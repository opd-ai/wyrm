//go:build !noebiten

// Package audio provides procedural audio synthesis.
package audio

import (
	"io"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	// AudioSampleRate is the sample rate for audio playback.
	AudioSampleRate = 44100
	// BytesPerSample is 4 bytes for 16-bit stereo.
	BytesPerSample = 4
)

// Player handles Ebitengine audio playback from synthesized samples.
type Player struct {
	context   *audio.Context
	player    *audio.Player
	stream    *SampleStream
	mu        sync.Mutex
	isPlaying bool
}

// NewPlayer creates a new Ebitengine audio player.
func NewPlayer() (*Player, error) {
	ctx := audio.NewContext(AudioSampleRate)
	stream := NewSampleStream()

	player, err := ctx.NewPlayer(stream)
	if err != nil {
		return nil, err
	}

	return &Player{
		context: ctx,
		player:  player,
		stream:  stream,
	}, nil
}

// QueueSamples adds synthesized samples to the playback queue.
// Samples should be float64 values in range [-1, 1].
func (p *Player) QueueSamples(samples []float64) {
	p.stream.QueueSamples(samples)
}

// Play starts audio playback.
func (p *Player) Play() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isPlaying {
		p.player.Play()
		p.isPlaying = true
	}
}

// Pause pauses audio playback.
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isPlaying {
		p.player.Pause()
		p.isPlaying = false
	}
}

// IsPlaying returns whether audio is currently playing.
func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isPlaying
}

// Close releases the audio player resources.
func (p *Player) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.player != nil {
		p.player.Close()
		p.player = nil
	}
	p.isPlaying = false
	return nil
}

// SampleStream implements io.Reader for Ebitengine audio playback.
type SampleStream struct {
	mu      sync.Mutex
	buffer  []byte
	silence []byte
}

// NewSampleStream creates a new sample stream.
func NewSampleStream() *SampleStream {
	// Pre-allocate silence buffer for when no samples are available
	silence := make([]byte, 4096)
	return &SampleStream{
		buffer:  make([]byte, 0, AudioSampleRate*BytesPerSample),
		silence: silence,
	}
}

// QueueSamples converts float64 samples to 16-bit PCM stereo and queues them.
func (s *SampleStream) QueueSamples(samples []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sample := range samples {
		// Clamp sample to [-1, 1]
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}

		// Convert to 16-bit signed integer
		i16 := int16(sample * 32767)
		lo := byte(i16)
		hi := byte(i16 >> 8)

		// Stereo: same sample for left and right channels
		s.buffer = append(s.buffer, lo, hi, lo, hi)
	}
}

// Read implements io.Reader for Ebitengine audio context.
func (s *SampleStream) Read(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.buffer) == 0 {
		// Return silence when no samples are queued
		n = copy(p, s.silence[:min(len(p), len(s.silence))])
		return n, nil
	}

	n = copy(p, s.buffer)
	s.buffer = s.buffer[n:]
	return n, nil
}

// Seek implements io.Seeker (required by Ebitengine but not used for streaming).
func (s *SampleStream) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

// Length returns the length of the stream (required by Ebitengine).
func (s *SampleStream) Length() int64 {
	// Return a large value for streaming audio
	return 1<<62 - 1
}

// Verify SampleStream implements the required interface.
var _ io.ReadSeeker = (*SampleStream)(nil)

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
