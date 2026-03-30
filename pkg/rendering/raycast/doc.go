// Package raycast provides the first-person raycasting renderer for Wyrm.
//
// The raycaster uses the DDA (Digital Differential Analyzer) algorithm to
// cast rays from the player's viewpoint and determine what walls and surfaces
// are visible. This enables fast first-person rendering without a full 3D engine.
//
// Key features:
//   - DDA-based ray casting for wall detection
//   - Procedural texture mapping with distance-based shading
//   - Support for multiple wall heights
//   - Floor and ceiling rendering
//   - Integration with the chunk-based world system
//
// # Build Tags
//
// This package uses conditional compilation for headless testing:
//
//   - Default build (no tags): Full Ebiten-based rendering
//   - noebiten build tag: Stub implementations for CI/testing without graphics
//
// To run tests:
//
//	go test -tags=noebiten ./pkg/rendering/raycast/...
//
// The raycast_test.go file contains comprehensive tests that work with or
// without the noebiten tag, while draw_stub.go provides stub Draw() methods
// for headless builds.
package raycast
