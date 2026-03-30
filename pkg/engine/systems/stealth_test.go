package systems

import (
	"math"
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestStealthVisibility(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewStealthSystem()
	w.RegisterSystem(sys)

	// Create entity with stealth
	entity := w.CreateEntity()
	_ = w.AddComponent(entity, &components.Stealth{
		BaseVisibility:  1.0,
		SneakVisibility: 0.3,
		Sneaking:        false,
	})

	// Normal visibility
	w.Update(0.016)
	stealthComp, _ := w.GetComponent(entity, "Stealth")
	stealth := stealthComp.(*components.Stealth)
	if stealth.Visibility != 1.0 {
		t.Errorf("Expected visibility 1.0 when not sneaking, got %f", stealth.Visibility)
	}

	// Start sneaking
	stealth.Sneaking = true
	w.Update(0.016)
	if stealth.Visibility != 0.3 {
		t.Errorf("Expected visibility 0.3 when sneaking, got %f", stealth.Visibility)
	}
}

func TestStealthDetection(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewStealthSystem()
	w.RegisterSystem(sys)

	// Create NPC with awareness
	npc := w.CreateEntity()
	_ = w.AddComponent(npc, &components.Position{X: 0, Y: 0, Z: 0, Angle: 0})
	_ = w.AddComponent(npc, &components.Awareness{
		SightRange: 10.0,
		SightAngle: math.Pi / 2, // 90 degree FOV
	})

	// Create player sneaking in front of NPC
	player := w.CreateEntity()
	_ = w.AddComponent(player, &components.Position{X: 5, Y: 0, Z: 0})
	_ = w.AddComponent(player, &components.Stealth{
		BaseVisibility:  1.0,
		SneakVisibility: 0.5,
		Sneaking:        true,
		Visibility:      0.5,
	})

	w.Update(0.016)

	// Check that NPC detected player
	awarenessComp, _ := w.GetComponent(npc, "Awareness")
	awareness := awarenessComp.(*components.Awareness)

	if awareness.AlertLevel == 0 {
		t.Error("NPC should have detected player in front")
	}
}

func TestStealthOutOfSightCone(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewStealthSystem()
	w.RegisterSystem(sys)

	// Create NPC facing east (angle 0)
	npc := w.CreateEntity()
	_ = w.AddComponent(npc, &components.Position{X: 0, Y: 0, Z: 0, Angle: 0})
	_ = w.AddComponent(npc, &components.Awareness{
		SightRange: 10.0,
		SightAngle: math.Pi / 4, // 45 degree FOV
	})

	// Create player behind NPC (west)
	player := w.CreateEntity()
	_ = w.AddComponent(player, &components.Position{X: -5, Y: 0, Z: 0})
	_ = w.AddComponent(player, &components.Stealth{
		BaseVisibility:  1.0,
		SneakVisibility: 0.5,
		Sneaking:        false,
		Visibility:      1.0,
	})

	w.Update(0.016)

	// Check that NPC did NOT detect player behind
	awarenessComp, _ := w.GetComponent(npc, "Awareness")
	awareness := awarenessComp.(*components.Awareness)

	if awareness.AlertLevel > 0 {
		t.Error("NPC should not detect player behind them")
	}
}

func TestBackstab(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewStealthSystem()

	// Create attacker
	attacker := w.CreateEntity()
	_ = w.AddComponent(attacker, &components.Stealth{Sneaking: true})

	// Create unaware target
	target := w.CreateEntity()
	_ = w.AddComponent(target, &components.Awareness{
		AlertLevel:       0,
		DetectedEntities: make(map[uint64]float64),
	})

	// Check target is unaware
	if !sys.IsTargetUnaware(w, attacker, target) {
		t.Error("Target should be unaware")
	}

	// Calculate backstab damage
	baseDamage := 50.0
	backstabDamage := sys.GetBackstabDamage(w, baseDamage, attacker, target)
	expected := baseDamage * sys.BackstabMultiplier

	if backstabDamage != expected {
		t.Errorf("Expected backstab damage %f, got %f", expected, backstabDamage)
	}
}

func TestBackstabAwareTarget(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewStealthSystem()

	// Create attacker
	attacker := w.CreateEntity()

	// Create aware target
	target := w.CreateEntity()
	_ = w.AddComponent(target, &components.Awareness{
		AlertLevel: 1.0,
		DetectedEntities: map[uint64]float64{
			uint64(attacker): 1.0, // Fully aware of attacker
		},
	})

	// Check target is aware
	if sys.IsTargetUnaware(w, attacker, target) {
		t.Error("Target should be aware")
	}

	// Calculate damage (no backstab bonus)
	baseDamage := 50.0
	damage := sys.GetBackstabDamage(w, baseDamage, attacker, target)

	if damage != baseDamage {
		t.Errorf("Expected normal damage %f when target is aware, got %f", baseDamage, damage)
	}
}

func TestPickpocket(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewStealthSystem()

	// Create thief with pickpocket skill
	thief := w.CreateEntity()
	_ = w.AddComponent(thief, &components.Stealth{Sneaking: true})
	_ = w.AddComponent(thief, &components.Skills{
		Levels: map[string]int{
			"pickpocket": 10,
		},
		Experience: make(map[string]float64),
	})

	// Create unaware target
	target := w.CreateEntity()
	_ = w.AddComponent(target, &components.Awareness{})

	// Low difficulty pickpocket should succeed
	if !sys.AttemptPickpocket(w, thief, target, 0.5) {
		t.Error("Pickpocket should succeed with high skill and low difficulty")
	}

	// High difficulty pickpocket should fail
	if sys.AttemptPickpocket(w, thief, target, 2.0) {
		t.Error("Pickpocket should fail with very high difficulty")
	}
}

func TestPickpocketNotSneaking(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewStealthSystem()

	// Create thief NOT sneaking
	thief := w.CreateEntity()
	_ = w.AddComponent(thief, &components.Stealth{Sneaking: false})
	_ = w.AddComponent(thief, &components.Skills{
		Levels: map[string]int{
			"pickpocket": 100,
		},
	})

	// Create unaware target
	target := w.CreateEntity()
	_ = w.AddComponent(target, &components.Awareness{})

	// Should fail because not sneaking
	if sys.AttemptPickpocket(w, thief, target, 0.1) {
		t.Error("Pickpocket should fail when not sneaking")
	}
}

func TestSetSneaking(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewStealthSystem()

	entity := w.CreateEntity()
	_ = w.AddComponent(entity, &components.Stealth{Sneaking: false})

	// Set sneaking on
	result := sys.SetSneaking(w, entity, true)
	if !result {
		t.Error("SetSneaking should succeed")
	}

	stealthComp, _ := w.GetComponent(entity, "Stealth")
	stealth := stealthComp.(*components.Stealth)
	if !stealth.Sneaking {
		t.Error("Entity should be sneaking")
	}

	// Set sneaking off
	sys.SetSneaking(w, entity, false)
	if stealth.Sneaking {
		t.Error("Entity should not be sneaking")
	}
}

func TestStealthComponent(t *testing.T) {
	stealth := &components.Stealth{
		Visibility:      0.5,
		Sneaking:        true,
		DetectionRadius: 5.0,
	}

	if stealth.Type() != "Stealth" {
		t.Errorf("Stealth.Type() = %s, want 'Stealth'", stealth.Type())
	}
}

func TestAwarenessComponent(t *testing.T) {
	awareness := &components.Awareness{
		AlertLevel: 0.5,
		SightRange: 10.0,
		SightAngle: math.Pi / 2,
	}

	if awareness.Type() != "Awareness" {
		t.Errorf("Awareness.Type() = %s, want 'Awareness'", awareness.Type())
	}
}

func TestAlertDecay(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewStealthSystem()
	w.RegisterSystem(sys)

	// Create NPC with alert level
	npc := w.CreateEntity()
	_ = w.AddComponent(npc, &components.Awareness{
		AlertLevel: 1.0,
	})

	// Run several updates to decay alert
	for i := 0; i < 100; i++ {
		w.Update(0.1) // 10 seconds total
	}

	awarenessComp, _ := w.GetComponent(npc, "Awareness")
	awareness := awarenessComp.(*components.Awareness)

	if awareness.AlertLevel >= 1.0 {
		t.Error("Alert level should have decayed")
	}
}

// ============================================================================
// Hiding Spot System Tests
// ============================================================================

func TestNewHidingSpotSystem(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	if sys == nil {
		t.Fatal("System should not be nil")
	}
	if sys.ChunkSize != 512 {
		t.Errorf("Chunk size should be 512, got %f", sys.ChunkSize)
	}
}

func TestHidingSpotRegister(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	spot := sys.CreateHidingSpotFromEnvironment("test_spot", HidingSpotShadow, 100, 100, 0)
	sys.RegisterHidingSpot(spot)

	if sys.SpotCount() != 1 {
		t.Errorf("Should have 1 spot, got %d", sys.SpotCount())
	}

	retrieved := sys.GetHidingSpot("test_spot")
	if retrieved == nil {
		t.Error("Should retrieve registered spot")
	}
}

func TestHidingSpotTypes(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	types := []HidingSpotType{
		HidingSpotShadow, HidingSpotFoliage, HidingSpotContainer,
		HidingSpotFurniture, HidingSpotArchitectural, HidingSpotRooftop,
		HidingSpotUnderwater,
	}

	for _, spotType := range types {
		spot := sys.CreateHidingSpotFromEnvironment(string(spotType), spotType, 0, 0, 0)
		if spot == nil {
			t.Errorf("Should create spot of type %s", spotType)
		}
		if spot.Radius <= 0 {
			t.Errorf("Spot %s should have positive radius", spotType)
		}
		if spot.VisibilityReduction <= 0 || spot.VisibilityReduction > 1 {
			t.Errorf("Spot %s visibility reduction invalid: %f", spotType, spot.VisibilityReduction)
		}
	}
}

func TestHidingSpotEnterExit(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	spot := sys.CreateHidingSpotFromEnvironment("test_spot", HidingSpotFoliage, 100, 100, 0)
	sys.RegisterHidingSpot(spot)

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Add required components
	pos := &components.Position{X: 100, Y: 100, Z: 0}
	world.AddComponent(player, pos)

	stealth := &components.Stealth{
		BaseVisibility:  1.0,
		SneakVisibility: 0.3,
		Sneaking:        false,
	}
	world.AddComponent(player, stealth)

	skills := &components.Skills{
		Levels:     map[string]int{"sneak": 10},
		Experience: make(map[string]float64),
	}
	world.AddComponent(player, skills)

	// Enter hiding spot
	entered := sys.EnterHidingSpot(world, player, "test_spot")
	if !entered {
		t.Fatal("Should enter hiding spot")
	}

	if !sys.IsHiding(player) {
		t.Error("Player should be hiding")
	}

	currentSpot := sys.GetCurrentSpot(player)
	if currentSpot == nil || currentSpot.ID != "test_spot" {
		t.Error("Should be in test_spot")
	}

	// Exit hiding spot
	exited := sys.ExitHidingSpot(world, player)
	if !exited {
		t.Error("Should exit hiding spot")
	}

	if sys.IsHiding(player) {
		t.Error("Player should no longer be hiding")
	}
}

func TestHidingSpotCapacity(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	// Container has capacity 1
	spot := sys.CreateHidingSpotFromEnvironment("small_box", HidingSpotContainer, 100, 100, 0)
	sys.RegisterHidingSpot(spot)

	world := ecs.NewWorld()

	// Create two players
	player1 := world.CreateEntity()
	pos1 := &components.Position{X: 100, Y: 100, Z: 0}
	world.AddComponent(player1, pos1)
	world.AddComponent(player1, &components.Stealth{})
	world.AddComponent(player1, &components.Skills{Levels: map[string]int{"sneak": 20}})

	player2 := world.CreateEntity()
	pos2 := &components.Position{X: 100, Y: 100, Z: 0}
	world.AddComponent(player2, pos2)
	world.AddComponent(player2, &components.Stealth{})
	world.AddComponent(player2, &components.Skills{Levels: map[string]int{"sneak": 20}})

	// First player enters
	entered1 := sys.EnterHidingSpot(world, player1, "small_box")
	if !entered1 {
		t.Error("First player should enter")
	}

	// Second player should fail (capacity full)
	entered2 := sys.EnterHidingSpot(world, player2, "small_box")
	if entered2 {
		t.Error("Second player should not enter (at capacity)")
	}
}

func TestHidingSpotSkillRequirement(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	// Rooftop requires sneak 20
	spot := sys.CreateHidingSpotFromEnvironment("roof", HidingSpotRooftop, 100, 100, 0)
	sys.RegisterHidingSpot(spot)

	world := ecs.NewWorld()

	// Low skill player
	player := world.CreateEntity()
	world.AddComponent(player, &components.Position{X: 100, Y: 100, Z: 0})
	world.AddComponent(player, &components.Stealth{})
	world.AddComponent(player, &components.Skills{Levels: map[string]int{"sneak": 5}})

	canHide := sys.CanHide(world, player, "roof")
	if canHide {
		t.Error("Low skill player should not be able to use rooftop")
	}

	// High skill player
	player2 := world.CreateEntity()
	world.AddComponent(player2, &components.Position{X: 100, Y: 100, Z: 0})
	world.AddComponent(player2, &components.Stealth{})
	world.AddComponent(player2, &components.Skills{Levels: map[string]int{"sneak": 30}})

	canHide2 := sys.CanHide(world, player2, "roof")
	if !canHide2 {
		t.Error("High skill player should be able to use rooftop")
	}
}

func TestHidingSpotSearch(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	// Foliage can be searched
	spot := sys.CreateHidingSpotFromEnvironment("bush", HidingSpotFoliage, 100, 100, 0)
	sys.RegisterHidingSpot(spot)

	world := ecs.NewWorld()

	// Hider with low sneak
	hider := world.CreateEntity()
	world.AddComponent(hider, &components.Position{X: 100, Y: 100, Z: 0})
	world.AddComponent(hider, &components.Stealth{})
	world.AddComponent(hider, &components.Skills{Levels: map[string]int{"sneak": 10}})

	sys.EnterHidingSpot(world, hider, "bush")

	// Searcher with high perception
	searcher := world.CreateEntity()
	world.AddComponent(searcher, &components.Position{X: 105, Y: 100, Z: 0})
	world.AddComponent(searcher, &components.Skills{Levels: map[string]int{"perception": 50}})

	found := sys.SearchHidingSpot(world, searcher, "bush")
	if len(found) == 0 {
		t.Error("High perception searcher should find low sneak hider")
	}

	// Hider should be revealed
	if sys.IsHiding(hider) {
		t.Error("Found hider should no longer be hiding")
	}
}

func TestHidingSpotNearby(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	// Add several spots
	sys.RegisterHidingSpot(sys.CreateHidingSpotFromEnvironment("spot1", HidingSpotShadow, 100, 100, 0))
	sys.RegisterHidingSpot(sys.CreateHidingSpotFromEnvironment("spot2", HidingSpotFoliage, 110, 100, 0))
	sys.RegisterHidingSpot(sys.CreateHidingSpotFromEnvironment("spot3", HidingSpotContainer, 1000, 1000, 0))

	nearby := sys.GetSpotsNear(105, 100, 20)
	if len(nearby) < 2 {
		t.Errorf("Should find at least 2 nearby spots, found %d", len(nearby))
	}
}

func TestHidingSpotByType(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	sys.RegisterHidingSpot(sys.CreateHidingSpotFromEnvironment("shadow1", HidingSpotShadow, 0, 0, 0))
	sys.RegisterHidingSpot(sys.CreateHidingSpotFromEnvironment("shadow2", HidingSpotShadow, 100, 0, 0))
	sys.RegisterHidingSpot(sys.CreateHidingSpotFromEnvironment("bush", HidingSpotFoliage, 50, 0, 0))

	shadows := sys.GetSpotsByType(HidingSpotShadow)
	if len(shadows) != 2 {
		t.Errorf("Should have 2 shadow spots, got %d", len(shadows))
	}
}

func TestHidingSpotUpdate(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	spot := sys.CreateHidingSpotFromEnvironment("shadow", HidingSpotShadow, 100, 100, 0)
	sys.RegisterHidingSpot(spot)

	world := ecs.NewWorld()
	player := world.CreateEntity()
	world.AddComponent(player, &components.Position{X: 100, Y: 100, Z: 0})
	stealth := &components.Stealth{
		BaseVisibility:  1.0,
		SneakVisibility: 0.3,
		Sneaking:        false,
	}
	world.AddComponent(player, stealth)
	world.AddComponent(player, &components.Skills{Levels: map[string]int{"sneak": 10}})

	sys.EnterHidingSpot(world, player, "shadow")

	// Update should apply hiding effects
	sys.Update(world, 0.016)

	comp, _ := world.GetComponent(player, "Stealth")
	updatedStealth := comp.(*components.Stealth)

	// Visibility should be reduced while hiding
	if updatedStealth.Visibility >= 0.3 {
		t.Errorf("Visibility should be reduced below 0.3, got %f", updatedStealth.Visibility)
	}
}

func TestHidingSpotNearest(t *testing.T) {
	sys := NewHidingSpotSystem(512)

	sys.RegisterHidingSpot(sys.CreateHidingSpotFromEnvironment("far", HidingSpotShadow, 200, 200, 0))
	sys.RegisterHidingSpot(sys.CreateHidingSpotFromEnvironment("near", HidingSpotFoliage, 105, 100, 0))

	world := ecs.NewWorld()
	player := world.CreateEntity()
	world.AddComponent(player, &components.Position{X: 100, Y: 100, Z: 0})
	world.AddComponent(player, &components.Stealth{})
	world.AddComponent(player, &components.Skills{Levels: map[string]int{"sneak": 10}})

	nearest := sys.GetNearestAvailableSpot(world, player)
	if nearest == nil {
		t.Fatal("Should find nearest spot")
	}
	if nearest.ID != "near" {
		t.Errorf("Nearest should be 'near', got %s", nearest.ID)
	}
}

func BenchmarkHidingSpotSystem(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewHidingSpotSystem(512)
	}
}

func BenchmarkHidingSpotGetNear(b *testing.B) {
	sys := NewHidingSpotSystem(512)

	// Add many spots
	for i := 0; i < 1000; i++ {
		x := float64(i % 100 * 50)
		y := float64(i / 100 * 50)
		spot := sys.CreateHidingSpotFromEnvironment(
			"spot_"+string(rune(i)),
			HidingSpotShadow,
			x, y, 0,
		)
		sys.RegisterHidingSpot(spot)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GetSpotsNear(250, 250, 100)
	}
}

// ============================================================================
// Distraction System Tests
// ============================================================================

func TestNewDistractionSystem(t *testing.T) {
	sys := NewDistractionSystem()

	if sys == nil {
		t.Fatal("System should not be nil")
	}

	if sys.GetTemplateCount() == 0 {
		t.Error("Should have default templates")
	}
}

func TestDistractionTemplates(t *testing.T) {
	sys := NewDistractionSystem()

	templates := []string{
		"thrown_rock", "whistle", "broken_glass", "animal_call", "explosion",
		"torch_flicker", "smoke_signal", "bright_flash",
		"small_fire", "alarm_bell", "trap_triggered",
		"loud_argument", "bard_performance", "scream",
		"thunder", "falling_debris", "animal_panic",
	}

	for _, id := range templates {
		template := sys.DistractionTemplates[id]
		if template == nil {
			t.Errorf("Template %s should exist", id)
		}
	}
}

func TestDistractionCreate(t *testing.T) {
	sys := NewDistractionSystem()

	distraction := sys.CreateDistraction("thrown_rock", 100, 100, 0, 0)
	if distraction == nil {
		t.Fatal("Should create distraction")
	}

	if distraction.Position.X != 100 || distraction.Position.Y != 100 {
		t.Error("Position should match creation params")
	}

	if distraction.TimeRemaining <= 0 {
		t.Error("Should have positive time remaining")
	}

	if sys.GetActiveCount() != 1 {
		t.Errorf("Should have 1 active distraction, got %d", sys.GetActiveCount())
	}
}

func TestDistractionTypes(t *testing.T) {
	sys := NewDistractionSystem()

	// Create one of each type
	sys.CreateDistraction("thrown_rock", 0, 0, 0, 0)     // sound
	sys.CreateDistraction("torch_flicker", 100, 0, 0, 0) // visual
	sys.CreateDistraction("small_fire", 200, 0, 0, 0)    // tactical
	sys.CreateDistraction("loud_argument", 300, 0, 0, 0) // social
	sys.CreateDistraction("thunder", 400, 0, 0, 0)       // environmental

	soundDistractions := sys.GetDistractionsByType(DistractionSound)
	if len(soundDistractions) < 1 {
		t.Error("Should have at least 1 sound distraction")
	}

	visualDistractions := sys.GetDistractionsByType(DistractionVisual)
	if len(visualDistractions) < 1 {
		t.Error("Should have at least 1 visual distraction")
	}
}

func TestDistractionExpiration(t *testing.T) {
	sys := NewDistractionSystem()

	distraction := sys.CreateDistraction("thrown_rock", 100, 100, 0, 0)
	initialTime := distraction.TimeRemaining

	// Simulate time passing
	sys.updateDistractionTimers(1.0)

	if distraction.TimeRemaining >= initialTime {
		t.Error("Time remaining should decrease")
	}

	// Fast forward past duration
	sys.updateDistractionTimers(100.0)
	sys.cleanupExpiredDistractions()

	if sys.GetActiveCount() > 0 {
		t.Error("Expired distraction should be cleaned up")
	}
}

func TestDistractionNPCReaction(t *testing.T) {
	sys := NewDistractionSystem()

	world := ecs.NewWorld()
	npc := world.CreateEntity()

	world.AddComponent(npc, &components.Position{X: 100, Y: 100, Z: 0})
	world.AddComponent(npc, &components.Awareness{
		SightRange: 50, SightAngle: 3.14, AlertLevel: 0,
	})

	// Create distraction near NPC
	sys.CreateDistraction("whistle", 110, 100, 0, 0)

	// Process NPC reactions
	sys.Update(world, 0.016)

	if !sys.IsNPCDistracted(npc) {
		t.Error("NPC should be distracted by nearby whistle")
	}

	distraction := sys.GetNPCDistraction(npc)
	if distraction == nil {
		t.Error("Should return the distraction being investigated")
	}
}

func TestDistractionIgnored(t *testing.T) {
	sys := NewDistractionSystem()

	world := ecs.NewWorld()
	npc := world.CreateEntity()

	world.AddComponent(npc, &components.Position{X: 100, Y: 100, Z: 0})
	// High alert level NPC
	world.AddComponent(npc, &components.Awareness{
		SightRange: 50, SightAngle: 3.14, AlertLevel: 100,
	})

	// Create ignorable distraction
	sys.CreateDistraction("thrown_rock", 105, 100, 0, 0)

	sys.Update(world, 0.016)

	if sys.IsNPCDistracted(npc) {
		t.Error("High-alert NPC should ignore low-priority distraction")
	}
}

func TestDistractionOutOfRange(t *testing.T) {
	sys := NewDistractionSystem()

	world := ecs.NewWorld()
	npc := world.CreateEntity()

	world.AddComponent(npc, &components.Position{X: 100, Y: 100, Z: 0})
	world.AddComponent(npc, &components.Awareness{
		SightRange: 50, SightAngle: 3.14, AlertLevel: 0,
	})

	// Create distraction far from NPC (thrown_rock has radius 15)
	sys.CreateDistraction("thrown_rock", 200, 200, 0, 0)

	sys.Update(world, 0.016)

	if sys.IsNPCDistracted(npc) {
		t.Error("NPC should not react to out-of-range distraction")
	}
}

func TestDistractionInvestigationProgress(t *testing.T) {
	sys := NewDistractionSystem()

	world := ecs.NewWorld()
	npc := world.CreateEntity()

	world.AddComponent(npc, &components.Position{X: 100, Y: 100, Z: 0})
	world.AddComponent(npc, &components.Awareness{
		SightRange: 100, SightAngle: 3.14, AlertLevel: 0,
	})

	// small_fire has Duration: 60 and InvestigationTime: 20
	sys.CreateDistraction("small_fire", 105, 100, 0, 0)

	sys.Update(world, 0.016) // First update starts investigation
	if !sys.IsNPCDistracted(npc) {
		t.Fatal("NPC should be investigating")
	}

	// Progress after 10 seconds should be ~50%
	sys.Update(world, 10.0)
	progress := sys.GetInvestigationProgress(npc)
	if progress < 0.4 || progress > 0.6 {
		t.Errorf("Progress should be ~0.5, got %f", progress)
	}

	// Complete investigation (need 10 more seconds)
	sys.Update(world, 15.0)
	if sys.IsNPCDistracted(npc) {
		t.Error("Investigation should be complete")
	}
}

func TestDistractionCancel(t *testing.T) {
	sys := NewDistractionSystem()

	world := ecs.NewWorld()
	npc := world.CreateEntity()

	world.AddComponent(npc, &components.Position{X: 100, Y: 100, Z: 0})
	world.AddComponent(npc, &components.Awareness{
		SightRange: 50, SightAngle: 3.14, AlertLevel: 0,
	})

	distraction := sys.CreateDistraction("whistle", 105, 100, 0, 0)
	sys.Update(world, 0.016)

	if !sys.IsNPCDistracted(npc) {
		t.Fatal("NPC should be investigating")
	}

	sys.CancelDistraction(distraction.ID)

	sys.Update(world, 0.016)
	if sys.IsNPCDistracted(npc) {
		t.Error("NPC should stop investigating cancelled distraction")
	}
}

func TestDistractionNear(t *testing.T) {
	sys := NewDistractionSystem()

	sys.CreateDistraction("thrown_rock", 100, 100, 0, 0)
	sys.CreateDistraction("whistle", 110, 100, 0, 0)
	sys.CreateDistraction("explosion", 500, 500, 0, 0)

	nearby := sys.GetDistractionsNear(105, 100, 20)
	if len(nearby) < 2 {
		t.Errorf("Should find at least 2 nearby distractions, got %d", len(nearby))
	}
}

func TestDistractionHelpers(t *testing.T) {
	sys := NewDistractionSystem()

	world := ecs.NewWorld()
	player := world.CreateEntity()

	// Test helper functions
	rock := sys.ThrowDistraction(world, player, 100, 100, 0)
	if rock == nil || rock.Type != DistractionSound {
		t.Error("ThrowDistraction should create sound distraction")
	}

	whistle := sys.CreateWhistle(world, player, 200, 100, 0)
	if whistle == nil {
		t.Error("CreateWhistle should create distraction")
	}

	alarm := sys.TriggerAlarm(world, 300, 100, 0)
	if alarm == nil || alarm.Priority < 5 {
		t.Error("TriggerAlarm should create high-priority distraction")
	}

	fire := sys.CreateFire(world, 400, 100, 0, player)
	if fire == nil || fire.Type != DistractionTactical {
		t.Error("CreateFire should create tactical distraction")
	}
}

func TestDistractionPriority(t *testing.T) {
	sys := NewDistractionSystem()

	world := ecs.NewWorld()
	npc := world.CreateEntity()

	world.AddComponent(npc, &components.Position{X: 100, Y: 100, Z: 0})
	world.AddComponent(npc, &components.Awareness{
		SightRange: 100, SightAngle: 3.14, AlertLevel: 0,
	})

	// Create low and high priority distractions
	sys.CreateDistraction("thrown_rock", 105, 100, 0, 0) // priority 1
	sys.CreateDistraction("explosion", 110, 100, 0, 0)   // priority 5

	sys.Update(world, 0.016)

	distraction := sys.GetNPCDistraction(npc)
	if distraction == nil {
		t.Fatal("NPC should be investigating")
	}

	// Should prioritize explosion (higher priority)
	if distraction.Priority < 4 {
		t.Error("Should investigate higher priority distraction first")
	}
}

func BenchmarkDistractionSystem(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewDistractionSystem()
	}
}

func BenchmarkDistractionUpdate(b *testing.B) {
	sys := NewDistractionSystem()

	world := ecs.NewWorld()
	for i := 0; i < 100; i++ {
		npc := world.CreateEntity()
		world.AddComponent(npc, &components.Position{X: float64(i * 10), Y: 100, Z: 0})
		world.AddComponent(npc, &components.Awareness{
			SightRange: 50, SightAngle: 3.14, AlertLevel: 0,
		})
	}

	// Create several distractions
	for i := 0; i < 20; i++ {
		sys.CreateDistraction("thrown_rock", float64(i*50), 100, 0, 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world, 0.016)
	}
}
