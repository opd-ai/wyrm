//go:build noebiten

// Package main provides stub faction UI for noebiten builds.
package main

import (
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

// getFactionInfo retrieves all faction information for the player.
func (ui *FactionUI) getFactionInfo(world *ecs.World) []FactionDisplayInfo {
	base := &factionUIBase{playerEntity: ui.playerEntity}
	return base.getFactionInfo(world)
}

// getStandingLevel returns a textual standing level based on reputation.
func (ui *FactionUI) getStandingLevel(reputation float64) string {
	return getStandingLevel(reputation)
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
