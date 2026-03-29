// Package network provides client-server networking.
package network

import (
	"log"
	"net"
	"sync"
	"time"
)

// Server handles incoming client connections and authoritative game state.
type Server struct {
	Address  string
	mu       sync.Mutex
	listener net.Listener
	clients  []net.Conn
	running  bool
}

// NewServer creates a new network server.
func NewServer(address string) *Server {
	return &Server{
		Address: address,
		clients: make([]net.Conn, 0),
	}
}

// Start begins listening for client connections.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.listener = ln
	s.running = true
	s.mu.Unlock()

	go s.acceptLoop()
	return nil
}

// acceptLoop continuously accepts incoming connections.
func (s *Server) acceptLoop() {
	for {
		s.mu.Lock()
		if !s.running {
			s.mu.Unlock()
			return
		}
		listener := s.listener
		s.mu.Unlock()

		if listener == nil {
			return
		}

		conn, err := listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()
			if !running {
				return
			}
			log.Printf("accept error: %v", err)
			continue
		}

		s.mu.Lock()
		s.clients = append(s.clients, conn)
		s.mu.Unlock()
		log.Printf("client connected: %s", conn.RemoteAddr())

		go s.handleClient(conn)
	}
}

// handleClient manages a single client connection.
func (s *Server) handleClient(conn net.Conn) {
	defer func() {
		conn.Close()
		s.removeClient(conn)
		log.Printf("client disconnected: %s", conn.RemoteAddr())
	}()

	for {
		msgType, err := ReadMessageType(conn)
		if err != nil {
			return
		}

		switch msgType {
		case MsgTypePlayerInput:
			input, err := DecodePlayerInput(conn)
			if err != nil {
				log.Printf("decode PlayerInput: %v", err)
				return
			}
			s.handlePlayerInput(conn, input)

		case MsgTypePing:
			ping, err := DecodePing(conn)
			if err != nil {
				log.Printf("decode Ping: %v", err)
				return
			}
			s.handlePing(conn, ping)

		default:
			log.Printf("unknown message type: %d", msgType)
		}
	}
}

// handlePlayerInput processes player input from a client.
func (s *Server) handlePlayerInput(conn net.Conn, input *PlayerInput) {
	// For now, just acknowledge receipt with a world state
	// Future: apply to player entity, validate, broadcast
	_ = input
}

// handlePing responds to a ping with a pong.
func (s *Server) handlePing(conn net.Conn, ping *Ping) {
	pong := &Pong{
		ClientTimeMs: ping.ClientTimeMs,
		ServerTimeMs: uint32(serverTimeMs()),
	}
	_ = pong.Encode(conn)
}

// serverTimeMs returns the current server time in milliseconds.
func serverTimeMs() int64 {
	return time.Now().UnixMilli()
}

// removeClient removes a client from the tracked connections.
func (s *Server) removeClient(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.clients {
		if c == conn {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return
		}
	}
}

// Stop closes the server listener and all client connections.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	for _, c := range s.clients {
		c.Close()
	}
	s.clients = nil
	if s.listener != nil {
		err := s.listener.Close()
		s.listener = nil
		return err
	}
	return nil
}

// ClientCount returns the number of connected clients.
func (s *Server) ClientCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}

// Client handles the connection to a game server.
type Client struct {
	ServerAddress string
	mu            sync.Mutex
	conn          net.Conn
}

// NewClient creates a new network client.
func NewClient(serverAddress string) *Client {
	return &Client{ServerAddress: serverAddress}
}

// Connect establishes a connection to the server.
func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.ServerAddress)
	if err != nil {
		return err
	}
	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = conn
	c.mu.Unlock()
	return nil
}

// Disconnect closes the connection to the server.
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// IsConnected returns whether the client is connected.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}

// Send sends data to the server.
func (c *Client) Send(data []byte) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return net.ErrClosed
	}
	_, err := conn.Write(data)
	return err
}
