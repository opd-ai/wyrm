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
	gameTime, ok := s.getGameTime(w)
	if !ok {
		return
	}

	for _, e := range w.Entities("ResourceNode") {
		s.tryRespawnNode(w, e, gameTime)
	}
}

// getGameTime retrieves the current game time from WorldClock.
func (s *CraftingSystem) getGameTime(w *ecs.World) (float64, bool) {
	for _, ce := range w.Entities("WorldClock") {
		clockComp, ok := w.GetComponent(ce, "WorldClock")
		if ok {
			clock := clockComp.(*components.WorldClock)
			return float64(clock.Day*HoursPerDay+clock.Hour)*SecondsPerHour + clock.TimeAccum, true
		}
	}
	return 0, false
}

// tryRespawnNode checks and respawns a single resource node if ready.
func (s *CraftingSystem) tryRespawnNode(w *ecs.World, e ecs.Entity, gameTime float64) {
	comp, ok := w.GetComponent(e, "ResourceNode")
	if !ok {
		return
	}
	node := comp.(*components.ResourceNode)
	if !node.Depleted {
		return
	}
	if gameTime-node.LastGathered >= node.RespawnTime {
		node.Depleted = false
		node.Quantity = node.MaxQuantity
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

	s.returnPartialMaterials(w, crafter, state)
	s.resetCraftingState(state)
	return true
}

// returnPartialMaterials returns a fraction of consumed materials to inventory.
func (s *CraftingSystem) returnPartialMaterials(w *ecs.World, crafter ecs.Entity, state *components.CraftingState) {
	if state.ConsumedMaterials == nil {
		return
	}
	invComp, invOK := w.GetComponent(crafter, "Inventory")
	if !invOK {
		return
	}
	inv := invComp.(*components.Inventory)
	for mat, qty := range state.ConsumedMaterials {
		returnQty := qty / MaterialReturnDivisor
		if returnQty > 0 {
			inv.Items = append(inv.Items, mat)
			_ = returnQty // Would add returnQty times
		}
	}
}

// resetCraftingState clears the crafting state.
func (s *CraftingSystem) resetCraftingState(state *components.CraftingState) {
	state.IsCrafting = false
	state.Progress = 0
	state.CurrentRecipeID = ""
	state.ConsumedMaterials = nil
}

// gatherToolBonus contains tool effects for gathering.
type gatherToolBonus struct {
	speed   float64
	quality float64
}

// applyToolDurability reduces tool durability and returns bonuses.
func applyToolDurability(w *ecs.World, gatherer ecs.Entity) gatherToolBonus {
	result := gatherToolBonus{speed: ToolMatchedEfficiency, quality: ZeroQuality}

	toolComp, toolOK := w.GetComponent(gatherer, "Tool")
	if !toolOK {
		return result
	}
	tool := toolComp.(*components.Tool)
	if tool.Durability <= 0 {
		return result
	}

	result.speed = tool.GatherSpeed
	result.quality = tool.QualityBonus
	tool.Durability -= DefaultToolDurabilityLoss
	if tool.Durability < 0 {
		tool.Durability = 0
	}
	return result
}

// calculateGatherAmount determines how much resource to gather.
func calculateGatherAmount(available int, gatherSpeed float64) int {
	amount := BaseGatherAmount
	if gatherSpeed > ToolMatchedEfficiency {
		amount = int(math.Ceil(float64(amount) * gatherSpeed))
	}
	if amount > available {
		amount = available
	}
	return amount
}

// applyGatheringSkillBonus adds quality bonus from gathering skills.
func applyGatheringSkillBonus(w *ecs.World, gatherer ecs.Entity, baseQuality float64) float64 {
	skillsComp, skillsOK := w.GetComponent(gatherer, "Skills")
	if !skillsOK {
		return baseQuality
	}
	skills := skillsComp.(*components.Skills)
	quality := baseQuality
	for _, skillID := range []string{"herbalism", "mining", "woodcutting", "scavenging"} {
		if level, ok := skills.Levels[skillID]; ok {
			quality += float64(level) * GatheringSkillQualityBonus
		}
	}
	return clampQuality(quality)
}

// clampQuality ensures quality stays in valid range [0, 1].
func clampQuality(quality float64) float64 {
	if quality > MaximumQuality {
		return MaximumQuality
	}
	if quality < ZeroQuality {
		return ZeroQuality
	}
	return quality
}

// recordDepletionTime sets the depletion timestamp from world clock.
func recordDepletionTime(w *ecs.World, node *components.ResourceNode) {
	for _, ce := range w.Entities("WorldClock") {
		clockComp, clockOK := w.GetComponent(ce, "WorldClock")
		if clockOK {
			clock := clockComp.(*components.WorldClock)
			node.LastGathered = float64(clock.Day*HoursPerDay+clock.Hour)*SecondsPerHour + clock.TimeAccum
			return
		}
	}
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

	// Apply tool effects
	toolBonus := applyToolDurability(w, gatherer)

	// Calculate gathered amount
	amount := calculateGatherAmount(rn.Quantity, toolBonus.speed)

	// Calculate quality with tool and skill bonuses
	quality := clampQuality(rn.Quality + toolBonus.quality)
	quality = applyGatheringSkillBonus(w, gatherer, quality)

	// Update node state
	rn.Quantity -= amount
	if rn.Quantity <= 0 {
		rn.Depleted = true
		recordDepletionTime(w, rn)
	}

	return rn.ResourceType, amount, quality
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
	baseQuality += s.workbenchQualityBonus(w, workbench)
	baseQuality += s.crafterSkillBonus(w, crafter)
	return s.applyQualityVariance(baseQuality)
}

// workbenchQualityBonus returns quality bonus from workbench.
func (s *CraftingSystem) workbenchQualityBonus(w *ecs.World, workbench ecs.Entity) float64 {
	wbComp, ok := w.GetComponent(workbench, "Workbench")
	if !ok {
		return 0
	}
	wb := wbComp.(*components.Workbench)
	return wb.QualityBonus
}

// crafterSkillBonus returns quality bonus from crafter's crafting skills.
func (s *CraftingSystem) crafterSkillBonus(w *ecs.World, crafter ecs.Entity) float64 {
	skillsComp, ok := w.GetComponent(crafter, "Skills")
	if !ok {
		return 0
	}
	skills := skillsComp.(*components.Skills)
	var bonus float64
	for _, skillID := range []string{"smithing", "alchemy", "enchanting", "cooking", "crafting"} {
		if level, ok := skills.Levels[skillID]; ok {
			bonus += float64(level) * CraftingSkillQualityBonus
		}
	}
	return bonus
}

// applyQualityVariance adds random variance and clamps result.
func (s *CraftingSystem) applyQualityVariance(baseQuality float64) float64 {
	variance := (s.rng.Float64() - QualityVarianceCenter) * QualityVarianceFactor
	finalQuality := baseQuality + variance
	if finalQuality < MinimumQuality {
		finalQuality = MinimumQuality
	}
	if finalQuality > MaximumQuality {
		finalQuality = MaximumQuality
	}
	return finalQuality
}

// GetQualityTier returns the quality tier name for a quality value.
func GetQualityTier(quality float64) string {
	switch {
	case quality >= LegendaryQualityThreshold:
		return "Legendary"
	case quality >= EpicQualityThreshold:
		return "Epic"
	case quality >= RareQualityThreshold:
		return "Rare"
	case quality >= UncommonQualityThreshold:
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
		return ToolMuchBetterEfficiency // Tool much better than resource
	case diff == 1:
		return ToolSlightlyBetterEfficiency // Tool slightly better
	case diff == 0:
		return ToolMatchedEfficiency // Matched
	case diff == -1:
		return ToolSlightlyWorseEfficiency // Tool slightly worse
	default:
		return ToolMuchWorseEfficiency // Tool much worse
	}
}

// Minigame constants.
const (
	// MinigameTimingWindow is the window in seconds to hit a timing challenge.
	MinigameTimingWindow = 0.5
	// MinigamePrecisionBonus is quality bonus for perfect timing.
	MinigamePrecisionBonus = 0.15
	// MinigameFailurePenalty is quality penalty for missing timing.
	MinigameFailurePenalty = 0.1
)

// MinigameState tracks the state of a crafting minigame.
type MinigameState struct {
	// Type is the minigame type ("timing", "sequence", "precision").
	Type string
	// TargetTime is when the player should act (for timing games).
	TargetTime float64
	// CurrentTime tracks elapsed time in the minigame.
	CurrentTime float64
	// Sequence is the required input sequence (for sequence games).
	Sequence []string
	// CurrentIndex is progress through a sequence.
	CurrentIndex int
	// Score tracks accumulated minigame performance (0.0-1.0).
	Score float64
	// Attempts is the number of minigame steps completed.
	Attempts int
	// MaxAttempts is the total minigame steps.
	MaxAttempts int
}

// StartCraftingMinigame begins a minigame for a crafting operation.
func (s *CraftingSystem) StartCraftingMinigame(recipeType string) *MinigameState {
	// Determine minigame type based on recipe
	gameType := s.getMinigameType(recipeType)

	state := &MinigameState{
		Type:        gameType,
		Score:       ZeroQuality,
		Attempts:    0,
		MaxAttempts: MinigameDefaultMaxAttempts,
	}

	switch gameType {
	case "timing":
		// Set up timing challenge
		state.TargetTime = TimingMinTargetDelay + s.rng.Float64()*TimingMaxTargetVariance // Random between 1-3 seconds
		state.CurrentTime = 0
	case "sequence":
		// Generate random sequence
		inputs := []string{"up", "down", "left", "right"}
		state.Sequence = make([]string, MinigameSequenceLength)
		for i := range state.Sequence {
			state.Sequence[i] = inputs[s.rng.Intn(len(inputs))]
		}
		state.CurrentIndex = 0
	case "precision":
		// Precision is a series of timing windows
		state.TargetTime = PrecisionMinTargetDelay
		state.CurrentTime = 0
	}

	return state
}

// getMinigameType returns appropriate minigame type for recipe.
func (s *CraftingSystem) getMinigameType(recipeType string) string {
	switch recipeType {
	case "smithing", "weapon", "armor":
		return "timing" // Hammering rhythm
	case "alchemy", "potion", "chemistry":
		return "sequence" // Ingredient order
	case "enchanting", "magic":
		return "precision" // Rune tracing
	default:
		return "timing"
	}
}

// ProcessMinigameInput handles player input during a minigame.
func (s *CraftingSystem) ProcessMinigameInput(state *MinigameState, input string, currentTime float64) (complete, success bool) {
	if state == nil {
		return true, false
	}

	switch state.Type {
	case "timing":
		s.processTimingInput(state, currentTime)
	case "sequence":
		s.processSequenceInput(state, input)
	case "precision":
		s.processPrecisionInput(state, currentTime)
	}

	complete = s.isMinigameComplete(state)
	success = state.Score >= MinigamePassThreshold
	return complete, success
}

// processTimingInput handles a timing-based minigame input.
func (s *CraftingSystem) processTimingInput(state *MinigameState, currentTime float64) {
	state.CurrentTime = currentTime
	timeDiff := math.Abs(state.CurrentTime - state.TargetTime)
	if timeDiff <= MinigameTimingWindow {
		state.Score += MaximumQuality / float64(state.MaxAttempts)
	} else if timeDiff <= MinigameTimingWindow*MinigameTimingWindowDouble {
		state.Score += MinigameGoodTimingMultiplier / float64(state.MaxAttempts)
	}
	state.Attempts++
	state.TargetTime = currentTime + TimingMinTargetDelay + s.rng.Float64()*TimingMaxTargetVariance
}

// processSequenceInput handles a sequence-based minigame input.
func (s *CraftingSystem) processSequenceInput(state *MinigameState, input string) {
	if state.CurrentIndex < len(state.Sequence) {
		if input == state.Sequence[state.CurrentIndex] {
			state.Score += MaximumQuality / float64(len(state.Sequence))
			state.CurrentIndex++
		} else {
			state.Score -= MinigameSequenceWrongPenalty
			if state.Score < 0 {
				state.Score = 0
			}
			state.CurrentIndex = 0
		}
	}
	state.Attempts++
}

// processPrecisionInput handles a precision-based minigame input.
func (s *CraftingSystem) processPrecisionInput(state *MinigameState, currentTime float64) {
	state.CurrentTime = currentTime
	timeDiff := math.Abs(state.CurrentTime - state.TargetTime)
	if timeDiff <= MinigameTimingWindow*MinigamePrecisionWindowHalf {
		state.Score += MaximumQuality / float64(state.MaxAttempts)
	} else if timeDiff <= MinigameTimingWindow {
		state.Score += MinigamePrecisionGoodScore / float64(state.MaxAttempts)
	}
	state.Attempts++
	state.TargetTime = currentTime + PrecisionMinTargetDelay + s.rng.Float64()*PrecisionMaxTargetVariance
}

// isMinigameComplete determines if the minigame has ended.
func (s *CraftingSystem) isMinigameComplete(state *MinigameState) bool {
	if state.Type == "sequence" {
		return state.CurrentIndex >= len(state.Sequence) || state.Attempts >= state.MaxAttempts*MinigameTimingWindowDouble
	}
	return state.Attempts >= state.MaxAttempts
}

// ApplyMinigameBonus modifies craft quality based on minigame performance.
func (s *CraftingSystem) ApplyMinigameBonus(baseQuality, minigameScore float64) float64 {
	if minigameScore >= MinigameExcellentThreshold {
		return baseQuality + MinigamePrecisionBonus
	} else if minigameScore >= MinigameGoodThreshold {
		return baseQuality + MinigamePrecisionBonus*MinigameGoodTimingMultiplier
	} else if minigameScore < MinigamePoorThreshold {
		return baseQuality - MinigameFailurePenalty
	}
	return baseQuality
}

// Enchanting constants.
const (
	// BaseEnchantSuccessRate is the base success chance for enchanting.
	BaseEnchantSuccessRate = 0.7
	// EnchantSkillBonus is success bonus per enchanting skill level.
	EnchantSkillBonus = 0.02
	// MaxEnchantments is the maximum enchantments per item.
	MaxEnchantments = 3
)

// EnchantmentType defines an enchantment effect.
type EnchantmentType struct {
	ID           string
	Name         string
	Effect       string // "damage_fire", "defense_frost", "mana_regen", etc.
	MinMagnitude float64
	MaxMagnitude float64
	ManaCost     float64
}

// GenreEnchantments maps genres to available enchantment types.
var GenreEnchantments = map[string][]EnchantmentType{
	"fantasy": {
		{ID: "fire", Name: "Flame", Effect: "damage_fire", MinMagnitude: 5, MaxMagnitude: 25, ManaCost: 20},
		{ID: "frost", Name: "Frost", Effect: "damage_frost", MinMagnitude: 5, MaxMagnitude: 25, ManaCost: 20},
		{ID: "shock", Name: "Lightning", Effect: "damage_shock", MinMagnitude: 5, MaxMagnitude: 25, ManaCost: 25},
		{ID: "fortify", Name: "Fortification", Effect: "defense_all", MinMagnitude: 5, MaxMagnitude: 15, ManaCost: 30},
		{ID: "mana", Name: "Magicka", Effect: "mana_regen", MinMagnitude: 2, MaxMagnitude: 10, ManaCost: 35},
	},
	"sci-fi": {
		{ID: "plasma", Name: "Plasma", Effect: "damage_energy", MinMagnitude: 8, MaxMagnitude: 30, ManaCost: 25},
		{ID: "cryo", Name: "Cryo", Effect: "damage_frost", MinMagnitude: 5, MaxMagnitude: 20, ManaCost: 20},
		{ID: "shield", Name: "Shield Boost", Effect: "defense_energy", MinMagnitude: 10, MaxMagnitude: 25, ManaCost: 30},
		{ID: "battery", Name: "Power Cell", Effect: "energy_regen", MinMagnitude: 3, MaxMagnitude: 12, ManaCost: 35},
	},
	"horror": {
		{ID: "curse", Name: "Curse", Effect: "damage_dark", MinMagnitude: 10, MaxMagnitude: 35, ManaCost: 30},
		{ID: "drain", Name: "Life Drain", Effect: "lifesteal", MinMagnitude: 3, MaxMagnitude: 10, ManaCost: 40},
		{ID: "fear", Name: "Terror", Effect: "fear_aura", MinMagnitude: 5, MaxMagnitude: 15, ManaCost: 25},
	},
	"cyberpunk": {
		{ID: "emp", Name: "EMP", Effect: "damage_emp", MinMagnitude: 15, MaxMagnitude: 40, ManaCost: 30},
		{ID: "hack", Name: "Breach", Effect: "hack_bonus", MinMagnitude: 5, MaxMagnitude: 20, ManaCost: 25},
		{ID: "armor", Name: "Nano-Armor", Effect: "defense_physical", MinMagnitude: 8, MaxMagnitude: 20, ManaCost: 35},
	},
	"post-apocalyptic": {
		{ID: "rad", Name: "Radiation", Effect: "damage_radiation", MinMagnitude: 8, MaxMagnitude: 30, ManaCost: 20},
		{ID: "toxic", Name: "Toxin", Effect: "damage_poison", MinMagnitude: 5, MaxMagnitude: 20, ManaCost: 15},
		{ID: "salvage", Name: "Scavenger", Effect: "loot_bonus", MinMagnitude: 10, MaxMagnitude: 25, ManaCost: 25},
	},
}

// Enchantment represents an applied enchantment on an item.
type Enchantment struct {
	TypeID    string
	Name      string
	Effect    string
	Magnitude float64
	Charges   int // -1 = permanent, 0+ = uses remaining
}

// EnchantItem applies an enchantment to an item.
func (s *CraftingSystem) EnchantItem(w *ecs.World, crafter, item ecs.Entity, enchantType EnchantmentType) (*Enchantment, bool) {
	mana, ok := s.getCrafterMana(w, crafter)
	if !ok || mana.Current < enchantType.ManaCost {
		return nil, false
	}

	enchantLevel := s.getEnchantingLevel(w, crafter)
	successRate := s.calculateEnchantSuccessRate(enchantLevel)

	mana.Current -= enchantType.ManaCost

	if s.rng.Float64() > successRate {
		return nil, false
	}

	magnitude := s.calculateEnchantMagnitude(enchantType, enchantLevel)

	return &Enchantment{
		TypeID:    enchantType.ID,
		Name:      enchantType.Name,
		Effect:    enchantType.Effect,
		Magnitude: magnitude,
		Charges:   -1,
	}, true
}

// getCrafterMana retrieves the Mana component from a crafter entity.
func (s *CraftingSystem) getCrafterMana(w *ecs.World, crafter ecs.Entity) (*components.Mana, bool) {
	manaComp, ok := w.GetComponent(crafter, "Mana")
	if !ok {
		return nil, false
	}
	return manaComp.(*components.Mana), true
}

// getEnchantingLevel returns the crafter's enchanting skill level.
func (s *CraftingSystem) getEnchantingLevel(w *ecs.World, crafter ecs.Entity) int {
	skillsComp, ok := w.GetComponent(crafter, "Skills")
	if !ok {
		return 0
	}
	skills := skillsComp.(*components.Skills)
	if level, found := skills.Levels["enchanting"]; found {
		return level
	}
	return 0
}

// calculateEnchantSuccessRate computes success rate based on skill level.
func (s *CraftingSystem) calculateEnchantSuccessRate(enchantLevel int) float64 {
	rate := BaseEnchantSuccessRate + float64(enchantLevel)*EnchantSkillBonus
	if rate > 0.95 {
		return 0.95
	}
	return rate
}

// calculateEnchantMagnitude computes the enchantment magnitude.
func (s *CraftingSystem) calculateEnchantMagnitude(enchantType EnchantmentType, enchantLevel int) float64 {
	magnitudeRange := enchantType.MaxMagnitude - enchantType.MinMagnitude
	magnitude := enchantType.MinMagnitude + s.rng.Float64()*magnitudeRange
	magnitude *= 1.0 + float64(enchantLevel)*0.01
	if magnitude > enchantType.MaxMagnitude*1.5 {
		magnitude = enchantType.MaxMagnitude * 1.5
	}
	return magnitude
}

// GetAvailableEnchantments returns enchantments available for a genre.
func (s *CraftingSystem) GetAvailableEnchantments(genre string) []EnchantmentType {
	if enchants, ok := GenreEnchantments[genre]; ok {
		return enchants
	}
	return GenreEnchantments["fantasy"]
}

// DisassemblyResult contains materials recovered from disassembly.
type DisassemblyResult struct {
	Materials     map[string]int
	RareMaterials map[string]int
	Success       bool
	Message       string
}

// Disassembly constants.
const (
	// BaseDisassemblyRate is the base material recovery rate.
	BaseDisassemblyRate = 0.5
	// DisassemblySkillBonus is recovery bonus per relevant skill level.
	DisassemblySkillBonus = 0.02
	// RareMaterialChance is chance to recover rare materials.
	RareMaterialChance = 0.1
)

// disassemblySkills are the skills that affect disassembly recovery rate.
var disassemblySkills = []string{"crafting", "smithing", "engineering", "scavenging"}

// DisassembleItem breaks down an item into materials.
func (s *CraftingSystem) DisassembleItem(w *ecs.World, crafter ecs.Entity, itemQuality float64, itemType string) DisassemblyResult {
	result := DisassemblyResult{
		Materials:     make(map[string]int),
		RareMaterials: make(map[string]int),
		Success:       true,
	}

	recoveryRate := s.calculateRecoveryRate(w, crafter)
	s.recoverBaseMaterials(&result, itemType, recoveryRate)
	s.tryRecoverRareMaterial(&result, itemType, itemQuality)

	result.Message = "Item successfully disassembled"
	return result
}

// calculateRecoveryRate computes material recovery rate based on crafter skills.
func (s *CraftingSystem) calculateRecoveryRate(w *ecs.World, crafter ecs.Entity) float64 {
	rate := BaseDisassemblyRate
	skillsComp, ok := w.GetComponent(crafter, "Skills")
	if ok {
		skills := skillsComp.(*components.Skills)
		for _, skillID := range disassemblySkills {
			if level, found := skills.Levels[skillID]; found {
				rate += float64(level) * DisassemblySkillBonus
			}
		}
	}
	if rate > 0.9 {
		return 0.9
	}
	return rate
}

// recoverBaseMaterials adds recovered materials to the result.
func (s *CraftingSystem) recoverBaseMaterials(result *DisassemblyResult, itemType string, recoveryRate float64) {
	baseMaterials := s.getItemBaseMaterials(itemType)
	for mat, baseQty := range baseMaterials {
		recovered := int(float64(baseQty) * recoveryRate * (0.8 + s.rng.Float64()*0.4))
		if recovered > 0 {
			result.Materials[mat] = recovered
		}
	}
}

// tryRecoverRareMaterial attempts to add a rare material based on item quality.
func (s *CraftingSystem) tryRecoverRareMaterial(result *DisassemblyResult, itemType string, itemQuality float64) {
	rareChance := RareMaterialChance * (1.0 + itemQuality)
	if s.rng.Float64() < rareChance {
		rareMat := s.getRareMaterialForType(itemType)
		if rareMat != "" {
			result.RareMaterials[rareMat] = 1
		}
	}
}

// getItemBaseMaterials returns materials that make up an item type.
func (s *CraftingSystem) getItemBaseMaterials(itemType string) map[string]int {
	switch itemType {
	case "sword", "axe", "mace":
		return map[string]int{"metal_ingot": 3, "leather": 1, "wood": 1}
	case "bow", "staff":
		return map[string]int{"wood": 4, "leather": 1, "string": 2}
	case "armor", "helmet", "shield":
		return map[string]int{"metal_ingot": 5, "leather": 2, "padding": 1}
	case "robe", "clothing":
		return map[string]int{"cloth": 4, "thread": 2, "dye": 1}
	case "potion", "elixir":
		return map[string]int{"glass": 1, "herb": 2}
	case "ring", "amulet":
		return map[string]int{"metal_ingot": 1, "gem": 1}
	default:
		return map[string]int{"scrap": 2}
	}
}

// getRareMaterialForType returns a rare material that can drop from an item type.
func (s *CraftingSystem) getRareMaterialForType(itemType string) string {
	switch itemType {
	case "sword", "axe", "mace", "armor", "helmet", "shield":
		materials := []string{"rare_metal", "enchanted_fragment", "soul_gem_shard"}
		return materials[s.rng.Intn(len(materials))]
	case "bow", "staff":
		materials := []string{"heartwood", "enchanted_string", "focus_crystal"}
		return materials[s.rng.Intn(len(materials))]
	case "robe", "clothing":
		materials := []string{"magic_thread", "rare_dye", "enchanted_cloth"}
		return materials[s.rng.Intn(len(materials))]
	case "potion", "elixir":
		materials := []string{"rare_essence", "catalyst", "purified_water"}
		return materials[s.rng.Intn(len(materials))]
	case "ring", "amulet":
		materials := []string{"flawless_gem", "rare_metal", "enchantment_core"}
		return materials[s.rng.Intn(len(materials))]
	default:
		return "mystery_component"
	}
}
