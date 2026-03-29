// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/narrative"
)

// NarrativeAdapter wraps Venture's narrative generator for Wyrm's world event system.
type NarrativeAdapter struct {
	generator *narrative.StoryArcGenerator
}

// NewNarrativeAdapter creates a new narrative adapter.
func NewNarrativeAdapter() *NarrativeAdapter {
	return &NarrativeAdapter{
		generator: narrative.NewStoryArcGenerator(),
	}
}

// StoryArcData holds story arc information adapted for Wyrm.
type StoryArcData struct {
	Title        string
	MainConflict string
	Antagonist   string
	Ally         string
	PlotPoints   []PlotPointData
	Endings      []string
	Genre        string
	Difficulty   float64
	Seed         int64
}

// PlotPointData holds plot point information adapted for Wyrm.
type PlotPointData struct {
	Act               int
	Type              string
	Description       string
	Participants      []string
	Location          string
	TriggerConditions []string
	Consequences      []string
	PlayerChoices     []PlayerChoiceData
}

// PlayerChoiceData holds player choice information adapted for Wyrm.
type PlayerChoiceData struct {
	Description         string
	Options             []string
	Consequences        [][]string
	RelationshipImpacts []map[string]float64
}

// GenerateStoryArc generates a story arc for world events.
func (a *NarrativeAdapter) GenerateStoryArc(seed int64, genre string, difficulty float64) (*StoryArcData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: difficulty,
		Custom:     map[string]interface{}{},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("story arc generation failed: %w", err)
	}

	arc, ok := result.(*narrative.StoryArc)
	if !ok {
		return nil, fmt.Errorf("invalid story arc result type: expected *narrative.StoryArc, got %T", result)
	}

	return convertStoryArc(arc), nil
}

// convertStoryArc converts Venture's StoryArc to Wyrm's StoryArcData.
func convertStoryArc(arc *narrative.StoryArc) *StoryArcData {
	plotPoints := make([]PlotPointData, len(arc.PlotPoints))
	for i, pp := range arc.PlotPoints {
		plotPoints[i] = convertPlotPoint(pp)
	}

	return &StoryArcData{
		Title:        arc.Title,
		MainConflict: arc.MainConflict,
		Antagonist:   arc.Antagonist,
		Ally:         arc.Ally,
		PlotPoints:   plotPoints,
		Endings:      arc.PossibleEndings,
		Genre:        arc.Genre,
		Difficulty:   arc.Difficulty,
		Seed:         arc.Seed,
	}
}

// convertPlotPoint converts a Venture PlotPoint to Wyrm's PlotPointData.
func convertPlotPoint(pp narrative.PlotPoint) PlotPointData {
	choices := make([]PlayerChoiceData, len(pp.PlayerChoices))
	for i, pc := range pp.PlayerChoices {
		choices[i] = PlayerChoiceData{
			Description:         pc.Description,
			Options:             pc.Options,
			Consequences:        pc.Consequences,
			RelationshipImpacts: pc.RelationshipImpacts,
		}
	}

	return PlotPointData{
		Act:               pp.Act,
		Type:              pp.Type,
		Description:       pp.Description,
		Participants:      pp.Participants,
		Location:          pp.Location,
		TriggerConditions: pp.TriggerConditions,
		Consequences:      pp.Consequences,
		PlayerChoices:     choices,
	}
}

// GetActiveArcForRegion generates a regional story arc seed-derived from coordinates.
func (a *NarrativeAdapter) GetActiveArcForRegion(worldSeed int64, regionX, regionY int, genre string) (*StoryArcData, error) {
	// Derive a regional seed using FNV-like mixing
	regionSeed := worldSeed + int64(regionX)*31337 + int64(regionY)*65537
	return a.GenerateStoryArc(regionSeed, genre, 0.5)
}

// GetWorldEventArc generates a world-scale event arc (sieges, plagues, etc.).
func (a *NarrativeAdapter) GetWorldEventArc(worldSeed int64, eventID string, genre string, difficulty float64) (*StoryArcData, error) {
	// Derive event seed from world seed and event ID
	eventSeed := worldSeed
	for _, c := range eventID {
		eventSeed = eventSeed*31 + int64(c)
	}
	return a.GenerateStoryArc(eventSeed, genre, difficulty)
}
