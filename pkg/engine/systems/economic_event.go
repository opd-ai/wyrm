package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// EconomicEventType represents types of economic events.
type EconomicEventType int

const (
	EventMarketBoom EconomicEventType = iota
	EventMarketCrash
	EventShortage
	EventSurplus
	EventInflation
	EventDeflation
	EventTradeWar
	EventGoldRush
	EventPlague
	EventHarvest
	EventDrought
	EventDiscovery
	EventMonopoly
	EventBankruptcy
	EventTaxChange
)

// EconomicEventSeverity represents the impact level of an event.
type EconomicEventSeverity int

const (
	SeverityMinor EconomicEventSeverity = iota
	SeverityModerate
	SeverityMajor
	SeverityCatastrophic
)

// EconomicEvent represents an active economic event.
type EconomicEvent struct {
	ID             string
	Type           EconomicEventType
	Severity       EconomicEventSeverity
	Name           string
	Description    string
	StartTime      float64
	Duration       float64      // Hours
	Progress       float64      // 0-1
	AffectedItems  []string     // Specific items affected
	AffectedNodes  []ecs.Entity // Specific market nodes affected
	PriceModifier  float64      // Multiplier on prices
	SupplyModifier float64      // Multiplier on supply
	DemandModifier float64      // Multiplier on demand
	IsGlobal       bool         // Affects all markets
	IsResolved     bool
	Consequences   []EventConsequence
}

// EventConsequence represents a follow-on effect of an event.
type EventConsequence struct {
	Description  string
	PriceEffect  float64
	SupplyEffect float64
	DemandEffect float64
	Duration     float64
}

// EconomicEventSystem manages economic events.
type EconomicEventSystem struct {
	Seed           int64
	Genre          string
	Economy        *EconomySystem
	Events         map[string]*EconomicEvent
	ActiveEvents   []*EconomicEvent
	EventHistory   []*EconomicEvent
	GameTime       float64
	counter        uint64
	EventFrequency float64 // Base chance per hour for random event
	MaxEvents      int     // Maximum concurrent events
}

// NewEconomicEventSystem creates a new economic event system.
func NewEconomicEventSystem(seed int64, genre string, economy *EconomySystem) *EconomicEventSystem {
	return &EconomicEventSystem{
		Seed:           seed,
		Genre:          genre,
		Economy:        economy,
		Events:         make(map[string]*EconomicEvent),
		ActiveEvents:   make([]*EconomicEvent, 0),
		EventHistory:   make([]*EconomicEvent, 0),
		EventFrequency: 0.001, // ~0.1% chance per hour
		MaxEvents:      5,
	}
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *EconomicEventSystem) pseudoRandom() float64 {
	s.counter++
	x := uint64(s.Seed) + s.counter*6364136223846793005
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	return float64(x%10000) / 10000.0
}

// pseudoRandomInt generates a deterministic pseudo-random integer.
func (s *EconomicEventSystem) pseudoRandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	return int(s.pseudoRandom() * float64(max))
}

// Update processes economic events.
func (s *EconomicEventSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	hoursDelta := dt / 3600.0
	// Process active events
	for _, event := range s.ActiveEvents {
		s.updateEvent(w, event, hoursDelta)
	}
	// Clean up resolved events
	s.cleanupResolvedEvents()
	// Potentially generate new events
	if len(s.ActiveEvents) < s.MaxEvents && s.pseudoRandom() < s.EventFrequency*hoursDelta {
		s.generateRandomEvent(w)
	}
}

// updateEvent processes a single active event.
func (s *EconomicEventSystem) updateEvent(w *ecs.World, event *EconomicEvent, hours float64) {
	if event.IsResolved {
		return
	}
	// Update progress
	event.Progress += hours / event.Duration
	// Apply effects to economy
	s.applyEventEffects(w, event)
	// Check if event is complete
	if event.Progress >= 1.0 {
		event.IsResolved = true
		s.applyEventConsequences(w, event)
	}
}

// applyEventEffects applies event modifiers to the economy.
func (s *EconomicEventSystem) applyEventEffects(w *ecs.World, event *EconomicEvent) {
	if event.IsGlobal {
		// Apply to all economy nodes
		for _, e := range w.Entities("EconomyNode") {
			s.applyToNode(w, e, event)
		}
	} else {
		// Apply to specific nodes
		for _, nodeID := range event.AffectedNodes {
			s.applyToNode(w, nodeID, event)
		}
	}
}

// applyToNode applies event effects to a single economy node.
func (s *EconomicEventSystem) applyToNode(w *ecs.World, nodeID ecs.Entity, event *EconomicEvent) {
	comp, ok := w.GetComponent(nodeID, "EconomyNode")
	if !ok {
		return
	}
	node := comp.(*components.EconomyNode)
	if node.PriceTable == nil || node.Supply == nil || node.Demand == nil {
		return
	}
	// Determine which items to affect
	items := event.AffectedItems
	if len(items) == 0 {
		// Affect all items
		items = make([]string, 0)
		for item := range node.PriceTable {
			items = append(items, item)
		}
	}
	// Apply modifiers gradually
	effectStrength := event.Progress * 0.01 // Apply 1% per tick
	for _, item := range items {
		// Price modifier
		if event.PriceModifier != 1.0 {
			currentPrice := node.PriceTable[item]
			targetPrice := currentPrice * event.PriceModifier
			node.PriceTable[item] = currentPrice + (targetPrice-currentPrice)*effectStrength
		}
		// Supply modifier
		if event.SupplyModifier != 1.0 {
			currentSupply := node.Supply[item]
			targetSupply := int(float64(currentSupply) * event.SupplyModifier)
			diff := targetSupply - currentSupply
			if diff > 0 {
				node.Supply[item] = currentSupply + int(float64(diff)*effectStrength)
			} else if diff < 0 {
				node.Supply[item] = currentSupply + int(float64(diff)*effectStrength)
				if node.Supply[item] < 0 {
					node.Supply[item] = 0
				}
			}
		}
		// Demand modifier
		if event.DemandModifier != 1.0 {
			currentDemand := node.Demand[item]
			targetDemand := int(float64(currentDemand) * event.DemandModifier)
			diff := targetDemand - currentDemand
			if diff > 0 {
				node.Demand[item] = currentDemand + int(float64(diff)*effectStrength)
			} else if diff < 0 {
				node.Demand[item] = currentDemand + int(float64(diff)*effectStrength)
				if node.Demand[item] < 0 {
					node.Demand[item] = 0
				}
			}
		}
	}
}

// applyEventConsequences applies follow-on effects when an event ends.
func (s *EconomicEventSystem) applyEventConsequences(w *ecs.World, event *EconomicEvent) {
	for _, consequence := range event.Consequences {
		// Create a follow-on event for each consequence
		consequenceEvent := &EconomicEvent{
			ID:             event.ID + "_consequence",
			Type:           event.Type,
			Severity:       SeverityMinor,
			Name:           "Aftermath: " + event.Name,
			Description:    consequence.Description,
			StartTime:      s.GameTime,
			Duration:       consequence.Duration,
			Progress:       0,
			AffectedItems:  event.AffectedItems,
			AffectedNodes:  event.AffectedNodes,
			PriceModifier:  consequence.PriceEffect,
			SupplyModifier: consequence.SupplyEffect,
			DemandModifier: consequence.DemandEffect,
			IsGlobal:       event.IsGlobal,
		}
		s.Events[consequenceEvent.ID] = consequenceEvent
		s.ActiveEvents = append(s.ActiveEvents, consequenceEvent)
	}
}

// cleanupResolvedEvents removes resolved events from active list.
func (s *EconomicEventSystem) cleanupResolvedEvents() {
	activeEvents := make([]*EconomicEvent, 0)
	for _, event := range s.ActiveEvents {
		if !event.IsResolved {
			activeEvents = append(activeEvents, event)
		} else {
			s.EventHistory = append(s.EventHistory, event)
			if len(s.EventHistory) > 50 {
				s.EventHistory = s.EventHistory[1:]
			}
		}
	}
	s.ActiveEvents = activeEvents
}

// generateRandomEvent creates a new random economic event.
func (s *EconomicEventSystem) generateRandomEvent(w *ecs.World) *EconomicEvent {
	eventType := s.selectRandomEventType()
	severity := s.selectRandomSeverity()
	event := s.createEvent(eventType, severity)
	// Determine scope
	if s.pseudoRandom() < 0.2 {
		event.IsGlobal = true
	} else {
		// Select random affected nodes
		nodes := w.Entities("EconomyNode")
		if len(nodes) > 0 {
			numNodes := 1 + s.pseudoRandomInt(3)
			for i := 0; i < numNodes && i < len(nodes); i++ {
				event.AffectedNodes = append(event.AffectedNodes, nodes[s.pseudoRandomInt(len(nodes))])
			}
		}
	}
	// Select affected items
	event.AffectedItems = s.selectAffectedItems()
	s.Events[event.ID] = event
	s.ActiveEvents = append(s.ActiveEvents, event)
	return event
}

// selectRandomEventType chooses a random event type weighted by genre.
func (s *EconomicEventSystem) selectRandomEventType() EconomicEventType {
	eventTypes := s.getGenreEventTypes()
	return eventTypes[s.pseudoRandomInt(len(eventTypes))]
}

// getGenreEventTypes returns event types appropriate for the genre.
func (s *EconomicEventSystem) getGenreEventTypes() []EconomicEventType {
	common := []EconomicEventType{
		EventMarketBoom, EventMarketCrash, EventShortage, EventSurplus,
		EventInflation, EventDeflation, EventTradeWar,
	}
	switch s.Genre {
	case "fantasy":
		return append(common, EventHarvest, EventDrought, EventDiscovery)
	case "sci-fi":
		return append(common, EventDiscovery, EventMonopoly, EventBankruptcy)
	case "horror":
		return append(common, EventPlague, EventShortage, EventBankruptcy)
	case "cyberpunk":
		return append(common, EventMonopoly, EventBankruptcy, EventTaxChange)
	case "post-apocalyptic":
		return append(common, EventDrought, EventPlague, EventGoldRush)
	default:
		return common
	}
}

// selectRandomSeverity chooses a severity level.
func (s *EconomicEventSystem) selectRandomSeverity() EconomicEventSeverity {
	roll := s.pseudoRandom()
	if roll < 0.5 {
		return SeverityMinor
	} else if roll < 0.8 {
		return SeverityModerate
	} else if roll < 0.95 {
		return SeverityMajor
	}
	return SeverityCatastrophic
}

// selectAffectedItems chooses which items are affected.
func (s *EconomicEventSystem) selectAffectedItems() []string {
	if s.Economy == nil || s.Economy.BasePrices == nil {
		return nil
	}
	items := make([]string, 0)
	for item := range s.Economy.BasePrices {
		// 30% chance each item is affected
		if s.pseudoRandom() < 0.3 {
			items = append(items, item)
		}
	}
	return items
}

// createEvent creates an event with appropriate parameters.
func (s *EconomicEventSystem) createEvent(eventType EconomicEventType, severity EconomicEventSeverity) *EconomicEvent {
	s.counter++
	eventID := "event_" + string(rune('0'+s.counter%1000))
	name, description := s.getEventNameAndDescription(eventType)
	pricemod, supplymod, demandmod := s.getEventModifiers(eventType, severity)
	duration := s.getEventDuration(eventType, severity)
	consequences := s.getEventConsequences(eventType, severity)
	return &EconomicEvent{
		ID:             eventID,
		Type:           eventType,
		Severity:       severity,
		Name:           name,
		Description:    description,
		StartTime:      s.GameTime,
		Duration:       duration,
		Progress:       0,
		AffectedItems:  make([]string, 0),
		AffectedNodes:  make([]ecs.Entity, 0),
		PriceModifier:  pricemod,
		SupplyModifier: supplymod,
		DemandModifier: demandmod,
		Consequences:   consequences,
	}
}

// getEventNameAndDescription returns name and description for event type.
func (s *EconomicEventSystem) getEventNameAndDescription(eventType EconomicEventType) (string, string) {
	descriptions := s.getGenreEventDescriptions()
	if info, ok := descriptions[eventType]; ok {
		return info[0], info[1]
	}
	return "Economic Event", "An unexpected economic event has occurred."
}

// getGenreEventDescriptions returns genre-specific event descriptions.
func (s *EconomicEventSystem) getGenreEventDescriptions() map[EconomicEventType][2]string {
	switch s.Genre {
	case "fantasy":
		return map[EconomicEventType][2]string{
			EventMarketBoom:  {"Merchant Festival", "A grand trade festival has boosted commerce."},
			EventMarketCrash: {"Market Collapse", "Rumors of war have spooked the markets."},
			EventShortage:    {"Resource Scarcity", "Key materials have become scarce."},
			EventSurplus:     {"Abundant Harvest", "An abundant harvest has flooded markets."},
			EventInflation:   {"Coin Debasement", "The realm's currency is losing value."},
			EventDeflation:   {"Hoarding Crisis", "Merchants are hoarding coins, causing deflation."},
			EventTradeWar:    {"Guild Conflict", "Rival guilds are disrupting trade."},
			EventGoldRush:    {"Treasure Discovery", "A legendary treasure hoard has been found."},
			EventPlague:      {"Blight", "A magical blight affects trade goods."},
			EventHarvest:     {"Bountiful Harvest", "The gods have blessed the harvest."},
			EventDrought:     {"Famine", "Drought has struck the farmlands."},
			EventDiscovery:   {"New Trade Route", "A new trade route has been discovered."},
			EventMonopoly:    {"Guild Monopoly", "A powerful guild has cornered the market."},
			EventBankruptcy:  {"Noble Bankruptcy", "A major noble house has gone bankrupt."},
			EventTaxChange:   {"Royal Decree", "The king has changed taxation policies."},
		}
	case "sci-fi":
		return map[EconomicEventType][2]string{
			EventMarketBoom:  {"Trade Boom", "A new hyperspace route has boosted commerce."},
			EventMarketCrash: {"Market Crash", "Corporate scandal has crashed stock prices."},
			EventShortage:    {"Supply Crisis", "Critical components are in short supply."},
			EventSurplus:     {"Overproduction", "Automated factories have overproduced."},
			EventInflation:   {"Currency Crisis", "Credit inflation is running rampant."},
			EventDeflation:   {"Deflationary Spiral", "Consumer confidence has collapsed."},
			EventTradeWar:    {"Sector Embargo", "An interstellar trade embargo is in effect."},
			EventGoldRush:    {"Mineral Strike", "Rich mineral deposits have been discovered."},
			EventPlague:      {"Biohazard Outbreak", "A viral outbreak has disrupted supply chains."},
			EventHarvest:     {"Agri-World Surplus", "Colony farms report record yields."},
			EventDrought:     {"Terraform Failure", "A terraform failure has caused crop failures."},
			EventDiscovery:   {"Tech Breakthrough", "A major technological breakthrough."},
			EventMonopoly:    {"Corporate Takeover", "A megacorp has monopolized the sector."},
			EventBankruptcy:  {"Corporate Collapse", "A major corporation has collapsed."},
			EventTaxChange:   {"Tariff Change", "New import tariffs have been imposed."},
		}
	case "horror":
		return map[EconomicEventType][2]string{
			EventMarketBoom:  {"Desperate Trading", "Fear has driven frantic commerce."},
			EventMarketCrash: {"Panic Selling", "Terror has gripped the markets."},
			EventShortage:    {"Supply Vanishing", "Goods are mysteriously disappearing."},
			EventSurplus:     {"Suspicious Abundance", "An unnatural surplus has appeared."},
			EventInflation:   {"Worthless Currency", "Money holds little value in these dark times."},
			EventDeflation:   {"Hoarding", "Survivors hoard everything they can."},
			EventTradeWar:    {"Route Closure", "Dangerous routes have been abandoned."},
			EventGoldRush:    {"Relic Hunt", "A valuable relic location has been revealed."},
			EventPlague:      {"Corruption Spread", "An unnatural plague spreads."},
			EventHarvest:     {"Tainted Harvest", "The harvest carries a strange taint."},
			EventDrought:     {"Blighted Fields", "Fields have withered mysteriously."},
			EventDiscovery:   {"Forbidden Knowledge", "Dark secrets have been uncovered."},
			EventMonopoly:    {"Cult Control", "A cult has seized control of trade."},
			EventBankruptcy:  {"Merchant Disappearance", "A major merchant has vanished."},
			EventTaxChange:   {"Protection Tax", "New 'protection' fees are demanded."},
		}
	case "cyberpunk":
		return map[EconomicEventType][2]string{
			EventMarketBoom:  {"Bull Market", "Stock prices are surging."},
			EventMarketCrash: {"Flash Crash", "AI trading has triggered a crash."},
			EventShortage:    {"Chip Shortage", "Critical components are unavailable."},
			EventSurplus:     {"Market Flooding", "Counterfeit goods flood the market."},
			EventInflation:   {"Hyperinflation", "Corporate currencies are inflating."},
			EventDeflation:   {"Crypto Collapse", "A major cryptocurrency has collapsed."},
			EventTradeWar:    {"Corporate War", "Megacorps are disrupting each other."},
			EventGoldRush:    {"Data Bonanza", "Valuable data caches have been located."},
			EventPlague:      {"Cyber Plague", "A computer virus is disrupting trade."},
			EventHarvest:     {"Vertical Farm Boom", "Synth-food production is up."},
			EventDrought:     {"Water Crisis", "Clean water is becoming scarce."},
			EventDiscovery:   {"Black Market Find", "A black market cache has been found."},
			EventMonopoly:    {"Hostile Takeover", "A zaibatsu has taken over the sector."},
			EventBankruptcy:  {"Corporate Implosion", "A megacorp has imploded."},
			EventTaxChange:   {"Tax Haven Closed", "Offshore accounts have been frozen."},
		}
	case "post-apocalyptic":
		return map[EconomicEventType][2]string{
			EventMarketBoom:  {"Trade Caravan", "A major caravan has arrived."},
			EventMarketCrash: {"Raider Attack", "Raiders have disrupted trade routes."},
			EventShortage:    {"Scarcity", "Essential supplies are running out."},
			EventSurplus:     {"Cache Discovery", "A pre-war cache has been found."},
			EventInflation:   {"Caps Devaluation", "Currency is losing its value."},
			EventDeflation:   {"Barter Economy", "People are returning to barter."},
			EventTradeWar:    {"Faction Conflict", "Factions are fighting over trade routes."},
			EventGoldRush:    {"Tech Salvage", "Major tech salvage opportunities found."},
			EventPlague:      {"Rad Sickness", "Radiation sickness is spreading."},
			EventHarvest:     {"Good Season", "The growing season was successful."},
			EventDrought:     {"Dust Storm", "Dust storms have destroyed crops."},
			EventDiscovery:   {"Vault Opening", "A sealed vault has been opened."},
			EventMonopoly:    {"Warlord Control", "A warlord has seized key supplies."},
			EventBankruptcy:  {"Settlement Collapse", "A major settlement has failed."},
			EventTaxChange:   {"Tribute Demand", "New tribute demands from raiders."},
		}
	default:
		return map[EconomicEventType][2]string{
			EventMarketBoom:  {"Market Boom", "Economic activity is surging."},
			EventMarketCrash: {"Market Crash", "Prices are plummeting."},
			EventShortage:    {"Shortage", "Supplies are scarce."},
			EventSurplus:     {"Surplus", "There is an abundance of goods."},
		}
	}
}

// getEventModifiers returns price/supply/demand modifiers for an event.
func (s *EconomicEventSystem) getEventModifiers(eventType EconomicEventType, severity EconomicEventSeverity) (float64, float64, float64) {
	severityMult := s.getSeverityMultiplier(severity)
	switch eventType {
	case EventMarketBoom:
		return 1.0 + 0.2*severityMult, 1.0 + 0.1*severityMult, 1.0 + 0.3*severityMult
	case EventMarketCrash:
		return 1.0 - 0.3*severityMult, 1.0, 1.0 - 0.4*severityMult
	case EventShortage:
		return 1.0 + 0.4*severityMult, 1.0 - 0.5*severityMult, 1.0
	case EventSurplus:
		return 1.0 - 0.3*severityMult, 1.0 + 0.5*severityMult, 1.0
	case EventInflation:
		return 1.0 + 0.3*severityMult, 1.0, 1.0 - 0.2*severityMult
	case EventDeflation:
		return 1.0 - 0.2*severityMult, 1.0, 1.0 - 0.3*severityMult
	case EventTradeWar:
		return 1.0 + 0.2*severityMult, 1.0 - 0.3*severityMult, 1.0
	case EventGoldRush:
		return 1.0 - 0.1*severityMult, 1.0 + 0.4*severityMult, 1.0 + 0.2*severityMult
	case EventPlague:
		return 1.0 + 0.3*severityMult, 1.0 - 0.4*severityMult, 1.0 - 0.2*severityMult
	case EventHarvest:
		return 1.0 - 0.2*severityMult, 1.0 + 0.4*severityMult, 1.0
	case EventDrought:
		return 1.0 + 0.4*severityMult, 1.0 - 0.5*severityMult, 1.0 + 0.2*severityMult
	case EventDiscovery:
		return 1.0 - 0.1*severityMult, 1.0 + 0.3*severityMult, 1.0 + 0.2*severityMult
	case EventMonopoly:
		return 1.0 + 0.5*severityMult, 1.0 - 0.2*severityMult, 1.0
	case EventBankruptcy:
		return 1.0 - 0.2*severityMult, 1.0 + 0.3*severityMult, 1.0 - 0.3*severityMult
	case EventTaxChange:
		return 1.0 + 0.15*severityMult, 1.0, 1.0 - 0.1*severityMult
	default:
		return 1.0, 1.0, 1.0
	}
}

// getSeverityMultiplier returns a multiplier based on severity.
func (s *EconomicEventSystem) getSeverityMultiplier(severity EconomicEventSeverity) float64 {
	switch severity {
	case SeverityMinor:
		return 0.5
	case SeverityModerate:
		return 1.0
	case SeverityMajor:
		return 1.5
	case SeverityCatastrophic:
		return 2.0
	default:
		return 1.0
	}
}

// getEventDuration returns duration based on event type and severity.
func (s *EconomicEventSystem) getEventDuration(eventType EconomicEventType, severity EconomicEventSeverity) float64 {
	baseDuration := 48.0 // Default 2 days
	switch eventType {
	case EventMarketBoom, EventMarketCrash:
		baseDuration = 24.0
	case EventShortage, EventSurplus:
		baseDuration = 72.0
	case EventInflation, EventDeflation:
		baseDuration = 168.0 // 1 week
	case EventTradeWar:
		baseDuration = 120.0
	case EventPlague, EventDrought:
		baseDuration = 336.0 // 2 weeks
	case EventGoldRush, EventDiscovery:
		baseDuration = 48.0
	case EventMonopoly:
		baseDuration = 240.0
	case EventBankruptcy:
		baseDuration = 72.0
	case EventTaxChange:
		baseDuration = 720.0 // 1 month
	}
	// Severity affects duration
	severityMult := s.getSeverityMultiplier(severity)
	return baseDuration * severityMult
}

// getEventConsequences returns follow-on consequences for an event.
func (s *EconomicEventSystem) getEventConsequences(eventType EconomicEventType, severity EconomicEventSeverity) []EventConsequence {
	if severity < SeverityModerate {
		return nil // Minor events have no consequences
	}
	switch eventType {
	case EventMarketCrash:
		return []EventConsequence{
			{Description: "Recovery period", PriceEffect: 1.1, SupplyEffect: 0.9, DemandEffect: 0.95, Duration: 72},
		}
	case EventShortage:
		return []EventConsequence{
			{Description: "Price adjustment", PriceEffect: 1.15, SupplyEffect: 1.1, DemandEffect: 1.0, Duration: 48},
		}
	case EventPlague:
		return []EventConsequence{
			{Description: "Quarantine aftermath", PriceEffect: 1.1, SupplyEffect: 0.95, DemandEffect: 1.1, Duration: 96},
		}
	case EventDrought:
		return []EventConsequence{
			{Description: "Rationing period", PriceEffect: 1.2, SupplyEffect: 0.8, DemandEffect: 1.1, Duration: 120},
		}
	default:
		return nil
	}
}

// TriggerEvent manually triggers an economic event.
func (s *EconomicEventSystem) TriggerEvent(eventType EconomicEventType, severity EconomicEventSeverity, affectedNodes []ecs.Entity, affectedItems []string) *EconomicEvent {
	event := s.createEvent(eventType, severity)
	event.AffectedNodes = affectedNodes
	event.AffectedItems = affectedItems
	if len(affectedNodes) == 0 {
		event.IsGlobal = true
	}
	s.Events[event.ID] = event
	s.ActiveEvents = append(s.ActiveEvents, event)
	return event
}

// GetActiveEvents returns all currently active events.
func (s *EconomicEventSystem) GetActiveEvents() []*EconomicEvent {
	return s.ActiveEvents
}

// GetEventHistory returns recent event history.
func (s *EconomicEventSystem) GetEventHistory() []*EconomicEvent {
	return s.EventHistory
}

// GetEvent returns a specific event by ID.
func (s *EconomicEventSystem) GetEvent(eventID string) (*EconomicEvent, bool) {
	event, exists := s.Events[eventID]
	return event, exists
}

// SetEventFrequency sets the base chance for random events.
func (s *EconomicEventSystem) SetEventFrequency(frequency float64) {
	s.EventFrequency = clampFloat(frequency, 0, 1)
}

// GetSeverityDescription returns a description of severity level.
func (s *EconomicEventSystem) GetSeverityDescription(severity EconomicEventSeverity) string {
	switch severity {
	case SeverityMinor:
		return "Minor - Limited impact on a few markets"
	case SeverityModerate:
		return "Moderate - Noticeable effect on regional trade"
	case SeverityMajor:
		return "Major - Significant disruption to commerce"
	case SeverityCatastrophic:
		return "Catastrophic - Devastating impact on the economy"
	default:
		return "Unknown severity"
	}
}

// EstimateEventImpact estimates the economic impact of an event.
func (s *EconomicEventSystem) EstimateEventImpact(event *EconomicEvent) float64 {
	// Calculate overall impact score
	priceImpact := absFloat(event.PriceModifier - 1.0)
	supplyImpact := absFloat(event.SupplyModifier - 1.0)
	demandImpact := absFloat(event.DemandModifier - 1.0)
	baseImpact := (priceImpact + supplyImpact + demandImpact) / 3.0
	// Scale by scope
	scopeMultiplier := 1.0
	if event.IsGlobal {
		scopeMultiplier = 3.0
	} else {
		scopeMultiplier = float64(len(event.AffectedNodes)) * 0.5
	}
	// Scale by duration
	durationFactor := event.Duration / 168.0 // Relative to 1 week
	return baseImpact * scopeMultiplier * durationFactor * 100.0
}

// absFloat returns the absolute value of a float.
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
