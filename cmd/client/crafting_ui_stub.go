//go:build noebiten

// Package main provides stub crafting UI for noebiten builds.
package main

import (
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/input"
)

// CraftingUI stub for noebiten builds.
type CraftingUI struct {
	isOpen           bool
	playerEntity     ecs.Entity
	workbenchEntity  ecs.Entity
	selectedCategory int
	selectedRecipe   int
	scrollOffset     int
	categories       []string
}

// NewCraftingUI creates a stub crafting UI.
func NewCraftingUI(genre string, playerEntity ecs.Entity, inputManager *input.Manager) *CraftingUI {
	return &CraftingUI{
		isOpen:       false,
		playerEntity: playerEntity,
		categories:   []string{"All", "Weapons", "Armor", "Tools", "Consumables"},
	}
}

// IsOpen returns whether the crafting UI is open.
func (ui *CraftingUI) IsOpen() bool {
	return ui.isOpen
}

// OpenWorkbench opens the crafting UI for a workbench.
func (ui *CraftingUI) OpenWorkbench(world *ecs.World, workbench ecs.Entity) {
	ui.isOpen = true
	ui.workbenchEntity = workbench
}

// Close closes the crafting UI.
func (ui *CraftingUI) Close() {
	ui.isOpen = false
}

// Update is a no-op for stub.
func (ui *CraftingUI) Update(world *ecs.World) {}

// RecipeInfo holds recipe display information (uses shared recipeInfo internally).
type RecipeInfo = recipeInfo

// getKnownRecipes retrieves recipes the player knows filtered by category.
func (ui *CraftingUI) getKnownRecipes(world *ecs.World) []RecipeInfo {
	return getKnownRecipesShared(world, ui.playerEntity, ui.selectedCategory, ui.categories, ui.getRecipeInfo)
}

// getRecipeInfo returns recipe display info based on ID patterns.
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
