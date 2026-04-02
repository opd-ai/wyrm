package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
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

// extremeEventConfig holds default parameters for an extreme event type.
type extremeEventConfig struct {
	Intensity   float64
	Duration    float64
	DamageRate  float64
	WarningTime float64
	Radius      float64
	Speed       float64
}

// extremeEventConfigs maps event types to their default configurations.
var extremeEventConfigs = map[string]extremeEventConfig{
	ExtremeEventTornado:       {0.8, 120.0, 10.0, 30.0, 50.0, 15.0},
	ExtremeEventBlizzard:      {0.7, 600.0, 2.0, 60.0, 500.0, 0.0},
	ExtremeEventHurricane:     {0.9, 300.0, 5.0, 120.0, 300.0, 8.0},
	ExtremeEventSolarFlare:    {0.6, 60.0, 3.0, 15.0, 1000.0, 0.0},
	ExtremeEventRadiationWave: {0.7, 180.0, 4.0, 45.0, 400.0, 20.0},
	ExtremeEventMeteorShower:  {0.5, 90.0, 15.0, 20.0, 200.0, 0.0},
	ExtremeEventEarthquake:    {0.8, 30.0, 8.0, 5.0, 300.0, 0.0},
	ExtremeEventFlood:         {0.6, 900.0, 1.0, 180.0, 250.0, 5.0},
	ExtremeEventDarkRitual:    {0.9, 180.0, 5.0, 60.0, 150.0, 0.0},
	ExtremeEventDragonFlight:  {0.7, 120.0, 12.0, 30.0, 100.0, 30.0},
	ExtremeEventAcidStorm:     {0.8, 240.0, 3.0, 45.0, 350.0, 10.0},
}

// extremeEventModifierConfig holds base modifiers for an extreme event type.
// Values of 1.0 mean no effect; lower values mean more severe impact.
type extremeEventModifierConfig struct {
	Visibility float64
	Movement   float64
	Accuracy   float64
	Stealth    float64
}

// extremeEventModifiers maps event types to their gameplay modifiers.
var extremeEventModifiers = map[string]extremeEventModifierConfig{
	ExtremeEventTornado:       {0.2, 0.3, 0.1, 0.3},
	ExtremeEventBlizzard:      {0.1, 0.4, 0.2, 0.4},
	ExtremeEventHurricane:     {0.3, 0.4, 0.2, 0.5},
	ExtremeEventSolarFlare:    {0.5, 1.0, 0.6, 1.0},
	ExtremeEventRadiationWave: {0.6, 0.8, 1.0, 1.0},
	ExtremeEventMeteorShower:  {0.7, 0.9, 1.0, 0.8},
	ExtremeEventEarthquake:    {1.0, 0.5, 0.4, 1.0},
	ExtremeEventFlood:         {1.0, 0.3, 1.0, 0.6},
	ExtremeEventDarkRitual:    {0.2, 1.0, 0.5, 0.3},
	ExtremeEventDragonFlight:  {0.6, 0.9, 1.0, 0.8},
	ExtremeEventAcidStorm:     {0.4, 0.7, 0.5, 1.0},
}

// weatherModifierConfig holds gameplay modifiers for a weather type.
type weatherModifierConfig struct {
	Visibility float64
	Movement   float64
	Accuracy   float64
	Stealth    float64
	Damage     float64
}

// weatherModifiers maps weather types to their gameplay effects.
var weatherModifiers = map[string]weatherModifierConfig{
	// Common weather types
	"clear":        {1.0, 1.0, 1.0, 1.0, 0.0},
	"cloudy":       {0.9, 1.0, 1.0, 1.0, 0.0},
	"overcast":     {0.9, 1.0, 1.0, 1.0, 0.0},
	"rain":         {0.7, 0.9, 0.85, 0.8, 0.0},
	"fog":          {0.3, 1.0, 0.7, 0.5, 0.0},
	"mist":         {0.3, 1.0, 0.7, 0.5, 0.0},
	"thunderstorm": {0.4, 0.7, 0.6, 0.6, 0.5},
	// Sci-fi weather
	"dust":            {0.5, 1.0, 0.75, 0.7, 0.0},
	"ion_storm":       {0.6, 1.0, 0.5, 1.0, 1.0},
	"radiation_burst": {0.8, 0.8, 1.0, 1.0, 2.0},
	// Horror weather
	"blood_moon": {0.5, 1.0, 1.0, 1.2, 0.3},
	// Cyberpunk weather
	"smog":      {0.6, 0.95, 1.0, 0.75, 0.0},
	"acid_rain": {0.7, 0.85, 1.0, 1.0, 0.8},
	"neon_haze": {0.75, 1.0, 0.9, 1.0, 0.0},
	// Post-apocalyptic weather
	"dust_storm":    {0.2, 0.6, 0.4, 0.4, 0.3},
	"ash_fall":      {0.5, 0.85, 0.8, 1.0, 0.0},
	"radiation_fog": {0.3, 1.0, 1.0, 0.5, 1.5},
	"scorching":     {1.0, 0.7, 1.0, 1.0, 1.0},
}

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
		Type:    eventType,
		CenterX: x,
		CenterY: y,
	}

	// Apply configuration from map, or use defaults
	if cfg, ok := extremeEventConfigs[eventType]; ok {
		event.Intensity = cfg.Intensity
		event.Duration = cfg.Duration
		event.MaxDuration = cfg.Duration
		event.DamageRate = cfg.DamageRate
		event.WarningTime = cfg.WarningTime
		event.Radius = cfg.Radius
		event.Speed = cfg.Speed
	} else {
		// Fallback defaults for unknown event types
		event.Intensity = 0.5
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

	// Apply modifiers from config map
	if cfg, ok := extremeEventModifiers[e.Type]; ok {
		mods.Visibility = cfg.Visibility
		mods.Movement = cfg.Movement
		mods.Accuracy = cfg.Accuracy
		mods.Stealth = cfg.Stealth
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

	// Apply modifiers from config map
	if cfg, ok := weatherModifiers[s.CurrentWeather]; ok {
		mods.Visibility = cfg.Visibility
		mods.Movement = cfg.Movement
		mods.Accuracy = cfg.Accuracy
		mods.Stealth = cfg.Stealth
		mods.Damage = cfg.Damage
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

	// Update or create Weather component for rendering system
	s.syncWeatherComponent(w)
}

// syncWeatherComponent updates the Weather ECS component for rendering sync.
func (s *WeatherSystem) syncWeatherComponent(w *ecs.World) {
	// Find existing weather entity or create one
	weatherEntities := w.Entities("Weather")
	var weatherEntity ecs.Entity

	if len(weatherEntities) > 0 {
		weatherEntity = weatherEntities[0]
	} else {
		// Create a weather entity if none exists
		weatherEntity = w.CreateEntity()
		w.AddComponent(weatherEntity, &components.Weather{})
	}

	// Get and update the component
	weatherComp, ok := w.GetComponent(weatherEntity, "Weather")
	if !ok {
		return
	}
	weather := weatherComp.(*components.Weather)

	// Update weather state
	weather.WeatherType = s.CurrentWeather
	weather.CloudCover = s.GetCloudCover()
	weather.Intensity = s.GetIntensity()
}

// GetCloudCover returns cloud coverage for the current weather type.
func (s *WeatherSystem) GetCloudCover() float64 {
	switch s.CurrentWeather {
	case "clear":
		return 0.0
	case "overcast":
		return 0.8
	case "rain", "storm":
		return 1.0
	case "snow":
		return 0.9
	case "fog":
		return 0.6
	default:
		return 0.0
	}
}

// GetIntensity returns intensity for the current weather type.
func (s *WeatherSystem) GetIntensity() float64 {
	switch s.CurrentWeather {
	case "clear", "overcast":
		return 0.0
	case "rain":
		return 0.5
	case "storm":
		return 1.0
	case "snow":
		return 0.6
	case "fog":
		return 0.4
	default:
		return 0.0
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
