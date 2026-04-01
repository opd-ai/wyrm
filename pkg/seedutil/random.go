// Package seedutil provides deterministic seed-based random generation utilities for Wyrm.
package seedutil

import "math"

// PseudoRandom provides deterministic pseudo-random number generation.
// It uses a fast xorshift algorithm suitable for game systems that need
// reproducible results from a given seed.
type PseudoRandom struct {
	Seed    int64
	counter uint64
}

// NewPseudoRandom creates a new PseudoRandom generator with the given seed.
func NewPseudoRandom(seed int64) *PseudoRandom {
	return &PseudoRandom{Seed: seed}
}

// Float64 generates a deterministic pseudo-random float64 in [0, 1).
// Uses xorshift algorithm for high-quality distribution.
func (p *PseudoRandom) Float64() float64 {
	p.counter++
	x := uint64(p.Seed) + p.counter*6364136223846793005
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	return float64(x%10000) / 10000.0
}

// Int generates a deterministic pseudo-random integer in [0, max).
// Returns 0 if max <= 0.
func (p *PseudoRandom) Int(max int) int {
	if max <= 0 {
		return 0
	}
	return int(p.Float64() * float64(max))
}

// PseudoRandomLCG provides deterministic pseudo-random number generation
// using an LCG (Linear Congruential Generator) algorithm.
type PseudoRandomLCG struct {
	Seed    int64
	counter int64
}

// NewPseudoRandomLCG creates a new LCG-based random generator with the given seed.
func NewPseudoRandomLCG(seed int64) *PseudoRandomLCG {
	return &PseudoRandomLCG{Seed: seed}
}

// Float64 generates a deterministic pseudo-random float64 in [0, 1).
// Uses LCG algorithm for compatibility with systems expecting this pattern.
func (p *PseudoRandomLCG) Float64() float64 {
	p.counter++
	x := p.Seed*1103515245 + p.counter*12345
	return float64((x>>16)&0x7FFF) / 32768.0
}

// FormatPrefixedID creates a prefixed ID string (e.g., "CQ-123", "EV-456").
// The prefix should be 2 characters. Returns prefix + "-" + decimal number.
func FormatPrefixedID(prefix string, n int) string {
	result := make([]byte, 0, 12)
	result = append(result, prefix...)
	result = append(result, '-')
	if n == 0 {
		return string(append(result, '0'))
	}
	digits := make([]byte, 0, 8)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	for i := len(digits) - 1; i >= 0; i-- {
		result = append(result, digits[i])
	}
	return string(result)
}

// NormalizeAngle normalizes an angle in radians to the range [0, 2π).
func NormalizeAngle(angle float64) float64 {
	for angle >= 2*math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < 0 {
		angle += 2 * math.Pi
	}
	return angle
}

// ClampDelta clamps a delta value to a maximum magnitude.
func ClampDelta(delta, maxDelta float64) float64 {
	if delta > maxDelta {
		return maxDelta
	}
	if delta < -maxDelta {
		return -maxDelta
	}
	return delta
}
