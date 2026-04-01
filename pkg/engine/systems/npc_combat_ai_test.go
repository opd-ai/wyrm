package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewNPCCombatAISystem(t *testing.T) {
	combatSystem := NewCombatSystem()
	memorySystem := NewNPCMemorySystem()
	system := NewNPCCombatAISystem(combatSystem, memorySystem)

	if system == nil {
		t.Fatal("expected system to be created")
	}
	if system.combatSystem != combatSystem {
		t.Error("combatSystem not set")
	}
	if system.memorySystem != memorySystem {
		t.Error("memorySystem not set")
	}
}

func TestNPCCombatAISystem_Update(t *testing.T) {
	world := ecs.NewWorld()
	combatSystem := NewCombatSystem()
	memorySystem := NewNPCMemorySystem()
	system := NewNPCCombatAISystem(combatSystem, memorySystem)

	// Create an NPC
	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0, Z: 0})
	world.AddComponent(npc, &components.Health{Current: 100, Max: 100})
	world.AddComponent(npc, &components.CombatState{})

	// Update should not panic with no targets
	system.Update(world, 0.1)
}

func TestNPCCombatAISystem_IsDead(t *testing.T) {
	world := ecs.NewWorld()
	system := NewNPCCombatAISystem(nil, nil)

	// Entity with no health component
	noHealth := world.CreateEntity()
	if system.isDead(world, noHealth) {
		t.Error("entity without health should not be considered dead")
	}

	// Entity with positive health
	alive := world.CreateEntity()
	world.AddComponent(alive, &components.Health{Current: 50, Max: 100})
	if system.isDead(world, alive) {
		t.Error("entity with positive health should not be dead")
	}

	// Entity with zero health
	dead := world.CreateEntity()
	world.AddComponent(dead, &components.Health{Current: 0, Max: 100})
	if !system.isDead(world, dead) {
		t.Error("entity with zero health should be dead")
	}
}

func TestNPCCombatAISystem_ShouldAttackHostileMemory(t *testing.T) {
	world := ecs.NewWorld()
	memorySystem := NewNPCMemorySystem()
	system := NewNPCCombatAISystem(nil, memorySystem)

	npc := world.CreateEntity()
	player := world.CreateEntity()

	// Initially not hostile
	if system.shouldAttackTarget(world, npc, player) {
		t.Error("should not attack neutral player")
	}

	// Set hostile disposition
	memorySystem.SetDisposition(world, npc, player, -0.8)
	if !system.shouldAttackTarget(world, npc, player) {
		t.Error("should attack hostile player")
	}
}

func TestNPCCombatAISystem_ShouldAttackHostileFaction(t *testing.T) {
	world := ecs.NewWorld()
	system := NewNPCCombatAISystem(nil, nil)

	npc := world.CreateEntity()
	player := world.CreateEntity()

	// Add factions
	world.AddComponent(npc, &components.Faction{ID: "guards"})
	world.AddComponent(npc, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"guards": {
				FactionID:  "guards",
				Reputation: -80, // Hostile reputation
			},
		},
	})
	world.AddComponent(player, &components.Faction{ID: "bandits"})

	// Should attack hostile faction
	if !system.shouldAttackTarget(world, npc, player) {
		t.Error("should attack hostile faction member")
	}

	// Same faction should not attack
	world.AddComponent(player, &components.Faction{ID: "guards"})
	if system.shouldAttackTarget(world, npc, player) {
		t.Error("should not attack same faction")
	}
}

func TestNPCCombatAISystem_GuardTargetsCriminal(t *testing.T) {
	world := ecs.NewWorld()
	system := NewNPCCombatAISystem(nil, nil)

	guard := world.CreateEntity()
	criminal := world.CreateEntity()

	// Add Guard component
	world.AddComponent(guard, &components.Guard{})

	// Criminal with wanted level
	world.AddComponent(criminal, &components.Crime{
		WantedLevel: 3,
		InJail:      false,
	})

	if !system.shouldAttackTarget(world, guard, criminal) {
		t.Error("guard should attack wanted criminal")
	}

	// Criminal in jail should not be attacked
	crimeComp, _ := world.GetComponent(criminal, "Crime")
	crime := crimeComp.(*components.Crime)
	crime.InJail = true

	if system.shouldAttackTarget(world, guard, criminal) {
		t.Error("guard should not attack criminal in jail")
	}
}

func TestNPCCombatAISystem_FindTarget(t *testing.T) {
	world := ecs.NewWorld()
	memorySystem := NewNPCMemorySystem()
	system := NewNPCCombatAISystem(nil, memorySystem)

	// Create NPC
	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0, Z: 0})
	world.AddComponent(npc, &components.Health{Current: 100, Max: 100})

	// Create hostile target in range
	hostile := world.CreateEntity()
	world.AddComponent(hostile, &components.Position{X: 5, Y: 0, Z: 0})
	world.AddComponent(hostile, &components.Health{Current: 100, Max: 100})
	memorySystem.SetDisposition(world, npc, hostile, -0.8)

	target := system.findTarget(world, npc)
	if target != hostile {
		t.Errorf("expected target %v, got %v", hostile, target)
	}

	// Target out of range should not be found
	hostilePos, _ := world.GetComponent(hostile, "Position")
	pos := hostilePos.(*components.Position)
	pos.X = 100 // Far away

	target = system.findTarget(world, npc)
	if target != 0 {
		t.Error("should not find target out of range")
	}
}

func TestNPCCombatAISystem_ProcessNPCCombat(t *testing.T) {
	world := ecs.NewWorld()
	combatSystem := NewCombatSystem()
	memorySystem := NewNPCMemorySystem()
	system := NewNPCCombatAISystem(combatSystem, memorySystem)

	// Create NPC
	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0, Z: 0})
	world.AddComponent(npc, &components.Health{Current: 100, Max: 100})
	world.AddComponent(npc, &components.CombatState{})

	// Create hostile target in combat range
	target := world.CreateEntity()
	world.AddComponent(target, &components.Position{X: 1, Y: 0, Z: 0})
	world.AddComponent(target, &components.Health{Current: 100, Max: 100})
	memorySystem.SetDisposition(world, npc, target, -0.8)

	// Process combat
	system.processNPCCombat(world, npc)

	// Check NPC is in combat
	combatComp, _ := world.GetComponent(npc, "CombatState")
	combat := combatComp.(*components.CombatState)
	if !combat.InCombat {
		t.Error("NPC should be in combat")
	}
	if combat.TargetEntity != uint64(target) {
		t.Errorf("expected target %v, got %v", uint64(target), combat.TargetEntity)
	}
}

func TestNPCCombatAISystem_SkipsDeadNPC(t *testing.T) {
	world := ecs.NewWorld()
	system := NewNPCCombatAISystem(nil, nil)

	// Create dead NPC
	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0, Z: 0})
	world.AddComponent(npc, &components.Health{Current: 0, Max: 100})
	world.AddComponent(npc, &components.CombatState{})

	// Should not panic and should exit early
	system.processNPCCombat(world, npc)

	combatComp, _ := world.GetComponent(npc, "CombatState")
	combat := combatComp.(*components.CombatState)
	if combat.InCombat {
		t.Error("dead NPC should not enter combat")
	}
}

func TestNPCCombatAISystem_DistanceSquared(t *testing.T) {
	system := NewNPCCombatAISystem(nil, nil)

	a := &components.Position{X: 0, Y: 0, Z: 0}
	b := &components.Position{X: 3, Y: 4, Z: 0}

	distSq := system.distanceSquared(a, b)
	expected := 25.0 // 3^2 + 4^2 = 9 + 16 = 25

	if distSq != expected {
		t.Errorf("expected distance squared %f, got %f", expected, distSq)
	}
}

func TestNPCCombatAISystem_SetSystems(t *testing.T) {
	system := NewNPCCombatAISystem(nil, nil)

	combatSystem := NewCombatSystem()
	memorySystem := NewNPCMemorySystem()

	system.SetCombatSystem(combatSystem)
	system.SetMemorySystem(memorySystem)

	if system.combatSystem != combatSystem {
		t.Error("combatSystem not set correctly")
	}
	if system.memorySystem != memorySystem {
		t.Error("memorySystem not set correctly")
	}
}
