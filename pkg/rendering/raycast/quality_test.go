//go:build noebiten

package raycast

import (
	"testing"
)

func TestQualityPresets(t *testing.T) {
	presets := []QualityPreset{QualityUltra, QualityHigh, QualityMedium, QualityLow, QualityMinimal}

	for _, preset := range presets {
		cfg := NewQualityConfig(preset)
		if cfg == nil {
			t.Errorf("NewQualityConfig(%d) returned nil", preset)
			continue
		}
		if cfg.Preset != preset {
			t.Errorf("NewQualityConfig(%d).Preset = %d", preset, cfg.Preset)
		}
	}
}

func TestQualityConfigDefaults(t *testing.T) {
	cfg := DefaultQualityConfig()
	if cfg == nil {
		t.Fatal("DefaultQualityConfig() returned nil")
	}
	if cfg.Preset != QualityHigh {
		t.Errorf("Default preset = %d, want QualityHigh (%d)", cfg.Preset, QualityHigh)
	}
}

func TestQualityPresetProgression(t *testing.T) {
	ultra := NewQualityConfig(QualityUltra)
	high := NewQualityConfig(QualityHigh)
	medium := NewQualityConfig(QualityMedium)
	low := NewQualityConfig(QualityLow)
	minimal := NewQualityConfig(QualityMinimal)

	// Max visible sprites should decrease with lower quality
	if ultra.MaxVisibleSprites <= high.MaxVisibleSprites {
		t.Error("Ultra should have more sprites than High")
	}
	if high.MaxVisibleSprites <= medium.MaxVisibleSprites {
		t.Error("High should have more sprites than Medium")
	}
	if medium.MaxVisibleSprites <= low.MaxVisibleSprites {
		t.Error("Medium should have more sprites than Low")
	}
	if low.MaxVisibleSprites <= minimal.MaxVisibleSprites {
		t.Error("Low should have more sprites than Minimal")
	}

	// Draw distance should decrease with lower quality
	if ultra.DrawDistance <= high.DrawDistance {
		t.Error("Ultra should have greater draw distance than High")
	}
	if high.DrawDistance <= medium.DrawDistance {
		t.Error("High should have greater draw distance than Medium")
	}
	if medium.DrawDistance <= low.DrawDistance {
		t.Error("Medium should have greater draw distance than Low")
	}
	if low.DrawDistance <= minimal.DrawDistance {
		t.Error("Low should have greater draw distance than Minimal")
	}
}

func TestQualityConfigFeatureDisabling(t *testing.T) {
	ultra := NewQualityConfig(QualityUltra)
	minimal := NewQualityConfig(QualityMinimal)

	// Ultra should have everything enabled
	if !ultra.NormalMapsEnabled {
		t.Error("Ultra should have normal maps enabled")
	}
	if !ultra.ShadowsEnabled {
		t.Error("Ultra should have shadows enabled")
	}
	if !ultra.ParticlesEnabled {
		t.Error("Ultra should have particles enabled")
	}
	if !ultra.PostProcessingEnabled {
		t.Error("Ultra should have post-processing enabled")
	}

	// Minimal should have most features disabled
	if minimal.NormalMapsEnabled {
		t.Error("Minimal should have normal maps disabled")
	}
	if minimal.ShadowsEnabled {
		t.Error("Minimal should have shadows disabled")
	}
	if minimal.ParticlesEnabled {
		t.Error("Minimal should have particles disabled")
	}
	if minimal.PostProcessingEnabled {
		t.Error("Minimal should have post-processing disabled")
	}
}

func TestQualityPresetName(t *testing.T) {
	tests := []struct {
		preset QualityPreset
		want   string
	}{
		{QualityUltra, "Ultra"},
		{QualityHigh, "High"},
		{QualityMedium, "Medium"},
		{QualityLow, "Low"},
		{QualityMinimal, "Minimal"},
		{QualityPreset(99), "Unknown"},
	}

	for _, tt := range tests {
		got := QualityPresetName(tt.preset)
		if got != tt.want {
			t.Errorf("QualityPresetName(%d) = %q, want %q", tt.preset, got, tt.want)
		}
	}
}

func TestQualityConfigShouldRenderSprite(t *testing.T) {
	cfg := NewQualityConfig(QualityMedium)

	// Within limits
	if !cfg.ShouldRenderSprite(0, 10.0) {
		t.Error("Should render sprite within limits")
	}

	// Beyond draw distance
	if cfg.ShouldRenderSprite(0, cfg.DrawDistance+10.0) {
		t.Error("Should not render sprite beyond draw distance")
	}

	// At max visible sprites
	if cfg.ShouldRenderSprite(cfg.MaxVisibleSprites, 10.0) {
		t.Error("Should not render sprite when at max visible count")
	}
}

func TestQualityConfigEffectiveSizes(t *testing.T) {
	cfg := NewQualityConfig(QualityLow)

	// Texture size should be reduced
	effectiveSize := cfg.EffectiveTextureSize(256)
	expectedSize := int(256 * cfg.TextureDetailLevel)
	if effectiveSize != expectedSize {
		t.Errorf("EffectiveTextureSize(256) = %d, want %d", effectiveSize, expectedSize)
	}

	// Sprite size should be reduced
	w, h := cfg.EffectiveSpriteSize(64, 96)
	expectedW := int(64 * cfg.SpriteDetailLevel)
	expectedH := int(96 * cfg.SpriteDetailLevel)
	if w != expectedW || h != expectedH {
		t.Errorf("EffectiveSpriteSize(64, 96) = (%d, %d), want (%d, %d)", w, h, expectedW, expectedH)
	}
}

func TestQualityConfigBarrierPattern(t *testing.T) {
	ultra := NewQualityConfig(QualityUltra)
	minimal := NewQualityConfig(QualityMinimal)

	// Count gaps in a 16x16 area for each quality
	ultraGaps := 0
	minimalGaps := 0
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if ultra.BarrierGapPattern(x, y) {
				ultraGaps++
			}
			if minimal.BarrierGapPattern(x, y) {
				minimalGaps++
			}
		}
	}

	// Ultra should have more gaps (more detail)
	if ultraGaps <= minimalGaps {
		t.Errorf("Ultra should have more gaps (%d) than Minimal (%d)", ultraGaps, minimalGaps)
	}
}

func TestQualityConfigApplyToLOD(t *testing.T) {
	cfg := NewQualityConfig(QualityLow)
	lod := DefaultLODConfig()

	originalHigh := lod.HighThreshold
	cfg.ApplyToLODConfig(lod)

	// LOD thresholds should be scaled
	expectedHigh := originalHigh * cfg.LODDistanceMultiplier
	if lod.HighThreshold != expectedHigh {
		t.Errorf("LOD HighThreshold = %f, want %f", lod.HighThreshold, expectedHigh)
	}

	// Normal maps disabled at low quality
	if !cfg.NormalMapsEnabled && lod.SkipNormalMapsAtLOD != LODFull {
		t.Error("LOD should skip normal maps at all distances when disabled")
	}
}

func TestRendererQualityConfig(t *testing.T) {
	r := NewRenderer(320, 240)

	// Default config
	cfg := r.GetQualityConfig()
	if cfg == nil {
		t.Fatal("GetQualityConfig should not return nil")
	}

	// Set custom config
	custom := NewQualityConfig(QualityLow)
	r.SetQualityConfig(custom)
	if r.GetQualityConfig() != custom {
		t.Error("SetQualityConfig should update the config")
	}

	// Set by preset
	r.SetQualityPreset(QualityMinimal)
	if r.GetQualityConfig().Preset != QualityMinimal {
		t.Error("SetQualityPreset should update to minimal")
	}
}

func TestQualityConfigInvalidPreset(t *testing.T) {
	// Invalid preset should return medium
	cfg := NewQualityConfig(QualityPreset(999))
	if cfg.Preset != QualityMedium {
		t.Errorf("Invalid preset should default to Medium, got %d", cfg.Preset)
	}
}

func BenchmarkNewQualityConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewQualityConfig(QualityMedium)
	}
}

func BenchmarkShouldRenderSprite(b *testing.B) {
	cfg := NewQualityConfig(QualityMedium)
	for i := 0; i < b.N; i++ {
		cfg.ShouldRenderSprite(50, 15.0)
	}
}
