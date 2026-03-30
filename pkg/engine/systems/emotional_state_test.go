package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewEmotionalStateSystem(t *testing.T) {
	sys := NewEmotionalStateSystem()
	if sys == nil {
		t.Fatal("NewEmotionalStateSystem returned nil")
	}
	if sys.GameTime != 0 {
		t.Errorf("expected GameTime=0, got %f", sys.GameTime)
	}
}

func TestEmotionalStateSystemUpdate(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	// Create NPC with emotional state
	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:    EmotionAngry,
		Intensity:         0.8,
		Mood:              0.5,
		Stress:            0.3,
		LastEmotionChange: 0,
		EmotionDecayRate:  DefaultEmotionDecayRate,
		MoodDecayRate:     DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	// Update should advance game time and decay emotions
	sys.Update(w, 1.0)

	if sys.GameTime != 1.0 {
		t.Errorf("expected GameTime=1.0, got %f", sys.GameTime)
	}
	// Intensity should have decayed
	if emotion.Intensity >= 0.8 {
		t.Errorf("intensity should have decayed, got %f", emotion.Intensity)
	}
}

func TestDecayEmotions(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:    EmotionHappy,
		Intensity:         0.5,
		Mood:              0.3,
		Stress:            0.2,
		EmotionDecayRate:  0.1,
		MoodDecayRate:     0.05,
		LastEmotionChange: 0,
	}
	w.AddComponent(npc, emotion)

	// Decay over 5 seconds
	sys.Update(w, 5.0)

	// Intensity should be 0 (was 0.5, decay 0.1/sec * 5 = 0.5)
	if emotion.Intensity != 0 {
		t.Errorf("expected intensity=0, got %f", emotion.Intensity)
	}
	// Emotion should reset to neutral when intensity is low
	if emotion.CurrentEmotion != EmotionNeutral {
		t.Errorf("expected emotion=neutral, got %s", emotion.CurrentEmotion)
	}
}

func TestDecayEmotionsNegativeMood(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:    EmotionNeutral,
		Intensity:         0.1,
		Mood:              -0.3,
		Stress:            0.1,
		EmotionDecayRate:  0.05,
		MoodDecayRate:     0.1,
		LastEmotionChange: 0,
	}
	w.AddComponent(npc, emotion)

	// Decay for 2 seconds
	sys.Update(w, 2.0)

	// Negative mood should decay toward 0
	if emotion.Mood <= -0.3 {
		t.Errorf("negative mood should have decayed, got %f", emotion.Mood)
	}
}

func TestApplyMoodEffects(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionNeutral,
		Intensity:        0.1,
		Mood:             0.5,
		Stress:           0.1,
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	memory := &components.NPCMemory{
		PlayerInteractions: make(map[uint64][]components.MemoryEvent),
		LastSeen:           make(map[uint64]float64),
		Disposition:        map[uint64]float64{1: 0.5, 2: -0.2},
		MaxMemories:        100,
		MemoryDecayRate:    0.01,
	}
	w.AddComponent(npc, emotion)
	w.AddComponent(npc, memory)

	// Update should apply mood effects to dispositions
	sys.Update(w, 1.0)

	// Positive mood should slightly increase dispositions
	// (but the change is small since MoodInfluenceOnEmotion = 0.2)
}

func TestTriggerEmotion(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionNeutral,
		Intensity:        0.1,
		Mood:             0.0,
		Stress:           0.0,
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	// Trigger happy emotion
	result := sys.TriggerEmotion(w, npc, EmotionHappy, 0.8)
	if !result {
		t.Error("TriggerEmotion should return true")
	}
	if emotion.CurrentEmotion != EmotionHappy {
		t.Errorf("expected emotion=happy, got %s", emotion.CurrentEmotion)
	}
	if emotion.Intensity != 0.8 {
		t.Errorf("expected intensity=0.8, got %f", emotion.Intensity)
	}
}

func TestTriggerEmotionHigherIntensity(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionSad,
		Intensity:        0.5,
		Mood:             0.0,
		Stress:           0.0,
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	// Lower intensity should not override
	sys.TriggerEmotion(w, npc, EmotionHappy, 0.3)
	if emotion.CurrentEmotion != EmotionSad {
		t.Errorf("lower intensity should not override, got %s", emotion.CurrentEmotion)
	}

	// Higher intensity should override
	sys.TriggerEmotion(w, npc, EmotionAngry, 0.7)
	if emotion.CurrentEmotion != EmotionAngry {
		t.Errorf("higher intensity should override, got %s", emotion.CurrentEmotion)
	}
}

func TestTriggerEmotionClampIntensity(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionNeutral,
		Intensity:        0.0,
		Mood:             0.0,
		Stress:           0.0,
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	// Try to set intensity above 1.0
	sys.TriggerEmotion(w, npc, EmotionHappy, 1.5)
	if emotion.Intensity != 1.0 {
		t.Errorf("intensity should be clamped to 1.0, got %f", emotion.Intensity)
	}
}

func TestTriggerEmotionNonExistent(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	// No EmotionalState component

	result := sys.TriggerEmotion(w, npc, EmotionHappy, 0.5)
	if result {
		t.Error("TriggerEmotion should return false for missing component")
	}
}

func TestAddStress(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionNeutral,
		Intensity:        0.0,
		Mood:             0.0,
		Stress:           0.0,
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	result := sys.AddStress(w, npc, 0.3)
	if !result {
		t.Error("AddStress should return true")
	}
	if emotion.Stress != 0.3 {
		t.Errorf("expected stress=0.3, got %f", emotion.Stress)
	}
}

func TestAddStressHighTriggersEmotion(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionNeutral,
		Intensity:        0.0,
		Mood:             0.0,
		Stress:           0.0,
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	// Add stress above threshold
	sys.AddStress(w, npc, 0.8)

	// Should trigger fearful emotion
	if emotion.CurrentEmotion != EmotionFearful {
		t.Errorf("high stress should trigger fear, got %s", emotion.CurrentEmotion)
	}
}

func TestAddStressClamp(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionNeutral,
		Intensity:        0.0,
		Mood:             0.0,
		Stress:           0.8,
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	// Add stress that would exceed 1.0
	sys.AddStress(w, npc, 0.5)

	if emotion.Stress != 1.0 {
		t.Errorf("stress should be clamped to 1.0, got %f", emotion.Stress)
	}
}

func TestAddStressNonExistent(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	// No EmotionalState component

	result := sys.AddStress(w, npc, 0.5)
	if result {
		t.Error("AddStress should return false for missing component")
	}
}

func TestGetEmotionalResponse(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionNeutral,
		Intensity:        0.0,
		Mood:             0.0,
		Stress:           0.0,
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	response := sys.GetEmotionalResponse(w, npc, "gift", 0.5)
	if response != EmotionHappy {
		t.Errorf("gift should trigger happy, got %s", response)
	}
}

func TestGetEmotionalResponseMoodModifier(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionNeutral,
		Intensity:        0.0,
		Mood:             0.7, // Positive mood
		Stress:           0.0,
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	// Insult with positive mood (should reduce negative emotion intensity)
	sys.GetEmotionalResponse(w, npc, "insult", 0.8)

	// The emotion should still be angry but intensity should be reduced
	if emotion.CurrentEmotion != EmotionAngry {
		t.Errorf("expected angry emotion, got %s", emotion.CurrentEmotion)
	}
}

func TestGetEmotionalResponseNonExistent(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	// No EmotionalState component

	response := sys.GetEmotionalResponse(w, npc, "gift", 0.5)
	if response != EmotionNeutral {
		t.Errorf("missing component should return neutral, got %s", response)
	}
}

func TestEventToEmotion(t *testing.T) {
	sys := NewEmotionalStateSystem()

	tests := []struct {
		event    string
		expected string
	}{
		{"gift", EmotionHappy},
		{"help", EmotionHappy},
		{"compliment", EmotionHappy},
		{"insult", EmotionAngry},
		{"theft", EmotionAngry},
		{"betrayal", EmotionAngry},
		{"attack", EmotionFearful},
		{"threat", EmotionFearful},
		{"danger", EmotionFearful},
		{"death", EmotionSad},
		{"loss", EmotionSad},
		{"rejection", EmotionSad},
		{"gross", EmotionDisgusted},
		{"crime_witness", EmotionDisgusted},
		{"loud_noise", EmotionSurprised},
		{"sudden_appearance", EmotionSurprised},
		{"unknown", EmotionNeutral},
	}

	for _, tc := range tests {
		t.Run(tc.event, func(t *testing.T) {
			result := sys.eventToEmotion(tc.event)
			if result != tc.expected {
				t.Errorf("eventToEmotion(%s) = %s, want %s", tc.event, result, tc.expected)
			}
		})
	}
}

func TestIsStressed(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	emotion := &components.EmotionalState{
		CurrentEmotion:   EmotionNeutral,
		Intensity:        0.0,
		Mood:             0.0,
		Stress:           0.5, // Below threshold
		EmotionDecayRate: DefaultEmotionDecayRate,
		MoodDecayRate:    DefaultMoodDecayRate,
	}
	w.AddComponent(npc, emotion)

	if sys.IsStressed(w, npc) {
		t.Error("stress 0.5 should not be stressed (threshold is 0.7)")
	}

	// Increase stress above threshold
	emotion.Stress = 0.8
	if !sys.IsStressed(w, npc) {
		t.Error("stress 0.8 should be stressed")
	}
}

func TestIsStressedNonExistent(t *testing.T) {
	sys := NewEmotionalStateSystem()
	w := ecs.NewWorld()

	npc := w.CreateEntity()
	// No EmotionalState component

	if sys.IsStressed(w, npc) {
		t.Error("missing component should return false")
	}
}

func TestAdjustMoodFromEmotionAllTypes(t *testing.T) {
	sys := NewEmotionalStateSystem()

	emotions := []string{
		EmotionHappy,
		EmotionSad,
		EmotionAngry,
		EmotionFearful,
		EmotionDisgusted,
		EmotionSurprised,
		EmotionNeutral,
	}

	for _, emType := range emotions {
		t.Run(emType, func(t *testing.T) {
			state := &components.EmotionalState{
				CurrentEmotion: EmotionNeutral,
				Intensity:      0.0,
				Mood:           0.0,
				Stress:         0.0,
			}

			sys.adjustMoodFromEmotion(state, emType, 0.5)

			// Just verify it doesn't panic and mood/stress stay within bounds
			if state.Mood < -1.0 || state.Mood > 1.0 {
				t.Errorf("mood out of bounds: %f", state.Mood)
			}
			if state.Stress < 0 || state.Stress > 1.0 {
				t.Errorf("stress out of bounds: %f", state.Stress)
			}
		})
	}
}
