package raycast

import (
	"image/color"
	"math"
	"testing"
)

func TestDefaultEdgeHighlightConfig(t *testing.T) {
	config := DefaultEdgeHighlightConfig()

	if !config.Enabled {
		t.Error("expected Enabled to be true by default")
	}
	if config.Genre != "fantasy" {
		t.Errorf("expected Genre 'fantasy', got '%s'", config.Genre)
	}
	if config.PulseSpeed != 3.0 {
		t.Errorf("expected PulseSpeed 3.0, got %f", config.PulseSpeed)
	}
	if config.PulseAmplitude != 0.3 {
		t.Errorf("expected PulseAmplitude 0.3, got %f", config.PulseAmplitude)
	}
	if config.BaseIntensity != 0.7 {
		t.Errorf("expected BaseIntensity 0.7, got %f", config.BaseIntensity)
	}
}

func TestGetAccentColor(t *testing.T) {
	tests := []struct {
		genre    string
		expected color.RGBA
	}{
		{"fantasy", color.RGBA{R: 255, G: 215, B: 0, A: 255}},
		{"sci-fi", color.RGBA{R: 0, G: 255, B: 255, A: 255}},
		{"horror", color.RGBA{R: 220, G: 20, B: 60, A: 255}},
		{"cyberpunk", color.RGBA{R: 255, G: 20, B: 147, A: 255}},
		{"post-apocalyptic", color.RGBA{R: 255, G: 140, B: 0, A: 255}},
		{"post-apoc", color.RGBA{R: 255, G: 140, B: 0, A: 255}},
		{"unknown", color.RGBA{R: 255, G: 215, B: 0, A: 255}}, // Falls back to fantasy
	}

	for _, tt := range tests {
		c := GetAccentColor(tt.genre)
		if c != tt.expected {
			t.Errorf("GetAccentColor(%s): expected %v, got %v", tt.genre, tt.expected, c)
		}
	}
}

func TestNewEdgeHighlightRenderer(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(320, 200, config)

	if hr.width != 320 {
		t.Errorf("expected width 320, got %d", hr.width)
	}
	if hr.height != 200 {
		t.Errorf("expected height 200, got %d", hr.height)
	}
	if hr.config.Genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got '%s'", hr.config.Genre)
	}
}

func TestSetFramebuffer(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(10, 10, config)

	fb := make([]byte, 10*10*4)
	hr.SetFramebuffer(fb)

	if hr.framebuffer == nil {
		t.Error("framebuffer should be set")
	}
	if len(hr.framebuffer) != len(fb) {
		t.Errorf("framebuffer length mismatch: expected %d, got %d", len(fb), len(hr.framebuffer))
	}
}

func TestSetTime(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(10, 10, config)

	hr.SetTime(1.5)
	if hr.config.Time != 1.5 {
		t.Errorf("expected time 1.5, got %f", hr.config.Time)
	}
}

func TestEdgeHighlightRenderer_SetGenre(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(10, 10, config)

	hr.SetGenre("cyberpunk")
	if hr.config.Genre != "cyberpunk" {
		t.Errorf("expected genre 'cyberpunk', got '%s'", hr.config.Genre)
	}
}

func TestCalculatePulseIntensity(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(10, 10, config)

	// At time 0, sin(0) = 0, so intensity = 0*0.3 + 0.7 = 0.7
	hr.SetTime(0)
	intensity := hr.CalculatePulseIntensity()
	if math.Abs(intensity-0.7) > 0.001 {
		t.Errorf("expected intensity ~0.7 at time 0, got %f", intensity)
	}

	// At time π/6 (≈0.5236), sin(π/6*3) = sin(π/2) = 1, so intensity = 1*0.3 + 0.7 = 1.0
	hr.SetTime(math.Pi / 6)
	intensity = hr.CalculatePulseIntensity()
	if math.Abs(intensity-1.0) > 0.001 {
		t.Errorf("expected intensity ~1.0 at time π/6, got %f", intensity)
	}

	// At time π/3 (≈1.047), sin(π/3*3) = sin(π) = 0, so intensity = 0.7
	hr.SetTime(math.Pi / 3)
	intensity = hr.CalculatePulseIntensity()
	if math.Abs(intensity-0.7) > 0.001 {
		t.Errorf("expected intensity ~0.7 at time π/3, got %f", intensity)
	}
}

func TestApplyHighlightDisabled(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	config.Enabled = false
	hr := NewEdgeHighlightRenderer(10, 10, config)

	fb := make([]byte, 10*10*4)
	hr.SetFramebuffer(fb)

	region := HighlightRegion{
		MinX:           2,
		MaxX:           7,
		MinY:           2,
		MaxY:           7,
		HighlightState: 2,
	}

	// Should not modify framebuffer when disabled
	hr.ApplyHighlight(region)

	// Check that framebuffer is unchanged
	for i := range fb {
		if fb[i] != 0 {
			t.Error("framebuffer should not be modified when highlighting is disabled")
			break
		}
	}
}

func TestApplyHighlightNoFramebuffer(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(10, 10, config)
	// Don't set framebuffer

	region := HighlightRegion{
		MinX:           2,
		MaxX:           7,
		MinY:           2,
		MaxY:           7,
		HighlightState: 2,
	}

	// Should not panic with nil framebuffer
	hr.ApplyHighlight(region)
}

func TestApplyHighlightZeroState(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(10, 10, config)

	fb := make([]byte, 10*10*4)
	hr.SetFramebuffer(fb)

	region := HighlightRegion{
		MinX:           2,
		MaxX:           7,
		MinY:           2,
		MaxY:           7,
		HighlightState: 0, // No highlight
	}

	hr.ApplyHighlight(region)

	// Check that framebuffer is unchanged
	for i := range fb {
		if fb[i] != 0 {
			t.Error("framebuffer should not be modified when HighlightState is 0")
			break
		}
	}
}

func TestApplyHighlightEdgeDetection(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	config.Genre = "fantasy" // Gold color
	hr := NewEdgeHighlightRenderer(10, 10, config)

	fb := make([]byte, 10*10*4)
	// Create a small solid square with alpha
	// Set pixels (3,3), (4,3), (3,4), (4,4) to have alpha
	for y := 3; y <= 4; y++ {
		for x := 3; x <= 4; x++ {
			idx := (y*10 + x) * 4
			fb[idx] = 100   // R
			fb[idx+1] = 100 // G
			fb[idx+2] = 100 // B
			fb[idx+3] = 255 // A (visible)
		}
	}

	hr.SetFramebuffer(fb)

	region := HighlightRegion{
		MinX:           2,
		MaxX:           6,
		MinY:           2,
		MaxY:           6,
		HighlightState: 2,
	}

	hr.ApplyHighlight(region)

	// The edge pixels (3,3), (4,3), (3,4), (4,4) should be highlighted
	// because they all have at least one transparent neighbor
	edgeModified := false
	for y := 3; y <= 4; y++ {
		for x := 3; x <= 4; x++ {
			idx := (y*10 + x) * 4
			// Gold highlight should increase R and G values
			if fb[idx] > 100 || fb[idx+1] > 100 {
				edgeModified = true
				break
			}
		}
		if edgeModified {
			break
		}
	}

	if !edgeModified {
		t.Error("edge pixels should be modified with highlight color")
	}
}

func TestIsEdgePixel(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(10, 10, config)

	fb := make([]byte, 10*10*4)
	// Create a 3x3 solid block at (3,3) to (5,5)
	for y := 3; y <= 5; y++ {
		for x := 3; x <= 5; x++ {
			idx := (y*10 + x) * 4
			fb[idx+3] = 255 // Set alpha
		}
	}
	hr.SetFramebuffer(fb)

	// Center pixel (4,4) should NOT be an edge (all neighbors have alpha)
	if hr.isEdgePixel(4, 4, 0, 9, 0, 9) {
		t.Error("center pixel (4,4) should not be an edge pixel")
	}

	// Corner pixel (3,3) should be an edge (has transparent neighbors)
	if !hr.isEdgePixel(3, 3, 0, 9, 0, 9) {
		t.Error("corner pixel (3,3) should be an edge pixel")
	}

	// Edge pixel (4,3) should be an edge (top neighbor is transparent)
	if !hr.isEdgePixel(4, 3, 0, 9, 0, 9) {
		t.Error("top edge pixel (4,3) should be an edge pixel")
	}
}

func TestBlendHighlightColor(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(10, 10, config)

	fb := make([]byte, 4)
	fb[0] = 100 // R
	fb[1] = 100 // G
	fb[2] = 100 // B
	fb[3] = 255 // A
	hr.framebuffer = fb

	// Blend with gold (255, 215, 0) at 50% intensity
	hr.blendHighlightColor(0, color.RGBA{R: 255, G: 215, B: 0, A: 255}, 0.5)

	// R = min(255, 100 + 255*0.5) = min(255, 227.5) = 227
	// G = min(255, 100 + 215*0.5) = min(255, 207.5) = 207
	// B = min(255, 100 + 0*0.5) = 100
	expectedR := uint8(227)
	expectedG := uint8(207)
	expectedB := uint8(100)

	if fb[0] != expectedR {
		t.Errorf("expected R=%d, got %d", expectedR, fb[0])
	}
	if fb[1] != expectedG {
		t.Errorf("expected G=%d, got %d", expectedG, fb[1])
	}
	if fb[2] != expectedB {
		t.Errorf("expected B=%d, got %d", expectedB, fb[2])
	}
}

func TestCreateRegionFromScreenBounds(t *testing.T) {
	region := CreateRegionFromScreenBounds(42, 10, 20, 50, 30, 2)

	if region.EntityID != 42 {
		t.Errorf("expected EntityID 42, got %d", region.EntityID)
	}
	if region.MinX != 10 {
		t.Errorf("expected MinX 10, got %d", region.MinX)
	}
	if region.MaxX != 59 { // 10 + 50 - 1
		t.Errorf("expected MaxX 59, got %d", region.MaxX)
	}
	if region.MinY != 20 {
		t.Errorf("expected MinY 20, got %d", region.MinY)
	}
	if region.MaxY != 49 { // 20 + 30 - 1
		t.Errorf("expected MaxY 49, got %d", region.MaxY)
	}
	if region.HighlightState != 2 {
		t.Errorf("expected HighlightState 2, got %d", region.HighlightState)
	}
}

func TestApplyHighlightToRegions(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(20, 20, config)

	fb := make([]byte, 20*20*4)
	// Create two separate sprites
	// Sprite 1: pixel at (3,3)
	fb[(3*20+3)*4+3] = 255
	fb[(3*20+3)*4] = 100
	fb[(3*20+3)*4+1] = 100
	fb[(3*20+3)*4+2] = 100

	// Sprite 2: pixel at (15,15)
	fb[(15*20+15)*4+3] = 255
	fb[(15*20+15)*4] = 100
	fb[(15*20+15)*4+1] = 100
	fb[(15*20+15)*4+2] = 100

	hr.SetFramebuffer(fb)

	regions := []HighlightRegion{
		{MinX: 2, MaxX: 5, MinY: 2, MaxY: 5, HighlightState: 1},
		{MinX: 14, MaxX: 17, MinY: 14, MaxY: 17, HighlightState: 2},
	}

	hr.ApplyHighlightToRegions(regions)

	// Both pixels should be modified
	if fb[(3*20+3)*4] <= 100 {
		t.Error("first sprite should be highlighted")
	}
	if fb[(15*20+15)*4] <= 100 {
		t.Error("second sprite should be highlighted")
	}
}

func TestHighlightIntensityByState(t *testing.T) {
	config := DefaultEdgeHighlightConfig()
	// Use lower intensity to avoid saturation
	config.BaseIntensity = 0.3
	config.PulseAmplitude = 0.1
	hr := NewEdgeHighlightRenderer(10, 10, config)

	// Create two framebuffers to compare
	fb1 := make([]byte, 10*10*4)
	fb2 := make([]byte, 10*10*4)

	// Set up identical single pixels with low base value to avoid saturation
	fb1[(5*10+5)*4+3] = 255
	fb1[(5*10+5)*4] = 50
	fb1[(5*10+5)*4+1] = 50
	fb1[(5*10+5)*4+2] = 50

	fb2[(5*10+5)*4+3] = 255
	fb2[(5*10+5)*4] = 50
	fb2[(5*10+5)*4+1] = 50
	fb2[(5*10+5)*4+2] = 50

	// Apply state 1 highlight
	hr.SetFramebuffer(fb1)
	hr.ApplyHighlight(HighlightRegion{MinX: 4, MaxX: 6, MinY: 4, MaxY: 6, HighlightState: 1})

	// Apply state 2 highlight (targeted - should be brighter)
	hr.SetFramebuffer(fb2)
	hr.ApplyHighlight(HighlightRegion{MinX: 4, MaxX: 6, MinY: 4, MaxY: 6, HighlightState: 2})

	// State 2 should have higher values due to 1.3x intensity multiplier
	r1 := fb1[(5*10+5)*4]
	r2 := fb2[(5*10+5)*4]

	// Both should be modified (higher than 50)
	if r1 <= 50 {
		t.Errorf("state 1 highlight should modify pixel: R=%d", r1)
	}
	if r2 <= 50 {
		t.Errorf("state 2 highlight should modify pixel: R=%d", r2)
	}
	// State 2 should be brighter (unless both saturate, which they shouldn't at these intensities)
	if r2 < r1 {
		t.Errorf("targeted highlight (state 2) should be >= state 1: state1 R=%d, state2 R=%d", r1, r2)
	}
}

func BenchmarkApplyHighlight(b *testing.B) {
	config := DefaultEdgeHighlightConfig()
	hr := NewEdgeHighlightRenderer(320, 200, config)

	fb := make([]byte, 320*200*4)
	// Create a sprite region with alpha
	for y := 50; y < 150; y++ {
		for x := 100; x < 220; x++ {
			idx := (y*320 + x) * 4
			fb[idx] = 128
			fb[idx+1] = 128
			fb[idx+2] = 128
			fb[idx+3] = 255
		}
	}
	hr.SetFramebuffer(fb)

	region := HighlightRegion{
		MinX:           100,
		MaxX:           220,
		MinY:           50,
		MaxY:           150,
		HighlightState: 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hr.ApplyHighlight(region)
	}
}
