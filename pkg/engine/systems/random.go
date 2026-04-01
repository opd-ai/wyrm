// Package systems provides ECS system implementations for Wyrm.
//
// This file contains shared random number generation utilities used across
// multiple systems to ensure deterministic, seed-based procedural generation.
package systems

import "github.com/opd-ai/wyrm/pkg/seedutil"

// PseudoRandom is an alias for seedutil.PseudoRandom for local convenience.
type PseudoRandom = seedutil.PseudoRandom

// NewPseudoRandom creates a new PseudoRandom generator with the given seed.
func NewPseudoRandom(seed int64) *PseudoRandom {
	return seedutil.NewPseudoRandom(seed)
}

// PseudoRandomLCG is an alias for seedutil.PseudoRandomLCG for local convenience.
type PseudoRandomLCG = seedutil.PseudoRandomLCG

// NewPseudoRandomLCG creates a new LCG-based random generator with the given seed.
func NewPseudoRandomLCG(seed int64) *PseudoRandomLCG {
	return seedutil.NewPseudoRandomLCG(seed)
}
