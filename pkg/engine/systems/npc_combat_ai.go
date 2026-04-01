package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// NPC Combat AI constants.
const (
	// NPCAICombatRange is the distance at which NPCs can engage in combat.
	NPCAICombatRange = 3.0
	// NPCAIAggroRange is the distance at which hostile NPCs become aggressive.
	NPCAIAggroRange = 8.0
	// NPCAISearchRange is the distance at which NPCs search for targets.
	NPCAISearchRange = 10.0
	// NPCAIAttackCooldown is the minimum time between NPC attacks in seconds.
	NPCAIAttackCooldown = 1.5
	// HostileDispositionThreshold is the disposition below which an NPC is hostile.
	HostileDispositionThreshold = -0.5
)

// NPCCombatAISystem controls NPC combat behavior including target selection and attacks.
type NPCCombatAISystem struct {
	combatSystem *CombatSystem
	memorySystem *NPCMemorySystem
	GameTime     float64
}

// NewNPCCombatAISystem creates a new NPC combat AI system.
func NewNPCCombatAISystem(combatSystem *CombatSystem, memorySystem *NPCMemorySystem) *NPCCombatAISystem {
	return &NPCCombatAISystem{
		combatSystem: combatSystem,
		memorySystem: memorySystem,
		GameTime:     0,
	}
}

// Update processes NPC combat decisions each tick.
func (s *NPCCombatAISystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt

	// Process all NPCs with combat capability
	for _, npc := range w.Entities("CombatState", "Position", "Health") {
		s.processNPCCombat(w, npc)
	}
}

// processNPCCombat handles combat logic for a single NPC.
func (s *NPCCombatAISystem) processNPCCombat(w *ecs.World, npc ecs.Entity) {
	// Skip dead NPCs
	if s.isDead(w, npc) {
		return
	}

	// Get combat state
	combatComp, ok := w.GetComponent(npc, "CombatState")
	if !ok {
		return
	}
	combat := combatComp.(*components.CombatState)

	// If already attacking, let combat system handle it
	if combat.IsAttacking {
		return
	}

	// Check if on attack cooldown
	if combat.Cooldown > 0 {
		return
	}

	// Find a target
	target := s.findTarget(w, npc)
	if target == 0 {
		combat.InCombat = false
		combat.TargetEntity = 0
		return
	}

	// Set combat state
	combat.InCombat = true
	combat.TargetEntity = uint64(target)

	// Attempt to attack if in range
	if s.combatSystem != nil {
		s.combatSystem.InitiateAttack(w, npc, target)
	}
}

// findTarget finds a valid combat target for the NPC.
func (s *NPCCombatAISystem) findTarget(w *ecs.World, npc ecs.Entity) ecs.Entity {
	npcPos := s.getPosition(w, npc)
	if npcPos == nil {
		return 0
	}

	var bestTarget ecs.Entity
	bestDistSq := math.MaxFloat64

	// Look for targets with Health component
	for _, target := range w.Entities("Health", "Position") {
		if target == npc {
			continue
		}

		// Skip dead targets
		if s.isDead(w, target) {
			continue
		}

		// Check if target is within aggro range
		targetPos := s.getPosition(w, target)
		if targetPos == nil {
			continue
		}

		distSq := s.distanceSquared(npcPos, targetPos)
		if distSq > NPCAIAggroRange*NPCAIAggroRange {
			continue
		}

		// Check hostility
		if !s.shouldAttackTarget(w, npc, target) {
			continue
		}

		// Track closest valid target
		if distSq < bestDistSq {
			bestDistSq = distSq
			bestTarget = target
		}
	}

	return bestTarget
}

// shouldAttackTarget determines if the NPC should attack a target.
func (s *NPCCombatAISystem) shouldAttackTarget(w *ecs.World, npc, target ecs.Entity) bool {
	// Check memory-based hostility
	if s.memorySystem != nil && s.memorySystem.IsHostile(w, npc, target) {
		return true
	}

	// Check faction-based hostility
	if s.areFactionsHostile(w, npc, target) {
		return true
	}

	// Check if NPC has a Guard component and target is wanted
	if s.isGuardTargetingCriminal(w, npc, target) {
		return true
	}

	return false
}

// areFactionsHostile checks if two entities are from hostile factions.
func (s *NPCCombatAISystem) areFactionsHostile(w *ecs.World, entityA, entityB ecs.Entity) bool {
	factionA := s.getFactionID(w, entityA)
	factionB := s.getFactionID(w, entityB)

	// No faction = no factional hostility
	if factionA == "" || factionB == "" {
		return false
	}

	// Same faction = never hostile
	if factionA == factionB {
		return false
	}

	// Check faction membership relations
	factionCompA, okA := w.GetComponent(entityA, "FactionMembership")
	if okA {
		membership := factionCompA.(*components.FactionMembership)
		if membership.Memberships != nil {
			for _, info := range membership.Memberships {
				if info.FactionID == factionA && info.Reputation < HostileDispositionThreshold*100 {
					// This entity is hostile
					return true
				}
			}
		}
	}

	return false
}

// isGuardTargetingCriminal checks if NPC is a guard and target has crimes.
func (s *NPCCombatAISystem) isGuardTargetingCriminal(w *ecs.World, npc, target ecs.Entity) bool {
	// Check if NPC is a guard
	_, hasGuard := w.GetComponent(npc, "Guard")
	if !hasGuard {
		return false
	}

	// Check if target has crimes
	crimeComp, hasCrime := w.GetComponent(target, "Crime")
	if !hasCrime {
		return false
	}
	crime := crimeComp.(*components.Crime)

	// Only attack if wanted level > 0 and not already in jail
	return crime.WantedLevel > 0 && !crime.InJail
}

// isDead checks if an entity is dead.
func (s *NPCCombatAISystem) isDead(w *ecs.World, entity ecs.Entity) bool {
	healthComp, ok := w.GetComponent(entity, "Health")
	if !ok {
		return false
	}
	health := healthComp.(*components.Health)
	return health.Current <= 0
}

// getPosition retrieves an entity's position.
func (s *NPCCombatAISystem) getPosition(w *ecs.World, entity ecs.Entity) *components.Position {
	posComp, ok := w.GetComponent(entity, "Position")
	if !ok {
		return nil
	}
	return posComp.(*components.Position)
}

// getFactionID retrieves an entity's faction ID.
func (s *NPCCombatAISystem) getFactionID(w *ecs.World, entity ecs.Entity) string {
	factionComp, ok := w.GetComponent(entity, "Faction")
	if !ok {
		return ""
	}
	faction := factionComp.(*components.Faction)
	return faction.ID
}

// distanceSquared calculates squared distance between two positions.
func (s *NPCCombatAISystem) distanceSquared(a, b *components.Position) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	dz := b.Z - a.Z
	return dx*dx + dy*dy + dz*dz
}

// SetCombatSystem sets the combat system reference.
func (s *NPCCombatAISystem) SetCombatSystem(cs *CombatSystem) {
	s.combatSystem = cs
}

// SetMemorySystem sets the memory system reference.
func (s *NPCCombatAISystem) SetMemorySystem(ms *NPCMemorySystem) {
	s.memorySystem = ms
}
