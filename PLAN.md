# Implementation Plan: Code Quality & Documentation Alignment

## Project Context
- **What it does**: Wyrm is a 100% procedurally generated first-person open-world RPG built in Go on Ebitengine, generating all content (graphics, audio, levels) at runtime from a deterministic seed.
- **Current goal**: Maintain code quality and align documentation with actual implementation state (FEATURES.md summary table is out of sync with 201/200 features implemented).
- **Estimated Scope**: Medium (8 items above threshold across complexity, duplication, and documentation)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| 200 features across 20 categories | ✅ Achieved (201/200) | No |
| Zero external assets | ✅ Achieved | No |
| ECS architecture | ✅ Achieved | No |
| Five genre themes | ✅ Achieved | No |
| 60 FPS at 1280×720 | ⚠️ Unverifiable | Yes (benchmarks) |
| 200–5000ms latency tolerance | ✅ Achieved | No |
| Test coverage ≥70% per package | ✅ Achieved (91.4% avg) | Yes (entry points) |
| Cyclomatic complexity ≤10 | ⚠️ 2 functions above | Yes |
| Code duplication <3% | ⚠️ 2.90% (72 clone pairs) | Yes |
| Documentation sync | ❌ FEATURES.md outdated | Yes |

## Metrics Summary

- **Complexity hotspots on goal-critical paths**: 2 functions above threshold
  - `drawQuadruped` (pkg/rendering/sprite/generator.go): cyclomatic 11, overall 16.3
  - `drawSerpentine` (pkg/rendering/sprite/generator.go): cyclomatic 10, overall 15.0
- **Duplication ratio**: 2.90% (1,536 lines / 72 clone pairs)
  - Primary clusters: weather.go (12 instances), stealth.go (8 instances), economic systems (7 instances)
- **Doc coverage**: 87.3% (above 80% target)
- **Package coupling**: 0 circular dependencies (excellent)
- **Test coverage**: 91.4% average across 28 packages

## Research Findings

### Project Landscape
- No open issues or discussions on GitHub — project is in maintenance/polish phase
- Part of opd-ai Procedural Game Suite with 8 sibling repositories sharing zero-external-assets philosophy

### Dependency Health
- **Ebitengine v2.9.3**: Current, requires Go 1.24+ (project uses 1.24.5 ✅)
  - Deprecations: `vector.AppendVerticesAndIndicesForFilling` (project doesn't use)
  - New APIs available: `DrawTriangles32` for optimization opportunities
- **Viper v1.19.0**: Stable, no known issues
- **golang.org/x packages**: All current versions

## Implementation Steps

### Step 1: Synchronize FEATURES.md Summary Table
- **Deliverable**: Update `FEATURES.md` summary table to reflect actual 201/200 (100%+) feature completion instead of incorrect 147/200 (73.5%)
- **Dependencies**: None
- **Goal Impact**: Documentation accuracy — removes confusion for contributors and users about project completeness
- **Acceptance**: `grep -c '\[x\]' FEATURES.md` matches sum of "Implemented" column values; percentage shows 100%+
- **Validation**: 
  ```bash
  # Count checked boxes
  grep -c '\[x\]' FEATURES.md
  # Should return 200 or 201
  
  # Verify summary table matches
  grep -A25 "Summary by Category" FEATURES.md | grep "TOTAL"
  # Should show 200+/200 and 100%+
  ```
- **Estimated Effort**: 1 hour

### Step 2: Refactor `drawQuadruped` Sprite Generator
- **Deliverable**: Reduce cyclomatic complexity of `drawQuadruped` function in `pkg/rendering/sprite/generator.go` from 11 to ≤9 by extracting body-part-specific helpers
- **Dependencies**: None
- **Goal Impact**: Maintainability — reduces bug probability in sprite generation (key visual feature)
- **Acceptance**: `go-stats-generator analyze pkg/rendering/sprite --skip-tests | grep drawQuadruped` shows cyclomatic ≤9
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/rendering/sprite --skip-tests --format json | \
    jq '.functions[] | select(.name=="drawQuadruped") | .complexity.cyclomatic'
  # Should return ≤9
  
  go test ./pkg/rendering/sprite/...
  # All tests must pass
  ```
- **Estimated Effort**: 3-4 hours
- **Approach**: Extract `drawQuadrupedHead()`, `drawQuadrupedBody()`, `drawQuadrupedLegs()`, `drawQuadrupedTail()` helper functions

### Step 3: Refactor `drawSerpentine` Sprite Generator
- **Deliverable**: Reduce cyclomatic complexity of `drawSerpentine` function from 10 to ≤9 by extracting segment-specific helpers
- **Dependencies**: Step 2 (establishes pattern)
- **Goal Impact**: Maintainability — consistent sprite generator architecture
- **Acceptance**: `go-stats-generator analyze pkg/rendering/sprite --skip-tests | grep drawSerpentine` shows cyclomatic ≤9
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/rendering/sprite --skip-tests --format json | \
    jq '.functions[] | select(.name=="drawSerpentine") | .complexity.cyclomatic'
  # Should return ≤9
  
  go test ./pkg/rendering/sprite/...
  # All tests must pass
  ```
- **Estimated Effort**: 2-3 hours
- **Approach**: Extract `drawSerpentineBody()`, `drawSerpentineHead()`, `drawSerpentineDetails()` helper functions

### Step 4: Consolidate Weather System Duplication
- **Deliverable**: Reduce clone instances in `pkg/engine/systems/weather.go` from 12 to 0 by extracting parameterized helper functions
- **Dependencies**: None
- **Goal Impact**: Maintainability — ensures weather effect bug fixes propagate to all weather types
- **Acceptance**: `go-stats-generator analyze pkg/engine/systems --skip-tests` shows weather.go duplication reduced by 80%+
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/engine/systems/weather.go --skip-tests --format json | \
    jq '.duplication.clones | length'
  # Should return 0-2
  
  go test ./pkg/engine/systems/...
  # All tests must pass
  ```
- **Estimated Effort**: 4-6 hours
- **Approach**: Extract `applyWeatherModifier(weatherType string, intensity float64, effects ...WeatherEffect)` parameterized helper

### Step 5: Consolidate Stealth System Duplication
- **Deliverable**: Reduce clone instances in `pkg/engine/systems/stealth.go` from 8 to 0
- **Dependencies**: None (can parallel with Step 4)
- **Goal Impact**: Maintainability — visibility calculations are critical for stealth gameplay
- **Acceptance**: `go-stats-generator analyze pkg/engine/systems/stealth.go --skip-tests` shows 0 clone pairs
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/engine/systems/stealth.go --skip-tests --format json | \
    jq '.duplication.clones | length'
  # Should return 0
  
  go test ./pkg/engine/systems/...
  # All tests must pass
  ```
- **Estimated Effort**: 3-4 hours
- **Approach**: Extract `calculateVisibilityFactor(lightLevel, movementSpeed, coverBonus float64) float64` helper

### Step 6: Consolidate Economic System Duplication
- **Deliverable**: Reduce clone instances across `economic_event.go`, `investment.go`, `market_manipulation.go`, `trade_route.go` from 7 to 0
- **Dependencies**: None (can parallel with Steps 4-5)
- **Goal Impact**: Maintainability — economic systems are interconnected; divergent clones cause inconsistent behavior
- **Acceptance**: Economic system files have 0 shared clones
- **Validation**:
  ```bash
  go-stats-generator analyze pkg/engine/systems --skip-tests --format json | \
    jq '[.duplication.clones[] | select(.instances[].file | contains("economic") or contains("investment") or contains("market") or contains("trade"))] | length'
  # Should return 0
  
  go test ./pkg/engine/systems/...
  # All tests must pass
  ```
- **Estimated Effort**: 4-5 hours
- **Approach**: Create shared `applyEconomicModifier()` in a new `economic_helpers.go` file

### Step 7: Add Performance Benchmarks
- **Deliverable**: Add benchmark tests to verify 60 FPS claim for raycaster, ECS update, and sprite generation
- **Dependencies**: Steps 2-3 (sprite generator changes)
- **Goal Impact**: Verifiability — enables objective validation of README performance claims
- **Acceptance**: Benchmarks exist and render completes in <16.67ms (60 FPS target)
- **Validation**:
  ```bash
  go test -bench=. ./pkg/rendering/raycast/... | grep "ns/op"
  # Render benchmark should show <16,670,000 ns/op
  
  go test -bench=. ./pkg/engine/ecs/... | grep "ns/op"
  # World update benchmark should exist
  
  go test -bench=. ./pkg/rendering/sprite/... | grep "ns/op"
  # Sprite generation benchmark should exist
  ```
- **Estimated Effort**: 1 day
- **Files to modify**:
  - `pkg/rendering/raycast/raycast_test.go` — add `BenchmarkRender`, `BenchmarkDDA`
  - `pkg/engine/ecs/world_test.go` — add `BenchmarkWorldUpdate`
  - `pkg/rendering/sprite/generator_test.go` — add `BenchmarkGenerateSprite`

### Step 8: Add Entry Point Tests
- **Deliverable**: Achieve ≥30% test coverage for `cmd/client/` and `cmd/server/` initialization logic
- **Dependencies**: None
- **Goal Impact**: Regression safety — initialization changes (system registration, config loading) are high-risk
- **Acceptance**: `go test -tags=noebiten -cover ./cmd/...` shows ≥30% coverage
- **Validation**:
  ```bash
  go test -tags=noebiten -cover ./cmd/client/... 2>&1 | grep coverage
  # Should show ≥30%
  
  go test -tags=noebiten -cover ./cmd/server/... 2>&1 | grep coverage
  # Should show ≥30%
  ```
- **Estimated Effort**: 3-5 days
- **Files to create**:
  - `cmd/client/main_noebiten_test.go` — test config loading, system registration helpers
  - `cmd/server/main_noebiten_test.go` — test initialization, federation setup, tick loop utilities

---

## Summary

| Step | Deliverable | Effort | Impact |
|------|-------------|--------|--------|
| 1 | FEATURES.md sync | 1 hour | Documentation accuracy |
| 2 | Refactor drawQuadruped | 3-4 hours | Complexity -2 |
| 3 | Refactor drawSerpentine | 2-3 hours | Complexity -1 |
| 4 | Consolidate weather.go | 4-6 hours | -12 clone instances |
| 5 | Consolidate stealth.go | 3-4 hours | -8 clone instances |
| 6 | Consolidate economic systems | 4-5 hours | -7 clone instances |
| 7 | Performance benchmarks | 1 day | Verify 60 FPS claim |
| 8 | Entry point tests | 3-5 days | +30% cmd/ coverage |

**Total Estimated Effort**: 2-3 weeks
**Primary Focus**: Maintainability and verifiability improvements for a feature-complete codebase

---

## Metrics Validation Commands

After completing all steps, run:

```bash
# Full validation suite
go build ./cmd/client && go build ./cmd/server
go test ./...
go vet ./...

# Complexity check (should show 0 functions with cyclomatic >9)
go-stats-generator analyze . --skip-tests --format json | \
  jq '[.functions[] | select(.complexity.cyclomatic > 9)] | length'

# Duplication check (should show <2.0%)
go-stats-generator analyze . --skip-tests | grep "Duplication Ratio"

# Performance verification
go test -bench=BenchmarkRender ./pkg/rendering/raycast/... | grep "ns/op"

# Entry point coverage
go test -tags=noebiten -cover ./cmd/...
```

---

*Generated by go-stats-generator v1.0.0 analysis combined with ROADMAP.md, GAPS.md, and FEATURES.md review*
