// Package network provides client-server networking.
package network

import "net"

// Server handles incoming client connections and authoritative game state.
type Server struct {
	Address  string
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
	s.listener = ln
	return nil
}

// Stop closes the server listener.
func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Client handles the connection to a game server.
type Client struct {
	ServerAddress string
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
	c.conn = conn
	return nil
}

// Disconnect closes the connection to the server.
func (c *Client) Disconnect() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
