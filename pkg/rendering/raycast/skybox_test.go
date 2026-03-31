//go:build noebiten

package raycast

import (
	"image/color"
	"testing"
)

func TestNewSkyConfig(t *testing.T) {
	cfg := NewSkyConfig()
	if cfg == nil {
		t.Fatal("NewSkyConfig returned nil")
	}
	if cfg.TimeOfDay != 12.0 {
		t.Errorf("expected TimeOfDay 12.0, got %f", cfg.TimeOfDay)
	}
	if cfg.Genre != "fantasy" {
		t.Errorf("expected Genre 'fantasy', got %s", cfg.Genre)
	}
	if cfg.CloudCover != 0.3 {
		t.Errorf("expected CloudCover 0.3, got %f", cfg.CloudCover)
	}
	if cfg.WeatherType != "clear" {
		t.Errorf("expected WeatherType 'clear', got %s", cfg.WeatherType)
	}
	if cfg.Indoor {
		t.Error("expected Indoor to be false")
	}
}

func TestNewSkybox(t *testing.T) {
	s := NewSkybox()
	if s == nil {
		t.Fatal("NewSkybox returned nil")
	}
	if s.config == nil {
		t.Error("skybox config should be initialized")
	}
	// Check colors are initialized
	if s.zenithColor == (color.RGBA{}) {
		t.Error("zenith color should be initialized")
	}
	if s.horizonColor == (color.RGBA{}) {
		t.Error("horizon color should be initialized")
	}
}

func TestSkyboxSetTimeOfDay(t *testing.T) {
	s := NewSkybox()

	tests := []struct {
		input    float64
		expected float64
	}{
		{12.0, 12.0},
		{0.0, 0.0},
		{24.0, 0.0},
		{25.0, 1.0},
		{-1.0, 23.0},
		{36.0, 12.0},
	}

	for _, tt := range tests {
		s.SetTimeOfDay(tt.input)
		if s.config.TimeOfDay != tt.expected {
			t.Errorf("SetTimeOfDay(%f) = %f, want %f", tt.input, s.config.TimeOfDay, tt.expected)
		}
	}
}

func TestSkyboxSetGenre(t *testing.T) {
	s := NewSkybox()
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		s.SetGenre(genre)
		if s.config.Genre != genre {
			t.Errorf("expected genre %s, got %s", genre, s.config.Genre)
		}
		// Colors should be updated
		if s.zenithColor == (color.RGBA{}) {
			t.Errorf("zenith color not updated for genre %s", genre)
		}
	}
}

func TestSkyboxSetWeather(t *testing.T) {
	s := NewSkybox()

	s.SetWeather("rain", 0.8)
	if s.config.WeatherType != "rain" {
		t.Errorf("expected weather 'rain', got %s", s.config.WeatherType)
	}
	if s.config.CloudCover != 0.8 {
		t.Errorf("expected cloud cover 0.8, got %f", s.config.CloudCover)
	}

	// Test clamping
	s.SetWeather("storm", 1.5)
	if s.config.CloudCover != 1.0 {
		t.Errorf("cloud cover should clamp to 1.0, got %f", s.config.CloudCover)
	}

	s.SetWeather("clear", -0.5)
	if s.config.CloudCover != 0.0 {
		t.Errorf("cloud cover should clamp to 0.0, got %f", s.config.CloudCover)
	}
}

func TestSkyboxIndoorMode(t *testing.T) {
	s := NewSkybox()

	if s.IsIndoor() {
		t.Error("should not be indoor by default")
	}

	s.SetIndoor(true)
	if !s.IsIndoor() {
		t.Error("should be indoor after SetIndoor(true)")
	}

	s.SetIndoor(false)
	if s.IsIndoor() {
		t.Error("should not be indoor after SetIndoor(false)")
	}
}

func TestSkyboxIsDaytime(t *testing.T) {
	s := NewSkybox()

	dayTimes := []float64{6.0, 9.0, 12.0, 15.0, 17.99}
	for _, hour := range dayTimes {
		s.SetTimeOfDay(hour)
		if !s.isDaytime() {
			t.Errorf("time %f should be daytime", hour)
		}
	}

	nightTimes := []float64{0.0, 3.0, 5.99, 18.0, 21.0, 23.0}
	for _, hour := range nightTimes {
		s.SetTimeOfDay(hour)
		if s.isDaytime() {
			t.Errorf("time %f should be nighttime", hour)
		}
	}
}

func TestSkyboxGetSkyColorAt(t *testing.T) {
	s := NewSkybox()
	s.SetTimeOfDay(12.0) // Noon

	// Test zenith (top of sky)
	zenithColor := s.GetSkyColorAt(0.5, 0.0)
	if zenithColor.A != 255 {
		t.Error("sky color should have full alpha")
	}

	// Test horizon (bottom of sky portion)
	horizonColor := s.GetSkyColorAt(0.5, 1.0)
	if horizonColor.A != 255 {
		t.Error("horizon color should have full alpha")
	}

	// Horizon should be different from zenith
	if zenithColor == horizonColor {
		t.Error("zenith and horizon colors should differ")
	}
}

func TestSkyboxGetSkyColorAtIndoor(t *testing.T) {
	s := NewSkybox()
	s.SetIndoor(true)

	// Indoor should return consistent ceiling color regardless of position
	color1 := s.GetSkyColorAt(0.0, 0.0)
	color2 := s.GetSkyColorAt(0.5, 0.5)
	color3 := s.GetSkyColorAt(1.0, 1.0)

	if color1 != color2 || color2 != color3 {
		t.Error("indoor ceiling should return same color for all positions")
	}
}

func TestSkyboxGetHorizonColor(t *testing.T) {
	s := NewSkybox()

	// Test outdoor
	outdoor := s.GetHorizonColor()
	if outdoor.A != 255 {
		t.Error("horizon color should have full alpha")
	}

	// Test indoor
	s.SetIndoor(true)
	indoor := s.GetHorizonColor()
	if indoor != s.getIndoorCeiling() {
		t.Error("indoor horizon should return ceiling color")
	}
}

func TestSkyboxGetZenithColor(t *testing.T) {
	s := NewSkybox()

	zenith := s.GetZenithColor()
	if zenith.A != 255 {
		t.Error("zenith color should have full alpha")
	}

	// Test indoor
	s.SetIndoor(true)
	indoor := s.GetZenithColor()
	if indoor != s.getIndoorCeiling() {
		t.Error("indoor zenith should return ceiling color")
	}
}

func TestSkyboxCelestialPositions(t *testing.T) {
	s := NewSkybox()

	// At noon, sun should be high (low Y) and centered (X=0.5)
	s.SetTimeOfDay(12.0)
	sunX, sunY := s.GetSunPosition()
	if sunY > 0.2 {
		t.Errorf("sun at noon should be high (low Y), got Y=%f", sunY)
	}
	if sunX < 0.45 || sunX > 0.55 {
		t.Errorf("sun at noon should be centered (X~0.5), got X=%f", sunX)
	}

	// At midnight, moon should be high and centered
	s.SetTimeOfDay(0.0)
	moonX, moonY := s.GetMoonPosition()
	if moonY > 0.2 {
		t.Errorf("moon at midnight should be high, got Y=%f", moonY)
	}
	if moonX < 0.45 || moonX > 0.55 {
		t.Errorf("moon at midnight should be centered, got X=%f", moonX)
	}
}

func TestSkyboxBlendColors(t *testing.T) {
	s := NewSkybox()

	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// t=0 should be black
	result := s.blendColors(black, white, 0)
	if result != black {
		t.Errorf("blend(black, white, 0) should be black, got %v", result)
	}

	// t=1 should be white
	result = s.blendColors(black, white, 1)
	if result.R != 255 || result.G != 255 || result.B != 255 {
		t.Errorf("blend(black, white, 1) should be white, got %v", result)
	}

	// t=0.5 should be mid-gray
	result = s.blendColors(black, white, 0.5)
	if result.R < 125 || result.R > 130 {
		t.Errorf("blend(black, white, 0.5) should be ~127, got R=%d", result.R)
	}

	// Clamping test
	result = s.blendColors(black, white, 1.5)
	if result.R != 255 {
		t.Errorf("blend should clamp t > 1, got R=%d", result.R)
	}

	result = s.blendColors(black, white, -0.5)
	if result.R != 0 {
		t.Errorf("blend should clamp t < 0, got R=%d", result.R)
	}
}

func TestSkyboxApplyWeatherEffects(t *testing.T) {
	s := NewSkybox()
	base := color.RGBA{R: 100, G: 150, B: 200, A: 255}

	weatherTypes := []string{"clear", "overcast", "rain", "storm", "snow", "fog"}

	for _, weather := range weatherTypes {
		s.SetWeather(weather, 0.5)
		result := s.applyWeatherEffects(base, 0.5, 0.5)

		// Should return valid color
		if result.A != 255 {
			t.Errorf("weather %s: should have full alpha", weather)
		}

		// Clear weather should return base unchanged
		if weather == "clear" && result != base {
			t.Errorf("clear weather should not modify base color")
		}

		// Other weather should modify base
		if weather != "clear" && result == base {
			t.Errorf("weather %s should modify base color", weather)
		}
	}
}

func TestSkyboxGetIndoorCeiling(t *testing.T) {
	s := NewSkybox()
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		s.SetGenre(genre)
		ceiling := s.getIndoorCeiling()

		if ceiling.A != 255 {
			t.Errorf("genre %s: ceiling should have full alpha", genre)
		}

		// Should be a dark/neutral color
		if ceiling.R > 100 || ceiling.G > 100 || ceiling.B > 100 {
			t.Errorf("genre %s: indoor ceiling should be dark, got %v", genre, ceiling)
		}
	}
}

func TestSkyboxGetGenrePalette(t *testing.T) {
	s := NewSkybox()
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		s.SetGenre(genre)
		palette := s.getGenrePalette()

		// Check all palette colors are set
		if palette.DayZenith == (color.RGBA{}) {
			t.Errorf("genre %s: DayZenith should be set", genre)
		}
		if palette.DayHorizon == (color.RGBA{}) {
			t.Errorf("genre %s: DayHorizon should be set", genre)
		}
		if palette.NightZenith == (color.RGBA{}) {
			t.Errorf("genre %s: NightZenith should be set", genre)
		}
		if palette.Sun == (color.RGBA{}) {
			t.Errorf("genre %s: Sun should be set", genre)
		}
		if palette.Moon == (color.RGBA{}) {
			t.Errorf("genre %s: Moon should be set", genre)
		}
	}
}

func TestSkyboxTimeTransitions(t *testing.T) {
	s := NewSkybox()

	// Test that colors change smoothly through day
	var prevZenith color.RGBA
	changes := 0

	for hour := 0.0; hour < 24.0; hour += 1.0 {
		s.SetTimeOfDay(hour)
		if s.zenithColor != prevZenith {
			changes++
			prevZenith = s.zenithColor
		}
	}

	// Should have multiple color transitions
	if changes < 4 {
		t.Errorf("expected at least 4 color transitions, got %d", changes)
	}
}

func TestSkyboxSetConfig(t *testing.T) {
	s := NewSkybox()

	// Test nil config
	s.SetConfig(nil)
	// Should not panic or change state

	// Test valid config
	cfg := &SkyConfig{
		TimeOfDay:   18.0,
		Genre:       "horror",
		CloudCover:  0.9,
		WeatherType: "storm",
		Indoor:      true,
	}

	s.SetConfig(cfg)
	if s.config.TimeOfDay != 18.0 {
		t.Errorf("expected TimeOfDay 18.0, got %f", s.config.TimeOfDay)
	}
	if s.config.Genre != "horror" {
		t.Errorf("expected Genre 'horror', got %s", s.config.Genre)
	}
	if s.config.Indoor != true {
		t.Error("expected Indoor true")
	}
}

func TestSkyboxGetConfig(t *testing.T) {
	s := NewSkybox()
	cfg := s.GetConfig()

	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}
	if cfg != s.config {
		t.Error("GetConfig should return the same config instance")
	}
}

func TestSkyboxGetTimeOfDay(t *testing.T) {
	s := NewSkybox()

	s.SetTimeOfDay(15.5)
	if s.GetTimeOfDay() != 15.5 {
		t.Errorf("expected GetTimeOfDay 15.5, got %f", s.GetTimeOfDay())
	}
}

func TestSkyboxAddGlow(t *testing.T) {
	s := NewSkybox()

	base := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	glow := color.RGBA{R: 255, G: 200, B: 0, A: 255}

	// No glow
	result := s.addGlow(base, glow, 0)
	if result != base {
		t.Error("addGlow with 0 intensity should return base")
	}

	// Full glow should add without overflow
	result = s.addGlow(base, glow, 1.0)
	if result.R != 255 {
		t.Errorf("expected R clamped to 255, got %d", result.R)
	}
	if result.G != 255 {
		t.Errorf("expected G clamped to 255, got %d", result.G)
	}

	// Partial glow
	result = s.addGlow(base, glow, 0.5)
	if result.R <= 100 || result.R >= 255 {
		t.Errorf("partial glow should add some R, got %d", result.R)
	}
}

func BenchmarkSkyboxGetSkyColorAt(b *testing.B) {
	s := NewSkybox()
	s.SetTimeOfDay(12.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64(i%1000) / 1000.0
		y := float64((i/1000)%500) / 500.0
		_ = s.GetSkyColorAt(x, y)
	}
}

func BenchmarkSkyboxSetTimeOfDay(b *testing.B) {
	s := NewSkybox()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SetTimeOfDay(float64(i % 24))
	}
}
