// Package postprocess provides genre-specific visual post-processing effects.
// Per ROADMAP Phase 4 item 17:
// - fantasy=warm color grade
// - sci-fi=scanlines
// - horror=vignette+desaturate
// - cyberpunk=chromatic aberration+bloom
// - post-apoc=sepia+grain
package postprocess

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// Effect represents a post-processing effect that can be applied to an image.
type Effect interface {
	Apply(img *image.RGBA) *image.RGBA
	Name() string
}

// InPlaceEffect extends Effect with a method to write directly to a destination buffer.
type InPlaceEffect interface {
	Effect
	ApplyTo(src, dst *image.RGBA)
}

// Pipeline chains multiple effects together.
type Pipeline struct {
	effects []Effect
	genre   string
	// workBufferA and workBufferB are pre-allocated RGBA buffers for effect processing.
	// Effects alternate between reading from one and writing to the other.
	workBufferA *image.RGBA
	workBufferB *image.RGBA
	// lastWidth and lastHeight track buffer dimensions for reallocation.
	lastWidth  int
	lastHeight int
}

// NewPipeline creates a post-processing pipeline for the given genre.
func NewPipeline(genre string) *Pipeline {
	p := &Pipeline{
		genre:   genre,
		effects: make([]Effect, 0),
	}

	// Add genre-specific effects with intensities to achieve >20% pixel delta
	switch genre {
	case "fantasy":
		p.effects = append(p.effects, &WarmColorGrade{Intensity: 0.6})
	case "sci-fi":
		p.effects = append(p.effects, &Scanlines{Spacing: 2, Intensity: 0.3})
		p.effects = append(p.effects, &Bloom{Threshold: 0.6, Intensity: 0.4})
		p.effects = append(p.effects, &CoolColorGrade{Intensity: 0.4})
	case "horror":
		p.effects = append(p.effects, &Desaturate{Amount: 0.7})
		p.effects = append(p.effects, &Vignette{Radius: 0.5, Softness: 0.3})
		p.effects = append(p.effects, &DarkenOverall{Amount: 0.2})
	case "cyberpunk":
		p.effects = append(p.effects, &ChromaticAberration{Offset: 3})
		p.effects = append(p.effects, &Bloom{Threshold: 0.5, Intensity: 0.5})
		p.effects = append(p.effects, &NeonGlow{Intensity: 0.4})
	case "post-apocalyptic":
		p.effects = append(p.effects, &Sepia{Intensity: 0.8})
		p.effects = append(p.effects, &FilmGrain{Amount: 0.15})
		p.effects = append(p.effects, &Desaturate{Amount: 0.3})
	}

	return p
}

// ensureBuffers allocates or reallocates work buffers if dimensions changed.
func (p *Pipeline) ensureBuffers(bounds image.Rectangle) {
	width := bounds.Dx()
	height := bounds.Dy()
	if p.lastWidth != width || p.lastHeight != height {
		p.workBufferA = image.NewRGBA(bounds)
		p.workBufferB = image.NewRGBA(bounds)
		p.lastWidth = width
		p.lastHeight = height
	}
}

// Apply runs all effects in the pipeline on the image.
// Uses pre-allocated buffers to minimize allocations.
func (p *Pipeline) Apply(img *image.RGBA) *image.RGBA {
	if len(p.effects) == 0 {
		return img
	}

	p.ensureBuffers(img.Bounds())

	// Copy input to work buffer A
	copy(p.workBufferA.Pix, img.Pix)

	// Process effects, alternating between buffers
	src := p.workBufferA
	dst := p.workBufferB

	for _, effect := range p.effects {
		// Check if effect implements InPlaceEffect for optimized path
		if ipe, ok := effect.(InPlaceEffect); ok {
			ipe.ApplyTo(src, dst)
		} else {
			// Fallback: use the allocating Apply method (less efficient)
			result := effect.Apply(src)
			copy(dst.Pix, result.Pix)
		}
		// Swap buffers for next effect
		src, dst = dst, src
	}

	// Copy result back to input image (src now holds the final result)
	copy(img.Pix, src.Pix)
	return img
}

// Genre returns the genre this pipeline was configured for.
func (p *Pipeline) Genre() string {
	return p.genre
}

// Effects returns the list of effects in the pipeline.
func (p *Pipeline) Effects() []Effect {
	return p.effects
}

// WarmColorGrade shifts colors toward warm tones (gold/amber).
type WarmColorGrade struct {
	Intensity float64 // 0.0 to 1.0
}

// Apply applies warm color grading.
func (w *WarmColorGrade) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	w.ApplyTo(img, result)
	return result
}

// ApplyTo applies warm color grading to the destination buffer.
func (w *WarmColorGrade) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)

			// Shift toward warm tones
			r := float64(c.R) + (w.Intensity * 30)
			g := float64(c.G) + (w.Intensity * 15)
			b := float64(c.B) - (w.Intensity * 20)

			dst.SetRGBA(x, y, color.RGBA{
				R: clampByte(r),
				G: clampByte(g),
				B: clampByte(b),
				A: c.A,
			})
		}
	}
}

// Name returns the effect name.
func (w *WarmColorGrade) Name() string { return "WarmColorGrade" }

// Scanlines adds horizontal scanlines typical of CRT displays.
type Scanlines struct {
	Spacing   int     // pixels between lines
	Intensity float64 // 0.0 to 1.0 darkness
}

// Apply adds scanlines to the image.
func (s *Scanlines) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	s.ApplyTo(img, result)
	return result
}

// ApplyTo adds scanlines to the destination buffer.
func (s *Scanlines) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	copy(dst.Pix, src.Pix)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		if y%s.Spacing == 0 {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := dst.RGBAAt(x, y)
				darkFactor := 1.0 - s.Intensity
				dst.SetRGBA(x, y, color.RGBA{
					R: uint8(float64(c.R) * darkFactor),
					G: uint8(float64(c.G) * darkFactor),
					B: uint8(float64(c.B) * darkFactor),
					A: c.A,
				})
			}
		}
	}
}

// Name returns the effect name.
func (s *Scanlines) Name() string { return "Scanlines" }

// Desaturate reduces color saturation.
type Desaturate struct {
	Amount float64 // 0.0 (full color) to 1.0 (grayscale)
}

// Apply reduces saturation.
func (d *Desaturate) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	d.ApplyTo(img, result)
	return result
}

// ApplyTo reduces saturation to the destination buffer.
func (d *Desaturate) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)

			// Calculate luminance
			gray := 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)

			// Blend between original and grayscale
			r := float64(c.R)*(1-d.Amount) + gray*d.Amount
			g := float64(c.G)*(1-d.Amount) + gray*d.Amount
			b := float64(c.B)*(1-d.Amount) + gray*d.Amount

			dst.SetRGBA(x, y, color.RGBA{
				R: clampByte(r),
				G: clampByte(g),
				B: clampByte(b),
				A: c.A,
			})
		}
	}
}

// Name returns the effect name.
func (d *Desaturate) Name() string { return "Desaturate" }

// Vignette darkens the edges of the image.
type Vignette struct {
	Radius   float64 // 0.0 to 1.0, where 1.0 is full radius
	Softness float64 // falloff softness
}

// Apply adds vignette effect.
func (v *Vignette) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	v.ApplyTo(img, result)
	return result
}

// ApplyTo adds vignette effect to the destination buffer.
func (v *Vignette) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	centerX := width / 2
	centerY := height / 2
	maxDist := math.Sqrt(centerX*centerX + centerY*centerY)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)

			// Calculate distance from center
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx+dy*dy) / maxDist

			// Calculate vignette factor
			vignetteFactor := 1.0
			if dist > v.Radius {
				// Guard against division by zero when Radius == 1.0 and Softness == 0.0
				denom := 1.0 - v.Radius + v.Softness
				if denom <= 0 {
					denom = 0.001
				}
				falloff := (dist - v.Radius) / denom
				vignetteFactor = 1.0 - math.Min(1.0, falloff)
			}

			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(float64(c.R) * vignetteFactor),
				G: uint8(float64(c.G) * vignetteFactor),
				B: uint8(float64(c.B) * vignetteFactor),
				A: c.A,
			})
		}
	}
}

// Name returns the effect name.
func (v *Vignette) Name() string { return "Vignette" }

// ChromaticAberration separates RGB channels for a glitch effect.
type ChromaticAberration struct {
	Offset int // pixel offset for R and B channels
}

// Apply adds chromatic aberration.
func (ca *ChromaticAberration) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	ca.ApplyTo(img, result)
	return result
}

// ApplyTo adds chromatic aberration to the destination buffer.
func (ca *ChromaticAberration) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get R channel from offset left
			rx := clampInt(x-ca.Offset, bounds.Min.X, bounds.Max.X-1)
			rColor := src.RGBAAt(rx, y)

			// Get G channel from center
			gColor := src.RGBAAt(x, y)

			// Get B channel from offset right
			bx := clampInt(x+ca.Offset, bounds.Min.X, bounds.Max.X-1)
			bColor := src.RGBAAt(bx, y)

			dst.SetRGBA(x, y, color.RGBA{
				R: rColor.R,
				G: gColor.G,
				B: bColor.B,
				A: gColor.A,
			})
		}
	}
}

// Name returns the effect name.
func (ca *ChromaticAberration) Name() string { return "ChromaticAberration" }

// Bloom adds glow to bright areas.
type Bloom struct {
	Threshold float64 // brightness threshold (0.0 to 1.0)
	Intensity float64 // bloom strength
}

// Apply adds bloom effect.
func (b *Bloom) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	b.ApplyTo(img, result)
	return result
}

// ApplyTo adds bloom effect to the destination buffer.
func (b *Bloom) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	copy(dst.Pix, src.Pix)
	blurRadius := 3
	b.applyBloomToImage(src, dst, bounds, blurRadius)
}

// applyBloomToImage processes all eligible pixels and applies bloom effect.
func (b *Bloom) applyBloomToImage(img, result *image.RGBA, bounds image.Rectangle, blurRadius int) {
	for y := bounds.Min.Y + blurRadius; y < bounds.Max.Y-blurRadius; y++ {
		for x := bounds.Min.X + blurRadius; x < bounds.Max.X-blurRadius; x++ {
			c := img.RGBAAt(x, y)
			brightness := b.calculateBrightness(c)
			if brightness > b.Threshold {
				b.applyBloomToSurrounding(result, x, y, blurRadius, brightness)
			}
		}
	}
}

// calculateBrightness returns the normalized brightness of a pixel (0.0-1.0).
func (b *Bloom) calculateBrightness(c color.RGBA) float64 {
	return (float64(c.R) + float64(c.G) + float64(c.B)) / (255 * 3)
}

// applyBloomToSurrounding adds bloom effect to pixels surrounding a bright pixel.
func (b *Bloom) applyBloomToSurrounding(result *image.RGBA, centerX, centerY, blurRadius int, brightness float64) {
	for dy := -blurRadius; dy <= blurRadius; dy++ {
		for dx := -blurRadius; dx <= blurRadius; dx++ {
			b.addBloomToPixel(result, centerX+dx, centerY+dy, dx, dy, blurRadius, brightness)
		}
	}
}

// addBloomToPixel adds bloom contribution to a single pixel.
func (b *Bloom) addBloomToPixel(result *image.RGBA, nx, ny, dx, dy, blurRadius int, brightness float64) {
	nc := result.RGBAAt(nx, ny)
	dist := math.Sqrt(float64(dx*dx + dy*dy))
	falloff := 1.0 - (dist / float64(blurRadius+1))
	bloomAdd := b.Intensity * falloff * brightness * 30

	result.SetRGBA(nx, ny, color.RGBA{
		R: clampByte(float64(nc.R) + bloomAdd),
		G: clampByte(float64(nc.G) + bloomAdd),
		B: clampByte(float64(nc.B) + bloomAdd),
		A: nc.A,
	})
}

// Name returns the effect name.
func (b *Bloom) Name() string { return "Bloom" }

// Sepia applies a sepia tone.
type Sepia struct {
	Intensity float64 // 0.0 to 1.0
}

// Apply adds sepia tone.
func (s *Sepia) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	s.ApplyTo(img, result)
	return result
}

// ApplyTo adds sepia tone to the destination buffer.
func (s *Sepia) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)

			// Sepia transformation
			sepiaR := 0.393*float64(c.R) + 0.769*float64(c.G) + 0.189*float64(c.B)
			sepiaG := 0.349*float64(c.R) + 0.686*float64(c.G) + 0.168*float64(c.B)
			sepiaB := 0.272*float64(c.R) + 0.534*float64(c.G) + 0.131*float64(c.B)

			// Blend with original
			r := float64(c.R)*(1-s.Intensity) + sepiaR*s.Intensity
			g := float64(c.G)*(1-s.Intensity) + sepiaG*s.Intensity
			b := float64(c.B)*(1-s.Intensity) + sepiaB*s.Intensity

			dst.SetRGBA(x, y, color.RGBA{
				R: clampByte(r),
				G: clampByte(g),
				B: clampByte(b),
				A: c.A,
			})
		}
	}
}

// Name returns the effect name.
func (s *Sepia) Name() string { return "Sepia" }

// FilmGrain adds noise for a film-like effect.
type FilmGrain struct {
	Amount float64 // 0.0 to 1.0
	rng    *rand.Rand
}

// Apply adds film grain.
func (f *FilmGrain) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	f.ApplyTo(img, result)
	return result
}

// ApplyTo adds film grain to the destination buffer.
func (f *FilmGrain) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()

	// Use deterministic seed based on image dimensions for consistency
	if f.rng == nil {
		seed := int64(bounds.Dx()*bounds.Dy() + 42)
		f.rng = rand.New(rand.NewSource(seed))
	}

	noiseRange := f.Amount * 50

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)

			// Add random noise
			noise := (f.rng.Float64() - 0.5) * noiseRange

			dst.SetRGBA(x, y, color.RGBA{
				R: clampByte(float64(c.R) + noise),
				G: clampByte(float64(c.G) + noise),
				B: clampByte(float64(c.B) + noise),
				A: c.A,
			})
		}
	}
}

// Name returns the effect name.
func (f *FilmGrain) Name() string { return "FilmGrain" }

// CoolColorGrade shifts colors toward cool tones (blue/white).
type CoolColorGrade struct {
	Intensity float64 // 0.0 to 1.0
}

// Apply applies cool color grading.
func (c *CoolColorGrade) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	c.ApplyTo(img, result)
	return result
}

// ApplyTo applies cool color grading to the destination buffer.
func (c *CoolColorGrade) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := src.RGBAAt(x, y)

			// Shift toward cool tones
			r := float64(pixel.R) - (c.Intensity * 25)
			g := float64(pixel.G) + (c.Intensity * 10)
			b := float64(pixel.B) + (c.Intensity * 35)

			dst.SetRGBA(x, y, color.RGBA{
				R: clampByte(r),
				G: clampByte(g),
				B: clampByte(b),
				A: pixel.A,
			})
		}
	}
}

// Name returns the effect name.
func (c *CoolColorGrade) Name() string { return "CoolColorGrade" }

// DarkenOverall reduces overall brightness.
type DarkenOverall struct {
	Amount float64 // 0.0 to 1.0
}

// Apply darkens the image.
func (d *DarkenOverall) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	d.ApplyTo(img, result)
	return result
}

// ApplyTo darkens the image to the destination buffer.
func (d *DarkenOverall) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	factor := 1.0 - d.Amount

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(float64(c.R) * factor),
				G: uint8(float64(c.G) * factor),
				B: uint8(float64(c.B) * factor),
				A: c.A,
			})
		}
	}
}

// Name returns the effect name.
func (d *DarkenOverall) Name() string { return "DarkenOverall" }

// NeonGlow adds neon-like glow to bright colors (pink/cyan emphasis).
type NeonGlow struct {
	Intensity float64 // 0.0 to 1.0
}

// Apply adds neon glow effect.
func (n *NeonGlow) Apply(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	n.ApplyTo(img, result)
	return result
}

// ApplyTo adds neon glow effect to the destination buffer.
func (n *NeonGlow) ApplyTo(src, dst *image.RGBA) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)

			// Enhance pinks and cyans (cyberpunk palette)
			brightness := (float64(c.R) + float64(c.G) + float64(c.B)) / 765.0

			// Add pink/magenta tint
			r := float64(c.R) + (n.Intensity * brightness * 40)
			// Reduce green slightly
			g := float64(c.G) - (n.Intensity * 15)
			// Add cyan tint
			b := float64(c.B) + (n.Intensity * brightness * 30)

			dst.SetRGBA(x, y, color.RGBA{
				R: clampByte(r),
				G: clampByte(g),
				B: clampByte(b),
				A: c.A,
			})
		}
	}
}

// Name returns the effect name.
func (n *NeonGlow) Name() string { return "NeonGlow" }

// clampByte clamps a float64 to uint8 range.
func clampByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// clampInt clamps an int to a range.
func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
