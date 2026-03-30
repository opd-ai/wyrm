// Package systems implements ECS system behaviors.
package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// MultiNPCConversationSystem manages conversations involving multiple NPCs.
type MultiNPCConversationSystem struct {
	// ActiveConversations tracks ongoing multi-NPC conversations.
	ActiveConversations map[string]*MultiNPCConversation
}

// MultiNPCConversation represents a conversation with multiple participants.
type MultiNPCConversation struct {
	// ConversationID uniquely identifies this conversation.
	ConversationID string
	// Participants are the entities in the conversation.
	Participants []ecs.Entity
	// PlayerEntity is the player (if involved).
	PlayerEntity ecs.Entity
	// HasPlayer indicates if a player is part of the conversation.
	HasPlayer bool
	// CurrentSpeaker is the entity currently speaking.
	CurrentSpeaker ecs.Entity
	// SpeakerQueue orders who speaks next.
	SpeakerQueue []ecs.Entity
	// TopicID is the current conversation topic.
	TopicID string
	// TurnCount tracks conversation turns.
	TurnCount int
	// MaxTurns limits conversation length (0 = unlimited).
	MaxTurns int
	// IsActive indicates if conversation is ongoing.
	IsActive bool
	// Exchanges records the conversation history.
	Exchanges []MultiNPCExchange
	// AudienceEntities are NPCs listening but not participating.
	AudienceEntities []ecs.Entity
}

// MultiNPCExchange records a single exchange in a multi-NPC conversation.
type MultiNPCExchange struct {
	// Speaker is who spoke.
	Speaker ecs.Entity
	// Text is what was said.
	Text string
	// ResponseTo is the speaker being responded to (0 if not a response).
	ResponseTo ecs.Entity
	// Sentiment is the emotional tone (-1 to 1).
	Sentiment float64
	// TargetAudience specifies who this was directed at (empty = everyone).
	TargetAudience []ecs.Entity
}

// NewMultiNPCConversationSystem creates a new multi-NPC conversation system.
func NewMultiNPCConversationSystem() *MultiNPCConversationSystem {
	return &MultiNPCConversationSystem{
		ActiveConversations: make(map[string]*MultiNPCConversation),
	}
}

// Update processes all active multi-NPC conversations.
func (s *MultiNPCConversationSystem) Update(w *ecs.World, dt float64) {
	// Process each active conversation
	for id, conv := range s.ActiveConversations {
		if !conv.IsActive {
			delete(s.ActiveConversations, id)
			continue
		}

		// Check if any participants have left the area
		if s.checkParticipantsPresent(w, conv) {
			s.advanceConversation(w, conv, dt)
		} else {
			// End conversation if participants separated
			s.EndConversation(w, id)
		}
	}
}

// checkParticipantsPresent verifies all participants are close enough.
func (s *MultiNPCConversationSystem) checkParticipantsPresent(w *ecs.World, conv *MultiNPCConversation) bool {
	if len(conv.Participants) < 2 {
		return false
	}

	// Get center position of first participant
	firstPos, ok := w.GetComponent(conv.Participants[0], "Position")
	if !ok {
		return false
	}
	pos1 := firstPos.(*components.Position)

	// Check all others are within conversation range
	for i := 1; i < len(conv.Participants); i++ {
		posComp, ok := w.GetComponent(conv.Participants[i], "Position")
		if !ok {
			return false
		}
		pos2 := posComp.(*components.Position)

		dx := pos2.X - pos1.X
		dy := pos2.Y - pos1.Y
		distSq := dx*dx + dy*dy
		if distSq > MultiNPCConversationMaxDistance*MultiNPCConversationMaxDistance {
			return false
		}
	}

	return true
}

// advanceConversation progresses the conversation.
func (s *MultiNPCConversationSystem) advanceConversation(w *ecs.World, conv *MultiNPCConversation, dt float64) {
	// If waiting for player input, don't auto-advance
	if conv.HasPlayer && conv.CurrentSpeaker == conv.PlayerEntity {
		return
	}

	// Check turn limit
	if conv.MaxTurns > 0 && conv.TurnCount >= conv.MaxTurns {
		s.EndConversation(w, conv.ConversationID)
		return
	}

	// NPC turn - generate response based on context
	if conv.CurrentSpeaker != 0 && conv.CurrentSpeaker != conv.PlayerEntity {
		// NPCs respond automatically
		s.npcRespond(w, conv)
	}

	// Advance to next speaker
	if len(conv.SpeakerQueue) > 0 {
		conv.CurrentSpeaker = conv.SpeakerQueue[0]
		conv.SpeakerQueue = conv.SpeakerQueue[1:]
	} else {
		// Rebuild speaker queue - round robin
		conv.SpeakerQueue = make([]ecs.Entity, len(conv.Participants))
		copy(conv.SpeakerQueue, conv.Participants)
		if len(conv.SpeakerQueue) > 0 {
			conv.CurrentSpeaker = conv.SpeakerQueue[0]
			conv.SpeakerQueue = conv.SpeakerQueue[1:]
		}
	}

	conv.TurnCount++
}

// npcRespond generates an NPC's response to the conversation.
func (s *MultiNPCConversationSystem) npcRespond(w *ecs.World, conv *MultiNPCConversation) {
	// Get NPC's personality/occupation for context
	speaker := conv.CurrentSpeaker

	// Find who they're responding to (last speaker that wasn't them)
	var respondingTo ecs.Entity
	for i := len(conv.Exchanges) - 1; i >= 0; i-- {
		if conv.Exchanges[i].Speaker != speaker {
			respondingTo = conv.Exchanges[i].Speaker
			break
		}
	}

	// Generate contextual response based on topic and NPC state
	response := s.generateNPCResponse(w, speaker, conv.TopicID, respondingTo)

	// Add to conversation
	exchange := MultiNPCExchange{
		Speaker:    speaker,
		Text:       response.Text,
		ResponseTo: respondingTo,
		Sentiment:  response.Sentiment,
	}
	conv.Exchanges = append(conv.Exchanges, exchange)

	// Update NPC's dialog memory
	s.updateNPCMemory(w, speaker, conv)
}

// NPCResponse holds a generated NPC response.
type NPCResponse struct {
	Text      string
	Sentiment float64
}

// generateNPCResponse creates an appropriate response for an NPC.
func (s *MultiNPCConversationSystem) generateNPCResponse(w *ecs.World, npc ecs.Entity, topicID string, respondingTo ecs.Entity) NPCResponse {
	// Get NPC's occupation for context
	occupation := "person"
	occComp, ok := w.GetComponent(npc, "NPCOccupation")
	if ok {
		occ := occComp.(*components.NPCOccupation)
		occupation = occ.OccupationType
	}

	// Get NPC's attitude
	sentiment := 0.0
	if respondingTo != 0 {
		memComp, ok := w.GetComponent(npc, "DialogMemory")
		if ok {
			mem := memComp.(*components.DialogMemory)
			sentiment = mem.Attitude
		}
	}

	// Generate response based on context
	// In a full implementation, this would use Venture's dialog generator
	response := NPCResponse{
		Text:      getTopicResponse(topicID, occupation),
		Sentiment: sentiment,
	}

	return response
}

// getTopicResponse returns a template response for a topic.
func getTopicResponse(topicID, occupation string) string {
	// Template responses by topic and occupation
	templates := map[string]map[string]string{
		"greeting": {
			"merchant":   "Welcome! Looking to buy or sell?",
			"guard":      "Stay out of trouble.",
			"innkeeper":  "Come in, come in! What can I get you?",
			"blacksmith": "Need something forged?",
			"default":    "Hello there.",
		},
		"weather": {
			"farmer":   "This weather will be good for the crops.",
			"guard":    "At least it's not raining on patrol.",
			"merchant": "Good weather means good business!",
			"default":  "Yes, the weather is something.",
		},
		"rumors": {
			"innkeeper": "I hear all sorts of things in this place...",
			"guard":     "I shouldn't be spreading rumors, but...",
			"merchant":  "My customers tell me interesting things.",
			"default":   "I may have heard something...",
		},
		"work": {
			"merchant":   "Business has been steady.",
			"blacksmith": "Always more weapons to forge.",
			"farmer":     "The fields keep me busy.",
			"guard":      "The patrols never end.",
			"default":    "Work is work.",
		},
	}

	if topicTemplates, ok := templates[topicID]; ok {
		if response, ok := topicTemplates[occupation]; ok {
			return response
		}
		return topicTemplates["default"]
	}

	return "I don't know much about that."
}

// updateNPCMemory records the conversation in NPC's memory.
func (s *MultiNPCConversationSystem) updateNPCMemory(w *ecs.World, npc ecs.Entity, conv *MultiNPCConversation) {
	memComp, ok := w.GetComponent(npc, "DialogMemory")
	if !ok {
		return
	}
	mem := memComp.(*components.DialogMemory)

	// Record topic discussed
	if mem.TopicsDiscussed == nil {
		mem.TopicsDiscussed = make(map[string]int)
	}
	mem.TopicsDiscussed[conv.TopicID]++
}

// StartMultiNPCConversation begins a conversation with multiple NPCs.
func (s *MultiNPCConversationSystem) StartMultiNPCConversation(w *ecs.World, participants []ecs.Entity, player ecs.Entity, topicID string) string {
	if len(participants) < 2 {
		return ""
	}

	// Generate conversation ID
	convID := generateConversationID(participants)

	// Check if already in a conversation
	for _, p := range participants {
		stateComp, ok := w.GetComponent(p, "DialogState")
		if ok {
			state := stateComp.(*components.DialogState)
			if state.IsInDialog {
				return "" // Can't start - someone is busy
			}
		}
	}

	// Create conversation
	conv := &MultiNPCConversation{
		ConversationID: convID,
		Participants:   participants,
		PlayerEntity:   player,
		HasPlayer:      player != 0,
		TopicID:        topicID,
		IsActive:       true,
		Exchanges:      make([]MultiNPCExchange, 0),
		SpeakerQueue:   make([]ecs.Entity, 0),
	}

	// Build initial speaker queue
	for _, p := range participants {
		if p != player { // NPCs first if player involved
			conv.SpeakerQueue = append(conv.SpeakerQueue, p)
		}
	}
	if player != 0 {
		conv.SpeakerQueue = append(conv.SpeakerQueue, player)
	}

	// Set first speaker
	if len(conv.SpeakerQueue) > 0 {
		conv.CurrentSpeaker = conv.SpeakerQueue[0]
		conv.SpeakerQueue = conv.SpeakerQueue[1:]
	}

	// Update participants' dialog states
	for _, p := range participants {
		stateComp, ok := w.GetComponent(p, "DialogState")
		if !ok {
			w.AddComponent(p, &components.DialogState{})
			stateComp, _ = w.GetComponent(p, "DialogState")
		}
		state := stateComp.(*components.DialogState)
		state.IsInDialog = true
		state.CurrentTopicID = topicID
	}

	s.ActiveConversations[convID] = conv
	return convID
}

// EndConversation terminates a multi-NPC conversation.
func (s *MultiNPCConversationSystem) EndConversation(w *ecs.World, convID string) {
	conv, ok := s.ActiveConversations[convID]
	if !ok {
		return
	}

	conv.IsActive = false

	// Update all participants' dialog states
	for _, p := range conv.Participants {
		stateComp, ok := w.GetComponent(p, "DialogState")
		if ok {
			state := stateComp.(*components.DialogState)
			state.IsInDialog = false
		}
	}

	delete(s.ActiveConversations, convID)
}

// PlayerSpeak handles player input in a multi-NPC conversation.
func (s *MultiNPCConversationSystem) PlayerSpeak(w *ecs.World, convID, text string, targetNPC ecs.Entity) bool {
	conv, ok := s.ActiveConversations[convID]
	if !ok || !conv.HasPlayer {
		return false
	}

	if conv.CurrentSpeaker != conv.PlayerEntity {
		return false // Not player's turn
	}

	// Record player's speech
	exchange := MultiNPCExchange{
		Speaker:    conv.PlayerEntity,
		Text:       text,
		ResponseTo: targetNPC,
		Sentiment:  0, // Could analyze text for sentiment
	}
	if targetNPC != 0 {
		exchange.TargetAudience = []ecs.Entity{targetNPC}
	}
	conv.Exchanges = append(conv.Exchanges, exchange)

	// Move to next speaker
	if len(conv.SpeakerQueue) > 0 {
		conv.CurrentSpeaker = conv.SpeakerQueue[0]
		conv.SpeakerQueue = conv.SpeakerQueue[1:]
	}

	conv.TurnCount++
	return true
}

// AddParticipant adds an entity to an ongoing conversation.
func (s *MultiNPCConversationSystem) AddParticipant(w *ecs.World, convID string, entity ecs.Entity) bool {
	conv, ok := s.ActiveConversations[convID]
	if !ok || !conv.IsActive {
		return false
	}

	// Check entity not already in conversation
	for _, p := range conv.Participants {
		if p == entity {
			return false
		}
	}

	// Check entity is not in another conversation
	stateComp, ok := w.GetComponent(entity, "DialogState")
	if ok {
		state := stateComp.(*components.DialogState)
		if state.IsInDialog {
			return false
		}
	}

	// Add to conversation
	conv.Participants = append(conv.Participants, entity)
	conv.SpeakerQueue = append(conv.SpeakerQueue, entity)

	// Update entity's dialog state
	if !ok {
		w.AddComponent(entity, &components.DialogState{})
		stateComp, _ = w.GetComponent(entity, "DialogState")
	}
	state := stateComp.(*components.DialogState)
	state.IsInDialog = true
	state.CurrentTopicID = conv.TopicID

	return true
}

// RemoveParticipant removes an entity from a conversation.
func (s *MultiNPCConversationSystem) RemoveParticipant(w *ecs.World, convID string, entity ecs.Entity) bool {
	conv, ok := s.ActiveConversations[convID]
	if !ok || !conv.IsActive {
		return false
	}

	// Remove from participants
	found := false
	for i, p := range conv.Participants {
		if p == entity {
			conv.Participants = append(conv.Participants[:i], conv.Participants[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return false
	}

	// Remove from speaker queue
	for i, p := range conv.SpeakerQueue {
		if p == entity {
			conv.SpeakerQueue = append(conv.SpeakerQueue[:i], conv.SpeakerQueue[i+1:]...)
			break
		}
	}

	// Update entity's dialog state
	stateComp, ok := w.GetComponent(entity, "DialogState")
	if ok {
		state := stateComp.(*components.DialogState)
		state.IsInDialog = false
	}

	// End conversation if fewer than 2 participants
	if len(conv.Participants) < 2 {
		s.EndConversation(w, convID)
	}

	return true
}

// GetConversationHistory returns the exchange history for a conversation.
func (s *MultiNPCConversationSystem) GetConversationHistory(convID string) []MultiNPCExchange {
	conv, ok := s.ActiveConversations[convID]
	if !ok {
		return nil
	}
	return conv.Exchanges
}

// IsPlayerTurn checks if it's the player's turn to speak.
func (s *MultiNPCConversationSystem) IsPlayerTurn(convID string) bool {
	conv, ok := s.ActiveConversations[convID]
	if !ok || !conv.HasPlayer {
		return false
	}
	return conv.CurrentSpeaker == conv.PlayerEntity
}

// GetCurrentSpeaker returns the entity whose turn it is.
func (s *MultiNPCConversationSystem) GetCurrentSpeaker(convID string) ecs.Entity {
	conv, ok := s.ActiveConversations[convID]
	if !ok {
		return 0
	}
	return conv.CurrentSpeaker
}

// generateConversationID creates a unique ID for a conversation.
func generateConversationID(participants []ecs.Entity) string {
	// Simple ID generation - in production would use UUID
	id := "conv"
	for _, p := range participants {
		id += "_" + string(rune('0'+p%10))
	}
	return id
}
