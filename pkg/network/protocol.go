// Package network provides client-server networking.
package network

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Message type identifiers.
const (
	MsgTypePlayerInput  uint8 = 1
	MsgTypeWorldState   uint8 = 2
	MsgTypeEntityUpdate uint8 = 3
	MsgTypeChunkData    uint8 = 4
	MsgTypePing         uint8 = 5
	MsgTypePong         uint8 = 6
)

// Message is the interface all network messages implement.
type Message interface {
	Type() uint8
	Encode(w io.Writer) error
}

// PlayerInput represents client input commands.
type PlayerInput struct {
	MoveForward  float32
	MoveRight    float32
	Turn         float32
	Jump         bool
	Attack       bool
	Use          bool
	SequenceNum  uint32 // For lag compensation
	ClientTimeMs uint32 // Client timestamp
}

// Type returns the message type identifier.
func (m *PlayerInput) Type() uint8 { return MsgTypePlayerInput }

// Encode writes the message to a writer.
func (m *PlayerInput) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, m.Type()); err != nil {
		return fmt.Errorf("encode type: %w", err)
	}
	if err := m.encodeMovement(w); err != nil {
		return err
	}
	if err := m.encodeFlags(w); err != nil {
		return err
	}
	return m.encodeTimestamps(w)
}

// encodeMovement writes movement fields to the writer.
func (m *PlayerInput) encodeMovement(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, m.MoveForward); err != nil {
		return fmt.Errorf("encode MoveForward: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, m.MoveRight); err != nil {
		return fmt.Errorf("encode MoveRight: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, m.Turn); err != nil {
		return fmt.Errorf("encode Turn: %w", err)
	}
	return nil
}

// encodeFlags packs boolean flags into a byte and writes it.
func (m *PlayerInput) encodeFlags(w io.Writer) error {
	var flags uint8
	if m.Jump {
		flags |= 1
	}
	if m.Attack {
		flags |= 2
	}
	if m.Use {
		flags |= 4
	}
	if err := binary.Write(w, binary.LittleEndian, flags); err != nil {
		return fmt.Errorf("encode flags: %w", err)
	}
	return nil
}

// encodeTimestamps writes sequence number and client timestamp.
func (m *PlayerInput) encodeTimestamps(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, m.SequenceNum); err != nil {
		return fmt.Errorf("encode SequenceNum: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, m.ClientTimeMs); err != nil {
		return fmt.Errorf("encode ClientTimeMs: %w", err)
	}
	return nil
}

// DecodePlayerInput reads a PlayerInput from a reader.
func DecodePlayerInput(r io.Reader) (*PlayerInput, error) {
	m := &PlayerInput{}
	if err := m.decodeMovement(r); err != nil {
		return nil, err
	}
	if err := m.decodeFlags(r); err != nil {
		return nil, err
	}
	if err := m.decodeTimestamps(r); err != nil {
		return nil, err
	}
	return m, nil
}

// decodeMovement reads movement fields from the reader.
func (m *PlayerInput) decodeMovement(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &m.MoveForward); err != nil {
		return fmt.Errorf("decode MoveForward: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &m.MoveRight); err != nil {
		return fmt.Errorf("decode MoveRight: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &m.Turn); err != nil {
		return fmt.Errorf("decode Turn: %w", err)
	}
	return nil
}

// decodeFlags reads and unpacks boolean flags from a byte.
func (m *PlayerInput) decodeFlags(r io.Reader) error {
	var flags uint8
	if err := binary.Read(r, binary.LittleEndian, &flags); err != nil {
		return fmt.Errorf("decode flags: %w", err)
	}
	m.Jump = flags&1 != 0
	m.Attack = flags&2 != 0
	m.Use = flags&4 != 0
	return nil
}

// decodeTimestamps reads sequence number and client timestamp.
func (m *PlayerInput) decodeTimestamps(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &m.SequenceNum); err != nil {
		return fmt.Errorf("decode SequenceNum: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &m.ClientTimeMs); err != nil {
		return fmt.Errorf("decode ClientTimeMs: %w", err)
	}
	return nil
}

// EntityState represents a single entity's state in a world update.
type EntityState struct {
	EntityID uint64
	X, Y, Z  float32
	Angle    float32
	Health   float32
}

// WorldState represents the authoritative server state.
type WorldState struct {
	ServerTimeMs uint32
	AckSequence  uint32 // Last acknowledged client input
	Entities     []EntityState
}

// Type returns the message type identifier.
func (m *WorldState) Type() uint8 { return MsgTypeWorldState }

// Encode writes the message to a writer.
func (m *WorldState) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, m.Type()); err != nil {
		return fmt.Errorf("encode type: %w", err)
	}
	if err := m.encodeHeader(w); err != nil {
		return err
	}
	return m.encodeEntities(w)
}

// encodeHeader writes server time, ack sequence, and entity count.
func (m *WorldState) encodeHeader(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, m.ServerTimeMs); err != nil {
		return fmt.Errorf("encode ServerTimeMs: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, m.AckSequence); err != nil {
		return fmt.Errorf("encode AckSequence: %w", err)
	}
	entityCount := uint16(len(m.Entities))
	if err := binary.Write(w, binary.LittleEndian, entityCount); err != nil {
		return fmt.Errorf("encode entity count: %w", err)
	}
	return nil
}

// encodeEntities writes all entity states to the writer.
func (m *WorldState) encodeEntities(w io.Writer) error {
	for _, e := range m.Entities {
		if err := encodeEntityState(w, &e); err != nil {
			return err
		}
	}
	return nil
}

// encodeEntityState writes a single entity state to the writer.
func encodeEntityState(w io.Writer, e *EntityState) error {
	if err := binary.Write(w, binary.LittleEndian, e.EntityID); err != nil {
		return fmt.Errorf("encode EntityID: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, e.X); err != nil {
		return fmt.Errorf("encode X: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, e.Y); err != nil {
		return fmt.Errorf("encode Y: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, e.Z); err != nil {
		return fmt.Errorf("encode Z: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, e.Angle); err != nil {
		return fmt.Errorf("encode Angle: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, e.Health); err != nil {
		return fmt.Errorf("encode Health: %w", err)
	}
	return nil
}

// DecodeWorldState reads a WorldState from a reader.
func DecodeWorldState(r io.Reader) (*WorldState, error) {
	m := &WorldState{}
	if err := m.decodeHeader(r); err != nil {
		return nil, err
	}
	if err := m.decodeEntities(r); err != nil {
		return nil, err
	}
	return m, nil
}

// decodeHeader reads server time, ack sequence, and allocates entity slice.
func (m *WorldState) decodeHeader(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &m.ServerTimeMs); err != nil {
		return fmt.Errorf("decode ServerTimeMs: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &m.AckSequence); err != nil {
		return fmt.Errorf("decode AckSequence: %w", err)
	}
	var entityCount uint16
	if err := binary.Read(r, binary.LittleEndian, &entityCount); err != nil {
		return fmt.Errorf("decode entity count: %w", err)
	}
	m.Entities = make([]EntityState, entityCount)
	return nil
}

// decodeEntities reads all entity states from the reader.
func (m *WorldState) decodeEntities(r io.Reader) error {
	for i := range m.Entities {
		if err := decodeEntityState(r, &m.Entities[i]); err != nil {
			return err
		}
	}
	return nil
}

// decodeEntityState reads a single entity state from the reader.
func decodeEntityState(r io.Reader, e *EntityState) error {
	if err := binary.Read(r, binary.LittleEndian, &e.EntityID); err != nil {
		return fmt.Errorf("decode EntityID: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &e.X); err != nil {
		return fmt.Errorf("decode X: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &e.Y); err != nil {
		return fmt.Errorf("decode Y: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &e.Z); err != nil {
		return fmt.Errorf("decode Z: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &e.Angle); err != nil {
		return fmt.Errorf("decode Angle: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &e.Health); err != nil {
		return fmt.Errorf("decode Health: %w", err)
	}
	return nil
}

// Ping is sent by client to measure latency.
type Ping struct {
	ClientTimeMs uint32
}

// Type returns the message type identifier.
func (m *Ping) Type() uint8 { return MsgTypePing }

// Encode writes the message to a writer.
func (m *Ping) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, m.Type()); err != nil {
		return fmt.Errorf("encode type: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, m.ClientTimeMs); err != nil {
		return fmt.Errorf("encode ClientTimeMs: %w", err)
	}
	return nil
}

// DecodePing reads a Ping from a reader.
func DecodePing(r io.Reader) (*Ping, error) {
	m := &Ping{}
	if err := binary.Read(r, binary.LittleEndian, &m.ClientTimeMs); err != nil {
		return nil, fmt.Errorf("decode ClientTimeMs: %w", err)
	}
	return m, nil
}

// Pong is sent by server in response to Ping.
type Pong struct {
	ClientTimeMs uint32
	ServerTimeMs uint32
}

// Type returns the message type identifier.
func (m *Pong) Type() uint8 { return MsgTypePong }

// Encode writes the message to a writer.
func (m *Pong) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, m.Type()); err != nil {
		return fmt.Errorf("encode type: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, m.ClientTimeMs); err != nil {
		return fmt.Errorf("encode ClientTimeMs: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, m.ServerTimeMs); err != nil {
		return fmt.Errorf("encode ServerTimeMs: %w", err)
	}
	return nil
}

// DecodePong reads a Pong from a reader.
func DecodePong(r io.Reader) (*Pong, error) {
	m := &Pong{}
	if err := binary.Read(r, binary.LittleEndian, &m.ClientTimeMs); err != nil {
		return nil, fmt.Errorf("decode ClientTimeMs: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &m.ServerTimeMs); err != nil {
		return nil, fmt.Errorf("decode ServerTimeMs: %w", err)
	}
	return m, nil
}

// ReadMessageType reads just the message type byte.
func ReadMessageType(r io.Reader) (uint8, error) {
	var msgType uint8
	if err := binary.Read(r, binary.LittleEndian, &msgType); err != nil {
		return 0, err
	}
	return msgType, nil
}
