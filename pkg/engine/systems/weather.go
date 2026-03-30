package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// WeatherModifiers contains multipliers that affect gameplay systems.
type WeatherModifiers struct {
	Visibility float64 // Multiplier for draw distance (0.0-1.0)
	Movement   float64 // Multiplier for movement speed (0.0-1.0)
	Accuracy   float64 // Multiplier for ranged/magic hit chance (0.0-1.0)
	Damage     float64 // Periodic environmental damage (0 = none)
	Stealth    float64 // Multiplier for detection distance (lower = easier to hide)
}

// WeatherSystem controls dynamic weather and environmental hazards.
type WeatherSystem struct {
	CurrentWeather  string
	TimeAccum       float64
	WeatherDuration float64 // Duration in seconds before weather change
	Genre           string  // Affects available weather types
	weatherIndex    int     // For deterministic cycling
}

// NewWeatherSystem creates a new weather system.
func NewWeatherSystem(genre string, duration float64) *WeatherSystem {
	return &WeatherSystem{
		Genre:           genre,
		WeatherDuration: duration,
		CurrentWeather:  "clear",
	}
}

// getWeatherPool returns genre-appropriate weather types.
func (s *WeatherSystem) getWeatherPool() []string {
	switch s.Genre {
	case "fantasy":
		return []string{"clear", "cloudy", "rain", "fog", "thunderstorm"}
	case "sci-fi":
		return []string{"clear", "dust", "ion_storm", "radiation_burst"}
	case "horror":
		return []string{"fog", "overcast", "rain", "blood_moon", "mist"}
	case "cyberpunk":
		return []string{"smog", "acid_rain", "clear", "neon_haze"}
	case "post-apocalyptic":
		return []string{"dust_storm", "clear", "ash_fall", "radiation_fog", "scorching"}
	default:
		return []string{"clear", "cloudy", "rain", "fog"}
	}
}

// GetWeatherModifiers returns the current weather's gameplay effects.
func (s *WeatherSystem) GetWeatherModifiers() WeatherModifiers {
	// Default neutral modifiers
	mods := WeatherModifiers{
		Visibility: 1.0,
		Movement:   1.0,
		Accuracy:   1.0,
		Damage:     0.0,
		Stealth:    1.0,
	}

	switch s.CurrentWeather {
	// Common weather types
	case "clear":
		// No modifications
	case "cloudy", "overcast":
		mods.Visibility = 0.9
	case "rain":
		mods.Visibility = 0.7
		mods.Movement = 0.9
		mods.Accuracy = 0.85
		mods.Stealth = 0.8 // Rain makes it easier to hide
	case "fog", "mist":
		mods.Visibility = 0.3
		mods.Accuracy = 0.7
		mods.Stealth = 0.5 // Much easier to hide in fog
	case "thunderstorm":
		mods.Visibility = 0.4
		mods.Movement = 0.7
		mods.Accuracy = 0.6
		mods.Stealth = 0.6
		mods.Damage = 0.5 // Lightning risk

	// Sci-fi weather
	case "dust":
		mods.Visibility = 0.5
		mods.Accuracy = 0.75
		mods.Stealth = 0.7
	case "ion_storm":
		mods.Visibility = 0.6
		mods.Accuracy = 0.5 // Electronics interference
		mods.Damage = 1.0
	case "radiation_burst":
		mods.Visibility = 0.8
		mods.Damage = 2.0 // High radiation damage
		mods.Movement = 0.8

	// Horror weather
	case "blood_moon":
		mods.Visibility = 0.5
		mods.Stealth = 1.2 // Enemies are more alert
		mods.Damage = 0.3  // Cursed damage

	// Cyberpunk weather
	case "smog":
		mods.Visibility = 0.6
		mods.Movement = 0.95
		mods.Stealth = 0.75
	case "acid_rain":
		mods.Visibility = 0.7
		mods.Movement = 0.85
		mods.Damage = 0.8 // Acid damage
	case "neon_haze":
		mods.Visibility = 0.75
		mods.Accuracy = 0.9

	// Post-apocalyptic weather
	case "dust_storm":
		mods.Visibility = 0.2
		mods.Movement = 0.6
		mods.Accuracy = 0.4
		mods.Stealth = 0.4
		mods.Damage = 0.3 // Abrasive damage
	case "ash_fall":
		mods.Visibility = 0.5
		mods.Movement = 0.85
		mods.Accuracy = 0.8
	case "radiation_fog":
		mods.Visibility = 0.3
		mods.Damage = 1.5 // Radiation damage
		mods.Stealth = 0.5
	case "scorching":
		mods.Movement = 0.7
		mods.Damage = 1.0 // Heat damage
	}

	return mods
}

// GetVisibilityMultiplier returns the current visibility modifier.
func (s *WeatherSystem) GetVisibilityMultiplier() float64 {
	return s.GetWeatherModifiers().Visibility
}

// GetMovementMultiplier returns the current movement speed modifier.
func (s *WeatherSystem) GetMovementMultiplier() float64 {
	return s.GetWeatherModifiers().Movement
}

// GetAccuracyMultiplier returns the current ranged accuracy modifier.
func (s *WeatherSystem) GetAccuracyMultiplier() float64 {
	return s.GetWeatherModifiers().Accuracy
}

// GetEnvironmentalDamage returns periodic damage from weather.
func (s *WeatherSystem) GetEnvironmentalDamage() float64 {
	return s.GetWeatherModifiers().Damage
}

// GetStealthMultiplier returns the stealth detection distance modifier.
func (s *WeatherSystem) GetStealthMultiplier() float64 {
	return s.GetWeatherModifiers().Stealth
}

// IsHazardousWeather returns true if the current weather deals damage.
func (s *WeatherSystem) IsHazardousWeather() bool {
	return s.GetWeatherModifiers().Damage > 0
}

// SetWeather forces a specific weather condition (for testing/scripting).
func (s *WeatherSystem) SetWeather(weather string) {
	s.CurrentWeather = weather
	s.TimeAccum = 0
}

// GetWeatherDescription returns a human-readable description.
func (s *WeatherSystem) GetWeatherDescription() string {
	descriptions := map[string]string{
		"clear":           "The weather is clear and visibility is good.",
		"cloudy":          "Clouds gather overhead, slightly dimming the light.",
		"overcast":        "A thick layer of clouds blocks the sky.",
		"rain":            "Rain falls steadily, reducing visibility and making surfaces slick.",
		"fog":             "Dense fog blankets the area, severely limiting sight.",
		"mist":            "A thin mist hangs in the air, obscuring distant objects.",
		"thunderstorm":    "Lightning flashes and thunder roars as a violent storm rages.",
		"dust":            "Fine dust particles fill the air, reducing visibility.",
		"ion_storm":       "An electromagnetic storm crackles through the atmosphere.",
		"radiation_burst": "A wave of radiation sweeps through the area. Seek shelter!",
		"blood_moon":      "An ominous red moon hangs in the sky. Creatures are restless.",
		"smog":            "Thick industrial smog hangs heavy in the air.",
		"acid_rain":       "Corrosive rain falls from polluted clouds. Find cover!",
		"neon_haze":       "A haze of reflected neon light diffuses through the air.",
		"dust_storm":      "A massive dust storm reduces visibility to nearly nothing.",
		"ash_fall":        "Volcanic ash drifts down from the sky.",
		"radiation_fog":   "A glowing fog carries dangerous radiation levels.",
		"scorching":       "The relentless heat beats down without mercy.",
	}

	if desc, ok := descriptions[s.CurrentWeather]; ok {
		return desc
	}
	return "The weather is unremarkable."
}

// Update advances weather simulation each tick.
func (s *WeatherSystem) Update(w *ecs.World, dt float64) {
	s.TimeAccum += dt
	if s.CurrentWeather == "" {
		s.CurrentWeather = "clear"
	}
	if s.WeatherDuration <= 0 {
		s.WeatherDuration = DefaultWeatherDuration
	}
	// Change weather after duration
	if s.TimeAccum >= s.WeatherDuration {
		s.TimeAccum = 0
		pool := s.getWeatherPool()
		s.weatherIndex = (s.weatherIndex + 1) % len(pool)
		s.CurrentWeather = pool[s.weatherIndex]
	}
}
