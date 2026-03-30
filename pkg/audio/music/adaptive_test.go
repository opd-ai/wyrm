package music

import (
	"testing"
	"time"
)

func TestNewAdaptiveMusic(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	if am == nil {
		t.Fatal("NewAdaptiveMusic returned nil")
	}
	if am.genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got %q", am.genre)
	}
	if am.GetCurrentState() != StateExploration {
		t.Error("should start in exploration state")
	}
}

func TestAdaptiveMusicAllGenres(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		am := NewAdaptiveMusic(genre, 42)
		if am == nil {
			t.Errorf("genre %q: NewAdaptiveMusic returned nil", genre)
			continue
		}

		// Verify motifs exist
		if _, ok := am.motifs["exploration"]; !ok {
			t.Errorf("genre %q: missing exploration motif", genre)
		}
		if _, ok := am.motifs["combat"]; !ok {
			t.Errorf("genre %q: missing combat motif", genre)
		}

		// Verify layers exist
		if _, ok := am.layers["exploration"]; !ok {
			t.Errorf("genre %q: missing exploration layer", genre)
		}
		if _, ok := am.layers["combat"]; !ok {
			t.Errorf("genre %q: missing combat layer", genre)
		}
	}
}

func TestEnterCombat(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	// Initially in exploration
	if am.GetCurrentState() != StateExploration {
		t.Error("should start in exploration")
	}

	// Enter combat
	am.EnterCombat()

	if am.GetCurrentState() != StateCombat {
		t.Error("should be in combat state after EnterCombat")
	}

	// Combat layer should have target 1.0
	if am.layers["combat"].Target != 1.0 {
		t.Errorf("combat layer target should be 1.0, got %f", am.layers["combat"].Target)
	}

	// Exploration layer should reduce
	if am.layers["exploration"].Target >= 1.0 {
		t.Error("exploration layer should reduce during combat")
	}
}

func TestExitCombat(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	am.EnterCombat()
	am.ExitCombat()

	if am.GetCurrentState() != StateExploration {
		t.Error("should return to exploration after ExitCombat")
	}

	// Layers should target exploration state
	if am.layers["exploration"].Target != 1.0 {
		t.Errorf("exploration layer target should be 1.0, got %f", am.layers["exploration"].Target)
	}
	if am.layers["combat"].Target != 0.0 {
		t.Errorf("combat layer target should be 0.0, got %f", am.layers["combat"].Target)
	}
}

func TestCrossfade(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	am.crossfadeDuration = 1.0 // 1 second for easier testing

	am.EnterCombat()

	// Initial exploration volume is 1.0, target is reduced
	initialExplorationVol := am.GetLayerVolume("exploration")

	// Simulate time passing
	for i := 0; i < 50; i++ {
		am.Update(0.02) // 20ms per update
	}

	// Volumes should have moved toward targets
	explorationVol := am.GetLayerVolume("exploration")
	combatVol := am.GetLayerVolume("combat")

	if explorationVol >= initialExplorationVol {
		t.Error("exploration volume should decrease during combat")
	}
	if combatVol <= 0 {
		t.Error("combat volume should increase after entering combat")
	}
}

func TestCombatTransitionTiming(t *testing.T) {
	// ROADMAP AC: Music transitions within 2s of entering combat
	am := NewAdaptiveMusic("fantasy", 42)
	am.crossfadeDuration = 2.0 // 2 second transition

	am.EnterCombat()

	// Simulate 2 seconds of updates
	for i := 0; i < 100; i++ {
		am.Update(0.02) // 20ms = 2 seconds total
	}

	// Combat layer should be at or very close to target
	combatVol := am.GetLayerVolume("combat")
	if combatVol < 0.9 {
		t.Errorf("combat layer should be near 1.0 after 2s, got %f", combatVol)
	}
}

func TestAutoCombatExitTiming(t *testing.T) {
	// ROADMAP AC: Music reverts within 5s of last enemy death
	// This test verifies the logic, not real-time waiting
	am := NewAdaptiveMusic("fantasy", 42)
	am.crossfadeDuration = 1.0

	am.EnterCombat()

	// Record enemy death
	am.EnemyDied()

	// Manually set the lastEnemyDeath to 5+ seconds ago
	am.mu.Lock()
	am.lastEnemyDeath = time.Now().Add(-6 * time.Second)
	am.mu.Unlock()

	// Call update which should trigger the exit
	am.Update(0.02)

	// Should have automatically exited combat
	if am.GetCurrentState() != StateExploration {
		t.Error("should auto-exit combat 5s after last enemy death")
	}
}

func TestGenerateSamples(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	samples := am.GenerateSamples(0.5) // 0.5 seconds

	expectedLen := int(0.5 * float64(am.sampleRate))
	if len(samples) != expectedLen {
		t.Errorf("expected %d samples, got %d", expectedLen, len(samples))
	}

	// Check samples are in valid range
	for i, s := range samples {
		if s < -1.0 || s > 1.0 {
			t.Errorf("sample %d out of range: %f", i, s)
			break
		}
	}
}

func TestGenerateSamplesInCombat(t *testing.T) {
	am := NewAdaptiveMusic("cyberpunk", 42)

	// Generate exploration samples
	am.layers["exploration"].Volume = 1.0
	am.layers["combat"].Volume = 0.0
	explorationSamples := am.GenerateSamples(0.2)

	// Switch to combat
	am.EnterCombat()
	am.layers["exploration"].Volume = 0.3
	am.layers["combat"].Volume = 1.0
	combatSamples := am.GenerateSamples(0.2)

	// Samples should be different
	differences := 0
	for i := range explorationSamples {
		if explorationSamples[i] != combatSamples[i] {
			differences++
		}
	}

	if differences == 0 {
		t.Error("combat and exploration samples should differ")
	}
}

func TestTimeSinceCombatEntry(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	if am.TimeSinceCombatEntry() != 0 {
		t.Error("should be 0 before entering combat")
	}

	am.EnterCombat()
	time.Sleep(10 * time.Millisecond)

	elapsed := am.TimeSinceCombatEntry()
	if elapsed < 10*time.Millisecond {
		t.Error("should track time since combat entry")
	}
}

func TestTimeSinceLastEnemyDeath(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	if am.TimeSinceLastEnemyDeath() != 0 {
		t.Error("should be 0 before any enemy death")
	}

	am.EnemyDied()
	time.Sleep(10 * time.Millisecond)

	elapsed := am.TimeSinceLastEnemyDeath()
	if elapsed < 10*time.Millisecond {
		t.Error("should track time since last enemy death")
	}
}

func BenchmarkGenerateSamples(b *testing.B) {
	am := NewAdaptiveMusic("fantasy", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = am.GenerateSamples(1.0)
	}
}

func BenchmarkUpdate(b *testing.B) {
	am := NewAdaptiveMusic("fantasy", 42)
	am.EnterCombat()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am.Update(0.016) // 60 FPS
	}
}

// ============================================================================
// Genre Style Tests
// ============================================================================

func TestGetGenreStyle(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		style := GetGenreStyle(genre)
		if style == nil {
			t.Errorf("genre %s: GetGenreStyle returned nil", genre)
			continue
		}

		if style.Genre != genre {
			t.Errorf("genre %s: style.Genre = %s", genre, style.Genre)
		}

		// Verify scale has notes
		if len(style.BaseScale) == 0 {
			t.Errorf("genre %s: empty BaseScale", genre)
		}

		// Verify tempo is reasonable
		if style.Tempo < 40 || style.Tempo > 200 {
			t.Errorf("genre %s: tempo %f out of range", genre, style.Tempo)
		}

		// Verify instruments are defined
		if len(style.InstrumentMix) == 0 {
			t.Errorf("genre %s: empty InstrumentMix", genre)
		}
	}
}

func TestGetGenreStyleDefaults(t *testing.T) {
	// Unknown genre should default to fantasy
	style := GetGenreStyle("unknown")
	if style == nil {
		t.Fatal("GetGenreStyle returned nil for unknown genre")
	}
	if style.Genre != "fantasy" {
		t.Errorf("unknown genre should default to fantasy, got %s", style.Genre)
	}
}

func TestGenreStyleUniqueCharacteristics(t *testing.T) {
	fantasy := GetGenreStyle("fantasy")
	horror := GetGenreStyle("horror")
	cyberpunk := GetGenreStyle("cyberpunk")

	// Horror should be slower
	if horror.Tempo >= fantasy.Tempo {
		t.Error("horror should have slower tempo than fantasy")
	}

	// Cyberpunk should be faster
	if cyberpunk.Tempo <= fantasy.Tempo {
		t.Error("cyberpunk should have faster tempo than fantasy")
	}

	// Horror should have more reverb
	if horror.ReverbAmount <= fantasy.ReverbAmount {
		t.Error("horror should have more reverb than fantasy")
	}

	// Cyberpunk should have more distortion
	if cyberpunk.DistortionMix <= fantasy.DistortionMix {
		t.Error("cyberpunk should have more distortion than fantasy")
	}
}

func TestApplyGenreStyle(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	sciFiStyle := GetGenreStyle("sci-fi")
	am.ApplyGenreStyle(sciFiStyle)

	if am.GetGenre() != "sci-fi" {
		t.Errorf("genre should be sci-fi, got %s", am.GetGenre())
	}

	// Motifs should be regenerated for sci-fi
	motif := am.motifs["exploration"]
	if motif.Genre != "sci-fi" {
		t.Error("motif should be regenerated for sci-fi")
	}
}

// ============================================================================
// Location Music Tests
// ============================================================================

func TestNewLocationMusicManager(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	lmm := NewLocationMusicManager(am)

	if lmm == nil {
		t.Fatal("NewLocationMusicManager returned nil")
	}

	if lmm.GetCurrentLocation() != LocationWilderness {
		t.Error("should start in wilderness")
	}

	if lmm.IsInTransition() {
		t.Error("should not be in transition initially")
	}
}

func TestLocationMusicSetLocation(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	lmm := NewLocationMusicManager(am)

	lmm.SetLocation(LocationTown)

	if !lmm.IsInTransition() {
		t.Error("should be in transition after SetLocation")
	}

	progress := lmm.GetTransitionProgress()
	if progress < 0 || progress > 1 {
		t.Errorf("transition progress %f out of range", progress)
	}
}

func TestLocationMusicTransitionComplete(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	lmm := NewLocationMusicManager(am)

	lmm.SetLocation(LocationDungeon)

	// Get transition time
	config := lmm.GetLocationConfig(LocationDungeon)
	transitionTime := config.TransitionTime

	// Update past transition time
	for elapsed := 0.0; elapsed < transitionTime+0.5; elapsed += 0.1 {
		lmm.Update(0.1)
	}

	if lmm.GetCurrentLocation() != LocationDungeon {
		t.Error("should have transitioned to dungeon")
	}

	if lmm.IsInTransition() {
		t.Error("transition should be complete")
	}

	if lmm.GetTransitionProgress() != 1.0 {
		t.Errorf("transition progress should be 1.0, got %f", lmm.GetTransitionProgress())
	}
}

func TestLocationMusicNoChangeOnSameLocation(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	lmm := NewLocationMusicManager(am)

	// Setting same location should not start transition
	lmm.SetLocation(LocationWilderness)

	if lmm.IsInTransition() {
		t.Error("should not transition to same location")
	}
}

func TestLocationConfigsExist(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	lmm := NewLocationMusicManager(am)

	locations := []LocationType{
		LocationWilderness,
		LocationTown,
		LocationDungeon,
		LocationTavern,
		LocationTemple,
		LocationCastle,
		LocationShop,
		LocationCombatArena,
		LocationBossRoom,
		LocationSafeZone,
	}

	for _, loc := range locations {
		config := lmm.GetLocationConfig(loc)
		if config == nil {
			t.Errorf("location %d has no config", loc)
			continue
		}

		if config.TransitionTime <= 0 {
			t.Errorf("location %d has invalid transition time", loc)
		}

		if len(config.LayerWeights) < 3 {
			t.Errorf("location %d has insufficient layer weights", loc)
		}
	}
}

func TestLocationMusicLayerIntegration(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	lmm := NewLocationMusicManager(am)

	// Go to dungeon (increases tension)
	lmm.SetLocation(LocationDungeon)

	// Complete transition
	config := lmm.GetLocationConfig(LocationDungeon)
	for elapsed := 0.0; elapsed < config.TransitionTime+0.5; elapsed += 0.1 {
		lmm.Update(0.1)
	}

	// Tension layer should have been activated (dungeon has 0.5 tension weight)
	tensionTarget := am.layers["tension"].Target
	if tensionTarget < 0.4 {
		t.Errorf("dungeon should increase tension layer, got target %f", tensionTarget)
	}
}

func BenchmarkLocationMusicTransition(b *testing.B) {
	am := NewAdaptiveMusic("fantasy", 42)
	lmm := NewLocationMusicManager(am)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lmm.SetLocation(LocationType(i % 10))
		lmm.Update(0.016)
	}
}
