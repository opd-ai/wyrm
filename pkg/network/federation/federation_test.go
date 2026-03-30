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

// TestFullTransferIntegration tests the complete transfer workflow between two
// Federation instances, simulating the actual network transfer by encoding
// and decoding the transfer data as would happen over a real connection.
// This tests the full InitiateTransfer → Encode → Decode → AcceptTransfer → CompleteTransfer flow.
func TestFullTransferIntegration(t *testing.T) {
	// Set up two federation instances simulating two servers
	sourceServer := NewFederation("source-server-001")
	destServer := NewFederation("dest-server-002")

	// Register peer nodes (as would happen in production)
	sourceServer.RegisterNode(&FederationNode{
		ServerID: "dest-server-002",
		Address:  "192.168.1.2:7777",
	})
	destServer.RegisterNode(&FederationNode{
		ServerID: "source-server-001",
		Address:  "192.168.1.1:7777",
	})

	// Create a comprehensive player transfer with all state
	transfer := &PlayerTransfer{
		PlayerID:      42,
		AccountID:     "user_12345",
		DisplayName:   "TestHero",
		DestX:         1000.5,
		DestY:         0.0,
		DestZ:         2500.75,
		DestAngle:     1.57,
		HealthCurrent: 85.5,
		HealthMax:     100.0,
		Inventory:     []string{"legendary_sword", "health_potion_x3", "map_fragment", "gold_coins"},
		SkillLevels: map[string]int{
			"combat":    35,
			"magic":     20,
			"stealth":   15,
			"crafting":  10,
			"diplomacy": 5,
		},
		SkillExperience: map[string]float64{
			"combat":    12500.75,
			"magic":     5000.25,
			"stealth":   2500.0,
			"crafting":  1000.0,
			"diplomacy": 250.0,
		},
		QuestFlags: map[string]map[string]bool{
			"main_story": {
				"prologue_complete": true,
				"chapter1_started":  true,
				"met_mentor":        true,
				"defeated_boss1":    false,
			},
			"faction_quest_1": {
				"accepted": true,
				"step1":    true,
				"step2":    false,
			},
		},
		Standings: map[string]float64{
			"merchants_guild": 75.5,
			"thieves_guild":   -25.0,
			"royal_guard":     50.0,
			"mage_circle":     30.0,
		},
	}

	// STEP 1: Source server initiates transfer
	if err := sourceServer.InitiateTransfer(transfer); err != nil {
		t.Fatalf("InitiateTransfer failed: %v", err)
	}

	// Verify transfer is pending on source
	pendingTransfer := sourceServer.GetPendingTransfer(transfer.PlayerID)
	if pendingTransfer == nil {
		t.Fatal("Transfer should be pending on source server")
	}
	if pendingTransfer.SourceServer != "source-server-001" {
		t.Errorf("SourceServer = %s, want source-server-001", pendingTransfer.SourceServer)
	}

	// STEP 2: Encode transfer data for network transmission
	var networkBuffer bytes.Buffer
	if err := EncodeTransfer(&networkBuffer, pendingTransfer); err != nil {
		t.Fatalf("EncodeTransfer failed: %v", err)
	}

	// Verify data was written
	if networkBuffer.Len() == 0 {
		t.Fatal("Encoded transfer should have non-zero length")
	}

	// STEP 3: Destination server decodes transfer data
	receivedTransfer, err := DecodeTransfer(&networkBuffer)
	if err != nil {
		t.Fatalf("DecodeTransfer failed: %v", err)
	}

	// STEP 4: Destination server accepts transfer
	acceptedTransfer, err := destServer.AcceptTransfer(receivedTransfer)
	if err != nil {
		t.Fatalf("AcceptTransfer failed: %v", err)
	}

	// STEP 5: Verify all player state was preserved
	// Per ROADMAP Phase 5 Item 24 AC: "Player moves between two local test servers
	// retaining inventory and quest state"

	// Verify identity
	if acceptedTransfer.PlayerID != 42 {
		t.Errorf("PlayerID = %d, want 42", acceptedTransfer.PlayerID)
	}
	if acceptedTransfer.DisplayName != "TestHero" {
		t.Errorf("DisplayName = %s, want TestHero", acceptedTransfer.DisplayName)
	}

	// Verify position
	if acceptedTransfer.DestX != 1000.5 {
		t.Errorf("DestX = %f, want 1000.5", acceptedTransfer.DestX)
	}
	if acceptedTransfer.DestZ != 2500.75 {
		t.Errorf("DestZ = %f, want 2500.75", acceptedTransfer.DestZ)
	}

	// Verify health
	if acceptedTransfer.HealthCurrent != 85.5 {
		t.Errorf("HealthCurrent = %f, want 85.5", acceptedTransfer.HealthCurrent)
	}

	// Verify inventory
	if len(acceptedTransfer.Inventory) != 4 {
		t.Errorf("Inventory length = %d, want 4", len(acceptedTransfer.Inventory))
	}
	if acceptedTransfer.Inventory[0] != "legendary_sword" {
		t.Errorf("Inventory[0] = %s, want legendary_sword", acceptedTransfer.Inventory[0])
	}

	// Verify skills
	if acceptedTransfer.SkillLevels["combat"] != 35 {
		t.Errorf("SkillLevels[combat] = %d, want 35", acceptedTransfer.SkillLevels["combat"])
	}
	if acceptedTransfer.SkillExperience["combat"] != 12500.75 {
		t.Errorf("SkillExperience[combat] = %f, want 12500.75", acceptedTransfer.SkillExperience["combat"])
	}

	// Verify quest state
	mainQuestFlags := acceptedTransfer.QuestFlags["main_story"]
	if mainQuestFlags == nil {
		t.Fatal("main_story quest flags not preserved")
	}
	if !mainQuestFlags["prologue_complete"] {
		t.Error("prologue_complete should be true")
	}
	if mainQuestFlags["defeated_boss1"] {
		t.Error("defeated_boss1 should be false")
	}

	// Verify faction standings
	if acceptedTransfer.Standings["merchants_guild"] != 75.5 {
		t.Errorf("merchants_guild standing = %f, want 75.5", acceptedTransfer.Standings["merchants_guild"])
	}
	if acceptedTransfer.Standings["thieves_guild"] != -25.0 {
		t.Errorf("thieves_guild standing = %f, want -25.0", acceptedTransfer.Standings["thieves_guild"])
	}

	// STEP 6: Source server completes transfer
	sourceServer.CompleteTransfer(transfer.PlayerID, true)
	if sourceServer.GetPendingTransfer(transfer.PlayerID) != nil {
		t.Error("Transfer should be cleared after completion")
	}

	// STEP 7: Verify duplicate prevention on destination
	_, err = destServer.AcceptTransfer(receivedTransfer)
	if err == nil {
		t.Error("Duplicate transfer should be rejected")
	}
}

// TestConcurrentTransfers tests that multiple transfers can be processed concurrently
// without data corruption or race conditions.
func TestConcurrentTransfers(t *testing.T) {
	source := NewFederation("source")
	dest := NewFederation("dest")

	// Start multiple transfers concurrently
	const numTransfers = 10
	errCh := make(chan error, numTransfers*2)

	for i := 0; i < numTransfers; i++ {
		playerID := uint64(i + 1)
		go func(id uint64) {
			transfer := &PlayerTransfer{
				PlayerID:    id,
				AccountID:   "acc-" + string(rune('A'+int(id))),
				DisplayName: "Player" + string(rune('0'+int(id))),
				Inventory:   []string{"item1", "item2"},
				SkillLevels: map[string]int{"combat": int(id * 10)},
				QuestFlags:  map[string]map[string]bool{"q1": {"done": true}},
				Standings:   map[string]float64{"f1": float64(id)},
			}
			errCh <- source.InitiateTransfer(transfer)
		}(playerID)
	}

	// Wait for all initiations
	for i := 0; i < numTransfers; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("Concurrent InitiateTransfer failed: %v", err)
		}
	}

	// Verify all transfers are pending
	for i := uint64(1); i <= numTransfers; i++ {
		if source.GetPendingTransfer(i) == nil {
			t.Errorf("Transfer %d should be pending", i)
		}
	}

	// Accept all transfers concurrently on destination
	for i := uint64(1); i <= numTransfers; i++ {
		transfer := source.GetPendingTransfer(i)
		go func(pt *PlayerTransfer) {
			_, err := dest.AcceptTransfer(pt)
			errCh <- err
		}(transfer)
	}

	// Wait for all acceptances
	for i := 0; i < numTransfers; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("Concurrent AcceptTransfer failed: %v", err)
		}
	}

	// Complete all transfers
	for i := uint64(1); i <= numTransfers; i++ {
		source.CompleteTransfer(i, true)
	}

	// Verify all are cleared
	for i := uint64(1); i <= numTransfers; i++ {
		if source.GetPendingTransfer(i) != nil {
			t.Errorf("Transfer %d should be completed", i)
		}
	}
}
