package federation

import (
	"encoding/gob"
	"io"
	"time"
)

// PriceSignal broadcasts economy price updates across federation.
type PriceSignal struct {
	ServerID   string
	CityID     string
	PriceTable map[string]float64
	Timestamp  time.Time
}

// GlobalEvent represents a world event broadcast to all servers.
type GlobalEvent struct {
	EventID      string
	EventType    string
	Description  string
	AffectedArea struct {
		CenterX, CenterZ float64
		Radius           float64
	}
	StartTime time.Time
	Duration  time.Duration
}

// ProcessPriceSignal updates remote price information.
func (f *Federation) ProcessPriceSignal(signal *PriceSignal) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := signal.ServerID + ":" + signal.CityID
	existing := f.remotePrices[key]

	// Only accept newer signals
	if existing == nil || signal.Timestamp.After(existing.Timestamp) {
		f.remotePrices[key] = signal
	}
}

// GetRemotePrices returns price signals from other servers.
func (f *Federation) GetRemotePrices() []*PriceSignal {
	f.mu.RLock()
	defer f.mu.RUnlock()

	signals := make([]*PriceSignal, 0, len(f.remotePrices))
	for _, s := range f.remotePrices {
		signals = append(signals, s)
	}
	return signals
}

// BroadcastEvent registers a global event.
func (f *Federation) BroadcastEvent(event *GlobalEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activeEvents[event.EventID] = event
}

// GetActiveEvents returns currently active global events.
func (f *Federation) GetActiveEvents() []*GlobalEvent {
	f.mu.RLock()
	defer f.mu.RUnlock()

	now := time.Now()
	events := make([]*GlobalEvent, 0)
	for _, e := range f.activeEvents {
		if now.Sub(e.StartTime) < e.Duration {
			events = append(events, e)
		}
	}
	return events
}

// EncodePriceSignal writes a price signal to a writer.
func EncodePriceSignal(w io.Writer, signal *PriceSignal) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(signal)
}

// DecodePriceSignal reads a price signal from a reader.
func DecodePriceSignal(r io.Reader) (*PriceSignal, error) {
	dec := gob.NewDecoder(r)
	signal := &PriceSignal{}
	if err := dec.Decode(signal); err != nil {
		return nil, err
	}
	return signal, nil
}

// EncodeGlobalEvent writes a global event to a writer.
func EncodeGlobalEvent(w io.Writer, event *GlobalEvent) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(event)
}

// DecodeGlobalEvent reads a global event from a reader.
func DecodeGlobalEvent(r io.Reader) (*GlobalEvent, error) {
	dec := gob.NewDecoder(r)
	event := &GlobalEvent{}
	if err := dec.Decode(event); err != nil {
		return nil, err
	}
	return event, nil
}
