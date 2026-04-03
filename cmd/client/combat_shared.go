// combat_shared.go provides shared combat logic used by both
// ebiten and noebiten builds.
package main

import (
	"math"
	"time"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/input"
)

// combatManagerBase contains the common fields and methods for CombatManager.
// This struct is embedded in the platform-specific CombatManager implementations.
type combatManagerBase struct {
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
	blockReduction float64

	// Death and respawn
	isDead       bool
	deathTime    time.Time
	respawnDelay time.Duration
	respawnPos   position3D

	// Aim direction for ranged attacks (normalized)
	aimDirX, aimDirY float64

	// Magic state
	selectedSpellIndex int
	lastSpellCastTime  time.Time
}

// position3D represents a 3D position for respawn point.
type position3D struct {
	X, Y, Z float64
}

// newCombatManagerBase creates a new combatManagerBase with default values.
func newCombatManagerBase(playerEntity ecs.Entity, inputManager *input.Manager) combatManagerBase {
	return combatManagerBase{
		playerEntity:       playerEntity,
		combatSystem:       systems.NewCombatSystem(),
		projectileSystem:   systems.NewProjectileSystem(),
		magicSystem:        systems.NewMagicSystem(),
		stealthSystem:      systems.NewStealthSystem(),
		inputManager:       inputManager,
		attackCooldown:     500 * time.Millisecond,
		comboWindowSecs:    1.5,
		respawnDelay:       3 * time.Second,
		respawnPos:         position3D{X: 8.5, Y: 8.5, Z: 0},
		aimDirX:            0,
		aimDirY:            1,
		selectedSpellIndex: 0,
		dodgeCooldown:      1 * time.Second,
		dodgeDuration:      300 * time.Millisecond,
		blockReduction:     0.5,
	}
}

// getEquippedWeaponTypeShared returns the type of weapon currently equipped.
func getEquippedWeaponTypeShared(world *ecs.World, playerEntity ecs.Entity) string {
	weaponComp, exists := world.GetComponent(playerEntity, "Weapon")
	if !exists || weaponComp == nil {
		return "melee"
	}
	weapon, ok := weaponComp.(*components.Weapon)
	if !ok || weapon.WeaponType == "" {
		return "melee"
	}
	return weapon.WeaponType
}

// getRangedWeaponStatsShared returns damage, speed, and range for the equipped ranged weapon.
func getRangedWeaponStatsShared(world *ecs.World, playerEntity ecs.Entity) (damage, speed, weaponRange float64) {
	weaponComp, exists := world.GetComponent(playerEntity, "Weapon")
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

// getWeaponDamageShared returns the equipped weapon's damage.
func getWeaponDamageShared(world *ecs.World, playerEntity ecs.Entity) float64 {
	weaponComp, exists := world.GetComponent(playerEntity, "Weapon")
	if !exists || weaponComp == nil {
		return 10.0
	}
	weapon, ok := weaponComp.(*components.Weapon)
	if !ok || weapon.Damage <= 0 {
		return 10.0
	}
	return weapon.Damage
}

// getSelectedSpellIDShared returns the spell ID at the selected index from the player's spellbook.
func getSelectedSpellIDShared(world *ecs.World, playerEntity ecs.Entity, selectedSpellIndex int) string {
	spellBookComp, exists := world.GetComponent(playerEntity, "Spellbook")
	if !exists || spellBookComp == nil {
		return ""
	}
	spellBook, ok := spellBookComp.(*components.Spellbook)
	if !ok || len(spellBook.Spells) == 0 {
		return ""
	}

	// If there's an active spell already selected, use that
	if spellBook.ActiveSpellID != "" {
		return spellBook.ActiveSpellID
	}

	// Convert map to slice for indexed access
	spellIDs := make([]string, 0, len(spellBook.Spells))
	for id := range spellBook.Spells {
		spellIDs = append(spellIDs, id)
	}

	// Clamp selected index to valid range
	idx := selectedSpellIndex
	if idx < 0 {
		idx = 0
	}
	if idx >= len(spellIDs) {
		idx = len(spellIDs) - 1
	}
	return spellIDs[idx]
}

// canAttackShared checks if the player can initiate an attack.
func canAttackShared(isDead, isBlocking bool, lastAttackTime time.Time, attackCooldown time.Duration) bool {
	if isDead || isBlocking {
		return false
	}
	return time.Since(lastAttackTime) >= attackCooldown
}

// canCastSpellShared checks if the player can cast a spell.
func canCastSpellShared(isDead, isBlocking bool, lastSpellCastTime time.Time) bool {
	if isDead || isBlocking {
		return false
	}
	return time.Since(lastSpellCastTime) >= 200*time.Millisecond
}

// canDodgeShared checks if the player can perform a dodge roll.
func canDodgeShared(isDead, isDodging, isBlocking bool, lastDodgeTime time.Time, dodgeCooldown time.Duration) bool {
	if isDead || isDodging || isBlocking {
		return false
	}
	return time.Since(lastDodgeTime) >= dodgeCooldown
}

// calculateIncomingDamageShared adjusts damage based on block/dodge state.
func calculateIncomingDamageShared(baseDamage float64, isDodging, isBlocking bool, blockReduction float64) float64 {
	if isDodging {
		return 0
	}
	if isBlocking {
		return baseDamage * (1.0 - blockReduction)
	}
	return baseDamage
}

// updateAimDirectionShared calculates aim direction from player position and facing.
func updateAimDirectionShared(world *ecs.World, playerEntity ecs.Entity) (aimDirX, aimDirY float64) {
	posComp, exists := world.GetComponent(playerEntity, "Position")
	if !exists || posComp == nil {
		return 0, 1
	}
	pos, ok := posComp.(*components.Position)
	if !ok {
		return 0, 1
	}
	// Use Angle field for aim direction (in radians)
	return math.Cos(pos.Angle), math.Sin(pos.Angle)
}
