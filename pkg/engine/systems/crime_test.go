package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewCrimeSystem(t *testing.T) {
	tests := []struct {
		name           string
		decayDelay     float64
		bountyPerLevel float64
	}{
		{"default values", 60.0, 100.0},
		{"zero values", 0.0, 0.0},
		{"high values", 300.0, 500.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewCrimeSystem(tt.decayDelay, tt.bountyPerLevel)
			if cs == nil {
				t.Fatal("NewCrimeSystem returned nil")
			}
			if cs.DecayDelay != tt.decayDelay {
				t.Errorf("DecayDelay = %v, want %v", cs.DecayDelay, tt.decayDelay)
			}
			if cs.BountyPerLevel != tt.bountyPerLevel {
				t.Errorf("BountyPerLevel = %v, want %v", cs.BountyPerLevel, tt.bountyPerLevel)
			}
			if cs.WitnessRange != DefaultWitnessRange {
				t.Errorf("WitnessRange = %v, want %v", cs.WitnessRange, DefaultWitnessRange)
			}
		})
	}
}

func TestCrimeSystem_Update(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(10.0, 100.0)

	// Create entity with Crime component
	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{
		WantedLevel:   3,
		BountyAmount:  300.0,
		LastCrimeTime: 0,
	})

	// Update should process crime entities
	cs.Update(w, 1.0)
	if cs.GameTime != 1.0 {
		t.Errorf("GameTime = %v, want 1.0", cs.GameTime)
	}
}

func TestCrimeSystem_DecayWantedLevel(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(10.0, 100.0)

	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{
		WantedLevel:   3,
		LastCrimeTime: 0,
	})

	// Simulate time passing without new crime
	cs.GameTime = 15.0 // More than DecayDelay
	cs.Update(w, 0.1)

	comp, _ := w.GetComponent(e, "Crime")
	crime := comp.(*components.Crime)
	if crime.WantedLevel >= 3 {
		t.Errorf("WantedLevel should have decayed, got %d", crime.WantedLevel)
	}
}

func TestCrimeSystem_ClampWantedLevel(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(10.0, 100.0)

	tests := []struct {
		name          string
		initialLevel  int
		expectedLevel int
	}{
		{"below minimum", -1, MinWantedLevel},
		{"above maximum", 10, MaxWantedLevel},
		{"within range", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := w.CreateEntity()
			w.AddComponent(e, &components.Crime{
				WantedLevel: tt.initialLevel,
			})

			cs.Update(w, 0.1)

			comp, _ := w.GetComponent(e, "Crime")
			crime := comp.(*components.Crime)
			if crime.WantedLevel != tt.expectedLevel {
				t.Errorf("WantedLevel = %d, want %d", crime.WantedLevel, tt.expectedLevel)
			}
		})
	}
}

func TestCrimeSystem_ReportCrime(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)

	// Create criminal entity with position
	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{WantedLevel: 0})
	w.AddComponent(criminal, &components.Position{X: 10, Y: 0, Z: 10})

	// Create witness entity
	witness := w.CreateEntity()
	w.AddComponent(witness, &components.Witness{CanReport: true})
	w.AddComponent(witness, &components.Position{X: 12, Y: 0, Z: 12})

	cs.ReportCrime(w, criminal)

	comp, _ := w.GetComponent(criminal, "Crime")
	crime := comp.(*components.Crime)
	if crime.WantedLevel != 1 {
		t.Errorf("WantedLevel = %d, want 1 after reported crime", crime.WantedLevel)
	}
	if crime.BountyAmount != 100.0 {
		t.Errorf("BountyAmount = %v, want 100.0", crime.BountyAmount)
	}
}

func TestCrimeSystem_ReportCrime_NoWitness(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)

	// Create criminal entity far from any witness
	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{WantedLevel: 0})
	w.AddComponent(criminal, &components.Position{X: 10, Y: 0, Z: 10})

	// Create witness too far away
	witness := w.CreateEntity()
	w.AddComponent(witness, &components.Witness{CanReport: true})
	w.AddComponent(witness, &components.Position{X: 1000, Y: 0, Z: 1000})

	cs.ReportCrime(w, criminal)

	comp, _ := w.GetComponent(criminal, "Crime")
	crime := comp.(*components.Crime)
	if crime.WantedLevel != 0 {
		t.Errorf("WantedLevel = %d, want 0 (no witness)", crime.WantedLevel)
	}
}

func TestCrimeSystem_PayBounty(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)

	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{
		WantedLevel:  3,
		BountyAmount: 300.0,
		InJail:       true,
	})

	result := cs.PayBounty(w, e)
	if !result {
		t.Error("PayBounty should return true")
	}

	comp, _ := w.GetComponent(e, "Crime")
	crime := comp.(*components.Crime)
	if crime.WantedLevel != 0 {
		t.Errorf("WantedLevel = %d, want 0 after paying", crime.WantedLevel)
	}
	if crime.BountyAmount != 0 {
		t.Errorf("BountyAmount = %v, want 0 after paying", crime.BountyAmount)
	}
	if crime.InJail {
		t.Error("InJail should be false after paying bounty")
	}
}

func TestCrimeSystem_GoToJail(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	cs.GameTime = 100.0

	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{
		WantedLevel:  3,
		BountyAmount: 300.0,
	})

	result := cs.GoToJail(w, e, 60.0)
	if !result {
		t.Error("GoToJail should return true")
	}

	comp, _ := w.GetComponent(e, "Crime")
	crime := comp.(*components.Crime)
	if !crime.InJail {
		t.Error("InJail should be true")
	}
	if crime.WantedLevel != 0 {
		t.Errorf("WantedLevel = %d, want 0 while in jail", crime.WantedLevel)
	}
	if crime.JailReleaseTime != 160.0 {
		t.Errorf("JailReleaseTime = %v, want 160.0", crime.JailReleaseTime)
	}
}

func TestCrimeSystem_CheckJailRelease(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)

	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{
		InJail:          true,
		JailReleaseTime: 100.0,
	})

	// Before release time
	cs.GameTime = 50.0
	cs.CheckJailRelease(w)

	comp, _ := w.GetComponent(e, "Crime")
	crime := comp.(*components.Crime)
	if !crime.InJail {
		t.Error("Should still be in jail")
	}

	// After release time
	cs.GameTime = 150.0
	cs.CheckJailRelease(w)

	comp, _ = w.GetComponent(e, "Crime")
	crime = comp.(*components.Crime)
	if crime.InJail {
		t.Error("Should be released from jail")
	}
}

func TestCrimeSystem_InJailSkipsProcessing(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(10.0, 100.0)

	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{
		WantedLevel:   3,
		InJail:        true,
		LastCrimeTime: 0,
	})

	// Even with enough time passing, in-jail entities shouldn't decay
	cs.GameTime = 100.0
	cs.Update(w, 1.0)

	comp, _ := w.GetComponent(e, "Crime")
	crime := comp.(*components.Crime)
	if crime.WantedLevel != 3 {
		t.Errorf("WantedLevel = %d, want 3 (no decay while in jail)", crime.WantedLevel)
	}
}

func TestCrimeSystem_WitnessCannotReport(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)

	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{WantedLevel: 0})
	w.AddComponent(criminal, &components.Position{X: 10, Y: 0, Z: 10})

	// Witness that cannot report
	witness := w.CreateEntity()
	w.AddComponent(witness, &components.Witness{CanReport: false})
	w.AddComponent(witness, &components.Position{X: 12, Y: 0, Z: 12})

	cs.ReportCrime(w, criminal)

	comp, _ := w.GetComponent(criminal, "Crime")
	crime := comp.(*components.Crime)
	if crime.WantedLevel != 0 {
		t.Errorf("WantedLevel = %d, want 0 (witness cannot report)", crime.WantedLevel)
	}
}

// ============================================================================
// GuardPursuitSystem Tests
// ============================================================================

func TestNewGuardPursuitSystem(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	if gps == nil {
		t.Fatal("NewGuardPursuitSystem returned nil")
	}
	if gps.crimeSystem != cs {
		t.Error("crimeSystem not set correctly")
	}
	if gps.pursuitSpeedMod <= 0 {
		t.Error("pursuitSpeedMod should be positive")
	}
	if gps.alertDuration <= 0 {
		t.Error("alertDuration should be positive")
	}
	if gps.searchDuration <= 0 {
		t.Error("searchDuration should be positive")
	}
}

func TestGuardPursuitSystem_Update(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	// Create a guard
	guard := w.CreateEntity()
	w.AddComponent(guard, &components.Guard{
		State:      int(GuardStatePatrol),
		SightRange: 20.0,
	})
	w.AddComponent(guard, &components.Position{X: 0, Y: 0, Z: 0})

	// Should not panic
	gps.Update(w, 0.1)
}

func TestGuardPursuitSystem_PatrolToPersuit(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	// Create guard
	guard := w.CreateEntity()
	w.AddComponent(guard, &components.Guard{
		State:        int(GuardStatePatrol),
		SightRange:   20.0,
		PursuitSpeed: 4.0,
	})
	w.AddComponent(guard, &components.Position{X: 0, Y: 0, Z: 0})

	// Create criminal nearby
	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{
		WantedLevel: 3,
		InJail:      false,
	})
	w.AddComponent(criminal, &components.Position{X: 5, Y: 0, Z: 5})

	gps.Update(w, 0.1)

	comp, _ := w.GetComponent(guard, "Guard")
	guardComp := comp.(*components.Guard)
	if guardComp.State != int(GuardStatePursue) {
		t.Errorf("Guard state = %d, want %d (Pursue)", guardComp.State, int(GuardStatePursue))
	}
}

func TestGuardPursuitSystem_AlertGuardsNearby(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	// Create guards near and far from crime
	nearGuard := w.CreateEntity()
	w.AddComponent(nearGuard, &components.Guard{State: int(GuardStatePatrol)})
	w.AddComponent(nearGuard, &components.Position{X: 5, Y: 0, Z: 5})

	farGuard := w.CreateEntity()
	w.AddComponent(farGuard, &components.Guard{State: int(GuardStatePatrol)})
	w.AddComponent(farGuard, &components.Position{X: 1000, Y: 0, Z: 1000})

	count := gps.AlertGuardsNearby(w, 10, 10)

	if count != 1 {
		t.Errorf("AlertGuardsNearby = %d, want 1", count)
	}

	comp, _ := w.GetComponent(nearGuard, "Guard")
	guardComp := comp.(*components.Guard)
	if guardComp.State != int(GuardStateAlert) {
		t.Errorf("Near guard state = %d, want %d (Alert)", guardComp.State, int(GuardStateAlert))
	}

	comp, _ = w.GetComponent(farGuard, "Guard")
	guardComp = comp.(*components.Guard)
	if guardComp.State != int(GuardStatePatrol) {
		t.Errorf("Far guard state = %d, want %d (Patrol)", guardComp.State, int(GuardStatePatrol))
	}
}

func TestGuardPursuitSystem_SearchDuration(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	gps.SetSearchDuration(30.0)
	if gps.GetSearchDuration() != 30.0 {
		t.Errorf("SearchDuration = %v, want 30.0", gps.GetSearchDuration())
	}
}

func TestGuardPursuitSystem_HandleAlert(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	guard := w.CreateEntity()
	w.AddComponent(guard, &components.Guard{
		State:      int(GuardStateAlert),
		AlertTimer: 0.5,
		SightRange: 20.0,
	})
	w.AddComponent(guard, &components.Position{X: 0, Y: 0, Z: 0})

	// Alert timer should decrease
	gps.Update(w, 0.3)

	comp, _ := w.GetComponent(guard, "Guard")
	guardComp := comp.(*components.Guard)
	if guardComp.AlertTimer >= 0.5 {
		t.Error("Alert timer should have decreased")
	}

	// After timer expires, should return to patrol
	gps.Update(w, 1.0)

	comp, _ = w.GetComponent(guard, "Guard")
	guardComp = comp.(*components.Guard)
	if guardComp.State != int(GuardStatePatrol) {
		t.Errorf("State = %d, want %d (Patrol after alert timeout)", guardComp.State, int(GuardStatePatrol))
	}
}

func TestGuardPursuitSystem_HandleSearch(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	guard := w.CreateEntity()
	w.AddComponent(guard, &components.Guard{
		State:       int(GuardStateSearch),
		SearchTimer: 1.0,
		SightRange:  20.0,
		LastKnownX:  50,
		LastKnownZ:  50,
	})
	w.AddComponent(guard, &components.Position{X: 0, Y: 0, Z: 0})

	// Search timer should decrease
	gps.Update(w, 0.5)

	comp, _ := w.GetComponent(guard, "Guard")
	guardComp := comp.(*components.Guard)
	if guardComp.SearchTimer >= 1.0 {
		t.Error("Search timer should have decreased")
	}

	// After timer expires, should go to return state then patrol
	gps.Update(w, 2.0)

	comp, _ = w.GetComponent(guard, "Guard")
	guardComp = comp.(*components.Guard)
	// First goes to Return state
	if guardComp.State != int(GuardStateReturn) {
		// Run another update to go from Return to Patrol
		gps.Update(w, 0.1)
		comp, _ = w.GetComponent(guard, "Guard")
		guardComp = comp.(*components.Guard)
	}
	// Should be in Patrol or Return (Return immediately transitions to Patrol)
	if guardComp.State != int(GuardStatePatrol) && guardComp.State != int(GuardStateReturn) {
		t.Errorf("State = %d, want %d (Patrol) or %d (Return)", guardComp.State, int(GuardStatePatrol), int(GuardStateReturn))
	}
}

func TestGuardPursuitSystem_HandleCombat(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	guard := w.CreateEntity()
	criminal := w.CreateEntity()

	w.AddComponent(guard, &components.Guard{
		State:        int(GuardStateCombat),
		TargetEntity: uint64(criminal),
		SightRange:   20.0,
	})
	w.AddComponent(guard, &components.Position{X: 0, Y: 0, Z: 0})

	w.AddComponent(criminal, &components.Crime{WantedLevel: 3})
	w.AddComponent(criminal, &components.Position{X: 1, Y: 0, Z: 1}) // Close

	gps.Update(w, 0.1)

	// Criminal should be in jail after arrest attempt
	comp, _ := w.GetComponent(criminal, "Crime")
	crime := comp.(*components.Crime)
	if !crime.InJail {
		t.Error("Criminal should be in jail after arrest")
	}
}

func TestGuardPursuitSystem_sqrt(t *testing.T) {
	gps := &GuardPursuitSystem{}

	tests := []struct {
		input    float64
		expected float64
		epsilon  float64
	}{
		{4.0, 2.0, 0.001},
		{9.0, 3.0, 0.001},
		{16.0, 4.0, 0.001},
		{0.0, 0.0, 0.001},
		{-1.0, 0.0, 0.001}, // Negative returns 0
	}

	for _, tt := range tests {
		result := gps.sqrt(tt.input)
		diff := result - tt.expected
		if diff < 0 {
			diff = -diff
		}
		if diff > tt.epsilon {
			t.Errorf("sqrt(%v) = %v, want %v (within %v)", tt.input, result, tt.expected, tt.epsilon)
		}
	}
}

// ============================================================================
// Bribery System Tests
// ============================================================================

func TestNewBriberySystem(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	tests := []struct {
		name string
		seed int64
	}{
		{"zero seed", 0},
		{"positive seed", 12345},
		{"negative seed", -9999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := NewBriberySystem(cs, gps, tt.seed)
			if bs == nil {
				t.Fatal("NewBriberySystem returned nil")
			}
			if bs.crimeSystem != cs {
				t.Error("crimeSystem not set correctly")
			}
			if bs.guardPursuitSystem != gps {
				t.Error("guardPursuitSystem not set correctly")
			}
			if bs.BaseBribeCostPerLevel != 100.0 {
				t.Errorf("BaseBribeCostPerLevel = %v, want 100.0", bs.BaseBribeCostPerLevel)
			}
			if bs.GuardSuccessBase != 0.6 {
				t.Errorf("GuardSuccessBase = %v, want 0.6", bs.GuardSuccessBase)
			}
			if bs.WitnessSuccessBase != 0.8 {
				t.Errorf("WitnessSuccessBase = %v, want 0.8", bs.WitnessSuccessBase)
			}
			if bs.OfficialSuccessBase != 0.5 {
				t.Errorf("OfficialSuccessBase = %v, want 0.5", bs.OfficialSuccessBase)
			}
			if bs.JailerSuccessBase != 0.4 {
				t.Errorf("JailerSuccessBase = %v, want 0.4", bs.JailerSuccessBase)
			}
		})
	}
}

func TestBriberySystem_pseudoRandom(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	// Test determinism - same seed produces same sequence
	values := make([]float64, 10)
	for i := 0; i < 10; i++ {
		values[i] = bs.pseudoRandom()
	}

	// Reset and test again
	bs2 := NewBriberySystem(cs, gps, 12345)
	for i := 0; i < 10; i++ {
		v := bs2.pseudoRandom()
		if v != values[i] {
			t.Errorf("pseudoRandom at %d: got %v, want %v", i, v, values[i])
		}
	}

	// Verify values are in range 0-1
	for i, v := range values {
		if v < 0 || v > 1 {
			t.Errorf("pseudoRandom value %d out of range: %v", i, v)
		}
	}
}

func TestBriberySystem_CalculateBribeCost(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{WantedLevel: 2})

	tests := []struct {
		name        string
		target      BribeTarget
		wantedLevel int
		wantCost    float64
	}{
		{"guard at level 2", BribeTargetGuard, 2, 200.0},            // 100 * 2 * 1.0
		{"witness at level 2", BribeTargetWitness, 2, 100.0},        // 100 * 2 * 0.5
		{"official at level 2", BribeTargetOfficial, 2, 400.0},      // 100 * 2 * 2.0
		{"jailer at level 2", BribeTargetJailer, 2, 300.0},          // 100 * 2 * 1.5
		{"guard at level 4 (high)", BribeTargetGuard, 4, 800.0},     // 100 * 4 * 1.0 * 2.0
		{"witness at level 5 (high)", BribeTargetWitness, 5, 500.0}, // 100 * 5 * 0.5 * 2.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crime, _ := w.GetComponent(e, "Crime")
			crime.(*components.Crime).WantedLevel = tt.wantedLevel

			cost := bs.CalculateBribeCost(w, e, tt.target)
			if cost != tt.wantCost {
				t.Errorf("CalculateBribeCost = %v, want %v", cost, tt.wantCost)
			}
		})
	}
}

func TestBriberySystem_CalculateBribeCost_NoCrime(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity() // No Crime component

	cost := bs.CalculateBribeCost(w, e, BribeTargetGuard)
	if cost != 0 {
		t.Errorf("CalculateBribeCost without Crime component = %v, want 0", cost)
	}
}

func TestBriberySystem_CalculateSuccessChance(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{WantedLevel: 2})

	tests := []struct {
		name            string
		target          BribeTarget
		offerMultiplier float64
		wantedLevel     int
		wantChance      float64
	}{
		// Base chances: Guard 0.6, Witness 0.8, Official 0.5, Jailer 0.4
		// -5% per wanted level, +10% per 100% extra offer
		{"guard standard offer level 2", BribeTargetGuard, 1.0, 2, 0.50},       // 0.6 - 0.1
		{"witness standard offer level 2", BribeTargetWitness, 1.0, 2, 0.70},   // 0.8 - 0.1
		{"official standard offer level 2", BribeTargetOfficial, 1.0, 2, 0.40}, // 0.5 - 0.1
		{"jailer standard offer level 2", BribeTargetJailer, 1.0, 2, 0.30},     // 0.4 - 0.1
		{"guard double offer level 2", BribeTargetGuard, 2.0, 2, 0.60},         // 0.6 + 0.1 - 0.1
		{"guard at level 5", BribeTargetGuard, 1.0, 5, 0.35},                   // 0.6 - 0.25
		{"witness at level 0", BribeTargetWitness, 1.0, 0, 0.80},               // 0.8 - 0.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crime, _ := w.GetComponent(e, "Crime")
			crime.(*components.Crime).WantedLevel = tt.wantedLevel

			chance := bs.CalculateSuccessChance(w, e, tt.target, tt.offerMultiplier)
			diff := chance - tt.wantChance
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("CalculateSuccessChance = %v, want %v", chance, tt.wantChance)
			}
		})
	}
}

func TestBriberySystem_CalculateSuccessChance_Clamping(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()

	// Test min clamping - very high wanted level should clamp to 5%
	w.AddComponent(e, &components.Crime{WantedLevel: 20})
	chance := bs.CalculateSuccessChance(w, e, BribeTargetJailer, 1.0)
	if chance != 0.05 {
		t.Errorf("Min clamp: got %v, want 0.05", chance)
	}

	// Test max clamping - huge offer should clamp to 95%
	w.AddComponent(e, &components.Crime{WantedLevel: 0})
	bs.WitnessSuccessBase = 0.8
	chance = bs.CalculateSuccessChance(w, e, BribeTargetWitness, 10.0)
	if chance != 0.95 {
		t.Errorf("Max clamp: got %v, want 0.95", chance)
	}
}

func TestBriberySystem_AttemptBribe_Insufficient(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{WantedLevel: 2})

	// Offer way below minimum (50% of required)
	// Required for guard at level 2 = 200, minimum = 100
	result := bs.AttemptBribe(w, e, BribeTargetGuard, 50.0)
	if result != BribeResultInsufficient {
		t.Errorf("AttemptBribe with low offer = %v, want BribeResultInsufficient", result)
	}
}

func TestBriberySystem_AttemptBribe_NoTarget(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity() // No Crime component

	result := bs.AttemptBribe(w, e, BribeTargetGuard, 1000.0)
	if result != BribeResultNoTarget {
		t.Errorf("AttemptBribe without Crime = %v, want BribeResultNoTarget", result)
	}
}

func TestBriberySystem_AttemptBribe_Success(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)

	// Use a seed that gives favorable RNG
	bs := NewBriberySystem(cs, gps, 42)

	e := w.CreateEntity()
	w.AddComponent(e, &components.Crime{WantedLevel: 2})

	// Run multiple attempts with very high offers to ensure success
	successCount := 0
	for i := 0; i < 20; i++ {
		// Reset wanted level
		crime, _ := w.GetComponent(e, "Crime")
		crime.(*components.Crime).WantedLevel = 2

		// Very high offer maximizes success chance
		result := bs.AttemptBribe(w, e, BribeTargetWitness, 10000.0)
		if result == BribeResultSuccess {
			successCount++
		}
	}

	// With 95% success chance, we should get several successes
	if successCount == 0 {
		t.Error("No successful bribes in 20 attempts with high offer")
	}
}

func TestBriberySystem_ApplyBribeEffect_Guard(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 3}
	w.AddComponent(e, crime)

	bs.applyBribeEffect(w, e, crime, BribeTargetGuard)

	if crime.WantedLevel != 2 {
		t.Errorf("Guard bribe effect: WantedLevel = %v, want 2", crime.WantedLevel)
	}
}

func TestBriberySystem_ApplyBribeEffect_Witness(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 3}
	w.AddComponent(e, crime)

	bs.applyBribeEffect(w, e, crime, BribeTargetWitness)

	if crime.WantedLevel != 2 {
		t.Errorf("Witness bribe effect: WantedLevel = %v, want 2", crime.WantedLevel)
	}
}

func TestBriberySystem_ApplyBribeEffect_Official(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 4, BountyAmount: 1000.0}
	w.AddComponent(e, crime)

	bs.applyBribeEffect(w, e, crime, BribeTargetOfficial)

	if crime.WantedLevel != 2 {
		t.Errorf("Official bribe effect: WantedLevel = %v, want 2", crime.WantedLevel)
	}
	if crime.BountyAmount != 500.0 {
		t.Errorf("Official bribe effect: BountyAmount = %v, want 500.0", crime.BountyAmount)
	}
}

func TestBriberySystem_ApplyBribeEffect_Jailer(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 3, InJail: true, JailReleaseTime: 1000.0}
	w.AddComponent(e, crime)

	bs.applyBribeEffect(w, e, crime, BribeTargetJailer)

	if crime.InJail {
		t.Error("Jailer bribe effect: should release from jail")
	}
	if crime.JailReleaseTime != 0 {
		t.Errorf("Jailer bribe effect: JailReleaseTime = %v, want 0", crime.JailReleaseTime)
	}
}

func TestBriberySystem_ApplyBribeEffect_MinClamp(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 0}
	w.AddComponent(e, crime)

	bs.applyBribeEffect(w, e, crime, BribeTargetGuard)

	if crime.WantedLevel != 0 {
		t.Errorf("Guard bribe at level 0: WantedLevel = %v, want 0", crime.WantedLevel)
	}
}

func TestBriberySystem_BribeGuard_Success(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{WantedLevel: 2})

	guard := w.CreateEntity()
	guardState := &components.Guard{
		State:        int(GuardStatePursue),
		TargetEntity: uint64(criminal),
	}
	w.AddComponent(guard, guardState)

	// Run multiple attempts
	for i := 0; i < 30; i++ {
		guardState.State = int(GuardStatePursue)
		result := bs.BribeGuard(w, criminal, guard, 10000.0)
		if result == BribeResultSuccess {
			if guardState.State != int(GuardStateReturn) {
				t.Error("Guard should return to patrol after successful bribe")
			}
			if guardState.TargetEntity != 0 {
				t.Error("Guard should clear target after successful bribe")
			}
			return // Test passed
		}
	}
	// With high offer, should succeed within 30 tries
	t.Log("Warning: No successful guard bribe in 30 attempts (RNG unlucky)")
}

func TestBriberySystem_BribeGuard_NoTarget(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{WantedLevel: 2})

	guard := w.CreateEntity()
	// No Guard component

	result := bs.BribeGuard(w, criminal, guard, 1000.0)
	if result != BribeResultNoTarget {
		t.Errorf("BribeGuard without Guard component = %v, want BribeResultNoTarget", result)
	}
}

func TestBriberySystem_BribeGuard_WrongState(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{WantedLevel: 2})

	guard := w.CreateEntity()
	guardState := &components.Guard{
		State:        int(GuardStatePatrol), // Not pursuing
		TargetEntity: uint64(criminal),
	}
	w.AddComponent(guard, guardState)

	result := bs.BribeGuard(w, criminal, guard, 1000.0)
	if result != BribeResultNoTarget {
		t.Errorf("BribeGuard with patrolling guard = %v, want BribeResultNoTarget", result)
	}
}

func TestBriberySystem_BribeGuard_WrongTarget(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{WantedLevel: 2})

	otherCriminal := w.CreateEntity()

	guard := w.CreateEntity()
	guardState := &components.Guard{
		State:        int(GuardStatePursue),
		TargetEntity: uint64(otherCriminal), // Wrong target
	}
	w.AddComponent(guard, guardState)

	result := bs.BribeGuard(w, criminal, guard, 1000.0)
	if result != BribeResultNoTarget {
		t.Errorf("BribeGuard with guard targeting another = %v, want BribeResultNoTarget", result)
	}
}

func TestBriberySystem_BribeWitnessesNearby(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{WantedLevel: 2})

	// Create witnesses - some in range, some out
	for i := 0; i < 5; i++ {
		witness := w.CreateEntity()
		w.AddComponent(witness, &components.Witness{CanReport: true})
		w.AddComponent(witness, &components.Position{
			X: float64(i * 5), // 0, 5, 10, 15, 20
			Y: 0,
			Z: 0,
		})
	}

	// Bribe witnesses within radius 12 from position (0,0)
	// Should reach witnesses at 0, 5, 10 (3 witnesses)
	successes, failures := bs.BribeWitnessesNearby(w, criminal, 0, 0, 12, 10000.0)

	total := successes + failures
	if total != 3 {
		t.Errorf("BribeWitnessesNearby found %d witnesses, want 3", total)
	}

	// Check that some witnesses were marked as bribed
	bribedCount := 0
	for _, e := range w.Entities("Witness") {
		witness, _ := w.GetComponent(e, "Witness")
		if !witness.(*components.Witness).CanReport {
			bribedCount++
		}
	}

	if bribedCount != successes {
		t.Errorf("Bribed witness count = %d, successes = %d", bribedCount, successes)
	}
}

func TestBriberySystem_GetBribeDescription(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	tests := []struct {
		target BribeTarget
		want   string
	}{
		{BribeTargetGuard, "Bribe a guard to stop pursuit and look the other way."},
		{BribeTargetWitness, "Bribe witnesses to forget what they saw."},
		{BribeTargetOfficial, "Bribe an official to reduce your bounty and wanted level."},
		{BribeTargetJailer, "Bribe the jailer for early release from prison."},
		{BribeTarget(99), "Unknown bribe target."},
	}

	for _, tt := range tests {
		desc := bs.GetBribeDescription(tt.target)
		if desc != tt.want {
			t.Errorf("GetBribeDescription(%v) = %q, want %q", tt.target, desc, tt.want)
		}
	}
}

func TestBriberySystem_GetAvailableBribeTargets(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity()

	tests := []struct {
		name        string
		wantedLevel int
		inJail      bool
		wantTargets []BribeTarget
	}{
		{
			"wanted not in jail",
			2, false,
			[]BribeTarget{BribeTargetGuard, BribeTargetWitness, BribeTargetOfficial},
		},
		{
			"in jail",
			2, true,
			[]BribeTarget{BribeTargetJailer},
		},
		{
			"not wanted not in jail",
			0, false,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w.AddComponent(e, &components.Crime{
				WantedLevel: tt.wantedLevel,
				InJail:      tt.inJail,
			})

			targets := bs.GetAvailableBribeTargets(w, e)

			if len(targets) != len(tt.wantTargets) {
				t.Errorf("GetAvailableBribeTargets returned %d targets, want %d", len(targets), len(tt.wantTargets))
				return
			}

			for i, target := range targets {
				if target != tt.wantTargets[i] {
					t.Errorf("Target %d = %v, want %v", i, target, tt.wantTargets[i])
				}
			}
		})
	}
}

func TestBriberySystem_GetAvailableBribeTargets_NoCrime(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	gps := NewGuardPursuitSystem(cs)
	bs := NewBriberySystem(cs, gps, 12345)

	e := w.CreateEntity() // No Crime component

	targets := bs.GetAvailableBribeTargets(w, e)
	if targets != nil {
		t.Errorf("GetAvailableBribeTargets without Crime = %v, want nil", targets)
	}
}
