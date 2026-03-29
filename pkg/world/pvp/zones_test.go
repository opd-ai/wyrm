package pvp

import (
	"math/rand"
	"testing"
)

func TestZoneContainsPoint(t *testing.T) {
	zone := &Zone{
		ID:   "test",
		MinX: 0, MaxX: 100,
		MinZ: 0, MaxZ: 100,
	}

	tests := []struct {
		x, z     float64
		expected bool
	}{
		{50, 50, true},
		{0, 0, true},
		{100, 100, true},
		{-1, 50, false},
		{101, 50, false},
		{50, -1, false},
		{50, 101, false},
	}

	for _, tt := range tests {
		got := zone.ContainsPoint(tt.x, tt.z)
		if got != tt.expected {
			t.Errorf("ContainsPoint(%f, %f) = %v, want %v", tt.x, tt.z, got, tt.expected)
		}
	}
}

func TestZoneManagerAddRemove(t *testing.T) {
	m := NewZoneManager()

	zone := &Zone{ID: "zone1", Type: ZoneContested}
	m.AddZone(zone)

	if m.ZoneCount() != 1 {
		t.Errorf("ZoneCount = %d, want 1", m.ZoneCount())
	}

	got := m.GetZone("zone1")
	if got == nil {
		t.Error("GetZone returned nil")
	}
	if got.Type != ZoneContested {
		t.Errorf("Zone type = %d, want %d", got.Type, ZoneContested)
	}

	m.RemoveZone("zone1")
	if m.ZoneCount() != 0 {
		t.Errorf("ZoneCount after remove = %d, want 0", m.ZoneCount())
	}
}

func TestZoneManagerGetZoneAt(t *testing.T) {
	m := NewZoneManager()

	m.AddZone(&Zone{ID: "safe", Type: ZoneSafe, MinX: 0, MaxX: 100, MinZ: 0, MaxZ: 100})
	m.AddZone(&Zone{ID: "hostile", Type: ZoneHostile, MinX: 200, MaxX: 300, MinZ: 0, MaxZ: 100})

	tests := []struct {
		x, z       float64
		expectedID string
	}{
		{50, 50, "safe"},
		{250, 50, "hostile"},
		{150, 50, ""}, // No zone
	}

	for _, tt := range tests {
		zone := m.GetZoneAt(tt.x, tt.z)
		gotID := ""
		if zone != nil {
			gotID = zone.ID
		}
		if gotID != tt.expectedID {
			t.Errorf("GetZoneAt(%f, %f) = %s, want %s", tt.x, tt.z, gotID, tt.expectedID)
		}
	}
}

func TestPlayerFlags(t *testing.T) {
	m := NewZoneManager()

	// Initially unflagged
	if m.IsPlayerFlagged(1) {
		t.Error("Player should not be flagged initially")
	}

	// Set flag
	m.SetPlayerFlag(1, true)
	if !m.IsPlayerFlagged(1) {
		t.Error("Player should be flagged after SetPlayerFlag(true)")
	}

	if m.FlaggedPlayerCount() != 1 {
		t.Errorf("FlaggedPlayerCount = %d, want 1", m.FlaggedPlayerCount())
	}

	// Clear flag
	m.SetPlayerFlag(1, false)
	if m.IsPlayerFlagged(1) {
		t.Error("Player should not be flagged after SetPlayerFlag(false)")
	}
}

func TestClearExpiredFlags(t *testing.T) {
	m := NewZoneManager()

	m.SetPlayerFlag(1, true)
	m.SetFlagCooldown(1, 1000)

	m.SetPlayerFlag(2, true)
	m.SetFlagCooldown(2, 2000)

	// Clear at time 1500 - player 1 should unflag, player 2 should stay
	m.ClearExpiredFlags(1500)

	if m.IsPlayerFlagged(1) {
		t.Error("Player 1 flag should have expired")
	}
	if !m.IsPlayerFlagged(2) {
		t.Error("Player 2 flag should still be active")
	}
}

func TestCheckCombatSafeZone(t *testing.T) {
	m := NewZoneManager()
	m.AddZone(&Zone{ID: "safe", Type: ZoneSafe, MinX: 0, MaxX: 100, MinZ: 0, MaxZ: 100})

	// Both in safe zone - no damage regardless of flags
	m.SetPlayerFlag(1, true)
	m.SetPlayerFlag(2, true)

	result := m.CheckCombat(1, 50, 50, 2, 50, 50, 100)
	if result.DamageAllowed {
		t.Error("Damage should not be allowed in safe zone")
	}
}

func TestCheckCombatContestedZone(t *testing.T) {
	m := NewZoneManager()
	m.AddZone(&Zone{ID: "contested", Type: ZoneContested, MinX: 0, MaxX: 100, MinZ: 0, MaxZ: 100})

	// Both flagged in contested zone - damage allowed
	m.SetPlayerFlag(1, true)
	m.SetPlayerFlag(2, true)

	result := m.CheckCombat(1, 50, 50, 2, 50, 50, 100)
	if !result.DamageAllowed {
		t.Error("Damage should be allowed when both flagged in contested zone")
	}
	if result.DamageAmount != 100 {
		t.Errorf("DamageAmount = %f, want 100", result.DamageAmount)
	}

	// Attacker flagged, defender unflagged - no damage (per AC)
	m.SetPlayerFlag(2, false)
	result = m.CheckCombat(1, 50, 50, 2, 50, 50, 100)
	if result.DamageAllowed {
		t.Error("Unflagged player should take no damage from flagged")
	}
}

func TestCheckCombatHostileZone(t *testing.T) {
	m := NewZoneManager()
	m.AddZone(&Zone{ID: "hostile", Type: ZoneHostile, MinX: 0, MaxX: 100, MinZ: 0, MaxZ: 100})

	// Hostile zone - damage always allowed
	result := m.CheckCombat(1, 50, 50, 2, 50, 50, 100)
	if !result.DamageAllowed {
		t.Error("Damage should be allowed in hostile zone")
	}
}

func TestCalculateDeathLoot(t *testing.T) {
	m := NewZoneManager()
	m.AddZone(&Zone{
		ID:   "pvp",
		Type: ZoneHostile,
		MinX: 0, MaxX: 100,
		MinZ: 0, MaxZ: 100,
		LootDropRate: 0.5,
	})

	rng := rand.New(rand.NewSource(42))
	inventory := []string{"sword", "shield", "potion", "gold"}

	loot := m.CalculateDeathLoot(1, inventory, 50, 50, rng)

	if loot == nil {
		t.Fatal("Loot should not be nil")
	}

	// Per AC: drops ≥1 item
	if len(loot.Items) < 1 {
		t.Errorf("Should drop at least 1 item, got %d", len(loot.Items))
	}

	// Should drop ~50% (2 items) with this rate
	if len(loot.Items) > len(inventory) {
		t.Errorf("Should not drop more than inventory size")
	}
}

func TestCalculateDeathLootMinimumOne(t *testing.T) {
	m := NewZoneManager()
	m.AddZone(&Zone{
		ID:   "pvp",
		Type: ZoneHostile,
		MinX: 0, MaxX: 100,
		MinZ: 0, MaxZ: 100,
		LootDropRate: 0.01, // Very low rate
	})

	rng := rand.New(rand.NewSource(42))
	inventory := []string{"sword"}

	loot := m.CalculateDeathLoot(1, inventory, 50, 50, rng)

	// Per AC: drops ≥1 item
	if loot == nil {
		t.Fatal("Loot should not be nil")
	}
	if len(loot.Items) < 1 {
		t.Errorf("Should drop at least 1 item, got %d", len(loot.Items))
	}
}

func TestCalculateDeathLootSafeZone(t *testing.T) {
	m := NewZoneManager()
	m.AddZone(&Zone{ID: "safe", Type: ZoneSafe, MinX: 0, MaxX: 100, MinZ: 0, MaxZ: 100})

	rng := rand.New(rand.NewSource(42))
	inventory := []string{"sword", "shield"}

	loot := m.CalculateDeathLoot(1, inventory, 50, 50, rng)

	if loot != nil {
		t.Error("Should not drop loot in safe zone")
	}
}

func TestRespawnPoint(t *testing.T) {
	m := NewZoneManager()
	m.AddZone(&Zone{
		ID:   "pvp",
		Type: ZoneHostile,
		MinX: 0, MaxX: 100,
		MinZ: 0, MaxZ: 100,
		RespawnX: 10,
		RespawnZ: 20,
	})

	x, z := m.RespawnPoint(50, 50)
	if x != 10 || z != 20 {
		t.Errorf("RespawnPoint = (%f, %f), want (10, 20)", x, z)
	}

	// Position outside zones
	x, z = m.RespawnPoint(500, 500)
	if x != 0 || z != 0 {
		t.Errorf("RespawnPoint outside zones = (%f, %f), want (0, 0)", x, z)
	}
}
