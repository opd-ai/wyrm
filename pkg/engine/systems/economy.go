package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

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
	s.ensureBasePricesInitialized()
	for _, e := range w.Entities("EconomyNode") {
		s.processEconomyEntity(w, e)
	}
}

// ensureBasePricesInitialized initializes the BasePrices map if nil.
func (s *EconomySystem) ensureBasePricesInitialized() {
	if s.BasePrices == nil {
		s.BasePrices = make(map[string]float64)
	}
}

// processEconomyEntity updates a single economy node entity.
func (s *EconomySystem) processEconomyEntity(w *ecs.World, e ecs.Entity) {
	comp, ok := w.GetComponent(e, "EconomyNode")
	if !ok {
		return
	}
	node := comp.(*components.EconomyNode)
	s.initializeNodeMaps(node)
	s.updateNodePrices(node)
	s.normalizeSupplyDemand(node)
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
	ratio := BasePriceMultiplier
	if supply > 0 {
		ratio = float64(demand) / float64(supply)
	} else if demand > 0 {
		ratio = HighDemandPriceMultiplier // High demand, no supply = double price
	}
	priceMod := BasePriceMultiplier + (ratio-BasePriceMultiplier)*s.PriceFluctuation
	return clampFloat(priceMod, MinPriceMultiplier, MaxPriceMultiplier)
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

// PlayerShop represents a player-owned shop.
type PlayerShop struct {
	OwnerID     ecs.Entity
	ShopName    string
	ShopType    string // "general", "weapons", "armor", "potions", "magic"
	Location    [3]float64
	Inventory   map[string]int     // item -> quantity
	PriceMarkup map[string]float64 // item -> markup multiplier (1.0 = base)
	DailyProfit float64
	Employees   []ecs.Entity
	IsOpen      bool
	OpenHours   [2]int // [open hour, close hour]
	Reputation  float64
}

// NewPlayerShop creates a new player-owned shop.
func NewPlayerShop(owner ecs.Entity, name, shopType string) *PlayerShop {
	return &PlayerShop{
		OwnerID:     owner,
		ShopName:    name,
		ShopType:    shopType,
		Inventory:   make(map[string]int),
		PriceMarkup: make(map[string]float64),
		Employees:   make([]ecs.Entity, 0),
		IsOpen:      true,
		OpenHours:   [2]int{8, 20}, // 8 AM to 8 PM
		Reputation:  50.0,          // Start with neutral reputation
	}
}

// PlayerShopSystem manages player-owned shops.
type PlayerShopSystem struct {
	Shops        map[ecs.Entity]*PlayerShop
	Economy      *EconomySystem
	CurrentHour  int
	SalesHistory map[ecs.Entity][]float64 // owner -> daily sales
}

// NewPlayerShopSystem creates a new player shop system.
func NewPlayerShopSystem(economy *EconomySystem) *PlayerShopSystem {
	return &PlayerShopSystem{
		Shops:        make(map[ecs.Entity]*PlayerShop),
		Economy:      economy,
		SalesHistory: make(map[ecs.Entity][]float64),
	}
}

// OpenShop opens a new player shop.
func (s *PlayerShopSystem) OpenShop(owner ecs.Entity, name, shopType string, loc [3]float64) *PlayerShop {
	shop := NewPlayerShop(owner, name, shopType)
	shop.Location = loc
	s.Shops[owner] = shop
	return shop
}

// StockItem adds items to a player's shop inventory.
func (s *PlayerShopSystem) StockItem(owner ecs.Entity, item string, qty int, markup float64) bool {
	shop, ok := s.Shops[owner]
	if !ok {
		return false
	}
	shop.Inventory[item] += qty
	if markup > 0 {
		shop.PriceMarkup[item] = markup
	}
	return true
}

// GetShopPrice gets the price for an item at a player shop.
func (s *PlayerShopSystem) GetShopPrice(owner ecs.Entity, item string) float64 {
	shop, ok := s.Shops[owner]
	if !ok {
		return 0
	}
	basePrice := s.Economy.BasePrices[item]
	markup := shop.PriceMarkup[item]
	if markup <= 0 {
		markup = 1.2 // Default 20% markup
	}
	return basePrice * markup
}

// ProcessSale handles a customer buying from a player shop.
func (s *PlayerShopSystem) ProcessSale(owner ecs.Entity, item string, qty int) float64 {
	shop, ok := s.Shops[owner]
	if !ok || !shop.IsOpen {
		return 0
	}
	if shop.Inventory[item] < qty {
		return 0
	}
	shop.Inventory[item] -= qty
	price := s.GetShopPrice(owner, item) * float64(qty)
	shop.DailyProfit += price
	// Successful sales improve reputation
	shop.Reputation = clampFloat(shop.Reputation+0.1, 0, 100)
	return price
}

// HireEmployee adds an NPC employee to a shop.
func (s *PlayerShopSystem) HireEmployee(owner, employee ecs.Entity) bool {
	shop, ok := s.Shops[owner]
	if !ok {
		return false
	}
	shop.Employees = append(shop.Employees, employee)
	return true
}

// Update processes player shops each tick.
func (s *PlayerShopSystem) Update(w *ecs.World, dt float64) {
	for owner, shop := range s.Shops {
		// Check if shop should be open based on hours
		shop.IsOpen = s.CurrentHour >= shop.OpenHours[0] && s.CurrentHour < shop.OpenHours[1]
		// Process NPC customers (simplified)
		if shop.IsOpen && shop.Reputation > 30 {
			s.simulateCustomers(owner, shop)
		}
	}
}

// simulateCustomers simulates NPC customers visiting shops.
func (s *PlayerShopSystem) simulateCustomers(owner ecs.Entity, shop *PlayerShop) {
	// Higher reputation = more customers
	customerChance := shop.Reputation / 200.0 // 50% at max rep
	if len(shop.Inventory) == 0 {
		return
	}
	// Simulate a potential customer
	if customerChance > 0.1 {
		for item, qty := range shop.Inventory {
			if qty > 0 {
				s.ProcessSale(owner, item, 1)
				break
			}
		}
	}
}

// GetDailyProfit returns the daily profit for a shop.
func (s *PlayerShopSystem) GetDailyProfit(owner ecs.Entity) float64 {
	shop, ok := s.Shops[owner]
	if !ok {
		return 0
	}
	return shop.DailyProfit
}

// ResetDailyProfit resets daily profit tracking.
func (s *PlayerShopSystem) ResetDailyProfit(owner ecs.Entity) {
	shop, ok := s.Shops[owner]
	if !ok {
		return
	}
	// Add to history
	if s.SalesHistory[owner] == nil {
		s.SalesHistory[owner] = make([]float64, 0)
	}
	s.SalesHistory[owner] = append(s.SalesHistory[owner], shop.DailyProfit)
	if len(s.SalesHistory[owner]) > 30 {
		s.SalesHistory[owner] = s.SalesHistory[owner][1:]
	}
	shop.DailyProfit = 0
}
