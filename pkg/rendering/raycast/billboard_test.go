//go:build noebiten

package raycast

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/rendering/sprite"
)

func TestTransformEntityToScreen(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0 // Facing right (+X direction)

	tests := []struct {
		name         string
		entityX      float64
		entityY      float64
		wantOnScreen bool
	}{
		{
			name:         "entity in front",
			entityX:      8.0,
			entityY:      5.0,
			wantOnScreen: true,
		},
		{
			name:         "entity behind camera",
			entityX:      2.0,
			entityY:      5.0,
			wantOnScreen: false, // Behind camera
		},
		{
			name:         "entity to the right",
			entityX:      8.0,
			entityY:      7.0,
			wantOnScreen: true,
		},
		{
			name:         "entity to the left",
			entityX:      8.0,
			entityY:      3.0,
			wantOnScreen: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &SpriteEntity{
				X:       tt.entityX,
				Y:       tt.entityY,
				Visible: true,
			}
			got := r.TransformEntityToScreen(e)
			if got != tt.wantOnScreen {
				t.Errorf("TransformEntityToScreen() = %v, want %v", got, tt.wantOnScreen)
			}
		})
	}
}

func TestTransformEntityToScreenInvisible(t *testing.T) {
	r := NewRenderer(640, 480)
	e := &SpriteEntity{
		X:       10.0,
		Y:       5.0,
		Visible: false,
	}
	if r.TransformEntityToScreen(e) {
		t.Error("invisible entity should return false")
	}
}

func TestGetSpriteScreenBounds(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	e := &SpriteEntity{
		X:       8.0, // 3 units away
		Y:       5.0,
		Scale:   1.0,
		Visible: true,
	}
	r.TransformEntityToScreen(e)

	startX, endX, startY, endY, spriteW, spriteH := r.GetSpriteScreenBounds(e, 32, 48)

	// Sprite should be centered around screen center
	if startX >= endX {
		t.Errorf("startX %d >= endX %d", startX, endX)
	}
	if startY >= endY {
		t.Errorf("startY %d >= endY %d", startY, endY)
	}
	if spriteW <= 0 {
		t.Errorf("spriteW %d <= 0", spriteW)
	}
	if spriteH <= 0 {
		t.Errorf("spriteH %d <= 0", spriteH)
	}
}

func TestSortSpritesByDistance(t *testing.T) {
	entities := []*SpriteEntity{
		{Distance: 5.0},
		{Distance: 10.0},
		{Distance: 2.0},
		{Distance: 8.0},
	}

	SortSpritesByDistance(entities)

	// Should be sorted back-to-front (furthest first)
	if entities[0].Distance != 10.0 {
		t.Errorf("first entity distance = %v, want 10.0", entities[0].Distance)
	}
	if entities[1].Distance != 8.0 {
		t.Errorf("second entity distance = %v, want 8.0", entities[1].Distance)
	}
	if entities[2].Distance != 5.0 {
		t.Errorf("third entity distance = %v, want 5.0", entities[2].Distance)
	}
	if entities[3].Distance != 2.0 {
		t.Errorf("fourth entity distance = %v, want 2.0", entities[3].Distance)
	}
}

func TestIsSpriteColumnVisible(t *testing.T) {
	r := NewRenderer(640, 480)

	// Set some z-buffer values
	r.ZBuffer[100] = 5.0
	r.ZBuffer[200] = 10.0

	tests := []struct {
		name     string
		screenX  int
		distance float64
		want     bool
	}{
		{name: "visible", screenX: 100, distance: 3.0, want: true},
		{name: "behind wall", screenX: 100, distance: 7.0, want: false},
		{name: "at wall", screenX: 100, distance: 5.0, want: false},
		{name: "off screen left", screenX: -1, distance: 3.0, want: false},
		{name: "off screen right", screenX: 700, distance: 3.0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.IsSpriteColumnVisible(tt.screenX, tt.distance)
			if got != tt.want {
				t.Errorf("IsSpriteColumnVisible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyFogToColor(t *testing.T) {
	r := NewRenderer(640, 480)

	// Close distance - minimal fog
	c := sprite.PixelRGBA{R: 255, G: 128, B: 64, A: 255}
	result := r.ApplyFogToColor(c, 1.0)
	if result.R < 200 {
		t.Errorf("close distance fog applied too strongly: R=%d", result.R)
	}

	// Far distance - heavy fog
	result = r.ApplyFogToColor(c, FogDistance)
	if result.R > 100 {
		t.Errorf("far distance fog not applied strongly enough: R=%d", result.R)
	}

	// Alpha should be preserved
	if result.A != 255 {
		t.Errorf("alpha changed from %d to %d", 255, result.A)
	}
}

func TestApplyOpacity(t *testing.T) {
	c := sprite.PixelRGBA{R: 255, G: 128, B: 64, A: 255}

	// Full opacity
	result := ApplyOpacity(c, 1.0)
	if result.A != 255 {
		t.Errorf("full opacity alpha = %d, want 255", result.A)
	}

	// Half opacity
	result = ApplyOpacity(c, 0.5)
	if result.A < 125 || result.A > 130 {
		t.Errorf("half opacity alpha = %d, want ~127", result.A)
	}

	// Zero opacity
	result = ApplyOpacity(c, 0.0)
	if result.A != 0 {
		t.Errorf("zero opacity alpha = %d, want 0", result.A)
	}

	// RGB should be unchanged
	if result.R != 255 || result.G != 128 || result.B != 64 {
		t.Error("opacity should not change RGB values")
	}
}

func TestGetSpritePixel(t *testing.T) {
	frame := sprite.NewSprite(8, 8)
	// Set a pixel
	frame.SetPixel(2, 3, sprite.TestPixelColor)

	// Normal access
	p := GetSpritePixel(frame, 2, 3, false)
	if p.R == 0 && p.G == 0 && p.B == 0 {
		t.Error("expected non-zero pixel")
	}

	// Flipped access - should get pixel from other side
	pFlip := GetSpritePixel(frame, 5, 3, true) // 8-1-2 = 5
	if pFlip.R == 0 && pFlip.G == 0 && pFlip.B == 0 {
		t.Error("expected non-zero pixel with flip")
	}

	// Nil frame
	pNil := GetSpritePixel(nil, 0, 0, false)
	if pNil.A != 0 {
		t.Error("nil frame should return transparent pixel")
	}
}

func TestPrepareSpriteDrawContext(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	// Create a sprite sheet with idle animation
	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	e := &SpriteEntity{
		X:         8.0,
		Y:         5.0,
		Sheet:     sheet,
		AnimState: sprite.AnimIdle,
		AnimFrame: 0,
		Scale:     1.0,
		Opacity:   1.0,
		Visible:   true,
	}
	r.TransformEntityToScreen(e)

	ctx := r.PrepareSpriteDrawContext(e)
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	if ctx.CurrentFrame == nil {
		t.Error("expected non-nil current frame")
	}
	if ctx.Distance <= 0 {
		t.Error("expected positive distance")
	}
}

func TestPrepareSpriteDrawContextNilCases(t *testing.T) {
	r := NewRenderer(640, 480)

	// Nil entity
	if r.PrepareSpriteDrawContext(nil) != nil {
		t.Error("nil entity should return nil context")
	}

	// Nil sheet
	e := &SpriteEntity{Visible: true}
	if r.PrepareSpriteDrawContext(e) != nil {
		t.Error("nil sheet should return nil context")
	}

	// Invisible
	sheet := sprite.NewSpriteSheet(32, 48)
	e = &SpriteEntity{Sheet: sheet, Visible: false}
	if r.PrepareSpriteDrawContext(e) != nil {
		t.Error("invisible entity should return nil context")
	}
}

func TestDrawSprites(t *testing.T) {
	r := NewRenderer(64, 48)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	// Initialize z-buffer to far distance
	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	// Create a sprite sheet with colored pixel
	sheet := sprite.NewSpriteSheet(8, 12)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	frame := sprite.NewSprite(8, 12)
	// Fill with visible pixels
	for y := 0; y < 12; y++ {
		for x := 0; x < 8; x++ {
			frame.SetPixel(x, y, sprite.TestPixelColor)
		}
	}
	idleAnim.AddFrame(frame)
	sheet.AddAnimation(idleAnim)

	entities := []*SpriteEntity{
		{
			X:         8.0,
			Y:         5.0,
			Sheet:     sheet,
			AnimState: sprite.AnimIdle,
			AnimFrame: 0,
			Scale:     1.0,
			Opacity:   1.0,
			Visible:   true,
		},
	}

	pixels := make([]byte, 64*48*4)
	r.DrawSprites(entities, pixels)

	// Check that some pixels were drawn (non-zero in the buffer)
	hasDrawnPixels := false
	for i := 0; i < len(pixels); i += 4 {
		if pixels[i] > 0 || pixels[i+1] > 0 || pixels[i+2] > 0 {
			hasDrawnPixels = true
			break
		}
	}
	if !hasDrawnPixels {
		t.Error("no pixels were drawn")
	}
}

func TestDrawSpritesEmpty(t *testing.T) {
	r := NewRenderer(64, 48)
	pixels := make([]byte, 64*48*4)

	// Should not panic with empty input
	r.DrawSprites(nil, pixels)
	r.DrawSprites([]*SpriteEntity{}, pixels)
}

func TestVisibleSpriteCount(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	// Initialize z-buffer
	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	// Create valid sprite sheet
	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	entities := []*SpriteEntity{
		{X: 8.0, Y: 5.0, Sheet: sheet, AnimState: sprite.AnimIdle, Scale: 1.0, Opacity: 1.0, Visible: true},  // In front
		{X: 2.0, Y: 5.0, Sheet: sheet, AnimState: sprite.AnimIdle, Scale: 1.0, Opacity: 1.0, Visible: true},  // Behind
		{X: 8.0, Y: 5.5, Sheet: sheet, AnimState: sprite.AnimIdle, Scale: 1.0, Opacity: 1.0, Visible: false}, // Invisible
		{X: 9.0, Y: 5.0, Sheet: sheet, AnimState: sprite.AnimIdle, Scale: 1.0, Opacity: 1.0, Visible: true},  // In front
	}

	count := r.VisibleSpriteCount(entities)
	// Should be 2 (in front + in front, excluding behind and invisible)
	if count != 2 {
		t.Errorf("VisibleSpriteCount() = %d, want 2", count)
	}
}

func BenchmarkTransformEntityToScreen(b *testing.B) {
	r := NewRenderer(1280, 720)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.5

	e := &SpriteEntity{
		X:       10.0,
		Y:       7.0,
		Visible: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.TransformEntityToScreen(e)
	}
}

func BenchmarkSortSpritesByDistance(b *testing.B) {
	entities := make([]*SpriteEntity, 50)
	for i := range entities {
		entities[i] = &SpriteEntity{Distance: float64(i % 10)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SortSpritesByDistance(entities)
	}
}

func BenchmarkDrawSprites(b *testing.B) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	// Initialize z-buffer
	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	// Create sprite sheets
	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	frame := sprite.NewSprite(32, 48)
	for y := 0; y < 48; y++ {
		for x := 0; x < 32; x++ {
			frame.SetPixel(x, y, sprite.TestPixelColor)
		}
	}
	idleAnim.AddFrame(frame)
	sheet.AddAnimation(idleAnim)

	entities := make([]*SpriteEntity, 20)
	for i := range entities {
		entities[i] = &SpriteEntity{
			X:         6.0 + float64(i)*0.5,
			Y:         5.0 + float64(i%5)*0.3,
			Sheet:     sheet,
			AnimState: sprite.AnimIdle,
			AnimFrame: 0,
			Scale:     1.0,
			Opacity:   1.0,
			Visible:   true,
		}
	}

	pixels := make([]byte, 640*480*4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.DrawSprites(entities, pixels)
	}
}
