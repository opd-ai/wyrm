// Package raycast provides the first-person raycasting renderer.
// This file implements fallback rendering modes for low-end hardware.
package raycast

// QualityPreset represents a predefined rendering quality level.
type QualityPreset int

const (
	// QualityUltra uses maximum quality settings.
	QualityUltra QualityPreset = iota
	// QualityHigh uses high quality with slight optimizations.
	QualityHigh
	// QualityMedium balances quality and performance.
	QualityMedium
	// QualityLow prioritizes performance over quality.
	QualityLow
	// QualityMinimal uses minimum quality for very low-end hardware.
	QualityMinimal
)

// QualityConfig holds all rendering quality settings.
// Use NewQualityConfig or a preset to create a configuration.
type QualityConfig struct {
	// Preset is the base quality preset this config was derived from.
	Preset QualityPreset

	// NormalMapsEnabled controls whether normal map lighting is calculated.
	// Disabling saves significant CPU time.
	NormalMapsEnabled bool

	// ShadowsEnabled controls whether shadow casting is calculated.
	ShadowsEnabled bool

	// SpecularLightingEnabled controls whether specular highlights are rendered.
	SpecularLightingEnabled bool

	// TextureFilteringEnabled controls whether textures use bilinear filtering.
	// Disabling uses nearest-neighbor for a pixelated look.
	TextureFilteringEnabled bool

	// ParticlesEnabled controls whether particle effects are rendered.
	ParticlesEnabled bool

	// PostProcessingEnabled controls whether post-processing effects run.
	PostProcessingEnabled bool

	// WeatherEffectsEnabled controls whether weather overlays are rendered.
	WeatherEffectsEnabled bool

	// AnimatedTexturesEnabled controls whether textures animate.
	AnimatedTexturesEnabled bool

	// MaxVisibleSprites is the maximum number of sprites to render per frame.
	// Excess sprites are culled based on distance.
	MaxVisibleSprites int

	// TextureDetailLevel controls texture resolution multiplier (1.0 = full, 0.5 = half).
	TextureDetailLevel float64

	// SpriteDetailLevel controls sprite resolution multiplier.
	SpriteDetailLevel float64

	// DrawDistance is the maximum distance to render objects.
	DrawDistance float64

	// LODDistanceMultiplier scales the LOD thresholds.
	// Lower values switch to lower LOD earlier.
	LODDistanceMultiplier float64

	// BarrierDetailLevel controls fence/grate pattern complexity.
	// 1 = full detail, 2 = half pattern, 4 = quarter pattern.
	BarrierDetailLevel int

	// HighlightEffectsEnabled controls whether interaction highlights pulse/glow.
	HighlightEffectsEnabled bool

	// SkyboxDetailLevel controls sky rendering detail.
	// 0 = solid color, 1 = gradient, 2 = full procedural.
	SkyboxDetailLevel int

	// ReflectionsEnabled controls whether reflective surfaces show reflections.
	ReflectionsEnabled bool
}

// NewQualityConfig creates a quality config from a preset.
func NewQualityConfig(preset QualityPreset) *QualityConfig {
	switch preset {
	case QualityUltra:
		return &QualityConfig{
			Preset:                  QualityUltra,
			NormalMapsEnabled:       true,
			ShadowsEnabled:          true,
			SpecularLightingEnabled: true,
			TextureFilteringEnabled: true,
			ParticlesEnabled:        true,
			PostProcessingEnabled:   true,
			WeatherEffectsEnabled:   true,
			AnimatedTexturesEnabled: true,
			MaxVisibleSprites:       512,
			TextureDetailLevel:      1.0,
			SpriteDetailLevel:       1.0,
			DrawDistance:            50.0,
			LODDistanceMultiplier:   1.5,
			BarrierDetailLevel:      1,
			HighlightEffectsEnabled: true,
			SkyboxDetailLevel:       2,
			ReflectionsEnabled:      true,
		}
	case QualityHigh:
		return &QualityConfig{
			Preset:                  QualityHigh,
			NormalMapsEnabled:       true,
			ShadowsEnabled:          true,
			SpecularLightingEnabled: true,
			TextureFilteringEnabled: true,
			ParticlesEnabled:        true,
			PostProcessingEnabled:   true,
			WeatherEffectsEnabled:   true,
			AnimatedTexturesEnabled: true,
			MaxVisibleSprites:       256,
			TextureDetailLevel:      1.0,
			SpriteDetailLevel:       1.0,
			DrawDistance:            40.0,
			LODDistanceMultiplier:   1.0,
			BarrierDetailLevel:      1,
			HighlightEffectsEnabled: true,
			SkyboxDetailLevel:       2,
			ReflectionsEnabled:      true,
		}
	case QualityMedium:
		return &QualityConfig{
			Preset:                  QualityMedium,
			NormalMapsEnabled:       false, // Disable for performance
			ShadowsEnabled:          true,
			SpecularLightingEnabled: false,
			TextureFilteringEnabled: true,
			ParticlesEnabled:        true,
			PostProcessingEnabled:   false, // Disable for performance
			WeatherEffectsEnabled:   true,
			AnimatedTexturesEnabled: true,
			MaxVisibleSprites:       128,
			TextureDetailLevel:      0.75,
			SpriteDetailLevel:       1.0,
			DrawDistance:            30.0,
			LODDistanceMultiplier:   0.8,
			BarrierDetailLevel:      2,
			HighlightEffectsEnabled: true,
			SkyboxDetailLevel:       1,
			ReflectionsEnabled:      false,
		}
	case QualityLow:
		return &QualityConfig{
			Preset:                  QualityLow,
			NormalMapsEnabled:       false,
			ShadowsEnabled:          false,
			SpecularLightingEnabled: false,
			TextureFilteringEnabled: false,
			ParticlesEnabled:        false,
			PostProcessingEnabled:   false,
			WeatherEffectsEnabled:   false,
			AnimatedTexturesEnabled: false,
			MaxVisibleSprites:       64,
			TextureDetailLevel:      0.5,
			SpriteDetailLevel:       0.75,
			DrawDistance:            20.0,
			LODDistanceMultiplier:   0.5,
			BarrierDetailLevel:      4,
			HighlightEffectsEnabled: true,
			SkyboxDetailLevel:       1,
			ReflectionsEnabled:      false,
		}
	case QualityMinimal:
		return &QualityConfig{
			Preset:                  QualityMinimal,
			NormalMapsEnabled:       false,
			ShadowsEnabled:          false,
			SpecularLightingEnabled: false,
			TextureFilteringEnabled: false,
			ParticlesEnabled:        false,
			PostProcessingEnabled:   false,
			WeatherEffectsEnabled:   false,
			AnimatedTexturesEnabled: false,
			MaxVisibleSprites:       32,
			TextureDetailLevel:      0.25,
			SpriteDetailLevel:       0.5,
			DrawDistance:            15.0,
			LODDistanceMultiplier:   0.25,
			BarrierDetailLevel:      8,
			HighlightEffectsEnabled: false,
			SkyboxDetailLevel:       0,
			ReflectionsEnabled:      false,
		}
	default:
		return NewQualityConfig(QualityMedium)
	}
}

// DefaultQualityConfig returns a sensible default quality configuration.
func DefaultQualityConfig() *QualityConfig {
	return NewQualityConfig(QualityHigh)
}

// QualityPresetName returns a human-readable name for a preset.
func QualityPresetName(preset QualityPreset) string {
	switch preset {
	case QualityUltra:
		return "Ultra"
	case QualityHigh:
		return "High"
	case QualityMedium:
		return "Medium"
	case QualityLow:
		return "Low"
	case QualityMinimal:
		return "Minimal"
	default:
		return "Unknown"
	}
}

// ApplyToLODConfig updates an LOD config based on the quality settings.
func (q *QualityConfig) ApplyToLODConfig(lod *LODConfig) {
	if lod == nil {
		return
	}
	// Scale LOD thresholds
	lod.HighThreshold *= q.LODDistanceMultiplier
	lod.MediumThreshold *= q.LODDistanceMultiplier
	lod.LowThreshold *= q.LODDistanceMultiplier

	// Disable features at lower quality
	if !q.NormalMapsEnabled {
		lod.SkipNormalMapsAtLOD = LODFull // Skip at all distances
	}
	if !q.HighlightEffectsEnabled {
		lod.SkipHighlightsAtLOD = LODFull
	}
}

// ShouldRenderSprite returns true if a sprite should be rendered given
// the current visible count and quality settings.
func (q *QualityConfig) ShouldRenderSprite(visibleCount int, distance float64) bool {
	if distance > q.DrawDistance {
		return false
	}
	if visibleCount >= q.MaxVisibleSprites {
		return false
	}
	return true
}

// EffectiveTextureSize returns the texture size to use given quality settings.
func (q *QualityConfig) EffectiveTextureSize(baseSize int) int {
	return int(float64(baseSize) * q.TextureDetailLevel)
}

// EffectiveSpriteSize returns the sprite size to use given quality settings.
func (q *QualityConfig) EffectiveSpriteSize(baseWidth, baseHeight int) (int, int) {
	return int(float64(baseWidth) * q.SpriteDetailLevel),
		int(float64(baseHeight) * q.SpriteDetailLevel)
}

// BarrierGapPattern returns whether a barrier position should be a gap
// at the current quality level.
func (q *QualityConfig) BarrierGapPattern(x, y int) bool {
	detail := q.BarrierDetailLevel
	if detail <= 1 {
		// Full detail - use normal pattern
		return (x+y)%4 == 0
	}
	// Simplified pattern - fewer gaps at lower quality
	// Higher detail level = fewer gaps (1/detail^2 as many)
	return (x/detail+y/detail)%(4) == 0 && (x%detail == 0) && (y%detail == 0)
}
