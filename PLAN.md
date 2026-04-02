# Implementation Plan: Performance & UI Rendering Migration

## Project Context
- **What it does**: A 100% procedurally generated first-person open-world RPG built in Go on Ebitengine with zero external assets
- **Current goal**: Achieve consistent 60 FPS by migrating UI rendering from per-pixel `Set()` calls to batch rendering APIs
- **Estimated Scope**: Medium (27 `screen.Set()` call sites across 4 UI files)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|----------------|---------------------|
| 200 Features | ✅ Complete (200/200) | No |
| 60 FPS at 1280×720 | ⚠️ Partial — raycaster migrated, UI not | **Yes** |
| Zero external assets | ✅ Complete | No |
| Five genre themes | ✅ Complete | No |
| ECS architecture | ✅ Complete (58 systems) | No |
| Multiplayer (200-5000ms latency) | ⚠️ Protocol defined, sync incomplete | No |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 9 functions above threshold (complexity >10)
  - Top: `GenerateRoads` (24.1), `Draw` (16.6), `main` (16.1)
- **Duplication ratio**: 0.98% (654 lines) — ✅ excellent
- **Doc coverage**: 86.9% — ✅ above 80% target
- **Package coupling**: `adapters` (10.0), `main` (10.0) — high coupling in entrypoints as expected

## Research Findings

### Ebitengine v2.9 Performance Best Practices
From official documentation and community resources:
1. **Avoid `WritePixels` every frame** — reserve for static textures or rarely-updated data
2. **Maximize draw call batching** — use sprite atlases, keep transformations uniform
3. **Use `DrawImage()` with `ColorM`/`GeoM`** for hardware-accelerated UI overlays
4. **Use `Fill()` for background clearing** instead of per-pixel writes
5. **Avoid `At()` pixel accessors** — they pull data back from GPU, causing stalls

### Current UI Rendering Problem
The raycaster core successfully uses `WritePixels()` for framebuffer rendering, but 27 `screen.Set()` calls remain in UI code:
- `cmd/client/main.go`: 11 calls (minimap, crosshair, combat effects)
- `cmd/client/quest_ui.go`: 10 calls (panel backgrounds, borders)
- `cmd/client/inventory_ui.go`: 6 calls (grid, slots, weight bar)
- `cmd/client/dialog_ui.go`: 1 call (background)

Each `Set()` call triggers GPU pipeline synchronization. At 60 FPS with UI open, this creates thousands of unnecessary sync points per second.

---

## Implementation Steps

### Step 1: Create Shared UI Framebuffer Infrastructure ✅
- **Deliverable**: New `UIFramebuffer` struct in `cmd/client/ui_buffer.go` with pre-allocated `[]byte` buffer and `DrawRect()`, `DrawBorder()`, `BlendPixel()` helper methods
- **Dependencies**: None
- **Goal Impact**: Foundation for all subsequent UI migration steps
- **Acceptance**: `UIFramebuffer` can render filled rectangles and borders to `[]byte` buffer
- **Validation**: `go test -v ./cmd/client/... -run TestUIFramebuffer`
- **Status**: COMPLETED 2026-04-01 — Created `ui_buffer.go` with full UIFramebuffer API + tests

### Step 2: Migrate Minimap Rendering ✅
- **Deliverable**: Replace 5 `screen.Set()` calls in `drawMinimap()` (`cmd/client/main.go:1302-1381`) with `UIFramebuffer` writes
- **Dependencies**: Step 1
- **Goal Impact**: Reduces per-frame GPU syncs; minimap always visible during gameplay
- **Acceptance**: Minimap renders identically using framebuffer; zero `Set()` calls in `drawMinimap()`
- **Validation**: `grep -c 'screen.Set' cmd/client/main.go` shows 6 remaining (down from 11)
- **Status**: COMPLETED 2026-04-01 — Removed fallback path, now uses WritePixels exclusively (6 Set() calls removed)

### Step 3: Migrate Combat Effect Overlays ✅
- **Deliverable**: Replace `screen.Set()` calls for damage flash and screen shake in `cmd/client/main.go` with `DrawImage()` using `ColorM` for tinting
- **Dependencies**: Step 1
- **Goal Impact**: Hardware-accelerated damage feedback effects
- **Acceptance**: Combat effects use `DrawImage()` with color transforms; visual quality preserved
- **Validation**: `grep -c 'screen.Set' cmd/client/main.go` shows 1 remaining (crosshair)
- **Status**: COMPLETED 2026-04-01 — Speech bubble and bar fallbacks removed, now uses DrawImage exclusively

### Step 4: Migrate Crosshair Rendering ✅
- **Deliverable**: Replace 5 `screen.Set()` calls for crosshair (`cmd/client/main.go:1377-1381`) with pre-rendered `ebiten.Image` drawn via `DrawImage()`
- **Dependencies**: None (parallel to Steps 2-3)
- **Goal Impact**: Eliminates per-pixel crosshair rendering; single draw call per frame
- **Acceptance**: Crosshair renders identically; zero `Set()` calls for crosshair
- **Validation**: `grep -c 'screen.Set' cmd/client/main.go` shows 0 remaining
- **Status**: COMPLETED 2026-04-01 — Crosshair already used pre-rendered image; minimap was the remaining Set() source (now fixed in Step 2)

### Step 5: Migrate Quest UI Panel Backgrounds
- **Deliverable**: Replace 6 `screen.Set()` loops in `cmd/client/quest_ui.go:399-416` with `UIFramebuffer.DrawRect()` and upload via `WritePixels()` or use `ebiten.Image.Fill()` for solid colors
- **Dependencies**: Step 1
- **Goal Impact**: Quest panel opens without GPU sync stalls
- **Acceptance**: Quest panel backgrounds render identically; no `Set()` loops for fill operations
- **Validation**: `grep -c 'screen.Set' cmd/client/quest_ui.go` shows 4 remaining (borders only)

### Step 6: Migrate Quest UI Borders and Selection
- **Deliverable**: Replace remaining 4 `screen.Set()` calls in quest UI for borders and selection highlighting with `DrawImage()` using pre-rendered border images or `UIFramebuffer.DrawBorder()`
- **Dependencies**: Step 5
- **Goal Impact**: Complete quest UI migration
- **Acceptance**: Zero `Set()` calls in `quest_ui.go`
- **Validation**: `grep -c 'screen.Set' cmd/client/quest_ui.go` shows 0

### Step 7: Migrate Inventory UI Grid and Slots
- **Deliverable**: Replace 6 `screen.Set()` loops in `cmd/client/inventory_ui.go:362-526` with `UIFramebuffer` or `DrawImage()` compositing
- **Dependencies**: Step 1
- **Goal Impact**: Inventory screen opens without frame drops
- **Acceptance**: Inventory renders identically; zero `Set()` calls
- **Validation**: `grep -c 'screen.Set' cmd/client/inventory_ui.go` shows 0

### Step 8: Migrate Dialog UI Background
- **Deliverable**: Replace 1 `screen.Set()` loop in `cmd/client/dialog_ui.go:444` with `Fill()` or pre-rendered image overlay
- **Dependencies**: None (parallel to Steps 5-7)
- **Goal Impact**: Dialog opens without GPU sync
- **Acceptance**: Dialog background renders identically; zero `Set()` calls
- **Validation**: `grep -c 'screen.Set' cmd/client/dialog_ui.go` shows 0

### Step 9: Consolidate and Upload UIFramebuffer
- **Deliverable**: In `Draw()` method, composite all UI framebuffer writes and upload once via `WritePixels()` or `DrawImage()` to a UI layer image
- **Dependencies**: Steps 2-8
- **Goal Impact**: Single GPU upload for all UI elements
- **Acceptance**: All UI rendering uses single-upload pattern
- **Validation**: Profile frame time with all UI panels open; target <16ms total

### Step 10: Reduce High-Complexity Functions
- **Deliverable**: Refactor top 5 high-complexity functions:
  - `GenerateRoads` (24.1) → extract road segment helpers
  - `Draw` (16.6) → split into `drawWorld()`, `drawUI()`, `drawEffects()`
  - `main` server (16.1) → extract system registration to `initSystems()`
  - `runServerLoop` (15.8) → extract tick phases
  - `handleFactionToggle` (15.8) → table-driven logic
- **Dependencies**: Steps 1-9 complete (avoid conflicts with UI changes)
- **Goal Impact**: Improves maintainability; reduces bug risk in hot paths
- **Acceptance**: All functions have cyclomatic complexity ≤10
- **Validation**: `go-stats-generator analyze . --skip-tests --format json | jq '.functions[] | select(.complexity.overall > 10)' | wc -l` shows 0

### Step 11: Pre-allocate Per-Frame Buffers
- **Deliverable**: 
  - Move particle buffer from per-frame `make([]byte)` to struct field
  - Move post-processing `image.RGBA` buffers to struct fields
  - Move sprite sort slice to renderer struct with `[:0]` reuse
- **Dependencies**: None (parallel to UI migration)
- **Goal Impact**: Reduces GC pressure from ~1 GB/sec to near-zero per-frame allocations
- **Acceptance**: Zero `image.NewRGBA()` or `make([]byte)` calls in hot paths
- **Validation**: `go test -bench=. -benchmem ./cmd/client/... | grep 'allocs/op'` shows ≥80% reduction

### Step 12: Performance Validation
- **Deliverable**: Run full benchmark suite; document frame times with all UI panels active
- **Dependencies**: Steps 1-11
- **Goal Impact**: Verify 60 FPS target achievement
- **Acceptance**: Average frame time <16ms with quest+inventory+dialog panels open
- **Validation**: `go test -bench=BenchmarkDraw -benchtime=5s ./cmd/client/...` shows stable frame times

---

## Scope Assessment Rationale

| Metric | Value | Assessment |
|--------|-------|------------|
| `screen.Set()` call sites | 27 | Medium (5-15 threshold) |
| Files requiring changes | 4 | Low |
| High-complexity functions | 9 | Medium |
| Duplication ratio | 0.98% | ✅ No action needed |
| Doc coverage gap | 13.1% | ✅ Acceptable |

**Classification: Medium** — 27 call sites require migration, but changes are mechanical and isolated to UI rendering paths. Each step is independently testable with clear validation commands.

---

## Risk Mitigation

1. **Visual regression risk**: Each step includes validation that visual output matches before/after
2. **Performance regression risk**: Steps are ordered so baseline rendering (raycaster) is not modified
3. **Merge conflict risk**: High-complexity refactoring (Step 10) is scheduled after UI migration to avoid overlapping changes

---

## Timeline Estimate

| Phase | Steps | Duration |
|-------|-------|----------|
| Infrastructure | 1 | 0.5 days |
| Main.go migration | 2-4 | 1 day |
| Quest UI migration | 5-6 | 0.5 days |
| Inventory/Dialog migration | 7-8 | 0.5 days |
| Consolidation | 9 | 0.5 days |
| Complexity refactor | 10 | 1 day |
| Buffer pre-allocation | 11 | 0.5 days |
| Validation | 12 | 0.5 days |
| **Total** | **12** | **5 days** |

---

## Verification Commands

```bash
# Count remaining Set() calls (target: 0)
grep -c 'screen.Set' cmd/client/*.go

# Verify build success
go build ./cmd/client && go build ./cmd/server

# Run tests with race detection
go test -race ./cmd/client/...

# Check complexity after refactoring
go-stats-generator analyze . --skip-tests --format json --sections functions | \
  jq '[.functions[] | select(.complexity.overall > 10)] | length'

# Benchmark frame performance
go test -bench=BenchmarkDraw -benchmem ./cmd/client/...

# Profile memory allocations
go test -bench=. -memprofile=mem.prof ./cmd/client/...
go tool pprof -top mem.prof | head -20
```

---

*Generated 2026-04-01 from go-stats-generator metrics analysis and project documentation review.*
