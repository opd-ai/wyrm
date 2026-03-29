// Package federation provides cross-server player travel and economy synchronization.
// Per ROADMAP Phase 5 item 24:
// AC: Player moves between two local test servers retaining inventory and quest state.
//
// The package is organized into focused files:
//   - federation.go: Core Federation struct and node management
//   - types.go: Shared types and message constants
//   - transfer.go: Player transfer logic
//   - gossip.go: Price signals and global events
package federation

import (
	"sync"
	"time"
)

// Federation manages cross-server communication.
type Federation struct {
	mu sync.RWMutex

	localServerID string
	nodes         map[string]*FederationNode

	// Transfer tracking
	pendingTransfers   map[uint64]*PlayerTransfer
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
