//go:build noebiten

// Test file for inventory UI logic (noebiten build tag for CI).
package main

import (
	"testing"
)

func TestNewInventoryUI(t *testing.T) {
	ui := NewInventoryUI("fantasy", 1, nil)
	if ui == nil {
		t.Fatal("NewInventoryUI returned nil")
	}
	if ui.IsOpen() {
		t.Error("inventory should not be open by default")
	}
}

func TestInventoryUIToggle(t *testing.T) {
	ui := NewInventoryUI("fantasy", 1, nil)

	// Initially closed
	if ui.IsOpen() {
		t.Error("inventory should be closed initially")
	}

	// Open
	ui.Open()
	if !ui.IsOpen() {
		t.Error("inventory should be open after Open()")
	}

	// Close
	ui.Close()
	if ui.IsOpen() {
		t.Error("inventory should be closed after Close()")
	}

	// Toggle to open
	ui.Toggle()
	if !ui.IsOpen() {
		t.Error("inventory should be open after Toggle()")
	}

	// Toggle to close
	ui.Toggle()
	if ui.IsOpen() {
		t.Error("inventory should be closed after second Toggle()")
	}
}

func TestContainsStr(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"Health Potion", "potion", true},
		{"Health Potion", "POTION", true},
		{"Mana Elixir", "mana", true},
		{"Iron Sword", "potion", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := containsStr(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("containsStr(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestIsConsumable(t *testing.T) {
	tests := []struct {
		itemName string
		want     bool
	}{
		{"Health Potion", true},
		{"Mana Potion", true},
		{"Food Ration", true},
		{"Healing Herbs", true},
		{"Iron Sword", false},
		{"Leather Armor", false},
	}

	for _, tt := range tests {
		t.Run(tt.itemName, func(t *testing.T) {
			got := isConsumable(tt.itemName)
			if got != tt.want {
				t.Errorf("isConsumable(%q) = %v, want %v", tt.itemName, got, tt.want)
			}
		})
	}
}
