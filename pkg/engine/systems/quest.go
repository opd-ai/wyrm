package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// QuestSystem manages quest state, branching, and consequence flags.
type QuestSystem struct {
	// QuestStages maps quest ID to list of stage conditions.
	QuestStages map[string][]QuestStageCondition
}

// QuestStageCondition defines what must be true to advance a quest stage.
type QuestStageCondition struct {
	RequiredFlag string // Flag that must be true to advance
	FromStage    int    // Stage this condition applies from
	NextStage    int    // Stage to advance to
	Completes    bool   // If true, this transition completes the quest
	LocksBranch  string // Branch ID to lock when this transition is taken
	BranchID     string // Branch ID this condition belongs to (blocked if locked)
}

// NewQuestSystem creates a new quest system.
func NewQuestSystem() *QuestSystem {
	return &QuestSystem{
		QuestStages: make(map[string][]QuestStageCondition),
	}
}

// DefineQuest adds stage conditions for a quest.
func (s *QuestSystem) DefineQuest(questID string, stages []QuestStageCondition) {
	if s.QuestStages == nil {
		s.QuestStages = make(map[string][]QuestStageCondition)
	}
	s.QuestStages[questID] = stages
}

// Update processes quest state transitions each tick.
func (s *QuestSystem) Update(w *ecs.World, dt float64) {
	if s.QuestStages == nil {
		s.QuestStages = make(map[string][]QuestStageCondition)
	}
	for _, e := range w.Entities("Quest") {
		comp, ok := w.GetComponent(e, "Quest")
		if !ok {
			continue
		}
		quest := comp.(*components.Quest)
		s.processQuestStage(quest)
	}
}

// processQuestStage checks and advances a single quest's stage.
func (s *QuestSystem) processQuestStage(quest *components.Quest) {
	if quest.Completed {
		return
	}
	if quest.Flags == nil {
		quest.Flags = make(map[string]bool)
	}
	stages, ok := s.QuestStages[quest.ID]
	if !ok {
		return
	}
	s.checkStageConditions(quest, stages)
}

// checkStageConditions evaluates stage conditions and advances the quest.
func (s *QuestSystem) checkStageConditions(quest *components.Quest, stages []QuestStageCondition) {
	for _, cond := range stages {
		if cond.FromStage != quest.CurrentStage {
			continue
		}
		// Skip if this branch is locked
		if cond.BranchID != "" && quest.IsBranchLocked(cond.BranchID) {
			continue
		}
		if quest.Flags[cond.RequiredFlag] {
			s.advanceQuest(quest, cond)
			break
		}
	}
}

// advanceQuest moves the quest to the next stage or completes it.
func (s *QuestSystem) advanceQuest(quest *components.Quest, cond QuestStageCondition) {
	// Lock the competing branch if specified
	if cond.LocksBranch != "" {
		quest.LockBranch(cond.LocksBranch)
	}
	if cond.Completes {
		quest.Completed = true
	} else {
		quest.CurrentStage = cond.NextStage
	}
}
