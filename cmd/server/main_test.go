//go:build noebiten

// Package main contains tests for the Wyrm server.
// These tests use the noebiten build tag to run without Ebiten initialization.
package main

import (
	"testing"
	"time"

	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network/federation"
	"github.com/opd-ai/wyrm/pkg/procgen/city"
)

func TestInitializeFactions(t *testing.T) {
	world := ecs.NewWorld()
	cfg := &config.Config{
		Genre: "fantasy",
		World: config.WorldConfig{
			Seed: 12345,
		},
	}

	fps := initializeFactions(world, cfg)
	if fps == nil {
		t.Error("initializeFactions returned nil")
	}
}

func TestInitializeFactions_DifferentGenres(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := ecs.NewWorld()
			cfg := &config.Config{
				Genre: genre,
				World: config.WorldConfig{
					Seed: 12345,
				},
			}

			fps := initializeFactions(world, cfg)
			if fps == nil {
				t.Errorf("initializeFactions returned nil for genre %s", genre)
			}
		})
	}
}

func TestCreateDistrictEntity(t *testing.T) {
	world := ecs.NewWorld()
	district := city.District{
		Name:      "Test District",
		Type:      "residential",
		CenterX:   100.0,
		CenterY:   200.0,
		Buildings: 10,
	}

	createDistrictEntity(world, district)

	// Verify entities were created
	entities := world.Entities("Position", "EconomyNode")
	if len(entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(entities))
	}

	// Verify position component
	pos, ok := world.GetComponent(entities[0], "Position")
	if !ok {
		t.Fatal("Position component not found")
	}
	position := pos.(*components.Position)
	if position.X != district.CenterX || position.Y != district.CenterY {
		t.Errorf("position mismatch: got (%f, %f), want (%f, %f)",
			position.X, position.Y, district.CenterX, district.CenterY)
	}

	// Verify economy node component
	econ, ok := world.GetComponent(entities[0], "EconomyNode")
	if !ok {
		t.Fatal("EconomyNode component not found")
	}
	economyNode := econ.(*components.EconomyNode)
	if economyNode.PriceTable == nil {
		t.Error("PriceTable is nil")
	}
	if economyNode.Supply == nil {
		t.Error("Supply is nil")
	}
	if economyNode.Demand == nil {
		t.Error("Demand is nil")
	}
}

func TestInitializeWorldClock(t *testing.T) {
	world := ecs.NewWorld()
	initializeWorldClock(world)

	// Verify clock entity was created
	entities := world.Entities("WorldClock")
	if len(entities) != 1 {
		t.Errorf("expected 1 WorldClock entity, got %d", len(entities))
	}

	// Verify clock component values
	clock, ok := world.GetComponent(entities[0], "WorldClock")
	if !ok {
		t.Fatal("WorldClock component not found")
	}
	wc := clock.(*components.WorldClock)
	if wc.Hour != 8 {
		t.Errorf("expected hour 8, got %d", wc.Hour)
	}
	if wc.Day != 1 {
		t.Errorf("expected day 1, got %d", wc.Day)
	}
	if wc.HourLength != 60.0 {
		t.Errorf("expected hour length 60.0, got %f", wc.HourLength)
	}
}

func TestInitializeFederation(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: "localhost:7777",
		},
		Federation: config.FederationConfig{
			Enabled: true,
			NodeID:  "test-node",
			Peers:   []string{"peer1:7777", "peer2:7777"},
		},
	}

	fed := initializeFederation(cfg)
	if fed == nil {
		t.Fatal("initializeFederation returned nil")
	}

	// Check that peers were registered
	if fed.NodeCount() != 2 {
		t.Errorf("expected 2 peer nodes, got %d", fed.NodeCount())
	}
}

func TestInitializeFederation_AutoNodeID(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: "localhost:7777",
		},
		Federation: config.FederationConfig{
			Enabled: true,
			NodeID:  "", // Empty - should auto-generate
			Peers:   []string{},
		},
	}

	fed := initializeFederation(cfg)
	if fed == nil {
		t.Fatal("initializeFederation returned nil")
	}
}

func TestRunFederationCleanup_StopsOnSignal(t *testing.T) {
	fed := federation.NewFederation("test-node")
	stopCh := make(chan struct{})

	done := make(chan bool)
	go func() {
		runFederationCleanup(fed, stopCh)
		done <- true
	}()

	// Stop immediately
	close(stopCh)

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("runFederationCleanup did not stop within timeout")
	}
}

func TestRunFederationCleanup_NilFederation(t *testing.T) {
	stopCh := make(chan struct{})

	done := make(chan bool)
	go func() {
		runFederationCleanup(nil, stopCh)
		done <- true
	}()

	// Should return immediately for nil federation
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("runFederationCleanup did not return immediately for nil federation")
	}
}

func BenchmarkInitializeFactions(b *testing.B) {
	cfg := &config.Config{
		Genre: "fantasy",
		World: config.WorldConfig{
			Seed: 12345,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := ecs.NewWorld()
		_ = initializeFactions(world, cfg)
	}
}

func BenchmarkCreateDistrictEntity(b *testing.B) {
	district := city.District{
		Name:      "Benchmark District",
		Type:      "commercial",
		CenterX:   100.0,
		CenterY:   200.0,
		Buildings: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := ecs.NewWorld()
		createDistrictEntity(world, district)
	}
}

func BenchmarkInitializeWorldClock(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := ecs.NewWorld()
		initializeWorldClock(world)
	}
}

// BenchmarkServerTickWith200NPCs benchmarks the server tick loop with 200 NPC entities.
// This validates the ROADMAP.md requirement: "200 NPCs + 32 players in ≤20ms server tick".
func BenchmarkServerTickWith200NPCs(b *testing.B) {
	world := ecs.NewWorld()

	// Initialize world clock (required by NPCScheduleSystem)
	initializeWorldClock(world)

	// Create 200 NPCs with typical component sets
	for i := 0; i < 200; i++ {
		e := world.CreateEntity()
		_ = world.AddComponent(e, &components.Position{
			X: float64(i%50) * 10.0,
			Y: float64(i/50) * 10.0,
			Z: 0,
		})
		_ = world.AddComponent(e, &components.Health{
			Current: 100,
			Max:     100,
		})
		_ = world.AddComponent(e, &components.Faction{
			ID: "TestFaction",
		})
		_ = world.AddComponent(e, &components.Schedule{
			CurrentActivity: "idle",
			TimeSlots: map[int]string{
				0: "sleep", 1: "sleep", 2: "sleep", 3: "sleep", 4: "sleep", 5: "sleep",
				6: "wake", 7: "eat", 8: "work", 9: "work", 10: "work", 11: "work",
				12: "eat", 13: "work", 14: "work", 15: "work", 16: "work", 17: "eat",
				18: "socialize", 19: "socialize", 20: "rest", 21: "rest", 22: "sleep", 23: "sleep",
			},
		})
	}

	// Create 32 player entities
	for i := 0; i < 32; i++ {
		e := world.CreateEntity()
		_ = world.AddComponent(e, &components.Position{
			X: float64(i%8) * 100.0,
			Y: float64(i/8) * 100.0,
			Z: 0,
		})
		_ = world.AddComponent(e, &components.Health{
			Current: 100,
			Max:     100,
		})
	}

	// Register systems typically running in server tick
	world.RegisterSystem(&systems.NPCScheduleSystem{})
	world.RegisterSystem(systems.NewFactionPoliticsSystem(0.1))

	dt := 1.0 / 20.0 // 20 Hz tick rate (50ms)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world.Update(dt)
	}
}

// BenchmarkServerTickWith1000Entities benchmarks a stress test scenario.
func BenchmarkServerTickWith1000Entities(b *testing.B) {
	world := ecs.NewWorld()

	// Initialize world clock
	initializeWorldClock(world)

	// Create 1000 entities with various components
	for i := 0; i < 1000; i++ {
		e := world.CreateEntity()
		_ = world.AddComponent(e, &components.Position{
			X: float64(i%100) * 5.0,
			Y: float64(i/100) * 5.0,
			Z: 0,
		})
		if i%2 == 0 {
			_ = world.AddComponent(e, &components.Health{
				Current: 100,
				Max:     100,
			})
		}
		if i%3 == 0 {
			_ = world.AddComponent(e, &components.Schedule{
				CurrentActivity: "idle",
				TimeSlots: map[int]string{
					8: "work", 12: "eat", 17: "rest", 22: "sleep",
				},
			})
		}
	}

	// Register systems
	world.RegisterSystem(&systems.NPCScheduleSystem{})

	dt := 1.0 / 20.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world.Update(dt)
	}
}
