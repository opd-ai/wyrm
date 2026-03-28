# Implementation Plan: Phase 1 Foundation Completion

## Project Context
- **What it does**: Wyrm is a 100% procedurally generated first-person open-world RPG built in Go on Ebitengine, generating all content (terrain, NPCs, quests, audio, textures) from a deterministic seed with zero external assets.
- **Current goal**: Complete Phase 1 Foundation — fully functional ECS pipeline, V-Series generator integration, player entity with movement, and raycaster connected to world data.
- **Estimated Scope**: **Medium** (7 functions above complexity threshold 9.0, 1.6% duplication, 85% doc coverage)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|----------------|---------------------|
| ECS Framework functional | ✅ Achieved | No |
| 11 Systems registered & executing | ⚠️ Partial (5/11 are stubs) | **Yes** |
| V-Series dependency imported | ❌ Missing | **Yes** |
| Player entity with movement | ❌ Missing | **Yes** |
| Raycaster connected to chunks | ❌ Missing | **Yes** |
| City generator called at runtime | ❌ Missing | **Yes** |
| Audio integrated with Ebitengine | ⚠️ Partial (samples generated, no playback) | **Yes** |
| Network protocol defined | ❌ Missing (echo-only) | **Yes** |
| High-latency tolerance (200-5000ms) | ❌ Missing | Deferred to Phase 5 |
| 60 FPS target | ⚠️ Untested | Verify after integration |

## Metrics Summary

- **Complexity hotspots on goal-critical paths**: 7 functions above 9.0 threshold
  - `castRay` (17.1) — raycaster core, highest complexity
  - `Draw` in raycast (11.9) — rendering path
  - `GenerateWithSeed` (10.6) — texture generation
  - `main` in server (10.1) — server initialization
  - 3 system `Update` methods (9.3) — ECS update loop
- **Duplication ratio**: 1.6% (25 duplicated lines)
  - Noise functions duplicated between `pkg/rendering/texture/` and `pkg/world/chunk/`
- **Documentation coverage**: 85.3% overall (100% functions, 77.8% types)
- **Package coupling**: `main` package has highest coupling (4.0) with 8 dependencies
- **Anti-patterns detected**:
  - 2 critical resource leaks (client main, network)
  - 2 goroutine leaks (network accept/handle without context)
  - 7 unused receivers (stub system Update methods)
  - 3 bare error returns without wrapping

## Implementation Steps

### Step 1: Extract Shared Noise Package
- **Deliverable**: New `pkg/procgen/noise/noise.go` with `Noise2D`, `HashToFloat`, `Smoothstep`, `Lerp` functions; update `pkg/rendering/texture/` and `pkg/world/chunk/` to import shared code.
- **Dependencies**: None (can start immediately)
- **Goal Impact**: Reduces duplication ratio; improves maintainability; prepares for V-Series integration
- **Acceptance**: Duplication ratio drops from 1.6% to <0.5%
- **Validation**: `go-stats-generator analyze . --skip-tests --format json 2>/dev/null | jq '.duplication.duplication_ratio < 0.005'`

### Step 2: Create Player Entity and Input Handling
- **Deliverable**: 
  - Add `pkg/engine/components/input.go` with `Input` component (MoveForward, MoveRight, Turn, etc.)
  - Add `pkg/engine/systems/input.go` with `InputSystem` that reads `ebiten.IsKeyPressed()` and modifies player Position
  - In `cmd/client/main.go`, create player entity with Position, Health, Input components
  - Update `Game.Update()` to process keyboard/mouse input before `world.Update()`
- **Dependencies**: None
- **Goal Impact**: Enables basic gameplay; unblocks raycaster integration (camera needs player position)
- **Acceptance**: WASD keys move player position; mouse look changes angle; position changes visible in debug output
- **Validation**: `go build ./cmd/client && echo "Build passed"`

### Step 3: Connect Raycaster to World Chunks
- **Deliverable**:
  - Add `SetWorldData(cm *chunk.ChunkManager, playerChunkX, playerChunkY int)` method to `Renderer`
  - Modify `Renderer` to read wall data from chunk heightmaps (threshold-based conversion)
  - Update `RenderSystem.Update()` to fetch player position and pass to renderer
  - Update `cmd/client/main.go` to wire ChunkManager to Renderer
- **Dependencies**: Step 2 (player entity must exist)
- **Goal Impact**: Critical—connects procedural world to visual output; completes rendering pipeline
- **Acceptance**: Terrain changes when world seed changes; chunk boundaries visible when walking
- **Validation**: Visual inspection + `go build ./cmd/client`

### Step 4: Implement FactionPoliticsSystem Logic
- **Deliverable**:
  - Add `pkg/engine/components/reputation.go` with `Reputation` component (map of faction→score)
  - Implement `FactionPoliticsSystem.Update()` with:
    - Faction relations map (ally/neutral/hostile)
    - Reputation decay/growth per tick
    - War/peace state transitions based on player actions
  - Add test `pkg/engine/systems/faction_test.go`
- **Dependencies**: None
- **Goal Impact**: Enables faction gameplay; unblocks crime system (crimes affect faction reputation)
- **Acceptance**: Killing faction member reduces reputation; reputation decays toward neutral over time
- **Validation**: `go test -v ./pkg/engine/systems/ -run TestFaction`

### Step 5: Implement CrimeSystem Logic
- **Deliverable**:
  - Add `Crime` component with WantedLevel (0-5), BountyAmount, LastCrimeTime
  - Add `Witness` tag component for NPCs that can report crimes
  - Implement `CrimeSystem.Update()` with:
    - Query for entities with Crime component
    - Line-of-sight check to witnesses (use spatial query)
    - Wanted level escalation/decay logic
  - Add test `pkg/engine/systems/crime_test.go`
- **Dependencies**: Step 4 (crime affects faction reputation)
- **Goal Impact**: Core RPG mechanic; enables law enforcement gameplay
- **Acceptance**: Crime within NPC LOS raises wanted level within 2 ticks; wanted level decays when out of sight
- **Validation**: `go test -v ./pkg/engine/systems/ -run TestCrime`

### Step 6: Implement EconomySystem Logic
- **Deliverable**:
  - Add `EconomyNode` component with PriceTable, Supply, Demand per item type
  - Implement `EconomySystem.Update()` with:
    - Supply/demand formula affecting prices
    - Transaction processing (buy/sell modifies supply)
    - Price normalization over time
  - Add test `pkg/engine/systems/economy_test.go`
- **Dependencies**: None (can run in parallel with Step 5)
- **Goal Impact**: Enables trading gameplay; shop prices become dynamic
- **Acceptance**: Selling 50 items to vendor reduces buy price by ≥10%
- **Validation**: `go test -v ./pkg/engine/systems/ -run TestEconomy`

### Step 7: Implement QuestSystem Logic
- **Deliverable**:
  - Add `Quest` component with ID, CurrentStage, Flags (map[string]bool), Completed
  - Implement `QuestSystem.Update()` with:
    - Stage condition checking
    - Flag setting on completion
    - Consequence application (modify world state)
  - Add test `pkg/engine/systems/quest_test.go`
- **Dependencies**: None (can run in parallel with Step 5-6)
- **Goal Impact**: Core RPG mechanic; enables quest-driven gameplay
- **Acceptance**: Quest advances stages when conditions met; flags persist in component state
- **Validation**: `go test -v ./pkg/engine/systems/ -run TestQuest`

### Step 8: Add World Clock and Weather Transitions
- **Deliverable**:
  - Add `WorldClock` component with Hour (0-23), Day, Season
  - Add `WorldClockSystem` that increments time based on dt
  - Update `NPCScheduleSystem` to read from WorldClock
  - Implement `WeatherSystem.Update()` with:
    - Genre-specific weather pools
    - Transition logic (change weather every N hours)
    - Weather effects on gameplay (rain reduces fire, etc.)
  - Add tests for both systems
- **Dependencies**: None
- **Goal Impact**: Enables day/night cycle, NPC schedules actually function, weather affects gameplay
- **Acceptance**: Hour advances over time; NPCs change activity at schedule transitions; weather changes
- **Validation**: `go test -v ./pkg/engine/systems/ -run TestWorldClock && go test -v ./pkg/engine/systems/ -run TestWeather`

### Step 9: Integrate Audio Engine with Ebitengine
- **Deliverable**:
  - Update `pkg/audio/audio.go` to create `ebiten.AudioContext` on init
  - Add `Play(sound Sound)` method that creates Ebitengine player from generated samples
  - Wire `AudioSystem` to trigger sounds on game events (footsteps, combat, ambient)
  - Add genre-specific SFX modifications per ROADMAP spec
- **Dependencies**: Step 2 (player movement triggers footstep sounds)
- **Goal Impact**: Game produces sound; completes audio pipeline
- **Acceptance**: Running client produces audible footstep sounds when moving
- **Validation**: Manual test—run client and hear audio

### Step 10: Define Network Message Protocol
- **Deliverable**:
  - Add `pkg/network/protocol.go` with message types:
    - `PlayerInput` (movement commands)
    - `WorldState` (entity positions, health, etc.)
    - `EntityUpdate` (delta updates)
    - `ChunkData` (terrain data)
  - Implement message serialization/deserialization
  - Update `Server.handleClient()` to dispatch by message type
  - Update `Client` to send PlayerInput and receive WorldState
- **Dependencies**: Step 2 (input system generates input commands)
- **Goal Impact**: Enables multiplayer; server-authoritative architecture functional
- **Acceptance**: Two clients connected to server see each other's position
- **Validation**: `go test -v ./pkg/network/ -run TestProtocol`

### Step 11: Call City Generator at Runtime
- **Deliverable**:
  - In `cmd/server/main.go`, call `city.Generate(cfg.World.Seed, cfg.Genre)` during world init
  - Create building entities at generated positions with Position components
  - Update ChunkManager to incorporate city building positions into terrain
- **Dependencies**: Step 3 (raycaster must be connected to see buildings)
- **Goal Impact**: Cities appear in world; completes basic world population
- **Acceptance**: Cities visible at deterministic locations based on seed; genre affects building style
- **Validation**: Visual inspection with different seeds and genres

### Step 12: Fix Resource Leaks and Anti-Patterns
- **Deliverable**:
  - Add `defer conn.Close()` in `cmd/client/main.go` line 65
  - Add `defer conn.Close()` in `pkg/network/network.go` line 147
  - Add `context.Context` parameter to `Server.Start()` and goroutines
  - Wrap bare error returns with `fmt.Errorf("context: %w", err)`
- **Dependencies**: None (can run in parallel)
- **Goal Impact**: Prevents resource leaks; enables graceful shutdown; improves error debugging
- **Acceptance**: Zero critical/high anti-patterns in go-stats-generator output
- **Validation**: `go-stats-generator analyze . --skip-tests --format json 2>/dev/null | jq '[.patterns.anti_patterns.performance_antipatterns[] | select(.severity == "critical" or .severity == "high")] | length == 0'`

### Step 13: Refactor castRay for Maintainability
- **Deliverable**:
  - Extract helper functions from `castRay` (complexity 17.1):
    - `calculateDeltaDist(rayDirX, rayDirY float64) (float64, float64)`
    - `calculateSideDist(posX, posY, rayDirX, rayDirY, deltaDistX, deltaDistY float64) (float64, float64, int, int)`
    - `ddaStep(sideDistX, sideDistY, deltaDistX, deltaDistY, stepX, stepY, mapX, mapY int) (float64, float64, int, int, int)`
  - Reduce `castRay` complexity to <10
- **Dependencies**: Step 3 (raycaster must be functional first to verify no regression)
- **Goal Impact**: Improves testability and maintainability of critical rendering code
- **Acceptance**: `castRay` complexity <10; all raycaster tests pass
- **Validation**: `go-stats-generator analyze . --skip-tests --format json 2>/dev/null | jq '.functions[] | select(.name == "castRay") | .complexity.overall < 10'`

---

## Dependency Graph

```
Step 1 (Noise) ────────────────────────────────────────────────────┐
                                                                    │
Step 2 (Player/Input) ──────┬──────────────────────────────────────┼──► Step 9 (Audio)
                            │                                       │
                            ▼                                       │
                    Step 3 (Raycaster-World) ──────────────────────┼──► Step 11 (City Runtime)
                            │                                       │           │
                            ▼                                       │           ▼
                    Step 13 (castRay Refactor)                      │    Step 11 (City Runtime)
                                                                    │
Step 4 (Faction) ──────────► Step 5 (Crime)                        │
                                                                    │
Step 6 (Economy) ────────────────────────────────────────(parallel)─┤
                                                                    │
Step 7 (Quest) ──────────────────────────────────────────(parallel)─┤
                                                                    │
Step 8 (Clock/Weather) ──────────────────────────────────(parallel)─┤
                                                                    │
Step 10 (Network Protocol) ───────────────────────────────(parallel)┤
                                                                    │
Step 12 (Fix Leaks) ─────────────────────────────────────(parallel)─┘
```

## Execution Order (Optimized for Parallel Work)

**Wave 1** (No dependencies, immediate start):
- Step 1: Extract Shared Noise Package
- Step 4: Implement FactionPoliticsSystem
- Step 6: Implement EconomySystem
- Step 7: Implement QuestSystem
- Step 8: Add World Clock and Weather
- Step 12: Fix Resource Leaks

**Wave 2** (After Wave 1 items complete):
- Step 2: Create Player Entity and Input Handling
- Step 5: Implement CrimeSystem (needs Step 4)
- Step 10: Define Network Protocol

**Wave 3** (After Wave 2 items complete):
- Step 3: Connect Raycaster to World Chunks
- Step 9: Integrate Audio Engine

**Wave 4** (After Wave 3 items complete):
- Step 11: Call City Generator at Runtime
- Step 13: Refactor castRay

---

## Success Criteria

Upon completion of all steps:
1. ✅ All 11 ECS systems have functional implementations (no empty Update methods)
2. ✅ Player can move with WASD/mouse in first-person view
3. ✅ Raycaster renders procedurally generated terrain from chunk system
4. ✅ Cities appear in world at deterministic positions
5. ✅ Audio plays during gameplay
6. ✅ Network protocol defined and functional for multiplayer
7. ✅ Zero critical/high anti-patterns
8. ✅ Duplication ratio <0.5%
9. ✅ All `go test ./...` passes

## Metrics Targets Post-Implementation

| Metric | Current | Target |
|--------|---------|--------|
| Functions >9.0 complexity | 7 | 5 |
| Duplication ratio | 1.6% | <0.5% |
| Stub systems (empty Update) | 5 | 0 |
| Critical anti-patterns | 2 | 0 |
| High anti-patterns | 5 | 0 |
| Test coverage | ~91% | ≥90% |

---

## Notes on Research Findings

### Ebitengine v2.9 Considerations
- Vector graphics API deprecated (`AppendVerticesAndIndicesForFilling/Stroke`); new APIs `vector.FillPath()`, `vector.StrokePath()` preferred
- Requires Go 1.24+ (project already on 1.24.5 ✅)
- Stencil-based rendering improves performance for complex shapes—consider for UI rendering

### ECS Best Practices (2026)
- Current architecture follows best practices: pure data components, all logic in systems
- Consider goroutine offloading for heavy generation (chunk pre-fetching) with channel communication
- Spatial indexing recommended for entity queries >100 entities (add in Phase 3)

### Competitive Context
- Wyrm's zero-asset philosophy is unique in the procedural RPG space
- Similar projects (Dwarf Fortress, Caves of Qud) use pre-authored assets
- 200-5000ms latency tolerance for Tor support is industry-leading for RPGs
