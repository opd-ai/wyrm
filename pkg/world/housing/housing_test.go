package housing

import (
	"fmt"
	"sync"
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

// ============================================================================
// Guild Hall System Tests
// ============================================================================

func TestNewGuildHallSystem(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	if ghs.Genre != "fantasy" {
		t.Errorf("Genre = %s, want fantasy", ghs.Genre)
	}
	if len(ghs.Permissions) == 0 {
		t.Error("Permissions not initialized")
	}
}

func TestCreateGuildHall(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	hall, err := ghs.CreateGuildHall("guild1", "Test Guild Hall", 1, location)
	if err != nil {
		t.Errorf("CreateGuildHall failed: %v", err)
	}
	if hall == nil {
		t.Fatal("Hall is nil")
	}
	if hall.Name != "Test Guild Hall" {
		t.Errorf("Name = %s, want Test Guild Hall", hall.Name)
	}
	if hall.Tier != GuildHallTierBasic {
		t.Errorf("Tier = %d, want GuildHallTierBasic", hall.Tier)
	}
	if len(hall.Facilities) == 0 {
		t.Error("No basic facilities added")
	}

	// Check leader rank
	rank := ghs.GetMemberRank("guild1", 1)
	if rank != GuildRankLeader {
		t.Errorf("Leader rank = %d, want GuildRankLeader", rank)
	}
}

func TestCreateGuildHallDuplicate(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall 1", 1, location)

	_, err := ghs.CreateGuildHall("guild1", "Hall 2", 2, location)
	if err == nil {
		t.Error("Should fail for duplicate guild hall")
	}
}

func TestGetGuildHall(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	hall := ghs.GetGuildHall("guild1")
	if hall == nil {
		t.Error("GetGuildHall returned nil")
	}

	hall = ghs.GetGuildHall("nonexistent")
	if hall != nil {
		t.Error("GetGuildHall should return nil for nonexistent guild")
	}
}

func TestGuildHallUpgrade(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	// Deposit funds for upgrade
	ghs.DepositToTreasury("guild1", 1, 10000)

	err := ghs.UpgradeGuildHall("guild1", 1)
	if err != nil {
		t.Errorf("UpgradeGuildHall failed: %v", err)
	}

	hall := ghs.GetGuildHall("guild1")
	if hall.Tier != GuildHallTierStandard {
		t.Errorf("Tier = %d, want GuildHallTierStandard", hall.Tier)
	}

	// Should have new facility (training_ground)
	if _, ok := hall.Facilities["training_ground"]; !ok {
		t.Error("Training ground should be added at Standard tier")
	}
}

func TestGuildHallUpgradeErrors(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankMember)

	// Non-leader cannot upgrade
	err := ghs.UpgradeGuildHall("guild1", 2)
	if err == nil {
		t.Error("Non-leader should not be able to upgrade")
	}

	// Insufficient funds
	err = ghs.UpgradeGuildHall("guild1", 1)
	if err == nil {
		t.Error("Should fail with insufficient funds")
	}
}

func TestGuildMemberRanks(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	ghs.AddMemberRank("guild1", 2, GuildRankMember)
	ghs.AddMemberRank("guild1", 3, GuildRankOfficer)

	rank := ghs.GetMemberRank("guild1", 2)
	if rank != GuildRankMember {
		t.Errorf("Rank = %d, want GuildRankMember", rank)
	}

	rank = ghs.GetMemberRank("guild1", 3)
	if rank != GuildRankOfficer {
		t.Errorf("Rank = %d, want GuildRankOfficer", rank)
	}
}

func TestPromoteMember(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankMember)

	err := ghs.PromoteMember("guild1", 1, 2) // Leader promotes member
	if err != nil {
		t.Errorf("PromoteMember failed: %v", err)
	}

	rank := ghs.GetMemberRank("guild1", 2)
	if rank != GuildRankOfficer {
		t.Errorf("Rank after promotion = %d, want GuildRankOfficer", rank)
	}
}

func TestPromoteMemberErrors(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankMember)
	ghs.AddMemberRank("guild1", 3, GuildRankOfficer)

	// Officer cannot promote (doesn't have manage_ranks)
	err := ghs.PromoteMember("guild1", 3, 2)
	if err == nil {
		t.Error("Officer should not be able to promote")
	}

	// Cannot promote equal rank
	ghs.AddMemberRank("guild1", 4, GuildRankCouncil)
	err = ghs.PromoteMember("guild1", 1, 1) // Leader promoting self
	if err == nil {
		t.Error("Cannot promote equal or higher rank")
	}
}

func TestDemoteMember(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankOfficer)

	err := ghs.DemoteMember("guild1", 1, 2)
	if err != nil {
		t.Errorf("DemoteMember failed: %v", err)
	}

	rank := ghs.GetMemberRank("guild1", 2)
	if rank != GuildRankMember {
		t.Errorf("Rank after demotion = %d, want GuildRankMember", rank)
	}
}

func TestHasPermission(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankMember)
	ghs.AddMemberRank("guild1", 3, GuildRankOfficer)

	// Members can use facilities
	if !ghs.HasPermission("guild1", 2, "use_facilities") {
		t.Error("Members should be able to use facilities")
	}

	// Members cannot access treasury
	if ghs.HasPermission("guild1", 2, "access_treasury") {
		t.Error("Members should not access treasury")
	}

	// Officers can invite
	if !ghs.HasPermission("guild1", 3, "invite_members") {
		t.Error("Officers should be able to invite")
	}

	// Leaders can manage ranks
	if !ghs.HasPermission("guild1", 1, "manage_ranks") {
		t.Error("Leaders should be able to manage ranks")
	}
}

func TestDepositWithdrawTreasury(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankCouncil)

	// Deposit
	err := ghs.DepositToTreasury("guild1", 2, 1000)
	if err != nil {
		t.Errorf("Deposit failed: %v", err)
	}

	balance := ghs.GetTreasuryBalance("guild1")
	if balance != 1000 {
		t.Errorf("Balance = %f, want 1000", balance)
	}

	// Withdraw
	amount, err := ghs.WithdrawFromTreasury("guild1", 2, 500)
	if err != nil {
		t.Errorf("Withdraw failed: %v", err)
	}
	if amount != 500 {
		t.Errorf("Withdrawn = %f, want 500", amount)
	}

	balance = ghs.GetTreasuryBalance("guild1")
	if balance != 500 {
		t.Errorf("Balance after withdraw = %f, want 500", balance)
	}
}

func TestTreasuryErrors(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankMember)

	ghs.DepositToTreasury("guild1", 1, 100)

	// Member cannot withdraw
	_, err := ghs.WithdrawFromTreasury("guild1", 2, 50)
	if err == nil {
		t.Error("Member should not be able to withdraw")
	}

	// Cannot withdraw more than balance
	_, err = ghs.WithdrawFromTreasury("guild1", 1, 1000)
	if err == nil {
		t.Error("Cannot withdraw more than balance")
	}
}

func TestGuildBank(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankOfficer)

	// Deposit item
	err := ghs.DepositToBank("guild1", 2, "item1")
	if err != nil {
		t.Errorf("Deposit to bank failed: %v", err)
	}

	contents, _ := ghs.GetBankContents("guild1", 2)
	if len(contents) != 1 {
		t.Errorf("Bank has %d items, want 1", len(contents))
	}

	// Withdraw item
	err = ghs.WithdrawFromBank("guild1", 2, "item1")
	if err != nil {
		t.Errorf("Withdraw from bank failed: %v", err)
	}

	contents, _ = ghs.GetBankContents("guild1", 2)
	if len(contents) != 0 {
		t.Errorf("Bank has %d items after withdraw, want 0", len(contents))
	}
}

func TestGuildBankErrors(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankMember)

	// Member cannot deposit
	err := ghs.DepositToBank("guild1", 2, "item1")
	if err == nil {
		t.Error("Member should not be able to deposit to guild bank")
	}

	// Withdraw non-existent item
	ghs.AddMemberRank("guild1", 3, GuildRankOfficer)
	err = ghs.WithdrawFromBank("guild1", 3, "nonexistent")
	if err == nil {
		t.Error("Cannot withdraw non-existent item")
	}
}

func TestUseFacility(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankMember)

	err := ghs.UseFacility("guild1", 2, "meeting_hall")
	if err != nil {
		t.Errorf("UseFacility failed: %v", err)
	}

	facility := ghs.GetFacility("guild1", "meeting_hall")
	if facility.CurrentUse != 1 {
		t.Errorf("CurrentUse = %d, want 1", facility.CurrentUse)
	}

	ghs.ReleaseFacility("guild1", "meeting_hall")
	if facility.CurrentUse != 0 {
		t.Errorf("CurrentUse after release = %d, want 0", facility.CurrentUse)
	}
}

func TestUseFacilityErrors(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	// Non-member cannot use
	err := ghs.UseFacility("guild1", 999, "meeting_hall")
	if err == nil {
		t.Error("Non-member should not use facility")
	}

	// Non-existent facility
	ghs.AddMemberRank("guild1", 2, GuildRankMember)
	err = ghs.UseFacility("guild1", 2, "nonexistent")
	if err == nil {
		t.Error("Cannot use non-existent facility")
	}
}

func TestUpgradeFacility(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankCouncil)
	ghs.DepositToTreasury("guild1", 1, 1000)

	facility := ghs.GetFacility("guild1", "meeting_hall")
	initialLevel := facility.Level

	err := ghs.UpgradeFacility("guild1", 2, "meeting_hall")
	if err != nil {
		t.Errorf("UpgradeFacility failed: %v", err)
	}

	if facility.Level != initialLevel+1 {
		t.Errorf("Level = %d, want %d", facility.Level, initialLevel+1)
	}
}

func TestGetAllFacilities(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	facilities := ghs.GetAllFacilities("guild1")
	if len(facilities) < 2 {
		t.Errorf("Got %d facilities, want at least 2 (basic)", len(facilities))
	}
}

func TestGuildHallUpdate(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	hall := ghs.GetGuildHall("guild1")
	initialLastUpkeep := hall.LastUpkeep

	// Update less than a day - no upkeep
	ghs.Update(3600.0 * 12) // 12 hours

	if hall.LastUpkeep != initialLastUpkeep {
		t.Error("Upkeep should not be processed before a full day")
	}
}

func TestGuildHallDebt(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	hall := ghs.GetGuildHall("guild1")
	// Ensure no funds
	hall.TreasuryGold = 0

	// Process a day
	ghs.Update(3600.0 * 24)

	if !hall.InDebt {
		t.Error("Hall should be in debt with no funds")
	}
	if hall.DebtDays != 1 {
		t.Errorf("DebtDays = %d, want 1", hall.DebtDays)
	}
}

func TestGuildHallDebtFacilitiesDisabled(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	hall := ghs.GetGuildHall("guild1")
	hall.TreasuryGold = 0

	// 3 days in debt
	for i := 0; i < 3; i++ {
		ghs.Update(3600.0 * 24)
	}

	// Facilities should be disabled
	for _, facility := range hall.Facilities {
		if facility.Operational {
			t.Errorf("Facility %s should be non-operational after 3 days debt", facility.ID)
		}
	}
}

func TestGetGuildHallStats(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.DepositToTreasury("guild1", 1, 500)
	ghs.DepositToBank("guild1", 1, "item1")

	tier, treasury, bankUsed, bankCap, _ := ghs.GetGuildHallStats("guild1")

	if tier != GuildHallTierBasic {
		t.Errorf("Tier = %d, want GuildHallTierBasic", tier)
	}
	if treasury != 500 {
		t.Errorf("Treasury = %f, want 500", treasury)
	}
	if bankUsed != 1 {
		t.Errorf("BankUsed = %d, want 1", bankUsed)
	}
	if bankCap != 50 {
		t.Errorf("BankCap = %d, want 50", bankCap)
	}
}

func TestGetTotalBonuses(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	bonuses := ghs.GetTotalBonuses("guild1")

	// Basic hall has social and storage bonuses
	if bonuses.SocialBonus <= 0 {
		t.Error("Should have social bonus from meeting hall")
	}
	if bonuses.StorageBonus <= 0 {
		t.Error("Should have storage bonus from storage room")
	}
}

func TestDisbandGuildHall(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.DepositToTreasury("guild1", 1, 1000)

	remaining, err := ghs.DisbandGuildHall("guild1", 1)
	if err != nil {
		t.Errorf("DisbandGuildHall failed: %v", err)
	}
	if remaining != 1000 {
		t.Errorf("Remaining treasury = %f, want 1000", remaining)
	}

	// Hall should be gone
	if ghs.GetGuildHall("guild1") != nil {
		t.Error("Hall should be removed after disband")
	}
}

func TestDisbandGuildHallNonLeader(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankOfficer)

	_, err := ghs.DisbandGuildHall("guild1", 2)
	if err == nil {
		t.Error("Non-leader should not be able to disband")
	}
}

func TestGetMemberCount(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	count := ghs.GetMemberCount("guild1")
	if count != 1 {
		t.Errorf("Initial count = %d, want 1 (leader)", count)
	}

	ghs.AddMemberRank("guild1", 2, GuildRankMember)
	ghs.AddMemberRank("guild1", 3, GuildRankOfficer)

	count = ghs.GetMemberCount("guild1")
	if count != 3 {
		t.Errorf("Count after adds = %d, want 3", count)
	}
}

func TestRemoveMember(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)
	ghs.AddMemberRank("guild1", 2, GuildRankMember)

	ghs.RemoveMember("guild1", 2)

	rank := ghs.GetMemberRank("guild1", 2)
	if rank != GuildRankMember {
		// After removal, getting rank of non-member returns default (Member)
		// which is fine since they're no longer tracked
	}

	count := ghs.GetMemberCount("guild1")
	if count != 1 {
		t.Errorf("Count after remove = %d, want 1", count)
	}
}

func TestIsInDebt(t *testing.T) {
	ghs := NewGuildHallSystem(12345, "fantasy")

	location := [3]float64{100.0, 0.0, 200.0}
	ghs.CreateGuildHall("guild1", "Hall", 1, location)

	if ghs.IsInDebt("guild1") {
		t.Error("New hall should not be in debt")
	}

	hall := ghs.GetGuildHall("guild1")
	hall.InDebt = true

	if !ghs.IsInDebt("guild1") {
		t.Error("Hall marked as in debt should return true")
	}
}

func TestGenreFacilityNames(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		ghs := NewGuildHallSystem(12345, genre)
		location := [3]float64{100.0, 0.0, 200.0}
		ghs.CreateGuildHall("guild1", "Hall", 1, location)

		hall := ghs.GetGuildHall("guild1")
		meetingHall := hall.Facilities["meeting_hall"]
		if meetingHall == nil {
			t.Errorf("No meeting hall for genre %s", genre)
			continue
		}
		if meetingHall.Name == "" {
			t.Errorf("Empty meeting hall name for genre %s", genre)
		}
		t.Logf("%s: meeting hall = %s", genre, meetingHall.Name)
	}
}

// ============================================================================
// Shared Storage System Tests
// ============================================================================

func TestNewSharedStorageSystem(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")

	if sys == nil {
		t.Fatal("NewSharedStorageSystem returned nil")
	}
	if sys.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", sys.Seed)
	}
	if sys.Genre != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got %s", sys.Genre)
	}
	if sys.Storages == nil {
		t.Error("Storages map not initialized")
	}
	if sys.PlayerStorage == nil {
		t.Error("PlayerStorage map not initialized")
	}
	if sys.StorageCount() != 0 {
		t.Errorf("Expected 0 storages, got %d", sys.StorageCount())
	}
}

func TestCreateStorage(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")

	// Create storage
	storage, err := sys.CreateStorage("storage1", "Guild Vault", 100, "guild", 50)
	if err != nil {
		t.Fatalf("CreateStorage failed: %v", err)
	}

	if storage.ID != "storage1" {
		t.Errorf("Expected ID 'storage1', got %s", storage.ID)
	}
	if storage.Name != "Guild Vault" {
		t.Errorf("Expected name 'Guild Vault', got %s", storage.Name)
	}
	if storage.OwnerID != 100 {
		t.Errorf("Expected owner 100, got %d", storage.OwnerID)
	}
	if storage.OwnerType != "guild" {
		t.Errorf("Expected ownerType 'guild', got %s", storage.OwnerType)
	}
	if storage.Capacity != 50 {
		t.Errorf("Expected capacity 50, got %d", storage.Capacity)
	}

	// Owner should have full permission
	perm := sys.GetPermission("storage1", 100)
	if perm != StoragePermissionFull {
		t.Errorf("Expected owner to have full permission, got %v", perm)
	}

	// Storage count
	if sys.StorageCount() != 1 {
		t.Errorf("Expected 1 storage, got %d", sys.StorageCount())
	}

	// Duplicate should fail
	_, err = sys.CreateStorage("storage1", "Another", 200, "player", 10)
	if err == nil {
		t.Error("Expected error for duplicate storage ID")
	}
}

func TestGetStorage(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Test Storage", 100, "player", 20)

	storage := sys.GetStorage("storage1")
	if storage == nil {
		t.Fatal("GetStorage returned nil for existing storage")
	}
	if storage.Name != "Test Storage" {
		t.Errorf("Expected name 'Test Storage', got %s", storage.Name)
	}

	// Non-existent storage
	storage = sys.GetStorage("nonexistent")
	if storage != nil {
		t.Error("GetStorage should return nil for non-existent storage")
	}
}

func TestAddRemoveMember(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Shared Chest", 100, "player", 30)

	// Add member with deposit permission
	err := sys.AddMember("storage1", 100, 200, StoragePermissionDeposit)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	perm := sys.GetPermission("storage1", 200)
	if perm != StoragePermissionDeposit {
		t.Errorf("Expected deposit permission, got %v", perm)
	}

	// Non-owner cannot add members
	err = sys.AddMember("storage1", 200, 300, StoragePermissionView)
	if err == nil {
		t.Error("Non-owner should not be able to add members")
	}

	// Member count
	if sys.MemberCount("storage1") != 2 {
		t.Errorf("Expected 2 members, got %d", sys.MemberCount("storage1"))
	}

	// Remove member
	err = sys.RemoveMember("storage1", 100, 200)
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}

	perm = sys.GetPermission("storage1", 200)
	if perm != StoragePermissionNone {
		t.Errorf("Removed member should have no permission, got %v", perm)
	}

	// Cannot remove owner
	err = sys.RemoveMember("storage1", 100, 100)
	if err == nil {
		t.Error("Should not be able to remove owner")
	}

	// Non-owner cannot remove members
	sys.AddMember("storage1", 100, 300, StoragePermissionView)
	err = sys.RemoveMember("storage1", 300, 300)
	if err == nil {
		t.Error("Non-owner should not be able to remove members")
	}
}

func TestSetPermission(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Test", 100, "player", 20)
	sys.AddMember("storage1", 100, 200, StoragePermissionView)

	// Owner changes permission
	err := sys.SetPermission("storage1", 100, 200, StoragePermissionWithdraw)
	if err != nil {
		t.Fatalf("SetPermission failed: %v", err)
	}

	perm := sys.GetPermission("storage1", 200)
	if perm != StoragePermissionWithdraw {
		t.Errorf("Expected withdraw permission, got %v", perm)
	}

	// Non-owner cannot change permission
	err = sys.SetPermission("storage1", 200, 200, StoragePermissionFull)
	if err == nil {
		t.Error("Non-owner should not be able to change permissions")
	}

	// Cannot set permission for non-member
	err = sys.SetPermission("storage1", 100, 999, StoragePermissionView)
	if err == nil {
		t.Error("Should not be able to set permission for non-member")
	}

	// Non-existent storage
	err = sys.SetPermission("nonexistent", 100, 200, StoragePermissionFull)
	if err == nil {
		t.Error("Should error for non-existent storage")
	}
}

func TestDepositWithdrawItem(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Vault", 100, "player", 10)
	sys.AddMember("storage1", 100, 200, StoragePermissionDeposit)
	sys.AddMember("storage1", 100, 300, StoragePermissionWithdraw)

	// Member with deposit permission can deposit
	item := &StorageItem{
		ID:       "item1",
		Name:     "Golden Sword",
		Type:     "weapon",
		Quantity: 1,
	}
	err := sys.DepositItem("storage1", 200, item)
	if err != nil {
		t.Fatalf("DepositItem failed: %v", err)
	}

	if sys.ItemCount("storage1") != 1 {
		t.Errorf("Expected 1 item, got %d", sys.ItemCount("storage1"))
	}

	// Deposited item should have owner set
	if item.OwnerID != 200 {
		t.Errorf("Expected item owner 200, got %d", item.OwnerID)
	}

	// Owner can withdraw
	withdrawn, err := sys.WithdrawItem("storage1", 100, "item1")
	if err != nil {
		t.Fatalf("WithdrawItem failed: %v", err)
	}
	if withdrawn.Name != "Golden Sword" {
		t.Errorf("Expected 'Golden Sword', got %s", withdrawn.Name)
	}

	if sys.ItemCount("storage1") != 0 {
		t.Errorf("Expected 0 items after withdrawal, got %d", sys.ItemCount("storage1"))
	}

	// Deposit-only member cannot withdraw
	item2 := &StorageItem{ID: "item2", Name: "Potion", Type: "consumable", Quantity: 5}
	sys.DepositItem("storage1", 200, item2)
	_, err = sys.WithdrawItem("storage1", 200, "item2")
	if err == nil {
		t.Error("Deposit-only member should not be able to withdraw")
	}

	// View-only member cannot deposit
	sys.AddMember("storage1", 100, 400, StoragePermissionView)
	item3 := &StorageItem{ID: "item3", Name: "Ring", Type: "accessory", Quantity: 1}
	err = sys.DepositItem("storage1", 400, item3)
	if err == nil {
		t.Error("View-only member should not be able to deposit")
	}
}

func TestStorageCapacity(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Small Chest", 100, "player", 3)

	// Fill storage
	for i := 1; i <= 3; i++ {
		item := &StorageItem{ID: fmt.Sprintf("item%d", i), Name: "Item", Quantity: 1}
		err := sys.DepositItem("storage1", 100, item)
		if err != nil {
			t.Fatalf("DepositItem %d failed: %v", i, err)
		}
	}

	// Should be full
	item := &StorageItem{ID: "item4", Name: "Extra", Quantity: 1}
	err := sys.DepositItem("storage1", 100, item)
	if err == nil {
		t.Error("Should not be able to deposit when full")
	}

	// Check capacity
	used, capacity := sys.GetStorageCapacity("storage1")
	if used != 3 {
		t.Errorf("Expected used 3, got %d", used)
	}
	if capacity != 3 {
		t.Errorf("Expected capacity 3, got %d", capacity)
	}

	// Expand capacity
	err = sys.SetCapacity("storage1", 100, 5)
	if err != nil {
		t.Fatalf("SetCapacity failed: %v", err)
	}

	// Now can deposit
	err = sys.DepositItem("storage1", 100, item)
	if err != nil {
		t.Fatalf("DepositItem after expansion failed: %v", err)
	}

	// Cannot shrink below usage
	err = sys.SetCapacity("storage1", 100, 2)
	if err == nil {
		t.Error("Should not be able to shrink below current usage")
	}
}

func TestReserveUnreserveItem(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Vault", 100, "player", 20)
	sys.AddMember("storage1", 100, 200, StoragePermissionWithdraw)
	sys.AddMember("storage1", 100, 300, StoragePermissionWithdraw)

	// Deposit item
	item := &StorageItem{ID: "item1", Name: "Legendary Sword", Type: "weapon", Quantity: 1}
	sys.DepositItem("storage1", 100, item)

	// Owner reserves for member 200
	err := sys.ReserveItem("storage1", 100, "item1", 200)
	if err != nil {
		t.Fatalf("ReserveItem failed: %v", err)
	}

	// Check reservation
	storage := sys.GetStorage("storage1")
	if !storage.Items["item1"].Reserved {
		t.Error("Item should be reserved")
	}
	if storage.Items["item1"].ReservedFor != 200 {
		t.Errorf("Item should be reserved for 200, got %d", storage.Items["item1"].ReservedFor)
	}

	// Member 300 cannot withdraw reserved item
	_, err = sys.WithdrawItem("storage1", 300, "item1")
	if err == nil {
		t.Error("Should not be able to withdraw item reserved for another")
	}

	// Member 200 can withdraw reserved item
	withdrawn, err := sys.WithdrawItem("storage1", 200, "item1")
	if err != nil {
		t.Fatalf("Reserved member should be able to withdraw: %v", err)
	}
	if withdrawn.ID != "item1" {
		t.Error("Wrong item withdrawn")
	}

	// Deposit new item for unreserve test
	item2 := &StorageItem{ID: "item2", Name: "Shield", Type: "armor", Quantity: 1}
	sys.DepositItem("storage1", 100, item2)
	sys.ReserveItem("storage1", 100, "item2", 300)

	// Cannot reserve already reserved item
	err = sys.ReserveItem("storage1", 100, "item2", 200)
	if err == nil {
		t.Error("Should not be able to reserve already reserved item")
	}

	// Unreserve
	err = sys.UnreserveItem("storage1", 100, "item2")
	if err != nil {
		t.Fatalf("UnreserveItem failed: %v", err)
	}

	storage = sys.GetStorage("storage1")
	if storage.Items["item2"].Reserved {
		t.Error("Item should no longer be reserved")
	}

	// Cannot unreserve non-reserved item
	err = sys.UnreserveItem("storage1", 100, "item2")
	if err == nil {
		t.Error("Should not be able to unreserve non-reserved item")
	}

	// Non-full permission cannot reserve
	err = sys.ReserveItem("storage1", 200, "item2", 200)
	if err == nil {
		t.Error("Withdraw-only member should not be able to reserve")
	}
}

func TestGetItems(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Vault", 100, "player", 20)
	sys.AddMember("storage1", 100, 200, StoragePermissionView)
	sys.AddMember("storage1", 100, 300, StoragePermissionNone)

	// Add items
	for i := 1; i <= 3; i++ {
		item := &StorageItem{ID: fmt.Sprintf("item%d", i), Name: fmt.Sprintf("Item %d", i), Quantity: i}
		sys.DepositItem("storage1", 100, item)
	}

	// View permission can see items
	items, err := sys.GetItems("storage1", 200)
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// No permission cannot see items
	_, err = sys.GetItems("storage1", 300)
	if err == nil {
		t.Error("No-permission member should not be able to see items")
	}

	// Non-member cannot see items
	_, err = sys.GetItems("storage1", 999)
	if err == nil {
		t.Error("Non-member should not be able to see items")
	}
}

func TestAccessLog(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Vault", 100, "player", 20)

	// Deposit and withdraw to generate logs
	item := &StorageItem{ID: "item1", Name: "Test", Quantity: 1}
	sys.DepositItem("storage1", 100, item)
	sys.WithdrawItem("storage1", 100, "item1")

	// Owner can view log
	log, err := sys.GetAccessLog("storage1", 100)
	if err != nil {
		t.Fatalf("GetAccessLog failed: %v", err)
	}
	if len(log) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(log))
	}

	// Check log entries
	if log[0].Action != "deposit" {
		t.Errorf("Expected first action 'deposit', got %s", log[0].Action)
	}
	if log[1].Action != "withdraw" {
		t.Errorf("Expected second action 'withdraw', got %s", log[1].Action)
	}

	// Non-full permission cannot view log
	sys.AddMember("storage1", 100, 200, StoragePermissionWithdraw)
	_, err = sys.GetAccessLog("storage1", 200)
	if err == nil {
		t.Error("Withdraw-only member should not be able to view log")
	}
}

func TestAccessLogTrimming(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	storage, _ := sys.CreateStorage("storage1", "Vault", 100, "player", 200)

	// Override max entries for test
	storage.MaxLogEntries = 5

	// Generate more logs than max
	for i := 1; i <= 10; i++ {
		item := &StorageItem{ID: fmt.Sprintf("item%d", i), Name: "Test", Quantity: 1}
		sys.DepositItem("storage1", 100, item)
	}

	log, _ := sys.GetAccessLog("storage1", 100)
	if len(log) != 5 {
		t.Errorf("Expected log trimmed to 5 entries, got %d", len(log))
	}

	// Should have the most recent entries
	if log[0].ItemID != "item6" {
		t.Errorf("Expected oldest entry to be item6, got %s", log[0].ItemID)
	}
}

func TestGetMembers(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Vault", 100, "player", 20)
	sys.AddMember("storage1", 100, 200, StoragePermissionView)
	sys.AddMember("storage1", 100, 300, StoragePermissionDeposit)

	// Member with view permission can see members
	members, err := sys.GetMembers("storage1", 200)
	if err != nil {
		t.Fatalf("GetMembers failed: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}

	// Check permissions
	if members[100] != StoragePermissionFull {
		t.Error("Owner should have full permission")
	}
	if members[200] != StoragePermissionView {
		t.Error("Member 200 should have view permission")
	}
	if members[300] != StoragePermissionDeposit {
		t.Error("Member 300 should have deposit permission")
	}

	// No permission cannot see members
	sys.AddMember("storage1", 100, 400, StoragePermissionNone)
	_, err = sys.GetMembers("storage1", 400)
	if err == nil {
		t.Error("No-permission member should not be able to see members")
	}
}

func TestDeleteStorage(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Vault", 100, "player", 20)
	sys.AddMember("storage1", 100, 200, StoragePermissionFull)

	// Cannot delete with items
	item := &StorageItem{ID: "item1", Name: "Test", Quantity: 1}
	sys.DepositItem("storage1", 100, item)

	err := sys.DeleteStorage("storage1", 100)
	if err == nil {
		t.Error("Should not be able to delete storage with items")
	}

	// Remove item
	sys.WithdrawItem("storage1", 100, "item1")

	// Non-owner cannot delete
	err = sys.DeleteStorage("storage1", 200)
	if err == nil {
		t.Error("Non-owner should not be able to delete storage")
	}

	// Owner can delete
	err = sys.DeleteStorage("storage1", 100)
	if err != nil {
		t.Fatalf("DeleteStorage failed: %v", err)
	}

	if sys.StorageCount() != 0 {
		t.Error("Storage should be deleted")
	}

	// Check removed from player storages
	storages := sys.GetPlayerStorages(100)
	if len(storages) != 0 {
		t.Error("Storage should be removed from player's list")
	}

	storages = sys.GetPlayerStorages(200)
	if len(storages) != 0 {
		t.Error("Storage should be removed from member's list")
	}
}

func TestSharedStorageTransferOwnership(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Vault", 100, "player", 20)
	sys.AddMember("storage1", 100, 200, StoragePermissionDeposit)

	// Non-owner cannot transfer
	err := sys.TransferOwnership("storage1", 200, 300)
	if err == nil {
		t.Error("Non-owner should not be able to transfer ownership")
	}

	// Owner transfers to member
	err = sys.TransferOwnership("storage1", 100, 200)
	if err != nil {
		t.Fatalf("TransferOwnership failed: %v", err)
	}

	// Check new owner
	storage := sys.GetStorage("storage1")
	if storage.OwnerID != 200 {
		t.Errorf("Expected new owner 200, got %d", storage.OwnerID)
	}

	// New owner has full permission
	perm := sys.GetPermission("storage1", 200)
	if perm != StoragePermissionFull {
		t.Error("New owner should have full permission")
	}

	// New owner should be in player storage list
	storages := sys.GetPlayerStorages(200)
	found := false
	for _, s := range storages {
		if s.ID == "storage1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Storage should be in new owner's list")
	}

	// Transfer to non-member (should work and add them)
	err = sys.TransferOwnership("storage1", 200, 300)
	if err != nil {
		t.Fatalf("Transfer to non-member failed: %v", err)
	}
	if storage.OwnerID != 300 {
		t.Errorf("Expected owner 300, got %d", storage.OwnerID)
	}
}

func TestGetPlayerStorages(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")

	// Create multiple storages for player
	sys.CreateStorage("storage1", "Vault 1", 100, "player", 20)
	sys.CreateStorage("storage2", "Vault 2", 100, "player", 20)
	sys.CreateStorage("storage3", "Vault 3", 200, "player", 20)

	// Add player 100 to storage3
	sys.AddMember("storage3", 200, 100, StoragePermissionView)

	// Player 100 should have access to all three
	storages := sys.GetPlayerStorages(100)
	if len(storages) != 3 {
		t.Errorf("Expected 3 storages, got %d", len(storages))
	}

	// Player 200 should only have storage3
	storages = sys.GetPlayerStorages(200)
	if len(storages) != 1 {
		t.Errorf("Expected 1 storage for player 200, got %d", len(storages))
	}
}

func TestStorageUpdate(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")

	if sys.GameTime != 0 {
		t.Errorf("Initial game time should be 0, got %f", sys.GameTime)
	}

	sys.Update(1.5)
	if sys.GameTime != 1.5 {
		t.Errorf("Expected game time 1.5, got %f", sys.GameTime)
	}

	sys.Update(0.5)
	if sys.GameTime != 2.0 {
		t.Errorf("Expected game time 2.0, got %f", sys.GameTime)
	}
}

func TestSharedStorageNonExistent(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")

	// All operations on non-existent storage should error
	err := sys.AddMember("nonexistent", 100, 200, StoragePermissionView)
	if err == nil {
		t.Error("AddMember on non-existent storage should error")
	}

	err = sys.RemoveMember("nonexistent", 100, 200)
	if err == nil {
		t.Error("RemoveMember on non-existent storage should error")
	}

	err = sys.DepositItem("nonexistent", 100, &StorageItem{ID: "item1"})
	if err == nil {
		t.Error("DepositItem on non-existent storage should error")
	}

	_, err = sys.WithdrawItem("nonexistent", 100, "item1")
	if err == nil {
		t.Error("WithdrawItem on non-existent storage should error")
	}

	err = sys.ReserveItem("nonexistent", 100, "item1", 200)
	if err == nil {
		t.Error("ReserveItem on non-existent storage should error")
	}

	err = sys.UnreserveItem("nonexistent", 100, "item1")
	if err == nil {
		t.Error("UnreserveItem on non-existent storage should error")
	}

	_, err = sys.GetItems("nonexistent", 100)
	if err == nil {
		t.Error("GetItems on non-existent storage should error")
	}

	_, err = sys.GetAccessLog("nonexistent", 100)
	if err == nil {
		t.Error("GetAccessLog on non-existent storage should error")
	}

	_, err = sys.GetMembers("nonexistent", 100)
	if err == nil {
		t.Error("GetMembers on non-existent storage should error")
	}

	err = sys.SetCapacity("nonexistent", 100, 50)
	if err == nil {
		t.Error("SetCapacity on non-existent storage should error")
	}

	err = sys.DeleteStorage("nonexistent", 100)
	if err == nil {
		t.Error("DeleteStorage on non-existent storage should error")
	}

	err = sys.TransferOwnership("nonexistent", 100, 200)
	if err == nil {
		t.Error("TransferOwnership on non-existent storage should error")
	}

	// These should return safe defaults
	perm := sys.GetPermission("nonexistent", 100)
	if perm != StoragePermissionNone {
		t.Error("GetPermission on non-existent storage should return None")
	}

	used, cap := sys.GetStorageCapacity("nonexistent")
	if used != 0 || cap != 0 {
		t.Error("GetStorageCapacity on non-existent storage should return 0,0")
	}

	count := sys.MemberCount("nonexistent")
	if count != 0 {
		t.Error("MemberCount on non-existent storage should return 0")
	}

	count = sys.ItemCount("nonexistent")
	if count != 0 {
		t.Error("ItemCount on non-existent storage should return 0")
	}
}

func TestSharedStorageConcurrency(t *testing.T) {
	sys := NewSharedStorageSystem(12345, "fantasy")
	sys.CreateStorage("storage1", "Test", 100, "player", 1000)

	// Add members
	for i := uint64(101); i <= 110; i++ {
		sys.AddMember("storage1", 100, i, StoragePermissionFull)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent deposits
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			memberID := uint64(100 + (idx % 11))
			item := &StorageItem{
				ID:       fmt.Sprintf("item_%d", idx),
				Name:     fmt.Sprintf("Item %d", idx),
				Quantity: 1,
			}
			err := sys.DepositItem("storage1", memberID, item)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errCount := 0
	for err := range errors {
		t.Logf("Error during concurrent deposit: %v", err)
		errCount++
	}

	if errCount > 0 {
		t.Errorf("Had %d errors during concurrent deposits", errCount)
	}

	// Verify items were deposited
	if sys.ItemCount("storage1") != 50 {
		t.Errorf("Expected 50 items, got %d", sys.ItemCount("storage1"))
	}
}

func TestSharedStorageAllGenres(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		sys := NewSharedStorageSystem(12345, genre)
		if sys.Genre != genre {
			t.Errorf("Expected genre %s, got %s", genre, sys.Genre)
		}

		// Create and use storage
		_, err := sys.CreateStorage("vault", "Test Vault", 100, "player", 10)
		if err != nil {
			t.Errorf("Failed to create storage for genre %s: %v", genre, err)
		}

		item := &StorageItem{ID: "item1", Name: "Test", Quantity: 1}
		err = sys.DepositItem("vault", 100, item)
		if err != nil {
			t.Errorf("Failed to deposit for genre %s: %v", genre, err)
		}
	}
}

// ============================================================================
// Indoor/Outdoor Detection Tests
// ============================================================================

func TestNewIndoorDetectionSystem(t *testing.T) {
	hm := NewHouseManager()
	ids := NewIndoorDetectionSystem(hm)
	if ids == nil {
		t.Fatal("NewIndoorDetectionSystem returned nil")
	}
	if ids.DefaultType != LocationOutdoor {
		t.Errorf("DefaultType = %v, want LocationOutdoor", ids.DefaultType)
	}
}

func TestIndoorDetectionSystem_RegisterZone(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	zone := &IndoorZone{
		ID:   "test_zone",
		Name: "Test Zone",
		Type: LocationIndoor,
		MinX: 0, MaxX: 10,
		MinY: 0, MaxY: 5,
		MinZ: 0, MaxZ: 10,
	}
	ids.RegisterZone(zone)

	if ids.ZoneCount() != 1 {
		t.Errorf("ZoneCount = %d, want 1", ids.ZoneCount())
	}
}

func TestIndoorDetectionSystem_RegisterZone_AutoID(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	zone := &IndoorZone{
		Name: "Auto ID Zone",
		Type: LocationIndoor,
		MinX: 0, MaxX: 10,
		MinY: 0, MaxY: 5,
		MinZ: 0, MaxZ: 10,
	}
	ids.RegisterZone(zone)

	if zone.ID == "" {
		t.Error("Zone should have auto-generated ID")
	}
}

func TestIndoorDetectionSystem_UnregisterZone(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	zone := &IndoorZone{
		ID:   "test_zone",
		Name: "Test Zone",
		Type: LocationIndoor,
		MinX: 0, MaxX: 10,
		MinY: 0, MaxY: 5,
		MinZ: 0, MaxZ: 10,
	}
	ids.RegisterZone(zone)
	ids.UnregisterZone("test_zone")

	if ids.ZoneCount() != 0 {
		t.Errorf("ZoneCount = %d, want 0", ids.ZoneCount())
	}
}

func TestIndoorDetectionSystem_GetLocationType(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	ids.RegisterZone(&IndoorZone{
		ID:   "indoor",
		Type: LocationIndoor,
		MinX: 0, MaxX: 10,
		MinY: 0, MaxY: 5,
		MinZ: 0, MaxZ: 10,
	})

	// Inside the zone
	if lt := ids.GetLocationType(5, 2, 5); lt != LocationIndoor {
		t.Errorf("Inside zone: LocationType = %v, want LocationIndoor", lt)
	}

	// Outside the zone
	if lt := ids.GetLocationType(20, 2, 20); lt != LocationOutdoor {
		t.Errorf("Outside zone: LocationType = %v, want LocationOutdoor", lt)
	}
}

func TestIndoorDetectionSystem_IsIndoors(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	ids.RegisterZone(&IndoorZone{
		ID:   "indoor",
		Type: LocationIndoor,
		MinX: 0, MaxX: 10,
		MinY: 0, MaxY: 5,
		MinZ: 0, MaxZ: 10,
	})

	if !ids.IsIndoors(5, 2, 5) {
		t.Error("Point inside zone should be indoors")
	}
	if ids.IsIndoors(20, 2, 20) {
		t.Error("Point outside zone should not be indoors")
	}
}

func TestIndoorDetectionSystem_IsOutdoors(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	ids.RegisterZone(&IndoorZone{
		ID:   "indoor",
		Type: LocationIndoor,
		MinX: 0, MaxX: 10,
		MinY: 0, MaxY: 5,
		MinZ: 0, MaxZ: 10,
	})

	if ids.IsOutdoors(5, 2, 5) {
		t.Error("Point inside zone should not be outdoors")
	}
	if !ids.IsOutdoors(20, 2, 20) {
		t.Error("Point outside zone should be outdoors")
	}
}

func TestIndoorDetectionSystem_Priority(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)

	// Low priority zone
	ids.RegisterZone(&IndoorZone{
		ID:   "building",
		Type: LocationBuilding,
		MinX: 0, MaxX: 20,
		MinY: 0, MaxY: 10,
		MinZ: 0, MaxZ: 20,
		Priority: 5,
	})

	// High priority zone (overlapping)
	ids.RegisterZone(&IndoorZone{
		ID:   "dungeon",
		Type: LocationDungeon,
		MinX: 5, MaxX: 15,
		MinY: 0, MaxY: 10,
		MinZ: 5, MaxZ: 15,
		Priority: 15,
	})

	// Point in both zones - should return higher priority
	if lt := ids.GetLocationType(10, 5, 10); lt != LocationDungeon {
		t.Errorf("Overlapping point: LocationType = %v, want LocationDungeon", lt)
	}

	// Point only in building zone
	if lt := ids.GetLocationType(1, 5, 1); lt != LocationBuilding {
		t.Errorf("Building only point: LocationType = %v, want LocationBuilding", lt)
	}
}

func TestIndoorDetectionSystem_GetZoneAt(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	zone := &IndoorZone{
		ID:   "test",
		Name: "Test Zone",
		Type: LocationIndoor,
		MinX: 0, MaxX: 10,
		MinY: 0, MaxY: 5,
		MinZ: 0, MaxZ: 10,
	}
	ids.RegisterZone(zone)

	got := ids.GetZoneAt(5, 2, 5)
	if got == nil {
		t.Fatal("GetZoneAt returned nil")
	}
	if got.ID != "test" {
		t.Errorf("Zone ID = %s, want test", got.ID)
	}

	// Outside zone
	got = ids.GetZoneAt(20, 2, 20)
	if got != nil {
		t.Error("GetZoneAt should return nil for point outside zones")
	}
}

func TestIndoorDetectionSystem_GetAllZonesAt(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)

	ids.RegisterZone(&IndoorZone{
		ID:   "zone1",
		Type: LocationBuilding,
		MinX: 0, MaxX: 20,
		MinY: 0, MaxY: 10,
		MinZ: 0, MaxZ: 20,
	})
	ids.RegisterZone(&IndoorZone{
		ID:   "zone2",
		Type: LocationDungeon,
		MinX: 5, MaxX: 15,
		MinY: 0, MaxY: 10,
		MinZ: 5, MaxZ: 15,
	})

	zones := ids.GetAllZonesAt(10, 5, 10)
	if len(zones) != 2 {
		t.Errorf("GetAllZonesAt returned %d zones, want 2", len(zones))
	}
}

func TestIndoorDetectionSystem_RegisterHouseAsZone(t *testing.T) {
	hm := NewHouseManager()
	ids := NewIndoorDetectionSystem(hm)

	house := &House{
		ID:     "house1",
		WorldX: 100,
		WorldZ: 100,
	}
	ids.RegisterHouseAsZone(house, 10, 5, 10)

	if ids.ZoneCount() != 1 {
		t.Errorf("ZoneCount = %d, want 1", ids.ZoneCount())
	}

	// Check point inside house
	if !ids.IsIndoors(100, 2, 100) {
		t.Error("Point inside house should be indoors")
	}
}

func TestIndoorDetectionSystem_RegisterBuildingZone(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	ids.RegisterBuildingZone("shop1", "General Store", 50, 50, 20, 8, 20)

	if ids.ZoneCount() != 1 {
		t.Errorf("ZoneCount = %d, want 1", ids.ZoneCount())
	}

	zone := ids.GetZoneAt(50, 4, 50)
	if zone == nil {
		t.Fatal("Zone not found")
	}
	if zone.Type != LocationBuilding {
		t.Errorf("Zone type = %v, want LocationBuilding", zone.Type)
	}
}

func TestIndoorDetectionSystem_RegisterDungeonZone(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	ids.RegisterDungeonZone("dungeon1", "Dark Cavern", 0, 100, -50, 0, 0, 100)

	zone := ids.GetZoneAt(50, -25, 50)
	if zone == nil {
		t.Fatal("Zone not found")
	}
	if zone.Type != LocationDungeon {
		t.Errorf("Zone type = %v, want LocationDungeon", zone.Type)
	}
}

func TestIndoorDetectionSystem_RegisterCaveZone(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	ids.RegisterCaveZone("cave1", "Crystal Cave", 0, 50, -100, -20, 0, 50)

	zone := ids.GetZoneAt(25, -50, 25)
	if zone == nil {
		t.Fatal("Zone not found")
	}
	if zone.Type != LocationCave {
		t.Errorf("Zone type = %v, want LocationCave", zone.Type)
	}
}

func TestGetLocationTypeName(t *testing.T) {
	tests := []struct {
		locType LocationType
		want    string
	}{
		{LocationOutdoor, "Outdoors"},
		{LocationIndoor, "Indoors"},
		{LocationDungeon, "Dungeon"},
		{LocationCave, "Cave"},
		{LocationBuilding, "Building"},
		{LocationType(99), "Unknown"},
	}

	for _, tc := range tests {
		got := GetLocationTypeName(tc.locType)
		if got != tc.want {
			t.Errorf("GetLocationTypeName(%v) = %s, want %s", tc.locType, got, tc.want)
		}
	}
}

func TestIndoorDetectionSystem_EdgeCases(t *testing.T) {
	ids := NewIndoorDetectionSystem(nil)
	ids.RegisterZone(&IndoorZone{
		ID:   "exact",
		Type: LocationIndoor,
		MinX: 0, MaxX: 10,
		MinY: 0, MaxY: 5,
		MinZ: 0, MaxZ: 10,
	})

	// Test exact boundary points (should be inside)
	if !ids.IsIndoors(0, 0, 0) {
		t.Error("Point at min boundary should be indoors")
	}
	if !ids.IsIndoors(10, 5, 10) {
		t.Error("Point at max boundary should be indoors")
	}

	// Test just outside boundary
	if ids.IsIndoors(-0.1, 0, 0) {
		t.Error("Point just outside min X should be outdoors")
	}
	if ids.IsIndoors(10.1, 0, 0) {
		t.Error("Point just outside max X should be outdoors")
	}
}

func BenchmarkIndoorDetectionSystem_GetLocationType(b *testing.B) {
	ids := NewIndoorDetectionSystem(nil)

	// Add 100 zones
	for i := 0; i < 100; i++ {
		x := float64(i * 20)
		ids.RegisterZone(&IndoorZone{
			ID:   fmt.Sprintf("zone_%d", i),
			Type: LocationIndoor,
			MinX: x, MaxX: x + 10,
			MinY: 0, MaxY: 5,
			MinZ: 0, MaxZ: 100,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ids.GetLocationType(500, 2, 50)
	}
}
