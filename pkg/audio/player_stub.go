//go:build noebiten

// Package audio provides procedural audio synthesis.
package audio

// Player is a stub for non-Ebitengine builds.
type Player struct {
	samples   []float64
	isPlaying bool
}

// NewPlayer creates a stub player for testing.
func NewPlayer() (*Player, error) {
	return &Player{
		samples: make([]float64, 0),
	}, nil
}

// QueueSamples stores samples for testing verification.
func (p *Player) QueueSamples(samples []float64) {
	p.samples = append(p.samples, samples...)
}

// Play marks the player as playing.
func (p *Player) Play() {
	p.isPlaying = true
}

// Pause marks the player as paused.
func (p *Player) Pause() {
	p.isPlaying = false
}

// IsPlaying returns whether the player is playing.
func (p *Player) IsPlaying() bool {
	return p.isPlaying
}
