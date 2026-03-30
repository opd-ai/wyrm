package housing

import (
	"testing"
)

func TestHouseManagerRegisterAndGet(t *testing.T) {
	m := NewHouseManager()

	house := &House{
		ID:      "house1",
		OwnerID: 1,
		WorldX:  100,
		WorldZ:  200,
	}
	m.RegisterHouse(house)

	got := m.GetHouse("house1")
	if got == nil {
		t.Fatal("GetHouse returned nil")
	}
	if got.OwnerID != 1 {
		t.Errorf("OwnerID = %d, want 1", got.OwnerID)
	}
}

func TestHouseManagerGetPlayerHouses(t *testing.T) {
	m := NewHouseManager()

	m.RegisterHouse(&House{ID: "house1", OwnerID: 1})
	m.RegisterHouse(&House{ID: "house2", OwnerID: 1})
	m.RegisterHouse(&House{ID: "house3", OwnerID: 2})

	houses := m.GetPlayerHouses(1)
	if len(houses) != 2 {
		t.Errorf("GetPlayerHouses(1) returned %d houses, want 2", len(houses))
	}

	houses = m.GetPlayerHouses(2)
	if len(houses) != 1 {
		t.Errorf("GetPlayerHouses(2) returned %d houses, want 1", len(houses))
	}

	houses = m.GetPlayerHouses(3)
	if len(houses) != 0 {
		t.Errorf("GetPlayerHouses(3) returned %d houses, want 0", len(houses))
	}
}

func TestPlaceFurniture(t *testing.T) {
	m := NewHouseManager()
	m.RegisterHouse(&House{ID: "house1", OwnerID: 1})

	item := FurnitureItem{
		ID:   "bed1",
		Type: "bed",
		X:    5, Y: 0, Z: 5,
		Rotation: 0,
	}

	if err := m.PlaceFurniture("house1", item); err != nil {
		t.Errorf("PlaceFurniture failed: %v", err)
	}

	house := m.GetHouse("house1")
	if len(house.Furniture) != 1 {
		t.Errorf("Furniture count = %d, want 1", len(house.Furniture))
	}
	if house.Furniture[0].Type != "bed" {
		t.Errorf("Furniture type = %s, want bed", house.Furniture[0].Type)
	}
}

func TestRemoveFurniture(t *testing.T) {
	m := NewHouseManager()
	m.RegisterHouse(&House{
		ID:      "house1",
		OwnerID: 1,
		Furniture: []FurnitureItem{
			{ID: "bed1", Type: "bed"},
			{ID: "table1", Type: "table"},
		},
	})

	if err := m.RemoveFurniture("house1", "bed1"); err != nil {
		t.Errorf("RemoveFurniture failed: %v", err)
	}

	house := m.GetHouse("house1")
	if len(house.Furniture) != 1 {
		t.Errorf("Furniture count = %d, want 1", len(house.Furniture))
	}
	if house.Furniture[0].ID != "table1" {
		t.Errorf("Remaining furniture = %s, want table1", house.Furniture[0].ID)
	}

	// Non-existent furniture
	err := m.RemoveFurniture("house1", "nonexistent")
	if err == nil {
		t.Error("RemoveFurniture should error for non-existent furniture")
	}
}

func TestMoveFurniture(t *testing.T) {
	m := NewHouseManager()
	m.RegisterHouse(&House{
		ID:      "house1",
		OwnerID: 1,
		Furniture: []FurnitureItem{
			{ID: "bed1", Type: "bed", X: 0, Y: 0, Z: 0, Rotation: 0},
		},
	})

	if err := m.MoveFurniture("house1", "bed1", 10, 0, 10, 1.57); err != nil {
		t.Errorf("MoveFurniture failed: %v", err)
	}

	house := m.GetHouse("house1")
	if house.Furniture[0].X != 10 {
		t.Errorf("Furniture X = %f, want 10", house.Furniture[0].X)
	}
	if house.Furniture[0].Rotation != 1.57 {
		t.Errorf("Furniture Rotation = %f, want 1.57", house.Furniture[0].Rotation)
	}
}

func TestTransferOwnership(t *testing.T) {
	m := NewHouseManager()
	m.RegisterHouse(&House{ID: "house1", OwnerID: 1})

	if err := m.TransferOwnership("house1", 2); err != nil {
		t.Errorf("TransferOwnership failed: %v", err)
	}

	// Player 1 should have no houses
	houses := m.GetPlayerHouses(1)
	if len(houses) != 0 {
		t.Errorf("Old owner still has %d houses", len(houses))
	}

	// Player 2 should have the house
	houses = m.GetPlayerHouses(2)
	if len(houses) != 1 {
		t.Errorf("New owner has %d houses, want 1", len(houses))
	}
}

func TestExportImportHouses(t *testing.T) {
	m := NewHouseManager()
	m.RegisterHouse(&House{
		ID:      "house1",
		OwnerID: 1,
		Furniture: []FurnitureItem{
			{ID: "bed1", Type: "bed", X: 5, Y: 0, Z: 5},
		},
	})
	m.RegisterHouse(&House{ID: "house2", OwnerID: 2})

	exported := m.ExportAll()
	if len(exported) != 2 {
		t.Errorf("Exported %d houses, want 2", len(exported))
	}

	// Import into new manager
	m2 := NewHouseManager()
	m2.ImportAll(exported)

	if m2.HouseCount() != 2 {
		t.Errorf("Imported HouseCount = %d, want 2", m2.HouseCount())
	}

	house := m2.GetHouse("house1")
	if house == nil {
		t.Fatal("house1 not found after import")
	}
	if len(house.Furniture) != 1 {
		t.Errorf("house1 furniture count = %d, want 1", len(house.Furniture))
	}

	// Per AC: furniture persists
	if house.Furniture[0].X != 5 {
		t.Errorf("Furniture X = %f, want 5 (persistence check)", house.Furniture[0].X)
	}
}

func TestGuildTerritoryContainsPoint(t *testing.T) {
	territory := &GuildTerritory{
		CenterX: 100,
		CenterZ: 100,
		Radius:  50,
	}

	tests := []struct {
		x, z     float64
		expected bool
	}{
		{100, 100, true},  // center
		{150, 100, true},  // edge
		{100, 150, true},  // edge
		{100, 151, false}, // just outside
		{0, 0, false},     // far outside
	}

	for _, tt := range tests {
		got := territory.ContainsPoint(tt.x, tt.z)
		if got != tt.expected {
			t.Errorf("ContainsPoint(%f, %f) = %v, want %v", tt.x, tt.z, got, tt.expected)
		}
	}
}

func TestGuildManagerClaimTerritory(t *testing.T) {
	m := NewGuildManager()

	t1 := &GuildTerritory{
		GuildID: "guild1",
		Name:    "Territory 1",
		CenterX: 100, CenterZ: 100,
		Radius: 50,
	}

	if err := m.ClaimTerritory(t1); err != nil {
		t.Errorf("ClaimTerritory failed: %v", err)
	}

	if m.TerritoryCount() != 1 {
		t.Errorf("TerritoryCount = %d, want 1", m.TerritoryCount())
	}

	// Overlapping territory should fail
	t2 := &GuildTerritory{
		GuildID: "guild2",
		Name:    "Territory 2",
		CenterX: 120, CenterZ: 100, // Overlaps with t1
		Radius: 50,
	}

	if err := m.ClaimTerritory(t2); err == nil {
		t.Error("Overlapping ClaimTerritory should fail")
	}

	// Non-overlapping should succeed
	t3 := &GuildTerritory{
		GuildID: "guild3",
		Name:    "Territory 3",
		CenterX: 300, CenterZ: 300,
		Radius: 50,
	}

	if err := m.ClaimTerritory(t3); err != nil {
		t.Errorf("Non-overlapping ClaimTerritory failed: %v", err)
	}
}

func TestGuildManagerMembership(t *testing.T) {
	m := NewGuildManager()

	m.AddMember("guild1", 1)
	m.AddMember("guild1", 2)
	m.AddMember("guild2", 3)

	if !m.IsMember("guild1", 1) {
		t.Error("Player 1 should be member of guild1")
	}
	if m.IsMember("guild1", 3) {
		t.Error("Player 3 should not be member of guild1")
	}

	members := m.GetMembers("guild1")
	if len(members) != 2 {
		t.Errorf("guild1 has %d members, want 2", len(members))
	}

	m.RemoveMember("guild1", 1)
	if m.IsMember("guild1", 1) {
		t.Error("Player 1 should not be member after removal")
	}
}

func TestCanAccessTerritory(t *testing.T) {
	m := NewGuildManager()

	// Claim territory for guild1
	m.ClaimTerritory(&GuildTerritory{
		GuildID: "guild1",
		CenterX: 100, CenterZ: 100,
		Radius: 50,
	})

	// Add member
	m.AddMember("guild1", 1)

	// Per AC: guild territory boundary enforced
	// Member can access
	if !m.CanAccessTerritory(1, 100, 100) {
		t.Error("Guild member should access territory")
	}

	// Non-member cannot access
	if m.CanAccessTerritory(2, 100, 100) {
		t.Error("Non-member should not access territory")
	}

	// Outside territory is accessible to all
	if !m.CanAccessTerritory(2, 300, 300) {
		t.Error("Outside territory should be accessible")
	}
}

func TestExportImportTerritories(t *testing.T) {
	m := NewGuildManager()
	m.ClaimTerritory(&GuildTerritory{
		GuildID: "guild1",
		Name:    "Territory 1",
		CenterX: 100, CenterZ: 100,
		Radius: 50,
	})
	m.AddMember("guild1", 1)
	m.AddMember("guild1", 2)

	// Export
	territories := m.ExportTerritories()
	members := m.ExportMembers()

	// Import into new manager
	m2 := NewGuildManager()
	m2.ImportTerritories(territories)
	m2.ImportMembers(members)

	// Verify persistence
	if m2.TerritoryCount() != 1 {
		t.Errorf("Imported TerritoryCount = %d, want 1", m2.TerritoryCount())
	}

	territory := m2.GetTerritory("guild1")
	if territory == nil {
		t.Fatal("Territory not found after import")
	}
	if territory.Radius != 50 {
		t.Errorf("Territory Radius = %f, want 50", territory.Radius)
	}

	if !m2.IsMember("guild1", 1) {
		t.Error("Membership not preserved after import")
	}
}

// ============================================================================
// Property Market Tests
// ============================================================================

func TestPropertyMarketAddListing(t *testing.T) {
	hm := NewHouseManager()
	pm := NewPropertyMarket(hm)

	listing := &PropertyListing{
		ID:         "prop1",
		Name:       "Cozy Cottage",
		BasePrice:  1000,
		Size:       1,
		Quality:    1.0,
		DistrictID: "residential",
	}
	pm.AddListing(listing)

	if pm.ListingCount() != 1 {
		t.Errorf("ListingCount = %d, want 1", pm.ListingCount())
	}

	listings := pm.GetAvailableListings()
	if len(listings) != 1 {
		t.Errorf("GetAvailableListings returned %d, want 1", len(listings))
	}
	if listings[0].Name != "Cozy Cottage" {
		t.Errorf("Listing name = %s, want Cozy Cottage", listings[0].Name)
	}
}

func TestPropertyMarketPricing(t *testing.T) {
	hm := NewHouseManager()
	pm := NewPropertyMarket(hm)

	listing := &PropertyListing{
		ID:         "prop1",
		BasePrice:  1000,
		Size:       2,   // Medium
		Quality:    1.0, // Full quality
		DistrictID: "noble",
	}
	pm.AddListing(listing)

	// Without district factor, price = 1000 * 1.0 * 2 = 2000
	price := pm.GetCurrentPrice("prop1")
	if price != 2000 {
		t.Errorf("Base price = %d, want 2000", price)
	}

	// With 1.5x district factor
	pm.SetDistrictPriceFactor("noble", 1.5)
	price = pm.GetCurrentPrice("prop1")
	if price != 3000 {
		t.Errorf("Price with factor = %d, want 3000", price)
	}
}

func TestPropertyMarketPurchaseSuccess(t *testing.T) {
	hm := NewHouseManager()
	pm := NewPropertyMarket(hm)

	listing := &PropertyListing{
		ID:        "prop1",
		Name:      "Starter Home",
		BasePrice: 1000,
		Size:      1,
		Quality:   1.0,
		WorldX:    100,
		WorldZ:    200,
	}
	pm.AddListing(listing)

	result := pm.PurchaseProperty("prop1", 1, 5000, 1)

	if !result.Success {
		t.Errorf("Purchase failed: %s", result.Message)
	}
	if result.PricePaid != 1000 {
		t.Errorf("PricePaid = %d, want 1000", result.PricePaid)
	}
	if result.House == nil {
		t.Fatal("Result house is nil")
	}
	if result.House.OwnerID != 1 {
		t.Errorf("House owner = %d, want 1", result.House.OwnerID)
	}

	// Listing should be removed
	if pm.ListingCount() != 0 {
		t.Error("Listing should be removed after purchase")
	}

	// House should be registered
	if hm.HouseCount() != 1 {
		t.Error("House should be registered in manager")
	}
}

func TestPropertyMarketPurchaseInsufficientFunds(t *testing.T) {
	hm := NewHouseManager()
	pm := NewPropertyMarket(hm)

	listing := &PropertyListing{
		ID:        "prop1",
		BasePrice: 1000,
		Size:      1,
		Quality:   1.0,
	}
	pm.AddListing(listing)

	result := pm.PurchaseProperty("prop1", 1, 500, 1) // Only 500 gold

	if result.Success {
		t.Error("Purchase should fail with insufficient funds")
	}
	if result.Message != "Insufficient gold" {
		t.Errorf("Message = %s, want 'Insufficient gold'", result.Message)
	}

	// Listing should still exist
	if pm.ListingCount() != 1 {
		t.Error("Listing should remain after failed purchase")
	}
}

func TestPropertyMarketPurchaseNotFound(t *testing.T) {
	hm := NewHouseManager()
	pm := NewPropertyMarket(hm)

	result := pm.PurchaseProperty("nonexistent", 1, 5000, 1)

	if result.Success {
		t.Error("Purchase should fail for non-existent property")
	}
	if result.Message != "Property not found" {
		t.Errorf("Message = %s, want 'Property not found'", result.Message)
	}
}

func TestPropertyMarketGetByDistrict(t *testing.T) {
	hm := NewHouseManager()
	pm := NewPropertyMarket(hm)

	pm.AddListing(&PropertyListing{ID: "prop1", DistrictID: "noble"})
	pm.AddListing(&PropertyListing{ID: "prop2", DistrictID: "noble"})
	pm.AddListing(&PropertyListing{ID: "prop3", DistrictID: "slums"})

	noble := pm.GetListingsByDistrict("noble")
	if len(noble) != 2 {
		t.Errorf("Noble district has %d listings, want 2", len(noble))
	}

	slums := pm.GetListingsByDistrict("slums")
	if len(slums) != 1 {
		t.Errorf("Slums district has %d listings, want 1", len(slums))
	}
}

// ============================================================================
// Furniture Placement Tests
// ============================================================================

func TestFurniturePlacementModes(t *testing.T) {
	fp := NewFurniturePlacement()

	if fp.IsInPlacementMode() {
		t.Error("Should not be in placement mode initially")
	}

	// Test place mode
	fp.StartPlaceMode("house1", "bed")
	if !fp.IsInPlacementMode() {
		t.Error("Should be in placement mode after StartPlaceMode")
	}
	if fp.GetCurrentHouse() != "house1" {
		t.Errorf("CurrentHouse = %s, want house1", fp.GetCurrentHouse())
	}

	fp.ExitMode()
	if fp.IsInPlacementMode() {
		t.Error("Should not be in placement mode after ExitMode")
	}

	// Test move mode
	fp.StartMoveMode("house1", "bed1")
	mode, _, _, _, _, _ := fp.GetPreviewState()
	if mode != PlacementModeMove {
		t.Errorf("Mode = %d, want PlacementModeMove", mode)
	}

	// Test rotate mode
	fp.StartRotateMode("house1", "bed1")
	mode, _, _, _, _, _ = fp.GetPreviewState()
	if mode != PlacementModeRotate {
		t.Errorf("Mode = %d, want PlacementModeRotate", mode)
	}

	// Test remove mode
	fp.StartRemoveMode("house1", "bed1")
	mode, _, _, _, _, _ = fp.GetPreviewState()
	if mode != PlacementModeRemove {
		t.Errorf("Mode = %d, want PlacementModeRemove", mode)
	}
}

func TestFurniturePlacementUpdatePreview(t *testing.T) {
	fp := NewFurniturePlacement()
	fp.StartPlaceMode("house1", "bed")
	fp.SetGridSnap(false, 0) // Disable snap for precise test

	// Player at (0,0) looking at angle 0 (east), distance 2
	fp.UpdatePreview(0, 0, 0, 2)

	mode, x, _, z, _, valid := fp.GetPreviewState()
	if mode != PlacementModePlace {
		t.Errorf("Mode = %d, want PlacementModePlace", mode)
	}
	if !valid {
		t.Error("Position should be valid after update")
	}

	// cos(0) = 1, sin(0) = 0, so position should be (2, 0)
	if x < 1.9 || x > 2.1 {
		t.Errorf("Preview X = %f, want ~2", x)
	}
	if z < -0.1 || z > 0.1 {
		t.Errorf("Preview Z = %f, want ~0", z)
	}
}

func TestFurniturePlacementGridSnap(t *testing.T) {
	fp := NewFurniturePlacement()
	fp.StartPlaceMode("house1", "bed")
	fp.SetGridSnap(true, 1.0) // 1-unit grid

	// Position that should snap
	fp.UpdatePreview(0.3, 0.7, 0, 0) // Player at (0.3, 0.7)

	_, x, _, z, _, _ := fp.GetPreviewState()
	// With grid snap at 1.0, 0.3 floors to 0, 0.7 floors to 0
	if x != 0 {
		t.Errorf("Snapped X = %f, want 0", x)
	}
	if z != 0 {
		t.Errorf("Snapped Z = %f, want 0", z)
	}
}

func TestFurniturePlacementRotate(t *testing.T) {
	fp := NewFurniturePlacement()
	fp.StartPlaceMode("house1", "bed")

	fp.RotatePreview(1.57) // ~90 degrees
	_, _, _, _, rot, _ := fp.GetPreviewState()
	if rot < 1.5 || rot > 1.6 {
		t.Errorf("Rotation = %f, want ~1.57", rot)
	}

	// Rotate more to test normalization
	fp.RotatePreview(6.28) // Full rotation
	_, _, _, _, rot, _ = fp.GetPreviewState()
	if rot < 1.5 || rot > 1.7 {
		t.Errorf("Normalized rotation = %f, want ~1.57", rot)
	}
}

func TestFurniturePlacementConfirm(t *testing.T) {
	hm := NewHouseManager()
	hm.RegisterHouse(&House{ID: "house1", OwnerID: 1})

	fp := NewFurniturePlacement()
	fp.StartPlaceMode("house1", "bed")
	fp.SetGridSnap(false, 0)
	fp.UpdatePreview(5, 5, 0, 0) // Place at (5, 5)

	err := fp.ConfirmPlacement(hm, "bed1")
	if err != nil {
		t.Errorf("ConfirmPlacement failed: %v", err)
	}

	// Should exit placement mode
	if fp.IsInPlacementMode() {
		t.Error("Should exit placement mode after confirm")
	}

	// Furniture should be placed
	house := hm.GetHouse("house1")
	if len(house.Furniture) != 1 {
		t.Errorf("House has %d furniture, want 1", len(house.Furniture))
	}
	if house.Furniture[0].ID != "bed1" {
		t.Errorf("Furniture ID = %s, want bed1", house.Furniture[0].ID)
	}
}

func TestFurniturePlacementConfirmNoValidPosition(t *testing.T) {
	hm := NewHouseManager()
	fp := NewFurniturePlacement()

	// Not in placement mode
	err := fp.ConfirmPlacement(hm, "bed1")
	if err == nil {
		t.Error("Should error when not in placement mode")
	}

	// In mode but no valid position
	fp.StartPlaceMode("house1", "bed")
	err = fp.ConfirmPlacement(hm, "bed1")
	if err == nil {
		t.Error("Should error with no valid position")
	}
}
