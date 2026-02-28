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

5. First-person raycaster in `pkg/rendering/raycast/` using Ebiten; procedural wall/floor/ceiling textures from `pkg/rendering/texture/` (no image files). **AC:** 60 fps at 1280×720 on reference hardware; genre changes wall palette.
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
| Tor-mode | Functional at 2000 ms RTT | Netem 2000 ms test: movement and combat resolve correctly |
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
