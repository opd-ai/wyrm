//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import "math/rand"

// SkillsAdapter wraps Venture's skill tree generator.
// Stub implementation for headless testing.
type SkillsAdapter struct{}

// NewSkillsAdapter creates a new skills adapter.
func NewSkillsAdapter() *SkillsAdapter { return &SkillsAdapter{} }

// SkillTreeData holds a generated skill tree.
type SkillTreeData struct {
	Name   string
	School string
	Nodes  []SkillNodeData
}

// SkillNodeData holds a node in the skill tree.
type SkillNodeData struct {
	ID            string
	Name          string
	Description   string
	Skill         *SkillData
	Prerequisites []string
	Cost          int
}

// SkillData holds skill information.
type SkillData struct {
	Name         string
	Description  string
	MaxLevel     int
	Requirements *SkillRequirements
	Effects      []SkillEffect
}

// SkillRequirements holds skill unlock requirements.
type SkillRequirements struct {
	Level      int
	Attributes map[string]int
}

// SkillEffect holds a skill's effect.
type SkillEffect struct {
	Type   string
	Value  float64
	Target string
}

// GenerateSkillTrees creates multiple skill trees.
func (a *SkillsAdapter) GenerateSkillTrees(seed int64, genre string, count int) ([]*SkillTreeData, error) {
	trees := make([]*SkillTreeData, count)
	for i := 0; i < count; i++ {
		trees[i], _ = a.GenerateSkillTree(seed+int64(i)*1000, genre)
	}
	return trees, nil
}

// GenerateSkillTree creates a single skill tree.
func (a *SkillsAdapter) GenerateSkillTree(seed int64, genre string) (*SkillTreeData, error) {
	rng := rand.New(rand.NewSource(seed))
	schools := []string{"Combat", "Magic", "Stealth", "Crafting", "Social"}
	school := schools[rng.Intn(len(schools))]

	nodeCount := 5 + rng.Intn(5)
	nodes := make([]SkillNodeData, nodeCount)
	for i := 0; i < nodeCount; i++ {
		prereqs := []string{}
		if i > 0 {
			prereqs = []string{nodes[rng.Intn(i)].ID}
		}
		nodes[i] = SkillNodeData{
			ID:          school + "_node_" + string(rune('A'+i)),
			Name:        school + " Skill " + string(rune('A'+i)),
			Description: "Improves " + school + " abilities",
			Skill: &SkillData{
				Name:         school + " Mastery",
				Description:  "Increases effectiveness",
				MaxLevel:     5,
				Requirements: &SkillRequirements{Level: i * 5},
				Effects:      []SkillEffect{{Type: "bonus", Value: float64(5 + i*2), Target: school}},
			},
			Prerequisites: prereqs,
			Cost:          10 + i*5,
		}
	}

	return &SkillTreeData{
		Name:   school + " Tree",
		School: school,
		Nodes:  nodes,
	}, nil
}
