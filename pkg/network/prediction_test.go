package network

import (
	"testing"
	"time"
)

func TestNewClientPredictor(t *testing.T) {
	cp := NewClientPredictor()
	if cp == nil {
		t.Fatal("NewClientPredictor returned nil")
	}
	if cp.PendingInputCount() != 0 {
		t.Error("should start with no pending inputs")
	}
}

func TestRecordInput(t *testing.T) {
	cp := NewClientPredictor()
	cp.SetPosition(Position3D{X: 0, Y: 0, Z: 0})

	input := &PlayerInput{
		MoveForward: 1.0,
		MoveRight:   0.0,
		Turn:        0.0,
	}

	seq := cp.RecordInput(input, 0.016)

	if seq != 0 {
		t.Errorf("first sequence should be 0, got %d", seq)
	}
	if cp.PendingInputCount() != 1 {
		t.Errorf("should have 1 pending input, got %d", cp.PendingInputCount())
	}

	// Position should have changed
	pos := cp.GetPredictedPosition()
	if pos.X == 0 && pos.Z == 0 {
		t.Error("position should have changed after input")
	}
}

func TestReconcile(t *testing.T) {
	cp := NewClientPredictor()
	cp.SetPosition(Position3D{X: 0, Y: 0, Z: 0})

	// Record some inputs
	for i := 0; i < 5; i++ {
		input := &PlayerInput{
			MoveForward: 1.0,
		}
		cp.RecordInput(input, 0.016)
	}

	// Server acknowledges first 3 inputs
	serverState := &WorldState{
		AckSequence: 2,
		Entities: []EntityState{
			{
				EntityID: 1,
				X:        10.0,
				Y:        0,
				Z:        5.0,
				Angle:    0,
			},
		},
	}

	cp.Reconcile(serverState, 1)

	// Should only have 2 pending inputs (seq 3 and 4)
	if cp.PendingInputCount() != 2 {
		t.Errorf("should have 2 pending inputs after reconcile, got %d", cp.PendingInputCount())
	}
}

func TestPredictedPositionAfterReconcile(t *testing.T) {
	cp := NewClientPredictor()
	cp.SetPosition(Position3D{X: 0, Y: 0, Z: 0})

	// Record input
	input := &PlayerInput{
		MoveForward: 1.0,
	}
	cp.RecordInput(input, 0.1)

	// Server sets position
	serverState := &WorldState{
		AckSequence: 0,
		Entities: []EntityState{
			{
				EntityID: 1,
				X:        100.0,
				Y:        0,
				Z:        50.0,
			},
		},
	}

	cp.Reconcile(serverState, 1)

	// Position should be server position (since input was acknowledged)
	pos := cp.GetPredictedPosition()
	if pos.X != 100.0 || pos.Z != 50.0 {
		t.Errorf("position should match server after reconcile: got (%f, %f)", pos.X, pos.Z)
	}
}

func TestInputBufferLimit(t *testing.T) {
	cp := NewClientPredictor()

	// Record more than buffer limit
	for i := 0; i < 200; i++ {
		input := &PlayerInput{
			MoveForward: 0.1,
		}
		cp.RecordInput(input, 0.016)
	}

	// Should be capped at 128
	if cp.PendingInputCount() > 128 {
		t.Errorf("pending inputs should be capped at 128, got %d", cp.PendingInputCount())
	}
}

func TestNewLagCompensator(t *testing.T) {
	lc := NewLagCompensator()
	if lc == nil {
		t.Fatal("NewLagCompensator returned nil")
	}
	if lc.EntityCount() != 0 {
		t.Error("should start with no tracked entities")
	}
}

func TestRecordEntityState(t *testing.T) {
	lc := NewLagCompensator()

	pos := Position3D{X: 10, Y: 0, Z: 20}
	hitMin := Position3D{X: -0.5, Y: 0, Z: -0.5}
	hitMax := Position3D{X: 0.5, Y: 2, Z: 0.5}

	lc.RecordEntityState(1, pos, 45.0, hitMin, hitMax)

	if lc.EntityCount() != 1 {
		t.Errorf("should have 1 tracked entity, got %d", lc.EntityCount())
	}
}

func TestRemoveEntity(t *testing.T) {
	lc := NewLagCompensator()

	lc.RecordEntityState(1, Position3D{}, 0, Position3D{}, Position3D{})
	lc.RecordEntityState(2, Position3D{}, 0, Position3D{}, Position3D{})

	lc.RemoveEntity(1)

	if lc.EntityCount() != 1 {
		t.Errorf("should have 1 tracked entity after removal, got %d", lc.EntityCount())
	}
}

func TestHitTestMiss(t *testing.T) {
	lc := NewLagCompensator()

	// Record target at position (10, 0, 10)
	targetPos := Position3D{X: 10, Y: 0, Z: 10}
	hitMin := Position3D{X: -0.5, Y: 0, Z: -0.5}
	hitMax := Position3D{X: 0.5, Y: 2, Z: 0.5}
	lc.RecordEntityState(2, targetPos, 0, hitMin, hitMax)

	// Shoot in wrong direction
	origin := Position3D{X: 0, Y: 1, Z: 0}
	direction := Position3D{X: 0, Y: 0, Z: -1} // Shooting away from target

	result := lc.HitTest(1, 2, origin, direction, time.Now(), 100*time.Millisecond)

	if result.Hit {
		t.Error("shot in wrong direction should miss")
	}
}

func TestHitTestHit(t *testing.T) {
	lc := NewLagCompensator()

	// Record target at position (10, 0, 0)
	targetPos := Position3D{X: 10, Y: 1, Z: 0}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, targetPos, 0, hitMin, hitMax)

	// Shoot toward target
	origin := Position3D{X: 0, Y: 1, Z: 0}
	direction := Position3D{X: 1, Y: 0, Z: 0} // Shooting toward target

	result := lc.HitTest(1, 2, origin, direction, time.Now(), 100*time.Millisecond)

	if !result.Hit {
		t.Error("shot toward target should hit")
	}
	if result.TargetID != 2 {
		t.Errorf("hit should report target ID 2, got %d", result.TargetID)
	}
}

func TestHitTestWithRewind(t *testing.T) {
	lc := NewLagCompensator()

	// Record target at old position
	oldPos := Position3D{X: 10, Y: 1, Z: 0}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, oldPos, 0, hitMin, hitMax)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Record target at new position (moved)
	newPos := Position3D{X: 20, Y: 1, Z: 0}
	lc.RecordEntityState(2, newPos, 0, hitMin, hitMax)

	// Client shot at the OLD position (client time was before movement)
	clientTime := time.Now().Add(-5 * time.Millisecond)
	origin := Position3D{X: 0, Y: 1, Z: 0}
	direction := Position3D{X: 1, Y: 0, Z: 0}

	result := lc.HitTest(1, 2, origin, direction, clientTime, 100*time.Millisecond)

	// Should hit because we rewind to when target was at old position
	if !result.Hit {
		t.Error("lag-compensated shot should hit old position")
	}
}

func TestHitTestExceedsMaxRewind(t *testing.T) {
	lc := NewLagCompensator()

	// Record target
	pos := Position3D{X: 10, Y: 1, Z: 0}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, pos, 0, hitMin, hitMax)

	// Client time is way too old
	clientTime := time.Now().Add(-1 * time.Second) // 1 second ago, exceeds 500ms max
	origin := Position3D{X: 0, Y: 1, Z: 0}
	direction := Position3D{X: 1, Y: 0, Z: 0}

	result := lc.HitTest(1, 2, origin, direction, clientTime, 100*time.Millisecond)

	// Should still try with clamped rewind time
	// The hit test logic clamps to MaxRewindTime
	if result.CompensatedBy > MaxRewindTime+50*time.Millisecond {
		t.Error("compensation should be clamped to MaxRewindTime")
	}
}

func TestIsTorMode(t *testing.T) {
	lc := NewLagCompensator()

	if lc.IsTorMode(500 * time.Millisecond) {
		t.Error("500ms RTT should not be Tor mode")
	}
	if !lc.IsTorMode(900 * time.Millisecond) {
		t.Error("900ms RTT should be Tor mode")
	}
}

func TestClientPredictorTorMode(t *testing.T) {
	cp := NewClientPredictor()

	// Initially not in Tor mode
	if cp.IsTorMode() {
		t.Error("should not start in Tor mode")
	}

	// Normal mode settings
	if cp.GetInputRateHz() != NormalInputRate {
		t.Errorf("normal input rate should be %d, got %d", NormalInputRate, cp.GetInputRateHz())
	}
	if cp.GetPredictionWindow() != NormalPredictionWindow {
		t.Errorf("normal prediction window should be %v, got %v", NormalPredictionWindow, cp.GetPredictionWindow())
	}
	if cp.GetInterpolationBlend() != NormalBlendTime {
		t.Errorf("normal blend time should be %v, got %v", NormalBlendTime, cp.GetInterpolationBlend())
	}
}

func TestTorModeThresholds(t *testing.T) {
	// Verify threshold constants
	if TorModeThreshold != 800*time.Millisecond {
		t.Errorf("TorModeThreshold should be 800ms, got %v", TorModeThreshold)
	}
	if TorModePredictionWindow != 1500*time.Millisecond {
		t.Errorf("TorModePredictionWindow should be 1500ms, got %v", TorModePredictionWindow)
	}
	if TorModeInputRate != 10 {
		t.Errorf("TorModeInputRate should be 10 Hz, got %d", TorModeInputRate)
	}
	if TorModeBlendTime != 300*time.Millisecond {
		t.Errorf("TorModeBlendTime should be 300ms, got %v", TorModeBlendTime)
	}
}

// TestHighLatencyModeThresholds verifies the high-latency mode constants.
// Per README: Supports 200-5000ms latency tolerance.
func TestHighLatencyModeThresholds(t *testing.T) {
	// Verify high-latency threshold constants
	if HighLatencyThreshold != 3000*time.Millisecond {
		t.Errorf("HighLatencyThreshold should be 3000ms, got %v", HighLatencyThreshold)
	}
	if ExtremeLatencyThreshold != 5000*time.Millisecond {
		t.Errorf("ExtremeLatencyThreshold should be 5000ms, got %v", ExtremeLatencyThreshold)
	}

	// Verify prediction windows scale appropriately
	if HighLatencyPredictionWindow != 3500*time.Millisecond {
		t.Errorf("HighLatencyPredictionWindow should be 3500ms, got %v", HighLatencyPredictionWindow)
	}
	if ExtremeLatencyPredictionWindow != 6000*time.Millisecond {
		t.Errorf("ExtremeLatencyPredictionWindow should be 6000ms, got %v", ExtremeLatencyPredictionWindow)
	}

	// Verify input rates decrease at higher latencies
	if HighLatencyInputRate != 5 {
		t.Errorf("HighLatencyInputRate should be 5 Hz, got %d", HighLatencyInputRate)
	}
	if ExtremeLatencyInputRate != 2 {
		t.Errorf("ExtremeLatencyInputRate should be 2 Hz, got %d", ExtremeLatencyInputRate)
	}
}

// TestLatencyModeClassification tests the latency mode classification logic.
func TestLatencyModeClassification(t *testing.T) {
	cp := NewClientPredictor()

	tests := []struct {
		rtt      time.Duration
		expected LatencyMode
	}{
		{200 * time.Millisecond, LatencyModeNormal},
		{500 * time.Millisecond, LatencyModeNormal},
		{799 * time.Millisecond, LatencyModeNormal},
		{800 * time.Millisecond, LatencyModeTor},
		{1500 * time.Millisecond, LatencyModeTor},
		{2999 * time.Millisecond, LatencyModeTor},
		{3000 * time.Millisecond, LatencyModeHigh},
		{4000 * time.Millisecond, LatencyModeHigh},
		{4999 * time.Millisecond, LatencyModeHigh},
		{5000 * time.Millisecond, LatencyModeExtreme},
		{7000 * time.Millisecond, LatencyModeExtreme},
		{10000 * time.Millisecond, LatencyModeExtreme},
	}

	for _, tt := range tests {
		mode := cp.classifyLatency(tt.rtt)
		if mode != tt.expected {
			t.Errorf("classifyLatency(%v) = %d, want %d", tt.rtt, mode, tt.expected)
		}
	}
}

// TestHighLatencyAdaptation tests that prediction parameters adapt at 5000ms RTT.
func TestHighLatencyAdaptation(t *testing.T) {
	cp := NewClientPredictor()

	// Verify initial state is normal mode
	if cp.GetLatencyMode() != LatencyModeNormal {
		t.Errorf("initial mode should be Normal, got %d", cp.GetLatencyMode())
	}

	// Simulate extreme latency by setting smoothedRTT directly
	cp.mu.Lock()
	cp.smoothedRTT = 5500 * time.Millisecond
	cp.adaptToLatency()
	cp.mu.Unlock()

	// Verify extreme latency mode is active
	if cp.GetLatencyMode() != LatencyModeExtreme {
		t.Errorf("mode should be Extreme at 5500ms RTT, got %d", cp.GetLatencyMode())
	}

	// Verify prediction window is set for extreme latency
	if cp.GetPredictionWindow() != ExtremeLatencyPredictionWindow {
		t.Errorf("prediction window should be %v, got %v",
			ExtremeLatencyPredictionWindow, cp.GetPredictionWindow())
	}

	// Verify input rate is reduced
	if cp.GetInputRateHz() != ExtremeLatencyInputRate {
		t.Errorf("input rate should be %d Hz, got %d Hz",
			ExtremeLatencyInputRate, cp.GetInputRateHz())
	}

	// Verify buffer size is increased
	if cp.GetMaxPendingInputs() != 512 {
		t.Errorf("max pending inputs should be 512 in extreme mode, got %d",
			cp.GetMaxPendingInputs())
	}

	// Verify IsTorMode returns true (backwards compatibility)
	if !cp.IsTorMode() {
		t.Error("IsTorMode should return true for extreme latency mode")
	}
}

// TestLagCompensatorHighLatencyRewind tests rewind time increases with RTT.
func TestLagCompensatorHighLatencyRewind(t *testing.T) {
	lc := NewLagCompensator()

	tests := []struct {
		rtt        time.Duration
		maxRewind  time.Duration
		latencyMod LatencyMode
	}{
		{200 * time.Millisecond, MaxRewindTimeNormal, LatencyModeNormal},
		{1000 * time.Millisecond, MaxRewindTimeTor, LatencyModeTor},
		{3500 * time.Millisecond, MaxRewindTimeHigh, LatencyModeHigh},
		{6000 * time.Millisecond, MaxRewindTimeExtreme, LatencyModeExtreme},
	}

	for _, tt := range tests {
		maxRewind := lc.getMaxRewindForRTT(tt.rtt)
		if maxRewind != tt.maxRewind {
			t.Errorf("getMaxRewindForRTT(%v) = %v, want %v", tt.rtt, maxRewind, tt.maxRewind)
		}

		mode := lc.GetLatencyMode(tt.rtt)
		if mode != tt.latencyMod {
			t.Errorf("GetLatencyMode(%v) = %d, want %d", tt.rtt, mode, tt.latencyMod)
		}
	}
}

func TestShouldSendInput(t *testing.T) {
	cp := NewClientPredictor()

	// First call should always return true (initialize lastInputTime)
	if !cp.ShouldSendInput() {
		t.Error("first ShouldSendInput should return true")
	}

	// Immediate second call should return false (rate limited)
	if cp.ShouldSendInput() {
		t.Error("immediate second ShouldSendInput should return false (rate limited)")
	}

	// After waiting for the interval, should return true
	time.Sleep(time.Second/time.Duration(NormalInputRate) + 2*time.Millisecond)
	if !cp.ShouldSendInput() {
		t.Error("ShouldSendInput should return true after waiting for interval")
	}
}

func TestStateHistoryRingBuffer(t *testing.T) {
	sh := NewStateHistory()

	// Record more than buffer size
	for i := 0; i < HistoryBufferSize+10; i++ {
		sh.Record(EntitySnapshot{
			EntityID:  1,
			Timestamp: time.Now(),
			Position:  Position3D{X: float32(i), Y: 0, Z: 0},
		})
	}

	// Count should be capped
	if sh.count != HistoryBufferSize {
		t.Errorf("count should be %d, got %d", HistoryBufferSize, sh.count)
	}
}

func BenchmarkRecordInput(b *testing.B) {
	cp := NewClientPredictor()
	input := &PlayerInput{
		MoveForward: 1.0,
		MoveRight:   0.5,
		Turn:        0.1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cp.RecordInput(input, 0.016)
	}
}

func BenchmarkRecordEntityState(b *testing.B) {
	lc := NewLagCompensator()
	pos := Position3D{X: 10, Y: 0, Z: 20}
	hitMin := Position3D{X: -0.5, Y: 0, Z: -0.5}
	hitMax := Position3D{X: 0.5, Y: 2, Z: 0.5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.RecordEntityState(1, pos, 45.0, hitMin, hitMax)
	}
}

func BenchmarkHitTest(b *testing.B) {
	lc := NewLagCompensator()
	pos := Position3D{X: 10, Y: 1, Z: 0}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, pos, 0, hitMin, hitMax)

	origin := Position3D{X: 0, Y: 1, Z: 0}
	direction := Position3D{X: 1, Y: 0, Z: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.HitTest(1, 2, origin, direction, time.Now(), 100*time.Millisecond)
	}
}

// ============================================================================
// PvP Validator Tests
// ============================================================================

func TestNewPvPValidator(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)
	if pv == nil {
		t.Fatal("NewPvPValidator returned nil")
	}

	// Verify default damage rates are set
	if pv.maxDamageRates[PvPMeleeAttack] <= 0 {
		t.Error("melee damage rate should be set")
	}
	if pv.maxDamageRates[PvPRangedAttack] <= 0 {
		t.Error("ranged damage rate should be set")
	}
	if pv.maxDamageRates[PvPMagicAttack] <= 0 {
		t.Error("magic damage rate should be set")
	}
}

func TestPvPValidatorZoneConfig(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	// Initially not enabled
	if pv.IsZonePvPEnabled("test_zone") {
		t.Error("zone should not be PvP enabled by default")
	}

	// Enable PvP
	pv.SetZonePvPEnabled("test_zone", true)
	if !pv.IsZonePvPEnabled("test_zone") {
		t.Error("zone should be PvP enabled after SetZonePvPEnabled(true)")
	}

	// Disable PvP
	pv.SetZonePvPEnabled("test_zone", false)
	if pv.IsZonePvPEnabled("test_zone") {
		t.Error("zone should not be PvP enabled after SetZonePvPEnabled(false)")
	}
}

func TestPvPValidatorValidateActionZoneNotEnabled(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	action := &PvPCombatAction{
		AttackerID:  1,
		TargetID:    2,
		ActionType:  PvPMeleeAttack,
		DamageClaim: 10.0,
		ClientTime:  time.Now(),
	}

	result := pv.ValidateAction(action, 100*time.Millisecond, "disabled_zone")

	if result.Valid {
		t.Error("action should be invalid in non-PvP zone")
	}
	if result.RejectionReason != "PvP not enabled in zone" {
		t.Errorf("wrong rejection reason: %s", result.RejectionReason)
	}
}

func TestPvPValidatorValidateActionDamageRateExceeded(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	pv.SetZonePvPEnabled("pvp_zone", true)

	// Set up target entity for hit detection
	targetPos := Position3D{X: 10, Y: 1, Z: 0}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, targetPos, 0, hitMin, hitMax)

	action := &PvPCombatAction{
		AttackerID:  1,
		TargetID:    2,
		ActionType:  PvPMeleeAttack,
		DamageClaim: 1000.0, // Way over the limit
		ClientTime:  time.Now(),
		Position:    Position3D{X: 0, Y: 1, Z: 0},
		Direction:   Position3D{X: 1, Y: 0, Z: 0},
	}

	result := pv.ValidateAction(action, 100*time.Millisecond, "pvp_zone")

	if result.Valid {
		t.Error("action should be invalid with excessive damage")
	}
	if result.RejectionReason != "damage rate exceeded" {
		t.Errorf("wrong rejection reason: %s", result.RejectionReason)
	}
}

func TestPvPValidatorValidateActionCooldown(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	pv.SetZonePvPEnabled("pvp_zone", true)

	// Set up target entity
	targetPos := Position3D{X: 10, Y: 1, Z: 0}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, targetPos, 0, hitMin, hitMax)

	action := &PvPCombatAction{
		AttackerID:  1,
		TargetID:    2,
		ActionType:  PvPMagicAttack,
		DamageClaim: 30.0,
		ClientTime:  time.Now(),
		Position:    Position3D{X: 0, Y: 1, Z: 0},
		Direction:   Position3D{X: 1, Y: 0, Z: 0},
		AbilityID:   "fireball",
	}

	// First use should succeed
	result1 := pv.ValidateAction(action, 100*time.Millisecond, "pvp_zone")
	if !result1.Valid {
		t.Errorf("first use should be valid: %s", result1.RejectionReason)
	}

	// Immediate second use should fail (cooldown)
	result2 := pv.ValidateAction(action, 100*time.Millisecond, "pvp_zone")
	if result2.Valid {
		t.Error("immediate second use should be on cooldown")
	}
	if result2.RejectionReason != "ability on cooldown" {
		t.Errorf("wrong rejection reason: %s", result2.RejectionReason)
	}
}

func TestPvPValidatorValidateActionBasicAttackNoCooldown(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	pv.SetZonePvPEnabled("pvp_zone", true)

	// Set up target entity
	targetPos := Position3D{X: 10, Y: 1, Z: 0}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, targetPos, 0, hitMin, hitMax)

	action := &PvPCombatAction{
		AttackerID:  1,
		TargetID:    2,
		ActionType:  PvPMeleeAttack,
		DamageClaim: 30.0,
		ClientTime:  time.Now(),
		Position:    Position3D{X: 0, Y: 1, Z: 0},
		Direction:   Position3D{X: 1, Y: 0, Z: 0},
		AbilityID:   "", // Empty = basic attack
	}

	// Both should succeed (no cooldown for basic attacks)
	result1 := pv.ValidateAction(action, 100*time.Millisecond, "pvp_zone")
	if !result1.Valid {
		t.Errorf("first basic attack should be valid: %s", result1.RejectionReason)
	}

	result2 := pv.ValidateAction(action, 100*time.Millisecond, "pvp_zone")
	if !result2.Valid {
		t.Errorf("second basic attack should be valid (no cooldown): %s", result2.RejectionReason)
	}
}

func TestPvPValidatorGetCooldownRemaining(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	// No cooldown initially
	remaining := pv.GetCooldownRemaining(1, "fireball")
	if remaining != 0 {
		t.Errorf("should have no cooldown initially, got %v", remaining)
	}

	// Record a cooldown manually
	pv.mu.Lock()
	pv.cooldowns[1] = make(map[string]time.Time)
	pv.cooldowns[1]["fireball"] = time.Now()
	pv.mu.Unlock()

	// Should have some cooldown remaining
	remaining = pv.GetCooldownRemaining(1, "fireball")
	if remaining <= 0 {
		t.Error("should have cooldown remaining after use")
	}
	if remaining > time.Second {
		t.Errorf("cooldown should be at most 1 second, got %v", remaining)
	}
}

func TestPvPValidatorCleanupCooldowns(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	// Add old cooldown
	pv.mu.Lock()
	pv.cooldowns[1] = make(map[string]time.Time)
	pv.cooldowns[1]["old_ability"] = time.Now().Add(-2 * time.Minute)
	pv.cooldowns[1]["recent_ability"] = time.Now()
	pv.mu.Unlock()

	// Cleanup entries older than 1 minute
	pv.CleanupCooldowns(1 * time.Minute)

	pv.mu.RLock()
	_, hasOld := pv.cooldowns[1]["old_ability"]
	_, hasRecent := pv.cooldowns[1]["recent_ability"]
	pv.mu.RUnlock()

	if hasOld {
		t.Error("old cooldown should have been cleaned up")
	}
	if !hasRecent {
		t.Error("recent cooldown should not have been cleaned up")
	}
}

func TestPvPValidatorCleanupEmptyEntity(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	// Add only old cooldowns for an entity
	pv.mu.Lock()
	pv.cooldowns[1] = make(map[string]time.Time)
	pv.cooldowns[1]["old_ability"] = time.Now().Add(-2 * time.Minute)
	pv.mu.Unlock()

	// Cleanup should remove the entity entirely
	pv.CleanupCooldowns(1 * time.Minute)

	pv.mu.RLock()
	_, hasEntity := pv.cooldowns[1]
	pv.mu.RUnlock()

	if hasEntity {
		t.Error("empty entity cooldown map should be removed")
	}
}

func TestPvPValidatorHitNotConfirmed(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	pv.SetZonePvPEnabled("pvp_zone", true)

	// Set up target entity at position far away
	targetPos := Position3D{X: 100, Y: 1, Z: 100}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, targetPos, 0, hitMin, hitMax)

	action := &PvPCombatAction{
		AttackerID:  1,
		TargetID:    2,
		ActionType:  PvPMeleeAttack,
		DamageClaim: 30.0,
		ClientTime:  time.Now(),
		Position:    Position3D{X: 0, Y: 1, Z: 0},
		Direction:   Position3D{X: 0, Y: 0, Z: -1}, // Wrong direction
	}

	result := pv.ValidateAction(action, 100*time.Millisecond, "pvp_zone")

	if result.Valid {
		t.Error("action should be invalid when hit not confirmed")
	}
	if result.RejectionReason != "hit not confirmed" {
		t.Errorf("wrong rejection reason: %s", result.RejectionReason)
	}
}

func TestPvPValidatorSuccessfulHit(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	pv.SetZonePvPEnabled("pvp_zone", true)

	// Set up target entity
	targetPos := Position3D{X: 10, Y: 1, Z: 0}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, targetPos, 0, hitMin, hitMax)

	action := &PvPCombatAction{
		AttackerID:  1,
		TargetID:    2,
		ActionType:  PvPMeleeAttack,
		DamageClaim: 30.0,
		ClientTime:  time.Now(),
		Position:    Position3D{X: 0, Y: 1, Z: 0},
		Direction:   Position3D{X: 1, Y: 0, Z: 0}, // Toward target
	}

	result := pv.ValidateAction(action, 100*time.Millisecond, "pvp_zone")

	if !result.Valid {
		t.Errorf("action should be valid: %s", result.RejectionReason)
	}
	if result.ActualDamage != 30.0 {
		t.Errorf("actual damage should be 30.0, got %f", result.ActualDamage)
	}
}

func TestPvPCombatTypes(t *testing.T) {
	// Verify all combat types exist and have different values
	types := []PvPCombatType{
		PvPMeleeAttack,
		PvPRangedAttack,
		PvPMagicAttack,
		PvPAreaEffect,
		PvPStatusEffect,
	}

	seen := make(map[PvPCombatType]bool)
	for _, ct := range types {
		if seen[ct] {
			t.Errorf("duplicate combat type value: %d", ct)
		}
		seen[ct] = true
	}

	if len(seen) != 5 {
		t.Errorf("expected 5 unique combat types, got %d", len(seen))
	}
}

// ============================================================================
// Additional Lag Compensator Edge Case Tests
// ============================================================================

func TestStateHistoryGetAtTimeNoMatching(t *testing.T) {
	sh := NewStateHistory()

	// Record entity 1
	sh.Record(EntitySnapshot{
		EntityID:  1,
		Timestamp: time.Now(),
		Position:  Position3D{X: 10, Y: 0, Z: 0},
	})

	// Try to get entity 2 (doesn't exist)
	result := sh.GetAtTime(2, time.Now())
	if result != nil {
		t.Error("should return nil for non-existent entity")
	}
}

func TestStateHistoryGetAtTimeOutsideWindow(t *testing.T) {
	sh := NewStateHistory()

	// Record entity with very old timestamp
	oldTime := time.Now().Add(-1 * time.Second)
	sh.Record(EntitySnapshot{
		EntityID:  1,
		Timestamp: oldTime,
		Position:  Position3D{X: 10, Y: 0, Z: 0},
	})

	// Try to get state at a time way after the snapshot (beyond 500ms window)
	result := sh.GetAtTime(1, time.Now())
	if result != nil {
		t.Error("should return nil when snapshot is outside max rewind window")
	}
}

func TestHitTestTargetNotTracked(t *testing.T) {
	lc := NewLagCompensator()

	// No entities recorded
	origin := Position3D{X: 0, Y: 1, Z: 0}
	direction := Position3D{X: 1, Y: 0, Z: 0}

	result := lc.HitTest(1, 2, origin, direction, time.Now(), 100*time.Millisecond)

	if result.Hit {
		t.Error("should not hit when target is not tracked")
	}
}

func TestRayAABBIntersectParallelRay(t *testing.T) {
	// Ray parallel to X axis, passing through box
	origin := Position3D{X: 0, Y: 0.5, Z: 0.5}
	direction := Position3D{X: 1, Y: 0, Z: 0}
	boxMin := Position3D{X: -1, Y: -1, Z: -1}
	boxMax := Position3D{X: 1, Y: 1, Z: 1}
	boxCenter := Position3D{X: 5, Y: 0, Z: 0}

	hit, dist := rayAABBIntersect(origin, direction, boxMin, boxMax, boxCenter)

	if !hit {
		t.Error("ray should hit box")
	}
	if dist <= 0 {
		t.Errorf("distance should be positive, got %f", dist)
	}
}

func TestRayAABBIntersectParallelMiss(t *testing.T) {
	// Ray parallel to X axis but above the box
	origin := Position3D{X: 0, Y: 10, Z: 0}
	direction := Position3D{X: 1, Y: 0, Z: 0}
	boxMin := Position3D{X: -1, Y: -1, Z: -1}
	boxMax := Position3D{X: 1, Y: 1, Z: 1}
	boxCenter := Position3D{X: 5, Y: 0, Z: 0}

	hit, _ := rayAABBIntersect(origin, direction, boxMin, boxMax, boxCenter)

	if hit {
		t.Error("ray parallel but outside should miss")
	}
}

func TestTranslateBox(t *testing.T) {
	boxMin := Position3D{X: -1, Y: -2, Z: -3}
	boxMax := Position3D{X: 1, Y: 2, Z: 3}
	center := Position3D{X: 10, Y: 20, Z: 30}

	worldMin, worldMax := translateBox(boxMin, boxMax, center)

	expectedMin := Position3D{X: 9, Y: 18, Z: 27}
	expectedMax := Position3D{X: 11, Y: 22, Z: 33}

	if worldMin != expectedMin {
		t.Errorf("wrong world min: got %+v, expected %+v", worldMin, expectedMin)
	}
	if worldMax != expectedMax {
		t.Errorf("wrong world max: got %+v, expected %+v", worldMax, expectedMax)
	}
}

func TestSlabIntersect(t *testing.T) {
	// Test non-parallel case
	tmin, tmax, ok := slabIntersect(0, 1, 5, 10, 0, 100)
	if !ok {
		t.Error("should intersect")
	}
	if tmin != 5 || tmax != 10 {
		t.Errorf("wrong t values: tmin=%f, tmax=%f", tmin, tmax)
	}

	// Test parallel inside case
	tmin2, tmax2, ok2 := slabIntersect(7, 0, 5, 10, 0, 100)
	if !ok2 {
		t.Error("parallel inside should intersect")
	}
	if tmin2 != 0 || tmax2 != 100 {
		t.Errorf("parallel inside should not modify t: tmin=%f, tmax=%f", tmin2, tmax2)
	}

	// Test parallel outside case
	_, _, ok3 := slabIntersect(15, 0, 5, 10, 0, 100)
	if ok3 {
		t.Error("parallel outside should not intersect")
	}
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestLagCompensatorConcurrent(t *testing.T) {
	lc := NewLagCompensator()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				entityID := uint64(id*100 + j)
				pos := Position3D{X: float32(j), Y: 0, Z: 0}
				hitMin := Position3D{X: -1, Y: -1, Z: -1}
				hitMax := Position3D{X: 1, Y: 1, Z: 1}
				lc.RecordEntityState(entityID, pos, 0, hitMin, hitMax)
				_ = lc.EntityCount()
				lc.RemoveEntity(entityID)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestPvPValidatorConcurrent(t *testing.T) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			zone := "zone_" + string(rune('A'+id))
			for j := 0; j < 100; j++ {
				pv.SetZonePvPEnabled(zone, j%2 == 0)
				_ = pv.IsZonePvPEnabled(zone)
				_ = pv.GetCooldownRemaining(uint64(id), "ability")
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// ============================================================================
// Math Helper Tests
// ============================================================================

func TestCosHelper(t *testing.T) {
	// Test cos at known values
	// Note: This uses a Taylor series approximation, so larger epsilon needed
	tests := []struct {
		input    float64
		expected float64
		epsilon  float64
	}{
		{0, 1.0, 0.01},
		{1.5707963268, 0.0, 0.1},   // pi/2
		{3.14159265359, -1.0, 0.3}, // pi (larger epsilon due to Taylor approximation)
	}

	for _, tc := range tests {
		result := cos(tc.input)
		if diff := result - tc.expected; diff < -tc.epsilon || diff > tc.epsilon {
			t.Errorf("cos(%f) = %f, expected %f (within %f)", tc.input, result, tc.expected, tc.epsilon)
		}
	}
}

func TestSinHelper(t *testing.T) {
	// Test sin at known values
	tests := []struct {
		input    float64
		expected float64
		epsilon  float64
	}{
		{0, 0.0, 0.01},
		{1.5707963268, 1.0, 0.01},
		{3.14159265359, 0.0, 0.01},
	}

	for _, tc := range tests {
		result := sin(tc.input)
		if diff := result - tc.expected; diff < -tc.epsilon || diff > tc.epsilon {
			t.Errorf("sin(%f) = %f, expected %f", tc.input, result, tc.expected)
		}
	}
}

func TestModHelper(t *testing.T) {
	tests := []struct {
		a, b, expected float64
	}{
		{5.0, 3.0, 2.0},
		{7.0, 3.0, 1.0},
		{-1.0, 3.0, 2.0},
		{6.0, 3.0, 0.0},
	}

	for _, tc := range tests {
		result := mod(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("mod(%f, %f) = %f, expected %f", tc.a, tc.b, result, tc.expected)
		}
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkPvPValidateAction(b *testing.B) {
	lc := NewLagCompensator()
	pv := NewPvPValidator(lc)
	pv.SetZonePvPEnabled("pvp_zone", true)

	targetPos := Position3D{X: 10, Y: 1, Z: 0}
	hitMin := Position3D{X: -1, Y: -1, Z: -1}
	hitMax := Position3D{X: 1, Y: 1, Z: 1}
	lc.RecordEntityState(2, targetPos, 0, hitMin, hitMax)

	action := &PvPCombatAction{
		AttackerID:  1,
		TargetID:    2,
		ActionType:  PvPMeleeAttack,
		DamageClaim: 30.0,
		Position:    Position3D{X: 0, Y: 1, Z: 0},
		Direction:   Position3D{X: 1, Y: 0, Z: 0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		action.ClientTime = time.Now()
		pv.ValidateAction(action, 100*time.Millisecond, "pvp_zone")
	}
}
