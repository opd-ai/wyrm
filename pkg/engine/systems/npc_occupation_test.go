package systems

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewNPCOccupationSystem(t *testing.T) {
	system := NewNPCOccupationSystem(12345)
	if system == nil {
		t.Fatal("NewNPCOccupationSystem returned nil")
	}
	if system.Seed != 12345 {
		t.Errorf("Seed = %d, want 12345", system.Seed)
	}
	if system.rng == nil {
		t.Error("rng should be initialized")
	}
}

func TestNPCOccupationComponent(t *testing.T) {
	occ := &components.NPCOccupation{
		OccupationType: OccupationMerchant,
		SkillLevel:     0.5,
	}

	if occ.Type() != "NPCOccupation" {
		t.Errorf("Type() = %v, want NPCOccupation", occ.Type())
	}
}

func TestInitializeOccupation(t *testing.T) {
	occ := &components.NPCOccupation{}
	rng := rand.New(rand.NewSource(12345))
	InitializeOccupation(occ, OccupationMerchant, "fantasy", rng)

	if occ.OccupationType != OccupationMerchant {
		t.Errorf("OccupationType = %v, want %v", occ.OccupationType, OccupationMerchant)
	}
	if occ.SkillLevel < 0.3 || occ.SkillLevel > 0.8 {
		t.Errorf("SkillLevel = %v, want 0.3-0.8", occ.SkillLevel)
	}
	if !occ.CanTrade {
		t.Error("Merchant should be able to trade")
	}
	if occ.WorkEfficiency != 1.0 {
		t.Errorf("WorkEfficiency = %v, want 1.0", occ.WorkEfficiency)
	}
}

func TestGetOccupationTasks(t *testing.T) {
	tests := []struct {
		occupation string
		wantTasks  int
	}{
		{OccupationMerchant, 4},
		{OccupationBlacksmith, 4},
		{OccupationGuard, 4},
		{OccupationInnkeeper, 4},
		{OccupationHealer, 4},
		{OccupationFarmer, 4},
		{"unknown", 2}, // default tasks
	}

	for _, tt := range tests {
		t.Run(tt.occupation, func(t *testing.T) {
			tasks := GetOccupationTasks(tt.occupation)
			if len(tasks) != tt.wantTasks {
				t.Errorf("GetOccupationTasks(%v) = %d tasks, want %d", tt.occupation, len(tasks), tt.wantTasks)
			}
		})
	}
}

func TestGetGenreOccupations(t *testing.T) {
	tests := []struct {
		genre    string
		wantMin  int
		checkHas string
	}{
		{"fantasy", 10, OccupationBlacksmith},
		{"sci-fi", 8, OccupationTechnician},
		{"horror", 8, OccupationMortician},
		{"cyberpunk", 8, OccupationHacker},
		{"post-apocalyptic", 8, OccupationScavenger},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			occs := GetGenreOccupations(tt.genre)
			if len(occs) < tt.wantMin {
				t.Errorf("GetGenreOccupations(%v) = %d occupations, want >= %d", tt.genre, len(occs), tt.wantMin)
			}
			found := false
			for _, o := range occs {
				if o == tt.checkHas {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Genre %v should have occupation %v", tt.genre, tt.checkHas)
			}
		})
	}
}

func TestCustomerQueue(t *testing.T) {
	occ := &components.NPCOccupation{}

	// Add customers
	AddToCustomerQueue(occ, 100)
	AddToCustomerQueue(occ, 200)
	AddToCustomerQueue(occ, 300)

	if len(occ.CustomerQueue) != 3 {
		t.Errorf("Queue length = %d, want 3", len(occ.CustomerQueue))
	}
	if occ.CustomerQueue[0] != 100 {
		t.Errorf("First customer = %d, want 100", occ.CustomerQueue[0])
	}

	// Remove middle customer
	RemoveFromCustomerQueue(occ, 200)
	if len(occ.CustomerQueue) != 2 {
		t.Errorf("Queue length after removal = %d, want 2", len(occ.CustomerQueue))
	}
	for _, id := range occ.CustomerQueue {
		if id == 200 {
			t.Error("Customer 200 should be removed")
		}
	}
}

func TestGetNPCEfficiency(t *testing.T) {
	tests := []struct {
		name       string
		efficiency float64
		fatigue    float64
		wantMin    float64
		wantMax    float64
	}{
		{"fresh", 1.0, 0.0, 0.99, 1.01},
		{"tired", 1.0, 0.5, 0.74, 0.76},
		{"exhausted", 1.0, 1.0, 0.49, 0.51},
		{"minimum_clamp", 1.0, 2.0, OccupationMinEfficiency - 0.01, OccupationMinEfficiency + 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			occ := &components.NPCOccupation{
				WorkEfficiency: tt.efficiency,
				Fatigue:        tt.fatigue,
			}
			eff := GetNPCEfficiency(occ)
			if eff < tt.wantMin || eff > tt.wantMax {
				t.Errorf("GetNPCEfficiency = %v, want %v-%v", eff, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestOccupationSystemUpdate(t *testing.T) {
	system := NewNPCOccupationSystem(12345)
	world := ecs.NewWorld()

	// Create an NPC with occupation and schedule
	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0})
	world.AddComponent(npc, &components.Schedule{
		CurrentActivity: "working",
		TimeSlots:       map[int]string{9: "working"},
	})
	occ := &components.NPCOccupation{
		OccupationType: OccupationMerchant,
		SkillLevel:     0.5,
		WorkEfficiency: 1.0,
	}
	world.AddComponent(npc, occ)

	// Run update
	system.Update(world, 1.0)

	// Check NPC is working
	occComp, _ := world.GetComponent(npc, "NPCOccupation")
	occResult := occComp.(*components.NPCOccupation)
	if !occResult.IsWorking {
		t.Error("NPC should be working")
	}
	if occResult.CurrentTask == "" {
		t.Error("NPC should have a task")
	}
	if occResult.Fatigue <= 0 {
		t.Error("Fatigue should accumulate while working")
	}
}

func TestOccupationRestRecovery(t *testing.T) {
	system := NewNPCOccupationSystem(12345)
	world := ecs.NewWorld()

	// Create an NPC that is resting
	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0})
	world.AddComponent(npc, &components.Schedule{
		CurrentActivity: "resting",
		TimeSlots:       map[int]string{22: "resting"},
	})
	occ := &components.NPCOccupation{
		OccupationType: OccupationMerchant,
		SkillLevel:     0.5,
		WorkEfficiency: 1.0,
		Fatigue:        0.5, // Start fatigued
	}
	world.AddComponent(npc, occ)

	// Run update
	system.Update(world, 10.0) // 10 seconds of rest

	// Check fatigue recovered
	occComp, _ := world.GetComponent(npc, "NPCOccupation")
	occResult := occComp.(*components.NPCOccupation)
	if occResult.Fatigue >= 0.5 {
		t.Errorf("Fatigue should decrease while resting, got %v", occResult.Fatigue)
	}
	if occResult.IsWorking {
		t.Error("NPC should not be working while resting")
	}
}

func TestTaskCompletion(t *testing.T) {
	system := NewNPCOccupationSystem(12345)
	world := ecs.NewWorld()

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0})
	world.AddComponent(npc, &components.Schedule{
		CurrentActivity: "working",
		TimeSlots:       map[int]string{9: "working"},
	})
	occ := &components.NPCOccupation{
		OccupationType: OccupationMerchant,
		SkillLevel:     0.5,
		WorkEfficiency: 1.0,
		CurrentTask:    "counting_gold",
		TaskProgress:   0.9,
		TaskDuration:   10.0,
	}
	world.AddComponent(npc, occ)

	initialSkill := occ.SkillLevel

	// Run enough updates to complete task
	for i := 0; i < 20; i++ {
		system.Update(world, 1.0)
	}

	// Check task completed and new one started
	occComp, _ := world.GetComponent(npc, "NPCOccupation")
	occResult := occComp.(*components.NPCOccupation)

	// Skill should have improved
	if occResult.SkillLevel <= initialSkill {
		t.Error("Skill should improve after completing tasks")
	}
}

func TestOccupationCapabilities(t *testing.T) {
	tests := []struct {
		occupation string
		canTrade   bool
		canCraft   bool
	}{
		{OccupationMerchant, true, false},
		{OccupationBlacksmith, true, true},
		{OccupationHealer, true, true},
		{OccupationGuard, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.occupation, func(t *testing.T) {
			occ := &components.NPCOccupation{}
			InitializeOccupation(occ, tt.occupation, "fantasy", nil)
			if occ.CanTrade != tt.canTrade {
				t.Errorf("%v CanTrade = %v, want %v", tt.occupation, occ.CanTrade, tt.canTrade)
			}
			if occ.CanCraft != tt.canCraft {
				t.Errorf("%v CanCraft = %v, want %v", tt.occupation, occ.CanCraft, tt.canCraft)
			}
		})
	}
}

func TestDeterministicBehavior(t *testing.T) {
	seed := int64(99999)

	// Run twice with same seed
	var tasks1, tasks2 []string
	for i := 0; i < 2; i++ {
		system := NewNPCOccupationSystem(seed)
		occ := &components.NPCOccupation{
			OccupationType: OccupationMerchant,
		}
		var tasks []string
		for j := 0; j < 10; j++ {
			task := system.selectTask(occ)
			tasks = append(tasks, task)
		}
		if i == 0 {
			tasks1 = tasks
		} else {
			tasks2 = tasks
		}
	}

	// Tasks should be identical
	for i := range tasks1 {
		if tasks1[i] != tasks2[i] {
			t.Errorf("Task %d differs: %v vs %v", i, tasks1[i], tasks2[i])
		}
	}
}
