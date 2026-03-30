// Package main contains pure utility functions for the Wyrm client.
// These functions have no external dependencies and can be tested with noebiten tag.
package main

// heightToWallType converts a height value to a wall type.
// Returns:
//   - 0: No wall (height <= threshold)
//   - 1: Low wall (height > threshold && height <= 0.6)
//   - 2: Medium wall (height > 0.6 && height <= 0.8)
//   - 3: High wall (height > 0.8)
func heightToWallType(height, threshold float64) int {
	if height <= threshold {
		return 0 // No wall
	}
	if height > 0.8 {
		return 3 // High wall (blue)
	}
	if height > 0.6 {
		return 2 // Medium wall (green)
	}
	return 1 // Low wall (red-brown)
}
