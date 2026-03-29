// Package music provides an adaptive music system.
// Per ROADMAP Phase 4 item 16:
// - Motifs per faction/region
// - Combat intensity layer mixing
// - Exploration vs combat transitions
// AC: Music transitions within 2s of entering combat; reverts within 5s of last enemy death
package music

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// State represents the current game state for music adaptation.
type State int

const (
	StateExploration State = iota
	StateCombat
	StateTense
	StateVictory
	StateDefeat
)

// Layer represents a music layer that can be mixed in/out.
type Layer struct {
	Name     string
	Samples  []float64
	Volume   float64 // Current volume (0.0 to 1.0)
	Target   float64 // Target volume for crossfade
	Active   bool
	Position int
}

// Motif represents a musical phrase tied to a faction or region.
type Motif struct {
	Name      string
	BaseFreq  float64
	Notes     []float64 // Frequency multipliers
	Durations []float64 // Duration in seconds for each note
	Genre     string
}

// AdaptiveMusic manages dynamic music based on game state.
type AdaptiveMusic struct {
	mu            sync.Mutex
	genre         string
	sampleRate    int
	currentState  State
	previousState State
	layers        map[string]*Layer
	motifs        map[string]*Motif

	// Timing for transitions (per ROADMAP AC)
	combatEntryTime time.Time
	lastEnemyDeath  time.Time

	// Crossfade settings
	crossfadeDuration float64 // seconds

	rng *rand.Rand
}

// NewAdaptiveMusic creates a new adaptive music system.
func NewAdaptiveMusic(genre string, seed int64) *AdaptiveMusic {
	am := &AdaptiveMusic{
		genre:             genre,
		sampleRate:        DefaultSampleRate,
		currentState:      StateExploration,
		previousState:     StateExploration,
		layers:            make(map[string]*Layer),
		motifs:            make(map[string]*Motif),
		crossfadeDuration: DefaultCrossfadeDuration, // 2 second transition per AC
		rng:               rand.New(rand.NewSource(seed)),
	}

	// Initialize genre-specific motifs
	am.initializeMotifs()

	// Initialize layers
	am.initializeLayers()

	return am
}

// initializeMotifs creates genre-appropriate musical motifs.
func (am *AdaptiveMusic) initializeMotifs() {
	switch am.genre {
	case "fantasy":
		am.motifs["exploration"] = &Motif{
			Name:      "exploration",
			BaseFreq:  FreqA3,
			Notes:     []float64{IntervalUnison, IntervalMajorThird, IntervalPerfectFifth, IntervalMajorThird, IntervalUnison, IntervalDownMinor3rd, IntervalUnison},
			Durations: []float64{EighthNote, EighthNote, SixteenthNote, SixteenthNote, EighthNote, EighthNote, QuarterNote},
			Genre:     "fantasy",
		}
		am.motifs["combat"] = &Motif{
			Name:      "combat",
			BaseFreq:  FreqA2,
			Notes:     []float64{IntervalUnison, IntervalUnison, IntervalPerfectFifth, IntervalUnison, IntervalPerfectFifth, IntervalOctave, IntervalPerfectFifth},
			Durations: []float64{SixteenthNote, SixteenthNote, SixteenthNote, SixteenthNote, SixteenthNote, SixteenthNote, EighthNote},
			Genre:     "fantasy",
		}
	case "sci-fi":
		am.motifs["exploration"] = &Motif{
			Name:      "exploration",
			BaseFreq:  FreqE4,
			Notes:     []float64{IntervalUnison, IntervalMinorThird, IntervalTritone, IntervalMinorThird, IntervalUnison},
			Durations: []float64{QuarterNote, EighthNote, EighthNote, EighthNote, DottedQuarter},
			Genre:     "sci-fi",
		}
		am.motifs["combat"] = &Motif{
			Name:      "combat",
			BaseFreq:  FreqE3,
			Notes:     []float64{IntervalUnison, IntervalUnison, IntervalTritone, IntervalUnison, IntervalMajorSixth, IntervalTritone},
			Durations: []float64{ThirtySecondNote, ThirtySecondNote, SixteenthNote, ThirtySecondNote, SixteenthNote, ThirtySecondNote},
			Genre:     "sci-fi",
		}
	case "horror":
		am.motifs["exploration"] = &Motif{
			Name:      "exploration",
			BaseFreq:  FreqA1,
			Notes:     []float64{IntervalUnison, IntervalMinorSecond, IntervalUnison, IntervalDownMinor2nd, IntervalUnison},
			Durations: []float64{HalfNote, QuarterNote, QuarterNote, QuarterNote, HalfNote},
			Genre:     "horror",
		}
		am.motifs["combat"] = &Motif{
			Name:      "combat",
			BaseFreq:  FreqE2,
			Notes:     []float64{IntervalUnison, IntervalMinorSecond, IntervalMajorSecond, IntervalUnison, IntervalDownMinor2nd, IntervalUnison},
			Durations: []float64{EighthNote, EighthNote, EighthNote, SixteenthNote, SixteenthNote, EighthNote},
			Genre:     "horror",
		}
	case "cyberpunk":
		am.motifs["exploration"] = &Motif{
			Name:      "exploration",
			BaseFreq:  FreqA4,
			Notes:     []float64{IntervalUnison, IntervalDown5th, IntervalUnison, IntervalPerfectFifth, IntervalUnison},
			Durations: []float64{EighthNote, SixteenthNote, SixteenthNote, EighthNote, EighthNote},
			Genre:     "cyberpunk",
		}
		am.motifs["combat"] = &Motif{
			Name:      "combat",
			BaseFreq:  FreqA3,
			Notes:     []float64{IntervalUnison, IntervalUnison, IntervalPerfectFifth, IntervalPerfectFifth, IntervalOctave, IntervalUnison},
			Durations: []float64{ThirtySecondNote, ThirtySecondNote, ThirtySecondNote, ThirtySecondNote, SixteenthNote, SixteenthNote},
			Genre:     "cyberpunk",
		}
	case "post-apocalyptic":
		am.motifs["exploration"] = &Motif{
			Name:      "exploration",
			BaseFreq:  FreqE3,
			Notes:     []float64{IntervalUnison, IntervalDownMajor2nd, IntervalUnison, IntervalMajorSecond, IntervalUnison},
			Durations: []float64{QuarterNote, EighthNote, EighthNote, EighthNote, DottedQuarter},
			Genre:     "post-apocalyptic",
		}
		am.motifs["combat"] = &Motif{
			Name:      "combat",
			BaseFreq:  FreqA2,
			Notes:     []float64{IntervalUnison, IntervalUnison, IntervalPerfectFifth, IntervalUnison, IntervalDown5th, IntervalUnison},
			Durations: []float64{SixteenthNote, SixteenthNote, SixteenthNote, SixteenthNote, SixteenthNote, SixteenthNote},
			Genre:     "post-apocalyptic",
		}
	default:
		// Default to fantasy
		am.motifs["exploration"] = &Motif{
			Name:      "exploration",
			BaseFreq:  FreqA3,
			Notes:     []float64{IntervalUnison, IntervalMajorThird, IntervalPerfectFifth, IntervalMajorThird, IntervalUnison},
			Durations: []float64{EighthNote, EighthNote, EighthNote, EighthNote, QuarterNote},
			Genre:     "default",
		}
		am.motifs["combat"] = &Motif{
			Name:      "combat",
			BaseFreq:  FreqA2,
			Notes:     []float64{IntervalUnison, IntervalUnison, IntervalPerfectFifth, IntervalUnison, IntervalPerfectFifth, IntervalUnison},
			Durations: []float64{SixteenthNote, SixteenthNote, SixteenthNote, SixteenthNote, SixteenthNote, SixteenthNote},
			Genre:     "default",
		}
	}
}

// initializeLayers sets up the music layers.
func (am *AdaptiveMusic) initializeLayers() {
	// Base exploration layer
	am.layers["exploration"] = &Layer{
		Name:   "exploration",
		Volume: MaxVolume,
		Target: MaxVolume,
		Active: true,
	}

	// Combat intensity layer
	am.layers["combat"] = &Layer{
		Name:   "combat",
		Volume: 0.0,
		Target: 0.0,
		Active: false,
	}

	// Tension layer
	am.layers["tension"] = &Layer{
		Name:   "tension",
		Volume: 0.0,
		Target: 0.0,
		Active: false,
	}
}

// EnterCombat signals that combat has started.
func (am *AdaptiveMusic) EnterCombat() {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.currentState != StateCombat {
		am.previousState = am.currentState
		am.currentState = StateCombat
		am.combatEntryTime = time.Now()

		// Set layer targets for crossfade
		am.layers["exploration"].Target = CombatVolumeReduction // Reduce exploration
		am.layers["combat"].Target = MaxVolume
		am.layers["combat"].Active = true
	}
}

// EnemyDied records an enemy death for transition timing.
func (am *AdaptiveMusic) EnemyDied() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.lastEnemyDeath = time.Now()
}

// ExitCombat signals that combat has ended.
func (am *AdaptiveMusic) ExitCombat() {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.currentState == StateCombat {
		am.currentState = StateExploration

		// Set layer targets for crossfade back to exploration
		am.layers["exploration"].Target = MaxVolume
		am.layers["combat"].Target = 0.0
	}
}

// Update advances the music system by one tick.
// This handles crossfading between layers.
func (am *AdaptiveMusic) Update(dt float64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.checkAutomaticCombatExit()
	am.crossfadeLayers(dt)
}

// checkAutomaticCombatExit handles automatic combat exit (5 seconds after last enemy death per AC).
func (am *AdaptiveMusic) checkAutomaticCombatExit() {
	if am.currentState != StateCombat || am.lastEnemyDeath.IsZero() {
		return
	}
	if time.Since(am.lastEnemyDeath).Seconds() >= CombatExitDelay {
		am.currentState = StateExploration
		am.layers["exploration"].Target = MaxVolume
		am.layers["combat"].Target = 0.0
	}
}

// crossfadeLayers gradually transitions layer volumes toward their targets.
func (am *AdaptiveMusic) crossfadeLayers(dt float64) {
	fadeRate := dt / am.crossfadeDuration
	for _, layer := range am.layers {
		am.updateLayerVolume(layer, fadeRate)
	}
}

// updateLayerVolume adjusts a single layer's volume toward its target.
func (am *AdaptiveMusic) updateLayerVolume(layer *Layer, fadeRate float64) {
	if layer.Volume < layer.Target {
		layer.Volume = math.Min(layer.Volume+fadeRate, layer.Target)
	} else if layer.Volume > layer.Target {
		layer.Volume = math.Max(layer.Volume-fadeRate, layer.Target)
	}

	// Deactivate layers that have faded out
	if layer.Volume <= MinVolume && layer.Target <= 0 {
		layer.Active = false
	}
}

// getMotifForLayer returns the appropriate motif for a layer name.
func (am *AdaptiveMusic) getMotifForLayer(name string) *Motif {
	switch name {
	case "exploration", "tension":
		return am.motifs["exploration"]
	case "combat":
		return am.motifs["combat"]
	default:
		return nil
	}
}

// mixLayerIntoSamples adds a single layer's contribution to the output samples.
func (am *AdaptiveMusic) mixLayerIntoSamples(samples []float64, name string, layer *Layer) {
	if !layer.Active || layer.Volume < MinVolume {
		return
	}

	motif := am.getMotifForLayer(name)
	if motif == nil {
		return
	}

	layerSamples := am.generateMotifSamples(motif, len(samples))
	for i := range samples {
		samples[i] += layerSamples[i] * layer.Volume
	}
}

// normalizeSamples prevents clipping by scaling samples to max amplitude of 1.0.
func normalizeSamples(samples []float64) {
	maxAmp := 0.0
	for _, s := range samples {
		if abs := math.Abs(s); abs > maxAmp {
			maxAmp = abs
		}
	}
	if maxAmp > 1.0 {
		for i := range samples {
			samples[i] /= maxAmp
		}
	}
}

// GenerateSamples produces mixed audio samples for the current state.
func (am *AdaptiveMusic) GenerateSamples(duration float64) []float64 {
	am.mu.Lock()
	defer am.mu.Unlock()

	numSamples := int(duration * float64(am.sampleRate))
	samples := make([]float64, numSamples)

	// Mix active layers
	for name, layer := range am.layers {
		am.mixLayerIntoSamples(samples, name, layer)
	}

	normalizeSamples(samples)

	return samples
}

// generateMotifSamples creates audio samples from a motif.
func (am *AdaptiveMusic) generateMotifSamples(motif *Motif, numSamples int) []float64 {
	samples := make([]float64, numSamples)
	position := 0
	noteIndex := 0

	for position < numSamples && noteIndex < len(motif.Notes) {
		freq := motif.BaseFreq * motif.Notes[noteIndex]
		dur := motif.Durations[noteIndex]
		noteSamples := int(dur * float64(am.sampleRate))

		// Generate note with envelope
		for i := 0; i < noteSamples && position+i < numSamples; i++ {
			t := float64(i) / float64(am.sampleRate)
			sample := math.Sin(TwoPi * freq * t)

			// Simple ADSR envelope
			env := 1.0
			if t < DefaultAttackTime {
				env = t / DefaultAttackTime
			} else if t > dur-DefaultReleaseTime {
				env = (dur - t) / DefaultReleaseTime
			}

			samples[position+i] = sample * env
		}

		position += noteSamples
		noteIndex = (noteIndex + 1) % len(motif.Notes)
	}

	return samples
}

// GetCurrentState returns the current music state.
func (am *AdaptiveMusic) GetCurrentState() State {
	am.mu.Lock()
	defer am.mu.Unlock()
	return am.currentState
}

// GetLayerVolume returns the current volume of a layer.
func (am *AdaptiveMusic) GetLayerVolume(name string) float64 {
	am.mu.Lock()
	defer am.mu.Unlock()
	if layer, ok := am.layers[name]; ok {
		return layer.Volume
	}
	return 0.0
}

// TimeSinceCombatEntry returns time since combat was entered.
func (am *AdaptiveMusic) TimeSinceCombatEntry() time.Duration {
	am.mu.Lock()
	defer am.mu.Unlock()
	if am.combatEntryTime.IsZero() {
		return 0
	}
	return time.Since(am.combatEntryTime)
}

// TimeSinceLastEnemyDeath returns time since the last enemy death.
func (am *AdaptiveMusic) TimeSinceLastEnemyDeath() time.Duration {
	am.mu.Lock()
	defer am.mu.Unlock()
	if am.lastEnemyDeath.IsZero() {
		return 0
	}
	return time.Since(am.lastEnemyDeath)
}
