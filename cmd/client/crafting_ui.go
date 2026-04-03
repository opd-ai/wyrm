//go:build !noebiten

// Package main provides the crafting UI overlay for the Wyrm client.
package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// CraftingUI manages the crafting overlay display and interaction.
type CraftingUI struct {
	isOpen           bool
	playerEntity     ecs.Entity
	workbenchEntity  ecs.Entity
	selectedRecipe   int
	selectedCategory int
	categories       []string
	scrollOffset     int
	genre            string
}

// NewCraftingUI creates a new crafting UI.
func NewCraftingUI(genre string, playerEntity ecs.Entity) *CraftingUI {
	return &CraftingUI{
		isOpen:           false,
		playerEntity:     playerEntity,
		selectedRecipe:   0,
		selectedCategory: 0,
		categories:       []string{"All", "Weapons", "Armor", "Tools", "Consumables"},
		scrollOffset:     0,
		genre:            genre,
	}
}

// IsOpen returns whether the crafting UI is currently open.
func (ui *CraftingUI) IsOpen() bool {
	return ui.isOpen
}

// OpenWorkbench opens the crafting UI for a specific workbench.
func (ui *CraftingUI) OpenWorkbench(world *ecs.World, workbench ecs.Entity) {
	ui.isOpen = true
	ui.workbenchEntity = workbench
	ui.selectedRecipe = 0
	ui.selectedCategory = 0
	ui.scrollOffset = 0
}

// Close closes the crafting UI.
func (ui *CraftingUI) Close() {
	ui.isOpen = false
}

// Update handles input for the crafting UI.
func (ui *CraftingUI) Update(world *ecs.World) {
	if !ui.isOpen {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.Close()
		return
	}
	ui.handleRecipeNavigation()
	ui.handleCategoryNavigation()
	ui.handleCraftAction(world)
}

// handleRecipeNavigation processes up/down navigation through recipes.
func (ui *CraftingUI) handleRecipeNavigation() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		ui.selectedRecipe--
		if ui.selectedRecipe < 0 {
			ui.selectedRecipe = 0
		}
		ui.adjustScroll()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		ui.selectedRecipe++
		ui.adjustScroll()
	}
}

// handleCategoryNavigation processes left/right navigation between categories.
func (ui *CraftingUI) handleCategoryNavigation() {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		ui.selectedCategory--
		if ui.selectedCategory < 0 {
			ui.selectedCategory = len(ui.categories) - 1
		}
		ui.resetRecipeSelection()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		ui.selectedCategory++
		if ui.selectedCategory >= len(ui.categories) {
			ui.selectedCategory = 0
		}
		ui.resetRecipeSelection()
	}
}

// resetRecipeSelection resets recipe selection when changing categories.
func (ui *CraftingUI) resetRecipeSelection() {
	ui.selectedRecipe = 0
	ui.scrollOffset = 0
}

// handleCraftAction initiates crafting when action key is pressed.
func (ui *CraftingUI) handleCraftAction(world *ecs.World) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyE) {
		ui.startCraft(world)
	}
}

// adjustScroll adjusts scroll offset to keep selected recipe visible.
func (ui *CraftingUI) adjustScroll() {
	visibleRows := 8
	if ui.selectedRecipe < ui.scrollOffset {
		ui.scrollOffset = ui.selectedRecipe
	}
	if ui.selectedRecipe >= ui.scrollOffset+visibleRows {
		ui.scrollOffset = ui.selectedRecipe - visibleRows + 1
	}
}

// startCraft attempts to start crafting the selected recipe.
func (ui *CraftingUI) startCraft(world *ecs.World) {
	recipes := ui.getKnownRecipes(world)
	if ui.selectedRecipe < 0 || ui.selectedRecipe >= len(recipes) {
		return
	}
	recipe := recipes[ui.selectedRecipe]

	// Check if we have materials
	if !ui.hasMaterials(world, recipe) {
		return
	}

	// Get crafting state component
	craftComp, ok := world.GetComponent(ui.playerEntity, "CraftingState")
	if !ok {
		return
	}
	craftState := craftComp.(*components.CraftingState)
	if craftState.IsCrafting {
		return // Already crafting
	}

	// Consume materials
	ui.consumeMaterials(world, recipe)

	// Start crafting via component state
	craftTime := ui.getCraftTime(recipe)
	craftState.IsCrafting = true
	craftState.CurrentRecipeID = recipe.ID
	craftState.Progress = 0
	craftState.TotalTime = craftTime
	craftState.WorkbenchEntity = uint64(ui.workbenchEntity)
	craftState.ConsumedMaterials = make(map[string]int)
	for mat, qty := range recipe.Materials {
		craftState.ConsumedMaterials[mat] = qty
	}
}

// getKnownRecipes retrieves recipes the player knows filtered by category.
func (ui *CraftingUI) getKnownRecipes(world *ecs.World) []RecipeInfo {
	knowledgeComp, ok := world.GetComponent(ui.playerEntity, "RecipeKnowledge")
	if !ok {
		return nil
	}
	knowledge := knowledgeComp.(*components.RecipeKnowledge)

	var recipes []RecipeInfo
	for recipeID := range knowledge.KnownRecipes {
		if !knowledge.KnownRecipes[recipeID] {
			continue
		}
		info := ui.getRecipeInfo(recipeID)
		if ui.matchesCategory(info) {
			recipes = append(recipes, info)
		}
	}
	return recipes
}

// RecipeInfo holds display information for a recipe (alias to shared type).
type RecipeInfo = recipeInfo

// getRecipeInfo returns recipe display info (would normally come from a registry).
func (ui *CraftingUI) getRecipeInfo(recipeID string) RecipeInfo {
	return getRecipeInfoShared(recipeID)
}

// contains checks if the string contains any of the substrings.
func contains(s string, subs ...string) bool {
	return containsAny(s, subs...)
}

// matchesCategory checks if recipe matches current category filter.
func (ui *CraftingUI) matchesCategory(info RecipeInfo) bool {
	return matchesCategoryShared(info, ui.selectedCategory, ui.categories)
}

// hasMaterials checks if player has required materials.
func (ui *CraftingUI) hasMaterials(world *ecs.World, recipe RecipeInfo) bool {
	return hasMaterialsShared(world, ui.playerEntity, recipe)
}

// consumeMaterials removes materials from inventory.
func (ui *CraftingUI) consumeMaterials(world *ecs.World, recipe RecipeInfo) {
	consumeMaterialsShared(world, ui.playerEntity, recipe)
}

// getCraftTime returns the crafting time for a recipe.
func (ui *CraftingUI) getCraftTime(recipe RecipeInfo) float64 {
	if recipe.CraftTime > 0 {
		return recipe.CraftTime
	}
	return 5.0 // Default 5 seconds
}

// Draw renders the crafting UI.
func (ui *CraftingUI) Draw(screen *ebiten.Image, world *ecs.World) {
	if !ui.isOpen {
		return
	}

	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Draw semi-transparent background
	bgColor := color.RGBA{0, 0, 0, 180}
	panelW, panelH := 500, 400
	panelX := (screenW - panelW) / 2
	panelY := (screenH - panelH) / 2

	ebitenutil.DrawRect(screen, float64(panelX), float64(panelY), float64(panelW), float64(panelH), bgColor)

	// Draw border
	borderColor := ui.getGenreColor()
	ebitenutil.DrawRect(screen, float64(panelX), float64(panelY), float64(panelW), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(panelX), float64(panelY+panelH-2), float64(panelW), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(panelX), float64(panelY), 2, float64(panelH), borderColor)
	ebitenutil.DrawRect(screen, float64(panelX+panelW-2), float64(panelY), 2, float64(panelH), borderColor)

	// Draw title
	title := ui.getWorkbenchTitle(world)
	ebitenutil.DebugPrintAt(screen, title, panelX+10, panelY+10)

	// Draw categories
	catY := panelY + 30
	for i, cat := range ui.categories {
		catStr := cat
		if i == ui.selectedCategory {
			catStr = "[" + cat + "]"
		}
		ebitenutil.DebugPrintAt(screen, catStr, panelX+10+(i*80), catY)
	}

	// Draw recipe list
	recipes := ui.getKnownRecipes(world)
	listY := catY + 25
	visibleRows := 8
	for i := ui.scrollOffset; i < len(recipes) && i < ui.scrollOffset+visibleRows; i++ {
		recipe := recipes[i]
		y := listY + (i-ui.scrollOffset)*20

		// Highlight selected
		if i == ui.selectedRecipe {
			ebitenutil.DrawRect(screen, float64(panelX+10), float64(y-2), float64(panelW-20), 18, color.RGBA{80, 80, 80, 200})
		}

		// Check if craftable
		canCraft := ui.hasMaterials(world, recipe)
		nameStr := recipe.Name
		if !canCraft {
			nameStr = "(!) " + recipe.Name
		}
		ebitenutil.DebugPrintAt(screen, nameStr, panelX+15, y)
	}

	// Draw selected recipe details
	if ui.selectedRecipe >= 0 && ui.selectedRecipe < len(recipes) {
		recipe := recipes[ui.selectedRecipe]
		detailY := panelY + panelH - 100

		ebitenutil.DebugPrintAt(screen, "Materials:", panelX+10, detailY)
		matY := detailY + 15
		for mat, qty := range recipe.Materials {
			matStr := fmt.Sprintf("  %s x%d", mat, qty)
			ebitenutil.DebugPrintAt(screen, matStr, panelX+10, matY)
			matY += 15
		}
	}

	// Draw crafting progress if active
	ui.drawCraftingProgress(screen, world, panelX, panelY, panelW, panelH)

	// Draw help text
	helpText := "[UP/DOWN] Select  [LEFT/RIGHT] Category  [ENTER] Craft  [ESC] Close"
	ebitenutil.DebugPrintAt(screen, helpText, panelX+10, panelY+panelH-20)
}

// drawCraftingProgress draws the crafting progress bar.
func (ui *CraftingUI) drawCraftingProgress(screen *ebiten.Image, world *ecs.World, panelX, panelY, panelW, panelH int) {
	craftComp, ok := world.GetComponent(ui.playerEntity, "CraftingState")
	if !ok {
		return
	}
	craftState := craftComp.(*components.CraftingState)
	if !craftState.IsCrafting {
		return
	}

	// Draw progress bar
	barX := panelX + 10
	barY := panelY + panelH - 50
	barW := panelW - 20
	barH := 20

	// Background
	ebitenutil.DrawRect(screen, float64(barX), float64(barY), float64(barW), float64(barH), color.RGBA{40, 40, 40, 255})

	// Progress fill
	fillW := int(float64(barW) * craftState.Progress)
	ebitenutil.DrawRect(screen, float64(barX), float64(barY), float64(fillW), float64(barH), ui.getGenreColor())

	// Progress text
	progressStr := fmt.Sprintf("Crafting: %s (%.0f%%)", craftState.CurrentRecipeID, craftState.Progress*100)
	ebitenutil.DebugPrintAt(screen, progressStr, barX+5, barY+3)
}

// getWorkbenchTitle returns the title for the current workbench.
func (ui *CraftingUI) getWorkbenchTitle(world *ecs.World) string {
	wbComp, ok := world.GetComponent(ui.workbenchEntity, "Workbench")
	if !ok {
		return "Crafting"
	}
	wb := wbComp.(*components.Workbench)
	if wb.WorkbenchType != "" {
		return fmt.Sprintf("Crafting - %s", wb.WorkbenchType)
	}
	return "Crafting"
}

// getGenreColor returns the accent color for the current genre.
func (ui *CraftingUI) getGenreColor() color.RGBA {
	switch ui.genre {
	case "fantasy":
		return color.RGBA{218, 165, 32, 255} // Gold
	case "sci-fi":
		return color.RGBA{0, 191, 255, 255} // Deep sky blue
	case "horror":
		return color.RGBA{139, 0, 0, 255} // Dark red
	case "cyberpunk":
		return color.RGBA{255, 0, 255, 255} // Magenta
	case "post-apocalyptic":
		return color.RGBA{210, 105, 30, 255} // Chocolate
	default:
		return color.RGBA{100, 100, 100, 255}
	}
}
