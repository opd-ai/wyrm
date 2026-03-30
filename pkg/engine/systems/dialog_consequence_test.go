package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewDialogConsequenceSystem(t *testing.T) {
	system := NewDialogConsequenceSystem()
	if system == nil {
		t.Fatal("NewDialogConsequenceSystem returned nil")
	}
	if system.PendingConsequences == nil {
		t.Error("PendingConsequences should be initialized")
	}
}

func TestDialogStateComponent(t *testing.T) {
	state := &components.DialogState{
		IsInDialog:          true,
		ConversationPartner: 123,
		CurrentTopicID:      "greeting",
	}

	if state.Type() != "DialogState" {
		t.Errorf("Type() = %v, want DialogState", state.Type())
	}
}

func TestDialogMemoryComponent(t *testing.T) {
	mem := &components.DialogMemory{
		ConversationCount: 5,
		Attitude:          0.3,
	}

	if mem.Type() != "DialogMemory" {
		t.Errorf("Type() = %v, want DialogMemory", mem.Type())
	}
}

func TestQueueConsequence(t *testing.T) {
	system := NewDialogConsequenceSystem()

	cons := components.DialogConsequences{
		ReputationChange: 0.1,
		FactionID:        "guards",
	}

	system.QueueConsequence(1, 2, "option1", cons)

	if len(system.PendingConsequences) != 1 {
		t.Fatalf("Queue length = %d, want 1", len(system.PendingConsequences))
	}

	pending := system.PendingConsequences[0]
	if pending.PlayerEntity != 1 {
		t.Errorf("PlayerEntity = %v, want 1", pending.PlayerEntity)
	}
	if pending.NPCEntity != 2 {
		t.Errorf("NPCEntity = %v, want 2", pending.NPCEntity)
	}
	if pending.OptionID != "option1" {
		t.Errorf("OptionID = %v, want option1", pending.OptionID)
	}
}

func TestStartConversation(t *testing.T) {
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Position{X: 0, Y: 0})

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 5, Y: 0})
	world.AddComponent(npc, &components.DialogMemory{
		TopicsDiscussed: make(map[string]int),
		KnownFacts:      make(map[string]bool),
	})

	// Start conversation
	result := StartConversation(world, player, npc)
	if !result {
		t.Error("StartConversation should return true")
	}

	// Check player state
	stateComp, ok := world.GetComponent(player, "DialogState")
	if !ok {
		t.Fatal("Player should have DialogState")
	}
	pState := stateComp.(*components.DialogState)
	if !pState.IsInDialog {
		t.Error("Player should be in dialog")
	}
	if pState.ConversationPartner != uint64(npc) {
		t.Errorf("ConversationPartner = %v, want %v", pState.ConversationPartner, npc)
	}

	// Check NPC state
	npcStateComp, ok := world.GetComponent(npc, "DialogState")
	if !ok {
		t.Fatal("NPC should have DialogState")
	}
	nState := npcStateComp.(*components.DialogState)
	if !nState.IsInDialog {
		t.Error("NPC should be in dialog")
	}

	// Check conversation count updated
	memComp, _ := world.GetComponent(npc, "DialogMemory")
	mem := memComp.(*components.DialogMemory)
	if mem.ConversationCount != 1 {
		t.Errorf("ConversationCount = %d, want 1", mem.ConversationCount)
	}
}

func TestStartConversationWhileInDialog(t *testing.T) {
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.DialogState{IsInDialog: true})

	npc := world.CreateEntity()

	// Should fail - player already in dialog
	result := StartConversation(world, player, npc)
	if result {
		t.Error("Should not start conversation when already in dialog")
	}
}

func TestEndConversation(t *testing.T) {
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.DialogState{
		IsInDialog:          true,
		ConversationPartner: 2,
	})

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.DialogState{
		IsInDialog:          true,
		ConversationPartner: uint64(player),
	})

	EndConversation(world, player, npc)

	// Check player state
	stateComp, _ := world.GetComponent(player, "DialogState")
	pState := stateComp.(*components.DialogState)
	if pState.IsInDialog {
		t.Error("Player should not be in dialog")
	}

	// Check NPC state
	npcStateComp, _ := world.GetComponent(npc, "DialogState")
	nState := npcStateComp.(*components.DialogState)
	if nState.IsInDialog {
		t.Error("NPC should not be in dialog")
	}
}

func TestCheckDialogRequirements(t *testing.T) {
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Faction{ID: "guards", Reputation: 0.5})
	world.AddComponent(player, &components.Inventory{
		Items:    []string{"gold_key", "sword"},
		Capacity: 10,
	})

	tests := []struct {
		name     string
		reqs     components.DialogRequirements
		expected bool
	}{
		{
			name:     "no_requirements",
			reqs:     components.DialogRequirements{},
			expected: true,
		},
		{
			name:     "reputation_met",
			reqs:     components.DialogRequirements{MinReputation: 0.3},
			expected: true,
		},
		{
			name:     "reputation_not_met",
			reqs:     components.DialogRequirements{MinReputation: 0.8},
			expected: false,
		},
		{
			name:     "item_present",
			reqs:     components.DialogRequirements{RequiredItems: []string{"gold_key"}},
			expected: true,
		},
		{
			name:     "item_missing",
			reqs:     components.DialogRequirements{RequiredItems: []string{"silver_key"}},
			expected: false,
		},
		{
			name:     "multiple_items_present",
			reqs:     components.DialogRequirements{RequiredItems: []string{"gold_key", "sword"}},
			expected: true,
		},
		{
			name:     "one_item_missing",
			reqs:     components.DialogRequirements{RequiredItems: []string{"gold_key", "shield"}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckDialogRequirements(world, player, tt.reqs)
			if result != tt.expected {
				t.Errorf("CheckDialogRequirements() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUpdateDialogOptions(t *testing.T) {
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Faction{ID: "guards", Reputation: 0.5})
	world.AddComponent(player, &components.Inventory{Items: []string{"gold"}, Capacity: 10})

	options := []components.DialogOption{
		{
			ID:           "opt1",
			Text:         "Hello",
			Requirements: components.DialogRequirements{}, // No requirements
		},
		{
			ID:           "opt2",
			Text:         "Give gold",
			Requirements: components.DialogRequirements{RequiredItems: []string{"gold"}},
		},
		{
			ID:           "opt3",
			Text:         "Give sword",
			Requirements: components.DialogRequirements{RequiredItems: []string{"sword"}},
		},
	}

	updated := UpdateDialogOptions(world, player, options)

	if len(updated) != 3 {
		t.Fatalf("Updated options length = %d, want 3", len(updated))
	}

	// All should be visible
	for i, opt := range updated {
		if !opt.IsVisible {
			t.Errorf("Option %d should be visible", i)
		}
	}

	// Check enabled status
	if !updated[0].IsEnabled {
		t.Error("Option 1 should be enabled (no requirements)")
	}
	if !updated[1].IsEnabled {
		t.Error("Option 2 should be enabled (has gold)")
	}
	if updated[2].IsEnabled {
		t.Error("Option 3 should not be enabled (no sword)")
	}
}

func TestSelectDialogOption(t *testing.T) {
	system := NewDialogConsequenceSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.DialogState{
		IsInDialog:          true,
		ConversationPartner: 2,
		CurrentTopicID:      "greeting",
		AvailableResponses: []components.DialogOption{
			{
				ID:          "friendly",
				Text:        "Hello friend!",
				NextTopicID: "friendly_response",
				IsVisible:   true,
				IsEnabled:   true,
				Consequences: components.DialogConsequences{
					RelationshipChange: 0.1,
				},
			},
			{
				ID:          "hostile",
				Text:        "Go away!",
				NextTopicID: "hostile_response",
				IsVisible:   true,
				IsEnabled:   true,
				Consequences: components.DialogConsequences{
					RelationshipChange: -0.2,
				},
			},
		},
		DialogHistory: make([]components.DialogExchange, 0),
	})

	npc := world.CreateEntity()

	// Select friendly option
	result := SelectDialogOption(world, system, player, "friendly")
	if !result {
		t.Error("SelectDialogOption should return true")
	}

	// Check consequence was queued
	if len(system.PendingConsequences) != 1 {
		t.Fatalf("Should have 1 pending consequence, got %d", len(system.PendingConsequences))
	}

	pending := system.PendingConsequences[0]
	if pending.OptionID != "friendly" {
		t.Errorf("OptionID = %v, want friendly", pending.OptionID)
	}

	// Check dialog history updated
	stateComp, _ := world.GetComponent(player, "DialogState")
	state := stateComp.(*components.DialogState)
	if len(state.DialogHistory) != 1 {
		t.Errorf("Dialog history length = %d, want 1", len(state.DialogHistory))
	}
	if state.CurrentTopicID != "friendly_response" {
		t.Errorf("CurrentTopicID = %v, want friendly_response", state.CurrentTopicID)
	}
	_ = npc
}

func TestSelectDialogOptionDisabled(t *testing.T) {
	system := NewDialogConsequenceSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.DialogState{
		IsInDialog:          true,
		ConversationPartner: 2,
		AvailableResponses: []components.DialogOption{
			{
				ID:        "locked",
				Text:      "Secret option",
				IsVisible: true,
				IsEnabled: false, // Disabled
			},
		},
	})

	// Should fail - option is disabled
	result := SelectDialogOption(world, system, player, "locked")
	if result {
		t.Error("Should not select disabled option")
	}
}

func TestApplyReputationChange(t *testing.T) {
	system := NewDialogConsequenceSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Faction{ID: "guards", Reputation: 0.5})

	npc := world.CreateEntity()

	cons := components.DialogConsequences{
		ReputationChange: 0.2,
		FactionID:        "guards",
	}

	system.QueueConsequence(player, npc, "test", cons)
	system.Update(world, 0)

	// Check reputation changed
	facComp, _ := world.GetComponent(player, "Faction")
	fac := facComp.(*components.Faction)
	if fac.Reputation < 0.69 || fac.Reputation > 0.71 {
		t.Errorf("Reputation = %v, want ~0.7", fac.Reputation)
	}
}

func TestApplyItemConsequences(t *testing.T) {
	system := NewDialogConsequenceSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()
	world.AddComponent(player, &components.Inventory{
		Items:    []string{"gold"},
		Capacity: 10,
	})

	npc := world.CreateEntity()

	cons := components.DialogConsequences{
		ItemsGiven: []string{"silver_key"},
		ItemsTaken: []string{"gold"},
	}

	system.QueueConsequence(player, npc, "test", cons)
	system.Update(world, 0)

	// Check inventory changed
	invComp, _ := world.GetComponent(player, "Inventory")
	inv := invComp.(*components.Inventory)

	hasGold := false
	hasSilverKey := false
	for _, item := range inv.Items {
		if item == "gold" {
			hasGold = true
		}
		if item == "silver_key" {
			hasSilverKey = true
		}
	}

	if hasGold {
		t.Error("Gold should have been taken")
	}
	if !hasSilverKey {
		t.Error("Silver key should have been given")
	}
}

func TestRelationshipChange(t *testing.T) {
	system := NewDialogConsequenceSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.DialogMemory{
		Attitude:        0.0,
		TopicsDiscussed: make(map[string]int),
		KnownFacts:      make(map[string]bool),
	})

	cons := components.DialogConsequences{
		RelationshipChange: 0.3,
	}

	system.QueueConsequence(player, npc, "test", cons)
	system.Update(world, 0)

	// Check attitude changed
	memComp, _ := world.GetComponent(npc, "DialogMemory")
	mem := memComp.(*components.DialogMemory)
	if mem.Attitude < 0.29 || mem.Attitude > 0.31 {
		t.Errorf("Attitude = %v, want ~0.3", mem.Attitude)
	}
}

func TestDialogMemoryEventRecording(t *testing.T) {
	system := NewDialogConsequenceSystem()
	world := ecs.NewWorld()

	player := world.CreateEntity()

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.DialogMemory{
		Attitude:        0.0,
		TopicsDiscussed: make(map[string]int),
		ImportantEvents: make([]components.DialogMemoryEvent, 0),
		KnownFacts:      make(map[string]bool),
	})

	cons := components.DialogConsequences{
		QuestStart: "find_artifact",
	}

	system.QueueConsequence(player, npc, "accept_quest", cons)
	system.Update(world, 0)

	// Check event was recorded
	memComp, _ := world.GetComponent(npc, "DialogMemory")
	mem := memComp.(*components.DialogMemory)

	if len(mem.ImportantEvents) != 1 {
		t.Fatalf("ImportantEvents length = %d, want 1", len(mem.ImportantEvents))
	}

	event := mem.ImportantEvents[0]
	if event.EventType != "quest_given" {
		t.Errorf("EventType = %v, want quest_given", event.EventType)
	}
}
