package sprite

import (
	"sync"
	"testing"
)

func TestNewSpriteCache(t *testing.T) {
	t.Run("with defaults", func(t *testing.T) {
		cache := NewSpriteCache(0, 0)
		if cache.maxSize != DefaultMaxSheets {
			t.Errorf("expected max size %d, got %d", DefaultMaxSheets, cache.maxSize)
		}
		if cache.maxMem != DefaultMaxMemory {
			t.Errorf("expected max memory %d, got %d", DefaultMaxMemory, cache.maxMem)
		}
	})

	t.Run("with custom values", func(t *testing.T) {
		cache := NewSpriteCache(100, 1024*1024)
		if cache.maxSize != 100 {
			t.Errorf("expected max size 100, got %d", cache.maxSize)
		}
		if cache.maxMem != 1024*1024 {
			t.Errorf("expected max memory 1MB, got %d", cache.maxMem)
		}
	})
}

func TestSpriteCacheGetPut(t *testing.T) {
	cache := NewSpriteCache(10, 1024*1024)
	key := SpriteCacheKey{
		Category: CategoryHumanoid,
		BodyPlan: "warrior",
		GenreID:  "fantasy",
		Seed:     12345,
	}

	t.Run("get missing returns nil", func(t *testing.T) {
		got := cache.Get(key)
		if got != nil {
			t.Error("expected nil for missing key")
		}
	})

	t.Run("put and get", func(t *testing.T) {
		sheet := NewSpriteSheet(32, 48)
		cache.Put(key, sheet)

		got := cache.Get(key)
		if got != sheet {
			t.Error("didn't retrieve the same sheet")
		}
	})

	t.Run("put nil does nothing", func(t *testing.T) {
		sizeBefore := cache.Size()
		cache.Put(SpriteCacheKey{}, nil)
		if cache.Size() != sizeBefore {
			t.Error("nil put should not affect cache")
		}
	})
}

func TestSpriteCacheSize(t *testing.T) {
	cache := NewSpriteCache(10, 1024*1024)

	if cache.Size() != 0 {
		t.Error("new cache should be empty")
	}

	for i := 0; i < 5; i++ {
		key := SpriteCacheKey{Seed: int64(i)}
		cache.Put(key, NewSpriteSheet(32, 48))
	}

	if cache.Size() != 5 {
		t.Errorf("expected size 5, got %d", cache.Size())
	}
}

func TestSpriteCacheMemoryUsage(t *testing.T) {
	cache := NewSpriteCache(100, 1024*1024)

	sheet := NewSpriteSheet(32, 48)
	idle := NewAnimation(AnimIdle, true)
	idle.AddFrame(NewSprite(32, 48))
	sheet.AddAnimation(idle)

	key := SpriteCacheKey{Seed: 1}
	cache.Put(key, sheet)

	expectedMem := sheet.MemorySize()
	if cache.MemoryUsage() != expectedMem {
		t.Errorf("expected memory %d, got %d", expectedMem, cache.MemoryUsage())
	}
}

func TestSpriteCacheLRUEviction(t *testing.T) {
	cache := NewSpriteCache(3, 100*1024*1024) // Small count limit

	// Add 3 entries
	for i := 0; i < 3; i++ {
		key := SpriteCacheKey{Seed: int64(i)}
		cache.Put(key, NewSpriteSheet(32, 48))
	}

	// Access entry 0 to make it recently used
	cache.Get(SpriteCacheKey{Seed: 0})

	// Add entry 3 - should evict entry 1 (least recently used)
	cache.Put(SpriteCacheKey{Seed: 3}, NewSpriteSheet(32, 48))

	if cache.Size() != 3 {
		t.Errorf("expected size 3, got %d", cache.Size())
	}

	// Entry 1 should be evicted
	if cache.Contains(SpriteCacheKey{Seed: 1}) {
		t.Error("entry 1 should have been evicted")
	}

	// Entries 0, 2, 3 should still be present
	if !cache.Contains(SpriteCacheKey{Seed: 0}) {
		t.Error("entry 0 should still be present")
	}
	if !cache.Contains(SpriteCacheKey{Seed: 2}) {
		t.Error("entry 2 should still be present")
	}
	if !cache.Contains(SpriteCacheKey{Seed: 3}) {
		t.Error("entry 3 should still be present")
	}
}

func TestSpriteCacheMemoryEviction(t *testing.T) {
	// Each 32x48 sprite is 32*48*4 = 6144 bytes
	// Each sheet with one frame uses ~6KB
	cache := NewSpriteCache(100, 15000) // Allow ~2-3 sheets

	for i := 0; i < 5; i++ {
		key := SpriteCacheKey{Seed: int64(i)}
		sheet := NewSpriteSheet(32, 48)
		idle := NewAnimation(AnimIdle, true)
		idle.AddFrame(NewSprite(32, 48))
		sheet.AddAnimation(idle)
		cache.Put(key, sheet)
	}

	// Should have evicted some due to memory limit
	if cache.MemoryUsage() > 15000 {
		t.Errorf("memory usage %d exceeds limit 15000", cache.MemoryUsage())
	}
}

func TestSpriteCacheStats(t *testing.T) {
	cache := NewSpriteCache(10, 1024*1024)
	key := SpriteCacheKey{Seed: 1}

	// Miss
	cache.Get(key)

	// Put and hit
	cache.Put(key, NewSpriteSheet(32, 48))
	cache.Get(key)
	cache.Get(key)

	hits, misses, _ := cache.Stats()
	if hits != 2 {
		t.Errorf("expected 2 hits, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("expected 1 miss, got %d", misses)
	}
}

func TestSpriteCacheClear(t *testing.T) {
	cache := NewSpriteCache(10, 1024*1024)

	for i := 0; i < 5; i++ {
		key := SpriteCacheKey{Seed: int64(i)}
		cache.Put(key, NewSpriteSheet(32, 48))
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Error("cache should be empty after clear")
	}
	if cache.MemoryUsage() != 0 {
		t.Error("memory usage should be 0 after clear")
	}
}

func TestSpriteCacheContains(t *testing.T) {
	cache := NewSpriteCache(10, 1024*1024)
	key := SpriteCacheKey{Seed: 1}

	if cache.Contains(key) {
		t.Error("should not contain key before put")
	}

	cache.Put(key, NewSpriteSheet(32, 48))

	if !cache.Contains(key) {
		t.Error("should contain key after put")
	}
}

func TestSpriteCacheKeys(t *testing.T) {
	cache := NewSpriteCache(10, 1024*1024)

	for i := 0; i < 3; i++ {
		key := SpriteCacheKey{Seed: int64(i)}
		cache.Put(key, NewSpriteSheet(32, 48))
	}

	keys := cache.Keys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

func TestSpriteCacheUpdateExisting(t *testing.T) {
	cache := NewSpriteCache(10, 1024*1024)
	key := SpriteCacheKey{Seed: 1}

	sheet1 := NewSpriteSheet(32, 48)
	cache.Put(key, sheet1)

	sheet2 := NewSpriteSheet(64, 96)
	cache.Put(key, sheet2)

	// Should still be size 1
	if cache.Size() != 1 {
		t.Errorf("expected size 1, got %d", cache.Size())
	}

	// Should return the updated sheet
	got := cache.Get(key)
	if got != sheet2 {
		t.Error("should return updated sheet")
	}
}

func TestSpriteCacheGetOrGenerate(t *testing.T) {
	cache := NewSpriteCache(10, 1024*1024)
	key := SpriteCacheKey{
		Category: CategoryHumanoid,
		BodyPlan: "warrior",
		Seed:     12345,
	}

	generateCalls := 0
	generator := func(k SpriteCacheKey) *SpriteSheet {
		generateCalls++
		return NewSpriteSheet(32, 48)
	}

	// First call should generate
	sheet1 := cache.GetOrGenerate(key, generator)
	if generateCalls != 1 {
		t.Errorf("expected 1 generate call, got %d", generateCalls)
	}
	if sheet1 == nil {
		t.Error("expected non-nil sheet")
	}

	// Second call should use cache
	sheet2 := cache.GetOrGenerate(key, generator)
	if generateCalls != 1 {
		t.Errorf("expected still 1 generate call, got %d", generateCalls)
	}
	if sheet2 != sheet1 {
		t.Error("should return cached sheet")
	}
}

func TestSpriteCacheConcurrency(t *testing.T) {
	cache := NewSpriteCache(100, 10*1024*1024)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := SpriteCacheKey{Seed: int64(id % 10)}
			cache.Put(key, NewSpriteSheet(32, 48))
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := SpriteCacheKey{Seed: int64(id % 10)}
			cache.Get(key)
		}(i)
	}

	wg.Wait()

	// Should not panic and should have some entries
	if cache.Size() > 10 {
		t.Error("should have at most 10 unique entries")
	}
}
