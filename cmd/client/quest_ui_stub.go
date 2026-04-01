//go:build noebiten

// Package main provides stub types for noebiten builds.
package main

import (
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/input"
)

// QuestUI is a stub for noebiten builds.
type QuestUI struct {
	isOpen         bool
	trackedQuestID string
}

// NewQuestUI creates a new quest UI stub for noebiten builds.
func NewQuestUI(genre string, playerEntity ecs.Entity, inputManager *input.Manager) *QuestUI {
	return &QuestUI{}
}

// IsOpen returns whether the quest log is open.
func (q *QuestUI) IsOpen() bool { return q.isOpen }

// Open opens the quest log.
func (q *QuestUI) Open() { q.isOpen = true }

// Close closes the quest log.
func (q *QuestUI) Close() { q.isOpen = false }

// Toggle toggles the quest log open/closed state.
func (q *QuestUI) Toggle() { q.isOpen = !q.isOpen }

// Update is a no-op for noebiten builds.
func (q *QuestUI) Update(world *ecs.World, dt float64) {}

// formatQuestID converts a quest ID to a readable name.
func formatQuestID(questID string) string {
	if len(questID) == 0 {
		return "Unknown Quest"
	}
	result := make([]byte, 0, len(questID))
	capitalize := true
	for i := 0; i < len(questID); i++ {
		c := questID[i]
		if c == '_' {
			result = append(result, ' ')
			capitalize = true
		} else if capitalize {
			if c >= 'a' && c <= 'z' {
				result = append(result, c-32)
			} else {
				result = append(result, c)
			}
			capitalize = false
		} else {
			result = append(result, c)
		}
	}
	return string(result)
}

// wrapTextQuest splits text into lines of max width.
func wrapTextQuest(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = 40
	}
	var lines []string
	for len(text) > maxWidth {
		breakIdx := maxWidth
		for i := maxWidth; i > 0; i-- {
			if text[i] == ' ' {
				breakIdx = i
				break
			}
		}
		lines = append(lines, text[:breakIdx])
		text = text[breakIdx:]
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:]
		}
	}
	if len(text) > 0 {
		lines = append(lines, text)
	}
	return lines
}
