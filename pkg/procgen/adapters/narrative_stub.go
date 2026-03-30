//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import "math/rand"

// NarrativeAdapter wraps Venture's narrative generator.
// Stub implementation for headless testing.
type NarrativeAdapter struct{}

// NewNarrativeAdapter creates a new narrative adapter.
func NewNarrativeAdapter() *NarrativeAdapter { return &NarrativeAdapter{} }

// StoryArcData holds a generated story arc.
type StoryArcData struct {
	Title      string
	Theme      string
	PlotPoints []PlotPointData
	Difficulty float64
}

// PlotPointData holds a single plot point in a story arc.
type PlotPointData struct {
	Description string
	Type        string
	Choices     []PlayerChoiceData
}

// PlayerChoiceData holds a player choice in narrative.
type PlayerChoiceData struct {
	Text        string
	Consequence string
}

// GenerateStoryArc creates a story arc for the world.
func (a *NarrativeAdapter) GenerateStoryArc(seed int64, genre string, difficulty float64) (*StoryArcData, error) {
	rng := rand.New(rand.NewSource(seed))
	themes := []string{"Redemption", "Revenge", "Discovery", "Survival"}
	points := make([]PlotPointData, 3+rng.Intn(3))
	for i := range points {
		points[i] = PlotPointData{
			Description: "A pivotal moment",
			Type:        "event",
			Choices:     []PlayerChoiceData{{Text: "Accept", Consequence: "continue"}, {Text: "Refuse", Consequence: "alternate"}},
		}
	}
	return &StoryArcData{
		Title:      themes[rng.Intn(len(themes))] + " Arc",
		Theme:      themes[rng.Intn(len(themes))],
		PlotPoints: points,
		Difficulty: difficulty,
	}, nil
}

// GetActiveArcForRegion returns the active story arc for a region.
func (a *NarrativeAdapter) GetActiveArcForRegion(worldSeed int64, regionX, regionY int, genre string) (*StoryArcData, error) {
	regionSeed := worldSeed + int64(regionX)*1000 + int64(regionY)
	return a.GenerateStoryArc(regionSeed, genre, 0.5)
}

// GetWorldEventArc returns a story arc for a world event.
func (a *NarrativeAdapter) GetWorldEventArc(worldSeed int64, eventID, genre string, difficulty float64) (*StoryArcData, error) {
	eventSeed := worldSeed ^ int64(len(eventID)*31)
	return a.GenerateStoryArc(eventSeed, genre, difficulty)
}
