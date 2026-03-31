package sprite

import (
	"image/color"
	"testing"
)

func TestNewSprite(t *testing.T) {
	t.Run("default dimensions", func(t *testing.T) {
		s := NewSprite(0, 0)
		if s.Width != DefaultSpriteWidth {
			t.Errorf("expected width %d, got %d", DefaultSpriteWidth, s.Width)
		}
		if s.Height != DefaultSpriteHeight {
			t.Errorf("expected height %d, got %d", DefaultSpriteHeight, s.Height)
		}
	})

	t.Run("custom dimensions", func(t *testing.T) {
		s := NewSprite(64, 96)
		if s.Width != 64 {
			t.Errorf("expected width 64, got %d", s.Width)
		}
		if s.Height != 96 {
			t.Errorf("expected height 96, got %d", s.Height)
		}
	})

	t.Run("clamped to max", func(t *testing.T) {
		s := NewSprite(500, 500)
		if s.Width != MaxSpriteWidth {
			t.Errorf("expected width clamped to %d, got %d", MaxSpriteWidth, s.Width)
		}
		if s.Height != MaxSpriteHeight {
			t.Errorf("expected height clamped to %d, got %d", MaxSpriteHeight, s.Height)
		}
	})

	t.Run("pixels initialized", func(t *testing.T) {
		s := NewSprite(10, 10)
		if len(s.Pixels) != 100 {
			t.Errorf("expected 100 pixels, got %d", len(s.Pixels))
		}
	})
}

func TestSpriteGetSetPixel(t *testing.T) {
	s := NewSprite(10, 10)
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	t.Run("set and get valid pixel", func(t *testing.T) {
		s.SetPixel(5, 5, red)
		got := s.GetPixel(5, 5)
		if got != red {
			t.Errorf("expected %v, got %v", red, got)
		}
	})

	t.Run("get out of bounds returns transparent", func(t *testing.T) {
		got := s.GetPixel(-1, 0)
		if got.A != 0 {
			t.Error("expected transparent pixel for negative x")
		}
		got = s.GetPixel(100, 0)
		if got.A != 0 {
			t.Error("expected transparent pixel for x > width")
		}
	})

	t.Run("set out of bounds does nothing", func(t *testing.T) {
		s.SetPixel(-1, -1, red)
		s.SetPixel(100, 100, red)
		// Should not panic
	})
}

func TestSpriteMemorySize(t *testing.T) {
	s := NewSprite(32, 48)
	expected := int64(32 * 48 * 4) // 4 bytes per RGBA pixel
	if s.MemorySize() != expected {
		t.Errorf("expected memory size %d, got %d", expected, s.MemorySize())
	}
}

func TestSpriteClone(t *testing.T) {
	s := NewSprite(10, 10)
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	s.SetPixel(5, 5, red)

	clone := s.Clone()

	if clone.Width != s.Width || clone.Height != s.Height {
		t.Error("clone dimensions don't match")
	}
	if clone.GetPixel(5, 5) != red {
		t.Error("clone pixel doesn't match")
	}

	// Modify original, clone should be unaffected
	s.SetPixel(5, 5, color.RGBA{G: 255, A: 255})
	if clone.GetPixel(5, 5) != red {
		t.Error("clone was affected by original modification")
	}
}

func TestSpriteFill(t *testing.T) {
	s := NewSprite(10, 10)
	blue := color.RGBA{B: 255, A: 255}
	s.Fill(blue)

	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			if s.GetPixel(x, y) != blue {
				t.Errorf("pixel at (%d,%d) not filled", x, y)
			}
		}
	}
}

func TestSpriteFlipHorizontal(t *testing.T) {
	s := NewSprite(10, 10)
	red := color.RGBA{R: 255, A: 255}
	s.SetPixel(0, 0, red) // Top-left

	flipped := s.FlipHorizontal()

	// Top-left should now be at top-right
	if flipped.GetPixel(9, 0) != red {
		t.Error("flipped pixel not in expected position")
	}
	if flipped.GetPixel(0, 0).A != 0 {
		t.Error("original position should be transparent after flip")
	}
}

func TestNewAnimation(t *testing.T) {
	anim := NewAnimation(AnimWalk, true)

	if anim.Name != AnimWalk {
		t.Errorf("expected name %s, got %s", AnimWalk, anim.Name)
	}
	if !anim.Loop {
		t.Error("expected looping animation")
	}
	if anim.FrameDuration != 1.0/AnimFrameRate {
		t.Errorf("unexpected frame duration: %f", anim.FrameDuration)
	}
}

func TestAnimationAddFrame(t *testing.T) {
	anim := NewAnimation(AnimIdle, true)
	frame := NewSprite(10, 10)

	anim.AddFrame(frame)

	if anim.FrameCount() != 1 {
		t.Errorf("expected 1 frame, got %d", anim.FrameCount())
	}
}

func TestAnimationGetFrame(t *testing.T) {
	anim := NewAnimation(AnimWalk, true)

	t.Run("empty animation returns nil", func(t *testing.T) {
		if anim.GetFrame(0) != nil {
			t.Error("expected nil for empty animation")
		}
	})

	// Add frames
	for i := 0; i < 4; i++ {
		anim.AddFrame(NewSprite(10, 10))
	}

	t.Run("valid index", func(t *testing.T) {
		if anim.GetFrame(2) == nil {
			t.Error("expected non-nil frame")
		}
	})

	t.Run("looping wraps index", func(t *testing.T) {
		frame4 := anim.GetFrame(4)
		frame0 := anim.GetFrame(0)
		if frame4 != frame0 {
			t.Error("looping animation should wrap")
		}
	})

	t.Run("non-looping clamps", func(t *testing.T) {
		nonLoop := NewAnimation(AnimAttack, false)
		for i := 0; i < 3; i++ {
			nonLoop.AddFrame(NewSprite(10, 10))
		}
		frame10 := nonLoop.GetFrame(10)
		frame2 := nonLoop.GetFrame(2)
		if frame10 != frame2 {
			t.Error("non-looping should clamp to last frame")
		}
	})

	t.Run("negative index returns first", func(t *testing.T) {
		if anim.GetFrame(-1) != anim.GetFrame(0) {
			t.Error("negative index should return first frame")
		}
	})
}

func TestAnimationDuration(t *testing.T) {
	anim := NewAnimation(AnimIdle, true)
	for i := 0; i < 8; i++ {
		anim.AddFrame(NewSprite(10, 10))
	}

	expected := 8.0 / AnimFrameRate
	if anim.Duration() != expected {
		t.Errorf("expected duration %f, got %f", expected, anim.Duration())
	}
}

func TestNewSpriteSheet(t *testing.T) {
	sheet := NewSpriteSheet(32, 48)

	if sheet.BaseWidth != 32 || sheet.BaseHeight != 48 {
		t.Error("unexpected base dimensions")
	}
	if len(sheet.Animations) != 0 {
		t.Error("expected empty animations map")
	}
}

func TestSpriteSheetAddGetAnimation(t *testing.T) {
	sheet := NewSpriteSheet(32, 48)

	idle := NewAnimation(AnimIdle, true)
	idle.AddFrame(NewSprite(32, 48))
	sheet.AddAnimation(idle)

	walk := NewAnimation(AnimWalk, true)
	walk.AddFrame(NewSprite(32, 48))
	sheet.AddAnimation(walk)

	t.Run("get existing animation", func(t *testing.T) {
		got := sheet.GetAnimation(AnimWalk)
		if got != walk {
			t.Error("didn't get expected animation")
		}
	})

	t.Run("fallback to idle", func(t *testing.T) {
		got := sheet.GetAnimation("nonexistent")
		if got != idle {
			t.Error("should fall back to idle")
		}
	})

	t.Run("add nil does nothing", func(t *testing.T) {
		sheet.AddAnimation(nil)
		if len(sheet.Animations) != 2 {
			t.Error("nil animation should not be added")
		}
	})
}

func TestSpriteSheetGetFrame(t *testing.T) {
	sheet := NewSpriteSheet(32, 48)
	idle := NewAnimation(AnimIdle, true)
	frame := NewSprite(32, 48)
	idle.AddFrame(frame)
	sheet.AddAnimation(idle)

	got := sheet.GetFrame(AnimIdle, 0)
	if got != frame {
		t.Error("didn't get expected frame")
	}
}

func TestSpriteSheetMemorySize(t *testing.T) {
	sheet := NewSpriteSheet(32, 48)
	idle := NewAnimation(AnimIdle, true)
	idle.AddFrame(NewSprite(32, 48))
	idle.AddFrame(NewSprite(32, 48))
	sheet.AddAnimation(idle)

	expected := int64(32 * 48 * 4 * 2) // 2 frames
	if sheet.MemorySize() != expected {
		t.Errorf("expected memory size %d, got %d", expected, sheet.MemorySize())
	}
}

func TestSpriteSheetAnimationNames(t *testing.T) {
	sheet := NewSpriteSheet(32, 48)
	sheet.AddAnimation(NewAnimation(AnimIdle, true))
	sheet.AddAnimation(NewAnimation(AnimWalk, true))

	names := sheet.AnimationNames()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}

	hasIdle := false
	hasWalk := false
	for _, n := range names {
		if n == AnimIdle {
			hasIdle = true
		}
		if n == AnimWalk {
			hasWalk = true
		}
	}
	if !hasIdle || !hasWalk {
		t.Error("missing expected animation names")
	}
}
