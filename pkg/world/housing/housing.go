// Package housing provides player housing and guild territory management.
// Per ROADMAP Phase 5 item 22:
// AC: Player-placed furniture persists across 3 server restarts;
// guild territory boundary enforced.
package housing

import (
	"encoding/gob"
	"fmt"
	"sync"
)

func init() {
	// Register types for gob encoding
	gob.Register(&House{})
	gob.Register(&FurnitureItem{})
	gob.Register(&GuildTerritory{})
}

// FurnitureItem represents a piece of furniture placed in a house.
type FurnitureItem struct {
	ID        string
	Type      string // furniture type (bed, table, chair, etc.)
	X, Y, Z   float64
	Rotation  float64 // rotation angle in radians
	Condition float64 // 0.0 - 1.0 durability
}

// House represents a player-owned house with interior layout.
type House struct {
	ID          string
	OwnerID     uint64 // Player entity ID
	WorldX      float64 // World position
	WorldZ      float64
	InteriorID  string // Interior instance ID
	Furniture   []FurnitureItem
	PurchaseDay int // Game day when purchased
}

// HouseManager manages player housing.
type HouseManager struct {
	mu         sync.RWMutex
	houses     map[string]*House // ID -> House
	ownerIndex map[uint64][]string // OwnerID -> House IDs
}

// NewHouseManager creates a new house manager.
func NewHouseManager() *HouseManager {
	return &HouseManager{
		houses:     make(map[string]*House),
		ownerIndex: make(map[uint64][]string),
	}
}

// RegisterHouse adds a house to the manager.
func (m *HouseManager) RegisterHouse(house *House) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.houses[house.ID] = house
	m.ownerIndex[house.OwnerID] = append(m.ownerIndex[house.OwnerID], house.ID)
}

// GetHouse returns a house by ID.
func (m *HouseManager) GetHouse(houseID string) *House {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.houses[houseID]
}

// GetPlayerHouses returns all houses owned by a player.
func (m *HouseManager) GetPlayerHouses(ownerID uint64) []*House {
	m.mu.RLock()
	defer m.mu.RUnlock()

	houseIDs := m.ownerIndex[ownerID]
	houses := make([]*House, 0, len(houseIDs))
	for _, id := range houseIDs {
		if h := m.houses[id]; h != nil {
			houses = append(houses, h)
		}
	}
	return houses
}

// PlaceFurniture adds furniture to a house.
func (m *HouseManager) PlaceFurniture(houseID string, item FurnitureItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	house := m.houses[houseID]
	if house == nil {
		return fmt.Errorf("house not found: %s", houseID)
	}

	house.Furniture = append(house.Furniture, item)
	return nil
}

// RemoveFurniture removes furniture from a house.
func (m *HouseManager) RemoveFurniture(houseID, furnitureID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	house := m.houses[houseID]
	if house == nil {
		return fmt.Errorf("house not found: %s", houseID)
	}

	for i, item := range house.Furniture {
		if item.ID == furnitureID {
			house.Furniture = append(house.Furniture[:i], house.Furniture[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("furniture not found: %s", furnitureID)
}

// MoveFurniture updates furniture position.
func (m *HouseManager) MoveFurniture(houseID, furnitureID string, x, y, z, rotation float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	house := m.houses[houseID]
	if house == nil {
		return fmt.Errorf("house not found: %s", houseID)
	}

	for i, item := range house.Furniture {
		if item.ID == furnitureID {
			house.Furniture[i].X = x
			house.Furniture[i].Y = y
			house.Furniture[i].Z = z
			house.Furniture[i].Rotation = rotation
			return nil
		}
	}
	return fmt.Errorf("furniture not found: %s", furnitureID)
}

// TransferOwnership changes house ownership.
func (m *HouseManager) TransferOwnership(houseID string, newOwnerID uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	house := m.houses[houseID]
	if house == nil {
		return fmt.Errorf("house not found: %s", houseID)
	}

	// Remove from old owner's index
	oldOwnerID := house.OwnerID
	oldHouses := m.ownerIndex[oldOwnerID]
	for i, id := range oldHouses {
		if id == houseID {
			m.ownerIndex[oldOwnerID] = append(oldHouses[:i], oldHouses[i+1:]...)
			break
		}
	}

	// Add to new owner's index
	house.OwnerID = newOwnerID
	m.ownerIndex[newOwnerID] = append(m.ownerIndex[newOwnerID], houseID)

	return nil
}

// HouseCount returns the total number of houses.
func (m *HouseManager) HouseCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.houses)
}

// ExportAll returns all houses for persistence.
func (m *HouseManager) ExportAll() []*House {
	m.mu.RLock()
	defer m.mu.RUnlock()

	houses := make([]*House, 0, len(m.houses))
	for _, h := range m.houses {
		houses = append(houses, h)
	}
	return houses
}

// ImportAll loads houses from persistence.
func (m *HouseManager) ImportAll(houses []*House) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.houses = make(map[string]*House)
	m.ownerIndex = make(map[uint64][]string)

	for _, h := range houses {
		m.houses[h.ID] = h
		m.ownerIndex[h.OwnerID] = append(m.ownerIndex[h.OwnerID], h.ID)
	}
}

// GuildTerritory represents a guild-claimed region.
type GuildTerritory struct {
	GuildID     string
	Name        string
	CenterX     float64
	CenterZ     float64
	Radius      float64 // Circular territory
	ClaimDay    int     // Game day when claimed
	ControlFlag float64 // 0.0 - 1.0 control strength
}

// ContainsPoint checks if a point is within the territory.
func (t *GuildTerritory) ContainsPoint(x, z float64) bool {
	dx := x - t.CenterX
	dz := z - t.CenterZ
	return dx*dx+dz*dz <= t.Radius*t.Radius
}

// GuildManager manages guild territories.
type GuildManager struct {
	mu          sync.RWMutex
	territories map[string]*GuildTerritory // GuildID -> Territory
	members     map[string][]uint64        // GuildID -> Member EntityIDs
}

// NewGuildManager creates a new guild manager.
func NewGuildManager() *GuildManager {
	return &GuildManager{
		territories: make(map[string]*GuildTerritory),
		members:     make(map[string][]uint64),
	}
}

// ClaimTerritory registers a guild territory claim.
func (m *GuildManager) ClaimTerritory(territory *GuildTerritory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for overlap with existing territories
	for _, existing := range m.territories {
		if m.territoriesOverlap(existing, territory) {
			return fmt.Errorf("territory overlaps with %s", existing.Name)
		}
	}

	m.territories[territory.GuildID] = territory
	return nil
}

// territoriesOverlap checks if two circular territories overlap.
func (m *GuildManager) territoriesOverlap(a, b *GuildTerritory) bool {
	dx := a.CenterX - b.CenterX
	dz := a.CenterZ - b.CenterZ
	distSq := dx*dx + dz*dz
	sumRadii := a.Radius + b.Radius
	return distSq < sumRadii*sumRadii
}

// GetTerritory returns a guild's territory.
func (m *GuildManager) GetTerritory(guildID string) *GuildTerritory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.territories[guildID]
}

// GetTerritoryAt returns the territory containing a point, or nil.
func (m *GuildManager) GetTerritoryAt(x, z float64) *GuildTerritory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, t := range m.territories {
		if t.ContainsPoint(x, z) {
			return t
		}
	}
	return nil
}

// ReleaseTerritory removes a guild's territory claim.
func (m *GuildManager) ReleaseTerritory(guildID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.territories, guildID)
}

// AddMember adds a player to a guild.
func (m *GuildManager) AddMember(guildID string, entityID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.members[guildID] = append(m.members[guildID], entityID)
}

// RemoveMember removes a player from a guild.
func (m *GuildManager) RemoveMember(guildID string, entityID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	members := m.members[guildID]
	for i, id := range members {
		if id == entityID {
			m.members[guildID] = append(members[:i], members[i+1:]...)
			return
		}
	}
}

// GetMembers returns all members of a guild.
func (m *GuildManager) GetMembers(guildID string) []uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.members[guildID]
}

// IsMember checks if a player is in a guild.
func (m *GuildManager) IsMember(guildID string, entityID uint64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, id := range m.members[guildID] {
		if id == entityID {
			return true
		}
	}
	return false
}

// CanAccessTerritory checks if a player can access a territory at position.
// Per AC: guild territory boundary enforced.
func (m *GuildManager) CanAccessTerritory(entityID uint64, x, z float64) bool {
	territory := m.GetTerritoryAt(x, z)
	if territory == nil {
		return true // No territory = public access
	}
	return m.IsMember(territory.GuildID, entityID)
}

// TerritoryCount returns the number of claimed territories.
func (m *GuildManager) TerritoryCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.territories)
}

// ExportTerritories returns all territories for persistence.
func (m *GuildManager) ExportTerritories() []*GuildTerritory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	territories := make([]*GuildTerritory, 0, len(m.territories))
	for _, t := range m.territories {
		territories = append(territories, t)
	}
	return territories
}

// ImportTerritories loads territories from persistence.
func (m *GuildManager) ImportTerritories(territories []*GuildTerritory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.territories = make(map[string]*GuildTerritory)
	for _, t := range territories {
		m.territories[t.GuildID] = t
	}
}

// ExportMembers returns guild membership for persistence.
func (m *GuildManager) ExportMembers() map[string][]uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]uint64)
	for guildID, members := range m.members {
		result[guildID] = append([]uint64{}, members...)
	}
	return result
}

// ImportMembers loads guild membership from persistence.
func (m *GuildManager) ImportMembers(members map[string][]uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.members = members
}
