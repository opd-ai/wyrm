// Package main contains utility functions for the Wyrm server.
// These functions are extracted to enable testing without Ebiten dependency.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network/federation"
	"github.com/opd-ai/wyrm/pkg/procgen/adapters"
	"github.com/opd-ai/wyrm/pkg/procgen/city"
	"github.com/opd-ai/wyrm/pkg/procgen/dungeon"
)

// initializeFactions generates factions using V-Series generators.
func initializeFactions(world *ecs.World, cfg *config.Config) *systems.FactionPoliticsSystem {
	factionAdapter := adapters.NewFactionAdapter()
	factions, err := factionAdapter.GenerateFactions(cfg.World.Seed, cfg.Genre, 20)
	if err != nil {
		log.Printf("warning: faction generation failed: %v", err)
		return systems.NewFactionPoliticsSystem(0.1)
	}

	log.Printf("generated %d factions for genre %s", len(factions), cfg.Genre)
	for _, f := range factions {
		log.Printf("  - %s (%s): %d members", f.Name, f.Type, f.MemberCount)
	}

	fps := systems.NewFactionPoliticsSystem(0.1)
	adapters.RegisterFactionsWithPoliticsSystem(fps, factions)
	return fps
}

// createDistrictEntity creates an entity for a single district.
func createDistrictEntity(world *ecs.World, district city.District) {
	districtEntity := world.CreateEntity()
	if err := world.AddComponent(districtEntity, &components.Position{
		X: district.CenterX,
		Y: district.CenterY,
		Z: 0,
	}); err != nil {
		log.Printf("warning: failed to add district position: %v", err)
	}
	if err := world.AddComponent(districtEntity, &components.EconomyNode{
		PriceTable: make(map[string]float64),
		Supply:     make(map[string]int),
		Demand:     make(map[string]int),
	}); err != nil {
		log.Printf("warning: failed to add economy node: %v", err)
	}
}

// initializeWorldClock creates the world clock entity.
func initializeWorldClock(world *ecs.World) {
	clockEntity := world.CreateEntity()
	if err := world.AddComponent(clockEntity, &components.WorldClock{
		Hour:       8, // Start at 8 AM
		Day:        1,
		HourLength: 60.0,
	}); err != nil {
		log.Fatalf("failed to add WorldClock component: %v", err)
	}
}

// initializeFederation sets up cross-server federation.
func initializeFederation(cfg *config.Config) *federation.Federation {
	nodeID := cfg.Federation.NodeID
	if nodeID == "" {
		// Generate a node ID from server address if not specified
		nodeID = fmt.Sprintf("node-%s", cfg.Server.Address)
	}

	fed := federation.NewFederation(nodeID)

	// Register peer nodes
	for _, peerAddr := range cfg.Federation.Peers {
		node := &federation.Node{
			ServerID: fmt.Sprintf("peer-%s", peerAddr),
			Address:  peerAddr,
		}
		fed.RegisterNode(node)
		log.Printf("registered federation peer: %s", peerAddr)
	}

	log.Printf("federation initialized with %d peers", fed.NodeCount())
	return fed
}

// runFederationCleanup handles periodic federation cleanup in a separate goroutine.
func runFederationCleanup(fed *federation.Federation, stopCh <-chan struct{}) {
	if fed == nil {
		return
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fed.CleanupExpired()
		case <-stopCh:
			return
		}
	}
}

// initializeDungeons generates instanced dungeon content for quests.
// Dungeons are generated with varying depths based on world seed and genre.
func initializeDungeons(world *ecs.World, cfg *config.Config) {
	dungeonGen := dungeon.NewGenerator(cfg.World.Seed, cfg.Genre)

	// Generate dungeons at varying depths (1-5) for quest content
	dungeonCount := 5
	for i := 0; i < dungeonCount; i++ {
		depth := (i % 5) + 1
		width := 50 + (depth * 10)
		height := 50 + (depth * 10)

		d := dungeonGen.Generate(width, height, depth)
		if d == nil {
			log.Printf("warning: dungeon generation failed for depth %d", depth)
			continue
		}

		// Create dungeon entrance entity in the world
		entranceEntity := world.CreateEntity()
		entranceX := float64(i*200) + 100 // Spread dungeons apart
		entranceY := float64(i*200) + 100

		if err := world.AddComponent(entranceEntity, &components.Position{
			X: entranceX,
			Y: entranceY,
			Z: 0,
		}); err != nil {
			log.Printf("warning: failed to add dungeon entrance position: %v", err)
			continue
		}

		// Store dungeon metadata in Interior component
		if err := world.AddComponent(entranceEntity, &components.Interior{
			ParentBuilding: uint64(entranceEntity),
			Width:          d.Width,
			Height:         d.Height,
			FloorType:      fmt.Sprintf("dungeon_%s_depth%d", cfg.Genre, depth),
		}); err != nil {
			log.Printf("warning: failed to add dungeon interior: %v", err)
			continue
		}

		log.Printf("  generated dungeon depth=%d with %d rooms at (%.0f, %.0f)",
			depth, len(d.Rooms), entranceX, entranceY)
	}

	log.Printf("generated %d dungeons for genre %s", dungeonCount, cfg.Genre)
}
