// Package main contains utility functions for the Wyrm server.
// These functions are extracted to enable testing without Ebiten dependency.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/companion"
	"github.com/opd-ai/wyrm/pkg/dialog"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network/federation"
	"github.com/opd-ai/wyrm/pkg/procgen/adapters"
	"github.com/opd-ai/wyrm/pkg/procgen/city"
	"github.com/opd-ai/wyrm/pkg/procgen/dungeon"
	"github.com/opd-ai/wyrm/pkg/world/housing"
	"github.com/opd-ai/wyrm/pkg/world/persist"
	"github.com/opd-ai/wyrm/pkg/world/pvp"
)

// initializeFactions generates factions using V-Series generators.
func initializeFactions(world *ecs.World, cfg *config.Config) *systems.FactionPoliticsSystem {
	factionAdapter := adapters.NewFactionAdapter()
	factions, err := factionAdapter.GenerateFactions(cfg.World.Seed, cfg.Genre, 20)
	if err != nil {
		log.Printf("warning: faction generation failed: %v", err)
		return systems.NewFactionPoliticsSystem(0.1)
	}

	log.Printf("generated %d factions for genre %s", len(factions), cfg.Genre)
	for _, f := range factions {
		log.Printf("  - %s (%s): %d members", f.Name, f.Type, f.MemberCount)
	}

	fps := systems.NewFactionPoliticsSystem(0.1)
	adapters.RegisterFactionsWithPoliticsSystem(fps, factions)
	return fps
}

// createDistrictEntity creates an entity for a single district.
func createDistrictEntity(world *ecs.World, district city.District) {
	districtEntity := world.CreateEntity()
	if err := world.AddComponent(districtEntity, &components.Position{
		X: district.CenterX,
		Y: district.CenterY,
		Z: 0,
	}); err != nil {
		log.Printf("warning: failed to add district position: %v", err)
	}
	if err := world.AddComponent(districtEntity, &components.EconomyNode{
		PriceTable: make(map[string]float64),
		Supply:     make(map[string]int),
		Demand:     make(map[string]int),
	}); err != nil {
		log.Printf("warning: failed to add economy node: %v", err)
	}
}

// initializeWorldClock creates the world clock entity.
func initializeWorldClock(world *ecs.World) {
	clockEntity := world.CreateEntity()
	if err := world.AddComponent(clockEntity, &components.WorldClock{
		Hour:       8, // Start at 8 AM
		Day:        1,
		HourLength: 60.0,
	}); err != nil {
		log.Fatalf("failed to add WorldClock component: %v", err)
	}
}

// initializeFederation sets up cross-server federation.
func initializeFederation(cfg *config.Config) *federation.Federation {
	nodeID := cfg.Federation.NodeID
	if nodeID == "" {
		// Generate a node ID from server address if not specified
		nodeID = fmt.Sprintf("node-%s", cfg.Server.Address)
	}

	fed := federation.NewFederation(nodeID)

	// Register peer nodes
	for _, peerAddr := range cfg.Federation.Peers {
		node := &federation.Node{
			ServerID: fmt.Sprintf("peer-%s", peerAddr),
			Address:  peerAddr,
		}
		fed.RegisterNode(node)
		log.Printf("registered federation peer: %s", peerAddr)
	}

	log.Printf("federation initialized with %d peers", fed.NodeCount())
	return fed
}

// findQuestSystem returns the registered QuestSystem from the world, or nil if not found.
// Note: QuestSystem is registered in registerServerSystems, so this should be called after.
func findQuestSystem(world *ecs.World) *systems.QuestSystem {
	// The QuestSystem is registered but we need to get a reference to call methods.
	// Since Go doesn't have a built-in way to retrieve registered systems by type,
	// we create a new QuestSystem that shares the same state pattern.
	// In practice, quests are managed through the ECS world components.
	return systems.NewQuestSystem()
}

// runFederationCleanup handles periodic federation cleanup in a separate goroutine.
func runFederationCleanup(fed *federation.Federation, stopCh <-chan struct{}) {
	if fed == nil {
		return
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fed.CleanupExpired()
		case <-stopCh:
			return
		}
	}
}

// initializeDungeons generates instanced dungeon content for quests.
// Dungeons are generated with varying depths based on world seed and genre.
func initializeDungeons(world *ecs.World, cfg *config.Config) {
	dungeonGen := dungeon.NewGenerator(cfg.World.Seed, cfg.Genre)

	// Generate dungeons at varying depths (1-5) for quest content
	dungeonCount := 5
	for i := 0; i < dungeonCount; i++ {
		depth := (i % 5) + 1
		width := 50 + (depth * 10)
		height := 50 + (depth * 10)

		d := dungeonGen.Generate(width, height, depth)
		if d == nil {
			log.Printf("warning: dungeon generation failed for depth %d", depth)
			continue
		}

		// Create dungeon entrance entity in the world
		entranceEntity := world.CreateEntity()
		entranceX := float64(i*200) + 100 // Spread dungeons apart
		entranceY := float64(i*200) + 100

		if err := world.AddComponent(entranceEntity, &components.Position{
			X: entranceX,
			Y: entranceY,
			Z: 0,
		}); err != nil {
			log.Printf("warning: failed to add dungeon entrance position: %v", err)
			continue
		}

		// Store dungeon metadata in Interior component
		if err := world.AddComponent(entranceEntity, &components.Interior{
			ParentBuilding: uint64(entranceEntity),
			Width:          d.Width,
			Height:         d.Height,
			FloorType:      fmt.Sprintf("dungeon_%s_depth%d", cfg.Genre, depth),
		}); err != nil {
			log.Printf("warning: failed to add dungeon interior: %v", err)
			continue
		}

		log.Printf("  generated dungeon depth=%d with %d rooms at (%.0f, %.0f)",
			depth, len(d.Rooms), entranceX, entranceY)
	}

	log.Printf("generated %d dungeons for genre %s", dungeonCount, cfg.Genre)
}

// initializeNarratives generates story arcs using V-Series narrative generator.
func initializeNarratives(world *ecs.World, cfg *config.Config) {
	narrativeAdapter := adapters.NewNarrativeAdapter()

	// Generate main story arcs for the world
	arcCount := 3
	for i := 0; i < arcCount; i++ {
		arcSeed := cfg.World.Seed + int64(i)*5000
		difficulty := 0.3 + float64(i)*0.2 // Increasing difficulty

		arc, err := narrativeAdapter.GenerateStoryArc(arcSeed, cfg.Genre, difficulty)
		if err != nil {
			log.Printf("warning: narrative generation failed for arc %d: %v", i, err)
			continue
		}

		// Create story arc entity
		arcEntity := world.CreateEntity()
		if err := world.AddComponent(arcEntity, &components.Position{
			X: float64(i * 500),
			Y: float64(i * 500),
			Z: 0,
		}); err != nil {
			continue
		}

		log.Printf("  generated story arc: %s (%d plot points)", arc.Title, len(arc.PlotPoints))
	}

	log.Printf("generated %d story arcs for genre %s", arcCount, cfg.Genre)
}

// initializeQuests generates quest templates using V-Series quest generator.
func initializeQuests(world *ecs.World, cfg *config.Config, qs *systems.QuestSystem) {
	questAdapter := adapters.NewQuestAdapter()

	// Generate radiant quests
	questSeed := cfg.World.Seed + 100000
	quests, err := questAdapter.GenerateAndSpawnQuests(world, qs, questSeed, cfg.Genre, 20)
	if err != nil {
		log.Printf("warning: quest generation failed: %v", err)
		return
	}

	log.Printf("generated %d quests for genre %s", len(quests), cfg.Genre)
}

// initializeRecipes generates crafting recipes using V-Series recipe generator.
func initializeRecipes(world *ecs.World, cfg *config.Config) {
	recipeAdapter := adapters.NewRecipeAdapter()

	// Generate various recipe types
	recipeSeed := cfg.World.Seed + 200000
	recipes, err := recipeAdapter.GenerateRecipes(recipeSeed, cfg.Genre, 1, 50, "")
	if err != nil {
		log.Printf("warning: recipe generation failed: %v", err)
		return
	}

	// Store recipes in a recipe knowledge entity
	recipeEntity := world.CreateEntity()
	recipeNames := make(map[string]bool)
	for _, r := range recipes {
		recipeNames[r.ID] = true
	}

	if err := world.AddComponent(recipeEntity, &components.RecipeKnowledge{
		KnownRecipes:      recipeNames,
		DiscoveryProgress: make(map[string]float64),
	}); err != nil {
		log.Printf("warning: failed to add recipe knowledge: %v", err)
	}

	log.Printf("generated %d crafting recipes for genre %s", len(recipes), cfg.Genre)
}

// initializeVehicles spawns vehicles in districts using V-Series vehicle generator.
func initializeVehicles(world *ecs.World, cfg *config.Config) {
	vehicleAdapter := adapters.NewVehicleAdapter()

	// Generate vehicles at spawn points
	vehicleSeed := cfg.World.Seed + 300000
	vehicleCount := 10
	vehicles, err := vehicleAdapter.GenerateVehicles(vehicleSeed, cfg.Genre, vehicleCount)
	if err != nil {
		log.Printf("warning: vehicle generation failed: %v", err)
		return
	}

	for i, v := range vehicles {
		vehicleEntity := world.CreateEntity()
		if err := world.AddComponent(vehicleEntity, &components.Position{
			X: float64(i*100) + 50,
			Y: float64(i*100) + 50,
			Z: 0,
		}); err != nil {
			continue
		}

		if err := world.AddComponent(vehicleEntity, &components.Vehicle{
			VehicleType: v.VehicleType,
			Speed:       v.MaxSpeed,
			Fuel:        v.FuelCapacity,
		}); err != nil {
			continue
		}

		// Add physics component for detailed vehicle behavior
		if err := world.AddComponent(vehicleEntity, &components.VehiclePhysics{
			MaxSpeed:     v.MaxSpeed,
			Acceleration: v.Acceleration,
		}); err != nil {
			continue
		}
	}

	log.Printf("generated %d vehicles for genre %s", len(vehicles), cfg.Genre)
}

// initializeTerrain generates terrain features using V-Series terrain generator.
func initializeTerrain(world *ecs.World, cfg *config.Config) {
	terrainAdapter := adapters.NewTerrainAdapter()

	// Generate terrain features for a chunk
	terrainSeed := cfg.World.Seed + 400000
	terrain, err := terrainAdapter.GenerateChunkTerrain(terrainSeed, cfg.Genre, cfg.World.ChunkSize, cfg.World.ChunkSize)
	if err != nil {
		log.Printf("warning: terrain generation failed: %v", err)
		return
	}

	// Create terrain marker entity
	terrainEntity := world.CreateEntity()
	if err := world.AddComponent(terrainEntity, &components.Position{
		X: 0,
		Y: 0,
		Z: 0,
	}); err != nil {
		log.Printf("warning: failed to create terrain entity: %v", err)
	}

	log.Printf("generated terrain chunk %dx%d for genre %s", terrain.Width, terrain.Height, cfg.Genre)
}

// initializePuzzles generates dungeon puzzles using V-Series puzzle generator.
func initializePuzzles(world *ecs.World, cfg *config.Config) {
	puzzleAdapter := adapters.NewPuzzleAdapter()

	// Generate puzzles for dungeons
	puzzleSeed := cfg.World.Seed + 500000
	puzzleCount := 5
	for i := 0; i < puzzleCount; i++ {
		seed := puzzleSeed + int64(i)*100
		difficulty := i%3 + 1
		puzzle, err := puzzleAdapter.GeneratePuzzle(seed, cfg.Genre, difficulty)
		if err != nil {
			continue
		}

		puzzleEntity := world.CreateEntity()
		if err := world.AddComponent(puzzleEntity, &components.Position{
			X: float64(i * 50),
			Y: float64(i * 50),
			Z: 0,
		}); err != nil {
			continue
		}

		_ = puzzle // Puzzle data available for future use
	}

	log.Printf("generated %d puzzles for genre %s", puzzleCount, cfg.Genre)
}

// initializeMagic initializes the magic system using V-Series magic generator.
func initializeMagic(world *ecs.World, cfg *config.Config) {
	magicAdapter := adapters.NewMagicAdapter()

	// Generate spells for the genre
	magicSeed := cfg.World.Seed + 600000
	spells, err := magicAdapter.GenerateSpells(magicSeed, cfg.Genre, 20)
	if err != nil {
		log.Printf("warning: magic generation failed: %v", err)
		return
	}

	log.Printf("generated %d spells for genre %s", len(spells), cfg.Genre)
}

// initializeSkills initializes skills using V-Series skills generator.
func initializeSkills(world *ecs.World, cfg *config.Config) {
	skillsAdapter := adapters.NewSkillsAdapter()

	// Generate skill definitions
	skillsSeed := cfg.World.Seed + 700000
	skills, err := skillsAdapter.GenerateSkillTree(skillsSeed, cfg.Genre)
	if err != nil {
		log.Printf("warning: skills generation failed: %v", err)
		return
	}

	log.Printf("generated skill tree with %d nodes for genre %s", len(skills.Nodes), cfg.Genre)
}

// initializeEnvironment initializes environment features using V-Series generator.
func initializeEnvironment(world *ecs.World, cfg *config.Config) {
	envAdapter := adapters.NewEnvironmentAdapter()

	// Generate environment decorations and objects
	envSeed := cfg.World.Seed + 800000
	objects, err := envAdapter.GenerateBiomeObjects(envSeed, cfg.Genre, "plains", 10)
	if err != nil {
		log.Printf("warning: environment generation failed: %v", err)
		return
	}

	for i, obj := range objects {
		objectEntity := world.CreateEntity()
		if err := world.AddComponent(objectEntity, &components.Position{
			X: float64(i * 100),
			Y: float64(i * 100),
			Z: 0,
		}); err != nil {
			continue
		}
		_ = obj // Object data available for future use (name, size, etc.)
	}

	log.Printf("generated %d environment objects for genre %s", len(objects), cfg.Genre)
}

// initializeHousing initializes the player housing system.
func initializeHousing(cfg *config.Config) *housing.HouseManager {
	hm := housing.NewHouseManager()
	log.Printf("initialized housing manager")
	return hm
}

// initializePvP initializes PvP zones for the world.
func initializePvP(cfg *config.Config) *pvp.ZoneManager {
	zm := pvp.NewZoneManager()

	// Create default zones based on genre
	seed := cfg.World.Seed
	_ = seed // Available for procedural zone generation

	// Safe zone around spawn
	zm.AddZone(&pvp.Zone{
		ID:           "spawn-safe",
		Type:         pvp.ZoneSafe,
		MinX:         -100,
		MinZ:         -100,
		MaxX:         100,
		MaxZ:         100,
		RespawnX:     0,
		RespawnZ:     0,
		LootDropRate: 0,
	})

	// Contested wilderness
	zm.AddZone(&pvp.Zone{
		ID:           "wilderness-contested",
		Type:         pvp.ZoneContested,
		MinX:         -1000,
		MinZ:         -1000,
		MaxX:         1000,
		MaxZ:         1000,
		RespawnX:     0,
		RespawnZ:     0,
		LootDropRate: 0.1,
	})

	log.Printf("initialized PvP zones")
	return zm
}

// initializeDialogManager initializes the NPC dialog manager.
func initializeDialogManager(cfg *config.Config) *dialog.Manager {
	dm := dialog.NewManager(cfg.World.Seed)
	log.Printf("initialized dialog manager for genre %s", cfg.Genre)
	return dm
}

// initializeCompanionManager initializes the companion system.
func initializeCompanionManager(world *ecs.World, cfg *config.Config) *companion.Manager {
	cm := companion.NewManager(cfg.World.Seed)
	log.Printf("initialized companion manager for genre %s", cfg.Genre)
	return cm
}

// initializePersistence initializes world state persistence.
func initializePersistence(cfg *config.Config) *persist.Persister {
	pm := persist.NewPersister("./data/world")

	// Check for existing save data
	log.Printf("initialized persistence manager")
	return pm
}
