//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import "math/rand"

// RecipeAdapter wraps Venture's recipe generator.
// Stub implementation for headless testing.
type RecipeAdapter struct{}

// NewRecipeAdapter creates a new recipe adapter.
func NewRecipeAdapter() *RecipeAdapter { return &RecipeAdapter{} }

// RecipeData holds generated recipe information.
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

// MaterialData holds recipe material requirements.
type MaterialData struct {
	ItemName string
	Quantity int
	Optional bool
}

// GenerateRecipes creates multiple recipes.
func (a *RecipeAdapter) GenerateRecipes(seed int64, genre string, depth, count int, recipeType string) ([]*RecipeData, error) {
	recipes := make([]*RecipeData, count)
	for i := 0; i < count; i++ {
		rng := rand.New(rand.NewSource(seed + int64(i)))
		matCount := 2 + rng.Intn(3)
		materials := make([]MaterialData, matCount)
		for j := 0; j < matCount; j++ {
			materials[j] = MaterialData{
				ItemName: "material_" + string(rune('A'+j)),
				Quantity: 1 + rng.Intn(5),
				Optional: j > 1 && rng.Float64() > 0.7,
			}
		}
		recipes[i] = &RecipeData{
			ID:                "recipe_" + string(rune('A'+i)),
			Name:              recipeType + " recipe",
			Description:       "A " + recipeType + " recipe",
			Type:              recipeType,
			Rarity:            "common",
			Materials:         materials,
			GoldCost:          10 + rng.Intn(100),
			SkillRequired:     depth * 5,
			BaseSuccessChance: 0.5 + float64(depth)*0.05,
			CraftTimeSec:      5.0 + rng.Float64()*10,
			OutputItemSeed:    seed + int64(i)*100,
			OutputItemType:    recipeType + "_output",
			Genre:             genre,
		}
	}
	return recipes, nil
}

// GeneratePotionRecipes creates potion recipes.
func (a *RecipeAdapter) GeneratePotionRecipes(seed int64, genre string, depth, count int) ([]*RecipeData, error) {
	return a.GenerateRecipes(seed, genre, depth, count, "potion")
}

// GenerateEnchantingRecipes creates enchanting recipes.
func (a *RecipeAdapter) GenerateEnchantingRecipes(seed int64, genre string, depth, count int) ([]*RecipeData, error) {
	return a.GenerateRecipes(seed, genre, depth, count, "enchant")
}

// GenerateMagicItemRecipes creates magic item recipes.
func (a *RecipeAdapter) GenerateMagicItemRecipes(seed int64, genre string, depth, count int) ([]*RecipeData, error) {
	return a.GenerateRecipes(seed, genre, depth, count, "magic_item")
}

// GenerateWorkbenchRecipes creates workbench recipes based on skill.
func (a *RecipeAdapter) GenerateWorkbenchRecipes(seed int64, genre string, craftingSkill int) ([]*RecipeData, error) {
	count := 5 + craftingSkill/10
	return a.GenerateRecipes(seed, genre, craftingSkill/5, count, "craft")
}

// GetEffectiveSuccessChance calculates actual success chance.
func GetEffectiveSuccessChance(recipe *RecipeData, skillLevel int) float64 {
	bonus := float64(skillLevel-recipe.SkillRequired) * 0.02
	chance := recipe.BaseSuccessChance + bonus
	if chance > 1.0 {
		return 1.0
	}
	if chance < 0.1 {
		return 0.1
	}
	return chance
}

// CanCraft checks if player can craft a recipe.
func CanCraft(recipe *RecipeData, playerSkill, playerGold int, inventory map[string]int) (bool, string) {
	if playerSkill < recipe.SkillRequired {
		return false, "insufficient skill"
	}
	if playerGold < recipe.GoldCost {
		return false, "insufficient gold"
	}
	for _, mat := range recipe.Materials {
		if mat.Optional {
			continue
		}
		if inventory[mat.ItemName] < mat.Quantity {
			return false, "missing materials: " + mat.ItemName
		}
	}
	return true, ""
}
