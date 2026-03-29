package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// NPCScheduleSystem drives NPC daily activity cycles.
type NPCScheduleSystem struct {
	// WorldHour is updated externally by WorldClockSystem
	WorldHour int
}

// Update processes NPC schedules based on the current world hour.
func (s *NPCScheduleSystem) Update(w *ecs.World, dt float64) {
	s.syncWorldHour(w)
	s.updateNPCSchedules(w)
}

// syncWorldHour reads the world clock from a WorldClock entity.
func (s *NPCScheduleSystem) syncWorldHour(w *ecs.World) {
	for _, e := range w.Entities("WorldClock") {
		comp, ok := w.GetComponent(e, "WorldClock")
		if ok {
			clock := comp.(*components.WorldClock)
			s.WorldHour = clock.Hour
			return
		}
	}
}

// updateNPCSchedules updates all NPC schedules based on the current hour.
func (s *NPCScheduleSystem) updateNPCSchedules(w *ecs.World) {
	for _, e := range w.Entities("Schedule") {
		s.updateEntitySchedule(w, e)
	}
}

// updateEntitySchedule updates a single entity's current activity.
func (s *NPCScheduleSystem) updateEntitySchedule(w *ecs.World, e ecs.Entity) {
	comp, ok := w.GetComponent(e, "Schedule")
	if !ok {
		return
	}
	sched := comp.(*components.Schedule)
	if activity, ok := sched.TimeSlots[s.WorldHour]; ok && activity != sched.CurrentActivity {
		sched.CurrentActivity = activity
	}
}
