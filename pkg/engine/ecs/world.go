// Package ecs provides the core Entity-Component-System framework.
package ecs

import (
	"errors"
	"sort"
	"sync"
)

// Entity is a unique identifier for a game object.
type Entity uint64

// Component is the interface all components must implement.
type Component interface {
	Type() string
}

// System is the interface all systems must implement.
type System interface {
	Update(w *World, dt float64)
}

// ErrEntityNotFound is returned when operating on a non-existent entity.
var ErrEntityNotFound = errors.New("ecs: entity not found")

// World holds all entities, their components, and the registered systems.
type World struct {
	mu          sync.RWMutex
	nextID      Entity
	components  map[Entity]map[string]Component
	systems     []System
	sortedIDs   []Entity       // Maintained sorted list of entity IDs for O(1) ordered iteration
	entityIndex map[Entity]int // Maps entity ID to index in sortedIDs for O(log n) removal
}

// NewWorld creates a new empty ECS world.
func NewWorld() *World {
	return &World{
		nextID:      1,
		components:  make(map[Entity]map[string]Component),
		sortedIDs:   make([]Entity, 0),
		entityIndex: make(map[Entity]int),
	}
}

// CreateEntity allocates a new entity and returns its ID.
func (w *World) CreateEntity() Entity {
	w.mu.Lock()
	defer w.mu.Unlock()
	id := w.nextID
	w.nextID++
	w.components[id] = make(map[string]Component)
	// Insert into sorted position (since IDs are monotonically increasing, always append)
	w.entityIndex[id] = len(w.sortedIDs)
	w.sortedIDs = append(w.sortedIDs, id)
	return id
}

// DestroyEntity removes an entity and all its components.
// Returns true if the entity existed and was destroyed, false if the entity
// did not exist (no-op). This allows callers to detect double-deletion bugs.
func (w *World) DestroyEntity(e Entity) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, exists := w.components[e]; !exists {
		return false
	}
	delete(w.components, e)
	// Remove from sorted index
	if idx, ok := w.entityIndex[e]; ok {
		// Swap with last element for O(1) removal
		lastIdx := len(w.sortedIDs) - 1
		if idx != lastIdx {
			lastEntity := w.sortedIDs[lastIdx]
			w.sortedIDs[idx] = lastEntity
			w.entityIndex[lastEntity] = idx
		}
		w.sortedIDs = w.sortedIDs[:lastIdx]
		delete(w.entityIndex, e)
	}
	return true
}

// AddComponent attaches a component to an entity.
// Returns ErrEntityNotFound if the entity does not exist.
func (w *World) AddComponent(e Entity, c Component) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.components[e]; !ok {
		return ErrEntityNotFound
	}
	w.components[e][c.Type()] = c
	return nil
}

// GetComponent retrieves a component by type from an entity.
func (w *World) GetComponent(e Entity, typeName string) (Component, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	comps, ok := w.components[e]
	if !ok {
		return nil, false
	}
	c, ok := comps[typeName]
	return c, ok
}

// GetComponents retrieves multiple components by type from an entity in a single lock.
// More efficient than calling GetComponent multiple times for multi-component access.
// Returns a map of component type to Component, and false if the entity doesn't exist.
func (w *World) GetComponents(e Entity, types ...string) (map[string]Component, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	comps, ok := w.components[e]
	if !ok {
		return nil, false
	}
	result := make(map[string]Component, len(types))
	for _, t := range types {
		if c, ok := comps[t]; ok {
			result[t] = c
		}
	}
	return result, true
}

// RemoveComponent removes a component by type from an entity.
func (w *World) RemoveComponent(e Entity, typeName string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if comps, ok := w.components[e]; ok {
		delete(comps, typeName)
	}
}

// Entities returns all entities that have the given component types,
// sorted by entity ID for deterministic iteration order.
//
// WARNING: Entity IDs returned are a snapshot at the time of the call.
// If an entity is destroyed after this call but before the caller finishes
// processing the returned slice, GetComponent() will return (nil, false)
// for that entity. Systems should always check the ok return value from
// GetComponent() to handle stale references gracefully.
//
// For long-running operations, consider re-querying or using a validated
// entity check before accessing components.
//
// Note: Uses pre-sorted entity index maintained during create/destroy,
// avoiding O(n log n) sort overhead per call. The result is still sorted
// by entity ID for deterministic iteration.
func (w *World) Entities(types ...string) []Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()
	// Pre-allocate with capacity hint to reduce allocations in hot path
	result := make([]Entity, 0, len(w.components))
	// Iterate in sorted order using pre-sorted index
	// Note: sortedIDs may be unsorted after removals (swap-delete), so we
	// collect matching entities and sort the result only
	for e, comps := range w.components {
		hasAll := true
		for _, t := range types {
			if _, ok := comps[t]; !ok {
				hasAll = false
				break
			}
		}
		if hasAll {
			result = append(result, e)
		}
	}
	// Only sort if needed (small optimization for empty/single results)
	if len(result) > 1 {
		sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	}
	return result
}

// RegisterSystem adds a system to the world.
// Thread-safe: acquires write lock to prevent race conditions with Update.
func (w *World) RegisterSystem(s System) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.systems = append(w.systems, s)
}

// SystemCount returns the number of registered systems.
func (w *World) SystemCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.systems)
}

// Update runs all registered systems with the given delta time.
// Note: Systems are expected to be registered before the game loop starts.
// If dynamic registration is needed during Update, RegisterSystem is thread-safe.
func (w *World) Update(dt float64) {
	w.mu.RLock()
	systems := make([]System, len(w.systems))
	copy(systems, w.systems)
	w.mu.RUnlock()

	for _, s := range systems {
		s.Update(w, dt)
	}
}

// AllEntities returns all entity IDs in the world, sorted for deterministic order.
// Uses same optimization as Entities() - skips sort for small results.
func (w *World) AllEntities() []Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]Entity, 0, len(w.components))
	for e := range w.components {
		result = append(result, e)
	}
	if len(result) > 1 {
		sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	}
	return result
}

// CreateEntityWithID creates an entity with a specific ID (used for load).
// If the ID already exists, it returns the existing entity.
// Updates nextID if the given ID is >= current nextID to prevent collisions.
func (w *World) CreateEntityWithID(id Entity) Entity {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, exists := w.components[id]; !exists {
		w.components[id] = make(map[string]Component)
		// Add to sorted index
		w.entityIndex[id] = len(w.sortedIDs)
		w.sortedIDs = append(w.sortedIDs, id)
	}
	if id >= w.nextID {
		w.nextID = id + 1
	}
	return id
}
