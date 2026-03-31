package audio

import (
	"testing"
)

func TestUISoundGenerator_NewUISoundGenerator(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	if gen == nil {
		t.Fatal("expected non-nil generator")
	}
	if gen.engine != engine {
		t.Error("generator should reference provided engine")
	}
}

func TestUISoundGenerator_Generate_AllTypes(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	soundTypes := []struct {
		name      string
		soundType UISoundType
	}{
		{"MenuSelect", UISoundMenuSelect},
		{"MenuNavigate", UISoundMenuNavigate},
		{"ButtonClick", UISoundButtonClick},
		{"ButtonHover", UISoundButtonHover},
		{"Notification", UISoundNotification},
		{"Error", UISoundError},
		{"Success", UISoundSuccess},
		{"InventoryOpen", UISoundInventoryOpen},
		{"InventoryClose", UISoundInventoryClose},
		{"ItemPickup", UISoundItemPickup},
		{"ItemDrop", UISoundItemDrop},
		{"GoldCoins", UISoundGoldCoins},
		{"LevelUp", UISoundLevelUp},
		{"QuestComplete", UISoundQuestComplete},
		{"QuestAccept", UISoundQuestAccept},
		{"MapOpen", UISoundMapOpen},
		{"DialogAdvance", UISoundDialogAdvance},
	}

	for _, genre := range genres {
		engine := NewEngine(genre)
		gen := NewUISoundGenerator(engine)

		for _, st := range soundTypes {
			t.Run(genre+"_"+st.name, func(t *testing.T) {
				samples := gen.Generate(st.soundType)

				if len(samples) == 0 {
					t.Errorf("expected non-empty samples for %s in %s genre", st.name, genre)
					return
				}

				// Verify sample values are in valid range
				for i, s := range samples {
					if s < -2.0 || s > 2.0 {
						t.Errorf("sample %d out of expected range: %f", i, s)
						break
					}
				}
			})
		}
	}
}

func TestUISoundGenerator_Generate_DefaultCase(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	// Test with an invalid sound type
	invalidType := UISoundType(999)
	samples := gen.Generate(invalidType)

	// Should fall back to button click
	if len(samples) == 0 {
		t.Error("expected default fallback to produce samples")
	}
}

func TestUISoundGenerator_GetUIBaseFrequency(t *testing.T) {
	testCases := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", 523.25},
		{"sci-fi", 659.26},
		{"horror", 311.13},
		{"cyberpunk", 783.99},
		{"post-apocalyptic", 392.00},
		{"unknown", 523.25},
	}

	for _, tc := range testCases {
		t.Run(tc.genre, func(t *testing.T) {
			engine := NewEngine(tc.genre)
			gen := NewUISoundGenerator(engine)
			freq := gen.getUIBaseFrequency()

			if freq != tc.expected {
				t.Errorf("expected %f, got %f for genre %s", tc.expected, freq, tc.genre)
			}
		})
	}
}

func TestUISoundGenerator_GenerateUITone(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	testCases := []struct {
		frequency float64
		duration  float64
	}{
		{440.0, 0.1},
		{880.0, 0.05},
		{220.0, 0.2},
	}

	for _, tc := range testCases {
		samples := gen.generateUITone(tc.frequency, tc.duration)
		expectedLen := int(tc.duration * float64(engine.SampleRate))

		if len(samples) != expectedLen {
			t.Errorf("expected %d samples, got %d for frequency %f duration %f",
				expectedLen, len(samples), tc.frequency, tc.duration)
		}
	}
}

func TestUISoundGenerator_GenerateNoise(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	duration := 0.1
	samples := gen.generateNoise(duration)
	expectedLen := int(duration * float64(engine.SampleRate))

	if len(samples) != expectedLen {
		t.Errorf("expected %d samples, got %d", expectedLen, len(samples))
	}

	// Noise should have variation
	allSame := true
	first := samples[0]
	for _, s := range samples[1:] {
		if s != first {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("noise samples should not all be the same value")
	}
}

func TestUISoundGenerator_Concatenate(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	a := []float64{1.0, 2.0, 3.0}
	b := []float64{4.0, 5.0}

	result := gen.concatenate(a, b)

	if len(result) != 5 {
		t.Errorf("expected length 5, got %d", len(result))
	}

	expected := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("at index %d: expected %f, got %f", i, expected[i], v)
		}
	}
}

func TestUISoundGenerator_MixSamples(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	a := []float64{1.0, 0.0, 1.0}
	b := []float64{0.0, 1.0, 0.0}

	result := gen.mixSamples(a, b, 0.5, 0.5)

	if len(result) != 3 {
		t.Errorf("expected length 3, got %d", len(result))
	}

	// All values should be 0.5 (50% of each)
	for i, v := range result {
		if v != 0.5 {
			t.Errorf("at index %d: expected 0.5, got %f", i, v)
		}
	}
}

func TestUISoundGenerator_MixSamples_DifferentLengths(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	a := []float64{1.0, 1.0, 1.0, 1.0}
	b := []float64{1.0, 1.0}

	result := gen.mixSamples(a, b, 0.5, 0.5)

	if len(result) != 4 {
		t.Errorf("expected length 4 (max of inputs), got %d", len(result))
	}

	// First two should mix, last two should only have a's contribution
	if result[0] != 1.0 {
		t.Errorf("index 0: expected 1.0, got %f", result[0])
	}
	if result[3] != 0.5 {
		t.Errorf("index 3: expected 0.5 (only a's contribution), got %f", result[3])
	}
}

func TestUISoundGenerator_ApplyClickEnvelope(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	samples := []float64{1.0, 1.0, 1.0, 1.0, 1.0}
	result := gen.applyClickEnvelope(samples)

	// First sample should be near 1.0, later samples should decay
	if result[0] < 0.8 {
		t.Errorf("first sample should be near 1.0, got %f", result[0])
	}
	if result[len(result)-1] >= result[0] {
		t.Error("envelope should decay over time")
	}
}

func TestUISoundGenerator_ApplyFadeEnvelope(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	samples := []float64{1.0, 1.0, 1.0, 1.0, 1.0}
	result := gen.applyFadeEnvelope(samples, 0.8)

	// First sample should be at volume (0.8), later should fade
	if result[0] < 0.7 || result[0] > 0.85 {
		t.Errorf("first sample should be near 0.8, got %f", result[0])
	}
	// Should fade towards end - last sample should be less than first
	if result[len(result)-1] >= result[0] {
		t.Error("last sample should be less than first (fading)")
	}
}

func TestUISoundGenerator_ApplySweep(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	samples := []float64{1.0, 1.0, 1.0, 1.0, 1.0}
	result := gen.applySweep(samples, 0.5, 1.5)

	// First sample should be 0.5
	if result[0] != 0.5 {
		t.Errorf("first sample should be 0.5, got %f", result[0])
	}
	// Last sample should be greater than first (sweeping up)
	if result[len(result)-1] <= result[0] {
		t.Errorf("last sample should be greater than first in upward sweep, got first=%f last=%f", result[0], result[len(result)-1])
	}
}

func TestUISoundGenerator_ApplyMetallicSheen(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	samples := []float64{0.5, 0.5, 0.5, 0.5, 0.5}
	result := gen.applyMetallicSheen(samples)

	// Result should be modified (not all identical)
	allSame := true
	first := result[0]
	for _, s := range result[1:] {
		if s != first {
			allSame = false
			break
		}
	}
	if allSame && len(result) > 1 {
		t.Error("metallic sheen should add harmonics/variation")
	}
}

func TestUISoundGenerator_ApplyBandpass(t *testing.T) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	// Generate noise
	samples := gen.generateNoise(0.1)
	result := gen.applyBandpass(samples, 500.0, 100.0)

	if len(result) != len(samples) {
		t.Errorf("bandpass should preserve length: expected %d, got %d", len(samples), len(result))
	}
}

func TestUISoundPlayer_NewUISoundPlayer(t *testing.T) {
	engine := NewEngine("fantasy")
	player := NewUISoundPlayer(engine)

	if player == nil {
		t.Fatal("expected non-nil player")
	}
	if player.generator == nil {
		t.Error("player should have a generator")
	}
	if player.minInterval != 0.05 {
		t.Errorf("expected default minInterval 0.05, got %f", player.minInterval)
	}
}

func TestUISoundPlayer_CanPlay(t *testing.T) {
	engine := NewEngine("fantasy")
	player := NewUISoundPlayer(engine)

	// First play should always be allowed
	if !player.CanPlay(UISoundMenuSelect, 0.0) {
		t.Error("first play should be allowed")
	}

	// Simulate playing
	player.GetSamples(UISoundMenuSelect, 0.0)

	// Immediate replay should be blocked
	if player.CanPlay(UISoundMenuSelect, 0.01) {
		t.Error("replay within minInterval should be blocked")
	}

	// After interval, should be allowed
	if !player.CanPlay(UISoundMenuSelect, 0.1) {
		t.Error("replay after minInterval should be allowed")
	}
}

func TestUISoundPlayer_GetSamples(t *testing.T) {
	engine := NewEngine("fantasy")
	player := NewUISoundPlayer(engine)

	samples := player.GetSamples(UISoundButtonClick, 0.0)
	if len(samples) == 0 {
		t.Error("expected non-empty samples")
	}

	// Immediate call should return nil (blocked)
	samples = player.GetSamples(UISoundButtonClick, 0.01)
	if samples != nil {
		t.Error("expected nil samples when blocked by rate limit")
	}
}

func TestUISoundPlayer_SetMinInterval(t *testing.T) {
	engine := NewEngine("fantasy")
	player := NewUISoundPlayer(engine)

	player.SetMinInterval(0.2)
	if player.minInterval != 0.2 {
		t.Errorf("expected 0.2, got %f", player.minInterval)
	}

	// Negative values should be ignored
	player.SetMinInterval(-0.1)
	if player.minInterval != 0.2 {
		t.Error("negative interval should be ignored")
	}
}

func TestUISoundPlayer_GetMinInterval(t *testing.T) {
	engine := NewEngine("fantasy")
	player := NewUISoundPlayer(engine)

	if player.GetMinInterval() != 0.05 {
		t.Errorf("expected default 0.05, got %f", player.GetMinInterval())
	}

	player.SetMinInterval(0.1)
	if player.GetMinInterval() != 0.1 {
		t.Errorf("expected 0.1, got %f", player.GetMinInterval())
	}
}

func TestUISoundPlayer_Reset(t *testing.T) {
	engine := NewEngine("fantasy")
	player := NewUISoundPlayer(engine)

	// Play a sound
	player.GetSamples(UISoundMenuSelect, 0.0)

	// Should be blocked immediately
	if player.CanPlay(UISoundMenuSelect, 0.01) {
		t.Error("should be blocked before reset")
	}

	// Reset and try again
	player.Reset()
	if !player.CanPlay(UISoundMenuSelect, 0.01) {
		t.Error("should be allowed after reset")
	}
}

func TestUISoundPlayer_DifferentSoundTypes(t *testing.T) {
	engine := NewEngine("fantasy")
	player := NewUISoundPlayer(engine)

	// Play one type
	player.GetSamples(UISoundMenuSelect, 0.0)

	// Different type should not be blocked
	if !player.CanPlay(UISoundButtonClick, 0.01) {
		t.Error("different sound types should not block each other")
	}
}

// Benchmark tests
func BenchmarkUISoundGenerator_MenuSelect(b *testing.B) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(UISoundMenuSelect)
	}
}

func BenchmarkUISoundGenerator_LevelUp(b *testing.B) {
	engine := NewEngine("fantasy")
	gen := NewUISoundGenerator(engine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(UISoundLevelUp)
	}
}

func BenchmarkUISoundGenerator_GoldCoins(b *testing.B) {
	engine := NewEngine("cyberpunk")
	gen := NewUISoundGenerator(engine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(UISoundGoldCoins)
	}
}

func BenchmarkUISoundPlayer_GetSamples(b *testing.B) {
	engine := NewEngine("fantasy")
	player := NewUISoundPlayer(engine)
	player.SetMinInterval(0) // Disable rate limiting for benchmark

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player.GetSamples(UISoundButtonClick, float64(i))
	}
}
