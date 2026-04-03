// crafting_shared.go provides shared crafting logic used by both
// ebiten and noebiten builds.
package main

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// recipeInfo holds display information for a recipe.
type recipeInfo struct {
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

// getKnownRecipesShared retrieves recipes the player knows filtered by category.
func getKnownRecipesShared(world *ecs.World, playerEntity ecs.Entity, selectedCategory int, categories []string, getRecipeInfo func(string) recipeInfo) []recipeInfo {
	knowledgeComp, ok := world.GetComponent(playerEntity, "RecipeKnowledge")
	if !ok {
		return nil
	}
	knowledge := knowledgeComp.(*components.RecipeKnowledge)

	var recipes []recipeInfo
	for recipeID := range knowledge.KnownRecipes {
		if !knowledge.KnownRecipes[recipeID] {
			continue
		}
		info := getRecipeInfo(recipeID)
		if matchesCategoryShared(info, selectedCategory, categories) {
			recipes = append(recipes, info)
		}
	}
	return recipes
}

// getRecipeInfoShared returns recipe display info based on ID patterns.
func getRecipeInfoShared(recipeID string) recipeInfo {
	info := recipeInfo{
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
	case containsAny(recipeID, "sword", "axe", "bow", "dagger", "spear", "mace"):
		info.Category = "Weapons"
	case containsAny(recipeID, "helm", "chest", "legs", "boots", "gloves", "shield"):
		info.Category = "Armor"
	case containsAny(recipeID, "pick", "hammer", "saw", "needle", "tool"):
		info.Category = "Tools"
	case containsAny(recipeID, "potion", "food", "bandage", "elixir"):
		info.Category = "Consumables"
	}

	return info
}

// containsAny checks if the string contains any of the substrings.
func containsAny(s string, subs ...string) bool {
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

// matchesCategoryShared checks if recipe matches current category filter.
func matchesCategoryShared(info recipeInfo, selectedCategory int, categories []string) bool {
	if selectedCategory == 0 { // "All"
		return true
	}
	return info.Category == categories[selectedCategory]
}

// hasMaterialsShared checks if player has required materials.
func hasMaterialsShared(world *ecs.World, playerEntity ecs.Entity, recipe recipeInfo) bool {
	invComp, ok := world.GetComponent(playerEntity, "Inventory")
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

// consumeMaterialsShared removes materials from inventory.
func consumeMaterialsShared(world *ecs.World, playerEntity ecs.Entity, recipe recipeInfo) {
	invComp, ok := world.GetComponent(playerEntity, "Inventory")
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
