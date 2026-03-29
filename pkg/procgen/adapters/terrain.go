//go:build !noebiten

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// BiomeType represents different biome categories.
type BiomeType int

const (
	// BiomeForest represents forested areas.
	BiomeForest BiomeType = iota
	// BiomeMountain represents mountainous terrain.
	BiomeMountain
	// BiomeLake represents water bodies.
	BiomeLake
	// BiomeSwamp represents swampy wetlands.
	BiomeSwamp
	// BiomeWasteland represents barren wasteland.
	BiomeWasteland
	// BiomeUrban represents urban/city areas.
	BiomeUrban
	// BiomeIndustrial represents industrial zones.
	BiomeIndustrial
	// BiomeRuins represents ruined structures.
	BiomeRuins
	// BiomeCrater represents impact craters.
	BiomeCrater
	// BiomeTech represents high-tech structures.
	BiomeTech
)

// GenreBiomeDistribution defines biome weights for each genre.
type GenreBiomeDistribution struct {
	PrimaryBiomes   []BiomeType
	SecondaryBiomes []BiomeType
	Weights         map[BiomeType]float64
}

// genreBiomeDistributions maps genre to biome distribution.
var genreBiomeDistributions = map[string]*GenreBiomeDistribution{
	"fantasy": {
		PrimaryBiomes:   []BiomeType{BiomeForest, BiomeMountain, BiomeLake},
		SecondaryBiomes: []BiomeType{BiomeRuins},
		Weights: map[BiomeType]float64{
			BiomeForest:   0.40,
			BiomeMountain: 0.30,
			BiomeLake:     0.20,
			BiomeRuins:    0.10,
		},
	},
	"sci-fi": {
		PrimaryBiomes:   []BiomeType{BiomeCrater, BiomeTech},
		SecondaryBiomes: []BiomeType{BiomeIndustrial},
		Weights: map[BiomeType]float64{
			BiomeCrater:     0.35,
			BiomeTech:       0.35,
			BiomeIndustrial: 0.30,
		},
	},
	"horror": {
		PrimaryBiomes:   []BiomeType{BiomeSwamp, BiomeForest},
		SecondaryBiomes: []BiomeType{BiomeRuins},
		Weights: map[BiomeType]float64{
			BiomeSwamp:  0.40,
			BiomeForest: 0.35, // "Dead forests" - style applied via palette
			BiomeRuins:  0.25,
		},
	},
	"cyberpunk": {
		PrimaryBiomes:   []BiomeType{BiomeUrban, BiomeIndustrial},
		SecondaryBiomes: []BiomeType{BiomeTech},
		Weights: map[BiomeType]float64{
			BiomeUrban:      0.50,
			BiomeIndustrial: 0.35,
			BiomeTech:       0.15,
		},
	},
	"post-apocalyptic": {
		PrimaryBiomes:   []BiomeType{BiomeWasteland, BiomeRuins},
		SecondaryBiomes: []BiomeType{BiomeCrater},
		Weights: map[BiomeType]float64{
			BiomeWasteland: 0.45,
			BiomeRuins:     0.35,
			BiomeCrater:    0.20,
		},
	},
}

// GetGenreBiomeDistribution returns the biome distribution for a genre.
func GetGenreBiomeDistribution(genre string) *GenreBiomeDistribution {
	if dist, ok := genreBiomeDistributions[genre]; ok {
		return dist
	}
	// Default to fantasy if unknown genre
	return genreBiomeDistributions["fantasy"]
}

// TerrainAdapter wraps Venture's terrain generators for Wyrm's chunk system.
type TerrainAdapter struct {
	generator *terrain.BSPGenerator
}

// NewTerrainAdapter creates a new terrain adapter using BSP generation.
func NewTerrainAdapter() *TerrainAdapter {
	return &TerrainAdapter{
		generator: terrain.NewBSPGenerator(),
	}
}

// ChunkTerrainData holds generated terrain for a chunk.
type ChunkTerrainData struct {
	Width        int
	Height       int
	Tiles        [][]int // TileType as int for portability
	HeightMap    [][]int // Height values for 3D rendering
	BiomeMap     [][]BiomeType
	Rooms        []RoomData
	Seed         int64
	Genre        string
	PrimaryBiome BiomeType
}

// RoomData holds room information from generation.
type RoomData struct {
	X, Y          int
	Width, Height int
	Type          string
}

// GenerateChunkTerrain generates terrain data for a world chunk.
func (a *TerrainAdapter) GenerateChunkTerrain(seed int64, genre string, width, height int) (*ChunkTerrainData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: 0.5,
		Custom: map[string]interface{}{
			"width":  width,
			"height": height,
		},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("terrain generation failed: %w", err)
	}

	ventureTerrain, ok := result.(*terrain.Terrain)
	if !ok {
		return nil, fmt.Errorf("invalid terrain result type: expected *terrain.Terrain, got %T", result)
	}

	chunkData := convertToChunkData(ventureTerrain, seed)
	chunkData.Genre = genre

	// Apply genre-specific biome distribution
	applyGenreBiomes(chunkData, seed, genre)

	return chunkData, nil
}

// applyGenreBiomes applies genre-specific biome distribution to chunk data.
func applyGenreBiomes(chunk *ChunkTerrainData, seed int64, genre string) {
	dist := GetGenreBiomeDistribution(genre)

	// Initialize biome map
	chunk.BiomeMap = make([][]BiomeType, chunk.Height)
	for y := 0; y < chunk.Height; y++ {
		chunk.BiomeMap[y] = make([]BiomeType, chunk.Width)
	}

	// Determine primary biome for this chunk based on seed
	chunk.PrimaryBiome = selectBiomeFromWeights(seed, dist)

	// Apply biomes based on tile types and genre weights
	for y := 0; y < chunk.Height; y++ {
		for x := 0; x < chunk.Width; x++ {
			tileType := chunk.Tiles[y][x]
			height := chunk.HeightMap[y][x]
			chunk.BiomeMap[y][x] = determineBiome(tileType, height, chunk.PrimaryBiome, dist)
		}
	}
}

// selectBiomeFromWeights selects a biome based on weighted random from seed.
func selectBiomeFromWeights(seed int64, dist *GenreBiomeDistribution) BiomeType {
	// Use seed to generate a deterministic value between 0 and 1
	seedVal := float64(seed%10000) / 10000.0

	// Iterate in deterministic order: primary biomes first, then secondary
	// This ensures deterministic selection despite Go's randomized map iteration
	allBiomes := append([]BiomeType{}, dist.PrimaryBiomes...)
	allBiomes = append(allBiomes, dist.SecondaryBiomes...)

	cumulative := 0.0
	for _, biome := range allBiomes {
		weight, ok := dist.Weights[biome]
		if !ok {
			continue
		}
		cumulative += weight
		if seedVal < cumulative {
			return biome
		}
	}

	// Fallback to first primary biome
	if len(dist.PrimaryBiomes) > 0 {
		return dist.PrimaryBiomes[0]
	}
	return BiomeForest
}

// determineBiome determines the biome for a specific tile.
func determineBiome(tileType, height int, primaryBiome BiomeType, dist *GenreBiomeDistribution) BiomeType {
	// Water tiles get lake/swamp based on genre
	if tileType == int(terrain.TileWaterShallow) || tileType == int(terrain.TileWaterDeep) {
		if _, ok := dist.Weights[BiomeSwamp]; ok {
			return BiomeSwamp
		}
		return BiomeLake
	}

	// High ground gets mountain/crater based on genre
	if height >= 2 {
		if _, ok := dist.Weights[BiomeMountain]; ok {
			return BiomeMountain
		}
		if _, ok := dist.Weights[BiomeCrater]; ok {
			return BiomeCrater
		}
	}

	// Default to primary biome
	return primaryBiome
}

// convertToChunkData transforms Venture terrain to Wyrm chunk format.
func convertToChunkData(t *terrain.Terrain, seed int64) *ChunkTerrainData {
	tiles := make([][]int, t.Height)
	heightMap := make([][]int, t.Height)

	for y := 0; y < t.Height; y++ {
		tiles[y] = make([]int, t.Width)
		heightMap[y] = make([]int, t.Width)
		for x := 0; x < t.Width; x++ {
			tiles[y][x] = int(t.Tiles[y][x])
			heightMap[y][x] = calculateHeight(t.Tiles[y][x])
		}
	}

	rooms := make([]RoomData, len(t.Rooms))
	for i, room := range t.Rooms {
		rooms[i] = RoomData{
			X:      room.X,
			Y:      room.Y,
			Width:  room.Width,
			Height: room.Height,
			Type:   room.Type.String(),
		}
	}

	return &ChunkTerrainData{
		Width:     t.Width,
		Height:    t.Height,
		Tiles:     tiles,
		HeightMap: heightMap,
		Rooms:     rooms,
		Seed:      seed,
	}
}

// calculateHeight determines height from tile type for 3D raycaster.
func calculateHeight(tile terrain.TileType) int {
	switch tile {
	case terrain.TileWall:
		return 2
	case terrain.TileFloor, terrain.TileCorridor, terrain.TileDoor:
		return 0
	case terrain.TilePlatform:
		return 1
	case terrain.TileWaterShallow:
		return -1
	case terrain.TileWaterDeep:
		return -2
	default:
		return 0
	}
}

// IsWalkable returns true if the tile type is walkable.
func IsWalkable(tileType int) bool {
	return terrain.TileType(tileType).IsWalkableTile()
}

// GetTileMovementCost returns movement cost for a tile type.
func GetTileMovementCost(tileType int) float64 {
	return terrain.TileType(tileType).MovementCost()
}
