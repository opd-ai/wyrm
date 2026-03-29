// Package chunk manages world chunk data and streaming.
package chunk

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
	"sync"

	"github.com/opd-ai/wyrm/pkg/procgen/noise"
)

// Chunk represents a single world chunk with a deterministic seed.
type Chunk struct {
	X, Y      int
	Size      int
	Seed      int64
	HeightMap []float64
}

// NewChunk creates a new chunk at the given coordinates with generated terrain.
func NewChunk(x, y, size int, seed int64) *Chunk {
	heightMap := generateHeightMap(size, seed)
	return &Chunk{
		X:         x,
		Y:         y,
		Size:      size,
		Seed:      seed,
		HeightMap: heightMap,
	}
}

// generateHeightMap creates a procedural heightmap using noise.
func generateHeightMap(size int, seed int64) []float64 {
	heightMap := make([]float64, size*size)
	rng := rand.New(rand.NewSource(seed))

	// Generate base offset for this chunk
	offsetX := rng.Float64() * 1000
	offsetY := rng.Float64() * 1000

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Multi-octave noise for terrain
			nx := (float64(x) + offsetX) * 0.01
			ny := (float64(y) + offsetY) * 0.01

			height := 0.0
			amplitude := 1.0
			frequency := 1.0

			// 4 octaves of noise
			for oct := 0; oct < 4; oct++ {
				height += noise.Noise2DSigned(nx*frequency, ny*frequency, seed+int64(oct)) * amplitude
				amplitude *= 0.5
				frequency *= 2.0
			}

			// Normalize to [0, 1]
			height = (height + 1.0) / 2.0
			heightMap[y*size+x] = height
		}
	}

	return heightMap
}

// GetHeight returns the height at the given local coordinates.
func (c *Chunk) GetHeight(localX, localY int) float64 {
	if localX < 0 || localX >= c.Size || localY < 0 || localY >= c.Size {
		return 0
	}
	return c.HeightMap[localY*c.Size+localX]
}

// Manager handles loading, caching, and streaming of chunks.
type Manager struct {
	ChunkSize int
	Seed      int64
	mu        sync.RWMutex
	loaded    map[[2]int]*Chunk
}

// NewManager creates a new chunk manager.
func NewManager(chunkSize int, seed int64) *Manager {
	return &Manager{
		ChunkSize: chunkSize,
		Seed:      seed,
		loaded:    make(map[[2]int]*Chunk),
	}
}

// GetChunk returns the chunk at the given coordinates, loading it if needed.
func (cm *Manager) GetChunk(x, y int) *Chunk {
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
func (cm *Manager) UnloadChunk(x, y int) {
	key := [2]int{x, y}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.loaded, key)
}

// LoadedCount returns the number of chunks currently loaded.
func (cm *Manager) LoadedCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.loaded)
}
