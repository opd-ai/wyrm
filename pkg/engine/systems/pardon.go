// Package systems implements ECS system logic.
package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// ============================================================================
// Pardons and Amnesty System
// ============================================================================

// PardonType represents different types of pardons.
type PardonType int

const (
	// PardonTypeFull completely clears criminal record.
	PardonTypeFull PardonType = iota
	// PardonTypePartial reduces wanted level and bounty.
	PardonTypePartial
	// PardonTypeAmnesty is a faction-wide or region-wide pardon event.
	PardonTypeAmnesty
	// PardonTypeBribed obtained through corruption.
	PardonTypeBribed
	// PardonTypeService earned through community service.
	PardonTypeService
	// PardonTypePolitical granted for political reasons.
	PardonTypePolitical
	// PardonTypeReligious granted by religious authority.
	PardonTypeReligious
	// PardonTypeMilitary granted for military service.
	PardonTypeMilitary
)

// String returns the pardon type name.
func (p PardonType) String() string {
	switch p {
	case PardonTypeFull:
		return "Full Pardon"
	case PardonTypePartial:
		return "Partial Pardon"
	case PardonTypeAmnesty:
		return "General Amnesty"
	case PardonTypeBribed:
		return "Bribed Pardon"
	case PardonTypeService:
		return "Service Pardon"
	case PardonTypePolitical:
		return "Political Pardon"
	case PardonTypeReligious:
		return "Religious Absolution"
	case PardonTypeMilitary:
		return "Military Pardon"
	default:
		return "Unknown"
	}
}

// PardonRequirement defines what's needed to obtain a pardon.
type PardonRequirement struct {
	Type              PardonType // Type of pardon
	GoldCost          int        // Currency required
	ReputationMin     float64    // Minimum reputation with granting faction
	QuestRequired     string     // Quest that must be completed (empty if none)
	ServiceHours      float64    // Hours of community service required
	ItemRequired      string     // Item that must be turned in
	WantedLevelMax    int        // Maximum wanted level that can be pardoned
	Description       string     // Human-readable description
	AvailableInRegion string     // Region where this pardon is available
}

// PardonRecord tracks a granted pardon.
type PardonRecord struct {
	ID             string     // Unique pardon identifier
	Type           PardonType // Type of pardon
	RecipientID    uint64     // Entity that received the pardon
	GranterID      uint64     // Entity or faction that granted the pardon
	GranterName    string     // Name of granting authority
	Timestamp      float64    // When pardon was granted
	CrimesCleared  []string   // Crime IDs that were cleared
	WasWantedLevel int        // Wanted level before pardon
	BountyCleared  float64    // Bounty amount cleared
	Conditions     []string   // Any conditions attached to the pardon
}

// AmnestyEvent represents a mass pardon event.
type AmnestyEvent struct {
	ID             string   // Unique event identifier
	Name           string   // Event name
	Description    string   // Event description
	StartTime      float64  // When amnesty begins
	EndTime        float64  // When amnesty ends
	Region         string   // Affected region (empty = worldwide)
	FactionID      string   // Faction granting amnesty (empty = government)
	MaxWantedLevel int      // Maximum wanted level covered
	CrimeTypes     []string // Crime types covered (empty = all)
	ParticipantIDs []uint64 // Entities who received amnesty
	IsActive       bool     // Currently active
}

// PardonSystem manages pardons and amnesty events.
type PardonSystem struct {
	crimeSystem         *CrimeSystem
	crimeEvidenceSystem *CrimeEvidenceSystem
	// Pardon storage
	pardons         map[string]*PardonRecord
	pardonsByEntity map[uint64][]string // Entity ID -> Pardon IDs
	// Amnesty events
	amnestyEvents map[string]*AmnestyEvent
	activeAmnesty []string // Currently active amnesty event IDs
	// Requirements for obtaining pardons
	pardonRequirements map[PardonType]*PardonRequirement
	// Settings
	genre string
	// Tracking
	gameTime      float64
	nextPardonID  int
	nextAmnestyID int
	// Random seed
	rngSeed    int64
	rngCounter int64
	// Configuration
	BaseBribeCost          int     // Base cost to bribe for pardon
	BaseServiceHours       float64 // Base hours for service pardon
	PartialPardonReduction float64 // How much partial pardon reduces wanted level (0.5 = 50%)
}

// NewPardonSystem creates a new pardon system.
func NewPardonSystem(crimeSystem *CrimeSystem, evidenceSystem *CrimeEvidenceSystem, genre string, seed int64) *PardonSystem {
	ps := &PardonSystem{
		crimeSystem:            crimeSystem,
		crimeEvidenceSystem:    evidenceSystem,
		pardons:                make(map[string]*PardonRecord),
		pardonsByEntity:        make(map[uint64][]string),
		amnestyEvents:          make(map[string]*AmnestyEvent),
		activeAmnesty:          make([]string, 0),
		pardonRequirements:     make(map[PardonType]*PardonRequirement),
		genre:                  genre,
		rngSeed:                seed,
		BaseBribeCost:          1000,
		BaseServiceHours:       10.0,
		PartialPardonReduction: 0.5,
	}
	ps.initializeRequirements()
	return ps
}

// initializeRequirements sets up default pardon requirements.
func (s *PardonSystem) initializeRequirements() {
	s.pardonRequirements[PardonTypeFull] = &PardonRequirement{
		Type:           PardonTypeFull,
		GoldCost:       5000,
		ReputationMin:  50.0,
		WantedLevelMax: 3,
		Description:    "A complete pardon that clears all crimes and bounties.",
	}
	s.pardonRequirements[PardonTypePartial] = &PardonRequirement{
		Type:           PardonTypePartial,
		GoldCost:       1000,
		ReputationMin:  0.0,
		WantedLevelMax: 5,
		Description:    "Reduces wanted level and bounty but does not clear record.",
	}
	s.pardonRequirements[PardonTypeAmnesty] = &PardonRequirement{
		Type:           PardonTypeAmnesty,
		GoldCost:       0,
		ReputationMin:  -100.0,
		WantedLevelMax: 5,
		Description:    "A general amnesty granted during special events.",
	}
	s.pardonRequirements[PardonTypeBribed] = &PardonRequirement{
		Type:           PardonTypeBribed,
		GoldCost:       3000,
		ReputationMin:  -100.0, // Available even with bad rep
		WantedLevelMax: 4,
		Description:    "An unofficial pardon obtained through bribery.",
	}
	s.pardonRequirements[PardonTypeService] = &PardonRequirement{
		Type:           PardonTypeService,
		GoldCost:       0,
		ServiceHours:   20.0,
		ReputationMin:  -50.0,
		WantedLevelMax: 2,
		Description:    "Earn a pardon through community service.",
	}
	s.pardonRequirements[PardonTypePolitical] = &PardonRequirement{
		Type:           PardonTypePolitical,
		GoldCost:       0,
		ReputationMin:  75.0,
		QuestRequired:  "political_favor",
		WantedLevelMax: 5,
		Description:    "A pardon granted for political services rendered.",
	}
	s.pardonRequirements[PardonTypeReligious] = &PardonRequirement{
		Type:           PardonTypeReligious,
		GoldCost:       500,
		ReputationMin:  25.0,
		WantedLevelMax: 3,
		Description:    s.getGenreReligiousDesc(),
	}
	s.pardonRequirements[PardonTypeMilitary] = &PardonRequirement{
		Type:           PardonTypeMilitary,
		GoldCost:       0,
		ServiceHours:   50.0,
		QuestRequired:  "military_service",
		WantedLevelMax: 4,
		Description:    "Earn a pardon through military service.",
	}
}

// getGenreReligiousDesc returns genre-appropriate religious pardon description.
func (s *PardonSystem) getGenreReligiousDesc() string {
	switch s.genre {
	case "fantasy":
		return "Seek absolution from the temple priests."
	case "sci-fi":
		return "Obtain clearance from the Ethics Council."
	case "horror":
		return "Perform the ritual of atonement."
	case "cyberpunk":
		return "Hack your criminal records through a priest-hacker."
	case "post-apocalyptic":
		return "Prove your worth to the settlement elders."
	default:
		return "Seek spiritual absolution."
	}
}

// Update processes amnesty events.
func (s *PardonSystem) Update(w *ecs.World, dt float64) {
	s.gameTime += dt
	s.processAmnestyEvents(w)
}

// processAmnestyEvents handles active amnesty events.
func (s *PardonSystem) processAmnestyEvents(w *ecs.World) {
	// Check for expired events
	for i := len(s.activeAmnesty) - 1; i >= 0; i-- {
		eventID := s.activeAmnesty[i]
		event := s.amnestyEvents[eventID]
		if event == nil {
			continue
		}
		if event.EndTime > 0 && s.gameTime > event.EndTime {
			event.IsActive = false
			// Remove from active list
			s.activeAmnesty = append(s.activeAmnesty[:i], s.activeAmnesty[i+1:]...)
		}
	}
}

// CanObtainPardon checks if an entity can obtain a specific pardon type.
func (s *PardonSystem) CanObtainPardon(w *ecs.World, entity ecs.Entity, pardonType PardonType) (bool, string) {
	req := s.pardonRequirements[pardonType]
	if req == nil {
		return false, "Unknown pardon type"
	}
	// Check wanted level
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return false, "No criminal record"
	}
	crime := comp.(*components.Crime)
	if crime.WantedLevel > req.WantedLevelMax {
		return false, "Wanted level too high for this pardon"
	}
	if crime.WantedLevel <= 0 {
		return false, "No crimes to pardon"
	}
	// Check reputation (would need faction system integration)
	// For now, skip reputation check
	// Check quest requirement
	if req.QuestRequired != "" {
		questComp, ok := w.GetComponent(entity, "Quest")
		if !ok {
			return false, "Required quest not completed"
		}
		quest := questComp.(*components.Quest)
		if quest.ID != req.QuestRequired || !quest.Completed {
			return false, "Required quest not completed"
		}
	}
	return true, ""
}

// GrantPardon grants a pardon to an entity.
func (s *PardonSystem) GrantPardon(w *ecs.World, entity ecs.Entity, pardonType PardonType, granterName string) (*PardonRecord, error) {
	canObtain, reason := s.CanObtainPardon(w, entity, pardonType)
	if !canObtain {
		return nil, &pardonError{reason}
	}
	comp, _ := w.GetComponent(entity, "Crime")
	crime := comp.(*components.Crime)
	// Create pardon record
	s.nextPardonID++
	id := formatPardonID(s.nextPardonID)
	record := &PardonRecord{
		ID:             id,
		Type:           pardonType,
		RecipientID:    uint64(entity),
		GranterName:    granterName,
		Timestamp:      s.gameTime,
		WasWantedLevel: crime.WantedLevel,
		BountyCleared:  crime.BountyAmount,
		CrimesCleared:  make([]string, 0),
	}
	// Apply pardon effects
	switch pardonType {
	case PardonTypeFull, PardonTypePolitical, PardonTypeMilitary:
		// Full pardon - clear everything
		record.CrimesCleared = s.clearAllCrimes(entity)
		crime.WantedLevel = 0
		crime.BountyAmount = 0
		crime.InJail = false
		crime.JailReleaseTime = 0
	case PardonTypePartial:
		// Partial - reduce but don't clear
		crime.WantedLevel = int(float64(crime.WantedLevel) * (1.0 - s.PartialPardonReduction))
		crime.BountyAmount *= (1.0 - s.PartialPardonReduction)
		record.BountyCleared = crime.BountyAmount * s.PartialPardonReduction
	case PardonTypeBribed, PardonTypeReligious, PardonTypeService:
		// Clear current crimes but leave record
		record.CrimesCleared = s.clearAllCrimes(entity)
		crime.WantedLevel = 0
		crime.BountyAmount = 0
	case PardonTypeAmnesty:
		// Amnesty - full clear
		record.CrimesCleared = s.clearAllCrimes(entity)
		crime.WantedLevel = 0
		crime.BountyAmount = 0
	}
	// Store record
	s.pardons[id] = record
	s.pardonsByEntity[uint64(entity)] = append(s.pardonsByEntity[uint64(entity)], id)
	return record, nil
}

// pardonError implements error interface for pardon errors.
type pardonError struct {
	message string
}

// Error returns the error message for the pardonError.
func (e *pardonError) Error() string {
	return e.message
}

// clearAllCrimes clears crimes from the evidence system.
func (s *PardonSystem) clearAllCrimes(entity ecs.Entity) []string {
	if s.crimeEvidenceSystem == nil {
		return nil
	}
	crimeIDs := s.crimeEvidenceSystem.GetCrimesForEntity(uint64(entity))
	cleared := make([]string, 0, len(crimeIDs))
	for _, crimeID := range crimeIDs {
		if s.crimeEvidenceSystem.ClearRecord(crimeID) {
			cleared = append(cleared, crimeID)
		}
	}
	return cleared
}

// formatPardonID creates a pardon ID string.
func formatPardonID(n int) string {
	result := make([]byte, 0, 10)
	result = append(result, 'P', 'D', '-')
	if n == 0 {
		return string(append(result, '0'))
	}
	digits := make([]byte, 0, 8)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	for i := len(digits) - 1; i >= 0; i-- {
		result = append(result, digits[i])
	}
	return string(result)
}

// formatAmnestyID creates an amnesty ID string.
func formatAmnestyID(n int) string {
	result := make([]byte, 0, 10)
	result = append(result, 'A', 'M', '-')
	if n == 0 {
		return string(append(result, '0'))
	}
	digits := make([]byte, 0, 8)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	for i := len(digits) - 1; i >= 0; i-- {
		result = append(result, digits[i])
	}
	return string(result)
}

// StartAmnestyEvent begins a new amnesty event.
func (s *PardonSystem) StartAmnestyEvent(name, description, region, factionID string, duration float64, maxWantedLevel int, crimeTypes []string) *AmnestyEvent {
	s.nextAmnestyID++
	id := formatAmnestyID(s.nextAmnestyID)
	event := &AmnestyEvent{
		ID:             id,
		Name:           name,
		Description:    description,
		StartTime:      s.gameTime,
		EndTime:        s.gameTime + duration,
		Region:         region,
		FactionID:      factionID,
		MaxWantedLevel: maxWantedLevel,
		CrimeTypes:     crimeTypes,
		ParticipantIDs: make([]uint64, 0),
		IsActive:       true,
	}
	if duration <= 0 {
		event.EndTime = 0 // No end time
	}
	s.amnestyEvents[id] = event
	s.activeAmnesty = append(s.activeAmnesty, id)
	return event
}

// ClaimAmnesty allows an entity to receive an amnesty pardon.
func (s *PardonSystem) ClaimAmnesty(w *ecs.World, entity ecs.Entity, eventID string) (*PardonRecord, error) {
	event := s.amnestyEvents[eventID]
	if event == nil {
		return nil, &pardonError{"Amnesty event not found"}
	}
	if !event.IsActive {
		return nil, &pardonError{"Amnesty event has ended"}
	}
	// Check wanted level
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return nil, &pardonError{"No criminal record"}
	}
	crime := comp.(*components.Crime)
	if crime.WantedLevel > event.MaxWantedLevel {
		return nil, &pardonError{"Wanted level too high for this amnesty"}
	}
	if crime.WantedLevel <= 0 {
		return nil, &pardonError{"No crimes to pardon"}
	}
	// Check if already claimed
	for _, pid := range event.ParticipantIDs {
		if pid == uint64(entity) {
			return nil, &pardonError{"Already claimed this amnesty"}
		}
	}
	// Grant amnesty pardon
	record, err := s.GrantPardon(w, entity, PardonTypeAmnesty, event.Name)
	if err != nil {
		return nil, err
	}
	// Track participation
	event.ParticipantIDs = append(event.ParticipantIDs, uint64(entity))
	return record, nil
}

// EndAmnestyEvent ends an amnesty event early.
func (s *PardonSystem) EndAmnestyEvent(eventID string) bool {
	event := s.amnestyEvents[eventID]
	if event == nil || !event.IsActive {
		return false
	}
	event.IsActive = false
	event.EndTime = s.gameTime
	// Remove from active list
	for i, id := range s.activeAmnesty {
		if id == eventID {
			s.activeAmnesty = append(s.activeAmnesty[:i], s.activeAmnesty[i+1:]...)
			break
		}
	}
	return true
}

// GetPardon returns a pardon record by ID.
func (s *PardonSystem) GetPardon(pardonID string) *PardonRecord {
	return s.pardons[pardonID]
}

// GetEntityPardons returns all pardons granted to an entity.
func (s *PardonSystem) GetEntityPardons(entityID uint64) []*PardonRecord {
	pardonIDs := s.pardonsByEntity[entityID]
	result := make([]*PardonRecord, 0, len(pardonIDs))
	for _, id := range pardonIDs {
		if pardon := s.pardons[id]; pardon != nil {
			result = append(result, pardon)
		}
	}
	return result
}

// GetAmnestyEvent returns an amnesty event by ID.
func (s *PardonSystem) GetAmnestyEvent(eventID string) *AmnestyEvent {
	return s.amnestyEvents[eventID]
}

// GetActiveAmnestyEvents returns all currently active amnesty events.
func (s *PardonSystem) GetActiveAmnestyEvents() []*AmnestyEvent {
	result := make([]*AmnestyEvent, 0, len(s.activeAmnesty))
	for _, id := range s.activeAmnesty {
		if event := s.amnestyEvents[id]; event != nil && event.IsActive {
			result = append(result, event)
		}
	}
	return result
}

// GetPardonRequirement returns the requirements for a pardon type.
func (s *PardonSystem) GetPardonRequirement(pardonType PardonType) *PardonRequirement {
	return s.pardonRequirements[pardonType]
}

// GetAvailablePardons returns pardon types available to an entity.
func (s *PardonSystem) GetAvailablePardons(w *ecs.World, entity ecs.Entity) []PardonType {
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return nil
	}
	crime := comp.(*components.Crime)
	if crime.WantedLevel <= 0 {
		return nil
	}
	var available []PardonType
	for pardonType := range s.pardonRequirements {
		canObtain, _ := s.CanObtainPardon(w, entity, pardonType)
		if canObtain {
			available = append(available, pardonType)
		}
	}
	return available
}

// CalculatePardonCost calculates the cost for a pardon.
func (s *PardonSystem) CalculatePardonCost(w *ecs.World, entity ecs.Entity, pardonType PardonType) int {
	req := s.pardonRequirements[pardonType]
	if req == nil {
		return 0
	}
	baseCost := req.GoldCost
	// Adjust based on wanted level
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return baseCost
	}
	crime := comp.(*components.Crime)
	// Higher wanted level = higher cost
	multiplier := 1.0 + float64(crime.WantedLevel)*0.2
	return int(float64(baseCost) * multiplier)
}

// GetPardonDescription returns a genre-appropriate description.
func (s *PardonSystem) GetPardonDescription(pardonType PardonType) string {
	switch s.genre {
	case "fantasy":
		return s.fantasyPardonDesc(pardonType)
	case "sci-fi":
		return s.sciFiPardonDesc(pardonType)
	case "horror":
		return s.horrorPardonDesc(pardonType)
	case "cyberpunk":
		return s.cyberpunkPardonDesc(pardonType)
	case "post-apocalyptic":
		return s.postApocPardonDesc(pardonType)
	default:
		return s.fantasyPardonDesc(pardonType)
	}
}

func (s *PardonSystem) fantasyPardonDesc(pardonType PardonType) string {
	switch pardonType {
	case PardonTypeFull:
		return "A royal decree absolving all crimes."
	case PardonTypePartial:
		return "A noble's intervention to reduce your sentence."
	case PardonTypeBribed:
		return "Gold speaks louder than law."
	case PardonTypeService:
		return "Labor in service to the realm."
	case PardonTypePolitical:
		return "A pardon for services to the crown."
	case PardonTypeReligious:
		return "Seek absolution at the temple."
	case PardonTypeMilitary:
		return "Serve in the army to clear your name."
	default:
		return "A path to redemption."
	}
}

func (s *PardonSystem) sciFiPardonDesc(pardonType PardonType) string {
	switch pardonType {
	case PardonTypeFull:
		return "Full criminal record expungement."
	case PardonTypePartial:
		return "Reduced security classification."
	case PardonTypeBribed:
		return "Credits can rewrite history."
	case PardonTypeService:
		return "Colonial service assignment."
	case PardonTypePolitical:
		return "Political asylum granted."
	case PardonTypeReligious:
		return "Ethics Council rehabilitation."
	case PardonTypeMilitary:
		return "Frontier defense conscription."
	default:
		return "Legal resolution available."
	}
}

func (s *PardonSystem) horrorPardonDesc(pardonType PardonType) string {
	switch pardonType {
	case PardonTypeFull:
		return "The Dark Council erases your sins."
	case PardonTypePartial:
		return "A blood pact reduces your debt."
	case PardonTypeBribed:
		return "The shadows accept payment."
	case PardonTypeService:
		return "Serve the coven to atone."
	case PardonTypePolitical:
		return "Favors for the Elder require... commitment."
	case PardonTypeReligious:
		return "The ritual of atonement awaits."
	case PardonTypeMilitary:
		return "Hunt the creatures that threaten us."
	default:
		return "Redemption comes at a price."
	}
}

func (s *PardonSystem) cyberpunkPardonDesc(pardonType PardonType) string {
	switch pardonType {
	case PardonTypeFull:
		return "Complete data wipe from all databases."
	case PardonTypePartial:
		return "Partial record scrub, some flags remain."
	case PardonTypeBribed:
		return "Eddies make problems disappear."
	case PardonTypeService:
		return "Corp contract work to clear your debt."
	case PardonTypePolitical:
		return "Political connections have their uses."
	case PardonTypeReligious:
		return "The techno-priests can rewrite your sins."
	case PardonTypeMilitary:
		return "Militech always needs expendable assets."
	default:
		return "There's always a way out... for a price."
	}
}

func (s *PardonSystem) postApocPardonDesc(pardonType PardonType) string {
	switch pardonType {
	case PardonTypeFull:
		return "The settlement council clears your record."
	case PardonTypePartial:
		return "Reduced exile time for good behavior."
	case PardonTypeBribed:
		return "Trade goods can buy forgiveness."
	case PardonTypeService:
		return "Work the fields to earn your place."
	case PardonTypePolitical:
		return "The leader vouches for you personally."
	case PardonTypeReligious:
		return "The elders perform the cleansing ritual."
	case PardonTypeMilitary:
		return "Defend the walls against raiders."
	default:
		return "Prove your worth to the community."
	}
}

// IsEligibleForAmnesty checks if an entity is eligible for a specific amnesty.
func (s *PardonSystem) IsEligibleForAmnesty(w *ecs.World, entity ecs.Entity, eventID string) (bool, string) {
	event := s.amnestyEvents[eventID]
	if event == nil {
		return false, "Amnesty event not found"
	}
	if !event.IsActive {
		return false, "Amnesty event has ended"
	}
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return false, "No criminal record"
	}
	crime := comp.(*components.Crime)
	if crime.WantedLevel > event.MaxWantedLevel {
		return false, "Wanted level too high for this amnesty"
	}
	if crime.WantedLevel <= 0 {
		return false, "No crimes to pardon"
	}
	// Check if already claimed
	for _, pid := range event.ParticipantIDs {
		if pid == uint64(entity) {
			return false, "Already claimed this amnesty"
		}
	}
	return true, ""
}
