package systems

import (
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// Building type constants.
const (
	BuildingTypeShop       = "shop"
	BuildingTypeResidence  = "residence"
	BuildingTypeGovernment = "government"
	BuildingTypeIndustrial = "industrial"
	BuildingTypeInn        = "inn"
	BuildingTypeTemple     = "temple"
)

// POI icon type constants.
const (
	POIIconShop       = "shop"
	POIIconQuest      = "quest"
	POIIconDanger     = "danger"
	POIIconGuild      = "guild"
	POIIconInn        = "inn"
	POIIconBlacksmith = "blacksmith"
	POIIconTemple     = "temple"
	POIIconGate       = "gate"
)

// Shop type constants.
const (
	ShopTypeGeneral    = "general"
	ShopTypeBlacksmith = "blacksmith"
	ShopTypeAlchemist  = "alchemist"
	ShopTypeTailor     = "tailor"
	ShopTypeWeapons    = "weapons"
	ShopTypeArmor      = "armor"
	ShopTypeMagic      = "magic"
	ShopTypeFood       = "food"
)

// CityBuildingSystem manages city buildings, POIs, and shop operations.
type CityBuildingSystem struct {
	// Genre affects building styles and shop types.
	Genre string
	// Seed for deterministic generation.
	Seed int64
}

// NewCityBuildingSystem creates a new city building system.
func NewCityBuildingSystem(genre string, seed int64) *CityBuildingSystem {
	return &CityBuildingSystem{
		Genre: genre,
		Seed:  seed,
	}
}

// Update processes shop restocking and building operations.
func (s *CityBuildingSystem) Update(w *ecs.World, dt float64) {
	currentHour := s.getCurrentHour(w)
	s.updateBuildingOpenStatus(w, currentHour)
	s.updateShopRestocking(w, currentHour)
}

// getCurrentHour retrieves the current game hour from WorldClock.
func (s *CityBuildingSystem) getCurrentHour(w *ecs.World) int {
	for _, e := range w.Entities("WorldClock") {
		clockComp, ok := w.GetComponent(e, "WorldClock")
		if !ok {
			continue
		}
		clock := clockComp.(*components.WorldClock)
		return clock.Hour
	}
	return 12 // Default to noon
}

// updateBuildingOpenStatus updates IsOpen based on operating hours.
func (s *CityBuildingSystem) updateBuildingOpenStatus(w *ecs.World, currentHour int) {
	for _, e := range w.Entities("Building") {
		buildingComp, ok := w.GetComponent(e, "Building")
		if !ok {
			continue
		}
		building := buildingComp.(*components.Building)

		// Residences are always closed to public
		if building.BuildingType == BuildingTypeResidence {
			building.IsOpen = false
			continue
		}

		// Check operating hours
		if building.OpenHour <= building.CloseHour {
			building.IsOpen = currentHour >= building.OpenHour && currentHour < building.CloseHour
		} else {
			// Handles overnight hours (e.g., 22:00 - 06:00)
			building.IsOpen = currentHour >= building.OpenHour || currentHour < building.CloseHour
		}
	}
}

// updateShopRestocking restocks shops when their interval passes.
func (s *CityBuildingSystem) updateShopRestocking(w *ecs.World, currentHour int) {
	for _, e := range w.Entities("ShopInventory") {
		shopComp, ok := w.GetComponent(e, "ShopInventory")
		if !ok {
			continue
		}
		shop := shopComp.(*components.ShopInventory)

		// Calculate hours since last restock
		hoursSinceRestock := currentHour - shop.LastRestock
		if hoursSinceRestock < 0 {
			hoursSinceRestock += HoursPerDay
		}

		if hoursSinceRestock >= shop.RestockInterval {
			s.restockShop(shop, e)
			shop.LastRestock = currentHour
		}
	}
}

// restockShop refills a shop's inventory.
func (s *CityBuildingSystem) restockShop(shop *components.ShopInventory, entity ecs.Entity) {
	rng := rand.New(rand.NewSource(s.Seed + int64(entity)))

	baseStock := s.getBaseStockForShopType(shop.ShopType)
	for itemID, baseQty := range baseStock {
		// Add some random variance
		variance := rng.Intn(baseQty/2+1) - baseQty/4
		qty := baseQty + variance
		if qty < 1 {
			qty = 1
		}
		shop.Items[itemID] = qty
	}

	// Restore gold reserve
	shop.GoldReserve = s.getGoldReserveForShopType(shop.ShopType)
}

// getBaseStockForShopType returns base inventory for a shop type.
func (s *CityBuildingSystem) getBaseStockForShopType(shopType string) map[string]int {
	switch shopType {
	case ShopTypeGeneral:
		return map[string]int{"torch": 10, "rope": 5, "rations": 20, "bandage": 10, "waterskin": 5}
	case ShopTypeBlacksmith:
		return map[string]int{"iron_sword": 3, "iron_shield": 2, "iron_armor": 2, "repair_kit": 5}
	case ShopTypeAlchemist:
		return map[string]int{"health_potion": 10, "mana_potion": 8, "antidote": 5, "buff_potion": 3}
	case ShopTypeWeapons:
		return map[string]int{"sword": 4, "axe": 3, "bow": 2, "arrows": 50, "dagger": 5}
	case ShopTypeArmor:
		return map[string]int{"leather_armor": 3, "chainmail": 2, "helmet": 4, "shield": 3}
	case ShopTypeMagic:
		return map[string]int{"scroll_fireball": 2, "scroll_heal": 3, "wand": 1, "staff": 1}
	case ShopTypeFood:
		return map[string]int{"bread": 15, "meat": 10, "fruit": 12, "ale": 20, "wine": 5}
	default:
		return map[string]int{"misc_item": 5}
	}
}

// getGoldReserveForShopType returns base gold reserve for a shop type.
func (s *CityBuildingSystem) getGoldReserveForShopType(shopType string) float64 {
	switch shopType {
	case ShopTypeGeneral:
		return 200.0
	case ShopTypeBlacksmith:
		return 500.0
	case ShopTypeAlchemist:
		return 300.0
	case ShopTypeWeapons:
		return 400.0
	case ShopTypeArmor:
		return 450.0
	case ShopTypeMagic:
		return 600.0
	case ShopTypeFood:
		return 150.0
	default:
		return 100.0
	}
}

// CreateShopBuilding creates a shop building with interior and inventory.
func (s *CityBuildingSystem) CreateShopBuilding(w *ecs.World, shopType, name string, x, y, z float64, factionID string) ecs.Entity {
	building := w.CreateEntity()

	// Add position
	w.AddComponent(building, &components.Position{X: x, Y: y, Z: z})

	// Add building component
	w.AddComponent(building, &components.Building{
		BuildingType: BuildingTypeShop,
		Name:         name,
		OwnerFaction: factionID,
		Floors:       1,
		Width:        10.0,
		Height:       8.0,
		EntranceX:    x + 5.0,
		EntranceY:    y,
		EntranceZ:    z,
		IsOpen:       true,
		OpenHour:     8,
		CloseHour:    20,
	})

	// Add shop inventory
	items := s.getBaseStockForShopType(shopType)
	prices := make(map[string]float64)
	for itemID := range items {
		prices[itemID] = s.getBasePriceForItem(itemID)
	}

	w.AddComponent(building, &components.ShopInventory{
		ShopType:        shopType,
		Items:           items,
		Prices:          prices,
		RestockInterval: 24,
		LastRestock:     0,
		GoldReserve:     s.getGoldReserveForShopType(shopType),
	})

	// Add POI marker
	w.AddComponent(building, &components.POIMarker{
		IconType:          s.getPOIIconForShopType(shopType),
		Name:              name,
		Description:       shopType + " shop",
		Visible:           true,
		MinimapVisible:    true,
		DiscoveryRequired: false,
		Discovered:        true,
	})

	// Create and link interior
	interior := s.CreateShopInterior(w, shopType)
	buildingComp, _ := w.GetComponent(building, "Building")
	buildingComp.(*components.Building).InteriorEntity = uint64(interior)

	return building
}

// CreateShopInterior creates an interior for a shop.
func (s *CityBuildingSystem) CreateShopInterior(w *ecs.World, shopType string) ecs.Entity {
	interior := w.CreateEntity()

	// Generate shop layout based on type
	width, height := 8, 6
	rooms := []components.Room{
		{ID: "main", Name: "Shop Floor", X: 0, Y: 0, Width: 6, Height: 6, Purpose: "shop"},
		{ID: "storage", Name: "Storage", X: 6, Y: 0, Width: 2, Height: 6, Purpose: "storage"},
	}

	// Create wall tiles
	wallTiles := make([][]int, height)
	for y := 0; y < height; y++ {
		wallTiles[y] = make([]int, width)
		for x := 0; x < width; x++ {
			// Walls on edges
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				wallTiles[y][x] = 1
			}
			// Internal wall between shop and storage
			if x == 6 && y > 1 && y < height-1 {
				wallTiles[y][x] = 1
			}
		}
	}
	// Door opening
	wallTiles[height-1][3] = 0

	w.AddComponent(interior, &components.Interior{
		Width:     width,
		Height:    height,
		Rooms:     rooms,
		Furniture: []uint64{},
		WallTiles: wallTiles,
		FloorType: s.getFloorTypeForShopType(shopType),
	})

	return interior
}

// CreateGovernmentBuilding creates a government/faction building.
func (s *CityBuildingSystem) CreateGovernmentBuilding(w *ecs.World, govType, name string, x, y, z float64, factionID string) ecs.Entity {
	building := w.CreateEntity()

	w.AddComponent(building, &components.Position{X: x, Y: y, Z: z})

	w.AddComponent(building, &components.Building{
		BuildingType: BuildingTypeGovernment,
		Name:         name,
		OwnerFaction: factionID,
		Floors:       2,
		Width:        20.0,
		Height:       15.0,
		EntranceX:    x + 10.0,
		EntranceY:    y,
		EntranceZ:    z,
		IsOpen:       true,
		OpenHour:     6,
		CloseHour:    22,
	})

	services := s.getServicesForGovType(govType)
	npcRoles := s.getNPCRolesForGovType(govType)

	w.AddComponent(building, &components.GovernmentBuilding{
		GovernmentType:     govType,
		ControllingFaction: factionID,
		Services:           services,
		NPCRoles:           npcRoles,
	})

	w.AddComponent(building, &components.POIMarker{
		IconType:          POIIconGuild,
		Name:              name,
		Description:       govType + " building",
		Visible:           true,
		MinimapVisible:    true,
		DiscoveryRequired: false,
		Discovered:        true,
	})

	return building
}

// CreatePOI creates a standalone point of interest marker.
func (s *CityBuildingSystem) CreatePOI(w *ecs.World, iconType, name, description string, x, y, z float64, requiresDiscovery bool) ecs.Entity {
	poi := w.CreateEntity()

	w.AddComponent(poi, &components.Position{X: x, Y: y, Z: z})
	w.AddComponent(poi, &components.POIMarker{
		IconType:          iconType,
		Name:              name,
		Description:       description,
		Visible:           true,
		MinimapVisible:    true,
		DiscoveryRequired: requiresDiscovery,
		Discovered:        !requiresDiscovery,
	})

	return poi
}

// DiscoverPOI marks a POI as discovered for the player.
func (s *CityBuildingSystem) DiscoverPOI(w *ecs.World, poi ecs.Entity) bool {
	poiComp, ok := w.GetComponent(poi, "POIMarker")
	if !ok {
		return false
	}
	marker := poiComp.(*components.POIMarker)
	marker.Discovered = true
	return true
}

// GetNearbyPOIs returns POIs within radius of a position.
func (s *CityBuildingSystem) GetNearbyPOIs(w *ecs.World, x, y, z, radius float64, onlyDiscovered bool) []ecs.Entity {
	var result []ecs.Entity
	radiusSq := radius * radius

	for _, e := range w.Entities("POIMarker", "Position") {
		posComp, _ := w.GetComponent(e, "Position")
		pos := posComp.(*components.Position)

		dx := pos.X - x
		dy := pos.Y - y
		dz := pos.Z - z
		distSq := dx*dx + dy*dy + dz*dz

		if distSq > radiusSq {
			continue
		}

		poiComp, _ := w.GetComponent(e, "POIMarker")
		poi := poiComp.(*components.POIMarker)

		if onlyDiscovered && !poi.Discovered && poi.DiscoveryRequired {
			continue
		}

		result = append(result, e)
	}

	return result
}

// Helper functions

func (s *CityBuildingSystem) getBasePriceForItem(itemID string) float64 {
	prices := map[string]float64{
		"torch": 5, "rope": 10, "rations": 3, "bandage": 8, "waterskin": 12,
		"iron_sword": 100, "iron_shield": 80, "iron_armor": 150, "repair_kit": 25,
		"health_potion": 50, "mana_potion": 50, "antidote": 30, "buff_potion": 75,
		"sword": 80, "axe": 70, "bow": 90, "arrows": 1, "dagger": 30,
		"leather_armor": 60, "chainmail": 120, "helmet": 50, "shield": 70,
		"scroll_fireball": 100, "scroll_heal": 80, "wand": 150, "staff": 200,
		"bread": 2, "meat": 5, "fruit": 3, "ale": 4, "wine": 15,
	}
	if price, exists := prices[itemID]; exists {
		return price
	}
	return 10.0
}

func (s *CityBuildingSystem) getPOIIconForShopType(shopType string) string {
	switch shopType {
	case ShopTypeBlacksmith:
		return POIIconBlacksmith
	default:
		return POIIconShop
	}
}

func (s *CityBuildingSystem) getFloorTypeForShopType(shopType string) string {
	switch shopType {
	case ShopTypeBlacksmith:
		return "stone"
	case ShopTypeAlchemist:
		return "tile"
	default:
		return "wood"
	}
}

func (s *CityBuildingSystem) getServicesForGovType(govType string) []string {
	switch govType {
	case "barracks":
		return []string{"training", "quest_board", "bounty_payment"}
	case "courthouse":
		return []string{"bounty_payment", "pardons", "disputes"}
	case "guild_hall":
		return []string{"quest_board", "storage", "training", "recruitment"}
	case "palace":
		return []string{"audience", "quests", "politics"}
	case "prison":
		return []string{"incarceration", "bounty_payment"}
	default:
		return []string{"general_services"}
	}
}

func (s *CityBuildingSystem) getNPCRolesForGovType(govType string) []string {
	switch govType {
	case "barracks":
		return []string{"guard", "captain", "trainer"}
	case "courthouse":
		return []string{"judge", "clerk", "guard"}
	case "guild_hall":
		return []string{"guildmaster", "clerk", "trainer"}
	case "palace":
		return []string{"ruler", "advisor", "guard", "servant"}
	case "prison":
		return []string{"warden", "guard"}
	default:
		return []string{"clerk"}
	}
}
