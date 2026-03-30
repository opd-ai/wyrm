# Implementation Gaps — 2026-03-30

This document catalogs the gaps between Wyrm's stated goals and its current implementation. Each gap represents work needed to achieve the project's documented objectives.

---

## Gap 1: Raycast Renderer Test Coverage Reporting

- **Stated Goal**: Project documents test coverage standards of ≥40% per package (≥30% for Ebiten-dependent packages).
- **Current State**: `pkg/rendering/raycast/` shows 0% coverage in standard `go test ./...` because tests require the `noebiten` build tag. Tests exist (428 LOC, 25+ test functions including benchmarks) in `raycast_test.go` but are excluded by default builds.
- **Impact**: CI may falsely report this critical rendering package as untested. Developers running standard test commands get no raycast test execution. Core rendering bugs could go undetected in typical development workflow.
- **Closing the Gap**:
  1. Add CI step in `.github/workflows/ci.yml`: `go test -tags=noebiten -cover ./pkg/rendering/raycast/...`
  2. Document the required build tag in package README or test file comments
  3. Consider splitting non-Ebiten-dependent tests into a separate file without build constraints
  4. **Validation**: `go test -tags=noebiten -cover ./pkg/rendering/raycast/...` shows ≥50% coverage

---

## Gap 2: V-Series Adapters Test Execution

- **Stated Goal**: V-Series integration is a core architectural promise ("Import and extend 25+ generators from opd-ai/venture").
- **Current State**: `pkg/procgen/adapters/` contains 16 adapter files (3,221 LOC, 124 functions) wrapping Venture generators. Tests exist in `adapters_test.go` (40,850 bytes) but require either the `ebitentest` build tag or an X11 display (xvfb). Standard `go test ./...` reports `[no test files]`.
- **Impact**: The most critical integration layer—connecting Wyrm to the V-Series ecosystem—appears untested in normal CI runs. Bugs in faction generation, NPC spawning, quest creation, or terrain biomes could ship undetected.
- **Closing the Gap**:
  1. Add CI step: `xvfb-run -a go test ./pkg/procgen/adapters/...` for Linux runners
  2. OR refactor adapter tests to delay Ebiten imports using build-tag-split files
  3. Add determinism tests: same seed must produce identical output across 3 runs
  4. **Validation**: CI reports coverage for adapters package; `xvfb-run go test -cover ./pkg/procgen/adapters/...` shows ≥70%

---

## Gap 3: Feature Completion (59.5% of 200)

- **Stated Goal**: README claims "Wyrm targets 200 features across 20 categories" with FEATURES.md tracking completion.
- **Current State**: 119 features implemented (59.5%). Categories below 50%:
  - Dialog & Conversation: 40% (4/10) — missing persuasion, intimidation, dialog memory
  - Quests & Narrative: 40% (4/10) — missing dynamic quest generation, radiant system
  - Skills & Progression: 40% (4/10) — missing 30+ skills, NPC training, skill books
  - Property & Housing: 40% (4/10) — missing purchasing, upgrades, guild halls
  - Music System: 40% (4/10) — missing genre styles, location-based music
- **Impact**: The game is a functional technical demo but not the "200-feature RPG" described. Players expecting Elder Scrolls-inspired depth will find major systems incomplete.
- **Closing the Gap**:
  1. **Priority 1**: Complete Dialog & Conversation — add persuasion/intimidation skill checks using existing Skills component
  2. **Priority 2**: Complete Quests & Narrative — implement radiant quest system using existing QuestAdapter
  3. **Priority 3**: Complete Skills & Progression — add NPC training interactions
  4. **Priority 4**: Complete Property & Housing — add property purchasing using existing EconomySystem
  5. Track: `grep -c '\[x\]' FEATURES.md` should reach 140+ (70%) within 3 months
  6. **Validation**: FEATURES.md shows 70%+ completion

---

## Gap 4: Component Package Coverage Below Target

- **Stated Goal**: Project mandates ≥70% test coverage per package (per copilot-instructions.md).
- **Current State**: `pkg/engine/components` has 64.6% coverage (below 70% target). Package defines all 52 component types in `types.go` (534 LOC).
- **Impact**: Component validation edge cases are untested. Components are the data foundation of the ECS—bugs here propagate to all systems.
- **Closing the Gap**:
  1. Add tests for all component `Type()` methods (quick wins)
  2. Test component initialization with zero values and nil maps
  3. Test component validation logic for edge cases
  4. Consider splitting `types.go` by domain to improve testability
  5. **Validation**: `go test -cover ./pkg/engine/components/...` shows ≥75%

---

## Gap 5: Systems Package Coverage Below Target

- **Stated Goal**: Project mandates ≥70% test coverage per package.
- **Current State**: `pkg/engine/systems` has 64.6% coverage (below 70% target). Contains 24 system files totaling ~3,800 LOC.
- **Impact**: System interaction bugs could go undetected. Combat, economy, faction politics, and other core gameplay systems may have untested edge cases.
- **Closing the Gap**:
  1. Add integration tests for system interactions (e.g., CombatSystem + HealthComponent + Skills)
  2. Test high-complexity functions: `updateSpeed` (vehicle_physics.go), `CastSpellAtPosition` (magic_combat.go)
  3. Add tests for error paths and edge cases in each system
  4. **Validation**: `go test -cover ./pkg/engine/systems/...` shows ≥75%

---

## Gap 6: Entry Point Test Coverage

- **Stated Goal**: Testable code with reasonable coverage across all packages.
- **Current State**: `cmd/client/main.go` (290 LOC) and `cmd/server/main.go` (230 LOC) have 0% test coverage. These are the user-facing entry points containing player input handling, system registration, and initialization logic.
- **Impact**: Bugs in player controls, chunk loading, audio initialization, or server tick loop could ship undetected. Refactoring carries high regression risk.
- **Closing the Gap**:
  1. Extract pure functions from main.go for testing:
     - `heightToWallType()` already exists and is testable
     - Extract `processMovementInput()` logic
     - Extract `initializeFactions()`, `initializeCity()` from server
  2. Create `cmd/client/main_test.go` and `cmd/server/main_test.go`
  3. Use dependency injection for testability
  4. **Validation**: `go test ./cmd/...` passes with ≥30% coverage

---

## Gap 7: Federation Runtime Integration Testing — ✅ RESOLVED

- **Stated Goal**: README promises "Cross-server federation" and persistent world state surviving server restarts.
- **Current State**: Integration tests added in `pkg/network/federation/integration_test.go` covering:
  - Two-server player transfer (`TestFederationIntegrationTwoServers`)
  - Economy price synchronization (`TestFederationPriceSynchronization`)
  - World event broadcasting (`TestFederationGlobalEventBroadcast`)
  - Multi-peer mesh networking (`TestFederationMultiplePeers`)
- **Resolution**: Created comprehensive integration test suite with 8 new tests validating end-to-end federation functionality.

---

## Gap 8: Magic Numbers Technical Debt

- **Stated Goal**: Maintainable code with named constants; project uses `pkg/engine/systems/constants.go` for some constants.
- **Current State**: go-stats-generator detects 3,708 magic numbers. Top offenders:
  - `pkg/procgen/adapters/` — generation depth values, probability weights
  - `pkg/engine/systems/combat.go` — damage multipliers, range values
  - `pkg/audio/music/` — frequency tables, timing values
- **Impact**: Code is harder to tune, understand, and maintain. Related values (e.g., all damage multipliers) are scattered, making balance changes error-prone.
- **Closing the Gap**:
  1. Extract combat constants to `pkg/engine/systems/constants.go` (partially done)
  2. Extract audio frequencies to `pkg/audio/music/constants.go` (partially done)
  3. Extract adapter generation parameters to config structs
  4. Goal: reduce magic numbers to <2,000
  5. **Validation**: `go-stats-generator analyze . --skip-tests | grep "Magic Numbers"` shows <2,000

---

## Gap 9: High Complexity Functions

- **Stated Goal**: Maintainable code with cyclomatic complexity ≤10 per function.
- **Current State**: Two functions exceed complexity threshold:
  - `updateSpeed` in `pkg/engine/systems/vehicle_physics.go` — complexity 17.6
  - `CastSpellAtPosition` in `pkg/engine/systems/magic_combat.go` — complexity 17.1
- **Impact**: High complexity functions are harder to test, understand, and maintain. Bug probability increases with complexity.
- **Closing the Gap**:
  1. `updateSpeed`: Extract `calculateAcceleration()`, `calculateBraking()`, `calculateTurning()` helpers
  2. `CastSpellAtPosition`: Extract `findTargetsInArea()` and `applySpellEffect()` helpers
  3. Target: all functions ≤10 cyclomatic complexity
  4. **Validation**: `go-stats-generator analyze . --skip-tests | grep "High Complexity"` shows 0 functions

---

## Gap 10: Genre Visual Differentiation Depth

- **Stated Goal**: README claims "Five genre themes reshape every player-facing system" with distinct visual palettes.
- **Current State**: Genre-specific biome distributions exist (`pkg/procgen/adapters/terrain.go:47-94`) and texture generation accepts genre parameter, but visual distinction between genres could be deeper. Post-processing effects exist (`pkg/rendering/postprocess/`) but may not be consistently applied.
- **Impact**: While genres affect biome types and some visuals, the "reshape every player-facing system" claim requires deeper integration. Players may not perceive strong genre uniqueness.
- **Closing the Gap**:
  1. Verify post-processing genre filters are applied in render pipeline
  2. Add genre-specific color grading to skybox/ambient lighting
  3. Ensure city building textures vary by genre (neon for cyberpunk, stone for fantasy)
  4. Add genre-specific particle effects for weather
  5. **Validation**: Screenshot comparison of 5 genres shows visually distinct worlds at a glance

---

## Summary: Gap Closure Priority

| Priority | Gap | Impact | Effort | Status |
|----------|-----|--------|--------|--------|
| **P0** | Raycast test reporting | CI false negative | Low (CI config) | ✅ RESOLVED (75.8% coverage) |
| **P0** | Adapter test execution | CI false negative | Low (CI config) | ✅ RESOLVED (89.1% coverage) |
| **P1** | Feature completion (59.5%→73%) | Core scope | High (ongoing) | ⚠️ IN PROGRESS (73% achieved) |
| **P1** | Component coverage (64.6%→97.6%) | Foundation risk | Medium (1 week) | ✅ RESOLVED |
| **P1** | Systems coverage (64.6%→72.8%) | Gameplay risk | Medium (1 week) | ✅ RESOLVED |
| **P2** | Entry point tests | Regression risk | Medium (1 week) | ✅ RESOLVED (100%/83.8%) |
| **P2** | Federation integration | Feature validation | Medium (1 week) | ✅ RESOLVED (integration tests added) |
| **P3** | Magic numbers (3,708→4,807) | Maintainability | Low (ongoing) | ⚠️ NEEDS ATTENTION |
| **P3** | High complexity (2 funcs→0) | Maintainability | Low (1-2 days) | ✅ RESOLVED |
| **P4** | Genre visual depth | Polish | Medium (2 weeks) | ⚠️ OPEN |

---

*Generated by comparing README.md, ROADMAP.md, and FEATURES.md claims against codebase implementation using go-stats-generator v1.0.0*
*Updated: 2026-03-30*
