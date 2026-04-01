//go:build noebiten

// Package main provides stub types for noebiten builds.
package main

import (
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/rendering/sprite"
)

// NPCRenderer is a stub for noebiten builds.
type NPCRenderer struct {
	generator *sprite.Generator
	cache     *sprite.SpriteCache
	genre     string
	seed      int64
}

// NewNPCRenderer creates a new NPC renderer stub for noebiten builds.
func NewNPCRenderer(genre string, seed int64) *NPCRenderer {
	return &NPCRenderer{
		generator: sprite.NewGenerator(genre, seed),
		cache:     sprite.NewSpriteCache(256, 20*1024*1024),
		genre:     genre,
		seed:      seed,
	}
}

// BuildSpriteEntities returns nil for noebiten builds.
func (r *NPCRenderer) BuildSpriteEntities(world *ecs.World, playerEntity ecs.Entity) []interface{} {
	return nil
}

// UpdateAnimations is a no-op for noebiten builds.
func (r *NPCRenderer) UpdateAnimations(world *ecs.World, dt float64) {}

// SyncAnimationWithSchedule is a no-op for noebiten builds.
func (r *NPCRenderer) SyncAnimationWithSchedule(world *ecs.World) {}

// CacheStats returns cache statistics.
func (r *NPCRenderer) CacheStats() (hits, misses, evicts int64) {
	return r.cache.Stats()
}

// CacheSize returns the current cache size.
func (r *NPCRenderer) CacheSize() int {
	return r.cache.Size()
}

// CacheMemory returns the current cache memory usage.
func (r *NPCRenderer) CacheMemory() int64 {
	return r.cache.MemoryUsage()
}

// mockAppearance is a minimal appearance struct for testing.
type mockAppearance struct {
	animState string
}

// getAnimationFrameCount returns frame counts for animations.
func (r *NPCRenderer) getAnimationFrameCount(app *mockAppearance) int {
	switch app.animState {
	case sprite.AnimIdle:
		return 4
	case sprite.AnimWalk:
		return 8
	case sprite.AnimRun:
		return 8
	case sprite.AnimAttack:
		return 6
	case sprite.AnimCast:
		return 8
	case sprite.AnimSneak:
		return 8
	case sprite.AnimDead:
		return 1
	case sprite.AnimSit:
		return 1
	case sprite.AnimWork:
		return 4
	default:
		return 4
	}
}

// activityToAnimState maps schedule activities to sprite animation states.
func activityToAnimState(activity string) string {
	switch activity {
	case "sleep", "rest":
		return sprite.AnimDead
	case "walk", "travel", "patrol":
		return sprite.AnimWalk
	case "run", "flee", "chase":
		return sprite.AnimRun
	case "work", "craft", "trade":
		return sprite.AnimWork
	case "eat", "drink":
		return sprite.AnimSit
	case "socialize", "talk":
		return sprite.AnimIdle
	case "fight", "attack":
		return sprite.AnimAttack
	case "sneak", "hide":
		return sprite.AnimSneak
	case "cast", "magic":
		return sprite.AnimCast
	default:
		return sprite.AnimIdle
	}
}
