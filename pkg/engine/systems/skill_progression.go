package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// SkillProgressionSystem manages skill experience gain and level-ups.
type SkillProgressionSystem struct {
	// XPPerLevel is the base XP required per level (scales with level).
	XPPerLevel float64
	// LevelCap is the maximum skill level.
	LevelCap int
}

// NewSkillProgressionSystem creates a new skill progression system.
func NewSkillProgressionSystem(xpPerLevel float64, levelCap int) *SkillProgressionSystem {
	if levelCap <= 0 {
		levelCap = 100
	}
	if xpPerLevel <= 0 {
		xpPerLevel = DefaultXPPerLevel
	}
	return &SkillProgressionSystem{
		XPPerLevel: xpPerLevel,
		LevelCap:   levelCap,
	}
}

// Update processes skill experience and level-ups each tick.
func (s *SkillProgressionSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Skills") {
		comp, ok := w.GetComponent(e, "Skills")
		if !ok {
			continue
		}
		skills := comp.(*components.Skills)
		s.processSkillProgression(skills)
	}
}

// processSkillProgression checks all skills for level-up conditions.
func (s *SkillProgressionSystem) processSkillProgression(skills *components.Skills) {
	if skills.Levels == nil || skills.Experience == nil {
		return
	}
	for skillID, xp := range skills.Experience {
		level := skills.Levels[skillID]
		if level >= s.LevelCap {
			continue
		}
		xpRequired := s.calculateXPRequired(level)
		if xp >= xpRequired {
			skills.Levels[skillID] = level + 1
			skills.Experience[skillID] = xp - xpRequired
		}
	}
}

// calculateXPRequired computes XP needed for the next level.
// Uses a simple scaling formula: base * (1 + level * LevelScalingFactor)
func (s *SkillProgressionSystem) calculateXPRequired(currentLevel int) float64 {
	return s.XPPerLevel * (BasePriceMultiplier + float64(currentLevel)*LevelScalingFactor)
}

// GrantSkillXP adds experience to a skill for an entity.
func (s *SkillProgressionSystem) GrantSkillXP(w *ecs.World, entity ecs.Entity, skillID string, xp float64) bool {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return false
	}
	skills := comp.(*components.Skills)
	if skills.Experience == nil {
		skills.Experience = make(map[string]float64)
	}
	if _, exists := skills.Levels[skillID]; !exists {
		return false
	}
	skills.Experience[skillID] += xp
	return true
}

// GetSkillLevel returns the current level of a skill for an entity.
func (s *SkillProgressionSystem) GetSkillLevel(w *ecs.World, entity ecs.Entity, skillID string) int {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return 0
	}
	skills := comp.(*components.Skills)
	if skills.Levels == nil {
		return 0
	}
	return skills.Levels[skillID]
}
