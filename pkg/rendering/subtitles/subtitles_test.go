package subtitles

import (
	"sync"
	"testing"
	"time"
)

func TestDefaultStyle(t *testing.T) {
	style := DefaultStyle()
	if style.FontSize != 1.0 {
		t.Errorf("expected FontSize 1.0, got %f", style.FontSize)
	}
	if style.Position != PositionBottom {
		t.Errorf("expected PositionBottom, got %v", style.Position)
	}
	if style.Alignment != AlignCenter {
		t.Errorf("expected AlignCenter, got %v", style.Alignment)
	}
}

func TestHighContrastStyle(t *testing.T) {
	style := HighContrastStyle()
	if style.FontSize <= 1.0 {
		t.Error("high contrast style should have larger font")
	}
	if style.BackgroundOpacity < 0.8 {
		t.Error("high contrast style should have higher opacity")
	}
}

func TestSubtitleExpiry(t *testing.T) {
	now := time.Now()
	sub := &Subtitle{
		Text:      "Test",
		Duration:  2 * time.Second,
		StartTime: now,
	}

	// Should not be expired immediately
	if sub.IsExpired(now) {
		t.Error("subtitle should not be expired immediately")
	}

	// Should not be expired after 1 second
	if sub.IsExpired(now.Add(1 * time.Second)) {
		t.Error("subtitle should not be expired after 1 second")
	}

	// Should be expired after 3 seconds
	if !sub.IsExpired(now.Add(3 * time.Second)) {
		t.Error("subtitle should be expired after 3 seconds")
	}
}

func TestSubtitleRemainingTime(t *testing.T) {
	now := time.Now()
	sub := &Subtitle{
		Text:      "Test",
		Duration:  4 * time.Second,
		StartTime: now,
	}

	remaining := sub.RemainingTime(now.Add(1 * time.Second))
	if remaining < 2*time.Second || remaining > 4*time.Second {
		t.Errorf("expected ~3 seconds remaining, got %v", remaining)
	}

	remaining = sub.RemainingTime(now.Add(5 * time.Second))
	if remaining != 0 {
		t.Errorf("expected 0 remaining after expiry, got %v", remaining)
	}
}

func TestSubtitleQueue(t *testing.T) {
	sq := NewSubtitleQueue()

	if sq.QueueLength() != 0 {
		t.Error("new queue should be empty")
	}

	sq.AddText("First subtitle")
	if sq.QueueLength() != 1 {
		t.Errorf("expected 1 in queue, got %d", sq.QueueLength())
	}

	sq.AddText("Second subtitle")
	if sq.QueueLength() != 2 {
		t.Errorf("expected 2 in queue, got %d", sq.QueueLength())
	}
}

func TestSubtitleQueueUpdate(t *testing.T) {
	sq := NewSubtitleQueue()
	sq.SetDefaultDuration(100 * time.Millisecond)
	sq.SetMinDisplayTime(50 * time.Millisecond)

	sq.AddText("Test")

	// First update should move subtitle to current
	sq.Update()

	if sq.Current() == nil {
		t.Error("expected current subtitle after update")
	}
	if sq.Current().Text != "Test" {
		t.Errorf("expected 'Test', got '%s'", sq.Current().Text)
	}
	if sq.QueueLength() != 0 {
		t.Error("queue should be empty after update")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)
	sq.Update()

	if sq.Current() != nil {
		t.Error("current should be nil after expiry")
	}
}

func TestSubtitleQueuePriority(t *testing.T) {
	sq := NewSubtitleQueue()

	sq.Add(&Subtitle{Text: "Low priority", Priority: 1})
	sq.Add(&Subtitle{Text: "High priority", Priority: 10})
	sq.Add(&Subtitle{Text: "Medium priority", Priority: 5})

	// High priority should be first
	sq.Update()
	if sq.Current().Text != "High priority" {
		t.Errorf("expected 'High priority', got '%s'", sq.Current().Text)
	}

	// Skip to next
	sq.Skip()
	sq.Update()
	if sq.Current().Text != "Medium priority" {
		t.Errorf("expected 'Medium priority', got '%s'", sq.Current().Text)
	}
}

func TestSubtitleQueueClear(t *testing.T) {
	sq := NewSubtitleQueue()

	sq.AddText("One")
	sq.AddText("Two")
	sq.Update()

	sq.Clear()

	if sq.Current() != nil {
		t.Error("current should be nil after clear")
	}
	if sq.QueueLength() != 0 {
		t.Error("queue should be empty after clear")
	}
}

func TestSubtitleQueueSkip(t *testing.T) {
	sq := NewSubtitleQueue()

	sq.AddText("First")
	sq.AddText("Second")
	sq.Update()

	if sq.Current().Text != "First" {
		t.Errorf("expected 'First', got '%s'", sq.Current().Text)
	}

	sq.Skip()
	sq.Update()

	if sq.Current().Text != "Second" {
		t.Errorf("expected 'Second', got '%s'", sq.Current().Text)
	}
}

func TestSubtitleQueueHistory(t *testing.T) {
	sq := NewSubtitleQueue()
	sq.SetDefaultDuration(10 * time.Millisecond)
	sq.SetMinDisplayTime(5 * time.Millisecond)

	sq.AddText("First")
	sq.Update() // Start displaying "First"
	time.Sleep(20 * time.Millisecond)
	sq.Update() // "First" expires, goes to history

	sq.AddText("Second")
	sq.Update() // Start displaying "Second"
	time.Sleep(20 * time.Millisecond)
	sq.Update() // "Second" expires, goes to history

	history := sq.GetHistory(10)
	if len(history) != 2 {
		t.Errorf("expected 2 in history, got %d", len(history))
		return
	}
	if history[0].Text != "First" {
		t.Errorf("expected 'First' first in history, got '%s'", history[0].Text)
	}
}

func TestSubtitleSystem(t *testing.T) {
	ss := NewSubtitleSystem(true)

	if !ss.IsEnabled() {
		t.Error("system should be enabled")
	}

	ss.AddText("Test subtitle")
	ss.Update()

	current := ss.Current()
	if current == nil {
		t.Error("expected current subtitle")
	}
	if current.Text != "Test subtitle" {
		t.Errorf("expected 'Test subtitle', got '%s'", current.Text)
	}
}

func TestSubtitleSystemDisabled(t *testing.T) {
	ss := NewSubtitleSystem(false)

	if ss.IsEnabled() {
		t.Error("system should be disabled")
	}

	ss.AddText("Test")
	ss.Update()

	if ss.Current() != nil {
		t.Error("disabled system should not display subtitles")
	}
}

func TestSubtitleSystemStyle(t *testing.T) {
	ss := NewSubtitleSystem(true)

	ss.SetFontSize(1.5)
	ss.SetPosition(PositionTop)
	ss.SetBackgroundOpacity(0.5)

	style := ss.GetStyle()
	if style.FontSize != 1.5 {
		t.Errorf("expected FontSize 1.5, got %f", style.FontSize)
	}
	if style.Position != PositionTop {
		t.Errorf("expected PositionTop, got %v", style.Position)
	}
	if style.BackgroundOpacity != 0.5 {
		t.Errorf("expected BackgroundOpacity 0.5, got %f", style.BackgroundOpacity)
	}
}

func TestSubtitleSystemSpeakerColors(t *testing.T) {
	ss := NewSubtitleSystem(true)

	redColor := [4]uint8{255, 0, 0, 255}
	ss.AddSpeakerColor("Hero", redColor)

	ss.AddDialog("Hero", "Hello!")
	ss.Update()

	rd := ss.GetRenderData()
	if rd == nil {
		t.Fatal("expected render data")
	}
	if rd.TextColor != redColor {
		t.Error("expected hero's color to be applied")
	}
}

func TestSubtitleRenderData(t *testing.T) {
	ss := NewSubtitleSystem(true)

	ss.AddDialog("NPC", "Welcome traveler!")
	ss.Update()

	rd := ss.GetRenderData()
	if rd == nil {
		t.Fatal("expected render data")
	}

	formatted := rd.FormatText()
	if formatted != "NPC: Welcome traveler!" {
		t.Errorf("expected 'NPC: Welcome traveler!', got '%s'", formatted)
	}
}

func TestSubtitleRenderDataNoSpeaker(t *testing.T) {
	ss := NewSubtitleSystem(true)

	ss.AddText("System message")
	ss.Update()

	rd := ss.GetRenderData()
	if rd == nil {
		t.Fatal("expected render data")
	}

	formatted := rd.FormatText()
	if formatted != "System message" {
		t.Errorf("expected 'System message', got '%s'", formatted)
	}
}

func TestConcurrentAccess(t *testing.T) {
	ss := NewSubtitleSystem(true)

	var wg sync.WaitGroup

	// Multiple goroutines adding subtitles
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				ss.AddText("Test")
				ss.Update()
				_ = ss.Current()
				_ = ss.GetRenderData()
			}
		}()
	}

	// Goroutine modifying style
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 50; j++ {
			ss.SetFontSize(1.0 + float64(j)*0.01)
			ss.SetPosition(Position(j % 5))
			ss.SetEnabled(j%2 == 0)
		}
	}()

	wg.Wait()
}

func TestAddImportant(t *testing.T) {
	sq := NewSubtitleQueue()

	sq.AddImportant("Important message!")
	sq.Update()

	current := sq.Current()
	if current == nil {
		t.Fatal("expected current subtitle")
	}
	if !current.Important {
		t.Error("expected subtitle to be marked important")
	}
	if current.Priority < 5 {
		t.Error("expected high priority for important subtitle")
	}
}

func TestSubtitleSystemAddDialog(t *testing.T) {
	ss := NewSubtitleSystem(true)

	ss.AddDialog("Guard", "Halt!")
	ss.Update()

	current := ss.Current()
	if current == nil {
		t.Fatal("expected current subtitle")
	}
	if current.Speaker != "Guard" {
		t.Errorf("expected speaker 'Guard', got '%s'", current.Speaker)
	}
	if current.Text != "Halt!" {
		t.Errorf("expected text 'Halt!', got '%s'", current.Text)
	}
}

func TestMinDisplayTime(t *testing.T) {
	sq := NewSubtitleQueue()
	sq.SetMinDisplayTime(1 * time.Second)

	// Add subtitle with very short duration
	sq.Add(&Subtitle{
		Text:     "Short",
		Duration: 10 * time.Millisecond,
	})
	sq.Update()

	current := sq.Current()
	if current == nil {
		t.Fatal("expected current subtitle")
	}
	if current.Duration < 1*time.Second {
		t.Errorf("duration should be at least min display time, got %v", current.Duration)
	}
}
