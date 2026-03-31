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
	StateMenu      // Main menu, title screen
	StatePauseMenu // In-game pause menu
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

	// Menu music layer
	am.layers["menu"] = &Layer{
		Name:   "menu",
		Volume: 0.0,
		Target: 0.0,
		Active: false,
	}
}

// EnterMenu transitions to menu music state.
func (am *AdaptiveMusic) EnterMenu() {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.currentState != StateMenu {
		am.previousState = am.currentState
		am.currentState = StateMenu

		// Fade out all gameplay layers
		am.layers["exploration"].Target = 0.0
		am.layers["combat"].Target = 0.0
		am.layers["tension"].Target = 0.0

		// Fade in menu layer
		am.layers["menu"].Target = MaxVolume
		am.layers["menu"].Active = true
	}
}

// EnterPauseMenu transitions to pause menu music state.
func (am *AdaptiveMusic) EnterPauseMenu() {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.currentState != StatePauseMenu {
		am.previousState = am.currentState
		am.currentState = StatePauseMenu

		// Reduce gameplay music volume but don't mute
		am.layers["exploration"].Target = MenuMusicReduction
		am.layers["combat"].Target = 0.0
		am.layers["tension"].Target = 0.0

		// Fade in menu layer at reduced volume
		am.layers["menu"].Target = MenuMusicVolume
		am.layers["menu"].Active = true
	}
}

// ExitMenu transitions back from menu state.
func (am *AdaptiveMusic) ExitMenu() {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.currentState == StateMenu || am.currentState == StatePauseMenu {
		// Restore previous state
		am.currentState = am.previousState

		// Fade out menu layer
		am.layers["menu"].Target = 0.0

		// Restore gameplay layers based on previous state
		switch am.previousState {
		case StateCombat:
			am.layers["exploration"].Target = CombatVolumeReduction
			am.layers["combat"].Target = MaxVolume
			am.layers["combat"].Active = true
		case StateTense:
			am.layers["exploration"].Target = MaxVolume * TensionMusicReduction
			am.layers["tension"].Target = MaxVolume
			am.layers["tension"].Active = true
		default:
			am.layers["exploration"].Target = MaxVolume
		}
	}
}

// IsInMenu returns true if currently in a menu state.
func (am *AdaptiveMusic) IsInMenu() bool {
	am.mu.Lock()
	defer am.mu.Unlock()
	return am.currentState == StateMenu || am.currentState == StatePauseMenu
}

// GenerateMenuMusic generates menu-specific music samples.
func (am *AdaptiveMusic) GenerateMenuMusic(duration float64) []float64 {
	am.mu.Lock()
	defer am.mu.Unlock()

	numSamples := int(duration * float64(am.sampleRate))
	samples := make([]float64, numSamples)

	// Generate genre-appropriate menu music
	baseFreq := am.getMenuBaseFrequency()
	motif := am.getMenuMotif()

	// Generate a gentle, ambient version of the motif
	samplePos := 0
	for samplePos < numSamples {
		for i, noteMultiplier := range motif.Notes {
			if samplePos >= numSamples {
				break
			}
			freq := baseFreq * noteMultiplier
			noteDuration := motif.Durations[i] * 2.0 // Slower for menu
			noteSamples := int(noteDuration * float64(am.sampleRate))

			for j := 0; j < noteSamples && samplePos < numSamples; j++ {
				t := float64(j) / float64(am.sampleRate)
				// Gentle sine wave with soft attack/release
				envelope := am.getMenuEnvelope(j, noteSamples)
				sample := math.Sin(2*math.Pi*freq*t) * envelope * MenuMusicVolume
				samples[samplePos] = sample
				samplePos++
			}
		}
	}

	return samples
}

// getMenuBaseFrequency returns genre-appropriate base frequency for menu music.
func (am *AdaptiveMusic) getMenuBaseFrequency() float64 {
	switch am.genre {
	case "fantasy":
		return FreqA3 * 0.5 // Lower, warmer
	case "sci-fi":
		return FreqE4
	case "horror":
		return FreqA1 // Deep, ominous
	case "cyberpunk":
		return FreqA4 * 0.75 // Slightly lower synth
	case "post-apocalyptic":
		return FreqE3 * 0.75 // Muted
	default:
		return FreqA3 * 0.5
	}
}

// getMenuMotif returns the menu-appropriate motif.
func (am *AdaptiveMusic) getMenuMotif() *Motif {
	// Use exploration motif as base, which is more ambient
	if motif, exists := am.motifs["exploration"]; exists {
		return motif
	}
	// Fallback motif
	return &Motif{
		Name:      "menu_default",
		BaseFreq:  FreqA3,
		Notes:     []float64{IntervalUnison, IntervalMajorThird, IntervalPerfectFifth, IntervalMajorThird, IntervalUnison},
		Durations: []float64{QuarterNote, QuarterNote, QuarterNote, QuarterNote, HalfNote},
		Genre:     am.genre,
	}
}

// getMenuEnvelope returns a soft envelope for menu music notes.
func (am *AdaptiveMusic) getMenuEnvelope(sampleIdx, totalSamples int) float64 {
	t := float64(sampleIdx) / float64(totalSamples)

	// Soft attack (first 20%)
	if t < 0.2 {
		return t / 0.2
	}
	// Sustain (20%-70%)
	if t < 0.7 {
		return 1.0
	}
	// Release (70%-100%)
	return 1.0 - (t-0.7)/0.3
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

// ============================================================================
// Genre Music Styles
// ============================================================================

// GenreStyle represents a complete music style configuration for a genre.
type GenreStyle struct {
	Genre          string
	BaseScale      []float64 // Scale intervals for melody generation
	Tempo          float64   // BPM
	TimeSig        int       // Beats per measure (3, 4, etc.)
	PreferredMode  string    // "major", "minor", "pentatonic", "chromatic"
	InstrumentMix  []string  // Instrument types to use
	RhythmPattern  []float64 // Rhythm emphasis pattern
	HarmonyDensity float64   // 0-1, how many simultaneous notes
	MelodyRange    float64   // Octave range for melodies
	ReverbAmount   float64   // 0-1, environmental reverb
	DistortionMix  float64   // 0-1, for grittier sounds
}

// GetGenreStyle returns the complete style configuration for a genre.
func GetGenreStyle(genre string) *GenreStyle {
	switch genre {
	case "fantasy":
		return &GenreStyle{
			Genre:          "fantasy",
			BaseScale:      []float64{1.0, 9.0 / 8, 5.0 / 4, 4.0 / 3, 3.0 / 2, 5.0 / 3, 15.0 / 8}, // Major scale
			Tempo:          90,
			TimeSig:        4,
			PreferredMode:  "major",
			InstrumentMix:  []string{"strings", "woodwind", "harp", "choir"},
			RhythmPattern:  []float64{1.0, 0.5, 0.7, 0.5},
			HarmonyDensity: 0.6,
			MelodyRange:    2.0,
			ReverbAmount:   0.4,
			DistortionMix:  0.0,
		}
	case "sci-fi":
		return &GenreStyle{
			Genre:          "sci-fi",
			BaseScale:      []float64{1.0, 16.0 / 15, 6.0 / 5, 45.0 / 32, 3.0 / 2, 8.0 / 5, 15.0 / 8}, // Chromatic-ish
			Tempo:          120,
			TimeSig:        4,
			PreferredMode:  "chromatic",
			InstrumentMix:  []string{"synth_pad", "synth_lead", "electronic_bass", "arpeggiator"},
			RhythmPattern:  []float64{1.0, 0.3, 0.8, 0.3},
			HarmonyDensity: 0.4,
			MelodyRange:    3.0,
			ReverbAmount:   0.6,
			DistortionMix:  0.1,
		}
	case "horror":
		return &GenreStyle{
			Genre:          "horror",
			BaseScale:      []float64{1.0, 16.0 / 15, 6.0 / 5, 4.0 / 3, 45.0 / 32, 8.0 / 5, 15.0 / 8}, // Locrian-ish
			Tempo:          60,
			TimeSig:        3,
			PreferredMode:  "minor",
			InstrumentMix:  []string{"low_strings", "dissonant_pad", "glass_harmonics", "whispers"},
			RhythmPattern:  []float64{1.0, 0.2, 0.4},
			HarmonyDensity: 0.2,
			MelodyRange:    1.5,
			ReverbAmount:   0.8,
			DistortionMix:  0.2,
		}
	case "cyberpunk":
		return &GenreStyle{
			Genre:          "cyberpunk",
			BaseScale:      []float64{1.0, 9.0 / 8, 6.0 / 5, 4.0 / 3, 3.0 / 2, 8.0 / 5, 9.0 / 5}, // Minor scale
			Tempo:          140,
			TimeSig:        4,
			PreferredMode:  "minor",
			InstrumentMix:  []string{"heavy_synth", "glitch", "drum_machine", "distorted_bass"},
			RhythmPattern:  []float64{1.0, 0.8, 0.6, 0.9},
			HarmonyDensity: 0.5,
			MelodyRange:    2.5,
			ReverbAmount:   0.3,
			DistortionMix:  0.4,
		}
	case "post-apocalyptic":
		return &GenreStyle{
			Genre:          "post-apocalyptic",
			BaseScale:      []float64{1.0, 9.0 / 8, 4.0 / 3, 3.0 / 2, 16.0 / 9}, // Pentatonic
			Tempo:          75,
			TimeSig:        4,
			PreferredMode:  "pentatonic",
			InstrumentMix:  []string{"acoustic_guitar", "harmonica", "sparse_drums", "distant_vocals"},
			RhythmPattern:  []float64{1.0, 0.4, 0.6, 0.3},
			HarmonyDensity: 0.3,
			MelodyRange:    1.5,
			ReverbAmount:   0.5,
			DistortionMix:  0.15,
		}
	default:
		return GetGenreStyle("fantasy")
	}
}

// ApplyGenreStyle configures the adaptive music system with a genre's style.
func (am *AdaptiveMusic) ApplyGenreStyle(style *GenreStyle) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.genre = style.Genre
	// Re-initialize motifs with the new genre
	am.initializeMotifs()
}

// GetGenre returns the current genre.
func (am *AdaptiveMusic) GetGenre() string {
	am.mu.Lock()
	defer am.mu.Unlock()
	return am.genre
}

// ============================================================================
// Location-Based Music System
// ============================================================================

// LocationType represents different types of locations for music selection.
type LocationType int

const (
	LocationWilderness LocationType = iota
	LocationTown
	LocationDungeon
	LocationTavern
	LocationTemple
	LocationCastle
	LocationShop
	LocationCombatArena
	LocationBossRoom
	LocationSafeZone
)

// LocationMusicConfig holds music configuration for a location type.
type LocationMusicConfig struct {
	Location        LocationType
	BaseIntensity   float64   // 0-1, base layer mixing
	TempoMultiplier float64   // Adjust tempo
	PitchShift      float64   // Semitones to shift
	LayerWeights    []float64 // Weight for each layer
	TransitionTime  float64   // Seconds to transition
}

// LocationMusicManager handles location-based music transitions.
type LocationMusicManager struct {
	mu              sync.RWMutex
	adaptiveMusic   *AdaptiveMusic
	currentLocation LocationType
	targetLocation  LocationType
	transitionTimer float64
	configs         map[LocationType]*LocationMusicConfig
}

// NewLocationMusicManager creates a location-based music manager.
func NewLocationMusicManager(adaptiveMusic *AdaptiveMusic) *LocationMusicManager {
	lmm := &LocationMusicManager{
		adaptiveMusic:   adaptiveMusic,
		currentLocation: LocationWilderness,
		targetLocation:  LocationWilderness,
		transitionTimer: 0,
		configs:         make(map[LocationType]*LocationMusicConfig),
	}
	lmm.initializeConfigs()
	return lmm
}

// initializeConfigs sets up music configs for each location type.
func (lmm *LocationMusicManager) initializeConfigs() {
	lmm.configs[LocationWilderness] = &LocationMusicConfig{
		Location:        LocationWilderness,
		BaseIntensity:   0.5,
		TempoMultiplier: 1.0,
		PitchShift:      0,
		LayerWeights:    []float64{1.0, 0.0, 0.2}, // exploration, combat, tension
		TransitionTime:  3.0,
	}
	lmm.configs[LocationTown] = &LocationMusicConfig{
		Location:        LocationTown,
		BaseIntensity:   0.6,
		TempoMultiplier: 1.1,
		PitchShift:      2, // Brighter
		LayerWeights:    []float64{0.8, 0.0, 0.0},
		TransitionTime:  2.0,
	}
	lmm.configs[LocationDungeon] = &LocationMusicConfig{
		Location:        LocationDungeon,
		BaseIntensity:   0.7,
		TempoMultiplier: 0.9,
		PitchShift:      -3, // Darker
		LayerWeights:    []float64{0.5, 0.0, 0.5},
		TransitionTime:  4.0,
	}
	lmm.configs[LocationTavern] = &LocationMusicConfig{
		Location:        LocationTavern,
		BaseIntensity:   0.8,
		TempoMultiplier: 1.2,
		PitchShift:      0,
		LayerWeights:    []float64{1.0, 0.0, 0.0},
		TransitionTime:  1.5,
	}
	lmm.configs[LocationTemple] = &LocationMusicConfig{
		Location:        LocationTemple,
		BaseIntensity:   0.4,
		TempoMultiplier: 0.7,
		PitchShift:      0,
		LayerWeights:    []float64{0.6, 0.0, 0.3},
		TransitionTime:  5.0,
	}
	lmm.configs[LocationCastle] = &LocationMusicConfig{
		Location:        LocationCastle,
		BaseIntensity:   0.6,
		TempoMultiplier: 0.95,
		PitchShift:      1,
		LayerWeights:    []float64{0.7, 0.0, 0.3},
		TransitionTime:  3.0,
	}
	lmm.configs[LocationShop] = &LocationMusicConfig{
		Location:        LocationShop,
		BaseIntensity:   0.5,
		TempoMultiplier: 1.0,
		PitchShift:      1,
		LayerWeights:    []float64{0.9, 0.0, 0.0},
		TransitionTime:  1.0,
	}
	lmm.configs[LocationCombatArena] = &LocationMusicConfig{
		Location:        LocationCombatArena,
		BaseIntensity:   0.9,
		TempoMultiplier: 1.3,
		PitchShift:      -1,
		LayerWeights:    []float64{0.2, 0.8, 0.5},
		TransitionTime:  1.0,
	}
	lmm.configs[LocationBossRoom] = &LocationMusicConfig{
		Location:        LocationBossRoom,
		BaseIntensity:   1.0,
		TempoMultiplier: 1.2,
		PitchShift:      -2,
		LayerWeights:    []float64{0.1, 1.0, 0.6},
		TransitionTime:  0.5,
	}
	lmm.configs[LocationSafeZone] = &LocationMusicConfig{
		Location:        LocationSafeZone,
		BaseIntensity:   0.3,
		TempoMultiplier: 0.8,
		PitchShift:      2,
		LayerWeights:    []float64{1.0, 0.0, 0.0},
		TransitionTime:  2.0,
	}
}

// SetLocation triggers a music transition to a new location.
func (lmm *LocationMusicManager) SetLocation(location LocationType) {
	lmm.mu.Lock()
	defer lmm.mu.Unlock()

	if location == lmm.currentLocation {
		return
	}

	lmm.targetLocation = location
	config := lmm.configs[location]
	if config != nil {
		lmm.transitionTimer = config.TransitionTime
	} else {
		lmm.transitionTimer = 2.0 // Default
	}
}

// Update processes location music transitions.
func (lmm *LocationMusicManager) Update(dt float64) {
	lmm.mu.Lock()
	defer lmm.mu.Unlock()

	if lmm.currentLocation == lmm.targetLocation {
		return
	}

	lmm.transitionTimer -= dt
	if lmm.transitionTimer <= 0 {
		lmm.currentLocation = lmm.targetLocation
		lmm.transitionTimer = 0
		lmm.applyLocationConfig()
	}
}

// applyLocationConfig applies the current location's music configuration.
func (lmm *LocationMusicManager) applyLocationConfig() {
	config := lmm.configs[lmm.currentLocation]
	if config == nil || lmm.adaptiveMusic == nil {
		return
	}

	lmm.adaptiveMusic.mu.Lock()
	defer lmm.adaptiveMusic.mu.Unlock()

	if len(config.LayerWeights) >= 3 {
		lmm.applyLayerWeights(config.LayerWeights)
	}
}

// applyLayerWeights sets layer targets from config weights.
func (lmm *LocationMusicManager) applyLayerWeights(weights []float64) {
	if layer, ok := lmm.adaptiveMusic.layers["exploration"]; ok {
		layer.Target = weights[0]
	}
	if layer, ok := lmm.adaptiveMusic.layers["combat"]; ok {
		if lmm.adaptiveMusic.currentState != StateCombat {
			layer.Target = weights[1]
		}
	}
	if layer, ok := lmm.adaptiveMusic.layers["tension"]; ok {
		layer.Target = weights[2]
		layer.Active = weights[2] > 0
	}
}

// GetCurrentLocation returns the current location type.
func (lmm *LocationMusicManager) GetCurrentLocation() LocationType {
	lmm.mu.RLock()
	defer lmm.mu.RUnlock()
	return lmm.currentLocation
}

// GetTransitionProgress returns progress of current transition (0-1).
func (lmm *LocationMusicManager) GetTransitionProgress() float64 {
	lmm.mu.RLock()
	defer lmm.mu.RUnlock()

	if lmm.currentLocation == lmm.targetLocation {
		return 1.0
	}

	config := lmm.configs[lmm.targetLocation]
	if config == nil || config.TransitionTime <= 0 {
		return 1.0
	}

	return 1.0 - (lmm.transitionTimer / config.TransitionTime)
}

// GetLocationConfig returns the config for a location type.
func (lmm *LocationMusicManager) GetLocationConfig(location LocationType) *LocationMusicConfig {
	lmm.mu.RLock()
	defer lmm.mu.RUnlock()
	return lmm.configs[location]
}

// IsInTransition returns whether a music transition is in progress.
func (lmm *LocationMusicManager) IsInTransition() bool {
	lmm.mu.RLock()
	defer lmm.mu.RUnlock()
	return lmm.currentLocation != lmm.targetLocation
}

// BossMusicPhase represents a phase of boss fight music.
type BossMusicPhase int

const (
	BossPhaseIntro BossMusicPhase = iota
	BossPhaseMain
	BossPhaseIntense
	BossPhaseFinal
	BossPhaseVictory
)

// BossMusicConfig defines configuration for boss fight music.
type BossMusicConfig struct {
	BossName       string
	BaseTempo      float64          // BPM
	BaseIntensity  float64          // 0.0 to 1.0
	Phases         []BossMusicPhase // Sequence of phases
	PhaseThreshold []float64        // Boss HP% thresholds for phase changes
	Genre          string
	LoopPoints     map[BossMusicPhase][2]float64 // Start/end loop points per phase
	Motif          *Motif                        // Boss's signature motif
}

// BossMusicManager handles boss fight music with phase transitions.
type BossMusicManager struct {
	mu             sync.RWMutex
	configs        map[string]*BossMusicConfig
	currentBoss    string
	currentPhase   BossMusicPhase
	bossHealth     float64 // 0.0 to 1.0
	isActive       bool
	genre          string
	transitionTime float64
}

// NewBossMusicManager creates a new boss music manager.
func NewBossMusicManager(genre string) *BossMusicManager {
	bmm := &BossMusicManager{
		configs:        make(map[string]*BossMusicConfig),
		genre:          genre,
		transitionTime: 0.5, // Fast transitions for boss music
	}
	bmm.initDefaultConfigs()
	return bmm
}

// initDefaultConfigs sets up default boss music configurations.
func (bmm *BossMusicManager) initDefaultConfigs() {
	// Generic boss config
	bmm.configs["generic"] = &BossMusicConfig{
		BossName:       "Generic Boss",
		BaseTempo:      140.0,
		BaseIntensity:  0.8,
		Phases:         []BossMusicPhase{BossPhaseIntro, BossPhaseMain, BossPhaseIntense, BossPhaseFinal},
		PhaseThreshold: []float64{1.0, 0.75, 0.5, 0.25},
		Genre:          bmm.genre,
		LoopPoints:     make(map[BossMusicPhase][2]float64),
	}

	// Final boss config
	bmm.configs["final"] = &BossMusicConfig{
		BossName:       "Final Boss",
		BaseTempo:      160.0,
		BaseIntensity:  0.9,
		Phases:         []BossMusicPhase{BossPhaseIntro, BossPhaseMain, BossPhaseIntense, BossPhaseFinal},
		PhaseThreshold: []float64{1.0, 0.66, 0.33, 0.1},
		Genre:          bmm.genre,
		LoopPoints:     make(map[BossMusicPhase][2]float64),
	}
}

// StartBossFight initiates boss fight music.
func (bmm *BossMusicManager) StartBossFight(bossType string) {
	bmm.mu.Lock()
	defer bmm.mu.Unlock()
	if _, ok := bmm.configs[bossType]; !ok {
		bossType = "generic"
	}
	bmm.currentBoss = bossType
	bmm.currentPhase = BossPhaseIntro
	bmm.bossHealth = 1.0
	bmm.isActive = true
}

// UpdateBossHealth updates the boss health and checks for phase transitions.
func (bmm *BossMusicManager) UpdateBossHealth(health float64) {
	bmm.mu.Lock()
	defer bmm.mu.Unlock()
	bmm.bossHealth = clampFloat(health, 0, 1)
	if !bmm.isActive {
		return
	}
	bmm.checkPhaseTransitions()
}

// checkPhaseTransitions evaluates and applies boss phase transitions.
func (bmm *BossMusicManager) checkPhaseTransitions() {
	config := bmm.configs[bmm.currentBoss]
	if config == nil {
		return
	}
	for i, threshold := range config.PhaseThreshold {
		if bmm.shouldTransitionToPhase(i, threshold, config) {
			bmm.currentPhase = config.Phases[i]
		}
	}
}

// shouldTransitionToPhase checks if a phase transition should occur.
func (bmm *BossMusicManager) shouldTransitionToPhase(phaseIndex int, threshold float64, config *BossMusicConfig) bool {
	return bmm.bossHealth <= threshold &&
		phaseIndex < len(config.Phases) &&
		config.Phases[phaseIndex] > bmm.currentPhase
}

// EndBossFight ends boss fight music (victory or defeat).
func (bmm *BossMusicManager) EndBossFight(victory bool) {
	bmm.mu.Lock()
	defer bmm.mu.Unlock()
	if victory {
		bmm.currentPhase = BossPhaseVictory
	}
	bmm.isActive = false
}

// GetCurrentTempo returns the tempo for the current boss phase.
func (bmm *BossMusicManager) GetCurrentTempo() float64 {
	bmm.mu.RLock()
	defer bmm.mu.RUnlock()
	config := bmm.configs[bmm.currentBoss]
	if config == nil {
		return 120.0
	}
	// Tempo increases with phase intensity
	phaseMultiplier := 1.0 + float64(bmm.currentPhase)*0.1
	return config.BaseTempo * phaseMultiplier
}

// GetCurrentIntensity returns the intensity for the current boss phase.
func (bmm *BossMusicManager) GetCurrentIntensity() float64 {
	bmm.mu.RLock()
	defer bmm.mu.RUnlock()
	config := bmm.configs[bmm.currentBoss]
	if config == nil {
		return 0.5
	}
	// Intensity increases with phase
	phaseBonus := float64(bmm.currentPhase) * 0.1
	return clampFloat(config.BaseIntensity+phaseBonus, 0, 1)
}

// IsActive returns whether boss music is currently playing.
func (bmm *BossMusicManager) IsActive() bool {
	bmm.mu.RLock()
	defer bmm.mu.RUnlock()
	return bmm.isActive
}

// GetCurrentPhase returns the current boss music phase.
func (bmm *BossMusicManager) GetCurrentPhase() BossMusicPhase {
	bmm.mu.RLock()
	defer bmm.mu.RUnlock()
	return bmm.currentPhase
}

// DynamicLayerConfig defines configuration for a dynamic music layer.
type DynamicLayerConfig struct {
	Name         string
	BaseVolume   float64
	TriggerState State   // Game state that activates this layer
	FadeInTime   float64 // Seconds to fade in
	FadeOutTime  float64 // Seconds to fade out
	LoopEnabled  bool
	Priority     int      // Higher priority overrides lower
	Tags         []string // Tags for filtering/grouping
}

// DynamicLayerManager manages multiple music layers with state-based mixing.
type DynamicLayerManager struct {
	mu            sync.RWMutex
	layers        map[string]*DynamicLayerConfig
	activeStates  map[State]bool
	layerVolumes  map[string]float64
	targetVolumes map[string]float64
	fadeProgress  map[string]float64
	masterVolume  float64
}

// NewDynamicLayerManager creates a new dynamic layer manager.
func NewDynamicLayerManager() *DynamicLayerManager {
	dlm := &DynamicLayerManager{
		layers:        make(map[string]*DynamicLayerConfig),
		activeStates:  make(map[State]bool),
		layerVolumes:  make(map[string]float64),
		targetVolumes: make(map[string]float64),
		fadeProgress:  make(map[string]float64),
		masterVolume:  1.0,
	}
	dlm.initDefaultLayers()
	return dlm
}

// initDefaultLayers sets up default music layers.
func (dlm *DynamicLayerManager) initDefaultLayers() {
	// Base exploration layer
	dlm.layers["exploration_base"] = &DynamicLayerConfig{
		Name:         "exploration_base",
		BaseVolume:   0.6,
		TriggerState: StateExploration,
		FadeInTime:   2.0,
		FadeOutTime:  1.5,
		LoopEnabled:  true,
		Priority:     1,
		Tags:         []string{"ambient", "exploration"},
	}

	// Combat percussion layer
	dlm.layers["combat_percussion"] = &DynamicLayerConfig{
		Name:         "combat_percussion",
		BaseVolume:   0.8,
		TriggerState: StateCombat,
		FadeInTime:   0.5,
		FadeOutTime:  2.0,
		LoopEnabled:  true,
		Priority:     3,
		Tags:         []string{"combat", "percussion"},
	}

	// Combat strings layer
	dlm.layers["combat_strings"] = &DynamicLayerConfig{
		Name:         "combat_strings",
		BaseVolume:   0.7,
		TriggerState: StateCombat,
		FadeInTime:   0.3,
		FadeOutTime:  1.5,
		LoopEnabled:  true,
		Priority:     2,
		Tags:         []string{"combat", "melody"},
	}

	// Tense layer
	dlm.layers["tense_drone"] = &DynamicLayerConfig{
		Name:         "tense_drone",
		BaseVolume:   0.5,
		TriggerState: StateTense,
		FadeInTime:   3.0,
		FadeOutTime:  2.0,
		LoopEnabled:  true,
		Priority:     2,
		Tags:         []string{"tension", "ambient"},
	}

	// Victory fanfare
	dlm.layers["victory_fanfare"] = &DynamicLayerConfig{
		Name:         "victory_fanfare",
		BaseVolume:   0.9,
		TriggerState: StateVictory,
		FadeInTime:   0.1,
		FadeOutTime:  3.0,
		LoopEnabled:  false,
		Priority:     5,
		Tags:         []string{"victory", "stinger"},
	}
}

// SetState activates or deactivates a game state.
func (dlm *DynamicLayerManager) SetState(state State, active bool) {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()
	dlm.activeStates[state] = active
	dlm.updateTargetVolumes()
}

// updateTargetVolumes recalculates target volumes based on active states.
func (dlm *DynamicLayerManager) updateTargetVolumes() {
	for name, config := range dlm.layers {
		if dlm.activeStates[config.TriggerState] {
			dlm.targetVolumes[name] = config.BaseVolume
		} else {
			dlm.targetVolumes[name] = 0
		}
	}
}

// Update processes layer fading each frame.
func (dlm *DynamicLayerManager) Update(dt float64) {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()
	for name, config := range dlm.layers {
		current := dlm.layerVolumes[name]
		target := dlm.targetVolumes[name]
		if current < target {
			// Fade in
			fadeRate := 1.0 / config.FadeInTime
			dlm.layerVolumes[name] = clampFloat(current+fadeRate*dt, 0, target)
		} else if current > target {
			// Fade out
			fadeRate := 1.0 / config.FadeOutTime
			dlm.layerVolumes[name] = clampFloat(current-fadeRate*dt, target, 1)
		}
	}
}

// GetLayerVolume returns the current volume for a layer.
func (dlm *DynamicLayerManager) GetLayerVolume(name string) float64 {
	dlm.mu.RLock()
	defer dlm.mu.RUnlock()
	return dlm.layerVolumes[name] * dlm.masterVolume
}

// GetActiveLayersByTag returns all active layers with a given tag.
func (dlm *DynamicLayerManager) GetActiveLayersByTag(tag string) []string {
	dlm.mu.RLock()
	defer dlm.mu.RUnlock()
	var result []string
	for name, config := range dlm.layers {
		if dlm.layerVolumes[name] > 0 {
			for _, t := range config.Tags {
				if t == tag {
					result = append(result, name)
					break
				}
			}
		}
	}
	return result
}

// SetMasterVolume sets the master volume for all layers.
func (dlm *DynamicLayerManager) SetMasterVolume(volume float64) {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()
	dlm.masterVolume = clampFloat(volume, 0, 1)
}

// AddLayer adds a custom layer configuration.
func (dlm *DynamicLayerManager) AddLayer(config *DynamicLayerConfig) {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()
	dlm.layers[config.Name] = config
	dlm.layerVolumes[config.Name] = 0
	dlm.targetVolumes[config.Name] = 0
}

// clampFloat clamps a value between min and max.
func clampFloat(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
