// Package systems implements ECS system behaviors.
package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// DialogConsequenceSystem processes dialog choices and applies their effects.
type DialogConsequenceSystem struct {
	// PendingConsequences queues consequences to process.
	PendingConsequences []PendingConsequence
}

// PendingConsequence tracks a consequence waiting to be applied.
type PendingConsequence struct {
	// PlayerEntity is the player who made the choice.
	PlayerEntity ecs.Entity
	// NPCEntity is the NPC involved in the dialog.
	NPCEntity ecs.Entity
	// Consequences are the effects to apply.
	Consequences components.DialogConsequences
	// OptionID is the dialog option that triggered this.
	OptionID string
}

// NewDialogConsequenceSystem creates a new dialog consequence system.
func NewDialogConsequenceSystem() *DialogConsequenceSystem {
	return &DialogConsequenceSystem{
		PendingConsequences: make([]PendingConsequence, 0),
	}
}

// Update processes pending dialog consequences.
func (s *DialogConsequenceSystem) Update(w *ecs.World, dt float64) {
	// Process all pending consequences
	for len(s.PendingConsequences) > 0 {
		pending := s.PendingConsequences[0]
		s.PendingConsequences = s.PendingConsequences[1:]
		s.applyConsequences(w, pending)
	}

	// Check for broken promises and update relationships
	s.checkPromises(w, dt)
}

// applyConsequences applies a single set of dialog consequences.
func (s *DialogConsequenceSystem) applyConsequences(w *ecs.World, pending PendingConsequence) {
	cons := pending.Consequences

	// Apply reputation change
	if cons.ReputationChange != 0 && cons.FactionID != "" {
		s.applyReputationChange(w, pending.PlayerEntity, cons.FactionID, cons.ReputationChange)
	}

	// Apply gold change
	if cons.GoldChange != 0 {
		s.applyGoldChange(w, pending.PlayerEntity, cons.GoldChange)
	}

	// Give items to player
	for _, itemID := range cons.ItemsGiven {
		s.giveItem(w, pending.PlayerEntity, itemID)
	}

	// Take items from player
	for _, itemID := range cons.ItemsTaken {
		s.takeItem(w, pending.PlayerEntity, itemID)
	}

	// Start quest
	if cons.QuestStart != "" {
		s.startQuest(w, pending.PlayerEntity, cons.QuestStart)
	}

	// Progress quest
	if cons.QuestProgress != "" {
		s.progressQuest(w, pending.PlayerEntity, cons.QuestProgress)
	}

	// Complete quest
	if cons.QuestComplete != "" {
		s.completeQuest(w, pending.PlayerEntity, cons.QuestComplete)
	}

	// Set world flag
	if cons.FlagSet != "" {
		s.setWorldFlag(w, cons.FlagSet, true)
	}

	// Clear world flag
	if cons.FlagClear != "" {
		s.setWorldFlag(w, cons.FlagClear, false)
	}

	// Apply relationship change
	if cons.RelationshipChange != 0 {
		s.applyRelationshipChange(w, pending.NPCEntity, pending.PlayerEntity, cons.RelationshipChange)
	}

	// Update NPC mood
	if cons.NPCMood != "" {
		s.setNPCMood(w, pending.NPCEntity, cons.NPCMood)
	}

	// Trigger combat
	if cons.TriggerCombat {
		s.triggerCombat(w, pending.PlayerEntity, pending.NPCEntity)
	}

	// Record in NPC's dialog memory
	s.recordDialogEvent(w, pending.NPCEntity, pending.PlayerEntity, pending.OptionID, cons)
}

// applyReputationChange modifies faction reputation.
func (s *DialogConsequenceSystem) applyReputationChange(w *ecs.World, player ecs.Entity, factionID string, change float64) {
	facComp, ok := w.GetComponent(player, "Faction")
	if !ok {
		return
	}
	fac := facComp.(*components.Faction)
	if fac.ID == factionID {
		fac.Reputation += change
		if fac.Reputation > 1.0 {
			fac.Reputation = 1.0
		}
		if fac.Reputation < -1.0 {
			fac.Reputation = -1.0
		}
	}
}

// applyGoldChange modifies player gold (stub - needs currency component).
func (s *DialogConsequenceSystem) applyGoldChange(w *ecs.World, player ecs.Entity, change float64) {
	// Gold tracking would need a Currency component or extended Inventory
	// For now, this is a placeholder
	_ = player
	_ = change
}

// giveItem adds an item to player inventory.
func (s *DialogConsequenceSystem) giveItem(w *ecs.World, player ecs.Entity, itemID string) {
	invComp, ok := w.GetComponent(player, "Inventory")
	if !ok {
		return
	}
	inv := invComp.(*components.Inventory)
	if len(inv.Items) < inv.Capacity {
		inv.Items = append(inv.Items, itemID)
	}
}

// takeItem removes an item from player inventory.
func (s *DialogConsequenceSystem) takeItem(w *ecs.World, player ecs.Entity, itemID string) {
	invComp, ok := w.GetComponent(player, "Inventory")
	if !ok {
		return
	}
	inv := invComp.(*components.Inventory)
	for i, item := range inv.Items {
		if item == itemID {
			inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
			return
		}
	}
}

// startQuest begins a new quest (stub - needs quest system integration).
func (s *DialogConsequenceSystem) startQuest(w *ecs.World, player ecs.Entity, questID string) {
	// Quest system integration would go here
	_ = player
	_ = questID
}

// progressQuest advances a quest stage (stub).
func (s *DialogConsequenceSystem) progressQuest(w *ecs.World, player ecs.Entity, questProgress string) {
	// Quest system integration would go here
	_ = player
	_ = questProgress
}

// completeQuest finishes a quest (stub).
func (s *DialogConsequenceSystem) completeQuest(w *ecs.World, player ecs.Entity, questID string) {
	// Quest system integration would go here
	_ = player
	_ = questID
}

// setWorldFlag sets or clears a global flag.
func (s *DialogConsequenceSystem) setWorldFlag(w *ecs.World, flag string, value bool) {
	// World flag system integration would go here
	// This would typically use a WorldState component on a singleton entity
	_ = flag
	_ = value
}

// applyRelationshipChange modifies NPC relationship with player.
func (s *DialogConsequenceSystem) applyRelationshipChange(w *ecs.World, npc, player ecs.Entity, change float64) {
	memComp, ok := w.GetComponent(npc, "DialogMemory")
	if !ok {
		// Create dialog memory if it doesn't exist
		mem := &components.DialogMemory{
			TopicsDiscussed: make(map[string]int),
			ImportantEvents: make([]components.DialogMemoryEvent, 0),
			KnownFacts:      make(map[string]bool),
			GiftsReceived:   make([]string, 0),
			PromisesMade:    make([]components.DialogPromise, 0),
		}
		w.AddComponent(npc, mem)
		memComp, _ = w.GetComponent(npc, "DialogMemory")
	}
	mem := memComp.(*components.DialogMemory)
	mem.Attitude += change
	if mem.Attitude > 1.0 {
		mem.Attitude = 1.0
	}
	if mem.Attitude < -1.0 {
		mem.Attitude = -1.0
	}
}

// setNPCMood changes the NPC's emotional state (stub - needs mood component).
func (s *DialogConsequenceSystem) setNPCMood(w *ecs.World, npc ecs.Entity, mood string) {
	// NPC mood system integration would go here
	_ = npc
	_ = mood
}

// triggerCombat initiates combat between player and NPC (stub).
func (s *DialogConsequenceSystem) triggerCombat(w *ecs.World, player, npc ecs.Entity) {
	// Combat system integration would go here
	_ = player
	_ = npc
}

// recordDialogEvent adds an event to NPC's dialog memory.
func (s *DialogConsequenceSystem) recordDialogEvent(w *ecs.World, npc, player ecs.Entity, optionID string, cons components.DialogConsequences) {
	memComp, ok := w.GetComponent(npc, "DialogMemory")
	if !ok {
		return
	}
	mem := memComp.(*components.DialogMemory)

	// Determine event type and sentiment based on consequences
	eventType := "conversation"
	sentiment := 0.0

	if cons.QuestStart != "" {
		eventType = "quest_given"
		sentiment = 0.3
	}
	if cons.QuestComplete != "" {
		eventType = "quest_completed"
		sentiment = 0.5
	}
	if cons.TriggerCombat {
		eventType = "combat_started"
		sentiment = -1.0
	}
	if cons.RelationshipChange > 0 {
		eventType = "positive_interaction"
		sentiment = cons.RelationshipChange
	}
	if cons.RelationshipChange < 0 {
		eventType = "negative_interaction"
		sentiment = cons.RelationshipChange
	}

	event := components.DialogMemoryEvent{
		EventType:   eventType,
		Description: optionID,
		Sentiment:   sentiment,
	}
	mem.ImportantEvents = append(mem.ImportantEvents, event)

	// Keep memory bounded
	if len(mem.ImportantEvents) > DialogMemoryMaxEvents {
		mem.ImportantEvents = mem.ImportantEvents[1:]
	}
}

// checkPromises checks for broken promises and updates relationships.
func (s *DialogConsequenceSystem) checkPromises(w *ecs.World, dt float64) {
	entities := w.Entities("DialogMemory")
	for _, e := range entities {
		memComp, ok := w.GetComponent(e, "DialogMemory")
		if !ok {
			continue
		}
		mem := memComp.(*components.DialogMemory)

		// Check each promise
		for i := range mem.PromisesMade {
			promise := &mem.PromisesMade[i]
			if promise.IsFulfilled || promise.IsBroken {
				continue
			}
			// If promise has deadline and it's passed, mark as broken
			if promise.Deadline > 0 {
				// Would need world time to check this properly
				// For now, this is structural
			}
		}
	}
}

// QueueConsequence adds a consequence to be processed.
func (s *DialogConsequenceSystem) QueueConsequence(player, npc ecs.Entity, optionID string, cons components.DialogConsequences) {
	s.PendingConsequences = append(s.PendingConsequences, PendingConsequence{
		PlayerEntity: player,
		NPCEntity:    npc,
		OptionID:     optionID,
		Consequences: cons,
	})
}

// SelectDialogOption handles a player selecting a dialog option.
func SelectDialogOption(w *ecs.World, s *DialogConsequenceSystem, player ecs.Entity, optionID string) bool {
	// Get player's dialog state
	stateComp, ok := w.GetComponent(player, "DialogState")
	if !ok {
		return false
	}
	state := stateComp.(*components.DialogState)
	if !state.IsInDialog {
		return false
	}

	// Find the selected option
	var selectedOption *components.DialogOption
	for i := range state.AvailableResponses {
		if state.AvailableResponses[i].ID == optionID {
			selectedOption = &state.AvailableResponses[i]
			break
		}
	}
	if selectedOption == nil {
		return false
	}

	// Check if option is enabled
	if !selectedOption.IsEnabled || !selectedOption.IsVisible {
		return false
	}

	// Queue the consequences
	s.QueueConsequence(player, ecs.Entity(state.ConversationPartner), optionID, selectedOption.Consequences)

	// Record in dialog history
	state.DialogHistory = append(state.DialogHistory, components.DialogExchange{
		Speaker:  uint64(player),
		Text:     selectedOption.Text,
		OptionID: optionID,
	})

	// Update current topic
	state.CurrentTopicID = selectedOption.NextTopicID

	return true
}

// StartConversation begins a dialog between two entities.
func StartConversation(w *ecs.World, player, npc ecs.Entity) bool {
	// Ensure both entities can have dialog
	playerState, ok := w.GetComponent(player, "DialogState")
	if !ok {
		w.AddComponent(player, &components.DialogState{})
		playerState, _ = w.GetComponent(player, "DialogState")
	}
	pState := playerState.(*components.DialogState)

	// Check if either is already in dialog
	if pState.IsInDialog {
		return false
	}

	npcState, ok := w.GetComponent(npc, "DialogState")
	if !ok {
		w.AddComponent(npc, &components.DialogState{})
		npcState, _ = w.GetComponent(npc, "DialogState")
	}
	nState := npcState.(*components.DialogState)
	if nState.IsInDialog {
		return false
	}

	// Start the conversation
	pState.IsInDialog = true
	pState.ConversationPartner = uint64(npc)
	pState.CurrentTopicID = "greeting"
	pState.DialogHistory = make([]components.DialogExchange, 0)

	nState.IsInDialog = true
	nState.ConversationPartner = uint64(player)
	nState.CurrentTopicID = "greeting"
	nState.DialogHistory = make([]components.DialogExchange, 0)

	// Update conversation count in NPC's memory
	memComp, ok := w.GetComponent(npc, "DialogMemory")
	if ok {
		mem := memComp.(*components.DialogMemory)
		mem.ConversationCount++
	}

	return true
}

// EndConversation terminates a dialog between two entities.
func EndConversation(w *ecs.World, player, npc ecs.Entity) {
	playerState, ok := w.GetComponent(player, "DialogState")
	if ok {
		pState := playerState.(*components.DialogState)
		pState.IsInDialog = false
		pState.ConversationPartner = 0
		pState.AvailableResponses = nil
	}

	npcState, ok := w.GetComponent(npc, "DialogState")
	if ok {
		nState := npcState.(*components.DialogState)
		nState.IsInDialog = false
		nState.ConversationPartner = 0
		nState.AvailableResponses = nil
	}
}

// CheckDialogRequirements verifies if a player meets option requirements.
func CheckDialogRequirements(w *ecs.World, player ecs.Entity, reqs components.DialogRequirements) bool {
	// Check reputation
	if reqs.MinReputation != 0 {
		facComp, ok := w.GetComponent(player, "Faction")
		if !ok || facComp.(*components.Faction).Reputation < reqs.MinReputation {
			return false
		}
	}

	// Check required items
	if len(reqs.RequiredItems) > 0 {
		invComp, ok := w.GetComponent(player, "Inventory")
		if !ok {
			return false
		}
		inv := invComp.(*components.Inventory)
		for _, reqItem := range reqs.RequiredItems {
			found := false
			for _, item := range inv.Items {
				if item == reqItem {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Additional requirements (skill, quest, flag, gold) would need
	// their respective systems to be implemented

	return true
}

// UpdateDialogOptions refreshes available options based on requirements.
func UpdateDialogOptions(w *ecs.World, player ecs.Entity, options []components.DialogOption) []components.DialogOption {
	result := make([]components.DialogOption, len(options))
	for i, opt := range options {
		result[i] = opt
		result[i].IsVisible = true
		result[i].IsEnabled = CheckDialogRequirements(w, player, opt.Requirements)
	}
	return result
}
