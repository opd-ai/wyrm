// Package systems provides ECS systems for game logic.
package systems

import (
	"sync"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// TradeStatus represents the current state of a trade.
type TradeStatus string

const (
	TradeStatusPending   TradeStatus = "pending"
	TradeStatusAccepted  TradeStatus = "accepted"
	TradeStatusDeclined  TradeStatus = "declined"
	TradeStatusCancelled TradeStatus = "cancelled"
	TradeStatusCompleted TradeStatus = "completed"
	TradeStatusExpired   TradeStatus = "expired"
)

// TradeItem represents an item being traded.
type TradeItem struct {
	ItemID   string
	Name     string
	Quantity int
	Quality  float64
}

// PlayerTradeOffer represents one side of a trade.
type PlayerTradeOffer struct {
	PlayerID ecs.Entity
	Items    []TradeItem
	Gold     int
	Locked   bool // Player has confirmed their offer
}

// Trade represents an active trade between two players.
type Trade struct {
	ID         string
	Initiator  *PlayerTradeOffer
	Recipient  *PlayerTradeOffer
	Status     TradeStatus
	Created    float64
	ExpiresAt  float64
	LastUpdate float64
}

// TradingSystem manages player-to-player item trading.
type TradingSystem struct {
	mu           sync.RWMutex
	Trades       map[string]*Trade
	PlayerTrades map[ecs.Entity]string // Player -> active trade ID
	TradeHistory []*Trade
	HistoryLimit int
	TradeExpiry  float64 // How long trades last before expiring
	TradeRange   float64 // Maximum distance between trading players
	GameTime     float64
	tradeCounter uint64
	OnTradeStart func(trade *Trade)
	OnTradeEnd   func(trade *Trade)
}

// NewTradingSystem creates a new trading management system.
func NewTradingSystem() *TradingSystem {
	return &TradingSystem{
		Trades:       make(map[string]*Trade),
		PlayerTrades: make(map[ecs.Entity]string),
		TradeHistory: make([]*Trade, 0),
		HistoryLimit: 100,
		TradeExpiry:  120.0, // 2 minutes to complete trade
		TradeRange:   10.0,  // 10 units range
	}
}

// Update processes trading system updates each tick.
func (s *TradingSystem) Update(w *ecs.World, dt float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GameTime += dt
	s.expireTrades()
}

// InitiateTrade starts a new trade between two players.
func (s *TradingSystem) InitiateTrade(initiatorID, recipientID ecs.Entity) (*Trade, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validateNewTrade(initiatorID, recipientID); err != nil {
		return nil, err
	}

	s.tradeCounter++
	tradeID := "trade_" + uint64ToString(s.tradeCounter)

	trade := &Trade{
		ID: tradeID,
		Initiator: &PlayerTradeOffer{
			PlayerID: initiatorID,
			Items:    make([]TradeItem, 0),
		},
		Recipient: &PlayerTradeOffer{
			PlayerID: recipientID,
			Items:    make([]TradeItem, 0),
		},
		Status:     TradeStatusPending,
		Created:    s.GameTime,
		ExpiresAt:  s.GameTime + s.TradeExpiry,
		LastUpdate: s.GameTime,
	}

	s.Trades[tradeID] = trade
	s.PlayerTrades[initiatorID] = tradeID
	s.PlayerTrades[recipientID] = tradeID

	if s.OnTradeStart != nil {
		s.OnTradeStart(trade)
	}

	return trade, nil
}

// AddItemToTrade adds an item to a player's trade offer.
func (s *TradingSystem) AddItemToTrade(playerID ecs.Entity, item TradeItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	trade, offer, err := s.getTradeAndOffer(playerID)
	if err != nil {
		return err
	}

	if offer.Locked {
		return &TradeError{Code: ErrOfferLocked}
	}

	offer.Items = append(offer.Items, item)
	trade.LastUpdate = s.GameTime
	s.unlockBothOffers(trade)

	return nil
}

// RemoveItemFromTrade removes an item from a player's trade offer.
func (s *TradingSystem) RemoveItemFromTrade(playerID ecs.Entity, itemIndex int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	trade, offer, err := s.getTradeAndOffer(playerID)
	if err != nil {
		return err
	}

	if offer.Locked {
		return &TradeError{Code: ErrOfferLocked}
	}

	if itemIndex < 0 || itemIndex >= len(offer.Items) {
		return &TradeError{Code: ErrInvalidItemIndex}
	}

	offer.Items = append(offer.Items[:itemIndex], offer.Items[itemIndex+1:]...)
	trade.LastUpdate = s.GameTime
	s.unlockBothOffers(trade)

	return nil
}

// SetTradeGold sets the gold amount in a player's trade offer.
func (s *TradingSystem) SetTradeGold(playerID ecs.Entity, gold int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	trade, offer, err := s.getTradeAndOffer(playerID)
	if err != nil {
		return err
	}

	if offer.Locked {
		return &TradeError{Code: ErrOfferLocked}
	}

	if gold < 0 {
		return &TradeError{Code: ErrInvalidGoldAmount}
	}

	offer.Gold = gold
	trade.LastUpdate = s.GameTime
	s.unlockBothOffers(trade)

	return nil
}

// LockOffer confirms a player's current offer.
func (s *TradingSystem) LockOffer(playerID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	trade, offer, err := s.getTradeAndOffer(playerID)
	if err != nil {
		return err
	}

	offer.Locked = true
	trade.LastUpdate = s.GameTime

	// Check if both offers are locked to complete trade
	if trade.Initiator.Locked && trade.Recipient.Locked {
		s.completeTrade(trade)
	}

	return nil
}

// UnlockOffer unlocks a player's offer for modification.
func (s *TradingSystem) UnlockOffer(playerID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	trade, offer, err := s.getTradeAndOffer(playerID)
	if err != nil {
		return err
	}

	offer.Locked = false
	trade.LastUpdate = s.GameTime

	return nil
}

// endTradeWithStatus terminates a trade with the given status.
// Caller must hold the write lock.
func (s *TradingSystem) endTradeWithStatus(playerID ecs.Entity, status TradeStatus) error {
	tradeID, ok := s.PlayerTrades[playerID]
	if !ok {
		return &TradeError{Code: ErrNotInTrade}
	}

	trade := s.Trades[tradeID]
	if trade == nil {
		delete(s.PlayerTrades, playerID)
		return &TradeError{Code: ErrTradeNotFound}
	}

	trade.Status = status
	s.endTrade(trade)
	return nil
}

// CancelTrade cancels an active trade.
func (s *TradingSystem) CancelTrade(playerID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.endTradeWithStatus(playerID, TradeStatusCancelled)
}

// AcceptTrade is an alias for LockOffer (both players locking = trade complete).
func (s *TradingSystem) AcceptTrade(playerID ecs.Entity) error {
	return s.LockOffer(playerID)
}

// DeclineTrade declines a pending trade (same as cancel for pending).
func (s *TradingSystem) DeclineTrade(playerID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.endTradeWithStatus(playerID, TradeStatusDeclined)
}

// GetTrade returns the active trade for a player.
func (s *TradingSystem) GetTrade(playerID ecs.Entity) *Trade {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tradeID, ok := s.PlayerTrades[playerID]
	if !ok {
		return nil
	}
	return s.Trades[tradeID]
}

// GetPlayerTradeOffer returns a player's current offer in their active trade.
func (s *TradingSystem) GetPlayerTradeOffer(playerID ecs.Entity) *PlayerTradeOffer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trade, offer, err := s.getTradeAndOfferLocked(playerID)
	if err != nil || trade == nil {
		return nil
	}
	return offer
}

// GetPartnerOffer returns the other player's offer in a trade.
func (s *TradingSystem) GetPartnerOffer(playerID ecs.Entity) *PlayerTradeOffer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tradeID, ok := s.PlayerTrades[playerID]
	if !ok {
		return nil
	}

	trade := s.Trades[tradeID]
	if trade == nil {
		return nil
	}

	if trade.Initiator.PlayerID == playerID {
		return trade.Recipient
	}
	return trade.Initiator
}

// IsTrading checks if a player is currently in a trade.
func (s *TradingSystem) IsTrading(playerID ecs.Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.PlayerTrades[playerID]
	return ok
}

// IsTradingWith checks if two players are trading with each other.
func (s *TradingSystem) IsTradingWith(playerA, playerB ecs.Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tradeA := s.PlayerTrades[playerA]
	tradeB := s.PlayerTrades[playerB]
	return tradeA != "" && tradeA == tradeB
}

// GetRecentTrades returns recent completed trades.
func (s *TradingSystem) GetRecentTrades(limit int) []*Trade {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.TradeHistory) {
		limit = len(s.TradeHistory)
	}

	result := make([]*Trade, limit)
	start := len(s.TradeHistory) - limit
	copy(result, s.TradeHistory[start:])
	return result
}

// validateNewTrade checks if a new trade can be started.
func (s *TradingSystem) validateNewTrade(initiatorID, recipientID ecs.Entity) error {
	if initiatorID == recipientID {
		return &TradeError{Code: ErrCannotTradeWithSelf}
	}

	if _, ok := s.PlayerTrades[initiatorID]; ok {
		return &TradeError{Code: ErrAlreadyInTrade}
	}

	if _, ok := s.PlayerTrades[recipientID]; ok {
		return &TradeError{Code: ErrTargetAlreadyTrading}
	}

	return nil
}

// getTradeAndOffer returns trade and offer for a player (requires lock).
func (s *TradingSystem) getTradeAndOffer(playerID ecs.Entity) (*Trade, *PlayerTradeOffer, error) {
	tradeID, ok := s.PlayerTrades[playerID]
	if !ok {
		return nil, nil, &TradeError{Code: ErrNotInTrade}
	}

	trade := s.Trades[tradeID]
	if trade == nil {
		delete(s.PlayerTrades, playerID)
		return nil, nil, &TradeError{Code: ErrTradeNotFound}
	}

	if trade.Status != TradeStatusPending {
		return nil, nil, &TradeError{Code: ErrTradeNotActive}
	}

	var offer *PlayerTradeOffer
	if trade.Initiator.PlayerID == playerID {
		offer = trade.Initiator
	} else {
		offer = trade.Recipient
	}

	return trade, offer, nil
}

// getTradeAndOfferLocked is getTradeAndOffer for already-locked contexts.
func (s *TradingSystem) getTradeAndOfferLocked(playerID ecs.Entity) (*Trade, *PlayerTradeOffer, error) {
	tradeID, ok := s.PlayerTrades[playerID]
	if !ok {
		return nil, nil, &TradeError{Code: ErrNotInTrade}
	}

	trade := s.Trades[tradeID]
	if trade == nil {
		return nil, nil, &TradeError{Code: ErrTradeNotFound}
	}

	var offer *PlayerTradeOffer
	if trade.Initiator.PlayerID == playerID {
		offer = trade.Initiator
	} else {
		offer = trade.Recipient
	}

	return trade, offer, nil
}

// unlockBothOffers unlocks both offers when trade terms change.
func (s *TradingSystem) unlockBothOffers(trade *Trade) {
	trade.Initiator.Locked = false
	trade.Recipient.Locked = false
}

// completeTrade finalizes a trade.
func (s *TradingSystem) completeTrade(trade *Trade) {
	trade.Status = TradeStatusCompleted
	s.endTrade(trade)
}

// endTrade cleans up after a trade ends.
func (s *TradingSystem) endTrade(trade *Trade) {
	delete(s.PlayerTrades, trade.Initiator.PlayerID)
	delete(s.PlayerTrades, trade.Recipient.PlayerID)
	delete(s.Trades, trade.ID)

	s.TradeHistory = append(s.TradeHistory, trade)
	if len(s.TradeHistory) > s.HistoryLimit {
		s.TradeHistory = s.TradeHistory[1:]
	}

	if s.OnTradeEnd != nil {
		s.OnTradeEnd(trade)
	}
}

// expireTrades marks expired trades.
func (s *TradingSystem) expireTrades() {
	for _, trade := range s.Trades {
		if trade.Status == TradeStatusPending && s.GameTime > trade.ExpiresAt {
			trade.Status = TradeStatusExpired
			s.endTrade(trade)
		}
	}
}

// TradeErrorCode represents specific trade errors.
type TradeErrorCode int

const (
	ErrCannotTradeWithSelf TradeErrorCode = iota
	ErrAlreadyInTrade
	ErrTargetAlreadyTrading
	ErrNotInTrade
	ErrTradeNotFound
	ErrTradeNotActive
	ErrOfferLocked
	ErrInvalidItemIndex
	ErrInvalidGoldAmount
	ErrInsufficientItems
	ErrInsufficientGold
)

// TradeError represents a trade-related error.
type TradeError struct {
	Code TradeErrorCode
}

// Error returns a human-readable message for the TradeError.
func (e *TradeError) Error() string {
	switch e.Code {
	case ErrCannotTradeWithSelf:
		return "cannot trade with yourself"
	case ErrAlreadyInTrade:
		return "already in a trade"
	case ErrTargetAlreadyTrading:
		return "target player is already trading"
	case ErrNotInTrade:
		return "not in a trade"
	case ErrTradeNotFound:
		return "trade not found"
	case ErrTradeNotActive:
		return "trade is not active"
	case ErrOfferLocked:
		return "offer is locked"
	case ErrInvalidItemIndex:
		return "invalid item index"
	case ErrInvalidGoldAmount:
		return "invalid gold amount"
	case ErrInsufficientItems:
		return "insufficient items"
	case ErrInsufficientGold:
		return "insufficient gold"
	default:
		return "trade error"
	}
}
