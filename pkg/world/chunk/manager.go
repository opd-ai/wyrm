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
	TerrainFlat   = 0 // Normal walkable terrain
	TerrainHill   = 1 // Elevated terrain (gradual slope)
	TerrainCliff  = 2 // Steep terrain (impassable edge)
	TerrainPeak   = 3 // High elevation peak
	TerrainValley = 4 // Low-lying depression
	TerrainWater  = 5 // Water surface (below water level)
	TerrainForest = 6 // Tree-covered terrain
	TerrainRoad   = 7 // Flat walkway connecting POIs
)

// Terrain elevation constants.
const (
	// HillThreshold is the height value above which terrain becomes a hill.
	HillThreshold = 0.5
	// CliffThreshold is the height difference that creates a cliff edge.
	CliffThreshold = 0.15
	// PeakThreshold is the height value above which terrain becomes a peak.
	PeakThreshold = 0.8
	// ValleyThreshold is the height value below which terrain becomes a valley.
	ValleyThreshold = 0.2
	// WaterLevel is the height value below which terrain becomes water.
	WaterLevel = 0.15
	// ForestNoiseThreshold is the noise value above which forest appears.
	ForestNoiseThreshold = 0.55
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
	ElevationMap []float64     // Terrain elevation in world units
	TerrainTypes []int         // Terrain type per cell
	BiomeMap     []float64     // Biome noise values for vegetation density
	DetailSpawns []DetailSpawn // Vegetation, rocks, and other detail entities
	WallHeights  []float64     // Per-cell wall height multipliers for variable-height rendering
}

// DetailSpawnType represents different types of detail entities.
type DetailSpawnType int

const (
	// Trees and vegetation
	DetailSpawnTree DetailSpawnType = iota
	DetailSpawnBush
	DetailSpawnDeadTree
	DetailSpawnGrass
	DetailSpawnFlower

	// Rocks and terrain details
	DetailSpawnRock
	DetailSpawnBoulder
	DetailSpawnDebris

	// Urban details
	DetailSpawnLampPost
	DetailSpawnBench
	DetailSpawnTrashCan

	// Wasteland details
	DetailSpawnScrap
	DetailSpawnRubble
	DetailSpawnBones
)

// DetailSpawn represents a spawned detail entity in the chunk.
type DetailSpawn struct {
	Type      DetailSpawnType
	LocalX    float64 // X position within chunk (0 to Size)
	LocalY    float64 // Y position within chunk (0 to Size)
	Scale     float64 // Size multiplier (0.5 to 1.5 typical)
	Rotation  float64 // Rotation in radians
	Variation int     // Variation index for visual variety
}

// NewChunk creates a new chunk at the given coordinates with generated terrain.
func NewChunk(x, y, size int, seed int64) *Chunk {
	return NewChunkWithNoiseType(x, y, size, seed, noise.NoiseTypeValue)
}

// NewChunkWithNoiseType creates a new chunk with the specified noise algorithm.
func NewChunkWithNoiseType(x, y, size int, seed int64, noiseType noise.NoiseType) *Chunk {
	heightMap := generateHeightMapWithNoiseType(size, seed, noiseType)
	elevationMap := generateElevationMapWithNoiseType(size, seed, heightMap, noiseType)
	biomeMap := generateBiomeMapWorldSpace(x, y, size, seed)
	terrainTypes := generateTerrainTypes(size, heightMap, elevationMap, biomeMap)
	detailSpawns := generateDetailSpawns(size, seed, terrainTypes, biomeMap)
	wallHeights := generateWallHeights(size, heightMap, elevationMap, terrainTypes)
	return &Chunk{
		X:            x,
		Y:            y,
		Size:         size,
		Seed:         seed,
		HeightMap:    heightMap,
		ElevationMap: elevationMap,
		TerrainTypes: terrainTypes,
		BiomeMap:     biomeMap,
		DetailSpawns: detailSpawns,
		WallHeights:  wallHeights,
	}
}

// generateHeightMap creates a procedural heightmap using value noise.
func generateHeightMap(size int, seed int64) []float64 {
	return generateHeightMapWithNoiseType(size, seed, noise.NoiseTypeValue)
}

// generateHeightMapWithNoiseType creates a procedural heightmap using the specified noise.
func generateHeightMapWithNoiseType(size int, seed int64, noiseType noise.NoiseType) []float64 {
	heightMap := make([]float64, size*size)
	rng := rand.New(rand.NewSource(seed))

	// Generate base offset for this chunk
	offsetX := rng.Float64() * 1000
	offsetY := rng.Float64() * 1000

	// Select noise function based on type
	noiseFunc := noise.Noise2DSigned
	if noiseType == noise.NoiseTypeGradient {
		noiseFunc = noise.GradientNoise2D
	}

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
				height += noiseFunc(nx*frequency, ny*frequency, seed+int64(oct)) * amplitude
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
	return generateElevationMapWithNoiseType(size, seed, heightMap, noise.NoiseTypeValue)
}

// generateElevationMapWithNoiseType converts the heightmap to world-unit elevations.
func generateElevationMapWithNoiseType(size int, seed int64, heightMap []float64, noiseType noise.NoiseType) []float64 {
	elevationMap := make([]float64, size*size)
	rng := rand.New(rand.NewSource(seed + 1000))

	offsetX := rng.Float64() * 1000
	offsetY := rng.Float64() * 1000

	// Select noise function based on type
	noiseFunc := noise.Noise2DSigned
	if noiseType == noise.NoiseTypeGradient {
		noiseFunc = noise.GradientNoise2D
	}

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
				detail += noiseFunc(nx*frequency, ny*frequency, seed+500+int64(oct)) * amplitude
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

// generateBiomeMap creates a noise-based biome map for vegetation placement.
func generateBiomeMap(size int, seed int64) []float64 {
	biomeMap := make([]float64, size*size)
	rng := rand.New(rand.NewSource(seed + 2000))

	offsetX := rng.Float64() * 1000
	offsetY := rng.Float64() * 1000

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			idx := y*size + x
			// Use larger scale noise for biome regions
			nx := (float64(x) + offsetX) * 0.008
			ny := (float64(y) + offsetY) * 0.008

			biomeValue := 0.0
			amplitude := 1.0
			frequency := 1.0

			// 3 octaves for biome variation
			for oct := 0; oct < 3; oct++ {
				biomeValue += noise.Noise2DSigned(nx*frequency, ny*frequency, seed+2000+int64(oct)) * amplitude
				amplitude *= 0.5
				frequency *= 2.0
			}

			// Normalize to [0, 1]
			biomeMap[idx] = (biomeValue + 1.0) / 2.0
		}
	}

	return biomeMap
}

// BiomeScale controls the world-space frequency of biome noise.
// Lower values create larger biome regions.
const BiomeScale = 0.003

// generateBiomeMapWorldSpace creates biome values using world-space coordinates for seamless transitions.
// This ensures consistent biome values across chunk boundaries.
func generateBiomeMapWorldSpace(chunkX, chunkY, size int, seed int64) []float64 {
	biomeMap := make([]float64, size*size)

	// Calculate world-space offset for this chunk
	worldOffsetX := float64(chunkX * size)
	worldOffsetY := float64(chunkY * size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			idx := y*size + x

			// Calculate world-space coordinates
			worldX := worldOffsetX + float64(x)
			worldY := worldOffsetY + float64(y)

			// Use world-space coordinates for seamless noise
			nx := worldX * BiomeScale
			ny := worldY * BiomeScale

			biomeValue := 0.0
			amplitude := 1.0
			frequency := 1.0

			// 4 octaves for richer biome variation
			for oct := 0; oct < 4; oct++ {
				biomeValue += noise.Noise2DSigned(nx*frequency, ny*frequency, seed+2000+int64(oct)) * amplitude
				amplitude *= 0.5
				frequency *= 2.0
			}

			// Normalize to [0, 1]
			biomeValue = (biomeValue + 1.0) / 2.0

			// Apply edge blending to smooth transitions
			biomeValue = applyBiomeEdgeBlend(x, y, size, biomeValue)

			biomeMap[idx] = biomeValue
		}
	}

	return biomeMap
}

// applyBiomeEdgeBlend applies smooth blending at chunk edges to prevent abrupt transitions.
// Uses a smooth interpolation function near chunk boundaries.
func applyBiomeEdgeBlend(localX, localY, size int, value float64) float64 {
	// Calculate distance from nearest edge
	distFromLeft := localX
	distFromRight := size - 1 - localX
	distFromTop := localY
	distFromBottom := size - 1 - localY

	minDistX := minInt(distFromLeft, distFromRight)
	minDistY := minInt(distFromTop, distFromBottom)
	minDist := minInt(minDistX, minDistY)

	// If outside blend zone, return unmodified value
	if minDist >= BiomeBlendWidth {
		return value
	}

	// Calculate blend factor (0 at edge, 1 at blend boundary)
	t := float64(minDist) / float64(BiomeBlendWidth)

	// Smoothstep interpolation for gradual transition
	blendFactor := biomeSmooth(t)

	// Blend toward center value (0.5) at edges for smoother chunk boundary matching
	centerValue := 0.5
	return lerp(centerValue, value, blendFactor)
}

// biomeSmooth performs smooth Hermite interpolation between 0 and 1 for biome blending.
func biomeSmooth(t float64) float64 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return t * t * (3 - 2*t)
}

// lerp performs linear interpolation between a and b.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// minInt returns the smaller of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Detail spawn density constants.
const (
	// BaseDensity is the base number of spawns per 100 cells.
	BaseDensity = 5
	// ForestDensityMultiplier increases density in forested areas.
	ForestDensityMultiplier = 3.0
	// MountainDensityMultiplier for rocky mountain areas.
	MountainDensityMultiplier = 2.0
)

// generateDetailSpawns creates vegetation and detail entities for the chunk.
func generateDetailSpawns(size int, seed int64, terrainTypes []int, biomeMap []float64) []DetailSpawn {
	rng := rand.New(rand.NewSource(seed + 3000))
	spawns := make([]DetailSpawn, 0, size)

	// Calculate base number of potential spawn attempts
	baseAttempts := (size * size * BaseDensity) / 100

	for i := 0; i < baseAttempts; i++ {
		// Random position within chunk
		localX := rng.Float64() * float64(size)
		localY := rng.Float64() * float64(size)

		// Get terrain type at this position
		gridX := int(localX)
		gridY := int(localY)
		if gridX >= size {
			gridX = size - 1
		}
		if gridY >= size {
			gridY = size - 1
		}
		idx := gridY*size + gridX
		terrainType := terrainTypes[idx]

		// Get biome value for density variation
		biomeValue := 0.5
		if biomeMap != nil && idx < len(biomeMap) {
			biomeValue = biomeMap[idx]
		}

		// Determine if we should spawn based on terrain and biome
		spawn := selectDetailSpawn(terrainType, biomeValue, rng)
		if spawn == nil {
			continue
		}

		spawn.LocalX = localX
		spawn.LocalY = localY
		spawn.Scale = 0.7 + rng.Float64()*0.6    // 0.7 to 1.3
		spawn.Rotation = rng.Float64() * 6.28318 // 0 to 2π
		spawn.Variation = rng.Intn(4)            // 0 to 3

		spawns = append(spawns, *spawn)
	}

	return spawns
}

// selectDetailSpawn determines what type of detail to spawn based on terrain.
func selectDetailSpawn(terrainType int, biomeValue float64, rng *rand.Rand) *DetailSpawn {
	// Skip spawning on impassable terrain
	if terrainType == TerrainWater || terrainType == TerrainCliff {
		return nil
	}

	// Use terrain type and biome to determine spawn type
	switch terrainType {
	case TerrainForest:
		return selectForestSpawn(biomeValue, rng)
	case TerrainHill, TerrainPeak:
		return selectMountainSpawn(biomeValue, rng)
	case TerrainValley:
		return selectValleySpawn(biomeValue, rng)
	case TerrainFlat:
		return selectFlatSpawn(biomeValue, rng)
	default:
		return nil
	}
}

// selectForestSpawn chooses a spawn for forested terrain.
func selectForestSpawn(biomeValue float64, rng *rand.Rand) *DetailSpawn {
	// Higher biome value = denser forest
	if rng.Float64() > biomeValue*ForestDensityMultiplier {
		return nil
	}

	roll := rng.Float64()
	switch {
	case roll < 0.6:
		return &DetailSpawn{Type: DetailSpawnTree}
	case roll < 0.8:
		return &DetailSpawn{Type: DetailSpawnBush}
	case roll < 0.95:
		return &DetailSpawn{Type: DetailSpawnGrass}
	default:
		return &DetailSpawn{Type: DetailSpawnFlower}
	}
}

// selectMountainSpawn chooses a spawn for mountain terrain.
func selectMountainSpawn(biomeValue float64, rng *rand.Rand) *DetailSpawn {
	if rng.Float64() > MountainDensityMultiplier*0.3 {
		return nil
	}

	roll := rng.Float64()
	switch {
	case roll < 0.5:
		return &DetailSpawn{Type: DetailSpawnRock}
	case roll < 0.8:
		return &DetailSpawn{Type: DetailSpawnBoulder}
	default:
		return &DetailSpawn{Type: DetailSpawnGrass}
	}
}

// selectValleySpawn chooses a spawn for valley terrain.
func selectValleySpawn(biomeValue float64, rng *rand.Rand) *DetailSpawn {
	if rng.Float64() > biomeValue {
		return nil
	}

	roll := rng.Float64()
	switch {
	case roll < 0.4:
		return &DetailSpawn{Type: DetailSpawnGrass}
	case roll < 0.7:
		return &DetailSpawn{Type: DetailSpawnFlower}
	case roll < 0.9:
		return &DetailSpawn{Type: DetailSpawnBush}
	default:
		return &DetailSpawn{Type: DetailSpawnDeadTree}
	}
}

// selectFlatSpawn chooses a spawn for flat terrain.
func selectFlatSpawn(biomeValue float64, rng *rand.Rand) *DetailSpawn {
	// Sparse spawning on flat terrain
	if rng.Float64() > biomeValue*0.5 {
		return nil
	}

	roll := rng.Float64()
	switch {
	case roll < 0.3:
		return &DetailSpawn{Type: DetailSpawnGrass}
	case roll < 0.5:
		return &DetailSpawn{Type: DetailSpawnRock}
	case roll < 0.7:
		return &DetailSpawn{Type: DetailSpawnBush}
	default:
		return &DetailSpawn{Type: DetailSpawnFlower}
	}
}

// generateTerrainTypes classifies each cell based on height and slope.
func generateTerrainTypes(size int, heightMap, elevationMap, biomeMap []float64) []int {
	terrainTypes := make([]int, size*size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			idx := y*size + x
			height := heightMap[idx]
			biome := 0.0
			if biomeMap != nil && idx < len(biomeMap) {
				biome = biomeMap[idx]
			}
			terrainTypes[idx] = classifyTerrain(x, y, size, height, biome, elevationMap)
		}
	}

	return terrainTypes
}

// classifyTerrain determines the terrain type for a single cell.
func classifyTerrain(x, y, size int, height, biome float64, elevationMap []float64) int {
	// Check for water first (lowest priority terrain)
	if height < WaterLevel {
		return TerrainWater
	}

	// Check for valleys (low-lying areas above water)
	if height < ValleyThreshold {
		return TerrainValley
	}

	// Check for peaks (highest priority terrain)
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

	// Check for forest (mid-elevation with high biome value)
	if biome >= ForestNoiseThreshold && height >= ValleyThreshold && height < HillThreshold {
		return TerrainForest
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

// Wall height constants for variable-height rendering.
const (
	// MinWallHeight is the minimum wall height multiplier.
	MinWallHeight = 0.25
	// MaxWallHeightMultiplier is the maximum wall height multiplier.
	MaxWallHeightMultiplier = 3.0
	// DefaultWallHeight is the standard wall height multiplier.
	DefaultWallHeight = 1.0
)

// generateWallHeights creates per-cell wall height multipliers based on terrain.
// Wall heights vary based on terrain type and elevation to create visual variety.
func generateWallHeights(size int, heightMap, elevationMap []float64, terrainTypes []int) []float64 {
	wallHeights := make([]float64, size*size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			idx := y*size + x
			wallHeights[idx] = calculateWallHeight(idx, heightMap, elevationMap, terrainTypes)
		}
	}

	return wallHeights
}

// calculateWallHeight determines the wall height for a single cell.
func calculateWallHeight(idx int, heightMap, elevationMap []float64, terrainTypes []int) float64 {
	// Get base values
	height := 0.0
	if idx < len(heightMap) {
		height = heightMap[idx]
	}

	elevation := 0.0
	if elevationMap != nil && idx < len(elevationMap) {
		elevation = elevationMap[idx]
	}

	terrainType := TerrainFlat
	if idx < len(terrainTypes) {
		terrainType = terrainTypes[idx]
	}

	return wallHeightFromTerrain(height, elevation, terrainType)
}

// wallHeightFromTerrain calculates wall height based on terrain properties.
func wallHeightFromTerrain(height, elevation float64, terrainType int) float64 {
	// Base height from terrain height
	baseHeight := DefaultWallHeight

	switch terrainType {
	case TerrainFlat:
		// Flat terrain has standard walls with slight variation
		baseHeight = DefaultWallHeight + (height-0.5)*0.5
	case TerrainHill:
		// Hills have taller structures (towers, lookouts)
		baseHeight = DefaultWallHeight + height*0.5
	case TerrainCliff:
		// Cliffs have tall walls representing the cliff face
		baseHeight = DefaultWallHeight + 1.0
	case TerrainPeak:
		// Peaks have the tallest structures (castle towers, monuments)
		baseHeight = DefaultWallHeight + 1.5
	case TerrainValley:
		// Valleys have shorter structures
		baseHeight = DefaultWallHeight * 0.75
	case TerrainWater:
		// Water has no walls (returns default for potential bridges/piers)
		return DefaultWallHeight
	case TerrainForest:
		// Forest can have varying height trees represented as walls
		baseHeight = DefaultWallHeight + height*0.8
	case TerrainRoad:
		// Roads are flat, minimal height variation
		return DefaultWallHeight
	}

	// Add variation from elevation
	elevationFactor := elevation / MaxElevation
	baseHeight += elevationFactor * 0.3

	// Clamp to valid range
	if baseHeight < MinWallHeight {
		baseHeight = MinWallHeight
	}
	if baseHeight > MaxWallHeightMultiplier {
		baseHeight = MaxWallHeightMultiplier
	}

	return baseHeight
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

// GetWallHeight returns the wall height multiplier at the given local coordinates.
// Returns DefaultWallHeight if coordinates are out of bounds or WallHeights is nil.
func (c *Chunk) GetWallHeight(localX, localY int) float64 {
	if localX < 0 || localX >= c.Size || localY < 0 || localY >= c.Size {
		return DefaultWallHeight
	}
	if c.WallHeights == nil {
		return DefaultWallHeight
	}
	return c.WallHeights[localY*c.Size+localX]
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

// IsWater returns true if the terrain at the given coordinates is water.
func (c *Chunk) IsWater(localX, localY int) bool {
	return c.GetTerrainType(localX, localY) == TerrainWater
}

// IsValley returns true if the terrain at the given coordinates is a valley.
func (c *Chunk) IsValley(localX, localY int) bool {
	return c.GetTerrainType(localX, localY) == TerrainValley
}

// IsForest returns true if the terrain at the given coordinates is forested.
func (c *Chunk) IsForest(localX, localY int) bool {
	return c.GetTerrainType(localX, localY) == TerrainForest
}

// IsRoad returns true if the terrain at the given coordinates is a road.
func (c *Chunk) IsRoad(localX, localY int) bool {
	return c.GetTerrainType(localX, localY) == TerrainRoad
}

// IsPassable returns true if the terrain can be walked on.
func (c *Chunk) IsPassable(localX, localY int) bool {
	t := c.GetTerrainType(localX, localY)
	// Water and cliffs are impassable
	return t != TerrainWater && t != TerrainCliff
}

// GetBiomeValue returns the biome noise value at the given local coordinates.
func (c *Chunk) GetBiomeValue(localX, localY int) float64 {
	if localX < 0 || localX >= c.Size || localY < 0 || localY >= c.Size {
		return 0
	}
	if c.BiomeMap == nil {
		return 0
	}
	return c.BiomeMap[localY*c.Size+localX]
}

// GetDetailSpawns returns all detail spawns in this chunk.
func (c *Chunk) GetDetailSpawns() []DetailSpawn {
	return c.DetailSpawns
}

// GetDetailSpawnsInArea returns detail spawns within a local area.
func (c *Chunk) GetDetailSpawnsInArea(minX, minY, maxX, maxY float64) []DetailSpawn {
	if c.DetailSpawns == nil {
		return nil
	}

	result := make([]DetailSpawn, 0)
	for _, spawn := range c.DetailSpawns {
		if spawn.LocalX >= minX && spawn.LocalX <= maxX &&
			spawn.LocalY >= minY && spawn.LocalY <= maxY {
			result = append(result, spawn)
		}
	}
	return result
}

// DetailSpawnCount returns the number of detail spawns in this chunk.
func (c *Chunk) DetailSpawnCount() int {
	return len(c.DetailSpawns)
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
	NoiseType noise.NoiseType // Noise algorithm to use for generation
	mu        sync.RWMutex
	loaded    map[[2]int]*Chunk

	// Async generation
	asyncEnabled bool
	workQueue    chan chunkGenRequest
	pending      map[[2]int]bool
	pendingMu    sync.Mutex
	stopChan     chan struct{}
}

// chunkGenRequest represents a request to generate a chunk.
type chunkGenRequest struct {
	X, Y int
}

// NewManager creates a new chunk manager with default value noise.
func NewManager(chunkSize int, seed int64) *Manager {
	return &Manager{
		ChunkSize: chunkSize,
		Seed:      seed,
		NoiseType: noise.NoiseTypeValue,
		loaded:    make(map[[2]int]*Chunk),
		pending:   make(map[[2]int]bool),
	}
}

// NewManagerWithNoiseType creates a new chunk manager with specified noise type.
func NewManagerWithNoiseType(chunkSize int, seed int64, noiseType noise.NoiseType) *Manager {
	return &Manager{
		ChunkSize: chunkSize,
		Seed:      seed,
		NoiseType: noiseType,
		loaded:    make(map[[2]int]*Chunk),
		pending:   make(map[[2]int]bool),
	}
}

// EnableAsyncGeneration starts background chunk generation workers.
func (cm *Manager) EnableAsyncGeneration(numWorkers int) {
	if cm.asyncEnabled {
		return
	}
	cm.asyncEnabled = true
	cm.workQueue = make(chan chunkGenRequest, 64)
	cm.stopChan = make(chan struct{})

	for i := 0; i < numWorkers; i++ {
		go cm.chunkGenerationWorker()
	}
}

// DisableAsyncGeneration stops background chunk generation workers.
func (cm *Manager) DisableAsyncGeneration() {
	if !cm.asyncEnabled {
		return
	}
	cm.asyncEnabled = false
	close(cm.stopChan)
	close(cm.workQueue)
}

// chunkGenerationWorker processes chunk generation requests.
func (cm *Manager) chunkGenerationWorker() {
	for {
		select {
		case <-cm.stopChan:
			return
		case req, ok := <-cm.workQueue:
			if !ok {
				return
			}
			cm.generateAndStoreChunk(req.X, req.Y)
		}
	}
}

// generateAndStoreChunk generates a chunk and stores it in the cache.
func (cm *Manager) generateAndStoreChunk(x, y int) {
	key := [2]int{x, y}

	// Check if already loaded
	cm.mu.RLock()
	if _, ok := cm.loaded[key]; ok {
		cm.mu.RUnlock()
		cm.markPendingComplete(key)
		return
	}
	cm.mu.RUnlock()

	// Generate the chunk
	chunkSeed := mixChunkSeed(cm.Seed, x, y)
	c := NewChunkWithNoiseType(x, y, cm.ChunkSize, chunkSeed, cm.NoiseType)

	// Store it
	cm.mu.Lock()
	// Double-check under write lock
	if _, ok := cm.loaded[key]; !ok {
		cm.loaded[key] = c
	}
	cm.mu.Unlock()

	cm.markPendingComplete(key)
}

// markPendingComplete removes a chunk from the pending map.
func (cm *Manager) markPendingComplete(key [2]int) {
	cm.pendingMu.Lock()
	delete(cm.pending, key)
	cm.pendingMu.Unlock()
}

// RequestChunkAsync queues a chunk for async generation if not already loaded.
// Returns immediately. Use GetChunk or GetChunkOrPlaceholder to retrieve.
func (cm *Manager) RequestChunkAsync(x, y int) {
	if !cm.asyncEnabled {
		return
	}

	key := [2]int{x, y}

	// Check if already loaded
	cm.mu.RLock()
	if _, ok := cm.loaded[key]; ok {
		cm.mu.RUnlock()
		return
	}
	cm.mu.RUnlock()

	// Check if already pending
	cm.pendingMu.Lock()
	if cm.pending[key] {
		cm.pendingMu.Unlock()
		return
	}
	cm.pending[key] = true
	cm.pendingMu.Unlock()

	// Queue for generation
	select {
	case cm.workQueue <- chunkGenRequest{X: x, Y: y}:
	default:
		// Queue full, mark as not pending so it can be retried
		cm.pendingMu.Lock()
		delete(cm.pending, key)
		cm.pendingMu.Unlock()
	}
}

// GetChunkOrPlaceholder returns the chunk if loaded, or a placeholder if pending generation.
// The placeholder is a flat, average-height chunk using the biome color.
func (cm *Manager) GetChunkOrPlaceholder(x, y int) (*Chunk, bool) {
	key := [2]int{x, y}

	cm.mu.RLock()
	if c, ok := cm.loaded[key]; ok {
		cm.mu.RUnlock()
		return c, true // true = real chunk
	}
	cm.mu.RUnlock()

	// If async is enabled, request generation
	if cm.asyncEnabled {
		cm.RequestChunkAsync(x, y)
	}

	// Return a placeholder
	return cm.createPlaceholderChunk(x, y), false // false = placeholder
}

// createPlaceholderChunk creates a minimal placeholder chunk for display while waiting.
func (cm *Manager) createPlaceholderChunk(x, y int) *Chunk {
	size := cm.ChunkSize
	chunkSeed := mixChunkSeed(cm.Seed, x, y)

	// Create flat heightmap at 0.5 (middle height)
	heightMap := make([]float64, size*size)
	elevationMap := make([]float64, size*size)
	terrainTypes := make([]int, size*size)
	biomeMap := make([]float64, size*size)

	for i := range heightMap {
		heightMap[i] = 0.5
		elevationMap[i] = MaxElevation * 0.25 // Quarter elevation
		terrainTypes[i] = TerrainFlat
		biomeMap[i] = 0.5
	}

	return &Chunk{
		X:            x,
		Y:            y,
		Size:         size,
		Seed:         chunkSeed,
		HeightMap:    heightMap,
		ElevationMap: elevationMap,
		TerrainTypes: terrainTypes,
		BiomeMap:     biomeMap,
		DetailSpawns: nil, // No details in placeholder
	}
}

// IsChunkPending returns true if a chunk is queued for generation.
func (cm *Manager) IsChunkPending(x, y int) bool {
	cm.pendingMu.Lock()
	defer cm.pendingMu.Unlock()
	return cm.pending[[2]int{x, y}]
}

// IsChunkLoaded returns true if a chunk is loaded in the cache.
func (cm *Manager) IsChunkLoaded(x, y int) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, ok := cm.loaded[[2]int{x, y}]
	return ok
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
	c := NewChunkWithNoiseType(x, y, cm.ChunkSize, chunkSeed, cm.NoiseType)
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

// ========== Biome Blending ==========

// BiomeBlendWidth is the number of cells from chunk edge where blending occurs.
const BiomeBlendWidth = 32

// neighborValueFunc retrieves a value from a neighboring chunk at given coordinates.
type neighborValueFunc func(chunkX, chunkY, localX, localY int) float64

// blendValueAtEdges performs edge-blending for any chunk value type using the provided
// neighbor lookup function. This consolidates the common blending algorithm used by
// biome, height, and similar per-cell values that need smooth chunk transitions.
func (cm *Manager) blendValueAtEdges(chunkX, chunkY, localX, localY int, baseValue, defaultValue float64, getNeighbor neighborValueFunc) float64 {
	distFromLeft := localX
	distFromRight := cm.ChunkSize - 1 - localX
	distFromTop := localY
	distFromBottom := cm.ChunkSize - 1 - localY

	// If not in blend zone, return base value
	if distFromLeft >= BiomeBlendWidth && distFromRight >= BiomeBlendWidth &&
		distFromTop >= BiomeBlendWidth && distFromBottom >= BiomeBlendWidth {
		return baseValue
	}

	blendedValue := baseValue
	totalWeight := 1.0

	// Blend with left neighbor
	if distFromLeft < BiomeBlendWidth {
		weight := smoothstep(float64(BiomeBlendWidth-distFromLeft) / float64(BiomeBlendWidth))
		neighborValue := getNeighbor(chunkX-1, chunkY, cm.ChunkSize-1-distFromLeft, localY)
		blendedValue, totalWeight = blendValues(blendedValue, neighborValue, totalWeight, weight)
	}

	// Blend with right neighbor
	if distFromRight < BiomeBlendWidth {
		weight := smoothstep(float64(BiomeBlendWidth-distFromRight) / float64(BiomeBlendWidth))
		neighborValue := getNeighbor(chunkX+1, chunkY, BiomeBlendWidth-1-distFromRight, localY)
		blendedValue, totalWeight = blendValues(blendedValue, neighborValue, totalWeight, weight)
	}

	// Blend with top neighbor
	if distFromTop < BiomeBlendWidth {
		weight := smoothstep(float64(BiomeBlendWidth-distFromTop) / float64(BiomeBlendWidth))
		neighborValue := getNeighbor(chunkX, chunkY-1, localX, cm.ChunkSize-1-distFromTop)
		blendedValue, totalWeight = blendValues(blendedValue, neighborValue, totalWeight, weight)
	}

	// Blend with bottom neighbor
	if distFromBottom < BiomeBlendWidth {
		weight := smoothstep(float64(BiomeBlendWidth-distFromBottom) / float64(BiomeBlendWidth))
		neighborValue := getNeighbor(chunkX, chunkY+1, localX, BiomeBlendWidth-1-distFromBottom)
		blendedValue, totalWeight = blendValues(blendedValue, neighborValue, totalWeight, weight)
	}

	return blendedValue / totalWeight
}

// GetBlendedBiomeValue returns a biome value that blends with neighboring chunks.
// This creates smooth biome transitions at chunk boundaries.
func (cm *Manager) GetBlendedBiomeValue(chunkX, chunkY, localX, localY int) float64 {
	chunk := cm.GetChunk(chunkX, chunkY)
	if chunk == nil {
		return 0.5
	}
	baseBiome := chunk.GetBiomeValue(localX, localY)
	return cm.blendValueAtEdges(chunkX, chunkY, localX, localY, baseBiome, 0.5, cm.getNeighborBiomeValue)
}

// getNeighborBiomeValue retrieves biome value from a neighboring chunk.
func (cm *Manager) getNeighborBiomeValue(chunkX, chunkY, localX, localY int) float64 {
	chunk := cm.GetChunk(chunkX, chunkY)
	if chunk == nil {
		return 0.5
	}
	return chunk.GetBiomeValue(localX, localY)
}

// blendValues adds a weighted contribution to the running blend.
func blendValues(current, neighbor, totalWeight, weight float64) (float64, float64) {
	return current + neighbor*weight, totalWeight + weight
}

// smoothstep provides smooth interpolation for blend factor (t in [0,1]).
func smoothstep(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	return t * t * (3 - 2*t)
}

// GetBlendedHeight returns a height value that blends with neighboring chunks.
func (cm *Manager) GetBlendedHeight(chunkX, chunkY, localX, localY int) float64 {
	chunk := cm.GetChunk(chunkX, chunkY)
	if chunk == nil {
		return 0
	}
	baseHeight := chunk.GetHeight(localX, localY)
	return cm.blendValueAtEdges(chunkX, chunkY, localX, localY, baseHeight, 0.5, cm.getNeighborHeight)
}

// getNeighborHeight retrieves height value from a neighboring chunk.
func (cm *Manager) getNeighborHeight(chunkX, chunkY, localX, localY int) float64 {
	chunk := cm.GetChunk(chunkX, chunkY)
	if chunk == nil {
		return 0.5
	}
	return chunk.GetHeight(localX, localY)
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
	if !mc.isValidPosition(mod.X, mod.Y) {
		return
	}

	idx := mod.Y*mc.Size + mod.X
	mc.storeOriginalValues(&mod, idx)
	mc.applyModificationType(mod)
	mc.recalculateDerivedTerrain(mod.X, mod.Y)
	mc.recordModification(mod, idx)
}

// isValidPosition checks if coordinates are within chunk bounds.
func (mc *ModifiedChunk) isValidPosition(x, y int) bool {
	return x >= 0 && x < mc.Size && y >= 0 && y < mc.Size
}

// storeOriginalValues saves original values for potential undo.
func (mc *ModifiedChunk) storeOriginalValues(mod *TerrainModification, idx int) {
	mod.OldHeight = mc.HeightMap[idx]
	if mc.ElevationMap != nil {
		mod.OldElevation = mc.ElevationMap[idx]
	}
	if mc.TerrainTypes != nil {
		mod.OldTerrainType = mc.TerrainTypes[idx]
	}
}

// applyModificationType applies the terrain modification based on type.
func (mc *ModifiedChunk) applyModificationType(mod TerrainModification) {
	idx := mod.Y*mc.Size + mod.X
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
}

// recalculateDerivedTerrain updates elevation and terrain type at a position.
func (mc *ModifiedChunk) recalculateDerivedTerrain(x, y int) {
	mc.recalculateElevation(x, y)
	mc.recalculateTerrainType(x, y)
}

// recordModification tracks the modification for history.
func (mc *ModifiedChunk) recordModification(mod TerrainModification, idx int) {
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
	biome := 0.0
	if mc.BiomeMap != nil && idx < len(mc.BiomeMap) {
		biome = mc.BiomeMap[idx]
	}
	mc.TerrainTypes[idx] = classifyTerrain(x, y, mc.Size, height, biome, mc.ElevationMap)
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

// GetChunkLODAsync returns a chunk at the appropriate LOD, using placeholder if not yet loaded.
// Returns the LODChunk and a boolean indicating if it's a real chunk (true) or placeholder (false).
func (lm *LODManager) GetChunkLODAsync(chunkX, chunkY int) (*LODChunk, bool) {
	chunk, isReal := lm.manager.GetChunkOrPlaceholder(chunkX, chunkY)

	// Calculate chunk center in world coordinates
	chunkCenterX := float64(chunkX*lm.chunkSize) + float64(lm.chunkSize)/2
	chunkCenterY := float64(chunkY*lm.chunkSize) + float64(lm.chunkSize)/2

	// Distance squared from viewpoint to chunk center
	dx := chunkCenterX - lm.viewX
	dy := chunkCenterY - lm.viewY
	distSq := dx*dx + dy*dy

	level := CalculateLODLevel(distSq)
	return lm.cache.GetLOD(chunk, level), isReal
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
