package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestFactionCoupSystemCreation(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	politicsSys := NewFactionPoliticsSystem(0.01)
	sys := NewFactionCoupSystem(rankSys, politicsSys, 12345, "fantasy")

	if sys == nil {
		t.Fatal("NewFactionCoupSystem returned nil")
	}
	if sys.Genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got '%s'", sys.Genre)
	}
	if sys.rng == nil {
		t.Error("RNG should be initialized")
	}
}

func TestStartCoup(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")

	// Start a coup
	result := sys.StartCoup("knights", 0, "npc", "disputed succession")
	if !result {
		t.Error("StartCoup should return true for new coup")
	}

	// Verify coup was created
	coup := sys.GetCoup("knights")
	if coup == nil {
		t.Fatal("coup should exist after StartCoup")
	}
	if coup.State != CoupStatePlotting {
		t.Errorf("expected CoupStatePlotting, got %d", coup.State)
	}
	if coup.Reason != "disputed succession" {
		t.Errorf("expected reason 'disputed succession', got '%s'", coup.Reason)
	}

	// Try to start another coup in same faction
	result = sys.StartCoup("knights", 0, "npc", "another reason")
	if result {
		t.Error("StartCoup should return false when coup already exists")
	}
}

func TestPlayerStartCoup(t *testing.T) {
	world := ecs.NewWorld()
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")
	entity := world.CreateEntity()

	// Try to start coup without membership
	result := sys.PlayerStartCoup(world, entity, "knights")
	if result {
		t.Error("should not start coup without membership")
	}

	// Join faction at rank 1
	rankSys.JoinFaction(world, entity, "knights", "military", 0)

	// Try to start coup at low rank
	result = sys.PlayerStartCoup(world, entity, "knights")
	if result {
		t.Error("should not start coup at rank 1")
	}

	// Promote to rank 5
	rankSys.AddXP(world, entity, "knights", 2000)

	// Now should be able to start coup
	result = sys.PlayerStartCoup(world, entity, "knights")
	if !result {
		t.Error("rank 5+ should be able to start coup")
	}

	coup := sys.GetCoup("knights")
	if coup == nil {
		t.Fatal("coup should exist")
	}
	if coup.InstigatorType != "player" {
		t.Errorf("expected instigator 'player', got '%s'", coup.InstigatorType)
	}
	if coup.LeaderEntity != entity {
		t.Error("player should be coup leader")
	}
}

func TestSupportAndOpposeCoup(t *testing.T) {
	world := ecs.NewWorld()
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")
	entity := world.CreateEntity()

	// Join faction
	rankSys.JoinFaction(world, entity, "knights", "military", 0)
	rankSys.AddXP(world, entity, "knights", 500) // Rank 3

	// Start a coup
	sys.StartCoup("knights", 0, "npc", "test")
	coup := sys.GetCoup("knights")
	initialSupport := coup.SupportLevel
	initialResistance := coup.ResistanceLevel

	// Support the coup
	result := sys.SupportCoup(world, entity, "knights")
	if !result {
		t.Error("SupportCoup should return true")
	}
	if coup.SupportLevel <= initialSupport {
		t.Error("support should increase")
	}

	// Oppose the coup
	result = sys.OpposeCoup(world, entity, "knights")
	if !result {
		t.Error("OpposeCoup should return true")
	}
	if coup.ResistanceLevel <= initialResistance {
		t.Error("resistance should increase")
	}
}

func TestCoupStateProgression(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")
	sys.MinPlotDuration = 1.0 // Short for testing

	// Start a coup
	sys.StartCoup("knights", 0, "npc", "test")
	coup := sys.GetCoup("knights")

	// Initial state should be plotting
	if coup.State != CoupStatePlotting {
		t.Errorf("expected CoupStatePlotting, got %d", coup.State)
	}

	// Set high support to trigger active state
	coup.SupportLevel = 0.6
	coup.Duration = 2.0 // Past min plot duration

	// Update to trigger state change
	world := ecs.NewWorld()
	sys.Update(world, 0.1)

	if coup.State != CoupStateActive {
		t.Errorf("expected CoupStateActive with high support, got %d", coup.State)
	}
}

func TestCoupSuccess(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")
	sys.MinPlotDuration = 0.1

	// Start a coup
	sys.StartCoup("knights", 0, "npc", "test")
	coup := sys.GetCoup("knights")

	// Set to active state with high support
	coup.State = CoupStateActive
	coup.SupportLevel = 0.7
	coup.ResistanceLevel = 0.2
	coup.Duration = 0.5

	// Update to trigger resolution
	world := ecs.NewWorld()
	sys.Update(world, 0.1)

	if coup.State != CoupStateSucceeded {
		t.Errorf("expected CoupStateSucceeded, got %d", coup.State)
	}
}

func TestCoupFailure(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")
	sys.MinPlotDuration = 0.1

	// Start a coup
	sys.StartCoup("knights", 0, "npc", "test")
	coup := sys.GetCoup("knights")

	// Set to active state with high resistance
	coup.State = CoupStateActive
	coup.SupportLevel = 0.2
	coup.ResistanceLevel = 0.8
	coup.Duration = 0.5

	// Update to trigger resolution
	world := ecs.NewWorld()
	sys.Update(world, 0.1)

	if coup.State != CoupStateFailed {
		t.Errorf("expected CoupStateFailed, got %d", coup.State)
	}
}

func TestCoupFinalization(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")

	// Start and complete a coup
	sys.StartCoup("knights", 0, "npc", "test")
	coup := sys.GetCoup("knights")
	coup.State = CoupStateSucceeded

	// Update to finalize
	world := ecs.NewWorld()
	sys.Update(world, 0.1)

	// Coup should be removed from active
	if sys.GetCoup("knights") != nil {
		t.Error("finished coup should be removed from active coups")
	}

	// Should be in history
	history := sys.GetCoupHistory("knights")
	if len(history) != 1 {
		t.Errorf("expected 1 coup in history, got %d", len(history))
	}
}

func TestIsCoupActive(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")

	// No coup - not active
	if sys.IsCoupActive("knights") {
		t.Error("should not be active without coup")
	}

	// Start coup - should be active
	sys.StartCoup("knights", 0, "npc", "test")
	if !sys.IsCoupActive("knights") {
		t.Error("plotting coup should be active")
	}

	// Set to succeeded - not active
	coup := sys.GetCoup("knights")
	coup.State = CoupStateSucceeded
	if sys.IsCoupActive("knights") {
		t.Error("succeeded coup should not be active")
	}
}

func TestGetCoupSuccessChance(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")

	// No coup
	chance := sys.GetCoupSuccessChance("knights")
	if chance != 0 {
		t.Errorf("expected 0%% chance with no coup, got %f", chance)
	}

	// Start coup with equal support/resistance
	sys.StartCoup("knights", 0, "npc", "test")
	coup := sys.GetCoup("knights")
	coup.SupportLevel = 0.5
	coup.ResistanceLevel = 0.5

	chance = sys.GetCoupSuccessChance("knights")
	if chance < 45 || chance > 55 {
		t.Errorf("expected ~50%% chance, got %f", chance)
	}

	// Higher support
	coup.SupportLevel = 0.8
	coup.ResistanceLevel = 0.2

	chance = sys.GetCoupSuccessChance("knights")
	if chance < 75 || chance > 85 {
		t.Errorf("expected ~80%% chance, got %f", chance)
	}
}

func TestForceCoupResolution(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")

	sys.StartCoup("knights", 0, "npc", "test")

	// Force success
	sys.ForceCoupResolution("knights", true)
	coup := sys.GetCoup("knights")
	if coup.State != CoupStateSucceeded {
		t.Errorf("expected CoupStateSucceeded after force, got %d", coup.State)
	}

	// Start another and force failure
	sys.Update(ecs.NewWorld(), 0.1) // Finalize previous
	sys.StartCoup("knights", 0, "npc", "test")
	sys.ForceCoupResolution("knights", false)
	coup = sys.GetCoup("knights")
	if coup.State != CoupStateFailed {
		t.Errorf("expected CoupStateFailed after force, got %d", coup.State)
	}
}

func TestGetAllActiveCoups(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")

	// Start multiple coups
	sys.StartCoup("knights", 0, "npc", "test1")
	sys.StartCoup("mages", 0, "npc", "test2")
	sys.StartCoup("thieves", 0, "npc", "test3")

	all := sys.GetAllActiveCoups()
	if len(all) != 3 {
		t.Errorf("expected 3 active coups, got %d", len(all))
	}
}

func TestGenreSpecificReasons(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		rankSys := NewFactionRankSystem(genre)
		sys := NewFactionCoupSystem(rankSys, nil, 12345, genre)

		reasons := sys.getGenreReasons()
		if len(reasons) == 0 {
			t.Errorf("no coup reasons for genre '%s'", genre)
		}
	}
}

func TestCoupStateConstants(t *testing.T) {
	// Verify state constants are distinct
	states := []CoupState{
		CoupStateNone,
		CoupStatePlotting,
		CoupStateActive,
		CoupStateSucceeded,
		CoupStateFailed,
	}
	seen := make(map[CoupState]bool)
	for _, s := range states {
		if seen[s] {
			t.Errorf("duplicate coup state value: %d", s)
		}
		seen[s] = true
	}
}

func TestCoupWithFactionTerritory(t *testing.T) {
	world := ecs.NewWorld()
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionCoupSystem(rankSys, nil, 12345, "fantasy")

	// Create a faction territory entity
	territoryEntity := world.CreateEntity()
	world.AddComponent(territoryEntity, &components.FactionTerritory{
		FactionID:    "knights",
		ControlLevel: 1.0,
	})

	// Update should find the faction and potentially start coup (based on RNG)
	// This tests the collectFactions function
	factions := sys.collectFactions(world)
	found := false
	for _, f := range factions {
		if f == "knights" {
			found = true
			break
		}
	}
	if !found {
		t.Error("collectFactions should find 'knights' faction")
	}
}
