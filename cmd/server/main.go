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
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	world := ecs.NewWorld()
	cm := chunk.NewChunkManager(cfg.World.ChunkSize, cfg.World.Seed)

	// Register server-side systems
	world.RegisterSystem(systems.NewWorldChunkSystem(cm, cfg.World.ChunkSize))
	world.RegisterSystem(&systems.NPCScheduleSystem{})
	world.RegisterSystem(&systems.FactionPoliticsSystem{})
	world.RegisterSystem(&systems.CrimeSystem{})
	world.RegisterSystem(&systems.EconomySystem{})
	world.RegisterSystem(&systems.CombatSystem{})
	world.RegisterSystem(&systems.VehicleSystem{})
	world.RegisterSystem(&systems.QuestSystem{})
	world.RegisterSystem(&systems.WeatherSystem{})

	srv := network.NewServer(cfg.Server.Address)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server start: %v\n", err)
		os.Exit(1)
	}
	defer srv.Stop()

	log.Printf("server listening on %s (tick_rate=%d)", cfg.Server.Address, cfg.Server.TickRate)

	// Start tick loop
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
