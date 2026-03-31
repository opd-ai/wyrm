// Package housing provides player housing and guild territory management.
// This file contains guild hall, shared storage, and indoor detection systems.
package housing

import (
	"fmt"
	"sync"

	"github.com/opd-ai/wyrm/pkg/util"
)

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
	Genre       string
	GuildHalls  map[string]*GuildHall           // GuildID -> Hall
	MemberRanks map[string]map[uint64]GuildRank // GuildID -> MemberID -> Rank
	Permissions map[GuildRank][]string          // Rank -> allowed actions
	GameTime    float64
	rng         *util.PseudoRandom
}

// NewGuildHallSystem creates a new guild hall system.
func NewGuildHallSystem(seed int64, genre string) *GuildHallSystem {
	sys := &GuildHallSystem{
		Genre:       genre,
		GuildHalls:  make(map[string]*GuildHall),
		MemberRanks: make(map[string]map[uint64]GuildRank),
		Permissions: make(map[GuildRank][]string),
		rng:         util.NewPseudoRandom(seed),
	}
	sys.initializePermissions()
	return sys
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *GuildHallSystem) pseudoRandom() float64 {
	return s.rng.Float64()
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

// getHallAndFacility validates hall exists, permission is granted, and facility exists.
// Caller must hold the lock.
func (s *GuildHallSystem) getHallAndFacility(guildID string, memberID uint64, facilityID, permission string) (*GuildHall, *GuildFacility, error) {
	hall := s.GuildHalls[guildID]
	if hall == nil {
		return nil, nil, fmt.Errorf("guild hall not found")
	}
	if !s.hasPermissionLocked(guildID, memberID, permission) {
		return nil, nil, fmt.Errorf("insufficient rank to %s", permission)
	}
	facility := hall.Facilities[facilityID]
	if facility == nil {
		return nil, nil, fmt.Errorf("facility not found")
	}
	return hall, facility, nil
}

// UseFacility marks a facility as being used.
func (s *GuildHallSystem) UseFacility(guildID string, memberID uint64, facilityID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, facility, err := s.getHallAndFacility(guildID, memberID, facilityID, "use_facilities")
	if err != nil {
		return err
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

	hall, facility, err := s.getHallAndFacility(guildID, memberID, facilityID, "manage_facilities")
	if err != nil {
		return err
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
		if s.isDailyUpkeepDue(hall, dayLength) {
			s.processDailyUpkeep(hall)
		}
	}
}

// isDailyUpkeepDue checks if daily upkeep should be processed.
func (s *GuildHallSystem) isDailyUpkeepDue(hall *GuildHall, dayLength float64) bool {
	return s.GameTime-hall.LastUpkeep >= dayLength
}

// processDailyUpkeep handles daily upkeep payment for a guild hall.
func (s *GuildHallSystem) processDailyUpkeep(hall *GuildHall) {
	hall.LastUpkeep = s.GameTime

	if hall.TreasuryGold >= hall.Upkeep {
		s.processSuccessfulPayment(hall)
	} else {
		s.processFailedPayment(hall)
	}
}

// processSuccessfulPayment handles successful upkeep payment.
func (s *GuildHallSystem) processSuccessfulPayment(hall *GuildHall) {
	hall.TreasuryGold -= hall.Upkeep
	hall.InDebt = false
	hall.DebtDays = 0
}

// processFailedPayment handles failed upkeep payment.
func (s *GuildHallSystem) processFailedPayment(hall *GuildHall) {
	hall.InDebt = true
	hall.DebtDays++
	if hall.DebtDays >= 3 {
		s.disableFacilities(hall)
	}
}

// disableFacilities marks all facilities as non-operational.
func (s *GuildHallSystem) disableFacilities(hall *GuildHall) {
	for _, facility := range hall.Facilities {
		facility.Operational = false
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

// getStorageWithPermission returns a storage and validates member permission.
// Caller must hold the lock. Returns nil storage and error on failure.
func (s *SharedStorageSystem) getStorageWithPermission(storageID string, memberID uint64, required StoragePermission, action string) (*SharedStorage, error) {
	storage := s.Storages[storageID]
	if storage == nil {
		return nil, fmt.Errorf("storage not found: %s", storageID)
	}
	if storage.Members[memberID] < required {
		return nil, fmt.Errorf("insufficient permission to %s", action)
	}
	return storage, nil
}

// getStorageItem returns an item from storage, validating it exists.
// Caller must hold the lock.
func (s *SharedStorageSystem) getStorageItem(storage *SharedStorage, itemID string) (*StorageItem, error) {
	item := storage.Items[itemID]
	if item == nil {
		return nil, fmt.Errorf("item not found: %s", itemID)
	}
	return item, nil
}

// DepositItem adds an item to shared storage.
func (s *SharedStorageSystem) DepositItem(storageID string, memberID uint64, item *StorageItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	storage, err := s.getStorageWithPermission(storageID, memberID, StoragePermissionDeposit, "deposit")
	if err != nil {
		return err
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

	storage, err := s.getStorageWithPermission(storageID, memberID, StoragePermissionWithdraw, "withdraw")
	if err != nil {
		return nil, err
	}

	item, err := s.getStorageItem(storage, itemID)
	if err != nil {
		return nil, err
	}

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

	storage, err := s.getStorageWithPermission(storageID, memberID, StoragePermissionFull, "reserve items")
	if err != nil {
		return err
	}

	item, err := s.getStorageItem(storage, itemID)
	if err != nil {
		return err
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

	storage, err := s.getStorageWithPermission(storageID, memberID, StoragePermissionFull, "unreserve items")
	if err != nil {
		return err
	}

	item, err := s.getStorageItem(storage, itemID)
	if err != nil {
		return err
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
	if err := s.validateStorageDeletion(storage, ownerID); err != nil {
		return err
	}

	s.removeStorageFromMembers(storage, storageID)
	delete(s.Storages, storageID)
	return nil
}

// validateStorageDeletion checks if storage can be deleted.
func (s *SharedStorageSystem) validateStorageDeletion(storage *SharedStorage, ownerID uint64) error {
	if storage == nil {
		return fmt.Errorf("storage not found")
	}
	if storage.OwnerID != ownerID {
		return fmt.Errorf("only owner can delete storage")
	}
	if len(storage.Items) > 0 {
		return fmt.Errorf("cannot delete storage with items")
	}
	return nil
}

// removeStorageFromMembers removes storage ID from all member player lists.
func (s *SharedStorageSystem) removeStorageFromMembers(storage *SharedStorage, storageID string) {
	for memberID := range storage.Members {
		s.removeStorageFromPlayer(memberID, storageID)
	}
}

// removeStorageFromPlayer removes a storage ID from a player's storage list.
func (s *SharedStorageSystem) removeStorageFromPlayer(playerID uint64, storageID string) {
	playerStorages := s.PlayerStorage[playerID]
	for i, id := range playerStorages {
		if id == storageID {
			s.PlayerStorage[playerID] = append(playerStorages[:i], playerStorages[i+1:]...)
			return
		}
	}
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
