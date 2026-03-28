// Package chunk manages world chunk data and streaming.
package chunk

import (
	"encoding/binary"
	"hash/fnv"
	"sync"
)

// Chunk represents a single world chunk with a deterministic seed.
type Chunk struct {
	X, Y      int
	Size      int
	Seed      int64
	HeightMap []float64
}

// NewChunk creates a new chunk at the given coordinates.
func NewChunk(x, y, size int, seed int64) *Chunk {
	return &Chunk{
		X:         x,
		Y:         y,
		Size:      size,
		Seed:      seed,
		HeightMap: make([]float64, size*size),
	}
}

// ChunkManager handles loading, caching, and streaming of chunks.
type ChunkManager struct {
	ChunkSize int
	Seed      int64
	mu        sync.RWMutex
	loaded    map[[2]int]*Chunk
}

// NewChunkManager creates a new chunk manager.
func NewChunkManager(chunkSize int, seed int64) *ChunkManager {
	return &ChunkManager{
		ChunkSize: chunkSize,
		Seed:      seed,
		loaded:    make(map[[2]int]*Chunk),
	}
}

// GetChunk returns the chunk at the given coordinates, loading it if needed.
func (cm *ChunkManager) GetChunk(x, y int) *Chunk {
	key := [2]int{x, y}
	cm.mu.RLock()
	if c, ok := cm.loaded[key]; ok {
		cm.mu.RUnlock()
		return c
	}
	cm.mu.RUnlock()

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check under write lock in case another goroutine populated it.
	if c, ok := cm.loaded[key]; ok {
		return c
	}

	chunkSeed := mixChunkSeed(cm.Seed, x, y)
	c := NewChunk(x, y, cm.ChunkSize, chunkSeed)
	cm.loaded[key] = c
	return c
}

// mixChunkSeed derives a deterministic chunk seed using FNV-1a hashing.
func mixChunkSeed(baseSeed int64, x, y int) int64 {
	h := fnv.New64a()
	_ = binary.Write(h, binary.LittleEndian, baseSeed)
	_ = binary.Write(h, binary.LittleEndian, int64(x))
	_ = binary.Write(h, binary.LittleEndian, int64(y))
	return int64(h.Sum64())
}

// UnloadChunk removes a chunk from the cache.
func (cm *ChunkManager) UnloadChunk(x, y int) {
	key := [2]int{x, y}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.loaded, key)
}

// LoadedCount returns the number of chunks currently loaded.
func (cm *ChunkManager) LoadedCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.loaded)
}
