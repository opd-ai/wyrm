# Implementation Plan: Complete Wyrm RPG Foundation

## Project Context
- **What it does**: 100% procedurally generated first-person open-world RPG built in Go on Ebitengine — all content generated at runtime from a deterministic seed with no external assets
- **Current goal**: Achieve 60% feature completion (120/200 features) and eliminate critical coverage gaps
- **Estimated Scope**: Large (>15 items requiring work)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| 200 features across 20 categories | ⚠️ 93/200 (46.5%) | Yes |
| Zero external assets | ✅ Achieved | No |
| ECS architecture with 11+ systems | ✅ Achieved (15 systems) | No |
| Five genre themes reshaping gameplay | ⚠️ Partial (terrain/NPCs generic) | Yes |
| 200–5000ms latency tolerance (Tor-mode) | ✅ Achieved | No |
| V-Series generator integration | ⚠️ 0% test coverage on adapters | Yes |
| Crafting & Resources system | ❌ 0% category completion | Yes |
| Combat triangle (melee/ranged/magic) | ⚠️ Melee only (80%) | Yes |
| Cross-server federation | ⚠️ Code exists, not wired to runtime | Yes |
| 60 FPS at 1280×720 | ✅ Achieved | No |

## Metrics Summary

### Codebase Overview
- **Total Lines**: 6,431 across 60 files
- **Packages**: 23
- **Functions**: 223 + 522 methods
- **Structs**: 171
- **Avg Complexity**: 3.5 (excellent, target <10)
- **Avg Function Length**: 9.8 lines (excellent, target <20)

### Complexity Hotspots (Functions with complexity >8.0)
| Function | Package | Complexity | Lines | Risk |
|----------|---------|------------|-------|------|
| `ReportKill` | systems | 8.8 | 25 | Medium |
| `SignTreaty` | systems | 8.8 | 22 | Medium |
| `processSpatialAudio` | systems | 8.8 | 30 | Medium |
| `GetAtTime` | network | 8.8 | 31 | Medium |
| `FindNearestTarget` | systems | 8.8 | 30 | Medium |
| `GenerateDungeonPuzzles` | adapters | 8.8 | 24 | Medium |
| `CanCraft` | adapters | 8.8 | 15 | Medium |
| `determineBiome` | adapters | 8.8 | 20 | Medium |

**All complexity scores under 10** — no critical refactoring needed.

### Duplication
- **Duplicated Lines**: 0
- **Duplication Ratio**: 0%

### Documentation Coverage
- **Package coverage**: 100%
- **Function coverage**: 97.2%
- **Type coverage**: 81.4%
- **Method coverage**: 86.1%
- **Overall**: 86.6% (good, target >70%)

### Test Coverage by Package
| Package | Coverage | Status |
|---------|----------|--------|
| `pkg/engine/ecs` | 100% | ✅ |
| `pkg/procgen/city` | 100% | ✅ |
| `pkg/procgen/noise` | 100% | ✅ |
| `pkg/rendering/postprocess` | 100% | ✅ |
| `pkg/audio/music` | 95.9% | ✅ |
| `pkg/world/housing` | 94.8% | ✅ |
| `pkg/rendering/texture` | 93.8% | ✅ |
| `config` | 92.9% | ✅ |
| `pkg/procgen/dungeon` | 91.7% | ✅ |
| `pkg/engine/components` | 91.4% | ✅ |
| `pkg/dialog` | 90.9% | ✅ |
| `pkg/network/federation` | 90.4% | ✅ |
| `pkg/world/persist` | 89.5% | ✅ |
| `pkg/world/pvp` | 89.4% | ✅ |
| `pkg/audio/ambient` | 87.0% | ✅ |
| `pkg/audio` | 85.1% | ✅ |
| `pkg/engine/systems` | 80.9% | ✅ |
| `pkg/network` | 80.7% | ✅ |
| `pkg/companion` | 78.8% | ✅ |
| `pkg/world/chunk` | 98.0% | ✅ |
| **`pkg/procgen/adapters`** | **0%** | ❌ |
| `cmd/client` | 0% | ⚠️ |
| `pkg/rendering/raycast` | 0% (requires Ebiten) | ⚠️ |

### Package Coupling Analysis
| Package | Coupling Score | Cohesion | Risk |
|---------|----------------|----------|------|
| `adapters` | 10 | 2.15 | High — critical integration layer |
| `main` | 6.5 | 2.6 | Medium |
| `systems` | 1.5 | 1.8 | Low |
| All others | <1 | Varies | Low |

### Feature Gap Analysis (FEATURES.md)
| Category | Implemented | Total | Gap |
|----------|-------------|-------|-----|
| Crafting & Resources | 0 | 10 | **10 features** |
| Cities & Structures | 3 | 10 | 7 features |
| NPCs & Social | 3 | 10 | 7 features |
| Vehicles & Mounts | 3 | 10 | 7 features |
| Weather & Environment | 3 | 10 | 7 features |
| Combat System | 8 | 10 | 2 features |

---

## Implementation Steps

### Step 1: Add Test Coverage for V-Series Adapters [COMPLETED]

- **Deliverable**: `pkg/procgen/adapters/adapters_test.go` with comprehensive tests for all 16 adapter types (Entity, Faction, Quest, Dialog, Terrain, Building, Vehicle, Magic, Skills, Recipe, Narrative, Puzzle, Item, Environment, Furniture, and doc)
- **Dependencies**: None
- **Goal Impact**: Addresses 0% coverage on the critical V-Series integration layer (124 functions, 2,788 lines); prevents silent breakage of procedural generation foundation
- **Acceptance**: Test coverage ≥70% for `pkg/procgen/adapters/`
- **Validation**: `xvfb-run -a go test -cover ./pkg/procgen/adapters/...` shows 82.4%
- **Status**: ✅ Completed - 82.4% coverage achieved

**Test categories to implement:**
1. Determinism verification (same seed → identical output)
2. Genre parameter routing (all 5 genres produce genre-appropriate content)
3. Error handling for invalid inputs (zero seed, empty genre string)
4. Integration with Venture dependency (imports resolve, types convert correctly)

**Files to create:**
- `pkg/procgen/adapters/adapters_test.go` — main test file
- Use `//go:build !ebitentest` tag for headless execution

---

### Step 2: Implement Crafting System Foundation [COMPLETED]

- **Deliverable**: Complete Crafting & Resources category from 0% to ≥50% (5 features)
- **Dependencies**: Step 1 (adapters tested for recipe generation)
- **Goal Impact**: Fills largest feature gap; `RecipeAdapter` already exists but has no gameplay integration
- **Acceptance**: FEATURES.md Crafting category shows ≥5 `[x]` marks
- **Validation**: `grep -c '\[x\]' FEATURES.md` increases by ≥5; player can gather material and craft at workbench
- **Status**: ✅ Completed - 7/10 crafting features implemented (70%); Material, ResourceNode, Workbench, CraftingState, Tool, RecipeKnowledge components added; CraftingSystem with gathering, quality tiers, tool durability, recipe discovery, respawning

**Components added in `pkg/engine/components/types.go`:**
- Material (ResourceType, Quantity, Quality, Rarity)
- ResourceNode (respawning resource locations)
- Workbench (crafting stations with type-specific recipe support)
- CraftingState (tracks ongoing crafting)
- Tool (durability, speed, quality bonuses)
- RecipeKnowledge (recipe discovery tracking)

**System added in `pkg/engine/systems/crafting.go`:**
- CraftingSystem.Update() advances crafting progress and resource respawning
- GatherResource() with tool efficiency and skill bonuses
- StartCraft()/CancelCraft() for workbench interaction
- Recipe discovery via DiscoverRecipe()/ProgressRecipeDiscovery()
- Quality tier calculation (Common/Uncommon/Rare/Epic/Legendary)

**Features enabled:**
- [x] Material gathering
- [x] Workbench system
- [x] Recipe discovery (via adapter)
- [x] Quality tiers (skill-based)
- [x] Tool durability
- [x] Resource respawning
- [x] Rare materials (quality/rarity system)

---

### Step 3: Implement Ranged Combat [COMPLETED]

- **Deliverable**: Ranged weapon system completing the combat triangle
- **Dependencies**: None (existing `CombatSystem` and `Weapon` component sufficient)
- **Goal Impact**: Advances Combat category from 80% to 90%; required for balanced gameplay
- **Acceptance**: Player can fire ranged weapon; projectile deals damage on collision
- **Validation**: `go test ./pkg/engine/systems/... -run TestRangedCombat` passes
- **Status**: ✅ Completed - ProjectileSystem with movement, collision, lifetime; CombatSystem.InitiateRangedAttack(); Projectile, Mana, Spell, Spellbook components added

**Components added in `pkg/engine/components/types.go`:**
- Projectile (velocity, damage, lifetime, hit radius, pierce)
- Mana (current, max, regen rate)
- SpellEffect (type, magnitude, duration)
- Spell (mana cost, cooldown, effect type, AoE)
- Spellbook (spell collection, active spell)

**System added in `pkg/engine/systems/ranged_combat.go`:**
- ProjectileSystem.Update() handles movement, collision, lifetime cleanup
- SpawnProjectile() creates projectile entities with physics
- InitiateRangedAttack() for ranged weapon firing
- IsRangedWeapon() weapon type detection
- getRangedSkillModifier() for archery/firearms skill bonuses

---

### Step 4: Implement Magic Combat

- **Deliverable**: Magic/ability system completing the combat triangle
- **Dependencies**: Step 3 (reuses projectile infrastructure for spell projectiles)
- **Goal Impact**: Advances Combat category to 100%; enables genre-appropriate combat (Fantasy spells, Sci-Fi tech, Cyberpunk hacks)
- **Acceptance**: Player can cast spell consuming mana; spell applies effect to target
- **Validation**: `go test ./pkg/engine/systems/... -run TestMagicCombat` passes

**Components to add:**
```go
type Mana struct {
    Current   float64
    Max       float64
    RegenRate float64
}

type SpellEffect struct {
    EffectType string  // "damage", "heal", "buff", "debuff"
    Magnitude  float64
    Duration   float64
    Remaining  float64
}

type Spell struct {
    Name       string
    ManaCost   float64
    Cooldown   float64
    LastCast   float64
    EffectType string
    Magnitude  float64
}
```

**System integration:**
- Use existing `MagicAdapter` to generate spells at character creation
- Add `ManaSystem` for regeneration per tick
- Extend `CombatSystem` for spell casting input
- Add area-of-effect targeting via spatial query

---

### Step 5: Add Genre-Specific Terrain Biomes [COMPLETED]

- **Deliverable**: Terrain generation differentiated by genre in `pkg/procgen/adapters/terrain.go` and `pkg/rendering/texture/patterns.go`
- **Dependencies**: Step 1 (adapter tests verify genre routing)
- **Goal Impact**: Delivers on "Five genre themes reshape every player-facing system" promise; terrain is most visible genre differentiator
- **Acceptance**: Visual inspection of 5 genre seeds shows distinct biome distributions
- **Validation**: Test verifying genre textures produce different patterns (`TestGenreTexturesAreDifferent` passes)
- **Status**: ✅ Completed - Genre biome distributions in terrain.go + genre-specific texture patterns in patterns.go (grid for sci-fi/cyberpunk, voronoi for horror, distortion for post-apocalyptic, layered for fantasy)

**Biome weight tables to implement:**

| Genre | Primary Biome | Secondary | Tertiary | Rare |
|-------|--------------|-----------|----------|------|
| Fantasy | Forest (40%) | Mountain (30%) | Plains (20%) | Lake (10%) |
| Sci-Fi | Crater (40%) | Tech-Structure (30%) | Alien-Flora (20%) | Mining-Site (10%) |
| Horror | Swamp (40%) | Dead-Forest (30%) | Fog-Zone (20%) | Graveyard (10%) |
| Cyberpunk | Urban (60%) | Industrial (25%) | Neon-District (15%) | — |
| Post-Apoc | Wasteland (50%) | Ruins (30%) | Radiation-Zone (15%) | Shanty (5%) |

**Files to modify:**
- `pkg/procgen/adapters/terrain.go` — add `genreBiomeWeights` map and modify `determineBiome()`

---

### Step 6: Wire Federation to Server Runtime

- **Deliverable**: Server federation enabled via configuration, allowing cross-server player transfer
- **Dependencies**: None
- **Goal Impact**: Delivers on "Cross-server federation" multiplayer promise; code exists (90.4% tested) but is never instantiated
- **Acceptance**: Two servers with federation enabled; player transfers between them
- **Validation**: Start two servers with `federation.enabled: true`; call transfer endpoint; player appears on peer server

**Configuration to add in `config.yaml`:**
```yaml
federation:
  enabled: false
  node_id: ""       # Auto-generated UUID if empty
  peers: []         # List of peer server addresses
  gossip_interval: 5s
```

**Code changes in `cmd/server/main.go`:**
```go
if cfg.Federation.Enabled {
    node := federation.NewFederationNode(cfg.Federation.NodeID)
    for _, peer := range cfg.Federation.Peers {
        node.ConnectPeer(peer)
    }
    go node.StartGossip(cfg.Federation.GossipInterval)
}
```

**Wire `CrossServerTransfer` into player disconnect handling to persist inventory/position for peer pickup.**

---

### Step 7: Add NPC Memory and Relationships

- **Deliverable**: NPCs remember player interactions and adjust behavior accordingly
- **Dependencies**: None
- **Goal Impact**: Advances NPCs & Social category from 30% to 50%+; core RPG immersion depends on reactive NPCs
- **Acceptance**: NPC remembers past attack; disposition decreases; future dialog options change
- **Validation**: `go test ./pkg/engine/systems/... -run TestNPCMemory` passes

**Components to add:**
```go
type NPCMemory struct {
    PlayerInteractions map[ecs.Entity][]MemoryEvent
    LastSeen           map[ecs.Entity]int64  // Unix timestamp
    Disposition        map[ecs.Entity]float64 // -1.0 to +1.0
}

type MemoryEvent struct {
    EventType  string  // "gift", "attack", "quest_complete", "dialog"
    Timestamp  int64
    Impact     float64 // Disposition change
}
```

**System to add in `pkg/engine/systems/npc_memory.go`:**
- Records player interactions with NPCs
- Decays old memories over time (configurable window)
- Updates disposition based on weighted interaction history
- Feeds disposition into `DialogAdapter` for response filtering

---

### Step 8: Expand City Structure Features

- **Deliverable**: City interiors and POIs from Cities & Structures category
- **Dependencies**: Step 5 (terrain genres may affect city placement)
- **Goal Impact**: Advances Cities & Structures from 30% to 60%+; buildings feel empty without interiors
- **Acceptance**: Player can enter shop building and see interior with shelves/counters
- **Validation**: `GenerateCity()` produces buildings with `Interior` component; raycaster renders interior walls

**Features to enable:**
- Shop building interiors (uses existing `BuildingAdapter` + `FurnitureAdapter`)
- Points of interest markers (component with map icon type)
- Government buildings (faction headquarters)

**Components to add:**
```go
type Interior struct {
    ParentBuilding ecs.Entity
    Rooms          []Room
    Furniture      []ecs.Entity
}

type POIMarker struct {
    IconType string // "shop", "quest", "danger", "guild"
    Visible  bool
}
```

---

### Step 9: Add Vehicle Physics and Cockpit View

- **Deliverable**: Vehicles with steering, acceleration, fuel; first-person cockpit rendering
- **Dependencies**: None
- **Goal Impact**: Advances Vehicles & Mounts from 30% to 60%+; vehicle system exists but lacks gameplay depth
- **Acceptance**: Player enters vehicle; first-person view changes to cockpit; WASD controls steering/acceleration
- **Validation**: `go test ./pkg/engine/systems/... -run TestVehiclePhysics` passes

**System enhancements in `pkg/engine/systems/vehicle.go`:**
- Add `steeringAngle`, `currentSpeed` fields to `VehicleState` component
- Implement acceleration/deceleration based on input
- Implement steering with turning radius
- Track fuel consumption per distance traveled
- Add cockpit view flag for render system

---

### Step 10: Weather Gameplay Effects

- **Deliverable**: Weather conditions affect gameplay mechanics (movement, visibility, combat)
- **Dependencies**: None
- **Goal Impact**: Advances Weather & Environment from 30% to 50%+; weather feels cosmetic without gameplay impact
- **Acceptance**: Rain reduces visibility in raycaster; snow slows movement; combat accuracy affected
- **Validation**: `go test ./pkg/engine/systems/... -run TestWeatherEffects` passes

**System enhancements in `pkg/engine/systems/weather.go`:**
- Add `GetWeatherModifiers()` method returning `{VisibilityMult, MovementMult, AccuracyMult}`
- Feed modifiers into `RenderSystem` (draw distance), `MovementSystem` (speed), `CombatSystem` (hit chance)

**Weather effect table:**
| Condition | Visibility | Movement | Accuracy |
|-----------|------------|----------|----------|
| Clear | 1.0 | 1.0 | 1.0 |
| Rain | 0.7 | 0.9 | 0.85 |
| Storm | 0.4 | 0.7 | 0.6 |
| Fog | 0.3 | 1.0 | 0.7 |
| Snow | 0.6 | 0.75 | 0.9 |
| Blizzard | 0.2 | 0.5 | 0.5 |

---

## Summary: Expected Outcome

After completing all 10 steps:

| Metric | Current | Target | Change |
|--------|---------|--------|--------|
| Feature completion | 93/200 (46.5%) | 120/200 (60%) | +27 features |
| Adapter test coverage | 0% | ≥70% | Critical gap closed |
| Crafting category | 0% | ≥50% | Major gap addressed |
| Combat category | 80% | 100% | Triangle complete |
| NPCs & Social | 30% | ≥50% | Core RPG depth |
| Vehicles & Mounts | 30% | ≥60% | Gameplay depth |
| Cities & Structures | 30% | ≥60% | World richness |
| Weather & Environment | 30% | ≥50% | Immersion |
| Federation | Code-only | Runtime-enabled | Multiplayer ready |

**Dependency Graph:**
```
Step 1 (Adapter Tests) ─────┬──▶ Step 2 (Crafting)
                            │
                            └──▶ Step 5 (Genre Terrain)

Step 3 (Ranged Combat) ─────▶ Step 4 (Magic Combat)

Steps 6-10 are independent and can be parallelized.
```

**Recommended Execution Order:**
1. Step 1 (unblocks Steps 2 and 5)
2. Steps 3 + 6 + 7 (in parallel)
3. Steps 4 + 5 + 8 (in parallel, after 1 and 3)
4. Steps 9 + 10 (in parallel)
5. Step 2 (after Step 1 complete)

---

## Validation Commands Reference

```bash
# Overall test coverage
go test -cover ./...

# Adapter coverage specifically
go test -cover ./pkg/procgen/adapters/...

# Feature count
grep -c '\[x\]' FEATURES.md
grep -c '\[ \]' FEATURES.md

# Complexity check (should remain <10 for all functions)
go-stats-generator analyze . --skip-tests --format json | jq '[.functions[] | select(.complexity.overall > 10)]'

# Build verification
go build ./cmd/client && go build ./cmd/server

# Race detection
go test -race ./...

# Static analysis
go vet ./...
```

---

## External Factors (from Research)

### Ebitengine v2.9 Compatibility
- **Go 1.24+ required** — project already uses 1.24.5 ✅
- **Vector API deprecations** — `vector.AppendVerticesAndIndices*` deprecated; migrate to `vector.FillPath()`/`vector.StrokePath()` if used
- **No blocking issues identified** for current codebase

### V-Series Ecosystem
- Venture dependency (`opd-ai/venture`) is stable at v0.0.0-20260321140920
- 16 adapters successfully import from Venture's `pkg/procgen/*` packages
- No breaking changes detected in dependent generators

### Community Status
- No active GitHub issues or discussions requiring attention
- Project is nascent but following opd-ai procedural generation conventions
