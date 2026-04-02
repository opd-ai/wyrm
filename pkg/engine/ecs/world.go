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
	mu         sync.RWMutex
	nextID     Entity
	components map[Entity]map[string]Component
	systems    []System
}

// NewWorld creates a new empty ECS world.
func NewWorld() *World {
	return &World{
		nextID:     1,
		components: make(map[Entity]map[string]Component),
	}
}

// CreateEntity allocates a new entity and returns its ID.
func (w *World) CreateEntity() Entity {
	w.mu.Lock()
	defer w.mu.Unlock()
	id := w.nextID
	w.nextID++
	w.components[id] = make(map[string]Component)
	return id
}

// DestroyEntity removes an entity and all its components.
func (w *World) DestroyEntity(e Entity) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.components, e)
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
func (w *World) Entities(types ...string) []Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()
	var result []Entity
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
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

// RegisterSystem adds a system to the world.
// Thread-safe: acquires write lock to prevent race conditions with Update.
func (w *World) RegisterSystem(s System) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.systems = append(w.systems, s)
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
func (w *World) AllEntities() []Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]Entity, 0, len(w.components))
	for e := range w.components {
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
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
	}
	if id >= w.nextID {
		w.nextID = id + 1
	}
	return id
}
