//go:build noebiten

// Package main provides stub crafting UI for noebiten builds.
package main

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
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

// RecipeInfo holds recipe display information.
type RecipeInfo struct {
	ID          string
	Name        string
	Category    string
	Description string
	Materials   map[string]int
	OutputItem  string
	OutputQty   int
	CraftTime   float64
	SkillReq    map[string]int
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

// getRecipeInfo returns recipe display info based on ID patterns.
func (ui *CraftingUI) getRecipeInfo(recipeID string) RecipeInfo {
	info := RecipeInfo{
		ID:          recipeID,
		Name:        recipeID,
		Category:    "All",
		Description: "A crafted item",
		Materials:   map[string]int{"material": 1},
		OutputItem:  recipeID,
		OutputQty:   1,
		CraftTime:   5.0,
		SkillReq:    map[string]int{},
	}

	switch {
	case contains(recipeID, "sword", "axe", "bow", "dagger", "spear", "mace"):
		info.Category = "Weapons"
	case contains(recipeID, "helm", "chest", "legs", "boots", "gloves", "shield"):
		info.Category = "Armor"
	case contains(recipeID, "pick", "hammer", "saw", "needle", "tool"):
		info.Category = "Tools"
	case contains(recipeID, "potion", "food", "bandage", "elixir"):
		info.Category = "Consumables"
	}

	return info
}

// contains checks if the string contains any of the substrings.
func contains(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// matchesCategory checks if recipe matches current category filter.
func (ui *CraftingUI) matchesCategory(info RecipeInfo) bool {
	if ui.selectedCategory == 0 { // "All"
		return true
	}
	return info.Category == ui.categories[ui.selectedCategory]
}

// hasMaterials checks if player has required materials.
func (ui *CraftingUI) hasMaterials(world *ecs.World, recipe RecipeInfo) bool {
	invComp, ok := world.GetComponent(ui.playerEntity, "Inventory")
	if !ok {
		return false
	}
	inv := invComp.(*components.Inventory)

	itemCounts := make(map[string]int)
	for _, item := range inv.Items {
		itemCounts[item]++
	}

	for mat, needed := range recipe.Materials {
		if itemCounts[mat] < needed {
			return false
		}
	}
	return true
}

// consumeMaterials removes materials from inventory.
func (ui *CraftingUI) consumeMaterials(world *ecs.World, recipe RecipeInfo) {
	invComp, ok := world.GetComponent(ui.playerEntity, "Inventory")
	if !ok {
		return
	}
	inv := invComp.(*components.Inventory)

	for mat, needed := range recipe.Materials {
		removed := 0
		newItems := make([]string, 0, len(inv.Items))
		for _, item := range inv.Items {
			if item == mat && removed < needed {
				removed++
				continue
			}
			newItems = append(newItems, item)
		}
		inv.Items = newItems
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
