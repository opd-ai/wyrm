package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewEconomySystem(t *testing.T) {
	tests := []struct {
		name        string
		fluctuation float64
		normRate    float64
	}{
		{"default", 0.5, 0.1},
		{"high fluctuation", 1.0, 0.5},
		{"zero values", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := NewEconomySystem(tt.fluctuation, tt.normRate)
			if es == nil {
				t.Fatal("NewEconomySystem returned nil")
			}
			if es.PriceFluctuation != tt.fluctuation {
				t.Errorf("PriceFluctuation = %v, want %v", es.PriceFluctuation, tt.fluctuation)
			}
			if es.NormalizationRate != tt.normRate {
				t.Errorf("NormalizationRate = %v, want %v", es.NormalizationRate, tt.normRate)
			}
			if es.BasePrices == nil {
				t.Error("BasePrices map should be initialized")
			}
		})
	}
}

func TestEconomySystem_SetBasePrice(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)

	es.SetBasePrice("sword", 100.0)
	es.SetBasePrice("potion", 25.0)

	if es.BasePrices["sword"] != 100.0 {
		t.Errorf("BasePrices[sword] = %v, want 100.0", es.BasePrices["sword"])
	}
	if es.BasePrices["potion"] != 25.0 {
		t.Errorf("BasePrices[potion] = %v, want 25.0", es.BasePrices["potion"])
	}
}

func TestEconomySystem_SetBasePrice_NilMap(t *testing.T) {
	es := &EconomySystem{
		BasePrices: nil,
	}

	// Should not panic, should initialize map
	es.SetBasePrice("sword", 100.0)

	if es.BasePrices == nil {
		t.Error("BasePrices should be initialized")
	}
	if es.BasePrices["sword"] != 100.0 {
		t.Errorf("BasePrices[sword] = %v, want 100.0", es.BasePrices["sword"])
	}
}

func TestEconomySystem_Update(t *testing.T) {
	w := ecs.NewWorld()
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)

	// Create economy node entity
	e := w.CreateEntity()
	w.AddComponent(e, &components.EconomyNode{
		Supply: map[string]int{"sword": 10},
		Demand: map[string]int{"sword": 20},
	})

	// Update should process economy nodes
	es.Update(w, 1.0)

	comp, _ := w.GetComponent(e, "EconomyNode")
	node := comp.(*components.EconomyNode)

	if node.PriceTable == nil {
		t.Error("PriceTable should be initialized after update")
	}
}

func TestEconomySystem_CalculatePriceModifier(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)

	tests := []struct {
		name   string
		supply int
		demand int
		minMod float64
		maxMod float64
	}{
		{"balanced", 10, 10, 0.9, 1.1},
		{"high demand", 10, 20, 1.0, 2.0},
		{"low demand", 20, 10, 0.5, 1.0},
		{"zero supply high demand", 0, 10, 1.5, 2.5},
		{"zero both", 0, 0, 0.9, 1.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := es.calculatePriceModifier(tt.supply, tt.demand)
			if mod < MinPriceMultiplier || mod > MaxPriceMultiplier {
				t.Errorf("Price modifier %v out of valid range [%v, %v]", mod, MinPriceMultiplier, MaxPriceMultiplier)
			}
		})
	}
}

func TestEconomySystem_SellItem(t *testing.T) {
	w := ecs.NewWorld()
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)

	vendor := w.CreateEntity()
	w.AddComponent(vendor, &components.EconomyNode{
		Supply:     map[string]int{"sword": 5},
		Demand:     map[string]int{"sword": 10},
		PriceTable: make(map[string]float64),
	})

	initialSupply := 5

	// Sell 3 swords (supply increases when selling TO vendor)
	revenue := es.SellItem(w, vendor, "sword", 3)

	if revenue <= 0 {
		t.Error("SellItem should return positive revenue")
	}

	comp, _ := w.GetComponent(vendor, "EconomyNode")
	node := comp.(*components.EconomyNode)

	// Supply should have increased by 3
	expectedSupply := initialSupply + 3
	if node.Supply["sword"] != expectedSupply {
		t.Errorf("Supply[sword] = %d, want %d", node.Supply["sword"], expectedSupply)
	}
}

func TestEconomySystem_BuyItem(t *testing.T) {
	w := ecs.NewWorld()
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)

	vendor := w.CreateEntity()
	w.AddComponent(vendor, &components.EconomyNode{
		Supply:     map[string]int{"sword": 10},
		Demand:     map[string]int{"sword": 5},
		PriceTable: make(map[string]float64),
	})

	initialSupply := 10

	// Buy 3 swords (supply decreases when buying FROM vendor)
	cost := es.BuyItem(w, vendor, "sword", 3)

	if cost <= 0 {
		t.Error("BuyItem should return positive cost")
	}

	comp, _ := w.GetComponent(vendor, "EconomyNode")
	node := comp.(*components.EconomyNode)

	// Supply should have decreased by 3
	expectedSupply := initialSupply - 3
	if node.Supply["sword"] != expectedSupply {
		t.Errorf("Supply[sword] = %d, want %d", node.Supply["sword"], expectedSupply)
	}
}

func TestEconomySystem_BuyItem_InsufficientSupply(t *testing.T) {
	w := ecs.NewWorld()
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)

	vendor := w.CreateEntity()
	w.AddComponent(vendor, &components.EconomyNode{
		Supply:     map[string]int{"sword": 2},
		Demand:     map[string]int{"sword": 5},
		PriceTable: make(map[string]float64),
	})

	es.Update(w, 0.1)

	// Try to buy more than available
	es.BuyItem(w, vendor, "sword", 5)

	comp, _ := w.GetComponent(vendor, "EconomyNode")
	node := comp.(*components.EconomyNode)

	// Supply should be 0, not negative
	if node.Supply["sword"] != 0 {
		t.Errorf("Supply[sword] = %d, want 0", node.Supply["sword"])
	}
}

func TestEconomySystem_GetBuyPrice(t *testing.T) {
	w := ecs.NewWorld()
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)

	vendor := w.CreateEntity()
	w.AddComponent(vendor, &components.EconomyNode{
		Supply:     map[string]int{"sword": 10},
		Demand:     map[string]int{"sword": 10},
		PriceTable: map[string]float64{"sword": 120.0},
	})

	price := es.GetBuyPrice(w, vendor, "sword")
	if price != 120.0 {
		t.Errorf("GetBuyPrice = %v, want 120.0", price)
	}
}

func TestEconomySystem_GetBuyPrice_NoEntity(t *testing.T) {
	w := ecs.NewWorld()
	es := NewEconomySystem(0.5, 0.1)

	price := es.GetBuyPrice(w, 999, "sword")
	if price != 0 {
		t.Errorf("GetBuyPrice for non-existent entity = %v, want 0", price)
	}
}

func TestEconomySystem_NormalizeSupplyDemand(t *testing.T) {
	w := ecs.NewWorld()
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)

	vendor := w.CreateEntity()
	w.AddComponent(vendor, &components.EconomyNode{
		Supply:     map[string]int{"sword": 5},
		Demand:     map[string]int{"sword": 10},
		PriceTable: make(map[string]float64),
	})

	// Multiple updates should drift supply toward demand
	for i := 0; i < 10; i++ {
		es.Update(w, 1.0)
	}

	comp, _ := w.GetComponent(vendor, "EconomyNode")
	node := comp.(*components.EconomyNode)

	// Supply should have increased toward demand
	if node.Supply["sword"] <= 5 {
		t.Errorf("Supply[sword] = %d, should have increased toward demand", node.Supply["sword"])
	}
}

func TestClampFloat(t *testing.T) {
	tests := []struct {
		value, min, max, expected float64
	}{
		{5.0, 0.0, 10.0, 5.0},
		{-5.0, 0.0, 10.0, 0.0},
		{15.0, 0.0, 10.0, 10.0},
		{0.0, 0.0, 10.0, 0.0},
		{10.0, 0.0, 10.0, 10.0},
	}

	for _, tt := range tests {
		result := clampFloat(tt.value, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clampFloat(%v, %v, %v) = %v, want %v", tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}

// ============================================================================
// PlayerShop Tests
// ============================================================================

func TestNewPlayerShop(t *testing.T) {
	shop := NewPlayerShop(1, "My Shop", "weapons")

	if shop == nil {
		t.Fatal("NewPlayerShop returned nil")
	}
	if shop.OwnerID != 1 {
		t.Errorf("OwnerID = %d, want 1", shop.OwnerID)
	}
	if shop.ShopName != "My Shop" {
		t.Errorf("ShopName = %s, want 'My Shop'", shop.ShopName)
	}
	if shop.ShopType != "weapons" {
		t.Errorf("ShopType = %s, want 'weapons'", shop.ShopType)
	}
	if shop.Inventory == nil {
		t.Error("Inventory should be initialized")
	}
	if shop.PriceMarkup == nil {
		t.Error("PriceMarkup should be initialized")
	}
	if shop.Employees == nil {
		t.Error("Employees should be initialized")
	}
	if !shop.IsOpen {
		t.Error("Shop should be open by default")
	}
	if shop.OpenHours != [2]int{8, 20} {
		t.Errorf("OpenHours = %v, want [8, 20]", shop.OpenHours)
	}
	if shop.Reputation != 50.0 {
		t.Errorf("Reputation = %v, want 50.0", shop.Reputation)
	}
}

func TestNewPlayerShopSystem(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	pss := NewPlayerShopSystem(es)

	if pss == nil {
		t.Fatal("NewPlayerShopSystem returned nil")
	}
	if pss.Economy != es {
		t.Error("Economy not set correctly")
	}
	if pss.Shops == nil {
		t.Error("Shops map should be initialized")
	}
}

func TestPlayerShopSystem_OpenShop(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	pss := NewPlayerShopSystem(es)

	loc := [3]float64{100, 0, 200}
	shop := pss.OpenShop(1, "Test Shop", "general", loc)

	if shop == nil {
		t.Fatal("OpenShop returned nil")
	}
	if shop.Location != loc {
		t.Errorf("Location = %v, want %v", shop.Location, loc)
	}
	if pss.Shops[1] != shop {
		t.Error("Shop not registered in system")
	}
}

func TestPlayerShopSystem_StockItem(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "general", [3]float64{0, 0, 0})

	result := pss.StockItem(1, "sword", 10, 1.3)
	if !result {
		t.Error("StockItem should return true")
	}

	shop := pss.Shops[1]
	if shop.Inventory["sword"] != 10 {
		t.Errorf("Inventory[sword] = %d, want 10", shop.Inventory["sword"])
	}
	if shop.PriceMarkup["sword"] != 1.3 {
		t.Errorf("PriceMarkup[sword] = %v, want 1.3", shop.PriceMarkup["sword"])
	}
}

func TestPlayerShopSystem_StockItem_NoShop(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	pss := NewPlayerShopSystem(es)

	result := pss.StockItem(999, "sword", 10, 1.3)
	if result {
		t.Error("StockItem should return false for non-existent shop")
	}
}

func TestPlayerShopSystem_GetShopPrice(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})
	pss.StockItem(1, "sword", 10, 1.5) // 50% markup

	price := pss.GetShopPrice(1, "sword")
	if price != 150.0 {
		t.Errorf("GetShopPrice = %v, want 150.0", price)
	}
}

func TestPlayerShopSystem_GetShopPrice_DefaultMarkup(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})
	pss.StockItem(1, "sword", 10, 0) // No markup specified

	price := pss.GetShopPrice(1, "sword")
	// Default markup is 1.2 (20%)
	if price != 120.0 {
		t.Errorf("GetShopPrice with default markup = %v, want 120.0", price)
	}
}

func TestPlayerShopSystem_ProcessSale(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})
	pss.StockItem(1, "sword", 10, 1.2)

	revenue := pss.ProcessSale(1, "sword", 2)
	if revenue != 240.0 { // 2 * 100 * 1.2
		t.Errorf("ProcessSale revenue = %v, want 240.0", revenue)
	}

	shop := pss.Shops[1]
	if shop.Inventory["sword"] != 8 {
		t.Errorf("Inventory[sword] = %d, want 8", shop.Inventory["sword"])
	}
	if shop.DailyProfit != 240.0 {
		t.Errorf("DailyProfit = %v, want 240.0", shop.DailyProfit)
	}
}

func TestPlayerShopSystem_ProcessSale_InsufficientInventory(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})
	pss.StockItem(1, "sword", 2, 1.2)

	revenue := pss.ProcessSale(1, "sword", 5) // Try to sell more than available
	if revenue != 0 {
		t.Errorf("ProcessSale with insufficient inventory = %v, want 0", revenue)
	}
}

func TestPlayerShopSystem_ProcessSale_ClosedShop(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})
	pss.StockItem(1, "sword", 10, 1.2)
	pss.Shops[1].IsOpen = false

	revenue := pss.ProcessSale(1, "sword", 2)
	if revenue != 0 {
		t.Errorf("ProcessSale on closed shop = %v, want 0", revenue)
	}
}

func TestPlayerShopSystem_HireEmployee(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})

	result := pss.HireEmployee(1, 100)
	if !result {
		t.Error("HireEmployee should return true")
	}

	shop := pss.Shops[1]
	if len(shop.Employees) != 1 {
		t.Errorf("Employees count = %d, want 1", len(shop.Employees))
	}
	if shop.Employees[0] != 100 {
		t.Errorf("Employee ID = %d, want 100", shop.Employees[0])
	}
}

func TestPlayerShopSystem_Update(t *testing.T) {
	w := ecs.NewWorld()
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)
	pss := NewPlayerShopSystem(es)
	pss.CurrentHour = 12 // Midday

	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})
	pss.StockItem(1, "sword", 100, 1.2)
	pss.Shops[1].Reputation = 80.0 // High reputation

	// Should not panic
	pss.Update(w, 1.0)

	// Shop should remain open during business hours
	if !pss.Shops[1].IsOpen {
		t.Error("Shop should be open during business hours")
	}
}

func TestPlayerShopSystem_Update_ClosedHours(t *testing.T) {
	w := ecs.NewWorld()
	es := NewEconomySystem(0.5, 0.1)
	pss := NewPlayerShopSystem(es)
	pss.CurrentHour = 3 // 3 AM

	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})

	pss.Update(w, 1.0)

	if pss.Shops[1].IsOpen {
		t.Error("Shop should be closed at 3 AM")
	}
}

func TestPlayerShopSystem_GetDailyProfit(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})
	pss.StockItem(1, "sword", 10, 1.2)

	pss.ProcessSale(1, "sword", 2)

	profit := pss.GetDailyProfit(1)
	if profit != 240.0 {
		t.Errorf("GetDailyProfit = %v, want 240.0", profit)
	}
}

func TestPlayerShopSystem_ResetDailyProfit(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})
	pss.StockItem(1, "sword", 10, 1.2)

	pss.ProcessSale(1, "sword", 2)
	pss.ResetDailyProfit(1)

	if pss.Shops[1].DailyProfit != 0 {
		t.Errorf("DailyProfit = %v, want 0 after reset", pss.Shops[1].DailyProfit)
	}

	// Should have sales history
	history := pss.SalesHistory[1]
	if len(history) != 1 {
		t.Errorf("SalesHistory length = %d, want 1", len(history))
	}
	if history[0] != 240.0 {
		t.Errorf("SalesHistory[0] = %v, want 240.0", history[0])
	}
}

func TestPlayerShopSystem_SalesHistoryLimit(t *testing.T) {
	es := NewEconomySystem(0.5, 0.1)
	es.SetBasePrice("sword", 100.0)
	pss := NewPlayerShopSystem(es)
	pss.OpenShop(1, "Test Shop", "weapons", [3]float64{0, 0, 0})

	// Add more than 30 days of history
	for i := 0; i < 35; i++ {
		pss.Shops[1].DailyProfit = float64(i * 10)
		pss.ResetDailyProfit(1)
	}

	// History should be limited to 30 entries
	if len(pss.SalesHistory[1]) != 30 {
		t.Errorf("SalesHistory length = %d, want 30 (limited)", len(pss.SalesHistory[1]))
	}
}
