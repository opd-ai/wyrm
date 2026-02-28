// Package network provides client-server networking.
package network

import (
	"net"
	"sync"
)

// Server handles incoming client connections and authoritative game state.
type Server struct {
	Address  string
	mu       sync.Mutex
	listener net.Listener
}

// NewServer creates a new network server.
func NewServer(address string) *Server {
	return &Server{Address: address}
}

// Start begins listening for client connections.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()
	return nil
}

// Stop closes the server listener.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		err := s.listener.Close()
		s.listener = nil
		return err
	}
	return nil
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
