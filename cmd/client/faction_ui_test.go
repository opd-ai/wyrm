//go:build noebiten

package main

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/input"
)

func TestNewFactionUI(t *testing.T) {
	inputMgr := input.NewManager()
	ui := NewFactionUI("fantasy", 1, inputMgr)
	if ui == nil {
		t.Fatal("NewFactionUI returned nil")
	}
	if ui.IsOpen() {
		t.Error("FactionUI should start closed")
	}
}

func TestFactionUI_OpenClose(t *testing.T) {
	inputMgr := input.NewManager()
	ui := NewFactionUI("fantasy", 1, inputMgr)

	ui.Open()
	if !ui.IsOpen() {
		t.Error("FactionUI should be open after Open()")
	}

	ui.Close()
	if ui.IsOpen() {
		t.Error("FactionUI should be closed after Close()")
	}
}

func TestFactionUI_Toggle(t *testing.T) {
	inputMgr := input.NewManager()
	ui := NewFactionUI("fantasy", 1, inputMgr)

	ui.Toggle()
	if !ui.IsOpen() {
		t.Error("Toggle should open closed UI")
	}

	ui.Toggle()
	if ui.IsOpen() {
		t.Error("Toggle should close open UI")
	}
}

func TestFactionUI_GetFactionInfo(t *testing.T) {
	w := ecs.NewWorld()
	inputMgr := input.NewManager()
	player := w.CreateEntity()

	// Add faction membership
	membership := &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {
				FactionID:       "thieves_guild",
				Rank:            3,
				RankTitle:       "Pickpocket",
				Reputation:      45.0,
				QuestsCompleted: 5,
			},
			"mages_college": {
				FactionID:  "mages_college",
				Rank:       1,
				RankTitle:  "Apprentice",
				Reputation: 15.0,
			},
		},
	}
	w.AddComponent(player, membership)

	ui := NewFactionUI("fantasy", player, inputMgr)
	factions := ui.getFactionInfo(w)

	if len(factions) != 2 {
		t.Errorf("Expected 2 factions, got %d", len(factions))
	}

	// Should be sorted by reputation descending
	if factions[0].ID != "thieves_guild" {
		t.Error("Factions should be sorted by reputation descending")
	}
}

func TestFactionUI_GetStandingLevel(t *testing.T) {
	inputMgr := input.NewManager()
	ui := NewFactionUI("fantasy", 1, inputMgr)

	tests := []struct {
		rep      float64
		expected string
	}{
		{90, "Exalted"},
		{60, "Allied"},
		{30, "Friendly"},
		{0, "Neutral"},
		{-30, "Unfriendly"},
		{-60, "Hostile"},
	}

	for _, tc := range tests {
		result := ui.getStandingLevel(tc.rep)
		if result != tc.expected {
			t.Errorf("getStandingLevel(%.0f) = %s, expected %s", tc.rep, result, tc.expected)
		}
	}
}

func TestFactionUI_GetFactionDisplayName(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"thieves_guild", "Thieves Guild"},
		{"mages_college", "Mages College"},
		{"dark_brotherhood", "Dark Brotherhood"},
		{"guards", "Guards"},
	}

	for _, tc := range tests {
		result := getFactionDisplayName(tc.id)
		if result != tc.expected {
			t.Errorf("getFactionDisplayName(%s) = %s, expected %s", tc.id, result, tc.expected)
		}
	}
}

func TestFactionUI_AdjustScroll(t *testing.T) {
	inputMgr := input.NewManager()
	ui := NewFactionUI("fantasy", 1, inputMgr)

	// Test scrolling down
	ui.selectedFaction = 8
	ui.scrollOffset = 0
	ui.adjustScroll()

	if ui.scrollOffset == 0 {
		t.Error("Scroll should have adjusted for faction 8")
	}

	// Test scrolling up
	ui.selectedFaction = 0
	ui.adjustScroll()
	if ui.scrollOffset != 0 {
		t.Error("Scroll should reset to 0 for faction 0")
	}
}

func TestFactionUI_NoMembership(t *testing.T) {
	w := ecs.NewWorld()
	inputMgr := input.NewManager()
	player := w.CreateEntity()
	// No FactionMembership component

	ui := NewFactionUI("fantasy", player, inputMgr)
	factions := ui.getFactionInfo(w)

	if factions != nil {
		t.Error("Should return nil when no FactionMembership component")
	}
}
