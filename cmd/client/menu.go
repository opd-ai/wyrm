//go:build !noebiten

// menu.go provides the game menu system for pause, settings, and quit.
package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/input"
)

// MenuState represents the current menu state.
type MenuState int

const (
	MenuStateNone MenuState = iota
	MenuStatePause
	MenuStateSettings
	MenuStateQuit
)

// MenuItem represents a single menu item.
type MenuItem struct {
	Label    string
	Action   func()
	Disabled bool
}

// Menu manages the game menu system.
type Menu struct {
	state         MenuState
	items         []MenuItem
	selectedIndex int
	cfg           *config.Config
	inputManager  *input.Manager
	quitRequested bool
	// Settings state
	volumeLevel   int // 0-10
	musicEnabled  bool
	sfxEnabled    bool
	fullscreen    bool
	showFPS       bool
	settingsPage  int // 0 = main settings, 1 = keybindings
	settingsItems []MenuItem
	settingsIndex int
}

// NewMenu creates a new menu system.
func NewMenu(cfg *config.Config, inputManager *input.Manager) *Menu {
	m := &Menu{
		state:         MenuStateNone,
		selectedIndex: 0,
		cfg:           cfg,
		inputManager:  inputManager,
		volumeLevel:   7,
		musicEnabled:  true,
		sfxEnabled:    true,
		fullscreen:    false,
		showFPS:       true,
		settingsPage:  0,
	}
	m.buildPauseMenu()
	m.buildSettingsMenu()
	return m
}

// buildPauseMenu constructs the pause menu items.
func (m *Menu) buildPauseMenu() {
	m.items = []MenuItem{
		{Label: "Resume", Action: m.resume},
		{Label: "Settings", Action: m.openSettings},
		{Label: "Save Game", Action: m.saveGame, Disabled: true}, // TODO: Implement save
		{Label: "Load Game", Action: m.loadGame, Disabled: true}, // TODO: Implement load
		{Label: "Quit to Desktop", Action: m.requestQuit},
	}
}

// buildSettingsMenu constructs the settings menu items.
func (m *Menu) buildSettingsMenu() {
	m.settingsItems = []MenuItem{
		{Label: fmt.Sprintf("Music Volume: %d", m.volumeLevel), Action: m.cycleVolume},
		{Label: fmt.Sprintf("Music: %s", boolToOnOff(m.musicEnabled)), Action: m.toggleMusic},
		{Label: fmt.Sprintf("Sound Effects: %s", boolToOnOff(m.sfxEnabled)), Action: m.toggleSFX},
		{Label: fmt.Sprintf("Fullscreen: %s", boolToOnOff(m.fullscreen)), Action: m.toggleFullscreen},
		{Label: fmt.Sprintf("Show FPS: %s", boolToOnOff(m.showFPS)), Action: m.toggleFPS},
		{Label: "Key Bindings...", Action: m.openKeybindings, Disabled: true}, // TODO: Implement keybindings UI
		{Label: "Back", Action: m.closeSettings},
	}
}

// boolToOnOff converts a boolean to "ON" or "OFF" string.
func boolToOnOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

// IsOpen returns true if any menu is currently open.
func (m *Menu) IsOpen() bool {
	return m.state != MenuStateNone
}

// QuitRequested returns true if the user requested to quit.
func (m *Menu) QuitRequested() bool {
	return m.quitRequested
}

// Toggle opens/closes the pause menu.
func (m *Menu) Toggle() {
	if m.state == MenuStateNone {
		m.state = MenuStatePause
		m.selectedIndex = 0
	} else {
		m.state = MenuStateNone
	}
}

// Update processes menu input.
func (m *Menu) Update() {
	if m.state == MenuStateNone {
		return
	}

	// Navigate menu
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.moveSelection(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.moveSelection(1)
	}

	// Select item
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		m.selectCurrentItem()
	}

	// Back/cancel
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		m.handleBack()
	}

	// Left/right for value adjustment (volume, etc.)
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		m.adjustValue(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		m.adjustValue(1)
	}
}

// moveSelection moves the selection cursor up or down.
func (m *Menu) moveSelection(delta int) {
	items := m.getCurrentItems()
	if len(items) == 0 {
		return
	}
	index := m.getCurrentIndex()
	newIndex := index + delta
	if newIndex < 0 {
		newIndex = len(items) - 1
	}
	if newIndex >= len(items) {
		newIndex = 0
	}
	m.setCurrentIndex(newIndex)
}

// getCurrentItems returns the currently active menu items.
func (m *Menu) getCurrentItems() []MenuItem {
	if m.state == MenuStateSettings {
		return m.settingsItems
	}
	return m.items
}

// getCurrentIndex returns the current selection index.
func (m *Menu) getCurrentIndex() int {
	if m.state == MenuStateSettings {
		return m.settingsIndex
	}
	return m.selectedIndex
}

// setCurrentIndex sets the current selection index.
func (m *Menu) setCurrentIndex(index int) {
	if m.state == MenuStateSettings {
		m.settingsIndex = index
	} else {
		m.selectedIndex = index
	}
}

// selectCurrentItem activates the currently selected menu item.
func (m *Menu) selectCurrentItem() {
	items := m.getCurrentItems()
	index := m.getCurrentIndex()
	if index < 0 || index >= len(items) {
		return
	}
	item := items[index]
	if item.Disabled || item.Action == nil {
		return
	}
	item.Action()
}

// handleBack goes back one menu level or closes the menu.
func (m *Menu) handleBack() {
	switch m.state {
	case MenuStateSettings:
		m.closeSettings()
	case MenuStatePause:
		m.resume()
	default:
		m.state = MenuStateNone
	}
}

// adjustValue adjusts the value of a settings item (for sliders).
func (m *Menu) adjustValue(delta int) {
	if m.state != MenuStateSettings {
		return
	}
	// Currently only volume can be adjusted with left/right
	if m.settingsIndex == 0 { // Volume item
		m.volumeLevel += delta
		if m.volumeLevel < 0 {
			m.volumeLevel = 0
		}
		if m.volumeLevel > 10 {
			m.volumeLevel = 10
		}
		m.buildSettingsMenu() // Rebuild to update label
	}
}

// Menu actions

func (m *Menu) resume() {
	m.state = MenuStateNone
}

func (m *Menu) openSettings() {
	m.state = MenuStateSettings
	m.settingsIndex = 0
}

func (m *Menu) closeSettings() {
	m.state = MenuStatePause
	m.buildSettingsMenu() // Rebuild to update labels
}

func (m *Menu) saveGame() {
	// TODO: Implement save game
}

func (m *Menu) loadGame() {
	// TODO: Implement load game
}

func (m *Menu) requestQuit() {
	m.quitRequested = true
}

func (m *Menu) cycleVolume() {
	m.volumeLevel = (m.volumeLevel + 1) % 11
	m.buildSettingsMenu()
}

func (m *Menu) toggleMusic() {
	m.musicEnabled = !m.musicEnabled
	m.buildSettingsMenu()
}

func (m *Menu) toggleSFX() {
	m.sfxEnabled = !m.sfxEnabled
	m.buildSettingsMenu()
}

func (m *Menu) toggleFullscreen() {
	m.fullscreen = !m.fullscreen
	ebiten.SetFullscreen(m.fullscreen)
	m.buildSettingsMenu()
}

func (m *Menu) toggleFPS() {
	m.showFPS = !m.showFPS
	m.buildSettingsMenu()
}

func (m *Menu) openKeybindings() {
	// TODO: Implement keybindings UI
	m.settingsPage = 1
}

// Draw renders the menu overlay.
func (m *Menu) Draw(screen *ebiten.Image) {
	if m.state == MenuStateNone {
		return
	}

	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	// Get screen dimensions
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	centerX := w / 2
	centerY := h / 2

	// Draw title
	title := m.getTitle()
	ebitenutil.DebugPrintAt(screen, title, centerX-len(title)*3, centerY-100)

	// Draw menu items
	items := m.getCurrentItems()
	index := m.getCurrentIndex()
	startY := centerY - 50
	for i, item := range items {
		label := item.Label
		if i == index {
			label = "> " + label + " <"
		}
		if item.Disabled {
			label = "  " + item.Label + " (disabled)"
		}
		x := centerX - len(label)*3
		y := startY + i*25
		ebitenutil.DebugPrintAt(screen, label, x, y)
	}

	// Draw navigation hints
	hints := "UP/DOWN: Navigate | ENTER: Select | ESC: Back"
	if m.state == MenuStateSettings && m.settingsIndex == 0 {
		hints = "LEFT/RIGHT: Adjust | " + hints
	}
	ebitenutil.DebugPrintAt(screen, hints, centerX-len(hints)*3, h-30)
}

// getTitle returns the title for the current menu state.
func (m *Menu) getTitle() string {
	switch m.state {
	case MenuStatePause:
		return "=== PAUSED ==="
	case MenuStateSettings:
		return "=== SETTINGS ==="
	default:
		return ""
	}
}

// GetVolumeLevel returns the current volume level (0-10).
func (m *Menu) GetVolumeLevel() int {
	return m.volumeLevel
}

// IsMusicEnabled returns whether music is enabled.
func (m *Menu) IsMusicEnabled() bool {
	return m.musicEnabled
}

// IsSFXEnabled returns whether sound effects are enabled.
func (m *Menu) IsSFXEnabled() bool {
	return m.sfxEnabled
}

// ShowFPS returns whether FPS should be displayed.
func (m *Menu) ShowFPS() bool {
	return m.showFPS
}
