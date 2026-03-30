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

// ============================================================================
// Reverb Effects System
// ============================================================================

// ReverbConfig holds reverb effect parameters.
type ReverbConfig struct {
	RoomSize  float64 // 0-1, larger = longer reverb tail
	Damping   float64 // 0-1, higher = more high frequency absorption
	WetMix    float64 // 0-1, amount of reverb in final mix
	DryMix    float64 // 0-1, amount of original in final mix
	PreDelay  float64 // seconds of delay before reverb starts
	DecayTime float64 // seconds for reverb to decay
}

// DefaultReverbConfig returns a standard reverb configuration.
func DefaultReverbConfig() *ReverbConfig {
	return &ReverbConfig{
		RoomSize:  0.5,
		Damping:   0.4,
		WetMix:    0.3,
		DryMix:    0.7,
		PreDelay:  0.02,
		DecayTime: 1.5,
	}
}

// GenreReverbConfig returns reverb settings appropriate for a genre.
func GenreReverbConfig(genre string) *ReverbConfig {
	switch genre {
	case "fantasy":
		return &ReverbConfig{
			RoomSize:  0.6,
			Damping:   0.3,
			WetMix:    0.35,
			DryMix:    0.65,
			PreDelay:  0.025,
			DecayTime: 1.8,
		}
	case "sci-fi":
		return &ReverbConfig{
			RoomSize:  0.7,
			Damping:   0.2,
			WetMix:    0.4,
			DryMix:    0.6,
			PreDelay:  0.03,
			DecayTime: 2.0,
		}
	case "horror":
		return &ReverbConfig{
			RoomSize:  0.9,
			Damping:   0.5,
			WetMix:    0.5,
			DryMix:    0.5,
			PreDelay:  0.05,
			DecayTime: 3.0,
		}
	case "cyberpunk":
		return &ReverbConfig{
			RoomSize:  0.4,
			Damping:   0.6,
			WetMix:    0.25,
			DryMix:    0.75,
			PreDelay:  0.01,
			DecayTime: 1.0,
		}
	case "post-apocalyptic":
		return &ReverbConfig{
			RoomSize:  0.8,
			Damping:   0.4,
			WetMix:    0.45,
			DryMix:    0.55,
			PreDelay:  0.04,
			DecayTime: 2.5,
		}
	default:
		return DefaultReverbConfig()
	}
}

// ReverbProcessor applies reverb effects to audio.
type ReverbProcessor struct {
	config     *ReverbConfig
	sampleRate int

	// Delay lines for Schroeder reverb algorithm
	combDelays     [][]float64
	combIndices    []int
	allpassDelays  [][]float64
	allpassIndices []int

	// Comb filter feedback gains
	combGains []float64
}

// NewReverbProcessor creates a new reverb processor.
func NewReverbProcessor(config *ReverbConfig, sampleRate int) *ReverbProcessor {
	rp := &ReverbProcessor{
		config:     config,
		sampleRate: sampleRate,
	}
	rp.initialize()
	return rp
}

// initialize sets up the reverb delay lines.
func (rp *ReverbProcessor) initialize() {
	// Schroeder reverb uses 4 parallel comb filters + 2 series allpass filters
	combDelayMs := []float64{29.7, 37.1, 41.1, 43.7}
	allpassDelayMs := []float64{5.0, 1.7}

	rp.combDelays = make([][]float64, 4)
	rp.combIndices = make([]int, 4)
	rp.combGains = make([]float64, 4)

	for i, delayMs := range combDelayMs {
		// Scale delay by room size
		scaledDelay := delayMs * rp.config.RoomSize * 2
		delaySamples := int(scaledDelay * float64(rp.sampleRate) / 1000.0)
		if delaySamples < 1 {
			delaySamples = 1
		}
		rp.combDelays[i] = make([]float64, delaySamples)
		rp.combIndices[i] = 0

		// Calculate feedback gain based on decay time and delay
		rt60 := rp.config.DecayTime
		gain := math.Pow(0.001, float64(delaySamples)/float64(rp.sampleRate)/rt60)
		// Apply damping
		rp.combGains[i] = gain * (1.0 - rp.config.Damping*0.3)
	}

	rp.allpassDelays = make([][]float64, 2)
	rp.allpassIndices = make([]int, 2)

	for i, delayMs := range allpassDelayMs {
		delaySamples := int(delayMs * float64(rp.sampleRate) / 1000.0)
		if delaySamples < 1 {
			delaySamples = 1
		}
		rp.allpassDelays[i] = make([]float64, delaySamples)
		rp.allpassIndices[i] = 0
	}
}

// Process applies reverb to the input samples.
func (rp *ReverbProcessor) Process(input []float64) []float64 {
	output := make([]float64, len(input))

	// Pre-delay (circular buffer simulation)
	preDelaySamples := int(rp.config.PreDelay * float64(rp.sampleRate))
	if preDelaySamples > len(input) {
		preDelaySamples = len(input)
	}

	for i := range input {
		// Get input with pre-delay
		delayedIdx := i - preDelaySamples
		var delayedInput float64
		if delayedIdx >= 0 {
			delayedInput = input[delayedIdx]
		}

		// Sum of 4 parallel comb filters
		combSum := 0.0
		for c := 0; c < 4; c++ {
			combSum += rp.processComb(c, delayedInput)
		}
		combSum /= 4.0

		// Series allpass filters
		allpassOut := rp.processAllpass(0, combSum)
		allpassOut = rp.processAllpass(1, allpassOut)

		// Mix dry and wet
		output[i] = input[i]*rp.config.DryMix + allpassOut*rp.config.WetMix
	}

	return output
}

// processComb processes a single comb filter.
func (rp *ReverbProcessor) processComb(index int, input float64) float64 {
	delay := rp.combDelays[index]
	idx := rp.combIndices[index]

	// Read from delay line
	output := delay[idx]

	// Write new value with feedback
	delay[idx] = input + output*rp.combGains[index]

	// Advance index
	rp.combIndices[index] = (idx + 1) % len(delay)

	return output
}

// processAllpass processes a single allpass filter.
func (rp *ReverbProcessor) processAllpass(index int, input float64) float64 {
	delay := rp.allpassDelays[index]
	idx := rp.allpassIndices[index]
	g := 0.5 // allpass coefficient

	// Read from delay line
	delayed := delay[idx]

	// Allpass formula
	output := -g*input + delayed + g*delayed

	// Write to delay line
	delay[idx] = input + delayed*g

	// Advance index
	rp.allpassIndices[index] = (idx + 1) % len(delay)

	return output
}

// Reset clears the reverb delay lines.
func (rp *ReverbProcessor) Reset() {
	for i := range rp.combDelays {
		for j := range rp.combDelays[i] {
			rp.combDelays[i][j] = 0
		}
		rp.combIndices[i] = 0
	}
	for i := range rp.allpassDelays {
		for j := range rp.allpassDelays[i] {
			rp.allpassDelays[i][j] = 0
		}
		rp.allpassIndices[i] = 0
	}
}

// SetConfig updates the reverb configuration.
func (rp *ReverbProcessor) SetConfig(config *ReverbConfig) {
	rp.config = config
	rp.initialize()
}

// GetConfig returns the current reverb configuration.
func (rp *ReverbProcessor) GetConfig() *ReverbConfig {
	return rp.config
}

// ApplyReverb applies reverb to samples using the engine's genre settings.
func (e *Engine) ApplyReverb(samples []float64) []float64 {
	config := GenreReverbConfig(e.Genre)
	processor := NewReverbProcessor(config, e.SampleRate)
	return processor.Process(samples)
}

// VoiceSynthesizer generates synthetic speech for dialog.
type VoiceSynthesizer struct {
	// SampleRate is the audio sample rate.
	SampleRate int
	// Genre affects voice characteristics.
	Genre string
	// VoiceProfiles maps NPC types to voice configurations.
	VoiceProfiles map[string]*VoiceProfile
}

// VoiceProfile defines characteristics for synthetic speech.
type VoiceProfile struct {
	// BaseFrequency is the fundamental pitch in Hz.
	BaseFrequency float64
	// FrequencyVariation adds pitch variation (0-1).
	FrequencyVariation float64
	// SpeakingRate affects speech speed (1.0 = normal).
	SpeakingRate float64
	// Breathiness adds noise to the voice (0-1).
	Breathiness float64
	// Roughness adds harmonic distortion (0-1).
	Roughness float64
	// Vibrato is the amount of pitch oscillation (0-1).
	Vibrato float64
	// VibratoRate is the vibrato frequency in Hz.
	VibratoRate float64
	// FormantShift adjusts formant frequencies (-1 to 1).
	FormantShift float64
	// ResonanceBoost boosts resonant frequencies (0-1).
	ResonanceBoost float64
}

// NewVoiceSynthesizer creates a new voice synthesizer.
func NewVoiceSynthesizer(sampleRate int, genre string) *VoiceSynthesizer {
	vs := &VoiceSynthesizer{
		SampleRate:    sampleRate,
		Genre:         genre,
		VoiceProfiles: make(map[string]*VoiceProfile),
	}
	vs.initializeProfiles()
	return vs
}

// initializeProfiles sets up default voice profiles.
func (vs *VoiceSynthesizer) initializeProfiles() {
	// Male deep voice (guards, warriors)
	vs.VoiceProfiles["male_deep"] = &VoiceProfile{
		BaseFrequency:      110,
		FrequencyVariation: 0.1,
		SpeakingRate:       0.9,
		Breathiness:        0.1,
		Roughness:          0.2,
		Vibrato:            0.05,
		VibratoRate:        5.0,
		FormantShift:       -0.2,
		ResonanceBoost:     0.3,
	}

	// Male medium voice (merchants, commoners)
	vs.VoiceProfiles["male_medium"] = &VoiceProfile{
		BaseFrequency:      140,
		FrequencyVariation: 0.15,
		SpeakingRate:       1.0,
		Breathiness:        0.15,
		Roughness:          0.1,
		Vibrato:            0.1,
		VibratoRate:        5.5,
		FormantShift:       0.0,
		ResonanceBoost:     0.2,
	}

	// Female voice
	vs.VoiceProfiles["female"] = &VoiceProfile{
		BaseFrequency:      220,
		FrequencyVariation: 0.2,
		SpeakingRate:       1.1,
		Breathiness:        0.2,
		Roughness:          0.05,
		Vibrato:            0.15,
		VibratoRate:        6.0,
		FormantShift:       0.3,
		ResonanceBoost:     0.4,
	}

	// Elderly voice
	vs.VoiceProfiles["elderly"] = &VoiceProfile{
		BaseFrequency:      130,
		FrequencyVariation: 0.25,
		SpeakingRate:       0.8,
		Breathiness:        0.3,
		Roughness:          0.25,
		Vibrato:            0.2,
		VibratoRate:        4.5,
		FormantShift:       -0.1,
		ResonanceBoost:     0.2,
	}

	// Child voice
	vs.VoiceProfiles["child"] = &VoiceProfile{
		BaseFrequency:      300,
		FrequencyVariation: 0.25,
		SpeakingRate:       1.2,
		Breathiness:        0.25,
		Roughness:          0.0,
		Vibrato:            0.05,
		VibratoRate:        6.5,
		FormantShift:       0.5,
		ResonanceBoost:     0.5,
	}

	// Monster/creature voice
	vs.VoiceProfiles["creature"] = &VoiceProfile{
		BaseFrequency:      80,
		FrequencyVariation: 0.3,
		SpeakingRate:       0.7,
		Breathiness:        0.4,
		Roughness:          0.5,
		Vibrato:            0.1,
		VibratoRate:        3.0,
		FormantShift:       -0.5,
		ResonanceBoost:     0.6,
	}

	// Robot voice (sci-fi)
	vs.VoiceProfiles["robot"] = &VoiceProfile{
		BaseFrequency:      160,
		FrequencyVariation: 0.05,
		SpeakingRate:       1.0,
		Breathiness:        0.0,
		Roughness:          0.0,
		Vibrato:            0.0,
		VibratoRate:        0.0,
		FormantShift:       0.0,
		ResonanceBoost:     0.8,
	}

	// Whisper voice
	vs.VoiceProfiles["whisper"] = &VoiceProfile{
		BaseFrequency:      150,
		FrequencyVariation: 0.1,
		SpeakingRate:       0.85,
		Breathiness:        0.9,
		Roughness:          0.0,
		Vibrato:            0.0,
		VibratoRate:        0.0,
		FormantShift:       0.0,
		ResonanceBoost:     0.1,
	}
}

// SynthesizeSpeech generates audio samples for a text phrase.
func (vs *VoiceSynthesizer) SynthesizeSpeech(text, voiceType string) []float64 {
	profile, ok := vs.VoiceProfiles[voiceType]
	if !ok {
		profile = vs.VoiceProfiles["male_medium"]
	}

	// Simple formant synthesis approximation
	// Each character gets a short "phoneme" duration
	phonemeDuration := 0.08 / profile.SpeakingRate
	totalDuration := float64(len(text)) * phonemeDuration
	numSamples := int(totalDuration * float64(vs.SampleRate))

	samples := make([]float64, numSamples)
	phase := 0.0
	vibratoPhase := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(vs.SampleRate)
		charIndex := int(t / phonemeDuration)
		if charIndex >= len(text) {
			charIndex = len(text) - 1
		}

		// Get character for formant modulation
		char := ' '
		if charIndex < len(text) {
			char = rune(text[charIndex])
		}

		// Calculate frequency with vibrato
		vibratoPhase += 2 * math.Pi * profile.VibratoRate / float64(vs.SampleRate)
		vibrato := profile.Vibrato * math.Sin(vibratoPhase)

		freq := profile.BaseFrequency * (1 + vibrato)
		freq *= 1 + (getCharacterPitch(char) * profile.FrequencyVariation)

		// Advance phase
		phase += 2 * math.Pi * freq / float64(vs.SampleRate)

		// Generate voice signal
		voice := vs.generateVoiceWaveform(phase, profile, char)

		// Apply amplitude envelope for phoneme
		phonemeT := math.Mod(t, phonemeDuration) / phonemeDuration
		envelope := vs.phonemeEnvelope(phonemeT)

		samples[i] = voice * envelope * 0.3 // Scale to avoid clipping
	}

	return samples
}

// generateVoiceWaveform creates the voice signal for one sample.
func (vs *VoiceSynthesizer) generateVoiceWaveform(phase float64, profile *VoiceProfile, char rune) float64 {
	// Base glottal pulse (simplified using harmonics)
	signal := math.Sin(phase)
	signal += 0.5 * math.Sin(2*phase)  // Second harmonic
	signal += 0.25 * math.Sin(3*phase) // Third harmonic

	// Add roughness (higher harmonics)
	if profile.Roughness > 0 {
		signal += profile.Roughness * 0.3 * math.Sin(5*phase)
		signal += profile.Roughness * 0.15 * math.Sin(7*phase)
	}

	// Add breathiness (noise)
	if profile.Breathiness > 0 {
		noise := (rand() - 0.5) * 2 * profile.Breathiness
		signal = signal*(1-profile.Breathiness*0.5) + noise
	}

	// Apply formant filtering based on character (simplified)
	formantMod := getFormantModifier(char)
	signal *= formantMod

	// Resonance boost
	if profile.ResonanceBoost > 0 {
		signal *= 1 + profile.ResonanceBoost*0.5
	}

	return signal
}

// phonemeEnvelope generates an amplitude envelope for a phoneme.
func (vs *VoiceSynthesizer) phonemeEnvelope(t float64) float64 {
	// ADSR-like envelope within each phoneme
	attack := 0.1
	decay := 0.1
	sustain := 0.7
	release := 0.1

	if t < attack {
		return t / attack
	} else if t < attack+decay {
		return 1.0 - (1.0-sustain)*(t-attack)/decay
	} else if t < 1.0-release {
		return sustain
	} else {
		return sustain * (1.0 - (t-(1.0-release))/release)
	}
}

// getCharacterPitch returns pitch modification for a character.
func getCharacterPitch(char rune) float64 {
	// Vowels have slightly higher pitch
	switch char {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return 0.1
	case '!', '?':
		return 0.2
	case '.':
		return -0.1
	default:
		return 0
	}
}

// getFormantModifier returns amplitude modification for a character.
func getFormantModifier(char rune) float64 {
	// Different characters produce different resonances
	switch char {
	case 'a', 'A':
		return 1.0
	case 'e', 'E':
		return 0.9
	case 'i', 'I':
		return 0.85
	case 'o', 'O':
		return 0.95
	case 'u', 'U':
		return 0.8
	case ' ', '.', ',':
		return 0.1 // Quieter for pauses
	default:
		return 0.7
	}
}

// Simple pseudo-random for voice synthesis (deterministic within sample).
var randState uint32 = 12345

func rand() float64 {
	randState = randState*1103515245 + 12345
	return float64(randState%65536) / 65536.0
}

// GetGenreVoiceProfile returns a voice profile appropriate for the genre.
func (vs *VoiceSynthesizer) GetGenreVoiceProfile(occupation string) string {
	switch vs.Genre {
	case "sci-fi":
		if occupation == "technician" || occupation == "scientist" {
			return "robot"
		}
	case "horror":
		if occupation == "priest" || occupation == "mortician" {
			return "whisper"
		}
	case "fantasy":
		if occupation == "creature" || occupation == "monster" {
			return "creature"
		}
	}

	// Default selection based on occupation
	switch occupation {
	case "guard", "blacksmith", "warrior":
		return "male_deep"
	case "healer", "innkeeper":
		return "female"
	case "priest", "scribe":
		return "elderly"
	default:
		return "male_medium"
	}
}

// GenerateDialogAudio creates audio for NPC dialog text.
func (e *Engine) GenerateDialogAudio(text, npcType string) []float64 {
	vs := NewVoiceSynthesizer(e.SampleRate, e.Genre)
	voiceProfile := vs.GetGenreVoiceProfile(npcType)
	return vs.SynthesizeSpeech(text, voiceProfile)
}
