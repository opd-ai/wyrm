// Command client launches the Wyrm game client with an Ebitengine window.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/rendering/raycast"
)

// Game implements the ebiten.Game interface.
type Game struct {
	cfg       *config.Config
	world     *ecs.World
	renderer  *raycast.Renderer
	client    *network.Client
	connected bool
}

func (g *Game) Update() error {
	const dt = 1.0 / 60.0
	g.world.Update(dt)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen)
	status := "offline"
	if g.connected {
		status = "online"
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Wyrm [%s] %s", g.cfg.Genre, status))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.cfg.Window.Width, g.cfg.Window.Height
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	world := ecs.NewWorld()
	renderer := raycast.NewRenderer(cfg.Window.Width, cfg.Window.Height)

	// Register client-side systems
	world.RegisterSystem(&systems.RenderSystem{})
	world.RegisterSystem(&systems.AudioSystem{Genre: cfg.Genre})
	world.RegisterSystem(&systems.WeatherSystem{})

	// Create network client
	client := network.NewClient(cfg.Server.Address)
	connected := false

	// Attempt to connect (non-blocking, graceful failure for offline mode)
	if err := client.Connect(); err != nil {
		log.Printf("running in offline mode: %v", err)
	} else {
		connected = true
		log.Printf("connected to server at %s", cfg.Server.Address)
	}

	game := &Game{
		cfg:       cfg,
		world:     world,
		renderer:  renderer,
		client:    client,
		connected: connected,
	}

	ebiten.SetWindowSize(cfg.Window.Width, cfg.Window.Height)
	ebiten.SetWindowTitle(cfg.Window.Title)

	if err := ebiten.RunGame(game); err != nil {
		if connected {
			client.Disconnect()
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if connected {
		client.Disconnect()
	}
}
