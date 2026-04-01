# Implementation Gaps — 2026-04-01

This document details the gaps between Wyrm's stated goals and current implementation.

---

## 1. High-Latency Network Support (200-5000ms)

- **Stated Goal**: README claims "200–5000 ms latency tolerance (designed for Tor-routed connections)"
- **Current State**: 
  - Normal mode: 500ms prediction window, adequate for 200ms RTT
  - Tor-mode activates at 800ms threshold, increases to 1500ms prediction window
  - No explicit support for RTT > 2000ms
  - `HistoryBufferSize` is 64 snapshots (insufficient for multi-second rewind)
  - `MaxRewindTime` capped at 500ms (`pkg/network/lagcomp.go:11`)
- **Impact**: Players on Tor or satellite connections (3000-5000ms RTT) will experience severe rubber-banding and failed hit registration. The game becomes effectively unplayable above ~2000ms RTT.
- **Closing the Gap**:
  1. Add additional RTT thresholds: 2000ms, 3000ms, 5000ms with proportional prediction windows
  2. Increase `HistoryBufferSize` from 64 to 256 entries
  3. Extend `MaxRewindTime` to 2000ms for extreme latency cases
  4. Add RTT-proportional input rate scaling (currently jumps from 60Hz to 10Hz)
  5. Implement aggressive entity culling for high-latency clients

**Files to modify**:
- `pkg/network/prediction.go:9-32` (constants and `adaptToLatency()`)
- `pkg/network/lagcomp.go:9-14` (buffer size and max rewind)
- `pkg/network/client.go` (input rate adjustment)

---

## 2. Delta Compression

- **Stated Goal**: README claims "delta compression" as a networking feature
- **Current State**:
  - `EntityUpdate` struct exists (`pkg/network/protocol.go:335-344`) for individual entity changes
  - All fields are always transmitted (no presence bitmask)
  - No bit-packing for floats (full 32-bit float for X, Y, Z, Angle)
  - No variable-length encoding for entity IDs
  - `WorldState` sends ALL entities each tick (`protocol.go:162-181`)
  - Minimum message size: ~40 bytes per entity update
- **Impact**: Bandwidth usage is ~2-4x higher than necessary. On constrained networks (mobile, Tor), this causes packet loss and increased latency variance.
- **Closing the Gap**:
  1. Add field presence bitmask to `EntityUpdate` (only send changed fields)
  2. Implement delta-from-baseline encoding for positions (send offset from last-known)
  3. Use variable-length integer encoding for entity IDs
  4. Add quantization for positions (e.g., fixed-point 16.16 instead of float32)
  5. Implement RLE or dictionary compression for state strings
  6. Send full state periodically (every N ticks), deltas otherwise

**Files to modify**:
- `pkg/network/protocol.go` (EntityUpdate struct and Encode/Decode)
- `pkg/network/server.go` (BroadcastEntityUpdate to track baselines)
- `pkg/network/client.go` (delta reconstruction)

---

## 3. UI Rendering Performance

- **Stated Goal**: README claims "60 FPS at 1280×720"
- **Current State**:
  - Raycaster correctly uses `WritePixels()` batch API (`pkg/rendering/raycast/draw.go:19`)
  - UI layer uses 27 `screen.Set()` per-pixel calls:
    - `cmd/client/main.go:1082,1091` — speech bubbles
    - `cmd/client/main.go:1302,1311` — UI elements
    - `cmd/client/main.go:1370,1377-1381` — crosshair (5 Set calls)
  - Each `Set()` triggers GPU pipeline synchronization
- **Impact**: Frame time spikes when UI panels are open (inventory, quest log, dialog). At 60 FPS, 27 Set calls = 1,620 GPU sync points/sec. Performance degrades on integrated graphics.
- **Closing the Gap**:
  1. Pre-allocate UI framebuffer (`[]byte`) in Game struct
  2. Render all UI to framebuffer using direct pixel writes
  3. Single `WritePixels()` call to upload UI layer
  4. Alternative: Use `Image.Fill()` for solid rectangles + `DrawImage()` with ColorM for tinted elements
  5. Consider sprite atlas for pre-rendered UI elements

**Files to modify**:
- `cmd/client/main.go` (Game struct, drawHUD, drawMinimap)
- `cmd/client/dialog_ui.go` (drawDialogBackground)
- `cmd/client/quest_ui.go` (drawQuestPanel)
- `cmd/client/inventory_ui.go` (drawInventoryGrid)

---

## 4. High-Complexity Functions

- **Stated Goal**: Maintainable codebase (implicit in professional software)
- **Current State**: 9 functions exceed cyclomatic complexity 10:

| Function | File | Lines | Complexity |
|----------|------|-------|------------|
| `GenerateRoads` | `pkg/procgen/city/generator.go` | 111 | 17 |
| `Draw` | `cmd/client/main.go` | 76 | 12 |
| `main` | `cmd/server/main.go` | 105 | 12 |
| `runServerLoop` | `cmd/server/main.go` | 61 | 11 |
| `handleFactionToggle` | `cmd/client/main.go` | 36 | 11 |
| `updateFurnitureMode` | `cmd/client/housing_ui.go` | 53 | 11 |
| `Update` (crafting) | `cmd/client/crafting_ui.go` | 45 | 11 |
| `updateSkillAllocation` | `cmd/client/character_ui.go` | 39 | 11 |
| `Encode` | `pkg/network/protocol.go` | 31 | 11 |

- **Impact**: High complexity correlates with bugs and difficulty in maintenance. These are critical code paths (rendering, networking, generation).
- **Closing the Gap**:

### GenerateRoads (complexity 17 → target ≤10)
Extract into:
- `generateMainRoads()` — district center connections
- `generateDistrictConnectors()` — secondary roads
- `generatePOIAccessRoads()` — POI connections
- `pruneOverlappingRoads()` — cleanup pass

### Draw (complexity 12 → target ≤10)
Split into:
- `drawWorld()` — raycaster + NPCs
- `drawEffects()` — particles, combat feedback
- `drawUI()` — HUD, panels

### main (server) (complexity 12 → target ≤10)
Extract:
- `initSystems()` — all RegisterSystem calls
- `initializeGameWorld()` — world population

### Encode (complexity 11 → target ≤10)
Replace switch with lookup table:
```go
var encoders = map[MessageType]func(*bytes.Buffer, Message) error{
    MsgConnect: encodeConnect,
    MsgWorldState: encodeWorldState,
    // ...
}
```

**Validation**: `go-stats-generator analyze . --skip-tests | grep "High Complexity"` shows 0 functions

---

## 5. TODO Annotations in Production Code

- **Stated Goal**: Production-ready codebase
- **Current State**: 4 TODO comments remain:
  - `cmd/client/dialog_ui.go:379` — "TODO: Get from player Skills component" (hardcoded SkillLevel: 50)
  - `cmd/client/dialog_ui.go:394` — "TODO: Get from player Skills component" (hardcoded SkillLevel: 30)
  - `cmd/client/main.go:290` — "TODO: Open container UI"
  - `cmd/client/main.go:293` — "TODO: Open/close door"
- **Impact**: Dialog skill checks use hardcoded values instead of actual player skills. Container and door interactions are not functional.
- **Closing the Gap**:
  1. Replace skill level TODO with actual component lookup:
     ```go
     skillComp, _ := g.world.GetComponent(g.playerEntity, "Skills")
     skills := skillComp.(*components.Skills)
     skillLevel := skills.Levels["Persuasion"]
     ```
  2. Implement container UI (similar to inventory UI)
  3. Implement door state toggling (modify door entity's DoorState component)

**Files to modify**:
- `cmd/client/dialog_ui.go:379,394`
- `cmd/client/main.go:290,293`

---

## 6. Naming Convention Violations

- **Stated Goal**: Follow Go naming conventions (implicit)
- **Current State**:
  - 19 file naming violations (stuttering, generic names)
  - 25 identifier stuttering violations
  - 1 package name violation (`pkg/util`)
- **Impact**: Reduced code discoverability and Go ecosystem conventions not followed.
- **Closing the Gap**:

### Package Rename
- `pkg/util` → `pkg/seedutil` or `pkg/mathutil` (based on actual contents)

### File Renames (examples)
- `server_init.go` → `init.go`
- `companion/companion.go` → `companion/behavior.go`
- `dialog/dialog.go` → `dialog/tree.go`
- `constants.go` files → merge into relevant type files or use `config.go`
- `types.go` → split by domain (e.g., `position.go`, `health.go`, `faction.go`)

### Identifier Fixes (examples)
- `DialogMemoryEvent` → `MemoryEvent` (in `pkg/engine/components`)
- `VehiclePhysics` → `Physics` (in vehicle context)
- `FactionMembership` → `Membership` (in faction context)

**Validation**: `go-stats-generator analyze . | grep "Violations"` returns 0 for all categories

---

## 7. Per-Frame Buffer Allocations

- **Stated Goal**: 60 FPS consistency, <500 MB RAM
- **Current State**:
  - Post-process buffers: `image.NewRGBA()` allocated per effect pass
  - Particle buffer: `make([]byte, w×h×4)` allocated per frame
  - Sprite sort slice: May allocate per frame depending on usage
- **Impact**: GC pressure causes periodic frame drops. Memory churn affects cache efficiency.
- **Closing the Gap**:
  1. Pre-allocate `image.RGBA` buffers in Pipeline struct (`pkg/rendering/postprocess/effects.go`)
  2. Pre-allocate particle pixel buffer in renderer struct
  3. Use `slice[:0]` reuse pattern for sorting (already done in raycast)
  4. Add `sync.Pool` for temporary buffers if allocation remains necessary

**Files to modify**:
- `pkg/rendering/postprocess/effects.go` (Pipeline struct)
- `pkg/rendering/particles/renderer.go`

**Validation**: `go test -bench=. -benchmem ./pkg/rendering/...` shows ≥80% allocation reduction

---

## 8. LOD System Not Wired to Renderer

- **Stated Goal**: Performance optimization for distant terrain
- **Current State**:
  - 4 LOD levels defined: `LODFull`, `LODHalf`, `LODQuarter`, `LODEighth` (`pkg/world/chunk/manager.go`)
  - `ChunkLODCache` struct exists
  - No rendering code selects LOD based on distance
- **Impact**: Memory usage higher than necessary; all chunks rendered at full detail regardless of distance.
- **Closing the Gap**:
  1. Calculate chunk distance from player in `ChunkManager.GetVisibleChunks()`
  2. Return appropriate LOD level based on distance thresholds
  3. Feed lower LOD data to raycaster for distant chunks
  4. Reduce texture resolution for lower LOD levels

**Files to modify**:
- `pkg/world/chunk/manager.go` (GetVisibleChunks, LOD selection)
- `pkg/rendering/raycast/core.go` (LOD-aware rendering)

**Validation**: Memory profiling shows reduced heap usage with LOD active

---

## 9. Runtime Profiling Infrastructure

- **Stated Goal**: Diagnosable production performance
- **Current State**:
  - No `net/http/pprof` import
  - No `runtime.MemStats` monitoring
  - No frame time tracking beyond Ebitengine's TPS
- **Impact**: Cannot diagnose performance issues without rebuilding with profiling code.
- **Closing the Gap**:
  1. Add `debug.profiling` config option (default: false)
  2. When enabled, start `net/http/pprof` endpoint on configurable port
  3. Add frame time tracking (last 60 frames) to debug overlay
  4. Add memory stats display (HeapAlloc, NumGC, GCPauseNs)
  5. Add entity count display

**Files to modify**:
- `config/config.go` (add Debug.Profiling, Debug.ProfilingPort)
- `cmd/client/main.go` (conditional pprof startup, debug overlay)

**Validation**: `curl http://localhost:6060/debug/pprof/` returns profile index when config enabled

---

## 10. Missing Packet Loss Handling

- **Stated Goal**: Robust networking for Tor/high-latency environments
- **Current State**:
  - TCP transport (reliable but high latency)
  - No explicit packet loss handling mentioned
  - No retransmission logic for UDP (if planned)
- **Impact**: On unreliable networks, packet loss causes state desync without recovery mechanism.
- **Closing the Gap**:
  1. Document that TCP provides reliability at cost of latency
  2. If UDP planned: implement selective ACK with retransmission
  3. Add connection quality metrics (packet loss %, jitter)
  4. Display connection quality indicator in HUD

**Files to modify**:
- `pkg/network/client.go` (connection quality tracking)
- `cmd/client/main.go` (HUD indicator)

---

## Gap Priority Matrix

| Gap | Severity | Effort | Impact on Stated Goals |
|-----|----------|--------|------------------------|
| 1. High-latency support | CRITICAL | High (2 weeks) | Directly contradicts README claim |
| 2. Delta compression | HIGH | Medium (1 week) | Partial implementation of claim |
| 3. UI rendering | HIGH | Medium (1 week) | Degrades 60 FPS claim |
| 4. High complexity | MEDIUM | Medium (1 week) | Maintenance risk |
| 5. TODO annotations | MEDIUM | Low (2-3 days) | Incomplete features |
| 6. Naming violations | LOW | Low (1-2 days) | Code quality |
| 7. Buffer allocations | MEDIUM | Low (3-4 days) | Performance consistency |
| 8. LOD system | LOW | Low (2-3 days) | Memory optimization |
| 9. Profiling | LOW | Low (1-2 days) | Diagnostics |
| 10. Packet loss | LOW | Low (1 day) | Documentation |

---

## Recommended Closing Order

1. **Week 1-2**: High-latency support (Gap 1) — addresses CRITICAL discrepancy
2. **Week 2-3**: Delta compression (Gap 2) + UI rendering (Gap 3) — HIGH priority
3. **Week 3-4**: High complexity refactoring (Gap 4) — prevents future bugs
4. **Week 4**: TODO cleanup (Gap 5) + buffer allocations (Gap 7)
5. **Ongoing**: Naming (Gap 6), LOD (Gap 8), profiling (Gap 9), docs (Gap 10)

**Total estimated effort: 5-6 weeks to close all gaps**

---

*Generated 2026-04-01. See AUDIT.md for full audit report.*
