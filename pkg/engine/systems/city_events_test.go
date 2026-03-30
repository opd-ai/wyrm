package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewCityEventSystem(t *testing.T) {
	tests := []struct {
		name  string
		genre string
		seed  int64
	}{
		{"fantasy system", "fantasy", 12345},
		{"sci-fi system", "sci-fi", 54321},
		{"cyberpunk system", "cyberpunk", 99999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewCityEventSystem(tt.genre, tt.seed)
			if system == nil {
				t.Fatal("NewCityEventSystem returned nil")
			}
			if system.Genre != tt.genre {
				t.Errorf("Genre = %v, want %v", system.Genre, tt.genre)
			}
			if system.Seed != tt.seed {
				t.Errorf("Seed = %v, want %v", system.Seed, tt.seed)
			}
			if system.rng == nil {
				t.Error("RNG is nil")
			}
		})
	}
}

func TestGetEventPool(t *testing.T) {
	tests := []struct {
		name           string
		genre          string
		expectedEvents []string
	}{
		{
			"fantasy events",
			"fantasy",
			[]string{EventTypeFestival, EventTypeMarketDay, EventTypeTournament},
		},
		{
			"sci-fi events",
			"sci-fi",
			[]string{EventTypeFestival, EventTypeBlackout, EventTypeMartialLaw},
		},
		{
			"horror events",
			"horror",
			[]string{EventTypeCultRitual, EventTypeHaunting, EventTypePlague},
		},
		{
			"cyberpunk events",
			"cyberpunk",
			[]string{EventTypeRiot, EventTypeHacking, EventTypeBlackout},
		},
		{
			"post-apocalyptic events",
			"post-apocalyptic",
			[]string{EventTypeRadiation, EventTypeMutantAttack, EventTypePlague},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewCityEventSystem(tt.genre, 12345)
			pool := system.getEventPool()

			if len(pool) == 0 {
				t.Error("Event pool is empty")
			}

			// Check that expected events are in the pool
			for _, expected := range tt.expectedEvents {
				found := false
				for _, event := range pool {
					if event == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected event %v not found in pool for %v", expected, tt.genre)
				}
			}
		})
	}
}

func TestCreateEvent(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)

	tests := []struct {
		eventType      string
		expectedName   string
		minDuration    float64
		hasPriceEffect bool
	}{
		{EventTypeFestival, "", CityEventDurationFestival, true},
		{EventTypeMarketDay, "Grand Market Day", CityEventDurationMarket, true},
		{EventTypePlague, "Outbreak", CityEventDurationPlague, true},
		{EventTypeRiot, "Civil Unrest", CityEventDurationRiot, true},
		{EventTypeSiege, "City Under Siege", CityEventDurationSiege, true},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			event := system.createEvent(tt.eventType, 100.0)

			if event == nil {
				t.Fatal("createEvent returned nil")
			}

			if event.EventType != tt.eventType {
				t.Errorf("EventType = %v, want %v", event.EventType, tt.eventType)
			}

			if tt.expectedName != "" && event.Name != tt.expectedName {
				t.Errorf("Name = %v, want %v", event.Name, tt.expectedName)
			}

			if event.Duration != tt.minDuration {
				t.Errorf("Duration = %v, want %v", event.Duration, tt.minDuration)
			}

			if !event.Active {
				t.Error("Event should be active")
			}

			if event.StartTime != 100.0 {
				t.Errorf("StartTime = %v, want 100.0", event.StartTime)
			}

			if tt.hasPriceEffect && event.Effects.ShopPriceMultiplier == 0 {
				t.Error("ShopPriceMultiplier should not be zero")
			}
		})
	}
}

func TestCityEventSystemUpdate(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// Create a world clock
	clockEntity := world.CreateEntity()
	world.AddComponent(clockEntity, &components.WorldClock{
		Hour:       12,
		Day:        10,
		TimeAccum:  0,
		HourLength: 60.0,
	})

	// Run several updates
	for i := 0; i < 100; i++ {
		system.Update(world, 0.016)
	}

	// The system should handle updates without error
	// Events may or may not spawn depending on RNG
}

func TestUpdateActiveEvents(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// Create a world clock at a specific time
	clockEntity := world.CreateEntity()
	world.AddComponent(clockEntity, &components.WorldClock{
		Hour:       12,
		Day:        10,
		TimeAccum:  0,
		HourLength: 60.0,
	})

	// Manually create an event that should expire
	eventEntity := world.CreateEntity()
	world.AddComponent(eventEntity, &components.CityEvent{
		EventType: EventTypeFestival,
		Name:      "Test Festival",
		StartTime: 0, // Started at time 0
		Duration:  1, // Lasts 1 hour
		Active:    true,
	})

	// Current time is day 10 hour 12 = 240+12 = 252 hours
	// Event started at 0, duration 1, so it should end at 1
	system.updateActiveEvents(world, 252.0)

	// Check that the event is now inactive
	for _, e := range world.Entities("CityEvent") {
		eventComp, ok := world.GetComponent(e, "CityEvent")
		if !ok {
			continue
		}
		event := eventComp.(*components.CityEvent)
		if event.Active {
			t.Error("Event should be inactive after expiration")
		}
	}
}

func TestGetActiveEvents(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// Create multiple events, some active, some not
	e1 := world.CreateEntity()
	world.AddComponent(e1, &components.CityEvent{
		EventType: EventTypeFestival,
		Active:    true,
	})

	e2 := world.CreateEntity()
	world.AddComponent(e2, &components.CityEvent{
		EventType: EventTypeMarketDay,
		Active:    false,
	})

	e3 := world.CreateEntity()
	world.AddComponent(e3, &components.CityEvent{
		EventType: EventTypePlague,
		Active:    true,
	})

	activeEvents := system.GetActiveEvents(world)
	if len(activeEvents) != 2 {
		t.Errorf("GetActiveEvents returned %d events, want 2", len(activeEvents))
	}
}

func TestGetShopPriceMultiplier(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// No events - should return 1.0
	mult := system.GetShopPriceMultiplier(world)
	if mult != 1.0 {
		t.Errorf("Multiplier with no events = %v, want 1.0", mult)
	}

	// Add an event with price multiplier
	e := world.CreateEntity()
	world.AddComponent(e, &components.CityEvent{
		EventType: EventTypeFestival,
		Active:    true,
		Effects: components.CityEventEffects{
			ShopPriceMultiplier: 1.2,
		},
	})

	mult = system.GetShopPriceMultiplier(world)
	if mult != 1.2 {
		t.Errorf("Multiplier = %v, want 1.2", mult)
	}
}

func TestGetCrimePenaltyMultiplier(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// Add a martial law event
	e := world.CreateEntity()
	world.AddComponent(e, &components.CityEvent{
		EventType: EventTypeMartialLaw,
		Active:    true,
		Effects: components.CityEventEffects{
			CrimePenaltyMultiplier: 3.0,
		},
	})

	mult := system.GetCrimePenaltyMultiplier(world)
	if mult != 3.0 {
		t.Errorf("Multiplier = %v, want 3.0", mult)
	}
}

func TestGetSpawnRateMultiplier(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// Add a siege event (high spawn rate)
	e := world.CreateEntity()
	world.AddComponent(e, &components.CityEvent{
		EventType: EventTypeSiege,
		Active:    true,
		Effects: components.CityEventEffects{
			SpawnRateMultiplier: 3.0,
		},
	})

	mult := system.GetSpawnRateMultiplier(world)
	if mult != 3.0 {
		t.Errorf("Multiplier = %v, want 3.0", mult)
	}
}

func TestTriggerEvent(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// Create a world clock
	clockEntity := world.CreateEntity()
	world.AddComponent(clockEntity, &components.WorldClock{
		Hour:       12,
		Day:        10,
		TimeAccum:  0,
		HourLength: 60.0,
	})

	event := system.TriggerEvent(world, EventTypeTournament)
	if event == nil {
		t.Fatal("TriggerEvent returned nil")
	}

	if event.EventType != EventTypeTournament {
		t.Errorf("EventType = %v, want %v", event.EventType, EventTypeTournament)
	}

	if !event.Active {
		t.Error("Triggered event should be active")
	}

	// Verify event was added to world
	activeEvents := system.GetActiveEvents(world)
	if len(activeEvents) != 1 {
		t.Errorf("Expected 1 active event, got %d", len(activeEvents))
	}
}

func TestHasActiveEventOfType(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// No events
	if system.HasActiveEventOfType(world, EventTypeFestival) {
		t.Error("Should not have festival event")
	}

	// Add a festival
	e := world.CreateEntity()
	world.AddComponent(e, &components.CityEvent{
		EventType: EventTypeFestival,
		Active:    true,
	})

	if !system.HasActiveEventOfType(world, EventTypeFestival) {
		t.Error("Should have festival event")
	}

	if system.HasActiveEventOfType(world, EventTypePlague) {
		t.Error("Should not have plague event")
	}
}

func TestGetEventSeverity(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// No events - severity should be 0
	sev := system.GetEventSeverity(world)
	if sev != 0.0 {
		t.Errorf("Severity with no events = %v, want 0.0", sev)
	}

	// Add events with different severities
	e1 := world.CreateEntity()
	world.AddComponent(e1, &components.CityEvent{
		EventType: EventTypeFestival,
		Active:    true,
		Severity:  0.2,
	})

	e2 := world.CreateEntity()
	world.AddComponent(e2, &components.CityEvent{
		EventType: EventTypeSiege,
		Active:    true,
		Severity:  1.0,
	})

	sev = system.GetEventSeverity(world)
	if sev != 1.0 {
		t.Errorf("Severity = %v, want 1.0 (highest)", sev)
	}
}

func TestGenerateFestivalName(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "cyberpunk", "post-apocalyptic", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			system := NewCityEventSystem(genre, 12345)
			name := system.generateFestivalName()
			if name == "" {
				t.Error("Festival name should not be empty")
			}
		})
	}
}

func TestCityEventComponent(t *testing.T) {
	event := &components.CityEvent{
		EventType:   EventTypeFestival,
		Name:        "Test Festival",
		Description: "A test event",
		CityID:      "city1",
		StartTime:   100.0,
		Duration:    48.0,
		Severity:    0.5,
		Active:      true,
	}

	if event.Type() != "CityEvent" {
		t.Errorf("Type() = %v, want CityEvent", event.Type())
	}
}

func TestCountActiveEvents(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	count := system.countActiveEvents(world)
	if count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	// Add some events
	for i := 0; i < 3; i++ {
		e := world.CreateEntity()
		world.AddComponent(e, &components.CityEvent{
			EventType: EventTypeFestival,
			Active:    true,
		})
	}

	// Add one inactive event
	e := world.CreateEntity()
	world.AddComponent(e, &components.CityEvent{
		EventType: EventTypeFestival,
		Active:    false,
	})

	count = system.countActiveEvents(world)
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}
}

func TestMultipleEventMultipliers(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// Add two events with price multipliers
	e1 := world.CreateEntity()
	world.AddComponent(e1, &components.CityEvent{
		EventType: EventTypeFestival,
		Active:    true,
		Effects: components.CityEventEffects{
			ShopPriceMultiplier: 1.2,
		},
	})

	e2 := world.CreateEntity()
	world.AddComponent(e2, &components.CityEvent{
		EventType: EventTypePlague,
		Active:    true,
		Effects: components.CityEventEffects{
			ShopPriceMultiplier: 1.5,
		},
	})

	// Combined multiplier should be 1.2 * 1.5 = 1.8
	mult := system.GetShopPriceMultiplier(world)
	expected := 1.8
	// Use epsilon comparison for floating point
	epsilon := 0.0001
	if mult < expected-epsilon || mult > expected+epsilon {
		t.Errorf("Combined multiplier = %v, want %v", mult, expected)
	}
}

func TestGetGuardPatrolMultiplier(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// No events - should return 1.0
	mult := system.GetGuardPatrolMultiplier(world)
	if mult != 1.0 {
		t.Errorf("Multiplier with no events = %v, want 1.0", mult)
	}

	// Add a martial law event (high guard presence)
	e := world.CreateEntity()
	world.AddComponent(e, &components.CityEvent{
		EventType: EventTypeMartialLaw,
		Active:    true,
		Effects: components.CityEventEffects{
			GuardPatrolMultiplier: 4.0,
		},
	})

	mult = system.GetGuardPatrolMultiplier(world)
	if mult != 4.0 {
		t.Errorf("Multiplier = %v, want 4.0", mult)
	}
}

func TestGetQuestRewardMultiplier(t *testing.T) {
	system := NewCityEventSystem("fantasy", 12345)
	world := ecs.NewWorld()

	// No events - should return 1.0
	mult := system.GetQuestRewardMultiplier(world)
	if mult != 1.0 {
		t.Errorf("Multiplier with no events = %v, want 1.0", mult)
	}

	// Add a siege event (high rewards)
	e := world.CreateEntity()
	world.AddComponent(e, &components.CityEvent{
		EventType: EventTypeSiege,
		Active:    true,
		Effects: components.CityEventEffects{
			QuestRewardMultiplier: 2.0,
		},
	})

	mult = system.GetQuestRewardMultiplier(world)
	if mult != 2.0 {
		t.Errorf("Multiplier = %v, want 2.0", mult)
	}
}
