# Wyrm — Implementation Roadmap

## 1. Project Overview

Wyrm is a 100% procedurally generated first-person open-world RPG built in Go 1.24+ on the Ebiten v2 engine using an Entity-Component-System (ECS) architecture. Inspired by Elder Scrolls (open-world exploration, NPC schedules, faction politics), Fallout (post-apocalyptic tone, skill trees, dialogue consequences), and GTA (freeform crime/law systems, vehicles, persistent city life), Wyrm is the most ambitious W-Series title. It extends the V-Series — especially Venture's `pkg/procgen/` generators (terrain, entity, faction, quest, dialog, narrative, building, vehicle, magic, skills) — from a top-down roguelike to a persistent first-person open world. Every element is generated at runtime from a deterministic seed: no image files, no audio files, no level data. Five genre themes (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic) reshape every player-facing system from world aesthetics to NPC culture, making each playthrough a distinct RPG experience across 200–5000 ms network latency.

---

## 2. Core Architecture

**ECS:** Entities are `uint64` IDs. Components (`Position`, `Health`, `Faction`, `Schedule`, `Inventory`, `Vehicle`) are pure data structs in `pkg/engine/components/`. Systems in `pkg/engine/systems/` contain all logic and operate on component queries each tick.

**Key systems:** WorldChunkSystem, NPCScheduleSystem, FactionPoliticsSystem, CrimeSystem, EconomySystem, CombatSystem, VehicleSystem, QuestSystem, WeatherSystem, RenderSystem (first-person raycast + Ebiten draw calls), AudioSystem (procedural synthesis).

**V-Series reuse:** Import Venture's `pkg/procgen/{terrain,entity,faction,quest,dialog,narrative,building,vehicle,magic,skills,recipe,class,companion,station,story}` as direct Go module dependencies; wrap each generator in open-world adapters that pass `GenerationParams{GenreID, Difficulty, Depth, Custom}` per chunk/region.

**Project layout:**
```
cmd/client/        cmd/server/
pkg/engine/        pkg/procgen/       pkg/rendering/
pkg/audio/         pkg/network/       pkg/world/
```

---

## 3. Implementation Phases

### Phase 1 — Foundation (8 weeks)
*ECS core, Go module scaffold, V-Series integration, headless server skeleton*

1. Initialize Go module `github.com/opd-ai/wyrm`; import Venture as dependency; define `pkg/engine/ecs/` (entity registry, component store, system runner). **AC:** `go test ./pkg/engine/...` passes; 10,000 entities created/destroyed in <5 ms.
2. Implement `pkg/world/chunk/` — 512×512-unit chunks with deterministic seed per coordinate; wire Venture's `pkg/procgen/terrain` generator. **AC:** Same seed+genre produces identical chunk height maps across 3 independent runs.
3. Scaffold `cmd/server/` with authoritative TCP loop; `cmd/client/` with Ebiten window + stub renderer. **AC:** Client connects to server, receives empty world state, renders black screen without panic.
4. Implement `GenerationParams` genre routing: each Venture generator called with correct `GenreID`. **AC:** Terrain palette differs visually for all 5 genres on same seed; unit test asserts non-equal tile color arrays.

### Phase 2 — Open World & First-Person Rendering (10 weeks)
*Seamless chunk streaming, raycaster, procedural sprites/lighting, NPC schedules*

5. First-person raycaster in `pkg/rendering/raycast/` using Ebiten; **consider re-using (and possibly extending) the raycasting engine from opd-ai/violence to reduce code duplication and accelerate development**; procedural wall/floor/ceiling textures from `pkg/rendering/texture/` (no image files). **AC:** 60 fps at 1280×720 on reference hardware; genre changes wall palette.
6. Seamless chunk streaming: server sends delta-compressed chunk diffs; client pre-fetches 3×3 chunk window. **AC:** Player walks across chunk boundary with <1 frame stutter at 50 ms simulated latency.
7. NPC schedule system (via Venture's `pkg/procgen/entity` + new `NPCScheduleSystem`): sleep/work/wander/trade/patrol cycles keyed to world clock. **AC:** 100 NPCs in a city simulate 24 h in <10 ms server tick.
8. Procedural city generator in `pkg/procgen/city/` wrapping Venture's `pkg/procgen/building` and `pkg/procgen/station`; road network via minimum spanning tree on district centroids. **AC:** Genre-appropriate city (fantasy=stone/timber, cyberpunk=neon towers) generated in <200 ms.

### Phase 3 — Gameplay Systems (10 weeks)
*Combat, skills, factions, quests, economy, crime/law, vehicles*

9. Skill-based character progression using Venture's `pkg/procgen/skills` and `pkg/procgen/class`; 30 skills across 6 schools; genre renames schools (fantasy: Destruction/Alteration; sci-fi: Weaponry/Biotech). **AC:** All 5 genres produce distinct non-overlapping skill name sets.
10. Faction territory system extending Venture's `pkg/procgen/faction`: territory polygons, reputation scores per player, dynamic war/treaty events. **AC:** Killing 3 faction members triggers hostility within 10 server ticks; peace treaty reduces hostility.
11. Quest system using Venture's `pkg/procgen/quest` + `pkg/procgen/narrative`: branching quests with consequence flags persisted per player; min 5 branch points per major quest. **AC:** Completing quest branch A locks branch B; consequence flag persists across server restart.
12. Dynamic economy: supply/demand per city node, player-owned property (houses, shops) via Venture's `pkg/procgen/building`. **AC:** Selling 50 identical items to one vendor reduces buy price by ≥10%.
13. Crime/law system: wanted level 0–5 stars, NPC witness reports, bounty system, jail mechanic. **AC:** Committing crime within NPC LOS raises wanted level within 2 ticks; paying bounty resets to 0.
14. Vehicle system wrapping Venture's `pkg/procgen/vehicle`: 3 vehicle archetypes per genre (fantasy=horse/cart/ship; cyberpunk=hover/cycle/truck); physics approximated via ECS components. **AC:** All 5 genres spawn ≥3 distinct drivable vehicle types.

### Phase 4 — Audio & Visual Polish (6 weeks)
*Procedural audio synthesis, adaptive music, post-processing, genre visual effects*

15. Procedural audio `pkg/audio/`: oscillators + ADSR envelopes; genre SFX modifications (sci-fi +30% pitch, horror −30% + vibrato, cyberpunk +40% + hard clipping) matching V-Series `pkg/audio/sfx/` patterns. **AC:** SFX for footsteps differs measurably (pitch deviation >15%) across all 5 genres.
16. Adaptive music system: motifs per faction/region, combat intensity layer mixing, exploration vs combat transitions. **AC:** Music transitions within 2 s of entering combat; reverts within 5 s of last enemy death.
17. Genre post-processing in `pkg/rendering/postprocess/`: fantasy=warm color grade; sci-fi=scanlines; horror=vignette+desaturate; cyberpunk=chromatic aberration+bloom; post-apoc=sepia+grain. **AC:** Screenshot diff between genres on identical geometry >20% mean pixel delta.
18. Location-based ambient soundscapes: cave drip, city crowd murmur, wind on plains — synthesized, no files. **AC:** Ambient sound type changes within 1 s of entering new region type.

### Phase 5 — Multiplayer & Persistence (8 weeks)
*Shared persistent world, PvP, player housing, guild territories, economy sync, Tor latency tolerance*

19. Client-side prediction + server reconciliation for movement/combat per V-Series `pkg/network/` patterns. **AC:** Movement feels responsive at 200 ms RTT; no visible rubber-banding at 500 ms RTT.
20. World persistence: chunk state, NPC schedules, economy, quest flags serialized to disk on server; hot-reload on restart. **AC:** Server restart with same seed restores world state diff <5% from pre-restart snapshot.
21. PvP zone system: flagged regions with opt-in PvP, loot-on-death rules, respawn points. **AC:** Player flagged PvP in zone; killed player drops ≥1 inventory item; unflagged player takes no damage from flagged.
22. Player housing + guild territories: instanced interiors owned by players; guild claims on world map regions via faction system. **AC:** Player-placed furniture persists across 3 server restarts; guild territory boundary enforced.
23. Lag compensation: server rewinds entity state up to 500 ms for hit detection; Tor-mode activates at RTT >800 ms. **AC:** Hit registration correct at 500 ms simulated RTT in automated test harness.
24. Federation protocol: cross-server player travel, economy price signals, global event broadcasts. **AC:** Player moves between two local test servers retaining inventory and quest state.

### Phase 6 — Content Depth & Release (6 weeks)
*Dungeon generation, dialog depth, companion AI, end-to-end genre playthrough validation*

25. Dungeon generator `pkg/procgen/dungeon/` extending Venture's `pkg/procgen/terrain` with room graphs, trap/puzzle rooms via `pkg/procgen/puzzle`, boss encounters. **AC:** 100 generated dungeons have 0 unreachable rooms; each genre produces distinct tile aesthetics.
26. Deep dialog via Venture's `pkg/procgen/dialog` + `pkg/procgen/narrative`: topic memory, emotional state modifiers, genre-appropriate vocabulary sets. **AC:** NPC recalls player's previous interaction topic in follow-up conversation; emotional state (fearful/hostile/friendly) changes NPC response vocabulary; unit test asserts all 5 genres produce non-overlapping common word sets.
27. Companion AI using Venture's `pkg/procgen/companion`: persistent companion with unique personality, combat role, and genre-themed backstory. **AC:** Companion uses class-appropriate abilities; dialog references player actions from last 10 events.
28. Full genre playthrough validation: automated CI test boots each of 5 genres for 60 s simulated time and asserts non-overlapping city names, NPC faction names, item names, SFX pitch profiles, and terrain palettes. **AC:** CI passes for all 5 genres; zero name collisions across genre pairs.

---

## 4. PCG Systems Inventory

| Generator | Package (source) | Algorithm | Fantasy | Sci-Fi | Horror | Cyberpunk | Post-Apoc |
|---|---|---|---|---|---|---|---|
| World Terrain | Venture `pkg/procgen/terrain` + extend | Multi-octave Perlin + biome lookup | Forest/mountain/river | Crater/tundra/metallic plains | Swamp/fog/dead forest | Urban sprawl/toxic canal | Dust bowl/ruins/rad zones |
| City Layout | new `pkg/procgen/city/` | MST on district centroids + Venture building | Stone/timber districts | Arcology towers/domes | Abandoned districts/boarded | Neon megablock/underpass | Shantytown/fortified compound |
| Dungeon | new `pkg/procgen/dungeon/` | BSP room graph + Venture terrain tile | Catacombs/castle vault | Research facility/ship hold | Asylum/sewer labyrinth | Server farm/black market | Bunker/collapsed mall |
| NPC Entity | Venture `pkg/procgen/entity` | Stat template + name grammar | Peasant/knight/mage | Engineer/marine/android | Cultist/survivor/ghoul | Corpo/hacker/street samurai | Raider/settler/mutant |
| NPC Schedule | new `NPCScheduleSystem` | Finite-state machine on world clock | Medieval day cycle | Shift rotation/night ops | Erratic/fearful patterns | 24 h hustle/club hours | Patrol/scavenge cycles |
| Faction | Venture `pkg/procgen/faction` | Graph relations + territory Voronoi | Guild/kingdom/church | Corporation/military/cult | Cult/survivor group/demons | Megacorp/gang/resistance | Tribe/militia/trader guild |
| Quest | Venture `pkg/procgen/quest` | Template graph + consequence flags | Fetch/slay/diplomacy | Data heist/rescue/sabotage | Ritual/escape/investigation | Corp contract/street job | Scavenge/territory/rescue |
| Dialog | Venture `pkg/procgen/dialog` | Topic graph + sentiment model | Old English vocabulary | Technical jargon | Paranoid/whispering tone | Street slang/corpo speak | Gruff/desperate tone |
| Item/Weapon | Venture `pkg/procgen/recipe` | Affix table + material grammar | Iron/wood/enchanted | Alloy/polymer/energy cell | Bone/rusted/cursed | Synth/chrome/overclocked | Salvaged/jury-rigged/rad |
| Magic/Tech | Venture `pkg/procgen/magic` | Spell component grammar | Arcane/elemental magic | Energy weapons/implants | Dark ritual/curse | Cyberware/hacking | Chem boosts/jury-rigged tech |
| Vehicle | Venture `pkg/procgen/vehicle` | Archetype params + genre skin | Horse/cart/sailing ship | Hover-bike/shuttle/mech | Hearse/bone cart/raft | Motorbike/APC/drone | Buggy/armored truck/gyroplane |
| Building | Venture `pkg/procgen/building` | Room template + façade grammar | Castle/cottage/tavern | Station/lab/bunker | Haunted manor/asylum | Nightclub/server tower | Scrapyard/fortified house |
| Narrative | Venture `pkg/procgen/narrative` | Story arc templates | Epic heroism/prophecy | First contact/rebellion | Survival horror/madness | Corpo dystopia/resistance | Reclamation/survival |
| Audio SFX | new `pkg/audio/sfx/` + V-Series patterns | Oscillator + envelope + genre mods | Warm/acoustic timbre | +30% pitch/metallic | −30% pitch/vibrato | +40% pitch/hard clip | Muffled/distorted/crackle |
| Music | new `pkg/audio/music/` | Motif sequencer + layer mixing | Orchestral/lute motifs | Synth arpeggios/drone | Dissonant strings/silence | EDM/glitch | Folk/acoustic/wind |
| Visual Palette | new `pkg/rendering/palette/` | HSL shift + genre LUT | Warm gold/green | Cool blue/white/grey | Desaturated grey/green | Neon pink/cyan/black | Sepia/orange/dust |
| Post-Processing | new `pkg/rendering/postprocess/` | Shader-equivalent fragment passes | Warm color grade | Scanlines/bloom | Vignette/desaturate | Chromatic aberration/bloom | Sepia/grain |
| Ambient Sound | new `pkg/audio/ambient/` | Noise oscillator + filter envelopes | Wind/birds/water | Hum/static/ventilation | Drip/creak/distant screams | City noise/bass throb | Wind/fire/geiger clicks |

---

## 5. Multiplayer Design

**Authority model:** Server is authoritative for all world state. Clients send input commands; server applies, broadcasts delta states.

**Chunk streaming:** Server tracks each client's 3×3 active chunk window. On chunk entry, server sends full chunk snapshot (compressed). On exit, server stops sending that chunk's deltas. Client interpolates entity positions between server snapshots.

**NPC authority:** Server owns all NPC state (schedule, faction, dialog). Clients render NPCs via interpolated position components. NPC AI runs server-side only.

**Quest instances:** Instanced quests (dungeon runs) spin up a sub-world with a unique seed derived from a stable 64-bit hash / mixing function over `(worldSeed, questID, playerGroupID)` (instead of a simple XOR). Party members join the same instance. Completion state written back to persistent world.

**PvP zones:** Regions flagged `PvPZone=true` by faction generator. Client UI shows zone boundary. Server enforces damage rules by zone flag on combat resolution.

**Economy sync:** City economy nodes sync price arrays via gossip protocol between server tick groups. Player transactions broadcast economy delta to all nodes within 2-city radius.

**Player housing:** Interior chunks linked to player `EntityID`. Furniture component layout serialized per interior chunk. Loaded on demand when any player enters.

**Lag compensation:** Server maintains 500 ms entity position history ring buffer. Each client input is tagged with a monotonic tick sequence number and mapped to server time via a synchronized clock offset. On hit registration, the server rewinds to the corresponding server timestamp (clamped to the last 500 ms of history and within an allowed drift window), checks AABB overlap, then re-advances; timestamps outside the drift window are rejected and processed at current server time.

**Tor/high-latency mode:** At RTT >800 ms, client increases prediction window to 1500 ms, reduces input send rate to 10 Hz, and enables aggressive visual interpolation with 300 ms blend time.

**Federation:** Cross-server travel via redirect token. Sending server serializes player ECS state bundle; destination server deserializes. Economy prices gossip across federation peers via UDP multicast on LAN or TCP on WAN.

---

## 6. Genre Differentiation Matrix

| Dimension | Fantasy | Sci-Fi | Horror | Cyberpunk | Post-Apocalyptic |
|---|---|---|---|---|---|
| World theme | Magic-infused medieval continent | Colony planet / space station | Cursed island / haunted city | Megacity sprawl 2140 | Irradiated wasteland |
| Faction archetypes | Kingdoms, guilds, church orders | Corporations, military fleets, cults | Cults, survivor bands, monsters | Megacorps, street gangs, hackers | Tribes, raider clans, trader caravans |
| Magic/tech system | Elemental spells, enchantments, alchemy | Energy weapons, cybernetic implants, hacking | Dark rituals, curses, exorcism | Cyberware, ICE hacking, drone ops | Chem boosters, jury-rigged explosives |
| Vehicle types | Horse, war cart, sailing ship | Hover-bike, shuttle, exo-mech | Bone cart, plague barge, hearse | Motorbike, APC, aerial drone | Dune buggy, armored bus, gyrocopter |
| Quest themes | Prophecy, diplomacy, monster hunts | Data heist, first contact, rebellion | Ritual disruption, escape, survival | Corp contract, turf war, data theft | Scavenge, territory claim, rescue |
| Hazards | Dragons, floods, magic surges | Radiation zones, AI rogue, vacuum | Madness meter, monster spawns, darkness | EMP bursts, hacker intrusion, smog | Rad zones, mutant hordes, dust storms |
| Music style | Orchestral / lute / choir | Synth arpeggios / electronic drone | Dissonant strings / silence / stings | EDM / glitch / bass | Acoustic folk / wind instruments |
| Visual palette | Warm gold/green, soft lighting | Cool blue/white, hard edges, glow | Desaturated grey-green, deep shadow | Neon pink/cyan on black, bloom | Sepia/orange dust, grain overlay |
| NPC culture | Medieval social hierarchy | Meritocracy / military rank | Fearful isolation / cult devotion | Class stratification / street code | Survival tribalism / barter culture |
| Post-processing | Warm color grade | Scanlines + edge bloom | Vignette + desaturate | Chromatic aberration + bloom | Sepia + film grain |

---

## 7. Success Criteria

| Indicator | Target | Measurement |
|---|---|---|
| Deterministic PCG | Identical world on same seed+genre | Automated test: 3 independent runs, byte-compare chunk data |
| Genre differentiation | ≥5 distinct systems differ per genre pair | CI test: assert non-equal names, palettes, SFX pitch per genre |
| Performance (server) | 200 NPCs, 32 players, ≤20 ms tick | Benchmark: `go test -bench=BenchmarkServerTick` |
| Performance (client) | 60 fps at 1280×720 | Ebiten FPS counter in headless Xvfb CI run |
| Multiplayer latency tolerance | Playable at 500 ms RTT | Automated test: netem 500 ms, play 60 s, assert 0 desync events |
| Tor-mode (mid latency) | Functional at 2000 ms RTT | Netem 2000 ms test: movement and combat resolve correctly |
| Tor-mode (upper bound) | Functional at 5000 ms RTT (hard constraint) | Netem 5000 ms test: player can move, cast, and receive world updates; no client crash |
| Zero static assets | No files in `assets/` at build time | CI: `find assets/ -type f ! -name '.gitignore' | wc -l` == 0 |
| Single binary | `go build ./cmd/client` produces one binary | CI: binary runs without external files on clean machine |
| Genre playthrough | All 5 genres boot and run 60 s without panic | CI matrix job: one run per genre |
| Quest persistence | Quest flags survive server restart | Integration test: restart server, assert flag present |

---

## 8. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| First-person raycaster performance in Ebiten | Medium | High | Profile early (Phase 2); fall back to simplified billboard rendering if needed |
| V-Series API surface changes breaking imports | Low | Medium | Pin Venture to a specific module version; maintain adapter layer |
| Open-world chunk streaming bandwidth | Medium | High | Delta compression + interest management; benchmark in Phase 2 before full NPC load |
| NPC schedule simulation CPU cost at scale | Medium | High | Coarse-grained time steps for distant NPCs; LOD simulation tiers |
| Deterministic PCG across Go versions | Low | High | Seed own PRNG (`xoshiro256**`); never use `math/rand` global state |
| Tor latency making game feel unplayable | Medium | High | Generous client-side prediction; animate all local actions immediately |
| Five-genre scope creep | High | Medium | Genre routing is a single `GenreID` parameter; enforce via interface; add genres incrementally |
| Economy exploit in player trading | Medium | Medium | Server-side price bounds clamping; rate-limit transactions per player per minute |

---

## 9. V-Series PCG Library Reuse Guide

Wyrm does not build from scratch. It imports, adapts, and extends the procedural generation, rendering, audio, and networking subsystems already proven in the four V-Series sibling games. This section is the authoritative reference for what to import from each repo and how to adapt it.

### 9.1 Venture (`opd-ai/venture`) — The Content Engine

Venture's `pkg/procgen/` (~25 sub-packages, ~399k LOC total repo) is the richest procedural content library in the suite. Wyrm treats it as a **direct Go module dependency** and wraps every generator in an open-world adapter that passes `GenerationParams{GenreID, Difficulty, Depth, Custom}` per chunk or region.

**Generators to import and their Wyrm adaptation:**

| Venture Package | Wyrm Adapter | Key Extension for Open World |
|-----------------|-------------|------------------------------|
| `procgen/terrain` (BSP, cellular, city, forest, composite, grammar, maze) | `pkg/world/chunk/` adapter | Tile terrain across 512×512 chunk boundaries; add open-world biome transitions; height-map for 3D raycaster |
| `procgen/entity` (stat templates, name grammar) | NPC factory in `WorldChunkSystem` | Add `Schedule`, `Reputation`, `Memory` components; wire into `NPCScheduleSystem` for daily routines |
| `procgen/faction` (graph relations, territory Voronoi) | `FactionPoliticsSystem` seed input | Layer dynamic war/treaty events; territory shifts at runtime driven by player actions and NPC battles |
| `procgen/quest` (template graphs, consequence flags) | `QuestSystem` template source | Persistent consequence flags (survive server restart); cross-quest dependency chains; journal tracking |
| `procgen/dialog` (topic graph, sentiment model) | Dialog presenter in client | NPC memory (recall previous topics); emotional state modifiers; genre-vocabulary overlays |
| `procgen/narrative` (story arc templates) | World event seeder | Multiple concurrent arcs per player; faction-driven plot forks; environmental storytelling triggers |
| `procgen/building` (room templates, façade grammar) | City generator + housing system | Genre-specific exteriors; player-placeable furniture; instanced interiors linked to owner EntityID |
| `procgen/vehicle` (archetype params, genre skin) | `VehicleSystem` entity factory | First-person cockpit rendering; physics (steering, acceleration, fuel); mount/dismount transitions |
| `procgen/magic` (spell component grammar) | Magic/ability subsystem | Cooldown system; mana/energy costs; visual effect triggers for raycaster; spell combos |
| `procgen/skills` (skill trees across schools) | Progression system | XP-based unlocks; genre-renamed schools; multi-class hybrid picks |
| `procgen/class` (character archetypes) | Character creation + NPC assignment | Multi-class hybrid support; genre-appropriate archetype names |
| `procgen/companion` (personality, combat role, backstory) | Companion AI system | Follow/fight/wait commands; relationship score; dialog references to shared adventures |
| `procgen/recipe` (affix table, material grammar) | Crafting system | First-person workbench UI; material gathering from world objects; quality tiers |
| `procgen/station` (shop generation) | POI placement + economy | Dynamic supply/demand pricing via `EconomySystem`; shop hours keyed to world clock |
| `procgen/story` (story events) | World event seeds | Persistent consequences (destroyed buildings stay destroyed, killed NPCs stay dead) |
| `procgen/book` (lore text) | Interior decoration | Place generated books in building interiors via furniture system |
| `procgen/furniture` (interior placement) | Housing system | Player-placeable furniture; interior decoration generation for NPC buildings |
| `procgen/environment` (detail generation) | Chunk decoration pass | Trees, rocks, debris, genre-appropriate clutter seeded per chunk |
| `procgen/item` (weapons, armor, consumables) | Loot + inventory system | Durability; enchantment; repair; weight-based carry capacity |
| `procgen/legendary` (unique items) | Boss/quest rewards | Wire into boss drop tables and quest reward pools |
| `procgen/puzzle` (puzzle rooms) | Dungeon puzzle encounters | Adapt from top-down interaction to first-person (lever pulling, object manipulation) |
| `procgen/minigame` (templates) | Crafting/lockpicking/hacking | Adapt UI from top-down to first-person or overlay-screen |
| `procgen/genre` (registry, blending) | Canonical genre source | `cfg.Genre` maps to Venture's genre registry; single source of truth |
| `procgen/audit` (output validation) | World validation CI | Extended checks: chunk connectivity, NPC schedule coverage, economy balance |

### 9.2 Violence (`opd-ai/violence`) — The First-Person Engine

Violence (~140 packages) provides Wyrm's first-person rendering infrastructure, real-time combat resolution, and multiplayer networking.

**Key imports:**

| Violence Package | What Wyrm Takes | Adaptation Notes |
|------------------|-----------------|------------------|
| `pkg/raycaster` (DDA raycaster, trig tables) | **Primary rendering engine** — wall/floor/ceiling casting | Extend for multi-height walls (buildings), open terrain elevation, skybox, procedural LOD |
| `pkg/audio` (synthesis, ambient, reverb, SFX) | Oscillator + ADSR + reverb core | Add genre-specific SFX mods, adaptive music layer, 3D spatial audio for open world |
| `pkg/network/lagcomp` (500ms rewind buffer) | Server lag compensation | Transplant directly; extend rewind window from Violence's 500ms to configurable 500–1500ms for Tor mode |
| `pkg/network/delta` (delta compression) | Chunk/entity state diffs | Apply to chunk streaming; extend with entity-level interest management |
| `pkg/network/latency` (RTT estimation) | Latency monitoring + Tor-mode trigger | Add automatic quality degradation (reduce tick rate, increase prediction) at high latency |
| `pkg/network/anticheat` (server-side validation) | Input validation | Extend with RPG-specific checks (impossible stat values, speed hacks, economy exploits) |
| `pkg/combat/spatial_hash` (spatial partitioning) | Entity collision queries | Use for NPC aggro ranges, area-of-effect spells, and proximity-based interactions |
| `pkg/combat/positional` (direction-based damage) | Melee/ranged hit resolution | Add RPG stat modifiers (strength, weapon skill, armor rating); stealth backstab multiplier |
| `pkg/combat/telegraph` (attack warnings) | Enemy attack indicators | Extend with magic casting indicators, environmental hazard warnings |
| `pkg/combat/boss_phase` (multi-phase boss AI) | Dungeon boss encounters | Add RPG boss mechanics: enrage timers, phase-locked invulnerability, add-spawning |
| `pkg/collision` (detection + response) | Player-world collision | Extend with NPC pathfinding obstacles, vehicle collision, destructible objects |
| `pkg/camera` / `pkg/camerafx` (FP camera + effects) | First-person view system | Add dialog camera (face NPC), vehicle camera, cinematic camera, scope/zoom |
| `pkg/weapon` / `pkg/weaponanim` / `pkg/weaponsway` (viewmodels) | Weapon rendering | Add melee weapons, magic staves, genre-specific models; extend with weapon condition visual degradation |
| `pkg/particle` (emitters + lifecycle) | Spell/explosion/weather effects | Extend with Wyrm's weather (rain, snow, dust), magic auras, environment particles |
| `pkg/fog` / `pkg/lighting` (atmosphere) | Atmospheric rendering | Add time-of-day cycle, genre-specific atmosphere (perpetual fog for horror, neon for cyberpunk) |
| `pkg/texture` / `pkg/walltex` (procedural textures) | Base texture pipeline | Add biome-aware textures, genre palette application, wear/age effects |
| `pkg/loot` / `pkg/inventory` / `pkg/equipment` (item management) | Inventory data structures + equipment slots | Extend with weight capacity, durability, paper-doll UI, equipment comparison |
| `pkg/ai` (patrol/chase/attack FSM) | NPC combat AI base | Layer schedule-awareness: NPC transitions from routine→alert→combat→resume |
| `pkg/hazard` / `pkg/trap` (environmental dangers) | Dungeon + open-world hazards | Add radiation zones, magic anomalies, weather dangers, genre-specific hazards |
| `pkg/bsp` (BSP tree) | Building interior geometry | Extend for indoor/outdoor transitions; combine with Venture's building generator |
| `pkg/decal` / `pkg/destruct` (surface effects) | Environmental effects | Add weapon impact decals, spell burn marks, environmental destruction |

### 9.3 Velocity (`opd-ai/velocity`) — Patterns & Architecture

Velocity is a Galaga-style shmup (~25 packages), so direct code import is limited. Its value is in **architectural patterns**.

| Velocity Package | Wyrm Pattern Reuse |
|------------------|-------------------|
| `pkg/procgen/spawner` (deterministic spawning, difficulty curves) | Adapt for open-world enemy placement density and dungeon depth scaling |
| `pkg/procgen/wave_manager` (escalating encounters) | Adapt wave concept for siege events, bandit raids, dungeon encounter rooms |
| `pkg/procgen/genre` (genre routing) | Match genre enum for suite-wide consistency |
| `pkg/audio` (Ebiten integration + stubs) | Match audio architecture pattern (platform-aware Ebiten layer + test stubs) |
| `pkg/rendering/culling` (off-screen entity culling) | Adapt culling algorithm for raycaster (don't render entities behind walls or beyond draw distance) |
| `pkg/balance` (parameter tuning framework) | Adapt for RPG stat balancing (damage formulas, XP curves, economy pricing) |
| `pkg/companion` (real-time AI positioning) | Adapt companion follow/formation logic for Wyrm's 3D NPC companions |
| `pkg/config` (Viper configuration) | Match loading pattern for suite-wide consistency |

### 9.4 Vania (`opd-ai/vania`) — Reference Implementations

Vania uses `internal/` packages (not directly importable as Go modules). Its value is as **reference implementations** to replicate in Wyrm's `pkg/` tree.

| Vania Package | Pattern to Replicate |
|---------------|---------------------|
| `internal/pcg/seed` (FNV-based seed mixing) | **Canonical seed derivation pattern.** Replicate in Wyrm's `pkg/procgen/` for chunk, quest, and instance seed derivation |
| `internal/pcg/cache` (LRU content caching) | **Caching pattern** for generated chunks, textures, building interiors, and NPC templates |
| `internal/pcg/validator` (output validation) | **Validation pattern** for world integrity checks (chunk connectivity, quest reachability, economy solvability) |
| `internal/camera` (smooth follow + bounds) | Camera interpolation math for look-direction smoothing and position transitions |
| `internal/physics` (gravity, collision, slopes) | Physics architecture for Wyrm's 3D approximated physics (projectile arcs, vehicle movement, falling) |
| `internal/animation` (sprite animation states) | Animation state machine pattern for weapon viewmodels and NPC animation |
| `internal/narrative` (runtime story delivery) | Narrative presentation pattern: text boxes, cutscene triggers, environmental storytelling cues |
| `internal/particle` (emitter/lifecycle pattern) | Particle system architecture (shared with Violence's more mature implementation) |

---

## 10. How Wyrm Distinguishes Itself from V-Series Games

### 10.1 Core Identity

The V-Series games are **session-based** or **run-based** experiences:
- **Venture** = session-based co-op roguelike (enter dungeon → clear rooms → extract)
- **Violence** = match-based FPS (join → play round → next map)
- **Velocity** = run-based shmup (start → survive waves → game over)
- **Vania** = single-player campaign (linear progression through generated world)

**Wyrm is a persistent, seamless open world.** The world exists continuously on the server. Players inhabit, modify, and leave lasting marks on a shared world. There are no "runs," "matches," or "levels" — just an ongoing life in a generated universe.

### 10.2 Architectural Distinctions

| Dimension | V-Series Approach | Wyrm Approach |
|-----------|-------------------|---------------|
| World scope | Bounded levels / rooms / arenas / stages | Seamless infinite terrain via chunk streaming |
| Persistence | Session state lost on exit or save-file snapshots | Server-authoritative persistent world (survives restarts) |
| NPC behavior | Simple AI states (patrol / chase / attack) | Full daily schedules, topic memory, reputation, relationships |
| Economy | Static loot tables or fixed shop inventories | Dynamic supply / demand, player-driven market, property ownership |
| Faction system | Static teams or allegiance flags | Dynamic territorial control, inter-faction wars, per-player reputation |
| Quest system | Linear or template-based single objectives | Branching narratives with persistent consequences that reshape the world |
| Player agency | Combat + movement | Combat, dialog, crafting, trading, stealing, building, governing, exploring |
| Multiplayer | Match / session with lobby or single-player | Persistent shared world with federated cross-server travel |
| Perspective | Top-down / side-scroll / fixed first-person arenas | First-person immersive with contextual camera (dialog, vehicle, cinematic) |
| Game loop | Tight action loops (seconds to minutes) | Long-form RPG loops (hours to hundreds of hours) |
| World mutation | Level-reset between runs/matches | Permanent world changes (buildings destroyed, NPCs killed, territory conquered) |

### 10.3 Unique Wyrm Systems (Not Present in Any V-Series Game)

1. **Persistent World Clock** — Time-of-day / calendar driving NPC schedules, shop hours, faction events, weather cycles, seasonal changes
2. **NPC Memory & Relationships** — NPCs remember player actions, form opinions, spread gossip, change behavior over time
3. **Property Ownership** — Buy / build / furnish houses; own shops with passive income; claim territory for guilds
4. **Crime & Law Simulation** — Witness-based detection, bounties, jail, reputation ripple effects
5. **Deep Crafting** — Material gathering, workbench minigames, recipe discovery, skill-based quality scaling
6. **Environmental Storytelling** — Ruined buildings from faction wars, abandoned camps, environmental quest clues
7. **Dynamic World Events** — Sieges, plagues, market crashes, dragon attacks, coups — reshaping the world
8. **Cross-Server Federation** — Seamless travel between server instances with inventory / quest persistence
9. **First-Person Dialog** — Face-to-face NPC conversations with expression rendering and gesture animation
10. **Vehicle Exploration** — Horses, buggies, airships — traversal as a core open-world mechanic

---

## 11. Exhaustive Feature Target List

This section catalogs **every** feature Wyrm targets, organized by system. Features marked with ★ go significantly beyond what any V-Series game implements. Features marked with (V) indicate reuse from a V-Series library.

### 11.1 Open World & Exploration

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 1 | ★ Seamless chunk streaming | Infinite terrain via 512×512 chunks loaded/unloaded around the player with zero stutter | New (chunk adapter around Venture terrain) |
| 2 | ★ Multi-biome open world | Forests, deserts, mountains, swamps, tundra, ocean coasts — biome determined by Perlin noise + genre | (V) Venture `procgen/terrain` extended |
| 3 | ★ Vertical terrain | Hills, cliffs, valleys, caves — height-mapped terrain rendered via extended raycaster | New |
| 4 | ★ Day / night cycle | 24-hour world clock affecting lighting, NPC behavior, enemy spawns, shop availability | New |
| 5 | ★ Weather system | Rain, snow, fog, sandstorms, thunderstorms — genre-specific weather events affecting gameplay | New |
| 6 | ★ Seasonal changes | Terrain palette shifts, crop availability, NPC dialogue references, festival events | New |
| 7 | ★ Procedural road network | Roads connecting cities/towns via MST on settlement centroids with terrain-following pathfinding | New |
| 8 | ★ Points of interest | Ruins, camps, shrines, caves, towers scattered across wilderness with loot and lore | (V) Venture `procgen/environment` extended |
| 9 | ★ Underwater exploration | Submersible areas with breath meter, underwater creatures, sunken treasure, genre-specific underwater environments | New |
| 10 | ★ Fast travel network | Discovered locations become fast-travel points (carriages/teleporters/shuttles per genre) | New |
| 11 | World map | Procedurally filled map revealed by exploration; fog-of-war on unvisited areas | New |
| 12 | ★ Environmental hazards | Radiation zones, magic anomalies, toxic swamps, lava flows — genre-specific dangers requiring preparation | (V) Violence `pkg/hazard` extended |
| 13 | ★ Discoverable landmarks | Unique generated structures (ancient towers, crashed ships, ritual circles) with associated lore | New |
| 14 | ★ Hidden areas | Secret caves, hidden passages, illusory walls — rewarding thorough exploration | New |
| 15 | ★ Climbing & traversal | Climb ledges, swim rivers, vault obstacles — movement beyond flat ground | New |

### 11.2 Cities, Settlements & Civilization

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 16 | ★ Procedural cities | Multi-district cities with residential, commercial, industrial, and governmental zones | New (wraps Venture `procgen/building` + `procgen/station`) |
| 17 | ★ City population simulation | Hundreds of NPCs with daily schedules (sleep/work/eat/socialize/patrol) | (V) Venture `procgen/entity` + new `NPCScheduleSystem` |
| 18 | ★ Dynamic shop hours | Shops open/close based on world clock; merchants have personal schedules | New |
| 19 | ★ City guards & law enforcement | Patrol routes, crime response, chase/arrest behavior, jailing | New |
| 20 | ★ Town crier / notice boards | Procedurally generated news reflecting recent world events, quest hooks | New |
| 21 | ★ Taverns & inns | Rest to advance time, hear rumors (quest hooks), hire companions, gamble | (V) Venture `procgen/building` extended |
| 22 | Villages & hamlets | Smaller settlements with fewer services, different architecture | (V) Venture `procgen/building` |
| 23 | ★ Settlement growth/decline | Cities grow or shrink based on economy and safety; new buildings constructed, abandoned ones decay | New |
| 24 | ★ Market districts | Open-air markets with vendor stalls, dynamic pricing, haggling minigame | New |
| 25 | ★ Sewers & underground | Below-city dungeon layer accessible from surface; connects to criminal underground | New |

### 11.3 NPCs & Social Simulation

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 26 | ★ NPC daily schedules | Wake, eat, work, socialize, worship, sleep — every NPC has a 24-hour routine | New |
| 27 | ★ NPC memory | NPCs remember the player's actions, gifts, crimes, and quest outcomes | New |
| 28 | ★ NPC relationships | NPCs have opinions of each other; friendships, rivalries, romances between NPCs | New |
| 29 | ★ NPC gossip network | Information spreads between NPCs — commit a crime and witnesses tell others | New |
| 30 | ★ NPC emotional states | Fear, anger, gratitude, suspicion — affecting dialogue tone and behavior | (V) Venture `procgen/dialog` sentiment extended |
| 31 | ★ Hireable NPCs | Hire mercenaries, mages, guides, pack mules — each with unique personality | (V) Venture `procgen/companion` extended |
| 32 | ★ NPC mortality | Important NPCs can die permanently — the world reacts (funerals, succession, quest branches close) | New |
| 33 | ★ Children & families | NPC family units with child NPCs (non-combatant); family ties affect quests | New |
| 34 | ★ NPC reactions to player equipment | NPCs comment on player armor, weapons, faction insignia; guards react to drawn weapons | New |
| 35 | ★ Beggar / street performer NPCs | Ambient city life: beggars, musicians, preachers, drunks — genre-appropriate | New |
| 36 | Deep dialog trees | Multi-path conversations with skill checks, persuasion, intimidation, bribery | (V) Venture `procgen/dialog` |
| 37 | ★ Dialog consequences | Conversation choices affect NPC disposition, quest availability, faction standing | New |
| 38 | ★ Lie detection / Persuasion | NPCs can detect lies based on player Charisma; persuasion skill opens dialog options | New |

### 11.4 Quests & Narrative

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 39 | Branching quest lines | Min 5 decision points per major quest; branches lock out alternatives permanently | (V) Venture `procgen/quest` extended |
| 40 | ★ Persistent consequences | Quest outcomes permanently change the world (NPCs alive/dead, buildings intact/destroyed) | New |
| 41 | ★ Faction quest chains | Multi-quest story arcs per faction with rising stakes and mutual exclusivity between factions | New |
| 42 | ★ Radiant quests | Infinitely generated side quests (bounty hunts, fetch, delivery, escort) from notice boards | (V) Venture `procgen/quest` template extended |
| 43 | ★ World-changing events | Completing major quests can trigger sieges, regime changes, new areas opening | New |
| 44 | ★ Player-driven narrative | Player actions (not just quest choices) influence the world — kill a merchant, lose a trade route | New |
| 45 | ★ Moral ambiguity | No quest has a clearly "right" answer; every choice involves tradeoffs | New (extends Venture `procgen/narrative`) |
| 46 | ★ Dynamic quest generation | New quests generated in response to world state (famine → food quest, war → spy quest) | New |
| 47 | ★ Quest journal & tracking | In-world journal tracking active, completed, and failed quests with map markers | New |
| 48 | ★ Quest timers | Some quests have deadlines; failure to act in time has consequences | New |
| 49 | ★ Multi-player quest cooperation | Party members share quest progress; instanced dungeons for quest climaxes | New |

### 11.5 Combat & Skills

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 50 | First-person melee combat | Sword/axe/mace with timing-based blocking and directional attacks | (V) Violence `pkg/combat` adapted |
| 51 | First-person ranged combat | Bows/guns/crossbows with projectile physics and distance falloff | (V) Violence `pkg/combat` adapted |
| 52 | ★ Magic / ability system | 30+ spells across 6 schools, each genre-renamed; visual effects via raycaster | (V) Venture `procgen/magic` extended |
| 53 | ★ Stealth system | Sneak, pickpocket, backstab — light/shadow, noise, line-of-sight detection | New |
| 54 | ★ Status effects | Poison, burning, frozen, stunned, blinded, cursed — 20+ status effects with visual indicators | New |
| 55 | ★ Dual-wielding | Wield weapon + shield, two weapons, or weapon + spell in off-hand | New |
| 56 | ★ Combo system | Chain attacks for bonus damage; different combos per weapon type | (V) Violence `pkg/combat/combo` adapted |
| 57 | ★ Dodge / roll | Active dodge mechanic with i-frames; stamina cost | New |
| 58 | ★ Critical hits | Location-based critical hits (headshots, backstabs) with bonus damage and effects | (V) Violence `pkg/combat/positional` extended |
| 59 | ★ Skill-based progression | 30 skills across 6 schools; skills improve by use (Elder Scrolls-style) | (V) Venture `procgen/skills` extended |
| 60 | ★ Perk trees | Unlockable perks at skill milestones granting unique abilities | New |
| 61 | ★ Enemy scaling | Enemies scale to player level / area danger level; elite and legendary variants | New |
| 62 | ★ Boss encounters | Multi-phase bosses in dungeons with unique mechanics and loot | (V) Violence `pkg/combat/boss_phase` extended |
| 63 | ★ Mounted combat | Fight from horseback / vehicles with modified attack patterns | New |
| 64 | ★ Siege combat | Large-scale faction battles with siege weapons, walls, and objectives | New |
| 65 | ★ Arena combat | Gladiatorial arenas in cities with ranked challenges and betting | New |

### 11.6 Character Progression & Classes

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 66 | Character creation | Choose race, class, background, starting skills; appearance procedurally generated | (V) Venture `procgen/class` + `procgen/entity` |
| 67 | ★ Level-less progression | Skills improve through use, not XP-based leveling (learn by doing) | New (inspired by Elder Scrolls) |
| 68 | ★ Multi-class freedom | No class lock — any character can learn any skill, but specialization yields bonuses | New |
| 69 | ★ Attribute system | Strength, Dexterity, Intelligence, Charisma, Constitution, Perception — genre-renamed | New |
| 70 | ★ Reputation system | Per-faction, per-city, and global fame/infamy affecting NPC behavior and quest access | New |
| 71 | ★ Title system | Earn titles through achievements (Dragonslayer, Archmage, Crime Lord) affecting dialog options | New |
| 72 | ★ Racial/background bonuses | Starting bonuses and unique dialog options based on character background | New |

### 11.7 Inventory, Crafting & Economy

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 73 | Inventory system | Grid/list inventory with item categories, stacking, and sorting | (V) Violence `pkg/inventory` extended |
| 74 | Equipment system | Head, chest, legs, feet, hands, ring×2, amulet, weapon, shield/off-hand slots | (V) Violence `pkg/equipment` extended |
| 75 | ★ Weight/encumbrance | Items have weight; over-encumbered slows movement; pack animals and containers help | New |
| 76 | ★ Item durability & repair | Equipment degrades with use; repair at workbenches or blacksmiths; condition affects stats | New |
| 77 | ★ Crafting system | Gather materials → use workbench → crafting minigame → produce item; skill affects quality | (V) Venture `procgen/recipe` extended |
| 78 | ★ Alchemy / cooking | Combine ingredients for potions/food with procedurally generated effects | New |
| 79 | ★ Enchanting | Add magical properties to equipment using soul gems/power cells/genre equivalent | New |
| 80 | ★ Mining / harvesting | Extract resources from world nodes (ore veins, herb patches, salvage piles) | New |
| 81 | ★ Dynamic economy | Supply/demand per city node; prices fluctuate based on player actions and world events | New |
| 82 | ★ Player-owned shops | Buy a shop → stock inventory → earn passive income; hire NPC shopkeeper | New |
| 83 | ★ Black market | Fence stolen goods; buy illegal items; accessible through crime contacts | New |
| 84 | ★ Haggling minigame | Negotiate prices with merchants; success depends on Charisma skill | New |
| 85 | ★ Trade routes | Cities connected by trade routes; disrupting routes affects prices; escort merchant caravans | New |

### 11.8 Factions & Politics

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 86 | Faction generation | 5–8 factions per genre with unique names, culture, territory, and relationships | (V) Venture `procgen/faction` |
| 87 | ★ Faction territory control | Factions hold regions on the world map; borders shift through war and diplomacy | New (extends Venture Voronoi) |
| 88 | ★ Faction wars | Factions wage war — sieges, skirmishes, territory capture; player can participate or avoid | New |
| 89 | ★ Faction diplomacy | Treaties, alliances, trade agreements between factions; player actions influence negotiations | New |
| 90 | ★ Faction rank progression | Join a faction → complete tasks → rise through ranks → unlock exclusive quests, gear, and abilities | New |
| 91 | ★ Faction betrayal | Double-agent quests; defecting to rival factions with consequences | New |
| 92 | ★ Faction-exclusive content | Each faction has unique quest lines, gear sets, skills, and story arcs | New |
| 93 | ★ Civil war mechanic | Major factions can split into civil war; player's faction choice affects world state | New |
| 94 | ★ Coup / regime change | NPCs can overthrow faction leaders; player can instigate or prevent coups | New |
| 95 | ★ Faction economy | Factions have treasuries, tax income, military spending; economy affects military strength | New |

### 11.9 Crime & Law

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 96 | ★ Crime detection | Crimes only detected if witnessed by NPCs with line-of-sight; stealth avoids witnesses | New |
| 97 | ★ Wanted level (0–5 stars) | Escalating law enforcement response: patrol → investigate → chase → overwhelm → military | New |
| 98 | ★ Bounty system | Bounties accrue per faction; pay at jails or kill witnesses to reduce | New |
| 99 | ★ Jail mechanic | Arrested player serves time (accelerated); loses stolen goods; can attempt jailbreak | New |
| 100 | ★ Pickpocketing | Skill-based minigame; risk/reward scales with target value and detection chance | New |
| 101 | ★ Breaking & entering | Lockpicking minigame for doors/chests; guards respond to broken locks | New |
| 102 | ★ Murder investigation | Kill an NPC → guards investigate → witnesses questioned → bounty issued | New |
| 103 | ★ Thieves Guild / Criminal faction | Join criminal organizations for exclusive quests, fencing services, and safe houses | New |
| 104 | ★ Contraband system | Certain items illegal in certain factions; guards search suspicious players | New |
| 105 | ★ Vigilante system | Players can hunt down criminals (other players or NPCs) for bounty rewards | New |

### 11.10 Vehicles & Mounts

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 106 | Vehicle archetypes | 3+ vehicle types per genre (horse/cart/ship, hover-bike/shuttle/mech, etc.) | (V) Venture `procgen/vehicle` |
| 107 | ★ Vehicle physics | Steering, acceleration, braking, fuel/charge, damage; terrain affects handling | New |
| 108 | ★ First-person cockpit | Drive/ride from first-person with dashboard/reins/handlebars visible | New |
| 109 | ★ Vehicle combat | Attack from vehicles; vehicles take damage and can be destroyed | New |
| 110 | ★ Vehicle customization | Upgrade engine, armor, weapons; paint/decal customization | New |
| 111 | ★ Mount system | Tame / purchase mounts; mounts have stats (speed, stamina, loyalty) | New |
| 112 | ★ Naval / water vehicles | Boats, ships, submarines — water traversal with its own physics | New |
| 113 | ★ Flying vehicles | Genre-specific (griffins, jetpacks, gyrocopters) — 3D movement over terrain | New |
| 114 | ★ Vehicle storage | Vehicles have inventory space; use as mobile base | New |
| 115 | ★ Public transport | NPC-operated carriages / buses / shuttles between cities (genre-appropriate) | New |

### 11.11 Player Housing & Property

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 116 | ★ Purchasable houses | Buy houses in cities; instanced interiors; multiple properties allowed | New |
| 117 | ★ Furniture placement | First-person drag-and-drop furniture placement; genre-specific furniture sets | (V) Venture `procgen/furniture` extended |
| 118 | ★ Home upgrades | Expand rooms, add crafting stations, hire servants, build stables/garages | New |
| 119 | ★ Player shop ownership | Buy commercial property; stock with crafted/gathered goods; hire NPC shopkeeper | New |
| 120 | ★ Guild halls | Guilds pool resources to buy/build large properties with shared storage and amenities | New |
| 121 | ★ Trophy display | Mount trophies from boss kills, rare items, and achievements in your home | New |
| 122 | ★ Home invasion defense | Occasionally defend your property from thieves/raiders; better security = fewer attacks | New |
| 123 | ★ Tenant income | Own rental properties; NPCs pay rent providing passive income | New |

### 11.12 Dungeons & Instanced Content

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 124 | ★ BSP dungeon generation | Procedural dungeons with room graphs, corridors, traps, puzzles, boss rooms | New (seed-derived from Venture `procgen/terrain` BSP algorithm) |
| 125 | Puzzle rooms | Lever puzzles, pressure plates, pattern locks, riddles — genre-themed | (V) Venture `procgen/puzzle` adapted |
| 126 | ★ Dungeon ecology | Dungeon inhabitants have patrol routes, sleep schedules, and inter-species conflicts | New |
| 127 | ★ Environmental traps | Spike pits, flame jets, poison gas, laser grids — genre-appropriate | (V) Violence `pkg/trap` extended |
| 128 | ★ Secret rooms | Hidden passages revealed by perception skill or environmental clues | New |
| 129 | ★ Dungeon difficulty tiers | Easy/Medium/Hard/Legendary — affecting enemy stats, loot quality, and trap density | New |
| 130 | ★ Raid dungeons | Multi-party instanced content for 8–16 players with complex boss mechanics | New |
| 131 | ★ Dungeon keys & progression | Some areas locked behind keys/items found elsewhere in the dungeon | New |
| 132 | ★ Dynamic dungeon events | Random events during dungeon runs (cave-in, flooding, reinforcements) | New |
| 133 | ★ Dungeon loot rooms | Treasure vaults guarded by elite enemies or complex puzzles | New |

### 11.13 Weather, Environment & Atmosphere

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 134 | ★ Real-time weather | Rain, snow, fog, sandstorms, thunderstorms — rendered via particle overlay + post-processing | New |
| 135 | ★ Weather gameplay effects | Rain slows fire damage; sandstorms reduce visibility; cold drains stamina | New |
| 136 | ★ Time-of-day lighting | Dawn/day/dusk/night cycle affecting color grading, shadow direction, and NPC behavior | New |
| 137 | ★ Genre post-processing | Fantasy=warm grade; sci-fi=scanlines; horror=vignette+desaturate; cyberpunk=chromatic aberration; post-apoc=sepia+grain | New |
| 138 | ★ Procedural skybox | Sun/moon/stars position calculated from world clock; genre-appropriate sky palette | New |
| 139 | ★ Location-based ambience | Cave drip, city crowd, wind on plains, forest birdsong — synthesized per environment type | (V) Violence `pkg/audio/ambient` extended |
| 140 | ★ Destructible environment | Doors can be smashed; walls can be breached; trees can be felled (with appropriate tools/abilities) | (V) Violence `pkg/destruct` extended |

### 11.14 Audio & Music

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 141 | Procedural SFX | All sound effects synthesized from oscillators + envelopes; genre modifications | (V) Violence `pkg/audio/sfx` + Venture patterns |
| 142 | ★ Adaptive music | Exploration/combat/danger/town themes with seamless crossfading; per-faction motifs | New |
| 143 | ★ Genre music styles | Fantasy=orchestral; sci-fi=synth; horror=dissonant; cyberpunk=EDM; post-apoc=folk | New |
| 144 | ★ 3D spatial audio | Sound sources have world position; volume/pan calculated from player distance/direction | New |
| 145 | ★ Ambient soundscapes | Layer synthesized ambience (wind, water, crowd) based on environment type and weather | (V) Violence `pkg/audio/ambient` extended |
| 146 | ★ Reverb by environment | Cave = heavy reverb; outdoors = minimal; cathedral = long decay | (V) Violence `pkg/audio/reverb` extended |
| 147 | ★ Dynamic music intensity | Combat music intensifies with enemy count/threat level; exploration music responds to scenery | New |

### 11.15 Multiplayer & Networking

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 148 | Client-server authority | Server authoritative for all world state; clients send input commands only | (V) Violence `pkg/network/gameserver` pattern |
| 149 | Client-side prediction | Client predicts movement/actions locally; reconciles on server state arrival | (V) Violence `pkg/network/lagcomp` adapted |
| 150 | Delta compression | Only changed entity state transmitted; chunk diffs instead of full snapshots | (V) Violence `pkg/network/delta` |
| 151 | ★ Persistent world server | World state persists across server restarts; chunk state, NPC state, economy all serialized | New |
| 152 | ★ 200–5000ms latency tolerance | Playable even on Tor; prediction windows, jitter buffers, graceful degradation | (V) Violence `pkg/network/latency` extended |
| 153 | ★ Chunk streaming protocol | 3×3 active window; full snapshot on enter, delta on update, stop on exit | New |
| 154 | ★ Federation / cross-server travel | Players travel between server instances; redirect token carries ECS state | New |
| 155 | ★ Economy sync | Price gossip protocol across federation peers | New |
| 156 | ★ PvP zones | Opt-in PvP regions with loot-on-death rules; respawn points | New |
| 157 | ★ Player trading | Direct player-to-player item/gold exchange with anti-exploit validation | New |
| 158 | ★ Chat system | Local, faction, party, and global chat channels; profanity filter | New |
| 159 | ★ Party system | Form parties for shared quest progress, loot distribution, and instanced content | New |
| 160 | ★ Guild system | Create/join guilds with ranks, permissions, shared bank, territory, and guild quests | New |
| 161 | Anti-cheat | Server-side validation of all client inputs; rate limiting; sanity checks | (V) Violence `pkg/network/anticheat` extended |

### 11.16 UI & HUD

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 162 | ★ First-person HUD | Health/mana/stamina bars, minimap, compass, quest tracker, status effects — all procedurally rendered | New |
| 163 | ★ Inventory screen | Grid or list view with item icons (procedurally drawn), weight display, equipment comparison | New |
| 164 | ★ Character sheet | Stats, skills, perks, reputation, active effects — full RPG character overview | New |
| 165 | ★ World map | Reveal-on-explore map with settlement icons, fast-travel markers, quest objectives | New |
| 166 | ★ Dialog interface | NPC portrait (procedurally rendered), dialog text, response choices, skill check indicators | New |
| 167 | ★ Crafting interface | Material list, recipe browser, quality preview, workbench minigame | New |
| 168 | ★ Quest journal | Active/completed/failed quest log with descriptions, objectives, and map markers | New |
| 169 | ★ Settings menu | Graphics, audio, controls, network — configurable without config file editing | New |
| 170 | ★ Loading screens | Genre-themed procedural art + tip text during chunk loading (if needed) | New |

### 11.17 Persistence & World State

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 171 | ★ Chunk state persistence | Modified chunks (placed/destroyed objects, looted containers) survive server restarts | New |
| 172 | ★ NPC state persistence | NPC alive/dead, disposition, schedule modifications, inventory changes all persisted | New |
| 173 | ★ Quest flag persistence | All quest consequence flags serialized; branching state preserved | New |
| 174 | ★ Economy persistence | Price history, supply/demand state, player shop inventories all serialized | New |
| 175 | ★ Faction state persistence | Territory boundaries, war/peace status, treasury levels, player reputation | New |
| 176 | ★ Player data persistence | Inventory, skills, quest progress, housing, reputation — all server-side | New |
| 177 | ★ World event history | Log of significant events (wars, plagues, boss kills) for environmental storytelling | New |
| 178 | ★ Hot reload | Server restarts restore world state with <5% diff from pre-restart snapshot | New |

### 11.18 Modding & Extensibility

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 179 | ★ Generator override system | Server operators can replace any procedural generator with custom implementations | New |
| 180 | ★ Custom quest injection | Define new quest templates that integrate with the procedural quest system | New |
| 181 | ★ Custom faction definitions | Add new factions with custom relationship graphs and territory seeds | New |
| 182 | ★ Genre definition extension | Create new genres beyond the base 5 with custom palettes, names, and parameters | (V) Venture `procgen/genre` blending extended |
| 183 | ★ Server plugin API | Go plugin interface for custom systems, components, and event handlers | New |

### 11.19 Accessibility & Quality of Life

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 184 | ★ Configurable difficulty | Separate sliders for combat, exploration, economy, and puzzle difficulty | New |
| 185 | ★ Colorblind modes | Palette adjustments for protanopia, deuteranopia, and tritanopia | New |
| 186 | ★ Subtitle system | All dialog and important audio cues displayed as subtitles | New |
| 187 | ★ Key rebinding | Full input rebinding for keyboard and gamepad | New |
| 188 | ★ FOV slider | Adjustable field of view for first-person camera (60°–120°) | New |
| 189 | ★ Auto-save | Periodic server-side state snapshots; manual save points at beds/save stations | New |

### 11.20 Performance & Technical

| # | Feature | Description | V-Series Source |
|---|---------|-------------|-----------------|
| 190 | 60 FPS target | Client renders at 60 FPS at 1280×720 on mid-range hardware | (V) Suite-wide target |
| 191 | 20 Hz server tick | Server updates at 20 ticks/second with all systems | New |
| 192 | <500MB client RAM | Memory budget for single-binary client | (V) Suite-wide target |
| 193 | Single binary | No external assets; `go build ./cmd/client` produces self-contained binary | (V) Suite-wide philosophy |
| 194 | ★ LOD system | Distant entities rendered with reduced detail; far terrain simplified | New |
| 195 | ★ Object pooling | Pooled allocation for particles, projectiles, and ephemeral entities | New |
| 196 | ★ Spatial indexing | Spatial hash for entity queries, collision, and interest management | (V) Violence `pkg/combat/spatial_hash` |
| 197 | ★ Texture caching | LRU cache for generated textures; never regenerate same texture twice | (V) Vania `internal/pcg/cache` pattern |
| 198 | ★ Chunk pre-fetching | Background goroutines generate chunks ahead of player movement | New |
| 199 | WASM support | Client builds to WebAssembly for browser play | (V) Suite-wide (Ebiten) |
| 200 | Cross-platform | Linux, macOS, Windows, WASM from single codebase | (V) Suite-wide (Ebiten) |
