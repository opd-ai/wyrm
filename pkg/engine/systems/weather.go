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

	// Extreme weather event tracking
	ExtremeEvent       *ExtremeWeatherEvent
	ExtremeEventChance float64 // Probability per weather transition (0.0-1.0)
	lastExtremeCheck   float64
}

// ExtremeWeatherEvent represents a rare, powerful weather phenomenon.
type ExtremeWeatherEvent struct {
	Type        string  // Event type identifier
	Intensity   float64 // Severity (0.0-1.0)
	Duration    float64 // Remaining duration in seconds
	MaxDuration float64 // Original duration
	DamageRate  float64 // Damage per second to exposed entities
	WarningTime float64 // Time remaining for warning (0 = active)
	CenterX     float64 // Event center X coordinate
	CenterY     float64 // Event center Y coordinate
	Radius      float64 // Affected area radius
	MovementX   float64 // Event movement direction X
	MovementY   float64 // Event movement direction Y
	Speed       float64 // Event movement speed
}

// ExtremeEventType constants
const (
	ExtremeEventTornado       = "tornado"
	ExtremeEventBlizzard      = "blizzard"
	ExtremeEventHurricane     = "hurricane"
	ExtremeEventVolcanic      = "volcanic"
	ExtremeEventSolarFlare    = "solar_flare"
	ExtremeEventRadiationWave = "radiation_wave"
	ExtremeEventMeteorShower  = "meteor_shower"
	ExtremeEventEarthquake    = "earthquake"
	ExtremeEventFlood         = "flood"
	ExtremeEventDarkRitual    = "dark_ritual"
	ExtremeEventDragonFlight  = "dragon_flight"
	ExtremeEventAcidStorm     = "acid_storm"
)

// GetExtremeEventPool returns genre-appropriate extreme events.
func (s *WeatherSystem) GetExtremeEventPool() []string {
	switch s.Genre {
	case "fantasy":
		return []string{
			ExtremeEventTornado, ExtremeEventBlizzard, ExtremeEventFlood,
			ExtremeEventEarthquake, ExtremeEventDarkRitual, ExtremeEventDragonFlight,
		}
	case "sci-fi":
		return []string{
			ExtremeEventSolarFlare, ExtremeEventRadiationWave,
			ExtremeEventMeteorShower, ExtremeEventEarthquake,
		}
	case "horror":
		return []string{
			ExtremeEventBlizzard, ExtremeEventFlood, ExtremeEventDarkRitual,
			ExtremeEventEarthquake,
		}
	case "cyberpunk":
		return []string{ExtremeEventAcidStorm, ExtremeEventSolarFlare, ExtremeEventFlood}
	case "post-apocalyptic":
		return []string{
			ExtremeEventRadiationWave, ExtremeEventTornado,
			ExtremeEventEarthquake, ExtremeEventAcidStorm,
		}
	default:
		return []string{ExtremeEventTornado, ExtremeEventBlizzard, ExtremeEventFlood}
	}
}

// CreateExtremeEvent spawns a new extreme weather event.
func (s *WeatherSystem) CreateExtremeEvent(eventType string, x, y float64) *ExtremeWeatherEvent {
	event := &ExtremeWeatherEvent{
		Type:      eventType,
		CenterX:   x,
		CenterY:   y,
		Intensity: 0.5,
	}

	// Set event-specific defaults
	switch eventType {
	case ExtremeEventTornado:
		event.Intensity = 0.8
		event.Duration = 120.0
		event.MaxDuration = 120.0
		event.DamageRate = 10.0
		event.WarningTime = 30.0
		event.Radius = 50.0
		event.Speed = 15.0

	case ExtremeEventBlizzard:
		event.Intensity = 0.7
		event.Duration = 600.0
		event.MaxDuration = 600.0
		event.DamageRate = 2.0
		event.WarningTime = 60.0
		event.Radius = 500.0
		event.Speed = 0.0 // Stationary

	case ExtremeEventHurricane:
		event.Intensity = 0.9
		event.Duration = 300.0
		event.MaxDuration = 300.0
		event.DamageRate = 5.0
		event.WarningTime = 120.0
		event.Radius = 300.0
		event.Speed = 8.0

	case ExtremeEventSolarFlare:
		event.Intensity = 0.6
		event.Duration = 60.0
		event.MaxDuration = 60.0
		event.DamageRate = 3.0
		event.WarningTime = 15.0
		event.Radius = 1000.0 // Global effect
		event.Speed = 0.0

	case ExtremeEventRadiationWave:
		event.Intensity = 0.7
		event.Duration = 180.0
		event.MaxDuration = 180.0
		event.DamageRate = 4.0
		event.WarningTime = 45.0
		event.Radius = 400.0
		event.Speed = 20.0

	case ExtremeEventMeteorShower:
		event.Intensity = 0.5
		event.Duration = 90.0
		event.MaxDuration = 90.0
		event.DamageRate = 15.0 // High damage but random impacts
		event.WarningTime = 20.0
		event.Radius = 200.0
		event.Speed = 0.0

	case ExtremeEventEarthquake:
		event.Intensity = 0.8
		event.Duration = 30.0
		event.MaxDuration = 30.0
		event.DamageRate = 8.0
		event.WarningTime = 5.0
		event.Radius = 300.0
		event.Speed = 0.0

	case ExtremeEventFlood:
		event.Intensity = 0.6
		event.Duration = 900.0
		event.MaxDuration = 900.0
		event.DamageRate = 1.0
		event.WarningTime = 180.0
		event.Radius = 250.0
		event.Speed = 5.0

	case ExtremeEventDarkRitual:
		event.Intensity = 0.9
		event.Duration = 180.0
		event.MaxDuration = 180.0
		event.DamageRate = 5.0
		event.WarningTime = 60.0
		event.Radius = 150.0
		event.Speed = 0.0

	case ExtremeEventDragonFlight:
		event.Intensity = 0.7
		event.Duration = 120.0
		event.MaxDuration = 120.0
		event.DamageRate = 12.0
		event.WarningTime = 30.0
		event.Radius = 100.0
		event.Speed = 30.0 // Fast moving

	case ExtremeEventAcidStorm:
		event.Intensity = 0.8
		event.Duration = 240.0
		event.MaxDuration = 240.0
		event.DamageRate = 3.0
		event.WarningTime = 45.0
		event.Radius = 350.0
		event.Speed = 10.0

	default:
		event.Duration = 60.0
		event.MaxDuration = 60.0
		event.DamageRate = 1.0
		event.WarningTime = 30.0
		event.Radius = 100.0
		event.Speed = 0.0
	}

	s.ExtremeEvent = event
	return event
}

// UpdateExtremeEvent advances the extreme event state.
func (s *WeatherSystem) UpdateExtremeEvent(dt float64) bool {
	if s.ExtremeEvent == nil {
		return false
	}

	e := s.ExtremeEvent

	// Update warning phase
	if e.WarningTime > 0 {
		e.WarningTime -= dt
		if e.WarningTime < 0 {
			e.WarningTime = 0
		}
		return true
	}

	// Update active phase
	e.Duration -= dt
	if e.Duration <= 0 {
		s.ExtremeEvent = nil
		return false
	}

	// Move the event
	if e.Speed > 0 {
		e.CenterX += e.MovementX * e.Speed * dt
		e.CenterY += e.MovementY * e.Speed * dt
	}

	return true
}

// IsExtremeEventActive returns whether an extreme event is in progress.
func (s *WeatherSystem) IsExtremeEventActive() bool {
	return s.ExtremeEvent != nil
}

// IsExtremeEventWarning returns whether we're in the warning phase.
func (s *WeatherSystem) IsExtremeEventWarning() bool {
	return s.ExtremeEvent != nil && s.ExtremeEvent.WarningTime > 0
}

// GetExtremeEventDamage returns damage for a position relative to event.
func (s *WeatherSystem) GetExtremeEventDamage(x, y float64) float64 {
	if s.ExtremeEvent == nil {
		return 0
	}
	e := s.ExtremeEvent

	// No damage during warning
	if e.WarningTime > 0 {
		return 0
	}

	// Calculate distance from event center
	dx := x - e.CenterX
	dy := y - e.CenterY
	dist := dx*dx + dy*dy
	radiusSq := e.Radius * e.Radius

	if dist > radiusSq {
		return 0 // Outside affected area
	}

	// Damage falls off with distance from center
	distRatio := dist / radiusSq
	falloff := 1.0 - distRatio
	return e.DamageRate * e.Intensity * falloff
}

// GetExtremeEventModifiers returns gameplay modifiers for the active event.
func (s *WeatherSystem) GetExtremeEventModifiers() WeatherModifiers {
	mods := WeatherModifiers{
		Visibility: 1.0,
		Movement:   1.0,
		Accuracy:   1.0,
		Damage:     0.0,
		Stealth:    1.0,
	}

	if s.ExtremeEvent == nil || s.ExtremeEvent.WarningTime > 0 {
		return mods
	}

	e := s.ExtremeEvent
	intensity := e.Intensity

	switch e.Type {
	case ExtremeEventTornado:
		mods.Visibility = 0.2
		mods.Movement = 0.3
		mods.Accuracy = 0.1
		mods.Stealth = 0.3
	case ExtremeEventBlizzard:
		mods.Visibility = 0.1
		mods.Movement = 0.4
		mods.Accuracy = 0.2
		mods.Stealth = 0.4
	case ExtremeEventHurricane:
		mods.Visibility = 0.3
		mods.Movement = 0.4
		mods.Accuracy = 0.2
		mods.Stealth = 0.5
	case ExtremeEventSolarFlare:
		mods.Visibility = 0.5 // Blinding light
		mods.Accuracy = 0.6
	case ExtremeEventRadiationWave:
		mods.Visibility = 0.6
		mods.Movement = 0.8
	case ExtremeEventMeteorShower:
		mods.Visibility = 0.7
		mods.Movement = 0.9
		mods.Stealth = 0.8
	case ExtremeEventEarthquake:
		mods.Movement = 0.5
		mods.Accuracy = 0.4
	case ExtremeEventFlood:
		mods.Movement = 0.3
		mods.Stealth = 0.6
	case ExtremeEventDarkRitual:
		mods.Visibility = 0.2
		mods.Accuracy = 0.5
		mods.Stealth = 0.3
	case ExtremeEventDragonFlight:
		mods.Visibility = 0.6
		mods.Movement = 0.9
		mods.Stealth = 0.8
	case ExtremeEventAcidStorm:
		mods.Visibility = 0.4
		mods.Movement = 0.7
		mods.Accuracy = 0.5
	}

	// Scale modifiers by intensity
	mods.Visibility = 1.0 - (1.0-mods.Visibility)*intensity
	mods.Movement = 1.0 - (1.0-mods.Movement)*intensity
	mods.Accuracy = 1.0 - (1.0-mods.Accuracy)*intensity
	mods.Stealth = 1.0 - (1.0-mods.Stealth)*intensity

	return mods
}

// GetExtremeEventDescription returns a warning/description string.
func (s *WeatherSystem) GetExtremeEventDescription() string {
	if s.ExtremeEvent == nil {
		return ""
	}

	e := s.ExtremeEvent
	descriptions := map[string]string{
		ExtremeEventTornado:       "A massive tornado tears through the area!",
		ExtremeEventBlizzard:      "A deadly blizzard has engulfed the region!",
		ExtremeEventHurricane:     "A devastating hurricane approaches!",
		ExtremeEventSolarFlare:    "A powerful solar flare disrupts all systems!",
		ExtremeEventRadiationWave: "A dangerous radiation wave sweeps across the land!",
		ExtremeEventMeteorShower:  "Meteors rain down from the sky!",
		ExtremeEventEarthquake:    "The ground shakes violently!",
		ExtremeEventFlood:         "Rising waters flood the area!",
		ExtremeEventDarkRitual:    "Dark energies pulse from an unholy ritual!",
		ExtremeEventDragonFlight:  "A dragon circles overhead, raining fire!",
		ExtremeEventAcidStorm:     "Corrosive acid rain falls from toxic clouds!",
	}

	warnings := map[string]string{
		ExtremeEventTornado:       "WARNING: Tornado detected! Seek shelter immediately!",
		ExtremeEventBlizzard:      "WARNING: Blizzard approaching! Find warmth and shelter!",
		ExtremeEventHurricane:     "WARNING: Hurricane warning! Evacuate low-lying areas!",
		ExtremeEventSolarFlare:    "WARNING: Solar flare detected! Protect electronics!",
		ExtremeEventRadiationWave: "WARNING: Radiation spike detected! Find cover!",
		ExtremeEventMeteorShower:  "WARNING: Meteor shower incoming! Get to cover!",
		ExtremeEventEarthquake:    "WARNING: Seismic activity detected!",
		ExtremeEventFlood:         "WARNING: Flash flood alert! Move to high ground!",
		ExtremeEventDarkRitual:    "WARNING: Dark magic detected nearby!",
		ExtremeEventDragonFlight:  "WARNING: Dragon spotted in the area!",
		ExtremeEventAcidStorm:     "WARNING: Acid rain imminent! Seek shelter!",
	}

	if e.WarningTime > 0 {
		if msg, ok := warnings[e.Type]; ok {
			return msg
		}
		return "WARNING: Extreme weather event approaching!"
	}

	if msg, ok := descriptions[e.Type]; ok {
		return msg
	}
	return "An extreme weather event is in progress!"
}

// GetExtremeEventProgress returns how far through the event we are (0.0-1.0).
func (s *WeatherSystem) GetExtremeEventProgress() float64 {
	if s.ExtremeEvent == nil {
		return 0
	}
	e := s.ExtremeEvent
	if e.MaxDuration <= 0 {
		return 1.0
	}
	return 1.0 - (e.Duration / e.MaxDuration)
}

// IsPositionInExtremeEvent checks if a position is affected by the event.
func (s *WeatherSystem) IsPositionInExtremeEvent(x, y float64) bool {
	if s.ExtremeEvent == nil {
		return false
	}
	e := s.ExtremeEvent

	dx := x - e.CenterX
	dy := y - e.CenterY
	distSq := dx*dx + dy*dy
	return distSq <= e.Radius*e.Radius
}

// ClearExtremeEvent ends the current extreme event.
func (s *WeatherSystem) ClearExtremeEvent() {
	s.ExtremeEvent = nil
}

// SetExtremeEventMovement sets the event's movement direction.
func (s *WeatherSystem) SetExtremeEventMovement(dx, dy float64) {
	if s.ExtremeEvent != nil {
		// Normalize direction
		length := dx*dx + dy*dy
		if length > 0 {
			length = 1.0 / (length * length)
			s.ExtremeEvent.MovementX = dx * length
			s.ExtremeEvent.MovementY = dy * length
		}
	}
}

// NewWeatherSystem creates a new weather system.
func NewWeatherSystem(genre string, duration float64) *WeatherSystem {
	return &WeatherSystem{
		Genre:              genre,
		WeatherDuration:    duration,
		CurrentWeather:     "clear",
		ExtremeEventChance: 0.05, // 5% chance per weather transition
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

// GetLightLevel returns ambient light level for a given hour, using day 0 as reference.
func (s *WeatherSystem) GetLightLevel(hour int) float64 {
	return s.GetAmbientLightLevel(0, hour)
}

// LocationType represents where an entity is located.
type LocationType int

const (
	LocationOutdoor LocationType = iota
	LocationIndoor
	LocationUnderground
	LocationUnderwater
)

// IndoorOutdoorZone represents a zone with indoor/outdoor properties.
type IndoorOutdoorZone struct {
	ID               string
	LocationType     LocationType
	MinX, MinY, MinZ float64
	MaxX, MaxY, MaxZ float64
	WeatherShielded  bool
	LightOverride    float64
	AmbientSound     string
}

// IndoorOutdoorSystem detects whether entities are inside or outside.
type IndoorOutdoorSystem struct {
	Zones       map[string]*IndoorOutdoorZone
	EntityZones map[ecs.Entity]string
	DefaultType LocationType
	weatherSys  *WeatherSystem
}

// NewIndoorOutdoorSystem creates a new indoor/outdoor detection system.
func NewIndoorOutdoorSystem(weatherSys *WeatherSystem) *IndoorOutdoorSystem {
	return &IndoorOutdoorSystem{
		Zones:       make(map[string]*IndoorOutdoorZone),
		EntityZones: make(map[ecs.Entity]string),
		DefaultType: LocationOutdoor,
		weatherSys:  weatherSys,
	}
}

// Update checks all entities' positions and updates their location status.
func (s *IndoorOutdoorSystem) Update(w *ecs.World, dt float64) {
	if w == nil {
		return
	}
	for _, e := range w.Entities("Position") {
		s.updateEntityLocation(w, e)
	}
}

// updateEntityLocation determines which zone an entity is in.
func (s *IndoorOutdoorSystem) updateEntityLocation(w *ecs.World, e ecs.Entity) {
	posComp, ok := w.GetComponent(e, "Position")
	if !ok {
		return
	}

	type positioner interface {
		GetX() float64
		GetY() float64
		GetZ() float64
	}

	if pos, ok := posComp.(positioner); ok {
		x, y, z := pos.GetX(), pos.GetY(), pos.GetZ()
		for id, zone := range s.Zones {
			if s.isInZone(x, y, z, zone) {
				s.EntityZones[e] = id
				return
			}
		}
	}
	delete(s.EntityZones, e)
}

// isInZone checks if coordinates are within a zone's bounds.
func (s *IndoorOutdoorSystem) isInZone(x, y, z float64, zone *IndoorOutdoorZone) bool {
	return x >= zone.MinX && x <= zone.MaxX &&
		y >= zone.MinY && y <= zone.MaxY &&
		z >= zone.MinZ && z <= zone.MaxZ
}

// RegisterZone adds a new zone to the system.
func (s *IndoorOutdoorSystem) RegisterZone(zone *IndoorOutdoorZone) {
	s.Zones[zone.ID] = zone
}

// UnregisterZone removes a zone from the system.
func (s *IndoorOutdoorSystem) UnregisterZone(id string) {
	delete(s.Zones, id)
	for e, zoneID := range s.EntityZones {
		if zoneID == id {
			delete(s.EntityZones, e)
		}
	}
}

// GetEntityLocationType returns the location type for an entity.
func (s *IndoorOutdoorSystem) GetEntityLocationType(e ecs.Entity) LocationType {
	if zoneID, ok := s.EntityZones[e]; ok {
		if zone, ok := s.Zones[zoneID]; ok {
			return zone.LocationType
		}
	}
	return s.DefaultType
}

// GetEntityZone returns the zone ID for an entity, or empty string if outside.
func (s *IndoorOutdoorSystem) GetEntityZone(e ecs.Entity) string {
	return s.EntityZones[e]
}

// IsEntityIndoor checks if an entity is in an indoor location.
func (s *IndoorOutdoorSystem) IsEntityIndoor(e ecs.Entity) bool {
	return s.GetEntityLocationType(e) == LocationIndoor
}

// IsEntityOutdoor checks if an entity is in an outdoor location.
func (s *IndoorOutdoorSystem) IsEntityOutdoor(e ecs.Entity) bool {
	return s.GetEntityLocationType(e) == LocationOutdoor
}

// IsEntityUnderground checks if an entity is underground.
func (s *IndoorOutdoorSystem) IsEntityUnderground(e ecs.Entity) bool {
	return s.GetEntityLocationType(e) == LocationUnderground
}

// IsEntityUnderwater checks if an entity is underwater.
func (s *IndoorOutdoorSystem) IsEntityUnderwater(e ecs.Entity) bool {
	return s.GetEntityLocationType(e) == LocationUnderwater
}

// IsWeatherShielded checks if an entity is protected from weather effects.
func (s *IndoorOutdoorSystem) IsWeatherShielded(e ecs.Entity) bool {
	if zoneID, ok := s.EntityZones[e]; ok {
		if zone, ok := s.Zones[zoneID]; ok {
			return zone.WeatherShielded
		}
	}
	return false
}

// GetEffectiveWeatherModifiers returns weather modifiers adjusted for location.
func (s *IndoorOutdoorSystem) GetEffectiveWeatherModifiers(e ecs.Entity) WeatherModifiers {
	if s.weatherSys == nil {
		return WeatherModifiers{
			Visibility: 1.0,
			Movement:   1.0,
			Accuracy:   1.0,
			Damage:     0.0,
			Stealth:    1.0,
		}
	}

	baseMods := s.weatherSys.GetWeatherModifiers()

	if s.IsWeatherShielded(e) {
		return WeatherModifiers{
			Visibility: 1.0,
			Movement:   1.0,
			Accuracy:   1.0,
			Damage:     0.0,
			Stealth:    baseMods.Stealth,
		}
	}

	locType := s.GetEntityLocationType(e)
	switch locType {
	case LocationUnderground:
		return WeatherModifiers{
			Visibility: 0.5,
			Movement:   1.0,
			Accuracy:   1.0,
			Damage:     0.0,
			Stealth:    0.7,
		}
	case LocationUnderwater:
		return WeatherModifiers{
			Visibility: 0.4,
			Movement:   0.6,
			Accuracy:   0.5,
			Damage:     0.0,
			Stealth:    0.8,
		}
	}

	return baseMods
}

// GetEffectiveLightLevel returns the light level adjusted for location.
func (s *IndoorOutdoorSystem) GetEffectiveLightLevel(e ecs.Entity, hour int) float64 {
	if zoneID, ok := s.EntityZones[e]; ok {
		if zone, ok := s.Zones[zoneID]; ok {
			if zone.LightOverride > 0 {
				return zone.LightOverride
			}
			switch zone.LocationType {
			case LocationIndoor:
				return 0.7
			case LocationUnderground:
				return 0.2
			case LocationUnderwater:
				return 0.4
			}
		}
	}

	if s.weatherSys != nil {
		return s.weatherSys.GetLightLevel(hour)
	}
	return 1.0
}

// GetAmbientSound returns the appropriate ambient sound for an entity's location.
func (s *IndoorOutdoorSystem) GetAmbientSound(e ecs.Entity) string {
	if zoneID, ok := s.EntityZones[e]; ok {
		if zone, ok := s.Zones[zoneID]; ok {
			if zone.AmbientSound != "" {
				return zone.AmbientSound
			}
			switch zone.LocationType {
			case LocationIndoor:
				return "ambient_indoor"
			case LocationUnderground:
				return "ambient_cave"
			case LocationUnderwater:
				return "ambient_underwater"
			}
		}
	}

	if s.weatherSys != nil {
		switch s.weatherSys.CurrentWeather {
		case "rain", "thunderstorm", "acid_rain":
			return "ambient_rain"
		case "fog", "mist":
			return "ambient_fog"
		case "dust_storm", "ash_fall":
			return "ambient_wind"
		}
	}
	return "ambient_outdoor"
}

// CreateBuildingZone creates a standard indoor zone for a building.
func (s *IndoorOutdoorSystem) CreateBuildingZone(id string, minX, minY, minZ, maxX, maxY, maxZ float64) *IndoorOutdoorZone {
	zone := &IndoorOutdoorZone{
		ID:              id,
		LocationType:    LocationIndoor,
		MinX:            minX,
		MinY:            minY,
		MinZ:            minZ,
		MaxX:            maxX,
		MaxY:            maxY,
		MaxZ:            maxZ,
		WeatherShielded: true,
		LightOverride:   0.7,
		AmbientSound:    "ambient_indoor",
	}
	s.RegisterZone(zone)
	return zone
}

// CreateCaveZone creates a standard underground zone.
func (s *IndoorOutdoorSystem) CreateCaveZone(id string, minX, minY, minZ, maxX, maxY, maxZ float64) *IndoorOutdoorZone {
	zone := &IndoorOutdoorZone{
		ID:              id,
		LocationType:    LocationUnderground,
		MinX:            minX,
		MinY:            minY,
		MinZ:            minZ,
		MaxX:            maxX,
		MaxY:            maxY,
		MaxZ:            maxZ,
		WeatherShielded: true,
		LightOverride:   0.2,
		AmbientSound:    "ambient_cave",
	}
	s.RegisterZone(zone)
	return zone
}

// CreateUnderwaterZone creates a standard underwater zone.
func (s *IndoorOutdoorSystem) CreateUnderwaterZone(id string, minX, minY, minZ, maxX, maxY, maxZ float64) *IndoorOutdoorZone {
	zone := &IndoorOutdoorZone{
		ID:              id,
		LocationType:    LocationUnderwater,
		MinX:            minX,
		MinY:            minY,
		MinZ:            minZ,
		MaxX:            maxX,
		MaxY:            maxY,
		MaxZ:            maxZ,
		WeatherShielded: true,
		LightOverride:   0.4,
		AmbientSound:    "ambient_underwater",
	}
	s.RegisterZone(zone)
	return zone
}

// GetZoneCount returns the number of registered zones.
func (s *IndoorOutdoorSystem) GetZoneCount() int {
	return len(s.Zones)
}

// GetTrackedEntityCount returns the number of entities in zones.
func (s *IndoorOutdoorSystem) GetTrackedEntityCount() int {
	return len(s.EntityZones)
}

// ClearEntityTracking removes all entity zone associations.
func (s *IndoorOutdoorSystem) ClearEntityTracking() {
	s.EntityZones = make(map[ecs.Entity]string)
}

// GetZone returns a zone by ID.
func (s *IndoorOutdoorSystem) GetZone(id string) *IndoorOutdoorZone {
	return s.Zones[id]
}

// SetEntityZone manually sets an entity's zone.
func (s *IndoorOutdoorSystem) SetEntityZone(e ecs.Entity, zoneID string) bool {
	if _, ok := s.Zones[zoneID]; ok {
		s.EntityZones[e] = zoneID
		return true
	}
	return false
}

// ClearEntityZone removes an entity's zone association.
func (s *IndoorOutdoorSystem) ClearEntityZone(e ecs.Entity) {
	delete(s.EntityZones, e)
}
