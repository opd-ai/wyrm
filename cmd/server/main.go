//go:build !noebiten

// Command server launches the Wyrm authoritative game server.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/network/federation"
	"github.com/opd-ai/wyrm/pkg/procgen/adapters"
	"github.com/opd-ai/wyrm/pkg/procgen/city"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	world := ecs.NewWorld()
	cm := chunk.NewManager(cfg.World.ChunkSize, cfg.World.Seed)

	fps := initializeFactions(world, cfg)
	initializeCity(world, cfg, fps)
	initializeWorldClock(world)
	registerServerSystems(world, cm, cfg, fps)

	// Initialize federation if enabled
	var fed *federation.Federation
	if cfg.Federation.Enabled {
		fed = initializeFederation(cfg)
	}

	srv := network.NewServer(cfg.Server.Address)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server start: %v\n", err)
		os.Exit(1)
	}
	defer srv.Stop()

	log.Printf("server listening on %s (tick_rate=%d)", cfg.Server.Address, cfg.Server.TickRate)
	if cfg.Federation.Enabled {
		log.Printf("federation enabled: node_id=%s, peers=%v", cfg.Federation.NodeID, cfg.Federation.Peers)
	}
	runServerLoop(world, cfg, srv, fed)
}

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

// initializeCity generates the city and spawns district entities.
func initializeCity(world *ecs.World, cfg *config.Config, fps *systems.FactionPoliticsSystem) {
	generatedCity := city.Generate(cfg.World.Seed, cfg.Genre)
	log.Printf("generated city: %s (%s) with %d districts", generatedCity.Name, cfg.Genre, len(generatedCity.Districts))

	// Generate NPCs using V-Series entity generator
	entityAdapter := adapters.NewEntityAdapter()
	factionAdapter := adapters.NewFactionAdapter()
	factions, _ := factionAdapter.GenerateFactions(cfg.World.Seed, cfg.Genre, 20)

	for i, district := range generatedCity.Districts {
		createDistrictEntity(world, district)

		// Spawn NPCs in each district using V-Series generator
		factionID := "neutral"
		if len(factions) > 0 {
			factionID = factions[i%len(factions)].ID
		}

		districtSeed := cfg.World.Seed + int64(i)*10000
		npcsPerDistrict := 5 + (district.Buildings / 10)
		if npcsPerDistrict > 20 {
			npcsPerDistrict = 20
		}

		npcCfg := adapters.NPCSpawnConfig{
			Seed:      districtSeed,
			Genre:     cfg.Genre,
			FactionID: factionID,
			Count:     npcsPerDistrict,
			CenterX:   district.CenterX,
			CenterY:   district.CenterY,
			Radius:    100.0,
		}
		entities, err := entityAdapter.GenerateAndSpawnNPCs(world, npcCfg)
		if err != nil {
			log.Printf("warning: NPC spawn failed for district %s: %v", district.Name, err)
			continue
		}
		log.Printf("  spawned %d NPCs in district %s", len(entities), district.Name)
	}
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

// registerServerSystems registers all server-side ECS systems.
func registerServerSystems(world *ecs.World, cm *chunk.Manager, cfg *config.Config, fps *systems.FactionPoliticsSystem) {
	world.RegisterSystem(systems.NewWorldClockSystem(60.0))
	world.RegisterSystem(systems.NewWorldChunkSystem(cm, cfg.World.ChunkSize))
	world.RegisterSystem(&systems.NPCScheduleSystem{})
	world.RegisterSystem(fps) // Use the pre-initialized faction politics system
	world.RegisterSystem(systems.NewCrimeSystem(60.0, 100.0))
	world.RegisterSystem(systems.NewEconomySystem(0.5, 0.1))
	world.RegisterSystem(&systems.CombatSystem{})
	world.RegisterSystem(&systems.VehicleSystem{})
	world.RegisterSystem(systems.NewQuestSystem())
	world.RegisterSystem(systems.NewWeatherSystem(cfg.Genre, 300.0))
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
		node := &federation.FederationNode{
			ServerID: fmt.Sprintf("peer-%s", peerAddr),
			Address:  peerAddr,
		}
		fed.RegisterNode(node)
		log.Printf("registered federation peer: %s", peerAddr)
	}

	log.Printf("federation initialized with %d peers", fed.NodeCount())
	return fed
}

// runServerLoop runs the main server tick loop until shutdown.
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

func runServerLoop(world *ecs.World, cfg *config.Config, srv *network.Server, fed *federation.Federation) {
	tickInterval := time.Second / time.Duration(cfg.Server.TickRate)
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start federation cleanup goroutine
	fedStopCh := make(chan struct{})
	go runFederationCleanup(fed, fedStopCh)
	defer close(fedStopCh)

	for {
		select {
		case <-ticker.C:
			world.Update(tickInterval.Seconds())
		case <-sigCh:
			log.Println("shutting down")
			return
		}
	}
}
