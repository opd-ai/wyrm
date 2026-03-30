// Package housing provides player housing and guild territory management.
// Per ROADMAP Phase 5 item 22:
// AC: Player-placed furniture persists across 3 server restarts;
// guild territory boundary enforced.
package housing

import (
	"encoding/gob"
	"fmt"
	"math"
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
	OwnerID     uint64  // Player entity ID
	WorldX      float64 // World position
	WorldZ      float64
	InteriorID  string // Interior instance ID
	Furniture   []FurnitureItem
	PurchaseDay int // Game day when purchased
}

// HouseManager manages player housing.
type HouseManager struct {
	mu         sync.RWMutex
	houses     map[string]*House   // ID -> House
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

// ============================================================================
// Property Purchasing System
// ============================================================================

// PropertyListing represents a house available for purchase.
type PropertyListing struct {
	ID          string
	Name        string
	Description string
	WorldX      float64
	WorldZ      float64
	BasePrice   int
	Size        int     // Small=1, Medium=2, Large=3
	Quality     float64 // 0.0-1.0 affects price
	DistrictID  string
	Genre       string
}

// PropertyMarket manages property listings and transactions.
type PropertyMarket struct {
	mu           sync.RWMutex
	listings     map[string]*PropertyListing
	priceFactors map[string]float64 // District -> price multiplier
	houseManager *HouseManager
}

// NewPropertyMarket creates a new property market.
func NewPropertyMarket(houseManager *HouseManager) *PropertyMarket {
	return &PropertyMarket{
		listings:     make(map[string]*PropertyListing),
		priceFactors: make(map[string]float64),
		houseManager: houseManager,
	}
}

// AddListing adds a property to the market.
func (m *PropertyMarket) AddListing(listing *PropertyListing) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listings[listing.ID] = listing
}

// RemoveListing removes a property from the market.
func (m *PropertyMarket) RemoveListing(listingID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.listings, listingID)
}

// SetDistrictPriceFactor sets price multiplier for a district.
func (m *PropertyMarket) SetDistrictPriceFactor(districtID string, factor float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.priceFactors[districtID] = factor
}

// GetCurrentPrice calculates current price with district factor.
func (m *PropertyMarket) GetCurrentPrice(listingID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	listing := m.listings[listingID]
	if listing == nil {
		return 0
	}

	factor := m.priceFactors[listing.DistrictID]
	if factor <= 0 {
		factor = 1.0
	}

	// Price = base * quality * size * district factor
	price := float64(listing.BasePrice) * (0.5 + listing.Quality*0.5)
	price *= float64(listing.Size)
	price *= factor

	return int(price)
}

// GetAvailableListings returns all available properties.
func (m *PropertyMarket) GetAvailableListings() []*PropertyListing {
	m.mu.RLock()
	defer m.mu.RUnlock()

	listings := make([]*PropertyListing, 0, len(m.listings))
	for _, l := range m.listings {
		listings = append(listings, l)
	}
	return listings
}

// GetListingsByDistrict filters listings by district.
func (m *PropertyMarket) GetListingsByDistrict(districtID string) []*PropertyListing {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var listings []*PropertyListing
	for _, l := range m.listings {
		if l.DistrictID == districtID {
			listings = append(listings, l)
		}
	}
	return listings
}

// PurchaseResult contains the outcome of a property purchase.
type PurchaseResult struct {
	Success   bool
	House     *House
	Message   string
	PricePaid int
}

// PurchaseProperty attempts to purchase a property for a player.
func (m *PropertyMarket) PurchaseProperty(
	listingID string,
	buyerID uint64,
	buyerGold int,
	currentDay int,
) PurchaseResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	listing := m.listings[listingID]
	if listing == nil {
		return PurchaseResult{
			Success: false,
			Message: "Property not found",
		}
	}

	// Calculate price
	factor := m.priceFactors[listing.DistrictID]
	if factor <= 0 {
		factor = 1.0
	}
	price := float64(listing.BasePrice) * (0.5 + listing.Quality*0.5)
	price *= float64(listing.Size)
	price *= factor
	finalPrice := int(price)

	// Check if buyer can afford
	if buyerGold < finalPrice {
		return PurchaseResult{
			Success: false,
			Message: "Insufficient gold",
		}
	}

	// Create the house
	house := &House{
		ID:          listing.ID,
		OwnerID:     buyerID,
		WorldX:      listing.WorldX,
		WorldZ:      listing.WorldZ,
		InteriorID:  listing.ID + "-interior",
		Furniture:   []FurnitureItem{},
		PurchaseDay: currentDay,
	}

	// Register with house manager
	m.houseManager.RegisterHouse(house)

	// Remove from listings
	delete(m.listings, listingID)

	return PurchaseResult{
		Success:   true,
		House:     house,
		Message:   "Property purchased successfully",
		PricePaid: finalPrice,
	}
}

// ListingCount returns the number of available listings.
func (m *PropertyMarket) ListingCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.listings)
}

// ============================================================================
// First-Person Furniture Placement System
// ============================================================================

// PlacementMode represents the furniture placement UI mode.
type PlacementMode int

const (
	PlacementModeNone PlacementMode = iota
	PlacementModePlace
	PlacementModeMove
	PlacementModeRotate
	PlacementModeRemove
)

// FurniturePlacement manages first-person furniture placement UI state.
type FurniturePlacement struct {
	mu            sync.RWMutex
	Mode          PlacementMode
	SelectedID    string  // Currently selected furniture ID
	SelectedType  string  // Type of furniture being placed
	PreviewX      float64 // Preview position X
	PreviewY      float64 // Preview position Y
	PreviewZ      float64 // Preview position Z
	PreviewRot    float64 // Preview rotation
	CurrentHouse  string  // Current house being edited
	SnapToGrid    bool    // Whether to snap to grid
	GridSize      float64 // Grid snap size
	ValidPosition bool    // Whether preview position is valid
}

// NewFurniturePlacement creates a new furniture placement controller.
func NewFurniturePlacement() *FurniturePlacement {
	return &FurniturePlacement{
		Mode:       PlacementModeNone,
		SnapToGrid: true,
		GridSize:   0.5, // Half-unit grid
	}
}

// StartPlaceMode enters furniture placement mode.
func (fp *FurniturePlacement) StartPlaceMode(houseID, furnitureType string) {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	fp.Mode = PlacementModePlace
	fp.CurrentHouse = houseID
	fp.SelectedType = furnitureType
	fp.SelectedID = ""
	fp.ValidPosition = false
}

// StartMoveMode enters furniture move mode.
func (fp *FurniturePlacement) StartMoveMode(houseID, furnitureID string) {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	fp.Mode = PlacementModeMove
	fp.CurrentHouse = houseID
	fp.SelectedID = furnitureID
	fp.SelectedType = ""
	fp.ValidPosition = false
}

// StartRotateMode enters furniture rotation mode.
func (fp *FurniturePlacement) StartRotateMode(houseID, furnitureID string) {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	fp.Mode = PlacementModeRotate
	fp.CurrentHouse = houseID
	fp.SelectedID = furnitureID
}

// StartRemoveMode enters furniture removal mode.
func (fp *FurniturePlacement) StartRemoveMode(houseID, furnitureID string) {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	fp.Mode = PlacementModeRemove
	fp.CurrentHouse = houseID
	fp.SelectedID = furnitureID
}

// ExitMode exits the current placement mode.
func (fp *FurniturePlacement) ExitMode() {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	fp.Mode = PlacementModeNone
	fp.SelectedID = ""
	fp.SelectedType = ""
	fp.CurrentHouse = ""
	fp.ValidPosition = false
}

// UpdatePreview updates the preview position from player view direction.
func (fp *FurniturePlacement) UpdatePreview(playerX, playerZ, viewAngle, distance float64) {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	if fp.Mode == PlacementModeNone {
		return
	}

	// Calculate position in front of player
	newX := playerX + math.Cos(viewAngle)*distance
	newZ := playerZ + math.Sin(viewAngle)*distance

	// Apply grid snap if enabled
	if fp.SnapToGrid && fp.GridSize > 0 {
		newX = math.Floor(newX/fp.GridSize) * fp.GridSize
		newZ = math.Floor(newZ/fp.GridSize) * fp.GridSize
	}

	fp.PreviewX = newX
	fp.PreviewY = 0 // Ground level by default
	fp.PreviewZ = newZ
	fp.ValidPosition = true
}

// RotatePreview rotates the preview furniture.
func (fp *FurniturePlacement) RotatePreview(deltaAngle float64) {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	fp.PreviewRot += deltaAngle
	// Normalize to 0-2π
	for fp.PreviewRot < 0 {
		fp.PreviewRot += 2 * 3.14159265
	}
	for fp.PreviewRot >= 2*3.14159265 {
		fp.PreviewRot -= 2 * 3.14159265
	}
}

// SetGridSnap enables or disables grid snapping.
func (fp *FurniturePlacement) SetGridSnap(enabled bool, gridSize float64) {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	fp.SnapToGrid = enabled
	if gridSize > 0 {
		fp.GridSize = gridSize
	}
}

// GetPreviewState returns current preview state for rendering.
func (fp *FurniturePlacement) GetPreviewState() (mode PlacementMode, x, y, z, rot float64, valid bool) {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	return fp.Mode, fp.PreviewX, fp.PreviewY, fp.PreviewZ, fp.PreviewRot, fp.ValidPosition
}

// ConfirmPlacement commits the current placement to the house.
func (fp *FurniturePlacement) ConfirmPlacement(houseManager *HouseManager, newID string) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	if fp.Mode == PlacementModeNone || !fp.ValidPosition {
		return fmt.Errorf("no valid placement to confirm")
	}

	switch fp.Mode {
	case PlacementModePlace:
		item := FurnitureItem{
			ID:        newID,
			Type:      fp.SelectedType,
			X:         fp.PreviewX,
			Y:         fp.PreviewY,
			Z:         fp.PreviewZ,
			Rotation:  fp.PreviewRot,
			Condition: 1.0,
		}
		err := houseManager.PlaceFurniture(fp.CurrentHouse, item)
		if err != nil {
			return err
		}

	case PlacementModeMove:
		err := houseManager.MoveFurniture(
			fp.CurrentHouse,
			fp.SelectedID,
			fp.PreviewX,
			fp.PreviewY,
			fp.PreviewZ,
			fp.PreviewRot,
		)
		if err != nil {
			return err
		}

	case PlacementModeRemove:
		err := houseManager.RemoveFurniture(fp.CurrentHouse, fp.SelectedID)
		if err != nil {
			return err
		}
	}

	// Reset mode
	fp.Mode = PlacementModeNone
	fp.SelectedID = ""
	fp.SelectedType = ""
	fp.ValidPosition = false

	return nil
}

// IsInPlacementMode returns whether placement UI is active.
func (fp *FurniturePlacement) IsInPlacementMode() bool {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.Mode != PlacementModeNone
}

// GetCurrentHouse returns the house being edited.
func (fp *FurniturePlacement) GetCurrentHouse() string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.CurrentHouse
}

// ============================================================================
// Rent Collection System
// ============================================================================

// RentStatus represents the rental status of a property.
type RentStatus int

const (
	RentStatusVacant RentStatus = iota
	RentStatusOccupied
	RentStatusOverdue
	RentStatusEvicting
)

// RentalProperty represents a property available for rent.
type RentalProperty struct {
	ID            string
	OwnerID       uint64 // Player who owns and rents out the property
	TenantID      uint64 // NPC or player tenant (0 if vacant)
	Name          string
	MonthlyRent   float64 // Monthly rent amount
	LastPayment   float64 // Game time of last payment
	NextPayment   float64 // Game time when next payment is due
	Status        RentStatus
	Deposit       float64    // Security deposit held
	LeaseDuration float64    // Lease length in game hours
	LeaseStart    float64    // When current lease started
	Quality       float64    // Property quality affects rent
	Location      [3]float64 // World position
	Condition     float64    // Property condition affects tenant satisfaction
	RentHistory   []RentPayment
	OverdueDays   int // Days payment is overdue
}

// RentPayment records a rent payment.
type RentPayment struct {
	Amount    float64
	Timestamp float64
	OnTime    bool
}

// TenantInfo represents information about a tenant.
type TenantInfo struct {
	ID             uint64
	Name           string
	Reliability    float64 // 0-1 chance to pay on time
	WealthLevel    float64 // 0-1 affects ability to pay
	Satisfaction   float64 // 0-1 tenant satisfaction
	LeasesHeld     int     // Number of properties rented
	PaymentsMade   int     // Total payments made
	PaymentsMissed int     // Total payments missed
}

// RentCollectionSystem manages rental properties and rent collection.
type RentCollectionSystem struct {
	mu              sync.RWMutex
	Seed            int64
	Genre           string
	Properties      map[string]*RentalProperty
	Tenants         map[uint64]*TenantInfo
	OwnerProperties map[uint64][]string // Owner -> property IDs
	OwnerIncome     map[uint64]float64  // Accumulated rental income
	GameTime        float64
	counter         uint64
	PaymentPeriod   float64 // Hours between rent payments
	EvictionGrace   int     // Days before eviction starts
	DepositMultiple float64 // Deposit as multiple of rent
}

// NewRentCollectionSystem creates a new rent collection system.
func NewRentCollectionSystem(seed int64, genre string) *RentCollectionSystem {
	return &RentCollectionSystem{
		Seed:            seed,
		Genre:           genre,
		Properties:      make(map[string]*RentalProperty),
		Tenants:         make(map[uint64]*TenantInfo),
		OwnerProperties: make(map[uint64][]string),
		OwnerIncome:     make(map[uint64]float64),
		PaymentPeriod:   720.0, // 30 days (720 hours)
		EvictionGrace:   7,     // 7 days grace period
		DepositMultiple: 2.0,   // 2 months deposit
	}
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *RentCollectionSystem) pseudoRandom() float64 {
	s.counter++
	x := uint64(s.Seed) + s.counter*6364136223846793005
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	return float64(x%10000) / 10000.0
}

// RegisterProperty registers a property for rent.
func (s *RentCollectionSystem) RegisterProperty(prop *RentalProperty) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Properties[prop.ID] = prop
	s.OwnerProperties[prop.OwnerID] = append(s.OwnerProperties[prop.OwnerID], prop.ID)
}

// ListPropertyForRent makes a property available for rent.
func (s *RentCollectionSystem) ListPropertyForRent(ownerID uint64, propID, name string, monthlyRent, quality float64, location [3]float64) *RentalProperty {
	s.mu.Lock()
	defer s.mu.Unlock()

	prop := &RentalProperty{
		ID:          propID,
		OwnerID:     ownerID,
		TenantID:    0,
		Name:        name,
		MonthlyRent: monthlyRent,
		Status:      RentStatusVacant,
		Deposit:     monthlyRent * s.DepositMultiple,
		Quality:     quality,
		Location:    location,
		Condition:   1.0,
		RentHistory: make([]RentPayment, 0),
	}

	s.Properties[propID] = prop
	s.OwnerProperties[ownerID] = append(s.OwnerProperties[ownerID], propID)

	return prop
}

// AddTenant creates a new tenant profile.
func (s *RentCollectionSystem) AddTenant(tenantID uint64, name string, reliability, wealth float64) *TenantInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	tenant := &TenantInfo{
		ID:           tenantID,
		Name:         name,
		Reliability:  clampFloat64(reliability, 0, 1),
		WealthLevel:  clampFloat64(wealth, 0, 1),
		Satisfaction: 0.5, // Neutral starting satisfaction
	}

	s.Tenants[tenantID] = tenant
	return tenant
}

// RentProperty assigns a tenant to a property.
func (s *RentCollectionSystem) RentProperty(propID string, tenantID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prop := s.Properties[propID]
	if prop == nil {
		return fmt.Errorf("property not found: %s", propID)
	}

	if prop.Status != RentStatusVacant {
		return fmt.Errorf("property is not vacant")
	}

	tenant := s.Tenants[tenantID]
	if tenant == nil {
		return fmt.Errorf("tenant not found: %d", tenantID)
	}

	prop.TenantID = tenantID
	prop.Status = RentStatusOccupied
	prop.LeaseStart = s.GameTime
	prop.LeaseDuration = s.PaymentPeriod * 12 // 1 year default
	prop.LastPayment = s.GameTime
	prop.NextPayment = s.GameTime + s.PaymentPeriod
	prop.OverdueDays = 0

	tenant.LeasesHeld++

	return nil
}

// EvictTenant removes a tenant from a property.
func (s *RentCollectionSystem) EvictTenant(propID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prop := s.Properties[propID]
	if prop == nil {
		return fmt.Errorf("property not found: %s", propID)
	}

	if prop.TenantID == 0 {
		return fmt.Errorf("property has no tenant")
	}

	if tenant := s.Tenants[prop.TenantID]; tenant != nil {
		tenant.LeasesHeld--
		if tenant.LeasesHeld < 0 {
			tenant.LeasesHeld = 0
		}
	}

	prop.TenantID = 0
	prop.Status = RentStatusVacant
	prop.OverdueDays = 0
	prop.LeaseStart = 0
	prop.NextPayment = 0

	return nil
}

// ProcessRentPayment processes a rent payment from a tenant.
func (s *RentCollectionSystem) ProcessRentPayment(propID string, amount float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prop := s.Properties[propID]
	if prop == nil {
		return 0, fmt.Errorf("property not found: %s", propID)
	}

	if prop.TenantID == 0 {
		return 0, fmt.Errorf("property has no tenant")
	}

	// Check if payment is on time
	onTime := s.GameTime <= prop.NextPayment

	// Record payment
	payment := RentPayment{
		Amount:    amount,
		Timestamp: s.GameTime,
		OnTime:    onTime,
	}
	prop.RentHistory = append(prop.RentHistory, payment)
	if len(prop.RentHistory) > 24 {
		prop.RentHistory = prop.RentHistory[1:]
	}

	// Update tenant stats
	if tenant := s.Tenants[prop.TenantID]; tenant != nil {
		tenant.PaymentsMade++
		if onTime {
			tenant.Satisfaction = clampFloat64(tenant.Satisfaction+0.05, 0, 1)
		}
	}

	// Update property status
	prop.LastPayment = s.GameTime
	prop.NextPayment = s.GameTime + s.PaymentPeriod
	prop.Status = RentStatusOccupied
	prop.OverdueDays = 0

	// Credit owner
	s.OwnerIncome[prop.OwnerID] += amount

	return amount, nil
}

// Update processes rent collection each tick.
func (s *RentCollectionSystem) Update(dt float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GameTime += dt
	hoursDelta := dt / 3600.0

	for _, prop := range s.Properties {
		if prop.TenantID == 0 {
			continue
		}

		s.updateProperty(prop, hoursDelta)
	}
}

// updateProperty processes a single rental property.
func (s *RentCollectionSystem) updateProperty(prop *RentalProperty, hours float64) {
	// Check if payment is due
	if s.GameTime >= prop.NextPayment {
		// Check if tenant pays
		tenant := s.Tenants[prop.TenantID]
		if tenant != nil && s.tenantPays(tenant, prop) {
			// Auto-pay rent
			payment := RentPayment{
				Amount:    prop.MonthlyRent,
				Timestamp: s.GameTime,
				OnTime:    prop.OverdueDays == 0,
			}
			prop.RentHistory = append(prop.RentHistory, payment)
			prop.LastPayment = s.GameTime
			prop.NextPayment = s.GameTime + s.PaymentPeriod
			prop.OverdueDays = 0
			prop.Status = RentStatusOccupied
			tenant.PaymentsMade++
			s.OwnerIncome[prop.OwnerID] += prop.MonthlyRent
		} else {
			// Payment missed
			prop.OverdueDays++
			if tenant != nil {
				tenant.PaymentsMissed++
				tenant.Satisfaction = clampFloat64(tenant.Satisfaction-0.1, 0, 1)
			}

			if prop.OverdueDays >= s.EvictionGrace {
				prop.Status = RentStatusEvicting
			} else {
				prop.Status = RentStatusOverdue
			}
		}
	}

	// Degrade property condition over time
	prop.Condition = clampFloat64(prop.Condition-0.0001*hours, 0.1, 1.0)

	// Tenant satisfaction affected by condition
	if tenant := s.Tenants[prop.TenantID]; tenant != nil {
		if prop.Condition < 0.5 {
			tenant.Satisfaction = clampFloat64(tenant.Satisfaction-0.001*hours, 0, 1)
		}
	}
}

// tenantPays determines if a tenant will pay rent.
func (s *RentCollectionSystem) tenantPays(tenant *TenantInfo, prop *RentalProperty) bool {
	// Base chance from reliability
	chance := tenant.Reliability

	// Wealth affects ability to pay
	chance *= (0.5 + tenant.WealthLevel*0.5)

	// Satisfaction affects willingness
	chance *= (0.5 + tenant.Satisfaction*0.5)

	// Property condition affects payment
	chance *= (0.5 + prop.Condition*0.5)

	return s.pseudoRandom() < chance
}

// CollectIncome withdraws accumulated rental income for an owner.
func (s *RentCollectionSystem) CollectIncome(ownerID uint64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	income := s.OwnerIncome[ownerID]
	s.OwnerIncome[ownerID] = 0
	return income
}

// GetOwnerIncome returns accumulated income without collecting it.
func (s *RentCollectionSystem) GetOwnerIncome(ownerID uint64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.OwnerIncome[ownerID]
}

// GetOwnerProperties returns all properties owned by a player.
func (s *RentCollectionSystem) GetOwnerProperties(ownerID uint64) []*RentalProperty {
	s.mu.RLock()
	defer s.mu.RUnlock()

	propIDs := s.OwnerProperties[ownerID]
	props := make([]*RentalProperty, 0, len(propIDs))
	for _, id := range propIDs {
		if prop := s.Properties[id]; prop != nil {
			props = append(props, prop)
		}
	}
	return props
}

// GetProperty returns a specific property.
func (s *RentCollectionSystem) GetProperty(propID string) *RentalProperty {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Properties[propID]
}

// GetTenant returns a specific tenant.
func (s *RentCollectionSystem) GetTenant(tenantID uint64) *TenantInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Tenants[tenantID]
}

// GetVacantProperties returns all vacant properties.
func (s *RentCollectionSystem) GetVacantProperties() []*RentalProperty {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vacant := make([]*RentalProperty, 0)
	for _, prop := range s.Properties {
		if prop.Status == RentStatusVacant {
			vacant = append(vacant, prop)
		}
	}
	return vacant
}

// SetRent updates the monthly rent for a property.
func (s *RentCollectionSystem) SetRent(propID string, newRent float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prop := s.Properties[propID]
	if prop == nil {
		return fmt.Errorf("property not found: %s", propID)
	}

	prop.MonthlyRent = newRent
	prop.Deposit = newRent * s.DepositMultiple

	return nil
}

// RepairProperty improves property condition.
func (s *RentCollectionSystem) RepairProperty(propID string, repairAmount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prop := s.Properties[propID]
	if prop == nil {
		return fmt.Errorf("property not found: %s", propID)
	}

	prop.Condition = clampFloat64(prop.Condition+repairAmount, 0, 1)

	// Improving condition improves tenant satisfaction
	if tenant := s.Tenants[prop.TenantID]; tenant != nil {
		tenant.Satisfaction = clampFloat64(tenant.Satisfaction+repairAmount*0.5, 0, 1)
	}

	return nil
}

// GetRentHistory returns payment history for a property.
func (s *RentCollectionSystem) GetRentHistory(propID string) []RentPayment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prop := s.Properties[propID]
	if prop == nil {
		return nil
	}
	return prop.RentHistory
}

// CalculateTotalRentalValue calculates total monthly rental income potential.
func (s *RentCollectionSystem) CalculateTotalRentalValue(ownerID uint64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0.0
	for _, propID := range s.OwnerProperties[ownerID] {
		if prop := s.Properties[propID]; prop != nil {
			total += prop.MonthlyRent
		}
	}
	return total
}

// CalculateOccupancyRate calculates percentage of occupied properties.
func (s *RentCollectionSystem) CalculateOccupancyRate(ownerID uint64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	propIDs := s.OwnerProperties[ownerID]
	if len(propIDs) == 0 {
		return 0
	}

	occupied := 0
	for _, propID := range propIDs {
		if prop := s.Properties[propID]; prop != nil && prop.TenantID != 0 {
			occupied++
		}
	}

	return float64(occupied) / float64(len(propIDs))
}

// GetOverdueProperties returns properties with overdue rent.
func (s *RentCollectionSystem) GetOverdueProperties(ownerID uint64) []*RentalProperty {
	s.mu.RLock()
	defer s.mu.RUnlock()

	overdue := make([]*RentalProperty, 0)
	for _, propID := range s.OwnerProperties[ownerID] {
		if prop := s.Properties[propID]; prop != nil {
			if prop.Status == RentStatusOverdue || prop.Status == RentStatusEvicting {
				overdue = append(overdue, prop)
			}
		}
	}
	return overdue
}

// clampFloat64 clamps a value between min and max.
func clampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ============================================================================
// Home Upgrades System
// ============================================================================

// UpgradeCategory represents a category of home upgrades.
type UpgradeCategory int

const (
	UpgradeCategorySecurity UpgradeCategory = iota
	UpgradeCategoryComfort
	UpgradeCategoryStorage
	UpgradeCategoryAesthetics
	UpgradeCategoryUtility
	UpgradeCategoryDefense
)

// UpgradeStatus represents the status of an upgrade.
type UpgradeStatus int

const (
	UpgradeStatusAvailable UpgradeStatus = iota
	UpgradeStatusInProgress
	UpgradeStatusCompleted
	UpgradeStatusLocked
)

// HomeUpgrade represents a purchasable home upgrade.
type HomeUpgrade struct {
	ID            string
	Name          string
	Description   string
	Category      UpgradeCategory
	Cost          float64        // Gold cost
	MaterialCost  map[string]int // Required materials
	InstallTime   float64        // Hours to install
	Status        UpgradeStatus
	Progress      float64        // 0-1 installation progress
	Effects       UpgradeEffects // What the upgrade provides
	Prerequisites []string       // Required upgrade IDs
	Genre         string         // Genre-specific upgrade
	Level         int            // Upgrade tier (1-5)
}

// UpgradeEffects describes what a home upgrade provides.
type UpgradeEffects struct {
	SecurityBonus float64 // Reduces theft/break-in chance
	ComfortBonus  float64 // Increases rest effectiveness
	StorageBonus  int     // Additional storage slots
	ValueBonus    float64 // Increases property value
	TenantBonus   float64 // Increases rent potential
	CraftingBonus float64 // Improves crafting in home
	DefenseBonus  float64 // Combat defense bonus at home
	SpecialEffect string  // Special effect identifier
}

// UpgradedHome tracks upgrades applied to a house.
type UpgradedHome struct {
	HouseID       string
	Upgrades      map[string]*HomeUpgrade // Upgrade ID -> upgrade
	TotalValue    float64
	SecurityLevel float64
	ComfortLevel  float64
	StorageSlots  int
	DefenseLevel  float64
	InstalledDate map[string]float64 // Upgrade ID -> installation time
}

// HomeUpgradeSystem manages home upgrades.
type HomeUpgradeSystem struct {
	mu                sync.RWMutex
	Seed              int64
	Genre             string
	AvailableUpgrades map[string]*HomeUpgrade  // All possible upgrades
	HomeUpgrades      map[string]*UpgradedHome // HouseID -> upgrades
	GameTime          float64
	counter           uint64
}

// NewHomeUpgradeSystem creates a new home upgrade system.
func NewHomeUpgradeSystem(seed int64, genre string) *HomeUpgradeSystem {
	sys := &HomeUpgradeSystem{
		Seed:              seed,
		Genre:             genre,
		AvailableUpgrades: make(map[string]*HomeUpgrade),
		HomeUpgrades:      make(map[string]*UpgradedHome),
	}
	sys.initializeUpgrades()
	return sys
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *HomeUpgradeSystem) pseudoRandom() float64 {
	s.counter++
	x := uint64(s.Seed) + s.counter*6364136223846793005
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	return float64(x%10000) / 10000.0
}

// initializeUpgrades populates the available upgrades based on genre.
func (s *HomeUpgradeSystem) initializeUpgrades() {
	// Security upgrades (all genres)
	s.addUpgrade(&HomeUpgrade{
		ID:          "lock_basic",
		Name:        s.getGenreName("Basic Lock", "Basic Lock", "Basic Lock", "Basic Keypad", "Basic Lock"),
		Description: "A simple lock for your door",
		Category:    UpgradeCategorySecurity,
		Cost:        100,
		InstallTime: 1.0,
		Status:      UpgradeStatusAvailable,
		Effects:     UpgradeEffects{SecurityBonus: 0.1},
		Level:       1,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "lock_advanced",
		Name:          s.getGenreName("Reinforced Lock", "Biometric Lock", "Warded Lock", "Quantum Lock", "Reinforced Lock"),
		Description:   "An advanced locking mechanism",
		Category:      UpgradeCategorySecurity,
		Cost:          500,
		InstallTime:   2.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{SecurityBonus: 0.25},
		Prerequisites: []string{"lock_basic"},
		Level:         2,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "alarm_system",
		Name:          s.getGenreName("Guard Bell", "Motion Sensors", "Spirit Ward", "Security Drones", "Trip Wires"),
		Description:   "Alerts you to intruders",
		Category:      UpgradeCategorySecurity,
		Cost:          1000,
		InstallTime:   4.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{SecurityBonus: 0.35, SpecialEffect: "intruder_alert"},
		Prerequisites: []string{"lock_advanced"},
		Level:         3,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "vault",
		Name:          s.getGenreName("Iron Vault", "Secure Vault", "Cursed Vault", "Quantum Vault", "Bunker Safe"),
		Description:   "A secure vault for valuables",
		Category:      UpgradeCategorySecurity,
		Cost:          2500,
		InstallTime:   8.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{SecurityBonus: 0.5, StorageBonus: 10, SpecialEffect: "secure_storage"},
		Prerequisites: []string{"alarm_system"},
		Level:         4,
	})

	// Comfort upgrades
	s.addUpgrade(&HomeUpgrade{
		ID:          "bedding_basic",
		Name:        s.getGenreName("Straw Mattress", "Foam Mattress", "Simple Cot", "Gel Mattress", "Salvaged Bed"),
		Description: "Basic sleeping accommodations",
		Category:    UpgradeCategoryComfort,
		Cost:        50,
		InstallTime: 0.5,
		Status:      UpgradeStatusAvailable,
		Effects:     UpgradeEffects{ComfortBonus: 0.1},
		Level:       1,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "bedding_luxury",
		Name:          s.getGenreName("Feather Bed", "Memory Foam", "Restful Bed", "Neural Rest Pod", "Luxury Mattress"),
		Description:   "Luxurious sleeping quarters",
		Category:      UpgradeCategoryComfort,
		Cost:          300,
		InstallTime:   1.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{ComfortBonus: 0.25, SpecialEffect: "better_rest"},
		Prerequisites: []string{"bedding_basic"},
		Level:         2,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:          "heating",
		Name:        s.getGenreName("Stone Fireplace", "Climate Control", "Brazier", "Thermal System", "Wood Stove"),
		Description: "Temperature control for comfort",
		Category:    UpgradeCategoryComfort,
		Cost:        400,
		InstallTime: 4.0,
		Status:      UpgradeStatusAvailable,
		Effects:     UpgradeEffects{ComfortBonus: 0.15, TenantBonus: 0.1},
		Level:       2,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "bath",
		Name:          s.getGenreName("Copper Tub", "Shower Unit", "Ritual Bath", "Sonic Cleaner", "Rain Barrel"),
		Description:   "Bathing facilities",
		Category:      UpgradeCategoryComfort,
		Cost:          600,
		InstallTime:   6.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{ComfortBonus: 0.2, TenantBonus: 0.15, SpecialEffect: "cleanse"},
		Prerequisites: []string{"heating"},
		Level:         3,
	})

	// Storage upgrades
	s.addUpgrade(&HomeUpgrade{
		ID:          "chest_basic",
		Name:        s.getGenreName("Wooden Chest", "Storage Locker", "Trunk", "Storage Unit", "Footlocker"),
		Description: "Basic storage container",
		Category:    UpgradeCategoryStorage,
		Cost:        75,
		InstallTime: 0.5,
		Status:      UpgradeStatusAvailable,
		Effects:     UpgradeEffects{StorageBonus: 20},
		Level:       1,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "wardrobe",
		Name:          s.getGenreName("Oak Wardrobe", "Closet System", "Armoire", "Nano-Closet", "Salvaged Cabinet"),
		Description:   "Organized clothing storage",
		Category:      UpgradeCategoryStorage,
		Cost:          200,
		InstallTime:   2.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{StorageBonus: 30, SpecialEffect: "outfit_storage"},
		Prerequisites: []string{"chest_basic"},
		Level:         2,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "pantry",
		Name:          s.getGenreName("Root Cellar", "Refrigeration Unit", "Cold Storage", "Stasis Chamber", "Food Cache"),
		Description:   "Food preservation storage",
		Category:      UpgradeCategoryStorage,
		Cost:          350,
		InstallTime:   4.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{StorageBonus: 25, SpecialEffect: "food_preservation"},
		Prerequisites: []string{"chest_basic"},
		Level:         2,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "armory",
		Name:          s.getGenreName("Weapon Rack", "Arsenal", "Dark Armory", "Weapon Cache", "Gun Locker"),
		Description:   "Secure weapon storage",
		Category:      UpgradeCategoryStorage,
		Cost:          500,
		InstallTime:   3.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{StorageBonus: 15, SecurityBonus: 0.1, SpecialEffect: "weapon_storage"},
		Prerequisites: []string{"wardrobe"},
		Level:         3,
	})

	// Aesthetics upgrades
	s.addUpgrade(&HomeUpgrade{
		ID:          "decor_basic",
		Name:        s.getGenreName("Tapestries", "Wall Art", "Candles", "Holograms", "Salvaged Decor"),
		Description: "Basic decorative items",
		Category:    UpgradeCategoryAesthetics,
		Cost:        100,
		InstallTime: 1.0,
		Status:      UpgradeStatusAvailable,
		Effects:     UpgradeEffects{ComfortBonus: 0.05, ValueBonus: 0.1},
		Level:       1,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "decor_luxury",
		Name:          s.getGenreName("Fine Art", "Designer Pieces", "Occult Artifacts", "Cyber Art", "Pre-War Relics"),
		Description:   "Luxurious decorations",
		Category:      UpgradeCategoryAesthetics,
		Cost:          750,
		InstallTime:   2.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{ComfortBonus: 0.15, ValueBonus: 0.3, TenantBonus: 0.2},
		Prerequisites: []string{"decor_basic"},
		Level:         3,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:          "garden",
		Name:        s.getGenreName("Herb Garden", "Hydroponic Bay", "Cursed Garden", "Rooftop Garden", "Survival Garden"),
		Description: "A small garden area",
		Category:    UpgradeCategoryAesthetics,
		Cost:        400,
		InstallTime: 6.0,
		Status:      UpgradeStatusAvailable,
		Effects:     UpgradeEffects{ComfortBonus: 0.1, ValueBonus: 0.15, SpecialEffect: "ingredient_source"},
		Level:       2,
	})

	// Utility upgrades
	s.addUpgrade(&HomeUpgrade{
		ID:          "workbench",
		Name:        s.getGenreName("Crafting Table", "Fabricator", "Ritual Table", "Nano-Printer", "Workshop Bench"),
		Description: "A place to craft items",
		Category:    UpgradeCategoryUtility,
		Cost:        300,
		InstallTime: 3.0,
		Status:      UpgradeStatusAvailable,
		Effects:     UpgradeEffects{CraftingBonus: 0.1, SpecialEffect: "home_crafting"},
		Level:       1,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "workbench_advanced",
		Name:          s.getGenreName("Master Forge", "Advanced Fabricator", "Dark Forge", "Matter Compiler", "Full Workshop"),
		Description:   "Advanced crafting station",
		Category:      UpgradeCategoryUtility,
		Cost:          1500,
		InstallTime:   8.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{CraftingBonus: 0.3, SpecialEffect: "advanced_crafting"},
		Prerequisites: []string{"workbench"},
		Level:         3,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:          "study",
		Name:        s.getGenreName("Study Desk", "Data Terminal", "Occult Library", "Neural Interface", "Research Station"),
		Description: "A place to study and research",
		Category:    UpgradeCategoryUtility,
		Cost:        250,
		InstallTime: 2.0,
		Status:      UpgradeStatusAvailable,
		Effects:     UpgradeEffects{SpecialEffect: "skill_training"},
		Level:       1,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "alchemy_lab",
		Name:          s.getGenreName("Alchemy Lab", "Chem Lab", "Dark Lab", "Synthesis Lab", "Chem Station"),
		Description:   "Laboratory for creating potions and chemicals",
		Category:      UpgradeCategoryUtility,
		Cost:          800,
		InstallTime:   6.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{CraftingBonus: 0.2, SpecialEffect: "potion_crafting"},
		Prerequisites: []string{"workbench", "study"},
		Level:         3,
	})

	// Defense upgrades
	s.addUpgrade(&HomeUpgrade{
		ID:          "fortification_basic",
		Name:        s.getGenreName("Wooden Shutters", "Blast Shutters", "Iron Bars", "Armored Panels", "Barricades"),
		Description: "Basic fortifications",
		Category:    UpgradeCategoryDefense,
		Cost:        200,
		InstallTime: 4.0,
		Status:      UpgradeStatusAvailable,
		Effects:     UpgradeEffects{DefenseBonus: 0.1, SecurityBonus: 0.1},
		Level:       1,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "fortification_advanced",
		Name:          s.getGenreName("Stone Walls", "Force Fields", "Warded Walls", "Energy Shields", "Reinforced Walls"),
		Description:   "Advanced fortifications",
		Category:      UpgradeCategoryDefense,
		Cost:          1200,
		InstallTime:   12.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{DefenseBonus: 0.3, SecurityBonus: 0.25},
		Prerequisites: []string{"fortification_basic"},
		Level:         3,
	})

	s.addUpgrade(&HomeUpgrade{
		ID:            "turret",
		Name:          s.getGenreName("Guard Golem", "Auto-Turret", "Spirit Guardian", "Defense Drone", "Mounted Gun"),
		Description:   "Automated defense system",
		Category:      UpgradeCategoryDefense,
		Cost:          3000,
		InstallTime:   16.0,
		Status:        UpgradeStatusLocked,
		Effects:       UpgradeEffects{DefenseBonus: 0.5, SecurityBonus: 0.4, SpecialEffect: "auto_defense"},
		Prerequisites: []string{"fortification_advanced", "alarm_system"},
		Level:         5,
	})
}

// getGenreName returns the appropriate name based on genre.
func (s *HomeUpgradeSystem) getGenreName(fantasy, scifi, horror, cyberpunk, postapoc string) string {
	switch s.Genre {
	case "sci-fi":
		return scifi
	case "horror":
		return horror
	case "cyberpunk":
		return cyberpunk
	case "post-apocalyptic":
		return postapoc
	default:
		return fantasy
	}
}

// addUpgrade adds an upgrade to available upgrades.
func (s *HomeUpgradeSystem) addUpgrade(upgrade *HomeUpgrade) {
	upgrade.Genre = s.Genre
	s.AvailableUpgrades[upgrade.ID] = upgrade
}

// RegisterHome registers a home for upgrades.
func (s *HomeUpgradeSystem) RegisterHome(houseID string) *UpgradedHome {
	s.mu.Lock()
	defer s.mu.Unlock()

	home := &UpgradedHome{
		HouseID:       houseID,
		Upgrades:      make(map[string]*HomeUpgrade),
		StorageSlots:  10, // Base storage
		InstalledDate: make(map[string]float64),
	}
	s.HomeUpgrades[houseID] = home
	return home
}

// GetUpgradedHome returns upgrade info for a home.
func (s *HomeUpgradeSystem) GetUpgradedHome(houseID string) *UpgradedHome {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.HomeUpgrades[houseID]
}

// GetAvailableUpgrades returns all upgrades available for a home.
func (s *HomeUpgradeSystem) GetAvailableUpgrades(houseID string) []*HomeUpgrade {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return nil
	}

	available := make([]*HomeUpgrade, 0)
	for _, upgrade := range s.AvailableUpgrades {
		status := s.getUpgradeStatus(home, upgrade)
		if status == UpgradeStatusAvailable {
			upgradeCopy := *upgrade
			upgradeCopy.Status = status
			available = append(available, &upgradeCopy)
		}
	}
	return available
}

// getUpgradeStatus determines the current status of an upgrade for a home.
func (s *HomeUpgradeSystem) getUpgradeStatus(home *UpgradedHome, upgrade *HomeUpgrade) UpgradeStatus {
	// Already installed
	if _, ok := home.Upgrades[upgrade.ID]; ok {
		return UpgradeStatusCompleted
	}

	// Check prerequisites
	for _, prereq := range upgrade.Prerequisites {
		if _, ok := home.Upgrades[prereq]; !ok {
			return UpgradeStatusLocked
		}
	}

	return UpgradeStatusAvailable
}

// CanInstallUpgrade checks if an upgrade can be installed.
func (s *HomeUpgradeSystem) CanInstallUpgrade(houseID, upgradeID string) (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return false, "home not registered"
	}

	upgrade := s.AvailableUpgrades[upgradeID]
	if upgrade == nil {
		return false, "upgrade not found"
	}

	// Already installed
	if _, ok := home.Upgrades[upgradeID]; ok {
		return false, "upgrade already installed"
	}

	// Check prerequisites
	for _, prereq := range upgrade.Prerequisites {
		if _, ok := home.Upgrades[prereq]; !ok {
			prereqUpgrade := s.AvailableUpgrades[prereq]
			if prereqUpgrade != nil {
				return false, fmt.Sprintf("requires %s", prereqUpgrade.Name)
			}
			return false, "missing prerequisite"
		}
	}

	return true, ""
}

// StartInstallation begins installing an upgrade.
func (s *HomeUpgradeSystem) StartInstallation(houseID, upgradeID string, playerGold float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return 0, fmt.Errorf("home not registered")
	}

	upgrade := s.AvailableUpgrades[upgradeID]
	if upgrade == nil {
		return 0, fmt.Errorf("upgrade not found")
	}

	// Check prerequisites
	for _, prereq := range upgrade.Prerequisites {
		if _, ok := home.Upgrades[prereq]; !ok {
			return 0, fmt.Errorf("missing prerequisite: %s", prereq)
		}
	}

	// Check if already installed
	if _, ok := home.Upgrades[upgradeID]; ok {
		return 0, fmt.Errorf("upgrade already installed")
	}

	// Check cost
	if playerGold < upgrade.Cost {
		return 0, fmt.Errorf("insufficient gold: need %.0f, have %.0f", upgrade.Cost, playerGold)
	}

	// Start installation
	newUpgrade := *upgrade
	newUpgrade.Status = UpgradeStatusInProgress
	newUpgrade.Progress = 0
	home.Upgrades[upgradeID] = &newUpgrade

	return upgrade.Cost, nil
}

// CompleteInstallation instantly completes an upgrade installation.
func (s *HomeUpgradeSystem) CompleteInstallation(houseID, upgradeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return fmt.Errorf("home not registered")
	}

	upgrade := home.Upgrades[upgradeID]
	if upgrade == nil {
		return fmt.Errorf("upgrade not found in home")
	}

	upgrade.Status = UpgradeStatusCompleted
	upgrade.Progress = 1.0
	home.InstalledDate[upgradeID] = s.GameTime

	// Apply effects
	s.applyUpgradeEffects(home, upgrade)

	return nil
}

// applyUpgradeEffects applies the effects of an upgrade to a home.
func (s *HomeUpgradeSystem) applyUpgradeEffects(home *UpgradedHome, upgrade *HomeUpgrade) {
	home.SecurityLevel += upgrade.Effects.SecurityBonus
	home.ComfortLevel += upgrade.Effects.ComfortBonus
	home.StorageSlots += upgrade.Effects.StorageBonus
	home.TotalValue += upgrade.Cost * (1 + upgrade.Effects.ValueBonus)
	home.DefenseLevel += upgrade.Effects.DefenseBonus
}

// Update processes upgrade installations.
func (s *HomeUpgradeSystem) Update(dt float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GameTime += dt
	hoursDelta := dt / 3600.0

	for _, home := range s.HomeUpgrades {
		for _, upgrade := range home.Upgrades {
			if upgrade.Status == UpgradeStatusInProgress {
				// Progress installation
				progressRate := 1.0 / upgrade.InstallTime // Progress per hour
				upgrade.Progress += progressRate * hoursDelta

				if upgrade.Progress >= 1.0 {
					upgrade.Progress = 1.0
					upgrade.Status = UpgradeStatusCompleted
					home.InstalledDate[upgrade.ID] = s.GameTime
					s.applyUpgradeEffects(home, upgrade)
				}
			}
		}
	}
}

// GetUpgradesByCategory returns upgrades filtered by category.
func (s *HomeUpgradeSystem) GetUpgradesByCategory(category UpgradeCategory) []*HomeUpgrade {
	s.mu.RLock()
	defer s.mu.RUnlock()

	upgrades := make([]*HomeUpgrade, 0)
	for _, upgrade := range s.AvailableUpgrades {
		if upgrade.Category == category {
			upgrades = append(upgrades, upgrade)
		}
	}
	return upgrades
}

// GetHomeStats returns aggregate stats for a home.
func (s *HomeUpgradeSystem) GetHomeStats(houseID string) (security, comfort, defense, value float64, storage int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return 0, 0, 0, 0, 10
	}

	return home.SecurityLevel, home.ComfortLevel, home.DefenseLevel, home.TotalValue, home.StorageSlots
}

// HasUpgrade checks if a home has a specific upgrade installed.
func (s *HomeUpgradeSystem) HasUpgrade(houseID, upgradeID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return false
	}

	upgrade, ok := home.Upgrades[upgradeID]
	return ok && upgrade.Status == UpgradeStatusCompleted
}

// HasSpecialEffect checks if a home has a specific special effect.
func (s *HomeUpgradeSystem) HasSpecialEffect(houseID, effect string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return false
	}

	for _, upgrade := range home.Upgrades {
		if upgrade.Status == UpgradeStatusCompleted && upgrade.Effects.SpecialEffect == effect {
			return true
		}
	}
	return false
}

// GetTenantBonus calculates total tenant bonus for rental properties.
func (s *HomeUpgradeSystem) GetTenantBonus(houseID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return 0
	}

	bonus := 0.0
	for _, upgrade := range home.Upgrades {
		if upgrade.Status == UpgradeStatusCompleted {
			bonus += upgrade.Effects.TenantBonus
		}
	}
	return bonus
}

// GetCraftingBonus calculates total crafting bonus for a home.
func (s *HomeUpgradeSystem) GetCraftingBonus(houseID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return 0
	}

	bonus := 0.0
	for _, upgrade := range home.Upgrades {
		if upgrade.Status == UpgradeStatusCompleted {
			bonus += upgrade.Effects.CraftingBonus
		}
	}
	return bonus
}

// GetInstalledUpgrades returns all installed upgrades for a home.
func (s *HomeUpgradeSystem) GetInstalledUpgrades(houseID string) []*HomeUpgrade {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return nil
	}

	installed := make([]*HomeUpgrade, 0)
	for _, upgrade := range home.Upgrades {
		if upgrade.Status == UpgradeStatusCompleted {
			installed = append(installed, upgrade)
		}
	}
	return installed
}

// GetUpgradeProgress returns the progress of an in-progress upgrade.
func (s *HomeUpgradeSystem) GetUpgradeProgress(houseID, upgradeID string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return 0, false
	}

	upgrade, ok := home.Upgrades[upgradeID]
	if !ok {
		return 0, false
	}

	return upgrade.Progress, upgrade.Status == UpgradeStatusInProgress
}

// UpgradeCount returns the number of installed upgrades.
func (s *HomeUpgradeSystem) UpgradeCount(houseID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return 0
	}

	count := 0
	for _, upgrade := range home.Upgrades {
		if upgrade.Status == UpgradeStatusCompleted {
			count++
		}
	}
	return count
}

// ============================================================================
// Guild Halls System
// ============================================================================

// GuildHallTier represents the tier of a guild hall.
type GuildHallTier int

const (
	GuildHallTierBasic GuildHallTier = iota + 1
	GuildHallTierStandard
	GuildHallTierGrand
	GuildHallTierImperial
	GuildHallTierLegendary
)

// GuildRank represents a member's rank in a guild.
type GuildRank int

const (
	GuildRankMember GuildRank = iota
	GuildRankOfficer
	GuildRankVeteran
	GuildRankCouncil
	GuildRankLeader
)

// GuildFacility represents a facility in a guild hall.
type GuildFacility struct {
	ID          string
	Name        string
	Description string
	Type        string // crafting, training, storage, defense, social
	Level       int    // 1-5
	Operational bool
	Capacity    int     // How many can use at once
	CurrentUse  int     // Current users
	Cooldown    float64 // Time between uses
	Bonuses     FacilityBonuses
}

// FacilityBonuses describes bonuses from a facility.
type FacilityBonuses struct {
	CraftingBonus float64
	TrainingBonus float64
	StorageBonus  int
	DefenseBonus  float64
	SocialBonus   float64
	IncomeBonus   float64
}

// GuildHall represents a guild's headquarters.
type GuildHall struct {
	ID            string
	GuildID       string
	Name          string
	Tier          GuildHallTier
	Location      [3]float64
	Facilities    map[string]*GuildFacility
	TreasuryGold  float64
	BankCapacity  int      // Shared storage capacity
	BankItems     []string // Item IDs in guild bank
	DefenseRating float64
	Upkeep        float64 // Daily upkeep cost
	Founded       float64 // Game time when founded
	LastUpkeep    float64 // Last upkeep payment time
	InDebt        bool    // Guild is in debt
	DebtDays      int     // Days in debt
}

// GuildHallSystem manages guild halls.
type GuildHallSystem struct {
	mu          sync.RWMutex
	Seed        int64
	Genre       string
	GuildHalls  map[string]*GuildHall           // GuildID -> Hall
	MemberRanks map[string]map[uint64]GuildRank // GuildID -> MemberID -> Rank
	Permissions map[GuildRank][]string          // Rank -> allowed actions
	GameTime    float64
	counter     uint64
}

// NewGuildHallSystem creates a new guild hall system.
func NewGuildHallSystem(seed int64, genre string) *GuildHallSystem {
	sys := &GuildHallSystem{
		Seed:        seed,
		Genre:       genre,
		GuildHalls:  make(map[string]*GuildHall),
		MemberRanks: make(map[string]map[uint64]GuildRank),
		Permissions: make(map[GuildRank][]string),
	}
	sys.initializePermissions()
	return sys
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *GuildHallSystem) pseudoRandom() float64 {
	s.counter++
	x := uint64(s.Seed) + s.counter*6364136223846793005
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	return float64(x%10000) / 10000.0
}

// initializePermissions sets up rank-based permissions.
func (s *GuildHallSystem) initializePermissions() {
	s.Permissions[GuildRankMember] = []string{
		"use_facilities", "access_bank_personal", "view_roster",
	}
	s.Permissions[GuildRankOfficer] = []string{
		"use_facilities", "access_bank_personal", "view_roster",
		"access_bank_guild", "invite_members",
	}
	s.Permissions[GuildRankVeteran] = []string{
		"use_facilities", "access_bank_personal", "view_roster",
		"access_bank_guild", "invite_members", "kick_members",
	}
	s.Permissions[GuildRankCouncil] = []string{
		"use_facilities", "access_bank_personal", "view_roster",
		"access_bank_guild", "invite_members", "kick_members",
		"manage_facilities", "access_treasury",
	}
	s.Permissions[GuildRankLeader] = []string{
		"use_facilities", "access_bank_personal", "view_roster",
		"access_bank_guild", "invite_members", "kick_members",
		"manage_facilities", "access_treasury", "manage_ranks",
		"upgrade_hall", "disband_guild",
	}
}

// CreateGuildHall creates a new guild hall.
func (s *GuildHallSystem) CreateGuildHall(guildID, hallName string, leaderID uint64, location [3]float64) (*GuildHall, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.GuildHalls[guildID]; exists {
		return nil, fmt.Errorf("guild already has a hall")
	}

	hall := &GuildHall{
		ID:           fmt.Sprintf("hall_%s", guildID),
		GuildID:      guildID,
		Name:         hallName,
		Tier:         GuildHallTierBasic,
		Location:     location,
		Facilities:   make(map[string]*GuildFacility),
		BankCapacity: 50,
		BankItems:    make([]string, 0),
		Upkeep:       10.0,
		Founded:      s.GameTime,
		LastUpkeep:   s.GameTime,
	}

	// Add basic facilities
	s.addBasicFacilities(hall)

	s.GuildHalls[guildID] = hall

	// Set up ranks
	s.MemberRanks[guildID] = make(map[uint64]GuildRank)
	s.MemberRanks[guildID][leaderID] = GuildRankLeader

	return hall, nil
}

// addBasicFacilities adds starting facilities to a guild hall.
func (s *GuildHallSystem) addBasicFacilities(hall *GuildHall) {
	hall.Facilities["meeting_hall"] = &GuildFacility{
		ID:          "meeting_hall",
		Name:        s.getGenreFacilityName("meeting_hall"),
		Description: "A place for guild members to gather",
		Type:        "social",
		Level:       1,
		Operational: true,
		Capacity:    20,
		Bonuses:     FacilityBonuses{SocialBonus: 0.1},
	}

	hall.Facilities["storage_room"] = &GuildFacility{
		ID:          "storage_room",
		Name:        s.getGenreFacilityName("storage_room"),
		Description: "Basic guild storage",
		Type:        "storage",
		Level:       1,
		Operational: true,
		Capacity:    1,
		Bonuses:     FacilityBonuses{StorageBonus: 50},
	}
}

// getGenreFacilityName returns genre-appropriate facility names.
func (s *GuildHallSystem) getGenreFacilityName(facilityType string) string {
	names := map[string]map[string]string{
		"meeting_hall": {
			"fantasy":          "Great Hall",
			"sci-fi":           "Conference Deck",
			"horror":           "Gathering Chamber",
			"cyberpunk":        "NetHub",
			"post-apocalyptic": "Community Room",
		},
		"storage_room": {
			"fantasy":          "Vault Chamber",
			"sci-fi":           "Cargo Bay",
			"horror":           "Storage Cellar",
			"cyberpunk":        "Secure Locker",
			"post-apocalyptic": "Supply Cache",
		},
		"training_ground": {
			"fantasy":          "Training Grounds",
			"sci-fi":           "Simulation Deck",
			"horror":           "Dark Arena",
			"cyberpunk":        "Combat Sim",
			"post-apocalyptic": "Fighting Pit",
		},
		"crafting_hall": {
			"fantasy":          "Artisan Workshop",
			"sci-fi":           "Fabrication Lab",
			"horror":           "Ritual Forge",
			"cyberpunk":        "Tech Lab",
			"post-apocalyptic": "Workshop",
		},
		"defense_tower": {
			"fantasy":          "Guard Tower",
			"sci-fi":           "Defense Array",
			"horror":           "Watchtower",
			"cyberpunk":        "Security Node",
			"post-apocalyptic": "Sniper Nest",
		},
	}

	if facilityNames, ok := names[facilityType]; ok {
		if name, ok := facilityNames[s.Genre]; ok {
			return name
		}
		return facilityNames["fantasy"]
	}
	return facilityType
}

// GetGuildHall returns a guild's hall.
func (s *GuildHallSystem) GetGuildHall(guildID string) *GuildHall {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.GuildHalls[guildID]
}

// UpgradeGuildHall upgrades a guild hall to the next tier.
func (s *GuildHallSystem) UpgradeGuildHall(guildID string, memberID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return fmt.Errorf("guild hall not found")
	}

	if !s.hasPermissionLocked(guildID, memberID, "upgrade_hall") {
		return fmt.Errorf("insufficient rank to upgrade hall")
	}

	if hall.Tier >= GuildHallTierLegendary {
		return fmt.Errorf("hall is already at maximum tier")
	}

	// Check treasury for upgrade cost
	cost := s.getUpgradeCost(hall.Tier)
	if hall.TreasuryGold < cost {
		return fmt.Errorf("insufficient treasury funds: need %.0f, have %.0f", cost, hall.TreasuryGold)
	}

	hall.TreasuryGold -= cost
	hall.Tier++
	hall.BankCapacity += 50
	hall.Upkeep *= 1.5

	// Add new facilities for higher tiers
	s.addTierFacilities(hall)

	return nil
}

// getUpgradeCost returns the cost to upgrade from current tier.
func (s *GuildHallSystem) getUpgradeCost(currentTier GuildHallTier) float64 {
	costs := map[GuildHallTier]float64{
		GuildHallTierBasic:    1000,
		GuildHallTierStandard: 5000,
		GuildHallTierGrand:    25000,
		GuildHallTierImperial: 100000,
	}
	return costs[currentTier]
}

// addTierFacilities adds facilities appropriate for the hall's tier.
func (s *GuildHallSystem) addTierFacilities(hall *GuildHall) {
	switch hall.Tier {
	case GuildHallTierStandard:
		hall.Facilities["training_ground"] = &GuildFacility{
			ID:          "training_ground",
			Name:        s.getGenreFacilityName("training_ground"),
			Description: "Train combat and skills",
			Type:        "training",
			Level:       1,
			Operational: true,
			Capacity:    5,
			Cooldown:    3600, // 1 hour
			Bonuses:     FacilityBonuses{TrainingBonus: 0.15},
		}
	case GuildHallTierGrand:
		hall.Facilities["crafting_hall"] = &GuildFacility{
			ID:          "crafting_hall",
			Name:        s.getGenreFacilityName("crafting_hall"),
			Description: "Advanced crafting facilities",
			Type:        "crafting",
			Level:       2,
			Operational: true,
			Capacity:    3,
			Cooldown:    1800, // 30 min
			Bonuses:     FacilityBonuses{CraftingBonus: 0.2},
		}
	case GuildHallTierImperial:
		hall.Facilities["defense_tower"] = &GuildFacility{
			ID:          "defense_tower",
			Name:        s.getGenreFacilityName("defense_tower"),
			Description: "Defensive fortifications",
			Type:        "defense",
			Level:       3,
			Operational: true,
			Capacity:    2,
			Bonuses:     FacilityBonuses{DefenseBonus: 0.3},
		}
	case GuildHallTierLegendary:
		// Upgrade all existing facilities
		for _, facility := range hall.Facilities {
			facility.Level++
			facility.Capacity++
		}
	}
}

// AddMemberRank adds or updates a member's rank.
func (s *GuildHallSystem) AddMemberRank(guildID string, memberID uint64, rank GuildRank) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.MemberRanks[guildID] == nil {
		s.MemberRanks[guildID] = make(map[uint64]GuildRank)
	}
	s.MemberRanks[guildID][memberID] = rank
}

// GetMemberRank returns a member's rank.
func (s *GuildHallSystem) GetMemberRank(guildID string, memberID uint64) GuildRank {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ranks, ok := s.MemberRanks[guildID]; ok {
		return ranks[memberID]
	}
	return GuildRankMember
}

// PromoteMember promotes a member to a higher rank.
func (s *GuildHallSystem) PromoteMember(guildID string, promoterID, targetID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.hasPermissionLocked(guildID, promoterID, "manage_ranks") {
		return fmt.Errorf("insufficient rank to promote members")
	}

	promoterRank := s.MemberRanks[guildID][promoterID]
	currentRank := s.MemberRanks[guildID][targetID]

	if currentRank >= promoterRank {
		return fmt.Errorf("cannot promote to equal or higher rank")
	}

	if currentRank >= GuildRankCouncil {
		return fmt.Errorf("cannot promote beyond council")
	}

	s.MemberRanks[guildID][targetID] = currentRank + 1
	return nil
}

// DemoteMember demotes a member to a lower rank.
func (s *GuildHallSystem) DemoteMember(guildID string, demoterID, targetID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.hasPermissionLocked(guildID, demoterID, "manage_ranks") {
		return fmt.Errorf("insufficient rank to demote members")
	}

	demoterRank := s.MemberRanks[guildID][demoterID]
	currentRank := s.MemberRanks[guildID][targetID]

	if currentRank >= demoterRank {
		return fmt.Errorf("cannot demote equal or higher rank")
	}

	if currentRank <= GuildRankMember {
		return fmt.Errorf("cannot demote below member")
	}

	s.MemberRanks[guildID][targetID] = currentRank - 1
	return nil
}

// hasPermission checks if a member has a specific permission.
func (s *GuildHallSystem) hasPermission(guildID string, memberID uint64, permission string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasPermissionLocked(guildID, memberID, permission)
}

// hasPermissionLocked checks permission without locking (caller must hold lock).
func (s *GuildHallSystem) hasPermissionLocked(guildID string, memberID uint64, permission string) bool {
	ranks, guildExists := s.MemberRanks[guildID]
	if !guildExists {
		return false
	}
	rank, memberExists := ranks[memberID]
	if !memberExists {
		return false
	}
	perms := s.Permissions[rank]
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// HasPermission checks if a member has a specific permission (public).
func (s *GuildHallSystem) HasPermission(guildID string, memberID uint64, permission string) bool {
	return s.hasPermission(guildID, memberID, permission)
}

// DepositToTreasury adds gold to the guild treasury.
func (s *GuildHallSystem) DepositToTreasury(guildID string, memberID uint64, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return fmt.Errorf("guild hall not found")
	}

	hall.TreasuryGold += amount
	if hall.InDebt && hall.TreasuryGold >= hall.Upkeep {
		hall.InDebt = false
		hall.DebtDays = 0
	}

	return nil
}

// WithdrawFromTreasury removes gold from the guild treasury.
func (s *GuildHallSystem) WithdrawFromTreasury(guildID string, memberID uint64, amount float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return 0, fmt.Errorf("guild hall not found")
	}

	if !s.hasPermissionLocked(guildID, memberID, "access_treasury") {
		return 0, fmt.Errorf("insufficient rank to access treasury")
	}

	if hall.TreasuryGold < amount {
		return 0, fmt.Errorf("insufficient treasury funds")
	}

	hall.TreasuryGold -= amount
	return amount, nil
}

// GetTreasuryBalance returns the guild treasury balance.
func (s *GuildHallSystem) GetTreasuryBalance(guildID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return 0
	}
	return hall.TreasuryGold
}

// DepositToBank adds an item to the guild bank.
func (s *GuildHallSystem) DepositToBank(guildID string, memberID uint64, itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return fmt.Errorf("guild hall not found")
	}

	if !s.hasPermissionLocked(guildID, memberID, "access_bank_guild") {
		return fmt.Errorf("insufficient rank to deposit to guild bank")
	}

	if len(hall.BankItems) >= hall.BankCapacity {
		return fmt.Errorf("guild bank is full")
	}

	hall.BankItems = append(hall.BankItems, itemID)
	return nil
}

// WithdrawFromBank removes an item from the guild bank.
func (s *GuildHallSystem) WithdrawFromBank(guildID string, memberID uint64, itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return fmt.Errorf("guild hall not found")
	}

	if !s.hasPermissionLocked(guildID, memberID, "access_bank_guild") {
		return fmt.Errorf("insufficient rank to withdraw from guild bank")
	}

	for i, id := range hall.BankItems {
		if id == itemID {
			hall.BankItems = append(hall.BankItems[:i], hall.BankItems[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("item not found in guild bank")
}

// GetBankContents returns the guild bank contents.
func (s *GuildHallSystem) GetBankContents(guildID string, memberID uint64) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return nil, fmt.Errorf("guild hall not found")
	}

	// Any member can view, but we check for basic access
	if _, ok := s.MemberRanks[guildID][memberID]; !ok {
		return nil, fmt.Errorf("not a guild member")
	}

	return hall.BankItems, nil
}

// UseFacility marks a facility as being used.
func (s *GuildHallSystem) UseFacility(guildID string, memberID uint64, facilityID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return fmt.Errorf("guild hall not found")
	}

	if !s.hasPermissionLocked(guildID, memberID, "use_facilities") {
		return fmt.Errorf("insufficient rank to use facilities")
	}

	facility := hall.Facilities[facilityID]
	if facility == nil {
		return fmt.Errorf("facility not found")
	}

	if !facility.Operational {
		return fmt.Errorf("facility is not operational")
	}

	if facility.CurrentUse >= facility.Capacity {
		return fmt.Errorf("facility is at capacity")
	}

	facility.CurrentUse++
	return nil
}

// ReleaseFacility releases a facility slot.
func (s *GuildHallSystem) ReleaseFacility(guildID, facilityID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return
	}

	facility := hall.Facilities[facilityID]
	if facility == nil {
		return
	}

	if facility.CurrentUse > 0 {
		facility.CurrentUse--
	}
}

// GetFacility returns a specific facility.
func (s *GuildHallSystem) GetFacility(guildID, facilityID string) *GuildFacility {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return nil
	}
	return hall.Facilities[facilityID]
}

// GetAllFacilities returns all facilities in a guild hall.
func (s *GuildHallSystem) GetAllFacilities(guildID string) []*GuildFacility {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return nil
	}

	facilities := make([]*GuildFacility, 0, len(hall.Facilities))
	for _, f := range hall.Facilities {
		facilities = append(facilities, f)
	}
	return facilities
}

// UpgradeFacility upgrades a specific facility.
func (s *GuildHallSystem) UpgradeFacility(guildID string, memberID uint64, facilityID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return fmt.Errorf("guild hall not found")
	}

	if !s.hasPermissionLocked(guildID, memberID, "manage_facilities") {
		return fmt.Errorf("insufficient rank to manage facilities")
	}

	facility := hall.Facilities[facilityID]
	if facility == nil {
		return fmt.Errorf("facility not found")
	}

	if facility.Level >= 5 {
		return fmt.Errorf("facility is at maximum level")
	}

	cost := float64(facility.Level) * 500
	if hall.TreasuryGold < cost {
		return fmt.Errorf("insufficient treasury funds")
	}

	hall.TreasuryGold -= cost
	facility.Level++
	facility.Capacity++

	// Increase bonuses
	facility.Bonuses.CraftingBonus *= 1.25
	facility.Bonuses.TrainingBonus *= 1.25
	facility.Bonuses.StorageBonus += 10
	facility.Bonuses.DefenseBonus *= 1.25
	facility.Bonuses.SocialBonus *= 1.25

	return nil
}

// Update processes guild hall upkeep.
func (s *GuildHallSystem) Update(dt float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GameTime += dt
	dayLength := 24.0 * 3600.0 // 24 hours in seconds

	for _, hall := range s.GuildHalls {
		// Check for daily upkeep
		if s.GameTime-hall.LastUpkeep >= dayLength {
			hall.LastUpkeep = s.GameTime

			if hall.TreasuryGold >= hall.Upkeep {
				hall.TreasuryGold -= hall.Upkeep
				hall.InDebt = false
				hall.DebtDays = 0
			} else {
				hall.InDebt = true
				hall.DebtDays++

				// Facilities become non-operational after 3 days in debt
				if hall.DebtDays >= 3 {
					for _, facility := range hall.Facilities {
						facility.Operational = false
					}
				}
			}
		}
	}
}

// GetGuildHallStats returns aggregate stats for a guild hall.
func (s *GuildHallSystem) GetGuildHallStats(guildID string) (tier GuildHallTier, treasury float64, bankUsed, bankCap int, defense float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return 0, 0, 0, 0, 0
	}

	// Calculate defense from facilities
	for _, f := range hall.Facilities {
		if f.Operational {
			defense += f.Bonuses.DefenseBonus
		}
	}

	return hall.Tier, hall.TreasuryGold, len(hall.BankItems), hall.BankCapacity, defense
}

// GetTotalBonuses calculates total bonuses from all operational facilities.
func (s *GuildHallSystem) GetTotalBonuses(guildID string) FacilityBonuses {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return FacilityBonuses{}
	}

	total := FacilityBonuses{}
	for _, f := range hall.Facilities {
		if f.Operational {
			total.CraftingBonus += f.Bonuses.CraftingBonus
			total.TrainingBonus += f.Bonuses.TrainingBonus
			total.StorageBonus += f.Bonuses.StorageBonus
			total.DefenseBonus += f.Bonuses.DefenseBonus
			total.SocialBonus += f.Bonuses.SocialBonus
			total.IncomeBonus += f.Bonuses.IncomeBonus
		}
	}
	return total
}

// IsInDebt returns whether a guild is in debt.
func (s *GuildHallSystem) IsInDebt(guildID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return false
	}
	return hall.InDebt
}

// DisbandGuildHall removes a guild hall.
func (s *GuildHallSystem) DisbandGuildHall(guildID string, memberID uint64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hall := s.GuildHalls[guildID]
	if hall == nil {
		return 0, fmt.Errorf("guild hall not found")
	}

	if !s.hasPermissionLocked(guildID, memberID, "disband_guild") {
		return 0, fmt.Errorf("only the leader can disband the guild hall")
	}

	// Return remaining treasury
	remaining := hall.TreasuryGold

	delete(s.GuildHalls, guildID)
	delete(s.MemberRanks, guildID)

	return remaining, nil
}

// GetMemberCount returns the number of members in a guild.
func (s *GuildHallSystem) GetMemberCount(guildID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ranks, ok := s.MemberRanks[guildID]; ok {
		return len(ranks)
	}
	return 0
}

// RemoveMember removes a member from the guild hall system.
func (s *GuildHallSystem) RemoveMember(guildID string, memberID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ranks, ok := s.MemberRanks[guildID]; ok {
		delete(ranks, memberID)
	}
}

// ============================================================================
// Shared Storage System
// ============================================================================

// StoragePermission represents access level for shared storage.
type StoragePermission int

const (
	StoragePermissionNone StoragePermission = iota
	StoragePermissionView
	StoragePermissionDeposit
	StoragePermissionWithdraw
	StoragePermissionFull
)

// StorageItem represents an item in shared storage.
type StorageItem struct {
	ID          string
	OwnerID     uint64 // Player who deposited it
	Name        string
	Type        string
	Quantity    int
	DepositTime float64
	Reserved    bool // Reserved for specific player
	ReservedFor uint64
}

// StorageLog records storage activity.
type StorageLog struct {
	Timestamp float64
	MemberID  uint64
	Action    string // deposit, withdraw, reserve, unreserve
	ItemID    string
	Quantity  int
}

// SharedStorage represents a shared storage container.
type SharedStorage struct {
	ID            string
	Name          string
	OwnerID       uint64 // Player or guild owner
	OwnerType     string // "player", "guild", "household"
	Items         map[string]*StorageItem
	Capacity      int
	Members       map[uint64]StoragePermission
	AccessLog     []StorageLog
	MaxLogEntries int
}

// SharedStorageSystem manages shared storage containers.
type SharedStorageSystem struct {
	mu            sync.RWMutex
	Seed          int64
	Genre         string
	Storages      map[string]*SharedStorage
	PlayerStorage map[uint64][]string // PlayerID -> Storage IDs they have access to
	GameTime      float64
	counter       uint64
}

// NewSharedStorageSystem creates a new shared storage system.
func NewSharedStorageSystem(seed int64, genre string) *SharedStorageSystem {
	return &SharedStorageSystem{
		Seed:          seed,
		Genre:         genre,
		Storages:      make(map[string]*SharedStorage),
		PlayerStorage: make(map[uint64][]string),
	}
}

// CreateStorage creates a new shared storage container.
func (s *SharedStorageSystem) CreateStorage(storageID, name string, ownerID uint64, ownerType string, capacity int) (*SharedStorage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Storages[storageID]; exists {
		return nil, fmt.Errorf("storage already exists: %s", storageID)
	}

	storage := &SharedStorage{
		ID:            storageID,
		Name:          name,
		OwnerID:       ownerID,
		OwnerType:     ownerType,
		Items:         make(map[string]*StorageItem),
		Capacity:      capacity,
		Members:       make(map[uint64]StoragePermission),
		AccessLog:     make([]StorageLog, 0),
		MaxLogEntries: 100,
	}

	// Owner has full access
	storage.Members[ownerID] = StoragePermissionFull

	s.Storages[storageID] = storage
	s.PlayerStorage[ownerID] = append(s.PlayerStorage[ownerID], storageID)

	return storage, nil
}

// GetStorage returns a storage by ID.
func (s *SharedStorageSystem) GetStorage(storageID string) *SharedStorage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Storages[storageID]
}

// AddMember adds a member with specific permissions to a storage.
func (s *SharedStorageSystem) AddMember(storageID string, ownerID, memberID uint64, permission StoragePermission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	// Only owner can add members
	if storage.OwnerID != ownerID {
		return fmt.Errorf("only owner can add members")
	}

	storage.Members[memberID] = permission
	s.PlayerStorage[memberID] = append(s.PlayerStorage[memberID], storageID)

	return nil
}

// RemoveMember removes a member from a storage.
func (s *SharedStorageSystem) RemoveMember(storageID string, ownerID, memberID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	if storage.OwnerID != ownerID {
		return fmt.Errorf("only owner can remove members")
	}

	if memberID == ownerID {
		return fmt.Errorf("cannot remove owner")
	}

	delete(storage.Members, memberID)

	// Remove from player's storage list
	playerStorages := s.PlayerStorage[memberID]
	for i, id := range playerStorages {
		if id == storageID {
			s.PlayerStorage[memberID] = append(playerStorages[:i], playerStorages[i+1:]...)
			break
		}
	}

	return nil
}

// SetPermission updates a member's permission level.
func (s *SharedStorageSystem) SetPermission(storageID string, ownerID, memberID uint64, permission StoragePermission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	if storage.OwnerID != ownerID {
		return fmt.Errorf("only owner can change permissions")
	}

	if _, exists := storage.Members[memberID]; !exists {
		return fmt.Errorf("member not found in storage")
	}

	storage.Members[memberID] = permission
	return nil
}

// GetPermission returns a member's permission level.
func (s *SharedStorageSystem) GetPermission(storageID string, memberID uint64) StoragePermission {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return StoragePermissionNone
	}

	return storage.Members[memberID]
}

// DepositItem adds an item to shared storage.
func (s *SharedStorageSystem) DepositItem(storageID string, memberID uint64, item *StorageItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	perm := storage.Members[memberID]
	if perm < StoragePermissionDeposit {
		return fmt.Errorf("insufficient permission to deposit")
	}

	if len(storage.Items) >= storage.Capacity {
		return fmt.Errorf("storage is full")
	}

	item.OwnerID = memberID
	item.DepositTime = s.GameTime
	storage.Items[item.ID] = item

	s.logAction(storage, memberID, "deposit", item.ID, item.Quantity)

	return nil
}

// WithdrawItem removes an item from shared storage.
func (s *SharedStorageSystem) WithdrawItem(storageID string, memberID uint64, itemID string) (*StorageItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return nil, fmt.Errorf("storage not found: %s", storageID)
	}

	perm := storage.Members[memberID]
	if perm < StoragePermissionWithdraw {
		return nil, fmt.Errorf("insufficient permission to withdraw")
	}

	item := storage.Items[itemID]
	if item == nil {
		return nil, fmt.Errorf("item not found: %s", itemID)
	}

	// Check reservation
	if item.Reserved && item.ReservedFor != memberID {
		return nil, fmt.Errorf("item is reserved for another member")
	}

	delete(storage.Items, itemID)
	s.logAction(storage, memberID, "withdraw", itemID, item.Quantity)

	return item, nil
}

// ReserveItem reserves an item for a specific member.
func (s *SharedStorageSystem) ReserveItem(storageID string, memberID uint64, itemID string, forMemberID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	perm := storage.Members[memberID]
	if perm < StoragePermissionFull {
		return fmt.Errorf("insufficient permission to reserve items")
	}

	item := storage.Items[itemID]
	if item == nil {
		return fmt.Errorf("item not found: %s", itemID)
	}

	if item.Reserved {
		return fmt.Errorf("item is already reserved")
	}

	item.Reserved = true
	item.ReservedFor = forMemberID
	s.logAction(storage, memberID, "reserve", itemID, 1)

	return nil
}

// UnreserveItem removes a reservation from an item.
func (s *SharedStorageSystem) UnreserveItem(storageID string, memberID uint64, itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	perm := storage.Members[memberID]
	if perm < StoragePermissionFull {
		return fmt.Errorf("insufficient permission to unreserve items")
	}

	item := storage.Items[itemID]
	if item == nil {
		return fmt.Errorf("item not found: %s", itemID)
	}

	if !item.Reserved {
		return fmt.Errorf("item is not reserved")
	}

	item.Reserved = false
	item.ReservedFor = 0
	s.logAction(storage, memberID, "unreserve", itemID, 1)

	return nil
}

// logAction records an action in the storage log.
func (s *SharedStorageSystem) logAction(storage *SharedStorage, memberID uint64, action, itemID string, quantity int) {
	log := StorageLog{
		Timestamp: s.GameTime,
		MemberID:  memberID,
		Action:    action,
		ItemID:    itemID,
		Quantity:  quantity,
	}

	storage.AccessLog = append(storage.AccessLog, log)

	// Trim log if too long
	if len(storage.AccessLog) > storage.MaxLogEntries {
		storage.AccessLog = storage.AccessLog[1:]
	}
}

// GetItems returns all items in a storage.
func (s *SharedStorageSystem) GetItems(storageID string, memberID uint64) ([]*StorageItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return nil, fmt.Errorf("storage not found: %s", storageID)
	}

	perm := storage.Members[memberID]
	if perm < StoragePermissionView {
		return nil, fmt.Errorf("insufficient permission to view items")
	}

	items := make([]*StorageItem, 0, len(storage.Items))
	for _, item := range storage.Items {
		items = append(items, item)
	}

	return items, nil
}

// GetAccessLog returns the storage access log.
func (s *SharedStorageSystem) GetAccessLog(storageID string, memberID uint64) ([]StorageLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return nil, fmt.Errorf("storage not found: %s", storageID)
	}

	perm := storage.Members[memberID]
	if perm < StoragePermissionFull {
		return nil, fmt.Errorf("insufficient permission to view log")
	}

	return storage.AccessLog, nil
}

// GetPlayerStorages returns all storages a player has access to.
func (s *SharedStorageSystem) GetPlayerStorages(playerID uint64) []*SharedStorage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storageIDs := s.PlayerStorage[playerID]
	storages := make([]*SharedStorage, 0, len(storageIDs))

	for _, id := range storageIDs {
		if storage := s.Storages[id]; storage != nil {
			storages = append(storages, storage)
		}
	}

	return storages
}

// GetStorageCapacity returns current usage and capacity.
func (s *SharedStorageSystem) GetStorageCapacity(storageID string) (used, capacity int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return 0, 0
	}

	return len(storage.Items), storage.Capacity
}

// SetCapacity updates the storage capacity.
func (s *SharedStorageSystem) SetCapacity(storageID string, ownerID uint64, newCapacity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	if storage.OwnerID != ownerID {
		return fmt.Errorf("only owner can change capacity")
	}

	if newCapacity < len(storage.Items) {
		return fmt.Errorf("cannot reduce capacity below current usage")
	}

	storage.Capacity = newCapacity
	return nil
}

// GetMembers returns all members and their permissions.
func (s *SharedStorageSystem) GetMembers(storageID string, memberID uint64) (map[uint64]StoragePermission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return nil, fmt.Errorf("storage not found: %s", storageID)
	}

	if storage.Members[memberID] < StoragePermissionView {
		return nil, fmt.Errorf("insufficient permission")
	}

	return storage.Members, nil
}

// DeleteStorage removes a storage container.
func (s *SharedStorageSystem) DeleteStorage(storageID string, ownerID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	if storage.OwnerID != ownerID {
		return fmt.Errorf("only owner can delete storage")
	}

	if len(storage.Items) > 0 {
		return fmt.Errorf("cannot delete storage with items")
	}

	// Remove from all members' lists
	for memberID := range storage.Members {
		playerStorages := s.PlayerStorage[memberID]
		for i, id := range playerStorages {
			if id == storageID {
				s.PlayerStorage[memberID] = append(playerStorages[:i], playerStorages[i+1:]...)
				break
			}
		}
	}

	delete(s.Storages, storageID)
	return nil
}

// TransferOwnership transfers storage ownership to another player.
func (s *SharedStorageSystem) TransferOwnership(storageID string, currentOwnerID, newOwnerID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	if storage.OwnerID != currentOwnerID {
		return fmt.Errorf("only owner can transfer ownership")
	}

	// Update owner
	storage.OwnerID = newOwnerID
	storage.Members[newOwnerID] = StoragePermissionFull

	// Add to new owner's list if not already there
	found := false
	for _, id := range s.PlayerStorage[newOwnerID] {
		if id == storageID {
			found = true
			break
		}
	}
	if !found {
		s.PlayerStorage[newOwnerID] = append(s.PlayerStorage[newOwnerID], storageID)
	}

	return nil
}

// Update processes shared storage (mainly for time tracking).
func (s *SharedStorageSystem) Update(dt float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GameTime += dt
}

// StorageCount returns the number of storages.
func (s *SharedStorageSystem) StorageCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Storages)
}

// MemberCount returns the number of members in a storage.
func (s *SharedStorageSystem) MemberCount(storageID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return 0
	}
	return len(storage.Members)
}

// ItemCount returns the number of items in a storage.
func (s *SharedStorageSystem) ItemCount(storageID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storage := s.Storages[storageID]
	if storage == nil {
		return 0
	}
	return len(storage.Items)
}

// ============================================================================
// Indoor/Outdoor Detection System
// ============================================================================

// LocationType represents whether a position is indoors or outdoors.
type LocationType int

const (
	LocationOutdoor LocationType = iota
	LocationIndoor
	LocationDungeon
	LocationCave
	LocationBuilding
)

// IndoorZone represents a region that is considered "indoors".
type IndoorZone struct {
	ID       string
	Name     string
	Type     LocationType
	MinX     float64
	MaxX     float64
	MinY     float64 // Height bounds
	MaxY     float64
	MinZ     float64
	MaxZ     float64
	Priority int // Higher priority zones override lower ones
}

// IndoorDetectionSystem tracks indoor/outdoor areas for position checks.
type IndoorDetectionSystem struct {
	mu           sync.RWMutex
	Zones        map[string]*IndoorZone
	DefaultType  LocationType
	zoneCounter  uint64
	HouseManager *HouseManager
}

// NewIndoorDetectionSystem creates a new indoor detection system.
func NewIndoorDetectionSystem(hm *HouseManager) *IndoorDetectionSystem {
	return &IndoorDetectionSystem{
		Zones:        make(map[string]*IndoorZone),
		DefaultType:  LocationOutdoor,
		HouseManager: hm,
	}
}

// RegisterZone adds an indoor zone to the detection system.
func (s *IndoorDetectionSystem) RegisterZone(zone *IndoorZone) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if zone.ID == "" {
		s.zoneCounter++
		zone.ID = "zone_" + fmt.Sprintf("%d", s.zoneCounter)
	}
	s.Zones[zone.ID] = zone
}

// UnregisterZone removes an indoor zone.
func (s *IndoorDetectionSystem) UnregisterZone(zoneID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Zones, zoneID)
}

// GetLocationType returns the location type at a given position.
func (s *IndoorDetectionSystem) GetLocationType(x, y, z float64) LocationType {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestZone *IndoorZone
	for _, zone := range s.Zones {
		if s.pointInZone(x, y, z, zone) {
			if bestZone == nil || zone.Priority > bestZone.Priority {
				bestZone = zone
			}
		}
	}

	if bestZone != nil {
		return bestZone.Type
	}
	return s.DefaultType
}

// IsIndoors checks if a position is considered indoors.
func (s *IndoorDetectionSystem) IsIndoors(x, y, z float64) bool {
	locType := s.GetLocationType(x, y, z)
	return locType != LocationOutdoor
}

// IsOutdoors checks if a position is considered outdoors.
func (s *IndoorDetectionSystem) IsOutdoors(x, y, z float64) bool {
	return s.GetLocationType(x, y, z) == LocationOutdoor
}

// GetZoneAt returns the zone containing the given position.
func (s *IndoorDetectionSystem) GetZoneAt(x, y, z float64) *IndoorZone {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestZone *IndoorZone
	for _, zone := range s.Zones {
		if s.pointInZone(x, y, z, zone) {
			if bestZone == nil || zone.Priority > bestZone.Priority {
				bestZone = zone
			}
		}
	}
	return bestZone
}

// GetAllZonesAt returns all zones containing the given position.
func (s *IndoorDetectionSystem) GetAllZonesAt(x, y, z float64) []*IndoorZone {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zones := make([]*IndoorZone, 0)
	for _, zone := range s.Zones {
		if s.pointInZone(x, y, z, zone) {
			zones = append(zones, zone)
		}
	}
	return zones
}

// RegisterHouseAsZone creates an indoor zone for a player house.
func (s *IndoorDetectionSystem) RegisterHouseAsZone(house *House, width, height, depth float64) {
	zone := &IndoorZone{
		ID:       "house_" + house.ID,
		Name:     "House Interior",
		Type:     LocationIndoor,
		MinX:     house.WorldX - width/2,
		MaxX:     house.WorldX + width/2,
		MinY:     0,
		MaxY:     height,
		MinZ:     house.WorldZ - depth/2,
		MaxZ:     house.WorldZ + depth/2,
		Priority: 10,
	}
	s.RegisterZone(zone)
}

// RegisterBuildingZone creates an indoor zone for a building.
func (s *IndoorDetectionSystem) RegisterBuildingZone(id, name string, x, z, width, height, depth float64) {
	zone := &IndoorZone{
		ID:       "building_" + id,
		Name:     name,
		Type:     LocationBuilding,
		MinX:     x - width/2,
		MaxX:     x + width/2,
		MinY:     0,
		MaxY:     height,
		MinZ:     z - depth/2,
		MaxZ:     z + depth/2,
		Priority: 5,
	}
	s.RegisterZone(zone)
}

// RegisterDungeonZone creates an indoor zone for a dungeon area.
func (s *IndoorDetectionSystem) RegisterDungeonZone(id, name string, minX, maxX, minY, maxY, minZ, maxZ float64) {
	zone := &IndoorZone{
		ID:       "dungeon_" + id,
		Name:     name,
		Type:     LocationDungeon,
		MinX:     minX,
		MaxX:     maxX,
		MinY:     minY,
		MaxY:     maxY,
		MinZ:     minZ,
		MaxZ:     maxZ,
		Priority: 15,
	}
	s.RegisterZone(zone)
}

// RegisterCaveZone creates an indoor zone for a cave area.
func (s *IndoorDetectionSystem) RegisterCaveZone(id, name string, minX, maxX, minY, maxY, minZ, maxZ float64) {
	zone := &IndoorZone{
		ID:       "cave_" + id,
		Name:     name,
		Type:     LocationCave,
		MinX:     minX,
		MaxX:     maxX,
		MinY:     minY,
		MaxY:     maxY,
		MinZ:     minZ,
		MaxZ:     maxZ,
		Priority: 12,
	}
	s.RegisterZone(zone)
}

// pointInZone checks if a point is within a zone's bounds.
func (s *IndoorDetectionSystem) pointInZone(x, y, z float64, zone *IndoorZone) bool {
	return x >= zone.MinX && x <= zone.MaxX &&
		y >= zone.MinY && y <= zone.MaxY &&
		z >= zone.MinZ && z <= zone.MaxZ
}

// ZoneCount returns the number of registered zones.
func (s *IndoorDetectionSystem) ZoneCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Zones)
}

// GetLocationTypeName returns a human-readable name for a location type.
func GetLocationTypeName(locType LocationType) string {
	switch locType {
	case LocationOutdoor:
		return "Outdoors"
	case LocationIndoor:
		return "Indoors"
	case LocationDungeon:
		return "Dungeon"
	case LocationCave:
		return "Cave"
	case LocationBuilding:
		return "Building"
	default:
		return "Unknown"
	}
}
