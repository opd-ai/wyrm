//go:build !noebiten

// Command client launches the Wyrm game client with an Ebitengine window.
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

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

// GameState represents the current state of the game.
type GameState int

const (
	GameStateCharacterCreation GameState = iota
	GameStatePlaying
)

// Movement constants for player physics.
const (
	playerMoveSpeed = 3.0 // units per second
	playerTurnSpeed = 2.0 // radians per second
	playerRadius    = 0.3 // collision radius
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
	lodManager    *chunk.LODManager // LOD management for distant terrain optimization
	lastChunkX    int
	lastChunkY    int
	mapSize       int // Size of the local map grid
	wallThreshold float64
	audioEngine   *audio.Engine
	audioPlayer   *audio.Player
	worldMap      [][]int // Local world map for collision detection
	inputManager  *input.Manager
	paused        bool
	// Game state
	gameState         GameState
	characterCreation *CharacterCreation
	// Mount/Vehicle state
	isRidingMount        bool
	currentMountEntity   ecs.Entity
	isInVehicle          bool
	currentVehicleEntity ecs.Entity
	// Rendering subpackages
	lightingSystem   *lighting.System
	postprocessPipe  *postprocess.Pipeline
	particleSystem   *particles.System
	particleRenderer *particles.Renderer
	spriteGenerator  *sprite.Generator
	spriteCache      *sprite.SpriteCache
	textureCache     map[string]*texture.Texture
	subtitleSystem   *subtitles.SubtitleSystem
	// Pre-allocated rendering buffers (Phase 2 optimization)
	particleBuffer     []byte // Pre-allocated buffer for particle rendering
	particleBufferSize int    // Current buffer size for reallocation check
	// Pre-allocated UI rendering images
	minimapImage    *ebiten.Image // Pre-allocated minimap image
	minimapPixels   []byte        // Pixel buffer for minimap
	barImage        *ebiten.Image // Pre-allocated bar image (health/mana)
	barPixels       []byte        // Pixel buffer for bar rendering
	crosshairImage  *ebiten.Image // Pre-allocated crosshair image
	speechBubbleImg *ebiten.Image // Pre-allocated speech bubble image
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
	// Inventory UI
	inventoryUI *InventoryUI
	// Quest UI
	questUI *QuestUI
	// Crafting UI
	craftingUI *CraftingUI
	// Faction UI
	factionUI *FactionUI
	// Housing UI
	housingUI *HousingUI
	// PvP UI
	pvpUI *PvPUI
	// State synchronization (online mode)
	stateSync *StateSynchronizer
	// NPC rendering
	npcRenderer *NPCRenderer
	// Debug/profiling fields
	frameTimeHistory []float64 // Ring buffer of recent frame times
	frameTimeIndex   int       // Current index in frame time history
	lastMemStats     runtime.MemStats
	// Mouse look state
	mouseCaptured    bool    // Is the mouse cursor captured for FPS-style look?
	lastMouseX       int     // Previous frame's mouse X position
	lastMouseY       int     // Previous frame's mouse Y position
	mouseInitialized bool    // Has the mouse position been initialized?
	smoothedDeltaX   float64 // Smoothed mouse delta X
	smoothedDeltaY   float64 // Smoothed mouse delta Y
}

// Update advances game state by one tick, processing player input and ECS systems.
func (g *Game) Update() error {
	// Use actual TPS for frame-rate independence
	actualTPS := ebiten.ActualTPS()
	dt := 1.0 / 60.0 // fallback
	if actualTPS > 0 {
		dt = 1.0 / actualTPS
	}
	g.recordFrameTime(dt)
	g.syncInputState()

	if g.gameState == GameStateCharacterCreation {
		return g.updateCharacterCreation()
	}

	g.handlePauseToggle()
	g.handleInventoryToggle()
	g.handleQuestToggle()
	g.handleFactionToggle()

	if err := g.handleQuitRequest(); err != nil {
		return err
	}

	if g.updateActiveOverlay(dt) {
		return nil
	}

	if !g.paused {
		g.updateGameplay(dt)
	}
	return nil
}

// handleQuitRequest checks if quit was requested from menu.
// Returns ebiten.Termination to gracefully exit via Ebitengine's shutdown path.
func (g *Game) handleQuitRequest() error {
	if g.menu != nil && g.menu.QuitRequested() {
		return ebiten.Termination
	}
	return nil
}

// updateActiveOverlay updates the currently open UI overlay.
func (g *Game) updateActiveOverlay(dt float64) bool {
	if g.menu != nil && g.menu.IsOpen() {
		g.menu.Update()
		return true
	}
	if g.inventoryUI != nil && g.inventoryUI.IsOpen() {
		g.inventoryUI.Update(g.world)
		return true
	}
	if g.questUI != nil && g.questUI.IsOpen() {
		g.questUI.Update(g.world, dt)
		return true
	}
	if g.craftingUI != nil && g.craftingUI.IsOpen() {
		g.craftingUI.Update(g.world)
		return true
	}
	if g.dialogUI != nil && g.dialogUI.IsOpen() {
		g.dialogUI.Update()
		return true
	}
	if g.factionUI != nil && g.factionUI.IsOpen() {
		g.factionUI.Update(g.world)
		return true
	}
	if g.housingUI != nil && g.housingUI.IsActive() {
		g.housingUI.Update()
		return true
	}
	return false
}

// updateGameplay handles all active gameplay updates.
func (g *Game) updateGameplay(dt float64) {
	g.handlePlayerInput(dt)
	g.handleInteraction()
	g.handleCombat(dt)
	g.world.Update(dt)
	g.updateChunkMap()
	g.updateRenderingSubsystems(dt)
	g.updateSubtitles(dt)
	g.updateNetworkSync(dt)
	g.updateBackgroundUI(dt)
	g.syncRendererPosition()
}

// syncRendererPosition updates the renderer's camera position from player entity.
// Must be called in Update(), not Draw(), to avoid race conditions per Ebitengine contract.
func (g *Game) syncRendererPosition() {
	if g.playerEntity == 0 || g.renderer == nil {
		return
	}
	if comp, ok := g.world.GetComponent(g.playerEntity, "Position"); ok {
		pos := comp.(*components.Position)
		g.renderer.SetPlayerPos(pos.X, pos.Y, pos.Angle)
	}
}

// updateNetworkSync synchronizes state with server if connected.
func (g *Game) updateNetworkSync(dt float64) {
	if g.stateSync == nil {
		return
	}
	g.stateSync.Update(dt)
	g.sendPlayerInputToServer()
}

// updateBackgroundUI updates UI elements that run in the background.
func (g *Game) updateBackgroundUI(dt float64) {
	// Only update questUI here if it's not the active overlay (avoids double-update)
	if g.questUI != nil && !g.questUI.IsOpen() {
		g.questUI.Update(g.world, dt)
	}
	if g.pvpUI != nil {
		playerPos := g.getPlayerPosition()
		g.pvpUI.Update(uint64(g.playerEntity), playerPos.X, playerPos.Y)
	}
	if g.housingUI != nil {
		playerPos := g.getPlayerPosition()
		angle := g.renderer.GetPlayerAngle()
		g.housingUI.UpdateFurniturePreview(playerPos.X, playerPos.Y, angle)
	}
}

// updateCharacterCreation handles the character creation screen state.
func (g *Game) updateCharacterCreation() error {
	if g.characterCreation == nil {
		g.characterCreation = NewCharacterCreation()
	}

	g.characterCreation.Update()

	if g.characterCreation.IsComplete() {
		// Apply character creation choices to player entity
		g.characterCreation.ApplyToPlayer(g.world, g.playerEntity)

		// Update config with selected genre if changed
		if g.characterCreation.GetSelectedGenre() != "" {
			g.cfg.Genre = g.characterCreation.GetSelectedGenre()
		}

		// Transition to playing state
		g.gameState = GameStatePlaying
		g.characterCreation = nil
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
		// Open crafting UI for this workbench
		if g.craftingUI != nil {
			g.craftingUI.OpenWorkbench(g.world, target.Entity)
		}
	case InteractionContainer:
		g.handleContainerInteraction(target.Entity, target.Name)
	case InteractionDoor:
		g.handleDoorInteraction(target.Entity)
	case InteractionMount:
		// Mount or dismount the creature
		g.handleMountInteraction(target.Entity)
	case InteractionVehicle:
		// Enter or exit the vehicle
		g.handleVehicleInteraction(target.Entity)
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

// handleContainerInteraction opens the container inventory UI.
func (g *Game) handleContainerInteraction(containerEntity ecs.Entity, containerName string) {
	containerInvComp, ok := g.world.GetComponent(containerEntity, "Inventory")
	if !ok {
		log.Printf("Container has no inventory component")
		return
	}
	containerInv := containerInvComp.(*components.Inventory)

	// Get player inventory for transfer
	playerInvComp, ok := g.world.GetComponent(g.playerEntity, "Inventory")
	if !ok {
		log.Printf("Player has no inventory component")
		return
	}

	// Open inventory UI with container view
	if g.inventoryUI != nil {
		g.inventoryUI.OpenContainer(containerEntity, containerName, containerInv, playerInvComp.(*components.Inventory))
		log.Printf("Opened container: %s (items: %d/%d)", containerName, len(containerInv.Items), containerInv.Capacity)
	}
}

// handleDoorInteraction opens or closes a door.
func (g *Game) handleDoorInteraction(doorEntity ecs.Entity) {
	// Check for Position component to determine door state
	posComp, ok := g.world.GetComponent(doorEntity, "Position")
	if !ok {
		return
	}
	pos := posComp.(*components.Position)

	// Toggle door state: doors store open/closed state in the Angle field (radians)
	// Open door has angle ~ PI/2, closed door has angle = 0
	doorOpenAngle := math.Pi / 2 // PI/2 radians (90 degrees)
	const doorCloseAngle = 0.0   // 0 radians
	doorThreshold := math.Pi / 4 // PI/4 radians (45 degrees)

	if pos.Angle > doorThreshold { // Currently open
		pos.Angle = doorCloseAngle
		log.Printf("Closed door")
	} else { // Currently closed
		pos.Angle = doorOpenAngle
		log.Printf("Opened door")
	}
}

// handleMountInteraction mounts or dismounts the player from a mount creature.
func (g *Game) handleMountInteraction(mountEntity ecs.Entity) {
	miComp, ok := g.world.GetComponent(mountEntity, "MountInfo")
	if !ok {
		return
	}
	mi := miComp.(*components.MountInfo)

	if mi.IsMounted && mi.RiderEntity == uint64(g.playerEntity) {
		// Dismount
		mi.IsMounted = false
		mi.RiderEntity = 0
		g.isRidingMount = false
		g.currentMountEntity = 0
		log.Printf("Dismounted from %s", mi.Name)
	} else if !mi.IsMounted {
		// Mount
		mi.IsMounted = true
		mi.RiderEntity = uint64(g.playerEntity)
		g.isRidingMount = true
		g.currentMountEntity = mountEntity
		log.Printf("Mounted %s", mi.Name)
	}
}

// handleVehicleInteraction enters or exits a vehicle.
func (g *Game) handleVehicleInteraction(vehicleEntity ecs.Entity) {
	stateComp, ok := g.world.GetComponent(vehicleEntity, "VehicleState")
	if !ok {
		return
	}
	state := stateComp.(*components.VehicleState)

	if state.IsOccupied && state.DriverEntity == uint64(g.playerEntity) {
		// Exit vehicle
		state.IsOccupied = false
		state.DriverEntity = 0
		state.EngineRunning = false
		state.InCockpitView = false
		g.isInVehicle = false
		g.currentVehicleEntity = 0
		log.Printf("Exited vehicle")
	} else if !state.IsOccupied {
		// Enter vehicle
		state.IsOccupied = true
		state.DriverEntity = uint64(g.playerEntity)
		state.EngineRunning = true
		state.InCockpitView = true
		g.isInVehicle = true
		g.currentVehicleEntity = vehicleEntity
		log.Printf("Entered vehicle")
	}
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

	// Update NPC animations
	if g.npcRenderer != nil {
		// Update animation frame timers
		g.npcRenderer.UpdateAnimations(g.world, dt)
		// Sync animation states with NPC schedules
		g.npcRenderer.SyncAnimationWithSchedule(g.world)
	}
}

// initUIBuffers initializes pre-allocated images for UI rendering.
// Uses WritePixels batch API instead of per-pixel Set() calls.
func (g *Game) initUIBuffers() {
	const minimapSize = 64
	const barMaxWidth = 150
	const barMaxHeight = 16
	const speechBubbleW = 42
	const speechBubbleH = 18
	const crosshairSize = 11

	// Pre-allocate minimap image and pixel buffer
	g.minimapImage = ebiten.NewImage(minimapSize, minimapSize)
	g.minimapPixels = make([]byte, minimapSize*minimapSize*4)

	// Pre-allocate bar image and pixel buffer
	g.barImage = ebiten.NewImage(barMaxWidth, barMaxHeight)
	g.barPixels = make([]byte, barMaxWidth*barMaxHeight*4)

	// Pre-allocate crosshair image (11x11 for the 5-pixel cross)
	g.crosshairImage = ebiten.NewImage(crosshairSize, crosshairSize)
	g.initCrosshairImage(crosshairSize)

	// Pre-allocate speech bubble image
	g.speechBubbleImg = ebiten.NewImage(speechBubbleW, speechBubbleH)
	g.initSpeechBubbleImage(speechBubbleW, speechBubbleH)
}

// initCrosshairImage renders the crosshair once into the pre-allocated image.
func (g *Game) initCrosshairImage(size int) {
	pixels := make([]byte, size*size*4)
	center := size / 2
	green := []byte{0, 255, 0, 255}

	// Horizontal line (center row: center-1, center, center+1)
	for dx := -1; dx <= 1; dx++ {
		px := center + dx
		py := center
		idx := (py*size + px) * 4
		copy(pixels[idx:idx+4], green)
	}
	// Vertical line (center col: center-1, center+1)
	for dy := -1; dy <= 1; dy++ {
		if dy == 0 {
			continue // Already drawn as part of horizontal
		}
		px := center
		py := center + dy
		idx := (py*size + px) * 4
		copy(pixels[idx:idx+4], green)
	}

	g.crosshairImage.WritePixels(pixels)
}

// initSpeechBubbleImage renders the speech bubble once into the pre-allocated image.
func (g *Game) initSpeechBubbleImage(w, h int) {
	pixels := make([]byte, w*h*4)
	bubbleR, bubbleG, bubbleB, bubbleA := byte(255), byte(255), byte(255), byte(200)
	textR, textG, textB, textA := byte(50), byte(50), byte(50), byte(255)

	centerX, centerY := w/2, h/2

	// Draw bubble background (ellipse approximation)
	for dy := -8; dy <= 8; dy++ {
		for dx := -20; dx <= 20; dx++ {
			if float64(dx*dx)/400+float64(dy*dy)/64 <= 1 {
				px := centerX + dx
				py := centerY + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					idx := (py*w + px) * 4
					pixels[idx] = bubbleR
					pixels[idx+1] = bubbleG
					pixels[idx+2] = bubbleB
					pixels[idx+3] = bubbleA
				}
			}
		}
	}

	// Draw "..." text (three dots)
	for i := -8; i <= 8; i += 8 {
		for ddx := 0; ddx < 3; ddx++ {
			for ddy := 0; ddy < 3; ddy++ {
				px := centerX + i + ddx - 1
				py := centerY + ddy - 1
				if px >= 0 && px < w && py >= 0 && py < h {
					idx := (py*w + px) * 4
					pixels[idx] = textR
					pixels[idx+1] = textG
					pixels[idx+2] = textB
					pixels[idx+3] = textA
				}
			}
		}
	}

	g.speechBubbleImg.WritePixels(pixels)
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
// Also toggles mouse capture state - captured during gameplay, released for menus.
func (g *Game) handlePauseToggle() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.menu != nil {
			g.menu.Toggle()
			g.paused = g.menu.IsOpen()
			// Release mouse when pausing, capture when resuming
			g.setMouseCaptured(!g.paused)
		} else {
			g.paused = !g.paused
			g.setMouseCaptured(!g.paused)
		}
	}
}

// handleInventoryToggle checks for inventory action and toggles the inventory UI.
func (g *Game) handleInventoryToggle() {
	// Don't toggle inventory if menu or dialog is open
	if g.menu != nil && g.menu.IsOpen() {
		return
	}
	if g.dialogUI != nil && g.dialogUI.IsOpen() {
		return
	}

	// Check for I key press
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		if g.inventoryUI != nil {
			g.inventoryUI.Toggle()
		}
	}
}

// handleQuestToggle checks for quest log action and toggles the quest UI.
func (g *Game) handleQuestToggle() {
	// Don't toggle quest log if menu, dialog, or inventory is open
	if g.menu != nil && g.menu.IsOpen() {
		return
	}
	if g.dialogUI != nil && g.dialogUI.IsOpen() {
		return
	}
	if g.inventoryUI != nil && g.inventoryUI.IsOpen() {
		return
	}

	// Check for J key press
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		if g.questUI != nil {
			g.questUI.Toggle()
		}
	}
}

// handleFactionToggle checks for faction action and toggles the faction UI.
func (g *Game) handleFactionToggle() {
	// Don't toggle if any overlay is open
	if g.isAnyOverlayOpen() {
		return
	}

	// Check for F key press (for Factions)
	if inpututil.IsKeyJustPressed(ebiten.KeyF) && g.factionUI != nil {
		g.factionUI.Toggle()
	}

	// Check for H key press (for Housing)
	if inpututil.IsKeyJustPressed(ebiten.KeyH) && g.housingUI != nil {
		g.toggleHousingUI()
	}
}

// isAnyOverlayOpen returns true if any UI overlay is currently open.
func (g *Game) isAnyOverlayOpen() bool {
	return (g.menu != nil && g.menu.IsOpen()) ||
		(g.dialogUI != nil && g.dialogUI.IsOpen()) ||
		(g.inventoryUI != nil && g.inventoryUI.IsOpen()) ||
		(g.questUI != nil && g.questUI.IsOpen()) ||
		(g.craftingUI != nil && g.craftingUI.IsOpen())
}

// toggleHousingUI opens or closes the housing UI.
func (g *Game) toggleHousingUI() {
	if g.housingUI.IsActive() {
		g.housingUI.Close()
	} else {
		g.housingUI.SetPlayerState(uint64(g.playerEntity), getPlayerGold(g), 1)
		g.housingUI.Open()
	}
}

// getPlayerPosition returns the player's current position, or a default if not found.
func (g *Game) getPlayerPosition() *components.Position {
	if g.world == nil || g.playerEntity == 0 {
		return &components.Position{X: 0, Y: 0, Z: 0}
	}
	comp, ok := g.world.GetComponent(g.playerEntity, "Position")
	if !ok {
		return &components.Position{X: 0, Y: 0, Z: 0}
	}
	return comp.(*components.Position)
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
// Uses LOD system for distant chunks and async generation to avoid frame stutter.
func (g *Game) rebuildWorldMap(centerChunkX, centerChunkY int) {
	worldMap := make([][]int, g.mapSize)
	for i := range worldMap {
		worldMap[i] = make([]int, g.mapSize)
	}
	chunkSize := g.cfg.World.ChunkSize

	// Update LOD viewpoint to player position (center of center chunk in world coords)
	viewX := float64(centerChunkX*chunkSize) + float64(chunkSize)/2
	viewY := float64(centerChunkY*chunkSize) + float64(chunkSize)/2
	g.lodManager.SetViewpoint(viewX, viewY)

	// Load 3x3 chunk window using LOD-based selection with async generation
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			chunkX := centerChunkX + dx
			chunkY := centerChunkY + dy
			// GetChunkLODAsync returns placeholder while chunk generates in background
			lodChunk, ok := g.lodManager.GetChunkLODAsync(chunkX, chunkY)
			if !ok || lodChunk == nil {
				// Skip chunk on error - player can still move through pending chunks
				continue
			}
			g.sampleLODChunkIntoMap(worldMap, lodChunk, dx, dy, chunkSize)
		}
	}
	g.worldMap = worldMap // Store for collision detection
	g.renderer.SetWorldMapDirect(worldMap)
}

// sampleLODChunkIntoMap samples an LOD chunk's heightmap into the local world map.
// Uses the LOD-appropriate heightmap which may be downsampled for distant chunks.
func (g *Game) sampleLODChunkIntoMap(worldMap [][]int, lod *chunk.LODChunk, dx, dy, chunkSize int) {
	if lod == nil {
		return
	}
	sectionSize := g.mapSize / 3
	startX := (dx + 1) * sectionSize
	startY := (dy + 1) * sectionSize

	lodSize := lod.LODSize
	heightMap := lod.HeightMap

	for ly := 0; ly < sectionSize; ly++ {
		for lx := 0; lx < sectionSize; lx++ {
			// Map local coords to LOD chunk coords
			cx := lx * lodSize / sectionSize
			cy := ly * lodSize / sectionSize
			if cx >= lodSize {
				cx = lodSize - 1
			}
			if cy >= lodSize {
				cy = lodSize - 1
			}
			idx := cy*lodSize + cx
			if idx < len(heightMap) {
				worldMap[startY+ly][startX+lx] = heightToWallType(heightMap[idx], g.wallThreshold)
			}
		}
	}
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

// handlePlayerInput processes keyboard and mouse input for player movement.
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
	g.handleMouseLook(pos)
}

// canMoveTo checks if a position is valid (not inside a wall).
// Uses player radius for wall sliding behavior.
func (g *Game) canMoveTo(x, y float64) bool {
	if g.worldMap == nil || len(g.worldMap) == 0 {
		return true // No map loaded yet, allow movement
	}

	// Check multiple points around player position for collision
	for _, offset := range [][2]float64{{0, 0}, {playerRadius, 0}, {-playerRadius, 0}, {0, playerRadius}, {0, -playerRadius}} {
		checkX := x + offset[0]
		checkY := y + offset[1]

		// Convert world position to map coordinates
		mapX := int(checkX)
		mapY := int(checkY)

		// Out of bounds - allow movement for seamless chunk transitions
		if mapY < 0 || mapY >= len(g.worldMap) || mapX < 0 || mapX >= len(g.worldMap[0]) {
			continue // Allow movement into areas outside loaded map
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
	// Use input manager if available, fallback to direct key checks
	moveForward := g.isActionOrKeyPressed(input.ActionMoveForward, ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)
	moveBackward := g.isActionOrKeyPressed(input.ActionMoveBackward, ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)
	turnLeft := g.isActionOrKeyPressed(input.ActionMoveLeft, ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft)
	turnRight := g.isActionOrKeyPressed(input.ActionMoveRight, ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight)

	if moveForward {
		dx := math.Cos(pos.Angle) * playerMoveSpeed * dt
		dy := math.Sin(pos.Angle) * playerMoveSpeed * dt
		g.tryMove(pos, dx, dy)
	}
	if moveBackward {
		dx := -math.Cos(pos.Angle) * playerMoveSpeed * dt
		dy := -math.Sin(pos.Angle) * playerMoveSpeed * dt
		g.tryMove(pos, dx, dy)
	}
	if turnLeft {
		pos.Angle -= playerTurnSpeed * dt
	}
	if turnRight {
		pos.Angle += playerTurnSpeed * dt
	}
}

// processStrafeInput handles left/right strafe movement.
func (g *Game) processStrafeInput(pos *components.Position, dt float64) {
	if g.isActionOrKeyPressed(input.ActionStrafeLeft, ebiten.KeyQ) {
		dx := math.Cos(pos.Angle-math.Pi/2) * playerMoveSpeed * dt
		dy := math.Sin(pos.Angle-math.Pi/2) * playerMoveSpeed * dt
		g.tryMove(pos, dx, dy)
	}
	if g.isActionOrKeyPressed(input.ActionStrafeRight, ebiten.KeyE) {
		dx := math.Cos(pos.Angle+math.Pi/2) * playerMoveSpeed * dt
		dy := math.Sin(pos.Angle+math.Pi/2) * playerMoveSpeed * dt
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

// isActionOrMousePressed checks if an action is pressed via input manager or fallback mouse button.
func (g *Game) isActionOrMousePressed(action input.Action, fallbackButton ebiten.MouseButton) bool {
	if g.inputManager != nil {
		return g.inputManager.IsActionPressed(action)
	}
	return ebiten.IsMouseButtonPressed(fallbackButton)
}

// sendPlayerInputToServer sends current player input state to the server.
func (g *Game) sendPlayerInputToServer() {
	if g.stateSync == nil {
		return
	}

	// Gather current input state
	var moveForward, moveRight, turn float32

	// Forward/backward
	if g.isActionOrKeyPressed(input.ActionMoveForward, ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		moveForward = 1.0
	} else if g.isActionOrKeyPressed(input.ActionMoveBackward, ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		moveForward = -1.0
	}

	// Strafe left/right (Q/E)
	if g.isActionOrKeyPressed(input.ActionStrafeLeft, ebiten.KeyQ) {
		moveRight = -1.0
	} else if g.isActionOrKeyPressed(input.ActionStrafeRight, ebiten.KeyE) {
		moveRight = 1.0
	}

	// Turning (A/D or arrows)
	if g.isActionOrKeyPressed(input.ActionMoveLeft, ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		turn = -0.05
	} else if g.isActionOrKeyPressed(input.ActionMoveRight, ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		turn = 0.05
	}

	// Jump (Space)
	jump := g.isActionOrKeyPressed(input.ActionJump, ebiten.KeySpace)

	// Attack (Left mouse) - uses ActionAttack
	attack := g.isActionOrMousePressed(input.ActionAttack, ebiten.MouseButtonLeft)

	// Block/Use (Right mouse) - uses ActionBlock
	use := g.isActionOrMousePressed(input.ActionBlock, ebiten.MouseButtonRight)

	// Only send if there's actual input
	if moveForward != 0 || moveRight != 0 || turn != 0 || jump || attack || use {
		if err := g.stateSync.SendPlayerInput(moveForward, moveRight, turn, jump, attack, use); err != nil {
			log.Printf("failed to send player input: %v", err)
		}
	}
}

// toggleMouseCapture toggles the mouse capture state.
// When captured, the cursor is hidden and mouse movement controls camera.
func (g *Game) toggleMouseCapture() {
	g.mouseCaptured = !g.mouseCaptured
	if g.mouseCaptured {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		// Reset mouse tracking on capture to prevent jumps
		g.mouseInitialized = false
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	}
}

// setMouseCaptured explicitly sets the mouse capture state.
func (g *Game) setMouseCaptured(captured bool) {
	if g.mouseCaptured == captured {
		return
	}
	g.mouseCaptured = captured
	if captured {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		g.mouseInitialized = false
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	}
}

// handleMouseLook processes mouse movement for FPS-style camera control.
// Updates player yaw (horizontal) and renderer pitch (vertical).
func (g *Game) handleMouseLook(pos *components.Position) {
	if !g.mouseCaptured || g.cfg == nil {
		return
	}

	// Get current mouse position
	cursorX, cursorY := ebiten.CursorPosition()

	// Initialize on first frame to prevent jump
	if !g.mouseInitialized {
		g.lastMouseX = cursorX
		g.lastMouseY = cursorY
		g.mouseInitialized = true
		return
	}

	// Calculate raw mouse delta
	deltaX := float64(cursorX - g.lastMouseX)
	deltaY := float64(cursorY - g.lastMouseY)
	g.lastMouseX = cursorX
	g.lastMouseY = cursorY

	// Skip if no movement
	if deltaX == 0 && deltaY == 0 {
		return
	}

	// Apply mouse acceleration if enabled
	if g.cfg.Mouse.AccelerationOn {
		magnitude := math.Sqrt(deltaX*deltaX + deltaY*deltaY)
		accelerationFactor := 1.0 + (magnitude * g.cfg.Mouse.Acceleration * 0.01)
		deltaX *= accelerationFactor
		deltaY *= accelerationFactor
	}

	// Apply sensitivity
	sensitivity := g.cfg.Mouse.Sensitivity * 0.005 // Scale to reasonable range
	deltaX *= sensitivity
	deltaY *= sensitivity

	// Apply smoothing if enabled
	if g.cfg.Mouse.SmoothingOn {
		factor := g.cfg.Mouse.SmoothingFactor
		g.smoothedDeltaX = g.smoothedDeltaX*(1-factor) + deltaX*factor
		g.smoothedDeltaY = g.smoothedDeltaY*(1-factor) + deltaY*factor
		deltaX = g.smoothedDeltaX
		deltaY = g.smoothedDeltaY
	}

	// Invert Y if configured
	if g.cfg.Mouse.InvertY {
		deltaY = -deltaY
	}

	// Update player yaw (horizontal rotation)
	pos.Angle += deltaX

	// Normalize angle to [-π, π]
	for pos.Angle > math.Pi {
		pos.Angle -= 2 * math.Pi
	}
	for pos.Angle < -math.Pi {
		pos.Angle += 2 * math.Pi
	}

	// Update renderer pitch (vertical look)
	if g.renderer != nil {
		g.renderer.PlayerPitch -= deltaY // Subtract because screen Y is inverted
		// Clamp pitch to prevent over-rotation
		if g.renderer.PlayerPitch > raycast.MaxPitchAngle {
			g.renderer.PlayerPitch = raycast.MaxPitchAngle
		}
		if g.renderer.PlayerPitch < -raycast.MaxPitchAngle {
			g.renderer.PlayerPitch = -raycast.MaxPitchAngle
		}
	}
}

// Draw renders the current frame using the raycaster and displays debug info.
func (g *Game) Draw(screen *ebiten.Image) {
	// Handle character creation state
	if g.gameState == GameStateCharacterCreation {
		g.drawCharacterCreation(screen)
		return
	}

	// Render 3D world (walls, floor, ceiling, NPCs, effects)
	g.drawWorld(screen)

	// Render UI overlays on top of world
	g.drawUIOverlays(screen)
}

// drawWorld renders the 3D world, NPCs, post-processing, and combat effects.
func (g *Game) drawWorld(screen *ebiten.Image) {
	// Player position is synced in Update() via syncRendererPosition() to avoid
	// race conditions between Update and Draw goroutines per Ebitengine contract.

	// Render base scene (walls, floor, ceiling)
	g.renderer.Draw(screen)

	// Render NPC sprites as billboards
	g.drawNPCs(screen)

	// Apply post-processing effects (genre-specific)
	g.applyPostProcessing(screen)

	// Draw particles (weather effects) on top
	g.drawParticles(screen)

	// Apply combat visual feedback (screen shake, damage flash)
	g.applyCombatVisualFeedback(screen)

	// Draw HUD last (on top of everything)
	g.drawHUD(screen)
}

// drawUIOverlays renders all UI overlay elements.
func (g *Game) drawUIOverlays(screen *ebiten.Image) {
	// Draw inventory UI overlay if active
	if g.inventoryUI != nil && g.inventoryUI.IsOpen() {
		g.inventoryUI.Draw(screen, g.world)
	}

	// Draw quest UI overlay and tracker
	if g.questUI != nil {
		g.questUI.Draw(screen)
	}

	// Draw crafting UI overlay if active
	if g.craftingUI != nil && g.craftingUI.IsOpen() {
		g.craftingUI.Draw(screen, g.world)
	}

	// Draw faction UI overlay if active
	if g.factionUI != nil && g.factionUI.IsOpen() {
		g.factionUI.Draw(screen, g.world)
	}

	// Draw housing UI overlay if active
	if g.housingUI != nil && g.housingUI.IsActive() {
		g.housingUI.Draw(screen)
	}

	// Draw PvP zone indicators (always visible)
	if g.pvpUI != nil {
		playerPos := g.getPlayerPosition()
		g.pvpUI.Draw(screen, uint64(g.playerEntity), playerPos.X, playerPos.Y)
		g.pvpUI.DrawZoneBoundaryIndicator(screen, playerPos.X, playerPos.Y)
	}

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
// Uses the renderer's framebuffer to avoid per-pixel GPU calls.
func (g *Game) applyCombatVisualFeedback(screen *ebiten.Image) {
	if g.combatManager == nil {
		return
	}

	// Apply damage flash overlay
	flashAlpha := g.combatManager.GetDamageFlashAlpha()
	if flashAlpha > 0 {
		framebuffer := g.renderer.GetFramebuffer()
		if framebuffer == nil {
			return
		}

		blendFactor := float64(flashAlpha) / 255.0
		invBlend := 1.0 - blendFactor
		flashR := 255.0 * blendFactor // Pre-compute flash color contribution

		// Blend red flash directly into framebuffer
		for i := 0; i < len(framebuffer); i += 4 {
			r := float64(framebuffer[i])
			gr := float64(framebuffer[i+1])
			b := float64(framebuffer[i+2])

			framebuffer[i] = uint8(r*invBlend + flashR)
			framebuffer[i+1] = uint8(gr * invBlend)
			framebuffer[i+2] = uint8(b * invBlend)
			// Alpha stays at 255
		}

		// Upload modified framebuffer
		screen.WritePixels(framebuffer)
	}
}

// drawCharacterCreation renders the character creation screen.
func (g *Game) drawCharacterCreation(screen *ebiten.Image) {
	if g.characterCreation == nil {
		return
	}
	g.characterCreation.Draw(screen)
}

// drawDeathScreen draws the death overlay when the player is dead.
// Uses the framebuffer for overlay, then falls back to ebitenutil for text.
func (g *Game) drawDeathScreen(screen *ebiten.Image) {
	if g.combatManager == nil || !g.combatManager.IsDead() {
		return
	}

	framebuffer := g.renderer.GetFramebuffer()
	if framebuffer != nil {
		// Apply dark overlay directly to framebuffer
		// alpha = 180/255 ≈ 0.706
		blendFactor := 180.0 / 255.0
		invBlend := 1.0 - blendFactor

		for i := 0; i < len(framebuffer); i += 4 {
			// Blend toward black (R=0, G=0, B=0)
			framebuffer[i] = uint8(float64(framebuffer[i]) * invBlend)
			framebuffer[i+1] = uint8(float64(framebuffer[i+1]) * invBlend)
			framebuffer[i+2] = uint8(float64(framebuffer[i+2]) * invBlend)
		}

		// Upload modified framebuffer
		screen.WritePixels(framebuffer)
	}

	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()

	// Draw death message (using ebitenutil which is GPU-accelerated)
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

// drawNPCs renders NPC entities as billboard sprites in the first-person view.
func (g *Game) drawNPCs(screen *ebiten.Image) {
	if g.npcRenderer == nil {
		return
	}

	// Build sprite entities from ECS world
	spriteEntities := g.npcRenderer.BuildSpriteEntities(g.world, g.playerEntity)
	if len(spriteEntities) == 0 {
		return
	}

	// Render sprites to screen using the raycast renderer's billboard system
	g.renderer.DrawSpritesToScreen(spriteEntities, screen)

	// Draw conversation indicators for NPCs in dialog
	g.drawNPCConversations(screen)
}

// drawNPCConversations draws speech bubble indicators for NPCs engaged in conversation.
func (g *Game) drawNPCConversations(screen *ebiten.Image) {
	if g.world == nil || g.renderer == nil {
		return
	}

	playerPos := g.getPlayerPositionOrNil()
	if playerPos == nil {
		return
	}

	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()
	playerAngle := g.renderer.GetPlayerAngle()

	for _, entity := range g.world.Entities("DialogState", "Position") {
		if !g.shouldDrawNPCConversation(entity) {
			continue
		}

		npcPos := g.getNPCPosition(entity)
		if npcPos == nil {
			continue
		}

		screenX, screenY, visible := g.projectNPCToScreen(playerPos, npcPos, playerAngle, screenW, screenH)
		if visible {
			g.drawSpeechBubble(screen, screenX, screenY)
		}
	}
}

// getPlayerPositionOrNil returns player position or nil if unavailable.
func (g *Game) getPlayerPositionOrNil() *components.Position {
	playerPosComp, ok := g.world.GetComponent(g.playerEntity, "Position")
	if !ok {
		return nil
	}
	return playerPosComp.(*components.Position)
}

// shouldDrawNPCConversation checks if an NPC should have a conversation indicator.
func (g *Game) shouldDrawNPCConversation(entity ecs.Entity) bool {
	if entity == g.playerEntity {
		return false
	}
	dialogComp, ok := g.world.GetComponent(entity, "DialogState")
	if !ok {
		return false
	}
	dialogState := dialogComp.(*components.DialogState)
	if !dialogState.IsInDialog {
		return false
	}
	return dialogState.ConversationPartner != uint64(g.playerEntity)
}

// getNPCPosition returns the position of an NPC entity.
func (g *Game) getNPCPosition(entity ecs.Entity) *components.Position {
	posComp, ok := g.world.GetComponent(entity, "Position")
	if !ok {
		return nil
	}
	return posComp.(*components.Position)
}

// projectNPCToScreen converts NPC world position to screen coordinates.
func (g *Game) projectNPCToScreen(playerPos, npcPos *components.Position, playerAngle float64, screenW, screenH int) (int, int, bool) {
	dx := npcPos.X - playerPos.X
	dy := npcPos.Y - playerPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance > 20 {
		return 0, 0, false
	}

	angle := math.Atan2(dy, dx)
	relAngle := normalizeAngle(angle - playerAngle)

	const fov = math.Pi / 3 // 60 degrees
	if math.Abs(relAngle) > fov {
		return 0, 0, false
	}

	screenX := screenW/2 + int(float64(screenW/2)*(relAngle/fov))
	screenY := screenH/2 - int(50/distance) - 20

	return screenX, screenY, true
}

// normalizeAngle normalizes an angle to the range -PI to PI.
func normalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// drawSpeechBubble draws a simple "..." speech indicator at the given position.
// Uses pre-allocated speech bubble image for GPU-efficient rendering.
func (g *Game) drawSpeechBubble(screen *ebiten.Image, x, y int) {
	// Ensure speech bubble image is initialized (lazy init)
	if g.speechBubbleImg == nil {
		const w, h = 42, 18
		g.speechBubbleImg = ebiten.NewImage(w, h)
		g.initSpeechBubbleImage(w, h)
	}

	op := &ebiten.DrawImageOptions{}
	// Center the bubble at (x, y)
	op.GeoM.Translate(float64(x-21), float64(y-9))
	screen.DrawImage(g.speechBubbleImg, op)
}

// applyPostProcessing applies genre-specific visual effects to the rendered frame.
// Uses pre-allocated buffers in the pipeline for efficiency.
func (g *Game) applyPostProcessing(screen *ebiten.Image) {
	if g.postprocessPipe == nil {
		return
	}

	// Use the renderer's framebuffer directly
	framebuffer := g.renderer.GetFramebuffer()
	if framebuffer == nil {
		return
	}

	// Create an RGBA image view over the framebuffer
	bounds := screen.Bounds()
	rgba := &image.RGBA{
		Pix:    framebuffer,
		Stride: bounds.Dx() * 4,
		Rect:   bounds,
	}

	// Apply post-processing pipeline (uses pre-allocated buffers internally)
	g.postprocessPipe.Apply(rgba)

	// Upload the modified framebuffer back to screen
	screen.WritePixels(framebuffer)
}

// drawParticles renders weather particles to the screen.
// Uses pre-allocated buffer to minimize allocations.
func (g *Game) drawParticles(screen *ebiten.Image) {
	if g.particleSystem == nil || g.particleRenderer == nil {
		return
	}

	// Use the renderer's framebuffer directly for particle rendering
	// The framebuffer already contains the rendered scene
	framebuffer := g.renderer.GetFramebuffer()
	if framebuffer == nil {
		return
	}

	// Render particles directly to the framebuffer
	g.particleRenderer.Draw(g.particleSystem, framebuffer)

	// Upload the modified framebuffer back to screen
	screen.WritePixels(framebuffer)
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

	// Draw bounty/wanted status (below minimap)
	g.drawWantedStatus(screen, screenWidth-80, 80)

	// Draw interaction prompt (bottom-center)
	g.drawInteractionPrompt(screen)

	// Draw debug info if enabled
	g.drawDebugInfo(screen)
}

// drawDebugInfo renders frame time, memory stats, and entity count when debug options are enabled.
func (g *Game) drawDebugInfo(screen *ebiten.Image) {
	if g.cfg == nil {
		return
	}
	cfg := g.cfg.Debug
	if !cfg.ShowFrameTime && !cfg.ShowMemStats && !cfg.ShowEntityCount {
		return
	}

	y := 100 // Start below the position info
	if cfg.ShowFrameTime {
		avgFrameTime := g.getAverageFrameTime()
		fps := 1.0 / avgFrameTime
		debugText := fmt.Sprintf("Frame: %.2fms (%.0f FPS)", avgFrameTime*1000, fps)
		ebitenutil.DebugPrintAt(screen, debugText, 10, y)
		y += 16
	}
	if cfg.ShowMemStats {
		runtime.ReadMemStats(&g.lastMemStats)
		heapMB := float64(g.lastMemStats.HeapAlloc) / 1024.0 / 1024.0
		numGC := g.lastMemStats.NumGC
		debugText := fmt.Sprintf("Heap: %.1f MB | GC: %d", heapMB, numGC)
		ebitenutil.DebugPrintAt(screen, debugText, 10, y)
		y += 16
	}
	if cfg.ShowEntityCount {
		entityCount := len(g.world.AllEntities())
		debugText := fmt.Sprintf("Entities: %d", entityCount)
		ebitenutil.DebugPrintAt(screen, debugText, 10, y)
	}
}

// getAverageFrameTime calculates the average frame time from the history buffer.
func (g *Game) getAverageFrameTime() float64 {
	var sum float64
	var count int
	for _, ft := range g.frameTimeHistory {
		if ft > 0 {
			sum += ft
			count++
		}
	}
	if count == 0 {
		return 1.0 / 60.0 // Default to 60 FPS
	}
	return sum / float64(count)
}

// recordFrameTime records a frame time measurement.
func (g *Game) recordFrameTime(dt float64) {
	g.frameTimeHistory[g.frameTimeIndex] = dt
	g.frameTimeIndex = (g.frameTimeIndex + 1) % len(g.frameTimeHistory)
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

// drawWantedStatus renders the player's bounty and wanted status.
func (g *Game) drawWantedStatus(screen *ebiten.Image, x, y int) {
	if g.playerEntity == 0 {
		return
	}

	crimeComp, ok := g.world.GetComponent(g.playerEntity, "Crime")
	if !ok {
		return
	}
	crime := crimeComp.(*components.Crime)

	// Only show if wanted or has bounty
	if crime.WantedLevel == 0 && crime.BountyAmount == 0 && !crime.InJail {
		return
	}

	// Draw wanted status box
	boxWidth := 70
	boxHeight := 40
	if crime.InJail {
		boxHeight = 50
	}

	// Background
	bgColor := color.RGBA{0, 0, 0, 180}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(boxWidth), float64(boxHeight), bgColor)

	// Border color based on wanted level
	var borderColor color.RGBA
	switch crime.WantedLevel {
	case 1:
		borderColor = color.RGBA{255, 255, 0, 255} // Yellow
	case 2:
		borderColor = color.RGBA{255, 165, 0, 255} // Orange
	case 3, 4, 5:
		borderColor = color.RGBA{255, 0, 0, 255} // Red
	default:
		borderColor = color.RGBA{100, 100, 100, 255}
	}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(boxWidth), 2, borderColor)

	// Wanted level text
	wantedStr := ""
	if crime.WantedLevel > 0 {
		stars := ""
		for i := 0; i < crime.WantedLevel; i++ {
			stars += "*"
		}
		wantedStr = fmt.Sprintf("WANTED %s", stars)
		ebitenutil.DebugPrintAt(screen, wantedStr, x+5, y+5)
	}

	// Bounty amount
	if crime.BountyAmount > 0 {
		bountyStr := fmt.Sprintf("Bounty: %.0fg", crime.BountyAmount)
		ebitenutil.DebugPrintAt(screen, bountyStr, x+5, y+20)
	}

	// Jail status
	if crime.InJail {
		ebitenutil.DebugPrintAt(screen, "IN JAIL", x+5, y+35)
	}
}

// drawBar renders a horizontal bar (health/mana style) using batch pixel writes.
// Uses WritePixels() for GPU-efficient rendering with zero per-pixel Set() calls.
func (g *Game) drawBar(screen *ebiten.Image, x, y, width, height int, percent float64, fillColor, bgColor uint32) {
	g.ensureBarBuffers(width, height)

	// Use pre-allocated pixel buffer (slice to required size)
	bufSize := width * height * 4
	pixels := g.barPixels[:bufSize]

	g.fillBarPixels(pixels, width, height, percent, fillColor, bgColor)

	// Create a sub-image of the correct size and write pixels
	subImg := g.barImage.SubImage(image.Rect(0, 0, width, height)).(*ebiten.Image)
	subImg.WritePixels(pixels)

	// Draw to screen
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(subImg, op)
}

// ensureBarBuffers initializes or resizes bar rendering buffers.
func (g *Game) ensureBarBuffers(width, height int) {
	if g.barImage == nil {
		g.barImage = ebiten.NewImage(200, 32)
		g.barPixels = make([]byte, 200*32*4)
	}

	bounds := g.barImage.Bounds()
	if bounds.Dx() < width || bounds.Dy() < height {
		newW := max(bounds.Dx(), width)
		newH := max(bounds.Dy(), height)
		g.barImage = ebiten.NewImage(newW, newH)
		g.barPixels = make([]byte, newW*newH*4)
	} else if len(g.barPixels) < width*height*4 {
		g.barPixels = make([]byte, bounds.Dx()*bounds.Dy()*4)
	}

	g.barImage.Clear()
}

// fillBarPixels renders the bar content to the pixel buffer.
func (g *Game) fillBarPixels(pixels []byte, width, height int, percent float64, fillColor, bgColor uint32) {
	bgR, bgG, bgB, bgA := uint8(bgColor>>24), uint8(bgColor>>16), uint8(bgColor>>8), uint8(bgColor)
	fillR, fillG, fillB, fillA := uint8(fillColor>>24), uint8(fillColor>>16), uint8(fillColor>>8), uint8(fillColor)

	// Fill background
	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			idx := (py*width + px) * 4
			pixels[idx] = bgR
			pixels[idx+1] = bgG
			pixels[idx+2] = bgB
			pixels[idx+3] = bgA
		}
	}

	// Fill progress (with 1px border)
	fillWidth := int(float64(width) * percent)
	for py := 1; py < height-1; py++ {
		for px := 1; px < fillWidth-1 && px < width-1; px++ {
			idx := (py*width + px) * 4
			pixels[idx] = fillR
			pixels[idx+1] = fillG
			pixels[idx+2] = fillB
			pixels[idx+3] = fillA
		}
	}
}

// drawMinimap renders a small top-down view of the nearby area using batch pixel writes.
// Uses WritePixels() for GPU-efficient rendering with zero per-pixel Set() calls.
func (g *Game) drawMinimap(screen *ebiten.Image, x, y, size int) {
	if g.worldMap == nil || len(g.worldMap) == 0 {
		return
	}

	g.ensureMinimapBuffers(size)

	playerMapX, playerMapY := g.getMinimapPlayerPosition()
	territories := g.getFactionTerritories()

	pixels := g.minimapPixels[:size*size*4]
	g.fillMinimapPixels(pixels, size, playerMapX, playerMapY, territories)
	g.drawMinimapPlayerMarker(pixels, size)

	// Write pixels and draw to screen
	subImg := g.minimapImage.SubImage(image.Rect(0, 0, size, size)).(*ebiten.Image)
	subImg.WritePixels(pixels)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(subImg, op)
}

// ensureMinimapBuffers initializes or resizes minimap pixel buffers.
func (g *Game) ensureMinimapBuffers(size int) {
	if g.minimapImage == nil || g.minimapPixels == nil {
		g.minimapImage = ebiten.NewImage(size, size)
		g.minimapPixels = make([]byte, size*size*4)
		return
	}
	bounds := g.minimapImage.Bounds()
	if bounds.Dx() < size || bounds.Dy() < size {
		g.minimapImage = ebiten.NewImage(size, size)
		g.minimapPixels = make([]byte, size*size*4)
	} else if len(g.minimapPixels) < size*size*4 {
		g.minimapPixels = make([]byte, size*size*4)
	}
}

// getMinimapPlayerPosition returns the player's position for minimap centering.
func (g *Game) getMinimapPlayerPosition() (int, int) {
	if g.playerEntity == 0 {
		return 0, 0
	}
	comp, ok := g.world.GetComponent(g.playerEntity, "Position")
	if !ok {
		return 0, 0
	}
	pos := comp.(*components.Position)
	return int(pos.X), int(pos.Y)
}

// fillMinimapPixels renders terrain and territory colors to the pixel buffer.
func (g *Game) fillMinimapPixels(pixels []byte, size, playerX, playerY int, territories []*components.FactionTerritory) {
	mapRadius := 16
	for my := 0; my < size; my++ {
		for mx := 0; mx < size; mx++ {
			worldX := playerX - mapRadius + (mx * mapRadius * 2 / size)
			worldY := playerY - mapRadius + (my * mapRadius * 2 / size)

			baseColor := g.getMinimapTileColor(worldX, worldY)
			if tc := g.getTerritoryColor(float64(worldX), float64(worldY), territories); tc != 0 {
				baseColor = blendColors(baseColor, tc, 0.4)
			}

			idx := (my*size + mx) * 4
			pixels[idx] = uint8(baseColor >> 24)
			pixels[idx+1] = uint8(baseColor >> 16)
			pixels[idx+2] = uint8(baseColor >> 8)
			pixels[idx+3] = uint8(baseColor)
		}
	}
}

// getMinimapTileColor returns the base color for a minimap tile.
func (g *Game) getMinimapTileColor(worldX, worldY int) uint32 {
	if worldY < 0 || worldY >= len(g.worldMap) || worldX < 0 || worldX >= len(g.worldMap[0]) {
		return 0x111111FF
	}
	if g.worldMap[worldY][worldX] > 0 {
		return 0x666666FF
	}
	return 0x333333FF
}

// drawMinimapPlayerMarker draws the green player marker in the center of the minimap.
func (g *Game) drawMinimapPlayerMarker(pixels []byte, size int) {
	center := size / 2
	green := []byte{0, 255, 0, 255}
	offsets := [][2]int{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for _, off := range offsets {
		px, py := center+off[0], center+off[1]
		if px >= 0 && px < size && py >= 0 && py < size {
			idx := (py*size + px) * 4
			copy(pixels[idx:idx+4], green)
		}
	}
}

// getFactionTerritories retrieves all faction territory entities.
func (g *Game) getFactionTerritories() []*components.FactionTerritory {
	var territories []*components.FactionTerritory
	for _, e := range g.world.Entities("FactionTerritory") {
		if comp, ok := g.world.GetComponent(e, "FactionTerritory"); ok {
			territories = append(territories, comp.(*components.FactionTerritory))
		}
	}
	return territories
}

// getTerritoryColor returns a color for the territory at the given point.
func (g *Game) getTerritoryColor(x, y float64, territories []*components.FactionTerritory) uint32 {
	for _, t := range territories {
		if t.ContainsPoint(x, y) {
			return g.factionToColor(t.FactionID)
		}
	}
	return 0
}

// factionToColor maps faction IDs to minimap colors.
func (gm *Game) factionToColor(factionID string) uint32 {
	// Use a simple hash to get consistent colors per faction
	hash := uint32(0)
	for _, c := range factionID {
		hash = hash*31 + uint32(c)
	}
	// Generate color with moderate saturation and alpha
	rr := uint8(80 + (hash % 80))
	gg := uint8(80 + ((hash / 256) % 80))
	bb := uint8(80 + ((hash / 65536) % 80))
	return uint32(rr)<<24 | uint32(gg)<<16 | uint32(bb)<<8 | 0xCC
}

// blendColors blends two RGBA colors by the given factor (0=first, 1=second).
func blendColors(c1, c2 uint32, factor float64) uint32 {
	r1, g1, b1, a1 := uint8((c1>>24)&0xFF), uint8((c1>>16)&0xFF), uint8((c1>>8)&0xFF), uint8(c1&0xFF)
	r2, g2, b2, a2 := uint8((c2>>24)&0xFF), uint8((c2>>16)&0xFF), uint8((c2>>8)&0xFF), uint8(c2&0xFF)

	r := uint8(float64(r1)*(1-factor) + float64(r2)*factor)
	g := uint8(float64(g1)*(1-factor) + float64(g2)*factor)
	b := uint8(float64(b1)*(1-factor) + float64(b2)*factor)
	a := uint8(float64(a1)*(1-factor) + float64(a2)*factor)

	return uint32(r)<<24 | uint32(g)<<16 | uint32(b)<<8 | uint32(a)
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
	// Use actual window dimensions for responsive layout
	return outsideWidth, outsideHeight
}

// startProfileServer starts the pprof HTTP server for runtime profiling.
func startProfileServer(port int) {
	addr := fmt.Sprintf("localhost:%d", port)
	log.Printf("Starting pprof server at http://%s/debug/pprof/", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("pprof server error: %v", err)
		}
	}()
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Start profiling server if enabled
	if cfg.Debug.ProfilingEnabled {
		startProfileServer(cfg.Debug.ProfilingPort)
	}

	world := ecs.NewWorld()
	renderer := raycast.NewRenderer(cfg.Window.Width, cfg.Window.Height)
	player := createPlayerEntity(world)
	client, connected := connectToServer(cfg)
	// Register client systems; in offline mode (!connected), also register game logic systems
	registerClientSystems(world, player, cfg, !connected)
	chunkMgr := chunk.NewManager(cfg.World.ChunkSize, cfg.World.Seed)
	chunkMgr.EnableAsyncGeneration(4) // Enable background chunk generation with 4 workers
	lodMgr := chunk.NewLODManager(chunkMgr)

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

	// Set up save/load handlers if connected
	if connected {
		menu.SetSaveHandler(func() error {
			return client.SendSaveRequest()
		})
		menu.SetLoadHandler(func() error {
			return client.SendLoadRequest()
		})
	}

	// Initialize inventory UI
	inventoryUI := NewInventoryUI(cfg.Genre, player, inputMgr)

	// Initialize quest UI
	questUI := NewQuestUI(cfg.Genre, player, inputMgr)

	// Initialize crafting UI
	craftingUI := NewCraftingUI(cfg.Genre, player)

	// Initialize faction UI
	factionUI := NewFactionUI(cfg.Genre, player)

	// Initialize housing UI
	housingUI := NewHousingUI()
	housingUI.Initialize()

	// Initialize PvP UI
	pvpUI := NewPvPUI()

	// Initialize NPC renderer
	npcRend := NewNPCRenderer(cfg.Genre, cfg.World.Seed)

	game := &Game{
		cfg:               cfg,
		world:             world,
		renderer:          renderer,
		client:            client,
		connected:         connected,
		playerEntity:      player,
		chunkManager:      chunkMgr,
		lodManager:        lodMgr,
		lastChunkX:        -999, // Force initial map build
		lastChunkY:        -999,
		mapSize:           localMapSize,
		wallThreshold:     wallThreshold,
		audioEngine:       audioEngine,
		audioPlayer:       audioPlayer,
		inputManager:      inputMgr,
		gameState:         GameStateCharacterCreation,
		characterCreation: NewCharacterCreation(),
		lightingSystem:    lightingSys,
		postprocessPipe:   postprocessPipeline,
		particleSystem:    particleSys,
		particleRenderer:  particleRend,
		spriteGenerator:   spriteGen,
		spriteCache:       spriteCache,
		textureCache:      textureCache,
		subtitleSystem:    subtitleSys,
		ambientMixer:      ambientMix,
		adaptiveMusic:     adaptiveMusic,
		currentRegion:     ambient.RegionPlains, // Default starting region
		interactionSys:    interactionSys,
		dialogUI:          dialogUI,
		combatManager:     combatMgr,
		menu:              menu,
		inventoryUI:       inventoryUI,
		questUI:           questUI,
		craftingUI:        craftingUI,
		factionUI:         factionUI,
		housingUI:         housingUI,
		pvpUI:             pvpUI,
		npcRenderer:       npcRend,
		frameTimeHistory:  make([]float64, 60), // Track last 60 frame times
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

	// Initialize pre-allocated UI rendering images
	game.initUIBuffers()

	ebiten.SetWindowSize(cfg.Window.Width, cfg.Window.Height)
	ebiten.SetWindowTitle(cfg.Window.Title)

	runGame(game, connected, client)
}

// createPlayerEntity creates and configures the player entity.
func createPlayerEntity(world *ecs.World) ecs.Entity {
	player := world.CreateEntity()
	addPlayerComponents(world, player)
	return player
}

// addPlayerComponents adds all required components to the player entity.
func addPlayerComponents(world *ecs.World, player ecs.Entity) {
	componentList := []ecs.Component{
		&components.Position{X: 8.5, Y: 8.5, Z: 0},
		&components.Health{Current: 100, Max: 100},
		&components.Mana{Current: 50, Max: 50, RegenRate: 1.0},
		&components.Skills{
			Levels:        make(map[string]int),
			Experience:    make(map[string]float64),
			SchoolBonuses: make(map[string]float64),
		},
		&components.Inventory{Items: []string{}, Capacity: 30},
		&components.Faction{ID: "player", Reputation: 0},
		&components.Reputation{Standings: make(map[string]float64)},
		&components.Stealth{
			Visibility:      1.0,
			BaseVisibility:  1.0,
			SneakVisibility: 0.3,
			DetectionRadius: 15.0,
		},
		&components.CombatState{},
		&components.AudioListener{Volume: 1.0, Enabled: true},
		&components.Weapon{
			Name:        "Fists",
			Damage:      5,
			Range:       1.5,
			AttackSpeed: 1.0,
			WeaponType:  "melee",
		},
	}

	for _, c := range componentList {
		if err := world.AddComponent(player, c); err != nil {
			log.Fatalf("failed to add %s component: %v", c.Type(), err)
		}
	}
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
