//go:build noebiten

// Package main provides stub types for noebiten builds.
package main

import (
	"strings"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/input"
)

// InventoryUI is a stub for noebiten builds.
type InventoryUI struct {
	isOpen bool
}

// NewInventoryUI creates a new inventory UI stub for noebiten builds.
func NewInventoryUI(genre string, playerEntity ecs.Entity, inputManager *input.Manager) *InventoryUI {
	return &InventoryUI{}
}

// IsOpen returns whether the inventory is open.
func (ui *InventoryUI) IsOpen() bool { return ui.isOpen }

// Open opens the inventory UI.
func (ui *InventoryUI) Open() { ui.isOpen = true }

// Close closes the inventory UI.
func (ui *InventoryUI) Close() { ui.isOpen = false }

// Toggle toggles the inventory UI open/closed state.
func (ui *InventoryUI) Toggle() { ui.isOpen = !ui.isOpen }

// Update is a no-op for noebiten builds.
func (ui *InventoryUI) Update(world *ecs.World) {}

// containsStr returns whether s contains substr (case-insensitive).
func containsStr(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// isConsumable returns whether the item name indicates a consumable item.
func isConsumable(itemName string) bool {
	consumableKeywords := []string{"potion", "food", "drink", "herb", "elixir", "scroll"}
	for _, kw := range consumableKeywords {
		if containsStr(itemName, kw) {
			return true
		}
	}
	return false
}
