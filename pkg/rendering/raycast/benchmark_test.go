//go:build noebiten

// Package raycast provides first-person raycasting rendering.
// This file contains additional benchmarks for hot path performance.
// Note: Some benchmarks exist in other test files (raycast_test.go, billboard_test.go, etc.)
package raycast

import (
	"image/color"
	"testing"
)

// Benchmarks for LOD calculations
func BenchmarkLODGetLevel(b *testing.B) {
	cfg := DefaultLODConfig()
	distances := []float64{2.0, 7.0, 12.0, 25.0, 50.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d := range distances {
			cfg.GetLODLevel(d)
		}
	}
}

func BenchmarkLODRenderFlags(b *testing.B) {
	cfg := DefaultLODConfig()
	distances := []float64{2.0, 7.0, 12.0, 25.0, 50.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d := range distances {
			cfg.GetRenderFlags(d)
		}
	}
}

func BenchmarkLODColumnSkip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ColumnSkipForLOD(LODFull)
		ColumnSkipForLOD(LODHigh)
		ColumnSkipForLOD(LODMedium)
		ColumnSkipForLOD(LODLow)
	}
}

// Benchmarks for quality config
func BenchmarkQualityConfigCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewQualityConfig(QualityMedium)
	}
}

func BenchmarkQualityConfigShouldRenderSprite(b *testing.B) {
	cfg := NewQualityConfig(QualityMedium)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.ShouldRenderSprite(50, 15.0)
	}
}

func BenchmarkQualityApplyToLOD(b *testing.B) {
	cfg := NewQualityConfig(QualityMedium)
	lod := DefaultLODConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.ApplyToLODConfig(lod)
	}
}

// Benchmarks for accessibility color operations
func BenchmarkAccessibleHighlightColor(b *testing.B) {
	cfg := DefaultAccessibilityConfig()
	palette := DefaultAccessiblePalette()
	types := []string{"weapon", "armor", "consumable", "quest", "hostile"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, t := range types {
			GetAccessibleHighlightColor(t, cfg, palette)
		}
	}
}

func BenchmarkApplyBrightnessColors(b *testing.B) {
	colors := []color.RGBA{
		{100, 150, 200, 255},
		{50, 100, 150, 255},
		{200, 200, 200, 255},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range colors {
			ApplyBrightness(c, 1.5)
		}
	}
}

func BenchmarkApplyHighContrastColors(b *testing.B) {
	colors := []color.RGBA{
		{100, 150, 200, 255},
		{50, 50, 50, 255},
		{200, 200, 200, 255},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range colors {
			ApplyHighContrast(c, HighContrastYellowBlack)
		}
	}
}

// Benchmarks for spatial hash with varying sizes
func BenchmarkSpatialHashQuery100(b *testing.B) {
	hash := NewSpatialHash(5.0)
	for i := 0; i < 100; i++ {
		hash.Insert(&SpriteEntity{
			X: float64(i%10) * 10,
			Y: float64(i/10) * 10,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.QueryRadius(50.0, 50.0, 10.0)
	}
}

func BenchmarkSpatialHashQuery1000(b *testing.B) {
	hash := NewSpatialHash(5.0)
	for i := 0; i < 1000; i++ {
		hash.Insert(&SpriteEntity{
			X: float64(i % 100),
			Y: float64(i / 100),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.QueryRadius(50.0, 50.0, 10.0)
	}
}

func BenchmarkSpatialHashQueryRect(b *testing.B) {
	hash := NewSpatialHash(5.0)
	for i := 0; i < 1000; i++ {
		hash.Insert(&SpriteEntity{
			X: float64(i % 100),
			Y: float64(i / 100),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.QueryRect(40.0, 40.0, 60.0, 60.0)
	}
}

func BenchmarkSpatialHashNearestInteractable(b *testing.B) {
	hash := NewSpatialHash(5.0)
	for i := 0; i < 100; i++ {
		hash.Insert(&SpriteEntity{
			X:              float64(i%10) * 10,
			Y:              float64(i/10) * 10,
			IsInteractable: i%3 == 0,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.NearestInteractable(50.0, 50.0, 15.0)
	}
}

// Benchmarks for highlight system
func BenchmarkHighlightColorForState(b *testing.B) {
	cfg := DefaultHighlightConfig()
	states := []int{HighlightNone, HighlightInRange, HighlightTargeted}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range states {
			cfg.HighlightColorForState(s)
		}
	}
}

func BenchmarkUpdateHighlightPulse(b *testing.B) {
	r := NewRenderer(640, 480)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.UpdateHighlightPulse(0.016)
	}
}

// Benchmarks for targeting
func BenchmarkFindTargetedEntity10(b *testing.B) {
	benchmarkFindTargetedEntityHelper(b, 10)
}

func BenchmarkFindTargetedEntity50(b *testing.B) {
	benchmarkFindTargetedEntityHelper(b, 50)
}

func BenchmarkFindTargetedEntity100(b *testing.B) {
	benchmarkFindTargetedEntityHelper(b, 100)
}

func benchmarkFindTargetedEntityHelper(b *testing.B, count int) {
	r := NewRenderer(640, 480)
	r.PlayerX = 50.0
	r.PlayerY = 50.0
	r.PlayerA = 0.0
	r.PlayerZ = 0.5

	sprites := make([]*SpriteEntity, count)
	for i := range sprites {
		sprites[i] = &SpriteEntity{
			X:                float64(i%10)*5 + 45,
			Y:                float64(i/10)*5 + 45,
			Visible:          true,
			IsInteractable:   i%3 == 0,
			InteractionRange: 5.0,
		}
	}

	cfg := DefaultTargetingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.FindTargetedEntity(sprites, cfg)
	}
}

// Benchmarks for context pool
func BenchmarkResetContextPool(b *testing.B) {
	r := NewRenderer(640, 480)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ResetContextPool()
	}
}

// Benchmarks for config getters
func BenchmarkGetLODConfig(b *testing.B) {
	r := NewRenderer(640, 480)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetLODConfig()
	}
}

func BenchmarkGetQualityConfig(b *testing.B) {
	r := NewRenderer(640, 480)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetQualityConfig()
	}
}

func BenchmarkGetAccessibilityConfig(b *testing.B) {
	r := NewRenderer(640, 480)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetAccessibilityConfig()
	}
}

// Combined workflow benchmarks
func BenchmarkTypicalFrameWorkflow(b *testing.B) {
	r := NewRenderer(640, 480)
	r.PlayerX = 50.0
	r.PlayerY = 50.0
	r.PlayerA = 0.0
	r.PlayerZ = 0.5

	sprites := make([]*SpriteEntity, 50)
	for i := range sprites {
		sprites[i] = &SpriteEntity{
			X:                float64(i%10)*5 + 45,
			Y:                float64(i/10)*5 + 45,
			Visible:          true,
			IsInteractable:   i%5 == 0,
			InteractionRange: 5.0,
		}
	}

	cfg := DefaultTargetingConfig()
	lodCfg := r.GetLODConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Frame start
		r.ResetContextPool()
		r.UpdateHighlightPulse(0.016)

		// Transform and cull sprites
		for _, s := range sprites {
			r.TransformEntityToScreen(s)
			lodCfg.GetLODLevel(s.Distance)
		}

		// Find target
		r.FindTargetedEntity(sprites, cfg)
	}
}

func BenchmarkHighQualityWorkflow(b *testing.B) {
	r := NewRenderer(1920, 1080)
	r.SetQualityPreset(QualityUltra)
	r.PlayerX = 50.0
	r.PlayerY = 50.0
	r.PlayerA = 0.0
	r.PlayerZ = 0.5

	sprites := make([]*SpriteEntity, 100)
	for i := range sprites {
		sprites[i] = &SpriteEntity{
			X:                float64(i%10) * 10,
			Y:                float64(i/10) * 10,
			Visible:          true,
			IsInteractable:   i%5 == 0,
			InteractionRange: 10.0,
		}
	}

	cfg := DefaultTargetingConfig()
	qualCfg := r.GetQualityConfig()
	lodCfg := r.GetLODConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ResetContextPool()
		r.UpdateHighlightPulse(0.016)

		visibleCount := 0
		for _, s := range sprites {
			r.TransformEntityToScreen(s)
			if qualCfg.ShouldRenderSprite(visibleCount, s.Distance) {
				lodCfg.GetRenderFlags(s.Distance)
				visibleCount++
			}
		}

		r.FindTargetedEntity(sprites, cfg)
	}
}

func BenchmarkLowQualityWorkflow(b *testing.B) {
	r := NewRenderer(640, 480)
	r.SetQualityPreset(QualityMinimal)
	r.PlayerX = 50.0
	r.PlayerY = 50.0
	r.PlayerA = 0.0
	r.PlayerZ = 0.5

	sprites := make([]*SpriteEntity, 30)
	for i := range sprites {
		sprites[i] = &SpriteEntity{
			X:                float64(i%5)*5 + 45,
			Y:                float64(i/5)*5 + 45,
			Visible:          true,
			IsInteractable:   i%5 == 0,
			InteractionRange: 3.0,
		}
	}

	cfg := DefaultTargetingConfig()
	qualCfg := r.GetQualityConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ResetContextPool()
		r.UpdateHighlightPulse(0.016)

		visibleCount := 0
		for _, s := range sprites {
			r.TransformEntityToScreen(s)
			if qualCfg.ShouldRenderSprite(visibleCount, s.Distance) {
				visibleCount++
			}
		}

		r.FindTargetedEntity(sprites, cfg)
	}
}
