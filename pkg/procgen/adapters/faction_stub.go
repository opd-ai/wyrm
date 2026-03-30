//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import (
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/engine/systems"
)

// FactionAdapter wraps Venture's faction generator.
// Stub implementation for headless testing.
type FactionAdapter struct{}

// NewFactionAdapter creates a new faction adapter.
func NewFactionAdapter() *FactionAdapter { return &FactionAdapter{} }

// FactionData holds generated faction information.
type FactionData struct {
	ID             string
	Name           string
	Type           string
	Description    string
	MemberCount    int
	TerritoryColor [4]uint8
	Relationships  map[string]int
}

// GenerateFactions creates multiple factions for the world.
func (a *FactionAdapter) GenerateFactions(seed int64, genre string, depth int) ([]*FactionData, error) {
	rng := rand.New(rand.NewSource(seed))
	names := []string{"Warriors Guild", "Merchants Union", "Thieves Brotherhood", "Mages Circle", "Royal Guard"}
	count := 3 + rng.Intn(3)
	factions := make([]*FactionData, count)
	for i := 0; i < count; i++ {
		factions[i] = &FactionData{
			ID:             "faction_" + string(rune('A'+i)),
			Name:           names[i%len(names)],
			Type:           "guild",
			Description:    "A " + names[i%len(names)],
			MemberCount:    20 + rng.Intn(80),
			TerritoryColor: [4]uint8{uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256)), 255},
			Relationships:  make(map[string]int),
		}
	}
	return factions, nil
}

// RegisterFactionsWithPoliticsSystem registers factions with the politics system.
func RegisterFactionsWithPoliticsSystem(fps *systems.FactionPoliticsSystem, factions []*FactionData) {
	for _, f1 := range factions {
		for f2ID, relationship := range f1.Relationships {
			rel := relationshipToFactionRelation(relationship)
			fps.SetRelation(f1.ID, f2ID, rel)
		}
	}
}

// relationshipToFactionRelation converts numeric relationship to enum.
func relationshipToFactionRelation(relationship int) systems.FactionRelation {
	if relationship <= -50 {
		return systems.RelationHostile
	}
	if relationship >= 51 {
		return systems.RelationAlly
	}
	return systems.RelationNeutral
}
