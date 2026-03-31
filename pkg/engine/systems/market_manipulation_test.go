package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// TestNewMarketManipulationSystem tests system initialization.
func TestNewMarketManipulationSystem(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	if system == nil {
		t.Fatal("NewMarketManipulationSystem returned nil")
	}
	if system.Genre != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", system.Genre)
	}
	if system.Manipulations == nil {
		t.Error("Manipulations map should be initialized")
	}
}

// TestStartManipulation tests starting a new manipulation.
func TestStartManipulation(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	manip, err := system.StartManipulation(playerID, ManipulationRumor, "ore", nodeID, 1000.0)
	if err != nil {
		t.Fatalf("StartManipulation failed: %v", err)
	}
	if manip == nil {
		t.Fatal("StartManipulation returned nil")
	}
	if manip.Type != ManipulationRumor {
		t.Errorf("expected type Rumor, got %d", manip.Type)
	}
	if manip.TargetItem != "ore" {
		t.Errorf("expected target 'ore', got '%s'", manip.TargetItem)
	}
	if manip.Status != ManipulationPlanning {
		t.Errorf("expected status Planning, got %d", manip.Status)
	}
}

// TestStartManipulationInsufficientInvestment tests minimum investment check.
func TestStartManipulationInsufficientInvestment(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	// Try to start with insufficient investment
	_, err := system.StartManipulation(playerID, ManipulationMonopoly, "ore", nodeID, 100.0)
	if err != ErrInsufficientInvestment {
		t.Errorf("expected ErrInsufficientInvestment, got %v", err)
	}
}

// TestStartManipulationCooldown tests cooldown enforcement.
func TestStartManipulationCooldown(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	// Set cooldown
	system.Cooldowns[playerID] = 1000000.0
	_, err := system.StartManipulation(playerID, ManipulationRumor, "ore", nodeID, 1000.0)
	if err != ErrManipulationCooldown {
		t.Errorf("expected ErrManipulationCooldown, got %v", err)
	}
}

// TestAddAccomplice tests adding accomplices to a manipulation.
func TestAddAccomplice(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	accompliceID := ecs.Entity(101)
	manip, _ := system.StartManipulation(playerID, ManipulationCartel, "ore", nodeID, 5000.0)
	initialSuccess := manip.SuccessChance
	initialDetection := manip.DetectionRisk
	success := system.AddAccomplice(manip.ID, accompliceID)
	if !success {
		t.Error("AddAccomplice should return true")
	}
	if len(manip.Accomplices) != 1 {
		t.Errorf("expected 1 accomplice, got %d", len(manip.Accomplices))
	}
	// Success should increase, detection should also increase
	if manip.SuccessChance <= initialSuccess {
		t.Error("accomplice should increase success chance")
	}
	if manip.DetectionRisk <= initialDetection {
		t.Error("accomplice should increase detection risk")
	}
}

// TestAddAccompliceInvalidStatus tests adding accomplice to non-planning manipulation.
func TestAddAccompliceInvalidStatus(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	manip, _ := system.StartManipulation(playerID, ManipulationRumor, "ore", nodeID, 1000.0)
	system.ExecuteManipulation(manip.ID) // Move to active
	success := system.AddAccomplice(manip.ID, ecs.Entity(101))
	if success {
		t.Error("should not be able to add accomplice to active manipulation")
	}
}

// TestExecuteManipulation tests starting the active phase.
func TestExecuteManipulation(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	manip, _ := system.StartManipulation(playerID, ManipulationRumor, "ore", nodeID, 1000.0)
	success := system.ExecuteManipulation(manip.ID)
	if !success {
		t.Error("ExecuteManipulation should return true")
	}
	if manip.Status != ManipulationActive {
		t.Errorf("expected status Active, got %d", manip.Status)
	}
}

// TestUpdate tests the update cycle.
func TestMarketManipulationUpdate(t *testing.T) {
	world := ecs.NewWorld()
	economy := NewEconomySystem(0.5, 0.01)
	economy.SetBasePrice("ore", 100.0)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	// Create a market node
	nodeID := world.CreateEntity()
	node := &components.EconomyNode{
		PriceTable: make(map[string]float64),
		Supply:     make(map[string]int),
		Demand:     make(map[string]int),
	}
	node.PriceTable["ore"] = 100.0
	world.AddComponent(nodeID, node)
	playerID := ecs.Entity(100)
	manip, _ := system.StartManipulation(playerID, ManipulationCorner, "ore", nodeID, 5000.0)
	system.ExecuteManipulation(manip.ID)
	initialProgress := manip.Progress
	// Simulate time passing
	for i := 0; i < 100; i++ {
		system.Update(world, 360.0) // 6 minutes per tick
	}
	if manip.Progress <= initialProgress {
		t.Error("manipulation should have made progress")
	}
}

// TestCancelManipulation tests canceling a manipulation.
func TestCancelManipulation(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	manip, _ := system.StartManipulation(playerID, ManipulationRumor, "ore", nodeID, 1000.0)
	refund, success := system.CancelManipulation(manip.ID)
	if !success {
		t.Error("CancelManipulation should return true")
	}
	// Should refund 80%
	if refund != 800.0 {
		t.Errorf("expected refund 800.0, got %f", refund)
	}
	// Manipulation should be removed
	_, exists := system.GetManipulationStatus(manip.ID)
	if exists {
		t.Error("manipulation should be removed after cancel")
	}
}

// TestCancelActiveManipulation tests canceling an active manipulation.
func TestCancelActiveManipulation(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	manip, _ := system.StartManipulation(playerID, ManipulationRumor, "ore", nodeID, 1000.0)
	system.ExecuteManipulation(manip.ID)
	_, success := system.CancelManipulation(manip.ID)
	if success {
		t.Error("should not be able to cancel active manipulation")
	}
}

// TestGetManipulationDescription tests genre-specific descriptions.
func TestGetManipulationDescription(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		system := NewMarketManipulationSystem(12345, genre, economy)
		desc := system.GetManipulationDescription(ManipulationCorner)
		if desc == "" || desc == "Unknown manipulation type" {
			t.Errorf("genre %s should have description for Corner", genre)
		}
	}
}

// TestGetActiveManipulations tests retrieving active manipulations.
func TestGetActiveManipulations(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	// Create two manipulations
	manip1, _ := system.StartManipulation(playerID, ManipulationRumor, "ore", nodeID, 1000.0)
	system.ExecuteManipulation(manip1.ID)
	manip2, _ := system.StartManipulation(playerID, ManipulationDump, "cloth", nodeID, 3000.0)
	system.ExecuteManipulation(manip2.ID)
	active := system.GetActiveManipulations(playerID)
	if len(active) != 2 {
		t.Errorf("expected 2 active manipulations, got %d", len(active))
	}
}

// TestSetManipulatorSkill tests setting player skill.
func TestSetManipulatorSkill(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	system.SetManipulatorSkill(playerID, 0.8)
	if system.ManipulatorSkill[playerID] != 0.8 {
		t.Errorf("expected skill 0.8, got %f", system.ManipulatorSkill[playerID])
	}
	// Test clamping
	system.SetManipulatorSkill(playerID, 1.5)
	if system.ManipulatorSkill[playerID] != 1.0 {
		t.Errorf("skill should be clamped to 1.0, got %f", system.ManipulatorSkill[playerID])
	}
}

// TestSetDetectionCapability tests setting node detection capability.
func TestSetDetectionCapability(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	nodeID := ecs.Entity(200)
	system.SetDetectionCapability(nodeID, 0.7)
	if system.DetectionAgents[nodeID] != 0.7 {
		t.Errorf("expected detection 0.7, got %f", system.DetectionAgents[nodeID])
	}
}

// TestEstimateOutcome tests outcome estimation.
func TestEstimateOutcome(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	// High skill should improve estimates
	lowSkillOutcome := system.EstimateOutcome(ManipulationRumor, 1000.0, 0.0)
	highSkillOutcome := system.EstimateOutcome(ManipulationRumor, 1000.0, 1.0)
	if highSkillOutcome.Profit <= lowSkillOutcome.Profit {
		t.Error("higher skill should improve expected profit")
	}
}

// TestGetCooldownRemaining tests cooldown tracking.
func TestGetCooldownRemaining(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	// No cooldown initially
	remaining := system.GetCooldownRemaining(playerID)
	if remaining != 0 {
		t.Errorf("expected 0 cooldown, got %f", remaining)
	}
	// Set future cooldown
	system.GameTime = 1000.0
	system.Cooldowns[playerID] = 2000.0
	remaining = system.GetCooldownRemaining(playerID)
	if remaining != 1000.0 {
		t.Errorf("expected 1000.0 cooldown remaining, got %f", remaining)
	}
}

// TestAllManipulationTypes tests all manipulation type configurations.
func TestAllManipulationTypes(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	types := []ManipulationType{
		ManipulationCorner,
		ManipulationDump,
		ManipulationRumor,
		ManipulationSabotage,
		ManipulationMonopoly,
		ManipulationCartel,
		ManipulationInsider,
		ManipulationCounterfeit,
	}
	for _, mType := range types {
		minInvest := system.getMinimumInvestment(mType)
		if minInvest <= 0 {
			t.Errorf("type %d should have positive minimum investment", mType)
		}
		baseSuccess := system.getBaseSuccessChance(mType)
		if baseSuccess <= 0 || baseSuccess > 1 {
			t.Errorf("type %d success chance should be 0-1, got %f", mType, baseSuccess)
		}
		detectionRisk := system.getBaseDetectionRisk(mType)
		if detectionRisk < 0 || detectionRisk > 1 {
			t.Errorf("type %d detection risk should be 0-1, got %f", mType, detectionRisk)
		}
		duration := system.getManipulationDuration(mType)
		if duration <= 0 {
			t.Errorf("type %d should have positive duration", mType)
		}
		priceEffect := system.getPriceEffect(mType)
		if priceEffect <= 0 {
			t.Errorf("type %d should have positive price effect", mType)
		}
		profitMult := system.getProfitMultiplier(mType)
		if profitMult <= 0 {
			t.Errorf("type %d should have positive profit multiplier", mType)
		}
	}
}

// TestPseudoRandom tests deterministic random generation.
func TestMarketManipulationPseudoRandom(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system1 := NewMarketManipulationSystem(12345, "fantasy", economy)
	system2 := NewMarketManipulationSystem(12345, "fantasy", economy)
	// Same seed should produce same sequence
	for i := 0; i < 10; i++ {
		r1 := system1.pseudoRandom()
		r2 := system2.pseudoRandom()
		if r1 != r2 {
			t.Errorf("iteration %d: same seed should produce same random values", i)
		}
	}
}

// TestManipulationWithSkillBonus tests skill affecting success chance.
func TestManipulationWithSkillBonus(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	nodeID := ecs.Entity(200)
	// Start without skill
	manip1, _ := system.StartManipulation(playerID, ManipulationRumor, "ore", nodeID, 1000.0)
	successNoSkill := manip1.SuccessChance
	// Add skill
	system.SetManipulatorSkill(ecs.Entity(101), 1.0)
	manip2, _ := system.StartManipulation(ecs.Entity(101), ManipulationRumor, "ore", nodeID, 1000.0)
	successWithSkill := manip2.SuccessChance
	if successWithSkill <= successNoSkill {
		t.Error("skill should improve success chance")
	}
}

// TestGetHistory tests manipulation history retrieval.
func TestGetHistory(t *testing.T) {
	economy := NewEconomySystem(0.5, 0.01)
	system := NewMarketManipulationSystem(12345, "fantasy", economy)
	playerID := ecs.Entity(100)
	// Initially empty
	history := system.GetHistory(playerID)
	if len(history) != 0 {
		t.Errorf("initial history should be empty, got %d items", len(history))
	}
	// Add an outcome directly
	outcome := ManipulationOutcome{
		Success: true,
		Profit:  1000.0,
	}
	system.recordOutcomeStruct(playerID, outcome)
	history = system.GetHistory(playerID)
	if len(history) != 1 {
		t.Errorf("expected 1 history item, got %d", len(history))
	}
}
