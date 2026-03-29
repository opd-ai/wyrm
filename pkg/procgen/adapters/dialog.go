//go:build !noebiten

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

// DialogAdapter wraps Venture's dialog generators for Wyrm's NPC conversations.
type DialogAdapter struct {
	generators map[string]*dialog.MarkovGenerator
}

// NewDialogAdapter creates a new dialog adapter.
func NewDialogAdapter() *DialogAdapter {
	return &DialogAdapter{
		generators: make(map[string]*dialog.MarkovGenerator),
	}
}

// DialogLine represents a single line of dialog for an NPC.
type DialogLine struct {
	Text        string
	Personality PersonalityTraits
}

// PersonalityTraits holds NPC personality values affecting dialog.
type PersonalityTraits struct {
	Friendliness float64 // 0-1: hostile to friendly
	Verbosity    float64 // 0-1: terse to verbose
	Formality    float64 // 0-1: casual to formal
	Humor        float64 // 0-1: serious to humorous
	Knowledge    float64 // 0-1: ignorant to knowledgeable
}

// GetOrCreateGenerator returns a trained generator for the given genre.
func (a *DialogAdapter) GetOrCreateGenerator(seed int64, genre string) *dialog.MarkovGenerator {
	mappedGenre := mapGenreID(genre)
	key := mappedGenre

	if gen, exists := a.generators[key]; exists {
		return gen
	}

	gen := dialog.NewMarkovGenerator(seed, mappedGenre, dialog.Order2)
	corpus := dialog.GetCorpus(mappedGenre)
	if corpus != nil {
		gen.TrainFromCorpus(corpus.Sentences)
	}

	a.generators[key] = gen
	return gen
}

// GenerateDialogLine creates a single dialog line for an NPC.
func (a *DialogAdapter) GenerateDialogLine(seed int64, genre string) (*DialogLine, error) {
	gen := a.GetOrCreateGenerator(seed, genre)

	params := dialog.GenerateParams{
		MaxWords: 50,
		MinWords: 5,
	}
	text := gen.GenerateDeterministic(params)

	return &DialogLine{
		Text: text,
		Personality: PersonalityTraits{
			Friendliness: 0.5,
			Verbosity:    0.5,
			Formality:    0.5,
			Humor:        0.5,
			Knowledge:    0.5,
		},
	}, nil
}

// GenerateGreeting creates a greeting dialog line.
func (a *DialogAdapter) GenerateGreeting(seed int64, genre string, personality PersonalityTraits) (*DialogLine, error) {
	gen := a.GetOrCreateGenerator(seed, genre)

	params := dialog.GenerateParams{
		MaxWords: 20,
		MinWords: 3,
	}
	text := gen.GenerateDeterministic(params)

	return &DialogLine{
		Text:        text,
		Personality: personality,
	}, nil
}

// GenerateDialogLines creates multiple dialog lines for conversation variety.
func (a *DialogAdapter) GenerateDialogLines(seed int64, genre string, count int) ([]*DialogLine, error) {
	lines := make([]*DialogLine, count)

	for i := 0; i < count; i++ {
		lineSeed := seed + int64(i)*1000
		line, err := a.GenerateDialogLine(lineSeed, genre)
		if err != nil {
			continue
		}
		lines[i] = line
	}

	return lines, nil
}

// PersonalityFromType returns personality traits for a personality archetype.
func PersonalityFromType(ptype string) PersonalityTraits {
	switch ptype {
	case "helpful":
		return PersonalityTraits{
			Friendliness: 0.8,
			Verbosity:    0.6,
			Formality:    0.4,
			Humor:        0.6,
			Knowledge:    0.5,
		}
	case "merchant":
		return PersonalityTraits{
			Friendliness: 0.7,
			Verbosity:    0.7,
			Formality:    0.6,
			Humor:        0.4,
			Knowledge:    0.6,
		}
	case "hostile":
		return PersonalityTraits{
			Friendliness: 0.2,
			Verbosity:    0.3,
			Formality:    0.3,
			Humor:        0.1,
			Knowledge:    0.4,
		}
	case "mysterious":
		return PersonalityTraits{
			Friendliness: 0.4,
			Verbosity:    0.3,
			Formality:    0.7,
			Humor:        0.2,
			Knowledge:    0.8,
		}
	case "scholar":
		return PersonalityTraits{
			Friendliness: 0.5,
			Verbosity:    0.8,
			Formality:    0.8,
			Humor:        0.3,
			Knowledge:    0.9,
		}
	case "guard":
		return PersonalityTraits{
			Friendliness: 0.4,
			Verbosity:    0.4,
			Formality:    0.7,
			Humor:        0.2,
			Knowledge:    0.5,
		}
	default:
		return PersonalityTraits{
			Friendliness: 0.5,
			Verbosity:    0.5,
			Formality:    0.5,
			Humor:        0.5,
			Knowledge:    0.5,
		}
	}
}
