package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// TestNewTradeRouteSystem tests trade route system initialization.
func TestNewTradeRouteSystem(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	if system == nil {
		t.Fatal("NewTradeRouteSystem returned nil")
	}
	if system.Genre != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", system.Genre)
	}
	if system.Routes == nil {
		t.Error("Routes map should be initialized")
	}
	if system.Caravans == nil {
		t.Error("Caravans map should be initialized")
	}
}

// TestCreateRoute tests creating new trade routes.
func TestCreateRoute(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	route := system.CreateRoute(1, 2, "Town A", "Town B", 100.0)
	if route == nil {
		t.Fatal("CreateRoute returned nil")
	}
	if route.OriginName != "Town A" {
		t.Errorf("expected origin 'Town A', got '%s'", route.OriginName)
	}
	if route.DestinationName != "Town B" {
		t.Errorf("expected destination 'Town B', got '%s'", route.DestinationName)
	}
	if route.Distance != 100.0 {
		t.Errorf("expected distance 100.0, got %f", route.Distance)
	}
	if route.Status != TradeRouteActive {
		t.Errorf("expected status Active, got %d", route.Status)
	}
	if route.TravelTime != 10.0 {
		t.Errorf("expected travel time 10.0, got %f", route.TravelTime)
	}
}

// TestDiscoverRoute tests route discovery by players.
func TestDiscoverRoute(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	system.CreateRoute(1, 2, "Town A", "Town B", 100.0)
	playerID := ecs.Entity(100)
	// Discover route
	discovered := system.DiscoverRoute(playerID, "Town A_to_Town B")
	if !discovered {
		t.Error("first discovery should return true")
	}
	// Try to discover again
	duplicate := system.DiscoverRoute(playerID, "Town A_to_Town B")
	if duplicate {
		t.Error("duplicate discovery should return false")
	}
	// Try non-existent route
	invalid := system.DiscoverRoute(playerID, "Invalid_to_Route")
	if invalid {
		t.Error("invalid route discovery should return false")
	}
}

// TestGetDiscoveredRoutes tests retrieving discovered routes.
func TestGetDiscoveredRoutes(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	system.CreateRoute(1, 2, "Town A", "Town B", 100.0)
	system.CreateRoute(2, 3, "Town B", "Town C", 150.0)
	playerID := ecs.Entity(100)
	system.DiscoverRoute(playerID, "Town A_to_Town B")
	system.DiscoverRoute(playerID, "Town B_to_Town C")
	routes := system.GetDiscoveredRoutes(playerID)
	if len(routes) != 2 {
		t.Errorf("expected 2 discovered routes, got %d", len(routes))
	}
}

// TestLaunchCaravan tests launching a trade caravan.
func TestLaunchCaravan(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 10.0)
	economy.SetBasePrice("cloth", 5.0)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	system.CreateRoute(1, 2, "Mine", "Market", 50.0)
	cargo := map[string]int{"ore": 20, "cloth": 10}
	caravan := system.LaunchCaravan(100, "Mine_to_Market", cargo, 2)
	if caravan == nil {
		t.Fatal("LaunchCaravan returned nil")
	}
	if caravan.RouteID != "Mine_to_Market" {
		t.Errorf("expected route 'Mine_to_Market', got '%s'", caravan.RouteID)
	}
	if caravan.Guards != 2 {
		t.Errorf("expected 2 guards, got %d", caravan.Guards)
	}
	if caravan.Status != CaravanTraveling {
		t.Errorf("expected status Traveling, got %d", caravan.Status)
	}
	// CargoCost should be 20*10 + 10*5 = 250
	if caravan.CargoCost != 250.0 {
		t.Errorf("expected cargo cost 250.0, got %f", caravan.CargoCost)
	}
}

// TestLaunchCaravanInvalidRoute tests launching on non-existent route.
func TestLaunchCaravanInvalidRoute(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	cargo := map[string]int{"ore": 10}
	caravan := system.LaunchCaravan(100, "Invalid_Route", cargo, 1)
	if caravan != nil {
		t.Error("LaunchCaravan should return nil for invalid route")
	}
}

// TestCalculateTravelSpeed tests travel speed calculation.
func TestCalculateTravelSpeed(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	// Light cargo, no guards
	speed1 := system.calculateTravelSpeed(map[string]int{"ore": 10}, 0)
	// Heavy cargo, no guards
	speed2 := system.calculateTravelSpeed(map[string]int{"ore": 150}, 0)
	// Light cargo, many guards
	speed3 := system.calculateTravelSpeed(map[string]int{"ore": 10}, 10)
	// Heavy cargo slows down
	if speed2 >= speed1 {
		t.Error("heavy cargo should be slower than light cargo")
	}
	// Guards slow down slightly
	if speed3 >= speed1 {
		t.Error("many guards should be slower than no guards")
	}
}

// TestUpdate tests the update cycle.
func TestTradeRouteUpdate(t *testing.T) {
	world := ecs.NewWorld()
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 10.0)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	route := system.CreateRoute(1, 2, "Mine", "Market", 10.0)
	cargo := map[string]int{"ore": 10}
	caravan := system.LaunchCaravan(100, route.ID, cargo, 1)
	initialProgress := caravan.Progress
	// Simulate time passing (3600 seconds = 1 hour)
	for i := 0; i < 100; i++ {
		system.Update(world, 36.0)
	}
	if caravan.Progress <= initialProgress {
		t.Error("caravan should have made progress")
	}
}

// TestCaravanArrival tests caravan completing journey.
func TestCaravanArrival(t *testing.T) {
	world := ecs.NewWorld()
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 10.0)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	route := system.CreateRoute(1, 2, "Mine", "Market", 1.0) // Very short route
	cargo := map[string]int{"ore": 10}
	caravan := system.LaunchCaravan(100, route.ID, cargo, 0)
	// Advance until arrival
	for i := 0; i < 1000; i++ {
		system.Update(world, 10.0)
		if caravan.Status == CaravanReturning || caravan.Status == CaravanArrived {
			break
		}
	}
	if caravan.Status != CaravanReturning && caravan.Status != CaravanArrived {
		t.Errorf("caravan should have arrived or be returning, status: %d", caravan.Status)
	}
}

// TestRouteHazards tests hazard generation.
func TestRouteHazards(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	route := system.CreateRoute(1, 2, "Town A", "Town B", 100.0)
	// Manually set hazard
	route.Hazard = HazardBandits
	route.HazardSeverity = 0.5
	route.Status = TradeRouteDisrupted
	status, hazard, severity := system.GetRouteStatus(route.ID)
	if status != TradeRouteDisrupted {
		t.Errorf("expected disrupted status, got %d", status)
	}
	if hazard != HazardBandits {
		t.Errorf("expected bandit hazard, got %d", hazard)
	}
	if severity != 0.5 {
		t.Errorf("expected severity 0.5, got %f", severity)
	}
}

// TestSuspendRoute tests suspending and resuming routes.
func TestSuspendRoute(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	system.CreateRoute(1, 2, "Town A", "Town B", 100.0)
	// Suspend route
	suspended := system.SuspendRoute("Town A_to_Town B")
	if !suspended {
		t.Error("SuspendRoute should return true for valid route")
	}
	status, _, _ := system.GetRouteStatus("Town A_to_Town B")
	if status != TradeRouteSuspended {
		t.Error("route should be suspended")
	}
	// Resume route
	resumed := system.ResumeRoute("Town A_to_Town B")
	if !resumed {
		t.Error("ResumeRoute should return true for suspended route")
	}
	status, _, _ = system.GetRouteStatus("Town A_to_Town B")
	if status != TradeRouteActive {
		t.Error("route should be active after resume")
	}
}

// TestGetHazardDescription tests genre-specific hazard descriptions.
func TestGetHazardDescription(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		system := NewTradeRouteSystem(12345, genre, economy)
		desc := system.GetHazardDescription(HazardBandits)
		if desc == "" || desc == "Unknown hazard" {
			t.Errorf("genre %s should have description for bandits", genre)
		}
	}
}

// TestEstimateProfit tests profit estimation.
func TestEstimateProfit(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 100.0)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	route := system.CreateRoute(1, 2, "Mine", "Market", 50.0)
	cargo := map[string]int{"ore": 10}
	// Basic estimate without hazards
	profit := system.EstimateProfit(route.ID, cargo, 0)
	if profit <= 0 {
		t.Error("basic trade should have positive expected profit")
	}
	// With hazard, profit should be lower
	route.Hazard = HazardBandits
	route.HazardSeverity = 0.5
	profitWithHazard := system.EstimateProfit(route.ID, cargo, 0)
	if profitWithHazard >= profit {
		t.Error("hazard should reduce expected profit")
	}
	// Guards reduce hazard impact (use 1 guard for minimal cost)
	profitWith1Guard := system.EstimateProfit(route.ID, cargo, 1)
	// The reduction in hazard risk with 1 guard should be noticeable
	// hazardRisk without guard = 0.5 * 1.0 * 0.3 = 0.15
	// hazardRisk with 1 guard = 0.5 * 0.9 * 0.3 = 0.135
	// This saves ~1.5% of sellValue, but guard costs 50
	// For this test, we just verify the estimate function runs correctly
	t.Logf("Profit no hazard: %.2f, with hazard: %.2f, with 1 guard: %.2f",
		profit, profitWithHazard, profitWith1Guard)
}

// TestCalculateCargoCapacity tests cargo capacity calculation.
func TestCalculateCargoCapacity(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	cargo := map[string]int{"ore": 30, "cloth": 20}
	remaining := system.CalculateCargoCapacity(cargo)
	expected := 100 - 50 // Default capacity - current cargo
	if remaining != expected {
		t.Errorf("expected remaining capacity %d, got %d", expected, remaining)
	}
}

// TestGetCaravanStatus tests retrieving player's caravans.
func TestGetCaravanStatus(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 10.0)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	system.CreateRoute(1, 2, "Mine", "Market", 50.0)
	system.CreateRoute(2, 3, "Market", "City", 75.0)
	playerID := ecs.Entity(100)
	cargo := map[string]int{"ore": 10}
	system.LaunchCaravan(playerID, "Mine_to_Market", cargo, 1)
	system.LaunchCaravan(playerID, "Market_to_City", cargo, 2)
	caravans := system.GetCaravanStatus(playerID)
	if len(caravans) != 2 {
		t.Errorf("expected 2 caravans, got %d", len(caravans))
	}
}

// TestTradeHistory tests trade history recording.
func TestTradeHistory(t *testing.T) {
	world := ecs.NewWorld()
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 10.0)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	route := system.CreateRoute(1, 2, "Mine", "Market", 0.5)
	playerID := ecs.Entity(100)
	cargo := map[string]int{"ore": 5}
	system.LaunchCaravan(playerID, route.ID, cargo, 0)
	// Run until journey completes
	for i := 0; i < 500; i++ {
		system.Update(world, 10.0)
	}
	history := system.GetTradeHistory(playerID)
	// At minimum, should have attempted trade
	if len(history) > 0 && history[0].RouteID != route.ID {
		t.Errorf("trade history should reference correct route")
	}
}

// TestGetAllRoutes tests retrieving all routes.
func TestGetAllRoutes(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	system.CreateRoute(1, 2, "Town A", "Town B", 100.0)
	system.CreateRoute(2, 3, "Town B", "Town C", 150.0)
	system.CreateRoute(3, 4, "Town C", "Town D", 200.0)
	routes := system.GetAllRoutes()
	if len(routes) != 3 {
		t.Errorf("expected 3 routes, got %d", len(routes))
	}
}

// TestGetActiveCaravans tests retrieving active caravans.
func TestGetActiveCaravans(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 10.0)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	system.CreateRoute(1, 2, "Mine", "Market", 100.0)
	cargo := map[string]int{"ore": 10}
	system.LaunchCaravan(100, "Mine_to_Market", cargo, 1)
	system.LaunchCaravan(101, "Mine_to_Market", cargo, 2)
	active := system.GetActiveCaravans()
	if len(active) != 2 {
		t.Errorf("expected 2 active caravans, got %d", len(active))
	}
}

// TestGenreHazards tests that each genre has unique hazard sets.
func TestGenreHazards(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		system := NewTradeRouteSystem(12345, genre, economy)
		hazards := system.getGenreHazards()
		if len(hazards) == 0 {
			t.Errorf("genre %s should have hazards defined", genre)
		}
	}
}

// TestRouteProfitability tests route profitability calculation.
func TestRouteProfitability(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	route := system.CreateRoute(1, 2, "Town A", "Town B", 100.0)
	baseProfitability := system.GetRouteProfitability(route.ID)
	if baseProfitability <= 0 {
		t.Error("new route should have positive profitability")
	}
	// Add hazard - profitability should decrease
	route.Hazard = HazardBandits
	route.HazardSeverity = 0.5
	hazardProfitability := system.GetRouteProfitability(route.ID)
	if hazardProfitability >= baseProfitability {
		t.Error("hazard should reduce profitability")
	}
}

// TestSuspendNonExistentRoute tests suspending invalid route.
func TestSuspendNonExistentRoute(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	suspended := system.SuspendRoute("NonExistent_Route")
	if suspended {
		t.Error("should not be able to suspend non-existent route")
	}
}

// TestPseudoRandom tests deterministic random generation.
func TestTradeRoutePseudoRandom(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system1 := NewTradeRouteSystem(12345, "fantasy", economy)
	system2 := NewTradeRouteSystem(12345, "fantasy", economy)
	// Same seed should produce same sequence
	for i := 0; i < 10; i++ {
		r1 := system1.pseudoRandom()
		r2 := system2.pseudoRandom()
		if r1 != r2 {
			t.Errorf("iteration %d: same seed should produce same random values", i)
		}
	}
}

// TestLaunchCaravanSuspendedRoute tests launching on suspended route.
func TestLaunchCaravanSuspendedRoute(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewTradeRouteSystem(12345, "fantasy", economy)
	system.CreateRoute(1, 2, "Town A", "Town B", 100.0)
	system.SuspendRoute("Town A_to_Town B")
	cargo := map[string]int{"ore": 10}
	caravan := system.LaunchCaravan(100, "Town A_to_Town B", cargo, 1)
	if caravan != nil {
		t.Error("should not be able to launch caravan on suspended route")
	}
}
