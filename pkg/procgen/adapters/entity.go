// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// EntityAdapter wraps Venture's entity generator for Wyrm's ECS world.
type EntityAdapter struct {
	generator *entity.EntityGenerator
}

// NewEntityAdapter creates a new entity adapter.
func NewEntityAdapter() *EntityAdapter {
	return &EntityAdapter{
		generator: entity.NewEntityGenerator(),
	}
}

// mapGenreID normalizes Wyrm genre IDs to Venture's format.
func mapGenreID(genre string) string {
	switch genre {
	case "sci-fi":
		return "scifi"
	case "post-apocalyptic":
		return "postapoc"
	default:
		return genre
	}
}

// GenerateEntity creates a single entity and returns its components.
func (a *EntityAdapter) GenerateEntity(seed int64, genre string, depth int) (*NPCData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Depth:      depth,
		Difficulty: float64(depth) / 100.0, // Scale difficulty with depth
		Custom:     map[string]interface{}{"count": 1},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("entity generation failed: %w", err)
	}

	entities, ok := result.([]*entity.Entity)
	if !ok || len(entities) == 0 {
		return nil, fmt.Errorf("invalid entity result type")
	}

	e := entities[0]
	return &NPCData{
		Name:   e.Name,
		Health: float64(e.Stats.MaxHealth),
		Tags:   e.Tags,
	}, nil
}

// NPCData holds generated NPC data before ECS registration.
type NPCData struct {
	Name   string
	Health float64
	Tags   []string
}

// SpawnNPC creates an NPC entity in the ECS world from generated data.
func SpawnNPC(world *ecs.World, data *NPCData, x, y float64, factionID string) (ecs.Entity, error) {
	e := world.CreateEntity()

	if err := world.AddComponent(e, &components.Position{X: x, Y: y, Z: 0}); err != nil {
		return 0, fmt.Errorf("failed to add Position: %w", err)
	}

	if err := world.AddComponent(e, &components.Health{
		Current: data.Health,
		Max:     data.Health,
	}); err != nil {
		return 0, fmt.Errorf("failed to add Health: %w", err)
	}

	if err := world.AddComponent(e, &components.Faction{
		ID:         factionID,
		Reputation: 0,
	}); err != nil {
		return 0, fmt.Errorf("failed to add Faction: %w", err)
	}

	// Add a basic schedule for NPCs
	schedule := &components.Schedule{
		CurrentActivity: "idle",
		TimeSlots: map[int]string{
			0: "sleep", 1: "sleep", 2: "sleep", 3: "sleep",
			4: "sleep", 5: "sleep", 6: "wake", 7: "eat",
			8: "work", 9: "work", 10: "work", 11: "work",
			12: "eat", 13: "work", 14: "work", 15: "work",
			16: "work", 17: "eat", 18: "socialize", 19: "socialize",
			20: "socialize", 21: "relax", 22: "sleep", 23: "sleep",
		},
	}
	if err := world.AddComponent(e, schedule); err != nil {
		return 0, fmt.Errorf("failed to add Schedule: %w", err)
	}

	return e, nil
}

// GenerateAndSpawnNPCs generates multiple NPCs and spawns them in the world.
func (a *EntityAdapter) GenerateAndSpawnNPCs(world *ecs.World, seed int64, genre, factionID string, count int, centerX, centerY, radius float64) ([]ecs.Entity, error) {
	rng := rand.New(rand.NewSource(seed))
	entities := make([]ecs.Entity, 0, count)

	for i := 0; i < count; i++ {
		npcSeed := seed + int64(i)*1000
		data, err := a.GenerateEntity(npcSeed, genre, i/10)
		if err != nil {
			continue // Skip failed generations
		}

		// Place NPC at random position within radius
		angle := rng.Float64() * 6.283185307 // 2*pi
		dist := rng.Float64() * radius
		x := centerX + dist*float64(rng.Intn(2)*2-1)*0.5
		y := centerY + dist*float64(rng.Intn(2)*2-1)*0.5
		x = centerX + dist*0.7071*float64(rng.Intn(3)-1)
		y = centerY + dist*0.7071*float64(rng.Intn(3)-1)

		// Use deterministic placement
		x = centerX + float64(i%5-2)*radius/5
		y = centerY + float64(i/5-count/10)*radius/5

		e, err := SpawnNPC(world, data, x, y, factionID)
		if err != nil {
			continue
		}
		entities = append(entities, e)
	}

	return entities, nil
}
