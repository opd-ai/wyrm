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

// ============================================================================
// AmbientMixer Tests
// ============================================================================

func TestNewAmbientMixer(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	if mixer == nil {
		t.Fatal("NewAmbientMixer returned nil")
	}
	if mixer.masterVolume != 1.0 {
		t.Errorf("expected masterVolume 1.0, got %f", mixer.masterVolume)
	}
	if mixer.GetLayerCount() != 0 {
		t.Error("new mixer should have no layers")
	}
}

func TestAmbientMixer_AddLayer(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)

	mixer.AddLayer("background", RegionPlains, 0.5, 0)

	if mixer.GetLayerCount() != 1 {
		t.Errorf("expected 1 layer, got %d", mixer.GetLayerCount())
	}
	if !mixer.HasLayer("background") {
		t.Error("layer 'background' should exist")
	}
}

func TestAmbientMixer_AddMultipleLayers(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)

	mixer.AddLayer("background", RegionPlains, 0.5, 0)
	mixer.AddLayer("weather", RegionForest, 0.3, 1)
	mixer.AddLayer("events", RegionCave, 0.2, 2)

	if mixer.GetLayerCount() != 3 {
		t.Errorf("expected 3 layers, got %d", mixer.GetLayerCount())
	}
}

func TestAmbientMixer_MaxLayers(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)

	// Add max layers
	for i := 0; i < mixer.maxLayers; i++ {
		mixer.AddLayer(string(rune('a'+i)), RegionPlains, 0.1, i)
	}

	// Try to add one more
	mixer.AddLayer("overflow", RegionCave, 0.5, 99)

	if mixer.GetLayerCount() != mixer.maxLayers {
		t.Errorf("should not exceed maxLayers (%d), got %d", mixer.maxLayers, mixer.GetLayerCount())
	}
}

func TestAmbientMixer_RemoveLayer(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("test", RegionPlains, 0.5, 0)

	// Force the volume to non-zero so it fades
	mixer.mu.Lock()
	mixer.layers["test"].volume = 0.5
	mixer.mu.Unlock()

	mixer.RemoveLayer("test")

	// Layer should still exist but have targetVol 0
	mixer.mu.Lock()
	layer := mixer.layers["test"]
	mixer.mu.Unlock()

	if layer.targetVol != 0 {
		t.Errorf("removed layer should have targetVol 0, got %f", layer.targetVol)
	}

	// After enough updates, layer should be removed
	for i := 0; i < 100; i++ {
		mixer.Update(0.05)
	}

	if mixer.HasLayer("test") {
		t.Error("layer should be removed after fadeout")
	}
}

func TestAmbientMixer_SetLayerVolume(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("test", RegionPlains, 0.5, 0)

	mixer.SetLayerVolume("test", 0.8)

	mixer.mu.Lock()
	targetVol := mixer.layers["test"].targetVol
	mixer.mu.Unlock()

	if targetVol != 0.8 {
		t.Errorf("expected targetVol 0.8, got %f", targetVol)
	}
}

func TestAmbientMixer_SetLayerVolume_Clamping(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("test", RegionPlains, 0.5, 0)

	// Test clamping above 1
	mixer.SetLayerVolume("test", 1.5)
	mixer.mu.Lock()
	targetVol := mixer.layers["test"].targetVol
	mixer.mu.Unlock()
	if targetVol != 1.0 {
		t.Errorf("volume should be clamped to 1.0, got %f", targetVol)
	}

	// Test clamping below 0
	mixer.SetLayerVolume("test", -0.5)
	mixer.mu.Lock()
	targetVol = mixer.layers["test"].targetVol
	mixer.mu.Unlock()
	if targetVol != 0.0 {
		t.Errorf("volume should be clamped to 0.0, got %f", targetVol)
	}
}

func TestAmbientMixer_SetLayerPan(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("test", RegionPlains, 0.5, 0)

	mixer.SetLayerPan("test", -0.5)

	mixer.mu.Lock()
	pan := mixer.layers["test"].pan
	mixer.mu.Unlock()

	if pan != -0.5 {
		t.Errorf("expected pan -0.5, got %f", pan)
	}
}

func TestAmbientMixer_SetLayerPan_Clamping(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("test", RegionPlains, 0.5, 0)

	mixer.SetLayerPan("test", 2.0)
	mixer.mu.Lock()
	pan := mixer.layers["test"].pan
	mixer.mu.Unlock()
	if pan != 1.0 {
		t.Errorf("pan should be clamped to 1.0, got %f", pan)
	}

	mixer.SetLayerPan("test", -2.0)
	mixer.mu.Lock()
	pan = mixer.layers["test"].pan
	mixer.mu.Unlock()
	if pan != -1.0 {
		t.Errorf("pan should be clamped to -1.0, got %f", pan)
	}
}

func TestAmbientMixer_MasterVolume(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)

	mixer.SetMasterVolume(0.7)
	if mixer.GetMasterVolume() != 0.7 {
		t.Errorf("expected 0.7, got %f", mixer.GetMasterVolume())
	}

	// Test clamping
	mixer.SetMasterVolume(1.5)
	if mixer.GetMasterVolume() != 1.0 {
		t.Error("master volume should be clamped to 1.0")
	}

	mixer.SetMasterVolume(-0.5)
	if mixer.GetMasterVolume() != 0.0 {
		t.Error("master volume should be clamped to 0.0")
	}
}

func TestAmbientMixer_CrossfadeTime(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)

	mixer.SetCrossfadeTime(0.25)
	if mixer.GetCrossfadeTime() != 0.25 {
		t.Errorf("expected 0.25, got %f", mixer.GetCrossfadeTime())
	}

	// Test minimum clamping
	mixer.SetCrossfadeTime(0.001)
	if mixer.GetCrossfadeTime() < 0.01 {
		t.Error("crossfade time should have minimum 0.01s")
	}
}

func TestAmbientMixer_Update_FadeIn(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.SetCrossfadeTime(0.5)

	mixer.AddLayer("test", RegionPlains, 0.8, 0)

	// Initial volume should be 0
	if mixer.GetLayerVolume("test") != 0 {
		t.Error("initial volume should be 0")
	}

	// After some updates, volume should increase toward target
	mixer.Update(0.25)
	vol := mixer.GetLayerVolume("test")
	if vol <= 0 || vol >= 0.8 {
		t.Errorf("volume should be between 0 and 0.8 during fade, got %f", vol)
	}

	// After full crossfade time, should be at target
	mixer.Update(0.5)
	vol = mixer.GetLayerVolume("test")
	if vol != 0.8 {
		t.Errorf("volume should reach target 0.8, got %f", vol)
	}
}

func TestAmbientMixer_GenerateMixedSamples(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("background", RegionPlains, 0.5, 0)

	// Set volume directly for testing
	mixer.mu.Lock()
	mixer.layers["background"].volume = 0.5
	mixer.mu.Unlock()

	samples := mixer.GenerateMixedSamples(0.1)
	expectedLen := int(0.1 * 44100)

	if len(samples) != expectedLen {
		t.Errorf("expected %d samples, got %d", expectedLen, len(samples))
	}

	// Samples should have some signal
	hasNonZero := false
	for _, s := range samples {
		if s != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Error("mixed samples should contain non-zero values")
	}
}

func TestAmbientMixer_GenerateStereoSamples(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("left", RegionPlains, 0.5, 0)
	mixer.AddLayer("right", RegionForest, 0.5, 1)

	mixer.mu.Lock()
	mixer.layers["left"].volume = 0.5
	mixer.layers["left"].pan = -1.0
	mixer.layers["right"].volume = 0.5
	mixer.layers["right"].pan = 1.0
	mixer.mu.Unlock()

	left, right := mixer.GenerateStereoSamples(0.1)

	if len(left) != len(right) {
		t.Error("left and right channels should have same length")
	}

	expectedLen := int(0.1 * 44100)
	if len(left) != expectedLen {
		t.Errorf("expected %d samples, got %d", expectedLen, len(left))
	}
}

func TestAmbientMixer_SoftClipping(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)

	// Create loud samples that would clip
	samples := []float64{2.0, -2.0, 1.5, -1.5}
	mixer.applySoftClipping(samples)

	for i, s := range samples {
		if s > 1.0 || s < -1.0 {
			t.Errorf("sample %d should be clipped, got %f", i, s)
		}
	}
}

func TestAmbientMixer_GetLayerNames(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("alpha", RegionPlains, 0.5, 0)
	mixer.AddLayer("beta", RegionForest, 0.5, 1)

	names := mixer.GetLayerNames()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}

	hasAlpha, hasBeta := false, false
	for _, name := range names {
		if name == "alpha" {
			hasAlpha = true
		}
		if name == "beta" {
			hasBeta = true
		}
	}
	if !hasAlpha || !hasBeta {
		t.Error("should return both layer names")
	}
}

func TestAmbientMixer_CrossfadeTo(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("old", RegionPlains, 0.5, 0)

	// Force volume
	mixer.mu.Lock()
	mixer.layers["old"].volume = 0.5
	mixer.mu.Unlock()

	mixer.CrossfadeTo("old", "new", RegionCave, 0.6, 0)

	// Should have both layers during crossfade
	if !mixer.HasLayer("new") {
		t.Error("new layer should exist")
	}

	// Old layer should be fading out
	mixer.mu.Lock()
	oldTarget := mixer.layers["old"].targetVol
	mixer.mu.Unlock()
	if oldTarget != 0 {
		t.Error("old layer should be fading out")
	}
}

func TestAmbientMixer_LayerPriority(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)

	// Add layers in non-priority order
	mixer.AddLayer("high", RegionCave, 0.5, 10)
	mixer.AddLayer("low", RegionPlains, 0.5, 1)
	mixer.AddLayer("med", RegionForest, 0.5, 5)

	// Verify order is sorted by priority
	mixer.mu.Lock()
	order := mixer.layerOrder
	mixer.mu.Unlock()

	if len(order) != 3 {
		t.Fatalf("expected 3 layers in order, got %d", len(order))
	}

	// Should be: low (1), med (5), high (10)
	if order[0] != "low" || order[1] != "med" || order[2] != "high" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestAmbientMixer_ConcurrentAccess(t *testing.T) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("test", RegionPlains, 0.5, 0)

	// Force volume
	mixer.mu.Lock()
	mixer.layers["test"].volume = 0.5
	mixer.mu.Unlock()

	done := make(chan bool, 3)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = mixer.GetLayerCount()
			_ = mixer.GetLayerVolume("test")
		}
		done <- true
	}()

	// Concurrent updates
	go func() {
		for i := 0; i < 100; i++ {
			mixer.Update(0.01)
		}
		done <- true
	}()

	// Concurrent sample generation
	go func() {
		for i := 0; i < 10; i++ {
			_ = mixer.GenerateMixedSamples(0.01)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

func BenchmarkAmbientMixer_GenerateMixedSamples(b *testing.B) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("bg", RegionPlains, 0.5, 0)
	mixer.AddLayer("weather", RegionForest, 0.3, 1)

	// Set volumes
	mixer.mu.Lock()
	mixer.layers["bg"].volume = 0.5
	mixer.layers["weather"].volume = 0.3
	mixer.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mixer.GenerateMixedSamples(0.1)
	}
}

func BenchmarkAmbientMixer_GenerateStereoSamples(b *testing.B) {
	mixer := NewAmbientMixer("fantasy", 42)
	mixer.AddLayer("left", RegionPlains, 0.5, 0)
	mixer.AddLayer("right", RegionForest, 0.5, 1)

	mixer.mu.Lock()
	mixer.layers["left"].volume = 0.5
	mixer.layers["left"].pan = -0.5
	mixer.layers["right"].volume = 0.5
	mixer.layers["right"].pan = 0.5
	mixer.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mixer.GenerateStereoSamples(0.1)
	}
}
