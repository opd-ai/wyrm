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

**Resolution Checklist:**

- [x] Register NPC behavior systems (6 systems: NPCPathfindingSystem, NPCNeedsSystem, NPCOccupationSystem, EmotionalStateSystem, NPCMemorySystem, GossipSystem)
- [x] Register faction depth systems (4 systems: FactionRankSystem, FactionCoupSystem, FactionExclusiveContentSystem, DynamicFactionWarSystem)
- [x] Register crime depth systems (5 systems: GuardPursuitSystem, BriberySystem, CrimeEvidenceSystem, PardonSystem, CriminalFactionQuestSystem)
- [x] Register economy depth systems (8 systems: EconomicEventSystem, MarketManipulationSystem, TradeRouteSystem, InvestmentSystem, PlayerShopSystem, CityBuildingSystem, CityEventSystem, TradingSystem)
- [x] Register combat depth systems (10 systems: MagicSystem, ProjectileSystem, StealthSystem, DistractionSystem, HidingSpotSystem, VehiclePhysicsSystem, VehicleCombatSystem, FlyingVehicleSystem, NavalVehicleSystem, MountSystem)
- [x] Register skills/crafting systems (6 systems: SkillProgressionSystem, SkillBookSystem, SkillSynergySystem, ActionUnlockSystem, NPCTrainingSystem, CraftingSystem)
- [x] Register dialog/social systems (4 systems: DialogConsequenceSystem, MultiNPCConversationSystem, PartySystem, VehicleCustomizationSystem)
- [x] Register environment systems (2 systems: IndoorOutdoorSystem, HazardSystem)
- [x] Verify `grep -c 'RegisterSystem' cmd/server/main.go` shows 57 registrations (actually shows 58)

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

**Resolution Checklist:**

- [x] Decide approach: client-side systems for offline mode OR entity sync protocol (offline mode with full system registration chosen)
- [x] If offline mode: register necessary client-side game logic systems (Combat, Quest, NPC, etc.)
- [x] If entity sync: implement server → client EntityUpdate message pipeline (N/A — offline mode chosen)
- [x] Verify client game loop executes meaningful gameplay logic beyond rendering (56 systems registered)

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

**Resolution Checklist:**

- [x] Replace `&systems.WeatherSystem{}` with `systems.NewWeatherSystem(cfg.Genre, 300.0)` in `cmd/client/main.go`
- [x] Verify client builds: `go build ./cmd/client`
- [x] Verify weather cycling uses correct genre and transition duration

---

## Category 2: Missing Core Gameplay Features

### Gap 2.1: No Player Collision Detection

**Severity: HIGH** · **Impact: Player walks through all walls and terrain**

Player movement in `cmd/client/main.go:123-153` directly modifies Position X/Y without checking the world map for walls. The raycaster renders walls but the movement code does not test against them.

**Location:** `cmd/client/main.go:123-153` (`processMovementInput`, `processStrafeInput`)
**Fix:** Before applying movement, check `worldMap[newY][newX]` for wall cells. Reject movement into cells where `heightToWallType() > 0`.

**Resolution Checklist:**

- [x] Store `worldMap` reference on the Game struct
- [x] Add bounds-checking for worldMap indices in `processMovementInput()`
- [x] Add bounds-checking for worldMap indices in `processStrafeInput()`
- [x] Reject movement into wall cells (value > 0)
- [x] Implement player radius (0.3 units) for wall sliding
- [x] Test that player cannot walk through rendered walls

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

**Resolution Checklist:**

- [x] Implement health bar (red bar, bottom-left, reads Health component)
- [x] Implement mana bar (blue bar, below health bar, reads Mana component)
- [x] Implement minimap (top-right, terrain from worldMap)
- [x] Implement compass (cardinal direction from player angle)
- [x] Implement inventory screen (I key toggle)
- [x] Implement quest log screen (J key toggle)
- [x] Implement dialog interface for NPC conversations
- [x] Implement pause/settings menu (Escape key)
- [x] Implement genre selection screen
- [x] Implement character creation screen

---

### Gap 2.3: No Interaction System (E Key)

**Severity: HIGH** · **Impact: Cannot interact with any world object**

The `pkg/input/` package defines 40+ bindable actions including `interact` (default: E key), but the key binding system is never imported or used by the client. The client only handles WASD/QE movement directly via Ebiten key checks.

**Location:** `cmd/client/main.go:108-153` (input handling)
**Fix:** Import `pkg/input`, use `Rebinder` for all key checks, implement interaction raycasting to find the entity the player is looking at.

**Resolution Checklist:**

- [x] Import `pkg/input` in client
- [x] Create `input.Rebinder` with config-loaded key bindings
- [x] Replace all `ebiten.IsKeyPressed()` calls with `rebinder.IsPressed()` equivalents (note: arrow keys retained as accessibility fallbacks)
- [x] Implement interaction ray cast from player position in look direction
- [x] Implement E key interaction with nearest entity (NPC, item, workbench, door)
- [x] Display interaction prompt on screen ("Press E to ...")

---

### Gap 2.4: No Save/Load Integration

**Severity: HIGH** · **Impact: All progress lost when game exits**

`pkg/world/persist/` provides a `PersistenceManager` with save/load functionality and test coverage, but it is never imported by either `cmd/client/` or `cmd/server/`.

**Location:** `pkg/world/persist/`
**Fix:** Call `PersistenceManager.Save()` on server shutdown and `PersistenceManager.Load()` on startup.

**Resolution Checklist:**

- [x] Import `pkg/world/persist/` in server
- [x] Call `PersistenceManager.Load()` on server startup (check for existing save file)
- [x] Call `PersistenceManager.Save()` on server shutdown (SIGINT/SIGTERM handler)
- [x] Implement periodic auto-save at configurable interval
- [x] Verify save includes entities, components, chunk modifications, and quest states
- [x] Test save/load round-trip preserves world state

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

**Resolution Checklist:**

- [x] Implement genre selection screen (5 genres with preview descriptions)
- [x] Implement character name input
- [x] Implement starting skill point allocation
- [x] Implement starting equipment choice
- [x] Add Mana component to player entity
- [x] Add Skills component with initialized maps
- [x] Add Inventory component with capacity
- [x] Add Faction, Reputation, Stealth, CombatState, AudioListener, and Weapon components
- [x] Create player entity with user-chosen attributes

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

**Resolution Checklist:**

- [x] Implement game state machine (Playing, Paused, Settings, CharacterCreation, Quit)
- [x] Implement Escape key → pause state toggle and menu overlay
- [x] Implement Resume menu option
- [x] Implement Settings screen (keybinds, audio, graphics)
- [x] Implement settings persistence via config file update
- [x] Implement Quit option with confirmation dialog

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

**Resolution Checklist:**

- [x] Call `adapters.BuildingAdapter` to generate building interiors for city districts
- [x] Call `adapters.DialogAdapter` to generate NPC dialog trees
- [x] Call `adapters.ItemAdapter` to populate building inventories with items
- [x] Call `adapters.FurnitureAdapter` to furnish building interiors
- [x] Call `adapters.NarrativeAdapter` to generate story arcs
- [x] Call `adapters.QuestAdapter` to generate quest templates for NPCs
- [x] Call `adapters.RecipeAdapter` to generate crafting recipes
- [x] Call `adapters.TerrainAdapter` to generate terrain features
- [x] Call `adapters.VehicleAdapter` to spawn vehicles in districts
- [x] Call `adapters.PuzzleAdapter` to generate dungeon puzzles
- [x] Implement `adapters.MagicAdapter` beyond stub
- [x] Implement `adapters.SkillsAdapter` beyond stub
- [x] Implement `adapters.EnvironmentAdapter` beyond stub
- [x] Verify all generators produce output consumed by game systems

---

### Gap 3.2: Dungeon Generator Orphaned

**Severity: MODERATE** · **Impact: No dungeons in game despite BSP generator being complete**

`pkg/procgen/dungeon/` generates fully connected dungeon layouts with rooms, doors, traps, puzzle areas, boss arenas, and genre-specific tile aesthetics. It has comprehensive tests proving 100 generated dungeons have 0 unreachable rooms. But it is never imported by any runtime code.

**Location:** `pkg/procgen/dungeon/`
**Fix:** Call dungeon generator for instanced content (quest dungeons, building basements) during world initialization.

**Resolution Checklist:**

- [x] Import `pkg/procgen/dungeon/` in server initialization code
- [x] Call `dungeon.Generate()` for quest-related instanced dungeon content
- [x] Wire generated dungeon layouts into quest objective locations
- [x] Verify dungeon rooms are reachable and correctly connected

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

**Resolution Checklist:**

- [x] Import and integrate `pkg/rendering/sprite/` for NPC/entity billboard rendering
- [x] Import and integrate `pkg/rendering/texture/` for procedural wall/floor textures
- [x] Import and integrate `pkg/rendering/lighting/` with time-of-day cycle
- [x] Import and integrate `pkg/rendering/particles/` for weather-driven particles
- [x] Import and integrate `pkg/rendering/postprocess/` for genre-specific effects
- [x] Import and integrate `pkg/rendering/subtitles/` for dialog text rendering
- [x] Verify all rendering subpackages produce visible output in the client

---

### Gap 4.2: No Entity Rendering

**Severity: HIGH** · **Impact: NPCs, items, vehicles invisible even if present**

The raycast renderer supports billboard rendering (`pkg/rendering/raycast/billboard.go`) and the sprite generator (`pkg/rendering/sprite/`) can create entity sprites, but:
- No entities have `Appearance` components on the client
- No system populates the renderer's billboard list
- Entity positions are not synced from server to client

**Location:** `pkg/rendering/raycast/billboard.go`, `pkg/rendering/sprite/`
**Fix:** Create entity Appearance components, sync entity positions from server, feed billboard list to renderer each frame.

**Resolution Checklist:**

- [x] Create `Appearance` components for all entity types (NPCs, items, vehicles)
- [x] Integrate `pkg/rendering/sprite/Generator` to produce entity sprites
- [x] Sync entity positions from server to client (or generate locally for offline)
- [x] Feed billboard list (position + sprite) to raycast renderer each frame
- [x] Verify entities are visible in the first-person view

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

**Resolution Checklist:**

- [ ] Server: broadcast `EntityUpdate` messages to connected clients each tick
- [ ] Server: stream `ChunkData` messages when client enters new chunk
- [ ] Server: receive and process `PlayerInput` messages from clients
- [ ] Client: receive and decode `WorldState` and `EntityUpdate` messages
- [ ] Client: apply server state to local ECS world entities
- [ ] Client: send `PlayerInput` messages to server each frame
- [ ] Implement client-side prediction using `pkg/network/prediction.go`
- [ ] Verify two clients can connect and observe shared world state

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

**Resolution Checklist:**

- [ ] Integrate player transfer messaging into server tick loop
- [ ] Integrate economy gossip exchange into server tick loop
- [ ] Integrate global event broadcast into server tick loop
- [ ] Verify federation features function beyond just cleanup

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

**Resolution Checklist:**

- [x] Add `AudioListener` component to player entity (in `createPlayerEntity`)
- [x] Import `pkg/audio/ambient/` and generate biome-aware ambient soundscapes
- [x] Import `pkg/audio/music/` and generate adaptive genre-specific music
- [x] Create `AudioSource` entities for world environmental sounds
- [x] Replace single sine wave with ambient soundscape and music output
- [x] Verify `AudioSystem.findListenerPosition()` finds the player's AudioListener

---

## Category 7: Package Integration Gaps

### Gap 7.1: Input Rebinding System Not Used

**Severity: MODERATE** · **Impact: Hardcoded keys; no rebinding**

`pkg/input/` provides a complete `Rebinder` with 40+ bindable actions and config loading, but the client uses hardcoded `ebiten.IsKeyPressed()` calls instead.

**Location:** `cmd/client/main.go:127-153`
**Fix:** Replace direct Ebiten key checks with `input.Rebinder.IsPressed()` calls.

**Resolution Checklist:**

- [x] Import `pkg/input` in client
- [x] Create `input.Rebinder` with config-loaded key bindings
- [x] Replace all `ebiten.IsKeyPressed()` calls with `rebinder.IsPressed()` equivalents (note: arrow keys retained as accessibility fallbacks)
- [x] Verify all movement and action keys work through the rebinder

---

### Gap 7.2: Dialog Package Not Integrated

**Severity: MODERATE** · **Impact: Cannot converse with NPCs**

`pkg/dialog/` provides genre-aware dialog with emotional states, topic memory, and branching conversations. It is fully tested but never imported by runtime code.

**Location:** `pkg/dialog/`
**Fix:** Create a dialog UI in the client and use the dialog system for NPC conversations.

**Resolution Checklist:**

- [x] Import `pkg/dialog/` in client
- [x] Create dialog overlay UI (NPC name, emotional state, dialog text, response options)
- [x] Wire dialog choice selection into `DialogConsequenceSystem`
- [x] Integrate `pkg/rendering/subtitles/` for accessible text rendering
- [x] Verify branching conversation flow with multiple responses

---

### Gap 7.3: Companion Package Not Integrated

**Severity: LOW** · **Impact: No companion NPCs**

`pkg/companion/` provides companion AI with personality, combat roles, and action memory. Never imported at runtime.

**Location:** `pkg/companion/`
**Fix:** Integrate companion spawning during character creation or quest progression.

**Resolution Checklist:**

- [x] Import `pkg/companion/` in server initialization
- [x] Spawn companion NPCs during character creation or quest progression
- [x] Wire companion AI (follow/fight/wait commands)
- [ ] Verify companion follows player and participates in combat

---

### Gap 7.4: Housing/PvP/Persist Packages Not Integrated

**Severity: LOW** · **Impact: No player housing, PvP zones, or data persistence**

Three `pkg/world/` subpackages with full implementations and test coverage are unused:
- `pkg/world/housing/` — Player houses, furniture, guild territories
- `pkg/world/pvp/` — PvP zone management with flags and loot mechanics
- `pkg/world/persist/` — World state serialization and persistence

**Location:** `pkg/world/housing/`, `pkg/world/pvp/`, `pkg/world/persist/`
**Fix:** Import into server initialization for housing registration, PvP zone creation, and state persistence.

**Resolution Checklist:**

- [x] Import `pkg/world/housing/` and register player houses during server initialization
- [x] Import `pkg/world/pvp/` and create PvP zone entities during server initialization
- [x] Import `pkg/world/persist/` and integrate save/load on server startup/shutdown
- [x] Verify housing, PvP, and persistence features are functional at runtime

---

## Category 8: Terrain Generation Quality Gaps

### Gap 8.1: Limited Geometric Variety — Only 4 Terrain Types

**Severity: HIGH** · **Impact: Terrain feels repetitive; no valleys, water, vegetation, or caves**

The heightmap system (`pkg/world/chunk/manager.go`) classifies terrain into only 4 types (Flat, Hill, Cliff, Peak) using simple height thresholds. The single-valued heightmap fundamentally prevents caves, overhangs, and complex 3D geometry.

**Missing terrain features:**

| Feature | Impact on Visual Quality | Complexity |
|---------|-------------------------|------------|
| Valleys/depressions (height < 0.2) | High — adds concavity | Low |
| Water bodies (lakes, rivers) | High — breaks monotony | Medium |
| Rock formations | Medium — adds detail objects | Medium |
| Vegetation placement | High — fills empty terrain | Medium |
| Roads between POIs | High — guides exploration | Medium |
| Caves/overhangs | Medium — adds depth | High (requires 3D data) |

**Location:** `pkg/world/chunk/manager.go` (lines 142-176, terrain classification)
**Fix:** Add valley detection (height < 0.2 threshold), water plane at configurable elevation, and vegetation/rock entity spawning based on biome + height.

**Resolution Checklist:**

- [x] Add `TerrainValley` type for heights < 0.2 in terrain classification
- [x] Add `TerrainWater` type for heights below a configurable water plane elevation
- [x] Implement vegetation entity spawning based on biome type and height (DetailSpawn system with Tree, Bush, Grass, Flower types)
- [x] Implement rock/boulder entity spawning on cliff and mountain terrain (DetailSpawn system with Rock, Boulder, Debris types)
- [ ] Implement road/path generation connecting city districts and POIs
- [ ] Verify new terrain types render distinctly in the raycaster

---

### Gap 8.2: No Biome Blending at Chunk Boundaries

**Severity: HIGH** · **Impact: Abrupt visual transitions at chunk edges break immersion**

Each chunk independently selects its biome from genre-weighted distribution (`pkg/procgen/adapters/terrain.go`). Adjacent chunks may have completely different biomes (e.g., Forest→Desert) with zero transition.

**Location:** `pkg/procgen/adapters/terrain.go` (biome selection), `pkg/world/chunk/manager.go` (chunk generation)
**Fix:** Sample adjacent chunk biomes within a configurable border width (e.g., 32 cells) and interpolate terrain parameters (heightmap amplitude, texture palette, vegetation density).

**Resolution Checklist:**

- [ ] Implement biome query for neighboring chunks during generation
- [ ] Add interpolation zone (32-cell border) that blends between adjacent biomes
- [ ] Blend heightmap parameters (amplitude, frequency) in transition zones
- [ ] Blend texture/color palettes in transition zones
- [ ] Verify smooth visual transitions at all chunk boundaries

---

### Gap 8.3: LOD System Defined But Not Integrated

**Severity: MODERATE** · **Impact: Distant terrain rendered at full resolution unnecessarily**

Four LOD levels are defined in `pkg/world/chunk/manager.go` (`LODFull`, `LODHalf`, `LODQuarter`, `LODEighth`) with a `ChunkLODCache` struct, but no rendering code selects LOD levels based on distance.

**Location:** `pkg/world/chunk/manager.go` (lines ~220-240)
**Fix:** Wire LOD selection into chunk streaming based on distance from player. Use lower LOD levels for chunks farther from the player.

**Resolution Checklist:**

- [ ] Implement distance-based LOD level selection in chunk streaming
- [ ] Feed LOD-appropriate chunk data to raycaster for distant terrain
- [ ] Verify visual quality at LOD boundaries
- [ ] Benchmark memory savings from LOD usage

---

### Gap 8.4: Value Noise Instead of Gradient Noise

**Severity: MODERATE** · **Impact: Terrain has blocky quality compared to Perlin/simplex noise**

The noise generator (`pkg/procgen/noise/generator.go`, 56 lines) uses hash-based value noise with smoothstep interpolation. Value noise interpolates between random values at grid points, producing less natural-looking terrain than gradient noise (Perlin/simplex).

**Location:** `pkg/procgen/noise/generator.go`
**Fix:** Implement gradient noise (Perlin or simplex) as an alternative/replacement noise function.

**Resolution Checklist:**

- [x] Implement 2D gradient (Perlin) noise function (GradientNoise2D in pkg/procgen/noise/generator.go)
- [x] Add noise function selection parameter to terrain generator (NoiseType parameter, used in chunk generation)
- [x] Verify deterministic output with gradient noise (tests in pkg/procgen/noise/generator_test.go)
- [ ] Benchmark gradient noise vs current value noise performance
- [ ] Compare terrain visual quality between noise algorithms

---

### Gap 8.5: Synchronous Chunk Generation Causes Frame Stutter

**Severity: MODERATE** · **Impact: Frame drops when crossing chunk boundaries for first time**

`GetChunk()` in `pkg/world/chunk/manager.go` generates chunks synchronously on the calling thread. First access triggers full heightmap + elevation + terrain type generation, blocking the game loop.

**Location:** `pkg/world/chunk/manager.go` (lines 291-312, `GetChunk`)
**Fix:** Implement async chunk generation with a worker goroutine. Return a placeholder/flat chunk while generation completes in background.

**Resolution Checklist:**

- [ ] Implement background chunk generation goroutine with work queue
- [ ] Return placeholder chunk for not-yet-generated chunks
- [ ] Swap placeholder with real chunk when generation completes
- [ ] Verify no frame stutter when crossing chunk boundaries
- [ ] Add benchmark for chunk generation latency

---

## Category 9: Ebitengine Performance & Rendering Bottlenecks

### Gap 9.1: Per-Pixel `screen.Set()` Rendering — Fundamental Performance Bottleneck

**Severity: CRITICAL** · **Impact: Prevents 60 FPS at any resolution; ~78M GPU sync points/sec**

All rendering (walls, floors, ceilings, sprites, UI, particles) uses `ebiten.Image.Set()` per-pixel calls instead of Ebitengine's batch rendering APIs. At 1280×720, the raycaster makes ~1.3M `Set()` calls per frame. Each call triggers GPU pipeline synchronization.

**36 `screen.Set()` call sites across 5 files:**

| File | Set() calls | Hot path? |
|------|-------------|-----------|
| `pkg/rendering/raycast/draw.go` | 5 | ✅ Every frame |
| `cmd/client/main.go` | 14 | ✅ Every frame |
| `cmd/client/quest_ui.go` | 10 | When UI open |
| `cmd/client/inventory_ui.go` | 6 | When UI open |
| `cmd/client/dialog_ui.go` | 1 | When dialog open |

**Location:** See AUDIT.md §4.1 for full analysis
**Fix:** Render to a `[]byte` software framebuffer, then upload once per frame via `ebiten.Image.WritePixels()`. This reduces GPU sync from ~1.3M to 1 call per frame.

**Resolution Checklist:**

- [ ] Create persistent `[]byte` software framebuffer (width×height×4) in Renderer
- [ ] Replace all `screen.Set(x, y, color)` in `draw.go` with framebuffer writes
- [ ] Replace all `screen.Set()` in `main.go` combat/particle/minimap code with framebuffer writes
- [ ] Upload framebuffer to screen via single `screen.WritePixels()` call in `Draw()`
- [ ] Replace UI `screen.Set()` calls with framebuffer writes or `DrawImage()` overlays
- [ ] Benchmark before/after: target ≥10× frame time improvement
- [ ] Verify visual output is identical

---

### Gap 9.2: Per-Frame Buffer Allocations — ~40 MB/Frame GC Pressure

**Severity: CRITICAL** · **Impact: GC stalls cause frame drops; 432 MB–1.08 GB/sec allocation rate**

Every frame allocates fresh buffers that are immediately discarded:

| Source | File | Allocation | Size (1280×720) |
|--------|------|-----------|-----------------|
| Particle buffer | `cmd/client/main.go:1103` | `make([]byte, w×h×4)` | 3.6 MB |
| Post-process input | `cmd/client/main.go:1074` | `image.NewRGBA()` | 3.6 MB |
| Post-process effects | `pkg/rendering/postprocess/effects.go` × 11 sites | `image.NewRGBA()` | 3.6 MB each |
| Z-Buffer | `pkg/rendering/raycast/draw.go:209` | `make([]float64, width)` | 10 KB |

**Total per frame:** 7.2 MB baseline, up to 18 MB with 3 post-processing effects active.
**At 60 FPS:** 432 MB/sec–1.08 GB/sec allocation pressure with zero `sync.Pool` usage anywhere.

**Location:** See AUDIT.md §8.2 for full breakdown
**Fix:** Pre-allocate all buffers once. Reuse across frames. Clear instead of reallocate.

**Resolution Checklist:**

- [ ] Pre-allocate persistent particle pixel buffer in particle renderer struct
- [ ] Pre-allocate persistent `image.RGBA` buffers for post-processing pipeline
- [ ] Move Z-buffer to Renderer struct field (allocate once in constructor)
- [ ] Move sprite sort slice to Renderer struct field
- [ ] Implement `sync.Pool` for any remaining dynamic buffer needs
- [ ] Benchmark allocation rate before/after: target ≥80% reduction
- [ ] Verify no visual regression

---

### Gap 9.3: Zero Batch Rendering API Usage

**Severity: HIGH** · **Impact: All GPU rendering capabilities unused**

The codebase uses virtually none of Ebitengine's batch rendering APIs:

| API | Current Usage | Optimal Usage |
|-----|---------------|---------------|
| `DrawImage()` | 1 (menu overlay) | UI overlays, sprite rendering |
| `DrawTriangles()` | 0 | Textured geometry |
| `WritePixels()` | 0 | Framebuffer upload (replaces all Set()) |
| `Fill()` | 0 | Background clearing |
| `ColorM` | 0 | Hardware color transforms |
| `GeoM` | 0 | Hardware geometry transforms |

**Location:** See AUDIT.md §4.2, §4.4
**Fix:** Use `WritePixels()` for raycaster output. Use `DrawImage()` with `ColorM`/`GeoM` for UI overlays and post-processing effects that can be expressed as color transforms (vignette, color grade, screen tint).

**Resolution Checklist:**

- [ ] Replace per-pixel raycaster output with `WritePixels()` framebuffer upload
- [ ] Use `DrawImage()` with `ColorM` for vignette, color grading, and tint effects
- [ ] Use `DrawImage()` with `GeoM` for UI element positioning
- [ ] Use `Fill()` for screen clearing
- [ ] Benchmark GPU utilization improvement

---

### Gap 9.4: Post-Processing Double Copy Loop

**Severity: HIGH** · **Impact: 2× full-screen pixel traversals per frame for format conversion**

Post-processing requires converting between `ebiten.Image` and `image.RGBA` (`cmd/client/main.go` lines 1074-1089):
1. Copy screen → `image.RGBA` (921,600 reads via `screen.At()`)
2. Apply N effects (921,600 ops each)
3. Copy `image.RGBA` → screen (921,600 writes via `screen.Set()`)

**Location:** `cmd/client/main.go` (lines 1074-1089), `pkg/rendering/postprocess/effects.go`
**Fix:** Operate post-processing directly on the software framebuffer before uploading to GPU. Eliminate the screen→RGBA→screen copy loop entirely.

**Resolution Checklist:**

- [ ] Apply post-processing to software framebuffer `[]byte` before `WritePixels()` upload
- [ ] Eliminate `screen.At()` → `image.RGBA` copy loop
- [ ] Eliminate `image.RGBA` → `screen.Set()` copy loop
- [ ] Verify post-processing visual effects are preserved
- [ ] Benchmark: target elimination of ~1.8M pixel operations per frame

---

### Gap 9.5: Sprite Alpha Blending via GPU Read-Modify-Write

**Severity: HIGH** · **Impact: Slowest possible per-pixel pattern for sprite rendering**

Sprite alpha blending (`pkg/rendering/raycast/draw.go` lines 120-129) uses `screen.At()` + CPU blend + `screen.Set()` per pixel — a GPU read + CPU compute + GPU write per pixel.

**Location:** `pkg/rendering/raycast/draw.go` (lines 120-129)
**Fix:** Perform alpha blending on the software framebuffer where it's a simple memory read-modify-write operation.

**Resolution Checklist:**

- [ ] Blend sprites directly into software framebuffer `[]byte`
- [ ] Eliminate `screen.At()` reads for alpha blending
- [ ] Verify sprite visual quality is preserved
- [ ] Benchmark sprite rendering improvement

---

### Gap 9.6: Particle System Full-Screen Copy Loops

**Severity: MODERATE** · **Impact: 1.8M extra pixel operations per frame for particle compositing**

Particle rendering (`cmd/client/main.go` lines 1099-1133) copies the entire screen to a `[]byte` buffer, renders particles, then copies back. This adds two full-screen traversals on top of particle rendering.

**Location:** `cmd/client/main.go` (lines 1099-1133)
**Fix:** Render particles directly into the shared software framebuffer before `WritePixels()` upload.

**Resolution Checklist:**

- [ ] Integrate particle rendering into main software framebuffer
- [ ] Eliminate screen → pixel buffer copy loop
- [ ] Eliminate pixel buffer → screen copy loop
- [ ] Verify particle visual effects preserved
- [ ] Benchmark: target elimination of 1.8M pixel operations per frame

---

### Gap 9.7: No Runtime Profiling Infrastructure

**Severity: LOW** · **Impact: Cannot diagnose performance issues in production builds**

No `net/http/pprof` import, no `runtime.MemStats` monitoring, no frame time tracking or debug overlay beyond a single `ebitenutil.DebugPrint` line.

**Location:** `cmd/client/main.go` (Draw method)
**Fix:** Add optional pprof HTTP endpoint (behind config flag) and frame time histogram to debug overlay.

**Resolution Checklist:**

- [ ] Add `debug.profiling` config option (default: false)
- [ ] When enabled, start `net/http/pprof` endpoint on configurable port
- [ ] Add frame time tracking (Update + Draw durations) to debug overlay
- [ ] Add memory stats (HeapAlloc, NumGC, GCPauseNs) to debug overlay

---

## Category 10: Previously Resolved Code Quality Gaps

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

### Performance & Rendering Gaps (NEW — from AUDIT.md terrain/Ebitengine analysis)

| Priority | Gap | Player Impact | Effort | Category |
|----------|-----|---------------|--------|----------|
| **P0** | 9.1 — Per-pixel Set() rendering | Cannot reach 60 FPS; ~78M GPU syncs/sec | High | Performance |
| **P0** | 9.2 — Per-frame buffer allocations | GC stalls cause frame drops; ~1 GB/sec alloc | Medium | Memory |
| **P1** | 9.3 — Zero batch rendering API usage | All GPU capabilities unused | High | Performance |
| **P1** | 9.4 — Post-process double copy loop | 1.8M extra pixel ops per frame | Medium | Performance |
| **P1** | 9.5 — Sprite read-modify-write | Slowest sprite rendering pattern | Medium | Performance |
| **P1** | 8.1 — Only 4 terrain types | Repetitive terrain; no water/vegetation | Medium | Terrain |
| **P1** | 8.2 — No biome blending | Abrupt chunk boundary transitions | Medium | Terrain |
| **P2** | 9.6 — Particle full-screen copy | 1.8M extra pixel ops per frame | Low | Performance |
| **P2** | 8.3 — LOD system not integrated | Distant terrain at full resolution | Low | Terrain |
| **P2** | 8.4 — Value noise only | Blocky terrain quality | Medium | Terrain |
| **P2** | 8.5 — Synchronous chunk generation | Frame stutter on chunk crossing | Medium | Terrain |
| **P3** | 9.7 — No profiling infrastructure | Cannot diagnose production perf | Low | Tooling |

### Playability & Integration Gaps (existing)

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

**Total: 32 active gaps** (9 code quality gaps resolved, 20 playability gaps + 12 terrain/performance gaps identified)

---

*Updated 2026-04-01. Terrain quality and Ebitengine performance gaps added from AUDIT.md analysis. See AUDIT.md for full technical assessment and PLAN.md for implementation roadmap.*
