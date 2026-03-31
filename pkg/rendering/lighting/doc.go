// Package lighting implements dynamic lighting for the first-person raycaster.
//
// The lighting system supports:
//   - Point lights (torches, lamps, candles)
//   - Directional lights (sun, moon)
//   - Ambient lighting
//   - Time-of-day simulation
//   - Genre-specific lighting palettes
//
// Lights are rendered as overlays on the raycaster output, modifying pixel
// brightness based on distance to light sources.
package lighting
