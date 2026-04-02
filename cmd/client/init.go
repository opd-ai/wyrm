//go:build !noebiten

// Package main provides the game client entry point.
// init.go contains game initialization functions including player setup,
// ECS system registration, audio initialization, and weather setup.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/network"
	"github.com/opd-ai/wyrm/pkg/rendering/particles"
)

// createPlayerEntity creates and configures the player entity.
func createPlayerEntity(world *ecs.World) ecs.Entity {
	player := world.CreateEntity()
	addPlayerComponents(world, player)
	return player
}

// addPlayerComponents adds all required components to the player entity.
func addPlayerComponents(world *ecs.World, player ecs.Entity) {
	componentList := []ecs.Component{
		&components.Position{X: 8.5, Y: 8.5, Z: 0},
		&components.Health{Current: 100, Max: 100},
		&components.Mana{Current: 50, Max: 50, RegenRate: 1.0},
		&components.Skills{
			Levels:        make(map[string]int),
			Experience:    make(map[string]float64),
			SchoolBonuses: make(map[string]float64),
		},
		&components.Inventory{Items: []string{}, Capacity: 30},
		&components.Faction{ID: "player", Reputation: 0},
		&components.Reputation{Standings: make(map[string]float64)},
		&components.Stealth{
			Visibility:      1.0,
			BaseVisibility:  1.0,
			SneakVisibility: sneakBaseVis,
			DetectionRadius: 15.0,
		},
		&components.CombatState{},
		&components.AudioListener{Volume: 1.0, Enabled: true},
		&components.Weapon{
			Name:        "Fists",
			Damage:      5,
			Range:       1.5,
			AttackSpeed: 1.0,
			WeaponType:  "melee",
		},
	}

	for _, c := range componentList {
		if err := world.AddComponent(player, c); err != nil {
			log.Fatalf("failed to add %s component: %v", c.Type(), err)
		}
	}
}

// registerClientSystems registers all client-side ECS systems.
// In single-player mode (offline), this also registers core game logic systems.
func registerClientSystems(world *ecs.World, player ecs.Entity, cfg *config.Config, offline bool) {
	// Note: RenderSystem is not registered as camera sync is handled directly
	// in the game loop via syncRendererPosition(). Register it when render
	// preparation logic (culling, LOD selection) is implemented.
	_ = player // Available when RenderSystem is needed

	// Audio and weather systems (always needed)
	world.RegisterSystem(&systems.AudioSystem{Genre: cfg.Genre})
	weatherSys := systems.NewWeatherSystem(cfg.Genre, 300.0)
	world.RegisterSystem(weatherSys)

	// In offline mode, register essential gameplay systems for single-player
	if offline {
		registerSinglePlayerSystems(world, cfg, weatherSys)
	}
}

// registerSinglePlayerSystems registers game logic systems for offline/single-player mode.
// These systems provide the core RPG mechanics when not connected to a server.
func registerSinglePlayerSystems(world *ecs.World, cfg *config.Config, weatherSys *systems.WeatherSystem) {
	seed := cfg.World.Seed
	genre := cfg.Genre

	// World time (drives NPC schedules, shop hours, etc.)
	world.RegisterSystem(systems.NewWorldClockSystem(secondsPerMinute))

	// NPC behavior systems
	world.RegisterSystem(&systems.NPCScheduleSystem{})
	world.RegisterSystem(systems.NewNPCPathfindingSystem())
	world.RegisterSystem(systems.NewNPCNeedsSystem())
	world.RegisterSystem(systems.NewNPCOccupationSystem(seed))
	world.RegisterSystem(systems.NewEmotionalStateSystem())
	world.RegisterSystem(systems.NewNPCMemorySystem())
	world.RegisterSystem(systems.NewGossipSystem())

	// Faction systems
	fps := systems.NewFactionPoliticsSystem(0.1)
	world.RegisterSystem(fps)
	factionRankSystem := systems.NewFactionRankSystem(genre)
	world.RegisterSystem(factionRankSystem)
	world.RegisterSystem(systems.NewFactionCoupSystem(factionRankSystem, fps, seed, genre))
	world.RegisterSystem(systems.NewFactionExclusiveContentSystem(factionRankSystem, genre))
	world.RegisterSystem(systems.NewDynamicFactionWarSystem(fps))

	// Crime and law systems
	crimeSystem := systems.NewCrimeSystem(secondsPerMinute, baseWantedLevel)
	world.RegisterSystem(crimeSystem)
	guardPursuitSystem := systems.NewGuardPursuitSystem(crimeSystem)
	world.RegisterSystem(guardPursuitSystem)
	world.RegisterSystem(systems.NewBriberySystem(crimeSystem, guardPursuitSystem, seed))
	crimeEvidenceSystem := systems.NewCrimeEvidenceSystem(crimeSystem, genre, seed)
	world.RegisterSystem(crimeEvidenceSystem)
	world.RegisterSystem(systems.NewPardonSystem(crimeSystem, crimeEvidenceSystem, genre, seed))
	world.RegisterSystem(systems.NewCriminalFactionQuestSystem(factionRankSystem, genre, seed))

	// Economy systems
	economySystem := systems.NewEconomySystem(0.5, 0.1)
	world.RegisterSystem(economySystem)
	world.RegisterSystem(systems.NewEconomicEventSystem(seed, genre, economySystem))
	world.RegisterSystem(systems.NewMarketManipulationSystem(seed, genre, economySystem))
	world.RegisterSystem(systems.NewTradeRouteSystem(seed, genre, economySystem))
	world.RegisterSystem(systems.NewInvestmentSystem(seed, genre))
	world.RegisterSystem(systems.NewPlayerShopSystem(economySystem))
	world.RegisterSystem(systems.NewCityBuildingSystem(genre, seed))
	world.RegisterSystem(systems.NewCityEventSystem(genre, seed))
	world.RegisterSystem(systems.NewTradingSystem())

	// Combat systems
	world.RegisterSystem(systems.NewCombatSystem())
	world.RegisterSystem(systems.NewMagicSystem())
	world.RegisterSystem(systems.NewProjectileSystem())
	world.RegisterSystem(systems.NewStealthSystem())
	world.RegisterSystem(systems.NewDistractionSystem())
	world.RegisterSystem(systems.NewHidingSpotSystem(float64(cfg.World.ChunkSize)))

	// Vehicle systems
	world.RegisterSystem(&systems.VehicleSystem{})
	world.RegisterSystem(systems.NewVehiclePhysicsSystem(genre))
	world.RegisterSystem(systems.NewVehicleCombatSystem())
	world.RegisterSystem(systems.NewFlyingVehicleSystem(genre))
	world.RegisterSystem(systems.NewNavalVehicleSystem(genre))
	world.RegisterSystem(systems.NewMountSystem(seed, genre))

	// Quest system
	world.RegisterSystem(systems.NewQuestSystem())

	// Skills and crafting systems
	skillRegistry := systems.NewSkillRegistry()
	skillProgressionSystem := systems.NewSkillProgressionSystem(100.0, 100)
	world.RegisterSystem(skillProgressionSystem)
	world.RegisterSystem(systems.NewSkillBookSystem(skillRegistry, skillProgressionSystem))
	world.RegisterSystem(systems.NewSkillSynergySystem(skillRegistry))
	world.RegisterSystem(systems.NewActionUnlockSystem(skillRegistry, skillProgressionSystem))
	world.RegisterSystem(systems.NewNPCTrainingSystem(skillRegistry, skillProgressionSystem))
	world.RegisterSystem(systems.NewCraftingSystem(seed))

	// Dialog and social systems
	world.RegisterSystem(systems.NewDialogConsequenceSystem())
	world.RegisterSystem(systems.NewMultiNPCConversationSystem())
	world.RegisterSystem(systems.NewPartySystem())
	world.RegisterSystem(systems.NewVehicleCustomizationSystem(seed, genre))

	// Environment systems
	world.RegisterSystem(systems.NewIndoorOutdoorSystem(weatherSys))
	world.RegisterSystem(systems.NewHazardSystem(genre))

	// Physics and destruction systems
	world.RegisterSystem(systems.NewPhysicsSystem())
	world.RegisterSystem(systems.NewBarrierDestructionSystem())

	// Health systems
	world.RegisterSystem(systems.NewHealthRegenSystem())
}

// startAmbientAudio initializes and plays genre-appropriate ambient audio.
func (g *Game) startAmbientAudio() {
	if g.audioPlayer == nil {
		return
	}

	// Use the ambient soundscape mixer if available
	if g.ambientMixer != nil {
		duration := 2.0 // seconds
		samples := g.ambientMixer.GenerateMixedSamples(duration)
		// Reduce volume for ambient background
		for i := range samples {
			samples[i] *= 0.15
		}
		g.audioPlayer.QueueSamples(samples)
		g.audioPlayer.Play()
		return
	}

	// Fallback to simple sine wave if ambient mixer not available
	if g.audioEngine == nil {
		return
	}
	freq := g.audioEngine.GetGenreBaseFrequency()
	duration := 2.0 // seconds
	samples := g.audioEngine.GenerateSineWave(freq, duration)
	// Apply a gentle ADSR envelope for smooth ambient sound
	samples = g.audioEngine.ApplyADSR(samples, 0.5, 0.2, 0.3, 0.5)
	// Reduce volume for ambient background
	for i := range samples {
		samples[i] *= 0.1
	}
	g.audioPlayer.QueueSamples(samples)
	g.audioPlayer.Play()
}

// connectToServer attempts to connect to the game server.
func connectToServer(cfg *config.Config) (*network.Client, bool) {
	client := network.NewClient(cfg.Server.Address)
	if err := client.Connect(); err != nil {
		log.Printf("running in offline mode: %v", err)
		return client, false
	}
	log.Printf("connected to server at %s", cfg.Server.Address)
	return client, true
}

// runGame starts the game loop and handles cleanup.
func runGame(game *Game, connected bool, client *network.Client) {
	if err := ebiten.RunGame(game); err != nil {
		if connected {
			client.Disconnect()
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if connected {
		client.Disconnect()
	}
}

// setupWeatherParticles creates genre-appropriate weather particle emitters.
func setupWeatherParticles(sys *particles.System, genre string, seed int64) {
	if sys == nil {
		return
	}

	// Genre-specific ambient particles
	var weatherType string
	var intensity float64

	switch genre {
	case "fantasy":
		// Light dust motes in sunbeams
		weatherType = particles.TypeDust
		intensity = 0.3
	case "sci-fi":
		// Subtle atmospheric particles
		weatherType = particles.TypeDust
		intensity = 0.2
	case "horror":
		// Ash/fog wisps
		weatherType = particles.TypeAsh
		intensity = 0.5
	case "cyberpunk":
		// Rain is iconic for cyberpunk
		weatherType = particles.TypeRain
		intensity = 0.4
	case "post-apocalyptic":
		// Dust and ash
		weatherType = particles.TypeDust
		intensity = 0.6
	default:
		return
	}

	// Create weather emitters using preset
	preset := &particles.WeatherPreset{
		Type:      weatherType,
		Intensity: intensity,
		Direction: 1.57, // Straight down
	}
	emitters := particles.CreateWeatherEmitters(preset, seed)
	for _, e := range emitters {
		sys.AddEmitter(e)
	}
}

// startProfileServer starts the pprof HTTP server for runtime profiling.
func startProfileServer(port int) {
	addr := fmt.Sprintf("localhost:%d", port)
	log.Printf("Starting pprof server at http://%s/debug/pprof/", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("pprof server error: %v", err)
		}
	}()
}
