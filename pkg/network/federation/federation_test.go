package federation

import (
	"bytes"
	"testing"
	"time"
)

func TestPlayerTransferValidate(t *testing.T) {
	tests := []struct {
		name     string
		transfer *PlayerTransfer
		wantErr  bool
	}{
		{
			name: "valid transfer",
			transfer: &PlayerTransfer{
				PlayerID:     1,
				AccountID:    "acc1",
				SourceServer: "server1",
				TransferTime: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing player ID",
			transfer: &PlayerTransfer{
				AccountID:    "acc1",
				SourceServer: "server1",
				TransferTime: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing account ID",
			transfer: &PlayerTransfer{
				PlayerID:     1,
				SourceServer: "server1",
				TransferTime: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "expired transfer",
			transfer: &PlayerTransfer{
				PlayerID:     1,
				AccountID:    "acc1",
				SourceServer: "server1",
				TransferTime: time.Now().Add(-10 * time.Minute),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transfer.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFederationRegisterNode(t *testing.T) {
	f := NewFederation("local")

	node := &FederationNode{
		ServerID: "remote1",
		Address:  "192.168.1.1:7777",
	}
	f.RegisterNode(node)

	if f.NodeCount() != 1 {
		t.Errorf("NodeCount = %d, want 1", f.NodeCount())
	}

	got := f.GetNode("remote1")
	if got == nil {
		t.Error("GetNode returned nil")
	}
	if got.Address != "192.168.1.1:7777" {
		t.Errorf("Node address = %s, want 192.168.1.1:7777", got.Address)
	}

	f.UnregisterNode("remote1")
	if f.NodeCount() != 0 {
		t.Errorf("NodeCount after unregister = %d, want 0", f.NodeCount())
	}
}

func TestFederationTransfer(t *testing.T) {
	source := NewFederation("server1")
	dest := NewFederation("server2")

	// Create transfer with inventory and quest state
	transfer := &PlayerTransfer{
		PlayerID:    1,
		AccountID:   "acc1",
		DisplayName: "TestPlayer",
		DestX:       100, DestY: 0, DestZ: 200,
		HealthCurrent: 80,
		HealthMax:     100,
		Inventory:     []string{"sword", "shield", "potion"},
		SkillLevels:   map[string]int{"combat": 25, "magic": 10},
		QuestFlags: map[string]map[string]bool{
			"main_quest": {"objective1": true, "objective2": false},
		},
		Standings: map[string]float64{"faction1": 50, "faction2": -20},
	}

	// Initiate on source
	if err := source.InitiateTransfer(transfer); err != nil {
		t.Fatalf("InitiateTransfer failed: %v", err)
	}

	pending := source.GetPendingTransfer(1)
	if pending == nil {
		t.Error("Pending transfer not found")
	}

	// Accept on destination
	received, err := dest.AcceptTransfer(transfer)
	if err != nil {
		t.Fatalf("AcceptTransfer failed: %v", err)
	}

	// Per AC: inventory retained
	if len(received.Inventory) != 3 {
		t.Errorf("Inventory size = %d, want 3", len(received.Inventory))
	}
	if received.Inventory[0] != "sword" {
		t.Errorf("Inventory[0] = %s, want sword", received.Inventory[0])
	}

	// Per AC: quest state retained
	questFlags := received.QuestFlags["main_quest"]
	if questFlags == nil {
		t.Error("Quest flags not retained")
	}
	if !questFlags["objective1"] {
		t.Error("Quest objective1 should be true")
	}

	// Skills retained
	if received.SkillLevels["combat"] != 25 {
		t.Errorf("Skill combat = %d, want 25", received.SkillLevels["combat"])
	}

	// Faction standings retained
	if received.Standings["faction1"] != 50 {
		t.Errorf("Faction1 standing = %f, want 50", received.Standings["faction1"])
	}

	// Complete transfer
	source.CompleteTransfer(1, true)
	if source.GetPendingTransfer(1) != nil {
		t.Error("Pending transfer should be cleared after complete")
	}
}

func TestFederationDuplicateTransfer(t *testing.T) {
	f := NewFederation("server1")

	transfer := &PlayerTransfer{
		PlayerID:     1,
		AccountID:    "acc1",
		SourceServer: "server2",
		TransferTime: time.Now(),
	}

	// Accept first transfer
	_, err := f.AcceptTransfer(transfer)
	if err != nil {
		t.Fatalf("First AcceptTransfer failed: %v", err)
	}

	// Duplicate should fail
	_, err = f.AcceptTransfer(transfer)
	if err == nil {
		t.Error("Duplicate AcceptTransfer should fail")
	}
}

func TestFederationPriceSignal(t *testing.T) {
	f := NewFederation("local")

	signal := &PriceSignal{
		ServerID: "server1",
		CityID:   "city1",
		PriceTable: map[string]float64{
			"sword":  100,
			"shield": 75,
		},
		Timestamp: time.Now(),
	}

	f.ProcessPriceSignal(signal)

	prices := f.GetRemotePrices()
	if len(prices) != 1 {
		t.Errorf("GetRemotePrices returned %d signals, want 1", len(prices))
	}

	if prices[0].PriceTable["sword"] != 100 {
		t.Errorf("Sword price = %f, want 100", prices[0].PriceTable["sword"])
	}
}

func TestFederationGlobalEvent(t *testing.T) {
	f := NewFederation("local")

	event := &GlobalEvent{
		EventID:     "event1",
		EventType:   "dragon_attack",
		Description: "A dragon attacks the region",
		StartTime:   time.Now(),
		Duration:    time.Hour,
	}
	event.AffectedArea.CenterX = 100
	event.AffectedArea.CenterZ = 200
	event.AffectedArea.Radius = 500

	f.BroadcastEvent(event)

	events := f.GetActiveEvents()
	if len(events) != 1 {
		t.Errorf("GetActiveEvents returned %d events, want 1", len(events))
	}

	if events[0].EventType != "dragon_attack" {
		t.Errorf("Event type = %s, want dragon_attack", events[0].EventType)
	}
}

func TestFederationCleanupExpired(t *testing.T) {
	f := NewFederation("local")

	// Add expired transfer
	f.pendingTransfers[1] = &PlayerTransfer{
		PlayerID:     1,
		TransferTime: time.Now().Add(-10 * time.Minute),
	}

	// Add active transfer
	f.pendingTransfers[2] = &PlayerTransfer{
		PlayerID:     2,
		TransferTime: time.Now(),
	}

	// Add expired event
	f.activeEvents["expired"] = &GlobalEvent{
		EventID:   "expired",
		StartTime: time.Now().Add(-2 * time.Hour),
		Duration:  time.Hour,
	}

	// Add active event
	f.activeEvents["active"] = &GlobalEvent{
		EventID:   "active",
		StartTime: time.Now(),
		Duration:  time.Hour,
	}

	f.CleanupExpired()

	if _, exists := f.pendingTransfers[1]; exists {
		t.Error("Expired pending transfer should be cleaned up")
	}
	if _, exists := f.pendingTransfers[2]; !exists {
		t.Error("Active pending transfer should remain")
	}

	if _, exists := f.activeEvents["expired"]; exists {
		t.Error("Expired event should be cleaned up")
	}
	if _, exists := f.activeEvents["active"]; !exists {
		t.Error("Active event should remain")
	}
}

func TestEncodeDecodeTransfer(t *testing.T) {
	original := &PlayerTransfer{
		PlayerID:        1,
		AccountID:       "acc1",
		DisplayName:     "TestPlayer",
		DestX:           100,
		DestY:           0,
		DestZ:           200,
		HealthCurrent:   80,
		HealthMax:       100,
		Inventory:       []string{"sword", "shield"},
		SkillLevels:     map[string]int{"combat": 25},
		SkillExperience: map[string]float64{"combat": 500.5},
		QuestFlags:      map[string]map[string]bool{"quest1": {"done": true}},
		Standings:       map[string]float64{"faction1": 50},
		TransferTime:    time.Now(),
		SourceServer:    "server1",
	}

	var buf bytes.Buffer
	if err := EncodeTransfer(&buf, original); err != nil {
		t.Fatalf("EncodeTransfer failed: %v", err)
	}

	decoded, err := DecodeTransfer(&buf)
	if err != nil {
		t.Fatalf("DecodeTransfer failed: %v", err)
	}

	if decoded.PlayerID != original.PlayerID {
		t.Errorf("PlayerID = %d, want %d", decoded.PlayerID, original.PlayerID)
	}
	if decoded.DisplayName != original.DisplayName {
		t.Errorf("DisplayName = %s, want %s", decoded.DisplayName, original.DisplayName)
	}
	if len(decoded.Inventory) != len(original.Inventory) {
		t.Errorf("Inventory length = %d, want %d", len(decoded.Inventory), len(original.Inventory))
	}
	if decoded.SkillLevels["combat"] != original.SkillLevels["combat"] {
		t.Errorf("SkillLevels[combat] = %d, want %d", decoded.SkillLevels["combat"], original.SkillLevels["combat"])
	}
}

func TestEncodeDecodePriceSignal(t *testing.T) {
	original := &PriceSignal{
		ServerID: "server1",
		CityID:   "city1",
		PriceTable: map[string]float64{
			"sword":  100,
			"shield": 75,
		},
		Timestamp: time.Now(),
	}

	var buf bytes.Buffer
	if err := EncodePriceSignal(&buf, original); err != nil {
		t.Fatalf("EncodePriceSignal failed: %v", err)
	}

	decoded, err := DecodePriceSignal(&buf)
	if err != nil {
		t.Fatalf("DecodePriceSignal failed: %v", err)
	}

	if decoded.ServerID != original.ServerID {
		t.Errorf("ServerID = %s, want %s", decoded.ServerID, original.ServerID)
	}
	if decoded.PriceTable["sword"] != original.PriceTable["sword"] {
		t.Errorf("PriceTable[sword] = %f, want %f", decoded.PriceTable["sword"], original.PriceTable["sword"])
	}
}

func TestEncodeDecodeGlobalEvent(t *testing.T) {
	original := &GlobalEvent{
		EventID:     "event1",
		EventType:   "dragon_attack",
		Description: "A dragon attacks!",
		StartTime:   time.Now(),
		Duration:    time.Hour,
	}
	original.AffectedArea.CenterX = 100
	original.AffectedArea.CenterZ = 200
	original.AffectedArea.Radius = 500

	var buf bytes.Buffer
	if err := EncodeGlobalEvent(&buf, original); err != nil {
		t.Fatalf("EncodeGlobalEvent failed: %v", err)
	}

	decoded, err := DecodeGlobalEvent(&buf)
	if err != nil {
		t.Fatalf("DecodeGlobalEvent failed: %v", err)
	}

	if decoded.EventID != original.EventID {
		t.Errorf("EventID = %s, want %s", decoded.EventID, original.EventID)
	}
	if decoded.AffectedArea.Radius != original.AffectedArea.Radius {
		t.Errorf("Radius = %f, want %f", decoded.AffectedArea.Radius, original.AffectedArea.Radius)
	}
}
