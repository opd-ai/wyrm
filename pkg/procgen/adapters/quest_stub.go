//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import (
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
)

// QuestAdapter wraps Venture's quest generator.
// Stub implementation for headless testing.
type QuestAdapter struct{}

// NewQuestAdapter creates a new quest adapter.
func NewQuestAdapter() *QuestAdapter { return &QuestAdapter{} }

// QuestData holds generated quest information.
type QuestData struct {
	ID          string
	Name        string
	Description string
	Type        string
	Difficulty  float64
	Objectives  []ObjectiveData
	Rewards     map[string]int
}

// ObjectiveData holds a single quest objective.
type ObjectiveData struct {
	Type        string
	Description string
	Target      string
	Required    int
	Current     int
}

// GenerateQuests creates multiple quests.
func (a *QuestAdapter) GenerateQuests(seed int64, genre string, count int, difficulty float64) ([]*QuestData, error) {
	quests := make([]*QuestData, count)
	rng := rand.New(rand.NewSource(seed))
	types := []string{"fetch", "kill", "escort", "explore", "deliver"}
	for i := 0; i < count; i++ {
		qSeed := seed + int64(i)*1000
		qrng := rand.New(rand.NewSource(qSeed))
		objCount := 1 + qrng.Intn(3)
		objectives := make([]ObjectiveData, objCount)
		for j := 0; j < objCount; j++ {
			objectives[j] = ObjectiveData{
				Description: "Complete objective",
				Type:        types[qrng.Intn(len(types))],
				Target:      "target",
				Required:    1 + qrng.Intn(5),
			}
		}
		quests[i] = &QuestData{
			ID:          "quest_" + string(rune('A'+i)),
			Name:        types[rng.Intn(len(types))] + " Quest",
			Description: "A quest for a brave adventurer",
			Type:        types[rng.Intn(len(types))],
			Difficulty:  difficulty,
			Objectives:  objectives,
			Rewards:     map[string]int{"gold": 50 + rng.Intn(200), "xp": 100 + rng.Intn(500)},
		}
	}
	return quests, nil
}

// SpawnQuestEntity creates a quest entity in the ECS world.
func SpawnQuestEntity(world *ecs.World, data *QuestData) (ecs.Entity, error) {
	e := world.CreateEntity()
	// Quest entities don't have position components
	return e, nil
}

// RegisterQuestWithSystem registers a quest with the quest system.
// Note: QuestSystem uses QuestStages map, not RegisterQuest method.
func RegisterQuestWithSystem(qs *systems.QuestSystem, data *QuestData) {
	conditions := make([]systems.QuestStageCondition, len(data.Objectives))
	for i, obj := range data.Objectives {
		conditions[i] = systems.QuestStageCondition{
			RequiredFlag: obj.Target + "_complete",
			FromStage:    i,
			NextStage:    i + 1,
			Completes:    i == len(data.Objectives)-1,
		}
	}
	qs.QuestStages[data.ID] = conditions
}

// GenerateAndSpawnQuests generates and registers quests.
func (a *QuestAdapter) GenerateAndSpawnQuests(world *ecs.World, qs *systems.QuestSystem, seed int64, genre string, count int) ([]ecs.Entity, error) {
	quests, err := a.GenerateQuests(seed, genre, count, 0.5)
	if err != nil {
		return nil, err
	}
	entities := make([]ecs.Entity, len(quests))
	for i, q := range quests {
		entities[i], _ = SpawnQuestEntity(world, q)
		RegisterQuestWithSystem(qs, q)
	}
	return entities, nil
}
