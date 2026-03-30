package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewNPCNeedsSystem(t *testing.T) {
	ns := NewNPCNeedsSystem()
	if ns == nil {
		t.Fatal("NewNPCNeedsSystem returned nil")
	}
	if ns.DefaultHungerRate <= 0 {
		t.Error("DefaultHungerRate should be positive")
	}
	if ns.DefaultEnergyRate <= 0 {
		t.Error("DefaultEnergyRate should be positive")
	}
	if ns.DefaultSocialDecayRate <= 0 {
		t.Error("DefaultSocialDecayRate should be positive")
	}
}

func TestNPCNeedsSystem_Update(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()
	ns.GameTime = 1.0
	ns.LastUpdateTime = 0.0

	e := w.CreateEntity()
	w.AddComponent(e, &components.NPCNeeds{
		Hunger: 0.5,
		Energy: 0.8,
		Social: 0.6,
		Safety: 1.0,
	})
	w.AddComponent(e, &components.Position{X: 0, Y: 0, Z: 0})

	ns.Update(w, 1.0)

	comp, _ := w.GetComponent(e, "NPCNeeds")
	needs := comp.(*components.NPCNeeds)

	// Hunger should increase over time
	if needs.Hunger <= 0.5 {
		t.Error("Hunger should have increased")
	}
	// Energy should decrease when awake
	if needs.Energy >= 0.8 {
		t.Error("Energy should have decreased")
	}
}

func TestNPCNeedsSystem_Update_NoTimeElapsed(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()
	ns.GameTime = 1.0
	ns.LastUpdateTime = 1.0 // Same as GameTime

	e := w.CreateEntity()
	w.AddComponent(e, &components.NPCNeeds{
		Hunger: 0.5,
		Energy: 0.8,
		Social: 0.6,
		Safety: 1.0,
	})

	ns.Update(w, 1.0)

	comp, _ := w.GetComponent(e, "NPCNeeds")
	needs := comp.(*components.NPCNeeds)

	// Nothing should change
	if needs.Hunger != 0.5 {
		t.Errorf("Hunger = %v, want 0.5 (no change)", needs.Hunger)
	}
}

func TestNPCNeedsSystem_IsAwake(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()

	// Entity without schedule (default awake)
	e1 := w.CreateEntity()
	if !ns.isAwake(w, e1) {
		t.Error("Entity without schedule should be awake")
	}

	// Entity with sleep activity
	e2 := w.CreateEntity()
	w.AddComponent(e2, &components.Schedule{
		CurrentActivity: "sleep",
	})
	if ns.isAwake(w, e2) {
		t.Error("Entity with sleep activity should not be awake")
	}

	// Entity with rest activity
	e3 := w.CreateEntity()
	w.AddComponent(e3, &components.Schedule{
		CurrentActivity: "rest",
	})
	if ns.isAwake(w, e3) {
		t.Error("Entity with rest activity should not be awake")
	}

	// Entity with work activity
	e4 := w.CreateEntity()
	w.AddComponent(e4, &components.Schedule{
		CurrentActivity: "work",
	})
	if !ns.isAwake(w, e4) {
		t.Error("Entity with work activity should be awake")
	}
}

func TestNPCNeedsSystem_HasNearbyNPCs(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()

	// Create NPC1
	e1 := w.CreateEntity()
	w.AddComponent(e1, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(e1, &components.NPCNeeds{})

	// No nearby NPCs yet
	if ns.hasNearbyNPCs(w, e1) {
		t.Error("Should have no nearby NPCs")
	}

	// Create NPC2 nearby
	e2 := w.CreateEntity()
	w.AddComponent(e2, &components.Position{X: 5, Y: 0, Z: 5})
	w.AddComponent(e2, &components.NPCNeeds{})

	if !ns.hasNearbyNPCs(w, e1) {
		t.Error("Should have nearby NPC")
	}
}

func TestNPCNeedsSystem_HasNearbyNPCs_FarAway(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()

	e1 := w.CreateEntity()
	w.AddComponent(e1, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(e1, &components.NPCNeeds{})

	// Create NPC far away
	e2 := w.CreateEntity()
	w.AddComponent(e2, &components.Position{X: 1000, Y: 0, Z: 1000})
	w.AddComponent(e2, &components.NPCNeeds{})

	if ns.hasNearbyNPCs(w, e1) {
		t.Error("Should not have nearby NPC (too far)")
	}
}

func TestNPCNeedsSystem_HasNearbyNPCs_NoPosition(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()

	e1 := w.CreateEntity()
	w.AddComponent(e1, &components.NPCNeeds{})
	// No Position component

	if ns.hasNearbyNPCs(w, e1) {
		t.Error("Entity without position should have no nearby NPCs")
	}
}

func TestNPCNeedsSystem_CalculateSafety(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()

	// NPC with position
	e := w.CreateEntity()
	w.AddComponent(e, &components.Position{X: 0, Y: 0, Z: 0})

	// No threats - should be safe
	safety := ns.calculateSafety(w, e)
	if safety != 1.0 {
		t.Errorf("Safety = %v, want 1.0 (no threats)", safety)
	}

	// Add a nearby threat
	threat := w.CreateEntity()
	w.AddComponent(threat, &components.Position{X: 5, Y: 0, Z: 5})
	w.AddComponent(threat, &components.CombatState{InCombat: true})

	safety = ns.calculateSafety(w, e)
	if safety >= 1.0 {
		t.Error("Safety should be reduced by nearby threat")
	}
}

func TestNPCNeedsSystem_CalculateSafety_NoPosition(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()

	e := w.CreateEntity()
	// No Position component

	safety := ns.calculateSafety(w, e)
	if safety != 1.0 {
		t.Errorf("Safety = %v, want 1.0 (default for no position)", safety)
	}
}

func TestNPCNeedsSystem_CalculateSafety_MultipleThreat(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()

	e := w.CreateEntity()
	w.AddComponent(e, &components.Position{X: 0, Y: 0, Z: 0})

	// Add multiple threats
	for i := 0; i < 5; i++ {
		threat := w.CreateEntity()
		w.AddComponent(threat, &components.Position{X: float64(i) + 1, Y: 0, Z: float64(i) + 1})
		w.AddComponent(threat, &components.CombatState{InCombat: true})
	}

	safety := ns.calculateSafety(w, e)
	// With 5 threats at -0.2 each, should be clamped to 0
	if safety != 0.0 {
		t.Errorf("Safety = %v, want 0.0 (max threats)", safety)
	}
}

func TestNPCNeedsSystem_GetNeedPriority(t *testing.T) {
	ns := NewNPCNeedsSystem()

	tests := []struct {
		name     string
		needs    *components.NPCNeeds
		expected string
	}{
		{
			name:     "low safety",
			needs:    &components.NPCNeeds{Safety: 0.1, Energy: 0.5, Hunger: 0.5, Social: 0.5},
			expected: "safety",
		},
		{
			name:     "low energy",
			needs:    &components.NPCNeeds{Safety: 0.5, Energy: 0.1, Hunger: 0.5, Social: 0.5},
			expected: "sleep",
		},
		{
			name:     "high hunger",
			needs:    &components.NPCNeeds{Safety: 0.5, Energy: 0.5, Hunger: 0.9, Social: 0.5},
			expected: "eat",
		},
		{
			name:     "low social",
			needs:    &components.NPCNeeds{Safety: 0.5, Energy: 0.5, Hunger: 0.5, Social: 0.1},
			expected: "socialize",
		},
		{
			name:     "all needs satisfied",
			needs:    &components.NPCNeeds{Safety: 0.8, Energy: 0.8, Hunger: 0.3, Social: 0.8},
			expected: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ns.GetNeedPriority(tt.needs)
			if result != tt.expected {
				t.Errorf("GetNeedPriority = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestNPCNeedsSystem_GetNeedModifier(t *testing.T) {
	ns := NewNPCNeedsSystem()

	needs := &components.NPCNeeds{
		Safety: 0.5,
		Energy: 0.5,
		Hunger: 0.5,
		Social: 0.5,
	}

	mod := ns.GetNeedModifier(needs)

	// All modifiers should be non-zero
	if mod.SpeedModifier <= 0 {
		t.Error("SpeedModifier should be positive")
	}
	if mod.AlertnessModifier <= 0 {
		t.Error("AlertnessModifier should be positive")
	}
	if mod.MoodModifier <= 0 {
		t.Error("MoodModifier should be positive")
	}
}

func TestCalculateSpeedModifier(t *testing.T) {
	tests := []struct {
		name    string
		needs   *components.NPCNeeds
		wantMin float64
		wantMax float64
	}{
		{
			name:    "high energy normal hunger",
			needs:   &components.NPCNeeds{Energy: 1.0, Hunger: 0.3},
			wantMin: 0.9,
			wantMax: 1.1,
		},
		{
			name:    "low energy",
			needs:   &components.NPCNeeds{Energy: 0.0, Hunger: 0.3},
			wantMin: 0.4,
			wantMax: 0.6,
		},
		{
			name:    "very hungry",
			needs:   &components.NPCNeeds{Energy: 1.0, Hunger: 0.9},
			wantMin: 0.8,
			wantMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSpeedModifier(tt.needs)
			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("calculateSpeedModifier = %v, want [%v, %v]", result, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateAlertnessModifier(t *testing.T) {
	tests := []struct {
		name     string
		needs    *components.NPCNeeds
		expected float64
	}{
		{
			name:     "low energy",
			needs:    &components.NPCNeeds{Energy: 0.1, Safety: 1.0},
			expected: 0.6,
		},
		{
			name:     "low safety",
			needs:    &components.NPCNeeds{Energy: 1.0, Safety: 0.3},
			expected: 1.3,
		},
		{
			name:     "normal",
			needs:    &components.NPCNeeds{Energy: 0.5, Safety: 0.8},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAlertnessModifier(tt.needs)
			if result != tt.expected {
				t.Errorf("calculateAlertnessModifier = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateMoodModifier(t *testing.T) {
	// Test with all good needs
	goodNeeds := &components.NPCNeeds{
		Hunger: 0.1, // Low is good
		Energy: 0.9,
		Social: 0.9,
		Safety: 0.9,
	}
	goodMood := calculateMoodModifier(goodNeeds)

	// Test with all bad needs
	badNeeds := &components.NPCNeeds{
		Hunger: 0.9, // High is bad
		Energy: 0.1,
		Social: 0.1,
		Safety: 0.1,
	}
	badMood := calculateMoodModifier(badNeeds)

	if goodMood <= badMood {
		t.Errorf("Good needs mood (%v) should be > bad needs mood (%v)", goodMood, badMood)
	}
}

func TestClampNeed(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},
		{0.0, 0.0},
		{1.0, 1.0},
		{-0.5, 0.0},
		{1.5, 1.0},
	}

	for _, tt := range tests {
		result := clampNeed(tt.input)
		if result != tt.expected {
			t.Errorf("clampNeed(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestNPCNeedsSystem_UpdateSleeping(t *testing.T) {
	w := ecs.NewWorld()
	ns := NewNPCNeedsSystem()
	ns.GameTime = 1.0
	ns.LastUpdateTime = 0.0

	e := w.CreateEntity()
	w.AddComponent(e, &components.NPCNeeds{
		Hunger: 0.5,
		Energy: 0.5,
		Social: 0.6,
		Safety: 1.0,
	})
	w.AddComponent(e, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(e, &components.Schedule{
		CurrentActivity: "sleep",
	})

	initialEnergy := 0.5
	ns.Update(w, 1.0)

	comp, _ := w.GetComponent(e, "NPCNeeds")
	needs := comp.(*components.NPCNeeds)

	// Energy should increase when sleeping
	if needs.Energy <= initialEnergy {
		t.Error("Energy should increase when sleeping")
	}
}
