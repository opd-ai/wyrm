// Package subtitles provides text overlay rendering for dialog and audio.
package subtitles

import (
	"sync"
	"time"
)

// Position represents where subtitles appear on screen.
type Position int

const (
	PositionBottom Position = iota
	PositionTop
	PositionLeft
	PositionRight
	PositionCenter
)

// Alignment represents text alignment within the subtitle area.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// SubtitleStyle defines the visual appearance of subtitles.
type SubtitleStyle struct {
	// FontSize is the base font size multiplier (1.0 = normal)
	FontSize float64
	// BackgroundOpacity is 0.0 (transparent) to 1.0 (opaque)
	BackgroundOpacity float64
	// BackgroundColor is the RGBA color for the background
	BackgroundColor [4]uint8
	// TextColor is the RGBA color for the text
	TextColor [4]uint8
	// OutlineColor is the RGBA color for text outline
	OutlineColor [4]uint8
	// OutlineWidth is the outline thickness in pixels
	OutlineWidth float64
	// Padding is the space around the text in pixels
	Padding float64
	// MaxWidth is the maximum width as a fraction of screen width (0.0 to 1.0)
	MaxWidth float64
	// Position determines where on screen subtitles appear
	Position Position
	// Alignment determines text alignment
	Alignment Alignment
	// SpeakerColor enables per-speaker color coding
	SpeakerColors map[string][4]uint8
}

// DefaultStyle returns a sensible default subtitle style.
func DefaultStyle() SubtitleStyle {
	return SubtitleStyle{
		FontSize:          1.0,
		BackgroundOpacity: 0.7,
		BackgroundColor:   [4]uint8{0, 0, 0, 178},
		TextColor:         [4]uint8{255, 255, 255, 255},
		OutlineColor:      [4]uint8{0, 0, 0, 255},
		OutlineWidth:      1.0,
		Padding:           8.0,
		MaxWidth:          0.8,
		Position:          PositionBottom,
		Alignment:         AlignCenter,
		SpeakerColors:     make(map[string][4]uint8),
	}
}

// HighContrastStyle returns a style optimized for visibility.
func HighContrastStyle() SubtitleStyle {
	return SubtitleStyle{
		FontSize:          1.2,
		BackgroundOpacity: 0.9,
		BackgroundColor:   [4]uint8{0, 0, 0, 230},
		TextColor:         [4]uint8{255, 255, 0, 255}, // Yellow for better contrast
		OutlineColor:      [4]uint8{0, 0, 0, 255},
		OutlineWidth:      2.0,
		Padding:           12.0,
		MaxWidth:          0.9,
		Position:          PositionBottom,
		Alignment:         AlignCenter,
		SpeakerColors:     make(map[string][4]uint8),
	}
}

// Subtitle represents a single subtitle entry.
type Subtitle struct {
	ID        string
	Speaker   string // Who is speaking (for color coding)
	Text      string // The text to display
	Duration  time.Duration
	StartTime time.Time
	Priority  int  // Higher priority subtitles replace lower ones
	Important bool // Important subtitles stay visible longer
}

// IsExpired returns true if the subtitle has been displayed long enough.
func (s *Subtitle) IsExpired(now time.Time) bool {
	return now.After(s.StartTime.Add(s.Duration))
}

// RemainingTime returns how much time is left before expiry.
func (s *Subtitle) RemainingTime(now time.Time) time.Duration {
	endTime := s.StartTime.Add(s.Duration)
	if now.After(endTime) {
		return 0
	}
	return endTime.Sub(now)
}

// SubtitleQueue manages subtitle display with queuing and priorities.
type SubtitleQueue struct {
	mu              sync.RWMutex
	queue           []*Subtitle
	current         *Subtitle
	history         []*Subtitle
	maxHistory      int
	defaultDuration time.Duration
	minDisplayTime  time.Duration
}

// NewSubtitleQueue creates a new subtitle queue.
func NewSubtitleQueue() *SubtitleQueue {
	return &SubtitleQueue{
		queue:           make([]*Subtitle, 0),
		history:         make([]*Subtitle, 0),
		maxHistory:      100,
		defaultDuration: 4 * time.Second,
		minDisplayTime:  1 * time.Second,
	}
}

// SetDefaultDuration sets the default display duration for subtitles.
func (sq *SubtitleQueue) SetDefaultDuration(d time.Duration) {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	sq.defaultDuration = d
}

// SetMinDisplayTime sets the minimum time a subtitle must be displayed.
func (sq *SubtitleQueue) SetMinDisplayTime(d time.Duration) {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	sq.minDisplayTime = d
}

// Add adds a subtitle to the queue.
func (sq *SubtitleQueue) Add(sub *Subtitle) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	// Set defaults if not specified
	if sub.Duration == 0 {
		sub.Duration = sq.defaultDuration
	}
	if sub.Duration < sq.minDisplayTime {
		sub.Duration = sq.minDisplayTime
	}

	// Insert in priority order (higher priority first)
	inserted := false
	for i, existing := range sq.queue {
		if sub.Priority > existing.Priority {
			sq.queue = append(sq.queue[:i], append([]*Subtitle{sub}, sq.queue[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		sq.queue = append(sq.queue, sub)
	}
}

// AddText is a convenience method to add a simple text subtitle.
func (sq *SubtitleQueue) AddText(text string) {
	sq.Add(&Subtitle{
		Text:      text,
		StartTime: time.Now(),
	})
}

// AddDialog adds a dialog subtitle with speaker name.
func (sq *SubtitleQueue) AddDialog(speaker, text string) {
	sq.Add(&Subtitle{
		Speaker:   speaker,
		Text:      text,
		StartTime: time.Now(),
	})
}

// AddImportant adds an important subtitle that displays longer.
func (sq *SubtitleQueue) AddImportant(text string) {
	sq.Add(&Subtitle{
		Text:      text,
		Duration:  8 * time.Second,
		StartTime: time.Now(),
		Priority:  10,
		Important: true,
	})
}

// Update advances the subtitle queue, called each frame.
func (sq *SubtitleQueue) Update() {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	now := time.Now()

	// Check if current subtitle has expired
	if sq.current != nil && sq.current.IsExpired(now) {
		sq.archiveSubtitle(sq.current)
		sq.current = nil
	}

	// If no current subtitle and queue has items, display next
	if sq.current == nil && len(sq.queue) > 0 {
		sq.current = sq.queue[0]
		sq.current.StartTime = now
		sq.queue = sq.queue[1:]
	}
}

// archiveSubtitle moves a subtitle to history.
func (sq *SubtitleQueue) archiveSubtitle(sub *Subtitle) {
	sq.history = append(sq.history, sub)
	if len(sq.history) > sq.maxHistory {
		sq.history = sq.history[1:]
	}
}

// Current returns the currently displayed subtitle, or nil if none.
func (sq *SubtitleQueue) Current() *Subtitle {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	return sq.current
}

// QueueLength returns the number of queued subtitles.
func (sq *SubtitleQueue) QueueLength() int {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	return len(sq.queue)
}

// Clear removes all queued subtitles and the current subtitle.
func (sq *SubtitleQueue) Clear() {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	sq.queue = sq.queue[:0]
	sq.current = nil
}

// Skip immediately expires the current subtitle.
func (sq *SubtitleQueue) Skip() {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	if sq.current != nil {
		sq.archiveSubtitle(sq.current)
		sq.current = nil
	}
}

// GetHistory returns recent subtitle history.
func (sq *SubtitleQueue) GetHistory(limit int) []*Subtitle {
	sq.mu.RLock()
	defer sq.mu.RUnlock()

	if limit <= 0 || limit > len(sq.history) {
		limit = len(sq.history)
	}
	start := len(sq.history) - limit
	result := make([]*Subtitle, limit)
	copy(result, sq.history[start:])
	return result
}

// SubtitleSystem manages subtitle rendering and display logic.
type SubtitleSystem struct {
	mu      sync.RWMutex
	queue   *SubtitleQueue
	style   SubtitleStyle
	enabled bool
}

// NewSubtitleSystem creates a new subtitle system.
func NewSubtitleSystem(enabled bool) *SubtitleSystem {
	return &SubtitleSystem{
		queue:   NewSubtitleQueue(),
		style:   DefaultStyle(),
		enabled: enabled,
	}
}

// SetEnabled enables or disables subtitle display.
func (ss *SubtitleSystem) SetEnabled(enabled bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.enabled = enabled
}

// IsEnabled returns whether subtitles are enabled.
func (ss *SubtitleSystem) IsEnabled() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.enabled
}

// SetStyle sets the subtitle style.
func (ss *SubtitleSystem) SetStyle(style SubtitleStyle) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.style = style
}

// GetStyle returns the current subtitle style.
func (ss *SubtitleSystem) GetStyle() SubtitleStyle {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.style
}

// SetFontSize sets the font size multiplier.
func (ss *SubtitleSystem) SetFontSize(size float64) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.style.FontSize = size
}

// SetPosition sets where subtitles appear on screen.
func (ss *SubtitleSystem) SetPosition(pos Position) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.style.Position = pos
}

// SetBackgroundOpacity sets the background opacity.
func (ss *SubtitleSystem) SetBackgroundOpacity(opacity float64) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.style.BackgroundOpacity = opacity
	ss.style.BackgroundColor[3] = uint8(opacity * 255)
}

// AddSpeakerColor sets a color for a specific speaker.
func (ss *SubtitleSystem) AddSpeakerColor(speaker string, color [4]uint8) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.style.SpeakerColors == nil {
		ss.style.SpeakerColors = make(map[string][4]uint8)
	}
	ss.style.SpeakerColors[speaker] = color
}

// Update advances the subtitle system.
func (ss *SubtitleSystem) Update() {
	ss.queue.Update()
}

// AddSubtitle adds a subtitle through the system.
func (ss *SubtitleSystem) AddSubtitle(sub *Subtitle) {
	if !ss.IsEnabled() {
		return
	}
	ss.queue.Add(sub)
}

// AddText adds a simple text subtitle.
func (ss *SubtitleSystem) AddText(text string) {
	if !ss.IsEnabled() {
		return
	}
	ss.queue.AddText(text)
}

// AddDialog adds a dialog subtitle.
func (ss *SubtitleSystem) AddDialog(speaker, text string) {
	if !ss.IsEnabled() {
		return
	}
	ss.queue.AddDialog(speaker, text)
}

// AddImportant adds an important subtitle.
func (ss *SubtitleSystem) AddImportant(text string) {
	if !ss.IsEnabled() {
		return
	}
	ss.queue.AddImportant(text)
}

// Current returns the current subtitle to render.
func (ss *SubtitleSystem) Current() *Subtitle {
	if !ss.IsEnabled() {
		return nil
	}
	return ss.queue.Current()
}

// Clear clears all subtitles.
func (ss *SubtitleSystem) Clear() {
	ss.queue.Clear()
}

// Skip skips the current subtitle.
func (ss *SubtitleSystem) Skip() {
	ss.queue.Skip()
}

// GetRenderData returns data needed to render the current subtitle.
// Returns nil if there's nothing to render.
func (ss *SubtitleSystem) GetRenderData() *SubtitleRenderData {
	if !ss.IsEnabled() {
		return nil
	}

	sub := ss.queue.Current()
	if sub == nil {
		return nil
	}

	ss.mu.RLock()
	style := ss.style
	ss.mu.RUnlock()

	// Get speaker color if available
	textColor := style.TextColor
	if sub.Speaker != "" {
		if speakerColor, ok := style.SpeakerColors[sub.Speaker]; ok {
			textColor = speakerColor
		}
	}

	return &SubtitleRenderData{
		Text:              sub.Text,
		Speaker:           sub.Speaker,
		FontSize:          style.FontSize,
		TextColor:         textColor,
		BackgroundColor:   style.BackgroundColor,
		OutlineColor:      style.OutlineColor,
		OutlineWidth:      style.OutlineWidth,
		Padding:           style.Padding,
		MaxWidth:          style.MaxWidth,
		Position:          style.Position,
		Alignment:         style.Alignment,
		RemainingFraction: float64(sub.RemainingTime(time.Now())) / float64(sub.Duration),
	}
}

// SubtitleRenderData contains all data needed to render a subtitle.
type SubtitleRenderData struct {
	Text              string
	Speaker           string
	FontSize          float64
	TextColor         [4]uint8
	BackgroundColor   [4]uint8
	OutlineColor      [4]uint8
	OutlineWidth      float64
	Padding           float64
	MaxWidth          float64
	Position          Position
	Alignment         Alignment
	RemainingFraction float64 // 1.0 = just started, 0.0 = about to expire
}

// FormatText returns the display text with optional speaker prefix.
func (rd *SubtitleRenderData) FormatText() string {
	if rd.Speaker != "" {
		return rd.Speaker + ": " + rd.Text
	}
	return rd.Text
}
