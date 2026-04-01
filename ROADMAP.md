# Goal-Achievement Assessment

**Generated**: 2026-03-31  
**Tool**: `go-stats-generator analyze . --skip-tests`  
**Codebase Version**: 28,050 lines of Go code (non-test) across 142 source files

---

## Project Context

### What It Claims To Do

Wyrm is described as a **"100% procedurally generated first-person open-world RPG"** built in Go on Ebitengine. The README makes the following key claims:

1. **Zero External Assets**: "No image files, no audio files, no level data. The game compiles to a single binary that runs anywhere without external assets."
2. **200 Features**: "Wyrm targets 200 features across 20 categories"
3. **Five Genre Themes**: Fantasy, Sci-Fi, Horror, Cyberpunk, Post-Apocalyptic — each reshapes every player-facing system
4. **Multiplayer**: "Authoritative server with client-side prediction and delta compression" with "200–5000 ms latency tolerance (designed for Tor-routed connections)"
5. **V-Series Integration**: Import and extend 25+ generators from `opd-ai/venture` and rendering/networking from `opd-ai/violence`
6. **Performance Targets**: "60 FPS at 1280×720; 20 Hz server tick; <500 MB client RAM"
7. **ECS Architecture**: Entity-Component-System with 11+ named systems registered and operational
8. **Core Gameplay**: "First-person melee, ranged, and magic combat", "crafting via material gathering", "NPC memory, relationships", "dynamic faction territory"

### Target Audience

- Players seeking procedurally generated open-world RPG experiences
- Developers interested in deterministic PCG techniques
- The opd-ai procedural game suite ecosystem

### Architecture

| Layer | Packages | Responsibility |
|-------|----------|----------------|
| **Entrypoints** | `cmd/client`, `cmd/server` | Game client (Ebitengine) and authoritative server |
| **ECS Core** | `pkg/engine/ecs`, `pkg/engine/components`, `pkg/engine/systems` | Entity-Component-System with 54 system files |
| **World** | `pkg/world/chunk`, `pkg/world/housing`, `pkg/world/persist`, `pkg/world/pvp` | Chunk streaming, player housing, persistence, PvP zones |
| **Rendering** | `pkg/rendering/raycast`, `pkg/rendering/texture`, `pkg/rendering/postprocess`, `pkg/rendering/sprite`, `pkg/rendering/lighting`, `pkg/rendering/particles`, `pkg/rendering/subtitles` | First-person raycaster with procedural textures, sprites, lighting, particles, and subtitles |
| **Procgen** | `pkg/procgen/adapters`, `pkg/procgen/city`, `pkg/procgen/dungeon`, `pkg/procgen/noise` | V-Series adapters (34 files) and local generators |
| **Audio** | `pkg/audio`, `pkg/audio/ambient`, `pkg/audio/music` | Procedural synthesis with adaptive music |
| **Network** | `pkg/network`, `pkg/network/federation` | Client-server with federation support |
| **Gameplay** | `pkg/companion`, `pkg/dialog`, `pkg/input` | Companion AI, dialog trees, and key rebinding |

### Existing CI/Quality Gates

- **CI Pipeline**: `.github/workflows/ci.yml` implements:
  - Build verification (`go build ./cmd/client`, `go build ./cmd/server`)
  - Test with race detection (`xvfb-run -a go test -race ./...`)
  - Build-tag-specific tests (`go test -tags=noebiten ./pkg/procgen/adapters/...`, etc.)
  - Static analysis (`go vet ./...`, `gofmt -l .`)
  - Security scanning (`govulncheck ./...`)
  - Coverage upload to Codecov
- **Build**: ✅ PASSES — Both client and server build successfully
- **Vet**: ✅ PASSES — No static analysis issues
- **Tests**: ✅ PASSES — All 30 packages pass (27 with standard tests, 4 require `noebiten` tag)

---

## Goal-Achievement Summary

| # | Stated Goal | Status | Evidence | Gap Description |
|---|-------------|--------|----------|-----------------|
| 1 | Zero external assets | ✅ Achieved | No PNG/WAV/OGG files in repo; procedural texture/audio in `pkg/` | — |
| 2 | Single binary distribution | ✅ Achieved | `go build ./cmd/client` produces standalone executable | — |
| 3 | ECS architecture | ✅ Achieved | `pkg/engine/ecs/` with World, Entity, Component, System; 54 system files | — |
| 4 | Five genre themes | ✅ Achieved | Genre-specific vehicles, weather pools, textures, biomes; adapters accept genre parameter | — |
| 5 | Chunk streaming (512×512) | ✅ Achieved | `pkg/world/chunk/` with Manager, 3×3 window, raycaster integration | 95.1% test coverage |
| 6 | First-person raycaster | ✅ Achieved | `pkg/rendering/raycast/` with DDA, floor/ceiling, textured walls | 90.7% coverage with `noebiten` tag |
| 7 | Procedural textures | ✅ Achieved | `pkg/rendering/texture/` with noise-based generation | 98.2% test coverage |
| 8 | Day/night cycle & world clock | ✅ Achieved | `WorldClockSystem` advances time; `WorldClock` component | Fully implemented |
| 9 | NPC schedules | ✅ Achieved | `NPCScheduleSystem` reads WorldClock, updates `Schedule.CurrentActivity` | Implemented |
| 10 | NPC memory and relationships | ✅ Achieved | `NPCMemorySystem` with event recording, disposition tracking | Implemented |
| 11 | Faction politics | ✅ Achieved | `FactionPoliticsSystem` with relations map, decay, kill tracking | Implemented |
| 12 | Crime system (0-5 stars) | ✅ Achieved | `CrimeSystem` with witness LOS, bounty, wanted level decay, jail | Implemented |
| 13 | Economy system | ✅ Achieved | `EconomySystem` with supply/demand, price fluctuation | Implemented |
| 14 | Quest system with branching | ✅ Achieved | `QuestSystem` with stage conditions, branch locking, consequence flags | Implemented |
| 15 | Vehicle system | ✅ Achieved | `VehicleSystem` with movement, fuel consumption; genre archetypes | Multi-file implementation |
| 16 | Weather system | ✅ Achieved | `WeatherSystem` with genre-specific pools, transitions | Implemented |
| 17 | Procedural audio synthesis | ✅ Achieved | `pkg/audio/` with sine waves, ADSR envelopes | 94.3% coverage |
| 18 | Adaptive music | ✅ Achieved | `pkg/audio/music/` with motifs, intensity states, combat detection | 97.6% test coverage |
| 19 | Spatial audio | ✅ Achieved | `AudioSystem` with distance attenuation | Implemented |
| 20 | V-Series integration | ✅ Achieved | 34 adapter files in `pkg/procgen/adapters/` wrapping Venture generators | 89.2% coverage with `noebiten` tag |
| 21 | City generation | ✅ Achieved | `pkg/procgen/city/` generates districts; server spawns NPCs | 98.0% test coverage |
| 22 | Dungeon generation | ✅ Achieved | `pkg/procgen/dungeon/` with BSP rooms, boss areas, puzzles | 92.6% test coverage |
| 23 | Melee combat | ✅ Achieved | `CombatSystem` with melee, damage calc, cooldowns, target finding | Skill modifiers implemented |
| 24 | Ranged combat | ✅ Achieved | `ProjectileSystem` with spawn, movement, collision detection | Implemented |
| 25 | Magic combat | ✅ Achieved | `MagicCombatSystem` with mana, spell effects, AoE targeting | Implemented |
| 26 | Stealth system | ✅ Achieved | `StealthSystem` with visibility, sneak, sight cones, backstab | Multi-file implementation |
| 27 | Network server | ✅ Achieved | `pkg/network/server.go` with TCP, client tracking, message dispatch | Implemented |
| 28 | Client-side prediction | ✅ Achieved | `pkg/network/prediction.go` with input buffer, reconciliation, Tor-mode | 83.2% coverage |
| 29 | Lag compensation | ✅ Achieved | `pkg/network/lagcomp.go` with position history ring buffer | 500ms rewind window |
| 30 | Tor-mode adaptive networking | ✅ Achieved | `IsTorMode()`, adaptive prediction window (1500ms), input rate (10Hz), blend time (300ms) | Fully implemented |
| 31 | Server federation | ✅ Achieved | `pkg/network/federation/` with FederationNode, gossip, transfer | 90.4% coverage |
| 32 | Player housing | ✅ Achieved | `pkg/world/housing/` with rooms, furniture, ownership | 93.0% test coverage |
| 33 | PvP zones | ✅ Achieved | `pkg/world/pvp/` with zone definitions, combat validation | 89.4% test coverage |
| 34 | World persistence | ✅ Achieved | `pkg/world/persist/` with entity serialization, chunk saves | 93.0% test coverage |
| 35 | Dialog system | ✅ Achieved | `pkg/dialog/` with topics, sentiment, responses | 87.8% test coverage |
| 36 | Companion AI | ✅ Achieved | `pkg/companion/` with behaviors, combat roles, relationship | 87.1% test coverage |
| 37 | Post-processing effects | ✅ Achieved | `pkg/rendering/postprocess/` with 13 effect types | 100.0% test coverage |
| 38 | 60 FPS target | ⚠️ Partial | Benchmarks exist but runtime profiling not automated in CI | Benchmarks show 7ns/op for raycasting core |
| 39 | 200–5000ms latency tolerance | ✅ Achieved | Tor-mode with 800ms threshold, adaptive prediction, blend time | Full implementation |
| 40 | CI/CD pipeline | ✅ Achieved | `.github/workflows/ci.yml` with build/test/lint/security | Fully functional |
| 41 | 200 features | ✅ Achieved | 200/200 features marked `[x]` in FEATURES.md | — |
| 42 | Skill progression | ✅ Achieved | `SkillProgressionSystem` with XP, levels, genre naming | Multi-file implementation |
| 43 | Crafting system | ✅ Achieved | `CraftingSystem` with workbench, materials, recipes, tool durability | Implemented |
| 44 | Sprite rendering | ✅ Achieved | `pkg/rendering/sprite/` with generator, cache, animation | 97.8% test coverage |
| 45 | Particle effects | ✅ Achieved | `pkg/rendering/particles/` with emitters, renderer | 91.7% test coverage |
| 46 | Lighting system | ✅ Achieved | `pkg/rendering/lighting/` with point/spot/directional lights | 95.5% test coverage |
| 47 | Subtitle system | ✅ Achieved | `pkg/rendering/subtitles/` with text overlay | 88.4% test coverage |
| 48 | Key rebinding | ✅ Achieved | `pkg/input/rebind.go` with config-driven mapping | 98.3% coverage |
| 49 | Party system | ✅ Achieved | `pkg/engine/systems/party.go` with invites, XP/loot sharing | Implemented |
| 50 | Player trading | ✅ Achieved | `pkg/engine/systems/trading.go` with trade protocol, validation | Implemented |

**Overall: 49/50 goals fully achieved (98%), 1 partial (60 FPS requires runtime profiling), 0 missing**

---

## Metrics Summary

### Code Quality (from go-stats-generator)

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines (non-test) | 28,050 | Substantial codebase |
| Total Functions | 501 | Well-structured |
| Total Methods | 2,301 | Method-heavy (good OO separation) |
| Total Structs | 502 | Rich type system |
| Total Interfaces | 8 | Minimal interface use |
| Total Packages | 29 | Good modularity |
| Source Files | 142 | Reasonable |
| Duplication Ratio | 1.70% (905 lines) | ✅ Below 2.0% target |
| Circular Dependencies | 0 | ✅ Excellent |
| Average Complexity | 3.5 | ✅ Good (target <5) |
| High Complexity (>10) | 0 functions | ✅ Excellent |
| Functions >50 lines | 38 (1.4%) | ✅ Acceptable |
| Documentation Coverage | 87.4% | ✅ Above 80% target |

### Top 10 Complex Functions (all below threshold)

| Function | Package | Lines | Cyclomatic | Overall |
|----------|---------|-------|------------|---------|
| `StartInstallation` | housing | 37 | 7 | 10.1 |
| `AddParticipant` | systems | 35 | 7 | 10.1 |
| `CompleteContent` | systems | 32 | 7 | 10.1 |
| `processPlottingCoup` | systems | 31 | 7 | 10.1 |
| `GetUnreadBooks` | systems | 30 | 7 | 10.1 |
| `calculateBounds` | city | 27 | 7 | 10.1 |
| `GetSpotsNear` | systems | 23 | 6 | 9.8 |
| `processHazardEncounter` | systems | 23 | 6 | 9.8 |
| `carveTunnel` | dungeon | 21 | 6 | 9.8 |
| `applyExplosionCrater` | chunk | 20 | 6 | 9.8 |

All functions are at or below cyclomatic complexity 10.

### Test Coverage by Package

| Package | Coverage | Status |
|---------|----------|--------|
| `pkg/procgen/noise` | 100.0% | ✅ Excellent |
| `pkg/rendering/postprocess` | 100.0% | ✅ Excellent |
| `cmd/client` (noebiten) | 100.0% | ✅ Excellent |
| `pkg/input` | 98.3% | ✅ Excellent |
| `pkg/rendering/texture` | 98.2% | ✅ Excellent |
| `pkg/procgen/city` | 98.0% | ✅ Excellent |
| `pkg/rendering/sprite` | 97.8% | ✅ Excellent |
| `pkg/audio/music` | 97.6% | ✅ Excellent |
| `pkg/rendering/lighting` | 95.5% | ✅ Excellent |
| `pkg/world/chunk` | 95.1% | ✅ Excellent |
| `pkg/audio` | 94.3% | ✅ Excellent |
| `pkg/engine/ecs` | 93.8% | ✅ Excellent |
| `pkg/world/housing` | 93.0% | ✅ Excellent |
| `pkg/world/persist` | 93.0% | ✅ Excellent |
| `pkg/procgen/dungeon` | 92.6% | ✅ Excellent |
| `pkg/rendering/particles` | 91.7% | ✅ Excellent |
| `pkg/rendering/raycast` (noebiten) | 90.7% | ✅ Excellent |
| `pkg/audio/ambient` | 90.5% | ✅ Good |
| `pkg/network/federation` | 90.4% | ✅ Good |
| `pkg/world/pvp` | 89.4% | ✅ Good |
| `pkg/procgen/adapters` (noebiten) | 89.2% | ✅ Good |
| `pkg/rendering/subtitles` | 88.4% | ✅ Good |
| `pkg/dialog` | 87.8% | ✅ Good |
| `pkg/companion` | 87.1% | ✅ Good |
| `pkg/engine/components` | 86.0% | ✅ Good |
| `config` | 85.9% | ✅ Good |
| `cmd/server` (noebiten) | 83.8% | ✅ Good |
| `pkg/network` | 83.2% | ✅ Good |
| `pkg/engine/systems` | 78.2% | ✅ Good |

**Average Package Coverage: 91.8%** — Exceeds 70% target across all packages with tests.

---

## Roadmap

### Priority 1: Automated Performance Benchmarking in CI

**Impact**: README claims "60 FPS at 1280×720" — benchmarks exist but aren't validated automatically  
**Effort**: Low (1-2 days)  
**Risk**: Performance regressions could go undetected

The codebase has comprehensive benchmarks but they don't run in CI or report against baselines.

- [x] Add benchmark job to `.github/workflows/ci.yml`:
  ```yaml
  benchmark:
    name: Benchmarks
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - run: go test -tags=noebiten -bench=. -benchmem ./pkg/rendering/raycast/... ./pkg/engine/ecs/... | tee bench.txt
      - uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: bench.txt
  ```
- [x] Add benchmark baseline comparison (fail CI if >10% regression)
- [x] **Validation**: CI shows benchmark results and alerts on significant regressions

### Priority 2: Reduce Code Duplication Further (1.70% → <1.5%)

**Impact**: 51 clone pairs / 905 duplicated lines — still 0.2% above ideal  
**Effort**: Low (2-3 days)  
**Risk**: Bug fixes may not propagate to all clone instances

Top duplication clusters identified by go-stats-generator:

| Location | Lines | Instances | Action |
|----------|-------|-----------|--------|
| `pkg/input/rebind.go` | 21-23 | 4 | Extract `validateBinding(action, key)` |
| `config/load.go` | 11 | 3 | Extract `parseConfigSection(...)` |
| `pkg/engine/systems/vehicle_*.go` | 10-11 | 6 | Extract `applyVehiclePhysics(...)` |
| `pkg/engine/systems/party.go` | 10 | 2 | Extract `distributeToParty(...)` |
| `pkg/engine/systems/trading.go` | 10 | 2 | Extract `validateTradeItem(...)` |

- [ ] Extract `validateBinding` helper from `pkg/input/rebind.go:106` and `:158`
- [ ] Consolidate config section parsing in `config/load.go`
- [ ] Extract shared vehicle physics helper across vehicle_*.go files
- [ ] **Validation**: `go-stats-generator analyze . --skip-tests | grep "Duplication Ratio"` shows <1.5%

### Priority 3: Add Tests for Missing Test Packages

**Impact**: 4 packages show 0% coverage without build tags: `cmd/client`, `cmd/server`, `pkg/procgen/adapters`, `pkg/rendering/raycast`  
**Effort**: Low — tests exist, build tags needed  
**Risk**: Standard `go test ./...` may mislead about coverage

The tests exist for these packages but require the `noebiten` build tag:

- [ ] Update CI to clearly report both standard and build-tag coverage
- [ ] Add comment to README test section explaining build tag requirements
- [ ] Consider adding stub tests that run without `noebiten` tag for basic import verification
- [ ] **Validation**: Coverage report shows all packages with appropriate context

### Priority 4: Address Naming Convention Violations

**Impact**: 23 identifier violations, 19 file violations, 1 package violation  
**Effort**: Low-Medium (3-5 days)  
**Risk**: API changes may affect consumers

Remaining violations:

**Identifier Violations (23)** — mostly in `components` package:
- `VehiclePhysics`, `FactionMembership`, `SkillSchool`, `CityEventEffects`, etc.
- These are acceptable since `components` is a collection package, not domain-specific

**File Violations (19)** — generic names:
- `pkg/engine/systems/constants.go` → could split by subsystem
- `pkg/engine/components/types.go` → could split by category
- `pkg/procgen/adapters/constants.go` → could move to respective adapter files

**Package Violation (1)**:
- `pkg/util` → rename to `pkg/random` or inline the single function

- [ ] Rename `pkg/util` to `pkg/random` (contains only `PseudoRandom` function)
- [ ] Move constants from `constants.go` files to their related source files (37 placement suggestions)
- [ ] **Validation**: `go-stats-generator analyze . --skip-tests | grep "Identifier Violations"` shows <10

### Priority 5: Improve Low-Cohesion Files

**Impact**: 38 low-cohesion files identified by go-stats-generator  
**Effort**: Medium (1-2 weeks)  
**Risk**: Lower developer productivity

Files with 0.00 cohesion score indicate unrelated functions grouped together:

| File | Current Cohesion | Suggested Splits |
|------|------------------|------------------|
| `pkg/engine/components/types.go` (845 LOC) | 0.00 | Split by component category |
| `pkg/procgen/adapters/constants.go` | 0.00 | Move to respective adapter files |
| `pkg/engine/systems/constants.go` | 0.00 | Move to respective system files |
| `pkg/audio/music/constants.go` | 0.00 | Inline into `adaptive.go` |

- [ ] Split `pkg/engine/components/types.go` into category-based files (position.go, combat.go, economy.go, etc.)
- [ ] Distribute constants to their usage sites
- [ ] **Validation**: Average file cohesion improves from 0.50 to >0.65

### Priority 6: Address Oversized Packages

**Impact**: 2 packages identified as potentially oversized  
**Effort**: Medium-High (2-3 weeks)  
**Risk**: Harder to navigate and test

| Package | Files | Exports | Recommendation |
|---------|-------|---------|----------------|
| `systems` | 54 | 1519 | Consider splitting into `systems/combat`, `systems/economy`, etc. |
| `adapters` | 34 | 314 | Consider per-generator subpackages |

- [ ] Evaluate splitting `pkg/engine/systems/` into subdirectories (combat, economy, world, etc.)
- [ ] Document system interdependencies before refactoring
- [ ] **Validation**: Each subpackage has <500 exports

### Priority 7: Document Architecture Decisions

**Impact**: Architecture is sound but not documented  
**Effort**: Medium (1 week)  
**Risk**: Onboarding difficulty for new contributors

- [ ] Create `docs/ARCHITECTURE.md` explaining:
  - ECS pattern and system registration order
  - Chunk streaming lifecycle
  - Network message flow
  - Procgen adapter pattern for V-Series integration
- [ ] Add inline architecture comments to key entry points
- [ ] **Validation**: New contributor can understand system flow from docs alone

---

## Dependency Health

| Dependency | Version | Status |
|------------|---------|--------|
| `github.com/hajimehoshi/ebiten/v2` | v2.9.3 | ✅ Current — Go 1.24+ required |
| `github.com/opd-ai/venture` | v0.0.0-20260321 | ✅ V-Series sibling |
| `github.com/spf13/viper` | v1.19.0 | ✅ Stable |
| `golang.org/x/sync` | v0.17.0 | ✅ Current |
| `golang.org/x/text` | v0.30.0 | ✅ Current |
| `golang.org/x/image` | v0.32.0 | ✅ Current |
| `golang.org/x/sys` | v0.37.0 | ✅ Current |

**Ebitengine v2.9 Notes**:
- Requires Go 1.24+ (project uses Go 1.24.5 ✅)
- Vector graphics API improvements available
- New `DrawTriangles32` APIs available for optimization if needed
- No breaking changes affecting this codebase

---

## Build & Test Commands

```bash
# Build (both pass)
go build ./cmd/client && go build ./cmd/server

# Test with race detection (requires xvfb for Ebiten)
xvfb-run -a go test -race ./...

# Test with build tags for headless packages
go test -tags=noebiten -cover ./pkg/procgen/adapters/...
go test -tags=noebiten -cover ./pkg/rendering/raycast/...
go test -tags=noebiten -cover ./cmd/client/...
go test -tags=noebiten -cover ./cmd/server/...

# Test with coverage
go test -cover ./...

# Static analysis
go vet ./...

# Security scan
govulncheck ./...

# Metrics
go-stats-generator analyze . --skip-tests
```

---

## Key Files Reference

| File | Purpose | Lines |
|------|---------|-------|
| `cmd/client/main.go` | Game client entry, Ebitengine loop | ~300 |
| `cmd/server/main.go` | Server entry, tick loop, system registration | ~140 |
| `pkg/engine/components/types.go` | All component definitions | 845 |
| `pkg/engine/systems/*.go` | 54 ECS system files | ~25,000 total |
| `pkg/world/housing/housing_guild.go` | Guild hall system | 1,237 |
| `pkg/audio/music/adaptive.go` | Adaptive music | 1,041 |
| `pkg/rendering/sprite/generator.go` | Procedural sprite generation | ~675 |
| `pkg/network/server.go` | TCP server, client handling | ~400 |
| `pkg/rendering/raycast/core.go` | DDA raycaster | ~385 |

---

## Summary

Wyrm is a well-architected, extensively tested procedural RPG that achieves **98% of its stated goals**. The codebase demonstrates:

**Strengths**:
- ✅ Build and all tests pass with no errors
- ✅ Average test coverage of 91.8% across all packages
- ✅ Zero circular dependencies
- ✅ Zero high-complexity functions (all ≤10 cyclomatic)
- ✅ Low code duplication (1.70%)
- ✅ Comprehensive ECS with 54 system files
- ✅ Full V-Series integration via 34 adapter files  
- ✅ Functional CI/CD pipeline with race detection, security scanning
- ✅ All 200 claimed features implemented per FEATURES.md
- ✅ Complete rendering pipeline: raycaster + sprites + lighting + particles + subtitles
- ✅ Full multiplayer stack: prediction, lag compensation, federation, Tor-mode
- ✅ Documentation coverage at 87.4%

**Priority Work Items**:

| Priority | Item | Impact | Effort | Status |
|----------|------|--------|--------|--------|
| **P1** | Automated performance benchmarking in CI | High | Low | Not started |
| **P2** | Further reduce code duplication (→<1.5%) | Medium | Low | Not started |
| **P3** | Clarify build-tag test coverage reporting | Medium | Low | Not started |
| **P4** | Address naming convention violations | Low | Medium | Not started |
| **P5** | Improve file cohesion | Low | Medium | Not started |
| **P6** | Consider package restructuring | Low | High | Not started |
| **P7** | Document architecture decisions | Medium | Medium | Not started |

The project is in excellent shape — a mature, feature-complete implementation with strong test coverage and clean architecture. The remaining work is maintenance and polish, not feature development. Compared to notable procedurally generated games like No Man's Sky, Minecraft, and Dwarf Fortress, Wyrm demonstrates an ambitious and largely successful attempt to build a genre-spanning RPG with fully procedural content.

---

*Generated by go-stats-generator v1.0.0 and manual analysis*
*Report date: 2026-03-31*
