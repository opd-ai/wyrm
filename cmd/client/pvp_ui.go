//go:build !noebiten

// Package main provides the Wyrm game client with PvP zone indicators.
package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/world/pvp"
)

// PvPUI manages PvP zone indicators and flagging.
type PvPUI struct {
	zoneManager    *pvp.ZoneManager
	showIndicator  bool
	indicatorFlash int
	lastZoneType   pvp.ZoneType
	rng            *rand.Rand

	// Flag toggle state
	wantsFlagged bool

	// Loot drop display
	recentDrop   *pvp.DeathLoot
	dropShowTime time.Time
}

// NewPvPUI creates a new PvP UI manager.
func NewPvPUI() *PvPUI {
	ui := &PvPUI{
		zoneManager:   pvp.NewZoneManager(),
		showIndicator: true,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	ui.initializeZones()
	return ui
}

// initializeZones sets up default PvP zones for the world.
func (ui *PvPUI) initializeZones() {
	// City center - safe zone
	ui.zoneManager.AddZone(&pvp.Zone{
		ID:           "city-center",
		Type:         pvp.ZoneSafe,
		MinX:         -50,
		MinZ:         -50,
		MaxX:         50,
		MaxZ:         50,
		RespawnX:     0,
		RespawnZ:     0,
		LootDropRate: 0,
	})

	// Contested outskirts
	ui.zoneManager.AddZone(&pvp.Zone{
		ID:           "outskirts-north",
		Type:         pvp.ZoneContested,
		MinX:         -100,
		MinZ:         50,
		MaxX:         100,
		MaxZ:         150,
		RespawnX:     0,
		RespawnZ:     45,
		LootDropRate: 0.1,
	})

	ui.zoneManager.AddZone(&pvp.Zone{
		ID:           "outskirts-south",
		Type:         pvp.ZoneContested,
		MinX:         -100,
		MinZ:         -150,
		MaxX:         100,
		MaxZ:         -50,
		RespawnX:     0,
		RespawnZ:     -45,
		LootDropRate: 0.1,
	})

	// Hostile wilderness
	ui.zoneManager.AddZone(&pvp.Zone{
		ID:           "wilderness-far",
		Type:         pvp.ZoneHostile,
		MinX:         -500,
		MinZ:         -500,
		MaxX:         500,
		MaxZ:         -150,
		RespawnX:     0,
		RespawnZ:     -45,
		LootDropRate: 0.25,
	})

	ui.zoneManager.AddZone(&pvp.Zone{
		ID:           "wilderness-north",
		Type:         pvp.ZoneHostile,
		MinX:         -500,
		MinZ:         150,
		MaxX:         500,
		MaxZ:         500,
		RespawnX:     0,
		RespawnZ:     45,
		LootDropRate: 0.25,
	})
}

// Update handles PvP UI input and state.
func (ui *PvPUI) Update(playerEntity uint64, playerX, playerZ float64) {
	// Update flash animation
	ui.indicatorFlash++
	if ui.indicatorFlash > 60 {
		ui.indicatorFlash = 0
	}

	// Track zone changes
	currentZoneType := ui.zoneManager.GetZoneTypeAt(playerX, playerZ)
	if currentZoneType != ui.lastZoneType {
		ui.lastZoneType = currentZoneType
		// Could play zone transition sound here
	}

	// Toggle PvP flag with P key
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		ui.wantsFlagged = !ui.wantsFlagged
		ui.zoneManager.SetPlayerFlag(playerEntity, ui.wantsFlagged)

		// Set cooldown (5 minutes to unflag)
		if ui.wantsFlagged {
			ui.zoneManager.SetFlagCooldown(playerEntity, time.Now().Add(5*time.Minute).Unix())
		}
	}

	// Clear expired flags
	ui.zoneManager.ClearExpiredFlags(time.Now().Unix())

	// Clear old loot drop display
	if ui.recentDrop != nil && time.Since(ui.dropShowTime) > 5*time.Second {
		ui.recentDrop = nil
	}
}

// Draw renders PvP zone indicators.
func (ui *PvPUI) Draw(screen *ebiten.Image, playerEntity uint64, playerX, playerZ float64) {
	if !ui.showIndicator {
		return
	}

	// Get current zone info
	zoneType := ui.zoneManager.GetZoneTypeAt(playerX, playerZ)
	zone := ui.zoneManager.GetZoneAt(playerX, playerZ)
	isFlagged := ui.zoneManager.IsPlayerFlagged(playerEntity)

	// Draw zone indicator
	ui.drawZoneIndicator(screen, zoneType, zone, isFlagged)

	// Draw PvP status
	ui.drawPvPStatus(screen, isFlagged)

	// Draw loot drop notification
	if ui.recentDrop != nil {
		ui.drawLootDrop(screen)
	}
}

// drawZoneIndicator renders the current zone type indicator.
func (ui *PvPUI) drawZoneIndicator(screen *ebiten.Image, zoneType pvp.ZoneType, zone *pvp.Zone, isFlagged bool) {
	// Position in top-right area
	x, y := screen.Bounds().Dx()-200, 10

	var zoneName, zoneDesc string
	var bgColor, textColor color.RGBA

	switch zoneType {
	case pvp.ZoneSafe:
		zoneName = "SAFE ZONE"
		zoneDesc = "No PvP allowed"
		bgColor = color.RGBA{20, 80, 20, 200}
		textColor = color.RGBA{100, 255, 100, 255}

	case pvp.ZoneContested:
		zoneName = "CONTESTED"
		zoneDesc = "Opt-in PvP"
		bgColor = color.RGBA{80, 80, 20, 200}
		textColor = color.RGBA{255, 255, 100, 255}
		// Flash if flagged
		if isFlagged && ui.indicatorFlash < 30 {
			bgColor = color.RGBA{100, 100, 30, 200}
		}

	case pvp.ZoneHostile:
		zoneName = "HOSTILE"
		zoneDesc = "Full PvP - Loot drops!"
		bgColor = color.RGBA{100, 20, 20, 200}
		textColor = color.RGBA{255, 100, 100, 255}
		// Flash warning
		if ui.indicatorFlash < 30 {
			bgColor = color.RGBA{120, 30, 30, 200}
		}
	}

	// Draw background
	drawRect(screen, x, y, 190, 50, bgColor)

	// Draw zone name and description
	drawText(screen, zoneName, x+10, y+5, textColor)
	drawText(screen, zoneDesc, x+10, y+25, color.RGBA{180, 180, 180, 255})

	// Show zone ID if available
	if zone != nil {
		drawText(screen, zone.ID, x+10, y+38, color.RGBA{120, 120, 120, 255})
	}
}

// drawPvPStatus renders the player's PvP flag status.
func (ui *PvPUI) drawPvPStatus(screen *ebiten.Image, isFlagged bool) {
	x, y := screen.Bounds().Dx()-200, 70

	var statusText string
	var statusColor color.RGBA

	if isFlagged {
		statusText = "PvP FLAGGED"
		statusColor = color.RGBA{255, 80, 80, 255}
		// Flash warning
		if ui.indicatorFlash < 30 {
			statusColor = color.RGBA{255, 150, 150, 255}
		}
	} else {
		statusText = "PvP: Safe"
		statusColor = color.RGBA{100, 200, 100, 255}
	}

	drawRect(screen, x, y, 190, 25, color.RGBA{30, 30, 30, 200})
	drawText(screen, statusText, x+10, y+5, statusColor)
	drawText(screen, "[P] Toggle", x+120, y+5, color.RGBA{100, 100, 100, 255})
}

// drawLootDrop renders a loot drop notification.
func (ui *PvPUI) drawLootDrop(screen *ebiten.Image) {
	if ui.recentDrop == nil {
		return
	}

	// Center of screen notification
	screenW := screen.Bounds().Dx()
	x := screenW/2 - 100
	y := 150

	// Fade based on time
	elapsed := time.Since(ui.dropShowTime).Seconds()
	alpha := uint8(255 - int(elapsed*50))
	if alpha < 50 {
		alpha = 50
	}

	bgColor := color.RGBA{100, 30, 30, alpha}
	drawRect(screen, x, y, 200, 60, bgColor)

	textColor := color.RGBA{255, 200, 200, alpha}
	drawText(screen, "LOOT DROPPED!", x+50, y+5, textColor)
	drawText(screen, fmt.Sprintf("%d items", len(ui.recentDrop.Items)), x+70, y+25, textColor)
	drawText(screen, fmt.Sprintf("At: (%.0f, %.0f)", ui.recentDrop.X, ui.recentDrop.Z), x+50, y+40, color.RGBA{180, 180, 180, alpha})
}

// CheckCombat evaluates if PvP damage should be applied between two players.
func (ui *PvPUI) CheckCombat(
	attackerID uint64, attackerX, attackerZ float64,
	defenderID uint64, defenderX, defenderZ float64,
	baseDamage float64,
) *pvp.CombatResult {
	return ui.zoneManager.CheckCombat(
		attackerID, attackerX, attackerZ,
		defenderID, defenderX, defenderZ,
		baseDamage,
	)
}

// ProcessDeath handles a player death in PvP, returning dropped loot.
func (ui *PvPUI) ProcessDeath(entityID uint64, inventory []string, x, z float64) *pvp.DeathLoot {
	loot := ui.zoneManager.CalculateDeathLoot(entityID, inventory, x, z, ui.rng)
	if loot != nil {
		ui.recentDrop = loot
		ui.dropShowTime = time.Now()
	}
	return loot
}

// GetRespawnPoint returns where a dead player should respawn.
func (ui *PvPUI) GetRespawnPoint(x, z float64) (respawnX, respawnZ float64) {
	return ui.zoneManager.RespawnPoint(x, z)
}

// GetZoneManager returns the underlying zone manager.
func (ui *PvPUI) GetZoneManager() *pvp.ZoneManager {
	return ui.zoneManager
}

// IsInSafeZone checks if the position is in a safe zone.
func (ui *PvPUI) IsInSafeZone(x, z float64) bool {
	return ui.zoneManager.GetZoneTypeAt(x, z) == pvp.ZoneSafe
}

// IsInHostileZone checks if the position is in a hostile zone.
func (ui *PvPUI) IsInHostileZone(x, z float64) bool {
	return ui.zoneManager.GetZoneTypeAt(x, z) == pvp.ZoneHostile
}

// SetPlayerFlag sets a player's PvP flag status.
func (ui *PvPUI) SetPlayerFlag(entityID uint64, flagged bool) {
	ui.zoneManager.SetPlayerFlag(entityID, flagged)
}

// IsPlayerFlagged returns whether a player is flagged for PvP.
func (ui *PvPUI) IsPlayerFlagged(entityID uint64) bool {
	return ui.zoneManager.IsPlayerFlagged(entityID)
}

// DrawZoneBoundaryIndicator draws an on-screen indicator when near zone boundaries.
func (ui *PvPUI) DrawZoneBoundaryIndicator(screen *ebiten.Image, playerX, playerZ float64) {
	// Check distance to any zone boundary
	zone := ui.zoneManager.GetZoneAt(playerX, playerZ)
	if zone == nil {
		return
	}

	// Calculate distances to edges
	distToLeft := playerX - zone.MinX
	distToRight := zone.MaxX - playerX
	distToBack := playerZ - zone.MinZ
	distToFront := zone.MaxZ - playerZ

	minDist := distToLeft
	direction := "west"
	if distToRight < minDist {
		minDist = distToRight
		direction = "east"
	}
	if distToBack < minDist {
		minDist = distToBack
		direction = "south"
	}
	if distToFront < minDist {
		minDist = distToFront
		direction = "north"
	}

	// Show warning if near boundary
	if minDist < 20 {
		screenW := screen.Bounds().Dx()
		x := screenW/2 - 80
		y := 100

		alpha := uint8(255 - int(minDist*10))
		if alpha < 100 {
			alpha = 100
		}

		bgColor := color.RGBA{80, 80, 20, alpha}
		drawRect(screen, x, y, 160, 30, bgColor)
		drawText(screen, fmt.Sprintf("Zone boundary %s", direction), x+10, y+8, color.RGBA{255, 255, 150, alpha})
	}
}

// Helper to get player inventory as string list for loot drops.
func getPlayerInventoryStrings(g *Game) []string {
	if g.world == nil {
		return nil
	}
	invComp, ok := g.world.GetComponent(g.playerEntity, "Inventory")
	if !ok {
		return nil
	}
	inv := invComp.(*components.Inventory)

	items := make([]string, 0, len(inv.Items))
	for _, item := range inv.Items {
		items = append(items, item.ID)
	}
	return items
}
