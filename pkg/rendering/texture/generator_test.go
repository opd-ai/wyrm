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

// ============================================================
// Material Registry Tests
// ============================================================

func TestNewMaterialRegistry(t *testing.T) {
	r := NewMaterialRegistry()
	if r == nil {
		t.Fatal("NewMaterialRegistry returned nil")
	}
	if r.Count() == 0 {
		t.Error("registry should have default materials")
	}
}

func TestMaterialRegistryGet(t *testing.T) {
	r := NewMaterialRegistry()

	// Test getting by ID
	stone := r.Get(MaterialStone)
	if stone == nil {
		t.Fatal("MaterialStone not found")
	}
	if stone.Name != "stone" {
		t.Errorf("expected name='stone', got '%s'", stone.Name)
	}
	if stone.ID != MaterialStone {
		t.Errorf("expected ID=%d, got %d", MaterialStone, stone.ID)
	}

	// Test getting non-existent material
	if m := r.Get(9999); m != nil {
		t.Error("should return nil for non-existent ID")
	}
}

func TestMaterialRegistryGetByName(t *testing.T) {
	r := NewMaterialRegistry()

	// Test getting by name
	wood := r.GetByName("wood")
	if wood == nil {
		t.Fatal("'wood' material not found")
	}
	if wood.ID != MaterialWood {
		t.Errorf("expected ID=%d, got %d", MaterialWood, wood.ID)
	}

	// Test non-existent name
	if m := r.GetByName("unobtainium"); m != nil {
		t.Error("should return nil for non-existent name")
	}
}

func TestMaterialRegistryGetID(t *testing.T) {
	r := NewMaterialRegistry()

	id := r.GetID("metal")
	if id != MaterialMetal {
		t.Errorf("expected MaterialMetal, got %d", id)
	}

	id = r.GetID("nonexistent")
	if id != MaterialNone {
		t.Errorf("expected MaterialNone, got %d", id)
	}
}

func TestMaterialRegistryList(t *testing.T) {
	r := NewMaterialRegistry()
	ids := r.List()
	if len(ids) == 0 {
		t.Error("List() returned empty slice")
	}
	if len(ids) != r.Count() {
		t.Errorf("List() length %d != Count() %d", len(ids), r.Count())
	}
}

func TestMaterialRegistryRegister(t *testing.T) {
	r := NewMaterialRegistry()
	initialCount := r.Count()

	// Register a custom material
	custom := &Material{
		ID:       MaterialCustom + 1,
		Name:     "mithril",
		Category: "metal",
		Physical: PhysicalProperties{
			Hardness: 0.99,
			Density:  0.3,
		},
		Visual: VisualProperties{
			Roughness: 0.1,
			Metalness: 1.0,
		},
	}
	r.Register(custom)

	if r.Count() != initialCount+1 {
		t.Error("Count should increase by 1")
	}

	// Retrieve it
	m := r.GetByName("mithril")
	if m == nil {
		t.Fatal("custom material not found by name")
	}
	if m.Physical.Hardness != 0.99 {
		t.Errorf("expected Hardness=0.99, got %f", m.Physical.Hardness)
	}
}

func TestMaterialPhysicalProperties(t *testing.T) {
	r := NewMaterialRegistry()

	tests := []struct {
		id          MaterialID
		minHardness float64
		maxHardness float64
	}{
		{MaterialStone, 0.7, 0.9},
		{MaterialWood, 0.4, 0.6},
		{MaterialMetal, 0.8, 1.0},
		{MaterialGlass, 0.5, 0.7},
		{MaterialDirt, 0.1, 0.3},
	}

	for _, tt := range tests {
		t.Run(r.Get(tt.id).Name, func(t *testing.T) {
			m := r.Get(tt.id)
			if m == nil {
				t.Fatal("material not found")
			}
			if m.Physical.Hardness < tt.minHardness || m.Physical.Hardness > tt.maxHardness {
				t.Errorf("hardness %f outside expected range [%f, %f]",
					m.Physical.Hardness, tt.minHardness, tt.maxHardness)
			}
		})
	}
}

func TestMaterialVisualProperties(t *testing.T) {
	r := NewMaterialRegistry()

	// Metal should have high metalness
	metal := r.Get(MaterialMetal)
	if metal.Visual.Metalness != 1.0 {
		t.Errorf("metal should have Metalness=1.0, got %f", metal.Visual.Metalness)
	}

	// Glass should have high transparency
	glass := r.Get(MaterialGlass)
	if glass.Visual.Transparency < 0.8 {
		t.Errorf("glass should have high transparency, got %f", glass.Visual.Transparency)
	}

	// Chrome should have low roughness (smooth)
	chrome := r.Get(MaterialChrome)
	if chrome.Visual.Roughness > 0.1 {
		t.Errorf("chrome should have low roughness, got %f", chrome.Visual.Roughness)
	}

	// Neon should have high emissive
	neon := r.Get(MaterialNeon)
	if neon.Visual.Emissive < 0.8 {
		t.Errorf("neon should have high emissive, got %f", neon.Visual.Emissive)
	}
}

func TestMaterialAcousticProperties(t *testing.T) {
	r := NewMaterialRegistry()

	// Metal should ring (high resonance)
	metal := r.Get(MaterialMetal)
	if metal.Acoustic.Resonance < 0.7 {
		t.Errorf("metal should have high resonance, got %f", metal.Acoustic.Resonance)
	}
	if metal.Acoustic.ImpactSound != "metal" {
		t.Errorf("metal impact sound should be 'metal', got '%s'", metal.Acoustic.ImpactSound)
	}

	// Dirt should absorb sound
	dirt := r.Get(MaterialDirt)
	if dirt.Acoustic.SoundAbsorption < 0.8 {
		t.Errorf("dirt should absorb sound, got %f", dirt.Acoustic.SoundAbsorption)
	}
}

func TestMaterialGetColorsForGenre(t *testing.T) {
	r := NewMaterialRegistry()

	// Stone has genre variants
	stoneColors := r.GetColorsForGenre(MaterialStone, "fantasy")
	if len(stoneColors) == 0 {
		t.Error("should return colors for stone/fantasy")
	}

	// Horror variant should differ from default
	horrorColors := r.GetColorsForGenre(MaterialStone, "horror")
	if len(horrorColors) == 0 {
		t.Error("should return colors for stone/horror")
	}

	// Non-existent material
	noColors := r.GetColorsForGenre(9999, "fantasy")
	if noColors != nil {
		t.Error("should return nil for non-existent material")
	}

	// Genre without specific variant falls back to base colors
	baseColors := r.GetColorsForGenre(MaterialBrick, "sci-fi")
	if len(baseColors) == 0 {
		t.Error("should fall back to base colors")
	}
}

func TestMaterialCategories(t *testing.T) {
	r := NewMaterialRegistry()

	categoryMap := make(map[string][]MaterialID)
	for _, id := range r.List() {
		m := r.Get(id)
		categoryMap[m.Category] = append(categoryMap[m.Category], id)
	}

	// Should have multiple categories
	if len(categoryMap) < 4 {
		t.Errorf("expected at least 4 categories, got %d", len(categoryMap))
	}

	// Check expected categories exist
	expectedCategories := []string{"mineral", "organic", "metal", "natural"}
	for _, cat := range expectedCategories {
		if len(categoryMap[cat]) == 0 {
			t.Errorf("expected materials in category '%s'", cat)
		}
	}
}

func TestDefaultMaterialRegistry(t *testing.T) {
	// DefaultMaterialRegistry should be initialized
	if DefaultMaterialRegistry == nil {
		t.Fatal("DefaultMaterialRegistry is nil")
	}
	if DefaultMaterialRegistry.Count() == 0 {
		t.Error("DefaultMaterialRegistry should have materials")
	}

	// Should be able to get materials from it
	stone := DefaultMaterialRegistry.Get(MaterialStone)
	if stone == nil {
		t.Error("should find MaterialStone in DefaultMaterialRegistry")
	}
}

func TestMaterialPropertyRanges(t *testing.T) {
	r := NewMaterialRegistry()

	for _, id := range r.List() {
		m := r.Get(id)
		t.Run(m.Name, func(t *testing.T) {
			// Physical properties should be in [0, 1]
			checkRange(t, "Hardness", m.Physical.Hardness, 0, 1)
			checkRange(t, "Density", m.Physical.Density, 0, 1)
			checkRange(t, "Friction", m.Physical.Friction, 0, 1)
			checkRange(t, "Elasticity", m.Physical.Elasticity, 0, 1)
			checkRange(t, "Conductivity", m.Physical.Conductivity, 0, 1)
			checkRange(t, "Flammability", m.Physical.Flammability, 0, 1)
			checkRange(t, "Brittleness", m.Physical.Brittleness, 0, 1)

			// Visual properties should be in [0, 1] (except refraction)
			checkRange(t, "Roughness", m.Visual.Roughness, 0, 1)
			checkRange(t, "Metalness", m.Visual.Metalness, 0, 1)
			checkRange(t, "Transparency", m.Visual.Transparency, 0, 1)
			checkRange(t, "Emissive", m.Visual.Emissive, 0, 1)
			checkRange(t, "Reflectivity", m.Visual.Reflectivity, 0, 1)
			checkRange(t, "Refraction", m.Visual.Refraction, 1, 3)
			checkRange(t, "Subsurface", m.Visual.Subsurface, 0, 1)

			// Acoustic properties should be in [0, 1]
			checkRange(t, "Resonance", m.Acoustic.Resonance, 0, 1)
			checkRange(t, "SoundAbsorption", m.Acoustic.SoundAbsorption, 0, 1)
		})
	}
}

func checkRange(t *testing.T, name string, value, min, max float64) {
	t.Helper()
	if value < min || value > max {
		t.Errorf("%s=%f outside range [%f, %f]", name, value, min, max)
	}
}

func TestMaterialBaseColors(t *testing.T) {
	r := NewMaterialRegistry()

	for _, id := range r.List() {
		m := r.Get(id)
		if len(m.BaseColors) == 0 {
			t.Errorf("material '%s' has no base colors", m.Name)
		}
	}
}

func BenchmarkMaterialRegistryGet(b *testing.B) {
	r := NewMaterialRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Get(MaterialID(i%15 + 1))
	}
}

func BenchmarkMaterialRegistryGetByName(b *testing.B) {
	r := NewMaterialRegistry()
	names := []string{"stone", "wood", "metal", "glass", "concrete"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.GetByName(names[i%len(names)])
	}
}
