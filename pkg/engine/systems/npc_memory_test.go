package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNPCMemorySystem_RecordEvent(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	// Record an event
	memSys.RecordEvent(w, npc, player, "gift", 0.1, "Received apple")

	// Check event was recorded
	memories := memSys.GetMemories(w, npc, player)
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}
	if memories[0].EventType != "gift" {
		t.Errorf("Expected event type 'gift', got %s", memories[0].EventType)
	}
}

func TestNPCMemorySystem_DispositionChange(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	// Initial disposition should be neutral
	initial := memSys.GetDisposition(w, npc, player)
	if initial != NeutralDisposition {
		t.Errorf("Expected neutral disposition, got %f", initial)
	}

	// Record positive event
	memSys.RecordEvent(w, npc, player, "gift", 0.3, "Big gift")

	disposition := memSys.GetDisposition(w, npc, player)
	if disposition != 0.3 {
		t.Errorf("Expected disposition 0.3, got %f", disposition)
	}
}

func TestNPCMemorySystem_DispositionClamp(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	// Record many positive events
	for i := 0; i < 10; i++ {
		memSys.RecordEvent(w, npc, player, "gift", 0.5, "Gift")
	}

	disposition := memSys.GetDisposition(w, npc, player)
	if disposition > DispositionClamp {
		t.Errorf("Disposition should be clamped to %f, got %f", DispositionClamp, disposition)
	}
}

func TestNPCMemorySystem_NegativeDisposition(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	// Record attack
	memSys.RecordAttack(w, npc, player, false)

	disposition := memSys.GetDisposition(w, npc, player)
	if disposition >= 0 {
		t.Errorf("Expected negative disposition after attack, got %f", disposition)
	}

	// Single attack doesn't make hostile (needs disposition < -0.5)
	// Attack twice to cross threshold
	memSys.RecordAttack(w, npc, player, false)

	if !memSys.IsHostile(w, npc, player) {
		t.Error("NPC should be hostile after multiple attacks")
	}
}

func TestNPCMemorySystem_HasMemoryOf(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()
	stranger := w.CreateEntity()

	// Initially no memories
	if memSys.HasMemoryOf(w, npc, player) {
		t.Error("NPC should not have memories of player initially")
	}

	// Record event
	memSys.RecordEvent(w, npc, player, "dialog", 0.05, "Small talk")

	if !memSys.HasMemoryOf(w, npc, player) {
		t.Error("NPC should have memories of player after interaction")
	}

	if memSys.HasMemoryOf(w, npc, stranger) {
		t.Error("NPC should not have memories of stranger")
	}
}

func TestNPCMemorySystem_ForgetPlayer(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	memSys.RecordEvent(w, npc, player, "gift", 0.2, "Gift")

	if !memSys.HasMemoryOf(w, npc, player) {
		t.Error("Should have memory before forget")
	}

	memSys.ForgetPlayer(w, npc, player)

	if memSys.HasMemoryOf(w, npc, player) {
		t.Error("Should have no memory after forget")
	}

	if memSys.GetDisposition(w, npc, player) != NeutralDisposition {
		t.Error("Disposition should be neutral after forget")
	}
}

func TestNPCMemorySystem_MemoryLimit(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	// Add memory component with low limit
	memory := &components.NPCMemory{
		PlayerInteractions: make(map[uint64][]MemoryEvent),
		LastSeen:           make(map[uint64]float64),
		Disposition:        make(map[uint64]float64),
		MaxMemories:        5,
	}
	w.AddComponent(npc, memory)

	// Record more events than the limit
	for i := 0; i < 10; i++ {
		memSys.RecordEvent(w, npc, player, "dialog", 0.01, "Chat")
	}

	memories := memSys.GetMemories(w, npc, player)
	if len(memories) > 5 {
		t.Errorf("Expected max 5 memories, got %d", len(memories))
	}
}

func TestNPCMemorySystem_DispositionDecay(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	// Set initial disposition
	memSys.SetDisposition(w, npc, player, 0.5)

	// Simulate time passing with decay
	for i := 0; i < 100; i++ {
		memSys.Update(w, 10.0) // 10 seconds per update
	}

	disposition := memSys.GetDisposition(w, npc, player)
	if disposition >= 0.5 {
		t.Errorf("Disposition should have decayed from 0.5, got %f", disposition)
	}
}

func TestNPCMemorySystem_IsFriendly(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	// Not friendly initially
	if memSys.IsFriendly(w, npc, player) {
		t.Error("Should not be friendly at neutral")
	}

	// Set high disposition
	memSys.SetDisposition(w, npc, player, 0.7)

	if !memSys.IsFriendly(w, npc, player) {
		t.Error("Should be friendly at 0.7 disposition")
	}
}

func TestNPCMemorySystem_RecordGift(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	memSys.RecordGift(w, npc, player, "Diamond Ring", true)

	memories := memSys.GetMemories(w, npc, player)
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}
	if memories[0].EventType != "gift" {
		t.Errorf("Expected gift event type, got %s", memories[0].EventType)
	}

	disposition := memSys.GetDisposition(w, npc, player)
	if disposition != ImpactGift {
		t.Errorf("Expected disposition %f, got %f", ImpactGift, disposition)
	}
}

func TestNPCMemorySystem_RecordTheft(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	memSys.RecordTheft(w, npc, player, "Gold Coin")

	disposition := memSys.GetDisposition(w, npc, player)
	if disposition >= 0 {
		t.Errorf("Expected negative disposition after theft, got %f", disposition)
	}
}

func TestNPCMemorySystem_RecordQuestComplete(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	memSys.RecordQuestComplete(w, npc, player, "Save the Village")

	disposition := memSys.GetDisposition(w, npc, player)
	if disposition != ImpactQuestComplete {
		t.Errorf("Expected disposition %f after quest, got %f", ImpactQuestComplete, disposition)
	}
}

func TestNPCMemorySystem_LastSeen(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	// Not seen initially
	_, seen := memSys.GetLastSeen(w, npc, player)
	if seen {
		t.Error("Should not have last seen time initially")
	}

	// Simulate time passing
	memSys.Update(w, 100.0)

	// Record event
	memSys.RecordEvent(w, npc, player, "dialog", 0.0, "Hello")

	lastSeen, seen := memSys.GetLastSeen(w, npc, player)
	if !seen {
		t.Error("Should have last seen time after interaction")
	}
	if lastSeen != 100.0 {
		t.Errorf("Expected last seen at 100.0, got %f", lastSeen)
	}
}

func TestGetDispositionCategory(t *testing.T) {
	tests := []struct {
		disposition float64
		expected    string
	}{
		{0.9, "adoring"},
		{0.6, "friendly"},
		{0.3, "warm"},
		{0.0, "neutral"},
		{-0.3, "cold"},
		{-0.6, "hostile"},
		{-0.9, "hated"},
	}

	for _, tt := range tests {
		result := GetDispositionCategory(tt.disposition)
		if result != tt.expected {
			t.Errorf("GetDispositionCategory(%f) = %s, expected %s", tt.disposition, result, tt.expected)
		}
	}
}

func TestNPCMemorySystem_RecordWitnessedCrime(t *testing.T) {
	w := ecs.NewWorld()
	memSys := NewNPCMemorySystem()

	npc := w.CreateEntity()
	player := w.CreateEntity()

	memSys.RecordWitnessedCrime(w, npc, player, "assault")

	memories := memSys.GetMemories(w, npc, player)
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}

	disposition := memSys.GetDisposition(w, npc, player)
	if disposition != ImpactWitness {
		t.Errorf("Expected disposition %f, got %f", ImpactWitness, disposition)
	}
}
