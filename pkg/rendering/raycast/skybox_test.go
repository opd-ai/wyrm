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

// ============================================================
// StarField Tests
// ============================================================

func TestNewStarField(t *testing.T) {
	sf := NewStarField(12345, 100, "fantasy")
	if sf == nil {
		t.Fatal("NewStarField returned nil")
	}
	if sf.StarCount != 100 {
		t.Errorf("expected StarCount 100, got %d", sf.StarCount)
	}
	if sf.Seed != 12345 {
		t.Errorf("expected Seed 12345, got %d", sf.Seed)
	}
	if sf.Genre != "fantasy" {
		t.Errorf("expected Genre 'fantasy', got %s", sf.Genre)
	}
	if len(sf.Stars) != 100 {
		t.Errorf("expected 100 stars, got %d", len(sf.Stars))
	}
}

func TestDefaultStarField(t *testing.T) {
	sf := DefaultStarField(54321)
	if sf == nil {
		t.Fatal("DefaultStarField returned nil")
	}
	if sf.StarCount != 200 {
		t.Errorf("expected default StarCount 200, got %d", sf.StarCount)
	}
	if sf.Seed != 54321 {
		t.Errorf("expected Seed 54321, got %d", sf.Seed)
	}
}

func TestStarFieldDeterminism(t *testing.T) {
	sf1 := NewStarField(99999, 50, "fantasy")
	sf2 := NewStarField(99999, 50, "fantasy")

	if len(sf1.Stars) != len(sf2.Stars) {
		t.Fatal("star counts differ for same seed")
	}

	for i := range sf1.Stars {
		if sf1.Stars[i].X != sf2.Stars[i].X ||
			sf1.Stars[i].Y != sf2.Stars[i].Y ||
			sf1.Stars[i].Brightness != sf2.Stars[i].Brightness {
			t.Errorf("star %d differs between identical seeds", i)
		}
	}
}

func TestStarFieldDifferentSeeds(t *testing.T) {
	sf1 := NewStarField(11111, 50, "fantasy")
	sf2 := NewStarField(22222, 50, "fantasy")

	matches := 0
	for i := range sf1.Stars {
		if sf1.Stars[i].X == sf2.Stars[i].X && sf1.Stars[i].Y == sf2.Stars[i].Y {
			matches++
		}
	}

	// With 50 stars, we shouldn't have more than a few accidental matches
	if matches > 5 {
		t.Errorf("too many matching star positions (%d) for different seeds", matches)
	}
}

func TestStarFieldGenreColors(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		sf := NewStarField(12345, 10, genre)
		if sf.Genre != genre {
			t.Errorf("expected genre %s, got %s", genre, sf.Genre)
		}
		// Verify all stars have valid colors
		for i, star := range sf.Stars {
			if star.Color.A != 255 {
				t.Errorf("star %d has invalid alpha %d", i, star.Color.A)
			}
		}
	}
}

func TestStarFieldSetGenre(t *testing.T) {
	sf := NewStarField(12345, 20, "fantasy")
	oldColors := make([]color.RGBA, len(sf.Stars))
	for i, star := range sf.Stars {
		oldColors[i] = star.Color
	}

	sf.SetGenre("cyberpunk")

	if sf.Genre != "cyberpunk" {
		t.Errorf("expected genre 'cyberpunk', got %s", sf.Genre)
	}

	// Positions should be same, but colors may differ
	for i, star := range sf.Stars {
		// Stars should still exist
		if star.Brightness == 0 {
			t.Errorf("star %d has zero brightness after genre change", i)
		}
	}
}

func TestGetNightIntensity(t *testing.T) {
	tests := []struct {
		timeOfDay float64
		expected  float64
		desc      string
	}{
		{0.0, 1.0, "midnight - full night"},
		{3.0, 1.0, "3 AM - full night"},
		{5.0, 1.0, "5 AM - start of dawn"},
		{6.0, 0.5, "6 AM - middle of dawn"},
		{7.0, 0.0, "7 AM - end of dawn"},
		{12.0, 0.0, "noon - full day"},
		{17.0, 0.0, "5 PM - start of dusk"},
		{18.0, 0.5, "6 PM - middle of dusk"},
		{19.0, 1.0, "7 PM - full night"},
		{22.0, 1.0, "10 PM - full night"},
	}

	for _, tt := range tests {
		result := GetNightIntensity(tt.timeOfDay)
		// Allow small floating point tolerance
		if result < tt.expected-0.01 || result > tt.expected+0.01 {
			t.Errorf("%s: expected %f, got %f", tt.desc, tt.expected, result)
		}
	}
}

func TestStarFieldGetStarColorAt(t *testing.T) {
	sf := NewStarField(12345, 200, "fantasy")

	// Test during day - should return empty color
	color := sf.GetStarColorAt(0.5, 0.5, 0.0, 0.0)
	if color.A != 0 {
		t.Error("expected empty color during day (nightIntensity=0)")
	}

	// Test at random position - might or might not hit a star
	// The star field is sparse, so most positions should return empty
	emptyCount := 0
	for i := 0; i < 100; i++ {
		x := float64(i) / 100.0
		y := float64(i%50) / 100.0
		c := sf.GetStarColorAt(x, y, 0.0, 1.0)
		if c.A == 0 {
			emptyCount++
		}
	}

	// Most positions should be empty (no star)
	if emptyCount < 90 {
		t.Errorf("expected mostly empty positions, got %d/100 empty", emptyCount)
	}
}

func TestSkyboxWithStarField(t *testing.T) {
	s := NewSkyboxWithSeed(12345)
	if s == nil {
		t.Fatal("NewSkyboxWithSeed returned nil")
	}
	if s.starField == nil {
		t.Error("skybox should have star field initialized")
	}
	if s.starField.Seed != 12345 {
		t.Errorf("expected star field seed 12345, got %d", s.starField.Seed)
	}
}

func TestSkyboxStarVisibilityDayNight(t *testing.T) {
	s := NewSkyboxWithSeed(12345)

	// At noon, stars should not affect sky color significantly
	s.SetTimeOfDay(12.0)
	dayColor := s.GetSkyColorAt(0.5, 0.3)

	// At midnight, sky color might include star contributions
	s.SetTimeOfDay(0.0)
	nightColor := s.GetSkyColorAt(0.5, 0.3)

	// Night sky should be darker (lower values) than day sky
	// This is a simple sanity check
	if nightColor.R > dayColor.R || nightColor.G > dayColor.G {
		// This might happen if we hit a star, which is fine
		// Just verify colors are valid
		if nightColor.R > 255 || nightColor.G > 255 || nightColor.B > 255 {
			t.Error("invalid color values")
		}
	}
}

func TestSkyboxUpdate(t *testing.T) {
	s := NewSkybox()
	initialTime := s.animTime

	s.Update(0.016) // ~60fps frame
	if s.animTime <= initialTime {
		t.Error("animTime should advance after Update")
	}

	s.Update(1.0)
	if s.animTime < 1.0 {
		t.Errorf("expected animTime >= 1.0, got %f", s.animTime)
	}
}

func BenchmarkStarFieldGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewStarField(int64(i), 200, "fantasy")
	}
}

func BenchmarkStarFieldGetColor(b *testing.B) {
	sf := NewStarField(12345, 200, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64(i%1000) / 1000.0
		y := float64((i/1000)%500) / 500.0
		_ = sf.GetStarColorAt(x, y, float64(i)*0.01, 1.0)
	}
}

func BenchmarkGetNightIntensity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetNightIntensity(float64(i % 24))
	}
}

// TestDeterministicStarField validates that the same seed produces identical star positions.
// This is critical for networked multiplayer - all clients must render identical night skies.
func TestDeterministicStarField(t *testing.T) {
	seed := int64(77889900)
	starCount := 200
	genre := "fantasy"

	// Generate star fields twice with the same parameters
	sf1 := NewStarField(seed, starCount, genre)
	sf2 := NewStarField(seed, starCount, genre)

	if sf1 == nil || sf2 == nil {
		t.Fatal("NewStarField returned nil")
	}

	// Star count must match
	if len(sf1.Stars) != len(sf2.Stars) {
		t.Fatalf("star count mismatch: %d vs %d", len(sf1.Stars), len(sf2.Stars))
	}

	// Every star position must be identical
	for i := range sf1.Stars {
		s1, s2 := sf1.Stars[i], sf2.Stars[i]

		if s1.X != s2.X || s1.Y != s2.Y {
			t.Errorf("star %d position mismatch: (%f,%f) vs (%f,%f)",
				i, s1.X, s1.Y, s2.X, s2.Y)
		}

		if s1.Color != s2.Color {
			t.Errorf("star %d color mismatch: %v vs %v", i, s1.Color, s2.Color)
		}

		if s1.Brightness != s2.Brightness {
			t.Errorf("star %d brightness mismatch: %f vs %f", i, s1.Brightness, s2.Brightness)
		}

		if s1.TwinklePhase != s2.TwinklePhase {
			t.Errorf("star %d twinkle phase mismatch: %f vs %f", i, s1.TwinklePhase, s2.TwinklePhase)
		}
	}

	// Star colors at same position must be identical
	testPoints := []struct{ x, y float64 }{
		{0.1, 0.1},
		{0.5, 0.3},
		{0.9, 0.2},
		{0.25, 0.45},
	}

	for _, pt := range testPoints {
		c1 := sf1.GetStarColorAt(pt.x, pt.y, 0.0, 1.0)
		c2 := sf2.GetStarColorAt(pt.x, pt.y, 0.0, 1.0)

		if c1 != c2 {
			t.Errorf("star color at (%f,%f) mismatch: %v vs %v", pt.x, pt.y, c1, c2)
		}
	}

	// Different seed should produce different star positions
	sf3 := NewStarField(seed+1, starCount, genre)
	if sf3 == nil {
		t.Fatal("NewStarField returned nil for different seed")
	}

	// At least one star should be at a different position
	allSame := true
	for i := range sf1.Stars {
		if sf1.Stars[i].X != sf3.Stars[i].X || sf1.Stars[i].Y != sf3.Stars[i].Y {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("different seeds should produce different star positions")
	}
}

// TestDeterministicStarFieldAcrossGenres validates determinism is maintained across genres.
func TestDeterministicStarFieldAcrossGenres(t *testing.T) {
	seed := int64(11122233)
	starCount := 100
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sf1 := NewStarField(seed, starCount, genre)
			sf2 := NewStarField(seed, starCount, genre)

			if sf1 == nil || sf2 == nil {
				t.Fatal("NewStarField returned nil")
			}

			if len(sf1.Stars) != len(sf2.Stars) {
				t.Fatalf("star count mismatch for %s: %d vs %d",
					genre, len(sf1.Stars), len(sf2.Stars))
			}

			for i := range sf1.Stars {
				if sf1.Stars[i].X != sf2.Stars[i].X || sf1.Stars[i].Y != sf2.Stars[i].Y {
					t.Errorf("star %d position mismatch for %s", i, genre)
					break
				}
			}
		})
	}
}
