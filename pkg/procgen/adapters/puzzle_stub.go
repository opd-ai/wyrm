//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import "math/rand"

// PuzzleAdapter wraps Venture's puzzle generator.
// Stub implementation for headless testing.
type PuzzleAdapter struct{}

// NewPuzzleAdapter creates a new puzzle adapter.
func NewPuzzleAdapter() *PuzzleAdapter { return &PuzzleAdapter{} }

// PuzzleData holds generated puzzle information.
type PuzzleData struct {
	Name       string
	Type       string
	Difficulty int
	Elements   []PuzzleElementData
	Solution   []string
	Hints      []string
}

// PuzzleElementData holds an element of a puzzle.
type PuzzleElementData struct {
	Type     string
	State    string
	Position int
}

// GeneratePuzzle creates a puzzle of random type.
func (a *PuzzleAdapter) GeneratePuzzle(seed int64, genre string, difficulty int) (*PuzzleData, error) {
	types := []string{"switch", "sequence", "pattern", "logic"}
	rng := rand.New(rand.NewSource(seed))
	return a.GeneratePuzzleOfType(seed, genre, difficulty, types[rng.Intn(len(types))])
}

// GeneratePuzzleOfType creates a puzzle of specific type.
func (a *PuzzleAdapter) GeneratePuzzleOfType(seed int64, genre string, difficulty int, puzzleType string) (*PuzzleData, error) {
	elemCount := 3 + difficulty
	elements := make([]PuzzleElementData, elemCount)
	solution := make([]string, elemCount)
	for i := 0; i < elemCount; i++ {
		elements[i] = PuzzleElementData{Type: puzzleType, State: "inactive", Position: i}
		solution[i] = "activate_" + string(rune('A'+i))
	}
	return &PuzzleData{
		Name:       puzzleType + " puzzle",
		Type:       puzzleType,
		Difficulty: difficulty,
		Elements:   elements,
		Solution:   solution,
		Hints:      []string{"Look for patterns", "Try the obvious first"},
	}, nil
}

// GenerateDungeonPuzzles creates puzzles for a dungeon.
func (a *PuzzleAdapter) GenerateDungeonPuzzles(seed int64, genre string, dungeonDepth, roomCount int) ([]*PuzzleData, error) {
	count := roomCount / 3
	if count < 1 {
		count = 1
	}
	puzzles := make([]*PuzzleData, count)
	for i := 0; i < count; i++ {
		puzzles[i], _ = a.GeneratePuzzle(seed+int64(i), genre, dungeonDepth)
	}
	return puzzles, nil
}

// ValidateSolution checks if actions solve the puzzle.
func ValidateSolution(puzzle *PuzzleData, actions []string) bool {
	if len(actions) != len(puzzle.Solution) {
		return false
	}
	for i, action := range actions {
		if action != puzzle.Solution[i] {
			return false
		}
	}
	return true
}

// GetPuzzleHint returns a hint based on attempts.
func GetPuzzleHint(puzzle *PuzzleData, attempts int) string {
	if attempts < len(puzzle.Hints) {
		return puzzle.Hints[attempts]
	}
	return puzzle.Hints[len(puzzle.Hints)-1]
}

// GetPuzzleTypes returns available puzzle types.
func GetPuzzleTypes() []string {
	return []string{"switch", "sequence", "pattern", "logic", "riddle"}
}
