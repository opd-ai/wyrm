// Package systems implements all ECS game systems.
package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// FactionRankSystem manages player faction membership and rank progression.
type FactionRankSystem struct {
	// RankTitles maps genre -> faction type -> rank -> title
	RankTitles map[string]map[string][]string
	// RankXPThresholds defines XP required per rank (1-10)
	RankXPThresholds []int
	// Genre is the current game genre
	Genre string
}

// NewFactionRankSystem creates a new faction rank system.
func NewFactionRankSystem(genre string) *FactionRankSystem {
	sys := &FactionRankSystem{
		RankTitles: make(map[string]map[string][]string),
		RankXPThresholds: []int{
			0,     // Rank 0: Not a member
			100,   // Rank 1: Entry
			250,   // Rank 2
			500,   // Rank 3
			1000,  // Rank 4
			2000,  // Rank 5
			4000,  // Rank 6
			7000,  // Rank 7
			11000, // Rank 8
			16000, // Rank 9
			25000, // Rank 10: Maximum
		},
		Genre: genre,
	}
	sys.initializeRankTitles()
	return sys
}

// initializeRankTitles sets up genre-appropriate rank titles.
func (s *FactionRankSystem) initializeRankTitles() {
	s.RankTitles["fantasy"] = map[string][]string{
		"guild":     {"Outsider", "Initiate", "Apprentice", "Journeyman", "Artisan", "Expert", "Master", "Elder", "Grandmaster", "Exalted", "Legendary"},
		"military":  {"Civilian", "Recruit", "Private", "Corporal", "Sergeant", "Lieutenant", "Captain", "Major", "Colonel", "General", "Marshal"},
		"religious": {"Heathen", "Acolyte", "Initiate", "Adept", "Priest", "High Priest", "Bishop", "Archbishop", "Cardinal", "Pope", "Saint"},
	}
	s.RankTitles["sci-fi"] = map[string][]string{
		"corporation": {"Outsider", "Intern", "Associate", "Analyst", "Specialist", "Manager", "Director", "VP", "SVP", "C-Suite", "Chairman"},
		"military":    {"Civilian", "Cadet", "Ensign", "Lieutenant", "Commander", "Captain", "Commodore", "Admiral", "Fleet Admiral", "Grand Admiral", "Supreme Commander"},
		"scientific":  {"Layperson", "Lab Assistant", "Technician", "Researcher", "Scientist", "Senior Scientist", "Lead Researcher", "Professor", "Director", "Chief Science Officer", "Luminary"},
	}
	s.RankTitles["horror"] = map[string][]string{
		"cult":     {"Uninitiated", "Seeker", "Acolyte", "Disciple", "Devoted", "Fanatic", "Zealot", "Prophet", "High Priest", "Voice", "Vessel"},
		"survivor": {"Stranger", "Scavenger", "Scout", "Guard", "Defender", "Warden", "Veteran", "Elder", "Leader", "Patriarch", "Legend"},
	}
	s.RankTitles["cyberpunk"] = map[string][]string{
		"megacorp": {"Nobody", "Intern", "Wage Slave", "Specialist", "Manager", "Executive", "VP", "Director", "CXO", "Board Member", "CEO"},
		"gang":     {"Outsider", "Runner", "Soldier", "Enforcer", "Lieutenant", "Underboss", "Captain", "Warlord", "Boss", "Kingpin", "Legend"},
		"hacker":   {"Script Kiddie", "Newbie", "Hacker", "Cracker", "Phreaker", "Elite", "Ghost", "Architect", "Netrunner", "Daemon", "Singularity"},
	}
	s.RankTitles["post-apocalyptic"] = map[string][]string{
		"tribe":  {"Outsider", "Initiate", "Scout", "Hunter", "Warrior", "Elder", "War Chief", "Shaman", "Chieftain", "High Chief", "Legend"},
		"raider": {"Meat", "Prospect", "Blood", "Marauder", "Reaver", "Wardog", "Captain", "Warlord", "Overlord", "Tyrant", "Apocalypse King"},
		"trader": {"Unknown", "Peddler", "Merchant", "Trader", "Dealer", "Broker", "Factor", "Guild Master", "Magnate", "Baron", "Trade Emperor"},
	}
}

// Update processes faction rank progression each tick.
func (s *FactionRankSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("FactionMembership") {
		s.processEntity(w, e)
	}
}

// processEntity checks and applies rank promotions for an entity.
func (s *FactionRankSystem) processEntity(w *ecs.World, e ecs.Entity) {
	comp, ok := w.GetComponent(e, "FactionMembership")
	if !ok {
		return
	}
	membership := comp.(*components.FactionMembership)
	if membership.Memberships == nil {
		return
	}
	for _, info := range membership.Memberships {
		s.checkPromotion(info)
	}
}

// checkPromotion checks if a faction membership qualifies for promotion.
// Loops to allow multiple rank ups if enough XP was gained at once.
func (s *FactionRankSystem) checkPromotion(info *components.FactionMemberInfo) {
	for {
		if info.Rank >= 10 || info.IsExalted {
			return // Already at max rank
		}
		nextRank := info.Rank + 1
		if nextRank >= len(s.RankXPThresholds) {
			return
		}
		if info.XP >= s.RankXPThresholds[nextRank] {
			s.promoteRank(info)
		} else {
			return // Not enough XP for next rank
		}
	}
}

// promoteRank increases the player's rank in a faction.
func (s *FactionRankSystem) promoteRank(info *components.FactionMemberInfo) {
	info.Rank++
	info.RankTitle = s.GetRankTitle(info.FactionID, info.Rank)
	if info.Rank < len(s.RankXPThresholds) {
		info.XPToNext = s.RankXPThresholds[info.Rank]
	}
	if info.Rank >= 10 {
		info.IsExalted = true
	}
}

// GetRankTitle returns the title for a rank in a faction.
func (s *FactionRankSystem) GetRankTitle(factionType string, rank int) string {
	genreTitles, ok := s.RankTitles[s.Genre]
	if !ok {
		genreTitles = s.RankTitles["fantasy"]
	}
	titles, ok := genreTitles[factionType]
	if !ok {
		// Default to first available faction type
		for _, t := range genreTitles {
			titles = t
			break
		}
	}
	if titles == nil || rank >= len(titles) {
		return "Member"
	}
	return titles[rank]
}

// JoinFaction adds a player to a faction at rank 1.
func (s *FactionRankSystem) JoinFaction(w *ecs.World, entity ecs.Entity, factionID, factionType string, gameTime float64) bool {
	comp, ok := w.GetComponent(entity, "FactionMembership")
	if !ok {
		// Create new membership component
		membership := &components.FactionMembership{
			Memberships: make(map[string]*components.FactionMemberInfo),
		}
		w.AddComponent(entity, membership)
		comp = membership
	}
	membership := comp.(*components.FactionMembership)
	if membership.Memberships == nil {
		membership.Memberships = make(map[string]*components.FactionMemberInfo)
	}
	// Check if already a member
	if _, exists := membership.Memberships[factionID]; exists {
		return false
	}
	// Add membership
	membership.Memberships[factionID] = &components.FactionMemberInfo{
		FactionID:  factionID,
		Rank:       1,
		RankTitle:  s.GetRankTitle(factionType, 1),
		Reputation: 0,
		JoinedAt:   gameTime,
		XP:         0,
		XPToNext:   s.RankXPThresholds[1],
	}
	return true
}

// LeaveFaction removes a player from a faction.
func (s *FactionRankSystem) LeaveFaction(w *ecs.World, entity ecs.Entity, factionID string) bool {
	comp, ok := w.GetComponent(entity, "FactionMembership")
	if !ok {
		return false
	}
	membership := comp.(*components.FactionMembership)
	if membership.Memberships == nil {
		return false
	}
	if _, exists := membership.Memberships[factionID]; !exists {
		return false
	}
	delete(membership.Memberships, factionID)
	return true
}

// AddXP adds experience points to a player's faction membership.
func (s *FactionRankSystem) AddXP(w *ecs.World, entity ecs.Entity, factionID string, xp int) {
	comp, ok := w.GetComponent(entity, "FactionMembership")
	if !ok {
		return
	}
	membership := comp.(*components.FactionMembership)
	if membership.Memberships == nil {
		return
	}
	info, exists := membership.Memberships[factionID]
	if !exists {
		return
	}
	info.XP += xp
	s.checkPromotion(info)
}

// AddQuestCompletion increments quest completion count and adds bonus XP.
func (s *FactionRankSystem) AddQuestCompletion(w *ecs.World, entity ecs.Entity, factionID string) {
	comp, ok := w.GetComponent(entity, "FactionMembership")
	if !ok {
		return
	}
	membership := comp.(*components.FactionMembership)
	if membership.Memberships == nil {
		return
	}
	info, exists := membership.Memberships[factionID]
	if !exists {
		return
	}
	info.QuestsCompleted++
	// Award bonus XP for quest completion
	bonusXP := 50 + (info.Rank * 10)
	info.XP += bonusXP
	s.checkPromotion(info)
}

// AddDonation records a donation and adds XP based on amount.
func (s *FactionRankSystem) AddDonation(w *ecs.World, entity ecs.Entity, factionID string, amount int) {
	comp, ok := w.GetComponent(entity, "FactionMembership")
	if !ok {
		return
	}
	membership := comp.(*components.FactionMembership)
	if membership.Memberships == nil {
		return
	}
	info, exists := membership.Memberships[factionID]
	if !exists {
		return
	}
	info.DonationTotal += amount
	// Award XP for donations (1 XP per 10 gold)
	bonusXP := amount / 10
	if bonusXP > 0 {
		info.XP += bonusXP
		s.checkPromotion(info)
	}
}

// GetMembershipInfo returns detailed membership info for a player and faction.
func (s *FactionRankSystem) GetMembershipInfo(w *ecs.World, entity ecs.Entity, factionID string) *components.FactionMemberInfo {
	comp, ok := w.GetComponent(entity, "FactionMembership")
	if !ok {
		return nil
	}
	membership := comp.(*components.FactionMembership)
	return membership.GetMembership(factionID)
}

// GetAllMemberships returns all faction memberships for a player.
func (s *FactionRankSystem) GetAllMemberships(w *ecs.World, entity ecs.Entity) map[string]*components.FactionMemberInfo {
	comp, ok := w.GetComponent(entity, "FactionMembership")
	if !ok {
		return nil
	}
	membership := comp.(*components.FactionMembership)
	return membership.Memberships
}

// CanAccessRankContent checks if a player meets the rank requirement for content.
func (s *FactionRankSystem) CanAccessRankContent(w *ecs.World, entity ecs.Entity, factionID string, requiredRank int) bool {
	comp, ok := w.GetComponent(entity, "FactionMembership")
	if !ok {
		return false
	}
	membership := comp.(*components.FactionMembership)
	return membership.GetRank(factionID) >= requiredRank
}

// GetProgressToNextRank returns progress percentage (0-100) to next rank.
func (s *FactionRankSystem) GetProgressToNextRank(w *ecs.World, entity ecs.Entity, factionID string) float64 {
	info := s.GetMembershipInfo(w, entity, factionID)
	if info == nil || info.Rank >= 10 {
		return 100.0
	}
	currentThreshold := s.RankXPThresholds[info.Rank]
	nextThreshold := s.RankXPThresholds[info.Rank+1]
	xpInRank := info.XP - currentThreshold
	xpNeeded := nextThreshold - currentThreshold
	if xpNeeded <= 0 {
		return 100.0
	}
	return float64(xpInRank) / float64(xpNeeded) * 100.0
}
