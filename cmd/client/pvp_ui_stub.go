//go:build noebiten

// Stub for PvP UI when building without Ebiten.
package main

import "github.com/opd-ai/wyrm/pkg/world/pvp"

// PvPUI stub for testing.
type PvPUI struct{}

// NewPvPUI stub.
func NewPvPUI() *PvPUI { return &PvPUI{} }

// Update stub.
func (ui *PvPUI) Update(playerEntity uint64, playerX, playerZ float64) {}

// CheckCombat stub.
func (ui *PvPUI) CheckCombat(
	attackerID uint64, attackerX, attackerZ float64,
	defenderID uint64, defenderX, defenderZ float64,
	baseDamage float64,
) *pvp.CombatResult {
	return &pvp.CombatResult{DamageAllowed: false}
}

// ProcessDeath stub.
func (ui *PvPUI) ProcessDeath(entityID uint64, inventory []string, x, z float64) *pvp.DeathLoot {
	return nil
}

// GetRespawnPoint stub.
func (ui *PvPUI) GetRespawnPoint(x, z float64) (respawnX, respawnZ float64) {
	return 0, 0
}

// GetZoneManager stub.
func (ui *PvPUI) GetZoneManager() *pvp.ZoneManager {
	return pvp.NewZoneManager()
}

// IsInSafeZone stub.
func (ui *PvPUI) IsInSafeZone(x, z float64) bool { return true }

// IsInHostileZone stub.
func (ui *PvPUI) IsInHostileZone(x, z float64) bool { return false }

// SetPlayerFlag stub.
func (ui *PvPUI) SetPlayerFlag(entityID uint64, flagged bool) {}

// IsPlayerFlagged stub.
func (ui *PvPUI) IsPlayerFlagged(entityID uint64) bool { return false }

// getPlayerInventoryStrings stub - returns nil for test builds.
func getPlayerInventoryStrings(_ interface{}) []string { return nil }
