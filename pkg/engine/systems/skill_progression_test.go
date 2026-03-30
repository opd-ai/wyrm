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
