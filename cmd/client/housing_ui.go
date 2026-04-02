//go:build !noebiten

// Package main provides the Wyrm game client with housing UI.
package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/world/housing"
)

// HousingUI manages player housing purchase and furniture placement UI.
type HousingUI struct {
	active             bool
	mode               HousingUIMode
	houseManager       *housing.HouseManager
	propertyMarket     *housing.PropertyMarket
	furniturePlacement *housing.FurniturePlacement
	guildManager       *housing.GuildManager

	// Property listing selection
	listings      []*housing.PropertyListing
	selectedIndex int

	// Player state
	playerGold   int
	playerEntity uint64

	// Owned houses
	playerHouses []*housing.House
	houseIndex   int

	// Furniture types available
	furnitureTypes []string
	furnitureIndex int

	// Current world day for purchases
	currentDay int
}

// HousingUIMode indicates what screen is showing.
type HousingUIMode int

const (
	HousingModeNone      HousingUIMode = iota
	HousingModeListings                // Browsing property listings
	HousingModeMyHouses                // Viewing owned houses
	HousingModeFurniture               // Placing furniture
	HousingModeGuild                   // Guild territory management
)

// NewHousingUI creates a new housing UI.
func NewHousingUI() *HousingUI {
	return &HousingUI{
		active:             false,
		mode:               HousingModeNone,
		houseManager:       housing.NewHouseManager(),
		guildManager:       housing.NewGuildManager(),
		furniturePlacement: housing.NewFurniturePlacement(),
		furnitureTypes:     []string{"bed", "table", "chair", "chest", "bookshelf", "lamp"},
	}
}

// Initialize sets up property market with house manager reference.
func (ui *HousingUI) Initialize() {
	ui.propertyMarket = housing.NewPropertyMarket(ui.houseManager)

	// Add some sample listings
	ui.propertyMarket.AddListing(&housing.PropertyListing{
		ID:          "house-001",
		Name:        "Small Cottage",
		Description: "A cozy starter home",
		WorldX:      50,
		WorldZ:      50,
		BasePrice:   1000,
		Size:        1,
		Quality:     0.5,
		DistrictID:  "residential",
		Genre:       "fantasy",
	})

	ui.propertyMarket.AddListing(&housing.PropertyListing{
		ID:          "house-002",
		Name:        "Town House",
		Description: "A comfortable mid-sized home",
		WorldX:      100,
		WorldZ:      75,
		BasePrice:   3000,
		Size:        2,
		Quality:     0.7,
		DistrictID:  "residential",
		Genre:       "fantasy",
	})

	ui.propertyMarket.AddListing(&housing.PropertyListing{
		ID:          "house-003",
		Name:        "Grand Manor",
		Description: "A luxurious estate",
		WorldX:      200,
		WorldZ:      150,
		BasePrice:   10000,
		Size:        3,
		Quality:     0.9,
		DistrictID:  "noble",
		Genre:       "fantasy",
	})
}

// SetPlayerState updates player information for purchases.
func (ui *HousingUI) SetPlayerState(entity uint64, gold, day int) {
	ui.playerEntity = entity
	ui.playerGold = gold
	ui.currentDay = day
}

// IsActive returns whether the housing UI is open.
func (ui *HousingUI) IsActive() bool {
	return ui.active
}

// Open opens the housing UI in listings mode.
func (ui *HousingUI) Open() {
	ui.active = true
	ui.mode = HousingModeListings
	ui.refreshListings()
}

// OpenFurnitureMode enters furniture placement mode.
func (ui *HousingUI) OpenFurnitureMode(houseID string) {
	ui.active = true
	ui.mode = HousingModeFurniture
	ui.furniturePlacement.StartPlaceMode(houseID, ui.furnitureTypes[ui.furnitureIndex])
}

// Close closes the housing UI.
func (ui *HousingUI) Close() {
	ui.active = false
	ui.mode = HousingModeNone
	ui.furniturePlacement.ExitMode()
}

// refreshListings updates the available property listings.
func (ui *HousingUI) refreshListings() {
	ui.listings = ui.propertyMarket.GetAvailableListings()
	ui.selectedIndex = 0
}

// refreshPlayerHouses updates the player's owned houses.
func (ui *HousingUI) refreshPlayerHouses() {
	ui.playerHouses = ui.houseManager.GetPlayerHouses(ui.playerEntity)
	ui.houseIndex = 0
}

// Update handles input for the housing UI.
func (ui *HousingUI) Update() {
	if !ui.active {
		return
	}

	// Tab to switch modes
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		switch ui.mode {
		case HousingModeListings:
			ui.mode = HousingModeMyHouses
			ui.refreshPlayerHouses()
		case HousingModeMyHouses:
			ui.mode = HousingModeGuild
		case HousingModeGuild:
			ui.mode = HousingModeListings
			ui.refreshListings()
		case HousingModeFurniture:
			// Stay in furniture mode until explicit exit
		}
	}

	// Escape to close
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.Close()
		return
	}

	switch ui.mode {
	case HousingModeListings:
		ui.updateListingsMode()
	case HousingModeMyHouses:
		ui.updateMyHousesMode()
	case HousingModeFurniture:
		ui.updateFurnitureMode()
	case HousingModeGuild:
		ui.updateGuildMode()
	}
}

// updateListingsMode handles property listing browsing.
func (ui *HousingUI) updateListingsMode() {
	// Navigate listings
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		if ui.selectedIndex > 0 {
			ui.selectedIndex--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		if ui.selectedIndex < len(ui.listings)-1 {
			ui.selectedIndex++
		}
	}

	// Purchase property
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(ui.listings) > 0 {
		listing := ui.listings[ui.selectedIndex]
		result := ui.propertyMarket.PurchaseProperty(
			listing.ID,
			ui.playerEntity,
			ui.playerGold,
			ui.currentDay,
		)
		if result.Success {
			ui.playerGold -= result.PricePaid
			ui.refreshListings()
		}
	}
}

// updateMyHousesMode handles owned houses browsing.
func (ui *HousingUI) updateMyHousesMode() {
	// Navigate houses
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		if ui.houseIndex > 0 {
			ui.houseIndex--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		if ui.houseIndex < len(ui.playerHouses)-1 {
			ui.houseIndex++
		}
	}

	// Enter furniture mode for selected house
	if inpututil.IsKeyJustPressed(ebiten.KeyF) && len(ui.playerHouses) > 0 {
		house := ui.playerHouses[ui.houseIndex]
		ui.OpenFurnitureMode(house.ID)
	}
}

// updateFurnitureMode handles furniture placement.
func (ui *HousingUI) updateFurnitureMode() {
	ui.handleFurnitureNavigation()
	ui.handleFurnitureActions()
}

// handleFurnitureNavigation processes left/right navigation between furniture types.
func (ui *HousingUI) handleFurnitureNavigation() {
	navigated := false
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) && ui.furnitureIndex > 0 {
		ui.furnitureIndex--
		navigated = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) && ui.furnitureIndex < len(ui.furnitureTypes)-1 {
		ui.furnitureIndex++
		navigated = true
	}
	if navigated {
		ui.furniturePlacement.StartPlaceMode(
			ui.furniturePlacement.GetCurrentHouse(),
			ui.furnitureTypes[ui.furnitureIndex],
		)
	}
}

// handleFurnitureActions processes furniture rotation, grid snap, placement, and exit.
func (ui *HousingUI) handleFurnitureActions() {
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		ui.furniturePlacement.RotatePreview(0.785) // 45 degrees
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		ui.toggleGridSnap()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		ui.confirmFurniturePlacement()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		ui.exitFurnitureMode()
	}
}

// toggleGridSnap enables grid snapping if placement mode is active.
func (ui *HousingUI) toggleGridSnap() {
	mode, _, _, _, _, _ := ui.furniturePlacement.GetPreviewState()
	if mode != housing.PlacementModeNone {
		ui.furniturePlacement.SetGridSnap(true, 0.5)
	}
}

// confirmFurniturePlacement places the furniture and restarts placement mode.
func (ui *HousingUI) confirmFurniturePlacement() {
	newID := fmt.Sprintf("furn-%d", ui.currentDay)
	err := ui.furniturePlacement.ConfirmPlacement(ui.houseManager, newID)
	if err == nil {
		ui.furniturePlacement.StartPlaceMode(
			ui.furniturePlacement.GetCurrentHouse(),
			ui.furnitureTypes[ui.furnitureIndex],
		)
	}
}

// exitFurnitureMode returns to house view and refreshes the house list.
func (ui *HousingUI) exitFurnitureMode() {
	ui.mode = HousingModeMyHouses
	ui.furniturePlacement.ExitMode()
	ui.refreshPlayerHouses()
}

// updateGuildMode handles guild territory management.
func (ui *HousingUI) updateGuildMode() {
	// Guild mode is simpler - just viewing
	// Full implementation would have claim/release options
}

// Draw renders the housing UI.
func (ui *HousingUI) Draw(screen *ebiten.Image) {
	if !ui.active {
		return
	}

	// Draw semi-transparent background
	ebitenutil.DrawRect(screen, 50, 50, 540, 380, color.RGBA{20, 20, 20, 220})
	ebitenutil.DrawRect(screen, 52, 52, 536, 376, color.RGBA{40, 40, 60, 255})

	// Draw title based on mode
	var title string
	switch ui.mode {
	case HousingModeListings:
		title = "Property Listings"
	case HousingModeMyHouses:
		title = "My Properties"
	case HousingModeFurniture:
		title = "Furniture Placement"
	case HousingModeGuild:
		title = "Guild Territories"
	}
	ebitenutil.DebugPrintAt(screen, title, 60, 60)

	// Draw mode-specific content
	switch ui.mode {
	case HousingModeListings:
		ui.drawListingsMode(screen)
	case HousingModeMyHouses:
		ui.drawMyHousesMode(screen)
	case HousingModeFurniture:
		ui.drawFurnitureMode(screen)
	case HousingModeGuild:
		ui.drawGuildMode(screen)
	}

	// Draw controls help
	ebitenutil.DebugPrintAt(screen, "[Tab] Switch Mode  [Esc] Close", 60, 410)
}

// drawListingsMode renders property listings.
func (ui *HousingUI) drawListingsMode(screen *ebiten.Image) {
	y := 90
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Gold: %d", ui.playerGold), 450, y)
	y += 25

	if len(ui.listings) == 0 {
		ebitenutil.DebugPrintAt(screen, "No properties available", 60, y)
		return
	}

	for i, listing := range ui.listings {
		if i == ui.selectedIndex {
			ebitenutil.DrawRect(screen, 55, float64(y-3), 530, 50, color.RGBA{60, 60, 80, 255})
		}

		price := ui.propertyMarket.GetCurrentPrice(listing.ID)
		ebitenutil.DebugPrintAt(screen, listing.Name, 60, y)
		ebitenutil.DebugPrintAt(screen, listing.Description, 60, y+15)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Price: %d gold", price), 60, y+30)

		if price > ui.playerGold {
			ebitenutil.DebugPrintAt(screen, "(Cannot afford)", 400, y)
		}

		y += 60
	}

	ebitenutil.DebugPrintAt(screen, "[Enter] Purchase  [Up/Down] Navigate", 60, 380)
}

// drawMyHousesMode renders owned houses.
func (ui *HousingUI) drawMyHousesMode(screen *ebiten.Image) {
	y := 90

	if len(ui.playerHouses) == 0 {
		ebitenutil.DebugPrintAt(screen, "You don't own any properties yet", 60, y)
		return
	}

	for i, house := range ui.playerHouses {
		if i == ui.houseIndex {
			ebitenutil.DrawRect(screen, 55, float64(y-3), 530, 50, color.RGBA{60, 60, 80, 255})
		}

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("House: %s", house.ID), 60, y)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Location: (%.0f, %.0f)", house.WorldX, house.WorldZ), 60, y+15)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Furniture: %d items", len(house.Furniture)), 60, y+30)

		y += 60
	}

	ebitenutil.DebugPrintAt(screen, "[F] Place Furniture  [Up/Down] Navigate", 60, 380)
}

// drawFurnitureMode renders furniture placement interface.
func (ui *HousingUI) drawFurnitureMode(screen *ebiten.Image) {
	y := 90

	// Show selected furniture type
	ebitenutil.DebugPrintAt(screen, "Selected Furniture:", 60, y)
	y += 20

	for i, ftype := range ui.furnitureTypes {
		prefix := "  "
		if i == ui.furnitureIndex {
			prefix = "> "
		}
		ebitenutil.DebugPrintAt(screen, prefix+ftype, 70+i*80, y)
	}
	y += 30

	// Show placement preview info
	mode, px, py, pz, rot, valid := ui.furniturePlacement.GetPreviewState()
	if mode != housing.PlacementModeNone {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Position: (%.1f, %.1f, %.1f)", px, py, pz), 60, y)
		y += 20
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rotation: %.0f deg", rot*57.3), 60, y)
		y += 20
		if valid {
			ebitenutil.DebugPrintAt(screen, "Position: Valid", 60, y)
		} else {
			ebitenutil.DebugPrintAt(screen, "Position: Invalid", 60, y)
		}
	}

	ebitenutil.DebugPrintAt(screen, "[Left/Right] Type  [R] Rotate  [Enter] Place  [Backspace] Exit", 60, 380)
}

// drawGuildMode renders guild territory information.
func (ui *HousingUI) drawGuildMode(screen *ebiten.Image) {
	y := 90

	territories := ui.guildManager.ExportTerritories()
	if len(territories) == 0 {
		ebitenutil.DebugPrintAt(screen, "No guild territories claimed", 60, y)
		return
	}

	for _, territory := range territories {
		ebitenutil.DebugPrintAt(screen, territory.Name, 60, y)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Guild: %s", territory.GuildID), 60, y+15)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Location: (%.0f, %.0f) R:%.0f", territory.CenterX, territory.CenterZ, territory.Radius), 60, y+30)
		y += 50
	}
}

// UpdateFurniturePreview updates furniture preview position from player view.
func (ui *HousingUI) UpdateFurniturePreview(playerX, playerZ, viewAngle float64) {
	if ui.mode == HousingModeFurniture {
		ui.furniturePlacement.UpdatePreview(playerX, playerZ, viewAngle, 2.0)
	}
}

// GetHouseManager returns the house manager for other systems.
func (ui *HousingUI) GetHouseManager() *housing.HouseManager {
	return ui.houseManager
}

// GetGuildManager returns the guild manager for other systems.
func (ui *HousingUI) GetGuildManager() *housing.GuildManager {
	return ui.guildManager
}

// IsInPlayerHouse checks if position is inside a player's house.
func (ui *HousingUI) IsInPlayerHouse(playerEntity uint64, x, z float64) bool {
	houses := ui.houseManager.GetPlayerHouses(playerEntity)
	for _, house := range houses {
		// Simple proximity check (houses are at specific world positions)
		dx := x - house.WorldX
		dz := z - house.WorldZ
		if dx*dx+dz*dz < 100 { // Within 10 units
			return true
		}
	}
	return false
}

// GetNearbyHouseForPurchase returns a property listing near the player position.
func (ui *HousingUI) GetNearbyHouseForPurchase(x, z float64) *housing.PropertyListing {
	for _, listing := range ui.propertyMarket.GetAvailableListings() {
		dx := x - listing.WorldX
		dz := z - listing.WorldZ
		if dx*dx+dz*dz < 100 { // Within 10 units
			return listing
		}
	}
	return nil
}

// Helper to get player gold from components.
func getPlayerGold(g *Game) int {
	if g.world == nil {
		return 0
	}
	currComp, ok := g.world.GetComponent(g.playerEntity, "Currency")
	if !ok {
		return 0
	}
	curr := currComp.(*components.Currency)
	return curr.Gold
}
