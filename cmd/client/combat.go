//go:build !noebiten

// combat.go provides client-side combat mechanics and visual feedback.
// Per PLAN.md Phase 2 Task 2C.

package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/input"
)

// CombatManager handles player combat input and visual feedback.
type CombatManager struct {
	playerEntity ecs.Entity
	combatSystem *systems.CombatSystem
	inputManager *input.Manager

	// Visual feedback state
	screenShakeMagnitude float64
	screenShakeDuration  float64
	damageFlashAlpha     float64
	damageFlashDuration  float64

	// Attack state
	isBlocking      bool
	lastAttackTime  time.Time
	attackCooldown  time.Duration
	comboCount      int
	comboResetTime  time.Time
	comboWindowSecs float64

	// Death and respawn
	isDead       bool
	deathTime    time.Time
	respawnDelay time.Duration
	respawnPos   Position3D
}

// Position3D represents a 3D position for respawn point.
type Position3D struct {
	X, Y, Z float64
}

// NewCombatManager creates a new combat manager.
func NewCombatManager(playerEntity ecs.Entity, inputManager *input.Manager) *CombatManager {
	return &CombatManager{
		playerEntity:    playerEntity,
		combatSystem:    systems.NewCombatSystem(),
		inputManager:    inputManager,
		attackCooldown:  500 * time.Millisecond,
		comboWindowSecs: 1.5,
		respawnDelay:    3 * time.Second,
		respawnPos:      Position3D{X: 8.5, Y: 8.5, Z: 0},
	}
}

// Update processes combat input and updates visual feedback.
func (cm *CombatManager) Update(world *ecs.World, dt float64) {
	// Update visual feedback timers
	cm.updateVisualFeedback(dt)

	// Check for player death
	if cm.checkPlayerDeath(world) {
		cm.handleDeath(world)
		return
	}

	// Handle death recovery
	if cm.isDead {
		cm.handleDeathRecovery(world)
		return
	}

	// Process combat input
	cm.handleCombatInput(world)

	// Update combo system
	cm.updateComboSystem()

	// Run combat system update
	cm.combatSystem.Update(world, dt)
}

// handleCombatInput processes player attack and block inputs.
func (cm *CombatManager) handleCombatInput(world *ecs.World) {
	// Check for attack input (left mouse click or action key)
	attackPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) ||
		cm.isActionPressed(input.ActionAttack)

	if attackPressed && cm.canAttack() {
		cm.performAttack(world)
	}

	// Check for block input (right mouse click or action key)
	blockPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) ||
		cm.isActionPressed(input.ActionBlock)

	cm.isBlocking = blockPressed
}

// isActionPressed checks if an input action is pressed.
func (cm *CombatManager) isActionPressed(action input.Action) bool {
	if cm.inputManager == nil {
		return false
	}
	return cm.inputManager.IsActionPressed(action)
}

// canAttack checks if the player can initiate an attack.
func (cm *CombatManager) canAttack() bool {
	if cm.isDead || cm.isBlocking {
		return false
	}
	return time.Since(cm.lastAttackTime) >= cm.attackCooldown
}

// performAttack executes a melee attack against the nearest target.
func (cm *CombatManager) performAttack(world *ecs.World) {
	// Find nearest target in attack range
	target := cm.combatSystem.FindNearestTarget(world, cm.playerEntity)

	if target != 0 {
		// Initiate attack through combat system
		if cm.combatSystem.InitiateAttack(world, cm.playerEntity, target) {
			cm.lastAttackTime = time.Now()
			cm.incrementCombo()

			// Trigger screen shake for attack feedback
			cm.triggerScreenShake(0.05, 0.1)
		}
	} else {
		// Swing attack animation even without target (miss)
		cm.lastAttackTime = time.Now()
		cm.resetCombo()
	}
}

// updateVisualFeedback updates screen shake and damage flash timers.
func (cm *CombatManager) updateVisualFeedback(dt float64) {
	// Decay screen shake
	if cm.screenShakeDuration > 0 {
		cm.screenShakeDuration -= dt
		if cm.screenShakeDuration <= 0 {
			cm.screenShakeDuration = 0
			cm.screenShakeMagnitude = 0
		}
	}

	// Decay damage flash
	if cm.damageFlashDuration > 0 {
		cm.damageFlashDuration -= dt
		if cm.damageFlashDuration <= 0 {
			cm.damageFlashDuration = 0
			cm.damageFlashAlpha = 0
		} else {
			// Fade out the flash
			cm.damageFlashAlpha = cm.damageFlashDuration / 0.3 * 128
		}
	}
}

// triggerScreenShake initiates screen shake effect.
func (cm *CombatManager) triggerScreenShake(magnitude, duration float64) {
	if magnitude > cm.screenShakeMagnitude {
		cm.screenShakeMagnitude = magnitude
	}
	if duration > cm.screenShakeDuration {
		cm.screenShakeDuration = duration
	}
}

// TriggerDamageFlash initiates red flash effect when player takes damage.
func (cm *CombatManager) TriggerDamageFlash() {
	cm.damageFlashAlpha = 128
	cm.damageFlashDuration = 0.3
	cm.triggerScreenShake(0.1, 0.15)
}

// GetScreenShakeOffset returns the current screen shake offset.
func (cm *CombatManager) GetScreenShakeOffset() (float64, float64) {
	if cm.screenShakeDuration <= 0 {
		return 0, 0
	}

	// Generate pseudo-random shake based on time
	t := float64(time.Now().UnixNano()) / 1e9
	offsetX := math.Sin(t*50) * cm.screenShakeMagnitude * 10
	offsetY := math.Cos(t*47) * cm.screenShakeMagnitude * 10

	return offsetX, offsetY
}

// GetDamageFlashAlpha returns the current damage flash alpha value.
func (cm *CombatManager) GetDamageFlashAlpha() float64 {
	return cm.damageFlashAlpha
}

// updateComboSystem manages the combo counter and timeout.
func (cm *CombatManager) updateComboSystem() {
	if cm.comboCount > 0 && time.Since(cm.comboResetTime).Seconds() > cm.comboWindowSecs {
		cm.resetCombo()
	}
}

// incrementCombo increases the combo counter.
func (cm *CombatManager) incrementCombo() {
	cm.comboCount++
	cm.comboResetTime = time.Now()

	// Combo bonuses
	if cm.comboCount >= 3 {
		// Reduce cooldown for combo attacks
		reduction := time.Duration(cm.comboCount*50) * time.Millisecond
		if reduction > 300*time.Millisecond {
			reduction = 300 * time.Millisecond
		}
		cm.attackCooldown = 500*time.Millisecond - reduction
	}
}

// resetCombo resets the combo counter.
func (cm *CombatManager) resetCombo() {
	cm.comboCount = 0
	cm.attackCooldown = 500 * time.Millisecond
}

// GetComboCount returns the current combo count.
func (cm *CombatManager) GetComboCount() int {
	return cm.comboCount
}

// checkPlayerDeath checks if the player's health has reached zero.
func (cm *CombatManager) checkPlayerDeath(world *ecs.World) bool {
	if cm.isDead {
		return false // Already dead
	}

	healthComp, ok := world.GetComponent(cm.playerEntity, "Health")
	if !ok {
		return false
	}
	health := healthComp.(*components.Health)

	return health.Current <= 0
}

// handleDeath processes player death state.
func (cm *CombatManager) handleDeath(world *ecs.World) {
	if cm.isDead {
		return
	}

	cm.isDead = true
	cm.deathTime = time.Now()

	// Trigger death visual effects
	cm.triggerScreenShake(0.3, 1.0)
	cm.damageFlashAlpha = 200
	cm.damageFlashDuration = 1.0

	// Could add death sound, camera effects, etc.
}

// handleDeathRecovery handles respawn logic after death.
func (cm *CombatManager) handleDeathRecovery(world *ecs.World) {
	if !cm.isDead {
		return
	}

	// Check if respawn delay has passed
	if time.Since(cm.deathTime) < cm.respawnDelay {
		return
	}

	// Respawn player
	cm.respawnPlayer(world)
}

// respawnPlayer resets the player's state and position.
func (cm *CombatManager) respawnPlayer(world *ecs.World) {
	// Reset health
	healthComp, ok := world.GetComponent(cm.playerEntity, "Health")
	if ok {
		health := healthComp.(*components.Health)
		health.Current = health.Max * 0.5 // Respawn with half health
	}

	// Reset mana
	manaComp, ok := world.GetComponent(cm.playerEntity, "Mana")
	if ok {
		mana := manaComp.(*components.Mana)
		mana.Current = mana.Max * 0.5 // Respawn with half mana
	}

	// Reset position to respawn point
	posComp, ok := world.GetComponent(cm.playerEntity, "Position")
	if ok {
		pos := posComp.(*components.Position)
		pos.X = cm.respawnPos.X
		pos.Y = cm.respawnPos.Y
		pos.Z = cm.respawnPos.Z
	}

	// Reset combat state
	combatComp, ok := world.GetComponent(cm.playerEntity, "CombatState")
	if ok {
		combat := combatComp.(*components.CombatState)
		combat.InCombat = false
		combat.IsAttacking = false
		combat.Cooldown = 0
	}

	cm.isDead = false
	cm.isBlocking = false
	cm.resetCombo()

	// Clear visual effects
	cm.screenShakeMagnitude = 0
	cm.screenShakeDuration = 0
	cm.damageFlashAlpha = 0
	cm.damageFlashDuration = 0
}

// SetRespawnPoint sets the player's respawn location.
func (cm *CombatManager) SetRespawnPoint(x, y, z float64) {
	cm.respawnPos = Position3D{X: x, Y: y, Z: z}
}

// IsDead returns whether the player is currently dead.
func (cm *CombatManager) IsDead() bool {
	return cm.isDead
}

// IsBlocking returns whether the player is currently blocking.
func (cm *CombatManager) IsBlocking() bool {
	return cm.isBlocking
}

// OnPlayerDamaged should be called when the player takes damage.
func (cm *CombatManager) OnPlayerDamaged(damage float64) {
	// Reduce damage if blocking
	if cm.isBlocking {
		damage *= 0.3 // Block reduces 70% damage
	}

	cm.TriggerDamageFlash()

	// Scale shake with damage
	shakeMagnitude := math.Min(damage/20.0, 0.2)
	cm.triggerScreenShake(shakeMagnitude, 0.2)
}

// GetTimeUntilRespawn returns seconds until respawn, or 0 if not dead.
func (cm *CombatManager) GetTimeUntilRespawn() float64 {
	if !cm.isDead {
		return 0
	}

	remaining := cm.respawnDelay - time.Since(cm.deathTime)
	if remaining < 0 {
		return 0
	}
	return remaining.Seconds()
}
