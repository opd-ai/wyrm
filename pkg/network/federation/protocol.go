// Package federation provides cross-server player travel and economy synchronization.
// Per ROADMAP Phase 5 item 24:
// AC: Player moves between two local test servers retaining inventory and quest state.
package federation

import (
	"encoding/gob"
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

// TransferAck acknowledges a player transfer.
type TransferAck struct {
	PlayerID  uint64
	Success   bool
	ErrorMsg  string
	Timestamp time.Time
}

// Node represents a peer server in the federation.
type Node struct {
	ServerID    string
	Address     string
	LastSeen    time.Time
	LatencyMs   int64
	PlayerCount int
}
