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
	if cm.magicSystem == nil {
		t.Error("magicSystem should not be nil")
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

func TestCombatManager_SpellSelection(t *testing.T) {
	cm := NewCombatManager(1, nil)

	// Default spell index should be 0
	if cm.GetSelectedSpellIndex() != 0 {
		t.Errorf("expected default spell index 0, got %d", cm.GetSelectedSpellIndex())
	}

	// Set spell index
	cm.SetSelectedSpellIndex(3)
	if cm.GetSelectedSpellIndex() != 3 {
		t.Errorf("expected spell index 3, got %d", cm.GetSelectedSpellIndex())
	}

	// Invalid index (too high) should not change
	cm.SetSelectedSpellIndex(10)
	if cm.GetSelectedSpellIndex() != 3 {
		t.Errorf("expected spell index 3 (unchanged), got %d", cm.GetSelectedSpellIndex())
	}

	// Invalid index (negative) should not change
	cm.SetSelectedSpellIndex(-1)
	if cm.GetSelectedSpellIndex() != 3 {
		t.Errorf("expected spell index 3 (unchanged), got %d", cm.GetSelectedSpellIndex())
	}
}

func TestCombatManager_GetSelectedSpellID(t *testing.T) {
	world := ecs.NewWorld()
	player := world.CreateEntity()
	cm := NewCombatManager(player, nil)

	// No spellbook = empty string
	spellID := cm.getSelectedSpellID(world)
	if spellID != "" {
		t.Errorf("expected empty spell ID with no spellbook, got %s", spellID)
	}

	// Add spellbook with spells
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{
			"fireball": {ID: "fireball", Name: "Fireball", ManaCost: 10},
			"heal":     {ID: "heal", Name: "Heal", ManaCost: 15},
		},
	}
	world.AddComponent(player, spellbook)

	// Should return some spell (map order is random)
	spellID = cm.getSelectedSpellID(world)
	if spellID != "fireball" && spellID != "heal" {
		t.Errorf("expected fireball or heal, got %s", spellID)
	}

	// With active spell set, should return that
	spellbook.ActiveSpellID = "heal"
	spellID = cm.getSelectedSpellID(world)
	if spellID != "heal" {
		t.Errorf("expected active spell 'heal', got %s", spellID)
	}
}

func TestCombatManager_CanCastSpell(t *testing.T) {
	cm := NewCombatManager(1, nil)

	// Initially can cast
	if !cm.canCastSpell() {
		t.Error("should be able to cast spell initially")
	}

	// Dead = cannot cast
	cm.isDead = true
	if cm.canCastSpell() {
		t.Error("should not be able to cast while dead")
	}
	cm.isDead = false

	// Blocking = cannot cast
	cm.isBlocking = true
	if cm.canCastSpell() {
		t.Error("should not be able to cast while blocking")
	}
}

func TestCombatManager_Stealth(t *testing.T) {
	world := ecs.NewWorld()
	player := world.CreateEntity()
	cm := NewCombatManager(player, nil)

	// Initially not sneaking
	if cm.IsSneaking() {
		t.Error("should not be sneaking initially")
	}

	// Add stealth component to player
	stealth := &components.Stealth{
		Sneaking:        false,
		Visibility:      1.0,
		DetectionRadius: 10.0,
		BaseVisibility:  1.0,
		SneakVisibility: 0.3,
	}
	world.AddComponent(player, stealth)

	// Toggle sneak
	cm.toggleSneak(world)
	if !cm.IsSneaking() {
		t.Error("should be sneaking after toggle")
	}

	// Toggle sneak again
	cm.toggleSneak(world)
	if cm.IsSneaking() {
		t.Error("should not be sneaking after second toggle")
	}

	// Break stealth
	cm.toggleSneak(world) // Start sneaking
	cm.breakStealth(world)
	if cm.IsSneaking() {
		t.Error("should not be sneaking after breakStealth")
	}
}

func TestCombatManager_GetWeaponDamage(t *testing.T) {
	world := ecs.NewWorld()
	player := world.CreateEntity()
	cm := NewCombatManager(player, nil)

	// No weapon = default 10
	damage := cm.getWeaponDamage(world)
	if damage != 10.0 {
		t.Errorf("expected default damage 10.0, got %f", damage)
	}

	// Add weapon with damage
	weapon := &components.Weapon{
		Name:       "Dagger",
		Damage:     15,
		WeaponType: "melee",
	}
	world.AddComponent(player, weapon)

	damage = cm.getWeaponDamage(world)
	if damage != 15.0 {
		t.Errorf("expected weapon damage 15.0, got %f", damage)
	}
}

func TestNewCombatManager_StealthSystem(t *testing.T) {
	cm := NewCombatManager(1, nil)
	if cm.stealthSystem == nil {
		t.Error("stealthSystem should not be nil")
	}
}

func TestCombatManager_CanDodge(t *testing.T) {
	cm := NewCombatManager(1, nil)

	// Initially can dodge
	if !cm.canDodge() {
		t.Error("should be able to dodge initially")
	}

	// Dead = cannot dodge
	cm.isDead = true
	if cm.canDodge() {
		t.Error("should not be able to dodge while dead")
	}
	cm.isDead = false

	// Already dodging = cannot dodge
	cm.isDodging = true
	if cm.canDodge() {
		t.Error("should not be able to dodge while already dodging")
	}
	cm.isDodging = false

	// Blocking = cannot dodge
	cm.isBlocking = true
	if cm.canDodge() {
		t.Error("should not be able to dodge while blocking")
	}
}

func TestCombatManager_IsDodging(t *testing.T) {
	cm := NewCombatManager(1, nil)

	// Initially not dodging
	if cm.IsDodging() {
		t.Error("should not be dodging initially")
	}

	cm.isDodging = true
	if !cm.IsDodging() {
		t.Error("should report dodging when isDodging is true")
	}
}

func TestCombatManager_GetBlockReduction(t *testing.T) {
	cm := NewCombatManager(1, nil)

	// Not blocking = no reduction
	if cm.GetBlockReduction() != 0 {
		t.Errorf("expected 0 reduction when not blocking, got %f", cm.GetBlockReduction())
	}

	// Blocking = should return blockReduction value
	cm.isBlocking = true
	if cm.GetBlockReduction() != cm.blockReduction {
		t.Errorf("expected block reduction %f, got %f", cm.blockReduction, cm.GetBlockReduction())
	}
}

func TestCombatManager_CalculateIncomingDamage(t *testing.T) {
	cm := NewCombatManager(1, nil)

	baseDamage := 100.0

	// No dodge/block = full damage
	result := cm.CalculateIncomingDamage(baseDamage)
	if result != baseDamage {
		t.Errorf("expected full damage %f, got %f", baseDamage, result)
	}

	// Dodging = zero damage
	cm.isDodging = true
	result = cm.CalculateIncomingDamage(baseDamage)
	if result != 0 {
		t.Errorf("expected zero damage while dodging, got %f", result)
	}
	cm.isDodging = false

	// Blocking = reduced damage
	cm.isBlocking = true
	expected := baseDamage * (1.0 - cm.blockReduction)
	result = cm.CalculateIncomingDamage(baseDamage)
	if result != expected {
		t.Errorf("expected blocked damage %f, got %f", expected, result)
	}
}
