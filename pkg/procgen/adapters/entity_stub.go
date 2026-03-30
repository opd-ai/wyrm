//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// EntityAdapter wraps Venture's entity generator for Wyrm's ECS world.
// Stub implementation for headless testing.
type EntityAdapter struct {
	seed int64
}

// NewEntityAdapter creates a new entity adapter.
func NewEntityAdapter() *EntityAdapter {
	return &EntityAdapter{}
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
// Stub implementation for headless testing.
func (a *EntityAdapter) GenerateEntity(seed int64, genre string, depth int) (*NPCData, error) {
	rng := rand.New(rand.NewSource(seed))
	names := []string{"Guard", "Merchant", "Traveler", "Scholar", "Warrior"}
	name := names[rng.Intn(len(names))]
	health := BaseNPCHealth + float64(rng.Intn(NPCHealthVariance))

	return &NPCData{
		Name:   fmt.Sprintf("%s_%s", name, genre),
		Health: health,
		Tags:   []string{genre, "npc"},
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
		Reputation: DefaultFactionReputation,
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

// NPCSpawnConfig holds parameters for spawning multiple NPCs.
type NPCSpawnConfig struct {
	Seed      int64
	Genre     string
	FactionID string
	Count     int
	CenterX   float64
	CenterY   float64
	Radius    float64
}

// GenerateAndSpawnNPCs generates multiple NPCs and spawns them in the world.
func (a *EntityAdapter) GenerateAndSpawnNPCs(world *ecs.World, cfg NPCSpawnConfig) ([]ecs.Entity, error) {
	rng := rand.New(rand.NewSource(cfg.Seed))
	entities := make([]ecs.Entity, 0, cfg.Count)

	for i := 0; i < cfg.Count; i++ {
		npcSeed := cfg.Seed + int64(i)*NPCSeedMultiplier
		data, err := a.GenerateEntity(npcSeed, cfg.Genre, i/DefaultNPCDepthDivisor)
		if err != nil {
			continue // Skip failed generations
		}

		// Use deterministic grid-based placement within radius
		// Consume RNG values for reproducibility
		_ = rng.Float64() // Reserved for future random offset
		_ = rng.Float64() // Reserved for future random offset

		x := cfg.CenterX + float64(i%NPCGridColumns-NPCGridOffset)*cfg.Radius/NPCGridSpacingDivisor
		y := cfg.CenterY + float64(i/NPCGridColumns-cfg.Count/DefaultNPCDepthDivisor)*cfg.Radius/NPCGridSpacingDivisor

		e, err := SpawnNPC(world, data, x, y, cfg.FactionID)
		if err != nil {
			continue
		}
		entities = append(entities, e)
	}

	return entities, nil
}
