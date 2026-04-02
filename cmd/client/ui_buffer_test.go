//go:build !noebiten

package main

import (
	"image/color"
	"testing"
)

func TestNewUIFramebuffer(t *testing.T) {
	fb := NewUIFramebuffer(64, 32)
	if fb == nil {
		t.Fatal("NewUIFramebuffer returned nil")
	}
	if fb.Width() != 64 {
		t.Errorf("expected width 64, got %d", fb.Width())
	}
	if fb.Height() != 32 {
		t.Errorf("expected height 32, got %d", fb.Height())
	}
	if len(fb.pixels) != 64*32*4 {
		t.Errorf("expected pixel buffer size %d, got %d", 64*32*4, len(fb.pixels))
	}
}

func TestUIFramebuffer_Clear(t *testing.T) {
	fb := NewUIFramebuffer(4, 4)
	// Set some pixels first
	fb.SetPixel(0, 0, color.RGBA{255, 0, 0, 255})
	fb.Clear()
	// Check all pixels are zero
	for i := range fb.pixels {
		if fb.pixels[i] != 0 {
			t.Errorf("pixel at index %d should be 0, got %d", i, fb.pixels[i])
		}
	}
}

func TestUIFramebuffer_SetPixel(t *testing.T) {
	fb := NewUIFramebuffer(4, 4)
	c := color.RGBA{255, 128, 64, 200}
	fb.SetPixel(2, 1, c)

	idx := (1*4 + 2) * 4
	if fb.pixels[idx] != 255 || fb.pixels[idx+1] != 128 ||
		fb.pixels[idx+2] != 64 || fb.pixels[idx+3] != 200 {
		t.Errorf("pixel not set correctly, got R=%d G=%d B=%d A=%d",
			fb.pixels[idx], fb.pixels[idx+1], fb.pixels[idx+2], fb.pixels[idx+3])
	}
}

func TestUIFramebuffer_SetPixel_OutOfBounds(t *testing.T) {
	fb := NewUIFramebuffer(4, 4)
	// Should not panic
	fb.SetPixel(-1, 0, color.RGBA{255, 0, 0, 255})
	fb.SetPixel(0, -1, color.RGBA{255, 0, 0, 255})
	fb.SetPixel(4, 0, color.RGBA{255, 0, 0, 255})
	fb.SetPixel(0, 4, color.RGBA{255, 0, 0, 255})
	// All pixels should still be zero
	for i := range fb.pixels {
		if fb.pixels[i] != 0 {
			t.Errorf("out-of-bounds write affected pixel at index %d", i)
		}
	}
}

func TestUIFramebuffer_SetPixelUint32(t *testing.T) {
	fb := NewUIFramebuffer(4, 4)
	fb.SetPixelUint32(1, 1, 0xAABBCCDD)

	idx := (1*4 + 1) * 4
	if fb.pixels[idx] != 0xAA || fb.pixels[idx+1] != 0xBB ||
		fb.pixels[idx+2] != 0xCC || fb.pixels[idx+3] != 0xDD {
		t.Errorf("uint32 pixel not set correctly")
	}
}

func TestUIFramebuffer_DrawRect(t *testing.T) {
	fb := NewUIFramebuffer(8, 8)
	c := color.RGBA{100, 100, 100, 255}
	fb.DrawRect(2, 2, 3, 2, c)

	// Check that pixels inside the rect are set
	for y := 2; y < 4; y++ {
		for x := 2; x < 5; x++ {
			idx := (y*8 + x) * 4
			if fb.pixels[idx] != 100 || fb.pixels[idx+3] != 255 {
				t.Errorf("pixel at (%d,%d) not in rect", x, y)
			}
		}
	}

	// Check that pixels outside are not set
	idx := (0*8 + 0) * 4
	if fb.pixels[idx] != 0 {
		t.Error("pixel outside rect was modified")
	}
}

func TestUIFramebuffer_DrawBorder(t *testing.T) {
	fb := NewUIFramebuffer(8, 8)
	c := color.RGBA{255, 0, 0, 255}
	fb.DrawBorder(1, 1, 4, 3, c)

	// Top edge: y=1, x=1,2,3,4
	for x := 1; x < 5; x++ {
		idx := (1*8 + x) * 4
		if fb.pixels[idx] != 255 || fb.pixels[idx+3] != 255 {
			t.Errorf("top edge pixel at x=%d not set", x)
		}
	}

	// Bottom edge: y=3, x=1,2,3,4
	for x := 1; x < 5; x++ {
		idx := (3*8 + x) * 4
		if fb.pixels[idx] != 255 || fb.pixels[idx+3] != 255 {
			t.Errorf("bottom edge pixel at x=%d not set", x)
		}
	}

	// Left edge: x=1, y=1,2,3
	for y := 1; y < 4; y++ {
		idx := (y*8 + 1) * 4
		if fb.pixels[idx] != 255 || fb.pixels[idx+3] != 255 {
			t.Errorf("left edge pixel at y=%d not set", y)
		}
	}

	// Right edge: x=4, y=1,2,3
	for y := 1; y < 4; y++ {
		idx := (y*8 + 4) * 4
		if fb.pixels[idx] != 255 || fb.pixels[idx+3] != 255 {
			t.Errorf("right edge pixel at y=%d not set", y)
		}
	}

	// Interior should be empty
	idx := (2*8 + 2) * 4
	if fb.pixels[idx] != 0 {
		t.Error("interior pixel was set")
	}
}

func TestUIFramebuffer_Resize(t *testing.T) {
	fb := NewUIFramebuffer(4, 4)
	fb.Resize(8, 8)
	if fb.Width() != 8 || fb.Height() != 8 {
		t.Errorf("resize failed: got %dx%d", fb.Width(), fb.Height())
	}
	if len(fb.pixels) != 8*8*4 {
		t.Errorf("pixel buffer not resized correctly")
	}
}

func TestUIFramebuffer_Resize_SameSize(t *testing.T) {
	fb := NewUIFramebuffer(4, 4)
	originalPixels := fb.pixels
	fb.Resize(4, 4) // Same size
	if &fb.pixels[0] != &originalPixels[0] {
		t.Error("resize with same size should not reallocate")
	}
}

func BenchmarkUIFramebuffer_DrawRect(b *testing.B) {
	fb := NewUIFramebuffer(256, 256)
	c := color.RGBA{100, 100, 100, 255}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fb.DrawRect(10, 10, 100, 50, c)
	}
}

func BenchmarkUIFramebuffer_DrawBorder(b *testing.B) {
	fb := NewUIFramebuffer(256, 256)
	c := color.RGBA{255, 0, 0, 255}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fb.DrawBorder(10, 10, 100, 50, c)
	}
}
