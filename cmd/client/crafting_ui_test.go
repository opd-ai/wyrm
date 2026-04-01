//go:build noebiten

package main

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/input"
)

func TestNewCraftingUI(t *testing.T) {
	inputMgr := input.NewManager()
	ui := NewCraftingUI("fantasy", 1, inputMgr)
	if ui == nil {
		t.Fatal("NewCraftingUI returned nil")
	}
	if ui.IsOpen() {
		t.Error("CraftingUI should start closed")
	}
}

func TestCraftingUI_OpenWorkbench(t *testing.T) {
	w := ecs.NewWorld()
	inputMgr := input.NewManager()
	player := w.CreateEntity()
	workbench := w.CreateEntity()

	ui := NewCraftingUI("fantasy", player, inputMgr)
	ui.OpenWorkbench(w, workbench)

	if !ui.IsOpen() {
		t.Error("CraftingUI should be open after OpenWorkbench")
	}
}

func TestCraftingUI_Close(t *testing.T) {
	w := ecs.NewWorld()
	inputMgr := input.NewManager()
	player := w.CreateEntity()
	workbench := w.CreateEntity()

	ui := NewCraftingUI("fantasy", player, inputMgr)
	ui.OpenWorkbench(w, workbench)
	ui.Close()

	if ui.IsOpen() {
		t.Error("CraftingUI should be closed after Close")
	}
}

func TestCraftingUI_GetKnownRecipes(t *testing.T) {
	w := ecs.NewWorld()
	inputMgr := input.NewManager()
	player := w.CreateEntity()

	// Add recipe knowledge
	knowledge := &components.RecipeKnowledge{
		KnownRecipes: map[string]bool{
			"iron_sword": true,
			"steel_helm": true,
		},
	}
	w.AddComponent(player, knowledge)

	ui := NewCraftingUI("fantasy", player, inputMgr)
	recipes := ui.getKnownRecipes(w)

	if len(recipes) != 2 {
		t.Errorf("Expected 2 recipes, got %d", len(recipes))
	}
}

func TestCraftingUI_HasMaterials(t *testing.T) {
	w := ecs.NewWorld()
	inputMgr := input.NewManager()
	player := w.CreateEntity()

	// Add inventory with materials
	inv := &components.Inventory{
		Items: []string{"iron_ore", "iron_ore", "wood"},
	}
	w.AddComponent(player, inv)

	ui := NewCraftingUI("fantasy", player, inputMgr)

	recipe := RecipeInfo{
		ID:        "test",
		Materials: map[string]int{"iron_ore": 2},
	}

	if !ui.hasMaterials(w, recipe) {
		t.Error("Should have materials for recipe")
	}

	insufficientRecipe := RecipeInfo{
		ID:        "test2",
		Materials: map[string]int{"iron_ore": 5},
	}

	if ui.hasMaterials(w, insufficientRecipe) {
		t.Error("Should not have enough materials")
	}
}

func TestCraftingUI_ConsumeMaterials(t *testing.T) {
	w := ecs.NewWorld()
	inputMgr := input.NewManager()
	player := w.CreateEntity()

	inv := &components.Inventory{
		Items: []string{"iron_ore", "iron_ore", "wood", "iron_ore"},
	}
	w.AddComponent(player, inv)

	ui := NewCraftingUI("fantasy", player, inputMgr)

	recipe := RecipeInfo{
		ID:        "test",
		Materials: map[string]int{"iron_ore": 2},
	}

	ui.consumeMaterials(w, recipe)

	// Count remaining iron_ore
	ironCount := 0
	for _, item := range inv.Items {
		if item == "iron_ore" {
			ironCount++
		}
	}

	if ironCount != 1 {
		t.Errorf("Expected 1 iron_ore remaining, got %d", ironCount)
	}
}

func TestCraftingUI_CategoryFilter(t *testing.T) {
	inputMgr := input.NewManager()
	ui := NewCraftingUI("fantasy", 1, inputMgr)

	// Test "All" category (index 0)
	ui.selectedCategory = 0
	weaponRecipe := RecipeInfo{Category: "Weapons"}
	armorRecipe := RecipeInfo{Category: "Armor"}

	if !ui.matchesCategory(weaponRecipe) {
		t.Error("All category should match weapons")
	}
	if !ui.matchesCategory(armorRecipe) {
		t.Error("All category should match armor")
	}

	// Test "Weapons" category (index 1)
	ui.selectedCategory = 1
	if !ui.matchesCategory(weaponRecipe) {
		t.Error("Weapons category should match weapons")
	}
	if ui.matchesCategory(armorRecipe) {
		t.Error("Weapons category should not match armor")
	}
}

func TestCraftingUI_GetRecipeInfo(t *testing.T) {
	inputMgr := input.NewManager()
	ui := NewCraftingUI("fantasy", 1, inputMgr)

	// Test weapon categorization
	swordInfo := ui.getRecipeInfo("iron_sword")
	if swordInfo.Category != "Weapons" {
		t.Errorf("Expected Weapons category for sword, got %s", swordInfo.Category)
	}

	// Test armor categorization
	helmInfo := ui.getRecipeInfo("steel_helm")
	if helmInfo.Category != "Armor" {
		t.Errorf("Expected Armor category for helm, got %s", helmInfo.Category)
	}

	// Test consumable categorization
	potionInfo := ui.getRecipeInfo("health_potion")
	if potionInfo.Category != "Consumables" {
		t.Errorf("Expected Consumables category for potion, got %s", potionInfo.Category)
	}
}

func TestCraftingUI_AdjustScroll(t *testing.T) {
	inputMgr := input.NewManager()
	ui := NewCraftingUI("fantasy", 1, inputMgr)

	// Test scrolling down
	ui.selectedRecipe = 10
	ui.scrollOffset = 0
	ui.adjustScroll()

	if ui.scrollOffset == 0 {
		t.Error("Scroll should have adjusted for recipe 10")
	}

	// Test scrolling up
	ui.selectedRecipe = 0
	ui.adjustScroll()
	if ui.scrollOffset != 0 {
		t.Error("Scroll should reset to 0 for recipe 0")
	}
}
