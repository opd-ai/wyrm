// Package systems implements all ECS game systems.
package systems

import (
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// CoupState represents the state of a faction coup.
type CoupState int

const (
	// CoupStateNone means no coup is active.
	CoupStateNone CoupState = iota
	// CoupStatePlotting means conspirators are gathering support.
	CoupStatePlotting
	// CoupStateActive means the coup is underway.
	CoupStateActive
	// CoupStateSucceeded means the coup succeeded.
	CoupStateSucceeded
	// CoupStateFailed means the coup was crushed.
	CoupStateFailed
)

// FactionCoup represents an active or planned coup within a faction.
type FactionCoup struct {
	FactionID       string
	State           CoupState
	PlotStartTime   float64    // When plotting began
	CoupStartTime   float64    // When active coup began
	LeaderEntity    ecs.Entity // NPC or player leading the coup
	SupportLevel    float64    // 0.0 to 1.0 faction support
	ResistanceLevel float64    // 0.0 to 1.0 loyalist resistance
	Duration        float64    // How long the coup has been active
	InstigatorType  string     // "npc", "player", "external"
	Reason          string     // Why the coup was triggered
}

// FactionCoupSystem manages internal faction coups and leadership changes.
type FactionCoupSystem struct {
	// ActiveCoups maps faction ID to active coup
	ActiveCoups map[string]*FactionCoup
	// CoupHistory tracks past coups for narrative
	CoupHistory map[string][]*FactionCoup
	// CoupChancePerTick is the base probability of a coup starting per tick
	CoupChancePerTick float64
	// MinPlotDuration is the minimum time before a coup can go active
	MinPlotDuration float64
	// MaxPlotDuration is the maximum time a coup can stay in plotting
	MaxPlotDuration float64
	// SuccessThreshold is the support level needed for success
	SuccessThreshold float64
	// FailureThreshold is the resistance level that causes failure
	FailureThreshold float64
	// RankSystem for checking player ranks
	RankSystem *FactionRankSystem
	// PoliticsSystem for faction relations
	PoliticsSystem *FactionPoliticsSystem
	// RNG for deterministic randomness
	rng *rand.Rand
	// Genre for narrative theming
	Genre string
}

// NewFactionCoupSystem creates a new faction coup system.
func NewFactionCoupSystem(rankSystem *FactionRankSystem, politicsSystem *FactionPoliticsSystem, seed int64, genre string) *FactionCoupSystem {
	return &FactionCoupSystem{
		ActiveCoups:       make(map[string]*FactionCoup),
		CoupHistory:       make(map[string][]*FactionCoup),
		CoupChancePerTick: 0.0001, // Very low base chance
		MinPlotDuration:   120.0,  // 2 minutes minimum plotting
		MaxPlotDuration:   600.0,  // 10 minutes max plotting
		SuccessThreshold:  0.6,    // 60% support for success
		FailureThreshold:  0.7,    // 70% resistance causes failure
		RankSystem:        rankSystem,
		PoliticsSystem:    politicsSystem,
		rng:               rand.New(rand.NewSource(seed)),
		Genre:             genre,
	}
}

// Update processes faction coups each tick.
func (s *FactionCoupSystem) Update(w *ecs.World, dt float64) {
	s.processActiveCoups(w, dt)
	s.checkForNewCoups(w, dt)
}

// processActiveCoups handles ongoing coups.
func (s *FactionCoupSystem) processActiveCoups(w *ecs.World, dt float64) {
	for factionID, coup := range s.ActiveCoups {
		coup.Duration += dt
		switch coup.State {
		case CoupStatePlotting:
			s.processPlottingCoup(w, factionID, coup, dt)
		case CoupStateActive:
			s.processActiveCoup(w, factionID, coup, dt)
		case CoupStateSucceeded, CoupStateFailed:
			s.finalizeCoup(factionID, coup)
		}
	}
}

// processPlottingCoup handles coups in the plotting phase.
func (s *FactionCoupSystem) processPlottingCoup(w *ecs.World, factionID string, coup *FactionCoup, dt float64) {
	// Gather support over time
	coup.SupportLevel += s.calculateSupportGrowth(w, factionID, coup) * dt
	if coup.SupportLevel > 1.0 {
		coup.SupportLevel = 1.0
	}

	// Loyalists may detect the plot
	coup.ResistanceLevel += s.calculateResistanceGrowth(w, factionID, coup) * dt
	if coup.ResistanceLevel > 1.0 {
		coup.ResistanceLevel = 1.0
	}

	// Check if plot is discovered and crushed
	if coup.ResistanceLevel >= s.FailureThreshold {
		coup.State = CoupStateFailed
		return
	}

	// Check if ready to go active
	plotDuration := coup.Duration
	if plotDuration >= s.MinPlotDuration && coup.SupportLevel >= 0.3 {
		if coup.SupportLevel >= 0.5 || plotDuration >= s.MaxPlotDuration {
			coup.State = CoupStateActive
			coup.CoupStartTime = 0 // Would use world clock
		}
	}

	// Forced failure if plotting too long without support
	if plotDuration >= s.MaxPlotDuration && coup.SupportLevel < 0.3 {
		coup.State = CoupStateFailed
	}
}

// processActiveCoup handles coups in the active phase.
func (s *FactionCoupSystem) processActiveCoup(w *ecs.World, factionID string, coup *FactionCoup, dt float64) {
	// Active coups resolve quickly
	timeSinceActive := coup.Duration - s.MinPlotDuration

	// Support and resistance continue to shift
	coup.SupportLevel += s.calculateSupportGrowth(w, factionID, coup) * dt * 2.0
	coup.ResistanceLevel += s.calculateResistanceGrowth(w, factionID, coup) * dt * 2.0

	// Clamp values
	if coup.SupportLevel > 1.0 {
		coup.SupportLevel = 1.0
	}
	if coup.ResistanceLevel > 1.0 {
		coup.ResistanceLevel = 1.0
	}

	// Check for resolution
	if coup.SupportLevel >= s.SuccessThreshold {
		coup.State = CoupStateSucceeded
	} else if coup.ResistanceLevel >= s.FailureThreshold {
		coup.State = CoupStateFailed
	} else if timeSinceActive > 60.0 {
		// After 1 minute of active coup, force resolution
		if coup.SupportLevel > coup.ResistanceLevel {
			coup.State = CoupStateSucceeded
		} else {
			coup.State = CoupStateFailed
		}
	}
}

// calculateSupportGrowth calculates how fast support grows for a coup.
func (s *FactionCoupSystem) calculateSupportGrowth(w *ecs.World, factionID string, coup *FactionCoup) float64 {
	baseGrowth := 0.01 // 1% per second base

	// Modifier based on faction instability (would check faction health)
	instabilityModifier := 1.0

	// Modifier based on leader's influence
	leaderModifier := 1.0
	if coup.LeaderEntity != 0 && s.RankSystem != nil {
		info := s.RankSystem.GetMembershipInfo(w, coup.LeaderEntity, factionID)
		if info != nil {
			// Higher rank = more influence
			leaderModifier = 1.0 + float64(info.Rank)*0.1
		}
	}

	return baseGrowth * instabilityModifier * leaderModifier
}

// calculateResistanceGrowth calculates how fast resistance grows against a coup.
func (s *FactionCoupSystem) calculateResistanceGrowth(w *ecs.World, factionID string, coup *FactionCoup) float64 {
	baseGrowth := 0.008 // 0.8% per second base

	// Stronger factions resist better
	stabilityModifier := 1.0

	// Active coups generate more resistance
	if coup.State == CoupStateActive {
		stabilityModifier *= 1.5
	}

	return baseGrowth * stabilityModifier
}

// finalizeCoup handles coup resolution.
func (s *FactionCoupSystem) finalizeCoup(factionID string, coup *FactionCoup) {
	// Record in history
	s.CoupHistory[factionID] = append(s.CoupHistory[factionID], coup)

	// Remove from active coups
	delete(s.ActiveCoups, factionID)
}

// checkForNewCoups checks if any faction should start a coup.
func (s *FactionCoupSystem) checkForNewCoups(w *ecs.World, dt float64) {
	// Get all factions
	factions := s.collectFactions(w)

	for _, factionID := range factions {
		// Skip if already has a coup
		if _, exists := s.ActiveCoups[factionID]; exists {
			continue
		}

		// Check if coup should start
		if s.shouldStartCoup(w, factionID, dt) {
			s.StartCoup(factionID, 0, "npc", s.getCoupReason())
		}
	}
}

// collectFactions gathers all faction IDs from territories.
func (s *FactionCoupSystem) collectFactions(w *ecs.World) []string {
	factionSet := make(map[string]bool)
	for _, e := range w.Entities("FactionTerritory") {
		comp, ok := w.GetComponent(e, "FactionTerritory")
		if !ok {
			continue
		}
		territory := comp.(*components.FactionTerritory)
		factionSet[territory.FactionID] = true
	}

	factions := make([]string, 0, len(factionSet))
	for id := range factionSet {
		factions = append(factions, id)
	}
	return factions
}

// shouldStartCoup checks if a coup should spontaneously start.
func (s *FactionCoupSystem) shouldStartCoup(w *ecs.World, factionID string, dt float64) bool {
	// Very low base chance
	if s.rng.Float64() > s.CoupChancePerTick*dt {
		return false
	}

	// Modifiers based on faction state
	// (Would check faction stability, wars, economic state)
	return true
}

// getCoupReason returns a narrative reason for the coup.
func (s *FactionCoupSystem) getCoupReason() string {
	reasons := s.getGenreReasons()
	if len(reasons) == 0 {
		return "internal power struggle"
	}
	return reasons[s.rng.Intn(len(reasons))]
}

// getGenreReasons returns genre-appropriate coup reasons.
func (s *FactionCoupSystem) getGenreReasons() []string {
	switch s.Genre {
	case "fantasy":
		return []string{
			"disputed succession",
			"accusations of dark magic",
			"failed military campaign",
			"treasury embezzlement",
			"religious heresy",
			"alliance with enemies",
		}
	case "sci-fi":
		return []string{
			"failed colonial expansion",
			"resource mismanagement",
			"alien collaboration accusations",
			"military AI malfunction",
			"corporate takeover attempt",
			"terraforming disaster",
		}
	case "horror":
		return []string{
			"ritual failure",
			"possession accusations",
			"sanctuary breach",
			"madness spreading",
			"ancient evil awakening",
			"sacrifice dispute",
		}
	case "cyberpunk":
		return []string{
			"stock manipulation",
			"data breach scandal",
			"failed hostile takeover",
			"AI rights dispute",
			"black ops exposure",
			"augmentation mandate",
		}
	case "post-apocalyptic":
		return []string{
			"resource hoarding",
			"failed expedition",
			"contamination cover-up",
			"trade route dispute",
			"mutation persecution",
			"bunker access denial",
		}
	default:
		return []string{"internal power struggle"}
	}
}

// StartCoup begins a coup in a faction.
func (s *FactionCoupSystem) StartCoup(factionID string, leaderEntity ecs.Entity, instigatorType, reason string) bool {
	// Check if coup already exists
	if _, exists := s.ActiveCoups[factionID]; exists {
		return false
	}

	s.ActiveCoups[factionID] = &FactionCoup{
		FactionID:       factionID,
		State:           CoupStatePlotting,
		PlotStartTime:   0,
		LeaderEntity:    leaderEntity,
		SupportLevel:    0.1, // Start with 10% support
		ResistanceLevel: 0.1, // Start with 10% resistance
		Duration:        0,
		InstigatorType:  instigatorType,
		Reason:          reason,
	}
	return true
}

// PlayerStartCoup allows a player to initiate a coup.
func (s *FactionCoupSystem) PlayerStartCoup(w *ecs.World, playerEntity ecs.Entity, factionID string) bool {
	// Check if player is a member
	if s.RankSystem == nil {
		return false
	}
	info := s.RankSystem.GetMembershipInfo(w, playerEntity, factionID)
	if info == nil {
		return false
	}

	// Require minimum rank (5+) to lead a coup
	if info.Rank < 5 {
		return false
	}

	// Higher rank = higher starting support
	startingSupport := 0.1 + float64(info.Rank)*0.03

	if s.StartCoup(factionID, playerEntity, "player", "player-led uprising") {
		s.ActiveCoups[factionID].SupportLevel = startingSupport
		return true
	}
	return false
}

// SupportCoup allows a player to support an active coup.
func (s *FactionCoupSystem) SupportCoup(w *ecs.World, playerEntity ecs.Entity, factionID string) bool {
	coup, exists := s.ActiveCoups[factionID]
	if !exists || coup.State != CoupStatePlotting && coup.State != CoupStateActive {
		return false
	}

	// Check if player is a member
	if s.RankSystem == nil {
		return false
	}
	info := s.RankSystem.GetMembershipInfo(w, playerEntity, factionID)
	if info == nil {
		return false
	}

	// Add support based on rank
	supportBonus := 0.02 + float64(info.Rank)*0.01
	coup.SupportLevel += supportBonus
	if coup.SupportLevel > 1.0 {
		coup.SupportLevel = 1.0
	}

	return true
}

// OpposeCoup allows a player to oppose an active coup.
func (s *FactionCoupSystem) OpposeCoup(w *ecs.World, playerEntity ecs.Entity, factionID string) bool {
	coup, exists := s.ActiveCoups[factionID]
	if !exists || coup.State != CoupStatePlotting && coup.State != CoupStateActive {
		return false
	}

	// Check if player is a member
	if s.RankSystem == nil {
		return false
	}
	info := s.RankSystem.GetMembershipInfo(w, playerEntity, factionID)
	if info == nil {
		return false
	}

	// Add resistance based on rank
	resistanceBonus := 0.02 + float64(info.Rank)*0.01
	coup.ResistanceLevel += resistanceBonus
	if coup.ResistanceLevel > 1.0 {
		coup.ResistanceLevel = 1.0
	}

	return true
}

// GetCoup returns the active coup for a faction, or nil if none.
func (s *FactionCoupSystem) GetCoup(factionID string) *FactionCoup {
	return s.ActiveCoups[factionID]
}

// GetCoupHistory returns the coup history for a faction.
func (s *FactionCoupSystem) GetCoupHistory(factionID string) []*FactionCoup {
	return s.CoupHistory[factionID]
}

// GetAllActiveCoups returns all currently active coups.
func (s *FactionCoupSystem) GetAllActiveCoups() map[string]*FactionCoup {
	result := make(map[string]*FactionCoup)
	for k, v := range s.ActiveCoups {
		result[k] = v
	}
	return result
}

// IsCoupActive checks if a faction has an active coup.
func (s *FactionCoupSystem) IsCoupActive(factionID string) bool {
	coup := s.GetCoup(factionID)
	return coup != nil && (coup.State == CoupStatePlotting || coup.State == CoupStateActive)
}

// GetCoupSuccessChance returns the estimated chance of coup success (0-100%).
func (s *FactionCoupSystem) GetCoupSuccessChance(factionID string) float64 {
	coup := s.GetCoup(factionID)
	if coup == nil {
		return 0
	}

	// Calculate based on support vs resistance
	if coup.SupportLevel <= 0 {
		return 0
	}
	total := coup.SupportLevel + coup.ResistanceLevel
	if total <= 0 {
		return 50
	}
	return (coup.SupportLevel / total) * 100
}

// ForceCoupResolution immediately resolves a coup (for testing/debug).
func (s *FactionCoupSystem) ForceCoupResolution(factionID string, success bool) {
	coup := s.GetCoup(factionID)
	if coup == nil {
		return
	}
	if success {
		coup.State = CoupStateSucceeded
	} else {
		coup.State = CoupStateFailed
	}
}
