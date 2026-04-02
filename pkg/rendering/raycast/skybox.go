// Package raycast provides DDA-based raycasting rendering.
//
// This file implements procedural skybox rendering with:
// - Genre-specific sky colors and atmospheres
// - Time-of-day simulation with sun/moon positioning
// - Weather integration (clouds, precipitation effects)
// - Horizon blending for seamless terrain transitions

package raycast

import (
	"image/color"
	"math"
)

// SkyConfig holds configuration for skybox rendering.
type SkyConfig struct {
	// Time of day (0-24 hours)
	TimeOfDay float64

	// Genre affects color palette and atmosphere
	Genre string

	// Weather conditions
	CloudCover  float64 // 0-1, amount of cloud coverage
	WeatherType string  // clear, overcast, rain, storm, snow

	// Indoor flag disables skybox rendering
	Indoor bool
}

// NewSkyConfig creates a default sky configuration.
func NewSkyConfig() *SkyConfig {
	return &SkyConfig{
		TimeOfDay:   12.0,
		Genre:       "fantasy",
		CloudCover:  0.3,
		WeatherType: "clear",
		Indoor:      false,
	}
}

// Skybox handles procedural sky rendering.
type Skybox struct {
	config *SkyConfig

	// Cached colors for current state
	zenithColor  color.RGBA
	horizonColor color.RGBA
	sunColor     color.RGBA
	moonColor    color.RGBA

	// Sun/moon position (normalized screen coordinates)
	sunX, sunY   float64
	moonX, moonY float64

	// Rendering parameters
	sunRadius  float64
	moonRadius float64

	// Star field for nighttime sky
	starField *StarField

	// Animation time for star twinkle
	animTime float64
}

// NewSkybox creates a new skybox renderer.
func NewSkybox() *Skybox {
	s := &Skybox{
		config:     NewSkyConfig(),
		sunRadius:  0.05,
		moonRadius: 0.03,
		starField:  DefaultStarField(12345), // Default seed
		animTime:   0,
	}
	s.updateColors()
	s.updateCelestialPositions()
	return s
}

// NewSkyboxWithSeed creates a new skybox renderer with a specific seed for stars.
func NewSkyboxWithSeed(seed int64) *Skybox {
	s := &Skybox{
		config:     NewSkyConfig(),
		sunRadius:  0.05,
		moonRadius: 0.03,
		starField:  DefaultStarField(seed),
		animTime:   0,
	}
	s.updateColors()
	s.updateCelestialPositions()
	return s
}

// SetConfig updates the sky configuration.
func (s *Skybox) SetConfig(cfg *SkyConfig) {
	if cfg != nil {
		s.config = cfg
		s.updateColors()
		s.updateCelestialPositions()
	}
}

// SetTimeOfDay updates the time and recalculates sky state.
func (s *Skybox) SetTimeOfDay(t float64) {
	s.config.TimeOfDay = math.Mod(t, 24.0)
	if s.config.TimeOfDay < 0 {
		s.config.TimeOfDay += 24.0
	}
	s.updateColors()
	s.updateCelestialPositions()
}

// SetGenre updates the genre palette and star field colors.
func (s *Skybox) SetGenre(genre string) {
	s.config.Genre = genre
	s.updateColors()
	if s.starField != nil {
		s.starField.SetGenre(genre)
	}
}

// Update advances animation time for star twinkle effects.
// dt is the time since the last update in seconds.
func (s *Skybox) Update(dt float64) {
	s.animTime += dt
}

// SetWeather updates weather conditions.
func (s *Skybox) SetWeather(weatherType string, cloudCover float64) {
	s.config.WeatherType = weatherType
	s.config.CloudCover = math.Max(0, math.Min(1, cloudCover))
	s.updateColors()
}

// SetIndoor toggles indoor mode (disables sky rendering).
func (s *Skybox) SetIndoor(indoor bool) {
	s.config.Indoor = indoor
}

// IsIndoor returns whether indoor mode is active.
func (s *Skybox) IsIndoor() bool {
	return s.config.Indoor
}

// GetSkyColorAt returns the sky color for a given screen position.
// x is normalized screen X (0-1), y is vertical position (0=top, 1=horizon).
func (s *Skybox) GetSkyColorAt(x, y float64) color.RGBA {
	if s.config.Indoor {
		return s.getIndoorCeiling()
	}

	// Base gradient from zenith to horizon
	baseColor := s.blendColors(s.zenithColor, s.horizonColor, y)

	// Add stars during night (before celestial bodies so they don't overlap)
	if s.starField != nil {
		nightIntensity := GetNightIntensity(s.config.TimeOfDay)
		if nightIntensity > 0 {
			starColor := s.starField.GetStarColorAt(x, y, s.animTime, nightIntensity)
			if starColor.A > 0 {
				// Blend star onto sky background
				baseColor = s.addGlow(baseColor, starColor, 1.0)
			}
		}
	}

	// Add celestial bodies
	baseColor = s.addCelestialBody(baseColor, x, y, s.sunX, s.sunY,
		s.sunRadius, s.sunColor, s.isDaytime())
	baseColor = s.addCelestialBody(baseColor, x, y, s.moonX, s.moonY,
		s.moonRadius, s.moonColor, !s.isDaytime())

	// Apply weather effects
	baseColor = s.applyWeatherEffects(baseColor, x, y)

	return baseColor
}

// GetHorizonColor returns the current horizon color for terrain blending.
func (s *Skybox) GetHorizonColor() color.RGBA {
	if s.config.Indoor {
		return s.getIndoorCeiling()
	}
	return s.horizonColor
}

// GetZenithColor returns the current zenith (overhead) color.
func (s *Skybox) GetZenithColor() color.RGBA {
	if s.config.Indoor {
		return s.getIndoorCeiling()
	}
	return s.zenithColor
}

// GetSunPosition returns normalized sun screen coordinates.
func (s *Skybox) GetSunPosition() (x, y float64) {
	return s.sunX, s.sunY
}

// GetMoonPosition returns normalized moon screen coordinates.
func (s *Skybox) GetMoonPosition() (x, y float64) {
	return s.moonX, s.moonY
}

// isDaytime returns true if it's currently daytime (6-18).
func (s *Skybox) isDaytime() bool {
	return s.config.TimeOfDay >= 6 && s.config.TimeOfDay < 18
}

// updateColors recalculates sky colors based on time, genre, weather.
func (s *Skybox) updateColors() {
	palette := s.getGenrePalette()
	t := s.config.TimeOfDay

	// Determine phase: night, dawn, day, dusk
	var zenith, horizon, sun, moon color.RGBA

	switch {
	case t < 5: // Night
		zenith = palette.NightZenith
		horizon = palette.NightHorizon
		sun = palette.Sun
		moon = palette.Moon
	case t < 7: // Dawn
		progress := (t - 5) / 2
		zenith = s.blendColors(palette.NightZenith, palette.DawnZenith, progress)
		horizon = s.blendColors(palette.NightHorizon, palette.DawnHorizon, progress)
		sun = palette.Sun
		moon = palette.Moon
	case t < 9: // Morning transition
		progress := (t - 7) / 2
		zenith = s.blendColors(palette.DawnZenith, palette.DayZenith, progress)
		horizon = s.blendColors(palette.DawnHorizon, palette.DayHorizon, progress)
		sun = palette.Sun
		moon = palette.Moon
	case t < 17: // Day
		zenith = palette.DayZenith
		horizon = palette.DayHorizon
		sun = palette.Sun
		moon = palette.Moon
	case t < 19: // Dusk
		progress := (t - 17) / 2
		zenith = s.blendColors(palette.DayZenith, palette.DuskZenith, progress)
		horizon = s.blendColors(palette.DayHorizon, palette.DuskHorizon, progress)
		sun = palette.Sun
		moon = palette.Moon
	case t < 21: // Evening transition
		progress := (t - 19) / 2
		zenith = s.blendColors(palette.DuskZenith, palette.NightZenith, progress)
		horizon = s.blendColors(palette.DuskHorizon, palette.NightHorizon, progress)
		sun = palette.Sun
		moon = palette.Moon
	default: // Night
		zenith = palette.NightZenith
		horizon = palette.NightHorizon
		sun = palette.Sun
		moon = palette.Moon
	}

	s.zenithColor = zenith
	s.horizonColor = horizon
	s.sunColor = sun
	s.moonColor = moon
}

// updateCelestialPositions calculates sun and moon screen positions.
func (s *Skybox) updateCelestialPositions() {
	t := s.config.TimeOfDay

	// Sun: rises at 6, peaks at 12, sets at 18
	// X: rises in east (0.1), peaks at center (0.5), sets in west (0.9)
	// Y: 1 at horizon (rise/set), 0 at zenith (noon)
	// Map 6AM-6PM (0-12 hours of daylight) to angle 0-π
	if t >= 6 && t < 18 {
		// Daytime - sun visible
		dayProgress := (t - 6) / 12.0            // 0 at 6AM, 1 at 6PM
		s.sunX = 0.1 + 0.8*dayProgress           // 0.1 to 0.9
		s.sunY = math.Sin(dayProgress * math.Pi) // Parabolic arc, 0 at rise/set, 1 at noon
		s.sunY = 1.0 - s.sunY                    // Invert so 0 is high, 1 is at horizon
	} else {
		// Nighttime - sun below horizon
		s.sunX = 0.5
		s.sunY = 1.5 // Below horizon
	}

	// Moon: opposite of sun (rises at 18, peaks at 0, sets at 6)
	// Map 6PM-6AM (12 hours of nighttime) to X position
	if t >= 18 || t < 6 {
		// Nighttime - moon visible
		var nightProgress float64
		if t >= 18 {
			nightProgress = (t - 18) / 12.0 // 18:00 = 0, 24:00 = 0.5
		} else {
			nightProgress = (t + 6) / 12.0 // 00:00 = 0.5, 06:00 = 1.0
		}
		s.moonX = 0.1 + 0.8*nightProgress
		s.moonY = math.Sin(nightProgress * math.Pi)
		s.moonY = 1.0 - s.moonY
	} else {
		// Daytime - moon below horizon
		s.moonX = 0.5
		s.moonY = 1.5
	}
}

// blendColors linearly interpolates between two colors.
// Uses math.Round for accurate uint8 conversion to prevent ±1 LSB color errors.
func (s *Skybox) blendColors(a, b color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))
	return color.RGBA{
		R: uint8(math.Round(float64(a.R)*(1-t) + float64(b.R)*t)),
		G: uint8(math.Round(float64(a.G)*(1-t) + float64(b.G)*t)),
		B: uint8(math.Round(float64(a.B)*(1-t) + float64(b.B)*t)),
		A: 255,
	}
}

// addCelestialBody adds sun or moon glow to the sky color.
func (s *Skybox) addCelestialBody(base color.RGBA, x, y, bodyX, bodyY,
	radius float64, bodyColor color.RGBA, visible bool,
) color.RGBA {
	if !visible {
		return base
	}

	// Distance from celestial body center
	dx := x - bodyX
	dy := y - bodyY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > radius*3 {
		return base // Too far for any effect
	}

	// Core glow
	if dist < radius {
		// Inside the body - bright center
		intensity := 1.0 - (dist / radius)
		return s.addGlow(base, bodyColor, intensity*0.8)
	}

	// Outer glow (corona)
	glowDist := (dist - radius) / (radius * 2)
	if glowDist < 1.0 {
		intensity := (1.0 - glowDist) * 0.3
		return s.addGlow(base, bodyColor, intensity)
	}

	return base
}

// addGlow adds a glow effect to a base color.
func (s *Skybox) addGlow(base, glow color.RGBA, intensity float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(255, float64(base.R)+float64(glow.R)*intensity)),
		G: uint8(math.Min(255, float64(base.G)+float64(glow.G)*intensity)),
		B: uint8(math.Min(255, float64(base.B)+float64(glow.B)*intensity)),
		A: 255,
	}
}

// applyWeatherEffects modifies sky color based on weather.
func (s *Skybox) applyWeatherEffects(base color.RGBA, x, y float64) color.RGBA {
	switch s.config.WeatherType {
	case "overcast":
		gray := color.RGBA{R: 150, G: 155, B: 160, A: 255}
		return s.blendColors(base, gray, s.config.CloudCover*0.7)

	case "rain":
		gray := color.RGBA{R: 100, G: 105, B: 115, A: 255}
		return s.blendColors(base, gray, s.config.CloudCover*0.8)

	case "storm":
		dark := color.RGBA{R: 50, G: 55, B: 65, A: 255}
		return s.blendColors(base, dark, s.config.CloudCover*0.9)

	case "snow":
		white := color.RGBA{R: 200, G: 205, B: 210, A: 255}
		return s.blendColors(base, white, s.config.CloudCover*0.6)

	case "fog":
		fog := color.RGBA{R: 180, G: 185, B: 190, A: 255}
		// Fog is thicker near horizon
		fogIntensity := s.config.CloudCover * (0.3 + 0.7*y)
		return s.blendColors(base, fog, fogIntensity)

	default: // clear
		return base
	}
}

// getIndoorCeiling returns a neutral ceiling color for indoor spaces.
func (s *Skybox) getIndoorCeiling() color.RGBA {
	switch s.config.Genre {
	case "sci-fi":
		return color.RGBA{R: 40, G: 45, B: 55, A: 255}
	case "horror":
		return color.RGBA{R: 25, G: 20, B: 20, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 30, G: 25, B: 40, A: 255}
	case "post-apocalyptic":
		return color.RGBA{R: 50, G: 45, B: 40, A: 255}
	default: // fantasy
		return color.RGBA{R: 45, G: 40, B: 35, A: 255}
	}
}

// SkyPalette defines colors for different times of day.
type SkyPalette struct {
	NightZenith  color.RGBA
	NightHorizon color.RGBA
	DawnZenith   color.RGBA
	DawnHorizon  color.RGBA
	DayZenith    color.RGBA
	DayHorizon   color.RGBA
	DuskZenith   color.RGBA
	DuskHorizon  color.RGBA
	Sun          color.RGBA
	Moon         color.RGBA
}

// getGenrePalette returns sky colors for the current genre.
func (s *Skybox) getGenrePalette() SkyPalette {
	switch s.config.Genre {
	case "sci-fi":
		return SkyPalette{
			NightZenith:  color.RGBA{R: 5, G: 10, B: 30, A: 255},
			NightHorizon: color.RGBA{R: 20, G: 30, B: 50, A: 255},
			DawnZenith:   color.RGBA{R: 50, G: 80, B: 120, A: 255},
			DawnHorizon:  color.RGBA{R: 100, G: 130, B: 180, A: 255},
			DayZenith:    color.RGBA{R: 100, G: 150, B: 200, A: 255},
			DayHorizon:   color.RGBA{R: 150, G: 180, B: 210, A: 255},
			DuskZenith:   color.RGBA{R: 80, G: 60, B: 120, A: 255},
			DuskHorizon:  color.RGBA{R: 150, G: 100, B: 150, A: 255},
			Sun:          color.RGBA{R: 255, G: 250, B: 230, A: 255},
			Moon:         color.RGBA{R: 200, G: 210, B: 255, A: 255},
		}

	case "horror":
		return SkyPalette{
			NightZenith:  color.RGBA{R: 10, G: 5, B: 15, A: 255},
			NightHorizon: color.RGBA{R: 30, G: 20, B: 35, A: 255},
			DawnZenith:   color.RGBA{R: 60, G: 50, B: 70, A: 255},
			DawnHorizon:  color.RGBA{R: 120, G: 80, B: 90, A: 255},
			DayZenith:    color.RGBA{R: 120, G: 130, B: 140, A: 255},
			DayHorizon:   color.RGBA{R: 150, G: 155, B: 160, A: 255},
			DuskZenith:   color.RGBA{R: 80, G: 50, B: 60, A: 255},
			DuskHorizon:  color.RGBA{R: 140, G: 80, B: 70, A: 255},
			Sun:          color.RGBA{R: 255, G: 200, B: 180, A: 255},
			Moon:         color.RGBA{R: 220, G: 180, B: 180, A: 255},
		}

	case "cyberpunk":
		return SkyPalette{
			NightZenith:  color.RGBA{R: 15, G: 10, B: 30, A: 255},
			NightHorizon: color.RGBA{R: 40, G: 20, B: 60, A: 255},
			DawnZenith:   color.RGBA{R: 80, G: 50, B: 100, A: 255},
			DawnHorizon:  color.RGBA{R: 180, G: 100, B: 150, A: 255},
			DayZenith:    color.RGBA{R: 100, G: 120, B: 160, A: 255},
			DayHorizon:   color.RGBA{R: 150, G: 140, B: 180, A: 255},
			DuskZenith:   color.RGBA{R: 100, G: 60, B: 120, A: 255},
			DuskHorizon:  color.RGBA{R: 200, G: 100, B: 150, A: 255},
			Sun:          color.RGBA{R: 255, G: 200, B: 150, A: 255},
			Moon:         color.RGBA{R: 180, G: 200, B: 255, A: 255},
		}

	case "post-apocalyptic":
		return SkyPalette{
			NightZenith:  color.RGBA{R: 20, G: 15, B: 10, A: 255},
			NightHorizon: color.RGBA{R: 40, G: 30, B: 25, A: 255},
			DawnZenith:   color.RGBA{R: 100, G: 70, B: 50, A: 255},
			DawnHorizon:  color.RGBA{R: 180, G: 120, B: 80, A: 255},
			DayZenith:    color.RGBA{R: 160, G: 140, B: 120, A: 255},
			DayHorizon:   color.RGBA{R: 200, G: 170, B: 140, A: 255},
			DuskZenith:   color.RGBA{R: 140, G: 80, B: 50, A: 255},
			DuskHorizon:  color.RGBA{R: 200, G: 120, B: 70, A: 255},
			Sun:          color.RGBA{R: 255, G: 180, B: 100, A: 255},
			Moon:         color.RGBA{R: 200, G: 180, B: 150, A: 255},
		}

	default: // fantasy
		return SkyPalette{
			NightZenith:  color.RGBA{R: 10, G: 15, B: 40, A: 255},
			NightHorizon: color.RGBA{R: 30, G: 40, B: 70, A: 255},
			DawnZenith:   color.RGBA{R: 80, G: 100, B: 150, A: 255},
			DawnHorizon:  color.RGBA{R: 200, G: 150, B: 120, A: 255},
			DayZenith:    color.RGBA{R: 100, G: 150, B: 220, A: 255},
			DayHorizon:   color.RGBA{R: 180, G: 200, B: 230, A: 255},
			DuskZenith:   color.RGBA{R: 120, G: 80, B: 100, A: 255},
			DuskHorizon:  color.RGBA{R: 220, G: 140, B: 100, A: 255},
			Sun:          color.RGBA{R: 255, G: 240, B: 200, A: 255},
			Moon:         color.RGBA{R: 220, G: 230, B: 255, A: 255},
		}
	}
}

// GetConfig returns the current sky configuration.
func (s *Skybox) GetConfig() *SkyConfig {
	return s.config
}

// GetTimeOfDay returns the current time of day.
func (s *Skybox) GetTimeOfDay() float64 {
	return s.config.TimeOfDay
}

// ============================================================
// Star Field Implementation
// ============================================================

// Star represents a single star in the night sky.
type Star struct {
	// X, Y are normalized sky coordinates (0-1)
	X, Y float64
	// Brightness is the star intensity (0.0-1.0)
	Brightness float64
	// Color is the star's color (typically white or slight blue/yellow tint)
	Color color.RGBA
	// Twinkle phase offset for per-star twinkle variation
	TwinklePhase float64
}

// StarField manages a deterministic set of stars for nighttime sky rendering.
type StarField struct {
	// Stars is the list of generated stars
	Stars []Star
	// Seed used for generation (for determinism)
	Seed int64
	// StarCount is the number of stars to generate
	StarCount int
	// TwinkleSpeed controls how fast stars twinkle
	TwinkleSpeed float64
	// Genre affects star colors and density
	Genre string
}

// NewStarField creates a new deterministic star field from the given seed.
func NewStarField(seed int64, starCount int, genre string) *StarField {
	sf := &StarField{
		Seed:         seed,
		StarCount:    starCount,
		TwinkleSpeed: 2.0,
		Genre:        genre,
	}
	sf.generateStars()
	return sf
}

// DefaultStarField creates a star field with default parameters.
func DefaultStarField(seed int64) *StarField {
	return NewStarField(seed, 200, "fantasy")
}

// generateStars creates the star positions and properties deterministically.
func (sf *StarField) generateStars() {
	sf.Stars = make([]Star, sf.StarCount)

	// Simple linear congruential generator for determinism
	rngState := uint64(sf.Seed)
	nextRand := func() float64 {
		rngState = rngState*6364136223846793005 + 1442695040888963407
		return float64(rngState>>33) / float64(1<<31)
	}

	for i := 0; i < sf.StarCount; i++ {
		// Position: distribute across the sky (avoid horizon)
		x := nextRand()
		y := nextRand() * 0.7 // Stars mostly in upper 70% of sky

		// Brightness: most stars are dim, few are bright
		brightness := nextRand() * nextRand() // Squared for more dim stars
		brightness = 0.3 + brightness*0.7     // Range 0.3-1.0

		// Color based on genre and random variation
		starColor := sf.getStarColor(nextRand(), brightness)

		// Twinkle phase offset
		twinklePhase := nextRand() * 2 * 3.14159

		sf.Stars[i] = Star{
			X:            x,
			Y:            y,
			Brightness:   brightness,
			Color:        starColor,
			TwinklePhase: twinklePhase,
		}
	}
}

// getStarColor returns a star color based on genre and random value.
func (sf *StarField) getStarColor(rand, brightness float64) color.RGBA {
	// Base white with slight color variation
	var r, g, b uint8

	switch sf.Genre {
	case "sci-fi":
		// Blue-white stars
		if rand < 0.3 {
			r, g, b = 200, 220, 255 // Blue
		} else if rand < 0.6 {
			r, g, b = 255, 255, 255 // White
		} else {
			r, g, b = 180, 255, 220 // Cyan-ish
		}
	case "horror":
		// Dim, cold stars
		if rand < 0.5 {
			r, g, b = 180, 180, 200 // Cold white
		} else {
			r, g, b = 200, 150, 150 // Dim reddish
		}
	case "cyberpunk":
		// Neon-tinted stars (light pollution effect)
		if rand < 0.4 {
			r, g, b = 255, 200, 255 // Pink-ish
		} else if rand < 0.7 {
			r, g, b = 200, 255, 255 // Cyan
		} else {
			r, g, b = 255, 255, 200 // Yellow
		}
	case "post-apocalyptic":
		// Orange-tinted (dust in atmosphere)
		if rand < 0.6 {
			r, g, b = 255, 230, 200 // Warm white
		} else {
			r, g, b = 255, 200, 150 // Orange-ish
		}
	default: // fantasy
		// Classic star colors
		if rand < 0.4 {
			r, g, b = 255, 255, 255 // White
		} else if rand < 0.7 {
			r, g, b = 255, 255, 220 // Warm white
		} else {
			r, g, b = 220, 230, 255 // Blue-ish
		}
	}

	// Apply brightness
	br := uint8(float64(r) * brightness)
	bg := uint8(float64(g) * brightness)
	bb := uint8(float64(b) * brightness)

	return color.RGBA{R: br, G: bg, B: bb, A: 255}
}

// GetStarColorAt returns the star contribution at the given sky position.
// Returns black (0,0,0,0) if no star is at this position.
// time is the current animation time for twinkle effects.
func (sf *StarField) GetStarColorAt(x, y, time, nightIntensity float64) color.RGBA {
	if nightIntensity <= 0 {
		return color.RGBA{} // No stars during day
	}

	// Check each star (could optimize with spatial partitioning for large star counts)
	for _, star := range sf.Stars {
		// Distance from star center
		dx := x - star.X
		dy := y - star.Y
		dist := dx*dx + dy*dy

		// Stars are small points, check if within radius
		starRadius := 0.003 // Very small
		if dist < starRadius*starRadius {
			// Calculate twinkle effect
			twinkle := 0.7 + 0.3*math.Sin(time*sf.TwinkleSpeed+star.TwinklePhase)

			// Apply night intensity (stars fade at dawn/dusk)
			intensity := star.Brightness * twinkle * nightIntensity

			return color.RGBA{
				R: uint8(float64(star.Color.R) * intensity),
				G: uint8(float64(star.Color.G) * intensity),
				B: uint8(float64(star.Color.B) * intensity),
				A: 255,
			}
		}
	}

	return color.RGBA{} // No star at this position
}

// SetGenre updates the genre and regenerates stars with new colors.
func (sf *StarField) SetGenre(genre string) {
	if sf.Genre != genre {
		sf.Genre = genre
		sf.generateStars()
	}
}

// GetNightIntensity calculates how visible stars should be based on time of day.
// Returns 0 during day, 1 at full night, with gradual transitions at dawn/dusk.
func GetNightIntensity(timeOfDay float64) float64 {
	// Night is roughly 19:00 to 05:00
	// Transition periods: 05:00-07:00 (dawn), 17:00-19:00 (dusk)

	if timeOfDay >= 7 && timeOfDay < 17 {
		return 0 // Full day, no stars
	} else if timeOfDay >= 19 || timeOfDay < 5 {
		return 1 // Full night, full stars
	} else if timeOfDay >= 5 && timeOfDay < 7 {
		// Dawn transition: 5:00 = 1.0, 7:00 = 0.0
		return 1 - (timeOfDay-5)/2
	} else {
		// Dusk transition: 17:00 = 0.0, 19:00 = 1.0
		return (timeOfDay - 17) / 2
	}
}
