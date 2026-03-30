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

// Season constants.
const (
	SeasonSpring = "spring"
	SeasonSummer = "summer"
	SeasonAutumn = "autumn"
	SeasonWinter = "winter"
	// DaysPerSeason is the number of game days per season.
	DaysPerSeason = 30
	// DaysPerYear is the total days in a game year.
	DaysPerYear = 120
)

// SeasonalModifiers contains season-specific effects.
type SeasonalModifiers struct {
	// Temperature affects survival mechanics (-1.0 = freezing, 1.0 = scorching).
	Temperature float64
	// DaylightHours affects the length of daylight.
	DaylightHours float64
	// GrowthRate affects plant/crop growth speed.
	GrowthRate float64
	// WeatherBias adjusts weather probabilities.
	WeatherBias map[string]float64
}

// GetCurrentSeason returns the season based on world day.
func (s *WeatherSystem) GetCurrentSeason(worldDay int) string {
	dayOfYear := worldDay % DaysPerYear
	switch {
	case dayOfYear < DaysPerSeason:
		return SeasonSpring
	case dayOfYear < DaysPerSeason*2:
		return SeasonSummer
	case dayOfYear < DaysPerSeason*3:
		return SeasonAutumn
	default:
		return SeasonWinter
	}
}

// GetSeasonalModifiers returns modifiers for the current season.
func (s *WeatherSystem) GetSeasonalModifiers(worldDay int) SeasonalModifiers {
	season := s.GetCurrentSeason(worldDay)

	mods := SeasonalModifiers{
		WeatherBias: make(map[string]float64),
	}

	switch season {
	case SeasonSpring:
		mods.Temperature = 0.2
		mods.DaylightHours = 13.0
		mods.GrowthRate = 1.5
		mods.WeatherBias["rain"] = 1.3
		mods.WeatherBias["clear"] = 0.9
	case SeasonSummer:
		mods.Temperature = 0.7
		mods.DaylightHours = 16.0
		mods.GrowthRate = 1.2
		mods.WeatherBias["clear"] = 1.4
		mods.WeatherBias["thunderstorm"] = 1.2
		mods.WeatherBias["rain"] = 0.7
	case SeasonAutumn:
		mods.Temperature = 0.1
		mods.DaylightHours = 11.0
		mods.GrowthRate = 0.5
		mods.WeatherBias["fog"] = 1.5
		mods.WeatherBias["cloudy"] = 1.3
		mods.WeatherBias["rain"] = 1.1
	case SeasonWinter:
		mods.Temperature = -0.5
		mods.DaylightHours = 8.0
		mods.GrowthRate = 0.0
		mods.WeatherBias["fog"] = 1.3
		mods.WeatherBias["clear"] = 0.8
		// Add snow for applicable genres
		if s.Genre == "fantasy" || s.Genre == "post-apocalyptic" {
			mods.WeatherBias["snow"] = 1.5
		}
	}

	return mods
}

// EnvironmentalSound represents an ambient sound triggered by weather/environment.
type EnvironmentalSound struct {
	// SoundID identifies the sound to play.
	SoundID string
	// Volume is the playback volume (0.0-1.0).
	Volume float64
	// Looping indicates if the sound should loop.
	Looping bool
	// Priority determines which sound plays if limited channels.
	Priority int
}

// GetEnvironmentalSounds returns sounds appropriate for current weather.
func (s *WeatherSystem) GetEnvironmentalSounds() []EnvironmentalSound {
	sounds := []EnvironmentalSound{}

	switch s.CurrentWeather {
	case "clear":
		sounds = append(sounds, EnvironmentalSound{SoundID: "ambient_wind_light", Volume: 0.2, Looping: true, Priority: 1})
	case "cloudy", "overcast":
		sounds = append(sounds, EnvironmentalSound{SoundID: "ambient_wind_medium", Volume: 0.3, Looping: true, Priority: 1})
	case "rain":
		sounds = append(sounds, EnvironmentalSound{SoundID: "rain_steady", Volume: 0.6, Looping: true, Priority: 2})
		sounds = append(sounds, EnvironmentalSound{SoundID: "ambient_wind_light", Volume: 0.2, Looping: true, Priority: 1})
	case "fog", "mist":
		sounds = append(sounds, EnvironmentalSound{SoundID: "ambient_eerie", Volume: 0.3, Looping: true, Priority: 1})
	case "thunderstorm":
		sounds = append(sounds, EnvironmentalSound{SoundID: "rain_heavy", Volume: 0.8, Looping: true, Priority: 3})
		sounds = append(sounds, EnvironmentalSound{SoundID: "thunder_rolling", Volume: 0.7, Looping: false, Priority: 4})
		sounds = append(sounds, EnvironmentalSound{SoundID: "ambient_wind_strong", Volume: 0.5, Looping: true, Priority: 2})
	case "dust", "dust_storm":
		sounds = append(sounds, EnvironmentalSound{SoundID: "wind_howling", Volume: 0.7, Looping: true, Priority: 3})
		sounds = append(sounds, EnvironmentalSound{SoundID: "sand_particles", Volume: 0.4, Looping: true, Priority: 2})
	case "ion_storm":
		sounds = append(sounds, EnvironmentalSound{SoundID: "electrical_crackle", Volume: 0.6, Looping: true, Priority: 3})
		sounds = append(sounds, EnvironmentalSound{SoundID: "static_hum", Volume: 0.3, Looping: true, Priority: 2})
	case "radiation_burst", "radiation_fog":
		sounds = append(sounds, EnvironmentalSound{SoundID: "geiger_counter", Volume: 0.5, Looping: true, Priority: 3})
		sounds = append(sounds, EnvironmentalSound{SoundID: "ambient_ominous", Volume: 0.4, Looping: true, Priority: 2})
	case "blood_moon":
		sounds = append(sounds, EnvironmentalSound{SoundID: "ambient_horror", Volume: 0.5, Looping: true, Priority: 2})
		sounds = append(sounds, EnvironmentalSound{SoundID: "distant_howl", Volume: 0.3, Looping: false, Priority: 3})
	case "smog":
		sounds = append(sounds, EnvironmentalSound{SoundID: "industrial_hum", Volume: 0.4, Looping: true, Priority: 2})
	case "acid_rain":
		sounds = append(sounds, EnvironmentalSound{SoundID: "rain_sizzle", Volume: 0.6, Looping: true, Priority: 3})
		sounds = append(sounds, EnvironmentalSound{SoundID: "chemical_drip", Volume: 0.3, Looping: true, Priority: 2})
	case "neon_haze":
		sounds = append(sounds, EnvironmentalSound{SoundID: "city_ambient", Volume: 0.5, Looping: true, Priority: 2})
		sounds = append(sounds, EnvironmentalSound{SoundID: "neon_buzz", Volume: 0.2, Looping: true, Priority: 1})
	case "ash_fall":
		sounds = append(sounds, EnvironmentalSound{SoundID: "ambient_wind_light", Volume: 0.3, Looping: true, Priority: 1})
		sounds = append(sounds, EnvironmentalSound{SoundID: "ash_settling", Volume: 0.2, Looping: true, Priority: 2})
	case "scorching":
		sounds = append(sounds, EnvironmentalSound{SoundID: "heat_shimmer", Volume: 0.3, Looping: true, Priority: 2})
		sounds = append(sounds, EnvironmentalSound{SoundID: "desert_wind", Volume: 0.4, Looping: true, Priority: 1})
	}

	return sounds
}

// GetDaylightInfo returns sunrise/sunset times for a given day.
func (s *WeatherSystem) GetDaylightInfo(worldDay int) (sunriseHour, sunsetHour int) {
	seasonMods := s.GetSeasonalModifiers(worldDay)
	daylightHours := seasonMods.DaylightHours

	// Center daylight around noon (hour 12)
	sunriseHour = 12 - int(daylightHours/2)
	sunsetHour = 12 + int(daylightHours/2)

	if sunriseHour < 0 {
		sunriseHour = 0
	}
	if sunsetHour > 24 {
		sunsetHour = 24
	}

	return sunriseHour, sunsetHour
}

// IsDaytime returns true if the current hour is during daylight.
func (s *WeatherSystem) IsDaytime(worldDay, hour int) bool {
	sunrise, sunset := s.GetDaylightInfo(worldDay)
	return hour >= sunrise && hour < sunset
}

// GetAmbientLightLevel returns light level based on time and weather (0.0-1.0).
func (s *WeatherSystem) GetAmbientLightLevel(worldDay, hour int) float64 {
	sunrise, sunset := s.GetDaylightInfo(worldDay)
	baseLight := s.calculateBaseLightLevel(hour, sunrise, sunset)
	weatherMods := s.GetWeatherModifiers()
	return baseLight * weatherMods.Visibility
}

// calculateBaseLightLevel determines base light from time of day.
func (s *WeatherSystem) calculateBaseLightLevel(hour, sunrise, sunset int) float64 {
	switch {
	case hour < sunrise:
		return 0.1 + 0.1*float64(hour)/float64(sunrise)
	case hour < sunrise+1:
		return 0.3 + 0.4*float64(hour-sunrise)
	case hour < sunset-1:
		return 1.0
	case hour < sunset:
		return 1.0 - 0.4*float64(hour-sunset+1)
	default:
		return s.calculateNightLightLevel(hour, sunset)
	}
}

// calculateNightLightLevel returns light level for nighttime hours.
func (s *WeatherSystem) calculateNightLightLevel(hour, sunset int) float64 {
	hoursAfterSunset := hour - sunset
	baseLight := 0.3 - 0.2*float64(hoursAfterSunset)/float64(24-sunset)
	if baseLight < 0.1 {
		baseLight = 0.1
	}
	return baseLight
}
