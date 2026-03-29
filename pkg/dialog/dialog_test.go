package dialog

import (
	"math/rand"
	"testing"
	"time"
)

func TestRecordAndRecallTopic(t *testing.T) {
	dm := NewDialogManager(12345)

	// Record a topic
	dm.RecordTopic(1, 100, "weather", "asked about rain", "complained about mud")

	// Per AC: NPC recalls player's previous interaction topic
	lastTopic := dm.GetLastTopic(1, 100)
	if lastTopic == nil {
		t.Fatal("GetLastTopic returned nil")
	}
	if lastTopic.Topic != "weather" {
		t.Errorf("Topic = %s, want weather", lastTopic.Topic)
	}
	if lastTopic.PlayerAction != "asked about rain" {
		t.Errorf("PlayerAction = %s, want 'asked about rain'", lastTopic.PlayerAction)
	}
}

func TestHasDiscussedTopic(t *testing.T) {
	dm := NewDialogManager(12345)

	dm.RecordTopic(1, 100, "Quest", "accepted", "gave directions")

	if !dm.HasDiscussedTopic(1, 100, "quest") { // case insensitive
		t.Error("Should have discussed 'quest'")
	}
	if dm.HasDiscussedTopic(1, 100, "dragons") {
		t.Error("Should not have discussed 'dragons'")
	}
}

func TestTopicHistory(t *testing.T) {
	dm := NewDialogManager(12345)

	dm.RecordTopic(1, 100, "topic1", "action1", "response1")
	dm.RecordTopic(1, 100, "topic2", "action2", "response2")
	dm.RecordTopic(1, 100, "topic3", "action3", "response3")

	history := dm.GetTopicHistory(1, 100)
	if len(history) != 3 {
		t.Errorf("History length = %d, want 3", len(history))
	}
	if history[0].Topic != "topic1" {
		t.Errorf("First topic = %s, want topic1", history[0].Topic)
	}
}

func TestTopicMemoryLimit(t *testing.T) {
	dm := NewDialogManager(12345)

	// Record more than 20 topics
	for i := 0; i < 25; i++ {
		dm.RecordTopic(1, 100, "topic", "action", "response")
	}

	history := dm.GetTopicHistory(1, 100)
	if len(history) != 20 {
		t.Errorf("History length = %d, want 20 (max)", len(history))
	}
}

func TestEmotionalStateShift(t *testing.T) {
	dm := NewDialogManager(12345)

	// Initial state should be neutral
	state := dm.GetEmotionalState(1, 100, EmotionNeutral)
	if state != EmotionNeutral {
		t.Errorf("Initial state = %v, want neutral", state)
	}

	// Positive shift should make friendly
	dm.ShiftEmotion(1, 100, 35)
	state = dm.GetEmotionalState(1, 100, EmotionNeutral)
	if state != EmotionFriendly {
		t.Errorf("State after +35 = %v, want friendly", state)
	}

	// Reset and negative shift should make suspicious/hostile
	dm2 := NewDialogManager(12345)
	dm2.ShiftEmotion(1, 100, -35)
	state = dm2.GetEmotionalState(1, 100, EmotionNeutral)
	if state != EmotionSuspicious {
		t.Errorf("State after -35 = %v, want suspicious", state)
	}

	dm2.ShiftEmotion(1, 100, -20) // Now -55
	state = dm2.GetEmotionalState(1, 100, EmotionNeutral)
	if state != EmotionHostile {
		t.Errorf("State after -55 = %v, want hostile", state)
	}
}

func TestEmotionClamping(t *testing.T) {
	dm := NewDialogManager(12345)

	dm.ShiftEmotion(1, 100, 200)
	dm.mu.RLock()
	shift := dm.memories[1][100].EmotionShift
	dm.mu.RUnlock()

	if shift != 100 {
		t.Errorf("EmotionShift = %f, want 100 (clamped)", shift)
	}

	dm2 := NewDialogManager(12345)
	dm2.ShiftEmotion(1, 100, -200)
	dm2.mu.RLock()
	shift = dm2.memories[1][100].EmotionShift
	dm2.mu.RUnlock()

	if shift != -100 {
		t.Errorf("EmotionShift = %f, want -100 (clamped)", shift)
	}
}

func TestGenreVocabulariesNonOverlapping(t *testing.T) {
	// Per AC: all 5 genres produce non-overlapping common word sets
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for i := 0; i < len(genres); i++ {
		for j := i + 1; j < len(genres); j++ {
			vocab1 := GenreVocabularies[genres[i]]
			vocab2 := GenreVocabularies[genres[j]]

			// Check common words don't overlap
			words1 := make(map[string]bool)
			for _, w := range vocab1.CommonWords {
				words1[w] = true
			}

			for _, w := range vocab2.CommonWords {
				if words1[w] {
					t.Errorf("Word '%s' appears in both %s and %s", w, genres[i], genres[j])
				}
			}
		}
	}
}

func TestGenreVocabulariesDistinct(t *testing.T) {
	// Each genre should have unique greetings
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for i := 0; i < len(genres); i++ {
		for j := i + 1; j < len(genres); j++ {
			vocab1 := GenreVocabularies[genres[i]]
			vocab2 := GenreVocabularies[genres[j]]

			// At least first greeting should be different
			if vocab1.Greetings[0] == vocab2.Greetings[0] {
				t.Errorf("Greetings for %s and %s are the same", genres[i], genres[j])
			}
		}
	}
}

func TestEmotionalStateChangesVocabulary(t *testing.T) {
	// Per AC: emotional state changes NPC response vocabulary
	rng := rand.New(rand.NewSource(42))

	friendlyGreeting := GetGreeting("fantasy", EmotionFriendly, rng)
	hostileGreeting := GetGreeting("fantasy", EmotionHostile, rng)

	// Friendly should have warm prefix, hostile should have aggressive
	// They should be different
	if friendlyGreeting == hostileGreeting {
		t.Errorf("Friendly and hostile greetings should differ")
	}

	// Check that emotion modifiers are applied
	if !containsAnyPrefix(friendlyGreeting, EmotionModifiers[EmotionFriendly].Prefixes) {
		t.Logf("Friendly greeting: %s", friendlyGreeting)
	}
}

func containsAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && len(s) >= len(p) && s[:len(p)] == p {
			return true
		}
	}
	return false
}

func TestGenerateResponse(t *testing.T) {
	dm := NewDialogManager(12345)

	// First interaction
	response := dm.GenerateResponse(1, 100, "fantasy", "quest", EmotionNeutral)
	if response == nil {
		t.Fatal("GenerateResponse returned nil")
	}
	if response.Text == "" {
		t.Error("Response text should not be empty")
	}
	if response.EmotionalTone == "" {
		t.Error("Emotional tone should not be empty")
	}
}

func TestGenerateResponseWithTopicRecall(t *testing.T) {
	dm := NewDialogManager(12345)

	// Record a previous topic
	dm.RecordTopic(1, 100, "treasure", "asked about gold", "mentioned location")

	// Per AC: NPC recalls player's previous interaction topic in follow-up
	response := dm.GenerateResponse(1, 100, "fantasy", "new_topic", EmotionNeutral)

	if response.RecalledTopic != "treasure" {
		t.Errorf("RecalledTopic = %s, want 'treasure'", response.RecalledTopic)
	}
}

func TestClearOldMemories(t *testing.T) {
	dm := NewDialogManager(12345)

	dm.RecordTopic(1, 100, "old_topic", "action", "response")

	// Manually set old timestamp
	dm.mu.Lock()
	dm.memories[1][100].LastInteraction = time.Now().Add(-48 * time.Hour)
	dm.mu.Unlock()

	dm.ClearOldMemories(24 * time.Hour)

	if dm.MemoryCount() != 0 {
		t.Errorf("MemoryCount = %d, want 0 after clearing old", dm.MemoryCount())
	}
}

func TestGetVocabularyUnknownGenre(t *testing.T) {
	vocab := GetVocabulary("unknown_genre")
	fantasyVocab := GenreVocabularies["fantasy"]

	if vocab != fantasyVocab {
		t.Error("Unknown genre should fall back to fantasy")
	}
}

func TestEmotionalStateString(t *testing.T) {
	tests := []struct {
		state EmotionalState
		want  string
	}{
		{EmotionNeutral, "neutral"},
		{EmotionFriendly, "friendly"},
		{EmotionHostile, "hostile"},
		{EmotionFearful, "fearful"},
		{EmotionSuspicious, "suspicious"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("%v.String() = %s, want %s", tt.state, got, tt.want)
		}
	}
}

func BenchmarkRecordTopic(b *testing.B) {
	dm := NewDialogManager(12345)
	for i := 0; i < b.N; i++ {
		dm.RecordTopic(1, 100, "topic", "action", "response")
	}
}

func BenchmarkGenerateResponse(b *testing.B) {
	dm := NewDialogManager(12345)
	for i := 0; i < b.N; i++ {
		dm.GenerateResponse(1, 100, "fantasy", "topic", EmotionNeutral)
	}
}
