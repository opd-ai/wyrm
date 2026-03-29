// Package network provides client-server networking.
package network

import (
	"sync"
	"time"
)

// PredictedInput stores input and the predicted state resulting from it.
type PredictedInput struct {
	SequenceNum uint32
	Input       *PlayerInput
	Position    Position3D
	Angle       float32
	Timestamp   time.Time
}

// Position3D represents a 3D position.
type Position3D struct {
	X, Y, Z float32
}

// ClientPredictor handles client-side prediction and server reconciliation.
// Per ROADMAP Phase 5 item 19:
// AC: Movement feels responsive at 200ms RTT; no visible rubber-banding at 500ms RTT.
type ClientPredictor struct {
	mu sync.Mutex

	// Pending inputs awaiting server acknowledgment
	pendingInputs []PredictedInput

	// Current predicted state
	predictedPosition Position3D
	predictedAngle    float32

	// Last acknowledged sequence from server
	lastAckedSequence uint32

	// Next sequence number for outgoing inputs
	nextSequence uint32

	// Movement parameters
	moveSpeed float32
	turnSpeed float32

	// RTT tracking
	smoothedRTT time.Duration
	rttAlpha    float64 // Smoothing factor
}

// NewClientPredictor creates a new client-side predictor.
func NewClientPredictor() *ClientPredictor {
	return &ClientPredictor{
		pendingInputs: make([]PredictedInput, 0, 64),
		moveSpeed:     5.0,  // Units per second
		turnSpeed:     90.0, // Degrees per second
		rttAlpha:      0.125,
	}
}

// RecordInput records a new input and predicts its effect locally.
// Returns the sequence number assigned to this input.
func (cp *ClientPredictor) RecordInput(input *PlayerInput, dt float32) uint32 {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Assign sequence number
	input.SequenceNum = cp.nextSequence
	input.ClientTimeMs = uint32(time.Now().UnixMilli())
	cp.nextSequence++

	// Apply input to predicted state
	cp.applyInput(input, dt)

	// Store for reconciliation
	cp.pendingInputs = append(cp.pendingInputs, PredictedInput{
		SequenceNum: input.SequenceNum,
		Input:       input,
		Position:    cp.predictedPosition,
		Angle:       cp.predictedAngle,
		Timestamp:   time.Now(),
	})

	// Limit buffer size (keep last 128 inputs)
	if len(cp.pendingInputs) > 128 {
		cp.pendingInputs = cp.pendingInputs[len(cp.pendingInputs)-128:]
	}

	return input.SequenceNum
}

// applyInput updates predicted state based on input.
func (cp *ClientPredictor) applyInput(input *PlayerInput, dt float32) {
	// Apply turning
	cp.predictedAngle += input.Turn * cp.turnSpeed * dt

	// Apply movement (in the direction we're facing)
	forwardRad := float32(cp.predictedAngle) * (3.14159265 / 180.0)
	cp.predictedPosition.X += input.MoveForward * cp.moveSpeed * dt * float32(cos(float64(forwardRad)))
	cp.predictedPosition.Z += input.MoveForward * cp.moveSpeed * dt * float32(sin(float64(forwardRad)))
	cp.predictedPosition.X += input.MoveRight * cp.moveSpeed * dt * float32(sin(float64(forwardRad)))
	cp.predictedPosition.Z -= input.MoveRight * cp.moveSpeed * dt * float32(cos(float64(forwardRad)))
}

// cos calculates cosine for float64.
func cos(x float64) float64 {
	// Simple approximation using Taylor series
	x = mod(x, 6.28318530718)
	if x > 3.14159265359 {
		x -= 6.28318530718
	}
	x2 := x * x
	return 1 - x2/2 + x2*x2/24 - x2*x2*x2/720
}

// sin calculates sine for float64.
func sin(x float64) float64 {
	return cos(x - 1.5707963268)
}

// mod performs modulo for float64.
func mod(a, b float64) float64 {
	for a >= b {
		a -= b
	}
	for a < 0 {
		a += b
	}
	return a
}

// Reconcile handles server state updates and replays un-acknowledged inputs.
func (cp *ClientPredictor) Reconcile(serverState *WorldState, localEntityID uint64) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	serverPos, serverAngle, found := cp.findLocalEntityState(serverState, localEntityID)
	if !found {
		return
	}

	cp.updateRTT(serverState)
	cp.removeAcknowledgedInputs(serverState.AckSequence)
	cp.resetToServerState(serverPos, serverAngle)
	cp.replayPendingInputs()
}

// findLocalEntityState locates the local player's entity in the server state.
func (cp *ClientPredictor) findLocalEntityState(serverState *WorldState, localEntityID uint64) (Position3D, float32, bool) {
	for _, e := range serverState.Entities {
		if e.EntityID == localEntityID {
			return Position3D{X: e.X, Y: e.Y, Z: e.Z}, e.Angle, true
		}
	}
	return Position3D{}, 0, false
}

// removeAcknowledgedInputs removes inputs that have been acknowledged by the server.
func (cp *ClientPredictor) removeAcknowledgedInputs(ackedSeq uint32) {
	newPending := cp.pendingInputs[:0]
	for _, pi := range cp.pendingInputs {
		if pi.SequenceNum > ackedSeq {
			newPending = append(newPending, pi)
		}
	}
	cp.pendingInputs = newPending
	cp.lastAckedSequence = ackedSeq
}

// resetToServerState resets the predicted state to match server authoritative state.
func (cp *ClientPredictor) resetToServerState(pos Position3D, angle float32) {
	cp.predictedPosition = pos
	cp.predictedAngle = angle
}

// replayPendingInputs re-applies all unacknowledged inputs after server state reset.
func (cp *ClientPredictor) replayPendingInputs() {
	const defaultDT = float32(0.016) // Default 60fps
	for _, pi := range cp.pendingInputs {
		cp.applyInput(pi.Input, defaultDT)
	}
}

// updateRTT updates the smoothed RTT estimate.
func (cp *ClientPredictor) updateRTT(serverState *WorldState) {
	// Find the input that was acknowledged
	for _, pi := range cp.pendingInputs {
		if pi.SequenceNum == serverState.AckSequence {
			rtt := time.Since(pi.Timestamp)
			// Exponential moving average
			if cp.smoothedRTT == 0 {
				cp.smoothedRTT = rtt
			} else {
				cp.smoothedRTT = time.Duration(
					float64(cp.smoothedRTT)*(1-cp.rttAlpha) +
						float64(rtt)*cp.rttAlpha,
				)
			}
			break
		}
	}
}

// GetPredictedPosition returns the current predicted position.
func (cp *ClientPredictor) GetPredictedPosition() Position3D {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.predictedPosition
}

// GetPredictedAngle returns the current predicted angle.
func (cp *ClientPredictor) GetPredictedAngle() float32 {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.predictedAngle
}

// GetSmoothedRTT returns the current smoothed RTT estimate.
func (cp *ClientPredictor) GetSmoothedRTT() time.Duration {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.smoothedRTT
}

// PendingInputCount returns the number of inputs awaiting acknowledgment.
func (cp *ClientPredictor) PendingInputCount() int {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return len(cp.pendingInputs)
}

// SetPosition sets the predicted position (for initialization).
func (cp *ClientPredictor) SetPosition(pos Position3D) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.predictedPosition = pos
}

// SetAngle sets the predicted angle (for initialization).
func (cp *ClientPredictor) SetAngle(angle float32) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.predictedAngle = angle
}
