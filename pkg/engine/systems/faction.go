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
