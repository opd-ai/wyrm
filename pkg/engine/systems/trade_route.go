package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// TradeRouteStatus represents the current status of a trade route.
type TradeRouteStatus int

const (
	TradeRouteActive TradeRouteStatus = iota
	TradeRouteSuspended
	TradeRouteBlocked
	TradeRouteDisrupted
)

// RouteHazardType represents types of hazards affecting trade routes.
type RouteHazardType int

const (
	HazardNone RouteHazardType = iota
	HazardBandits
	HazardWeather
	HazardWar
	HazardPlague
	HazardMonsters
	HazardTax
	HazardPirates
)

// TradeRoute represents a trade route between two locations.
type TradeRoute struct {
	ID              string
	OriginNode      ecs.Entity
	DestinationNode ecs.Entity
	OriginName      string
	DestinationName string
	Distance        float64
	TravelTime      float64 // Hours
	Status          TradeRouteStatus
	Hazard          RouteHazardType
	HazardSeverity  float64 // 0-1
	TradedGoods     map[string]int
	TollCost        float64
	ProfitMargin    float64 // Expected profit multiplier
	LastUsed        float64 // Game time when last used
	UsageCount      int
	Reputation      float64 // Route reputation affects usage
}

// TradeCaravan represents a caravan traveling along a route.
type TradeCaravan struct {
	ID           string
	OwnerID      ecs.Entity
	RouteID      string
	Cargo        map[string]int
	CargoCost    float64 // Initial cost of cargo
	Guards       int
	Progress     float64 // 0-1 progress along route
	TravelSpeed  float64 // Modified by cargo, guards, weather
	IsReturning  bool
	Status       CaravanStatus
	HazardDamage float64 // Cargo lost to hazards
	StartTime    float64
	ExpectedTime float64
	ActualProfit float64
}

// CaravanStatus represents the current state of a caravan.
type CaravanStatus int

const (
	CaravanPreparing CaravanStatus = iota
	CaravanTraveling
	CaravanTrading
	CaravanReturning
	CaravanArrived
	CaravanLost
	CaravanRobbed
)

// TradeOffer represents goods available for trade.
type TradeOffer struct {
	ItemType      string
	Quantity      int
	BuyPrice      float64
	SellPrice     float64
	ProfitPerUnit float64
	Demand        float64 // Demand at destination
	Supply        float64 // Supply at origin
}

// TradeRouteSystem manages trade routes and caravans.
type TradeRouteSystem struct {
	Genre           string
	Routes          map[string]*TradeRoute
	Caravans        map[string]*TradeCaravan
	Economy         *EconomySystem
	GameTime        float64
	rng             *PseudoRandom
	idCounter       uint64                  // Counter for generating unique IDs
	RouteDiscovery  map[ecs.Entity][]string // Player discovered routes
	TradeHistory    map[ecs.Entity][]TradeRecord
	CaravanCapacity int
	BaseSpeed       float64
}

// TradeRecord tracks historical trade data.
type TradeRecord struct {
	RouteID    string
	Profit     float64
	CargoValue float64
	GameTime   float64
	Success    bool
}

// NewTradeRouteSystem creates a new trade route system.
func NewTradeRouteSystem(seed int64, genre string, economy *EconomySystem) *TradeRouteSystem {
	return &TradeRouteSystem{
		Genre:           genre,
		Routes:          make(map[string]*TradeRoute),
		Caravans:        make(map[string]*TradeCaravan),
		Economy:         economy,
		rng:             NewPseudoRandom(seed),
		RouteDiscovery:  make(map[ecs.Entity][]string),
		TradeHistory:    make(map[ecs.Entity][]TradeRecord),
		CaravanCapacity: 100,
		BaseSpeed:       1.0,
	}
}

// pseudoRandom generates a deterministic pseudo-random number.
func (s *TradeRouteSystem) pseudoRandom() float64 {
	return s.rng.Float64()
}

// pseudoRandomInt generates a deterministic pseudo-random integer.
func (s *TradeRouteSystem) pseudoRandomInt(max int) int {
	return s.rng.Int(max)
}

// CreateRoute establishes a new trade route between two economy nodes.
func (s *TradeRouteSystem) CreateRoute(origin, dest ecs.Entity, originName, destName string, dist float64) *TradeRoute {
	routeID := s.generateRouteID(originName, destName)
	route := &TradeRoute{
		ID:              routeID,
		OriginNode:      origin,
		DestinationNode: dest,
		OriginName:      originName,
		DestinationName: destName,
		Distance:        dist,
		TravelTime:      dist / 10.0, // Base travel time
		Status:          TradeRouteActive,
		Hazard:          HazardNone,
		HazardSeverity:  0,
		TradedGoods:     make(map[string]int),
		TollCost:        dist * 0.1,
		ProfitMargin:    1.0 + (dist * 0.01), // Longer = more profit potential
		Reputation:      50.0,
	}
	s.Routes[routeID] = route
	return route
}

// generateRouteID creates a unique route identifier.
func (s *TradeRouteSystem) generateRouteID(origin, dest string) string {
	return origin + "_to_" + dest
}

// DiscoverRoute marks a route as discovered by a player.
func (s *TradeRouteSystem) DiscoverRoute(player ecs.Entity, routeID string) bool {
	if _, exists := s.Routes[routeID]; !exists {
		return false
	}
	if s.RouteDiscovery[player] == nil {
		s.RouteDiscovery[player] = make([]string, 0)
	}
	for _, r := range s.RouteDiscovery[player] {
		if r == routeID {
			return false // Already discovered
		}
	}
	s.RouteDiscovery[player] = append(s.RouteDiscovery[player], routeID)
	return true
}

// GetDiscoveredRoutes returns routes discovered by a player.
func (s *TradeRouteSystem) GetDiscoveredRoutes(player ecs.Entity) []*TradeRoute {
	routes := make([]*TradeRoute, 0)
	for _, routeID := range s.RouteDiscovery[player] {
		if route, exists := s.Routes[routeID]; exists {
			routes = append(routes, route)
		}
	}
	return routes
}

// GetTradeOffers calculates profitable trades for a route.
func (s *TradeRouteSystem) GetTradeOffers(w *ecs.World, routeID string) []TradeOffer {
	route, exists := s.Routes[routeID]
	if !exists {
		return nil
	}

	origin, dest, ok := s.getRouteNodes(w, route)
	if !ok {
		return nil
	}

	return s.calculateProfitableOffers(origin, dest, route)
}

// economyNodeInterface defines the methods needed from economy nodes.
type economyNodeInterface interface {
	GetSupply(string) int
	GetDemand(string) int
	GetPrice(string) float64
}

// getRouteNodes retrieves the origin and destination economy nodes for a route.
func (s *TradeRouteSystem) getRouteNodes(w *ecs.World, route *TradeRoute) (origin, dest economyNodeInterface, ok bool) {
	originComp, ok := w.GetComponent(route.OriginNode, "EconomyNode")
	if !ok {
		return nil, nil, false
	}
	destComp, ok := w.GetComponent(route.DestinationNode, "EconomyNode")
	if !ok {
		return nil, nil, false
	}
	return originComp.(economyNodeInterface), destComp.(economyNodeInterface), true
}

// calculateProfitableOffers generates trade offers for profitable items.
func (s *TradeRouteSystem) calculateProfitableOffers(origin, dest economyNodeInterface, route *TradeRoute) []TradeOffer {
	offers := make([]TradeOffer, 0)
	for itemType, basePrice := range s.Economy.BasePrices {
		offer, profitable := s.evaluateTradeItem(itemType, basePrice, origin, dest, route)
		if profitable {
			offers = append(offers, offer)
		}
	}
	return offers
}

// evaluateTradeItem determines if an item is profitable to trade.
func (s *TradeRouteSystem) evaluateTradeItem(itemType string, basePrice float64, origin, dest economyNodeInterface, route *TradeRoute) (TradeOffer, bool) {
	buyPrice := origin.GetPrice(itemType)
	if buyPrice <= 0 {
		buyPrice = basePrice
	}
	sellPrice := dest.GetPrice(itemType)
	if sellPrice <= 0 {
		sellPrice = basePrice
	}

	profit := sellPrice - buyPrice - route.TollCost
	if profit <= 0 {
		return TradeOffer{}, false
	}

	return TradeOffer{
		ItemType:      itemType,
		Quantity:      origin.GetSupply(itemType),
		BuyPrice:      buyPrice,
		SellPrice:     sellPrice,
		ProfitPerUnit: profit,
		Demand:        float64(dest.GetDemand(itemType)),
		Supply:        float64(origin.GetSupply(itemType)),
	}, true
}

// LaunchCaravan starts a trading caravan on a route.
func (s *TradeRouteSystem) LaunchCaravan(owner ecs.Entity, routeID string, cargo map[string]int, guards int) *TradeCaravan {
	route, exists := s.Routes[routeID]
	if !exists || route.Status != TradeRouteActive {
		return nil
	}
	caravanID := s.generateCaravanID(owner)
	cargoCost := s.calculateCargoCost(cargo)
	caravan := &TradeCaravan{
		ID:           caravanID,
		OwnerID:      owner,
		RouteID:      routeID,
		Cargo:        cargo,
		CargoCost:    cargoCost,
		Guards:       guards,
		Progress:     0,
		TravelSpeed:  s.calculateTravelSpeed(cargo, guards),
		IsReturning:  false,
		Status:       CaravanTraveling,
		StartTime:    s.GameTime,
		ExpectedTime: route.TravelTime / s.calculateTravelSpeed(cargo, guards),
	}
	s.Caravans[caravanID] = caravan
	return caravan
}

// generateCaravanID creates a unique caravan identifier.
func (s *TradeRouteSystem) generateCaravanID(owner ecs.Entity) string {
	s.idCounter++
	return "caravan_" + string(rune('0'+owner%10)) + "_" + string(rune('0'+s.idCounter%1000))
}

// calculateCargoCost computes the total cost of cargo.
func (s *TradeRouteSystem) calculateCargoCost(cargo map[string]int) float64 {
	total := 0.0
	for item, qty := range cargo {
		if price, ok := s.Economy.BasePrices[item]; ok {
			total += price * float64(qty)
		}
	}
	return total
}

// calculateTravelSpeed computes travel speed based on load and guards.
func (s *TradeRouteSystem) calculateTravelSpeed(cargo map[string]int, guards int) float64 {
	totalCargo := 0
	for _, qty := range cargo {
		totalCargo += qty
	}
	// Heavy loads slow down
	loadPenalty := 1.0 - (float64(totalCargo) / float64(s.CaravanCapacity*2))
	if loadPenalty < 0.5 {
		loadPenalty = 0.5
	}
	// More guards = slightly slower but safer
	guardPenalty := 1.0 - (float64(guards) * 0.02)
	if guardPenalty < 0.8 {
		guardPenalty = 0.8
	}
	return s.BaseSpeed * loadPenalty * guardPenalty
}

// Update processes all active caravans and route conditions.
func (s *TradeRouteSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	// Update route hazards
	s.updateRouteHazards()
	// Process caravans
	for caravanID, caravan := range s.Caravans {
		s.updateCaravan(w, caravanID, caravan, dt)
	}
}

// updateRouteHazards randomly generates and clears hazards on routes.
func (s *TradeRouteSystem) updateRouteHazards() {
	for _, route := range s.Routes {
		// Small chance of new hazard
		if route.Hazard == HazardNone && s.pseudoRandom() < 0.001 {
			route.Hazard = s.generateRandomHazard()
			route.HazardSeverity = 0.2 + s.pseudoRandom()*0.6
			route.Status = TradeRouteDisrupted
		}
		// Hazards clear over time
		if route.Hazard != HazardNone {
			route.HazardSeverity -= 0.0001
			if route.HazardSeverity <= 0 {
				route.Hazard = HazardNone
				route.HazardSeverity = 0
				if route.Status == TradeRouteDisrupted {
					route.Status = TradeRouteActive
				}
			}
		}
	}
}

// generateRandomHazard creates a genre-appropriate hazard.
func (s *TradeRouteSystem) generateRandomHazard() RouteHazardType {
	hazards := s.getGenreHazards()
	if len(hazards) == 0 {
		return HazardBandits
	}
	return hazards[s.pseudoRandomInt(len(hazards))]
}

// getGenreHazards returns hazard types appropriate for the genre.
func (s *TradeRouteSystem) getGenreHazards() []RouteHazardType {
	switch s.Genre {
	case "fantasy":
		return []RouteHazardType{HazardBandits, HazardMonsters, HazardWeather}
	case "sci-fi":
		return []RouteHazardType{HazardPirates, HazardTax, HazardWeather}
	case "horror":
		return []RouteHazardType{HazardMonsters, HazardPlague, HazardWeather}
	case "cyberpunk":
		return []RouteHazardType{HazardBandits, HazardTax, HazardWar}
	case "post-apocalyptic":
		return []RouteHazardType{HazardBandits, HazardPlague, HazardMonsters}
	default:
		return []RouteHazardType{HazardBandits, HazardWeather}
	}
}

// updateCaravan processes a single caravan's progress.
func (s *TradeRouteSystem) updateCaravan(w *ecs.World, caravanID string, caravan *TradeCaravan, dt float64) {
	if caravan.Status == CaravanArrived || caravan.Status == CaravanLost || caravan.Status == CaravanRobbed {
		return
	}
	route, exists := s.Routes[caravan.RouteID]
	if !exists {
		caravan.Status = CaravanLost
		return
	}
	// Update progress
	progressDelta := (dt / 3600.0) * caravan.TravelSpeed / route.TravelTime
	caravan.Progress += progressDelta
	// Check for hazard encounters
	if route.Hazard != HazardNone && s.pseudoRandom() < route.HazardSeverity*0.01 {
		s.processHazardEncounter(caravan, route)
	}
	// Check if arrived
	if caravan.Progress >= 1.0 {
		s.processCaravanArrival(w, caravan, route)
	}
}

// processHazardEncounter handles a caravan encountering a hazard.
func (s *TradeRouteSystem) processHazardEncounter(caravan *TradeCaravan, route *TradeRoute) {
	// Guards reduce hazard impact
	guardProtection := float64(caravan.Guards) * 0.1
	if guardProtection > 0.8 {
		guardProtection = 0.8
	}
	damage := route.HazardSeverity * (1.0 - guardProtection)
	caravan.HazardDamage += damage
	// Severe damage can destroy caravan
	if caravan.HazardDamage > 0.8 {
		if s.pseudoRandom() < 0.5 {
			caravan.Status = CaravanLost
		} else {
			caravan.Status = CaravanRobbed
			// Lose some cargo
			for item := range caravan.Cargo {
				loss := int(float64(caravan.Cargo[item]) * damage)
				caravan.Cargo[item] -= loss
				if caravan.Cargo[item] < 0 {
					caravan.Cargo[item] = 0
				}
			}
		}
	}
}

// processCaravanArrival handles a caravan reaching its destination.
func (s *TradeRouteSystem) processCaravanArrival(w *ecs.World, caravan *TradeCaravan, route *TradeRoute) {
	if caravan.IsReturning {
		caravan.Status = CaravanArrived
		s.recordTrade(caravan, route, true)
		return
	}
	// Sell cargo at destination
	profit := s.sellCargoAtDestination(w, caravan, route)
	caravan.ActualProfit = profit - caravan.CargoCost - route.TollCost
	// Start return journey
	caravan.IsReturning = true
	caravan.Progress = 0
	caravan.Status = CaravanReturning
	// Update route usage
	route.UsageCount++
	route.LastUsed = s.GameTime
	if caravan.ActualProfit > 0 {
		route.Reputation = clampFloat(route.Reputation+0.5, 0, 100)
	}
}

// sellCargoAtDestination sells all caravan cargo at destination prices.
func (s *TradeRouteSystem) sellCargoAtDestination(w *ecs.World, caravan *TradeCaravan, route *TradeRoute) float64 {
	dest := s.getDestinationPricer(w, route)
	return s.calculateCargoValue(caravan, route, dest)
}

// getDestinationPricer returns a price getter interface for the destination, or nil.
func (s *TradeRouteSystem) getDestinationPricer(w *ecs.World, route *TradeRoute) interface{ GetPrice(string) float64 } {
	destComp, ok := w.GetComponent(route.DestinationNode, "EconomyNode")
	if !ok {
		return nil
	}
	dest, ok := destComp.(interface{ GetPrice(string) float64 })
	if !ok {
		return nil
	}
	return dest
}

// calculateCargoValue computes the total value of caravan cargo.
func (s *TradeRouteSystem) calculateCargoValue(caravan *TradeCaravan, route *TradeRoute, dest interface{ GetPrice(string) float64 }) float64 {
	total := 0.0
	for item, qty := range caravan.Cargo {
		price := s.getItemPrice(item, route, dest)
		total += price * float64(qty) * s.getDamageMultiplier(caravan, dest != nil)
	}
	return total
}

// getItemPrice returns the sale price for an item.
func (s *TradeRouteSystem) getItemPrice(item string, route *TradeRoute, dest interface{ GetPrice(string) float64 }) float64 {
	if dest != nil {
		if price := dest.GetPrice(item); price > 0 {
			return price
		}
	}
	if basePrice, ok := s.Economy.BasePrices[item]; ok {
		return basePrice * route.ProfitMargin
	}
	return 0
}

// getDamageMultiplier returns the cargo damage multiplier.
func (s *TradeRouteSystem) getDamageMultiplier(caravan *TradeCaravan, hasDestPricing bool) float64 {
	if hasDestPricing {
		return 1.0 - caravan.HazardDamage*0.5
	}
	return 1.0
}

// recordTrade records trade history for a player.
func (s *TradeRouteSystem) recordTrade(caravan *TradeCaravan, route *TradeRoute, success bool) {
	record := TradeRecord{
		RouteID:    route.ID,
		Profit:     caravan.ActualProfit,
		CargoValue: caravan.CargoCost,
		GameTime:   s.GameTime,
		Success:    success && caravan.ActualProfit > 0,
	}
	if s.TradeHistory[caravan.OwnerID] == nil {
		s.TradeHistory[caravan.OwnerID] = make([]TradeRecord, 0)
	}
	s.TradeHistory[caravan.OwnerID] = append(s.TradeHistory[caravan.OwnerID], record)
	// Keep only last 50 records
	if len(s.TradeHistory[caravan.OwnerID]) > 50 {
		s.TradeHistory[caravan.OwnerID] = s.TradeHistory[caravan.OwnerID][1:]
	}
}

// GetCaravanStatus returns the status of a player's caravans.
func (s *TradeRouteSystem) GetCaravanStatus(owner ecs.Entity) []*TradeCaravan {
	caravans := make([]*TradeCaravan, 0)
	for _, caravan := range s.Caravans {
		if caravan.OwnerID == owner {
			caravans = append(caravans, caravan)
		}
	}
	return caravans
}

// GetTradeHistory returns trading history for a player.
func (s *TradeRouteSystem) GetTradeHistory(owner ecs.Entity) []TradeRecord {
	return s.TradeHistory[owner]
}

// GetRouteProfitability calculates expected profit for a route.
func (s *TradeRouteSystem) GetRouteProfitability(routeID string) float64 {
	route, exists := s.Routes[routeID]
	if !exists {
		return 0
	}
	// Base profitability from margin
	base := route.ProfitMargin
	// Hazards reduce profitability
	if route.Hazard != HazardNone {
		base *= (1.0 - route.HazardSeverity)
	}
	// High usage means more competition
	if route.UsageCount > 100 {
		base *= 0.9
	}
	// Route reputation affects expected outcome
	base *= route.Reputation / 50.0
	return base
}

// SuspendRoute temporarily disables a trade route.
func (s *TradeRouteSystem) SuspendRoute(routeID string) bool {
	route, exists := s.Routes[routeID]
	if !exists {
		return false
	}
	route.Status = TradeRouteSuspended
	return true
}

// ResumeRoute reactivates a suspended trade route.
func (s *TradeRouteSystem) ResumeRoute(routeID string) bool {
	route, exists := s.Routes[routeID]
	if !exists {
		return false
	}
	if route.Status == TradeRouteSuspended {
		route.Status = TradeRouteActive
		return true
	}
	return false
}

// GetHazardDescription returns a genre-appropriate hazard description.
func (s *TradeRouteSystem) GetHazardDescription(hazard RouteHazardType) string {
	descriptions := s.getHazardDescriptions()
	if desc, ok := descriptions[hazard]; ok {
		return desc
	}
	return "Unknown hazard"
}

// getHazardDescriptions returns genre-appropriate hazard descriptions.
func (s *TradeRouteSystem) getHazardDescriptions() map[RouteHazardType]string {
	switch s.Genre {
	case "fantasy":
		return map[RouteHazardType]string{
			HazardBandits:  "Brigands lurk along the road, preying on merchants.",
			HazardWeather:  "Fierce storms make the path treacherous.",
			HazardMonsters: "Fell beasts have been sighted near the route.",
			HazardWar:      "Armed conflict has spilled onto the trade roads.",
			HazardPlague:   "A mysterious plague has closed the route.",
			HazardTax:      "Heavy tolls are being demanded by local lords.",
			HazardPirates:  "River pirates control key crossing points.",
		}
	case "sci-fi":
		return map[RouteHazardType]string{
			HazardBandits:  "Space pirates patrol the shipping lanes.",
			HazardWeather:  "Solar storms disrupt navigation systems.",
			HazardMonsters: "Hostile alien lifeforms threaten the route.",
			HazardWar:      "Military operations have closed the sector.",
			HazardPlague:   "Quarantine protocols are in effect.",
			HazardTax:      "Excessive import tariffs at the checkpoint.",
			HazardPirates:  "Raider fleets control the hyperspace routes.",
		}
	case "horror":
		return map[RouteHazardType]string{
			HazardBandits:  "Cultists waylay travelers for their rituals.",
			HazardWeather:  "Unnatural fog blankets the path, hiding horrors.",
			HazardMonsters: "Nameless things hunt in the darkness.",
			HazardWar:      "The war between light and shadow rages here.",
			HazardPlague:   "A supernatural blight corrupts all who pass.",
			HazardTax:      "The toll collector demands more than gold.",
			HazardPirates:  "Ghost ships patrol these cursed waters.",
		}
	case "cyberpunk":
		return map[RouteHazardType]string{
			HazardBandits:  "Chrome-augmented gangs raid the transport lines.",
			HazardWeather:  "Acid rain has damaged critical infrastructure.",
			HazardMonsters: "Rogue AIs control autonomous defense systems.",
			HazardWar:      "Corporate warfare has made the area a warzone.",
			HazardPlague:   "A nano-virus outbreak has closed the district.",
			HazardTax:      "Megacorp tollgates extract excessive fees.",
			HazardPirates:  "Data pirates intercept all transmissions.",
		}
	case "post-apocalyptic":
		return map[RouteHazardType]string{
			HazardBandits:  "Raider gangs control this stretch of wasteland.",
			HazardWeather:  "Rad storms make travel extremely dangerous.",
			HazardMonsters: "Mutated creatures roam the area freely.",
			HazardWar:      "Faction warfare has turned this into a killzone.",
			HazardPlague:   "Contamination levels exceed safe limits.",
			HazardTax:      "Warlords demand tribute for safe passage.",
			HazardPirates:  "Water pirates control the river crossings.",
		}
	default:
		return map[RouteHazardType]string{
			HazardBandits:  "Bandits threaten the trade route.",
			HazardWeather:  "Severe weather has disrupted travel.",
			HazardMonsters: "Dangerous creatures block the path.",
			HazardWar:      "Armed conflict has closed the route.",
			HazardPlague:   "Disease outbreak has halted trade.",
			HazardTax:      "High taxes are being charged.",
			HazardPirates:  "Pirates control key waterways.",
		}
	}
}

// CalculateCargoCapacity returns the remaining capacity for a caravan.
func (s *TradeRouteSystem) CalculateCargoCapacity(cargo map[string]int) int {
	total := 0
	for _, qty := range cargo {
		total += qty
	}
	return s.CaravanCapacity - total
}

// EstimateProfit estimates profit for a potential trade.
func (s *TradeRouteSystem) EstimateProfit(routeID string, cargo map[string]int, guards int) float64 {
	route, exists := s.Routes[routeID]
	if !exists {
		return 0
	}
	// Calculate cargo cost
	cargoCost := s.calculateCargoCost(cargo)
	// Estimate sell value with margin
	sellValue := cargoCost * route.ProfitMargin
	// Account for hazard risk
	hazardRisk := 0.0
	if route.Hazard != HazardNone {
		guardProtection := float64(guards) * 0.1
		if guardProtection > 0.8 {
			guardProtection = 0.8
		}
		hazardRisk = route.HazardSeverity * (1.0 - guardProtection) * 0.3
	}
	// Guard costs
	guardCost := float64(guards) * 50.0
	// Final estimate
	estimate := sellValue*(1.0-hazardRisk) - cargoCost - route.TollCost - guardCost
	return estimate
}

// GetRouteStatus returns the current status of a route.
func (s *TradeRouteSystem) GetRouteStatus(routeID string) (TradeRouteStatus, RouteHazardType, float64) {
	route, exists := s.Routes[routeID]
	if !exists {
		return TradeRouteBlocked, HazardNone, 0
	}
	return route.Status, route.Hazard, route.HazardSeverity
}

// GetAllRoutes returns all registered trade routes.
func (s *TradeRouteSystem) GetAllRoutes() []*TradeRoute {
	routes := make([]*TradeRoute, 0, len(s.Routes))
	for _, route := range s.Routes {
		routes = append(routes, route)
	}
	return routes
}

// GetActiveCaravans returns all currently active caravans.
func (s *TradeRouteSystem) GetActiveCaravans() []*TradeCaravan {
	caravans := make([]*TradeCaravan, 0)
	for _, caravan := range s.Caravans {
		if caravan.Status == CaravanTraveling || caravan.Status == CaravanReturning {
			caravans = append(caravans, caravan)
		}
	}
	return caravans
}
