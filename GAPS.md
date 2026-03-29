# Implementation Gaps — 2026-03-28

This document catalogs the gaps between Wyrm's stated goals and its current implementation. Each gap represents work needed to achieve the project's documented objectives.

---

## Gap 1: V-Series Generator Integration

- **Stated Goal**: ROADMAP.md Section 9 specifies importing 25+ generators from `opd-ai/venture` as "a direct Go module dependency" including terrain, entity, faction, quest, dialog, narrative, building, vehicle, magic, skills, recipe, class, companion, station, story, book, furniture, environment, item, legendary, puzzle, minigame, genre, and audit.
- **Current State**: `go.mod` has zero dependency on `opd-ai/venture` or any V-Series repository. All procedural content is built from scratch with minimal implementations.
- **Impact**: Dramatically increased development effort. The project cannot leverage proven, tested generators for terrain biomes, NPC entities, faction relationships, quest templates, dialog trees, building layouts, vehicle archetypes, magic systems, or skill trees. Every system must be reimplemented.
- **Closing the Gap**:
  1. Run `go get github.com/opd-ai/venture@latest`
  2. Create `pkg/procgen/adapters/` directory
  3. For each Venture generator, create an adapter that wraps the generator with Wyrm's genre parameters
  4. Update `WorldChunkSystem` to call terrain adapter
  5. Verify: `go test ./pkg/procgen/adapters/...`

---

## Gap 2: Stub System Implementations

- **Stated Goal**: README claims functional systems for faction politics, crime/law, dynamic economy, and branching quests with "persistent world-changing consequences."
- **Current State**: Five systems have empty or near-empty `Update()` methods:
  - `FactionPoliticsSystem` — queries entities but performs no faction logic
  - `CrimeSystem` — comment "Future: query witness entities" with no implementation
  - `EconomySystem` — comment "Future: update supply/demand" with no implementation
  - `QuestSystem` — comment "Future: check quest conditions" with no implementation
  - `WeatherSystem` — only initializes to "clear", never changes
- **Impact**: Core RPG gameplay loops do not function. Players cannot experience faction relationships, crime consequences, economic trading, quest progression, or weather effects.
- **Closing the Gap**:
  1. `FactionPoliticsSystem`: Add faction relations map, decay/grow logic, war/peace events
  2. `CrimeSystem`: Add Crime component with WantedLevel, witness LOS queries
  3. `EconomySystem`: Add EconomyNode component, supply/demand formula
  4. `QuestSystem`: Add Quest component with stages, flags, condition checking
  5. `WeatherSystem`: Define genre weather pools, add transition logic
  6. Verify: `go test -v ./pkg/engine/systems/`

---

## Gap 3: Raycaster-World Disconnect

- **Stated Goal**: README claims "Seamless infinite terrain via 512x512 chunk streaming" with first-person rendering.
- **Current State**: The raycaster creates a hardcoded 16x16 test map. The chunk system generates terrain heightmaps but the raycaster never reads them. Two parallel world representations exist without connection.
- **Impact**: Players see a static test level instead of the procedurally generated world. Chunk streaming is implemented but invisible.
- **Closing the Gap**:
  1. Add `SetWorldData(chunks [][]*chunk.Chunk)` method to Renderer
  2. Convert chunk heightmaps to wall grid with threshold logic
  3. Update client main loop to fetch chunks and pass to renderer
  4. Handle chunk boundary transitions
  5. Verify: Visual inspection that terrain changes match chunk seeds

---

## Gap 4: No Player Entity

- **Stated Goal**: README describes player character with "Position, Health, Faction, Schedule, Inventory, Vehicle" components.
- **Current State**: Neither client nor server creates a player entity. The ECS world is empty at startup. No entity has a Position component the renderer can track.
- **Impact**: The game renders from a fixed position. Players cannot move, have no health, belong to no faction, and cannot carry items.
- **Closing the Gap**:
  1. In server main() after world creation, call `world.CreateEntity()` and add components
  2. Store player entity ID and pass to RenderSystem
  3. In client, track local player entity for camera positioning
  4. Verify: `world.Entities("Position")` returns player ID

---

## Gap 5: No Input Handling

- **Stated Goal**: README describes "First-person melee, ranged, and magic combat" and player movement.
- **Current State**: `Game.Update()` only calls `world.Update(dt)`. No keyboard or mouse input is processed.
- **Impact**: Players cannot move, look around, attack, or interact. The game is a passive slideshow.
- **Closing the Gap**:
  1. Add input processing in `Game.Update()` using `ebiten.IsKeyPressed()`
  2. Create `InputSystem` to modify Position/Direction components
  3. Add mouse look using cursor position delta
  4. Verify: WASD keys change player position

---

## Gap 6: Audio Not Producing Sound

- **Stated Goal**: README claims "All sound effects synthesized procedurally" with "3D spatial audio."
- **Current State**: `pkg/audio/audio.go` generates sample arrays but never outputs sound. The Ebitengine audio context is never created.
- **Impact**: The game is silent. No footsteps, combat sounds, ambient audio, or music.
- **Closing the Gap**:
  1. Initialize Ebitengine audio context with `audio.NewContext(44100)`
  2. Create streaming player that reads from generated samples
  3. Wire AudioSystem to trigger sounds on game events
  4. Verify: Running client produces audible sound

---

## Gap 7: Network Protocol Incomplete

- **Stated Goal**: README claims "Authoritative server with client-side prediction and delta compression."
- **Current State**: The server echoes received bytes back unchanged. No message types are defined. No game state is synchronized.
- **Impact**: Multiplayer is non-functional. Clients cannot see other players or world state.
- **Closing the Gap**:
  1. Define message protocol with types for input, world state, and entity updates
  2. Implement message serialization
  3. Replace echo with message dispatch
  4. Add client-side prediction and server reconciliation
  5. Verify: Two clients see each other move

---

## Gap 8: High-Latency Networking Missing

- **Stated Goal**: ROADMAP.md specifies "200-5000ms latency tolerance (designed for Tor-routed connections)."
- **Current State**: No latency measurement, prediction buffers, jitter compensation, or Tor-mode detection exists.
- **Impact**: Game would be unplayable on high-latency connections. Stated Tor support does not exist.
- **Closing the Gap**:
  1. Implement RTT measurement via timestamp echo
  2. Add prediction buffer storing N frames of input
  3. Implement interpolation between server snapshots
  4. Add Tor-mode detection when RTT exceeds 800ms
  5. Verify: Game remains playable with simulated 2000ms latency

---

## Gap 9: World Clock Not Advancing

- **Stated Goal**: README claims "Day/night cycle, real-time weather, and seasonal changes" affecting "NPC schedules, shop hours, faction events."
- **Current State**: `NPCScheduleSystem.WorldHour` is never incremented. No world clock exists. Time does not pass.
- **Impact**: NPCs never change activity. Shops never open/close. Day never becomes night.
- **Closing the Gap**:
  1. Create WorldClock component with Hour, Day, Season fields
  2. Create WorldClockSystem that advances time based on dt
  3. NPCScheduleSystem reads hour from WorldClock
  4. RenderSystem adjusts lighting based on hour
  5. Verify: Log output shows hour advancing over time

---

## Gap 10: Feature Scope (200 vs ~15)

- **Stated Goal**: ROADMAP.md lists 200 features across 20 categories.
- **Current State**: Approximately 15 features have working implementations.
- **Impact**: The game is a technical demo, not the full RPG experience described.
- **Closing the Gap**: Continue Phase 2-6 implementation per ROADMAP.md milestones.

---

## Summary: Implementation Completion

| Category | Stated Features | Implemented | Completion |
|----------|-----------------|-------------|------------|
| ECS Framework | Core + 6 components + 11 systems | Core + 6 components + 6 working systems | ~70% |
| Rendering | First-person raycaster + world integration | DDA raycaster (hardcoded map) | ~40% |
| Procedural Generation | 16 generator types | 4 generators | ~25% |
| Networking | Authoritative multiplayer + lag comp | TCP echo server | ~15% |
| Audio | Procedural synthesis + spatial | Sample generation only | ~30% |
| V-Series Integration | 25+ generators | 0 imported | 0% |
| Feature Target | 200 features | ~15 features | ~8% |

**Overall Project Completion**: ~25% of Phase 1 Foundation, ~8% of full ROADMAP scope.

---

## Prioritized Remediation Order

### Immediate (Blocks basic gameplay)
- [x] Create player entity with Position component
- [x] Add input handling (WASD movement, mouse look)
- [x] Connect raycaster to chunk terrain data

### Short-term (Core systems)
- [x] Implement stub systems (Crime, Economy, Quest, Faction)
- [x] Add world clock for time progression
- [x] Define network message protocol

### Medium-term (Feature depth)
- [ ] Integrate V-Series generators
- [x] Wire audio engine to Ebitengine
- [x] Add weather transitions

### Long-term (Full scope)
- [ ] Implement remaining 185 features per ROADMAP.md phases
