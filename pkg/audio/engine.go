// Package audio provides procedural audio synthesis.
package audio

import (
	"math"
)

// Engine handles procedural audio generation and playback.
type Engine struct {
	SampleRate int
	Genre      string
	phase      float64
	playing    bool
}

// NewEngine creates a new audio engine.
func NewEngine(genre string) *Engine {
	return &Engine{
		SampleRate: 44100,
		Genre:      genre,
		phase:      0,
		playing:    false,
	}
}

// Update advances the audio engine by one tick.
func (e *Engine) Update() {
	if e.playing {
		e.phase += 0.01
		if e.phase > 2*math.Pi {
			e.phase -= 2 * math.Pi
		}
	}
}

// GenerateSineWave generates a sine wave buffer at the given frequency.
func (e *Engine) GenerateSineWave(frequency, duration float64) []float64 {
	numSamples := int(duration * float64(e.SampleRate))
	samples := make([]float64, numSamples)
	phaseIncrement := 2 * math.Pi * frequency / float64(e.SampleRate)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		samples[i] = math.Sin(phase)
		phase += phaseIncrement
		if phase > 2*math.Pi {
			phase -= 2 * math.Pi
		}
	}
	return samples
}

// ApplyADSR applies an Attack-Decay-Sustain-Release envelope to samples.
func (e *Engine) ApplyADSR(samples []float64, attack, decay, sustain, release float64) []float64 {
	result := make([]float64, len(samples))
	sampleRate := float64(e.SampleRate)

	attackSamples := int(attack * sampleRate)
	decaySamples := int(decay * sampleRate)
	releaseSamples := int(release * sampleRate)
	sustainStart := attackSamples + decaySamples
	releaseStart := len(samples) - releaseSamples

	for i, s := range samples {
		var envelope float64
		switch {
		case i < attackSamples:
			envelope = float64(i) / float64(attackSamples)
		case i < sustainStart:
			decayPos := float64(i-attackSamples) / float64(decaySamples)
			envelope = 1.0 - decayPos*(1.0-sustain)
		case i < releaseStart:
			envelope = sustain
		default:
			releasePos := float64(i-releaseStart) / float64(releaseSamples)
			envelope = sustain * (1.0 - releasePos)
		}
		result[i] = s * envelope
	}
	return result
}

// GetGenreBaseFrequency returns a base frequency appropriate for the genre.
func (e *Engine) GetGenreBaseFrequency() float64 {
	switch e.Genre {
	case "fantasy":
		return 220.0 // A3 - orchestral feel
	case "sci-fi":
		return 110.0 // A2 - deep synth bass
	case "horror":
		return 55.0 // A1 - ominous low drone
	case "cyberpunk":
		return 440.0 // A4 - bright synth
	case "post-apocalyptic":
		return 165.0 // E3 - desolate feel
	default:
		return 220.0
	}
}

// Play starts audio playback.
func (e *Engine) Play() {
	e.playing = true
}

// Stop stops audio playback.
func (e *Engine) Stop() {
	e.playing = false
}

// IsPlaying returns whether audio is currently playing.
func (e *Engine) IsPlaying() bool {
	return e.playing
}
