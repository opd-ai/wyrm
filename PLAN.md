# Implementation Plan: Feature Completion to 70%

## Project Context
- **What it does**: A 100% procedurally generated first-person open-world RPG built in Go on Ebitengine, generating all content (terrain, audio, NPCs, quests) from a deterministic seed at runtime with zero external assets.
- **Current goal**: Reach 70% feature completion (140/200 features) while maintaining code quality and test coverage standards.
- **Estimated Scope**: Large (81+ features incomplete, multiple subsystems below coverage targets)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| 200 features across 20 categories | ✅ 70% (140/200) | Completed |
| Zero external assets | ✅ Achieved | No |
| ECS architecture with 11+ systems | ✅ Achieved | No |
| Five genre themes | ✅ Achieved | No |
| First-person raycaster at 60 FPS | ✅ Tests with noebiten tag (75.8% cov) | Completed |
| Chunk streaming (512×512) | ✅ 98.0% coverage | No |
| Multiplayer (200–5000ms latency) | ✅ Achieved | No |
| V-Series integration (25+ generators) | ✅ 16 adapters with noebiten tests | Completed |
| ≥70% test coverage per package | ⚠️ components 97.6%, systems 63.5% | Partial |
| High complexity functions ≤10 | ✅ 0 functions exceed threshold | Completed |

## Metrics Summary

- **Complexity hotspots on goal-critical paths**: 0 functions above threshold (resolved)
- **Duplication ratio**: 0.35% (excellent)
- **Doc coverage**: 85.4% (above 70% target)
- **Package coupling**: `adapters` at 10.0 coupling (20 dependencies), manageable
- **Test coverage status**: `pkg/engine/components` 97.6% ✅, `pkg/engine/systems` 63.5% ⚠️ (below 70%)

## Feature Completion Analysis

Categories below 50% completion (blocking 70% goal):

| Category | Current | Target | Features Needed |
|----------|---------|--------|-----------------|
| Dialog & Conversation | 40% (4/10) | 60% | +2 features |
| Quests & Narrative | 40% (4/10) | 60% | +2 features |
| Skills & Progression | 40% (4/10) | 60% | +2 features |
| Property & Housing | 40% (4/10) | 60% | +2 features |
| Music System | 40% (4/10) | 60% | +2 features |
| Cities & Structures | 60% (6/10) | 70% | +1 feature |
| NPCs & Social | 70% (7/10) | 80% | +1 feature |
| Factions & Politics | 50% (5/10) | 60% | +1 feature |
| Crime & Law | 50% (5/10) | 60% | +1 feature |
| Economy & Trade | 50% (5/10) | 60% | +1 feature |
| Rendering & Graphics | 50% (5/10) | 60% | +1 feature |
| World & Exploration | 50% (5/10) | 60% | +1 feature |

**Path to 70%**: Need 21 additional features (from 119 to 140).

---

## Implementation Steps

### Step 1: Fix CI Test Reporting (P0) ✅ COMPLETED

- **Deliverable**: CI pipeline reports accurate coverage for raycast and adapters packages
- **Dependencies**: None
- **Goal Impact**: Unblocks accurate quality measurement; no false negatives in CI
- **Acceptance**: `go test -tags=noebiten ./pkg/rendering/raycast/...` shows coverage in CI logs
- **Result**: Raycast tests report 75.8% coverage with noebiten tag

### Step 2: Improve pkg/engine/components Coverage to ≥70% ✅ COMPLETED

- **Deliverable**: Add tests in `pkg/engine/components/types_test.go` covering component Type() methods and edge cases
- **Dependencies**: None
- **Goal Impact**: Meets project's stated ≥70% coverage standard; stabilizes ECS foundation
- **Acceptance**: `go test -cover ./pkg/engine/components/...` shows ≥70%
- **Result**: Coverage at 97.6%

### Step 3: Improve pkg/engine/systems Coverage to ≥70% ⚠️ PARTIAL

- **Deliverable**: Add integration tests for system interactions in `pkg/engine/systems/*_test.go`
- **Dependencies**: Step 2 (components must be stable for system tests)
- **Goal Impact**: Meets ≥70% standard; reduces gameplay bug risk
- **Acceptance**: `go test -cover ./pkg/engine/systems/...` shows ≥70%
- **Status**: Coverage at 63.5% (below 70% target). Goal-Achievement Status claims 71.4% but current measurement shows regression or different measurement point.

### Step 4: Reduce High Complexity Functions to ≤10 ✅ COMPLETED

- **Deliverable**: Refactor 3 functions exceeding complexity threshold
- **Dependencies**: Step 3 (tests must cover functions before refactoring)
- **Goal Impact**: Maintainability; enables safer feature additions in vehicle/NPC/crafting
- **Acceptance**: All functions ≤10 cyclomatic complexity
- **Result**: 0 functions exceed complexity threshold

### Step 5: Add Dialog & Conversation Features (+2) ✅ COMPLETED

- **Deliverable**: Persuasion and Intimidation skill checks in `pkg/dialog/`
- **Dependencies**: None
- **Goal Impact**: Dialog category expanded
- **Result**: FEATURES.md shows both features implemented

### Step 6: Add Quests & Narrative Features (+2) ✅ COMPLETED

- **Deliverable**: Dynamic quest generation and radiant quest system
- **Dependencies**: Step 5
- **Goal Impact**: Quests category expanded
- **Result**: FEATURES.md shows both features implemented

### Step 7: Add Skills & Progression Features (+2) ✅ COMPLETED

- **Deliverable**: NPC training system and 30+ unique skills
- **Dependencies**: Step 5
- **Goal Impact**: Skills category expanded
- **Result**: FEATURES.md shows both features implemented

### Step 8: Add Property & Housing Features (+2) ✅ COMPLETED

- **Deliverable**: Property purchasing and first-person furniture placement
- **Dependencies**: None
- **Goal Impact**: Property category expanded
- **Result**: FEATURES.md shows both features implemented

### Step 9: Add Music System Features (+2) ✅ COMPLETED

- **Deliverable**: Genre-specific music styles and location-based transitions
- **Dependencies**: None
- **Goal Impact**: Music category expanded
- **Result**: FEATURES.md shows both features implemented

### Step 10: Add Remaining High-Impact Features (+11) ✅ COMPLETED

- **Deliverable**: Add features across remaining incomplete categories:
  - Cities & Structures: Residential areas (+1)
  - NPCs & Social: NPC pathfinding to schedule locations (+1)
  - Factions & Politics: Dynamic faction wars (+1)
  - Crime & Law: Guard pursuit AI (+1)
  - Economy & Trade: Player-owned shops (+1)
  - Rendering & Graphics: Sprite rendering for NPCs (+1)
  - World & Exploration: Vertical terrain (hills, cliffs) (+1)
  - Audio System: Reverb effects (+1)
  - Vehicles & Mounts: Vehicle combat (+1)
  - Networking: PvP combat validation (+1)
  - Technical: Colorblind modes (+1)
- **Dependencies**: Steps 5-9 (core system features first)
- **Goal Impact**: Reaches 140/200 (70%) feature completion
- **Acceptance**: `grep -c '\[x\]' FEATURES.md` returns 140 or higher
- **Result**: 140/200 features (70%) achieved

### Step 11: Add Entry Point Tests (P2) ✅ COMPLETED

- **Deliverable**: 
  - Create `cmd/client/main_test.go` testing extracted pure functions
  - Create `cmd/server/main_test.go` testing initialization logic
- **Dependencies**: Step 10 (features implemented to test)
- **Goal Impact**: Reduces regression risk; 0% → ≥30% coverage for entry points
- **Acceptance**: `go test ./cmd/...` passes with ≥30% coverage per entry point
- **Validation**:
  ```bash
  go test -cover ./cmd/... | grep -E "coverage: ([3-9][0-9]|100)\.[0-9]+%"
  ```
- **Result**: Tests exist with noebiten build tag; client util.go 100%, server 83.8% coverage

### Step 12: Add Federation Integration Test (P2) ✅ COMPLETED

- **Deliverable**: Integration test in `pkg/network/federation/integration_test.go` with 2 server instances
- **Dependencies**: None
- **Goal Impact**: Validates cross-server feature actually works at runtime
- **Acceptance**: Integration test passes with player transfer between 2 nodes
- **Validation**:
  ```bash
  go test -v ./pkg/network/federation/... -run Integration | grep -E "PASS.*Integration"
  ```
- **Result**: TestFullTransferIntegration and TestConcurrentTransfers implemented

---

## Milestone Summary

| Milestone | Steps | Features Added | Coverage Impact | Validation |
|-----------|-------|----------------|-----------------|------------|
| CI Quality | 1 | 0 | Accurate metrics | CI green with coverage |
| Foundation | 2-4 | 0 | +10-15% systems/components | Coverage ≥70% |
| Core Features | 5-9 | +10 | Maintained | 129/200 (64.5%) |
| Feature Complete | 10 | +11 | Maintained | 140/200 (70%) |
| Polish | 11-12 | 0 | +30% entry points | Full integration |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Feature interdependencies | Medium | High | Ordered steps by dependency |
| Complexity refactoring breaks tests | Low | Medium | Step 3 before Step 4 |
| CI xvfb unavailable | Low | Low | Document manual test procedure |
| Ebitengine API changes | Low | Medium | Pin to v2.9.3, monitor releases |

## Estimated Timeline

- **Steps 1-4 (Quality)**: 2 weeks
- **Steps 5-9 (Core Features)**: 4 weeks
- **Step 10 (Breadth Features)**: 4 weeks
- **Steps 11-12 (Polish)**: 2 weeks
- **Total**: ~12 weeks to 70% feature completion

---

## Appendix: Validation Commands

```bash
# Build verification
go build ./cmd/client && go build ./cmd/server

# Full test suite
go test -race ./...

# Coverage report
go test -cover ./...

# Raycast tests (special tag)
go test -tags=noebiten -cover ./pkg/rendering/raycast/...

# Adapter tests (requires xvfb)
xvfb-run -a go test ./pkg/procgen/adapters/...

# Static analysis
go vet ./...

# Metrics refresh
go-stats-generator analyze . --skip-tests

# Feature count
grep -c '\[x\]' FEATURES.md

# High complexity check
go-stats-generator analyze . --skip-tests 2>&1 | grep "High Complexity"
```

---

*Generated: 2026-03-30 using go-stats-generator v1.0.0*
*Based on: README.md, ROADMAP.md, FEATURES.md, GAPS.md, AUDIT.md, and runtime metrics*
