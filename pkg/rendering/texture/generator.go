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

// ============================================================
// Material-Based Texture Generation
// ============================================================

// GenerateForMaterial creates a texture based on a material's visual properties.
func GenerateForMaterial(width, height int, materialID MaterialID, seed int64, genre string) *Texture {
	return GenerateForMaterialWithRegistry(width, height, materialID, seed, genre, DefaultMaterialRegistry)
}

// GenerateForMaterialWithRegistry creates a texture with a custom registry.
func GenerateForMaterialWithRegistry(width, height int, materialID MaterialID, seed int64, genre string, registry *MaterialRegistry) *Texture {
	if width <= 0 || height <= 0 {
		return nil
	}
	if registry == nil {
		registry = DefaultMaterialRegistry
	}

	material := registry.Get(materialID)
	if material == nil {
		// Fall back to generic texture
		return GenerateWithSeed(width, height, seed, genre)
	}

	rng := rand.New(rand.NewSource(seed))
	pixels := make([]color.RGBA, width*height)

	// Get genre-specific colors or fall back to base colors
	colors := registry.GetColorsForGenre(materialID, genre)
	if len(colors) == 0 {
		colors = material.BaseColors
	}

	// Generate texture based on material category
	switch material.Category {
	case "metal":
		generateMetalTexture(pixels, width, height, seed, rng, colors, &material.Visual)
	case "organic":
		generateOrganicTexture(pixels, width, height, seed, rng, colors, &material.Visual)
	case "mineral":
		generateMineralTexture(pixels, width, height, seed, rng, colors, &material.Visual)
	case "natural":
		generateNaturalTexture(pixels, width, height, seed, rng, colors, &material.Visual)
	case "synthetic":
		generateSyntheticTexture(pixels, width, height, seed, rng, colors, &material.Visual)
	default:
		generateNoiseTexture(pixels, width, height, seed, rng, colors)
	}

	// Apply visual properties post-processing
	applyVisualProperties(pixels, width, height, &material.Visual)

	return &Texture{Width: width, Height: height, Pixels: pixels}
}

// generateMetalTexture creates a metallic texture with streaks and reflections.
func generateMetalTexture(pixels []color.RGBA, width, height int, seed int64, rng *rand.Rand, palette []color.RGBA, visual *VisualProperties) {
	// Metals have directional streaks (like brushed metal)
	streakDirection := rng.Float64() * 3.14159 // Random direction

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create directional noise for brushed metal effect
			fx := float64(x) * 0.05
			fy := float64(y) * 0.05

			// Rotate coordinates by streak direction
			rx := fx*cos(streakDirection) - fy*sin(streakDirection)

			// High-frequency directional noise
			noiseVal := noise.Noise2D(rx*2.0, float64(y)*0.02, seed)

			// Add some random sparkle based on roughness
			if visual.Roughness < 0.5 {
				sparkle := noise.Noise2D(float64(x)*0.5, float64(y)*0.5, seed+1)
				if sparkle > 0.95 {
					noiseVal = 1.0
				}
			}

			paletteIdx := mapNoiseToIndex(noiseVal, len(palette))
			baseColor := palette[paletteIdx]

			// Adjust brightness based on metalness and roughness
			brightness := 1.0 + (1.0-visual.Roughness)*0.3*noiseVal
			pixels[y*width+x] = scaleBrightness(baseColor, brightness)
		}
	}
}

// generateOrganicTexture creates textures for wood, flesh, plants.
func generateOrganicTexture(pixels []color.RGBA, width, height int, seed int64, rng *rand.Rand, palette []color.RGBA, visual *VisualProperties) {
	// Organic materials have grain/fiber patterns
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Multi-octave noise for organic look
			n1 := noise.Noise2D(float64(x)*0.1, float64(y)*0.02, seed)    // Long grain
			n2 := noise.Noise2D(float64(x)*0.05, float64(y)*0.05, seed+1) // Detail
			n3 := noise.Noise2D(float64(x)*0.2, float64(y)*0.2, seed+2)   // Fine detail

			noiseVal := n1*0.6 + n2*0.3 + n3*0.1

			paletteIdx := mapNoiseToIndex(noiseVal, len(palette))
			baseColor := palette[paletteIdx]
			pixels[y*width+x] = applyColorVariation(baseColor, rng)
		}
	}
}

// generateMineralTexture creates textures for stone, brick, crystal.
func generateMineralTexture(pixels []color.RGBA, width, height int, seed int64, rng *rand.Rand, palette []color.RGBA, visual *VisualProperties) {
	// Minerals have blocky, crystalline patterns
	blockSize := 8 + rng.Intn(8) // Variable block size

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Quantize to blocks
			bx := x / blockSize
			by := y / blockSize

			// Block-level noise
			blockNoise := noise.HashToFloat(bx, by, seed)

			// Edge highlighting
			edgeX := x % blockSize
			edgeY := y % blockSize
			isEdge := edgeX == 0 || edgeY == 0

			// Detail noise within block
			detailNoise := noise.Noise2D(float64(x)*0.1, float64(y)*0.1, seed+1)

			noiseVal := blockNoise*0.7 + detailNoise*0.3
			if isEdge {
				noiseVal *= 0.8 // Darken edges
			}

			paletteIdx := mapNoiseToIndex(noiseVal, len(palette))
			baseColor := palette[paletteIdx]
			pixels[y*width+x] = applyColorVariation(baseColor, rng)
		}
	}
}

// generateNaturalTexture creates textures for dirt, grass, water.
func generateNaturalTexture(pixels []color.RGBA, width, height int, seed int64, rng *rand.Rand, palette []color.RGBA, visual *VisualProperties) {
	// Natural materials have irregular, flowing patterns
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Multi-scale noise for natural variety
			n1 := noise.Noise2D(float64(x)*0.05, float64(y)*0.05, seed)
			n2 := noise.Noise2D(float64(x)*0.1, float64(y)*0.1, seed+1)
			n3 := noise.Noise2D(float64(x)*0.3, float64(y)*0.3, seed+2)

			noiseVal := n1*0.5 + n2*0.35 + n3*0.15

			paletteIdx := mapNoiseToIndex(noiseVal, len(palette))
			baseColor := palette[paletteIdx]
			pixels[y*width+x] = applyColorVariation(baseColor, rng)
		}
	}
}

// generateSyntheticTexture creates textures for plastic, neon, etc.
func generateSyntheticTexture(pixels []color.RGBA, width, height int, seed int64, rng *rand.Rand, palette []color.RGBA, visual *VisualProperties) {
	// Synthetic materials are smoother with less variation
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Smoother noise for synthetic look
			noiseVal := noise.Noise2D(float64(x)*0.03, float64(y)*0.03, seed)

			// Add emissive glow variation if applicable
			if visual.Emissive > 0.5 {
				glowNoise := noise.Noise2D(float64(x)*0.1, float64(y)*0.1, seed+1)
				noiseVal = noiseVal*0.5 + glowNoise*0.5
			}

			paletteIdx := mapNoiseToIndex(noiseVal, len(palette))
			baseColor := palette[paletteIdx]

			// Less variation for synthetic materials
			variation := int(rng.Intn(11)) - 5
			pixels[y*width+x] = color.RGBA{
				R: clampColor(int(baseColor.R) + variation),
				G: clampColor(int(baseColor.G) + variation),
				B: clampColor(int(baseColor.B) + variation),
				A: baseColor.A,
			}
		}
	}
}

// applyVisualProperties applies post-processing based on material visual properties.
func applyVisualProperties(pixels []color.RGBA, width, height int, visual *VisualProperties) {
	for i := range pixels {
		// Apply transparency
		if visual.Transparency > 0 {
			pixels[i].A = uint8(255 * (1 - visual.Transparency))
		}

		// Apply emissive boost (brighter colors)
		if visual.Emissive > 0.2 {
			factor := 1.0 + visual.Emissive*0.5
			pixels[i] = adjustBrightness(pixels[i], factor)
		}
	}
}

// scaleBrightness multiplies RGB values by a factor while preserving alpha.
func scaleBrightness(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: clampColor(int(float64(c.R) * factor)),
		G: clampColor(int(float64(c.G) * factor)),
		B: clampColor(int(float64(c.B) * factor)),
		A: c.A,
	}
}

// sin and cos wrappers for texture generation.
func sin(x float64) float64 {
	// Use math.Sin via simple approximation for texture gen
	// This keeps the noise package dependency minimal
	x = x - float64(int(x/(2*3.14159)))*2*3.14159
	if x > 3.14159 {
		x -= 2 * 3.14159
	}
	// Taylor series approximation
	x2 := x * x
	return x * (1 - x2/6 + x2*x2/120)
}

func cos(x float64) float64 {
	return sin(x + 3.14159/2)
}
