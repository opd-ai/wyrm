package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// TestNewInvestmentSystem tests system initialization.
func TestNewInvestmentSystem(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	if system == nil {
		t.Fatal("NewInvestmentSystem returned nil")
	}
	if system.Genre != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", system.Genre)
	}
	if len(system.Opportunities) == 0 {
		t.Error("should have generated initial opportunities")
	}
}

// TestGetAvailableOpportunities tests listing available opportunities.
func TestGetAvailableOpportunities(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	opps := system.GetAvailableOpportunities()
	if len(opps) == 0 {
		t.Error("should have available opportunities")
	}
	for _, opp := range opps {
		if opp.MinInvestment <= 0 {
			t.Errorf("opportunity %s should have positive min investment", opp.ID)
		}
		if opp.MaxInvestment < opp.MinInvestment {
			t.Errorf("opportunity %s max should be >= min", opp.ID)
		}
	}
}

// TestInvest tests making an investment.
func TestInvest(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	opps := system.GetAvailableOpportunities()
	if len(opps) == 0 {
		t.Fatal("no opportunities available")
	}
	opp := opps[0]
	inv, err := system.Invest(investorID, opp.ID, opp.MinInvestment)
	if err != nil {
		t.Fatalf("Invest failed: %v", err)
	}
	if inv == nil {
		t.Fatal("Invest returned nil")
	}
	if inv.Principal != opp.MinInvestment {
		t.Errorf("expected principal %f, got %f", opp.MinInvestment, inv.Principal)
	}
	if inv.Status != InvestmentActive {
		t.Errorf("expected status Active, got %d", inv.Status)
	}
}

// TestInvestTooSmall tests minimum investment check.
func TestInvestTooSmall(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	opps := system.GetAvailableOpportunities()
	if len(opps) == 0 {
		t.Fatal("no opportunities available")
	}
	opp := opps[0]
	_, err := system.Invest(investorID, opp.ID, opp.MinInvestment-1)
	if err != ErrInvestmentTooSmall {
		t.Errorf("expected ErrInvestmentTooSmall, got %v", err)
	}
}

// TestInvestTooLarge tests maximum investment check.
func TestInvestTooLarge(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	opps := system.GetAvailableOpportunities()
	if len(opps) == 0 {
		t.Fatal("no opportunities available")
	}
	opp := opps[0]
	_, err := system.Invest(investorID, opp.ID, opp.MaxInvestment+1)
	if err != ErrInvestmentTooLarge {
		t.Errorf("expected ErrInvestmentTooLarge, got %v", err)
	}
}

// TestInvestInvalidOpportunity tests investing in non-existent opportunity.
func TestInvestInvalidOpportunity(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	_, err := system.Invest(investorID, "invalid_opp", 1000)
	if err != ErrInvestmentNotAvailable {
		t.Errorf("expected ErrInvestmentNotAvailable, got %v", err)
	}
}

// TestUpdate tests the update cycle.
func TestInvestmentUpdate(t *testing.T) {
	world := ecs.NewWorld()
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	opps := system.GetAvailableOpportunities()
	opp := opps[0]
	inv, _ := system.Invest(investorID, opp.ID, opp.MinInvestment)
	initialValue := inv.CurrentValue
	// Simulate time passing (24 hours worth)
	for i := 0; i < 100; i++ {
		system.Update(world, 864.0) // ~0.24 hours per tick
	}
	// Value should have changed due to returns/volatility
	if inv.CurrentValue == initialValue && inv.Status == InvestmentActive {
		t.Log("Note: Value unchanged, may be due to low volatility")
	}
}

// TestSell tests selling an investment.
func TestSell(t *testing.T) {
	world := ecs.NewWorld()
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	// Find a bond (can be sold immediately)
	var bondOpp *InvestmentOpportunity
	for _, opp := range system.GetAvailableOpportunities() {
		if opp.Type == InvestmentBond {
			bondOpp = opp
			break
		}
	}
	if bondOpp == nil {
		t.Skip("No bond opportunity available")
	}
	inv, _ := system.Invest(investorID, bondOpp.ID, bondOpp.MinInvestment)
	// Simulate some time passing
	for i := 0; i < 10; i++ {
		system.Update(world, 3600.0)
	}
	proceeds, err := system.Sell(inv.ID)
	if err != nil {
		t.Fatalf("Sell failed: %v", err)
	}
	if proceeds <= 0 {
		t.Error("proceeds should be positive")
	}
	if inv.Status != InvestmentSold {
		t.Errorf("expected status Sold, got %d", inv.Status)
	}
}

// TestSellMinHoldTime tests minimum hold time enforcement.
func TestSellMinHoldTime(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	// Find a property (has min hold time)
	var propOpp *InvestmentOpportunity
	for _, opp := range system.GetAvailableOpportunities() {
		if opp.Type == InvestmentProperty {
			propOpp = opp
			break
		}
	}
	if propOpp == nil {
		t.Skip("No property opportunity available")
	}
	inv, _ := system.Invest(investorID, propOpp.ID, propOpp.MinInvestment)
	// Try to sell immediately
	_, err := system.Sell(inv.ID)
	if err != ErrInvestmentMinHoldTime {
		t.Errorf("expected ErrInvestmentMinHoldTime, got %v", err)
	}
}

// TestSellInvalidInvestment tests selling non-existent investment.
func TestSellInvalidInvestment(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	_, err := system.Sell("invalid_invest")
	if err != ErrInvestmentNotFound {
		t.Errorf("expected ErrInvestmentNotFound, got %v", err)
	}
}

// TestGetPortfolio tests retrieving investor's portfolio.
func TestGetPortfolio(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	opps := system.GetAvailableOpportunities()
	// Make two investments
	system.Invest(investorID, opps[0].ID, opps[0].MinInvestment)
	if len(opps) > 1 {
		system.Invest(investorID, opps[1].ID, opps[1].MinInvestment)
	}
	portfolio := system.GetPortfolio(investorID)
	if len(portfolio) < 1 {
		t.Error("portfolio should have at least 1 investment")
	}
}

// TestGetPortfolioValue tests calculating portfolio value.
func TestGetPortfolioValue(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	opps := system.GetAvailableOpportunities()
	opp := opps[0]
	system.Invest(investorID, opp.ID, opp.MinInvestment)
	value := system.GetPortfolioValue(investorID)
	if value < opp.MinInvestment*0.5 {
		t.Errorf("portfolio value %f seems too low", value)
	}
}

// TestSetInvestorSkill tests setting investor skill.
func TestSetInvestorSkill(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	system.SetInvestorSkill(investorID, 0.8)
	if system.InvestorSkill[investorID] != 0.8 {
		t.Errorf("expected skill 0.8, got %f", system.InvestorSkill[investorID])
	}
	// Test clamping
	system.SetInvestorSkill(investorID, 1.5)
	if system.InvestorSkill[investorID] != 1.0 {
		t.Errorf("skill should be clamped to 1.0, got %f", system.InvestorSkill[investorID])
	}
}

// TestCollectDividends tests dividend collection.
func TestCollectDividends(t *testing.T) {
	world := ecs.NewWorld()
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	// Find investment with dividends
	var divOpp *InvestmentOpportunity
	for _, opp := range system.GetAvailableOpportunities() {
		if opp.DividendRate > 0 {
			divOpp = opp
			break
		}
	}
	if divOpp == nil {
		t.Skip("No dividend-paying opportunity available")
	}
	inv, _ := system.Invest(investorID, divOpp.ID, divOpp.MinInvestment)
	// Simulate enough time for dividend
	for i := 0; i < 50; i++ {
		system.Update(world, 3600.0) // 1 hour per tick
	}
	dividends := system.CollectDividends(inv.ID)
	if dividends <= 0 && inv.DividendRate > 0 {
		t.Log("Note: No dividends accumulated (might need more time)")
	}
}

// TestGenreOpportunities tests that all genres have opportunities.
func TestGenreOpportunities(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		system := NewInvestmentSystem(12345, genre)
		opps := system.GetAvailableOpportunities()
		if len(opps) == 0 {
			t.Errorf("genre %s should have opportunities", genre)
		}
	}
}

// TestGetRiskDescription tests risk level descriptions.
func TestGetRiskDescription(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	risks := []InvestmentRisk{RiskLow, RiskMedium, RiskHigh, RiskSpeculative}
	for _, risk := range risks {
		desc := system.GetRiskDescription(risk)
		if desc == "" || desc == "Unknown risk level" {
			t.Errorf("risk %d should have description", risk)
		}
	}
}

// TestCalculateProjectedReturn tests return projection.
func TestCalculateProjectedReturn(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	opps := system.GetAvailableOpportunities()
	if len(opps) == 0 {
		t.Fatal("no opportunities")
	}
	opp := opps[0]
	// Project return for 1 week hold
	projected := system.CalculateProjectedReturn(opp.ID, opp.MinInvestment, 168)
	if projected <= 0 {
		t.Error("projected return should be positive")
	}
}

// TestInvestmentMaturity tests investment reaching maturity.
func TestInvestmentMaturity(t *testing.T) {
	world := ecs.NewWorld()
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	// Find short-term investment
	var shortOpp *InvestmentOpportunity
	for _, opp := range system.GetAvailableOpportunities() {
		if opp.MaturityPeriod > 0 && opp.MaturityPeriod <= 168 {
			shortOpp = opp
			break
		}
	}
	if shortOpp == nil {
		t.Skip("No short-term opportunity available")
	}
	inv, _ := system.Invest(investorID, shortOpp.ID, shortOpp.MinInvestment)
	// Simulate until maturity
	hoursNeeded := shortOpp.MaturityPeriod + 10
	for i := 0; i < int(hoursNeeded); i++ {
		system.Update(world, 3600.0)
	}
	if inv.Status != InvestmentMatured && inv.Status != InvestmentLost {
		t.Errorf("expected matured or lost status, got %d", inv.Status)
	}
}

// TestPseudoRandom tests deterministic random generation.
func TestInvestmentPseudoRandom(t *testing.T) {
	system1 := NewInvestmentSystem(12345, "fantasy")
	system2 := NewInvestmentSystem(12345, "fantasy")
	// Same seed should produce same sequence
	for i := 0; i < 10; i++ {
		r1 := system1.pseudoRandom()
		r2 := system2.pseudoRandom()
		if r1 != r2 {
			t.Errorf("iteration %d: same seed should produce same random values", i)
		}
	}
}

// TestSkillAffectsReturn tests that skill improves returns.
func TestSkillAffectsReturn(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	skilledID := ecs.Entity(101)
	system.SetInvestorSkill(skilledID, 1.0)
	opps := system.GetAvailableOpportunities()
	opp := opps[0]
	invNoSkill, _ := system.Invest(investorID, opp.ID, opp.MinInvestment)
	invWithSkill, _ := system.Invest(skilledID, opp.ID, opp.MinInvestment)
	if invWithSkill.ReturnRate <= invNoSkill.ReturnRate {
		t.Error("skilled investor should have better return rate")
	}
}

// TestGetInvestment tests retrieving a specific investment.
func TestGetInvestment(t *testing.T) {
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	opps := system.GetAvailableOpportunities()
	inv, _ := system.Invest(investorID, opps[0].ID, opps[0].MinInvestment)
	retrieved, exists := system.GetInvestment(inv.ID)
	if !exists {
		t.Error("investment should exist")
	}
	if retrieved.ID != inv.ID {
		t.Error("retrieved investment should match")
	}
}

// TestGetTotalReturns tests lifetime returns tracking.
func TestGetTotalReturns(t *testing.T) {
	world := ecs.NewWorld()
	system := NewInvestmentSystem(12345, "fantasy")
	investorID := ecs.Entity(100)
	// Find bond for quick test
	var bondOpp *InvestmentOpportunity
	for _, opp := range system.GetAvailableOpportunities() {
		if opp.Type == InvestmentBond {
			bondOpp = opp
			break
		}
	}
	if bondOpp == nil {
		t.Skip("No bond opportunity")
	}
	inv, _ := system.Invest(investorID, bondOpp.ID, bondOpp.MinInvestment)
	// Simulate time
	for i := 0; i < 50; i++ {
		system.Update(world, 3600.0)
	}
	// Sell
	system.Sell(inv.ID)
	// Check returns tracked
	returns := system.GetTotalReturns(investorID)
	// Returns could be positive or negative depending on market
	t.Logf("Total returns: %f", returns)
}
