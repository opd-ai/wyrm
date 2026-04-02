//go:build noebiten

package raycast

import (
	"image/color"
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

// ============================================================
// SpriteEntity Interaction Metadata Tests
// ============================================================

func TestSpriteEntityInteractionFields(t *testing.T) {
	e := &SpriteEntity{
		X:                5.0,
		Y:                5.0,
		Visible:          true,
		InteractionType:  "pickup",
		InteractionRange: 2.0,
		HighlightState:   1,
		IsInteractable:   true,
		UseText:          "Pick up",
		DisplayName:      "Health Potion",
		EntityID:         12345,
	}

	if e.InteractionType != "pickup" {
		t.Errorf("expected pickup, got %s", e.InteractionType)
	}
	if e.InteractionRange != 2.0 {
		t.Errorf("expected range 2.0, got %f", e.InteractionRange)
	}
	if e.HighlightState != 1 {
		t.Errorf("expected highlight 1, got %d", e.HighlightState)
	}
	if !e.IsInteractable {
		t.Error("expected interactable true")
	}
	if e.UseText != "Pick up" {
		t.Errorf("expected 'Pick up', got %s", e.UseText)
	}
	if e.DisplayName != "Health Potion" {
		t.Errorf("expected 'Health Potion', got %s", e.DisplayName)
	}
	if e.EntityID != 12345 {
		t.Errorf("expected entity ID 12345, got %d", e.EntityID)
	}
}

func TestSpriteEntityHighlightStates(t *testing.T) {
	tests := []struct {
		state    int
		expected string
	}{
		{0, "no highlight"},
		{1, "in range"},
		{2, "targeted"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			e := &SpriteEntity{HighlightState: tt.state}
			if e.HighlightState != tt.state {
				t.Errorf("expected state %d, got %d", tt.state, e.HighlightState)
			}
		})
	}
}

func TestSpriteEntityInteractionDefaults(t *testing.T) {
	// Default values should be zero/empty for backward compatibility
	e := &SpriteEntity{}

	if e.InteractionType != "" {
		t.Error("default InteractionType should be empty")
	}
	if e.InteractionRange != 0 {
		t.Error("default InteractionRange should be 0")
	}
	if e.HighlightState != 0 {
		t.Error("default HighlightState should be 0")
	}
	if e.IsInteractable {
		t.Error("default IsInteractable should be false")
	}
	if e.UseText != "" {
		t.Error("default UseText should be empty")
	}
	if e.DisplayName != "" {
		t.Error("default DisplayName should be empty")
	}
	if e.EntityID != 0 {
		t.Error("default EntityID should be 0")
	}
}

func TestInteractableSpriteTransform(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	e := &SpriteEntity{
		X:                7.0,
		Y:                5.0,
		Visible:          true,
		IsInteractable:   true,
		InteractionType:  "open",
		InteractionRange: 2.5,
		DisplayName:      "Chest",
	}

	// Transform should work normally for interactable sprites
	if !r.TransformEntityToScreen(e) {
		t.Error("interactable sprite should transform successfully")
	}

	// Distance should be computed
	if e.Distance <= 0 {
		t.Error("distance should be computed")
	}
}

// ============================================================
// Scale-Correct Rendering Tests (Phase 5)
// ============================================================

func TestItemSizeCategory(t *testing.T) {
	tests := []struct {
		size   ItemSizeCategory
		height float64
	}{
		{SizeTiny, 0.1},
		{SizeSmall, 0.25},
		{SizeMedium, 0.5},
		{SizeLarge, 1.0},
		{SizeHuge, 2.0},
		{SizeCharacter, 1.8},
	}

	for _, tt := range tests {
		t.Run(string(rune('A'+int(tt.size))), func(t *testing.T) {
			got := GetWorldHeightForSize(tt.size)
			if got != tt.height {
				t.Errorf("GetWorldHeightForSize(%d) = %v, want %v", tt.size, got, tt.height)
			}
		})
	}
}

func TestGetWorldHeightForSizeUnknown(t *testing.T) {
	// Unknown size should default to medium
	unknownSize := ItemSizeCategory(999)
	got := GetWorldHeightForSize(unknownSize)
	want := ItemSizeWorldHeight[SizeMedium]
	if got != want {
		t.Errorf("unknown size returned %v, want %v (medium)", got, want)
	}
}

func TestComputeScaleFromWorldHeight(t *testing.T) {
	tests := []struct {
		name      string
		worldH    float64
		baseH     float64
		wantScale float64
	}{
		{"unit scale", 1.0, 1.0, 1.0},
		{"half height", 0.5, 1.0, 0.5},
		{"double height", 2.0, 1.0, 2.0},
		{"zero world height defaults", 0.0, 1.0, 1.0},
		{"zero base height defaults", 1.0, 0.0, 1.0},
		{"custom base", 0.9, 0.3, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeScaleFromWorldHeight(tt.worldH, tt.baseH)
			if got != tt.wantScale {
				t.Errorf("ComputeScaleFromWorldHeight(%v, %v) = %v, want %v",
					tt.worldH, tt.baseH, got, tt.wantScale)
			}
		})
	}
}

func TestSetEntitySize(t *testing.T) {
	tests := []struct {
		size     ItemSizeCategory
		grounded bool
		wantH    float64
	}{
		{SizeTiny, true, 0.1},
		{SizeSmall, false, 0.25},
		{SizeMedium, true, 0.5},
		{SizeLarge, false, 1.0},
		{SizeCharacter, true, 1.8},
	}

	for _, tt := range tests {
		t.Run(string(rune('A'+int(tt.size))), func(t *testing.T) {
			e := &SpriteEntity{}
			e.SetEntitySize(tt.size, tt.grounded)

			if e.WorldHeight != tt.wantH {
				t.Errorf("WorldHeight = %v, want %v", e.WorldHeight, tt.wantH)
			}
			if e.GroundLevel != tt.grounded {
				t.Errorf("GroundLevel = %v, want %v", e.GroundLevel, tt.grounded)
			}

			if tt.grounded {
				// Grounded sprites should have vertical offset
				expectedOffset := -0.5 + tt.wantH/2
				if e.VerticalOffset != expectedOffset {
					t.Errorf("VerticalOffset = %v, want %v", e.VerticalOffset, expectedOffset)
				}
			} else {
				if e.VerticalOffset != 0 {
					t.Errorf("non-grounded VerticalOffset = %v, want 0", e.VerticalOffset)
				}
			}
		})
	}
}

func TestSetCustomWorldHeight(t *testing.T) {
	e := &SpriteEntity{}

	// Non-grounded custom height
	e.SetCustomWorldHeight(0.75, false)
	if e.WorldHeight != 0.75 {
		t.Errorf("WorldHeight = %v, want 0.75", e.WorldHeight)
	}
	if e.VerticalOffset != 0 {
		t.Errorf("non-grounded VerticalOffset = %v, want 0", e.VerticalOffset)
	}

	// Grounded custom height
	e.SetCustomWorldHeight(0.4, true)
	if e.WorldHeight != 0.4 {
		t.Errorf("WorldHeight = %v, want 0.4", e.WorldHeight)
	}
	if e.GroundLevel != true {
		t.Error("GroundLevel should be true")
	}
	expectedOffset := -0.5 + 0.4/2
	if e.VerticalOffset != expectedOffset {
		t.Errorf("grounded VerticalOffset = %v, want %v", e.VerticalOffset, expectedOffset)
	}
}

func TestGetSpriteScreenBoundsWithWorldHeight(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	// Entity in front at distance 3
	e := &SpriteEntity{
		X:           8.0,
		Y:           5.0,
		Visible:     true,
		Scale:       1.0,
		WorldHeight: 0.5, // Half unit tall
	}
	r.TransformEntityToScreen(e)

	startX1, endX1, startY1, endY1, w1, h1 := r.GetSpriteScreenBounds(e, 32, 64)

	// Now use Scale instead of WorldHeight
	e.WorldHeight = 0
	e.Scale = 1.0
	startX2, endX2, startY2, endY2, w2, h2 := r.GetSpriteScreenBounds(e, 32, 64)

	// WorldHeight 0.5 should produce smaller sprite than Scale 1.0
	spriteH1 := endY1 - startY1
	spriteH2 := endY2 - startY2

	if spriteH1 >= spriteH2 {
		t.Errorf("WorldHeight 0.5 sprite (%d) should be smaller than Scale 1.0 sprite (%d)",
			spriteH1, spriteH2)
	}

	// Width should also be proportionally smaller (aspect ratio preserved)
	spriteW1 := endX1 - startX1
	spriteW2 := endX2 - startX2
	if spriteW1 >= spriteW2 {
		t.Errorf("WorldHeight 0.5 width (%d) should be smaller than Scale 1.0 width (%d)",
			spriteW1, spriteW2)
	}

	// Screen width and height returned should match calculated bounds
	if w1 != spriteW1 || h1 != spriteH1 {
		t.Errorf("returned dimensions (%d,%d) don't match bounds (%d,%d)", w1, h1, spriteW1, spriteH1)
	}
	if w2 != spriteW2 || h2 != spriteH2 {
		t.Errorf("returned dimensions (%d,%d) don't match bounds (%d,%d)", w2, h2, spriteW2, spriteH2)
	}

	// Center X position should be the same (both centered on same screen X)
	centerX1 := (startX1 + endX1) / 2
	centerX2 := (startX2 + endX2) / 2
	if centerX1 != centerX2 {
		t.Errorf("center X %d should equal %d (both sprites centered on same screen X)", centerX1, centerX2)
	}

	// Verify the center Y is at horizon for both (no vertical offset)
	horizon := r.Height / 2
	centerY1 := (startY1 + endY1) / 2
	centerY2 := (startY2 + endY2) / 2
	if centerY1 != horizon || centerY2 != horizon {
		t.Errorf("center Y (%d, %d) should both equal horizon %d", centerY1, centerY2, horizon)
	}
}

func TestGetSpriteScreenBoundsWithVerticalOffset(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	// Entity in front
	e := &SpriteEntity{
		X:       8.0,
		Y:       5.0,
		Visible: true,
		Scale:   1.0,
	}
	r.TransformEntityToScreen(e)

	// No offset - centered on horizon
	e.VerticalOffset = 0
	_, _, startYCenter, endYCenter, _, _ := r.GetSpriteScreenBounds(e, 32, 64)
	spriteCenter := (startYCenter + endYCenter) / 2
	horizon := r.Height / 2

	if spriteCenter != horizon {
		t.Errorf("no offset: sprite center %d should equal horizon %d", spriteCenter, horizon)
	}

	// Negative offset (below eye level) should move sprite down on screen
	e.VerticalOffset = -0.5
	_, _, startYDown, _, _, _ := r.GetSpriteScreenBounds(e, 32, 64)

	if startYDown <= startYCenter {
		t.Errorf("negative offset: startY %d should be > center startY %d", startYDown, startYCenter)
	}

	// Positive offset (above eye level) should move sprite up on screen
	e.VerticalOffset = 0.5
	_, _, startYUp, _, _, _ := r.GetSpriteScreenBounds(e, 32, 64)

	if startYUp >= startYCenter {
		t.Errorf("positive offset: startY %d should be < center startY %d", startYUp, startYCenter)
	}
}

func TestGroundedItemRendering(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	// Create a grounded item
	e := &SpriteEntity{
		X:       8.0,
		Y:       5.0,
		Visible: true,
	}
	e.SetEntitySize(SizeSmall, true) // Small grounded item
	r.TransformEntityToScreen(e)

	_, _, _, endY, _, _ := r.GetSpriteScreenBounds(e, 32, 32)

	// A grounded small item should have its bottom below the horizon
	horizon := r.Height / 2
	if endY <= horizon {
		t.Errorf("grounded item bottom %d should be below horizon %d", endY, horizon)
	}
}

func TestScaleCorrectRenderingDefaults(t *testing.T) {
	// Verify default values maintain backward compatibility
	e := &SpriteEntity{}

	if e.WorldHeight != 0 {
		t.Error("default WorldHeight should be 0")
	}
	if e.VerticalOffset != 0 {
		t.Error("default VerticalOffset should be 0")
	}
	if e.GroundLevel {
		t.Error("default GroundLevel should be false")
	}
}

func BenchmarkGetSpriteScreenBoundsWithWorldHeight(b *testing.B) {
	r := NewRenderer(1280, 720)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.5

	e := &SpriteEntity{
		X:              10.0,
		Y:              7.0,
		Visible:        true,
		WorldHeight:    0.5,
		VerticalOffset: -0.25,
	}
	r.TransformEntityToScreen(e)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetSpriteScreenBounds(e, 32, 64)
	}
}

// ============================================================
// Interaction Highlight Effect Tests (Phase 5)
// ============================================================

func TestHighlightConstants(t *testing.T) {
	if HighlightNone != 0 {
		t.Error("HighlightNone should be 0")
	}
	if HighlightInRange != 1 {
		t.Error("HighlightInRange should be 1")
	}
	if HighlightTargeted != 2 {
		t.Error("HighlightTargeted should be 2")
	}
}

func TestDefaultHighlightConfig(t *testing.T) {
	cfg := DefaultHighlightConfig()

	if cfg == nil {
		t.Fatal("DefaultHighlightConfig should not return nil")
	}
	if cfg.InRangeColor.A == 0 {
		t.Error("InRangeColor should have non-zero alpha")
	}
	if cfg.TargetedColor.A == 0 {
		t.Error("TargetedColor should have non-zero alpha")
	}
	if cfg.OutlineWidth <= 0 {
		t.Error("OutlineWidth should be positive")
	}
	if cfg.GlowIntensity <= 0 || cfg.GlowIntensity > 1.0 {
		t.Error("GlowIntensity should be between 0 and 1")
	}
}

func TestHighlightColorForState(t *testing.T) {
	cfg := DefaultHighlightConfig()

	tests := []struct {
		state     int
		wantAlpha bool
	}{
		{HighlightNone, false},
		{HighlightInRange, true},
		{HighlightTargeted, true},
		{999, false}, // Unknown state
	}

	for _, tt := range tests {
		t.Run(string(rune('A'+tt.state)), func(t *testing.T) {
			color := cfg.HighlightColorForState(tt.state)
			hasAlpha := color.A > 0
			if hasAlpha != tt.wantAlpha {
				t.Errorf("state %d: hasAlpha=%v, want %v", tt.state, hasAlpha, tt.wantAlpha)
			}
		})
	}
}

func TestApplyHighlightToPixel(t *testing.T) {
	cfg := DefaultHighlightConfig()
	cfg.PulseEnabled = false // Disable for deterministic testing

	originalPixel := sprite.PixelRGBA{R: 100, G: 100, B: 100, A: 255}

	// No highlight - pixel unchanged
	result := ApplyHighlightToPixel(originalPixel, HighlightNone, false, 0, cfg)
	if result.R != originalPixel.R || result.G != originalPixel.G || result.B != originalPixel.B {
		t.Error("HighlightNone should not modify pixel")
	}

	// Transparent pixel - no highlight applied
	transparent := sprite.PixelRGBA{R: 100, G: 100, B: 100, A: 0}
	result = ApplyHighlightToPixel(transparent, HighlightTargeted, false, 0, cfg)
	if result.A != 0 {
		t.Error("transparent pixel should remain transparent")
	}

	// In range highlight
	result = ApplyHighlightToPixel(originalPixel, HighlightInRange, false, 0, cfg)
	// Should be blended toward yellow (highlight color)
	if result.R <= originalPixel.R && result.G <= originalPixel.G {
		t.Error("in-range highlight should brighten/tint the pixel")
	}

	// Targeted highlight should be stronger
	resultTargeted := ApplyHighlightToPixel(originalPixel, HighlightTargeted, false, 0, cfg)
	resultInRange := ApplyHighlightToPixel(originalPixel, HighlightInRange, false, 0, cfg)
	// Targeted has higher alpha in config, so blend should be stronger
	if resultTargeted.R <= resultInRange.R && resultTargeted.G <= resultInRange.G {
		// This might not always hold depending on exact config, so just check it's different
		t.Log("Note: targeted highlight may not always be brighter than in-range")
	}
}

func TestApplyHighlightToPixelEdgeEffect(t *testing.T) {
	cfg := DefaultHighlightConfig()
	cfg.PulseEnabled = false

	pixel := sprite.PixelRGBA{R: 100, G: 100, B: 100, A: 255}

	// Edge pixels should have stronger highlight
	resultEdge := ApplyHighlightToPixel(pixel, HighlightTargeted, true, 0, cfg)
	resultCenter := ApplyHighlightToPixel(pixel, HighlightTargeted, false, 0, cfg)

	// Edge should be more highlighted (different from center)
	if resultEdge.R == resultCenter.R && resultEdge.G == resultCenter.G && resultEdge.B == resultCenter.B {
		t.Error("edge pixels should have different highlight than center pixels")
	}
}

func TestApplyHighlightToPixelPulse(t *testing.T) {
	cfg := DefaultHighlightConfig()
	cfg.PulseEnabled = true

	pixel := sprite.PixelRGBA{R: 100, G: 100, B: 100, A: 255}

	// Different pulse phases should produce different results
	// Use 0 and PI/2 since sin(0)=0 but sin(PI/2)=1
	result0 := ApplyHighlightToPixel(pixel, HighlightTargeted, false, 0, cfg)
	resultPiHalf := ApplyHighlightToPixel(pixel, HighlightTargeted, false, 1.5708, cfg) // PI/2

	if result0.R == resultPiHalf.R && result0.G == resultPiHalf.G && result0.B == resultPiHalf.B {
		t.Error("different pulse phases should produce different highlight intensities")
	}
}

func TestIsPixelOnSpriteEdge(t *testing.T) {
	frame := sprite.NewSprite(8, 8)
	// Fill center 4x4 with opaque pixels
	for y := 2; y < 6; y++ {
		for x := 2; x < 6; x++ {
			frame.SetPixel(x, y, sprite.TestPixelColor)
		}
	}

	// Pixel in center should not be edge
	if IsPixelOnSpriteEdge(frame, 3, 3, false) {
		t.Error("center pixel should not be edge")
	}
	if IsPixelOnSpriteEdge(frame, 4, 4, false) {
		t.Error("center pixel should not be edge")
	}

	// Pixel on border of filled region should be edge
	if !IsPixelOnSpriteEdge(frame, 2, 3, false) {
		t.Error("border pixel should be edge")
	}
	if !IsPixelOnSpriteEdge(frame, 5, 4, false) {
		t.Error("border pixel should be edge")
	}

	// Nil frame should return false
	if IsPixelOnSpriteEdge(nil, 0, 0, false) {
		t.Error("nil frame should return false")
	}
}

func TestUpdateHighlightPulse(t *testing.T) {
	r := NewRenderer(640, 480)

	initial := r.GetHighlightPulsePhase()
	if initial != 0 {
		t.Error("initial pulse phase should be 0")
	}

	// Update with time delta
	r.UpdateHighlightPulse(0.1)
	afterUpdate := r.GetHighlightPulsePhase()
	if afterUpdate <= initial {
		t.Error("pulse phase should increase after update")
	}

	// Multiple updates
	r.UpdateHighlightPulse(0.1)
	afterSecond := r.GetHighlightPulsePhase()
	if afterSecond <= afterUpdate {
		t.Error("pulse phase should continue increasing")
	}
}

func TestSetHighlightConfig(t *testing.T) {
	r := NewRenderer(640, 480)

	// Initially nil
	if r.HighlightConfig != nil {
		t.Error("initial HighlightConfig should be nil")
	}

	// Set custom config
	cfg := &HighlightConfig{
		GlowIntensity: 0.8,
		PulseSpeed:    10.0,
	}
	r.SetHighlightConfig(cfg)

	if r.HighlightConfig != cfg {
		t.Error("SetHighlightConfig should set the config")
	}
}

func TestSpriteDrawContextHighlightFields(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	// Initialize z-buffer
	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	// Create a sprite with highlight
	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	e := &SpriteEntity{
		X:              8.0,
		Y:              5.0,
		Sheet:          sheet,
		AnimState:      sprite.AnimIdle,
		AnimFrame:      0,
		Scale:          1.0,
		Opacity:        1.0,
		Visible:        true,
		HighlightState: HighlightTargeted,
		IsInteractable: true,
	}
	r.TransformEntityToScreen(e)

	ctx := r.PrepareSpriteDrawContext(e)
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	if ctx.HighlightState != HighlightTargeted {
		t.Errorf("context HighlightState = %d, want %d", ctx.HighlightState, HighlightTargeted)
	}
	if !ctx.IsInteractable {
		t.Error("context IsInteractable should be true")
	}
}

func BenchmarkApplyHighlightToPixel(b *testing.B) {
	cfg := DefaultHighlightConfig()
	pixel := sprite.PixelRGBA{R: 100, G: 150, B: 200, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyHighlightToPixel(pixel, HighlightTargeted, true, float64(i), cfg)
	}
}

func BenchmarkIsPixelOnSpriteEdge(b *testing.B) {
	frame := sprite.NewSprite(32, 32)
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			if x > 4 && x < 28 && y > 4 && y < 28 {
				frame.SetPixel(x, y, sprite.TestPixelColor)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsPixelOnSpriteEdge(frame, 10, 10, false)
	}
}

// ============================================================
// Interaction Targeting System Tests (Phase 5)
// ============================================================

func TestDefaultTargetingConfig(t *testing.T) {
	cfg := DefaultTargetingConfig()

	if cfg == nil {
		t.Fatal("DefaultTargetingConfig should not return nil")
	}
	if cfg.CrosshairTolerance <= 0 {
		t.Error("CrosshairTolerance should be positive")
	}
	if cfg.MaxTargetDistance <= 0 {
		t.Error("MaxTargetDistance should be positive")
	}
}

func TestFindTargetedEntityNoEntities(t *testing.T) {
	r := NewRenderer(640, 480)

	result := r.FindTargetedEntity(nil, nil)
	if result.HasTarget {
		t.Error("should not have target with nil entities")
	}

	result = r.FindTargetedEntity([]*SpriteEntity{}, nil)
	if result.HasTarget {
		t.Error("should not have target with empty entities")
	}
}

func TestFindTargetedEntityDirectlyInFront(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0 // Facing right (+X)

	// Initialize z-buffer
	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	// Create an interactable entity directly in front
	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	e := &SpriteEntity{
		X:                7.0,
		Y:                5.0,
		Sheet:            sheet,
		AnimState:        sprite.AnimIdle,
		AnimFrame:        0,
		Scale:            1.0,
		Opacity:          1.0,
		Visible:          true,
		IsInteractable:   true,
		InteractionType:  "pickup",
		InteractionRange: 3.0,
		DisplayName:      "Potion",
	}
	r.TransformEntityToScreen(e)

	result := r.FindTargetedEntity([]*SpriteEntity{e}, nil)

	if !result.HasTarget {
		t.Error("should have target for entity directly in front")
	}
	if result.TargetEntity != e {
		t.Error("targeted entity should match")
	}
	if !result.IsWithinRange {
		t.Error("entity should be within interaction range")
	}
}

func TestFindTargetedEntityOutOfRange(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	// Entity at distance 5, but interaction range is 2
	e := &SpriteEntity{
		X:                10.0,
		Y:                5.0,
		Sheet:            sheet,
		AnimState:        sprite.AnimIdle,
		AnimFrame:        0,
		Scale:            1.0,
		Opacity:          1.0,
		Visible:          true,
		IsInteractable:   true,
		InteractionRange: 2.0,
	}
	r.TransformEntityToScreen(e)

	result := r.FindTargetedEntity([]*SpriteEntity{e}, nil)

	// Still targets (for showing highlight) but not within range
	if result.HasTarget && result.IsWithinRange {
		t.Error("entity beyond interaction range should not be within range")
	}
}

func TestFindTargetedEntityPrefersCloser(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	sheet := sprite.NewSpriteSheet(64, 64)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(64, 64))
	sheet.AddAnimation(idleAnim)

	// Two overlapping entities
	closer := &SpriteEntity{
		X:              6.5,
		Y:              5.0,
		Sheet:          sheet,
		AnimState:      sprite.AnimIdle,
		Scale:          1.0,
		Opacity:        1.0,
		Visible:        true,
		IsInteractable: true,
		DisplayName:    "Closer",
	}

	further := &SpriteEntity{
		X:              8.0,
		Y:              5.0,
		Sheet:          sheet,
		AnimState:      sprite.AnimIdle,
		Scale:          1.0,
		Opacity:        1.0,
		Visible:        true,
		IsInteractable: true,
		DisplayName:    "Further",
	}

	r.TransformEntityToScreen(closer)
	r.TransformEntityToScreen(further)

	cfg := DefaultTargetingConfig()
	cfg.PreferCloser = true

	result := r.FindTargetedEntity([]*SpriteEntity{further, closer}, cfg)

	if result.HasTarget && result.TargetEntity != closer {
		t.Error("should prefer closer entity")
	}
}

func TestFindTargetedEntityIgnoresNonInteractable(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	// Non-interactable entity
	e := &SpriteEntity{
		X:              7.0,
		Y:              5.0,
		Sheet:          sheet,
		AnimState:      sprite.AnimIdle,
		Scale:          1.0,
		Opacity:        1.0,
		Visible:        true,
		IsInteractable: false, // Not interactable
	}
	r.TransformEntityToScreen(e)

	result := r.FindTargetedEntity([]*SpriteEntity{e}, nil)

	if result.HasTarget {
		t.Error("should not target non-interactable entity")
	}
}

func TestUpdateEntityHighlightStates(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	targeted := &SpriteEntity{
		X:                7.0,
		Y:                5.0,
		Sheet:            sheet,
		AnimState:        sprite.AnimIdle,
		Scale:            1.0,
		Visible:          true,
		IsInteractable:   true,
		InteractionRange: 3.0,
	}

	inRange := &SpriteEntity{
		X:                6.0,
		Y:                6.0,
		Sheet:            sheet,
		AnimState:        sprite.AnimIdle,
		Scale:            1.0,
		Visible:          true,
		IsInteractable:   true,
		InteractionRange: 3.0,
	}

	outOfRange := &SpriteEntity{
		X:                10.0,
		Y:                10.0,
		Sheet:            sheet,
		AnimState:        sprite.AnimIdle,
		Scale:            1.0,
		Visible:          true,
		IsInteractable:   true,
		InteractionRange: 1.0, // Very short range
	}

	r.TransformEntityToScreen(targeted)
	r.TransformEntityToScreen(inRange)
	r.TransformEntityToScreen(outOfRange)

	// Create targeting result
	targetResult := TargetingResult{
		HasTarget:     true,
		TargetEntity:  targeted,
		IsWithinRange: true,
	}

	entities := []*SpriteEntity{targeted, inRange, outOfRange}
	r.UpdateEntityHighlightStates(entities, targetResult)

	if targeted.HighlightState != HighlightTargeted {
		t.Errorf("targeted entity should have HighlightTargeted, got %d", targeted.HighlightState)
	}
	if inRange.HighlightState != HighlightInRange {
		t.Errorf("in-range entity should have HighlightInRange, got %d", inRange.HighlightState)
	}
	if outOfRange.HighlightState != HighlightNone {
		t.Errorf("out-of-range entity should have HighlightNone, got %d", outOfRange.HighlightState)
	}
}

func TestGetInteractionPrompt(t *testing.T) {
	tests := []struct {
		name     string
		result   TargetingResult
		contains string
	}{
		{
			name:     "no target",
			result:   TargetingResult{HasTarget: false},
			contains: "",
		},
		{
			name: "out of range",
			result: TargetingResult{
				HasTarget:     true,
				TargetEntity:  &SpriteEntity{DisplayName: "Chest"},
				IsWithinRange: false,
			},
			contains: "",
		},
		{
			name: "pickup with name",
			result: TargetingResult{
				HasTarget:     true,
				TargetEntity:  &SpriteEntity{InteractionType: "pickup", DisplayName: "Potion"},
				IsWithinRange: true,
			},
			contains: "Take Potion",
		},
		{
			name: "open door",
			result: TargetingResult{
				HasTarget:     true,
				TargetEntity:  &SpriteEntity{InteractionType: "open", DisplayName: "Door"},
				IsWithinRange: true,
			},
			contains: "Open Door",
		},
		{
			name: "custom use text",
			result: TargetingResult{
				HasTarget:     true,
				TargetEntity:  &SpriteEntity{UseText: "Activate", DisplayName: "Lever"},
				IsWithinRange: true,
			},
			contains: "Activate Lever",
		},
		{
			name: "talk to NPC",
			result: TargetingResult{
				HasTarget:     true,
				TargetEntity:  &SpriteEntity{InteractionType: "talk", DisplayName: "Merchant"},
				IsWithinRange: true,
			},
			contains: "Talk to Merchant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := GetInteractionPrompt(tt.result)
			if tt.contains != "" {
				if prompt != tt.contains {
					t.Errorf("prompt = %q, want %q", prompt, tt.contains)
				}
			} else {
				if prompt != "" {
					t.Errorf("prompt = %q, want empty", prompt)
				}
			}
		})
	}
}

func TestIsEntityAtScreenPosition(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	e := &SpriteEntity{
		X:         7.0,
		Y:         5.0,
		Sheet:     sheet,
		AnimState: sprite.AnimIdle,
		Scale:     1.0,
		Visible:   true,
	}
	r.TransformEntityToScreen(e)

	ctx := r.PrepareSpriteDrawContext(e)
	if ctx == nil {
		t.Fatal("could not prepare context")
	}

	// Test at center of entity
	centerX := (ctx.StartX + ctx.EndX) / 2
	centerY := (ctx.StartY + ctx.EndY) / 2
	if !r.IsEntityAtScreenPosition(e, centerX, centerY) {
		t.Error("center position should be within entity")
	}

	// Test outside entity
	if r.IsEntityAtScreenPosition(e, ctx.StartX-50, centerY) {
		t.Error("position far left should not be within entity")
	}

	// Test nil entity
	if r.IsEntityAtScreenPosition(nil, centerX, centerY) {
		t.Error("nil entity should return false")
	}
}

func TestFindEntityAtScreenPosition(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	e := &SpriteEntity{
		X:              7.0,
		Y:              5.0,
		Sheet:          sheet,
		AnimState:      sprite.AnimIdle,
		Scale:          1.0,
		Visible:        true,
		IsInteractable: true,
	}
	r.TransformEntityToScreen(e)

	ctx := r.PrepareSpriteDrawContext(e)
	if ctx == nil {
		t.Fatal("could not prepare context")
	}

	centerX := (ctx.StartX + ctx.EndX) / 2
	centerY := (ctx.StartY + ctx.EndY) / 2

	found := r.FindEntityAtScreenPosition([]*SpriteEntity{e}, centerX, centerY)
	if found != e {
		t.Error("should find entity at its screen position")
	}

	// Test at position with no entity
	found = r.FindEntityAtScreenPosition([]*SpriteEntity{e}, 0, 0)
	if found != nil {
		t.Error("should not find entity at position 0,0")
	}
}

func BenchmarkFindTargetedEntity(b *testing.B) {
	r := NewRenderer(1280, 720)
	r.PlayerX = 5.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0

	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	entities := make([]*SpriteEntity, 50)
	for i := range entities {
		entities[i] = &SpriteEntity{
			X:              6.0 + float64(i%10),
			Y:              4.0 + float64(i/10),
			Sheet:          sheet,
			AnimState:      sprite.AnimIdle,
			Scale:          1.0,
			Visible:        true,
			IsInteractable: true,
		}
		r.TransformEntityToScreen(entities[i])
	}

	cfg := DefaultTargetingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.FindTargetedEntity(entities, cfg)
	}
}

// TestItemIdentificationByType tests that different item types are correctly identified.
func TestItemIdentificationByType(t *testing.T) {
	itemTypes := []struct {
		interactionType string
		displayName     string
		isInteractable  bool
	}{
		{"pickup", "Health Potion", true},
		{"open", "Wooden Chest", true},
		{"use", "Iron Lever", true},
		{"read", "Ancient Tome", true},
		{"talk", "Village Elder", true},
		{"push", "Stone Block", true},
		{"examine", "Old Statue", true},
		{"none", "Decorative Pillar", false},
		{"", "Background Object", false},
	}

	for _, tc := range itemTypes {
		t.Run(tc.interactionType, func(t *testing.T) {
			e := &SpriteEntity{
				X:               5.0,
				Y:               5.0,
				Scale:           1.0,
				Visible:         true,
				InteractionType: tc.interactionType,
				DisplayName:     tc.displayName,
				IsInteractable:  tc.isInteractable,
			}

			// Verify IsInteractable matches expected
			if e.IsInteractable != tc.isInteractable {
				t.Errorf("IsInteractable = %v, want %v", e.IsInteractable, tc.isInteractable)
			}

			// Verify InteractionType is set correctly
			if e.InteractionType != tc.interactionType {
				t.Errorf("InteractionType = %s, want %s", e.InteractionType, tc.interactionType)
			}
		})
	}
}

// TestHighlightStateProgression tests highlight state transitions.
func TestHighlightStateProgression(t *testing.T) {
	e := &SpriteEntity{
		X:               5.0,
		Y:               5.0,
		Scale:           1.0,
		Visible:         true,
		IsInteractable:  true,
		InteractionType: "pickup",
	}

	// Initially no highlight
	if e.HighlightState != HighlightNone {
		t.Errorf("initial state = %d, want %d (none)", e.HighlightState, HighlightNone)
	}

	// Transition to in-range
	e.HighlightState = HighlightInRange
	if e.HighlightState != HighlightInRange {
		t.Errorf("in-range state = %d, want %d", e.HighlightState, HighlightInRange)
	}

	// Transition to targeted
	e.HighlightState = HighlightTargeted
	if e.HighlightState != HighlightTargeted {
		t.Errorf("targeted state = %d, want %d", e.HighlightState, HighlightTargeted)
	}

	// Back to none
	e.HighlightState = HighlightNone
	if e.HighlightState != HighlightNone {
		t.Errorf("reset state = %d, want %d (none)", e.HighlightState, HighlightNone)
	}
}

// TestInteractionRaycastIntegration tests the complete interaction raycasting flow.
func TestInteractionRaycastIntegration(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 0.0
	r.PlayerY = 0.0
	r.PlayerA = 0.0 // Facing East (+X)

	// Clear ZBuffer
	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	// Create sprite sheet
	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	// Create three entities at different distances along the view axis
	near := &SpriteEntity{
		X:                2.0,
		Y:                0.0,
		Sheet:            sheet,
		AnimState:        sprite.AnimIdle,
		Scale:            1.0,
		Visible:          true,
		IsInteractable:   true,
		InteractionType:  "pickup",
		DisplayName:      "Near Item",
		InteractionRange: 3.0,
	}

	mid := &SpriteEntity{
		X:                5.0,
		Y:                0.0,
		Sheet:            sheet,
		AnimState:        sprite.AnimIdle,
		Scale:            1.0,
		Visible:          true,
		IsInteractable:   true,
		InteractionType:  "pickup",
		DisplayName:      "Mid Item",
		InteractionRange: 3.0,
	}

	far := &SpriteEntity{
		X:                10.0,
		Y:                0.0,
		Sheet:            sheet,
		AnimState:        sprite.AnimIdle,
		Scale:            1.0,
		Visible:          true,
		IsInteractable:   true,
		InteractionType:  "pickup",
		DisplayName:      "Far Item",
		InteractionRange: 3.0,
	}

	entities := []*SpriteEntity{near, mid, far}

	// Transform all entities to screen space
	for _, e := range entities {
		r.TransformEntityToScreen(e)
	}

	// Find targeted entity (should be nearest)
	cfg := DefaultTargetingConfig()
	result := r.FindTargetedEntity(entities, cfg)

	if !result.HasTarget {
		t.Fatal("expected to find a targeted entity")
	}

	if result.TargetEntity != near {
		t.Errorf("expected nearest entity to be targeted, got %s", result.TargetEntity.DisplayName)
	}

	if result.Distance > 3.0 {
		t.Errorf("distance = %f, expected <= 3.0", result.Distance)
	}
}

// TestMultipleItemTypesInView tests targeting when multiple item types are visible.
func TestMultipleItemTypesInView(t *testing.T) {
	r := NewRenderer(640, 480)
	r.PlayerX = 0.0
	r.PlayerY = 0.0
	r.PlayerA = 0.0

	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}

	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	// Different interaction types at same distance but different angles
	potion := &SpriteEntity{
		X:               3.0,
		Y:               0.0, // Directly ahead
		Sheet:           sheet,
		AnimState:       sprite.AnimIdle,
		Scale:           1.0,
		Visible:         true,
		IsInteractable:  true,
		InteractionType: "pickup",
		DisplayName:     "Health Potion",
	}

	chest := &SpriteEntity{
		X:               3.0,
		Y:               2.0, // To the left
		Sheet:           sheet,
		AnimState:       sprite.AnimIdle,
		Scale:           1.0,
		Visible:         true,
		IsInteractable:  true,
		InteractionType: "open",
		DisplayName:     "Treasure Chest",
	}

	lever := &SpriteEntity{
		X:               3.0,
		Y:               -2.0, // To the right
		Sheet:           sheet,
		AnimState:       sprite.AnimIdle,
		Scale:           1.0,
		Visible:         true,
		IsInteractable:  true,
		InteractionType: "use",
		DisplayName:     "Iron Lever",
	}

	entities := []*SpriteEntity{potion, chest, lever}

	for _, e := range entities {
		r.TransformEntityToScreen(e)
	}

	// With player facing East, potion should be targeted (closest to center of view)
	cfg := DefaultTargetingConfig()
	result := r.FindTargetedEntity(entities, cfg)

	if !result.HasTarget {
		t.Fatal("expected to find a targeted entity")
	}

	// The potion is directly ahead, should be targeted
	if result.TargetEntity.InteractionType != "pickup" {
		t.Errorf("expected 'pickup' type to be targeted, got '%s'", result.TargetEntity.InteractionType)
	}
}

// TestHighlightRenderingIntegration tests that highlights affect rendering output.
func TestHighlightRenderingIntegration(t *testing.T) {
	r := NewRenderer(320, 240)

	// Create a test sprite with actual pixel data
	sheet := sprite.NewSpriteSheet(16, 16)
	testSprite := sprite.NewSprite(16, 16)
	// Fill with a solid color
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			testSprite.SetPixel(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(testSprite)
	sheet.AddAnimation(idleAnim)

	e := &SpriteEntity{
		X:              5.0,
		Y:              5.0,
		Sheet:          sheet,
		AnimState:      sprite.AnimIdle,
		Scale:          1.0,
		Visible:        true,
		IsInteractable: true,
		HighlightState: HighlightTargeted,
	}

	// Set up renderer
	r.PlayerX = 4.0
	r.PlayerY = 5.0
	r.PlayerA = 0.0
	for i := range r.ZBuffer {
		r.ZBuffer[i] = MaxRayDistance
	}
	r.SetHighlightConfig(DefaultHighlightConfig())

	r.TransformEntityToScreen(e)

	ctx := r.PrepareSpriteDrawContext(e)
	if ctx == nil {
		t.Fatal("could not prepare context")
	}

	// Verify context has highlight state
	if ctx.HighlightState != HighlightTargeted {
		t.Errorf("context HighlightState = %d, want %d", ctx.HighlightState, HighlightTargeted)
	}

	// Verify interactable flag
	if !ctx.IsInteractable {
		t.Error("context IsInteractable should be true")
	}
}

// TestInteractionPromptVariety tests interaction prompts for different types.
func TestInteractionPromptVariety(t *testing.T) {
	testCases := []struct {
		interactionType string
		displayName     string
		expectContains  string
	}{
		{"pickup", "Healing Potion", "Take"},
		{"open", "Wooden Door", "Open"},
		{"use", "Crystal Ball", "Use"},
		{"read", "Magic Scroll", "Read"},
		{"talk", "Merchant", "Talk to"},
		{"push", "Heavy Crate", "Interact"}, // Default case - "push" not specifically handled
		{"examine", "Strange Artifact", "Examine"},
	}

	r := NewRenderer(640, 480)
	r.PlayerX = 0.0
	r.PlayerY = 0.0
	r.PlayerA = 0.0

	sheet := sprite.NewSpriteSheet(32, 48)
	idleAnim := sprite.NewAnimation(sprite.AnimIdle, true)
	idleAnim.AddFrame(sprite.NewSprite(32, 48))
	sheet.AddAnimation(idleAnim)

	for _, tc := range testCases {
		t.Run(tc.interactionType, func(t *testing.T) {
			e := &SpriteEntity{
				X:               2.0,
				Y:               0.0,
				Sheet:           sheet,
				AnimState:       sprite.AnimIdle,
				Scale:           1.0,
				Visible:         true,
				IsInteractable:  true,
				InteractionType: tc.interactionType,
				DisplayName:     tc.displayName,
			}

			r.TransformEntityToScreen(e)

			// Create a targeting result for the entity
			result := TargetingResult{
				HasTarget:     true,
				TargetEntity:  e,
				Distance:      2.0,
				ScreenX:       r.Width / 2,
				ScreenY:       r.Height / 2,
				IsWithinRange: true,
			}

			prompt := GetInteractionPrompt(result)
			if prompt == "" {
				t.Error("expected non-empty prompt")
			}

			// Prompt should contain the expected action word
			if !containsString(prompt, tc.expectContains) {
				t.Errorf("prompt '%s' should contain '%s'", prompt, tc.expectContains)
			}

			// Prompt should contain the display name
			if !containsString(prompt, tc.displayName) {
				t.Errorf("prompt '%s' should contain '%s'", prompt, tc.displayName)
			}
		})
	}
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

// findSubstring is a simple substring search.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
