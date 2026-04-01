//go:build noebiten

// Package main provides stub faction UI for noebiten builds.
package main

import (
	"sort"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/input"
)

// FactionUI stub for noebiten builds.
type FactionUI struct {
	isOpen          bool
	playerEntity    ecs.Entity
	selectedFaction int
	scrollOffset    int
	genre           string
}

// NewFactionUI creates a stub faction UI.
func NewFactionUI(genre string, playerEntity ecs.Entity, inputManager *input.Manager) *FactionUI {
	return &FactionUI{
		isOpen:       false,
		playerEntity: playerEntity,
		genre:        genre,
	}
}

// IsOpen returns whether the faction UI is open.
func (ui *FactionUI) IsOpen() bool {
	return ui.isOpen
}

// Open opens the faction UI.
func (ui *FactionUI) Open() {
	ui.isOpen = true
	ui.selectedFaction = 0
	ui.scrollOffset = 0
}

// Close closes the faction UI.
func (ui *FactionUI) Close() {
	ui.isOpen = false
}

// Toggle toggles the faction UI open/closed state.
func (ui *FactionUI) Toggle() {
	if ui.isOpen {
		ui.Close()
	} else {
		ui.Open()
	}
}

// Update is a no-op for stub.
func (ui *FactionUI) Update(world *ecs.World) {}

// FactionDisplayInfo holds faction display data.
type FactionDisplayInfo struct {
	ID              string
	Name            string
	Reputation      float64
	Rank            int
	RankTitle       string
	XP              int
	XPToNext        int
	IsMember        bool
	IsExalted       bool
	QuestsCompleted int
	StandingLevel   string
}

// getFactionInfo retrieves all faction information for the player.
func (ui *FactionUI) getFactionInfo(world *ecs.World) []FactionDisplayInfo {
	membershipComp, ok := world.GetComponent(ui.playerEntity, "FactionMembership")
	if !ok {
		return nil
	}
	membership := membershipComp.(*components.FactionMembership)

	var factions []FactionDisplayInfo

	if membership.Memberships == nil {
		return factions
	}

	for factionID, info := range membership.Memberships {
		fdi := FactionDisplayInfo{
			ID:              factionID,
			Name:            ui.getFactionDisplayName(factionID),
			Reputation:      info.Reputation,
			Rank:            info.Rank,
			RankTitle:       info.RankTitle,
			XP:              info.XP,
			XPToNext:        info.XPToNext,
			IsMember:        info.Rank > 0,
			IsExalted:       info.IsExalted,
			QuestsCompleted: info.QuestsCompleted,
			StandingLevel:   ui.getStandingLevel(info.Reputation),
		}
		factions = append(factions, fdi)
	}

	sort.Slice(factions, func(i, j int) bool {
		return factions[i].Reputation > factions[j].Reputation
	})

	return factions
}

// getFactionDisplayName returns a formatted faction name.
func (ui *FactionUI) getFactionDisplayName(factionID string) string {
	name := ""
	capitalize := true
	for _, c := range factionID {
		if c == '_' {
			name += " "
			capitalize = true
		} else {
			if capitalize && c >= 'a' && c <= 'z' {
				name += string(c - 32)
				capitalize = false
			} else {
				name += string(c)
			}
		}
	}
	return name
}

// getStandingLevel returns a textual standing level based on reputation.
func (ui *FactionUI) getStandingLevel(reputation float64) string {
	switch {
	case reputation >= 80:
		return "Exalted"
	case reputation >= 50:
		return "Allied"
	case reputation >= 20:
		return "Friendly"
	case reputation >= -20:
		return "Neutral"
	case reputation >= -50:
		return "Unfriendly"
	default:
		return "Hostile"
	}
}

// adjustScroll adjusts scroll offset to keep selected faction visible.
func (ui *FactionUI) adjustScroll() {
	visibleRows := 6
	if ui.selectedFaction < ui.scrollOffset {
		ui.scrollOffset = ui.selectedFaction
	}
	if ui.selectedFaction >= ui.scrollOffset+visibleRows {
		ui.scrollOffset = ui.selectedFaction - visibleRows + 1
	}
}
