// Package systems provides ECS systems for game logic.
package systems

import (
	"sync"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// PartyRole represents a player's role within a party.
type PartyRole string

const (
	PartyRoleLeader PartyRole = "leader"
	PartyRoleMember PartyRole = "member"
)

// PartyInviteStatus represents the status of a party invitation.
type PartyInviteStatus string

const (
	PartyInvitePending  PartyInviteStatus = "pending"
	PartyInviteAccepted PartyInviteStatus = "accepted"
	PartyInviteDeclined PartyInviteStatus = "declined"
	PartyInviteExpired  PartyInviteStatus = "expired"
)

// Party represents a group of players.
type Party struct {
	ID         string
	LeaderID   ecs.Entity
	Members    map[ecs.Entity]PartyRole
	MaxSize    int
	ShareXP    bool
	ShareLoot  bool
	Created    float64 // Game time created
	InviteOnly bool
}

// PartyInvite represents a pending party invitation.
type PartyInvite struct {
	ID          string
	PartyID     string
	InviterID   ecs.Entity
	InviteeID   ecs.Entity
	Status      PartyInviteStatus
	Created     float64
	ExpiresAt   float64
	InviteCount int // Allow up to 3 attempts
}

// PartySystem manages player parties for cooperative gameplay.
type PartySystem struct {
	mu           sync.RWMutex
	Parties      map[string]*Party
	Invites      map[string]*PartyInvite
	PlayerParty  map[ecs.Entity]string // Player -> Party ID mapping
	MaxPartySize int
	InviteExpiry float64 // How long invites last (seconds)
	GameTime     float64
	partyCounter uint64
	inviteCount  uint64
}

// NewPartySystem creates a new party management system.
func NewPartySystem() *PartySystem {
	return &PartySystem{
		Parties:      make(map[string]*Party),
		Invites:      make(map[string]*PartyInvite),
		PlayerParty:  make(map[ecs.Entity]string),
		MaxPartySize: 6,
		InviteExpiry: 60.0, // 60 seconds to accept
	}
}

// Update processes party system updates each tick.
func (s *PartySystem) Update(w *ecs.World, dt float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GameTime += dt
	s.expireInvites()
	s.cleanupEmptyParties()
}

// CreateParty creates a new party with the given player as leader.
func (s *PartySystem) CreateParty(leaderID ecs.Entity) (*Party, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.PlayerParty[leaderID]; ok {
		if p := s.Parties[existing]; p != nil {
			return nil, &PartyError{Code: ErrAlreadyInParty}
		}
	}

	s.partyCounter++
	partyID := generatePartyID(s.partyCounter)

	party := &Party{
		ID:         partyID,
		LeaderID:   leaderID,
		Members:    make(map[ecs.Entity]PartyRole),
		MaxSize:    s.MaxPartySize,
		ShareXP:    true,
		ShareLoot:  true,
		Created:    s.GameTime,
		InviteOnly: true,
	}
	party.Members[leaderID] = PartyRoleLeader

	s.Parties[partyID] = party
	s.PlayerParty[leaderID] = partyID

	return party, nil
}

// InviteToParty sends an invitation to join a party.
func (s *PartySystem) InviteToParty(inviterID, inviteeID ecs.Entity) (*PartyInvite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	partyID, ok := s.PlayerParty[inviterID]
	if !ok {
		return nil, &PartyError{Code: ErrNotInParty}
	}

	party := s.Parties[partyID]
	if party == nil {
		return nil, &PartyError{Code: ErrPartyNotFound}
	}

	if party.LeaderID != inviterID {
		return nil, &PartyError{Code: ErrNotPartyLeader}
	}

	if len(party.Members) >= party.MaxSize {
		return nil, &PartyError{Code: ErrPartyFull}
	}

	if _, inParty := s.PlayerParty[inviteeID]; inParty {
		return nil, &PartyError{Code: ErrTargetInParty}
	}

	s.inviteCount++
	inviteID := generateInviteID(s.inviteCount)

	invite := &PartyInvite{
		ID:        inviteID,
		PartyID:   partyID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		Status:    PartyInvitePending,
		Created:   s.GameTime,
		ExpiresAt: s.GameTime + s.InviteExpiry,
	}

	s.Invites[inviteID] = invite
	return invite, nil
}

// AcceptInvite accepts a party invitation.
func (s *PartySystem) AcceptInvite(inviteID string, playerID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	invite := s.Invites[inviteID]
	if invite == nil {
		return &PartyError{Code: ErrInviteNotFound}
	}

	if invite.InviteeID != playerID {
		return &PartyError{Code: ErrWrongInvitee}
	}

	if invite.Status != PartyInvitePending {
		return &PartyError{Code: ErrInviteNotPending}
	}

	if s.GameTime > invite.ExpiresAt {
		invite.Status = PartyInviteExpired
		return &PartyError{Code: ErrInviteExpired}
	}

	party := s.Parties[invite.PartyID]
	if party == nil {
		return &PartyError{Code: ErrPartyNotFound}
	}

	if len(party.Members) >= party.MaxSize {
		return &PartyError{Code: ErrPartyFull}
	}

	party.Members[playerID] = PartyRoleMember
	s.PlayerParty[playerID] = party.ID
	invite.Status = PartyInviteAccepted

	return nil
}

// DeclineInvite declines a party invitation.
func (s *PartySystem) DeclineInvite(inviteID string, playerID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	invite := s.Invites[inviteID]
	if invite == nil {
		return &PartyError{Code: ErrInviteNotFound}
	}

	if invite.InviteeID != playerID {
		return &PartyError{Code: ErrWrongInvitee}
	}

	invite.Status = PartyInviteDeclined
	return nil
}

// LeaveParty removes a player from their current party.
func (s *PartySystem) LeaveParty(playerID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	partyID, ok := s.PlayerParty[playerID]
	if !ok {
		return &PartyError{Code: ErrNotInParty}
	}

	party := s.Parties[partyID]
	if party == nil {
		delete(s.PlayerParty, playerID)
		return nil
	}

	delete(party.Members, playerID)
	delete(s.PlayerParty, playerID)

	// If leader left, promote someone else or disband
	if party.LeaderID == playerID {
		s.handleLeaderLeft(party)
	}

	return nil
}

// KickFromParty removes a player from the party (leader only).
func (s *PartySystem) KickFromParty(leaderID, targetID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	partyID, ok := s.PlayerParty[leaderID]
	if !ok {
		return &PartyError{Code: ErrNotInParty}
	}

	party := s.Parties[partyID]
	if party == nil {
		return &PartyError{Code: ErrPartyNotFound}
	}

	if party.LeaderID != leaderID {
		return &PartyError{Code: ErrNotPartyLeader}
	}

	if leaderID == targetID {
		return &PartyError{Code: ErrCannotKickSelf}
	}

	if _, isMember := party.Members[targetID]; !isMember {
		return &PartyError{Code: ErrTargetNotInParty}
	}

	delete(party.Members, targetID)
	delete(s.PlayerParty, targetID)

	return nil
}

// PromoteToLeader promotes a member to party leader.
func (s *PartySystem) PromoteToLeader(currentLeaderID, newLeaderID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	partyID, ok := s.PlayerParty[currentLeaderID]
	if !ok {
		return &PartyError{Code: ErrNotInParty}
	}

	party := s.Parties[partyID]
	if party == nil {
		return &PartyError{Code: ErrPartyNotFound}
	}

	if party.LeaderID != currentLeaderID {
		return &PartyError{Code: ErrNotPartyLeader}
	}

	if _, isMember := party.Members[newLeaderID]; !isMember {
		return &PartyError{Code: ErrTargetNotInParty}
	}

	party.Members[currentLeaderID] = PartyRoleMember
	party.Members[newLeaderID] = PartyRoleLeader
	party.LeaderID = newLeaderID

	return nil
}

// DisbandParty disbands the party (leader only).
func (s *PartySystem) DisbandParty(leaderID ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	partyID, ok := s.PlayerParty[leaderID]
	if !ok {
		return &PartyError{Code: ErrNotInParty}
	}

	party := s.Parties[partyID]
	if party == nil {
		return &PartyError{Code: ErrPartyNotFound}
	}

	if party.LeaderID != leaderID {
		return &PartyError{Code: ErrNotPartyLeader}
	}

	for memberID := range party.Members {
		delete(s.PlayerParty, memberID)
	}
	delete(s.Parties, partyID)

	return nil
}

// GetParty returns the party a player belongs to.
func (s *PartySystem) GetParty(playerID ecs.Entity) *Party {
	s.mu.RLock()
	defer s.mu.RUnlock()

	partyID, ok := s.PlayerParty[playerID]
	if !ok {
		return nil
	}
	return s.Parties[partyID]
}

// GetPartyMembers returns all members of a player's party.
func (s *PartySystem) GetPartyMembers(playerID ecs.Entity) []ecs.Entity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	partyID, ok := s.PlayerParty[playerID]
	if !ok {
		return nil
	}

	party := s.Parties[partyID]
	if party == nil {
		return nil
	}

	members := make([]ecs.Entity, 0, len(party.Members))
	for memberID := range party.Members {
		members = append(members, memberID)
	}
	return members
}

// IsPartyLeader checks if a player is the leader of their party.
func (s *PartySystem) IsPartyLeader(playerID ecs.Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	partyID, ok := s.PlayerParty[playerID]
	if !ok {
		return false
	}

	party := s.Parties[partyID]
	return party != nil && party.LeaderID == playerID
}

// IsInSameParty checks if two players are in the same party.
func (s *PartySystem) IsInSameParty(playerA, playerB ecs.Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	partyA := s.PlayerParty[playerA]
	partyB := s.PlayerParty[playerB]
	return partyA != "" && partyA == partyB
}

// GetPendingInvites returns all pending invites for a player.
func (s *PartySystem) GetPendingInvites(playerID ecs.Entity) []*PartyInvite {
	s.mu.RLock()
	defer s.mu.RUnlock()

	invites := make([]*PartyInvite, 0)
	for _, invite := range s.Invites {
		if invite.InviteeID == playerID && invite.Status == PartyInvitePending {
			invites = append(invites, invite)
		}
	}
	return invites
}

// SetLootSharing enables or disables loot sharing for a party.
func (s *PartySystem) SetLootSharing(leaderID ecs.Entity, share bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	partyID, ok := s.PlayerParty[leaderID]
	if !ok {
		return &PartyError{Code: ErrNotInParty}
	}

	party := s.Parties[partyID]
	if party == nil {
		return &PartyError{Code: ErrPartyNotFound}
	}

	if party.LeaderID != leaderID {
		return &PartyError{Code: ErrNotPartyLeader}
	}

	party.ShareLoot = share
	return nil
}

// SetXPSharing enables or disables XP sharing for a party.
func (s *PartySystem) SetXPSharing(leaderID ecs.Entity, share bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	partyID, ok := s.PlayerParty[leaderID]
	if !ok {
		return &PartyError{Code: ErrNotInParty}
	}

	party := s.Parties[partyID]
	if party == nil {
		return &PartyError{Code: ErrPartyNotFound}
	}

	if party.LeaderID != leaderID {
		return &PartyError{Code: ErrNotPartyLeader}
	}

	party.ShareXP = share
	return nil
}

// expireInvites marks expired invites.
func (s *PartySystem) expireInvites() {
	for _, invite := range s.Invites {
		if invite.Status == PartyInvitePending && s.GameTime > invite.ExpiresAt {
			invite.Status = PartyInviteExpired
		}
	}
}

// cleanupEmptyParties removes parties with no members.
func (s *PartySystem) cleanupEmptyParties() {
	for partyID, party := range s.Parties {
		if len(party.Members) == 0 {
			delete(s.Parties, partyID)
		}
	}
}

// handleLeaderLeft handles leadership transfer when leader leaves.
func (s *PartySystem) handleLeaderLeft(party *Party) {
	if len(party.Members) == 0 {
		return
	}
	// Promote first available member
	for memberID := range party.Members {
		party.LeaderID = memberID
		party.Members[memberID] = PartyRoleLeader
		return
	}
}

// generatePartyID creates a unique party ID.
func generatePartyID(counter uint64) string {
	return "party_" + uint64ToString(counter)
}

// generateInviteID creates a unique invite ID.
func generateInviteID(counter uint64) string {
	return "invite_" + uint64ToString(counter)
}

// uint64ToString converts uint64 to string without fmt.
func uint64ToString(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// PartyErrorCode represents specific party errors.
type PartyErrorCode int

const (
	ErrAlreadyInParty PartyErrorCode = iota
	ErrNotInParty
	ErrPartyNotFound
	ErrNotPartyLeader
	ErrPartyFull
	ErrTargetInParty
	ErrTargetNotInParty
	ErrInviteNotFound
	ErrWrongInvitee
	ErrInviteNotPending
	ErrInviteExpired
	ErrCannotKickSelf
)

// PartyError represents a party-related error.
type PartyError struct {
	Code PartyErrorCode
}

// Error returns a human-readable message for the PartyError.
func (e *PartyError) Error() string {
	switch e.Code {
	case ErrAlreadyInParty:
		return "player is already in a party"
	case ErrNotInParty:
		return "player is not in a party"
	case ErrPartyNotFound:
		return "party not found"
	case ErrNotPartyLeader:
		return "not the party leader"
	case ErrPartyFull:
		return "party is full"
	case ErrTargetInParty:
		return "target player is already in a party"
	case ErrTargetNotInParty:
		return "target player is not in this party"
	case ErrInviteNotFound:
		return "invite not found"
	case ErrWrongInvitee:
		return "not the intended invite recipient"
	case ErrInviteNotPending:
		return "invite is no longer pending"
	case ErrInviteExpired:
		return "invite has expired"
	case ErrCannotKickSelf:
		return "cannot kick yourself from the party"
	default:
		return "party error"
	}
}
