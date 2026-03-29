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
// Colors are based on ROADMAP.md genre visual specifications.
var GenrePalette = map[string][]color.RGBA{
	"fantasy": {
		{R: 0xD4, G: 0xA5, B: 0x74, A: 255}, // warm gold
		{R: 0x4A, G: 0x7C, B: 0x23, A: 255}, // green
		{R: 0x8B, G: 0x45, B: 0x13, A: 255}, // brown
		{R: 0xC0, G: 0xA0, B: 0x60, A: 255}, // light gold
	},
	"sci-fi": {
		{R: 0x1E, G: 0x90, B: 0xFF, A: 255}, // cool blue
		{R: 0xF0, G: 0xF0, B: 0xF0, A: 255}, // white
		{R: 0xC0, G: 0xC0, B: 0xC0, A: 255}, // chrome
		{R: 0x40, G: 0x60, B: 0x90, A: 255}, // steel blue
	},
	"horror": {
		{R: 0x55, G: 0x6B, B: 0x2F, A: 255}, // desaturated grey-green
		{R: 0x1A, G: 0x1A, B: 0x1A, A: 255}, // near black
		{R: 0x8B, G: 0x00, B: 0x00, A: 255}, // blood red
		{R: 0x3F, G: 0x3F, B: 0x3F, A: 255}, // dark grey
	},
	"cyberpunk": {
		{R: 0xFF, G: 0x00, B: 0xFF, A: 255}, // neon pink
		{R: 0x00, G: 0xFF, B: 0xFF, A: 255}, // cyan
		{R: 0x2F, G: 0x2F, B: 0x2F, A: 255}, // dark grey
		{R: 0x80, G: 0x00, B: 0x80, A: 255}, // deep purple
	},
	"post-apocalyptic": {
		{R: 0x70, G: 0x42, B: 0x14, A: 255}, // sepia
		{R: 0xCC, G: 0x77, B: 0x22, A: 255}, // orange dust
		{R: 0xB7, G: 0x41, B: 0x0E, A: 255}, // rust
		{R: 0x8B, G: 0x63, B: 0x33, A: 255}, // weathered tan
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
