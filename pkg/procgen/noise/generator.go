// Package noise provides shared procedural noise functions.
package noise

import "math"

// NoiseType represents different noise algorithms.
type NoiseType int

const (
	// NoiseTypeValue is traditional value noise.
	NoiseTypeValue NoiseType = iota
	// NoiseTypeGradient is Perlin-style gradient noise.
	NoiseTypeGradient
)

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

// GradientNoise2D generates 2D gradient (Perlin-style) noise.
// Uses per-grid-point gradient vectors derived from a hash function.
// Returns a value in [-1, 1] range with smoother transitions than value noise.
func GradientNoise2D(x, y float64, seed int64) float64 {
	xi := int(math.Floor(x))
	yi := int(math.Floor(y))
	xf := x - float64(xi)
	yf := y - float64(yi)

	// Get gradient vectors and compute dot products at each corner
	g00 := gradientDot(xi, yi, xf, yf, seed)
	g10 := gradientDot(xi+1, yi, xf-1, yf, seed)
	g01 := gradientDot(xi, yi+1, xf, yf-1, seed)
	g11 := gradientDot(xi+1, yi+1, xf-1, yf-1, seed)

	// Use quintic smoothstep for smoother gradients
	sx := QuinticSmooth(xf)
	sy := QuinticSmooth(yf)

	// Bilinear interpolation of gradient values
	n0 := Lerp(g00, g10, sx)
	n1 := Lerp(g01, g11, sx)

	return Lerp(n0, n1, sy)
}

// GradientNoise2DNormalized generates gradient noise normalized to [0, 1] range.
func GradientNoise2DNormalized(x, y float64, seed int64) float64 {
	// Gradient noise typically outputs in ~[-0.7, 0.7] range
	// Scale and shift to [0, 1]
	n := GradientNoise2D(x, y, seed)
	return (n + 1.0) / 2.0
}

// gradientDot computes the dot product of a pseudo-random gradient with the distance vector.
func gradientDot(xi, yi int, dx, dy float64, seed int64) float64 {
	gx, gy := hashToGradient(xi, yi, seed)
	return gx*dx + gy*dy
}

// hashToGradient generates a pseudo-random unit gradient vector from grid coordinates.
func hashToGradient(x, y int, seed int64) (float64, float64) {
	// Use hash to select one of 8 gradient directions for better distribution
	h := hashCoords(x, y, seed)

	// 8 uniformly distributed gradient directions
	switch h % 8 {
	case 0:
		return 1.0, 0.0
	case 1:
		return -1.0, 0.0
	case 2:
		return 0.0, 1.0
	case 3:
		return 0.0, -1.0
	case 4:
		return 0.7071067811865476, 0.7071067811865476 // 1/sqrt(2)
	case 5:
		return -0.7071067811865476, 0.7071067811865476
	case 6:
		return 0.7071067811865476, -0.7071067811865476
	default:
		return -0.7071067811865476, -0.7071067811865476
	}
}

// hashCoords generates a deterministic integer hash from coordinates.
func hashCoords(x, y int, seed int64) uint64 {
	h := uint64(seed)
	h ^= uint64(x) * 0x9E3779B97F4A7C15
	h ^= uint64(y) * 0xBF58476D1CE4E5B9
	h = (h ^ (h >> 30)) * 0xBF58476D1CE4E5B9
	h = (h ^ (h >> 27)) * 0x94D049BB133111EB
	h ^= h >> 31
	return h
}

// HashToFloat converts coordinates to a pseudo-random float in [0, 1].
func HashToFloat(x, y int, seed int64) float64 {
	h := hashCoords(x, y, seed)
	return float64(h&0x7FFFFFFFFFFFFFFF) / float64(0x7FFFFFFFFFFFFFFF)
}

// Smoothstep applies smoothstep interpolation for smoother gradients.
func Smoothstep(t float64) float64 {
	return t * t * (3 - 2*t)
}

// QuinticSmooth applies quintic smoothstep for even smoother gradients.
// Used in improved Perlin noise for second-derivative continuity.
func QuinticSmooth(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

// Lerp performs linear interpolation between a and b by factor t.
func Lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

// NoiseFunc2D is a function type for 2D noise generators.
type NoiseFunc2D func(x, y float64, seed int64) float64

// fbmCore is the shared fractal Brownian motion implementation. It accumulates
// noise samples across octaves using the provided noise function, applying
// persistence (amplitude decay) and lacunarity (frequency growth) per octave.
func fbmCore(x, y float64, octaves int, persistence, lacunarity float64, seed int64, noiseFn NoiseFunc2D) float64 {
	total := 0.0
	amplitude := 1.0
	frequency := 1.0
	maxValue := 0.0

	for i := 0; i < octaves; i++ {
		total += noiseFn(x*frequency, y*frequency, seed+int64(i)) * amplitude
		maxValue += amplitude
		amplitude *= persistence
		frequency *= lacunarity
	}

	return total / maxValue
}

// FBM generates fractal Brownian motion using value noise.
func FBM(x, y float64, octaves int, persistence, lacunarity float64, seed int64) float64 {
	return fbmCore(x, y, octaves, persistence, lacunarity, seed, Noise2DSigned)
}

// GradientFBM generates fractal Brownian motion using gradient noise.
func GradientFBM(x, y float64, octaves int, persistence, lacunarity float64, seed int64) float64 {
	return fbmCore(x, y, octaves, persistence, lacunarity, seed, GradientNoise2D)
}

// Turbulence generates turbulent noise (absolute value FBM).
func Turbulence(x, y float64, octaves int, persistence, lacunarity float64, seed int64) float64 {
	absNoise := func(x, y float64, seed int64) float64 {
		return math.Abs(Noise2DSigned(x, y, seed))
	}
	return fbmCore(x, y, octaves, persistence, lacunarity, seed, absNoise)
}

// Ridged generates ridged multifractal noise.
func Ridged(x, y float64, octaves int, persistence, lacunarity float64, seed int64) float64 {
	total := 0.0
	amplitude := 1.0
	frequency := 1.0
	weight := 1.0

	for i := 0; i < octaves; i++ {
		signal := GradientNoise2D(x*frequency, y*frequency, seed+int64(i))
		signal = 1.0 - math.Abs(signal)
		signal *= signal
		signal *= weight
		weight = signal
		if weight > 1.0 {
			weight = 1.0
		}
		if weight < 0.0 {
			weight = 0.0
		}
		total += signal * amplitude
		amplitude *= persistence
		frequency *= lacunarity
	}

	return total
}
