// Package texture provides procedural texture generation.
package texture

import (
	"image/color"
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/procgen/noise"
)

// Texture represents a procedurally generated texture.
type Texture struct {
	Width  int
	Height int
	Pixels []color.RGBA
}

// GenrePalette holds color palettes for different genres.
var GenrePalette = map[string][]color.RGBA{
	"fantasy": {
		{R: 139, G: 119, B: 101, A: 255}, // warm brown
		{R: 169, G: 149, B: 121, A: 255}, // light tan
		{R: 101, G: 139, B: 101, A: 255}, // forest green
		{R: 199, G: 179, B: 139, A: 255}, // gold tint
	},
	"sci-fi": {
		{R: 80, G: 100, B: 130, A: 255},  // steel blue
		{R: 200, G: 210, B: 220, A: 255}, // bright white
		{R: 60, G: 80, B: 100, A: 255},   // dark blue
		{R: 100, G: 120, B: 150, A: 255}, // medium blue
	},
	"horror": {
		{R: 60, G: 55, B: 50, A: 255},  // dark grey
		{R: 80, G: 70, B: 65, A: 255},  // muted brown
		{R: 100, G: 85, B: 75, A: 255}, // dusty tan
		{R: 50, G: 45, B: 45, A: 255},  // near black
	},
	"cyberpunk": {
		{R: 40, G: 30, B: 50, A: 255},  // dark purple
		{R: 255, G: 0, B: 128, A: 255}, // neon pink
		{R: 0, G: 255, B: 255, A: 255}, // cyan
		{R: 80, G: 60, B: 100, A: 255}, // muted purple
	},
	"post-apocalyptic": {
		{R: 139, G: 119, B: 91, A: 255},  // rust orange
		{R: 169, G: 149, B: 111, A: 255}, // dusty tan
		{R: 99, G: 89, B: 79, A: 255},    // weathered grey
		{R: 119, G: 99, B: 79, A: 255},   // mud brown
	},
}

// Generate creates a procedural texture of the given size.
// Returns nil if width or height is <= 0.
func Generate(width, height int) *Texture {
	return GenerateWithSeed(width, height, 0, "fantasy")
}

// GenerateWithSeed creates a procedural texture using the given seed and genre.
func GenerateWithSeed(width, height int, seed int64, genre string) *Texture {
	if width <= 0 || height <= 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(seed))
	pixels := make([]color.RGBA, width*height)
	palette := getGenrePalette(genre)

	generateNoiseTexture(pixels, width, height, seed, rng, palette)

	return &Texture{Width: width, Height: height, Pixels: pixels}
}

// getGenrePalette returns the color palette for the given genre.
func getGenrePalette(genre string) []color.RGBA {
	palette, ok := GenrePalette[genre]
	if !ok {
		palette = GenrePalette["fantasy"]
	}
	return palette
}

// generateNoiseTexture fills the pixel array with noise-based colors.
func generateNoiseTexture(pixels []color.RGBA, width, height int, seed int64, rng *rand.Rand, palette []color.RGBA) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixels[y*width+x] = generatePixelColor(x, y, seed, rng, palette)
		}
	}
}

// generatePixelColor computes the color for a single pixel using noise.
func generatePixelColor(x, y int, seed int64, rng *rand.Rand, palette []color.RGBA) color.RGBA {
	noiseVal := noise.Noise2D(float64(x)*0.1, float64(y)*0.1, seed)
	paletteIdx := mapNoiseToIndex(noiseVal, len(palette))
	baseColor := palette[paletteIdx]
	return applyColorVariation(baseColor, rng)
}

// mapNoiseToIndex maps a noise value [0,1] to a palette index.
func mapNoiseToIndex(noiseVal float64, paletteLen int) int {
	idx := int(noiseVal * float64(paletteLen))
	if idx >= paletteLen {
		idx = paletteLen - 1
	}
	if idx < 0 {
		idx = 0
	}
	return idx
}

// applyColorVariation adds subtle random variation to a base color.
func applyColorVariation(baseColor color.RGBA, rng *rand.Rand) color.RGBA {
	variation := int(rng.Intn(21)) - 10
	return color.RGBA{
		R: clampColor(int(baseColor.R) + variation),
		G: clampColor(int(baseColor.G) + variation),
		B: clampColor(int(baseColor.B) + variation),
		A: 255,
	}
}

// clampColor clamps a color value to [0, 255].
func clampColor(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
