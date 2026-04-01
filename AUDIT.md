# Wyrm System Integration Audit — 2026-03-31

This document provides a comprehensive technical assessment of every system, package, and integration point in the Wyrm codebase. It is designed to give any developer an instant, accurate picture of what works, what is wired, and what is idle.

---

## 1. System Status Matrix

### 1.1 Server-Side Systems Registered in `cmd/server/main.go`

| # | System | Constructor Used | Registered | Update Logic | Consumed By | Notes |
|---|--------|-----------------|------------|--------------|-------------|-------|
| 1 | WorldClockSystem | `NewWorldClockSystem(60.0)` | ✅ | ✅ Full | NPCScheduleSystem, CraftingSystem, CityBuildingSystem, CityEventSystem | Drives in-game time; WorldClock component advances hour/day |
| 2 | WorldChunkSystem | `NewWorldChunkSystem(cm, chunkSize)` | ✅ | ✅ Full | Client chunk streaming | Loads/unloads chunks around entities with Position |
| 3 | NPCScheduleSystem | `&systems.NPCScheduleSystem{}` | ✅ | ✅ Full | NPCPathfindingSystem | Reads WorldClock, sets CurrentActivity on Schedule components |
| 4 | FactionPoliticsSystem | `NewFactionPoliticsSystem(0.1)` | ✅ | ✅ Full | FactionCoupSystem, FactionRankSystem | Manages inter-faction relations and reputation decay |
| 5 | CrimeSystem | `NewCrimeSystem(60.0, 100.0)` | ✅ | ✅ Full | GuardPursuitSystem, EvidenceSystem, PardonSystem | Tracks crimes, wanted levels, bounties |
| 6 | EconomySystem | `NewEconomySystem(0.5, 0.1)` | ✅ | ✅ Full | EconomicEventSystem, TradeRouteSystem, MarketManipulationSystem | Supply/demand pricing on EconomyNode components |
| 7 | CombatSystem | `&systems.CombatSystem{}` | ✅ | ✅ Full | RenderSystem, AudioSystem (combat intensity) | Melee damage resolution, cooldowns, attack processing |
| 8 | VehicleSystem | `&systems.VehicleSystem{}` | ✅ | ✅ Full | VehiclePhysicsSystem | Basic vehicle movement |
| 9 | QuestSystem | `NewQuestSystem()` | ✅ | ✅ Full | DialogConsequenceSystem | Quest state machine, stage transitions, branching |
| 10 | WeatherSystem | `NewWeatherSystem(cfg.Genre, 300.0)` | ✅ | ✅ Full | IndoorOutdoorSystem, HazardSystem | Dynamic weather cycling with extreme events |

### 1.2 Client-Side Systems Registered in `cmd/client/main.go`

| # | System | Constructor Used | Registered | Update Logic | Consumed By | Notes |
|---|--------|-----------------|------------|--------------|-------------|-------|
| 1 | RenderSystem | `&systems.RenderSystem{PlayerEntity: player}` | ✅ | ⚠️ Minimal | Draw() in Game | Only reads player Position — no rendering logic in Update() |
| 2 | AudioSystem | `&systems.AudioSystem{Genre: cfg.Genre}` | ✅ | ✅ Full | Audio playback | Spatial audio, combat intensity, ambient selection |
| 3 | WeatherSystem | `&systems.WeatherSystem{}` | ✅ | ✅ Full | RenderSystem (indirectly) | Weather cycling; **not** initialized with `NewWeatherSystem()` |

### 1.3 Systems Defined But NOT Registered Anywhere (44 Systems)

These systems exist in `pkg/engine/systems/` with full implementations but are never instantiated or registered in either `cmd/client/main.go` or `cmd/server/main.go`.

| # | System | File | Has Constructor | Update Logic | Purpose | Impact of Not Being Registered |
|---|--------|------|-----------------|--------------|---------|-------------------------------|
| 1 | ActionUnlockSystem | skill_unlock.go | ✅ | ✅ Full | Skill-based action unlocks | Players can't unlock abilities at skill levels |
| 2 | BriberySystem | crime.go | ✅ | ✅ Full | Bribe guards to reduce bounty | No guard bribery mechanics |
| 3 | CityBuildingSystem | city_buildings.go | ✅ | ✅ Full | Building operation hours, POIs | Buildings don't open/close, no shop schedules |
| 4 | CityEventSystem | city_events.go | ✅ | ✅ Full | Dynamic city events | No festivals, riots, plagues, or market crashes |
| 5 | CraftingSystem | crafting.go | ✅ | ✅ Full | Workbench crafting | Players can't craft items |
| 6 | CrimeEvidenceSystem | evidence.go | ✅ | ✅ Full | Crime evidence collection | No forensic investigation, evidence decay |
| 7 | CriminalFactionQuestSystem | criminal_quest.go | ✅ | ✅ Full | Criminal faction quests | No thieves guild / criminal questlines |
| 8 | DialogConsequenceSystem | dialog_consequence.go | ✅ | ✅ Full | Dialog choice consequences | Dialog choices have no lasting effects |
| 9 | DistractionSystem | stealth_distraction.go | ✅ | ✅ Full | Stealth distractions | Can't throw objects to distract guards |
| 10 | DynamicFactionWarSystem | faction.go | ✅ | ✅ Full | Faction wars | No inter-faction warfare events |
| 11 | EconomicEventSystem | economic_event.go | ✅ | ✅ Full | Dynamic economic events | No market booms, crashes, or shortages |
| 12 | EmotionalStateSystem | emotional_state.go | ✅ | ✅ Full | NPC emotions | NPCs have no emotional responses |
| 13 | FactionCoupSystem | faction_coup.go | ✅ | ✅ Full | Faction leadership coups | No internal faction power struggles |
| 14 | FactionExclusiveContentSystem | faction_exclusive.go | ✅ | ✅ Full | Faction-exclusive content | Rank-gated content not accessible |
| 15 | FactionRankSystem | faction_rank.go | ✅ | ✅ Full | Faction rank progression | Players can't advance in factions |
| 16 | FlyingVehicleSystem | vehicle_flying.go | ✅ | ✅ Full | Flying vehicles / airships | No flying vehicle operation |
| 17 | GossipSystem | gossip.go | ✅ | ✅ Full | NPC gossip propagation | NPCs don't share information |
| 18 | GuardPursuitSystem | crime.go | ✅ | ✅ Full | Guard pursuit AI | Guards don't pursue criminals |
| 19 | HazardSystem | hazard.go | ✅ | ✅ Full | Environmental hazards | Hazard zones do no damage |
| 20 | HidingSpotSystem | stealth_hiding.go | ✅ | ✅ Full | Hiding spots | Can't hide from NPCs |
| 21 | IndoorOutdoorSystem | weather_indoor.go | ✅ | ✅ Full | Indoor/outdoor detection | Weather affects players indoors |
| 22 | InvestmentSystem | investment.go | ✅ | ✅ Full | Player investments | Can't invest gold for returns |
| 23 | MagicSystem | magic_combat.go | ✅ | ✅ Full | Mana regen, spells, effects | No magic casting |
| 24 | MarketManipulationSystem | market_manipulation.go | ✅ | ✅ Full | Market manipulation | No market manipulation schemes |
| 25 | MountSystem | vehicle_mount.go | ✅ | ✅ Full | Mount riding | Can't ride mounts/horses |
| 26 | MultiNPCConversationSystem | multi_npc_conversation.go | ✅ | ✅ Full | Multi-NPC conversations | Can't participate in group conversations |
| 27 | NavalVehicleSystem | vehicle_naval.go | ✅ | ✅ Full | Naval vehicles / ships | No ship sailing |
| 28 | NPCMemorySystem | npc_memory.go | ✅ | ✅ Full | NPC player memory | NPCs don't remember player interactions |
| 29 | NPCNeedsSystem | npc_needs.go | ✅ | ✅ Full | NPC basic needs | NPCs have no hunger/energy/social needs |
| 30 | NPCOccupationSystem | npc_occupation.go | ✅ | ✅ Full | NPC work behavior | NPCs don't perform their occupations |
| 31 | NPCPathfindingSystem | npc_pathfinding.go | ✅ | ✅ Full | NPC pathfinding | NPCs don't move to activity locations |
| 32 | NPCTrainingSystem | skill_progression.go | ✅ | ✅ Full | NPC skill training | Can't train skills with NPCs |
| 33 | PardonSystem | pardon.go | ✅ | ✅ Full | Crime pardons | Can't get pardoned for crimes |
| 34 | PartySystem | party.go | ✅ | ✅ Full | Player party management | Can't form parties |
| 35 | PlayerShopSystem | economy.go | ✅ | ✅ Full | Player-owned shops | Can't own/operate shops |
| 36 | ProjectileSystem | ranged_combat.go | ✅ | ✅ Full | Projectile physics | No arrows, bullets, or ranged attacks |
| 37 | SkillBookSystem | skill_book.go | ✅ | ✅ Full | Skill books for training | Can't read skill books |
| 38 | SkillProgressionSystem | skill_progression.go | ✅ | ✅ Full | Skill XP and leveling | No skill progression |
| 39 | SkillSynergySystem | skill_book.go | ✅ | ✅ Full | Skill synergy bonuses | No cross-skill synergies |
| 40 | StealthSystem | stealth.go | ✅ | ✅ Full | Stealth detection | Can't sneak |
| 41 | TradeRouteSystem | trade_route.go | ✅ | ✅ Full | Trade routes and caravans | No trade route economy |
| 42 | TradingSystem | trading.go | ✅ | ✅ Full | Player-to-player trading | Can't trade with other players |
| 43 | VehicleCombatSystem | vehicle_combat.go | ✅ | ✅ Full | Vehicle-mounted combat | No vehicle weapons |
| 44 | VehicleCustomizationSystem | vehicle_customization.go | ✅ | ✅ Full | Vehicle upgrades | Can't customize vehicles |
| 45 | VehiclePhysicsSystem | vehicle_physics.go | ✅ | ✅ Full | Realistic vehicle physics | Vehicles use basic movement only |

**Registration Summary:**
- **Server registered:** 10 / 57 systems (17.5%)
- **Client registered:** 3 / 57 systems (5.3%)
- **Unregistered:** 44 systems with full implementations (77.2%)

### 1.5 System Registration Completion Checklist

Track progress toward registering all 44 unregistered systems:

**NPC Behavior Systems:**
- [x] Register NPCPathfindingSystem in `cmd/server/main.go`
- [x] Register NPCNeedsSystem in `cmd/server/main.go`
- [x] Register NPCOccupationSystem in `cmd/server/main.go`
- [x] Register EmotionalStateSystem in `cmd/server/main.go`
- [x] Register NPCMemorySystem in `cmd/server/main.go`
- [x] Register GossipSystem in `cmd/server/main.go`

**Faction Depth Systems:**
- [x] Register FactionRankSystem in `cmd/server/main.go`
- [x] Register FactionCoupSystem in `cmd/server/main.go`
- [x] Register FactionExclusiveContentSystem in `cmd/server/main.go`
- [x] Register DynamicFactionWarSystem in `cmd/server/main.go`

**Crime Depth Systems:**
- [x] Register GuardPursuitSystem in `cmd/server/main.go`
- [x] Register BriberySystem in `cmd/server/main.go`
- [x] Register CrimeEvidenceSystem in `cmd/server/main.go`
- [x] Register PardonSystem in `cmd/server/main.go`
- [x] Register CriminalFactionQuestSystem in `cmd/server/main.go`

**Economy Depth Systems:**
- [x] Register EconomicEventSystem in `cmd/server/main.go`
- [x] Register MarketManipulationSystem in `cmd/server/main.go`
- [x] Register TradeRouteSystem in `cmd/server/main.go`
- [x] Register InvestmentSystem in `cmd/server/main.go`
- [x] Register PlayerShopSystem in `cmd/server/main.go`
- [x] Register CityBuildingSystem in `cmd/server/main.go`
- [x] Register CityEventSystem in `cmd/server/main.go`
- [x] Register TradingSystem in `cmd/server/main.go`

**Combat Depth Systems:**
- [x] Register MagicSystem in `cmd/server/main.go`
- [x] Register ProjectileSystem in `cmd/server/main.go`
- [x] Register StealthSystem in `cmd/server/main.go`
- [x] Register DistractionSystem in `cmd/server/main.go`
- [x] Register HidingSpotSystem in `cmd/server/main.go`
- [x] Register VehiclePhysicsSystem in `cmd/server/main.go`
- [x] Register VehicleCombatSystem in `cmd/server/main.go`
- [x] Register FlyingVehicleSystem in `cmd/server/main.go`
- [x] Register NavalVehicleSystem in `cmd/server/main.go`
- [x] Register MountSystem in `cmd/server/main.go`

**Skills/Crafting Systems:**
- [x] Register SkillProgressionSystem in `cmd/server/main.go`
- [x] Register SkillBookSystem in `cmd/server/main.go`
- [x] Register SkillSynergySystem in `cmd/server/main.go`
- [x] Register ActionUnlockSystem in `cmd/server/main.go`
- [x] Register NPCTrainingSystem in `cmd/server/main.go`
- [x] Register CraftingSystem in `cmd/server/main.go`

**Dialog/Social Systems:**
- [x] Register DialogConsequenceSystem in `cmd/server/main.go`
- [x] Register MultiNPCConversationSystem in `cmd/server/main.go`
- [x] Register PartySystem in `cmd/server/main.go`
- [x] Register VehicleCustomizationSystem in `cmd/server/main.go`

**Environment Systems:**
- [x] Register IndoorOutdoorSystem in `cmd/server/main.go`
- [x] Register HazardSystem in `cmd/server/main.go`

### 1.4 Support Types (Not Systems)

| Type | File | Purpose |
|------|------|---------|
| FactionArcManager | quest_faction_arcs.go | Faction quest arc management (manager, no Update method) |
| DynamicQuestGenerator | quest.go | Quest generation helper (used by QuestSystem) |
| RadiantQuestBoard | quest.go | Radiant quest helper (used by QuestSystem) |
| SkillRegistry | skill_progression.go | Skill definition registry (used by SkillProgressionSystem) |

---

## 2. Code Architecture Analysis

### 2.1 ECS Implementation

| Aspect | Status | Details |
|--------|--------|---------|
| Entity type | ✅ Complete | `uint64` ID, monotonic allocation via `World.nextID` |
| Component interface | ✅ Complete | `Type() string` method, stored in `map[Entity]map[string]Component` |
| System interface | ✅ Complete | `Update(w *World, dt float64)` |
| World struct | ✅ Complete | Entity CRUD, component add/get/remove, system registration, sorted entity queries |
| Entity queries | ✅ Complete | `Entities(types ...string)` with deterministic sort order |
| System execution | ✅ Complete | Sequential in registration order via `World.Update(dt)` |

**Component Count:** 59 distinct component types with `Type()` method + 18 supporting structs

**Key Component Categories:**
- **Core:** Position, Health, Mana (3)
- **NPC Behavior:** Schedule, NPCPathfinding, NPCMemory, NPCRelationships, NPCNeeds, NPCOccupation, EmotionalState, SocialStatus, Awareness, GossipNetwork (10)
- **Combat:** Weapon, CombatState, Projectile, Spell, Spellbook, SpellEffect, Stealth (7)
- **Economy/Crafting:** EconomyNode, Inventory, Material, ResourceNode, Workbench, CraftingState, Tool, RecipeKnowledge, ShopInventory (9)
- **Faction/Crime:** Faction, FactionMembership, FactionTerritory, Reputation, Crime, Witness, Guard (7)
- **World:** WorldClock, Building, Interior, POIMarker, GovernmentBuilding (5)
- **Vehicle:** Vehicle, VehiclePhysics, VehicleState (3)
- **Dialog:** DialogState, DialogMemory (2)
- **Audio/Visual:** AudioListener, AudioSource, AudioState, Appearance (4)
- **Hazards:** EnvironmentalHazard, HazardResistance, HazardEffect, HazardZone, WeatherHazard, TrapMechanism (6)
- **Events/Quests:** Quest, CityEvent, Skills (3)

### 2.2 Procedural Generation Pipeline

| Generator | Package | Called at Runtime | Called By | Seed-Based | Genre-Aware |
|-----------|---------|-------------------|-----------|------------|-------------|
| City Generator | `pkg/procgen/city/` | ✅ `city.Generate()` | `cmd/server/main.go:60` | ✅ | ✅ |
| Entity/NPC Generator | `pkg/procgen/adapters/entity.go` | ✅ `GenerateAndSpawnNPCs()` | `cmd/server/main.go:92` | ✅ | ✅ |
| Faction Generator | `pkg/procgen/adapters/faction.go` | ✅ `GenerateFactions()` | `cmd/server/server_init.go:22` | ✅ | ✅ |
| Dungeon Generator | `pkg/procgen/dungeon/` | ❌ Never called | — | ✅ | ✅ |
| Noise Functions | `pkg/procgen/noise/` | ✅ (by chunk gen) | `pkg/world/chunk/` | ✅ | N/A |
| Building Adapter | `pkg/procgen/adapters/building.go` | ❌ Never called | — | ✅ | ✅ |
| Dialog Adapter | `pkg/procgen/adapters/dialog.go` | ❌ Never called | — | ✅ | ✅ |
| Item Adapter | `pkg/procgen/adapters/item.go` | ❌ Never called | — | ✅ | ✅ |
| Furniture Adapter | `pkg/procgen/adapters/furniture.go` | ❌ Never called | — | ✅ | ✅ |
| Narrative Adapter | `pkg/procgen/adapters/narrative.go` | ❌ Never called | — | ✅ | ✅ |
| Quest Adapter | `pkg/procgen/adapters/quest.go` | ❌ Never called | — | ✅ | ✅ |
| Recipe Adapter | `pkg/procgen/adapters/recipe.go` | ❌ Never called | — | ✅ | ✅ |
| Terrain Adapter | `pkg/procgen/adapters/terrain.go` | ❌ Never called | — | ✅ | ✅ |
| Vehicle Adapter | `pkg/procgen/adapters/vehicle.go` | ❌ Never called | — | ✅ | ✅ |
| Puzzle Adapter | `pkg/procgen/adapters/puzzle.go` | ❌ Never called | — | ✅ | ✅ |
| Magic Adapter | `pkg/procgen/adapters/magic.go` | ❌ Stub only | — | — | — |
| Skills Adapter | `pkg/procgen/adapters/skills.go` | ❌ Stub only | — | — | — |
| Environment Adapter | `pkg/procgen/adapters/environment.go` | ❌ Stub only | — | — | — |

**Generators called at runtime:** 3 out of 18 (16.7%)
**Generators with stub-only implementations:** 3

### 2.3 Generator Integration Completion Checklist

Track progress toward calling all generators at runtime:

- [x] City Generator — called in `cmd/server/main.go:60`
- [x] Entity/NPC Generator — called in `cmd/server/main.go:92`
- [x] Faction Generator — called in `cmd/server/server_init.go:22`
- [x] Noise Functions — called by `pkg/world/chunk/`
- [x] Dungeon Generator — called in `cmd/server/server_init.go:initializeDungeons()`
- [x] Building Adapter — called in `cmd/server/main.go:initializeCity()` for district buildings
- [x] Dialog Adapter — called in `cmd/server/main.go:initializeCity()` to generate NPC dialog trees
- [x] Item Adapter — called in `cmd/server/main.go:initializeCity()` to populate shop inventories
- [x] Furniture Adapter — called in `cmd/server/main.go:initializeCity()` to furnish building interiors
- [ ] Narrative Adapter — call to generate story arcs
- [ ] Quest Adapter — call to generate quest templates for NPCs
- [ ] Recipe Adapter — call to generate crafting recipes
- [ ] Terrain Adapter — call to generate terrain features
- [ ] Vehicle Adapter — call to spawn vehicles in districts
- [ ] Puzzle Adapter — call to generate dungeon puzzles
- [ ] Magic Adapter — implement beyond stub and call at runtime
- [ ] Skills Adapter — implement beyond stub and call at runtime
- [ ] Environment Adapter — implement beyond stub and call at runtime

### 2.3 Networking Infrastructure

| Aspect | Status | Location | Details |
|--------|--------|----------|---------|
| Server (TCP listener) | ✅ Complete | `pkg/network/server.go` | Accepts connections, interface-based (`net.Listener`) |
| Client (TCP dialer) | ✅ Complete | `pkg/network/server.go` | Connects to server, graceful disconnect |
| Binary Protocol | ✅ Complete | `pkg/network/protocol.go` | Message types: PlayerInput, WorldState, EntityUpdate, ChunkData, Ping/Pong |
| Client-Side Prediction | ✅ Complete | `pkg/network/prediction.go` | Input prediction with server reconciliation |
| Lag Compensation | ✅ Complete | `pkg/network/lagcomp.go` | Server-side position history rewind |
| Federation | ✅ Complete | `pkg/network/federation/` | Cross-server player transfer, economy sync, gossip protocol |
| **Chunk Streaming Protocol** | ❌ Not implemented | — | Protocol messages exist but no chunk streaming in server loop |
| **Entity Sync** | ❌ Not implemented | — | No entity state broadcast from server to clients |
| **Input Processing** | ❌ Not implemented | — | Server accepts connections but doesn't process PlayerInput messages |

### 2.4 Package Usage Matrix

| Package | Imported by Client | Imported by Server | Imported by Systems | Imported by Tests Only |
|---------|-------------------|-------------------|--------------------|-----------------------|
| `pkg/engine/ecs` | ✅ | ✅ | ✅ | — |
| `pkg/engine/components` | ✅ | ✅ | ✅ | — |
| `pkg/engine/systems` | ✅ | ✅ | — | — |
| `pkg/rendering/raycast` | ✅ | ❌ | ❌ | — |
| `pkg/rendering/sprite` | ❌ | ❌ | ❌ | ✅ (indirectly via raycast) |
| `pkg/rendering/texture` | ❌ | ❌ | ❌ | ✅ |
| `pkg/rendering/lighting` | ❌ | ❌ | ❌ | ✅ |
| `pkg/rendering/particles` | ❌ | ❌ | ❌ | ✅ |
| `pkg/rendering/postprocess` | ❌ | ❌ | ❌ | ✅ |
| `pkg/rendering/subtitles` | ❌ | ❌ | ❌ | ✅ |
| `pkg/audio` | ✅ | ❌ | ❌ | — |
| `pkg/audio/ambient` | ❌ | ❌ | ❌ | ✅ |
| `pkg/audio/music` | ❌ | ❌ | ❌ | ✅ |
| `pkg/network` | ✅ | ✅ | ❌ | — |
| `pkg/network/federation` | ❌ | ✅ | ❌ | — |
| `pkg/world/chunk` | ✅ | ✅ | ✅ | — |
| `pkg/world/housing` | ❌ | ❌ | ❌ | ✅ |
| `pkg/world/persist` | ❌ | ❌ | ❌ | ✅ |
| `pkg/world/pvp` | ❌ | ❌ | ❌ | ✅ |
| `pkg/procgen/city` | ❌ | ✅ | ❌ | — |
| `pkg/procgen/dungeon` | ❌ | ❌ | ❌ | ✅ |
| `pkg/procgen/adapters` | ❌ | ✅ | ❌ | — |
| `pkg/procgen/noise` | ❌ | ❌ | ❌ | (by chunk) |
| `pkg/dialog` | ❌ | ❌ | ❌ | ✅ |
| `pkg/companion` | ❌ | ❌ | ❌ | ✅ |
| `pkg/input` | ❌ | ❌ | ❌ | ✅ |
| `pkg/util` | ❌ | ❌ | ✅ | — |
| `config` | ✅ | ✅ | ❌ | — |

**Packages unused at runtime:** 14 out of 28 (50%)

### 2.5 Package Integration Completion Checklist

Track progress toward integrating all packages at runtime:

- [x] `pkg/engine/ecs` — used by client and server
- [x] `pkg/engine/components` — used by client and server
- [x] `pkg/engine/systems` — used by client and server
- [x] `pkg/rendering/raycast` — used by client
- [x] `pkg/audio` — used by client
- [x] `pkg/network` — used by client and server
- [x] `pkg/network/federation` — used by server
- [x] `pkg/world/chunk` — used by client and server
- [x] `pkg/procgen/city` — used by server
- [x] `pkg/procgen/adapters` — used by server
- [x] `pkg/procgen/noise` — used by chunk system
- [x] `pkg/util` — used by systems
- [x] `config` — used by client and server
- [ ] `pkg/rendering/sprite` — integrate for NPC/entity billboard rendering
- [ ] `pkg/rendering/texture` — integrate for procedural wall/floor textures
- [ ] `pkg/rendering/lighting` — integrate for time-of-day lighting
- [ ] `pkg/rendering/particles` — integrate for weather particle effects
- [ ] `pkg/rendering/postprocess` — integrate for genre-specific post-processing
- [ ] `pkg/rendering/subtitles` — integrate for dialog subtitle rendering
- [ ] `pkg/audio/ambient` — integrate for biome-aware ambient soundscapes
- [ ] `pkg/audio/music` — integrate for adaptive genre-specific music
- [ ] `pkg/world/housing` — integrate for player housing and guild territory
- [ ] `pkg/world/persist` — integrate for world state persistence
- [ ] `pkg/world/pvp` — integrate for PvP zone management
- [ ] `pkg/procgen/dungeon` — integrate for instanced dungeon content
- [ ] `pkg/dialog` — integrate for NPC conversation UI
- [ ] `pkg/companion` — integrate for companion NPC spawning
- [ ] `pkg/input` — integrate for key rebinding

---

## 3. Build and Runtime Analysis

### 3.1 Compilation Status

| Target | Command | Result | Notes |
|--------|---------|--------|-------|
| Client (with Ebiten) | `go build ./cmd/client` | ✅ Success | Requires libasound2-dev, libgl1-mesa-dev, xorg-dev |
| Server (with Ebiten) | `go build ./cmd/server` | ✅ Success | Same system deps |
| Client (noebiten) | `go build -tags=noebiten ./cmd/client` | ❌ Fails | `main` undeclared — no stub `main()` for noebiten tag. Tests pass because `main_test.go` has `//go:build noebiten` and tests only testable functions, not `main()` |
| Server (noebiten) | `go build -tags=noebiten ./cmd/server` | ❌ Fails | `main` undeclared — same reason. Tests in `main_test.go` cover exported helper functions via noebiten tag |

### 3.2 Test Execution

| Package | Tests | Result | Coverage Notes |
|---------|-------|--------|----------------|
| `cmd/client` | ✅ Pass | ✅ | 100% (with `-tags=noebiten`) |
| `cmd/server` | ✅ Pass | ✅ | 83.8% (with `-tags=noebiten`) |
| `config` | ✅ Pass | ✅ | Viper config loading |
| `pkg/audio` | ✅ Pass | ✅ | Engine synthesis tests |
| `pkg/audio/ambient` | ✅ Pass | ✅ | Ambient soundscapes |
| `pkg/audio/music` | ✅ Pass | ✅ | Music generation |
| `pkg/companion` | ✅ Pass | ✅ | 87.1% coverage |
| `pkg/dialog` | ✅ Pass | ✅ | Genre-aware dialog |
| `pkg/engine/components` | ✅ Pass | ✅ | Type() method validation |
| `pkg/engine/ecs` | ✅ Pass | ✅ | World CRUD, queries |
| `pkg/engine/systems` | ✅ Pass | ✅ | 37 test files |
| `pkg/input` | ✅ Pass | ✅ | Key rebinding |
| `pkg/network` | ✅ Pass | ✅ | Protocol, prediction |
| `pkg/network/federation` | ✅ Pass | ✅ | Cross-server federation |
| `pkg/procgen/adapters` | ✅ Pass | ✅ | 89.2% coverage |
| `pkg/procgen/city` | ✅ Pass | ✅ | City generation |
| `pkg/procgen/dungeon` | ✅ Pass | ✅ | BSP dungeon generation |
| `pkg/procgen/noise` | ✅ Pass | ✅ | Noise functions |
| `pkg/rendering/*` (7 pkgs) | ✅ Pass | ✅ | Raycast, sprite, texture, lighting, particles, postprocess, subtitles |
| `pkg/world/chunk` | ✅ Pass | ✅ | Chunk generation |
| `pkg/world/housing` | ✅ Pass | ✅ | Housing/guild territory |
| `pkg/world/persist` | ✅ Pass | ✅ | Persistence |
| `pkg/world/pvp` | ✅ Pass | ✅ | PvP zones |

**All 30 packages pass.** 79 test files, 45,230 lines of test code.

### 3.3 Codebase Metrics

| Metric | Value |
|--------|-------|
| Total Go files | 221 |
| Implementation LOC | 53,193 |
| Test LOC | 45,230 |
| Test files | 79 |
| Component types | 59 |
| System types | 57 |
| Package count | 28 |
| Dependencies (direct) | 3 (ebiten, viper, venture) |

### 3.4 Runtime Behavior (Expected)

Based on code analysis, when the client launches:

1. **Config loads** from `config.yaml` with Viper defaults ✅
2. **ECS World created** with 3 registered systems (Render, Audio, Weather) ✅
3. **Player entity** spawned at (8.5, 8.5) with Position + Health ✅
4. **Chunk manager** initialized, 3×3 chunk window loaded ✅
5. **Raycaster** initialized at configured resolution ✅
6. **Audio engine** generates ambient sine wave tone ✅
7. **Server connection** attempted, falls back to offline mode ✅
8. **Game loop** runs: input → world.Update(dt) → chunk map update → draw ✅

**What the player sees:**
- First-person raycasted view of procedurally generated terrain
- Height-based wall coloring (3 wall types from chunk heightmaps)
- WASD movement, left/right turning, Q/E strafing
- Debug text overlay: "Wyrm [fantasy] offline"
- Ambient audio tone (genre-based frequency)

**What the player cannot do:**
- No collision detection (walks through walls)
- No NPC interaction (NPCs only exist on server)
- No HUD (no health/mana bars, minimap, inventory)
- No menu system (no pause, settings, save/load)
- No combat, crafting, dialog, quests, or any gameplay mechanics
- No genre selection UI
- No character creation

---

## 4. Configuration System

### 4.1 Config Structure

```yaml
window:    { width: 1280, height: 720, title: "Wyrm" }
server:    { address: "localhost:7777", protocol: "tcp", tick_rate: 20 }
world:     { seed: 0, chunk_size: 512 }
federation: { enabled: false, node_id: "", peers: [], gossip_interval: 5 }
genre:     "fantasy"
```

### 4.2 Extended Config (in code, not in default YAML)

The `Config` struct in `config/load.go` also supports:
- **Accessibility:** colorblind modes (5 types), high contrast, large text, reduced motion, screen reader, subtitles
- **Difficulty:** level (easy/normal/hard/custom), damage/xp/price multipliers, perma-death, friendly fire, auto-aim
- **KeyBindings:** 40+ bindable actions (WASD, interact, inventory, combat, etc.)

These are defined with defaults but not exposed in `config.yaml`.

---

## 5. Dependency Analysis

### 5.1 External Dependencies

| Dependency | Version | Used By | Purpose |
|------------|---------|---------|---------|
| `ebiten/v2` | v2.9.3 | Client, rendering | Game engine, window, input, audio context |
| `viper` | v1.19.0 | Config | YAML config + env vars |
| `venture` | v0.0.0-20260321 | Server (adapters) | V-Series procedural generation library |

### 5.2 Cross-Package Dependencies (Internal)

```
cmd/client → config, ecs, components, systems, network, raycast, chunk, audio
cmd/server → config, ecs, components, systems, network, federation, adapters, city, chunk
systems → ecs, components, util, chunk (via interface)
raycast → sprite
chunk → noise
adapters → venture (external)
```

No circular dependencies detected.

---

## 6. Architecture Observations

### 6.1 Strengths

1. **Clean ECS separation** — Components are pure data, systems contain all logic
2. **Comprehensive system implementations** — 57 systems with full Update() logic (only RenderSystem is minimal)
3. **Extensive test coverage** — 79 test files covering all major subsystems
4. **Interface-based networking** — Uses `net.Listener`, `net.Conn` (not concrete types)
5. **Deterministic generation** — Seed-based RNG throughout, explicit `*rand.Rand` instances
6. **Genre parameterization** — 20+ systems accept genre parameter
7. **Thread safety** — Mutex-protected shared state in Party, Trading, Vehicle systems
8. **Builds cleanly** — No compiler warnings, all tests pass

### 6.2 Critical Architecture Gaps

1. **77% of systems never execute** — 44 systems with full logic but no registration
2. **Client has almost no systems** — Only 3 systems vs 57 available
3. **No client-server protocol usage** — Server accepts connections but doesn't send/receive game state
4. **50% of packages unused at runtime** — 14 packages exist only for tests
5. **No UI framework** — No menus, HUD, dialogs, or character creation
6. **No collision detection** — Player walks through terrain walls
7. **No save/load integration** — `pkg/world/persist` exists but is never called
8. **No game state synchronization** — Client and server maintain independent worlds

### 6.3 Architecture Gap Resolution Checklist

- [x] Register all 44 unregistered systems (see Section 1.5 checklist)
- [ ] Add client-side systems for single-player mode or implement entity sync protocol
- [ ] Implement client-server game state protocol (send/receive EntityUpdate, ChunkData, PlayerInput)
- [ ] Integrate all 14 unused packages at runtime (see Section 2.5 checklist)
- [ ] Implement HUD overlay system (health, mana, compass, minimap)
- [ ] Implement menu system (pause, settings, character creation, quit)
- [ ] Implement dialog UI for NPC conversations
- [ ] Add player collision detection against worldMap wall cells
- [ ] Integrate `pkg/world/persist/` for save/load on server startup/shutdown
- [ ] Implement game state synchronization between client and server ECS worlds
