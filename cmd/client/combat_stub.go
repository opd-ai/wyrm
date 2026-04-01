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
	spellBookComp, exists := world.GetComponent(cm.playerEntity, "Spellbook")
	if !exists || spellBookComp == nil {
		return ""
	}
	spellBook, ok := spellBookComp.(*components.Spellbook)
	if !ok || len(spellBook.Spells) == 0 {
		return ""
	}
	if spellBook.ActiveSpellID != "" {
		return spellBook.ActiveSpellID
	}
	spellIDs := make([]string, 0, len(spellBook.Spells))
	for id := range spellBook.Spells {
		spellIDs = append(spellIDs, id)
	}
	idx := cm.selectedSpellIndex
	if idx < 0 {
		idx = 0
	}
	if idx >= len(spellIDs) {
		idx = len(spellIDs) - 1
	}
	return spellIDs[idx]
}

// canCastSpell checks if player can cast a spell (stub).
func (cm *CombatManager) canCastSpell() bool {
	if cm.isDead || cm.isBlocking {
		return false
	}
	return time.Since(cm.lastSpellCastTime) >= 200*time.Millisecond
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
	weaponComp, exists := world.GetComponent(cm.playerEntity, "Weapon")
	if !exists || weaponComp == nil {
		return 10.0
	}
	weapon, ok := weaponComp.(*components.Weapon)
	if !ok || weapon.Damage <= 0 {
		return 10.0
	}
	return weapon.Damage
}

// canDodge checks if the player can perform a dodge roll (stub).
func (cm *CombatManager) canDodge() bool {
	if cm.isDead || cm.isDodging || cm.isBlocking {
		return false
	}
	return time.Since(cm.lastDodgeTime) >= cm.dodgeCooldown
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
	if cm.isDodging {
		return 0
	}
	if cm.isBlocking {
		return baseDamage * (1.0 - cm.blockReduction)
	}
	return baseDamage
}
