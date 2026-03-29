// Package systems contains all ECS system implementations.
package systems

import (
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
}

// NewWorldChunkSystem creates a new chunk system with the given manager.
func NewWorldChunkSystem(manager ChunkLoader, chunkSize int) *WorldChunkSystem {
	return &WorldChunkSystem{
		Manager:   manager,
		chunkSize: chunkSize,
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
	s.loadChunkWindow(chunkX, chunkY)
}

// loadChunkWindow loads the 3x3 chunk window centered on the given chunk coordinates.
func (s *WorldChunkSystem) loadChunkWindow(centerX, centerY int) {
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			_ = s.Manager.GetChunk(centerX+dx, centerY+dy)
		}
	}
}
