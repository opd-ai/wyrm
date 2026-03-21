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
