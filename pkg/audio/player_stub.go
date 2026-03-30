//go:build noebiten

// Package audio provides procedural audio synthesis.
package audio

import "sync"

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

// SampleStream is a stub for non-Ebitengine builds.
type SampleStream struct {
	mu      sync.Mutex
	buffer  []byte
	silence []byte
}

// NewSampleStream creates a new sample stream.
func NewSampleStream() *SampleStream {
	silence := make([]byte, 4096)
	return &SampleStream{
		buffer:  make([]byte, 0, AudioSampleRate*BytesPerSample),
		silence: silence,
	}
}

// QueueSamples converts float64 samples to 16-bit PCM stereo and queues them (stub).
func (s *SampleStream) QueueSamples(samples []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Stub: just track that samples were queued
	for range samples {
		s.buffer = append(s.buffer, 0, 0, 0, 0) // 4 bytes per stereo sample
	}
}

// Read implements io.Reader for audio playback (stub).
func (s *SampleStream) Read(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.buffer) == 0 {
		copy(p, s.silence)
		return len(p), nil
	}

	n = copy(p, s.buffer)
	s.buffer = s.buffer[n:]
	return n, nil
}

// Clear empties the sample buffer.
func (s *SampleStream) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buffer = s.buffer[:0]
}

// BufferLength returns the current buffer length in bytes.
func (s *SampleStream) BufferLength() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.buffer)
}
