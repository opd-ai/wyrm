//go:build noebiten

// Test file for quest UI logic (noebiten build tag for CI).
package main

import (
	"testing"
)

func TestNewQuestUI(t *testing.T) {
	ui := NewQuestUI("fantasy", 1, nil)
	if ui == nil {
		t.Fatal("NewQuestUI returned nil")
	}
	if ui.IsOpen() {
		t.Error("quest UI should not be open by default")
	}
}

func TestQuestUIToggle(t *testing.T) {
	ui := NewQuestUI("fantasy", 1, nil)

	// Initially closed
	if ui.IsOpen() {
		t.Error("quest UI should be closed initially")
	}

	// Open
	ui.Open()
	if !ui.IsOpen() {
		t.Error("quest UI should be open after Open()")
	}

	// Close
	ui.Close()
	if ui.IsOpen() {
		t.Error("quest UI should be closed after Close()")
	}

	// Toggle to open
	ui.Toggle()
	if !ui.IsOpen() {
		t.Error("quest UI should be open after Toggle()")
	}

	// Toggle to close
	ui.Toggle()
	if ui.IsOpen() {
		t.Error("quest UI should be closed after second Toggle()")
	}
}

func TestFormatQuestID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"famine_A0", "Famine A0"},
		{"war_quest_B1", "War Quest B1"},
		{"", "Unknown Quest"},
		{"simple", "Simple"},
		{"multiple_word_quest", "Multiple Word Quest"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatQuestID(tt.input)
			if result != tt.expected {
				t.Errorf("formatQuestID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWrapTextQuest(t *testing.T) {
	tests := []struct {
		text     string
		maxWidth int
		wantLen  int
	}{
		{"short text", 40, 1},
		{"", 40, 0},
		{"a longer text that needs wrapping at some point", 20, 3},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			lines := wrapTextQuest(tt.text, tt.maxWidth)
			if len(lines) != tt.wantLen {
				t.Errorf("wrapTextQuest(%q, %d) returned %d lines, want %d", tt.text, tt.maxWidth, len(lines), tt.wantLen)
			}
		})
	}
}
