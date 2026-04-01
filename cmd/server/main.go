//go:build !noebiten

// Command server launches the Wyrm authoritative game server.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/network/federation"
	"github.com/opd-ai/wyrm/pkg/procgen/adapters"
	"github.com/opd-ai/wyrm/pkg/procgen/city"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	world := ecs.NewWorld()
	cm := chunk.NewManager(cfg.World.ChunkSize, cfg.World.Seed)

	fps := initializeFactions(world, cfg)
	initializeCity(world, cfg, fps)
	initializeWorldClock(world)
	initializeDungeons(world, cfg)
	initializeNarratives(world, cfg)
	initializeTerrain(world, cfg)
	initializeVehicles(world, cfg)
	initializePuzzles(world, cfg)
	initializeMagic(world, cfg)
	initializeSkills(world, cfg)
	initializeEnvironment(world, cfg)
	registerServerSystems(world, cm, cfg, fps)

	// Initialize quests after systems are registered (needs QuestSystem)
	qs := findQuestSystem(world)
	if qs != nil {
		initializeQuests(world, cfg, qs)
		initializeRecipes(world, cfg)
	}

	// Initialize federation if enabled
	var fed *federation.Federation
	if cfg.Federation.Enabled {
		fed = initializeFederation(cfg)
	}

	srv := network.NewServer(cfg.Server.Address)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server start: %v\n", err)
		os.Exit(1)
	}
	defer srv.Stop()

	log.Printf("server listening on %s (tick_rate=%d)", cfg.Server.Address, cfg.Server.TickRate)
	if cfg.Federation.Enabled {
		log.Printf("federation enabled: node_id=%s, peers=%v", cfg.Federation.NodeID, cfg.Federation.Peers)
	}
	runServerLoop(world, cfg, srv, fed)
}

// initializeCity generates the city and spawns district entities.
func initializeCity(world *ecs.World, cfg *config.Config, fps *systems.FactionPoliticsSystem) {
	generatedCity := city.Generate(cfg.World.Seed, cfg.Genre)
	log.Printf("generated city: %s (%s) with %d districts", generatedCity.Name, cfg.Genre, len(generatedCity.Districts))

	// Generate NPCs using V-Series entity generator
	entityAdapter := adapters.NewEntityAdapter()
	factionAdapter := adapters.NewFactionAdapter()
	buildingAdapter := adapters.NewBuildingAdapter()
	dialogAdapter := adapters.NewDialogAdapter()
	itemAdapter := adapters.NewItemAdapter()
	furnitureAdapter := adapters.NewFurnitureAdapter()
	factions, _ := factionAdapter.GenerateFactions(cfg.World.Seed, cfg.Genre, 20)

	for i, district := range generatedCity.Districts {
		createDistrictEntity(world, district)

		// Generate building interiors for this district
		districtSeed := cfg.World.Seed + int64(i)*10000
		buildingsInDistrict := district.Buildings
		if buildingsInDistrict > 10 {
			buildingsInDistrict = 10 // Cap buildings per district for performance
		}

		itemsGenerated := 0
		furnitureGenerated := 0
		for b := 0; b < buildingsInDistrict; b++ {
			buildingSeed := districtSeed + int64(b)*100
			buildingType := b % 5 // Cycle through building types
			floors := 1 + (b % 3)

			bldg, err := buildingAdapter.GenerateBuilding(buildingSeed, cfg.Genre, buildingType, floors)
			if err != nil {
				log.Printf("warning: building generation failed in district %s: %v", district.Name, err)
				continue
			}

			// Create building entity in the world
			buildingEntity := world.CreateEntity()
			buildingX := district.CenterX + float64((b%3)-1)*30
			buildingY := district.CenterY + float64((b/3)-1)*30

			if err := world.AddComponent(buildingEntity, &components.Position{
				X: buildingX,
				Y: buildingY,
				Z: 0,
			}); err != nil {
				continue
			}

			if err := world.AddComponent(buildingEntity, &components.Building{
				BuildingType: bldg.Type,
				Width:        float64(bldg.Width),
				Height:       float64(bldg.Height),
				Floors:       bldg.Floors,
				IsOpen:       true,
			}); err != nil {
				continue
			}

			// Create interior for the building
			if err := world.AddComponent(buildingEntity, &components.Interior{
				ParentBuilding: uint64(buildingEntity),
				Width:          bldg.Width,
				Height:         bldg.Height,
				FloorType:      bldg.Style,
			}); err != nil {
				continue
			}

			// Generate furniture for the building interior
			furnitureSeed := buildingSeed + 300
			roomType := []string{"common", "bedroom", "shop", "kitchen", "storage"}[buildingType%5]
			furnitureItems, err := furnitureAdapter.GenerateRoomFurniture(furnitureSeed, cfg.Genre, roomType, 3)
			if err == nil && len(furnitureItems) > 0 {
				furnitureIDs := make([]uint64, 0, len(furnitureItems))
				for fi, furn := range furnitureItems {
					furnEntity := world.CreateEntity()
					if err := world.AddComponent(furnEntity, &components.Position{
						X: buildingX + float64(fi%3)*2,
						Y: buildingY + float64(fi/3)*2,
						Z: 0,
					}); err != nil {
						continue
					}
					furnitureIDs = append(furnitureIDs, uint64(furnEntity))
					furnitureGenerated++
					_ = furn // Furniture data available for future use
				}
			}

			// For shop buildings (type 0), generate inventory items
			if buildingType == 0 {
				itemSeed := buildingSeed + 500
				items, err := itemAdapter.GenerateItems(itemSeed, cfg.Genre, 1, 10, "")
				if err == nil && len(items) > 0 {
					shopItems := make(map[string]int)
					shopPrices := make(map[string]float64)
					for _, itm := range items {
						shopItems[itm.ID] = 1 + (b % 3) // Quantity varies
						shopPrices[itm.ID] = float64(itm.Stats.Value)
					}

					if err := world.AddComponent(buildingEntity, &components.ShopInventory{
						ShopType:        "general",
						Items:           shopItems,
						Prices:          shopPrices,
						RestockInterval: 24,
						GoldReserve:     500.0 + float64(b*100),
					}); err == nil {
						itemsGenerated += len(items)
					}
				}
			}
		}
		log.Printf("  generated %d buildings with %d items, %d furniture in district %s", buildingsInDistrict, itemsGenerated, furnitureGenerated, district.Name)

		// Spawn NPCs in each district using V-Series generator
		factionID := "neutral"
		if len(factions) > 0 {
			factionID = factions[i%len(factions)].ID
		}

		npcsPerDistrict := 5 + (district.Buildings / 10)
		if npcsPerDistrict > 20 {
			npcsPerDistrict = 20
		}

		npcCfg := adapters.NPCSpawnConfig{
			Seed:      districtSeed,
			Genre:     cfg.Genre,
			FactionID: factionID,
			Count:     npcsPerDistrict,
			CenterX:   district.CenterX,
			CenterY:   district.CenterY,
			Radius:    100.0,
		}
		entities, err := entityAdapter.GenerateAndSpawnNPCs(world, npcCfg)
		if err != nil {
			log.Printf("warning: NPC spawn failed for district %s: %v", district.Name, err)
			continue
		}

		// Generate dialog trees for each spawned NPC
		dialogsGenerated := 0
		for j, npcEntity := range entities {
			npcDialogSeed := districtSeed + int64(j)*7
			dialogLines, err := dialogAdapter.GenerateDialogLines(npcDialogSeed, cfg.Genre, 5)
			if err != nil {
				continue
			}

			// Create dialog options from generated lines
			options := make([]components.DialogOption, len(dialogLines))
			for k, line := range dialogLines {
				options[k] = components.DialogOption{
					ID:          fmt.Sprintf("topic_%d", k),
					Text:        line.Text,
					NextTopicID: fmt.Sprintf("topic_%d", (k+1)%len(dialogLines)),
				}
			}

			if err := world.AddComponent(npcEntity, &components.DialogState{
				CurrentTopicID:     "greeting",
				AvailableResponses: options,
			}); err != nil {
				continue
			}
			dialogsGenerated++
		}

		log.Printf("  spawned %d NPCs with %d dialogs in district %s", len(entities), dialogsGenerated, district.Name)
	}
}

// registerServerSystems registers all 57 server-side ECS systems.
// Systems are registered in dependency order according to PLAN.md.
func registerServerSystems(world *ecs.World, cm *chunk.Manager, cfg *config.Config, fps *systems.FactionPoliticsSystem) {
	seed := cfg.World.Seed
	genre := cfg.Genre

	// Foundation systems (Phase 1)
	world.RegisterSystem(systems.NewWorldClockSystem(60.0))
	world.RegisterSystem(systems.NewWorldChunkSystem(cm, cfg.World.ChunkSize))
	world.RegisterSystem(&systems.NPCScheduleSystem{})
	world.RegisterSystem(fps) // Pre-initialized faction politics system
	crimeSystem := systems.NewCrimeSystem(60.0, 100.0)
	world.RegisterSystem(crimeSystem)
	economySystem := systems.NewEconomySystem(0.5, 0.1)
	world.RegisterSystem(economySystem)
	world.RegisterSystem(systems.NewCombatSystem())
	world.RegisterSystem(&systems.VehicleSystem{})
	world.RegisterSystem(systems.NewQuestSystem())
	weatherSystem := systems.NewWeatherSystem(genre, 300.0)
	world.RegisterSystem(weatherSystem)

	// NPC behavior systems (Phase 2)
	world.RegisterSystem(systems.NewNPCPathfindingSystem())
	world.RegisterSystem(systems.NewNPCNeedsSystem())
	world.RegisterSystem(systems.NewNPCOccupationSystem(seed))
	world.RegisterSystem(systems.NewEmotionalStateSystem())
	world.RegisterSystem(systems.NewNPCMemorySystem())
	world.RegisterSystem(systems.NewGossipSystem())

	// Faction depth systems (Phase 3)
	factionRankSystem := systems.NewFactionRankSystem(genre)
	world.RegisterSystem(factionRankSystem)
	world.RegisterSystem(systems.NewFactionCoupSystem(factionRankSystem, fps, seed, genre))
	world.RegisterSystem(systems.NewFactionExclusiveContentSystem(factionRankSystem, genre))
	world.RegisterSystem(systems.NewDynamicFactionWarSystem(fps))

	// Crime depth systems (Phase 4)
	guardPursuitSystem := systems.NewGuardPursuitSystem(crimeSystem)
	world.RegisterSystem(guardPursuitSystem)
	world.RegisterSystem(systems.NewBriberySystem(crimeSystem, guardPursuitSystem, seed))
	crimeEvidenceSystem := systems.NewCrimeEvidenceSystem(crimeSystem, genre, seed)
	world.RegisterSystem(crimeEvidenceSystem)
	world.RegisterSystem(systems.NewPardonSystem(crimeSystem, crimeEvidenceSystem, genre, seed))
	world.RegisterSystem(systems.NewCriminalFactionQuestSystem(factionRankSystem, genre, seed))

	// Economy depth systems (Phase 5)
	world.RegisterSystem(systems.NewEconomicEventSystem(seed, genre, economySystem))
	world.RegisterSystem(systems.NewMarketManipulationSystem(seed, genre, economySystem))
	world.RegisterSystem(systems.NewTradeRouteSystem(seed, genre, economySystem))
	world.RegisterSystem(systems.NewInvestmentSystem(seed, genre))
	world.RegisterSystem(systems.NewPlayerShopSystem(economySystem))
	world.RegisterSystem(systems.NewCityBuildingSystem(genre, seed))
	world.RegisterSystem(systems.NewCityEventSystem(genre, seed))
	world.RegisterSystem(systems.NewTradingSystem())

	// Combat depth systems (Phase 6)
	world.RegisterSystem(systems.NewMagicSystem())
	world.RegisterSystem(systems.NewProjectileSystem())
	world.RegisterSystem(systems.NewStealthSystem())
	world.RegisterSystem(systems.NewDistractionSystem())
	world.RegisterSystem(systems.NewHidingSpotSystem(float64(cfg.World.ChunkSize)))
	world.RegisterSystem(systems.NewVehiclePhysicsSystem(genre))
	world.RegisterSystem(systems.NewVehicleCombatSystem())
	world.RegisterSystem(systems.NewFlyingVehicleSystem(genre))
	world.RegisterSystem(systems.NewNavalVehicleSystem(genre))
	world.RegisterSystem(systems.NewMountSystem(seed, genre))

	// Skills and crafting systems (Phase 7)
	skillRegistry := systems.NewSkillRegistry()
	skillProgressionSystem := systems.NewSkillProgressionSystem(100.0, 100)
	world.RegisterSystem(skillProgressionSystem)
	world.RegisterSystem(systems.NewSkillBookSystem(skillRegistry, skillProgressionSystem))
	world.RegisterSystem(systems.NewSkillSynergySystem(skillRegistry))
	world.RegisterSystem(systems.NewActionUnlockSystem(skillRegistry, skillProgressionSystem))
	world.RegisterSystem(systems.NewNPCTrainingSystem(skillRegistry, skillProgressionSystem))
	world.RegisterSystem(systems.NewCraftingSystem(seed))

	// Dialog and social systems (Phase 8)
	world.RegisterSystem(systems.NewDialogConsequenceSystem())
	world.RegisterSystem(systems.NewMultiNPCConversationSystem())
	world.RegisterSystem(systems.NewPartySystem())
	world.RegisterSystem(systems.NewVehicleCustomizationSystem(seed, genre))

	// Environment systems (Phase 9)
	world.RegisterSystem(systems.NewIndoorOutdoorSystem(weatherSystem))
	world.RegisterSystem(systems.NewHazardSystem(genre))

	log.Printf("registered %d server systems", 57)
}

// runServerLoop runs the main server tick loop until shutdown.
func runServerLoop(world *ecs.World, cfg *config.Config, srv *network.Server, fed *federation.Federation) {
	tickInterval := time.Second / time.Duration(cfg.Server.TickRate)
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start federation cleanup goroutine
	fedStopCh := make(chan struct{})
	go runFederationCleanup(fed, fedStopCh)
	defer close(fedStopCh)

	for {
		select {
		case <-ticker.C:
			world.Update(tickInterval.Seconds())
		case <-sigCh:
			log.Println("shutting down")
			return
		}
	}
}
