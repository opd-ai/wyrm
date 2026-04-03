//go:build !noebiten

// Command client launches the Wyrm game client with an Ebitengine window.
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	_ "net/http/pprof"
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

// Timing constants for game loop and simulation.
const (
	targetFPS        = 60.0       // target frames per second
	defaultDeltaTime = 1.0 / 60.0 // default delta time for fixed timestep
	secondsPerMinute = 60.0       // conversion factor for world clock
	interactionRange = 2.0        // maximum range for player interactions in units
	defaultDuration  = 2.0        // default duration for effects/animations in seconds
	sneakBaseVis     = 0.3        // base visibility when sneaking
	baseWantedLevel  = 100.0      // base wanted level threshold
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
	chunkMapInit  bool // Whether the chunk map has been initialized
	mapSize       int  // Size of the local map grid
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
	interactionSys       *InteractionSystem
	interactionThisFrame bool // Whether E key was used for interaction this frame
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
	dt := defaultDeltaTime // fallback to 60 FPS
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
	g.handleInteraction() // Check for interaction first to set interactionThisFrame
	g.handlePlayerInput(dt)
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
// Sets interactionThisFrame flag when E key is consumed for interaction.
func (g *Game) handleInteraction() {
	g.interactionThisFrame = false // Reset each frame
	if g.interactionSys == nil {
		return
	}

	// Update interaction system to find current target
	g.interactionSys.Update(g.world)

	// Check if interaction key is pressed and there's a valid target
	if g.isActionOrKeyPressed(input.ActionInteract, ebiten.KeyE) {
		target := g.interactionSys.GetCurrentTarget()
		if target != nil {
			g.processInteraction(target)
			g.interactionThisFrame = true // Consume the E key this frame
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

	// Sync skybox with ECS world state (weather and time of day)
	g.syncSkyboxWithWorld()

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

// syncSkyboxWithWorld updates the renderer's skybox based on ECS weather and clock state.
func (g *Game) syncSkyboxWithWorld() {
	if g.renderer == nil || g.renderer.Skybox == nil {
		return
	}

	skybox := g.renderer.Skybox

	// Find Weather component and sync current weather
	for _, e := range g.world.Entities("Weather") {
		weatherComp, ok := g.world.GetComponent(e, "Weather")
		if ok {
			weather := weatherComp.(*components.Weather)
			skybox.SetWeather(weather.WeatherType, weather.CloudCover)
			break // Only need one weather entity
		}
	}

	// Find WorldClock and sync time of day
	for _, e := range g.world.Entities("WorldClock") {
		clockComp, ok := g.world.GetComponent(e, "WorldClock")
		if ok {
			clock := clockComp.(*components.WorldClock)
			// Convert hour (0-23) to normalized time (0.0-1.0)
			// 0.0 = midnight, 0.5 = noon
			// TimeAccum accumulates seconds toward next hour change
			hourFraction := 0.0
			if clock.HourLength > 0 {
				hourFraction = clock.TimeAccum / clock.HourLength
			}
			timeOfDay := (float64(clock.Hour) + hourFraction) / 24.0
			skybox.SetTimeOfDay(timeOfDay)
			break
		}
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
	// Only rebuild map when entering a new chunk or on first run
	if g.chunkMapInit && chunkX == g.lastChunkX && chunkY == g.lastChunkY {
		return
	}
	g.lastChunkX = chunkX
	g.lastChunkY = chunkY
	g.chunkMapInit = true
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

	moveForward := g.gatherMovementForward()
	moveRight := g.gatherMovementStrafe()
	turn := g.gatherTurnInput()
	jump := g.isActionOrKeyPressed(input.ActionJump, ebiten.KeySpace)
	attack := g.isActionOrMousePressed(input.ActionAttack, ebiten.MouseButtonLeft)
	use := g.isActionOrMousePressed(input.ActionBlock, ebiten.MouseButtonRight)

	if g.hasActiveInput(moveForward, moveRight, turn, jump, attack, use) {
		if err := g.stateSync.SendPlayerInput(moveForward, moveRight, turn, jump, attack, use); err != nil {
			log.Printf("failed to send player input: %v", err)
		}
	}
}

// gatherMovementForward returns forward/backward movement input (-1, 0, or 1).
func (g *Game) gatherMovementForward() float32 {
	if g.isActionOrKeyPressed(input.ActionMoveForward, ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		return 1.0
	}
	if g.isActionOrKeyPressed(input.ActionMoveBackward, ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		return -1.0
	}
	return 0
}

// gatherMovementStrafe returns strafe input (-1 for left, 1 for right, 0 otherwise).
// Skips strafe right if E was consumed for interaction this frame.
func (g *Game) gatherMovementStrafe() float32 {
	if g.isActionOrKeyPressed(input.ActionStrafeLeft, ebiten.KeyQ) {
		return -1.0
	}
	// Don't strafe right on E if it was used for interaction
	if !g.interactionThisFrame && g.isActionOrKeyPressed(input.ActionStrafeRight, ebiten.KeyE) {
		return 1.0
	}
	return 0
}

// gatherTurnInput returns turn input from keyboard.
// Returns ±1.0 representing desired turn direction; server scales by tick rate.
func (g *Game) gatherTurnInput() float32 {
	if g.isActionOrKeyPressed(input.ActionMoveLeft, ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		return -1.0
	}
	if g.isActionOrKeyPressed(input.ActionMoveRight, ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		return 1.0
	}
	return 0
}

// hasActiveInput returns true if any input is active.
func (g *Game) hasActiveInput(moveForward, moveRight, turn float32, jump, attack, use bool) bool {
	return moveForward != 0 || moveRight != 0 || turn != 0 || jump || attack || use
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

	deltaX, deltaY, ok := g.calculateMouseDelta()
	if !ok {
		return
	}

	deltaX, deltaY = g.applyMouseModifiers(deltaX, deltaY)

	g.updatePlayerYaw(pos, deltaX)
	g.updateRendererPitch(deltaY)
}

// calculateMouseDelta computes the raw mouse movement delta.
// Returns false if mouse is not yet initialized or no movement occurred.
func (g *Game) calculateMouseDelta() (deltaX, deltaY float64, ok bool) {
	cursorX, cursorY := ebiten.CursorPosition()

	if !g.mouseInitialized {
		g.lastMouseX = cursorX
		g.lastMouseY = cursorY
		g.mouseInitialized = true
		return 0, 0, false
	}

	deltaX = float64(cursorX - g.lastMouseX)
	deltaY = float64(cursorY - g.lastMouseY)
	g.lastMouseX = cursorX
	g.lastMouseY = cursorY

	if deltaX == 0 && deltaY == 0 {
		return 0, 0, false
	}
	return deltaX, deltaY, true
}

// applyMouseModifiers applies acceleration, sensitivity, smoothing, and inversion.
func (g *Game) applyMouseModifiers(deltaX, deltaY float64) (float64, float64) {
	if g.cfg.Mouse.AccelerationOn {
		magnitude := math.Sqrt(deltaX*deltaX + deltaY*deltaY)
		accelerationFactor := 1.0 + (magnitude * g.cfg.Mouse.Acceleration * 0.01)
		deltaX *= accelerationFactor
		deltaY *= accelerationFactor
	}

	sensitivity := g.cfg.Mouse.Sensitivity * 0.005
	deltaX *= sensitivity
	deltaY *= sensitivity

	if g.cfg.Mouse.SmoothingOn {
		factor := g.cfg.Mouse.SmoothingFactor
		// Clamp smoothing factor to valid range [0, 1]
		if factor < 0 {
			factor = 0
		} else if factor > 1 {
			factor = 1
		}
		g.smoothedDeltaX = g.smoothedDeltaX*(1-factor) + deltaX*factor
		g.smoothedDeltaY = g.smoothedDeltaY*(1-factor) + deltaY*factor
		// Apply dead-zone to prevent phantom drift when mouse is stationary
		if math.Abs(g.smoothedDeltaX) < 0.001 {
			g.smoothedDeltaX = 0
		}
		if math.Abs(g.smoothedDeltaY) < 0.001 {
			g.smoothedDeltaY = 0
		}
		deltaX = g.smoothedDeltaX
		deltaY = g.smoothedDeltaY
	}

	if g.cfg.Mouse.InvertY {
		deltaY = -deltaY
	}
	return deltaX, deltaY
}

// updatePlayerYaw updates the player's horizontal angle and normalizes to [-π, π].
func (g *Game) updatePlayerYaw(pos *components.Position, deltaX float64) {
	pos.Angle += deltaX
	for pos.Angle > math.Pi {
		pos.Angle -= 2 * math.Pi
	}
	for pos.Angle < -math.Pi {
		pos.Angle += 2 * math.Pi
	}
}

// updateRendererPitch updates vertical look angle with clamping.
func (g *Game) updateRendererPitch(deltaY float64) {
	if g.renderer == nil {
		return
	}
	g.renderer.PlayerPitch -= deltaY
	if g.renderer.PlayerPitch > raycast.MaxPitchAngle {
		g.renderer.PlayerPitch = raycast.MaxPitchAngle
	}
	if g.renderer.PlayerPitch < -raycast.MaxPitchAngle {
		g.renderer.PlayerPitch = -raycast.MaxPitchAngle
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
// Uses DrawRect overlay instead of framebuffer modification to preserve sprites.
func (g *Game) applyCombatVisualFeedback(screen *ebiten.Image) {
	if g.combatManager == nil {
		return
	}

	// Apply damage flash as semi-transparent red overlay
	flashAlpha := g.combatManager.GetDamageFlashAlpha()
	if flashAlpha > 0 {
		screenWidth := screen.Bounds().Dx()
		screenHeight := screen.Bounds().Dy()
		// Draw red overlay with calculated alpha
		ebitenutil.DrawRect(screen, 0, 0, float64(screenWidth), float64(screenHeight),
			color.RGBA{255, 0, 0, uint8(flashAlpha)})
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
// Uses DrawRect for overlay, then ebitenutil for text.
func (g *Game) drawDeathScreen(screen *ebiten.Image) {
	if g.combatManager == nil || !g.combatManager.IsDead() {
		return
	}

	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()

	// Apply dark overlay using DrawRect (preserves previously drawn content)
	ebitenutil.DrawRect(screen, 0, 0, float64(screenWidth), float64(screenHeight),
		color.RGBA{0, 0, 0, 180})

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
		return defaultDeltaTime // Default to 60 FPS
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

// Layout returns the game's logical screen dimensions.
// Returns fixed dimensions matching the renderer's initialized size to prevent
// framebuffer/screen size mismatch on window resize.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.cfg.Window.Width, g.cfg.Window.Height
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
	interactionSys := NewInteractionSystem(player, interactionRange)

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
		lastChunkX:        0,
		lastChunkY:        0,
		chunkMapInit:      false, // Triggers initial map build on first update
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
