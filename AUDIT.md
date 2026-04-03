# Ebitengine Game Audit Report
Generated: 2026-04-03

## Executive Summary
- **Total Issues**: 32
- **Critical**: 3 - Crashes, game-breaking bugs
- **High**: 7 - Major functionality/UX problems
- **Medium**: 8 - Noticeable bugs, moderate impact
- **Low**: 4 - Minor issues, edge cases
- **Optimizations**: 5 - Performance improvements
- **Code Quality**: 5 - Maintainability concerns

## Critical Issues

### [C-001] Custom Trigonometric Functions Are Inaccurate in Network Prediction
- **Location**: `pkg/network/prediction.go:194-219`
- **Category**: Logic
- **Description**: The `cos()`, `sin()`, and `mod()` functions are custom Taylor-series approximations instead of using `math.Cos`/`math.Sin`. The Taylor expansion `1 - x²/2 + x⁴/24 - x⁶/720` only uses 4 terms (up to x⁶), yielding significant error for large angles. The `mod()` function uses a loop-based subtraction (`for a >= b { a -= b }`) instead of `math.Mod`, which is O(n) for large values and will hang or become extremely slow for very large negative inputs.
- **Impact**: Client-side movement prediction accumulates angular drift over time, causing visible rubber-banding and server-client position desync. The `mod()` loop can cause frame stalls on extreme inputs.
- **Reproduction**:
  1. Connect a client to the server
  2. Rotate the player continuously (high accumulated angle values)
  3. Observe desynchronization between predicted and server-confirmed positions
- **Root Cause**: Reinventing `math.Cos`/`math.Sin`/`math.Mod` with inferior approximations instead of using the standard library.
- **Suggested Fix**: Replace `cos()`, `sin()`, `mod()` with `math.Cos()`, `math.Sin()`, `math.Mod()`.
- [x] **Resolved**

### [C-002] Operator Precedence Bug in Coup System Guard Clause
- **Location**: `pkg/engine/systems/faction_coup.go:395`
- **Category**: Logic
- **Description**: The condition `if !exists || coup.State != CoupStatePlotting && coup.State != CoupStateActive` is parsed by Go as `!exists || (coup.State != CoupStatePlotting && coup.State != CoupStateActive)` due to `&&` having higher precedence than `||`. The intended logic is likely `!exists || (coup.State != CoupStatePlotting && coup.State != CoupStateActive)` which actually matches Go's parsing, BUT the logical intent appears to be "return nil if not exists OR if the state is neither Plotting nor Active." The current expression correctly implements that intent. **However**, when `exists` is `false`, Go short-circuits the `||`, so `coup` is `nil` and `coup.State` is never evaluated — this is safe. The real issue is the **missing parentheses** make the code hard to read and review, and any refactor could accidentally introduce a nil-dereference.
- **Impact**: Code is fragile — a minor edit (e.g., changing `||` to `&&`) would cause a nil pointer dereference panic at runtime. The lack of parentheses violates code clarity principles.
- **Reproduction**:
  1. Call `getActiveCoupWithMembership()` for a faction with no active coup
  2. Currently safe due to short-circuit, but fragile
- **Root Cause**: Missing explicit parentheses around compound boolean expressions.
- **Suggested Fix**: Add parentheses: `if !exists || (coup.State != CoupStatePlotting && coup.State != CoupStateActive)`.
- [x] **Resolved**

### [C-003] Division by Zero in Sprite Texture Coordinate Calculation
- **Location**: `pkg/rendering/raycast/draw.go:127`
- **Category**: Rendering
- **Description**: `calculateSpriteTexX()` computes `(screenX - ctx.StartX) * ctx.SpriteWidth / ctx.ScreenSpriteWidth`. If `ctx.ScreenSpriteWidth` is 0 (which happens when `TransformY` is extremely large, i.e., a sprite is infinitely far away), this causes an integer division by zero panic.
- **Impact**: Game crashes when a sprite entity has extreme transform distance values.
- **Reproduction**:
  1. Place a sprite entity very far from the camera
  2. When `ScreenSpriteWidth` rounds to 0, the division panics
- **Root Cause**: No guard against zero denominator before integer division.
- **Suggested Fix**: Add `if ctx.ScreenSpriteWidth == 0 { return 0 }` guard before the division. Similarly check `ctx.ScreenSpriteHeight` at line 137.
- [x] **Resolved**

## High Priority Issues

### [H-001] FactionCoupSystem Public Methods Are Not Thread-Safe
- **Location**: `pkg/engine/systems/faction_coup.go:344-460`
- **Category**: State
- **Description**: The `FactionCoupSystem` has public methods (`StartCoup`, `PlayerStartCoup`, `GetCoup`, `GetCoupHistory`, `GetAllActiveCoups`, `SupportCoup`, `OpposeCoup`) that read and write `ActiveCoups` and `CoupHistory` maps without any synchronization. The `Update()` method (line 85) iterates and modifies these same maps. If any public method is called concurrently with `Update()` (e.g., from a network handler goroutine), this results in a concurrent map read/write panic.
- **Impact**: Runtime panic (`concurrent map read and map write`) under concurrent access, crashing the server.
- **Root Cause**: No mutex protection on shared map state.
- **Suggested Fix**: Add a `sync.RWMutex` to `FactionCoupSystem`. Use `RLock` in read-only methods (`GetCoup`, `GetCoupHistory`, `GetAllActiveCoups`) and `Lock` in mutating methods (`StartCoup`, `Update`, `finalizeCoup`).
- [x] **Resolved**

### [H-002] Vehicle Combat Cooldown Logic Is Inverted
- **Location**: `pkg/engine/systems/vehicle_combat.go:128-130`
- **Category**: Logic
- **Description**: The `Fire()` method checks `if weapon.LastFired < cooldown { return 0 }`. The field `LastFired` accumulates elapsed time since last shot. The guard prevents firing when `LastFired < cooldown` (i.e., not enough time has passed). However, `LastFired` is reset to 0 on line 131 immediately after the check passes. The issue is that `LastFired` starts at the default float64 zero value (0.0), meaning a weapon **cannot fire on its first attempt** after initialization because `0 < cooldown` is true.
- **Impact**: Newly equipped or first-use weapons fail to fire until `LastFired` accumulates past `cooldown` through system updates. Players experience "dead" weapons on first equip.
- **Root Cause**: `LastFired` is initialized to 0 instead of a value >= cooldown (e.g., `math.MaxFloat64` or initializing to `cooldown` value).
- **Suggested Fix**: Initialize `weapon.LastFired = cooldown` when weapons are created, or change the guard to `if weapon.LastFired < cooldown && weapon.LastFired > 0`.
- [x] **Resolved**

### [H-003] Auto-Save Creates Unbounded Goroutines Without Rate Limiting
- **Location**: `cmd/server/main.go:602-612`
- **Category**: Performance / State
- **Description**: `performAutoSave()` spawns a new goroutine each time it's called. If `pm.Save()` is slow (disk I/O, large world), the ticker fires again before the previous save completes, spawning another goroutine that also calls `createWorldSnapshot()`. There is no check for an in-progress save and no goroutine semaphore.
- **Impact**: Under heavy load or slow disk, multiple concurrent goroutines snapshot and save simultaneously, causing memory pressure (multiple full world snapshots) and potential file corruption from concurrent writes.
- **Root Cause**: No guard against concurrent auto-save operations.
- **Suggested Fix**: Add a `sync.Mutex` or `atomic.Bool` flag (`saving`) to skip the auto-save if one is already in progress.
- [x] **Resolved**

### [H-004] Network Server Spawns Goroutine Per Client Input
- **Location**: `pkg/network/server.go:359-360`
- **Category**: Performance
- **Description**: For every client input received, the server calls `s.wg.Add(1); go s.sendWorldState(conn, entityID, input.SequenceNum)`. At 60 Hz input rate per client, this spawns 60 goroutines/second/client. With 32 players, that's ~1920 goroutines/second.
- **Impact**: Goroutine churn increases GC pressure and scheduler overhead. While Go handles many goroutines, this pattern is wasteful — a per-client send channel with a dedicated writer goroutine would be more efficient.
- **Root Cause**: Using goroutine-per-message instead of goroutine-per-connection pattern.
- **Suggested Fix**: Use a buffered channel per connection with a dedicated sender goroutine, rather than spawning a goroutine per input message.
- [x] **Resolved**

### [H-005] Lag Compensator Has TOCTOU Race on Entity History
- **Location**: `pkg/network/lagcomp.go:179-182`
- **Category**: State
- **Description**: `HitTest()` acquires an `RLock`, copies `lc.entities[targetID]` pointer, then releases the lock. Subsequent operations on `targetHistory` (line 197: `GetAtTimeWithLimit`) happen outside the lock. If another goroutine removes the entity from `lc.entities` between the lock release and the method call, the `StateHistory` object may be concurrently modified or invalidated.
- **Impact**: Potential data race or stale hit results in lag-compensated combat.
- **Root Cause**: Lock scope is too narrow — should hold the read lock through the `GetAtTimeWithLimit` call.
- **Suggested Fix**: Extend the `RLock` scope to cover the `GetAtTimeWithLimit` call, or ensure `StateHistory` is independently thread-safe.
- [x] **Resolved**

### [H-006] Framebuffer Modified in Draw() Then Re-uploaded
- **Location**: `cmd/client/main.go:1287-1321, 1331-1354`
- **Category**: Rendering / Ebitengine-Specific
- **Description**: `applyCombatVisualFeedback()` and `drawDeathScreen()` are called from `Draw()`. They retrieve the renderer's framebuffer via `g.renderer.GetFramebuffer()`, modify pixels directly, then call `screen.WritePixels(framebuffer)`. While these methods don't mutate game *state* (they modify a rendering buffer), they do re-upload the framebuffer, overwriting whatever was previously drawn to `screen`. This means any NPC sprites drawn between the initial `Draw()` upload (line 20) and the combat flash are erased by the second `WritePixels()`.
- **Impact**: NPC sprites and post-processing effects are briefly visible then overwritten by the combat flash framebuffer upload. During damage flash, NPCs disappear for one frame.
- **Root Cause**: Double `WritePixels()` to the same screen in a single Draw call. The combat flash should be applied as a semi-transparent overlay using `DrawImage` with `ColorScale`, not by re-uploading the raw framebuffer.
- **Suggested Fix**: Apply combat flash as an overlay image with alpha blending using `screen.DrawImage()` instead of modifying and re-uploading the framebuffer.
- [x] **Resolved**

### [H-007] Client-Side Prediction Uses Degrees While Client Uses Radians
- **Location**: `pkg/network/prediction.go:187` vs `cmd/client/main.go:975-988`
- **Category**: Logic
- **Description**: The `ClientPredictor.applyInput()` converts `predictedAngle` from degrees to radians with `forwardRad := float32(cp.predictedAngle) * (3.14159265 / 180.0)`. However, the client's `processMovementInput()` (main.go:975) uses `pos.Angle` directly with `math.Cos(pos.Angle)`, treating it as radians. The server's `handleInput()` (server.go:348) also uses `math.Cos(float64(state.Angle))` directly. This means the client predictor expects degrees, but the rest of the system uses radians.
- **Impact**: Client-side prediction moves in completely wrong directions compared to actual server movement, causing severe rubber-banding when server corrections arrive.
- **Root Cause**: Mismatch between angle units (degrees vs. radians) across subsystems.
- **Suggested Fix**: Standardize on radians throughout. Remove the degree-to-radian conversion in `prediction.go:187` and use `math.Cos`/`math.Sin` directly.
- [x] **Resolved** (Fixed as part of C-001 - prediction.go now uses radians directly)

## Medium Priority Issues

### [M-001] CoupHistory Grows Without Bound
- **Location**: `pkg/engine/systems/faction_coup.go:227`
- **Category**: Performance / State
- **Description**: `finalizeCoup()` appends to `s.CoupHistory[factionID]` without any size limit. Over long server sessions with many faction coups, this slice grows indefinitely.
- **Impact**: Gradual memory leak proportional to server uptime and faction activity.
- **Root Cause**: No maximum history size or eviction policy.
- **Suggested Fix**: Add a history limit (e.g., 50 per faction) and trim like the `TradingSystem` does at line 432-434.
- [x] **Resolved**

### [M-002] DialogHistory in DialogConsequenceSystem Grows Without Bound
- **Location**: `pkg/engine/systems/dialog_consequence.go:364`
- **Category**: Performance / State
- **Description**: `state.DialogHistory = append(state.DialogHistory, ...)` has no size limit. Each NPC dialog exchange adds to the history indefinitely.
- **Impact**: Memory grows proportional to total NPC conversations. With 200+ NPCs and long play sessions, this accumulates significantly.
- **Root Cause**: No trimming of dialog history.
- **Suggested Fix**: Add a maximum dialog history size per entity (e.g., 100 entries) with FIFO eviction.
- [x] **Resolved**

### [M-003] Dead Code — `updateLinear` Method Never Called
- **Location**: `pkg/engine/systems/physics.go:71-126`
- **Category**: Code Quality
- **Description**: `updateLinear()` is defined but never called. Only `updateLinearWithCollision()` is used in `Update()` at line 65. This is 55 lines of dead code.
- **Impact**: Maintenance burden — developers may modify `updateLinear()` thinking it affects gameplay.
- **Root Cause**: Method was superseded by `updateLinearWithCollision()` but not removed.
- **Suggested Fix**: Remove `updateLinear()` or document it as an internal helper if it serves a purpose.
- [x] **Resolved**

### [M-004] Business Logic in ECS Components Violates Architecture
- **Location**: `pkg/engine/components/definitions.go:57-74, 97-115`
- **Category**: Code Quality
- **Description**: `FactionMembership` has methods `GetMembership()`, `GetRank()`, `IsMember()` containing query logic. `FactionTerritory` has `ContainsPoint()` implementing ray-casting point-in-polygon collision. Per ECS architecture, components should be pure data containers with only `Type()` methods; all logic belongs in systems.
- **Impact**: Violates the project's stated ECS discipline. Makes it harder to move logic to systems later. Creates coupling between component definitions and business rules.
- **Root Cause**: Convenience methods added directly to components rather than in systems.
- **Suggested Fix**: Move `GetMembership`/`GetRank`/`IsMember` logic into `FactionRankSystem`. Move `ContainsPoint` into a spatial query utility or the `FactionPoliticsSystem`.
- [ ] **Resolved**

### [M-005] Sentinel Values for Chunk Coordinates
- **Location**: `cmd/client/main.go:1753-1754`
- **Category**: Logic
- **Description**: `lastChunkX: -999, lastChunkY: -999` uses magic sentinel values to force initial map build. Negative chunk coordinates are theoretically valid in an infinite world.
- **Impact**: If a player spawns at or near chunk (-999, -999), the initial map build is skipped because the sentinel matches the actual position.
- **Root Cause**: Using magic numbers instead of an explicit `needsRebuild` boolean flag.
- **Suggested Fix**: Add a `chunkMapInitialized bool` field to `Game` and check it instead of relying on sentinel coordinates.
- [x] **Resolved**

### [M-006] Mouse Smoothing Accumulates State Between Frames Incorrectly
- **Location**: `cmd/client/main.go:1161-1167`
- **Category**: Input
- **Description**: Mouse smoothing applies a low-pass filter: `smoothedDeltaX = smoothedDeltaX*(1-factor) + deltaX*factor`. When the mouse is stationary (`deltaX = 0`), the smoothed value decays toward zero but never reaches it, causing phantom camera drift. The smoothing factor comes from config but has no bounds validation.
- **Impact**: Very slow residual camera drift when the mouse is still. With high smoothing factor values (>1.0 due to missing validation), camera behavior becomes erratic.
- **Root Cause**: No dead-zone threshold to zero out small smoothed deltas.
- **Suggested Fix**: Add a dead-zone: `if math.Abs(g.smoothedDeltaX) < 0.001 { g.smoothedDeltaX = 0 }`. Validate smoothing factor is in [0, 1].
- [x] **Resolved**

### [M-007] Turn Input Sends Hardcoded Values Instead of Delta-Time Scaled
- **Location**: `cmd/client/main.go:1065-1073`
- **Category**: Logic
- **Description**: `gatherTurnInput()` returns hardcoded `±0.05` for turn input sent to the server. This value is not scaled by delta time or TPS, meaning the turn rate varies based on the client's frame rate. The server applies this directly at line 347.
- **Impact**: Players with higher frame rates send more turn inputs per second, rotating faster than players with lower frame rates.
- **Root Cause**: Turn input should represent desired angular velocity (scaled by dt), not raw per-frame delta.
- **Suggested Fix**: Return `±1.0` and let the server scale by its tick rate, or scale by `dt` before sending.
- [x] **Resolved**

### [M-008] Strafe and Interaction Share Key E
- **Location**: `cmd/client/main.go:999, 318`
- **Category**: Input
- **Description**: `processStrafeInput()` uses `ebiten.KeyE` for strafe right (line 999). `handleInteraction()` uses `ebiten.KeyE` for interaction (line 318). Both are checked in `updateGameplay()`, so pressing E triggers both strafe movement AND interaction simultaneously.
- **Impact**: Player strafes right every time they interact with an NPC/item, causing unintended movement during dialog or pickups.
- **Root Cause**: Key binding conflict — same physical key mapped to two actions.
- **Suggested Fix**: Separate the bindings. Use a dedicated interaction key or consume the input after interaction handling.
- [x] **Resolved**

## Low Priority Issues

### [L-001] `particleBuffer` and `particleBufferSize` Fields Declared But Never Used
- **Location**: `cmd/client/main.go:98-99`
- **Category**: Code Quality
- **Description**: The `Game` struct declares `particleBuffer []byte` and `particleBufferSize int` but they are never assigned or read anywhere in the codebase.
- **Impact**: Minor dead code. No runtime effect.
- **Suggested Fix**: Remove the unused fields or implement particle buffer optimization.
- [x] **Resolved**

### [L-002] Layout() Returns Dynamic Dimensions — Renderer Not Resized
- **Location**: `cmd/client/main.go:1637-1640`
- **Category**: Ebitengine-Specific
- **Description**: `Layout()` returns `outsideWidth, outsideHeight`, making the game responsive to window resizing. However, the raycaster `Renderer` is initialized with fixed `cfg.Window.Width/Height` (line 1654) and its internal framebuffer, z-buffer, and floor/ceiling buffers are never resized.
- **Impact**: If the window is resized, `Layout()` returns new dimensions but the renderer still draws at the original resolution. The framebuffer upload via `WritePixels` may cause artifacts or panics if the screen size no longer matches the framebuffer size.
- **Root Cause**: Renderer dimensions are fixed at initialization and not updated on resize.
- **Suggested Fix**: Either return fixed dimensions from `Layout()` (matching renderer), or implement renderer resize support.
- [x] **Resolved**

### [L-003] Degree-to-Radian Uses Hardcoded Pi Constant
- **Location**: `pkg/network/prediction.go:187`
- **Category**: Code Quality
- **Description**: Uses `3.14159265 / 180.0` instead of `math.Pi / 180.0`. The hardcoded value has only 8 decimal places vs `math.Pi`'s 15.
- **Impact**: Minor precision loss in angle calculations.
- **Suggested Fix**: Use `math.Pi` from the standard library.
- [x] **Resolved** (Fixed as part of C-001 - degree conversion removed, angles are now radians throughout)

### [L-004] `pprof` Imported Unconditionally
- **Location**: `cmd/client/main.go:11`
- **Category**: Performance
- **Description**: `_ "net/http/pprof"` is imported with a blank identifier, registering pprof HTTP handlers in the default mux even when profiling is disabled. The profile server is only started conditionally (line 1649), but the handlers are always registered.
- **Impact**: Minor: pprof routes exist in the default HTTP mux even when not used. Not a security risk since no default HTTP server is started unless profiling is enabled.
- **Suggested Fix**: Move the import inside a build-tagged file (e.g., `//go:build debug`).
- [x] **Resolved**

## Performance Optimization Opportunities

### [P-001] Post-Processing Uses `RGBAAt()` Per-Pixel Instead of Direct Slice Access
- **Location**: `pkg/rendering/postprocess/effects.go` (multiple effects)
- **Category**: Rendering
- **Current Impact**: Post-processing effects iterate pixel-by-pixel using `image.RGBAAt()` which involves bounds checks per call.
- **Optimization**: Access the `Pix` slice directly (e.g., `img.Pix[offset:offset+4]`) to bypass per-pixel bounds checking.
- **Expected Improvement**: 20-40% speedup for post-processing passes based on typical Go image benchmarks.
- [x] **Resolved**

### [P-002] Replace Custom `mod()` Loop with `math.Mod`
- **Location**: `pkg/network/prediction.go:210-219`
- **Category**: Logic / Performance
- **Current Impact**: `mod()` uses iterative subtraction, which is O(n) where n = a/b. For very large accumulated angles, this becomes a hot loop.
- **Optimization**: Replace with `math.Mod(a, b)` which is O(1).
- **Expected Improvement**: Eliminates potential frame stalls from large angle accumulation.
- [x] **Resolved** (Fixed as part of C-001 - custom mod() removed, math.Mod used)

### [P-003] Goroutine-Per-Message Pattern in Network Server
- **Location**: `pkg/network/server.go:359-360`
- **Category**: Performance
- **Current Impact**: Creates ~60 goroutines/second/client for world state sends.
- **Optimization**: Use a per-connection buffered channel with a single dedicated sender goroutine.
- **Expected Improvement**: Reduces goroutine creation overhead and GC pressure. Estimated 50%+ reduction in scheduler overhead for network I/O.
- [x] **Resolved**

### [P-004] World Map Fully Rebuilt on Chunk Transition
- **Location**: `cmd/client/main.go:816-843`
- **Category**: Performance
- **Current Impact**: `rebuildWorldMap()` allocates a fresh `[][]int` grid (48×48) with 48 inner slice allocations every time the player crosses a chunk boundary.
- **Optimization**: Pre-allocate the world map once and reuse the buffer. Shift existing data and only generate the new edge chunks.
- **Expected Improvement**: Eliminates 49 allocations per chunk transition. Reduces GC pressure during movement.
- [x] **Resolved**

### [P-005] `syncSkyboxWithWorld()` Iterates All Entities Every Frame
- **Location**: `cmd/client/main.go:526-551`
- **Category**: Performance
- **Current Impact**: Calls `g.world.Entities("Weather")` and `g.world.Entities("WorldClock")` every frame to find singleton entities. Each call scans all entities.
- **Optimization**: Cache references to the weather and clock entities at creation time, or use a dedicated singleton lookup method.
- **Expected Improvement**: Eliminates O(n) entity scans per frame for what should be O(1) singleton lookups.
- [x] **Resolved**

## Code Quality Observations

### [Q-001] Magic Numbers for Turn Input Rate
- **Location**: `cmd/client/main.go:1067, 1070`
- **Category**: Code Quality
- **Issue**: `gatherTurnInput()` returns hardcoded `0.05` and `-0.05` without a named constant. The meaning of this value (radians per frame? per second?) is unclear.
- **Suggestion**: Define a constant like `keyboardTurnRate = 0.05` with a comment explaining units.
- [x] **Resolved** (Fixed as part of M-007 - now returns ±1.0 with documented behavior)

### [Q-002] Inconsistent Angle Units Across Codebase
- **Location**: `pkg/network/prediction.go:187` (degrees), `cmd/client/main.go:975` (radians), `pkg/network/server.go:348` (radians)
- **Category**: Code Quality
- **Issue**: The network prediction system assumes angles in degrees while the client and server use radians. This inconsistency makes it difficult to reason about angle-related code.
- **Suggestion**: Standardize all angle representations to radians throughout the codebase. Document the convention in the project's coding guidelines.
- [x] **Resolved** (Fixed as part of C-001/H-007 - all systems now use radians)

### [Q-003] Large `Game` Struct With 40+ Fields
- **Location**: `cmd/client/main.go:62-146`
- **Category**: Code Quality
- **Issue**: The `Game` struct has 40+ fields spanning rendering, audio, input, UI, networking, and state. This makes it a God Object that is difficult to test, extend, and reason about.
- **Suggestion**: Extract subsystem groups into dedicated structs (e.g., `RenderState`, `AudioState`, `UIState`, `NetworkState`).
- [ ] **Resolved**

### [Q-004] Multiple Input Checking Paths for Same Actions
- **Location**: `cmd/client/main.go:967-990` (processMovementInput), `1043-1061` (gatherMovementForward/Strafe), `1065-1073` (gatherTurnInput)
- **Category**: Code Quality
- **Issue**: Movement input is checked in two separate code paths: once for local application (processMovementInput) and once for network sending (gatherMovement*). The logic is duplicated and could diverge, causing local/remote inconsistency.
- **Suggestion**: Unify input gathering into a single `PlayerInputState` struct produced once per frame, then consumed by both local application and network sending.
- [ ] **Resolved**

### [Q-005] Missing `initUIBuffers` Guard for Window Resize
- **Location**: `cmd/client/main.go:556-579`
- **Category**: Code Quality
- **Issue**: `initUIBuffers()` pre-allocates fixed-size images (e.g., minimap 64×64, bars 150×16) but these are never re-created if the window resolution changes. With a responsive `Layout()`, the UI elements may be incorrectly sized.
- **Suggestion**: Either make UI buffer sizes relative to screen dimensions and re-create on resize, or fix `Layout()` to return constant dimensions.
- [x] **Resolved** (Not an issue: `Layout()` returns constant `cfg.Window.Width/Height`, so window resize is not supported; UI buffers remain correctly sized)

## Recommendations by Priority

1. **Immediate Action Required**
   - [C-001]: Replace custom trig functions with `math.Cos`/`math.Sin` — prediction is fundamentally broken
   - [C-003]: Add zero-guard in `calculateSpriteTexX` to prevent division-by-zero crash
   - [H-007]: Fix degree/radian mismatch in prediction — causes severe rubber-banding
   - [H-002]: Fix weapon cooldown initialization — weapons won't fire on first use

2. **High Priority (Next Sprint)**
   - [H-001]: Add mutex to FactionCoupSystem — concurrent map access panics
   - [H-006]: Fix double-WritePixels in Draw — NPCs disappear during combat flash
   - [H-004]: Refactor per-input goroutine spawning — scalability concern
   - [H-003]: Add auto-save rate limiting — potential data corruption
   - [M-008]: Fix E key conflict between strafe and interaction

3. **Medium Priority (Backlog)**
   - [H-005]: Extend lag compensator lock scope
   - [M-001]: Add CoupHistory size limit
   - [M-002]: Add DialogHistory size limit
   - [M-006]: Add mouse smoothing dead zone
   - [M-007]: Fix turn input delta-time scaling
   - [M-005]: Replace sentinel chunk values with boolean flag
   - [L-002]: Fix Layout/Renderer dimension mismatch

4. **Technical Debt**
   - [Q-003]: Decompose Game God Object into subsystem structs
   - [Q-004]: Unify input gathering into single code path
   - [Q-002]: Standardize angle units across codebase
   - [M-003]: Remove dead `updateLinear()` code
   - [M-004]: Move business logic out of ECS components

## Testing Recommendations
- **Network prediction accuracy**: Unit test comparing `cos()`/`sin()` output against `math.Cos()`/`math.Sin()` across full angle range to quantify drift
- **Weapon cooldown**: Test that newly created weapons can fire immediately (currently they cannot)
- **Concurrent map access**: Run `go test -race` on faction_coup package with concurrent `Update()` and `StartCoup()` calls
- **Sprite rendering edge cases**: Test with `ScreenSpriteWidth = 0` to verify division-by-zero guard
- **Window resize**: Test Layout() returns vs actual renderer dimensions after resize
- **Key conflict**: Automated input test that presses E and verifies only one action fires
- **Performance benchmarks**: Establish baselines for chunk rebuild time, entity iteration, and post-processing throughput

## Audit Methodology Notes
- **Analysis approach**: Static analysis of all Go source files in `cmd/`, `pkg/`, and `config/` directories. Cross-referenced function calls, data flow, and Ebitengine API contracts. Focused on files in the game loop hot path.
- **Areas not covered**: Test files (`*_test.go`) were not audited for correctness. Build tag variations (`noebiten`) were not tested. External dependency source code (Ebitengine, Viper) was not reviewed.
- **Assumptions**: Ebitengine v2.9.3 API contracts (Update/Draw/Layout thread model) are as documented. Go 1.24 memory model applies.
- **Limitations**: Static analysis cannot detect all race conditions or runtime-dependent bugs. Performance impact estimates are approximate. Some findings marked "Potential issue" where runtime behavior may differ from static reading.

## Positive Observations
- **Well-structured ECS core**: `pkg/engine/ecs/world.go` uses proper `sync.RWMutex` locking for entity/component operations. The `World` type is clean and well-tested.
- **Frame-rate independent game loop**: `Update()` correctly computes `dt` from `ebiten.ActualTPS()` with a sensible fallback (line 150-155). Movement code in `processMovementInput()` properly scales by `dt`.
- **Pre-allocated UI buffers**: The `initUIBuffers()` pattern (line 556) pre-allocates images and pixel buffers for minimap, health bars, crosshair, and speech bubbles. This avoids allocations in the Draw loop.
- **Wall-sliding collision**: The `tryMove()` / `canMoveTo()` pattern (lines 912-964) implements proper axis-separated wall sliding, checking X-only and Y-only movement when full movement is blocked.
- **Comprehensive system architecture**: 57 ECS system types covering factions, economy, combat, quests, crafting, and more show ambitious and thoughtful game design.
- **Proper Ebitengine patterns**: `Layout()` correctly uses responsive dimensions. `Draw()` does not modify game state (the framebuffer modification is a rendering concern, not game state). `Update()` returns `ebiten.Termination` for clean shutdown.
- **Lag compensation design**: The `LagCompensator` in `pkg/network/lagcomp.go` implements proper RTT-based rewind with configurable time windows — a sophisticated networking feature.
- **LOD system for chunk streaming**: `pkg/world/chunk/` implements LOD-based chunk management with async generation, showing good architectural planning for open-world performance.
