//go:build !noebiten

// Command client launches the Wyrm game client with an Ebitengine window.
package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/rendering/raycast"
)

// Game implements the ebiten.Game interface.
type Game struct {
	cfg          *config.Config
	world        *ecs.World
	renderer     *raycast.Renderer
	client       *network.Client
	connected    bool
	playerEntity ecs.Entity
}

// Update advances game state by one tick, processing player input and ECS systems.
func (g *Game) Update() error {
	const dt = 1.0 / 60.0

	// Handle player input
	if g.playerEntity != 0 {
		if comp, ok := g.world.GetComponent(g.playerEntity, "Position"); ok {
			pos := comp.(*components.Position)
			const moveSpeed = 3.0 // units per second
			const turnSpeed = 2.0 // radians per second

			// Movement (WASD or arrow keys)
			if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
				pos.X += math.Cos(pos.Angle) * moveSpeed * dt
				pos.Y += math.Sin(pos.Angle) * moveSpeed * dt
			}
			if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
				pos.X -= math.Cos(pos.Angle) * moveSpeed * dt
				pos.Y -= math.Sin(pos.Angle) * moveSpeed * dt
			}
			if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
				pos.Angle -= turnSpeed * dt
			}
			if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
				pos.Angle += turnSpeed * dt
			}
			// Strafe
			if ebiten.IsKeyPressed(ebiten.KeyQ) {
				pos.X += math.Cos(pos.Angle-math.Pi/2) * moveSpeed * dt
				pos.Y += math.Sin(pos.Angle-math.Pi/2) * moveSpeed * dt
			}
			if ebiten.IsKeyPressed(ebiten.KeyE) {
				pos.X += math.Cos(pos.Angle+math.Pi/2) * moveSpeed * dt
				pos.Y += math.Sin(pos.Angle+math.Pi/2) * moveSpeed * dt
			}
		}
	}

	g.world.Update(dt)
	return nil
}

// Draw renders the current frame using the raycaster and displays debug info.
func (g *Game) Draw(screen *ebiten.Image) {
	// Sync player position to renderer
	if g.playerEntity != 0 {
		if comp, ok := g.world.GetComponent(g.playerEntity, "Position"); ok {
			pos := comp.(*components.Position)
			g.renderer.SetPlayerPos(pos.X, pos.Y, pos.Angle)
		}
	}
	g.renderer.Draw(screen)
	status := "offline"
	if g.connected {
		status = "online"
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Wyrm [%s] %s", g.cfg.Genre, status))
}

// Layout returns the game's logical screen dimensions.
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

	// Create player entity
	player := world.CreateEntity()
	if err := world.AddComponent(player, &components.Position{X: 8.5, Y: 8.5, Z: 0}); err != nil {
		log.Fatalf("failed to add Position component: %v", err)
	}
	if err := world.AddComponent(player, &components.Health{Current: 100, Max: 100}); err != nil {
		log.Fatalf("failed to add Health component: %v", err)
	}

	// Register client-side systems
	world.RegisterSystem(&systems.RenderSystem{PlayerEntity: player})
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
		cfg:          cfg,
		world:        world,
		renderer:     renderer,
		client:       client,
		connected:    connected,
		playerEntity: player,
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
