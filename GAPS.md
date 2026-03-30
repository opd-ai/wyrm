# Implementation Gaps — 2026-03-30

This document catalogs the gaps between Wyrm's stated goals and its current implementation. Each gap represents work needed to achieve the project's documented objectives.

---

## Gap 1: Feature Completion (94% → 100%)

- **Stated Goal**: README claims "Wyrm targets 200 features across 20 categories" with full implementation.
- **Current State**: 188/200 features implemented (94%). 12 features remain unimplemented according to FEATURES.md:
  - Weather & Environment: Indoor/outdoor detection, Extreme weather events
  - Audio System: UI sounds, Ambient sound mixing
  - Music System: Menu music
  - Rendering & Graphics: Sprite rendering, Particle effects, Lighting system, Skybox rendering
  - Networking & Multiplayer: Party system, Player trading
  - Technical & Accessibility: Subtitle system, Key rebinding
- **Impact**: The game delivers 94% of promised features, but key user-facing polish features (sprite rendering for NPCs, particle effects for combat/weather, lighting for atmosphere) are missing. Party and trading features impact multiplayer experience.
- **Closing the Gap**:
  1. **Priority 1**: Party system — networking infrastructure exists; add party creation, member tracking, shared quest state
  2. **Priority 2**: Player trading — economy system exists; add trade request protocol, item transfer validation
  3. **Priority 3**: Indoor/outdoor detection — housing system has room definitions; add `IsIndoors` component check
  4. **Priority 4**: Sprite rendering — integrate with raycaster for NPC/item visibility (see [SPRITE_PLAN.md](SPRITE_PLAN.md) for complete design spec)
  5. **Priority 5**: Lighting system — extend raycaster with light source attenuation
  6. **Priority 6**: Particle effects — add particle emitter system for weather, combat
  7. **Priority 7**: Key rebinding — extend config system with input mapping
  8. **Priority 8**: Remaining audio/visual polish (UI sounds, ambient mixing, menu music, skybox, subtitles, extreme weather)
  - **Validation**: `grep -c '\[x\]' FEATURES.md` returns 200

---

## Gap 2: Music System Test Coverage (62.9% → 75%)

- **Stated Goal**: Project targets ≥70% test coverage per package per copilot-instructions.md.
- **Current State**: `pkg/audio/music/` has 62.9% coverage, the lowest of all tested packages. Package handles adaptive music generation, intensity state tracking, and genre-specific styles.
- **Impact**: Music is a key atmospheric feature. Low coverage means genre music differences, combat detection triggers, and state transitions may have untested edge cases. Bugs here directly impact player immersion.
- **Closing the Gap**:
  1. Add tests for `Update()` method state machine transitions (idle → exploration → combat → victory)
  2. Test genre-specific motif generation for all 5 genres
  3. Add edge case tests for music intensity boundary values
  4. Test combat detection trigger timing
  - **Validation**: `go test -cover ./pkg/audio/music/...` shows ≥75%

---

## Gap 3: Network Package Coverage Floor (70.2% → 80%)

- **Stated Goal**: ≥70% test coverage per package; additionally, README claims "200-5000ms latency tolerance."
- **Current State**: `pkg/network/` has 70.2% coverage — at the minimum threshold. This package implements client prediction, lag compensation, and Tor-mode adaptive behavior — all critical for the latency tolerance claim.
- **Impact**: The high-latency networking claim is the project's key differentiator. Insufficient test coverage risks subtle bugs in reconciliation, rewind accuracy, or Tor-mode transitions that would break the >800ms latency experience.
- **Closing the Gap**:
  1. Add integration tests for `ClientPredictor.Reconcile()` with simulated high-latency state updates
  2. Test `StateHistory.GetAtTime()` at exact boundary conditions (499ms, 500ms, 501ms)
  3. Test Tor-mode state transitions at RTT thresholds (799ms, 800ms, 801ms)
  4. Add fuzz tests for delta compression/decompression roundtrip
  - **Validation**: `go test -cover ./pkg/network/...` shows ≥80%

---

## Gap 4: Entry Point Test Coverage (0% → 30%)

- **Stated Goal**: Testable code with reasonable coverage across all packages.
- **Current State**: `cmd/client/main.go` and `cmd/server/main.go` report `[no test files]`. These entry points contain player input handling, system registration, chunk management initialization, audio startup, and server tick loop logic.
- **Impact**: Changes to initialization code (system registration order, config loading, audio setup) carry high regression risk. Federation initialization in server and chunk map building in client are untested critical paths.
- **Closing the Gap**:
  1. Create `cmd/client/main_noebiten_test.go` with `//go:build noebiten` tag
  2. Create `cmd/server/main_noebiten_test.go` with `//go:build noebiten` tag
  3. Test utility functions: `heightToWallType()`, `initializeFactions()`, `createDistrictEntity()`, `initializeWorldClock()`, `initializeFederation()`
  4. Use dependency injection to test `registerServerSystems()` and `registerClientSystems()`
  - **Validation**: `go test -tags=noebiten ./cmd/...` shows ≥30% coverage

---

## Gap 5: Build-Tag-Dependent Test Visibility

- **Stated Goal**: CI pipeline provides clear test status for all packages.
- **Current State**: Three packages require the `noebiten` build tag for tests to run:
  - `pkg/procgen/adapters/` — 89.2% coverage, but shows `[no test files]` without tag
  - `pkg/rendering/raycast/` — 94.6% coverage, but shows `[no test files]` without tag
  - `cmd/client/` — tests exist with tag, shows `[no test files]` without tag
  
  Developers running `go test ./...` see these critical packages as untested.
- **Impact**: False negative perception of test coverage. Developers may believe these packages are untested and skip them during refactoring, or add redundant tests.
- **Closing the Gap**:
  1. Add prominent comment in each package explaining build tag requirement
  2. Ensure CI workflow (`.github/workflows/ci.yml`) includes explicit steps:
     ```yaml
     - run: go test -tags=noebiten ./pkg/procgen/adapters/...
     - run: go test -tags=noebiten ./pkg/rendering/raycast/...
     ```
  3. Consider splitting tests into `*_ebiten_test.go` (requires Ebiten) and `*_test.go` (no dependency)
  - **Validation**: CI logs show explicit coverage numbers for all three packages

---

## Gap 6: High Complexity Functions (4 → 0)

- **Stated Goal**: Maintainable code with cyclomatic complexity ≤10 per function (per copilot-instructions.md).
- **Current State**: Four functions exceed the complexity threshold:
  - `applyToNode` in `pkg/engine/systems/economic_event.go` — complexity 22.0
  - `updateMount` in `pkg/engine/systems/` — complexity 17.9
  - `completeReading` in `pkg/engine/systems/skill_progression.go` — complexity 15.8
  - `generateSceneEvidence` in `pkg/engine/systems/evidence.go` — complexity 15.3
- **Impact**: High complexity functions are harder to test, understand, and maintain. Bug probability correlates with complexity. These functions handle economic events, mount state, skill progression, and crime evidence — all gameplay-critical systems.
- **Closing the Gap**:
  1. **applyToNode (22.0)**: Extract per-effect-type handlers: `applySupplyEffect()`, `applyDemandEffect()`, `applyPriceModifier()`, `applyTradeRestriction()`, etc.
  2. **updateMount (17.9)**: Extract `MountStateMachine` struct with `Transition()` method; handle mount/dismount/combat as explicit states
  3. **completeReading (15.8)**: Split into `validateBookCompletion()` and `applyBookSkillEffects()`
  4. **generateSceneEvidence (15.3)**: Extract `createWitnessEvidence()`, `createPhysicalEvidence()`, `createEnvironmentEvidence()`
  - **Validation**: `go-stats-generator analyze . --skip-tests | grep "High Complexity"` shows 0 functions >10

---

## Gap 7: Code Duplication in Vehicle System

- **Stated Goal**: DRY code with duplication ratio <3%.
- **Current State**: `pkg/engine/systems/vehicle.go` (3,256 LOC) contains 4 clone pairs of 10-12 lines each for vehicle state handling. Additional clones exist in `faction_rank.go` (3 exact 12-line clones) and `stealth.go` (8 renamed 6-7 line clones).
- **Impact**: Duplicated code means bug fixes must be applied in multiple places. Vehicle state handling is complex; divergence between clones could cause inconsistent behavior.
- **Closing the Gap**:
  1. **vehicle.go**: Extract `updateVehicleState()`, `handleVehicleTransition()`, `validateVehicleAction()` helpers
  2. **faction_rank.go**: Consolidate into `progressFactionRank(faction, rankType, amount)` with faction-type parameter
  3. **stealth.go**: Extract `calculateVisibilityModifier(factor, lightLevel, movement)` parameterized helper
  - **Validation**: `go-stats-generator analyze . --skip-tests | grep "Duplication Ratio"` shows <1.0%

---

## Gap 8: Incomplete Hazard System (2 TODOs)

- **Stated Goal**: Fully implemented environmental hazards per "Environmental hazards" feature in FEATURES.md.
- **Current State**: `pkg/engine/systems/hazard.go` contains two TODO comments:
  - `// TODO: Check if entity is indoors (shelter check)`
  - `// TODO: Get from WorldClock component`
  
  Indoor/shelter detection and world clock integration are incomplete.
- **Impact**: Weather hazards (radiation, cold, heat) may incorrectly apply to players inside buildings. Time-based hazard intensity may not function correctly.
- **Closing the Gap**:
  1. Implement indoor detection by checking entity proximity to `Housing` rooms or adding `IsIndoors` component
  2. Integrate with `WorldClockSystem` by querying world for `WorldClock` component to get current hour
  3. Add tests for indoor shelter protection and time-based hazard scaling
  - **Validation**: `grep -c TODO pkg/engine/systems/hazard.go` returns 0

---

## Gap 9: 60 FPS Performance Target Unverifiable

- **Stated Goal**: README claims "60 FPS at 1280×720" as a performance target.
- **Current State**: No performance benchmarks exist in the test suite. The raycaster has good test coverage (94.6%) but no benchmark tests. Average function complexity (3.7) suggests efficient code, but actual frame time is unmeasured.
- **Impact**: Cannot verify the 60 FPS claim without runtime profiling. Performance regressions would go undetected until manual testing.
- **Closing the Gap**:
  1. Add benchmark tests to `pkg/rendering/raycast/raycast_test.go`:
     ```go
     func BenchmarkRender(b *testing.B) { ... }
     func BenchmarkDDA(b *testing.B) { ... }
     ```
  2. Add benchmark to ECS world update:
     ```go
     func BenchmarkWorldUpdate(b *testing.B) { ... }
     ```
  3. Document expected performance characteristics in package comments
  4. Add CI step to run benchmarks and detect regressions
  - **Validation**: `go test -bench=. ./pkg/rendering/raycast/...` completes render in <16.67ms (60 FPS)

---

## Gap 10: Naming Convention Violations

- **Stated Goal**: Follow Go naming conventions per project style guide.
- **Current State**: go-stats-generator detected 23 identifier naming violations and 14 file naming violations:
  - Package-stuttering identifiers: `DialogManager` in `dialog`, `CompanionManager` in `companion`, `AmbientManager` in `ambient`
  - Generic file names: `util.go`, `constants.go`, `types.go`
- **Impact**: Package-stuttering names are idiomatic violations (`dialog.DialogManager` vs preferred `dialog.Manager`). Generic file names reduce code discoverability.
- **Closing the Gap**:
  1. Rename stuttering identifiers: `DialogManager` → `Manager`, `CompanionManager` → `Manager`, etc.
  2. Rename generic files: `cmd/client/util.go` → `client_helpers.go`, `cmd/server/util.go` → `server_init.go`
  3. Consider splitting large `types.go` files by domain
  - **Validation**: `go-stats-generator analyze . --skip-tests | grep "Identifier Violations"` shows ≤5

---

## Summary: Gap Closure Priority

| Priority | Gap | Impact | Effort | Current State |
|----------|-----|--------|--------|---------------|
| **P1** | Feature completion (94%→100%) | Core scope | High | 12 features remaining |
| **P1** | Music coverage (62.9%→75%) | Quality risk | Medium | Below threshold |
| **P1** | Network coverage (70.2%→80%) | Reliability risk | Medium | At floor |
| **P2** | Entry point tests (0%→30%) | Regression risk | Medium | No tests |
| **P2** | Build-tag visibility | Developer experience | Low | CI config needed |
| **P2** | High complexity (4→0) | Maintainability | Medium | 4 functions |
| **P3** | Code duplication | Maintainability | Low | 1.18% ratio |
| **P3** | Hazard TODOs | Feature completion | Low | 2 TODOs |
| **P3** | 60 FPS verification | Quality claim | Medium | No benchmarks |
| **P4** | Naming violations | Code style | Low | 23 identifiers |

---

## Next Steps

1. **Immediate (Week 1)**: Add tests to bring `pkg/audio/music/` and `pkg/network/` above coverage threshold
2. **Short-term (Week 2-3)**: Implement Party system and Player trading features (highest-impact missing features)
3. **Medium-term (Week 4-6)**: Refactor high-complexity functions and add entry point tests
4. **Ongoing**: Address naming violations and code duplication during normal development

---

*Generated by comparing README.md, ROADMAP.md, and FEATURES.md claims against codebase implementation using go-stats-generator v1.0.0*
