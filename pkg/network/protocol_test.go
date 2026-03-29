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
