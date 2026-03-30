package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// ManipulationType represents types of market manipulation.
type ManipulationType int

const (
	ManipulationNone        ManipulationType = iota
	ManipulationCorner                       // Buy up all supply to control prices
	ManipulationDump                         // Flood market to crash prices
	ManipulationRumor                        // Spread false information
	ManipulationSabotage                     // Sabotage competitor supply
	ManipulationMonopoly                     // Establish exclusive control
	ManipulationCartel                       // Form price-fixing agreements
	ManipulationInsider                      // Trade on non-public information
	ManipulationCounterfeit                  // Introduce fake goods
)

// ManipulationStatus represents the state of a manipulation scheme.
type ManipulationStatus int

const (
	ManipulationPlanning ManipulationStatus = iota
	ManipulationActive
	ManipulationSucceeded
	ManipulationFailed
	ManipulationDiscovered
)

// MarketManipulation represents an active manipulation scheme.
type MarketManipulation struct {
	ID            string
	Manipulator   ecs.Entity
	Type          ManipulationType
	TargetItem    string
	TargetNode    ecs.Entity // Target market/location
	Status        ManipulationStatus
	StartTime     float64
	Duration      float64 // How long the manipulation lasts
	Investment    float64 // Gold invested
	Progress      float64 // 0-1 progress
	SuccessChance float64 // Current success probability
	DetectionRisk float64 // Risk of being caught
	PriceEffect   float64 // Multiplier on target price
	Accomplices   []ecs.Entity
	Evidence      float64 // How much evidence exists (0-1)
}

// ManipulationOutcome represents the result of a manipulation.
type ManipulationOutcome struct {
	Success       bool
	Profit        float64
	Discovered    bool
	Penalty       float64 // Fine if discovered
	ReputationHit float64 // Reputation damage if discovered
	CrimeID       string  // If discovered, crime record ID
}

// MarketManipulationSystem manages market manipulation schemes.
type MarketManipulationSystem struct {
	Seed             int64
	Genre            string
	Economy          *EconomySystem
	Manipulations    map[string]*MarketManipulation
	GameTime         float64
	counter          uint64
	DetectionAgents  map[ecs.Entity]float64 // Node -> detection capability
	ManipulatorSkill map[ecs.Entity]float64 // Player skill level
	Cooldowns        map[ecs.Entity]float64 // Player -> next allowed time
	History          map[ecs.Entity][]ManipulationOutcome
}

// NewMarketManipulationSystem creates a new market manipulation system.
func NewMarketManipulationSystem(seed int64, genre string, economy *EconomySystem) *MarketManipulationSystem {
	return &MarketManipulationSystem{
		Seed:             seed,
		Genre:            genre,
		Economy:          economy,
		Manipulations:    make(map[string]*MarketManipulation),
		DetectionAgents:  make(map[ecs.Entity]float64),
		ManipulatorSkill: make(map[ecs.Entity]float64),
		Cooldowns:        make(map[ecs.Entity]float64),
		History:          make(map[ecs.Entity][]ManipulationOutcome),
	}
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *MarketManipulationSystem) pseudoRandom() float64 {
	s.counter++
	x := uint64(s.Seed) + s.counter*6364136223846793005
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	return float64(x%10000) / 10000.0
}

// StartManipulation initiates a new market manipulation scheme.
func (s *MarketManipulationSystem) StartManipulation(manipulator ecs.Entity, mType ManipulationType, targetItem string, targetNode ecs.Entity, investment float64) (*MarketManipulation, error) {
	// Check cooldown
	if s.Cooldowns[manipulator] > s.GameTime {
		return nil, ErrManipulationCooldown
	}
	// Check minimum investment
	minInvestment := s.getMinimumInvestment(mType)
	if investment < minInvestment {
		return nil, ErrInsufficientInvestment
	}
	manipID := s.generateManipulationID()
	baseSuccess := s.getBaseSuccessChance(mType)
	skillBonus := s.ManipulatorSkill[manipulator] * 0.2
	manipulation := &MarketManipulation{
		ID:            manipID,
		Manipulator:   manipulator,
		Type:          mType,
		TargetItem:    targetItem,
		TargetNode:    targetNode,
		Status:        ManipulationPlanning,
		StartTime:     s.GameTime,
		Duration:      s.getManipulationDuration(mType),
		Investment:    investment,
		Progress:      0,
		SuccessChance: clampFloat(baseSuccess+skillBonus, 0.1, 0.95),
		DetectionRisk: s.getBaseDetectionRisk(mType),
		PriceEffect:   s.getPriceEffect(mType),
		Accomplices:   make([]ecs.Entity, 0),
		Evidence:      0,
	}
	s.Manipulations[manipID] = manipulation
	return manipulation, nil
}

// generateManipulationID creates a unique manipulation identifier.
func (s *MarketManipulationSystem) generateManipulationID() string {
	s.counter++
	return "manip_" + string(rune('0'+s.counter%1000))
}

// getMinimumInvestment returns minimum gold required for manipulation type.
func (s *MarketManipulationSystem) getMinimumInvestment(mType ManipulationType) float64 {
	switch mType {
	case ManipulationCorner:
		return 5000.0
	case ManipulationDump:
		return 2000.0
	case ManipulationRumor:
		return 500.0
	case ManipulationSabotage:
		return 1500.0
	case ManipulationMonopoly:
		return 10000.0
	case ManipulationCartel:
		return 3000.0
	case ManipulationInsider:
		return 1000.0
	case ManipulationCounterfeit:
		return 2500.0
	default:
		return 1000.0
	}
}

// getBaseSuccessChance returns base success probability for manipulation type.
func (s *MarketManipulationSystem) getBaseSuccessChance(mType ManipulationType) float64 {
	switch mType {
	case ManipulationCorner:
		return 0.6
	case ManipulationDump:
		return 0.75
	case ManipulationRumor:
		return 0.7
	case ManipulationSabotage:
		return 0.5
	case ManipulationMonopoly:
		return 0.3
	case ManipulationCartel:
		return 0.55
	case ManipulationInsider:
		return 0.8
	case ManipulationCounterfeit:
		return 0.45
	default:
		return 0.5
	}
}

// getBaseDetectionRisk returns base detection risk for manipulation type.
func (s *MarketManipulationSystem) getBaseDetectionRisk(mType ManipulationType) float64 {
	switch mType {
	case ManipulationCorner:
		return 0.4
	case ManipulationDump:
		return 0.3
	case ManipulationRumor:
		return 0.2
	case ManipulationSabotage:
		return 0.5
	case ManipulationMonopoly:
		return 0.6
	case ManipulationCartel:
		return 0.45
	case ManipulationInsider:
		return 0.35
	case ManipulationCounterfeit:
		return 0.7
	default:
		return 0.4
	}
}

// getManipulationDuration returns how long a manipulation takes (hours).
func (s *MarketManipulationSystem) getManipulationDuration(mType ManipulationType) float64 {
	switch mType {
	case ManipulationCorner:
		return 48.0
	case ManipulationDump:
		return 6.0
	case ManipulationRumor:
		return 24.0
	case ManipulationSabotage:
		return 12.0
	case ManipulationMonopoly:
		return 168.0 // 1 week
	case ManipulationCartel:
		return 72.0
	case ManipulationInsider:
		return 2.0
	case ManipulationCounterfeit:
		return 48.0
	default:
		return 24.0
	}
}

// getPriceEffect returns the price multiplier for manipulation type.
func (s *MarketManipulationSystem) getPriceEffect(mType ManipulationType) float64 {
	switch mType {
	case ManipulationCorner:
		return 2.0 // Double prices
	case ManipulationDump:
		return 0.4 // 60% price drop
	case ManipulationRumor:
		return 1.5 // 50% increase/decrease based on rumor
	case ManipulationSabotage:
		return 1.8 // Supply shortage
	case ManipulationMonopoly:
		return 3.0 // Triple prices
	case ManipulationCartel:
		return 1.7 // Agreed markup
	case ManipulationInsider:
		return 1.0 // No direct effect, just profit
	case ManipulationCounterfeit:
		return 0.6 // Flood with fakes
	default:
		return 1.0
	}
}

// AddAccomplice adds an accomplice to a manipulation scheme.
func (s *MarketManipulationSystem) AddAccomplice(manipID string, accomplice ecs.Entity) bool {
	manip, exists := s.Manipulations[manipID]
	if !exists || manip.Status != ManipulationPlanning {
		return false
	}
	manip.Accomplices = append(manip.Accomplices, accomplice)
	// More accomplices = higher success but higher detection
	manip.SuccessChance = clampFloat(manip.SuccessChance+0.05, 0, 0.95)
	manip.DetectionRisk = clampFloat(manip.DetectionRisk+0.1, 0, 0.95)
	return true
}

// ExecuteManipulation starts the active phase of a manipulation.
func (s *MarketManipulationSystem) ExecuteManipulation(manipID string) bool {
	manip, exists := s.Manipulations[manipID]
	if !exists || manip.Status != ManipulationPlanning {
		return false
	}
	manip.Status = ManipulationActive
	manip.StartTime = s.GameTime
	return true
}

// Update processes all active manipulations.
func (s *MarketManipulationSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	for manipID, manip := range s.Manipulations {
		if manip.Status == ManipulationActive {
			s.updateManipulation(w, manipID, manip, dt)
		}
	}
}

// updateManipulation processes a single active manipulation.
func (s *MarketManipulationSystem) updateManipulation(w *ecs.World, manipID string, manip *MarketManipulation, dt float64) {
	// Update progress
	progressDelta := dt / (manip.Duration * 3600.0)
	manip.Progress += progressDelta
	// Accumulate evidence
	manip.Evidence += dt * manip.DetectionRisk * 0.0001
	// Apply price effects while active
	s.applyPriceEffect(w, manip)
	// Check for detection
	if s.pseudoRandom() < manip.Evidence*manip.DetectionRisk*0.001 {
		manip.Status = ManipulationDiscovered
		s.handleDiscovery(manip)
		return
	}
	// Check for completion
	if manip.Progress >= 1.0 {
		s.completeManipulation(manip)
	}
}

// applyPriceEffect modifies market prices based on manipulation.
func (s *MarketManipulationSystem) applyPriceEffect(w *ecs.World, manip *MarketManipulation) {
	comp, ok := w.GetComponent(manip.TargetNode, "EconomyNode")
	if !ok {
		return
	}
	node := comp.(*components.EconomyNode)
	if node.PriceTable == nil {
		return
	}
	basePrice := s.Economy.BasePrices[manip.TargetItem]
	// Gradually shift price toward target
	targetPrice := basePrice * manip.PriceEffect
	currentPrice := node.PriceTable[manip.TargetItem]
	if currentPrice == 0 {
		currentPrice = basePrice
	}
	// Move 1% toward target per update
	diff := targetPrice - currentPrice
	node.PriceTable[manip.TargetItem] = currentPrice + diff*0.01
}

// completeManipulation finalizes a successful manipulation.
func (s *MarketManipulationSystem) completeManipulation(manip *MarketManipulation) {
	// Determine success
	if s.pseudoRandom() < manip.SuccessChance {
		manip.Status = ManipulationSucceeded
		s.recordOutcome(manip, true)
	} else {
		manip.Status = ManipulationFailed
		s.recordOutcome(manip, false)
	}
	// Set cooldown
	cooldownDuration := manip.Duration * 2 * 3600.0
	s.Cooldowns[manip.Manipulator] = s.GameTime + cooldownDuration
}

// handleDiscovery processes a discovered manipulation.
func (s *MarketManipulationSystem) handleDiscovery(manip *MarketManipulation) {
	outcome := ManipulationOutcome{
		Success:       false,
		Profit:        -manip.Investment, // Lose investment
		Discovered:    true,
		Penalty:       manip.Investment * 2, // Double investment as fine
		ReputationHit: 20.0,
		CrimeID:       "market_manipulation_" + manip.ID,
	}
	s.recordOutcomeStruct(manip.Manipulator, outcome)
}

// recordOutcome records the outcome of a manipulation.
func (s *MarketManipulationSystem) recordOutcome(manip *MarketManipulation, success bool) {
	profit := 0.0
	if success {
		// Calculate profit based on type and investment
		profit = manip.Investment * s.getProfitMultiplier(manip.Type)
	} else {
		// Partial loss on failure
		profit = -manip.Investment * 0.5
	}
	outcome := ManipulationOutcome{
		Success:       success,
		Profit:        profit,
		Discovered:    false,
		Penalty:       0,
		ReputationHit: 0,
	}
	s.recordOutcomeStruct(manip.Manipulator, outcome)
}

// recordOutcomeStruct records a manipulation outcome.
func (s *MarketManipulationSystem) recordOutcomeStruct(entity ecs.Entity, outcome ManipulationOutcome) {
	if s.History[entity] == nil {
		s.History[entity] = make([]ManipulationOutcome, 0)
	}
	s.History[entity] = append(s.History[entity], outcome)
	// Keep last 20 outcomes
	if len(s.History[entity]) > 20 {
		s.History[entity] = s.History[entity][1:]
	}
}

// getProfitMultiplier returns expected profit multiplier for manipulation type.
func (s *MarketManipulationSystem) getProfitMultiplier(mType ManipulationType) float64 {
	switch mType {
	case ManipulationCorner:
		return 2.5
	case ManipulationDump:
		return 1.8
	case ManipulationRumor:
		return 2.0
	case ManipulationSabotage:
		return 1.5
	case ManipulationMonopoly:
		return 5.0
	case ManipulationCartel:
		return 2.2
	case ManipulationInsider:
		return 3.0
	case ManipulationCounterfeit:
		return 1.7
	default:
		return 1.5
	}
}

// GetManipulationStatus returns the current status of a manipulation.
func (s *MarketManipulationSystem) GetManipulationStatus(manipID string) (*MarketManipulation, bool) {
	manip, exists := s.Manipulations[manipID]
	return manip, exists
}

// GetActiveManipulations returns all active manipulations by a player.
func (s *MarketManipulationSystem) GetActiveManipulations(manipulator ecs.Entity) []*MarketManipulation {
	result := make([]*MarketManipulation, 0)
	for _, manip := range s.Manipulations {
		if manip.Manipulator == manipulator && manip.Status == ManipulationActive {
			result = append(result, manip)
		}
	}
	return result
}

// GetHistory returns manipulation history for a player.
func (s *MarketManipulationSystem) GetHistory(manipulator ecs.Entity) []ManipulationOutcome {
	return s.History[manipulator]
}

// SetManipulatorSkill sets a player's manipulation skill level.
func (s *MarketManipulationSystem) SetManipulatorSkill(entity ecs.Entity, skill float64) {
	s.ManipulatorSkill[entity] = clampFloat(skill, 0, 1)
}

// SetDetectionCapability sets a market node's detection capability.
func (s *MarketManipulationSystem) SetDetectionCapability(node ecs.Entity, capability float64) {
	s.DetectionAgents[node] = clampFloat(capability, 0, 1)
}

// CancelManipulation cancels a manipulation in planning phase.
func (s *MarketManipulationSystem) CancelManipulation(manipID string) (float64, bool) {
	manip, exists := s.Manipulations[manipID]
	if !exists {
		return 0, false
	}
	if manip.Status != ManipulationPlanning {
		return 0, false // Can only cancel during planning
	}
	// Refund 80% of investment
	refund := manip.Investment * 0.8
	delete(s.Manipulations, manipID)
	return refund, true
}

// GetManipulationDescription returns a genre-appropriate description.
func (s *MarketManipulationSystem) GetManipulationDescription(mType ManipulationType) string {
	descriptions := s.getGenreDescriptions()
	if desc, ok := descriptions[mType]; ok {
		return desc
	}
	return "Unknown manipulation type"
}

// getGenreDescriptions returns genre-specific manipulation descriptions.
func (s *MarketManipulationSystem) getGenreDescriptions() map[ManipulationType]string {
	switch s.Genre {
	case "fantasy":
		return map[ManipulationType]string{
			ManipulationCorner:      "Buy up all supplies of a rare potion ingredient to control its market.",
			ManipulationDump:        "Flood the market with stockpiled goods to crash prices.",
			ManipulationRumor:       "Spread tales through taverns about a coming shortage or surplus.",
			ManipulationSabotage:    "Hire rogues to disrupt a competitor's caravans.",
			ManipulationMonopoly:    "Establish exclusive guild contracts for a resource.",
			ManipulationCartel:      "Form a merchants' compact to fix prices.",
			ManipulationInsider:     "Trade on advance knowledge of kingdom decrees.",
			ManipulationCounterfeit: "Introduce illusion-enchanted fake goods.",
		}
	case "sci-fi":
		return map[ManipulationType]string{
			ManipulationCorner:      "Purchase all available supply contracts for critical components.",
			ManipulationDump:        "Release stockpiled inventory to destabilize market prices.",
			ManipulationRumor:       "Leak false manufacturing data to financial networks.",
			ManipulationSabotage:    "Hire hackers to disrupt competitor logistics.",
			ManipulationMonopoly:    "Acquire exclusive mining rights across the sector.",
			ManipulationCartel:      "Coordinate with other corporations on price floors.",
			ManipulationInsider:     "Trade on classified colony development plans.",
			ManipulationCounterfeit: "Manufacture indistinguishable knockoff goods.",
		}
	case "horror":
		return map[ManipulationType]string{
			ManipulationCorner:      "Hoard protective talismans as fear spreads.",
			ManipulationDump:        "Release cursed goods to destabilize the market.",
			ManipulationRumor:       "Whisper of dark omens affecting certain goods.",
			ManipulationSabotage:    "Send entities to plague a competitor's operations.",
			ManipulationMonopoly:    "Make dark bargains for exclusive access to supplies.",
			ManipulationCartel:      "Form a cult-backed trading syndicate.",
			ManipulationInsider:     "Trade on knowledge gained from forbidden rituals.",
			ManipulationCounterfeit: "Sell cursed items disguised as genuine goods.",
		}
	case "cyberpunk":
		return map[ManipulationType]string{
			ManipulationCorner:      "Execute algorithmic trades to corner the market.",
			ManipulationDump:        "Trigger an automated sell-off to crash prices.",
			ManipulationRumor:       "Plant fake news stories in media feeds.",
			ManipulationSabotage:    "Deploy netrunners against competitor infrastructure.",
			ManipulationMonopoly:    "Leverage corporate influence for exclusive contracts.",
			ManipulationCartel:      "Coordinate zaibatsu price-fixing protocols.",
			ManipulationInsider:     "Trade on stolen corporate intelligence.",
			ManipulationCounterfeit: "3D-print knockoffs with forged authenticity chips.",
		}
	case "post-apocalyptic":
		return map[ManipulationType]string{
			ManipulationCorner:      "Seize all available stockpiles of vital supplies.",
			ManipulationDump:        "Flood the market with scavenged goods.",
			ManipulationRumor:       "Spread word of contamination or bounty at certain traders.",
			ManipulationSabotage:    "Raid a rival settlement's supply depot.",
			ManipulationMonopoly:    "Control the only source of clean water in the region.",
			ManipulationCartel:      "Ally with other warlords to fix trade rates.",
			ManipulationInsider:     "Trade on scout reports of resource discoveries.",
			ManipulationCounterfeit: "Pass off irradiated goods as clean.",
		}
	default:
		return map[ManipulationType]string{
			ManipulationCorner:      "Buy up all available supply to control prices.",
			ManipulationDump:        "Flood the market to crash prices.",
			ManipulationRumor:       "Spread false information about supply or demand.",
			ManipulationSabotage:    "Disrupt competitor operations.",
			ManipulationMonopoly:    "Establish exclusive control over a resource.",
			ManipulationCartel:      "Form price-fixing agreements.",
			ManipulationInsider:     "Trade on non-public information.",
			ManipulationCounterfeit: "Introduce fake goods to the market.",
		}
	}
}

// EstimateOutcome estimates the expected outcome of a manipulation.
func (s *MarketManipulationSystem) EstimateOutcome(mType ManipulationType, investment, skill float64) ManipulationOutcome {
	baseSuccess := s.getBaseSuccessChance(mType)
	skillBonus := skill * 0.2
	successChance := clampFloat(baseSuccess+skillBonus, 0.1, 0.95)
	detectionRisk := s.getBaseDetectionRisk(mType)
	profitMult := s.getProfitMultiplier(mType)
	// Expected value calculation
	expectedProfit := investment * profitMult * successChance
	expectedLoss := investment * 0.5 * (1 - successChance)
	discoveryPenalty := investment * 2 * detectionRisk * (1 - successChance)
	return ManipulationOutcome{
		Success:       successChance > 0.5,
		Profit:        expectedProfit - expectedLoss - discoveryPenalty,
		Discovered:    detectionRisk > 0.5,
		Penalty:       discoveryPenalty,
		ReputationHit: 20.0 * detectionRisk,
	}
}

// GetCooldownRemaining returns remaining cooldown time for a player.
func (s *MarketManipulationSystem) GetCooldownRemaining(entity ecs.Entity) float64 {
	if s.Cooldowns[entity] <= s.GameTime {
		return 0
	}
	return s.Cooldowns[entity] - s.GameTime
}

// Error types for market manipulation.
type ManipulationError string

func (e ManipulationError) Error() string { return string(e) }

const (
	ErrManipulationCooldown    ManipulationError = "manipulation on cooldown"
	ErrInsufficientInvestment  ManipulationError = "insufficient investment"
	ErrInvalidManipulationType ManipulationError = "invalid manipulation type"
)
