package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestCityBuildingSystem_CreateShopBuilding(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	shop := sys.CreateShopBuilding(w, ShopTypeGeneral, "Bob's General Store", 100, 100, 0, "merchant_guild")

	// Check building component
	buildingComp, ok := w.GetComponent(shop, "Building")
	if !ok {
		t.Fatal("Shop should have Building component")
	}
	building := buildingComp.(*components.Building)

	if building.BuildingType != BuildingTypeShop {
		t.Errorf("Expected shop building type, got %s", building.BuildingType)
	}
	if building.Name != "Bob's General Store" {
		t.Errorf("Expected name 'Bob's General Store', got %s", building.Name)
	}

	// Check shop inventory
	shopComp, ok := w.GetComponent(shop, "ShopInventory")
	if !ok {
		t.Fatal("Shop should have ShopInventory component")
	}
	inventory := shopComp.(*components.ShopInventory)

	if inventory.ShopType != ShopTypeGeneral {
		t.Errorf("Expected general shop type, got %s", inventory.ShopType)
	}
	if len(inventory.Items) == 0 {
		t.Error("Shop should have items in inventory")
	}

	// Check POI marker
	poiComp, ok := w.GetComponent(shop, "POIMarker")
	if !ok {
		t.Fatal("Shop should have POIMarker component")
	}
	poi := poiComp.(*components.POIMarker)

	if poi.IconType != POIIconShop {
		t.Errorf("Expected shop icon type, got %s", poi.IconType)
	}
	if !poi.Visible {
		t.Error("Shop POI should be visible")
	}

	// Check interior was created
	if building.InteriorEntity == 0 {
		t.Error("Shop should have linked interior entity")
	}
}

func TestCityBuildingSystem_CreateGovernmentBuilding(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	gov := sys.CreateGovernmentBuilding(w, "guild_hall", "Adventurer's Guild", 200, 200, 0, "adventurers")

	// Check building component
	buildingComp, ok := w.GetComponent(gov, "Building")
	if !ok {
		t.Fatal("Government building should have Building component")
	}
	building := buildingComp.(*components.Building)

	if building.BuildingType != BuildingTypeGovernment {
		t.Errorf("Expected government building type, got %s", building.BuildingType)
	}

	// Check government building component
	govComp, ok := w.GetComponent(gov, "GovernmentBuilding")
	if !ok {
		t.Fatal("Should have GovernmentBuilding component")
	}
	govData := govComp.(*components.GovernmentBuilding)

	if govData.GovernmentType != "guild_hall" {
		t.Errorf("Expected guild_hall type, got %s", govData.GovernmentType)
	}
	if len(govData.Services) == 0 {
		t.Error("Guild hall should have services")
	}
	if len(govData.NPCRoles) == 0 {
		t.Error("Guild hall should have NPC roles")
	}
}

func TestCityBuildingSystem_CreatePOI(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	poi := sys.CreatePOI(w, POIIconDanger, "Dragon's Lair", "A dangerous cave", 500, 500, 0, true)

	poiComp, ok := w.GetComponent(poi, "POIMarker")
	if !ok {
		t.Fatal("Should have POIMarker component")
	}
	marker := poiComp.(*components.POIMarker)

	if marker.IconType != POIIconDanger {
		t.Errorf("Expected danger icon, got %s", marker.IconType)
	}
	if !marker.DiscoveryRequired {
		t.Error("POI should require discovery")
	}
	if marker.Discovered {
		t.Error("POI should not be discovered initially")
	}
}

func TestCityBuildingSystem_DiscoverPOI(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	poi := sys.CreatePOI(w, POIIconQuest, "Hidden Shrine", "An ancient shrine", 300, 300, 0, true)

	// Initially not discovered
	poiComp, _ := w.GetComponent(poi, "POIMarker")
	marker := poiComp.(*components.POIMarker)

	if marker.Discovered {
		t.Error("POI should not be discovered initially")
	}

	// Discover it
	success := sys.DiscoverPOI(w, poi)
	if !success {
		t.Error("DiscoverPOI should succeed")
	}
	if !marker.Discovered {
		t.Error("POI should be discovered after DiscoverPOI")
	}
}

func TestCityBuildingSystem_GetNearbyPOIs(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	// Create POIs at different distances
	sys.CreatePOI(w, POIIconShop, "Nearby Shop", "A shop", 10, 10, 0, false)
	sys.CreatePOI(w, POIIconInn, "Far Inn", "An inn", 1000, 1000, 0, false)
	sys.CreatePOI(w, POIIconDanger, "Hidden Cave", "A cave", 20, 20, 0, true) // Undiscovered

	// Find POIs within 100 units of origin, including undiscovered
	pois := sys.GetNearbyPOIs(w, 0, 0, 0, 100, false)
	if len(pois) != 2 {
		t.Errorf("Expected 2 nearby POIs, got %d", len(pois))
	}

	// Find only discovered POIs
	discoveredPOIs := sys.GetNearbyPOIs(w, 0, 0, 0, 100, true)
	if len(discoveredPOIs) != 1 {
		t.Errorf("Expected 1 discovered nearby POI, got %d", len(discoveredPOIs))
	}
}

func TestCityBuildingSystem_BuildingOpenStatus(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	// Create world clock at different hours
	createWorldClockAt := func(hour int) {
		// Remove existing clocks
		for _, e := range w.Entities("WorldClock") {
			w.DestroyEntity(e)
		}
		clockEntity := w.CreateEntity()
		w.AddComponent(clockEntity, &components.WorldClock{Hour: hour, Day: 1, HourLength: 60.0})
	}

	shop := sys.CreateShopBuilding(w, ShopTypeGeneral, "Test Shop", 0, 0, 0, "")

	// During open hours (8-20)
	createWorldClockAt(12)
	sys.Update(w, 1.0)

	buildingComp, _ := w.GetComponent(shop, "Building")
	building := buildingComp.(*components.Building)

	if !building.IsOpen {
		t.Error("Shop should be open at noon")
	}

	// During closed hours
	createWorldClockAt(3)
	sys.Update(w, 1.0)

	if building.IsOpen {
		t.Error("Shop should be closed at 3 AM")
	}
}

func TestCityBuildingSystem_ShopRestocking(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	// Create clock
	clockEntity := w.CreateEntity()
	w.AddComponent(clockEntity, &components.WorldClock{Hour: 8, Day: 1, HourLength: 60.0})

	shop := sys.CreateShopBuilding(w, ShopTypeGeneral, "Test Shop", 0, 0, 0, "")

	shopComp, _ := w.GetComponent(shop, "ShopInventory")
	inventory := shopComp.(*components.ShopInventory)

	// Deplete inventory
	inventory.Items["torch"] = 0
	inventory.LastRestock = 0

	// Fast forward past restock interval
	clockComp, _ := w.GetComponent(clockEntity, "WorldClock")
	clock := clockComp.(*components.WorldClock)
	clock.Hour = inventory.RestockInterval + 1

	sys.Update(w, 1.0)

	if inventory.Items["torch"] == 0 {
		t.Error("Shop should have restocked torches")
	}
}

func TestCityBuildingSystem_CreateShopInterior(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	interior := sys.CreateShopInterior(w, ShopTypeBlacksmith)

	interiorComp, ok := w.GetComponent(interior, "Interior")
	if !ok {
		t.Fatal("Interior should have Interior component")
	}
	interiorData := interiorComp.(*components.Interior)

	if interiorData.Width == 0 || interiorData.Height == 0 {
		t.Error("Interior should have dimensions")
	}
	if len(interiorData.Rooms) == 0 {
		t.Error("Interior should have rooms")
	}
	if interiorData.WallTiles == nil {
		t.Error("Interior should have wall tiles")
	}
	if interiorData.FloorType != "stone" {
		t.Errorf("Blacksmith should have stone floor, got %s", interiorData.FloorType)
	}
}

func TestCityBuildingSystem_ShopTypes(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	shopTypes := []string{
		ShopTypeGeneral, ShopTypeBlacksmith, ShopTypeAlchemist,
		ShopTypeWeapons, ShopTypeArmor, ShopTypeMagic, ShopTypeFood,
	}

	for _, shopType := range shopTypes {
		t.Run(shopType, func(t *testing.T) {
			shop := sys.CreateShopBuilding(w, shopType, shopType+" Shop", 0, 0, 0, "")

			shopComp, ok := w.GetComponent(shop, "ShopInventory")
			if !ok {
				t.Fatal("Shop should have inventory")
			}
			inventory := shopComp.(*components.ShopInventory)

			if len(inventory.Items) == 0 {
				t.Errorf("%s shop should have items", shopType)
			}
			if inventory.GoldReserve <= 0 {
				t.Errorf("%s shop should have gold reserve", shopType)
			}
		})
	}
}

func TestCityBuildingSystem_GovernmentTypes(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCityBuildingSystem("fantasy", 12345)

	govTypes := []string{"barracks", "courthouse", "guild_hall", "palace", "prison"}

	for _, govType := range govTypes {
		t.Run(govType, func(t *testing.T) {
			gov := sys.CreateGovernmentBuilding(w, govType, govType+" Building", 0, 0, 0, "faction")

			govComp, ok := w.GetComponent(gov, "GovernmentBuilding")
			if !ok {
				t.Fatal("Should have GovernmentBuilding component")
			}
			govData := govComp.(*components.GovernmentBuilding)

			if len(govData.Services) == 0 {
				t.Errorf("%s should have services", govType)
			}
			if len(govData.NPCRoles) == 0 {
				t.Errorf("%s should have NPC roles", govType)
			}
		})
	}
}
