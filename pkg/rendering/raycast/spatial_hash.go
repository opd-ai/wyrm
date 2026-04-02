// Package raycast provides the first-person raycasting renderer.
// This file implements a spatial hash for efficient sprite queries.
package raycast

import (
	"sync"
)

// SpatialHash is a grid-based spatial partitioning structure for sprites.
// It allows O(1) lookups for sprites in a given cell and efficient
// range queries for sprites near a position.
type SpatialHash struct {
	// CellSize is the width/height of each grid cell in world units.
	CellSize float64
	// cells maps grid coordinates to sprites in that cell.
	// Key format: (x << 16) | (y & 0xFFFF) for efficient hashing.
	cells map[int64][]*SpriteEntity
	// entityCells tracks which cells each entity is in.
	// Used for efficient removal/update.
	entityCells map[*SpriteEntity]int64
	mu          sync.RWMutex
}

// NewSpatialHash creates a new spatial hash with the given cell size.
// Cell size should be roughly the size of the largest objects to ensure
// efficient queries. A cell size of 5-10 world units is typical.
func NewSpatialHash(cellSize float64) *SpatialHash {
	if cellSize <= 0 {
		cellSize = 5.0 // Default cell size
	}
	return &SpatialHash{
		CellSize:    cellSize,
		cells:       make(map[int64][]*SpriteEntity),
		entityCells: make(map[*SpriteEntity]int64),
	}
}

// cellKey computes the hash key for a cell at (cx, cy).
func cellKey(cx, cy int) int64 {
	return int64(cx)<<16 | int64(uint16(cy))
}

// worldToCell converts world coordinates to cell coordinates.
func (sh *SpatialHash) worldToCell(x, y float64) (int, int) {
	cx := int(x / sh.CellSize)
	cy := int(y / sh.CellSize)
	return cx, cy
}

// Insert adds an entity to the spatial hash.
// If the entity is already in the hash, it is first removed.
func (sh *SpatialHash) Insert(entity *SpriteEntity) {
	if entity == nil {
		return
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	// Remove from old cell if present
	if oldKey, exists := sh.entityCells[entity]; exists {
		sh.removeFromCell(entity, oldKey)
	}

	// Add to new cell
	cx, cy := sh.worldToCell(entity.X, entity.Y)
	key := cellKey(cx, cy)
	sh.cells[key] = append(sh.cells[key], entity)
	sh.entityCells[entity] = key
}

// Remove removes an entity from the spatial hash.
func (sh *SpatialHash) Remove(entity *SpriteEntity) {
	if entity == nil {
		return
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	if key, exists := sh.entityCells[entity]; exists {
		sh.removeFromCell(entity, key)
		delete(sh.entityCells, entity)
	}
}

// removeFromCell removes an entity from a specific cell (internal, must hold lock).
func (sh *SpatialHash) removeFromCell(entity *SpriteEntity, key int64) {
	cell := sh.cells[key]
	for i, e := range cell {
		if e == entity {
			// Swap with last and truncate
			cell[i] = cell[len(cell)-1]
			sh.cells[key] = cell[:len(cell)-1]
			break
		}
	}
}

// Update updates an entity's position in the spatial hash.
// This is more efficient than Remove + Insert when the entity
// might be in the same cell.
func (sh *SpatialHash) Update(entity *SpriteEntity) {
	if entity == nil {
		return
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	newCX, newCY := sh.worldToCell(entity.X, entity.Y)
	newKey := cellKey(newCX, newCY)

	if oldKey, exists := sh.entityCells[entity]; exists {
		if oldKey == newKey {
			// Same cell, no update needed
			return
		}
		sh.removeFromCell(entity, oldKey)
	}

	sh.cells[newKey] = append(sh.cells[newKey], entity)
	sh.entityCells[entity] = newKey
}

// Query returns all entities in the cell containing (x, y).
func (sh *SpatialHash) Query(x, y float64) []*SpriteEntity {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	cx, cy := sh.worldToCell(x, y)
	key := cellKey(cx, cy)
	return sh.cells[key]
}

// QueryRadius returns all entities within radius of (x, y).
// This checks all cells that could contain entities within the radius.
func (sh *SpatialHash) QueryRadius(x, y, radius float64) []*SpriteEntity {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.queryRadiusInternal(x, y, radius)
}

// QueryRect returns all entities within the rectangle from (minX, minY) to (maxX, maxY).
func (sh *SpatialHash) QueryRect(minX, minY, maxX, maxY float64) []*SpriteEntity {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	// Calculate cell range to check
	minCX, minCY := sh.worldToCell(minX, minY)
	maxCX, maxCY := sh.worldToCell(maxX, maxY)

	var result []*SpriteEntity

	for cx := minCX; cx <= maxCX; cx++ {
		for cy := minCY; cy <= maxCY; cy++ {
			key := cellKey(cx, cy)
			for _, e := range sh.cells[key] {
				// Bounds check
				if e.X >= minX && e.X <= maxX && e.Y >= minY && e.Y <= maxY {
					result = append(result, e)
				}
			}
		}
	}

	return result
}

// QueryFrustum returns all entities potentially visible from (px, py) facing angle
// within the field of view and max distance. This is a conservative check -
// it may include some entities that are actually outside the frustum.
func (sh *SpatialHash) QueryFrustum(px, py, angle, fov, maxDist float64) []*SpriteEntity {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	// For a simple approximation, we query a rectangle that bounds the frustum
	// The frustum width at maxDist is approximately 2 * maxDist * tan(fov/2)
	// We use a circular query for simplicity (conservative)
	return sh.queryRadiusInternal(px, py, maxDist)
}

// queryRadiusInternal is the internal radius query (must hold read lock).
func (sh *SpatialHash) queryRadiusInternal(x, y, radius float64) []*SpriteEntity {
	minCX, minCY := sh.worldToCell(x-radius, y-radius)
	maxCX, maxCY := sh.worldToCell(x+radius, y+radius)

	radiusSq := radius * radius
	var result []*SpriteEntity

	for cx := minCX; cx <= maxCX; cx++ {
		for cy := minCY; cy <= maxCY; cy++ {
			key := cellKey(cx, cy)
			for _, e := range sh.cells[key] {
				dx := e.X - x
				dy := e.Y - y
				if dx*dx+dy*dy <= radiusSq {
					result = append(result, e)
				}
			}
		}
	}

	return result
}

// Clear removes all entities from the spatial hash.
func (sh *SpatialHash) Clear() {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.cells = make(map[int64][]*SpriteEntity)
	sh.entityCells = make(map[*SpriteEntity]int64)
}

// Count returns the total number of entities in the spatial hash.
func (sh *SpatialHash) Count() int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	return len(sh.entityCells)
}

// CellCount returns the number of occupied cells.
func (sh *SpatialHash) CellCount() int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	return len(sh.cells)
}

// InsertAll adds multiple entities to the spatial hash.
func (sh *SpatialHash) InsertAll(entities []*SpriteEntity) {
	for _, e := range entities {
		sh.Insert(e)
	}
}

// NearestEntity finds the entity nearest to (x, y) within maxDist.
// Returns nil if no entity is found within the distance.
func (sh *SpatialHash) NearestEntity(x, y, maxDist float64) *SpriteEntity {
	entities := sh.QueryRadius(x, y, maxDist)

	var nearest *SpriteEntity
	nearestDistSq := maxDist * maxDist

	for _, e := range entities {
		dx := e.X - x
		dy := e.Y - y
		distSq := dx*dx + dy*dy
		if distSq < nearestDistSq {
			nearestDistSq = distSq
			nearest = e
		}
	}

	return nearest
}

// NearestInteractable finds the nearest interactable entity to (x, y) within maxDist.
func (sh *SpatialHash) NearestInteractable(x, y, maxDist float64) *SpriteEntity {
	entities := sh.QueryRadius(x, y, maxDist)

	var nearest *SpriteEntity
	nearestDistSq := maxDist * maxDist

	for _, e := range entities {
		if !e.IsInteractable || !e.Visible {
			continue
		}
		dx := e.X - x
		dy := e.Y - y
		distSq := dx*dx + dy*dy
		if distSq < nearestDistSq {
			nearestDistSq = distSq
			nearest = e
		}
	}

	return nearest
}
