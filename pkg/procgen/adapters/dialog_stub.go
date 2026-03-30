//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import "math/rand"

// DialogAdapter wraps Venture's dialog generator.
// Stub implementation for headless testing.
type DialogAdapter struct{}

// NewDialogAdapter creates a new dialog adapter.
func NewDialogAdapter() *DialogAdapter { return &DialogAdapter{} }

// DialogLine holds a single dialog line with metadata.
type DialogLine struct {
	Text      string
	Tone      string
	Speaker   string
	Sentiment float64
}

// PersonalityTraits defines NPC personality for dialog generation.
type PersonalityTraits struct {
	Friendliness float64 // 0-1: hostile to friendly
	Verbosity    float64 // 0-1: terse to verbose
	Formality    float64 // 0-1: casual to formal
	Humor        float64 // 0-1: serious to humorous
	Knowledge    float64 // 0-1: ignorant to knowledgeable
}

// GetOrCreateGenerator returns a stub generator.
func (a *DialogAdapter) GetOrCreateGenerator(seed int64, genre string) interface{} { return nil }

// GenerateDialogLine generates a single dialog line.
func (a *DialogAdapter) GenerateDialogLine(seed int64, genre string) (*DialogLine, error) {
	lines := []string{"Greetings, traveler.", "What do you need?", "Safe travels."}
	rng := rand.New(rand.NewSource(seed))
	return &DialogLine{
		Text:      lines[rng.Intn(len(lines))],
		Tone:      "neutral",
		Sentiment: 0.5,
	}, nil
}

// GenerateGreeting generates a greeting based on personality.
func (a *DialogAdapter) GenerateGreeting(seed int64, genre string, personality PersonalityTraits) (*DialogLine, error) {
	return a.GenerateDialogLine(seed, genre)
}

// GenerateDialogLines generates multiple dialog lines.
func (a *DialogAdapter) GenerateDialogLines(seed int64, genre string, count int) ([]*DialogLine, error) {
	lines := make([]*DialogLine, count)
	for i := 0; i < count; i++ {
		lines[i], _ = a.GenerateDialogLine(seed+int64(i), genre)
	}
	return lines, nil
}

// PersonalityFromType returns personality traits based on type.
func PersonalityFromType(ptype string) PersonalityTraits {
	return PersonalityTraits{
		Friendliness: 0.5,
		Verbosity:    0.5,
		Formality:    0.5,
		Humor:        0.5,
		Knowledge:    0.5,
	}
}
