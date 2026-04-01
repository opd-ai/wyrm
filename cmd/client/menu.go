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
	MenuStateQuitConfirm
	MenuStateKeyBindings
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
	// Quit confirmation state
	quitConfirmItems []MenuItem
	quitConfirmIndex int
	// Key binding state
	keybindItems        []MenuItem
	keybindIndex        int
	waitingForKeyInput  bool   // true when waiting for player to press a key
	keyBindingToChange  string // the action being rebound
	keybindScrollOffset int    // for scrolling through long list
	// Save/Load callbacks (set by Game)
	onSaveRequest func() error
	onLoadRequest func() error
	saveMessage   string // status message after save/load
}

// NewMenu creates a new menu system.
func NewMenu(cfg *config.Config, inputManager *input.Manager) *Menu {
	m := &Menu{
		state:         MenuStateNone,
		selectedIndex: 0,
		cfg:           cfg,
		inputManager:  inputManager,
		volumeLevel:   cfg.Audio.MasterVolume,
		musicEnabled:  cfg.Audio.MusicEnabled,
		sfxEnabled:    cfg.Audio.SFXEnabled,
		fullscreen:    cfg.Window.Fullscreen,
		showFPS:       cfg.Window.ShowFPS,
		settingsPage:  0,
	}
	m.buildPauseMenu()
	m.buildSettingsMenu()
	m.buildQuitConfirmMenu()
	m.buildKeybindMenu()
	return m
}

// SetSaveHandler sets the callback for save requests.
func (m *Menu) SetSaveHandler(fn func() error) {
	m.onSaveRequest = fn
	m.buildPauseMenu() // Rebuild to enable Save item
}

// SetLoadHandler sets the callback for load requests.
func (m *Menu) SetLoadHandler(fn func() error) {
	m.onLoadRequest = fn
	m.buildPauseMenu() // Rebuild to enable Load item
}

// buildPauseMenu constructs the pause menu items.
func (m *Menu) buildPauseMenu() {
	m.items = []MenuItem{
		{Label: "Resume", Action: m.resume},
		{Label: "Save Game", Action: m.saveGame, Disabled: m.onSaveRequest == nil},
		{Label: "Load Game", Action: m.loadGame, Disabled: m.onLoadRequest == nil},
		{Label: "Settings", Action: m.openSettings},
		{Label: "Save Settings", Action: m.saveSettings},
		{Label: "Quit to Desktop", Action: m.showQuitConfirm},
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
		{Label: "Key Bindings...", Action: m.openKeybindings},
		{Label: "Back", Action: m.closeSettings},
	}
}

// buildQuitConfirmMenu constructs the quit confirmation menu items.
func (m *Menu) buildQuitConfirmMenu() {
	m.quitConfirmItems = []MenuItem{
		{Label: "Yes, Quit", Action: m.confirmQuit},
		{Label: "No, Return to Game", Action: m.cancelQuit},
	}
}

// keybindActionOrder defines the order of actions in the keybind menu.
var keybindActionOrder = []input.Action{
	// Movement
	input.ActionMoveForward,
	input.ActionMoveBackward,
	input.ActionMoveLeft,
	input.ActionMoveRight,
	input.ActionJump,
	input.ActionCrouch,
	input.ActionSprint,
	// Combat
	input.ActionAttack,
	input.ActionBlock,
	input.ActionAbility1,
	input.ActionAbility2,
	input.ActionAbility3,
	input.ActionAbility4,
	input.ActionQuickHeal,
	input.ActionToggleWeapon,
	// Interaction
	input.ActionInteract,
	input.ActionPickUp,
	input.ActionDropItem,
	input.ActionUseItem,
	input.ActionTalk,
	input.ActionMount,
	input.ActionEnterVehicle,
	// UI
	input.ActionInventory,
	input.ActionMap,
	input.ActionQuestLog,
	input.ActionCharSheet,
	input.ActionSkillTree,
	input.ActionCrafting,
	input.ActionPause,
	input.ActionQuickSave,
	input.ActionQuickLoad,
}

// actionDisplayNames maps actions to user-friendly display names.
var actionDisplayNames = map[input.Action]string{
	input.ActionMoveForward:  "Move Forward",
	input.ActionMoveBackward: "Move Backward",
	input.ActionMoveLeft:     "Strafe Left",
	input.ActionMoveRight:    "Strafe Right",
	input.ActionJump:         "Jump",
	input.ActionCrouch:       "Crouch",
	input.ActionSprint:       "Sprint",
	input.ActionAttack:       "Attack",
	input.ActionBlock:        "Block/Parry",
	input.ActionAbility1:     "Ability 1",
	input.ActionAbility2:     "Ability 2",
	input.ActionAbility3:     "Ability 3",
	input.ActionAbility4:     "Ability 4",
	input.ActionQuickHeal:    "Quick Heal",
	input.ActionToggleWeapon: "Toggle Weapon",
	input.ActionInteract:     "Interact",
	input.ActionPickUp:       "Pick Up",
	input.ActionDropItem:     "Drop Item",
	input.ActionUseItem:      "Use Item",
	input.ActionTalk:         "Talk",
	input.ActionMount:        "Mount/Dismount",
	input.ActionEnterVehicle: "Enter Vehicle",
	input.ActionInventory:    "Inventory",
	input.ActionMap:          "Map",
	input.ActionQuestLog:     "Quest Log",
	input.ActionCharSheet:    "Character",
	input.ActionSkillTree:    "Skill Tree",
	input.ActionCrafting:     "Crafting",
	input.ActionPause:        "Pause/Menu",
	input.ActionQuickSave:    "Quick Save",
	input.ActionQuickLoad:    "Quick Load",
}

// buildKeybindMenu constructs the key bindings menu items.
func (m *Menu) buildKeybindMenu() {
	m.keybindItems = make([]MenuItem, 0, len(keybindActionOrder)+2)

	for _, action := range keybindActionOrder {
		displayName := actionDisplayNames[action]
		if displayName == "" {
			displayName = string(action)
		}
		currentKey := m.inputManager.GetBinding(action)
		label := fmt.Sprintf("%-16s: %s", displayName, currentKey)

		// Capture action in closure
		act := action
		m.keybindItems = append(m.keybindItems, MenuItem{
			Label: label,
			Action: func() {
				m.startRebind(act)
			},
		})
	}

	// Add Reset All and Back options
	m.keybindItems = append(m.keybindItems,
		MenuItem{Label: "Reset All to Defaults", Action: m.resetAllBindings},
		MenuItem{Label: "Back", Action: m.closeKeybindings},
	)
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

	// Handle key rebinding input capture
	if m.waitingForKeyInput {
		m.handleKeyCapture()
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

// handleKeyCapture processes key input during rebinding.
func (m *Menu) handleKeyCapture() {
	// Cancel rebinding with Escape
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		m.waitingForKeyInput = false
		m.keyBindingToChange = ""
		return
	}

	// Check for any key press
	keys := inpututil.AppendJustPressedKeys(nil)
	if len(keys) == 0 {
		return
	}

	// Get the first pressed key and convert to our key name format
	key := keys[0]
	keyName := ebitenKeyToName(key)
	if keyName == "" {
		return // Unknown key
	}

	// Validate the key
	if !input.ValidateKey(keyName) {
		return // Invalid key for binding
	}

	// Apply the new binding
	action := input.Action(m.keyBindingToChange)
	if err := m.inputManager.SetBinding(action, keyName); err == nil {
		// Update config
		m.cfg.KeyBindings = *m.inputManager.ExportToConfig()
	}

	// Rebuild the menu and exit capture mode
	m.buildKeybindMenu()
	m.waitingForKeyInput = false
	m.keyBindingToChange = ""
}

// ebitenKeyToName converts an ebiten.Key to our key name format.
func ebitenKeyToName(key ebiten.Key) string {
	keyNames := map[ebiten.Key]string{
		// Letters
		ebiten.KeyA: "A", ebiten.KeyB: "B", ebiten.KeyC: "C", ebiten.KeyD: "D",
		ebiten.KeyE: "E", ebiten.KeyF: "F", ebiten.KeyG: "G", ebiten.KeyH: "H",
		ebiten.KeyI: "I", ebiten.KeyJ: "J", ebiten.KeyK: "K", ebiten.KeyL: "L",
		ebiten.KeyM: "M", ebiten.KeyN: "N", ebiten.KeyO: "O", ebiten.KeyP: "P",
		ebiten.KeyQ: "Q", ebiten.KeyR: "R", ebiten.KeyS: "S", ebiten.KeyT: "T",
		ebiten.KeyU: "U", ebiten.KeyV: "V", ebiten.KeyW: "W", ebiten.KeyX: "X",
		ebiten.KeyY: "Y", ebiten.KeyZ: "Z",
		// Numbers
		ebiten.KeyDigit0: "0", ebiten.KeyDigit1: "1", ebiten.KeyDigit2: "2",
		ebiten.KeyDigit3: "3", ebiten.KeyDigit4: "4", ebiten.KeyDigit5: "5",
		ebiten.KeyDigit6: "6", ebiten.KeyDigit7: "7", ebiten.KeyDigit8: "8",
		ebiten.KeyDigit9: "9",
		// Function keys
		ebiten.KeyF1: "F1", ebiten.KeyF2: "F2", ebiten.KeyF3: "F3", ebiten.KeyF4: "F4",
		ebiten.KeyF5: "F5", ebiten.KeyF6: "F6", ebiten.KeyF7: "F7", ebiten.KeyF8: "F8",
		ebiten.KeyF9: "F9", ebiten.KeyF10: "F10", ebiten.KeyF11: "F11", ebiten.KeyF12: "F12",
		// Special keys
		ebiten.KeySpace:     "Space",
		ebiten.KeyTab:       "Tab",
		ebiten.KeyEnter:     "Enter",
		ebiten.KeyEscape:    "Escape",
		ebiten.KeyBackspace: "Backspace",
		ebiten.KeyDelete:    "Delete",
		ebiten.KeyInsert:    "Insert",
		ebiten.KeyHome:      "Home",
		ebiten.KeyEnd:       "End",
		ebiten.KeyPageUp:    "PageUp",
		ebiten.KeyPageDown:  "PageDown",
		ebiten.KeyBackquote: "Backquote",
		// Modifiers
		ebiten.KeyShiftLeft:    "ShiftLeft",
		ebiten.KeyShiftRight:   "ShiftRight",
		ebiten.KeyControlLeft:  "ControlLeft",
		ebiten.KeyControlRight: "ControlRight",
		ebiten.KeyAltLeft:      "AltLeft",
		ebiten.KeyAltRight:     "AltRight",
		// Arrows
		ebiten.KeyArrowUp:    "ArrowUp",
		ebiten.KeyArrowDown:  "ArrowDown",
		ebiten.KeyArrowLeft:  "ArrowLeft",
		ebiten.KeyArrowRight: "ArrowRight",
	}

	if name, ok := keyNames[key]; ok {
		return name
	}
	return ""
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
	switch m.state {
	case MenuStateSettings:
		return m.settingsItems
	case MenuStateQuitConfirm:
		return m.quitConfirmItems
	case MenuStateKeyBindings:
		return m.keybindItems
	default:
		return m.items
	}
}

// getCurrentIndex returns the current selection index.
func (m *Menu) getCurrentIndex() int {
	switch m.state {
	case MenuStateSettings:
		return m.settingsIndex
	case MenuStateQuitConfirm:
		return m.quitConfirmIndex
	case MenuStateKeyBindings:
		return m.keybindIndex
	default:
		return m.selectedIndex
	}
}

// setCurrentIndex sets the current selection index.
func (m *Menu) setCurrentIndex(index int) {
	switch m.state {
	case MenuStateSettings:
		m.settingsIndex = index
	case MenuStateQuitConfirm:
		m.quitConfirmIndex = index
	case MenuStateKeyBindings:
		m.keybindIndex = index
	default:
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
	case MenuStateQuitConfirm:
		m.cancelQuit()
	case MenuStateKeyBindings:
		m.closeKeybindings()
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

func (m *Menu) saveGame() {
	if m.onSaveRequest == nil {
		m.saveMessage = "Save not available"
		return
	}
	if err := m.onSaveRequest(); err != nil {
		m.saveMessage = fmt.Sprintf("Save failed: %v", err)
	} else {
		m.saveMessage = "Save requested..."
	}
}

func (m *Menu) loadGame() {
	if m.onLoadRequest == nil {
		m.saveMessage = "Load not available"
		return
	}
	if err := m.onLoadRequest(); err != nil {
		m.saveMessage = fmt.Sprintf("Load failed: %v", err)
	} else {
		m.saveMessage = "Load requested..."
	}
}

func (m *Menu) openSettings() {
	m.state = MenuStateSettings
	m.settingsIndex = 0
}

func (m *Menu) closeSettings() {
	m.state = MenuStatePause
	m.buildSettingsMenu() // Rebuild to update labels
}

func (m *Menu) saveSettings() {
	// Update config with current settings
	m.cfg.Audio.MasterVolume = m.volumeLevel
	m.cfg.Audio.MusicEnabled = m.musicEnabled
	m.cfg.Audio.SFXEnabled = m.sfxEnabled
	m.cfg.Window.Fullscreen = m.fullscreen
	m.cfg.Window.ShowFPS = m.showFPS

	// Save to config file
	if err := m.cfg.Save("config.yaml"); err != nil {
		// Log error but don't crash - settings are still applied in memory
		fmt.Printf("Warning: failed to save settings: %v\n", err)
	}
}

func (m *Menu) showQuitConfirm() {
	m.state = MenuStateQuitConfirm
	m.quitConfirmIndex = 1 // Default to "No" for safety
}

func (m *Menu) confirmQuit() {
	m.quitRequested = true
}

func (m *Menu) cancelQuit() {
	m.state = MenuStatePause
	m.quitConfirmIndex = 0
}

func (m *Menu) cycleVolume() {
	m.volumeLevel = (m.volumeLevel + 1) % 11
	m.cfg.Audio.MasterVolume = m.volumeLevel
	m.buildSettingsMenu()
}

func (m *Menu) toggleMusic() {
	m.musicEnabled = !m.musicEnabled
	m.cfg.Audio.MusicEnabled = m.musicEnabled
	m.buildSettingsMenu()
}

func (m *Menu) toggleSFX() {
	m.sfxEnabled = !m.sfxEnabled
	m.cfg.Audio.SFXEnabled = m.sfxEnabled
	m.buildSettingsMenu()
}

func (m *Menu) toggleFullscreen() {
	m.fullscreen = !m.fullscreen
	m.cfg.Window.Fullscreen = m.fullscreen
	ebiten.SetFullscreen(m.fullscreen)
	m.buildSettingsMenu()
}

func (m *Menu) toggleFPS() {
	m.showFPS = !m.showFPS
	m.cfg.Window.ShowFPS = m.showFPS
	m.buildSettingsMenu()
}

func (m *Menu) openKeybindings() {
	m.state = MenuStateKeyBindings
	m.keybindIndex = 0
	m.keybindScrollOffset = 0
	m.buildKeybindMenu()
}

func (m *Menu) closeKeybindings() {
	m.state = MenuStateSettings
	m.waitingForKeyInput = false
	m.keyBindingToChange = ""
}

func (m *Menu) startRebind(action input.Action) {
	m.waitingForKeyInput = true
	m.keyBindingToChange = string(action)
}

func (m *Menu) resetAllBindings() {
	m.inputManager.ResetAllBindings()
	m.buildKeybindMenu()
	// Update config with reset bindings
	m.cfg.KeyBindings = *m.inputManager.ExportToConfig()
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

	// Special handling for waiting for key input
	if m.waitingForKeyInput {
		msg := "Press a key to bind, or ESC to cancel"
		ebitenutil.DebugPrintAt(screen, msg, centerX-len(msg)*3, centerY)
		actionName := actionDisplayNames[input.Action(m.keyBindingToChange)]
		if actionName == "" {
			actionName = m.keyBindingToChange
		}
		bindMsg := fmt.Sprintf("Rebinding: %s", actionName)
		ebitenutil.DebugPrintAt(screen, bindMsg, centerX-len(bindMsg)*3, centerY+25)
		return
	}

	// Draw menu items (with scrolling for keybindings)
	items := m.getCurrentItems()
	index := m.getCurrentIndex()

	if m.state == MenuStateKeyBindings {
		m.drawKeybindMenu(screen, items, index, centerX, centerY, h)
	} else {
		m.drawStandardMenu(screen, items, index, centerX, centerY, h)
	}
}

// drawStandardMenu draws a standard menu (pause, settings, quit confirm).
func (m *Menu) drawStandardMenu(screen *ebiten.Image, items []MenuItem, index, centerX, centerY, h int) {
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

	// Draw save/load status message if present
	if m.state == MenuStatePause && m.saveMessage != "" {
		msgY := startY + len(items)*25 + 15
		ebitenutil.DebugPrintAt(screen, m.saveMessage, centerX-len(m.saveMessage)*3, msgY)
	}

	// Draw navigation hints
	hints := "UP/DOWN: Navigate | ENTER: Select | ESC: Back"
	if m.state == MenuStateSettings && m.settingsIndex == 0 {
		hints = "LEFT/RIGHT: Adjust | " + hints
	}
	ebitenutil.DebugPrintAt(screen, hints, centerX-len(hints)*3, h-30)
}

// drawKeybindMenu draws the keybinding menu with scrolling support.
func (m *Menu) drawKeybindMenu(screen *ebiten.Image, items []MenuItem, index, centerX, centerY, h int) {
	maxVisible := 15 // Max items visible at once

	// Calculate scroll offset to keep selection visible
	if index < m.keybindScrollOffset {
		m.keybindScrollOffset = index
	}
	if index >= m.keybindScrollOffset+maxVisible {
		m.keybindScrollOffset = index - maxVisible + 1
	}

	startY := centerY - 150
	visibleEnd := m.keybindScrollOffset + maxVisible
	if visibleEnd > len(items) {
		visibleEnd = len(items)
	}

	for i := m.keybindScrollOffset; i < visibleEnd; i++ {
		item := items[i]
		label := item.Label
		if i == index {
			label = "> " + label + " <"
		}
		if item.Disabled {
			label = "  " + item.Label + " (disabled)"
		}
		x := centerX - len(label)*3
		displayIndex := i - m.keybindScrollOffset
		y := startY + displayIndex*20
		ebitenutil.DebugPrintAt(screen, label, x, y)
	}

	// Draw scroll indicators
	if m.keybindScrollOffset > 0 {
		ebitenutil.DebugPrintAt(screen, "^ More above ^", centerX-42, startY-20)
	}
	if visibleEnd < len(items) {
		ebitenutil.DebugPrintAt(screen, "v More below v", centerX-42, startY+maxVisible*20)
	}

	// Draw navigation hints
	hints := "UP/DOWN: Navigate | ENTER: Rebind | ESC: Back"
	ebitenutil.DebugPrintAt(screen, hints, centerX-len(hints)*3, h-30)
}

// getTitle returns the title for the current menu state.
func (m *Menu) getTitle() string {
	switch m.state {
	case MenuStatePause:
		return "=== PAUSED ==="
	case MenuStateSettings:
		return "=== SETTINGS ==="
	case MenuStateQuitConfirm:
		return "=== QUIT GAME? ==="
	case MenuStateKeyBindings:
		return "=== KEY BINDINGS ==="
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
