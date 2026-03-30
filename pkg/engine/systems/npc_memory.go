package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// NPC memory constants.
const (
	// DefaultMaxMemories is the default number of events to remember per player.
	DefaultMaxMemories = 20
	// DefaultMemoryDecayRate is the default disposition decay per second.
	DefaultMemoryDecayRate = 0.001
	// DispositionClamp is the maximum absolute disposition value.
	DispositionClamp = 1.0
	// NeutralDisposition is the starting disposition for unknown players.
	NeutralDisposition = 0.0
)

// Event impact presets for common interaction types.
const (
	ImpactGift          = 0.15
	ImpactSmallGift     = 0.05
	ImpactAttack        = -0.3
	ImpactKill          = -0.8
	ImpactTheft         = -0.25
	ImpactHelp          = 0.2
	ImpactQuestComplete = 0.3
	ImpactDialogGood    = 0.05
	ImpactDialogBad     = -0.1
	ImpactWitness       = -0.15
)

// NPCMemorySystem manages NPC memories and disposition toward players.
type NPCMemorySystem struct {
	// GameTime tracks elapsed time for memory decay.
	GameTime float64
}

// NewNPCMemorySystem creates a new NPC memory system.
func NewNPCMemorySystem() *NPCMemorySystem {
	return &NPCMemorySystem{GameTime: 0}
}

// Update processes memory decay for all NPCs.
func (s *NPCMemorySystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	s.decayDispositions(w, dt)
}

// decayDispositions gradually moves dispositions toward neutral over time.
func (s *NPCMemorySystem) decayDispositions(w *ecs.World, dt float64) {
	for _, e := range w.Entities("NPCMemory") {
		memComp, ok := w.GetComponent(e, "NPCMemory")
		if !ok {
			continue
		}
		memory := memComp.(*components.NPCMemory)
		s.decayMemoryDispositions(memory, dt)
	}
}

// decayMemoryDispositions decays all dispositions in an NPC's memory.
func (s *NPCMemorySystem) decayMemoryDispositions(memory *components.NPCMemory, dt float64) {
	decayRate := memory.MemoryDecayRate
	if decayRate <= 0 {
		decayRate = DefaultMemoryDecayRate
	}

	for playerID, disposition := range memory.Disposition {
		memory.Disposition[playerID] = decayTowardNeutral(disposition, decayRate, dt)
	}
}

// decayTowardNeutral moves a value toward NeutralDisposition.
func decayTowardNeutral(value, decayRate, dt float64) float64 {
	if value > NeutralDisposition {
		value -= decayRate * dt
		if value < NeutralDisposition {
			return NeutralDisposition
		}
	} else if value < NeutralDisposition {
		value += decayRate * dt
		if value > NeutralDisposition {
			return NeutralDisposition
		}
	}
	return value
}

// RecordEvent adds a memory event to an NPC's memory of a player.
func (s *NPCMemorySystem) RecordEvent(w *ecs.World, npc, player ecs.Entity, eventType string, impact float64, details string) {
	memComp, ok := w.GetComponent(npc, "NPCMemory")
	if !ok {
		// Create memory component if it doesn't exist
		memory := &components.NPCMemory{
			PlayerInteractions: make(map[uint64][]MemoryEvent),
			LastSeen:           make(map[uint64]float64),
			Disposition:        make(map[uint64]float64),
			MaxMemories:        DefaultMaxMemories,
			MemoryDecayRate:    DefaultMemoryDecayRate,
		}
		w.AddComponent(npc, memory)
		memComp, _ = w.GetComponent(npc, "NPCMemory")
	}
	memory := memComp.(*components.NPCMemory)

	playerID := uint64(player)

	// Initialize maps if needed
	if memory.PlayerInteractions == nil {
		memory.PlayerInteractions = make(map[uint64][]MemoryEvent)
	}
	if memory.LastSeen == nil {
		memory.LastSeen = make(map[uint64]float64)
	}
	if memory.Disposition == nil {
		memory.Disposition = make(map[uint64]float64)
	}

	// Record the event
	event := MemoryEvent{
		EventType: eventType,
		Timestamp: s.GameTime,
		Impact:    impact,
		Details:   details,
	}
	memory.PlayerInteractions[playerID] = append(memory.PlayerInteractions[playerID], event)

	// Trim old memories if over limit
	maxMem := memory.MaxMemories
	if maxMem <= 0 {
		maxMem = DefaultMaxMemories
	}
	if len(memory.PlayerInteractions[playerID]) > maxMem {
		memory.PlayerInteractions[playerID] = memory.PlayerInteractions[playerID][1:]
	}

	// Update last seen
	memory.LastSeen[playerID] = s.GameTime

	// Update disposition
	currentDisposition := memory.Disposition[playerID]
	newDisposition := currentDisposition + impact
	memory.Disposition[playerID] = clampDisposition(newDisposition)
}

// MemoryEvent is a local type alias for the component type.
type MemoryEvent = components.MemoryEvent

// GetDisposition returns the NPC's disposition toward a player.
func (s *NPCMemorySystem) GetDisposition(w *ecs.World, npc, player ecs.Entity) float64 {
	memComp, ok := w.GetComponent(npc, "NPCMemory")
	if !ok {
		return NeutralDisposition
	}
	memory := memComp.(*components.NPCMemory)

	if memory.Disposition == nil {
		return NeutralDisposition
	}

	disposition, exists := memory.Disposition[uint64(player)]
	if !exists {
		return NeutralDisposition
	}
	return disposition
}

// SetDisposition directly sets the NPC's disposition toward a player.
func (s *NPCMemorySystem) SetDisposition(w *ecs.World, npc, player ecs.Entity, disposition float64) {
	memComp, ok := w.GetComponent(npc, "NPCMemory")
	if !ok {
		memory := &components.NPCMemory{
			PlayerInteractions: make(map[uint64][]MemoryEvent),
			LastSeen:           make(map[uint64]float64),
			Disposition:        make(map[uint64]float64),
			MaxMemories:        DefaultMaxMemories,
			MemoryDecayRate:    DefaultMemoryDecayRate,
		}
		w.AddComponent(npc, memory)
		memComp, _ = w.GetComponent(npc, "NPCMemory")
	}
	memory := memComp.(*components.NPCMemory)

	if memory.Disposition == nil {
		memory.Disposition = make(map[uint64]float64)
	}

	memory.Disposition[uint64(player)] = clampDisposition(disposition)
}

// GetMemories returns all memory events for a player.
func (s *NPCMemorySystem) GetMemories(w *ecs.World, npc, player ecs.Entity) []MemoryEvent {
	memComp, ok := w.GetComponent(npc, "NPCMemory")
	if !ok {
		return nil
	}
	memory := memComp.(*components.NPCMemory)

	if memory.PlayerInteractions == nil {
		return nil
	}

	return memory.PlayerInteractions[uint64(player)]
}

// HasMemoryOf checks if an NPC has any memories of a player.
func (s *NPCMemorySystem) HasMemoryOf(w *ecs.World, npc, player ecs.Entity) bool {
	memComp, ok := w.GetComponent(npc, "NPCMemory")
	if !ok {
		return false
	}
	memory := memComp.(*components.NPCMemory)

	if memory.PlayerInteractions == nil {
		return false
	}

	events, exists := memory.PlayerInteractions[uint64(player)]
	return exists && len(events) > 0
}

// GetLastSeen returns when the NPC last saw the player.
func (s *NPCMemorySystem) GetLastSeen(w *ecs.World, npc, player ecs.Entity) (float64, bool) {
	memComp, ok := w.GetComponent(npc, "NPCMemory")
	if !ok {
		return 0, false
	}
	memory := memComp.(*components.NPCMemory)

	if memory.LastSeen == nil {
		return 0, false
	}

	lastSeen, exists := memory.LastSeen[uint64(player)]
	return lastSeen, exists
}

// ForgetPlayer removes all memories of a player from an NPC.
func (s *NPCMemorySystem) ForgetPlayer(w *ecs.World, npc, player ecs.Entity) {
	memComp, ok := w.GetComponent(npc, "NPCMemory")
	if !ok {
		return
	}
	memory := memComp.(*components.NPCMemory)
	playerID := uint64(player)

	delete(memory.PlayerInteractions, playerID)
	delete(memory.LastSeen, playerID)
	delete(memory.Disposition, playerID)
}

// IsHostile returns true if the NPC's disposition is below -0.5.
func (s *NPCMemorySystem) IsHostile(w *ecs.World, npc, player ecs.Entity) bool {
	return s.GetDisposition(w, npc, player) < -0.5
}

// IsFriendly returns true if the NPC's disposition is above 0.5.
func (s *NPCMemorySystem) IsFriendly(w *ecs.World, npc, player ecs.Entity) bool {
	return s.GetDisposition(w, npc, player) > 0.5
}

// RecordGift records a gift interaction.
func (s *NPCMemorySystem) RecordGift(w *ecs.World, npc, player ecs.Entity, itemName string, isValuable bool) {
	impact := ImpactSmallGift
	if isValuable {
		impact = ImpactGift
	}
	s.RecordEvent(w, npc, player, "gift", impact, "Received "+itemName)
}

// RecordAttack records an attack interaction.
func (s *NPCMemorySystem) RecordAttack(w *ecs.World, npc, player ecs.Entity, wasKilled bool) {
	impact := ImpactAttack
	eventType := "attack"
	details := "Was attacked"
	if wasKilled {
		impact = ImpactKill
		eventType = "kill"
		details = "Was killed"
	}
	s.RecordEvent(w, npc, player, eventType, impact, details)
}

// RecordTheft records a theft interaction.
func (s *NPCMemorySystem) RecordTheft(w *ecs.World, npc, player ecs.Entity, itemName string) {
	s.RecordEvent(w, npc, player, "theft", ImpactTheft, "Stolen: "+itemName)
}

// RecordHelp records the player helping the NPC.
func (s *NPCMemorySystem) RecordHelp(w *ecs.World, npc, player ecs.Entity, details string) {
	s.RecordEvent(w, npc, player, "help", ImpactHelp, details)
}

// RecordQuestComplete records quest completion for the NPC.
func (s *NPCMemorySystem) RecordQuestComplete(w *ecs.World, npc, player ecs.Entity, questName string) {
	s.RecordEvent(w, npc, player, "quest_complete", ImpactQuestComplete, "Completed: "+questName)
}

// RecordWitnessedCrime records that the NPC witnessed the player commit a crime.
func (s *NPCMemorySystem) RecordWitnessedCrime(w *ecs.World, npc, player ecs.Entity, crimeType string) {
	s.RecordEvent(w, npc, player, "witness", ImpactWitness, "Witnessed: "+crimeType)
}

// clampDisposition ensures disposition stays within valid range.
func clampDisposition(d float64) float64 {
	if d > DispositionClamp {
		return DispositionClamp
	}
	if d < -DispositionClamp {
		return -DispositionClamp
	}
	return d
}

// GetDispositionCategory returns a human-readable category for the disposition.
func GetDispositionCategory(disposition float64) string {
	switch {
	case disposition >= 0.8:
		return "adoring"
	case disposition >= 0.5:
		return "friendly"
	case disposition >= 0.2:
		return "warm"
	case disposition >= -0.2:
		return "neutral"
	case disposition >= -0.5:
		return "cold"
	case disposition >= -0.8:
		return "hostile"
	default:
		return "hated"
	}
}
