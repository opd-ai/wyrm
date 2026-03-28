# Project Overview

Wyrm is a 100% procedurally generated first-person open-world RPG built in Go 1.24+ on Ebitengine v2. Inspired by Elder Scrolls (open-world exploration, NPC schedules, faction politics), Fallout (post-apocalyptic tone, skill trees, dialogue consequences), and GTA (freeform crime/law systems, vehicles, persistent city life), Wyrm is the most ambitious W-Series title. The game extends the V-Series generators (terrain, entity, faction, quest, dialog, narrative, building, vehicle, magic, skills) from top-down roguelike to a persistent first-person open world.

Every element is generated at runtime from a deterministic seed: no image files, no audio files, no level data. The game compiles to a single binary that runs anywhere without external assets. Five genre themes (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic) reshape every player-facing system—from world aesthetics to NPC culture—making each playthrough a distinct RPG experience. Multiplayer networking is designed for 200–5000ms latency tolerance, supporting diverse network conditions including Tor-routed connections.

## Sibling Repository Context

Wyrm is part of the opd-ai Procedural Game Suite—8 sibling repositories that all share a 100% procedural, zero-external-assets philosophy. Code patterns, conventions, and eventually shared library packages flow between these repos:

| Repo | Genre | Description |
|------|-------|-------------|
| `opd-ai/venture` | Co-op action-RPG | Top-down roguelike with extensive pkg/procgen/ generators |
| `opd-ai/vania` | Metroidvania | Procedural 2D platformer |
| `opd-ai/velocity` | Galaga-like shooter | Fast-paced arcade shmup |
| `opd-ai/violence` | Raycasting FPS | First-person shooter with libp2p networking |
| `opd-ai/way` | Battle-cart racer | Vehicular combat racing |
| `opd-ai/wyrm` | First-person survival RPG | Open-world Elder Scrolls-style RPG (this repo) |
| `opd-ai/where` | Wilderness survival | Survival crafting game |
| `opd-ai/whack` | Arena battle | Melee combat arena |

When implementing features, match patterns from sibling repos (especially Venture for procgen, Violence for networking) to enable future code sharing.

## Technical Stack

- **Primary Language**: Go 1.24.0
- **Game Framework**: Ebitengine v2.8.8 — 2D game engine with cross-platform + WASM support
- **Configuration**: Viper v1.19.0 — YAML config with environment variable override
- **Architecture**: Entity-Component-System (ECS) with authoritative server model
- **Testing**: Go standard `testing` package, table-driven tests, benchmarks
- **Build/Deploy**: `go build ./cmd/client` and `go build ./cmd/server`

### Key Dependencies (from go.mod)

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/hajimehoshi/ebiten/v2` | v2.8.8 | 2D game engine, rendering, input, audio |
| `github.com/spf13/viper` | v1.19.0 | Configuration management (YAML, env vars) |
| `golang.org/x/sync` | v0.8.0 | Extended synchronization primitives |
| `golang.org/x/text` | v0.18.0 | Unicode text processing |

## Project Structure

Wyrm uses a **venture-style layout** with separate client and server entrypoints plus a `pkg/` directory for public library packages:

```
cmd/client/              # Ebitengine client entrypoint
cmd/server/              # Authoritative game server entrypoint
config/                  # Viper configuration loading (config.go)
config.yaml              # Default configuration file
pkg/
├── engine/
│   ├── ecs/             # Entity-Component-System core (World, Entity, Component, System interfaces)
│   ├── components/      # ECS component definitions (Position, Health, Faction, Schedule, Inventory, Vehicle)
│   └── systems/         # ECS system implementations (WorldChunk, NPCSchedule, Combat, Weather, Render, Audio, etc.)
├── world/
│   └── chunk/           # World chunk management and streaming
├── rendering/
│   ├── raycast/         # First-person raycasting renderer
│   └── texture/         # Procedural texture generation
├── procgen/
│   └── city/            # Procedural city generation
├── audio/               # Procedural audio synthesis
└── network/             # Client-server networking (Server, Client types)
```

---

## ⚠️ CRITICAL: Complete Feature Integration (Zero Dangling Features)

**This is the single most important rule for this codebase.** Every feature, system, component, generator, and integration MUST be fully wired into the runtime. Dangling features are a maintenance burden, a source of frustration, and actively degrade code quality.

### The Dangling Feature Problem

Wyrm is an ambitious open-world RPG with many interconnected systems. The codebase currently has **significant integration gaps** that must be addressed as development continues. Many systems are defined but not yet wired into the game loop.

**Current Integration Status (as of codebase analysis):**

| System | Defined | Instantiated | Registered | Update Called | Produces Output | Output Consumed |
|--------|---------|--------------|------------|---------------|-----------------|-----------------|
| WorldChunkSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| NPCScheduleSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| FactionPoliticsSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| CrimeSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| EconomySystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| CombatSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| VehicleSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| QuestSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| WeatherSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| RenderSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| AudioSystem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

**All systems in `pkg/engine/systems/systems.go` have empty `Update()` method bodies.** This is expected for early development but represents a critical integration gap that must be addressed.

### Mandatory Checks Before Adding or Modifying Any Feature

**Before writing ANY new code, verify the full integration chain:**

1. **Definition → Instantiation**: Is the struct/system created at runtime? Trace from `main()` through system registration.
2. **Instantiation → Registration**: Is the system registered with `world.RegisterSystem()`? Check both client and server `main.go`.
3. **Registration → Update Loop**: Does the system's `Update()` method actually get called each frame/tick via `world.Update(dt)`?
4. **Update → Output**: Does the system produce outputs (component modifications, events, state changes) that other systems consume?
5. **Output → Consumer**: Is there at least one other system that reads this system's output?
6. **Consumer → Player Effect**: Does the chain ultimately produce something visible, audible, or mechanically felt by the player?

If ANY link in this chain is missing, the feature is dangling. **Do not submit dangling features.**

### Specific Anti-Patterns to Reject

```go
// ❌ BAD: System defined but never added to the game world
// Currently ALL systems in pkg/engine/systems/ fall into this category!
type WeatherSystem struct{}
func (s *WeatherSystem) Update(w *ecs.World, dt float64) {}
// ...but WeatherSystem{} is never created or registered in cmd/client/main.go or cmd/server/main.go

// ✅ GOOD: System defined, instantiated, registered, and consuming/producing
weather := &systems.WeatherSystem{Seed: cfg.World.Seed}
world.RegisterSystem(weather)
// AND the RenderSystem applies weather effects:
// func (r *RenderSystem) Update(w *ecs.World, dt float64) {
//     weather := getWeatherState(w)
//     r.ApplyWeatherVisuals(weather)
// }
```

```go
// ❌ BAD: Generator exists but is never called in runtime (only tests)
// pkg/procgen/city/city.go defines Generate() but it's never called from cmd/
func Generate(seed int64, genre string) *City { ... }

// ✅ GOOD: Generator is called during world initialization
func initializeWorld(cfg *config.Config) *ecs.World {
    world := ecs.NewWorld()
    city := city.Generate(cfg.World.Seed, cfg.Genre)
    spawnCityEntities(world, city)
    return world
}
```

```go
// ❌ BAD: Component defined but no system operates on it
type Schedule struct {
    CurrentActivity string
    TimeSlots       map[int]string
}
// Schedule is defined but NPCScheduleSystem.Update() is empty

// ✅ GOOD: Component has a system that reads and modifies it
func (s *NPCScheduleSystem) Update(w *ecs.World, dt float64) {
    for _, e := range w.Entities("Schedule", "Position") {
        sched, _ := w.GetComponent(e, "Schedule")
        schedule := sched.(*components.Schedule)
        hour := s.worldClock.Hour()
        activity := schedule.TimeSlots[hour]
        if activity != schedule.CurrentActivity {
            schedule.CurrentActivity = activity
            s.moveNPCToActivityLocation(w, e, activity)
        }
    }
}
```

### Integration Verification Checklist (run before every PR)

```bash
# Every constructor should have at least one non-test caller
grep -rn 'func New' --include='*.go' | grep -v _test.go

# All TODOs should be tracked
grep -rn 'TODO\|FIXME\|HACK\|XXX' --include='*.go'

# Find empty method bodies (potential stub implementations)
grep -rn '{}$' --include='*.go' | grep 'func'

# Verify systems are registered (should see RegisterSystem calls)
grep -rn 'RegisterSystem' --include='*.go'

# Verify generators are called at runtime
grep -rn 'Generate(' --include='*.go' | grep -v _test.go | grep -v 'func.*Generate'
```

**Current gaps to address:**
- No systems registered in either `cmd/client/main.go` or `cmd/server/main.go`
- ChunkManager created but unused (assigned to `_` in server)
- City generator never called outside its package
- Audio engine never instantiated

---

## Networking Best Practices (MANDATORY for all Go network code)

### Interface-Only Network Types (Hard Constraint)

When declaring network variables, ALWAYS use interface types. This is a **non-negotiable project rule**.

| ❌ Never Use (Concrete Type) | ✅ Always Use (Interface Type) |
|------------------------------|-------------------------------|
| `*net.UDPAddr` | `net.Addr` |
| `*net.IPAddr` | `net.Addr` |
| `*net.TCPAddr` | `net.Addr` |
| `*net.UDPConn` | `net.PacketConn` |
| `*net.TCPConn` | `net.Conn` |
| `*net.TCPListener` | `net.Listener` |

The current `pkg/network/network.go` correctly uses interface types:
```go
// ✅ GOOD: Current implementation uses interfaces
type Server struct {
    listener net.Listener  // Interface, not *net.TCPListener
}

type Client struct {
    conn net.Conn  // Interface, not *net.TCPConn
}
```

**Never use type switches or type assertions to convert from interface to concrete type:**

```go
// ❌ BAD: Type assertion to access concrete methods
if tcpConn, ok := conn.(*net.TCPConn); ok {
    tcpConn.SetNoDelay(true)
}

// ✅ GOOD: Use interface methods or TCP-agnostic approaches
conn.SetDeadline(time.Now().Add(timeout))
```

### High-Latency Network Design (200–5000ms)

Wyrm MUST function correctly under **200–5000ms round-trip latency** as specified in ROADMAP.md. This is non-negotiable for supporting Tor-routed connections and diverse global network conditions.

#### Mandatory Design Principles

1. **Client-Side Prediction**: The client must simulate game state locally and reconcile with server authoritative state when it arrives. Never block the game loop waiting for a server response.

2. **State Interpolation / Extrapolation**: Remote entity positions must be interpolated between known states. When packets are delayed beyond the interpolation window, extrapolate using last-known velocity.

3. **Jitter Buffers**: Incoming state updates must be buffered and played back at a consistent rate, absorbing latency variance (jitter). Design for ±500ms jitter tolerance minimum.

4. **Idempotent Messages**: Every network message must be safe to process multiple times. Retransmission at high latency is expected, not exceptional.

5. **No Synchronous RPC in Game Loops**: Never issue a blocking network call inside `Update()` or `Draw()`. All network I/O must be asynchronous with results consumed on the next available frame.

6. **Graceful Degradation**: At 5000ms latency the game must remain playable, not just connected. Reduce update frequency, increase prediction windows, and hide latency with animations.

7. **Timeout Tolerance**: Connection timeouts must be set to ≥10 seconds. Disconnect detection must use heartbeat absence over a sliding window (≥3 missed heartbeats), never a single missed packet.

```go
// ❌ BAD: Tight timeout that drops players on satellite connections
conn.SetReadDeadline(time.Now().Add(1 * time.Second))

// ✅ GOOD: Generous timeout for high-latency environments
conn.SetReadDeadline(time.Now().Add(10 * time.Second))

// ❌ BAD: Blocking network call in game loop
func (g *Game) Update() error {
    state, err := g.server.GetWorldState()  // blocks until response
    g.world = state
    return nil
}

// ✅ GOOD: Async receive with interpolation
func (g *Game) Update() error {
    select {
    case state := <-g.stateChannel:
        g.interpolator.PushServerState(state)
    default:
        // No new state — continue with prediction
    }
    g.world = g.interpolator.GetInterpolatedState(time.Now())
    return nil
}
```

#### Tor-Mode (RTT > 800ms)

Per ROADMAP.md, when RTT exceeds 800ms:
- Increase client prediction window to 1500ms
- Reduce input send rate to 10 Hz
- Enable aggressive visual interpolation with 300ms blend time

#### Lag Compensation

Server maintains 500ms entity position history ring buffer. Hit registration rewinds to client-tagged timestamp (clamped to history window), checks overlap, then re-advances.

---

## Code Assistance Guidelines

### 1. Deterministic Procedural Generation

All content generation MUST be deterministic and seed-based. Given the same seed, the game MUST produce identical output across all platforms and runs.

```go
// ✅ GOOD: Explicit seed-based RNG, never global
rng := rand.New(rand.NewSource(seed))
value := rng.Intn(100)

// ❌ BAD: Global rand (non-deterministic, not thread-safe)
value := rand.Intn(100)

// ❌ BAD: Time-based seeding in generation code
rng := rand.New(rand.NewSource(time.Now().UnixNano()))

// ✅ GOOD: Derived seeds for sub-generators (deterministic hierarchy)
// Use stable mixing, not simple XOR (per ROADMAP.md section 5)
terrainSeed := mixSeeds(seed, "terrain")
enemySeed := mixSeeds(seed, "enemy")
terrainRNG := rand.New(rand.NewSource(terrainSeed))
enemyRNG := rand.New(rand.NewSource(enemySeed))

// Seed mixing function (example using FNV-1a)
func mixSeeds(seed int64, tag string) int64 {
    h := fnv.New64a()
    binary.Write(h, binary.LittleEndian, seed)
    h.Write([]byte(tag))
    return int64(h.Sum64())
}
```

The current `pkg/world/chunk/chunk.go` uses a simple linear seed derivation:
```go
c := NewChunk(x, y, cm.ChunkSize, cm.Seed+int64(x)*31+int64(y)*37)
```
This is acceptable for chunk coordinates but should use FNV or xxHash for quest/instance seeds.

### 2. ECS Architecture Discipline

Wyrm uses a clean ECS architecture defined in `pkg/engine/ecs/`:

**Core Types:**
- `Entity` = `uint64` ID
- `Component` = interface with `Type() string` method (pure data, NO logic)
- `System` = interface with `Update(w *World, dt float64)` method (ALL logic here)
- `World` = entity registry + component store + system runner

**Rules:**
- Components are pure data structs. NO methods beyond `Type()`.
- Systems contain ALL game logic. Systems operate on entity collections via `world.Entities(componentTypes...)`.
- Never store entity references directly; use `Entity` IDs.
- Systems declare their dependencies by the component types they query.

```go
// ✅ GOOD: Component is pure data
type Position struct {
    X, Y, Z float64
}
func (p *Position) Type() string { return "Position" }

// ✅ GOOD: System contains logic, operates on component queries
func (s *MovementSystem) Update(w *ecs.World, dt float64) {
    for _, e := range w.Entities("Position", "Velocity") {
        pos, _ := w.GetComponent(e, "Position")
        vel, _ := w.GetComponent(e, "Velocity")
        position := pos.(*components.Position)
        velocity := vel.(*components.Velocity)
        position.X += velocity.X * dt
        position.Y += velocity.Y * dt
        position.Z += velocity.Z * dt
    }
}

// ❌ BAD: Logic in component
type Position struct {
    X, Y, Z float64
}
func (p *Position) Move(dx, dy, dz float64) { // NO! Logic belongs in systems
    p.X += dx; p.Y += dy; p.Z += dz
}
```

### 3. Configuration via Viper

Configuration follows the established pattern in `config/config.go`:

```go
// Load configuration with Viper
cfg, err := config.Load()
if err != nil {
    log.Fatalf("config: %v", err)
}

// Access configuration
seed := cfg.World.Seed
genre := cfg.Genre
width := cfg.Window.Width
```

**Environment Variable Override:**
- Prefix: `WYRM_`
- Separator: `_` (dots become underscores)
- Example: `WYRM_WORLD_SEED=12345` overrides `world.seed`

**Supported Genres** (from ROADMAP.md):
- `fantasy` (default)
- `sci-fi`
- `horror`
- `cyberpunk`
- `post-apocalyptic`

### 4. Performance Requirements

- **Target**: 60 FPS at 1280×720 on mid-range hardware
- **Server tick rate**: 20 Hz (configurable via `server.tick_rate`)
- **Memory budget**: <500MB client RAM
- **Chunk streaming**: 3×3 chunk window, delta-compressed updates
- **NPC simulation**: 200 NPCs + 32 players in ≤20ms server tick

**Optimization Patterns:**
- Use spatial partitioning for entity queries over collections >100
- Cache generated textures—never regenerate the same texture twice per session
- Use object pooling for frequently allocated/deallocated objects
- Benchmark hot paths with `go test -bench=. -benchmem`

### 5. Zero External Assets

The single-binary philosophy means ALL content is generated at runtime:

- **Graphics**: Procedurally generated via `pkg/rendering/texture/` (pixel manipulation, noise functions)
- **Audio**: Synthesized via `pkg/audio/` (oscillators, envelopes, effects)
- **Levels/Maps**: Generated via `pkg/world/chunk/` and `pkg/procgen/city/`
- **NPCs/Items/Quests**: Will use Venture-style generators with genre parameters
- **UI**: Built from code, not loaded from image files

**Never add asset files** (PNG, WAV, OGG, JSON level files) to the repository. If you need test fixtures, generate them in test setup code.

### 6. Error Handling

```go
// ✅ GOOD: Return errors, handle them at the call site
func GenerateChunk(x, y int, seed int64) (*Chunk, error) {
    if seed == 0 {
        return nil, fmt.Errorf("chunk generation requires non-zero seed")
    }
    // ...
}

// ❌ BAD: Panic in library/game code
func GenerateChunk(x, y int, seed int64) *Chunk {
    if seed == 0 {
        panic("zero seed")  // Never panic in game logic
    }
}

// ✅ GOOD: Log and recover gracefully in game systems
func (s *WorldChunkSystem) Update(w *ecs.World, dt float64) {
    chunk, err := s.chunkManager.GetChunk(playerChunkX, playerChunkY)
    if err != nil {
        log.Printf("chunk load failed at (%d,%d): %v", playerChunkX, playerChunkY, err)
        chunk = s.fallbackChunk()
    }
}
```

Panics are acceptable ONLY in `main()` for unrecoverable startup failures. All game systems must handle errors gracefully with fallbacks.

---

## Cross-Repository Code Sharing Patterns

### Shared Pattern Catalog

When implementing features, follow these patterns so code can be extracted into shared packages later:

| Pattern | Wyrm Package | Sibling Reference |
|---------|--------------|-------------------|
| ECS core (World, Entity, Component, System) | `pkg/engine/ecs/` | Venture `pkg/engine/` |
| Components (Position, Health, etc.) | `pkg/engine/components/` | Venture `pkg/engine/components/` |
| Systems (Combat, Weather, etc.) | `pkg/engine/systems/` | Venture `pkg/engine/systems/` |
| Procedural generation framework | `pkg/procgen/` | Venture `pkg/procgen/` (30+ sub-packages) |
| Chunk/world management | `pkg/world/chunk/` | — |
| First-person raycasting | `pkg/rendering/raycast/` | Violence `pkg/rendering/` |
| Procedural textures | `pkg/rendering/texture/` | — |
| Audio synthesis | `pkg/audio/` | Venture `pkg/audio/` |
| Networking | `pkg/network/` | Violence `pkg/network/` |
| Configuration | `config/` | Velocity, Violence (viper) |

### Interface Signatures (Match Across Repos)

**Component Interface** (universal):
```go
type Component interface {
    Type() string
}
```

**System Interface** (Wyrm uses World pointer):
```go
type System interface {
    Update(w *World, dt float64)
}
```

**Generator Pattern** (for procgen packages):
```go
type Generator[T any] interface {
    Generate(seed int64, params GenerationParams) (T, error)
}

type GenerationParams struct {
    GenreID    string
    Difficulty float64
    Depth      int
    Custom     map[string]any
}
```

### When Adding a Feature That Exists in a Sibling Repo

1. Check the sibling repo's implementation first
2. Use the same package structure and naming conventions
3. Match the interface signatures so future extraction is seamless
4. If the sibling implementation has known issues (check its GAPS.md), fix them in your implementation
5. Document divergences in ROADMAP.md with a note about future convergence

---

## V-Series PCG Library Reuse Guide

Wyrm is designed to stand on the shoulders of the four V-Series games. Each sibling repo contributes battle-tested procedural generation, rendering, audio, and networking subsystems that Wyrm wraps in open-world adapters. This section specifies **exactly** what to import, what to adapt, and what to build new.

### Venture (`opd-ai/venture`) — Primary PCG Source

Venture is the single richest source of reusable generators. Its `pkg/procgen/` tree contains 25+ sub-packages covering nearly every content type an RPG needs.

| Venture Package | What It Provides | How Wyrm Reuses It |
|-----------------|------------------|--------------------|
| `pkg/procgen/terrain` | Multi-algorithm terrain (BSP, cellular, city, forest, composite, grammar, maze) | Wrap as chunk-level heightmap + biome generator. Replace Venture's fixed-room output with Wyrm's 512×512 open-world chunks by tiling composite terrain across chunk boundaries. |
| `pkg/procgen/entity` | NPC stat templates + procedural name grammar | Import directly for NPC creation. Extend with Wyrm's `Schedule` and `Reputation` components to give NPCs daily routines and per-player disposition. |
| `pkg/procgen/faction` | Faction graph generation (relations, territory Voronoi) | Import as-is for initial faction seeding. Layer Wyrm's `FactionPoliticsSystem` on top to drive dynamic war/treaty events and territory shifts at runtime. |
| `pkg/procgen/quest` | Template-graph quest generation with consequence flags | Import the template graph. Extend with Wyrm's persistent consequence storage (flags survive server restart) and cross-quest dependency tracking. |
| `pkg/procgen/dialog` | Topic-graph dialog with sentiment model | Import topic graph + sentiment. Add Wyrm's NPC memory system (topic recall across conversations) and emotional state modifiers. |
| `pkg/procgen/narrative` | Story arc templates + plot progression | Import arc templates. Extend with Wyrm's open-world narrative layering (multiple concurrent story arcs per player, faction-driven plot forks). |
| `pkg/procgen/building` | Room-template buildings + façade grammar | Import for city building interiors. Extend façade grammar with genre-specific exterior details (neon signs for cyberpunk, vine overgrowth for fantasy). |
| `pkg/procgen/vehicle` | Vehicle archetype params + genre skin | Import archetypes. Add Wyrm's `VehicleSystem` physics (steering, acceleration, fuel/charge) and first-person cockpit view integration. |
| `pkg/procgen/magic` | Spell component grammar | Import spell generation. Extend with Wyrm's cooldown system, mana/energy costs, and visual effect triggers for the first-person renderer. |
| `pkg/procgen/skills` | Skill tree generation across schools | Import skill definitions. Wire into Wyrm's XP-based progression system with genre-renamed schools (Destruction→Weaponry for sci-fi). |
| `pkg/procgen/class` | Character class archetypes | Import for initial player and NPC class assignment. Extend with Wyrm's multi-class hybrid system allowing cross-school skill picks. |
| `pkg/procgen/companion` | Companion personality + combat role + backstory | Import companion generation. Add Wyrm's companion AI (follow/fight/wait commands), persistent relationship score, and dialog references to shared adventures. |
| `pkg/procgen/recipe` | Item crafting recipe generation (affix table + material grammar) | Import recipe system. Extend with Wyrm's workbench interaction (first-person crafting UI), material gathering from world objects, and item quality tiers. |
| `pkg/procgen/station` | Station/shop generation | Import for city POI placement. Wrap in Wyrm's `EconomySystem` to give stations dynamic supply/demand pricing. |
| `pkg/procgen/story` | Story event generation | Import for world event seeds. Layer Wyrm's persistent event consequences (destroyed buildings stay destroyed, killed NPCs stay dead). |
| `pkg/procgen/book` | In-game book/lore text generation | Import for world-building. Place generated books in Wyrm's building interiors via the furniture system. |
| `pkg/procgen/furniture` | Interior furniture placement | Import for player housing and building interiors. Extend with Wyrm's player-placeable furniture (drag-and-drop in first-person). |
| `pkg/procgen/environment` | Environmental detail generation | Import for ambient world decoration. Feed into Wyrm's chunk decoration pass (trees, rocks, debris, genre-appropriate clutter). |
| `pkg/procgen/item` | Item generation (weapons, armor, consumables) | Import item templates. Add Wyrm's durability, enchantment, and repair systems. |
| `pkg/procgen/legendary` | Legendary/unique item generation | Import for rare loot. Wire into Wyrm's boss drop tables and quest reward pools. |
| `pkg/procgen/puzzle` | Puzzle room generation | Import for dungeon puzzle rooms. Adapt from top-down interaction to first-person interaction (lever pulling, switch flipping, object manipulation). |
| `pkg/procgen/minigame` | Minigame generation | Import templates. Adapt UI from top-down to first-person or overlay-screen presentation (lockpicking, hacking, persuasion). |
| `pkg/procgen/genre` | Genre registry + blending | Import genre definitions as the canonical genre source. Wyrm's `cfg.Genre` maps directly to Venture's genre registry. |
| `pkg/procgen/audit` | Generator output validation | Import audit framework. Run Wyrm's extended open-world validation checks (chunk connectivity, NPC schedule coverage, economy balance). |

**Import pattern:**
```go
import (
    vterrain  "github.com/opd-ai/venture/pkg/procgen/terrain"
    ventity   "github.com/opd-ai/venture/pkg/procgen/entity"
    vfaction  "github.com/opd-ai/venture/pkg/procgen/faction"
    vquest    "github.com/opd-ai/venture/pkg/procgen/quest"
    // ... etc
)

// Wrap Venture generators with Wyrm's open-world params
func generateChunkTerrain(seed int64, genre string, cx, cy int) (*ChunkTerrain, error) {
    chunkSeed := mixSeeds(seed, fmt.Sprintf("terrain:%d:%d", cx, cy))
    params := vterrain.GenerationParams{GenreID: genre}
    raw, err := vterrain.Generate(chunkSeed, params)
    if err != nil {
        return nil, fmt.Errorf("terrain gen at (%d,%d): %w", cx, cy, err)
    }
    return adaptToChunk(raw, cx, cy), nil
}
```

### Violence (`opd-ai/violence`) — Rendering, Networking, Combat

Violence is the FPS sibling and the primary source for first-person rendering infrastructure, real-time combat, and multiplayer networking.

| Violence Package | What It Provides | How Wyrm Reuses It |
|------------------|------------------|--------------------|
| `pkg/raycaster` | DDA raycasting engine with trig lookup tables, wall/floor/ceiling rendering | **Primary rendering engine.** Import and extend with Wyrm's procedural texture mapping, multi-height walls (buildings), and skybox rendering. Violence's raycaster is optimized for grid-based maps; Wyrm extends it to handle open terrain elevation. |
| `pkg/audio` | Procedural audio synthesis (SFX, ambient, reverb) with oscillators and ADSR envelopes | Import the synthesis core (oscillators, envelopes, reverb). Extend with Wyrm's genre-specific SFX modifications, adaptive music system, and 3D spatial audio for open-world environments. |
| `pkg/network` | Game server, lag compensation (500ms rewind), latency estimation, delta compression, anti-cheat, matchmaking, co-op, FFA, team modes, territory control | Import lag compensation (`lagcomp.go`), delta compression (`delta.go`), latency estimation (`latency.go`), and anti-cheat validation (`anticheat.go`). Wyrm replaces Violence's match-based session model with persistent-world connections, but the underlying packet handling and lag compensation algorithms transfer directly. |
| `pkg/combat` | Combat resolution (boss phases, combo system, defense mechanics, positional damage, spatial hash, telegraph system) | Import spatial hash for efficient collision queries and positional damage calculations. Adapt Violence's fast-paced FPS combat to Wyrm's RPG combat (add stat-based damage modifiers, skill cooldowns, and magic/ability integration). Reuse telegraph system for enemy attack warnings. |
| `pkg/collision` | Collision detection and response | Import for player-world and entity-entity collision. Extend with Wyrm's vehicle collision physics and NPC pathfinding obstacle avoidance. |
| `pkg/bsp` | BSP tree for level geometry | Import for dungeon interior rendering. Extend with Wyrm's building-interior BSP for seamless indoor/outdoor transitions. |
| `pkg/camera` / `pkg/camerafx` | First-person camera system with effects | Import camera system. Extend with Wyrm's dialogue camera (face NPC), vehicle camera (cockpit view), and cinematic camera (cutscene framing). |
| `pkg/weapon` / `pkg/weaponanim` / `pkg/weaponsway` | Weapon rendering, animation, and view sway | Import weapon viewmodel rendering. Extend with Wyrm's melee weapons, magic staves/wands, and genre-specific weapon types (laser rifles for sci-fi, cursed blades for horror). |
| `pkg/particle` | Particle system | Import for spell effects, explosions, environmental particles. Extend with Wyrm's weather particles (rain, snow, dust, ash). |
| `pkg/fog` / `pkg/lighting` | Fog and lighting systems | Import for atmospheric rendering. Extend with Wyrm's time-of-day lighting cycle and genre-specific atmosphere (perpetual fog for horror, neon glow for cyberpunk). |
| `pkg/texture` / `pkg/walltex` | Procedural texture generation | Import base texture generation. Extend with Wyrm's biome-aware textures (grass/stone/sand) and genre palette application. |
| `pkg/loot` / `pkg/inventory` / `pkg/equipment` | Loot tables, inventory management, equipment slots | Import inventory data structures and equipment slot system. Extend with Wyrm's weight-based carry capacity, equipment condition/durability, and paper-doll visualization. |
| `pkg/lore` / `pkg/dialogue` | Lore and dialog systems | Reference for patterns but prefer Venture's more mature `pkg/procgen/dialog` and `pkg/procgen/narrative` for generation. Use Violence's dialog UI rendering approach for first-person dialog presentation. |
| `pkg/ai` | Enemy AI (patrol, chase, attack) | Import patrol/chase state machine. Extend with Wyrm's NPC schedule-aware AI (NPCs transition from daily routine to combat stance and back). |
| `pkg/hazard` / `pkg/trap` | Environmental hazards and traps | Import for dungeon hazards. Extend with Wyrm's open-world hazards (radiation zones, magic anomalies, weather dangers). |
| `pkg/save` | Save/load system | Reference for persistence patterns. Wyrm uses server-side persistence rather than client-side saves, but the serialization approach is reusable. |
| `pkg/progression` / `pkg/skills` / `pkg/class` | Character progression systems | Reference as supplementary to Venture's generators. Violence's progression is FPS-focused (accuracy, reload speed); Wyrm adapts to RPG-style (strength, intelligence, charisma). |

**Import pattern:**
```go
import (
    vraycast  "github.com/opd-ai/violence/pkg/raycaster"
    vlaudio   "github.com/opd-ai/violence/pkg/audio"
    vcombat   "github.com/opd-ai/violence/pkg/combat"
)

// Extend Violence's raycaster for open-world rendering
type WyrmRenderer struct {
    caster   *vraycast.Raycaster
    textures *texture.ProceduralTextureCache
    skybox   *SkyboxRenderer
    weather  *WeatherOverlay
}
```

### Velocity (`opd-ai/velocity`) — Wave/Spawning Patterns, Audio, Balance

Velocity is a Galaga-like shmup, so direct code reuse is limited, but several architectural patterns transfer well.

| Velocity Package | What It Provides | How Wyrm Reuses It |
|------------------|------------------|--------------------|
| `pkg/procgen/spawner` | Deterministic entity spawning with seed-based RNG, spawn rate control, difficulty curves | Adapt spawner pattern for Wyrm's open-world enemy placement. Use Velocity's difficulty-curve algorithms for dungeon depth scaling and world-zone danger levels. |
| `pkg/procgen/wave_manager` | Wave-based encounter sequencing with escalating difficulty | Adapt wave concept for Wyrm's siege events, bandit raids, and dungeon encounter rooms. Replace fixed waves with dynamic event-triggered spawns. |
| `pkg/procgen/genre` | Genre-based content theming | Reference for genre-routing consistency. Ensure Wyrm's genre enum matches Velocity's for future shared-library extraction. |
| `pkg/audio` | Procedural audio with Ebiten integration (platform-aware stubs) | Import audio architecture pattern (Ebiten integration layer + stub for testing). Wyrm's audio builds on this same Ebiten audio context pattern. |
| `pkg/rendering` | Sprite rendering with culling system | Reference culling algorithm for Wyrm's entity render culling in the raycaster (don't render entities behind walls or beyond draw distance). |
| `pkg/balance` | Game balance parameter tuning | Adapt balance framework for Wyrm's RPG stat balancing (damage formulas, XP curves, economy pricing). |
| `pkg/combat` | Combat resolution (hit/damage/death) | Reference for damage calculation patterns. Wyrm's RPG combat is more complex but follows the same resolution pipeline. |
| `pkg/companion` | Companion AI system | Reference alongside Venture's companion generator. Velocity's real-time companion positioning logic applies to Wyrm's companion follow behavior. |
| `pkg/config` | Viper-based configuration | Match configuration loading pattern for consistency across suite. |

### Vania (`opd-ai/vania`) — PCG Validation, Seed Mixing, Camera, Physics

Vania is a Metroidvania using `internal/` packages (not importable as Go modules), but its patterns are valuable reference implementations.

| Vania Package | What It Provides | How Wyrm References It |
|---------------|------------------|-----------------------|
| `internal/pcg/seed` | Seed derivation and mixing functions (FNV-based) | **Reference implementation** for Wyrm's `mixSeeds()` function. Vania's approach to deterministic sub-seed derivation is the pattern Wyrm should follow for chunk, quest, and instance seeds. |
| `internal/pcg/cache` | Generated content caching with LRU eviction | **Reference implementation** for Wyrm's chunk cache and texture cache. Same LRU pattern applies to caching generated buildings, NPCs, and quest templates. |
| `internal/pcg/validator` | PCG output validation (connectivity, balance, completeness) | **Reference implementation** for Wyrm's world validation (ensure all chunks are traversable, all quest objectives are reachable, all economy loops are solvable). |
| `internal/camera` | 2D camera with smooth follow and bounds | Reference for camera interpolation math. Wyrm's first-person camera uses similar smoothing for look direction and position transitions. |
| `internal/physics` | 2D platformer physics (gravity, collision, slopes) | Reference for physics architecture. Wyrm's 3D approximated physics (gravity on projectiles, vehicle movement, falling) follows the same component-based approach. |
| `internal/animation` | Sprite animation system | Reference for Wyrm's weapon viewmodel animation and NPC animation state machines. |
| `internal/narrative` | In-game narrative/story system | Reference alongside Venture's narrative generator. Vania's runtime narrative delivery (text boxes, cutscene triggers) informs Wyrm's first-person narrative presentation. |
| `internal/particle` | Particle effects system | Reference for Wyrm's particle architecture. Same emitter/particle/lifecycle pattern. |
| `internal/save` | Save/load serialization | Reference for entity state serialization patterns used in Wyrm's server-side persistence. |

**Note:** Because Vania uses `internal/` packages, code cannot be imported directly. Instead, replicate the patterns in Wyrm's `pkg/` tree with the same algorithmic approach.

### Summary: What Is Imported vs. New

| Category | Imported from V-Series | Built New for Wyrm |
|----------|------------------------|---------------------|
| **Content Generation** | Terrain, entity, faction, quest, dialog, narrative, building, vehicle, magic, skills, class, companion, recipe, station, story, book, furniture, environment, item, legendary, puzzle, minigame (all from Venture) | Open-world chunk adapter, city layout (MST roads), dungeon BSP, NPC schedule FSM, crime/law system, dynamic economy, weather generation, world event system |
| **Rendering** | Raycaster core, textures, lighting, fog, particles, camera, weapon viewmodels (all from Violence) | Open-world terrain rendering, skybox, time-of-day cycle, genre post-processing, building interior transitions, procedural LOD, weather overlays |
| **Audio** | Oscillators, ADSR envelopes, SFX synthesis, reverb, ambient (from Violence + Velocity patterns) | Adaptive music system, genre-specific SFX mods, 3D spatial audio, location-based ambience, dialog voice synthesis |
| **Networking** | Lag compensation, delta compression, latency estimation, anti-cheat (from Violence) | Persistent-world connections, chunk streaming protocol, NPC authority model, quest instance sub-worlds, federation/cross-server travel, economy sync gossip protocol |
| **Combat** | Spatial hash, positional damage, telegraph system (from Violence) | RPG stat-based damage resolution, skill/magic integration, status effects, melee/ranged/magic combat triangle, stealth mechanics |
| **Patterns** | Seed mixing (Vania), spawner/wave (Velocity), balance framework (Velocity), PCG validation (Vania) | World persistence, player housing, guild territories, PvP zones, crafting workbenches, NPC memory, reputation system |

---

## How Wyrm Distinguishes Itself from V-Series Games

While Wyrm shares the opd-ai suite's zero-asset philosophy and ECS architecture, it is fundamentally a different class of game from all four V-Series titles. Understanding these distinctions prevents scope confusion and guides architectural decisions.

### Core Identity: Persistent Open-World RPG

The V-Series games are all **session-based** or **run-based** experiences:
- **Venture** = session-based co-op roguelike (enter dungeon → clear rooms → extract)
- **Violence** = match-based FPS (join server → play round → next map)
- **Velocity** = run-based shmup (start → survive waves → game over)
- **Vania** = single-player campaign (linear progression through generated world)

**Wyrm is none of these.** Wyrm is a **persistent, seamless open world** where:
- The world exists continuously on the server, evolving even when players are offline
- Players inhabit, modify, and leave lasting marks on a shared world
- There is no "run" or "match" — just an ongoing life in a generated world
- Time passes, factions wage wars, economies fluctuate, and NPCs live schedules

### Architectural Distinctions

| Dimension | V-Series Games | Wyrm |
|-----------|----------------|------|
| **World scope** | Bounded levels/rooms/arenas | Seamless infinite terrain via chunk streaming |
| **Persistence** | Session state lost on exit (or save-file snapshots) | Server-authoritative persistent world state |
| **NPC behavior** | Simple AI states (patrol/chase/attack) | Full daily schedules, memory, reputation, relationships |
| **Economy** | Static loot tables or shop inventories | Dynamic supply/demand, player-driven market, property ownership |
| **Faction system** | Static teams or allegiance flags | Dynamic territorial control, inter-faction wars, player reputation per faction |
| **Quest system** | Linear or template-based objectives | Branching narratives with persistent consequences that reshape the world |
| **Player agency** | Combat + movement | Combat, dialog, crafting, trading, stealing, building, governing, exploring |
| **Multiplayer model** | Match/session with lobby | Persistent shared world with federated cross-server travel |
| **Camera perspective** | Top-down (Venture), first-person (Violence), side-scroll (Vania), top-down (Velocity) | First-person immersive with contextual camera switches (dialog, vehicle, cinematic) |
| **Game loop** | Tight action loops (seconds to minutes) | Long-form RPG loops (hours to hundreds of hours) |

### What Wyrm Adds That No V-Series Game Has

1. **Persistent World Clock** — Time-of-day and calendar that drives NPC schedules, shop hours, faction events, weather cycles, and seasonal changes
2. **NPC Memory & Relationships** — NPCs remember player actions, form opinions, spread gossip, and change behavior over time
3. **Property Ownership** — Players buy, build, and furnish houses; own shops that generate passive income; claim territory for guilds
4. **Crime & Law System** — Witness-based crime detection, bounties, jail, reputation consequences — a full justice simulation
5. **Deep Crafting** — Gather materials from the open world, use workbenches with minigames, discover recipes, improve quality with skills
6. **Environmental Storytelling** — The world shows history: ruined buildings from past faction wars, abandoned camps, environmental clues to quests
7. **Dynamic World Events** — Sieges, plagues, market crashes, dragon attacks, faction coups — events that reshape the world and create emergent stories
8. **Cross-Server Federation** — Players travel between server instances, carrying inventory and quest state; economies sync across federation peers
9. **First-Person Dialog** — Face-to-face conversations with NPCs using the first-person camera, with facial expression rendering and gesture animation
10. **Vehicle Exploration** — Mount horses, drive buggies, pilot airships — traversal is part of the open-world experience, not just combat

---

## Quality Standards

### Testing Requirements

- **Coverage**: ≥40% per package (≥30% for Ebiten-dependent packages requiring xvfb)
- **Table-driven tests** for all business logic and generation functions
- **Benchmarks** for all hot-path code (rendering, physics, generation)
- **Race detection**: All tests must pass under `go test -race ./...`
- **Determinism tests**: Same seed must produce identical output across 3 runs

**Note:** Currently no test files exist (`*_test.go`). This is a significant gap to address.

### Code Review Quality Gates

- Build success: `go build ./cmd/client && go build ./cmd/server`
- All tests pass: `go test ./...`
- Race-free: `go test -race ./...`
- Static analysis: `go vet ./...`
- No new TODO/FIXME without corresponding tracking

### Documentation Requirements

- Every exported type and function has a godoc comment
- README.md stays in sync with CLI flags and features
- ROADMAP.md reflects current priorities and phase status

---

## Naming Conventions

- **Packages**: lowercase, single-word when possible (`engine`, `procgen`, `audio`, `render`)
- **Files**: snake_case (`terrain_generator.go`, `combat_system.go`)
- **Types**: PascalCase (`TerrainGenerator`, `CombatSystem`, `HealthComponent`)
- **Interfaces**: PascalCase, often ending in `-er` for single-method interfaces (`Generator`, `Renderer`)
- **Component types**: PascalCase, typically no suffix (`Position`, `Health`, `Faction`)
- **System types**: PascalCase + "System" suffix (`CombatSystem`, `RenderSystem`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Seeds**: Always `int64`, always named `seed` in function parameters
- **Delta time**: Always `float64`, always named `dt` in Update methods

---

## Genre Differentiation

Wyrm supports 5 genres that affect all procedural generation. When implementing features, parameterize by genre:

| Dimension | Fantasy | Sci-Fi | Horror | Cyberpunk | Post-Apoc |
|-----------|---------|--------|--------|-----------|-----------|
| World theme | Medieval magic | Colony planet | Cursed island | Megacity 2140 | Irradiated wasteland |
| Factions | Guilds, kingdoms | Corps, military | Cults, survivors | Megacorps, gangs | Tribes, raiders |
| Vehicles | Horse, cart, ship | Hover-bike, mech | Bone cart, hearse | Motorbike, drone | Buggy, gyrocopter |
| Visual palette | Warm gold/green | Cool blue/white | Desaturated grey | Neon pink/cyan | Sepia/orange dust |
| Audio style | Orchestral/lute | Synth/electronic | Dissonant/silence | EDM/glitch | Folk/wind |

Access current genre via `cfg.Genre` and pass it to all generators.

---

## Development Status and Gaps

Wyrm is in **Phase 1 (Foundation)** per ROADMAP.md. Key items to address:

### Immediate Gaps (Critical)

1. **No systems registered** — all ECS systems defined but not instantiated/registered in main
2. **No tests** — zero `*_test.go` files exist
3. **Empty system implementations** — all `Update()` methods are empty stubs
4. **ChunkManager unused** — created but assigned to `_` in server
5. **City generator unused** — `pkg/procgen/city/` exists but is never called
6. **Audio engine unused** — `pkg/audio/` exists but is never instantiated

### Phase 1 Completion Criteria (from ROADMAP.md)

- [ ] `go test ./pkg/engine/...` passes
- [ ] 10,000 entities created/destroyed in <5ms
- [ ] Venture generator integration
- [ ] Client connects to server, receives empty world state
- [ ] Genre routing passes `GenreID` to all generators

### Phase 2 Preview

- First-person raycaster at 60 FPS
- Procedural textures with genre-appropriate palettes
- Chunk streaming with <1 frame stutter at 50ms latency

Refer to ROADMAP.md for the complete 6-phase implementation plan.

---

## Procedural Generation Systems (PCG)

### Generator Inventory

Wyrm's procedural generation follows Venture's pattern of parameterized generators. Each generator accepts a seed and genre to produce deterministic, theme-appropriate content:

| Generator | Package | Algorithm | Status |
|-----------|---------|-----------|--------|
| World Terrain | `pkg/world/chunk/` | Per-chunk seed derivation | Skeleton |
| City Layout | `pkg/procgen/city/` | District-based generation | Skeleton |
| Textures | `pkg/rendering/texture/` | Procedural pixel fill | Skeleton |

### Future Generators (from ROADMAP.md)

| Generator | Planned Package | Algorithm |
|-----------|-----------------|-----------|
| Dungeon | `pkg/procgen/dungeon/` | BSP room graph |
| NPC Entity | via Venture `pkg/procgen/entity` | Stat template + name grammar |
| Faction | via Venture `pkg/procgen/faction` | Graph relations + territory Voronoi |
| Quest | via Venture `pkg/procgen/quest` | Template graph + consequence flags |
| Dialog | via Venture `pkg/procgen/dialog` | Topic graph + sentiment model |
| Item/Weapon | via Venture `pkg/procgen/recipe` | Affix table + material grammar |
| Building | via Venture `pkg/procgen/building` | Room template + façade grammar |
| Vehicle | via Venture `pkg/procgen/vehicle` | Archetype params + genre skin |

### Generator Implementation Pattern

```go
// Standard generator signature
type TerrainGenerator struct {
    Seed  int64
    Genre string
    rng   *rand.Rand
}

func NewTerrainGenerator(seed int64, genre string) *TerrainGenerator {
    return &TerrainGenerator{
        Seed:  seed,
        Genre: genre,
        rng:   rand.New(rand.NewSource(seed)),
    }
}

func (g *TerrainGenerator) Generate(params GenerationParams) (*Terrain, error) {
    // Derive sub-seeds for different aspects
    heightSeed := mixSeeds(g.Seed, "height")
    biomeSeed := mixSeeds(g.Seed, "biome")
    
    // Generate with genre-appropriate parameters
    palette := g.getPaletteForGenre(g.Genre)
    
    return &Terrain{
        HeightMap: g.generateHeightMap(heightSeed),
        BiomeMap:  g.generateBiomeMap(biomeSeed),
        Palette:   palette,
    }, nil
}
```

---

## Build and Run

### Building

```bash
# Build client (Ebitengine window)
go build ./cmd/client

# Build server (authoritative game server)
go build ./cmd/server

# Build both
go build ./cmd/...
```

### Running

```bash
# Start server (listens on localhost:7777 by default)
./server

# Start client (connects to server)
./client
```

### Configuration

Configuration is loaded from `config.yaml` in the working directory or `./config/` directory:

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

Environment variables override config file values:
- `WYRM_WINDOW_WIDTH=1920`
- `WYRM_WORLD_SEED=12345`
- `WYRM_GENRE=cyberpunk`

---

## Multiplayer Architecture

### Authority Model

- **Server is authoritative** for all world state
- Clients send input commands (movement, actions, interactions)
- Server validates, applies, and broadcasts delta states
- Never trust client-computed positions or damage values

### Chunk Streaming Protocol

1. Server tracks each client's 3×3 active chunk window
2. On chunk entry, server sends full chunk snapshot (compressed)
3. On chunk exit, server stops sending that chunk's deltas
4. Client interpolates entity positions between server snapshots

### NPC Authority

- Server owns all NPC state (schedule, faction, dialog, inventory)
- Clients render NPCs via interpolated position components
- NPC AI runs server-side only—clients never simulate NPC logic

### Quest Instances

Instanced dungeons/quests spin up a sub-world with a unique seed derived from:
```go
instanceSeed := mixSeeds(worldSeed, fmt.Sprintf("%s:%s", questID, partyID))
```

Party members join the same instance. Completion state writes back to persistent world.

### Player Housing

- Interior chunks linked to player `EntityID`
- Furniture component layout serialized per interior chunk
- Loaded on demand when any player enters

---

## File Organization Guidelines

### Adding New Components

1. Define the component struct in `pkg/engine/components/components.go`
2. Implement the `Type() string` method returning a unique string
3. Ensure the component is pure data—no methods beyond `Type()`

```go
// In pkg/engine/components/components.go
type Stamina struct {
    Current, Max float64
    RegenRate    float64
}

func (s *Stamina) Type() string { return "Stamina" }
```

### Adding New Systems

1. Define the system struct in `pkg/engine/systems/systems.go`
2. Implement `Update(w *ecs.World, dt float64)`
3. **CRITICAL**: Register the system in `main()` (both client and server if applicable)
4. **CRITICAL**: Ensure the system produces output consumed by another system

```go
// In pkg/engine/systems/systems.go
type StaminaSystem struct {
    regenMultiplier float64
}

func (s *StaminaSystem) Update(w *ecs.World, dt float64) {
    for _, e := range w.Entities("Stamina") {
        comp, _ := w.GetComponent(e, "Stamina")
        stamina := comp.(*components.Stamina)
        if stamina.Current < stamina.Max {
            stamina.Current += stamina.RegenRate * s.regenMultiplier * dt
            if stamina.Current > stamina.Max {
                stamina.Current = stamina.Max
            }
        }
    }
}

// In cmd/client/main.go or cmd/server/main.go
world.RegisterSystem(&systems.StaminaSystem{regenMultiplier: 1.0})
```

### Adding New Generators

1. Create a new package under `pkg/procgen/` (e.g., `pkg/procgen/dungeon/`)
2. Define the generator struct with `Seed` and `Genre` fields
3. Implement `Generate()` method returning the generated content
4. **CRITICAL**: Call the generator from world initialization code

```go
// pkg/procgen/dungeon/dungeon.go
package dungeon

type Dungeon struct {
    Rooms []Room
    Seed  int64
}

type Generator struct {
    Seed  int64
    Genre string
}

func NewGenerator(seed int64, genre string) *Generator {
    return &Generator{Seed: seed, Genre: genre}
}

func (g *Generator) Generate(depth int) (*Dungeon, error) {
    // BSP room generation algorithm
    return &Dungeon{Seed: g.Seed}, nil
}
```

---

## Debugging and Profiling

### Debug Output

The client displays debug info via `ebitenutil.DebugPrint`:
```go
ebitenutil.DebugPrint(screen, fmt.Sprintf("Wyrm [%s]", g.cfg.Genre))
```

Add additional debug output as needed during development.

### Performance Profiling

```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=. ./pkg/engine/ecs/
go tool pprof cpu.prof

# Memory profile
go test -memprofile=mem.prof -bench=. ./pkg/engine/ecs/
go tool pprof mem.prof

# Trace
go test -trace=trace.out ./pkg/engine/ecs/
go tool trace trace.out
```

### Race Detection

Always run tests with race detection during development:
```bash
go test -race ./...
```

---

## Common Pitfalls to Avoid

### 1. Global Random State
```go
// ❌ NEVER do this
value := rand.Intn(100)

// ✅ Always use explicit seed
rng := rand.New(rand.NewSource(seed))
value := rng.Intn(100)
```

### 2. Forgetting to Register Systems
```go
// ❌ System exists but does nothing
type MySystem struct{}
func (s *MySystem) Update(w *ecs.World, dt float64) { /* logic */ }
// ...but never called world.RegisterSystem(&MySystem{})

// ✅ System is registered and participates in game loop
mySystem := &MySystem{}
world.RegisterSystem(mySystem)
```

### 3. Blocking in Update/Draw
```go
// ❌ Blocks the game loop
func (g *Game) Update() error {
    data, _ := http.Get("http://server/state")  // NEVER
    return nil
}

// ✅ Non-blocking with channel
func (g *Game) Update() error {
    select {
    case data := <-g.dataChan:
        g.processData(data)
    default:
    }
    return nil
}
```

### 4. Logic in Components
```go
// ❌ Components should be pure data
type Position struct {
    X, Y, Z float64
}
func (p *Position) MoveForward(dist float64) {  // NO!
    p.Z += dist
}

// ✅ Logic belongs in systems
func (s *MovementSystem) Update(w *ecs.World, dt float64) {
    // Movement logic here
}
```

### 5. Concrete Network Types
```go
// ❌ Concrete types reduce testability
var conn *net.TCPConn

// ✅ Interface types
var conn net.Conn
```

---

## Success Criteria Summary

A contribution to Wyrm is complete when:

1. ✅ All new code compiles without errors
2. ✅ All tests pass (including `go test -race ./...`)
3. ✅ Every new system is registered in `main()`
4. ✅ Every new generator is called from runtime code
5. ✅ Every new component has a system that operates on it
6. ✅ Deterministic output: same seed → same result
7. ✅ No external asset files added
8. ✅ Genre parameter passed to all generation code
9. ✅ Network code uses interface types only
10. ✅ No blocking calls in Update/Draw loops
