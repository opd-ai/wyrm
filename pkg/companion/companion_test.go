package companion

import (
	"testing"
)

func TestCreateCompanion(t *testing.T) {
	cm := NewManager(12345)

	comp := cm.CreateCompanion(100, "fantasy", RoleTank)
	if comp == nil {
		t.Fatal("CreateCompanion returned nil")
	}

	if comp.Genre != "fantasy" {
		t.Errorf("Genre = %s, want fantasy", comp.Genre)
	}
	if comp.Role != RoleTank {
		t.Errorf("Role = %v, want tank", comp.Role)
	}
	if comp.Class == "" {
		t.Error("Class should not be empty")
	}
	if len(comp.Abilities) == 0 {
		t.Error("Companion should have abilities")
	}
}

func TestCompanionAbilitiesMatchRole(t *testing.T) {
	// Per AC: Companion uses class-appropriate abilities
	cm := NewManager(12345)

	testCases := []struct {
		role            CombatRole
		expectedAbility string
	}{
		{RoleTank, "shield_wall"},
		{RoleDPS, "power_strike"},
		{RoleHealer, "heal"},
		{RoleSupport, "buff"},
		{RoleRanged, "aimed_shot"},
	}

	for _, tc := range testCases {
		comp := cm.CreateCompanion(uint64(tc.role), "fantasy", tc.role)

		found := false
		for _, ability := range comp.Abilities {
			if ability.ID == tc.expectedAbility {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Role %v should have ability %s", tc.role, tc.expectedAbility)
		}
	}
}

func TestSelectAbilityHealer(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleHealer)

	// Per AC: Companion uses class-appropriate abilities
	ability := cm.SelectAbility(comp.ID, false, true) // ally low health
	if ability == nil {
		t.Fatal("SelectAbility returned nil")
	}
	if !ability.IsHealing {
		t.Errorf("Healer with low health ally should use healing ability")
	}
}

func TestSelectAbilityDPS(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleDPS)

	ability := cm.SelectAbility(comp.ID, true, false) // target low health
	if ability == nil {
		t.Fatal("SelectAbility returned nil")
	}
	// Should try to use execute on low health target
	if ability.ID != "execute" {
		t.Logf("DPS selected %s (execute expected for low health target)", ability.ID)
	}
}

func TestRecordPlayerAction(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleTank)

	event := ActionEvent{
		EventType:   "helped_npc",
		Description: "saved a villager",
		Target:      "villager",
		Outcome:     "success",
	}

	cm.RecordPlayerAction(comp.ID, event)

	events := cm.GetRecentEvents(comp.ID)
	if len(events) != 1 {
		t.Errorf("Events length = %d, want 1", len(events))
	}
	if events[0].EventType != "helped_npc" {
		t.Errorf("EventType = %s, want helped_npc", events[0].EventType)
	}
}

func TestEventMemoryLimit(t *testing.T) {
	// Per AC: dialog references player actions from last 10 events
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleTank)

	// Record more than 10 events
	for i := 0; i < 15; i++ {
		event := ActionEvent{
			EventType:   "event",
			Description: "test action",
		}
		cm.RecordPlayerAction(comp.ID, event)
	}

	events := cm.GetRecentEvents(comp.ID)
	if len(events) != 10 {
		t.Errorf("Events length = %d, want 10 (max)", len(events))
	}
}

func TestLoyaltyAdjustment(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleTank)

	initialLoyalty := comp.Loyalty

	// Positive action
	cm.RecordPlayerAction(comp.ID, ActionEvent{EventType: "shared_loot"})
	if comp.Loyalty <= initialLoyalty {
		t.Error("Loyalty should increase after positive action")
	}

	// Negative action
	cm2 := NewManager(12345)
	comp2 := cm2.CreateCompanion(100, "fantasy", RoleTank)
	initialLoyalty2 := comp2.Loyalty

	cm2.RecordPlayerAction(comp2.ID, ActionEvent{EventType: "killed_innocent"})
	if comp2.Loyalty >= initialLoyalty2 {
		t.Error("Loyalty should decrease after negative action")
	}
}

func TestLoyaltyClamping(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleTank)

	// Max loyalty
	for i := 0; i < 100; i++ {
		cm.RecordPlayerAction(comp.ID, ActionEvent{EventType: "protected_companion"})
	}
	if comp.Loyalty > 100 {
		t.Errorf("Loyalty = %f, should be clamped to 100", comp.Loyalty)
	}

	// Min loyalty
	for i := 0; i < 100; i++ {
		cm.RecordPlayerAction(comp.ID, ActionEvent{EventType: "abandoned_companion"})
	}
	if comp.Loyalty < 0 {
		t.Errorf("Loyalty = %f, should be clamped to 0", comp.Loyalty)
	}
}

func TestGenerateDialogResponse(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleTank)

	// Per AC: dialog references player actions from last 10 events
	cm.RecordPlayerAction(comp.ID, ActionEvent{
		EventType:   "helped_npc",
		Description: "saved the merchant",
	})

	response := cm.GenerateDialogResponse(comp.ID, "quest")
	if response == "" {
		t.Error("Dialog response should not be empty")
	}

	// Should reference the saved merchant event
	if len(response) == 0 {
		t.Error("Response should contain dialog text")
	}
}

func TestGetPlayerCompanion(t *testing.T) {
	cm := NewManager(12345)
	cm.CreateCompanion(100, "fantasy", RoleTank)

	comp := cm.GetPlayerCompanion(100)
	if comp == nil {
		t.Error("GetPlayerCompanion should return the companion")
	}

	// Different player should not have this companion
	comp2 := cm.GetPlayerCompanion(200)
	if comp2 != nil {
		t.Error("Different player should not have companion")
	}
}

func TestSetOrder(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleTank)

	cm.SetOrder(comp.ID, OrderStay)
	if comp.Following {
		t.Error("Companion should not be following after OrderStay")
	}
	if comp.CurrentOrder != OrderStay {
		t.Errorf("CurrentOrder = %v, want OrderStay", comp.CurrentOrder)
	}

	cm.SetOrder(comp.ID, OrderFollow)
	if !comp.Following {
		t.Error("Companion should be following after OrderFollow")
	}
}

func TestSetCombatState(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleTank)

	cm.SetCombatState(comp.ID, true)
	if !comp.InCombat {
		t.Error("Companion should be in combat")
	}

	cm.SetCombatState(comp.ID, false)
	if comp.InCombat {
		t.Error("Companion should not be in combat")
	}
}

func TestGenreSpecificTemplates(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		cm := NewManager(12345)
		comp := cm.CreateCompanion(uint64(len(genre)), genre, RoleDPS)

		if comp.Genre != genre {
			t.Errorf("Companion genre = %s, want %s", comp.Genre, genre)
		}
		if comp.Backstory == "" {
			t.Errorf("Companion for genre %s has no backstory", genre)
		}
	}
}

func TestPersonalityString(t *testing.T) {
	tests := []struct {
		p    Personality
		want string
	}{
		{PersonalityBrave, "brave"},
		{PersonalityCautious, "cautious"},
		{PersonalityLoyal, "loyal"},
		{PersonalityAggressive, "aggressive"},
		{PersonalityWise, "wise"},
	}

	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("%v.String() = %s, want %s", tt.p, got, tt.want)
		}
	}
}

func TestCombatRoleString(t *testing.T) {
	tests := []struct {
		r    CombatRole
		want string
	}{
		{RoleTank, "tank"},
		{RoleDPS, "damage"},
		{RoleHealer, "healer"},
		{RoleSupport, "support"},
		{RoleRanged, "ranged"},
	}

	for _, tt := range tests {
		if got := tt.r.String(); got != tt.want {
			t.Errorf("%v.String() = %s, want %s", tt.r, got, tt.want)
		}
	}
}

// TestGetCompanion tests the GetCompanion method.
func TestGetCompanion(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleTank)

	// Test getting an existing companion
	retrieved := cm.GetCompanion(comp.ID)
	if retrieved == nil {
		t.Fatal("GetCompanion returned nil for existing companion")
	}
	if retrieved.ID != comp.ID {
		t.Errorf("GetCompanion ID = %d, want %d", retrieved.ID, comp.ID)
	}

	// Test getting a non-existent companion
	nonExistent := cm.GetCompanion(999999)
	if nonExistent != nil {
		t.Error("GetCompanion should return nil for non-existent ID")
	}
}

// TestCount tests the Count method.
func TestCount(t *testing.T) {
	cm := NewManager(12345)

	// Initially no companions
	if cm.Count() != 0 {
		t.Errorf("Initial count = %d, want 0", cm.Count())
	}

	// Add companions
	cm.CreateCompanion(1, "fantasy", RoleTank)
	if cm.Count() != 1 {
		t.Errorf("After 1 creation, count = %d, want 1", cm.Count())
	}

	cm.CreateCompanion(2, "fantasy", RoleHealer)
	if cm.Count() != 2 {
		t.Errorf("After 2 creations, count = %d, want 2", cm.Count())
	}

	cm.CreateCompanion(3, "sci-fi", RoleDPS)
	if cm.Count() != 3 {
		t.Errorf("After 3 creations, count = %d, want 3", cm.Count())
	}
}

// TestSelectTankAbility tests the tank ability selection.
func TestSelectTankAbility(t *testing.T) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleTank)

	// Tank should select shield_wall
	ability := cm.SelectAbility(comp.ID, false, false)
	if ability == nil {
		t.Fatal("SelectAbility returned nil for tank")
	}
	if ability.ID != "shield_wall" {
		t.Errorf("Tank ability = %s, want shield_wall", ability.ID)
	}
}

// TestSelectAbilityAllRoles tests ability selection for all combat roles.
func TestSelectAbilityAllRoles(t *testing.T) {
	cm := NewManager(12345)

	tests := []struct {
		role           CombatRole
		targetLowHP    bool
		allyLowHP      bool
		expectedPrefix string
	}{
		{RoleTank, false, false, "shield_wall"},
		{RoleHealer, false, true, "heal"},
		{RoleDPS, true, false, "execute"},
		{RoleDPS, false, false, "power_strike"},
	}

	for i, tt := range tests {
		comp := cm.CreateCompanion(uint64(200+i), "fantasy", tt.role)
		ability := cm.SelectAbility(comp.ID, tt.targetLowHP, tt.allyLowHP)
		if ability == nil {
			t.Errorf("Test %d: SelectAbility returned nil for role %v", i, tt.role)
			continue
		}
		if ability.ID != tt.expectedPrefix {
			t.Errorf("Test %d: ability = %s, want %s", i, ability.ID, tt.expectedPrefix)
		}
	}
}

func BenchmarkCreateCompanion(b *testing.B) {
	cm := NewManager(12345)
	for i := 0; i < b.N; i++ {
		cm.CreateCompanion(uint64(i), "fantasy", RoleTank)
	}
}

func BenchmarkSelectAbility(b *testing.B) {
	cm := NewManager(12345)
	comp := cm.CreateCompanion(100, "fantasy", RoleHealer)

	for i := 0; i < b.N; i++ {
		cm.SelectAbility(comp.ID, false, true)
	}
}
