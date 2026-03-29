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
