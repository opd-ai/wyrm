// Package pvp provides PvP zone management and combat rules.
// Per ROADMAP Phase 5 item 21:
// AC: Player flagged PvP in zone; killed player drops ≥1 inventory item;
// unflagged player takes no damage from flagged.
package pvp

import (
	"math/rand"
	"sync"
)

// ZoneType indicates the PvP rules for an area.
type ZoneType int

const (
	// ZoneSafe is a non-PvP zone (cities, sanctuaries).
	ZoneSafe ZoneType = iota
	// ZoneContested allows opt-in PvP.
	ZoneContested
	// ZoneHostile is full PvP (wilderness, faction territory borders).
	ZoneHostile
)

// Zone represents a PvP zone with boundaries and rules.
type Zone struct {
	ID           string
	Type         ZoneType
	MinX, MinZ   float64
	MaxX, MaxZ   float64
	RespawnX     float64
	RespawnZ     float64
	LootDropRate float64 // Percentage of inventory to drop on death (0.0-1.0)
}

// ContainsPoint checks if a point is within the zone boundaries.
func (z *Zone) ContainsPoint(x, zPos float64) bool {
	return x >= z.MinX && x <= z.MaxX && zPos >= z.MinZ && zPos <= z.MaxZ
}

// ZoneManager tracks PvP zones and player flags.
type ZoneManager struct {
	mu           sync.RWMutex
	zones        map[string]*Zone
	playerFlags  map[uint64]bool // EntityID -> PvP flagged
	flagCooldown map[uint64]int64 // EntityID -> Unix timestamp when flag expires
}

// NewZoneManager creates a new PvP zone manager.
func NewZoneManager() *ZoneManager {
	return &ZoneManager{
		zones:        make(map[string]*Zone),
		playerFlags:  make(map[uint64]bool),
		flagCooldown: make(map[uint64]int64),
	}
}

// AddZone registers a PvP zone.
func (m *ZoneManager) AddZone(zone *Zone) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.zones[zone.ID] = zone
}

// RemoveZone unregisters a PvP zone.
func (m *ZoneManager) RemoveZone(zoneID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.zones, zoneID)
}

// GetZone returns a zone by ID.
func (m *ZoneManager) GetZone(zoneID string) *Zone {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.zones[zoneID]
}

// GetZoneAt returns the zone containing the given position, or nil if none.
func (m *ZoneManager) GetZoneAt(x, z float64) *Zone {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, zone := range m.zones {
		if zone.ContainsPoint(x, z) {
			return zone
		}
	}
	return nil
}

// GetZoneTypeAt returns the zone type at a position.
// Returns ZoneSafe if no zone found (default safety).
func (m *ZoneManager) GetZoneTypeAt(x, z float64) ZoneType {
	zone := m.GetZoneAt(x, z)
	if zone == nil {
		return ZoneSafe
	}
	return zone.Type
}

// SetPlayerFlag sets or clears a player's PvP flag.
func (m *ZoneManager) SetPlayerFlag(entityID uint64, flagged bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playerFlags[entityID] = flagged
}

// IsPlayerFlagged returns whether a player is flagged for PvP.
func (m *ZoneManager) IsPlayerFlagged(entityID uint64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.playerFlags[entityID]
}

// SetFlagCooldown sets when a player's PvP flag will expire.
func (m *ZoneManager) SetFlagCooldown(entityID uint64, expiresUnix int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flagCooldown[entityID] = expiresUnix
}

// GetFlagCooldown returns when a player's PvP flag expires.
func (m *ZoneManager) GetFlagCooldown(entityID uint64) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.flagCooldown[entityID]
}

// ClearExpiredFlags removes flags that have expired.
func (m *ZoneManager) ClearExpiredFlags(currentUnix int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for entityID, expires := range m.flagCooldown {
		if currentUnix >= expires {
			m.playerFlags[entityID] = false
			delete(m.flagCooldown, entityID)
		}
	}
}

// CombatResult contains the outcome of a PvP damage check.
type CombatResult struct {
	DamageAllowed bool
	DamageAmount  float64
	AttackerFlagged bool
	DefenderFlagged bool
	Zone         *Zone
}

// CheckCombat determines if damage should be applied between two players.
// Returns the result including damage allowance and zone context.
func (m *ZoneManager) CheckCombat(
	attackerID uint64, attackerX, attackerZ float64,
	defenderID uint64, defenderX, defenderZ float64,
	baseDamage float64,
) *CombatResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attackerFlagged := m.playerFlags[attackerID]
	defenderFlagged := m.playerFlags[defenderID]

	// Get zones for both players
	attackerZone := m.getZoneAtLocked(attackerX, attackerZ)
	defenderZone := m.getZoneAtLocked(defenderX, defenderZ)

	result := &CombatResult{
		DamageAllowed:   false,
		DamageAmount:    0,
		AttackerFlagged: attackerFlagged,
		DefenderFlagged: defenderFlagged,
		Zone:            defenderZone,
	}

	// Determine zone type for combat (use defender's zone)
	zoneType := ZoneSafe
	if defenderZone != nil {
		zoneType = defenderZone.Type
	}

	// Per AC: unflagged player takes no damage from flagged
	switch zoneType {
	case ZoneSafe:
		// No PvP damage in safe zones
		result.DamageAllowed = false

	case ZoneContested:
		// Only allow damage if both players are flagged
		if attackerFlagged && defenderFlagged {
			result.DamageAllowed = true
			result.DamageAmount = baseDamage
		}

	case ZoneHostile:
		// Always allow damage in hostile zones
		result.DamageAllowed = true
		result.DamageAmount = baseDamage
	}

	// Additional check: never damage unflagged from flagged in any zone
	if attackerFlagged && !defenderFlagged && zoneType != ZoneHostile {
		result.DamageAllowed = false
		result.DamageAmount = 0
	}

	_ = attackerZone // May be used for future cross-zone rules

	return result
}

// getZoneAtLocked returns zone at position (caller must hold lock).
func (m *ZoneManager) getZoneAtLocked(x, z float64) *Zone {
	for _, zone := range m.zones {
		if zone.ContainsPoint(x, z) {
			return zone
		}
	}
	return nil
}

// DeathLoot represents items dropped on PvP death.
type DeathLoot struct {
	DropperEntityID uint64
	Items           []string
	X, Z            float64
}

// CalculateDeathLoot determines what items a player drops on death.
// Per AC: killed player drops ≥1 inventory item.
func (m *ZoneManager) CalculateDeathLoot(
	entityID uint64,
	inventory []string,
	x, z float64,
	rng *rand.Rand,
) *DeathLoot {
	zone := m.GetZoneAt(x, z)
	if !m.shouldDropLoot(zone, inventory) {
		return nil
	}

	numItems := m.calculateDropCount(zone, len(inventory))
	droppedItems := selectRandomItems(inventory, numItems, rng)

	return &DeathLoot{
		DropperEntityID: entityID,
		Items:           droppedItems,
		X:               x,
		Z:               z,
	}
}

// shouldDropLoot determines if loot should be dropped for a zone.
func (m *ZoneManager) shouldDropLoot(zone *Zone, inventory []string) bool {
	if zone == nil {
		return false
	}
	if zone.Type == ZoneSafe {
		return false
	}
	return len(inventory) > 0
}

// calculateDropCount calculates how many items to drop based on zone drop rate.
func (m *ZoneManager) calculateDropCount(zone *Zone, inventorySize int) int {
	dropRate := zone.LootDropRate
	if dropRate <= 0 {
		dropRate = 0.1 // Default 10% drop rate
	}

	numItems := int(float64(inventorySize) * dropRate)
	if numItems < 1 {
		numItems = 1 // Per AC: drops ≥1 item
	}
	if numItems > inventorySize {
		numItems = inventorySize
	}
	return numItems
}

// selectRandomItems selects n random items from the inventory.
func selectRandomItems(inventory []string, n int, rng *rand.Rand) []string {
	indices := rng.Perm(len(inventory))
	items := make([]string, n)
	for i := 0; i < n; i++ {
		items[i] = inventory[indices[i]]
	}
	return items
}

// RespawnPoint returns the respawn location for a player who died at position.
func (m *ZoneManager) RespawnPoint(x, z float64) (respawnX, respawnZ float64) {
	zone := m.GetZoneAt(x, z)
	if zone != nil && (zone.RespawnX != 0 || zone.RespawnZ != 0) {
		return zone.RespawnX, zone.RespawnZ
	}
	// Default: respawn at origin
	return 0, 0
}

// ZoneCount returns the number of registered zones.
func (m *ZoneManager) ZoneCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.zones)
}

// FlaggedPlayerCount returns the number of PvP-flagged players.
func (m *ZoneManager) FlaggedPlayerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, flagged := range m.playerFlags {
		if flagged {
			count++
		}
	}
	return count
}
