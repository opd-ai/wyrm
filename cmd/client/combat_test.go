//go:build noebiten

// Test file for combat manager logic (noebiten build tag for CI).
package main

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewCombatManager(t *testing.T) {
	cm := NewCombatManager(1, nil)
	if cm == nil {
		t.Fatal("NewCombatManager returned nil")
	}
	if cm.combatSystem == nil {
		t.Error("combatSystem should not be nil")
	}
	if cm.projectileSystem == nil {
		t.Error("projectileSystem should not be nil")
	}
}

func TestCombatManager_GetEquippedWeaponType(t *testing.T) {
	world := ecs.NewWorld()
	player := world.CreateEntity()
	cm := NewCombatManager(player, nil)

	// No weapon = melee default
	weaponType := cm.getEquippedWeaponType(world)
	if weaponType != "melee" {
		t.Errorf("expected melee, got %s", weaponType)
	}

	// Add ranged weapon
	rangedWeapon := &components.Weapon{
		Name:       "Bow",
		Damage:     15,
		Range:      25,
		WeaponType: "ranged",
	}
	world.AddComponent(player, rangedWeapon)

	weaponType = cm.getEquippedWeaponType(world)
	if weaponType != "ranged" {
		t.Errorf("expected ranged, got %s", weaponType)
	}
}

func TestCombatManager_GetRangedWeaponStats(t *testing.T) {
	world := ecs.NewWorld()
	player := world.CreateEntity()
	cm := NewCombatManager(player, nil)

	// No weapon = defaults
	damage, speed, weaponRange := cm.getRangedWeaponStats(world)
	if damage != 10.0 {
		t.Errorf("expected default damage 10.0, got %f", damage)
	}
	if speed != 15.0 {
		t.Errorf("expected default speed 15.0, got %f", speed)
	}
	if weaponRange != 20.0 {
		t.Errorf("expected default range 20.0, got %f", weaponRange)
	}

	// Add weapon with custom stats
	weapon := &components.Weapon{
		Name:       "Crossbow",
		Damage:     20,
		Range:      30,
		WeaponType: "ranged",
	}
	world.AddComponent(player, weapon)

	damage, speed, weaponRange = cm.getRangedWeaponStats(world)
	if damage != 20.0 {
		t.Errorf("expected damage 20.0, got %f", damage)
	}
	if weaponRange != 30.0 {
		t.Errorf("expected range 30.0, got %f", weaponRange)
	}
}

func TestCombatManager_CanAttack(t *testing.T) {
	cm := NewCombatManager(1, nil)

	// Initially can attack
	if !cm.canAttack() {
		t.Error("should be able to attack initially")
	}

	// Dead = cannot attack
	cm.isDead = true
	if cm.canAttack() {
		t.Error("should not be able to attack while dead")
	}
	cm.isDead = false

	// Blocking = cannot attack
	cm.isBlocking = true
	if cm.canAttack() {
		t.Error("should not be able to attack while blocking")
	}
}

func TestCombatManager_UpdateAimDirection(t *testing.T) {
	world := ecs.NewWorld()
	player := world.CreateEntity()
	cm := NewCombatManager(player, nil)

	// Add position with angle
	pos := &components.Position{X: 5, Y: 5, Z: 0, Angle: 0}
	world.AddComponent(player, pos)

	cm.updateAimDirection(world)

	// Angle 0 = aiming right (cos(0)=1, sin(0)=0)
	if cm.aimDirX < 0.99 || cm.aimDirX > 1.01 {
		t.Errorf("expected aimDirX ~1.0, got %f", cm.aimDirX)
	}
	if cm.aimDirY < -0.01 || cm.aimDirY > 0.01 {
		t.Errorf("expected aimDirY ~0.0, got %f", cm.aimDirY)
	}
}
