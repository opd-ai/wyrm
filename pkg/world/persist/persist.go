// Package persist provides world state persistence for server restart recovery.
// Per ROADMAP Phase 5 item 20:
// AC: Server restart with same seed restores world state diff <5% from pre-restart snapshot.
package persist

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WorldSnapshot represents the complete serializable world state.
type WorldSnapshot struct {
	// Metadata
	Seed      int64
	Genre     string
	Timestamp time.Time
	Version   int

	// Entity data
	Entities []EntityData

	// Economy data
	EconomyNodes []EconomyNodeData

	// Quest flags (persistent per-player)
	QuestFlags map[uint64]map[string]bool

	// NPC schedules state
	WorldHour int
	WorldDay  int
}

// EntityData represents a serializable entity with its components.
type EntityData struct {
	ID uint64

	// Position component
	HasPosition      bool
	PosX, PosY, PosZ float64
	PosAngle         float64

	// Health component
	HasHealth     bool
	HealthCurrent float64
	HealthMax     float64

	// Faction component
	HasFaction        bool
	FactionID         string
	FactionReputation float64

	// Crime component
	HasCrime        bool
	WantedLevel     int
	BountyAmount    float64
	LastCrimeTime   float64
	InJail          bool
	JailReleaseTime float64

	// Inventory component
	HasInventory      bool
	InventoryItems    []string
	InventoryCapacity int

	// Skills component
	HasSkills       bool
	SkillLevels     map[string]int
	SkillExperience map[string]float64
	SchoolBonuses   map[string]float64

	// Quest component
	HasQuest       bool
	QuestID        string
	QuestStage     int
	QuestFlags     map[string]bool
	QuestCompleted bool
	LockedBranches map[string]bool
}

// EconomyNodeData represents a serializable economy node.
type EconomyNodeData struct {
	EntityID   uint64
	PriceTable map[string]float64
	Supply     map[string]int
	Demand     map[string]int
}

// Persister handles world state persistence operations.
type Persister struct {
	mu       sync.Mutex
	dataDir  string
	autoSave bool
	interval time.Duration
	stopCh   chan struct{}
}

// NewPersister creates a new world state persister.
func NewPersister(dataDir string) *Persister {
	return &Persister{
		dataDir:  dataDir,
		interval: 5 * time.Minute,
		stopCh:   make(chan struct{}),
	}
}

// SetAutoSave enables or disables automatic periodic saving.
func (p *Persister) SetAutoSave(enabled bool, interval time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.autoSave = enabled
	if interval > 0 {
		p.interval = interval
	}
}

// EnsureDataDir creates the data directory if it doesn't exist.
func (p *Persister) EnsureDataDir() error {
	return os.MkdirAll(p.dataDir, 0755)
}

// snapshotPath returns the path for a world snapshot file.
func (p *Persister) snapshotPath(seed int64) string {
	filename := fmt.Sprintf("world_%d.gob.gz", seed)
	return filepath.Join(p.dataDir, filename)
}

// Save writes a world snapshot to disk.
func (p *Persister) Save(snapshot *WorldSnapshot) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.EnsureDataDir(); err != nil {
		return fmt.Errorf("ensure data dir: %w", err)
	}

	snapshot.Timestamp = time.Now()
	snapshot.Version = 1

	path := p.snapshotPath(snapshot.Seed)
	tempPath := path + ".tmp"

	f, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if err := p.writeSnapshot(f, snapshot); err != nil {
		f.Close()
		os.Remove(tempPath)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("rename to final path: %w", err)
	}

	return nil
}

// writeSnapshot writes a compressed snapshot to a writer.
func (p *Persister) writeSnapshot(w io.Writer, snapshot *WorldSnapshot) error {
	gzw := gzip.NewWriter(w)
	enc := gob.NewEncoder(gzw)

	if err := enc.Encode(snapshot); err != nil {
		gzw.Close()
		return fmt.Errorf("encode snapshot: %w", err)
	}

	if err := gzw.Close(); err != nil {
		return fmt.Errorf("close gzip writer: %w", err)
	}

	return nil
}

// Load reads a world snapshot from disk.
func (p *Persister) Load(seed int64) (*WorldSnapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	path := p.snapshotPath(seed)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No saved state exists
		}
		return nil, fmt.Errorf("open snapshot file: %w", err)
	}
	defer f.Close()

	return p.readSnapshot(f)
}

// readSnapshot reads a compressed snapshot from a reader.
func (p *Persister) readSnapshot(r io.Reader) (*WorldSnapshot, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("create gzip reader: %w", err)
	}
	defer gzr.Close()

	dec := gob.NewDecoder(gzr)
	snapshot := &WorldSnapshot{}

	if err := dec.Decode(snapshot); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}

	return snapshot, nil
}

// Exists checks if a saved world exists for the given seed.
func (p *Persister) Exists(seed int64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	path := p.snapshotPath(seed)
	_, err := os.Stat(path)
	return err == nil
}

// Delete removes a saved world snapshot.
func (p *Persister) Delete(seed int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	path := p.snapshotPath(seed)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// LastSaveTime returns when the world was last saved.
func (p *Persister) LastSaveTime(seed int64) (time.Time, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	path := p.snapshotPath(seed)
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// CalculateStateDiff compares two snapshots and returns difference percentage.
// Used for AC verification: diff <5% from pre-restart snapshot.
func CalculateStateDiff(before, after *WorldSnapshot) float64 {
	if before == nil || after == nil {
		return 100.0
	}

	totalFields, diffFields := compareMetadata(before, after)
	entityTotal, entityDiff := compareEntities(before.Entities, after.Entities)
	totalFields += entityTotal
	diffFields += entityDiff

	if totalFields == 0 {
		return 0.0
	}
	return float64(diffFields) / float64(totalFields) * 100.0
}

// compareMetadata compares snapshot metadata fields.
func compareMetadata(before, after *WorldSnapshot) (total, diff int) {
	total = 3 // Seed + Genre + entity count
	if before.Seed != after.Seed {
		diff++
	}
	if before.Genre != after.Genre {
		diff++
	}
	if len(before.Entities) != len(after.Entities) {
		diff++
	}
	return total, diff
}

// buildEntityMap creates a lookup map from entity ID to data.
func buildEntityMap(entities []EntityData) map[uint64]EntityData {
	m := make(map[uint64]EntityData, len(entities))
	for _, e := range entities {
		m[e.ID] = e
	}
	return m
}

// compareEntities compares entity lists and returns total/diff field counts.
func compareEntities(beforeEntities, afterEntities []EntityData) (total, diff int) {
	beforeMap := buildEntityMap(beforeEntities)

	for _, afterEntity := range afterEntities {
		beforeEntity, exists := beforeMap[afterEntity.ID]
		if !exists {
			total++
			diff++
			continue
		}
		fieldDiff := compareEntity(beforeEntity, afterEntity)
		total += fieldDiff.total
		diff += fieldDiff.diff
	}
	return total, diff
}

type fieldComparison struct {
	total int
	diff  int
}

// comparePosition compares position components between two entity data structures.
func comparePosition(before, after EntityData) fieldComparison {
	result := fieldComparison{}
	if before.HasPosition || after.HasPosition {
		result.total = 4
		if before.PosX != after.PosX {
			result.diff++
		}
		if before.PosY != after.PosY {
			result.diff++
		}
		if before.PosZ != after.PosZ {
			result.diff++
		}
		if before.PosAngle != after.PosAngle {
			result.diff++
		}
	}
	return result
}

// compareHealth compares health components between two entity data structures.
func compareHealth(before, after EntityData) fieldComparison {
	result := fieldComparison{}
	if before.HasHealth || after.HasHealth {
		result.total = 2
		if before.HealthCurrent != after.HealthCurrent {
			result.diff++
		}
		if before.HealthMax != after.HealthMax {
			result.diff++
		}
	}
	return result
}

// compareCrime compares crime components between two entity data structures.
func compareCrime(before, after EntityData) fieldComparison {
	result := fieldComparison{}
	if before.HasCrime || after.HasCrime {
		result.total = 3
		if before.WantedLevel != after.WantedLevel {
			result.diff++
		}
		if before.BountyAmount != after.BountyAmount {
			result.diff++
		}
		if before.InJail != after.InJail {
			result.diff++
		}
	}
	return result
}

// compareEntity compares two entity data structures.
func compareEntity(before, after EntityData) fieldComparison {
	result := fieldComparison{}

	pos := comparePosition(before, after)
	result.total += pos.total
	result.diff += pos.diff

	health := compareHealth(before, after)
	result.total += health.total
	result.diff += health.diff

	crime := compareCrime(before, after)
	result.total += crime.total
	result.diff += crime.diff

	return result
}

// NewEntityData creates an empty EntityData with the given ID.
func NewEntityData(id uint64) EntityData {
	return EntityData{
		ID:              id,
		SkillLevels:     make(map[string]int),
		SkillExperience: make(map[string]float64),
		SchoolBonuses:   make(map[string]float64),
		QuestFlags:      make(map[string]bool),
		LockedBranches:  make(map[string]bool),
	}
}

// NewWorldSnapshot creates an empty world snapshot.
func NewWorldSnapshot(seed int64, genre string) *WorldSnapshot {
	return &WorldSnapshot{
		Seed:         seed,
		Genre:        genre,
		Timestamp:    time.Now(),
		Version:      1,
		Entities:     make([]EntityData, 0),
		EconomyNodes: make([]EconomyNodeData, 0),
		QuestFlags:   make(map[uint64]map[string]bool),
	}
}
