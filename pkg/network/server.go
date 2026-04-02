// Package network provides client-server networking.
package network

import (
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Server handles incoming client connections and authoritative game state.
type Server struct {
	Address       string
	mu            sync.Mutex
	listener      net.Listener
	clients       []net.Conn
	clientPlayers map[net.Conn]uint64 // Connection to player entity ID
	playerStates  map[uint64]*PlayerState
	lastAck       map[net.Conn]uint32 // Last acknowledged input sequence per client
	clientChunks  map[net.Conn][2]int // Last known chunk position per client
	running       bool
	wg            sync.WaitGroup // Tracks all server goroutines for clean shutdown
}

// PlayerState stores authoritative player state.
type PlayerState struct {
	EntityID uint64
	X, Y, Z  float32
	Angle    float32
	Health   float32
}

// NewServer creates a new network server.
func NewServer(address string) *Server {
	return &Server{
		Address:       address,
		clients:       make([]net.Conn, 0),
		clientPlayers: make(map[net.Conn]uint64),
		playerStates:  make(map[uint64]*PlayerState),
		lastAck:       make(map[net.Conn]uint32),
		clientChunks:  make(map[net.Conn][2]int),
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

	s.wg.Add(1)
	go s.acceptLoop()
	return nil
}

// acceptLoop continuously accepts incoming connections.
func (s *Server) acceptLoop() {
	defer s.wg.Done()
	for {
		if s.shouldStopAccepting() {
			return
		}
		s.acceptConnection()
	}
}

// shouldStopAccepting checks if the accept loop should terminate.
func (s *Server) shouldStopAccepting() bool {
	return !s.isRunning() || s.getListener() == nil
}

// acceptConnection attempts to accept a single incoming connection.
func (s *Server) acceptConnection() {
	listener := s.getListener()
	if listener == nil {
		return
	}
	conn, err := listener.Accept()
	if err != nil {
		s.handleAcceptError(err)
		return
	}
	s.registerClient(conn)
	s.wg.Add(1)
	go s.handleClient(conn)
}

// handleAcceptError logs accept errors unless the server is shutting down.
func (s *Server) handleAcceptError(err error) {
	if s.isRunning() {
		log.Printf("accept error: %v", err)
	}
}

// isRunning checks if the server is still running.
func (s *Server) isRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// getListener returns the current listener.
func (s *Server) getListener() net.Listener {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener
}

// nextEntityID generates a unique entity ID for a new player.
var entityIDCounter uint64 = 1000

func nextEntityID() uint64 {
	return atomic.AddUint64(&entityIDCounter, 1)
}

// registerClient adds a new client connection and creates a player entity.
func (s *Server) registerClient(conn net.Conn) {
	entityID := nextEntityID()

	s.mu.Lock()
	s.clients = append(s.clients, conn)
	s.clientPlayers[conn] = entityID
	s.playerStates[entityID] = &PlayerState{
		EntityID: entityID,
		X:        0,
		Y:        0,
		Z:        0,
		Angle:    0,
		Health:   100,
	}
	s.lastAck[conn] = 0
	s.mu.Unlock()
	log.Printf("client connected: %s (entity %d)", conn.RemoteAddr(), entityID)
}

// handleClient manages a single client connection.
func (s *Server) handleClient(conn net.Conn) {
	defer s.wg.Done()
	defer s.cleanupClient(conn)

	for {
		if err := s.processClientMessage(conn); err != nil {
			return
		}
	}
}

// cleanupClient closes a connection and removes it from tracking.
func (s *Server) cleanupClient(conn net.Conn) {
	s.mu.Lock()
	entityID := s.clientPlayers[conn]
	delete(s.clientPlayers, conn)
	delete(s.playerStates, entityID)
	delete(s.lastAck, conn)
	delete(s.clientChunks, conn)
	s.mu.Unlock()

	conn.Close()
	s.removeClient(conn)
	log.Printf("client disconnected: %s (entity %d)", conn.RemoteAddr(), entityID)
}

// processClientMessage reads and handles a single message from a client.
func (s *Server) processClientMessage(conn net.Conn) error {
	msgType, err := ReadMessageType(conn)
	if err != nil {
		return err
	}
	return s.dispatchMessage(conn, msgType)
}

// dispatchMessage routes a message to the appropriate handler.
func (s *Server) dispatchMessage(conn net.Conn, msgType uint8) error {
	switch msgType {
	case MsgTypePlayerInput:
		return s.handlePlayerInputMessage(conn)
	case MsgTypePing:
		return s.handlePingMessage(conn)
	case MsgTypeSaveRequest:
		return s.handleSaveRequestMessage(conn)
	case MsgTypeLoadRequest:
		return s.handleLoadRequestMessage(conn)
	default:
		log.Printf("unknown message type: %d", msgType)
		return nil
	}
}

// handlePlayerInputMessage decodes and processes player input.
func (s *Server) handlePlayerInputMessage(conn net.Conn) error {
	input, err := DecodePlayerInput(conn)
	if err != nil {
		log.Printf("decode PlayerInput: %v", err)
		return err
	}
	s.handlePlayerInput(conn, input)
	return nil
}

// SaveHandler is a callback for save requests.
type SaveHandler func() error

// LoadHandler is a callback for load requests.
type LoadHandler func() error

var (
	saveHandler SaveHandler
	loadHandler LoadHandler
)

// SetSaveHandler registers a callback for save requests.
func SetSaveHandler(handler SaveHandler) {
	saveHandler = handler
}

// SetLoadHandler registers a callback for load requests.
func SetLoadHandler(handler LoadHandler) {
	loadHandler = handler
}

// handleSaveRequestMessage decodes and processes a save request.
func (s *Server) handleSaveRequestMessage(conn net.Conn) error {
	_, err := DecodeSaveRequest(conn)
	if err != nil {
		log.Printf("decode SaveRequest: %v", err)
		return err
	}

	// Trigger save via registered handler
	var resp *SaveResponse
	if saveHandler != nil {
		if err := saveHandler(); err != nil {
			resp = &SaveResponse{
				Success:      false,
				ServerTimeMs: uint32(serverTimeMs()),
				Message:      err.Error(),
			}
		} else {
			resp = &SaveResponse{
				Success:      true,
				ServerTimeMs: uint32(serverTimeMs()),
				Message:      "World saved successfully",
			}
		}
	} else {
		resp = &SaveResponse{
			Success:      false,
			ServerTimeMs: uint32(serverTimeMs()),
			Message:      "Save not available",
		}
	}

	if err := resp.Encode(conn); err != nil {
		s.disconnectOnError(conn, err, "handleSaveRequest")
	}
	return nil
}

// handleLoadRequestMessage decodes and processes a load request.
func (s *Server) handleLoadRequestMessage(conn net.Conn) error {
	_, err := DecodeLoadRequest(conn)
	if err != nil {
		log.Printf("decode LoadRequest: %v", err)
		return err
	}

	// Trigger load via registered handler
	var resp *LoadResponse
	if loadHandler != nil {
		if err := loadHandler(); err != nil {
			resp = &LoadResponse{
				Success:      false,
				ServerTimeMs: uint32(serverTimeMs()),
				Message:      err.Error(),
			}
		} else {
			resp = &LoadResponse{
				Success:      true,
				ServerTimeMs: uint32(serverTimeMs()),
				Message:      "World loaded successfully",
			}
		}
	} else {
		resp = &LoadResponse{
			Success:      false,
			ServerTimeMs: uint32(serverTimeMs()),
			Message:      "Load not available",
		}
	}

	if err := resp.Encode(conn); err != nil {
		s.disconnectOnError(conn, err, "handleLoadRequest")
	}
	return nil
}

// handlePingMessage decodes and responds to a ping.
func (s *Server) handlePingMessage(conn net.Conn) error {
	ping, err := DecodePing(conn)
	if err != nil {
		log.Printf("decode Ping: %v", err)
		return err
	}
	s.handlePing(conn, ping)
	return nil
}

// handlePlayerInput processes player input from a client.
func (s *Server) handlePlayerInput(conn net.Conn, input *PlayerInput) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entityID, ok := s.clientPlayers[conn]
	if !ok {
		return
	}
	state, ok := s.playerStates[entityID]
	if !ok {
		return
	}

	// Validate and clamp input values
	input.MoveForward = clampFloat32(input.MoveForward, -1.0, 1.0)
	input.MoveRight = clampFloat32(input.MoveRight, -1.0, 1.0)
	input.Turn = clampFloat32(input.Turn, -3.14159, 3.14159)

	// Apply movement to player state (simplified physics)
	const moveSpeed = 5.0
	const turnSpeed = 1.0

	state.Angle += input.Turn * turnSpeed
	state.X += input.MoveForward * moveSpeed * float32(cos64(float64(state.Angle)))
	state.Y += input.MoveForward * moveSpeed * float32(sin64(float64(state.Angle)))
	state.X += input.MoveRight * moveSpeed * float32(cos64(float64(state.Angle)+1.5708))
	state.Y += input.MoveRight * moveSpeed * float32(sin64(float64(state.Angle)+1.5708))

	// Update last acknowledged sequence
	s.lastAck[conn] = input.SequenceNum

	// Send acknowledgement with world state (unlocks after this line via defer)
	// Note: sendWorldState is called after unlock via goroutine to avoid holding lock
	go s.sendWorldState(conn, entityID, input.SequenceNum)
}

// clampFloat32 clamps a value between min and max.
func clampFloat32(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// cos64 is a helper for math.Cos with float64.
func cos64(x float64) float64 {
	return math.Cos(x)
}

// sin64 is a helper for math.Sin with float64.
func sin64(x float64) float64 {
	return math.Sin(x)
}

// sendWorldState sends the current world state to a client.
func (s *Server) sendWorldState(conn net.Conn, playerID uint64, ackSeq uint32) {
	s.mu.Lock()
	entities := make([]EntityState, 0, len(s.playerStates))
	for _, ps := range s.playerStates {
		entities = append(entities, EntityState{
			EntityID: ps.EntityID,
			X:        ps.X,
			Y:        ps.Y,
			Z:        ps.Z,
			Angle:    ps.Angle,
			Health:   ps.Health,
		})
	}
	s.mu.Unlock()

	state := &WorldState{
		ServerTimeMs: uint32(serverTimeMs()),
		AckSequence:  ackSeq,
		Entities:     entities,
	}
	if err := state.Encode(conn); err != nil {
		s.disconnectOnError(conn, err, "SendWorldState")
	}
}

// SendEntityUpdate sends a single entity's state update to a client.
func (s *Server) SendEntityUpdate(conn net.Conn, entityID uint64, x, y, z, angle, health, velocity float32, state uint8) {
	update := &EntityUpdate{
		ServerTimeMs: uint32(serverTimeMs()),
		EntityID:     entityID,
		X:            x,
		Y:            y,
		Z:            z,
		Angle:        angle,
		Health:       health,
		Velocity:     velocity,
		State:        state,
	}
	if err := update.Encode(conn); err != nil {
		s.disconnectOnError(conn, err, "SendEntityUpdate")
	}
}

// BroadcastEntityUpdate sends an entity update to all connected clients.
func (s *Server) BroadcastEntityUpdate(entityID uint64, x, y, z, angle, health, velocity float32, state uint8) {
	s.mu.Lock()
	clients := make([]net.Conn, len(s.clients))
	copy(clients, s.clients)
	s.mu.Unlock()

	for _, conn := range clients {
		s.SendEntityUpdate(conn, entityID, x, y, z, angle, health, velocity, state)
	}
}

// SendChunkData sends chunk terrain data to a client.
func (s *Server) SendChunkData(conn net.Conn, chunkX, chunkY int32, chunkSize uint16, heightData []uint16, biomeData []uint8) {
	chunk := &ChunkData{
		ServerTimeMs: uint32(serverTimeMs()),
		ChunkX:       chunkX,
		ChunkY:       chunkY,
		ChunkSize:    chunkSize,
		HeightData:   heightData,
		BiomeData:    biomeData,
	}
	if err := chunk.Encode(conn); err != nil {
		s.disconnectOnError(conn, err, "SendChunkData")
	}
}

// handlePing responds to a ping with a pong.
func (s *Server) handlePing(conn net.Conn, ping *Ping) {
	pong := &Pong{
		ClientTimeMs: ping.ClientTimeMs,
		ServerTimeMs: uint32(serverTimeMs()),
	}
	if err := pong.Encode(conn); err != nil {
		s.disconnectOnError(conn, err, "handlePing")
	}
}

// serverTimeMs returns the current server time in milliseconds.
func serverTimeMs() int64 {
	return time.Now().UnixMilli()
}

// UpdateClientChunkPosition checks if a client has moved to a new chunk
// and sends chunk data if so. Returns true if chunk data was sent.
// chunkSize is the size of each chunk in world units.
func (s *Server) UpdateClientChunkPosition(conn net.Conn, x, y float32, chunkSize int) bool {
	newChunkX := int(x) / chunkSize
	newChunkY := int(y) / chunkSize

	s.mu.Lock()
	oldChunk, exists := s.clientChunks[conn]
	if exists && oldChunk[0] == newChunkX && oldChunk[1] == newChunkY {
		s.mu.Unlock()
		return false // Same chunk, no update needed
	}
	s.clientChunks[conn] = [2]int{newChunkX, newChunkY}
	s.mu.Unlock()

	return true // Chunk changed, caller should send chunk data
}

// GetClientChunk returns the client's current chunk coordinates.
func (s *Server) GetClientChunk(conn net.Conn) (int, int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	chunk, exists := s.clientChunks[conn]
	if !exists {
		return 0, 0, false
	}
	return chunk[0], chunk[1], true
}

// GetConnectedClients returns a copy of the connected clients slice.
func (s *Server) GetConnectedClients() []net.Conn {
	s.mu.Lock()
	defer s.mu.Unlock()
	clients := make([]net.Conn, len(s.clients))
	copy(clients, s.clients)
	return clients
}

// GetPlayerState returns the state for a player entity, if it exists.
func (s *Server) GetPlayerState(entityID uint64) (*PlayerState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, exists := s.playerStates[entityID]
	return state, exists
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

// disconnectOnError closes the connection and removes the client if an encode error occurs.
// Returns true if an error occurred and the client was disconnected.
func (s *Server) disconnectOnError(conn net.Conn, err error, context string) bool {
	if err == nil {
		return false
	}
	log.Printf("encode error (%s) for %s: %v, disconnecting client", context, conn.RemoteAddr(), err)
	conn.Close()
	s.removeClient(conn)
	return true
}

// Stop closes the server listener and all client connections.
// Waits for all server goroutines to finish before returning.
func (s *Server) Stop() error {
	s.mu.Lock()
	s.running = false
	// Close listener first to unblock Accept() in acceptLoop
	var listenerErr error
	if s.listener != nil {
		listenerErr = s.listener.Close()
		s.listener = nil
	}
	// Close all client connections to unblock handleClient goroutines
	for _, c := range s.clients {
		c.Close()
	}
	s.clients = nil
	s.mu.Unlock()

	// Wait for all goroutines to finish
	s.wg.Wait()
	return listenerErr
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

// SendPlayerInput sends player input to the server.
func (c *Client) SendPlayerInput(input *PlayerInput) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return net.ErrClosed
	}
	return input.Encode(conn)
}

// ReceiveMessage reads and returns the next message from the server.
// Returns the message type and the decoded message, or an error.
func (c *Client) ReceiveMessage() (uint8, Message, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return 0, nil, net.ErrClosed
	}

	msgType, err := ReadMessageType(conn)
	if err != nil {
		return 0, nil, fmt.Errorf("read message type: %w", err)
	}

	switch msgType {
	case MsgTypeWorldState:
		msg, err := DecodeWorldState(conn)
		return msgType, msg, err
	case MsgTypeEntityUpdate:
		msg, err := DecodeEntityUpdate(conn)
		return msgType, msg, err
	case MsgTypeChunkData:
		msg, err := DecodeChunkData(conn)
		return msgType, msg, err
	case MsgTypePong:
		msg, err := DecodePong(conn)
		return msgType, msg, err
	default:
		return msgType, nil, fmt.Errorf("unknown message type: %d", msgType)
	}
}

// SendPing sends a ping message to measure latency.
func (c *Client) SendPing(clientTimeMs uint32) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return net.ErrClosed
	}
	ping := &Ping{ClientTimeMs: clientTimeMs}
	return ping.Encode(conn)
}

// SendSaveRequest sends a save request to the server.
func (c *Client) SendSaveRequest() error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return net.ErrClosed
	}
	req := &SaveRequest{ClientTimeMs: uint32(time.Now().UnixMilli())}
	return req.Encode(conn)
}

// SendLoadRequest sends a load request to the server.
func (c *Client) SendLoadRequest() error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return net.ErrClosed
	}
	req := &LoadRequest{ClientTimeMs: uint32(time.Now().UnixMilli())}
	return req.Encode(conn)
}
