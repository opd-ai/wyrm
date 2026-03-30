package chunk

import (
	"testing"
)

func TestNewChunk(t *testing.T) {
	c := NewChunk(5, 10, 16, 12345)

	if c.X != 5 {
		t.Errorf("expected X=5, got %d", c.X)
	}
	if c.Y != 10 {
		t.Errorf("expected Y=10, got %d", c.Y)
	}
	if c.Size != 16 {
		t.Errorf("expected Size=16, got %d", c.Size)
	}
	if c.Seed != 12345 {
		t.Errorf("expected Seed=12345, got %d", c.Seed)
	}
	if len(c.HeightMap) != 16*16 {
		t.Errorf("expected HeightMap size 256, got %d", len(c.HeightMap))
	}
}

func TestChunkHeightMapPopulated(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Check that heightmap has non-zero values
	hasNonZero := false
	for _, h := range c.HeightMap {
		if h != 0 {
			hasNonZero = true
			break
		}
	}

	if !hasNonZero {
		t.Error("heightmap should be populated with non-zero values")
	}
}

func TestChunkDeterminism(t *testing.T) {
	seed := int64(42)

	c1 := NewChunk(0, 0, 16, seed)
	c2 := NewChunk(0, 0, 16, seed)

	// Same seed should produce identical heightmaps
	for i := range c1.HeightMap {
		if c1.HeightMap[i] != c2.HeightMap[i] {
			t.Errorf("heightmap mismatch at index %d: %f != %f", i, c1.HeightMap[i], c2.HeightMap[i])
		}
	}
}

func TestChunkDifferentSeeds(t *testing.T) {
	c1 := NewChunk(0, 0, 16, 12345)
	c2 := NewChunk(0, 0, 16, 54321)

	// Different seeds should produce different heightmaps
	same := true
	for i := range c1.HeightMap {
		if c1.HeightMap[i] != c2.HeightMap[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("different seeds should produce different heightmaps")
	}
}

func TestChunkGetHeight(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Valid coordinates
	h := c.GetHeight(0, 0)
	if h < 0 || h > 1 {
		t.Errorf("height should be in [0,1] range, got %f", h)
	}

	h = c.GetHeight(15, 15)
	if h < 0 || h > 1 {
		t.Errorf("height should be in [0,1] range, got %f", h)
	}

	// Invalid coordinates should return 0
	h = c.GetHeight(-1, 0)
	if h != 0 {
		t.Errorf("expected 0 for invalid coordinates, got %f", h)
	}

	h = c.GetHeight(0, -1)
	if h != 0 {
		t.Errorf("expected 0 for invalid coordinates, got %f", h)
	}

	h = c.GetHeight(16, 0)
	if h != 0 {
		t.Errorf("expected 0 for out-of-bounds coordinates, got %f", h)
	}

	h = c.GetHeight(0, 16)
	if h != 0 {
		t.Errorf("expected 0 for out-of-bounds coordinates, got %f", h)
	}
}

func TestManagerGetChunk(t *testing.T) {
	cm := NewManager(16, 12345)

	c1 := cm.GetChunk(0, 0)
	if c1 == nil {
		t.Fatal("GetChunk returned nil")
	}
	if c1.X != 0 || c1.Y != 0 {
		t.Errorf("expected chunk at (0,0), got (%d,%d)", c1.X, c1.Y)
	}
}

func TestManagerCaching(t *testing.T) {
	cm := NewManager(16, 12345)

	c1 := cm.GetChunk(5, 10)
	c2 := cm.GetChunk(5, 10)

	// Should return the same cached chunk
	if c1 != c2 {
		t.Error("GetChunk should return cached chunk")
	}
}

func TestManagerDifferentCoordinates(t *testing.T) {
	cm := NewManager(16, 12345)

	c1 := cm.GetChunk(0, 0)
	c2 := cm.GetChunk(1, 0)

	if c1 == c2 {
		t.Error("different coordinates should return different chunks")
	}
}

func TestManagerSeedMixing(t *testing.T) {
	cm := NewManager(16, 12345)

	c00 := cm.GetChunk(0, 0)
	c01 := cm.GetChunk(0, 1)
	c10 := cm.GetChunk(1, 0)

	// Each chunk should have a unique seed derived from coordinates
	if c00.Seed == c01.Seed {
		t.Error("(0,0) and (0,1) should have different seeds")
	}
	if c00.Seed == c10.Seed {
		t.Error("(0,0) and (1,0) should have different seeds")
	}
	if c01.Seed == c10.Seed {
		t.Error("(0,1) and (1,0) should have different seeds")
	}
}

func TestManagerUnloadChunk(t *testing.T) {
	cm := NewManager(16, 12345)

	_ = cm.GetChunk(0, 0)
	if cm.LoadedCount() != 1 {
		t.Errorf("expected 1 loaded chunk, got %d", cm.LoadedCount())
	}

	cm.UnloadChunk(0, 0)
	if cm.LoadedCount() != 0 {
		t.Errorf("expected 0 loaded chunks after unload, got %d", cm.LoadedCount())
	}
}

func TestManagerLoadedCount(t *testing.T) {
	cm := NewManager(16, 12345)

	if cm.LoadedCount() != 0 {
		t.Errorf("expected 0 loaded chunks initially, got %d", cm.LoadedCount())
	}

	_ = cm.GetChunk(0, 0)
	_ = cm.GetChunk(1, 0)
	_ = cm.GetChunk(0, 1)

	if cm.LoadedCount() != 3 {
		t.Errorf("expected 3 loaded chunks, got %d", cm.LoadedCount())
	}
}

func TestMixChunkSeedDeterminism(t *testing.T) {
	baseSeed := int64(12345)

	s1 := mixChunkSeed(baseSeed, 10, 20)
	s2 := mixChunkSeed(baseSeed, 10, 20)

	if s1 != s2 {
		t.Error("mixChunkSeed should be deterministic")
	}
}

func TestMixChunkSeedUniqueness(t *testing.T) {
	baseSeed := int64(12345)

	s00 := mixChunkSeed(baseSeed, 0, 0)
	s01 := mixChunkSeed(baseSeed, 0, 1)
	s10 := mixChunkSeed(baseSeed, 1, 0)
	s11 := mixChunkSeed(baseSeed, 1, 1)

	seeds := []int64{s00, s01, s10, s11}
	seen := make(map[int64]bool)
	for _, s := range seeds {
		if seen[s] {
			t.Error("mixChunkSeed produced duplicate seeds for different coordinates")
		}
		seen[s] = true
	}
}

func BenchmarkNewChunk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewChunk(0, 0, 64, int64(i))
	}
}

func BenchmarkManagerGetChunk(b *testing.B) {
	cm := NewManager(64, 12345)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cm.GetChunk(i%100, i/100)
	}
}

// Tests for vertical terrain (hills, cliffs)

func TestChunkElevationMapPopulated(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	if c.ElevationMap == nil {
		t.Fatal("ElevationMap should not be nil")
	}
	if len(c.ElevationMap) != 16*16 {
		t.Errorf("expected ElevationMap size 256, got %d", len(c.ElevationMap))
	}

	// Check that elevation values are in valid range
	for _, e := range c.ElevationMap {
		if e < 0 || e > MaxElevation {
			t.Errorf("elevation %f out of range [0, %f]", e, MaxElevation)
		}
	}
}

func TestChunkTerrainTypesPopulated(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	if c.TerrainTypes == nil {
		t.Fatal("TerrainTypes should not be nil")
	}
	if len(c.TerrainTypes) != 16*16 {
		t.Errorf("expected TerrainTypes size 256, got %d", len(c.TerrainTypes))
	}

	// Check that terrain types are valid
	for _, tt := range c.TerrainTypes {
		if tt < TerrainFlat || tt > TerrainPeak {
			t.Errorf("invalid terrain type %d", tt)
		}
	}
}

func TestChunkGetElevation(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Valid coordinates
	e := c.GetElevation(0, 0)
	if e < 0 || e > MaxElevation {
		t.Errorf("elevation should be in [0, %f] range, got %f", MaxElevation, e)
	}

	e = c.GetElevation(15, 15)
	if e < 0 || e > MaxElevation {
		t.Errorf("elevation should be in [0, %f] range, got %f", MaxElevation, e)
	}

	// Invalid coordinates should return 0
	e = c.GetElevation(-1, 0)
	if e != 0 {
		t.Errorf("expected 0 for invalid coordinates, got %f", e)
	}

	e = c.GetElevation(16, 0)
	if e != 0 {
		t.Errorf("expected 0 for out-of-bounds coordinates, got %f", e)
	}
}

func TestChunkGetTerrainType(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Valid coordinates should return a valid type
	tt := c.GetTerrainType(8, 8)
	if tt < TerrainFlat || tt > TerrainPeak {
		t.Errorf("invalid terrain type %d", tt)
	}

	// Invalid coordinates should return TerrainFlat
	tt = c.GetTerrainType(-1, 0)
	if tt != TerrainFlat {
		t.Errorf("expected TerrainFlat for invalid coordinates, got %d", tt)
	}

	tt = c.GetTerrainType(16, 0)
	if tt != TerrainFlat {
		t.Errorf("expected TerrainFlat for out-of-bounds coordinates, got %d", tt)
	}
}

func TestChunkIsCliff(t *testing.T) {
	// Create a chunk and manually set a cliff for testing
	c := NewChunk(0, 0, 16, 12345)

	// Find a cliff cell if one exists
	hasCliff := false
	for y := 0; y < c.Size; y++ {
		for x := 0; x < c.Size; x++ {
			if c.TerrainTypes[y*c.Size+x] == TerrainCliff {
				if !c.IsCliff(x, y) {
					t.Error("IsCliff should return true for cliff terrain")
				}
				hasCliff = true
				break
			}
		}
		if hasCliff {
			break
		}
	}

	// A flat cell should not be a cliff
	for y := 0; y < c.Size; y++ {
		for x := 0; x < c.Size; x++ {
			if c.TerrainTypes[y*c.Size+x] == TerrainFlat {
				if c.IsCliff(x, y) {
					t.Error("IsCliff should return false for flat terrain")
				}
				return
			}
		}
	}
}

func TestChunkIsHill(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Test that hills and peaks return true for IsHill
	for y := 0; y < c.Size; y++ {
		for x := 0; x < c.Size; x++ {
			tt := c.TerrainTypes[y*c.Size+x]
			isHill := c.IsHill(x, y)
			expected := tt == TerrainHill || tt == TerrainPeak
			if isHill != expected {
				t.Errorf("IsHill(%d, %d) = %v, expected %v for terrain type %d", x, y, isHill, expected, tt)
			}
		}
	}
}

func TestChunkGetElevationDifference(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Same cell should have 0 difference
	diff := c.GetElevationDifference(5, 5, 5, 5)
	if diff != 0 {
		t.Errorf("expected 0 difference for same cell, got %f", diff)
	}

	// Different cells may have positive difference
	diff = c.GetElevationDifference(0, 0, 15, 15)
	if diff < 0 {
		t.Errorf("elevation difference should be non-negative, got %f", diff)
	}
}

func TestElevationDeterminism(t *testing.T) {
	seed := int64(42)

	c1 := NewChunk(0, 0, 16, seed)
	c2 := NewChunk(0, 0, 16, seed)

	// Same seed should produce identical elevation maps
	for i := range c1.ElevationMap {
		if c1.ElevationMap[i] != c2.ElevationMap[i] {
			t.Errorf("elevation mismatch at index %d: %f != %f", i, c1.ElevationMap[i], c2.ElevationMap[i])
		}
	}

	// Same seed should produce identical terrain types
	for i := range c1.TerrainTypes {
		if c1.TerrainTypes[i] != c2.TerrainTypes[i] {
			t.Errorf("terrain type mismatch at index %d: %d != %d", i, c1.TerrainTypes[i], c2.TerrainTypes[i])
		}
	}
}

func TestTerrainTypeDistribution(t *testing.T) {
	// Test that terrain types are distributed across the chunk
	c := NewChunk(0, 0, 64, 12345)

	counts := make(map[int]int)
	for _, tt := range c.TerrainTypes {
		counts[tt]++
	}

	// With a 64x64 chunk, we should have some variation
	// At minimum, we should have flat terrain
	if counts[TerrainFlat] == 0 {
		t.Error("expected some flat terrain cells")
	}

	// Log distribution for debugging (not an error if others are 0)
	t.Logf("Terrain distribution: Flat=%d, Hill=%d, Cliff=%d, Peak=%d",
		counts[TerrainFlat], counts[TerrainHill], counts[TerrainCliff], counts[TerrainPeak])
}

func BenchmarkNewChunkWithElevation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewChunk(0, 0, 64, int64(i))
	}
}
