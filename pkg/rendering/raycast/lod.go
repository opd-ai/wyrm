// Package raycast provides the first-person raycasting renderer.
// This file implements the Level of Detail (LOD) system for performance optimization.
package raycast

// LODLevel represents a level of detail for rendering.
type LODLevel int

const (
	// LODFull renders at maximum quality (close objects)
	LODFull LODLevel = iota
	// LODHigh renders at slightly reduced quality (medium distance)
	LODHigh
	// LODMedium renders at reduced quality (far distance)
	LODMedium
	// LODLow renders at minimum quality (very far distance)
	LODLow
)

// LODConfig holds configuration for the LOD system.
type LODConfig struct {
	// Enabled controls whether LOD is active.
	Enabled bool
	// HighThreshold is the distance at which LOD switches from Full to High.
	HighThreshold float64
	// MediumThreshold is the distance at which LOD switches from High to Medium.
	MediumThreshold float64
	// LowThreshold is the distance at which LOD switches from Medium to Low.
	LowThreshold float64
	// SkipHighlightsAtLOD is the LOD level at which highlights are skipped.
	// Set to LODLow to always render highlights, or LODMedium to skip at medium+.
	SkipHighlightsAtLOD LODLevel
	// SimplifySpritesAtLOD is the LOD level at which sprites use simplified rendering.
	// At this level and below, sprite detail (animations, effects) is reduced.
	SimplifySpritesAtLOD LODLevel
	// SkipNormalMapsAtLOD is the LOD level at which normal map sampling is skipped.
	SkipNormalMapsAtLOD LODLevel
}

// DefaultLODConfig returns the default LOD configuration.
func DefaultLODConfig() *LODConfig {
	return &LODConfig{
		Enabled:              true,
		HighThreshold:        5.0,
		MediumThreshold:      10.0,
		LowThreshold:         20.0,
		SkipHighlightsAtLOD:  LODLow,
		SimplifySpritesAtLOD: LODMedium,
		SkipNormalMapsAtLOD:  LODMedium,
	}
}

// defaultLODConfig is a package-level default to avoid allocations.
var defaultLODConfig = DefaultLODConfig()

// GetLODLevel returns the appropriate LOD level for a given distance.
func (cfg *LODConfig) GetLODLevel(distance float64) LODLevel {
	if !cfg.Enabled || distance < cfg.HighThreshold {
		return LODFull
	}
	if distance < cfg.MediumThreshold {
		return LODHigh
	}
	if distance < cfg.LowThreshold {
		return LODMedium
	}
	return LODLow
}

// ShouldRenderHighlight returns true if highlights should be rendered at this LOD level.
func (cfg *LODConfig) ShouldRenderHighlight(lod LODLevel) bool {
	return lod < cfg.SkipHighlightsAtLOD
}

// ShouldSimplifySprite returns true if sprites should use simplified rendering at this LOD level.
func (cfg *LODConfig) ShouldSimplifySprite(lod LODLevel) bool {
	return lod >= cfg.SimplifySpritesAtLOD
}

// ShouldSkipNormalMaps returns true if normal map sampling should be skipped at this LOD level.
func (cfg *LODConfig) ShouldSkipNormalMaps(lod LODLevel) bool {
	return lod >= cfg.SkipNormalMapsAtLOD
}

// LODRenderFlags contains flags for what to render at the current LOD level.
type LODRenderFlags struct {
	// RenderHighlights controls whether highlight effects are rendered.
	RenderHighlights bool
	// RenderNormalMaps controls whether normal map lighting is calculated.
	RenderNormalMaps bool
	// SimplifiedSprites controls whether sprites use simplified rendering.
	SimplifiedSprites bool
	// Level is the current LOD level.
	Level LODLevel
}

// GetRenderFlags returns the render flags for a given distance.
func (cfg *LODConfig) GetRenderFlags(distance float64) LODRenderFlags {
	lod := cfg.GetLODLevel(distance)
	return LODRenderFlags{
		RenderHighlights:  cfg.ShouldRenderHighlight(lod),
		RenderNormalMaps:  !cfg.ShouldSkipNormalMaps(lod),
		SimplifiedSprites: cfg.ShouldSimplifySprite(lod),
		Level:             lod,
	}
}

// ColumnSkipInterval returns how many columns to skip for simplified sprite rendering.
// At higher LOD levels (lower quality), we skip more columns for faster rendering.
// This creates a "scan line" effect at distance but greatly reduces CPU cost.
func ColumnSkipForLOD(lod LODLevel) int {
	switch lod {
	case LODFull, LODHigh:
		return 1 // Render every column
	case LODMedium:
		return 2 // Render every other column
	case LODLow:
		return 4 // Render every 4th column
	default:
		return 1
	}
}

// RowSkipForLOD returns how many rows to skip for simplified sprite rendering.
func RowSkipForLOD(lod LODLevel) int {
	switch lod {
	case LODFull, LODHigh:
		return 1 // Render every row
	case LODMedium:
		return 2 // Render every other row
	case LODLow:
		return 3 // Render every 3rd row
	default:
		return 1
	}
}
