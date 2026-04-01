//go:build !noebiten

// Package main provides the NPC sprite rendering integration for the Wyrm client.
package main

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/rendering/raycast"
	"github.com/opd-ai/wyrm/pkg/rendering/sprite"
)

// NPCRenderer handles the rendering of NPC entities as billboard sprites.
type NPCRenderer struct {
	generator *sprite.Generator
	cache     *sprite.SpriteCache
	genre     string
	seed      int64
}

// NewNPCRenderer creates a new NPC renderer with the given genre and seed.
func NewNPCRenderer(genre string, seed int64) *NPCRenderer {
	return &NPCRenderer{
		generator: sprite.NewGenerator(genre, seed),
		cache:     sprite.NewSpriteCache(256, 20*1024*1024), // 256 sheets, 20MB max
		genre:     genre,
		seed:      seed,
	}
}

// BuildSpriteEntities creates SpriteEntity objects for all NPCs with Position
// and Appearance components in the given world.
func (r *NPCRenderer) BuildSpriteEntities(world *ecs.World, playerEntity ecs.Entity) []*raycast.SpriteEntity {
	entities := world.Entities("Position", "Appearance")
	spriteEntities := make([]*raycast.SpriteEntity, 0, len(entities))

	for _, e := range entities {
		// Skip the player entity
		if e == playerEntity {
			continue
		}

		posComp, ok := world.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)

		appComp, ok := world.GetComponent(e, "Appearance")
		if !ok {
			continue
		}
		app := appComp.(*components.Appearance)

		// Skip invisible entities
		if !app.Visible {
			continue
		}

		// Get or generate the sprite sheet
		cacheKey := r.buildCacheKey(app, e)
		sheet := r.cache.GetOrGenerate(cacheKey, r.generator.GenerateSheet)
		if sheet == nil {
			continue
		}

		// Create the sprite entity for the renderer
		spriteEntity := &raycast.SpriteEntity{
			X:         pos.X,
			Y:         pos.Y,
			Sheet:     sheet,
			AnimState: app.AnimState,
			AnimFrame: app.AnimFrame,
			Scale:     app.Scale,
			Opacity:   app.Opacity,
			FlipH:     app.FlipH,
			Visible:   app.Visible,
		}

		spriteEntities = append(spriteEntities, spriteEntity)
	}

	return spriteEntities
}

// buildCacheKey creates a sprite cache key from an Appearance component.
func (r *NPCRenderer) buildCacheKey(app *components.Appearance, entityID ecs.Entity) sprite.SpriteCacheKey {
	return sprite.SpriteCacheKey{
		Category:       app.SpriteCategory,
		BodyPlan:       app.BodyPlan,
		GenreID:        app.GenreID,
		PrimaryColor:   app.PrimaryColor,
		SecondaryColor: app.SecondaryColor,
		AccentColor:    app.AccentColor,
		Scale:          app.Scale,
		Seed:           int64(entityID), // Use entity ID as seed for variation
	}
}

// UpdateAnimations updates animation timers for all Appearance components.
// Should be called each frame with the frame's delta time.
func (r *NPCRenderer) UpdateAnimations(world *ecs.World, dt float64) {
	entities := world.Entities("Appearance")

	for _, e := range entities {
		appComp, ok := world.GetComponent(e, "Appearance")
		if !ok {
			continue
		}
		app := appComp.(*components.Appearance)

		// Update animation timer
		app.AnimTimer += dt

		// Advance frame when timer exceeds frame duration
		frameDuration := 1.0 / sprite.AnimFrameRate
		if app.AnimTimer >= frameDuration {
			app.AnimTimer -= frameDuration
			app.AnimFrame++

			// Get frame count for current animation
			frameCount := r.getAnimationFrameCount(app)
			if app.AnimFrame >= frameCount {
				app.AnimFrame = 0 // Loop animation
			}
		}
	}
}

// getAnimationFrameCount returns the frame count for the current animation state.
func (r *NPCRenderer) getAnimationFrameCount(app *components.Appearance) int {
	// Default frame counts per animation state
	switch app.AnimState {
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

// SyncAnimationWithSchedule updates NPC animation state based on their schedule.
// Maps schedule activities to appropriate animation states.
func (r *NPCRenderer) SyncAnimationWithSchedule(world *ecs.World) {
	entities := world.Entities("Appearance", "Schedule")

	for _, e := range entities {
		appComp, ok := world.GetComponent(e, "Appearance")
		if !ok {
			continue
		}
		app := appComp.(*components.Appearance)

		schedComp, ok := world.GetComponent(e, "Schedule")
		if !ok {
			continue
		}
		sched := schedComp.(*components.Schedule)

		// Map schedule activity to animation state
		newAnimState := activityToAnimState(sched.CurrentActivity)

		// Only change if different (to preserve animation frame on same state)
		if app.AnimState != newAnimState {
			app.AnimState = newAnimState
			app.AnimFrame = 0
			app.AnimTimer = 0.0
		}
	}
}

// activityToAnimState maps schedule activities to sprite animation states.
func activityToAnimState(activity string) string {
	switch activity {
	case "sleep", "rest":
		return sprite.AnimDead // Lying down (reusing dead pose)
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

// CacheStats returns the sprite cache statistics for debugging.
func (r *NPCRenderer) CacheStats() (hits, misses, evicts int64) {
	return r.cache.Stats()
}

// CacheSize returns the current number of cached sprite sheets.
func (r *NPCRenderer) CacheSize() int {
	return r.cache.Size()
}

// CacheMemory returns the current memory usage of the sprite cache in bytes.
func (r *NPCRenderer) CacheMemory() int64 {
	return r.cache.MemoryUsage()
}
