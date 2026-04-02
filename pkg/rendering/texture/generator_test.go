package texture

import (
	"image/color"
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

// ============================================================
// Material-Based Texture Generation Tests
// ============================================================

func TestGenerateForMaterial(t *testing.T) {
	tests := []struct {
		material MaterialID
		genre    string
	}{
		{MaterialStone, "fantasy"},
		{MaterialWood, "fantasy"},
		{MaterialMetal, "sci-fi"},
		{MaterialGlass, "cyberpunk"},
		{MaterialDirt, "post-apocalyptic"},
		{MaterialNeon, "cyberpunk"},
		{MaterialChrome, "sci-fi"},
		{MaterialRust, "post-apocalyptic"},
	}

	for _, tt := range tests {
		name := DefaultMaterialRegistry.Get(tt.material).Name
		t.Run(name+"_"+tt.genre, func(t *testing.T) {
			tex := GenerateForMaterial(32, 32, tt.material, 12345, tt.genre)
			if tex == nil {
				t.Fatal("GenerateForMaterial returned nil")
			}
			if tex.Width != 32 || tex.Height != 32 {
				t.Errorf("expected 32x32, got %dx%d", tex.Width, tex.Height)
			}
			if len(tex.Pixels) != 32*32 {
				t.Errorf("expected %d pixels, got %d", 32*32, len(tex.Pixels))
			}
		})
	}
}

func TestGenerateForMaterialDeterminism(t *testing.T) {
	seed := int64(98765)

	tex1 := GenerateForMaterial(16, 16, MaterialMetal, seed, "fantasy")
	tex2 := GenerateForMaterial(16, 16, MaterialMetal, seed, "fantasy")

	if tex1 == nil || tex2 == nil {
		t.Fatal("textures should not be nil")
	}

	// Same seed should produce identical results
	for i := range tex1.Pixels {
		if tex1.Pixels[i] != tex2.Pixels[i] {
			t.Errorf("pixel %d differs: %v vs %v", i, tex1.Pixels[i], tex2.Pixels[i])
			break
		}
	}
}

func TestGenerateForMaterialDifferentSeeds(t *testing.T) {
	tex1 := GenerateForMaterial(16, 16, MaterialStone, 11111, "fantasy")
	tex2 := GenerateForMaterial(16, 16, MaterialStone, 22222, "fantasy")

	if tex1 == nil || tex2 == nil {
		t.Fatal("textures should not be nil")
	}

	// Different seeds should produce different results
	different := false
	for i := range tex1.Pixels {
		if tex1.Pixels[i] != tex2.Pixels[i] {
			different = true
			break
		}
	}
	if !different {
		t.Error("different seeds should produce different textures")
	}
}

func TestGenerateForMaterialInvalidSize(t *testing.T) {
	if tex := GenerateForMaterial(0, 32, MaterialStone, 0, "fantasy"); tex != nil {
		t.Error("should return nil for width=0")
	}
	if tex := GenerateForMaterial(32, 0, MaterialStone, 0, "fantasy"); tex != nil {
		t.Error("should return nil for height=0")
	}
	if tex := GenerateForMaterial(-1, 32, MaterialStone, 0, "fantasy"); tex != nil {
		t.Error("should return nil for negative width")
	}
}

func TestGenerateForMaterialUnknownMaterial(t *testing.T) {
	// Unknown material should fall back to generic texture
	tex := GenerateForMaterial(16, 16, MaterialID(9999), 12345, "fantasy")
	if tex == nil {
		t.Fatal("should fall back to generic texture for unknown material")
	}
	if tex.Width != 16 || tex.Height != 16 {
		t.Error("fallback texture dimensions wrong")
	}
}

func TestGenerateForMaterialCategories(t *testing.T) {
	// Test each material category produces valid textures
	categories := map[string]MaterialID{
		"metal":     MaterialMetal,
		"organic":   MaterialWood,
		"mineral":   MaterialStone,
		"natural":   MaterialDirt,
		"synthetic": MaterialNeon,
	}

	for category, materialID := range categories {
		t.Run(category, func(t *testing.T) {
			tex := GenerateForMaterial(32, 32, materialID, 12345, "fantasy")
			if tex == nil {
				t.Fatalf("GenerateForMaterial returned nil for %s", category)
			}

			// Check no nil/zero pixels
			hasColor := false
			for _, p := range tex.Pixels {
				if p.R > 0 || p.G > 0 || p.B > 0 {
					hasColor = true
					break
				}
			}
			if !hasColor {
				t.Errorf("texture for %s has no color", category)
			}
		})
	}
}

func TestGenerateForMaterialTransparency(t *testing.T) {
	// Glass has high transparency, should have non-255 alpha
	tex := GenerateForMaterial(16, 16, MaterialGlass, 12345, "fantasy")
	if tex == nil {
		t.Fatal("texture should not be nil")
	}

	// Check that some pixels have transparency
	hasTransparency := false
	for _, p := range tex.Pixels {
		if p.A < 255 {
			hasTransparency = true
			break
		}
	}
	if !hasTransparency {
		t.Error("glass texture should have transparency")
	}
}

func TestGenerateForMaterialEmissive(t *testing.T) {
	// Neon has high emissive
	tex := GenerateForMaterial(16, 16, MaterialNeon, 12345, "cyberpunk")
	if tex == nil {
		t.Fatal("texture should not be nil")
	}

	// Emissive materials should have bright colors
	brightPixels := 0
	for _, p := range tex.Pixels {
		if p.R > 200 || p.G > 200 || p.B > 200 {
			brightPixels++
		}
	}
	if brightPixels == 0 {
		t.Error("neon texture should have bright pixels")
	}
}

func TestGenerateForMaterialWithRegistry(t *testing.T) {
	// Test with custom registry
	r := NewMaterialRegistry()
	customMaterial := &Material{
		ID:       MaterialCustom + 5,
		Name:     "adamantium",
		Category: "metal",
		Physical: PhysicalProperties{Hardness: 1.0},
		Visual:   VisualProperties{Roughness: 0.1, Metalness: 1.0},
		BaseColors: []color.RGBA{
			{R: 0x40, G: 0x40, B: 0x50, A: 255},
		},
	}
	r.Register(customMaterial)

	tex := GenerateForMaterialWithRegistry(16, 16, MaterialCustom+5, 12345, "fantasy", r)
	if tex == nil {
		t.Fatal("should generate texture for custom material")
	}
}

func TestGenerateForMaterialNilRegistry(t *testing.T) {
	// Should fall back to default registry
	tex := GenerateForMaterialWithRegistry(16, 16, MaterialStone, 12345, "fantasy", nil)
	if tex == nil {
		t.Fatal("should use default registry when nil passed")
	}
}

func TestScaleBrightness(t *testing.T) {
	white := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	// No change
	result := scaleBrightness(white, 1.0)
	if result.R != 100 || result.G != 100 || result.B != 100 {
		t.Error("factor 1.0 should not change color")
	}

	// Double brightness
	result = scaleBrightness(white, 2.0)
	if result.R != 200 || result.G != 200 || result.B != 200 {
		t.Errorf("expected 200, got %d", result.R)
	}

	// Half brightness
	result = scaleBrightness(white, 0.5)
	if result.R != 50 || result.G != 50 || result.B != 50 {
		t.Errorf("expected 50, got %d", result.R)
	}

	// Clamping at max
	result = scaleBrightness(white, 3.0)
	if result.R != 255 {
		t.Errorf("expected clamped to 255, got %d", result.R)
	}

	// Alpha preserved
	if result.A != 255 {
		t.Error("alpha should be preserved")
	}
}

func TestSinCos(t *testing.T) {
	// Test sin approximation at key points
	if sin(0) != 0 {
		t.Error("sin(0) should be 0")
	}

	// Test cos approximation
	cosVal := cos(0)
	if cosVal < 0.99 || cosVal > 1.01 {
		t.Errorf("cos(0) should be ~1, got %f", cosVal)
	}
}

func BenchmarkGenerateForMaterial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateForMaterial(64, 64, MaterialStone, int64(i), "fantasy")
	}
}

func BenchmarkGenerateForMaterialMetal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateForMaterial(64, 64, MaterialMetal, int64(i), "sci-fi")
	}
}

func BenchmarkGenerateForMaterialOrganic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateForMaterial(64, 64, MaterialWood, int64(i), "fantasy")
	}
}

// ============================================================
// Normal Map Generation Tests
// ============================================================

func TestNormalToColor(t *testing.T) {
	tests := []struct {
		name   string
		normal Normal
		wantR  uint8
		wantG  uint8
		wantB  uint8
	}{
		{"flat", Normal{0, 0, 1}, 127, 127, 255},
		{"pointing right", Normal{1, 0, 0}, 255, 127, 127},
		{"pointing left", Normal{-1, 0, 0}, 0, 127, 127},
		{"pointing up", Normal{0, 1, 0}, 127, 255, 127},
		{"pointing down", Normal{0, -1, 0}, 127, 0, 127},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.normal.ToColor()
			if abs(int(c.R)-int(tt.wantR)) > 1 {
				t.Errorf("R: expected %d, got %d", tt.wantR, c.R)
			}
			if abs(int(c.G)-int(tt.wantG)) > 1 {
				t.Errorf("G: expected %d, got %d", tt.wantG, c.G)
			}
			if abs(int(c.B)-int(tt.wantB)) > 1 {
				t.Errorf("B: expected %d, got %d", tt.wantB, c.B)
			}
		})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestNormalFromColor(t *testing.T) {
	// Flat normal (blue = 255)
	flatColor := color.RGBA{R: 127, G: 127, B: 255, A: 255}
	n := NormalFromColor(flatColor)
	if n.Z < 0.99 {
		t.Errorf("flat normal Z should be ~1, got %f", n.Z)
	}
	if absFloat(n.X) > 0.01 || absFloat(n.Y) > 0.01 {
		t.Errorf("flat normal X,Y should be ~0, got %f, %f", n.X, n.Y)
	}

	// Right-pointing normal
	rightColor := color.RGBA{R: 255, G: 127, B: 127, A: 255}
	n = NormalFromColor(rightColor)
	if n.X < 0.99 {
		t.Errorf("right normal X should be ~1, got %f", n.X)
	}
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestNormalRoundTrip(t *testing.T) {
	// Test that Normal -> Color -> Normal preserves values
	normals := []Normal{
		{0, 0, 1},
		{0.5, 0.5, 0.707},
		{-0.5, 0.5, 0.707},
		{0, -1, 0},
	}

	for i, orig := range normals {
		c := orig.ToColor()
		result := NormalFromColor(c)

		// Allow small precision loss
		if absFloat(orig.X-result.X) > 0.02 {
			t.Errorf("normal %d: X mismatch: %f vs %f", i, orig.X, result.X)
		}
		if absFloat(orig.Y-result.Y) > 0.02 {
			t.Errorf("normal %d: Y mismatch: %f vs %f", i, orig.Y, result.Y)
		}
		if absFloat(orig.Z-result.Z) > 0.02 {
			t.Errorf("normal %d: Z mismatch: %f vs %f", i, orig.Z, result.Z)
		}
	}
}

func TestGenerateNormalMap(t *testing.T) {
	normalMap := GenerateNormalMap(32, 32, 12345, 1.0)
	if normalMap == nil {
		t.Fatal("GenerateNormalMap returned nil")
	}
	if len(normalMap) != 32*32 {
		t.Errorf("expected %d pixels, got %d", 32*32, len(normalMap))
	}

	// Check that normals are roughly unit length
	for i := 0; i < 10; i++ {
		n := NormalFromColor(normalMap[i*100])
		length := sqrt(n.X*n.X + n.Y*n.Y + n.Z*n.Z)
		if absFloat(length-1.0) > 0.1 {
			t.Errorf("normal %d not unit length: %f", i, length)
		}
	}
}

func TestGenerateNormalMapInvalidSize(t *testing.T) {
	if nm := GenerateNormalMap(0, 32, 0, 1.0); nm != nil {
		t.Error("should return nil for width=0")
	}
	if nm := GenerateNormalMap(32, 0, 0, 1.0); nm != nil {
		t.Error("should return nil for height=0")
	}
}

func TestGenerateNormalMapStrength(t *testing.T) {
	// Higher strength should produce more varied normals
	weakMap := GenerateNormalMap(16, 16, 12345, 0.1)
	strongMap := GenerateNormalMap(16, 16, 12345, 2.0)

	if weakMap == nil || strongMap == nil {
		t.Fatal("normal maps should not be nil")
	}

	// Calculate average deviation from flat normal
	weakDev := 0.0
	strongDev := 0.0
	for i := range weakMap {
		nw := NormalFromColor(weakMap[i])
		ns := NormalFromColor(strongMap[i])

		// Deviation from Z=1 (flat)
		weakDev += absFloat(1.0 - nw.Z)
		strongDev += absFloat(1.0 - ns.Z)
	}

	if strongDev <= weakDev {
		t.Error("stronger normal map should have more deviation from flat")
	}
}

func TestGenerateWithNormalMap(t *testing.T) {
	tex := GenerateWithNormalMap(32, 32, 12345, "fantasy", 1.0)
	if tex == nil {
		t.Fatal("GenerateWithNormalMap returned nil")
	}
	if len(tex.Pixels) != 32*32 {
		t.Errorf("expected %d pixels, got %d", 32*32, len(tex.Pixels))
	}
	if !tex.HasNormalMap() {
		t.Error("texture should have normal map")
	}
	if len(tex.NormalMap) != 32*32 {
		t.Errorf("expected %d normal map pixels, got %d", 32*32, len(tex.NormalMap))
	}
}

func TestGenerateWithNormalMapInvalidSize(t *testing.T) {
	if tex := GenerateWithNormalMap(0, 32, 0, "fantasy", 1.0); tex != nil {
		t.Error("should return nil for width=0")
	}
	if tex := GenerateWithNormalMap(32, -1, 0, "fantasy", 1.0); tex != nil {
		t.Error("should return nil for negative height")
	}
}

func TestTextureHasNormalMap(t *testing.T) {
	// Texture without normal map
	texNoNormal := GenerateWithSeed(16, 16, 12345, "fantasy")
	if texNoNormal.HasNormalMap() {
		t.Error("texture without normal map should return false")
	}

	// Texture with normal map
	texWithNormal := GenerateWithNormalMap(16, 16, 12345, "fantasy", 1.0)
	if !texWithNormal.HasNormalMap() {
		t.Error("texture with normal map should return true")
	}

	// Nil texture
	var nilTex *Texture
	if nilTex.HasNormalMap() {
		t.Error("nil texture should return false")
	}
}

func TestTextureGetNormalAt(t *testing.T) {
	tex := GenerateWithNormalMap(16, 16, 12345, "fantasy", 1.0)

	// Valid coordinates
	n := tex.GetNormalAt(5, 5)
	length := sqrt(n.X*n.X + n.Y*n.Y + n.Z*n.Z)
	if absFloat(length-1.0) > 0.1 {
		t.Errorf("normal at (5,5) not unit length: %f", length)
	}

	// Out of bounds - should return flat normal
	n = tex.GetNormalAt(-1, 5)
	if n.Z != 1 {
		t.Error("out of bounds should return flat normal")
	}

	n = tex.GetNormalAt(5, 100)
	if n.Z != 1 {
		t.Error("out of bounds should return flat normal")
	}
}

func TestTextureGetNormalAtNoNormalMap(t *testing.T) {
	tex := GenerateWithSeed(16, 16, 12345, "fantasy")
	n := tex.GetNormalAt(5, 5)
	if n.X != 0 || n.Y != 0 || n.Z != 1 {
		t.Error("texture without normal map should return flat normal")
	}
}

func TestNormalizeNormal(t *testing.T) {
	// Non-unit normal
	n := normalizeNormal(Normal{X: 3, Y: 4, Z: 0})
	length := sqrt(n.X*n.X + n.Y*n.Y + n.Z*n.Z)
	if absFloat(length-1.0) > 0.001 {
		t.Errorf("normalized normal should be unit length, got %f", length)
	}

	// Correct ratios (3-4-5 triangle)
	if absFloat(n.X-0.6) > 0.01 {
		t.Errorf("expected X=0.6, got %f", n.X)
	}
	if absFloat(n.Y-0.8) > 0.01 {
		t.Errorf("expected Y=0.8, got %f", n.Y)
	}

	// Zero-length normal
	n = normalizeNormal(Normal{X: 0, Y: 0, Z: 0})
	if n.Z != 1 {
		t.Error("zero-length normal should return default flat normal")
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0, 0},
		{1, 1},
		{4, 2},
		{9, 3},
		{16, 4},
		{2, 1.414},
	}

	for _, tt := range tests {
		result := sqrt(tt.input)
		if absFloat(result-tt.expected) > 0.01 {
			t.Errorf("sqrt(%f): expected %f, got %f", tt.input, tt.expected, result)
		}
	}
}

func BenchmarkGenerateNormalMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateNormalMap(64, 64, int64(i), 1.0)
	}
}

func BenchmarkGenerateWithNormalMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateWithNormalMap(64, 64, int64(i), "fantasy", 1.0)
	}
}

// ============================================================
// Surface Wear/Aging Tests
// ============================================================

func TestWearType(t *testing.T) {
	// Test wear type constants exist
	types := []WearType{WearNone, WearScratches, WearRust, WearDirt, WearFade, WearChip, WearMoss, WearStain}
	for i, wt := range types {
		if int(wt) != i {
			t.Errorf("wear type %d should have value %d", wt, i)
		}
	}
}

func TestGetWearConfigForMaterial(t *testing.T) {
	tests := []struct {
		material     MaterialID
		expectedWear WearType
	}{
		{MaterialMetal, WearRust},
		{MaterialWood, WearFade},
		{MaterialStone, WearChip},
		{MaterialDirt, WearDirt},
		{MaterialNeon, WearFade},
	}

	for _, tt := range tests {
		name := DefaultMaterialRegistry.Get(tt.material).Name
		t.Run(name, func(t *testing.T) {
			config := GetWearConfigForMaterial(tt.material, 0.5, 12345)
			if config.PrimaryWear != tt.expectedWear {
				t.Errorf("expected primary wear %d, got %d", tt.expectedWear, config.PrimaryWear)
			}
			if config.Age != 0.5 {
				t.Errorf("expected age 0.5, got %f", config.Age)
			}
		})
	}
}

func TestGetWearConfigForUnknownMaterial(t *testing.T) {
	config := GetWearConfigForMaterial(9999, 0.5, 12345)
	// Should return default config with dirt wear
	if config.PrimaryWear != WearDirt {
		t.Errorf("unknown material should default to WearDirt, got %d", config.PrimaryWear)
	}
}

func TestApplyWear(t *testing.T) {
	// Create a simple test texture
	pixels := make([]color.RGBA, 16*16)
	for i := range pixels {
		pixels[i] = color.RGBA{R: 150, G: 150, B: 150, A: 255}
	}

	config := WearConfig{
		Age:            0.5,
		WearResistance: 0.3,
		PrimaryWear:    WearRust,
		SecondaryWear:  WearScratches,
		WearSeed:       12345,
	}

	result := ApplyWear(pixels, 16, 16, config)
	if len(result) != len(pixels) {
		t.Errorf("result should have same length as input")
	}

	// Check that some pixels were modified
	modified := false
	for i := range result {
		if result[i] != pixels[i] {
			modified = true
			break
		}
	}
	if !modified {
		t.Error("wear should modify at least some pixels")
	}
}

func TestApplyWearZeroAge(t *testing.T) {
	pixels := make([]color.RGBA, 16*16)
	for i := range pixels {
		pixels[i] = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	}

	config := WearConfig{
		Age:         0.0,
		PrimaryWear: WearRust,
		WearSeed:    12345,
	}

	result := ApplyWear(pixels, 16, 16, config)

	// With zero age, pixels should be unchanged
	for i := range result {
		if result[i] != pixels[i] {
			t.Error("zero age should not modify pixels")
			break
		}
	}
}

func TestApplyWearHighAge(t *testing.T) {
	pixels := make([]color.RGBA, 16*16)
	for i := range pixels {
		pixels[i] = color.RGBA{R: 200, G: 200, B: 200, A: 255}
	}

	config := WearConfig{
		Age:            2.0, // Very old
		WearResistance: 0.0, // No resistance
		PrimaryWear:    WearDirt,
		WearSeed:       12345,
	}

	result := ApplyWear(pixels, 16, 16, config)

	// High age should significantly modify pixels
	totalDiff := 0
	for i := range result {
		totalDiff += abs(int(result[i].R) - int(pixels[i].R))
		totalDiff += abs(int(result[i].G) - int(pixels[i].G))
		totalDiff += abs(int(result[i].B) - int(pixels[i].B))
	}

	if totalDiff < 100 {
		t.Error("high age should significantly modify texture")
	}
}

func TestApplyWearEmptyPixels(t *testing.T) {
	config := WearConfig{Age: 0.5, PrimaryWear: WearRust}
	result := ApplyWear(nil, 0, 0, config)
	if result != nil {
		t.Error("empty input should return nil")
	}

	result = ApplyWear([]color.RGBA{}, 16, 16, config)
	if len(result) != 0 {
		t.Error("empty slice should return empty")
	}
}

func TestApplyWearDeterminism(t *testing.T) {
	pixels := make([]color.RGBA, 32*32)
	for i := range pixels {
		pixels[i] = color.RGBA{R: 128, G: 128, B: 128, A: 255}
	}

	config := WearConfig{
		Age:         0.7,
		PrimaryWear: WearRust,
		WearSeed:    54321,
	}

	result1 := ApplyWear(pixels, 32, 32, config)
	result2 := ApplyWear(pixels, 32, 32, config)

	// Same config should produce identical results
	for i := range result1 {
		if result1[i] != result2[i] {
			t.Errorf("wear should be deterministic: pixel %d differs", i)
			break
		}
	}
}

func TestWearEffectTypes(t *testing.T) {
	// Use larger texture to ensure all wear types have room to work
	width, height := 32, 32

	wearTypes := []WearType{WearScratches, WearRust, WearDirt, WearFade, WearChip, WearMoss, WearStain}

	for _, wt := range wearTypes {
		t.Run(string(rune('A'+int(wt))), func(t *testing.T) {
			// Use colored pixels for fade test, gray for others
			var baseColor color.RGBA
			if wt == WearFade {
				// WearFade needs a saturated color to show effect
				baseColor = color.RGBA{R: 200, G: 100, B: 50, A: 255}
			} else {
				baseColor = color.RGBA{R: 150, G: 150, B: 150, A: 255}
			}

			basePixels := make([]color.RGBA, width*height)
			for i := range basePixels {
				basePixels[i] = baseColor
			}

			pixels := make([]color.RGBA, len(basePixels))
			copy(pixels, basePixels)

			config := WearConfig{
				Age:         1.0, // Maximum age for best coverage
				PrimaryWear: wt,
				WearSeed:    12345,
			}

			result := ApplyWear(pixels, width, height, config)

			// Each wear type should produce some modification
			modified := false
			for i := range result {
				if result[i] != basePixels[i] {
					modified = true
					break
				}
			}
			if !modified {
				t.Errorf("wear type %d should modify pixels", wt)
			}
		})
	}
}

func TestBlendColor(t *testing.T) {
	a := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	b := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	// 0% blend = a
	result := blendColor(a, b, 0.0)
	if result != a {
		t.Error("0% blend should return a")
	}

	// 100% blend = b
	result = blendColor(a, b, 1.0)
	if result != b {
		t.Error("100% blend should return b")
	}

	// 50% blend
	result = blendColor(a, b, 0.5)
	if result.R != 150 || result.G != 150 || result.B != 150 {
		t.Errorf("50%% blend should be midpoint, got (%d,%d,%d)", result.R, result.G, result.B)
	}

	// Alpha preserved
	if result.A != a.A {
		t.Error("alpha should be preserved from source")
	}
}

func TestNoiseAt(t *testing.T) {
	// Test determinism
	n1 := noiseAt(1.5, 2.5, 12345)
	n2 := noiseAt(1.5, 2.5, 12345)
	if n1 != n2 {
		t.Error("noise should be deterministic")
	}

	// Test range [0, 1]
	for i := 0; i < 100; i++ {
		n := noiseAt(float64(i)*0.1, float64(i)*0.15, 12345)
		if n < 0 || n > 1 {
			t.Errorf("noise out of range: %f", n)
		}
	}

	// Different seeds produce different values
	n3 := noiseAt(1.5, 2.5, 12346)
	if n1 == n3 {
		t.Error("different seeds should produce different noise")
	}
}

func TestSimpleRng(t *testing.T) {
	rng := seedRng(12345)

	// Test Intn range
	for i := 0; i < 100; i++ {
		v := rng.Intn(10)
		if v < 0 || v >= 10 {
			t.Errorf("Intn(10) out of range: %d", v)
		}
	}

	// Test Float64 range
	rng = seedRng(12345)
	for i := 0; i < 100; i++ {
		f := rng.Float64()
		if f < 0 || f >= 1 {
			t.Errorf("Float64() out of range: %f", f)
		}
	}

	// Test determinism
	rng1 := seedRng(54321)
	rng2 := seedRng(54321)
	for i := 0; i < 10; i++ {
		if rng1.Intn(100) != rng2.Intn(100) {
			t.Error("RNG should be deterministic")
		}
	}
}

func TestSinCosApprox(t *testing.T) {
	// Test basic values
	if absFloat(sinApprox(0)) > 0.01 {
		t.Error("sin(0) should be 0")
	}
	if absFloat(cosApprox(0)-1) > 0.01 {
		t.Error("cos(0) should be 1")
	}

	// Test at π/2
	if absFloat(sinApprox(1.5708)-1) > 0.05 {
		t.Errorf("sin(π/2) should be ~1, got %f", sinApprox(1.5708))
	}
	if absFloat(cosApprox(1.5708)) > 0.05 {
		t.Errorf("cos(π/2) should be ~0, got %f", cosApprox(1.5708))
	}
}

func BenchmarkApplyWear(b *testing.B) {
	pixels := make([]color.RGBA, 64*64)
	for i := range pixels {
		pixels[i] = color.RGBA{R: 128, G: 128, B: 128, A: 255}
	}

	config := WearConfig{
		Age:            0.5,
		WearResistance: 0.3,
		PrimaryWear:    WearRust,
		SecondaryWear:  WearDirt,
		WearSeed:       12345,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyWear(pixels, 64, 64, config)
	}
}

// ============================================================
// Genre-Specific Palette Tests
// ============================================================

func TestGenrePalettes(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			p := GetGenreColorScheme(genre)
			if p == nil {
				t.Fatal("palette should not be nil")
			}
			if len(p.Primary) == 0 {
				t.Error("primary colors should not be empty")
			}
			if len(p.Accent) == 0 {
				t.Error("accent colors should not be empty")
			}
			if p.Saturation <= 0 {
				t.Error("saturation should be positive")
			}
			if p.Brightness <= 0 {
				t.Error("brightness should be positive")
			}
			if p.Contrast <= 0 {
				t.Error("contrast should be positive")
			}
		})
	}
}

func TestGetGenreColorSchemeUnknown(t *testing.T) {
	p := GetGenreColorScheme("unknown-genre")
	if p == nil {
		t.Fatal("should return default palette for unknown genre")
	}
	if len(p.Primary) == 0 {
		t.Error("default palette should have primary colors")
	}
}

func TestGetMaterialPaletteForGenre(t *testing.T) {
	tests := []struct {
		material  MaterialID
		genre     string
		condition float64
	}{
		{MaterialStone, "fantasy", 0.0},
		{MaterialStone, "horror", 0.5},
		{MaterialMetal, "sci-fi", 0.0},
		{MaterialMetal, "post-apocalyptic", 0.8},
		{MaterialWood, "fantasy", 0.3},
		{MaterialGlass, "cyberpunk", 0.0},
	}

	for _, tt := range tests {
		name := DefaultMaterialRegistry.Get(tt.material).Name
		t.Run(name+"/"+tt.genre, func(t *testing.T) {
			colors := GetMaterialPaletteForGenre(tt.material, tt.genre, tt.condition)
			if colors == nil {
				t.Fatal("should return colors")
			}
			if len(colors) == 0 {
				t.Error("colors should not be empty")
			}
			// Verify all colors have valid alpha
			for i, c := range colors {
				if c.A == 0 {
					t.Errorf("color %d has zero alpha", i)
				}
			}
		})
	}
}

func TestGetMaterialPaletteForGenreUnknownMaterial(t *testing.T) {
	colors := GetMaterialPaletteForGenre(9999, "fantasy", 0.5)
	if colors != nil {
		t.Error("unknown material should return nil")
	}
}

func TestApplyConditionToColors(t *testing.T) {
	base := []color.RGBA{
		{R: 200, G: 100, B: 50, A: 255},
	}

	// Zero condition should return unchanged colors
	result := applyConditionToColors(base, 0, "fantasy")
	if result[0] != base[0] {
		t.Error("zero condition should not modify colors")
	}

	// High condition should modify colors
	result = applyConditionToColors(base, 1.0, "post-apocalyptic")
	if result[0] == base[0] {
		t.Error("high condition should modify colors")
	}
}

func TestApplyGenreStyling(t *testing.T) {
	baseColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}

	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			palette := GetGenreColorScheme(genre)
			result := applyGenreStyling(baseColor, palette, "mineral")

			// Result should be different from base (unless palette has no effect)
			// We can't guarantee difference since some palettes might not change gray
			// Just verify it doesn't crash and returns valid color
			if result.A != baseColor.A {
				t.Error("alpha should be preserved")
			}
		})
	}
}

func TestGetRustyMetalPalette(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	severities := []float64{0.0, 0.5, 1.0}

	for _, genre := range genres {
		for _, sev := range severities {
			t.Run(genre, func(t *testing.T) {
				colors := GetRustyMetalPalette(genre, sev)
				if len(colors) != 4 {
					t.Errorf("expected 4 colors, got %d", len(colors))
				}
				// Verify rust colors are in reasonable range (reddish-orange-brown)
				for i, c := range colors {
					if c.R < c.B {
						t.Errorf("rust color %d should have R >= B (got R=%d, B=%d)", i, c.R, c.B)
					}
				}
			})
		}
	}
}

func TestGetPolishedChromePalette(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	reflectivities := []float64{0.0, 0.5, 1.0}

	for _, genre := range genres {
		for _, ref := range reflectivities {
			t.Run(genre, func(t *testing.T) {
				colors := GetPolishedChromePalette(genre, ref)
				if len(colors) != 4 {
					t.Errorf("expected 4 colors, got %d", len(colors))
				}
				// Chrome colors should be relatively bright
				for i, c := range colors {
					brightness := (int(c.R) + int(c.G) + int(c.B)) / 3
					if brightness < 100 {
						t.Errorf("chrome color %d should be bright (got avg=%d)", i, brightness)
					}
				}
			})
		}
	}
}

func TestGetWeatheredStonePalette(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	weatherings := []float64{0.0, 0.5, 1.0}

	for _, genre := range genres {
		for _, w := range weatherings {
			t.Run(genre, func(t *testing.T) {
				colors := GetWeatheredStonePalette(genre, w)
				if len(colors) != 4 {
					t.Errorf("expected 4 colors, got %d", len(colors))
				}
			})
		}
	}
}

func TestClampByte(t *testing.T) {
	tests := []struct {
		input    float64
		expected uint8
	}{
		{-10, 0},
		{0, 0},
		{128, 128},
		{255, 255},
		{300, 255},
	}

	for _, tt := range tests {
		result := clampByte(tt.input)
		if result != tt.expected {
			t.Errorf("clampByte(%f): expected %d, got %d", tt.input, tt.expected, result)
		}
	}
}

func BenchmarkGetMaterialPaletteForGenre(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetMaterialPaletteForGenre(MaterialStone, "fantasy", 0.5)
	}
}

func BenchmarkGetRustyMetalPalette(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetRustyMetalPalette("post-apocalyptic", 0.7)
	}
}
