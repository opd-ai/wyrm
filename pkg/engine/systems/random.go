// Package systems provides ECS system implementations for Wyrm.
//
// This file contains shared random number generation utilities used across
// multiple systems to ensure deterministic, seed-based procedural generation.
package systems

import "github.com/opd-ai/wyrm/pkg/util"

// PseudoRandom is an alias for util.PseudoRandom for local convenience.
type PseudoRandom = util.PseudoRandom

// NewPseudoRandom creates a new PseudoRandom generator with the given seed.
func NewPseudoRandom(seed int64) *PseudoRandom {
	return util.NewPseudoRandom(seed)
}

// PseudoRandomLCG is an alias for util.PseudoRandomLCG for local convenience.
type PseudoRandomLCG = util.PseudoRandomLCG

// NewPseudoRandomLCG creates a new LCG-based random generator with the given seed.
func NewPseudoRandomLCG(seed int64) *PseudoRandomLCG {
	return util.NewPseudoRandomLCG(seed)
}
