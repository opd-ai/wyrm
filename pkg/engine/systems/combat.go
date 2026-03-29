package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// CombatSystem handles combat resolution and damage.
type CombatSystem struct {
	// DefaultMeleeRange is the default attack range if no weapon equipped.
	DefaultMeleeRange float64
	// DefaultDamage is the base unarmed damage.
	DefaultDamage float64
	// GameTime tracks the current game time for cooldowns.
	GameTime float64
}

// NewCombatSystem creates a new combat system with default settings.
func NewCombatSystem() *CombatSystem {
	return &CombatSystem{
		DefaultMeleeRange: DefaultMeleeRangeUnits,
		DefaultDamage:     DefaultBaseDamage,
		GameTime:          0,
	}
}

// Update processes combat resolution each tick.
func (s *CombatSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	s.processCooldowns(w, dt)
	s.processActiveAttacks(w)
	s.clampHealth(w)
}

// processCooldowns reduces attack cooldowns for all combat entities.
func (s *CombatSystem) processCooldowns(w *ecs.World, dt float64) {
	for _, e := range w.Entities("CombatState") {
		comp, ok := w.GetComponent(e, "CombatState")
		if !ok {
			continue
		}
		combat := comp.(*components.CombatState)
		if combat.Cooldown > 0 {
			combat.Cooldown -= dt
			if combat.Cooldown < 0 {
				combat.Cooldown = 0
			}
		}
	}
}

// processActiveAttacks resolves pending attacks.
func (s *CombatSystem) processActiveAttacks(w *ecs.World) {
	for _, e := range w.Entities("CombatState", "Position") {
		comp, ok := w.GetComponent(e, "CombatState")
		if !ok {
			continue
		}
		combat := comp.(*components.CombatState)
		if !combat.IsAttacking || combat.TargetEntity == 0 {
			continue
		}

		// Resolve the attack
		s.resolveAttack(w, e, ecs.Entity(combat.TargetEntity))
		combat.IsAttacking = false
		combat.TargetEntity = 0
	}
}

// clampHealth ensures health doesn't exceed max for any entity.
func (s *CombatSystem) clampHealth(w *ecs.World) {
	for _, e := range w.Entities("Health") {
		comp, ok := w.GetComponent(e, "Health")
		if !ok {
			continue
		}
		health := comp.(*components.Health)
		if health.Current > health.Max {
			health.Current = health.Max
		}
		if health.Current < 0 {
			health.Current = 0
		}
	}
}

// InitiateAttack starts an attack against a target if in range.
func (s *CombatSystem) InitiateAttack(w *ecs.World, attacker, target ecs.Entity) bool {
	// Check attacker has combat state
	combatComp, ok := w.GetComponent(attacker, "CombatState")
	if !ok {
		return false
	}
	combat := combatComp.(*components.CombatState)

	// Check cooldown
	if combat.Cooldown > 0 {
		return false
	}

	// Check range
	if !s.isInMeleeRange(w, attacker, target) {
		return false
	}

	// Set up the attack
	combat.IsAttacking = true
	combat.TargetEntity = uint64(target)
	combat.InCombat = true
	combat.LastAttackTime = s.GameTime

	// Set cooldown based on weapon or default
	combat.Cooldown = s.getAttackCooldown(w, attacker)

	return true
}

// isInMeleeRange checks if target is within melee range of attacker.
func (s *CombatSystem) isInMeleeRange(w *ecs.World, attacker, target ecs.Entity) bool {
	attackerPos := s.getPosition(w, attacker)
	targetPos := s.getPosition(w, target)
	if attackerPos == nil || targetPos == nil {
		return false
	}

	meleeRange := s.getMeleeRange(w, attacker)
	dx := targetPos.X - attackerPos.X
	dy := targetPos.Y - attackerPos.Y
	distSq := dx*dx + dy*dy

	return distSq <= meleeRange*meleeRange
}

// getMeleeRange returns the melee attack range for an entity.
func (s *CombatSystem) getMeleeRange(w *ecs.World, entity ecs.Entity) float64 {
	weaponComp, ok := w.GetComponent(entity, "Weapon")
	if !ok {
		return s.DefaultMeleeRange
	}
	weapon := weaponComp.(*components.Weapon)
	if weapon.WeaponType != "melee" && weapon.WeaponType != "" {
		return s.DefaultMeleeRange
	}
	return weapon.Range
}

// getAttackCooldown returns the attack cooldown for an entity.
func (s *CombatSystem) getAttackCooldown(w *ecs.World, entity ecs.Entity) float64 {
	weaponComp, ok := w.GetComponent(entity, "Weapon")
	if !ok {
		return DefaultAttackCooldown
	}
	weapon := weaponComp.(*components.Weapon)
	if weapon.AttackSpeed <= 0 {
		return DefaultAttackCooldown
	}
	return DefaultAttackCooldown / weapon.AttackSpeed
}

// resolveAttack applies damage from attacker to target.
func (s *CombatSystem) resolveAttack(w *ecs.World, attacker, target ecs.Entity) {
	// Get target health
	healthComp, ok := w.GetComponent(target, "Health")
	if !ok {
		return
	}
	health := healthComp.(*components.Health)

	// Calculate damage
	damage := s.calculateDamage(w, attacker)

	// Apply damage
	health.Current -= damage
}

// calculateDamage computes the damage for an attack.
func (s *CombatSystem) calculateDamage(w *ecs.World, attacker ecs.Entity) float64 {
	baseDamage := s.DefaultDamage

	// Get weapon damage if equipped
	weaponComp, ok := w.GetComponent(attacker, "Weapon")
	if ok {
		weapon := weaponComp.(*components.Weapon)
		baseDamage = weapon.Damage
	}

	// Apply skill modifiers if available
	skillMod := s.getSkillModifier(w, attacker)
	return baseDamage * skillMod
}

// getSkillModifier returns a damage multiplier based on combat skills.
func (s *CombatSystem) getSkillModifier(w *ecs.World, entity ecs.Entity) float64 {
	skillsComp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return BasePriceMultiplier
	}
	skills := skillsComp.(*components.Skills)
	if skills.Levels == nil {
		return BasePriceMultiplier
	}

	// Check for combat-related skills
	combatLevel := 0
	for _, skillID := range []string{"melee", "combat", "strength", "blade"} {
		if level, exists := skills.Levels[skillID]; exists && level > combatLevel {
			combatLevel = level
		}
	}

	// Each skill level adds damage bonus
	return BasePriceMultiplier + float64(combatLevel)*SkillDamageBonus
}

// getPosition retrieves an entity's position component.
func (s *CombatSystem) getPosition(w *ecs.World, entity ecs.Entity) *components.Position {
	posComp, ok := w.GetComponent(entity, "Position")
	if !ok {
		return nil
	}
	return posComp.(*components.Position)
}

// FindNearestTarget finds the closest attackable entity within range.
func (s *CombatSystem) FindNearestTarget(w *ecs.World, attacker ecs.Entity) ecs.Entity {
	attackerPos := s.getPosition(w, attacker)
	if attackerPos == nil {
		return 0
	}

	meleeRange := s.getMeleeRange(w, attacker)
	var nearestTarget ecs.Entity
	nearestDistSq := math.MaxFloat64

	for _, target := range w.Entities("Health", "Position") {
		if target == attacker {
			continue
		}

		targetPos := s.getPosition(w, target)
		if targetPos == nil {
			continue
		}

		dx := targetPos.X - attackerPos.X
		dy := targetPos.Y - attackerPos.Y
		distSq := dx*dx + dy*dy

		if distSq <= meleeRange*meleeRange && distSq < nearestDistSq {
			nearestDistSq = distSq
			nearestTarget = target
		}
	}

	return nearestTarget
}
