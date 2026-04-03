//go:build noebiten

// Package main provides stub types for noebiten builds.
package main

import (
	"time"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/input"
)

// CombatManager handles player combat input and visual feedback (stub for noebiten).
type CombatManager struct {
	playerEntity     ecs.Entity
	combatSystem     *systems.CombatSystem
	projectileSystem *systems.ProjectileSystem
	magicSystem      *systems.MagicSystem
	stealthSystem    *systems.StealthSystem
	inputManager     *input.Manager

	screenShakeMagnitude float64
	screenShakeDuration  float64
	damageFlashAlpha     float64
	damageFlashDuration  float64
	isBlocking           bool
	isSneaking           bool
	isDodging            bool
	dodgeEndTime         time.Time
	dodgeCooldown        time.Duration
	lastDodgeTime        time.Time
	dodgeDuration        time.Duration
	blockReduction       float64
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
	selectedSpellIndex   int
	lastSpellCastTime    time.Time
}

// Position3D represents a 3D position for respawn point.
type Position3D struct {
	X, Y, Z float64
}

// NewCombatManager creates a new combat manager (stub for noebiten).
func NewCombatManager(playerEntity ecs.Entity, inputManager *input.Manager) *CombatManager {
	return &CombatManager{
		playerEntity:       playerEntity,
		combatSystem:       systems.NewCombatSystem(),
		projectileSystem:   systems.NewProjectileSystem(),
		magicSystem:        systems.NewMagicSystem(),
		stealthSystem:      systems.NewStealthSystem(),
		inputManager:       inputManager,
		attackCooldown:     500 * time.Millisecond,
		comboWindowSecs:    1.5,
		respawnDelay:       3 * time.Second,
		respawnPos:         Position3D{X: 8.5, Y: 8.5, Z: 0},
		aimDirX:            0,
		aimDirY:            1,
		selectedSpellIndex: 0,
		dodgeCooldown:      1 * time.Second,
		dodgeDuration:      300 * time.Millisecond,
		blockReduction:     0.5,
	}
}

// Update processes combat input and updates visual feedback (stub).
func (cm *CombatManager) Update(world *ecs.World, dt float64) {}

// getEquippedWeaponType returns the type of the player's equipped weapon.
func (cm *CombatManager) getEquippedWeaponType(world *ecs.World) string {
	return getEquippedWeaponTypeShared(world, cm.playerEntity)
}

// getRangedWeaponStats returns damage, speed, and range for the equipped ranged weapon.
func (cm *CombatManager) getRangedWeaponStats(world *ecs.World) (damage, speed, weaponRange float64) {
	return getRangedWeaponStatsShared(world, cm.playerEntity)
}

// canAttack checks if the player can initiate an attack.
func (cm *CombatManager) canAttack() bool {
	return canAttackShared(cm.isDead, cm.isBlocking, cm.lastAttackTime, cm.attackCooldown)
}

// updateAimDirection calculates aim direction from player facing angle.
func (cm *CombatManager) updateAimDirection(world *ecs.World) {
	cm.aimDirX, cm.aimDirY = updateAimDirectionShared(world, cm.playerEntity)
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

// GetSelectedSpellIndex returns the currently selected spell slot (stub).
func (cm *CombatManager) GetSelectedSpellIndex() int { return cm.selectedSpellIndex }

// SetSelectedSpellIndex sets the currently selected spell slot (stub).
func (cm *CombatManager) SetSelectedSpellIndex(index int) {
	if index >= 0 && index < 9 {
		cm.selectedSpellIndex = index
	}
}

// getSelectedSpellID returns the spell ID at the selected index (stub).
func (cm *CombatManager) getSelectedSpellID(world *ecs.World) string {
	return getSelectedSpellIDShared(world, cm.playerEntity, cm.selectedSpellIndex)
}

// canCastSpell checks if player can cast a spell (stub).
func (cm *CombatManager) canCastSpell() bool {
	return canCastSpellShared(cm.isDead, cm.isBlocking, cm.lastSpellCastTime)
}

// IsSneaking returns whether the player is sneaking (stub).
func (cm *CombatManager) IsSneaking() bool { return cm.isSneaking }

// toggleSneak toggles the player's sneaking state (stub).
func (cm *CombatManager) toggleSneak(world *ecs.World) {
	if cm.isDead {
		return
	}
	cm.isSneaking = !cm.isSneaking
	cm.stealthSystem.SetSneaking(world, cm.playerEntity, cm.isSneaking)
}

// breakStealth forces the player out of stealth mode (stub).
func (cm *CombatManager) breakStealth(world *ecs.World) {
	cm.isSneaking = false
	cm.stealthSystem.SetSneaking(world, cm.playerEntity, false)
}

// getWeaponDamage returns the equipped weapon's damage (stub).
func (cm *CombatManager) getWeaponDamage(world *ecs.World) float64 {
	return getWeaponDamageShared(world, cm.playerEntity)
}

// canDodge checks if the player can perform a dodge roll (stub).
func (cm *CombatManager) canDodge() bool {
	return canDodgeShared(cm.isDead, cm.isDodging, cm.isBlocking, cm.lastDodgeTime, cm.dodgeCooldown)
}

// IsDodging returns whether the player is currently dodging (stub).
func (cm *CombatManager) IsDodging() bool { return cm.isDodging }

// GetBlockReduction returns the current block damage reduction (stub).
func (cm *CombatManager) GetBlockReduction() float64 {
	if cm.isBlocking {
		return cm.blockReduction
	}
	return 0
}

// CalculateIncomingDamage adjusts damage based on block/dodge state (stub).
func (cm *CombatManager) CalculateIncomingDamage(baseDamage float64) float64 {
	return calculateIncomingDamageShared(baseDamage, cm.isDodging, cm.isBlocking, cm.blockReduction)
}
