# Implementation Plan: Phase 2 Completion & Quality Hardening

## Project Context
- **What it does**: 100% procedurally generated first-person open-world RPG built in Go on Ebitengine—no external assets, single binary distribution
- **Current goal**: Complete Phase 1→2 transition with zero-coverage packages addressed, CI/CD established, and high-latency networking finalized
- **Estimated Scope**: **Medium** (11 items above threshold across 3 categories)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Zero external assets | ✅ Achieved | No |
| Single binary distribution | ✅ Achieved | No |
| ECS architecture (11 systems) | ✅ Achieved | No |
| Five genre themes | ⚠️ Partial (genre doesn't affect terrain/textures deeply) | Yes |
| V-Series integration (25+ generators) | ✅ Achieved (16 adapters) | Partial |
| 200–5000ms latency tolerance | ⚠️ Partial (lag comp exists, no Tor-mode) | Yes |
| CI/CD pipeline | ❌ Missing | Yes |
| Test coverage >70% all packages | ⚠️ Partial (3 packages at 0%) | Yes |
| 200 features | ❌ 28% complete (~55/200) | Partial |

## Metrics Summary

| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Total LOC | 6,015 | — | Healthy |
| Functions above complexity 9.0 | **0** | <5 Small | ✅ Excellent |
| Duplication ratio | **0%** | <3% Small | ✅ Excellent |
| Doc coverage | **86.4%** | >70% | ✅ Good |
| High-complexity functions (>10) | 0 | <5 Small | ✅ Excellent |
| Packages at 0% coverage | **3** | 0 | ❌ Needs attention |
| Avg function complexity | 3.4 | <10 | ✅ Excellent |

### Complexity Hotspots (Top 5, all under threshold)

| Function | Package | Complexity | Lines |
|----------|---------|------------|-------|
| `GenerateDungeonPuzzles` | adapters | 8.8 | 24 |
| `GetAtTime` | network | 8.8 | 31 |
| `CanCraft` | adapters | 8.8 | 15 |
| `ReportKill` | systems | 8.8 | 25 |
| `processSpatialAudio` | systems | 8.8 | 30 |

### Package Coupling Analysis

| Package | Cohesion | Coupling | Assessment |
|---------|----------|----------|------------|
| systems | 10.0 | 1.5 | High cohesion, low coupling ✅ |
| adapters | 2.1 | 10.0 | Low cohesion, high coupling ⚠️ |
| network | 5.5 | 0.0 | Moderate cohesion, isolated ✅ |
| components | 9.0 | 0.0 | High cohesion ✅ |

### Zero-Coverage Packages (Critical)

| Package | Coverage | Risk | Impact |
|---------|----------|------|--------|
| `pkg/procgen/adapters` | 0.0% | High | V-Series integration layer |
| `pkg/rendering/raycast` | 0.0% | High | Core rendering path |
| `cmd/client` | 0.0% | Low | Entry point (acceptable) |
| `cmd/server` | 0.0% | Low | Entry point (acceptable) |

---

## Implementation Steps

### Step 1: Establish CI/CD Pipeline

- **Deliverable**: `.github/workflows/ci.yml` with build, test, lint, and coverage gates
- **Dependencies**: None (can start immediately)
- **Goal Impact**: Blocks merge of failing code; enables automated quality enforcement
- **Acceptance**: PR workflow runs on all pushes/PRs; badge shows passing status
- **Validation**: 
  ```bash
  # Verify workflow file exists and is valid YAML
  cat .github/workflows/ci.yml | yq '.'
  # After first run, check workflow status
  gh run list --workflow=ci.yml --limit=1 --json status | jq '.[0].status'
  ```

**Implementation Details**:
```yaml
# .github/workflows/ci.yml
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
      - run: go test -race -cover ./...
      - run: go vet ./...
```

---

### Step 2: Add Tests for `pkg/procgen/adapters`

- **Deliverable**: `pkg/procgen/adapters/adapters_test.go` with ≥70% coverage
- **Dependencies**: Step 1 (CI will validate coverage)
- **Goal Impact**: Validates V-Series integration layer; currently highest-risk zero-coverage package
- **Acceptance**: `go test -cover ./pkg/procgen/adapters/...` shows ≥70%
- **Validation**:
  ```bash
  go test -cover ./pkg/procgen/adapters/... | grep -E 'coverage: [7-9][0-9]|coverage: 100'
  go-stats-generator analyze ./pkg/procgen/adapters --skip-tests --format json | jq '.packages[0].documentation.quality_score > 0'
  ```

**Test Cases to Cover**:
1. `EntityAdapter` — generation produces valid entities with required components
2. `FactionAdapter` — faction relationships are symmetric
3. `QuestAdapter` — quest stages are properly ordered
4. `TerrainAdapter` — heightmaps are deterministic for same seed
5. `DialogAdapter` — topic graphs have no orphan nodes
6. Error handling — adapters return errors for invalid inputs (zero seed, empty genre)

---

### Step 3: Add Tests for `pkg/rendering/raycast`

- **Deliverable**: `pkg/rendering/raycast/raycast_test.go` with ≥50% coverage
- **Dependencies**: Step 1
- **Goal Impact**: Validates core rendering path; currently untested
- **Acceptance**: `go test -cover ./pkg/rendering/raycast/...` shows ≥50%
- **Validation**:
  ```bash
  go test -cover ./pkg/rendering/raycast/... | grep -E 'coverage: [5-9][0-9]|coverage: 100'
  ```

**Test Cases to Cover**:
1. `NewRenderer` — creates renderer with valid dimensions
2. DDA algorithm — ray hits wall at expected distance
3. `SetWorldData` — correctly updates internal map representation
4. `GetCameraPosition` — returns expected values after movement
5. Boundary conditions — rays at map edges don't panic

**Note**: Tests must use headless build tags or mock screen to avoid xvfb dependency.

---

### Step 4: Implement Tor-Mode High-Latency Adaptation

- **Deliverable**: RTT measurement and adaptive prediction in `pkg/network/prediction.go`
- **Dependencies**: None (isolated network package)
- **Goal Impact**: Fulfills "200–5000ms latency tolerance" claim from README
- **Acceptance**: Game adapts prediction window when RTT > 800ms
- **Validation**:
  ```bash
  go test -v ./pkg/network/... -run TestHighLatency
  go test -v ./pkg/network/... -run TestTorMode
  ```

**Implementation Requirements (from ROADMAP.md §5)**:
1. Measure RTT via Ping/Pong message timestamps
2. Detect Tor-mode when RTT > 800ms
3. Increase prediction window to 1500ms in Tor-mode
4. Reduce input send rate to 10 Hz in Tor-mode
5. Enable aggressive interpolation with 300ms blend time

**New Code Locations**:
- `pkg/network/prediction.go` — add `RTTEstimator` struct
- `pkg/network/protocol.go` — add timestamp fields to Ping/Pong
- `pkg/network/prediction_test.go` — add latency simulation tests

---

### Step 5: Add Genre-Specific Terrain Biomes

- **Deliverable**: `pkg/procgen/adapters/terrain.go` with genre-aware biome distribution
- **Dependencies**: Step 2 (adapter tests established)
- **Goal Impact**: Fulfills "Five genre themes reshape every player-facing system"
- **Acceptance**: Visual inspection of 5 genre screenshots shows distinct terrain
- **Validation**:
  ```bash
  go test -v ./pkg/procgen/adapters/... -run TestGenreTerrain
  # Compare biome distributions:
  go test -v ./pkg/procgen/adapters/... -run TestBiomeDistribution | grep -E 'fantasy|sci-fi|horror|cyberpunk|post-apocalyptic'
  ```

**Genre-Biome Mapping (from ROADMAP.md)**:
| Genre | Primary Biomes | Secondary Features |
|-------|---------------|-------------------|
| Fantasy | Forests, mountains, lakes | Ruins, shrines |
| Sci-Fi | Craters, tech structures | Alien flora, mining sites |
| Horror | Swamps, dead forests | Fog zones, graveyards |
| Cyberpunk | Urban sprawl, industrial | Neon, pollution |
| Post-Apocalyptic | Wasteland, ruins | Radiation zones, shanties |

---

### Step 6: Add Genre-Specific Texture Palettes

- **Deliverable**: Color palette selection in `pkg/rendering/texture/generator.go`
- **Dependencies**: Step 5 (terrain provides context for textures)
- **Goal Impact**: Visual differentiation across genres
- **Acceptance**: Textures use correct palette per genre
- **Validation**:
  ```bash
  go test -v ./pkg/rendering/texture/... -run TestGenrePalette
  # Verify palette constants:
  grep -n 'GenrePalette\|palette' pkg/rendering/texture/generator.go
  ```

**Genre Visual Palettes (from ROADMAP.md)**:
| Genre | Palette Colors (RGB ranges) |
|-------|----------------------------|
| Fantasy | Warm gold (#D4A574), green (#4A7C23), brown (#8B4513) |
| Sci-Fi | Cool blue (#1E90FF), white (#F0F0F0), chrome (#C0C0C0) |
| Horror | Desaturated grey-green (#556B2F), black (#1A1A1A), blood (#8B0000) |
| Cyberpunk | Neon pink (#FF00FF), cyan (#00FFFF), dark grey (#2F2F2F) |
| Post-Apocalyptic | Sepia (#704214), orange dust (#CC7722), rust (#B7410E) |

---

### Step 7: Integrate Federation at Server Runtime

- **Deliverable**: `FederationNode` initialization in `cmd/server/main.go`
- **Dependencies**: None (federation package is well-tested at 90.4%)
- **Goal Impact**: Enables cross-server player transfer
- **Acceptance**: Two server instances can exchange player entities
- **Validation**:
  ```bash
  # Start two servers, verify peer discovery
  go test -v ./pkg/network/federation/... -run TestCrossServerTransfer
  # Integration test (manual):
  # 1. Start server A with federation enabled
  # 2. Start server B pointing to server A
  # 3. Player on A can transfer to B
  ```

**Configuration Addition** (`config.yaml`):
```yaml
federation:
  enabled: false
  node_id: ""  # Auto-generated if empty
  peers: []    # List of peer server addresses
  gossip_interval: 5s
```

---

### Step 8: Split `pkg/engine/systems/registry.go`

- **Deliverable**: One file per system type in `pkg/engine/systems/`
- **Dependencies**: None (refactoring only, no behavior change)
- **Goal Impact**: Maintainability; current 1,235-line file is difficult to navigate
- **Acceptance**: No file in `pkg/engine/systems/` exceeds 200 lines; all tests pass
- **Validation**:
  ```bash
  wc -l pkg/engine/systems/*.go | grep -v total | awk '$1 > 200 {exit 1}'
  go test ./pkg/engine/systems/...
  go-stats-generator analyze ./pkg/engine/systems --skip-tests --format json | jq '.packages[0].cohesion_score'
  ```

**Proposed File Split**:
| New File | Contents | Est. Lines |
|----------|----------|-----------|
| `world_clock.go` | WorldClockSystem, WorldClock component support | 80 |
| `npc_schedule.go` | NPCScheduleSystem | 100 |
| `faction.go` | FactionPoliticsSystem | 150 |
| `crime.go` | CrimeSystem | 120 |
| `economy.go` | EconomySystem | 100 |
| `quest.go` | QuestSystem | 120 |
| `weather.go` | WeatherSystem | 80 |
| `combat.go` | CombatSystem | 100 |
| `vehicle.go` | VehicleSystem | 100 |
| `audio.go` | AudioSystem | 100 |
| `render.go` | RenderSystem | 100 |
| `registry.go` | System registration helpers only | 50 |

---

### Step 9: Add Combat Mechanics (First-Person Melee)

- **Deliverable**: Attack input handling, melee range detection, damage calculation
- **Dependencies**: Step 8 (combat.go split makes this cleaner)
- **Goal Impact**: Fulfills "First-person melee combat" claim
- **Acceptance**: Player can attack NPCs; damage is applied
- **Validation**:
  ```bash
  go test -v ./pkg/engine/systems/... -run TestMeleeCombat
  go test -v ./pkg/engine/systems/... -run TestDamageCalculation
  ```

**Implementation Requirements**:
1. Add attack input handling in `cmd/client/main.go` (mouse click)
2. Implement melee range detection using spatial queries (AABB or radius)
3. Add damage calculation with skill modifiers from `Skills` component
4. Wire to `Health` component reduction
5. Add combat feedback (future: audio cue, visual indicator)

**New Components** (if not present):
- `Weapon` — damage, range, attack speed
- `CombatState` — last attack time, cooldown

---

### Step 10: Add Stealth Mechanics

- **Deliverable**: `Stealth` component, `StealthSystem`, sneak movement
- **Dependencies**: Step 9 (combat provides foundation)
- **Goal Impact**: Fulfills "Stealth system with sneak, pickpocket, and backstab"
- **Acceptance**: Player can sneak past NPCs without detection
- **Validation**:
  ```bash
  go test -v ./pkg/engine/systems/... -run TestStealth
  go test -v ./pkg/engine/systems/... -run TestBackstab
  ```

**Implementation Requirements**:
1. Add `Stealth` component: `{Visibility float64, Sneaking bool, DetectionRadius float64}`
2. Add `StealthSystem` that checks NPC sight cones
3. Implement sneak movement (crouch key reduces speed, lowers visibility)
4. Implement backstab multiplier (2x damage when attacking unaware NPC)
5. Pickpocket action with skill check against NPC awareness

---

### Step 11: Create Feature Tracking Document

- **Deliverable**: `FEATURES.md` with checklist of 200 features
- **Dependencies**: All previous steps provide context
- **Goal Impact**: Tracks progress toward "200 features" claim
- **Acceptance**: Document lists all features with completion status
- **Validation**:
  ```bash
  # Count implemented vs total
  grep -c '\[x\]' FEATURES.md
  grep -c '\[ \]' FEATURES.md
  # Verify total is 200
  grep -cE '^\s*-\s*\[' FEATURES.md | grep '^200$'
  ```

**Document Structure**:
```markdown
# Feature Checklist (200 Features)

## World & Exploration (25 features)
- [x] Chunk streaming (512×512)
- [x] Multi-biome terrain
- [ ] Vertical terrain (caves)
...

## NPCs & Social (20 features)
- [x] NPC schedules
- [x] Faction relationships
- [ ] NPC gossip networks
...
```

---

## Dependency Graph

```
Step 1 (CI/CD)
    ├── Step 2 (adapters tests)
    │       └── Step 5 (genre terrain)
    │               └── Step 6 (genre palettes)
    ├── Step 3 (raycast tests)
    └── Step 4 (Tor-mode)

Step 7 (federation) ─── independent
Step 8 (split registry) ─── independent
        └── Step 9 (combat)
                └── Step 10 (stealth)

Step 11 (FEATURES.md) ─── after all others
```

---

## Research Findings (from web search)

### Ebitengine v2.9 Breaking Changes
- **Impact**: Wyrm uses Ebitengine v2.8.8 → v2.9.3 (current in go.mod)
- **Action Required**: None. Current version is compatible; deprecated vector APIs are not used.
- **Future Risk**: v3.0 will remove deprecated APIs. Monitor `vector.AppendVerticesAndIndicesForFilling` usage if added.

### Community Context
- **Wyrm GitHub**: No open issues; repository in early stage
- **Venture (sibling repo)**: v8.0 production-ready; active issue tracker can inform patterns
- **Recommendation**: Continue matching Venture patterns for future code sharing

---

## Validation Commands Summary

```bash
# Full validation suite
go build ./cmd/...
go test -race -cover ./...
go vet ./...

# Specific step validations
# Step 1: gh run list --workflow=ci.yml
# Step 2: go test -cover ./pkg/procgen/adapters/... | grep 'coverage:'
# Step 3: go test -cover ./pkg/rendering/raycast/... | grep 'coverage:'
# Step 4: go test -v ./pkg/network/... -run TestTorMode
# Step 5: go test -v ./pkg/procgen/adapters/... -run TestGenreTerrain
# Step 6: go test -v ./pkg/rendering/texture/... -run TestGenrePalette
# Step 7: go test -v ./pkg/network/federation/... -run TestCrossServer
# Step 8: wc -l pkg/engine/systems/*.go
# Step 9: go test -v ./pkg/engine/systems/... -run TestMeleeCombat
# Step 10: go test -v ./pkg/engine/systems/... -run TestStealth
# Step 11: grep -c '\[x\]' FEATURES.md

# Metrics re-check
go-stats-generator analyze . --skip-tests --format json --sections functions,packages | jq '.overview'
```

---

## Appendix: Metrics Baseline (2026-03-29)

```json
{
  "overview": {
    "total_lines_of_code": 6015,
    "total_functions": 214,
    "total_methods": 484,
    "total_structs": 164,
    "total_interfaces": 5,
    "total_packages": 23,
    "total_files": 46
  },
  "documentation": {
    "coverage": {
      "overall": 86.4
    }
  },
  "duplication": {
    "duplication_ratio": 0
  },
  "high_complexity_count": 0
}
```
