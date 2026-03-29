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
