// Package ambient provides location-based ambient soundscapes.
// Per ROADMAP Phase 4 item 18:
// - Cave drip, city crowd murmur, wind on plains — synthesized, no files
// AC: Ambient sound type changes within 1s of entering new region type
package ambient

import (
	"math"
	"math/rand"
	"sync"
)

// RegionType represents different ambient environments.
type RegionType int

const (
	RegionPlains RegionType = iota
	RegionForest
	RegionCave
	RegionCity
	RegionWater
	RegionDesert
	RegionMountain
	RegionDungeon
	RegionInterior
)

// Soundscape generates ambient audio for a specific region type.
type Soundscape struct {
	regionType RegionType
	genre      string
	sampleRate int
	rng        *rand.Rand
}

// NewSoundscape creates a soundscape for the given region and genre.
func NewSoundscape(region RegionType, genre string, seed int64) *Soundscape {
	return &Soundscape{
		regionType: region,
		genre:      genre,
		sampleRate: 44100,
		rng:        rand.New(rand.NewSource(seed)),
	}
}

// GenerateSamples creates ambient audio samples for the duration.
func (s *Soundscape) GenerateSamples(duration float64) []float64 {
	numSamples := int(duration * float64(s.sampleRate))
	var samples []float64

	switch s.regionType {
	case RegionPlains:
		samples = s.generateWindSamples(numSamples, 0.3)
	case RegionForest:
		samples = s.generateForestSamples(numSamples)
	case RegionCave:
		samples = s.generateCaveSamples(numSamples)
	case RegionCity:
		samples = s.generateCitySamples(numSamples)
	case RegionWater:
		samples = s.generateWaterSamples(numSamples)
	case RegionDesert:
		samples = s.generateDesertSamples(numSamples)
	case RegionMountain:
		samples = s.generateMountainSamples(numSamples)
	case RegionDungeon:
		samples = s.generateDungeonSamples(numSamples)
	case RegionInterior:
		samples = s.generateInteriorSamples(numSamples)
	default:
		samples = s.generateWindSamples(numSamples, 0.2)
	}

	// Apply genre-specific modifications
	samples = s.applyGenreModifications(samples)

	return samples
}

// generateFilteredNoise creates low-pass filtered noise samples.
// cutoff is the filter frequency in Hz, intensity is the noise amplitude.
func (s *Soundscape) generateFilteredNoise(numSamples int, cutoff, intensity float64) []float64 {
	samples := make([]float64, numSamples)
	alpha := 2 * math.Pi * cutoff / float64(s.sampleRate)
	alpha = alpha / (alpha + 1)

	prevSample := 0.0
	for i := 0; i < numSamples; i++ {
		noise := (s.rng.Float64()*2 - 1) * intensity
		sample := prevSample + alpha*(noise-prevSample)
		prevSample = sample
		samples[i] = sample
	}
	return samples
}

// generateWindSamples creates wind noise.
func (s *Soundscape) generateWindSamples(numSamples int, intensity float64) []float64 {
	samples := s.generateFilteredNoise(numSamples, 800.0, intensity)

	// Add slow modulation for gusts
	for i := range samples {
		gustMod := 0.7 + 0.3*math.Sin(float64(i)*0.0001)
		samples[i] *= gustMod
	}

	return samples
}

// generateForestSamples creates forest ambience (rustling leaves, birds).
func (s *Soundscape) generateForestSamples(numSamples int) []float64 {
	samples := s.generateWindSamples(numSamples, 0.15)
	s.addBirdChirps(samples, numSamples)
	s.addRustlingLeaves(samples, numSamples)
	return samples
}

// addBirdChirps adds occasional bird chirp sounds to samples.
func (s *Soundscape) addBirdChirps(samples []float64, numSamples int) {
	const chirpProbability = 0.00005
	chirpLen := int(0.1 * float64(s.sampleRate))

	for i := 0; i < numSamples; i++ {
		if s.rng.Float64() < chirpProbability {
			freq := 2000 + s.rng.Float64()*1500
			s.generateChirpAtPosition(samples, i, chirpLen, freq, numSamples)
		}
	}
}

// generateChirpAtPosition generates a single bird chirp starting at position i.
func (s *Soundscape) generateChirpAtPosition(samples []float64, startPos, chirpLen int, freq float64, numSamples int) {
	const chirpDuration = 0.1
	const chirpAmplitude = 0.3

	for j := 0; j < chirpLen && startPos+j < numSamples; j++ {
		t := float64(j) / float64(s.sampleRate)
		env := math.Sin(math.Pi * t / chirpDuration)
		chirp := math.Sin(2*math.Pi*freq*t) * env * chirpAmplitude
		samples[startPos+j] += chirp
	}
}

// addRustlingLeaves adds occasional rustling leaf sounds to samples.
func (s *Soundscape) addRustlingLeaves(samples []float64, numSamples int) {
	const rustleProbability = 0.0001
	rustleLen := int(0.05 * float64(s.sampleRate))

	for i := 0; i < numSamples; i++ {
		if s.rng.Float64() < rustleProbability {
			s.generateRustleAtPosition(samples, i, rustleLen, numSamples)
		}
	}
}

// generateRustleAtPosition generates a single rustle sound starting at position i.
func (s *Soundscape) generateRustleAtPosition(samples []float64, startPos, rustleLen, numSamples int) {
	const rustleAmplitude = 0.1

	for j := 0; j < rustleLen && startPos+j < numSamples; j++ {
		rustle := (s.rng.Float64()*2 - 1) * rustleAmplitude
		env := math.Sin(math.Pi * float64(j) / float64(rustleLen))
		samples[startPos+j] += rustle * env
	}
}

// generateCaveSamples creates cave ambience (drips, echoes).
func (s *Soundscape) generateCaveSamples(numSamples int) []float64 {
	samples := make([]float64, numSamples)
	s.generateCaveDrone(samples, numSamples)
	s.generateWaterDrips(samples, numSamples)
	s.applyCaveReverb(samples, numSamples)
	return samples
}

// generateCaveDrone adds a low ambient drone to the samples.
func (s *Soundscape) generateCaveDrone(samples []float64, numSamples int) {
	const droneFreq = 40.0
	const droneAmplitude = 0.05
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(s.sampleRate)
		samples[i] = math.Sin(2*math.Pi*droneFreq*t) * droneAmplitude
	}
}

// generateWaterDrips adds random water drip sounds to the samples.
func (s *Soundscape) generateWaterDrips(samples []float64, numSamples int) {
	const dripProbability = 0.00003
	dripLen := int(0.05 * float64(s.sampleRate))
	for i := 0; i < numSamples; i++ {
		if s.rng.Float64() < dripProbability {
			freq := 1500 + s.rng.Float64()*500
			s.generateDripAtPosition(samples, i, dripLen, freq, numSamples)
		}
	}
}

// generateDripAtPosition generates a single drip sound starting at position i.
func (s *Soundscape) generateDripAtPosition(samples []float64, startPos, dripLen int, freq float64, numSamples int) {
	const dripAmplitude = 0.2
	const decayRate = 50.0
	for j := 0; j < dripLen && startPos+j < numSamples; j++ {
		t := float64(j) / float64(s.sampleRate)
		env := math.Exp(-t * decayRate)
		drip := math.Sin(2*math.Pi*freq*t) * env * dripAmplitude
		samples[startPos+j] += drip
	}
}

// applyCaveReverb applies a simple delay-based reverb simulation.
func (s *Soundscape) applyCaveReverb(samples []float64, numSamples int) {
	delayLen := int(0.3 * float64(s.sampleRate))
	const reverbAmount = 0.3
	for i := delayLen; i < numSamples; i++ {
		samples[i] += samples[i-delayLen] * reverbAmount
	}
}

// generateCitySamples creates city ambience (crowd murmur, distant traffic).
func (s *Soundscape) generateCitySamples(numSamples int) []float64 {
	// Base murmur (filtered noise)
	samples := s.generateFilteredNoise(numSamples, 500.0, 0.15)

	// Distant traffic rumble
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(s.sampleRate)
		rumble := math.Sin(2*math.Pi*60*t) * 0.08
		rumble += math.Sin(2*math.Pi*80*t) * 0.05
		samples[i] += rumble
	}

	// Occasional footstep-like sounds
	for i := 0; i < numSamples; i++ {
		if s.rng.Float64() < 0.0001 {
			stepLen := int(0.08 * float64(s.sampleRate))
			for j := 0; j < stepLen && i+j < numSamples; j++ {
				t := float64(j) / float64(s.sampleRate)
				env := math.Exp(-t * 30)
				step := (s.rng.Float64()*2 - 1) * env * 0.1
				samples[i+j] += step
			}
		}
	}

	return samples
}

// generateWaterSamples creates water ambience (flowing, lapping).
func (s *Soundscape) generateWaterSamples(numSamples int) []float64 {
	samples := make([]float64, numSamples)

	// Flowing water (modulated noise)
	cutoff := 600.0
	alpha := 2 * math.Pi * cutoff / float64(s.sampleRate)
	alpha = alpha / (alpha + 1)

	prevSample := 0.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(s.sampleRate)
		mod := 0.8 + 0.2*math.Sin(2*math.Pi*0.3*t) // Slow modulation
		noise := (s.rng.Float64()*2 - 1) * 0.3 * mod
		sample := prevSample + alpha*(noise-prevSample)
		prevSample = sample
		samples[i] = sample
	}

	// Occasional splash
	for i := 0; i < numSamples; i++ {
		if s.rng.Float64() < 0.00002 {
			splashLen := int(0.15 * float64(s.sampleRate))
			for j := 0; j < splashLen && i+j < numSamples; j++ {
				t := float64(j) / float64(s.sampleRate)
				env := math.Exp(-t * 15)
				splash := (s.rng.Float64()*2 - 1) * env * 0.3
				samples[i+j] += splash
			}
		}
	}

	return samples
}

// generateDesertSamples creates desert ambience (wind, sand).
func (s *Soundscape) generateDesertSamples(numSamples int) []float64 {
	samples := s.generateWindSamples(numSamples, 0.25)

	// Add high-frequency sand hiss
	cutoffHigh := 3000.0
	alpha := 2 * math.Pi * cutoffHigh / float64(s.sampleRate)
	alpha = alpha / (alpha + 1)

	prevSand := 0.0
	for i := 0; i < numSamples; i++ {
		sandNoise := (s.rng.Float64()*2 - 1) * 0.05
		sandSample := prevSand + alpha*(sandNoise-prevSand)
		prevSand = sandSample
		samples[i] += sandSample
	}

	return samples
}

// generateMountainSamples creates mountain ambience (strong wind, echo).
func (s *Soundscape) generateMountainSamples(numSamples int) []float64 {
	samples := s.generateWindSamples(numSamples, 0.4)

	// Add echo effect
	delayLen := int(0.5 * float64(s.sampleRate))
	for i := delayLen; i < numSamples; i++ {
		samples[i] += samples[i-delayLen] * 0.2
	}

	return samples
}

// generateDungeonSamples creates dungeon ambience (drips, echoes, distant sounds).
func (s *Soundscape) generateDungeonSamples(numSamples int) []float64 {
	samples := s.generateCaveSamples(numSamples)

	// Add occasional distant metallic clang
	for i := 0; i < numSamples; i++ {
		if s.rng.Float64() < 0.00001 {
			clangLen := int(0.3 * float64(s.sampleRate))
			freq := 300 + s.rng.Float64()*200
			for j := 0; j < clangLen && i+j < numSamples; j++ {
				t := float64(j) / float64(s.sampleRate)
				env := math.Exp(-t * 10)
				// Metallic sound: multiple harmonics
				clang := (math.Sin(2*math.Pi*freq*t) +
					0.5*math.Sin(2*math.Pi*freq*2.3*t) +
					0.3*math.Sin(2*math.Pi*freq*3.7*t)) * env * 0.1
				samples[i+j] += clang
			}
		}
	}

	return samples
}

// generateInteriorSamples creates interior ambience (room tone, quiet).
func (s *Soundscape) generateInteriorSamples(numSamples int) []float64 {
	// Very quiet room tone using low-pass filtered noise
	return s.generateFilteredNoise(numSamples, 200.0, 0.03)
}

// applyHorrorModifications adds unsettling undertones for horror genre.
func (s *Soundscape) applyHorrorModifications(samples []float64) {
	for i := range samples {
		t := float64(i) / float64(s.sampleRate)
		undertone := math.Sin(2*math.Pi*37*t) * 0.02
		samples[i] += undertone
		if s.rng.Float64() < 0.00001 {
			samples[i] += (s.rng.Float64() - 0.5) * 0.1
		}
	}
}

// applyCyberpunkModifications adds electronic hum for cyberpunk genre.
func (s *Soundscape) applyCyberpunkModifications(samples []float64) {
	for i := range samples {
		t := float64(i) / float64(s.sampleRate)
		hum := math.Sin(2*math.Pi*60*t) * 0.04
		hum += math.Sin(2*math.Pi*120*t) * 0.02
		samples[i] += hum
	}
}

// applyPostApocModifications adds geiger-like clicks for post-apocalyptic genre.
func (s *Soundscape) applyPostApocModifications(samples []float64) {
	clickLen := int(0.005 * float64(s.sampleRate))
	for i := range samples {
		if s.rng.Float64() < 0.0002 {
			for j := 0; j < clickLen && i+j < len(samples); j++ {
				samples[i+j] += 0.15
			}
		}
	}
}

// applySciFiModifications adds subtle electronic ambience for sci-fi genre.
func (s *Soundscape) applySciFiModifications(samples []float64) {
	for i := range samples {
		t := float64(i) / float64(s.sampleRate)
		synth := math.Sin(2*math.Pi*220*t) * 0.01
		synth += math.Sin(2*math.Pi*330*t) * 0.005
		samples[i] += synth * (0.5 + 0.5*math.Sin(t*0.5))
	}
}

// applyGenreModifications adjusts ambient based on genre.
func (s *Soundscape) applyGenreModifications(samples []float64) []float64 {
	switch s.genre {
	case "horror":
		s.applyHorrorModifications(samples)
	case "cyberpunk":
		s.applyCyberpunkModifications(samples)
	case "post-apocalyptic":
		s.applyPostApocModifications(samples)
	case "sci-fi":
		s.applySciFiModifications(samples)
	}
	return samples
}

// RegionType returns the region type of this soundscape.
func (s *Soundscape) RegionType() RegionType {
	return s.regionType
}

// AmbientManager handles transitions between ambient soundscapes.
type AmbientManager struct {
	mu                 sync.Mutex
	currentRegion      RegionType
	previousRegion     RegionType
	currentSoundscape  *Soundscape
	genre              string
	seed               int64
	transitionTime     float64 // seconds
	transitionProgress float64
}

// NewAmbientManager creates a new ambient sound manager.
func NewAmbientManager(genre string, seed int64) *AmbientManager {
	return &AmbientManager{
		currentRegion:      RegionPlains,
		previousRegion:     RegionPlains,
		genre:              genre,
		seed:               seed,
		currentSoundscape:  NewSoundscape(RegionPlains, genre, seed),
		transitionTime:     1.0, // 1 second transition per AC
		transitionProgress: 1.0,
	}
}

// SetRegion changes the current region with transition.
func (am *AmbientManager) SetRegion(region RegionType) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if region != am.currentRegion {
		am.previousRegion = am.currentRegion
		am.currentRegion = region
		am.currentSoundscape = NewSoundscape(region, am.genre, am.seed)
		am.transitionProgress = 0.0
	}
}

// Update advances the ambient manager by dt seconds.
func (am *AmbientManager) Update(dt float64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.transitionProgress < 1.0 {
		am.transitionProgress += dt / am.transitionTime
		if am.transitionProgress > 1.0 {
			am.transitionProgress = 1.0
		}
	}
}

// GenerateSamples produces ambient audio for the current region.
func (am *AmbientManager) GenerateSamples(duration float64) []float64 {
	am.mu.Lock()
	defer am.mu.Unlock()

	return am.currentSoundscape.GenerateSamples(duration)
}

// GetCurrentRegion returns the current ambient region.
func (am *AmbientManager) GetCurrentRegion() RegionType {
	am.mu.Lock()
	defer am.mu.Unlock()
	return am.currentRegion
}

// IsTransitioning returns whether a region transition is in progress.
func (am *AmbientManager) IsTransitioning() bool {
	am.mu.Lock()
	defer am.mu.Unlock()
	return am.transitionProgress < 1.0
}
