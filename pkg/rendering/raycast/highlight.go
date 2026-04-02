// Package raycast provides the first-person raycasting renderer.
package raycast

import (
	"image/color"
	"math"
)

// EdgeHighlightConfig configures edge-based highlight rendering for interactive objects.
// This is separate from HighlightConfig in billboard.go which handles outline rendering.
type EdgeHighlightConfig struct {
	// Enabled controls whether edge highlighting is active.
	Enabled bool
	// Genre determines the accent color used for highlights.
	Genre string
	// Time is the current game time for animation.
	Time float64
	// PulseSpeed controls how fast the highlight pulses (default 3.0).
	PulseSpeed float64
	// PulseAmplitude controls the pulse intensity range (default 0.3).
	PulseAmplitude float64
	// BaseIntensity is the minimum intensity (default 0.7).
	BaseIntensity float64
}

// DefaultEdgeHighlightConfig returns an edge highlight config with default values.
func DefaultEdgeHighlightConfig() EdgeHighlightConfig {
	return EdgeHighlightConfig{
		Enabled:        true,
		Genre:          "fantasy",
		Time:           0,
		PulseSpeed:     3.0,
		PulseAmplitude: 0.3,
		BaseIntensity:  0.7,
	}
}

// GenreAccentColors maps genres to their accent colors for highlighting.
var GenreAccentColors = map[string]color.RGBA{
	"fantasy":          {R: 255, G: 215, B: 0, A: 255},   // Gold
	"sci-fi":           {R: 0, G: 255, B: 255, A: 255},   // Cyan
	"horror":           {R: 220, G: 20, B: 60, A: 255},   // Crimson red
	"cyberpunk":        {R: 255, G: 20, B: 147, A: 255},  // Neon pink
	"post-apocalyptic": {R: 255, G: 140, B: 0, A: 255},   // Orange
	"post-apoc":        {R: 255, G: 140, B: 0, A: 255},   // Orange (alias)
}

// GetAccentColor returns the accent color for a given genre.
func GetAccentColor(genre string) color.RGBA {
	if c, ok := GenreAccentColors[genre]; ok {
		return c
	}
	// Default to gold if genre not found
	return GenreAccentColors["fantasy"]
}

// HighlightRegion defines a screen region where a sprite was rendered.
type HighlightRegion struct {
	// MinX is the left edge of the region.
	MinX int
	// MaxX is the right edge of the region.
	MaxX int
	// MinY is the top edge of the region.
	MinY int
	// MaxY is the bottom edge of the region.
	MaxY int
	// EntityID identifies the object for this region.
	EntityID uint64
	// HighlightState indicates the highlight level (0=none, 1=in range, 2=targeted).
	HighlightState int
}

// EdgeHighlightRenderer handles edge-detection and highlight rendering for sprites.
type EdgeHighlightRenderer struct {
	config      EdgeHighlightConfig
	width       int
	height      int
	framebuffer []byte // RGBA framebuffer reference
}

// NewEdgeHighlightRenderer creates a highlight renderer for a given framebuffer size.
func NewEdgeHighlightRenderer(width, height int, config EdgeHighlightConfig) *EdgeHighlightRenderer {
	return &EdgeHighlightRenderer{
		config: config,
		width:  width,
		height: height,
	}
}

// SetFramebuffer sets the framebuffer reference for highlight operations.
func (hr *EdgeHighlightRenderer) SetFramebuffer(fb []byte) {
	hr.framebuffer = fb
}

// SetTime updates the animation time for pulsing.
func (hr *EdgeHighlightRenderer) SetTime(t float64) {
	hr.config.Time = t
}

// SetGenre updates the genre for accent color selection.
func (hr *EdgeHighlightRenderer) SetGenre(genre string) {
	hr.config.Genre = genre
}

// CalculatePulseIntensity calculates the current pulse intensity.
func (hr *EdgeHighlightRenderer) CalculatePulseIntensity() float64 {
	// sin(time * speed) * amplitude + baseIntensity
	// Results in range [baseIntensity - amplitude, baseIntensity + amplitude]
	return math.Sin(hr.config.Time*hr.config.PulseSpeed)*hr.config.PulseAmplitude + hr.config.BaseIntensity
}

// ApplyHighlight applies edge highlighting to a sprite region in the framebuffer.
// The region defines the screen area where the sprite was rendered.
func (hr *EdgeHighlightRenderer) ApplyHighlight(region HighlightRegion) {
	if !hr.config.Enabled || hr.framebuffer == nil {
		return
	}
	if region.HighlightState <= 0 {
		return
	}

	// Clamp region to framebuffer bounds
	minX := max(0, region.MinX)
	maxX := min(hr.width-1, region.MaxX)
	minY := max(0, region.MinY)
	maxY := min(hr.height-1, region.MaxY)

	if minX >= maxX || minY >= maxY {
		return
	}

	// Get accent color and pulse intensity
	accentColor := GetAccentColor(hr.config.Genre)
	intensity := hr.CalculatePulseIntensity()

	// Scale intensity based on highlight state (targeted = brighter)
	if region.HighlightState == 2 {
		intensity = math.Min(1.0, intensity*1.3)
	}

	// Perform edge detection and apply highlight
	hr.applyEdgeHighlight(minX, maxX, minY, maxY, accentColor, intensity)
}

// applyEdgeHighlight detects edges and applies highlight color.
func (hr *EdgeHighlightRenderer) applyEdgeHighlight(minX, maxX, minY, maxY int, accentColor color.RGBA, intensity float64) {
	// For each pixel in the region, check if it's on the edge of a sprite
	// An edge pixel has alpha > 0 and at least one neighbor with alpha == 0

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			idx := (y*hr.width + x) * 4
			if idx+3 >= len(hr.framebuffer) {
				continue
			}

			// Check if current pixel has content (alpha > 0)
			if hr.framebuffer[idx+3] == 0 {
				continue
			}

			// Check if this is an edge pixel (has transparent neighbor)
			if hr.isEdgePixel(x, y, minX, maxX, minY, maxY) {
				hr.blendHighlightColor(idx, accentColor, intensity)
			}
		}
	}
}

// isEdgePixel checks if a pixel has a transparent neighbor within the region.
func (hr *EdgeHighlightRenderer) isEdgePixel(x, y, minX, maxX, minY, maxY int) bool {
	// Check 4-connected neighbors
	neighbors := [][2]int{
		{x - 1, y}, // left
		{x + 1, y}, // right
		{x, y - 1}, // up
		{x, y + 1}, // down
	}

	for _, n := range neighbors {
		nx, ny := n[0], n[1]

		// Neighbor outside sprite region counts as transparent (edge of region)
		if nx < minX || nx > maxX || ny < minY || ny > maxY {
			return true
		}

		// Neighbor outside framebuffer counts as transparent
		if nx < 0 || nx >= hr.width || ny < 0 || ny >= hr.height {
			return true
		}

		// Check neighbor's alpha
		nidx := (ny*hr.width + nx) * 4
		if nidx+3 < len(hr.framebuffer) && hr.framebuffer[nidx+3] == 0 {
			return true
		}
	}

	return false
}

// blendHighlightColor blends the highlight color into the framebuffer at the given index.
func (hr *EdgeHighlightRenderer) blendHighlightColor(idx int, accentColor color.RGBA, intensity float64) {
	// Additive blending with intensity
	r := float64(hr.framebuffer[idx])
	g := float64(hr.framebuffer[idx+1])
	b := float64(hr.framebuffer[idx+2])

	// Add highlight color scaled by intensity
	r = math.Min(255, r+float64(accentColor.R)*intensity)
	g = math.Min(255, g+float64(accentColor.G)*intensity)
	b = math.Min(255, b+float64(accentColor.B)*intensity)

	hr.framebuffer[idx] = uint8(r)
	hr.framebuffer[idx+1] = uint8(g)
	hr.framebuffer[idx+2] = uint8(b)
}

// ApplyHighlightToRegions applies highlights to multiple regions.
func (hr *EdgeHighlightRenderer) ApplyHighlightToRegions(regions []HighlightRegion) {
	for _, region := range regions {
		hr.ApplyHighlight(region)
	}
}

// CreateRegionFromScreenBounds creates a highlight region from screen coordinates.
func CreateRegionFromScreenBounds(entityID uint64, screenX, screenY, width, height, highlightState int) HighlightRegion {
	return HighlightRegion{
		MinX:           screenX,
		MaxX:           screenX + width - 1,
		MinY:           screenY,
		MaxY:           screenY + height - 1,
		EntityID:       entityID,
		HighlightState: highlightState,
	}
}
