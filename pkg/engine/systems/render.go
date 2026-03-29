package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// RenderSystem handles first-person raycasting and Ebitengine draw calls.
type RenderSystem struct {
	PlayerEntity ecs.Entity
}

// Update prepares render state based on player position each tick.
func (s *RenderSystem) Update(w *ecs.World, dt float64) {
	// Get player position for camera
	if s.PlayerEntity != 0 {
		_, _ = w.GetComponent(s.PlayerEntity, "Position")
	}
}
