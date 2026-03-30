// Package chunk manages world chunk data and streaming.
package chunk

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
	"sync"

	"github.com/opd-ai/wyrm/pkg/procgen/noise"
)

// Terrain types for vertical features.
const (
	TerrainFlat  = 0 // Normal walkable terrain
	TerrainHill  = 1 // Elevated terrain (gradual slope)
	TerrainCliff = 2 // Steep terrain (impassable edge)
	TerrainPeak  = 3 // High elevation peak
)

// Terrain elevation constants.
const (
	// HillThreshold is the height value above which terrain becomes a hill.
	HillThreshold = 0.5
	// CliffThreshold is the height difference that creates a cliff edge.
	CliffThreshold = 0.15
	// PeakThreshold is the height value above which terrain becomes a peak.
	PeakThreshold = 0.8
	// MaxElevation is the maximum terrain elevation in world units.
	MaxElevation = 10.0
	// ElevationOctaves is the number of noise octaves for elevation.
	ElevationOctaves = 5
)

// Chunk represents a single world chunk with a deterministic seed.
type Chunk struct {
	X, Y         int
	Size         int
	Seed         int64
	HeightMap    []float64
	ElevationMap []float64 // Terrain elevation in world units
	TerrainTypes []int     // Terrain type per cell
}

// NewChunk creates a new chunk at the given coordinates with generated terrain.
func NewChunk(x, y, size int, seed int64) *Chunk {
	heightMap := generateHeightMap(size, seed)
	elevationMap := generateElevationMap(size, seed, heightMap)
	terrainTypes := generateTerrainTypes(size, heightMap, elevationMap)
	return &Chunk{
		X:            x,
		Y:            y,
		Size:         size,
		Seed:         seed,
		HeightMap:    heightMap,
		ElevationMap: elevationMap,
		TerrainTypes: terrainTypes,
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

// generateElevationMap converts the heightmap to world-unit elevations.
// Uses additional octaves to create more dramatic vertical features.
func generateElevationMap(size int, seed int64, heightMap []float64) []float64 {
	elevationMap := make([]float64, size*size)
	rng := rand.New(rand.NewSource(seed + 1000))

	offsetX := rng.Float64() * 1000
	offsetY := rng.Float64() * 1000

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			idx := y*size + x
			baseHeight := heightMap[idx]

			// Add additional detail noise for elevation
			nx := (float64(x) + offsetX) * 0.02
			ny := (float64(y) + offsetY) * 0.02

			detail := 0.0
			amplitude := 0.3
			frequency := 1.0

			for oct := 0; oct < ElevationOctaves; oct++ {
				detail += noise.Noise2DSigned(nx*frequency, ny*frequency, seed+500+int64(oct)) * amplitude
				amplitude *= 0.5
				frequency *= 2.0
			}

			// Combine base height with detail and apply elevation curve
			combined := baseHeight + detail*0.3
			if combined < 0 {
				combined = 0
			}
			if combined > 1 {
				combined = 1
			}

			// Apply exponential curve for more dramatic hills/peaks
			elevation := combined * combined * MaxElevation
			elevationMap[idx] = elevation
		}
	}

	return elevationMap
}

// generateTerrainTypes classifies each cell based on height and slope.
func generateTerrainTypes(size int, heightMap, elevationMap []float64) []int {
	terrainTypes := make([]int, size*size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			idx := y*size + x
			height := heightMap[idx]
			terrainTypes[idx] = classifyTerrain(x, y, size, height, elevationMap)
		}
	}

	return terrainTypes
}

// classifyTerrain determines the terrain type for a single cell.
func classifyTerrain(x, y, size int, height float64, elevationMap []float64) int {
	// Check for peaks first
	if height >= PeakThreshold {
		return TerrainPeak
	}

	// Check for cliffs (steep elevation changes)
	maxSlope := calculateMaxSlope(x, y, size, elevationMap)
	if maxSlope >= CliffThreshold {
		return TerrainCliff
	}

	// Check for hills
	if height >= HillThreshold {
		return TerrainHill
	}

	return TerrainFlat
}

// calculateMaxSlope finds the steepest slope from this cell to neighbors.
func calculateMaxSlope(x, y, size int, elevationMap []float64) float64 {
	idx := y*size + x
	if idx >= len(elevationMap) {
		return 0
	}
	centerElevation := elevationMap[idx]
	maxSlope := 0.0

	// Check all 8 neighbors
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx < 0 || nx >= size || ny < 0 || ny >= size {
				continue
			}
			neighborIdx := ny*size + nx
			if neighborIdx >= len(elevationMap) {
				continue
			}
			neighborElevation := elevationMap[neighborIdx]
			slope := abs(centerElevation - neighborElevation)
			if slope > maxSlope {
				maxSlope = slope
			}
		}
	}

	return maxSlope / MaxElevation // Normalize to [0, 1] range
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetHeight returns the height at the given local coordinates.
func (c *Chunk) GetHeight(localX, localY int) float64 {
	if localX < 0 || localX >= c.Size || localY < 0 || localY >= c.Size {
		return 0
	}
	return c.HeightMap[localY*c.Size+localX]
}

// GetElevation returns the terrain elevation in world units at the given local coordinates.
func (c *Chunk) GetElevation(localX, localY int) float64 {
	if localX < 0 || localX >= c.Size || localY < 0 || localY >= c.Size {
		return 0
	}
	if c.ElevationMap == nil {
		return 0
	}
	return c.ElevationMap[localY*c.Size+localX]
}

// GetTerrainType returns the terrain type at the given local coordinates.
func (c *Chunk) GetTerrainType(localX, localY int) int {
	if localX < 0 || localX >= c.Size || localY < 0 || localY >= c.Size {
		return TerrainFlat
	}
	if c.TerrainTypes == nil {
		return TerrainFlat
	}
	return c.TerrainTypes[localY*c.Size+localX]
}

// IsCliff returns true if the terrain at the given coordinates is a cliff.
func (c *Chunk) IsCliff(localX, localY int) bool {
	return c.GetTerrainType(localX, localY) == TerrainCliff
}

// IsHill returns true if the terrain at the given coordinates is a hill.
func (c *Chunk) IsHill(localX, localY int) bool {
	t := c.GetTerrainType(localX, localY)
	return t == TerrainHill || t == TerrainPeak
}

// GetElevationDifference returns the elevation difference between two cells.
func (c *Chunk) GetElevationDifference(x1, y1, x2, y2 int) float64 {
	e1 := c.GetElevation(x1, y1)
	e2 := c.GetElevation(x2, y2)
	return abs(e1 - e2)
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
