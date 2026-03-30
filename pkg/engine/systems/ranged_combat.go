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
		proj, projPos := s.getProjectileWithPosition(w, pe)
		if proj == nil || projPos == nil {
			continue
		}
		s.checkProjectileAgainstTargets(w, pe, proj, projPos, targets)
	}
}

// getProjectileWithPosition retrieves projectile and position components.
func (s *ProjectileSystem) getProjectileWithPosition(w *ecs.World, pe ecs.Entity) (*components.Projectile, *components.Position) {
	projComp, pOK := w.GetComponent(pe, "Projectile")
	projPosComp, ppOK := w.GetComponent(pe, "Position")
	if !pOK || !ppOK {
		return nil, nil
	}
	return projComp.(*components.Projectile), projPosComp.(*components.Position)
}

// checkProjectileAgainstTargets tests a single projectile against all targets.
func (s *ProjectileSystem) checkProjectileAgainstTargets(w *ecs.World, pe ecs.Entity, proj *components.Projectile, projPos *components.Position, targets []ecs.Entity) {
	for _, te := range targets {
		if s.shouldSkipTarget(proj, te) {
			continue
		}
		if s.isHit(w, projPos, proj.HitRadius, te) {
			s.applyProjectileHit(w, pe, te, proj)
		}
	}
}

// shouldSkipTarget returns true if target should not be considered for collision.
func (s *ProjectileSystem) shouldSkipTarget(proj *components.Projectile, te ecs.Entity) bool {
	if uint64(te) == proj.OwnerID {
		return true
	}
	if proj.HitEntities != nil && proj.HitEntities[uint64(te)] {
		return true
	}
	return false
}

// isHit checks if projectile is within hit radius of target.
func (s *ProjectileSystem) isHit(w *ecs.World, projPos *components.Position, hitRadius float64, te ecs.Entity) bool {
	targetPosComp, ok := w.GetComponent(te, "Position")
	if !ok {
		return false
	}
	targetPos := targetPosComp.(*components.Position)
	distSq := distanceSquared(targetPos, projPos)
	return distSq <= hitRadius*hitRadius
}

// distanceSquared calculates squared distance between two positions.
func distanceSquared(p1, p2 *components.Position) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	dz := p1.Z - p2.Z
	return dx*dx + dy*dy + dz*dz
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
	combat, ok := cs.getCombatStateReady(w, attacker)
	if !ok {
		return false
	}

	weapon, ok := cs.getRangedWeapon(w, attacker)
	if !ok {
		return false
	}

	attackerPos := cs.getPosition(w, attacker)
	if attackerPos == nil || !cs.isInRange(attackerPos, targetX, targetY, targetZ, weapon.Range) {
		return false
	}

	damage := weapon.Damage * cs.getRangedSkillModifier(w, attacker)
	speed := cs.getProjectileSpeed(weapon)
	projType := cs.determineProjectileType(weapon)

	if projectileSystem != nil {
		projectileSystem.SpawnProjectile(w, attacker, targetX, targetY, targetZ, damage, speed, projType)
	}

	cs.setAttackCooldown(w, attacker, combat)
	return true
}

// getCombatStateReady checks if attacker can attack (has combat state and no cooldown).
func (cs *CombatSystem) getCombatStateReady(w *ecs.World, attacker ecs.Entity) (*components.CombatState, bool) {
	combatComp, ok := w.GetComponent(attacker, "CombatState")
	if !ok {
		return nil, false
	}
	combat := combatComp.(*components.CombatState)
	if combat.Cooldown > 0 {
		return nil, false
	}
	return combat, true
}

// getRangedWeapon retrieves a ranged weapon from the attacker.
func (cs *CombatSystem) getRangedWeapon(w *ecs.World, attacker ecs.Entity) (*components.Weapon, bool) {
	weaponComp, ok := w.GetComponent(attacker, "Weapon")
	if !ok {
		return nil, false
	}
	weapon := weaponComp.(*components.Weapon)
	if !isRangedWeaponType(weapon.WeaponType) {
		return nil, false
	}
	return weapon, true
}

// isRangedWeaponType returns true if the weapon type is ranged.
func isRangedWeaponType(weaponType string) bool {
	return weaponType == "ranged" || weaponType == "bow" || weaponType == "gun" || weaponType == "crossbow"
}

// isInRange checks if target is within weapon range.
func (cs *CombatSystem) isInRange(pos *components.Position, tx, ty, tz, weaponRange float64) bool {
	dx := tx - pos.X
	dy := ty - pos.Y
	dz := tz - pos.Z
	return dx*dx+dy*dy+dz*dz <= weaponRange*weaponRange
}

// getProjectileSpeed returns projectile speed based on weapon.
func (cs *CombatSystem) getProjectileSpeed(weapon *components.Weapon) float64 {
	if weapon.AttackSpeed > 0 {
		return DefaultProjectileSpeed * weapon.AttackSpeed
	}
	return DefaultProjectileSpeed
}

// determineProjectileType returns the projectile type for a weapon.
func (cs *CombatSystem) determineProjectileType(weapon *components.Weapon) string {
	switch weapon.WeaponType {
	case "gun":
		return "bullet"
	case "crossbow":
		return "bolt"
	default:
		return "arrow"
	}
}

// setAttackCooldown sets cooldown and marks combat state.
func (cs *CombatSystem) setAttackCooldown(w *ecs.World, attacker ecs.Entity, combat *components.CombatState) {
	combat.Cooldown = cs.getAttackCooldown(w, attacker)
	combat.InCombat = true
	combat.LastAttackTime = cs.GameTime
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
