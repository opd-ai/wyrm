//go:build !noebiten

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/skills"
)

// SkillsAdapter wraps Venture's skill tree generator for Wyrm integration.
type SkillsAdapter struct {
	generator *skills.SkillTreeGenerator
}

// NewSkillsAdapter creates a new skills adapter.
func NewSkillsAdapter() *SkillsAdapter {
	return &SkillsAdapter{
		generator: skills.NewSkillTreeGenerator(),
	}
}

// SkillTreeData holds generated skill tree data for Wyrm integration.
type SkillTreeData struct {
	ID          string
	Name        string
	Description string
	Category    string
	Genre       string
	Nodes       []*SkillNodeData
	RootNodes   []*SkillNodeData
	MaxPoints   int
	Seed        int64
}

// SkillNodeData holds skill node data for Wyrm integration.
type SkillNodeData struct {
	Skill    *SkillData
	Children []*SkillNodeData
	X, Y     float64
}

// SkillData holds individual skill data for Wyrm integration.
type SkillData struct {
	ID           string
	Name         string
	Description  string
	Type         string
	Category     string
	Tier         string
	Level        int
	MaxLevel     int
	Requirements SkillRequirements
	Effects      []SkillEffect
	Tags         []string
	Seed         int64
}

// SkillRequirements holds skill unlock requirements.
type SkillRequirements struct {
	PlayerLevel       int
	SkillPoints       int
	PrerequisiteIDs   []string
	AttributeMinimums map[string]int
}

// SkillEffect holds a skill effect.
type SkillEffect struct {
	Type        string
	Value       float64
	IsPercent   bool
	Description string
}

// GenerateSkillTrees creates skill trees using Venture's generator.
func (a *SkillsAdapter) GenerateSkillTrees(seed int64, genre string, count int) ([]*SkillTreeData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: DefaultGenerationDifficulty,
		Custom: map[string]interface{}{
			"count": count,
		},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("skill tree generation failed: %w", err)
	}

	trees, ok := result.([]*skills.SkillTree)
	if !ok {
		return nil, fmt.Errorf("invalid skill tree result: expected []*skills.SkillTree, got %T", result)
	}

	treeData := make([]*SkillTreeData, len(trees))
	for i, tree := range trees {
		treeData[i] = convertSkillTree(tree)
	}

	return treeData, nil
}

// GenerateSkillTree creates a single skill tree.
func (a *SkillsAdapter) GenerateSkillTree(seed int64, genre string) (*SkillTreeData, error) {
	trees, err := a.GenerateSkillTrees(seed, genre, 1)
	if err != nil {
		return nil, err
	}
	if len(trees) == 0 {
		return nil, fmt.Errorf("no skill tree generated")
	}
	return trees[0], nil
}

// convertSkillTree transforms a Venture skill tree to Wyrm format.
func convertSkillTree(tree *skills.SkillTree) *SkillTreeData {
	nodes := make([]*SkillNodeData, len(tree.Nodes))
	for i, node := range tree.Nodes {
		nodes[i] = convertSkillNode(node)
	}

	rootNodes := make([]*SkillNodeData, len(tree.RootNodes))
	for i, node := range tree.RootNodes {
		rootNodes[i] = convertSkillNode(node)
	}

	return &SkillTreeData{
		ID:          tree.ID,
		Name:        tree.Name,
		Description: tree.Description,
		Category:    tree.Category.String(),
		Genre:       tree.Genre,
		Nodes:       nodes,
		RootNodes:   rootNodes,
		MaxPoints:   tree.MaxPoints,
		Seed:        tree.Seed,
	}
}

// convertSkillNode transforms a Venture skill node to Wyrm format.
func convertSkillNode(node *skills.SkillNode) *SkillNodeData {
	children := make([]*SkillNodeData, len(node.Children))
	for i, child := range node.Children {
		children[i] = convertSkillNode(child)
	}

	return &SkillNodeData{
		Skill:    convertSkill(node.Skill),
		Children: children,
		X:        float64(node.Position.X),
		Y:        float64(node.Position.Y),
	}
}

// convertSkill transforms a Venture skill to Wyrm format.
func convertSkill(skill *skills.Skill) *SkillData {
	effects := make([]SkillEffect, len(skill.Effects))
	for i, e := range skill.Effects {
		effects[i] = SkillEffect{
			Type:        e.Type,
			Value:       e.Value,
			IsPercent:   e.IsPercent,
			Description: e.Description,
		}
	}

	prereqs := make([]string, len(skill.Requirements.PrerequisiteIDs))
	copy(prereqs, skill.Requirements.PrerequisiteIDs)

	attrMins := make(map[string]int)
	for k, v := range skill.Requirements.AttributeMinimums {
		attrMins[k] = v
	}

	tags := make([]string, len(skill.Tags))
	copy(tags, skill.Tags)

	return &SkillData{
		ID:          skill.ID,
		Name:        skill.Name,
		Description: skill.Description,
		Type:        skill.Type.String(),
		Category:    skill.Category.String(),
		Tier:        skill.Tier.String(),
		Level:       skill.Level,
		MaxLevel:    skill.MaxLevel,
		Requirements: SkillRequirements{
			PlayerLevel:       skill.Requirements.PlayerLevel,
			SkillPoints:       skill.Requirements.SkillPoints,
			PrerequisiteIDs:   prereqs,
			AttributeMinimums: attrMins,
		},
		Effects: effects,
		Tags:    tags,
		Seed:    skill.Seed,
	}
}
