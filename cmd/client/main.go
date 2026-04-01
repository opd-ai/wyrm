//go:build !noebiten

// Command client launches the Wyrm game client with an Ebitengine window.
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/audio"
	"github.com/opd-ai/wyrm/pkg/audio/ambient"
	"github.com/opd-ai/wyrm/pkg/audio/music"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/input"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/rendering/lighting"
	"github.com/opd-ai/wyrm/pkg/rendering/particles"
	"github.com/opd-ai/wyrm/pkg/rendering/postprocess"
	"github.com/opd-ai/wyrm/pkg/rendering/raycast"
	"github.com/opd-ai/wyrm/pkg/rendering/sprite"
	"github.com/opd-ai/wyrm/pkg/rendering/subtitles"
	"github.com/opd-ai/wyrm/pkg/rendering/texture"
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
	worldMap      [][]int // Local world map for collision detection
	inputManager  *input.Manager
	paused        bool
	// Rendering subpackages
	lightingSystem   *lighting.System
	postprocessPipe  *postprocess.Pipeline
	particleSystem   *particles.System
	particleRenderer *particles.Renderer
	spriteGenerator  *sprite.Generator
	spriteCache      *sprite.SpriteCache
	textureCache     map[string]*texture.Texture
	subtitleSystem   *subtitles.SubtitleSystem
	// Audio subpackages
	ambientMixer  *ambient.Mixer
	adaptiveMusic *music.AdaptiveMusic
	currentRegion ambient.RegionType
	// Interaction system
	interactionSys *InteractionSystem
	// Dialog UI
	dialogUI *DialogUI
	// Combat system
	combatManager *CombatManager
	// Menu system
	menu *Menu
	// State synchronization (online mode)
	stateSync *StateSynchronizer
}

// Update advances game state by one tick, processing player input and ECS systems.
func (g *Game) Update() error {
	const dt = 1.0 / 60.0
	g.syncInputState()
	g.handlePauseToggle()

	// Check if quit requested
	if g.menu != nil && g.menu.QuitRequested() {
		os.Exit(0)
	}

	// Handle menu UI if open
	if g.menu != nil && g.menu.IsOpen() {
		g.menu.Update()
		return nil
	}

	// Handle dialog UI if open
	if g.dialogUI != nil && g.dialogUI.IsOpen() {
		g.dialogUI.Update()
		return nil
	}

	if !g.paused {
		g.handlePlayerInput(dt)
		g.handleInteraction()
		g.handleCombat(dt)
		g.world.Update(dt)
		g.updateChunkMap()
		g.updateRenderingSubsystems(dt)
		g.updateSubtitles(dt)
		// Synchronize state with server (if connected)
		if g.stateSync != nil {
			g.stateSync.Update(dt)
		}
	}
	return nil
}

// handleCombat processes combat mechanics.
func (g *Game) handleCombat(dt float64) {
	if g.combatManager != nil {
		g.combatManager.Update(g.world, dt)
	}
}

// handleInteraction checks for and processes entity interactions.
func (g *Game) handleInteraction() {
	if g.interactionSys == nil {
		return
	}

	// Update interaction system to find current target
	g.interactionSys.Update(g.world)

	// Check if interaction key is pressed
	if g.isActionOrKeyPressed(input.ActionInteract, ebiten.KeyE) {
		target := g.interactionSys.GetCurrentTarget()
		if target != nil {
			g.processInteraction(target)
		}
	}
}

// processInteraction handles the actual interaction with an entity.
func (g *Game) processInteraction(target *InteractionResult) {
	if target == nil {
		return
	}

	switch target.Type {
	case InteractionNPC:
		// Open dialog UI
		if g.dialogUI != nil {
			g.dialogUI.OpenDialog(g.world, target.Entity, target.Name)
		}
	case InteractionItem:
		// Pick up item
		g.pickupItem(target.Entity)
	case InteractionWorkbench:
		// TODO: Open crafting UI (Phase 4B)
		log.Printf("Using workbench: %s", target.Name)
	case InteractionContainer:
		// TODO: Open container UI
		log.Printf("Opening container: %s", target.Name)
	case InteractionDoor:
		// TODO: Open/close door
		log.Printf("Opening door")
	}
}

// pickupItem attempts to pick up an item entity.
func (g *Game) pickupItem(itemEntity ecs.Entity) {
	// Get player inventory
	playerInvComp, ok := g.world.GetComponent(g.playerEntity, "Inventory")
	if !ok {
		return
	}
	playerInv := playerInvComp.(*components.Inventory)

	// Get item inventory (the item itself)
	itemInvComp, ok := g.world.GetComponent(itemEntity, "Inventory")
	if !ok {
		return
	}
	itemInv := itemInvComp.(*components.Inventory)

	// Check capacity
	if len(playerInv.Items)+len(itemInv.Items) > playerInv.Capacity {
		log.Printf("Inventory full!")
		return
	}

	// Transfer items
	for _, item := range itemInv.Items {
		playerInv.Items = append(playerInv.Items, item)
	}

	// Remove item entity from world
	g.world.DestroyEntity(itemEntity)
	log.Printf("Picked up item")
}

// updateRenderingSubsystems updates lighting and particle systems.
func (g *Game) updateRenderingSubsystems(dt float64) {
	// Update lighting based on world clock
	if g.lightingSystem != nil {
		// Advance time-of-day (scaled: 1 real second = 1 game minute)
		g.lightingSystem.AdvanceTime(dt / 60.0)
	}

	// Update particle system
	if g.particleSystem != nil {
		g.particleSystem.Update(dt)
	}
}

// updateSubtitles updates the subtitle system for dialog display.
func (g *Game) updateSubtitles(dt float64) {
	if g.subtitleSystem != nil {
		g.subtitleSystem.Update()
	}
}

// syncInputState synchronizes Ebiten key state with the input manager.
func (g *Game) syncInputState() {
	if g.inputManager == nil {
		return
	}
	// Sync pressed keys
	for _, key := range inpututil.AppendJustPressedKeys(nil) {
		g.inputManager.OnKeyPressed(key.String())
	}
	// Sync released keys
	for _, key := range inpututil.AppendJustReleasedKeys(nil) {
		g.inputManager.OnKeyReleased(key.String())
	}
	// Sync mouse buttons
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.inputManager.OnKeyPressed("MouseButtonLeft")
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.inputManager.OnKeyReleased("MouseButtonLeft")
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.inputManager.OnKeyPressed("MouseButtonRight")
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		g.inputManager.OnKeyReleased("MouseButtonRight")
	}
}

// handlePauseToggle checks for pause action and toggles the menu.
func (g *Game) handlePauseToggle() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.menu != nil {
			g.menu.Toggle()
			g.paused = g.menu.IsOpen()
		} else {
			g.paused = !g.paused
		}
	}
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
	g.worldMap = worldMap // Store for collision detection
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

// canMoveTo checks if a position is valid (not inside a wall).
// Uses player radius for wall sliding behavior.
func (g *Game) canMoveTo(x, y float64) bool {
	if g.worldMap == nil || len(g.worldMap) == 0 {
		return true // No map loaded yet, allow movement
	}
	const playerRadius = 0.3

	// Check multiple points around player position for collision
	for _, offset := range [][2]float64{{0, 0}, {playerRadius, 0}, {-playerRadius, 0}, {0, playerRadius}, {0, -playerRadius}} {
		checkX := x + offset[0]
		checkY := y + offset[1]

		// Convert world position to map coordinates
		mapX := int(checkX)
		mapY := int(checkY)

		// Bounds check
		if mapY < 0 || mapY >= len(g.worldMap) || mapX < 0 || mapX >= len(g.worldMap[0]) {
			return false // Out of bounds is treated as wall
		}

		// Check if cell is a wall (value > 0)
		if g.worldMap[mapY][mapX] > 0 {
			return false
		}
	}
	return true
}

// tryMove attempts to move from current position by delta, with wall sliding.
func (g *Game) tryMove(pos *components.Position, dx, dy float64) {
	// Try full movement first
	newX := pos.X + dx
	newY := pos.Y + dy
	if g.canMoveTo(newX, newY) {
		pos.X = newX
		pos.Y = newY
		return
	}

	// Wall sliding: try X movement only
	if g.canMoveTo(newX, pos.Y) {
		pos.X = newX
		return
	}

	// Wall sliding: try Y movement only
	if g.canMoveTo(pos.X, newY) {
		pos.Y = newY
		return
	}
	// Movement blocked in both directions
}

// processMovementInput handles forward/back movement and turning.
func (g *Game) processMovementInput(pos *components.Position, dt float64) {
	const moveSpeed = 3.0 // units per second
	const turnSpeed = 2.0 // radians per second

	// Use input manager if available, fallback to direct key checks
	moveForward := g.isActionOrKeyPressed(input.ActionMoveForward, ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)
	moveBackward := g.isActionOrKeyPressed(input.ActionMoveBackward, ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)
	turnLeft := g.isActionOrKeyPressed(input.ActionMoveLeft, ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft)
	turnRight := g.isActionOrKeyPressed(input.ActionMoveRight, ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight)

	if moveForward {
		dx := math.Cos(pos.Angle) * moveSpeed * dt
		dy := math.Sin(pos.Angle) * moveSpeed * dt
		g.tryMove(pos, dx, dy)
	}
	if moveBackward {
		dx := -math.Cos(pos.Angle) * moveSpeed * dt
		dy := -math.Sin(pos.Angle) * moveSpeed * dt
		g.tryMove(pos, dx, dy)
	}
	if turnLeft {
		pos.Angle -= turnSpeed * dt
	}
	if turnRight {
		pos.Angle += turnSpeed * dt
	}
}

// processStrafeInput handles left/right strafe movement.
func (g *Game) processStrafeInput(pos *components.Position, dt float64) {
	const moveSpeed = 3.0
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		dx := math.Cos(pos.Angle-math.Pi/2) * moveSpeed * dt
		dy := math.Sin(pos.Angle-math.Pi/2) * moveSpeed * dt
		g.tryMove(pos, dx, dy)
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		dx := math.Cos(pos.Angle+math.Pi/2) * moveSpeed * dt
		dy := math.Sin(pos.Angle+math.Pi/2) * moveSpeed * dt
		g.tryMove(pos, dx, dy)
	}
}

// isActionOrKeyPressed checks if an action is pressed via input manager or fallback key.
func (g *Game) isActionOrKeyPressed(action input.Action, fallbackKey ebiten.Key) bool {
	if g.inputManager != nil {
		return g.inputManager.IsActionPressed(action)
	}
	return ebiten.IsKeyPressed(fallbackKey)
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

	// Render base scene
	g.renderer.Draw(screen)

	// Apply post-processing effects (genre-specific)
	g.applyPostProcessing(screen)

	// Draw particles (weather effects) on top
	g.drawParticles(screen)

	// Apply combat visual feedback (screen shake, damage flash)
	g.applyCombatVisualFeedback(screen)

	// Draw HUD last (on top of everything)
	g.drawHUD(screen)

	// Draw dialog UI overlay if active
	if g.dialogUI != nil && g.dialogUI.IsOpen() {
		g.dialogUI.Draw(screen)
	}

	// Draw menu overlay if active
	if g.menu != nil && g.menu.IsOpen() {
		g.menu.Draw(screen)
	}

	// Draw death screen if dead
	g.drawDeathScreen(screen)
}

// applyCombatVisualFeedback applies screen shake and damage flash effects.
func (g *Game) applyCombatVisualFeedback(screen *ebiten.Image) {
	if g.combatManager == nil {
		return
	}

	// Apply damage flash overlay
	flashAlpha := g.combatManager.GetDamageFlashAlpha()
	if flashAlpha > 0 {
		bounds := screen.Bounds()
		flashColor := color.RGBA{R: 255, G: 0, B: 0, A: uint8(flashAlpha)}
		for y := 0; y < bounds.Dy(); y++ {
			for x := 0; x < bounds.Dx(); x++ {
				// Blend red flash over existing pixels
				existing := screen.At(x, y)
				r, g, b, a := existing.RGBA()
				blendFactor := float64(flashAlpha) / 255.0
				newR := uint8((float64(r>>8)*(1-blendFactor) + float64(flashColor.R)*blendFactor))
				newG := uint8((float64(g>>8)*(1-blendFactor) + float64(flashColor.G)*blendFactor))
				newB := uint8((float64(b>>8)*(1-blendFactor) + float64(flashColor.B)*blendFactor))
				screen.Set(x, y, color.RGBA{R: newR, G: newG, B: newB, A: uint8(a >> 8)})
			}
		}
	}
}

// drawDeathScreen draws the death overlay when the player is dead.
func (g *Game) drawDeathScreen(screen *ebiten.Image) {
	if g.combatManager == nil || !g.combatManager.IsDead() {
		return
	}

	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()

	// Draw dark overlay
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	for y := 0; y < screenHeight; y++ {
		for x := 0; x < screenWidth; x++ {
			screen.Set(x, y, overlayColor)
		}
	}

	// Draw death message
	respawnTime := g.combatManager.GetTimeUntilRespawn()
	deathText := "YOU DIED"
	respawnText := fmt.Sprintf("Respawning in %.1f...", respawnTime)

	// Center the text
	deathX := (screenWidth - len(deathText)*6) / 2
	deathY := screenHeight/2 - 20
	respawnX := (screenWidth - len(respawnText)*6) / 2
	respawnY := screenHeight/2 + 10

	ebitenutil.DebugPrintAt(screen, deathText, deathX, deathY)
	ebitenutil.DebugPrintAt(screen, respawnText, respawnX, respawnY)
}

// applyPostProcessing applies genre-specific visual effects to the rendered frame.
func (g *Game) applyPostProcessing(screen *ebiten.Image) {
	if g.postprocessPipe == nil {
		return
	}

	// Get screen as RGBA image
	bounds := screen.Bounds()
	rgba := image.NewRGBA(bounds)

	// Copy screen pixels to RGBA
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, screen.At(x, y))
		}
	}

	// Apply post-processing pipeline
	processed := g.postprocessPipe.Apply(rgba)

	// Copy processed image back to screen
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			screen.Set(x, y, processed.At(x, y))
		}
	}
}

// drawParticles renders weather particles to the screen.
func (g *Game) drawParticles(screen *ebiten.Image) {
	if g.particleSystem == nil || g.particleRenderer == nil {
		return
	}

	// Get screen pixels for particle rendering
	bounds := screen.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixels := make([]byte, width*height*4)

	// Copy current screen to pixel buffer for blending
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, gr, b, a := screen.At(x, y).RGBA()
			idx := (y*width + x) * 4
			pixels[idx] = uint8(r >> 8)
			pixels[idx+1] = uint8(gr >> 8)
			pixels[idx+2] = uint8(b >> 8)
			pixels[idx+3] = uint8(a >> 8)
		}
	}

	// Render particles
	g.particleRenderer.Draw(g.particleSystem, pixels)

	// Copy back to screen
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			screen.Set(x, y, color.RGBA{
				R: pixels[idx],
				G: pixels[idx+1],
				B: pixels[idx+2],
				A: pixels[idx+3],
			})
		}
	}
}

// drawHUD renders the heads-up display elements.
func (g *Game) drawHUD(screen *ebiten.Image) {
	screenWidth := g.cfg.Window.Width
	screenHeight := g.cfg.Window.Height

	// Get player components
	var pos *components.Position
	var health *components.Health
	var mana *components.Mana
	if g.playerEntity != 0 {
		if comp, ok := g.world.GetComponent(g.playerEntity, "Position"); ok {
			pos = comp.(*components.Position)
		}
		if comp, ok := g.world.GetComponent(g.playerEntity, "Health"); ok {
			health = comp.(*components.Health)
		}
		if comp, ok := g.world.GetComponent(g.playerEntity, "Mana"); ok {
			mana = comp.(*components.Mana)
		}
	}

	// Draw health bar (bottom-left)
	barWidth := 150
	barHeight := 12
	barX := 10
	barY := screenHeight - 50
	if health != nil {
		healthPercent := health.Current / health.Max
		g.drawBar(screen, barX, barY, barWidth, barHeight, healthPercent, 0xCC0000FF, 0x440000FF)
	}

	// Draw mana bar (below health)
	if mana != nil {
		manaPercent := mana.Current / mana.Max
		g.drawBar(screen, barX, barY+16, barWidth, barHeight, manaPercent, 0x0066CCFF, 0x002244FF)
	}

	// Draw position and compass (top-left)
	status := "offline"
	if g.connected {
		status = "online"
	}
	coordText := fmt.Sprintf("Wyrm [%s] %s", g.cfg.Genre, status)
	if pos != nil {
		chunkX := int(pos.X) / g.cfg.World.ChunkSize
		chunkY := int(pos.Y) / g.cfg.World.ChunkSize
		direction := getCompassDirection(pos.Angle)
		coordText = fmt.Sprintf("Wyrm [%s] %s\nPos: %.1f, %.1f | Chunk: %d, %d\nHeading: %s",
			g.cfg.Genre, status, pos.X, pos.Y, chunkX, chunkY, direction)
	}
	ebitenutil.DebugPrint(screen, coordText)

	// Draw minimap (top-right)
	g.drawMinimap(screen, screenWidth-80, 10, 64)

	// Draw interaction prompt (bottom-center)
	g.drawInteractionPrompt(screen)
}

// drawInteractionPrompt displays the current interaction target prompt.
func (g *Game) drawInteractionPrompt(screen *ebiten.Image) {
	if g.interactionSys == nil {
		return
	}

	target := g.interactionSys.GetCurrentTarget()
	if target == nil || target.Prompt == "" {
		return
	}

	// Draw prompt at bottom center of screen
	screenWidth := g.cfg.Window.Width
	screenHeight := g.cfg.Window.Height

	// Estimate text width (rough approximation: 6 pixels per character)
	textWidth := len(target.Prompt) * 6
	x := (screenWidth - textWidth) / 2
	y := screenHeight - 100

	ebitenutil.DebugPrintAt(screen, target.Prompt, x, y)
}

// drawBar renders a horizontal bar (health/mana style).
func (g *Game) drawBar(screen *ebiten.Image, x, y, width, height int, percent float64, fillColor, bgColor uint32) {
	// Background
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			if px >= 0 && px < g.cfg.Window.Width && py >= 0 && py < g.cfg.Window.Height {
				screen.Set(px, py, uint32ToColor(bgColor))
			}
		}
	}
	// Fill
	fillWidth := int(float64(width) * percent)
	for py := y + 1; py < y+height-1; py++ {
		for px := x + 1; px < x+fillWidth-1; px++ {
			if px >= 0 && px < g.cfg.Window.Width && py >= 0 && py < g.cfg.Window.Height {
				screen.Set(px, py, uint32ToColor(fillColor))
			}
		}
	}
}

// drawMinimap renders a small top-down view of the nearby area.
func (g *Game) drawMinimap(screen *ebiten.Image, x, y, size int) {
	if g.worldMap == nil || len(g.worldMap) == 0 {
		return
	}

	// Get player position
	var playerMapX, playerMapY int
	if g.playerEntity != 0 {
		if comp, ok := g.world.GetComponent(g.playerEntity, "Position"); ok {
			pos := comp.(*components.Position)
			playerMapX = int(pos.X)
			playerMapY = int(pos.Y)
		}
	}

	// Draw minimap background and terrain
	mapRadius := 16 // Show 16 cells in each direction
	for my := 0; my < size; my++ {
		for mx := 0; mx < size; mx++ {
			// Map screen coords to world coords
			worldX := playerMapX - mapRadius + (mx * mapRadius * 2 / size)
			worldY := playerMapY - mapRadius + (my * mapRadius * 2 / size)

			screenX := x + mx
			screenY := y + my

			if screenX < 0 || screenX >= g.cfg.Window.Width || screenY < 0 || screenY >= g.cfg.Window.Height {
				continue
			}

			// Check bounds and draw
			if worldY >= 0 && worldY < len(g.worldMap) && worldX >= 0 && worldX < len(g.worldMap[0]) {
				if g.worldMap[worldY][worldX] > 0 {
					screen.Set(screenX, screenY, uint32ToColor(0x666666FF)) // Wall
				} else {
					screen.Set(screenX, screenY, uint32ToColor(0x333333FF)) // Floor
				}
			} else {
				screen.Set(screenX, screenY, uint32ToColor(0x111111FF)) // Out of bounds
			}
		}
	}

	// Draw player dot in center
	centerX := x + size/2
	centerY := y + size/2
	screen.Set(centerX, centerY, uint32ToColor(0x00FF00FF))
	screen.Set(centerX+1, centerY, uint32ToColor(0x00FF00FF))
	screen.Set(centerX-1, centerY, uint32ToColor(0x00FF00FF))
	screen.Set(centerX, centerY+1, uint32ToColor(0x00FF00FF))
	screen.Set(centerX, centerY-1, uint32ToColor(0x00FF00FF))
}

// getCompassDirection returns cardinal/ordinal direction name from angle.
func getCompassDirection(angle float64) string {
	// Normalize angle to 0-2π
	for angle < 0 {
		angle += 2 * math.Pi
	}
	for angle >= 2*math.Pi {
		angle -= 2 * math.Pi
	}

	// Convert to compass direction (0 = East in math coordinates)
	directions := []string{"E", "NE", "N", "NW", "W", "SW", "S", "SE"}
	index := int((angle+(math.Pi/8))/(math.Pi/4)) % 8
	return directions[index]
}

// uint32ToColor converts a hex color (RRGGBBAA) to color.RGBA.
func uint32ToColor(c uint32) color.RGBA {
	return color.RGBA{
		R: uint8((c >> 24) & 0xFF),
		G: uint8((c >> 16) & 0xFF),
		B: uint8((c >> 8) & 0xFF),
		A: uint8(c & 0xFF),
	}
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
	client, connected := connectToServer(cfg)
	// Register client systems; in offline mode (!connected), also register game logic systems
	registerClientSystems(world, player, cfg, !connected)
	chunkMgr := chunk.NewManager(cfg.World.ChunkSize, cfg.World.Seed)

	// Initialize input manager with config bindings
	inputMgr := input.NewManager()
	inputMgr.LoadFromConfig(&cfg.KeyBindings)

	// Initialize audio system
	audioEngine := audio.NewEngine(cfg.Genre)
	audioPlayer, audioErr := audio.NewPlayer()
	if audioErr != nil {
		log.Printf("audio initialization failed (continuing without audio): %v", audioErr)
	}

	// Initialize audio subpackages (ambient soundscapes and adaptive music)
	ambientMix := ambient.NewMixer(cfg.Genre, cfg.World.Seed)
	adaptiveMusic := music.NewAdaptiveMusic(cfg.Genre, cfg.World.Seed)

	// Initialize rendering subpackages
	lightingSys := lighting.NewSystem(cfg.Genre)
	postprocessPipeline := postprocess.NewPipeline(cfg.Genre)
	particleSys := particles.NewSystem(cfg.World.Seed)
	particleRend := particles.NewRenderer(cfg.Window.Width, cfg.Window.Height)
	spriteGen := sprite.NewGenerator(cfg.Genre, cfg.World.Seed)
	spriteCache := sprite.NewSpriteCache(100, 64*1024*1024) // 100 sheets, 64MB max
	textureCache := make(map[string]*texture.Texture)
	subtitleSys := subtitles.NewSubtitleSystem(true)

	// Pre-generate some common textures for the genre
	textureCache["wall"] = texture.GenerateWithSeed(64, 64, cfg.World.Seed, cfg.Genre)
	textureCache["floor"] = texture.GenerateWithSeed(64, 64, cfg.World.Seed+1, cfg.Genre)
	textureCache["ceiling"] = texture.GenerateWithSeed(64, 64, cfg.World.Seed+2, cfg.Genre)

	// Set up weather particles based on genre
	setupWeatherParticles(particleSys, cfg.Genre, cfg.World.Seed)

	// Local map size for raycaster (must be divisible by 3 for 3x3 chunk window)
	const localMapSize = 48
	const wallThreshold = 0.5

	// Initialize interaction system
	interactionSys := NewInteractionSystem(player, 2.0) // 2.0 units max interaction range

	// Initialize dialog UI
	dialogUI := NewDialogUI(cfg.Genre, player)

	// Initialize combat manager
	combatMgr := NewCombatManager(player, inputMgr)

	// Initialize menu system
	menu := NewMenu(cfg, inputMgr)

	game := &Game{
		cfg:              cfg,
		world:            world,
		renderer:         renderer,
		client:           client,
		connected:        connected,
		playerEntity:     player,
		chunkManager:     chunkMgr,
		lastChunkX:       -999, // Force initial map build
		lastChunkY:       -999,
		mapSize:          localMapSize,
		wallThreshold:    wallThreshold,
		audioEngine:      audioEngine,
		audioPlayer:      audioPlayer,
		inputManager:     inputMgr,
		lightingSystem:   lightingSys,
		postprocessPipe:  postprocessPipeline,
		particleSystem:   particleSys,
		particleRenderer: particleRend,
		spriteGenerator:  spriteGen,
		spriteCache:      spriteCache,
		textureCache:     textureCache,
		subtitleSystem:   subtitleSys,
		ambientMixer:     ambientMix,
		adaptiveMusic:    adaptiveMusic,
		currentRegion:    ambient.RegionPlains, // Default starting region
		interactionSys:   interactionSys,
		dialogUI:         dialogUI,
		combatManager:    combatMgr,
		menu:             menu,
	}

	// Initialize state synchronization if connected to server
	if connected {
		game.stateSync = NewStateSynchronizer(client, world, player)
		game.stateSync.Start()
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
	if err := world.AddComponent(player, &components.Mana{Current: 50, Max: 50, RegenRate: 1.0}); err != nil {
		log.Fatalf("failed to add Mana component: %v", err)
	}
	if err := world.AddComponent(player, &components.Skills{Levels: make(map[string]int), Experience: make(map[string]float64), SchoolBonuses: make(map[string]float64)}); err != nil {
		log.Fatalf("failed to add Skills component: %v", err)
	}
	if err := world.AddComponent(player, &components.Inventory{Items: []string{}, Capacity: 30}); err != nil {
		log.Fatalf("failed to add Inventory component: %v", err)
	}
	if err := world.AddComponent(player, &components.Faction{ID: "player", Reputation: 0}); err != nil {
		log.Fatalf("failed to add Faction component: %v", err)
	}
	if err := world.AddComponent(player, &components.Reputation{Standings: make(map[string]float64)}); err != nil {
		log.Fatalf("failed to add Reputation component: %v", err)
	}
	if err := world.AddComponent(player, &components.Stealth{Visibility: 1.0, BaseVisibility: 1.0, SneakVisibility: 0.3, DetectionRadius: 15.0}); err != nil {
		log.Fatalf("failed to add Stealth component: %v", err)
	}
	if err := world.AddComponent(player, &components.CombatState{}); err != nil {
		log.Fatalf("failed to add CombatState component: %v", err)
	}
	if err := world.AddComponent(player, &components.AudioListener{Volume: 1.0, Enabled: true}); err != nil {
		log.Fatalf("failed to add AudioListener component: %v", err)
	}
	if err := world.AddComponent(player, &components.Weapon{Name: "Fists", Damage: 5, Range: 1.5, AttackSpeed: 1.0, WeaponType: "melee"}); err != nil {
		log.Fatalf("failed to add Weapon component: %v", err)
	}
	return player
}

// registerClientSystems registers all client-side ECS systems.
// In single-player mode (offline), this also registers core game logic systems.
func registerClientSystems(world *ecs.World, player ecs.Entity, cfg *config.Config, offline bool) {
	// Rendering and audio systems (always needed)
	world.RegisterSystem(&systems.RenderSystem{PlayerEntity: player})
	world.RegisterSystem(&systems.AudioSystem{Genre: cfg.Genre})
	weatherSys := systems.NewWeatherSystem(cfg.Genre, 300.0)
	world.RegisterSystem(weatherSys)

	// In offline mode, register essential gameplay systems for single-player
	if offline {
		registerSinglePlayerSystems(world, cfg, weatherSys)
	}
}

// registerSinglePlayerSystems registers game logic systems for offline/single-player mode.
// These systems provide the core RPG mechanics when not connected to a server.
func registerSinglePlayerSystems(world *ecs.World, cfg *config.Config, weatherSys *systems.WeatherSystem) {
	seed := cfg.World.Seed
	genre := cfg.Genre

	// World time (drives NPC schedules, shop hours, etc.)
	world.RegisterSystem(systems.NewWorldClockSystem(60.0))

	// NPC behavior systems
	world.RegisterSystem(&systems.NPCScheduleSystem{})
	world.RegisterSystem(systems.NewNPCPathfindingSystem())
	world.RegisterSystem(systems.NewNPCNeedsSystem())
	world.RegisterSystem(systems.NewNPCOccupationSystem(seed))
	world.RegisterSystem(systems.NewEmotionalStateSystem())
	world.RegisterSystem(systems.NewNPCMemorySystem())
	world.RegisterSystem(systems.NewGossipSystem())

	// Faction systems
	fps := systems.NewFactionPoliticsSystem(0.1)
	world.RegisterSystem(fps)
	factionRankSystem := systems.NewFactionRankSystem(genre)
	world.RegisterSystem(factionRankSystem)
	world.RegisterSystem(systems.NewFactionCoupSystem(factionRankSystem, fps, seed, genre))
	world.RegisterSystem(systems.NewFactionExclusiveContentSystem(factionRankSystem, genre))
	world.RegisterSystem(systems.NewDynamicFactionWarSystem(fps))

	// Crime and law systems
	crimeSystem := systems.NewCrimeSystem(60.0, 100.0)
	world.RegisterSystem(crimeSystem)
	guardPursuitSystem := systems.NewGuardPursuitSystem(crimeSystem)
	world.RegisterSystem(guardPursuitSystem)
	world.RegisterSystem(systems.NewBriberySystem(crimeSystem, guardPursuitSystem, seed))
	crimeEvidenceSystem := systems.NewCrimeEvidenceSystem(crimeSystem, genre, seed)
	world.RegisterSystem(crimeEvidenceSystem)
	world.RegisterSystem(systems.NewPardonSystem(crimeSystem, crimeEvidenceSystem, genre, seed))
	world.RegisterSystem(systems.NewCriminalFactionQuestSystem(factionRankSystem, genre, seed))

	// Economy systems
	economySystem := systems.NewEconomySystem(0.5, 0.1)
	world.RegisterSystem(economySystem)
	world.RegisterSystem(systems.NewEconomicEventSystem(seed, genre, economySystem))
	world.RegisterSystem(systems.NewMarketManipulationSystem(seed, genre, economySystem))
	world.RegisterSystem(systems.NewTradeRouteSystem(seed, genre, economySystem))
	world.RegisterSystem(systems.NewInvestmentSystem(seed, genre))
	world.RegisterSystem(systems.NewPlayerShopSystem(economySystem))
	world.RegisterSystem(systems.NewCityBuildingSystem(genre, seed))
	world.RegisterSystem(systems.NewCityEventSystem(genre, seed))
	world.RegisterSystem(systems.NewTradingSystem())

	// Combat systems
	world.RegisterSystem(systems.NewCombatSystem())
	world.RegisterSystem(systems.NewMagicSystem())
	world.RegisterSystem(systems.NewProjectileSystem())
	world.RegisterSystem(systems.NewStealthSystem())
	world.RegisterSystem(systems.NewDistractionSystem())
	world.RegisterSystem(systems.NewHidingSpotSystem(float64(cfg.World.ChunkSize)))

	// Vehicle systems
	world.RegisterSystem(&systems.VehicleSystem{})
	world.RegisterSystem(systems.NewVehiclePhysicsSystem(genre))
	world.RegisterSystem(systems.NewVehicleCombatSystem())
	world.RegisterSystem(systems.NewFlyingVehicleSystem(genre))
	world.RegisterSystem(systems.NewNavalVehicleSystem(genre))
	world.RegisterSystem(systems.NewMountSystem(seed, genre))

	// Quest system
	world.RegisterSystem(systems.NewQuestSystem())

	// Skills and crafting systems
	skillRegistry := systems.NewSkillRegistry()
	skillProgressionSystem := systems.NewSkillProgressionSystem(100.0, 100)
	world.RegisterSystem(skillProgressionSystem)
	world.RegisterSystem(systems.NewSkillBookSystem(skillRegistry, skillProgressionSystem))
	world.RegisterSystem(systems.NewSkillSynergySystem(skillRegistry))
	world.RegisterSystem(systems.NewActionUnlockSystem(skillRegistry, skillProgressionSystem))
	world.RegisterSystem(systems.NewNPCTrainingSystem(skillRegistry, skillProgressionSystem))
	world.RegisterSystem(systems.NewCraftingSystem(seed))

	// Dialog and social systems
	world.RegisterSystem(systems.NewDialogConsequenceSystem())
	world.RegisterSystem(systems.NewMultiNPCConversationSystem())
	world.RegisterSystem(systems.NewPartySystem())
	world.RegisterSystem(systems.NewVehicleCustomizationSystem(seed, genre))

	// Environment systems
	world.RegisterSystem(systems.NewIndoorOutdoorSystem(weatherSys))
	world.RegisterSystem(systems.NewHazardSystem(genre))
}

// startAmbientAudio initializes and plays genre-appropriate ambient audio.
func (g *Game) startAmbientAudio() {
	if g.audioPlayer == nil {
		return
	}

	// Use the ambient soundscape mixer if available
	if g.ambientMixer != nil {
		duration := 2.0 // seconds
		samples := g.ambientMixer.GenerateMixedSamples(duration)
		// Reduce volume for ambient background
		for i := range samples {
			samples[i] *= 0.15
		}
		g.audioPlayer.QueueSamples(samples)
		g.audioPlayer.Play()
		return
	}

	// Fallback to simple sine wave if ambient mixer not available
	if g.audioEngine == nil {
		return
	}
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

// setupWeatherParticles creates genre-appropriate weather particle emitters.
func setupWeatherParticles(sys *particles.System, genre string, seed int64) {
	if sys == nil {
		return
	}

	// Genre-specific ambient particles
	var weatherType string
	var intensity float64

	switch genre {
	case "fantasy":
		// Light dust motes in sunbeams
		weatherType = particles.TypeDust
		intensity = 0.3
	case "sci-fi":
		// Subtle atmospheric particles
		weatherType = particles.TypeDust
		intensity = 0.2
	case "horror":
		// Ash/fog wisps
		weatherType = particles.TypeAsh
		intensity = 0.5
	case "cyberpunk":
		// Rain is iconic for cyberpunk
		weatherType = particles.TypeRain
		intensity = 0.4
	case "post-apocalyptic":
		// Dust and ash
		weatherType = particles.TypeDust
		intensity = 0.6
	default:
		return
	}

	// Create weather emitters using preset
	preset := &particles.WeatherPreset{
		Type:      weatherType,
		Intensity: intensity,
		Direction: 1.57, // Straight down
	}
	emitters := particles.CreateWeatherEmitters(preset, seed)
	for _, e := range emitters {
		sys.AddEmitter(e)
	}
}
