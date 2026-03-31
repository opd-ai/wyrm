package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestWeatherSystem_GetWeatherModifiers_Clear(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)
	sys.SetWeather("clear")

	mods := sys.GetWeatherModifiers()

	if mods.Visibility != 1.0 {
		t.Errorf("Clear weather should have visibility 1.0, got %f", mods.Visibility)
	}
	if mods.Movement != 1.0 {
		t.Errorf("Clear weather should have movement 1.0, got %f", mods.Movement)
	}
	if mods.Accuracy != 1.0 {
		t.Errorf("Clear weather should have accuracy 1.0, got %f", mods.Accuracy)
	}
	if mods.Damage != 0.0 {
		t.Errorf("Clear weather should have no damage, got %f", mods.Damage)
	}
}

func TestWeatherSystem_GetWeatherModifiers_Rain(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)
	sys.SetWeather("rain")

	mods := sys.GetWeatherModifiers()

	if mods.Visibility >= 1.0 {
		t.Error("Rain should reduce visibility")
	}
	if mods.Movement >= 1.0 {
		t.Error("Rain should reduce movement speed")
	}
	if mods.Accuracy >= 1.0 {
		t.Error("Rain should reduce accuracy")
	}
	if mods.Stealth >= 1.0 {
		t.Error("Rain should make it easier to hide (lower stealth multiplier)")
	}
}

func TestWeatherSystem_GetWeatherModifiers_Fog(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)
	sys.SetWeather("fog")

	mods := sys.GetWeatherModifiers()

	if mods.Visibility > 0.5 {
		t.Error("Fog should severely reduce visibility")
	}
	if mods.Stealth > 0.6 {
		t.Error("Fog should make it much easier to hide")
	}
}

func TestWeatherSystem_GetWeatherModifiers_Thunderstorm(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)
	sys.SetWeather("thunderstorm")

	mods := sys.GetWeatherModifiers()

	if mods.Visibility > 0.5 {
		t.Error("Thunderstorm should severely reduce visibility")
	}
	if mods.Movement > 0.8 {
		t.Error("Thunderstorm should slow movement")
	}
	if mods.Damage <= 0 {
		t.Error("Thunderstorm should deal environmental damage (lightning)")
	}
}

func TestWeatherSystem_GetWeatherModifiers_SciFi(t *testing.T) {
	sys := NewWeatherSystem("sci-fi", 300)

	// Test ion_storm
	sys.SetWeather("ion_storm")
	mods := sys.GetWeatherModifiers()

	if mods.Accuracy >= 0.6 {
		t.Error("Ion storm should significantly reduce accuracy (electronics interference)")
	}
	if mods.Damage <= 0 {
		t.Error("Ion storm should deal damage")
	}

	// Test radiation_burst
	sys.SetWeather("radiation_burst")
	mods = sys.GetWeatherModifiers()

	if mods.Damage < 1.5 {
		t.Error("Radiation burst should deal significant damage")
	}
}

func TestWeatherSystem_GetWeatherModifiers_Horror(t *testing.T) {
	sys := NewWeatherSystem("horror", 300)
	sys.SetWeather("blood_moon")

	mods := sys.GetWeatherModifiers()

	if mods.Stealth <= 1.0 {
		t.Error("Blood moon should make enemies more alert (higher stealth mult)")
	}
	if mods.Damage <= 0 {
		t.Error("Blood moon should deal cursed damage")
	}
}

func TestWeatherSystem_GetWeatherModifiers_Cyberpunk(t *testing.T) {
	sys := NewWeatherSystem("cyberpunk", 300)

	// Test acid_rain
	sys.SetWeather("acid_rain")
	mods := sys.GetWeatherModifiers()

	if mods.Damage <= 0 {
		t.Error("Acid rain should deal damage")
	}
	if mods.Movement >= 1.0 {
		t.Error("Acid rain should reduce movement (avoiding puddles)")
	}

	// Test smog
	sys.SetWeather("smog")
	mods = sys.GetWeatherModifiers()

	if mods.Visibility >= 0.8 {
		t.Error("Smog should reduce visibility")
	}
}

func TestWeatherSystem_GetWeatherModifiers_PostApocalyptic(t *testing.T) {
	sys := NewWeatherSystem("post-apocalyptic", 300)

	// Test dust_storm
	sys.SetWeather("dust_storm")
	mods := sys.GetWeatherModifiers()

	if mods.Visibility > 0.3 {
		t.Error("Dust storm should severely reduce visibility")
	}
	if mods.Movement > 0.7 {
		t.Error("Dust storm should significantly slow movement")
	}
	if mods.Accuracy > 0.5 {
		t.Error("Dust storm should significantly reduce accuracy")
	}
	if mods.Damage <= 0 {
		t.Error("Dust storm should deal abrasive damage")
	}

	// Test radiation_fog
	sys.SetWeather("radiation_fog")
	mods = sys.GetWeatherModifiers()

	if mods.Damage < 1.0 {
		t.Error("Radiation fog should deal significant radiation damage")
	}

	// Test scorching
	sys.SetWeather("scorching")
	mods = sys.GetWeatherModifiers()

	if mods.Movement >= 0.8 {
		t.Error("Scorching heat should slow movement")
	}
	if mods.Damage <= 0 {
		t.Error("Scorching heat should deal heat damage")
	}
}

func TestWeatherSystem_GetVisibilityMultiplier(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	sys.SetWeather("clear")
	if sys.GetVisibilityMultiplier() != 1.0 {
		t.Error("Clear weather visibility should be 1.0")
	}

	sys.SetWeather("fog")
	if sys.GetVisibilityMultiplier() >= 0.5 {
		t.Error("Fog visibility should be below 0.5")
	}
}

func TestWeatherSystem_GetMovementMultiplier(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	sys.SetWeather("clear")
	if sys.GetMovementMultiplier() != 1.0 {
		t.Error("Clear weather movement should be 1.0")
	}

	sys.SetWeather("thunderstorm")
	if sys.GetMovementMultiplier() >= 0.8 {
		t.Error("Thunderstorm movement should be below 0.8")
	}
}

func TestWeatherSystem_GetAccuracyMultiplier(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	sys.SetWeather("clear")
	if sys.GetAccuracyMultiplier() != 1.0 {
		t.Error("Clear weather accuracy should be 1.0")
	}

	sys.SetWeather("rain")
	if sys.GetAccuracyMultiplier() >= 1.0 {
		t.Error("Rain should reduce accuracy")
	}
}

func TestWeatherSystem_GetEnvironmentalDamage(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	sys.SetWeather("clear")
	if sys.GetEnvironmentalDamage() != 0.0 {
		t.Error("Clear weather should not deal damage")
	}

	sys.SetWeather("thunderstorm")
	if sys.GetEnvironmentalDamage() <= 0.0 {
		t.Error("Thunderstorm should deal environmental damage")
	}
}

func TestWeatherSystem_GetStealthMultiplier(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	sys.SetWeather("clear")
	if sys.GetStealthMultiplier() != 1.0 {
		t.Error("Clear weather stealth should be 1.0")
	}

	sys.SetWeather("fog")
	if sys.GetStealthMultiplier() >= 1.0 {
		t.Error("Fog should make it easier to hide (lower stealth mult)")
	}
}

func TestWeatherSystem_IsHazardousWeather(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	nonHazardous := []string{"clear", "cloudy", "rain", "fog"}
	for _, weather := range nonHazardous {
		sys.SetWeather(weather)
		if weather == "thunderstorm" {
			continue // Skip, thunderstorm is hazardous
		}
		if sys.IsHazardousWeather() && weather != "thunderstorm" {
			t.Errorf("%s should not be hazardous", weather)
		}
	}

	// Test hazardous weather
	hazardous := []string{"thunderstorm", "ion_storm", "radiation_burst", "acid_rain", "dust_storm", "radiation_fog", "scorching", "blood_moon"}
	for _, weather := range hazardous {
		sys.SetWeather(weather)
		if !sys.IsHazardousWeather() {
			t.Errorf("%s should be hazardous", weather)
		}
	}
}

func TestWeatherSystem_SetWeather(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	sys.SetWeather("rain")
	if sys.CurrentWeather != "rain" {
		t.Error("SetWeather should change current weather")
	}
	if sys.TimeAccum != 0 {
		t.Error("SetWeather should reset time accumulator")
	}
}

func TestWeatherSystem_GetWeatherDescription(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	weatherTypes := []string{
		"clear", "cloudy", "rain", "fog", "thunderstorm",
		"dust", "ion_storm", "radiation_burst",
		"blood_moon", "mist",
		"smog", "acid_rain", "neon_haze",
		"dust_storm", "ash_fall", "radiation_fog", "scorching",
	}

	for _, weather := range weatherTypes {
		sys.SetWeather(weather)
		desc := sys.GetWeatherDescription()
		if desc == "" {
			t.Errorf("Weather %s should have a description", weather)
		}
		if desc == "The weather is unremarkable." {
			t.Errorf("Weather %s should have a custom description", weather)
		}
	}

	// Test unknown weather
	sys.SetWeather("unknown_weather")
	desc := sys.GetWeatherDescription()
	if desc != "The weather is unremarkable." {
		t.Error("Unknown weather should return default description")
	}
}

func TestWeatherSystem_GenreWeatherPools(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewWeatherSystem(genre, 300)
			pool := sys.getWeatherPool()

			if len(pool) == 0 {
				t.Errorf("Genre %s should have weather pool", genre)
			}

			// Verify all weather types have modifiers
			for _, weather := range pool {
				sys.SetWeather(weather)
				mods := sys.GetWeatherModifiers()

				if mods.Visibility <= 0 || mods.Visibility > 1.5 {
					t.Errorf("Weather %s has invalid visibility: %f", weather, mods.Visibility)
				}
				if mods.Movement <= 0 || mods.Movement > 1.5 {
					t.Errorf("Weather %s has invalid movement: %f", weather, mods.Movement)
				}
				if mods.Accuracy <= 0 || mods.Accuracy > 1.5 {
					t.Errorf("Weather %s has invalid accuracy: %f", weather, mods.Accuracy)
				}
				if mods.Stealth <= 0 || mods.Stealth > 1.5 {
					t.Errorf("Weather %s has invalid stealth: %f", weather, mods.Stealth)
				}
			}
		})
	}
}

func TestWeatherSystem_Update_WeatherChanges(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewWeatherSystem("fantasy", 10) // Short duration for testing

	initialWeather := sys.CurrentWeather

	// Simulate time passing
	for i := 0; i < 15; i++ {
		sys.Update(w, 1.0)
	}

	if sys.CurrentWeather == initialWeather {
		t.Error("Weather should change after duration")
	}
}

func TestWeatherSystem_WeatherEffectsVaryByCondition(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	// Get modifiers for different weather conditions
	conditions := []string{"clear", "rain", "fog", "thunderstorm"}
	modifiersMap := make(map[string]WeatherModifiers)

	for _, condition := range conditions {
		sys.SetWeather(condition)
		modifiersMap[condition] = sys.GetWeatherModifiers()
	}

	// Verify conditions have different effects
	if modifiersMap["clear"].Visibility == modifiersMap["fog"].Visibility {
		t.Error("Clear and fog should have different visibility")
	}
	if modifiersMap["clear"].Movement == modifiersMap["thunderstorm"].Movement {
		t.Error("Clear and thunderstorm should have different movement")
	}
	if modifiersMap["clear"].Accuracy == modifiersMap["rain"].Accuracy {
		t.Error("Clear and rain should have different accuracy")
	}
}

func TestWeatherSystem_WorstCaseModifiers(t *testing.T) {
	sys := NewWeatherSystem("post-apocalyptic", 300)
	sys.SetWeather("dust_storm") // Worst visibility

	mods := sys.GetWeatherModifiers()

	// Even worst case should not completely disable gameplay
	if mods.Visibility < 0.1 {
		t.Error("Visibility should never go below 0.1 (10%)")
	}
	if mods.Movement < 0.5 {
		t.Error("Movement should never go below 0.5 (50%)")
	}
	if mods.Accuracy < 0.3 {
		t.Error("Accuracy should never go below 0.3 (30%)")
	}
}

func TestWeatherSystem_OvercastHasMinimalEffect(t *testing.T) {
	sys := NewWeatherSystem("horror", 300)
	sys.SetWeather("overcast")

	mods := sys.GetWeatherModifiers()

	// Overcast should be very mild
	if mods.Visibility < 0.85 {
		t.Error("Overcast should have only mild visibility reduction")
	}
	if mods.Movement != 1.0 {
		t.Error("Overcast should not affect movement")
	}
	if mods.Accuracy != 1.0 {
		t.Error("Overcast should not affect accuracy")
	}
	if mods.Damage != 0 {
		t.Error("Overcast should not deal damage")
	}
}

// Tests for Extreme Weather Events

func TestExtremeEventPool(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		sys := NewWeatherSystem(genre, 300)
		pool := sys.GetExtremeEventPool()

		if len(pool) == 0 {
			t.Errorf("Genre %s should have at least one extreme event", genre)
		}
	}
}

func TestCreateExtremeEvent(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	event := sys.CreateExtremeEvent(ExtremeEventTornado, 100, 100)

	if event == nil {
		t.Fatal("CreateExtremeEvent should return an event")
	}
	if event.Type != ExtremeEventTornado {
		t.Errorf("Event type should be tornado, got %s", event.Type)
	}
	if event.CenterX != 100 || event.CenterY != 100 {
		t.Error("Event should be at specified position")
	}
	if event.Duration <= 0 {
		t.Error("Event should have positive duration")
	}
	if event.WarningTime <= 0 {
		t.Error("Tornado should have warning time")
	}
}

func TestExtremeEventTypes(t *testing.T) {
	eventTypes := []string{
		ExtremeEventTornado,
		ExtremeEventBlizzard,
		ExtremeEventHurricane,
		ExtremeEventSolarFlare,
		ExtremeEventRadiationWave,
		ExtremeEventMeteorShower,
		ExtremeEventEarthquake,
		ExtremeEventFlood,
		ExtremeEventDarkRitual,
		ExtremeEventDragonFlight,
		ExtremeEventAcidStorm,
	}

	for _, et := range eventTypes {
		sys := NewWeatherSystem("fantasy", 300)
		event := sys.CreateExtremeEvent(et, 0, 0)

		if event == nil {
			t.Errorf("Event type %s should create an event", et)
			continue
		}
		if event.Duration <= 0 {
			t.Errorf("Event %s should have positive duration", et)
		}
		if event.DamageRate <= 0 {
			t.Errorf("Event %s should have damage rate", et)
		}
		if event.Radius <= 0 {
			t.Errorf("Event %s should have positive radius", et)
		}
	}
}

func TestUpdateExtremeEvent(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)
	sys.CreateExtremeEvent(ExtremeEventTornado, 100, 100)

	initialWarning := sys.ExtremeEvent.WarningTime
	initialDuration := sys.ExtremeEvent.Duration

	// Update during warning phase
	result := sys.UpdateExtremeEvent(10.0)
	if !result {
		t.Error("UpdateExtremeEvent should return true while event active")
	}
	if sys.ExtremeEvent.WarningTime >= initialWarning {
		t.Error("Warning time should decrease")
	}

	// Skip past warning
	sys.ExtremeEvent.WarningTime = 0

	// Update during active phase
	sys.UpdateExtremeEvent(10.0)
	if sys.ExtremeEvent.Duration >= initialDuration {
		t.Error("Duration should decrease after warning ends")
	}

	// Exhaust the event
	sys.ExtremeEvent.Duration = 0.5
	result = sys.UpdateExtremeEvent(1.0)
	if result {
		t.Error("UpdateExtremeEvent should return false when event ends")
	}
	if sys.ExtremeEvent != nil {
		t.Error("Event should be cleared when duration expires")
	}
}

func TestExtremeEventDamage(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)
	sys.CreateExtremeEvent(ExtremeEventTornado, 100, 100)

	// No damage during warning
	damage := sys.GetExtremeEventDamage(100, 100)
	if damage != 0 {
		t.Error("No damage should occur during warning phase")
	}

	// Clear warning
	sys.ExtremeEvent.WarningTime = 0

	// Damage at center
	damage = sys.GetExtremeEventDamage(100, 100)
	if damage <= 0 {
		t.Error("Should have damage at event center")
	}

	// Damage falls off with distance
	centerDamage := sys.GetExtremeEventDamage(100, 100)
	edgeDamage := sys.GetExtremeEventDamage(100+sys.ExtremeEvent.Radius*0.8, 100)
	if edgeDamage >= centerDamage {
		t.Error("Damage should be lower at edge than center")
	}

	// No damage outside radius
	outsideDamage := sys.GetExtremeEventDamage(100+sys.ExtremeEvent.Radius*2, 100)
	if outsideDamage != 0 {
		t.Error("No damage outside event radius")
	}
}

func TestExtremeEventModifiers(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	// No modifiers without event
	mods := sys.GetExtremeEventModifiers()
	if mods.Visibility != 1.0 || mods.Movement != 1.0 {
		t.Error("Should have neutral modifiers without event")
	}

	// Create event
	sys.CreateExtremeEvent(ExtremeEventTornado, 0, 0)

	// No modifiers during warning
	mods = sys.GetExtremeEventModifiers()
	if mods.Visibility != 1.0 {
		t.Error("Should have neutral modifiers during warning")
	}

	// Clear warning
	sys.ExtremeEvent.WarningTime = 0

	// Should have reduced modifiers during active event
	mods = sys.GetExtremeEventModifiers()
	if mods.Visibility >= 1.0 {
		t.Error("Tornado should reduce visibility")
	}
	if mods.Movement >= 1.0 {
		t.Error("Tornado should reduce movement")
	}
	if mods.Accuracy >= 1.0 {
		t.Error("Tornado should reduce accuracy")
	}
}

func TestExtremeEventDescription(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	// No description without event
	desc := sys.GetExtremeEventDescription()
	if desc != "" {
		t.Error("Should have no description without event")
	}

	// Warning description
	sys.CreateExtremeEvent(ExtremeEventTornado, 0, 0)
	desc = sys.GetExtremeEventDescription()
	if desc == "" {
		t.Error("Should have warning description")
	}
	if len(desc) < 10 {
		t.Error("Description should be meaningful")
	}

	// Active description
	sys.ExtremeEvent.WarningTime = 0
	desc = sys.GetExtremeEventDescription()
	if desc == "" {
		t.Error("Should have active description")
	}
}

func TestExtremeEventProgress(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	// No progress without event
	progress := sys.GetExtremeEventProgress()
	if progress != 0 {
		t.Error("Should have 0 progress without event")
	}

	// Create event
	sys.CreateExtremeEvent(ExtremeEventTornado, 0, 0)
	sys.ExtremeEvent.WarningTime = 0

	initialDuration := sys.ExtremeEvent.Duration
	progress = sys.GetExtremeEventProgress()
	if progress != 0 {
		t.Error("Should start at 0 progress")
	}

	// Halfway through
	sys.ExtremeEvent.Duration = initialDuration / 2
	progress = sys.GetExtremeEventProgress()
	if progress < 0.4 || progress > 0.6 {
		t.Errorf("Progress should be ~0.5 at halfway, got %f", progress)
	}

	// Near end
	sys.ExtremeEvent.Duration = initialDuration * 0.1
	progress = sys.GetExtremeEventProgress()
	if progress < 0.8 {
		t.Error("Progress should be high near end")
	}
}

func TestIsPositionInExtremeEvent(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	// No event
	if sys.IsPositionInExtremeEvent(0, 0) {
		t.Error("Should return false without event")
	}

	// Create event at origin with radius 50
	sys.CreateExtremeEvent(ExtremeEventTornado, 0, 0)
	sys.ExtremeEvent.Radius = 50

	// Inside
	if !sys.IsPositionInExtremeEvent(0, 0) {
		t.Error("Center should be in event")
	}
	if !sys.IsPositionInExtremeEvent(25, 25) {
		t.Error("Inside radius should be in event")
	}

	// Outside
	if sys.IsPositionInExtremeEvent(100, 100) {
		t.Error("Far position should not be in event")
	}
}

func TestClearExtremeEvent(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)
	sys.CreateExtremeEvent(ExtremeEventTornado, 0, 0)

	if !sys.IsExtremeEventActive() {
		t.Error("Event should be active")
	}

	sys.ClearExtremeEvent()

	if sys.IsExtremeEventActive() {
		t.Error("Event should be cleared")
	}
}

func TestIsExtremeEventWarning(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	// No event
	if sys.IsExtremeEventWarning() {
		t.Error("Should not be warning without event")
	}

	// With event in warning
	sys.CreateExtremeEvent(ExtremeEventTornado, 0, 0)
	if !sys.IsExtremeEventWarning() {
		t.Error("Should be warning with active warning time")
	}

	// After warning
	sys.ExtremeEvent.WarningTime = 0
	if sys.IsExtremeEventWarning() {
		t.Error("Should not be warning after warning time expires")
	}
}

func TestSetExtremeEventMovement(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	// No crash without event
	sys.SetExtremeEventMovement(1, 0)

	// With event
	sys.CreateExtremeEvent(ExtremeEventTornado, 0, 0)
	sys.SetExtremeEventMovement(1, 0)

	// Zero movement
	sys.SetExtremeEventMovement(0, 0)
	// Should not crash
}

func TestExtremeEventChance(t *testing.T) {
	sys := NewWeatherSystem("fantasy", 300)

	if sys.ExtremeEventChance <= 0 {
		t.Error("Default extreme event chance should be positive")
	}
	if sys.ExtremeEventChance > 1.0 {
		t.Error("Extreme event chance should not exceed 1.0")
	}
}
