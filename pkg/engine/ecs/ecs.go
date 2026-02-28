// Package ecs provides the core Entity-Component-System framework.
package ecs

import (
	"errors"
	"sort"
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
	id := w.nextID
	w.nextID++
	w.components[id] = make(map[string]Component)
	return id
}

// DestroyEntity removes an entity and all its components.
func (w *World) DestroyEntity(e Entity) {
	delete(w.components, e)
}

// AddComponent attaches a component to an entity.
// Returns ErrEntityNotFound if the entity does not exist.
func (w *World) AddComponent(e Entity, c Component) error {
	if _, ok := w.components[e]; !ok {
		return ErrEntityNotFound
	}
	w.components[e][c.Type()] = c
	return nil
}

// GetComponent retrieves a component by type from an entity.
func (w *World) GetComponent(e Entity, typeName string) (Component, bool) {
	comps, ok := w.components[e]
	if !ok {
		return nil, false
	}
	c, ok := comps[typeName]
	return c, ok
}

// Entities returns all entities that have the given component types,
// sorted by entity ID for deterministic iteration order.
func (w *World) Entities(types ...string) []Entity {
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
func (w *World) RegisterSystem(s System) {
	w.systems = append(w.systems, s)
}

// Update runs all registered systems with the given delta time.
func (w *World) Update(dt float64) {
	for _, s := range w.systems {
		s.Update(w, dt)
	}
}
