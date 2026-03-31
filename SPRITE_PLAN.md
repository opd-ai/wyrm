# SPRITE_PLAN.md — Wyrm Entity Rendering System Design

**Status**: ✅ Implemented (FEATURES.md §18 — Sprite rendering)
**Phase**: Completed (Entity Rendering)
**Last Updated**: 2026-03-31

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current Rendering Pipeline](#2-current-rendering-pipeline)
3. [ECS Architecture for Sprites](#3-ecs-architecture-for-sprites)
4. [New Component: Appearance](#4-new-component-appearance)
5. [Sprite Generation System](#5-sprite-generation-system)
6. [Billboard Rendering Integration](#6-billboard-rendering-integration)
7. [Genre Visual Specifications](#7-genre-visual-specifications)
8. [Animation System](#8-animation-system)
9. [Performance Optimization](#9-performance-optimization)
10. [Multiplayer Considerations](#10-multiplayer-considerations)
11. [Implementation Roadmap](#11-implementation-roadmap)
12. [Risk Assessment](#12-risk-assessment)
13. [Testing Strategy](#13-testing-strategy)
14. [Dependencies](#14-dependencies)
15. [Appendix: Entity Type Catalog](#15-appendix-entity-type-catalog)

---

## 1. Executive Summary

Wyrm's first-person raycasting renderer currently supports textured walls, floors, and ceilings
with genre-specific color palettes and 11 post-processing effects. It does **not** render entities
(NPCs, creatures, vehicles, objects) in the 3D view. This document specifies the complete
architecture for adding billboard-based sprite rendering to the existing pipeline.

### Current State (from FEATURES.md §18)

| Feature | Status |
|---------|--------|
| First-person raycaster (DDA) | ✅ `pkg/rendering/raycast/core.go` (386 LOC) |
| Procedural texture generation | ✅ `pkg/rendering/texture/generator.go` (132 LOC) |
| Genre-specific color palettes | ✅ 5 genres × 4 colors each |
| Wall/floor/ceiling rendering | ✅ `pkg/rendering/raycast/draw.go` (209 LOC) |
| Post-processing effects | ✅ 11 effect types, `pkg/rendering/postprocess/effects.go` (523 LOC) |
| **Sprite rendering** | ❌ **Not implemented** |
| Particle effects | ❌ Not implemented |
| Lighting system | ❌ Not implemented |
| Fog effects | ✅ Distance-based fog in raycaster |
| Skybox rendering | ❌ Not implemented |

### Goal

Add billboard-based sprite rendering for all entity types (NPCs, creatures, vehicles,
interactive objects) while maintaining 60 FPS at 1280×720, zero external assets, genre-specific
visuals, ECS composition, server-authoritative state, and deterministic procedural appearance.

---

## 2. Current Rendering Pipeline

### 2.1 Raycaster Core (`pkg/rendering/raycast/core.go`)

The renderer uses Digital Differential Analysis (DDA) for wall intersection detection.

**Renderer struct:**
```go
type Renderer struct {
    Width, Height       int
    PlayerX, PlayerY    float64     // World position
    PlayerA             float64     // Heading angle (radians)
    WorldMap            [][]int     // 2D wall grid (0=empty, 1-3=wall type)
    FOV                 float64     // Field of view (default π/3 = 60°)
    Genre               string
    WallTextures        []*texture.Texture
    FloorTexture        *texture.Texture
    CeilTexture         *texture.Texture
    textureSeed         int64
}
```

**Key constants:**

| Constant | Value | Purpose |
|----------|-------|---------|
| `DefaultFOV` | π/3 (60°) | Field of view |
| `MaxRaySteps` | 64 | DDA max iterations |
| `MaxRayDistance` | 100.0 | Far plane distance |
| `FogDistance` | 16.0 | Fog falloff radius |
| `MinFogFactor` | 0.2 | Minimum brightness in fog |
| `MinWallDistance` | 0.1 | Prevents division by zero |
| `TextureSize` | 64 | Procedural texture resolution |

**Ray casting pipeline:**
1. `calculateDeltaDist(rayDirX, rayDirY)` — step lengths for DDA grid traversal
2. `calculateSideDist(...)` — initial distances to first axis-aligned grid line
3. `performDDA(...)` — march through grid cells, return hit wall + side + distance
4. `castRayWithTexCoord(...)` — full ray cast with texture U coordinate
5. `GetWallTextureColor(wallType, texX, texY, distance)` — sample texture + fog

**Fog formula:** `fogFactor = clamp(1.0 - distance/FogDistance, MinFogFactor, 1.0)`
**Side darkening:** Y-axis wall hits are darkened to 0.8× brightness.

### 2.2 Draw Pipeline (`pkg/rendering/raycast/draw.go`)

```
Draw(screen)
├── drawFloorCeiling(screen)    ← per-scanline raycasted textured floor/ceiling
└── drawWalls(screen)           ← per-column DDA wall strips with texture + fog
```

**Wall rendering per column:**
1. Calculate ray direction from FOV and column index
2. Cast ray → get perpendicular distance, wall type, texture X coordinate
3. Apply fisheye correction: `distance *= cos(cameraX × FOV/2)`
4. Calculate wall height: `wallHeight = screenHeight / distance`
5. Sample wall texture at computed `(texX, texY)` coordinates
6. Apply side darkening and fog
7. Draw vertical strip from `drawStart` to `drawEnd`

**Critical for sprites:** The per-column wall distance values (the "z-buffer") are needed to
correctly occlude sprites behind walls. Currently these values are local to `drawWalls()` and
must be exposed.

### 2.3 Texture Generation (`pkg/rendering/texture/generator.go`)

Procedural textures are generated using noise-based palette sampling with ±10 RGB variation.

**Genre palettes (5 genres × 4 colors):**

| Genre | Color 1 | Color 2 | Color 3 | Color 4 | Character |
|-------|---------|---------|---------|---------|-----------|
| Fantasy | Gold `D4A574` | Green `4A7C23` | Brown `8B4513` | Light Gold `C0A060` | Warm, natural, earthy |
| Sci-Fi | Blue `1E90FF` | White `F0F0F0` | Chrome `C0C0C0` | Steel Blue `406090` | Metallic, cold, industrial |
| Horror | Grey-Green `556B2F` | Near Black `1A1A1A` | Blood Red `8B0000` | Dark Grey `3F3F3F` | Decay, darkness, blood |
| Cyberpunk | Neon Pink `FF00FF` | Cyan `00FFFF` | Dark Grey `2F2F2F` | Purple `800080` | Neon, synthetic, bright |
| Post-Apoc | Sepia `704214` | Orange `CC7722` | Rust `B7410E` | Tan `8B6333` | Oxidized, sandy, weathered |

### 2.4 Pattern Generation (`pkg/rendering/texture/patterns.go`)

Genre-specific texture patterns control noise frequency, contrast, and saturation.

| Genre | NoiseScale | Pattern | Detail | Contrast | Saturation | Secondary |
|-------|-----------|---------|--------|----------|-----------|-----------|
| Fantasy | 0.08 | Layered | 0.6 | 1.0 | 0.9 | 0.20 |
| Sci-Fi | 0.10 | Grid | 0.8 | 1.2 | 0.7 | 0.05 |
| Horror | 0.06 | Voronoi | 0.5 | 1.5 | 0.3 | 0.15 |
| Cyberpunk | 0.12 | Grid | 0.9 | 1.4 | 1.0 | 0.08 |
| Post-Apoc | 0.07 | Distortion | 0.4 | 0.9 | 0.5 | 0.25 |

Pattern types: `PatternNoise` (0), `PatternGrid` (1), `PatternVoronoi` (2),
`PatternDistortion` (3), `PatternLayered` (4).

### 2.5 Post-Processing (`pkg/rendering/postprocess/effects.go`)

Genre-specific post-processing pipelines are applied after the 3D render, which means
sprite rendering must happen **before** post-processing to receive the same treatment.

| Genre | Effects Applied |
|-------|----------------|
| Fantasy | WarmColorGrade(0.6) |
| Sci-Fi | Scanlines(2, 0.3) → Bloom(0.6, 0.4) → CoolColorGrade(0.4) |
| Horror | Desaturate(0.7) → Vignette(0.5, 0.3) → DarkenOverall(0.2) |
| Cyberpunk | ChromaticAberration(3) → Bloom(0.5, 0.5) → NeonGlow(0.4) |
| Post-Apoc | Sepia(0.8) → FilmGrain(0.15) → Desaturate(0.3) |

11 total implemented effect types: WarmColorGrade, CoolColorGrade, Scanlines, Bloom, Desaturate,
Vignette, DarkenOverall, ChromaticAberration, NeonGlow, Sepia, FilmGrain. Future/planned
post-process effects (not yet implemented in `pkg/rendering/postprocess/effects.go`): Fog, LightRays.

### 2.6 Noise Generation (`pkg/procgen/noise/generator.go`)

Value noise with XOR-based hashing, bilinear interpolation, and smoothstep easing.

- `Noise2D(x, y, seed)` → `[0, 1]` — base noise for sprite texture generation
- `Noise2DSigned(x, y, seed)` → `[-1, 1]` — signed variant for offset/distortion
- `HashToFloat(x, y, seed)` → `[0, 1]` — deterministic per-coordinate hash
- `Smoothstep(t)` — cubic interpolation `3t² - 2t³`
- `Lerp(a, b, t)` — linear interpolation

---

## 3. ECS Architecture for Sprites

### 3.1 Current ECS (`pkg/engine/ecs/`)

- `Entity` = `uint64` ID
- `Component` = interface with `Type() string` (pure data, NO logic)
- `System` = interface with `Update(w *World, dt float64)` (ALL logic)
- `World` = entity registry + component store + system runner

### 3.2 Existing Visual-Adjacent Components

| Component | Location | Visual Relevance |
|-----------|----------|-----------------|
| `Position` | types.go:5 | X, Y, Z, Angle — entity world position and heading |
| `Health` | types.go:14 | Current/Max — affects visual state (wounded/dead) |
| `CombatState` | types.go:556 | InCombat, IsAttacking — animation triggers |
| `Stealth` | types.go:573 | Visibility (0.0–1.0), Sneaking — alpha/pose |
| `EmotionalState` | types.go:1021 | CurrentEmotion, Intensity — facial expression |
| `HazardEffect` | types.go:1131 | VisualEffect string — overlay effects |
| `Vehicle` | types.go:171 | VehicleType, Speed — which sprite to show |
| `VehicleState` | types.go:217 | IsOccupied, DamagePercent — visual damage states |

### 3.3 Entity Creation Paths

**NPCs** are created via `pkg/procgen/adapters/entity.go`:
```go
SpawnNPC(world, data, x, y, factionID) → Entity with {Position, Health, Faction, Schedule}
```

**Vehicles** are created via `pkg/procgen/adapters/vehicle.go`:
```go
SpawnVehicleEntity(world, vehicleData, x, y, z) → Entity with {Position, Vehicle}
// Note: VehiclePhysics and VehicleState are not yet attached by the current adapter.
```

**Companions** are generated via `CompanionManager` in `pkg/companion/companion.go`:
```go
// CompanionManager.CreateCompanion deterministically generates a Companion.
// Seed and genre are threaded through the manager; see pkg/companion/companion.go for details.
func (cm *CompanionManager) CreateCompanion(playerID uint64, genre string, preferredRole CombatRole) *Companion
// Returns a data-only Companion struct — not yet spawned as an ECS entity.
```

### 3.4 Current RenderSystem (`pkg/engine/systems/render.go` — 18 LOC, stub)

```go
type RenderSystem struct {
    PlayerEntity ecs.Entity
}

func (s *RenderSystem) Update(w *ecs.World, dt float64) {
    if s.PlayerEntity != 0 {
        _, _ = w.GetComponent(s.PlayerEntity, "Position")
    }
}
```

This stub must be expanded to drive sprite rendering.

---

## 4. New Component: Appearance

### 4.1 Component Definition

Add to `pkg/engine/components/types.go`:

```go
// Appearance defines the visual representation of an entity in the first-person view.
// It is pure data — rendering logic belongs in the SpriteRenderSystem.
type Appearance struct {
    // SpriteCategory selects the generation algorithm.
    // One of: "humanoid", "creature", "vehicle", "object", "effect"
    SpriteCategory string

    // BodyPlan selects the silhouette template within the category.
    // Examples: "warrior", "merchant", "wolf", "dragon", "horse", "buggy"
    BodyPlan string

    // PrimaryColor and SecondaryColor are packed RGBA values (genre-derived).
    PrimaryColor   uint32
    SecondaryColor uint32

    // AccentColor for details (belt, trim, insignia).
    AccentColor uint32

    // Scale multiplier relative to default entity height (1.0 = standard humanoid).
    // Range: 0.25 (small critter) to 4.0 (dragon/mech).
    Scale float64

    // AnimState is the current animation state identifier.
    // One of: "idle", "walk", "run", "attack", "cast", "sneak", "dead", "sit", "work"
    AnimState string

    // AnimFrame is the current frame index within the animation.
    AnimFrame int

    // AnimTimer accumulates dt for frame advancement.
    AnimTimer float64

    // Visible controls whether the entity is rendered at all.
    // Set to false for hidden/despawned entities.
    Visible bool

    // Opacity controls alpha blending (0.0 = invisible, 1.0 = opaque).
    // Driven by Stealth.Visibility when sneaking.
    Opacity float64

    // FlipH mirrors the sprite horizontally (for facing direction).
    FlipH bool

    // Decorations are additional overlay identifiers (armor, hat, weapon held).
    Decorations []string

    // DamageOverlay intensity (0.0 = pristine, 1.0 = heavily damaged).
    // Driven by Health.Current/Health.Max or VehicleState.DamagePercent.
    DamageOverlay float64

    // GenreID is stored for sprite generation cache keying.
    GenreID string
}

func (a *Appearance) Type() string { return "Appearance" }
```

### 4.2 Integration Points

When spawning entities, the `Appearance` component must be added alongside existing
components. Modify the adapters:

**Entity adapter** (`pkg/procgen/adapters/entity.go`):
```go
func SpawnNPC(world *ecs.World, data *NPCData, x, y float64, factionID string, genre string) (ecs.Entity, error) {
    e := world.CreateEntity()
    if err := world.AddComponent(e, &components.Position{X: x, Y: y}); err != nil {
        return 0, fmt.Errorf("failed to add Position: %w", err)
    }
    if err := world.AddComponent(e, &components.Health{Current: data.Health, Max: data.Health}); err != nil {
        return 0, fmt.Errorf("failed to add Health: %w", err)
    }
    if err := world.AddComponent(e, &components.Faction{ID: factionID}); err != nil {
        return 0, fmt.Errorf("failed to add Faction: %w", err)
    }
    if err := world.AddComponent(e, &components.Schedule{}); err != nil {
        return 0, fmt.Errorf("failed to add Schedule: %w", err)
    }
    // NEW: Add Appearance based on NPC tags and genre
    if err := world.AddComponent(e, generateNPCAppearance(data, genre)); err != nil {
        return 0, fmt.Errorf("failed to add Appearance: %w", err)
    }
    return e, nil
}
```

**Vehicle adapter** (`pkg/procgen/adapters/vehicle.go`):
```go
func SpawnVehicleEntity(world *ecs.World, v *VehicleData, x, y, z float64) ecs.Entity {
    e := world.CreateEntity()
    // ... existing components ...
    // NEW: Add Appearance from VehicleData colors and type
    world.AddComponent(e, &components.Appearance{
        SpriteCategory: "vehicle",
        BodyPlan:       v.VehicleType,
        PrimaryColor:   v.Color,
        SecondaryColor: v.SecondaryColor,
        Scale:          1.0,
        AnimState:      "idle",
        Visible:        true,
        Opacity:        1.0,
        GenreID:        v.GenreID,
    })
    return e
}
```

---

## 5. Sprite Generation System

### 5.1 Architecture

Sprites are procedurally generated pixel arrays, cached by a composite key of
`(SpriteCategory, BodyPlan, GenreID, PrimaryColor, SecondaryColor, QuantizedScale)`.
`AnimState` selects which animation sequence/sheet to use for a given cached sprite,
and `AnimFrame` is an index into that sequence — neither is part of the cache key itself.

```
pkg/rendering/sprite/
├── sprite.go        # Sprite struct, cache, lookup
├── generator.go     # Top-level generation dispatch
├── humanoid.go      # Humanoid silhouette algorithm
├── creature.go      # Creature body plan algorithms
├── vehicle.go       # Vehicle silhouette algorithms
├── object.go        # Static object sprites (crates, signs, doors)
└── animation.go     # Frame sequence definitions
```

### 5.2 Sprite Struct

```go
// Sprite is a single rendered frame of an entity's visual.
type Sprite struct {
    Width, Height int
    Pixels        []color.RGBA // Row-major, Width × Height
}

// SpriteSheet is a collection of animation frames for one entity configuration.
type SpriteSheet struct {
    Frames    []Sprite       // Indexed by frame number
    FrameRate float64        // Frames per second
    Looping   bool           // Whether animation loops
}
```

### 5.3 Humanoid Generation Algorithm

Humanoid sprites use a **template silhouette** approach: a small pixel grid (e.g., 32×64) is
filled with body regions, then colored using the entity's palette.

**Body regions:**

| Region | Approximate Location | Purpose |
|--------|---------------------|---------|
| Head | Top 8 rows, centered | Face/helmet area |
| Torso | Rows 8–24 | Main body |
| Arms | Columns 0–8 and 24–32, rows 8–24 | Weapon holding, gestures |
| Legs | Rows 24–48 | Walking animation |
| Feet | Bottom 4 rows | Ground contact |
| Equipment overlay | Full sprite | Armor, robes, weapons |

**Color mapping:**
- Skin/face → derive from `PrimaryColor` (lighter variant)
- Clothing/armor → `PrimaryColor`
- Trim/accents → `SecondaryColor`
- Equipment → `AccentColor`
- Hair → derived from genre palette index 2

**Occupation-based BodyPlan mapping (fantasy genre example):**

| Occupation | BodyPlan | Visual Cues |
|------------|----------|-------------|
| merchant | `merchant` | Wide torso (robes), hands at waist |
| guard | `guard` | Tall, spear/shield silhouette |
| blacksmith | `smith` | Broad shoulders, hammer in hand |
| innkeeper | `innkeeper` | Apron, hands spread |
| healer | `healer` | Staff, flowing robes |
| farmer | `farmer` | Hoe/tool, wide hat |
| bard | `bard` | Instrument outline, light build |
| priest | `priest` | Tall hat, vestments |

### 5.4 Creature Generation Algorithm

Creatures use a **body plan matrix** approach: a template defines the proportional layout of
body segments (head, body, limbs, tail, wings), which are then scaled, colored, and
distorted per-seed.

**Body plan templates:**

| BodyPlan | Segments | Example Entities |
|----------|----------|-----------------|
| `quadruped` | Head, body, 4 legs, tail | Wolf, horse, dog, rat |
| `biped_large` | Head, torso, 2 arms, 2 legs | Ogre, troll, golem |
| `serpentine` | Head, elongated body, no legs | Snake, wyrm, tentacle |
| `avian` | Head, body, 2 wings, 2 legs | Bird, harpy, griffin |
| `insectoid` | Head, thorax, abdomen, 6 legs | Giant spider, beetle |
| `amorphous` | Central mass, pseudopods | Slime, ooze, ghost |
| `dragon` | Head, body, 4 legs, 2 wings, tail | Dragon, wyvern |

**Generation steps:**
1. Select body plan template from `BodyPlan` field
2. Derive sub-seed per segment: `segmentSeed = mixSeeds(entitySeed, segmentName)`
3. For each segment: generate silhouette pixels using noise-distorted ellipses
4. Apply genre palette colors (body → primary, details → secondary, eyes → accent)
5. Apply noise-based surface detail (scales, fur, chitin patterns) using genre pattern config

### 5.5 Vehicle Sprite Generation

Vehicle sprites are derived from `VehicleData` properties. The silhouette is selected by
`VehicleType` and colored with `VehicleData.Color`/`SecondaryColor`.

**Vehicle archetypes per genre (from `GenreVehicleArchetypes`):**

| Genre | Vehicle 1 | Vehicle 2 | Vehicle 3 |
|-------|-----------|-----------|-----------|
| Fantasy | Horse | Horse Cart | Sailing Ship |
| Sci-Fi | Hover-Bike | Shuttle | Mech Walker |
| Horror | Hearse | Bone Cart | Swamp Raft |
| Cyberpunk | Street Bike | APC | Personal Drone |
| Post-Apoc | Wasteland Buggy | Armored Truck | Gyroplane |

**Damage states:** The `DamageOverlay` field (from `VehicleState.DamagePercent`) adds
progressive visual degradation: scratches at 0.25, dents at 0.50, missing panels at 0.75,
fire/smoke at 1.0.

### 5.6 Object Sprite Generation

Static world objects use simple template-based sprites:

| Object Type | Examples | Generation |
|-------------|----------|-----------|
| Container | Chest, crate, barrel, locker | Rectangle + genre color + lid/lock detail |
| Furniture | Chair, table, bed, workbench | Silhouette template + genre material colors |
| Sign | Road sign, shop sign, notice | Rectangle + text placeholder area |
| Resource | Tree, ore vein, herb patch | Organic noise silhouette + biome color |
| Door | Building entrance, dungeon gate | Rectangle with handle detail |

### 5.7 Sprite Cache

```go
type SpriteCache struct {
    mu      sync.RWMutex
    cache   map[SpriteCacheKey]*SpriteSheet
    lru     *list.List          // For LRU eviction
    maxSize int                 // Max cached sprite sheets (default 256)
    maxMem  int64               // Max memory in bytes (default 20MB)
    curMem  int64               // Current memory usage
}

type SpriteCacheKey struct {
    Category       string
    BodyPlan       string
    Genre          string
    Primary        uint32
    Secondary      uint32
    QuantizedScale uint32 // fixed-point scale (e.g., scale*1000, rounded) to avoid float map-key issues
}
```

Cache eviction uses LRU. Memory is estimated as `width × height × 4 × frameCount` per sheet.

---

## 6. Billboard Rendering Integration

### 6.1 Modified Draw Pipeline

The draw pipeline must be extended to render sprites after `drawWalls()` (and before any
post-processing), using the wall-populated depth buffer (z-buffer) to correctly occlude
sprites behind wall columns.

**New pipeline:**

```
Draw(screen)
├── drawFloorCeiling(screen)          ← existing (unchanged)
├── drawWalls(screen, zBuffer)        ← modified: populate z-buffer per column
└── drawSprites(screen, zBuffer)      ← NEW: billboard sprites with depth test (after walls)
```

### 6.2 Z-Buffer

The z-buffer is a `[]float64` of length `screenWidth`, storing the perpendicular distance
to the closest wall for each screen column. This is populated during `drawWalls()` and
consumed during `drawSprites()` to skip sprite pixels that are behind walls.

```go
// In Renderer struct:
type Renderer struct {
    // ... existing fields ...
    ZBuffer []float64  // Per-column wall distance, populated each frame
}
```

### 6.3 Billboard Transform (World → Screen)

For each visible entity, compute screen position and size:

```
Given:
    entityPos   = (ex, ey)       — entity world position
    playerPos   = (px, py)       — camera world position
    playerAngle = pa             — camera heading (radians)
    FOV         = fov            — field of view (radians)
    screenW, screenH             — screen dimensions

Step 1: Translate to camera-relative coordinates
    dx = ex - px
    dy = ey - py

Step 2: Rotate into camera space
    invDet = 1.0 / (planeX * dirY - dirX * planeY)
    transformX = invDet * (dirY * dx - dirX * dy)    // lateral offset
    transformY = invDet * (-planeY * dx + planeX * dy) // depth

    where:
        dirX = cos(pa), dirY = sin(pa)                // camera direction
        planeX = -sin(pa) * tan(fov/2)                // camera plane X
        planeY = cos(pa) * tan(fov/2)                 // camera plane Y

Step 3: Skip if behind camera
    if transformY <= 0: skip

Step 4: Calculate screen position
    spriteScreenX = (screenW / 2) * (1 + transformX / transformY)

Step 5: Calculate sprite screen height (perspective projection)
    spriteHeight = abs(screenH / transformY) * appearance.Scale
    spriteWidth  = spriteHeight * (sprite.Width / sprite.Height)  // maintain aspect ratio

Step 6: Calculate vertical position
    drawStartY = (screenH / 2) - (spriteHeight / 2)   // centered vertically
    drawEndY   = (screenH / 2) + (spriteHeight / 2)

Step 7: Calculate horizontal bounds
    drawStartX = spriteScreenX - (spriteWidth / 2)
    drawEndX   = spriteScreenX + (spriteWidth / 2)
```

### 6.4 Depth-Tested Sprite Drawing

For each screen column within the sprite's horizontal bounds:

```
for stripe = drawStartX to drawEndX:
    if stripe < 0 or stripe >= screenW: continue
    if transformY >= zBuffer[stripe]: continue  // sprite column is behind wall

    texX = (stripe - drawStartX) * sprite.Width / spriteWidth  // texture U

    for y = drawStartY to drawEndY:
        if y < 0 or y >= screenH: continue

        texY = (y - drawStartY) * sprite.Height / spriteHeight  // texture V
        pixel = sprite.Pixels[texY * sprite.Width + texX]

        if pixel.A == 0: continue  // transparent pixel

        // Apply fog (same formula as walls)
        fogFactor = clamp(1.0 - transformY / FogDistance, MinFogFactor, 1.0)
        pixel.R = uint8(float64(pixel.R) * fogFactor)
        pixel.G = uint8(float64(pixel.G) * fogFactor)
        pixel.B = uint8(float64(pixel.B) * fogFactor)

        // Apply opacity (from Stealth.Visibility)
        pixel.A = uint8(float64(pixel.A) * appearance.Opacity)

        screen.Set(stripe, y, pixel)
```

### 6.5 Sprite Sorting

Sprites must be drawn back-to-front (painter's algorithm) so that nearer sprites overwrite
farther ones:

```go
sort.Slice(visibleEntities, func(i, j int) bool {
    return visibleEntities[i].Distance > visibleEntities[j].Distance
})
```

Distance is computed as `transformY` (perpendicular distance in camera space) to match the
wall distance convention used by the z-buffer.

### 6.6 Facing Direction

Sprites always face the camera (billboard behavior). The `FlipH` field determines which
direction the entity appears to face relative to the player. When the entity's `Position.Angle`
points away from the player, `FlipH = false`; when facing toward, `FlipH = true`.

```go
angleDiff := normalizeAngle(entity.Position.Angle - angleToPlayer)
appearance.FlipH = math.Abs(angleDiff) < math.Pi/2
```

---

## 7. Genre Visual Specifications

### 7.1 NPC Occupation Sprites by Genre

Each genre has unique occupations (from `GetGenreOccupations` in `npc_occupation.go`) that
require distinct visual representations.

**Fantasy occupations:**
merchant, guard, healer, blacksmith, innkeeper, farmer, scribe, bard, priest, miner

- Color palette: warm golds, greens, browns
- Clothing style: medieval robes, leather armor, tunics
- Equipment: swords, staves, aprons, instruments, pickaxes

**Sci-Fi occupations:**
merchant, guard, healer, technician, scientist, pilot, engineer, medic

- Color palette: cool blues, whites, chromes
- Clothing style: jumpsuits, lab coats, flight suits, powered armor
- Equipment: datapads, plasma tools, medical scanners, jet packs

**Horror occupations:**
merchant, guard, healer, priest, mortician, hunter, herbalist, gravedigger

- Color palette: desaturated grey-greens, near-blacks, blood reds
- Clothing style: tattered, stained, patched, hooded
- Equipment: lanterns, crosses, shovels, bone implements

**Cyberpunk occupations:**
merchant, guard, medic, hacker, fixer, bodyguard, street vendor, tech dealer

- Color palette: neon pink, cyan, dark greys, purple
- Clothing style: leather jackets, visors, chrome implants, holographic
- Equipment: cyberjacks, neon weapons, holoprojectors

**Post-Apocalyptic occupations:**
merchant, guard, healer, scavenger, mechanic, farmer, hunter, water merchant

- Color palette: sepia, orange, rust, tan
- Clothing style: patched, makeshift armor, gas masks, goggles
- Equipment: jury-rigged tools, scrap weapons, water containers

### 7.2 Weather Visual Effects on Sprites

Weather conditions (from `pkg/engine/systems/weather.go`) affect sprite rendering:

| Weather | Visibility Mod | Sprite Effect |
|---------|---------------|--------------|
| clear | 1.0 | No modification |
| rain | 0.7 | Blue tint overlay, slight vertical streaks |
| fog / mist | 0.3 | Increased fog falloff, reduced contrast |
| thunderstorm | 0.4 | Flash-lit intermittently, heavy rain overlay |
| dust | 0.5 | Sepia tint, particle overlay |
| dust_storm | 0.2 | Heavy sepia tint, dense particle overlay |
| blood_moon | 0.5 | Red tint, enhanced shadow contrast |
| smog | 0.6 | Yellow-grey tint, reduced saturation |
| acid_rain | 0.7 | Green tint, damage particle overlay |
| ion_storm | 0.6 | Blue-white flicker, scanline distortion |
| radiation_burst | 0.8 | Green glow overlay, distortion |
| ash_fall | 0.5 | Grey tint, falling particle overlay |
| neon_haze | 0.75 | Bloom around bright sprite colors |

### 7.3 Genre Post-Processing Interaction

Sprites are rendered into the same framebuffer as walls/floors before post-processing runs.
This means:

- **Fantasy**: sprites receive warm gold color grading
- **Sci-Fi**: sprites have scanlines and bloom applied
- **Horror**: sprites are desaturated and vignetted (enhances dread)
- **Cyberpunk**: sprites receive chromatic aberration and neon glow
- **Post-Apoc**: sprites are sepia-toned with film grain

No special handling needed — the post-processing pipeline applies uniformly.

### 7.4 Companion Visual Distinction

Companions (from `pkg/companion/`) have 5 personalities × 5 combat roles that affect their
visual presentation:

**Personalities:** Brave, Cautious, Loyal, Aggressive, Wise

- Brave: upright posture, forward-leaning idle
- Cautious: hunched slightly, weapon-ready idle
- Loyal: positioned close to player, matching player facing
- Aggressive: wide stance, weapon-drawn idle
- Wise: tall posture, staff/book held

**Combat Roles:** Tank, DPS, Healer, Support, Ranged

- Tank: heavy armor, shield, wide silhouette
- DPS: medium armor, dual weapons or two-hander
- Healer: light robes, staff with glow accent
- Support: medium gear, utility belt/pouches
- Ranged: light armor, bow/rifle, quiver/ammo

---

## 8. Animation System

### 8.1 Animation State Machine

```
    idle ←──────────── return to idle after timeout
     │
     ├── walk ──→ run (speed > threshold)
     │     ↑        │
     │     └────────┘ (speed < threshold)
     │
     ├── attack ──→ idle (after attack duration)
     │
     ├── cast ──→ idle (after cast duration)
     │
     ├── sneak (while Stealth.Sneaking == true)
     │
     ├── sit / work (occupation-driven)
     │
     └── dead (Health.Current <= 0, terminal state)
```

### 8.2 Frame Sequences

| AnimState | Frame Count | Frame Rate (FPS) | Looping | Notes |
|-----------|-------------|-------------------|---------|-------|
| idle | 2 | 1 | Yes | Subtle breathing/sway |
| walk | 4 | 5 | Yes | Leg alternation cycle |
| run | 6 | 10 | Yes | Extended stride |
| attack | 3 | 8 | No | Weapon swing/thrust |
| cast | 4 | 6 | No | Spell gesture sequence |
| sneak | 4 | 3 | Yes | Low, slow movement |
| sit | 1 | 0 | No | Static seated pose |
| work | 4 | 4 | Yes | Occupation-specific action |
| dead | 1 | 0 | No | Fallen/collapsed pose |

### 8.3 Animation Update Logic

The `SpriteAnimationSystem` advances animation frames based on elapsed time:

```go
func (s *SpriteAnimationSystem) Update(w *ecs.World, dt float64) {
    for _, e := range w.Entities("Appearance") {
        comp, _ := w.GetComponent(e, "Appearance")
        app := comp.(*components.Appearance)
        if !app.Visible { continue }

        sheet := s.getAnimSheet(app.AnimState)
        app.AnimTimer += dt
        if app.AnimTimer >= 1.0/sheet.FrameRate {
            app.AnimTimer -= 1.0 / sheet.FrameRate
            app.AnimFrame++
            if app.AnimFrame >= len(sheet.Frames) {
                if sheet.Looping {
                    app.AnimFrame = 0
                } else {
                    app.AnimFrame = len(sheet.Frames) - 1
                }
            }
        }
    }
}
```

### 8.4 State Transitions (Driven by Other Systems)

| Source System | Reads | Sets |
|--------------|-------|------|
| NPCScheduleSystem | Schedule.CurrentActivity | AnimState = "work" / "sit" / "walk" |
| CombatSystem | CombatState.IsAttacking | AnimState = "attack" |
| MagicCombatSystem | Spell casting | AnimState = "cast" |
| StealthSystem | Stealth.Sneaking | AnimState = "sneak" |
| VehiclePhysicsSystem | VehiclePhysics.CurrentSpeed | AnimState = "idle" / "walk" |
| HealthSystem | Health.Current ≤ 0 | AnimState = "dead" |
| MovementSystem | Entity speed | AnimState = "walk" / "run" |

---

## 9. Performance Optimization

### 9.1 Level of Detail (LOD)

| Distance (units) | Resolution | Pixel Budget | Approach |
|-------------------|------------|-------------|----------|
| 0–10 | 32×64 | 2,048 px | Full detail |
| 10–25 | 16×32 | 512 px | Reduced detail, skip decorations |
| 25–50 | 8×16 | 128 px | Silhouette + primary color only |
| > 50 | — | 0 | Culled (not rendered) |

LOD transitions use a 2-unit hysteresis band to prevent popping:
- LOD increases at threshold distance
- LOD decreases at threshold - 2 units

### 9.2 Culling

**Frustum culling:** Entities outside the FOV (±30° from center) are skipped.
```go
angleToEntity := math.Atan2(dy, dx) - playerAngle
if math.Abs(normalizeAngle(angleToEntity)) > FOV/2 + spriteAngularWidth/2 {
    continue // outside view frustum
}
```

**Distance culling:** Entities beyond `MaxSpriteDistance` (50.0 units) are skipped.

**Occlusion culling:** Entities fully behind walls (all columns have `zBuffer[col] < entityDist`)
can be skipped early.

### 9.3 Sprite Batching

Entities sharing the same `SpriteCacheKey` use the same generated pixel data. The renderer
only generates unique sprites, then instances them at different screen positions.

### 9.4 Performance Budget

| Metric | Target | Rationale |
|--------|--------|-----------|
| Max visible sprites per frame | 50 | Maintain 60 FPS headroom |
| Sprite render time per frame | < 5 ms | Out of 16.67 ms budget |
| Sprite cache memory | < 20 MB | Within 500 MB total client budget |
| Max cached sprite sheets | 256 | LRU eviction after this |
| Sprite generation time | < 2 ms each | On-demand generation |

### 9.5 Memory Estimation

At full LOD (32×64 × 4 bytes/pixel × average 4 frames):
```
32 × 64 × 4 × 4 = 32,768 bytes per sprite sheet
256 sheets × 32,768 = 8,388,608 bytes ≈ 8 MB
```

Well within the 20 MB budget.

---

## 10. Multiplayer Considerations

### 10.1 Server-Authoritative Appearance

The server owns all `Appearance` component data. Clients receive `Appearance` fields as part
of entity delta updates. The sprite cache key is computed client-side from received fields.

### 10.2 Network Efficiency

Only changed `Appearance` fields are delta-compressed. Typical changes per tick:
- `AnimState` changes: rare (state transitions)
- `AnimFrame` changes: NOT networked (computed client-side from `AnimState` + time)
- `Opacity` changes: rare (stealth transitions)
- `DamageOverlay` changes: rare (on damage events)
- `Visible` changes: rare (spawn/despawn)

**Estimated overhead:** 2–4 bytes per entity per tick (mostly unchanged).

### 10.3 High-Latency Sprite Interpolation (200–5000ms RTT)

Per Wyrm's networking requirements:
- Animation frame advancement runs client-side (no server sync needed)
- Entity position interpolation provides smooth movement between server updates
- State transitions (idle→walk→attack) use the most recent server state
- At >800ms RTT (Tor mode): increase animation blend time to 300ms

---

## 11. Implementation Roadmap

### Phase 1: Foundation (Week 1)

- [x] Add `Appearance` component to `pkg/engine/components/types.go`
- [x] Add `Appearance` unit tests
- [x] Create `pkg/rendering/sprite/` package scaffold
- [x] Define `Sprite` and `SpriteSheet` structs
- [x] Implement `SpriteCache` with LRU eviction
- [x] Expose z-buffer from `drawWalls()` in `Renderer` struct
- [x] Register new systems in `cmd/client/main.go`

### Phase 2: Sprite Generation (Weeks 2–3)

- [x] Implement humanoid silhouette generator with body region mapping
- [x] Implement creature body plan generator (quadruped, biped, serpentine, avian, insectoid)
- [x] Implement vehicle silhouette generator (per genre archetype)
- [x] Implement object sprite generator (container, furniture, sign, resource, door)
- [x] Wire genre palettes into sprite coloring
- [x] Add occupation→body plan mapping for all 5 genres
- [x] Generate animation frame sequences (idle, walk, attack, dead as minimum)
- [x] Unit tests for all generators (determinism tests: same seed → same output)

### Phase 3: Rendering Integration (Weeks 3–4)

- [x] Implement billboard transform (world→screen math)
- [x] Implement depth-tested sprite column drawing against z-buffer
- [x] Implement back-to-front sprite sorting
- [x] Implement frustum culling
- [x] Implement distance culling
- [x] Add fog application to sprites (matching wall fog formula)
- [x] Add opacity support (alpha blending for stealth)
- [x] Add `FlipH` facing-direction support
- [x] Verify post-processing effects apply correctly to sprites

### Phase 4: Animation (Week 4)

- [x] Implement `SpriteAnimationSystem`
- [x] Wire `CombatSystem` → `AnimState` transitions
- [x] Wire `NPCScheduleSystem` → `AnimState` transitions
- [x] Wire `StealthSystem` → `AnimState` / `Opacity` transitions
- [x] Wire `Health` → `AnimState = "dead"` transition
- [x] Wire `VehicleState.DamagePercent` → `DamageOverlay`

### Phase 5: Optimization (Week 5)

- [x] Implement 3-tier LOD system with hysteresis
- [x] Implement occlusion culling
- [x] Profile at 50 visible sprites, 1280×720
- [x] Optimize hot path (cache sprite sheet lookup, minimize allocations)
- [x] Benchmark: `go test -bench=. -benchmem ./pkg/rendering/sprite/`

### Phase 6: Polish & Testing (Week 6)

- [x] Integration tests with full raycaster pipeline
- [x] Determinism tests (same seed → same pixels across 3 runs)
- [x] Multiplayer delta compression for Appearance component
- [x] Weather visual effect integration on sprites
- [x] Edge cases: sprites at screen edges, very close sprites, overlapping sprites
- [x] Update FEATURES.md: mark "Sprite rendering" as complete

---

## 12. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| FPS regression below 60 | Medium | High | Early profiling; LOD + culling; fall back to lower max sprites |
| Z-buffer depth artifacts | Low | High | Careful matching of sprite distance calc to wall distance calc |
| Sprite generation too slow | Low | Medium | Aggressive caching; pre-generate common archetypes at load |
| Sprite/wall seam artifacts | Medium | Medium | Clamp sprite draw bounds; half-pixel bias on edges |
| Animation sync desync in MP | Medium | Low | Client-side only animation; reset on state change |
| Memory pressure from cache | Low | Medium | Strict LRU eviction; 20 MB hard cap; monitor with profiler |
| Genre sprites look too similar | Medium | Low | Distinct silhouette templates per genre; strong palette differentiation |
| Stealth opacity blending | Low | Low | Pre-multiply alpha; test with various opacity values |

---

## 13. Testing Strategy

### 13.1 Unit Tests

| Test Area | Package | Key Tests |
|-----------|---------|-----------|
| Appearance component | `pkg/engine/components/` | Type() returns "Appearance", default values |
| Sprite generation | `pkg/rendering/sprite/` | Determinism (same seed → same pixels × 3 runs) |
| Humanoid generator | `pkg/rendering/sprite/` | Each occupation produces non-empty sprite |
| Creature generator | `pkg/rendering/sprite/` | Each body plan produces valid dimensions |
| Vehicle generator | `pkg/rendering/sprite/` | Each archetype ID produces non-empty sprite |
| Sprite cache | `pkg/rendering/sprite/` | LRU eviction, memory accounting, cache hits |
| Billboard math | `pkg/rendering/raycast/` | Screen position for known world positions |
| Z-buffer | `pkg/rendering/raycast/` | Sprites behind walls are fully occluded |
| Animation | `pkg/engine/systems/` | Frame advancement, looping, terminal states |
| LOD | `pkg/rendering/sprite/` | Correct LOD tier for boundary distances |

### 13.2 Benchmarks

```bash
# Sprite generation (target: <2ms per unique sprite)
go test -bench=BenchmarkSpriteGeneration -benchmem ./pkg/rendering/sprite/

# Sprite rendering (target: <5ms for 50 sprites at 1280×720)
go test -bench=BenchmarkSpriteRendering -benchmem ./pkg/rendering/raycast/

# Sprite cache operations (target: <100ns per lookup)
go test -bench=BenchmarkSpriteCache -benchmem ./pkg/rendering/sprite/
```

### 13.3 Integration Tests

- Full render pipeline: floor/ceiling + walls + sprites + post-processing
- Spawn 50 entities in a test world, render one frame, verify non-zero sprite pixels
- Verify fog application matches between walls and sprites at same distance
- Verify post-processing applies uniformly to wall and sprite pixels

---

## 14. Dependencies

### 14.1 Existing (No New Dependencies)

| Package | Version | Used For |
|---------|---------|----------|
| `github.com/hajimehoshi/ebiten/v2` | v2.8.8 | Rendering target (`*ebiten.Image`) |
| `image/color` | stdlib | RGBA pixel manipulation |
| `math` | stdlib | Trig, floor, abs for billboard math |
| `math/rand` | stdlib | Seeded RNG for sprite generation |
| `sort` | stdlib | Back-to-front sprite ordering |
| `sync` | stdlib | RWMutex for sprite cache |
| `container/list` | stdlib | LRU eviction list |

### 14.2 Internal Dependencies

| Package | Depends On |
|---------|-----------|
| `pkg/rendering/sprite/` | `pkg/rendering/texture/` (genre palettes), `pkg/procgen/noise/` (texture detail) |
| `pkg/rendering/raycast/` | `pkg/rendering/sprite/` (sprite drawing), `pkg/rendering/texture/` (wall textures) |
| `pkg/engine/systems/` | `pkg/engine/components/` (Appearance), `pkg/rendering/sprite/` (cache lookup) |

---

## 15. Appendix: Entity Type Catalog

### 15.1 All Renderable Entity Types

| Category | Entity Type | SpriteCategory | BodyPlan | Scale | Genre Variants |
|----------|-------------|---------------|----------|-------|---------------|
| **NPC** | Merchant | humanoid | merchant | 1.0 | All 5 genres |
| **NPC** | Guard | humanoid | guard | 1.1 | All 5 genres |
| **NPC** | Healer | humanoid | healer | 1.0 | All 5 genres |
| **NPC** | Blacksmith | humanoid | smith | 1.1 | Fantasy |
| **NPC** | Innkeeper | humanoid | innkeeper | 1.0 | Fantasy |
| **NPC** | Farmer | humanoid | farmer | 1.0 | Fantasy, Post-Apoc |
| **NPC** | Miner | humanoid | miner | 1.0 | Fantasy |
| **NPC** | Scribe | humanoid | scribe | 0.9 | Fantasy |
| **NPC** | Bard | humanoid | bard | 1.0 | Fantasy |
| **NPC** | Priest | humanoid | priest | 1.0 | Fantasy, Horror |
| **NPC** | Technician | humanoid | technician | 1.0 | Sci-Fi |
| **NPC** | Scientist | humanoid | scientist | 1.0 | Sci-Fi |
| **NPC** | Pilot | humanoid | pilot | 1.0 | Sci-Fi |
| **NPC** | Engineer | humanoid | engineer | 1.0 | Sci-Fi |
| **NPC** | Medic | humanoid | medic | 1.0 | Sci-Fi, Cyberpunk |
| **NPC** | Mortician | humanoid | mortician | 1.0 | Horror |
| **NPC** | Hunter | humanoid | hunter | 1.0 | Horror, Post-Apoc |
| **NPC** | Herbalist | humanoid | herbalist | 0.9 | Horror |
| **NPC** | Gravedigger | humanoid | gravedigger | 1.1 | Horror |
| **NPC** | Hacker | humanoid | hacker | 0.9 | Cyberpunk |
| **NPC** | Fixer | humanoid | fixer | 1.0 | Cyberpunk |
| **NPC** | Bodyguard | humanoid | bodyguard | 1.2 | Cyberpunk |
| **NPC** | Street Vendor | humanoid | vendor | 1.0 | Cyberpunk |
| **NPC** | Tech Dealer | humanoid | dealer | 1.0 | Cyberpunk |
| **NPC** | Scavenger | humanoid | scavenger | 1.0 | Post-Apoc |
| **NPC** | Mechanic | humanoid | mechanic | 1.0 | Post-Apoc |
| **NPC** | Water Merchant | humanoid | water_merchant | 1.0 | Post-Apoc |
| **Companion** | Tank | humanoid | tank_companion | 1.2 | All 5 genres |
| **Companion** | DPS | humanoid | dps_companion | 1.0 | All 5 genres |
| **Companion** | Healer | humanoid | healer_companion | 1.0 | All 5 genres |
| **Companion** | Support | humanoid | support_companion | 1.0 | All 5 genres |
| **Companion** | Ranged | humanoid | ranged_companion | 1.0 | All 5 genres |
| **Vehicle** | Horse | vehicle | horse | 1.5 | Fantasy |
| **Vehicle** | Horse Cart | vehicle | cart | 2.0 | Fantasy |
| **Vehicle** | Sailing Ship | vehicle | ship | 3.0 | Fantasy |
| **Vehicle** | Hover-Bike | vehicle | hoverbike | 1.3 | Sci-Fi |
| **Vehicle** | Shuttle | vehicle | shuttle | 2.5 | Sci-Fi |
| **Vehicle** | Mech Walker | vehicle | mech | 2.5 | Sci-Fi |
| **Vehicle** | Hearse | vehicle | hearse | 2.0 | Horror |
| **Vehicle** | Bone Cart | vehicle | bonecart | 1.8 | Horror |
| **Vehicle** | Swamp Raft | vehicle | raft | 1.5 | Horror |
| **Vehicle** | Street Bike | vehicle | motorbike | 1.3 | Cyberpunk |
| **Vehicle** | APC | vehicle | apc | 2.5 | Cyberpunk |
| **Vehicle** | Personal Drone | vehicle | drone | 1.0 | Cyberpunk |
| **Vehicle** | Wasteland Buggy | vehicle | buggy | 1.8 | Post-Apoc |
| **Vehicle** | Armored Truck | vehicle | truck | 2.5 | Post-Apoc |
| **Vehicle** | Gyroplane | vehicle | gyroplane | 2.0 | Post-Apoc |
| **Object** | Chest | object | chest | 0.5 | All |
| **Object** | Crate | object | crate | 0.5 | All |
| **Object** | Barrel | object | barrel | 0.5 | All |
| **Object** | Workbench | object | workbench | 0.7 | All |
| **Object** | Sign | object | sign | 0.8 | All |
| **Object** | Door | object | door | 1.2 | All |
| **Object** | Resource Node | object | resource | 0.8 | All |

### 15.2 Creature Body Plans (Future — Post-Creature System)

| BodyPlan | Description | Scale Range | Example Uses |
|----------|-------------|-------------|-------------|
| quadruped | 4-legged animal | 0.5–2.0 | Wolf, horse, dog, rat, bear |
| biped_large | Large humanoid | 1.5–3.0 | Ogre, troll, golem, giant |
| serpentine | Elongated, legless | 0.5–4.0 | Snake, wyrm, tentacle beast |
| avian | Winged biped | 0.8–2.5 | Bird, harpy, griffin, bat |
| insectoid | Multi-legged arthropod | 0.3–2.0 | Giant spider, beetle, scorpion |
| amorphous | Shapeless mass | 0.5–2.0 | Slime, ooze, ghost, wraith |
| dragon | Winged quadruped | 2.0–4.0 | Dragon, wyvern, drake |

---

*End of document.*
