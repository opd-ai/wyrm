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
	"github.com/opd-ai/wyrm/pkg/audio"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/rendering/raycast"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

// Game implements the ebiten.Game interface.
type Game struct {
	cfg           *config.Config
	world         *ecs.World
	renderer      *raycast.Renderer
	client        *network.Client
	connected     bool
	playerEntity  ecs.Entity
	chunkManager  *chunk.Manager
	lastChunkX    int
	lastChunkY    int
	mapSize       int // Size of the local map grid
	wallThreshold float64
	audioEngine   *audio.Engine
	audioPlayer   *audio.Player
}

// Update advances game state by one tick, processing player input and ECS systems.
func (g *Game) Update() error {
	const dt = 1.0 / 60.0
	g.handlePlayerInput(dt)
	g.world.Update(dt)
	g.updateChunkMap()
	return nil
}

// updateChunkMap refreshes the world map when player moves to a new chunk.
func (g *Game) updateChunkMap() {
	if g.playerEntity == 0 || g.chunkManager == nil {
		return
	}
	comp, ok := g.world.GetComponent(g.playerEntity, "Position")
	if !ok {
		return
	}
	pos := comp.(*components.Position)
	// Calculate current chunk coordinates
	chunkX := int(pos.X) / g.cfg.World.ChunkSize
	chunkY := int(pos.Y) / g.cfg.World.ChunkSize
	// Only rebuild map when entering a new chunk
	if chunkX == g.lastChunkX && chunkY == g.lastChunkY {
		return
	}
	g.lastChunkX = chunkX
	g.lastChunkY = chunkY
	g.rebuildWorldMap(chunkX, chunkY)
}

// rebuildWorldMap constructs the local world map from surrounding chunks.
func (g *Game) rebuildWorldMap(centerChunkX, centerChunkY int) {
	worldMap := make([][]int, g.mapSize)
	for i := range worldMap {
		worldMap[i] = make([]int, g.mapSize)
	}
	chunkSize := g.cfg.World.ChunkSize
	// Load 3x3 chunk window and sample into local map
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			c := g.chunkManager.GetChunk(centerChunkX+dx, centerChunkY+dy)
			g.sampleChunkIntoMap(worldMap, c, dx, dy, chunkSize)
		}
	}
	g.renderer.SetWorldMapDirect(worldMap)
}

// sampleChunkIntoMap samples a chunk's heightmap into the local world map.
func (g *Game) sampleChunkIntoMap(worldMap [][]int, c *chunk.Chunk, dx, dy, chunkSize int) {
	// Each chunk maps to a section of the local map
	// Local map is divided into 3x3 sections for the 3x3 chunk window
	sectionSize := g.mapSize / 3
	startX := (dx + 1) * sectionSize
	startY := (dy + 1) * sectionSize
	// Sample from chunk heightmap into local map section
	for ly := 0; ly < sectionSize; ly++ {
		for lx := 0; lx < sectionSize; lx++ {
			// Map local coords to chunk coords
			cx := lx * chunkSize / sectionSize
			cy := ly * chunkSize / sectionSize
			height := c.GetHeight(cx, cy)
			worldMap[startY+ly][startX+lx] = heightToWallType(height, g.wallThreshold)
		}
	}
}

// heightToWallType converts a height value to a wall type.
func heightToWallType(height, threshold float64) int {
	if height <= threshold {
		return 0 // No wall
	}
	if height > 0.8 {
		return 3 // High wall (blue)
	}
	if height > 0.6 {
		return 2 // Medium wall (green)
	}
	return 1 // Low wall (red-brown)
}

// handlePlayerInput processes keyboard input for player movement.
func (g *Game) handlePlayerInput(dt float64) {
	if g.playerEntity == 0 {
		return
	}
	comp, ok := g.world.GetComponent(g.playerEntity, "Position")
	if !ok {
		return
	}
	pos := comp.(*components.Position)
	g.processMovementInput(pos, dt)
	g.processStrafeInput(pos, dt)
}

// processMovementInput handles forward/back movement and turning.
func (g *Game) processMovementInput(pos *components.Position, dt float64) {
	const moveSpeed = 3.0 // units per second
	const turnSpeed = 2.0 // radians per second

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
}

// processStrafeInput handles left/right strafe movement.
func (g *Game) processStrafeInput(pos *components.Position, dt float64) {
	const moveSpeed = 3.0
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		pos.X += math.Cos(pos.Angle-math.Pi/2) * moveSpeed * dt
		pos.Y += math.Sin(pos.Angle-math.Pi/2) * moveSpeed * dt
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		pos.X += math.Cos(pos.Angle+math.Pi/2) * moveSpeed * dt
		pos.Y += math.Sin(pos.Angle+math.Pi/2) * moveSpeed * dt
	}
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
	player := createPlayerEntity(world)
	registerClientSystems(world, player, cfg)
	client, connected := connectToServer(cfg)
	chunkMgr := chunk.NewManager(cfg.World.ChunkSize, cfg.World.Seed)

	// Initialize audio system
	audioEngine := audio.NewEngine(cfg.Genre)
	audioPlayer, audioErr := audio.NewPlayer()
	if audioErr != nil {
		log.Printf("audio initialization failed (continuing without audio): %v", audioErr)
	}

	// Local map size for raycaster (must be divisible by 3 for 3x3 chunk window)
	const localMapSize = 48
	const wallThreshold = 0.5

	game := &Game{
		cfg:           cfg,
		world:         world,
		renderer:      renderer,
		client:        client,
		connected:     connected,
		playerEntity:  player,
		chunkManager:  chunkMgr,
		lastChunkX:    -999, // Force initial map build
		lastChunkY:    -999,
		mapSize:       localMapSize,
		wallThreshold: wallThreshold,
		audioEngine:   audioEngine,
		audioPlayer:   audioPlayer,
	}

	// Start ambient audio if available
	if audioPlayer != nil {
		game.startAmbientAudio()
	}

	ebiten.SetWindowSize(cfg.Window.Width, cfg.Window.Height)
	ebiten.SetWindowTitle(cfg.Window.Title)

	runGame(game, connected, client)
}

// createPlayerEntity creates and configures the player entity.
func createPlayerEntity(world *ecs.World) ecs.Entity {
	player := world.CreateEntity()
	if err := world.AddComponent(player, &components.Position{X: 8.5, Y: 8.5, Z: 0}); err != nil {
		log.Fatalf("failed to add Position component: %v", err)
	}
	if err := world.AddComponent(player, &components.Health{Current: 100, Max: 100}); err != nil {
		log.Fatalf("failed to add Health component: %v", err)
	}
	return player
}

// registerClientSystems registers all client-side ECS systems.
func registerClientSystems(world *ecs.World, player ecs.Entity, cfg *config.Config) {
	world.RegisterSystem(&systems.RenderSystem{PlayerEntity: player})
	world.RegisterSystem(&systems.AudioSystem{Genre: cfg.Genre})
	world.RegisterSystem(&systems.WeatherSystem{})
}

// startAmbientAudio initializes and plays genre-appropriate ambient audio.
func (g *Game) startAmbientAudio() {
	if g.audioEngine == nil || g.audioPlayer == nil {
		return
	}
	// Generate a simple ambient tone based on genre
	freq := g.audioEngine.GetGenreBaseFrequency()
	duration := 2.0 // seconds
	samples := g.audioEngine.GenerateSineWave(freq, duration)
	// Apply a gentle ADSR envelope for smooth ambient sound
	samples = g.audioEngine.ApplyADSR(samples, 0.5, 0.2, 0.3, 0.5)
	// Reduce volume for ambient background
	for i := range samples {
		samples[i] *= 0.1
	}
	g.audioPlayer.QueueSamples(samples)
	g.audioPlayer.Play()
}

// connectToServer attempts to connect to the game server.
func connectToServer(cfg *config.Config) (*network.Client, bool) {
	client := network.NewClient(cfg.Server.Address)
	if err := client.Connect(); err != nil {
		log.Printf("running in offline mode: %v", err)
		return client, false
	}
	log.Printf("connected to server at %s", cfg.Server.Address)
	return client, true
}

// runGame starts the game loop and handles cleanup.
func runGame(game *Game, connected bool, client *network.Client) {
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
