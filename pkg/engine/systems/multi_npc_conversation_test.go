package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewMultiNPCConversationSystem(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	if system == nil {
		t.Fatal("NewMultiNPCConversationSystem returned nil")
	}
	if system.ActiveConversations == nil {
		t.Error("ActiveConversations should be initialized")
	}
}

func TestStartMultiNPCConversation(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	// Create NPCs
	npc1 := world.CreateEntity()
	world.AddComponent(npc1, &components.Position{X: 0, Y: 0})

	npc2 := world.CreateEntity()
	world.AddComponent(npc2, &components.Position{X: 5, Y: 0})

	npc3 := world.CreateEntity()
	world.AddComponent(npc3, &components.Position{X: 3, Y: 3})

	participants := []ecs.Entity{npc1, npc2, npc3}
	convID := system.StartMultiNPCConversation(world, participants, 0, "greeting")

	if convID == "" {
		t.Fatal("StartMultiNPCConversation should return a valid ID")
	}

	// Check conversation exists
	conv, ok := system.ActiveConversations[convID]
	if !ok {
		t.Fatal("Conversation should exist in ActiveConversations")
	}
	if len(conv.Participants) != 3 {
		t.Errorf("Participants = %d, want 3", len(conv.Participants))
	}
	if !conv.IsActive {
		t.Error("Conversation should be active")
	}
	if conv.HasPlayer {
		t.Error("Conversation should not have player")
	}

	// Check participants are in dialog
	for _, p := range participants {
		stateComp, ok := world.GetComponent(p, "DialogState")
		if !ok {
			t.Errorf("Participant %v should have DialogState", p)
			continue
		}
		state := stateComp.(*components.DialogState)
		if !state.IsInDialog {
			t.Errorf("Participant %v should be in dialog", p)
		}
	}
}

func TestStartMultiNPCConversationWithPlayer(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Position{X: 0, Y: 0})

	npc1 := world.CreateEntity()
	world.AddComponent(npc1, &components.Position{X: 5, Y: 0})

	npc2 := world.CreateEntity()
	world.AddComponent(npc2, &components.Position{X: 3, Y: 3})

	participants := []ecs.Entity{player, npc1, npc2}
	convID := system.StartMultiNPCConversation(world, participants, player, "greeting")

	if convID == "" {
		t.Fatal("Should create conversation with player")
	}

	conv := system.ActiveConversations[convID]
	if !conv.HasPlayer {
		t.Error("Conversation should have player")
	}
	if conv.PlayerEntity != player {
		t.Errorf("PlayerEntity = %v, want %v", conv.PlayerEntity, player)
	}
}

func TestStartMultiNPCConversationTooFewParticipants(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{npc}, 0, "greeting")

	if convID != "" {
		t.Error("Should not create conversation with only 1 participant")
	}
}

func TestMultiNPCEndConversation(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	npc1 := world.CreateEntity()
	world.AddComponent(npc1, &components.Position{X: 0, Y: 0})

	npc2 := world.CreateEntity()
	world.AddComponent(npc2, &components.Position{X: 5, Y: 0})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{npc1, npc2}, 0, "greeting")
	system.EndConversation(world, convID)

	// Check conversation removed
	_, ok := system.ActiveConversations[convID]
	if ok {
		t.Error("Conversation should be removed after ending")
	}

	// Check participants not in dialog
	for _, npc := range []ecs.Entity{npc1, npc2} {
		stateComp, ok := world.GetComponent(npc, "DialogState")
		if ok {
			state := stateComp.(*components.DialogState)
			if state.IsInDialog {
				t.Error("Participant should not be in dialog after ending")
			}
		}
	}
}

func TestPlayerSpeak(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Position{X: 0, Y: 0})

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 5, Y: 0})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{player, npc}, player, "greeting")

	// NPC speaks first in our implementation
	// Wait for player's turn
	conv := system.ActiveConversations[convID]
	conv.CurrentSpeaker = player

	result := system.PlayerSpeak(world, convID, "Hello everyone!", npc)
	if !result {
		t.Error("PlayerSpeak should succeed on player's turn")
	}

	// Check exchange recorded
	if len(conv.Exchanges) != 1 {
		t.Fatalf("Exchanges = %d, want 1", len(conv.Exchanges))
	}
	exchange := conv.Exchanges[0]
	if exchange.Speaker != player {
		t.Errorf("Speaker = %v, want %v", exchange.Speaker, player)
	}
	if exchange.Text != "Hello everyone!" {
		t.Errorf("Text = %v, want 'Hello everyone!'", exchange.Text)
	}
}

func TestPlayerSpeakNotTurn(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Position{X: 0, Y: 0})

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 5, Y: 0})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{player, npc}, player, "greeting")

	// It's NPC's turn initially
	conv := system.ActiveConversations[convID]
	conv.CurrentSpeaker = npc

	result := system.PlayerSpeak(world, convID, "Out of turn!", 0)
	if result {
		t.Error("PlayerSpeak should fail when not player's turn")
	}
}

func TestAddParticipant(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	npc1 := world.CreateEntity()
	world.AddComponent(npc1, &components.Position{X: 0, Y: 0})

	npc2 := world.CreateEntity()
	world.AddComponent(npc2, &components.Position{X: 5, Y: 0})

	npc3 := world.CreateEntity()
	world.AddComponent(npc3, &components.Position{X: 3, Y: 3})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{npc1, npc2}, 0, "greeting")

	result := system.AddParticipant(world, convID, npc3)
	if !result {
		t.Error("AddParticipant should succeed")
	}

	conv := system.ActiveConversations[convID]
	if len(conv.Participants) != 3 {
		t.Errorf("Participants = %d, want 3", len(conv.Participants))
	}

	// Check NPC3 is in dialog
	stateComp, _ := world.GetComponent(npc3, "DialogState")
	state := stateComp.(*components.DialogState)
	if !state.IsInDialog {
		t.Error("Added participant should be in dialog")
	}
}

func TestAddParticipantAlreadyInDialog(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	npc1 := world.CreateEntity()
	world.AddComponent(npc1, &components.Position{X: 0, Y: 0})

	npc2 := world.CreateEntity()
	world.AddComponent(npc2, &components.Position{X: 5, Y: 0})

	npc3 := world.CreateEntity()
	world.AddComponent(npc3, &components.Position{X: 10, Y: 0})
	world.AddComponent(npc3, &components.DialogState{IsInDialog: true})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{npc1, npc2}, 0, "greeting")

	result := system.AddParticipant(world, convID, npc3)
	if result {
		t.Error("Should not add participant who is already in dialog")
	}
}

func TestRemoveParticipant(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	npc1 := world.CreateEntity()
	world.AddComponent(npc1, &components.Position{X: 0, Y: 0})

	npc2 := world.CreateEntity()
	world.AddComponent(npc2, &components.Position{X: 5, Y: 0})

	npc3 := world.CreateEntity()
	world.AddComponent(npc3, &components.Position{X: 3, Y: 3})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{npc1, npc2, npc3}, 0, "greeting")

	result := system.RemoveParticipant(world, convID, npc3)
	if !result {
		t.Error("RemoveParticipant should succeed")
	}

	conv := system.ActiveConversations[convID]
	if len(conv.Participants) != 2 {
		t.Errorf("Participants = %d, want 2", len(conv.Participants))
	}

	// Check NPC3 is no longer in dialog
	stateComp, _ := world.GetComponent(npc3, "DialogState")
	state := stateComp.(*components.DialogState)
	if state.IsInDialog {
		t.Error("Removed participant should not be in dialog")
	}
}

func TestRemoveParticipantEndsConversation(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	npc1 := world.CreateEntity()
	world.AddComponent(npc1, &components.Position{X: 0, Y: 0})

	npc2 := world.CreateEntity()
	world.AddComponent(npc2, &components.Position{X: 5, Y: 0})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{npc1, npc2}, 0, "greeting")

	system.RemoveParticipant(world, convID, npc2)

	// Conversation should end with only 1 participant
	_, ok := system.ActiveConversations[convID]
	if ok {
		t.Error("Conversation should end when fewer than 2 participants")
	}
}

func TestIsPlayerTurn(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Position{X: 0, Y: 0})

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 5, Y: 0})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{player, npc}, player, "greeting")

	conv := system.ActiveConversations[convID]
	conv.CurrentSpeaker = player

	if !system.IsPlayerTurn(convID) {
		t.Error("Should be player's turn")
	}

	conv.CurrentSpeaker = npc
	if system.IsPlayerTurn(convID) {
		t.Error("Should not be player's turn")
	}
}

func TestGetCurrentSpeaker(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	npc1 := world.CreateEntity()
	world.AddComponent(npc1, &components.Position{X: 0, Y: 0})

	npc2 := world.CreateEntity()
	world.AddComponent(npc2, &components.Position{X: 5, Y: 0})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{npc1, npc2}, 0, "greeting")

	speaker := system.GetCurrentSpeaker(convID)
	if speaker != npc1 && speaker != npc2 {
		t.Errorf("Current speaker should be one of the participants")
	}
}

func TestGetConversationHistory(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Position{X: 0, Y: 0})

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 5, Y: 0})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{player, npc}, player, "greeting")

	// Add some exchanges
	conv := system.ActiveConversations[convID]
	conv.Exchanges = append(conv.Exchanges, MultiNPCExchange{
		Speaker: npc,
		Text:    "Hello!",
	})
	conv.CurrentSpeaker = player
	system.PlayerSpeak(world, convID, "Hi there!", npc)

	history := system.GetConversationHistory(convID)
	if len(history) != 2 {
		t.Errorf("History length = %d, want 2", len(history))
	}
}

func TestCheckParticipantsPresent(t *testing.T) {
	system := NewMultiNPCConversationSystem()
	world := ecs.NewWorld()

	npc1 := world.CreateEntity()
	world.AddComponent(npc1, &components.Position{X: 0, Y: 0})

	npc2 := world.CreateEntity()
	world.AddComponent(npc2, &components.Position{X: 5, Y: 0})

	convID := system.StartMultiNPCConversation(world, []ecs.Entity{npc1, npc2}, 0, "greeting")
	conv := system.ActiveConversations[convID]

	// Participants close together - should be present
	if !system.checkParticipantsPresent(world, conv) {
		t.Error("Participants should be present when close")
	}

	// Move NPC2 far away
	posComp, _ := world.GetComponent(npc2, "Position")
	pos := posComp.(*components.Position)
	pos.X = 100
	pos.Y = 100

	if system.checkParticipantsPresent(world, conv) {
		t.Error("Participants should not be present when far apart")
	}
}

func TestGetTopicResponse(t *testing.T) {
	tests := []struct {
		topic      string
		occupation string
		wantLen    int // Just check we get a non-empty response
	}{
		{"greeting", "merchant", 5},
		{"greeting", "guard", 5},
		{"weather", "farmer", 5},
		{"rumors", "innkeeper", 5},
		{"work", "blacksmith", 5},
		{"unknown", "merchant", 5},
	}

	for _, tt := range tests {
		t.Run(tt.topic+"_"+tt.occupation, func(t *testing.T) {
			response := getTopicResponse(tt.topic, tt.occupation)
			if len(response) < tt.wantLen {
				t.Errorf("Response too short: %v", response)
			}
		})
	}
}
