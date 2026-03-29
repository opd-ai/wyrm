# Implementation Gaps — 2026-03-29

This document catalogs the gaps between Wyrm's stated goals and its current implementation. Each gap represents work needed to achieve the project's documented objectives.

---

## Gap 1: Feature Target (200 vs ~55)

- **Stated Goal**: README claims "Wyrm targets 200 features across 20 categories" and ROADMAP.md §11 lists all 200 features with completion tracking.
- **Current State**: Approximately 55 features are implemented (~28% of target). Core systems (ECS, chunk streaming, raycasting, basic networking) are functional, but many gameplay features remain unimplemented.
- **Impact**: The game is a functional technical demo but not the full RPG experience described. Key missing features include:
  - Multi-phase boss encounters with unique mechanics
  - Stealth system (sneak, pickpocket, backstab)
  - Player-joinable factions with rank progression
  - First-person furniture placement in housing
  - Guild halls with shared storage
  - Mount system and naval/flying vehicles
  - Crafting workbench minigames
  - Full melee/ranged/magic combat with timing-based blocking
- **Closing the Gap**:
  1. Continue Phase 2-6 implementation per ROADMAP.md milestones
  2. Create `FEATURES.md` checklist tracking all 200 features
  3. Prioritize combat, stealth, and faction mechanics for core gameplay loop
  4. **Validation**: `grep -c '\[x\]' FEATURES.md` shows increasing completion

---

## Gap 2: CI/CD Pipeline

- **Stated Goal**: Quality gates should prevent regressions from merging. The project's own documentation (copilot-instructions.md) mandates: "Build success: `go build ./cmd/...`; All tests pass: `go test ./...`; Race-free: `go test -race ./...`; Static analysis: `go vet ./...`".
- **Current State**: No `.github/workflows/` directory exists. No CI configuration files are present. Quality checks are manual.
- **Impact**: Regressions can merge undetected. Contributors lack feedback on whether their changes break the build or tests. No automated coverage reporting.
- **Closing the Gap**:
  1. Create `.github/workflows/ci.yml`:
     ```yaml
     name: CI
     on: [push, pull_request]
     jobs:
       test:
         runs-on: ubuntu-latest
         steps:
           - uses: actions/checkout@v4
           - uses: actions/setup-go@v5
             with:
               go-version: '1.24'
           - run: go build ./cmd/...
           - run: go test -race ./...
           - run: go test -tags=noebiten ./pkg/rendering/raycast/...
           - run: go vet ./...
     ```
  2. Enable branch protection requiring CI pass
  3. **Validation**: `gh run list --workflow=ci.yml` shows passing status

---

## Gap 3: Tor-Mode Adaptive Networking

- **Stated Goal**: README claims "200–5000 ms latency tolerance (designed for Tor-routed connections)". ROADMAP.md §5 specifies: "Tor-mode detection when RTT exceeds 800ms; increase prediction window to 1500ms; reduce input send rate to 10 Hz; enable aggressive interpolation with 300ms blend time."
- **Current State**: 
  - `LagCompensator.IsTorMode()` exists in `pkg/network/lagcomp.go:192` with 800ms threshold
  - RTT measurement exists via `ClientPredictor.updateRTT()` in `pkg/network/prediction.go:184`
  - However, the prediction window does NOT adapt when Tor-mode is detected
  - Input send rate is not configurable based on RTT
- **Impact**: High-latency players (Tor, satellite, poor connections) will experience degraded gameplay. The stated Tor compatibility is incomplete.
- **Closing the Gap**:
  1. Add `torModeActive` field to `ClientPredictor`
  2. In `Reconcile()`, check if `smoothedRTT > 800ms` and set `torModeActive = true`
  3. When `torModeActive`, increase pending input buffer from 128 to 256
  4. Add `GetRecommendedInputRate()` method returning 10 Hz when in Tor-mode, 30 Hz otherwise
  5. Add interpolation blend time field that increases to 300ms in Tor-mode
  6. **Validation**: `go test -v ./pkg/network/... -run TestTorMode` with simulated 2000ms latency

---

## Gap 4: Server Federation Runtime Integration

- **Stated Goal**: README claims "Cross-server federation" enabling player transfer between server instances. ROADMAP.md §5 specifies federation for scaling.
- **Current State**: 
  - `pkg/network/federation/` package exists with 90.4% test coverage
  - `FederationNode`, `CrossServerTransfer`, `GossipProtocol` are implemented
  - But `cmd/server/main.go` never instantiates `FederationNode`
  - No federation configuration in `config.yaml`
- **Impact**: Federation code is thoroughly tested but never used. Multi-server deployment is not possible.
- **Closing the Gap**:
  1. Add federation config section to `config.yaml`:
     ```yaml
     federation:
       enabled: false
       node_id: ""  # Auto-generated if empty
       peers: []    # List of peer server addresses
       gossip_interval: 5s
     ```
  2. In `cmd/server/main.go`, add conditional federation initialization:
     ```go
     if cfg.Federation.Enabled {
         node := federation.NewFederationNode(cfg.Federation.NodeID)
         for _, peer := range cfg.Federation.Peers {
             node.ConnectPeer(peer)
         }
         go node.StartGossip(cfg.Federation.GossipInterval)
     }
     ```
  3. Wire `CrossServerTransfer` into player disconnect handling
  4. **Validation**: Start two servers with federation enabled; player transfers between them

---

## Gap 5: V-Series Adapter Test Coverage

- **Stated Goal**: Project mandates ≥40% test coverage per package (per copilot-instructions.md Quality Standards section).
- **Current State**: `pkg/procgen/adapters/` shows 0% coverage with default `go test`, and only 11.4% with `xvfb-run go test -tags=ebitentest`. This is the critical V-Series integration layer with 16 adapters.
- **Impact**: V-Series integration—the foundation for terrain, entity, faction, quest, dialog, and other generators—is minimally tested. Bugs in adapters could cascade throughout the game.
- **Closing the Gap**:
  1. Add comprehensive tests to `pkg/procgen/adapters/adapters_test.go` covering:
     - All 16 adapters (EntityAdapter, FactionAdapter, QuestAdapter, DialogAdapter, TerrainAdapter, etc.)
     - Determinism verification (same seed → same output)
     - Error handling for invalid inputs (zero seed, empty genre)
  2. Use build tags to allow headless testing where possible
  3. Update CI to run: `xvfb-run go test -tags=ebitentest ./pkg/procgen/adapters/...`
  4. **Validation**: `go test -cover ./pkg/procgen/adapters/...` shows ≥70%

---

## Gap 6: Genre Depth in Terrain and NPCs

- **Stated Goal**: README claims "Five genre themes reshape every player-facing system." Specific visual palettes, faction types, and NPC behaviors are promised for each genre.
- **Current State**:
  - ✅ Vehicles have genre-specific archetypes (`GenreVehicleArchetypes` in `pkg/engine/components/types.go`)
  - ✅ Weather has genre-specific pools (`genreWeatherPools` in systems)
  - ✅ Textures use genre palettes (`pkg/rendering/texture/`)
  - ⚠️ Terrain biome distribution is not genre-differentiated
  - ⚠️ NPC archetypes and dialog vocabulary don't vary by genre
  - ⚠️ Quest templates are genre-agnostic
- **Impact**: Different genres feel similar aside from visual palette. The promised "distinct RPG experience" per genre is not fully achieved.
- **Closing the Gap**:
  1. `pkg/procgen/adapters/terrain.go`: Add genre-specific biome weights
     - Fantasy: 40% forest, 30% mountain, 20% plains, 10% lake
     - Sci-Fi: 40% crater, 30% tech-structure, 20% alien-flora, 10% mining-site
     - Horror: 40% swamp, 30% dead-forest, 20% fog-zone, 10% graveyard
     - Cyberpunk: 60% urban, 25% industrial, 15% neon-district
     - Post-Apoc: 50% wasteland, 30% ruins, 15% radiation-zone, 5% shanty
  2. `pkg/procgen/adapters/entity.go`: Add genre-specific NPC templates
  3. `pkg/procgen/adapters/quest.go`: Add genre-flavored objective text
  4. **Validation**: Visual comparison of 5 genre screenshots shows distinct biome distributions

---

## Gap 7: Combat Mechanics Depth

- **Stated Goal**: README claims "First-person melee, ranged, and magic combat with timing-based blocking" and "Multi-phase boss encounters with unique mechanics."
- **Current State**:
  - `CombatSystem.Update()` only clamps health to max (`pkg/engine/systems/registry.go:688-700`)
  - No attack input handling in client
  - No melee range detection or damage calculation
  - No timing-based blocking window
  - No ranged projectile system
  - No magic/ability system
- **Impact**: Combat—a core RPG mechanic—is essentially non-functional. Players cannot attack or be attacked.
- **Closing the Gap**:
  1. Add attack input handling in `cmd/client/main.go` (mouse click triggers attack)
  2. Add `Weapon` component with damage, range, attack speed
  3. Add `CombatState` component tracking last attack time, cooldown, blocking state
  4. Implement melee range detection using spatial queries (AABB or radius)
  5. Add damage calculation formula incorporating `Skills` component modifiers
  6. Add blocking window detection (e.g., 200ms window where block reduces damage by 75%)
  7. Add ranged projectile entities with trajectory and collision
  8. **Validation**: Player can attack NPC, NPC takes damage, blocking reduces damage

---

## Gap 8: Stealth System

- **Stated Goal**: README claims "Stealth system with sneak, pickpocket, and backstab mechanics."
- **Current State**: No `Stealth` component or `StealthSystem` exists. No sneak movement, detection mechanics, or stealth-related actions.
- **Impact**: Stealth gameplay—a promised feature—is completely missing.
- **Closing the Gap**:
  1. Add `Stealth` component: `{Visibility float64, Sneaking bool, DetectionRadius float64}`
  2. Add `StealthSystem` that:
     - Checks NPC sight cones against player position
     - Reduces visibility when sneaking (crouch key)
     - Triggers detection events when player enters NPC awareness
  3. Implement sneak movement (reduced speed, lower visibility)
  4. Implement backstab multiplier (2x damage when attacking unaware NPC)
  5. Implement pickpocket action with skill check against NPC awareness
  6. **Validation**: Player can sneak past NPCs; backstab deals bonus damage

---

## Gap 9: Registry.go File Size

- **Stated Goal**: Maintainable codebase with navigable file sizes. copilot-instructions.md suggests no file should exceed 200 lines for optimal maintainability.
- **Current State**: `pkg/engine/systems/registry.go` is 950 lines with 87 functions and 16 structs. All 11 ECS systems are in a single file.
- **Impact**: Difficult to navigate; hard to understand system boundaries; merge conflicts likely when multiple developers work on different systems.
- **Closing the Gap**:
  1. Split into per-system files:
     - `world_clock.go`: WorldClockSystem (~80 lines)
     - `npc_schedule.go`: NPCScheduleSystem (~100 lines)
     - `faction.go`: FactionPoliticsSystem (~150 lines)
     - `crime.go`: CrimeSystem (~120 lines)
     - `economy.go`: EconomySystem (~100 lines)
     - `quest.go`: QuestSystem (~120 lines)
     - `weather.go`: WeatherSystem (~80 lines)
     - `combat.go`: CombatSystem (~100 lines)
     - `vehicle.go`: VehicleSystem (~100 lines)
     - `audio.go`: AudioSystem (~100 lines)
     - `render.go`: RenderSystem (~100 lines)
     - `registry.go`: System registration helpers only (~50 lines)
  2. **Validation**: `wc -l pkg/engine/systems/*.go | awk '$1 > 200 {fail=1} END {exit fail}'`

---

## Summary: Implementation Completion

| Category | Stated Features | Implemented | Completion |
|----------|-----------------|-------------|------------|
| ECS Framework | Core + 11 systems | Core + 11 working systems | ~100% |
| Rendering | Raycaster + textures + post-processing | All implemented | ~90% |
| Procedural Generation | 16+ generator types | 16 adapters + local generators | ~85% |
| Networking | Authoritative multiplayer + lag comp + federation | Server + prediction + lag comp (no federation runtime) | ~75% |
| Audio | Synthesis + spatial + adaptive music | All implemented | ~85% |
| V-Series Integration | 25+ generators | 16 adapters | ~64% |
| World Systems | Chunks + housing + PvP + persistence | All implemented | ~90% |
| Gameplay | Combat + stealth + dialog + companions | Dialog + companions (no combat/stealth) | ~40% |
| CI/CD | Automated quality gates | None | 0% |
| Feature Target | 200 features | ~55 features | ~28% |

**Overall Project Completion**: ~85% of Phase 1 Foundation, ~28% of full ROADMAP scope.

---

## Prioritized Remediation Order

### Immediate (Blocks core functionality)
- [ ] Add CI/CD pipeline (Gap 2)
- [ ] Implement combat mechanics (Gap 7)

### Short-term (Core systems)
- [ ] Complete Tor-mode adaptive networking (Gap 3)
- [ ] Integrate federation at runtime (Gap 4)
- [ ] Split registry.go (Gap 9)

### Medium-term (Feature depth)
- [ ] Add stealth system (Gap 8)
- [ ] Deepen genre differentiation (Gap 6)
- [ ] Increase adapter test coverage (Gap 5)

### Long-term (Full scope)
- [ ] Implement remaining ~145 features per ROADMAP.md phases (Gap 1)
