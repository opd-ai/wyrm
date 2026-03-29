// Package systems contains all ECS system implementations.
package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

// ChunkLoader defines the interface for loading and unloading chunks.
type ChunkLoader interface {
	GetChunk(x, y int) *chunk.Chunk
}

// WorldChunkSystem manages loading and unloading of world chunks.
type WorldChunkSystem struct {
	Manager   ChunkLoader
	chunkSize int
}

// NewWorldChunkSystem creates a new chunk system with the given manager.
func NewWorldChunkSystem(manager ChunkLoader, chunkSize int) *WorldChunkSystem {
	return &WorldChunkSystem{
		Manager:   manager,
		chunkSize: chunkSize,
	}
}

// Update loads chunks around entities with Position components.
func (s *WorldChunkSystem) Update(w *ecs.World, dt float64) {
	if s.Manager == nil {
		return
	}
	for _, e := range w.Entities("Position") {
		comp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := comp.(*components.Position)
		chunkX := int(pos.X) / s.chunkSize
		chunkY := int(pos.Y) / s.chunkSize
		// Load the 3x3 chunk window around the entity
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				_ = s.Manager.GetChunk(chunkX+dx, chunkY+dy)
			}
		}
	}
}

// NPCScheduleSystem drives NPC daily activity cycles.
type NPCScheduleSystem struct {
	// WorldHour is updated externally by WorldClockSystem
	WorldHour int
}

// Update processes NPC schedules based on the current world hour.
func (s *NPCScheduleSystem) Update(w *ecs.World, dt float64) {
	s.syncWorldHour(w)
	s.updateNPCSchedules(w)
}

// syncWorldHour reads the world clock from a WorldClock entity.
func (s *NPCScheduleSystem) syncWorldHour(w *ecs.World) {
	for _, e := range w.Entities("WorldClock") {
		comp, ok := w.GetComponent(e, "WorldClock")
		if ok {
			clock := comp.(*components.WorldClock)
			s.WorldHour = clock.Hour
			return
		}
	}
}

// updateNPCSchedules updates all NPC schedules based on the current hour.
func (s *NPCScheduleSystem) updateNPCSchedules(w *ecs.World) {
	for _, e := range w.Entities("Schedule") {
		s.updateEntitySchedule(w, e)
	}
}

// updateEntitySchedule updates a single entity's current activity.
func (s *NPCScheduleSystem) updateEntitySchedule(w *ecs.World, e ecs.Entity) {
	comp, ok := w.GetComponent(e, "Schedule")
	if !ok {
		return
	}
	sched := comp.(*components.Schedule)
	if activity, ok := sched.TimeSlots[s.WorldHour]; ok && activity != sched.CurrentActivity {
		sched.CurrentActivity = activity
	}
}

// WorldClockSystem advances the in-game time.
type WorldClockSystem struct {
	// DefaultHourLength is seconds per game hour if no clock entity exists.
	DefaultHourLength float64
}

// NewWorldClockSystem creates a new world clock system.
func NewWorldClockSystem(hourLength float64) *WorldClockSystem {
	return &WorldClockSystem{DefaultHourLength: hourLength}
}

// Update advances the game clock each tick.
func (s *WorldClockSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("WorldClock") {
		comp, ok := w.GetComponent(e, "WorldClock")
		if !ok {
			continue
		}
		clock := comp.(*components.WorldClock)
		if clock.HourLength <= 0 {
			clock.HourLength = s.DefaultHourLength
		}
		clock.TimeAccum += dt
		if clock.TimeAccum >= clock.HourLength {
			clock.TimeAccum -= clock.HourLength
			clock.Hour++
			if clock.Hour >= 24 {
				clock.Hour = 0
				clock.Day++
			}
		}
	}
}

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
		ReputationPerKill: -25.0,
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
	if s.Relations == nil {
		s.Relations = make(map[[2]string]FactionRelation)
	}
	for _, e := range w.Entities("Reputation") {
		comp, ok := w.GetComponent(e, "Reputation")
		if !ok {
			continue
		}
		rep := comp.(*components.Reputation)
		s.decayReputationStandings(rep, dt)
	}
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

// CrimeSystem tracks crimes, wanted levels, witnesses, and bounties.
type CrimeSystem struct {
	// DecayDelay is seconds without new crime before wanted level decreases.
	DecayDelay float64
	// BountyPerLevel is bounty added per wanted level.
	BountyPerLevel float64
	// GameTime is the current game time for tracking decay.
	GameTime float64
	// WitnessRange is the maximum distance for witnesses to observe crimes.
	WitnessRange float64
}

// NewCrimeSystem creates a new crime system with specified decay delay.
func NewCrimeSystem(decayDelay, bountyPerLevel float64) *CrimeSystem {
	return &CrimeSystem{
		DecayDelay:     decayDelay,
		BountyPerLevel: bountyPerLevel,
		WitnessRange:   50.0, // Default witness range
	}
}

// Update processes crime detection and bounty updates each tick.
func (s *CrimeSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	for _, e := range w.Entities("Crime") {
		s.processCrimeEntity(w, e)
	}
}

// processCrimeEntity updates a single entity's crime state.
func (s *CrimeSystem) processCrimeEntity(w *ecs.World, e ecs.Entity) {
	comp, ok := w.GetComponent(e, "Crime")
	if !ok {
		return
	}
	crime := comp.(*components.Crime)

	if crime.InJail {
		return
	}

	s.decayWantedLevel(crime)
	s.clampWantedLevel(crime)
}

// decayWantedLevel decreases wanted level if enough time has passed.
func (s *CrimeSystem) decayWantedLevel(crime *components.Crime) {
	if crime.WantedLevel <= 0 {
		return
	}
	timeSinceCrime := s.GameTime - crime.LastCrimeTime
	if timeSinceCrime >= s.DecayDelay {
		crime.WantedLevel--
		crime.LastCrimeTime = s.GameTime
	}
}

// clampWantedLevel constrains wanted level to valid range 0-5.
func (s *CrimeSystem) clampWantedLevel(crime *components.Crime) {
	if crime.WantedLevel < 0 {
		crime.WantedLevel = 0
	}
	if crime.WantedLevel > 5 {
		crime.WantedLevel = 5
	}
}

// ReportCrime increases wanted level for an entity if witnessed.
func (s *CrimeSystem) ReportCrime(w *ecs.World, criminal ecs.Entity) {
	comp, ok := w.GetComponent(criminal, "Crime")
	if !ok {
		return
	}
	crime := comp.(*components.Crime)
	criminalPos := s.getEntityPosition(w, criminal)
	if !s.isWitnessed(w, criminalPos) {
		return
	}
	crime.WantedLevel++
	crime.BountyAmount += s.BountyPerLevel
	crime.LastCrimeTime = s.GameTime
}

// isWitnessed checks if any witness can observe the given position.
func (s *CrimeSystem) isWitnessed(w *ecs.World, pos [3]float64) bool {
	for _, witness := range w.Entities("Witness", "Position") {
		if s.canWitnessObserve(w, witness, pos) {
			return true
		}
	}
	return false
}

// canWitnessObserve checks if a specific witness can observe a position.
func (s *CrimeSystem) canWitnessObserve(w *ecs.World, witness ecs.Entity, pos [3]float64) bool {
	wComp, ok := w.GetComponent(witness, "Witness")
	if !ok {
		return false
	}
	wState := wComp.(*components.Witness)
	if !wState.CanReport {
		return false
	}
	witnessPos := s.getEntityPosition(w, witness)
	return s.canWitnessSee(pos, witnessPos)
}

// canWitnessSee checks if a witness can see the crime location (LOS + range).
func (s *CrimeSystem) canWitnessSee(crimePos, witnessPos [3]float64) bool {
	// Simple distance-based check (future: actual line-of-sight)
	dx := crimePos[0] - witnessPos[0]
	dy := crimePos[1] - witnessPos[1]
	distSq := dx*dx + dy*dy
	return distSq <= s.WitnessRange*s.WitnessRange
}

// getEntityPosition returns an entity's position or zero if not found.
func (s *CrimeSystem) getEntityPosition(w *ecs.World, e ecs.Entity) [3]float64 {
	comp, ok := w.GetComponent(e, "Position")
	if !ok {
		return [3]float64{}
	}
	pos := comp.(*components.Position)
	return [3]float64{pos.X, pos.Y, pos.Z}
}

// PayBounty allows an entity to pay off their bounty and reset wanted level.
func (s *CrimeSystem) PayBounty(w *ecs.World, entity ecs.Entity) bool {
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return false
	}
	crime := comp.(*components.Crime)
	// Could integrate with Economy system for actual payment
	crime.WantedLevel = 0
	crime.BountyAmount = 0
	crime.InJail = false
	return true
}

// GoToJail sends an entity to jail (sets flag, would teleport in full impl).
func (s *CrimeSystem) GoToJail(w *ecs.World, entity ecs.Entity, jailTime float64) bool {
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return false
	}
	crime := comp.(*components.Crime)
	crime.InJail = true
	crime.JailReleaseTime = s.GameTime + jailTime
	// Note: bounty remains, wanted level clears after jail time
	crime.WantedLevel = 0
	return true
}

// CheckJailRelease checks if entity should be released from jail.
func (s *CrimeSystem) CheckJailRelease(w *ecs.World) {
	for _, e := range w.Entities("Crime") {
		comp, ok := w.GetComponent(e, "Crime")
		if !ok {
			continue
		}
		crime := comp.(*components.Crime)
		if crime.InJail && s.GameTime >= crime.JailReleaseTime {
			crime.InJail = false
		}
	}
}

// EconomySystem manages supply, demand, and pricing across city nodes.
type EconomySystem struct {
	// BasePrices maps item type to base price before supply/demand adjustment.
	BasePrices map[string]float64
	// PriceFluctuation controls how much supply/demand affects price (0-1).
	PriceFluctuation float64
	// NormalizationRate is how fast supply/demand drift to equilibrium per second.
	NormalizationRate float64
}

// NewEconomySystem creates a new economy system.
func NewEconomySystem(fluctuation, normRate float64) *EconomySystem {
	return &EconomySystem{
		BasePrices:        make(map[string]float64),
		PriceFluctuation:  fluctuation,
		NormalizationRate: normRate,
	}
}

// SetBasePrice sets the base price for an item type.
func (s *EconomySystem) SetBasePrice(itemType string, price float64) {
	if s.BasePrices == nil {
		s.BasePrices = make(map[string]float64)
	}
	s.BasePrices[itemType] = price
}

// Update processes economic simulation each tick.
func (s *EconomySystem) Update(w *ecs.World, dt float64) {
	if s.BasePrices == nil {
		s.BasePrices = make(map[string]float64)
	}
	for _, e := range w.Entities("EconomyNode") {
		comp, ok := w.GetComponent(e, "EconomyNode")
		if !ok {
			continue
		}
		node := comp.(*components.EconomyNode)
		s.initializeNodeMaps(node)
		s.updateNodePrices(node)
		s.normalizeSupplyDemand(node)
	}
}

// initializeNodeMaps ensures all economy node maps are initialized.
func (s *EconomySystem) initializeNodeMaps(node *components.EconomyNode) {
	if node.PriceTable == nil {
		node.PriceTable = make(map[string]float64)
	}
	if node.Supply == nil {
		node.Supply = make(map[string]int)
	}
	if node.Demand == nil {
		node.Demand = make(map[string]int)
	}
}

// calculatePriceModifier computes the price modifier based on supply and demand.
func (s *EconomySystem) calculatePriceModifier(supply, demand int) float64 {
	ratio := 1.0
	if supply > 0 {
		ratio = float64(demand) / float64(supply)
	} else if demand > 0 {
		ratio = 2.0 // High demand, no supply = double price
	}
	priceMod := 1.0 + (ratio-1.0)*s.PriceFluctuation
	return clampFloat(priceMod, 0.5, 2.0)
}

// updateNodePrices updates all item prices based on supply vs demand.
func (s *EconomySystem) updateNodePrices(node *components.EconomyNode) {
	for itemType, basePrice := range s.BasePrices {
		priceMod := s.calculatePriceModifier(node.Supply[itemType], node.Demand[itemType])
		node.PriceTable[itemType] = basePrice * priceMod
	}
}

// normalizeSupplyDemand drifts supply toward demand over time.
func (s *EconomySystem) normalizeSupplyDemand(node *components.EconomyNode) {
	for itemType := range node.Supply {
		target := node.Demand[itemType]
		if node.Supply[itemType] < target {
			node.Supply[itemType]++
		} else if node.Supply[itemType] > target {
			node.Supply[itemType]--
		}
	}
}

// SellItem processes a sale of items to a vendor, increasing supply.
func (s *EconomySystem) SellItem(w *ecs.World, vendor ecs.Entity, itemType string, quantity int) float64 {
	comp, ok := w.GetComponent(vendor, "EconomyNode")
	if !ok {
		return 0
	}
	node := comp.(*components.EconomyNode)
	s.initializeNodeMaps(node)
	// Calculate price before supply increase
	currentPrice := s.GetBuyPrice(w, vendor, itemType)
	// Increase supply (vendor now has more stock)
	node.Supply[itemType] += quantity
	// Recalculate price after supply change
	s.updateNodePrices(node)
	return currentPrice * float64(quantity)
}

// BuyItem processes a purchase of items from a vendor, decreasing supply.
func (s *EconomySystem) BuyItem(w *ecs.World, vendor ecs.Entity, itemType string, quantity int) float64 {
	comp, ok := w.GetComponent(vendor, "EconomyNode")
	if !ok {
		return 0
	}
	node := comp.(*components.EconomyNode)
	s.initializeNodeMaps(node)
	// Calculate price
	currentPrice := s.GetBuyPrice(w, vendor, itemType)
	// Decrease supply (vendor sold stock)
	if node.Supply[itemType] >= quantity {
		node.Supply[itemType] -= quantity
	} else {
		node.Supply[itemType] = 0
	}
	// Recalculate price after supply change
	s.updateNodePrices(node)
	return currentPrice * float64(quantity)
}

// GetBuyPrice returns the current buying price at a vendor for an item.
func (s *EconomySystem) GetBuyPrice(w *ecs.World, vendor ecs.Entity, itemType string) float64 {
	comp, ok := w.GetComponent(vendor, "EconomyNode")
	if !ok {
		return 0
	}
	node := comp.(*components.EconomyNode)
	if node.PriceTable == nil {
		return s.BasePrices[itemType]
	}
	price, ok := node.PriceTable[itemType]
	if !ok {
		return s.BasePrices[itemType]
	}
	return price
}

// clampFloat clamps a value between min and max.
func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// CombatSystem handles combat resolution and damage.
type CombatSystem struct{}

// Update processes combat resolution each tick.
func (s *CombatSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Health") {
		comp, ok := w.GetComponent(e, "Health")
		if !ok {
			continue
		}
		health := comp.(*components.Health)
		// Clamp health to max
		if health.Current > health.Max {
			health.Current = health.Max
		}
	}
}

// VehicleSystem manages vehicle movement and physics.
type VehicleSystem struct{}

// Update processes vehicle physics each tick.
func (s *VehicleSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Vehicle", "Position") {
		vComp, ok := w.GetComponent(e, "Vehicle")
		if !ok {
			continue
		}
		pComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		vehicle := vComp.(*components.Vehicle)
		pos := pComp.(*components.Position)
		// Apply vehicle movement based on speed and direction
		if vehicle.Fuel > 0 && vehicle.Speed > 0 {
			pos.X += math.Cos(vehicle.Direction) * vehicle.Speed * dt
			pos.Y += math.Sin(vehicle.Direction) * vehicle.Speed * dt
			vehicle.Fuel -= vehicle.Speed * dt * 0.01
			if vehicle.Fuel < 0 {
				vehicle.Fuel = 0
			}
		}
	}
}

// QuestSystem manages quest state, branching, and consequence flags.
type QuestSystem struct {
	// QuestStages maps quest ID to list of stage conditions.
	QuestStages map[string][]QuestStageCondition
}

// QuestStageCondition defines what must be true to advance a quest stage.
type QuestStageCondition struct {
	RequiredFlag string // Flag that must be true to advance
	FromStage    int    // Stage this condition applies from
	NextStage    int    // Stage to advance to
	Completes    bool   // If true, this transition completes the quest
	LocksBranch  string // Branch ID to lock when this transition is taken
	BranchID     string // Branch ID this condition belongs to (blocked if locked)
}

// NewQuestSystem creates a new quest system.
func NewQuestSystem() *QuestSystem {
	return &QuestSystem{
		QuestStages: make(map[string][]QuestStageCondition),
	}
}

// DefineQuest adds stage conditions for a quest.
func (s *QuestSystem) DefineQuest(questID string, stages []QuestStageCondition) {
	if s.QuestStages == nil {
		s.QuestStages = make(map[string][]QuestStageCondition)
	}
	s.QuestStages[questID] = stages
}

// Update processes quest state transitions each tick.
func (s *QuestSystem) Update(w *ecs.World, dt float64) {
	if s.QuestStages == nil {
		s.QuestStages = make(map[string][]QuestStageCondition)
	}
	for _, e := range w.Entities("Quest") {
		comp, ok := w.GetComponent(e, "Quest")
		if !ok {
			continue
		}
		quest := comp.(*components.Quest)
		s.processQuestStage(quest)
	}
}

// processQuestStage checks and advances a single quest's stage.
func (s *QuestSystem) processQuestStage(quest *components.Quest) {
	if quest.Completed {
		return
	}
	if quest.Flags == nil {
		quest.Flags = make(map[string]bool)
	}
	stages, ok := s.QuestStages[quest.ID]
	if !ok {
		return
	}
	s.checkStageConditions(quest, stages)
}

// checkStageConditions evaluates stage conditions and advances the quest.
func (s *QuestSystem) checkStageConditions(quest *components.Quest, stages []QuestStageCondition) {
	for _, cond := range stages {
		if cond.FromStage != quest.CurrentStage {
			continue
		}
		// Skip if this branch is locked
		if cond.BranchID != "" && quest.IsBranchLocked(cond.BranchID) {
			continue
		}
		if quest.Flags[cond.RequiredFlag] {
			s.advanceQuest(quest, cond)
			break
		}
	}
}

// advanceQuest moves the quest to the next stage or completes it.
func (s *QuestSystem) advanceQuest(quest *components.Quest, cond QuestStageCondition) {
	// Lock the competing branch if specified
	if cond.LocksBranch != "" {
		quest.LockBranch(cond.LocksBranch)
	}
	if cond.Completes {
		quest.Completed = true
	} else {
		quest.CurrentStage = cond.NextStage
	}
}

// WeatherSystem controls dynamic weather and environmental hazards.
type WeatherSystem struct {
	CurrentWeather  string
	TimeAccum       float64
	WeatherDuration float64 // Duration in seconds before weather change
	Genre           string  // Affects available weather types
	weatherIndex    int     // For deterministic cycling
}

// NewWeatherSystem creates a new weather system.
func NewWeatherSystem(genre string, duration float64) *WeatherSystem {
	return &WeatherSystem{
		Genre:           genre,
		WeatherDuration: duration,
		CurrentWeather:  "clear",
	}
}

// getWeatherPool returns genre-appropriate weather types.
func (s *WeatherSystem) getWeatherPool() []string {
	switch s.Genre {
	case "fantasy":
		return []string{"clear", "cloudy", "rain", "fog", "thunderstorm"}
	case "sci-fi":
		return []string{"clear", "dust", "ion_storm", "radiation_burst"}
	case "horror":
		return []string{"fog", "overcast", "rain", "blood_moon", "mist"}
	case "cyberpunk":
		return []string{"smog", "acid_rain", "clear", "neon_haze"}
	case "post-apocalyptic":
		return []string{"dust_storm", "clear", "ash_fall", "radiation_fog", "scorching"}
	default:
		return []string{"clear", "cloudy", "rain", "fog"}
	}
}

// Update advances weather simulation each tick.
func (s *WeatherSystem) Update(w *ecs.World, dt float64) {
	s.TimeAccum += dt
	if s.CurrentWeather == "" {
		s.CurrentWeather = "clear"
	}
	if s.WeatherDuration <= 0 {
		s.WeatherDuration = 300.0 // Default 5 minutes per weather
	}
	// Change weather after duration
	if s.TimeAccum >= s.WeatherDuration {
		s.TimeAccum = 0
		pool := s.getWeatherPool()
		s.weatherIndex = (s.weatherIndex + 1) % len(pool)
		s.CurrentWeather = pool[s.weatherIndex]
	}
}

// RenderSystem handles first-person raycasting and Ebitengine draw calls.
type RenderSystem struct {
	PlayerEntity ecs.Entity
}

// Update prepares render state based on player position each tick.
func (s *RenderSystem) Update(w *ecs.World, dt float64) {
	// Get player position for camera
	if s.PlayerEntity != 0 {
		_, _ = w.GetComponent(s.PlayerEntity, "Position")
	}
}

// AudioSystem drives procedural audio synthesis and spatial audio.
type AudioSystem struct {
	Genre string
	// CombatDetectionRange is the distance to detect combat for music intensity.
	CombatDetectionRange float64
	// AmbientUpdateInterval is seconds between ambient sound checks.
	AmbientUpdateInterval float64
	// timeAccum tracks time for periodic ambient updates.
	timeAccum float64
}

// NewAudioSystem creates a new audio system with default settings.
func NewAudioSystem(genre string) *AudioSystem {
	return &AudioSystem{
		Genre:                 genre,
		CombatDetectionRange:  50.0,
		AmbientUpdateInterval: 5.0,
		timeAccum:             0,
	}
}

// Update advances audio synthesis based on player position and game state.
func (s *AudioSystem) Update(w *ecs.World, dt float64) {
	s.timeAccum += dt

	// Find the audio listener (typically the player)
	listenerPos, listenerFound := s.findListenerPosition(w)
	if !listenerFound {
		return
	}

	// Update audio state component if it exists
	s.updateAudioState(w, listenerPos)

	// Process audio sources for spatial audio calculations
	s.processSpatialAudio(w, listenerPos)

	// Periodically update ambient sounds
	if s.timeAccum >= s.AmbientUpdateInterval {
		s.timeAccum = 0
		s.updateAmbientSounds(w, listenerPos)
	}
}

// findListenerPosition locates the audio listener entity and returns its position.
func (s *AudioSystem) findListenerPosition(w *ecs.World) ([2]float64, bool) {
	for _, e := range w.Entities("AudioListener", "Position") {
		listenerComp, ok := w.GetComponent(e, "AudioListener")
		if !ok {
			continue
		}
		listener := listenerComp.(*components.AudioListener)
		if !listener.Enabled {
			continue
		}
		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)
		return [2]float64{pos.X, pos.Y}, true
	}
	return [2]float64{}, false
}

// updateAudioState updates the AudioState component with current conditions.
func (s *AudioSystem) updateAudioState(w *ecs.World, listenerPos [2]float64) {
	// Calculate combat intensity based on nearby hostile entities
	combatIntensity := s.calculateCombatIntensity(w, listenerPos)

	// Update all AudioState components
	for _, e := range w.Entities("AudioState") {
		comp, ok := w.GetComponent(e, "AudioState")
		if !ok {
			continue
		}
		state := comp.(*components.AudioState)
		state.CombatIntensity = combatIntensity
		state.LastPositionX = listenerPos[0]
		state.LastPositionY = listenerPos[1]
	}
}

// calculateCombatIntensity returns 0.0-1.0 based on nearby hostile entities.
func (s *AudioSystem) calculateCombatIntensity(w *ecs.World, listenerPos [2]float64) float64 {
	maxHostiles := 10 // Cap for intensity calculation
	hostileCount := s.countNearbyHostiles(w, listenerPos)

	if hostileCount >= maxHostiles {
		return 1.0
	}
	return float64(hostileCount) / float64(maxHostiles)
}

// countNearbyHostiles counts hostile entities within detection range.
func (s *AudioSystem) countNearbyHostiles(w *ecs.World, listenerPos [2]float64) int {
	hostileCount := 0
	rangeSquared := s.CombatDetectionRange * s.CombatDetectionRange

	for _, e := range w.Entities("Health", "Position", "Faction") {
		if s.isEntityHostileAndNearby(w, e, listenerPos, rangeSquared) {
			hostileCount++
		}
	}
	return hostileCount
}

// isEntityHostileAndNearby checks if an entity is both hostile and within range.
func (s *AudioSystem) isEntityHostileAndNearby(w *ecs.World, e ecs.Entity, listenerPos [2]float64, rangeSquared float64) bool {
	posComp, ok := w.GetComponent(e, "Position")
	if !ok {
		return false
	}
	pos := posComp.(*components.Position)

	if !s.isWithinRange(pos, listenerPos, rangeSquared) {
		return false
	}

	return s.isEntityHostile(w, e)
}

// isWithinRange checks if a position is within squared range of listener.
func (s *AudioSystem) isWithinRange(pos *components.Position, listenerPos [2]float64, rangeSquared float64) bool {
	dx := pos.X - listenerPos[0]
	dy := pos.Y - listenerPos[1]
	distSq := dx*dx + dy*dy
	return distSq <= rangeSquared
}

// isEntityHostile checks if an entity has hostile faction reputation.
func (s *AudioSystem) isEntityHostile(w *ecs.World, e ecs.Entity) bool {
	factionComp, ok := w.GetComponent(e, "Faction")
	if !ok {
		return false
	}
	faction := factionComp.(*components.Faction)
	return faction.Reputation < -50
}

// processSpatialAudio calculates volume/pan for audio sources based on distance.
func (s *AudioSystem) processSpatialAudio(w *ecs.World, listenerPos [2]float64) {
	for _, e := range w.Entities("AudioSource", "Position") {
		sourceComp, ok := w.GetComponent(e, "AudioSource")
		if !ok {
			continue
		}
		source := sourceComp.(*components.AudioSource)
		if !source.Playing {
			continue
		}

		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)

		// Calculate distance-based attenuation
		dx := pos.X - listenerPos[0]
		dy := pos.Y - listenerPos[1]
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist >= source.Range {
			// Out of range - effectively muted
			continue
		}

		// Linear falloff for now (could be improved to inverse-square)
		attenuation := 1.0 - (dist / source.Range)
		_ = source.Volume * attenuation // Calculated volume for playback
	}
}

// updateAmbientSounds selects appropriate ambient sound based on environment.
func (s *AudioSystem) updateAmbientSounds(w *ecs.World, listenerPos [2]float64) {
	// Determine ambient type based on location and world state
	ambientType := s.selectAmbientType(w, listenerPos)

	// Update AudioState with new ambient
	for _, e := range w.Entities("AudioState") {
		comp, ok := w.GetComponent(e, "AudioState")
		if !ok {
			continue
		}
		state := comp.(*components.AudioState)
		state.CurrentAmbient = ambientType
	}
}

// selectAmbientType chooses the ambient sound type based on location.
func (s *AudioSystem) selectAmbientType(w *ecs.World, listenerPos [2]float64) string {
	// Check if in a city (near EconomyNode entities)
	for _, e := range w.Entities("EconomyNode", "Position") {
		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)
		dx := pos.X - listenerPos[0]
		dy := pos.Y - listenerPos[1]
		if dx*dx+dy*dy < 100*100 { // Within 100 units of a city node
			return s.getCityAmbient()
		}
	}

	// Default to wilderness ambient
	return s.getWildernessAmbient()
}

// getCityAmbient returns the genre-appropriate city ambient sound type.
func (s *AudioSystem) getCityAmbient() string {
	switch s.Genre {
	case "fantasy":
		return "city_medieval"
	case "sci-fi":
		return "city_station"
	case "horror":
		return "city_abandoned"
	case "cyberpunk":
		return "city_neon"
	case "post-apocalyptic":
		return "city_ruins"
	default:
		return "city_generic"
	}
}

// getWildernessAmbient returns the genre-appropriate wilderness ambient sound type.
func (s *AudioSystem) getWildernessAmbient() string {
	switch s.Genre {
	case "fantasy":
		return "wilderness_forest"
	case "sci-fi":
		return "wilderness_alien"
	case "horror":
		return "wilderness_dark"
	case "cyberpunk":
		return "wilderness_industrial"
	case "post-apocalyptic":
		return "wilderness_wasteland"
	default:
		return "wilderness_generic"
	}
}

// SkillProgressionSystem manages skill experience gain and level-ups.
type SkillProgressionSystem struct {
	// XPPerLevel is the base XP required per level (scales with level).
	XPPerLevel float64
	// LevelCap is the maximum skill level.
	LevelCap int
}

// NewSkillProgressionSystem creates a new skill progression system.
func NewSkillProgressionSystem(xpPerLevel float64, levelCap int) *SkillProgressionSystem {
	if levelCap <= 0 {
		levelCap = 100
	}
	if xpPerLevel <= 0 {
		xpPerLevel = 100
	}
	return &SkillProgressionSystem{
		XPPerLevel: xpPerLevel,
		LevelCap:   levelCap,
	}
}

// Update processes skill experience and level-ups each tick.
func (s *SkillProgressionSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Skills") {
		comp, ok := w.GetComponent(e, "Skills")
		if !ok {
			continue
		}
		skills := comp.(*components.Skills)
		s.processSkillProgression(skills)
	}
}

// processSkillProgression checks all skills for level-up conditions.
func (s *SkillProgressionSystem) processSkillProgression(skills *components.Skills) {
	if skills.Levels == nil || skills.Experience == nil {
		return
	}
	for skillID, xp := range skills.Experience {
		level := skills.Levels[skillID]
		if level >= s.LevelCap {
			continue
		}
		xpRequired := s.calculateXPRequired(level)
		if xp >= xpRequired {
			skills.Levels[skillID] = level + 1
			skills.Experience[skillID] = xp - xpRequired
		}
	}
}

// calculateXPRequired computes XP needed for the next level.
// Uses a simple scaling formula: base * (1 + level * 0.1)
func (s *SkillProgressionSystem) calculateXPRequired(currentLevel int) float64 {
	return s.XPPerLevel * (1.0 + float64(currentLevel)*0.1)
}

// GrantSkillXP adds experience to a skill for an entity.
func (s *SkillProgressionSystem) GrantSkillXP(w *ecs.World, entity ecs.Entity, skillID string, xp float64) bool {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return false
	}
	skills := comp.(*components.Skills)
	if skills.Experience == nil {
		skills.Experience = make(map[string]float64)
	}
	if _, exists := skills.Levels[skillID]; !exists {
		return false
	}
	skills.Experience[skillID] += xp
	return true
}

// GetSkillLevel returns the current level of a skill for an entity.
func (s *SkillProgressionSystem) GetSkillLevel(w *ecs.World, entity ecs.Entity, skillID string) int {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return 0
	}
	skills := comp.(*components.Skills)
	if skills.Levels == nil {
		return 0
	}
	return skills.Levels[skillID]
}
