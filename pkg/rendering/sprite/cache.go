package sprite

import (
	"container/list"
	"sync"
)

// Cache size defaults.
const (
	// DefaultMaxSheets is the default maximum number of cached sprite sheets.
	DefaultMaxSheets = 256
	// DefaultMaxMemory is the default maximum memory in bytes (20 MB).
	DefaultMaxMemory = 20 * 1024 * 1024
)

// SpriteCacheKey uniquely identifies a sprite sheet for caching.
// Two entities with identical keys share the same generated sprite sheet.
type SpriteCacheKey struct {
	// Category is the sprite type (humanoid, creature, vehicle, object, effect).
	Category string
	// BodyPlan is the silhouette template within the category.
	BodyPlan string
	// GenreID determines color palette and visual style.
	GenreID string
	// PrimaryColor is the main color (packed RGBA).
	PrimaryColor uint32
	// SecondaryColor is the accent color (packed RGBA).
	SecondaryColor uint32
	// AccentColor is the detail color (packed RGBA).
	AccentColor uint32
	// Scale affects the generated dimensions.
	Scale float64
	// Seed is the deterministic generation seed.
	Seed int64
}

// cacheEntry holds a sprite sheet and its LRU list element.
type cacheEntry struct {
	key     SpriteCacheKey
	sheet   *SpriteSheet
	element *list.Element
}

// SpriteCache provides LRU-cached access to generated sprite sheets.
// It is safe for concurrent access.
type SpriteCache struct {
	mu      sync.RWMutex
	cache   map[SpriteCacheKey]*cacheEntry
	lru     *list.List // front = most recently used
	maxSize int        // maximum number of sprite sheets
	maxMem  int64      // maximum memory in bytes
	curMem  int64      // current memory usage
	hits    int64      // cache hit count (for stats)
	misses  int64      // cache miss count (for stats)
	evicts  int64      // eviction count (for stats)
}

// NewSpriteCache creates a new sprite cache with the given limits.
// If maxSheets or maxMemory are <= 0, defaults are used.
func NewSpriteCache(maxSheets int, maxMemory int64) *SpriteCache {
	if maxSheets <= 0 {
		maxSheets = DefaultMaxSheets
	}
	if maxMemory <= 0 {
		maxMemory = DefaultMaxMemory
	}
	return &SpriteCache{
		cache:   make(map[SpriteCacheKey]*cacheEntry),
		lru:     list.New(),
		maxSize: maxSheets,
		maxMem:  maxMemory,
		curMem:  0,
	}
}

// Get retrieves a sprite sheet from the cache.
// Returns nil if not found. Marks the entry as recently used.
func (c *SpriteCache) Get(key SpriteCacheKey) *SpriteSheet {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.cache[key]
	if !ok {
		c.misses++
		return nil
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(entry.element)
	c.hits++
	return entry.sheet
}

// Put adds a sprite sheet to the cache.
// If the cache is full, the least recently used entries are evicted.
func (c *SpriteCache) Put(key SpriteCacheKey, sheet *SpriteSheet) {
	if sheet == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if entry, ok := c.cache[key]; ok {
		// Update existing entry
		c.curMem -= entry.sheet.MemorySize()
		entry.sheet = sheet
		c.curMem += sheet.MemorySize()
		c.lru.MoveToFront(entry.element)
		return
	}

	// Calculate new sheet memory
	sheetMem := sheet.MemorySize()

	// Evict if necessary
	c.evictIfNeeded(sheetMem)

	// Add new entry
	entry := &cacheEntry{
		key:   key,
		sheet: sheet,
	}
	entry.element = c.lru.PushFront(entry)
	c.cache[key] = entry
	c.curMem += sheetMem
}

// evictIfNeeded removes LRU entries until there's room for a new sheet.
// Caller must hold the write lock.
func (c *SpriteCache) evictIfNeeded(newSheetMem int64) {
	// Evict while over count limit
	for len(c.cache) >= c.maxSize {
		c.evictLRU()
	}
	// Evict while over memory limit
	for c.curMem+newSheetMem > c.maxMem && c.lru.Len() > 0 {
		c.evictLRU()
	}
}

// evictLRU removes the least recently used entry.
// Caller must hold the write lock.
func (c *SpriteCache) evictLRU() {
	elem := c.lru.Back()
	if elem == nil {
		return
	}

	entry := elem.Value.(*cacheEntry)
	c.lru.Remove(elem)
	delete(c.cache, entry.key)
	c.curMem -= entry.sheet.MemorySize()
	c.evicts++
}

// GetOrGenerate retrieves a cached sprite sheet or generates a new one.
// The generator function is called only if the key is not cached.
func (c *SpriteCache) GetOrGenerate(key SpriteCacheKey, generator func(SpriteCacheKey) *SpriteSheet) *SpriteSheet {
	// Try cache first (read lock)
	sheet := c.Get(key)
	if sheet != nil {
		return sheet
	}

	// Generate and cache (write lock in Put)
	if generator != nil {
		sheet = generator(key)
		if sheet != nil {
			c.Put(key, sheet)
		}
	}
	return sheet
}

// Size returns the number of cached sprite sheets.
func (c *SpriteCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// MemoryUsage returns the current memory usage in bytes.
func (c *SpriteCache) MemoryUsage() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.curMem
}

// Stats returns cache statistics (hits, misses, evictions).
func (c *SpriteCache) Stats() (hits, misses, evicts int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, c.evicts
}

// Clear removes all entries from the cache.
func (c *SpriteCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[SpriteCacheKey]*cacheEntry)
	c.lru = list.New()
	c.curMem = 0
}

// Contains checks if a key is in the cache without updating LRU order.
func (c *SpriteCache) Contains(key SpriteCacheKey) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.cache[key]
	return ok
}

// Keys returns all keys currently in the cache.
func (c *SpriteCache) Keys() []SpriteCacheKey {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]SpriteCacheKey, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	return keys
}
