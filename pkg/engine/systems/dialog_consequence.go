// Package systems implements ECS system behaviors.
package systems

import (
	"sync"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// maxDialogHistoryPerEntity limits stored dialog history to prevent unbounded growth.
const maxDialogHistoryPerEntity = 100

// DialogConsequenceSystem processes dialog choices and applies their effects.
type DialogConsequenceSystem struct {
	// PendingConsequences queues consequences to process.
	PendingConsequences []PendingConsequence
	// mu protects PendingConsequences from concurrent access.
	mu sync.Mutex
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
	s.mu.Lock()
	for len(s.PendingConsequences) > 0 {
		pending := s.PendingConsequences[0]
		s.PendingConsequences = s.PendingConsequences[1:]
		s.mu.Unlock()
		s.applyConsequences(w, pending)
		s.mu.Lock()
	}
	s.mu.Unlock()

	// Check for broken promises and update relationships
	s.checkPromises(w, dt)
}

// applyConsequences applies a single set of dialog consequences.
func (s *DialogConsequenceSystem) applyConsequences(w *ecs.World, pending PendingConsequence) {
	cons := pending.Consequences

	s.applyResourceChanges(w, pending.PlayerEntity, cons)
	s.applyInventoryChanges(w, pending.PlayerEntity, cons)
	s.applyQuestChanges(w, pending.PlayerEntity, cons)
	s.applyWorldFlagChanges(w, cons)
	s.applyRelationshipAndMoodChanges(w, pending, cons)

	if cons.TriggerCombat {
		s.triggerCombat(w, pending.PlayerEntity, pending.NPCEntity)
	}

	s.recordDialogEvent(w, pending.NPCEntity, pending.PlayerEntity, pending.OptionID, cons)
}

// applyResourceChanges handles reputation and gold changes.
func (s *DialogConsequenceSystem) applyResourceChanges(w *ecs.World, player ecs.Entity, cons components.DialogConsequences) {
	if cons.ReputationChange != 0 && cons.FactionID != "" {
		s.applyReputationChange(w, player, cons.FactionID, cons.ReputationChange)
	}
	if cons.GoldChange != 0 {
		s.applyGoldChange(w, player, cons.GoldChange)
	}
}

// applyInventoryChanges handles item giving and taking.
func (s *DialogConsequenceSystem) applyInventoryChanges(w *ecs.World, player ecs.Entity, cons components.DialogConsequences) {
	for _, itemID := range cons.ItemsGiven {
		s.giveItem(w, player, itemID)
	}
	for _, itemID := range cons.ItemsTaken {
		s.takeItem(w, player, itemID)
	}
}

// applyQuestChanges handles quest start, progress, and completion.
func (s *DialogConsequenceSystem) applyQuestChanges(w *ecs.World, player ecs.Entity, cons components.DialogConsequences) {
	if cons.QuestStart != "" {
		s.startQuest(w, player, cons.QuestStart)
	}
	if cons.QuestProgress != "" {
		s.progressQuest(w, player, cons.QuestProgress)
	}
	if cons.QuestComplete != "" {
		s.completeQuest(w, player, cons.QuestComplete)
	}
}

// applyWorldFlagChanges handles world flag setting and clearing.
func (s *DialogConsequenceSystem) applyWorldFlagChanges(w *ecs.World, cons components.DialogConsequences) {
	if cons.FlagSet != "" {
		s.setWorldFlag(w, cons.FlagSet, true)
	}
	if cons.FlagClear != "" {
		s.setWorldFlag(w, cons.FlagClear, false)
	}
}

// applyRelationshipAndMoodChanges handles NPC relationship and mood updates.
func (s *DialogConsequenceSystem) applyRelationshipAndMoodChanges(w *ecs.World, pending PendingConsequence, cons components.DialogConsequences) {
	if cons.RelationshipChange != 0 {
		s.applyRelationshipChange(w, pending.NPCEntity, pending.PlayerEntity, cons.RelationshipChange)
	}
	if cons.NPCMood != "" {
		s.setNPCMood(w, pending.NPCEntity, cons.NPCMood)
	}
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

	eventType, sentiment := determineEventTypeAndSentiment(cons)

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

// determineEventTypeAndSentiment categorizes a dialog consequence.
func determineEventTypeAndSentiment(cons components.DialogConsequences) (eventType string, sentiment float64) {
	eventType = "conversation"
	sentiment = 0.0

	switch {
	case cons.TriggerCombat:
		return "combat_started", -1.0
	case cons.RelationshipChange < 0:
		return "negative_interaction", cons.RelationshipChange
	case cons.RelationshipChange > 0:
		return "positive_interaction", cons.RelationshipChange
	case cons.QuestComplete != "":
		return "quest_completed", 0.5
	case cons.QuestStart != "":
		return "quest_given", 0.3
	}
	return eventType, sentiment
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PendingConsequences = append(s.PendingConsequences, PendingConsequence{
		PlayerEntity: player,
		NPCEntity:    npc,
		OptionID:     optionID,
		Consequences: cons,
	})
}

// SelectDialogOption handles a player selecting a dialog option.
func SelectDialogOption(w *ecs.World, s *DialogConsequenceSystem, player ecs.Entity, optionID string) bool {
	state, ok := getPlayerDialogState(w, player)
	if !ok || !state.IsInDialog {
		return false
	}

	selectedOption := findDialogOption(state, optionID)
	if selectedOption == nil || !selectedOption.IsEnabled || !selectedOption.IsVisible {
		return false
	}

	s.QueueConsequence(player, ecs.Entity(state.ConversationPartner), optionID, selectedOption.Consequences)
	recordDialogSelection(state, player, selectedOption, optionID)
	return true
}

// getPlayerDialogState retrieves the dialog state component for a player.
func getPlayerDialogState(w *ecs.World, player ecs.Entity) (*components.DialogState, bool) {
	stateComp, ok := w.GetComponent(player, "DialogState")
	if !ok {
		return nil, false
	}
	return stateComp.(*components.DialogState), true
}

// findDialogOption locates a dialog option by ID in the available responses.
func findDialogOption(state *components.DialogState, optionID string) *components.DialogOption {
	for i := range state.AvailableResponses {
		if state.AvailableResponses[i].ID == optionID {
			return &state.AvailableResponses[i]
		}
	}
	return nil
}

// recordDialogSelection updates dialog state with the selected option.
func recordDialogSelection(state *components.DialogState, player ecs.Entity, option *components.DialogOption, optionID string) {
	state.DialogHistory = append(state.DialogHistory, components.DialogExchange{
		Speaker:  uint64(player),
		Text:     option.Text,
		OptionID: optionID,
	})
	// Trim oldest entries if over the limit
	if len(state.DialogHistory) > maxDialogHistoryPerEntity {
		state.DialogHistory = state.DialogHistory[1:]
	}
	state.CurrentTopicID = option.NextTopicID
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
	if !checkReputationRequirement(w, player, reqs.MinReputation) {
		return false
	}
	if !checkInventoryRequirements(w, player, reqs.RequiredItems) {
		return false
	}
	// Additional requirements (skill, quest, flag, gold) would need
	// their respective systems to be implemented
	return true
}

// checkReputationRequirement verifies minimum reputation is met.
func checkReputationRequirement(w *ecs.World, player ecs.Entity, minReputation float64) bool {
	if minReputation == 0 {
		return true
	}
	facComp, ok := w.GetComponent(player, "Faction")
	if !ok {
		return false
	}
	return facComp.(*components.Faction).Reputation >= minReputation
}

// checkInventoryRequirements verifies all required items are present.
func checkInventoryRequirements(w *ecs.World, player ecs.Entity, requiredItems []string) bool {
	if len(requiredItems) == 0 {
		return true
	}
	invComp, ok := w.GetComponent(player, "Inventory")
	if !ok {
		return false
	}
	inv := invComp.(*components.Inventory)
	for _, reqItem := range requiredItems {
		if !hasItem(inv.Items, reqItem) {
			return false
		}
	}
	return true
}

// hasItem checks if an item exists in the inventory.
func hasItem(items []string, itemID string) bool {
	for _, item := range items {
		if item == itemID {
			return true
		}
	}
	return false
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
