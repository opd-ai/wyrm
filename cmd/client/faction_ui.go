//go:build !noebiten

// Package main provides the faction rank UI overlay for the Wyrm client.
package main

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// FactionUI manages the faction standings overlay display.
type FactionUI struct {
	isOpen          bool
	playerEntity    ecs.Entity
	selectedFaction int
	scrollOffset    int
	genre           string
}

// NewFactionUI creates a new faction UI.
func NewFactionUI(genre string, playerEntity ecs.Entity) *FactionUI {
	return &FactionUI{
		isOpen:          false,
		playerEntity:    playerEntity,
		selectedFaction: 0,
		scrollOffset:    0,
		genre:           genre,
	}
}

// IsOpen returns whether the faction UI is currently open.
func (ui *FactionUI) IsOpen() bool {
	return ui.isOpen
}

// Open opens the faction UI.
func (ui *FactionUI) Open() {
	ui.isOpen = true
	ui.selectedFaction = 0
	ui.scrollOffset = 0
}

// Close closes the faction UI.
func (ui *FactionUI) Close() {
	ui.isOpen = false
}

// Toggle toggles the faction UI open/closed state.
func (ui *FactionUI) Toggle() {
	if ui.isOpen {
		ui.Close()
	} else {
		ui.Open()
	}
}

// Update handles input for the faction UI.
func (ui *FactionUI) Update(world *ecs.World) {
	if !ui.isOpen {
		return
	}

	// Check for cancel/close
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.Close()
		return
	}

	// Navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		ui.selectedFaction--
		if ui.selectedFaction < 0 {
			ui.selectedFaction = 0
		}
		ui.adjustScroll()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		ui.selectedFaction++
		ui.adjustScroll()
	}
}

// adjustScroll adjusts scroll offset to keep selected faction visible.
func (ui *FactionUI) adjustScroll() {
	visibleRows := 6
	if ui.selectedFaction < ui.scrollOffset {
		ui.scrollOffset = ui.selectedFaction
	}
	if ui.selectedFaction >= ui.scrollOffset+visibleRows {
		ui.scrollOffset = ui.selectedFaction - visibleRows + 1
	}
}

// FactionDisplayInfo holds faction display data.
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

// getFactionInfo retrieves all faction information for the player.
func (ui *FactionUI) getFactionInfo(world *ecs.World) []FactionDisplayInfo {
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
			Name:            ui.getFactionDisplayName(factionID),
			Reputation:      info.Reputation,
			Rank:            info.Rank,
			RankTitle:       info.RankTitle,
			XP:              info.XP,
			XPToNext:        info.XPToNext,
			IsMember:        info.Rank > 0,
			IsExalted:       info.IsExalted,
			QuestsCompleted: info.QuestsCompleted,
			StandingLevel:   ui.getStandingLevel(info.Reputation),
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
func (ui *FactionUI) getFactionDisplayName(factionID string) string {
	// Convert faction ID to display name (e.g., "thieves_guild" -> "Thieves Guild")
	name := ""
	capitalize := true
	for _, c := range factionID {
		if c == '_' {
			name += " "
			capitalize = true
		} else {
			if capitalize {
				name += string(c - 32) // Convert to uppercase if lowercase
				capitalize = false
			} else {
				name += string(c)
			}
		}
	}
	return name
}

// getStandingLevel returns a textual standing level based on reputation.
func (ui *FactionUI) getStandingLevel(reputation float64) string {
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

// getStandingColor returns a color based on standing level.
func (ui *FactionUI) getStandingColor(standing string) color.RGBA {
	switch standing {
	case "Exalted":
		return color.RGBA{255, 215, 0, 255} // Gold
	case "Allied":
		return color.RGBA{0, 255, 0, 255} // Green
	case "Friendly":
		return color.RGBA{144, 238, 144, 255} // Light green
	case "Neutral":
		return color.RGBA{200, 200, 200, 255} // Gray
	case "Unfriendly":
		return color.RGBA{255, 165, 0, 255} // Orange
	case "Hostile":
		return color.RGBA{255, 0, 0, 255} // Red
	default:
		return color.RGBA{200, 200, 200, 255}
	}
}

// Draw renders the faction UI.
func (ui *FactionUI) Draw(screen *ebiten.Image, world *ecs.World) {
	if !ui.isOpen {
		return
	}

	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()
	panelW, panelH := 450, 450
	panelX := (screenW - panelW) / 2
	panelY := (screenH - panelH) / 2

	ui.drawPanelBackground(screen, panelX, panelY, panelW, panelH)

	factions := ui.getFactionInfo(world)
	ui.clampSelectedFaction(len(factions))

	ui.drawFactionList(screen, factions, panelX, panelY, panelW)
	ui.drawFactionDetails(screen, factions, panelX, panelY, panelW, panelH)
	ui.drawHelpText(screen, panelX, panelY, panelH)
}

// drawPanelBackground draws the panel background and border.
func (ui *FactionUI) drawPanelBackground(screen *ebiten.Image, panelX, panelY, panelW, panelH int) {
	bgColor := color.RGBA{0, 0, 0, 200}
	ebitenutil.DrawRect(screen, float64(panelX), float64(panelY), float64(panelW), float64(panelH), bgColor)

	borderColor := ui.getGenreColor()
	ebitenutil.DrawRect(screen, float64(panelX), float64(panelY), float64(panelW), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(panelX), float64(panelY+panelH-2), float64(panelW), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(panelX), float64(panelY), 2, float64(panelH), borderColor)
	ebitenutil.DrawRect(screen, float64(panelX+panelW-2), float64(panelY), 2, float64(panelH), borderColor)

	ebitenutil.DebugPrintAt(screen, "=== Faction Standings ===", panelX+140, panelY+10)
}

// clampSelectedFaction ensures selected faction is within valid range.
func (ui *FactionUI) clampSelectedFaction(count int) {
	if ui.selectedFaction >= count {
		ui.selectedFaction = count - 1
	}
	if ui.selectedFaction < 0 {
		ui.selectedFaction = 0
	}
}

// drawFactionList draws the scrollable faction list.
func (ui *FactionUI) drawFactionList(screen *ebiten.Image, factions []FactionDisplayInfo, panelX, panelY, panelW int) {
	listY := panelY + 35
	visibleRows := 6
	rowHeight := 50

	for i := ui.scrollOffset; i < len(factions) && i < ui.scrollOffset+visibleRows; i++ {
		y := listY + (i-ui.scrollOffset)*rowHeight
		ui.drawFactionRow(screen, factions[i], i, panelX, y, panelW, rowHeight)
	}
}

// drawFactionRow draws a single faction row.
func (ui *FactionUI) drawFactionRow(screen *ebiten.Image, faction FactionDisplayInfo, index, panelX, y, panelW, rowHeight int) {
	if index == ui.selectedFaction {
		ebitenutil.DrawRect(screen, float64(panelX+5), float64(y-2), float64(panelW-10), float64(rowHeight-4), color.RGBA{60, 60, 80, 200})
	}

	standingColor := ui.getStandingColor(faction.StandingLevel)
	ebitenutil.DrawRect(screen, float64(panelX+10), float64(y+5), 10, 10, standingColor)
	ebitenutil.DebugPrintAt(screen, faction.Name, panelX+25, y+3)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[%s]", faction.StandingLevel), panelX+200, y+3)

	ui.drawReputationBar(screen, faction, standingColor, panelX, y+20, panelW)
	ui.drawRankInfo(screen, faction, panelX, y+35)
}

// drawReputationBar draws the reputation progress bar.
func (ui *FactionUI) drawReputationBar(screen *ebiten.Image, faction FactionDisplayInfo, standingColor color.RGBA, barX, barY, panelW int) {
	barW := panelW - 20
	barH := 12

	ebitenutil.DrawRect(screen, float64(barX+10), float64(barY), float64(barW), float64(barH), color.RGBA{40, 40, 40, 255})

	reputationNorm := (faction.Reputation + 100) / 200
	fillW := int(float64(barW) * reputationNorm)
	ebitenutil.DrawRect(screen, float64(barX+10), float64(barY), float64(fillW), float64(barH), standingColor)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rep: %.0f", faction.Reputation), barX+barW-50, barY-1)
}

// drawRankInfo draws member rank information if applicable.
func (ui *FactionUI) drawRankInfo(screen *ebiten.Image, faction FactionDisplayInfo, panelX, y int) {
	if !faction.IsMember {
		return
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rank %d: %s", faction.Rank, faction.RankTitle), panelX+10, y)

	if faction.XPToNext > 0 && !faction.IsExalted {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("XP: %d/%d", faction.XP, faction.XPToNext), panelX+200, y)
	} else if faction.IsExalted {
		ebitenutil.DebugPrintAt(screen, "MAX RANK", panelX+200, y)
	}
}

// drawFactionDetails draws the selected faction details section.
func (ui *FactionUI) drawFactionDetails(screen *ebiten.Image, factions []FactionDisplayInfo, panelX, panelY, panelW, panelH int) {
	if ui.selectedFaction < 0 || ui.selectedFaction >= len(factions) {
		return
	}
	faction := factions[ui.selectedFaction]
	detailY := panelY + panelH - 100
	borderColor := ui.getGenreColor()

	ebitenutil.DrawRect(screen, float64(panelX+10), float64(detailY-10), float64(panelW-20), 1, borderColor)
	ebitenutil.DebugPrintAt(screen, "Details:", panelX+10, detailY)

	if faction.IsMember {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Quests Completed: %d", faction.QuestsCompleted), panelX+10, detailY+15)
	} else {
		ebitenutil.DebugPrintAt(screen, "Not a member - complete quests to join", panelX+10, detailY+15)
	}

	ebitenutil.DebugPrintAt(screen, ui.getAvailableActions(faction), panelX+10, detailY+30)
}

// drawHelpText draws the help text at the bottom.
func (ui *FactionUI) drawHelpText(screen *ebiten.Image, panelX, panelY, panelH int) {
	ebitenutil.DebugPrintAt(screen, "[UP/DOWN] Select  [ESC] Close", panelX+10, panelY+panelH-20)
}

// getAvailableActions returns actions available based on faction standing.
func (ui *FactionUI) getAvailableActions(faction FactionDisplayInfo) string {
	switch faction.StandingLevel {
	case "Exalted":
		return "Actions: Exclusive quests, special vendor access"
	case "Allied":
		return "Actions: Faction quests, vendor discounts"
	case "Friendly":
		return "Actions: Can accept basic faction quests"
	case "Neutral":
		return "Actions: Can interact with faction NPCs"
	case "Unfriendly":
		return "Actions: Limited - faction members avoid you"
	case "Hostile":
		return "Actions: HOSTILE - faction members will attack!"
	default:
		return ""
	}
}

// getGenreColor returns the accent color for the current genre.
func (ui *FactionUI) getGenreColor() color.RGBA {
	switch ui.genre {
	case "fantasy":
		return color.RGBA{218, 165, 32, 255} // Gold
	case "sci-fi":
		return color.RGBA{0, 191, 255, 255} // Deep sky blue
	case "horror":
		return color.RGBA{139, 0, 0, 255} // Dark red
	case "cyberpunk":
		return color.RGBA{255, 0, 255, 255} // Magenta
	case "post-apocalyptic":
		return color.RGBA{210, 105, 30, 255} // Chocolate
	default:
		return color.RGBA{100, 100, 100, 255}
	}
}
