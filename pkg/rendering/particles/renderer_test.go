package particles

import (
	"image/color"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	r := NewRenderer(640, 480)
	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}
	if r.width != 640 || r.height != 480 {
		t.Errorf("expected 640x480, got %dx%d", r.width, r.height)
	}
}

func TestRendererSetDimensions(t *testing.T) {
	r := NewRenderer(640, 480)
	r.SetDimensions(1280, 720)
	if r.width != 1280 || r.height != 720 {
		t.Errorf("expected 1280x720, got %dx%d", r.width, r.height)
	}
}

func TestRendererDrawNilSystem(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)
	// Should not panic
	r.Draw(nil, pixels)
}

func TestRendererDrawSmallBuffer(t *testing.T) {
	r := NewRenderer(64, 48)
	s := NewSystem(12345)
	pixels := make([]byte, 100) // Too small
	// Should not panic
	r.Draw(s, pixels)
}

func TestRendererDrawParticles(t *testing.T) {
	r := NewRenderer(64, 48)
	s := NewSystem(12345)

	// Spawn particles in visible area using a burst
	s.SpawnBurst(TypeSparks, 0.5, 0.5, 20)

	if s.ActiveCount() == 0 {
		t.Fatal("no particles spawned")
	}

	// Update to get past the fade-in period
	s.Update(0.1)

	pixels := make([]byte, 64*48*4)
	r.Draw(s, pixels)

	// Check if any pixels were drawn
	hasDrawnPixels := false
	for i := 0; i < len(pixels); i += 4 {
		if pixels[i] > 0 || pixels[i+1] > 0 || pixels[i+2] > 0 {
			hasDrawnPixels = true
			break
		}
	}
	if !hasDrawnPixels {
		t.Error("no pixels were drawn")
	}
}

func TestDrawPixelBlendBounds(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)
	c := color.RGBA{255, 128, 64, 200}

	// Test bounds checking - should not panic
	r.drawPixelBlend(-1, 25, c, pixels)
	r.drawPixelBlend(25, -1, c, pixels)
	r.drawPixelBlend(100, 25, c, pixels)
	r.drawPixelBlend(25, 100, c, pixels)

	// Valid pixel
	r.drawPixelBlend(32, 24, c, pixels)
	idx := (24*64 + 32) * 4
	if pixels[idx] == 0 && pixels[idx+1] == 0 && pixels[idx+2] == 0 {
		t.Error("valid pixel was not drawn")
	}
}

func TestDrawPixelBlendTransparent(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)
	c := color.RGBA{255, 128, 64, 0} // Fully transparent

	r.drawPixelBlend(32, 24, c, pixels)
	idx := (24*64 + 32) * 4
	if pixels[idx] != 0 || pixels[idx+1] != 0 || pixels[idx+2] != 0 {
		t.Error("transparent pixel should not be drawn")
	}
}

func TestDrawCircle(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)
	c := color.RGBA{255, 255, 255, 255}

	r.drawCircle(32, 24, 5, c, pixels)

	// Center should be drawn
	idx := (24*64 + 32) * 4
	if pixels[idx] == 0 {
		t.Error("circle center was not drawn")
	}

	// Point at edge of radius should be drawn
	idx = (24*64 + 36) * 4 // x+4
	if pixels[idx] == 0 {
		t.Error("point at radius was not drawn")
	}
}

func TestDrawCircleSmall(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)
	c := color.RGBA{255, 255, 255, 255}

	// Zero radius should still draw center pixel
	r.drawCircle(32, 24, 0, c, pixels)
	idx := (24*64 + 32) * 4
	if pixels[idx] == 0 {
		t.Error("zero radius circle should draw center pixel")
	}
}

func TestDrawRainDrop(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)
	c := color.RGBA{200, 200, 255, 200}

	r.drawRainDrop(32, 20, 8, c, pixels)

	// Check multiple points along the drop
	for i := 0; i < 8; i++ {
		idx := ((20+i)*64 + 32) * 4
		if pixels[idx] == 0 && pixels[idx+1] == 0 && pixels[idx+2] == 0 {
			t.Errorf("rain drop pixel at offset %d was not drawn", i)
		}
	}
}

func TestDrawSnowFlake(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)
	c := color.RGBA{255, 255, 255, 200}

	r.drawSnowFlake(32, 24, 3, c, pixels)

	// Center
	idx := (24*64 + 32) * 4
	if pixels[idx] == 0 {
		t.Error("snowflake center was not drawn")
	}

	// Cross points
	idx = (24*64 + 31) * 4 // left
	if pixels[idx] == 0 {
		t.Error("snowflake left was not drawn")
	}
	idx = (24*64 + 33) * 4 // right
	if pixels[idx] == 0 {
		t.Error("snowflake right was not drawn")
	}
	idx = (23*64 + 32) * 4 // top
	if pixels[idx] == 0 {
		t.Error("snowflake top was not drawn")
	}
	idx = (25*64 + 32) * 4 // bottom
	if pixels[idx] == 0 {
		t.Error("snowflake bottom was not drawn")
	}
}

func TestDrawGlow(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)
	c := color.RGBA{255, 200, 50, 255}

	r.drawGlow(32, 24, 5, c, pixels)

	// Center should be brightest
	centerIdx := (24*64 + 32) * 4
	centerBrightness := int(pixels[centerIdx]) + int(pixels[centerIdx+1]) + int(pixels[centerIdx+2])

	// Edge should be dimmer
	edgeIdx := (24*64 + 36) * 4 // x+4
	if pixels[edgeIdx] > 0 {
		edgeBrightness := int(pixels[edgeIdx]) + int(pixels[edgeIdx+1]) + int(pixels[edgeIdx+2])
		if edgeBrightness >= centerBrightness {
			t.Error("glow edge should be dimmer than center")
		}
	}
}

func TestCreateWeatherEmitters(t *testing.T) {
	preset := &WeatherPreset{
		Type:      TypeRain,
		Intensity: 0.5,
	}

	emitters := CreateWeatherEmitters(preset, 12345)
	if len(emitters) == 0 {
		t.Error("should create emitters")
	}

	for _, e := range emitters {
		if e.Type != TypeRain {
			t.Errorf("expected type %s, got %s", TypeRain, e.Type)
		}
	}
}

func TestCreateWeatherEmittersNil(t *testing.T) {
	emitters := CreateWeatherEmitters(nil, 12345)
	if emitters != nil {
		t.Error("nil preset should return nil")
	}
}

func TestSpawnCombatEffect(t *testing.T) {
	s := NewSystem(12345)
	effect := &CombatEffect{
		Type:  TypeBlood,
		X:     0.5,
		Y:     0.5,
		Count: 20,
	}

	s.SpawnCombatEffect(effect)
	if s.ActiveCount() != 20 {
		t.Errorf("expected 20 particles, got %d", s.ActiveCount())
	}
}

func TestSpawnCombatEffectNil(t *testing.T) {
	s := NewSystem(12345)
	s.SpawnCombatEffect(nil) // Should not panic
	if s.ActiveCount() != 0 {
		t.Error("nil effect should not spawn particles")
	}
}

func TestDrawParticleTypes(t *testing.T) {
	r := NewRenderer(128, 96)
	types := []string{
		TypeRain, TypeSnow, TypeDust, TypeAsh, TypeSparks,
		TypeBlood, TypeMagic, TypeSmoke, TypeFire, TypeFogWisp,
	}

	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			s := NewSystem(12345)
			e := NewEmitter(typ, 12345)
			e.X = 0.5
			e.Y = 0.5
			e.Rate = 100
			s.AddEmitter(e)
			s.Update(0.1)

			pixels := make([]byte, 128*96*4)
			r.Draw(s, pixels)
			// Should not panic for any type
		})
	}
}

func TestAlphaBlending(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)

	// Set a background color
	idx := (24*64 + 32) * 4
	pixels[idx] = 100   // R
	pixels[idx+1] = 100 // G
	pixels[idx+2] = 100 // B
	pixels[idx+3] = 255 // A

	// Draw semi-transparent pixel
	c := color.RGBA{255, 0, 0, 128} // 50% red
	r.drawPixelBlend(32, 24, c, pixels)

	// Result should be blended
	if pixels[idx] <= 100 {
		t.Error("blending should increase red")
	}
	if pixels[idx] >= 255 {
		t.Error("blending should not be full red")
	}
}

func BenchmarkRendererDraw(b *testing.B) {
	r := NewRenderer(640, 480)
	s := NewSystem(12345)
	for i := 0; i < 5; i++ {
		e := NewEmitter(TypeRain, int64(i))
		e.Rate = 100
		s.AddEmitter(e)
	}
	// Fill with particles
	for i := 0; i < 20; i++ {
		s.Update(0.05)
	}

	pixels := make([]byte, 640*480*4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Draw(s, pixels)
	}
}

func BenchmarkDrawCircle(b *testing.B) {
	r := NewRenderer(640, 480)
	pixels := make([]byte, 640*480*4)
	c := color.RGBA{255, 255, 255, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.drawCircle(320, 240, 8, c, pixels)
	}
}

func BenchmarkDrawGlow(b *testing.B) {
	r := NewRenderer(640, 480)
	pixels := make([]byte, 640*480*4)
	c := color.RGBA{255, 200, 50, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.drawGlow(320, 240, 8, c, pixels)
	}
}
