// Package main provides shared faction UI logic for all build configurations.
package main

import (
	"sort"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// FactionDisplayInfo holds formatted faction info for UI display.
type FactionDisplayInfo struct {
	ID              string
	Name            string
	Reputation      float64
	Rank            int
	RankTitle       string
	XP              int
	XPToNext        int
	IsMember        bool
	IsExalted       bool
	QuestsCompleted int
	StandingLevel   string // "Hostile", "Unfriendly", "Neutral", "Friendly", "Allied", "Exalted"
}

// factionUIBase provides shared implementation for faction UI across builds.
type factionUIBase struct {
	playerEntity ecs.Entity
}

// getFactionInfo retrieves all faction information for the player.
func (ui *factionUIBase) getFactionInfo(world *ecs.World) []FactionDisplayInfo {
	membershipComp, ok := world.GetComponent(ui.playerEntity, "FactionMembership")
	if !ok {
		return nil
	}
	membership := membershipComp.(*components.FactionMembership)

	var factions []FactionDisplayInfo

	if membership.Memberships == nil {
		return factions
	}

	for factionID, info := range membership.Memberships {
		fdi := FactionDisplayInfo{
			ID:              factionID,
			Name:            getFactionDisplayName(factionID),
			Reputation:      info.Reputation,
			Rank:            info.Rank,
			RankTitle:       info.RankTitle,
			XP:              info.XP,
			XPToNext:        info.XPToNext,
			IsMember:        info.Rank > 0,
			IsExalted:       info.IsExalted,
			QuestsCompleted: info.QuestsCompleted,
			StandingLevel:   getStandingLevel(info.Reputation),
		}
		factions = append(factions, fdi)
	}

	// Sort by reputation descending
	sort.Slice(factions, func(i, j int) bool {
		return factions[i].Reputation > factions[j].Reputation
	})

	return factions
}

// getFactionDisplayName returns a formatted faction name.
func getFactionDisplayName(factionID string) string {
	name := ""
	capitalize := true
	for _, c := range factionID {
		if c == '_' {
			name += " "
			capitalize = true
			continue
		}
		if capitalize {
			if c >= 'a' && c <= 'z' {
				c -= 32 // uppercase
			}
			capitalize = false
		}
		name += string(c)
	}
	return name
}

// getStandingLevel returns the standing level string for a reputation value.
func getStandingLevel(reputation float64) string {
	switch {
	case reputation >= 80:
		return "Exalted"
	case reputation >= 50:
		return "Allied"
	case reputation >= 20:
		return "Friendly"
	case reputation >= -20:
		return "Neutral"
	case reputation >= -50:
		return "Unfriendly"
	default:
		return "Hostile"
	}
}
