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

func TestChunkManagerGetChunk(t *testing.T) {
	cm := NewChunkManager(16, 12345)

	c1 := cm.GetChunk(0, 0)
	if c1 == nil {
		t.Fatal("GetChunk returned nil")
	}
	if c1.X != 0 || c1.Y != 0 {
		t.Errorf("expected chunk at (0,0), got (%d,%d)", c1.X, c1.Y)
	}
}

func TestChunkManagerCaching(t *testing.T) {
	cm := NewChunkManager(16, 12345)

	c1 := cm.GetChunk(5, 10)
	c2 := cm.GetChunk(5, 10)

	// Should return the same cached chunk
	if c1 != c2 {
		t.Error("GetChunk should return cached chunk")
	}
}

func TestChunkManagerDifferentCoordinates(t *testing.T) {
	cm := NewChunkManager(16, 12345)

	c1 := cm.GetChunk(0, 0)
	c2 := cm.GetChunk(1, 0)

	if c1 == c2 {
		t.Error("different coordinates should return different chunks")
	}
}

func TestChunkManagerSeedMixing(t *testing.T) {
	cm := NewChunkManager(16, 12345)

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

func TestChunkManagerUnloadChunk(t *testing.T) {
	cm := NewChunkManager(16, 12345)

	_ = cm.GetChunk(0, 0)
	if cm.LoadedCount() != 1 {
		t.Errorf("expected 1 loaded chunk, got %d", cm.LoadedCount())
	}

	cm.UnloadChunk(0, 0)
	if cm.LoadedCount() != 0 {
		t.Errorf("expected 0 loaded chunks after unload, got %d", cm.LoadedCount())
	}
}

func TestChunkManagerLoadedCount(t *testing.T) {
	cm := NewChunkManager(16, 12345)

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

func BenchmarkChunkManagerGetChunk(b *testing.B) {
	cm := NewChunkManager(64, 12345)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cm.GetChunk(i%100, i/100)
	}
}
