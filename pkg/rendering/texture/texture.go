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

	// Get palette for genre
	palette, ok := GenrePalette[genre]
	if !ok {
		palette = GenrePalette["fantasy"]
	}

	// Generate noise-based texture
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Simple value noise
			noiseVal := noise.Noise2D(float64(x)*0.1, float64(y)*0.1, seed)

			// Map noise to palette index
			paletteIdx := int(noiseVal * float64(len(palette)))
			if paletteIdx >= len(palette) {
				paletteIdx = len(palette) - 1
			}
			if paletteIdx < 0 {
				paletteIdx = 0
			}

			baseColor := palette[paletteIdx]

			// Add subtle variation
			variation := int(rng.Intn(21)) - 10
			r := clampColor(int(baseColor.R) + variation)
			g := clampColor(int(baseColor.G) + variation)
			b := clampColor(int(baseColor.B) + variation)

			pixels[y*width+x] = color.RGBA{R: r, G: g, B: b, A: 255}
		}
	}

	return &Texture{Width: width, Height: height, Pixels: pixels}
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
