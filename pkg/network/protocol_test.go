package network

import (
	"bytes"
	"testing"
)

func TestPlayerInputEncodeDecode(t *testing.T) {
	original := &PlayerInput{
		MoveForward:  1.0,
		MoveRight:    -0.5,
		Turn:         0.25,
		Jump:         true,
		Attack:       false,
		Use:          true,
		SequenceNum:  42,
		ClientTimeMs: 12345,
	}

	var buf bytes.Buffer
	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Skip the type byte for decoding
	data := buf.Bytes()
	reader := bytes.NewReader(data[1:]) // Skip type byte
	decoded, err := DecodePlayerInput(reader)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.MoveForward != original.MoveForward {
		t.Errorf("MoveForward: got %f, want %f", decoded.MoveForward, original.MoveForward)
	}
	if decoded.MoveRight != original.MoveRight {
		t.Errorf("MoveRight: got %f, want %f", decoded.MoveRight, original.MoveRight)
	}
	if decoded.Turn != original.Turn {
		t.Errorf("Turn: got %f, want %f", decoded.Turn, original.Turn)
	}
	if decoded.Jump != original.Jump {
		t.Errorf("Jump: got %v, want %v", decoded.Jump, original.Jump)
	}
	if decoded.Attack != original.Attack {
		t.Errorf("Attack: got %v, want %v", decoded.Attack, original.Attack)
	}
	if decoded.Use != original.Use {
		t.Errorf("Use: got %v, want %v", decoded.Use, original.Use)
	}
	if decoded.SequenceNum != original.SequenceNum {
		t.Errorf("SequenceNum: got %d, want %d", decoded.SequenceNum, original.SequenceNum)
	}
	if decoded.ClientTimeMs != original.ClientTimeMs {
		t.Errorf("ClientTimeMs: got %d, want %d", decoded.ClientTimeMs, original.ClientTimeMs)
	}
}

func TestWorldStateEncodeDecode(t *testing.T) {
	original := &WorldState{
		ServerTimeMs: 54321,
		AckSequence:  10,
		Entities: []EntityState{
			{EntityID: 1, X: 10.5, Y: 20.5, Z: 0.5, Angle: 1.57, Health: 100},
			{EntityID: 2, X: -5.0, Y: 15.0, Z: 1.0, Angle: 3.14, Health: 50},
		},
	}

	var buf bytes.Buffer
	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	data := buf.Bytes()
	reader := bytes.NewReader(data[1:]) // Skip type byte
	decoded, err := DecodeWorldState(reader)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.ServerTimeMs != original.ServerTimeMs {
		t.Errorf("ServerTimeMs: got %d, want %d", decoded.ServerTimeMs, original.ServerTimeMs)
	}
	if decoded.AckSequence != original.AckSequence {
		t.Errorf("AckSequence: got %d, want %d", decoded.AckSequence, original.AckSequence)
	}
	if len(decoded.Entities) != len(original.Entities) {
		t.Fatalf("Entities count: got %d, want %d", len(decoded.Entities), len(original.Entities))
	}
	for i, e := range decoded.Entities {
		o := original.Entities[i]
		if e.EntityID != o.EntityID {
			t.Errorf("Entity[%d].EntityID: got %d, want %d", i, e.EntityID, o.EntityID)
		}
		if e.X != o.X {
			t.Errorf("Entity[%d].X: got %f, want %f", i, e.X, o.X)
		}
		if e.Y != o.Y {
			t.Errorf("Entity[%d].Y: got %f, want %f", i, e.Y, o.Y)
		}
		if e.Health != o.Health {
			t.Errorf("Entity[%d].Health: got %f, want %f", i, e.Health, o.Health)
		}
	}
}

func TestPingPongEncodeDecode(t *testing.T) {
	ping := &Ping{ClientTimeMs: 99999}
	var pingBuf bytes.Buffer
	if err := ping.Encode(&pingBuf); err != nil {
		t.Fatalf("Ping Encode failed: %v", err)
	}

	pingData := pingBuf.Bytes()
	pingReader := bytes.NewReader(pingData[1:])
	decodedPing, err := DecodePing(pingReader)
	if err != nil {
		t.Fatalf("Ping Decode failed: %v", err)
	}
	if decodedPing.ClientTimeMs != ping.ClientTimeMs {
		t.Errorf("Ping ClientTimeMs: got %d, want %d", decodedPing.ClientTimeMs, ping.ClientTimeMs)
	}

	pong := &Pong{ClientTimeMs: 99999, ServerTimeMs: 100100}
	var pongBuf bytes.Buffer
	if err := pong.Encode(&pongBuf); err != nil {
		t.Fatalf("Pong Encode failed: %v", err)
	}

	pongData := pongBuf.Bytes()
	pongReader := bytes.NewReader(pongData[1:])
	decodedPong, err := DecodePong(pongReader)
	if err != nil {
		t.Fatalf("Pong Decode failed: %v", err)
	}
	if decodedPong.ClientTimeMs != pong.ClientTimeMs {
		t.Errorf("Pong ClientTimeMs: got %d, want %d", decodedPong.ClientTimeMs, pong.ClientTimeMs)
	}
	if decodedPong.ServerTimeMs != pong.ServerTimeMs {
		t.Errorf("Pong ServerTimeMs: got %d, want %d", decodedPong.ServerTimeMs, pong.ServerTimeMs)
	}
}

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		msg      Message
		expected uint8
	}{
		{&PlayerInput{}, MsgTypePlayerInput},
		{&WorldState{}, MsgTypeWorldState},
		{&Ping{}, MsgTypePing},
		{&Pong{}, MsgTypePong},
	}

	for _, tc := range tests {
		if tc.msg.Type() != tc.expected {
			t.Errorf("Type() = %d, want %d", tc.msg.Type(), tc.expected)
		}
	}
}

func TestReadMessageType(t *testing.T) {
	input := &PlayerInput{MoveForward: 1.0}
	var buf bytes.Buffer
	if err := input.Encode(&buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	reader := bytes.NewReader(buf.Bytes())
	msgType, err := ReadMessageType(reader)
	if err != nil {
		t.Fatalf("ReadMessageType failed: %v", err)
	}
	if msgType != MsgTypePlayerInput {
		t.Errorf("ReadMessageType = %d, want %d", msgType, MsgTypePlayerInput)
	}
}

func TestEmptyWorldState(t *testing.T) {
	original := &WorldState{
		ServerTimeMs: 1000,
		AckSequence:  0,
		Entities:     []EntityState{},
	}

	var buf bytes.Buffer
	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	data := buf.Bytes()
	reader := bytes.NewReader(data[1:])
	decoded, err := DecodeWorldState(reader)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded.Entities) != 0 {
		t.Errorf("Expected empty entities, got %d", len(decoded.Entities))
	}
}

func BenchmarkPlayerInputEncode(b *testing.B) {
	input := &PlayerInput{
		MoveForward:  1.0,
		MoveRight:    -0.5,
		Turn:         0.25,
		Jump:         true,
		SequenceNum:  42,
		ClientTimeMs: 12345,
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = input.Encode(&buf)
	}
}

func BenchmarkWorldStateEncode(b *testing.B) {
	state := &WorldState{
		ServerTimeMs: 54321,
		AckSequence:  10,
		Entities:     make([]EntityState, 100),
	}
	for i := range state.Entities {
		state.Entities[i] = EntityState{
			EntityID: uint64(i),
			X:        float32(i) * 10,
			Y:        float32(i) * 10,
			Health:   100,
		}
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = state.Encode(&buf)
	}
}

// ============================================================================
// Delta Compression Tests
// ============================================================================

func TestDeltaEntityUpdateEncodeDecode(t *testing.T) {
	original := &DeltaEntityUpdate{
		ServerTimeMs: 1234567890,
		EntityID:     42,
		FieldMask:    FieldPosition | FieldAngle | FieldState,
		X:            10.5,
		Y:            20.5,
		Z:            30.5,
		Angle:        45.0,
		State:        7,
	}

	var buf bytes.Buffer
	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Skip the type byte for decoding
	data := buf.Bytes()
	reader := bytes.NewReader(data[1:])
	decoded, err := DecodeDeltaEntityUpdate(reader)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.ServerTimeMs != original.ServerTimeMs {
		t.Errorf("ServerTimeMs: got %d, want %d", decoded.ServerTimeMs, original.ServerTimeMs)
	}
	if decoded.EntityID != original.EntityID {
		t.Errorf("EntityID: got %d, want %d", decoded.EntityID, original.EntityID)
	}
	if decoded.FieldMask != original.FieldMask {
		t.Errorf("FieldMask: got %d, want %d", decoded.FieldMask, original.FieldMask)
	}
	if decoded.X != original.X {
		t.Errorf("X: got %f, want %f", decoded.X, original.X)
	}
	if decoded.Y != original.Y {
		t.Errorf("Y: got %f, want %f", decoded.Y, original.Y)
	}
	if decoded.Z != original.Z {
		t.Errorf("Z: got %f, want %f", decoded.Z, original.Z)
	}
	if decoded.Angle != original.Angle {
		t.Errorf("Angle: got %f, want %f", decoded.Angle, original.Angle)
	}
	if decoded.State != original.State {
		t.Errorf("State: got %d, want %d", decoded.State, original.State)
	}
}

func TestDeltaEntityUpdateMinimalEncoding(t *testing.T) {
	// Only position changed
	original := &DeltaEntityUpdate{
		ServerTimeMs: 12345,
		EntityID:     1,
		FieldMask:    FieldPosition,
		X:            1.0,
		Y:            2.0,
		Z:            3.0,
	}

	var buf bytes.Buffer
	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Calculate expected size:
	// 1 (type) + 4 (timestamp) + 1 (varint entityID=1) + 1 (fieldmask) + 12 (3 floats) = 19 bytes
	expectedSize := 19
	if buf.Len() != expectedSize {
		t.Errorf("Minimal delta size: got %d, want %d", buf.Len(), expectedSize)
	}
}

func TestComputeDelta(t *testing.T) {
	prev := &EntityUpdate{
		ServerTimeMs: 1000,
		EntityID:     42,
		X:            0, Y: 0, Z: 0,
		Angle:    0,
		Health:   100,
		Velocity: 0,
		State:    0,
	}

	curr := &EntityUpdate{
		ServerTimeMs: 2000,
		EntityID:     42,
		X:            10, Y: 0, Z: 5, // Position changed
		Angle:    45, // Angle changed
		Health:   100,
		Velocity: 0,
		State:    0,
	}

	delta := ComputeDelta(prev, curr)

	if delta.FieldMask&FieldPosition == 0 {
		t.Error("Position should be marked as changed")
	}
	if delta.FieldMask&FieldAngle == 0 {
		t.Error("Angle should be marked as changed")
	}
	if delta.FieldMask&FieldHealth != 0 {
		t.Error("Health should NOT be marked as changed")
	}
	if delta.FieldMask&FieldVelocity != 0 {
		t.Error("Velocity should NOT be marked as changed")
	}
	if delta.FieldMask&FieldState != 0 {
		t.Error("State should NOT be marked as changed")
	}
}

func TestApplyDelta(t *testing.T) {
	prev := &EntityUpdate{
		ServerTimeMs: 1000,
		EntityID:     42,
		X:            0, Y: 0, Z: 0,
		Angle:    0,
		Health:   100,
		Velocity: 5.0,
		State:    1,
	}

	delta := &DeltaEntityUpdate{
		ServerTimeMs: 2000,
		EntityID:     42,
		FieldMask:    FieldPosition | FieldHealth,
		X:            10, Y: 20, Z: 30,
		Health: 75,
	}

	result := ApplyDelta(prev, delta)

	// Changed fields should be updated
	if result.X != 10 || result.Y != 20 || result.Z != 30 {
		t.Errorf("Position not updated correctly: got (%f, %f, %f)", result.X, result.Y, result.Z)
	}
	if result.Health != 75 {
		t.Errorf("Health not updated: got %f, want 75", result.Health)
	}

	// Unchanged fields should preserve previous values
	if result.Angle != 0 {
		t.Errorf("Angle should be unchanged: got %f, want 0", result.Angle)
	}
	if result.Velocity != 5.0 {
		t.Errorf("Velocity should be unchanged: got %f, want 5.0", result.Velocity)
	}
	if result.State != 1 {
		t.Errorf("State should be unchanged: got %d, want 1", result.State)
	}
}

func TestDeltaCompressionSavings(t *testing.T) {
	// Full entity update
	full := &EntityUpdate{
		ServerTimeMs: 1234567890,
		EntityID:     12345,
		X:            100.5, Y: 50.25, Z: 200.75,
		Angle:    90.0,
		Health:   100.0,
		Velocity: 5.5,
		State:    3,
	}

	var fullBuf bytes.Buffer
	if err := full.Encode(&fullBuf); err != nil {
		t.Fatalf("Full encode failed: %v", err)
	}
	fullSize := fullBuf.Len()

	// Delta with only position changed
	delta := &DeltaEntityUpdate{
		ServerTimeMs: 1234567891,
		EntityID:     12345,
		FieldMask:    FieldPosition,
		X:            101.5, Y: 50.25, Z: 201.75,
	}

	var deltaBuf bytes.Buffer
	if err := delta.Encode(&deltaBuf); err != nil {
		t.Fatalf("Delta encode failed: %v", err)
	}
	deltaSize := deltaBuf.Len()

	// Delta should be smaller than full update
	if deltaSize >= fullSize {
		t.Errorf("Delta (%d bytes) should be smaller than full update (%d bytes)",
			deltaSize, fullSize)
	}

	// Calculate savings percentage
	savings := float64(fullSize-deltaSize) / float64(fullSize) * 100
	t.Logf("Delta compression savings: %.1f%% (full: %d bytes, delta: %d bytes)",
		savings, fullSize, deltaSize)
}

func TestVarUint64Encoding(t *testing.T) {
	tests := []uint64{
		0,
		1,
		127,
		128,
		255,
		256,
		16383,
		16384,
		2097151,
		2097152,
		268435455,
		268435456,
		0xFFFFFFFFFFFFFFFF,
	}

	for _, v := range tests {
		var buf bytes.Buffer
		if err := encodeVarUint64(&buf, v); err != nil {
			t.Errorf("encodeVarUint64(%d) failed: %v", v, err)
			continue
		}

		decoded, err := decodeVarUint64(&buf)
		if err != nil {
			t.Errorf("decodeVarUint64 failed for %d: %v", v, err)
			continue
		}

		if decoded != v {
			t.Errorf("VarUint64: got %d, want %d", decoded, v)
		}
	}
}

func BenchmarkFullEntityUpdate(b *testing.B) {
	full := &EntityUpdate{
		ServerTimeMs: 1234567890,
		EntityID:     12345,
		X:            100.5, Y: 50.25, Z: 200.75,
		Angle:    90.0,
		Health:   100.0,
		Velocity: 5.5,
		State:    3,
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = full.Encode(&buf)
	}
}

func BenchmarkDeltaEntityUpdate(b *testing.B) {
	delta := &DeltaEntityUpdate{
		ServerTimeMs: 1234567890,
		EntityID:     12345,
		FieldMask:    FieldPosition,
		X:            100.5, Y: 50.25, Z: 200.75,
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = delta.Encode(&buf)
	}
}
