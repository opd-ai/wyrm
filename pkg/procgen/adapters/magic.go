//go:build !noebiten

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/magic"
)

// MagicAdapter wraps Venture's spell generator for Wyrm integration.
type MagicAdapter struct {
	generator *magic.SpellGenerator
}

// NewMagicAdapter creates a new magic adapter.
func NewMagicAdapter() *MagicAdapter {
	return &MagicAdapter{
		generator: magic.NewSpellGenerator(),
	}
}

// SpellData holds generated spell data for Wyrm integration.
type SpellData struct {
	Name        string
	Type        string
	Element     string
	Rarity      string
	Target      string
	Description string
	Tags        []string
	Seed        int64
	// Stats
	Damage   int
	Healing  int
	ManaCost int
	Cooldown float64
	CastTime float64
	Range    float64
	AreaSize float64
	Duration float64
}

// GenerateSpells creates spells using Venture's generator.
func (a *MagicAdapter) GenerateSpells(seed int64, genre string, count int) ([]*SpellData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: DefaultGenerationDifficulty,
		Depth:      DefaultGenerationDepth,
		Custom: map[string]interface{}{
			"count": count,
		},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("spell generation failed: %w", err)
	}

	spells, ok := result.([]*magic.Spell)
	if !ok {
		return nil, fmt.Errorf("invalid spell result type: expected []*magic.Spell, got %T", result)
	}

	spellData := make([]*SpellData, len(spells))
	for i, s := range spells {
		spellData[i] = convertSpell(s)
	}

	return spellData, nil
}

// GenerateSpell creates a single spell.
func (a *MagicAdapter) GenerateSpell(seed int64, genre string) (*SpellData, error) {
	spells, err := a.GenerateSpells(seed, genre, 1)
	if err != nil {
		return nil, err
	}
	if len(spells) == 0 {
		return nil, fmt.Errorf("no spell generated")
	}
	return spells[0], nil
}

// convertSpell transforms a Venture spell to Wyrm format.
func convertSpell(s *magic.Spell) *SpellData {
	tags := make([]string, len(s.Tags))
	copy(tags, s.Tags)

	return &SpellData{
		Name:        s.Name,
		Type:        s.Type.String(),
		Element:     s.Element.String(),
		Rarity:      s.Rarity.String(),
		Target:      s.Target.String(),
		Description: s.Description,
		Tags:        tags,
		Seed:        s.Seed,
		Damage:      s.Stats.Damage,
		Healing:     s.Stats.Healing,
		ManaCost:    s.Stats.ManaCost,
		Cooldown:    s.Stats.Cooldown,
		CastTime:    s.Stats.CastTime,
		Range:       s.Stats.Range,
		AreaSize:    s.Stats.AreaSize,
		Duration:    s.Stats.Duration,
	}
}

// IsOffensive returns true if the spell deals damage.
func (s *SpellData) IsOffensive() bool {
	return s.Type == "Offensive" || s.Type == "Debuff"
}

// IsSupport returns true if the spell supports allies.
func (s *SpellData) IsSupport() bool {
	return s.Type == "Healing" || s.Type == "Buff" || s.Type == "Defensive"
}

// SpellRarityMultiplier returns the stat multiplier for a given rarity.
func SpellRarityMultiplier(rarity string) float64 {
	switch rarity {
	case "Common":
		return VehicleCommonStatMultiplier
	case "Uncommon":
		return VehicleUncommonStatMultiplier
	case "Rare":
		return VehicleRareStatMultiplier
	case "Epic":
		return VehicleEpicStatMultiplier
	case "Legendary":
		return VehicleLegendaryStatMultiplier
	default:
		return VehicleCommonStatMultiplier
	}
}
