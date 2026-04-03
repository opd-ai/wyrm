//go:build noebiten

// Package raycast provides first-person raycasting rendering.
// This file contains additional benchmarks for hot path performance.
// Note: Some benchmarks exist in other test files (raycast_test.go, billboard_test.go, etc.)
package raycast

import (
	"image/color"
	"testing"

	"github.com/opd-ai/wyrm/pkg/rendering/texture"
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

// === PLAN.md Required Benchmarks ===
// These benchmarks measure core rendering operations with target performance goals.
// NOTE: Some rendering methods require Ebiten and are tested via xvfb in CI.
// These benchmarks test the parts available without Ebiten.

// BenchmarkDrawWallsVariableHeight benchmarks map setup with variable height walls.
// Target: <8ms per frame at 1280×720 resolution.
// NOTE: Actual wall rendering requires Ebiten; this tests map cell setup.
func BenchmarkDrawWallsVariableHeight(b *testing.B) {
	r := NewRenderer(1280, 720)
	r.PlayerX = 8.5
	r.PlayerY = 8.5
	r.PlayerA = 0.0
	r.PlayerZ = 0.5
	r.PlayerPitch = 0.0

	mapSize := 32
	heightMap := make([]float64, mapSize*mapSize)
	wallHeights := make([]float64, mapSize*mapSize)

	for y := 0; y < mapSize; y++ {
		for x := 0; x < mapSize; x++ {
			idx := y*mapSize + x
			// Border walls with varying heights
			if x == 0 || y == 0 || x == mapSize-1 || y == mapSize-1 {
				heightMap[idx] = 0.8
				wallHeights[idx] = 0.5 + float64((x+y)%5)*0.4
			} else if (x+y)%5 == 0 && x > 2 && y > 2 && x < mapSize-3 && y < mapSize-3 {
				// Interior walls with variable heights
				heightMap[idx] = 0.7
				wallHeights[idx] = 0.7 + float64(x%4)*0.3
			} else {
				heightMap[idx] = 0.3
				wallHeights[idx] = DefaultWallHeight
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ClearFramebuffer()
		r.SetWorldMapWithWallHeights(heightMap, wallHeights, mapSize, 0.5)
	}
}

// BenchmarkDrawWallsWithNormals benchmarks normal lighting calculations.
// Target: <12ms per frame at 1280×720 resolution.
// NOTE: Wall rendering requires Ebiten; this tests normal map lighting calculations.
func BenchmarkDrawWallsWithNormals(b *testing.B) {
	nl := DefaultNormalLighting()

	// Create a simple test texture
	tex := &texture.Texture{
		Width:  64,
		Height: 64,
		Pixels: make([]color.RGBA, 64*64),
	}
	for i := range tex.Pixels {
		tex.Pixels[i] = color.RGBA{uint8(i % 256), uint8((i * 2) % 256), uint8((i * 3) % 256), 255}
	}

	baseColor := color.RGBA{128, 128, 128, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate applying normal lighting to a column of wall pixels
		for y := 0; y < 720; y++ {
			texX := float64(i%64) / 64.0
			texY := float64(y%64) / 64.0
			_ = nl.ApplyNormalMapLighting(tex, baseColor, texX, texY, i%2)
		}
	}
}

// BenchmarkPartialBarrierPass benchmarks transparency calculations for partial barriers.
// Target: <3ms per frame at 1280×720 resolution.
func BenchmarkPartialBarrierPass(b *testing.B) {
	r := NewRenderer(1280, 720)
	r.PlayerX = 8.5
	r.PlayerY = 8.5
	r.PlayerA = 0.0
	r.PlayerZ = 0.5
	r.PlayerPitch = 0.0

	// Set up a map with semi-opaque barriers
	mapSize := 32
	r.WorldMap = make([][]int, mapSize)
	r.WorldMapCells = make([][]MapCell, mapSize)
	for y := 0; y < mapSize; y++ {
		r.WorldMap[y] = make([]int, mapSize)
		r.WorldMapCells[y] = make([]MapCell, mapSize)
		for x := 0; x < mapSize; x++ {
			// Border walls
			if x == 0 || y == 0 || x == mapSize-1 || y == mapSize-1 {
				r.WorldMap[y][x] = 1
				r.WorldMapCells[y][x] = WallMapCell(1)
			} else if (x+y)%4 == 0 && x > 2 && y > 2 && x < mapSize-3 && y < mapSize-3 {
				// Semi-opaque barriers (fences, grates)
				r.WorldMap[y][x] = 2
				cell := WallMapCell(2)
				cell.Flags = FlagSemiOpaque
				r.WorldMapCells[y][x] = cell
			} else if (x+y)%6 == 0 && x > 2 && y > 2 && x < mapSize-3 && y < mapSize-3 {
				// Transparent barriers (glass)
				r.WorldMap[y][x] = 3
				cell := WallMapCell(3)
				cell.Flags = FlagTransparent
				r.WorldMapCells[y][x] = cell
			}
		}
	}

	// Benchmark the transparency flag checks and alpha calculations
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < mapSize; y++ {
			for x := 0; x < mapSize; x++ {
				cell := r.WorldMapCells[y][x]
				_ = getTransparencyForFlags(cell.Flags)
				if cell.Flags&FlagSemiOpaque != 0 {
					_ = isSemiOpaqueGap(float64(x), float64(y))
				}
			}
		}
	}
}

// BenchmarkInteractionRay benchmarks the interaction ray casting for targeting.
// Target: <0.05ms per frame.
func BenchmarkInteractionRay(b *testing.B) {
	r := NewRenderer(640, 480)
	r.PlayerX = 16.5
	r.PlayerY = 16.5
	r.PlayerA = 0.0
	r.PlayerZ = 0.5

	// Create sprites distributed around the player
	sprites := make([]*SpriteEntity, 50)
	for i := range sprites {
		angle := float64(i) * 0.1256 // Spread around
		dist := 3.0 + float64(i%10)
		sprites[i] = &SpriteEntity{
			X:                r.PlayerX + dist*cosApprox(angle),
			Y:                r.PlayerY + dist*sinApprox(angle),
			Scale:            0.5,
			Visible:          true,
			IsInteractable:   i%3 == 0,
			InteractionRange: 5.0,
			InteractionType:  "pickup",
		}
	}

	cfg := DefaultTargetingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.FindTargetedEntity(sprites, cfg)
	}
}

// cosApprox provides an approximate cosine for benchmarking setup.
func cosApprox(x float64) float64 {
	// Simple Taylor approximation for setup
	x2 := x * x
	return 1 - x2/2 + x2*x2/24
}

// sinApprox provides an approximate sine for benchmarking setup.
func sinApprox(x float64) float64 {
	// Simple Taylor approximation for setup
	x2 := x * x
	return x - x*x2/6 + x*x2*x2/120
}
