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

// ============================================================================
// Boss Music Manager Tests
// ============================================================================

func TestNewBossMusicManager(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	if bmm == nil {
		t.Fatal("NewBossMusicManager returned nil")
	}

	if bmm.genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got %q", bmm.genre)
	}

	if bmm.IsActive() {
		t.Error("should not be active initially")
	}
}

func TestBossMusicManagerAllGenres(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		bmm := NewBossMusicManager(genre)
		if bmm == nil {
			t.Errorf("genre %q: NewBossMusicManager returned nil", genre)
			continue
		}

		// Verify default configs exist
		if _, ok := bmm.configs["generic"]; !ok {
			t.Errorf("genre %q: missing 'generic' boss config", genre)
		}
		if _, ok := bmm.configs["final"]; !ok {
			t.Errorf("genre %q: missing 'final' boss config", genre)
		}
	}
}

func TestStartBossFight(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")

	bmm.StartBossFight("generic")

	if !bmm.IsActive() {
		t.Error("should be active after StartBossFight")
	}

	if bmm.GetCurrentPhase() != BossPhaseIntro {
		t.Errorf("should start at BossPhaseIntro, got %v", bmm.GetCurrentPhase())
	}

	// Check initial health
	bmm.mu.RLock()
	health := bmm.bossHealth
	bmm.mu.RUnlock()
	if health != 1.0 {
		t.Errorf("boss health should be 1.0, got %f", health)
	}
}

func TestStartBossFightUnknownType(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")

	// Unknown boss type should default to "generic"
	bmm.StartBossFight("unknown_boss_type")

	if !bmm.IsActive() {
		t.Error("should be active after StartBossFight with unknown type")
	}

	bmm.mu.RLock()
	bossType := bmm.currentBoss
	bmm.mu.RUnlock()
	if bossType != "generic" {
		t.Errorf("unknown boss type should default to 'generic', got %q", bossType)
	}
}

func TestBossPhaseTransitions(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	bmm.StartBossFight("generic")

	tests := []struct {
		health        float64
		expectedPhase BossMusicPhase
	}{
		{1.0, BossPhaseIntro},
		{0.74, BossPhaseMain},
		{0.49, BossPhaseIntense},
		{0.24, BossPhaseFinal},
	}

	for _, tc := range tests {
		bmm.UpdateBossHealth(tc.health)
		phase := bmm.GetCurrentPhase()
		if phase != tc.expectedPhase {
			t.Errorf("at health %f: expected phase %v, got %v", tc.health, tc.expectedPhase, phase)
		}
	}
}

func TestBossPhaseTransitionsNoReverse(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	bmm.StartBossFight("generic")

	// Go to intense phase
	bmm.UpdateBossHealth(0.4)
	phase1 := bmm.GetCurrentPhase()

	// Health goes back up (healing)
	bmm.UpdateBossHealth(0.8)
	phase2 := bmm.GetCurrentPhase()

	// Phase should not revert (phases only go forward)
	if phase2 < phase1 {
		t.Errorf("phase should not decrease: was %v, now %v", phase1, phase2)
	}
}

func TestEndBossFightVictory(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	bmm.StartBossFight("generic")

	bmm.EndBossFight(true)

	if bmm.IsActive() {
		t.Error("should not be active after EndBossFight")
	}

	if bmm.GetCurrentPhase() != BossPhaseVictory {
		t.Error("should be in victory phase after winning")
	}
}

func TestEndBossFightDefeat(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	bmm.StartBossFight("generic")
	bmm.UpdateBossHealth(0.5) // Get to a later phase

	initialPhase := bmm.GetCurrentPhase()
	bmm.EndBossFight(false)

	if bmm.IsActive() {
		t.Error("should not be active after EndBossFight")
	}

	// Phase should remain unchanged on defeat (not victory)
	if bmm.GetCurrentPhase() != initialPhase {
		t.Error("phase should remain unchanged on defeat")
	}
}

func TestGetCurrentTempo(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	bmm.StartBossFight("generic")

	// Get tempo at different phases
	tempoIntro := bmm.GetCurrentTempo()
	bmm.UpdateBossHealth(0.49)
	tempoIntense := bmm.GetCurrentTempo()

	// Tempo should increase with phase intensity
	if tempoIntense <= tempoIntro {
		t.Errorf("tempo should increase with phase: intro=%f, intense=%f", tempoIntro, tempoIntense)
	}
}

func TestGetCurrentTempoNoBoss(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	// Don't start a boss fight

	tempo := bmm.GetCurrentTempo()
	// Should return default tempo
	if tempo != 120.0 {
		t.Errorf("expected default tempo 120.0, got %f", tempo)
	}
}

func TestGetCurrentIntensity(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	bmm.StartBossFight("generic")

	intensityIntro := bmm.GetCurrentIntensity()
	bmm.UpdateBossHealth(0.24)
	intensityFinal := bmm.GetCurrentIntensity()

	// Intensity should increase with phase
	if intensityFinal <= intensityIntro {
		t.Errorf("intensity should increase: intro=%f, final=%f", intensityIntro, intensityFinal)
	}

	// Intensity should be clamped to 1.0
	if intensityFinal > 1.0 {
		t.Errorf("intensity should be clamped to 1.0, got %f", intensityFinal)
	}
}

func TestGetCurrentIntensityNoBoss(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	// Don't start a boss fight

	intensity := bmm.GetCurrentIntensity()
	// Should return default intensity
	if intensity != 0.5 {
		t.Errorf("expected default intensity 0.5, got %f", intensity)
	}
}

func TestUpdateBossHealthClamping(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	bmm.StartBossFight("generic")

	// Test clamping to min
	bmm.UpdateBossHealth(-0.5)
	bmm.mu.RLock()
	health := bmm.bossHealth
	bmm.mu.RUnlock()
	if health != 0.0 {
		t.Errorf("health should be clamped to 0.0, got %f", health)
	}

	// Test clamping to max
	bmm.UpdateBossHealth(1.5)
	bmm.mu.RLock()
	health = bmm.bossHealth
	bmm.mu.RUnlock()
	if health != 1.0 {
		t.Errorf("health should be clamped to 1.0, got %f", health)
	}
}

func TestUpdateBossHealthWhenInactive(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")
	// Don't start boss fight, just update health

	bmm.UpdateBossHealth(0.5)

	// Should not cause any errors and should not trigger phase transitions
	if bmm.GetCurrentPhase() != BossPhaseIntro {
		t.Error("phase should remain at intro when inactive")
	}
}

func BenchmarkBossMusicPhaseTransition(b *testing.B) {
	bmm := NewBossMusicManager("fantasy")
	bmm.StartBossFight("generic")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		health := float64(i%100) / 100.0
		bmm.UpdateBossHealth(health)
		bmm.GetCurrentTempo()
		bmm.GetCurrentIntensity()
	}
}

// ============================================================================
// Dynamic Layer Manager Tests
// ============================================================================

func TestNewDynamicLayerManager(t *testing.T) {
	dlm := NewDynamicLayerManager()
	if dlm == nil {
		t.Fatal("NewDynamicLayerManager returned nil")
	}

	// Check default layers exist
	expectedLayers := []string{
		"exploration_base",
		"combat_percussion",
		"combat_strings",
		"tense_drone",
		"victory_fanfare",
	}
	for _, name := range expectedLayers {
		if _, ok := dlm.layers[name]; !ok {
			t.Errorf("missing default layer: %s", name)
		}
	}

	// Master volume should be 1.0
	if dlm.masterVolume != 1.0 {
		t.Errorf("expected master volume 1.0, got %f", dlm.masterVolume)
	}
}

func TestDynamicLayerSetState(t *testing.T) {
	dlm := NewDynamicLayerManager()

	// Activate exploration state
	dlm.SetState(StateExploration, true)

	// Exploration layer should have a target volume
	dlm.mu.RLock()
	target := dlm.targetVolumes["exploration_base"]
	dlm.mu.RUnlock()

	if target <= 0 {
		t.Error("exploration_base should have target volume when exploration state is active")
	}
}

func TestDynamicLayerFadeIn(t *testing.T) {
	dlm := NewDynamicLayerManager()

	// Activate combat state
	dlm.SetState(StateCombat, true)

	// Initial volume should be 0
	initialVol := dlm.GetLayerVolume("combat_percussion")
	if initialVol != 0 {
		t.Errorf("initial volume should be 0, got %f", initialVol)
	}

	// Update for a while to fade in
	for i := 0; i < 50; i++ {
		dlm.Update(0.02) // 20ms per update = 1 second total
	}

	// Volume should have increased
	vol := dlm.GetLayerVolume("combat_percussion")
	if vol <= initialVol {
		t.Errorf("volume should increase during fade in: initial=%f, current=%f", initialVol, vol)
	}
}

func TestDynamicLayerFadeOut(t *testing.T) {
	dlm := NewDynamicLayerManager()

	// Activate then deactivate combat
	dlm.SetState(StateCombat, true)

	// Force volume to max for testing
	dlm.mu.Lock()
	dlm.layerVolumes["combat_percussion"] = 0.8
	dlm.mu.Unlock()

	// Deactivate
	dlm.SetState(StateCombat, false)

	// Update for a while to fade out
	for i := 0; i < 100; i++ {
		dlm.Update(0.02)
	}

	// Volume should have decreased toward 0
	vol := dlm.GetLayerVolume("combat_percussion")
	if vol >= 0.8 {
		t.Errorf("volume should decrease during fade out, got %f", vol)
	}
}

func TestDynamicLayerSetMasterVolume(t *testing.T) {
	dlm := NewDynamicLayerManager()

	dlm.SetMasterVolume(0.5)

	dlm.mu.RLock()
	master := dlm.masterVolume
	dlm.mu.RUnlock()

	if master != 0.5 {
		t.Errorf("expected master volume 0.5, got %f", master)
	}
}

func TestDynamicLayerSetMasterVolumeClamping(t *testing.T) {
	dlm := NewDynamicLayerManager()

	dlm.SetMasterVolume(-0.5)
	dlm.mu.RLock()
	master := dlm.masterVolume
	dlm.mu.RUnlock()
	if master != 0 {
		t.Errorf("master volume should clamp to 0, got %f", master)
	}

	dlm.SetMasterVolume(1.5)
	dlm.mu.RLock()
	master = dlm.masterVolume
	dlm.mu.RUnlock()
	if master != 1 {
		t.Errorf("master volume should clamp to 1, got %f", master)
	}
}

func TestDynamicLayerMasterVolumeAffectsOutput(t *testing.T) {
	dlm := NewDynamicLayerManager()

	// Set a layer volume directly for testing
	dlm.mu.Lock()
	dlm.layerVolumes["exploration_base"] = 0.6
	dlm.mu.Unlock()

	// With master at 1.0
	dlm.SetMasterVolume(1.0)
	vol1 := dlm.GetLayerVolume("exploration_base")

	// With master at 0.5
	dlm.SetMasterVolume(0.5)
	vol2 := dlm.GetLayerVolume("exploration_base")

	if vol2 >= vol1 {
		t.Errorf("lower master volume should reduce output: vol1=%f, vol2=%f", vol1, vol2)
	}

	// Should be exactly half
	if vol2 != vol1*0.5 {
		t.Errorf("expected vol2 = %f, got %f", vol1*0.5, vol2)
	}
}

func TestDynamicLayerGetActiveLayersByTag(t *testing.T) {
	dlm := NewDynamicLayerManager()

	// Set combat layers as active
	dlm.mu.Lock()
	dlm.layerVolumes["combat_percussion"] = 0.8
	dlm.layerVolumes["combat_strings"] = 0.7
	dlm.mu.Unlock()

	combatLayers := dlm.GetActiveLayersByTag("combat")
	if len(combatLayers) != 2 {
		t.Errorf("expected 2 combat layers, got %d", len(combatLayers))
	}

	// Check both layers are in result
	found := map[string]bool{}
	for _, name := range combatLayers {
		found[name] = true
	}
	if !found["combat_percussion"] || !found["combat_strings"] {
		t.Error("should find both combat layers")
	}
}

func TestDynamicLayerGetActiveLayersByTagNoMatches(t *testing.T) {
	dlm := NewDynamicLayerManager()

	// All layers start at 0 volume
	layers := dlm.GetActiveLayersByTag("combat")
	if len(layers) != 0 {
		t.Errorf("expected 0 layers when none active, got %d", len(layers))
	}
}

func TestDynamicLayerAddLayer(t *testing.T) {
	dlm := NewDynamicLayerManager()

	customConfig := &DynamicLayerConfig{
		Name:         "custom_layer",
		BaseVolume:   0.7,
		TriggerState: StateTense,
		FadeInTime:   1.0,
		FadeOutTime:  1.0,
		LoopEnabled:  true,
		Priority:     4,
		Tags:         []string{"custom", "test"},
	}

	dlm.AddLayer(customConfig)

	dlm.mu.RLock()
	_, ok := dlm.layers["custom_layer"]
	dlm.mu.RUnlock()

	if !ok {
		t.Error("custom layer should be added")
	}

	// Activate tense state
	dlm.SetState(StateTense, true)

	dlm.mu.RLock()
	target := dlm.targetVolumes["custom_layer"]
	dlm.mu.RUnlock()

	if target != 0.7 {
		t.Errorf("custom layer target should be 0.7, got %f", target)
	}
}

func TestDynamicLayerMultipleStates(t *testing.T) {
	dlm := NewDynamicLayerManager()

	// Activate both exploration and tense
	dlm.SetState(StateExploration, true)
	dlm.SetState(StateTense, true)

	dlm.mu.RLock()
	explorationTarget := dlm.targetVolumes["exploration_base"]
	tenseTarget := dlm.targetVolumes["tense_drone"]
	dlm.mu.RUnlock()

	if explorationTarget <= 0 {
		t.Error("exploration layer should have target when state active")
	}
	if tenseTarget <= 0 {
		t.Error("tense layer should have target when state active")
	}
}

func BenchmarkDynamicLayerUpdate(b *testing.B) {
	dlm := NewDynamicLayerManager()
	dlm.SetState(StateCombat, true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dlm.Update(0.016)
	}
}

// ============================================================================
// Additional Edge Case Tests
// ============================================================================

func TestStateTenseAndVictoryDefeat(t *testing.T) {
	// Test that StateTense, StateVictory, StateDefeat exist and can be set
	states := []State{StateExploration, StateCombat, StateTense, StateVictory, StateDefeat}

	for i, state := range states {
		if int(state) != i {
			t.Errorf("state %d should equal %d", state, i)
		}
	}
}

func TestMotifGeneration(t *testing.T) {
	am := NewAdaptiveMusic("horror", 42)

	// Verify horror has distinctive characteristics
	explorationMotif := am.motifs["exploration"]
	if explorationMotif.BaseFreq != FreqA1 {
		t.Errorf("horror exploration should use low frequency A1, got %f", explorationMotif.BaseFreq)
	}

	// Check that motif has notes
	if len(explorationMotif.Notes) == 0 {
		t.Error("motif should have notes")
	}
	if len(explorationMotif.Durations) == 0 {
		t.Error("motif should have durations")
	}
	if len(explorationMotif.Notes) != len(explorationMotif.Durations) {
		t.Error("notes and durations should have same length")
	}
}

func TestGetLayerVolumeNonexistent(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	vol := am.GetLayerVolume("nonexistent_layer")
	if vol != 0.0 {
		t.Errorf("nonexistent layer should return 0.0, got %f", vol)
	}
}

func TestNormalizeSamples(t *testing.T) {
	// Test that clipping is prevented
	samples := []float64{0.5, 1.5, -2.0, 0.3, 2.5}
	normalizeSamples(samples)

	for i, s := range samples {
		if s > 1.0 || s < -1.0 {
			t.Errorf("sample %d should be normalized: %f", i, s)
		}
	}
}

func TestNormalizeSamplesNoClip(t *testing.T) {
	// When no clipping needed, samples unchanged
	original := []float64{0.5, 0.3, -0.7, 0.2}
	samples := make([]float64, len(original))
	copy(samples, original)

	normalizeSamples(samples)

	for i := range samples {
		if samples[i] != original[i] {
			t.Errorf("samples should be unchanged when no clipping: %f != %f", samples[i], original[i])
		}
	}
}

func TestClampFloat(t *testing.T) {
	tests := []struct {
		val, min, max, expected float64
	}{
		{0.5, 0.0, 1.0, 0.5},
		{-0.5, 0.0, 1.0, 0.0},
		{1.5, 0.0, 1.0, 1.0},
		{0.0, 0.0, 1.0, 0.0},
		{1.0, 0.0, 1.0, 1.0},
	}

	for _, tc := range tests {
		result := clampFloat(tc.val, tc.min, tc.max)
		if result != tc.expected {
			t.Errorf("clampFloat(%f, %f, %f) = %f, expected %f",
				tc.val, tc.min, tc.max, result, tc.expected)
		}
	}
}

func TestDefaultGenreFallback(t *testing.T) {
	// Test that unknown genre falls back to default motifs
	am := NewAdaptiveMusic("unknown_genre", 42)

	if am.motifs["exploration"].Genre != "default" {
		t.Errorf("unknown genre should use default motifs, got %s", am.motifs["exploration"].Genre)
	}
}

func TestLocationMusicMissingConfig(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	lmm := NewLocationMusicManager(am)

	// Remove a config to test edge case
	lmm.mu.Lock()
	delete(lmm.configs, LocationTown)
	lmm.mu.Unlock()

	// Should use default transition time
	lmm.SetLocation(LocationTown)

	lmm.mu.RLock()
	timer := lmm.transitionTimer
	lmm.mu.RUnlock()

	if timer != 2.0 {
		t.Errorf("missing config should use default 2.0s, got %f", timer)
	}
}

func TestConcurrentAccess(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	// Test concurrent access to adaptive music
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				am.EnterCombat()
				am.Update(0.016)
				_ = am.GetCurrentState()
				_ = am.GetLayerVolume("combat")
				am.ExitCombat()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestBossMusicConcurrentAccess(t *testing.T) {
	bmm := NewBossMusicManager("fantasy")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				bmm.StartBossFight("generic")
				bmm.UpdateBossHealth(float64(j%100) / 100.0)
				_ = bmm.GetCurrentTempo()
				_ = bmm.GetCurrentIntensity()
				bmm.EndBossFight(j%2 == 0)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestDynamicLayerConcurrentAccess(t *testing.T) {
	dlm := NewDynamicLayerManager()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				dlm.SetState(StateCombat, j%2 == 0)
				dlm.Update(0.016)
				_ = dlm.GetLayerVolume("combat_percussion")
				_ = dlm.GetActiveLayersByTag("combat")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// ============================================================================
// Menu Music Tests
// ============================================================================

func TestEnterMenu(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	am.EnterMenu()

	if am.GetCurrentState() != StateMenu {
		t.Errorf("expected StateMenu, got %v", am.GetCurrentState())
	}

	// Menu layer should be fading in
	am.mu.Lock()
	menuTarget := am.layers["menu"].Target
	menuActive := am.layers["menu"].Active
	am.mu.Unlock()

	if menuTarget != MaxVolume {
		t.Errorf("menu target should be %f, got %f", MaxVolume, menuTarget)
	}
	if !menuActive {
		t.Error("menu layer should be active")
	}
}

func TestEnterMenu_FadesOutGameplay(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	// Simulate gameplay first
	am.EnterCombat()

	// Then enter menu
	am.EnterMenu()

	am.mu.Lock()
	explorationTarget := am.layers["exploration"].Target
	combatTarget := am.layers["combat"].Target
	am.mu.Unlock()

	if explorationTarget != 0.0 {
		t.Errorf("exploration should fade out to 0, got %f", explorationTarget)
	}
	if combatTarget != 0.0 {
		t.Errorf("combat should fade out to 0, got %f", combatTarget)
	}
}

func TestEnterPauseMenu(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	am.EnterPauseMenu()

	if am.GetCurrentState() != StatePauseMenu {
		t.Errorf("expected StatePauseMenu, got %v", am.GetCurrentState())
	}

	am.mu.Lock()
	menuTarget := am.layers["menu"].Target
	explorationTarget := am.layers["exploration"].Target
	am.mu.Unlock()

	// Pause menu has reduced menu music
	if menuTarget != MenuMusicVolume {
		t.Errorf("pause menu music should be at %f, got %f", MenuMusicVolume, menuTarget)
	}
	// Exploration music should be reduced, not muted
	if explorationTarget != MenuMusicReduction {
		t.Errorf("exploration should be reduced to %f, got %f", MenuMusicReduction, explorationTarget)
	}
}

func TestExitMenu(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	// Enter menu
	am.EnterMenu()
	if am.GetCurrentState() != StateMenu {
		t.Fatal("should be in menu state")
	}

	// Exit menu
	am.ExitMenu()

	// Should return to exploration
	if am.GetCurrentState() != StateExploration {
		t.Errorf("expected StateExploration after exit, got %v", am.GetCurrentState())
	}

	// Menu layer should be fading out
	am.mu.Lock()
	menuTarget := am.layers["menu"].Target
	am.mu.Unlock()

	if menuTarget != 0.0 {
		t.Errorf("menu should fade out to 0, got %f", menuTarget)
	}
}

func TestExitMenu_RestoresCombat(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	// Enter combat first
	am.EnterCombat()

	// Enter pause menu (preserves previous state)
	am.EnterPauseMenu()

	// Exit menu - should restore combat
	am.ExitMenu()

	if am.GetCurrentState() != StateCombat {
		t.Errorf("expected StateCombat after exit from pause, got %v", am.GetCurrentState())
	}

	am.mu.Lock()
	combatTarget := am.layers["combat"].Target
	combatActive := am.layers["combat"].Active
	am.mu.Unlock()

	if combatTarget != MaxVolume {
		t.Errorf("combat should restore to max, got %f", combatTarget)
	}
	if !combatActive {
		t.Error("combat layer should be active")
	}
}

func TestIsInMenu(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	if am.IsInMenu() {
		t.Error("should not be in menu initially")
	}

	am.EnterMenu()
	if !am.IsInMenu() {
		t.Error("should be in menu after EnterMenu")
	}

	am.ExitMenu()
	if am.IsInMenu() {
		t.Error("should not be in menu after ExitMenu")
	}
}

func TestIsInMenu_PauseMenu(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	am.EnterPauseMenu()
	if !am.IsInMenu() {
		t.Error("IsInMenu should return true for pause menu")
	}
}

func TestGenerateMenuMusic(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			am := NewAdaptiveMusic(genre, 42)
			am.EnterMenu()

			samples := am.GenerateMenuMusic(0.5)

			expectedLen := int(0.5 * float64(DefaultSampleRate))
			if len(samples) != expectedLen {
				t.Errorf("expected %d samples, got %d", expectedLen, len(samples))
			}

			// Should have non-zero samples
			hasNonZero := false
			for _, s := range samples {
				if s != 0 {
					hasNonZero = true
					break
				}
			}
			if !hasNonZero {
				t.Error("menu music should produce non-zero samples")
			}
		})
	}
}

func TestMenuStatesPreservePrevious(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	// Enter combat
	am.EnterCombat()
	previousState := am.GetCurrentState()

	// Enter menu
	am.EnterMenu()

	am.mu.Lock()
	storedPrevious := am.previousState
	am.mu.Unlock()

	if storedPrevious != previousState {
		t.Error("entering menu should preserve previous state")
	}
}

func TestExitMenuFromNonMenu(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	// Calling ExitMenu when not in menu should do nothing
	am.ExitMenu()

	// Should still be in exploration
	if am.GetCurrentState() != StateExploration {
		t.Error("state should remain unchanged when ExitMenu called outside menu")
	}
}

func TestMenuLayerExists(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	am.mu.Lock()
	_, exists := am.layers["menu"]
	am.mu.Unlock()

	if !exists {
		t.Error("menu layer should be initialized")
	}
}

func TestGetMenuBaseFrequency(t *testing.T) {
	testCases := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", FreqA3 * 0.5},
		{"sci-fi", FreqE4},
		{"horror", FreqA1},
		{"cyberpunk", FreqA4 * 0.75},
		{"post-apocalyptic", FreqE3 * 0.75},
		{"unknown", FreqA3 * 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.genre, func(t *testing.T) {
			am := NewAdaptiveMusic(tc.genre, 42)
			freq := am.getMenuBaseFrequency()

			if freq != tc.expected {
				t.Errorf("expected %f, got %f", tc.expected, freq)
			}
		})
	}
}

func TestGetMenuMotif(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)
	motif := am.getMenuMotif()

	if motif == nil {
		t.Fatal("getMenuMotif should not return nil")
	}
	if len(motif.Notes) == 0 {
		t.Error("motif should have notes")
	}
	if len(motif.Durations) == 0 {
		t.Error("motif should have durations")
	}
}

func TestGetMenuEnvelope(t *testing.T) {
	am := NewAdaptiveMusic("fantasy", 42)

	// Test attack phase (first 20%)
	env := am.getMenuEnvelope(0, 100)
	if env != 0 {
		t.Errorf("envelope at start should be 0, got %f", env)
	}

	env = am.getMenuEnvelope(10, 100)
	if env != 0.5 {
		t.Errorf("envelope at 10%% should be 0.5, got %f", env)
	}

	// Test sustain phase (20%-70%)
	env = am.getMenuEnvelope(50, 100)
	if env != 1.0 {
		t.Errorf("envelope in sustain should be 1.0, got %f", env)
	}

	// Test release phase (70%-100%)
	env = am.getMenuEnvelope(85, 100)
	if env < 0 || env > 1 {
		t.Errorf("envelope in release should be between 0-1, got %f", env)
	}
}

func BenchmarkGenerateMenuMusic(b *testing.B) {
	am := NewAdaptiveMusic("fantasy", 42)
	am.EnterMenu()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = am.GenerateMenuMusic(0.1)
	}
}
