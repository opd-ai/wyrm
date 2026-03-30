package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// Ranged combat constants.
const (
	// MeleeWeaponRangeThreshold separates melee from ranged weapons.
	MeleeWeaponRangeThreshold = 3.0
	// DefaultProjectileSpeed is the base projectile speed (units per second).
	DefaultProjectileSpeed = 20.0
	// DefaultProjectileLifetime is how long projectiles last (seconds).
	DefaultProjectileLifetime = 5.0
	// DefaultProjectileHitRadius is the default collision radius.
	DefaultProjectileHitRadius = 0.5
)

// ProjectileSystem handles projectile movement, collision, and lifetime.
type ProjectileSystem struct {
	// GameTime tracks time for lifetime calculations.
	GameTime float64
}

// NewProjectileSystem creates a new projectile system.
func NewProjectileSystem() *ProjectileSystem {
	return &ProjectileSystem{GameTime: 0}
}

// Update processes all projectiles: movement, collision, and cleanup.
func (s *ProjectileSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	s.updateProjectileMovement(w, dt)
	s.checkProjectileCollisions(w)
	s.cleanupExpiredProjectiles(w)
}

// updateProjectileMovement moves all projectiles based on velocity.
func (s *ProjectileSystem) updateProjectileMovement(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Projectile", "Position") {
		projComp, ok := w.GetComponent(e, "Projectile")
		if !ok {
			continue
		}
		proj := projComp.(*components.Projectile)

		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)

		// Move projectile
		pos.X += proj.VelocityX * dt
		pos.Y += proj.VelocityY * dt
		pos.Z += proj.VelocityZ * dt

		// Reduce lifetime
		proj.Lifetime -= dt
	}
}

// checkProjectileCollisions checks for projectile-entity collisions.
func (s *ProjectileSystem) checkProjectileCollisions(w *ecs.World) {
	projectiles := w.Entities("Projectile", "Position")
	targets := w.Entities("Health", "Position")

	for _, pe := range projectiles {
		projComp, pOK := w.GetComponent(pe, "Projectile")
		projPosComp, ppOK := w.GetComponent(pe, "Position")
		if !pOK || !ppOK {
			continue
		}
		proj := projComp.(*components.Projectile)
		projPos := projPosComp.(*components.Position)

		for _, te := range targets {
			// Skip owner entity
			if uint64(te) == proj.OwnerID {
				continue
			}

			// Skip already hit entities (for pierce)
			if proj.HitEntities != nil && proj.HitEntities[uint64(te)] {
				continue
			}

			targetPosComp, tpOK := w.GetComponent(te, "Position")
			if !tpOK {
				continue
			}
			targetPos := targetPosComp.(*components.Position)

			// Check collision
			dx := targetPos.X - projPos.X
			dy := targetPos.Y - projPos.Y
			dz := targetPos.Z - projPos.Z
			distSq := dx*dx + dy*dy + dz*dz
			hitRadiusSq := proj.HitRadius * proj.HitRadius

			if distSq <= hitRadiusSq {
				s.applyProjectileHit(w, pe, te, proj)
			}
		}
	}
}

// applyProjectileHit applies damage and handles pierce/despawn.
func (s *ProjectileSystem) applyProjectileHit(w *ecs.World, projEntity, targetEntity ecs.Entity, proj *components.Projectile) {
	// Apply damage to target
	healthComp, ok := w.GetComponent(targetEntity, "Health")
	if ok {
		health := healthComp.(*components.Health)
		health.Current -= proj.Damage
		if health.Current < 0 {
			health.Current = 0
		}
	}

	// Track hit for pierce mechanics
	if proj.HitEntities == nil {
		proj.HitEntities = make(map[uint64]bool)
	}
	proj.HitEntities[uint64(targetEntity)] = true

	// Check if projectile should despawn
	if proj.PierceCount > 0 && len(proj.HitEntities) >= proj.PierceCount {
		proj.Lifetime = 0 // Mark for cleanup
	}
}

// cleanupExpiredProjectiles removes projectiles that have expired.
func (s *ProjectileSystem) cleanupExpiredProjectiles(w *ecs.World) {
	for _, e := range w.Entities("Projectile") {
		projComp, ok := w.GetComponent(e, "Projectile")
		if !ok {
			continue
		}
		proj := projComp.(*components.Projectile)

		if proj.Lifetime <= 0 {
			w.DestroyEntity(e)
		}
	}
}

// SpawnProjectile creates a new projectile entity.
func (s *ProjectileSystem) SpawnProjectile(w *ecs.World, owner ecs.Entity, targetX, targetY, targetZ, damage, speed float64, projectileType string) ecs.Entity {
	// Get owner position
	ownerPosComp, ok := w.GetComponent(owner, "Position")
	if !ok {
		return 0
	}
	ownerPos := ownerPosComp.(*components.Position)

	// Calculate direction
	dx := targetX - ownerPos.X
	dy := targetY - ownerPos.Y
	dz := targetZ - ownerPos.Z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)

	if dist == 0 {
		return 0 // Can't fire at self
	}

	// Normalize and apply speed
	vx := (dx / dist) * speed
	vy := (dy / dist) * speed
	vz := (dz / dist) * speed

	// Create projectile entity
	projEntity := w.CreateEntity()

	// Add position (start at owner position)
	pos := &components.Position{
		X:     ownerPos.X,
		Y:     ownerPos.Y,
		Z:     ownerPos.Z,
		Angle: math.Atan2(dy, dx),
	}
	w.AddComponent(projEntity, pos)

	// Add projectile component
	proj := &components.Projectile{
		OwnerID:        uint64(owner),
		VelocityX:      vx,
		VelocityY:      vy,
		VelocityZ:      vz,
		Damage:         damage,
		Lifetime:       DefaultProjectileLifetime,
		HitRadius:      DefaultProjectileHitRadius,
		ProjectileType: projectileType,
		PierceCount:    1, // Default: single hit
		HitEntities:    make(map[uint64]bool),
	}
	w.AddComponent(projEntity, proj)

	return projEntity
}

// InitiateRangedAttack starts a ranged attack against a position.
func (cs *CombatSystem) InitiateRangedAttack(w *ecs.World, attacker ecs.Entity, targetX, targetY, targetZ float64, projectileSystem *ProjectileSystem) bool {
	// Check cooldown
	combatComp, ok := w.GetComponent(attacker, "CombatState")
	if !ok {
		return false
	}
	combat := combatComp.(*components.CombatState)
	if combat.Cooldown > 0 {
		return false
	}

	// Get weapon
	weaponComp, wOK := w.GetComponent(attacker, "Weapon")
	if !wOK {
		return false // Need ranged weapon
	}
	weapon := weaponComp.(*components.Weapon)

	// Check if weapon is ranged
	if weapon.WeaponType != "ranged" && weapon.WeaponType != "bow" && weapon.WeaponType != "gun" && weapon.WeaponType != "crossbow" {
		return false
	}

	// Check range
	attackerPos := cs.getPosition(w, attacker)
	if attackerPos == nil {
		return false
	}
	dx := targetX - attackerPos.X
	dy := targetY - attackerPos.Y
	dz := targetZ - attackerPos.Z
	distSq := dx*dx + dy*dy + dz*dz
	if distSq > weapon.Range*weapon.Range {
		return false // Target too far
	}

	// Calculate damage with skill modifier
	damage := weapon.Damage * cs.getRangedSkillModifier(w, attacker)

	// Determine projectile speed
	speed := DefaultProjectileSpeed
	if weapon.AttackSpeed > 0 {
		speed = DefaultProjectileSpeed * weapon.AttackSpeed
	}

	// Determine projectile type based on weapon
	projType := "arrow"
	switch weapon.WeaponType {
	case "gun":
		projType = "bullet"
	case "crossbow":
		projType = "bolt"
	}

	// Spawn projectile
	if projectileSystem != nil {
		projectileSystem.SpawnProjectile(w, attacker, targetX, targetY, targetZ, damage, speed, projType)
	}

	// Set cooldown
	combat.Cooldown = cs.getAttackCooldown(w, attacker)
	combat.InCombat = true
	combat.LastAttackTime = cs.GameTime

	return true
}

// getRangedSkillModifier returns damage multiplier for ranged skills.
func (cs *CombatSystem) getRangedSkillModifier(w *ecs.World, entity ecs.Entity) float64 {
	skillsComp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return 1.0
	}
	skills := skillsComp.(*components.Skills)
	if skills.Levels == nil {
		return 1.0
	}

	// Check for ranged combat skills
	rangedLevel := 0
	for _, skillID := range []string{"archery", "ranged", "firearms", "rifles", "pistols", "marksmanship"} {
		if level, exists := skills.Levels[skillID]; exists && level > rangedLevel {
			rangedLevel = level
		}
	}

	return 1.0 + float64(rangedLevel)*SkillDamageBonus
}

// IsRangedWeapon checks if an entity has a ranged weapon equipped.
func (cs *CombatSystem) IsRangedWeapon(w *ecs.World, entity ecs.Entity) bool {
	weaponComp, ok := w.GetComponent(entity, "Weapon")
	if !ok {
		return false
	}
	weapon := weaponComp.(*components.Weapon)
	switch weapon.WeaponType {
	case "ranged", "bow", "gun", "crossbow", "rifle", "pistol":
		return true
	default:
		return weapon.Range > MeleeWeaponRangeThreshold
	}
}
