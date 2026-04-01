package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewDeathPenaltySystem(t *testing.T) {
	config := DeathPenaltyConfig{
		XPLossPercent:   0.2,
		GoldLossPercent: 0.15,
	}
	system := NewDeathPenaltySystem(config)

	if system == nil {
		t.Fatal("expected system to be created")
	}
	if system.Config.XPLossPercent != 0.2 {
		t.Errorf("expected XP loss 0.2, got %f", system.Config.XPLossPercent)
	}
	if system.Config.GoldLossPercent != 0.15 {
		t.Errorf("expected gold loss 0.15, got %f", system.Config.GoldLossPercent)
	}
}

func TestNewDefaultDeathPenaltySystem(t *testing.T) {
	system := NewDefaultDeathPenaltySystem()

	if system == nil {
		t.Fatal("expected system to be created")
	}
	if system.Config.PermaDeath {
		t.Error("default should not have perma death")
	}
	if system.Config.XPLossPercent != 0.1 {
		t.Errorf("expected default XP loss 0.1, got %f", system.Config.XPLossPercent)
	}
}

func TestDeathPenaltySystem_ApplyXPLoss(t *testing.T) {
	world := ecs.NewWorld()
	system := NewDeathPenaltySystem(DeathPenaltyConfig{
		XPLossPercent: 0.5, // 50% loss
	})

	// Create entity with skills and XP
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 0, Max: 100})
	world.AddComponent(entity, &components.Skills{
		Experience: map[string]float64{
			"combat": 100.0,
			"magic":  50.0,
		},
	})

	// Process death
	system.Update(world, 0.1)

	// Check XP was reduced
	skillsComp, _ := world.GetComponent(entity, "Skills")
	skills := skillsComp.(*components.Skills)

	if skills.Experience["combat"] != 50.0 {
		t.Errorf("expected combat XP 50.0 after 50%% loss, got %f", skills.Experience["combat"])
	}
	if skills.Experience["magic"] != 25.0 {
		t.Errorf("expected magic XP 25.0 after 50%% loss, got %f", skills.Experience["magic"])
	}
}

func TestDeathPenaltySystem_ApplyGoldLoss(t *testing.T) {
	world := ecs.NewWorld()
	system := NewDeathPenaltySystem(DeathPenaltyConfig{
		GoldLossPercent: 0.2, // 20% loss
	})

	// Create entity with gold
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 0, Max: 100})
	world.AddComponent(entity, &components.Currency{Gold: 1000})

	// Process death
	system.Update(world, 0.1)

	// Check gold was reduced
	currencyComp, _ := world.GetComponent(entity, "Currency")
	currency := currencyComp.(*components.Currency)

	if currency.Gold != 800 {
		t.Errorf("expected gold 800 after 20%% loss, got %d", currency.Gold)
	}
}

func TestDeathPenaltySystem_ApplyDurabilityLoss(t *testing.T) {
	world := ecs.NewWorld()
	system := NewDeathPenaltySystem(DeathPenaltyConfig{
		DurabilityLoss: 0.1, // 10% loss
	})

	// Create entity with equipment
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 0, Max: 100})
	world.AddComponent(entity, &components.Equipment{
		Slots: map[string]*components.EquipmentSlot{
			"weapon": {
				ItemID:        "sword",
				Durability:    100,
				MaxDurability: 100,
			},
			"armor": {
				ItemID:        "chainmail",
				Durability:    80,
				MaxDurability: 100,
			},
		},
	})

	// Process death
	system.Update(world, 0.1)

	// Check durability was reduced
	equipComp, _ := world.GetComponent(entity, "Equipment")
	equip := equipComp.(*components.Equipment)

	if equip.Slots["weapon"].Durability != 90 {
		t.Errorf("expected weapon durability 90, got %f", equip.Slots["weapon"].Durability)
	}
	if equip.Slots["armor"].Durability != 70 {
		t.Errorf("expected armor durability 70, got %f", equip.Slots["armor"].Durability)
	}
}

func TestDeathPenaltySystem_CreateCorpse(t *testing.T) {
	world := ecs.NewWorld()
	system := NewDeathPenaltySystem(DeathPenaltyConfig{
		DropItems:         true,
		CorpseRetrievable: true,
	})

	// Create entity with inventory
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 0, Max: 100})
	world.AddComponent(entity, &components.Position{X: 10, Y: 20, Z: 0})
	world.AddComponent(entity, &components.Inventory{
		Items: []string{"item1", "item2"},
	})

	// Process death
	system.Update(world, 0.1)

	// Find corpse entity
	corpses := world.Entities("Corpse")
	if len(corpses) == 0 {
		t.Fatal("expected corpse to be created")
	}

	corpse := corpses[0]
	corpseComp, _ := world.GetComponent(corpse, "Corpse")
	corpseData := corpseComp.(*components.Corpse)

	if corpseData.OwnerEntity != uint64(entity) {
		t.Errorf("expected owner %v, got %v", uint64(entity), corpseData.OwnerEntity)
	}

	// Check corpse has inventory
	invComp, hasInv := world.GetComponent(corpse, "Inventory")
	if !hasInv {
		t.Fatal("corpse should have inventory")
	}
	inv := invComp.(*components.Inventory)
	if len(inv.Items) != 2 {
		t.Errorf("expected 2 items in corpse, got %d", len(inv.Items))
	}
}

func TestDeathPenaltySystem_NoDoubleProcessing(t *testing.T) {
	world := ecs.NewWorld()
	system := NewDeathPenaltySystem(DeathPenaltyConfig{
		XPLossPercent: 0.5,
	})

	// Create dead entity with skills
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 0, Max: 100})
	world.AddComponent(entity, &components.Skills{
		Experience: map[string]float64{"combat": 100.0},
	})

	// Process death twice
	system.Update(world, 0.1)
	system.Update(world, 0.1)

	// XP should only be reduced once
	skillsComp, _ := world.GetComponent(entity, "Skills")
	skills := skillsComp.(*components.Skills)

	if skills.Experience["combat"] != 50.0 {
		t.Errorf("expected combat XP 50.0 (single penalty), got %f", skills.Experience["combat"])
	}
}

func TestDeathPenaltySystem_SkipsAliveEntities(t *testing.T) {
	world := ecs.NewWorld()
	system := NewDeathPenaltySystem(DeathPenaltyConfig{
		XPLossPercent: 0.5,
	})

	// Create alive entity
	entity := world.CreateEntity()
	world.AddComponent(entity, &components.Health{Current: 100, Max: 100})
	world.AddComponent(entity, &components.Skills{
		Experience: map[string]float64{"combat": 100.0},
	})

	// Update should not affect alive entity
	system.Update(world, 0.1)

	skillsComp, _ := world.GetComponent(entity, "Skills")
	skills := skillsComp.(*components.Skills)

	if skills.Experience["combat"] != 100.0 {
		t.Errorf("alive entity XP should be unchanged, got %f", skills.Experience["combat"])
	}
}

func TestDeathPenaltySystem_SetConfig(t *testing.T) {
	system := NewDefaultDeathPenaltySystem()

	newConfig := DeathPenaltyConfig{
		PermaDeath:    true,
		XPLossPercent: 0.5,
	}
	system.SetConfig(newConfig)

	if !system.Config.PermaDeath {
		t.Error("expected perma death enabled")
	}
	if system.Config.XPLossPercent != 0.5 {
		t.Errorf("expected XP loss 0.5, got %f", system.Config.XPLossPercent)
	}
}

func TestDeathPenaltySystem_IsPermaDeath(t *testing.T) {
	system := NewDeathPenaltySystem(DeathPenaltyConfig{
		PermaDeath: true,
	})

	if !system.IsPermaDeath() {
		t.Error("expected IsPermaDeath to return true")
	}

	system.Config.PermaDeath = false
	if system.IsPermaDeath() {
		t.Error("expected IsPermaDeath to return false")
	}
}
