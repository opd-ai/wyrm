//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import "math/rand"

// MagicAdapter wraps Venture's magic/spell generator.
// Stub implementation for headless testing.
type MagicAdapter struct{}

// NewMagicAdapter creates a new magic adapter.
func NewMagicAdapter() *MagicAdapter { return &MagicAdapter{} }

// SpellData holds generated spell information.
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

// GenerateSpells creates multiple spells.
func (a *MagicAdapter) GenerateSpells(seed int64, genre string, count int) ([]*SpellData, error) {
	spells := make([]*SpellData, count)
	for i := 0; i < count; i++ {
		spells[i], _ = a.GenerateSpell(seed+int64(i), genre)
	}
	return spells, nil
}

// GenerateSpell creates a single spell.
func (a *MagicAdapter) GenerateSpell(seed int64, genre string) (*SpellData, error) {
	rng := rand.New(rand.NewSource(seed))
	types := []string{"Destruction", "Restoration", "Illusion", "Conjuration"}
	elements := []string{"Fire", "Ice", "Lightning", "Earth"}
	rarities := []string{"common", "uncommon", "rare", "legendary"}
	return &SpellData{
		Name:        elements[rng.Intn(len(elements))] + " " + types[rng.Intn(len(types))],
		Type:        types[rng.Intn(len(types))],
		Element:     elements[rng.Intn(len(elements))],
		Rarity:      rarities[rng.Intn(len(rarities))],
		Target:      "enemy",
		Description: "A magical spell",
		Tags:        []string{"magic"},
		Seed:        seed,
		Damage:      5 + rng.Intn(50),
		ManaCost:    10 + rng.Intn(90),
		Range:       5 + rng.Float64()*20,
		Cooldown:    1 + rng.Float64()*5,
	}, nil
}

// IsOffensive checks if spell deals damage.
func (s *SpellData) IsOffensive() bool {
	return s.Type == "Offensive" || s.Type == "Debuff"
}

// IsSupport checks if spell is supportive.
func (s *SpellData) IsSupport() bool {
	return s.Type == "Healing" || s.Type == "Buff" || s.Type == "Defensive"
}

// SpellRarityMultiplier returns rarity-based damage multiplier.
func SpellRarityMultiplier(rarity string) float64 {
	switch rarity {
	case "legendary":
		return 2.0
	case "rare":
		return 1.5
	case "uncommon":
		return 1.25
	default:
		return 1.0
	}
}
