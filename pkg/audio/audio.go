// Package audio provides procedural audio synthesis.
package audio

// Engine handles procedural audio generation and playback.
type Engine struct {
	SampleRate int
	Genre      string
}

// NewEngine creates a new audio engine.
func NewEngine(genre string) *Engine {
	return &Engine{
		SampleRate: 44100,
		Genre:      genre,
	}
}

// Update advances the audio engine by one tick.
func (e *Engine) Update() {}
