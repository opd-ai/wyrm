# PLAN.md — Terrain Quality & Ebitengine Performance Optimization Roadmap

> **Date:** 2026-04-01
> **Companion to:** AUDIT.md (technical assessment), GAPS.md (issue catalog)
> **Goal:** Achieve 60 FPS at 1280×720 with 5× terrain detail improvement and ≥80% memory allocation reduction.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Phase 1: Software Framebuffer Architecture](#2-phase-1-software-framebuffer-architecture-critical)
3. [Phase 2: Memory Allocation Elimination](#3-phase-2-memory-allocation-elimination-critical)
4. [Phase 3: Unified Rendering Pipeline](#4-phase-3-unified-rendering-pipeline)
5. [Phase 4: Terrain Generation Enhancement](#5-phase-4-terrain-generation-enhancement)
6. [Phase 5: GPU-Accelerated Effects](#6-phase-5-gpu-accelerated-effects)
7. [Phase 6: Advanced Terrain & Polish](#7-phase-6-advanced-terrain--polish)
8. [Validation Strategy](#8-validation-strategy)
9. [Risk Assessment](#9-risk-assessment)

---

## 1. Overview

### Current State (from AUDIT.md)

The Wyrm rendering pipeline has a fundamental architecture issue: **all output uses per-pixel `screen.Set()` calls** (~1.3M per frame at 1280×720), while Ebitengine's batch rendering APIs (`WritePixels`, `DrawImage`, `DrawTriangles`) are unused. Combined with per-frame buffer allocations (~40 MB/frame with post-processing), this makes 60 FPS unachievable.

The terrain generation system produces functional but visually monotonous terrain with only 4 terrain types, no biome blending, and an unused LOD system.

### Target State

| Metric | Current | Target | Improvement |
|--------|---------|--------|-------------|
| GPU sync calls/frame | ~1,300,000 | 1-5 | 1,000,000× |
| Per-frame allocations | 7-18 MB | <100 KB | ≥80% reduction |
| GC pressure (60 FPS) | 432 MB-1.08 GB/sec | <6 MB/sec | ≥99% reduction |
| Terrain types | 4 | 8+ | 2× |
| Terrain features | Hills, cliffs, peaks | + valleys, water, vegetation, roads | 5× detail |
| Biome transitions | Hard edge (abrupt) | 32-cell blended | Seamless |
| Chunk generation | Synchronous (stutter) | Async (no stutter) | Stutter-free |
| Frame rate (1280×720) | Below 60 FPS | Stable 60 FPS | Target met |

### Phase Dependencies

```
Phase 1 (Framebuffer)
  └── Phase 2 (Allocation pooling)
        └── Phase 3 (Unified pipeline)
              ├── Phase 5 (GPU effects)
              └── Phase 4 (Terrain enhancement)
                    └── Phase 6 (Advanced terrain)
```

Phases 1-3 are sequential (each depends on the prior). Phases 4 and 5 can proceed in parallel after Phase 3. Phase 6 depends on Phase 4.

---

## 2. Phase 1: Software Framebuffer Architecture (CRITICAL)

**Addresses:** GAPS 9.1 (per-pixel Set()), 9.5 (sprite read-modify-write)
**Estimated effort:** 3-4 hours
**Expected impact:** ≥10× frame time improvement

### Objective

Replace all `screen.Set()` rendering with a single `screen.WritePixels()` upload per frame. This is the highest-impact change in the entire plan.

### Tasks

#### 1.1 Create Software Framebuffer

**File:** `pkg/rendering/raycast/core.go`

Add a persistent `[]byte` framebuffer to the `Renderer` struct, allocated once in the constructor:

```
Renderer struct {
    ...
    framebuffer []byte  // width × height × 4 (RGBA), allocated once
}
```

Provide helper methods:
- `SetPixel(x, y int, r, g, b, a uint8)` — write to framebuffer (bounds-checked)
- `BlendPixel(x, y int, r, g, b, a uint8)` — alpha-blend into framebuffer
- `Clear()` — zero the framebuffer (or use `copy` from a pre-zeroed slice)
- `Upload(screen *ebiten.Image)` — single `screen.WritePixels(r.framebuffer)` call

#### 1.2 Migrate Raycaster Rendering

**File:** `pkg/rendering/raycast/draw.go`

Replace the 5 `screen.Set()` call sites with framebuffer writes:

| Method | Current | Replacement |
|--------|---------|-------------|
| `renderFloorCeilingRow` | `screen.Set(x, y, floorColor)` | `r.SetPixel(x, y, ...)` |
| `renderFloorCeilingRow` | `screen.Set(x, ceilY, ceilColor)` | `r.SetPixel(x, ceilY, ...)` |
| `renderWallStrip` | `screen.Set(x, y, wallColor)` | `r.SetPixel(x, y, ...)` |
| `DrawSpritesToScreen` | `screen.Set(screenX, screenY, ...)` | `r.BlendPixel(screenX, screenY, ...)` |
| `DrawSpritesToScreen` | `screen.At()` + blend + `screen.Set()` | `r.BlendPixel(...)` (direct memory blend) |

#### 1.3 Migrate Client Rendering

**File:** `cmd/client/main.go`

Replace the 14 `screen.Set()` call sites with framebuffer writes. Pass the renderer's framebuffer reference to client rendering methods, or provide a client-side wrapper.

#### 1.4 Upload Framebuffer

In `Draw()`, after all rendering is complete, call `screen.WritePixels(renderer.framebuffer)` once.

### Validation

- [x] `go build ./cmd/client` succeeds
- [x] `go test -tags=noebiten ./pkg/rendering/...` passes
- [ ] Visual output is pixel-identical to current rendering
- [x] `grep -rn 'screen\.Set(' cmd/ pkg/rendering/` returns 0 results in hot path
- [ ] Frame time benchmark shows ≥10× improvement

---

## 3. Phase 2: Memory Allocation Elimination (CRITICAL)

**Addresses:** GAPS 9.2 (per-frame allocations), GAP-M1 (zero sync.Pool)
**Estimated effort:** 2-3 hours
**Expected impact:** ≥80% GC pressure reduction

### Objective

Eliminate all per-frame buffer allocations by pre-allocating persistent buffers and reusing them across frames.

### Tasks

#### 2.1 Pre-allocate Z-Buffer

**File:** `pkg/rendering/raycast/draw.go` (line 209)

Move `make([]float64, r.Width)` from `drawWalls()` into the `Renderer` constructor. Store as `Renderer.zBuffer` field. Clear at start of each frame with a `for` loop or `copy`.

#### 2.2 Pre-allocate Post-Processing Buffers

**File:** `pkg/rendering/postprocess/effects.go`

Add persistent buffer fields to `Pipeline` struct:
- `inputBuffer *image.RGBA` — reused across frames
- `workBuffer *image.RGBA` — reused across effects in chain

Replace all 11 `image.NewRGBA()` calls with buffer reuse. Each effect's `Apply()` reads from one buffer, writes to the other, and they swap roles.

#### 2.3 Pre-allocate Particle Buffer

**File:** `cmd/client/main.go` (line 1103) or `pkg/rendering/particles/renderer.go`

Move `make([]byte, width*height*4)` into the `particles.Renderer` constructor. Store as `Renderer.pixelBuffer` field. Clear at start of each frame.

#### 2.4 Pre-allocate Sprite Sort Slice

**File:** `pkg/rendering/raycast/draw.go` (line 27)

Move sprite sort slice allocation into `Renderer` struct. Reuse with `slice[:0]` pattern.

### Validation

- [x] `go test -tags=noebiten ./...` passes
- [ ] `go test -bench=. -benchmem pkg/rendering/...` shows ≥80% B/op reduction
- [x] No `make([]byte` or `image.NewRGBA()` in rendering hot paths (grep verification) — hot paths use pre-allocated buffers; allocating methods only exist for non-hot-path compatibility
- [ ] Frame time stability improved (reduced GC pause spikes)

---

## 4. Phase 3: Unified Rendering Pipeline

**Addresses:** GAPS 9.4 (post-process copy loops), 9.6 (particle copy loops)
**Estimated effort:** 2-3 hours
**Expected impact:** Eliminates ~3.6M redundant pixel operations per frame

### Objective

Unify raycaster, post-processing, and particle rendering into a single software framebuffer pipeline, eliminating all copy loops between `ebiten.Image` and `image.RGBA`/`[]byte`.

### Tasks

#### 3.1 Establish Rendering Order

Define the single-buffer rendering pipeline:

```
1. Clear framebuffer
2. Raycaster renders walls/floor/ceiling → framebuffer
3. Raycaster renders sprites → framebuffer (with alpha blend)
4. Post-processing applies effects → framebuffer (in-place)
5. Particle system renders → framebuffer (with alpha blend)
6. UI overlays render → framebuffer
7. Upload framebuffer → screen via WritePixels()
```

#### 3.2 Refactor Post-Processing to Operate on Framebuffer

**File:** `pkg/rendering/postprocess/effects.go`

Change `Apply(img *image.RGBA) *image.RGBA` to `Apply(pixels []byte, width, height int)` (in-place modification). Eliminate all `image.NewRGBA()` allocations inside effects.

#### 3.3 Refactor Particle Rendering to Use Shared Framebuffer

**File:** `pkg/rendering/particles/renderer.go`, `cmd/client/main.go`

Change particle renderer to accept the shared framebuffer `[]byte` directly instead of allocating its own. Eliminate the screen→buffer→screen copy loops in `main.go`.

#### 3.4 Migrate UI Rendering

**Files:** `cmd/client/quest_ui.go`, `cmd/client/inventory_ui.go`, `cmd/client/dialog_ui.go`

Replace per-pixel `screen.Set()` in UI code with either:
- Framebuffer writes (for simple rectangles/borders)
- `DrawImage()` overlays rendered after `WritePixels()` (for complex UI with text)

### Validation

- [x] `go build ./cmd/client` succeeds
- [x] `go test -tags=noebiten ./...` passes
- [x] No `screen.Set()` calls remain in any hot path
- [x] No `screen.At()` calls remain
- [x] No `image.NewRGBA()` allocations in rendering pipeline
- [ ] Post-processing visual effects are preserved
- [ ] Particle visual effects are preserved

---

## 5. Phase 4: Terrain Generation Enhancement

**Addresses:** GAPS 8.1 (limited variety), 8.2 (biome blending), 8.4 (value noise), 8.5 (sync chunk gen)
**Estimated effort:** 4-5 hours
**Expected impact:** 5× terrain detail improvement

### Tasks

#### 4.1 Add New Terrain Types

**File:** `pkg/world/chunk/manager.go`

Add terrain classifications:

| Type | Condition | Visual |
|------|-----------|--------|
| `TerrainValley` | height < 0.2 | Low-lying depression |
| `TerrainWater` | height < WaterLevel (0.15) | Water surface |
| `TerrainForest` | biome == Forest && 0.2 ≤ height < 0.5 | Tree-covered terrain |
| `TerrainRoad` | On path between POIs | Flat walkway |

Update terrain classification function to detect valleys (height < 0.2) and water bodies (height below configurable water plane elevation).

#### 4.2 Implement Biome Blending

**Files:** `pkg/world/chunk/manager.go`, `pkg/procgen/adapters/terrain.go`

When generating a chunk, query neighboring chunks' biomes within a 32-cell border zone. Interpolate:
- Heightmap parameters (amplitude, frequency scaling)
- Terrain color palette (linear interpolation between genre palettes)
- Vegetation density

Use smoothstep interpolation for the blend factor: `blend = smoothstep(distFromBorder / blendWidth)`.

#### 4.3 Implement Gradient Noise

**File:** `pkg/procgen/noise/generator.go`

Add `GradientNoise2D()` function alongside existing `Noise2D()`:
- Use per-grid-point gradient vectors (from hash function)
- Dot product of gradient with distance vector
- Smoothstep interpolation (existing `Smoothstep()` function)
- Same deterministic seeding as current value noise

Add a `NoiseType` parameter to terrain generator to select between value and gradient noise.

#### 4.4 Implement Async Chunk Generation

**File:** `pkg/world/chunk/manager.go`

Add background chunk generation:
- Worker goroutine with buffered channel work queue
- `GetChunk()` returns placeholder (flat, average-height chunk) for not-yet-generated chunks
- Callback or polling mechanism to swap in generated chunk
- Placeholder chunks use the chunk's biome color to avoid visual pops

#### 4.5 Vegetation and Rock Entity Spawning

**File:** `pkg/world/chunk/manager.go` or new `pkg/procgen/vegetation/`

During chunk generation, spawn detail entities based on biome and terrain type:

| Biome | Terrain Type | Entities |
|-------|-------------|----------|
| Forest | Flat/Hill | Trees (high density), bushes |
| Mountain | Hill/Cliff | Rocks, boulders |
| Swamp | Valley/Flat | Dead trees, fog wisps |
| Urban | Flat | Lamp posts, benches |
| Wasteland | Flat/Hill | Debris, scrap |

Entities are billboard sprites rendered by the raycaster's existing billboard system.

### Validation

- [x] `go test -tags=noebiten ./pkg/world/...` passes
- [x] `go test -tags=noebiten ./pkg/procgen/...` passes
- [x] Same seed produces identical terrain (determinism test)
- [ ] Visual inspection confirms: valleys, water, vegetation visible
- [ ] Chunk boundary transitions are smooth (no hard edges)
- [ ] No frame stutter when crossing chunk boundaries

---

## 6. Phase 5: GPU-Accelerated Effects

**Addresses:** GAP 9.3 (batch rendering unused)
**Estimated effort:** 2-3 hours
**Expected impact:** GPU-accelerated overlays and UI

### Tasks

#### 5.1 Use `DrawImage()` with `ColorM` for Screen Effects

After `WritePixels()` uploads the framebuffer, apply screen-wide effects using GPU:

| Effect | Implementation |
|--------|---------------|
| Vignette | Pre-generated gradient `ebiten.Image`, composited via `DrawImage()` with multiply blend |
| Color grade (warm/cool/sepia) | `DrawImage()` with `ColorM` scale+translate on R/G/B channels |
| Screen tint | `DrawImage()` with `ColorM` color multiply |
| Damage flash | Full-screen red overlay via `Fill()` + `DrawImage()` with alpha |

#### 5.2 Use `DrawImage()` for UI Elements

Replace per-pixel UI rendering with pre-generated UI element images:
- Pre-generate UI panel backgrounds as `ebiten.Image` (once on init)
- Position via `GeoM` translate
- Text rendering via `ebitenutil.DebugPrintAt()` or text/v2

#### 5.3 Use `Fill()` for Screen Clearing

Replace manual framebuffer zeroing for the sky/background with `screen.Fill(skyColor)` when the skybox is a solid color.

### Validation

- [ ] Post-processing effects visually match pre-optimization output
- [ ] UI elements render correctly at all window sizes
- [ ] GPU utilization visible in profiling (not 100% CPU rendering)
- [ ] Frame time further improved over Phase 3

---

## 7. Phase 6: Advanced Terrain & Polish

**Addresses:** GAP 8.3 (LOD integration)
**Estimated effort:** 3-4 hours
**Expected impact:** Reduced memory usage for distant terrain; improved visual detail

### Tasks

#### 6.1 Wire LOD to Chunk Streaming

**File:** `pkg/world/chunk/manager.go`

Select LOD level based on Manhattan distance from player's current chunk:

| Distance | LOD Level | Resolution |
|----------|-----------|-----------|
| 0 (current chunk) | LODFull | Every cell |
| 1 (adjacent) | LODHalf | Every 2nd cell |
| 2 | LODQuarter | Every 4th cell |
| 3+ | LODEighth | Every 8th cell |

Feed LOD-appropriate chunk data to the raycaster for world map construction.

#### 6.2 Texture Mipmapping

**File:** `pkg/rendering/texture/generator.go`

Generate 2-3 mipmap levels per texture during initialization:

| Level | Resolution | Use Case |
|-------|-----------|----------|
| 0 | 64×64 | Near walls (distance < 3) |
| 1 | 32×32 | Medium walls (3 ≤ distance < 8) |
| 2 | 16×16 | Far walls (distance ≥ 8) |

Select mipmap in `GetWallTextureColor()` based on ray distance.

#### 6.3 Road Generation Between POIs

**File:** New `pkg/procgen/roads/` or extend `pkg/world/chunk/manager.go`

Generate roads connecting city districts and dungeon entrances:
- A* pathfinding on heightmap (prefer flat terrain)
- Flatten terrain along road path
- Mark road cells as `TerrainRoad` type
- Render with distinct road texture

#### 6.4 Water Rendering

**File:** `pkg/rendering/raycast/draw.go`

Add water rendering for `TerrainWater` cells:
- Animated blue surface (sinusoidal UV offset over time)
- Reflective tint (darken floor texture, add blue overlay)
- Transparency effect (blend with terrain color below water level)

### Validation

- [ ] LOD reduces memory usage by measurable amount at 5×5 chunk window
- [ ] Mipmap selection reduces aliasing artifacts at distance
- [ ] Roads visually connect POIs and are walkable
- [ ] Water surfaces animate and look distinct from land
- [ ] All tests pass: `go test -tags=noebiten ./...`

---

## 8. Validation Strategy

### Per-Phase Validation

Every phase must pass before proceeding to the next:

1. **Build:** `go build ./cmd/client && go build ./cmd/server`
2. **Tests:** `go test -tags=noebiten -count=1 ./...`
3. **Race detection:** `go test -tags=noebiten -race ./...`
4. **Visual verification:** Manual inspection that rendering output matches expectations
5. **Performance benchmark:** `go test -tags=noebiten -bench=. -benchmem ./pkg/rendering/... ./pkg/world/...`

### Key Performance Metrics

| Metric | Baseline | Phase 1 Target | Phase 3 Target | Final Target |
|--------|----------|---------------|----------------|-------------|
| `screen.Set()` calls/frame | ~1,300,000 | 0 (hot path) | 0 (all) | 0 |
| `WritePixels()` calls/frame | 0 | 1 | 1 | 1 |
| Per-frame alloc (MB) | 7-18 | 7-18 | <0.1 | <0.1 |
| GC pressure (MB/sec @60fps) | 432-1080 | 432-1080 | <6 | <6 |
| Terrain types | 4 | 4 | 4 | 8+ |
| Biome blending | None | None | None | 32-cell smooth |

### Determinism Validation

After each terrain change, verify deterministic generation:

```bash
# Run 3 times with same seed, compare output
go test -tags=noebiten -run TestDeterminism -count=3 ./pkg/world/chunk/
```

---

## 9. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| `WritePixels()` API change in future Ebitengine | Low | High | Pin Ebitengine version; WritePixels is stable API |
| Post-processing visual regression | Medium | Medium | Pixel comparison tests before/after |
| Framebuffer architecture breaks existing tests | Low | Medium | Tests use `noebiten` tag; rendering tests are visual |
| Async chunk generation introduces race conditions | Medium | High | Use `sync.RWMutex`; test with `-race` flag |
| Gradient noise breaks terrain determinism | Low | High | Use same seed derivation; determinism tests |
| LOD introduces visual artifacts at boundaries | Medium | Low | Smooth LOD blending; visual QA |
| Memory pre-allocation increases baseline memory | Low | Low | Total pre-allocated: ~15 MB (well within 500 MB budget) |

---

## Appendix: Gap-to-Phase Mapping

| Gap ID | Description | Phase |
|--------|-------------|-------|
| 9.1 | Per-pixel Set() rendering | Phase 1 |
| 9.5 | Sprite read-modify-write | Phase 1 |
| 9.2 | Per-frame buffer allocations | Phase 2 |
| 9.4 | Post-process double copy loop | Phase 3 |
| 9.6 | Particle full-screen copy | Phase 3 |
| 8.1 | Limited terrain types | Phase 4 |
| 8.2 | No biome blending | Phase 4 |
| 8.4 | Value noise only | Phase 4 |
| 8.5 | Synchronous chunk generation | Phase 4 |
| 9.3 | Zero batch rendering API | Phase 5 |
| 8.3 | LOD not integrated | Phase 6 |
| 9.7 | No profiling infrastructure | Any phase (independent) |

---

*See AUDIT.md for the full technical assessment backing this plan, and GAPS.md for the complete issue catalog.*
