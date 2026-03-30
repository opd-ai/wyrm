package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewPartySystem(t *testing.T) {
	ps := NewPartySystem()
	if ps == nil {
		t.Fatal("NewPartySystem returned nil")
	}
	if ps.MaxPartySize != 6 {
		t.Errorf("MaxPartySize = %d, want 6", ps.MaxPartySize)
	}
	if ps.InviteExpiry != 60.0 {
		t.Errorf("InviteExpiry = %f, want 60.0", ps.InviteExpiry)
	}
}

func TestPartySystem_CreateParty(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)

	party, err := ps.CreateParty(leaderID)
	if err != nil {
		t.Fatalf("CreateParty failed: %v", err)
	}
	if party == nil {
		t.Fatal("CreateParty returned nil party")
	}
	if party.LeaderID != leaderID {
		t.Errorf("LeaderID = %d, want %d", party.LeaderID, leaderID)
	}
	if len(party.Members) != 1 {
		t.Errorf("Members count = %d, want 1", len(party.Members))
	}
	if party.Members[leaderID] != PartyRoleLeader {
		t.Error("Leader should have PartyRoleLeader role")
	}
}

func TestPartySystem_CreateParty_AlreadyInParty(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)

	_, err := ps.CreateParty(leaderID)
	if err != nil {
		t.Fatalf("First CreateParty failed: %v", err)
	}

	_, err = ps.CreateParty(leaderID)
	if err == nil {
		t.Error("Expected error when creating party while already in one")
	}
	if pe, ok := err.(*PartyError); !ok || pe.Code != ErrAlreadyInParty {
		t.Errorf("Expected ErrAlreadyInParty, got %v", err)
	}
}

func TestPartySystem_InviteToParty(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	inviteeID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)

	invite, err := ps.InviteToParty(leaderID, inviteeID)
	if err != nil {
		t.Fatalf("InviteToParty failed: %v", err)
	}
	if invite == nil {
		t.Fatal("InviteToParty returned nil invite")
	}
	if invite.InviteeID != inviteeID {
		t.Errorf("InviteeID = %d, want %d", invite.InviteeID, inviteeID)
	}
	if invite.Status != PartyInvitePending {
		t.Errorf("Status = %s, want pending", invite.Status)
	}
}

func TestPartySystem_InviteToParty_NotInParty(t *testing.T) {
	ps := NewPartySystem()
	_, err := ps.InviteToParty(ecs.Entity(1), ecs.Entity(2))
	if err == nil {
		t.Error("Expected error when inviting without being in party")
	}
}

func TestPartySystem_InviteToParty_NotLeader(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	memberID := ecs.Entity(2)
	inviteeID := ecs.Entity(3)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, memberID)
	_ = ps.AcceptInvite(invite.ID, memberID)

	// Member tries to invite
	_, err := ps.InviteToParty(memberID, inviteeID)
	if err == nil {
		t.Error("Expected error when non-leader invites")
	}
}

func TestPartySystem_AcceptInvite(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	inviteeID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, inviteeID)

	err := ps.AcceptInvite(invite.ID, inviteeID)
	if err != nil {
		t.Fatalf("AcceptInvite failed: %v", err)
	}

	party := ps.GetParty(inviteeID)
	if party == nil {
		t.Fatal("Invitee should be in party after accepting")
	}
	if party.Members[inviteeID] != PartyRoleMember {
		t.Error("Invitee should have member role")
	}
}

func TestPartySystem_AcceptInvite_WrongPlayer(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	inviteeID := ecs.Entity(2)
	wrongID := ecs.Entity(3)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, inviteeID)

	err := ps.AcceptInvite(invite.ID, wrongID)
	if err == nil {
		t.Error("Expected error when wrong player accepts invite")
	}
}

func TestPartySystem_DeclineInvite(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	inviteeID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, inviteeID)

	err := ps.DeclineInvite(invite.ID, inviteeID)
	if err != nil {
		t.Fatalf("DeclineInvite failed: %v", err)
	}
	if invite.Status != PartyInviteDeclined {
		t.Errorf("Status = %s, want declined", invite.Status)
	}
}

func TestPartySystem_LeaveParty(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	memberID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, memberID)
	_ = ps.AcceptInvite(invite.ID, memberID)

	err := ps.LeaveParty(memberID)
	if err != nil {
		t.Fatalf("LeaveParty failed: %v", err)
	}

	party := ps.GetParty(memberID)
	if party != nil {
		t.Error("Member should not be in party after leaving")
	}
}

func TestPartySystem_LeaveParty_LeaderPromotion(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	memberID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, memberID)
	_ = ps.AcceptInvite(invite.ID, memberID)

	err := ps.LeaveParty(leaderID)
	if err != nil {
		t.Fatalf("LeaveParty (leader) failed: %v", err)
	}

	party := ps.GetParty(memberID)
	if party == nil {
		t.Fatal("Party should still exist")
	}
	if party.LeaderID != memberID {
		t.Error("Remaining member should become leader")
	}
}

func TestPartySystem_KickFromParty(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	memberID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, memberID)
	_ = ps.AcceptInvite(invite.ID, memberID)

	err := ps.KickFromParty(leaderID, memberID)
	if err != nil {
		t.Fatalf("KickFromParty failed: %v", err)
	}

	party := ps.GetParty(memberID)
	if party != nil {
		t.Error("Kicked member should not be in party")
	}
}

func TestPartySystem_KickFromParty_NotLeader(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	member1 := ecs.Entity(2)
	member2 := ecs.Entity(3)

	_, _ = ps.CreateParty(leaderID)
	invite1, _ := ps.InviteToParty(leaderID, member1)
	_ = ps.AcceptInvite(invite1.ID, member1)
	invite2, _ := ps.InviteToParty(leaderID, member2)
	_ = ps.AcceptInvite(invite2.ID, member2)

	// Member tries to kick
	err := ps.KickFromParty(member1, member2)
	if err == nil {
		t.Error("Expected error when non-leader kicks")
	}
}

func TestPartySystem_PromoteToLeader(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	memberID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, memberID)
	_ = ps.AcceptInvite(invite.ID, memberID)

	err := ps.PromoteToLeader(leaderID, memberID)
	if err != nil {
		t.Fatalf("PromoteToLeader failed: %v", err)
	}

	if !ps.IsPartyLeader(memberID) {
		t.Error("Promoted member should be leader")
	}
	if ps.IsPartyLeader(leaderID) {
		t.Error("Old leader should no longer be leader")
	}
}

func TestPartySystem_DisbandParty(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	memberID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, memberID)
	_ = ps.AcceptInvite(invite.ID, memberID)

	err := ps.DisbandParty(leaderID)
	if err != nil {
		t.Fatalf("DisbandParty failed: %v", err)
	}

	if ps.GetParty(leaderID) != nil {
		t.Error("Leader should not be in party after disband")
	}
	if ps.GetParty(memberID) != nil {
		t.Error("Member should not be in party after disband")
	}
}

func TestPartySystem_GetPartyMembers(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	memberID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, memberID)
	_ = ps.AcceptInvite(invite.ID, memberID)

	members := ps.GetPartyMembers(leaderID)
	if len(members) != 2 {
		t.Errorf("Member count = %d, want 2", len(members))
	}
}

func TestPartySystem_IsInSameParty(t *testing.T) {
	ps := NewPartySystem()
	player1 := ecs.Entity(1)
	player2 := ecs.Entity(2)
	player3 := ecs.Entity(3)

	_, _ = ps.CreateParty(player1)
	invite, _ := ps.InviteToParty(player1, player2)
	_ = ps.AcceptInvite(invite.ID, player2)

	if !ps.IsInSameParty(player1, player2) {
		t.Error("Players in same party should return true")
	}
	if ps.IsInSameParty(player1, player3) {
		t.Error("Players in different/no party should return false")
	}
}

func TestPartySystem_GetPendingInvites(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	inviteeID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	_, _ = ps.InviteToParty(leaderID, inviteeID)

	invites := ps.GetPendingInvites(inviteeID)
	if len(invites) != 1 {
		t.Errorf("Pending invites = %d, want 1", len(invites))
	}
}

func TestPartySystem_SetLootSharing(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)

	party, _ := ps.CreateParty(leaderID)
	if !party.ShareLoot {
		t.Error("Loot sharing should be enabled by default")
	}

	_ = ps.SetLootSharing(leaderID, false)
	if party.ShareLoot {
		t.Error("Loot sharing should be disabled")
	}
}

func TestPartySystem_SetXPSharing(t *testing.T) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)

	party, _ := ps.CreateParty(leaderID)
	if !party.ShareXP {
		t.Error("XP sharing should be enabled by default")
	}

	_ = ps.SetXPSharing(leaderID, false)
	if party.ShareXP {
		t.Error("XP sharing should be disabled")
	}
}

func TestPartySystem_Update_ExpireInvites(t *testing.T) {
	ps := NewPartySystem()
	ps.InviteExpiry = 10.0 // Short expiry for test

	leaderID := ecs.Entity(1)
	inviteeID := ecs.Entity(2)

	_, _ = ps.CreateParty(leaderID)
	invite, _ := ps.InviteToParty(leaderID, inviteeID)

	// Advance time past expiry
	w := ecs.NewWorld()
	ps.Update(w, 15.0)

	if invite.Status != PartyInviteExpired {
		t.Errorf("Invite status = %s, want expired", invite.Status)
	}
}

func TestPartySystem_PartyFull(t *testing.T) {
	ps := NewPartySystem()
	ps.MaxPartySize = 2 // Small party for test

	leaderID := ecs.Entity(1)
	member1 := ecs.Entity(2)
	member2 := ecs.Entity(3)

	_, _ = ps.CreateParty(leaderID)
	invite1, _ := ps.InviteToParty(leaderID, member1)
	_ = ps.AcceptInvite(invite1.ID, member1)

	// Party is now full (2 members)
	_, err := ps.InviteToParty(leaderID, member2)
	if err == nil {
		t.Error("Expected error when inviting to full party")
	}
	if pe, ok := err.(*PartyError); !ok || pe.Code != ErrPartyFull {
		t.Errorf("Expected ErrPartyFull, got %v", err)
	}
}

func TestPartyError_Error(t *testing.T) {
	tests := []struct {
		code PartyErrorCode
		want string
	}{
		{ErrAlreadyInParty, "player is already in a party"},
		{ErrNotInParty, "player is not in a party"},
		{ErrPartyNotFound, "party not found"},
		{ErrNotPartyLeader, "not the party leader"},
		{ErrPartyFull, "party is full"},
		{ErrTargetInParty, "target player is already in a party"},
		{ErrTargetNotInParty, "target player is not in this party"},
		{ErrInviteNotFound, "invite not found"},
		{ErrWrongInvitee, "not the intended invite recipient"},
		{ErrInviteNotPending, "invite is no longer pending"},
		{ErrInviteExpired, "invite has expired"},
		{ErrCannotKickSelf, "cannot kick yourself from the party"},
		{PartyErrorCode(999), "party error"}, // Unknown code
	}

	for _, tc := range tests {
		err := &PartyError{Code: tc.code}
		if err.Error() != tc.want {
			t.Errorf("Error() = %q, want %q", err.Error(), tc.want)
		}
	}
}

func BenchmarkPartySystem_CreateParty(b *testing.B) {
	ps := NewPartySystem()
	for i := 0; i < b.N; i++ {
		ps.partyCounter = 0 // Reset
		ps.Parties = make(map[string]*Party)
		ps.PlayerParty = make(map[ecs.Entity]string)
		_, _ = ps.CreateParty(ecs.Entity(i))
	}
}

func BenchmarkPartySystem_InviteAndAccept(b *testing.B) {
	ps := NewPartySystem()
	leaderID := ecs.Entity(1)
	_, _ = ps.CreateParty(leaderID)

	for i := 0; i < b.N; i++ {
		inviteeID := ecs.Entity(i + 100)
		invite, _ := ps.InviteToParty(leaderID, inviteeID)
		_ = ps.AcceptInvite(invite.ID, inviteeID)
		// Clean up for next iteration
		delete(ps.PlayerParty, inviteeID)
		party := ps.Parties[ps.PlayerParty[leaderID]]
		if party != nil {
			delete(party.Members, inviteeID)
		}
	}
}
