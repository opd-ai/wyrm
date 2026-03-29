package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// CombatSystem handles combat resolution and damage.
type CombatSystem struct{}

// Update processes combat resolution each tick.
func (s *CombatSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Health") {
		comp, ok := w.GetComponent(e, "Health")
		if !ok {
			continue
		}
		health := comp.(*components.Health)
		// Clamp health to max
		if health.Current > health.Max {
			health.Current = health.Max
		}
	}
}
