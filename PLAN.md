# PLAN.md — Advanced Raycasting Renderer Enhancement

## 1. Architecture Overview

### Current Renderer Pipeline

```
┌─────────────────────────────────────────────────────────────────────────┐
│ cmd/client/main.go  Game.Draw()                                        │
│                                                                         │
│  ┌──────────────┐    ┌───────────────┐    ┌─────────────────────┐      │
│  │ ClearFrame-  │───▶│ drawFloor-    │───▶│ drawWalls()         │      │
│  │ buffer()     │    │ Ceiling()     │    │  DDA raycasting     │      │
│  └──────────────┘    └───────────────┘    │  ZBuffer population │      │
│                                            └─────────┬───────────┘      │
│                                                      ▼                  │
│  ┌──────────────┐    ┌───────────────┐    ┌─────────────────────┐      │
│  │ Post-process │◀───│ DrawSprites-  │◀───│ WritePixels()       │      │
│  │ Pipeline     │    │ ToScreen()    │    │  framebuffer upload  │      │
│  └──────┬───────┘    └───────────────┘    └─────────────────────┘      │
│         ▼                                                               │
│  ┌──────────────┐    ┌───────────────┐    ┌─────────────────────┐      │
│  │ Particles    │───▶│ Lighting      │───▶│ UI Overlays         │      │
│  │ System       │    │ System        │    │ (HUD, menus)        │      │
│  └──────────────┘    └───────────────┘    └─────────────────────┘      │
└─────────────────────────────────────────────────────────────────────────┘
```

### Proposed Enhanced Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           ENHANCED RENDERER PIPELINE                            │
│                                                                                 │
│  ┌──────────────┐                                                               │
│  │ Mouse Input  │──▶ PlayerA/yaw (existing field) + PlayerPitch (new, ±85°)    │
│  └──────────────┘                                                               │
│         │                                                                       │
│         ▼                                                                       │
│  ┌──────────────┐    ┌───────────────────┐    ┌──────────────────────────┐      │
│  │ ClearFrame-  │───▶│ drawSkybox()      │───▶│ drawFloorCeiling()       │      │
│  │ buffer()     │    │  (replaces ceil    │    │  (pitch-offset rows,     │      │
│  └──────────────┘    │   above horizon)   │    │   material properties)   │      │
│                      └───────────────────┘    └──────────┬───────────────┘      │
│                                                          ▼                      │
│  ┌──────────────────────────────────────────────────────────────────────┐       │
│  │ drawWalls() — ENHANCED                                               │       │
│  │  ┌──────────────┐  ┌─────────────────┐  ┌────────────────────────┐  │       │
│  │  │ DDA with     │  │ Variable height │  │ Partial barrier        │  │       │
│  │  │ HeightMap    │──▶│ wall rendering  │──▶│ transparency pass      │  │       │
│  │  │ lookup       │  │ (0.5x–3x)      │  │ (alpha, gaps, density) │  │       │
│  │  └──────────────┘  └─────────────────┘  └────────────────────────┘  │       │
│  │  ┌──────────────┐  ┌─────────────────┐                              │       │
│  │  │ Normal map   │  │ Material-based  │                              │       │
│  │  │ sampling     │──▶│ shading         │                              │       │
│  │  └──────────────┘  └─────────────────┘                              │       │
│  └──────────────────────────────────────────────────────────────────────┘       │
│         │                                                                       │
│         ▼                                                                       │
│  ┌──────────────────────────────────────────────────────────────────────┐       │
│  │ drawEnvironmentObjects() — NEW                                       │       │
│  │  ┌──────────────┐  ┌─────────────────┐  ┌────────────────────────┐  │       │
│  │  │ Barrier      │  │ Item billboard  │  │ Interactive object     │  │       │
│  │  │ sprites      │──▶│ rendering       │──▶│ highlight pass         │  │       │
│  │  │ (shaped)     │  │ (scale-correct) │  │ (glow outline)        │  │       │
│  │  └──────────────┘  └─────────────────┘  └────────────────────────┘  │       │
│  └──────────────────────────────────────────────────────────────────────┘       │
│         │                                                                       │
│         ▼                                                                       │
│  ┌──────────────┐    ┌───────────────┐    ┌──────────────────────────┐          │
│  │ NPC/Entity   │───▶│ Lighting      │───▶│ Post-process Pipeline    │          │
│  │ Sprites      │    │ (enhanced     │    │ (existing 13 effects     │          │
│  │ (existing)   │    │  materials)   │    │  + interaction highlight) │          │
│  └──────────────┘    └───────────────┘    └──────────────────────────┘          │
│         │                                                                       │
│         ▼                                                                       │
│  ┌──────────────┐    ┌───────────────┐                                          │
│  │ Particles    │───▶│ UI + Cursor   │                                          │
│  │ (weather)    │    │ System        │                                          │
│  └──────────────┘    └───────────────┘                                          │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Key Architectural Changes

| Layer | Current State | Enhanced State |
|-------|---------------|----------------|
| **World Map** | `[][]int` (wall type only, uniform height) | `[][]MapCell` (wall type + height + material + barrier flags) |
| **Ray Result** | `(distance, wallType, wallX, side)` | `(distance, wallType, wallX, side, wallHeight, materialID, barrierFlags)` |
| **Floor/Ceiling** | Fixed horizon at `Height/2` | Pitch-offset horizon at `Height/2 + pitchOffset` |
| **Skybox** | Exists but not integrated into `Draw()` | Rendered to ceiling area above horizon before floor/ceiling pass |
| **Sprites** | Billboard NPC entities only | Billboards + shaped barrier sprites + item sprites + interaction highlights |
| **Materials** | 4 procedural textures (wall types 1-3 + default) | Material registry with per-material texture, normal map, reflectivity |
| **Input** | Keyboard only (`PlayerA` via arrow keys) | Mouse look (yaw + pitch) + keyboard movement |

---

## 2. Implementation Phases

### Phase 1: Core Map Data & Variable Height Walls
**Dependencies:** None (foundational)  
**Estimated Scope:** `pkg/rendering/raycast/`, `pkg/world/chunk/`

**Milestone:** Walls render at variable heights; multi-story buildings visible.

1. Define `MapCell` struct replacing `int` in `WorldMap`
2. Extend `Chunk` with per-cell wall height data
3. Modify `castRayWithTexCoord()` to return wall height
4. Modify `drawWallColumn()` to use per-cell height for `drawStart`/`drawEnd`
5. Render floor/ceiling between adjacent height-mismatched walls
6. Add chunk-to-renderer height data bridging in `SetWorldMap()`
7. Unit tests: variable height rendering, height transitions

### Phase 2: Sky Rendering & Mouse Viewport Control
**Dependencies:** None (parallel with Phase 1)  
**Estimated Scope:** `pkg/rendering/raycast/`, `cmd/client/`, `config/`

**Milestone:** Skybox renders above horizon; mouse controls camera yaw/pitch.

1. Integrate existing `Skybox` into `Draw()` — render sky pixels above horizon line
2. Add `PlayerPitch` field to `Renderer`; offset horizon line by pitch
3. Adjust `drawFloorCeiling()` and `drawWalls()` for pitch-shifted horizon
4. Add mouse input capture in `Game.Update()` using Ebiten's `CursorPosition()`
5. Implement `CursorModeCaptured` for FPS-style mouse capture
6. Add sensitivity/acceleration config to `config.Config`
7. Implement contextual cursor visibility (captured during gameplay, visible for UI)
8. Unit tests: pitch clamping, sky gradient, mouse sensitivity

### Phase 3: Environmental Barriers (Variable Shape)
**Dependencies:** Phase 1 (MapCell, variable heights)  
**Estimated Scope:** `pkg/rendering/raycast/`, `pkg/engine/components/`, `pkg/world/chunk/`

**Milestone:** Natural/constructed barriers render as shaped sprites with collision.

1. Define `BarrierComponent` ECS component with shape, material, genre data
2. Define barrier archetypes per genre (boulders, pillars, hedgerows, wreckage)
3. Implement shaped billboard rendering (non-rectangular silhouettes via alpha masks)
4. Implement polygon-based collision for irregular barrier shapes
5. Add barrier spawn data to chunk `DetailSpawn` system
6. Procedural barrier sprite generation in `pkg/rendering/sprite/`
7. Integration with existing `WorldChunkSystem`
8. Unit tests: collision detection, barrier sprite generation, genre variations

### Phase 4: Partial Barriers & Enhanced Materials
**Dependencies:** Phase 3 (barrier system), Phase 1 (MapCell)  
**Estimated Scope:** `pkg/rendering/raycast/`, `pkg/rendering/texture/`, `pkg/engine/components/`

**Milestone:** Semi-transparent barriers render with density; materials have physical properties.

1. Add barrier permeability flags to `MapCell` (transparency, climbable, destructible)
2. Implement alpha-blended wall rendering for partial barriers
3. Define `MaterialRegistry` with physical properties per material type
4. Implement per-material texture generation with appropriate visual properties
5. Add normal map generation to `texture.GenerateWithSeed()`
6. Implement specular highlight calculation in wall/floor rendering
7. Add surface wear/aging based on world age parameter
8. Genre-specific material palettes (rusty metal, polished chrome, weathered stone)
9. Unit tests: transparency rendering, material property lookups, normal sampling

### Phase 5: Environmental Object Representation
**Dependencies:** Phase 3 (barrier sprites), Phase 4 (materials)  
**Estimated Scope:** `pkg/rendering/raycast/`, `pkg/rendering/sprite/`, `pkg/engine/components/`, `pkg/engine/systems/`

**Milestone:** Items, chests, doors render in world; interaction highlight visible.

1. Categorize environment objects: inventoriable, interactive, decorative
2. Extend `SpriteEntity` with interaction metadata (type, range, highlight state)
3. Implement scale-correct item rendering (items appear correctly sized)
4. Implement interaction highlight effect (glow outline for objects in range)
5. Implement interaction targeting system (raycast from crosshair to determine target)
6. Add `InteractionSystem` ECS system for proximity detection and feedback
7. Procedural item sprite generation matching inventory icons
8. Physics integration for pushable/swinging objects
9. Unit tests: item identification, highlight rendering, interaction raycasting

### Phase 6: Integration, Performance & Polish
**Dependencies:** All previous phases  
**Estimated Scope:** All modified packages

**Milestone:** 60 FPS maintained; all features integrated end-to-end.

1. Performance profiling and optimization pass
2. LOD system for barrier/object detail reduction at distance
3. Frustum culling for environment objects
4. Spatial hash for efficient object/barrier queries
5. Fallback rendering for low-end hardware (disable normal maps, reduce barrier detail)
6. Accessibility: high-contrast interaction highlights, colorblind-friendly item indicators
7. Full integration test suite
8. Benchmark suite for rendering hot paths

### Phase Dependency Graph

```
Phase 1 (Variable Height) ──────────────┬──▶ Phase 3 (Barriers) ──▶ Phase 4 (Partial + Materials)
                                         │                                     │
Phase 2 (Sky + Mouse) ──────────────────┤                                     │
                                         │                                     ▼
                                         └─────────────────────────▶ Phase 5 (Objects)
                                                                              │
                                                                              ▼
                                                                    Phase 6 (Integration)
```

Phases 1 and 2 can proceed in parallel. Phase 3 requires Phase 1. Phase 4 requires Phase 3 and Phase 1. Phase 5 requires Phases 3 and 4. Phase 6 is the final integration pass.

---

## 3. Detailed Feature Specifications

### 3.1 Variable Height Walls

#### Data Structures

```go
// MapCell replaces the int in WorldMap[][]
// File: pkg/rendering/raycast/renderer.go
type MapCell struct {
    WallType   int     // 0=empty, 1-N=wall texture index
    WallHeight float64 // Height multiplier: 0.5=half, 1.0=standard, 3.0=triple
    FloorH     float64 // Floor elevation (0.0=ground level)
    CeilH      float64 // Ceiling height (defaults to WallHeight if 0)
    MaterialID int     // Index into MaterialRegistry
    Flags      uint16  // Bit flags: passable, transparent, climbable, destructible
}

// HeightMap stored alongside the 2D grid
// File: pkg/rendering/raycast/renderer.go (Renderer struct additions)
type Renderer struct {
    // ... existing fields ...
    WorldMapCells [][]MapCell // Replaces WorldMap [][]int
    PlayerPitch   float64     // Vertical look angle (radians, clamped ±85°)
    PlayerZ       float64     // Player eye height (default 0.5 = standing)
}
```

#### Algorithm Changes

**`castRayWithTexCoord()`** — After finding a wall hit via DDA, look up `WorldMapCells[mapX][mapY].WallHeight` to determine the actual wall height multiplier. Return this alongside existing return values.

**`drawWallColumn()`** — Replace the fixed `calculateWallHeight(screenHeight, distance)` with a height-aware version:

```
wallHeight = (screenHeight / distance) * cell.WallHeight
drawStart = horizonLine - wallHeight * (cell.CeilH - playerZ) / cell.WallHeight
drawEnd = horizonLine + wallHeight * (playerZ - cell.FloorH) / cell.WallHeight
```

For multi-story buildings: the `FloorH` and `CeilH` fields allow stacking (floor at 1.0, ceiling at 2.0 for second story). The raycaster checks if the player's Z is between FloorH and CeilH to determine which story is visible.

**Stepped Terrain:** Adjacent cells with different `FloorH` values create visible steps. The renderer draws the exposed side-wall of the step as a horizontal wall strip between the two floor levels.

#### Performance Impact
- **MapCell vs int:** 24 bytes per cell vs 8 bytes → ~3x memory per map cell. For a 512×512 chunk: 6.3 MB vs 2.1 MB. Acceptable within 500 MB budget.
- **DDA lookup:** One additional struct field read per step — negligible.
- **Wall rendering:** Additional multiplication per column for height scaling — negligible.

#### Integration Points
- `pkg/world/chunk/manager.go`: `Chunk.HeightMap` already stores per-cell float64. Extend chunk generation to produce `WallHeight` values from terrain type + noise.
- `pkg/world/chunk/chunk.go`: Add `WallHeights []float64` field (parallel to `HeightMap`).
- `pkg/rendering/raycast/renderer.go`: `SetWorldMap()` must convert chunk data to `MapCell` grid.

#### Genre Variations
| Genre | Height Characteristics |
|-------|----------------------|
| Fantasy | Castle towers (3x), cottage walls (1x), ruins (0.5x-1.5x random) |
| Sci-Fi | Uniform modular buildings (1x, 2x, 3x), observation domes (2x) |
| Horror | Decaying structures (0.8x-1.2x, irregular), crypt walls (0.7x) |
| Cyberpunk | Towering megastructures (3x), slum shacks (0.5x), neon pillars (2x) |
| Post-Apoc | Rubble (0.5x), reinforced shelters (1x), watchtowers (2.5x) |

---

### 3.2 Variable Shape Environmental Barriers

#### Data Structures

```go
// BarrierShape defines the collision and visual profile of a barrier.
// File: pkg/engine/components/definitions.go
type BarrierShape struct {
    ShapeType   string    // "cylinder", "box", "polygon", "billboard"
    Radius      float64   // For cylinder shapes
    Width       float64   // For box shapes
    Depth       float64   // For box shapes
    Height      float64   // World-space height
    Vertices    []float64 // For polygon shapes: [x0,y0, x1,y1, ...] relative to center
    SpriteKey   string    // Key into sprite cache for visual representation
    MaterialID  int       // Material for collision sound/effects
}

// Barrier is an ECS component for environmental barriers.
// File: pkg/engine/components/definitions.go
type Barrier struct {
    Shape       BarrierShape
    Genre       string  // Genre that generated this barrier
    Destructible bool
    HitPoints   float64 // For destructible barriers
    MaxHP       float64
}
```

#### Barrier Archetypes by Genre

| Category | Fantasy | Sci-Fi | Horror | Cyberpunk | Post-Apoc |
|----------|---------|--------|--------|-----------|-----------|
| Natural | Boulders, ancient trees, crystal formations | Alien rock, fungal growths, crystal nodes | Gnarled trees, bone piles, pulsing hives | Toxic waste drums, mutant flora | Rubble mounds, burnt trees, craters |
| Constructed | Stone pillars, archways, statues | Steel beams, energy pylons, antenna arrays | Iron gates, tombstones, ritual circles | Neon signs, holographic walls, vending machines | Barricades, wrecked cars, makeshift walls |
| Organic | Hedgerows, thornbushes, vine walls | Bio-pods, growth membranes, tendril curtains | Flesh walls, web clusters, fungal masses | Gang graffiti barriers, plant walls | Overgrown ruins, thorn thickets |

#### Algorithm: Shaped Billboard Rendering

Barriers use **shaped billboards** — sprites with alpha-mask silhouettes that are wider than a single grid cell. Unlike NPC billboards (always face camera), barrier billboards are rendered with perspective-correct width based on their `BarrierShape`.

1. During the entity sprite pass, barriers are sorted alongside NPCs by distance.
2. For each barrier, compute screen bounds using `GetSpriteScreenBounds()` with the barrier's width/height.
3. Sample the barrier's sprite with its alpha mask to produce the silhouette.
4. The alpha mask is generated procedurally from the `ShapeType` and `Vertices` data.

#### Algorithm: Polygon Collision Detection

For irregular barrier shapes, collision uses a 2D polygon intersection test:

1. Each barrier's `Vertices` define a convex hull in world-space relative to the barrier's center.
2. Player movement checks: for each movement vector, test line-segment vs polygon edge intersection.
3. Use separating axis theorem (SAT) for convex polygon vs circle (player bounding circle) collision.
4. Cylinder and box shapes use optimized fast-path checks (circle-circle, AABB).

#### Performance Impact
- **Barrier rendering:** Same cost as NPC sprite rendering (billboard transform + column draw). With 50 barriers visible: ~50× sprite column cost. Mitigated by frustum culling and distance culling.
- **Collision:** SAT test per barrier within player's cell neighborhood (3×3 grid). Typically <20 barriers in range. Sub-microsecond per test.

#### Integration Points
- `pkg/engine/components/definitions.go`: New `Barrier` component.
- `pkg/engine/systems/`: Barriers consumed by `WorldChunkSystem` (spawning) and collision system.
- `pkg/world/chunk/manager.go`: `DetailSpawn` extended with `BarrierShape` data for spawning.
- `pkg/rendering/sprite/generator.go`: New barrier sprite generation functions.

---

### 3.3 Partial Environmental Barriers

#### Data Structures

```go
// BarrierFlags bit constants for MapCell.Flags and Barrier properties.
// File: pkg/rendering/raycast/renderer.go
const (
    FlagSolid       uint16 = 1 << iota // Full collision
    FlagPassable                         // Can walk through (tall grass, shallow water)
    FlagTransparent                      // Rendered with alpha (ice, force fields)
    FlagClimbable                        // Can climb over (low walls, debris)
    FlagDestructible                     // Can be destroyed
    FlagSemiOpaque                       // Partial opacity (reeds, broken fence)
)

// PartialBarrierProperties extends Barrier with partial-permeability data.
// File: pkg/engine/components/definitions.go
type PartialBarrierProperties struct {
    Opacity       float64 // 0.0=fully transparent, 1.0=fully opaque
    Density       float64 // Material density for movement speed penalty
    GapPattern    string  // "none", "random_holes", "lattice", "vertical_bars"
    GapDensity    float64 // 0.0=no gaps, 1.0=mostly gaps
    ClimbHeight   float64 // Max height player can climb over
    BreakThreshold float64 // Damage needed to destroy
}
```

#### Algorithm: Alpha-Blended Wall Rendering

For walls/barriers with `FlagTransparent` or `FlagSemiOpaque`:

1. During `renderWallStrip()`, after sampling the wall texture color, check `MapCell.Flags`.
2. If transparent: apply `cell.Opacity` to the alpha channel. Blend with the sky/floor color behind.
3. If semi-opaque with gap pattern: use a procedural gap mask (based on seed + position) to determine per-pixel opacity. Pixels in "gap" regions get alpha 0 (show through to background).
4. For lattice patterns: `texX % spacing < bar_width` creates vertical bars; combine with horizontal for lattice.

**Rendering order change:** Partial barriers require a **two-pass approach**:
- Pass 1: Render all opaque walls (existing behavior, populates ZBuffer).
- Pass 2: Render partial barriers with alpha blending over the existing framebuffer content.

This avoids the need to sort walls by distance (which DDA handles implicitly for opaques).

#### Climbable Objects

When the player approaches a `FlagClimbable` barrier:
1. Check `barrier.ClimbHeight` vs player step height (configurable, default 0.5 world units).
2. If climbable: smoothly adjust `PlayerZ` over 0.3 seconds to rise over the barrier.
3. On the other side: smoothly return `PlayerZ` to ground level.

This reuses the `PlayerZ` field added for variable-height walls.

#### Destructible Elements

Destructible barriers have `HitPoints`. When attacked:
1. Reduce `HitPoints` by weapon damage.
2. Update `DamageOverlay` on the barrier's `Appearance` component.
3. At 50% HP: switch sprite to "damaged" variant (cracks, gaps increase).
4. At 0 HP: remove barrier entity, spawn debris particles, play destruction sound.

#### Performance Impact
- **Two-pass walls:** The second pass only touches partial barriers (typically <10% of walls). Minimal overhead.
- **Gap pattern calculation:** Per-pixel modulo operation — negligible.
- **Climb animation:** Only during player transition — no per-frame cost otherwise.

#### Genre Variations
| Genre | Semi-Permeable | Damaged Structures | Climbable | Transparent | Destructible |
|-------|---------------|-------------------|-----------|-------------|-------------|
| Fantasy | Tall grass, reed beds | Crumbling castle walls | Low stone walls, fallen trees | Ice walls, magic barriers | Wooden barricades, ice |
| Sci-Fi | Energy fields (low) | Damaged hull plating | Cargo crates, ledges | Force fields, glass panels | Glass panels, weak plating |
| Horror | Fog banks, cobwebs | Rotting walls, broken boards | Gravestones, debris piles | Ghostly barriers, thin walls | Rotten wood, brittle bone |
| Cyberpunk | Holographic ads, smoke | Broken neon signs, cracked glass | Dumpsters, pipe stacks | Holographic walls, glass | Cheap barriers, glass |
| Post-Apoc | Irradiated grass, ash clouds | Collapsed buildings | Rubble, wrecked cars | Thin sheet metal with holes | Rusted barriers, weak structures |

---

### 3.4 Enhanced Material Representation

#### Data Structures

```go
// MaterialProperties defines physical rendering properties.
// File: pkg/rendering/texture/material.go (new file)
type MaterialProperties struct {
    ID            int
    Name          string   // "stone", "wood", "metal", "glass", "fabric", "organic"
    Roughness     float64  // 0.0=mirror, 1.0=fully rough
    Metallic      float64  // 0.0=dielectric, 1.0=metallic
    Reflectivity  float64  // Specular reflection strength (0.0-1.0)
    Transparency  float64  // 0.0=opaque, 1.0=fully transparent
    EmissiveStr   float64  // Self-illumination strength (neon signs, lava)
    NormalStrength float64 // Normal map influence (0.0=flat, 1.0=full)
    WearFactor    float64  // 0.0=pristine, 1.0=heavily worn
    AgeMultiplier float64  // How fast this material visually ages
}

// MaterialRegistry manages all material types and their textures.
// File: pkg/rendering/texture/material.go (new file)
type MaterialRegistry struct {
    Materials     map[int]*MaterialProperties
    Textures      map[int]*Texture      // Albedo textures per material
    NormalMaps    map[int]*Texture      // Normal maps per material
    GenrePalettes map[string]map[int]GenreMaterialOverride
}

// GenreMaterialOverride adjusts material appearance per genre.
type GenreMaterialOverride struct {
    TintColor    color.RGBA
    WearBoost    float64  // Additional wear for this genre
    AgeBoost     float64  // Additional aging
    PaletteShift float64  // Hue shift for genre palette
}
```

#### Algorithm: Normal Map Sampling

Normal maps are procedurally generated alongside albedo textures. During wall rendering:

1. Sample normal map at `(texX, texY)` to get surface normal perturbation `(nx, ny, nz)`.
2. Transform the normal from tangent space to world space using the wall's orientation (side 0 = X-facing, side 1 = Y-facing).
3. Compute light direction from the lighting system's sun/point lights.
4. Apply `dot(normal, lightDir) * lightIntensity` as a brightness modifier.

The normal map is a `Texture` where RGB channels encode the normal vector: `R=nx*127+128, G=ny*127+128, B=nz*127+128`.

#### Algorithm: Specular Highlights

For materials with `Reflectivity > 0`:

1. Compute the reflection vector: `R = 2 * dot(N, L) * N - L`.
2. Compute specular intensity: `spec = pow(max(dot(R, viewDir), 0), shininess)`.
3. `shininess = (1.0 - Roughness) * 64.0` (rougher = wider, dimmer highlights).
4. Add `spec * Reflectivity * lightColor` to the final pixel color.

This is a simplified Blinn-Phong model suitable for CPU-based per-pixel computation.

#### Algorithm: Procedural Wear & Aging

Surface wear is applied as a texture-space modification:

1. Generate a "wear noise" texture at material creation time (low-frequency Perlin noise).
2. `wearIntensity = WearFactor * AgeMultiplier * worldAge`.
3. Where wear noise exceeds a threshold based on `wearIntensity`: darken the albedo, increase roughness, add color shift toward grey/brown.
4. Edge wear: increase wear at texture edges (top/bottom rows of wall textures) to simulate erosion.

#### Performance Impact
- **Normal map sampling:** One additional texture lookup per pixel + dot product + multiply. Approximately 2× the per-pixel cost of albedo-only rendering. At 1280×720 with ~30% wall pixels: ~276K additional lookups per frame. At 1ns per lookup: ~0.3ms. Acceptable.
- **Specular highlights:** One pow() call per pixel where specular is nonzero. Roughly ~10% of wall pixels have specular. Cost: ~0.1ms.
- **Material registry:** O(1) lookup per ray hit. Negligible.

**Optimization:** Normal maps and specular can be disabled per-quality-level for fallback rendering.

#### Genre-Specific Material Palettes

| Material | Fantasy | Sci-Fi | Horror | Cyberpunk | Post-Apoc |
|----------|---------|--------|--------|-----------|-----------|
| Stone | Warm grey, mossy | Clean grey, precise | Dark grey, stained | Concrete, tagged | Cracked, dusty |
| Wood | Rich brown, carved | — (rare) | Rotting, dark | — (rare) | Weathered, splintered |
| Metal | Bronze/iron, patina | Chrome, brushed steel | Rusted iron, corroded | Polished chrome, neon-lit | Rusted, dented, salvaged |
| Glass | Stained (colorful) | Clear, blue-tinted | Cracked, dirty | Neon-reflective | Shattered, yellowed |
| Fabric | Tapestry, banners | Synthetic, clean | Torn, blood-stained | Synthetic, LED-threaded | Patched, faded |
| Organic | Vines, bark | Bio-tech, membrane | Flesh, bone, fungus | — (rare) | Mutant growth, lichen |

---

### 3.5 Environmental Object Representation

#### Data Structures

```go
// EnvironmentObject extends SpriteEntity with interaction data.
// File: pkg/rendering/raycast/billboard.go (additions)
type EnvironmentObject struct {
    SpriteEntity                // Embedded billboard
    ObjectType    string        // "item", "interactive", "decorative"
    InteractRange float64       // Max interaction distance
    HighlightState float64      // 0.0=no highlight, 1.0=full highlight
    ItemID        string        // For inventoriable items: matches inventory item ID
    InteractionID string        // For interactive objects: "open_chest", "pull_lever", etc.
}

// InteractionTarget holds the result of an interaction raycast.
// File: pkg/rendering/raycast/renderer.go (additions)
type InteractionTarget struct {
    Entity    uint64  // ECS entity ID
    Distance  float64
    ScreenX   int     // Screen position of target center
    ScreenY   int
    Type      string  // "item", "interactive", "decorative"
}
```

#### Algorithm: Interaction Targeting

Each frame, cast a ray from the screen center (crosshair) into the world:

1. Use the same DDA algorithm as wall rendering, but for the center column only.
2. After the ray completes, check all `EnvironmentObject` entities within the ray path.
3. For each object: test if the ray passes within `object.Radius` of the object's world position.
4. Return the closest intersecting object within `InteractRange`.

This is a single additional ray cast per frame — negligible cost.

#### Algorithm: Highlight Rendering

For objects with `HighlightState > 0`:

1. After drawing the object's sprite to the framebuffer, perform an edge-detection pass on the sprite's screen region.
2. For each pixel on the sprite's boundary (where alpha transitions from >0 to 0): write a highlight color.
3. The highlight color uses the genre's accent color (gold for fantasy, cyan for sci-fi, red for horror, neon pink for cyberpunk, orange for post-apoc).
4. Pulse the highlight intensity using `sin(time * 3.0) * 0.3 + 0.7` for a subtle breathing effect.

**Optimization:** Only compute highlight for the one currently-targeted object, not all interactive objects.

#### Scale-Appropriate Rendering

Items must appear at correct real-world scale:

| Object Category | World Height | Scale Factor |
|----------------|-------------|-------------|
| Small items (keys, potions, coins) | 0.15 world units | 0.15 |
| Medium items (swords, books, tools) | 0.4 world units | 0.4 |
| Large items (shields, staves) | 0.6 world units | 0.6 |
| Furniture (chairs, tables) | 0.8 world units | 0.8 |
| Large objects (chests, workbenches) | 0.5 world units | 0.5 |
| Doors | 1.0 world units (full wall height) | 1.0 |

The `Scale` field on `SpriteEntity` is set based on the item's category during placement.

#### Physics Integration

Pushable objects (crates, barrels):
1. On player collision with a pushable object: apply force in the player's movement direction.
2. Move the object's `Position` component by `pushForce * dt` in the push direction.
3. Check collision of the pushed object against walls and other barriers.
4. Limit push speed to prevent objects from phasing through walls.

Swinging doors:
1. Doors have a `rotation` field in addition to position.
2. On interaction: animate the rotation from 0° to 90° over 0.5 seconds.
3. Update the door's collision polygon each frame during animation.
4. After animation: the door remains in the open state until interacted with again.

#### Integration Points
- `pkg/engine/components/definitions.go`: Extend or use existing `Appearance` component. No new component needed — use `SpriteCategory = "object"` with `BodyPlan` for item type.
- `pkg/engine/systems/`: `RenderSystem` already handles sprite rendering. Extend with interaction targeting.
- `cmd/client/main.go`: Add crosshair rendering and interaction key binding.

---

### 3.6 Sky Rendering System

#### Current State

The `Skybox` struct in `pkg/rendering/raycast/skybox.go` is **fully implemented** with:
- Genre-specific color palettes (5 genres × 10 colors)
- Time-of-day transitions (night → dawn → day → dusk → night)
- Celestial body positioning (sun parabolic arc, moon opposite)
- Weather effects (clear, overcast, rain, storm, snow, fog)
- Indoor mode

**Gap:** The skybox is not currently called from `Draw()`. The ceiling area renders as texture-mapped ceiling instead of sky.

#### Integration Plan

1. In `drawFloorCeiling()`: for rows above the horizon line (plus pitch offset), call `skybox.GetSkyColorAt(x, y)` instead of `GetCeilingTextureColor()`.
2. When `skybox.IsIndoor()` is true: fall back to existing ceiling texture rendering.
3. Connect `WeatherSystem` ECS output to `skybox.SetWeather()` each frame.
4. Connect `WorldClockSystem` output to `skybox.SetTimeOfDay()` each frame.

#### Enhancements

**Stars:** Add a star field for nighttime sky. Stars are rendered as bright pixels at fixed celestial coordinates (generated from seed). Stars fade in as `timeOfDay` approaches full night and fade out at dawn.

```go
// StarField generates deterministic star positions.
// File: pkg/rendering/raycast/skybox.go (additions)
type StarField struct {
    Stars []Star // Pre-generated from seed
}

type Star struct {
    X, Y       float64 // Normalized sky position (0-1)
    Brightness float64 // 0.0-1.0
    Color      color.RGBA
}
```

**Dynamic Lighting Influence:** The skybox sun position feeds into the lighting system's directional light. This connection already exists conceptually in `pkg/rendering/lighting/system.go` (`sun *Light`). Wire the skybox's sun position to the lighting system's directional light direction.

#### Performance Impact
- **Sky rendering:** Replaces ceiling texture sampling with gradient interpolation — approximately equal cost. Celestial body check adds one distance calculation per sky pixel, but short-circuits for pixels far from sun/moon.
- **Star rendering:** Only at night. ~200 stars × one pixel each = trivial.

---

### 3.7 Mouse-Based Viewport Control

#### Data Structures

```go
// MouseConfig holds mouse control settings.
// File: config/config.go (additions to Config struct)
type MouseConfig struct {
    Sensitivity    float64 // Base sensitivity multiplier (default 0.003)
    Acceleration   float64 // Mouse acceleration curve (0.0=none, 1.0=full, default 0.0)
    InvertY        bool    // Invert vertical axis (default false)
    SmoothingFrames int    // Number of frames to smooth input over (default 2)
    PitchLimitDeg  float64 // Max vertical look angle in degrees (default 85)
}
```

#### Algorithm: Mouse Look Implementation

Ebitengine provides `ebiten.CursorPosition()` for cursor position and `ebiten.SetCursorMode()` for cursor capture.

**Per-frame in `Game.Update()`:**

1. Read `ebiten.CursorPosition()` to get current cursor `(cx, cy)`.
2. Compute delta: `dx = cx - screenCenterX`, `dy = cy - screenCenterY`.
3. Apply sensitivity: `yawDelta = dx * sensitivity`, `pitchDelta = dy * sensitivity * (invertY ? -1 : 1)`.
4. Apply optional acceleration: `if |dx| > threshold: yawDelta *= 1.0 + acceleration * (|dx| / maxDelta)`.
5. Apply smoothing: average the last N frame deltas.
6. Update player angle: `PlayerA += yawDelta` (wrap to 0–2π). Note: `PlayerA` is the existing field name in `Renderer` (renderer.go:39). Consider renaming to `PlayerYaw` for consistency with the new `PlayerPitch`, but this is a separate refactoring task.
7. Update player pitch: `PlayerPitch = clamp(PlayerPitch + pitchDelta, -pitchLimit, +pitchLimit)`.
8. Reset cursor to screen center: use `ebiten.SetCursorMode(ebiten.CursorModeCaptured)` which automatically captures the cursor.

**Cursor Visibility:**
- During gameplay: `CursorModeCaptured` — cursor hidden, deltas computed from movement.
- During UI (inventory, menu, dialog): `CursorModeVisible` — cursor shown, used for UI interaction.
- Toggle via `Escape` key or UI open/close events.

#### Pitch Integration with Renderer

The `PlayerPitch` value shifts the rendering horizon:

```
pitchOffset = int(PlayerPitch / maxPitch * float64(Height / 2))
horizonLine = Height/2 + pitchOffset
```

- `drawFloorCeiling()`: The floor starts at `horizonLine` instead of `Height/2`. The ceiling (or sky) fills from 0 to `horizonLine`.
- `drawWallColumn()`: `drawStart` and `drawEnd` are offset by `pitchOffset`.
- `drawSpriteToFramebuffer()`: Sprite vertical position offset by `pitchOffset`.

This is the standard technique used in classic raycasters (Wolfenstein 3D-style) for pitch simulation. It provides convincing vertical look without true 3D projection.

#### Aim Assistance

For interaction targeting (Section 3.5), the crosshair position is always screen center. When an interactable object is within range and near the crosshair:

1. Compute angular distance from crosshair ray to object center.
2. If within `aimAssistAngle` (configurable, default 3°): snap the interaction target to that object.
3. Display a subtle reticle expansion to indicate aim assist is active.

This does NOT move the camera — only the interaction target selection is assisted.

#### Performance Impact
- **Mouse input:** One `CursorPosition()` call + arithmetic per frame. Negligible.
- **Pitch offset:** One addition per row/column in rendering. Negligible.

---

## 4. Code Modification Breakdown

### `pkg/rendering/raycast/renderer.go`

| Change | Type | Description |
|--------|------|-------------|
| Add `MapCell` struct | New type | Replace `int` wall type with rich cell data |
| Add `WorldMapCells` field | New field | Parallel to existing `WorldMap`, stores `MapCell` grid |
| Add `PlayerPitch` field | New field | Vertical look angle for mouse pitch |
| Add `PlayerZ` field | New field | Player eye height for variable-height rendering |
| Modify `SetWorldMap()` | Edit | Accept height data alongside heightmap, populate `MapCell` grid |
| Add `SetWorldMapCells()` | New method | Direct setter for `MapCell` grid |
| Add `castRayEnhanced()` | New method | Returns `MapCell` data instead of just wall type |
| Add `MaterialRegistry` integration | New field | Pointer to shared `MaterialRegistry` |

### `pkg/rendering/raycast/draw.go`

| Change | Type | Description |
|--------|------|-------------|
| Modify `Draw()` | Edit | Add skybox pass before floor/ceiling, pass pitch offset |
| Modify `drawFloorCeiling()` | Edit | Use `horizonLine` (pitch-adjusted) instead of `Height/2`, call skybox for ceiling pixels |
| Modify `drawWalls()` | Edit | Use `MapCell` height for per-column wall height calculation |
| Modify `drawWallColumn()` | Edit | Variable height + normal map + specular calculation |
| Modify `renderWallStrip()` | Edit | Material-aware shading, alpha blending for partial barriers |
| Add `drawPartialBarriers()` | New method | Second pass for transparent/semi-opaque walls |
| Add `drawEnvironmentObjects()` | New method | Render barrier sprites, items, interactive objects |
| Add `drawInteractionHighlight()` | New method | Glow outline for targeted interactive object |

### `pkg/rendering/raycast/skybox.go`

| Change | Type | Description |
|--------|------|-------------|
| Add `StarField` struct | New type | Deterministic star positions |
| Add `RenderToFramebuffer()` | New method | Write sky pixels directly to framebuffer for ceiling area |
| Add star rendering | New method | Render stars during nighttime |
| Wire into `Draw()` pipeline | Edit | Called from `draw.go` during ceiling pass |

### `pkg/rendering/raycast/billboard.go`

| Change | Type | Description |
|--------|------|-------------|
| Add `EnvironmentObject` struct | New type | Extended sprite with interaction data |
| Add `CastInteractionRay()` | New method | Single center-screen ray for interaction targeting |
| Modify `TransformEntityToScreen()` | Edit | Apply pitch offset to sprite vertical position |
| Add `DrawHighlight()` | New method | Edge-detect and glow for interaction highlight |

### `pkg/rendering/texture/` (new file: `material.go`)

| Change | Type | Description |
|--------|------|-------------|
| `MaterialProperties` struct | New type | Physical material properties |
| `MaterialRegistry` struct | New type | Material type registry with textures and normal maps |
| `GenerateNormalMap()` | New function | Procedural normal map generation from heightfield noise |
| `GenerateMaterialTexture()` | New function | Genre-aware material texture with wear/aging |
| `NewMaterialRegistry()` | New constructor | Initialize with standard materials |

### `pkg/rendering/texture/generator.go`

| Change | Type | Description |
|--------|------|-------------|
| Add `GenerateNormalMapWithSeed()` | New function | Normal map variant of texture generation |
| Add wear/aging overlay | Edit | Apply surface degradation based on age parameter |

### `cmd/client/main.go`

| Change | Type | Description |
|--------|------|-------------|
| Add mouse input handling | New code | `CursorPosition()` delta computation in `Update()` |
| Add `CursorModeCaptured` | New code | Mouse capture toggling |
| Add skybox integration | New code | Connect skybox to renderer, set time/weather each frame |
| Add interaction targeting | New code | Center-screen raycast + highlight management |
| Add crosshair rendering | New code | Simple crosshair drawn at screen center |

### `config/config.go`

| Change | Type | Description |
|--------|------|-------------|
| Add `MouseConfig` | New struct | Mouse sensitivity, acceleration, invert, smoothing |
| Add `RenderingConfig` | New struct | Quality levels for normal maps, specular, barrier detail |
| Add to `Config` struct | Edit | New fields for mouse and rendering config |

---

## 5. ECS Integration

### New Components

| Component | Type String | Fields | Purpose |
|-----------|-------------|--------|---------|
| `Barrier` | `"Barrier"` | Shape, Genre, Destructible, HitPoints, MaxHP | Environmental barrier collision and rendering |
| `Interactable` | `"Interactable"` | InteractionType, Range, Prompt, Cooldown, Locked | Objects the player can interact with |
| `WorldItem` | `"WorldItem"` | ItemID, Quantity, SpawnTime, Respawnable | Items placed in the world for pickup |
| `PhysicsBody` | `"PhysicsBody"` | Mass, Velocity, Pushable, Friction | Simple physics for pushable objects |

**Note:** The `Interactable` component does not currently exist in `pkg/engine/components/definitions.go` — it must be created as a new component. The `WorldItem` component is also new.

### New Systems

| System | Consumes | Produces | Priority |
|--------|----------|----------|----------|
| `BarrierCollisionSystem` | Position, Barrier | Position (clamped) | Before movement systems |
| `InteractionTargetSystem` | Position, Interactable, WorldItem | InteractionTarget (renderer state) | After movement, before render |
| `BarrierDestructionSystem` | Barrier, Health | Particle spawn, entity removal | After combat |
| `ObjectPhysicsSystem` | PhysicsBody, Position, Barrier | Position (updated) | After interaction |

### Component Interaction Diagram

```
┌─────────────┐     reads      ┌──────────────┐
│ Position    │◀──────────────│ BarrierColl- │
│ (Player)    │──────────────▶│ isionSystem  │
└─────────────┘     modifies   └──────────────┘
                                       │ reads
                                       ▼
                               ┌──────────────┐
                               │ Barrier      │
                               │ (all barriers)│
                               └──────────────┘
                                       │
                                       ▼
┌─────────────┐     reads      ┌──────────────┐      produces     ┌──────────────┐
│ Position    │◀──────────────│ Interaction- │──────────────────▶│ Renderer     │
│ (objects)   │               │ TargetSystem │                   │ highlight    │
└─────────────┘               └──────────────┘                   └──────────────┘
       ▲                              │ reads
       │                              ▼
┌─────────────┐               ┌──────────────┐
│ Interactable│               │ WorldItem    │
└─────────────┘               └──────────────┘

┌─────────────┐     reads      ┌──────────────┐      produces     ┌──────────────┐
│ Barrier     │◀──────────────│ Barrier      │──────────────────▶│ Particles    │
│ (destructed)│               │ Destruction  │                   │ (debris)     │
└─────────────┘               │ System       │                   └──────────────┘
                              └──────────────┘
                                      │ reads
                                      ▼
                              ┌──────────────┐
                              │ CombatSystem │
                              │ (damage)     │
                              └──────────────┘
```

### Registration in `cmd/client/main.go` and `cmd/server/main.go`

All new systems must be registered:

```go
// Barrier systems (server + client)
world.RegisterSystem(&systems.BarrierCollisionSystem{})
world.RegisterSystem(&systems.BarrierDestructionSystem{})
world.RegisterSystem(&systems.ObjectPhysicsSystem{})

// Interaction systems (client only)
world.RegisterSystem(&systems.InteractionTargetSystem{})
```

---

## 6. Testing Strategy

### Unit Tests

| Test File | Package | Tests |
|-----------|---------|-------|
| `renderer_height_test.go` | `raycast` | Variable height wall rendering, multi-story buildings, height transitions |
| `mapcell_test.go` | `raycast` | MapCell creation, flag operations, material lookup |
| `pitch_test.go` | `raycast` | Pitch offset calculation, horizon clamping, pitch limits |
| `skybox_integration_test.go` | `raycast` | Sky renders above horizon, indoor fallback, star field |
| `barrier_collision_test.go` | `raycast` (or `systems`) | Polygon SAT collision, cylinder collision, AABB collision |
| `material_test.go` | `texture` | Material registry, normal map generation, specular calculation |
| `partial_barrier_test.go` | `raycast` | Alpha blending, gap patterns, transparency rendering |
| `interaction_ray_test.go` | `raycast` | Center-screen ray, object targeting, range checking |
| `highlight_test.go` | `raycast` | Edge detection, glow rendering |
| `mouse_input_test.go` | `client` | Sensitivity, acceleration, smoothing, pitch clamping |
| `barrier_component_test.go` | `components` | Component creation, Type() string, flag operations |
| `barrier_system_test.go` | `systems` | System Update() with mock world, collision resolution |

### Integration Tests

| Test | Scope | Validates |
|------|-------|-----------|
| Variable height chunk rendering | `chunk` + `raycast` | Chunks generate height data → renderer displays variable walls |
| Skybox + weather | `raycast` + `systems` | WeatherSystem output → skybox color changes |
| Barrier spawn + collision | `chunk` + `components` + `systems` | Chunk generates barriers → collision system prevents walkthrough |
| Item pickup flow | `components` + `systems` | WorldItem targeted → interaction → added to Inventory |
| Material + lighting | `texture` + `lighting` + `raycast` | Normal maps + directional light → correct shading |

### Performance Benchmarks

| Benchmark | Target | Measures |
|-----------|--------|----------|
| `BenchmarkDrawWallsVariableHeight` | <8ms per frame (1280×720) | Wall rendering with height lookups |
| `BenchmarkDrawWallsWithNormals` | <12ms per frame | Wall rendering with normal map + specular |
| `BenchmarkBarrierCollision50` | <0.1ms | 50 barrier SAT collision checks |
| `BenchmarkSkyboxRender` | <2ms per frame | Full skybox with celestial bodies |
| `BenchmarkPartialBarrierPass` | <3ms per frame | Second-pass alpha blending for 20 partial barriers |
| `BenchmarkInteractionRay` | <0.05ms | Single center-screen interaction raycast |
| `BenchmarkMaterialRegistryLookup` | <10ns | Material property lookup by ID |

### Determinism Tests

| Test | Validates |
|------|-----------|
| `TestDeterministicBarrierSpawn` | Same seed → same barrier positions, shapes, materials |
| `TestDeterministicMaterialGeneration` | Same seed → identical textures and normal maps |
| `TestDeterministicStarField` | Same seed → identical star positions |

All tests run with `go test -tags=noebiten -count=1 ./...` for headless CI. Rendering-specific tests that require Ebiten use build tags and run under `xvfb` in CI.

---

## 7. Asset Pipeline (Zero External Assets)

All visual content is procedurally generated. No image files, model files, or external data are added.

### Procedural Generation Chain

| Asset Type | Generator | Input | Cache Strategy |
|------------|-----------|-------|----------------|
| **Wall textures** | `texture.GenerateWithSeed()` | seed + genre + materialID | Keyed by `(seed, genre, materialID)` in `textureCache` |
| **Normal maps** | `texture.GenerateNormalMapWithSeed()` | seed + genre + materialID | Keyed alongside albedo texture |
| **Barrier sprites** | `sprite.GenerateBarrier()` | seed + genre + shapeType + variation | LRU cache in `SpriteCache` |
| **Item sprites** | `sprite.GenerateItem()` | seed + genre + itemCategory + itemType | LRU cache in `SpriteCache` |
| **Star positions** | `StarField.Generate()` | seed | Generated once at startup, stored in `Skybox` |
| **Material palettes** | `MaterialRegistry.Init()` | genre | Generated once per genre change |

### Texture Generation Pipeline

```
Base noise (Perlin 2D, seeded)
     │
     ▼
Genre palette mapping (color from genre palette table)
     │
     ▼
Material modification (roughness darkens, metallic adds specular texture)
     │
     ▼
Wear/aging overlay (additional noise layer, genre-specific aging)
     │
     ▼
Normal map derivation (gradient of noise → normal vectors)
     │
     ▼
Cache storage (albedo + normal map stored as pair)
```

### Barrier Sprite Generation

Barrier sprites use the existing `sprite.Generator` with a new `CategoryBarrier` mode:

1. Generate a base shape silhouette from `BarrierShape.ShapeType`:
   - `"cylinder"` → oval silhouette
   - `"box"` → rectangular silhouette  
   - `"polygon"` → custom silhouette from vertices
   - `"billboard"` → rectangular with alpha-mask edges
2. Fill silhouette with material texture (sampled from `MaterialRegistry`).
3. Add genre-appropriate detail overlays (moss for fantasy, rust for post-apoc, neon for cyberpunk).
4. Generate multiple variations per archetype (3-5 variations) for visual diversity.

### Item Sprite Generation

Item sprites match their inventory representation:

1. Generate item silhouette from `BodyPlan` (sword shape, potion shape, book shape).
2. Apply genre palette colors.
3. Add material-appropriate texture fill (metal sheen for weapons, leather for armor).
4. Scale to world-appropriate size (see Section 3.5 scale table).
5. Store a thumbnail variant for inventory UI (same silhouette, smaller resolution).

---

## 8. Fallback Systems

### Quality Levels

```go
// RenderQuality configures rendering detail.
// File: config/config.go
type RenderQuality struct {
    Level          string // "low", "medium", "high"
    NormalMaps     bool   // Enable normal map sampling
    Specular       bool   // Enable specular highlights
    BarrierDetail  int    // 0=simple boxes, 1=shaped, 2=full detail
    SkyStars       bool   // Enable star rendering
    ParticleCount  int    // Max particles (100/500/2000)
    InteractionGlow bool  // Enable glow highlight effect
    ShadowQuality  int    // 0=none, 1=simple, 2=full
}
```

### Degradation Tiers

| Feature | High Quality | Medium Quality | Low Quality |
|---------|-------------|---------------|-------------|
| Wall rendering | Normal maps + specular + wear | Albedo + simple lighting | Flat color + fog |
| Barriers | Full shaped sprites + collision | Simple billboard sprites | Colored rectangles |
| Partial barriers | Per-pixel alpha + gap patterns | Uniform transparency | Opaque or invisible |
| Sky | Full gradient + celestial bodies + stars | Gradient + sun/moon | Solid color |
| Materials | Per-material textures + normals | Shared textures per type | Genre palette colors |
| Interaction highlight | Glow outline with pulse | Simple color tint | Text indicator only |
| Particles | 2000 max, full behavior | 500 max, simplified | 100 max, point particles |
| Object physics | Full push/swing simulation | Simplified movement | No physics |

### Automatic Quality Detection

On startup, run a quick benchmark (render 10 frames, measure average time):
- If avg frame time > 14ms (below 60 FPS): downgrade to medium.
- If avg frame time > 20ms (below 50 FPS): downgrade to low.
- User can override via `config.yaml` → `rendering.quality: "high"`.

### Graceful Degradation During Play

If frame time exceeds 18ms for 10 consecutive frames:
1. Reduce particle count by 50%.
2. If still slow: disable normal maps.
3. If still slow: reduce barrier detail to level 1.
4. If still slow: reduce barrier detail to level 0.

Recovery: if frame time drops below 12ms for 30 consecutive frames, restore one quality tier.

### Accessibility Fallbacks

| Feature | Standard | High Contrast | Colorblind Mode |
|---------|----------|---------------|-----------------|
| Interaction highlight | Genre accent color glow | Bright white outline (3px) | Pattern overlay + white outline |
| Item identification | Color + silhouette | Silhouette + thick outline | Silhouette + symbol overlay |
| Barrier types | Color differentiation | Shape + brightness contrast | Shape + pattern fill |
| Sky weather | Color shift | Color shift + icon indicator | Desaturated + icon indicator |

These integrate with the existing `AccessibilityConfig` in `config/config.go` which already supports `HighContrast`, `ColorblindMode` (4 modes), `ReducedMotion`, and `LargeText`.
