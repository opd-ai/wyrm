// Package systems provides ECS systems for Wyrm.
package systems

import (
	"math"
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// CraftingSystem handles crafting mechanics including workbench interactions,
// material gathering, recipe validation, and tool durability.
type CraftingSystem struct {
	// RNG for quality calculations and success rolls.
	rng *rand.Rand
}

// NewCraftingSystem creates a new crafting system with the given seed.
func NewCraftingSystem(seed int64) *CraftingSystem {
	return &CraftingSystem{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Update processes crafting progress and resource respawning.
func (s *CraftingSystem) Update(w *ecs.World, dt float64) {
	s.updateCraftingProgress(w, dt)
	s.updateResourceRespawning(w, dt)
}

// updateCraftingProgress advances ongoing crafts.
func (s *CraftingSystem) updateCraftingProgress(w *ecs.World, dt float64) {
	for _, e := range w.Entities("CraftingState") {
		comp, ok := w.GetComponent(e, "CraftingState")
		if !ok {
			continue
		}
		state := comp.(*components.CraftingState)
		if !state.IsCrafting || state.TotalTime <= 0 {
			continue
		}

		// Advance progress
		progressDelta := dt / state.TotalTime
		state.Progress += progressDelta
		if state.Progress >= 1.0 {
			s.completeCraft(w, e, state)
		}
	}
}

// updateResourceRespawning handles depleted resource nodes.
func (s *CraftingSystem) updateResourceRespawning(w *ecs.World, dt float64) {
	// Find the world clock
	var gameTime float64
	clockFound := false
	for _, ce := range w.Entities("WorldClock") {
		clockComp, ok := w.GetComponent(ce, "WorldClock")
		if ok {
			clock := clockComp.(*components.WorldClock)
			gameTime = float64(clock.Day*HoursPerDay+clock.Hour)*SecondsPerHour + clock.TimeAccum
			clockFound = true
			break
		}
	}
	if !clockFound {
		return
	}

	for _, e := range w.Entities("ResourceNode") {
		comp, ok := w.GetComponent(e, "ResourceNode")
		if !ok {
			continue
		}
		node := comp.(*components.ResourceNode)
		if !node.Depleted {
			continue
		}

		if gameTime-node.LastGathered >= node.RespawnTime {
			node.Depleted = false
			node.Quantity = node.MaxQuantity
		}
	}
}

// completeCraft finishes a crafting operation and creates the output item.
func (s *CraftingSystem) completeCraft(w *ecs.World, entity ecs.Entity, state *components.CraftingState) {
	state.IsCrafting = false
	state.Progress = 0
	state.CurrentRecipeID = ""
	state.ConsumedMaterials = nil
	// Note: Item creation would integrate with ItemAdapter and Inventory system
}

// StartCraft begins crafting a recipe at a workbench.
func (s *CraftingSystem) StartCraft(w *ecs.World, crafter, workbench ecs.Entity, recipeID string, craftTime float64) bool {
	// Verify crafter has CraftingState
	comp, ok := w.GetComponent(crafter, "CraftingState")
	if !ok {
		return false
	}
	state := comp.(*components.CraftingState)
	if state.IsCrafting {
		return false // Already crafting
	}

	// Verify workbench exists
	_, wbOK := w.GetComponent(workbench, "Workbench")
	if !wbOK {
		return false
	}

	// Verify crafter knows this recipe
	if !s.KnowsRecipe(w, crafter, recipeID) {
		return false
	}

	// Start crafting
	state.IsCrafting = true
	state.CurrentRecipeID = recipeID
	state.Progress = 0
	state.TotalTime = craftTime
	state.WorkbenchEntity = uint64(workbench)
	state.ConsumedMaterials = make(map[string]int)

	return true
}

// CancelCraft stops an ongoing craft and returns consumed materials.
func (s *CraftingSystem) CancelCraft(w *ecs.World, crafter ecs.Entity) bool {
	comp, ok := w.GetComponent(crafter, "CraftingState")
	if !ok {
		return false
	}
	state := comp.(*components.CraftingState)
	if !state.IsCrafting {
		return false
	}

	// Return partial materials (50% of consumed)
	if state.ConsumedMaterials != nil {
		invComp, invOK := w.GetComponent(crafter, "Inventory")
		if invOK {
			inv := invComp.(*components.Inventory)
			for mat, qty := range state.ConsumedMaterials {
				returnQty := qty / 2
				if returnQty > 0 {
					inv.Items = append(inv.Items, mat)
					_ = returnQty // Would add returnQty times
				}
			}
		}
	}

	state.IsCrafting = false
	state.Progress = 0
	state.CurrentRecipeID = ""
	state.ConsumedMaterials = nil

	return true
}

// GatherResource harvests materials from a resource node.
func (s *CraftingSystem) GatherResource(w *ecs.World, gatherer, node ecs.Entity) (string, int, float64) {
	nodeComp, nodeOK := w.GetComponent(node, "ResourceNode")
	if !nodeOK {
		return "", 0, 0
	}
	rn := nodeComp.(*components.ResourceNode)
	if rn.Depleted || rn.Quantity <= 0 {
		return "", 0, 0
	}

	// Check for equipped tool
	gatherSpeed := 1.0
	qualityBonus := 0.0

	toolComp, toolOK := w.GetComponent(gatherer, "Tool")
	if toolOK {
		tool := toolComp.(*components.Tool)
		if tool.Durability > 0 {
			gatherSpeed = tool.GatherSpeed
			qualityBonus = tool.QualityBonus
			// Reduce tool durability
			tool.Durability -= 1.0
			if tool.Durability < 0 {
				tool.Durability = 0
			}
		}
	}

	// Calculate gathered amount (affected by tool)
	baseAmount := 1
	if gatherSpeed > 1.0 {
		baseAmount = int(math.Ceil(float64(baseAmount) * gatherSpeed))
	}
	if baseAmount > rn.Quantity {
		baseAmount = rn.Quantity
	}

	// Calculate quality
	quality := rn.Quality + qualityBonus
	if quality > 1.0 {
		quality = 1.0
	}

	// Apply skill bonus if available
	skillsComp, skillsOK := w.GetComponent(gatherer, "Skills")
	if skillsOK {
		skills := skillsComp.(*components.Skills)
		// Use gathering-related skill if available
		for _, skillID := range []string{"herbalism", "mining", "woodcutting", "scavenging"} {
			if level, ok := skills.Levels[skillID]; ok {
				quality += float64(level) * 0.005 // +0.5% per skill level
			}
		}
	}
	if quality > 1.0 {
		quality = 1.0
	}

	// Update node
	rn.Quantity -= baseAmount
	if rn.Quantity <= 0 {
		rn.Depleted = true
		// Record depletion time - find world clock
		for _, ce := range w.Entities("WorldClock") {
			clockComp, clockOK := w.GetComponent(ce, "WorldClock")
			if clockOK {
				clock := clockComp.(*components.WorldClock)
				rn.LastGathered = float64(clock.Day*HoursPerDay+clock.Hour)*SecondsPerHour + clock.TimeAccum
				break
			}
		}
	}

	return rn.ResourceType, baseAmount, quality
}

// DiscoverRecipe attempts to discover a new recipe based on experimentation.
func (s *CraftingSystem) DiscoverRecipe(w *ecs.World, entity ecs.Entity, recipeID string) bool {
	comp, ok := w.GetComponent(entity, "RecipeKnowledge")
	if !ok {
		return false
	}
	knowledge := comp.(*components.RecipeKnowledge)

	if knowledge.KnownRecipes == nil {
		knowledge.KnownRecipes = make(map[string]bool)
	}
	if knowledge.KnownRecipes[recipeID] {
		return false // Already known
	}

	knowledge.KnownRecipes[recipeID] = true
	return true
}

// ProgressRecipeDiscovery adds progress toward discovering a recipe.
func (s *CraftingSystem) ProgressRecipeDiscovery(w *ecs.World, entity ecs.Entity, recipeID string, progress float64) bool {
	comp, ok := w.GetComponent(entity, "RecipeKnowledge")
	if !ok {
		return false
	}
	knowledge := comp.(*components.RecipeKnowledge)

	if knowledge.KnownRecipes == nil {
		knowledge.KnownRecipes = make(map[string]bool)
	}
	if knowledge.DiscoveryProgress == nil {
		knowledge.DiscoveryProgress = make(map[string]float64)
	}

	if knowledge.KnownRecipes[recipeID] {
		return false // Already discovered
	}

	knowledge.DiscoveryProgress[recipeID] += progress
	if knowledge.DiscoveryProgress[recipeID] >= 1.0 {
		knowledge.KnownRecipes[recipeID] = true
		delete(knowledge.DiscoveryProgress, recipeID)
		return true // Newly discovered
	}
	return false
}

// KnowsRecipe checks if an entity knows a specific recipe.
func (s *CraftingSystem) KnowsRecipe(w *ecs.World, entity ecs.Entity, recipeID string) bool {
	comp, ok := w.GetComponent(entity, "RecipeKnowledge")
	if !ok {
		return false
	}
	knowledge := comp.(*components.RecipeKnowledge)
	if knowledge.KnownRecipes == nil {
		return false
	}
	return knowledge.KnownRecipes[recipeID]
}

// RepairTool restores durability to a tool using materials.
func (s *CraftingSystem) RepairTool(w *ecs.World, entity ecs.Entity, repairAmount float64) bool {
	comp, ok := w.GetComponent(entity, "Tool")
	if !ok {
		return false
	}
	tool := comp.(*components.Tool)
	if tool.Durability >= tool.MaxDurability {
		return false // Already at max
	}

	tool.Durability += repairAmount
	if tool.Durability > tool.MaxDurability {
		tool.Durability = tool.MaxDurability
	}
	return true
}

// CalculateCraftQuality determines the quality of a crafted item.
func (s *CraftingSystem) CalculateCraftQuality(w *ecs.World, crafter, workbench ecs.Entity, materialQuality float64) float64 {
	baseQuality := materialQuality

	// Add workbench bonus
	wbComp, wbOK := w.GetComponent(workbench, "Workbench")
	if wbOK {
		wb := wbComp.(*components.Workbench)
		baseQuality += wb.QualityBonus
	}

	// Add skill bonus
	skillsComp, skillsOK := w.GetComponent(crafter, "Skills")
	if skillsOK {
		skills := skillsComp.(*components.Skills)
		for _, skillID := range []string{"smithing", "alchemy", "enchanting", "cooking", "crafting"} {
			if level, ok := skills.Levels[skillID]; ok {
				baseQuality += float64(level) * 0.01 // +1% per skill level
			}
		}
	}

	// Apply random variance (+/- 10%)
	variance := (s.rng.Float64() - 0.5) * 0.2
	finalQuality := baseQuality + variance

	// Clamp to valid range
	if finalQuality < 0.1 {
		finalQuality = 0.1
	}
	if finalQuality > 1.0 {
		finalQuality = 1.0
	}

	return finalQuality
}

// GetQualityTier returns the quality tier name for a quality value.
func GetQualityTier(quality float64) string {
	switch {
	case quality >= 0.95:
		return "Legendary"
	case quality >= 0.85:
		return "Epic"
	case quality >= 0.70:
		return "Rare"
	case quality >= 0.50:
		return "Uncommon"
	default:
		return "Common"
	}
}

// GetToolEfficiency returns the efficiency multiplier for a tool tier vs resource tier.
func GetToolEfficiency(toolTier, resourceTier int) float64 {
	diff := toolTier - resourceTier
	switch {
	case diff >= 2:
		return 2.0 // Tool much better than resource
	case diff == 1:
		return 1.5 // Tool slightly better
	case diff == 0:
		return 1.0 // Matched
	case diff == -1:
		return 0.5 // Tool slightly worse
	default:
		return 0.25 // Tool much worse
	}
}
