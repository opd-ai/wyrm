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
