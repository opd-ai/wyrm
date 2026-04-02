# Ebitengine Game Audit Report
Generated: 2026-04-02T04:40:06Z

## Executive Summary
- **Total Issues**: 52
- **Critical**: 5 — Crashes, division-by-zero, race conditions
- **High**: 8 — Major functionality/UX/stability problems
- **Medium**: 15 — Noticeable bugs, moderate impact
- **Low**: 8 — Minor issues, edge cases
- **Optimizations**: 7 — Performance improvements
- **Code Quality**: 9 — Maintainability concerns

**Codebase Scope**: 169 Go source files, 89 test files, ~53K implementation LOC, ~45K test LOC across 28 packages. Built on Ebitengine v2.9.3 with Go 1.24.5.

---

## Completion Checklist

Track issue resolution status below. Check off each item when the corresponding fix has been implemented and verified.

### Critical Issues
- [x] [C-001] Division by Zero in Floor/Ceiling Rendering
- [x] [C-002] Division by Zero in Billboard Camera Transform
- [x] [C-003] Race Condition in ECS RegisterSystem
- [x] [C-004] Division by Zero in Vignette Post-Processing
- [x] [C-005] Goroutine Leak in Network Server sendWorldState

### High Priority Issues
- [x] [H-001] Unsafe Component Access Pattern Across 11 System Files
- [x] [H-002] No Network Timeouts on Connections
- [ ] [H-003] Managers Initialized But Unused (Housing, PvP, Dialog, Companion)
- [ ] [H-004] FactionArcManager Created But Never Registered
- [x] [H-005] WorldMap Assumes Uniform Row Length
- [x] [H-006] uint8 Overflow in Subtitle Opacity
- [x] [H-007] Audio Player Resource Leak on Error Path
- [x] [H-008] No Configuration Value Validation

### Medium Priority Issues
- [ ] [M-001] Entities() Query Allocates on Every Call
- [ ] [M-002] time.Sleep in Network Sync Code
- [x] [M-003] Server Tick Rate Division by Zero Risk
- [x] [M-004] Z-Fighting in Billboard Rendering
- [x] [M-005] Ceiling Rendering Row Asymmetry
- [ ] [M-006] RenderSystem is an Empty Stub
- [x] [M-007] Magic Number 1.5708 (π/2) Used Directly
- [x] [M-008] isValidMapCellPosition Also Assumes Uniform Row Length
- [x] [M-009] Dialog System Returns Pointer to Slice Element
- [x] [M-010] Audio Player Has No Close/Cleanup Method
- [x] [M-011] Potential Integer Overflow in Subtitle Duration Division
- [ ] [M-012] Unused Puzzle and Object Data in Server Init
- [x] [M-013] Large Texture Distortion Scale Causes Aliasing
- [x] [M-014] Framebuffer Index Safety with Zero Dimensions
- [x] [M-015] Particle Truncation Instead of Rounding

### Low Priority Issues
- [ ] [L-001] Stale Entity References After Destruction
- [ ] [L-002] DestroyEntity Succeeds Silently for Non-Existent Entities
- [ ] [L-003] Color Blending Precision in Skybox
- [ ] [L-004] Particle Alpha Transition Sharp at Boundaries
- [x] [L-005] Type Assertion Without ok Check in Quest UI
- [x] [L-006] Unsafe Type Assertion in Trade Route System
- [x] [L-007] Lighting Direction Zero Vector
- [ ] [L-008] Bloom Edge Artifacts

### Performance Optimizations
- [ ] [P-001] ECS Entities() Query Hot Path Allocation
- [ ] [P-002] Sort on Every Entities() Call
- [ ] [P-003] FOV Ray Directions Recalculated Per Row
- [ ] [P-004] sendWorldState Copies All Player States Under Lock
- [ ] [P-005] Redundant Map Lookups in ECS Component Access
- [ ] [P-006] Auto-Save Creates Full World Snapshot Under Unknown Locking
- [ ] [P-007] Particle Glow Uses sqrt for Distance

### Code Quality
- [ ] [Q-001] 20 Component Structs Missing Type() Method
- [ ] [Q-002] Server System Count Hardcoded as String Literal
- [ ] [Q-003] Unused Variable Suppression Pattern
- [ ] [Q-004] Client Main File Exceeds 2,250 Lines
- [ ] [Q-005] Inconsistent Error Handling in AddComponent Callers
- [ ] [Q-006] Helper Functions for Trivial Math Operations
- [ ] [Q-007] Magic Numbers in Movement and Physics
- [ ] [Q-008] Computed Values in Systems Discarded with _ Assignment
- [ ] [Q-009] Non-Component Helper Structs Mixed with Components

---

## Critical Issues

### [C-001] Division by Zero in Floor/Ceiling Rendering
- **Status**: Open
- **Location**: `pkg/rendering/raycast/draw.go:155`
- **Category**: Rendering
- **Description**: `calculateRowDistance()` computes `posZ / float64(p)` where `p = y - halfHeight`. When `y == halfHeight` (the first iteration of the floor loop at line 132), `p == 0`, causing a division by zero that produces `+Inf`, which then propagates into texture coordinate calculations.
- **Impact**: Produces `+Inf`/`NaN` values in the floor rendering row at the horizon line. In Go, floating-point division by zero yields `+Inf` (not a panic), but this corrupts all downstream calculations for that row—`floorStepX/Y`, texture coordinates, and fog distance become invalid, potentially causing visual artifacts at the horizon.
- **Reproduction**:
  1. Start the client with any configuration
  2. The floor rendering loop at line 132 starts at `y = halfHeight`, so `p = 0` on the very first iteration
  3. Observe `rowDistance = +Inf`, cascading to `floorStepX/Y = +Inf`
- **Root Cause**: The loop starts at `y = halfHeight` (line 132) but the distance formula requires `y > halfHeight` to produce finite values.
- **Suggested Fix**: Start the loop at `y = halfHeight + 1`, or add a guard: `if p == 0 { continue }` before the division.

### [C-002] Division by Zero in Billboard Camera Transform
- **Status**: Open
- **Location**: `pkg/rendering/raycast/billboard.go:60`
- **Category**: Rendering
- **Description**: `invDet := 1.0 / (planeX*dirY - dirX*planeY)` computes the inverse of the camera matrix determinant. When the camera plane vector is parallel to the direction vector, the determinant is zero, causing division by zero.
- **Impact**: Produces `+Inf` in `TransformX` and `TransformY`, causing sprites to be drawn at extreme/invalid screen positions. Could cause massive overdraw or framebuffer index issues downstream.
- **Reproduction**:
  1. This occurs when `planeX*dirY == dirX*planeY`
  2. Since `planeX = -dirY * tan(FOV/2)` and `planeY = dirX * tan(FOV/2)` (lines 55-57), the determinant simplifies to `-(dirY² + dirX²) * tan(FOV/2)`, which is zero only if `dirX == dirY == 0` (zero-length direction vector) or `FOV == 0`
  3. With a valid FOV (>0) and non-zero direction, this is unlikely in normal play but could occur during initialization or edge-case angle resets
- **Root Cause**: No guard on the determinant value before computing its inverse.
- **Suggested Fix**: Check `det := planeX*dirY - dirX*planeY; if math.Abs(det) < 1e-10 { return false }` before line 60.

### [C-003] Race Condition in ECS RegisterSystem
- **Status**: Open
- **Location**: `pkg/engine/ecs/world.go:115-116`
- **Category**: State Management
- **Description**: `RegisterSystem()` appends to `w.systems` without holding the mutex lock. Meanwhile, `Update()` (line 120-123) iterates over `w.systems` without any lock. If a system is registered while `Update()` is running (e.g., during late initialization or dynamic system registration), the append may cause a data race on the slice header.
- **Impact**: Potential data race detectable by `go test -race`. Could cause the newly appended system to be missed, or in extreme cases, slice corruption leading to a panic.
- **Reproduction**:
  1. Call `world.RegisterSystem()` from one goroutine
  2. Call `world.Update()` from another goroutine concurrently
  3. Run with `-race` flag to detect
- **Root Cause**: `RegisterSystem` is the only World method that modifies shared state without acquiring the mutex.
- **Suggested Fix**: Add `w.mu.Lock()` / `w.mu.Unlock()` around the append in `RegisterSystem()`.

### [C-004] Division by Zero in Vignette Post-Processing
- **Status**: Open
- **Location**: `pkg/rendering/postprocess/effects.go:282`
- **Category**: Rendering
- **Description**: `falloff := (dist - v.Radius) / (1.0 - v.Radius + v.Softness)` divides by zero when `v.Radius == 1.0` and `v.Softness == 0.0`, producing `+Inf` or `NaN`.
- **Impact**: The vignette effect corrupts all pixels in the affected region, potentially causing a fully black or fully white screen. Since post-processing runs every frame, the effect is persistent.
- **Reproduction**:
  1. Configure a vignette effect with `Radius = 1.0` and `Softness = 0.0`
  2. Any pixel where `dist > Radius` triggers the division
- **Root Cause**: No guard on the denominator.
- **Suggested Fix**: Add `denom := 1.0 - v.Radius + v.Softness; if denom <= 0 { denom = 0.001 }` before the division.

### [C-005] Goroutine Leak in Network Server sendWorldState
- **Status**: Open
- **Location**: `pkg/network/server.go:350`
- **Category**: State Management / Performance
- **Description**: `go s.sendWorldState(conn, entityID, input.SequenceNum)` spawns a goroutine for every player input message, but these goroutines are not tracked in the `sync.WaitGroup`. If a client disconnects while `sendWorldState` is writing to the connection (line 395), the goroutine may block indefinitely on the write since no write deadline is set.
- **Impact**: Under high load or slow clients, goroutines accumulate. During server shutdown, `Stop()` waits on `s.wg.Wait()` (line 551) but these goroutines are not in the WaitGroup—so shutdown completes while goroutines still hold locks or write to closed connections.
- **Reproduction**:
  1. Connect multiple clients
  2. Have clients send input rapidly
  3. Disconnect one client's network abruptly (not gracefully)
  4. The `sendWorldState` goroutine for that client blocks on `Encode(conn)` at line 395
- **Root Cause**: `sendWorldState` goroutines not tracked in WaitGroup, and no write deadline on connections.
- **Suggested Fix**: Add `s.wg.Add(1)` before the goroutine spawn and `defer s.wg.Done()` inside `sendWorldState`. Set `conn.SetWriteDeadline()` before encoding.

---

## High Priority Issues

### [H-001] Unsafe Component Access Pattern Across 11 System Files
- **Status**: Open
- **Location**: `pkg/engine/systems/gossip.go:72,79,81,94,101`, `pkg/engine/systems/hazard.go:64,83,106,115,140,145,171,178,206,232,247,253,260`, `pkg/engine/systems/magic_combat.go:195,241,318`, `pkg/engine/systems/emotional_state.go:117`, `pkg/engine/systems/npc_memory.go:107`, `pkg/engine/systems/dialog_consequence.go:220`, `pkg/engine/systems/multi_npc_conversation.go:345,437`, `pkg/engine/systems/pardon.go:238`, `pkg/engine/systems/vehicle_physics.go:56-58`, `pkg/engine/systems/crime.go:799`, `pkg/engine/systems/city_buildings.go:242,368,380`
- **Category**: State Management
- **Description**: Over 40 instances of `comp, _ := w.GetComponent(e, "TypeName")` where the boolean ok value is discarded, followed by an unsafe type assertion `comp.(*components.SomeType)`. If `GetComponent` returns `(nil, false)`, the type assertion on nil causes a nil pointer dereference panic.
- **Impact**: Any entity missing an expected component crashes the server/client. While the ECS query `w.Entities("Type1", "Type2")` should guarantee those components exist, entities can be destroyed between the query and the access, or component names could be misspelled.
- **Suggested Fix**: Always check `ok` before casting: `comp, ok := w.GetComponent(e, "X"); if !ok { continue }`. Or create a helper: `MustGetComponent[T](w, e, name) *T` that handles the check.

### [H-002] No Network Timeouts on Connections
- **Status**: Open
- **Location**: `pkg/network/server.go:86`, `pkg/network/server.go:375-398`
- **Category**: Performance / State Management
- **Description**: No `SetReadDeadline`, `SetWriteDeadline`, or `SetDeadline` calls anywhere in the network package. The `Accept()` call at line 86 blocks indefinitely. Client read loops have no timeout. Write operations in `sendWorldState` (line 395) block indefinitely if the peer stops reading.
- **Impact**: A single slow or malicious client can cause goroutine accumulation. Under adversarial conditions, an attacker can exhaust server goroutines by connecting and never sending data. This violates the ROADMAP.md requirement for 200-5000ms latency tolerance.
- **Suggested Fix**: Set `conn.SetDeadline(time.Now().Add(10*time.Second))` per the project's high-latency design mandate. Add idle connection detection via heartbeat absence.

### [H-003] Managers Initialized But Unused (Housing, PvP, Dialog, Companion)
- **Status**: Open
- **Location**: `cmd/server/main.go:95-106`
- **Category**: Code Quality / State Management
- **Description**: `initializeManagers()` creates HouseManager, ZoneManager, DialogManager, and CompanionManager, then immediately discards them with `_ = hm`, `_ = zm`, `_ = dm`, `_ = compMgr`. These managers are allocated, consume memory, and are never connected to any system.
- **Impact**: Wasted initialization and memory. Any system that needs these managers has no way to access them. This is a dangling-feature anti-pattern per the project's copilot instructions.
- **Suggested Fix**: Either pass these managers to systems that need them, or defer their initialization until they are actually integrated.

### [H-004] FactionArcManager Created But Never Registered
- **Status**: Open
- **Location**: `cmd/server/main.go:455-456`
- **Category**: Code Quality / State Management
- **Description**: `factionArcManager := systems.NewFactionArcManager(genre)` is created but immediately discarded with `_ = factionArcManager`. It is never registered as a system and never connected to the quest system.
- **Impact**: Faction quest arcs are silently non-functional. Players expecting faction-driven quests will find none.
- **Suggested Fix**: Register the manager as a system, or pass it to the quest system.

### [H-005] WorldMap Assumes Uniform Row Length
- **Status**: Open
- **Location**: `pkg/rendering/raycast/renderer.go:390-392`
- **Category**: Rendering
- **Description**: `isValidMapPosition(x, y)` checks `y < len(r.WorldMap[0])` using the first row's length. If any row has a different length, accessing `r.WorldMap[x][y]` could panic with index out of range.
- **Impact**: If the map generator produces a jagged 2D slice (non-uniform row lengths), the renderer crashes.
- **Suggested Fix**: Check `y < len(r.WorldMap[x])` instead of `y < len(r.WorldMap[0])` after validating `x`.

### [H-006] uint8 Overflow in Subtitle Opacity
- **Status**: Open
- **Location**: `pkg/rendering/subtitles/renderer.go:344`
- **Category**: Rendering
- **Description**: `ss.style.BackgroundColor[3] = uint8(opacity * 255)` converts a `float64` to `uint8`. If `opacity > 1.0` (no clamping at call site), the result exceeds 255 and wraps around due to Go's uint8 truncation. E.g., `opacity = 1.1` → `280.5` → `uint8(280) = 24`.
- **Impact**: Setting opacity slightly above 1.0 produces a nearly transparent background instead of fully opaque, which is the opposite of user intent.
- **Suggested Fix**: Clamp opacity: `if opacity > 1.0 { opacity = 1.0 } else if opacity < 0 { opacity = 0 }`.

### [H-007] Audio Player Resource Leak on Error Path
- **Status**: Open
- **Location**: `pkg/audio/player.go:31-38`
- **Category**: Assets / State Management
- **Description**: `NewPlayer()` calls `audio.NewContext(AudioSampleRate)` at line 32, then `ctx.NewPlayer(stream)` at line 35. If `NewPlayer` fails, the function returns the error but never closes the `audio.Context`. The context is leaked.
- **Impact**: Each failed audio player creation leaks an Ebitengine audio context. Over repeated failures, this accumulates resources.
- **Suggested Fix**: On error, clean up the context before returning.

### [H-008] No Configuration Value Validation
- **Status**: Open
- **Location**: `config/load.go:246-257`
- **Category**: State Management
- **Description**: After unmarshaling config, no validation is performed. Values like `Server.TickRate = 0` would cause division by zero at `cmd/server/main.go:527` (`time.Second / time.Duration(0)`). Negative multipliers, zero window dimensions, and invalid port numbers are all accepted silently.
- **Impact**: Invalid config can crash the server or client at runtime with obscure panics.
- **Suggested Fix**: Add a `Validate()` method that checks: `TickRate > 0`, `Window.Width > 0`, `Window.Height > 0`, port range `1-65535`, non-negative multipliers.

---

## Medium Priority Issues

### [M-001] Entities() Query Allocates on Every Call
- **Status**: Open
- **Location**: `pkg/engine/ecs/world.go:97`
- **Category**: Performance
- **Description**: `var result []Entity` in `Entities()` creates a nil slice that grows via `append`. This method is called by every system on every tick for every component query. The `sort.Slice` at line 110 also allocates internally.
- **Impact**: High allocation rate in the hot path. With 60 systems each querying once per tick at 20 Hz server rate, this is 1,200 allocations/second minimum, increasing GC pressure.
- **Suggested Fix**: Pre-allocate with `result := make([]Entity, 0, len(w.components))` as a capacity hint.

### [M-002] time.Sleep in Network Sync Code
- **Status**: Open
- **Location**: `cmd/client/sync.go:70,77`
- **Category**: Performance
- **Description**: `time.Sleep(100 * time.Millisecond)` used in the network synchronization polling loop. This blocks the goroutine for a fixed duration regardless of actual network conditions.
- **Impact**: Adds 100ms of artificial latency to synchronization. Under good network conditions, this unnecessarily delays state sync.
- **Suggested Fix**: Use `time.NewTicker` or channel-based signaling instead of sleep-polling.

### [M-003] Server Tick Rate Division by Zero Risk
- **Status**: Open
- **Location**: `cmd/server/main.go:527`
- **Category**: Logic
- **Description**: `tickInterval := time.Second / time.Duration(cfg.Server.TickRate)` panics if `TickRate` is 0. No validation prevents this.
- **Impact**: Server crashes on startup with zero tick rate config.
- **Suggested Fix**: Validate `cfg.Server.TickRate > 0` in config validation, or add a guard: `if cfg.Server.TickRate == 0 { cfg.Server.TickRate = 20 }`.

### [M-004] Z-Fighting in Billboard Rendering
- **Status**: Open
- **Location**: `pkg/rendering/raycast/billboard.go:84`
- **Category**: Rendering
- **Description**: `if ctx.Distance >= r.GetZBufferAt(screenX)` uses `>=` for the depth test. Sprites at exactly the same distance as a wall are culled. This can cause sprites to flicker between visible and hidden when aligned with walls.
- **Impact**: Visual flickering of sprites near walls, especially noticeable with NPCs standing against walls.
- **Suggested Fix**: Use `>` instead of `>=` to render sprites at equal distance, or add a small epsilon bias.

### [M-005] Ceiling Rendering Row Asymmetry
- **Status**: Open
- **Location**: `pkg/rendering/raycast/draw.go:130-132`
- **Category**: Rendering
- **Description**: `halfHeight := r.Height / 2` uses integer division. For odd heights (e.g., 721), `halfHeight = 360`. The floor loop runs for rows 360-720 (361 rows), but ceiling mirroring via `ceilY = Height - y - 1` maps to rows 0-360 (361 rows). Row 360 is rendered twice (once as floor, once as ceiling), causing a 1-pixel overlap at the horizon.
- **Impact**: A faint visual artifact at the exact center of the screen. Barely noticeable at most resolutions.
- **Suggested Fix**: Use `halfHeight + 1` as the floor loop start, or handle the center row explicitly.

### [M-006] RenderSystem is an Empty Stub
- **Status**: Open
- **Location**: `pkg/engine/systems/render.go:15-18`
- **Category**: Code Quality
- **Description**: `RenderSystem.Update()` body contains only `_ = w` to suppress the unused variable warning. The system is registered but does nothing. The actual rendering is handled directly by the client's `Draw()` method.
- **Impact**: A registered system consuming iteration time with zero output. Misleading to developers who expect render preparation logic here.
- **Suggested Fix**: Either implement render preparation logic (culling, LOD selection) or remove the system registration.

### [M-007] Magic Number 1.5708 (π/2) Used Directly
- **Status**: Open
- **Location**: `pkg/network/server.go:342-343`, `cmd/client/main.go:411`
- **Category**: Code Quality
- **Description**: The value `1.5708` appears as a magic number for π/2 radians in movement calculations and door animations. It's an approximation (actual π/2 = 1.5707963...) with ~0.00002 error.
- **Impact**: Minor directional drift in strafing movement over long distances. Door angles slightly imprecise.
- **Suggested Fix**: Use `math.Pi / 2` for exact precision and readability.

### [M-008] isValidMapCellPosition Also Assumes Uniform Row Length
- **Status**: Open
- **Location**: `pkg/rendering/raycast/renderer.go:394-400`
- **Category**: Rendering
- **Description**: Same issue as H-005 but for `WorldMapCells`: checks `y < len(r.WorldMapCells[0])` instead of `y < len(r.WorldMapCells[x])`.
- **Impact**: Potential panic with jagged cell arrays.
- **Suggested Fix**: Same as H-005—check the specific row length.

### [M-009] Dialog System Returns Pointer to Slice Element
- **Status**: Open
- **Location**: `pkg/dialog/system.go:125`
- **Category**: State Management
- **Description**: `GetLastTopic()` returns `&memory.Topics[len(memory.Topics)-1]`, a pointer to the last element in the internal slice. If the slice is modified later (topics appended or trimmed), the pointer may reference stale or relocated data.
- **Impact**: Callers holding the returned pointer across modifications see corrupted data.
- **Suggested Fix**: Return a copy: `t := memory.Topics[len(memory.Topics)-1]; return &t`.

### [M-010] Audio Player Has No Close/Cleanup Method
- **Status**: Open
- **Location**: `pkg/audio/player.go:22-28`
- **Category**: Assets
- **Description**: The `Player` struct holds an `*audio.Context` and `*audio.Player` but provides no `Close()` method. When a Player is discarded, these Ebitengine resources are not explicitly released.
- **Impact**: Resource leak if players are created and destroyed (e.g., per-scene audio).
- **Suggested Fix**: Add a `Close()` method that calls `p.player.Close()`.

### [M-011] Potential Integer Overflow in Subtitle Duration Division
- **Status**: Open
- **Location**: `pkg/rendering/subtitles/renderer.go:448`
- **Category**: Logic
- **Description**: `RemainingFraction: float64(sub.RemainingTime(time.Now())) / float64(sub.Duration)` divides by `sub.Duration`. If a subtitle is created with `Duration = 0` (bypassing the `Add()` method which enforces `minDisplayTime`), this is division by zero.
- **Impact**: Produces `NaN` or `+Inf` in the subtitle animation state.
- **Suggested Fix**: Guard with `if sub.Duration == 0 { return 0 }`.

### [M-012] Unused Puzzle and Object Data in Server Init
- **Status**: Open
- **Location**: `cmd/server/init.go:344,401,420`
- **Category**: Code Quality
- **Description**: Puzzle data, object data, and zone seeds are generated during initialization then immediately discarded: `_ = puzzle`, `_ = obj`, `_ = seed`.
- **Impact**: Wasted CPU cycles generating content that is never used. Initialization takes longer than necessary.
- **Suggested Fix**: Defer generation until actually needed, or wire the data into the systems that will consume it.

### [M-013] Large Texture Distortion Scale Causes Aliasing
- **Status**: Open
- **Location**: `pkg/rendering/texture/patterns.go:219-220`
- **Category**: Rendering
- **Description**: Distortion pattern uses `distortX := noise.Noise2D(...) * 10.0` with a 10-pixel distortion amplitude. Combined with `NoiseScale`, this produces large jumps in noise space (±1.0), causing visible banding and aliasing in procedural textures.
- **Impact**: Generated textures may show harsh visual banding instead of smooth distortion.
- **Suggested Fix**: Reduce distortion amplitude or add anti-aliasing (bilinear sampling of the displaced coordinate).

### [M-014] Framebuffer Index Safety with Zero Dimensions
- **Status**: Open
- **Location**: `pkg/rendering/raycast/renderer.go:244-252`
- **Category**: Rendering
- **Description**: `SetPixel()` checks `x < r.Width && y < r.Height` but if `r.Width` or `r.Height` is 0, the `Framebuffer` slice is empty. While the bounds check prevents out-of-range x/y, the idx calculation `(y*Width + x) * 4` could be 0, but `r.Framebuffer[0]` on an empty slice would panic. This only occurs if the renderer is initialized with zero dimensions.
- **Impact**: Panic on zero-dimension renderer initialization.
- **Suggested Fix**: Validate `Width > 0 && Height > 0` in `NewRenderer()`.

### [M-015] Particle Truncation Instead of Rounding
- **Status**: Open
- **Location**: `pkg/rendering/particles/renderer.go:46`
- **Category**: Rendering
- **Description**: `screenX := int(p.X * float64(r.width))` truncates instead of rounding. A particle at `p.X = 0.999` with `width = 1280` produces `screenX = 1279` instead of `1280`. This creates a systematic 0.5-pixel leftward/upward bias for all particles.
- **Impact**: Minor visual offset—particles drift slightly toward the top-left corner.
- **Suggested Fix**: Use `int(math.Round(p.X * float64(r.width)))`.

---

## Low Priority Issues

### [L-001] Stale Entity References After Destruction
- **Status**: Open
- **Location**: `pkg/engine/ecs/world.go:53-57,94-112`
- **Category**: State Management
- **Description**: `DestroyEntity()` removes an entity from the components map, but any system that cached the entity ID from a previous `Entities()` call still holds a stale reference. Subsequent `GetComponent()` calls return `(nil, false)` but callers often ignore the boolean (see H-001).
- **Impact**: Silent nil returns that cascade to panics when the ok value is not checked.
- **Suggested Fix**: Document the stale-reference risk in `Entities()` godoc. Long-term: add entity generation numbers.

### [L-002] DestroyEntity Succeeds Silently for Non-Existent Entities
- **Status**: Open
- **Location**: `pkg/engine/ecs/world.go:53-57`
- **Category**: State Management
- **Description**: `DestroyEntity()` calls `delete(w.components, e)` without checking if the entity exists. Deleting a non-existent key is a no-op in Go, so this silently succeeds.
- **Impact**: Makes debugging double-deletion bugs difficult. No error path to detect logic errors.
- **Suggested Fix**: Optionally return `error` or `bool` to indicate if the entity existed.

### [L-003] Color Blending Precision in Skybox
- **Status**: Open
- **Location**: `pkg/rendering/raycast/skybox.go:269`
- **Category**: Rendering
- **Description**: Color interpolation `uint8(float64(a.R)*(1-t) + float64(b.R)*t)` can lose ±1 LSB due to float64→uint8 truncation.
- **Impact**: Imperceptible 1-bit color error in sky gradient. No visual significance.
- **Suggested Fix**: Use `math.Round()` before uint8 conversion for perfectionism.

### [L-004] Particle Alpha Transition Sharp at Boundaries
- **Status**: Open
- **Location**: `pkg/rendering/particles/emitter.go:411-418`
- **Category**: Rendering
- **Description**: Fade-in at `lifeRatio > 0.9` and fade-out at `lifeRatio < 0.3` use linear ramps. At the boundary values (0.9 and 0.3 exactly), the transition has a sharp discontinuity in the derivative (C0 continuity but not C1).
- **Impact**: Barely perceptible hard edge in particle fade animation.
- **Suggested Fix**: Use smoothstep instead of linear interpolation for smoother transitions.

### [L-005] Type Assertion Without ok Check in Quest UI
- **Status**: Open
- **Location**: `cmd/client/quest_ui.go:412-413`
- **Category**: State Management
- **Description**: `bgColor := q.getBackgroundColor().(color.RGBA)` performs a bare type assertion. If `getBackgroundColor()` returns a different `color.Color` implementation, this panics.
- **Impact**: Panic if the color type changes. Currently safe since `getBackgroundColor()` always returns `color.RGBA`, but brittle.
- **Suggested Fix**: Use `bgColor, ok := q.getBackgroundColor().(color.RGBA); if !ok { ... }`.

### [L-006] Unsafe Type Assertion in Trade Route System
- **Status**: Open
- **Location**: `pkg/engine/systems/trade_route.go:480`
- **Category**: State Management
- **Description**: `dest := destComp.(interface{ GetPrice(string) float64 })` asserts a structural interface on a component. If the component doesn't have a `GetPrice` method, this panics.
- **Impact**: Panic if component types change or if a different component is mistakenly registered with the same type name.
- **Suggested Fix**: Use comma-ok assertion: `dest, ok := destComp.(interface{ GetPrice(string) float64 })`.

### [L-007] Lighting Direction Zero Vector
- **Status**: Open
- **Location**: `pkg/rendering/lighting/system.go:66-70`
- **Category**: Rendering
- **Description**: If all direction components (dirX, dirY, dirZ) are zero, normalization is skipped, leaving a zero-length direction vector. Downstream dot-product lighting calculations produce zero illumination regardless of surface orientation.
- **Impact**: A light with zero direction produces no directional contribution, which may be confusing but is not a crash.
- **Suggested Fix**: Log a warning when a zero-direction light is encountered.

### [L-008] Bloom Edge Artifacts
- **Status**: Open
- **Location**: `pkg/rendering/postprocess/effects.go:365-366`
- **Category**: Rendering
- **Description**: The bloom blur loop starts at `Min + blurRadius` and ends at `Max - blurRadius`, leaving a `blurRadius`-wide border unblurred. If the blur radius is large relative to the image size, significant portions of the edges receive no bloom.
- **Impact**: Visible edge darkening when bloom is enabled with a large radius.
- **Suggested Fix**: Handle edge pixels with a clamped kernel or extend the loop to the full range.

---

## Performance Optimization Opportunities

### [P-001] ECS Entities() Query Hot Path Allocation
- **Status**: Open
- **Location**: `pkg/engine/ecs/world.go:94-112`
- **Current Impact**: 1,200+ allocations/second from system queries. Each allocation grows the result slice via append, then sort allocates internally.
- **Optimization**: Pre-allocate result slice with estimated capacity. Consider caching query results per frame (invalidated on entity create/destroy).
- **Expected Improvement**: 50-80% reduction in query-related GC pressure.

### [P-002] Sort on Every Entities() Call
- **Status**: Open
- **Location**: `pkg/engine/ecs/world.go:110`
- **Current Impact**: `sort.Slice` runs O(n log n) on every component query for determinism. With 200 NPCs, each query sorts ~200 entities.
- **Optimization**: Maintain a sorted entity index that's updated on create/destroy, avoiding per-query sorts.
- **Expected Improvement**: ~70% reduction in Entities() CPU time for large entity counts.

### [P-003] FOV Ray Directions Recalculated Per Row
- **Status**: Open
- **Location**: `pkg/rendering/raycast/draw.go:133`
- **Current Impact**: `calculateFOVRayDirections()` is called once per row in the floor/ceiling loop (lines 132-139). The result depends only on `PlayerA` and `FOV`, which are constant for the entire frame.
- **Optimization**: Hoist the call above the loop—compute once, reuse for all rows.
- **Expected Improvement**: Removes ~360 redundant trig calls per frame (2 cos + 2 sin per call).

### [P-004] sendWorldState Copies All Player States Under Lock
- **Status**: Open
- **Location**: `pkg/network/server.go:376-388`
- **Current Impact**: The server holds the mutex while iterating and copying all player states for every world state send. With 32 players, this copies 32 EntityState structs while blocking all other operations.
- **Optimization**: Use a lock-free snapshot mechanism (copy-on-write) or a ring buffer of pre-serialized states.
- **Expected Improvement**: Reduced lock contention under high player counts.

### [P-005] Redundant Map Lookups in ECS Component Access
- **Status**: Open
- **Location**: `pkg/engine/ecs/world.go:72-80`
- **Current Impact**: `GetComponent()` acquires RLock, looks up entity in map, looks up component in nested map, releases lock. For systems accessing multiple components on the same entity, this acquires/releases the lock N times.
- **Optimization**: Add a batch accessor: `GetComponents(e Entity, types ...string) (map[string]Component, bool)` that does one lock acquisition.
- **Expected Improvement**: ~30% reduction in lock overhead for multi-component systems.

### [P-006] Auto-Save Creates Full World Snapshot Under Unknown Locking
- **Status**: Open
- **Location**: `cmd/server/main.go:531`
- **Current Impact**: Auto-save every 5 minutes creates a snapshot of the entire world. If this blocks the tick loop, it introduces a latency spike.
- **Optimization**: Perform snapshot in a background goroutine with a copy-on-write mechanism.
- **Expected Improvement**: Eliminates save-related latency spikes.

### [P-007] Particle Glow Uses sqrt for Distance
- **Status**: Open
- **Location**: `pkg/rendering/particles/renderer.go:151`
- **Current Impact**: Each glow pixel computes `math.Sqrt(dx*dx + dy*dy)` for distance. For a glow radius of R, this is R² sqrt operations per particle per frame.
- **Optimization**: Use squared distance for comparisons (`dx*dx + dy*dy < radius*radius`), avoiding sqrt entirely.
- **Expected Improvement**: 2-5x speedup for particle glow rendering.

---

## Code Quality Observations

### [Q-001] 20 Component Structs Missing Type() Method
- **Status**: Open
- **Location**: `pkg/engine/components/definitions.go` — `ActivityLocation` (line 151), `CityEventEffects` (line 1356), `CityEventRequirements` (line 1372), `DialogConsequences` (line 1492), `DialogExchange` (line 1526), `DialogMemoryEvent` (line 1561), `DialogOption` (line 1458), `DialogPromise` (line 1573), `DialogRequirements` (line 1476), `EquipmentSlot` (line 750), `FactionMemberInfo` (line 37), `GossipItem` (line 1092), `MemoryEvent` (line 912), `OccupationItem` (line 1427), `Point2D` (line 91), `Relationship` (line 941), `Room` (line 995), `SkillSchool` (line 446), `VehicleArchetype` (line 274), `Waypoint` (line 157)
- **Issue**: These structs are in the components package but don't implement the `Component` interface. If mistakenly passed to `AddComponent()`, the `c.Type()` call at `world.go:67` will panic on nil method.
- **Suggestion**: Either add `Type()` methods or move them to a separate `types` sub-package to clarify they are not ECS components.

### [Q-002] Server System Count Hardcoded as String Literal
- **Status**: Open
- **Location**: `cmd/server/main.go:522`
- **Issue**: `log.Printf("registered %d server systems", 60)` hardcodes the count as 60 instead of counting actual registrations. If systems are added or removed, this log message becomes inaccurate.
- **Suggestion**: Track registration count dynamically: `count := len(world.systems)` or increment a counter.

### [Q-003] Unused Variable Suppression Pattern
- **Status**: Open
- **Location**: `cmd/server/main.go:102-105,456`, `cmd/server/init.go:344,401,420`
- **Issue**: Seven instances of `_ = variable` to suppress "unused variable" warnings. These represent initialized but unwired features (housing, PvP, dialog, companion managers; faction arcs; puzzles; zone seeds).
- **Suggestion**: Remove unused initializations or wire them into the game loop. Each represents a dangling feature per the project's integration mandate.

### [Q-004] Client Main File Exceeds 2,250 Lines
- **Status**: Open
- **Location**: `cmd/client/main.go` (2,259 lines)
- **Issue**: A single file containing game initialization, Update loop, Draw loop, gameplay logic, UI rendering, input handling, NPC rendering, vehicle logic, combat feedback, minimap, HUD, and debug display. This makes navigation and maintenance difficult.
- **Suggestion**: Extract logical sections into separate files: `cmd/client/update.go`, `cmd/client/draw.go`, `cmd/client/hud.go`, `cmd/client/combat.go`, etc.

### [Q-005] Inconsistent Error Handling in AddComponent Callers
- **Status**: Open
- **Location**: `cmd/client/sync.go:164,170`
- **Issue**: `_ = s.world.AddComponent(localEntity, ...)` discards the error return from `AddComponent`. If the entity doesn't exist, the component is silently not added.
- **Suggestion**: Check and handle the error, or at minimum log it.

### [Q-006] Helper Functions for Trivial Math Operations
- **Status**: Open
- **Location**: `pkg/network/server.go:364-372`
- **Issue**: `cos64()` and `sin64()` are single-line wrappers around `math.Cos()` and `math.Sin()` that add indirection without any value. They accept and return `float64`, which is what `math.Cos/Sin` already do.
- **Suggestion**: Call `math.Cos()` and `math.Sin()` directly.

### [Q-007] Magic Numbers in Movement and Physics
- **Status**: Open
- **Location**: `cmd/client/main.go:49-51,138,411-413,483`, `cmd/server/main.go:342-343`
- **Issue**: Numerous magic numbers (3.0, 2.0, 0.3, 1.5708, 0.785398, 60.0, etc.) used throughout movement, physics, and animation code. While some are defined as constants (lines 49-51), others appear inline.
- **Suggestion**: Define all physics/animation constants in a single constants block or config values.

### [Q-008] Computed Values in Systems Discarded with _ Assignment
- **Status**: Open
- **Location**: `pkg/engine/systems/npc_occupation.go:191,305`, `pkg/engine/systems/audio.go:179`
- **Issue**: Computed values like `occ.GoldPerHour * (occ.TaskDuration / 3600.0) * ...` and `source.Volume * attenuation` are calculated then discarded with `_ =`. This wastes CPU and confuses readers about the computation's purpose.
- **Suggestion**: Remove the computations or use the results. These appear to be placeholders for future functionality.

### [Q-009] Non-Component Helper Structs Mixed with Components
- **Status**: Open
- **Location**: `pkg/engine/components/definitions.go`
- **Issue**: The definitions file contains 85 structs, of which 20 are helper/nested types (Point2D, Waypoint, EquipmentSlot, etc.) mixed with 65 actual Component types. No organizational separation distinguishes them.
- **Suggestion**: Use comments or separate files to group components vs. helper types (e.g., `components_helpers.go`).

---

## Recommendations by Priority

### 1. Immediate Action Required
- [ ] **[C-001]**: Fix division by zero in floor rendering (start loop at `halfHeight + 1`)
- [ ] **[C-002]**: Add determinant guard in billboard camera transform
- [ ] **[C-003]**: Add mutex lock to `RegisterSystem()`
- [ ] **[C-004]**: Guard vignette falloff denominator
- [ ] **[C-005]**: Track sendWorldState goroutines in WaitGroup and set write deadlines

### 2. High Priority (Next Sprint)
- [ ] **[H-001]**: Add ok-check to all 40+ GetComponent calls across 11 system files
- [ ] **[H-002]**: Add read/write deadlines to all network connections
- [ ] **[H-003]**: Wire or remove unused managers (housing, PvP, dialog, companion)
- [ ] **[H-006]**: Clamp subtitle opacity to [0, 1]
- [ ] **[H-008]**: Add configuration validation after unmarshal

### 3. Medium Priority (Backlog)
- [ ] **[M-001]**: Pre-allocate Entities() result slice
- [ ] **[M-003]**: Validate TickRate > 0 before division
- [ ] **[M-004]**: Fix Z-fighting comparison in billboard rendering
- [ ] **[M-006]**: Implement or remove empty RenderSystem stub
- [ ] **[M-007]**: Replace magic number 1.5708 with math.Pi/2
- [ ] **[P-003]**: Hoist FOV ray direction calculation out of floor loop

### 4. Technical Debt
- [ ] **[Q-001]**: Clarify non-Component helper structs (20 types missing Type())
- [ ] **[Q-003]**: Eliminate `_ = variable` suppression pattern (7 instances)
- [ ] **[Q-004]**: Split 2,259-line client main.go into logical modules
- [ ] **[Q-008]**: Remove or use computed-then-discarded values

---

## Testing Recommendations

### Critical Test Scenarios
1. **Floor rendering at horizon**: Test `calculateRowDistance(halfHeight, halfHeight)` — should not produce Inf/NaN
2. **Billboard transform with zero direction vector**: Test `TransformEntityToScreen` when camera direction is (0,0)
3. **Concurrent RegisterSystem + Update**: Run with `-race` flag under concurrent system registration
4. **Server tick with TickRate=0**: Ensure config validation catches this before division
5. **Network timeout**: Connect a client that never sends data—verify server doesn't leak goroutines
6. **Entity destruction during iteration**: Destroy entity in one system while another system holds its ID from a query

### Input Edge Cases to Validate
- Window resize to 0×0 or 1×1 dimensions
- Opacity values of -0.1, 0.0, 1.0, 1.1 for subtitle system
- World maps with 0 rows, 1 row, or jagged row lengths
- Player at exact map boundaries (x=0, y=0, x=mapWidth-1, y=mapHeight-1)
- FOV set to 0 or π (180°)

### Performance Benchmarks to Establish
- `BenchmarkEntitiesQuery` with 100, 1000, 10000 entities
- `BenchmarkFloorCeiling` at 1280×720 and 1920×1080
- `BenchmarkSendWorldState` with 1, 10, 32 players
- `BenchmarkGetComponent` for single vs. batch access patterns
- `BenchmarkParticleRender` with 100, 500, 1000 active particles

---

## Audit Methodology Notes

### Analysis Approach
- **Static analysis only**: All findings based on source code inspection without execution
- **Systematic package-by-package review**: Every .go file in the repository examined
- **Line-number verified**: All critical and high-priority issues verified against actual file contents
- **Cross-reference analysis**: Traced data flow from system registration through Update/Draw loops

### Areas Covered
- All 169 non-test Go source files across 28 packages
- Entry points: `cmd/client/main.go` (2,259 lines), `cmd/server/main.go` (793 lines)
- Core ECS framework: `pkg/engine/ecs/world.go` (151 lines)
- All 56 system implementations in `pkg/engine/systems/`
- Complete rendering pipeline: raycast, texture, postprocess, particles, lighting, subtitles
- Network server and client code
- Procedural generation (city, dungeon, noise, adapters)
- Audio synthesis and playback
- World management (chunk, housing, PvP, persistence)
- Configuration loading and dialog system

### Areas Not Covered
- Test file correctness (89 test files not audited for testing quality)
- Build tag variations (`noebiten` vs default)
- Third-party dependency vulnerabilities (go.mod dependencies not audited)
- Runtime behavior under actual gameplay (requires execution)
- WASM build-specific issues

### Assumptions
- WorldMap is rectangular (uniform row lengths)—if jagged maps are intended, H-005 becomes critical
- Systems are registered only at startup (not dynamically)—if dynamic registration is intended, C-003 is more urgent
- Floor rendering loop starting at `y = halfHeight` is a bug, not an intentional "horizon row" feature
- The 20 Type()-less structs in components are helper types, not intended as ECS components

### Limitations of Static Analysis
- Cannot confirm rendering artifacts without visual inspection
- Cannot measure actual FPS impact of performance issues
- Race conditions identified may not manifest under typical single-threaded startup patterns
- Float precision issues may be masked by clamping in downstream code

---

## Positive Observations

### Well-Implemented Patterns

1. **Deterministic Procedural Generation**: All generators properly use `rand.New(rand.NewSource(seed))` with no global rand usage detected. Noise functions take explicit seeds. City and dungeon generators are verified deterministic via tests.

2. **ECS Core Design**: The World/Entity/Component/System architecture is clean and minimal (151 lines). Component storage uses string-keyed maps for flexibility. Deterministic iteration via sorted entity IDs is a good choice for reproducibility.

3. **Thread Safety in World Packages**: `pkg/world/housing/`, `pkg/world/pvp/`, and `pkg/dialog/` all use `sync.RWMutex` correctly—`RLock` for reads, `Lock` for writes, with proper defer patterns.

4. **Input Management Best Practices**: `pkg/input/rebind.go` demonstrates excellent concurrency patterns—locks are released before calling external listeners to prevent deadlocks, and state is copied while locked then used after unlock.

5. **Network Interface Usage**: The network package correctly uses `net.Conn` and `net.Listener` interfaces instead of concrete types like `*net.TCPConn`, enabling testability and transport flexibility.

6. **Genre Parameterization**: All generators accept a genre string and route to genre-specific content. The city generator has 5 complete genre profiles with distinct naming, districts, and housing types.

7. **Responsive Layout**: `Layout()` returns `outsideWidth, outsideHeight` (line 1822-1824), properly adapting to window size rather than returning hardcoded dimensions.

8. **Bounds Checking in Renderer**: `SetPixel()` performs bounds checking before framebuffer access (line 244-246). `GetZBufferAt()` validates array bounds (line 748-752). Floor/ceiling rendering checks ceiling Y bounds (line 189).

9. **Server Shutdown Handling**: The server uses `os.Signal` notification with `SIGINT`/`SIGTERM`, graceful `ticker.Stop()` with defers, and `WaitGroup`-tracked goroutines for clean shutdown (with the noted exception of C-005).

10. **Configuration Flexibility**: Viper-based config with YAML file, environment variable override (`WYRM_` prefix), and comprehensive defaults covers the full configuration lifecycle.
