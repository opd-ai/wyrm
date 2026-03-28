# Implementation Gaps — 2026-03-28

This document catalogs the gaps between Wyrm's stated goals and its current implementation. Each gap represents work needed to achieve the project's documented objectives.

---

## Gap 1: ECS Systems Are Skeleton Implementations

- **Stated Goal**: ROADMAP.md Section 2 describes 11 key systems (WorldChunkSystem, NPCScheduleSystem, FactionPoliticsSystem, CrimeSystem, EconomySystem, CombatSystem, VehicleSystem, QuestSystem, WeatherSystem, RenderSystem, AudioSystem) that "contain all logic and operate on component queries each tick."
- **Current State**: All 11 system types exist in `pkg/engine/systems/systems.go` with correct interface signatures, but every `Update()` method body is `{}` (empty). Zero game logic executes.
- **Impact**: The game has no behavior. World chunks don't load, NPCs don't follow schedules, combat doesn't resolve, nothing renders beyond a solid color screen. The ECS architecture is complete but entirely inert.
- **Closing the Gap**:
  1. Prioritize implementing `RenderSystem.Update()` to draw the first-person view using the raycaster
  2. Implement `WorldChunkSystem.Update()` to load/unload chunks based on player position
  3. Implement `NPCScheduleSystem.Update()` to drive NPC daily routines
  4. Add unit tests for each system verifying component queries and state changes
  5. Register all systems in `cmd/client/main.go` and `cmd/server/main.go`

---

## Gap 2: Systems Not Registered in Game Loop

- **Stated Goal**: ROADMAP.md and copilot-instructions.md emphasize that "every feature MUST be fully wired into the runtime" and that systems must be registered via `world.RegisterSystem()`.
- **Current State**: Neither `cmd/client/main.go` nor `cmd/server/main.go` call `RegisterSystem()`. The ECS `World.Update(dt)` method is called each frame/tick, but it iterates over an empty system slice.
- **Impact**: Even if systems were implemented, they would never execute. The game loop is disconnected from the ECS architecture.
- **Closing the Gap**:
  1. In `cmd/client/main.go` after `ecs.NewWorld()`, add:
     ```go
     world.RegisterSystem(&systems.WorldChunkSystem{})
     world.RegisterSystem(&systems.RenderSystem{Renderer: renderer})
     world.RegisterSystem(&systems.AudioSystem{})
     ```
  2. In `cmd/server/main.go` after `ecs.NewWorld()`, add:
     ```go
     world.RegisterSystem(&systems.WorldChunkSystem{Manager: cm})
     world.RegisterSystem(&systems.NPCScheduleSystem{})
     world.RegisterSystem(&systems.FactionPoliticsSystem{})
     world.RegisterSystem(&systems.EconomySystem{})
     world.RegisterSystem(&systems.CombatSystem{})
     ```
  3. Modify system struct definitions to accept required dependencies (Manager, Renderer, etc.)

---

## Gap 3: No Test Coverage

- **Stated Goal**: ROADMAP.md Phase 1 acceptance criteria: "`go test ./pkg/engine/...` passes". Success criteria: "10,000 entities created/destroyed in <5 ms" with benchmark. copilot-instructions.md requires "≥40% per package".
- **Current State**: Zero `*_test.go` files exist. `go test ./...` reports `[no test files]` for all 12 packages.
- **Impact**: Cannot verify correctness of any implementation. Cannot detect regressions. Cannot validate performance targets. Phase 1 completion criteria cannot be met.
- **Closing the Gap**:
  1. Create `pkg/engine/ecs/ecs_test.go`:
     - `TestCreateEntity` — verify ID uniqueness
     - `TestAddComponent` — verify component attachment
     - `TestEntities` — verify query returns correct entities
     - `BenchmarkCreateDestroy` — verify 10k entities in <5ms
  2. Create `pkg/world/chunk/chunk_test.go`:
     - `TestChunkDeterminism` — same seed produces identical chunks
     - `TestChunkManagerCaching` — same coordinates return cached chunk
  3. Create `config/config_test.go`:
     - `TestLoadDefaults` — verify default values load correctly
  4. Target ≥40% coverage in each package

---

## Gap 4: V-Series Generator Integration Missing

- **Stated Goal**: ROADMAP.md Section 9 details importing 25+ generators from `opd-ai/venture` including terrain, entity, faction, quest, dialog, narrative, building, vehicle, magic, skills, recipe, class, companion, etc. "Wyrm treats it as a direct Go module dependency."
- **Current State**: `go.mod` contains no dependency on `opd-ai/venture`. All procedural content must be built from scratch rather than wrapping existing generators.
- **Impact**: Cannot leverage proven V-Series generators. Must reimplement terrain, entity, faction, quest, dialog, building, vehicle, magic, and skills generation independently. Dramatically increases development effort.
- **Closing the Gap**:
  1. Add dependency: `go get github.com/opd-ai/venture@latest`
  2. Create adapter packages in `pkg/procgen/adapters/`:
     - `terrain_adapter.go` wrapping Venture's terrain generator
     - `entity_adapter.go` wrapping Venture's entity generator
     - `faction_adapter.go` wrapping Venture's faction generator
  3. Update `WorldChunkSystem` to use terrain adapter for heightmap generation
  4. Pass `GenerationParams{GenreID: cfg.Genre, ...}` to all wrapped generators

---

## Gap 5: First-Person Raycaster Not Implemented

- **Stated Goal**: README.md lists `pkg/rendering/raycast/` as "First-person raycasting renderer". ROADMAP.md Phase 2 targets "60 fps at 1280×720 on reference hardware" and "genre changes wall palette".
- **Current State**: `pkg/rendering/raycast/raycast.go` exists but `Draw()` only calls `screen.Fill()` with a solid dark color. No raycasting algorithm, no wall rendering, no floor/ceiling, no depth perception.
- **Impact**: No first-person view. Players see only a solid color screen. The game cannot be visually experienced. Genre visual differentiation is impossible.
- **Closing the Gap**:
  1. Implement DDA (Digital Differential Analyzer) raycasting in `Draw()`:
     - Cast rays from player position through each screen column
     - Calculate wall intersection distances
     - Draw vertical strips with height inversely proportional to distance
  2. Reference Violence `pkg/raycaster` for proven implementation patterns
  3. Integrate with `pkg/rendering/texture/` for wall textures
  4. Add player Position and Direction to ECS world
  5. Benchmark to verify 60 FPS at target resolution

---

## Gap 6: Procedural Generators Produce No Content

- **Stated Goal**: ROADMAP.md Section 4 lists 16 generator systems. README.md claims "Procedural city generation" in `pkg/procgen/city/`. copilot-instructions.md requires "deterministic output: same seed → same result".
- **Current State**:
  - `city.Generate()` returns `{Name: "Unnamed City", Seed: seed}` ignoring seed and genre
  - `texture.Generate()` fills all pixels with identical grey
  - `chunk.NewChunk()` creates empty heightmap (all zeros)
  - `audio.Engine.Update()` is empty
- **Impact**: World is empty. No cities, no terrain variation, no procedural textures, no procedural audio. The "100% procedurally generated" claim is not achieved.
- **Closing the Gap**:
  1. **City generator** — Use seed to derive district count, positions, names; generate genre-appropriate building types
  2. **Texture generator** — Implement Perlin noise; apply genre color palettes
  3. **Chunk terrain** — Generate heightmap using noise functions seeded per-chunk
  4. **Audio engine** — Implement basic oscillator synthesis with Ebitengine audio
  5. Add determinism tests: 3 runs with same seed must produce byte-identical output

---

## Gap 7: Network Server Non-Functional

- **Stated Goal**: ROADMAP.md Section 5: "Server is authoritative for all world state. Clients send input commands; server applies, broadcasts delta states." Phase 1 acceptance: "Client connects to server, receives empty world state".
- **Current State**:
  - Server calls `net.Listen()` but has no `Accept()` loop — clients cannot connect
  - Server has no tick loop — `world.Update()` never called
  - Client has no `network.Client` instantiation — no connection attempt
  - No message protocol defined — no way to exchange data
- **Impact**: Multiplayer is completely non-functional. Even single-player cannot receive server state. Client and server are entirely disconnected codebases.
- **Closing the Gap**:
  1. Add connection accept loop in `network.Server`:
     ```go
     func (s *Server) acceptLoop() {
         for {
             conn, err := s.listener.Accept()
             // handle connection...
         }
     }
     ```
  2. Add tick loop in `cmd/server/main.go` using `time.Ticker`
  3. Define message protocol (protobuf or msgpack) for state sync
  4. Implement client connection in `cmd/client/main.go`
  5. Test with simulated latency to verify 200ms+ tolerance

---

## Gap 8: Genre Parameter Not Routed to Generators

- **Stated Goal**: ROADMAP.md Section 6 defines 5 genres (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic) with distinct visual palettes, audio styles, vocabulary, and content. Phase 1 acceptance: "Genre routing passes GenreID to all generators."
- **Current State**: `config.Genre` is loaded but only `audio.Engine` stores it. City generator, texture generator, and chunk manager ignore genre entirely.
- **Impact**: All 5 genres produce identical content. No visual or audio differentiation. The genre system is defined but completely unused.
- **Closing the Gap**:
  1. Add `Genre string` field to all generator structs and `NewX()` constructors
  2. Pass `cfg.Genre` when instantiating generators in main functions
  3. Implement genre-specific behavior:
     - City: genre-appropriate building styles and names
     - Texture: genre color palettes (warm gold/green for fantasy, neon pink/cyan for cyberpunk)
     - Audio: genre pitch/envelope modifications
  4. Add CI test asserting non-equal output across all 5 genres

---

## Gap 9: ChunkManager Instantiated but Unused

- **Stated Goal**: ROADMAP.md Phase 1: "Implement `pkg/world/chunk/` — 512×512-unit chunks with deterministic seed per coordinate". Phase 2: "Seamless chunk streaming".
- **Current State**: `cmd/server/main.go:24` creates ChunkManager but assigns to blank identifier `_ =`. The manager is immediately discarded. No system references it.
- **Impact**: Chunk loading/unloading cannot occur. World cannot stream. The chunk system exists but is orphaned from the game.
- **Closing the Gap**:
  1. Store ChunkManager: `cm := chunk.NewChunkManager(...)`
  2. Pass to WorldChunkSystem: `world.RegisterSystem(&systems.WorldChunkSystem{Manager: cm})`
  3. Add `Manager *chunk.ChunkManager` field to WorldChunkSystem struct
  4. Implement chunk loading in `WorldChunkSystem.Update()` based on player position

---

## Gap 10: Documented Features vs. Implementation Scale

- **Stated Goal**: ROADMAP.md Section 11 lists 200 specific features across 20 categories, marked with ★ for novel features and (V) for V-Series reuse.
- **Current State**: Of 200 features, approximately 0 are fully implemented:
  - 0/15 Open World features (no terrain, no biomes, no day/night, no weather)
  - 0/10 City features (generator produces empty city)
  - 0/13 NPC features (no NPCs exist)
  - 0/11 Quest features (QuestSystem is empty)
  - 0/16 Combat features (CombatSystem is empty)
  - 0/10 Vehicle features (VehicleSystem is empty)
  - 0/14 Multiplayer features (network non-functional)
  - etc.
- **Impact**: The project has ~0.5% implementation toward its 200-feature target. The gap between documentation ambition and code reality is approximately two orders of magnitude.
- **Closing the Gap**:
  1. Focus on Phase 1 Foundation before Phase 2+ features
  2. Achieve Phase 1 acceptance criteria:
     - `go test ./pkg/engine/...` passes
     - 10,000 entities created/destroyed in <5ms
     - Venture generator integration
     - Client connects to server, receives empty world state
     - Genre routing to all generators
  3. Defer 180+ features until foundation is solid
  4. Consider reducing ROADMAP scope to achievable milestones

---

## Summary: Implementation Completion

| Category | Stated Features | Implemented | Completion |
|----------|-----------------|-------------|------------|
| ECS Framework | Core + 6 components + 11 systems | Core + 6 components + 11 stubs | ~30% |
| Rendering | First-person raycaster | Solid color fill | ~5% |
| Procedural Generation | 16 generator types | 4 skeleton generators | ~5% |
| Networking | Authoritative multiplayer | TCP listener only | ~5% |
| Audio | Procedural synthesis | Empty engine | ~2% |
| Tests | ≥40% coverage | 0% coverage | 0% |
| V-Series Integration | 25+ generators | 0 imported | 0% |
| Feature Target | 200 features | ~0 complete | ~0% |

**Overall Project Completion: ~5-10%** of Phase 1 Foundation, **~1%** of full ROADMAP scope.

---

## Prioritized Gap Closure Roadmap

### Immediate (Week 1-2)
1. Add tests for ECS core — establish baseline correctness
2. Register systems in main functions — connect architecture
3. Implement basic RenderSystem — visual feedback loop

### Short-term (Week 3-4)
4. Implement WorldChunkSystem — terrain loading
5. Import Venture terrain generator — deterministic terrain
6. Implement network accept loop — client connection

### Medium-term (Week 5-8)
7. Implement raycaster — first-person view
8. Implement genre routing — content differentiation
9. Implement NPC schedules — world feels alive
10. Achieve Phase 1 acceptance criteria

### Long-term (Week 9+)
11. Implement remaining systems one by one
12. Add V-Series generator adapters incrementally
13. Progress through Phases 2-6 per ROADMAP
