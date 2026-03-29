# Implementation Gaps — 2026-03-29

This document catalogs the gaps between Wyrm's stated goals and its current implementation. Each gap represents work needed to achieve the project's documented objectives.

---

## Gap 1: Server Build Constraint Prevents Normal Compilation

- **Stated Goal**: README states "Build: `go build ./cmd/server`" and describes the server as "authoritative game server entry point".
- **Current State**: `cmd/server/main.go` line 1 has `//go:build ebitentest` build tag. Running `go build ./cmd/server` fails with "build constraints exclude all Go files in /home/user/go/src/github.com/opd-ai/wyrm/cmd/server".
- **Impact**: Server cannot be built using documented instructions. Developers following README will fail. CI may pass only because it uses special tags.
- **Closing the Gap**:
  1. Change line 1 of `cmd/server/main.go` from `//go:build ebitentest` to `//go:build !noebiten` (matching client pattern)
  2. OR remove the build tag entirely if server doesn't require Ebiten
  3. Update CI to test `go build ./cmd/server` without special tags
  4. **Validation**: `go build ./cmd/server && ./server --help` succeeds

---

## Gap 2: V-Series Adapters Test Coverage (0%)

- **Stated Goal**: Project mandates "≥40% per package (≥30% for Ebiten-dependent packages)" per copilot-instructions.md quality standards.
- **Current State**: `pkg/procgen/adapters/` contains 16 adapter files (3,221 LOC, 124 functions) with 0% coverage in standard `go test ./...`. Tests exist in `adapters_test.go` but require `ebitentest` build tag. This is the critical V-Series integration layer.
- **Impact**: Integration bugs in adapters could silently break faction generation, NPC spawning, quest creation, terrain biomes — all core gameplay systems. The adapters import from `opd-ai/venture` and translate to Wyrm's ECS, making them high-risk.
- **Closing the Gap**:
  1. Refactor adapters to delay Ebiten imports (use interfaces or build-tag-split files)
  2. OR ensure CI runs `xvfb-run go test -tags=ebitentest ./pkg/procgen/adapters/...`
  3. Add determinism tests: same seed must produce identical output across 3 runs
  4. Add error handling tests: zero seed, empty genre, invalid depth
  5. **Validation**: `go test -cover ./pkg/procgen/adapters/...` shows ≥70% without special environment

---

## Gap 3: Raycast Renderer Test Coverage (0%)

- **Stated Goal**: Project emphasizes "first-person raycaster at 60 FPS" as a Phase 2 completion criterion. Core rendering is critical path.
- **Current State**: `pkg/rendering/raycast/` has only stub test files. `core.go` (385 LOC) contains the DDA raycasting algorithm with no automated tests. `draw.go` requires Ebiten for rendering.
- **Impact**: Raycaster bugs (wall clipping, texture coordinate errors, floor/ceiling artifacts) would be undetected until visual inspection. Changes could break rendering without any test failure.
- **Closing the Gap**:
  1. Create `core_test.go` with `//go:build noebiten` tag for headless testing
  2. Test `CastRay()` with known wall configurations and expected intersection points
  3. Test `calculateWallDistance()` with edge cases (parallel walls, corners)
  4. Test texture coordinate calculation for seam correctness
  5. **Validation**: `go test -tags=noebiten ./pkg/rendering/raycast/...` passes with ≥50% coverage

---

## Gap 4: Feature Target (92/200 = 46%)

- **Stated Goal**: README claims "Wyrm targets 200 features across 20 categories" with FEATURES.md tracking completion.
- **Current State**: 92 features implemented (46% per FEATURES.md). By category:
  - Crafting & Resources: **0%** (0/10) — Major gameplay pillar missing
  - Cities & Structures: **30%** (3/10)
  - NPCs & Social: **30%** (3/10) — NPC memory, relationships missing
  - Vehicles & Mounts: **30%** (3/10) — Physics, cockpit view missing
  - Weather & Environment: **30%** (3/10) — Gameplay effects missing
- **Impact**: The game is a technical demo, not the "200-feature RPG" described. Players expecting Elder Scrolls-inspired depth will find major systems absent.
- **Closing the Gap**:
  1. **Priority 1**: Implement crafting system (Material component, Workbench component, CraftingSystem) — uses existing RecipeAdapter
  2. **Priority 2**: Add NPC memory and relationships (NPCMemory component, memory events, disposition tracking)
  3. **Priority 3**: Complete combat triangle (ranged Projectile component, magic Mana/SpellEffect components)
  4. **Priority 4**: Add vehicle physics (steering, acceleration curves, fuel consumption rates)
  5. Track progress: `grep -c '\[x\]' FEATURES.md` should increase 5-10 features/sprint
  6. **Validation**: FEATURES.md shows 120+ features (60%) within 3 months

---

## Gap 5: Genre Terrain Differentiation

- **Stated Goal**: README claims "Five genre themes reshape every player-facing system" with distinct visual palettes per genre.
- **Current State**: `pkg/procgen/adapters/terrain.go` defines genre-specific biome distributions (Fantasy: 40% forest, Cyberpunk: 50% urban, etc.) but texture generation in `pkg/rendering/texture/` uses the same palettes regardless of genre. Visual distinction is minimal.
- **Impact**: Players won't perceive genre uniqueness in terrain — a core differentiator. Fantasy and Cyberpunk worlds look similar despite different biome types.
- **Closing the Gap**:
  1. Add `GenreTextureParams` to `pkg/rendering/texture/` with palette overrides per genre
  2. Fantasy: warm gold/green; Sci-Fi: cool blue/white; Horror: desaturated grey; Cyberpunk: neon pink/cyan; Post-Apoc: sepia/orange
  3. Apply genre palette to `generateNoiseTexture()` based on biome and genre
  4. Add post-processing genre filters (already in `pkg/rendering/postprocess/` but not wired to terrain)
  5. **Validation**: Screenshot comparison of 5 genres shows visually distinct terrain colors

---

## Gap 6: Ranged and Magic Combat Missing

- **Stated Goal**: README promises "First-person melee, ranged, and magic combat with timing-based blocking".
- **Current State**: `CombatSystem` in `pkg/engine/systems/combat.go` implements melee only (260 LOC). No projectile spawning, no mana system, no spell effects. FEATURES.md shows Combat System at 80% but ranged/magic specifically unchecked.
- **Impact**: Combat is limited to melee — 1/3 of the promised combat triangle. Players cannot play archers, mages, or use genre-appropriate ranged weapons (guns in Cyberpunk, spells in Fantasy).
- **Closing the Gap**:
  1. Add `Projectile` component: OwnerID, Velocity, Damage, Lifetime, ProjectileType
  2. Add projectile spawning in `CombatSystem` when weapon type is "ranged"
  3. Add projectile movement and collision system or extend CombatSystem
  4. Add `Mana` component: Current, Max, RegenRate
  5. Add `SpellEffect` component: Type, Duration, Magnitude, Area
  6. Wire existing `MagicAdapter` to generate runtime spells
  7. **Validation**: Player can fire ranged weapon and cast spell; both deal damage

---

## Gap 7: Client Entry Point Untested

- **Stated Goal**: Quality standards require testable code; `cmd/client/main.go` is the primary user-facing entry point.
- **Current State**: `cmd/client/main.go` (305 LOC) has 0% test coverage. Contains player input handling, chunk map updates, audio initialization, network connection logic — all untested.
- **Impact**: Bugs in player controls, chunk loading, or audio could ship undetected. Refactoring carries high regression risk.
- **Closing the Gap**:
  1. Extract pure functions from main.go that can be tested without Ebiten
  2. Create `cmd/client/main_test.go` with tests for:
     - `heightToWallType()` — already pure, easy to test
     - `processMovementInput()` — mock Position component
     - `updateChunkMap()` — mock ChunkManager
  3. Use dependency injection for testability
  4. **Validation**: `go test ./cmd/client/...` passes with ≥40% coverage

---

## Gap 8: Magic Numbers Technical Debt (2,365)

- **Stated Goal**: Maintainable code with named constants; project uses `pkg/engine/systems/constants.go` for some constants.
- **Current State**: go-stats-generator detects 2,365 magic numbers. Top offenders:
  - `pkg/procgen/adapters/` — generation depth values, probability weights
  - `pkg/engine/systems/combat.go` — damage multipliers, range values
  - `pkg/audio/music/adaptive.go` — frequency tables, timing values
- **Impact**: Code is harder to tune, understand, and maintain. Related values (e.g., all damage multipliers) are scattered, making balance changes error-prone.
- **Closing the Gap**:
  1. Extract combat constants from `combat.go` to named values or `constants.go`
  2. Extract audio frequencies to `audio_constants.go` or similar
  3. Extract adapter generation parameters to config structs
  4. Goal: reduce magic numbers to <1,500
  5. **Validation**: `go-stats-generator analyze . --skip-tests | grep "Magic Numbers"` shows <1,500

---

## Gap 9: NPC Memory and Relationships

- **Stated Goal**: README promises "NPC memory, relationships, gossip networks, and emotional states" and "NPCs remember player actions".
- **Current State**: `Schedule` component exists for NPC daily activities, but no `NPCMemory` component. NPCs have no persistent memory of player interactions, no disposition tracking, no relationship system. FEATURES.md shows "NPC memory system" and "NPC relationships" as unchecked.
- **Impact**: NPCs feel static. Attacking an NPC, completing quests for them, or trading has no lasting effect on how they treat the player — a core RPG expectation.
- **Closing the Gap**:
  1. Add `NPCMemory` component:
     ```go
     type NPCMemory struct {
         PlayerInteractions map[uint64][]MemoryEvent
         LastSeen           map[uint64]time.Time
         Disposition        map[uint64]float64 // -1 to +1
     }
     ```
  2. Add `MemoryEvent` struct with type (gift, attack, quest, dialog), timestamp, impact value
  3. Create `NPCMemorySystem` that records interactions and decays old memories
  4. Integrate with `DialogAdapter` to affect available dialog options based on disposition
  5. **Validation**: NPC remembers attack (disposition drops); affects future dialog

---

## Gap 10: Crafting System (0%)

- **Stated Goal**: README promises "Crafting via material gathering, workbench minigames, and recipe discovery". FEATURES.md lists 10 crafting features.
- **Current State**: `pkg/procgen/adapters/recipe.go` exists and can generate recipes via V-Series, but no gameplay implementation. No Material component, no Workbench component, no CraftingSystem. FEATURES.md shows Crafting & Resources at **0%** — the only category with zero implementation.
- **Impact**: Crafting is a core RPG loop (gather → craft → equip). Its complete absence means players cannot engage in meaningful item progression outside of loot drops.
- **Closing the Gap**:
  1. Add `Material` component: ResourceType, Quantity, Quality
  2. Add `Workbench` component: SupportedRecipeTypes, CurrentRecipe, Progress
  3. Create `CraftingSystem` that:
     - Checks player has required materials (via RecipeAdapter)
     - Validates workbench proximity via spatial query
     - Applies skill-based quality modifiers
     - Creates crafted item entity
  4. Add crafting UI showing available recipes and requirements
  5. **Validation**: Player can craft a basic item at a workbench

---

## Summary: Gap Closure Priority

| Priority | Gap | Impact | Effort | Dependencies |
|----------|-----|--------|--------|--------------|
| **P0** | Server build constraint | Blocks basic usage | Low (1 line change) | None |
| **P1** | V-Series adapter tests | Blocks confidence in core systems | Medium (1 week) | CI changes |
| **P1** | Raycast tests | Blocks confidence in rendering | Medium (1 week) | None |
| **P2** | Crafting system | Major feature gap | High (2-3 weeks) | RecipeAdapter |
| **P2** | Ranged/magic combat | Incomplete combat | High (2-3 weeks) | Projectile physics |
| **P3** | NPC memory | Incomplete social simulation | Medium (2 weeks) | DialogAdapter |
| **P3** | Genre terrain visuals | Incomplete differentiation | Medium (1 week) | Texture system |
| **P4** | Magic numbers | Tech debt | Low (ongoing) | None |
| **P4** | Feature target (200) | Scope completion | High (6+ months) | All above |

---

*Generated by comparing README.md, ROADMAP.md, and FEATURES.md claims against codebase implementation.*
