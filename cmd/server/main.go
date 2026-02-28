// Command server launches the Wyrm authoritative game server.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("config: %v, using defaults", err)
		cfg = &config.Config{
			Server:  config.ServerConfig{Address: "localhost:7777", TickRate: 20},
			World:   config.WorldConfig{Seed: 0, ChunkSize: 512},
			Network: config.NetworkConfig{Protocol: "tcp", Port: 7777},
		}
	}

	world := ecs.NewWorld()
	_ = chunk.NewChunkManager(cfg.World.ChunkSize, cfg.World.Seed)

	srv := network.NewServer(cfg.Server.Address)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server start: %v\n", err)
		os.Exit(1)
	}
	defer srv.Stop()

	log.Printf("server listening on %s (tick_rate=%d)", cfg.Server.Address, cfg.Server.TickRate)

	_ = world

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("shutting down")
}
