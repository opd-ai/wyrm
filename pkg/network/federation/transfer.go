package federation

import (
	"encoding/gob"
	"fmt"
	"io"
	"time"
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
