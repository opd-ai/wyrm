package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// NPCPathfindingSystem moves NPCs toward their scheduled activity locations.
type NPCPathfindingSystem struct {
	// DefaultMoveSpeed is used when an NPC has no specified speed.
	DefaultMoveSpeed float64
	// DefaultArrivalThreshold is the distance at which NPCs stop moving.
	DefaultArrivalThreshold float64
}

// NewNPCPathfindingSystem creates a new pathfinding system.
func NewNPCPathfindingSystem() *NPCPathfindingSystem {
	return &NPCPathfindingSystem{
		DefaultMoveSpeed:        NPCDefaultMoveSpeed,
		DefaultArrivalThreshold: NPCDefaultArrivalThreshold,
	}
}

// Update processes NPC movement toward destinations each tick.
func (s *NPCPathfindingSystem) Update(w *ecs.World, dt float64) {
	s.updateScheduleTargets(w)
	s.moveNPCs(w, dt)
}

// updateScheduleTargets sets pathfinding targets based on current activity.
func (s *NPCPathfindingSystem) updateScheduleTargets(w *ecs.World) {
	for _, e := range w.Entities("Schedule", "NPCPathfinding") {
		schedComp, ok := w.GetComponent(e, "Schedule")
		if !ok {
			continue
		}
		pathComp, ok := w.GetComponent(e, "NPCPathfinding")
		if !ok {
			continue
		}

		sched := schedComp.(*components.Schedule)
		path := pathComp.(*components.NPCPathfinding)

		s.updateTargetForActivity(path, sched.CurrentActivity)
	}
}

// updateTargetForActivity sets the pathfinding target based on activity.
func (s *NPCPathfindingSystem) updateTargetForActivity(path *components.NPCPathfinding, activity string) {
	if path.ActivityLocations == nil {
		return
	}

	loc, exists := path.ActivityLocations[activity]
	if !exists {
		return
	}

	// Only update target if it changed
	if path.HasTarget && path.TargetX == loc.X && path.TargetY == loc.Y {
		return
	}

	path.TargetX = loc.X
	path.TargetY = loc.Y
	path.HasTarget = true
	path.IsMoving = true
	path.StuckTime = 0

	// Generate simple path (direct line for now)
	path.CurrentPath = []components.Waypoint{{X: loc.X, Y: loc.Y}}
	path.CurrentWaypointIndex = 0
}

// moveNPCs moves all NPCs toward their targets.
func (s *NPCPathfindingSystem) moveNPCs(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Position", "NPCPathfinding") {
		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pathComp, ok := w.GetComponent(e, "NPCPathfinding")
		if !ok {
			continue
		}

		pos := posComp.(*components.Position)
		path := pathComp.(*components.NPCPathfinding)

		s.moveTowardTarget(pos, path, dt)
	}
}

// moveTowardTarget moves an NPC toward its current waypoint.
func (s *NPCPathfindingSystem) moveTowardTarget(pos *components.Position, path *components.NPCPathfinding, dt float64) {
	if !path.HasTarget || !path.IsMoving {
		return
	}

	waypoint, ok := s.getCurrentWaypoint(path)
	if !ok {
		path.IsMoving = false
		return
	}

	dx, dy, dist := calculateDistance(waypoint.X, waypoint.Y, pos.X, pos.Y)
	threshold := s.getArrivalThreshold(path)

	if s.checkArrival(path, dist, threshold) {
		return
	}

	moveAmount := s.calculateMovement(path, dist, dt)
	s.applyMovement(pos, dx, dy, dist, moveAmount)
	s.updateStuckDetection(path, moveAmount, dt)
}

// getCurrentWaypoint returns the current waypoint if valid.
func (s *NPCPathfindingSystem) getCurrentWaypoint(path *components.NPCPathfinding) (components.Waypoint, bool) {
	if len(path.CurrentPath) == 0 || path.CurrentWaypointIndex >= len(path.CurrentPath) {
		return components.Waypoint{}, false
	}
	return path.CurrentPath[path.CurrentWaypointIndex], true
}

// calculateDistance computes the delta and distance between two points.
func calculateDistance(targetX, targetY, posX, posY float64) (dx, dy, dist float64) {
	dx = targetX - posX
	dy = targetY - posY
	dist = math.Sqrt(dx*dx + dy*dy)
	return dx, dy, dist
}

// getArrivalThreshold returns the threshold for waypoint arrival.
func (s *NPCPathfindingSystem) getArrivalThreshold(path *components.NPCPathfinding) float64 {
	if path.ArrivalThreshold > 0 {
		return path.ArrivalThreshold
	}
	return s.DefaultArrivalThreshold
}

// checkArrival handles waypoint arrival and path completion.
func (s *NPCPathfindingSystem) checkArrival(path *components.NPCPathfinding, dist, threshold float64) bool {
	if dist > threshold {
		return false
	}
	path.CurrentWaypointIndex++
	if path.CurrentWaypointIndex >= len(path.CurrentPath) {
		path.IsMoving = false
		path.HasTarget = false
	}
	return true
}

// calculateMovement determines how far to move this frame.
func (s *NPCPathfindingSystem) calculateMovement(path *components.NPCPathfinding, dist, dt float64) float64 {
	speed := path.MoveSpeed
	if speed <= 0 {
		speed = s.DefaultMoveSpeed
	}
	moveAmount := speed * dt
	if moveAmount > dist {
		moveAmount = dist
	}
	return moveAmount
}

// applyMovement updates position and facing angle.
func (s *NPCPathfindingSystem) applyMovement(pos *components.Position, dx, dy, dist, moveAmount float64) {
	if dist > 0 {
		pos.X += (dx / dist) * moveAmount
		pos.Y += (dy / dist) * moveAmount
		pos.Angle = math.Atan2(dy, dx)
	}
}

// updateStuckDetection tracks if NPC is stuck and handles timeout.
func (s *NPCPathfindingSystem) updateStuckDetection(path *components.NPCPathfinding, moveAmount, dt float64) {
	if moveAmount < NPCMinMovementThreshold {
		path.StuckTime += dt
		maxStuck := path.MaxStuckTime
		if maxStuck <= 0 {
			maxStuck = NPCDefaultMaxStuckTime
		}
		if path.StuckTime > maxStuck {
			path.IsMoving = false
			path.StuckTime = 0
		}
	} else {
		path.StuckTime = 0
	}
}

// SetActivityLocation sets the location for a specific activity.
func SetActivityLocation(path *components.NPCPathfinding, activity string, x, y float64, locationID string) {
	if path.ActivityLocations == nil {
		path.ActivityLocations = make(map[string]components.ActivityLocation)
	}
	path.ActivityLocations[activity] = components.ActivityLocation{
		X:          x,
		Y:          y,
		LocationID: locationID,
	}
}

// GetDistanceToTarget returns the distance to the current target.
func GetDistanceToTarget(pos *components.Position, path *components.NPCPathfinding) float64 {
	if !path.HasTarget {
		return 0
	}
	dx := path.TargetX - pos.X
	dy := path.TargetY - pos.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// IsAtDestination checks if the NPC has reached their destination.
func IsAtDestination(pos *components.Position, path *components.NPCPathfinding, threshold float64) bool {
	if !path.HasTarget {
		return true
	}
	return GetDistanceToTarget(pos, path) <= threshold
}

// ClearPath stops the NPC and clears their current path.
func ClearPath(path *components.NPCPathfinding) {
	path.HasTarget = false
	path.IsMoving = false
	path.CurrentPath = nil
	path.CurrentWaypointIndex = 0
	path.StuckTime = 0
}

// SetDirectTarget sets a direct target position without using activity locations.
func SetDirectTarget(path *components.NPCPathfinding, x, y float64) {
	path.TargetX = x
	path.TargetY = y
	path.HasTarget = true
	path.IsMoving = true
	path.StuckTime = 0
	path.CurrentPath = []components.Waypoint{{X: x, Y: y}}
	path.CurrentWaypointIndex = 0
}

// GenerateScheduleLocations creates typical activity locations for an NPC.
func GenerateScheduleLocations(path *components.NPCPathfinding, homeX, homeY, workX, workY float64) {
	if path.ActivityLocations == nil {
		path.ActivityLocations = make(map[string]components.ActivityLocation)
	}

	// Common activity mappings
	path.ActivityLocations["sleeping"] = components.ActivityLocation{X: homeX, Y: homeY, LocationID: "home"}
	path.ActivityLocations["resting"] = components.ActivityLocation{X: homeX, Y: homeY, LocationID: "home"}
	path.ActivityLocations["eating"] = components.ActivityLocation{X: homeX, Y: homeY, LocationID: "home"}
	path.ActivityLocations["working"] = components.ActivityLocation{X: workX, Y: workY, LocationID: "work"}
	path.ActivityLocations["crafting"] = components.ActivityLocation{X: workX, Y: workY, LocationID: "work"}
	path.ActivityLocations["trading"] = components.ActivityLocation{X: workX, Y: workY, LocationID: "work"}
}
