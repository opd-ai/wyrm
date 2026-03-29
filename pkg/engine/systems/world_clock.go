package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// WorldClockSystem advances the in-game time.
type WorldClockSystem struct {
	// DefaultHourLength is seconds per game hour if no clock entity exists.
	DefaultHourLength float64
}

// NewWorldClockSystem creates a new world clock system.
func NewWorldClockSystem(hourLength float64) *WorldClockSystem {
	return &WorldClockSystem{DefaultHourLength: hourLength}
}

// Update advances the game clock each tick.
func (s *WorldClockSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("WorldClock") {
		s.updateClock(w, e, dt)
	}
}

// updateClock updates a single world clock component.
func (s *WorldClockSystem) updateClock(w *ecs.World, e ecs.Entity, dt float64) {
	comp, ok := w.GetComponent(e, "WorldClock")
	if !ok {
		return
	}
	clock := comp.(*components.WorldClock)
	s.ensureHourLength(clock)
	s.advanceTime(clock, dt)
}

// ensureHourLength sets a default hour length if not configured.
func (s *WorldClockSystem) ensureHourLength(clock *components.WorldClock) {
	if clock.HourLength <= 0 {
		clock.HourLength = s.DefaultHourLength
	}
}

// advanceTime accumulates delta time and advances hours/days when needed.
func (s *WorldClockSystem) advanceTime(clock *components.WorldClock, dt float64) {
	clock.TimeAccum += dt
	if clock.TimeAccum >= clock.HourLength {
		clock.TimeAccum -= clock.HourLength
		clock.Hour++
		if clock.Hour >= 24 {
			clock.Hour = 0
			clock.Day++
		}
	}
}
