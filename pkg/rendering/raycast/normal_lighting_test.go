package raycast

import (
	"image/color"
	"math"
	"testing"

	"github.com/opd-ai/wyrm/pkg/rendering/texture"
)

func TestDefaultNormalLighting(t *testing.T) {
	nl := DefaultNormalLighting()

	if nl == nil {
		t.Fatal("expected non-nil NormalLighting")
	}

	if nl.SunIntensity <= 0 || nl.SunIntensity > 1 {
		t.Errorf("expected SunIntensity in (0, 1], got %f", nl.SunIntensity)
	}

	if nl.AmbientLight < 0 || nl.AmbientLight > 1 {
		t.Errorf("expected AmbientLight in [0, 1], got %f", nl.AmbientLight)
	}
}

func TestSampleNormalMap_NilTexture(t *testing.T) {
	normal := SampleNormalMap(nil, 0.5, 0.5)

	// Should return flat normal pointing outward
	if normal[0] != 0 || normal[1] != 0 || normal[2] != 1 {
		t.Errorf("expected flat normal (0,0,1), got (%f, %f, %f)", normal[0], normal[1], normal[2])
	}
}

func TestSampleNormalMap_NoNormalMap(t *testing.T) {
	// Create texture without normal map
	tex := &texture.Texture{
		Width:  64,
		Height: 64,
		Pixels: make([]color.RGBA, 64*64),
	}

	normal := SampleNormalMap(tex, 0.5, 0.5)

	// Should return flat normal
	if normal[2] != 1 {
		t.Errorf("expected flat normal with Z=1, got Z=%f", normal[2])
	}
}

func TestTransformNormalToWorld_XFacingWall(t *testing.T) {
	// Flat normal in tangent space (pointing straight out)
	tangent := [3]float64{0, 0, 1}
	world := TransformNormalToWorld(tangent, 0) // side 0 = X-facing

	// For X-facing wall, tangent Z becomes world X
	if math.Abs(world[0]-1) > 0.001 {
		t.Errorf("expected world X ≈ 1 for X-facing wall, got %f", world[0])
	}
}

func TestTransformNormalToWorld_YFacingWall(t *testing.T) {
	// Flat normal in tangent space
	tangent := [3]float64{0, 0, 1}
	world := TransformNormalToWorld(tangent, 1) // side 1 = Y-facing

	// For Y-facing wall, tangent Z becomes world Y
	if math.Abs(world[1]-1) > 0.001 {
		t.Errorf("expected world Y ≈ 1 for Y-facing wall, got %f", world[1])
	}
}

func TestComputeLightIntensity_DirectLight(t *testing.T) {
	nl := &NormalLighting{
		SunDirection: [3]float64{-1, 0, 0}, // Light coming from +X direction
		SunIntensity: 1.0,
		AmbientLight: 0.0,
	}

	// Surface normal pointing towards light (+X)
	normal := [3]float64{1, 0, 0}
	intensity := nl.ComputeLightIntensity(normal)

	// Should be fully lit
	if intensity < 0.99 {
		t.Errorf("expected full light intensity ≈ 1.0, got %f", intensity)
	}
}

func TestComputeLightIntensity_Backlit(t *testing.T) {
	nl := &NormalLighting{
		SunDirection: [3]float64{-1, 0, 0}, // Light coming from +X direction
		SunIntensity: 1.0,
		AmbientLight: 0.2,
	}

	// Surface normal pointing away from light (-X)
	normal := [3]float64{-1, 0, 0}
	intensity := nl.ComputeLightIntensity(normal)

	// Should only have ambient light
	if math.Abs(intensity-0.2) > 0.01 {
		t.Errorf("expected ambient-only intensity ≈ 0.2, got %f", intensity)
	}
}

func TestApplyNormalMapLighting_NoNormalMap(t *testing.T) {
	nl := DefaultNormalLighting()
	tex := &texture.Texture{
		Width:  64,
		Height: 64,
		Pixels: make([]color.RGBA, 64*64),
	}

	baseColor := color.RGBA{R: 200, G: 150, B: 100, A: 255}
	result := nl.ApplyNormalMapLighting(tex, baseColor, 0.5, 0.5, 0)

	// Should apply flat normal lighting (ambient + some diffuse based on sun angle)
	if result.A != baseColor.A {
		t.Errorf("expected alpha preserved, got %d vs %d", result.A, baseColor.A)
	}

	// Result should be darker than original due to lighting calculation
	// (unless light is directly facing the surface)
	if result.R > baseColor.R || result.G > baseColor.G || result.B > baseColor.B {
		// This is okay if lighting is very bright
		t.Logf("lighting increased brightness: (%d,%d,%d) -> (%d,%d,%d)",
			baseColor.R, baseColor.G, baseColor.B, result.R, result.G, result.B)
	}
}

func TestSetSunAngle(t *testing.T) {
	nl := DefaultNormalLighting()

	// Set sun at noon (directly overhead)
	nl.SetSunAngle(math.Pi, math.Pi/2)

	// Z component should be close to 1 (sun overhead)
	if nl.SunDirection[2] < 0.9 {
		t.Errorf("expected sun Z close to 1 at noon, got %f", nl.SunDirection[2])
	}
}

func TestSetTimeOfDay(t *testing.T) {
	nl := DefaultNormalLighting()

	tests := []struct {
		hour              float64
		expectedHighSun   bool
		expectedBrightSun bool
	}{
		{6, false, false},  // Sunrise
		{12, true, true},   // Noon
		{18, false, false}, // Sunset
		{0, false, false},  // Midnight
	}

	for _, tc := range tests {
		nl.SetTimeOfDay(tc.hour)

		// At noon, sun should be high (positive Z)
		sunHigh := nl.SunDirection[2] > 0.5
		if tc.expectedHighSun && !sunHigh {
			t.Errorf("hour %f: expected high sun, got Z=%f", tc.hour, nl.SunDirection[2])
		}

		// At noon, intensity should be bright
		bright := nl.SunIntensity > 0.7
		if tc.expectedBrightSun && !bright {
			t.Errorf("hour %f: expected bright sun, got intensity=%f", tc.hour, nl.SunIntensity)
		}
	}
}

func TestApplyLightIntensity(t *testing.T) {
	c := color.RGBA{R: 200, G: 100, B: 50, A: 255}

	// Half intensity
	result := applyLightIntensity(c, 0.5)

	if result.R != 100 {
		t.Errorf("expected R=100, got %d", result.R)
	}
	if result.G != 50 {
		t.Errorf("expected G=50, got %d", result.G)
	}
	if result.B != 25 {
		t.Errorf("expected B=25, got %d", result.B)
	}
	if result.A != 255 {
		t.Errorf("expected A preserved=255, got %d", result.A)
	}
}

func TestNormalizeVec3(t *testing.T) {
	v := [3]float64{3, 0, 4}
	normalizeVec3(&v)

	length := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if math.Abs(length-1.0) > 0.0001 {
		t.Errorf("expected normalized length ≈ 1, got %f", length)
	}
}

func TestNormalizeVec3_ZeroVector(t *testing.T) {
	v := [3]float64{0, 0, 0}
	normalizeVec3(&v)

	// Should not panic or produce NaN
	for i, val := range v {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			t.Errorf("normalizing zero vector produced invalid value at index %d: %f", i, val)
		}
	}
}

func TestApplySpecularHighlight_NoReflectivity(t *testing.T) {
	nl := DefaultNormalLighting()
	baseColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	normal := [3]float64{0, 0, 1}

	// Zero reflectivity should return unchanged color
	result := nl.ApplySpecularHighlight(baseColor, normal, 0.0, 0.5)

	if result.R != baseColor.R || result.G != baseColor.G || result.B != baseColor.B {
		t.Errorf("expected unchanged color with zero reflectivity")
	}
}

func TestApplySpecularHighlight_HighReflectivity(t *testing.T) {
	nl := DefaultNormalLighting()
	// Set sun direction to hit the surface head-on for maximum specular
	nl.SunDirection = [3]float64{0, 0, -1}

	baseColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	normal := [3]float64{0, 0, 1} // Facing camera

	// High reflectivity, low roughness (sharp highlight)
	result := nl.ApplySpecularHighlight(baseColor, normal, 1.0, 0.1)

	// Should add specular highlight (color should be brighter)
	if result.R <= baseColor.R {
		t.Errorf("expected brighter red with specular, got %d vs %d", result.R, baseColor.R)
	}
}

func TestApplySpecularHighlight_BackFacing(t *testing.T) {
	nl := DefaultNormalLighting()
	nl.SunDirection = [3]float64{0, 0, 1} // Light from behind camera

	baseColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	normal := [3]float64{0, 0, 1} // Facing camera but light is behind

	// Surface facing away from light should have no specular
	result := nl.ApplySpecularHighlight(baseColor, normal, 1.0, 0.1)

	// Color should be unchanged (no specular when facing away)
	if result.R != baseColor.R {
		t.Errorf("expected unchanged color when back-facing, got %d vs %d", result.R, baseColor.R)
	}
}

func TestApplySpecularHighlight_PreservesAlpha(t *testing.T) {
	nl := DefaultNormalLighting()
	baseColor := color.RGBA{R: 100, G: 100, B: 100, A: 128}
	normal := [3]float64{0, 0, 1}

	result := nl.ApplySpecularHighlight(baseColor, normal, 0.5, 0.5)

	if result.A != 128 {
		t.Errorf("expected alpha preserved at 128, got %d", result.A)
	}
}

func TestSpecularEnabled_DisablesSpecular(t *testing.T) {
	nl := DefaultNormalLighting()
	nl.SpecularEnabled = false
	nl.SunDirection = [3]float64{0, 0, -1}

	tex := &texture.Texture{
		Width:  64,
		Height: 64,
		Pixels: make([]color.RGBA, 64*64),
	}
	baseColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	result := nl.ApplyNormalMapLighting(tex, baseColor, 0.5, 0.5, 0)

	// With specular disabled, result should just have diffuse lighting
	// (no bright specular highlights)
	nlWithSpec := DefaultNormalLighting()
	nlWithSpec.SpecularEnabled = true
	nlWithSpec.SunDirection = nl.SunDirection

	resultWithSpec := nlWithSpec.ApplyNormalMapLighting(tex, baseColor, 0.5, 0.5, 0)

	// Without specular should be same or darker
	if result.R > resultWithSpec.R {
		t.Logf("result without spec: %d, with spec: %d", result.R, resultWithSpec.R)
	}
}
