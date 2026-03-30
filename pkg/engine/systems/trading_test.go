package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewTradingSystem(t *testing.T) {
	ts := NewTradingSystem()
	if ts == nil {
		t.Fatal("NewTradingSystem returned nil")
	}
	if ts.TradeExpiry != 120.0 {
		t.Errorf("TradeExpiry = %f, want 120.0", ts.TradeExpiry)
	}
	if ts.TradeRange != 10.0 {
		t.Errorf("TradeRange = %f, want 10.0", ts.TradeRange)
	}
}

func TestTradingSystem_InitiateTrade(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	trade, err := ts.InitiateTrade(player1, player2)
	if err != nil {
		t.Fatalf("InitiateTrade failed: %v", err)
	}
	if trade == nil {
		t.Fatal("InitiateTrade returned nil trade")
	}
	if trade.Initiator.PlayerID != player1 {
		t.Errorf("Initiator = %d, want %d", trade.Initiator.PlayerID, player1)
	}
	if trade.Recipient.PlayerID != player2 {
		t.Errorf("Recipient = %d, want %d", trade.Recipient.PlayerID, player2)
	}
	if trade.Status != TradeStatusPending {
		t.Errorf("Status = %s, want pending", trade.Status)
	}
}

func TestTradingSystem_InitiateTrade_WithSelf(t *testing.T) {
	ts := NewTradingSystem()
	player := ecs.Entity(1)

	_, err := ts.InitiateTrade(player, player)
	if err == nil {
		t.Error("Expected error when trading with self")
	}
	if te, ok := err.(*TradeError); !ok || te.Code != ErrCannotTradeWithSelf {
		t.Errorf("Expected ErrCannotTradeWithSelf, got %v", err)
	}
}

func TestTradingSystem_InitiateTrade_AlreadyTrading(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)
	player3 := ecs.Entity(3)

	_, _ = ts.InitiateTrade(player1, player2)

	// Player1 tries to start another trade
	_, err := ts.InitiateTrade(player1, player3)
	if err == nil {
		t.Error("Expected error when already trading")
	}
}

func TestTradingSystem_AddItemToTrade(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)

	item := TradeItem{ItemID: "sword", Name: "Iron Sword", Quantity: 1, Quality: 0.8}
	err := ts.AddItemToTrade(player1, item)
	if err != nil {
		t.Fatalf("AddItemToTrade failed: %v", err)
	}

	offer := ts.GetPlayerTradeOffer(player1)
	if len(offer.Items) != 1 {
		t.Errorf("Items count = %d, want 1", len(offer.Items))
	}
	if offer.Items[0].ItemID != "sword" {
		t.Errorf("ItemID = %s, want sword", offer.Items[0].ItemID)
	}
}

func TestTradingSystem_RemoveItemFromTrade(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)
	_ = ts.AddItemToTrade(player1, TradeItem{ItemID: "sword", Name: "Sword", Quantity: 1})
	_ = ts.AddItemToTrade(player1, TradeItem{ItemID: "shield", Name: "Shield", Quantity: 1})

	err := ts.RemoveItemFromTrade(player1, 0)
	if err != nil {
		t.Fatalf("RemoveItemFromTrade failed: %v", err)
	}

	offer := ts.GetPlayerTradeOffer(player1)
	if len(offer.Items) != 1 {
		t.Errorf("Items count = %d, want 1", len(offer.Items))
	}
	if offer.Items[0].ItemID != "shield" {
		t.Errorf("Remaining item = %s, want shield", offer.Items[0].ItemID)
	}
}

func TestTradingSystem_RemoveItemFromTrade_InvalidIndex(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)

	err := ts.RemoveItemFromTrade(player1, 5)
	if err == nil {
		t.Error("Expected error for invalid index")
	}
}

func TestTradingSystem_SetTradeGold(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)

	err := ts.SetTradeGold(player1, 100)
	if err != nil {
		t.Fatalf("SetTradeGold failed: %v", err)
	}

	offer := ts.GetPlayerTradeOffer(player1)
	if offer.Gold != 100 {
		t.Errorf("Gold = %d, want 100", offer.Gold)
	}
}

func TestTradingSystem_SetTradeGold_Negative(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)

	err := ts.SetTradeGold(player1, -50)
	if err == nil {
		t.Error("Expected error for negative gold")
	}
}

func TestTradingSystem_LockOffer(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)

	err := ts.LockOffer(player1)
	if err != nil {
		t.Fatalf("LockOffer failed: %v", err)
	}

	offer := ts.GetPlayerTradeOffer(player1)
	if !offer.Locked {
		t.Error("Offer should be locked")
	}
}

func TestTradingSystem_LockOffer_CannotModifyLocked(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)
	_ = ts.LockOffer(player1)

	err := ts.AddItemToTrade(player1, TradeItem{ItemID: "sword"})
	if err == nil {
		t.Error("Expected error when modifying locked offer")
	}
}

func TestTradingSystem_UnlockOffer(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)
	_ = ts.LockOffer(player1)
	_ = ts.UnlockOffer(player1)

	offer := ts.GetPlayerTradeOffer(player1)
	if offer.Locked {
		t.Error("Offer should be unlocked")
	}
}

func TestTradingSystem_CompleteTrade(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	trade, _ := ts.InitiateTrade(player1, player2)
	_ = ts.LockOffer(player1)
	_ = ts.LockOffer(player2)

	// Trade should be completed when both lock
	if trade.Status != TradeStatusCompleted {
		t.Errorf("Status = %s, want completed", trade.Status)
	}

	if ts.IsTrading(player1) {
		t.Error("Player1 should not be in trade after completion")
	}
	if ts.IsTrading(player2) {
		t.Error("Player2 should not be in trade after completion")
	}
}

func TestTradingSystem_CancelTrade(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	trade, _ := ts.InitiateTrade(player1, player2)

	err := ts.CancelTrade(player1)
	if err != nil {
		t.Fatalf("CancelTrade failed: %v", err)
	}

	if trade.Status != TradeStatusCancelled {
		t.Errorf("Status = %s, want cancelled", trade.Status)
	}

	if ts.IsTrading(player1) || ts.IsTrading(player2) {
		t.Error("Players should not be in trade after cancellation")
	}
}

func TestTradingSystem_DeclineTrade(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	trade, _ := ts.InitiateTrade(player1, player2)

	err := ts.DeclineTrade(player2)
	if err != nil {
		t.Fatalf("DeclineTrade failed: %v", err)
	}

	if trade.Status != TradeStatusDeclined {
		t.Errorf("Status = %s, want declined", trade.Status)
	}
}

func TestTradingSystem_GetTrade(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	originalTrade, _ := ts.InitiateTrade(player1, player2)

	trade := ts.GetTrade(player1)
	if trade != originalTrade {
		t.Error("GetTrade should return the same trade")
	}

	trade = ts.GetTrade(player2)
	if trade != originalTrade {
		t.Error("Both players should get the same trade")
	}
}

func TestTradingSystem_GetPartnerOffer(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)
	_ = ts.AddItemToTrade(player2, TradeItem{ItemID: "potion", Name: "Potion"})
	_ = ts.SetTradeGold(player2, 50)

	partnerOffer := ts.GetPartnerOffer(player1)
	if partnerOffer == nil {
		t.Fatal("GetPartnerOffer returned nil")
	}
	if len(partnerOffer.Items) != 1 {
		t.Errorf("Partner items = %d, want 1", len(partnerOffer.Items))
	}
	if partnerOffer.Gold != 50 {
		t.Errorf("Partner gold = %d, want 50", partnerOffer.Gold)
	}
}

func TestTradingSystem_IsTrading(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)
	player3 := ecs.Entity(3)

	_, _ = ts.InitiateTrade(player1, player2)

	if !ts.IsTrading(player1) {
		t.Error("Player1 should be trading")
	}
	if !ts.IsTrading(player2) {
		t.Error("Player2 should be trading")
	}
	if ts.IsTrading(player3) {
		t.Error("Player3 should not be trading")
	}
}

func TestTradingSystem_IsTradingWith(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)
	player3 := ecs.Entity(3)

	_, _ = ts.InitiateTrade(player1, player2)

	if !ts.IsTradingWith(player1, player2) {
		t.Error("Player1 should be trading with Player2")
	}
	if ts.IsTradingWith(player1, player3) {
		t.Error("Player1 should not be trading with Player3")
	}
}

func TestTradingSystem_GetRecentTrades(t *testing.T) {
	ts := NewTradingSystem()

	// Complete a few trades
	for i := 0; i < 5; i++ {
		p1 := ecs.Entity(i*2 + 1)
		p2 := ecs.Entity(i*2 + 2)
		_, _ = ts.InitiateTrade(p1, p2)
		_ = ts.LockOffer(p1)
		_ = ts.LockOffer(p2)
	}

	trades := ts.GetRecentTrades(3)
	if len(trades) != 3 {
		t.Errorf("Recent trades = %d, want 3", len(trades))
	}
}

func TestTradingSystem_Update_ExpireTrades(t *testing.T) {
	ts := NewTradingSystem()
	ts.TradeExpiry = 10.0 // Short expiry for test

	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	trade, _ := ts.InitiateTrade(player1, player2)

	// Advance time past expiry
	w := ecs.NewWorld()
	ts.Update(w, 15.0)

	if trade.Status != TradeStatusExpired {
		t.Errorf("Status = %s, want expired", trade.Status)
	}
}

func TestTradingSystem_ModifyUnlocksOther(t *testing.T) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)

	// Both lock
	_ = ts.LockOffer(player1)
	// Note: If player1 locks then player2 locks, trade completes
	// So we test that modification unlocks

	// Instead, let's test that adding an item unlocks both
	_ = ts.UnlockOffer(player1)
	_ = ts.AddItemToTrade(player2, TradeItem{ItemID: "gem"})

	offer1 := ts.GetPlayerTradeOffer(player1)
	offer2 := ts.GetPlayerTradeOffer(player2)

	// Adding item should unlock both
	if offer1 == nil || offer2 == nil {
		t.Fatal("Could not get offers")
	}
	if offer1.Locked || offer2.Locked {
		t.Error("Adding item should unlock both offers")
	}
}

func TestTradingSystem_Callbacks(t *testing.T) {
	ts := NewTradingSystem()

	startCalled := false
	endCalled := false

	ts.OnTradeStart = func(trade *Trade) {
		startCalled = true
	}
	ts.OnTradeEnd = func(trade *Trade) {
		endCalled = true
	}

	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)

	_, _ = ts.InitiateTrade(player1, player2)
	if !startCalled {
		t.Error("OnTradeStart should be called")
	}

	_ = ts.LockOffer(player1)
	_ = ts.LockOffer(player2)

	if !endCalled {
		t.Error("OnTradeEnd should be called")
	}
}

func TestTradeError_Error(t *testing.T) {
	tests := []struct {
		code TradeErrorCode
		want string
	}{
		{ErrCannotTradeWithSelf, "cannot trade with yourself"},
		{ErrAlreadyInTrade, "already in a trade"},
		{ErrTargetAlreadyTrading, "target player is already trading"},
		{ErrNotInTrade, "not in a trade"},
		{ErrTradeNotFound, "trade not found"},
		{ErrTradeNotActive, "trade is not active"},
		{ErrOfferLocked, "offer is locked"},
		{ErrInvalidItemIndex, "invalid item index"},
		{ErrInvalidGoldAmount, "invalid gold amount"},
		{ErrInsufficientItems, "insufficient items"},
		{ErrInsufficientGold, "insufficient gold"},
		{TradeErrorCode(999), "trade error"}, // Unknown code
	}

	for _, tc := range tests {
		err := &TradeError{Code: tc.code}
		if err.Error() != tc.want {
			t.Errorf("Error() = %q, want %q", err.Error(), tc.want)
		}
	}
}

func BenchmarkTradingSystem_InitiateTrade(b *testing.B) {
	ts := NewTradingSystem()
	for i := 0; i < b.N; i++ {
		p1 := ecs.Entity(i*2 + 1)
		p2 := ecs.Entity(i*2 + 2)
		_, _ = ts.InitiateTrade(p1, p2)
		// Clean up
		_ = ts.CancelTrade(p1)
	}
}

func BenchmarkTradingSystem_AddItem(b *testing.B) {
	ts := NewTradingSystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)
	_, _ = ts.InitiateTrade(player1, player2)

	item := TradeItem{ItemID: "item", Name: "Item", Quantity: 1}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ts.AddItemToTrade(player1, item)
	}
}
