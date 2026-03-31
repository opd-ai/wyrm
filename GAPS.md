# Implementation Gaps — 2026-03-31

This document catalogs the gaps between Wyrm's stated goals and its current implementation. Each gap represents work needed to achieve full alignment between documentation claims and code reality.

---

## Gap 1: FEATURES.md Summary Table Severely Outdated

- **Stated Goal**: FEATURES.md header claims "192/200 implemented (96.0%)" with summary table showing 147/200 (73.5%).
- **Current State**: Individual feature checkboxes show 201/200 features marked `[x]`. The summary table at `FEATURES.md:318-340` is completely out of sync with the actual checkboxes.
- **Impact**: Critical documentation inconsistency misleads users and contributors about project completion status. The "Priority Implementation Order" section at the bottom recommends implementing features that are already implemented.
- **Closing the Gap**:
  1. Regenerate summary table by counting `[x]` checkboxes per category section
  2. Update header progress line to show 201/200 (100%+)
  3. Remove the obsolete "Priority Implementation Order" section
  4. Add automation script to keep summary in sync with checkboxes
  - **Validation**: `grep -c '\[x\]' FEATURES.md` returns 201, and header/summary match

---

## Gap 2: Entry Point Test Coverage (0% → 30%)

- **Stated Goal**: Testable code with reasonable coverage across all packages.
- **Current State**: `cmd/client/main.go` and `cmd/server/main.go` report `[no test files]` despite containing critical initialization logic:
  - Client: player entity creation, system registration, chunk map building, audio setup
  - Server: faction initialization, city/NPC spawning, federation setup, tick loop
- **Impact**: Changes to initialization code carry high regression risk. Federation initialization in server and chunk map building in client are untested critical paths.
- **Closing the Gap**:
  1. Create `cmd/client/main_noebiten_test.go` with `//go:build noebiten` tag
  2. Create `cmd/server/main_noebiten_test.go` with `//go:build noebiten` tag  
  3. Test utility functions: `heightToWallType()`, `createPlayerEntity()`, `registerClientSystems()`, `registerServerSystems()`, `initializeFactions()`, `initializeCity()`, `initializeWorldClock()`
  4. Use dependency injection to test system registration order
  - **Validation**: `go test -tags=noebiten -cover ./cmd/...` shows ≥30% coverage

---

## Gap 3: Performance Target Unverifiable (60 FPS)

- **Stated Goal**: README.md claims "60 FPS at 1280×720" as a performance target.
- **Current State**: No benchmark tests exist in the test suite. The raycaster has good test coverage but no `Benchmark*` functions. Average function complexity (3.6) suggests efficient code, but actual frame time is unmeasured.
- **Impact**: Cannot verify the 60 FPS claim without runtime profiling. Performance regressions would go undetected until manual testing.
- **Closing the Gap**:
  1. Add benchmark tests to `pkg/rendering/raycast/raycast_test.go`:
     ```go
     func BenchmarkRender(b *testing.B) { ... }
     func BenchmarkDDA(b *testing.B) { ... }
     ```
  2. Add benchmark to ECS world update in `pkg/engine/ecs/world_test.go`:
     ```go
     func BenchmarkWorldUpdate(b *testing.B) { ... }
     ```
  3. Add benchmark for sprite generation in `pkg/rendering/sprite/generator_test.go`
  4. Add CI step to run benchmarks and detect regressions
  - **Validation**: `go test -bench=. ./pkg/rendering/raycast/...` completes render in ≤16.67ms

---

## Gap 4: Build-Tag Test Visibility

- **Stated Goal**: CI pipeline provides clear test status for all packages.
- **Current State**: Three packages require the `noebiten` build tag for tests to run:
  - `pkg/procgen/adapters/` — 89.2% coverage with tag, shows `[no test files]` without
  - `pkg/rendering/raycast/` — 90.7% coverage with tag, shows `[no test files]` without
  - `cmd/client/` — tests exist with tag, shows `[no test files]` without
  
  Developers running `go test ./...` see these critical packages as untested.
- **Impact**: False negative perception of test coverage. Developers may believe these packages are untested and skip them during refactoring, or add redundant tests.
- **Closing the Gap**:
  1. Add prominent comment in each package's `doc.go` explaining build tag requirement
  2. Ensure CI workflow (`.github/workflows/ci.yml`) includes explicit steps:
     ```yaml
     - run: go test -tags=noebiten ./pkg/procgen/adapters/...
     - run: go test -tags=noebiten ./pkg/rendering/raycast/...
     ```
  3. Consider splitting tests into `*_ebiten_test.go` (requires Ebiten) and `*_test.go` (no dependency)
  - **Validation**: CI logs show explicit coverage numbers for all three packages

---

## Gap 5: High Complexity Functions (5 → 0)

- **Stated Goal**: Maintainable code with cyclomatic complexity ≤10 per function (per copilot-instructions.md).
- **Current State**: Five functions exceed the complexity threshold:

  | Function | File | Cyclomatic | Overall |
  |----------|------|------------|---------|
  | `drawQuadruped` | `pkg/rendering/sprite/generator.go` | 11 | 16.3 |
  | `drawSerpentine` | `pkg/rendering/sprite/generator.go` | 10 | 15.0 |
  | `GetNextUnlockForSkill` | `pkg/engine/systems/skill_progression.go` | 9 | 13.7 |
  | `drawAvian` | `pkg/rendering/sprite/generator.go` | 9 | 13.2 |
  | `GetAvailableUnlocks` | `pkg/engine/systems/skill_progression.go` | 9 | 13.2 |

- **Impact**: High complexity functions are harder to test, understand, and maintain. Bug probability correlates with complexity. These functions handle sprite rendering and skill progression — both player-visible systems.
- **Closing the Gap**:
  1. **drawQuadruped (16.3)**: Extract body-part-specific helpers: `drawQuadrupedHead()`, `drawQuadrupedBody()`, `drawQuadrupedLegs()`
  2. **drawSerpentine (15.0)**: Split into `drawSerpentineBody()`, `drawSerpentineHead()`, `drawSerpentineDetails()`
  3. **drawAvian (13.2)**: Extract genre-specific rendering into separate functions
  4. **GetNextUnlockForSkill (13.7)**: Split skill lookup from unlock condition checking
  5. **GetAvailableUnlocks (13.2)**: Extract prerequisite validation into helper
  - **Validation**: `go-stats-generator analyze . --skip-tests | grep "High Complexity"` shows 0 functions with cyclomatic >10

---

## Gap 6: Code Duplication (2.90% → <2.0%)

- **Stated Goal**: DRY code with minimal duplication.
- **Current State**: 72 clone pairs detected (1,536 duplicated lines, 2.90% ratio). Top duplication clusters:

  | Location | Lines | Instances |
  |----------|-------|-----------|
  | `pkg/engine/systems/weather.go` | 6-7 | 12 |
  | `pkg/engine/systems/stealth.go` | 6-7 | 8 |
  | `pkg/engine/systems/economic_event.go` | 6 | 7 |
  | `pkg/rendering/particles/particles.go` | 8 | 4+ |

- **Impact**: Duplicated code means bug fixes must be applied in multiple places. Weather and stealth handling divergence could cause inconsistent behavior.
- **Closing the Gap**:
  1. **weather.go**: Extract `applyWeatherModifier(type, intensity)` parameterized helper
  2. **stealth.go**: Extract `calculateVisibilityFactor(factor, lightLevel, movement)` helper
  3. **economic_event.go**: Extract `applyEconomicModifier(...)` helper
  4. **particles.go**: Extract `updateParticlePhysics(...)` helper
  - **Validation**: `go-stats-generator analyze . --skip-tests | grep "Duplication Ratio"` shows <2.0%

---

## Gap 7: Companion Package Coverage (78.8% → 85%)

- **Stated Goal**: ≥70% test coverage per package; companion AI is a key gameplay feature.
- **Current State**: `pkg/companion/` has 78.8% coverage — the lowest among gameplay packages. Package handles companion behaviors, combat roles, and relationship tracking.
- **Impact**: Companion behavior bugs could affect player experience. State machine transitions and combat role changes may have untested edge cases.
- **Closing the Gap**:
  1. Add tests for edge cases in `CompanionManager` state machine
  2. Test combat role transitions (protect → attack → support → retreat)
  3. Test relationship score boundary conditions (-100, 0, 100)
  4. Test companion command queuing under high-latency conditions
  - **Validation**: `go test -cover ./pkg/companion/...` shows ≥85%

---

## Gap 8: Large File Cohesion

- **Stated Goal**: Maintainable code with focused files.
- **Current State**: Several files exceed 1,000 LOC with low cohesion scores:

  | File | Lines | Recommendation |
  |------|-------|----------------|
  | `pkg/engine/systems/vehicle.go` | 3,291 | Split into `vehicle_movement.go`, `vehicle_fuel.go`, `vehicle_combat.go` |
  | `pkg/world/housing/housing.go` | 2,601 | Split into `rooms.go`, `furniture.go`, `ownership.go` |
  | `pkg/engine/systems/skill_progression.go` | 1,870 | Split into `skill_xp.go`, `skill_training.go` |
  | `pkg/engine/systems/quest.go` | 1,266 | Split into `quest_stages.go`, `quest_conditions.go` |
  | `pkg/engine/systems/stealth.go` | 1,218 | Split into `visibility.go`, `detection.go` |
  | `pkg/engine/systems/weather.go` | 1,189 | Split into `weather_effects.go`, `weather_transitions.go` |

- **Impact**: Large files are harder to navigate, review, and maintain. Go best practice suggests files under 500 LOC.
- **Closing the Gap**:
  1. Split `vehicle.go` into 3+ focused files (movement, fuel, combat)
  2. Split `housing.go` into domain-focused files
  3. Split other >1000 LOC files as capacity allows
  - **Validation**: No file exceeds 1,000 LOC in core packages

---

## Gap 9: Naming Convention Violations

- **Stated Goal**: Follow Go naming conventions per project style guide.
- **Current State**: go-stats-generator detected 34 identifier violations and 17 file violations:
  
  **Identifier Violations (package stuttering)**:
  - `DialogManager` in `dialog` → should be `Manager`
  - `CompanionManager` in `companion` → should be `Manager`
  - `AmbientManager` in `ambient` → should be `Manager`
  - `VehiclePhysics` in `components` (stuttering pattern)
  
  **File Violations (generic names)**:
  - `constants.go` (5 instances across packages)
  - `types.go` (4 instances)
  - `util.go` (2 instances)

- **Impact**: Package-stuttering names are Go idiomatic violations (`dialog.DialogManager` vs preferred `dialog.Manager`). Generic file names reduce code discoverability.
- **Closing the Gap**:
  1. Consider renaming stuttering identifiers during natural refactoring
  2. Rename generic files to domain-specific names
  3. Low priority as functionality is unaffected
  - **Validation**: `go-stats-generator analyze . --skip-tests | grep "Identifier Violations"` shows ≤10

---

## Summary: Gap Closure Priority

| Priority | Gap | Impact | Effort | Current State |
|----------|-----|--------|--------|---------------|
| **P1** | FEATURES.md sync | Documentation trust | Low | Summary shows 73.5% vs actual 100%+ |
| **P1** | Entry point tests | Regression risk | Medium | 0% coverage |
| **P2** | Performance benchmarks | Quality claim | Medium | No benchmarks |
| **P2** | Build-tag visibility | Developer experience | Low | CI config needed |
| **P2** | High complexity (5→0) | Maintainability | Medium | 5 functions >10 |
| **P3** | Code duplication | Maintainability | Medium | 2.90% ratio |
| **P3** | Companion coverage | Quality | Low | 78.8% (below 85% target) |
| **P4** | File cohesion | Maintainability | High | 6 files >1000 LOC |
| **P4** | Naming violations | Code style | Low | 34 identifiers |

---

## Next Steps

1. **Immediate (Day 1)**: Fix FEATURES.md summary table inconsistency — high visibility, low effort
2. **Short-term (Week 1)**: Add entry point tests and performance benchmarks
3. **Medium-term (Week 2-3)**: Refactor high-complexity sprite generator functions
4. **Ongoing**: Address duplication, cohesion, and naming during normal development

---

*Generated 2026-03-31 by comparing README.md, ROADMAP.md, and FEATURES.md claims against codebase implementation using go-stats-generator v1.0.0*
