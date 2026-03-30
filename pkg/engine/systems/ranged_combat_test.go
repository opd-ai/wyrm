package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewProjectileSystem(t *testing.T) {
	ps := NewProjectileSystem()
	if ps == nil {
		t.Fatal("NewProjectileSystem returned nil")
	}
}

func TestProjectileSystem_SpawnProjectile(t *testing.T) {
	w := ecs.NewWorld()
	ps := NewProjectileSystem()

	// Create owner entity with position
	owner := w.CreateEntity()
	w.AddComponent(owner, &components.Position{X: 0, Y: 0, Z: 0})

	// Spawn projectile toward target
	proj := ps.SpawnProjectile(w, owner, 10, 0, 0, 25.0, 20.0, "arrow")
	if proj == 0 {
		t.Fatal("SpawnProjectile returned 0")
	}

	// Verify projectile has Position
	posComp, ok := w.GetComponent(proj, "Position")
	if !ok {
		t.Fatal("Projectile missing Position component")
	}
	pos := posComp.(*components.Position)
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("Projectile should start at owner position, got (%f, %f)", pos.X, pos.Y)
	}

	// Verify projectile has Projectile component
	projComp, ok := w.GetComponent(proj, "Projectile")
	if !ok {
		t.Fatal("Projectile missing Projectile component")
	}
	projectile := projComp.(*components.Projectile)
	if projectile.Damage != 25.0 {
		t.Errorf("Damage should be 25.0, got %f", projectile.Damage)
	}
	if projectile.VelocityX <= 0 {
		t.Error("VelocityX should be positive for target to the right")
	}
	if projectile.ProjectileType != "arrow" {
		t.Errorf("ProjectileType should be 'arrow', got '%s'", projectile.ProjectileType)
	}
}

func TestProjectileSystem_Update_Movement(t *testing.T) {
	w := ecs.NewWorld()
	ps := NewProjectileSystem()

	// Create owner
	owner := w.CreateEntity()
	w.AddComponent(owner, &components.Position{X: 0, Y: 0, Z: 0})

	// Spawn projectile
	proj := ps.SpawnProjectile(w, owner, 100, 0, 0, 10.0, 20.0, "bullet")

	// Get initial position
	posComp, _ := w.GetComponent(proj, "Position")
	pos := posComp.(*components.Position)
	initialX := pos.X

	// Update system
	ps.Update(w, 1.0) // 1 second

	// Check position moved
	if pos.X <= initialX {
		t.Error("Projectile should have moved in positive X direction")
	}
	// At 20 units/sec, after 1 second should be at ~20
	if pos.X < 18 || pos.X > 22 {
		t.Errorf("Expected X ~20 after 1 second, got %f", pos.X)
	}
}

func TestProjectileSystem_Update_Collision(t *testing.T) {
	w := ecs.NewWorld()
	ps := NewProjectileSystem()

	// Create owner
	owner := w.CreateEntity()
	w.AddComponent(owner, &components.Position{X: 0, Y: 0, Z: 0})

	// Create target with health at position (1, 0, 0)
	target := w.CreateEntity()
	w.AddComponent(target, &components.Position{X: 1, Y: 0, Z: 0})
	w.AddComponent(target, &components.Health{Current: 100, Max: 100})

	// Spawn projectile toward target with large hit radius
	proj := ps.SpawnProjectile(w, owner, 1, 0, 0, 30.0, 10.0, "arrow")

	// Manually adjust hit radius to ensure collision
	projComp, _ := w.GetComponent(proj, "Projectile")
	projectile := projComp.(*components.Projectile)
	projectile.HitRadius = 1.0 // Increase hit radius for test

	// Manually move projectile to target position
	posComp, _ := w.GetComponent(proj, "Position")
	pos := posComp.(*components.Position)
	pos.X = 1.0

	// Update to check collision
	ps.Update(w, 0.1)

	// Check target took damage
	healthComp, _ := w.GetComponent(target, "Health")
	health := healthComp.(*components.Health)
	if health.Current >= 100 {
		t.Errorf("Target should have taken damage, health is %f", health.Current)
	}
	expectedHealth := 100.0 - 30.0
	if health.Current != expectedHealth {
		t.Errorf("Expected health %f, got %f", expectedHealth, health.Current)
	}
}

func TestProjectileSystem_Update_Lifetime(t *testing.T) {
	w := ecs.NewWorld()
	ps := NewProjectileSystem()

	// Create owner
	owner := w.CreateEntity()
	w.AddComponent(owner, &components.Position{X: 0, Y: 0, Z: 0})

	// Spawn projectile
	proj := ps.SpawnProjectile(w, owner, 100, 0, 0, 10.0, 20.0, "arrow")

	// Verify projectile exists
	_, exists := w.GetComponent(proj, "Projectile")
	if !exists {
		t.Fatal("Projectile should exist initially")
	}

	// Update past lifetime
	ps.Update(w, DefaultProjectileLifetime+1)

	// Projectile should be destroyed
	_, exists = w.GetComponent(proj, "Projectile")
	if exists {
		t.Error("Projectile should be destroyed after lifetime expires")
	}
}

func TestProjectileSystem_Pierce(t *testing.T) {
	w := ecs.NewWorld()
	ps := NewProjectileSystem()

	// Create owner
	owner := w.CreateEntity()
	w.AddComponent(owner, &components.Position{X: 0, Y: 0, Z: 0})

	// Create two targets
	target1 := w.CreateEntity()
	w.AddComponent(target1, &components.Position{X: 1, Y: 0, Z: 0})
	w.AddComponent(target1, &components.Health{Current: 100, Max: 100})

	target2 := w.CreateEntity()
	w.AddComponent(target2, &components.Position{X: 2, Y: 0, Z: 0})
	w.AddComponent(target2, &components.Health{Current: 100, Max: 100})

	// Spawn projectile with pierce = 2
	proj := ps.SpawnProjectile(w, owner, 5, 0, 0, 20.0, 10.0, "arrow")
	projComp, _ := w.GetComponent(proj, "Projectile")
	projectile := projComp.(*components.Projectile)
	projectile.PierceCount = 2

	// Move to hit first target
	posComp, _ := w.GetComponent(proj, "Position")
	pos := posComp.(*components.Position)
	pos.X = 1.0
	ps.Update(w, 0.01)

	// First target should be damaged
	health1, _ := w.GetComponent(target1, "Health")
	if health1.(*components.Health).Current == 100 {
		t.Error("First target should have taken damage")
	}

	// Projectile should still exist (pierce = 2)
	_, exists := w.GetComponent(proj, "Projectile")
	if !exists {
		t.Error("Projectile should still exist after first pierce")
	}

	// Move to hit second target
	pos.X = 2.0
	ps.Update(w, 0.01)

	// Second target should be damaged
	health2, _ := w.GetComponent(target2, "Health")
	if health2.(*components.Health).Current == 100 {
		t.Error("Second target should have taken damage")
	}

	// Projectile should be marked for cleanup (pierce exhausted)
	if projectile.Lifetime > 0 {
		t.Error("Projectile lifetime should be 0 after exhausting pierce count")
	}
}

func TestCombatSystem_InitiateRangedAttack(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCombatSystem()
	ps := NewProjectileSystem()

	// Create attacker with ranged weapon
	attacker := w.CreateEntity()
	w.AddComponent(attacker, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(attacker, &components.CombatState{})
	w.AddComponent(attacker, &components.Weapon{
		Name:        "Longbow",
		Damage:      15,
		Range:       50,
		AttackSpeed: 1.0,
		WeaponType:  "bow",
	})

	// Initiate ranged attack
	success := cs.InitiateRangedAttack(w, attacker, 30, 0, 0, ps)
	if !success {
		t.Error("InitiateRangedAttack should succeed")
	}

	// Check cooldown was set
	combatComp, _ := w.GetComponent(attacker, "CombatState")
	combat := combatComp.(*components.CombatState)
	if combat.Cooldown <= 0 {
		t.Error("Cooldown should be set after attack")
	}
	if !combat.InCombat {
		t.Error("Should be in combat after attack")
	}
}

func TestCombatSystem_InitiateRangedAttack_OutOfRange(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCombatSystem()
	ps := NewProjectileSystem()

	// Create attacker with short-range bow
	attacker := w.CreateEntity()
	w.AddComponent(attacker, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(attacker, &components.CombatState{})
	w.AddComponent(attacker, &components.Weapon{
		Name:        "Shortbow",
		Damage:      10,
		Range:       20,
		AttackSpeed: 1.5,
		WeaponType:  "bow",
	})

	// Try to attack target beyond range
	success := cs.InitiateRangedAttack(w, attacker, 100, 0, 0, ps)
	if success {
		t.Error("InitiateRangedAttack should fail when target is out of range")
	}
}

func TestCombatSystem_InitiateRangedAttack_MeleeWeapon(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCombatSystem()
	ps := NewProjectileSystem()

	// Create attacker with melee weapon
	attacker := w.CreateEntity()
	w.AddComponent(attacker, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(attacker, &components.CombatState{})
	w.AddComponent(attacker, &components.Weapon{
		Name:        "Sword",
		Damage:      20,
		Range:       2,
		AttackSpeed: 1.0,
		WeaponType:  "melee",
	})

	// Try to use ranged attack with melee weapon
	success := cs.InitiateRangedAttack(w, attacker, 10, 0, 0, ps)
	if success {
		t.Error("InitiateRangedAttack should fail with melee weapon")
	}
}

func TestCombatSystem_IsRangedWeapon(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCombatSystem()

	tests := []struct {
		weaponType string
		expected   bool
	}{
		{"melee", false},
		{"bow", true},
		{"ranged", true},
		{"gun", true},
		{"crossbow", true},
		{"rifle", true},
		{"sword", false},
	}

	for _, tt := range tests {
		entity := w.CreateEntity()
		w.AddComponent(entity, &components.Weapon{
			Name:       "Test",
			WeaponType: tt.weaponType,
			Range:      2,
		})

		result := cs.IsRangedWeapon(w, entity)
		if result != tt.expected {
			t.Errorf("IsRangedWeapon for '%s' = %v, want %v", tt.weaponType, result, tt.expected)
		}
	}
}

func TestCombatSystem_getRangedSkillModifier(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCombatSystem()

	// Entity without skills
	entity1 := w.CreateEntity()
	mod1 := cs.getRangedSkillModifier(w, entity1)
	if mod1 != 1.0 {
		t.Errorf("Without skills, modifier should be 1.0, got %f", mod1)
	}

	// Entity with archery skill
	entity2 := w.CreateEntity()
	w.AddComponent(entity2, &components.Skills{
		Levels: map[string]int{"archery": 50},
	})
	mod2 := cs.getRangedSkillModifier(w, entity2)
	expectedMod := 1.0 + 50.0*SkillDamageBonus
	if mod2 != expectedMod {
		t.Errorf("With archery 50, modifier should be %f, got %f", expectedMod, mod2)
	}
}

func TestProjectileComponent(t *testing.T) {
	p := &components.Projectile{}
	if p.Type() != "Projectile" {
		t.Errorf("Type() should return 'Projectile', got '%s'", p.Type())
	}
}

func TestManaComponent(t *testing.T) {
	m := &components.Mana{}
	if m.Type() != "Mana" {
		t.Errorf("Type() should return 'Mana', got '%s'", m.Type())
	}
}

func TestSpellComponent(t *testing.T) {
	s := &components.Spell{}
	if s.Type() != "Spell" {
		t.Errorf("Type() should return 'Spell', got '%s'", s.Type())
	}
}

func TestSpellbookComponent(t *testing.T) {
	sb := &components.Spellbook{}
	if sb.Type() != "Spellbook" {
		t.Errorf("Type() should return 'Spellbook', got '%s'", sb.Type())
	}
}

func TestSpellEffectComponent(t *testing.T) {
	se := &components.SpellEffect{}
	if se.Type() != "SpellEffect" {
		t.Errorf("Type() should return 'SpellEffect', got '%s'", se.Type())
	}
}
