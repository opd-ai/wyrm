# Wyrm Critical Gaps Analysis — 2026-03-31

This document catalogs the gaps between Wyrm's stated goals and its current implementation, organized by player impact. Each gap identifies what is broken or missing, why it matters, and what is needed to fix it.

> **Previous code quality gaps** (FEATURES.md sync, test coverage, complexity, duplication, naming) are all resolved and omitted from this revision. This document now focuses on **playability and integration gaps**.

---

## Category 1: System Integration Failures

### Gap 1.1: 77% of ECS Systems Never Execute

**Severity: CRITICAL** · **Impact: All gameplay beyond movement is non-functional**

Of 57 defined system types in `pkg/engine/systems/`, only 13 are registered at runtime:
- **Server:** 10 systems registered in `cmd/server/main.go:102-113`
- **Client:** 3 systems registered in `cmd/client/main.go:242-246`
- **Unregistered:** 44 systems with complete implementations that never run

All 44 unregistered systems have constructors, full `Update()` implementations with component queries, and test coverage. They simply are not instantiated or registered via `world.RegisterSystem()`.

**Key unregistered systems and their player-facing impact:**

| System | Player Impact of Absence |
|--------|--------------------------|
| CraftingSystem | Cannot craft any items |
| MagicSystem | Cannot cast spells; no mana regeneration |
| StealthSystem | Cannot sneak; no detection mechanics |
| ProjectileSystem | No ranged combat (arrows, projectiles) |
| NPCPathfindingSystem | NPCs stand motionless; never move to activity locations |
| NPCMemorySystem | NPCs don't remember player; no reputation tracking |
| NPCNeedsSystem | NPCs have no needs (hunger, energy, social); feel lifeless |
| NPCOccupationSystem | NPCs don't perform their jobs |
| EmotionalStateSystem | NPCs have no emotional reactions |
| GossipSystem | NPCs don't share information |
| GuardPursuitSystem | Guards don't pursue criminals |
| SkillProgressionSystem | No skill leveling or XP |
| FactionRankSystem | Cannot advance in any faction |
| DialogConsequenceSystem | Dialog choices have no lasting effects |
| HazardSystem | Environmental hazards do no damage |
| PartySystem | Cannot form or manage parties |
| TradingSystem | Cannot trade with other players |
| CityBuildingSystem | Buildings don't open/close; no shop schedules |
| CityEventSystem | No dynamic city events (festivals, raids, plagues) |
| IndoorOutdoorSystem | Weather affects entities while indoors |

**Location:** `cmd/client/main.go:242-246`, `cmd/server/main.go:102-113`
**Fix:** Add `world.RegisterSystem()` calls with proper initialization for all 44 systems.

---

### Gap 1.2: Client Has Almost No Game Logic

**Severity: CRITICAL** · **Impact: Client is a rendering shell with no gameplay**

The client registers only 3 systems:
1. `RenderSystem` — Minimal stub that reads player Position (no actual rendering logic in Update)
2. `AudioSystem` — Full spatial audio calculation (but no AudioListener or AudioSource entities exist on client)
3. `WeatherSystem` — Weather cycling (but `&systems.WeatherSystem{}` uses zero-value init, not `NewWeatherSystem()`)

The client relies entirely on the server for game logic, but there is no protocol to synchronize server state to the client. The client runs its own isolated ECS world with only a player entity.

**Location:** `cmd/client/main.go:242-246`
**Fix:** Either register client-side systems for single-player mode, or implement the entity synchronization protocol.

---

### Gap 1.3: Client WeatherSystem Improperly Initialized

**Severity: MODERATE** · **Impact: Weather on client uses zero-value defaults**

In `cmd/client/main.go:245`:
```go
world.RegisterSystem(&systems.WeatherSystem{})  // zero-value init
```

But `WeatherSystem` requires genre and duration parameters via `NewWeatherSystem(genre, duration)`. The server correctly uses `NewWeatherSystem(cfg.Genre, 300.0)` at line 112.

**Location:** `cmd/client/main.go:245`
**Fix:** `world.RegisterSystem(systems.NewWeatherSystem(cfg.Genre, 300.0))`

---

## Category 2: Missing Core Gameplay Features

### Gap 2.1: No Player Collision Detection

**Severity: HIGH** · **Impact: Player walks through all walls and terrain**

Player movement in `cmd/client/main.go:123-153` directly modifies Position X/Y without checking the world map for walls. The raycaster renders walls but the movement code does not test against them.

**Location:** `cmd/client/main.go:123-153` (`processMovementInput`, `processStrafeInput`)
**Fix:** Before applying movement, check `worldMap[newY][newX]` for wall cells. Reject movement into cells where `heightToWallType() > 0`.

---

### Gap 2.2: No HUD or UI System

**Severity: HIGH** · **Impact: Zero visual feedback for player state**

The only screen overlay is a single debug text line:
```go
ebitenutil.DebugPrint(screen, fmt.Sprintf("Wyrm [%s] %s", g.cfg.Genre, status))
```

Missing UI elements:
- Health/mana/stamina bars (player has Health component but it is never displayed)
- Minimap or compass
- Inventory screen
- Quest log
- Dialog interface
- Pause/settings menu
- Genre selection screen
- Character creation screen

**Location:** `cmd/client/main.go:157-171` (`Draw` method)
**Fix:** Implement an overlay UI system rendered after the raycaster in `Draw()`.

---

### Gap 2.3: No Interaction System (E Key)

**Severity: HIGH** · **Impact: Cannot interact with any world object**

The `pkg/input/` package defines 40+ bindable actions including `interact` (default: E key), but the key binding system is never imported or used by the client. The client only handles WASD/QE movement directly via Ebiten key checks.

**Location:** `cmd/client/main.go:108-153` (input handling)
**Fix:** Import `pkg/input`, use `Rebinder` for all key checks, implement interaction raycasting to find the entity the player is looking at.

---

### Gap 2.4: No Save/Load Integration

**Severity: HIGH** · **Impact: All progress lost when game exits**

`pkg/world/persist/` provides a `PersistenceManager` with save/load functionality and test coverage, but it is never imported by either `cmd/client/` or `cmd/server/`.

**Location:** `pkg/world/persist/`
**Fix:** Call `PersistenceManager.Save()` on server shutdown and `PersistenceManager.Load()` on startup.

---

### Gap 2.5: No Character Creation or Genre Selection

**Severity: MODERATE** · **Impact: Player starts with hardcoded position and stats; no genre choice UI**

Player entity is created with hardcoded values:
```go
player := world.CreateEntity()
world.AddComponent(player, &components.Position{X: 8.5, Y: 8.5, Z: 0})
world.AddComponent(player, &components.Health{Current: 100, Max: 100})
```

No Mana, Skills, Inventory, Faction, Reputation, or Stealth components are added. Genre is read from config.yaml but there is no in-game UI to select it.

**Location:** `cmd/client/main.go:230-239` (`createPlayerEntity`)
**Fix:** Add a genre selection screen before game start. Create player with full component set (Mana, Skills, Inventory, etc.).

---

### Gap 2.6: No Menu System (Pause, Settings, Quit)

**Severity: MODERATE** · **Impact: No way to pause, change settings, or quit gracefully from in-game**

The only exit mechanism is OS window close (Ebiten default). There is no:
- Pause menu (Escape key)
- Settings screen (keybinds, audio, graphics)
- Quit confirmation dialog
- Return-to-title option

**Location:** `cmd/client/main.go` (missing entirely)
**Fix:** Implement game state machine (Playing → Paused → Settings → Quit).

---

## Category 3: Procedural Generation Disconnects

### Gap 3.1: 83% of Generators Never Called at Runtime

**Severity: HIGH** · **Impact: Most procedural content exists only in tests**

Of 18 generator packages/adapters, only 3 are called at runtime:
- ✅ `city.Generate()` — Called in `cmd/server/main.go:60`
- ✅ `adapters.GenerateFactions()` — Called in `cmd/server/server_init.go:22`
- ✅ `adapters.GenerateAndSpawnNPCs()` — Called in `cmd/server/main.go:92`

Never called at runtime (15 generators):
- `pkg/procgen/dungeon/` — BSP dungeon generator (tested, never used)
- `adapters.BuildingAdapter` — Building generation
- `adapters.DialogAdapter` — NPC dialog trees
- `adapters.ItemAdapter` — Item/weapon generation
- `adapters.FurnitureAdapter` — Interior furniture
- `adapters.NarrativeAdapter` — Story arc generation
- `adapters.QuestAdapter` — Quest template generation
- `adapters.RecipeAdapter` — Crafting recipes
- `adapters.TerrainAdapter` — Terrain features
- `adapters.VehicleAdapter` — Vehicle generation
- `adapters.PuzzleAdapter` — Dungeon puzzles
- `adapters.MagicAdapter` — Spell generation (stub only)
- `adapters.SkillsAdapter` — Skill tree generation (stub only)
- `adapters.EnvironmentAdapter` — Environmental details (stub only)
- `adapters.BiomeAdapter` — Biome generation (stub only)

**Location:** `cmd/server/main.go`, `cmd/server/server_init.go`
**Fix:** Call generators during server world initialization to populate buildings, items, quests, vehicles, etc.

---

### Gap 3.2: Dungeon Generator Orphaned

**Severity: MODERATE** · **Impact: No dungeons in game despite BSP generator being complete**

`pkg/procgen/dungeon/` generates fully connected dungeon layouts with rooms, doors, traps, puzzle areas, boss arenas, and genre-specific tile aesthetics. It has comprehensive tests proving 100 generated dungeons have 0 unreachable rooms. But it is never imported by any runtime code.

**Location:** `pkg/procgen/dungeon/`
**Fix:** Call dungeon generator for instanced content (quest dungeons, building basements) during world initialization.

---

## Category 4: Rendering and Visual Gaps

### Gap 4.1: Rendering Subpackages Disconnected from Client

**Severity: HIGH** · **Impact: 6 of 7 rendering packages unused**

The client imports only `pkg/rendering/raycast/`. These rendering packages are fully implemented but never used at runtime:

| Package | LOC | Status |
|---------|-----|--------|
| `pkg/rendering/sprite/` | ~1,200 | Procedural sprite generation — indirectly referenced by raycast billboard.go but no sprites exist |
| `pkg/rendering/texture/` | ~600 | Procedural textures — never applied to rendered walls |
| `pkg/rendering/lighting/` | ~500 | Point/directional/spot lighting with day/night — never integrated |
| `pkg/rendering/particles/` | ~900 | 11 particle types (rain, snow, sparks, magic) — never rendered |
| `pkg/rendering/postprocess/` | ~400 | Genre-specific post-processing (scanlines, vignette, bloom) — never applied |
| `pkg/rendering/subtitles/` | ~500 | Subtitle rendering with accessibility options — never displayed |

**Location:** `cmd/client/main.go` (only imports raycast)
**Fix:** Integrate texture generator into raycast wall colors, enable particles for weather, apply post-processing in Draw().

---

### Gap 4.2: No Entity Rendering

**Severity: HIGH** · **Impact: NPCs, items, vehicles invisible even if present**

The raycast renderer supports billboard rendering (`pkg/rendering/raycast/billboard.go`) and the sprite generator (`pkg/rendering/sprite/`) can create entity sprites, but:
- No entities have `Appearance` components on the client
- No system populates the renderer's billboard list
- Entity positions are not synced from server to client

**Location:** `pkg/rendering/raycast/billboard.go`, `pkg/rendering/sprite/`
**Fix:** Create entity Appearance components, sync entity positions from server, feed billboard list to renderer each frame.

---

## Category 5: Networking and Client-Server Gaps

### Gap 5.1: No Game State Synchronization

**Severity: CRITICAL** · **Impact: Multiplayer is non-functional beyond TCP connection**

The server accepts TCP connections and runs a tick loop, but:
- No entity state is broadcast to clients
- No client input is processed by the server
- No chunk data is streamed to clients
- Client and server maintain completely independent ECS worlds

The protocol messages (`PlayerInput`, `WorldState`, `EntityUpdate`, `ChunkData`) are defined in `pkg/network/protocol.go` but never sent or received in the game loop.

**Location:** `cmd/server/main.go:116-138` (server loop), `cmd/client/main.go:42-48` (client loop)
**Fix:** Implement message send/receive in both server tick loop and client update loop.

---

### Gap 5.2: Federation Protocol Initialized But Unused in Game Loop

**Severity: LOW** · **Impact: Cross-server features don't affect gameplay**

Federation is initialized in `cmd/server/main.go:39-42` and cleanup runs in a goroutine, but:
- No player transfer messages are sent
- No economy gossip is exchanged
- No global events are broadcast
- The federation object is only used for cleanup

**Location:** `cmd/server/main.go:39-42`, `cmd/server/server_init.go:70-91`
**Fix:** Integrate federation messaging into the server tick loop.

---

## Category 6: Audio Integration Gaps

### Gap 6.1: Audio Subpackages Disconnected

**Severity: MODERATE** · **Impact: No ambient soundscapes or adaptive music**

The client creates an `audio.Engine` and `audio.Player` and plays a single ambient sine wave, but:
- `pkg/audio/ambient/` (soundscapes with biome awareness) — never imported
- `pkg/audio/music/` (adaptive music with genre motifs) — never imported
- `AudioSystem` runs but has no `AudioListener` entity, so `findListenerPosition()` always returns false
- No `AudioSource` entities exist on the client

**Location:** `cmd/client/main.go:191-215`, `pkg/engine/systems/audio.go`
**Fix:** Add AudioListener component to player entity. Import ambient/music packages. Create AudioSource entities for world sounds.

---

## Category 7: Package Integration Gaps

### Gap 7.1: Input Rebinding System Not Used

**Severity: MODERATE** · **Impact: Hardcoded keys; no rebinding**

`pkg/input/` provides a complete `Rebinder` with 40+ bindable actions and config loading, but the client uses hardcoded `ebiten.IsKeyPressed()` calls instead.

**Location:** `cmd/client/main.go:127-153`
**Fix:** Replace direct Ebiten key checks with `input.Rebinder.IsPressed()` calls.

---

### Gap 7.2: Dialog Package Not Integrated

**Severity: MODERATE** · **Impact: Cannot converse with NPCs**

`pkg/dialog/` provides genre-aware dialog with emotional states, topic memory, and branching conversations. It is fully tested but never imported by runtime code.

**Location:** `pkg/dialog/`
**Fix:** Create a dialog UI in the client and use the dialog system for NPC conversations.

---

### Gap 7.3: Companion Package Not Integrated

**Severity: LOW** · **Impact: No companion NPCs**

`pkg/companion/` provides companion AI with personality, combat roles, and action memory. Never imported at runtime.

**Location:** `pkg/companion/`
**Fix:** Integrate companion spawning during character creation or quest progression.

---

### Gap 7.4: Housing/PvP/Persist Packages Not Integrated

**Severity: LOW** · **Impact: No player housing, PvP zones, or data persistence**

Three `pkg/world/` subpackages with full implementations and test coverage are unused:
- `pkg/world/housing/` — Player houses, furniture, guild territories
- `pkg/world/pvp/` — PvP zone management with flags and loot mechanics
- `pkg/world/persist/` — World state serialization and persistence

**Location:** `pkg/world/housing/`, `pkg/world/pvp/`, `pkg/world/persist/`
**Fix:** Import into server initialization for housing registration, PvP zone creation, and state persistence.

---

## Category 8: Previously Resolved Code Quality Gaps

The following gaps from the original GAPS.md (2026-03-31) are resolved and retained for reference:

| # | Gap | Status |
|---|-----|--------|
| 1 | FEATURES.md Summary Table Outdated | ✅ RESOLVED — 200/200 (100.0%) |
| 2 | Entry Point Test Coverage (0% → 30%) | ✅ RESOLVED — Client 100%, Server 83.8% |
| 3 | Performance Target Unverifiable (60 FPS) | ✅ RESOLVED — 18+ benchmarks exist |
| 4 | Build-Tag Test Visibility | ✅ RESOLVED — noebiten coverage reported |
| 5 | High Complexity Functions (5 → 0) | ✅ RESOLVED — All refactored |
| 6 | Code Duplication (2.90% → <2.0%) | ✅ RESOLVED — Now 1.89% |
| 7 | Companion Package Coverage (78.8% → 85%) | ✅ RESOLVED — Now 87.1% |
| 8 | Large File Cohesion | ✅ RESOLVED — All files split |
| 9 | Naming Convention Violations (35 → 23) | ✅ RESOLVED |

---

## Summary: Gap Priority Matrix

| Priority | Gap | Player Impact | Effort | Category |
|----------|-----|---------------|--------|----------|
| **P0** | 1.1 — 44 unregistered systems | All gameplay non-functional | Medium | Integration |
| **P0** | 5.1 — No state sync | Multiplayer broken | High | Networking |
| **P0** | 1.2 — Client has no game logic | Client is empty shell | High | Integration |
| **P1** | 2.1 — No collision detection | Walk through walls | Low | Gameplay |
| **P1** | 2.2 — No HUD/UI | Zero visual feedback | Medium | UI/UX |
| **P1** | 4.1 — Rendering packages unused | No lighting, particles, post-processing | Medium | Rendering |
| **P1** | 3.1 — 83% generators unused | World feels empty | Medium | Content |
| **P2** | 2.3 — No interaction system | Can't interact with anything | Medium | Gameplay |
| **P2** | 2.5 — No character creation | Hardcoded starting state | Medium | Onboarding |
| **P2** | 4.2 — No entity rendering | NPCs invisible | Medium | Rendering |
| **P2** | 6.1 — Audio subpackages unused | No soundscapes or music | Medium | Audio |
| **P2** | 2.4 — No save/load | Progress lost on exit | Medium | Persistence |
| **P3** | 7.1 — Input rebinding unused | Hardcoded keys | Low | Input |
| **P3** | 7.2 — Dialog not integrated | No NPC conversations | Medium | Dialog |
| **P3** | 2.6 — No menu system | No pause/settings/quit | Medium | UI/UX |
| **P3** | 1.3 — WeatherSystem wrong init | Client weather uses defaults | Low | Integration |
| **P4** | 3.2 — Dungeon generator orphaned | No dungeons despite generator | Low | Content |
| **P4** | 7.3 — Companion not integrated | No companion NPCs | Low | Features |
| **P4** | 7.4 — Housing/PvP/Persist unused | No housing or PvP zones | Low | Features |
| **P4** | 5.2 — Federation unused in loop | Cross-server non-functional | Low | Networking |

**Total: 20 active gaps** (9 code quality gaps resolved, 20 playability gaps identified)

---

*Generated 2026-03-31 from comprehensive codebase audit. See AUDIT.md for full system status matrix.*
