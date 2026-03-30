package systems

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// City event type constants.
const (
	EventTypeFestival      = "festival"
	EventTypeMarketDay     = "market_day"
	EventTypePlague        = "plague"
	EventTypeRiot          = "riot"
	EventTypeSiege         = "siege"
	EventTypeCelebration   = "celebration"
	EventTypeMartialLaw    = "martial_law"
	EventTypeTradeCaravan  = "trade_caravan"
	EventTypeReligiousRite = "religious_rite"
	EventTypeTournament    = "tournament"
	EventTypeBlackout      = "blackout"
	EventTypeHacking       = "hacking"
	EventTypeRadiation     = "radiation"
	EventTypeMutantAttack  = "mutant_attack"
	EventTypeCultRitual    = "cult_ritual"
	EventTypeHaunting      = "haunting"
)

// CityEventSystem manages dynamic city events based on world state.
type CityEventSystem struct {
	// Genre affects available event types.
	Genre string
	// Seed for deterministic event generation.
	Seed int64
	// MinEventInterval is minimum hours between events.
	MinEventInterval float64
	// MaxActiveEvents limits simultaneous events per city.
	MaxActiveEvents int

	// Internal state
	lastEventTime    float64
	rng              *rand.Rand
	eventProbability float64
}

// NewCityEventSystem creates a new city event system.
func NewCityEventSystem(genre string, seed int64) *CityEventSystem {
	return &CityEventSystem{
		Genre:            genre,
		Seed:             seed,
		MinEventInterval: CityEventMinInterval,
		MaxActiveEvents:  CityEventMaxActive,
		eventProbability: CityEventBaseProbability,
		rng:              rand.New(rand.NewSource(seed)),
	}
}

// Update processes city events each tick.
func (s *CityEventSystem) Update(w *ecs.World, dt float64) {
	currentTime := s.getCurrentGameTime(w)
	s.updateActiveEvents(w, currentTime)
	s.checkForNewEvents(w, currentTime)
}

// getCurrentGameTime retrieves the current game time from WorldClock.
func (s *CityEventSystem) getCurrentGameTime(w *ecs.World) float64 {
	for _, e := range w.Entities("WorldClock") {
		clockComp, ok := w.GetComponent(e, "WorldClock")
		if !ok {
			continue
		}
		clock := clockComp.(*components.WorldClock)
		// Calculate total hours: days * 24 + hour + time accumulated as fraction of an hour
		hourFraction := 0.0
		if clock.HourLength > 0 {
			hourFraction = clock.TimeAccum / clock.HourLength
		}
		return float64(clock.Day*24) + float64(clock.Hour) + hourFraction
	}
	return 0
}

// updateActiveEvents processes ongoing events.
func (s *CityEventSystem) updateActiveEvents(w *ecs.World, currentTime float64) {
	for _, e := range w.Entities("CityEvent") {
		eventComp, ok := w.GetComponent(e, "CityEvent")
		if !ok {
			continue
		}
		event := eventComp.(*components.CityEvent)

		if !event.Active {
			continue
		}

		// Check if event has expired
		endTime := event.StartTime + event.Duration
		if currentTime >= endTime {
			event.Active = false
			s.onEventEnd(w, event)
		}
	}
}

// checkForNewEvents potentially spawns new events.
func (s *CityEventSystem) checkForNewEvents(w *ecs.World, currentTime float64) {
	// Check if enough time has passed since last event
	if currentTime-s.lastEventTime < s.MinEventInterval {
		return
	}

	// Random chance for new event
	if s.rng.Float64() > s.eventProbability {
		return
	}

	// Count active events
	activeCount := s.countActiveEvents(w)
	if activeCount >= s.MaxActiveEvents {
		return
	}

	// Generate a new event
	s.generateEvent(w, currentTime)
	s.lastEventTime = currentTime
}

// countActiveEvents returns the number of currently active events.
func (s *CityEventSystem) countActiveEvents(w *ecs.World) int {
	count := 0
	for _, e := range w.Entities("CityEvent") {
		eventComp, ok := w.GetComponent(e, "CityEvent")
		if !ok {
			continue
		}
		event := eventComp.(*components.CityEvent)
		if event.Active {
			count++
		}
	}
	return count
}

// generateEvent creates a new random event.
func (s *CityEventSystem) generateEvent(w *ecs.World, currentTime float64) {
	eventTypes := s.getEventPool()
	eventType := eventTypes[s.rng.Intn(len(eventTypes))]
	event := s.createEvent(eventType, currentTime)

	// Create entity for the event
	entity := w.CreateEntity()
	w.AddComponent(entity, event)
}

// getEventPool returns genre-appropriate event types.
func (s *CityEventSystem) getEventPool() []string {
	switch s.Genre {
	case "fantasy":
		return []string{
			EventTypeFestival, EventTypeMarketDay, EventTypePlague,
			EventTypeRiot, EventTypeSiege, EventTypeCelebration,
			EventTypeReligiousRite, EventTypeTournament,
		}
	case "sci-fi":
		return []string{
			EventTypeFestival, EventTypeMarketDay, EventTypePlague,
			EventTypeRiot, EventTypeMartialLaw, EventTypeTradeCaravan,
			EventTypeBlackout,
		}
	case "horror":
		return []string{
			EventTypePlague, EventTypeRiot, EventTypeMartialLaw,
			EventTypeCultRitual, EventTypeHaunting, EventTypeReligiousRite,
		}
	case "cyberpunk":
		return []string{
			EventTypeFestival, EventTypeMarketDay, EventTypeRiot,
			EventTypeMartialLaw, EventTypeBlackout, EventTypeHacking,
		}
	case "post-apocalyptic":
		return []string{
			EventTypeMarketDay, EventTypePlague, EventTypeRiot,
			EventTypeMartialLaw, EventTypeTradeCaravan, EventTypeRadiation,
			EventTypeMutantAttack,
		}
	default:
		return []string{
			EventTypeFestival, EventTypeMarketDay, EventTypePlague,
			EventTypeRiot, EventTypeCelebration,
		}
	}
}

// createEvent builds a complete event of the given type.
func (s *CityEventSystem) createEvent(eventType string, startTime float64) *components.CityEvent {
	event := &components.CityEvent{
		EventType: eventType,
		StartTime: startTime,
		Active:    true,
	}

	switch eventType {
	case EventTypeFestival:
		event.Name = s.generateFestivalName()
		event.Description = "The city celebrates with music, food, and entertainment."
		event.Duration = CityEventDurationFestival
		event.Severity = 0.2
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.2,
			CrimePenaltyMultiplier: 0.8,
			NPCActivityChange:      "celebrating",
			SpawnRateMultiplier:    0.5,
			QuestRewardMultiplier:  1.1,
			GuardPatrolMultiplier:  0.7,
		}

	case EventTypeMarketDay:
		event.Name = "Grand Market Day"
		event.Description = "Merchants from afar gather to trade rare goods."
		event.Duration = CityEventDurationMarket
		event.Severity = 0.1
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    0.85,
			CrimePenaltyMultiplier: 1.0,
			NPCActivityChange:      "shopping",
			SpawnRateMultiplier:    1.0,
			QuestRewardMultiplier:  1.0,
			GuardPatrolMultiplier:  1.2,
		}

	case EventTypePlague:
		event.Name = "Outbreak"
		event.Description = "A dangerous illness spreads through the population."
		event.Duration = CityEventDurationPlague
		event.Severity = 0.8
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.5,
			CrimePenaltyMultiplier: 0.5,
			NPCActivityChange:      "hiding",
			SpawnRateMultiplier:    0.3,
			QuestRewardMultiplier:  1.5,
			GuardPatrolMultiplier:  0.4,
		}

	case EventTypeRiot:
		event.Name = "Civil Unrest"
		event.Description = "Angry citizens take to the streets in protest."
		event.Duration = CityEventDurationRiot
		event.Severity = 0.7
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.3,
			CrimePenaltyMultiplier: 0.3,
			NPCActivityChange:      "rioting",
			SpawnRateMultiplier:    2.0,
			QuestRewardMultiplier:  1.2,
			GuardPatrolMultiplier:  2.5,
		}

	case EventTypeSiege:
		event.Name = "City Under Siege"
		event.Description = "Enemy forces surround the city. All hands to the walls!"
		event.Duration = CityEventDurationSiege
		event.Severity = 1.0
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    2.0,
			CrimePenaltyMultiplier: 2.0,
			NPCActivityChange:      "defending",
			SpawnRateMultiplier:    3.0,
			QuestRewardMultiplier:  2.0,
			GuardPatrolMultiplier:  3.0,
		}

	case EventTypeCelebration:
		event.Name = "Victory Celebration"
		event.Description = "The city rejoices in a recent triumph."
		event.Duration = CityEventDurationCelebration
		event.Severity = 0.1
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.1,
			CrimePenaltyMultiplier: 0.9,
			NPCActivityChange:      "celebrating",
			SpawnRateMultiplier:    0.4,
			QuestRewardMultiplier:  1.2,
			GuardPatrolMultiplier:  0.8,
		}

	case EventTypeMartialLaw:
		event.Name = "Martial Law Declared"
		event.Description = "The authorities have imposed strict controls."
		event.Duration = CityEventDurationMartialLaw
		event.Severity = 0.6
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.2,
			CrimePenaltyMultiplier: 3.0,
			NPCActivityChange:      "hiding",
			SpawnRateMultiplier:    0.2,
			QuestRewardMultiplier:  0.8,
			GuardPatrolMultiplier:  4.0,
		}

	case EventTypeTradeCaravan:
		event.Name = "Trade Caravan Arrives"
		event.Description = "A large caravan brings exotic goods from distant lands."
		event.Duration = CityEventDurationCaravan
		event.Severity = 0.1
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    0.75,
			CrimePenaltyMultiplier: 1.0,
			NPCActivityChange:      "trading",
			SpawnRateMultiplier:    0.8,
			QuestRewardMultiplier:  1.0,
			GuardPatrolMultiplier:  1.3,
		}

	case EventTypeReligiousRite:
		event.Name = "Sacred Ritual"
		event.Description = "A significant religious ceremony takes place."
		event.Duration = CityEventDurationRite
		event.Severity = 0.3
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.0,
			CrimePenaltyMultiplier: 1.5,
			NPCActivityChange:      "worshipping",
			SpawnRateMultiplier:    0.6,
			QuestRewardMultiplier:  1.0,
			GuardPatrolMultiplier:  1.0,
		}

	case EventTypeTournament:
		event.Name = "Grand Tournament"
		event.Description = "Warriors gather to compete for glory and prizes."
		event.Duration = CityEventDurationTournament
		event.Severity = 0.2
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.15,
			CrimePenaltyMultiplier: 1.2,
			NPCActivityChange:      "spectating",
			SpawnRateMultiplier:    0.7,
			QuestRewardMultiplier:  1.3,
			GuardPatrolMultiplier:  1.5,
		}

	case EventTypeBlackout:
		event.Name = "Power Grid Failure"
		event.Description = "The city's power systems have failed."
		event.Duration = CityEventDurationBlackout
		event.Severity = 0.5
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.4,
			CrimePenaltyMultiplier: 0.5,
			NPCActivityChange:      "panicking",
			SpawnRateMultiplier:    1.8,
			QuestRewardMultiplier:  1.2,
			GuardPatrolMultiplier:  0.6,
		}

	case EventTypeHacking:
		event.Name = "Corporate Data Breach"
		event.Description = "Hackers have compromised the city's systems."
		event.Duration = CityEventDurationHacking
		event.Severity = 0.4
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.1,
			CrimePenaltyMultiplier: 0.7,
			NPCActivityChange:      "panicking",
			SpawnRateMultiplier:    1.3,
			QuestRewardMultiplier:  1.4,
			GuardPatrolMultiplier:  1.8,
		}

	case EventTypeRadiation:
		event.Name = "Radiation Storm"
		event.Description = "Dangerous levels of radiation sweep through the area."
		event.Duration = CityEventDurationRadiation
		event.Severity = 0.9
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.6,
			CrimePenaltyMultiplier: 0.4,
			NPCActivityChange:      "sheltering",
			SpawnRateMultiplier:    0.2,
			QuestRewardMultiplier:  1.5,
			GuardPatrolMultiplier:  0.3,
		}

	case EventTypeMutantAttack:
		event.Name = "Mutant Incursion"
		event.Description = "Mutated creatures have breached the settlement's defenses."
		event.Duration = CityEventDurationMutantAttack
		event.Severity = 0.8
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.5,
			CrimePenaltyMultiplier: 0.5,
			NPCActivityChange:      "fighting",
			SpawnRateMultiplier:    2.5,
			QuestRewardMultiplier:  1.8,
			GuardPatrolMultiplier:  2.0,
		}

	case EventTypeCultRitual:
		event.Name = "Dark Ritual"
		event.Description = "Cultists perform a sinister ceremony in the shadows."
		event.Duration = CityEventDurationCultRitual
		event.Severity = 0.7
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.2,
			CrimePenaltyMultiplier: 0.6,
			NPCActivityChange:      "hiding",
			SpawnRateMultiplier:    1.6,
			QuestRewardMultiplier:  1.5,
			GuardPatrolMultiplier:  1.4,
		}

	case EventTypeHaunting:
		event.Name = "Supernatural Manifestation"
		event.Description = "Strange phenomena plague the city as spirits grow restless."
		event.Duration = CityEventDurationHaunting
		event.Severity = 0.6
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.3,
			CrimePenaltyMultiplier: 0.8,
			NPCActivityChange:      "frightened",
			SpawnRateMultiplier:    1.4,
			QuestRewardMultiplier:  1.3,
			GuardPatrolMultiplier:  0.9,
		}

	default:
		event.Name = "Unknown Event"
		event.Description = "Something unusual is happening."
		event.Duration = CityEventDurationDefault
		event.Severity = 0.5
		event.Effects = components.CityEventEffects{
			ShopPriceMultiplier:    1.0,
			CrimePenaltyMultiplier: 1.0,
			SpawnRateMultiplier:    1.0,
			QuestRewardMultiplier:  1.0,
			GuardPatrolMultiplier:  1.0,
		}
	}

	return event
}

// generateFestivalName creates a genre-appropriate festival name.
func (s *CityEventSystem) generateFestivalName() string {
	var prefixes, suffixes []string
	switch s.Genre {
	case "fantasy":
		prefixes = []string{"Harvest", "Spring", "Midsummer", "Dragon", "Moon", "Star"}
		suffixes = []string{"Festival", "Fair", "Feast", "Celebration", "Revelry"}
	case "sci-fi":
		prefixes = []string{"Founders", "Unity", "First Contact", "Nova", "Stellar", "Colony"}
		suffixes = []string{"Day", "Week", "Festival", "Celebration", "Jubilee"}
	case "cyberpunk":
		prefixes = []string{"Neon", "Grid", "Corporate", "Street", "Synth", "Chrome"}
		suffixes = []string{"Fest", "Rave", "Carnival", "Block Party", "Celebration"}
	case "post-apocalyptic":
		prefixes = []string{"Remembrance", "Founding", "Harvest", "Water", "Trade", "Survival"}
		suffixes = []string{"Day", "Gathering", "Festival", "Celebration", "Feast"}
	default:
		prefixes = []string{"Annual", "Grand", "Seasonal", "Traditional"}
		suffixes = []string{"Festival", "Fair", "Celebration"}
	}

	prefix := prefixes[s.rng.Intn(len(prefixes))]
	suffix := suffixes[s.rng.Intn(len(suffixes))]
	return fmt.Sprintf("%s %s", prefix, suffix)
}

// onEventEnd handles cleanup when an event ends.
func (s *CityEventSystem) onEventEnd(w *ecs.World, event *components.CityEvent) {
	// Could trigger follow-up events or quest availability here
	// For now, just mark as inactive (already done in updateActiveEvents)
}

// GetActiveEvents returns all currently active events.
func (s *CityEventSystem) GetActiveEvents(w *ecs.World) []*components.CityEvent {
	var events []*components.CityEvent
	for _, e := range w.Entities("CityEvent") {
		eventComp, ok := w.GetComponent(e, "CityEvent")
		if !ok {
			continue
		}
		event := eventComp.(*components.CityEvent)
		if event.Active {
			events = append(events, event)
		}
	}
	return events
}

// GetShopPriceMultiplier returns the combined price multiplier from all active events.
func (s *CityEventSystem) GetShopPriceMultiplier(w *ecs.World) float64 {
	multiplier := 1.0
	for _, event := range s.GetActiveEvents(w) {
		multiplier *= event.Effects.ShopPriceMultiplier
	}
	return multiplier
}

// GetCrimePenaltyMultiplier returns the combined crime penalty multiplier.
func (s *CityEventSystem) GetCrimePenaltyMultiplier(w *ecs.World) float64 {
	multiplier := 1.0
	for _, event := range s.GetActiveEvents(w) {
		multiplier *= event.Effects.CrimePenaltyMultiplier
	}
	return multiplier
}

// GetSpawnRateMultiplier returns the combined hostile spawn rate multiplier.
func (s *CityEventSystem) GetSpawnRateMultiplier(w *ecs.World) float64 {
	multiplier := 1.0
	for _, event := range s.GetActiveEvents(w) {
		multiplier *= event.Effects.SpawnRateMultiplier
	}
	return multiplier
}

// GetQuestRewardMultiplier returns the combined quest reward multiplier.
func (s *CityEventSystem) GetQuestRewardMultiplier(w *ecs.World) float64 {
	multiplier := 1.0
	for _, event := range s.GetActiveEvents(w) {
		multiplier *= event.Effects.QuestRewardMultiplier
	}
	return multiplier
}

// GetGuardPatrolMultiplier returns the combined guard patrol multiplier.
func (s *CityEventSystem) GetGuardPatrolMultiplier(w *ecs.World) float64 {
	multiplier := 1.0
	for _, event := range s.GetActiveEvents(w) {
		multiplier *= event.Effects.GuardPatrolMultiplier
	}
	return multiplier
}

// TriggerEvent manually triggers a specific event type.
func (s *CityEventSystem) TriggerEvent(w *ecs.World, eventType string) *components.CityEvent {
	currentTime := s.getCurrentGameTime(w)
	event := s.createEvent(eventType, currentTime)

	entity := w.CreateEntity()
	w.AddComponent(entity, event)

	return event
}

// HasActiveEventOfType checks if an event of the given type is active.
func (s *CityEventSystem) HasActiveEventOfType(w *ecs.World, eventType string) bool {
	for _, event := range s.GetActiveEvents(w) {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

// GetEventSeverity returns the highest severity among active events.
func (s *CityEventSystem) GetEventSeverity(w *ecs.World) float64 {
	maxSeverity := 0.0
	for _, event := range s.GetActiveEvents(w) {
		if event.Severity > maxSeverity {
			maxSeverity = event.Severity
		}
	}
	return maxSeverity
}
