//go:build !noebiten

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/puzzle"
)

// PuzzleAdapter wraps Venture's puzzle generator for Wyrm's dungeon system.
type PuzzleAdapter struct {
	generator *puzzle.Generator
}

// NewPuzzleAdapter creates a new puzzle adapter.
func NewPuzzleAdapter() *PuzzleAdapter {
	return &PuzzleAdapter{
		generator: puzzle.NewGenerator(),
	}
}

// PuzzleData holds puzzle information adapted for Wyrm's dungeons.
type PuzzleData struct {
	ID           string
	Type         string
	Difficulty   int
	Solution     []string
	ElementCount int
	Elements     []PuzzleElementData
	TimeLimit    float64
	MaxAttempts  int
	HintText     string
	Description  string
	RewardType   string
}

// PuzzleElementData holds puzzle element information adapted for Wyrm.
type PuzzleElementData struct {
	ID           string
	ElementType  string
	PositionX    int
	PositionY    int
	State        interface{}
	Interactable bool
}

// GeneratePuzzle generates a puzzle for a dungeon room.
func (a *PuzzleAdapter) GeneratePuzzle(seed int64, genre string, difficulty int) (*PuzzleData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: float64(difficulty) / 10.0,
		Depth:      difficulty,
		Custom:     map[string]interface{}{},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("puzzle generation failed: %w", err)
	}

	p, ok := result.(*puzzle.Puzzle)
	if !ok {
		return nil, fmt.Errorf("invalid puzzle result type: expected *puzzle.Puzzle, got %T", result)
	}

	return convertPuzzle(p), nil
}

// GeneratePuzzleOfType generates a puzzle of a specific type.
func (a *PuzzleAdapter) GeneratePuzzleOfType(seed int64, genre string, difficulty int, puzzleType string) (*PuzzleData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: float64(difficulty) / 10.0,
		Depth:      difficulty,
		Custom: map[string]interface{}{
			"type": puzzleType,
		},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("puzzle generation failed: %w", err)
	}

	p, ok := result.(*puzzle.Puzzle)
	if !ok {
		return nil, fmt.Errorf("invalid puzzle result type: expected *puzzle.Puzzle, got %T", result)
	}

	return convertPuzzle(p), nil
}

// GenerateDungeonPuzzles generates a set of puzzles for a dungeon level.
func (a *PuzzleAdapter) GenerateDungeonPuzzles(seed int64, genre string, dungeonDepth, roomCount int) ([]*PuzzleData, error) {
	// Number of puzzles scales with room count and depth
	puzzleCount := roomCount/5 + dungeonDepth/3
	if puzzleCount < 1 {
		puzzleCount = 1
	}
	if puzzleCount > 10 {
		puzzleCount = 10
	}

	puzzles := make([]*PuzzleData, 0, puzzleCount)
	for i := 0; i < puzzleCount; i++ {
		// Difficulty increases deeper in dungeon
		difficulty := dungeonDepth + i/2
		if difficulty > 10 {
			difficulty = 10
		}

		p, err := a.GeneratePuzzle(seed+int64(i)*1000, genre, difficulty)
		if err != nil {
			continue
		}
		puzzles = append(puzzles, p)
	}
	return puzzles, nil
}

// convertPuzzle converts Venture puzzle to Wyrm format.
func convertPuzzle(p *puzzle.Puzzle) *PuzzleData {
	elements := make([]PuzzleElementData, len(p.Elements))
	for i, elem := range p.Elements {
		elements[i] = PuzzleElementData{
			ID:           elem.ID,
			ElementType:  elem.ElementType,
			PositionX:    elem.Position[0],
			PositionY:    elem.Position[1],
			State:        elem.State,
			Interactable: elem.Interactable,
		}
	}

	return &PuzzleData{
		ID:           p.ID,
		Type:         string(p.Type),
		Difficulty:   p.Difficulty,
		Solution:     p.Solution,
		ElementCount: p.ElementCount,
		Elements:     elements,
		TimeLimit:    p.TimeLimit,
		MaxAttempts:  p.MaxAttempts,
		HintText:     p.HintText,
		Description:  p.Description,
		RewardType:   p.RewardType,
	}
}

// ValidateSolution checks if a sequence of actions solves the puzzle.
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

// GetPuzzleHint returns a hint for the puzzle based on attempts.
func GetPuzzleHint(puzzle *PuzzleData, attempts int) string {
	if attempts < 2 {
		return "" // No hint yet
	}
	if attempts >= 5 {
		// After 5 attempts, reveal first step
		if len(puzzle.Solution) > 0 {
			return "First step: " + puzzle.Solution[0]
		}
	}
	return puzzle.HintText
}

// GetPuzzleTypes returns all available puzzle types.
func GetPuzzleTypes() []string {
	return []string{
		"pressure_plate",
		"lever",
		"pattern_match",
		"block_push",
		"tile_rotation",
		"sequence",
		"riddle",
	}
}
