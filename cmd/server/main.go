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

	initializeCity(world, cfg)
	initializeWorldClock(world)
	registerServerSystems(world, cm, cfg)

	srv := network.NewServer(cfg.Server.Address)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server start: %v\n", err)
		os.Exit(1)
	}
	defer srv.Stop()

	log.Printf("server listening on %s (tick_rate=%d)", cfg.Server.Address, cfg.Server.TickRate)
	runServerLoop(world, cfg, srv)
}

// initializeCity generates the city and spawns district entities.
func initializeCity(world *ecs.World, cfg *config.Config) {
	generatedCity := city.Generate(cfg.World.Seed, cfg.Genre)
	log.Printf("generated city: %s (%s) with %d districts", generatedCity.Name, cfg.Genre, len(generatedCity.Districts))
	for _, district := range generatedCity.Districts {
		createDistrictEntity(world, district)
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
func registerServerSystems(world *ecs.World, cm *chunk.Manager, cfg *config.Config) {
	world.RegisterSystem(systems.NewWorldClockSystem(60.0))
	world.RegisterSystem(systems.NewWorldChunkSystem(cm, cfg.World.ChunkSize))
	world.RegisterSystem(&systems.NPCScheduleSystem{})
	world.RegisterSystem(systems.NewFactionPoliticsSystem(0.1))
	world.RegisterSystem(systems.NewCrimeSystem(60.0, 100.0))
	world.RegisterSystem(systems.NewEconomySystem(0.5, 0.1))
	world.RegisterSystem(&systems.CombatSystem{})
	world.RegisterSystem(&systems.VehicleSystem{})
	world.RegisterSystem(systems.NewQuestSystem())
	world.RegisterSystem(systems.NewWeatherSystem(cfg.Genre, 300.0))
}

// runServerLoop runs the main server tick loop until shutdown.
func runServerLoop(world *ecs.World, cfg *config.Config, srv *network.Server) {
	tickInterval := time.Second / time.Duration(cfg.Server.TickRate)
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			dt := tickInterval.Seconds()
			world.Update(dt)
		case <-sigCh:
			log.Println("shutting down")
			return
		}
	}
}
