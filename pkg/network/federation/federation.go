// Package federation provides cross-server player travel and economy synchronization.
// Per ROADMAP Phase 5 item 24:
// AC: Player moves between two local test servers retaining inventory and quest state.
package federation

import (
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"
)

func init() {
	// Register types for gob encoding
	gob.Register(&PlayerTransfer{})
	gob.Register(&PriceSignal{})
	gob.Register(&GlobalEvent{})
}

// Message type identifiers for federation protocol.
const (
	MsgTypeTransfer    uint8 = 100
	MsgTypePriceSignal uint8 = 101
	MsgTypeGlobalEvent uint8 = 102
	MsgTypeTransferAck uint8 = 103
)

// PlayerTransfer contains all state needed to move a player between servers.
type PlayerTransfer struct {
	// Player identification
	PlayerID    uint64
	AccountID   string
	DisplayName string

	// Position on destination server
	DestX, DestY, DestZ float64
	DestAngle           float64

	// Health state
	HealthCurrent float64
	HealthMax     float64

	// Inventory (item IDs)
	Inventory []string

	// Skills
	SkillLevels     map[string]int
	SkillExperience map[string]float64

	// Quest state (QuestID -> flags)
	QuestFlags map[string]map[string]bool

	// Faction standings
	Standings map[string]float64

	// Timestamp for validation
	TransferTime time.Time
	SourceServer string
}

// Validate checks if the transfer data is valid.
func (t *PlayerTransfer) Validate() error {
	if t.PlayerID == 0 {
		return fmt.Errorf("player ID required")
	}
	if t.AccountID == "" {
		return fmt.Errorf("account ID required")
	}
	if t.SourceServer == "" {
		return fmt.Errorf("source server required")
	}
	// Transfer should be recent (within 5 minutes)
	if time.Since(t.TransferTime) > 5*time.Minute {
		return fmt.Errorf("transfer expired")
	}
	return nil
}

// PriceSignal broadcasts economy price updates across federation.
type PriceSignal struct {
	ServerID   string
	CityID     string
	PriceTable map[string]float64
	Timestamp  time.Time
}

// GlobalEvent represents a world event broadcast to all servers.
type GlobalEvent struct {
	EventID     string
	EventType   string
	Description string
	AffectedArea struct {
		CenterX, CenterZ float64
		Radius           float64
	}
	StartTime time.Time
	Duration  time.Duration
}

// TransferAck acknowledges a player transfer.
type TransferAck struct {
	PlayerID  uint64
	Success   bool
	ErrorMsg  string
	Timestamp time.Time
}

// FederationNode represents a peer server in the federation.
type FederationNode struct {
	ServerID    string
	Address     string
	LastSeen    time.Time
	LatencyMs   int64
	PlayerCount int
}

// Federation manages cross-server communication.
type Federation struct {
	mu sync.RWMutex

	localServerID string
	nodes         map[string]*FederationNode

	// Transfer tracking
	pendingTransfers map[uint64]*PlayerTransfer
	completedTransfers map[uint64]time.Time

	// Price aggregation from other servers
	remotePrices map[string]*PriceSignal

	// Event log
	activeEvents map[string]*GlobalEvent
}

// NewFederation creates a new federation manager.
func NewFederation(localServerID string) *Federation {
	return &Federation{
		localServerID:      localServerID,
		nodes:              make(map[string]*FederationNode),
		pendingTransfers:   make(map[uint64]*PlayerTransfer),
		completedTransfers: make(map[uint64]time.Time),
		remotePrices:       make(map[string]*PriceSignal),
		activeEvents:       make(map[string]*GlobalEvent),
	}
}

// RegisterNode adds a peer server to the federation.
func (f *Federation) RegisterNode(node *FederationNode) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nodes[node.ServerID] = node
}

// UnregisterNode removes a peer server.
func (f *Federation) UnregisterNode(serverID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.nodes, serverID)
}

// GetNode returns a federation node by ID.
func (f *Federation) GetNode(serverID string) *FederationNode {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.nodes[serverID]
}

// NodeCount returns the number of registered peer servers.
func (f *Federation) NodeCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.nodes)
}

// InitiateTransfer starts a player transfer to another server.
// Returns the transfer data to be sent to the destination.
func (f *Federation) InitiateTransfer(transfer *PlayerTransfer) error {
	// Set source server and timestamp before validation
	transfer.SourceServer = f.localServerID
	transfer.TransferTime = time.Now()

	if err := transfer.Validate(); err != nil {
		return fmt.Errorf("invalid transfer: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if transfer already pending
	if _, exists := f.pendingTransfers[transfer.PlayerID]; exists {
		return fmt.Errorf("transfer already pending for player %d", transfer.PlayerID)
	}

	f.pendingTransfers[transfer.PlayerID] = transfer

	return nil
}

// AcceptTransfer processes an incoming player transfer.
// Returns the transfer data for creating the player on this server.
func (f *Federation) AcceptTransfer(transfer *PlayerTransfer) (*PlayerTransfer, error) {
	if err := transfer.Validate(); err != nil {
		return nil, fmt.Errorf("invalid transfer: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Check for duplicate
	if completedTime, exists := f.completedTransfers[transfer.PlayerID]; exists {
		if time.Since(completedTime) < 5*time.Minute {
			return nil, fmt.Errorf("duplicate transfer for player %d", transfer.PlayerID)
		}
	}

	// Mark as completed
	f.completedTransfers[transfer.PlayerID] = time.Now()

	return transfer, nil
}

// CompleteTransfer marks a transfer as complete.
func (f *Federation) CompleteTransfer(playerID uint64, success bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.pendingTransfers, playerID)
	if success {
		f.completedTransfers[playerID] = time.Now()
	}
}

// GetPendingTransfer returns a pending transfer if one exists.
func (f *Federation) GetPendingTransfer(playerID uint64) *PlayerTransfer {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.pendingTransfers[playerID]
}

// ProcessPriceSignal updates remote price information.
func (f *Federation) ProcessPriceSignal(signal *PriceSignal) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := signal.ServerID + ":" + signal.CityID
	existing := f.remotePrices[key]
	
	// Only accept newer signals
	if existing == nil || signal.Timestamp.After(existing.Timestamp) {
		f.remotePrices[key] = signal
	}
}

// GetRemotePrices returns price signals from other servers.
func (f *Federation) GetRemotePrices() []*PriceSignal {
	f.mu.RLock()
	defer f.mu.RUnlock()

	signals := make([]*PriceSignal, 0, len(f.remotePrices))
	for _, s := range f.remotePrices {
		signals = append(signals, s)
	}
	return signals
}

// BroadcastEvent registers a global event.
func (f *Federation) BroadcastEvent(event *GlobalEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activeEvents[event.EventID] = event
}

// GetActiveEvents returns currently active global events.
func (f *Federation) GetActiveEvents() []*GlobalEvent {
	f.mu.RLock()
	defer f.mu.RUnlock()

	now := time.Now()
	events := make([]*GlobalEvent, 0)
	for _, e := range f.activeEvents {
		if now.Sub(e.StartTime) < e.Duration {
			events = append(events, e)
		}
	}
	return events
}

// cleanupPendingTransfers removes pending transfers older than 5 minutes.
func (f *Federation) cleanupPendingTransfers(now time.Time) {
	for id, transfer := range f.pendingTransfers {
		if now.Sub(transfer.TransferTime) > 5*time.Minute {
			delete(f.pendingTransfers, id)
		}
	}
}

// cleanupCompletedTransfers removes completed transfers older than 1 hour.
func (f *Federation) cleanupCompletedTransfers(now time.Time) {
	for id, completedTime := range f.completedTransfers {
		if now.Sub(completedTime) > time.Hour {
			delete(f.completedTransfers, id)
		}
	}
}

// cleanupExpiredEvents removes events that have exceeded their duration.
func (f *Federation) cleanupExpiredEvents(now time.Time) {
	for id, event := range f.activeEvents {
		if now.Sub(event.StartTime) >= event.Duration {
			delete(f.activeEvents, id)
		}
	}
}

// cleanupStalePriceSignals removes price signals older than 5 minutes.
func (f *Federation) cleanupStalePriceSignals(now time.Time) {
	for key, signal := range f.remotePrices {
		if now.Sub(signal.Timestamp) > 5*time.Minute {
			delete(f.remotePrices, key)
		}
	}
}

// CleanupExpired removes expired transfers and events.
func (f *Federation) CleanupExpired() {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	f.cleanupPendingTransfers(now)
	f.cleanupCompletedTransfers(now)
	f.cleanupExpiredEvents(now)
	f.cleanupStalePriceSignals(now)
}

// EncodeTransfer writes a player transfer to a writer.
func EncodeTransfer(w io.Writer, transfer *PlayerTransfer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(transfer)
}

// DecodeTransfer reads a player transfer from a reader.
func DecodeTransfer(r io.Reader) (*PlayerTransfer, error) {
	dec := gob.NewDecoder(r)
	transfer := &PlayerTransfer{}
	if err := dec.Decode(transfer); err != nil {
		return nil, err
	}
	return transfer, nil
}

// EncodePriceSignal writes a price signal to a writer.
func EncodePriceSignal(w io.Writer, signal *PriceSignal) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(signal)
}

// DecodePriceSignal reads a price signal from a reader.
func DecodePriceSignal(r io.Reader) (*PriceSignal, error) {
	dec := gob.NewDecoder(r)
	signal := &PriceSignal{}
	if err := dec.Decode(signal); err != nil {
		return nil, err
	}
	return signal, nil
}

// EncodeGlobalEvent writes a global event to a writer.
func EncodeGlobalEvent(w io.Writer, event *GlobalEvent) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(event)
}

// DecodeGlobalEvent reads a global event from a reader.
func DecodeGlobalEvent(r io.Reader) (*GlobalEvent, error) {
	dec := gob.NewDecoder(r)
	event := &GlobalEvent{}
	if err := dec.Decode(event); err != nil {
		return nil, err
	}
	return event, nil
}
