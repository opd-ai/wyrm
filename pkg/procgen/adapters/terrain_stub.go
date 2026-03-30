//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import "math/rand"

// TerrainAdapter wraps Venture's terrain generator.
// Stub implementation for headless testing.
type TerrainAdapter struct{}

// NewTerrainAdapter creates a new terrain adapter.
func NewTerrainAdapter() *TerrainAdapter { return &TerrainAdapter{} }

// ChunkTerrainData holds generated chunk terrain data.
type ChunkTerrainData struct {
	Width  int
	Height int
	Tiles  [][]int
	Biomes [][]BiomeType
	Seed   int64
}

// RoomData holds room information within terrain.
type RoomData struct {
	X, Y   int
	Width  int
	Height int
}

// GenerateChunkTerrain creates terrain for a chunk.
func (a *TerrainAdapter) GenerateChunkTerrain(seed int64, genre string, width, height int) (*ChunkTerrainData, error) {
	rng := rand.New(rand.NewSource(seed))
	tiles := make([][]int, height)
	biomes := make([][]BiomeType, height)

	dist := GetGenreBiomeDistribution(genre)
	primaryBiome := selectBiomeFromWeights(seed, dist)

	for y := 0; y < height; y++ {
		tiles[y] = make([]int, width)
		biomes[y] = make([]BiomeType, width)
		for x := 0; x < width; x++ {
			tileType := 1 // floor
			if rng.Float64() < 0.1 {
				tileType = 0 // wall
			} else if rng.Float64() < 0.05 {
				tileType = 4 // water
			}
			tiles[y][x] = tileType
			biomes[y][x] = determineBiome(tileType, 0, primaryBiome, dist)
		}
	}

	return &ChunkTerrainData{
		Width:  width,
		Height: height,
		Tiles:  tiles,
		Biomes: biomes,
		Seed:   seed,
	}, nil
}

// selectBiomeFromWeights selects a biome based on weighted distribution.
func selectBiomeFromWeights(seed int64, dist *GenreBiomeDistribution) BiomeType {
	rng := rand.New(rand.NewSource(seed))
	roll := rng.Float64()
	cumulative := 0.0
	for _, biome := range dist.PrimaryBiomes {
		cumulative += dist.Weights[biome]
		if roll < cumulative {
			return biome
		}
	}
	if len(dist.PrimaryBiomes) > 0 {
		return dist.PrimaryBiomes[0]
	}
	return BiomeForest
}

// determineBiome determines the biome for a specific tile.
func determineBiome(tileType, height int, primaryBiome BiomeType, dist *GenreBiomeDistribution) BiomeType {
	// Water tiles become lake
	if tileType == 4 {
		return BiomeLake
	}
	// High ground becomes mountain
	if height > 1 {
		return BiomeMountain
	}
	return primaryBiome
}

// IsWalkable checks if a tile type is walkable.
func IsWalkable(tileType int) bool {
	return tileType == 1 || tileType == 2 || tileType == 3 // floor, door, corridor
}

// GetTileMovementCost returns movement cost for a tile.
func GetTileMovementCost(tileType int) float64 {
	switch tileType {
	case 0: // wall
		return 999.0
	case 4: // water
		return 2.0
	default:
		return 1.0
	}
}
