package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// FactionRelation represents the diplomatic state between two factions.
type FactionRelation int

const (
	// RelationHostile means factions are at war.
	RelationHostile FactionRelation = -1
	// RelationNeutral means factions have no special relationship.
	RelationNeutral FactionRelation = 0
	// RelationAlly means factions are allied.
	RelationAlly FactionRelation = 1
)

// FactionPoliticsSystem handles faction relationships, wars, and treaties.
type FactionPoliticsSystem struct {
	// Relations maps faction pair (sorted alphabetically) to relation state.
	Relations map[[2]string]FactionRelation
	// DecayRate is how fast reputation drifts toward neutral per second.
	DecayRate float64
	// KillsForHostility is how many kills trigger automatic hostility.
	KillsForHostility int
	// ReputationPerKill is how much reputation is lost per faction member killed.
	ReputationPerKill float64
}

// NewFactionPoliticsSystem creates a new faction politics system.
func NewFactionPoliticsSystem(decayRate float64) *FactionPoliticsSystem {
	return &FactionPoliticsSystem{
		Relations:         make(map[[2]string]FactionRelation),
		DecayRate:         decayRate,
		KillsForHostility: 3,
		ReputationPerKill: DefaultReputationPerKill,
	}
}

// SetRelation sets the diplomatic relation between two factions.
func (s *FactionPoliticsSystem) SetRelation(f1, f2 string, rel FactionRelation) {
	key := factionPairKey(f1, f2)
	s.Relations[key] = rel
}

// GetRelation returns the diplomatic relation between two factions.
func (s *FactionPoliticsSystem) GetRelation(f1, f2 string) FactionRelation {
	key := factionPairKey(f1, f2)
	rel, ok := s.Relations[key]
	if !ok {
		return RelationNeutral
	}
	return rel
}

// factionPairKey returns a sorted pair key for faction relations.
func factionPairKey(f1, f2 string) [2]string {
	if f1 < f2 {
		return [2]string{f1, f2}
	}
	return [2]string{f2, f1}
}

// Update processes faction politics each tick: decays reputation toward neutral.
func (s *FactionPoliticsSystem) Update(w *ecs.World, dt float64) {
	s.ensureRelationsInitialized()
	for _, e := range w.Entities("Reputation") {
		s.processReputationEntity(w, e, dt)
	}
}

// ensureRelationsInitialized initializes the Relations map if nil.
func (s *FactionPoliticsSystem) ensureRelationsInitialized() {
	if s.Relations == nil {
		s.Relations = make(map[[2]string]FactionRelation)
	}
}

// processReputationEntity decays reputation standings for a single entity.
func (s *FactionPoliticsSystem) processReputationEntity(w *ecs.World, e ecs.Entity, dt float64) {
	comp, ok := w.GetComponent(e, "Reputation")
	if !ok {
		return
	}
	rep := comp.(*components.Reputation)
	s.decayReputationStandings(rep, dt)
}

// decayReputationStandings drifts all faction standings toward neutral.
func (s *FactionPoliticsSystem) decayReputationStandings(rep *components.Reputation, dt float64) {
	if rep.Standings == nil {
		return
	}
	for factionID, standing := range rep.Standings {
		rep.Standings[factionID] = s.decayStanding(standing, dt)
	}
}

// decayStanding decays a single standing value toward neutral (0).
func (s *FactionPoliticsSystem) decayStanding(standing, dt float64) float64 {
	decay := s.DecayRate * dt
	if standing > 0 {
		standing -= decay
		if standing < 0 {
			return 0
		}
	} else if standing < 0 {
		standing += decay
		if standing > 0 {
			return 0
		}
	}
	return standing
}

// ReportKill tracks a faction member kill and updates hostility if threshold reached.
// Returns true if the kill triggered hostility.
func (s *FactionPoliticsSystem) ReportKill(w *ecs.World, killerEntity ecs.Entity, victimFactionID string) bool {
	// Find the faction territory for the victim
	for _, e := range w.Entities("FactionTerritory") {
		comp, ok := w.GetComponent(e, "FactionTerritory")
		if !ok {
			continue
		}
		territory := comp.(*components.FactionTerritory)
		if territory.FactionID != victimFactionID {
			continue
		}
		// Track the kill
		if territory.KillTracker == nil {
			territory.KillTracker = make(map[uint64]int)
		}
		territory.KillTracker[uint64(killerEntity)]++
		// Reduce killer's reputation with faction
		s.applyReputationPenalty(w, killerEntity, victimFactionID)
		// Check for hostility threshold
		if territory.KillTracker[uint64(killerEntity)] >= s.KillsForHostility {
			s.setPlayerHostile(w, killerEntity, victimFactionID)
			return true
		}
		return false
	}
	return false
}

// applyReputationPenalty reduces an entity's reputation with a faction.
func (s *FactionPoliticsSystem) applyReputationPenalty(w *ecs.World, entity ecs.Entity, factionID string) {
	comp, ok := w.GetComponent(entity, "Reputation")
	if !ok {
		return
	}
	rep := comp.(*components.Reputation)
	if rep.Standings == nil {
		rep.Standings = make(map[string]float64)
	}
	rep.Standings[factionID] += s.ReputationPerKill
	// Clamp to minimum
	if rep.Standings[factionID] < -100 {
		rep.Standings[factionID] = -100
	}
}

// setPlayerHostile marks a player as hostile with a faction.
func (s *FactionPoliticsSystem) setPlayerHostile(w *ecs.World, entity ecs.Entity, factionID string) {
	comp, ok := w.GetComponent(entity, "Reputation")
	if !ok {
		return
	}
	rep := comp.(*components.Reputation)
	if rep.Standings == nil {
		rep.Standings = make(map[string]float64)
	}
	// Set to maximum hostility
	rep.Standings[factionID] = -100
}

// SignTreaty establishes a peace treaty between player and faction, reducing hostility.
func (s *FactionPoliticsSystem) SignTreaty(w *ecs.World, playerEntity ecs.Entity, factionID string) bool {
	comp, ok := w.GetComponent(playerEntity, "Reputation")
	if !ok {
		return false
	}
	rep := comp.(*components.Reputation)
	if rep.Standings == nil {
		rep.Standings = make(map[string]float64)
	}
	// Reset hostility to neutral
	rep.Standings[factionID] = 0
	// Clear kill count
	for _, e := range w.Entities("FactionTerritory") {
		tComp, ok := w.GetComponent(e, "FactionTerritory")
		if !ok {
			continue
		}
		territory := tComp.(*components.FactionTerritory)
		if territory.FactionID == factionID && territory.KillTracker != nil {
			delete(territory.KillTracker, uint64(playerEntity))
		}
	}
	return true
}

// ============================================================================
// Dynamic Faction Wars System
// ============================================================================

// WarState represents the state of a war between factions.
type WarState int

const (
	WarStateNone WarState = iota
	WarStateTension
	WarStateActive
	WarStateCeasefire
)

// FactionWar represents an active conflict between two factions.
type FactionWar struct {
	Faction1       string
	Faction2       string
	State          WarState
	StartTime      float64 // Game time when war started
	Duration       float64 // How long the war has been active
	Faction1Score  int     // Territory captured, battles won
	Faction2Score  int
	CeasefireTimer float64 // Time until ceasefire expires
}

// DynamicFactionWarSystem manages dynamic faction wars.
type DynamicFactionWarSystem struct {
	wars              map[[2]string]*FactionWar
	tensionThreshold  float64 // When tension triggers war
	warDecayRate      float64 // How fast wars wind down
	ceasefireDuration float64 // How long ceasefires last
	politicsSystem    *FactionPoliticsSystem
}

// NewDynamicFactionWarSystem creates a new dynamic faction war system.
func NewDynamicFactionWarSystem(politicsSystem *FactionPoliticsSystem) *DynamicFactionWarSystem {
	return &DynamicFactionWarSystem{
		wars:              make(map[[2]string]*FactionWar),
		tensionThreshold:  0.7,
		warDecayRate:      0.01,
		ceasefireDuration: 300.0, // 5 minutes game time
		politicsSystem:    politicsSystem,
	}
}

// Update processes faction wars each tick.
func (s *DynamicFactionWarSystem) Update(w *ecs.World, dt float64) {
	s.updateExistingWars(dt)
	s.checkForNewWars(w)
	s.processWarEffects(w, dt)
}

// updateExistingWars updates the duration and state of active wars.
func (s *DynamicFactionWarSystem) updateExistingWars(dt float64) {
	for key, war := range s.wars {
		war.Duration += dt
		s.updateWarState(key, war, dt)
	}
}

// updateWarState handles state transitions for a single war.
func (s *DynamicFactionWarSystem) updateWarState(key [2]string, war *FactionWar, dt float64) {
	switch war.State {
	case WarStateActive:
		s.checkVictoryConditions(war)
	case WarStateCeasefire:
		s.processCeasefire(key, war, dt)
	case WarStateTension:
		s.checkEscalation(war)
	}
}

// checkVictoryConditions transitions to ceasefire if a side wins.
func (s *DynamicFactionWarSystem) checkVictoryConditions(war *FactionWar) {
	if war.Faction1Score >= 10 || war.Faction2Score >= 10 {
		war.State = WarStateCeasefire
		war.CeasefireTimer = s.ceasefireDuration
	}
}

// processCeasefire counts down and ends wars.
func (s *DynamicFactionWarSystem) processCeasefire(key [2]string, war *FactionWar, dt float64) {
	war.CeasefireTimer -= dt
	if war.CeasefireTimer <= 0 {
		delete(s.wars, key)
		if s.politicsSystem != nil {
			s.politicsSystem.SetRelation(war.Faction1, war.Faction2, RelationNeutral)
		}
	}
}

// checkEscalation escalates tension to active war after threshold duration.
func (s *DynamicFactionWarSystem) checkEscalation(war *FactionWar) {
	if war.Duration > 60.0 {
		war.State = WarStateActive
		if s.politicsSystem != nil {
			s.politicsSystem.SetRelation(war.Faction1, war.Faction2, RelationHostile)
		}
	}
}

// checkForNewWars looks for faction pairs that should start wars.
func (s *DynamicFactionWarSystem) checkForNewWars(w *ecs.World) {
	if s.politicsSystem == nil {
		return
	}

	factions := s.collectFactionIDs(w)
	s.checkHostilePairs(factions)
}

// collectFactionIDs gathers all unique faction IDs from territories.
func (s *DynamicFactionWarSystem) collectFactionIDs(w *ecs.World) []string {
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

// checkHostilePairs examines faction pairs and declares wars for hostile relations.
func (s *DynamicFactionWarSystem) checkHostilePairs(factions []string) {
	for i := 0; i < len(factions); i++ {
		for j := i + 1; j < len(factions); j++ {
			s.checkAndDeclareWar(factions[i], factions[j])
		}
	}
}

// checkAndDeclareWar declares war between two factions if they are hostile.
func (s *DynamicFactionWarSystem) checkAndDeclareWar(f1, f2 string) {
	key := factionPairKey(f1, f2)
	if _, exists := s.wars[key]; exists {
		return
	}
	if s.politicsSystem.GetRelation(f1, f2) == RelationHostile {
		s.DeclareWar(f1, f2)
	}
}

// processWarEffects applies war effects to the world.
func (s *DynamicFactionWarSystem) processWarEffects(w *ecs.World, dt float64) {
	for _, war := range s.wars {
		if war.State != WarStateActive {
			continue
		}

		// Process territory disputes
		s.processWarTerritories(w, war, dt)
	}
}

// processWarTerritories handles territory changes during wars.
func (s *DynamicFactionWarSystem) processWarTerritories(w *ecs.World, war *FactionWar, dt float64) {
	for _, e := range w.Entities("FactionTerritory") {
		s.processWarTerritory(w, e, war, dt)
	}
}

// processWarTerritory handles territory changes for a single territory.
func (s *DynamicFactionWarSystem) processWarTerritory(w *ecs.World, e ecs.Entity, war *FactionWar, dt float64) {
	comp, ok := w.GetComponent(e, "FactionTerritory")
	if !ok {
		return
	}
	territory := comp.(*components.FactionTerritory)

	if !s.isTerritoryInWar(territory, war) {
		return
	}

	territory.ControlLevel -= dt * 0.01
	if territory.ControlLevel < 0 {
		territory.ControlLevel = 0
	}

	if territory.ControlLevel < 0.2 {
		s.captureTerritory(territory, war)
	}
}

// isTerritoryInWar checks if territory belongs to one of the warring factions.
func (s *DynamicFactionWarSystem) isTerritoryInWar(territory *components.FactionTerritory, war *FactionWar) bool {
	return territory.FactionID == war.Faction1 || territory.FactionID == war.Faction2
}

// captureTerritory awards points and resets control level.
func (s *DynamicFactionWarSystem) captureTerritory(territory *components.FactionTerritory, war *FactionWar) {
	if territory.FactionID == war.Faction1 {
		war.Faction2Score++
	} else {
		war.Faction1Score++
	}
	territory.ControlLevel = 0.5
}

// DeclareWar starts a war between two factions.
func (s *DynamicFactionWarSystem) DeclareWar(faction1, faction2 string) {
	key := factionPairKey(faction1, faction2)
	if _, exists := s.wars[key]; exists {
		return // Already at war
	}

	s.wars[key] = &FactionWar{
		Faction1:  faction1,
		Faction2:  faction2,
		State:     WarStateTension,
		StartTime: 0,
		Duration:  0,
	}

	if s.politicsSystem != nil {
		s.politicsSystem.SetRelation(faction1, faction2, RelationHostile)
	}
}

// ForceWar immediately starts an active war.
func (s *DynamicFactionWarSystem) ForceWar(faction1, faction2 string) {
	key := factionPairKey(faction1, faction2)

	s.wars[key] = &FactionWar{
		Faction1:  faction1,
		Faction2:  faction2,
		State:     WarStateActive,
		StartTime: 0,
		Duration:  0,
	}

	if s.politicsSystem != nil {
		s.politicsSystem.SetRelation(faction1, faction2, RelationHostile)
	}
}

// RequestCeasefire attempts to end a war with a ceasefire.
func (s *DynamicFactionWarSystem) RequestCeasefire(faction1, faction2 string) bool {
	key := factionPairKey(faction1, faction2)
	war, exists := s.wars[key]
	if !exists || war.State != WarStateActive {
		return false
	}

	war.State = WarStateCeasefire
	war.CeasefireTimer = s.ceasefireDuration
	return true
}

// GetWar returns the war between two factions, or nil if none.
func (s *DynamicFactionWarSystem) GetWar(faction1, faction2 string) *FactionWar {
	key := factionPairKey(faction1, faction2)
	return s.wars[key]
}

// GetAllWars returns all active wars.
func (s *DynamicFactionWarSystem) GetAllWars() []*FactionWar {
	wars := make([]*FactionWar, 0, len(s.wars))
	for _, war := range s.wars {
		wars = append(wars, war)
	}
	return wars
}

// IsAtWar checks if two factions are at war.
func (s *DynamicFactionWarSystem) IsAtWar(faction1, faction2 string) bool {
	war := s.GetWar(faction1, faction2)
	return war != nil && war.State == WarStateActive
}

// GetWarCount returns the number of active wars.
func (s *DynamicFactionWarSystem) GetWarCount() int {
	count := 0
	for _, war := range s.wars {
		if war.State == WarStateActive {
			count++
		}
	}
	return count
}
