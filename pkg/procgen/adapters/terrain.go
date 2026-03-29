// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

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
	Width      int
	Height     int
	Tiles      [][]int  // TileType as int for portability
	HeightMap  [][]int  // Height values for 3D rendering
	Rooms      []RoomData
	Seed       int64
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

	return convertToChunkData(ventureTerrain, seed), nil
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
