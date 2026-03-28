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
