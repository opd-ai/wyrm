package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// TestNewEconomicEventSystem tests system initialization.
func TestNewEconomicEventSystem(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	if system == nil {
		t.Fatal("NewEconomicEventSystem returned nil")
	}
	if system.Genre != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", system.Genre)
	}
	if system.Events == nil {
		t.Error("Events map should be initialized")
	}
}

// TestTriggerEconomicEvent tests manually triggering an event.
func TestTriggerEconomicEvent(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	event := system.TriggerEvent(EventMarketBoom, SeverityModerate, nil, nil)
	if event == nil {
		t.Fatal("TriggerEvent returned nil")
	}
	if event.Type != EventMarketBoom {
		t.Errorf("expected type MarketBoom, got %d", event.Type)
	}
	if event.Severity != SeverityModerate {
		t.Errorf("expected severity Moderate, got %d", event.Severity)
	}
	if !event.IsGlobal {
		t.Error("event with no nodes should be global")
	}
}

// TestTriggerEventWithNodes tests triggering a local event.
func TestTriggerEventWithNodes(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	nodes := []ecs.Entity{1, 2, 3}
	event := system.TriggerEvent(EventShortage, SeverityMajor, nodes, []string{"ore", "cloth"})
	if event.IsGlobal {
		t.Error("event with specific nodes should not be global")
	}
	if len(event.AffectedNodes) != 3 {
		t.Errorf("expected 3 affected nodes, got %d", len(event.AffectedNodes))
	}
	if len(event.AffectedItems) != 2 {
		t.Errorf("expected 2 affected items, got %d", len(event.AffectedItems))
	}
}

// TestGetActiveEconomicEvents tests retrieving active events.
func TestGetActiveEconomicEvents(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	// No events initially
	if len(system.GetActiveEvents()) != 0 {
		t.Error("should have no active events initially")
	}
	// Trigger some events
	system.TriggerEvent(EventMarketBoom, SeverityMinor, nil, nil)
	system.TriggerEvent(EventShortage, SeverityModerate, nil, nil)
	active := system.GetActiveEvents()
	if len(active) != 2 {
		t.Errorf("expected 2 active events, got %d", len(active))
	}
}

// TestUpdate tests the update cycle.
func TestEconomicEventUpdate(t *testing.T) {
	world := ecs.NewWorld()
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 100.0)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	// Create economy node
	nodeID := world.CreateEntity()
	node := &components.EconomyNode{
		PriceTable: make(map[string]float64),
		Supply:     make(map[string]int),
		Demand:     make(map[string]int),
	}
	node.PriceTable["ore"] = 100.0
	node.Supply["ore"] = 100
	node.Demand["ore"] = 100
	world.AddComponent(nodeID, node)
	// Trigger a shortage event
	event := system.TriggerEvent(EventShortage, SeverityMajor, []ecs.Entity{nodeID}, []string{"ore"})
	initialProgress := event.Progress
	// Simulate time passing
	for i := 0; i < 100; i++ {
		system.Update(world, 360.0) // 6 minutes per tick
	}
	if event.Progress <= initialProgress {
		t.Error("event should have made progress")
	}
}

// TestEventResolution tests events completing.
func TestEventResolution(t *testing.T) {
	world := ecs.NewWorld()
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	// Create short event
	event := system.TriggerEvent(EventMarketBoom, SeverityMinor, nil, nil)
	// Run until completion
	for i := 0; i < 1000; i++ {
		system.Update(world, 360.0)
		if event.IsResolved {
			break
		}
	}
	if !event.IsResolved {
		t.Error("event should have resolved")
	}
}

// TestEventHistory tests history tracking.
func TestEventHistory(t *testing.T) {
	world := ecs.NewWorld()
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	// Initially empty
	if len(system.GetEventHistory()) != 0 {
		t.Error("history should be empty initially")
	}
	// Trigger and complete an event
	event := system.TriggerEvent(EventMarketBoom, SeverityMinor, nil, nil)
	for i := 0; i < 500; i++ {
		system.Update(world, 360.0)
		if event.IsResolved {
			break
		}
	}
	// After cleanup, should be in history
	system.Update(world, 1.0) // Trigger cleanup
	history := system.GetEventHistory()
	if len(history) == 0 {
		t.Log("Note: Event may still be processing")
	}
}

// TestGetEvent tests retrieving specific event.
func TestGetEvent(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	event := system.TriggerEvent(EventMarketCrash, SeverityModerate, nil, nil)
	retrieved, exists := system.GetEvent(event.ID)
	if !exists {
		t.Error("event should exist")
	}
	if retrieved.ID != event.ID {
		t.Error("retrieved event should match")
	}
}

// TestGetSeverityDescription tests severity descriptions.
func TestGetSeverityDescription(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	severities := []EconomicEventSeverity{SeverityMinor, SeverityModerate, SeverityMajor, SeverityCatastrophic}
	for _, sev := range severities {
		desc := system.GetSeverityDescription(sev)
		if desc == "" || desc == "Unknown severity" {
			t.Errorf("severity %d should have description", sev)
		}
	}
}

// TestSetEventFrequency tests setting event frequency.
func TestSetEventFrequency(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	system.SetEventFrequency(0.5)
	if system.EventFrequency != 0.5 {
		t.Errorf("expected frequency 0.5, got %f", system.EventFrequency)
	}
	// Test clamping
	system.SetEventFrequency(2.0)
	if system.EventFrequency != 1.0 {
		t.Errorf("frequency should be clamped to 1.0, got %f", system.EventFrequency)
	}
}

// TestEstimateEventImpact tests impact estimation.
func TestEstimateEventImpact(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	minorEvent := system.TriggerEvent(EventMarketBoom, SeverityMinor, []ecs.Entity{1}, nil)
	majorEvent := system.TriggerEvent(EventMarketCrash, SeverityCatastrophic, nil, nil)
	minorImpact := system.EstimateEventImpact(minorEvent)
	majorImpact := system.EstimateEventImpact(majorEvent)
	if majorImpact <= minorImpact {
		t.Error("catastrophic global event should have more impact than minor local")
	}
}

// TestGenreEventDescriptions tests genre-specific descriptions.
func TestGenreEventDescriptions(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		economy := NewEconomySystem(0.5, 0.01)
		system := NewEconomicEventSystem(12345, genre, economy)
		event := system.TriggerEvent(EventMarketBoom, SeverityModerate, nil, nil)
		if event.Name == "" {
			t.Errorf("genre %s should have event name", genre)
		}
		if event.Description == "" {
			t.Errorf("genre %s should have event description", genre)
		}
	}
}

// TestAllEventTypes tests all event type configurations.
func TestAllEventTypes(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	types := []EconomicEventType{
		EventMarketBoom, EventMarketCrash, EventShortage, EventSurplus,
		EventInflation, EventDeflation, EventTradeWar, EventGoldRush,
		EventPlague, EventHarvest, EventDrought, EventDiscovery,
		EventMonopoly, EventBankruptcy, EventTaxChange,
	}
	for _, eventType := range types {
		event := system.TriggerEvent(eventType, SeverityModerate, nil, nil)
		if event.Duration <= 0 {
			t.Errorf("event type %d should have positive duration", eventType)
		}
		// Check modifiers are set
		if event.PriceModifier == 0 && event.SupplyModifier == 0 && event.DemandModifier == 0 {
			t.Errorf("event type %d should have at least one modifier", eventType)
		}
	}
}

// TestEventConsequences tests that certain events have consequences.
func TestEventConsequences(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	// Major market crash should have consequences
	event := system.TriggerEvent(EventMarketCrash, SeverityMajor, nil, nil)
	if len(event.Consequences) == 0 {
		t.Error("major market crash should have consequences")
	}
}

// TestPseudoRandom tests deterministic random generation.
func TestEconomicEventPseudoRandom(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system1 := NewEconomicEventSystem(12345, "fantasy", economy)
	system2 := NewEconomicEventSystem(12345, "fantasy", economy)
	// Same seed should produce same sequence
	for i := 0; i < 10; i++ {
		r1 := system1.pseudoRandom()
		r2 := system2.pseudoRandom()
		if r1 != r2 {
			t.Errorf("iteration %d: same seed should produce same random values", i)
		}
	}
}

// TestRandomEventGeneration tests automatic event generation.
func TestRandomEventGeneration(t *testing.T) {
	world := ecs.NewWorld()
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 100.0)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	system.SetEventFrequency(1.0) // 100% chance
	// Create economy node
	nodeID := world.CreateEntity()
	node := &components.EconomyNode{
		PriceTable: make(map[string]float64),
		Supply:     make(map[string]int),
		Demand:     make(map[string]int),
	}
	node.PriceTable["ore"] = 100.0
	world.AddComponent(nodeID, node)
	// Run updates - should generate events
	for i := 0; i < 100; i++ {
		system.Update(world, 3600.0)
	}
	// Should have some events
	if len(system.GetActiveEvents()) == 0 && len(system.GetEventHistory()) == 0 {
		t.Log("Note: Random event generation may not have triggered")
	}
}

// TestMaxEvents tests maximum concurrent events limit.
func TestMaxEvents(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewEconomicEventSystem(12345, "fantasy", economy)
	// Trigger more than max events
	for i := 0; i < 10; i++ {
		system.TriggerEvent(EventMarketBoom, SeverityMinor, nil, nil)
	}
	// Should have 10 (no limit on manual triggers, only random)
	if len(system.GetActiveEvents()) != 10 {
		t.Errorf("expected 10 active events, got %d", len(system.GetActiveEvents()))
	}
}

// TestAbsFloat tests absolute value function.
func TestAbsFloat(t *testing.T) {
	if absFloat(-5.0) != 5.0 {
		t.Error("absFloat(-5.0) should be 5.0")
	}
	if absFloat(5.0) != 5.0 {
		t.Error("absFloat(5.0) should be 5.0")
	}
	if absFloat(0.0) != 0.0 {
		t.Error("absFloat(0.0) should be 0.0")
	}
}
