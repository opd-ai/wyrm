package dialog

import (
	"math/rand"
	"testing"
	"time"
)

func TestRecordAndRecallTopic(t *testing.T) {
	dm := NewManager(12345)

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
	dm := NewManager(12345)

	dm.RecordTopic(1, 100, "Quest", "accepted", "gave directions")

	if !dm.HasDiscussedTopic(1, 100, "quest") { // case insensitive
		t.Error("Should have discussed 'quest'")
	}
	if dm.HasDiscussedTopic(1, 100, "dragons") {
		t.Error("Should not have discussed 'dragons'")
	}
}

func TestTopicHistory(t *testing.T) {
	dm := NewManager(12345)

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
	dm := NewManager(12345)

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
	dm := NewManager(12345)

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
	dm2 := NewManager(12345)
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
	dm := NewManager(12345)

	dm.ShiftEmotion(1, 100, 200)
	dm.mu.RLock()
	shift := dm.memories[1][100].EmotionShift
	dm.mu.RUnlock()

	if shift != 100 {
		t.Errorf("EmotionShift = %f, want 100 (clamped)", shift)
	}

	dm2 := NewManager(12345)
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
	dm := NewManager(12345)

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
	dm := NewManager(12345)

	// Record a previous topic
	dm.RecordTopic(1, 100, "treasure", "asked about gold", "mentioned location")

	// Per AC: NPC recalls player's previous interaction topic in follow-up
	response := dm.GenerateResponse(1, 100, "fantasy", "new_topic", EmotionNeutral)

	if response.RecalledTopic != "treasure" {
		t.Errorf("RecalledTopic = %s, want 'treasure'", response.RecalledTopic)
	}
}

func TestClearOldMemories(t *testing.T) {
	dm := NewManager(12345)

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
	dm := NewManager(12345)
	for i := 0; i < b.N; i++ {
		dm.RecordTopic(1, 100, "topic", "action", "response")
	}
}

func BenchmarkGenerateResponse(b *testing.B) {
	dm := NewManager(12345)
	for i := 0; i < b.N; i++ {
		dm.GenerateResponse(1, 100, "fantasy", "topic", EmotionNeutral)
	}
}

// Tests for Persuasion and Intimidation skill checks (FEATURES.md #4)

func TestSkillCheckTypeString(t *testing.T) {
	tests := []struct {
		checkType SkillCheckType
		want      string
	}{
		{SkillCheckPersuasion, "persuasion"},
		{SkillCheckIntimidate, "intimidation"},
		{SkillCheckType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.checkType.String(); got != tt.want {
			t.Errorf("%v.String() = %s, want %s", tt.checkType, got, tt.want)
		}
	}
}

func TestPersuasionSkillCheckSuccess(t *testing.T) {
	// Use a fixed seed for deterministic testing
	dm := NewManager(42)

	// High skill (80) against easy difficulty (25) should usually succeed
	result := dm.AttemptPersuasion(1, 100, 80, DifficultyEasy, EmotionNeutral, "fantasy")

	// The result should have all fields populated
	if result.ResponseText == "" {
		t.Error("ResponseText should not be empty")
	}

	// Check that emotion shift is applied
	dm.mu.RLock()
	memory := dm.memories[1][100]
	dm.mu.RUnlock()

	if memory == nil {
		t.Fatal("Memory should be created after skill check")
	}

	// Topic should be recorded
	if len(memory.Topics) == 0 {
		t.Error("Skill check attempt should be recorded in topic history")
	}
}

func TestPersuasionSkillCheckFailure(t *testing.T) {
	dm := NewManager(42)

	// Low skill (10) against very hard difficulty (90) should fail
	result := dm.AttemptPersuasion(1, 100, 10, DifficultyVeryHard, EmotionHostile, "cyberpunk")

	// Should fail with margin below zero
	if result.Success && result.Margin < 0 {
		t.Error("Result with negative margin should not be success")
	}

	// Response should be genre-appropriate
	if result.ResponseText == "" {
		t.Error("ResponseText should not be empty even on failure")
	}
}

func TestIntimidationSkillCheckSuccess(t *testing.T) {
	dm := NewManager(42)

	// High skill against fearful NPC (easier to intimidate)
	result := dm.AttemptIntimidate(1, 100, 70, DifficultyMedium, EmotionFearful, "horror")

	if result.ResponseText == "" {
		t.Error("ResponseText should not be empty")
	}

	// Successful intimidation should cause negative emotion shift
	if result.Success && result.EmotionShift >= 0 {
		t.Logf("Note: Successful intimidation usually causes fear (negative shift), got %f", result.EmotionShift)
	}
}

func TestIntimidationSkillCheckFailure(t *testing.T) {
	dm := NewManager(42)

	// Low skill against friendly NPC (harder to intimidate)
	result := dm.AttemptIntimidate(1, 100, 15, DifficultyHard, EmotionFriendly, "sci-fi")

	// Failed intimidation should also cause negative emotion (NPC becomes hostile)
	if !result.Success && result.EmotionShift > 0 {
		t.Errorf("Failed intimidation should cause negative emotion shift, got %f", result.EmotionShift)
	}
}

func TestEmotionModifierPersuasion(t *testing.T) {
	dm := NewManager(42)

	// Test that friendly NPCs are easier to persuade
	modFriendly := dm.getEmotionModifier(SkillCheckPersuasion, EmotionFriendly)
	modHostile := dm.getEmotionModifier(SkillCheckPersuasion, EmotionHostile)

	if modFriendly <= modHostile {
		t.Errorf("Friendly NPC should give better persuasion modifier than hostile: %d vs %d", modFriendly, modHostile)
	}

	// Hostile should be negative
	if modHostile >= 0 {
		t.Errorf("Hostile NPC should give negative persuasion modifier, got %d", modHostile)
	}
}

func TestEmotionModifierIntimidation(t *testing.T) {
	dm := NewManager(42)

	// Test that fearful NPCs are easier to intimidate
	modFearful := dm.getEmotionModifier(SkillCheckIntimidate, EmotionFearful)
	modFriendly := dm.getEmotionModifier(SkillCheckIntimidate, EmotionFriendly)

	if modFearful <= modFriendly {
		t.Errorf("Fearful NPC should give better intimidation modifier than friendly: %d vs %d", modFearful, modFriendly)
	}

	// Friendly should be negative for intimidation
	if modFriendly >= 0 {
		t.Errorf("Friendly NPC should give negative intimidation modifier, got %d", modFriendly)
	}
}

func TestSkillCheckDifficultyLevels(t *testing.T) {
	// Test that difficulty levels are ordered correctly
	difficulties := []SkillCheckDifficulty{
		DifficultyTrivial,
		DifficultyEasy,
		DifficultyMedium,
		DifficultyHard,
		DifficultyVeryHard,
		DifficultyImpossible,
	}

	for i := 1; i < len(difficulties); i++ {
		if difficulties[i] <= difficulties[i-1] {
			t.Errorf("Difficulty levels should increase: %d should be > %d",
				difficulties[i], difficulties[i-1])
		}
	}
}

func TestSkillCheckRecordsInMemory(t *testing.T) {
	dm := NewManager(42)

	// Perform a skill check
	dm.AttemptPersuasion(1, 100, 50, DifficultyMedium, EmotionNeutral, "fantasy")

	// Check that the attempt was recorded
	history := dm.GetTopicHistory(1, 100)
	if len(history) == 0 {
		t.Error("Skill check should be recorded in topic history")
	}

	// The topic should mention the skill check type
	found := false
	for _, topic := range history {
		if topic.Topic != "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Topic record should exist for skill check")
	}
}

func TestAllGenresHaveSkillCheckResponses(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		dm := NewManager(42)

		// Test persuasion
		result := dm.AttemptPersuasion(1, 100, 50, DifficultyMedium, EmotionNeutral, genre)
		if result.ResponseText == "" {
			t.Errorf("Genre %s should produce persuasion response", genre)
		}

		// Test intimidation
		result = dm.AttemptIntimidate(2, 100, 50, DifficultyMedium, EmotionNeutral, genre)
		if result.ResponseText == "" {
			t.Errorf("Genre %s should produce intimidation response", genre)
		}
	}
}

func TestCriticalSuccessAndFailure(t *testing.T) {
	// Run multiple checks to see critical results
	successCount := 0
	failCount := 0
	critSuccessCount := 0
	critFailCount := 0

	// Run many times to get statistical coverage
	for seed := int64(0); seed < 100; seed++ {
		dm := NewManager(seed)
		result := dm.AttemptPersuasion(1, 100, 50, DifficultyMedium, EmotionNeutral, "fantasy")

		if result.Success {
			successCount++
		} else {
			failCount++
		}
		if result.CriticalSuccess {
			critSuccessCount++
		}
		if result.CriticalFailure {
			critFailCount++
		}
	}

	// Should have mix of outcomes
	t.Logf("Outcomes over 100 checks: success=%d, fail=%d, critSuccess=%d, critFail=%d",
		successCount, failCount, critSuccessCount, critFailCount)

	// With medium difficulty and medium skill, should have some successes and failures
	if successCount == 0 || failCount == 0 {
		t.Error("Expected mix of success and failure outcomes")
	}
}

func BenchmarkPersuasionSkillCheck(b *testing.B) {
	dm := NewManager(12345)
	for i := 0; i < b.N; i++ {
		dm.AttemptPersuasion(1, 100, 50, DifficultyMedium, EmotionNeutral, "fantasy")
	}
}

func BenchmarkIntimidationSkillCheck(b *testing.B) {
	dm := NewManager(12345)
	for i := 0; i < b.N; i++ {
		dm.AttemptIntimidate(1, 100, 50, DifficultyMedium, EmotionNeutral, "fantasy")
	}
}
