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
