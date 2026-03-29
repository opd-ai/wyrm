//go:build noebiten

// Package main contains tests for the Wyrm client.
// These tests use the noebiten build tag to run without Ebiten initialization.
package main

import (
	"testing"
)

func TestHeightToWallType(t *testing.T) {
	tests := []struct {
		name      string
		height    float64
		threshold float64
		expected  int
	}{
		{"below threshold returns no wall", 0.3, 0.5, 0},
		{"at threshold returns no wall", 0.5, 0.5, 0},
		{"just above threshold returns low wall", 0.51, 0.5, 1},
		{"at medium boundary returns low wall", 0.6, 0.5, 1},
		{"above medium boundary returns medium wall", 0.61, 0.5, 2},
		{"at high boundary returns medium wall", 0.8, 0.5, 2},
		{"above high boundary returns high wall", 0.81, 0.5, 3},
		{"max height returns high wall", 1.0, 0.5, 3},
		{"zero height returns no wall", 0.0, 0.5, 0},
		{"negative height returns no wall", -0.1, 0.5, 0},
		{"custom threshold low", 0.3, 0.2, 1},
		{"custom threshold high", 0.3, 0.4, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := heightToWallType(tt.height, tt.threshold)
			if result != tt.expected {
				t.Errorf("heightToWallType(%f, %f) = %d, want %d", tt.height, tt.threshold, result, tt.expected)
			}
		})
	}
}

func TestHeightToWallType_Consistency(t *testing.T) {
	// Test that the function is consistent across multiple calls
	threshold := 0.5
	for i := 0; i < 100; i++ {
		height := float64(i) / 100.0
		result1 := heightToWallType(height, threshold)
		result2 := heightToWallType(height, threshold)
		if result1 != result2 {
			t.Errorf("heightToWallType(%f, %f) inconsistent: got %d and %d", height, threshold, result1, result2)
		}
	}
}

func TestHeightToWallType_Boundaries(t *testing.T) {
	threshold := 0.5

	// Test boundary between no wall and low wall
	noWall := heightToWallType(threshold, threshold)
	lowWall := heightToWallType(threshold+0.01, threshold)
	if noWall != 0 {
		t.Errorf("height at threshold should return 0, got %d", noWall)
	}
	if lowWall != 1 {
		t.Errorf("height just above threshold should return 1, got %d", lowWall)
	}

	// Test boundary between low wall and medium wall (at 0.6)
	stillLow := heightToWallType(0.6, threshold)
	medium := heightToWallType(0.61, threshold)
	if stillLow != 1 {
		t.Errorf("height at 0.6 should return 1, got %d", stillLow)
	}
	if medium != 2 {
		t.Errorf("height above 0.6 should return 2, got %d", medium)
	}

	// Test boundary between medium wall and high wall (at 0.8)
	stillMedium := heightToWallType(0.8, threshold)
	high := heightToWallType(0.81, threshold)
	if stillMedium != 2 {
		t.Errorf("height at 0.8 should return 2, got %d", stillMedium)
	}
	if high != 3 {
		t.Errorf("height above 0.8 should return 3, got %d", high)
	}
}

func BenchmarkHeightToWallType(b *testing.B) {
	threshold := 0.5
	for i := 0; i < b.N; i++ {
		height := float64(i%100) / 100.0
		_ = heightToWallType(height, threshold)
	}
}
