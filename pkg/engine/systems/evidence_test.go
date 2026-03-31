package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewCrimeEvidenceSystem(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)

	tests := []struct {
		name  string
		genre string
		seed  int64
	}{
		{"fantasy genre", "fantasy", 12345},
		{"sci-fi genre", "sci-fi", 67890},
		{"zero seed", "horror", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ces := NewCrimeEvidenceSystem(cs, tt.genre, tt.seed)
			if ces == nil {
				t.Fatal("NewCrimeEvidenceSystem returned nil")
			}
			if ces.crimeSystem != cs {
				t.Error("crimeSystem not set correctly")
			}
			if ces.genre != tt.genre {
				t.Errorf("genre = %q, want %q", ces.genre, tt.genre)
			}
			if ces.MinConvictionScore != 0.75 {
				t.Errorf("MinConvictionScore = %v, want 0.75", ces.MinConvictionScore)
			}
		})
	}
}

func TestEvidenceType_String(t *testing.T) {
	tests := []struct {
		evidenceType EvidenceType
		want         string
	}{
		{EvidenceTypeFingerprint, "Fingerprint"},
		{EvidenceTypeWeapon, "Weapon"},
		{EvidenceTypeWitness, "Witness Testimony"},
		{EvidenceTypeFootprint, "Footprint"},
		{EvidenceTypeBlood, "Blood Sample"},
		{EvidenceTypeDocument, "Document"},
		{EvidenceTypeStolenGoods, "Stolen Goods"},
		{EvidenceTypeMagical, "Magical Trace"},
		{EvidenceTypeDigital, "Digital Record"},
		{EvidenceType(99), "Unknown Evidence"},
	}

	for _, tt := range tests {
		got := tt.evidenceType.String()
		if got != tt.want {
			t.Errorf("EvidenceType(%d).String() = %q, want %q", tt.evidenceType, got, tt.want)
		}
	}
}

func TestEvidenceType_EvidenceStrength(t *testing.T) {
	tests := []struct {
		evidenceType EvidenceType
		want         float64
	}{
		{EvidenceTypeFingerprint, 0.7},
		{EvidenceTypeWeapon, 0.8},
		{EvidenceTypeWitness, 0.5},
		{EvidenceTypeFootprint, 0.3},
		{EvidenceTypeBlood, 0.85},
		{EvidenceTypeDocument, 0.6},
		{EvidenceTypeStolenGoods, 0.9},
		{EvidenceTypeMagical, 0.75},
		{EvidenceTypeDigital, 0.95},
		{EvidenceType(99), 0.1},
	}

	for _, tt := range tests {
		got := tt.evidenceType.EvidenceStrength()
		if got != tt.want {
			t.Errorf("EvidenceType(%d).EvidenceStrength() = %v, want %v", tt.evidenceType, got, tt.want)
		}
	}
}

func TestEvidence_GetConvictionStrength(t *testing.T) {
	tests := []struct {
		name         string
		evidence     Evidence
		wantStrength float64
	}{
		{
			"full quality processed",
			Evidence{Type: EvidenceTypeBlood, Quality: 1.0, IsProcessed: true},
			0.85, // 0.85 * 1.0 * 1.0
		},
		{
			"half quality processed",
			Evidence{Type: EvidenceTypeBlood, Quality: 0.5, IsProcessed: true},
			0.425, // 0.85 * 0.5
		},
		{
			"full quality unprocessed",
			Evidence{Type: EvidenceTypeBlood, Quality: 1.0, IsProcessed: false},
			0.425, // 0.85 * 1.0 * 0.5
		},
		{
			"tampered evidence",
			Evidence{Type: EvidenceTypeBlood, Quality: 1.0, IsProcessed: true, IsTampered: true},
			0.085, // 0.85 * 1.0 * 0.1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.evidence.GetConvictionStrength()
			diff := got - tt.wantStrength
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("GetConvictionStrength() = %v, want %v", got, tt.wantStrength)
			}
		})
	}
}

func TestCrimeEvidenceSystem_RecordCrime(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("murder", 100, 200, 10.0, 20.0, 5)

	if crimeID == "" {
		t.Fatal("RecordCrime returned empty ID")
	}

	record := ces.GetCrimeRecord(crimeID)
	if record == nil {
		t.Fatal("GetCrimeRecord returned nil")
	}

	if record.CrimeType != "murder" {
		t.Errorf("CrimeType = %q, want %q", record.CrimeType, "murder")
	}
	if record.CriminalID != 100 {
		t.Errorf("CriminalID = %d, want 100", record.CriminalID)
	}
	if record.VictimID != 200 {
		t.Errorf("VictimID = %d, want 200", record.VictimID)
	}
	if record.LocationX != 10.0 {
		t.Errorf("LocationX = %v, want 10.0", record.LocationX)
	}
	if record.Severity != 5 {
		t.Errorf("Severity = %d, want 5", record.Severity)
	}
}

func TestCrimeEvidenceSystem_RecordCrime_SeverityClamp(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	// Test severity too low
	crimeID1 := ces.RecordCrime("theft", 100, 0, 0, 0, -5)
	record1 := ces.GetCrimeRecord(crimeID1)
	if record1.Severity != 1 {
		t.Errorf("Severity clamped low: got %d, want 1", record1.Severity)
	}

	// Test severity too high
	crimeID2 := ces.RecordCrime("theft", 100, 0, 0, 0, 99)
	record2 := ces.GetCrimeRecord(crimeID2)
	if record2.Severity != 5 {
		t.Errorf("Severity clamped high: got %d, want 5", record2.Severity)
	}
}

func TestCrimeEvidenceSystem_CreateEvidence(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 10.0, 20.0, 2)

	evidence := ces.CreateEvidence(crimeID, EvidenceTypeFingerprint, 100, 10.0, 20.0, 0.8)
	if evidence == nil {
		t.Fatal("CreateEvidence returned nil")
	}

	if evidence.Type != EvidenceTypeFingerprint {
		t.Errorf("Type = %v, want %v", evidence.Type, EvidenceTypeFingerprint)
	}
	if evidence.CrimeID != crimeID {
		t.Errorf("CrimeID = %q, want %q", evidence.CrimeID, crimeID)
	}
	if evidence.Quality != 0.8 {
		t.Errorf("Quality = %v, want 0.8", evidence.Quality)
	}
	if evidence.SuspectID != 100 {
		t.Errorf("SuspectID = %d, want 100", evidence.SuspectID)
	}

	// Verify linked to crime
	record := ces.GetCrimeRecord(crimeID)
	if len(record.EvidenceIDs) != 1 {
		t.Errorf("Crime has %d evidence IDs, want 1", len(record.EvidenceIDs))
	}
}

func TestCrimeEvidenceSystem_CreateEvidence_InvalidCrime(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	evidence := ces.CreateEvidence("invalid-crime", EvidenceTypeFingerprint, 100, 0, 0, 0.8)
	if evidence != nil {
		t.Error("CreateEvidence with invalid crime ID should return nil")
	}
}

func TestCrimeEvidenceSystem_ReportCrime(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 0, 0, 2)

	// Initially not reported
	record := ces.GetCrimeRecord(crimeID)
	if record.IsReported {
		t.Error("Crime should not be reported initially")
	}

	// Report crime
	result := ces.ReportCrime(crimeID)
	if !result {
		t.Error("ReportCrime returned false")
	}
	if !record.IsReported {
		t.Error("Crime should be reported after ReportCrime")
	}

	// Invalid crime
	result = ces.ReportCrime("invalid")
	if result {
		t.Error("ReportCrime with invalid ID should return false")
	}
}

func TestCrimeEvidenceSystem_AddWitness(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 10.0, 20.0, 2)

	// Add witness
	result := ces.AddWitness(crimeID, 200)
	if !result {
		t.Error("AddWitness returned false")
	}

	record := ces.GetCrimeRecord(crimeID)
	if len(record.WitnessIDs) != 1 {
		t.Errorf("WitnessIDs has %d items, want 1", len(record.WitnessIDs))
	}
	if record.WitnessIDs[0] != 200 {
		t.Errorf("WitnessID = %d, want 200", record.WitnessIDs[0])
	}

	// Adding same witness again should fail
	result = ces.AddWitness(crimeID, 200)
	if result {
		t.Error("Adding duplicate witness should return false")
	}

	// Witness testimony evidence should be created
	evidence := ces.GetEvidenceForCrime(crimeID)
	if len(evidence) != 1 {
		t.Errorf("Expected 1 evidence (witness testimony), got %d", len(evidence))
	}
}

func TestCrimeEvidenceSystem_ProcessEvidence(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 0, 0, 2)
	evidence := ces.CreateEvidence(crimeID, EvidenceTypeFingerprint, 100, 0, 0, 0.8)

	if evidence.IsProcessed {
		t.Error("Evidence should not be processed initially")
	}

	result := ces.ProcessEvidence(evidence.ID)
	if !result {
		t.Error("ProcessEvidence returned false")
	}
	if !evidence.IsProcessed {
		t.Error("Evidence should be processed after ProcessEvidence")
	}

	// Invalid evidence
	result = ces.ProcessEvidence("invalid")
	if result {
		t.Error("ProcessEvidence with invalid ID should return false")
	}
}

func TestCrimeEvidenceSystem_ProcessEvidence_Destroyed(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 0, 0, 2)
	evidence := ces.CreateEvidence(crimeID, EvidenceTypeFingerprint, 100, 0, 0, 0.8)

	// Destroy evidence
	evidence.Quality = 0

	result := ces.ProcessEvidence(evidence.ID)
	if result {
		t.Error("ProcessEvidence on destroyed evidence should return false")
	}
}

func TestCrimeEvidenceSystem_TamperWithEvidence(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 0, 0, 2)
	evidence := ces.CreateEvidence(crimeID, EvidenceTypeFingerprint, 100, 0, 0, 1.0)

	result := ces.TamperWithEvidence(evidence.ID)
	if !result {
		t.Error("TamperWithEvidence returned false")
	}
	if !evidence.IsTampered {
		t.Error("Evidence should be marked tampered")
	}
	if evidence.Quality != 0.3 {
		t.Errorf("Quality after tampering = %v, want 0.3", evidence.Quality)
	}
}

func TestCrimeEvidenceSystem_TamperWithEvidence_Processed(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 0, 0, 2)
	evidence := ces.CreateEvidence(crimeID, EvidenceTypeFingerprint, 100, 0, 0, 1.0)
	evidence.IsProcessed = true

	result := ces.TamperWithEvidence(evidence.ID)
	if result {
		t.Error("TamperWithEvidence on processed evidence should return false")
	}
}

func TestCrimeEvidenceSystem_DestroyEvidence(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 0, 0, 2)
	evidence := ces.CreateEvidence(crimeID, EvidenceTypeFingerprint, 100, 0, 0, 1.0)

	result := ces.DestroyEvidence(evidence.ID)
	if !result {
		t.Error("DestroyEvidence returned false")
	}
	if evidence.Quality != 0 {
		t.Errorf("Quality after destroy = %v, want 0", evidence.Quality)
	}
}

func TestCrimeEvidenceSystem_CalculateConvictionScore(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 0, 0, 2)

	// No evidence = 0 score
	score := ces.CalculateConvictionScore(crimeID)
	if score != 0 {
		t.Errorf("Score with no evidence = %v, want 0", score)
	}

	// Add processed evidence
	evidence := ces.CreateEvidence(crimeID, EvidenceTypeBlood, 100, 0, 0, 1.0)
	evidence.IsProcessed = true

	score = ces.CalculateConvictionScore(crimeID)
	if score != 0.85 { // Blood = 0.85 base strength
		t.Errorf("Score with blood evidence = %v, want 0.85", score)
	}

	// Add more evidence to exceed 1.0 (should cap)
	evidence2 := ces.CreateEvidence(crimeID, EvidenceTypeStolenGoods, 100, 0, 0, 1.0)
	evidence2.IsProcessed = true

	score = ces.CalculateConvictionScore(crimeID)
	if score != 1.0 {
		t.Errorf("Score should cap at 1.0, got %v", score)
	}
}

func TestCrimeEvidenceSystem_InvestigateCrimeScene(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("murder", 100, 200, 0, 0, 5)

	count := ces.InvestigateCrimeScene(crimeID)
	if count == 0 {
		t.Error("InvestigateCrimeScene should generate some evidence")
	}

	record := ces.GetCrimeRecord(crimeID)
	if !record.IsInvestigated {
		t.Error("Crime should be marked as investigated")
	}

	// Second investigation should return 0
	count2 := ces.InvestigateCrimeScene(crimeID)
	if count2 != 0 {
		t.Error("Second investigation should return 0")
	}
}

func TestCrimeEvidenceSystem_Update_EvidenceDecay(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ces.EvidenceDecayRate = 0.1 // 10% per second

	crimeID := ces.RecordCrime("theft", 100, 0, 0, 0, 2)
	evidence := ces.CreateEvidence(crimeID, EvidenceTypeFootprint, 100, 0, 0, 1.0)
	evidence.CanDecay = true

	// Simulate 5 seconds
	w := ecs.NewWorld()
	ces.Update(w, 5.0)

	if evidence.Quality != 0.5 {
		t.Errorf("Quality after 5s decay = %v, want 0.5", evidence.Quality)
	}

	// Processed evidence doesn't decay
	evidence2 := ces.CreateEvidence(crimeID, EvidenceTypeFootprint, 100, 0, 0, 1.0)
	evidence2.CanDecay = true
	evidence2.IsProcessed = true

	ces.Update(w, 5.0)

	if evidence2.Quality != 1.0 {
		t.Errorf("Processed evidence quality = %v, want 1.0", evidence2.Quality)
	}
}

func TestCrimeEvidenceSystem_Update_Conviction(t *testing.T) {
	w := ecs.NewWorld()
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	// Create criminal entity
	criminal := w.CreateEntity()
	w.AddComponent(criminal, &components.Crime{WantedLevel: 0})

	// Record and report crime
	crimeID := ces.RecordCrime("theft", uint64(criminal), 0, 0, 0, 3)
	ces.ReportCrime(crimeID)

	// Add strong evidence
	evidence := ces.CreateEvidence(crimeID, EvidenceTypeStolenGoods, uint64(criminal), 0, 0, 1.0)
	evidence.IsProcessed = true

	// Update should trigger conviction
	ces.Update(w, 1.0)

	record := ces.GetCrimeRecord(crimeID)
	if !record.IsConvicted {
		t.Error("Criminal should be convicted with strong evidence")
	}

	crime, _ := w.GetComponent(criminal, "Crime")
	if crime.(*components.Crime).WantedLevel < 3 {
		t.Error("Criminal wanted level should increase on conviction")
	}
}

func TestCrimeEvidenceSystem_GetCrimesForEntity(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	// Record multiple crimes by same criminal
	crime1 := ces.RecordCrime("theft", 100, 0, 0, 0, 2)
	crime2 := ces.RecordCrime("assault", 100, 200, 0, 0, 3)
	ces.RecordCrime("murder", 999, 0, 0, 0, 5) // Different criminal

	crimes := ces.GetCrimesForEntity(100)
	if len(crimes) != 2 {
		t.Errorf("GetCrimesForEntity returned %d crimes, want 2", len(crimes))
	}

	found1, found2 := false, false
	for _, id := range crimes {
		if id == crime1 {
			found1 = true
		}
		if id == crime2 {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Error("Expected both crimes to be found")
	}
}

func TestCrimeEvidenceSystem_GetUnsolvedCrimes(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crime1 := ces.RecordCrime("theft", 100, 0, 0, 0, 2)
	crime2 := ces.RecordCrime("assault", 200, 0, 0, 0, 3)
	ces.RecordCrime("murder", 300, 0, 0, 0, 5) // Not reported

	// Report crimes
	ces.ReportCrime(crime1)
	ces.ReportCrime(crime2)

	// Solve one crime
	record1 := ces.GetCrimeRecord(crime1)
	record1.IsSolved = true

	unsolved := ces.GetUnsolvedCrimes()
	if len(unsolved) != 1 {
		t.Errorf("GetUnsolvedCrimes returned %d crimes, want 1", len(unsolved))
	}
	if unsolved[0].ID != crime2 {
		t.Error("Expected crime2 to be in unsolved list")
	}
}

func TestCrimeEvidenceSystem_ClearRecord(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	crimeID := ces.RecordCrime("theft", 100, 0, 0, 0, 2)
	ces.ReportCrime(crimeID)

	result := ces.ClearRecord(crimeID)
	if !result {
		t.Error("ClearRecord returned false")
	}

	record := ces.GetCrimeRecord(crimeID)
	if !record.IsSolved {
		t.Error("Crime should be marked solved after clearing")
	}
	if record.ConvictionScore != 0 {
		t.Error("ConvictionScore should be 0 after clearing")
	}

	// Crime should be removed from entity tracking
	crimes := ces.GetCrimesForEntity(100)
	for _, id := range crimes {
		if id == crimeID {
			t.Error("Crime should be removed from entity tracking")
		}
	}

	// Invalid crime
	result = ces.ClearRecord("invalid")
	if result {
		t.Error("ClearRecord with invalid ID should return false")
	}
}

func TestCrimeEvidenceSystem_GetCrimeTypeDescription(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	tests := []struct {
		crimeType string
		wantLen   int // Just check non-empty
	}{
		{"murder", 10},
		{"assault", 10},
		{"theft", 10},
		{"robbery", 10},
		{"burglary", 10},
		{"fraud", 10},
		{"forgery", 10},
		{"trespass", 10},
		{"vandalism", 10},
		{"smuggling", 10},
		{"unknown", 10},
	}

	for _, tt := range tests {
		desc := ces.GetCrimeTypeDescription(tt.crimeType)
		if len(desc) < tt.wantLen {
			t.Errorf("GetCrimeTypeDescription(%q) too short: %q", tt.crimeType, desc)
		}
	}
}

func TestCrimeEvidenceSystem_GetCrimeSeverity(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	tests := []struct {
		crimeType    string
		wantSeverity int
	}{
		{"murder", 5},
		{"assault", 3},
		{"robbery", 4},
		{"theft", 2},
		{"burglary", 3},
		{"fraud", 2},
		{"forgery", 2},
		{"trespass", 1},
		{"vandalism", 1},
		{"smuggling", 3},
		{"unknown", 2},
	}

	for _, tt := range tests {
		severity := ces.GetCrimeSeverity(tt.crimeType)
		if severity != tt.wantSeverity {
			t.Errorf("GetCrimeSeverity(%q) = %d, want %d", tt.crimeType, severity, tt.wantSeverity)
		}
	}
}

func TestCrimeEvidenceSystem_GenreDescriptions(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			ces := NewCrimeEvidenceSystem(cs, genre, 12345)
			crimeID := ces.RecordCrime("murder", 100, 200, 0, 0, 5)

			// Create evidence of different types
			evidence := ces.CreateEvidence(crimeID, EvidenceTypeFingerprint, 100, 0, 0, 1.0)
			if evidence.Description == "" {
				t.Error("Evidence description should not be empty")
			}

			// Genre-specific evidence types
			if genre == "fantasy" {
				evidence = ces.CreateEvidence(crimeID, EvidenceTypeMagical, 100, 0, 0, 1.0)
				if evidence.Description == "" {
					t.Error("Magical evidence description should not be empty")
				}
			}
			if genre == "sci-fi" || genre == "cyberpunk" {
				evidence = ces.CreateEvidence(crimeID, EvidenceTypeDigital, 100, 0, 0, 1.0)
				if evidence.Description == "" {
					t.Error("Digital evidence description should not be empty")
				}
			}
		})
	}
}

func TestCrimeEvidenceSystem_pseudoRandom(t *testing.T) {
	cs := NewCrimeSystem(60.0, 100.0)
	ces1 := NewCrimeEvidenceSystem(cs, "fantasy", 12345)
	ces2 := NewCrimeEvidenceSystem(cs, "fantasy", 12345)

	// Test determinism
	for i := 0; i < 10; i++ {
		v1 := ces1.pseudoRandom()
		v2 := ces2.pseudoRandom()
		if v1 != v2 {
			t.Errorf("pseudoRandom not deterministic at %d: %v vs %v", i, v1, v2)
		}
		if v1 < 0 || v1 > 1 {
			t.Errorf("pseudoRandom out of range: %v", v1)
		}
	}
}

func TestFormatCrimeID(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "CR-0"},
		{1, "CR-1"},
		{123, "CR-123"},
		{99999, "CR-99999"},
	}

	for _, tt := range tests {
		got := formatCrimeID(tt.n)
		if got != tt.want {
			t.Errorf("formatCrimeID(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatEvidenceID(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "EV-0"},
		{1, "EV-1"},
		{123, "EV-123"},
		{99999, "EV-99999"},
	}

	for _, tt := range tests {
		got := formatEvidenceID(tt.n)
		if got != tt.want {
			t.Errorf("formatEvidenceID(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
