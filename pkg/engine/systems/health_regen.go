package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// Health regeneration constants.
const (
	// DefaultHealthRegenRate is HP regenerated per second outside combat.
	DefaultHealthRegenRate = 2.0
	// DefaultCombatCooldown is seconds after combat before regen starts.
	DefaultCombatCooldown = 5.0
	// BaseStaminaRegenRate is stamina regenerated per second.
	BaseStaminaRegenRate = 5.0
)

// HealthRegenSystem regenerates health, mana, and stamina outside of combat.
type HealthRegenSystem struct {
	// HealthRegenRate is HP regenerated per second.
	HealthRegenRate float64
	// CombatCooldown is seconds after combat before regen starts.
	CombatCooldown float64
	// ManaRegenRate is mana regenerated per second.
	ManaRegenRate float64
	// StaminaRegenRate is stamina regenerated per second.
	StaminaRegenRate float64
	// GameTime tracks elapsed game time.
	GameTime float64
}

// NewHealthRegenSystem creates a new health regeneration system.
func NewHealthRegenSystem() *HealthRegenSystem {
	return &HealthRegenSystem{
		HealthRegenRate:  DefaultHealthRegenRate,
		CombatCooldown:   DefaultCombatCooldown,
		ManaRegenRate:    DefaultManaRegenRate, // From magic_combat.go
		StaminaRegenRate: BaseStaminaRegenRate,
		GameTime:         0,
	}
}

// Update processes health regeneration for all entities.
func (s *HealthRegenSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	s.regenerateHealth(w, dt)
	s.regenerateMana(w, dt)
	s.regenerateStamina(w, dt)
}

// regenerateHealth restores health for entities not in combat.
func (s *HealthRegenSystem) regenerateHealth(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Health") {
		if !s.canRegenerate(w, e) {
			continue
		}

		healthComp, ok := w.GetComponent(e, "Health")
		if !ok {
			continue
		}
		health := healthComp.(*components.Health)

		// Skip dead entities
		if health.Current <= 0 {
			continue
		}

		// Skip if already at max
		if health.Current >= health.Max {
			continue
		}

		// Calculate regen rate (may be modified by entity's own regen rate)
		regenRate := s.getEffectiveHealthRegenRate(w, e)

		// Apply regeneration
		health.Current += regenRate * dt
		if health.Current > health.Max {
			health.Current = health.Max
		}
	}
}

// regenerateMana restores mana for entities with Mana component.
func (s *HealthRegenSystem) regenerateMana(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Mana") {
		manaComp, ok := w.GetComponent(e, "Mana")
		if !ok {
			continue
		}
		mana := manaComp.(*components.Mana)

		// Skip if already at max
		if mana.Current >= mana.Max {
			continue
		}

		// Calculate regen rate
		regenRate := s.ManaRegenRate
		if mana.RegenRate > 0 {
			regenRate = mana.RegenRate
		}

		// Apply regeneration
		mana.Current += regenRate * dt
		if mana.Current > mana.Max {
			mana.Current = mana.Max
		}
	}
}

// regenerateStamina restores stamina for entities with Stamina component.
func (s *HealthRegenSystem) regenerateStamina(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Stamina") {
		staminaComp, ok := w.GetComponent(e, "Stamina")
		if !ok {
			continue
		}
		stamina := staminaComp.(*components.Stamina)

		// Skip if already at max
		if stamina.Current >= stamina.Max {
			continue
		}

		// Calculate regen rate
		regenRate := s.StaminaRegenRate
		if stamina.RegenRate > 0 {
			regenRate = stamina.RegenRate
		}

		// Apply regeneration
		stamina.Current += regenRate * dt
		if stamina.Current > stamina.Max {
			stamina.Current = stamina.Max
		}
	}
}

// canRegenerate checks if an entity can regenerate health.
func (s *HealthRegenSystem) canRegenerate(w *ecs.World, entity ecs.Entity) bool {
	// Check if entity is in combat
	combatComp, hasCombat := w.GetComponent(entity, "CombatState")
	if hasCombat {
		combat := combatComp.(*components.CombatState)
		if combat.InCombat {
			return false
		}

		// Check if combat recently ended (within cooldown period)
		timeSinceCombat := s.GameTime - combat.LastAttackTime
		if timeSinceCombat < s.CombatCooldown {
			return false
		}
	}

	return true
}

// getEffectiveHealthRegenRate returns the health regen rate for an entity.
func (s *HealthRegenSystem) getEffectiveHealthRegenRate(w *ecs.World, entity ecs.Entity) float64 {
	baseRate := s.HealthRegenRate

	// Check for skills that boost regen
	skillsComp, hasSkills := w.GetComponent(entity, "Skills")
	if hasSkills {
		skills := skillsComp.(*components.Skills)
		if skills.Levels != nil {
			// Check for regeneration-related skills
			for _, skillID := range []string{"regeneration", "vitality", "endurance", "constitution"} {
				if level, exists := skills.Levels[skillID]; exists {
					// Each skill level adds 10% to regen rate
					baseRate *= 1.0 + float64(level)*0.1
					break
				}
			}
		}
	}

	return baseRate
}

// SetHealthRegenRate sets the base health regeneration rate.
func (s *HealthRegenSystem) SetHealthRegenRate(rate float64) {
	s.HealthRegenRate = rate
}

// SetCombatCooldown sets the combat cooldown before regen starts.
func (s *HealthRegenSystem) SetCombatCooldown(cooldown float64) {
	s.CombatCooldown = cooldown
}

// SetManaRegenRate sets the base mana regeneration rate.
func (s *HealthRegenSystem) SetManaRegenRate(rate float64) {
	s.ManaRegenRate = rate
}

// SetStaminaRegenRate sets the base stamina regeneration rate.
func (s *HealthRegenSystem) SetStaminaRegenRate(rate float64) {
	s.StaminaRegenRate = rate
}
