// Package audio provides procedural audio synthesis.
package audio

import (
	"math"
)

// SFXType represents a category of sound effect.
type SFXType int

const (
	SFXFootstep SFXType = iota
	SFXHit
	SFXExplosion
	SFXAmbient
	SFXUI
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

// GetGenrePitchModifier returns the pitch modifier for the genre.
// Per ROADMAP Phase 4: sci-fi +30%, horror -30%, cyberpunk +40%.
func (e *Engine) GetGenrePitchModifier() float64 {
	switch e.Genre {
	case "fantasy":
		return 1.0 // baseline
	case "sci-fi":
		return 1.30 // +30% pitch
	case "horror":
		return 0.70 // -30% pitch
	case "cyberpunk":
		return 1.40 // +40% pitch
	case "post-apocalyptic":
		return 0.85 // -15% for desolate feel
	default:
		return 1.0
	}
}

// GenerateSFX generates a sound effect with genre-specific modifications.
func (e *Engine) GenerateSFX(sfxType SFXType, baseFreq, duration float64) []float64 {
	// Apply genre pitch modifier
	freq := baseFreq * e.GetGenrePitchModifier()
	samples := e.GenerateSineWave(freq, duration)

	// Apply genre-specific effects
	switch e.Genre {
	case "horror":
		samples = e.applyVibrato(samples, 5.0, 0.1) // 5 Hz vibrato
	case "cyberpunk":
		samples = e.applyHardClipping(samples, 0.7) // hard clip at 70%
	case "post-apocalyptic":
		samples = e.applyDistortion(samples, 0.3) // subtle distortion
	case "sci-fi":
		samples = e.applyMetallicSheen(samples)
	}

	// Apply ADSR envelope based on SFX type
	switch sfxType {
	case SFXFootstep:
		samples = e.ApplyADSR(samples, 0.01, 0.05, 0.3, 0.1)
	case SFXHit:
		samples = e.ApplyADSR(samples, 0.005, 0.1, 0.4, 0.15)
	case SFXExplosion:
		samples = e.ApplyADSR(samples, 0.02, 0.2, 0.5, 0.3)
	case SFXAmbient:
		samples = e.ApplyADSR(samples, 0.3, 0.2, 0.8, 0.5)
	case SFXUI:
		samples = e.ApplyADSR(samples, 0.005, 0.02, 0.6, 0.05)
	}

	return samples
}

// applyVibrato applies pitch vibrato to samples.
func (e *Engine) applyVibrato(samples []float64, rate, depth float64) []float64 {
	result := make([]float64, len(samples))
	for i, s := range samples {
		t := float64(i) / float64(e.SampleRate)
		vibrato := 1.0 + depth*math.Sin(2*math.Pi*rate*t)
		result[i] = s * vibrato
	}
	return result
}

// applyHardClipping applies hard clipping distortion.
func (e *Engine) applyHardClipping(samples []float64, threshold float64) []float64 {
	result := make([]float64, len(samples))
	for i, s := range samples {
		if s > threshold {
			result[i] = threshold
		} else if s < -threshold {
			result[i] = -threshold
		} else {
			result[i] = s
		}
	}
	return result
}

// applyDistortion applies soft distortion using tanh.
func (e *Engine) applyDistortion(samples []float64, amount float64) []float64 {
	result := make([]float64, len(samples))
	drive := 1.0 + amount*5.0
	for i, s := range samples {
		result[i] = math.Tanh(s * drive)
	}
	return result
}

// applyMetallicSheen adds harmonics for a metallic sound.
func (e *Engine) applyMetallicSheen(samples []float64) []float64 {
	result := make([]float64, len(samples))
	// Mix in higher harmonics
	for i := range samples {
		harmonic := math.Sin(float64(i)*0.02) * 0.15
		result[i] = samples[i]*0.85 + harmonic
	}
	return result
}
