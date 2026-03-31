// Package particles implements a procedural particle system.
package particles

import (
	"image/color"
	"math"
	"math/rand"
)

// Particle type identifiers.
const (
	TypeRain    = "rain"
	TypeSnow    = "snow"
	TypeDust    = "dust"
	TypeAsh     = "ash"
	TypeSparks  = "sparks"
	TypeBlood   = "blood"
	TypeMagic   = "magic"
	TypeSmoke   = "smoke"
	TypeFire    = "fire"
	TypeFogWisp = "fog_wisp"
	TypeBubbles = "bubbles"
)

// Default particle system limits.
const (
	DefaultMaxParticles = 1000
	DefaultPoolSize     = 2000
	MinParticleLife     = 0.1
	MaxParticleLife     = 10.0
)

// Particle represents a single particle in the system.
type Particle struct {
	// Position in screen space (0-1 normalized)
	X, Y float64
	// Velocity in screen space per second
	VX, VY float64
	// Remaining lifetime in seconds
	Life float64
	// Maximum lifetime (for alpha fade calculation)
	MaxLife float64
	// Size in pixels
	Size float64
	// Color
	Color color.RGBA
	// Type identifier
	Type string
	// Active flag for pool management
	Active bool
	// Custom data for type-specific behavior
	Data float64
}

// Emitter generates particles of a specific type.
type Emitter struct {
	// Type of particles to emit
	Type string
	// Position in screen space (0-1 normalized)
	X, Y float64
	// Emission rate (particles per second)
	Rate float64
	// Particle lifetime range
	MinLife, MaxLife float64
	// Particle size range
	MinSize, MaxSize float64
	// Velocity ranges
	MinVX, MaxVX float64
	MinVY, MaxVY float64
	// Color
	Color color.RGBA
	// Active flag
	Active bool
	// Accumulated time for emission
	accumulator float64
	// RNG for variation
	rng *rand.Rand
}

// NewEmitter creates a new particle emitter.
func NewEmitter(particleType string, seed int64) *Emitter {
	e := &Emitter{
		Type:    particleType,
		X:       0.5,
		Y:       0.0,
		Rate:    50,
		MinLife: 1.0,
		MaxLife: 2.0,
		MinSize: 1.0,
		MaxSize: 3.0,
		MinVX:   -0.01,
		MaxVX:   0.01,
		MinVY:   0.1,
		MaxVY:   0.2,
		Color:   color.RGBA{255, 255, 255, 255},
		Active:  true,
		rng:     rand.New(rand.NewSource(seed)),
	}
	e.applyTypeDefaults()
	return e
}

// applyTypeDefaults sets default values based on particle type.
func (e *Emitter) applyTypeDefaults() {
	switch e.Type {
	case TypeRain:
		e.MinVY = 0.8
		e.MaxVY = 1.2
		e.MinVX = -0.02
		e.MaxVX = 0.02
		e.MinLife = 0.5
		e.MaxLife = 1.0
		e.MinSize = 1.0
		e.MaxSize = 2.0
		e.Color = color.RGBA{150, 180, 200, 180}
		e.Rate = 200
	case TypeSnow:
		e.MinVY = 0.05
		e.MaxVY = 0.15
		e.MinVX = -0.03
		e.MaxVX = 0.03
		e.MinLife = 3.0
		e.MaxLife = 5.0
		e.MinSize = 2.0
		e.MaxSize = 4.0
		e.Color = color.RGBA{240, 240, 255, 200}
		e.Rate = 80
	case TypeDust:
		e.MinVY = -0.01
		e.MaxVY = 0.02
		e.MinVX = -0.05
		e.MaxVX = 0.05
		e.MinLife = 2.0
		e.MaxLife = 4.0
		e.MinSize = 1.0
		e.MaxSize = 3.0
		e.Color = color.RGBA{180, 160, 120, 100}
		e.Rate = 30
	case TypeAsh:
		e.MinVY = -0.02
		e.MaxVY = 0.01
		e.MinVX = -0.02
		e.MaxVX = 0.02
		e.MinLife = 4.0
		e.MaxLife = 8.0
		e.MinSize = 2.0
		e.MaxSize = 4.0
		e.Color = color.RGBA{80, 80, 80, 150}
		e.Rate = 40
	case TypeSparks:
		e.MinVY = -0.3
		e.MaxVY = -0.1
		e.MinVX = -0.2
		e.MaxVX = 0.2
		e.MinLife = 0.3
		e.MaxLife = 0.8
		e.MinSize = 1.0
		e.MaxSize = 2.0
		e.Color = color.RGBA{255, 200, 50, 255}
		e.Rate = 100
	case TypeBlood:
		e.MinVY = 0.1
		e.MaxVY = 0.3
		e.MinVX = -0.15
		e.MaxVX = 0.15
		e.MinLife = 0.5
		e.MaxLife = 1.5
		e.MinSize = 2.0
		e.MaxSize = 4.0
		e.Color = color.RGBA{180, 20, 20, 200}
		e.Rate = 50
	case TypeMagic:
		e.MinVY = -0.1
		e.MaxVY = 0.1
		e.MinVX = -0.1
		e.MaxVX = 0.1
		e.MinLife = 0.5
		e.MaxLife = 1.5
		e.MinSize = 2.0
		e.MaxSize = 5.0
		e.Color = color.RGBA{100, 150, 255, 200}
		e.Rate = 60
	case TypeSmoke:
		e.MinVY = -0.08
		e.MaxVY = -0.02
		e.MinVX = -0.02
		e.MaxVX = 0.02
		e.MinLife = 2.0
		e.MaxLife = 4.0
		e.MinSize = 4.0
		e.MaxSize = 8.0
		e.Color = color.RGBA{100, 100, 100, 120}
		e.Rate = 20
	case TypeFire:
		e.MinVY = -0.15
		e.MaxVY = -0.05
		e.MinVX = -0.03
		e.MaxVX = 0.03
		e.MinLife = 0.3
		e.MaxLife = 0.8
		e.MinSize = 3.0
		e.MaxSize = 6.0
		e.Color = color.RGBA{255, 150, 50, 220}
		e.Rate = 80
	case TypeFogWisp:
		e.MinVY = -0.01
		e.MaxVY = 0.01
		e.MinVX = -0.03
		e.MaxVX = 0.03
		e.MinLife = 3.0
		e.MaxLife = 6.0
		e.MinSize = 8.0
		e.MaxSize = 16.0
		e.Color = color.RGBA{200, 200, 200, 60}
		e.Rate = 10
	case TypeBubbles:
		e.MinVY = -0.1
		e.MaxVY = -0.03
		e.MinVX = -0.02
		e.MaxVX = 0.02
		e.MinLife = 1.0
		e.MaxLife = 3.0
		e.MinSize = 2.0
		e.MaxSize = 5.0
		e.Color = color.RGBA{200, 220, 255, 100}
		e.Rate = 25
	}
}

// Emit creates a new particle at the emitter position.
func (e *Emitter) Emit() *Particle {
	if !e.Active {
		return nil
	}
	life := e.MinLife + e.rng.Float64()*(e.MaxLife-e.MinLife)
	return &Particle{
		X:       e.X + (e.rng.Float64()-0.5)*0.2,
		Y:       e.Y,
		VX:      e.MinVX + e.rng.Float64()*(e.MaxVX-e.MinVX),
		VY:      e.MinVY + e.rng.Float64()*(e.MaxVY-e.MinVY),
		Life:    life,
		MaxLife: life,
		Size:    e.MinSize + e.rng.Float64()*(e.MaxSize-e.MinSize),
		Color:   e.Color,
		Type:    e.Type,
		Active:  true,
	}
}

// System manages particles and emitters.
type System struct {
	particles    []*Particle
	emitters     []*Emitter
	pool         []*Particle
	maxParticles int
	activeCount  int
	seed         int64
	rng          *rand.Rand
}

// NewSystem creates a new particle system.
func NewSystem(seed int64) *System {
	s := &System{
		particles:    make([]*Particle, 0, DefaultMaxParticles),
		emitters:     make([]*Emitter, 0, 16),
		pool:         make([]*Particle, DefaultPoolSize),
		maxParticles: DefaultMaxParticles,
		seed:         seed,
		rng:          rand.New(rand.NewSource(seed)),
	}
	// Pre-allocate pool
	for i := range s.pool {
		s.pool[i] = &Particle{}
	}
	return s
}

// SetMaxParticles sets the maximum active particle count.
func (s *System) SetMaxParticles(max int) {
	if max > 0 {
		s.maxParticles = max
	}
}

// AddEmitter adds an emitter to the system.
func (s *System) AddEmitter(e *Emitter) {
	if e != nil {
		s.emitters = append(s.emitters, e)
	}
}

// RemoveEmitter removes an emitter from the system.
func (s *System) RemoveEmitter(e *Emitter) {
	for i, em := range s.emitters {
		if em == e {
			s.emitters = append(s.emitters[:i], s.emitters[i+1:]...)
			return
		}
	}
}

// ClearEmitters removes all emitters.
func (s *System) ClearEmitters() {
	s.emitters = s.emitters[:0]
}

// acquireParticle gets a particle from the pool or creates a new one.
func (s *System) acquireParticle() *Particle {
	// Try to reuse from pool
	for _, p := range s.pool {
		if !p.Active {
			return p
		}
	}
	// Pool exhausted, create new
	return &Particle{}
}

// Update advances the particle system by dt seconds.
func (s *System) Update(dt float64) {
	// Update existing particles and compact active ones
	s.updateParticles(dt)

	// Emit new particles from active emitters
	s.emitParticles(dt)
}

// updateParticles updates all particles and compacts the active list.
func (s *System) updateParticles(dt float64) {
	activeIdx := 0
	for i := 0; i < len(s.particles); i++ {
		p := s.particles[i]
		if s.updateSingleParticle(p, dt) {
			// Compact active particles
			if activeIdx != i {
				s.particles[activeIdx] = p
			}
			activeIdx++
		}
	}
	s.particles = s.particles[:activeIdx]
	s.activeCount = activeIdx
}

// updateSingleParticle updates one particle, returns true if still active.
func (s *System) updateSingleParticle(p *Particle, dt float64) bool {
	if !p.Active {
		return false
	}

	p.Life -= dt
	if p.Life <= 0 {
		p.Active = false
		return false
	}

	// Update position
	p.X += p.VX * dt
	p.Y += p.VY * dt

	// Apply type-specific behavior
	s.updateParticleBehavior(p, dt)

	// Remove if off screen
	if p.X < -0.1 || p.X > 1.1 || p.Y < -0.1 || p.Y > 1.1 {
		p.Active = false
		return false
	}

	return true
}

// emitParticles processes all emitters and spawns new particles.
func (s *System) emitParticles(dt float64) {
	for _, e := range s.emitters {
		if !e.Active {
			continue
		}
		s.emitFromEmitter(e, dt)
	}
}

// emitFromEmitter spawns particles from a single emitter.
func (s *System) emitFromEmitter(e *Emitter, dt float64) {
	e.accumulator += dt
	interval := 1.0 / e.Rate
	for e.accumulator >= interval {
		e.accumulator -= interval
		if s.activeCount < s.maxParticles {
			p := e.Emit()
			if p != nil {
				s.particles = append(s.particles, p)
				s.activeCount++
			}
		}
	}
}

// updateParticleBehavior applies type-specific physics.
func (s *System) updateParticleBehavior(p *Particle, dt float64) {
	switch p.Type {
	case TypeSnow:
		// Gentle oscillation
		p.VX += (s.rng.Float64()-0.5)*0.1*dt - p.VX*0.5*dt
	case TypeDust, TypeAsh:
		// Drifting
		p.VX += (s.rng.Float64() - 0.5) * 0.02 * dt
		p.VY += (s.rng.Float64() - 0.5) * 0.01 * dt
	case TypeSparks:
		// Gravity
		p.VY += 0.5 * dt
		// Fade color to red
		lifeRatio := p.Life / p.MaxLife
		p.Color.R = uint8(255 * lifeRatio)
		p.Color.G = uint8(200 * lifeRatio * lifeRatio)
	case TypeFire:
		// Slight expansion and color shift
		p.Size += dt * 2
		lifeRatio := p.Life / p.MaxLife
		p.Color.G = uint8(150 * lifeRatio)
		p.Color.B = uint8(50 * lifeRatio)
	case TypeSmoke:
		// Expand and slow
		p.Size += dt * 2
		p.VY *= 0.98
	case TypeMagic:
		// Spiral motion
		p.Data += dt * 5
		p.VX = 0.05 * math.Sin(p.Data)
		p.VY = 0.05 * math.Cos(p.Data)
	case TypeBubbles:
		// Wobble
		p.VX += (s.rng.Float64() - 0.5) * 0.05 * dt
	}
}

// Particles returns the current active particles.
func (s *System) Particles() []*Particle {
	return s.particles
}

// ActiveCount returns the number of active particles.
func (s *System) ActiveCount() int {
	return s.activeCount
}

// Clear removes all particles.
func (s *System) Clear() {
	for _, p := range s.particles {
		p.Active = false
	}
	s.particles = s.particles[:0]
	s.activeCount = 0
}

// EmitterCount returns the number of emitters.
func (s *System) EmitterCount() int {
	return len(s.emitters)
}

// Emitters returns the current emitters.
func (s *System) Emitters() []*Emitter {
	return s.emitters
}

// SpawnBurst spawns a burst of particles at a position.
func (s *System) SpawnBurst(particleType string, x, y float64, count int) {
	e := NewEmitter(particleType, s.rng.Int63())
	e.X = x
	e.Y = y

	for i := 0; i < count && s.activeCount < s.maxParticles; i++ {
		p := e.Emit()
		if p != nil {
			// Radial burst
			angle := float64(i) * 2 * math.Pi / float64(count)
			speed := e.MinVY + e.rng.Float64()*(e.MaxVY-e.MinVY)
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle) * speed
			s.particles = append(s.particles, p)
			s.activeCount++
		}
	}
}

// GetAlpha returns the alpha value for a particle based on its remaining life.
func GetAlpha(p *Particle) uint8 {
	if p == nil || p.MaxLife <= 0 {
		return 0
	}
	lifeRatio := p.Life / p.MaxLife
	// Fade in for first 10%, fade out for last 30%
	if lifeRatio > 0.9 {
		// Fade in: at 1.0 -> 0.0, at 0.9 -> 1.0
		fadeIn := (1.0 - lifeRatio) / 0.1 // 0 to 1
		return uint8(float64(p.Color.A) * fadeIn)
	}
	if lifeRatio < 0.3 {
		// Fade out: at 0.3 -> 1.0, at 0.0 -> 0.0
		return uint8(float64(p.Color.A) * lifeRatio / 0.3)
	}
	return p.Color.A
}

// GetColor returns the particle color with proper alpha.
func GetColor(p *Particle) color.RGBA {
	if p == nil {
		return color.RGBA{}
	}
	return color.RGBA{
		R: p.Color.R,
		G: p.Color.G,
		B: p.Color.B,
		A: GetAlpha(p),
	}
}
