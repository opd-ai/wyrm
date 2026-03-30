package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// Magic combat constants.
const (
	// DefaultManaRegenRate is mana regenerated per second.
	DefaultManaRegenRate = 2.0
	// DefaultSpellCooldown is the minimum time between spell casts.
	DefaultSpellCooldown = 1.0
	// DefaultSpellRange is the default casting distance.
	DefaultSpellRange = 30.0
)

// MagicSystem handles mana regeneration, spell casting, and spell effects.
type MagicSystem struct {
	// GameTime tracks time for cooldowns and effect duration.
	GameTime float64
}

// NewMagicSystem creates a new magic system.
func NewMagicSystem() *MagicSystem {
	return &MagicSystem{GameTime: 0}
}

// Update processes mana regeneration and active spell effects.
func (s *MagicSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	s.updateManaRegeneration(w, dt)
	s.updateSpellEffects(w, dt)
}

// updateManaRegeneration regenerates mana for all entities with Mana component.
func (s *MagicSystem) updateManaRegeneration(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Mana") {
		manaComp, ok := w.GetComponent(e, "Mana")
		if !ok {
			continue
		}
		mana := manaComp.(*components.Mana)

		if mana.Current < mana.Max {
			mana.Current += mana.RegenRate * dt
			if mana.Current > mana.Max {
				mana.Current = mana.Max
			}
		}
	}
}

// updateSpellEffects processes active spell effects (buffs, debuffs, DoTs).
func (s *MagicSystem) updateSpellEffects(w *ecs.World, dt float64) {
	for _, e := range w.Entities("SpellEffect") {
		effectComp, ok := w.GetComponent(e, "SpellEffect")
		if !ok {
			continue
		}
		effect := effectComp.(*components.SpellEffect)

		// Apply effect per tick
		s.applyEffectTick(w, e, effect, dt)

		// Reduce duration
		effect.Remaining -= dt
		if effect.Remaining <= 0 {
			// Remove expired effect
			w.RemoveComponent(e, "SpellEffect")
		}
	}
}

// applyEffectTick applies the per-tick effect of a spell effect.
func (s *MagicSystem) applyEffectTick(w *ecs.World, entity ecs.Entity, effect *components.SpellEffect, dt float64) {
	healthComp, ok := w.GetComponent(entity, "Health")
	if !ok {
		return
	}
	health := healthComp.(*components.Health)

	switch effect.EffectType {
	case "damage", "burn", "poison", "bleed":
		// Damage over time
		health.Current -= effect.Magnitude * dt
		if health.Current < 0 {
			health.Current = 0
		}
	case "heal", "regen":
		// Heal over time
		health.Current += effect.Magnitude * dt
		if health.Current > health.Max {
			health.Current = health.Max
		}
	}
}

// CastSpell attempts to cast a spell at a target entity.
func (s *MagicSystem) CastSpell(w *ecs.World, caster ecs.Entity, spellID string, targetEntity ecs.Entity, projectileSystem *ProjectileSystem) bool {
	// Get caster's spellbook
	spellbookComp, ok := w.GetComponent(caster, "Spellbook")
	if !ok {
		return false
	}
	spellbook := spellbookComp.(*components.Spellbook)

	// Get the spell
	spell, exists := spellbook.Spells[spellID]
	if !exists {
		return false
	}

	// Check cooldown (allow first cast when LastCast is 0)
	if spell.LastCast > 0 && s.GameTime-spell.LastCast < spell.Cooldown {
		return false
	}

	// Get caster mana
	manaComp, manaOK := w.GetComponent(caster, "Mana")
	if !manaOK {
		return false
	}
	mana := manaComp.(*components.Mana)

	// Check mana cost
	if mana.Current < spell.ManaCost {
		return false
	}

	// Check range
	if !s.isInSpellRange(w, caster, targetEntity, spell.Range) {
		return false
	}

	// Consume mana
	mana.Current -= spell.ManaCost

	// Update cooldown
	spell.LastCast = s.GameTime

	// Apply spell effect
	if spell.ProjectileSpeed > 0 && projectileSystem != nil {
		// Projectile spell
		s.castProjectileSpell(w, caster, targetEntity, spell, projectileSystem)
	} else {
		// Instant spell
		s.applySpellEffect(w, caster, targetEntity, spell)
	}

	return true
}

// CastSpellAtPosition casts a spell at a world position (for AoE or projectiles).
func (s *MagicSystem) CastSpellAtPosition(w *ecs.World, caster ecs.Entity, spellID string, targetX, targetY, targetZ float64, projectileSystem *ProjectileSystem) bool {
	// Get caster's spellbook
	spellbookComp, ok := w.GetComponent(caster, "Spellbook")
	if !ok {
		return false
	}
	spellbook := spellbookComp.(*components.Spellbook)

	// Get the spell
	spell, exists := spellbook.Spells[spellID]
	if !exists {
		return false
	}

	// Check cooldown (allow first cast when LastCast is 0)
	if spell.LastCast > 0 && s.GameTime-spell.LastCast < spell.Cooldown {
		return false
	}

	// Get caster mana
	manaComp, manaOK := w.GetComponent(caster, "Mana")
	if !manaOK {
		return false
	}
	mana := manaComp.(*components.Mana)

	// Check mana cost
	if mana.Current < spell.ManaCost {
		return false
	}

	// Check range
	casterPos := s.getPosition(w, caster)
	if casterPos == nil {
		return false
	}
	dx := targetX - casterPos.X
	dy := targetY - casterPos.Y
	dz := targetZ - casterPos.Z
	distSq := dx*dx + dy*dy + dz*dz
	if distSq > spell.Range*spell.Range {
		return false
	}

	// Consume mana
	mana.Current -= spell.ManaCost

	// Update cooldown
	spell.LastCast = s.GameTime

	// Cast the spell
	if spell.ProjectileSpeed > 0 && projectileSystem != nil {
		// Spawn spell projectile
		proj := projectileSystem.SpawnProjectile(w, caster, targetX, targetY, targetZ, spell.Magnitude, spell.ProjectileSpeed, "spell")
		if proj != 0 && spell.AreaOfEffect > 0 {
			// Mark projectile for AoE on impact
			projComp, _ := w.GetComponent(proj, "Projectile")
			if projComp != nil {
				projectile := projComp.(*components.Projectile)
				projectile.HitRadius = spell.AreaOfEffect
				projectile.PierceCount = 0 // AoE hits all in radius
			}
		}
	} else if spell.AreaOfEffect > 0 {
		// Instant AoE spell
		s.applyAoESpell(w, caster, targetX, targetY, targetZ, spell)
	}

	return true
}

// castProjectileSpell creates a spell projectile toward a target.
func (s *MagicSystem) castProjectileSpell(w *ecs.World, caster, target ecs.Entity, spell *components.Spell, projectileSystem *ProjectileSystem) {
	targetPos := s.getPosition(w, target)
	if targetPos == nil {
		return
	}

	proj := projectileSystem.SpawnProjectile(w, caster, targetPos.X, targetPos.Y, targetPos.Z, spell.Magnitude, spell.ProjectileSpeed, "spell")
	if proj != 0 && spell.AreaOfEffect > 0 {
		projComp, _ := w.GetComponent(proj, "Projectile")
		if projComp != nil {
			projectile := projComp.(*components.Projectile)
			projectile.HitRadius = spell.AreaOfEffect
			projectile.PierceCount = 0
		}
	}
}

// applySpellEffect applies an instant spell effect to a target.
func (s *MagicSystem) applySpellEffect(w *ecs.World, caster, target ecs.Entity, spell *components.Spell) {
	switch spell.EffectType {
	case "damage":
		s.applyDamage(w, target, spell.Magnitude, caster)
	case "heal":
		s.applyHeal(w, target, spell.Magnitude)
	case "buff", "debuff", "burn", "poison", "regen":
		s.applyStatusEffect(w, target, spell)
	}
}

// applyAoESpell applies an area-of-effect spell at a position.
func (s *MagicSystem) applyAoESpell(w *ecs.World, caster ecs.Entity, x, y, z float64, spell *components.Spell) {
	radiusSq := spell.AreaOfEffect * spell.AreaOfEffect

	for _, target := range w.Entities("Health", "Position") {
		if target == caster {
			continue
		}

		targetPos := s.getPosition(w, target)
		if targetPos == nil {
			continue
		}

		dx := targetPos.X - x
		dy := targetPos.Y - y
		dz := targetPos.Z - z
		distSq := dx*dx + dy*dy + dz*dz

		if distSq <= radiusSq {
			// Apply damage that falls off with distance
			falloff := 1.0 - math.Sqrt(distSq)/spell.AreaOfEffect
			if falloff < 0.25 {
				falloff = 0.25 // Minimum 25% damage at edge
			}
			damage := spell.Magnitude * falloff
			s.applyDamage(w, target, damage, caster)
		}
	}
}

// applyDamage deals damage to a target.
func (s *MagicSystem) applyDamage(w *ecs.World, target ecs.Entity, damage float64, source ecs.Entity) {
	healthComp, ok := w.GetComponent(target, "Health")
	if !ok {
		return
	}
	health := healthComp.(*components.Health)
	health.Current -= damage
	if health.Current < 0 {
		health.Current = 0
	}
	_ = source // Could be used for kill tracking
}

// applyHeal restores health to a target.
func (s *MagicSystem) applyHeal(w *ecs.World, target ecs.Entity, amount float64) {
	healthComp, ok := w.GetComponent(target, "Health")
	if !ok {
		return
	}
	health := healthComp.(*components.Health)
	health.Current += amount
	if health.Current > health.Max {
		health.Current = health.Max
	}
}

// applyStatusEffect applies a status effect to a target.
func (s *MagicSystem) applyStatusEffect(w *ecs.World, target ecs.Entity, spell *components.Spell) {
	// Create SpellEffect component
	effect := &components.SpellEffect{
		EffectType: spell.EffectType,
		Magnitude:  spell.Magnitude / spell.Cooldown, // DPS based on spell power over cooldown
		Duration:   spell.Cooldown,                   // Duration equals cooldown
		Remaining:  spell.Cooldown,
		Source:     0, // Could track caster
	}
	w.AddComponent(target, effect)
}

// isInSpellRange checks if a target is within casting range.
func (s *MagicSystem) isInSpellRange(w *ecs.World, caster, target ecs.Entity, spellRange float64) bool {
	casterPos := s.getPosition(w, caster)
	targetPos := s.getPosition(w, target)
	if casterPos == nil || targetPos == nil {
		return false
	}

	dx := targetPos.X - casterPos.X
	dy := targetPos.Y - casterPos.Y
	dz := targetPos.Z - casterPos.Z
	distSq := dx*dx + dy*dy + dz*dz

	return distSq <= spellRange*spellRange
}

// getPosition retrieves an entity's position component.
func (s *MagicSystem) getPosition(w *ecs.World, entity ecs.Entity) *components.Position {
	posComp, ok := w.GetComponent(entity, "Position")
	if !ok {
		return nil
	}
	return posComp.(*components.Position)
}

// GetMagicSkillModifier returns damage multiplier for magic skills.
func (s *MagicSystem) GetMagicSkillModifier(w *ecs.World, entity ecs.Entity) float64 {
	skillsComp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return 1.0
	}
	skills := skillsComp.(*components.Skills)
	if skills.Levels == nil {
		return 1.0
	}

	// Check for magic-related skills
	magicLevel := 0
	for _, skillID := range []string{"magic", "destruction", "conjuration", "fire_magic", "frost_magic", "shock_magic", "psi_ops", "hacking", "occult"} {
		if level, exists := skills.Levels[skillID]; exists && level > magicLevel {
			magicLevel = level
		}
	}

	return 1.0 + float64(magicLevel)*SkillDamageBonus
}

// LearnSpell adds a spell to an entity's spellbook.
func (s *MagicSystem) LearnSpell(w *ecs.World, entity ecs.Entity, spell *components.Spell) bool {
	spellbookComp, ok := w.GetComponent(entity, "Spellbook")
	if !ok {
		// Create spellbook if not exists
		spellbook := &components.Spellbook{
			Spells: make(map[string]*components.Spell),
		}
		w.AddComponent(entity, spellbook)
		spellbookComp, _ = w.GetComponent(entity, "Spellbook")
	}
	spellbook := spellbookComp.(*components.Spellbook)

	if spellbook.Spells == nil {
		spellbook.Spells = make(map[string]*components.Spell)
	}

	spellbook.Spells[spell.ID] = spell
	return true
}

// SetActiveSpell sets the currently selected spell.
func (s *MagicSystem) SetActiveSpell(w *ecs.World, entity ecs.Entity, spellID string) bool {
	spellbookComp, ok := w.GetComponent(entity, "Spellbook")
	if !ok {
		return false
	}
	spellbook := spellbookComp.(*components.Spellbook)

	if _, exists := spellbook.Spells[spellID]; !exists {
		return false
	}

	spellbook.ActiveSpellID = spellID
	return true
}
