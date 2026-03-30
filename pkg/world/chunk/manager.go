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
	maxSlope := findMaxNeighborSlope(x, y, size, centerElevation, elevationMap)
	return maxSlope / MaxElevation // Normalize to [0, 1] range
}

// findMaxNeighborSlope finds the steepest slope among all 8 neighbors.
func findMaxNeighborSlope(x, y, size int, centerElevation float64, elevationMap []float64) float64 {
	maxSlope := 0.0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			slope := getNeighborSlope(x+dx, y+dy, size, centerElevation, elevationMap)
			if slope > maxSlope {
				maxSlope = slope
			}
		}
	}
	return maxSlope
}

// getNeighborSlope calculates the slope to a single neighbor.
func getNeighborSlope(nx, ny, size int, centerElevation float64, elevationMap []float64) float64 {
	if nx < 0 || nx >= size || ny < 0 || ny >= size {
		return 0
	}
	neighborIdx := ny*size + nx
	if neighborIdx >= len(elevationMap) {
		return 0
	}
	return abs(centerElevation - elevationMap[neighborIdx])
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

// ========== Dynamic Terrain Modification ==========

// TerrainModification represents a single modification to the terrain.
type TerrainModification struct {
	X, Y           int     // Local coordinates within chunk
	NewHeight      float64 // New height value
	Timestamp      int64   // When the modification was made
	ModifierID     uint64  // Entity ID that caused the modification (0 for environment)
	ModType        int     // Type of modification
	OldHeight      float64 // Original height (for undo support)
	OldElevation   float64
	OldTerrainType int
}

// Modification types.
const (
	ModTypeDig     = iota // Lowering terrain
	ModTypeFill           // Raising terrain
	ModTypeSmooth         // Smoothing terrain
	ModTypeFlatten        // Flattening to a specific height
	ModTypeExplode        // Explosion crater
	ModTypeErode          // Natural erosion
)

// ModifiedChunk wraps a Chunk with a modification layer.
type ModifiedChunk struct {
	*Chunk
	modifications []TerrainModification
	modifiedMask  map[int]bool // index -> has modification
	dirty         bool         // true if modifications haven't been persisted
}

// NewModifiedChunk creates a new modified chunk wrapper.
func NewModifiedChunk(base *Chunk) *ModifiedChunk {
	return &ModifiedChunk{
		Chunk:         base,
		modifications: make([]TerrainModification, 0),
		modifiedMask:  make(map[int]bool),
		dirty:         false,
	}
}

// ApplyModification applies a terrain modification at the given location.
func (mc *ModifiedChunk) ApplyModification(mod TerrainModification) {
	if mod.X < 0 || mod.X >= mc.Size || mod.Y < 0 || mod.Y >= mc.Size {
		return
	}

	idx := mod.Y*mc.Size + mod.X

	// Store original values for undo
	mod.OldHeight = mc.HeightMap[idx]
	if mc.ElevationMap != nil {
		mod.OldElevation = mc.ElevationMap[idx]
	}
	if mc.TerrainTypes != nil {
		mod.OldTerrainType = mc.TerrainTypes[idx]
	}

	// Apply the modification based on type
	switch mod.ModType {
	case ModTypeDig:
		mc.HeightMap[idx] -= mod.NewHeight
		if mc.HeightMap[idx] < 0 {
			mc.HeightMap[idx] = 0
		}
	case ModTypeFill:
		mc.HeightMap[idx] += mod.NewHeight
		if mc.HeightMap[idx] > 1 {
			mc.HeightMap[idx] = 1
		}
	case ModTypeSmooth:
		mc.smoothTerrain(mod.X, mod.Y)
	case ModTypeFlatten:
		mc.HeightMap[idx] = mod.NewHeight
	case ModTypeExplode:
		mc.applyExplosionCrater(mod.X, mod.Y, mod.NewHeight)
	case ModTypeErode:
		mc.applyErosion(mod.X, mod.Y, mod.NewHeight)
	}

	// Recalculate derived terrain data at this location
	mc.recalculateElevation(mod.X, mod.Y)
	mc.recalculateTerrainType(mod.X, mod.Y)

	mc.modifications = append(mc.modifications, mod)
	mc.modifiedMask[idx] = true
	mc.dirty = true
}

// smoothTerrain averages height with neighbors.
func (mc *ModifiedChunk) smoothTerrain(x, y int) {
	sum := mc.HeightMap[y*mc.Size+x]
	count := 1.0

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < mc.Size && ny >= 0 && ny < mc.Size {
				sum += mc.HeightMap[ny*mc.Size+nx]
				count++
			}
		}
	}

	mc.HeightMap[y*mc.Size+x] = sum / count
}

// applyExplosionCrater creates a crater effect at the given location.
func (mc *ModifiedChunk) applyExplosionCrater(x, y int, radius float64) {
	radiusInt := int(radius)
	for dy := -radiusInt; dy <= radiusInt; dy++ {
		for dx := -radiusInt; dx <= radiusInt; dx++ {
			nx, ny := x+dx, y+dy
			if nx < 0 || nx >= mc.Size || ny < 0 || ny >= mc.Size {
				continue
			}
			dist := float64(dx*dx + dy*dy)
			maxDist := radius * radius
			if dist <= maxDist {
				// Crater depth decreases toward edges
				falloff := 1.0 - (dist / maxDist)
				idx := ny*mc.Size + nx
				mc.HeightMap[idx] -= 0.3 * falloff
				if mc.HeightMap[idx] < 0 {
					mc.HeightMap[idx] = 0
				}
			}
		}
	}
}

// applyErosion simulates natural erosion at the given location.
func (mc *ModifiedChunk) applyErosion(x, y int, intensity float64) {
	idx := y*mc.Size + x
	currentHeight := mc.HeightMap[idx]
	lowestHeight := mc.findLowestNeighborHeight(x, y, currentHeight)

	diff := currentHeight - lowestHeight
	mc.HeightMap[idx] -= diff * intensity * 0.5
	if mc.HeightMap[idx] < 0 {
		mc.HeightMap[idx] = 0
	}
}

// findLowestNeighborHeight finds the minimum height among 8-neighbors.
func (mc *ModifiedChunk) findLowestNeighborHeight(x, y int, currentHeight float64) float64 {
	lowestHeight := currentHeight
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < mc.Size && ny >= 0 && ny < mc.Size {
				h := mc.HeightMap[ny*mc.Size+nx]
				if h < lowestHeight {
					lowestHeight = h
				}
			}
		}
	}
	return lowestHeight
}

// recalculateElevation updates elevation at a modified point.
func (mc *ModifiedChunk) recalculateElevation(x, y int) {
	if mc.ElevationMap == nil {
		return
	}
	idx := y*mc.Size + x
	height := mc.HeightMap[idx]
	// Use exponential curve for more dramatic elevation
	mc.ElevationMap[idx] = height * height * MaxElevation
}

// recalculateTerrainType updates terrain type at a modified point.
func (mc *ModifiedChunk) recalculateTerrainType(x, y int) {
	if mc.TerrainTypes == nil || mc.ElevationMap == nil {
		return
	}
	idx := y*mc.Size + x
	height := mc.HeightMap[idx]
	mc.TerrainTypes[idx] = classifyTerrain(x, y, mc.Size, height, mc.ElevationMap)
}

// UndoLastModification reverts the most recent modification.
func (mc *ModifiedChunk) UndoLastModification() bool {
	if len(mc.modifications) == 0 {
		return false
	}

	mod := mc.modifications[len(mc.modifications)-1]
	mc.modifications = mc.modifications[:len(mc.modifications)-1]

	idx := mod.Y*mc.Size + mod.X
	mc.HeightMap[idx] = mod.OldHeight
	if mc.ElevationMap != nil {
		mc.ElevationMap[idx] = mod.OldElevation
	}
	if mc.TerrainTypes != nil {
		mc.TerrainTypes[idx] = mod.OldTerrainType
	}

	delete(mc.modifiedMask, idx)
	mc.dirty = len(mc.modifications) > 0

	return true
}

// IsModified returns true if the cell has been modified.
func (mc *ModifiedChunk) IsModified(x, y int) bool {
	if x < 0 || x >= mc.Size || y < 0 || y >= mc.Size {
		return false
	}
	return mc.modifiedMask[y*mc.Size+x]
}

// ModificationCount returns the number of modifications applied to this chunk.
func (mc *ModifiedChunk) ModificationCount() int {
	return len(mc.modifications)
}

// IsDirty returns true if there are unsaved modifications.
func (mc *ModifiedChunk) IsDirty() bool {
	return mc.dirty
}

// ClearDirty marks modifications as saved.
func (mc *ModifiedChunk) ClearDirty() {
	mc.dirty = false
}

// GetModifications returns a copy of all modifications for persistence.
func (mc *ModifiedChunk) GetModifications() []TerrainModification {
	result := make([]TerrainModification, len(mc.modifications))
	copy(result, mc.modifications)
	return result
}

// RestoreModifications applies a list of modifications (for loading from save).
func (mc *ModifiedChunk) RestoreModifications(mods []TerrainModification) {
	for _, mod := range mods {
		mc.ApplyModification(mod)
	}
	mc.dirty = false
}

// ModifyTerrain is a convenience method to create and apply a modification.
func (mc *ModifiedChunk) ModifyTerrain(x, y, modType int, value float64, modifierID uint64, timestamp int64) {
	mod := TerrainModification{
		X:          x,
		Y:          y,
		NewHeight:  value,
		Timestamp:  timestamp,
		ModifierID: modifierID,
		ModType:    modType,
	}
	mc.ApplyModification(mod)
}

// Dig lowers terrain at the given location.
func (mc *ModifiedChunk) Dig(x, y int, depth float64, modifierID uint64, timestamp int64) {
	mc.ModifyTerrain(x, y, ModTypeDig, depth, modifierID, timestamp)
}

// Fill raises terrain at the given location.
func (mc *ModifiedChunk) Fill(x, y int, height float64, modifierID uint64, timestamp int64) {
	mc.ModifyTerrain(x, y, ModTypeFill, height, modifierID, timestamp)
}

// Smooth averages terrain with neighbors.
func (mc *ModifiedChunk) Smooth(x, y int, modifierID uint64, timestamp int64) {
	mc.ModifyTerrain(x, y, ModTypeSmooth, 0, modifierID, timestamp)
}

// Flatten sets terrain to a specific height.
func (mc *ModifiedChunk) Flatten(x, y int, targetHeight float64, modifierID uint64, timestamp int64) {
	mc.ModifyTerrain(x, y, ModTypeFlatten, targetHeight, modifierID, timestamp)
}

// Explode creates a crater at the given location.
func (mc *ModifiedChunk) Explode(x, y int, radius float64, modifierID uint64, timestamp int64) {
	mc.ModifyTerrain(x, y, ModTypeExplode, radius, modifierID, timestamp)
}

// Erode applies natural erosion at the given location.
func (mc *ModifiedChunk) Erode(x, y int, intensity float64, modifierID uint64, timestamp int64) {
	mc.ModifyTerrain(x, y, ModTypeErode, intensity, modifierID, timestamp)
}

// ========== Terrain LOD (Level of Detail) System ==========

// LODLevel represents different detail levels for terrain rendering.
type LODLevel int

const (
	LODFull    LODLevel = 0 // Full detail (every cell)
	LODHalf    LODLevel = 1 // Half detail (every 2nd cell)
	LODQuarter LODLevel = 2 // Quarter detail (every 4th cell)
	LODEighth  LODLevel = 3 // Eighth detail (every 8th cell)
)

// LODChunk represents a chunk at a specific level of detail.
type LODChunk struct {
	*Chunk
	Level        LODLevel
	HeightMap    []float64 // Downsampled height map
	ElevationMap []float64 // Downsampled elevation map
	LODSize      int       // Size at this LOD level
}

// ChunkLODCache manages LOD versions of chunks.
type ChunkLODCache struct {
	chunks map[[2]int]map[LODLevel]*LODChunk
	mu     sync.RWMutex
}

// NewChunkLODCache creates a new LOD cache.
func NewChunkLODCache() *ChunkLODCache {
	return &ChunkLODCache{
		chunks: make(map[[2]int]map[LODLevel]*LODChunk),
	}
}

// GetLOD returns a chunk at the specified detail level.
func (c *ChunkLODCache) GetLOD(chunk *Chunk, level LODLevel) *LODChunk {
	if level == LODFull {
		return &LODChunk{
			Chunk:        chunk,
			Level:        LODFull,
			HeightMap:    chunk.HeightMap,
			ElevationMap: chunk.ElevationMap,
			LODSize:      chunk.Size,
		}
	}

	key := [2]int{chunk.X, chunk.Y}

	c.mu.RLock()
	if levels, ok := c.chunks[key]; ok {
		if lod, ok := levels[level]; ok {
			c.mu.RUnlock()
			return lod
		}
	}
	c.mu.RUnlock()

	// Generate LOD
	lod := generateLOD(chunk, level)

	c.mu.Lock()
	if c.chunks[key] == nil {
		c.chunks[key] = make(map[LODLevel]*LODChunk)
	}
	c.chunks[key][level] = lod
	c.mu.Unlock()

	return lod
}

// InvalidateLOD removes cached LOD data for a chunk (call after modification).
func (c *ChunkLODCache) InvalidateLOD(x, y int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.chunks, [2]int{x, y})
}

// generateLOD creates a downsampled version of a chunk.
func generateLOD(chunk *Chunk, level LODLevel) *LODChunk {
	step := 1 << int(level)
	lodSize := calculateLODSize(chunk.Size, step)

	heightMap := make([]float64, lodSize*lodSize)
	elevationMap := make([]float64, lodSize*lodSize)

	for ly := 0; ly < lodSize; ly++ {
		for lx := 0; lx < lodSize; lx++ {
			sumH, sumE, count := sampleChunkRegion(chunk, lx, ly, step)
			lodIdx := ly*lodSize + lx
			if count > 0 {
				heightMap[lodIdx] = sumH / count
				elevationMap[lodIdx] = sumE / count
			}
		}
	}

	return &LODChunk{
		Chunk:        chunk,
		Level:        level,
		HeightMap:    heightMap,
		ElevationMap: elevationMap,
		LODSize:      lodSize,
	}
}

// calculateLODSize computes the size of an LOD map.
func calculateLODSize(chunkSize, step int) int {
	lodSize := chunkSize / step
	if lodSize < 1 {
		lodSize = 1
	}
	return lodSize
}

// sampleChunkRegion averages height and elevation for an LOD cell.
func sampleChunkRegion(chunk *Chunk, lx, ly, step int) (sumH, sumE, count float64) {
	for dy := 0; dy < step; dy++ {
		for dx := 0; dx < step; dx++ {
			sx := lx*step + dx
			sy := ly*step + dy
			if sx < chunk.Size && sy < chunk.Size {
				idx := sy*chunk.Size + sx
				sumH += chunk.HeightMap[idx]
				if chunk.ElevationMap != nil {
					sumE += chunk.ElevationMap[idx]
				}
				count++
			}
		}
	}
	return sumH, sumE, count
}

// GetHeight returns the height at LOD coordinates.
func (lc *LODChunk) GetHeight(lodX, lodY int) float64 {
	if lodX < 0 || lodX >= lc.LODSize || lodY < 0 || lodY >= lc.LODSize {
		return 0
	}
	return lc.HeightMap[lodY*lc.LODSize+lodX]
}

// GetElevation returns the elevation at LOD coordinates.
func (lc *LODChunk) GetElevation(lodX, lodY int) float64 {
	if lodX < 0 || lodX >= lc.LODSize || lodY < 0 || lodY >= lc.LODSize {
		return 0
	}
	if lc.ElevationMap == nil {
		return 0
	}
	return lc.ElevationMap[lodY*lc.LODSize+lodX]
}

// ToFullCoords converts LOD coordinates to full-resolution coordinates.
func (lc *LODChunk) ToFullCoords(lodX, lodY int) (int, int) {
	step := 1 << int(lc.Level)
	return lodX * step, lodY * step
}

// FromFullCoords converts full-resolution coordinates to LOD coordinates.
func (lc *LODChunk) FromFullCoords(fullX, fullY int) (int, int) {
	step := 1 << int(lc.Level)
	return fullX / step, fullY / step
}

// VertexCount returns the number of vertices needed to render this LOD.
func (lc *LODChunk) VertexCount() int {
	return lc.LODSize * lc.LODSize
}

// TriangleCount returns the number of triangles needed to render this LOD.
func (lc *LODChunk) TriangleCount() int {
	if lc.LODSize < 2 {
		return 0
	}
	return (lc.LODSize - 1) * (lc.LODSize - 1) * 2
}

// CalculateLODLevel determines appropriate LOD based on distance from camera.
func CalculateLODLevel(distanceSquared float64) LODLevel {
	// Distance thresholds for LOD transitions (squared for efficiency)
	const (
		lodHalfDist    = 4096  // 64^2
		lodQuarterDist = 16384 // 128^2
		lodEighthDist  = 65536 // 256^2
	)

	switch {
	case distanceSquared < lodHalfDist:
		return LODFull
	case distanceSquared < lodQuarterDist:
		return LODHalf
	case distanceSquared < lodEighthDist:
		return LODQuarter
	default:
		return LODEighth
	}
}

// LODManager coordinates LOD for multiple chunks around a viewpoint.
type LODManager struct {
	cache     *ChunkLODCache
	manager   *Manager
	viewX     float64
	viewY     float64
	chunkSize int
}

// NewLODManager creates a new LOD manager.
func NewLODManager(chunkManager *Manager) *LODManager {
	return &LODManager{
		cache:     NewChunkLODCache(),
		manager:   chunkManager,
		chunkSize: chunkManager.ChunkSize,
	}
}

// SetViewpoint updates the camera/player position for LOD calculations.
func (lm *LODManager) SetViewpoint(worldX, worldY float64) {
	lm.viewX = worldX
	lm.viewY = worldY
}

// GetChunkLOD returns a chunk at the appropriate LOD for its distance from viewpoint.
func (lm *LODManager) GetChunkLOD(chunkX, chunkY int) *LODChunk {
	chunk := lm.manager.GetChunk(chunkX, chunkY)

	// Calculate chunk center in world coordinates
	chunkCenterX := float64(chunkX*lm.chunkSize) + float64(lm.chunkSize)/2
	chunkCenterY := float64(chunkY*lm.chunkSize) + float64(lm.chunkSize)/2

	// Distance squared from viewpoint to chunk center
	dx := chunkCenterX - lm.viewX
	dy := chunkCenterY - lm.viewY
	distSq := dx*dx + dy*dy

	level := CalculateLODLevel(distSq)
	return lm.cache.GetLOD(chunk, level)
}

// GetChunksInView returns all chunks visible from the viewpoint with appropriate LOD.
func (lm *LODManager) GetChunksInView(viewRadius int) []*LODChunk {
	centerChunkX := int(lm.viewX) / lm.chunkSize
	centerChunkY := int(lm.viewY) / lm.chunkSize

	chunks := make([]*LODChunk, 0)
	for dy := -viewRadius; dy <= viewRadius; dy++ {
		for dx := -viewRadius; dx <= viewRadius; dx++ {
			cx := centerChunkX + dx
			cy := centerChunkY + dy
			chunks = append(chunks, lm.GetChunkLOD(cx, cy))
		}
	}

	return chunks
}

// InvalidateChunk invalidates LOD cache for a modified chunk.
func (lm *LODManager) InvalidateChunk(chunkX, chunkY int) {
	lm.cache.InvalidateLOD(chunkX, chunkY)
}

// CacheStats returns statistics about the LOD cache.
func (lm *LODManager) CacheStats() (totalChunks, totalLODs int) {
	lm.cache.mu.RLock()
	defer lm.cache.mu.RUnlock()

	totalChunks = len(lm.cache.chunks)
	for _, levels := range lm.cache.chunks {
		totalLODs += len(levels)
	}
	return totalChunks, totalLODs
}
