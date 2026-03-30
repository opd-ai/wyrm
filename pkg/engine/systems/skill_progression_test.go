package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestSkillRegistry(t *testing.T) {
	reg := NewSkillRegistry()

	if reg == nil {
		t.Fatal("Registry should not be nil")
	}

	// Should have 30+ skills
	if reg.SkillCount() < 30 {
		t.Errorf("Expected at least 30 skills, got %d", reg.SkillCount())
	}

	// Should have 6 schools
	if reg.SchoolCount() != 6 {
		t.Errorf("Expected 6 schools, got %d", reg.SchoolCount())
	}
}

func TestSkillRegistryAllSchoolsHaveSkills(t *testing.T) {
	reg := NewSkillRegistry()

	schools := []string{"combat", "stealth", "magic", "crafting", "social", "survival"}

	for _, schoolID := range schools {
		skills := reg.GetSkillsForSchool(schoolID)
		if len(skills) < 5 {
			t.Errorf("School %s should have at least 5 skills, got %d", schoolID, len(skills))
		}

		// Verify each skill exists
		for _, skillID := range skills {
			if _, ok := reg.Skills[skillID]; !ok {
				t.Errorf("Skill %s in school %s not found in registry", skillID, schoolID)
			}
		}
	}
}

func TestSkillRegistryGenreNames(t *testing.T) {
	reg := NewSkillRegistry()

	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		// Check that melee skill has a genre-specific name
		name := reg.GetSkillName("melee", genre)
		if name == "" || name == "melee" {
			t.Errorf("Genre %s should have custom name for melee, got %s", genre, name)
		}
	}

	// Unknown genre should fall back to default
	name := reg.GetSkillName("melee", "unknown_genre")
	if name == "" {
		t.Error("Unknown genre should return skill's default name")
	}
}

func TestSkillRegistryAllSkillsHaveDefinitions(t *testing.T) {
	reg := NewSkillRegistry()

	for skillID, skill := range reg.Skills {
		if skill.Name == "" {
			t.Errorf("Skill %s has no name", skillID)
		}
		if skill.School == "" {
			t.Errorf("Skill %s has no school", skillID)
		}
		if skill.Description == "" {
			t.Errorf("Skill %s has no description", skillID)
		}
		if skill.MaxLevel != 100 {
			t.Errorf("Skill %s should have max level 100, got %d", skillID, skill.MaxLevel)
		}
	}
}

// NPC Training System Tests

func TestNPCTrainingSystem(t *testing.T) {
	reg := NewSkillRegistry()
	progSys := NewSkillProgressionSystem(100, 100)
	trainingSys := NewNPCTrainingSystem(reg, progSys)

	if trainingSys == nil {
		t.Fatal("Training system should not be nil")
	}

	if trainingSys.TrainerCount() != 0 {
		t.Error("Initial trainer count should be 0")
	}
}

func TestRegisterTrainer(t *testing.T) {
	reg := NewSkillRegistry()
	progSys := NewSkillProgressionSystem(100, 100)
	trainingSys := NewNPCTrainingSystem(reg, progSys)

	trainer := &NPCTrainer{
		NPCID:         1,
		Name:          "Master Smith",
		TrainedSkills: []string{"smithing", "melee"},
		MaxTrainLevel: 50,
		CostPerLevel:  10,
		TrainXPBonus:  0.5,
	}

	trainingSys.RegisterTrainer(trainer)

	if trainingSys.TrainerCount() != 1 {
		t.Error("Should have 1 trainer after registration")
	}

	if !trainingSys.IsTrainer(1) {
		t.Error("NPC 1 should be a trainer")
	}

	if trainingSys.IsTrainer(2) {
		t.Error("NPC 2 should not be a trainer")
	}

	retrieved := trainingSys.GetTrainer(1)
	if retrieved == nil {
		t.Fatal("Should retrieve trainer")
	}
	if retrieved.Name != "Master Smith" {
		t.Errorf("Trainer name = %s, want Master Smith", retrieved.Name)
	}
}

func TestGetTrainableSkills(t *testing.T) {
	reg := NewSkillRegistry()
	progSys := NewSkillProgressionSystem(100, 100)
	trainingSys := NewNPCTrainingSystem(reg, progSys)

	trainer := &NPCTrainer{
		NPCID:         1,
		Name:          "Combat Trainer",
		TrainedSkills: []string{"melee", "ranged", "blocking"},
		MaxTrainLevel: 50,
		CostPerLevel:  10,
		TrainXPBonus:  1.0,
	}
	trainingSys.RegisterTrainer(trainer)

	world := ecs.NewWorld()
	player := world.CreateEntity()
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 30, "ranged": 0, "blocking": 50},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	trainable := trainingSys.GetTrainableSkills(world, 1, player)

	// melee (30) and ranged (0) can be trained, blocking (50) is at max
	if len(trainable) != 2 {
		t.Errorf("Expected 2 trainable skills, got %d", len(trainable))
	}

	// Verify blocking is not in list
	for _, skill := range trainable {
		if skill == "blocking" {
			t.Error("Blocking should not be trainable (at max level)")
		}
	}
}

func TestCalculateTrainingCost(t *testing.T) {
	reg := NewSkillRegistry()
	progSys := NewSkillProgressionSystem(100, 100)
	trainingSys := NewNPCTrainingSystem(reg, progSys)

	trainer := &NPCTrainer{
		NPCID:         1,
		Name:          "Trainer",
		TrainedSkills: []string{"melee"},
		MaxTrainLevel: 100,
		CostPerLevel:  10,
		TrainXPBonus:  1.0,
	}
	trainingSys.RegisterTrainer(trainer)

	// Cost for training from 0 to 1
	cost1 := trainingSys.CalculateTrainingCost(1, 0, 1)
	if cost1 <= 0 {
		t.Error("Cost should be positive")
	}

	// Cost for training from 0 to 10 should be more than from 0 to 1
	cost10 := trainingSys.CalculateTrainingCost(1, 0, 10)
	if cost10 <= cost1 {
		t.Error("Cost for more levels should be higher")
	}

	// Cost for higher levels should be more expensive
	costHigh := trainingSys.CalculateTrainingCost(1, 50, 51)
	if costHigh <= cost1 {
		t.Error("Training at higher levels should cost more")
	}
}

func TestTrainSkillSuccess(t *testing.T) {
	reg := NewSkillRegistry()
	progSys := NewSkillProgressionSystem(100, 100)
	trainingSys := NewNPCTrainingSystem(reg, progSys)

	trainer := &NPCTrainer{
		NPCID:         1,
		Name:          "Trainer",
		TrainedSkills: []string{"melee"},
		MaxTrainLevel: 50,
		CostPerLevel:  10,
		TrainXPBonus:  1.0,
	}
	trainingSys.RegisterTrainer(trainer)

	world := ecs.NewWorld()
	player := world.CreateEntity()
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 0},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	playerGold := 1000

	result := trainingSys.TrainSkill(world, 1, player, "melee", 5, &playerGold)

	if !result.Success {
		t.Errorf("Training should succeed, got error: %s", result.ErrorMessage)
	}
	if result.LevelsGained != 5 {
		t.Errorf("Should gain 5 levels, got %d", result.LevelsGained)
	}
	if result.GoldSpent <= 0 {
		t.Error("Should spend gold")
	}
	if playerGold >= 1000 {
		t.Error("Player gold should be reduced")
	}
	if skills.Levels["melee"] != 5 {
		t.Errorf("Skill level should be 5, got %d", skills.Levels["melee"])
	}
}

func TestTrainSkillInsufficientGold(t *testing.T) {
	reg := NewSkillRegistry()
	progSys := NewSkillProgressionSystem(100, 100)
	trainingSys := NewNPCTrainingSystem(reg, progSys)

	trainer := &NPCTrainer{
		NPCID:         1,
		Name:          "Expensive Trainer",
		TrainedSkills: []string{"melee"},
		MaxTrainLevel: 50,
		CostPerLevel:  1000, // Very expensive
		TrainXPBonus:  1.0,
	}
	trainingSys.RegisterTrainer(trainer)

	world := ecs.NewWorld()
	player := world.CreateEntity()
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 0},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	playerGold := 10 // Not enough

	result := trainingSys.TrainSkill(world, 1, player, "melee", 5, &playerGold)

	if result.Success {
		t.Error("Training should fail due to insufficient gold")
	}
	if result.ErrorMessage == "" {
		t.Error("Should have error message")
	}
}

func TestTrainSkillAtMaxLevel(t *testing.T) {
	reg := NewSkillRegistry()
	progSys := NewSkillProgressionSystem(100, 100)
	trainingSys := NewNPCTrainingSystem(reg, progSys)

	trainer := &NPCTrainer{
		NPCID:         1,
		Name:          "Trainer",
		TrainedSkills: []string{"melee"},
		MaxTrainLevel: 50,
		CostPerLevel:  10,
		TrainXPBonus:  1.0,
	}
	trainingSys.RegisterTrainer(trainer)

	world := ecs.NewWorld()
	player := world.CreateEntity()
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 50}, // Already at max trainable
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	playerGold := 1000

	result := trainingSys.TrainSkill(world, 1, player, "melee", 5, &playerGold)

	if result.Success {
		t.Error("Training should fail when already at max trainable level")
	}
}

func TestTrainSkillNotTaught(t *testing.T) {
	reg := NewSkillRegistry()
	progSys := NewSkillProgressionSystem(100, 100)
	trainingSys := NewNPCTrainingSystem(reg, progSys)

	trainer := &NPCTrainer{
		NPCID:         1,
		Name:          "Melee Trainer",
		TrainedSkills: []string{"melee"}, // Only teaches melee
		MaxTrainLevel: 50,
		CostPerLevel:  10,
		TrainXPBonus:  1.0,
	}
	trainingSys.RegisterTrainer(trainer)

	world := ecs.NewWorld()
	player := world.CreateEntity()
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 0, "ranged": 0},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	playerGold := 1000

	result := trainingSys.TrainSkill(world, 1, player, "ranged", 5, &playerGold)

	if result.Success {
		t.Error("Training should fail - trainer doesn't teach ranged")
	}
}

func TestSkillProgressionWithTraining(t *testing.T) {
	progSys := NewSkillProgressionSystem(100, 100)

	world := ecs.NewWorld()
	player := world.CreateEntity()
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 5},
		Experience: map[string]float64{"melee": 500}, // Enough for level up
	}
	world.AddComponent(player, skills)

	// Run progression update
	progSys.Update(world, 0.016)

	// Should have leveled up
	if skills.Levels["melee"] <= 5 {
		t.Error("Should have gained levels from XP")
	}
}

func BenchmarkSkillRegistry(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewSkillRegistry()
	}
}

func BenchmarkGetSkillName(b *testing.B) {
	reg := NewSkillRegistry()
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		genre := genres[i%len(genres)]
		reg.GetSkillName("melee", genre)
	}
}

// ============================================================================
// Action Unlock System Tests
// ============================================================================

func TestNewActionUnlockSystem(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	if sys == nil {
		t.Fatal("System should not be nil")
	}

	// Should have default unlocks
	if sys.UnlockCount() == 0 {
		t.Error("Should have default unlocks registered")
	}
}

func TestActionUnlockSystemRegister(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	initialCount := sys.UnlockCount()

	customUnlock := &ActionUnlock{
		ID:            "custom_ability",
		Name:          "Custom Ability",
		Description:   "A test ability",
		SkillID:       "melee",
		RequiredLevel: 25,
		ActionType:    "ability",
		EffectType:    "damage",
		BasePower:     1.5,
	}

	sys.RegisterUnlock(customUnlock)

	if sys.UnlockCount() != initialCount+1 {
		t.Error("Should have one more unlock after registration")
	}

	retrieved := sys.GetUnlock("custom_ability")
	if retrieved == nil {
		t.Error("Should retrieve registered unlock")
	}
	if retrieved.Name != "Custom Ability" {
		t.Error("Retrieved unlock should have correct name")
	}
}

func TestActionUnlockSystemBySkill(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	// Melee should have unlocks
	meleeUnlocks := sys.GetUnlocksForSkill("melee")
	if len(meleeUnlocks) == 0 {
		t.Error("Melee skill should have unlocks")
	}

	// Verify all are for melee
	for _, unlock := range meleeUnlocks {
		if unlock.SkillID != "melee" {
			t.Errorf("Unlock %s should be for melee, got %s", unlock.ID, unlock.SkillID)
		}
	}
}

func TestActionUnlockSystemPlayerUnlocks(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Player with high melee skill
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 50, "sneak": 30},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	// Process unlocks
	sys.Update(world, 0.016)

	// Should have unlocked power_attack (requires level 15)
	if !sys.IsUnlocked(player, "power_attack") {
		t.Error("Should have unlocked power_attack at melee 50")
	}

	// Should have unlocked cleave (requires level 30, prereq power_attack)
	if !sys.IsUnlocked(player, "cleave") {
		t.Error("Should have unlocked cleave at melee 50")
	}

	// Should NOT have execute (requires level 75)
	if sys.IsUnlocked(player, "execute") {
		t.Error("Should not have execute at melee 50")
	}
}

func TestActionUnlockSystemPrerequisites(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Player with exactly level 30 melee - meets cleave level req
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 30},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	// First update - should unlock power_attack first
	sys.Update(world, 0.016)

	// Cleave requires power_attack as prereq
	if !sys.IsUnlocked(player, "power_attack") {
		t.Error("power_attack should be unlocked first")
	}
	if !sys.IsUnlocked(player, "cleave") {
		t.Error("cleave should be unlocked when prereq is met")
	}
}

func TestActionUnlockSystemGetUnlockedActions(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	skills := &components.Skills{
		Levels:     map[string]int{"melee": 25, "ranged": 20},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	unlocked := sys.GetUnlockedActions(player)
	if len(unlocked) == 0 {
		t.Error("Should have some unlocked actions")
	}

	// Verify each returned unlock is actually unlocked
	for _, unlock := range unlocked {
		if !sys.IsUnlocked(player, unlock.ID) {
			t.Errorf("Returned unlock %s should be marked as unlocked", unlock.ID)
		}
	}
}

func TestActionUnlockSystemAvailableUnlocks(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Low level player
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 5},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	available := sys.GetAvailableUnlocks(world, player)

	// Should have power_attack available (not yet unlocked but has the skill)
	found := false
	for _, unlock := range available {
		if unlock.ID == "power_attack" {
			found = true
			break
		}
	}
	if !found {
		t.Error("power_attack should be available (not yet unlocked)")
	}
}

func TestActionUnlockSystemNextUnlock(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	skills := &components.Skills{
		Levels:     map[string]int{"melee": 10},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	next := sys.GetNextUnlockForSkill(world, player, "melee")
	if next == nil {
		t.Fatal("Should have a next unlock")
	}

	// Should be power_attack at level 15
	if next.ID != "power_attack" {
		t.Errorf("Next unlock should be power_attack, got %s", next.ID)
	}
	if next.RequiredLevel != 15 {
		t.Errorf("power_attack requires level 15, got %d", next.RequiredLevel)
	}
}

func TestActionUnlockSystemByType(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	types := []string{"ability", "passive", "recipe", "interaction", "perk"}

	for _, actionType := range types {
		unlocks := sys.GetUnlocksByType(actionType)
		for _, unlock := range unlocks {
			if unlock.ActionType != actionType {
				t.Errorf("Unlock %s type should be %s, got %s", unlock.ID, actionType, unlock.ActionType)
			}
		}
	}

	// Abilities should be the most common
	abilities := sys.GetUnlocksByType("ability")
	if len(abilities) < 10 {
		t.Errorf("Should have many ability unlocks, got %d", len(abilities))
	}
}

func TestUnlockTiers(t *testing.T) {
	// Verify tier constants are in ascending order
	tiers := []int{
		UnlockTierNovice, UnlockTierApprentice, UnlockTierJourneyman,
		UnlockTierExpert, UnlockTierMaster, UnlockTierGrandmaster,
	}

	for i := 1; i < len(tiers); i++ {
		if tiers[i] <= tiers[i-1] {
			t.Errorf("Tier %d (%d) should be greater than tier %d (%d)",
				i, tiers[i], i-1, tiers[i-1])
		}
	}

	// Grandmaster should be 100
	if UnlockTierGrandmaster != 100 {
		t.Errorf("Grandmaster tier should be 100, got %d", UnlockTierGrandmaster)
	}
}

func BenchmarkActionUnlockSystem(b *testing.B) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewActionUnlockSystem(reg, prog)
	}
}

func BenchmarkActionUnlockUpdate(b *testing.B) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewActionUnlockSystem(reg, prog)

	world := ecs.NewWorld()
	for i := 0; i < 100; i++ {
		player := world.CreateEntity()
		skills := &components.Skills{
			Levels:     map[string]int{"melee": 50, "ranged": 40, "sneak": 30},
			Experience: make(map[string]float64),
		}
		world.AddComponent(player, skills)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world, 0.016)
	}
}

// ============================================================================
// Skill Book System Tests
// ============================================================================

func TestNewSkillBookSystem(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	if sys == nil {
		t.Fatal("System should not be nil")
	}

	if sys.BookCount() == 0 {
		t.Error("Should have default books")
	}
}

func TestSkillBookSystemRegister(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	initialCount := sys.BookCount()

	customBook := &SkillBook{
		ID:            "custom_tome",
		Name:          "Custom Tome",
		Description:   "A test book",
		SkillID:       "melee",
		XPGrant:       500,
		RequiredLevel: 10,
		MaxLevel:      50,
		Rarity:        50,
		IsOneTimeUse:  true,
		ReadTime:      20,
	}

	sys.RegisterBook(customBook)

	if sys.BookCount() != initialCount+1 {
		t.Error("Should have one more book")
	}

	retrieved := sys.GetBook("custom_tome")
	if retrieved == nil {
		t.Error("Should retrieve registered book")
	}
	if retrieved.Name != "Custom Tome" {
		t.Error("Retrieved book should have correct name")
	}
}

func TestSkillBookSystemBySkill(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	meleeBooks := sys.GetBooksForSkill("melee")
	if len(meleeBooks) == 0 {
		t.Error("Melee should have books")
	}

	for _, book := range meleeBooks {
		if book.SkillID != "melee" {
			t.Errorf("Book %s should be for melee", book.ID)
		}
	}
}

func TestSkillBookSystemByRarity(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	// Common books (rarity >= 80)
	common := sys.GetBooksByRarity(80, 100)
	if len(common) == 0 {
		t.Error("Should have common books")
	}

	// Rare books (rarity 10-30)
	rare := sys.GetBooksByRarity(10, 30)
	if len(rare) == 0 {
		t.Error("Should have rare books")
	}
}

func TestSkillBookReading(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	skills := &components.Skills{
		Levels:     map[string]int{"melee": 5},
		Experience: map[string]float64{"melee": 0},
	}
	world.AddComponent(player, skills)

	// Start reading basic melee book
	started := sys.StartReading(world, player, "tome_melee_basics")
	if !started {
		t.Fatal("Should start reading")
	}

	if !sys.IsReading(player) {
		t.Error("Should be marked as reading")
	}

	progress := sys.GetReadingProgress(player)
	if progress != 0 {
		t.Error("Progress should be 0 at start")
	}

	// Simulate partial read time (book has ReadTime: 10)
	sys.Update(world, 5.0)

	progress = sys.GetReadingProgress(player)
	if progress < 0.4 || progress > 0.6 {
		t.Errorf("Progress should be ~0.5, got %f", progress)
	}

	// Complete reading
	sys.Update(world, 6.0)

	if sys.IsReading(player) {
		t.Error("Should no longer be reading")
	}

	// Check XP was granted
	comp, _ := world.GetComponent(player, "Skills")
	updatedSkills := comp.(*components.Skills)
	if updatedSkills.Experience["melee"] < 100 {
		t.Errorf("Should have gained XP, got %f", updatedSkills.Experience["melee"])
	}
}

func TestSkillBookLevelGrant(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	skills := &components.Skills{
		Levels:     map[string]int{"melee": 50},
		Experience: map[string]float64{},
	}
	world.AddComponent(player, skills)

	// Start reading master book (LevelGrant: 2)
	started := sys.StartReading(world, player, "tome_melee_master")
	if !started {
		t.Fatal("Should start reading master tome")
	}

	// Complete reading
	sys.Update(world, 100.0)

	comp, _ := world.GetComponent(player, "Skills")
	updatedSkills := comp.(*components.Skills)
	if updatedSkills.Levels["melee"] != 52 {
		t.Errorf("Should have gained 2 levels, got %d", updatedSkills.Levels["melee"])
	}
}

func TestSkillBookRequirements(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Low level player
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 5},
		Experience: map[string]float64{},
	}
	world.AddComponent(player, skills)

	// Try to read advanced book (requires level 15)
	started := sys.StartReading(world, player, "tome_melee_advanced")
	if started {
		t.Error("Should not start reading - level too low")
	}
}

func TestSkillBookMaxLevel(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// High level player
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 25},
		Experience: map[string]float64{},
	}
	world.AddComponent(player, skills)

	// Try to read basic book (max level 20)
	started := sys.StartReading(world, player, "tome_melee_basics")
	if started {
		t.Error("Should not start reading - already beyond book's max level")
	}
}

func TestSkillBookOneTimeUse(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	skills := &components.Skills{
		Levels:     map[string]int{"melee": 5},
		Experience: map[string]float64{},
	}
	world.AddComponent(player, skills)

	// Read book once
	sys.StartReading(world, player, "tome_melee_basics")
	sys.Update(world, 100.0)

	if !sys.HasRead(player, "tome_melee_basics") {
		t.Error("Should have marked book as read")
	}

	// Try to read again
	started := sys.StartReading(world, player, "tome_melee_basics")
	if started {
		t.Error("Should not start reading - already read this one-time book")
	}
}

func TestSkillBookCancelReading(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	skills := &components.Skills{
		Levels:     map[string]int{"melee": 5},
		Experience: map[string]float64{},
	}
	world.AddComponent(player, skills)

	sys.StartReading(world, player, "tome_melee_basics")
	sys.Update(world, 5.0) // Partial read

	sys.CancelReading(player)

	if sys.IsReading(player) {
		t.Error("Should not be reading after cancel")
	}

	// No XP should have been granted
	comp, _ := world.GetComponent(player, "Skills")
	updatedSkills := comp.(*components.Skills)
	if updatedSkills.Experience["melee"] != 0 {
		t.Error("Should not have gained XP after cancel")
	}
}

func TestSkillBookGetReadAndUnread(t *testing.T) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	skills := &components.Skills{
		Levels:     map[string]int{"melee": 10},
		Experience: map[string]float64{},
	}
	world.AddComponent(player, skills)

	// Read one book
	sys.StartReading(world, player, "tome_melee_basics")
	sys.Update(world, 100.0)

	readBooks := sys.GetReadBooks(player)
	if len(readBooks) != 1 {
		t.Errorf("Should have 1 read book, got %d", len(readBooks))
	}

	unreadBooks := sys.GetUnreadBooks(world, player)
	// Should not include the read book or books with unmet requirements
	for _, book := range unreadBooks {
		if book.ID == "tome_melee_basics" {
			t.Error("Should not include already-read book")
		}
	}
}

func BenchmarkSkillBookSystem(b *testing.B) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewSkillBookSystem(reg, prog)
	}
}

func BenchmarkSkillBookReading(b *testing.B) {
	reg := NewSkillRegistry()
	prog := NewSkillProgressionSystem(100, 100)
	sys := NewSkillBookSystem(reg, prog)

	world := ecs.NewWorld()
	for i := 0; i < 100; i++ {
		player := world.CreateEntity()
		skills := &components.Skills{
			Levels:     map[string]int{"melee": 5},
			Experience: map[string]float64{},
		}
		world.AddComponent(player, skills)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world, 0.016)
	}
}

// ============================================================================
// Skill Synergy System Tests
// ============================================================================

func TestNewSkillSynergySystem(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	if sys == nil {
		t.Fatal("System should not be nil")
	}

	if sys.SynergyCount() == 0 {
		t.Error("Should have default synergies")
	}
}

func TestSkillSynergySystemRegister(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	initialCount := sys.SynergyCount()

	customSynergy := &SkillSynergy{
		ID:              "custom_synergy",
		Name:            "Custom Synergy",
		Description:     "A test synergy",
		PrimarySkill:    "melee",
		SecondarySkills: []string{"ranged"},
		MinLevel:        10,
		BonusType:       "xp_mult",
		BonusValue:      0.1,
		MaxBonus:        0.3,
	}

	sys.RegisterSynergy(customSynergy)

	if sys.SynergyCount() != initialCount+1 {
		t.Error("Should have one more synergy")
	}

	retrieved := sys.GetSynergy("custom_synergy")
	if retrieved == nil {
		t.Error("Should retrieve registered synergy")
	}
}

func TestSkillSynergySystemBySkill(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	// Melee should have synergies that benefit it
	meleeSynergies := sys.GetSynergiesForSkill("melee")
	if len(meleeSynergies) == 0 {
		t.Error("Melee should have benefiting synergies")
	}

	for _, synergy := range meleeSynergies {
		if synergy.PrimarySkill != "melee" {
			t.Errorf("Synergy %s should have melee as primary", synergy.ID)
		}
	}
}

func TestSkillSynergySystemRequiringSkill(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	// Defense should be required by some synergies
	synergies := sys.GetSynergiesRequiringSkill("defense")
	if len(synergies) == 0 {
		t.Error("Defense should be required by some synergies")
	}

	for _, synergy := range synergies {
		found := false
		for _, secondary := range synergy.SecondarySkills {
			if secondary == "defense" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Synergy %s should require defense", synergy.ID)
		}
	}
}

func TestSkillSynergyActivation(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Player with defense at 15 (activates warrior_prowess synergy for melee)
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 30, "defense": 15},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	// warrior_prowess should be active (defense 15 meets minimum)
	if !sys.IsSynergyActive(player, "warrior_prowess") {
		t.Error("warrior_prowess synergy should be active")
	}
}

func TestSkillSynergyNotActive(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Player with defense at 10 (below minimum of 15 for warrior_prowess)
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 30, "defense": 10},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	// warrior_prowess should NOT be active
	if sys.IsSynergyActive(player, "warrior_prowess") {
		t.Error("warrior_prowess synergy should not be active (defense too low)")
	}
}

func TestSkillSynergyBonus(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// High defense should give good bonus
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 30, "defense": 40},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	bonus := sys.GetSkillBonus(player, "melee")
	if bonus <= 0 {
		t.Errorf("Should have melee bonus, got %f", bonus)
	}
}

func TestSkillSynergyXPMultiplier(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	skills := &components.Skills{
		Levels:     map[string]int{"melee": 30, "defense": 20},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	multiplier := sys.GetXPMultiplier(player, "melee")
	if multiplier < 1.0 {
		t.Errorf("XP multiplier should be >= 1.0, got %f", multiplier)
	}
}

func TestSkillSynergyGetActive(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Player with multiple high skills
	skills := &components.Skills{
		Levels: map[string]int{
			"melee": 50, "defense": 30, "ranged": 25,
			"sneak": 40, "lockpicking": 20,
		},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	active := sys.GetActiveSynergies(player)
	if len(active) == 0 {
		t.Error("Should have active synergies")
	}
}

func TestSkillSynergyPotential(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Player with some skills but not meeting synergy requirements
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 30, "defense": 5},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	potential := sys.GetPotentialSynergies(world, player)

	// warrior_prowess should be potential (defense trained but not high enough)
	found := false
	for _, synergy := range potential {
		if synergy.ID == "warrior_prowess" {
			found = true
			break
		}
	}
	if !found {
		t.Error("warrior_prowess should be a potential synergy")
	}
}

func TestSkillSynergyByBonusType(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	xpSynergies := sys.GetSynergiesByBonusType("xp_mult")
	if len(xpSynergies) == 0 {
		t.Error("Should have XP multiplier synergies")
	}

	for _, synergy := range xpSynergies {
		if synergy.BonusType != "xp_mult" {
			t.Errorf("Synergy %s should have xp_mult type", synergy.ID)
		}
	}

	abilitySynergies := sys.GetSynergiesByBonusType("ability_boost")
	if len(abilitySynergies) == 0 {
		t.Error("Should have ability boost synergies")
	}
}

func TestSkillSynergyMaxBonus(t *testing.T) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Very high secondary skill - should cap at MaxBonus
	skills := &components.Skills{
		Levels:     map[string]int{"melee": 50, "defense": 100},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	sys.Update(world, 0.016)

	bonus := sys.GetSkillBonus(player, "melee")

	// warrior_prowess has MaxBonus of 0.25
	if bonus > 0.25 {
		t.Errorf("Bonus should be capped at 0.25, got %f", bonus)
	}
}

func BenchmarkSkillSynergySystem(b *testing.B) {
	reg := NewSkillRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewSkillSynergySystem(reg)
	}
}

func BenchmarkSkillSynergyUpdate(b *testing.B) {
	reg := NewSkillRegistry()
	sys := NewSkillSynergySystem(reg)

	world := ecs.NewWorld()
	for i := 0; i < 100; i++ {
		player := world.CreateEntity()
		skills := &components.Skills{
			Levels: map[string]int{
				"melee": 30, "defense": 20, "ranged": 25,
				"sneak": 15, "lockpicking": 10,
			},
			Experience: make(map[string]float64),
		}
		world.AddComponent(player, skills)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world, 0.016)
	}
}
