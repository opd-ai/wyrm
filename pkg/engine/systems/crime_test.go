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
