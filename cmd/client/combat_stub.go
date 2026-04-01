//go:build noebiten

// Package main provides stub types for noebiten builds.
package main

import (
	"math"
	"time"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/input"
)

// CombatManager handles player combat input and visual feedback (stub for noebiten).
type CombatManager struct {
	playerEntity     ecs.Entity
	combatSystem     *systems.CombatSystem
	projectileSystem *systems.ProjectileSystem
	inputManager     *input.Manager

	screenShakeMagnitude float64
	screenShakeDuration  float64
	damageFlashAlpha     float64
	damageFlashDuration  float64
	isBlocking           bool
	lastAttackTime       time.Time
	attackCooldown       time.Duration
	comboCount           int
	comboResetTime       time.Time
	comboWindowSecs      float64
	isDead               bool
	deathTime            time.Time
	respawnDelay         time.Duration
	respawnPos           Position3D
	aimDirX, aimDirY     float64
}

// Position3D represents a 3D position for respawn point.
type Position3D struct {
	X, Y, Z float64
}

// NewCombatManager creates a new combat manager (stub for noebiten).
func NewCombatManager(playerEntity ecs.Entity, inputManager *input.Manager) *CombatManager {
	return &CombatManager{
		playerEntity:     playerEntity,
		combatSystem:     systems.NewCombatSystem(),
		projectileSystem: systems.NewProjectileSystem(),
		inputManager:     inputManager,
		attackCooldown:   500 * time.Millisecond,
		comboWindowSecs:  1.5,
		respawnDelay:     3 * time.Second,
		respawnPos:       Position3D{X: 8.5, Y: 8.5, Z: 0},
		aimDirX:          0,
		aimDirY:          1,
	}
}

// Update processes combat input and updates visual feedback (stub).
func (cm *CombatManager) Update(world *ecs.World, dt float64) {}

// getEquippedWeaponType returns the type of the player's equipped weapon.
func (cm *CombatManager) getEquippedWeaponType(world *ecs.World) string {
	weaponComp, exists := world.GetComponent(cm.playerEntity, "Weapon")
	if !exists || weaponComp == nil {
		return "melee"
	}
	weapon, ok := weaponComp.(*components.Weapon)
	if !ok || weapon.WeaponType == "" {
		return "melee"
	}
	return weapon.WeaponType
}

// getRangedWeaponStats returns damage, speed, and range for the equipped ranged weapon.
func (cm *CombatManager) getRangedWeaponStats(world *ecs.World) (damage, speed, weaponRange float64) {
	weaponComp, exists := world.GetComponent(cm.playerEntity, "Weapon")
	if !exists || weaponComp == nil {
		return 10.0, 15.0, 20.0
	}
	weapon, ok := weaponComp.(*components.Weapon)
	if !ok {
		return 10.0, 15.0, 20.0
	}
	damage = weapon.Damage
	if damage <= 0 {
		damage = 10.0
	}
	speed = 15.0
	weaponRange = weapon.Range
	if weaponRange <= 0 {
		weaponRange = 20.0
	}
	return damage, speed, weaponRange
}

// canAttack checks if the player can initiate an attack.
func (cm *CombatManager) canAttack() bool {
	if cm.isDead || cm.isBlocking {
		return false
	}
	return time.Since(cm.lastAttackTime) >= cm.attackCooldown
}

// updateAimDirection calculates aim direction from player facing angle.
func (cm *CombatManager) updateAimDirection(world *ecs.World) {
	posComp, exists := world.GetComponent(cm.playerEntity, "Position")
	if !exists || posComp == nil {
		return
	}
	pos, ok := posComp.(*components.Position)
	if !ok {
		return
	}
	cm.aimDirX = math.Cos(pos.Angle)
	cm.aimDirY = math.Sin(pos.Angle)
}

// GetScreenShakeOffset returns the current screen shake offset (stub).
func (cm *CombatManager) GetScreenShakeOffset() (float64, float64) { return 0, 0 }

// GetDamageFlashAlpha returns the current damage flash alpha (stub).
func (cm *CombatManager) GetDamageFlashAlpha() float64 { return 0 }

// GetComboCount returns the current combo count (stub).
func (cm *CombatManager) GetComboCount() int { return 0 }

// IsDead returns whether the player is dead (stub).
func (cm *CombatManager) IsDead() bool { return cm.isDead }

// IsBlocking returns whether the player is blocking (stub).
func (cm *CombatManager) IsBlocking() bool { return cm.isBlocking }

// SetRespawnPoint sets the respawn position (stub).
func (cm *CombatManager) SetRespawnPoint(x, y, z float64) {
	cm.respawnPos = Position3D{X: x, Y: y, Z: z}
}

// OnPlayerDamaged handles player damage visual feedback (stub).
func (cm *CombatManager) OnPlayerDamaged(damage float64) {}

// TriggerDamageFlash triggers the damage flash effect (stub).
func (cm *CombatManager) TriggerDamageFlash() {}

// GetTimeUntilRespawn returns time until respawn (stub).
func (cm *CombatManager) GetTimeUntilRespawn() float64 { return 0 }
