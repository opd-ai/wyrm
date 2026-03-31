package lighting

import (
	"image/color"
	"testing"
)

func TestNewPointLight(t *testing.T) {
	l := NewPointLight(1, 2, 3, 0.8, color.RGBA{255, 128, 64, 255})
	if l == nil {
		t.Fatal("NewPointLight returned nil")
	}
	if l.Type != TypePoint {
		t.Errorf("expected type %s, got %s", TypePoint, l.Type)
	}
	if l.X != 1 || l.Y != 2 || l.Z != 3 {
		t.Error("position not set correctly")
	}
	if l.Intensity != 0.8 {
		t.Errorf("expected intensity 0.8, got %f", l.Intensity)
	}
	if !l.Enabled {
		t.Error("light should be enabled by default")
	}
}

func TestNewDirectionalLight(t *testing.T) {
	l := NewDirectionalLight(1, 0, 0, 1.0, color.RGBA{255, 255, 255, 255})
	if l == nil {
		t.Fatal("NewDirectionalLight returned nil")
	}
	if l.Type != TypeDirectional {
		t.Errorf("expected type %s, got %s", TypeDirectional, l.Type)
	}
	// Direction should be normalized
	if l.DirX != 1 || l.DirY != 0 || l.DirZ != 0 {
		t.Errorf("direction not normalized correctly: %f, %f, %f", l.DirX, l.DirY, l.DirZ)
	}
}

func TestNewAmbientLight(t *testing.T) {
	l := NewAmbientLight(0.3, color.RGBA{50, 50, 60, 255})
	if l == nil {
		t.Fatal("NewAmbientLight returned nil")
	}
	if l.Type != TypeAmbient {
		t.Errorf("expected type %s, got %s", TypeAmbient, l.Type)
	}
	if l.Intensity != 0.3 {
		t.Errorf("expected intensity 0.3, got %f", l.Intensity)
	}
}

func TestCalculateAttenuation(t *testing.T) {
	l := NewPointLight(0, 0, 0, 1.0, color.RGBA{255, 255, 255, 255})
	l.Range = 10.0
	l.Falloff = 2.0

	tests := []struct {
		distance float64
		wantMin  float64
		wantMax  float64
	}{
		{0, 0.99, 1.01},   // At light source
		{5, 0.1, 0.5},     // Half range
		{10, -0.01, 0.01}, // At max range
		{15, -0.01, 0.01}, // Beyond range
	}

	for _, tt := range tests {
		got := l.CalculateAttenuation(tt.distance)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("attenuation at distance %f = %f, want [%f, %f]",
				tt.distance, got, tt.wantMin, tt.wantMax)
		}
	}
}

func TestCalculateAttenuationDisabled(t *testing.T) {
	l := NewPointLight(0, 0, 0, 1.0, color.RGBA{255, 255, 255, 255})
	l.Enabled = false
	if l.CalculateAttenuation(5) != 1.0 {
		t.Error("disabled light should return 1.0")
	}
}

func TestCalculateAttenuationAmbient(t *testing.T) {
	l := NewAmbientLight(0.5, color.RGBA{50, 50, 50, 255})
	if l.CalculateAttenuation(100) != 1.0 {
		t.Error("ambient light should always return 1.0")
	}
}

func TestGetColorAtDistance(t *testing.T) {
	l := NewPointLight(0, 0, 0, 1.0, color.RGBA{200, 100, 50, 255})
	l.Range = 10.0

	// At source
	c := l.GetColorAtDistance(0)
	if c.R < 180 || c.G < 80 || c.B < 40 {
		t.Error("color at source should be close to light color")
	}

	// Far away
	c = l.GetColorAtDistance(15)
	if c.R != 0 && c.G != 0 && c.B != 0 {
		t.Error("color beyond range should be zero")
	}
}

func TestNewSystem(t *testing.T) {
	s := NewSystem("fantasy")
	if s == nil {
		t.Fatal("NewSystem returned nil")
	}
	if s.ambient == nil {
		t.Error("system should have ambient light")
	}
	if s.sun == nil {
		t.Error("system should have sun")
	}
	if s.timeOfDay != 12.0 {
		t.Errorf("default time should be 12.0, got %f", s.timeOfDay)
	}
}

func TestNewSystemUnknownGenre(t *testing.T) {
	s := NewSystem("unknown")
	if s == nil {
		t.Fatal("NewSystem returned nil for unknown genre")
	}
	// Should fall back to fantasy
	if s.genrePalette.Sunlight != GenrePalettes["fantasy"].Sunlight {
		t.Error("unknown genre should fall back to fantasy")
	}
}

func TestSetGenre(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetGenre("cyberpunk")
	if s.genrePalette.Sunlight != GenrePalettes["cyberpunk"].Sunlight {
		t.Error("genre palette not updated")
	}
}

func TestSetTimeOfDay(t *testing.T) {
	s := NewSystem("fantasy")

	tests := []struct {
		input    float64
		expected float64
	}{
		{6.0, 6.0},
		{18.0, 18.0},
		{25.0, 1.0},  // Wraps around
		{-1.0, 23.0}, // Negative wraps
		{48.0, 0.0},  // Multiple days
	}

	for _, tt := range tests {
		s.SetTimeOfDay(tt.input)
		if s.GetTimeOfDay() != tt.expected {
			t.Errorf("SetTimeOfDay(%f) = %f, want %f", tt.input, s.GetTimeOfDay(), tt.expected)
		}
	}
}

func TestAdvanceTime(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetTimeOfDay(12.0)
	s.AdvanceTime(3.0)
	if s.GetTimeOfDay() != 15.0 {
		t.Errorf("expected 15.0, got %f", s.GetTimeOfDay())
	}
	s.AdvanceTime(12.0) // Should wrap to 3.0
	if s.GetTimeOfDay() != 3.0 {
		t.Errorf("expected 3.0, got %f", s.GetTimeOfDay())
	}
}

func TestIsDaytime(t *testing.T) {
	s := NewSystem("fantasy")

	s.SetTimeOfDay(12.0)
	if !s.IsDaytime() {
		t.Error("noon should be daytime")
	}

	s.SetTimeOfDay(3.0)
	if s.IsDaytime() {
		t.Error("3AM should not be daytime")
	}

	s.SetTimeOfDay(6.0)
	if !s.IsDaytime() {
		t.Error("6AM should be daytime")
	}

	s.SetTimeOfDay(18.0)
	if s.IsDaytime() {
		t.Error("6PM should not be daytime")
	}
}

func TestIsNight(t *testing.T) {
	s := NewSystem("fantasy")

	s.SetTimeOfDay(23.0)
	if !s.IsNight() {
		t.Error("11PM should be night")
	}

	s.SetTimeOfDay(12.0)
	if s.IsNight() {
		t.Error("noon should not be night")
	}
}

func TestIndoorMode(t *testing.T) {
	s := NewSystem("fantasy")

	if s.IsIndoor() {
		t.Error("should not be indoor by default")
	}

	s.SetIndoorMode(true)
	if !s.IsIndoor() {
		t.Error("should be indoor after SetIndoorMode(true)")
	}

	s.SetIndoorMode(false)
	if s.IsIndoor() {
		t.Error("should not be indoor after SetIndoorMode(false)")
	}
}

func TestAddRemoveLight(t *testing.T) {
	s := NewSystem("fantasy")
	l := NewPointLight(5, 5, 0, 1.0, color.RGBA{255, 255, 255, 255})

	s.AddLight(l)
	if s.LightCount() != 1 {
		t.Errorf("expected 1 light, got %d", s.LightCount())
	}

	s.RemoveLight(l)
	if s.LightCount() != 0 {
		t.Errorf("expected 0 lights, got %d", s.LightCount())
	}

	// Should handle nil
	s.AddLight(nil)
	if s.LightCount() != 0 {
		t.Error("should not add nil light")
	}
}

func TestClearLights(t *testing.T) {
	s := NewSystem("fantasy")
	s.AddLight(NewPointLight(1, 1, 0, 1.0, color.RGBA{255, 255, 255, 255}))
	s.AddLight(NewPointLight(2, 2, 0, 1.0, color.RGBA{255, 255, 255, 255}))
	s.AddLight(NewPointLight(3, 3, 0, 1.0, color.RGBA{255, 255, 255, 255}))

	s.ClearLights()
	if s.LightCount() != 0 {
		t.Errorf("expected 0 lights after clear, got %d", s.LightCount())
	}
}

func TestLights(t *testing.T) {
	s := NewSystem("fantasy")
	l1 := NewPointLight(1, 1, 0, 1.0, color.RGBA{255, 255, 255, 255})
	l2 := NewPointLight(2, 2, 0, 1.0, color.RGBA{255, 255, 255, 255})
	s.AddLight(l1)
	s.AddLight(l2)

	lights := s.Lights()
	if len(lights) != 2 {
		t.Errorf("expected 2 lights, got %d", len(lights))
	}
}

func TestGetAmbientAndSun(t *testing.T) {
	s := NewSystem("fantasy")

	if s.GetAmbient() == nil {
		t.Error("GetAmbient should not return nil")
	}
	if s.GetSun() == nil {
		t.Error("GetSun should not return nil")
	}
}

func TestCalculateLightingAt(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetTimeOfDay(12.0) // Noon - max sun

	// At noon with no extra lights, should have good intensity
	intensity, _ := s.CalculateLightingAt(5, 5, 0)
	if intensity < 0.5 {
		t.Errorf("expected intensity >= 0.5 at noon, got %f", intensity)
	}

	// Add a point light nearby
	l := NewPointLight(5, 5, 0, 1.0, color.RGBA{255, 200, 100, 255})
	s.AddLight(l)

	intensityWithLight, _ := s.CalculateLightingAt(5, 5, 0)
	if intensityWithLight <= intensity {
		t.Error("intensity should increase with nearby light")
	}
}

func TestCalculateLightingAtIndoor(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetTimeOfDay(12.0)

	outdoorIntensity, _ := s.CalculateLightingAt(5, 5, 0)

	s.SetIndoorMode(true)
	indoorIntensity, _ := s.CalculateLightingAt(5, 5, 0)

	if indoorIntensity >= outdoorIntensity {
		t.Error("indoor should be darker than outdoor")
	}
}

func TestCalculateLightingAtNight(t *testing.T) {
	s := NewSystem("fantasy")

	s.SetTimeOfDay(12.0)
	dayIntensity, _ := s.CalculateLightingAt(5, 5, 0)

	s.SetTimeOfDay(0.0) // Midnight
	nightIntensity, _ := s.CalculateLightingAt(5, 5, 0)

	// Night should be noticeably darker but not completely black
	if nightIntensity >= dayIntensity*0.8 {
		t.Errorf("night (%.2f) should be significantly darker than day (%.2f)", nightIntensity, dayIntensity)
	}
}

func TestApplyLighting(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetTimeOfDay(12.0)

	original := color.RGBA{200, 150, 100, 255}
	lit := s.ApplyLighting(original, 5, 5, 0)

	// At noon, color should still be visible
	if lit.R == 0 && lit.G == 0 && lit.B == 0 {
		t.Error("lit pixel should not be black at noon")
	}
	if lit.A != 255 {
		t.Error("alpha should be preserved")
	}
}

func TestApplyLightingClamp(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetTimeOfDay(12.0)

	// Add many bright lights
	for i := 0; i < 5; i++ {
		l := NewPointLight(5, 5, 0, 2.0, color.RGBA{255, 255, 255, 255})
		s.AddLight(l)
	}

	original := color.RGBA{200, 150, 100, 255}
	lit := s.ApplyLighting(original, 5, 5, 0)

	// Should be clamped to 255
	if lit.R > 255 || lit.G > 255 || lit.B > 255 {
		t.Error("colors should be clamped to 255")
	}
}

func TestCreateTorch(t *testing.T) {
	s := NewSystem("fantasy")
	l := s.CreateTorch(5, 5, 1)
	if l == nil {
		t.Fatal("CreateTorch returned nil")
	}
	if l.Type != TypePoint {
		t.Error("torch should be point light")
	}
	if l.Color != s.genrePalette.Torchlight {
		t.Error("torch should use genre torchlight color")
	}
}

func TestCreateMagicLight(t *testing.T) {
	s := NewSystem("fantasy")
	l := s.CreateMagicLight(5, 5, 1, 1.5)
	if l == nil {
		t.Fatal("CreateMagicLight returned nil")
	}
	if l.Intensity != 1.5 {
		t.Errorf("expected intensity 1.5, got %f", l.Intensity)
	}
	if l.Color != s.genrePalette.Magic {
		t.Error("magic light should use genre magic color")
	}
}

func TestAllGenrePalettes(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			palette, ok := GenrePalettes[genre]
			if !ok {
				t.Errorf("missing palette for genre %s", genre)
				return
			}

			// Verify palette has valid colors (non-zero)
			if palette.Sunlight.A == 0 {
				t.Error("sunlight alpha should be non-zero")
			}
			if palette.Moonlight.A == 0 {
				t.Error("moonlight alpha should be non-zero")
			}
			if palette.Torchlight.A == 0 {
				t.Error("torchlight alpha should be non-zero")
			}
		})
	}
}

func TestSunPositionAtDifferentTimes(t *testing.T) {
	s := NewSystem("fantasy")

	// Test that sun intensity varies with time
	s.SetTimeOfDay(12.0)
	noonIntensity := s.sun.Intensity

	s.SetTimeOfDay(0.0) // Midnight
	midnightIntensity := s.sun.Intensity

	if midnightIntensity >= noonIntensity {
		t.Errorf("sun intensity at midnight (%.2f) should be less than noon (%.2f)", midnightIntensity, noonIntensity)
	}

	// Verify sun is using moonlight color at night
	s.SetTimeOfDay(0.0)
	if s.sun.Color != s.genrePalette.Moonlight {
		t.Error("sun should use moonlight color at night")
	}

	// Verify sun is using sunlight color at day
	s.SetTimeOfDay(12.0)
	if s.sun.Color != s.genrePalette.Sunlight {
		t.Error("sun should use sunlight color at noon")
	}
}

func TestDisabledLightNoContribution(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetTimeOfDay(12.0)
	s.ambient.Enabled = false
	s.sun.Enabled = false

	l := NewPointLight(5, 5, 0, 1.0, color.RGBA{255, 255, 255, 255})
	l.Enabled = false
	s.AddLight(l)

	intensity, _ := s.CalculateLightingAt(5, 5, 0)
	if intensity > 0.01 {
		t.Errorf("disabled lights should contribute no intensity, got %f", intensity)
	}
}

func BenchmarkCalculateLightingAt(b *testing.B) {
	s := NewSystem("fantasy")
	for i := 0; i < 10; i++ {
		s.AddLight(NewPointLight(float64(i), float64(i), 0, 1.0, color.RGBA{255, 200, 100, 255}))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.CalculateLightingAt(5, 5, 0)
	}
}

func BenchmarkApplyLighting(b *testing.B) {
	s := NewSystem("fantasy")
	for i := 0; i < 10; i++ {
		s.AddLight(NewPointLight(float64(i), float64(i), 0, 1.0, color.RGBA{255, 200, 100, 255}))
	}
	c := color.RGBA{200, 150, 100, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ApplyLighting(c, 5, 5, 0)
	}
}
