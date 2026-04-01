// Package housing provides player housing and guild territory management.
// This file contains rent collection and home upgrade systems.
package housing

import (
	"fmt"
	"sync"

	"github.com/opd-ai/wyrm/pkg/seedutil"
)

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
	Genre           string
	Properties      map[string]*RentalProperty
	Tenants         map[uint64]*TenantInfo
	OwnerProperties map[uint64][]string // Owner -> property IDs
	OwnerIncome     map[uint64]float64  // Accumulated rental income
	GameTime        float64
	rng             *seedutil.PseudoRandom
	PaymentPeriod   float64 // Hours between rent payments
	EvictionGrace   int     // Days before eviction starts
	DepositMultiple float64 // Deposit as multiple of rent
}

// NewRentCollectionSystem creates a new rent collection system.
func NewRentCollectionSystem(seed int64, genre string) *RentCollectionSystem {
	return &RentCollectionSystem{
		Genre:           genre,
		Properties:      make(map[string]*RentalProperty),
		Tenants:         make(map[uint64]*TenantInfo),
		OwnerProperties: make(map[uint64][]string),
		OwnerIncome:     make(map[uint64]float64),
		rng:             seedutil.NewPseudoRandom(seed),
		PaymentPeriod:   720.0, // 30 days (720 hours)
		EvictionGrace:   7,     // 7 days grace period
		DepositMultiple: 2.0,   // 2 months deposit
	}
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *RentCollectionSystem) pseudoRandom() float64 {
	return s.rng.Float64()
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
	if s.GameTime >= prop.NextPayment {
		s.processRentPayment(prop)
	}

	s.degradePropertyCondition(prop, hours)
	s.updateTenantSatisfaction(prop, hours)
}

// processRentPayment handles the rent payment cycle for a property.
func (s *RentCollectionSystem) processRentPayment(prop *RentalProperty) {
	tenant := s.Tenants[prop.TenantID]
	if tenant != nil && s.tenantPays(tenant, prop) {
		s.recordSuccessfulPayment(prop, tenant)
	} else {
		s.recordMissedPayment(prop, tenant)
	}
}

// recordSuccessfulPayment updates state for an on-time rent payment.
func (s *RentCollectionSystem) recordSuccessfulPayment(prop *RentalProperty, tenant *TenantInfo) {
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
}

// recordMissedPayment updates state for a missed rent payment.
func (s *RentCollectionSystem) recordMissedPayment(prop *RentalProperty, tenant *TenantInfo) {
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

// degradePropertyCondition reduces property condition over time.
func (s *RentCollectionSystem) degradePropertyCondition(prop *RentalProperty, hours float64) {
	prop.Condition = clampFloat64(prop.Condition-0.0001*hours, 0.1, 1.0)
}

// updateTenantSatisfaction adjusts tenant satisfaction based on property condition.
func (s *RentCollectionSystem) updateTenantSatisfaction(prop *RentalProperty, hours float64) {
	tenant := s.Tenants[prop.TenantID]
	if tenant != nil && prop.Condition < 0.5 {
		tenant.Satisfaction = clampFloat64(tenant.Satisfaction-0.001*hours, 0, 1)
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
	Genre             string
	AvailableUpgrades map[string]*HomeUpgrade  // All possible upgrades
	HomeUpgrades      map[string]*UpgradedHome // HouseID -> upgrades
	GameTime          float64
	rng               *seedutil.PseudoRandom
}

// NewHomeUpgradeSystem creates a new home upgrade system.
func NewHomeUpgradeSystem(seed int64, genre string) *HomeUpgradeSystem {
	sys := &HomeUpgradeSystem{
		Genre:             genre,
		AvailableUpgrades: make(map[string]*HomeUpgrade),
		HomeUpgrades:      make(map[string]*UpgradedHome),
		rng:               seedutil.NewPseudoRandom(seed),
	}
	sys.initializeUpgrades()
	return sys
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *HomeUpgradeSystem) pseudoRandom() float64 {
	return s.rng.Float64()
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

	if s.isUpgradeInstalled(home, upgradeID) {
		return false, "upgrade already installed"
	}

	if reason := s.checkPrerequisites(home, upgrade); reason != "" {
		return false, reason
	}

	return true, ""
}

// isUpgradeInstalled checks if an upgrade is already installed.
func (s *HomeUpgradeSystem) isUpgradeInstalled(home *UpgradedHome, upgradeID string) bool {
	_, ok := home.Upgrades[upgradeID]
	return ok
}

// checkPrerequisites verifies all upgrade prerequisites are met.
func (s *HomeUpgradeSystem) checkPrerequisites(home *UpgradedHome, upgrade *HomeUpgrade) string {
	for _, prereq := range upgrade.Prerequisites {
		if _, ok := home.Upgrades[prereq]; !ok {
			if prereqUpgrade := s.AvailableUpgrades[prereq]; prereqUpgrade != nil {
				return fmt.Sprintf("requires %s", prereqUpgrade.Name)
			}
			return "missing prerequisite"
		}
	}
	return ""
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

// sumUpgradeBonus calculates total bonus from completed upgrades using the given extractor.
func (s *HomeUpgradeSystem) sumUpgradeBonus(houseID string, extractBonus func(*UpgradeEffects) float64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	home := s.HomeUpgrades[houseID]
	if home == nil {
		return 0
	}

	bonus := 0.0
	for _, upgrade := range home.Upgrades {
		if upgrade.Status == UpgradeStatusCompleted {
			bonus += extractBonus(&upgrade.Effects)
		}
	}
	return bonus
}

// GetTenantBonus calculates total tenant bonus for rental properties.
func (s *HomeUpgradeSystem) GetTenantBonus(houseID string) float64 {
	return s.sumUpgradeBonus(houseID, func(e *UpgradeEffects) float64 { return e.TenantBonus })
}

// GetCraftingBonus calculates total crafting bonus for a home.
func (s *HomeUpgradeSystem) GetCraftingBonus(houseID string) float64 {
	return s.sumUpgradeBonus(houseID, func(e *UpgradeEffects) float64 { return e.CraftingBonus })
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
