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
