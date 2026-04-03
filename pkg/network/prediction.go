// Package network provides client-server networking.
package network

import (
	"math"
	"sync"
	"time"
)

// Latency mode thresholds for adaptive network behavior.
// Per README: "200-5000ms latency tolerance (designed for Tor-routed connections)"
const (
	// TorModeThreshold is the RTT threshold above which Tor-mode activates.
	// Per ROADMAP: When RTT exceeds 800ms, adaptive prediction engages.
	TorModeThreshold = 800 * time.Millisecond

	// HighLatencyThreshold activates enhanced prediction for 3000ms+ RTT.
	HighLatencyThreshold = 3000 * time.Millisecond

	// ExtremeLatencyThreshold activates maximum prediction for 5000ms+ RTT.
	ExtremeLatencyThreshold = 5000 * time.Millisecond
)

// Prediction windows for different latency modes.
const (
	// NormalPredictionWindow is the standard prediction window for low-latency connections.
	NormalPredictionWindow = 500 * time.Millisecond

	// TorModePredictionWindow is the increased prediction window in Tor-mode (800-3000ms RTT).
	// Per ROADMAP: Increase client prediction window to 1500ms.
	TorModePredictionWindow = 1500 * time.Millisecond

	// HighLatencyPredictionWindow for 3000-5000ms RTT connections.
	HighLatencyPredictionWindow = 3500 * time.Millisecond

	// ExtremeLatencyPredictionWindow for 5000ms+ RTT connections.
	ExtremeLatencyPredictionWindow = 6000 * time.Millisecond
)

// Input rates for different latency modes (Hz).
const (
	// NormalInputRate is the standard input send rate (Hz).
	NormalInputRate = 60

	// TorModeInputRate is the reduced input send rate in Tor-mode (Hz).
	// Per ROADMAP: Reduce input send rate to 10 Hz.
	TorModeInputRate = 10

	// HighLatencyInputRate for 3000-5000ms RTT (5 Hz).
	HighLatencyInputRate = 5

	// ExtremeLatencyInputRate for 5000ms+ RTT (2 Hz).
	ExtremeLatencyInputRate = 2
)

// Interpolation blend times for different latency modes.
const (
	// NormalBlendTime is the standard interpolation blend time.
	NormalBlendTime = 100 * time.Millisecond

	// TorModeBlendTime is the interpolation blend time in Tor-mode.
	// Per ROADMAP: Enable aggressive visual interpolation with 300ms blend time.
	TorModeBlendTime = 300 * time.Millisecond

	// HighLatencyBlendTime for 3000-5000ms RTT.
	HighLatencyBlendTime = 600 * time.Millisecond

	// ExtremeLatencyBlendTime for 5000ms+ RTT.
	ExtremeLatencyBlendTime = 1000 * time.Millisecond
)

// LatencyMode represents the current latency classification.
type LatencyMode int

const (
	// LatencyModeNormal for RTT < 800ms.
	LatencyModeNormal LatencyMode = iota
	// LatencyModeTor for 800ms <= RTT < 3000ms.
	LatencyModeTor
	// LatencyModeHigh for 3000ms <= RTT < 5000ms.
	LatencyModeHigh
	// LatencyModeExtreme for RTT >= 5000ms.
	LatencyModeExtreme
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
// Per README: Supports 200-5000ms latency tolerance for Tor-routed connections.
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

	// RTT tracking for adaptive latency modes
	smoothedRTT        time.Duration
	rttAlpha           float64 // Smoothing factor
	lastInputTime      time.Time
	latencyMode        LatencyMode
	torModeActive      bool // Backwards compatibility: true if latencyMode >= LatencyModeTor
	predictionWindow   time.Duration
	interpolationBlend time.Duration
	inputRateHz        int
	maxPendingInputs   int // Dynamic buffer size based on latency mode
}

// NewClientPredictor creates a new client-side predictor.
func NewClientPredictor() *ClientPredictor {
	return &ClientPredictor{
		pendingInputs:      make([]PredictedInput, 0, 256), // Increased for high-latency support
		moveSpeed:          5.0,                            // Units per second
		turnSpeed:          math.Pi / 2,                    // Radians per second (~90 degrees)
		rttAlpha:           0.125,
		latencyMode:        LatencyModeNormal,
		predictionWindow:   NormalPredictionWindow,
		interpolationBlend: NormalBlendTime,
		inputRateHz:        NormalInputRate,
		maxPendingInputs:   128, // Normal mode buffer size
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

	// Limit buffer size based on latency mode (dynamic sizing)
	if len(cp.pendingInputs) > cp.maxPendingInputs {
		cp.pendingInputs = cp.pendingInputs[len(cp.pendingInputs)-cp.maxPendingInputs:]
	}

	return input.SequenceNum
}

// applyInput updates predicted state based on input.
// Angles are in radians throughout for consistency with client and server.
func (cp *ClientPredictor) applyInput(input *PlayerInput, dt float32) {
	// Apply turning
	cp.predictedAngle += input.Turn * cp.turnSpeed * dt

	// Apply movement (in the direction we're facing)
	// Use standard library math functions for accuracy
	forwardRad := float64(cp.predictedAngle)
	cp.predictedPosition.X += input.MoveForward * cp.moveSpeed * dt * float32(math.Cos(forwardRad))
	cp.predictedPosition.Z += input.MoveForward * cp.moveSpeed * dt * float32(math.Sin(forwardRad))
	cp.predictedPosition.X += input.MoveRight * cp.moveSpeed * dt * float32(math.Sin(forwardRad))
	cp.predictedPosition.Z -= input.MoveRight * cp.moveSpeed * dt * float32(math.Cos(forwardRad))
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

// updateRTT updates the smoothed RTT estimate and adapts Tor-mode settings.
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
			// Adapt prediction parameters based on RTT
			cp.adaptToLatency()
			break
		}
	}
}

// adaptToLatency adjusts prediction parameters based on current RTT.
// Supports 4 latency modes per README: 200-5000ms latency tolerance.
func (cp *ClientPredictor) adaptToLatency() {
	prevMode := cp.latencyMode
	cp.latencyMode = cp.classifyLatency(cp.smoothedRTT)

	switch cp.latencyMode {
	case LatencyModeExtreme: // RTT >= 5000ms
		cp.predictionWindow = ExtremeLatencyPredictionWindow
		cp.interpolationBlend = ExtremeLatencyBlendTime
		cp.inputRateHz = ExtremeLatencyInputRate
		cp.maxPendingInputs = 512 // Large buffer for multi-second rewind

	case LatencyModeHigh: // 3000ms <= RTT < 5000ms
		cp.predictionWindow = HighLatencyPredictionWindow
		cp.interpolationBlend = HighLatencyBlendTime
		cp.inputRateHz = HighLatencyInputRate
		cp.maxPendingInputs = 256

	case LatencyModeTor: // 800ms <= RTT < 3000ms
		cp.predictionWindow = TorModePredictionWindow
		cp.interpolationBlend = TorModeBlendTime
		cp.inputRateHz = TorModeInputRate
		cp.maxPendingInputs = 192

	default: // RTT < 800ms
		cp.predictionWindow = NormalPredictionWindow
		cp.interpolationBlend = NormalBlendTime
		cp.inputRateHz = NormalInputRate
		cp.maxPendingInputs = 128
	}

	// Backwards compatibility: torModeActive = true for any elevated latency
	cp.torModeActive = cp.latencyMode >= LatencyModeTor

	// Log mode changes (will be visible in debug output)
	if prevMode != cp.latencyMode {
		// Mode changed - client code can check GetLatencyMode() for UI feedback
	}
}

// classifyLatency determines the latency mode based on RTT.
func (cp *ClientPredictor) classifyLatency(rtt time.Duration) LatencyMode {
	switch {
	case rtt >= ExtremeLatencyThreshold:
		return LatencyModeExtreme
	case rtt >= HighLatencyThreshold:
		return LatencyModeHigh
	case rtt >= TorModeThreshold:
		return LatencyModeTor
	default:
		return LatencyModeNormal
	}
}

// IsTorMode returns whether Tor-mode is currently active.
// Tor-mode is active for any RTT >= 800ms (LatencyModeTor, LatencyModeHigh, or LatencyModeExtreme).
func (cp *ClientPredictor) IsTorMode() bool {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.torModeActive
}

// GetLatencyMode returns the current latency mode classification.
func (cp *ClientPredictor) GetLatencyMode() LatencyMode {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.latencyMode
}

// GetMaxPendingInputs returns the current maximum pending inputs buffer size.
func (cp *ClientPredictor) GetMaxPendingInputs() int {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.maxPendingInputs
}

// ShouldSendInput returns true if enough time has passed to send another input.
// This enforces the adaptive input rate (10 Hz in Tor-mode, 60 Hz normally).
func (cp *ClientPredictor) ShouldSendInput() bool {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	interval := time.Second / time.Duration(cp.inputRateHz)
	if time.Since(cp.lastInputTime) >= interval {
		cp.lastInputTime = time.Now()
		return true
	}
	return false
}

// GetPredictionWindow returns the current prediction window duration.
func (cp *ClientPredictor) GetPredictionWindow() time.Duration {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.predictionWindow
}

// GetInterpolationBlend returns the current interpolation blend time.
func (cp *ClientPredictor) GetInterpolationBlend() time.Duration {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.interpolationBlend
}

// GetInputRateHz returns the current input send rate in Hz.
func (cp *ClientPredictor) GetInputRateHz() int {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return cp.inputRateHz
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
