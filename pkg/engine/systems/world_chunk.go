// Package systems contains all ECS system implementations.
package systems

import (
	"fmt"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

// ChunkLoader defines the interface for loading and unloading chunks.
type ChunkLoader interface {
	GetChunk(x, y int) *chunk.Chunk
}

// WorldChunkSystem manages loading and unloading of world chunks.
type WorldChunkSystem struct {
	Manager   ChunkLoader
	chunkSize int
	genre     string
	// loadedBarriers tracks which chunks have had their barriers spawned.
	// Key is "chunkX:chunkY", value is slice of entity IDs created.
	loadedBarriers map[string][]ecs.Entity
}

// NewWorldChunkSystem creates a new chunk system with the given manager.
func NewWorldChunkSystem(manager ChunkLoader, chunkSize int) *WorldChunkSystem {
	return &WorldChunkSystem{
		Manager:        manager,
		chunkSize:      chunkSize,
		genre:          "fantasy",
		loadedBarriers: make(map[string][]ecs.Entity),
	}
}

// NewWorldChunkSystemWithGenre creates a chunk system with a specific genre.
func NewWorldChunkSystemWithGenre(manager ChunkLoader, chunkSize int, genre string) *WorldChunkSystem {
	return &WorldChunkSystem{
		Manager:        manager,
		chunkSize:      chunkSize,
		genre:          genre,
		loadedBarriers: make(map[string][]ecs.Entity),
	}
}

// Update loads chunks around entities with Position components.
func (s *WorldChunkSystem) Update(w *ecs.World, dt float64) {
	if s.Manager == nil {
		return
	}
	for _, e := range w.Entities("Position") {
		s.loadChunksForEntity(w, e)
	}
}

// loadChunksForEntity loads the 3x3 chunk window around a positioned entity.
func (s *WorldChunkSystem) loadChunksForEntity(w *ecs.World, e ecs.Entity) {
	comp, ok := w.GetComponent(e, "Position")
	if !ok {
		return
	}
	pos := comp.(*components.Position)
	chunkX := int(pos.X) / s.chunkSize
	chunkY := int(pos.Y) / s.chunkSize
	s.loadChunkWindow(w, chunkX, chunkY)
}

// loadChunkWindow loads the 3x3 chunk window centered on the given chunk coordinates.
func (s *WorldChunkSystem) loadChunkWindow(w *ecs.World, centerX, centerY int) {
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			chunkData := s.Manager.GetChunk(centerX+dx, centerY+dy)
			if chunkData != nil {
				s.spawnBarriersForChunk(w, chunkData, centerX+dx, centerY+dy)
			}
		}
	}
}

// chunkKey generates a unique key for a chunk coordinate pair.
func chunkKey(x, y int) string {
	return fmt.Sprintf("%d:%d", x, y)
}

// spawnBarriersForChunk creates barrier entities for a loaded chunk.
func (s *WorldChunkSystem) spawnBarriersForChunk(w *ecs.World, c *chunk.Chunk, chunkX, chunkY int) {
	if s.loadedBarriers == nil {
		s.loadedBarriers = make(map[string][]ecs.Entity)
	}

	key := chunkKey(chunkX, chunkY)

	// Skip if barriers already spawned for this chunk
	if _, exists := s.loadedBarriers[key]; exists {
		return
	}

	// Get barrier spawns from chunk
	barriers := c.GetBarrierSpawns()
	if len(barriers) == 0 {
		// Mark as processed even if no barriers
		s.loadedBarriers[key] = nil
		return
	}

	// Calculate world-space offset for this chunk
	worldOffsetX := float64(chunkX * s.chunkSize)
	worldOffsetY := float64(chunkY * s.chunkSize)

	// Create entities for each barrier spawn
	entities := make([]ecs.Entity, 0, len(barriers))
	for _, spawn := range barriers {
		entity := s.createBarrierEntity(w, spawn, worldOffsetX, worldOffsetY)
		if entity != 0 {
			entities = append(entities, entity)
		}
	}

	s.loadedBarriers[key] = entities
}

// createBarrierEntity creates an ECS entity for a barrier spawn.
func (s *WorldChunkSystem) createBarrierEntity(w *ecs.World, spawn chunk.DetailSpawn, offsetX, offsetY float64) ecs.Entity {
	if spawn.BarrierData == nil {
		return 0
	}

	entity := w.CreateEntity()

	// Add Position component
	position := &components.Position{
		X: offsetX + spawn.LocalX,
		Y: offsetY + spawn.LocalY,
		Z: 0, // Ground level
	}
	w.AddComponent(entity, position)

	// Create Barrier component from spawn data
	data := spawn.BarrierData
	barrier := &components.Barrier{
		Shape: components.BarrierShape{
			ShapeType:  data.ShapeType,
			Radius:     data.Radius * spawn.Scale,
			Width:      data.Width * spawn.Scale,
			Depth:      data.Depth * spawn.Scale,
			Height:     data.Height * spawn.Scale,
			SpriteKey:  data.ArchetypeID,
			MaterialID: data.MaterialID,
		},
		Genre:        s.genre,
		Destructible: data.Destructible,
		HitPoints:    data.HitPoints,
		MaxHP:        data.HitPoints,
	}
	w.AddComponent(entity, barrier)

	return entity
}

// UnloadChunk removes barrier entities for a chunk that is being unloaded.
func (s *WorldChunkSystem) UnloadChunk(w *ecs.World, chunkX, chunkY int) {
	key := chunkKey(chunkX, chunkY)
	entities, exists := s.loadedBarriers[key]
	if !exists {
		return
	}

	// Remove all barrier entities for this chunk
	for _, entity := range entities {
		w.DestroyEntity(entity)
	}

	delete(s.loadedBarriers, key)
}

// GetLoadedBarrierCount returns the total number of barrier entities currently loaded.
func (s *WorldChunkSystem) GetLoadedBarrierCount() int {
	count := 0
	for _, entities := range s.loadedBarriers {
		count += len(entities)
	}
	return count
}

// GetChunksWithBarriers returns the number of chunks that have loaded barriers.
func (s *WorldChunkSystem) GetChunksWithBarriers() int {
	return len(s.loadedBarriers)
}
