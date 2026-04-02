package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/geom"
)

// BarrierCollisionSystem handles collision detection between entities and barriers.
type BarrierCollisionSystem struct {
	// PlayerRadius is the collision radius for player entities.
	PlayerRadius float64
	// NPCRadius is the collision radius for NPC entities.
	NPCRadius float64
}

// NewBarrierCollisionSystem creates a new barrier collision system with default radii.
func NewBarrierCollisionSystem() *BarrierCollisionSystem {
	return &BarrierCollisionSystem{
		PlayerRadius: 0.25,
		NPCRadius:    0.3,
	}
}

// getMoverRadius returns the collision radius for an entity.
func (s *BarrierCollisionSystem) getMoverRadius(w *ecs.World, entity ecs.Entity) float64 {
	if _, hasPlayer := w.GetComponent(entity, "Player"); hasPlayer {
		return s.PlayerRadius
	}
	return s.NPCRadius
}

// calculateNextPosition computes the simulated next position of a mover.
func calculateNextPosition(pos *components.Position, path *components.NPCPathfinding, dt float64) (float64, float64, bool) {
	dx := path.TargetX - pos.X
	dy := path.TargetY - pos.Y
	moveLen := math.Sqrt(dx*dx + dy*dy)
	if moveLen < 0.001 {
		return 0, 0, false
	}
	speed := path.MoveSpeed
	if speed <= 0 {
		speed = 5.0
	}
	newX := pos.X + (dx/moveLen)*speed*dt
	newY := pos.Y + (dy/moveLen)*speed*dt
	return newX, newY, true
}

// checkMoverBarrierCollision checks collision between a mover and all barriers.
func (s *BarrierCollisionSystem) checkMoverBarrierCollision(w *ecs.World, barriers []ecs.Entity,
	path *components.NPCPathfinding, newX, newY, radius, dt float64,
) {
	for _, be := range barriers {
		barrierComp, bOK := w.GetComponent(be, "Barrier")
		barrierPosComp, bpOK := w.GetComponent(be, "Position")
		if !bOK || !bpOK {
			continue
		}

		barrier := barrierComp.(*components.Barrier)
		barrierPos := barrierPosComp.(*components.Position)

		if barrier.IsDestroyed() {
			continue
		}

		if s.CheckCollision(newX, newY, radius, barrier, barrierPos) {
			path.StuckTime += dt
		}
	}
}

// Update processes collision between moving entities and barriers.
// This is a simplified version that checks positions rather than velocities.
func (s *BarrierCollisionSystem) Update(w *ecs.World, dt float64) {
	barriers := w.Entities("Barrier", "Position")
	movers := w.Entities("Position", "NPCPathfinding")

	for _, me := range movers {
		pathComp, pOK := w.GetComponent(me, "NPCPathfinding")
		posComp, posOK := w.GetComponent(me, "Position")
		if !pOK || !posOK {
			continue
		}

		path := pathComp.(*components.NPCPathfinding)
		pos := posComp.(*components.Position)

		if !path.IsMoving || !path.HasTarget {
			continue
		}

		radius := s.getMoverRadius(w, me)
		newX, newY, ok := calculateNextPosition(pos, path, dt)
		if !ok {
			continue
		}

		s.checkMoverBarrierCollision(w, barriers, path, newX, newY, radius, dt)
	}
}

// CheckCollision tests if a point (with radius) collides with a barrier.
func (s *BarrierCollisionSystem) CheckCollision(x, y, radius float64, barrier *components.Barrier, barrierPos *components.Position) bool {
	switch barrier.Shape.ShapeType {
	case "cylinder":
		return s.checkCylinderCollision(x, y, radius, barrier, barrierPos)
	case "box":
		return s.checkBoxCollision(x, y, radius, barrier, barrierPos)
	case "polygon":
		return s.checkPolygonCollision(x, y, radius, barrier, barrierPos)
	default:
		// Default to cylinder collision for unknown shapes
		return s.checkCylinderCollision(x, y, radius, barrier, barrierPos)
	}
}

// checkCylinderCollision tests circle-circle collision.
func (s *BarrierCollisionSystem) checkCylinderCollision(x, y, radius float64, barrier *components.Barrier, barrierPos *components.Position) bool {
	dx := x - barrierPos.X
	dy := y - barrierPos.Y
	distSq := dx*dx + dy*dy
	combinedRadius := radius + barrier.Shape.Radius
	return distSq < combinedRadius*combinedRadius
}

// checkBoxCollision tests circle-AABB collision.
func (s *BarrierCollisionSystem) checkBoxCollision(x, y, radius float64, barrier *components.Barrier, barrierPos *components.Position) bool {
	halfW := barrier.Shape.Width / 2
	halfD := barrier.Shape.Depth / 2

	// Find closest point on box to circle center
	closestX := barrierClamp(x, barrierPos.X-halfW, barrierPos.X+halfW)
	closestY := barrierClamp(y, barrierPos.Y-halfD, barrierPos.Y+halfD)

	// Check if closest point is within circle radius
	dx := x - closestX
	dy := y - closestY
	distSq := dx*dx + dy*dy

	return distSq < radius*radius
}

// checkPolygonCollision tests circle-polygon collision using SAT.
func (s *BarrierCollisionSystem) checkPolygonCollision(x, y, radius float64, barrier *components.Barrier, barrierPos *components.Position) bool {
	vertices := barrier.Shape.Vertices
	if len(vertices) < 6 {
		// Need at least 3 vertices (6 floats)
		return false
	}

	// Transform vertices to world space
	worldVerts := make([]float64, len(vertices))
	for i := 0; i < len(vertices); i += 2 {
		worldVerts[i] = vertices[i] + barrierPos.X
		if i+1 < len(vertices) {
			worldVerts[i+1] = vertices[i+1] + barrierPos.Y
		}
	}

	// Check if point is inside polygon
	if s.pointInPolygon(x, y, worldVerts) {
		return true
	}

	// Check if circle intersects any polygon edge
	numVerts := len(worldVerts) / 2
	for i := 0; i < numVerts; i++ {
		x1, y1 := worldVerts[i*2], worldVerts[i*2+1]
		j := (i + 1) % numVerts
		x2, y2 := worldVerts[j*2], worldVerts[j*2+1]

		// Distance from circle center to line segment
		dist := s.pointToSegmentDistance(x, y, x1, y1, x2, y2)
		if dist < radius {
			return true
		}
	}

	return false
}

// pointInPolygon tests if a point is inside a polygon using ray casting.
func (s *BarrierCollisionSystem) pointInPolygon(x, y float64, vertices []float64) bool {
	return geom.PointInPolygon(x, y, vertices)
}

// pointToSegmentDistance calculates the shortest distance from a point to a line segment.
func (s *BarrierCollisionSystem) pointToSegmentDistance(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	lengthSq := dx*dx + dy*dy

	if lengthSq == 0 {
		// Segment is a point
		return math.Sqrt((px-x1)*(px-x1) + (py-y1)*(py-y1))
	}

	// Project point onto line, clamping to segment
	t := barrierClamp(((px-x1)*dx+(py-y1)*dy)/lengthSq, 0, 1)

	// Find closest point on segment
	closestX := x1 + t*dx
	closestY := y1 + t*dy

	// Distance to closest point
	return math.Sqrt((px-closestX)*(px-closestX) + (py-closestY)*(py-closestY))
}

// ResolveCollision calculates the resolved position after collision.
func (s *BarrierCollisionSystem) ResolveCollision(startX, startY, endX, endY, radius float64, barrier *components.Barrier, barrierPos *components.Position) (float64, float64) {
	switch barrier.Shape.ShapeType {
	case "cylinder":
		return s.resolveCylinderCollision(startX, startY, endX, endY, radius, barrier, barrierPos)
	case "box":
		return s.resolveBoxCollision(startX, startY, endX, endY, radius, barrier, barrierPos)
	case "polygon":
		return s.resolvePolygonCollision(startX, startY, endX, endY, radius, barrier, barrierPos)
	default:
		return s.resolveCylinderCollision(startX, startY, endX, endY, radius, barrier, barrierPos)
	}
}

// resolveCylinderCollision resolves circle-circle collision by sliding along the edge.
func (s *BarrierCollisionSystem) resolveCylinderCollision(startX, startY, endX, endY, radius float64, barrier *components.Barrier, barrierPos *components.Position) (float64, float64) {
	// Direction from barrier to entity
	dx := endX - barrierPos.X
	dy := endY - barrierPos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist == 0 {
		// Entity at center of barrier, push away
		return startX, startY
	}

	// Push entity to edge of combined radius
	combinedRadius := radius + barrier.Shape.Radius + 0.001 // Small epsilon
	resolvedX := barrierPos.X + (dx/dist)*combinedRadius
	resolvedY := barrierPos.Y + (dy/dist)*combinedRadius

	return resolvedX, resolvedY
}

// resolveBoxCollision resolves circle-AABB collision.
func (s *BarrierCollisionSystem) resolveBoxCollision(startX, startY, endX, endY, radius float64, barrier *components.Barrier, barrierPos *components.Position) (float64, float64) {
	halfW := barrier.Shape.Width / 2
	halfD := barrier.Shape.Depth / 2

	// Find closest point on box to circle center
	closestX := barrierClamp(endX, barrierPos.X-halfW, barrierPos.X+halfW)
	closestY := barrierClamp(endY, barrierPos.Y-halfD, barrierPos.Y+halfD)

	// Direction from closest point to circle center
	dx := endX - closestX
	dy := endY - closestY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist == 0 {
		// Entity inside box, push to nearest edge
		return startX, startY
	}

	// Push entity away from closest point
	pushDist := radius + 0.001 - dist
	if pushDist > 0 {
		return endX + (dx/dist)*pushDist, endY + (dy/dist)*pushDist
	}

	return endX, endY
}

// resolvePolygonCollision resolves circle-polygon collision.
func (s *BarrierCollisionSystem) resolvePolygonCollision(startX, startY, endX, endY, radius float64, barrier *components.Barrier, barrierPos *components.Position) (float64, float64) {
	vertices := barrier.Shape.Vertices
	if len(vertices) < 6 {
		return endX, endY
	}

	// Transform vertices to world space
	worldVerts := make([]float64, len(vertices))
	for i := 0; i < len(vertices); i += 2 {
		worldVerts[i] = vertices[i] + barrierPos.X
		if i+1 < len(vertices) {
			worldVerts[i+1] = vertices[i+1] + barrierPos.Y
		}
	}

	// Find closest point on polygon boundary
	minDist := math.MaxFloat64
	var normalX, normalY float64

	numVerts := len(worldVerts) / 2
	for i := 0; i < numVerts; i++ {
		x1, y1 := worldVerts[i*2], worldVerts[i*2+1]
		j := (i + 1) % numVerts
		x2, y2 := worldVerts[j*2], worldVerts[j*2+1]

		// Find closest point on this edge
		closest := s.closestPointOnSegment(endX, endY, x1, y1, x2, y2)
		dist := math.Sqrt((endX-closest[0])*(endX-closest[0]) + (endY-closest[1])*(endY-closest[1]))

		if dist < minDist {
			minDist = dist

			// Calculate outward normal
			edgeDx, edgeDy := x2-x1, y2-y1
			normalX, normalY = -edgeDy, edgeDx
			normalLen := math.Sqrt(normalX*normalX + normalY*normalY)
			if normalLen > 0 {
				normalX, normalY = normalX/normalLen, normalY/normalLen
			}
		}
	}

	if minDist < radius {
		// Push entity away along normal
		pushDist := radius + 0.001 - minDist
		return endX + normalX*pushDist, endY + normalY*pushDist
	}

	return endX, endY
}

// closestPointOnSegment finds the closest point on a segment to a given point.
func (s *BarrierCollisionSystem) closestPointOnSegment(px, py, x1, y1, x2, y2 float64) [2]float64 {
	dx := x2 - x1
	dy := y2 - y1
	lengthSq := dx*dx + dy*dy

	if lengthSq == 0 {
		return [2]float64{x1, y1}
	}

	t := barrierClamp(((px-x1)*dx+(py-y1)*dy)/lengthSq, 0, 1)
	return [2]float64{x1 + t*dx, y1 + t*dy}
}

// barrierClamp clamps a float64 to [min, max].
// Named to avoid conflict with clampFloat in economy.go
func barrierClamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

// BarrierDamageSystem handles damage to destructible barriers.
type BarrierDamageSystem struct{}

// NewBarrierDamageSystem creates a new barrier damage system.
func NewBarrierDamageSystem() *BarrierDamageSystem {
	return &BarrierDamageSystem{}
}

// Update processes damage to barriers (from projectiles, melee, etc.).
func (s *BarrierDamageSystem) Update(w *ecs.World, dt float64) {
	// This system is event-driven via DamageBarrier, not tick-based
}

// DamageBarrier applies damage to a barrier and returns true if destroyed.
// Updates the barrier's Appearance.DamageOverlay based on current HP percentage.
func (s *BarrierDamageSystem) DamageBarrier(w *ecs.World, barrierEntity ecs.Entity, damage float64) bool {
	barrierComp, ok := w.GetComponent(barrierEntity, "Barrier")
	if !ok {
		return false
	}

	barrier := barrierComp.(*components.Barrier)
	if !barrier.Destructible {
		return false
	}

	barrier.HitPoints -= damage
	if barrier.HitPoints < 0 {
		barrier.HitPoints = 0
	}

	// Update damage overlay on appearance
	if appComp, hasApp := w.GetComponent(barrierEntity, "Appearance"); hasApp {
		appearance := appComp.(*components.Appearance)
		// DamageOverlay: 0.0 = pristine, 1.0 = heavily damaged
		if barrier.MaxHP > 0 {
			appearance.DamageOverlay = 1.0 - (barrier.HitPoints / barrier.MaxHP)
		}
		// Switch to damaged sprite variant at 50% HP
		if barrier.HitPoints <= barrier.MaxHP*0.5 && barrier.HitPoints > 0 {
			if appearance.AnimState != "damaged" {
				appearance.AnimState = "damaged"
			}
		}
	}

	return barrier.IsDestroyed()
}

// ============================================================
// BarrierDestructionSystem
// ============================================================

// BarrierDestructionSystem processes destroyed barriers, spawns debris particles,
// and removes barrier entities from the world. This system runs after damage is
// applied (e.g., by CombatSystem) and handles the visual/audio feedback of destruction.
type BarrierDestructionSystem struct {
	// DebrisParticlesPerBarrier is the number of debris particles to spawn per destroyed barrier.
	DebrisParticlesPerBarrier int
	// DebrisVelocityMax is the maximum velocity for debris particles.
	DebrisVelocityMax float64
	// DebrisLifetime is how long debris particles exist before despawning.
	DebrisLifetime float64
	// RemoveDestroyedEntities determines whether to remove entities or just mark them.
	RemoveDestroyedEntities bool
	// DestructionSound is the default sound to play on destruction.
	DestructionSound string
	// pendingRemovals tracks entities to remove after the update loop.
	pendingRemovals []ecs.Entity
	// NewlyDestroyed is the list of barriers destroyed this frame (for other systems).
	NewlyDestroyed []ecs.Entity
}

// NewBarrierDestructionSystem creates a new barrier destruction system with defaults.
func NewBarrierDestructionSystem() *BarrierDestructionSystem {
	return &BarrierDestructionSystem{
		DebrisParticlesPerBarrier: 8,
		DebrisVelocityMax:         5.0,
		DebrisLifetime:            2.0,
		RemoveDestroyedEntities:   true,
		DestructionSound:          "destruction_debris",
		pendingRemovals:           make([]ecs.Entity, 0),
		NewlyDestroyed:            make([]ecs.Entity, 0),
	}
}

// Update processes all barriers and handles destruction for those with zero health.
func (s *BarrierDestructionSystem) Update(w *ecs.World, dt float64) {
	// Clear previous frame's data
	s.pendingRemovals = s.pendingRemovals[:0]
	s.NewlyDestroyed = s.NewlyDestroyed[:0]

	barriers := w.Entities("Barrier", "Position")

	for _, e := range barriers {
		barrierComp, ok := w.GetComponent(e, "Barrier")
		if !ok {
			continue
		}
		barrier := barrierComp.(*components.Barrier)

		// Skip non-destructible or already processed barriers
		if !barrier.Destructible {
			continue
		}

		// Check if barrier should be destroyed
		if barrier.IsDestroyed() && !barrier.DestructionProcessed {
			s.processDestruction(w, e, barrier)
		}
	}

	// Remove entities after iteration to avoid modifying collection during loop
	if s.RemoveDestroyedEntities {
		for _, e := range s.pendingRemovals {
			w.DestroyEntity(e)
		}
	}
}

// processDestruction handles the destruction of a single barrier.
func (s *BarrierDestructionSystem) processDestruction(w *ecs.World, entity ecs.Entity, barrier *components.Barrier) {
	// Mark as processed to prevent duplicate processing
	barrier.DestructionProcessed = true

	// Get position for particle spawning
	posComp, ok := w.GetComponent(entity, "Position")
	if !ok {
		// No position, just mark for removal
		s.pendingRemovals = append(s.pendingRemovals, entity)
		s.NewlyDestroyed = append(s.NewlyDestroyed, entity)
		return
	}
	pos := posComp.(*components.Position)

	// Spawn debris particles
	s.spawnDebrisParticles(w, pos, barrier)

	// Trigger destruction sound (via audio system event)
	s.triggerDestructionSound(w, entity, pos, barrier)

	// Track for removal and notification
	s.pendingRemovals = append(s.pendingRemovals, entity)
	s.NewlyDestroyed = append(s.NewlyDestroyed, entity)
}

// spawnDebrisParticles creates debris particle entities at the barrier location.
func (s *BarrierDestructionSystem) spawnDebrisParticles(w *ecs.World, pos *components.Position, barrier *components.Barrier) {
	// Use deterministic seed based on barrier position for consistent debris
	seed := int64(pos.X*1000) ^ int64(pos.Y*1000) ^ int64(pos.Z*1000)
	rng := NewPseudoRandom(seed)

	// Determine debris color based on barrier material
	debrisColor := s.getDebrisColor(barrier)

	for i := 0; i < s.DebrisParticlesPerBarrier; i++ {
		// Create debris entity
		debris := w.CreateEntity()

		// Random velocity in all directions
		vx := (rng.Float64()*2 - 1) * s.DebrisVelocityMax
		vy := (rng.Float64()*2 - 1) * s.DebrisVelocityMax
		vz := rng.Float64() * s.DebrisVelocityMax * 0.5 // Bias upward

		// Add position component (slightly offset from center)
		offsetX := (rng.Float64()*2 - 1) * 0.3
		offsetY := (rng.Float64()*2 - 1) * 0.3
		offsetZ := rng.Float64() * 0.3
		w.AddComponent(debris, &components.Position{
			X: pos.X + offsetX,
			Y: pos.Y + offsetY,
			Z: pos.Z + offsetZ,
		})

		// Add physics for debris movement
		w.AddComponent(debris, &components.PhysicsBody{
			VelocityX:       vx,
			VelocityY:       vy,
			VelocityZ:       vz,
			Mass:            0.5,
			Friction:        0.8,
			Bounciness:      0.3,
			CollisionRadius: 0.1,
		})

		// Add particle component for rendering and lifetime
		w.AddComponent(debris, &components.Particle{
			Color:      debrisColor,
			Size:       0.1 + rng.Float64()*0.1,
			Lifetime:   s.DebrisLifetime,
			Age:        0,
			FadeOut:    true,
			ParticleID: "debris",
		})
	}
}

// getDebrisColor returns an appropriate debris color based on barrier material.
func (s *BarrierDestructionSystem) getDebrisColor(barrier *components.Barrier) [4]uint8 {
	switch barrier.MaterialType {
	case "wood":
		return [4]uint8{139, 90, 43, 255} // Brown
	case "stone":
		return [4]uint8{128, 128, 128, 255} // Gray
	case "metal":
		return [4]uint8{100, 100, 100, 255} // Dark gray
	case "glass":
		return [4]uint8{200, 230, 255, 200} // Light blue, semi-transparent
	case "ice":
		return [4]uint8{200, 240, 255, 220} // Cyan-tinted white
	default:
		return [4]uint8{128, 128, 128, 255} // Default gray
	}
}

// triggerDestructionSound notifies the audio system to play destruction sound.
func (s *BarrierDestructionSystem) triggerDestructionSound(w *ecs.World, entity ecs.Entity, pos *components.Position, barrier *components.Barrier) {
	// Create a sound event entity that the audio system will pick up
	soundEntity := w.CreateEntity()

	soundName := s.DestructionSound
	switch barrier.MaterialType {
	case "wood":
		soundName = "destruction_wood"
	case "stone":
		soundName = "destruction_stone"
	case "metal":
		soundName = "destruction_metal"
	case "glass":
		soundName = "destruction_glass"
	}

	w.AddComponent(soundEntity, &components.SoundEvent{
		SoundID:   soundName,
		X:         pos.X,
		Y:         pos.Y,
		Z:         pos.Z,
		Volume:    1.0,
		Radius:    20.0,
		OneShot:   true,
		Processed: false,
	})
}

// GetNewlyDestroyed returns the list of barriers destroyed this frame.
// Other systems can use this to react to destruction events.
func (s *BarrierDestructionSystem) GetNewlyDestroyed() []ecs.Entity {
	return s.NewlyDestroyed
}
