// Package adapters provides wrappers around Venture's procedural generators
// that adapt them for Wyrm's open-world RPG context.
//
// These adapters bridge the gap between Venture's roguelike generators and
// Wyrm's persistent first-person open world by:
//   - Converting Venture's room-based terrain to chunk-based open world
//   - Wrapping entity generation with Wyrm's ECS component attachment
//   - Adapting faction generation to Wyrm's persistent politics system
//   - Integrating quest generation with Wyrm's consequence tracking
//
// All adapters maintain deterministic generation: same seed produces same output.
package adapters
