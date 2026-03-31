package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// InvestmentType represents types of investments.
type InvestmentType int

const (
	InvestmentBusiness InvestmentType = iota
	InvestmentProperty
	InvestmentCommodity
	InvestmentGuild
	InvestmentExpedition
	InvestmentResearch
	InvestmentBond
	InvestmentVenture
)

// InvestmentRisk represents risk levels.
type InvestmentRisk int

const (
	RiskLow InvestmentRisk = iota
	RiskMedium
	RiskHigh
	RiskSpeculative
)

// Investment represents an active investment.
type Investment struct {
	ID           string
	InvestorID   ecs.Entity
	Type         InvestmentType
	Name         string
	Principal    float64 // Initial investment
	CurrentValue float64 // Current market value
	ReturnRate   float64 // Annual return rate
	Risk         InvestmentRisk
	Maturity     float64 // Time until maturity (hours)
	TimeHeld     float64 // Time investment has been held
	Status       InvestmentStatus
	Dividends    float64 // Accumulated dividends
	DividendRate float64 // Periodic dividend rate
	LastDividend float64 // Time of last dividend payment
	Volatility   float64 // Price volatility (0-1)
	MinHoldTime  float64 // Minimum hold time (hours)
}

// InvestmentStatus represents the state of an investment.
type InvestmentStatus int

const (
	InvestmentActive InvestmentStatus = iota
	InvestmentMatured
	InvestmentSold
	InvestmentLost
	InvestmentFrozen
)

// InvestmentOpportunity represents an available investment.
type InvestmentOpportunity struct {
	ID             string
	Type           InvestmentType
	Name           string
	Description    string
	MinInvestment  float64
	MaxInvestment  float64
	ExpectedReturn float64
	Risk           InvestmentRisk
	MaturityPeriod float64
	DividendRate   float64
	Volatility     float64
	Available      bool
	ExpiresAt      float64 // Game time when opportunity expires
}

// InvestmentSystem manages player investments.
type InvestmentSystem struct {
	Seed           int64
	Genre          string
	Investments    map[string]*Investment
	Opportunities  map[string]*InvestmentOpportunity
	GameTime       float64
	counter        uint64
	InvestorSkill  map[ecs.Entity]float64  // Skill affects returns
	Portfolio      map[ecs.Entity][]string // Investor -> investment IDs
	TotalReturns   map[ecs.Entity]float64  // Lifetime returns
	DividendPeriod float64                 // Hours between dividend payments
}

// NewInvestmentSystem creates a new investment system.
func NewInvestmentSystem(seed int64, genre string) *InvestmentSystem {
	s := &InvestmentSystem{
		Seed:           seed,
		Genre:          genre,
		Investments:    make(map[string]*Investment),
		Opportunities:  make(map[string]*InvestmentOpportunity),
		InvestorSkill:  make(map[ecs.Entity]float64),
		Portfolio:      make(map[ecs.Entity][]string),
		TotalReturns:   make(map[ecs.Entity]float64),
		DividendPeriod: 24.0, // Daily dividends
	}
	s.generateOpportunities()
	return s
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *InvestmentSystem) pseudoRandom() float64 {
	s.counter++
	x := uint64(s.Seed) + s.counter*6364136223846793005
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	return float64(x%10000) / 10000.0
}

// generateOpportunities creates initial investment opportunities.
func (s *InvestmentSystem) generateOpportunities() {
	opportunities := s.getGenreOpportunities()
	for _, opp := range opportunities {
		s.Opportunities[opp.ID] = opp
	}
}

// getGenreOpportunities returns genre-specific investment opportunities.
func (s *InvestmentSystem) getGenreOpportunities() []*InvestmentOpportunity {
	switch s.Genre {
	case "fantasy":
		return []*InvestmentOpportunity{
			{ID: "tavern_stake", Type: InvestmentBusiness, Name: "Tavern Partnership", Description: "Share in a profitable inn and tavern.", MinInvestment: 500, MaxInvestment: 5000, ExpectedReturn: 0.12, Risk: RiskLow, MaturityPeriod: 168, DividendRate: 0.01, Volatility: 0.1, Available: true},
			{ID: "mine_shares", Type: InvestmentBusiness, Name: "Mining Guild Shares", Description: "Invest in a dwarven mining operation.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.18, Risk: RiskMedium, MaturityPeriod: 336, DividendRate: 0.02, Volatility: 0.25, Available: true},
			{ID: "caravan_venture", Type: InvestmentExpedition, Name: "Trade Caravan Venture", Description: "Fund a merchant caravan to distant lands.", MinInvestment: 2000, MaxInvestment: 20000, ExpectedReturn: 0.35, Risk: RiskHigh, MaturityPeriod: 504, DividendRate: 0, Volatility: 0.4, Available: true},
			{ID: "magic_academy", Type: InvestmentResearch, Name: "Academy of Arcane Arts", Description: "Support magical research with potential patents.", MinInvestment: 5000, MaxInvestment: 50000, ExpectedReturn: 0.5, Risk: RiskSpeculative, MaturityPeriod: 720, DividendRate: 0, Volatility: 0.6, Available: true},
			{ID: "land_deed", Type: InvestmentProperty, Name: "Rural Estate", Description: "Purchase farmland for rental income.", MinInvestment: 3000, MaxInvestment: 30000, ExpectedReturn: 0.08, Risk: RiskLow, MaturityPeriod: 0, DividendRate: 0.015, Volatility: 0.05, Available: true},
			{ID: "guild_bond", Type: InvestmentBond, Name: "Merchants' Guild Bond", Description: "Secure loan to the merchants' guild.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.06, Risk: RiskLow, MaturityPeriod: 168, DividendRate: 0.005, Volatility: 0.02, Available: true},
		}
	case "sci-fi":
		return []*InvestmentOpportunity{
			{ID: "station_shares", Type: InvestmentBusiness, Name: "Station Commerce Hub", Description: "Equity in an orbital trading post.", MinInvestment: 500, MaxInvestment: 5000, ExpectedReturn: 0.12, Risk: RiskLow, MaturityPeriod: 168, DividendRate: 0.01, Volatility: 0.1, Available: true},
			{ID: "mining_corp", Type: InvestmentBusiness, Name: "Asteroid Mining Corp", Description: "Shares in a mining corporation.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.2, Risk: RiskMedium, MaturityPeriod: 336, DividendRate: 0.02, Volatility: 0.3, Available: true},
			{ID: "colony_expedition", Type: InvestmentExpedition, Name: "Colony Ship Venture", Description: "Fund a new colony establishment.", MinInvestment: 2000, MaxInvestment: 20000, ExpectedReturn: 0.4, Risk: RiskHigh, MaturityPeriod: 504, DividendRate: 0, Volatility: 0.45, Available: true},
			{ID: "biotech_startup", Type: InvestmentResearch, Name: "Biotech Startup", Description: "Invest in cutting-edge genetic research.", MinInvestment: 5000, MaxInvestment: 50000, ExpectedReturn: 0.6, Risk: RiskSpeculative, MaturityPeriod: 720, DividendRate: 0, Volatility: 0.7, Available: true},
			{ID: "hab_module", Type: InvestmentProperty, Name: "Habitat Module", Description: "Own a residential hab for rental income.", MinInvestment: 3000, MaxInvestment: 30000, ExpectedReturn: 0.09, Risk: RiskLow, MaturityPeriod: 0, DividendRate: 0.018, Volatility: 0.06, Available: true},
			{ID: "corp_bond", Type: InvestmentBond, Name: "Corporate Bond", Description: "Secure loan to a megacorporation.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.05, Risk: RiskLow, MaturityPeriod: 168, DividendRate: 0.004, Volatility: 0.02, Available: true},
		}
	case "horror":
		return []*InvestmentOpportunity{
			{ID: "occult_shop", Type: InvestmentBusiness, Name: "Occult Curiosities Shop", Description: "Partnership in a shop selling protective items.", MinInvestment: 500, MaxInvestment: 5000, ExpectedReturn: 0.15, Risk: RiskMedium, MaturityPeriod: 168, DividendRate: 0.012, Volatility: 0.2, Available: true},
			{ID: "silver_mine", Type: InvestmentBusiness, Name: "Silver Mining Operation", Description: "Shares in a silver mine (high demand for wards).", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.22, Risk: RiskMedium, MaturityPeriod: 336, DividendRate: 0.025, Volatility: 0.3, Available: true},
			{ID: "artifact_hunt", Type: InvestmentExpedition, Name: "Artifact Recovery Expedition", Description: "Fund a dangerous expedition to recover relics.", MinInvestment: 2000, MaxInvestment: 20000, ExpectedReturn: 0.5, Risk: RiskHigh, MaturityPeriod: 504, DividendRate: 0, Volatility: 0.55, Available: true},
			{ID: "ward_research", Type: InvestmentResearch, Name: "Ward Research Institute", Description: "Support research into protective magic.", MinInvestment: 5000, MaxInvestment: 50000, ExpectedReturn: 0.45, Risk: RiskSpeculative, MaturityPeriod: 720, DividendRate: 0, Volatility: 0.5, Available: true},
			{ID: "safe_house", Type: InvestmentProperty, Name: "Warded Safe House", Description: "Own a protected shelter for rent.", MinInvestment: 3000, MaxInvestment: 30000, ExpectedReturn: 0.1, Risk: RiskLow, MaturityPeriod: 0, DividendRate: 0.02, Volatility: 0.08, Available: true},
			{ID: "sanctuary_bond", Type: InvestmentBond, Name: "Sanctuary Bond", Description: "Loan to fund community defenses.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.07, Risk: RiskLow, MaturityPeriod: 168, DividendRate: 0.006, Volatility: 0.03, Available: true},
		}
	case "cyberpunk":
		return []*InvestmentOpportunity{
			{ID: "club_shares", Type: InvestmentBusiness, Name: "Nightclub Franchise", Description: "Stake in a popular entertainment venue.", MinInvestment: 500, MaxInvestment: 5000, ExpectedReturn: 0.14, Risk: RiskMedium, MaturityPeriod: 168, DividendRate: 0.015, Volatility: 0.2, Available: true},
			{ID: "corp_stock", Type: InvestmentBusiness, Name: "Zaibatsu Stock", Description: "Shares in a powerful megacorporation.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.16, Risk: RiskMedium, MaturityPeriod: 336, DividendRate: 0.018, Volatility: 0.25, Available: true},
			{ID: "heist_fund", Type: InvestmentExpedition, Name: "High-Risk Heist Fund", Description: "Back a crew for a major data heist.", MinInvestment: 2000, MaxInvestment: 20000, ExpectedReturn: 0.6, Risk: RiskHigh, MaturityPeriod: 168, DividendRate: 0, Volatility: 0.7, Available: true},
			{ID: "cyberware_rd", Type: InvestmentResearch, Name: "Cyberware R&D", Description: "Fund next-gen implant development.", MinInvestment: 5000, MaxInvestment: 50000, ExpectedReturn: 0.55, Risk: RiskSpeculative, MaturityPeriod: 720, DividendRate: 0, Volatility: 0.65, Available: true},
			{ID: "condo_unit", Type: InvestmentProperty, Name: "Megablock Condo Unit", Description: "Own residential space for rent.", MinInvestment: 3000, MaxInvestment: 30000, ExpectedReturn: 0.08, Risk: RiskLow, MaturityPeriod: 0, DividendRate: 0.016, Volatility: 0.05, Available: true},
			{ID: "syndicate_loan", Type: InvestmentBond, Name: "Syndicate Loan", Description: "Lend to an organized crime syndicate.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.12, Risk: RiskMedium, MaturityPeriod: 168, DividendRate: 0.01, Volatility: 0.15, Available: true},
		}
	case "post-apocalyptic":
		return []*InvestmentOpportunity{
			{ID: "trading_post", Type: InvestmentBusiness, Name: "Wasteland Trading Post", Description: "Share in a fortified trade hub.", MinInvestment: 500, MaxInvestment: 5000, ExpectedReturn: 0.18, Risk: RiskMedium, MaturityPeriod: 168, DividendRate: 0.02, Volatility: 0.25, Available: true},
			{ID: "water_purifier", Type: InvestmentBusiness, Name: "Water Purification Plant", Description: "Invest in clean water production.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.25, Risk: RiskMedium, MaturityPeriod: 336, DividendRate: 0.03, Volatility: 0.2, Available: true},
			{ID: "scav_expedition", Type: InvestmentExpedition, Name: "Scavenger Expedition", Description: "Fund a scavenging run into the ruins.", MinInvestment: 2000, MaxInvestment: 20000, ExpectedReturn: 0.45, Risk: RiskHigh, MaturityPeriod: 336, DividendRate: 0, Volatility: 0.5, Available: true},
			{ID: "tech_recovery", Type: InvestmentResearch, Name: "Tech Recovery Project", Description: "Support efforts to recover old-world tech.", MinInvestment: 5000, MaxInvestment: 50000, ExpectedReturn: 0.55, Risk: RiskSpeculative, MaturityPeriod: 720, DividendRate: 0, Volatility: 0.6, Available: true},
			{ID: "bunker_space", Type: InvestmentProperty, Name: "Bunker Living Space", Description: "Own secure shelter for rent.", MinInvestment: 3000, MaxInvestment: 30000, ExpectedReturn: 0.12, Risk: RiskLow, MaturityPeriod: 0, DividendRate: 0.025, Volatility: 0.08, Available: true},
			{ID: "settlement_bond", Type: InvestmentBond, Name: "Settlement Development Bond", Description: "Fund settlement infrastructure.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.08, Risk: RiskLow, MaturityPeriod: 168, DividendRate: 0.007, Volatility: 0.04, Available: true},
		}
	default:
		return []*InvestmentOpportunity{
			{ID: "business", Type: InvestmentBusiness, Name: "Local Business", Description: "Invest in a local business.", MinInvestment: 500, MaxInvestment: 5000, ExpectedReturn: 0.12, Risk: RiskLow, MaturityPeriod: 168, DividendRate: 0.01, Volatility: 0.1, Available: true},
			{ID: "property", Type: InvestmentProperty, Name: "Property", Description: "Buy property for rental income.", MinInvestment: 3000, MaxInvestment: 30000, ExpectedReturn: 0.08, Risk: RiskLow, MaturityPeriod: 0, DividendRate: 0.015, Volatility: 0.05, Available: true},
			{ID: "bond", Type: InvestmentBond, Name: "Secure Bond", Description: "Low-risk secured loan.", MinInvestment: 1000, MaxInvestment: 10000, ExpectedReturn: 0.06, Risk: RiskLow, MaturityPeriod: 168, DividendRate: 0.005, Volatility: 0.02, Available: true},
		}
	}
}

// GetAvailableOpportunities returns currently available investment opportunities.
func (s *InvestmentSystem) GetAvailableOpportunities() []*InvestmentOpportunity {
	result := make([]*InvestmentOpportunity, 0)
	for _, opp := range s.Opportunities {
		if opp.Available && (opp.ExpiresAt == 0 || opp.ExpiresAt > s.GameTime) {
			result = append(result, opp)
		}
	}
	return result
}

// Invest makes an investment in an opportunity.
func (s *InvestmentSystem) Invest(investor ecs.Entity, opportunityID string, amount float64) (*Investment, error) {
	opp, exists := s.Opportunities[opportunityID]
	if !exists || !opp.Available {
		return nil, ErrInvestmentNotAvailable
	}
	if amount < opp.MinInvestment {
		return nil, ErrInvestmentTooSmall
	}
	if amount > opp.MaxInvestment {
		return nil, ErrInvestmentTooLarge
	}
	investID := s.generateInvestmentID()
	skill := s.InvestorSkill[investor]
	// Skill improves expected return slightly
	adjustedReturn := opp.ExpectedReturn * (1.0 + skill*0.1)
	investment := &Investment{
		ID:           investID,
		InvestorID:   investor,
		Type:         opp.Type,
		Name:         opp.Name,
		Principal:    amount,
		CurrentValue: amount,
		ReturnRate:   adjustedReturn,
		Risk:         opp.Risk,
		Maturity:     opp.MaturityPeriod,
		TimeHeld:     0,
		Status:       InvestmentActive,
		Dividends:    0,
		DividendRate: opp.DividendRate,
		LastDividend: s.GameTime,
		Volatility:   opp.Volatility,
		MinHoldTime:  s.getMinHoldTime(opp.Type),
	}
	s.Investments[investID] = investment
	s.addToPortfolio(investor, investID)
	return investment, nil
}

// generateInvestmentID creates a unique investment identifier.
func (s *InvestmentSystem) generateInvestmentID() string {
	s.counter++
	return "invest_" + string(rune('0'+s.counter%1000))
}

// getMinHoldTime returns minimum hold time for investment type.
func (s *InvestmentSystem) getMinHoldTime(iType InvestmentType) float64 {
	switch iType {
	case InvestmentBond:
		return 0 // Bonds can be sold anytime (with penalty)
	case InvestmentProperty:
		return 168 // 1 week for property
	case InvestmentExpedition:
		return 0 // Can't sell expeditions early
	default:
		return 24 // 1 day default
	}
}

// addToPortfolio adds an investment to an investor's portfolio.
func (s *InvestmentSystem) addToPortfolio(investor ecs.Entity, investID string) {
	if s.Portfolio[investor] == nil {
		s.Portfolio[investor] = make([]string, 0)
	}
	s.Portfolio[investor] = append(s.Portfolio[investor], investID)
}

// removeFromPortfolio removes an investment from an investor's portfolio.
func (s *InvestmentSystem) removeFromPortfolio(investor ecs.Entity, investID string) {
	if s.Portfolio[investor] == nil {
		return
	}
	newPortfolio := make([]string, 0)
	for _, id := range s.Portfolio[investor] {
		if id != investID {
			newPortfolio = append(newPortfolio, id)
		}
	}
	s.Portfolio[investor] = newPortfolio
}

// Update processes all active investments.
func (s *InvestmentSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	hoursDelta := dt / 3600.0
	for _, investment := range s.Investments {
		if investment.Status == InvestmentActive {
			s.updateInvestment(investment, hoursDelta)
		}
	}
	// Occasionally generate new opportunities
	if s.pseudoRandom() < 0.0001 {
		s.generateNewOpportunity()
	}
}

// updateInvestment processes a single investment.
func (s *InvestmentSystem) updateInvestment(investment *Investment, hours float64) {
	investment.TimeHeld += hours
	// Apply market volatility
	volatilityEffect := (s.pseudoRandom() - 0.5) * 2 * investment.Volatility
	// Base growth rate (hourly)
	hourlyReturn := investment.ReturnRate / (365 * 24)
	// Apply changes
	growth := 1.0 + hourlyReturn + volatilityEffect*0.01
	investment.CurrentValue *= growth
	// Pay dividends
	if investment.DividendRate > 0 {
		hoursSinceDiv := s.GameTime - investment.LastDividend
		if hoursSinceDiv >= s.DividendPeriod {
			dividend := investment.Principal * investment.DividendRate
			investment.Dividends += dividend
			investment.LastDividend = s.GameTime
		}
	}
	// Check for risk events
	s.checkRiskEvents(investment)
	// Check maturity
	if investment.Maturity > 0 && investment.TimeHeld >= investment.Maturity {
		investment.Status = InvestmentMatured
	}
}

// checkRiskEvents checks for negative risk events.
func (s *InvestmentSystem) checkRiskEvents(investment *Investment) {
	riskChance := s.getRiskEventChance(investment.Risk)
	if s.pseudoRandom() < riskChance {
		// Risk event occurred
		loss := s.pseudoRandom() * 0.3 // Up to 30% loss
		investment.CurrentValue *= (1.0 - loss)
		// Severe risk can destroy investment
		if investment.CurrentValue < investment.Principal*0.1 {
			investment.Status = InvestmentLost
		}
	}
}

// getRiskEventChance returns chance of negative event per hour.
func (s *InvestmentSystem) getRiskEventChance(risk InvestmentRisk) float64 {
	switch risk {
	case RiskLow:
		return 0.00001
	case RiskMedium:
		return 0.0001
	case RiskHigh:
		return 0.001
	case RiskSpeculative:
		return 0.005
	default:
		return 0.0001
	}
}

// Sell liquidates an investment.
func (s *InvestmentSystem) Sell(investID string) (float64, error) {
	investment, exists := s.Investments[investID]
	if !exists {
		return 0, ErrInvestmentNotFound
	}
	if investment.Status != InvestmentActive && investment.Status != InvestmentMatured {
		return 0, ErrInvestmentNotSellable
	}
	// Check minimum hold time
	if investment.TimeHeld < investment.MinHoldTime {
		return 0, ErrInvestmentMinHoldTime
	}
	// Calculate proceeds
	proceeds := investment.CurrentValue + investment.Dividends
	// Early sale penalty (if not matured and not property)
	if investment.Status != InvestmentMatured && investment.Maturity > 0 {
		penalty := 0.1 // 10% early withdrawal penalty
		proceeds *= (1.0 - penalty)
	}
	// Update totals
	profit := proceeds - investment.Principal
	s.TotalReturns[investment.InvestorID] += profit
	// Update status
	investment.Status = InvestmentSold
	s.removeFromPortfolio(investment.InvestorID, investID)
	return proceeds, nil
}

// GetPortfolio returns an investor's current investments.
func (s *InvestmentSystem) GetPortfolio(investor ecs.Entity) []*Investment {
	result := make([]*Investment, 0)
	for _, investID := range s.Portfolio[investor] {
		if inv, exists := s.Investments[investID]; exists {
			result = append(result, inv)
		}
	}
	return result
}

// GetPortfolioValue returns total value of an investor's portfolio.
func (s *InvestmentSystem) GetPortfolioValue(investor ecs.Entity) float64 {
	total := 0.0
	for _, investID := range s.Portfolio[investor] {
		if inv, exists := s.Investments[investID]; exists {
			if inv.Status == InvestmentActive || inv.Status == InvestmentMatured {
				total += inv.CurrentValue + inv.Dividends
			}
		}
	}
	return total
}

// GetTotalReturns returns an investor's lifetime returns.
func (s *InvestmentSystem) GetTotalReturns(investor ecs.Entity) float64 {
	return s.TotalReturns[investor]
}

// SetInvestorSkill sets an investor's skill level.
func (s *InvestmentSystem) SetInvestorSkill(investor ecs.Entity, skill float64) {
	s.InvestorSkill[investor] = clampFloat(skill, 0, 1)
}

// CollectDividends withdraws accumulated dividends.
func (s *InvestmentSystem) CollectDividends(investID string) float64 {
	investment, exists := s.Investments[investID]
	if !exists || investment.Status != InvestmentActive {
		return 0
	}
	dividends := investment.Dividends
	investment.Dividends = 0
	s.TotalReturns[investment.InvestorID] += dividends
	return dividends
}

// generateNewOpportunity creates a new random opportunity.
func (s *InvestmentSystem) generateNewOpportunity() {
	opportunities := s.getGenreOpportunities()
	if len(opportunities) == 0 {
		return
	}
	base := opportunities[int(s.pseudoRandom()*float64(len(opportunities)))]
	newOpp := &InvestmentOpportunity{
		ID:             base.ID + "_" + string(rune('0'+s.counter%100)),
		Type:           base.Type,
		Name:           base.Name + " (Limited)",
		Description:    base.Description,
		MinInvestment:  base.MinInvestment * (0.8 + s.pseudoRandom()*0.4),
		MaxInvestment:  base.MaxInvestment * (0.8 + s.pseudoRandom()*0.4),
		ExpectedReturn: base.ExpectedReturn * (0.9 + s.pseudoRandom()*0.2),
		Risk:           base.Risk,
		MaturityPeriod: base.MaturityPeriod,
		DividendRate:   base.DividendRate,
		Volatility:     base.Volatility,
		Available:      true,
		ExpiresAt:      s.GameTime + 168.0, // Available for 1 week
	}
	s.Opportunities[newOpp.ID] = newOpp
}

// GetInvestment returns a specific investment.
func (s *InvestmentSystem) GetInvestment(investID string) (*Investment, bool) {
	inv, exists := s.Investments[investID]
	return inv, exists
}

// GetRiskDescription returns a description of risk level.
func (s *InvestmentSystem) GetRiskDescription(risk InvestmentRisk) string {
	switch risk {
	case RiskLow:
		return "Low risk - stable returns with minimal chance of loss"
	case RiskMedium:
		return "Medium risk - moderate volatility with potential for gains or losses"
	case RiskHigh:
		return "High risk - significant volatility, potential for large gains or losses"
	case RiskSpeculative:
		return "Speculative - extremely volatile, potential for total loss or exceptional gains"
	default:
		return "Unknown risk level"
	}
}

// CalculateProjectedReturn calculates projected return for an investment.
func (s *InvestmentSystem) CalculateProjectedReturn(opportunityID string, amount, holdPeriod float64) float64 {
	opp, exists := s.Opportunities[opportunityID]
	if !exists {
		return 0
	}
	// Annual return prorated to hold period
	years := holdPeriod / (365 * 24)
	baseReturn := amount * opp.ExpectedReturn * years
	// Add estimated dividends
	dividendPeriods := holdPeriod / s.DividendPeriod
	dividends := amount * opp.DividendRate * dividendPeriods
	return baseReturn + dividends
}

// Investment error types.
type InvestmentError string

// Error returns the error message for the InvestmentError.
func (e InvestmentError) Error() string { return string(e) }

const (
	ErrInvestmentNotAvailable InvestmentError = "investment opportunity not available"
	ErrInvestmentTooSmall     InvestmentError = "investment amount below minimum"
	ErrInvestmentTooLarge     InvestmentError = "investment amount above maximum"
	ErrInvestmentNotFound     InvestmentError = "investment not found"
	ErrInvestmentNotSellable  InvestmentError = "investment cannot be sold"
	ErrInvestmentMinHoldTime  InvestmentError = "minimum hold time not met"
)
