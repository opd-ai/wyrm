// Package systems implements ECS system behaviors.
package systems

import (
	"math"
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// NPCOccupationSystem handles NPC work behaviors and occupation-specific actions.
type NPCOccupationSystem struct {
	// Seed for deterministic randomness.
	Seed int64
	// rng is the random number generator.
	rng *rand.Rand
}

// NewNPCOccupationSystem creates a new occupation behavior system.
func NewNPCOccupationSystem(seed int64) *NPCOccupationSystem {
	return &NPCOccupationSystem{
		Seed: seed,
		rng:  rand.New(rand.NewSource(seed)),
	}
}

// Update processes all NPCs with occupations.
func (s *NPCOccupationSystem) Update(w *ecs.World, dt float64) {
	entities := w.Entities("NPCOccupation", "Schedule", "Position")
	for _, e := range entities {
		s.updateOccupation(w, e, dt)
	}
}

// updateOccupation handles a single NPC's occupation behavior.
func (s *NPCOccupationSystem) updateOccupation(w *ecs.World, e ecs.Entity, dt float64) {
	occComp, ok := w.GetComponent(e, "NPCOccupation")
	if !ok {
		return
	}
	occ := occComp.(*components.NPCOccupation)

	schedComp, ok := w.GetComponent(e, "Schedule")
	if !ok {
		return
	}
	sched := schedComp.(*components.Schedule)

	// Check if NPC should be working
	isWorkActivity := sched.CurrentActivity == "working" ||
		sched.CurrentActivity == "crafting" ||
		sched.CurrentActivity == "trading"

	if isWorkActivity {
		s.performWorkBehavior(w, e, occ, dt)
	} else if sched.CurrentActivity == "resting" || sched.CurrentActivity == "sleeping" {
		s.performRestBehavior(occ, dt)
	} else {
		occ.IsWorking = false
	}
}

// performWorkBehavior handles NPC work actions.
func (s *NPCOccupationSystem) performWorkBehavior(w *ecs.World, e ecs.Entity, occ *components.NPCOccupation, dt float64) {
	occ.IsWorking = true
	s.applyFatigue(occ, dt)
	efficiency := s.calculateEfficiency(occ)
	s.progressTask(w, e, occ, efficiency, dt)
	s.processCustomers(w, e, occ, dt)
}

// applyFatigue increases fatigue during work.
func (s *NPCOccupationSystem) applyFatigue(occ *components.NPCOccupation, dt float64) {
	occ.Fatigue += dt * OccupationFatigueRate
	if occ.Fatigue > 1.0 {
		occ.Fatigue = 1.0
	}
}

// calculateEfficiency computes current work efficiency based on fatigue.
func (s *NPCOccupationSystem) calculateEfficiency(occ *components.NPCOccupation) float64 {
	efficiency := occ.WorkEfficiency * (1.0 - occ.Fatigue*OccupationFatiguePenalty)
	if efficiency < OccupationMinEfficiency {
		efficiency = OccupationMinEfficiency
	}
	return efficiency
}

// progressTask updates current task progress and handles completion.
func (s *NPCOccupationSystem) progressTask(w *ecs.World, e ecs.Entity, occ *components.NPCOccupation, efficiency, dt float64) {
	if occ.CurrentTask == "" {
		occ.CurrentTask = s.selectTask(occ)
		occ.TaskProgress = 0
		occ.TaskDuration = s.getTaskDuration(occ.OccupationType, occ.CurrentTask)
	}

	if occ.TaskDuration > 0 {
		progressRate := (dt / occ.TaskDuration) * efficiency
		occ.TaskProgress += progressRate

		if occ.TaskProgress >= 1.0 {
			s.completeTask(w, e, occ)
			occ.CurrentTask = ""
			occ.TaskProgress = 0
		}
	}
}

// processCustomers handles customer queue for service occupations.
func (s *NPCOccupationSystem) processCustomers(w *ecs.World, e ecs.Entity, occ *components.NPCOccupation, dt float64) {
	if len(occ.CustomerQueue) > 0 && occ.CanTrade {
		s.processCustomerQueue(w, e, occ, dt)
	}
}

// performRestBehavior handles NPC rest and fatigue recovery.
func (s *NPCOccupationSystem) performRestBehavior(occ *components.NPCOccupation, dt float64) {
	occ.IsWorking = false

	// Recover fatigue while resting
	occ.Fatigue -= dt * OccupationFatigueRecoveryRate
	if occ.Fatigue < 0 {
		occ.Fatigue = 0
	}
}

// selectTask chooses an appropriate task based on occupation type.
func (s *NPCOccupationSystem) selectTask(occ *components.NPCOccupation) string {
	tasks := GetOccupationTasks(occ.OccupationType)
	if len(tasks) == 0 {
		return "idle"
	}
	return tasks[s.rng.Intn(len(tasks))]
}

// getTaskDuration returns the base duration for a task.
func (s *NPCOccupationSystem) getTaskDuration(occupationType, task string) float64 {
	baseDuration := OccupationTaskDurationBase
	multiplier := getTaskDurationMultiplier(occupationType, task)
	return baseDuration * multiplier
}

// getTaskDurationMultiplier returns the duration multiplier for a task.
func getTaskDurationMultiplier(occupationType, task string) float64 {
	multipliers := map[string]map[string]float64{
		OccupationMerchant: {
			"arranging_wares": 0.5,
			"counting_gold":   0.3,
			"haggling":        0.8,
		},
		OccupationBlacksmith: {
			"forging":    2.0,
			"sharpening": 0.7,
			"tempering":  1.5,
		},
		OccupationGuard: {
			"patrolling":     1.0,
			"standing_watch": 2.0,
			"inspecting":     0.5,
		},
		OccupationInnkeeper: {
			"cleaning": 0.8,
			"serving":  0.3,
			"cooking":  1.2,
		},
		OccupationHealer: {
			"treating_patient":   1.5,
			"preparing_medicine": 1.0,
			"studying":           2.0,
		},
		OccupationFarmer: {
			"planting":        1.2,
			"harvesting":      1.5,
			"tending_animals": 0.8,
		},
	}

	if occTasks, ok := multipliers[occupationType]; ok {
		if mult, ok := occTasks[task]; ok {
			return mult
		}
	}
	return 1.0
}

// completeTask handles task completion effects.
func (s *NPCOccupationSystem) completeTask(w *ecs.World, e ecs.Entity, occ *components.NPCOccupation) {
	// Gold earned is tracked but not stored in a simple inventory
	// This could be expanded to use a more detailed currency component
	_ = occ.GoldPerHour * (occ.TaskDuration / 3600.0) * (0.5 + occ.SkillLevel*0.5)

	// Crafting occupations may produce items
	if occ.CanCraft && (occ.CurrentTask == "forging" ||
		occ.CurrentTask == "cooking" || occ.CurrentTask == "preparing_medicine") {
		s.produceItem(occ)
	}

	// Improve skill slightly with practice
	occ.SkillLevel += OccupationSkillGainRate
	if occ.SkillLevel > 1.0 {
		occ.SkillLevel = 1.0
	}
}

// produceItem adds an item to the NPC's work inventory.
func (s *NPCOccupationSystem) produceItem(occ *components.NPCOccupation) {
	itemID := s.getProducedItem(occ.OccupationType, occ.CurrentTask)
	if itemID == "" {
		return
	}

	// Quality based on skill level with some randomness
	quality := occ.SkillLevel*0.7 + s.rng.Float64()*0.3

	// Check if item already exists in inventory
	for i := range occ.WorkInventory {
		if occ.WorkInventory[i].ItemID == itemID {
			occ.WorkInventory[i].Quantity++
			return
		}
	}

	// Add new item
	occ.WorkInventory = append(occ.WorkInventory, components.OccupationItem{
		ItemID:    itemID,
		Quantity:  1,
		BasePrice: s.getBasePrice(itemID),
		Quality:   quality,
	})
}

// getProducedItem returns the item ID produced by a task.
func (s *NPCOccupationSystem) getProducedItem(occupationType, task string) string {
	switch occupationType {
	case OccupationBlacksmith:
		switch task {
		case "forging":
			items := []string{"iron_sword", "iron_axe", "iron_hammer", "iron_nails"}
			return items[s.rng.Intn(len(items))]
		case "sharpening":
			return "" // Service, not product
		}
	case OccupationInnkeeper:
		if task == "cooking" {
			items := []string{"bread", "stew", "ale", "roast_meat"}
			return items[s.rng.Intn(len(items))]
		}
	case OccupationHealer:
		if task == "preparing_medicine" {
			items := []string{"health_potion", "antidote", "bandages", "salve"}
			return items[s.rng.Intn(len(items))]
		}
	case OccupationFarmer:
		if task == "harvesting" {
			items := []string{"wheat", "vegetables", "fruit", "herbs"}
			return items[s.rng.Intn(len(items))]
		}
	}
	return ""
}

// getBasePrice returns the base price for an item.
func (s *NPCOccupationSystem) getBasePrice(itemID string) float64 {
	prices := map[string]float64{
		"iron_sword":    50,
		"iron_axe":      40,
		"iron_hammer":   30,
		"iron_nails":    5,
		"bread":         2,
		"stew":          5,
		"ale":           3,
		"roast_meat":    8,
		"health_potion": 25,
		"antidote":      20,
		"bandages":      5,
		"salve":         10,
		"wheat":         3,
		"vegetables":    4,
		"fruit":         5,
		"herbs":         8,
	}
	if price, ok := prices[itemID]; ok {
		return price
	}
	return 10 // Default price
}

// processCustomerQueue handles customer service.
func (s *NPCOccupationSystem) processCustomerQueue(w *ecs.World, e ecs.Entity, occ *components.NPCOccupation, dt float64) {
	if len(occ.CustomerQueue) == 0 {
		return
	}

	// Service time based on efficiency
	serviceTime := OccupationServiceTimeBase / occ.WorkEfficiency

	// For now, just remove customers after service time
	// In a full implementation, this would trigger trade UI or auto-trade
	if occ.TaskProgress >= 1.0 && occ.CurrentTask == "serving_customer" {
		if len(occ.CustomerQueue) > 0 {
			occ.CustomerQueue = occ.CustomerQueue[1:]
		}
	}
	_ = serviceTime
}

// GetOccupationTasks returns tasks for an occupation type.
func GetOccupationTasks(occupationType string) []string {
	switch occupationType {
	case OccupationMerchant:
		return []string{"arranging_wares", "counting_gold", "haggling", "restocking"}
	case OccupationBlacksmith:
		return []string{"forging", "sharpening", "tempering", "repairing"}
	case OccupationGuard:
		return []string{"patrolling", "standing_watch", "inspecting", "training"}
	case OccupationInnkeeper:
		return []string{"cleaning", "serving", "cooking", "brewing"}
	case OccupationHealer:
		return []string{"treating_patient", "preparing_medicine", "studying", "praying"}
	case OccupationFarmer:
		return []string{"planting", "harvesting", "tending_animals", "repairing_fences"}
	case OccupationMiner:
		return []string{"mining", "sorting_ore", "reinforcing_tunnels", "resting"}
	case OccupationScribe:
		return []string{"copying", "translating", "cataloging", "researching"}
	case OccupationBard:
		return []string{"performing", "composing", "practicing", "socializing"}
	case OccupationPriest:
		return []string{"praying", "blessing", "counseling", "studying"}
	default:
		return []string{"working", "resting"}
	}
}

// GetGenreOccupations returns occupation types appropriate for a genre.
func GetGenreOccupations(genre string) []string {
	base := []string{
		OccupationMerchant, OccupationGuard, OccupationHealer,
	}

	switch genre {
	case "fantasy":
		return append(base, OccupationBlacksmith, OccupationInnkeeper, OccupationFarmer,
			OccupationScribe, OccupationBard, OccupationPriest, OccupationMiner)
	case "sci-fi":
		return []string{
			OccupationMerchant, OccupationGuard, OccupationHealer,
			OccupationTechnician, OccupationScientist, OccupationPilot,
			OccupationEngineer, OccupationMedic,
		}
	case "horror":
		return []string{
			OccupationMerchant, OccupationGuard, OccupationHealer,
			OccupationPriest, OccupationMortician, OccupationHunter,
			OccupationHerbalist, OccupationGravedigger,
		}
	case "cyberpunk":
		return []string{
			OccupationMerchant, OccupationGuard, OccupationMedic,
			OccupationHacker, OccupationFixer, OccupationBodyguard,
			OccupationStreetVendor, OccupationTechDealer,
		}
	case "post-apocalyptic":
		return []string{
			OccupationMerchant, OccupationGuard, OccupationHealer,
			OccupationScavenger, OccupationMechanic, OccupationFarmer,
			OccupationHunter, OccupationWaterMerchant,
		}
	}
	return base
}

// AddToCustomerQueue adds an entity to an NPC's customer queue.
func AddToCustomerQueue(occ *components.NPCOccupation, customerID uint64) {
	if occ.CustomerQueue == nil {
		occ.CustomerQueue = make([]uint64, 0)
	}
	occ.CustomerQueue = append(occ.CustomerQueue, customerID)
}

// RemoveFromCustomerQueue removes an entity from the queue.
func RemoveFromCustomerQueue(occ *components.NPCOccupation, customerID uint64) {
	for i, id := range occ.CustomerQueue {
		if id == customerID {
			occ.CustomerQueue = append(occ.CustomerQueue[:i], occ.CustomerQueue[i+1:]...)
			return
		}
	}
}

// GetNPCEfficiency returns the NPC's current work efficiency.
func GetNPCEfficiency(occ *components.NPCOccupation) float64 {
	efficiency := occ.WorkEfficiency * (1.0 - occ.Fatigue*OccupationFatiguePenalty)
	if efficiency < OccupationMinEfficiency {
		return OccupationMinEfficiency
	}
	return efficiency
}

// InitializeOccupation sets up default values for an occupation.
func InitializeOccupation(occ *components.NPCOccupation, occupationType, genre string) {
	occ.OccupationType = occupationType
	occ.SkillLevel = 0.3 + math.Floor(rand.Float64()*5)/10 // 0.3-0.8
	occ.WorkEfficiency = 1.0
	occ.Fatigue = 0
	occ.GoldPerHour = getBaseGoldPerHour(occupationType)

	// Set capabilities based on occupation
	switch occupationType {
	case OccupationMerchant, OccupationStreetVendor, OccupationWaterMerchant:
		occ.CanTrade = true
	case OccupationBlacksmith, OccupationFarmer, OccupationMiner:
		occ.CanCraft = true
		occ.CanTrade = true
	case OccupationHealer, OccupationMedic, OccupationHerbalist:
		occ.CanTrade = true
		occ.CanCraft = true
	case OccupationGuard, OccupationBodyguard:
		occ.CanProvideQuests = true
	case OccupationScribe, OccupationScientist:
		occ.CanTrain = true
		occ.TrainableSkills = []string{"lore", "research"}
	case OccupationBard:
		occ.CanTrain = true
		occ.TrainableSkills = []string{"speech", "persuasion"}
	case OccupationPriest:
		occ.CanTrain = true
		occ.CanProvideQuests = true
		occ.TrainableSkills = []string{"faith", "healing"}
	}

	occ.WorkInventory = make([]components.OccupationItem, 0)
	occ.CustomerQueue = make([]uint64, 0)
}

// getBaseGoldPerHour returns the base hourly wage for an occupation.
func getBaseGoldPerHour(occupationType string) float64 {
	wages := map[string]float64{
		OccupationMerchant:      10,
		OccupationBlacksmith:    15,
		OccupationGuard:         8,
		OccupationInnkeeper:     12,
		OccupationHealer:        20,
		OccupationFarmer:        5,
		OccupationMiner:         7,
		OccupationScribe:        12,
		OccupationBard:          10,
		OccupationPriest:        8,
		OccupationTechnician:    18,
		OccupationScientist:     25,
		OccupationPilot:         30,
		OccupationEngineer:      22,
		OccupationMedic:         20,
		OccupationMortician:     10,
		OccupationHunter:        12,
		OccupationHerbalist:     15,
		OccupationGravedigger:   6,
		OccupationHacker:        25,
		OccupationFixer:         20,
		OccupationBodyguard:     15,
		OccupationStreetVendor:  8,
		OccupationTechDealer:    18,
		OccupationScavenger:     10,
		OccupationMechanic:      15,
		OccupationWaterMerchant: 12,
	}
	if wage, ok := wages[occupationType]; ok {
		return wage
	}
	return 10
}
