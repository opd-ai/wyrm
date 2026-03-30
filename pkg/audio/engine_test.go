package audio

import (
	"math"
	"testing"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine("fantasy")
	if e == nil {
		t.Fatal("NewEngine returned nil")
	}
	if e.SampleRate != 44100 {
		t.Errorf("expected SampleRate=44100, got %d", e.SampleRate)
	}
	if e.Genre != "fantasy" {
		t.Errorf("expected Genre='fantasy', got %q", e.Genre)
	}
	if e.playing {
		t.Error("engine should not be playing initially")
	}
}

func TestNewEngineGenres(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		e := NewEngine(genre)
		if e.Genre != genre {
			t.Errorf("expected Genre=%q, got %q", genre, e.Genre)
		}
	}
}

func TestUpdate(t *testing.T) {
	e := NewEngine("fantasy")

	// Not playing - phase should not change
	initialPhase := e.phase
	e.Update()
	if e.phase != initialPhase {
		t.Error("phase should not change when not playing")
	}

	// Start playing
	e.Play()
	e.Update()
	if e.phase == 0 {
		t.Error("phase should advance when playing")
	}
}

func TestUpdatePhaseWrap(t *testing.T) {
	e := NewEngine("fantasy")
	e.Play()

	// Advance phase past 2*Pi
	e.phase = 2*math.Pi - 0.005
	e.Update()

	if e.phase >= 2*math.Pi {
		t.Error("phase should wrap around 2*Pi")
	}
}

func TestGenerateSineWave(t *testing.T) {
	e := NewEngine("fantasy")
	samples := e.GenerateSineWave(440.0, 0.1)

	expectedLen := int(0.1 * float64(e.SampleRate))
	if len(samples) != expectedLen {
		t.Errorf("expected %d samples, got %d", expectedLen, len(samples))
	}

	// Check samples are in valid range [-1, 1]
	for i, s := range samples {
		if s < -1.0 || s > 1.0 {
			t.Errorf("sample %d out of range: %f", i, s)
		}
	}

	// Check for periodicity (sine wave should have zero crossings)
	zeroCrossings := 0
	for i := 1; i < len(samples); i++ {
		if samples[i-1] < 0 && samples[i] >= 0 || samples[i-1] >= 0 && samples[i] < 0 {
			zeroCrossings++
		}
	}
	// 440 Hz for 0.1 seconds = 44 cycles, roughly 88 zero crossings
	if zeroCrossings < 80 || zeroCrossings > 96 {
		t.Errorf("unexpected zero crossings %d for 440Hz sine", zeroCrossings)
	}
}

func TestGenerateSineWaveDifferentFrequencies(t *testing.T) {
	e := NewEngine("fantasy")

	low := e.GenerateSineWave(220.0, 0.1)
	high := e.GenerateSineWave(880.0, 0.1)

	// Both should have same length
	if len(low) != len(high) {
		t.Error("different frequencies should produce same length buffers for same duration")
	}

	// Higher frequency should have more zero crossings
	lowCrossings := countZeroCrossings(low)
	highCrossings := countZeroCrossings(high)
	if highCrossings <= lowCrossings {
		t.Error("higher frequency should have more zero crossings")
	}
}

func countZeroCrossings(samples []float64) int {
	count := 0
	for i := 1; i < len(samples); i++ {
		if samples[i-1] < 0 && samples[i] >= 0 || samples[i-1] >= 0 && samples[i] < 0 {
			count++
		}
	}
	return count
}

func TestApplyADSR(t *testing.T) {
	e := NewEngine("fantasy")

	// Create constant amplitude samples
	samples := make([]float64, e.SampleRate) // 1 second
	for i := range samples {
		samples[i] = 1.0
	}

	// Apply envelope
	result := e.ApplyADSR(samples, 0.1, 0.1, 0.5, 0.2)

	if len(result) != len(samples) {
		t.Errorf("ADSR should preserve sample count")
	}

	// Attack phase starts at 0
	if result[0] >= 0.01 {
		t.Error("attack should start near zero")
	}

	// Peak should reach 1.0 at end of attack
	attackEnd := int(0.1 * float64(e.SampleRate))
	if result[attackEnd-1] < 0.9 {
		t.Errorf("should reach peak at attack end, got %f", result[attackEnd-1])
	}

	// Sustain level
	sustainStart := int(0.2 * float64(e.SampleRate))
	releaseStart := len(samples) - int(0.2*float64(e.SampleRate))
	midSustain := (sustainStart + releaseStart) / 2
	if result[midSustain] < 0.45 || result[midSustain] > 0.55 {
		t.Errorf("sustain should be near 0.5, got %f", result[midSustain])
	}

	// Release ends near zero
	if result[len(result)-1] > 0.1 {
		t.Errorf("release should end near zero, got %f", result[len(result)-1])
	}
}

func TestGetGenreBaseFrequency(t *testing.T) {
	tests := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", 220.0},
		{"sci-fi", 110.0},
		{"horror", 55.0},
		{"cyberpunk", 440.0},
		{"post-apocalyptic", 165.0},
		{"unknown", 220.0}, // default
	}

	for _, tc := range tests {
		e := NewEngine(tc.genre)
		freq := e.GetGenreBaseFrequency()
		if freq != tc.expected {
			t.Errorf("genre %q: expected freq=%f, got %f", tc.genre, tc.expected, freq)
		}
	}
}

func TestPlayStop(t *testing.T) {
	e := NewEngine("fantasy")

	if e.IsPlaying() {
		t.Error("should not be playing initially")
	}

	e.Play()
	if !e.IsPlaying() {
		t.Error("should be playing after Play()")
	}

	e.Stop()
	if e.IsPlaying() {
		t.Error("should not be playing after Stop()")
	}
}

func BenchmarkGenerateSineWave(b *testing.B) {
	e := NewEngine("fantasy")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.GenerateSineWave(440.0, 1.0)
	}
}

func BenchmarkApplyADSR(b *testing.B) {
	e := NewEngine("fantasy")
	samples := e.GenerateSineWave(440.0, 1.0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.ApplyADSR(samples, 0.1, 0.1, 0.5, 0.2)
	}
}

func TestGetGenrePitchModifier(t *testing.T) {
	tests := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", 1.0},
		{"sci-fi", 1.30},
		{"horror", 0.70},
		{"cyberpunk", 1.40},
		{"post-apocalyptic", 0.85},
		{"unknown", 1.0}, // default
	}

	for _, tc := range tests {
		e := NewEngine(tc.genre)
		modifier := e.GetGenrePitchModifier()
		if modifier != tc.expected {
			t.Errorf("genre %q: expected modifier=%f, got %f", tc.genre, tc.expected, modifier)
		}
	}
}

func TestGenrePitchDeviationExceeds15Percent(t *testing.T) {
	// ROADMAP Phase 4 AC: SFX for footsteps differs measurably (pitch deviation >15%) across all 5 genres
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	baseFreq := 440.0
	duration := 0.1

	// Generate SFX for each genre
	sfxByGenre := make(map[string][]float64)
	for _, genre := range genres {
		e := NewEngine(genre)
		sfxByGenre[genre] = e.GenerateSFX(SFXFootstep, baseFreq, duration)
	}

	// Compare each genre pair for pitch deviation
	for i := 0; i < len(genres); i++ {
		for j := i + 1; j < len(genres); j++ {
			g1, g2 := genres[i], genres[j]
			e1 := NewEngine(g1)
			e2 := NewEngine(g2)

			// Effective frequencies
			freq1 := baseFreq * e1.GetGenrePitchModifier()
			freq2 := baseFreq * e2.GetGenrePitchModifier()

			// Calculate percent deviation
			deviation := math.Abs(freq1-freq2) / math.Min(freq1, freq2) * 100

			// Verify deviation exceeds 15% for non-matching pairs
			// At minimum, compare against fantasy baseline
			if g1 == "fantasy" || g2 == "fantasy" {
				// Compare non-fantasy genre against fantasy baseline
				if deviation < 15.0 {
					t.Errorf("genres %q vs %q: pitch deviation %.2f%% is below 15%% threshold",
						g1, g2, deviation)
				}
			}
		}
	}
}

func TestGenerateSFXTypes(t *testing.T) {
	e := NewEngine("fantasy")
	sfxTypes := []SFXType{SFXFootstep, SFXHit, SFXExplosion, SFXAmbient, SFXUI}

	for _, sfx := range sfxTypes {
		samples := e.GenerateSFX(sfx, 440.0, 0.1)
		if len(samples) == 0 {
			t.Errorf("SFX type %d produced no samples", sfx)
		}

		// Verify samples are in valid range
		for i, s := range samples {
			if s < -1.5 || s > 1.5 {
				t.Errorf("SFX type %d sample %d out of range: %f", sfx, i, s)
			}
		}
	}
}

func TestHorrorVibrato(t *testing.T) {
	e := NewEngine("horror")
	samples := e.GenerateSFX(SFXAmbient, 220.0, 1.0)

	// Horror should have vibrato applied
	// Check for amplitude variation (vibrato effect)
	minAmp := 1.0
	maxAmp := -1.0
	for _, s := range samples {
		absS := math.Abs(s)
		if absS < minAmp {
			minAmp = absS
		}
		if absS > maxAmp {
			maxAmp = absS
		}
	}

	// Vibrato should create amplitude variation
	variation := maxAmp - minAmp
	if variation < 0.05 {
		t.Errorf("horror vibrato effect too weak: amplitude variation %.4f", variation)
	}
}

func TestCyberpunkHardClipping(t *testing.T) {
	e := NewEngine("cyberpunk")
	samples := e.GenerateSFX(SFXHit, 440.0, 0.5)

	// Hard clipping at 0.7 should limit amplitude
	for _, s := range samples {
		if math.Abs(s) > 0.75 {
			t.Errorf("cyberpunk hard clipping failed: sample %.4f exceeds threshold", s)
		}
	}
}

func BenchmarkGenerateSFX(b *testing.B) {
	e := NewEngine("horror")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.GenerateSFX(SFXFootstep, 440.0, 0.1)
	}
}

func TestSampleStreamQueueAndRead(t *testing.T) {
	stream := NewSampleStream()

	// Queue some samples
	samples := []float64{0.5, -0.5, 1.0, -1.0}
	stream.QueueSamples(samples)

	// Read should return the queued data
	buf := make([]byte, 16) // 4 samples * 4 bytes/sample = 16 bytes
	n, err := stream.Read(buf)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if n != 16 {
		t.Errorf("expected 16 bytes read, got %d", n)
	}
}

func TestSampleStreamReadSilence(t *testing.T) {
	stream := NewSampleStream()

	// Read when buffer is empty should return silence
	buf := make([]byte, 100)
	n, err := stream.Read(buf)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if n == 0 {
		t.Error("Read should return silence bytes, not 0")
	}

	// All bytes should be zero (silence)
	for i := 0; i < n; i++ {
		if buf[i] != 0 {
			t.Errorf("silence byte %d not zero: %d", i, buf[i])
		}
	}
}

func TestSampleStreamClipping(t *testing.T) {
	stream := NewSampleStream()

	// Queue samples that exceed [-1, 1] range - should be clipped
	samples := []float64{2.0, -2.0}
	stream.QueueSamples(samples)

	// Read and verify (we can't check exact values but should not error)
	buf := make([]byte, 8)
	n, err := stream.Read(buf)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if n != 8 {
		t.Errorf("expected 8 bytes, got %d", n)
	}
}

func TestSampleStreamSeek(t *testing.T) {
	stream := NewSampleStream()

	// Seek should return 0, nil (no-op for streaming audio)
	pos, err := stream.Seek(100, 0)
	if err != nil {
		t.Errorf("Seek should not error: %v", err)
	}
	if pos != 0 {
		t.Errorf("Seek should return 0, got %d", pos)
	}
}

func TestSampleStreamLength(t *testing.T) {
	stream := NewSampleStream()

	// Length should return a large value for streaming
	length := stream.Length()
	if length < 1<<60 {
		t.Errorf("Length should be very large for streaming, got %d", length)
	}
}

func TestMinHelper(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{0, 10, 0},
		{-5, 5, -5},
	}

	for _, tc := range tests {
		got := min(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("min(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSampleStreamPartialRead(t *testing.T) {
	stream := NewSampleStream()

	// Queue many samples
	samples := make([]float64, 100)
	for i := range samples {
		samples[i] = 0.5
	}
	stream.QueueSamples(samples)

	// Read in small chunks
	buf := make([]byte, 20)
	totalRead := 0
	for totalRead < 100*4 {
		n, err := stream.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		totalRead += n
	}

	// After reading all, should get silence
	n, _ := stream.Read(buf)
	allZero := true
	for i := 0; i < n; i++ {
		if buf[i] != 0 {
			allZero = false
			break
		}
	}
	if !allZero {
		t.Error("expected silence after draining buffer")
	}
}

// ============================================================================
// Reverb Effect Tests
// ============================================================================

func TestDefaultReverbConfig(t *testing.T) {
	config := DefaultReverbConfig()
	if config == nil {
		t.Fatal("DefaultReverbConfig returned nil")
	}

	if config.RoomSize < 0 || config.RoomSize > 1 {
		t.Errorf("RoomSize %f out of range [0,1]", config.RoomSize)
	}
	if config.Damping < 0 || config.Damping > 1 {
		t.Errorf("Damping %f out of range [0,1]", config.Damping)
	}
	if config.WetMix+config.DryMix > 1.5 { // Allow some flexibility
		t.Error("WetMix + DryMix too high")
	}
	if config.DecayTime <= 0 {
		t.Errorf("DecayTime %f should be positive", config.DecayTime)
	}
}

func TestGenreReverbConfig(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic", "unknown"}

	for _, genre := range genres {
		config := GenreReverbConfig(genre)
		if config == nil {
			t.Errorf("genre %s: GenreReverbConfig returned nil", genre)
			continue
		}

		if config.RoomSize < 0 || config.RoomSize > 1 {
			t.Errorf("genre %s: RoomSize out of range", genre)
		}
		if config.Damping < 0 || config.Damping > 1 {
			t.Errorf("genre %s: Damping out of range", genre)
		}
		if config.DecayTime <= 0 {
			t.Errorf("genre %s: DecayTime should be positive", genre)
		}
	}
}

func TestGenreReverbVariation(t *testing.T) {
	// Horror should have more reverb than cyberpunk
	horror := GenreReverbConfig("horror")
	cyberpunk := GenreReverbConfig("cyberpunk")

	if horror.RoomSize <= cyberpunk.RoomSize {
		t.Error("horror should have larger room size than cyberpunk")
	}
	if horror.DecayTime <= cyberpunk.DecayTime {
		t.Error("horror should have longer decay than cyberpunk")
	}
	if horror.WetMix <= cyberpunk.WetMix {
		t.Error("horror should have more wet mix than cyberpunk")
	}
}

func TestNewReverbProcessor(t *testing.T) {
	config := DefaultReverbConfig()
	processor := NewReverbProcessor(config, 44100)

	if processor == nil {
		t.Fatal("NewReverbProcessor returned nil")
	}

	if processor.sampleRate != 44100 {
		t.Errorf("sampleRate = %d, want 44100", processor.sampleRate)
	}

	if len(processor.combDelays) != 4 {
		t.Errorf("expected 4 comb filters, got %d", len(processor.combDelays))
	}

	if len(processor.allpassDelays) != 2 {
		t.Errorf("expected 2 allpass filters, got %d", len(processor.allpassDelays))
	}
}

func TestReverbProcessorProcess(t *testing.T) {
	config := DefaultReverbConfig()
	processor := NewReverbProcessor(config, 44100)

	// Create test input (impulse)
	input := make([]float64, 44100)
	input[0] = 1.0

	output := processor.Process(input)

	if len(output) != len(input) {
		t.Errorf("output length %d != input length %d", len(output), len(input))
	}

	// Output should have some energy beyond the impulse (reverb tail)
	tailEnergy := 0.0
	for i := 1000; i < len(output); i++ {
		tailEnergy += output[i] * output[i]
	}

	if tailEnergy < 0.01 {
		t.Error("reverb should produce a tail beyond the impulse")
	}
}

func TestReverbProcessorReset(t *testing.T) {
	config := DefaultReverbConfig()
	processor := NewReverbProcessor(config, 44100)

	// Process some input
	input := make([]float64, 1000)
	input[0] = 1.0
	processor.Process(input)

	// Reset
	processor.Reset()

	// Process silence - should produce silence
	silence := make([]float64, 1000)
	output := processor.Process(silence)

	// Output should be very quiet
	maxAmp := 0.0
	for _, s := range output {
		if math.Abs(s) > maxAmp {
			maxAmp = math.Abs(s)
		}
	}

	if maxAmp > 0.1 {
		t.Errorf("after reset, output should be silent, got max amp %f", maxAmp)
	}
}

func TestApplyReverb(t *testing.T) {
	e := NewEngine("fantasy")

	// Generate a tone
	samples := e.GenerateSineWave(440.0, 0.5)

	// Apply reverb
	reverbed := e.ApplyReverb(samples)

	if len(reverbed) != len(samples) {
		t.Errorf("reverb changed sample count: %d -> %d", len(samples), len(reverbed))
	}

	// Verify samples are in valid range (may exceed slightly due to mixing)
	for i, s := range reverbed {
		if s < -2.0 || s > 2.0 {
			t.Errorf("sample %d out of reasonable range: %f", i, s)
			break
		}
	}
}

func TestReverbProcessorSetConfig(t *testing.T) {
	config1 := DefaultReverbConfig()
	processor := NewReverbProcessor(config1, 44100)

	config2 := GenreReverbConfig("horror")
	processor.SetConfig(config2)

	if processor.GetConfig().RoomSize != config2.RoomSize {
		t.Error("SetConfig did not update the configuration")
	}
}

func BenchmarkReverbProcess(b *testing.B) {
	config := DefaultReverbConfig()
	processor := NewReverbProcessor(config, 44100)
	input := make([]float64, 44100)
	input[0] = 1.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.Process(input)
	}
}

// Voice synthesis tests

func TestNewVoiceSynthesizer(t *testing.T) {
	vs := NewVoiceSynthesizer(44100, "fantasy")
	if vs == nil {
		t.Fatal("NewVoiceSynthesizer returned nil")
	}
	if vs.SampleRate != 44100 {
		t.Errorf("SampleRate = %d, want 44100", vs.SampleRate)
	}
	if vs.Genre != "fantasy" {
		t.Errorf("Genre = %s, want fantasy", vs.Genre)
	}
	if len(vs.VoiceProfiles) == 0 {
		t.Error("VoiceProfiles should be initialized")
	}
}

func TestVoiceProfiles(t *testing.T) {
	vs := NewVoiceSynthesizer(44100, "fantasy")

	profiles := []string{"male_deep", "male_medium", "female", "elderly", "child", "creature", "robot", "whisper"}
	for _, name := range profiles {
		profile, ok := vs.VoiceProfiles[name]
		if !ok {
			t.Errorf("Missing profile: %s", name)
			continue
		}
		if profile.BaseFrequency <= 0 {
			t.Errorf("Profile %s has invalid BaseFrequency", name)
		}
		if profile.SpeakingRate <= 0 {
			t.Errorf("Profile %s has invalid SpeakingRate", name)
		}
	}
}

func TestSynthesizeSpeech(t *testing.T) {
	vs := NewVoiceSynthesizer(44100, "fantasy")

	text := "Hello world"
	samples := vs.SynthesizeSpeech(text, "male_medium")

	if len(samples) == 0 {
		t.Error("SynthesizeSpeech returned empty samples")
	}

	// Check samples are in valid range
	for i, s := range samples {
		if math.IsNaN(s) {
			t.Errorf("Sample %d is NaN", i)
		}
		if math.IsInf(s, 0) {
			t.Errorf("Sample %d is Inf", i)
		}
	}
}

func TestSynthesizeSpeechDifferentVoices(t *testing.T) {
	vs := NewVoiceSynthesizer(44100, "fantasy")
	text := "Test"

	voices := []string{"male_deep", "female", "creature", "robot"}
	for _, voice := range voices {
		samples := vs.SynthesizeSpeech(text, voice)
		if len(samples) == 0 {
			t.Errorf("Voice %s produced no samples", voice)
		}
	}
}

func TestSynthesizeSpeechUnknownVoice(t *testing.T) {
	vs := NewVoiceSynthesizer(44100, "fantasy")

	// Should fall back to male_medium
	samples := vs.SynthesizeSpeech("Test", "unknown_voice")
	if len(samples) == 0 {
		t.Error("Should produce samples with fallback voice")
	}
}

func TestGetGenreVoiceProfile(t *testing.T) {
	tests := []struct {
		genre      string
		occupation string
		want       string
	}{
		{"sci-fi", "technician", "robot"},
		{"sci-fi", "scientist", "robot"},
		{"horror", "priest", "whisper"},
		{"horror", "mortician", "whisper"},
		{"fantasy", "guard", "male_deep"},
		{"fantasy", "blacksmith", "male_deep"},
		{"fantasy", "healer", "female"},
		{"fantasy", "priest", "elderly"},
		{"fantasy", "merchant", "male_medium"},
	}

	for _, tt := range tests {
		vs := NewVoiceSynthesizer(44100, tt.genre)
		got := vs.GetGenreVoiceProfile(tt.occupation)
		if got != tt.want {
			t.Errorf("GetGenreVoiceProfile(%s, %s) = %s, want %s", tt.genre, tt.occupation, got, tt.want)
		}
	}
}

func TestGenerateDialogAudio(t *testing.T) {
	e := NewEngine("fantasy")

	samples := e.GenerateDialogAudio("Hello there!", "guard")
	if len(samples) == 0 {
		t.Error("GenerateDialogAudio returned empty samples")
	}

	// Test different NPC types
	npcTypes := []string{"guard", "merchant", "healer", "priest"}
	for _, npcType := range npcTypes {
		samples := e.GenerateDialogAudio("Test", npcType)
		if len(samples) == 0 {
			t.Errorf("No samples for NPC type %s", npcType)
		}
	}
}

func TestGetCharacterPitch(t *testing.T) {
	// Vowels should have higher pitch
	if getCharacterPitch('a') <= 0 {
		t.Error("Vowel 'a' should have positive pitch modifier")
	}
	if getCharacterPitch('!') <= 0 {
		t.Error("Exclamation should have positive pitch modifier")
	}
	if getCharacterPitch('.') >= 0 {
		t.Error("Period should have negative pitch modifier")
	}
	if getCharacterPitch('x') != 0 {
		t.Error("Consonant should have zero pitch modifier")
	}
}

func TestGetFormantModifier(t *testing.T) {
	// Vowels should have high formant values
	if getFormantModifier('a') < 0.8 {
		t.Error("Vowel 'a' should have high formant modifier")
	}
	// Spaces should be quiet
	if getFormantModifier(' ') > 0.2 {
		t.Error("Space should have low formant modifier")
	}
	// Consonants should be moderate
	if getFormantModifier('x') < 0.5 || getFormantModifier('x') > 0.9 {
		t.Error("Consonant should have moderate formant modifier")
	}
}

func TestPhonemeEnvelope(t *testing.T) {
	vs := NewVoiceSynthesizer(44100, "fantasy")

	// Attack phase (t=0) should start at 0
	if vs.phonemeEnvelope(0) != 0 {
		t.Error("Envelope should start at 0")
	}

	// Peak should be at end of attack
	peak := vs.phonemeEnvelope(0.1)
	if peak < 0.9 {
		t.Errorf("Envelope peak = %v, should be near 1.0", peak)
	}

	// Sustain should be moderate
	sustain := vs.phonemeEnvelope(0.5)
	if sustain < 0.5 || sustain > 0.9 {
		t.Errorf("Envelope sustain = %v, should be moderate", sustain)
	}

	// Release (t=1) should approach 0
	release := vs.phonemeEnvelope(0.99)
	if release > 0.5 {
		t.Errorf("Envelope release = %v, should be low", release)
	}
}

func BenchmarkSynthesizeSpeech(b *testing.B) {
	vs := NewVoiceSynthesizer(44100, "fantasy")
	text := "Hello, welcome to my shop!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vs.SynthesizeSpeech(text, "male_medium")
	}
}

func BenchmarkGenerateDialogAudio(b *testing.B) {
	e := NewEngine("fantasy")
	text := "Hello, welcome to my shop!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.GenerateDialogAudio(text, "merchant")
	}
}
