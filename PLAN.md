# Implementation Plan: Critical Multiplayer Fixes

**Generated**: 2026-04-03  
**Tool**: `go-stats-generator analyze . --skip-tests`  
**Codebase Version**: 42,606 lines of Go code across 189 source files

---

## Project Context

- **What it does**: Wyrm is a 100% procedurally generated first-person open-world RPG built in Go on Ebitengine, compiling to a single binary with zero external assets.
- **Current goal**: Close the 5 partial achievements blocking multiplayer from meeting README's stated claims (client-side prediction accuracy, thread safety, 200-5000ms latency support, delta compression, and 60 FPS UI performance).
- **Estimated Scope**: Medium (9 high-complexity functions, 392 duplicated lines, 3 outstanding TODOs in production code)

---

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Zero external assets | ✅ Achieved | No |
| 200 Features implemented | ✅ Achieved | No |
| ECS architecture | ✅ Achieved | No |
| Five genre themes | ✅ Achieved | No |
| Client-side prediction | ⚠️ Partial | **Yes** — custom trig functions cause drift |
| 200-5000ms latency tolerance | ⚠️ Partial | **Yes** — caps at ~2000ms |
| Delta compression | ⚠️ Partial | **Yes** — sends all fields unconditionally |
| 60 FPS performance | ⚠️ Partial | **Yes** — UI per-pixel Set() calls |
| Thread safety | ⚠️ Partial | **Yes** — FactionCoupSystem concurrent map access |

**Overall: 47/52 goals achieved (90%), 5 partial requiring attention**

---

## Metrics Summary

| Metric | Value | Target | Status |
|--------|-------|--------|--------|
| Complexity hotspots (overall >10) | 10 functions | 0 | ⚠️ Medium |
| Duplication ratio | 0.49% (392 lines) | <2.0% | ✅ Excellent |
| Doc coverage | 88.1% | >80% | ✅ Good |
| Circular dependencies | 0 | 0 | ✅ Excellent |
| High-coupling packages | 3 | <5 | ✅ Acceptable |

### Complexity Hotspots on Goal-Critical Paths

| Function | Package | Lines | Complexity | Goal Impact |
|----------|---------|-------|------------|-------------|
| `updateLinear` | systems | 54 | 14.0 | Dead code — remove |
| `initSpeechBubbleImage` | main | 41 | 13.7 | UI performance |
| `drawCoin` | sprite | 32 | 13.7 | Rendering |
| `DecodeEntityUpdate` | network | 29 | 13.5 | Delta compression |
| `findInteractableInRay` | main | 62 | 13.2 | Interaction |
| `SpawnNPCWithGenre` | adapters | 57 | 13.2 | NPC generation |
| `drawQuestList` | main | 51 | 13.2 | UI performance |
| `drawFood` | sprite | 36 | 13.2 | Rendering |
| `sellCargoAtDestination` | systems | 29 | 13.2 | Economy |
| `Draw` (crafting) | main | 77 | 12.7 | UI performance |

### Duplication on Goal-Critical Paths

| Clone Type | Lines | Locations | Impact |
|------------|-------|-----------|--------|
| System init | 7 | `cmd/client/init.go:117`, `cmd/server/main.go:461` | Maintainability |
| Vehicle physics | 9 | `vehicle_flying.go:625`, `vehicle_naval.go:536` | Bug propagation risk |
| UI buffer | 9 | `ui_buffer.go:59`, `ui_buffer_stub.go:51` | Build tag duplication |

---

## Dependency Health Notes

| Dependency | Version | Advisory |
|------------|---------|----------|
| Ebitengine | v2.9.3 | ✅ Stable — deprecated vector APIs not used |
| Viper | v1.19.0 | ⚠️ CVE-2026-33186 in remote gRPC dependency (not used by Wyrm) |
| golang.org/x/* | Current | ✅ No advisories |
| opd-ai/venture | v0.0.0-20260321 | ✅ Sibling repo |

No immediate dependency updates required. Monitor Viper for v2 release.

---

## Implementation Steps

### Step 1: Fix Client-Side Prediction Accuracy

- **Deliverable**: Replace custom `cos()`, `sin()`, `mod()` functions with `math` package equivalents and standardize on radians.
- **Dependencies**: None
- **Goal Impact**: **CRITICAL** — Enables functional multiplayer; fixes severe rubber-banding and desync (AUDIT.md C-001, H-007)
- **Acceptance**: Unit test comparing prediction output to `math.Cos`/`math.Sin` across full angle range; verify ≤0.001 radian drift after 1000 predictions

**Files to modify**:
- `pkg/network/prediction.go:187` — Remove degree-to-radian conversion
- `pkg/network/prediction.go:194-219` — Replace custom `cos()`, `sin()`, `mod()`

**Validation**:
```bash
go test -v ./pkg/network/... -run TestPrediction
# Verify no custom trig functions remain:
grep -n "func cos\|func sin\|func mod" pkg/network/prediction.go | wc -l  # Should be 0
```

---

### Step 2: Add Thread Safety to FactionCoupSystem

- **Deliverable**: Add `sync.RWMutex` to `FactionCoupSystem` protecting concurrent map access.
- **Dependencies**: None
- **Goal Impact**: **CRITICAL** — Prevents server crashes under concurrent faction activity (AUDIT.md H-001)
- **Acceptance**: `go test -race ./pkg/engine/systems/...` passes with concurrent coup operations

**Files to modify**:
- `pkg/engine/systems/faction_coup.go:344-460` — Add mutex, wrap reads with `RLock()`, writes with `Lock()`

**Validation**:
```bash
go test -race -v ./pkg/engine/systems/... -run TestFactionCoup
```

---

### Step 3: Guard Sprite Division by Zero

- **Deliverable**: Add zero-denominator guards to sprite texture coordinate calculations.
- **Dependencies**: None
- **Goal Impact**: **CRITICAL** — Prevents game crashes when sprites are at extreme distances (AUDIT.md C-003)
- **Acceptance**: Test with sprite at distance yielding 0 ScreenSpriteWidth; no panic

**Files to modify**:
- `pkg/rendering/raycast/draw.go:127` — Add `if ctx.ScreenSpriteWidth == 0 { return 0 }`
- `pkg/rendering/raycast/draw.go:137` — Add `if ctx.ScreenSpriteHeight == 0 { return 0 }`

**Validation**:
```bash
go test -v ./pkg/rendering/raycast/... -run TestSprite
```

---

### Step 4: Extend High-Latency Support to 5000ms

- **Deliverable**: Add tiered RTT thresholds, increase history buffer, extend max rewind time.
- **Dependencies**: Step 1 (prediction accuracy)
- **Goal Impact**: **HIGH** — Achieves README claim of 200-5000ms latency tolerance (GAPS.md §1)
- **Acceptance**: Connect client through artificial 5000ms latency; gameplay remains playable (no rubber-banding worse than 2s)

**Files to modify**:
- `pkg/network/prediction.go` — Add RTT thresholds: 800ms, 2000ms, 3500ms, 5000ms
- `pkg/network/lagcomp.go:9-14` — Increase `HistoryBufferSize` from 64 to 256
- `pkg/network/lagcomp.go` — Extend `MaxRewindTime` to 2000ms
- `pkg/network/client.go` — Add graduated input rate: 60Hz (<200ms), 30Hz (200-800ms), 20Hz (800-2000ms), 10Hz (>2000ms)

**Validation**:
```bash
# Constants check
grep -n "HistoryBufferSize.*256\|MaxRewindTime.*2000" pkg/network/lagcomp.go
```

---

### Step 5: Fix Vehicle Weapon Initialization

- **Deliverable**: Initialize `weapon.LastFired` to cooldown value when weapons are created.
- **Dependencies**: None
- **Goal Impact**: **HIGH** — Newly equipped weapons can fire on first attempt (AUDIT.md H-002)
- **Acceptance**: Newly spawned vehicle can fire weapon on first input

**Files to modify**:
- `pkg/engine/systems/vehicle_combat.go` — Initialize `weapon.LastFired = cooldown` on creation

**Validation**:
```bash
go test -v ./pkg/engine/systems/... -run TestVehicleCombat
```

---

### Step 6: Fix E Key Input Conflict

- **Deliverable**: Separate strafe right from interaction key binding.
- **Dependencies**: None
- **Goal Impact**: **HIGH** — Prevents unintended movement during interaction (AUDIT.md M-008)
- **Acceptance**: Press E near interactable; player interacts without moving

**Files to modify**:
- `cmd/client/main.go:999` — Change strafe right to `ebiten.KeyD` (standard WASD)

**Validation**:
```bash
# Ensure E is not used for strafe
grep -n "KeyE" cmd/client/main.go | grep -v "interact" | wc -l  # Should be 0 after fix
```

---

### Step 7: Rate-Limit Auto-Save

- **Deliverable**: Add mutex guard to prevent concurrent auto-save operations.
- **Dependencies**: None
- **Goal Impact**: **HIGH** — Prevents save corruption and memory pressure (AUDIT.md H-003)
- **Acceptance**: Only one save runs at a time under artificially slow disk I/O

**Files to modify**:
- `cmd/server/main.go:602-612` — Add `sync.Mutex` flag, skip save if previous in progress

**Validation**:
```bash
go test -race -v ./cmd/server/... -run TestAutoSave
```

---

### Step 8: Implement True Delta Compression

- **Deliverable**: Add field presence bitmask to `EntityUpdate`, only encode changed fields.
- **Dependencies**: None
- **Goal Impact**: **MEDIUM** — Reduces bandwidth usage by 50%+ for stationary entities (GAPS.md §2)
- **Acceptance**: Network traffic profiling shows ≥50% bandwidth reduction for stationary entities

**Files to modify**:
- `pkg/network/protocol.go` — Add `FieldMask uint8` to `EntityUpdate` struct
- `pkg/network/protocol.go` — Modify `Encode/Decode` to use bitmask
- `pkg/network/server.go` — Track baseline state per client, send deltas

**Validation**:
```bash
go-stats-generator analyze . --skip-tests --format json 2>&1 | grep -A 5 '"DecodeEntityUpdate"'
# Complexity should remain ≤13.5
```

---

### Step 9: Add Mutex to Lag Compensator

- **Deliverable**: Extend `RLock` scope to cover `GetAtTimeWithLimit` call.
- **Dependencies**: None
- **Goal Impact**: **MEDIUM** — Prevents race condition in hit registration (AUDIT.md H-005)
- **Acceptance**: `go test -race` passes with concurrent `HitTest` and `RecordState` calls

**Files to modify**:
- `pkg/network/lagcomp.go:179-182` — Extend lock scope

**Validation**:
```bash
go test -race -v ./pkg/network/... -run TestLagComp
```

---

### Step 10: Fix Double-WritePixels in Combat Flash

- **Deliverable**: Apply combat flash as overlay using `DrawImage` with `ColorScale` instead of re-uploading framebuffer.
- **Dependencies**: None
- **Goal Impact**: **MEDIUM** — NPCs remain visible during damage flash (AUDIT.md H-006)
- **Acceptance**: NPCs remain visible during damage flash

**Files to modify**:
- `cmd/client/main.go:1287-1321` — Replace `WritePixels()` with `DrawImage()` overlay

**Validation**:
```bash
# Ensure no double WritePixels in Draw path
grep -n "WritePixels" cmd/client/main.go | wc -l  # Should be 1 after fix
```

---

### Step 11: Bound History Data Structures

- **Deliverable**: Add max entry limits with FIFO eviction to `CoupHistory` and `DialogHistory`.
- **Dependencies**: Step 2 (FactionCoupSystem mutex)
- **Goal Impact**: **MEDIUM** — Prevents memory leaks over long server sessions (AUDIT.md M-001, M-002)
- **Acceptance**: Memory profiling after 10,000 simulated coups/dialogs shows bounded growth

**Files to modify**:
- `pkg/engine/systems/faction_coup.go:227` — Add max 50 entries per faction
- `pkg/engine/systems/dialog_consequence.go:364` — Add max 100 entries per entity

**Validation**:
```bash
go test -bench=. -benchmem ./pkg/engine/systems/... -run TestHistory
```

---

### Step 12: Fix Mouse Smoothing Dead Zone

- **Deliverable**: Add dead-zone threshold to zero out small smoothed deltas.
- **Dependencies**: None
- **Goal Impact**: **LOW** — Prevents phantom camera drift (AUDIT.md M-006)
- **Acceptance**: Mouse stationary for 5 seconds; camera completely still

**Files to modify**:
- `cmd/client/main.go:1161-1167` — Add `if math.Abs(g.smoothedDeltaX) < 0.001 { g.smoothedDeltaX = 0 }`

**Validation**:
```bash
go build ./cmd/client && echo "Build passes"
```

---

### Step 13: Remove Dead Code

- **Deliverable**: Remove unused `updateLinear()` function and unused struct fields.
- **Dependencies**: None
- **Goal Impact**: **LOW** — Reduces maintenance burden, removes complexity hotspot (AUDIT.md M-003, L-001)
- **Acceptance**: `go build` succeeds; no remaining references

**Files to modify**:
- `pkg/engine/systems/physics.go:71-126` — Remove `updateLinear()`
- `cmd/client/main.go:98-99` — Remove unused `particleBuffer`, `particleBufferSize`

**Validation**:
```bash
go-stats-generator analyze . --skip-tests 2>&1 | grep "updateLinear"  # Should return nothing
go build ./cmd/... && echo "Build passes"
```

---

### Step 14: Use Per-Connection Send Channel Pattern

- **Deliverable**: Replace goroutine-per-message with goroutine-per-connection pattern.
- **Dependencies**: Steps 1, 4, 8 (networking stack stabilization)
- **Goal Impact**: **LOW** — Reduces goroutine churn for better 32+ player scalability (AUDIT.md H-004)
- **Acceptance**: With 32 simulated clients at 60Hz input, goroutine count stays under 100

**Files to modify**:
- `pkg/network/server.go:359-360` — Create buffered channel per connection
- `pkg/network/server.go` — Spawn single sender goroutine per connection

**Validation**:
```bash
go test -v ./pkg/network/... -run TestGoroutineCount
```

---

## Implementation Schedule

| Week | Steps | Priority | Impact |
|------|-------|----------|--------|
| 1 | 1, 2, 3 | Critical | Enables functional multiplayer |
| 2 | 4, 5, 6, 7 | High | Achieves stated network/UX claims |
| 3 | 8, 9, 10 | Medium | Polish networking and rendering |
| 4 | 11, 12, 13, 14 | Low | Technical debt cleanup |

**Total estimated effort: 3-4 weeks**

---

## Success Metrics

After completing this plan:

| Metric | Before | After | Validation |
|--------|--------|-------|------------|
| Goals achieved | 47/52 (90%) | 52/52 (100%) | Manual verification |
| Critical audit issues | 3 | 0 | AUDIT.md review |
| High audit issues | 7 | 0 | AUDIT.md review |
| High-complexity functions | 10 | 9 | `go-stats-generator analyze . --skip-tests \| grep "High Complexity"` |
| Multiplayer playable | No | Yes | Connect 2+ clients, verify no desync |
| 5000ms latency tolerance | No | Yes | Artificial latency test |

---

## Validation Commands Summary

```bash
# Full validation suite
go build ./cmd/... && \
go test -race ./... && \
go vet ./... && \
go-stats-generator analyze . --skip-tests 2>&1 | grep -E "(High Complexity|Duplication Ratio|Overall Coverage)"

# Expected output after plan completion:
#   High Complexity (>10): 0 functions
#   Duplication Ratio: <0.5%
#   Overall Coverage: >88%
```

---

*Generated 2026-04-03 using `go-stats-generator`. Cross-referenced with ROADMAP.md, AUDIT.md, GAPS.md, and FEATURES.md.*
