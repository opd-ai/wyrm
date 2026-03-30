package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewPardonSystem(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	tests := []struct {
		name  string
		genre string
		seed  int64
	}{
		{"fantasy", "fantasy", 12345},
		{"cyberpunk", "cyberpunk", 67890},
		{"zero seed", "horror", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPardonSystem(cs, ces, tt.genre, tt.seed)
			if ps == nil {
				t.Fatal("NewPardonSystem returned nil")
			}
			if ps.genre != tt.genre {
				t.Errorf("genre = %q, want %q", ps.genre, tt.genre)
			}
			if ps.rngSeed != tt.seed {
				t.Errorf("rngSeed = %d, want %d", ps.rngSeed, tt.seed)
			}
			if len(ps.pardonRequirements) == 0 {
				t.Error("Pardon requirements not initialized")
			}
		})
	}
}

func TestPardonType_String(t *testing.T) {
	tests := []struct {
		pardonType PardonType
		want       string
	}{
		{PardonTypeFull, "Full Pardon"},
		{PardonTypePartial, "Partial Pardon"},
		{PardonTypeAmnesty, "General Amnesty"},
		{PardonTypeBribed, "Bribed Pardon"},
		{PardonTypeService, "Service Pardon"},
		{PardonTypePolitical, "Political Pardon"},
		{PardonTypeReligious, "Religious Absolution"},
		{PardonTypeMilitary, "Military Pardon"},
		{PardonType(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.pardonType.String()
		if got != tt.want {
			t.Errorf("PardonType(%d).String() = %q, want %q", tt.pardonType, got, tt.want)
		}
	}
}

func TestPardonSystem_CanObtainPardon(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	// Entity with crime
	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 2, BountyAmount: 500})

	// Should be able to get partial pardon
	canObtain, reason := ps.CanObtainPardon(w, entity, PardonTypePartial)
	if !canObtain {
		t.Errorf("Should be able to obtain partial pardon: %s", reason)
	}

	// Should be able to get full pardon (level 2 <= max 3)
	canObtain, _ = ps.CanObtainPardon(w, entity, PardonTypeFull)
	if !canObtain {
		t.Error("Should be able to obtain full pardon at level 2")
	}
}

func TestPardonSystem_CanObtainPardon_WantedTooHigh(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 5, BountyAmount: 2000})

	// Full pardon max is 3, should fail at level 5
	canObtain, reason := ps.CanObtainPardon(w, entity, PardonTypeFull)
	if canObtain {
		t.Error("Should not be able to obtain full pardon at level 5")
	}
	if reason != "Wanted level too high for this pardon" {
		t.Errorf("Wrong reason: %s", reason)
	}
}

func TestPardonSystem_CanObtainPardon_NoCrime(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	entity := w.CreateEntity()
	// No Crime component

	canObtain, reason := ps.CanObtainPardon(w, entity, PardonTypePartial)
	if canObtain {
		t.Error("Should not be able to obtain pardon without crime")
	}
	if reason != "No criminal record" {
		t.Errorf("Wrong reason: %s", reason)
	}
}

func TestPardonSystem_CanObtainPardon_NoWanted(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 0})

	canObtain, reason := ps.CanObtainPardon(w, entity, PardonTypePartial)
	if canObtain {
		t.Error("Should not be able to obtain pardon when not wanted")
	}
	if reason != "No crimes to pardon" {
		t.Errorf("Wrong reason: %s", reason)
	}
}

func TestPardonSystem_GrantPardon_Full(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{
		WantedLevel:  2,
		BountyAmount: 500,
		InJail:       true,
	})

	record, err := ps.GrantPardon(w, entity, PardonTypeFull, "King Arthur")
	if err != nil {
		t.Fatalf("GrantPardon failed: %v", err)
	}

	if record == nil {
		t.Fatal("Record is nil")
	}
	if record.Type != PardonTypeFull {
		t.Errorf("Type = %d, want Full", record.Type)
	}
	if record.GranterName != "King Arthur" {
		t.Errorf("GranterName = %q, want 'King Arthur'", record.GranterName)
	}
	if record.WasWantedLevel != 2 {
		t.Errorf("WasWantedLevel = %d, want 2", record.WasWantedLevel)
	}

	// Verify crime was cleared
	comp, _ := w.GetComponent(entity, "Crime")
	crime := comp.(*components.Crime)
	if crime.WantedLevel != 0 {
		t.Errorf("WantedLevel after pardon = %d, want 0", crime.WantedLevel)
	}
	if crime.BountyAmount != 0 {
		t.Errorf("BountyAmount after pardon = %v, want 0", crime.BountyAmount)
	}
	if crime.InJail {
		t.Error("Should be released from jail")
	}
}

func TestPardonSystem_GrantPardon_Partial(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)
	ps.PartialPardonReduction = 0.5

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{
		WantedLevel:  4,
		BountyAmount: 1000,
	})

	record, err := ps.GrantPardon(w, entity, PardonTypePartial, "Local Noble")
	if err != nil {
		t.Fatalf("GrantPardon failed: %v", err)
	}

	// Verify partial reduction
	comp, _ := w.GetComponent(entity, "Crime")
	crime := comp.(*components.Crime)
	if crime.WantedLevel != 2 { // 4 * (1 - 0.5) = 2
		t.Errorf("WantedLevel after partial = %d, want 2", crime.WantedLevel)
	}
	if crime.BountyAmount != 500 { // 1000 * (1 - 0.5) = 500
		t.Errorf("BountyAmount after partial = %v, want 500", crime.BountyAmount)
	}
	if record.Type != PardonTypePartial {
		t.Error("Record should be partial type")
	}
}

func TestPardonSystem_GrantPardon_CannotObtain(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 5}) // Too high for full pardon

	_, err := ps.GrantPardon(w, entity, PardonTypeFull, "King")
	if err == nil {
		t.Error("Should fail for wanted level too high")
	}
}

func TestPardonSystem_StartAmnestyEvent(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	event := ps.StartAmnestyEvent(
		"Victory Amnesty",
		"Celebrating the kingdom's victory",
		"capital_region",
		"",
		3600.0, // 1 hour
		3,      // Max level 3
		nil,    // All crimes
	)

	if event == nil {
		t.Fatal("StartAmnestyEvent returned nil")
	}
	if event.Name != "Victory Amnesty" {
		t.Errorf("Name = %q, want 'Victory Amnesty'", event.Name)
	}
	if !event.IsActive {
		t.Error("Event should be active")
	}
	if event.MaxWantedLevel != 3 {
		t.Errorf("MaxWantedLevel = %d, want 3", event.MaxWantedLevel)
	}

	// Verify in active list
	activeEvents := ps.GetActiveAmnestyEvents()
	if len(activeEvents) != 1 {
		t.Errorf("Expected 1 active event, got %d", len(activeEvents))
	}
}

func TestPardonSystem_ClaimAmnesty(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	// Start amnesty
	event := ps.StartAmnestyEvent("Test Amnesty", "Test", "", "", 3600.0, 3, nil)

	// Create criminal
	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 2, BountyAmount: 300})

	record, err := ps.ClaimAmnesty(w, entity, event.ID)
	if err != nil {
		t.Fatalf("ClaimAmnesty failed: %v", err)
	}
	if record == nil {
		t.Fatal("Record is nil")
	}
	if record.Type != PardonTypeAmnesty {
		t.Errorf("Type = %d, want Amnesty", record.Type)
	}

	// Verify crime cleared
	comp, _ := w.GetComponent(entity, "Crime")
	crime := comp.(*components.Crime)
	if crime.WantedLevel != 0 {
		t.Error("Wanted level should be cleared")
	}

	// Verify participation tracked
	if len(event.ParticipantIDs) != 1 {
		t.Errorf("Expected 1 participant, got %d", len(event.ParticipantIDs))
	}
}

func TestPardonSystem_ClaimAmnesty_AlreadyClaimed(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	event := ps.StartAmnestyEvent("Test Amnesty", "Test", "", "", 3600.0, 3, nil)

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 2})

	// Claim once
	ps.ClaimAmnesty(w, entity, event.ID)

	// Add crime again
	comp, _ := w.GetComponent(entity, "Crime")
	comp.(*components.Crime).WantedLevel = 2

	// Try to claim again
	_, err := ps.ClaimAmnesty(w, entity, event.ID)
	if err == nil {
		t.Error("Should not be able to claim amnesty twice")
	}
}

func TestPardonSystem_ClaimAmnesty_WantedTooHigh(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	event := ps.StartAmnestyEvent("Test Amnesty", "Test", "", "", 3600.0, 2, nil) // Max level 2

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 3}) // Level 3, too high

	_, err := ps.ClaimAmnesty(w, entity, event.ID)
	if err == nil {
		t.Error("Should fail for wanted level too high")
	}
}

func TestPardonSystem_EndAmnestyEvent(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	event := ps.StartAmnestyEvent("Test Amnesty", "Test", "", "", 3600.0, 3, nil)

	result := ps.EndAmnestyEvent(event.ID)
	if !result {
		t.Error("EndAmnestyEvent returned false")
	}
	if event.IsActive {
		t.Error("Event should not be active after ending")
	}

	activeEvents := ps.GetActiveAmnestyEvents()
	if len(activeEvents) != 0 {
		t.Error("Should have no active events")
	}
}

func TestPardonSystem_Update_AmnestyExpiration(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	event := ps.StartAmnestyEvent("Test Amnesty", "Test", "", "", 100.0, 3, nil) // 100 second duration

	// Simulate time passing beyond duration
	ps.Update(w, 150.0)

	if event.IsActive {
		t.Error("Event should expire after duration")
	}

	activeEvents := ps.GetActiveAmnestyEvents()
	if len(activeEvents) != 0 {
		t.Error("Should have no active events after expiration")
	}
}

func TestPardonSystem_GetEntityPardons(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	entity := w.CreateEntity()

	// No pardons initially
	pardons := ps.GetEntityPardons(uint64(entity))
	if len(pardons) != 0 {
		t.Error("Should have no pardons initially")
	}

	// Grant a pardon
	w.AddComponent(entity, &components.Crime{WantedLevel: 2})
	ps.GrantPardon(w, entity, PardonTypePartial, "Noble")

	// Re-add crime and get another pardon
	comp, _ := w.GetComponent(entity, "Crime")
	comp.(*components.Crime).WantedLevel = 2
	ps.GrantPardon(w, entity, PardonTypeBribed, "Corrupt Official")

	pardons = ps.GetEntityPardons(uint64(entity))
	if len(pardons) != 2 {
		t.Errorf("Expected 2 pardons, got %d", len(pardons))
	}
}

func TestPardonSystem_GetAvailablePardons(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 2})

	available := ps.GetAvailablePardons(w, entity)
	if len(available) == 0 {
		t.Error("Should have some available pardons")
	}

	// Check that partial is available (it has no special requirements)
	foundPartial := false
	for _, pt := range available {
		if pt == PardonTypePartial {
			foundPartial = true
			break
		}
	}
	if !foundPartial {
		t.Error("Partial pardon should be available")
	}
}

func TestPardonSystem_CalculatePardonCost(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 3})

	// Partial pardon base cost is 1000
	cost := ps.CalculatePardonCost(w, entity, PardonTypePartial)
	// 1000 * (1 + 3 * 0.2) = 1000 * 1.6 = 1600
	if cost != 1600 {
		t.Errorf("Cost = %d, want 1600", cost)
	}
}

func TestPardonSystem_GetPardonDescription_Genres(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			ps := NewPardonSystem(cs, ces, genre, 12345)
			desc := ps.GetPardonDescription(PardonTypeFull)
			if desc == "" {
				t.Errorf("Empty description for genre %s", genre)
			}
		})
	}
}

func TestPardonSystem_IsEligibleForAmnesty(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	event := ps.StartAmnestyEvent("Test Amnesty", "Test", "", "", 3600.0, 3, nil)

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Crime{WantedLevel: 2})

	eligible, reason := ps.IsEligibleForAmnesty(w, entity, event.ID)
	if !eligible {
		t.Errorf("Should be eligible: %s", reason)
	}

	// Make ineligible by high wanted level
	comp, _ := w.GetComponent(entity, "Crime")
	comp.(*components.Crime).WantedLevel = 5

	eligible, _ = ps.IsEligibleForAmnesty(w, entity, event.ID)
	if eligible {
		t.Error("Should not be eligible at level 5")
	}
}

func TestFormatPardonID(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "PD-0"},
		{1, "PD-1"},
		{123, "PD-123"},
		{99999, "PD-99999"},
	}

	for _, tt := range tests {
		got := formatPardonID(tt.n)
		if got != tt.want {
			t.Errorf("formatPardonID(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatAmnestyID(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "AM-0"},
		{1, "AM-1"},
		{123, "AM-123"},
		{99999, "AM-99999"},
	}

	for _, tt := range tests {
		got := formatAmnestyID(tt.n)
		if got != tt.want {
			t.Errorf("formatAmnestyID(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestPardonSystem_GetPardonRequirement(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ps := NewPardonSystem(cs, ces, "fantasy", 12345)

	req := ps.GetPardonRequirement(PardonTypeFull)
	if req == nil {
		t.Fatal("Requirement is nil")
	}
	if req.Type != PardonTypeFull {
		t.Errorf("Type = %d, want Full", req.Type)
	}
	if req.GoldCost <= 0 {
		t.Error("GoldCost should be positive")
	}
	if req.Description == "" {
		t.Error("Description should not be empty")
	}
}
