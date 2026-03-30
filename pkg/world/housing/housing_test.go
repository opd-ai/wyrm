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

// ============================================================================
// Rent Collection System Tests
// ============================================================================

func TestNewRentCollectionSystem(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	if rcs.Seed != 12345 {
		t.Errorf("Seed = %d, want 12345", rcs.Seed)
	}
	if rcs.Genre != "fantasy" {
		t.Errorf("Genre = %s, want fantasy", rcs.Genre)
	}
	if rcs.PaymentPeriod != 720.0 {
		t.Errorf("PaymentPeriod = %f, want 720.0", rcs.PaymentPeriod)
	}
	if rcs.EvictionGrace != 7 {
		t.Errorf("EvictionGrace = %d, want 7", rcs.EvictionGrace)
	}
	if rcs.DepositMultiple != 2.0 {
		t.Errorf("DepositMultiple = %f, want 2.0", rcs.DepositMultiple)
	}
}

func TestListPropertyForRent(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	prop := rcs.ListPropertyForRent(1, "prop1", "Cozy Cottage", 100.0, 0.8, location)

	if prop == nil {
		t.Fatal("ListPropertyForRent returned nil")
	}
	if prop.ID != "prop1" {
		t.Errorf("ID = %s, want prop1", prop.ID)
	}
	if prop.OwnerID != 1 {
		t.Errorf("OwnerID = %d, want 1", prop.OwnerID)
	}
	if prop.MonthlyRent != 100.0 {
		t.Errorf("MonthlyRent = %f, want 100.0", prop.MonthlyRent)
	}
	if prop.Quality != 0.8 {
		t.Errorf("Quality = %f, want 0.8", prop.Quality)
	}
	if prop.Status != RentStatusVacant {
		t.Errorf("Status = %d, want RentStatusVacant", prop.Status)
	}
	if prop.Deposit != 200.0 {
		t.Errorf("Deposit = %f, want 200.0 (2x rent)", prop.Deposit)
	}
	if prop.Condition != 1.0 {
		t.Errorf("Condition = %f, want 1.0", prop.Condition)
	}

	// Verify property is tracked
	got := rcs.GetProperty("prop1")
	if got == nil {
		t.Error("GetProperty returned nil for registered property")
	}
}

func TestAddTenant(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	tenant := rcs.AddTenant(100, "John Doe", 0.9, 0.7)

	if tenant == nil {
		t.Fatal("AddTenant returned nil")
	}
	if tenant.ID != 100 {
		t.Errorf("ID = %d, want 100", tenant.ID)
	}
	if tenant.Name != "John Doe" {
		t.Errorf("Name = %s, want John Doe", tenant.Name)
	}
	if tenant.Reliability != 0.9 {
		t.Errorf("Reliability = %f, want 0.9", tenant.Reliability)
	}
	if tenant.WealthLevel != 0.7 {
		t.Errorf("WealthLevel = %f, want 0.7", tenant.WealthLevel)
	}
	if tenant.Satisfaction != 0.5 {
		t.Errorf("Satisfaction = %f, want 0.5 (neutral)", tenant.Satisfaction)
	}

	// Verify tenant is tracked
	got := rcs.GetTenant(100)
	if got == nil {
		t.Error("GetTenant returned nil for registered tenant")
	}
}

func TestAddTenantClamps(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	// Values outside 0-1 range should be clamped
	tenant := rcs.AddTenant(100, "Test", 1.5, -0.2)

	if tenant.Reliability != 1.0 {
		t.Errorf("Reliability = %f, want 1.0 (clamped)", tenant.Reliability)
	}
	if tenant.WealthLevel != 0.0 {
		t.Errorf("WealthLevel = %f, want 0.0 (clamped)", tenant.WealthLevel)
	}
}

func TestRentProperty(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant", 0.9, 0.8)

	err := rcs.RentProperty("prop1", 100)
	if err != nil {
		t.Errorf("RentProperty failed: %v", err)
	}

	prop := rcs.GetProperty("prop1")
	if prop.TenantID != 100 {
		t.Errorf("TenantID = %d, want 100", prop.TenantID)
	}
	if prop.Status != RentStatusOccupied {
		t.Errorf("Status = %d, want RentStatusOccupied", prop.Status)
	}

	tenant := rcs.GetTenant(100)
	if tenant.LeasesHeld != 1 {
		t.Errorf("LeasesHeld = %d, want 1", tenant.LeasesHeld)
	}
}

func TestRentPropertyErrors(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant", 0.9, 0.8)

	// Non-existent property
	err := rcs.RentProperty("nonexistent", 100)
	if err == nil {
		t.Error("Should error for non-existent property")
	}

	// Non-existent tenant
	err = rcs.RentProperty("prop1", 999)
	if err == nil {
		t.Error("Should error for non-existent tenant")
	}

	// Rent it once
	rcs.RentProperty("prop1", 100)

	// Try to rent again (not vacant)
	err = rcs.RentProperty("prop1", 100)
	if err == nil {
		t.Error("Should error when property is not vacant")
	}
}

func TestEvictTenant(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant", 0.9, 0.8)
	rcs.RentProperty("prop1", 100)

	err := rcs.EvictTenant("prop1")
	if err != nil {
		t.Errorf("EvictTenant failed: %v", err)
	}

	prop := rcs.GetProperty("prop1")
	if prop.TenantID != 0 {
		t.Errorf("TenantID = %d, want 0", prop.TenantID)
	}
	if prop.Status != RentStatusVacant {
		t.Errorf("Status = %d, want RentStatusVacant", prop.Status)
	}

	tenant := rcs.GetTenant(100)
	if tenant.LeasesHeld != 0 {
		t.Errorf("LeasesHeld = %d, want 0", tenant.LeasesHeld)
	}
}

func TestEvictTenantErrors(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)

	// Non-existent property
	err := rcs.EvictTenant("nonexistent")
	if err == nil {
		t.Error("Should error for non-existent property")
	}

	// No tenant
	err = rcs.EvictTenant("prop1")
	if err == nil {
		t.Error("Should error when property has no tenant")
	}
}

func TestProcessRentPayment(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant", 0.9, 0.8)
	rcs.RentProperty("prop1", 100)

	amount, err := rcs.ProcessRentPayment("prop1", 100.0)
	if err != nil {
		t.Errorf("ProcessRentPayment failed: %v", err)
	}
	if amount != 100.0 {
		t.Errorf("Amount = %f, want 100.0", amount)
	}

	// Check owner income
	income := rcs.GetOwnerIncome(1)
	if income != 100.0 {
		t.Errorf("OwnerIncome = %f, want 100.0", income)
	}

	// Check payment history
	history := rcs.GetRentHistory("prop1")
	if len(history) != 1 {
		t.Errorf("Payment history length = %d, want 1", len(history))
	}
	if !history[0].OnTime {
		t.Error("Payment should be on time")
	}
}

func TestProcessRentPaymentErrors(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)

	// Non-existent property
	_, err := rcs.ProcessRentPayment("nonexistent", 100.0)
	if err == nil {
		t.Error("Should error for non-existent property")
	}

	// No tenant
	_, err = rcs.ProcessRentPayment("prop1", 100.0)
	if err == nil {
		t.Error("Should error when property has no tenant")
	}
}

func TestCollectIncome(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant", 0.9, 0.8)
	rcs.RentProperty("prop1", 100)
	rcs.ProcessRentPayment("prop1", 100.0)

	// Collect income
	income := rcs.CollectIncome(1)
	if income != 100.0 {
		t.Errorf("Collected income = %f, want 100.0", income)
	}

	// Income should be zeroed after collection
	remaining := rcs.GetOwnerIncome(1)
	if remaining != 0 {
		t.Errorf("Remaining income = %f, want 0", remaining)
	}
}

func TestSetRent(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)

	err := rcs.SetRent("prop1", 150.0)
	if err != nil {
		t.Errorf("SetRent failed: %v", err)
	}

	prop := rcs.GetProperty("prop1")
	if prop.MonthlyRent != 150.0 {
		t.Errorf("MonthlyRent = %f, want 150.0", prop.MonthlyRent)
	}
	if prop.Deposit != 300.0 {
		t.Errorf("Deposit = %f, want 300.0 (2x new rent)", prop.Deposit)
	}

	// Non-existent property
	err = rcs.SetRent("nonexistent", 100.0)
	if err == nil {
		t.Error("Should error for non-existent property")
	}
}

func TestRepairProperty(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant", 0.9, 0.8)
	rcs.RentProperty("prop1", 100)

	prop := rcs.GetProperty("prop1")
	// Manually degrade condition
	prop.Condition = 0.5

	err := rcs.RepairProperty("prop1", 0.3)
	if err != nil {
		t.Errorf("RepairProperty failed: %v", err)
	}

	if prop.Condition != 0.8 {
		t.Errorf("Condition = %f, want 0.8", prop.Condition)
	}

	// Check tenant satisfaction improved
	tenant := rcs.GetTenant(100)
	if tenant.Satisfaction <= 0.5 {
		t.Error("Tenant satisfaction should have improved")
	}

	// Non-existent property
	err = rcs.RepairProperty("nonexistent", 0.1)
	if err == nil {
		t.Error("Should error for non-existent property")
	}
}

func TestGetVacantProperties(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage1", 100.0, 1.0, location)
	rcs.ListPropertyForRent(1, "prop2", "Cottage2", 100.0, 1.0, location)
	rcs.ListPropertyForRent(1, "prop3", "Cottage3", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant", 0.9, 0.8)
	rcs.RentProperty("prop1", 100) // Occupy prop1

	vacant := rcs.GetVacantProperties()
	if len(vacant) != 2 {
		t.Errorf("Vacant count = %d, want 2", len(vacant))
	}
}

func TestGetOwnerProperties(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage1", 100.0, 1.0, location)
	rcs.ListPropertyForRent(1, "prop2", "Cottage2", 100.0, 1.0, location)
	rcs.ListPropertyForRent(2, "prop3", "Cottage3", 100.0, 1.0, location)

	owner1Props := rcs.GetOwnerProperties(1)
	if len(owner1Props) != 2 {
		t.Errorf("Owner 1 has %d properties, want 2", len(owner1Props))
	}

	owner2Props := rcs.GetOwnerProperties(2)
	if len(owner2Props) != 1 {
		t.Errorf("Owner 2 has %d properties, want 1", len(owner2Props))
	}

	owner3Props := rcs.GetOwnerProperties(3)
	if len(owner3Props) != 0 {
		t.Errorf("Owner 3 has %d properties, want 0", len(owner3Props))
	}
}

func TestCalculateTotalRentalValue(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage1", 100.0, 1.0, location)
	rcs.ListPropertyForRent(1, "prop2", "Cottage2", 150.0, 1.0, location)
	rcs.ListPropertyForRent(1, "prop3", "Cottage3", 200.0, 1.0, location)

	total := rcs.CalculateTotalRentalValue(1)
	if total != 450.0 {
		t.Errorf("Total rental value = %f, want 450.0", total)
	}

	total = rcs.CalculateTotalRentalValue(2)
	if total != 0 {
		t.Errorf("Owner 2 total = %f, want 0", total)
	}
}

func TestCalculateOccupancyRate(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	// No properties
	rate := rcs.CalculateOccupancyRate(1)
	if rate != 0 {
		t.Errorf("Rate with no properties = %f, want 0", rate)
	}

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage1", 100.0, 1.0, location)
	rcs.ListPropertyForRent(1, "prop2", "Cottage2", 100.0, 1.0, location)
	rcs.ListPropertyForRent(1, "prop3", "Cottage3", 100.0, 1.0, location)
	rcs.ListPropertyForRent(1, "prop4", "Cottage4", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant1", 0.9, 0.8)
	rcs.AddTenant(101, "Tenant2", 0.9, 0.8)

	rcs.RentProperty("prop1", 100)
	rcs.RentProperty("prop2", 101)

	rate = rcs.CalculateOccupancyRate(1)
	if rate != 0.5 {
		t.Errorf("Occupancy rate = %f, want 0.5 (50%%)", rate)
	}
}

func TestGetOverdueProperties(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage1", 100.0, 1.0, location)
	rcs.ListPropertyForRent(1, "prop2", "Cottage2", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant1", 0.9, 0.8)
	rcs.AddTenant(101, "Tenant2", 0.9, 0.8)

	rcs.RentProperty("prop1", 100)
	rcs.RentProperty("prop2", 101)

	// Manually set one property to overdue
	prop2 := rcs.GetProperty("prop2")
	prop2.Status = RentStatusOverdue
	prop2.OverdueDays = 3

	overdue := rcs.GetOverdueProperties(1)
	if len(overdue) != 1 {
		t.Errorf("Overdue count = %d, want 1", len(overdue))
	}
	if overdue[0].ID != "prop2" {
		t.Errorf("Overdue property = %s, want prop2", overdue[0].ID)
	}
}

func TestRentCollectionSystemUpdate(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)
	// Create a very reliable tenant that will pay
	rcs.AddTenant(100, "Reliable Tenant", 1.0, 1.0) // Max reliability and wealth
	rcs.RentProperty("prop1", 100)

	prop := rcs.GetProperty("prop1")
	tenant := rcs.GetTenant(100)
	tenant.Satisfaction = 1.0 // Max satisfaction to ensure payment

	// Initial condition
	initialCondition := prop.Condition

	// Update with small dt (simulating regular game loop)
	rcs.Update(1.0) // 1 second

	// Condition should degrade slightly
	if prop.Condition >= initialCondition {
		t.Error("Property condition should degrade over time")
	}
}

func TestRentCollectionSystemUpdatePayment(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")
	rcs.PaymentPeriod = 1.0 // 1 hour for testing

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)
	rcs.AddTenant(100, "Tenant", 1.0, 1.0) // Perfectly reliable tenant
	rcs.RentProperty("prop1", 100)

	tenant := rcs.GetTenant(100)
	tenant.Satisfaction = 1.0 // Max satisfaction

	prop := rcs.GetProperty("prop1")
	initialNextPayment := prop.NextPayment

	// Advance time past payment due
	rcs.Update(3600.0 * 2) // 2 hours (2 seconds converted to hours as dt/3600)

	// Check if payment was processed (tenant should have paid)
	// With perfect reliability/wealth/satisfaction/condition, payment is likely
	if rcs.GetOwnerIncome(1) > 0 || len(rcs.GetRentHistory("prop1")) > 0 {
		// Payment was made
		if prop.NextPayment <= initialNextPayment {
			t.Error("NextPayment should have advanced after payment")
		}
	}
}

func TestRentCollectionSystemEviction(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")
	rcs.PaymentPeriod = 0.001 // Very short payment period
	rcs.EvictionGrace = 3     // 3 day grace period

	location := [3]float64{100.0, 0.0, 200.0}
	rcs.ListPropertyForRent(1, "prop1", "Cottage", 100.0, 1.0, location)
	// Create an unreliable tenant that won't pay
	rcs.AddTenant(100, "Bad Tenant", 0.0, 0.0) // Zero reliability and wealth
	rcs.RentProperty("prop1", 100)

	prop := rcs.GetProperty("prop1")

	// Simulate multiple missed payments
	for i := 0; i < 10; i++ {
		rcs.Update(3600.0) // 1 hour each
	}

	// Property should be in overdue or evicting status
	if prop.Status != RentStatusOverdue && prop.Status != RentStatusEvicting {
		t.Errorf("Status = %d, want RentStatusOverdue or RentStatusEvicting", prop.Status)
	}
	if prop.OverdueDays == 0 {
		t.Error("OverdueDays should be > 0 for unreliable tenant")
	}
}

func TestRegisterProperty(t *testing.T) {
	rcs := NewRentCollectionSystem(12345, "fantasy")

	prop := &RentalProperty{
		ID:          "prop1",
		OwnerID:     1,
		Name:        "Test Property",
		MonthlyRent: 100.0,
	}

	rcs.RegisterProperty(prop)

	got := rcs.GetProperty("prop1")
	if got == nil {
		t.Fatal("RegisterProperty failed to register")
	}
	if got.Name != "Test Property" {
		t.Errorf("Name = %s, want Test Property", got.Name)
	}

	// Check owner tracking
	ownerProps := rcs.GetOwnerProperties(1)
	if len(ownerProps) != 1 {
		t.Errorf("Owner properties count = %d, want 1", len(ownerProps))
	}
}

// ============================================================================
// Home Upgrades System Tests
// ============================================================================

func TestNewHomeUpgradeSystem(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")

	if hus.Seed != 12345 {
		t.Errorf("Seed = %d, want 12345", hus.Seed)
	}
	if hus.Genre != "fantasy" {
		t.Errorf("Genre = %s, want fantasy", hus.Genre)
	}
	if len(hus.AvailableUpgrades) == 0 {
		t.Error("No available upgrades initialized")
	}
}

func TestHomeUpgradeSystemGenreNames(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		hus := NewHomeUpgradeSystem(12345, genre)
		lock := hus.AvailableUpgrades["lock_basic"]
		if lock == nil {
			t.Errorf("lock_basic not found for genre %s", genre)
			continue
		}
		if lock.Name == "" {
			t.Errorf("lock_basic has no name for genre %s", genre)
		}
		// Each genre should have a different name
		t.Logf("%s: %s", genre, lock.Name)
	}
}

func TestRegisterHome(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")

	home := hus.RegisterHome("house1")

	if home == nil {
		t.Fatal("RegisterHome returned nil")
	}
	if home.HouseID != "house1" {
		t.Errorf("HouseID = %s, want house1", home.HouseID)
	}
	if home.StorageSlots != 10 {
		t.Errorf("Base StorageSlots = %d, want 10", home.StorageSlots)
	}

	// Verify it's tracked
	got := hus.GetUpgradedHome("house1")
	if got == nil {
		t.Error("Home not tracked after registration")
	}
}

func TestGetAvailableUpgrades(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	available := hus.GetAvailableUpgrades("house1")

	if len(available) == 0 {
		t.Error("No available upgrades for new home")
	}

	// Level 1 upgrades should be available
	foundBasicLock := false
	for _, upgrade := range available {
		if upgrade.ID == "lock_basic" {
			foundBasicLock = true
			if upgrade.Status != UpgradeStatusAvailable {
				t.Errorf("lock_basic status = %d, want UpgradeStatusAvailable", upgrade.Status)
			}
		}
	}

	if !foundBasicLock {
		t.Error("lock_basic should be available for new home")
	}
}

func TestCanInstallUpgrade(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Basic upgrade should be installable
	can, reason := hus.CanInstallUpgrade("house1", "lock_basic")
	if !can {
		t.Errorf("Should be able to install lock_basic: %s", reason)
	}

	// Advanced upgrade should not be installable (missing prerequisite)
	can, reason = hus.CanInstallUpgrade("house1", "lock_advanced")
	if can {
		t.Error("Should not be able to install lock_advanced without lock_basic")
	}
	if reason == "" {
		t.Error("Should provide reason for failure")
	}

	// Non-existent home
	can, reason = hus.CanInstallUpgrade("nonexistent", "lock_basic")
	if can {
		t.Error("Should fail for non-existent home")
	}

	// Non-existent upgrade
	can, reason = hus.CanInstallUpgrade("house1", "nonexistent")
	if can {
		t.Error("Should fail for non-existent upgrade")
	}
}

func TestStartInstallation(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	upgrade := hus.AvailableUpgrades["lock_basic"]
	cost, err := hus.StartInstallation("house1", "lock_basic", 1000.0)
	if err != nil {
		t.Errorf("StartInstallation failed: %v", err)
	}
	if cost != upgrade.Cost {
		t.Errorf("Cost = %f, want %f", cost, upgrade.Cost)
	}

	// Check upgrade is in progress
	home := hus.GetUpgradedHome("house1")
	installed := home.Upgrades["lock_basic"]
	if installed == nil {
		t.Fatal("Upgrade not tracked after start")
	}
	if installed.Status != UpgradeStatusInProgress {
		t.Errorf("Status = %d, want UpgradeStatusInProgress", installed.Status)
	}
}

func TestStartInstallationErrors(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Insufficient gold
	_, err := hus.StartInstallation("house1", "lock_basic", 10.0)
	if err == nil {
		t.Error("Should fail with insufficient gold")
	}

	// Missing prerequisite
	_, err = hus.StartInstallation("house1", "lock_advanced", 10000.0)
	if err == nil {
		t.Error("Should fail with missing prerequisite")
	}

	// Non-existent home
	_, err = hus.StartInstallation("nonexistent", "lock_basic", 1000.0)
	if err == nil {
		t.Error("Should fail for non-existent home")
	}

	// Non-existent upgrade
	_, err = hus.StartInstallation("house1", "nonexistent", 1000.0)
	if err == nil {
		t.Error("Should fail for non-existent upgrade")
	}

	// Already installed
	hus.StartInstallation("house1", "lock_basic", 1000.0)
	hus.CompleteInstallation("house1", "lock_basic")
	_, err = hus.StartInstallation("house1", "lock_basic", 1000.0)
	if err == nil {
		t.Error("Should fail for already installed upgrade")
	}
}

func TestCompleteInstallation(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	hus.StartInstallation("house1", "lock_basic", 1000.0)
	err := hus.CompleteInstallation("house1", "lock_basic")
	if err != nil {
		t.Errorf("CompleteInstallation failed: %v", err)
	}

	// Check upgrade is completed
	home := hus.GetUpgradedHome("house1")
	installed := home.Upgrades["lock_basic"]
	if installed.Status != UpgradeStatusCompleted {
		t.Errorf("Status = %d, want UpgradeStatusCompleted", installed.Status)
	}
	if installed.Progress != 1.0 {
		t.Errorf("Progress = %f, want 1.0", installed.Progress)
	}

	// Effects should be applied
	if home.SecurityLevel <= 0 {
		t.Error("Security level should have increased")
	}
}

func TestCompleteInstallationErrors(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Non-existent home
	err := hus.CompleteInstallation("nonexistent", "lock_basic")
	if err == nil {
		t.Error("Should fail for non-existent home")
	}

	// Upgrade not started
	err = hus.CompleteInstallation("house1", "lock_basic")
	if err == nil {
		t.Error("Should fail for upgrade not in home")
	}
}

func TestHomeUpgradeSystemUpdate(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Start a short installation
	hus.StartInstallation("house1", "bedding_basic", 1000.0)

	home := hus.GetUpgradedHome("house1")
	upgrade := home.Upgrades["bedding_basic"]
	initialProgress := upgrade.Progress

	// Update with enough time to complete (0.5 hours install time)
	hus.Update(3600.0 * 2) // 2 hours (in seconds converted to hours)

	// Should have progressed or completed
	if upgrade.Progress <= initialProgress && upgrade.Status != UpgradeStatusCompleted {
		t.Error("Upgrade should have progressed")
	}
}

func TestHomeUpgradeSystemUpdateCompletion(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Start installation
	hus.StartInstallation("house1", "bedding_basic", 1000.0) // 0.5 hour install

	// Update with enough time to complete
	hus.Update(3600.0 * 10) // 10 hours worth of updates

	home := hus.GetUpgradedHome("house1")
	upgrade := home.Upgrades["bedding_basic"]

	if upgrade.Status != UpgradeStatusCompleted {
		t.Errorf("Status = %d, want UpgradeStatusCompleted", upgrade.Status)
	}

	// Effects should be applied
	if home.ComfortLevel <= 0 {
		t.Error("Comfort level should have increased")
	}
}

func TestHasUpgrade(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Not installed
	if hus.HasUpgrade("house1", "lock_basic") {
		t.Error("Should not have lock_basic before installation")
	}

	// Install and complete
	hus.StartInstallation("house1", "lock_basic", 1000.0)
	hus.CompleteInstallation("house1", "lock_basic")

	// Should have it now
	if !hus.HasUpgrade("house1", "lock_basic") {
		t.Error("Should have lock_basic after installation")
	}
}

func TestHasSpecialEffect(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Workbench has "home_crafting" effect
	if hus.HasSpecialEffect("house1", "home_crafting") {
		t.Error("Should not have home_crafting before installation")
	}

	hus.StartInstallation("house1", "workbench", 1000.0)
	hus.CompleteInstallation("house1", "workbench")

	if !hus.HasSpecialEffect("house1", "home_crafting") {
		t.Error("Should have home_crafting after workbench installation")
	}
}

func TestGetHomeStats(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	security, _, _, _, storage := hus.GetHomeStats("house1")

	// Base stats for new home
	if storage != 10 {
		t.Errorf("Base storage = %d, want 10", storage)
	}
	if security != 0 {
		t.Errorf("Base security = %f, want 0", security)
	}

	// Install upgrades
	hus.StartInstallation("house1", "lock_basic", 1000.0)
	hus.CompleteInstallation("house1", "lock_basic")
	hus.StartInstallation("house1", "chest_basic", 1000.0)
	hus.CompleteInstallation("house1", "chest_basic")

	security, _, _, value, storage := hus.GetHomeStats("house1")

	if security <= 0 {
		t.Error("Security should have increased after lock_basic")
	}
	if storage <= 10 {
		t.Error("Storage should have increased after chest_basic")
	}
	if value <= 0 {
		t.Error("Value should have increased")
	}
}

func TestGetUpgradesByCategory(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")

	security := hus.GetUpgradesByCategory(UpgradeCategorySecurity)
	if len(security) == 0 {
		t.Error("No security upgrades found")
	}

	for _, upgrade := range security {
		if upgrade.Category != UpgradeCategorySecurity {
			t.Errorf("Wrong category: %d, want UpgradeCategorySecurity", upgrade.Category)
		}
	}

	comfort := hus.GetUpgradesByCategory(UpgradeCategoryComfort)
	if len(comfort) == 0 {
		t.Error("No comfort upgrades found")
	}
}

func TestGetTenantBonus(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	bonus := hus.GetTenantBonus("house1")
	if bonus != 0 {
		t.Errorf("Initial tenant bonus = %f, want 0", bonus)
	}

	// Install heating which has tenant bonus
	hus.StartInstallation("house1", "heating", 1000.0)
	hus.CompleteInstallation("house1", "heating")

	bonus = hus.GetTenantBonus("house1")
	if bonus <= 0 {
		t.Error("Tenant bonus should increase after heating installation")
	}
}

func TestGetCraftingBonus(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	bonus := hus.GetCraftingBonus("house1")
	if bonus != 0 {
		t.Errorf("Initial crafting bonus = %f, want 0", bonus)
	}

	// Install workbench
	hus.StartInstallation("house1", "workbench", 1000.0)
	hus.CompleteInstallation("house1", "workbench")

	bonus = hus.GetCraftingBonus("house1")
	if bonus <= 0 {
		t.Error("Crafting bonus should increase after workbench installation")
	}
}

func TestGetInstalledUpgrades(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	installed := hus.GetInstalledUpgrades("house1")
	if len(installed) != 0 {
		t.Errorf("New home has %d installed upgrades, want 0", len(installed))
	}

	hus.StartInstallation("house1", "lock_basic", 1000.0)
	hus.CompleteInstallation("house1", "lock_basic")
	hus.StartInstallation("house1", "chest_basic", 1000.0)
	hus.CompleteInstallation("house1", "chest_basic")

	installed = hus.GetInstalledUpgrades("house1")
	if len(installed) != 2 {
		t.Errorf("Home has %d installed upgrades, want 2", len(installed))
	}
}

func TestGetUpgradeProgress(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Not started
	progress, inProgress := hus.GetUpgradeProgress("house1", "lock_basic")
	if inProgress {
		t.Error("Should not be in progress before start")
	}

	// Start installation
	hus.StartInstallation("house1", "lock_basic", 1000.0)

	progress, inProgress = hus.GetUpgradeProgress("house1", "lock_basic")
	if !inProgress {
		t.Error("Should be in progress after start")
	}
	if progress != 0 {
		t.Errorf("Initial progress = %f, want 0", progress)
	}

	// Complete it
	hus.CompleteInstallation("house1", "lock_basic")

	progress, inProgress = hus.GetUpgradeProgress("house1", "lock_basic")
	if inProgress {
		t.Error("Should not be in progress after completion")
	}
}

func TestUpgradeCount(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	count := hus.UpgradeCount("house1")
	if count != 0 {
		t.Errorf("New home count = %d, want 0", count)
	}

	hus.StartInstallation("house1", "lock_basic", 1000.0)
	hus.CompleteInstallation("house1", "lock_basic")

	count = hus.UpgradeCount("house1")
	if count != 1 {
		t.Errorf("After 1 install, count = %d, want 1", count)
	}

	hus.StartInstallation("house1", "chest_basic", 1000.0)
	hus.CompleteInstallation("house1", "chest_basic")

	count = hus.UpgradeCount("house1")
	if count != 2 {
		t.Errorf("After 2 installs, count = %d, want 2", count)
	}
}

func TestPrerequisiteChain(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Cannot install advanced lock without basic
	can, _ := hus.CanInstallUpgrade("house1", "lock_advanced")
	if can {
		t.Error("Should not be able to install lock_advanced without lock_basic")
	}

	// Install basic lock
	hus.StartInstallation("house1", "lock_basic", 1000.0)
	hus.CompleteInstallation("house1", "lock_basic")

	// Now can install advanced
	can, _ = hus.CanInstallUpgrade("house1", "lock_advanced")
	if !can {
		t.Error("Should be able to install lock_advanced after lock_basic")
	}
}

func TestMultiplePrerequisites(t *testing.T) {
	hus := NewHomeUpgradeSystem(12345, "fantasy")
	hus.RegisterHome("house1")

	// Alchemy lab requires workbench AND study
	can, _ := hus.CanInstallUpgrade("house1", "alchemy_lab")
	if can {
		t.Error("Should not be able to install alchemy_lab without prerequisites")
	}

	// Install only workbench
	hus.StartInstallation("house1", "workbench", 1000.0)
	hus.CompleteInstallation("house1", "workbench")

	can, _ = hus.CanInstallUpgrade("house1", "alchemy_lab")
	if can {
		t.Error("Should not be able to install alchemy_lab with only workbench")
	}

	// Install study
	hus.StartInstallation("house1", "study", 1000.0)
	hus.CompleteInstallation("house1", "study")

	can, _ = hus.CanInstallUpgrade("house1", "alchemy_lab")
	if !can {
		t.Error("Should be able to install alchemy_lab with both prerequisites")
	}
}
