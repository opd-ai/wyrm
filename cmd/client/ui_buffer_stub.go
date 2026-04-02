//go:build noebiten

// Package main provides stub UIFramebuffer for noebiten builds.
package main

import (
	"image/color"
)

// UIFramebuffer stub for noebiten builds (no Ebiten dependencies).
type UIFramebuffer struct {
	width  int
	height int
	pixels []byte
	dirty  bool
}

// NewUIFramebuffer creates a stub UIFramebuffer.
func NewUIFramebuffer(width, height int) *UIFramebuffer {
	return &UIFramebuffer{
		width:  width,
		height: height,
		pixels: make([]byte, width*height*4),
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
	for px := x; px < x+width; px++ {
		fb.SetPixel(px, y, c)
	}
	for px := x; px < x+width; px++ {
		fb.SetPixel(px, y+height-1, c)
	}
	for py := y; py < y+height; py++ {
		fb.SetPixel(x, py, c)
	}
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
	fb.dirty = true
}
