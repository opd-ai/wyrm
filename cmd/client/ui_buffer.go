//go:build !noebiten

// Package main provides UI framebuffer utilities for batch rendering.
package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// UIFramebuffer provides a pre-allocated pixel buffer for batch UI rendering.
// It uses WritePixels() to upload the entire buffer to GPU in a single call,
// avoiding per-pixel Set() calls that cause GPU pipeline synchronization.
type UIFramebuffer struct {
	width  int
	height int
	pixels []byte
	image  *ebiten.Image
	dirty  bool // Tracks whether pixels have changed since last upload
}

// NewUIFramebuffer creates a new UIFramebuffer with the specified dimensions.
func NewUIFramebuffer(width, height int) *UIFramebuffer {
	return &UIFramebuffer{
		width:  width,
		height: height,
		pixels: make([]byte, width*height*4),
		image:  ebiten.NewImage(width, height),
	}
}

// Clear resets all pixels to transparent black.
func (fb *UIFramebuffer) Clear() {
	for i := range fb.pixels {
		fb.pixels[i] = 0
	}
	fb.dirty = true
}

// ClearColor fills the entire framebuffer with the specified color.
func (fb *UIFramebuffer) ClearColor(c color.RGBA) {
	for y := 0; y < fb.height; y++ {
		for x := 0; x < fb.width; x++ {
			idx := (y*fb.width + x) * 4
			fb.pixels[idx] = c.R
			fb.pixels[idx+1] = c.G
			fb.pixels[idx+2] = c.B
			fb.pixels[idx+3] = c.A
		}
	}
	fb.dirty = true
}

// SetPixel sets a single pixel at (x, y) to the specified color.
// Bounds checking is performed; out-of-bounds pixels are ignored.
func (fb *UIFramebuffer) SetPixel(x, y int, c color.RGBA) {
	if x < 0 || x >= fb.width || y < 0 || y >= fb.height {
		return
	}
	idx := (y*fb.width + x) * 4
	fb.pixels[idx] = c.R
	fb.pixels[idx+1] = c.G
	fb.pixels[idx+2] = c.B
	fb.pixels[idx+3] = c.A
	fb.dirty = true
}

// SetPixelUint32 sets a single pixel using a uint32 color (RGBA format).
func (fb *UIFramebuffer) SetPixelUint32(x, y int, c uint32) {
	if x < 0 || x >= fb.width || y < 0 || y >= fb.height {
		return
	}
	idx := (y*fb.width + x) * 4
	fb.pixels[idx] = uint8(c >> 24)
	fb.pixels[idx+1] = uint8(c >> 16)
	fb.pixels[idx+2] = uint8(c >> 8)
	fb.pixels[idx+3] = uint8(c)
	fb.dirty = true
}

// BlendPixel alpha-blends a color onto the existing pixel at (x, y).
func (fb *UIFramebuffer) BlendPixel(x, y int, c color.RGBA) {
	if x < 0 || x >= fb.width || y < 0 || y >= fb.height {
		return
	}
	idx := (y*fb.width + x) * 4
	srcA := float64(c.A) / 255.0
	dstA := 1.0 - srcA

	fb.pixels[idx] = uint8(float64(c.R)*srcA + float64(fb.pixels[idx])*dstA)
	fb.pixels[idx+1] = uint8(float64(c.G)*srcA + float64(fb.pixels[idx+1])*dstA)
	fb.pixels[idx+2] = uint8(float64(c.B)*srcA + float64(fb.pixels[idx+2])*dstA)
	fb.pixels[idx+3] = uint8(float64(c.A)*srcA + float64(fb.pixels[idx+3])*dstA)
	fb.dirty = true
}

// DrawRect fills a rectangle with the specified color.
func (fb *UIFramebuffer) DrawRect(x, y, width, height int, c color.RGBA) {
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			fb.SetPixel(px, py, c)
		}
	}
}

// DrawRectUint32 fills a rectangle using a uint32 color (RGBA format).
func (fb *UIFramebuffer) DrawRectUint32(x, y, width, height int, c uint32) {
	rgba := color.RGBA{
		R: uint8(c >> 24),
		G: uint8(c >> 16),
		B: uint8(c >> 8),
		A: uint8(c),
	}
	fb.DrawRect(x, y, width, height, rgba)
}

// DrawBorder draws a rectangular border (outline only).
func (fb *UIFramebuffer) DrawBorder(x, y, width, height int, c color.RGBA) {
	// Top edge
	for px := x; px < x+width; px++ {
		fb.SetPixel(px, y, c)
	}
	// Bottom edge
	for px := x; px < x+width; px++ {
		fb.SetPixel(px, y+height-1, c)
	}
	// Left edge
	for py := y; py < y+height; py++ {
		fb.SetPixel(x, py, c)
	}
	// Right edge
	for py := y; py < y+height; py++ {
		fb.SetPixel(x+width-1, py, c)
	}
}

// DrawHLine draws a horizontal line from (x, y) with the specified width.
func (fb *UIFramebuffer) DrawHLine(x, y, width int, c color.RGBA) {
	for px := x; px < x+width; px++ {
		fb.SetPixel(px, y, c)
	}
}

// DrawVLine draws a vertical line from (x, y) with the specified height.
func (fb *UIFramebuffer) DrawVLine(x, y, height int, c color.RGBA) {
	for py := y; py < y+height; py++ {
		fb.SetPixel(x, py, c)
	}
}

// Upload writes the pixel buffer to the GPU and returns the image for drawing.
func (fb *UIFramebuffer) Upload() *ebiten.Image {
	if fb.dirty {
		fb.image.WritePixels(fb.pixels)
		fb.dirty = false
	}
	return fb.image
}

// UploadRegion uploads a specific region and returns a sub-image for drawing.
func (fb *UIFramebuffer) UploadRegion(x, y, width, height int) *ebiten.Image {
	if fb.dirty {
		fb.image.WritePixels(fb.pixels)
		fb.dirty = false
	}
	return fb.image.SubImage(image.Rect(x, y, x+width, y+height)).(*ebiten.Image)
}

// DrawTo draws the framebuffer to the target screen at the specified position.
func (fb *UIFramebuffer) DrawTo(screen *ebiten.Image, x, y int) {
	if fb.dirty {
		fb.image.WritePixels(fb.pixels)
		fb.dirty = false
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(fb.image, op)
}

// Width returns the framebuffer width.
func (fb *UIFramebuffer) Width() int {
	return fb.width
}

// Height returns the framebuffer height.
func (fb *UIFramebuffer) Height() int {
	return fb.height
}

// Resize changes the framebuffer dimensions and reallocates buffers.
func (fb *UIFramebuffer) Resize(width, height int) {
	if fb.width == width && fb.height == height {
		return
	}
	fb.width = width
	fb.height = height
	fb.pixels = make([]byte, width*height*4)
	fb.image = ebiten.NewImage(width, height)
	fb.dirty = true
}
