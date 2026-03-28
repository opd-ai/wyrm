// Package noise provides shared procedural noise functions.
package noise

import "math"

// Noise2D generates 2D value noise for the given coordinates.
// Returns a value in [0, 1] range.
func Noise2D(x, y float64, seed int64) float64 {
	xi := int(math.Floor(x))
	yi := int(math.Floor(y))
	xf := x - float64(xi)
	yf := y - float64(yi)

	// Get corner values
	v00 := HashToFloat(xi, yi, seed)
	v10 := HashToFloat(xi+1, yi, seed)
	v01 := HashToFloat(xi, yi+1, seed)
	v11 := HashToFloat(xi+1, yi+1, seed)

	// Smoothstep interpolation
	sx := Smoothstep(xf)
	sy := Smoothstep(yf)

	// Bilinear interpolation
	v0 := Lerp(v00, v10, sx)
	v1 := Lerp(v01, v11, sx)

	return Lerp(v0, v1, sy)
}

// Noise2DSigned generates 2D value noise in [-1, 1] range.
func Noise2DSigned(x, y float64, seed int64) float64 {
	return Noise2D(x, y, seed)*2 - 1
}

// HashToFloat converts coordinates to a pseudo-random float in [0, 1].
func HashToFloat(x, y int, seed int64) float64 {
	h := uint64(seed)
	h ^= uint64(x) * 0x9E3779B97F4A7C15
	h ^= uint64(y) * 0xBF58476D1CE4E5B9
	h = (h ^ (h >> 30)) * 0xBF58476D1CE4E5B9
	h = (h ^ (h >> 27)) * 0x94D049BB133111EB
	h ^= h >> 31
	return float64(h&0x7FFFFFFFFFFFFFFF) / float64(0x7FFFFFFFFFFFFFFF)
}

// Smoothstep applies smoothstep interpolation for smoother gradients.
func Smoothstep(t float64) float64 {
	return t * t * (3 - 2*t)
}

// Lerp performs linear interpolation between a and b by factor t.
func Lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}
