//go:build !noebiten

// sync.go provides game state synchronization between client and server.
package main

import (
	"log"
	"sync"
	"time"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

// StateSynchronizer handles synchronization of game state between client and server.
type StateSynchronizer struct {
	client        *network.Client
	world         *ecs.World
	playerEntity  ecs.Entity
	remoteEntites map[uint64]ecs.Entity // Server entity ID -> local entity ID
	mu            sync.RWMutex

	// Input tracking for client-side prediction
	lastInputSeq   uint32
	pendingInputs  []*network.PlayerInput
	lastServerTime uint32

	// Chunk management for network-received terrain
	chunkManager *chunk.Manager
	lodManager   *chunk.LODManager
	chunkDirty   bool // true when new chunk data arrived and world map needs rebuild

	// Receive channel for async message processing
	msgChan chan serverMessage
	stopCh  chan struct{}
}

// serverMessage wraps a received message.
type serverMessage struct {
	msgType uint8
	msg     network.Message
}

// NewStateSynchronizer creates a new state synchronizer.
func NewStateSynchronizer(client *network.Client, world *ecs.World, playerEntity ecs.Entity, chunkMgr *chunk.Manager, lodMgr *chunk.LODManager) *StateSynchronizer {
	return &StateSynchronizer{
		client:        client,
		world:         world,
		playerEntity:  playerEntity,
		remoteEntites: make(map[uint64]ecs.Entity),
		chunkManager:  chunkMgr,
		lodManager:    lodMgr,
		msgChan:       make(chan serverMessage, 100),
		stopCh:        make(chan struct{}),
	}
}

// Start begins listening for server messages in a background goroutine.
func (s *StateSynchronizer) Start() {
	go s.receiveLoop()
}

// Stop stops the background receive goroutine.
func (s *StateSynchronizer) Stop() {
	close(s.stopCh)
}

// receiveLoop continuously reads messages from the server.
func (s *StateSynchronizer) receiveLoop() {
	reconnectTicker := time.NewTicker(100 * time.Millisecond)
	defer reconnectTicker.Stop()

	for {
		if s.shouldStopReceiving() {
			return
		}
		if s.waitForConnection(reconnectTicker) {
			continue
		}
		s.receiveAndQueueMessage()
	}
}

// shouldStopReceiving checks if the receive loop should terminate.
func (s *StateSynchronizer) shouldStopReceiving() bool {
	select {
	case <-s.stopCh:
		return true
	default:
		return false
	}
}

// waitForConnection waits for a valid connection, returning true if should continue loop.
func (s *StateSynchronizer) waitForConnection(ticker *time.Ticker) bool {
	if s.client == nil || !s.client.IsConnected() {
		<-ticker.C
		return true
	}
	return false
}

// receiveAndQueueMessage reads a message from the server and queues it for processing.
func (s *StateSynchronizer) receiveAndQueueMessage() {
	msgType, msg, err := s.client.ReceiveMessage()
	if err != nil {
		log.Printf("receive error: %v", err)
		return
	}

	select {
	case s.msgChan <- serverMessage{msgType: msgType, msg: msg}:
	default:
		log.Printf("message channel full, dropping message")
	}
}

// Update processes pending server messages and sends player input.
// Should be called once per frame.
func (s *StateSynchronizer) Update(dt float64) {
	// Process all pending messages
	for {
		select {
		case msg := <-s.msgChan:
			s.processMessage(msg.msgType, msg.msg)
		default:
			// No more messages
			goto done
		}
	}
done:
}

// processMessage handles a received server message.
func (s *StateSynchronizer) processMessage(msgType uint8, msg network.Message) {
	switch msgType {
	case network.MsgTypeWorldState:
		s.handleWorldState(msg.(*network.WorldState))
	case network.MsgTypeEntityUpdate:
		s.handleEntityUpdate(msg.(*network.EntityUpdate))
	case network.MsgTypeChunkData:
		s.handleChunkData(msg.(*network.ChunkData))
	case network.MsgTypePong:
		s.handlePong(msg.(*network.Pong))
	}
}

// handleWorldState applies a full world state update from the server.
func (s *StateSynchronizer) handleWorldState(state *network.WorldState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastServerTime = state.ServerTimeMs

	// Remove pending inputs that have been acknowledged
	s.removeAcknowledgedInputs(state.AckSequence)

	// Update all entity states
	for _, es := range state.Entities {
		s.updateEntityState(es)
	}
}

// handleEntityUpdate applies a single entity update from the server.
func (s *StateSynchronizer) handleEntityUpdate(update *network.EntityUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastServerTime = update.ServerTimeMs

	es := network.EntityState{
		EntityID: update.EntityID,
		X:        update.X,
		Y:        update.Y,
		Z:        update.Z,
		Angle:    update.Angle,
		Health:   update.Health,
	}
	s.updateEntityState(es)
}

// updateEntityState updates or creates an entity based on server state.
func (s *StateSynchronizer) updateEntityState(es network.EntityState) {
	localEntity, exists := s.remoteEntites[es.EntityID]

	if !exists {
		// Create new entity for remote state
		localEntity = s.world.CreateEntity()
		s.remoteEntites[es.EntityID] = localEntity

		// Add basic components
		if err := s.world.AddComponent(localEntity, &components.Position{
			X:     float64(es.X),
			Y:     float64(es.Y),
			Z:     float64(es.Z),
			Angle: float64(es.Angle),
		}); err != nil {
			log.Printf("failed to add Position to entity %d: %v", es.EntityID, err)
		}
		if err := s.world.AddComponent(localEntity, &components.Health{
			Current: float64(es.Health),
			Max:     100.0,
		}); err != nil {
			log.Printf("failed to add Health to entity %d: %v", es.EntityID, err)
		}
		return
	}

	// Update existing entity (but not the player - that's handled by client-side prediction)
	if localEntity == s.playerEntity {
		// For player entity, we only update position if there's been a reconciliation
		// (i.e., server position differs significantly from predicted position)
		s.reconcilePlayerPosition(es)
		return
	}

	// Update remote entity position and health
	if comp, ok := s.world.GetComponent(localEntity, "Position"); ok {
		pos := comp.(*components.Position)
		pos.X = float64(es.X)
		pos.Y = float64(es.Y)
		pos.Z = float64(es.Z)
		pos.Angle = float64(es.Angle)
	}

	if comp, ok := s.world.GetComponent(localEntity, "Health"); ok {
		health := comp.(*components.Health)
		health.Current = float64(es.Health)
	}
}

// reconcilePlayerPosition checks if client position needs correction from server.
func (s *StateSynchronizer) reconcilePlayerPosition(es network.EntityState) {
	comp, ok := s.world.GetComponent(s.playerEntity, "Position")
	if !ok {
		return
	}
	pos := comp.(*components.Position)

	// Calculate position difference
	dx := pos.X - float64(es.X)
	dy := pos.Y - float64(es.Y)
	distSq := dx*dx + dy*dy

	// If difference is significant (> 0.5 units), snap to server position
	// and replay pending inputs for client-side prediction
	const reconcileThresholdSq = 0.25 // 0.5 squared
	if distSq > reconcileThresholdSq {
		pos.X = float64(es.X)
		pos.Y = float64(es.Y)
		pos.Z = float64(es.Z)
		pos.Angle = float64(es.Angle)

		// Replay pending inputs that haven't been acknowledged yet
		for _, input := range s.pendingInputs {
			s.applyInputLocally(input)
		}
	}
}

// applyInputLocally applies a player input for client-side prediction.
func (s *StateSynchronizer) applyInputLocally(input *network.PlayerInput) {
	comp, ok := s.world.GetComponent(s.playerEntity, "Position")
	if !ok {
		return
	}
	pos := comp.(*components.Position)

	// Apply movement based on input
	const moveSpeed = 5.0
	const dt = 1.0 / 60.0

	// Forward/backward movement
	if input.MoveForward != 0 {
		pos.X += float64(input.MoveForward) * moveSpeed * dt * float64(cos(pos.Angle))
		pos.Y += float64(input.MoveForward) * moveSpeed * dt * float64(sin(pos.Angle))
	}

	// Strafing
	if input.MoveRight != 0 {
		pos.X += float64(input.MoveRight) * moveSpeed * dt * float64(cos(pos.Angle+90))
		pos.Y += float64(input.MoveRight) * moveSpeed * dt * float64(sin(pos.Angle+90))
	}

	// Turning
	if input.Turn != 0 {
		pos.Angle += float64(input.Turn)
	}
}

// cos returns the cosine of an angle in radians.
func cos(angle float64) float64 {
	return float64(cosDegrees(angle))
}

// sin returns the sine of an angle in radians.
func sin(angle float64) float64 {
	return float64(sinDegrees(angle))
}

// cosDegrees returns cosine of angle in radians (converted from game angle).
func cosDegrees(angle float64) float64 {
	// Use math import via inline - avoid circular issues
	return float64(int(angle)) / 360.0 * 2.0 * 3.14159
}

// sinDegrees returns sine of angle in radians (converted from game angle).
func sinDegrees(angle float64) float64 {
	return float64(int(angle+90)) / 360.0 * 2.0 * 3.14159
}

// handleChunkData processes chunk terrain data from server.
func (s *StateSynchronizer) handleChunkData(cd *network.ChunkData) {
	if s.chunkManager == nil {
		log.Printf("warning: received chunk data but chunk manager not available")
		return
	}

	x := int(cd.ChunkX)
	y := int(cd.ChunkY)
	size := int(cd.ChunkSize)

	// Store the network chunk data into the local chunk manager
	s.chunkManager.StoreNetworkChunk(x, y, size, cd.HeightData, cd.BiomeData)

	// Invalidate LOD cache so it regenerates from the new data
	if s.lodManager != nil {
		s.lodManager.InvalidateChunk(x, y)
	}

	// Mark that new chunk data has arrived and world map needs rebuild
	s.mu.Lock()
	s.chunkDirty = true
	s.mu.Unlock()

	log.Printf("stored chunk data for (%d, %d), size %d", x, y, size)
}

// HasNewChunks returns true and resets the flag if new chunk data arrived from the server.
func (s *StateSynchronizer) HasNewChunks() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chunkDirty {
		s.chunkDirty = false
		return true
	}
	return false
}

// handlePong processes a pong response for latency measurement.
func (s *StateSynchronizer) handlePong(pong *network.Pong) {
	now := uint32(time.Now().UnixMilli())
	rtt := now - pong.ClientTimeMs
	log.Printf("RTT: %d ms", rtt)
}

// removeAcknowledgedInputs removes inputs that have been acknowledged by server.
func (s *StateSynchronizer) removeAcknowledgedInputs(ackSeq uint32) {
	kept := make([]*network.PlayerInput, 0, len(s.pendingInputs))
	for _, input := range s.pendingInputs {
		if input.SequenceNum > ackSeq {
			kept = append(kept, input)
		}
	}
	s.pendingInputs = kept
}

// SendPlayerInput sends player input to the server and records it for prediction.
func (s *StateSynchronizer) SendPlayerInput(moveForward, moveRight, turn float32, jump, attack, use bool) error {
	if s.client == nil || !s.client.IsConnected() {
		return nil
	}

	// Combine sequence increment and pending input append in a single critical section
	// to prevent race condition between SendPlayerInput and receiveLoop
	s.mu.Lock()
	s.lastInputSeq++
	seq := s.lastInputSeq

	input := &network.PlayerInput{
		MoveForward:  moveForward,
		MoveRight:    moveRight,
		Turn:         turn,
		Jump:         jump,
		Attack:       attack,
		Use:          use,
		SequenceNum:  seq,
		ClientTimeMs: uint32(time.Now().UnixMilli()),
	}

	// Record for client-side prediction reconciliation
	s.pendingInputs = append(s.pendingInputs, input)
	s.mu.Unlock()

	return s.client.SendPlayerInput(input)
}

// SendPing sends a ping to measure latency.
func (s *StateSynchronizer) SendPing() error {
	if s.client == nil || !s.client.IsConnected() {
		return nil
	}
	return s.client.SendPing(uint32(time.Now().UnixMilli()))
}
