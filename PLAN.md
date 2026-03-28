# Implementation Plan: Phase 1 Foundation Completion

## Project Context
- **What it does**: Wyrm is a 100% procedurally generated first-person open-world RPG built in Go 1.24+ on Ebitengine v2 with zero external assets
- **Current goal**: Complete Phase 1 Foundation — ECS integration, test coverage, V-Series imports, client-server connectivity, genre routing
- **Estimated Scope**: **Large** — 15+ critical integration gaps, 0% test coverage, 11 empty system implementations

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| ECS systems functional | ❌ All 11 `Update()` methods empty | Yes — Step 1-3 |
| Systems registered in game loop | ❌ Zero `RegisterSystem()` calls | Yes — Step 2 |
| `go test ./pkg/engine/...` passes | ❌ Zero test files exist | Yes — Step 4 |
| 10,000 entities in <5ms benchmark | ❌ No benchmarks exist | Yes — Step 4 |
| V-Series generator integration | ❌ Not imported in `go.mod` | Yes — Step 7 |
| Client connects to server | ❌ Server has no accept loop, no tick loop | Yes — Step 5-6 |
| Genre routing to all generators | ⚠️ Only audio uses genre | Yes — Step 8 |
| First-person raycaster rendering | ⚠️ Only fills solid color | Yes — Step 9 |
| Zero external assets | ✅ Achieved | No |
| Build succeeds | ✅ Achieved | No |

## Metrics Summary

**Source**: `go-stats-generator analyze . --skip-tests`

- **Complexity hotspots**: 0 functions above threshold (max is `Entities` at 8.0 cyclomatic)
- **Duplication ratio**: 0% (codebase too small for significant duplication)
- **Doc coverage**: 63.3% overall; **38.2% method coverage** (21 undocumented methods)
- **Package coupling**: `main` packages have highest coupling (7 deps) — expected for entry points
- **Test coverage**: **0%** (no `*_test.go` files)
- **Total LOC**: 189 lines (excluding tests)
- **Total functions/methods**: 47 (13 functions + 34 methods)

### Critical Observations

1. **All 11 ECS systems have empty `Update()` bodies** — the ECS architecture is sound but completely inert
2. **Zero systems registered** — `world.Update(dt)` iterates over empty slice
3. **ChunkManager discarded** — `_ = chunk.NewChunkManager(...)` in server
4. **No server tick loop** — server starts listener then blocks on signals
5. **No client connection** — client creates world but never connects to server
6. **Generators produce placeholder content** — city returns "Unnamed City", texture fills grey

## Implementation Steps

### Step 1: Implement Core ECS System Updates

- **Deliverable**: Implement working `Update()` methods for `WorldChunkSystem`, `RenderSystem`, and `NPCScheduleSystem` in `pkg/engine/systems/systems.go`
- **Dependencies**: None
- **Goal Impact**: Directly enables "ECS systems functional" — the foundational goal
- **Acceptance**: Systems query entities and modify component state; visible behavior change in client/server
- **Validation**: 
  ```bash
  go build ./... && ./client  # RenderSystem produces non-solid-color output
  ```

**Implementation Details**:
- `RenderSystem.Update()`: Call renderer with player camera state from world
- `WorldChunkSystem.Update()`: Query player position, load/unload chunks via ChunkManager
- `NPCScheduleSystem.Update()`: Query entities with Schedule component, update activity based on world clock

### Step 2: Register Systems in Game Loop

- **Deliverable**: Add `world.RegisterSystem()` calls in `cmd/client/main.go` and `cmd/server/main.go`
- **Dependencies**: Step 1 (systems must have implementations to register)
- **Goal Impact**: Connects ECS architecture to game loop — critical missing link
- **Acceptance**: `world.Update(dt)` invokes all registered system `Update()` methods each frame/tick
- **Validation**:
  ```bash
  # Add debug logging to systems, verify invocation
  go build ./cmd/client && ./client  # Should log system updates
  ```

**Implementation Details**:
- Client registers: `RenderSystem`, `AudioSystem`
- Server registers: `WorldChunkSystem`, `NPCScheduleSystem`, `FactionPoliticsSystem`, `CrimeSystem`, `EconomySystem`, `CombatSystem`, `VehicleSystem`, `QuestSystem`, `WeatherSystem`
- Modify system structs to accept dependencies (e.g., `WorldChunkSystem{Manager: cm}`)

### Step 3: Wire ChunkManager to WorldChunkSystem

- **Deliverable**: Store ChunkManager in server and pass to WorldChunkSystem; implement chunk loading in `WorldChunkSystem.Update()`
- **Dependencies**: Step 1, Step 2
- **Goal Impact**: Enables world terrain generation at runtime
- **Acceptance**: Chunks load/unload based on player position; ChunkManager.GetChunk() called during tick
- **Validation**:
  ```bash
  # Server should log chunk loading on player movement
  go build ./cmd/server && ./server  # Check logs for chunk operations
  ```

**Implementation Details**:
- Change `_ = chunk.NewChunkManager(...)` to `cm := chunk.NewChunkManager(...)`
- Add `Manager *chunk.ChunkManager` field to `WorldChunkSystem`
- Pass `cm` when registering: `world.RegisterSystem(&systems.WorldChunkSystem{Manager: cm})`

### Step 4: Create ECS Core Test Suite

- **Deliverable**: Create `pkg/engine/ecs/ecs_test.go` with comprehensive tests and benchmarks
- **Dependencies**: None (can be done in parallel with Steps 1-3)
- **Goal Impact**: Achieves Phase 1 criterion "`go test ./pkg/engine/...` passes" and "10,000 entities in <5ms"
- **Acceptance**: ≥40% coverage on `pkg/engine/ecs/`; benchmark proves 10k entity perf target
- **Validation**:
  ```bash
  go test -v -cover ./pkg/engine/ecs/...
  go test -bench=BenchmarkCreateDestroy -benchtime=3s ./pkg/engine/ecs/
  ```

**Test Cases Required**:
- `TestCreateEntity` — verify ID uniqueness, incrementing IDs
- `TestDestroyEntity` — verify entity removal
- `TestAddComponent` — verify component attachment, ErrEntityNotFound for missing entity
- `TestGetComponent` — verify retrieval, missing component returns false
- `TestEntities` — verify query returns correct entities, sorted order, multi-component query
- `TestRegisterSystem` — verify system added to slice
- `TestWorldUpdate` — verify all systems called with delta time
- `BenchmarkCreateDestroy` — 10,000 entities created/destroyed must complete in <5ms

### Step 5: Implement Server Accept Loop and Tick Loop

- **Deliverable**: Add connection accept goroutine to `pkg/network/network.go`; add tick loop to `cmd/server/main.go`
- **Dependencies**: Step 2 (systems must be registered for tick to be meaningful)
- **Goal Impact**: Enables client-server communication — Phase 1 criterion
- **Acceptance**: Server accepts TCP connections; world updates at configured tick rate
- **Validation**:
  ```bash
  go build ./cmd/server && ./server &
  nc localhost 7777  # Should connect without immediate disconnect
  # Server logs should show tick updates
  ```

**Implementation Details**:
- In `network.Server.Start()`: spawn `go s.acceptLoop()` after `net.Listen`
- Add `acceptLoop()` method that calls `s.listener.Accept()` in loop, stores connections
- In `cmd/server/main.go`: create `time.Ticker(time.Second / time.Duration(cfg.Server.TickRate))`
- Tick loop calls `world.Update(dt)` on each tick

### Step 6: Implement Client Network Connection

- **Deliverable**: Instantiate and connect `network.Client` in `cmd/client/main.go`
- **Dependencies**: Step 5 (server must accept connections)
- **Goal Impact**: Completes "client connects to server, receives empty world state"
- **Acceptance**: Client establishes TCP connection to server on startup
- **Validation**:
  ```bash
  ./server &  # Start server
  ./client    # Client should connect (verify via server logs)
  ```

**Implementation Details**:
- In `cmd/client/main.go` after config load: `client := network.NewClient(cfg.Server.Address)`
- Call `client.Connect()` before starting Ebiten
- Handle connection errors gracefully (offline mode for development)

### Step 7: Import V-Series Venture Generators

- **Deliverable**: Add `github.com/opd-ai/venture` dependency; create adapter packages in `pkg/procgen/adapters/`
- **Dependencies**: None (can be done in parallel)
- **Goal Impact**: Achieves "Venture generator integration" — massive code reuse benefit
- **Acceptance**: `go.mod` includes venture; terrain adapter produces genre-appropriate heightmaps
- **Validation**:
  ```bash
  go mod tidy
  go build ./pkg/procgen/adapters/...
  go test ./pkg/procgen/adapters/... -run TestTerrainDeterminism
  ```

**Implementation Details**:
- Run `go get github.com/opd-ai/venture@latest`
- Create `pkg/procgen/adapters/terrain.go` wrapping `venture/pkg/procgen/terrain`
- Create `pkg/procgen/adapters/entity.go` wrapping `venture/pkg/procgen/entity`
- All adapters accept `seed int64` and `genre string`, pass to Venture generators
- Add determinism test: 3 runs with same seed produce identical output

### Step 8: Implement Genre Routing to All Generators

- **Deliverable**: Pass `cfg.Genre` to city generator, texture generator, chunk manager, and all adapters
- **Dependencies**: Step 7 (adapters must exist to receive genre)
- **Goal Impact**: Achieves "Genre routing passes GenreID to all generators"
- **Acceptance**: Different genres produce visibly different output (palette, names, etc.)
- **Validation**:
  ```bash
  # Run with each genre, compare outputs
  WYRM_GENRE=fantasy ./client &
  WYRM_GENRE=cyberpunk ./client &
  # Visual inspection or automated screenshot diff
  ```

**Implementation Details**:
- Add `Genre string` field to `city.Generator`, `texture.Generator`, `chunk.ChunkManager`
- Modify all `NewX()` constructors to accept genre parameter
- Update `cmd/client/main.go` and `cmd/server/main.go` to pass `cfg.Genre`
- Implement genre-specific behavior:
  - City: genre-appropriate building names and district types
  - Texture: genre color palettes (warm gold/green for fantasy, neon pink/cyan for cyberpunk)
  - Audio: already implemented, verify functionality

### Step 9: Implement Basic Raycaster Rendering

- **Deliverable**: Implement DDA raycasting in `pkg/rendering/raycast/raycast.go` to render walls
- **Dependencies**: Step 1 (RenderSystem must call renderer), Step 8 (genre palettes)
- **Goal Impact**: Visual feedback for development; progress toward Phase 2 rendering target
- **Acceptance**: First-person view showing walls with depth perception; genre-appropriate colors
- **Validation**:
  ```bash
  ./client  # Visual inspection: should see 3D-like wall rendering, not solid color
  ```

**Implementation Details**:
- Add player `Position` and `Direction` to world
- Implement DDA algorithm in `Draw()`:
  - Cast rays from player position through each screen column
  - Calculate wall intersection distances
  - Draw vertical strips with height inversely proportional to distance
- Apply genre palette to wall colors
- Reference Violence `pkg/raycaster` for proven implementation patterns

### Step 10: Add Component and System Tests

- **Deliverable**: Create `pkg/engine/components/components_test.go` and `pkg/engine/systems/systems_test.go`
- **Dependencies**: Steps 1-3 (systems must have implementations to test)
- **Goal Impact**: Increases test coverage toward 40% target; validates system behavior
- **Acceptance**: ≥40% coverage on components and systems packages
- **Validation**:
  ```bash
  go test -v -cover ./pkg/engine/components/...
  go test -v -cover ./pkg/engine/systems/...
  ```

**Test Cases Required**:
- Components: `Type()` returns correct string for each component
- Systems: each system modifies expected components when entities exist
- Determinism: same seed + genre produces identical system behavior

### Step 11: Add Chunk and Network Tests

- **Deliverable**: Create `pkg/world/chunk/chunk_test.go` and `pkg/network/network_test.go`
- **Dependencies**: Steps 5-6 (network must be functional to test)
- **Goal Impact**: Validates critical infrastructure; enables CI quality gates
- **Acceptance**: ≥40% coverage on both packages; determinism verified
- **Validation**:
  ```bash
  go test -v -cover ./pkg/world/chunk/...
  go test -v -cover ./pkg/network/...
  ```

**Test Cases Required**:
- Chunk: `TestChunkDeterminism` — same seed produces identical chunks
- Chunk: `TestChunkManagerCaching` — same coordinates return cached chunk
- Network: `TestServerStartStop` — server starts and stops cleanly
- Network: `TestClientConnect` — client connects to running server

### Step 12: Documentation Coverage Improvement

- **Deliverable**: Add godoc comments to all 21 undocumented methods in systems and components
- **Dependencies**: None (can be done in parallel)
- **Goal Impact**: Improves method doc coverage from 38.2% to ~90%
- **Acceptance**: `go-stats-generator` reports ≥80% method documentation coverage
- **Validation**:
  ```bash
  go-stats-generator analyze . --skip-tests --format console --sections documentation | grep "Method Coverage"
  ```

**Methods to Document**:
- All 11 system `Update()` methods
- All 6 component `Type()` methods
- All 4 ECS core methods (`CreateEntity`, `DestroyEntity`, `AddComponent`, `GetComponent`)

---

## Phase 1 Completion Checklist

After completing all steps, verify Phase 1 acceptance criteria:

```bash
# 1. Tests pass
go test ./pkg/engine/...

# 2. 10,000 entities in <5ms
go test -bench=BenchmarkCreateDestroy ./pkg/engine/ecs/

# 3. Venture integration (verify import works)
go list -m github.com/opd-ai/venture

# 4. Client-server connection
./server &
./client  # Should connect

# 5. Genre routing
WYRM_GENRE=fantasy ./client &
WYRM_GENRE=cyberpunk ./client &
# Verify visual difference
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Venture API incompatibility | Pin to specific version; implement adapter layer with fallback stubs |
| Raycaster performance | Profile early; simplify to billboard rendering if needed |
| Network complexity | Start with minimal protocol; iterate after basic connectivity works |
| Test coverage target | Prioritize critical paths (ECS core, determinism); defer edge cases |

---

## Estimated Timeline

| Step | Effort | Parallel With |
|------|--------|---------------|
| 1. Core system implementations | 2-3 days | — |
| 2. System registration | 0.5 days | Step 4, 7, 12 |
| 3. ChunkManager wiring | 0.5 days | Step 4, 7, 12 |
| 4. ECS test suite | 1-2 days | Steps 1-3 |
| 5. Server accept/tick loop | 1 day | Step 4, 7 |
| 6. Client connection | 0.5 days | Step 4, 7 |
| 7. V-Series import | 1-2 days | Steps 1-6 |
| 8. Genre routing | 1 day | Step 10 |
| 9. Raycaster implementation | 2-3 days | Step 10-12 |
| 10. Component/system tests | 1 day | Step 9 |
| 11. Chunk/network tests | 1 day | Step 9 |
| 12. Documentation | 0.5 days | Any |

**Total Estimated Duration**: 4-6 weeks (with parallelization)

---

## Success Metrics

| Metric | Current | Target | Validation Command |
|--------|---------|--------|-------------------|
| Test coverage | 0% | ≥40% | `go test -cover ./...` |
| ECS benchmark | N/A | 10k entities <5ms | `go test -bench=. ./pkg/engine/ecs/` |
| Systems registered | 0 | 11 | `grep -r "RegisterSystem" cmd/` |
| Method doc coverage | 38.2% | ≥80% | `go-stats-generator analyze . --sections documentation` |
| Build success | ✅ | ✅ | `go build ./...` |
| Vet passes | ✅ | ✅ | `go vet ./...` |

---

*Generated by go-stats-generator analysis combined with ROADMAP.md Phase 1 goals*
*Report date: 2026-03-28*
