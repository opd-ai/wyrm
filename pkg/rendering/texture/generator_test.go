package texture

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/procgen/noise"
)

func TestGenerate(t *testing.T) {
	tex := Generate(64, 64)
	if tex == nil {
		t.Fatal("Generate returned nil")
	}
	if tex.Width != 64 {
		t.Errorf("expected Width=64, got %d", tex.Width)
	}
	if tex.Height != 64 {
		t.Errorf("expected Height=64, got %d", tex.Height)
	}
	if len(tex.Pixels) != 64*64 {
		t.Errorf("expected %d pixels, got %d", 64*64, len(tex.Pixels))
	}
}

func TestGenerateInvalidSize(t *testing.T) {
	if tex := Generate(0, 64); tex != nil {
		t.Error("should return nil for width=0")
	}
	if tex := Generate(64, 0); tex != nil {
		t.Error("should return nil for height=0")
	}
	if tex := Generate(-1, 64); tex != nil {
		t.Error("should return nil for negative width")
	}
	if tex := Generate(64, -1); tex != nil {
		t.Error("should return nil for negative height")
	}
}

func TestGenerateWithSeed(t *testing.T) {
	tex := GenerateWithSeed(32, 32, 12345, "fantasy")
	if tex == nil {
		t.Fatal("GenerateWithSeed returned nil")
	}
	if tex.Width != 32 || tex.Height != 32 {
		t.Error("dimensions mismatch")
	}
}

func TestGenerateWithSeedDeterminism(t *testing.T) {
	seed := int64(42)
	tex1 := GenerateWithSeed(16, 16, seed, "fantasy")
	tex2 := GenerateWithSeed(16, 16, seed, "fantasy")

	for i := range tex1.Pixels {
		if tex1.Pixels[i] != tex2.Pixels[i] {
			t.Errorf("determinism fail at pixel %d", i)
			break
		}
	}
}

func TestGenerateWithSeedDifferentSeeds(t *testing.T) {
	tex1 := GenerateWithSeed(16, 16, 12345, "fantasy")
	tex2 := GenerateWithSeed(16, 16, 54321, "fantasy")

	same := true
	for i := range tex1.Pixels {
		if tex1.Pixels[i] != tex2.Pixels[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("different seeds should produce different textures")
	}
}

func TestGenerateAllGenres(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	seed := int64(12345)

	textures := make([]*Texture, len(genres))
	for i, genre := range genres {
		tex := GenerateWithSeed(16, 16, seed, genre)
		if tex == nil {
			t.Fatalf("GenerateWithSeed returned nil for genre %q", genre)
		}
		textures[i] = tex
	}

	// Different genres should produce visually different textures
	for i := 0; i < len(genres); i++ {
		for j := i + 1; j < len(genres); j++ {
			diffCount := 0
			for k := range textures[i].Pixels {
				if textures[i].Pixels[k] != textures[j].Pixels[k] {
					diffCount++
				}
			}
			// At least 10% of pixels should differ between genres
			threshold := len(textures[i].Pixels) / 10
			if diffCount < threshold {
				t.Errorf("genres %q and %q too similar: only %d different pixels",
					genres[i], genres[j], diffCount)
			}
		}
	}
}

func TestGenerateUnknownGenre(t *testing.T) {
	tex := GenerateWithSeed(16, 16, 12345, "unknown-genre")
	if tex == nil {
		t.Fatal("should handle unknown genre gracefully")
	}
	if len(tex.Pixels) != 16*16 {
		t.Error("unknown genre should still generate valid texture")
	}
}

func TestPixelColorRange(t *testing.T) {
	tex := GenerateWithSeed(32, 32, 12345, "fantasy")

	for i, px := range tex.Pixels {
		if px.A != 255 {
			t.Errorf("pixel %d: alpha should be 255, got %d", i, px.A)
		}
	}
}

func TestGenrePaletteExists(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		palette, ok := GenrePalette[genre]
		if !ok {
			t.Errorf("genre %q missing from GenrePalette", genre)
		}
		if len(palette) < 2 {
			t.Errorf("genre %q palette too small: %d colors", genre, len(palette))
		}
	}
}

func TestNoise2D(t *testing.T) {
	seed := int64(12345)

	// Test determinism
	v1 := noise.Noise2D(0.5, 0.5, seed)
	v2 := noise.Noise2D(0.5, 0.5, seed)
	if v1 != v2 {
		t.Error("noise2D should be deterministic")
	}

	// Test range [0, 1]
	for i := 0; i < 100; i++ {
		x := float64(i) * 0.1
		y := float64(i) * 0.15
		v := noise.Noise2D(x, y, seed)
		if v < 0 || v > 1 {
			t.Errorf("noise2D value out of range: %f at (%f, %f)", v, x, y)
		}
	}
}

func TestHashToFloat(t *testing.T) {
	seed := int64(12345)

	// Test determinism
	v1 := noise.HashToFloat(10, 20, seed)
	v2 := noise.HashToFloat(10, 20, seed)
	if v1 != v2 {
		t.Error("hashToFloat should be deterministic")
	}

	// Test range [0, 1]
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			v := noise.HashToFloat(x, y, seed)
			if v < 0 || v > 1 {
				t.Errorf("hashToFloat value out of range: %f at (%d, %d)", v, x, y)
			}
		}
	}

	// Test uniqueness
	seen := make(map[float64]bool)
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			v := noise.HashToFloat(x, y, seed)
			if seen[v] {
				t.Logf("warning: duplicate hash value at (%d, %d)", x, y)
			}
			seen[v] = true
		}
	}
}

func TestSmoothstep(t *testing.T) {
	// Test boundary values
	if noise.Smoothstep(0) != 0 {
		t.Error("smoothstep(0) should be 0")
	}
	if noise.Smoothstep(1) != 1 {
		t.Error("smoothstep(1) should be 1")
	}

	// Test midpoint
	mid := noise.Smoothstep(0.5)
	if mid < 0.4 || mid > 0.6 {
		t.Errorf("smoothstep(0.5) should be near 0.5, got %f", mid)
	}

	// Test monotonicity
	prev := 0.0
	for i := 0; i <= 100; i++ {
		v := noise.Smoothstep(float64(i) / 100.0)
		if v < prev {
			t.Error("smoothstep should be monotonically increasing")
		}
		prev = v
	}
}

func TestLerp(t *testing.T) {
	// Test boundaries
	if noise.Lerp(0, 10, 0) != 0 {
		t.Error("lerp(0, 10, 0) should be 0")
	}
	if noise.Lerp(0, 10, 1) != 10 {
		t.Error("lerp(0, 10, 1) should be 10")
	}

	// Test midpoint
	if noise.Lerp(0, 10, 0.5) != 5 {
		t.Error("lerp(0, 10, 0.5) should be 5")
	}
}

func TestClampColor(t *testing.T) {
	if clampColor(-10) != 0 {
		t.Error("clampColor(-10) should be 0")
	}
	if clampColor(300) != 255 {
		t.Error("clampColor(300) should be 255")
	}
	if clampColor(128) != 128 {
		t.Error("clampColor(128) should be 128")
	}
}

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Generate(64, 64)
	}
}

func BenchmarkGenerateWithSeed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateWithSeed(64, 64, int64(i), "fantasy")
	}
}

func BenchmarkNoise2D(b *testing.B) {
	seed := int64(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = noise.Noise2D(float64(i%100)*0.1, float64(i/100)*0.1, seed)
	}
}
