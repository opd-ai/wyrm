// Package systems contains all ECS system implementations.
package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

// ChunkLoader defines the interface for loading and unloading chunks.
type ChunkLoader interface {
	GetChunk(x, y int) *chunk.Chunk
}

// WorldChunkSystem manages loading and unloading of world chunks.
type WorldChunkSystem struct {
	Manager   ChunkLoader
	chunkSize int
}

// NewWorldChunkSystem creates a new chunk system with the given manager.
func NewWorldChunkSystem(manager ChunkLoader, chunkSize int) *WorldChunkSystem {
	return &WorldChunkSystem{
		Manager:   manager,
		chunkSize: chunkSize,
	}
}

// Update loads chunks around entities with Position components.
func (s *WorldChunkSystem) Update(w *ecs.World, dt float64) {
	if s.Manager == nil {
		return
	}
	for _, e := range w.Entities("Position") {
		comp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := comp.(*components.Position)
		chunkX := int(pos.X) / s.chunkSize
		chunkY := int(pos.Y) / s.chunkSize
		// Load the 3x3 chunk window around the entity
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				_ = s.Manager.GetChunk(chunkX+dx, chunkY+dy)
			}
		}
	}
}

// NPCScheduleSystem drives NPC daily activity cycles.
type NPCScheduleSystem struct {
	WorldHour int
}

// Update processes NPC schedules based on the current world hour.
func (s *NPCScheduleSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Schedule") {
		comp, ok := w.GetComponent(e, "Schedule")
		if !ok {
			continue
		}
		sched := comp.(*components.Schedule)
		if activity, ok := sched.TimeSlots[s.WorldHour]; ok {
			if activity != sched.CurrentActivity {
				sched.CurrentActivity = activity
			}
		}
	}
}

// FactionPoliticsSystem handles faction relationships, wars, and treaties.
type FactionPoliticsSystem struct{}

// Update processes faction politics each tick.
func (s *FactionPoliticsSystem) Update(w *ecs.World, dt float64) {
	// Query entities with Faction component for relationship updates
	_ = w.Entities("Faction")
}

// CrimeSystem tracks crimes, wanted levels, witnesses, and bounties.
type CrimeSystem struct{}

// Update processes crime detection and bounty updates each tick.
func (s *CrimeSystem) Update(w *ecs.World, dt float64) {
	// Future: query witness entities, update wanted levels
}

// EconomySystem manages supply, demand, and pricing across city nodes.
type EconomySystem struct{}

// Update processes economic simulation each tick.
func (s *EconomySystem) Update(w *ecs.World, dt float64) {
	// Future: update supply/demand, adjust prices
}

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

// VehicleSystem manages vehicle movement and physics.
type VehicleSystem struct{}

// Update processes vehicle physics each tick.
func (s *VehicleSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Vehicle", "Position") {
		vComp, ok := w.GetComponent(e, "Vehicle")
		if !ok {
			continue
		}
		pComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		vehicle := vComp.(*components.Vehicle)
		pos := pComp.(*components.Position)
		// Apply vehicle movement based on speed
		if vehicle.Fuel > 0 && vehicle.Speed > 0 {
			pos.X += vehicle.Speed * dt
			vehicle.Fuel -= vehicle.Speed * dt * 0.01
			if vehicle.Fuel < 0 {
				vehicle.Fuel = 0
			}
		}
	}
}

// QuestSystem manages quest state, branching, and consequence flags.
type QuestSystem struct{}

// Update processes quest state transitions each tick.
func (s *QuestSystem) Update(w *ecs.World, dt float64) {
	// Future: check quest conditions, trigger state transitions
}

// WeatherSystem controls dynamic weather and environmental hazards.
type WeatherSystem struct {
	CurrentWeather string
	TimeAccum      float64
}

// Update advances weather simulation each tick.
func (s *WeatherSystem) Update(w *ecs.World, dt float64) {
	s.TimeAccum += dt
	// Weather changes occur periodically
	if s.CurrentWeather == "" {
		s.CurrentWeather = "clear"
	}
}

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

// AudioSystem drives procedural audio synthesis.
type AudioSystem struct {
	Genre string
}

// Update advances audio synthesis each tick.
func (s *AudioSystem) Update(w *ecs.World, dt float64) {
	// Future: update audio based on player position and game state
}
