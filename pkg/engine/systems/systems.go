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
	// First, read world clock from a WorldClock entity if it exists
	for _, e := range w.Entities("WorldClock") {
		comp, ok := w.GetComponent(e, "WorldClock")
		if ok {
			clock := comp.(*components.WorldClock)
			s.WorldHour = clock.Hour
			break
		}
	}
	// Then update NPC schedules based on the hour
	for _, e := range w.Entities("Schedule") {
		comp, ok := w.GetComponent(e, "Schedule")
		if !ok {
			continue
		}
		sched := comp.(*components.Schedule)
		if activity, ok := sched.TimeSlots[s.WorldHour]; ok {
			if activity != sched.CurrentActivity {
				sched.CurrentActivity = activity
			}
		}
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
}

// NewFactionPoliticsSystem creates a new faction politics system.
func NewFactionPoliticsSystem(decayRate float64) *FactionPoliticsSystem {
	return &FactionPoliticsSystem{
		Relations: make(map[[2]string]FactionRelation),
		DecayRate: decayRate,
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
		if rep.Standings == nil {
			continue
		}
		// Decay each faction standing toward neutral (0)
		for factionID, standing := range rep.Standings {
			if standing > 0 {
				standing -= s.DecayRate * dt
				if standing < 0 {
					standing = 0
				}
			} else if standing < 0 {
				standing += s.DecayRate * dt
				if standing > 0 {
					standing = 0
				}
			}
			rep.Standings[factionID] = standing
		}
	}
}

// CrimeSystem tracks crimes, wanted levels, witnesses, and bounties.
type CrimeSystem struct {
	// DecayDelay is seconds without new crime before wanted level decreases.
	DecayDelay float64
	// BountyPerLevel is bounty added per wanted level.
	BountyPerLevel float64
	// GameTime is the current game time for tracking decay.
	GameTime float64
}

// NewCrimeSystem creates a new crime system with specified decay delay.
func NewCrimeSystem(decayDelay, bountyPerLevel float64) *CrimeSystem {
	return &CrimeSystem{
		DecayDelay:     decayDelay,
		BountyPerLevel: bountyPerLevel,
	}
}

// Update processes crime detection and bounty updates each tick.
func (s *CrimeSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	for _, e := range w.Entities("Crime") {
		comp, ok := w.GetComponent(e, "Crime")
		if !ok {
			continue
		}
		crime := comp.(*components.Crime)
		// Decay wanted level if enough time has passed since last crime
		if crime.WantedLevel > 0 {
			timeSinceCrime := s.GameTime - crime.LastCrimeTime
			if timeSinceCrime >= s.DecayDelay {
				crime.WantedLevel--
				crime.LastCrimeTime = s.GameTime
			}
		}
		// Clamp wanted level to 0-5
		if crime.WantedLevel < 0 {
			crime.WantedLevel = 0
		}
		if crime.WantedLevel > 5 {
			crime.WantedLevel = 5
		}
	}
}

// ReportCrime increases wanted level for an entity with witnesses in range.
func (s *CrimeSystem) ReportCrime(w *ecs.World, criminal ecs.Entity) {
	comp, ok := w.GetComponent(criminal, "Crime")
	if !ok {
		return
	}
	crime := comp.(*components.Crime)
	// Check if any witnesses exist
	witnesses := w.Entities("Witness", "Position")
	if len(witnesses) == 0 {
		return // No witnesses, no crime reported
	}
	// For now, any witness in world reports the crime
	// Future: add line-of-sight and distance checks
	crime.WantedLevel++
	crime.BountyAmount += s.BountyPerLevel
	crime.LastCrimeTime = s.GameTime
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
		if node.PriceTable == nil {
			node.PriceTable = make(map[string]float64)
		}
		if node.Supply == nil {
			node.Supply = make(map[string]int)
		}
		if node.Demand == nil {
			node.Demand = make(map[string]int)
		}
		// Update prices based on supply vs demand
		for itemType, basePrice := range s.BasePrices {
			supply := float64(node.Supply[itemType])
			demand := float64(node.Demand[itemType])
			// Price goes up when demand > supply, down when supply > demand
			ratio := 1.0
			if supply > 0 {
				ratio = demand / supply
			} else if demand > 0 {
				ratio = 2.0 // High demand, no supply = double price
			}
			// Apply fluctuation with clamping
			priceMod := 1.0 + (ratio-1.0)*s.PriceFluctuation
			if priceMod < 0.5 {
				priceMod = 0.5 // Minimum 50% of base price
			}
			if priceMod > 2.0 {
				priceMod = 2.0 // Maximum 200% of base price
			}
			node.PriceTable[itemType] = basePrice * priceMod
		}
		// Normalize supply and demand toward equilibrium over time
		for itemType := range node.Supply {
			target := node.Demand[itemType]
			if node.Supply[itemType] < target {
				node.Supply[itemType]++
			} else if node.Supply[itemType] > target {
				node.Supply[itemType]--
			}
		}
	}
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
		if quest.Completed {
			continue
		}
		if quest.Flags == nil {
			quest.Flags = make(map[string]bool)
		}
		// Check if current stage conditions are met
		stages, ok := s.QuestStages[quest.ID]
		if !ok {
			continue
		}
		for _, cond := range stages {
			// Only check conditions for current stage
			if cond.FromStage != quest.CurrentStage {
				continue
			}
			// Check if required flag is set
			if quest.Flags[cond.RequiredFlag] {
				if cond.Completes {
					quest.Completed = true
				} else {
					quest.CurrentStage = cond.NextStage
				}
				break
			}
		}
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

// AudioSystem drives procedural audio synthesis.
type AudioSystem struct {
	Genre string
}

// Update advances audio synthesis each tick.
func (s *AudioSystem) Update(w *ecs.World, dt float64) {
	// Future: update audio based on player position and game state
}
