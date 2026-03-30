package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// Emotional state constants.
const (
	// DefaultEmotionDecayRate is how fast emotions return to neutral (per second).
	DefaultEmotionDecayRate = 0.05
	// DefaultMoodDecayRate is how fast mood returns to neutral (per second).
	DefaultMoodDecayRate = 0.01
	// StressThreshold is the stress level that triggers behavioral changes.
	StressThreshold = 0.7
	// EmotionIntensityThreshold is the minimum intensity for visible emotion.
	EmotionIntensityThreshold = 0.3
	// MoodInfluenceOnEmotion is how much mood affects emotion intensity.
	MoodInfluenceOnEmotion = 0.2
)

// Emotion types.
const (
	EmotionNeutral   = "neutral"
	EmotionHappy     = "happy"
	EmotionSad       = "sad"
	EmotionAngry     = "angry"
	EmotionFearful   = "fearful"
	EmotionDisgusted = "disgusted"
	EmotionSurprised = "surprised"
)

// EmotionalStateSystem handles NPC emotional states and mood.
type EmotionalStateSystem struct {
	// GameTime tracks elapsed time.
	GameTime float64
}

// NewEmotionalStateSystem creates a new emotional state system.
func NewEmotionalStateSystem() *EmotionalStateSystem {
	return &EmotionalStateSystem{GameTime: 0}
}

// Update processes emotion decay and mood effects.
func (s *EmotionalStateSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	s.decayEmotions(w, dt)
	s.applyMoodEffects(w, dt)
}

// decayEmotions gradually returns emotions to neutral.
func (s *EmotionalStateSystem) decayEmotions(w *ecs.World, dt float64) {
	for _, e := range w.Entities("EmotionalState") {
		emotionComp, ok := w.GetComponent(e, "EmotionalState")
		if !ok {
			continue
		}
		emotion := emotionComp.(*components.EmotionalState)

		// Decay intensity toward zero
		if emotion.Intensity > 0 {
			emotion.Intensity -= emotion.EmotionDecayRate * dt
			if emotion.Intensity < 0 {
				emotion.Intensity = 0
			}
		}

		// Reset emotion to neutral when intensity is low
		if emotion.Intensity < EmotionIntensityThreshold {
			emotion.CurrentEmotion = EmotionNeutral
		}

		// Decay mood toward zero
		if emotion.Mood > 0 {
			emotion.Mood -= emotion.MoodDecayRate * dt
			if emotion.Mood < 0 {
				emotion.Mood = 0
			}
		} else if emotion.Mood < 0 {
			emotion.Mood += emotion.MoodDecayRate * dt
			if emotion.Mood > 0 {
				emotion.Mood = 0
			}
		}

		// Decay stress
		if emotion.Stress > 0 {
			emotion.Stress -= emotion.MoodDecayRate * dt * 0.5
			if emotion.Stress < 0 {
				emotion.Stress = 0
			}
		}
	}
}

// applyMoodEffects influences NPC behavior based on mood.
func (s *EmotionalStateSystem) applyMoodEffects(w *ecs.World, dt float64) {
	for _, e := range w.Entities("EmotionalState", "NPCMemory") {
		emotionComp, _ := w.GetComponent(e, "EmotionalState")
		emotion := emotionComp.(*components.EmotionalState)

		memoryComp, ok := w.GetComponent(e, "NPCMemory")
		if !ok {
			continue
		}
		memory := memoryComp.(*components.NPCMemory)

		// Mood affects disposition toward all known players
		moodModifier := emotion.Mood * MoodInfluenceOnEmotion * dt
		for playerID := range memory.Disposition {
			memory.Disposition[playerID] += moodModifier
			// Clamp disposition
			if memory.Disposition[playerID] > DispositionClamp {
				memory.Disposition[playerID] = DispositionClamp
			}
			if memory.Disposition[playerID] < -DispositionClamp {
				memory.Disposition[playerID] = -DispositionClamp
			}
		}
	}
}

// TriggerEmotion sets an NPC's emotional state.
func (s *EmotionalStateSystem) TriggerEmotion(w *ecs.World, entity ecs.Entity, emotion string, intensity float64) bool {
	emotionComp, ok := w.GetComponent(entity, "EmotionalState")
	if !ok {
		return false
	}
	state := emotionComp.(*components.EmotionalState)

	// Only override if new emotion is more intense
	if intensity > state.Intensity || emotion == state.CurrentEmotion {
		state.CurrentEmotion = emotion
		state.Intensity = intensity
		if state.Intensity > 1.0 {
			state.Intensity = 1.0
		}
		state.LastEmotionChange = s.GameTime

		// Adjust mood based on emotion
		s.adjustMoodFromEmotion(state, emotion, intensity)
	}
	return true
}

// adjustMoodFromEmotion modifies mood based on triggered emotion.
func (s *EmotionalStateSystem) adjustMoodFromEmotion(state *components.EmotionalState, emotion string, intensity float64) {
	moodChange := intensity * 0.1
	switch emotion {
	case EmotionHappy:
		state.Mood += moodChange
	case EmotionSad, EmotionFearful:
		state.Mood -= moodChange
		state.Stress += intensity * 0.05
	case EmotionAngry:
		state.Mood -= moodChange * 0.5
		state.Stress += intensity * 0.1
	case EmotionDisgusted:
		state.Mood -= moodChange * 0.3
	case EmotionSurprised:
		// Surprise is neutral on mood
	}

	// Clamp values
	if state.Mood > 1.0 {
		state.Mood = 1.0
	}
	if state.Mood < -1.0 {
		state.Mood = -1.0
	}
	if state.Stress > 1.0 {
		state.Stress = 1.0
	}
}

// AddStress increases an NPC's stress level.
func (s *EmotionalStateSystem) AddStress(w *ecs.World, entity ecs.Entity, amount float64) bool {
	emotionComp, ok := w.GetComponent(entity, "EmotionalState")
	if !ok {
		return false
	}
	state := emotionComp.(*components.EmotionalState)

	state.Stress += amount
	if state.Stress > 1.0 {
		state.Stress = 1.0
	}

	// High stress can trigger negative emotions
	if state.Stress > StressThreshold && state.Intensity < state.Stress {
		s.TriggerEmotion(w, entity, EmotionFearful, state.Stress*0.5)
	}
	return true
}

// GetEmotionalResponse determines how an NPC should react emotionally.
func (s *EmotionalStateSystem) GetEmotionalResponse(w *ecs.World, entity ecs.Entity, eventType string, intensity float64) string {
	emotionComp, ok := w.GetComponent(entity, "EmotionalState")
	if !ok {
		return EmotionNeutral
	}
	state := emotionComp.(*components.EmotionalState)

	// Base response on event type
	baseEmotion := s.eventToEmotion(eventType)

	// Mood modifies response
	if state.Mood > 0.5 {
		// Positive mood reduces negative emotions
		if baseEmotion == EmotionAngry || baseEmotion == EmotionSad {
			intensity *= 0.5
		}
	} else if state.Mood < -0.5 {
		// Negative mood amplifies negative emotions
		if baseEmotion == EmotionAngry || baseEmotion == EmotionSad || baseEmotion == EmotionFearful {
			intensity *= 1.5
		}
	}

	// Apply the emotion
	s.TriggerEmotion(w, entity, baseEmotion, intensity)
	return baseEmotion
}

// eventToEmotion maps event types to emotions.
func (s *EmotionalStateSystem) eventToEmotion(eventType string) string {
	switch eventType {
	case "gift", "help", "compliment":
		return EmotionHappy
	case "insult", "theft", "betrayal":
		return EmotionAngry
	case "attack", "threat", "danger":
		return EmotionFearful
	case "death", "loss", "rejection":
		return EmotionSad
	case "gross", "crime_witness":
		return EmotionDisgusted
	case "loud_noise", "sudden_appearance":
		return EmotionSurprised
	default:
		return EmotionNeutral
	}
}

// IsStressed checks if an NPC is under high stress.
func (s *EmotionalStateSystem) IsStressed(w *ecs.World, entity ecs.Entity) bool {
	emotionComp, ok := w.GetComponent(entity, "EmotionalState")
	if !ok {
		return false
	}
	state := emotionComp.(*components.EmotionalState)
	return state.Stress > StressThreshold
}
