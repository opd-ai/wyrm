//go:build !noebiten

// Package main provides the game client entry point.
// hud.go contains HUD (heads-up display) rendering functions including
// health/mana bars, minimap, wanted status, compass direction, and related utilities.
package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/wyrm/pkg/engine/components"
)

// drawHUD renders the heads-up display elements.
func (g *Game) drawHUD(screen *ebiten.Image) {
	screenWidth := g.cfg.Window.Width
	screenHeight := g.cfg.Window.Height

	// Get player components
	var pos *components.Position
	var health *components.Health
	var mana *components.Mana
	if g.playerEntity != 0 {
		if comp, ok := g.world.GetComponent(g.playerEntity, "Position"); ok {
			pos = comp.(*components.Position)
		}
		if comp, ok := g.world.GetComponent(g.playerEntity, "Health"); ok {
			health = comp.(*components.Health)
		}
		if comp, ok := g.world.GetComponent(g.playerEntity, "Mana"); ok {
			mana = comp.(*components.Mana)
		}
	}

	// Draw health bar (bottom-left)
	barWidth := 150
	barHeight := 12
	barX := 10
	barY := screenHeight - 50
	if health != nil {
		healthPercent := health.Current / health.Max
		g.drawBar(screen, barX, barY, barWidth, barHeight, healthPercent, 0xCC0000FF, 0x440000FF)
	}

	// Draw mana bar (below health)
	if mana != nil {
		manaPercent := mana.Current / mana.Max
		g.drawBar(screen, barX, barY+16, barWidth, barHeight, manaPercent, 0x0066CCFF, 0x002244FF)
	}

	// Draw position and compass (top-left)
	status := "offline"
	if g.connected {
		status = "online"
	}
	coordText := fmt.Sprintf("Wyrm [%s] %s", g.cfg.Genre, status)
	if pos != nil {
		chunkX := int(pos.X) / g.cfg.World.ChunkSize
		chunkY := int(pos.Y) / g.cfg.World.ChunkSize
		direction := getCompassDirection(pos.Angle)
		coordText = fmt.Sprintf("Wyrm [%s] %s\nPos: %.1f, %.1f | Chunk: %d, %d\nHeading: %s",
			g.cfg.Genre, status, pos.X, pos.Y, chunkX, chunkY, direction)
	}
	ebitenutil.DebugPrint(screen, coordText)

	// Draw minimap (top-right)
	g.drawMinimap(screen, screenWidth-80, 10, 64)

	// Draw bounty/wanted status (below minimap)
	g.drawWantedStatus(screen, screenWidth-80, 80)

	// Draw interaction prompt (bottom-center)
	g.drawInteractionPrompt(screen)

	// Draw debug info if enabled
	g.drawDebugInfo(screen)
}

// drawWantedStatus renders the player's bounty and wanted status.
func (g *Game) drawWantedStatus(screen *ebiten.Image, x, y int) {
	crime := g.getPlayerCrime()
	if crime == nil || !g.shouldShowWantedStatus(crime) {
		return
	}

	boxWidth, boxHeight := g.calculateWantedBoxSize(crime)
	g.drawWantedBackground(screen, x, y, boxWidth, boxHeight)
	g.drawWantedBorder(screen, x, y, boxWidth, crime.WantedLevel)
	g.drawWantedText(screen, x, y, crime)
}

// getPlayerCrime returns the player's Crime component or nil.
func (g *Game) getPlayerCrime() *components.Crime {
	if g.playerEntity == 0 {
		return nil
	}
	crimeComp, ok := g.world.GetComponent(g.playerEntity, "Crime")
	if !ok {
		return nil
	}
	return crimeComp.(*components.Crime)
}

// shouldShowWantedStatus returns true if the wanted status should be displayed.
func (g *Game) shouldShowWantedStatus(crime *components.Crime) bool {
	return crime.WantedLevel > 0 || crime.BountyAmount > 0 || crime.InJail
}

// calculateWantedBoxSize returns the box dimensions based on jail status.
func (g *Game) calculateWantedBoxSize(crime *components.Crime) (width, height int) {
	width = 70
	height = 40
	if crime.InJail {
		height = 50
	}
	return width, height
}

// drawWantedBackground renders the background rectangle.
func (g *Game) drawWantedBackground(screen *ebiten.Image, x, y, w, h int) {
	bgColor := color.RGBA{0, 0, 0, 180}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), bgColor)
}

// drawWantedBorder renders the border with color based on wanted level.
func (g *Game) drawWantedBorder(screen *ebiten.Image, x, y, w, wantedLevel int) {
	borderColor := getWantedBorderColor(wantedLevel)
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w), 2, borderColor)
}

// getWantedBorderColor returns the border color for a given wanted level.
func getWantedBorderColor(level int) color.RGBA {
	switch level {
	case 1:
		return color.RGBA{255, 255, 0, 255} // Yellow
	case 2:
		return color.RGBA{255, 165, 0, 255} // Orange
	case 3, 4, 5:
		return color.RGBA{255, 0, 0, 255} // Red
	default:
		return color.RGBA{100, 100, 100, 255}
	}
}

// drawWantedText renders the wanted level, bounty, and jail status text.
func (g *Game) drawWantedText(screen *ebiten.Image, x, y int, crime *components.Crime) {
	if crime.WantedLevel > 0 {
		stars := ""
		for i := 0; i < crime.WantedLevel; i++ {
			stars += "*"
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("WANTED %s", stars), x+5, y+5)
	}
	if crime.BountyAmount > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Bounty: %.0fg", crime.BountyAmount), x+5, y+20)
	}
	if crime.InJail {
		ebitenutil.DebugPrintAt(screen, "IN JAIL", x+5, y+35)
	}
}

// drawBar renders a horizontal bar (health/mana style) using batch pixel writes.
// Uses WritePixels() for GPU-efficient rendering with zero per-pixel Set() calls.
func (g *Game) drawBar(screen *ebiten.Image, x, y, width, height int, percent float64, fillColor, bgColor uint32) {
	g.ensureBarBuffers(width, height)

	// Use pre-allocated pixel buffer (slice to required size)
	bufSize := width * height * 4
	pixels := g.barPixels[:bufSize]

	g.fillBarPixels(pixels, width, height, percent, fillColor, bgColor)

	// Create a sub-image of the correct size and write pixels
	subImg := g.barImage.SubImage(image.Rect(0, 0, width, height)).(*ebiten.Image)
	subImg.WritePixels(pixels)

	// Draw to screen
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(subImg, op)
}

// ensureBarBuffers initializes or resizes bar rendering buffers.
func (g *Game) ensureBarBuffers(width, height int) {
	if g.barImage == nil {
		g.barImage = ebiten.NewImage(200, 32)
		g.barPixels = make([]byte, 200*32*4)
	}

	bounds := g.barImage.Bounds()
	if bounds.Dx() < width || bounds.Dy() < height {
		newW := max(bounds.Dx(), width)
		newH := max(bounds.Dy(), height)
		g.barImage = ebiten.NewImage(newW, newH)
		g.barPixels = make([]byte, newW*newH*4)
	} else if len(g.barPixels) < width*height*4 {
		g.barPixels = make([]byte, bounds.Dx()*bounds.Dy()*4)
	}

	g.barImage.Clear()
}

// fillBarPixels renders the bar content to the pixel buffer.
func (g *Game) fillBarPixels(pixels []byte, width, height int, percent float64, fillColor, bgColor uint32) {
	bgR, bgG, bgB, bgA := uint8(bgColor>>24), uint8(bgColor>>16), uint8(bgColor>>8), uint8(bgColor)
	fillR, fillG, fillB, fillA := uint8(fillColor>>24), uint8(fillColor>>16), uint8(fillColor>>8), uint8(fillColor)

	// Fill background
	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			idx := (py*width + px) * 4
			pixels[idx] = bgR
			pixels[idx+1] = bgG
			pixels[idx+2] = bgB
			pixels[idx+3] = bgA
		}
	}

	// Fill progress (with 1px border)
	fillWidth := int(float64(width) * percent)
	for py := 1; py < height-1; py++ {
		for px := 1; px < fillWidth-1 && px < width-1; px++ {
			idx := (py*width + px) * 4
			pixels[idx] = fillR
			pixels[idx+1] = fillG
			pixels[idx+2] = fillB
			pixels[idx+3] = fillA
		}
	}
}

// drawMinimap renders a small top-down view of the nearby area using batch pixel writes.
// Uses WritePixels() for GPU-efficient rendering with zero per-pixel Set() calls.
func (g *Game) drawMinimap(screen *ebiten.Image, x, y, size int) {
	if g.worldMap == nil || len(g.worldMap) == 0 {
		return
	}

	g.ensureMinimapBuffers(size)

	playerMapX, playerMapY := g.getMinimapPlayerPosition()
	territories := g.getFactionTerritories()

	pixels := g.minimapPixels[:size*size*4]
	g.fillMinimapPixels(pixels, size, playerMapX, playerMapY, territories)
	g.drawMinimapPlayerMarker(pixels, size)

	// Write pixels and draw to screen
	subImg := g.minimapImage.SubImage(image.Rect(0, 0, size, size)).(*ebiten.Image)
	subImg.WritePixels(pixels)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(subImg, op)
}

// ensureMinimapBuffers initializes or resizes minimap pixel buffers.
func (g *Game) ensureMinimapBuffers(size int) {
	if g.minimapImage == nil || g.minimapPixels == nil {
		g.minimapImage = ebiten.NewImage(size, size)
		g.minimapPixels = make([]byte, size*size*4)
		return
	}
	bounds := g.minimapImage.Bounds()
	if bounds.Dx() < size || bounds.Dy() < size {
		g.minimapImage = ebiten.NewImage(size, size)
		g.minimapPixels = make([]byte, size*size*4)
	} else if len(g.minimapPixels) < size*size*4 {
		g.minimapPixels = make([]byte, size*size*4)
	}
}

// getMinimapPlayerPosition returns the player's position for minimap centering.
func (g *Game) getMinimapPlayerPosition() (int, int) {
	if g.playerEntity == 0 {
		return 0, 0
	}
	comp, ok := g.world.GetComponent(g.playerEntity, "Position")
	if !ok {
		return 0, 0
	}
	pos := comp.(*components.Position)
	return int(pos.X), int(pos.Y)
}

// fillMinimapPixels renders terrain and territory colors to the pixel buffer.
func (g *Game) fillMinimapPixels(pixels []byte, size, playerX, playerY int, territories []*components.FactionTerritory) {
	mapRadius := 16
	for my := 0; my < size; my++ {
		for mx := 0; mx < size; mx++ {
			worldX := playerX - mapRadius + (mx * mapRadius * 2 / size)
			worldY := playerY - mapRadius + (my * mapRadius * 2 / size)

			baseColor := g.getMinimapTileColor(worldX, worldY)
			if tc := g.getTerritoryColor(float64(worldX), float64(worldY), territories); tc != 0 {
				baseColor = blendColors(baseColor, tc, 0.4)
			}

			idx := (my*size + mx) * 4
			pixels[idx] = uint8(baseColor >> 24)
			pixels[idx+1] = uint8(baseColor >> 16)
			pixels[idx+2] = uint8(baseColor >> 8)
			pixels[idx+3] = uint8(baseColor)
		}
	}
}

// getMinimapTileColor returns the base color for a minimap tile.
func (g *Game) getMinimapTileColor(worldX, worldY int) uint32 {
	if worldY < 0 || worldY >= len(g.worldMap) || worldX < 0 || worldX >= len(g.worldMap[0]) {
		return 0x111111FF
	}
	if g.worldMap[worldY][worldX] > 0 {
		return 0x666666FF
	}
	return 0x333333FF
}

// drawMinimapPlayerMarker draws the green player marker in the center of the minimap.
func (g *Game) drawMinimapPlayerMarker(pixels []byte, size int) {
	center := size / 2
	green := []byte{0, 255, 0, 255}
	offsets := [][2]int{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for _, off := range offsets {
		px, py := center+off[0], center+off[1]
		if px >= 0 && px < size && py >= 0 && py < size {
			idx := (py*size + px) * 4
			copy(pixels[idx:idx+4], green)
		}
	}
}

// getFactionTerritories retrieves all faction territory entities.
func (g *Game) getFactionTerritories() []*components.FactionTerritory {
	var territories []*components.FactionTerritory
	for _, e := range g.world.Entities("FactionTerritory") {
		if comp, ok := g.world.GetComponent(e, "FactionTerritory"); ok {
			territories = append(territories, comp.(*components.FactionTerritory))
		}
	}
	return territories
}

// getTerritoryColor returns a color for the territory at the given point.
func (g *Game) getTerritoryColor(x, y float64, territories []*components.FactionTerritory) uint32 {
	for _, t := range territories {
		if t.ContainsPoint(x, y) {
			return g.factionToColor(t.FactionID)
		}
	}
	return 0
}

// factionToColor maps faction IDs to minimap colors.
func (g *Game) factionToColor(factionID string) uint32 {
	// Use a simple hash to get consistent colors per faction
	hash := uint32(0)
	for _, c := range factionID {
		hash = hash*31 + uint32(c)
	}
	// Generate color with moderate saturation and alpha
	rr := uint8(80 + (hash % 80))
	gg := uint8(80 + ((hash / 256) % 80))
	bb := uint8(80 + ((hash / 65536) % 80))
	return uint32(rr)<<24 | uint32(gg)<<16 | uint32(bb)<<8 | 0xCC
}

// blendColors blends two RGBA colors by the given factor (0=first, 1=second).
func blendColors(c1, c2 uint32, factor float64) uint32 {
	r1, g1, b1, a1 := uint8((c1>>24)&0xFF), uint8((c1>>16)&0xFF), uint8((c1>>8)&0xFF), uint8(c1&0xFF)
	r2, g2, b2, a2 := uint8((c2>>24)&0xFF), uint8((c2>>16)&0xFF), uint8((c2>>8)&0xFF), uint8(c2&0xFF)

	r := uint8(float64(r1)*(1-factor) + float64(r2)*factor)
	g := uint8(float64(g1)*(1-factor) + float64(g2)*factor)
	b := uint8(float64(b1)*(1-factor) + float64(b2)*factor)
	a := uint8(float64(a1)*(1-factor) + float64(a2)*factor)

	return uint32(r)<<24 | uint32(g)<<16 | uint32(b)<<8 | uint32(a)
}

// getCompassDirection returns cardinal/ordinal direction name from angle.
func getCompassDirection(angle float64) string {
	// Normalize angle to 0-2π
	for angle < 0 {
		angle += 2 * math.Pi
	}
	for angle >= 2*math.Pi {
		angle -= 2 * math.Pi
	}

	// Convert to compass direction (0 = East in math coordinates)
	directions := []string{"E", "NE", "N", "NW", "W", "SW", "S", "SE"}
	index := int((angle+(math.Pi/8))/(math.Pi/4)) % 8
	return directions[index]
}

// uint32ToColor converts a hex color (RRGGBBAA) to color.RGBA.
func uint32ToColor(c uint32) color.RGBA {
	return color.RGBA{
		R: uint8((c >> 24) & 0xFF),
		G: uint8((c >> 16) & 0xFF),
		B: uint8((c >> 8) & 0xFF),
		A: uint8(c & 0xFF),
	}
}
