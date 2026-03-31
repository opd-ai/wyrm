// Package sprite provides billboard-based sprite rendering for entities in the
// first-person raycasting view. All sprites are procedurally generated at runtime
// from entity Appearance components — no external image files.
//
// The package provides:
//   - Sprite and SpriteSheet types for holding generated pixel data
//   - SpriteCache with LRU eviction for efficient sprite reuse
//   - Genre-specific sprite generation for humanoids, creatures, vehicles, and objects
//   - Billboard rendering integration with the raycaster z-buffer
//
// # Build Tags
//
// Tests in this package do not require the noebiten build tag since they focus
// on sprite generation logic, not Ebitengine drawing. However, integration tests
// with the raycaster may require: go test -tags=noebiten ./pkg/rendering/sprite/...
//
// # Usage
//
// Create a sprite cache:
//
//	cache := sprite.NewSpriteCache(256, 20*1024*1024) // 256 sheets, 20MB max
//
// Generate a sprite for an entity:
//
//	key := sprite.SpriteCacheKey{
//	    Category: appearance.SpriteCategory,
//	    BodyPlan: appearance.BodyPlan,
//	    GenreID:  appearance.GenreID,
//	    // ... other fields
//	}
//	sheet := cache.GetOrGenerate(key, seed)
//	frame := sheet.GetFrame(appearance.AnimState, appearance.AnimFrame)
//
// Render sprite using the SpriteRenderer:
//
//	renderer.DrawSprite(screen, sprite, screenX, screenY, zBuffer)
package sprite
