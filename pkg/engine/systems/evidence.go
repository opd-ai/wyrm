// Package systems implements ECS system logic.
package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/seedutil"
)

// ============================================================================
// Crime Evidence System
// ============================================================================

// EvidenceType represents different types of crime evidence.
type EvidenceType int

const (
	// EvidenceTypeFingerprint is physical evidence left at a crime scene.
	EvidenceTypeFingerprint EvidenceType = iota
	// EvidenceTypeWeapon is a weapon used in the crime.
	EvidenceTypeWeapon
	// EvidenceTypeWitness is eyewitness testimony.
	EvidenceTypeWitness
	// EvidenceTypeFootprint is tracks left at the scene.
	EvidenceTypeFootprint
	// EvidenceTypeBlood is biological evidence.
	EvidenceTypeBlood
	// EvidenceTypeDocument is written evidence (forged documents, contracts).
	EvidenceTypeDocument
	// EvidenceTypeStolenGoods is recovered stolen property.
	EvidenceTypeStolenGoods
	// EvidenceTypeMagical is magical residue or traces (fantasy genre).
	EvidenceTypeMagical
	// EvidenceTypeDigital is electronic evidence (cyberpunk/sci-fi genre).
	EvidenceTypeDigital
)

// String returns a human-readable name for the evidence type.
func (e EvidenceType) String() string {
	switch e {
	case EvidenceTypeFingerprint:
		return "Fingerprint"
	case EvidenceTypeWeapon:
		return "Weapon"
	case EvidenceTypeWitness:
		return "Witness Testimony"
	case EvidenceTypeFootprint:
		return "Footprint"
	case EvidenceTypeBlood:
		return "Blood Sample"
	case EvidenceTypeDocument:
		return "Document"
	case EvidenceTypeStolenGoods:
		return "Stolen Goods"
	case EvidenceTypeMagical:
		return "Magical Trace"
	case EvidenceTypeDigital:
		return "Digital Record"
	default:
		return "Unknown Evidence"
	}
}

// EvidenceStrength returns base conviction strength for this evidence type.
func (e EvidenceType) EvidenceStrength() float64 {
	switch e {
	case EvidenceTypeFingerprint:
		return 0.7
	case EvidenceTypeWeapon:
		return 0.8
	case EvidenceTypeWitness:
		return 0.5
	case EvidenceTypeFootprint:
		return 0.3
	case EvidenceTypeBlood:
		return 0.85
	case EvidenceTypeDocument:
		return 0.6
	case EvidenceTypeStolenGoods:
		return 0.9
	case EvidenceTypeMagical:
		return 0.75
	case EvidenceTypeDigital:
		return 0.95
	default:
		return 0.1
	}
}

// Evidence represents a single piece of crime evidence.
type Evidence struct {
	ID          string       // Unique evidence identifier
	Type        EvidenceType // Type of evidence
	CrimeID     string       // Crime this evidence relates to
	SuspectID   uint64       // Entity suspected (0 if unknown)
	CollectedAt float64      // Game time when collected
	LocationX   float64      // X coordinate where found
	LocationZ   float64      // Z coordinate where found
	Quality     float64      // Evidence quality 0.0-1.0 (affects conviction)
	IsTampered  bool         // Has evidence been tampered with
	IsProcessed bool         // Has evidence been analyzed by authorities
	CanDecay    bool         // Can this evidence decay over time
	DecayTime   float64      // Time until evidence becomes unusable
	Description string       // Human-readable description
}

// GetConvictionStrength returns the effective strength for conviction.
func (e *Evidence) GetConvictionStrength() float64 {
	strength := e.Type.EvidenceStrength() * e.Quality
	if e.IsTampered {
		strength *= 0.1 // Tampered evidence is almost worthless
	}
	if !e.IsProcessed {
		strength *= 0.5 // Unprocessed evidence is less useful
	}
	return strength
}

// CrimeRecord represents a record of a committed crime.
type CrimeRecord struct {
	ID              string   // Unique crime identifier
	CrimeType       string   // Type of crime (theft, murder, assault, etc.)
	CriminalID      uint64   // Entity that committed the crime (0 if unknown)
	VictimID        uint64   // Entity that was the victim (0 if none)
	LocationX       float64  // X coordinate
	LocationZ       float64  // Z coordinate
	Timestamp       float64  // When the crime occurred
	IsReported      bool     // Has the crime been reported to authorities
	IsInvestigated  bool     // Has the crime been investigated
	IsSolved        bool     // Has the criminal been identified
	IsConvicted     bool     // Has conviction occurred
	EvidenceIDs     []string // Evidence associated with this crime
	WitnessIDs      []uint64 // Witnesses who saw the crime
	Severity        int      // Crime severity 1-5 (affects punishment)
	ConvictionScore float64  // Cumulative evidence strength
}

// CrimeEvidenceSystem manages crime evidence collection, processing, and investigation.
type CrimeEvidenceSystem struct {
	crimeSystem *CrimeSystem
	// Evidence storage
	evidenceByID    map[string]*Evidence
	evidenceByCrime map[string][]*Evidence // CrimeID -> Evidence list
	// Crime records
	crimeRecords   map[string]*CrimeRecord
	crimesByEntity map[uint64][]string // Entity ID -> Crime IDs
	// Investigation settings
	EvidenceDecayRate  float64 // Rate at which evidence degrades
	InvestigationRange float64 // Range to search for evidence
	MinConvictionScore float64 // Minimum score for conviction
	ProcessingTime     float64 // Time to process evidence
	// Genre-specific modifiers
	genre string
	// Tracking
	gameTime     float64
	nextCrimeID  int
	nextEvidence int
	// Random generator
	rng *PseudoRandomLCG
}

// NewCrimeEvidenceSystem creates a new crime evidence system.
func NewCrimeEvidenceSystem(crimeSystem *CrimeSystem, genre string, seed int64) *CrimeEvidenceSystem {
	return &CrimeEvidenceSystem{
		crimeSystem:        crimeSystem,
		evidenceByID:       make(map[string]*Evidence),
		evidenceByCrime:    make(map[string][]*Evidence),
		crimeRecords:       make(map[string]*CrimeRecord),
		crimesByEntity:     make(map[uint64][]string),
		EvidenceDecayRate:  0.01, // 1% quality loss per minute
		InvestigationRange: 50.0, // 50 unit search radius
		MinConvictionScore: 0.75, // 75% evidence needed
		ProcessingTime:     30.0, // 30 seconds to process evidence
		genre:              genre,
		rng:                NewPseudoRandomLCG(seed),
	}
}

// Update processes evidence decay and ongoing investigations.
func (s *CrimeEvidenceSystem) Update(w *ecs.World, dt float64) {
	s.gameTime += dt
	s.processEvidenceDecay(dt)
	s.processInvestigations(w)
}

// processEvidenceDecay degrades evidence quality over time.
func (s *CrimeEvidenceSystem) processEvidenceDecay(dt float64) {
	for _, evidence := range s.evidenceByID {
		if !evidence.CanDecay || evidence.IsProcessed {
			continue
		}
		evidence.Quality -= s.EvidenceDecayRate * dt
		if evidence.Quality < 0 {
			evidence.Quality = 0
		}
		// Check decay timer
		if evidence.DecayTime > 0 {
			evidence.DecayTime -= dt
			if evidence.DecayTime <= 0 {
				evidence.Quality = 0
			}
		}
	}
}

// processInvestigations updates ongoing crime investigations.
func (s *CrimeEvidenceSystem) processInvestigations(w *ecs.World) {
	for _, record := range s.crimeRecords {
		if !record.IsReported || record.IsConvicted {
			continue
		}
		// Recalculate conviction score based on current evidence
		score := s.CalculateConvictionScore(record.ID)
		record.ConvictionScore = score
		// Check if enough evidence for conviction
		if score >= s.MinConvictionScore && !record.IsConvicted {
			s.convictCriminal(w, record)
		}
	}
}

// convictCriminal handles the conviction process.
func (s *CrimeEvidenceSystem) convictCriminal(w *ecs.World, record *CrimeRecord) {
	if record.CriminalID == 0 {
		return // Cannot convict unknown criminal
	}
	record.IsConvicted = true
	record.IsSolved = true
	// Apply punishment based on severity
	comp, ok := w.GetComponent(ecs.Entity(record.CriminalID), "Crime")
	if !ok {
		return
	}
	crime := comp.(*components.Crime)
	// Increase wanted level based on severity
	crime.WantedLevel += record.Severity
	if crime.WantedLevel > MaxWantedLevel {
		crime.WantedLevel = MaxWantedLevel
	}
	// Add bounty
	if s.crimeSystem != nil {
		crime.BountyAmount += float64(record.Severity) * s.crimeSystem.BountyPerLevel
	}
}

// RecordCrime creates a new crime record.
func (s *CrimeEvidenceSystem) RecordCrime(crimeType string, criminalID, victimID uint64, x, z float64, severity int) string {
	s.nextCrimeID++
	id := formatCrimeID(s.nextCrimeID)
	record := &CrimeRecord{
		ID:          id,
		CrimeType:   crimeType,
		CriminalID:  criminalID,
		VictimID:    victimID,
		LocationX:   x,
		LocationZ:   z,
		Timestamp:   s.gameTime,
		Severity:    severity,
		EvidenceIDs: make([]string, 0),
		WitnessIDs:  make([]uint64, 0),
	}
	// Clamp severity
	if record.Severity < 1 {
		record.Severity = 1
	}
	if record.Severity > 5 {
		record.Severity = 5
	}
	s.crimeRecords[id] = record
	// Track by criminal
	if criminalID != 0 {
		s.crimesByEntity[criminalID] = append(s.crimesByEntity[criminalID], id)
	}
	return id
}

// formatCrimeID creates a crime ID string.
func formatCrimeID(n int) string {
	return seedutil.FormatPrefixedID("CR", n)
}

// formatEvidenceID creates an evidence ID string.
func formatEvidenceID(n int) string {
	return seedutil.FormatPrefixedID("EV", n)
}

// ReportCrime marks a crime as reported to authorities.
func (s *CrimeEvidenceSystem) ReportCrime(crimeID string) bool {
	record, ok := s.crimeRecords[crimeID]
	if !ok {
		return false
	}
	record.IsReported = true
	return true
}

// AddWitness adds a witness to a crime record.
func (s *CrimeEvidenceSystem) AddWitness(crimeID string, witnessID uint64) bool {
	record, ok := s.crimeRecords[crimeID]
	if !ok {
		return false
	}
	// Check for duplicate
	for _, w := range record.WitnessIDs {
		if w == witnessID {
			return false
		}
	}
	record.WitnessIDs = append(record.WitnessIDs, witnessID)
	// Create witness testimony evidence
	s.CreateEvidence(crimeID, EvidenceTypeWitness, record.CriminalID, record.LocationX, record.LocationZ, 0.8)
	return true
}

// CreateEvidence creates new evidence for a crime.
func (s *CrimeEvidenceSystem) CreateEvidence(crimeID string, evidenceType EvidenceType, suspectID uint64, x, z, quality float64) *Evidence {
	record, ok := s.crimeRecords[crimeID]
	if !ok {
		return nil
	}
	s.nextEvidence++
	id := formatEvidenceID(s.nextEvidence)
	evidence := &Evidence{
		ID:          id,
		Type:        evidenceType,
		CrimeID:     crimeID,
		SuspectID:   suspectID,
		CollectedAt: s.gameTime,
		LocationX:   x,
		LocationZ:   z,
		Quality:     quality,
		CanDecay:    evidenceType != EvidenceTypeDigital, // Digital evidence doesn't decay
		DecayTime:   s.getDecayTimeForType(evidenceType),
		Description: s.generateEvidenceDescription(evidenceType),
	}
	// Clamp quality
	if evidence.Quality < 0 {
		evidence.Quality = 0
	}
	if evidence.Quality > 1.0 {
		evidence.Quality = 1.0
	}
	s.evidenceByID[id] = evidence
	s.evidenceByCrime[crimeID] = append(s.evidenceByCrime[crimeID], evidence)
	record.EvidenceIDs = append(record.EvidenceIDs, id)
	return evidence
}

// getDecayTimeForType returns the decay time for an evidence type.
func (s *CrimeEvidenceSystem) getDecayTimeForType(evidenceType EvidenceType) float64 {
	switch evidenceType {
	case EvidenceTypeFingerprint:
		return 600.0 // 10 minutes
	case EvidenceTypeFootprint:
		return 300.0 // 5 minutes
	case EvidenceTypeBlood:
		return 900.0 // 15 minutes
	case EvidenceTypeWeapon:
		return 0 // Doesn't decay
	case EvidenceTypeDocument:
		return 0 // Doesn't decay
	case EvidenceTypeStolenGoods:
		return 0 // Doesn't decay
	case EvidenceTypeMagical:
		return 180.0 // 3 minutes (magic fades quickly)
	case EvidenceTypeDigital:
		return 0 // Doesn't decay
	default:
		return 300.0
	}
}

// generateEvidenceDescription creates genre-appropriate description.
func (s *CrimeEvidenceSystem) generateEvidenceDescription(evidenceType EvidenceType) string {
	switch s.genre {
	case "fantasy":
		return s.fantasyDescription(evidenceType)
	case "sci-fi":
		return s.sciFiDescription(evidenceType)
	case "horror":
		return s.horrorDescription(evidenceType)
	case "cyberpunk":
		return s.cyberpunkDescription(evidenceType)
	case "post-apocalyptic":
		return s.postApocDescription(evidenceType)
	default:
		return s.fantasyDescription(evidenceType)
	}
}

func (s *CrimeEvidenceSystem) fantasyDescription(evidenceType EvidenceType) string {
	switch evidenceType {
	case EvidenceTypeFingerprint:
		return "Residual fingerprints found at the scene."
	case EvidenceTypeWeapon:
		return "A weapon, possibly used in the crime."
	case EvidenceTypeWitness:
		return "Testimony from one who witnessed the deed."
	case EvidenceTypeFootprint:
		return "Boot prints leading away from the scene."
	case EvidenceTypeBlood:
		return "Blood stains that may identify the perpetrator."
	case EvidenceTypeMagical:
		return "Traces of arcane energy linger here."
	default:
		return "Evidence relating to the crime."
	}
}

func (s *CrimeEvidenceSystem) sciFiDescription(evidenceType EvidenceType) string {
	switch evidenceType {
	case EvidenceTypeFingerprint:
		return "Biometric fingerprint scan captured."
	case EvidenceTypeWeapon:
		return "Weapon recovered, serial number logged."
	case EvidenceTypeWitness:
		return "Witness statement recorded and verified."
	case EvidenceTypeDigital:
		return "Digital forensics data recovered from systems."
	case EvidenceTypeBlood:
		return "DNA sample collected for analysis."
	default:
		return "Evidence logged in the investigation database."
	}
}

func (s *CrimeEvidenceSystem) horrorDescription(evidenceType EvidenceType) string {
	switch evidenceType {
	case EvidenceTypeFingerprint:
		return "Smeared fingerprints, partially obscured by... something."
	case EvidenceTypeWeapon:
		return "A weapon stained with dark purpose."
	case EvidenceTypeWitness:
		return "A survivor's trembling account of events."
	case EvidenceTypeBlood:
		return "Blood, but its color seems... wrong."
	case EvidenceTypeMagical:
		return "An unnatural residue that chills the soul."
	default:
		return "Disturbing evidence of the incident."
	}
}

func (s *CrimeEvidenceSystem) cyberpunkDescription(evidenceType EvidenceType) string {
	switch evidenceType {
	case EvidenceTypeFingerprint:
		return "Biometric imprint captured by sensor grid."
	case EvidenceTypeWeapon:
		return "Smart-weapon traced via manufacturer backdoor."
	case EvidenceTypeWitness:
		return "Witness cyberware recording extracted."
	case EvidenceTypeDigital:
		return "Data trail recovered from the net."
	case EvidenceTypeBlood:
		return "Genetic material scanned and indexed."
	default:
		return "Evidence uploaded to investigation matrix."
	}
}

func (s *CrimeEvidenceSystem) postApocDescription(evidenceType EvidenceType) string {
	switch evidenceType {
	case EvidenceTypeFingerprint:
		return "Prints left in the dust and grime."
	case EvidenceTypeWeapon:
		return "A makeshift weapon, crude but effective."
	case EvidenceTypeWitness:
		return "A scavenger's account of what they saw."
	case EvidenceTypeFootprint:
		return "Tracks in the irradiated soil."
	case EvidenceTypeBlood:
		return "Blood, possibly contaminated with radiation."
	default:
		return "Evidence salvaged from the scene."
	}
}

// ProcessEvidence marks evidence as analyzed.
func (s *CrimeEvidenceSystem) ProcessEvidence(evidenceID string) bool {
	evidence, ok := s.evidenceByID[evidenceID]
	if !ok {
		return false
	}
	if evidence.Quality <= 0 {
		return false // Can't process destroyed evidence
	}
	evidence.IsProcessed = true
	return true
}

// TamperWithEvidence degrades evidence quality (criminal action).
func (s *CrimeEvidenceSystem) TamperWithEvidence(evidenceID string) bool {
	evidence, ok := s.evidenceByID[evidenceID]
	if !ok {
		return false
	}
	if evidence.IsProcessed {
		return false // Can't tamper with already processed evidence
	}
	evidence.IsTampered = true
	evidence.Quality *= 0.3 // Reduce quality significantly
	return true
}

// DestroyEvidence removes evidence from the system.
func (s *CrimeEvidenceSystem) DestroyEvidence(evidenceID string) bool {
	evidence, ok := s.evidenceByID[evidenceID]
	if !ok {
		return false
	}
	if evidence.IsProcessed {
		return false // Can't destroy processed evidence
	}
	evidence.Quality = 0
	evidence.CanDecay = false // Mark as destroyed
	return true
}

// CalculateConvictionScore calculates total evidence strength for a crime.
func (s *CrimeEvidenceSystem) CalculateConvictionScore(crimeID string) float64 {
	evidenceList := s.evidenceByCrime[crimeID]
	if len(evidenceList) == 0 {
		return 0
	}
	totalStrength := 0.0
	for _, evidence := range evidenceList {
		strength := evidence.GetConvictionStrength()
		totalStrength += strength
	}
	// Cap at 1.0 (100% certain)
	if totalStrength > 1.0 {
		totalStrength = 1.0
	}
	return totalStrength
}

// GetCrimeRecord returns a crime record by ID.
func (s *CrimeEvidenceSystem) GetCrimeRecord(crimeID string) *CrimeRecord {
	return s.crimeRecords[crimeID]
}

// GetEvidence returns evidence by ID.
func (s *CrimeEvidenceSystem) GetEvidence(evidenceID string) *Evidence {
	return s.evidenceByID[evidenceID]
}

// GetEvidenceForCrime returns all evidence for a crime.
func (s *CrimeEvidenceSystem) GetEvidenceForCrime(crimeID string) []*Evidence {
	return s.evidenceByCrime[crimeID]
}

// GetCrimesForEntity returns all crime IDs involving an entity.
func (s *CrimeEvidenceSystem) GetCrimesForEntity(entityID uint64) []string {
	return s.crimesByEntity[entityID]
}

// GetUnsolvedCrimes returns all unsolved reported crimes.
func (s *CrimeEvidenceSystem) GetUnsolvedCrimes() []*CrimeRecord {
	var unsolved []*CrimeRecord
	for _, record := range s.crimeRecords {
		if record.IsReported && !record.IsSolved {
			unsolved = append(unsolved, record)
		}
	}
	return unsolved
}

// InvestigateCrimeScene generates evidence at a crime location.
func (s *CrimeEvidenceSystem) InvestigateCrimeScene(crimeID string) int {
	record, ok := s.crimeRecords[crimeID]
	if !ok {
		return 0
	}
	if record.IsInvestigated {
		return 0 // Already investigated
	}
	record.IsInvestigated = true
	// Generate evidence based on crime type and genre
	evidenceCount := s.generateSceneEvidence(record)
	return evidenceCount
}

// generateSceneEvidence creates evidence appropriate for the crime.
func (s *CrimeEvidenceSystem) generateSceneEvidence(record *CrimeRecord) int {
	count := 0
	count += s.generateCommonEvidence(record)
	count += s.generateCrimeTypeEvidence(record)
	count += s.generateGenreEvidence(record)
	return count
}

// generateCommonEvidence generates fingerprints and footprints.
func (s *CrimeEvidenceSystem) generateCommonEvidence(record *CrimeRecord) int {
	count := 0
	if s.pseudoRandom() < 0.7 {
		s.CreateEvidence(record.ID, EvidenceTypeFingerprint, record.CriminalID, record.LocationX, record.LocationZ, 0.6+s.pseudoRandom()*0.3)
		count++
	}
	if s.pseudoRandom() < 0.6 {
		s.CreateEvidence(record.ID, EvidenceTypeFootprint, record.CriminalID, record.LocationX, record.LocationZ, 0.5+s.pseudoRandom()*0.4)
		count++
	}
	return count
}

// generateCrimeTypeEvidence generates evidence specific to the crime type.
func (s *CrimeEvidenceSystem) generateCrimeTypeEvidence(record *CrimeRecord) int {
	count := 0
	switch record.CrimeType {
	case "murder", "assault":
		count += s.generateViolentCrimeEvidence(record)
	case "theft", "robbery":
		count += s.generateTheftEvidence(record)
	case "fraud", "forgery":
		count += s.generateFraudEvidence(record)
	}
	return count
}

// generateViolentCrimeEvidence generates blood and weapon evidence.
func (s *CrimeEvidenceSystem) generateViolentCrimeEvidence(record *CrimeRecord) int {
	count := 0
	if s.pseudoRandom() < 0.8 {
		s.CreateEvidence(record.ID, EvidenceTypeBlood, record.CriminalID, record.LocationX, record.LocationZ, 0.7+s.pseudoRandom()*0.2)
		count++
	}
	if s.pseudoRandom() < 0.5 {
		s.CreateEvidence(record.ID, EvidenceTypeWeapon, record.CriminalID, record.LocationX, record.LocationZ, 0.8)
		count++
	}
	return count
}

// generateTheftEvidence generates stolen goods evidence.
func (s *CrimeEvidenceSystem) generateTheftEvidence(record *CrimeRecord) int {
	if s.pseudoRandom() < 0.4 {
		s.CreateEvidence(record.ID, EvidenceTypeStolenGoods, record.CriminalID, record.LocationX, record.LocationZ, 0.9)
		return 1
	}
	return 0
}

// generateFraudEvidence generates document evidence.
func (s *CrimeEvidenceSystem) generateFraudEvidence(record *CrimeRecord) int {
	if s.pseudoRandom() < 0.7 {
		s.CreateEvidence(record.ID, EvidenceTypeDocument, record.CriminalID, record.LocationX, record.LocationZ, 0.7)
		return 1
	}
	return 0
}

// generateGenreEvidence generates evidence specific to the game genre.
func (s *CrimeEvidenceSystem) generateGenreEvidence(record *CrimeRecord) int {
	switch s.genre {
	case "fantasy":
		if s.pseudoRandom() < 0.3 {
			s.CreateEvidence(record.ID, EvidenceTypeMagical, record.CriminalID, record.LocationX, record.LocationZ, 0.6+s.pseudoRandom()*0.3)
			return 1
		}
	case "sci-fi", "cyberpunk":
		if s.pseudoRandom() < 0.5 {
			s.CreateEvidence(record.ID, EvidenceTypeDigital, record.CriminalID, record.LocationX, record.LocationZ, 0.8+s.pseudoRandom()*0.15)
			return 1
		}
	}
	return 0
}

// pseudoRandom generates a deterministic pseudo-random number 0.0-1.0.
func (s *CrimeEvidenceSystem) pseudoRandom() float64 {
	return s.rng.Float64()
}

// GetCrimeTypeDescription returns a description of a crime type.
func (s *CrimeEvidenceSystem) GetCrimeTypeDescription(crimeType string) string {
	switch crimeType {
	case "murder":
		return "The unlawful killing of another person."
	case "assault":
		return "Physical attack causing injury to another."
	case "theft":
		return "Taking property belonging to another without consent."
	case "robbery":
		return "Theft using force or threat of force."
	case "burglary":
		return "Unlawful entry with intent to commit a crime."
	case "fraud":
		return "Deception for personal or financial gain."
	case "forgery":
		return "Creating false documents or counterfeits."
	case "trespass":
		return "Entering property without permission."
	case "vandalism":
		return "Willful destruction of property."
	case "smuggling":
		return "Illegal transport of goods or persons."
	default:
		return "A criminal offense."
	}
}

// GetCrimeSeverity returns the default severity for a crime type.
func (s *CrimeEvidenceSystem) GetCrimeSeverity(crimeType string) int {
	switch crimeType {
	case "murder":
		return 5
	case "assault":
		return 3
	case "robbery":
		return 4
	case "theft":
		return 2
	case "burglary":
		return 3
	case "fraud":
		return 2
	case "forgery":
		return 2
	case "trespass":
		return 1
	case "vandalism":
		return 1
	case "smuggling":
		return 3
	default:
		return 2
	}
}

// ClearRecord clears a crime record (for pardons).
func (s *CrimeEvidenceSystem) ClearRecord(crimeID string) bool {
	record, ok := s.crimeRecords[crimeID]
	if !ok {
		return false
	}
	// Mark as resolved without conviction
	record.IsSolved = true
	record.ConvictionScore = 0
	// Remove from entity tracking
	if record.CriminalID != 0 {
		crimes := s.crimesByEntity[record.CriminalID]
		for i, id := range crimes {
			if id == crimeID {
				s.crimesByEntity[record.CriminalID] = append(crimes[:i], crimes[i+1:]...)
				break
			}
		}
	}
	return true
}
