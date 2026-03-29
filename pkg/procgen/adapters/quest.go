// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
)

// QuestAdapter wraps Venture's quest generator for Wyrm's quest system.
type QuestAdapter struct {
	generator *quest.QuestGenerator
}

// NewQuestAdapter creates a new quest adapter.
func NewQuestAdapter() *QuestAdapter {
	return &QuestAdapter{
		generator: quest.NewQuestGenerator(),
	}
}

// QuestData holds generated quest data for Wyrm integration.
type QuestData struct {
	ID              string
	Name            string
	Type            string
	Description     string
	Objectives      []ObjectiveData
	RequiredLevel   int
	GiverNPC        string
	Location        string
	FactionA        string
	FactionB        string
	HasConsequences bool
	Seed            int64
}

// ObjectiveData holds objective information from generated quests.
type ObjectiveData struct {
	Type        string
	Description string
	Target      string
	Required    int
	Current     int
}

// GenerateQuests creates quests for the world using Venture's generator.
func (a *QuestAdapter) GenerateQuests(seed int64, genre string, count int, difficulty float64) ([]*QuestData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: difficulty,
		Depth:      int(difficulty * 100),
		Custom:     map[string]interface{}{"count": count},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("quest generation failed: %w", err)
	}

	ventureQuests, ok := result.([]*quest.Quest)
	if !ok {
		return nil, fmt.Errorf("invalid quest result type: expected []*quest.Quest, got %T", result)
	}

	return convertQuests(ventureQuests), nil
}

// convertQuests transforms Venture quests to Wyrm format.
func convertQuests(vq []*quest.Quest) []*QuestData {
	quests := make([]*QuestData, len(vq))
	for i, q := range vq {
		quests[i] = convertSingleQuest(q)
	}
	return quests
}

// convertSingleQuest transforms a single Venture quest.
func convertSingleQuest(q *quest.Quest) *QuestData {
	objectives := make([]ObjectiveData, len(q.Objectives))
	for j, obj := range q.Objectives {
		objectives[j] = ObjectiveData{
			Type:        q.Type.String(), // Use quest type for objective type
			Description: obj.Description,
			Target:      obj.Target,
			Required:    obj.Required,
			Current:     obj.Current,
		}
	}

	return &QuestData{
		ID:              q.ID,
		Name:            q.Name,
		Type:            q.Type.String(),
		Description:     q.Description,
		Objectives:      objectives,
		RequiredLevel:   q.RequiredLevel,
		GiverNPC:        q.GiverNPC,
		Location:        q.Location,
		FactionA:        q.FactionA,
		FactionB:        q.FactionB,
		HasConsequences: q.HasMoralConsequences,
		Seed:            q.Seed,
	}
}

// SpawnQuestEntity creates a Quest component in the ECS world.
func SpawnQuestEntity(world *ecs.World, data *QuestData) (ecs.Entity, error) {
	e := world.CreateEntity()

	questComp := &components.Quest{
		ID:             data.ID,
		CurrentStage:   0,
		Flags:          make(map[string]bool),
		Completed:      false,
		LockedBranches: make(map[string]bool),
	}

	if err := world.AddComponent(e, questComp); err != nil {
		return 0, fmt.Errorf("failed to add Quest component: %w", err)
	}

	return e, nil
}

// RegisterQuestWithSystem adds quest stage conditions to the QuestSystem.
func RegisterQuestWithSystem(qs *systems.QuestSystem, data *QuestData) {
	stages := make([]systems.QuestStageCondition, len(data.Objectives))

	for i := range data.Objectives {
		flagName := fmt.Sprintf("%s_obj_%d", data.ID, i)
		stages[i] = systems.QuestStageCondition{
			RequiredFlag: flagName,
			FromStage:    i,
			NextStage:    i + 1,
			Completes:    i == len(data.Objectives)-1,
		}
	}

	qs.DefineQuest(data.ID, stages)
}

// GenerateAndSpawnQuests generates quests and spawns them in the world.
func (a *QuestAdapter) GenerateAndSpawnQuests(world *ecs.World, qs *systems.QuestSystem, seed int64, genre string, count int) ([]ecs.Entity, error) {
	quests, err := a.GenerateQuests(seed, genre, count, 0.5)
	if err != nil {
		return nil, err
	}

	entities := make([]ecs.Entity, 0, len(quests))
	for _, q := range quests {
		e, err := SpawnQuestEntity(world, q)
		if err != nil {
			continue
		}
		RegisterQuestWithSystem(qs, q)
		entities = append(entities, e)
	}

	return entities, nil
}
