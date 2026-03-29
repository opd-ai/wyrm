// Package texture provides procedural texture generation.
//
// Genre patterns define how texture generation varies across game themes.
// Each genre has distinct noise settings, pattern types, and visual characteristics.
package texture

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/procgen/noise"
)

// GenrePatternConfig holds genre-specific texture generation parameters.
type GenrePatternConfig struct {
	// NoiseScale controls the frequency of the base noise pattern.
	NoiseScale float64
	// PatternType selects which pattern algorithm to use.
	PatternType PatternType
	// DetailLevel controls the amount of fine detail (0.0-1.0).
	DetailLevel float64
	// Contrast affects the difference between light and dark areas (0.0-2.0).
	Contrast float64
	// Saturation controls color intensity (0.0-1.0).
	Saturation float64
	// SecondaryNoiseScale for layered effects.
	SecondaryNoiseScale float64
}

// PatternType identifies the procedural pattern algorithm.
type PatternType int

const (
	// PatternNoise uses standard simplex noise.
	PatternNoise PatternType = iota
	// PatternGrid creates tech-style grid patterns.
	PatternGrid
	// PatternVoronoi creates organic cell patterns.
	PatternVoronoi
	// PatternDistortion applies wavy distortion.
	PatternDistortion
	// PatternLayered combines multiple noise layers.
	PatternLayered
)

// genrePatternConfigs maps genre to its pattern configuration.
var genrePatternConfigs = map[string]GenrePatternConfig{
	"fantasy": {
		NoiseScale:          0.08,
		PatternType:         PatternLayered,
		DetailLevel:         0.6,
		Contrast:            1.0,
		Saturation:          0.9,
		SecondaryNoiseScale: 0.2,
	},
	"sci-fi": {
		NoiseScale:          0.1,
		PatternType:         PatternGrid,
		DetailLevel:         0.8,
		Contrast:            1.2,
		Saturation:          0.7,
		SecondaryNoiseScale: 0.05,
	},
	"horror": {
		NoiseScale:          0.06,
		PatternType:         PatternVoronoi,
		DetailLevel:         0.5,
		Contrast:            1.5,
		Saturation:          0.3, // Desaturated
		SecondaryNoiseScale: 0.15,
	},
	"cyberpunk": {
		NoiseScale:          0.12,
		PatternType:         PatternGrid,
		DetailLevel:         0.9,
		Contrast:            1.4,
		Saturation:          1.0, // Full saturation for neon
		SecondaryNoiseScale: 0.08,
	},
	"post-apocalyptic": {
		NoiseScale:          0.07,
		PatternType:         PatternDistortion,
		DetailLevel:         0.4,
		Contrast:            0.9,
		Saturation:          0.5, // Faded
		SecondaryNoiseScale: 0.25,
	},
}

// GetGenrePatternConfig returns the pattern config for a genre.
func GetGenrePatternConfig(genre string) GenrePatternConfig {
	if cfg, ok := genrePatternConfigs[genre]; ok {
		return cfg
	}
	return genrePatternConfigs["fantasy"]
}

// GenerateGenreTexture creates a texture using genre-specific patterns.
func GenerateGenreTexture(width, height int, seed int64, genre string) *Texture {
	if width <= 0 || height <= 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(seed))
	pixels := make([]color.RGBA, width*height)
	palette := getGenrePalette(genre)
	config := GetGenrePatternConfig(genre)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixels[y*width+x] = generateGenrePixel(x, y, seed, rng, palette, config)
		}
	}

	return &Texture{Width: width, Height: height, Pixels: pixels}
}

// generateGenrePixel computes a pixel color using genre-specific patterns.
func generateGenrePixel(x, y int, seed int64, rng *rand.Rand, palette []color.RGBA, config GenrePatternConfig) color.RGBA {
	var noiseVal float64

	fx := float64(x)
	fy := float64(y)

	switch config.PatternType {
	case PatternGrid:
		noiseVal = gridPattern(fx, fy, seed, config)
	case PatternVoronoi:
		noiseVal = voronoiPattern(fx, fy, seed, config)
	case PatternDistortion:
		noiseVal = distortionPattern(fx, fy, seed, config)
	case PatternLayered:
		noiseVal = layeredPattern(fx, fy, seed, config)
	default:
		noiseVal = noise.Noise2D(fx*config.NoiseScale, fy*config.NoiseScale, seed)
	}

	// Apply contrast
	noiseVal = applyContrast(noiseVal, config.Contrast)

	// Map to palette
	paletteIdx := mapNoiseToIndex(noiseVal, len(palette))
	baseColor := palette[paletteIdx]

	// Apply saturation
	baseColor = applySaturation(baseColor, config.Saturation)

	// Apply detail variation
	if config.DetailLevel > 0 {
		detailNoise := noise.Noise2D(fx*0.5, fy*0.5, seed+1) * config.DetailLevel * 0.15
		baseColor = adjustBrightness(baseColor, detailNoise)
	}

	return applyColorVariation(baseColor, rng)
}

// gridPattern creates tech-style grid lines.
func gridPattern(x, y float64, seed int64, config GenrePatternConfig) float64 {
	// Base noise
	baseNoise := noise.Noise2D(x*config.NoiseScale, y*config.NoiseScale, seed)

	// Grid lines
	gridSize := 8.0
	xMod := math.Mod(x, gridSize)
	yMod := math.Mod(y, gridSize)

	// Thin lines at grid edges
	lineThickness := 0.5
	xLine := 0.0
	yLine := 0.0
	if xMod < lineThickness || xMod > gridSize-lineThickness {
		xLine = 0.3
	}
	if yMod < lineThickness || yMod > gridSize-lineThickness {
		yLine = 0.3
	}

	// Combine with base noise
	return clamp01(baseNoise*0.7 + math.Max(xLine, yLine))
}

// voronoiPattern creates organic cell patterns.
func voronoiPattern(x, y float64, seed int64, config GenrePatternConfig) float64 {
	// Use noise to create pseudo-Voronoi cells
	cellSize := 16.0
	cellX := math.Floor(x / cellSize)
	cellY := math.Floor(y / cellSize)

	minDist := 999.0

	// Check neighboring cells
	for dx := -1.0; dx <= 1.0; dx++ {
		for dy := -1.0; dy <= 1.0; dy++ {
			ncx := cellX + dx
			ncy := cellY + dy

			// Deterministic point in cell
			pointSeed := seed + int64(ncx*31+ncy*37)
			pointNoise := noise.Noise2D(ncx*0.1, ncy*0.1, pointSeed)
			px := ncx*cellSize + cellSize*0.5 + pointNoise*cellSize*0.3
			py := ncy*cellSize + cellSize*0.5 + noise.Noise2D(ncx*0.1+0.5, ncy*0.1+0.5, pointSeed)*cellSize*0.3

			dist := math.Sqrt((x-px)*(x-px) + (y-py)*(y-py))
			if dist < minDist {
				minDist = dist
			}
		}
	}

	// Normalize distance
	normalizedDist := minDist / cellSize
	return clamp01(normalizedDist * config.Contrast)
}

// distortionPattern creates wavy, unstable patterns.
func distortionPattern(x, y float64, seed int64, config GenrePatternConfig) float64 {
	// Apply distortion to coordinates
	distortX := noise.Noise2D(x*config.SecondaryNoiseScale, y*config.SecondaryNoiseScale, seed+100) * 10.0
	distortY := noise.Noise2D(x*config.SecondaryNoiseScale, y*config.SecondaryNoiseScale, seed+200) * 10.0

	// Sample noise at distorted position
	return noise.Noise2D((x+distortX)*config.NoiseScale, (y+distortY)*config.NoiseScale, seed)
}

// layeredPattern combines multiple noise octaves.
func layeredPattern(x, y float64, seed int64, config GenrePatternConfig) float64 {
	// Primary layer
	primary := noise.Noise2D(x*config.NoiseScale, y*config.NoiseScale, seed)

	// Secondary detail layer
	secondary := noise.Noise2D(x*config.SecondaryNoiseScale*4, y*config.SecondaryNoiseScale*4, seed+1) * 0.3

	// Tertiary fine detail
	tertiary := noise.Noise2D(x*0.4, y*0.4, seed+2) * 0.1

	return clamp01(primary*0.6 + secondary + tertiary)
}

// applyContrast adjusts the contrast of a noise value.
func applyContrast(val, contrast float64) float64 {
	// Center around 0.5, scale, then shift back
	centered := val - 0.5
	scaled := centered * contrast
	return clamp01(scaled + 0.5)
}

// applySaturation adjusts color saturation.
func applySaturation(c color.RGBA, saturation float64) color.RGBA {
	// Convert to grayscale
	gray := float64(c.R)*0.299 + float64(c.G)*0.587 + float64(c.B)*0.114

	// Interpolate between grayscale and original
	r := gray + (float64(c.R)-gray)*saturation
	g := gray + (float64(c.G)-gray)*saturation
	b := gray + (float64(c.B)-gray)*saturation

	return color.RGBA{
		R: clampColor(int(r)),
		G: clampColor(int(g)),
		B: clampColor(int(b)),
		A: c.A,
	}
}

// adjustBrightness shifts color brightness.
func adjustBrightness(c color.RGBA, delta float64) color.RGBA {
	shift := int(delta * 255)
	return color.RGBA{
		R: clampColor(int(c.R) + shift),
		G: clampColor(int(c.G) + shift),
		B: clampColor(int(c.B) + shift),
		A: c.A,
	}
}

// clamp01 clamps a value to [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
