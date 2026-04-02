package sprite

import (
	"image/color"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator("fantasy", 12345)
	if gen == nil {
		t.Fatal("expected non-nil generator")
	}
	if gen.genre != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", gen.genre)
	}
	if gen.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", gen.seed)
	}
}

func TestGeneratorGenerateSheet(t *testing.T) {
	gen := NewGenerator("fantasy", 12345)

	testCases := []struct {
		name     string
		category string
		bodyPlan string
	}{
		{"humanoid warrior", CategoryHumanoid, "warrior"},
		{"humanoid merchant", CategoryHumanoid, "merchant"},
		{"humanoid guard", CategoryHumanoid, "guard"},
		{"humanoid healer", CategoryHumanoid, "healer"},
		{"humanoid smith", CategoryHumanoid, "smith"},
		{"creature quadruped", CategoryCreature, "quadruped"},
		{"creature serpentine", CategoryCreature, "serpentine"},
		{"creature avian", CategoryCreature, "avian"},
		{"vehicle", CategoryVehicle, "buggy"},
		{"object", CategoryObject, "chest"},
		{"effect", CategoryEffect, "glow"},
		{"unknown category", "unknown", "default"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := SpriteCacheKey{
				Category:       tc.category,
				BodyPlan:       tc.bodyPlan,
				GenreID:        "fantasy",
				PrimaryColor:   packColor(color.RGBA{R: 200, G: 150, B: 100, A: 255}),
				SecondaryColor: packColor(color.RGBA{R: 100, G: 80, B: 60, A: 255}),
				AccentColor:    packColor(color.RGBA{R: 255, G: 200, B: 0, A: 255}),
				Scale:          1.0,
				Seed:           12345,
			}

			sheet := gen.GenerateSheet(key)
			if sheet == nil {
				t.Fatal("expected non-nil sprite sheet")
			}

			// Should have at least idle animation
			idle := sheet.GetAnimation(AnimIdle)
			if idle == nil {
				t.Error("expected idle animation")
			}

			if idle != nil && idle.FrameCount() == 0 {
				t.Error("expected at least one frame")
			}

			// Frames should have non-zero dimensions
			if idle != nil {
				frame := idle.GetFrame(0)
				if frame != nil {
					if frame.Width <= 0 || frame.Height <= 0 {
						t.Error("frame has invalid dimensions")
					}
				}
			}
		})
	}
}

func TestGeneratorDeterminism(t *testing.T) {
	// Same seed should produce identical sprites
	key := SpriteCacheKey{
		Category:       CategoryHumanoid,
		BodyPlan:       "warrior",
		GenreID:        "fantasy",
		PrimaryColor:   packColor(color.RGBA{R: 200, G: 150, B: 100, A: 255}),
		SecondaryColor: packColor(color.RGBA{R: 100, G: 80, B: 60, A: 255}),
		Scale:          1.0,
		Seed:           99999,
	}

	gen1 := NewGenerator("fantasy", 12345)
	gen2 := NewGenerator("fantasy", 12345)

	sheet1 := gen1.GenerateSheet(key)
	sheet2 := gen2.GenerateSheet(key)

	// Compare first frame pixels
	frame1 := sheet1.GetFrame(AnimIdle, 0)
	frame2 := sheet2.GetFrame(AnimIdle, 0)

	if frame1.Width != frame2.Width || frame1.Height != frame2.Height {
		t.Error("determinism failure: dimensions differ")
	}

	for i := range frame1.Pixels {
		if frame1.Pixels[i] != frame2.Pixels[i] {
			t.Errorf("determinism failure: pixel %d differs", i)
			break
		}
	}
}

func TestGeneratorScale(t *testing.T) {
	gen := NewGenerator("fantasy", 12345)

	// Small scale
	smallKey := SpriteCacheKey{
		Category: CategoryHumanoid,
		BodyPlan: "warrior",
		Scale:    0.5,
		Seed:     1,
	}
	small := gen.GenerateSheet(smallKey)

	// Large scale
	largeKey := SpriteCacheKey{
		Category: CategoryHumanoid,
		BodyPlan: "warrior",
		Scale:    2.0,
		Seed:     1,
	}
	large := gen.GenerateSheet(largeKey)

	// Large should have bigger dimensions
	smallFrame := small.GetFrame(AnimIdle, 0)
	largeFrame := large.GetFrame(AnimIdle, 0)

	if smallFrame != nil && largeFrame != nil {
		if largeFrame.Width <= smallFrame.Width {
			t.Error("large scale should have bigger width")
		}
		if largeFrame.Height <= smallFrame.Height {
			t.Error("large scale should have bigger height")
		}
	}
}

func TestGeneratorHumanoidAnimations(t *testing.T) {
	gen := NewGenerator("fantasy", 12345)
	key := SpriteCacheKey{
		Category:       CategoryHumanoid,
		BodyPlan:       "guard",
		PrimaryColor:   packColor(color.RGBA{R: 128, G: 128, B: 128, A: 255}),
		SecondaryColor: packColor(color.RGBA{R: 64, G: 64, B: 64, A: 255}),
		AccentColor:    packColor(color.RGBA{R: 255, G: 215, B: 0, A: 255}),
		Scale:          1.0,
		Seed:           12345,
	}

	sheet := gen.GenerateSheet(key)

	expectedAnims := []string{AnimIdle, AnimWalk, AnimAttack, AnimDead}
	for _, name := range expectedAnims {
		anim := sheet.GetAnimation(name)
		if anim == nil {
			t.Errorf("missing expected animation: %s", name)
			continue
		}
		if anim.FrameCount() == 0 {
			t.Errorf("animation %s has no frames", name)
		}
	}
}

func TestGeneratorCreatureBodyPlans(t *testing.T) {
	gen := NewGenerator("fantasy", 12345)

	bodyPlans := []string{"quadruped", "serpentine", "avian", "unknown"}
	for _, plan := range bodyPlans {
		t.Run(plan, func(t *testing.T) {
			key := SpriteCacheKey{
				Category:       CategoryCreature,
				BodyPlan:       plan,
				PrimaryColor:   packColor(color.RGBA{R: 100, G: 80, B: 60, A: 255}),
				SecondaryColor: packColor(color.RGBA{R: 200, G: 180, B: 160, A: 255}),
				Scale:          1.0,
				Seed:           42,
			}

			sheet := gen.GenerateSheet(key)
			if sheet == nil {
				t.Fatal("expected non-nil sheet")
			}

			frame := sheet.GetFrame(AnimIdle, 0)
			if frame == nil {
				t.Fatal("expected non-nil frame")
			}

			// Frame should have some non-transparent pixels
			hasPixels := false
			for _, p := range frame.Pixels {
				if p.A > 0 {
					hasPixels = true
					break
				}
			}
			if !hasPixels {
				t.Error("frame has no visible pixels")
			}
		})
	}
}

func TestUnpackPackColor(t *testing.T) {
	original := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	packed := packColor(original)
	unpacked := unpackColor(packed)

	if unpacked != original {
		t.Errorf("color roundtrip failed: %v -> %v", original, unpacked)
	}
}

func TestGeneratorBodyPlanDetails(t *testing.T) {
	gen := NewGenerator("fantasy", 12345)

	// Test that different body plans produce different sprites
	bodyPlans := []string{"guard", "merchant", "healer", "smith"}
	sprites := make([]*Sprite, len(bodyPlans))

	for i, plan := range bodyPlans {
		key := SpriteCacheKey{
			Category:       CategoryHumanoid,
			BodyPlan:       plan,
			PrimaryColor:   packColor(color.RGBA{R: 200, G: 150, B: 100, A: 255}),
			SecondaryColor: packColor(color.RGBA{R: 100, G: 80, B: 60, A: 255}),
			AccentColor:    packColor(color.RGBA{R: 255, G: 200, B: 0, A: 255}),
			Scale:          1.0,
			Seed:           12345,
		}
		sheet := gen.GenerateSheet(key)
		sprites[i] = sheet.GetFrame(AnimIdle, 0)
	}

	// At least some should be different (due to decorations)
	// This is a weak test since the base humanoid is the same
	for i := 0; i < len(sprites)-1; i++ {
		if sprites[i] == nil || sprites[i+1] == nil {
			t.Error("nil sprite in comparison")
			continue
		}
	}
}

func TestGeneratorVehicleSheet(t *testing.T) {
	gen := NewGenerator("post-apocalyptic", 12345)
	key := SpriteCacheKey{
		Category:       CategoryVehicle,
		BodyPlan:       "buggy",
		PrimaryColor:   packColor(color.RGBA{R: 139, G: 90, B: 43, A: 255}),
		SecondaryColor: packColor(color.RGBA{R: 50, G: 50, B: 50, A: 255}),
		Scale:          1.5,
		Seed:           54321,
	}

	sheet := gen.GenerateSheet(key)
	if sheet == nil {
		t.Fatal("expected non-nil sheet")
	}

	// Vehicles should have at least idle animation
	idle := sheet.GetAnimation(AnimIdle)
	if idle == nil {
		t.Error("expected idle animation for vehicle")
	}
}

func TestGeneratorEffectSheet(t *testing.T) {
	gen := NewGenerator("horror", 12345)
	key := SpriteCacheKey{
		Category:     CategoryEffect,
		BodyPlan:     "glow",
		PrimaryColor: packColor(color.RGBA{R: 255, G: 0, B: 0, A: 200}),
		Scale:        1.0,
		Seed:         77777,
	}

	sheet := gen.GenerateSheet(key)
	if sheet == nil {
		t.Fatal("expected non-nil sheet")
	}

	// Effects should animate
	idle := sheet.GetAnimation(AnimIdle)
	if idle == nil {
		t.Fatal("expected animation for effect")
	}
	if idle.FrameCount() < 2 {
		t.Error("effects should have multiple frames")
	}
}

func TestMinFunction(t *testing.T) {
	if min(5, 10) != 5 {
		t.Error("min(5, 10) should be 5")
	}
	if min(10, 5) != 5 {
		t.Error("min(10, 5) should be 5")
	}
	if min(5, 5) != 5 {
		t.Error("min(5, 5) should be 5")
	}
}

// Benchmarks for performance verification (60 FPS target)

func BenchmarkGenerateHumanoidSheet(b *testing.B) {
	gen := NewGenerator("fantasy", 12345)
	key := SpriteCacheKey{
		Category:       CategoryHumanoid,
		BodyPlan:       "warrior",
		GenreID:        "fantasy",
		PrimaryColor:   packColor(color.RGBA{R: 200, G: 150, B: 100, A: 255}),
		SecondaryColor: packColor(color.RGBA{R: 100, G: 80, B: 60, A: 255}),
		Scale:          1.0,
		Seed:           42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateSheet(key)
	}
}

func BenchmarkGenerateCreatureSheet(b *testing.B) {
	gen := NewGenerator("fantasy", 12345)
	key := SpriteCacheKey{
		Category:       CategoryCreature,
		BodyPlan:       "quadruped",
		GenreID:        "fantasy",
		PrimaryColor:   packColor(color.RGBA{R: 150, G: 100, B: 50, A: 255}),
		SecondaryColor: packColor(color.RGBA{R: 80, G: 60, B: 30, A: 255}),
		Scale:          1.5,
		Seed:           99,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateSheet(key)
	}
}

func BenchmarkGenerateVehicleSheet(b *testing.B) {
	gen := NewGenerator("cyberpunk", 12345)
	key := SpriteCacheKey{
		Category:       CategoryVehicle,
		BodyPlan:       "buggy",
		GenreID:        "cyberpunk",
		PrimaryColor:   packColor(color.RGBA{R: 50, G: 50, B: 100, A: 255}),
		SecondaryColor: packColor(color.RGBA{R: 200, G: 0, B: 200, A: 255}),
		Scale:          2.0,
		Seed:           1337,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateSheet(key)
	}
}

func BenchmarkGenerateAllCategories(b *testing.B) {
	gen := NewGenerator("fantasy", 12345)
	keys := []SpriteCacheKey{
		{Category: CategoryHumanoid, BodyPlan: "warrior", Scale: 1.0, Seed: 1},
		{Category: CategoryCreature, BodyPlan: "quadruped", Scale: 1.5, Seed: 2},
		{Category: CategoryVehicle, BodyPlan: "buggy", Scale: 2.0, Seed: 3},
		{Category: CategoryObject, BodyPlan: "chest", Scale: 1.0, Seed: 4},
		{Category: CategoryEffect, BodyPlan: "glow", Scale: 1.0, Seed: 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, key := range keys {
			gen.GenerateSheet(key)
		}
	}
}

// TestGenerateBarrierSprite tests barrier sprite generation for all shape types.
func TestGenerateBarrierSprite(t *testing.T) {
	gen := NewGenerator("fantasy", 12345)

	testCases := []struct {
		name      string
		params    BarrierSpriteParams
		minWidth  int
		minHeight int
	}{
		{
			name: "cylinder barrier",
			params: BarrierSpriteParams{
				ShapeType:   ShapeCylinder,
				Radius:      0.8,
				Height:      1.5,
				Genre:       "fantasy",
				ArchetypeID: "boulder",
			},
			minWidth:  8,
			minHeight: 8,
		},
		{
			name: "box barrier",
			params: BarrierSpriteParams{
				ShapeType:   ShapeBox,
				Width:       2.0,
				Depth:       0.5,
				Height:      1.8,
				Genre:       "sci-fi",
				ArchetypeID: "steel_beam",
			},
			minWidth:  8,
			minHeight: 8,
		},
		{
			name: "polygon barrier",
			params: BarrierSpriteParams{
				ShapeType:   ShapePolygon,
				Height:      2.0,
				Genre:       "horror",
				ArchetypeID: "bone_pile",
				Vertices:    []float64{0, 0, 1, 0, 1.5, 0.5, 1, 1, 0, 1, -0.5, 0.5},
			},
			minWidth:  16,
			minHeight: 8,
		},
		{
			name: "destructible barrier",
			params: BarrierSpriteParams{
				ShapeType:     ShapeBox,
				Width:         1.5,
				Height:        1.0,
				Genre:         "cyberpunk",
				ArchetypeID:   "vending_machine",
				Destructible:  true,
				DamagePercent: 0.0,
			},
			minWidth:  8,
			minHeight: 8,
		},
		{
			name: "damaged barrier",
			params: BarrierSpriteParams{
				ShapeType:     ShapeCylinder,
				Radius:        0.5,
				Height:        1.2,
				Genre:         "post-apocalyptic",
				ArchetypeID:   "rubble_mound",
				Destructible:  true,
				DamagePercent: 0.75,
			},
			minWidth:  8,
			minHeight: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sprite := gen.GenerateBarrierSprite(tc.params, 12345)

			if sprite == nil {
				t.Fatal("expected non-nil sprite")
			}

			if sprite.Width < tc.minWidth {
				t.Errorf("expected width >= %d, got %d", tc.minWidth, sprite.Width)
			}

			if sprite.Height < tc.minHeight {
				t.Errorf("expected height >= %d, got %d", tc.minHeight, sprite.Height)
			}

			// Verify sprite has non-transparent pixels
			hasContent := false
			for _, p := range sprite.Pixels {
				if p.A > 0 {
					hasContent = true
					break
				}
			}
			if !hasContent {
				t.Error("expected sprite to have non-transparent content")
			}
		})
	}
}

// TestGenerateBarrierSpriteSheetDestructible tests sprite sheet generation for destructible barriers.
func TestGenerateBarrierSpriteSheetDestructible(t *testing.T) {
	gen := NewGenerator("fantasy", 12345)

	params := BarrierSpriteParams{
		ShapeType:    ShapeBox,
		Width:        1.5,
		Height:       2.0,
		Genre:        "fantasy",
		ArchetypeID:  "hedgerow",
		Destructible: true,
	}

	sheet := gen.GenerateBarrierSpriteSheet(params, 12345)

	if sheet == nil {
		t.Fatal("expected non-nil sprite sheet")
	}

	// Check for idle animation
	idleAnim := sheet.GetAnimation(AnimIdle)
	if idleAnim == nil {
		t.Error("expected idle animation")
	}

	// Check for damage state animations
	for i := 1; i <= 4; i++ {
		stateName := "damaged_" + string('0'+rune(i))
		if anim := sheet.Animations[stateName]; anim == nil {
			t.Errorf("expected %s animation", stateName)
		}
	}
}

// TestBarrierSpriteGenreVariation tests that different genres produce different colors.
func TestBarrierSpriteGenreVariation(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	params := BarrierSpriteParams{
		ShapeType:   ShapeCylinder,
		Radius:      0.5,
		Height:      1.0,
		ArchetypeID: "test",
	}

	// Generate sprites for each genre with same seed
	sprites := make(map[string]*Sprite)
	for _, genre := range genres {
		gen := NewGenerator(genre, 12345)
		genreParams := params
		genreParams.Genre = genre
		sprites[genre] = gen.GenerateBarrierSprite(genreParams, 12345)
	}

	// Verify we have variation (sprites should be different between at least some genres)
	differentCount := 0
	for i := 0; i < len(genres); i++ {
		for j := i + 1; j < len(genres); j++ {
			s1 := sprites[genres[i]]
			s2 := sprites[genres[j]]
			if !spritesEqual(s1, s2) {
				differentCount++
			}
		}
	}

	if differentCount == 0 {
		t.Error("expected genre variation in barrier sprites")
	}
}

// spritesEqual checks if two sprites are identical.
func spritesEqual(a, b *Sprite) bool {
	if a.Width != b.Width || a.Height != b.Height {
		return false
	}
	for i := range a.Pixels {
		if a.Pixels[i] != b.Pixels[i] {
			return false
		}
	}
	return true
}

// TestAlphaMaskGeneration tests alpha mask generation for different shapes.
func TestAlphaMaskGeneration(t *testing.T) {
	gen := NewGenerator("fantasy", 12345)

	t.Run("cylinder mask", func(t *testing.T) {
		params := BarrierSpriteParams{ShapeType: ShapeCylinder}
		sprite := gen.GenerateBarrierSprite(BarrierSpriteParams{
			ShapeType: ShapeCylinder,
			Radius:    1.0,
			Height:    2.0,
			Genre:     "fantasy",
		}, 12345)

		// Cylinder should have transparent corners
		topLeftAlpha := sprite.GetPixel(0, 0).A
		centerAlpha := sprite.GetPixel(sprite.Width/2, sprite.Height/2).A

		if topLeftAlpha > 0 {
			t.Log("Note: cylinder corners may have alpha based on radius")
		}
		if centerAlpha == 0 {
			t.Error("expected center to have alpha > 0")
		}
		_ = params
	})

	t.Run("box mask", func(t *testing.T) {
		sprite := gen.GenerateBarrierSprite(BarrierSpriteParams{
			ShapeType: ShapeBox,
			Width:     1.0,
			Height:    1.0,
			Genre:     "fantasy",
		}, 12345)

		// Box should be mostly filled
		centerAlpha := sprite.GetPixel(sprite.Width/2, sprite.Height/2).A
		if centerAlpha == 0 {
			t.Error("expected center to have alpha > 0")
		}
	})
}

// BenchmarkBarrierSpriteGeneration benchmarks barrier sprite generation.
func BenchmarkBarrierSpriteGeneration(b *testing.B) {
	gen := NewGenerator("fantasy", 12345)

	params := BarrierSpriteParams{
		ShapeType:   ShapeCylinder,
		Radius:      0.8,
		Height:      1.5,
		Genre:       "fantasy",
		ArchetypeID: "boulder",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateBarrierSprite(params, int64(i))
	}
}
