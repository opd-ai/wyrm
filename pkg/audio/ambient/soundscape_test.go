package ambient

import (
	"testing"
)

func TestNewSoundscape(t *testing.T) {
	s := NewSoundscape(RegionCave, "fantasy", 42)
	if s == nil {
		t.Fatal("NewSoundscape returned nil")
	}
	if s.regionType != RegionCave {
		t.Errorf("expected RegionCave, got %v", s.regionType)
	}
	if s.genre != "fantasy" {
		t.Errorf("expected 'fantasy', got %q", s.genre)
	}
}

func TestAllRegionTypes(t *testing.T) {
	regions := []RegionType{
		RegionPlains, RegionForest, RegionCave, RegionCity,
		RegionWater, RegionDesert, RegionMountain, RegionDungeon,
		RegionInterior,
	}
	
	for _, region := range regions {
		s := NewSoundscape(region, "fantasy", 42)
		samples := s.GenerateSamples(0.1)
		
		if len(samples) == 0 {
			t.Errorf("region %v: no samples generated", region)
			continue
		}
		
		// Check samples are in valid range
		for i, sample := range samples {
			if sample < -1.0 || sample > 1.0 {
				t.Errorf("region %v: sample %d out of range: %f", region, i, sample)
				break
			}
		}
	}
}

func TestAllGenreModifications(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	
	for _, genre := range genres {
		s := NewSoundscape(RegionCave, genre, 42)
		samples := s.GenerateSamples(0.1)
		
		if len(samples) == 0 {
			t.Errorf("genre %q: no samples generated", genre)
		}
	}
}

func TestRegionSamplesAreDifferent(t *testing.T) {
	plains := NewSoundscape(RegionPlains, "fantasy", 42)
	cave := NewSoundscape(RegionCave, "fantasy", 42)
	city := NewSoundscape(RegionCity, "fantasy", 42)
	
	plainsSamples := plains.GenerateSamples(0.5)
	caveSamples := cave.GenerateSamples(0.5)
	citySamples := city.GenerateSamples(0.5)
	
	// Compare RMS amplitude (different regions have different characteristics)
	plainsRMS := calculateRMS(plainsSamples)
	caveRMS := calculateRMS(caveSamples)
	cityRMS := calculateRMS(citySamples)
	
	// At least one should differ significantly
	if almostEqual(plainsRMS, caveRMS) && almostEqual(caveRMS, cityRMS) {
		t.Error("different regions should produce different sound characteristics")
	}
}

func calculateRMS(samples []float64) float64 {
	sum := 0.0
	for _, s := range samples {
		sum += s * s
	}
	return sum / float64(len(samples))
}

func almostEqual(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.0001
}

func TestNewAmbientManager(t *testing.T) {
	am := NewAmbientManager("fantasy", 42)
	if am == nil {
		t.Fatal("NewAmbientManager returned nil")
	}
	if am.GetCurrentRegion() != RegionPlains {
		t.Error("should start in plains region")
	}
}

func TestSetRegion(t *testing.T) {
	am := NewAmbientManager("fantasy", 42)
	
	am.SetRegion(RegionCave)
	
	if am.GetCurrentRegion() != RegionCave {
		t.Error("region should change to cave")
	}
	
	if !am.IsTransitioning() {
		t.Error("should be transitioning after region change")
	}
}

func TestTransitionProgress(t *testing.T) {
	am := NewAmbientManager("fantasy", 42)
	am.transitionTime = 1.0 // 1 second
	
	am.SetRegion(RegionForest)
	
	// Should be transitioning
	if !am.IsTransitioning() {
		t.Error("should be transitioning")
	}
	
	// Update for half the transition time
	am.Update(0.5)
	
	// Still transitioning
	if !am.IsTransitioning() {
		t.Error("should still be transitioning at 50%")
	}
	
	// Update past transition time
	am.Update(0.6)
	
	// No longer transitioning
	if am.IsTransitioning() {
		t.Error("should not be transitioning after full duration")
	}
}

func TestTransitionTimingAC(t *testing.T) {
	// ROADMAP AC: Ambient sound type changes within 1s of entering new region type
	am := NewAmbientManager("fantasy", 42)
	
	// Default transition time should be 1 second
	if am.transitionTime != 1.0 {
		t.Errorf("transition time should be 1.0s, got %f", am.transitionTime)
	}
	
	am.SetRegion(RegionCave)
	
	// After 1 second, transition should be complete
	am.Update(1.0)
	
	if am.IsTransitioning() {
		t.Error("transition should complete within 1s per AC")
	}
}

func TestGenerateSamples(t *testing.T) {
	am := NewAmbientManager("fantasy", 42)
	
	samples := am.GenerateSamples(0.5)
	
	expectedLen := int(0.5 * 44100)
	if len(samples) != expectedLen {
		t.Errorf("expected %d samples, got %d", expectedLen, len(samples))
	}
}

func TestSameRegionNoTransition(t *testing.T) {
	am := NewAmbientManager("fantasy", 42)
	
	// Complete initial state
	am.transitionProgress = 1.0
	
	// Set to same region
	am.SetRegion(RegionPlains)
	
	// Should not trigger transition
	if am.IsTransitioning() {
		t.Error("setting same region should not trigger transition")
	}
}

func BenchmarkGenerateCaveSamples(b *testing.B) {
	s := NewSoundscape(RegionCave, "fantasy", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.GenerateSamples(1.0)
	}
}

func BenchmarkGenerateCitySamples(b *testing.B) {
	s := NewSoundscape(RegionCity, "cyberpunk", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.GenerateSamples(1.0)
	}
}
