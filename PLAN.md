# Implementation Plan: Wyrm 60% Feature Milestone

## Project Context
- **What it does**: A 100% procedurally generated first-person open-world RPG built in Go on Ebitengine, targeting 200 features across 20 categories with five genre themes (Fantasy, Sci-Fi, Horror, Cyberpunk, Post-Apocalyptic).
- **Current goal**: Achieve 60% feature completion (120/200 features) with full test coverage on critical paths
- **Estimated Scope**: Medium-Large (3-4 months, 9 features above complexity threshold, 0.23% duplication ratio, 87.9% doc coverage)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|----------------|---------------------|
| 200 features (ROADMAP.md) | ⚠️ 111/200 (55.5%) | Yes — target 120+ (60%) |
| ≥40% test coverage per package (copilot-instructions.md) | ⚠️ 4 packages below threshold | Yes — adapters (0%), raycast (0%), cmd/client (0%), cmd/server (0%) |
| Zero external assets | ✅ Achieved | No |
| Single binary distribution | ✅ Achieved | No |
| ECS architecture with 11+ systems | ✅ Achieved (15 systems) | No |
| Five genre themes | ⚠️ Partial | Yes — terrain differentiation |
| First-person raycaster 60 FPS | ✅ Achieved | No |
| Chunk streaming 512×512 | ✅ Achieved | No |
| 200–5000ms latency tolerance | ✅ Achieved | No |
| Crafting system (FEATURES.md) | ⚠️ 70% (7/10) | Yes — minigames, enchanting |
| Ranged/magic combat | ✅ Achieved (10/10) | No |
| NPC memory & relationships | ✅ Achieved (5/10) | Yes — gossip, needs, pathfinding |
| V-Series integration | ✅ Achieved (16 adapters) | Yes — test coverage |
| Server builds normally | ⚠️ Build tag issue | Yes |

## Metrics Summary

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 8,109 | Healthy |
| Total Functions/Methods | 891 | Well-distributed |
| Total Packages | 23 | Good modularity |
| Avg Function Length | 10.5 lines | Excellent (<20 target) |
| Avg Cyclomatic Complexity | 3.6 | Excellent (<10 target) |
| Functions >10 complexity | 9 | Needs attention |
| Duplication Ratio | 0.23% | Excellent (<3% threshold) |
| Documentation Coverage | 87.9% | Good (>70% target) |
| Magic Numbers | 2,879 | Moderate technical debt |
| Test Coverage (avg tested) | ~87% | Good |

### Complexity Hotspots (Goal-Critical Paths)

| Function | File | Lines | Complexity | Risk |
|----------|------|-------|------------|------|
| `GatherResource` | pkg/engine/systems/crafting.go | 73 | 22.3 | High |
| `updateSpeed` | pkg/engine/systems/vehicle.go | 44 | 17.6 | High |
| `CastSpellAtPosition` | pkg/engine/systems/magic_combat.go | 68 | 17.1 | High |
| `decayDispositions` | pkg/engine/systems/npc_memory.go | 26 | 14.2 | Medium |
| `InitiateRangedAttack` | pkg/engine/systems/ranged_combat.go | 64 | 13.5 | Medium |

### Test Coverage Gaps (Goal-Critical)

| Package | Coverage | Impact |
|---------|----------|--------|
| `pkg/procgen/adapters` | 0% (FAIL) | V-Series integration layer |
| `pkg/rendering/raycast` | 0% | Core rendering |
| `cmd/client` | 0% | User-facing entry point |
| `cmd/server` | 0% | Server entry point |
| `pkg/engine/components` | 67.1% | Below 70% target |

---

## Implementation Steps

### Step 1: Fix Server Build Constraint

- **Deliverable**: `cmd/server/main.go` builds without special tags
- **Dependencies**: None
- **Goal Impact**: Enables documented build instructions to work
- **Files**: `cmd/server/main.go` line 1
- **Acceptance**: `go build ./cmd/server && ./server --help` succeeds
- **Validation**: `go build ./cmd/server`

### Step 2: Add V-Series Adapter Tests

- **Deliverable**: Test file `pkg/procgen/adapters/adapters_test.go` with ≥70% coverage
- **Dependencies**: Step 1 (server builds)
- **Goal Impact**: V-Series integration is foundational; 0% coverage is critical risk
- **Files**: 
  - Create `pkg/procgen/adapters/adapters_test.go`
  - Test all 16 adapters for determinism and error handling
- **Acceptance**: `go test -cover ./pkg/procgen/adapters/...` shows ≥70%
- **Validation**: `go test -cover ./pkg/procgen/adapters/... | grep -E 'coverage: [7-9][0-9]|100'`

### Step 3: Add Raycast Core Tests

- **Deliverable**: Test file `pkg/rendering/raycast/core_test.go` with ≥50% coverage
- **Dependencies**: None
- **Goal Impact**: Core rendering algorithm untested; regressions undetected
- **Files**:
  - Create `pkg/rendering/raycast/core_test.go` with `//go:build noebiten`
  - Test `CastRay()`, `calculateWallDistance()`, texture coordinate calculations
- **Acceptance**: `go test -tags=noebiten -cover ./pkg/rendering/raycast/...` shows ≥50%
- **Validation**: `go test -tags=noebiten ./pkg/rendering/raycast/...`

### Step 4: Refactor GatherResource Complexity

- **Deliverable**: `GatherResource` function complexity reduced from 22.3 to ≤10
- **Dependencies**: Step 2 (tests ensure no regression)
- **Goal Impact**: Highest complexity function; maintenance burden
- **Files**: `pkg/engine/systems/crafting.go`
- **Acceptance**: Extract helper functions for resource validation, tool checks, skill modifiers
- **Validation**: `go-stats-generator analyze . --skip-tests | grep GatherResource`

### Step 5: Refactor CastSpellAtPosition Complexity

- **Deliverable**: `CastSpellAtPosition` function complexity reduced from 17.1 to ≤10
- **Dependencies**: None (has existing tests)
- **Goal Impact**: Magic combat is a core gameplay feature
- **Files**: `pkg/engine/systems/magic_combat.go`
- **Acceptance**: Extract spell validation, mana checks, effect application into helpers
- **Validation**: `go-stats-generator analyze . --skip-tests | grep CastSpellAtPosition`

### Step 6: Complete NPCs & Social Features (30% → 50%)

- **Deliverable**: Implement gossip network, NPC pathfinding to schedule locations
- **Dependencies**: None
- **Goal Impact**: NPCs & Social at 30% (3/10); core RPG experience
- **Files**:
  - Add `GossipNetwork` component to `pkg/engine/components/types.go`
  - Add `pkg/engine/systems/gossip.go` for gossip propagation
  - Add pathfinding to `pkg/engine/systems/npc_schedule.go`
- **Acceptance**: FEATURES.md NPCs & Social shows ≥50% (5/10)
- **Validation**: `grep -A 20 "NPCs & Social" FEATURES.md | grep -c '\[x\]'` shows ≥5

### Step 7: Complete Weather & Environment Features (50% → 70%)

- **Deliverable**: Add seasonal changes, environmental sounds
- **Dependencies**: None
- **Goal Impact**: Weather & Environment at 50%; atmosphere enhancement
- **Files**:
  - Extend `WeatherSystem` in `pkg/engine/systems/weather.go` for seasons
  - Add environmental sound triggers to `pkg/engine/systems/audio.go`
- **Acceptance**: FEATURES.md Weather & Environment shows ≥70% (7/10)
- **Validation**: `grep -A 15 "Weather & Environment" FEATURES.md | grep -c '\[x\]'` shows ≥7

### Step 8: Complete Crafting & Resources Features (70% → 90%)

- **Deliverable**: Add crafting minigames, enchanting system
- **Dependencies**: Step 4 (GatherResource refactor)
- **Goal Impact**: Crafting at 70%; minigames and enchanting are documented promises
- **Files**:
  - Add minigame logic to `pkg/engine/systems/crafting.go`
  - Add `EnchantingSystem` to `pkg/engine/systems/enchanting.go`
- **Acceptance**: FEATURES.md Crafting & Resources shows ≥90% (9/10)
- **Validation**: `grep -A 15 "Crafting & Resources" FEATURES.md | grep -c '\[x\]'` shows ≥9

### Step 9: Deepen Genre Terrain Differentiation

- **Deliverable**: Visually distinct terrain per genre using existing texture system
- **Dependencies**: None
- **Goal Impact**: README claims "five genre themes reshape every player-facing system"
- **Files**:
  - Extend `pkg/rendering/texture/patterns.go` with genre-specific palettes
  - Update terrain adapter to apply genre palette
- **Acceptance**: Screenshot comparison shows distinct terrain colors per genre
- **Validation**: `go test ./pkg/rendering/texture/... -run TestGenreTextures`

### Step 10: Add Client Entry Point Tests

- **Deliverable**: Test file `cmd/client/main_test.go` with ≥40% coverage
- **Dependencies**: None
- **Goal Impact**: Primary user-facing code untested
- **Files**:
  - Create `cmd/client/main_test.go`
  - Test `heightToWallType`, input processing, pure functions
- **Acceptance**: `go test ./cmd/client/...` shows ≥40% coverage
- **Validation**: `go test -cover ./cmd/client/...`

### Step 11: Extract Magic Numbers in Combat Systems

- **Deliverable**: Reduce magic numbers from 2,879 to <2,000
- **Dependencies**: None
- **Goal Impact**: Improves maintainability; reduces balance-change errors
- **Files**:
  - `pkg/engine/systems/combat.go` — extract damage multipliers
  - `pkg/engine/systems/ranged_combat.go` — extract range values
  - `pkg/engine/systems/magic_combat.go` — extract mana costs
- **Acceptance**: go-stats-generator shows <2,000 magic numbers
- **Validation**: `go-stats-generator analyze . --skip-tests | grep "Magic Numbers"`

### Step 12: Fix Clone Pair in Magic/Ranged Combat

- **Deliverable**: Eliminate 28-line duplicate block between magic_combat.go and ranged_combat.go
- **Dependencies**: Steps 5, 11
- **Goal Impact**: Reduces maintenance burden; keeps duplication ratio <1%
- **Files**:
  - `pkg/engine/systems/magic_combat.go:104-131` and `:159-186`
  - Extract shared logic to helper function
- **Acceptance**: go-stats-generator shows 0 clone pairs in these files
- **Validation**: `go-stats-generator analyze . --skip-tests | grep "Clone Pairs"`

---

## Milestone Validation

After completing all steps, verify the 60% milestone:

```bash
# Feature count (target: ≥120)
grep -c '\[x\]' FEATURES.md

# Test coverage (target: all packages ≥40%)
go test -cover ./...

# High complexity functions (target: 0 above 15.0)
go-stats-generator analyze . --skip-tests | grep -A 5 "Top Complex Functions"

# Magic numbers (target: <2000)
go-stats-generator analyze . --skip-tests | grep "Magic Numbers"

# Clone pairs (target: 0)
go-stats-generator analyze . --skip-tests | grep "Clone Pairs"
```

---

## Dependencies Graph

```
Step 1 (Server Build)
    └── Step 2 (Adapter Tests)
            └── Step 4 (GatherResource Refactor)
                    └── Step 8 (Crafting Features)
                            └── Step 12 (Clone Fix)
                                    └── Step 11 (Magic Numbers)
                                            └── Step 5 (CastSpell Refactor)

Step 3 (Raycast Tests) ──────────────────────────────────────────┐
Step 6 (NPC Features) ───────────────────────────────────────────┤
Step 7 (Weather Features) ───────────────────────────────────────┼── Parallel
Step 9 (Genre Terrain) ──────────────────────────────────────────┤
Step 10 (Client Tests) ──────────────────────────────────────────┘
```

---

## Risk Mitigations

| Risk | Mitigation |
|------|------------|
| Adapter tests require Ebiten | Use `//go:build noebiten` tag; mock Ebiten dependencies |
| GatherResource refactor breaks crafting | Run existing crafting tests after each change |
| Genre texture changes affect performance | Benchmark texture generation before/after |
| Ebitengine v2.9 deprecations | Monitor `ebiten.FillRule`, vector functions; plan migration |

---

## External Dependencies

| Dependency | Version | Notes |
|------------|---------|-------|
| Ebitengine | v2.9.3 | Vector graphics APIs deprecated; monitor for v3 migration |
| opd-ai/venture | v0.0.0-20260321 | V-Series generators; stable |
| Viper | v1.19.0 | Configuration; stable |
| Go | 1.24.5 | Current requirement; aligned with Ebitengine 2.9 |

---

## Cleanup

```bash
rm /tmp/metrics.json
```

---

*Generated: 2026-03-30 | Tool: go-stats-generator v1.0.0 | Codebase: 8,109 LOC across 72 files*
