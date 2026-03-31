# Wyrm

**A 100% procedurally generated first-person open-world RPG built in Go on [Ebitengine](https://ebitengine.org/).**

Inspired by *Elder Scrolls* (open-world exploration, NPC schedules, faction politics), *Fallout* (post-apocalyptic tone, skill trees, dialogue consequences), and *GTA* (freeform crime/law systems, vehicles, persistent city life), Wyrm generates every element at runtime from a deterministic seed — no image files, no audio files, no level data. The game compiles to a single binary that runs anywhere without external assets.

Five genre themes reshape every player-facing system, making each playthrough a distinct RPG experience:

| Genre | World Theme | Factions | Vehicles | Visual Palette |
|-------|-------------|----------|----------|----------------|
| **Fantasy** | Magic-infused medieval continent | Kingdoms, guilds, church orders | Horse, war cart, sailing ship | Warm gold/green |
| **Sci-Fi** | Colony planet / space station | Corporations, military, cults | Hover-bike, shuttle, exo-mech | Cool blue/white |
| **Horror** | Cursed island / haunted city | Cults, survivor bands, monsters | Bone cart, plague barge, hearse | Desaturated grey-green |
| **Cyberpunk** | Megacity sprawl 2140 | Megacorps, street gangs, hackers | Motorbike, APC, aerial drone | Neon pink/cyan |
| **Post-Apocalyptic** | Irradiated wasteland | Tribes, raider clans, trader caravans | Dune buggy, armored bus, gyrocopter | Sepia/orange dust |

---

## Key Features

Wyrm targets **200 features** across 20 categories (see [ROADMAP.md](ROADMAP.md) for the full list). Highlights include:

### Open World & Exploration
- Seamless infinite terrain via 512×512 chunk streaming
- Multi-biome open world with vertical terrain (hills, cliffs, caves)
- Day/night cycle, real-time weather, and seasonal changes
- Procedural road networks, points of interest, and discoverable landmarks

### Cities, NPCs & Social Simulation
- Procedural multi-district cities with dynamic shop hours and law enforcement
- NPCs with full 24-hour daily schedules (sleep, work, eat, socialize, patrol)
- NPC memory, relationships, gossip networks, and emotional states
- Deep dialog trees with persuasion, intimidation, consequences, and genre-appropriate vocabulary

### Quests & Narrative
- Branching quest lines with persistent world-changing consequences
- Faction-specific multi-quest story arcs with mutual exclusivity
- Dynamic quest generation responding to world state (famine → food quest, war → spy quest)
- Radiant (infinitely generated) side quests from notice boards

### Combat, Skills & Progression
- First-person melee, ranged, and magic combat with timing-based blocking
- 30+ skills across 6 genre-renamed schools; skills improve through use (Elder Scrolls-style)
- Stealth system with sneak, pickpocket, and backstab mechanics
- Multi-phase boss encounters with unique mechanics

### Factions, Crime & Law
- Dynamic faction territory control with wars, diplomacy, and coups
- Crime detection via NPC line-of-sight witnesses; wanted level 0–5 stars
- Bounty system, jail mechanic, and criminal faction questlines
- Player-joinable factions with rank progression and exclusive content

### Economy, Crafting & Property
- Dynamic supply/demand economy with player-owned shops and trade routes
- Crafting via material gathering, workbench minigames, and recipe discovery
- Purchasable houses with first-person furniture placement
- Guild halls with shared storage and territory claims

### Vehicles & Mounts
- 3+ vehicle archetypes per genre with first-person cockpit view
- Vehicle physics (steering, acceleration, fuel), combat, and customization
- Mount system, naval vehicles, and flying vehicles

### Audio & Atmosphere
- All sound effects synthesized procedurally (oscillators + ADSR envelopes)
- Adaptive music with genre styles (orchestral, synth, dissonant, EDM, folk)
- 3D spatial audio and environment-based reverb
- Genre-specific post-processing (warm color grade, scanlines, vignette, chromatic aberration, sepia grain)

### Multiplayer & Networking
- Authoritative server with client-side prediction and delta compression
- **200–5000 ms latency tolerance** (designed for Tor-routed connections)
- Persistent world state surviving server restarts
- Cross-server federation, PvP zones, party/guild systems, and player trading

### Accessibility & Technical
- Single binary, zero external assets, cross-platform (Linux, macOS, Windows, WASM)
- 60 FPS at 1280×720; 20 Hz server tick; <500 MB client RAM
- Configurable difficulty, colorblind modes, subtitle system, full key rebinding

---

## Architecture

Wyrm uses an **Entity-Component-System (ECS)** architecture with an authoritative client-server model.

- **Entities** are `uint64` IDs
- **Components** (`Position`, `Health`, `Faction`, `Schedule`, `Inventory`, `Vehicle`) are pure data structs — no logic
- **Systems** (`WorldChunkSystem`, `NPCScheduleSystem`, `FactionPoliticsSystem`, `CrimeSystem`, `EconomySystem`, `CombatSystem`, `VehicleSystem`, `QuestSystem`, `WeatherSystem`, `RenderSystem`, `AudioSystem`) contain all game logic and operate on component queries each tick

### V-Series Reuse

Wyrm is part of the **opd-ai Procedural Game Suite** — 8 sibling repositories sharing a zero-external-assets philosophy:

| Repo | Genre | Description |
|------|-------|-------------|
| [opd-ai/venture](https://github.com/opd-ai/venture) | Co-op action-RPG | Top-down roguelike with 25+ `pkg/procgen/` generators |
| [opd-ai/violence](https://github.com/opd-ai/violence) | Raycasting FPS | First-person shooter with rendering, combat, and networking |
| [opd-ai/velocity](https://github.com/opd-ai/velocity) | Galaga-like shooter | Spawner, wave manager, and balance patterns |
| [opd-ai/vania](https://github.com/opd-ai/vania) | Metroidvania | Seed mixing, caching, and validation reference implementations |
| [opd-ai/way](https://github.com/opd-ai/way) | Battle-cart racer | Vehicular combat racing |
| [opd-ai/where](https://github.com/opd-ai/where) | Wilderness survival | Survival crafting game |
| [opd-ai/whack](https://github.com/opd-ai/whack) | Arena battle | Melee combat arena |

Wyrm imports and extends generators from **Venture** (terrain, entity, faction, quest, dialog, narrative, building, vehicle, magic, skills, and more) and rendering/networking infrastructure from **Violence** (raycaster, lag compensation, delta compression, spatial hashing). See [ROADMAP.md § 9](ROADMAP.md) for the complete V-Series reuse guide.

---

## Implementation Phases

Development follows a 6-phase plan (see [ROADMAP.md § 3](ROADMAP.md) for full details):

| Phase | Focus | Duration |
|-------|-------|----------|
| **1 — Foundation** | ECS core, Go module scaffold, V-Series integration, headless server | 8 weeks |
| **2 — Open World & Rendering** | Chunk streaming, raycaster, procedural textures, NPC schedules | 10 weeks |
| **3 — Gameplay Systems** | Combat, skills, factions, quests, economy, crime/law, vehicles | 10 weeks |
| **4 — Audio & Visual Polish** | Procedural audio, adaptive music, post-processing, genre effects | 6 weeks |
| **5 — Multiplayer & Persistence** | Shared world, PvP, housing, guilds, economy sync, Tor tolerance | 8 weeks |
| **6 — Content Depth & Release** | Dungeons, deep dialog, companion AI, genre playthrough validation | 6 weeks |

---

## Directory Structure

```
cmd/client/              Client entry point (Ebitengine window)
cmd/server/              Authoritative game server entry point
config/                  Configuration loading (Viper)
pkg/engine/ecs/          Entity-Component-System core
pkg/engine/components/   ECS component definitions
pkg/engine/systems/      ECS system implementations
pkg/world/chunk/         World chunk management and streaming
pkg/rendering/raycast/   First-person raycasting renderer
pkg/rendering/texture/   Procedural texture generation
pkg/procgen/city/        Procedural city generation
pkg/audio/               Procedural audio synthesis
pkg/network/             Client-server networking
```

## Build

```bash
go build ./cmd/client
go build ./cmd/server
```

## Test

```bash
# Run all tests (requires X11 display or xvfb for Ebitengine packages)
xvfb-run -a go test -race ./...

# Run tests for headless packages (no display required)
go test -tags=noebiten ./pkg/procgen/adapters/...
go test -tags=noebiten ./pkg/rendering/raycast/...
go test -tags=noebiten ./cmd/client/...
go test -tags=noebiten ./cmd/server/...
```

The `noebiten` build tag enables testing of packages that have Ebitengine dependencies without requiring a graphical display. This is useful for CI environments and server deployments.

## Run

```bash
./server   # starts authoritative server on localhost:7777
./client   # launches game window and connects to server
```

## Configuration

Configuration is loaded from `config.yaml` in the working directory or `./config/` directory. Environment variables with the `WYRM_` prefix override config file values (e.g., `WYRM_WORLD_SEED=12345`, `WYRM_GENRE=cyberpunk`). Defaults are used when no config file is present.

```yaml
window:
  width: 1280
  height: 720
  title: "Wyrm"

server:
  address: "localhost:7777"
  protocol: "tcp"
  tick_rate: 20

world:
  seed: 0           # 0 = random seed on startup
  chunk_size: 512

genre: "fantasy"    # fantasy, sci-fi, horror, cyberpunk, post-apocalyptic
```

See [`config.yaml`](config.yaml) for all available settings.

## Dependencies

- [Ebitengine v2](https://ebitengine.org/) — 2D game engine with cross-platform and WASM support
- [Viper](https://github.com/spf13/viper) — configuration management (YAML, environment variables)

## License

See [LICENSE](LICENSE) for details.
