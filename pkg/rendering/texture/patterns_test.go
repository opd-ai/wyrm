package texture

import (
	"image/color"
	"testing"
)

func TestGetGenrePatternConfig(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		cfg := GetGenrePatternConfig(genre)
		if cfg.NoiseScale <= 0 {
			t.Errorf("genre %q has invalid NoiseScale: %f", genre, cfg.NoiseScale)
		}
		if cfg.DetailLevel < 0 || cfg.DetailLevel > 1 {
			t.Errorf("genre %q has invalid DetailLevel: %f", genre, cfg.DetailLevel)
		}
		if cfg.Saturation < 0 || cfg.Saturation > 1 {
			t.Errorf("genre %q has invalid Saturation: %f", genre, cfg.Saturation)
		}
	}
}

func TestGetGenrePatternConfigUnknown(t *testing.T) {
	cfg := GetGenrePatternConfig("unknown")
	defaultCfg := GetGenrePatternConfig("fantasy")
	if cfg.NoiseScale != defaultCfg.NoiseScale {
		t.Error("unknown genre should fall back to fantasy")
	}
}

func TestGenerateGenreTexture(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	seed := int64(42)

	for _, genre := range genres {
		tex := GenerateGenreTexture(32, 32, seed, genre)
		if tex == nil {
			t.Fatalf("GenerateGenreTexture returned nil for genre %q", genre)
		}
		if tex.Width != 32 || tex.Height != 32 {
			t.Errorf("dimensions mismatch for genre %q", genre)
		}
		if len(tex.Pixels) != 32*32 {
			t.Errorf("pixel count mismatch for genre %q", genre)
		}
	}
}

func TestGenerateGenreTextureInvalidSize(t *testing.T) {
	if tex := GenerateGenreTexture(0, 32, 42, "fantasy"); tex != nil {
		t.Error("should return nil for width=0")
	}
	if tex := GenerateGenreTexture(32, 0, 42, "fantasy"); tex != nil {
		t.Error("should return nil for height=0")
	}
}

func TestGenerateGenreTextureDeterminism(t *testing.T) {
	seed := int64(123)
	tex1 := GenerateGenreTexture(16, 16, seed, "cyberpunk")
	tex2 := GenerateGenreTexture(16, 16, seed, "cyberpunk")

	for i := range tex1.Pixels {
		if tex1.Pixels[i] != tex2.Pixels[i] {
			t.Errorf("determinism fail at pixel %d", i)
			break
		}
	}
}

func TestGenreTexturesAreDifferent(t *testing.T) {
	seed := int64(42)
	textures := map[string]*Texture{
		"fantasy":          GenerateGenreTexture(32, 32, seed, "fantasy"),
		"sci-fi":           GenerateGenreTexture(32, 32, seed, "sci-fi"),
		"horror":           GenerateGenreTexture(32, 32, seed, "horror"),
		"cyberpunk":        GenerateGenreTexture(32, 32, seed, "cyberpunk"),
		"post-apocalyptic": GenerateGenreTexture(32, 32, seed, "post-apocalyptic"),
	}

	// Each genre should produce meaningfully different textures
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for i := 0; i < len(genres); i++ {
		for j := i + 1; j < len(genres); j++ {
			tex1 := textures[genres[i]]
			tex2 := textures[genres[j]]

			diffCount := 0
			for k := range tex1.Pixels {
				if tex1.Pixels[k] != tex2.Pixels[k] {
					diffCount++
				}
			}

			// At least 15% should differ (genre patterns are more distinct)
			threshold := len(tex1.Pixels) * 15 / 100
			if diffCount < threshold {
				t.Errorf("genres %q and %q too similar: %d/%d pixels differ",
					genres[i], genres[j], diffCount, len(tex1.Pixels))
			}
		}
	}
}

func TestPatternTypes(t *testing.T) {
	// Verify each genre uses the expected pattern type
	expected := map[string]PatternType{
		"fantasy":          PatternLayered,
		"sci-fi":           PatternGrid,
		"horror":           PatternVoronoi,
		"cyberpunk":        PatternGrid,
		"post-apocalyptic": PatternDistortion,
	}

	for genre, expectedType := range expected {
		cfg := GetGenrePatternConfig(genre)
		if cfg.PatternType != expectedType {
			t.Errorf("genre %q: expected pattern type %d, got %d", genre, expectedType, cfg.PatternType)
		}
	}
}

func TestGridPattern(t *testing.T) {
	cfg := GenrePatternConfig{
		NoiseScale:  0.1,
		PatternType: PatternGrid,
	}

	// Grid pattern should have lines at regular intervals
	val := gridPattern(8.0, 8.0, 42, cfg) // On grid line
	if val <= 0 {
		t.Error("grid pattern should have non-zero value")
	}
}

func TestVoronoiPattern(t *testing.T) {
	cfg := GenrePatternConfig{
		NoiseScale: 0.1,
		Contrast:   1.0,
	}

	val := voronoiPattern(8.0, 8.0, 42, cfg)
	if val < 0 || val > 1 {
		t.Errorf("voronoi pattern value out of range: %f", val)
	}
}

func TestDistortionPattern(t *testing.T) {
	cfg := GenrePatternConfig{
		NoiseScale:          0.1,
		SecondaryNoiseScale: 0.1,
	}

	val := distortionPattern(8.0, 8.0, 42, cfg)
	if val < 0 || val > 1 {
		t.Errorf("distortion pattern value out of range: %f", val)
	}
}

func TestLayeredPattern(t *testing.T) {
	cfg := GenrePatternConfig{
		NoiseScale:          0.1,
		SecondaryNoiseScale: 0.2,
	}

	val := layeredPattern(8.0, 8.0, 42, cfg)
	if val < 0 || val > 1 {
		t.Errorf("layered pattern value out of range: %f", val)
	}
}

func TestApplyContrast(t *testing.T) {
	// No contrast change
	if applyContrast(0.5, 1.0) != 0.5 {
		t.Error("contrast 1.0 at 0.5 should not change value")
	}

	// Higher contrast pushes values away from 0.5
	high := applyContrast(0.7, 2.0)
	if high <= 0.7 {
		t.Error("high contrast should increase values above 0.5")
	}

	low := applyContrast(0.3, 2.0)
	if low >= 0.3 {
		t.Error("high contrast should decrease values below 0.5")
	}
}

func TestApplySaturation(t *testing.T) {
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	// Full saturation keeps original
	fullSat := applySaturation(red, 1.0)
	if fullSat.R != 255 || fullSat.G != 0 || fullSat.B != 0 {
		t.Error("saturation 1.0 should keep original color")
	}

	// Zero saturation becomes grayscale
	noSat := applySaturation(red, 0.0)
	if noSat.R != noSat.G || noSat.G != noSat.B {
		t.Error("saturation 0.0 should produce grayscale")
	}
}

func TestAdjustBrightness(t *testing.T) {
	base := color.RGBA{R: 128, G: 128, B: 128, A: 255}

	brighter := adjustBrightness(base, 0.1)
	if brighter.R <= base.R {
		t.Error("positive delta should increase brightness")
	}

	darker := adjustBrightness(base, -0.1)
	if darker.R >= base.R {
		t.Error("negative delta should decrease brightness")
	}
}

func TestClamp01(t *testing.T) {
	if clamp01(-0.5) != 0 {
		t.Error("clamp01 should clamp negative to 0")
	}
	if clamp01(1.5) != 1 {
		t.Error("clamp01 should clamp >1 to 1")
	}
	if clamp01(0.5) != 0.5 {
		t.Error("clamp01 should not change values in range")
	}
}

func BenchmarkGenerateGenreTexture(b *testing.B) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		b.Run(genre, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = GenerateGenreTexture(64, 64, int64(i), genre)
			}
		})
	}
}

func BenchmarkPatternGrid(b *testing.B) {
	cfg := GetGenrePatternConfig("sci-fi")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gridPattern(float64(i%100), float64(i/100), 42, cfg)
	}
}

func BenchmarkPatternVoronoi(b *testing.B) {
	cfg := GetGenrePatternConfig("horror")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = voronoiPattern(float64(i%100), float64(i/100), 42, cfg)
	}
}

func BenchmarkPatternDistortion(b *testing.B) {
	cfg := GetGenrePatternConfig("post-apocalyptic")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = distortionPattern(float64(i%100), float64(i/100), 42, cfg)
	}
}

func BenchmarkPatternLayered(b *testing.B) {
	cfg := GetGenrePatternConfig("fantasy")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = layeredPattern(float64(i%100), float64(i/100), 42, cfg)
	}
}
