// Package federation provides cross-server player travel and economy synchronization.
// This file contains integration tests that spin up multiple server instances
// to validate end-to-end federation functionality.
package federation

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// FederationServer simulates a game server with federation support for testing.
type FederationServer struct {
	ID         string
	Federation *Federation
	Address    string
	Listener   net.Listener
	mu         sync.Mutex
	running    bool
	peers      map[string]net.Conn // serverID -> connection
	received   []interface{}       // received messages for verification
}

// NewFederationServer creates a new test server instance.
func NewFederationServer(id, address string) *FederationServer {
	return &FederationServer{
		ID:         id,
		Federation: NewFederation(id),
		Address:    address,
		peers:      make(map[string]net.Conn),
		received:   make([]interface{}, 0),
	}
}

// Start begins listening for federation connections.
func (s *FederationServer) Start() error {
	ln, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.Listener = ln
	s.running = true
	s.mu.Unlock()

	go s.acceptLoop()
	return nil
}

// acceptLoop handles incoming connections.
func (s *FederationServer) acceptLoop() {
	for {
		s.mu.Lock()
		if !s.running || s.Listener == nil {
			s.mu.Unlock()
			return
		}
		listener := s.Listener
		s.mu.Unlock()

		conn, err := listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()
			if !running {
				return
			}
			continue
		}
		go s.handleConnection(conn)
	}
}

// handleConnection processes incoming federation messages.
func (s *FederationServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		// Read message type
		var msgType [1]byte
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, err := conn.Read(msgType[:])
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				s.mu.Lock()
				running := s.running
				s.mu.Unlock()
				if !running {
					return
				}
				continue
			}
			return
		}

		switch msgType[0] {
		case MsgTypeTransfer:
			transfer, err := DecodeTransfer(conn)
			if err != nil {
				return
			}
			s.mu.Lock()
			s.received = append(s.received, transfer)
			s.mu.Unlock()
			// Accept the transfer
			_, _ = s.Federation.AcceptTransfer(transfer)

		case MsgTypePriceSignal:
			signal, err := DecodePriceSignal(conn)
			if err != nil {
				return
			}
			s.mu.Lock()
			s.received = append(s.received, signal)
			s.mu.Unlock()
			s.Federation.ProcessPriceSignal(signal)

		case MsgTypeGlobalEvent:
			event, err := DecodeGlobalEvent(conn)
			if err != nil {
				return
			}
			s.mu.Lock()
			s.received = append(s.received, event)
			s.mu.Unlock()
			s.Federation.BroadcastEvent(event)
		}
	}
}

// ConnectToPeer establishes a connection to another federation server.
func (s *FederationServer) ConnectToPeer(peerID, address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.peers[peerID] = conn
	s.Federation.RegisterNode(&Node{
		ServerID: peerID,
		Address:  address,
		LastSeen: time.Now(),
	})
	s.mu.Unlock()
	return nil
}

// SendTransfer sends a player transfer to a peer server.
func (s *FederationServer) SendTransfer(peerID string, transfer *PlayerTransfer) error {
	s.mu.Lock()
	conn, ok := s.peers[peerID]
	s.mu.Unlock()
	if !ok {
		return net.ErrClosed
	}

	// Write message type
	if _, err := conn.Write([]byte{MsgTypeTransfer}); err != nil {
		return err
	}
	return EncodeTransfer(conn, transfer)
}

// SendPriceSignal sends a price signal to a peer server.
func (s *FederationServer) SendPriceSignal(peerID string, signal *PriceSignal) error {
	s.mu.Lock()
	conn, ok := s.peers[peerID]
	s.mu.Unlock()
	if !ok {
		return net.ErrClosed
	}

	if _, err := conn.Write([]byte{MsgTypePriceSignal}); err != nil {
		return err
	}
	return EncodePriceSignal(conn, signal)
}

// SendGlobalEvent sends a global event to a peer server.
func (s *FederationServer) SendGlobalEvent(peerID string, event *GlobalEvent) error {
	s.mu.Lock()
	conn, ok := s.peers[peerID]
	s.mu.Unlock()
	if !ok {
		return net.ErrClosed
	}

	if _, err := conn.Write([]byte{MsgTypeGlobalEvent}); err != nil {
		return err
	}
	return EncodeGlobalEvent(conn, event)
}

// GetReceivedMessages returns a copy of received messages.
func (s *FederationServer) GetReceivedMessages() []interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]interface{}, len(s.received))
	copy(result, s.received)
	return result
}

// Stop shuts down the server and closes all connections.
func (s *FederationServer) Stop() error {
	s.mu.Lock()
	s.running = false
	for _, conn := range s.peers {
		conn.Close()
	}
	s.peers = make(map[string]net.Conn)
	listener := s.Listener
	s.Listener = nil
	s.mu.Unlock()

	if listener != nil {
		return listener.Close()
	}
	return nil
}

// setupTestServerPair creates, starts, and connects two test servers.
// Returns the servers and a cleanup function. The cleanup function should be deferred.
func setupTestServerPair(t *testing.T, name1, addr1, name2, addr2 string) (*FederationServer, *FederationServer, func()) {
	t.Helper()
	server1 := NewFederationServer(name1, addr1)
	server2 := NewFederationServer(name2, addr2)

	if err := server1.Start(); err != nil {
		t.Fatalf("Failed to start server1: %v", err)
	}
	if err := server2.Start(); err != nil {
		server1.Stop()
		t.Fatalf("Failed to start server2: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := server1.ConnectToPeer(name2, addr2); err != nil {
		server1.Stop()
		server2.Stop()
		t.Fatalf("Failed to connect servers: %v", err)
	}

	cleanup := func() {
		server1.Stop()
		server2.Stop()
	}
	return server1, server2, cleanup
}

// TestFederationIntegrationTwoServers tests player transfer between two actual
// TCP server instances, validating the complete transfer workflow.
func TestFederationIntegrationTwoServers(t *testing.T) {
	server1, server2, cleanup := setupTestServerPair(t,
		"server-alpha", "127.0.0.1:17781",
		"server-beta", "127.0.0.1:17782")
	defer cleanup()

	// Create a comprehensive player transfer
	transfer := &PlayerTransfer{
		PlayerID:      123,
		AccountID:     "integration-test-user",
		DisplayName:   "IntegrationHero",
		DestX:         500.5,
		DestY:         0.0,
		DestZ:         750.25,
		DestAngle:     1.0,
		HealthCurrent: 90.0,
		HealthMax:     100.0,
		Inventory:     []string{"integration_sword", "test_potion", "debug_armor"},
		SkillLevels: map[string]int{
			"combat": 50,
			"magic":  25,
		},
		SkillExperience: map[string]float64{
			"combat": 25000.0,
			"magic":  6250.0,
		},
		QuestFlags: map[string]map[string]bool{
			"integration_quest": {
				"started":   true,
				"completed": false,
			},
		},
		Standings: map[string]float64{
			"test_faction": 100.0,
		},
		TransferTime: time.Now(),
		SourceServer: "server-alpha",
	}

	// Initiate transfer on server1
	if err := server1.Federation.InitiateTransfer(transfer); err != nil {
		t.Fatalf("InitiateTransfer failed: %v", err)
	}

	// Send transfer to server2 via TCP
	if err := server1.SendTransfer("server-beta", transfer); err != nil {
		t.Fatalf("SendTransfer failed: %v", err)
	}

	// Wait for message to be received
	time.Sleep(100 * time.Millisecond)

	// Verify server2 received the transfer
	messages := server2.GetReceivedMessages()
	if len(messages) == 0 {
		t.Fatal("Server2 should have received a transfer message")
	}

	receivedTransfer, ok := messages[0].(*PlayerTransfer)
	if !ok {
		t.Fatalf("Expected PlayerTransfer, got %T", messages[0])
	}

	// Validate all transfer data was preserved
	if receivedTransfer.PlayerID != 123 {
		t.Errorf("PlayerID = %d, want 123", receivedTransfer.PlayerID)
	}
	if receivedTransfer.DisplayName != "IntegrationHero" {
		t.Errorf("DisplayName = %s, want IntegrationHero", receivedTransfer.DisplayName)
	}
	if len(receivedTransfer.Inventory) != 3 {
		t.Errorf("Inventory length = %d, want 3", len(receivedTransfer.Inventory))
	}
	if receivedTransfer.SkillLevels["combat"] != 50 {
		t.Errorf("combat skill = %d, want 50", receivedTransfer.SkillLevels["combat"])
	}
	if !receivedTransfer.QuestFlags["integration_quest"]["started"] {
		t.Error("Quest flag 'started' should be true")
	}
	if receivedTransfer.Standings["test_faction"] != 100.0 {
		t.Errorf("test_faction standing = %f, want 100.0", receivedTransfer.Standings["test_faction"])
	}

	// Complete transfer on source server
	server1.Federation.CompleteTransfer(123, true)
	if server1.Federation.GetPendingTransfer(123) != nil {
		t.Error("Transfer should be cleared after completion")
	}
}

// TestFederationPriceSynchronization tests economy price gossip protocol
// between two federated servers.
func TestFederationPriceSynchronization(t *testing.T) {
	server1, server2, cleanup := setupTestServerPair(t,
		"economy-server-1", "127.0.0.1:17783",
		"economy-server-2", "127.0.0.1:17784")
	defer cleanup()

	// Create price signal with market data
	priceSignal := &PriceSignal{
		ServerID: "economy-server-1",
		CityID:   "capital-city",
		PriceTable: map[string]float64{
			"iron_ore":      25.50,
			"gold_ore":      150.00,
			"health_potion": 10.00,
			"magic_scroll":  75.25,
		},
		Timestamp: time.Now(),
	}

	// Send price signal
	if err := server1.SendPriceSignal("economy-server-2", priceSignal); err != nil {
		t.Fatalf("SendPriceSignal failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify server2 received and processed the price signal
	remotePrices := server2.Federation.GetRemotePrices()
	if len(remotePrices) == 0 {
		t.Fatal("Server2 should have received price signals")
	}

	// Find our price signal
	var foundSignal *PriceSignal
	for _, ps := range remotePrices {
		if ps.ServerID == "economy-server-1" && ps.CityID == "capital-city" {
			foundSignal = ps
			break
		}
	}

	if foundSignal == nil {
		t.Fatal("Price signal from economy-server-1 not found")
	}

	if foundSignal.PriceTable["iron_ore"] != 25.50 {
		t.Errorf("iron_ore price = %f, want 25.50", foundSignal.PriceTable["iron_ore"])
	}
	if foundSignal.PriceTable["gold_ore"] != 150.00 {
		t.Errorf("gold_ore price = %f, want 150.00", foundSignal.PriceTable["gold_ore"])
	}
	if foundSignal.PriceTable["magic_scroll"] != 75.25 {
		t.Errorf("magic_scroll price = %f, want 75.25", foundSignal.PriceTable["magic_scroll"])
	}
}

// TestFederationGlobalEventBroadcast tests world event broadcasting
// between federated servers.
func TestFederationGlobalEventBroadcast(t *testing.T) {
	server1, server2, cleanup := setupTestServerPair(t,
		"event-server-1", "127.0.0.1:17785",
		"event-server-2", "127.0.0.1:17786")
	defer cleanup()

	// Create a global event
	event := &GlobalEvent{
		EventID:     "dragon-raid-001",
		EventType:   "dragon_attack",
		Description: "A great dragon descends upon the realm!",
		StartTime:   time.Now(),
		Duration:    30 * time.Minute,
	}
	event.AffectedArea.CenterX = 1000.0
	event.AffectedArea.CenterZ = 2000.0
	event.AffectedArea.Radius = 500.0

	// Broadcast event locally first
	server1.Federation.BroadcastEvent(event)

	// Send to peer
	if err := server1.SendGlobalEvent("event-server-2", event); err != nil {
		t.Fatalf("SendGlobalEvent failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify both servers have the event
	server1Events := server1.Federation.GetActiveEvents()
	if len(server1Events) == 0 {
		t.Fatal("Server1 should have the event locally")
	}

	server2Events := server2.Federation.GetActiveEvents()
	if len(server2Events) == 0 {
		t.Fatal("Server2 should have received the event")
	}

	// Validate event data on server2
	var foundEvent *GlobalEvent
	for _, e := range server2Events {
		if e.EventID == "dragon-raid-001" {
			foundEvent = e
			break
		}
	}

	if foundEvent == nil {
		t.Fatal("Dragon raid event not found on server2")
	}

	if foundEvent.EventType != "dragon_attack" {
		t.Errorf("EventType = %s, want dragon_attack", foundEvent.EventType)
	}
	if foundEvent.AffectedArea.CenterX != 1000.0 {
		t.Errorf("CenterX = %f, want 1000.0", foundEvent.AffectedArea.CenterX)
	}
	if foundEvent.AffectedArea.Radius != 500.0 {
		t.Errorf("Radius = %f, want 500.0", foundEvent.AffectedArea.Radius)
	}
}

// TestFederationMultiplePeers tests federation with 3+ servers forming a mesh.
func TestFederationMultiplePeers(t *testing.T) {
	server1 := NewFederationServer("mesh-1", "127.0.0.1:17787")
	server2 := NewFederationServer("mesh-2", "127.0.0.1:17788")
	server3 := NewFederationServer("mesh-3", "127.0.0.1:17789")

	// Start all servers
	for _, s := range []*FederationServer{server1, server2, server3} {
		if err := s.Start(); err != nil {
			t.Fatalf("Failed to start %s: %v", s.ID, err)
		}
		defer s.Stop()
	}

	time.Sleep(50 * time.Millisecond)

	// Create mesh: 1 -> 2, 1 -> 3, 2 -> 3
	if err := server1.ConnectToPeer("mesh-2", "127.0.0.1:17788"); err != nil {
		t.Fatalf("Failed to connect 1->2: %v", err)
	}
	if err := server1.ConnectToPeer("mesh-3", "127.0.0.1:17789"); err != nil {
		t.Fatalf("Failed to connect 1->3: %v", err)
	}
	if err := server2.ConnectToPeer("mesh-3", "127.0.0.1:17789"); err != nil {
		t.Fatalf("Failed to connect 2->3: %v", err)
	}

	// Verify node counts
	if server1.Federation.NodeCount() != 2 {
		t.Errorf("Server1 should have 2 peers, got %d", server1.Federation.NodeCount())
	}
	if server2.Federation.NodeCount() != 1 {
		t.Errorf("Server2 should have 1 peer, got %d", server2.Federation.NodeCount())
	}

	// Broadcast event from server1 to both peers
	event := &GlobalEvent{
		EventID:     "mesh-event-001",
		EventType:   "server_merge",
		Description: "Multiple servers are merging",
		StartTime:   time.Now(),
		Duration:    time.Hour,
	}

	server1.Federation.BroadcastEvent(event)
	if err := server1.SendGlobalEvent("mesh-2", event); err != nil {
		t.Fatalf("Failed to send event to mesh-2: %v", err)
	}
	if err := server1.SendGlobalEvent("mesh-3", event); err != nil {
		t.Fatalf("Failed to send event to mesh-3: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// All three servers should have the event
	for _, s := range []*FederationServer{server1, server2, server3} {
		events := s.Federation.GetActiveEvents()
		found := false
		for _, e := range events {
			if e.EventID == "mesh-event-001" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Server %s should have mesh-event-001", s.ID)
		}
	}
}

// TestFederationTransferEncoding verifies that transfers survive network encoding.
func TestFederationTransferEncoding(t *testing.T) {
	// Test with a transfer containing complex data
	original := &PlayerTransfer{
		PlayerID:      999,
		AccountID:     "encoding-test",
		DisplayName:   "EncodingTestPlayer™",
		DestX:         -100.5,
		DestY:         50.0,
		DestZ:         -200.75,
		DestAngle:     3.14159,
		HealthCurrent: 50.5,
		HealthMax:     100.0,
		Inventory: []string{
			"special_item_with_long_name",
			"another-item",
			"item_3",
		},
		SkillLevels: map[string]int{
			"skill_a": 100,
			"skill_b": 50,
			"skill_c": 25,
			"skill_d": 1,
		},
		SkillExperience: map[string]float64{
			"skill_a": 999999.99,
			"skill_b": 50000.5,
		},
		QuestFlags: map[string]map[string]bool{
			"quest_1": {"a": true, "b": false, "c": true},
			"quest_2": {"x": false, "y": false},
		},
		Standings: map[string]float64{
			"faction_positive": 100.0,
			"faction_negative": -100.0,
			"faction_neutral":  0.0,
		},
		TransferTime: time.Now(),
		SourceServer: "encoding-source",
	}

	// Simulate network transmission with pipe
	reader, writer := io.Pipe()
	errCh := make(chan error, 1)

	go func() {
		errCh <- EncodeTransfer(writer, original)
		writer.Close()
	}()

	decoded, err := DecodeTransfer(reader)
	if err != nil {
		t.Fatalf("DecodeTransfer failed: %v", err)
	}

	if encErr := <-errCh; encErr != nil {
		t.Fatalf("EncodeTransfer failed: %v", encErr)
	}

	// Comprehensive validation
	if decoded.PlayerID != original.PlayerID {
		t.Errorf("PlayerID mismatch: got %d, want %d", decoded.PlayerID, original.PlayerID)
	}
	if decoded.DisplayName != original.DisplayName {
		t.Errorf("DisplayName mismatch: got %s, want %s", decoded.DisplayName, original.DisplayName)
	}
	if decoded.DestX != original.DestX {
		t.Errorf("DestX mismatch: got %f, want %f", decoded.DestX, original.DestX)
	}
	if decoded.DestAngle != original.DestAngle {
		t.Errorf("DestAngle mismatch: got %f, want %f", decoded.DestAngle, original.DestAngle)
	}
	if len(decoded.Inventory) != len(original.Inventory) {
		t.Errorf("Inventory length mismatch: got %d, want %d", len(decoded.Inventory), len(original.Inventory))
	}
	if len(decoded.SkillLevels) != len(original.SkillLevels) {
		t.Errorf("SkillLevels length mismatch: got %d, want %d", len(decoded.SkillLevels), len(original.SkillLevels))
	}
	if decoded.SkillLevels["skill_a"] != 100 {
		t.Errorf("skill_a mismatch: got %d, want 100", decoded.SkillLevels["skill_a"])
	}
	if len(decoded.QuestFlags) != 2 {
		t.Errorf("QuestFlags length mismatch: got %d, want 2", len(decoded.QuestFlags))
	}
	if decoded.Standings["faction_negative"] != -100.0 {
		t.Errorf("faction_negative mismatch: got %f, want -100.0", decoded.Standings["faction_negative"])
	}
}

// TestFederationNetworkResilience tests transfer behavior under network issues.
func TestFederationNetworkResilience(t *testing.T) {
	server := NewFederationServer("resilience-server", "127.0.0.1:17790")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Test: Initiate transfer but never complete (simulate network failure)
	transfer := &PlayerTransfer{
		PlayerID:     777,
		AccountID:    "resilience-test",
		DisplayName:  "ResiliencePlayer",
		TransferTime: time.Now(),
		SourceServer: "remote-server",
	}

	if err := server.Federation.InitiateTransfer(transfer); err != nil {
		t.Fatalf("InitiateTransfer failed: %v", err)
	}

	// Transfer should be pending
	pending := server.Federation.GetPendingTransfer(777)
	if pending == nil {
		t.Fatal("Transfer should be pending")
	}

	// Test failed completion
	server.Federation.CompleteTransfer(777, false)
	if server.Federation.GetPendingTransfer(777) != nil {
		t.Error("Failed transfer should be cleared")
	}
}

// TestFederationGossipProtocolOrdering tests that newer price signals
// correctly override older ones.
func TestFederationGossipProtocolOrdering(t *testing.T) {
	fed := NewFederation("ordering-test")

	// Send old signal
	oldSignal := &PriceSignal{
		ServerID:   "remote",
		CityID:     "city",
		PriceTable: map[string]float64{"sword": 100},
		Timestamp:  time.Now().Add(-time.Hour),
	}
	fed.ProcessPriceSignal(oldSignal)

	// Send newer signal
	newSignal := &PriceSignal{
		ServerID:   "remote",
		CityID:     "city",
		PriceTable: map[string]float64{"sword": 150},
		Timestamp:  time.Now(),
	}
	fed.ProcessPriceSignal(newSignal)

	// Only newer signal should be stored
	prices := fed.GetRemotePrices()
	if len(prices) != 1 {
		t.Fatalf("Expected 1 price signal, got %d", len(prices))
	}

	if prices[0].PriceTable["sword"] != 150 {
		t.Errorf("Expected newer price 150, got %f", prices[0].PriceTable["sword"])
	}

	// Try to send even older signal - should be rejected
	olderSignal := &PriceSignal{
		ServerID:   "remote",
		CityID:     "city",
		PriceTable: map[string]float64{"sword": 50},
		Timestamp:  time.Now().Add(-2 * time.Hour),
	}
	fed.ProcessPriceSignal(olderSignal)

	prices = fed.GetRemotePrices()
	if prices[0].PriceTable["sword"] != 150 {
		t.Errorf("Old signal should be rejected, price should still be 150, got %f", prices[0].PriceTable["sword"])
	}
}

// BenchmarkTransferEncodeDecode benchmarks the transfer serialization performance.
func BenchmarkTransferEncodeDecode(b *testing.B) {
	transfer := &PlayerTransfer{
		PlayerID:        1,
		AccountID:       "bench",
		DisplayName:     "BenchPlayer",
		DestX:           100,
		DestY:           0,
		DestZ:           200,
		HealthCurrent:   100,
		HealthMax:       100,
		Inventory:       []string{"a", "b", "c", "d", "e"},
		SkillLevels:     map[string]int{"s1": 10, "s2": 20, "s3": 30},
		SkillExperience: map[string]float64{"s1": 1000, "s2": 2000},
		QuestFlags:      map[string]map[string]bool{"q1": {"done": true}},
		Standings:       map[string]float64{"f1": 50, "f2": -25},
		TransferTime:    time.Now(),
		SourceServer:    "source",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = EncodeTransfer(&buf, transfer)
		_, _ = DecodeTransfer(&buf)
	}
}
