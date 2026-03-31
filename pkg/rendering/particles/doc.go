// Package particles implements a procedural particle system for weather,
// combat, and environmental effects in the first-person raycaster.
//
// The particle system supports:
//   - Weather effects (rain, snow, dust, ash)
//   - Combat effects (blood, sparks, magic)
//   - Environmental effects (smoke, fire, fog wisps)
//
// Particles are rendered as screen-space effects, applied after the raycaster
// draws walls and sprites but before post-processing.
//
// All particles are procedurally generated from seed values to maintain the
// zero-external-assets constraint.
package particles
