//go:build !noebiten

// Package main provides the inventory UI overlay for the Wyrm client.
package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/input"
)

// InventoryUI manages the inventory overlay display and interaction.
type InventoryUI struct {
	isOpen        bool
	playerEntity  ecs.Entity
	inputManager  *input.Manager
	selectedSlot  int
	gridCols      int
	gridRows      int
	slotSize      int
	padding       int
	scrollOffset  int
	confirmAction string // "use" or "drop" when confirming
	genre         string
}

// NewInventoryUI creates a new inventory UI.
func NewInventoryUI(genre string, playerEntity ecs.Entity, inputManager *input.Manager) *InventoryUI {
	return &InventoryUI{
		isOpen:       false,
		playerEntity: playerEntity,
		inputManager: inputManager,
		selectedSlot: 0,
		gridCols:     6, // 6 columns in the grid
		gridRows:     5, // 5 rows visible
		slotSize:     48,
		padding:      8,
		scrollOffset: 0,
		genre:        genre,
	}
}

// IsOpen returns whether the inventory UI is currently open.
func (ui *InventoryUI) IsOpen() bool {
	return ui.isOpen
}

// Open opens the inventory UI.
func (ui *InventoryUI) Open() {
	ui.isOpen = true
	ui.selectedSlot = 0
	ui.scrollOffset = 0
	ui.confirmAction = ""
}

// Close closes the inventory UI.
func (ui *InventoryUI) Close() {
	ui.isOpen = false
	ui.confirmAction = ""
}

// Toggle toggles the inventory UI open/closed state.
func (ui *InventoryUI) Toggle() {
	if ui.isOpen {
		ui.Close()
	} else {
		ui.Open()
	}
}

// Update handles input for the inventory UI.
func (ui *InventoryUI) Update(world *ecs.World) {
	if !ui.isOpen {
		return
	}

	if ui.handleConfirmationDialog(world) {
		return
	}

	if ui.handleCloseInput() {
		return
	}

	inv := ui.getInventory(world)
	if inv == nil {
		return
	}

	itemCount := len(inv.Items)
	ui.handleNavigationInput(itemCount)
	ui.handleMouseInput(itemCount)
	ui.handleItemActionInput(itemCount)
	ui.updateScrollOffset(itemCount)
}

// handleConfirmationDialog processes Y/N confirmation dialogs.
func (ui *InventoryUI) handleConfirmationDialog(world *ecs.World) bool {
	if ui.confirmAction == "" {
		return false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyY) {
		ui.executeConfirmedAction(world)
		ui.confirmAction = ""
	} else if inpututil.IsKeyJustPressed(ebiten.KeyN) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.confirmAction = ""
	}
	return true
}

// handleCloseInput checks for close key presses.
func (ui *InventoryUI) handleCloseInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyI) {
		ui.Close()
		return true
	}
	return false
}

// handleNavigationInput processes arrow key navigation.
func (ui *InventoryUI) handleNavigationInput(itemCount int) {
	maxSlot := itemCount - 1
	if maxSlot < 0 {
		maxSlot = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		ui.selectedSlot -= ui.gridCols
		if ui.selectedSlot < 0 {
			ui.selectedSlot = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		ui.selectedSlot += ui.gridCols
		if ui.selectedSlot > maxSlot {
			ui.selectedSlot = maxSlot
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) && ui.selectedSlot > 0 {
		ui.selectedSlot--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) && ui.selectedSlot < maxSlot {
		ui.selectedSlot++
	}
}

// handleMouseInput processes mouse click for slot selection.
func (ui *InventoryUI) handleMouseInput(itemCount int) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	mx, my := ebiten.CursorPosition()
	clickedSlot := ui.getSlotAtPosition(mx, my)
	if clickedSlot >= 0 && clickedSlot < itemCount {
		ui.selectedSlot = clickedSlot
	}
}

// handleItemActionInput processes use/drop item key presses.
func (ui *InventoryUI) handleItemActionInput(itemCount int) {
	if ui.selectedSlot >= itemCount {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyU) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		ui.confirmAction = "use"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		ui.confirmAction = "drop"
	}
}

// updateScrollOffset adjusts scroll to keep selected slot visible.
func (ui *InventoryUI) updateScrollOffset(itemCount int) {
	slotsPerPage := ui.gridCols * ui.gridRows
	selectedRow := ui.selectedSlot / ui.gridCols

	// Scroll down if selection is below visible area
	minVisibleRow := ui.scrollOffset
	maxVisibleRow := ui.scrollOffset + ui.gridRows - 1

	if selectedRow < minVisibleRow {
		ui.scrollOffset = selectedRow
	} else if selectedRow > maxVisibleRow {
		ui.scrollOffset = selectedRow - ui.gridRows + 1
	}

	// Clamp scroll offset
	maxScroll := (itemCount+ui.gridCols-1)/ui.gridCols - ui.gridRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if ui.scrollOffset > maxScroll {
		ui.scrollOffset = maxScroll
	}
	if ui.scrollOffset < 0 {
		ui.scrollOffset = 0
	}
	_ = slotsPerPage // suppress unused variable warning
}

// getSlotAtPosition returns the slot index at the given screen position.
func (ui *InventoryUI) getSlotAtPosition(x, y int) int {
	// Calculate grid position
	screenWidth, screenHeight := 1280, 720 // default dimensions
	gridWidth := ui.gridCols*ui.slotSize + (ui.gridCols-1)*ui.padding
	gridHeight := ui.gridRows*ui.slotSize + (ui.gridRows-1)*ui.padding
	startX := (screenWidth - gridWidth) / 2
	startY := (screenHeight-gridHeight)/2 + 60 // offset for header

	// Check if click is within grid
	if x < startX || y < startY {
		return -1
	}

	col := (x - startX) / (ui.slotSize + ui.padding)
	row := (y - startY) / (ui.slotSize + ui.padding)

	if col >= ui.gridCols || row >= ui.gridRows {
		return -1
	}

	return (row+ui.scrollOffset)*ui.gridCols + col
}

// executeConfirmedAction performs the confirmed use or drop action.
func (ui *InventoryUI) executeConfirmedAction(world *ecs.World) {
	inv := ui.getInventory(world)
	if inv == nil || ui.selectedSlot >= len(inv.Items) {
		return
	}

	switch ui.confirmAction {
	case "use":
		ui.useItem(world, ui.selectedSlot)
	case "drop":
		ui.dropItem(world, ui.selectedSlot)
	}
}

// useItem uses the item at the given slot.
func (ui *InventoryUI) useItem(world *ecs.World, slot int) {
	inv := ui.getInventory(world)
	if inv == nil || slot >= len(inv.Items) {
		return
	}

	itemName := inv.Items[slot]

	// Check for consumables (health potions, mana potions, etc.)
	if isConsumable(itemName) {
		ui.consumeItem(world, itemName)
		// Remove item from inventory
		inv.Items = append(inv.Items[:slot], inv.Items[slot+1:]...)
		if ui.selectedSlot >= len(inv.Items) && ui.selectedSlot > 0 {
			ui.selectedSlot--
		}
	}
	// Equipment would go here in a future implementation
}

// consumeItem applies the effects of a consumable item.
func (ui *InventoryUI) consumeItem(world *ecs.World, itemName string) {
	healthComp, ok := world.GetComponent(ui.playerEntity, "Health")
	if !ok {
		return
	}
	health := healthComp.(*components.Health)

	manaComp, _ := world.GetComponent(ui.playerEntity, "Mana")
	var mana *components.Mana
	if manaComp != nil {
		mana = manaComp.(*components.Mana)
	}

	// Apply effects based on item name
	switch {
	case containsStr(itemName, "health") || containsStr(itemName, "potion"):
		health.Current += 25
		if health.Current > health.Max {
			health.Current = health.Max
		}
	case containsStr(itemName, "mana") && mana != nil:
		mana.Current += 25
		if mana.Current > mana.Max {
			mana.Current = mana.Max
		}
	}
}

// dropItem drops the item at the given slot.
func (ui *InventoryUI) dropItem(world *ecs.World, slot int) {
	inv := ui.getInventory(world)
	if inv == nil || slot >= len(inv.Items) {
		return
	}

	// Remove item from inventory
	// In a full implementation, we would spawn an item entity in the world
	inv.Items = append(inv.Items[:slot], inv.Items[slot+1:]...)
	if ui.selectedSlot >= len(inv.Items) && ui.selectedSlot > 0 {
		ui.selectedSlot--
	}
}

// getInventory retrieves the player's inventory component.
func (ui *InventoryUI) getInventory(world *ecs.World) *components.Inventory {
	comp, ok := world.GetComponent(ui.playerEntity, "Inventory")
	if !ok {
		return nil
	}
	return comp.(*components.Inventory)
}

// Draw renders the inventory UI overlay.
func (ui *InventoryUI) Draw(screen *ebiten.Image, world *ecs.World) {
	if !ui.isOpen {
		return
	}

	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()

	// Draw semi-transparent background
	ui.drawBackground(screen, screenWidth, screenHeight)

	// Draw title
	ui.drawTitle(screen, screenWidth)

	// Get inventory data
	inv := ui.getInventory(world)
	weapon := ui.getWeapon(world)

	// Draw equipment slots
	ui.drawEquipmentSlots(screen, screenWidth, weapon)

	// Draw inventory grid
	ui.drawInventoryGrid(screen, screenWidth, screenHeight, inv)

	// Draw capacity indicator
	ui.drawCapacityIndicator(screen, screenWidth, screenHeight, inv)

	// Draw confirmation dialog if active
	if ui.confirmAction != "" {
		ui.drawConfirmDialog(screen, screenWidth, screenHeight, inv)
	}

	// Draw controls help
	ui.drawControlsHelp(screen, screenHeight)
}

// drawBackground draws the semi-transparent overlay background.
func (ui *InventoryUI) drawBackground(screen *ebiten.Image, width, height int) {
	bgColor := color.RGBA{R: 20, G: 20, B: 30, A: 220}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			screen.Set(x, y, bgColor)
		}
	}
}

// drawTitle draws the inventory title.
func (ui *InventoryUI) drawTitle(screen *ebiten.Image, screenWidth int) {
	title := "INVENTORY"
	titleX := (screenWidth - len(title)*6) / 2
	ebitenutil.DebugPrintAt(screen, title, titleX, 20)
}

// drawEquipmentSlots draws the equipment slot display.
func (ui *InventoryUI) drawEquipmentSlots(screen *ebiten.Image, screenWidth int, weapon *components.Weapon) {
	startY := 50

	// Weapon slot
	weaponText := "Weapon: [Empty]"
	if weapon != nil {
		weaponText = fmt.Sprintf("Weapon: %s (Dmg: %.0f)", weapon.Name, weapon.Damage)
	}
	ebitenutil.DebugPrintAt(screen, weaponText, 20, startY)

	// Armor slot (placeholder)
	ebitenutil.DebugPrintAt(screen, "Armor: [Empty]", 20, startY+16)

	// Accessory slot (placeholder)
	ebitenutil.DebugPrintAt(screen, "Accessory: [Empty]", 20, startY+32)
}

// drawInventoryGrid draws the main inventory item grid.
func (ui *InventoryUI) drawInventoryGrid(screen *ebiten.Image, screenWidth, screenHeight int, inv *components.Inventory) {
	gridWidth := ui.gridCols*ui.slotSize + (ui.gridCols-1)*ui.padding
	gridHeight := ui.gridRows*ui.slotSize + (ui.gridRows-1)*ui.padding
	startX := (screenWidth - gridWidth) / 2
	startY := 120

	// Draw grid background
	gridBgColor := color.RGBA{R: 40, G: 40, B: 50, A: 255}
	for y := startY - 5; y < startY+gridHeight+5; y++ {
		for x := startX - 5; x < startX+gridWidth+5; x++ {
			screen.Set(x, y, gridBgColor)
		}
	}

	// Draw slots
	itemCount := 0
	if inv != nil {
		itemCount = len(inv.Items)
	}

	for row := 0; row < ui.gridRows; row++ {
		for col := 0; col < ui.gridCols; col++ {
			slotIndex := (row+ui.scrollOffset)*ui.gridCols + col
			slotX := startX + col*(ui.slotSize+ui.padding)
			slotY := startY + row*(ui.slotSize+ui.padding)

			// Slot background
			slotColor := color.RGBA{R: 60, G: 60, B: 70, A: 255}
			if slotIndex == ui.selectedSlot {
				slotColor = color.RGBA{R: 100, G: 150, B: 200, A: 255} // Highlighted
			}

			for y := slotY; y < slotY+ui.slotSize; y++ {
				for x := slotX; x < slotX+ui.slotSize; x++ {
					screen.Set(x, y, slotColor)
				}
			}

			// Draw item if present
			if inv != nil && slotIndex < itemCount {
				itemName := inv.Items[slotIndex]
				// Draw item icon (just first letter for now)
				if len(itemName) > 0 {
					iconText := string(itemName[0])
					ebitenutil.DebugPrintAt(screen, iconText, slotX+ui.slotSize/2-3, slotY+ui.slotSize/2-8)
				}
			}
		}
	}

	// Draw selected item name below grid
	if inv != nil && ui.selectedSlot < itemCount {
		itemName := inv.Items[ui.selectedSlot]
		nameX := (screenWidth - len(itemName)*6) / 2
		ebitenutil.DebugPrintAt(screen, itemName, nameX, startY+gridHeight+15)
	}
}

// drawCapacityIndicator draws the weight/capacity indicator.
func (ui *InventoryUI) drawCapacityIndicator(screen *ebiten.Image, screenWidth, screenHeight int, inv *components.Inventory) {
	itemCount := 0
	capacity := 30 // default
	if inv != nil {
		itemCount = len(inv.Items)
		if inv.Capacity > 0 {
			capacity = inv.Capacity
		}
	}

	capacityText := fmt.Sprintf("Items: %d / %d", itemCount, capacity)
	capX := screenWidth - len(capacityText)*6 - 20
	ebitenutil.DebugPrintAt(screen, capacityText, capX, screenHeight-40)

	// Draw capacity bar
	barWidth := 100
	barHeight := 10
	barX := screenWidth - barWidth - 20
	barY := screenHeight - 25

	// Background
	for y := barY; y < barY+barHeight; y++ {
		for x := barX; x < barX+barWidth; x++ {
			screen.Set(x, y, color.RGBA{R: 50, G: 50, B: 50, A: 255})
		}
	}

	// Fill
	fillWidth := barWidth * itemCount / capacity
	if fillWidth > barWidth {
		fillWidth = barWidth
	}
	fillColor := color.RGBA{R: 100, G: 200, B: 100, A: 255}
	if itemCount > capacity*80/100 {
		fillColor = color.RGBA{R: 200, G: 200, B: 100, A: 255}
	}
	if itemCount >= capacity {
		fillColor = color.RGBA{R: 200, G: 100, B: 100, A: 255}
	}
	for y := barY; y < barY+barHeight; y++ {
		for x := barX; x < barX+fillWidth; x++ {
			screen.Set(x, y, fillColor)
		}
	}
}

// drawConfirmDialog draws a confirmation dialog for use/drop actions.
func (ui *InventoryUI) drawConfirmDialog(screen *ebiten.Image, screenWidth, screenHeight int, inv *components.Inventory) {
	if inv == nil || ui.selectedSlot >= len(inv.Items) {
		return
	}

	itemName := inv.Items[ui.selectedSlot]
	dialogWidth := 300
	dialogHeight := 80
	dialogX := (screenWidth - dialogWidth) / 2
	dialogY := (screenHeight - dialogHeight) / 2

	// Dialog background
	for y := dialogY; y < dialogY+dialogHeight; y++ {
		for x := dialogX; x < dialogX+dialogWidth; x++ {
			screen.Set(x, y, color.RGBA{R: 30, G: 30, B: 40, A: 250})
		}
	}

	// Dialog text
	actionText := ui.confirmAction
	promptText := fmt.Sprintf("%s %s?", actionText, itemName)
	promptX := dialogX + (dialogWidth-len(promptText)*6)/2
	ebitenutil.DebugPrintAt(screen, promptText, promptX, dialogY+20)

	confirmText := "Y = Yes   N = No"
	confirmX := dialogX + (dialogWidth-len(confirmText)*6)/2
	ebitenutil.DebugPrintAt(screen, confirmText, confirmX, dialogY+50)
}

// drawControlsHelp draws the controls help text.
func (ui *InventoryUI) drawControlsHelp(screen *ebiten.Image, screenHeight int) {
	helpText := "Arrows: Navigate | U/Enter: Use | Del: Drop | I/Esc: Close"
	ebitenutil.DebugPrintAt(screen, helpText, 20, screenHeight-20)
}

// getWeapon retrieves the player's weapon component.
func (ui *InventoryUI) getWeapon(world *ecs.World) *components.Weapon {
	comp, ok := world.GetComponent(ui.playerEntity, "Weapon")
	if !ok {
		return nil
	}
	return comp.(*components.Weapon)
}

// isConsumable checks if an item is a consumable.
func isConsumable(itemName string) bool {
	return containsStr(itemName, "potion") ||
		containsStr(itemName, "food") ||
		containsStr(itemName, "drink") ||
		containsStr(itemName, "health") ||
		containsStr(itemName, "mana")
}

// containsStr checks if a string contains a substring (case-insensitive).
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			// Simple lowercase comparison
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
