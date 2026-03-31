// Package audio provides procedural audio synthesis.
package audio

import (
	"math"
)

// UISoundType represents different UI sound effect types.
type UISoundType int

const (
	// UISoundMenuSelect is played when selecting a menu item.
	UISoundMenuSelect UISoundType = iota
	// UISoundMenuNavigate is played when navigating menu options.
	UISoundMenuNavigate
	// UISoundButtonClick is played when clicking a button.
	UISoundButtonClick
	// UISoundButtonHover is played when hovering over a button.
	UISoundButtonHover
	// UISoundNotification is played for notifications.
	UISoundNotification
	// UISoundError is played for error states.
	UISoundError
	// UISoundSuccess is played for successful actions.
	UISoundSuccess
	// UISoundInventoryOpen is played when opening inventory.
	UISoundInventoryOpen
	// UISoundInventoryClose is played when closing inventory.
	UISoundInventoryClose
	// UISoundItemPickup is played when picking up an item.
	UISoundItemPickup
	// UISoundItemDrop is played when dropping an item.
	UISoundItemDrop
	// UISoundGoldCoins is played for currency transactions.
	UISoundGoldCoins
	// UISoundLevelUp is played on level up.
	UISoundLevelUp
	// UISoundQuestComplete is played on quest completion.
	UISoundQuestComplete
	// UISoundQuestAccept is played when accepting a quest.
	UISoundQuestAccept
	// UISoundMapOpen is played when opening the map.
	UISoundMapOpen
	// UISoundDialogAdvance is played when advancing dialog.
	UISoundDialogAdvance
)

// UISoundGenerator generates procedural UI sound effects.
type UISoundGenerator struct {
	engine *Engine
}

// NewUISoundGenerator creates a new UI sound generator.
func NewUISoundGenerator(engine *Engine) *UISoundGenerator {
	return &UISoundGenerator{engine: engine}
}

// Generate creates samples for the specified UI sound type.
func (g *UISoundGenerator) Generate(soundType UISoundType) []float64 {
	switch soundType {
	case UISoundMenuSelect:
		return g.generateMenuSelect()
	case UISoundMenuNavigate:
		return g.generateMenuNavigate()
	case UISoundButtonClick:
		return g.generateButtonClick()
	case UISoundButtonHover:
		return g.generateButtonHover()
	case UISoundNotification:
		return g.generateNotification()
	case UISoundError:
		return g.generateError()
	case UISoundSuccess:
		return g.generateSuccess()
	case UISoundInventoryOpen:
		return g.generateInventoryOpen()
	case UISoundInventoryClose:
		return g.generateInventoryClose()
	case UISoundItemPickup:
		return g.generateItemPickup()
	case UISoundItemDrop:
		return g.generateItemDrop()
	case UISoundGoldCoins:
		return g.generateGoldCoins()
	case UISoundLevelUp:
		return g.generateLevelUp()
	case UISoundQuestComplete:
		return g.generateQuestComplete()
	case UISoundQuestAccept:
		return g.generateQuestAccept()
	case UISoundMapOpen:
		return g.generateMapOpen()
	case UISoundDialogAdvance:
		return g.generateDialogAdvance()
	default:
		return g.generateButtonClick()
	}
}

// generateMenuSelect creates a pleasant two-note ascending tone.
func (g *UISoundGenerator) generateMenuSelect() []float64 {
	baseFreq := g.getUIBaseFrequency()
	samples1 := g.generateUITone(baseFreq, 0.05)
	samples2 := g.generateUITone(baseFreq*1.25, 0.08)
	return g.concatenate(samples1, samples2)
}

// generateMenuNavigate creates a quick blip sound.
func (g *UISoundGenerator) generateMenuNavigate() []float64 {
	baseFreq := g.getUIBaseFrequency() * 1.5
	return g.generateUITone(baseFreq, 0.03)
}

// generateButtonClick creates a tactile click sound.
func (g *UISoundGenerator) generateButtonClick() []float64 {
	baseFreq := g.getUIBaseFrequency() * 2.0
	samples := g.generateUITone(baseFreq, 0.025)
	return g.applyClickEnvelope(samples)
}

// generateButtonHover creates a subtle hover indication.
func (g *UISoundGenerator) generateButtonHover() []float64 {
	baseFreq := g.getUIBaseFrequency() * 1.8
	samples := g.generateUITone(baseFreq, 0.015)
	return g.applyFadeEnvelope(samples, 0.3)
}

// generateNotification creates an attention-grabbing tone.
func (g *UISoundGenerator) generateNotification() []float64 {
	baseFreq := g.getUIBaseFrequency()
	samples1 := g.generateUITone(baseFreq, 0.06)
	samples2 := g.generateUITone(baseFreq*1.5, 0.06)
	samples3 := g.generateUITone(baseFreq*2.0, 0.08)
	return g.concatenate(g.concatenate(samples1, samples2), samples3)
}

// generateError creates a dissonant warning sound.
func (g *UISoundGenerator) generateError() []float64 {
	baseFreq := g.getUIBaseFrequency() * 0.7
	// Two notes that create mild dissonance
	samples1 := g.generateUITone(baseFreq, 0.1)
	samples2 := g.generateUITone(baseFreq*1.059, 0.15)
	return g.mixSamples(samples1, samples2, 0.6, 0.4)
}

// generateSuccess creates a pleasant ascending fanfare.
func (g *UISoundGenerator) generateSuccess() []float64 {
	baseFreq := g.getUIBaseFrequency()
	samples1 := g.generateUITone(baseFreq, 0.04)
	samples2 := g.generateUITone(baseFreq*1.25, 0.04)
	samples3 := g.generateUITone(baseFreq*1.5, 0.06)
	return g.concatenate(g.concatenate(samples1, samples2), samples3)
}

// generateInventoryOpen creates a bag/pack opening sound.
func (g *UISoundGenerator) generateInventoryOpen() []float64 {
	baseFreq := g.getUIBaseFrequency() * 0.8
	samples := g.generateUITone(baseFreq, 0.12)
	return g.applySweep(samples, 1.0, 1.3)
}

// generateInventoryClose creates a bag/pack closing sound.
func (g *UISoundGenerator) generateInventoryClose() []float64 {
	baseFreq := g.getUIBaseFrequency() * 0.8
	samples := g.generateUITone(baseFreq*1.3, 0.1)
	return g.applySweep(samples, 1.0, 0.75)
}

// generateItemPickup creates an item collection sound.
func (g *UISoundGenerator) generateItemPickup() []float64 {
	baseFreq := g.getUIBaseFrequency() * 1.2
	samples := g.generateUITone(baseFreq, 0.06)
	return g.applySweep(samples, 1.0, 1.4)
}

// generateItemDrop creates an item release sound.
func (g *UISoundGenerator) generateItemDrop() []float64 {
	baseFreq := g.getUIBaseFrequency() * 1.3
	samples := g.generateUITone(baseFreq, 0.05)
	return g.applySweep(samples, 1.0, 0.7)
}

// generateGoldCoins creates a coin jingle sound.
func (g *UISoundGenerator) generateGoldCoins() []float64 {
	baseFreq := g.getUIBaseFrequency() * 2.5
	samples := make([]float64, 0)
	for i := 0; i < 3; i++ {
		freq := baseFreq * (1.0 + float64(i)*0.15)
		tone := g.generateUITone(freq, 0.04)
		tone = g.applyMetallicSheen(tone)
		samples = g.concatenate(samples, tone)
	}
	return samples
}

// generateLevelUp creates an impressive level-up fanfare.
func (g *UISoundGenerator) generateLevelUp() []float64 {
	baseFreq := g.getUIBaseFrequency()
	// Major chord arpeggio
	samples1 := g.generateUITone(baseFreq, 0.08)
	samples2 := g.generateUITone(baseFreq*1.25, 0.08)
	samples3 := g.generateUITone(baseFreq*1.5, 0.08)
	samples4 := g.generateUITone(baseFreq*2.0, 0.15)

	result := g.concatenate(samples1, samples2)
	result = g.concatenate(result, samples3)
	result = g.concatenate(result, samples4)
	return result
}

// generateQuestComplete creates a triumphant completion sound.
func (g *UISoundGenerator) generateQuestComplete() []float64 {
	baseFreq := g.getUIBaseFrequency()
	// Victory fanfare
	samples1 := g.generateUITone(baseFreq*1.5, 0.1)
	samples2 := g.generateUITone(baseFreq*1.5, 0.1)
	samples3 := g.generateUITone(baseFreq*2.0, 0.2)

	result := g.concatenate(samples1, samples2)
	result = g.concatenate(result, samples3)
	return result
}

// generateQuestAccept creates a quest acceptance confirmation.
func (g *UISoundGenerator) generateQuestAccept() []float64 {
	baseFreq := g.getUIBaseFrequency()
	samples1 := g.generateUITone(baseFreq, 0.06)
	samples2 := g.generateUITone(baseFreq*1.5, 0.1)
	return g.concatenate(samples1, samples2)
}

// generateMapOpen creates a paper unfolding sound.
func (g *UISoundGenerator) generateMapOpen() []float64 {
	baseFreq := g.getUIBaseFrequency() * 0.6
	samples := g.generateNoise(0.15)
	samples = g.applyBandpass(samples, baseFreq, 2.0)
	return g.engine.ApplyADSR(samples, 0.02, 0.05, 0.5, 0.08)
}

// generateDialogAdvance creates a subtle page turn/advance sound.
func (g *UISoundGenerator) generateDialogAdvance() []float64 {
	baseFreq := g.getUIBaseFrequency() * 1.6
	return g.generateUITone(baseFreq, 0.02)
}

// getUIBaseFrequency returns genre-appropriate UI base frequency.
func (g *UISoundGenerator) getUIBaseFrequency() float64 {
	switch g.engine.Genre {
	case "fantasy":
		return 523.25 // C5 - bright, pleasant
	case "sci-fi":
		return 659.26 // E5 - high-tech feel
	case "horror":
		return 311.13 // Eb4 - unsettling
	case "cyberpunk":
		return 783.99 // G5 - digital, sharp
	case "post-apocalyptic":
		return 392.00 // G4 - muted, dusty
	default:
		return 523.25
	}
}

// generateUITone generates a basic tone for UI sounds.
func (g *UISoundGenerator) generateUITone(frequency, duration float64) []float64 {
	numSamples := int(duration * float64(g.engine.SampleRate))
	samples := make([]float64, numSamples)
	phaseIncrement := 2 * math.Pi * frequency / float64(g.engine.SampleRate)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		// Mix sine and triangle for richer UI tone
		sine := math.Sin(phase)
		triangle := 2*math.Abs(2*(phase/(2*math.Pi)-math.Floor(phase/(2*math.Pi)+0.5))) - 1
		samples[i] = sine*0.7 + triangle*0.3

		phase += phaseIncrement
		if phase > 2*math.Pi {
			phase -= 2 * math.Pi
		}
	}

	return g.engine.ApplyADSR(samples, 0.005, 0.02, 0.6, 0.05)
}

// generateNoise creates white noise for texture sounds.
func (g *UISoundGenerator) generateNoise(duration float64) []float64 {
	numSamples := int(duration * float64(g.engine.SampleRate))
	samples := make([]float64, numSamples)

	// Simple PRNG for deterministic noise
	state := uint32(12345)
	for i := 0; i < numSamples; i++ {
		// Xorshift32
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		samples[i] = (float64(state)/float64(^uint32(0)))*2 - 1
	}

	return samples
}

// concatenate joins two sample arrays.
func (g *UISoundGenerator) concatenate(a, b []float64) []float64 {
	result := make([]float64, len(a)+len(b))
	copy(result, a)
	copy(result[len(a):], b)
	return result
}

// mixSamples mixes two sample arrays with given weights.
func (g *UISoundGenerator) mixSamples(a, b []float64, wa, wb float64) []float64 {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	result := make([]float64, maxLen)
	for i := 0; i < maxLen; i++ {
		va, vb := 0.0, 0.0
		if i < len(a) {
			va = a[i]
		}
		if i < len(b) {
			vb = b[i]
		}
		result[i] = va*wa + vb*wb
	}
	return result
}

// applyClickEnvelope applies a sharp click envelope.
func (g *UISoundGenerator) applyClickEnvelope(samples []float64) []float64 {
	result := make([]float64, len(samples))
	for i, s := range samples {
		t := float64(i) / float64(len(samples))
		envelope := math.Exp(-t * 20)
		result[i] = s * envelope
	}
	return result
}

// applyFadeEnvelope applies a fade envelope with configurable speed.
func (g *UISoundGenerator) applyFadeEnvelope(samples []float64, volume float64) []float64 {
	result := make([]float64, len(samples))
	for i, s := range samples {
		t := float64(i) / float64(len(samples))
		envelope := volume * (1 - t)
		result[i] = s * envelope
	}
	return result
}

// applySweep applies a frequency sweep multiplier over time.
func (g *UISoundGenerator) applySweep(samples []float64, startMult, endMult float64) []float64 {
	result := make([]float64, len(samples))
	for i, s := range samples {
		t := float64(i) / float64(len(samples))
		mult := startMult + (endMult-startMult)*t
		result[i] = s * mult
	}
	return result
}

// applyMetallicSheen adds metallic harmonics for coin-like sounds.
func (g *UISoundGenerator) applyMetallicSheen(samples []float64) []float64 {
	result := make([]float64, len(samples))
	for i := range samples {
		harmonic := math.Sin(float64(i)*0.05) * 0.2
		harmonic2 := math.Sin(float64(i)*0.08) * 0.1
		result[i] = samples[i]*0.7 + harmonic + harmonic2
	}
	return result
}

// applyBandpass applies a simple bandpass filter.
func (g *UISoundGenerator) applyBandpass(samples []float64, centerFreq, bandwidth float64) []float64 {
	result := make([]float64, len(samples))
	sampleRate := float64(g.engine.SampleRate)

	// Simple resonant filter approximation
	f := centerFreq / sampleRate
	q := centerFreq / bandwidth
	if q < 0.5 {
		q = 0.5
	}

	// State variables
	low, band, high := 0.0, 0.0, 0.0

	for i, s := range samples {
		// State variable filter
		low = low + f*band
		high = s - low - band/q
		band = f*high + band
		result[i] = band * 0.5
	}

	return result
}

// UISoundPlayer manages playback of UI sounds with priority and deduplication.
type UISoundPlayer struct {
	generator    *UISoundGenerator
	lastPlayTime map[UISoundType]float64
	minInterval  float64 // minimum seconds between same sound type
}

// NewUISoundPlayer creates a new UI sound player.
func NewUISoundPlayer(engine *Engine) *UISoundPlayer {
	return &UISoundPlayer{
		generator:    NewUISoundGenerator(engine),
		lastPlayTime: make(map[UISoundType]float64),
		minInterval:  0.05,
	}
}

// CanPlay checks if a sound can be played (not too recently played).
func (p *UISoundPlayer) CanPlay(soundType UISoundType, currentTime float64) bool {
	lastTime, exists := p.lastPlayTime[soundType]
	if !exists {
		return true
	}
	return currentTime-lastTime >= p.minInterval
}

// GetSamples generates samples for a UI sound if allowed.
func (p *UISoundPlayer) GetSamples(soundType UISoundType, currentTime float64) []float64 {
	if !p.CanPlay(soundType, currentTime) {
		return nil
	}
	p.lastPlayTime[soundType] = currentTime
	return p.generator.Generate(soundType)
}

// SetMinInterval sets the minimum interval between same sound type.
func (p *UISoundPlayer) SetMinInterval(interval float64) {
	if interval >= 0 {
		p.minInterval = interval
	}
}

// GetMinInterval returns the minimum interval between same sound type.
func (p *UISoundPlayer) GetMinInterval() float64 {
	return p.minInterval
}

// Reset clears the play time history.
func (p *UISoundPlayer) Reset() {
	p.lastPlayTime = make(map[UISoundType]float64)
}
