package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// RenderSystem handles first-person raycasting and Ebitengine draw calls.
type RenderSystem struct {
	PlayerEntity ecs.Entity
	// TargetEntity is the entity currently under the crosshair (if any).
	TargetEntity ecs.Entity
	// TargetValid indicates whether TargetEntity is a valid interaction target.
	TargetValid bool
	// TargetDistance is the distance to the target entity.
	TargetDistance float64
	// MaxTargetDistance is the maximum distance for targeting entities.
	MaxTargetDistance float64
	// TargetingAngleTolerance is the angle tolerance in radians for targeting.
	TargetingAngleTolerance float64
}

// NewRenderSystem creates a RenderSystem with default targeting parameters.
func NewRenderSystem() *RenderSystem {
	return &RenderSystem{
		MaxTargetDistance:       10.0, // 10 world units
		TargetingAngleTolerance: 0.15, // ~8.5 degrees
	}
}

// Update prepares render state based on player position each tick.
// This includes updating the interaction target based on the player's look direction.
func (s *RenderSystem) Update(w *ecs.World, _ float64) {
	s.TargetValid = false
	s.TargetEntity = 0

	// Get player position and orientation
	playerPos, ok := w.GetComponent(s.PlayerEntity, "Position")
	if !ok {
		return
	}
	pos := playerPos.(*components.Position)

	// Calculate player look direction
	lookX := math.Cos(pos.Angle)
	lookY := math.Sin(pos.Angle)

	// Find the closest interactable entity in the look direction
	var bestEntity ecs.Entity
	bestDist := s.MaxTargetDistance + 1.0 // Start beyond max
	bestScore := -1.0                     // Higher score = better target

	for _, e := range w.Entities("EnvironmentObject", "Position") {
		if e == s.PlayerEntity {
			continue
		}

		targetPos, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		tPos := targetPos.(*components.Position)

		// Vector from player to target
		dx := tPos.X - pos.X
		dy := tPos.Y - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		// Skip if too far or too close
		if dist > s.MaxTargetDistance || dist < 0.1 {
			continue
		}

		// Normalize direction to target
		invDist := 1.0 / dist
		ndx := dx * invDist
		ndy := dy * invDist

		// Dot product with look direction (1 = looking directly at, -1 = behind)
		dot := ndx*lookX + ndy*lookY

		// Skip if not in front
		if dot < math.Cos(s.TargetingAngleTolerance) {
			continue
		}

		// Score: prefer higher dot product (more aligned) and closer distance
		score := dot - dist*0.1

		if score > bestScore {
			bestScore = score
			bestDist = dist
			bestEntity = e
		}
	}

	if bestDist <= s.MaxTargetDistance {
		s.TargetEntity = bestEntity
		s.TargetValid = true
		s.TargetDistance = bestDist
	}
}

// GetTarget returns the currently targeted entity and whether it's valid.
func (s *RenderSystem) GetTarget() (ecs.Entity, bool) {
	return s.TargetEntity, s.TargetValid
}

// SetPlayerEntity sets the player entity for the render system.
func (s *RenderSystem) SetPlayerEntity(e ecs.Entity) {
	s.PlayerEntity = e
}
