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
