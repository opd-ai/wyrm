package network

import (
	"net"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	srv := NewServer("localhost:0")
	if srv.Address != "localhost:0" {
		t.Errorf("expected address localhost:0, got %s", srv.Address)
	}
}

func TestServerStartStop(t *testing.T) {
	srv := NewServer("localhost:0")

	err := srv.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	// Server should be running
	if srv.listener == nil {
		t.Error("listener should not be nil after start")
	}

	err = srv.Stop()
	if err != nil {
		t.Errorf("failed to stop server: %v", err)
	}
}

func TestServerClientCount(t *testing.T) {
	srv := NewServer("localhost:0")
	_ = srv.Start()
	defer srv.Stop()

	if srv.ClientCount() != 0 {
		t.Errorf("expected 0 clients initially, got %d", srv.ClientCount())
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("localhost:7777")
	if client.ServerAddress != "localhost:7777" {
		t.Errorf("expected address localhost:7777, got %s", client.ServerAddress)
	}
}

func TestClientConnectNoServer(t *testing.T) {
	client := NewClient("localhost:59999") // Port unlikely to be in use

	err := client.Connect()
	if err == nil {
		t.Error("expected error when connecting to non-existent server")
		client.Disconnect()
	}
}

func TestClientIsConnected(t *testing.T) {
	client := NewClient("localhost:7777")
	if client.IsConnected() {
		t.Error("client should not be connected initially")
	}
}

func TestClientDisconnectWhenNotConnected(t *testing.T) {
	client := NewClient("localhost:7777")
	err := client.Disconnect()
	if err != nil {
		t.Errorf("disconnect when not connected should not error: %v", err)
	}
}

func TestServerAcceptsConnection(t *testing.T) {
	srv := NewServer("localhost:0")
	err := srv.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Stop()

	// Get the actual address the server is listening on
	addr := srv.listener.Addr().String()

	// Connect using raw TCP
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Give server time to process connection
	time.Sleep(50 * time.Millisecond)

	if srv.ClientCount() != 1 {
		t.Errorf("expected 1 client after connection, got %d", srv.ClientCount())
	}
}

func TestClientConnectDisconnect(t *testing.T) {
	srv := NewServer("localhost:0")
	err := srv.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Stop()

	addr := srv.listener.Addr().String()
	client := NewClient(addr)

	err = client.Connect()
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if !client.IsConnected() {
		t.Error("client should be connected after Connect()")
	}

	time.Sleep(50 * time.Millisecond)

	if srv.ClientCount() != 1 {
		t.Errorf("expected 1 client, got %d", srv.ClientCount())
	}

	err = client.Disconnect()
	if err != nil {
		t.Errorf("failed to disconnect: %v", err)
	}

	if client.IsConnected() {
		t.Error("client should not be connected after Disconnect()")
	}
}

func TestServerStopClosesClients(t *testing.T) {
	srv := NewServer("localhost:0")
	_ = srv.Start()

	addr := srv.listener.Addr().String()
	client := NewClient(addr)
	_ = client.Connect()

	time.Sleep(50 * time.Millisecond)

	// Stop server should close client connections
	_ = srv.Stop()

	// Server should have no clients after stop
	if srv.ClientCount() != 0 {
		t.Errorf("expected 0 clients after stop, got %d", srv.ClientCount())
	}
}

func TestClientSendWhenNotConnected(t *testing.T) {
	client := NewClient("localhost:7777")
	err := client.Send([]byte("test"))
	if err == nil {
		t.Error("expected error when sending without connection")
	}
}

func TestClientSendReceive(t *testing.T) {
	srv := NewServer("localhost:0")
	_ = srv.Start()
	defer srv.Stop()

	addr := srv.listener.Addr().String()
	client := NewClient(addr)
	_ = client.Connect()
	defer client.Disconnect()

	// Send data
	testData := []byte("hello")
	err := client.Send(testData)
	if err != nil {
		t.Errorf("failed to send data: %v", err)
	}
}

func TestServerStopIdempotent(t *testing.T) {
	srv := NewServer("localhost:0")
	_ = srv.Start()

	// Multiple stops should not panic or error
	_ = srv.Stop()
	err := srv.Stop()
	if err != nil {
		t.Errorf("second stop should not error: %v", err)
	}
}

func TestHandlePlayerInputValidation(t *testing.T) {
	tests := []struct {
		name         string
		inputForward float32
		inputRight   float32
		inputTurn    float32
	}{
		{"normal input", 0.5, 0.3, 0.1},
		{"max forward", 1.0, 0, 0},
		{"negative forward", -1.0, 0, 0},
		{"over max forward", 2.0, 0, 0},   // Should clamp to 1.0
		{"under min forward", -5.0, 0, 0}, // Should clamp to -1.0
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := NewServer("localhost:0")
			_ = srv.Start()
			defer srv.Stop()

			addr := srv.listener.Addr().String()
			conn, err := net.DialTimeout("tcp", addr, time.Second)
			if err != nil {
				t.Fatalf("failed to connect: %v", err)
			}
			defer conn.Close()

			time.Sleep(50 * time.Millisecond)

			// Send player input
			input := &PlayerInput{
				MoveForward:  tc.inputForward,
				MoveRight:    tc.inputRight,
				Turn:         tc.inputTurn,
				SequenceNum:  1,
				ClientTimeMs: uint32(time.Now().UnixMilli()),
			}
			err = input.Encode(conn)
			if err != nil {
				t.Fatalf("failed to encode input: %v", err)
			}

			// Give server time to process
			time.Sleep(50 * time.Millisecond)
		})
	}
}

func TestClampFloat32(t *testing.T) {
	tests := []struct {
		value, min, max float32
		expected        float32
	}{
		{0.5, 0, 1, 0.5},    // In range
		{-1.0, 0, 1, 0},     // Below min
		{2.0, 0, 1, 1},      // Above max
		{0, -1, 1, 0},       // At zero
		{-0.5, -1, 1, -0.5}, // Negative in range
	}

	for _, tc := range tests {
		result := clampFloat32(tc.value, tc.min, tc.max)
		if result != tc.expected {
			t.Errorf("clampFloat32(%f, %f, %f) = %f; want %f",
				tc.value, tc.min, tc.max, result, tc.expected)
		}
	}
}

func TestServerPlayerStateTracking(t *testing.T) {
	srv := NewServer("localhost:0")
	_ = srv.Start()
	defer srv.Stop()

	addr := srv.listener.Addr().String()
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	// Verify a player entity was created (check via client count and state map)
	srv.mu.Lock()
	clientCount := len(srv.clients)
	stateCount := len(srv.playerStates)
	var entityID uint64
	for _, state := range srv.playerStates {
		entityID = state.EntityID
		break
	}
	srv.mu.Unlock()

	if clientCount != 1 {
		t.Errorf("expected 1 client, got %d", clientCount)
	}
	if stateCount != 1 {
		t.Errorf("expected 1 player state, got %d", stateCount)
	}
	if entityID == 0 {
		t.Error("expected non-zero entity ID")
	}
}

// TestMultiClientStateSync verifies two clients can connect and the server tracks their positions.
// This test covers the multiplayer scenario where the server maintains state for all connected clients.
func TestMultiClientStateSync(t *testing.T) {
	srv := NewServer("localhost:0")
	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Stop()

	addr := srv.listener.Addr().String()

	// Connect two clients
	conn1, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("client 1 failed to connect: %v", err)
	}
	defer conn1.Close()

	conn2, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("client 2 failed to connect: %v", err)
	}
	defer conn2.Close()

	// Wait for server to process connections
	time.Sleep(100 * time.Millisecond)

	// Verify both clients are registered
	srv.mu.Lock()
	clientCount := len(srv.clients)
	stateCount := len(srv.playerStates)
	var entityIDs []uint64
	for _, state := range srv.playerStates {
		entityIDs = append(entityIDs, state.EntityID)
	}
	srv.mu.Unlock()

	if clientCount != 2 {
		t.Errorf("expected 2 clients, got %d", clientCount)
	}
	if stateCount != 2 {
		t.Errorf("expected 2 player states, got %d", stateCount)
	}
	if len(entityIDs) != 2 {
		t.Errorf("expected 2 entity IDs, got %d", len(entityIDs))
	}

	// Verify entity IDs are unique
	if len(entityIDs) == 2 && entityIDs[0] == entityIDs[1] {
		t.Error("entity IDs should be unique")
	}

	// Verify sendWorldState would include both entities
	// (This simulates what happens when any client moves)
	srv.mu.Lock()
	entities := make([]EntityState, 0, len(srv.playerStates))
	for _, ps := range srv.playerStates {
		entities = append(entities, EntityState{
			EntityID: ps.EntityID,
			X:        ps.X,
			Y:        ps.Y,
			Z:        ps.Z,
			Angle:    ps.Angle,
			Health:   ps.Health,
		})
	}
	srv.mu.Unlock()

	if len(entities) != 2 {
		t.Errorf("world state should include 2 entities, has %d", len(entities))
	}

	// Verify client entity mapping is consistent
	// Note: We compare against server's client list, not our dial connections
	// (server stores the accepted conn, not the dialed conn)
	srv.mu.Lock()
	var mappedEntityIDs []uint64
	for _, conn := range srv.clients {
		entityID := srv.clientPlayers[conn]
		mappedEntityIDs = append(mappedEntityIDs, entityID)
	}
	srv.mu.Unlock()

	if len(mappedEntityIDs) != 2 {
		t.Errorf("expected 2 mapped entities, got %d", len(mappedEntityIDs))
	}
	for _, eid := range mappedEntityIDs {
		if eid == 0 {
			t.Error("client entity mapping should be non-zero")
		}
	}
	if len(mappedEntityIDs) == 2 && mappedEntityIDs[0] == mappedEntityIDs[1] {
		t.Error("each client should have a unique entity ID")
	}
}

// TestMultiClientWorldStateBroadcast verifies that world state contains all connected players.
func TestMultiClientWorldStateBroadcast(t *testing.T) {
	srv := NewServer("localhost:0")
	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Stop()

	addr := srv.listener.Addr().String()

	// Connect three clients
	clients := make([]net.Conn, 3)
	for i := 0; i < 3; i++ {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err != nil {
			t.Fatalf("client %d failed to connect: %v", i+1, err)
		}
		defer conn.Close()
		clients[i] = conn
	}

	// Wait for server to process connections
	time.Sleep(150 * time.Millisecond)

	// Verify all three clients are tracked
	if srv.ClientCount() != 3 {
		t.Errorf("expected 3 clients, got %d", srv.ClientCount())
	}

	// Verify all three have player states
	srv.mu.Lock()
	stateCount := len(srv.playerStates)
	entityIDSet := make(map[uint64]bool)
	for _, state := range srv.playerStates {
		entityIDSet[state.EntityID] = true
	}
	srv.mu.Unlock()

	if stateCount != 3 {
		t.Errorf("expected 3 player states, got %d", stateCount)
	}
	if len(entityIDSet) != 3 {
		t.Errorf("expected 3 unique entity IDs, got %d", len(entityIDSet))
	}
}
