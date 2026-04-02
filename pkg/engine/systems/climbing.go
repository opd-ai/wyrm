package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// ClimbableSystem handles player climbing over low barriers.
// When a player approaches a climbable barrier, their Z position is smoothly
// adjusted to rise over the barrier and return to ground level on the other side.
type ClimbableSystem struct {
	// PlayerStepHeight is the maximum height the player can step up without climbing.
	PlayerStepHeight float64
	// ClimbDuration is the time in seconds to complete a climb animation.
	ClimbDuration float64
	// ActiveClimbs tracks ongoing climb animations per entity.
	ActiveClimbs map[ecs.Entity]*ClimbState
}

// ClimbState tracks the progress of a climb animation.
type ClimbState struct {
	StartZ       float64 // Starting Z position
	PeakZ        float64 // Peak Z position (top of barrier)
	EndZ         float64 // Ending Z position (ground level on other side)
	Progress     float64 // 0.0 to 1.0 animation progress
	Phase        int     // 0=ascending, 1=descending
	BarrierX     float64 // X position of barrier center
	BarrierY     float64 // Y position of barrier center
	ApproachDirX float64 // Direction player was moving when climb started
	ApproachDirY float64
}

// DefaultPlayerStepHeight is the default height a player can step up.
const DefaultPlayerStepHeight = 0.3

// DefaultClimbDuration is the default time to complete a climb.
const DefaultClimbDuration = 0.3

// NewClimbableSystem creates a new climbing system with default settings.
func NewClimbableSystem() *ClimbableSystem {
	return &ClimbableSystem{
		PlayerStepHeight: DefaultPlayerStepHeight,
		ClimbDuration:    DefaultClimbDuration,
		ActiveClimbs:     make(map[ecs.Entity]*ClimbState),
	}
}

// Update processes climbing animations and detects new climbs.
func (s *ClimbableSystem) Update(w *ecs.World, dt float64) {
	// Update active climb animations
	s.updateActiveClimbs(w, dt)

	// Check for new climb opportunities
	s.checkForNewClimbs(w)
}

// updateActiveClimbs progresses active climb animations.
func (s *ClimbableSystem) updateActiveClimbs(w *ecs.World, dt float64) {
	toRemove := []ecs.Entity{}

	for entity, state := range s.ActiveClimbs {
		posComp, ok := w.GetComponent(entity, "Position")
		if !ok {
			toRemove = append(toRemove, entity)
			continue
		}

		pos := posComp.(*components.Position)

		// Update progress
		state.Progress += dt / s.ClimbDuration
		if state.Progress >= 1.0 {
			state.Progress = 1.0
		}

		// Calculate current Z based on phase and progress
		if state.Phase == 0 {
			// Ascending phase: interpolate from StartZ to PeakZ
			pos.Z = lerp(state.StartZ, state.PeakZ, smoothStep(state.Progress))
			if state.Progress >= 1.0 {
				// Transition to descending phase
				state.Phase = 1
				state.Progress = 0.0
			}
		} else {
			// Descending phase: interpolate from PeakZ to EndZ
			pos.Z = lerp(state.PeakZ, state.EndZ, smoothStep(state.Progress))
			if state.Progress >= 1.0 {
				// Climb complete
				pos.Z = state.EndZ
				toRemove = append(toRemove, entity)
			}
		}
	}

	// Clean up completed climbs
	for _, entity := range toRemove {
		delete(s.ActiveClimbs, entity)
	}
}

// checkForNewClimbs detects when a player should start climbing.
func (s *ClimbableSystem) checkForNewClimbs(w *ecs.World) {
	// Get all climbable barriers
	barriers := w.Entities("Barrier", "Position")

	// Get all potential climbers (players with velocity)
	climbers := w.Entities("Position", "Player")

	for _, climberEntity := range climbers {
		// Skip if already climbing
		if _, climbing := s.ActiveClimbs[climberEntity]; climbing {
			continue
		}

		posComp, ok := w.GetComponent(climberEntity, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)

		// Check if player is near a climbable barrier
		for _, barrierEntity := range barriers {
			if s.shouldStartClimb(w, climberEntity, barrierEntity, pos) {
				break // Only climb one barrier at a time
			}
		}
	}
}

// shouldStartClimb checks if the player should begin climbing a barrier.
func (s *ClimbableSystem) shouldStartClimb(w *ecs.World, climber, barrierEntity ecs.Entity, climberPos *components.Position) bool {
	barrierComp, bOK := w.GetComponent(barrierEntity, "Barrier")
	barrierPosComp, bpOK := w.GetComponent(barrierEntity, "Position")
	if !bOK || !bpOK {
		return false
	}

	barrier := barrierComp.(*components.Barrier)
	barrierPos := barrierPosComp.(*components.Position)

	// Check if barrier is climbable
	if barrier.Shape.ClimbHeight <= 0 {
		return false
	}

	// Check if barrier height is within climbable range
	if barrier.Shape.Height > barrier.Shape.ClimbHeight {
		return false
	}

	// Calculate distance to barrier
	dx := climberPos.X - barrierPos.X
	dy := climberPos.Y - barrierPos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	// Check if within climb range (based on barrier shape)
	climbRange := s.getClimbRange(barrier)
	if dist > climbRange {
		return false
	}

	// Start climbing
	state := &ClimbState{
		StartZ:       climberPos.Z,
		PeakZ:        climberPos.Z + barrier.Shape.Height + 0.1, // Slightly above barrier
		EndZ:         climberPos.Z,                              // Return to ground level
		Progress:     0.0,
		Phase:        0,
		BarrierX:     barrierPos.X,
		BarrierY:     barrierPos.Y,
		ApproachDirX: -dx / dist,
		ApproachDirY: -dy / dist,
	}

	s.ActiveClimbs[climber] = state
	return true
}

// getClimbRange returns the distance at which climbing can be triggered.
func (s *ClimbableSystem) getClimbRange(barrier *components.Barrier) float64 {
	switch barrier.Shape.ShapeType {
	case "cylinder":
		return barrier.Shape.Radius + 0.3
	case "box":
		return math.Max(barrier.Shape.Width, barrier.Shape.Depth)/2 + 0.3
	default:
		return 0.5
	}
}

// IsClimbing returns true if the given entity is currently climbing.
func (s *ClimbableSystem) IsClimbing(entity ecs.Entity) bool {
	_, climbing := s.ActiveClimbs[entity]
	return climbing
}

// GetClimbProgress returns the climb progress (0.0-1.0) for an entity, or -1 if not climbing.
func (s *ClimbableSystem) GetClimbProgress(entity ecs.Entity) float64 {
	state, ok := s.ActiveClimbs[entity]
	if !ok {
		return -1
	}
	// Combine phases: 0-0.5 for ascending, 0.5-1.0 for descending
	if state.Phase == 0 {
		return state.Progress * 0.5
	}
	return 0.5 + state.Progress*0.5
}

// lerp performs linear interpolation.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// smoothStep provides smooth acceleration/deceleration curve.
func smoothStep(t float64) float64 {
	return t * t * (3 - 2*t)
}
