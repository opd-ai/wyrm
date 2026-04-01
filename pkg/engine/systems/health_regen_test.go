package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewHealthRegenSystem(t *testing.T) {
	system := NewHealthRegenSystem()

	if system == nil {
		t.Fatal("expected system to be created")
	}
	if system.HealthRegenRate != DefaultHealthRegenRate {
		t.Errorf("expected health regen rate %f, got %f", DefaultHealthRegenRate, system.HealthRegenRate)
	}
	if system.CombatCooldown != DefaultCombatCooldown {
		t.Errorf("expected combat cooldown %f, got %f", DefaultCombatCooldown, system.CombatCooldown)
	}
}

func TestHealthRegenSystem_RegenerateHealth(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()

	// Create entity with damaged health
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 50, Max: 100})

	// Update should regenerate health
	system.Update(world, 1.0)

	healthComp, _ := world.GetComponent(entity, "Health")
	health := healthComp.(*components.Health)

	expected := 50.0 + DefaultHealthRegenRate
	if health.Current != expected {
		t.Errorf("expected health %f, got %f", expected, health.Current)
	}
}

func TestHealthRegenSystem_NoRegenAtMaxHealth(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()

	// Create entity at max health
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 100, Max: 100})

	system.Update(world, 1.0)

	healthComp, _ := world.GetComponent(entity, "Health")
	health := healthComp.(*components.Health)

	// Should stay at max
	if health.Current != 100 {
		t.Errorf("expected health 100, got %f", health.Current)
	}
}

func TestHealthRegenSystem_NoRegenWhileDead(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()

	// Create dead entity
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 0, Max: 100})

	system.Update(world, 1.0)

	healthComp, _ := world.GetComponent(entity, "Health")
	health := healthComp.(*components.Health)

	// Should stay at 0
	if health.Current != 0 {
		t.Errorf("expected health 0, got %f", health.Current)
	}
}

func TestHealthRegenSystem_NoRegenInCombat(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()

	// Create entity in combat
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 50, Max: 100})
	world.AddComponent(entity, &components.CombatState{InCombat: true})

	system.Update(world, 1.0)

	healthComp, _ := world.GetComponent(entity, "Health")
	health := healthComp.(*components.Health)

	// Should not regenerate while in combat
	if health.Current != 50 {
		t.Errorf("expected health 50 (no regen in combat), got %f", health.Current)
	}
}

func TestHealthRegenSystem_RegenAfterCombatCooldown(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()
	system.CombatCooldown = 5.0

	// Create entity recently out of combat
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 50, Max: 100})
	world.AddComponent(entity, &components.CombatState{
		InCombat:       false,
		LastAttackTime: 0, // Combat ended at time 0
	})

	// Advance game time past cooldown
	system.GameTime = 10.0
	system.Update(world, 1.0)

	healthComp, _ := world.GetComponent(entity, "Health")
	health := healthComp.(*components.Health)

	// Should regenerate after cooldown
	if health.Current <= 50 {
		t.Errorf("expected health > 50 after cooldown, got %f", health.Current)
	}
}

func TestHealthRegenSystem_NoRegenDuringCooldown(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()
	system.CombatCooldown = 5.0

	// Create entity recently out of combat
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 50, Max: 100})
	world.AddComponent(entity, &components.CombatState{
		InCombat:       false,
		LastAttackTime: 2.0, // Combat ended at time 2
	})

	// Game time is still within cooldown
	system.GameTime = 4.0
	system.Update(world, 1.0)

	healthComp, _ := world.GetComponent(entity, "Health")
	health := healthComp.(*components.Health)

	// Should not regenerate during cooldown
	if health.Current != 50 {
		t.Errorf("expected health 50 (in cooldown), got %f", health.Current)
	}
}

func TestHealthRegenSystem_RegenerateMana(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()

	// Create entity with depleted mana
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Mana{Current: 30, Max: 100, RegenRate: 5.0})

	system.Update(world, 1.0)

	manaComp, _ := world.GetComponent(entity, "Mana")
	mana := manaComp.(*components.Mana)

	// Should regenerate using entity's regen rate
	if mana.Current != 35.0 {
		t.Errorf("expected mana 35, got %f", mana.Current)
	}
}

func TestHealthRegenSystem_RegenerateStamina(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()

	// Create entity with depleted stamina
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Stamina{Current: 50, Max: 100, RegenRate: 10.0})

	system.Update(world, 1.0)

	staminaComp, _ := world.GetComponent(entity, "Stamina")
	stamina := staminaComp.(*components.Stamina)

	// Should regenerate using entity's regen rate
	if stamina.Current != 60.0 {
		t.Errorf("expected stamina 60, got %f", stamina.Current)
	}
}

func TestHealthRegenSystem_HealthCapsAtMax(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()
	system.HealthRegenRate = 100.0 // High regen rate

	// Create entity with damaged health
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 90, Max: 100})

	system.Update(world, 1.0)

	healthComp, _ := world.GetComponent(entity, "Health")
	health := healthComp.(*components.Health)

	// Should cap at max
	if health.Current != 100 {
		t.Errorf("expected health 100 (capped), got %f", health.Current)
	}
}

func TestHealthRegenSystem_SkillBoostsRegen(t *testing.T) {
	world := ecs.NewWorld()
	system := NewHealthRegenSystem()

	// Create entity with regeneration skill
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 50, Max: 100})
	world.AddComponent(entity, &components.Skills{
		Levels: map[string]int{
			"regeneration": 2, // Level 2 adds 20% bonus
		},
	})

	system.Update(world, 1.0)

	healthComp, _ := world.GetComponent(entity, "Health")
	health := healthComp.(*components.Health)

	// Should regenerate with skill bonus
	expectedRate := DefaultHealthRegenRate * 1.2
	expected := 50.0 + expectedRate
	if health.Current != expected {
		t.Errorf("expected health %f (with skill bonus), got %f", expected, health.Current)
	}
}

func TestHealthRegenSystem_SetRates(t *testing.T) {
	system := NewHealthRegenSystem()

	system.SetHealthRegenRate(5.0)
	system.SetCombatCooldown(10.0)
	system.SetManaRegenRate(3.0)
	system.SetStaminaRegenRate(8.0)

	if system.HealthRegenRate != 5.0 {
		t.Errorf("expected health regen rate 5.0, got %f", system.HealthRegenRate)
	}
	if system.CombatCooldown != 10.0 {
		t.Errorf("expected combat cooldown 10.0, got %f", system.CombatCooldown)
	}
	if system.ManaRegenRate != 3.0 {
		t.Errorf("expected mana regen rate 3.0, got %f", system.ManaRegenRate)
	}
	if system.StaminaRegenRate != 8.0 {
		t.Errorf("expected stamina regen rate 8.0, got %f", system.StaminaRegenRate)
	}
}
