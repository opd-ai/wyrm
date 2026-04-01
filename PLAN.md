# Wyrm Playability Implementation Plan — 2026-03-31

This document provides an actionable, prioritized implementation roadmap to take Wyrm from its current state (working renderer + movement shell) to a playable first-person RPG. Each phase builds on the previous one, with clear milestones and success criteria.

> **Current State:** Client renders procedurally-generated terrain via raycasting with WASD movement. Server generates cities, factions, and NPCs but does not communicate game state to clients. 77% of implemented systems (44 of 57) are never registered. 50% of packages (14 of 28) are unused at runtime.

---

## Phase 1: Critical System Integration (Week 1 — ~40 hours)

**Goal:** Make the game mechanically functional in single-player offline mode. Player can walk through a world with collision, see a HUD, and interact with the environment.

### 1A: Register All Server Systems (4 hours)

Register the 44 unregistered systems in `cmd/server/main.go`. Systems with cross-dependencies must be registered in the correct order.

**Files:** `cmd/server/main.go`

**Registration Order (dependency-aware):**

```
Foundation (pre-existing, already registered in codebase): WorldClockSystem,
    WorldChunkSystem, NPCScheduleSystem, FactionPoliticsSystem, CrimeSystem,
    EconomySystem, CombatSystem, VehicleSystem, QuestSystem, WeatherSystem

Phase 2 (NPC behavior): NPCPathfindingSystem, NPCNeedsSystem,
    NPCOccupationSystem, EmotionalStateSystem, NPCMemorySystem, GossipSystem

Phase 3 (faction depth): FactionRankSystem, FactionCoupSystem,
    FactionExclusiveContentSystem, DynamicFactionWarSystem

Phase 4 (crime depth): GuardPursuitSystem, BriberySystem,
    CrimeEvidenceSystem, PardonSystem, CriminalFactionQuestSystem

Phase 5 (economy depth): EconomicEventSystem, MarketManipulationSystem,
    TradeRouteSystem, InvestmentSystem, PlayerShopSystem, CityBuildingSystem,
    CityEventSystem, TradingSystem

Phase 6 (combat depth): MagicSystem, ProjectileSystem, StealthSystem,
    DistractionSystem, HidingSpotSystem, VehiclePhysicsSystem,
    VehicleCombatSystem, FlyingVehicleSystem, NavalVehicleSystem, MountSystem

Phase 7 (skills/crafting): SkillProgressionSystem, SkillBookSystem,
    SkillSynergySystem, ActionUnlockSystem, NPCTrainingSystem, CraftingSystem

Phase 8 (dialog/social): DialogConsequenceSystem,
    MultiNPCConversationSystem, PartySystem, VehicleCustomizationSystem

Phase 9 (environment): IndoorOutdoorSystem, HazardSystem
```

**Completion Checklist:**

- [x] Register NPC behavior systems (NPCPathfindingSystem, NPCNeedsSystem, NPCOccupationSystem, EmotionalStateSystem, NPCMemorySystem, GossipSystem)
- [x] Register faction depth systems (FactionRankSystem, FactionCoupSystem, FactionExclusiveContentSystem, DynamicFactionWarSystem)
- [x] Register crime depth systems (GuardPursuitSystem, BriberySystem, CrimeEvidenceSystem, PardonSystem, CriminalFactionQuestSystem)
- [x] Register economy depth systems (EconomicEventSystem, MarketManipulationSystem, TradeRouteSystem, InvestmentSystem, PlayerShopSystem, CityBuildingSystem, CityEventSystem, TradingSystem)
- [x] Register combat depth systems (MagicSystem, ProjectileSystem, StealthSystem, DistractionSystem, HidingSpotSystem, VehiclePhysicsSystem, VehicleCombatSystem, FlyingVehicleSystem, NavalVehicleSystem, MountSystem)
- [x] Register skills/crafting systems (SkillProgressionSystem, SkillBookSystem, SkillSynergySystem, ActionUnlockSystem, NPCTrainingSystem, CraftingSystem)
- [x] Register dialog/social systems (DialogConsequenceSystem, MultiNPCConversationSystem, PartySystem, VehicleCustomizationSystem)
- [x] Register environment systems (IndoorOutdoorSystem, HazardSystem)
- [x] Verify all 57 systems registered: `grep -c 'RegisterSystem' cmd/server/main.go`

**Validation:** `grep -c 'RegisterSystem' cmd/server/main.go` shows 57 registrations.

### 1B: Fix Client WeatherSystem Initialization (15 minutes)

**File:** `cmd/client/main.go:245`

**Change:**
```go
// Before:
world.RegisterSystem(&systems.WeatherSystem{})
// After (300.0 = seconds between weather transitions):
world.RegisterSystem(systems.NewWeatherSystem(cfg.Genre, 300.0))
```

**Completion Checklist:**

- [x] Replace `&systems.WeatherSystem{}` with `systems.NewWeatherSystem(cfg.Genre, 300.0)` in `cmd/client/main.go`
- [x] Verify client builds: `go build ./cmd/client`

### 1C: Add Player Collision Detection (4 hours)

**File:** `cmd/client/main.go`

**Approach:**
1. Store the `worldMap` grid on the Game struct (currently only passed to renderer)
2. In `processMovementInput()` and `processStrafeInput()`, compute candidate position
3. Bounds-check candidate coordinates against worldMap dimensions before indexing
4. Check `worldMap[candidateY][candidateX]` — reject movement if cell is a wall (value > 0)
5. Use player radius (0.3 units) for wall sliding instead of hard rejection

**Completion Checklist:**

- [x] Store `worldMap` reference on the Game struct
- [x] Add bounds-checking before worldMap indexing in `processMovementInput()`
- [x] Add bounds-checking before worldMap indexing in `processStrafeInput()`
- [x] Implement wall-cell rejection (value > 0) for candidate positions
- [x] Implement player radius (0.3 units) wall sliding
- [x] Test that player cannot walk through rendered walls

**Validation:** Player cannot walk through rendered walls.

### 1D: Add Complete Player Components (2 hours)

**File:** `cmd/client/main.go` (`createPlayerEntity`)

**Add components:**
```go
world.AddComponent(player, &components.Mana{Current: 50, Max: 50, RegenRate: 1.0})
world.AddComponent(player, &components.Skills{Levels: make(map[string]int), Experience: make(map[string]float64), SchoolBonuses: make(map[string]float64)})
world.AddComponent(player, &components.Inventory{Items: []string{}, Capacity: 30})
world.AddComponent(player, &components.Faction{ID: "player", Reputation: 0})
world.AddComponent(player, &components.Reputation{Standings: make(map[string]float64)})
world.AddComponent(player, &components.Stealth{Visibility: 1.0, BaseVisibility: 1.0, SneakVisibility: 0.3, DetectionRadius: 15.0})
world.AddComponent(player, &components.CombatState{})
world.AddComponent(player, &components.AudioListener{Volume: 1.0, Enabled: true})
world.AddComponent(player, &components.Weapon{Name: "Fists", Damage: 5, Range: 1.5, AttackSpeed: 1.0, WeaponType: "melee"})
```

**Completion Checklist:**

- [x] Add Mana component to player entity
- [x] Add Skills component with initialized maps
- [x] Add Inventory component with capacity
- [x] Add Faction component with player ID
- [x] Add Reputation component with standings map
- [x] Add Stealth component with detection parameters
- [x] Add CombatState component
- [x] Add AudioListener component
- [x] Add Weapon component (Fists default)
- [x] Verify client builds and player entity has all components

### 1E: Minimal HUD Implementation (8 hours)

**File:** `cmd/client/main.go` (`Draw` method)

**HUD Elements (priority order):**
1. **Health bar** — Red bar at bottom-left, reads Health component
2. **Mana bar** — Blue bar below health bar, reads Mana component
3. **Coordinates** — Current position X, Y, chunk coordinates
4. **Genre indicator** — Current genre and weather state
5. **Connection status** — Online/offline indicator
6. **Compass** — Cardinal direction based on player angle
7. **Minimap** — Small top-right map showing nearby terrain from worldMap

**Implementation:** Use `ebitenutil.DebugPrintAt()` for text elements. Use `screen.Set()` or small `ebiten.Image` draws for bars and minimap.

**Completion Checklist:**

- [x] Implement health bar overlay (red bar, bottom-left)
- [x] Implement mana bar overlay (blue bar, below health)
- [x] Display current position X, Y and chunk coordinates
- [x] Display genre indicator and weather state
- [x] Display connection status (online/offline)
- [x] Implement compass showing cardinal direction from player angle
- [x] Implement minimap (top-right, terrain from worldMap)
- [x] Verify HUD renders correctly at target resolution (1280×720)

### 1F: Integrate Input Rebinding (4 hours)

**Files:** `cmd/client/main.go`

**Changes:**
1. Import `pkg/input`
2. Create `input.Rebinder` with config-loaded bindings in `main()`
3. Replace all `ebiten.IsKeyPressed()` calls with `rebinder.IsPressed("action_name")`
4. Add Escape key → pause state toggle

**Completion Checklist:**

- [x] Import `pkg/input` in client
- [x] Create `input.Rebinder` with config-loaded key bindings
- [x] Replace all `ebiten.IsKeyPressed()` calls with `rebinder.IsPressed()` equivalents
- [x] Bind Escape key to pause state toggle
- [x] Verify all movement and action keys work through the rebinder

### 1G: Integrate Rendering Subpackages (8 hours)

**Files:** `cmd/client/main.go`, `pkg/rendering/raycast/`

**Steps:**
1. Import `pkg/rendering/texture/` — Generate wall/floor textures and pass to raycast renderer
2. Import `pkg/rendering/lighting/` — Create `LightManager`, add time-of-day cycle
3. Import `pkg/rendering/postprocess/` — Apply genre-specific post-processing effects after raycast draw
4. Import `pkg/rendering/particles/` — Add weather particles (rain, snow based on WeatherSystem state)

**Completion Checklist:**

- [x] Import and integrate `pkg/rendering/texture/` for procedural wall/floor textures
- [x] Import and integrate `pkg/rendering/lighting/` with `LightManager` and time-of-day cycle
- [x] Import and integrate `pkg/rendering/postprocess/` with genre-specific effects
- [x] Import and integrate `pkg/rendering/particles/` with weather-driven particles
- [x] Verify all rendering subpackages produce visible output in the client
- [x] Confirm 60 FPS target at 1280×720 is maintained

### 1H: Add AudioListener and Audio Integration (4 hours)

**Files:** `cmd/client/main.go`

**Steps:**
1. Add `AudioListener` component to player entity (done in 1D)
2. Add `AudioState` entity for tracking ambient/combat state
3. Import `pkg/audio/ambient/` and `pkg/audio/music/`
4. Replace single sine wave with proper ambient soundscape generation
5. Feed AudioSystem output into audio player

**Completion Checklist:**

- [x] Verify `AudioListener` component is on player entity (from 1D)
- [x] Create `AudioState` entity for ambient/combat state tracking
- [x] Import `pkg/audio/ambient/` and generate biome-aware soundscapes
- [x] Import `pkg/audio/music/` and generate adaptive genre-specific music
- [x] Replace single sine wave with soundscape and music output
- [x] Verify AudioSystem finds AudioListener and produces spatial audio

---

**Phase 1 Milestone: "Basic Playability"**
- ✅ Player can launch game, see a 3D world, and move around
- ✅ Collision detection prevents walking through walls
- ✅ HUD shows health, mana, position, and compass
- ✅ Keys are rebindable
- ✅ Genre-appropriate textures, lighting, and post-processing
- ✅ Ambient audio matches environment
- ✅ All 57 server systems execute each tick

---

## Phase 2: Core Gameplay Loop (Weeks 2–3 — ~60 hours)

**Goal:** Player can interact with the world, talk to NPCs, accept and complete quests, and engage in combat.

### 2A: Interaction System (12 hours)

**Files:** New file `cmd/client/interaction.go`

**Implementation:**
1. Cast a ray from player position in look direction
2. Find nearest entity within interaction range (2.0 units)
3. Highlight interactable entities (NPCs, items, workbenches)
4. On E key press:
   - NPC → Open dialog UI
   - Item → Pick up (add to inventory)
   - Workbench → Open crafting UI
   - Door → Open/close
5. Display interaction prompt ("Press E to talk to [NPC Name]")

**Completion Checklist:**

- [x] Implement interaction ray cast from player position in look direction
- [x] Implement nearest-entity search within 2.0 unit range
- [x] Highlight interactable entities (NPCs, items, workbenches, doors)
- [x] Implement NPC interaction (open dialog UI on E key)
- [x] Implement item pickup (add to inventory on E key)
- [x] Implement workbench interaction (open crafting UI on E key)
- [x] Implement door open/close on E key
- [x] Display interaction prompt text on screen

### 2B: Dialog UI (8 hours)

**Files:** New file `cmd/client/dialog_ui.go`

**Implementation:**
1. Import `pkg/dialog/`
2. Create dialog overlay screen when talking to NPC
3. Show NPC name, emotional state, dialog text
4. Display response options (numbered or arrow-selectable)
5. Wire dialog choices into `DialogConsequenceSystem`
6. Use `pkg/rendering/subtitles/` for accessible text rendering

**Completion Checklist:**

- [x] Import `pkg/dialog/` in client
- [x] Create dialog overlay screen (semi-transparent background)
- [x] Display NPC name and emotional state
- [x] Display dialog text with line wrapping
- [x] Display response options (numbered or arrow-selectable)
- [x] Wire dialog choice selection into `DialogConsequenceSystem`
- [x] Integrate `pkg/rendering/subtitles/` for accessible text rendering
- [x] Test branching conversation flow with multiple responses

### 2C: Basic Combat Mechanics (10 hours)

**Files:** `cmd/client/main.go`, new file `cmd/client/combat.go`

**Implementation:**
1. Left mouse click → melee attack (swing weapon)
2. Right mouse click → block/defend
3. Check CombatState cooldown before allowing attack
4. Calculate damage using Weapon.Damage, target Health
5. Visual feedback: screen shake on hit, red flash on damage taken
6. Death handling: respawn at safe location

**Completion Checklist:**

- [x] Implement left-click melee attack (swing weapon)
- [x] Implement right-click block/defend
- [x] Add CombatState cooldown check before allowing attacks
- [x] Implement damage calculation using Weapon.Damage vs target Health
- [x] Add screen shake visual feedback on hitting an enemy
- [x] Add red flash visual feedback on taking damage
- [x] Implement death detection and respawn at safe location
- [x] Verify combat works against NPC entities

### 2D: NPC Entity Rendering (10 hours)

**Files:** `cmd/client/main.go`

**Steps:**
1. Create `Appearance` components for NPC entity types
2. Use `pkg/rendering/sprite/Generator` to create NPC sprites
3. Feed NPC positions and sprites into raycast billboard system
4. Animate sprites based on NPC state (idle, walking, working)

**Dependency:** Requires entity state sync from server (Gap 5.1) OR local NPC generation for offline mode.

**Completion Checklist:**

- [x] Create `Appearance` components for NPC entity types
- [x] Integrate `pkg/rendering/sprite/Generator` to produce NPC sprites
- [x] Feed NPC positions and sprites into raycast billboard renderer
- [x] Implement sprite animation based on NPC state (idle, walking, working)
- [x] Verify NPCs are visible in the first-person view
- [x] Ensure NPC rendering works in offline mode (local NPC generation)

### 2E: Inventory UI (8 hours)

**Files:** New file `cmd/client/inventory_ui.go`

**Implementation:**
1. Press I → Toggle inventory overlay
2. Grid display of held items
3. Click to select, use, or drop items
4. Equipment slots display (weapon, armor, accessories)
5. Weight/capacity indicator

**Completion Checklist:**

- [x] Implement I key toggle for inventory overlay
- [x] Implement grid display of held items
- [x] Implement click-to-select item interaction
- [x] Implement item use action
- [x] Implement item drop action
- [x] Display equipment slots (weapon, armor, accessories)
- [x] Display weight/capacity indicator
- [x] Verify inventory updates reflect in Inventory component

### 2F: Quest System Integration (6 hours)

**Files:** `cmd/client/main.go`, new file `cmd/client/quest_ui.go`

**Implementation:**
1. Press J → Toggle quest log overlay
2. Display active quests with descriptions and objectives
3. Quest tracker on-screen (active quest objective)
4. Quest completion notifications
5. Wire quest adapter for procedural quest generation in server init

**Completion Checklist:**

- [x] Implement J key toggle for quest log overlay
- [x] Display active quests with descriptions and objectives
- [x] Implement on-screen quest tracker showing active objective
- [x] Implement quest completion notification display
- [x] Wire `adapters.QuestAdapter` into server world initialization
- [x] Verify quest acceptance, tracking, and completion end-to-end

### 2G: Content Generator Integration (6 hours)

**Files:** `cmd/server/main.go`, `cmd/server/server_init.go`

**Call remaining generators during world initialization:**
1. `adapters.BuildingAdapter` → Generate building interiors for city districts
2. `adapters.ItemAdapter` → Populate building inventories with items
3. `adapters.QuestAdapter` → Generate quest templates for NPCs
4. `adapters.RecipeAdapter` → Generate crafting recipes
5. `adapters.VehicleAdapter` → Spawn vehicles in districts
6. `dungeon.Generate()` → Create dungeon instances for quest objectives

**Completion Checklist:**

- [x] Call `adapters.BuildingAdapter` to generate building interiors for city districts
- [x] Call `adapters.ItemAdapter` to populate building inventories with items
- [x] Call `adapters.QuestAdapter` to generate quest templates for NPCs
- [x] Call `adapters.RecipeAdapter` to generate crafting recipes
- [x] Call `adapters.VehicleAdapter` to spawn vehicles in districts
- [x] Call `dungeon.Generate()` to create dungeon instances for quest objectives
- [x] Verify all generators produce output consumed by game systems

---

**Phase 2 Milestone: "Minimal Gameplay"**
- ✅ Player can interact with NPCs via E key
- ✅ Dialog system with branching conversations
- ✅ NPCs visible and animated in the world
- ✅ Basic melee combat with damage and death
- ✅ Inventory management (pickup, drop, use)
- ✅ Quest acceptance, tracking, and completion
- ✅ Procedurally generated quests, items, and buildings

---

## Phase 3: Persistence and Multiplayer (Week 4 — ~40 hours)

**Goal:** Game state persists between sessions. Client and server communicate game state.

### 3A: Save/Load Implementation (8 hours)

**Files:** `cmd/server/main.go`

**Steps:**
1. Import `pkg/world/persist/`
2. On server startup, check for existing save file and load
3. On server shutdown (SIGINT/SIGTERM), save world state
4. Periodic auto-save every N minutes
5. Save includes: entities, components, chunk modifications, quest states

**Completion Checklist:**

- [x] Import `pkg/world/persist/` in server
- [x] Implement save-file check and load on server startup
- [x] Implement world state save on server shutdown (SIGINT/SIGTERM handler)
- [x] Implement periodic auto-save at configurable interval
- [x] Verify save includes entities, components, chunk modifications, and quest states
- [x] Test save/load round-trip: save → restart → load → verify state matches

### 3B: Client-Server Protocol Implementation (16 hours)

**Files:** `cmd/server/main.go`, `cmd/client/main.go`, `pkg/network/`

**Steps:**
1. Server: Accept client connections, assign player entity
2. Server: On each tick, encode entity updates as `EntityUpdate` messages
3. Server: Stream chunk data when client enters new chunk
4. Client: Receive `WorldState` and `EntityUpdate` messages
5. Client: Apply server state to local ECS world
6. Client: Send `PlayerInput` messages to server
7. Server: Process `PlayerInput` messages, update authoritative state
8. Implement client-side prediction using `pkg/network/prediction.go`

**Completion Checklist:**

- [x] Server: accept client connections and assign player entity ID
- [x] Server: encode and broadcast `EntityUpdate` messages each tick
- [x] Server: stream `ChunkData` messages when client enters new chunk
- [x] Client: receive and decode `WorldState` messages
- [x] Client: receive and decode `EntityUpdate` messages
- [x] Client: apply server state to local ECS world
- [x] Client: send `PlayerInput` messages to server on each frame
- [x] Server: receive and process `PlayerInput` messages (validate, apply)
- [x] Implement client-side prediction using `pkg/network/prediction.go`
- [x] Verify two clients can connect and see each other's movements

### 3C: Character Creation Screen (8 hours)

**Files:** New file `cmd/client/character_creation.go`

**Implementation:**
1. Before game start, show character creation screen
2. Genre selection (5 genres with preview descriptions)
3. Name input
4. Starting skill allocation (distribute points across skill schools)
5. Starting equipment choice
6. Store selections, create player entity with chosen attributes

**Completion Checklist:**

- [x] Create character creation screen displayed before game start
- [x] Implement genre selection UI (5 genres with preview descriptions)
- [x] Implement player name text input
- [x] Implement starting skill point allocation across skill schools
- [x] Implement starting equipment choice
- [x] Create player entity with chosen attributes (genre, name, skills, equipment)
- [x] Verify character creation flows into game start seamlessly

### 3D: Pause Menu and Settings (8 hours)

**Files:** New file `cmd/client/menu.go`

**Implementation:**
1. Escape → Pause game, show menu overlay
2. Menu options: Resume, Settings, Save, Load, Quit
3. Settings: Key bindings, audio volume, graphics options
4. Settings persistence via config file update
5. Quit confirmation dialog

**Completion Checklist:**

- [x] Implement Escape key → pause game state and show menu overlay
- [x] Implement Resume menu option (return to gameplay)
- [x] Implement Settings submenu with key binding configuration
- [x] Implement Settings submenu with audio volume controls
- [x] Implement Settings submenu with graphics options
- [x] Implement settings persistence via config file update
- [x] Implement Save menu option (trigger server save)
- [x] Implement Load menu option (trigger server load)
- [x] Implement Quit menu option with confirmation dialog

---

**Phase 3 Milestone: "Persistent World"**
- ✅ Game saves and loads world state
- ✅ Client receives entity updates from server
- ✅ Multiplayer: multiple clients see each other
- ✅ Character creation with genre/skill choice
- ✅ Pause menu with settings

---

## Phase 4: Systems Depth and Polish (Weeks 5–6 — ~80 hours)

**Goal:** Deep gameplay systems active and interconnected. World feels alive.

### 4A: Advanced Combat (16 hours)

- Ranged combat: bow/gun with ProjectileSystem
- Magic combat: spell casting with MagicSystem
- Stealth: sneak mode, backstab, detection with StealthSystem
- Block/dodge mechanics
- Enemy AI: NPCs fight back using CombatState
- Health regeneration outside combat
- Death penalties (configurable via Difficulty settings)

**Completion Checklist:**

- [x] Implement ranged combat (bow/gun) with ProjectileSystem
- [x] Implement magic combat (spell casting) with MagicSystem
- [x] Implement stealth mode (sneak, backstab, detection) with StealthSystem
- [x] Implement block/dodge mechanics
- [x] Implement enemy AI combat (NPCs fight back using CombatState)
- [x] Implement health regeneration outside combat
- [x] Implement configurable death penalties via Difficulty settings
- [x] Verify all three combat styles (melee, ranged, magic) are functional

### 4B: Crafting and Economy (12 hours)

- Crafting UI with workbench interaction
- Resource gathering from ResourceNode components
- Recipe discovery through exploration
- Player shops using PlayerShopSystem
- Dynamic pricing visible at shop entities
- Trade routes visible on map

**Completion Checklist:**

- [x] Implement crafting UI with workbench interaction
- [x] Implement resource gathering from ResourceNode components
- [x] Implement recipe discovery through exploration
- [x] Integrate PlayerShopSystem for player-owned shops
- [x] Display dynamic pricing at shop entities
- [x] Display trade routes on minimap/map
- [x] Verify crafting end-to-end: gather materials → use workbench → produce item

### 4C: Faction and Crime Depth (12 hours)

- Faction rank UI showing standing with each faction
- Faction quest arcs using FactionArcManager
- Crime consequences: guards pursue, jail time
- Bounty display and bribery option
- Faction territory visual boundaries

**Completion Checklist:**

- [x] Implement faction rank UI showing standing with each faction
- [x] Integrate FactionArcManager for faction quest arcs
- [x] Implement crime consequences (guard pursuit, jail time)
- [x] Implement bounty display in HUD
- [x] Implement bribery option during guard encounters
- [x] Display faction territory visual boundaries on map/minimap
- [x] Verify faction reputation changes based on player actions

### 4D: Vehicle and Mount Integration (8 hours)

- Mount riding with MountSystem
- Vehicle physics with VehiclePhysicsSystem
- Cockpit/rider view transition
- Vehicle fuel/durability management

**Completion Checklist:**

- [x] Implement mount riding with MountSystem (mount/dismount, movement)
- [x] Implement vehicle physics with VehiclePhysicsSystem (steering, acceleration)
- [x] Implement cockpit/rider camera view transition
- [x] Implement vehicle fuel/durability management
- [x] Verify mount and vehicle entities are spawned in the world
- [x] Verify player can mount, ride, and dismount vehicles

### 4E: Weather and Environment Polish (8 hours)

- Weather particles: rain, snow, dust, ash
- Indoor/outdoor transitions with IndoorOutdoorSystem
- Environmental hazard damage with HazardSystem
- Time-of-day lighting via LightManager
- Genre-specific atmospheric effects

**Completion Checklist:**

- [x] Implement weather particles (rain, snow, dust, ash) via particles package
- [x] Implement indoor/outdoor transitions with IndoorOutdoorSystem
- [x] Implement environmental hazard damage with HazardSystem
- [x] Implement time-of-day lighting cycle via LightManager
- [x] Implement genre-specific atmospheric effects (e.g., neon glow for cyberpunk, fog for horror)
- [x] Verify weather visuals match WeatherSystem state

### 4F: NPC Life Simulation (12 hours)

- NPCs move between activities via NPCPathfindingSystem
- NPCs eat, sleep, socialize via NPCNeedsSystem
- NPCs work at jobs via NPCOccupationSystem
- NPCs gossip and share information via GossipSystem
- NPCs remember player interactions via NPCMemorySystem
- NPCs show emotions via EmotionalStateSystem
- Multi-NPC conversations visible in cities

**Completion Checklist:**

- [x] Verify NPCPathfindingSystem moves NPCs between activity locations
- [x] Verify NPCNeedsSystem drives NPC eat, sleep, and socialize behaviors
- [x] Verify NPCOccupationSystem makes NPCs perform their jobs
- [x] Verify GossipSystem propagates information between NPCs
- [x] Verify NPCMemorySystem tracks player interactions per NPC
- [x] Verify EmotionalStateSystem produces visible emotional reactions
- [x] Implement visible multi-NPC conversations in city areas
- [x] Verify NPCs follow daily schedule cycle (work → eat → socialize → sleep)

### 4G: Housing and PvP (12 hours)

- Player house purchase UI
- Furniture placement in first-person
- Guild territory management
- PvP zone indicators and flagging
- Loot drop mechanics in hostile zones

**Completion Checklist:**

- [x] Import `pkg/world/housing/` and integrate house purchase UI
- [x] Implement first-person furniture placement
- [x] Implement guild territory management
- [x] Import `pkg/world/pvp/` and display PvP zone indicators
- [x] Implement PvP flagging system
- [x] Implement loot drop mechanics in hostile/PvP zones
- [x] Verify player can buy, enter, and furnish a house

---

**Phase 4 Milestone: "Compelling Experience"**
- ✅ 3 combat styles: melee, ranged, magic
- ✅ Crafting system with recipes and workbenches
- ✅ Dynamic economy with player participation
- ✅ Faction progression and consequences
- ✅ Living NPC population with daily routines
- ✅ Vehicles and mounts rideable
- ✅ Weather affects gameplay
- ✅ Player housing functional

---

## Success Milestones Summary

| Milestone | Week | Success Criteria |
|-----------|------|-----------------|
| **Basic Playability** | 1 | Launch → see world → move with collision → HUD → exit cleanly |
| **Minimal Gameplay** | 2–3 | Interact with NPCs → accept quest → complete objective → combat |
| **Persistent World** | 4 | Save/load → multiplayer → character creation → settings |
| **Compelling Experience** | 5–6 | Combat depth → crafting → factions → living NPCs → housing |

## Effort Estimates

| Phase | Hours | Systems Activated | Packages Integrated |
|-------|-------|-------------------|---------------------|
| Phase 1 | ~40 | 57 (all registered) | +7 (rendering, audio, input) |
| Phase 2 | ~60 | — (using registered systems) | +4 (dialog, dungeon, adapters) |
| Phase 3 | ~40 | — | +2 (persist, network protocol) |
| Phase 4 | ~80 | — | +3 (housing, pvp, companion) |
| **Total** | **~220** | **57 systems** | **28 packages (all)** |

## Risk Factors

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| System registration order causes panics | Medium | High | Add nil-safe checks in Update methods; register in dependency order |
| Entity sync protocol too complex for single sprint | High | High | Implement offline single-player mode first; add sync later |
| 57 systems per tick exceeds 50ms budget | Medium | Medium | Profile; disable non-essential systems; spatial indexing for entity queries |
| Ebiten limitations for complex UI | Medium | Medium | Use `ebiten.Image` composition; consider external UI library |
| Venture V-Series generator API changes | Low | Medium | Pin version in go.mod; wrap with adapter pattern (already done) |
| Collision detection out-of-bounds panic | Medium | High | Bounds-check worldMap indices before access; clamp to map edges |

---

*Generated 2026-03-31 from AUDIT.md and GAPS.md analysis. See those documents for the evidence supporting this plan.*
