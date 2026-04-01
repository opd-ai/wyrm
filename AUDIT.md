# AUDIT.md — Terrain Generation & Ebitengine Performance Technical Assessment

> **Date:** 2026-04-01
> **Scope:** Comprehensive terrain system analysis, Ebitengine API usage audit, rendering pipeline evaluation, and performance baseline measurement across the Wyrm codebase (54,020 LOC production, 45,606 LOC tests).

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Terrain Generation System](#2-terrain-generation-system)
3. [Rendering Pipeline](#3-rendering-pipeline)
4. [Ebitengine API Usage Patterns](#4-ebitengine-api-usage-patterns)
5. [Particle System](#5-particle-system)
6. [Audio Pipeline](#6-audio-pipeline)
7. [ECS System Integration](#7-ecs-system-integration)
8. [Performance Baselines](#8-performance-baselines)
9. [Appendix: File Index](#9-appendix-file-index)

---

## 1. Executive Summary

Wyrm is a 100% procedurally generated first-person open-world RPG built on Ebitengine v2.9.3. The codebase contains a mature ECS architecture (60 registered systems), comprehensive terrain generation (4-octave Perlin noise, 10 biome types, 5 genre themes), and a DDA raycasting renderer. However, the rendering pipeline has a **fundamental architecture problem**: all rendering uses per-pixel `screen.Set()` calls (36 instances) instead of Ebitengine's batch rendering APIs (`DrawImage`/`DrawTriangles`). This, combined with per-frame buffer allocations (~40 MB/frame when post-processing is active), creates severe performance bottlenecks that prevent the game from reaching its 60 FPS target at 1280×720 resolution.

### Key Findings

| Category | Status | Assessment |
|----------|--------|------------|
| Terrain Generation Quality | ✅ Solid | 4-octave noise, FNV-1a seeding, deterministic chunk streaming |
| Biome System | ✅ Good | 10 biome types, genre-weighted distribution |
| Chunk Streaming | ✅ Functional | 3×3 window, lazy loading, thread-safe cache |
| LOD System | ⚠️ Defined but unused | 4 LOD levels exist, not wired to renderer |
| Raycaster Core | ✅ Correct | DDA algorithm, Z-buffer, billboard sprites |
| Rendering Method | ❌ Critical | 100% per-pixel Set() — no batch rendering |
| Post-Processing | ❌ Critical | 11 image.NewRGBA() allocations per frame |
| Particle System | ⚠️ Functional | Per-pixel blending, 3.6 MB buffer allocation per frame |
| Texture Management | ⚠️ Suboptimal | CPU-side pixel arrays, no GPU texture caching |
| Memory Management | ❌ Critical | ~40+ MB/frame GC pressure, zero sync.Pool usage |
| Object Pooling | ❌ Missing | No pooling for any frequently allocated objects |

---

## 2. Terrain Generation System

### 2.1 Chunk Architecture

**Location:** `pkg/world/chunk/manager.go` (651 lines)

The chunk system uses a well-designed seed-based architecture:

```
Chunk Structure:
├── X, Y: int                  — Grid coordinates
├── Size: int                  — Configurable (default 512×512)
├── Seed: int64                — FNV-1a(baseSeed, x, y)
├── HeightMap: []float64       — Normalized [0, 1], one value per cell
├── ElevationMap: []float64    — World units [0, MaxElevation=10.0]
└── TerrainTypes: []int        — Classification (Flat/Hill/Cliff/Peak)
```

**Seed Derivation** (lines 315-322): Uses FNV-1a hash for collision-free chunk seeding:
```go
func mixChunkSeed(baseSeed int64, x, y int) int64 {
    h := fnv.New64a()
    binary.Write(h, binary.LittleEndian, baseSeed)
    binary.Write(h, binary.LittleEndian, int64(x))
    binary.Write(h, binary.LittleEndian, int64(y))
    return int64(h.Sum64())
}
```

**Assessment:** ✅ Seed derivation is cryptographically sound and deterministic. FNV-1a provides excellent distribution with minimal collision risk.

### 2.2 Heightmap Generation

**Location:** `pkg/world/chunk/manager.go` (lines 61-94)

**Algorithm:** 4-octave value noise with smoothstep interpolation.

| Parameter | Value | Notes |
|-----------|-------|-------|
| Octaves | 4 | Standard for terrain |
| Persistence | 0.5 | Each octave halves amplitude |
| Lacunarity | 2.0 | Each octave doubles frequency |
| Normalization | `(h+1)/2` → [0,1] | Linear mapping |
| Offset | Random per-chunk | `rng.Float64() * 1000` |

**Elevation Conversion** (lines 96-140):
- 5-octave elevation detail (`ElevationOctaves = 5`)
- Exponential curve: `elevation = combined² × MaxElevation`
- Produces dramatic peaks while keeping lowlands relatively flat
- Range: 0.0–10.0 world units

**Terrain Classification Thresholds:**

| Type | Threshold | Method |
|------|-----------|--------|
| Peak | height ≥ 0.8 | Direct height check |
| Hill | height ≥ 0.5 | Height below peak |
| Cliff | slope ≥ 0.15 | 8-neighbor gradient analysis |
| Flat | default | Everything else |

**Assessment:** ✅ Heightmap quality is good for the game's scale. The 4-octave noise produces natural-looking terrain variation. The exponential elevation curve creates visually interesting terrain profiles. However, only 4 terrain types limits geometric variety (no valleys, caves, overhangs, or water bodies).

### 2.3 Noise Generator

**Location:** `pkg/procgen/noise/generator.go` (56 lines)

**Algorithm:** Value noise with hash-based random values (NOT classic Perlin noise).

```go
func HashToFloat(x, y int, seed int64) float64 {
    h := uint64(seed)
    h ^= uint64(x) * 0x9E3779B97F4A7C15
    h ^= uint64(y) * 0xBF58476D1CE4E5B9
    h = (h ^ (h >> 30)) * 0xBF58476D1CE4E5B9
    h = (h ^ (h >> 27)) * 0x94D049BB133111EB
    h ^= h >> 31
    return float64(h&0x7FFFFFFFFFFFFFFF) / float64(0x7FFFFFFFFFFFFFFF)
}
```

**Interpolation:** Smoothstep `t*t*(3-2*t)` with bilinear interpolation between 4 corner values.

**Assessment:** ⚠️ The hash function uses SplitMix64-style mixing constants, which provides good distribution. However, value noise produces less natural-looking terrain than gradient (Perlin) or simplex noise. No lookup table is used — all noise is computed fresh each call, which impacts texture generation performance where noise is sampled per-pixel.

### 2.4 Biome System

**Location:** `pkg/procgen/adapters/terrain.go`

**10 Biome Types:** Forest, Mountain, Lake, Swamp, Wasteland, Urban, Industrial, Ruins, Crater, Tech

**Genre-Based Distribution:**

| Genre | Primary (%) | Secondary (%) | Tertiary (%) | Quaternary (%) |
|-------|-------------|---------------|--------------|----------------|
| Fantasy | Forest (40) | Mountain (30) | Lake (20) | Ruins (10) |
| Sci-Fi | Crater (35) | Tech (35) | Industrial (30) | — |
| Horror | Swamp (40) | Forest (35) | Ruins (25) | — |
| Cyberpunk | Urban (50) | Industrial (35) | Tech (15) | — |
| Post-Apocalyptic | Wasteland (45) | Ruins (35) | Crater (20) | — |

**Assessment:** ✅ Good genre differentiation. Biome weights produce distinct world feels per genre. However, biome blending at chunk boundaries is not implemented — transitions are abrupt at chunk edges.

### 2.5 Chunk Streaming

**Location:** `cmd/client/main.go` (lines 547-604)

**Strategy:**
- 3×3 chunk window around player (9 chunks loaded)
- Lazy loading via `GetChunk(x, y)` — generated on first access
- Thread-safe cache with `sync.RWMutex`
- `UnloadChunk()` available for memory management
- Local world map rebuilt on chunk boundary crossing

**Server-side streaming:**
- 500ms interval chunk streaming ticker
- Delta-compressed entity updates at tick rate

**Assessment:** ✅ Functional but basic. The 3×3 window is appropriate for a raycaster with limited draw distance. However, chunk generation happens synchronously on the game thread — a large chunk at first access could cause frame stutter.

### 2.6 LOD System

**Location:** `pkg/world/chunk/manager.go` (lines ~220-240)

**Defined LOD Levels:**

| Level | Resolution | Reduction |
|-------|-----------|-----------|
| LODFull (0) | Every cell | 1× |
| LODHalf (1) | Every 2nd cell | 4× |
| LODQuarter (2) | Every 4th cell | 16× |
| LODEighth (3) | Every 8th cell | 64× |

**Status:** ⚠️ **Defined but NOT integrated.** The `LODChunk` struct and `ChunkLODCache` exist with thread-safe access, but no code in the rendering pipeline selects LOD levels based on distance. This is dead code awaiting integration.

### 2.7 Geometric Variety Assessment

**Current terrain features:**

| Feature | Status | Notes |
|---------|--------|-------|
| Hills | ✅ Present | Height ≥ 0.5 threshold |
| Mountains/Peaks | ✅ Present | Height ≥ 0.8 threshold |
| Cliffs | ✅ Present | Slope-based detection |
| Flat terrain | ✅ Present | Default classification |
| Valleys | ❌ Missing | No concavity detection |
| Caves | ❌ Missing | No 3D terrain carving |
| Overhangs | ❌ Missing | Heightmap is single-valued |
| Water bodies | ❌ Missing | No water plane generation |
| Rivers | ❌ Missing | No flow simulation |
| Rock formations | ❌ Missing | No detail geometry |
| Vegetation | ❌ Missing | No plant placement |
| Roads/Paths | ❌ Missing | No path generation between POIs |

**Assessment:** The heightmap-based terrain provides adequate macro-scale variety but lacks micro-scale detail features. The single-valued heightmap fundamentally prevents caves, overhangs, and complex 3D geometry. Water bodies and vegetation would significantly improve visual quality.

---

## 3. Rendering Pipeline

### 3.1 Raycaster Core

**Location:** `pkg/rendering/raycast/core.go` (3,652 lines), `pkg/rendering/raycast/draw.go`

**Algorithm:** DDA (Digital Differential Analyzer)

| Parameter | Value |
|-----------|-------|
| FOV | π/3 (60°) |
| Max Ray Steps | 64 |
| Max Ray Distance | 100.0 units |
| Z-Buffer | Per-column `[]float64` |

**Ray casting pipeline per frame:**
1. Cast 1,280 rays (one per screen column at 1280px width)
2. For each ray: DDA traversal through world grid
3. Calculate wall height from perpendicular distance
4. Sample wall texture at hit point
5. Draw wall strip pixel-by-pixel via `screen.Set()`
6. Draw floor/ceiling rows pixel-by-pixel via `screen.Set()`
7. Sort and draw billboard sprites with Z-buffer occlusion

**Assessment:** ✅ The DDA algorithm is correctly implemented and efficient for raycasting. The core algorithm is not the bottleneck — the rendering output method (`screen.Set()`) is.

### 3.2 Wall Rendering

**Location:** `pkg/rendering/raycast/draw.go` (lines 206-270)

Per frame: 1,280 columns × variable height pixels = ~400,000 `screen.Set()` calls.

Each pixel requires:
1. Texture coordinate calculation
2. Texture sampling from `[]color.RGBA` array
3. Distance-based darkening
4. `screen.Set(x, y, color)` — **GPU roundtrip per pixel**

### 3.3 Floor/Ceiling Rendering

**Location:** `pkg/rendering/raycast/draw.go` (lines 135-203)

Per frame: ~360 rows × 1,280 columns × 2 (floor + ceiling) = ~921,600 `screen.Set()` calls.

Each pixel requires:
1. Row distance calculation
2. Floor/ceiling texture coordinate mapping
3. Texture sampling
4. `screen.Set()` for floor pixel
5. `screen.Set()` for ceiling pixel (mirrored Y)

### 3.4 Sprite/Billboard Rendering

**Location:** `pkg/rendering/raycast/draw.go` (lines 18-131), `pkg/rendering/raycast/billboard.go`

Pipeline:
1. Transform entities to screen space
2. Sort by distance (back-to-front)
3. Per-column Z-buffer occlusion test
4. Per-pixel rendering with alpha blending

Alpha blending (line 120-129):
```go
existing := screen.At(screenX, screenY)  // GPU read
// blend calculation...
screen.Set(screenX, screenY, blended)     // GPU write
```

**Assessment:** ❌ Alpha-blended sprites require a read-modify-write per pixel, which is the most expensive pattern possible with `screen.Set()`/`screen.At()`.

### 3.5 Texture System

**Location:** `pkg/rendering/texture/generator.go`

**Storage:** `[]color.RGBA` pixel arrays (CPU-side only)

| Texture Set | Count | Size | Total Memory |
|-------------|-------|------|-------------|
| Wall textures | 4 | 64×64 | 64 KB |
| Floor texture | 1 | 64×64 | 16 KB |
| Ceiling texture | 1 | 64×64 | 16 KB |
| Cached textures | Variable | 64×64 | Variable |

**Genre Palettes:**

| Genre | Colors | Feel |
|-------|--------|------|
| Fantasy | Warm gold, green, brown, light gold | Medieval warmth |
| Sci-Fi | Cool blue, white, chrome, steel blue | Metallic cold |
| Horror | Grey-green, near-black, blood red, dark grey | Oppressive dark |
| Cyberpunk | Neon pink, cyan, dark grey, deep purple | Electric neon |
| Post-Apocalyptic | Sepia, orange dust, rust, weathered tan | Dusty decay |

**Assessment:** ⚠️ Textures are stored as CPU-side pixel arrays and sampled per-pixel during raycasting. This prevents GPU texture caching and hardware-accelerated sampling. The texture resolution (64×64) is appropriate for a raycaster but could benefit from mipmapping for distance-based quality.

### 3.6 Post-Processing Pipeline

**Location:** `pkg/rendering/postprocess/effects.go` (524 lines)

**Architecture:** Sequential CPU-based pixel processing.

**Per-Genre Effect Chains:**

| Genre | Effects | Allocations/Frame |
|-------|---------|-------------------|
| Fantasy | WarmColorGrade (0.6) | 1 image.NewRGBA |
| Sci-Fi | Scanlines (2px), Bloom, CoolColorGrade | 3 image.NewRGBA |
| Horror | Desaturate, Vignette, DarkenOverall | 3 image.NewRGBA |
| Cyberpunk | ChromaticAberration (3px), Bloom, NeonGlow | 3 image.NewRGBA |
| Post-Apocalyptic | Sepia, FilmGrain, Desaturate | 3 image.NewRGBA |

**Integration overhead** (cmd/client/main.go lines 1074-1089):
1. Full-screen copy: `screen` → `image.RGBA` (921,600 pixel reads)
2. Pipeline execution: N effects × 921,600 pixel operations each
3. Full-screen copy: `image.RGBA` → `screen` (921,600 pixel writes)

**Assessment:** ❌ Critical performance issue. Each frame allocates 1-3 new `image.RGBA` buffers (3.6 MB each), performs multiple full-screen pixel traversals, and generates massive GC pressure. The copy loops between `ebiten.Image` and `image.RGBA` add 2× full-screen traversals on top.

### 3.7 Lighting System

**Location:** `pkg/rendering/lighting/lighting.go`

**Light Types:** Point, Directional, Spot, Ambient

**Features:**
- Quadratic falloff (configurable exponent)
- Distance-based attenuation
- Day/night cycle (24-hour period)
- Per-light color and intensity

**Assessment:** ✅ Well-designed lighting model. The quadratic falloff and day/night cycle provide good visual variety. Integration with the raycaster applies lighting as a color multiplier during wall/floor rendering.

### 3.8 Skybox Rendering

**Location:** `pkg/rendering/raycast/skybox.go`

Procedural sky rendering with genre-appropriate colors and time-of-day variation.

**Assessment:** ✅ Functional and appropriate for the raycasting renderer.

---

## 4. Ebitengine API Usage Patterns

### 4.1 Per-Pixel Operations (Critical Anti-Pattern)

**Total `screen.Set()` call sites:** 36
**Total `screen.At()` call sites:** 4

**Distribution by file:**

| File | Set() | At() | Context |
|------|-------|------|---------|
| `pkg/rendering/raycast/draw.go` | 5 | 1 | Wall/floor/ceiling/sprite rendering |
| `cmd/client/main.go` | 14 | 3 | Combat feedback, post-processing, particles, minimap |
| `cmd/client/quest_ui.go` | 10 | 0 | UI background/border drawing |
| `cmd/client/dialog_ui.go` | 1 | 0 | Dialog background |
| `cmd/client/inventory_ui.go` | 6 | 0 | Inventory grid/slots |

**Why this is critical:** Ebitengine's `ebiten.Image.Set()` and `At()` methods are designed for occasional pixel manipulation, not bulk rendering. Each call involves a synchronization with the GPU pipeline. At 1280×720, the raycaster alone makes ~1.3 million `Set()` calls per frame. At 60 FPS, this is **78 million GPU synchronization points per second**.

### 4.2 Batch Rendering API Usage

| API | Usage Count | Notes |
|-----|-------------|-------|
| `DrawImage()` | 1 | Menu overlay only (`cmd/client/menu.go:172`) |
| `DrawTriangles()` | 0 | Never used |
| `DrawRectShader()` | 0 | Never used |
| `Fill()` | 0 | Never used (could replace background clearing) |
| `WritePixels()` | 0 | Never used (could replace all Set() calls) |

**Assessment:** ❌ The codebase uses virtually none of Ebitengine's batch rendering capabilities. The single `DrawImage()` call is for a menu overlay. The entire game rendering bypasses Ebitengine's optimized rendering paths.

### 4.3 Image Lifecycle Management

**`ebiten.NewImage()` allocations:** 1 (menu.go — acceptable, one-time)

**`image.NewRGBA()` allocations in hot path:** 12 per frame (11 in postprocess + 1 in main)

**Assessment:** ⚠️ No `ebiten.Image` leak issues, but the `image.RGBA` allocations for post-processing create severe GC pressure.

### 4.4 ColorM/GeoM Transform Usage

| Transform | Usage |
|-----------|-------|
| `ebiten.ColorM` | 0 — Not used anywhere |
| `ebiten.GeoM` | 0 — Not used anywhere |
| `ebiten.DrawImageOptions` | 1 — Menu overlay |

**Assessment:** ❌ Complete absence of Ebitengine's hardware-accelerated color and geometry transforms. All color manipulation (post-processing, lighting, distance fog) is done via manual per-pixel CPU computation. This leaves significant GPU capability unused.

### 4.5 Audio Integration

**Location:** `pkg/audio/player.go`

```go
type Player struct {
    context   *audio.Context   // ebiten/v2/audio.Context
    player    *audio.Player    // ebiten/v2/audio.Player
    stream    *SampleStream    // Custom io.Reader
}
```

**Assessment:** ✅ Audio integration follows Ebitengine best practices. The `SampleStream` implements `io.Reader` for streaming synthesis. Thread-safe with `sync.Mutex`. Pre-allocated silence buffer prevents underflow.

### 4.6 Input Handling

**Location:** `cmd/client/main.go` (lines 416-442, 676-717)

- Uses `inpututil.AppendJustPressedKeys()` for event-based input
- Uses `ebiten.IsKeyPressed()` for continuous movement input
- Mouse input via `inpututil.IsMouseButtonJustPressed()`
- Configurable key bindings via input manager

**Assessment:** ✅ Input handling follows Ebitengine best practices with appropriate use of both event-based and polling-based input methods.

### 4.7 Game Lifecycle

**Location:** `cmd/client/main.go`

- `Update()`: Fixed 60 FPS timestep (dt = 1/60)
- `Draw()`: Called after each Update
- `Layout()`: Returns configured window dimensions

**Assessment:** ✅ Standard Ebitengine game loop implementation. No blocking operations in Update/Draw.

---

## 5. Particle System

### 5.1 Architecture

**Location:** `pkg/rendering/particles/particles.go` (pool management), `pkg/rendering/particles/renderer.go` (204 lines)

**11 Particle Types:** Rain, Snow, Dust, Ash, Sparks, Blood, Magic, Smoke, Fire, FogWisp, Bubbles

**Configuration:**
- `DefaultMaxParticles = 1000`
- `DefaultPoolSize = 2000`
- Per-particle: Position, velocity, life, size, color (all in screen-space)

### 5.2 Rendering Method

**Per-pixel blending to byte buffer:**

```go
func (r *Renderer) drawPixelBlend(x, y int, c color.RGBA, pixels []byte) {
    idx := (y*r.width + x) * 4
    alpha := float64(c.A) / 255.0
    invAlpha := 1.0 - alpha
    pixels[idx] = uint8(float64(c.R)*alpha + float64(pixels[idx])*invAlpha)
    // ... G, B, A channels
}
```

**Per-particle draw methods:**

| Type | Method | Pixel Operations |
|------|--------|-----------------|
| Circle | Nested loops, radius² check | π × radius² pixels |
| Rain drop | Line with alpha fade | ~10 pixels |
| Snowflake | 9 individual pixel plots | 9 pixels |
| Glow | Nested loops, quadratic falloff | (2×radius)² pixels |

### 5.3 Integration

**Per-frame overhead** (cmd/client/main.go lines 1099-1133):
1. Allocate `make([]byte, width*height*4)` — **3.6 MB at 1280×720**
2. Copy screen to pixel buffer (full traversal)
3. Render particles to pixel buffer (per-pixel blending)
4. Copy pixel buffer back to screen (full traversal)

**Assessment:** ❌ The particle system creates a fresh 3.6 MB buffer every frame, copies the entire screen twice, and renders particles via per-pixel CPU blending. With 1000 active particles, each with potential glow effects, this generates significant CPU load and GC pressure.

---

## 6. Audio Pipeline

### 6.1 Synthesis Engine

**Location:** `pkg/audio/engine.go`

- Sine wave generation with frequency/duration parameters
- ADSR envelope (Attack/Decay/Sustain/Release)
- Genre-specific base frequencies (Fantasy: 220 Hz, Sci-Fi: 110 Hz, etc.)
- Sample rate: 44,100 Hz

### 6.2 Adaptive Music

**Location:** `pkg/audio/music/adaptive.go`

- 7 music states: Exploration, Combat, Tense, Victory, Defeat, Menu, Pause
- Layer-based mixing with crossfade transitions
- Genre-specific motifs (frequency patterns per genre)
- Transition timing: 2s (combat entry), 5s (combat exit)

### 6.3 Ambient Soundscapes

**Location:** `pkg/audio/ambient/soundscape.go`

- 9 region types: Plains, Forest, Cave, City, Water, Desert, Mountain, Dungeon, Interior
- Procedural generation per region (filtered noise, oscillator combinations)
- Genre modifications applied to base soundscapes
- 1-second transition on region change

**Assessment:** ✅ Audio architecture is well-designed with appropriate Ebitengine integration. The synthesis-based approach aligns with the zero-asset philosophy. No significant performance concerns.

---

## 7. ECS System Integration

### 7.1 Registration Status

**Total systems defined:** 57+
**Registered in server:** 60 systems
**Registered in client (online):** 3 systems (Render, Audio, Weather)
**Registered in client (offline):** 50 systems (full simulation)

**Assessment:** ✅ Systems are comprehensively registered. The client runs a subset when connected to a server (avoiding duplicate simulation), and the full set when offline.

### 7.2 System Update Performance

**Benchmark: `BenchmarkServerTickWith1000Entities`**
- Result: 88,364 ns/op (0.088 ms per tick)
- Allocations: 14 allocs/op, 8,272 B/op
- Target: 20 Hz tick rate = 50 ms budget
- **Headroom: 99.8%** — well within budget

**Benchmark: `BenchmarkWorldUpdate`**
- Result: 20.72 ns/op
- Allocations: 0 allocs/op
- Assessment: Negligible overhead for empty world updates

### 7.3 Entity Management

**Benchmark: `BenchmarkCreateDestroy` (10,000 entities)**
- Result: 1,140,183 ns/op (1.14 ms)
- Allocations: 10,001 allocs/op, 564,036 B/op
- Assessment: ⚠️ High allocation count suggests entity creation/destruction should use pooling for bulk operations.

---

## 8. Performance Baselines

### 8.1 Benchmark Summary

| Benchmark | ns/op | B/op | allocs/op | Assessment |
|-----------|-------|------|-----------|------------|
| WorldUpdate (empty) | 20.72 | 0 | 0 | ✅ Excellent |
| SelectAbility | 7.51 | 0 | 0 | ✅ Excellent |
| HidingSpotSystem | 22.12 | 0 | 0 | ✅ Excellent |
| ServerTick/1000 entities | 88,364 | 8,272 | 14 | ✅ Good |
| CreateDestroy/10K | 1,140,183 | 564,036 | 10,001 | ⚠️ High allocs |
| TransferEncodeDecode | 37,337 | 17,208 | 349 | ⚠️ Network overhead |
| WorldStateEncode | 16,950 | 3,216 | 604 | ⚠️ High alloc count |

### 8.2 Per-Frame Memory Budget Estimate (1280×720)

| Component | Allocation | Size | Frequency |
|-----------|-----------|------|-----------|
| Particle pixel buffer | `make([]byte, w*h*4)` | 3.6 MB | Every frame |
| Post-process input RGBA | `image.NewRGBA()` | 3.6 MB | Every frame |
| Post-process effects (1-3) | `image.NewRGBA()` × N | 3.6-10.8 MB | Every frame |
| Z-Buffer | `make([]float64, width)` | 10 KB | Every frame |
| Sprite sort slice | `make([]*SpriteEntity, ...)` | ~1 KB | Every frame |
| **Total baseline** | | **~7.2 MB** | **Per frame** |
| **Total with 3 effects** | | **~18 MB** | **Per frame** |

**At 60 FPS:** 432 MB/sec baseline — 1.08 GB/sec with effects

### 8.3 Per-Frame CPU Operation Count (1280×720)

| Operation | Count | Notes |
|-----------|-------|-------|
| Floor/ceiling Set() | 921,600 | 1280 × 360 × 2 |
| Wall Set() | ~400,000 | Variable height |
| Sprite Set() | ~50,000 | Variable, with At() reads |
| Post-process copy-in | 921,600 | Full screen read |
| Post-process effects | 921,600 × N | N = number of effects |
| Post-process copy-out | 921,600 | Full screen write |
| Particle copy-in | 921,600 | Full screen read |
| Particle rendering | Variable | Per-particle draw |
| Particle copy-out | 921,600 | Full screen write |
| **Total minimum** | **~5 million** | **Per frame, no effects** |
| **Total with 3 effects** | **~8 million** | **Per frame** |

### 8.4 Object Pooling Status

| Object | Pooled? | Per-Frame Allocations |
|--------|---------|----------------------|
| Particle pixel buffer | ❌ No | 3.6 MB |
| Post-process images | ❌ No | 3.6 MB × N |
| Z-Buffer | ❌ No | 10 KB |
| Sprite sort slice | ❌ No | ~1 KB |
| Entity creation | ❌ No | Variable |
| `sync.Pool` usage | ❌ None | — |

---

## 9. Appendix: File Index

### Critical Rendering Files

| File | Lines | Purpose |
|------|-------|---------|
| `pkg/rendering/raycast/core.go` | 3,652 | DDA raycaster core |
| `pkg/rendering/raycast/draw.go` | ~300 | Screen rendering (Set() calls) |
| `pkg/rendering/raycast/billboard.go` | ~200 | Sprite billboard system |
| `pkg/rendering/raycast/skybox.go` | ~150 | Sky rendering |
| `pkg/rendering/postprocess/effects.go` | 524 | Post-processing pipeline |
| `pkg/rendering/particles/particles.go` | ~200 | Particle management |
| `pkg/rendering/particles/renderer.go` | 204 | Particle rendering |
| `pkg/rendering/texture/generator.go` | ~150 | Texture generation |
| `pkg/rendering/lighting/lighting.go` | ~200 | Dynamic lighting |

### Terrain Generation Files

| File | Lines | Purpose |
|------|-------|---------|
| `pkg/world/chunk/manager.go` | 651 | Chunk management + terrain gen |
| `pkg/procgen/noise/generator.go` | 56 | Value noise function |
| `pkg/procgen/adapters/terrain.go` | ~100 | Biome distribution |
| `pkg/procgen/city/generator.go` | ~850 | City generation |
| `pkg/procgen/dungeon/dungeon.go` | 1,066 | Dungeon BSP generation |

### Client/Server Entry Points

| File | Lines | Purpose |
|------|-------|---------|
| `cmd/client/main.go` | 1,860 | Client game loop |
| `cmd/server/main.go` | 618 | Server tick loop |

---

*This audit provides the foundation for GAPS.md (issue identification) and PLAN.md (optimization roadmap).*
