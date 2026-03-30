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
