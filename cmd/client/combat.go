//go:build !noebiten

// combat.go provides client-side combat mechanics and visual feedback.
// Per PLAN.md Phase 2 Task 2C.

package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/input"
)

// CombatManager handles player combat input and visual feedback.
type CombatManager struct {
	playerEntity     ecs.Entity
	combatSystem     *systems.CombatSystem
	projectileSystem *systems.ProjectileSystem
	magicSystem      *systems.MagicSystem
	stealthSystem    *systems.StealthSystem
	inputManager     *input.Manager

	// Visual feedback state
	screenShakeMagnitude float64
	screenShakeDuration  float64
	damageFlashAlpha     float64
	damageFlashDuration  float64

	// Attack state
	isBlocking      bool
	isSneaking      bool
	lastAttackTime  time.Time
	attackCooldown  time.Duration
	comboCount      int
	comboResetTime  time.Time
	comboWindowSecs float64

	// Dodge state
	isDodging      bool
	dodgeEndTime   time.Time
	dodgeCooldown  time.Duration
	lastDodgeTime  time.Time
	dodgeDuration  time.Duration
	blockReduction float64 // Percentage of damage blocked (0.5 = 50%)

	// Death and respawn
	isDead       bool
	deathTime    time.Time
	respawnDelay time.Duration
	respawnPos   Position3D

	// Aim direction for ranged attacks (normalized)
	aimDirX, aimDirY float64

	// Magic state
	selectedSpellIndex int // Index of currently selected spell
	lastSpellCastTime  time.Time
}

// Position3D represents a 3D position for respawn point.
type Position3D struct {
	X, Y, Z float64
}

// NewCombatManager creates a new combat manager.
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
		aimDirY:            1, // Default aim direction (forward)
		selectedSpellIndex: 0,
		dodgeCooldown:      1 * time.Second,
		dodgeDuration:      300 * time.Millisecond,
		blockReduction:     0.5, // Block reduces damage by 50%
	}
}

// Update processes combat input and updates visual feedback.
func (cm *CombatManager) Update(world *ecs.World, dt float64) {
	// Update visual feedback timers
	cm.updateVisualFeedback(dt)

	// Update dodge state
	cm.updateDodgeState(world)

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

	// Run projectile system update for ranged combat
	cm.projectileSystem.Update(world, dt)

	// Run magic system update for spell effects and mana regen
	cm.magicSystem.Update(world, dt)

	// Run stealth system update for detection and alert decay
	cm.stealthSystem.Update(world, dt)
}

// handleCombatInput processes player attack and block inputs.
func (cm *CombatManager) handleCombatInput(world *ecs.World) {
	// Update aim direction based on mouse position (for ranged weapons)
	cm.updateAimDirection(world)

	// Check for sneak toggle input (C key for stealth toggle)
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		cm.toggleSneak(world)
	}

	// Check for dodge input (Space key or dodge action)
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && cm.canDodge() {
		cm.performDodge(world)
	}

	// Check for attack input (left mouse click or action key)
	attackPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) ||
		cm.isActionPressed(input.ActionAttack)

	if attackPressed && cm.canAttack() {
		cm.performAttack(world)
	}

	// Check for spell cast input (number keys or spell action)
	cm.handleSpellInput(world)

	// Check for block input (right mouse click or action key)
	blockPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) ||
		cm.isActionPressed(input.ActionBlock)

	cm.isBlocking = blockPressed
	cm.updateBlockState(world)
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
	return canAttackShared(cm.isDead, cm.isBlocking, cm.lastAttackTime, cm.attackCooldown)
}

// performAttack executes an attack based on equipped weapon type.
func (cm *CombatManager) performAttack(world *ecs.World) {
	// Check equipped weapon type
	weaponType := cm.getEquippedWeaponType(world)

	switch weaponType {
	case "ranged":
		cm.performRangedAttack(world)
	case "magic":
		// Magic combat handled by MagicSystem (placeholder)
		cm.performMeleeAttack(world)
	default:
		cm.performMeleeAttack(world)
	}
}

// getEquippedWeaponType returns the type of the player's equipped weapon.
func (cm *CombatManager) getEquippedWeaponType(world *ecs.World) string {
	return getEquippedWeaponTypeShared(world, cm.playerEntity)
}

// performMeleeAttack executes a melee attack against the nearest target.
// If sneaking and target is unaware, performs a backstab with bonus damage.
func (cm *CombatManager) performMeleeAttack(world *ecs.World) {
	target := cm.combatSystem.FindNearestTarget(world, cm.playerEntity)

	if target != 0 {
		// Check for backstab opportunity
		if cm.isSneaking && cm.stealthSystem.IsTargetUnaware(world, cm.playerEntity, target) {
			cm.performBackstab(world, target)
			return
		}

		if cm.combatSystem.InitiateAttack(world, cm.playerEntity, target) {
			cm.lastAttackTime = time.Now()
			cm.incrementCombo()
			cm.triggerScreenShake(0.05, 0.1)

			// Attacking breaks stealth
			if cm.isSneaking {
				cm.breakStealth(world)
			}
		}
	} else {
		cm.lastAttackTime = time.Now()
		cm.resetCombo()
	}
}

// performBackstab executes a stealth backstab attack with bonus damage.
func (cm *CombatManager) performBackstab(world *ecs.World, target ecs.Entity) {
	// Get base weapon damage
	baseDamage := cm.getWeaponDamage(world)

	// Apply backstab multiplier from stealth system
	backstabDamage := cm.stealthSystem.GetBackstabDamage(world, baseDamage, cm.playerEntity, target)

	// Apply damage directly to target
	healthComp, exists := world.GetComponent(target, "Health")
	if exists && healthComp != nil {
		health, ok := healthComp.(*components.Health)
		if ok {
			health.Current -= backstabDamage
		}
	}

	cm.lastAttackTime = time.Now()
	cm.triggerScreenShake(0.1, 0.15) // Stronger feedback for backstab

	// Backstab breaks stealth
	cm.breakStealth(world)
}

// getWeaponDamage returns the equipped weapon's damage or default.
func (cm *CombatManager) getWeaponDamage(world *ecs.World) float64 {
	return getWeaponDamageShared(world, cm.playerEntity)
}

// toggleSneak toggles the player's sneaking state.
func (cm *CombatManager) toggleSneak(world *ecs.World) {
	if cm.isDead {
		return
	}
	cm.isSneaking = !cm.isSneaking
	cm.stealthSystem.SetSneaking(world, cm.playerEntity, cm.isSneaking)
}

// breakStealth forces the player out of stealth mode.
func (cm *CombatManager) breakStealth(world *ecs.World) {
	cm.isSneaking = false
	cm.stealthSystem.SetSneaking(world, cm.playerEntity, false)
}

// IsSneaking returns whether the player is currently sneaking.
func (cm *CombatManager) IsSneaking() bool {
	return cm.isSneaking
}

// canDodge checks if the player can perform a dodge roll.
func (cm *CombatManager) canDodge() bool {
	return canDodgeShared(cm.isDead, cm.isDodging, cm.isBlocking, cm.lastDodgeTime, cm.dodgeCooldown)
}

// performDodge initiates a dodge roll with invulnerability frames.
func (cm *CombatManager) performDodge(world *ecs.World) {
	cm.isDodging = true
	cm.lastDodgeTime = time.Now()
	cm.dodgeEndTime = time.Now().Add(cm.dodgeDuration)

	// Update CombatState component
	cm.updateCombatStateDodge(world, true)

	// Dodging breaks stealth
	if cm.isSneaking {
		cm.breakStealth(world)
	}
}

// updateDodgeState checks if dodge has ended.
func (cm *CombatManager) updateDodgeState(world *ecs.World) {
	if cm.isDodging && time.Now().After(cm.dodgeEndTime) {
		cm.isDodging = false
		cm.updateCombatStateDodge(world, false)
	}
}

// updateCombatStateDodge updates the CombatState component's dodge fields.
func (cm *CombatManager) updateCombatStateDodge(world *ecs.World, dodging bool) {
	combatComp, exists := world.GetComponent(cm.playerEntity, "CombatState")
	if !exists || combatComp == nil {
		return
	}
	combatState, ok := combatComp.(*components.CombatState)
	if !ok {
		return
	}
	combatState.IsDodging = dodging
	combatState.DodgeInvulnerable = dodging
}

// updateBlockState updates the CombatState component's block fields.
func (cm *CombatManager) updateBlockState(world *ecs.World) {
	combatComp, exists := world.GetComponent(cm.playerEntity, "CombatState")
	if !exists || combatComp == nil {
		return
	}
	combatState, ok := combatComp.(*components.CombatState)
	if !ok {
		return
	}
	combatState.IsBlocking = cm.isBlocking
	combatState.BlockReduction = cm.blockReduction
}

// IsDodging returns whether the player is currently dodging.
func (cm *CombatManager) IsDodging() bool {
	return cm.isDodging
}

// GetBlockReduction returns the current block damage reduction.
func (cm *CombatManager) GetBlockReduction() float64 {
	if cm.isBlocking {
		return cm.blockReduction
	}
	return 0
}

// CalculateIncomingDamage adjusts damage based on block/dodge state.
func (cm *CombatManager) CalculateIncomingDamage(baseDamage float64) float64 {
	return calculateIncomingDamageShared(baseDamage, cm.isDodging, cm.isBlocking, cm.blockReduction)
}

// performRangedAttack fires a projectile in the aim direction.
func (cm *CombatManager) performRangedAttack(world *ecs.World) {
	// Get player position
	posComp, exists := world.GetComponent(cm.playerEntity, "Position")
	if !exists || posComp == nil {
		return
	}
	pos, ok := posComp.(*components.Position)
	if !ok {
		return
	}

	// Get weapon damage and speed
	damage, speed, projRange := cm.getRangedWeaponStats(world)

	// Calculate target position based on aim direction and range
	targetX := pos.X + cm.aimDirX*projRange
	targetY := pos.Y + cm.aimDirY*projRange
	targetZ := pos.Z

	// Spawn projectile using ProjectileSystem
	cm.projectileSystem.SpawnProjectile(
		world, cm.playerEntity,
		targetX, targetY, targetZ,
		damage, speed, "arrow", // projectileType varies by weapon
	)

	cm.lastAttackTime = time.Now()
	cm.triggerScreenShake(0.02, 0.05) // Lighter shake for ranged
}

// getRangedWeaponStats returns damage, speed, and range for the equipped ranged weapon.
func (cm *CombatManager) getRangedWeaponStats(world *ecs.World) (damage, speed, weaponRange float64) {
	return getRangedWeaponStatsShared(world, cm.playerEntity)
}

// updateAimDirection calculates aim direction from player position to screen center.
// In first-person view, aim direction follows camera facing direction.
func (cm *CombatManager) updateAimDirection(world *ecs.World) {
	// Get player position and facing direction
	posComp, exists := world.GetComponent(cm.playerEntity, "Position")
	if !exists || posComp == nil {
		return
	}
	pos, ok := posComp.(*components.Position)
	if !ok {
		return
	}

	// Use player's Angle (from Position component) for aim direction
	// Angle is typically stored in radians
	cm.aimDirX = math.Cos(pos.Angle)
	cm.aimDirY = math.Sin(pos.Angle)
}

// handleSpellInput processes spell casting and selection input.
func (cm *CombatManager) handleSpellInput(world *ecs.World) {
	// Spell selection with number keys 1-9
	for i := 0; i < 9; i++ {
		key := ebiten.Key(int(ebiten.Key1) + i)
		if inpututil.IsKeyJustPressed(key) {
			cm.selectedSpellIndex = i
		}
	}

	// Cast spell with Q key or ActionCastSpell
	castPressed := inpututil.IsKeyJustPressed(ebiten.KeyQ) ||
		cm.isActionPressed(input.ActionCastSpell)

	if castPressed && cm.canCastSpell() {
		cm.performMagicAttack(world)
	}
}

// canCastSpell checks if player can cast a spell (not dead, has mana).
func (cm *CombatManager) canCastSpell() bool {
	return canCastSpellShared(cm.isDead, cm.isBlocking, cm.lastSpellCastTime)
}

// performMagicAttack casts the selected spell toward the aim direction.
func (cm *CombatManager) performMagicAttack(world *ecs.World) {
	// Get player's spell book
	spellID := cm.getSelectedSpellID(world)
	if spellID == "" {
		return
	}

	// Get player position for target calculation
	posComp, exists := world.GetComponent(cm.playerEntity, "Position")
	if !exists || posComp == nil {
		return
	}
	pos, ok := posComp.(*components.Position)
	if !ok {
		return
	}

	// Calculate target position based on aim direction
	spellRange := 15.0 // Default spell range
	targetX := pos.X + cm.aimDirX*spellRange
	targetY := pos.Y + cm.aimDirY*spellRange
	targetZ := pos.Z

	// Cast spell at target position
	if cm.magicSystem.CastSpellAtPosition(world, cm.playerEntity, spellID, targetX, targetY, targetZ, cm.projectileSystem) {
		cm.lastSpellCastTime = time.Now()
		cm.triggerScreenShake(0.03, 0.08) // Magic effect shake
	}
}

// getSelectedSpellID returns the spell ID at the selected index from the player's spellbook.
func (cm *CombatManager) getSelectedSpellID(world *ecs.World) string {
	return getSelectedSpellIDShared(world, cm.playerEntity, cm.selectedSpellIndex)
}

// GetSelectedSpellIndex returns the currently selected spell slot (0-8).
func (cm *CombatManager) GetSelectedSpellIndex() int {
	return cm.selectedSpellIndex
}

// SetSelectedSpellIndex sets the currently selected spell slot.
func (cm *CombatManager) SetSelectedSpellIndex(index int) {
	if index >= 0 && index < 9 {
		cm.selectedSpellIndex = index
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
