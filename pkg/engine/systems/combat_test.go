package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestMeleeCombat(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCombatSystem()
	w.RegisterSystem(sys)

	// Create attacker with weapon
	attacker := w.CreateEntity()
	_ = w.AddComponent(attacker, &components.Position{X: 0, Y: 0, Z: 0})
	_ = w.AddComponent(attacker, &components.Health{Current: 100, Max: 100})
	_ = w.AddComponent(attacker, &components.Weapon{
		Name:        "Sword",
		Damage:      25,
		Range:       2.0,
		AttackSpeed: 2.0,
		WeaponType:  "melee",
	})
	_ = w.AddComponent(attacker, &components.CombatState{})

	// Create target in range
	target := w.CreateEntity()
	_ = w.AddComponent(target, &components.Position{X: 1, Y: 0, Z: 0})
	_ = w.AddComponent(target, &components.Health{Current: 100, Max: 100})

	// Initiate attack
	success := sys.InitiateAttack(w, attacker, target)
	if !success {
		t.Fatal("InitiateAttack should succeed when in range")
	}

	// Run update to resolve attack
	w.Update(0.016)

	// Check target took damage
	healthComp, _ := w.GetComponent(target, "Health")
	health := healthComp.(*components.Health)
	if health.Current >= 100 {
		t.Errorf("Target should have taken damage, health = %f", health.Current)
	}
	if health.Current != 75 {
		t.Errorf("Expected health 75 (100 - 25 damage), got %f", health.Current)
	}
}

func TestMeleeCombatOutOfRange(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCombatSystem()
	w.RegisterSystem(sys)

	// Create attacker
	attacker := w.CreateEntity()
	_ = w.AddComponent(attacker, &components.Position{X: 0, Y: 0, Z: 0})
	_ = w.AddComponent(attacker, &components.Weapon{
		Range:      2.0,
		WeaponType: "melee",
	})
	_ = w.AddComponent(attacker, &components.CombatState{})

	// Create target out of range
	target := w.CreateEntity()
	_ = w.AddComponent(target, &components.Position{X: 10, Y: 0, Z: 0})
	_ = w.AddComponent(target, &components.Health{Current: 100, Max: 100})

	// Initiate attack
	success := sys.InitiateAttack(w, attacker, target)
	if success {
		t.Error("InitiateAttack should fail when out of range")
	}
}

func TestDamageCalculation(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCombatSystem()

	// Test without weapon (default damage)
	attacker := w.CreateEntity()
	damage := sys.calculateDamage(w, attacker)
	if damage != sys.DefaultDamage {
		t.Errorf("Expected default damage %f, got %f", sys.DefaultDamage, damage)
	}

	// Test with weapon
	_ = w.AddComponent(attacker, &components.Weapon{
		Damage: 50,
	})
	damage = sys.calculateDamage(w, attacker)
	if damage != 50 {
		t.Errorf("Expected weapon damage 50, got %f", damage)
	}
}

func TestDamageCalculationWithSkills(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCombatSystem()

	attacker := w.CreateEntity()
	_ = w.AddComponent(attacker, &components.Weapon{
		Damage: 100,
	})
	_ = w.AddComponent(attacker, &components.Skills{
		Levels: map[string]int{
			"melee": 10, // +20% damage
		},
		Experience: make(map[string]float64),
	})

	damage := sys.calculateDamage(w, attacker)
	expected := 100.0 * 1.20 // 120 damage
	if damage != expected {
		t.Errorf("Expected damage %f with skill modifier, got %f", expected, damage)
	}
}

func TestAttackCooldown(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCombatSystem()
	w.RegisterSystem(sys)

	// Create attacker
	attacker := w.CreateEntity()
	_ = w.AddComponent(attacker, &components.Position{X: 0, Y: 0, Z: 0})
	_ = w.AddComponent(attacker, &components.Weapon{
		Range:       2.0,
		AttackSpeed: 2.0, // 2 attacks per second = 0.5s cooldown
		WeaponType:  "melee",
	})
	_ = w.AddComponent(attacker, &components.CombatState{})

	// Create target
	target := w.CreateEntity()
	_ = w.AddComponent(target, &components.Position{X: 1, Y: 0, Z: 0})
	_ = w.AddComponent(target, &components.Health{Current: 100, Max: 100})

	// First attack should succeed
	success := sys.InitiateAttack(w, attacker, target)
	if !success {
		t.Fatal("First attack should succeed")
	}

	// Resolve attack
	w.Update(0.016)

	// Second attack should fail (on cooldown)
	success = sys.InitiateAttack(w, attacker, target)
	if success {
		t.Error("Second attack should fail due to cooldown")
	}

	// Wait for cooldown (0.5s)
	w.Update(0.5)

	// Third attack should succeed
	success = sys.InitiateAttack(w, attacker, target)
	if !success {
		t.Error("Third attack should succeed after cooldown")
	}
}

func TestFindNearestTarget(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCombatSystem()

	// Create attacker
	attacker := w.CreateEntity()
	_ = w.AddComponent(attacker, &components.Position{X: 0, Y: 0, Z: 0})
	_ = w.AddComponent(attacker, &components.Weapon{
		Range:      5.0,
		WeaponType: "melee",
	})

	// Create targets at different distances
	close := w.CreateEntity()
	_ = w.AddComponent(close, &components.Position{X: 2, Y: 0, Z: 0})
	_ = w.AddComponent(close, &components.Health{Current: 100, Max: 100})

	far := w.CreateEntity()
	_ = w.AddComponent(far, &components.Position{X: 4, Y: 0, Z: 0})
	_ = w.AddComponent(far, &components.Health{Current: 100, Max: 100})

	tooFar := w.CreateEntity()
	_ = w.AddComponent(tooFar, &components.Position{X: 10, Y: 0, Z: 0})
	_ = w.AddComponent(tooFar, &components.Health{Current: 100, Max: 100})

	nearest := sys.FindNearestTarget(w, attacker)
	if nearest != close {
		t.Errorf("Expected nearest target to be 'close' entity, got %d", nearest)
	}
}

func TestWeaponComponent(t *testing.T) {
	weapon := &components.Weapon{
		Name:        "Steel Sword",
		Damage:      30,
		Range:       2.5,
		AttackSpeed: 1.5,
		WeaponType:  "melee",
	}

	if weapon.Type() != "Weapon" {
		t.Errorf("Weapon.Type() = %s, want 'Weapon'", weapon.Type())
	}
}

func TestCombatStateComponent(t *testing.T) {
	combat := &components.CombatState{
		LastAttackTime: 1.0,
		Cooldown:       0.5,
		IsAttacking:    true,
		TargetEntity:   42,
		InCombat:       true,
	}

	if combat.Type() != "CombatState" {
		t.Errorf("CombatState.Type() = %s, want 'CombatState'", combat.Type())
	}
}
