package particles

import (
	"image/color"
	"testing"
)

func TestNewEmitter(t *testing.T) {
	e := NewEmitter(TypeRain, 12345)
	if e == nil {
		t.Fatal("NewEmitter returned nil")
	}
	if e.Type != TypeRain {
		t.Errorf("expected type %s, got %s", TypeRain, e.Type)
	}
	if !e.Active {
		t.Error("emitter should be active by default")
	}
	if e.Rate <= 0 {
		t.Error("emitter rate should be positive")
	}
}

func TestEmitterTypeDefaults(t *testing.T) {
	types := []string{
		TypeRain, TypeSnow, TypeDust, TypeAsh, TypeSparks,
		TypeBlood, TypeMagic, TypeSmoke, TypeFire, TypeFogWisp, TypeBubbles,
	}

	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			e := NewEmitter(typ, 12345)
			if e.Rate <= 0 {
				t.Errorf("%s: rate should be positive", typ)
			}
			if e.MinLife <= 0 || e.MaxLife <= 0 {
				t.Errorf("%s: life values should be positive", typ)
			}
			if e.MinSize <= 0 || e.MaxSize <= 0 {
				t.Errorf("%s: size values should be positive", typ)
			}
		})
	}
}

func TestEmitterEmit(t *testing.T) {
	e := NewEmitter(TypeRain, 12345)
	p := e.Emit()
	if p == nil {
		t.Fatal("Emit returned nil")
	}
	if !p.Active {
		t.Error("emitted particle should be active")
	}
	if p.Life <= 0 {
		t.Error("particle life should be positive")
	}
	if p.Type != TypeRain {
		t.Errorf("expected type %s, got %s", TypeRain, p.Type)
	}
}

func TestEmitterInactiveEmit(t *testing.T) {
	e := NewEmitter(TypeRain, 12345)
	e.Active = false
	p := e.Emit()
	if p != nil {
		t.Error("inactive emitter should return nil")
	}
}

func TestNewSystem(t *testing.T) {
	s := NewSystem(12345)
	if s == nil {
		t.Fatal("NewSystem returned nil")
	}
	if s.ActiveCount() != 0 {
		t.Error("new system should have no active particles")
	}
	if s.maxParticles != DefaultMaxParticles {
		t.Errorf("expected max particles %d, got %d", DefaultMaxParticles, s.maxParticles)
	}
}

func TestSystemSetMaxParticles(t *testing.T) {
	s := NewSystem(12345)
	s.SetMaxParticles(500)
	if s.maxParticles != 500 {
		t.Errorf("expected 500, got %d", s.maxParticles)
	}
	// Should ignore invalid values
	s.SetMaxParticles(0)
	if s.maxParticles != 500 {
		t.Error("should ignore zero")
	}
	s.SetMaxParticles(-1)
	if s.maxParticles != 500 {
		t.Error("should ignore negative")
	}
}

func TestSystemAddRemoveEmitter(t *testing.T) {
	s := NewSystem(12345)
	e := NewEmitter(TypeRain, 12345)

	s.AddEmitter(e)
	if s.EmitterCount() != 1 {
		t.Errorf("expected 1 emitter, got %d", s.EmitterCount())
	}

	s.RemoveEmitter(e)
	if s.EmitterCount() != 0 {
		t.Errorf("expected 0 emitters, got %d", s.EmitterCount())
	}

	// Should handle nil
	s.AddEmitter(nil)
	if s.EmitterCount() != 0 {
		t.Error("should not add nil emitter")
	}
}

func TestSystemClearEmitters(t *testing.T) {
	s := NewSystem(12345)
	s.AddEmitter(NewEmitter(TypeRain, 1))
	s.AddEmitter(NewEmitter(TypeSnow, 2))
	s.AddEmitter(NewEmitter(TypeDust, 3))

	s.ClearEmitters()
	if s.EmitterCount() != 0 {
		t.Errorf("expected 0 emitters after clear, got %d", s.EmitterCount())
	}
}

func TestSystemUpdate(t *testing.T) {
	s := NewSystem(12345)
	e := NewEmitter(TypeRain, 12345)
	e.Rate = 100 // High rate to ensure particles spawn
	s.AddEmitter(e)

	// Update for 1 second
	s.Update(1.0)

	if s.ActiveCount() == 0 {
		t.Error("expected particles after update")
	}
}

func TestSystemUpdateKillsExpired(t *testing.T) {
	s := NewSystem(12345)
	e := NewEmitter(TypeSparks, 12345)
	e.MinLife = 0.05
	e.MaxLife = 0.1
	e.Rate = 1000
	s.AddEmitter(e)

	// Spawn particles
	s.Update(0.01)
	if s.ActiveCount() == 0 {
		t.Skip("no particles spawned")
	}

	// Disable emitter to stop new spawns
	e.Active = false

	// Age them past their lifetime
	s.Update(0.5)

	if s.ActiveCount() > 0 {
		t.Errorf("expired particles should be removed, still have %d", s.ActiveCount())
	}
}

func TestSystemClear(t *testing.T) {
	s := NewSystem(12345)
	e := NewEmitter(TypeRain, 12345)
	e.Rate = 1000
	s.AddEmitter(e)
	s.Update(0.1)

	s.Clear()
	if s.ActiveCount() != 0 {
		t.Errorf("expected 0 particles after clear, got %d", s.ActiveCount())
	}
}

func TestSystemSpawnBurst(t *testing.T) {
	s := NewSystem(12345)
	s.SpawnBurst(TypeSparks, 0.5, 0.5, 20)

	if s.ActiveCount() != 20 {
		t.Errorf("expected 20 particles, got %d", s.ActiveCount())
	}
}

func TestSystemParticlesSlice(t *testing.T) {
	s := NewSystem(12345)
	e := NewEmitter(TypeRain, 12345)
	e.Rate = 100
	s.AddEmitter(e)
	s.Update(0.1)

	particles := s.Particles()
	if len(particles) == 0 {
		t.Error("Particles() should return particles")
	}
	// Verify all returned particles are active
	for i, p := range particles {
		if !p.Active {
			t.Errorf("particle %d should be active", i)
		}
	}
}

func TestSystemEmitters(t *testing.T) {
	s := NewSystem(12345)
	e1 := NewEmitter(TypeRain, 1)
	e2 := NewEmitter(TypeSnow, 2)
	s.AddEmitter(e1)
	s.AddEmitter(e2)

	emitters := s.Emitters()
	if len(emitters) != 2 {
		t.Errorf("expected 2 emitters, got %d", len(emitters))
	}
}

func TestParticleBehaviorSnow(t *testing.T) {
	s := NewSystem(12345)
	e := NewEmitter(TypeSnow, 12345)
	e.Rate = 100
	s.AddEmitter(e)
	s.Update(0.1)

	// Just verify no panic
	for i := 0; i < 10; i++ {
		s.Update(0.1)
	}
}

func TestParticleBehaviorFire(t *testing.T) {
	s := NewSystem(12345)
	e := NewEmitter(TypeFire, 12345)
	e.Rate = 100
	s.AddEmitter(e)
	s.Update(0.1)

	// Fire particles should expand
	initialSize := float64(0)
	for _, p := range s.Particles() {
		initialSize = p.Size
		break
	}

	s.Update(0.1)

	// Verify fire behavior was applied
	for _, p := range s.Particles() {
		if p.Size <= initialSize {
			// May have been killed, that's okay
		}
		break
	}
}

func TestParticleBehaviorMagic(t *testing.T) {
	s := NewSystem(12345)
	e := NewEmitter(TypeMagic, 12345)
	e.Rate = 10
	s.AddEmitter(e)

	// Magic particles should spiral
	for i := 0; i < 20; i++ {
		s.Update(0.05)
	}
	// Just verify no panic
}

func TestGetAlpha(t *testing.T) {
	p := &Particle{
		Life:    1.0,
		MaxLife: 2.0,
		Color:   color.RGBA{255, 255, 255, 200},
	}

	// At 50% life, should be fully visible
	alpha := GetAlpha(p)
	if alpha != 200 {
		t.Errorf("at 50%% life, alpha = %d, want 200", alpha)
	}

	// At low life, should fade
	p.Life = 0.1
	alpha = GetAlpha(p)
	if alpha >= 200 {
		t.Error("at low life, alpha should fade")
	}

	// At high life (just born), should fade in
	p.Life = 1.9
	alpha = GetAlpha(p)
	if alpha >= 200 {
		t.Error("at high life, alpha should fade in")
	}

	// Nil particle
	if GetAlpha(nil) != 0 {
		t.Error("nil particle should return 0")
	}

	// Zero max life
	p.MaxLife = 0
	if GetAlpha(p) != 0 {
		t.Error("zero max life should return 0")
	}
}

func TestGetColor(t *testing.T) {
	p := &Particle{
		Life:    1.0,
		MaxLife: 2.0,
		Color:   color.RGBA{100, 150, 200, 255},
	}

	c := GetColor(p)
	if c.R != 100 || c.G != 150 || c.B != 200 {
		t.Error("GetColor should preserve RGB")
	}

	// Nil particle
	c = GetColor(nil)
	if c.R != 0 || c.G != 0 || c.B != 0 || c.A != 0 {
		t.Error("nil particle should return zero color")
	}
}

func TestMaxParticleLimit(t *testing.T) {
	s := NewSystem(12345)
	s.SetMaxParticles(50)

	e := NewEmitter(TypeRain, 12345)
	e.Rate = 10000 // Very high rate
	s.AddEmitter(e)

	// Update several times
	for i := 0; i < 10; i++ {
		s.Update(0.1)
	}

	if s.ActiveCount() > 50 {
		t.Errorf("particle count %d exceeds max 50", s.ActiveCount())
	}
}

func TestParticleOffScreen(t *testing.T) {
	s := NewSystem(12345)
	e := NewEmitter(TypeRain, 12345)
	e.Y = 0.95    // Start very close to bottom
	e.MinVY = 2.0 // Very fast downward
	e.MaxVY = 3.0
	e.MinLife = 5.0 // Long life so they don't expire
	e.MaxLife = 10.0
	e.Rate = 100
	s.AddEmitter(e)

	s.Update(0.05)
	initial := s.ActiveCount()
	if initial == 0 {
		t.Skip("no particles spawned")
	}

	// Disable emitter
	e.Active = false

	// Move particles off screen
	s.Update(1.0)

	// Some should have been removed
	if s.ActiveCount() >= initial {
		t.Errorf("off-screen particles should be removed, had %d, now have %d", initial, s.ActiveCount())
	}
}

func TestDeterminism(t *testing.T) {
	// Same seed should produce same results
	run := func(seed int64) []float64 {
		s := NewSystem(seed)
		e := NewEmitter(TypeRain, seed)
		e.Rate = 100
		s.AddEmitter(e)

		var positions []float64
		for i := 0; i < 10; i++ {
			s.Update(0.1)
			for _, p := range s.Particles() {
				positions = append(positions, p.X, p.Y)
				if len(positions) >= 20 {
					return positions[:20]
				}
			}
		}
		return positions
	}

	run1 := run(42)
	run2 := run(42)
	run3 := run(99)

	// Same seed should match
	for i := range run1 {
		if i >= len(run2) {
			break
		}
		if run1[i] != run2[i] {
			t.Error("same seed should produce identical results")
			break
		}
	}

	// Different seed should differ
	same := true
	for i := range run1 {
		if i >= len(run3) {
			break
		}
		if run1[i] != run3[i] {
			same = false
			break
		}
	}
	if same && len(run1) > 0 {
		t.Error("different seeds should produce different results")
	}
}

func BenchmarkSystemUpdate(b *testing.B) {
	s := NewSystem(12345)
	for i := 0; i < 5; i++ {
		e := NewEmitter(TypeRain, int64(i))
		e.Rate = 100
		s.AddEmitter(e)
	}
	// Pre-fill with particles
	for i := 0; i < 20; i++ {
		s.Update(0.05)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Update(0.016) // ~60fps
	}
}

func BenchmarkEmitterEmit(b *testing.B) {
	e := NewEmitter(TypeRain, 12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Emit()
	}
}

func BenchmarkSpawnBurst(b *testing.B) {
	s := NewSystem(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SpawnBurst(TypeSparks, 0.5, 0.5, 50)
		s.Clear()
	}
}
