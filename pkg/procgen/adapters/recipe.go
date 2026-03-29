// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
)

// RecipeAdapter wraps Venture's recipe generator for Wyrm's crafting system.
type RecipeAdapter struct {
	generator *recipe.RecipeGenerator
}

// NewRecipeAdapter creates a new recipe adapter.
func NewRecipeAdapter() *RecipeAdapter {
	return &RecipeAdapter{
		generator: recipe.NewRecipeGenerator(),
	}
}

// RecipeData holds recipe information adapted for Wyrm's crafting system.
type RecipeData struct {
	ID                string
	Name              string
	Description       string
	Type              string
	Rarity            string
	Materials         []MaterialData
	GoldCost          int
	SkillRequired     int
	BaseSuccessChance float64
	CraftTimeSec      float64
	OutputItemSeed    int64
	OutputItemType    string
	Genre             string
}

// MaterialData holds material requirement information adapted for Wyrm.
type MaterialData struct {
	ItemName string
	Quantity int
	Optional bool
}

// GenerateRecipes generates a batch of recipes for a crafting station or shop.
func (a *RecipeAdapter) GenerateRecipes(seed int64, genre string, depth, count int, recipeType string) ([]*RecipeData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: float64(depth) / 100.0,
		Depth:      depth,
		Custom: map[string]interface{}{
			"count": count,
			"type":  recipeType,
		},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("recipe generation failed: %w", err)
	}

	recipes, ok := result.([]*engine.Recipe)
	if !ok {
		return nil, fmt.Errorf("invalid recipe result type: expected []*engine.Recipe, got %T", result)
	}

	return convertRecipes(recipes), nil
}

// GeneratePotionRecipes generates potion/consumable recipes.
func (a *RecipeAdapter) GeneratePotionRecipes(seed int64, genre string, depth, count int) ([]*RecipeData, error) {
	return a.GenerateRecipes(seed, genre, depth, count, "potion")
}

// GenerateEnchantingRecipes generates enchanting recipes.
func (a *RecipeAdapter) GenerateEnchantingRecipes(seed int64, genre string, depth, count int) ([]*RecipeData, error) {
	return a.GenerateRecipes(seed, genre, depth, count, "enchanting")
}

// GenerateMagicItemRecipes generates magic item recipes.
func (a *RecipeAdapter) GenerateMagicItemRecipes(seed int64, genre string, depth, count int) ([]*RecipeData, error) {
	return a.GenerateRecipes(seed, genre, depth, count, "magic_item")
}

// GenerateWorkbenchRecipes generates all recipe types for a workbench.
func (a *RecipeAdapter) GenerateWorkbenchRecipes(seed int64, genre string, craftingSkill int) ([]*RecipeData, error) {
	var allRecipes []*RecipeData

	// Generate potions (alchemist's table)
	potions, err := a.GeneratePotionRecipes(seed, genre, craftingSkill, 5)
	if err == nil {
		allRecipes = append(allRecipes, potions...)
	}

	// Generate enchantments (enchanting table)
	enchantments, err := a.GenerateEnchantingRecipes(seed+1, genre, craftingSkill, 3)
	if err == nil {
		allRecipes = append(allRecipes, enchantments...)
	}

	// Generate magic items (arcane workbench)
	magicItems, err := a.GenerateMagicItemRecipes(seed+2, genre, craftingSkill, 2)
	if err == nil {
		allRecipes = append(allRecipes, magicItems...)
	}

	return allRecipes, nil
}

// convertRecipes converts Venture recipes to Wyrm format.
func convertRecipes(recipes []*engine.Recipe) []*RecipeData {
	result := make([]*RecipeData, len(recipes))
	for i, r := range recipes {
		result[i] = convertRecipe(r)
	}
	return result
}

// convertRecipe converts a single Venture recipe to Wyrm format.
func convertRecipe(r *engine.Recipe) *RecipeData {
	materials := make([]MaterialData, len(r.Materials))
	for i, m := range r.Materials {
		materials[i] = MaterialData{
			ItemName: m.ItemName,
			Quantity: m.Quantity,
			Optional: m.Optional,
		}
	}

	return &RecipeData{
		ID:                r.ID,
		Name:              r.Name,
		Description:       r.Description,
		Type:              r.Type.String(),
		Rarity:            r.Rarity.String(),
		Materials:         materials,
		GoldCost:          r.GoldCost,
		SkillRequired:     r.SkillRequired,
		BaseSuccessChance: r.BaseSuccessChance,
		CraftTimeSec:      r.CraftTimeSec,
		OutputItemSeed:    r.OutputItemSeed,
		OutputItemType:    r.OutputItemType.String(),
		Genre:             r.GenreID,
	}
}

// GetEffectiveSuccessChance calculates the success chance for a recipe at a given skill.
func GetEffectiveSuccessChance(recipe *RecipeData, skillLevel int) float64 {
	// Base chance increases by 2% per skill level above requirement
	if skillLevel < recipe.SkillRequired {
		return recipe.BaseSuccessChance * 0.5 // 50% penalty if under-skilled
	}
	bonus := float64(skillLevel-recipe.SkillRequired) * 0.02
	chance := recipe.BaseSuccessChance + bonus
	if chance > 0.99 {
		chance = 0.99 // Never 100% guaranteed
	}
	return chance
}

// CanCraft checks if a player has the requirements to craft a recipe.
func CanCraft(recipe *RecipeData, playerSkill, playerGold int, inventory map[string]int) (bool, string) {
	if playerSkill < recipe.SkillRequired-3 {
		return false, "skill too low"
	}
	if playerGold < recipe.GoldCost {
		return false, "insufficient gold"
	}
	for _, mat := range recipe.Materials {
		if mat.Optional {
			continue
		}
		if inventory[mat.ItemName] < mat.Quantity {
			return false, "missing material: " + mat.ItemName
		}
	}
	return true, ""
}
