// Package input provides key rebinding and input handling for Wyrm.
package input

import (
	"fmt"
	"strings"
	"sync"

	"github.com/opd-ai/wyrm/config"
)

// Action represents a game action that can be bound to a key.
type Action string

// Movement actions.
const (
	ActionMoveForward  Action = "move_forward"
	ActionMoveBackward Action = "move_backward"
	ActionMoveLeft     Action = "move_left"
	ActionMoveRight    Action = "move_right"
	ActionJump         Action = "jump"
	ActionCrouch       Action = "crouch"
	ActionSprint       Action = "sprint"
)

// Combat actions.
const (
	ActionAttack       Action = "attack"
	ActionBlock        Action = "block"
	ActionCastSpell    Action = "cast_spell"
	ActionAbility1     Action = "ability_1"
	ActionAbility2     Action = "ability_2"
	ActionAbility3     Action = "ability_3"
	ActionAbility4     Action = "ability_4"
	ActionQuickHeal    Action = "quick_heal"
	ActionToggleWeapon Action = "toggle_weapon"
)

// Interaction actions.
const (
	ActionInteract     Action = "interact"
	ActionPickUp       Action = "pick_up"
	ActionDropItem     Action = "drop_item"
	ActionUseItem      Action = "use_item"
	ActionTalk         Action = "talk"
	ActionReadSign     Action = "read_sign"
	ActionMount        Action = "mount"
	ActionEnterVehicle Action = "enter_vehicle"
)

// UI actions.
const (
	ActionInventory   Action = "inventory"
	ActionMap         Action = "map"
	ActionQuestLog    Action = "quest_log"
	ActionCharSheet   Action = "character_sheet"
	ActionSkillTree   Action = "skill_tree"
	ActionCrafting    Action = "crafting"
	ActionFactions    Action = "factions"
	ActionPause       Action = "pause"
	ActionQuickSave   Action = "quick_save"
	ActionQuickLoad   Action = "quick_load"
	ActionScreenshot  Action = "screenshot"
	ActionToggleHUD   Action = "toggle_hud"
	ActionConsole     Action = "console"
	ActionChatWindow  Action = "chat_window"
	ActionSocialMenu  Action = "social_menu"
	ActionTradeWindow Action = "trade_window"
)

// Manager handles key bindings and input state.
type Manager struct {
	mu           sync.RWMutex
	bindings     map[Action]string   // Action -> key name
	keyToActions map[string][]Action // key name -> actions (for reverse lookup)
	pressedKeys  map[string]bool     // currently pressed keys
	listeners    []Listener
}

// Listener receives input events.
type Listener interface {
	OnActionPressed(action Action)
	OnActionReleased(action Action)
}

// NewManager creates a new input manager with default bindings.
func NewManager() *Manager {
	im := &Manager{
		bindings:     make(map[Action]string),
		keyToActions: make(map[string][]Action),
		pressedKeys:  make(map[string]bool),
		listeners:    make([]Listener, 0),
	}
	im.setDefaultBindings()
	return im
}

// LoadFromConfig loads key bindings from a config struct.
func (im *Manager) LoadFromConfig(cfg *config.KeyBindingsConfig) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Clear existing bindings
	im.bindings = make(map[Action]string)
	im.keyToActions = make(map[string][]Action)

	// Apply bindings from config using the binding table
	bindings := im.getConfigBindings(cfg)
	for _, b := range bindings {
		im.setBindingUnsafe(b.action, b.key)
	}
}

// actionBinding pairs an action with its key.
type actionBinding struct {
	action Action
	key    string
}

// getConfigBindings returns all bindings from the config.
func (im *Manager) getConfigBindings(cfg *config.KeyBindingsConfig) []actionBinding {
	return []actionBinding{
		// Movement
		{ActionMoveForward, cfg.MoveForward},
		{ActionMoveBackward, cfg.MoveBackward},
		{ActionMoveLeft, cfg.MoveLeft},
		{ActionMoveRight, cfg.MoveRight},
		{ActionJump, cfg.Jump},
		{ActionCrouch, cfg.Crouch},
		{ActionSprint, cfg.Sprint},
		// Combat
		{ActionAttack, cfg.Attack},
		{ActionBlock, cfg.Block},
		{ActionAbility1, cfg.UseAbility1},
		{ActionAbility2, cfg.UseAbility2},
		{ActionAbility3, cfg.UseAbility3},
		{ActionAbility4, cfg.UseAbility4},
		{ActionQuickHeal, cfg.QuickHeal},
		{ActionToggleWeapon, cfg.ToggleWeapon},
		// Interaction
		{ActionInteract, cfg.Interact},
		{ActionPickUp, cfg.PickUp},
		{ActionDropItem, cfg.DropItem},
		{ActionUseItem, cfg.UseItem},
		{ActionTalk, cfg.Talk},
		{ActionReadSign, cfg.ReadSign},
		{ActionMount, cfg.Mount},
		{ActionEnterVehicle, cfg.EnterVehicle},
		// UI
		{ActionInventory, cfg.Inventory},
		{ActionMap, cfg.Map},
		{ActionQuestLog, cfg.QuestLog},
		{ActionCharSheet, cfg.CharSheet},
		{ActionSkillTree, cfg.SkillTree},
		{ActionCrafting, cfg.Crafting},
		{ActionPause, cfg.Pause},
		{ActionQuickSave, cfg.QuickSave},
		{ActionQuickLoad, cfg.QuickLoad},
		{ActionScreenshot, cfg.Screenshot},
		{ActionToggleHUD, cfg.ToggleHUD},
		{ActionConsole, cfg.Console},
		{ActionChatWindow, cfg.ChatWindow},
		{ActionSocialMenu, cfg.SocialMenu},
		{ActionTradeWindow, cfg.TradeWindow},
	}
}

// setDefaultBindings sets up the default key bindings.
func (im *Manager) setDefaultBindings() {
	im.mu.Lock()
	defer im.mu.Unlock()

	for _, b := range defaultBindings {
		im.setBindingUnsafe(b.action, b.key)
	}
}

// defaultBindings defines the default key bindings for all actions.
var defaultBindings = []actionBinding{
	// Movement
	{ActionMoveForward, "W"},
	{ActionMoveBackward, "S"},
	{ActionMoveLeft, "A"},
	{ActionMoveRight, "D"},
	{ActionJump, "Space"},
	{ActionCrouch, "ControlLeft"},
	{ActionSprint, "ShiftLeft"},
	// Combat
	{ActionAttack, "MouseButtonLeft"},
	{ActionBlock, "MouseButtonRight"},
	{ActionCastSpell, "Q"},
	{ActionAbility1, "1"},
	{ActionAbility2, "2"},
	{ActionAbility3, "3"},
	{ActionAbility4, "4"},
	{ActionQuickHeal, "H"},
	{ActionToggleWeapon, "Tab"},
	// Interaction
	{ActionInteract, "E"},
	{ActionPickUp, "F"},
	{ActionDropItem, "G"},
	{ActionUseItem, "R"},
	{ActionTalk, "T"},
	{ActionReadSign, "V"},
	{ActionMount, "X"},
	{ActionEnterVehicle, "C"},
	// UI
	{ActionInventory, "I"},
	{ActionMap, "M"},
	{ActionQuestLog, "J"},
	{ActionCharSheet, "K"},
	{ActionSkillTree, "P"},
	{ActionCrafting, "B"},
	{ActionPause, "Escape"},
	{ActionQuickSave, "F5"},
	{ActionQuickLoad, "F9"},
	{ActionScreenshot, "F12"},
	{ActionToggleHUD, "F1"},
	{ActionConsole, "Backquote"},
	{ActionChatWindow, "Enter"},
	{ActionSocialMenu, "O"},
	{ActionTradeWindow, "Y"},
}

// setBindingUnsafe sets a binding without acquiring the lock.
func (im *Manager) setBindingUnsafe(action Action, key string) {
	key = strings.ToUpper(key)
	im.bindings[action] = key
	im.keyToActions[key] = append(im.keyToActions[key], action)
}

// SetBinding changes the key binding for an action.
func (im *Manager) SetBinding(action Action, key string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	key = strings.ToUpper(key)

	// Remove old binding
	if oldKey, exists := im.bindings[action]; exists {
		im.removeKeyActionUnsafe(oldKey, action)
	}

	// Set new binding
	im.bindings[action] = key
	im.keyToActions[key] = append(im.keyToActions[key], action)
	return nil
}

// removeKeyActionUnsafe removes an action from a key's action list.
func (im *Manager) removeKeyActionUnsafe(key string, action Action) {
	actions := im.keyToActions[key]
	for i, a := range actions {
		if a == action {
			im.keyToActions[key] = append(actions[:i], actions[i+1:]...)
			break
		}
	}
	if len(im.keyToActions[key]) == 0 {
		delete(im.keyToActions, key)
	}
}

// GetBinding returns the key bound to an action.
func (im *Manager) GetBinding(action Action) string {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.bindings[action]
}

// GetAllBindings returns a copy of all current bindings.
func (im *Manager) GetAllBindings() map[Action]string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	result := make(map[Action]string, len(im.bindings))
	for action, key := range im.bindings {
		result[action] = key
	}
	return result
}

// GetActionsForKey returns all actions bound to a key.
func (im *Manager) GetActionsForKey(key string) []Action {
	im.mu.RLock()
	defer im.mu.RUnlock()

	key = strings.ToUpper(key)
	if actions, exists := im.keyToActions[key]; exists {
		result := make([]Action, len(actions))
		copy(result, actions)
		return result
	}
	return nil
}

// AddListener registers an input listener.
func (im *Manager) AddListener(listener Listener) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.listeners = append(im.listeners, listener)
}

// RemoveListener unregisters an input listener.
func (im *Manager) RemoveListener(listener Listener) {
	im.mu.Lock()
	defer im.mu.Unlock()

	for i, l := range im.listeners {
		if l == listener {
			im.listeners = append(im.listeners[:i], im.listeners[i+1:]...)
			return
		}
	}
}

// OnKeyPressed should be called when a key is pressed.
func (im *Manager) OnKeyPressed(key string) {
	im.mu.Lock()
	key = strings.ToUpper(key)
	if im.pressedKeys[key] {
		im.mu.Unlock()
		return // Already pressed, ignore repeat
	}
	im.pressedKeys[key] = true
	actions := make([]Action, len(im.keyToActions[key]))
	copy(actions, im.keyToActions[key])
	listeners := make([]Listener, len(im.listeners))
	copy(listeners, im.listeners)
	im.mu.Unlock()

	// Notify listeners outside the lock
	for _, action := range actions {
		for _, listener := range listeners {
			listener.OnActionPressed(action)
		}
	}
}

// OnKeyReleased should be called when a key is released.
func (im *Manager) OnKeyReleased(key string) {
	im.mu.Lock()
	key = strings.ToUpper(key)
	if !im.pressedKeys[key] {
		im.mu.Unlock()
		return // Wasn't pressed
	}
	delete(im.pressedKeys, key)
	actions := make([]Action, len(im.keyToActions[key]))
	copy(actions, im.keyToActions[key])
	listeners := make([]Listener, len(im.listeners))
	copy(listeners, im.listeners)
	im.mu.Unlock()

	// Notify listeners outside the lock
	for _, action := range actions {
		for _, listener := range listeners {
			listener.OnActionReleased(action)
		}
	}
}

// IsActionPressed returns true if the action's key is currently pressed.
func (im *Manager) IsActionPressed(action Action) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	key, exists := im.bindings[action]
	if !exists {
		return false
	}
	return im.pressedKeys[key]
}

// IsKeyPressed returns true if the key is currently pressed.
func (im *Manager) IsKeyPressed(key string) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.pressedKeys[strings.ToUpper(key)]
}

// ResetBinding restores the default binding for an action.
func (im *Manager) ResetBinding(action Action) error {
	defaultKey, err := im.getDefaultKey(action)
	if err != nil {
		return err
	}
	return im.SetBinding(action, defaultKey)
}

// ResetAllBindings restores all bindings to defaults.
func (im *Manager) ResetAllBindings() {
	im.setDefaultBindings()
}

// getDefaultKey returns the default key for an action.
func (im *Manager) getDefaultKey(action Action) (string, error) {
	defaults := map[Action]string{
		ActionMoveForward:  "W",
		ActionMoveBackward: "S",
		ActionMoveLeft:     "A",
		ActionMoveRight:    "D",
		ActionJump:         "Space",
		ActionCrouch:       "ControlLeft",
		ActionSprint:       "ShiftLeft",
		ActionAttack:       "MouseButtonLeft",
		ActionBlock:        "MouseButtonRight",
		ActionAbility1:     "1",
		ActionAbility2:     "2",
		ActionAbility3:     "3",
		ActionAbility4:     "4",
		ActionQuickHeal:    "H",
		ActionToggleWeapon: "Tab",
		ActionInteract:     "E",
		ActionPickUp:       "F",
		ActionDropItem:     "G",
		ActionUseItem:      "R",
		ActionTalk:         "T",
		ActionReadSign:     "V",
		ActionMount:        "X",
		ActionEnterVehicle: "C",
		ActionInventory:    "I",
		ActionMap:          "M",
		ActionQuestLog:     "J",
		ActionCharSheet:    "K",
		ActionSkillTree:    "P",
		ActionCrafting:     "B",
		ActionPause:        "Escape",
		ActionQuickSave:    "F5",
		ActionQuickLoad:    "F9",
		ActionScreenshot:   "F12",
		ActionToggleHUD:    "F1",
		ActionConsole:      "Backquote",
		ActionChatWindow:   "Enter",
		ActionSocialMenu:   "O",
		ActionTradeWindow:  "Y",
	}

	if key, exists := defaults[action]; exists {
		return key, nil
	}
	return "", fmt.Errorf("unknown action: %s", action)
}

// ExportToConfig exports current bindings to a KeyBindingsConfig.
func (im *Manager) ExportToConfig() *config.KeyBindingsConfig {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return &config.KeyBindingsConfig{
		MoveForward:  im.bindings[ActionMoveForward],
		MoveBackward: im.bindings[ActionMoveBackward],
		MoveLeft:     im.bindings[ActionMoveLeft],
		MoveRight:    im.bindings[ActionMoveRight],
		Jump:         im.bindings[ActionJump],
		Crouch:       im.bindings[ActionCrouch],
		Sprint:       im.bindings[ActionSprint],

		Attack:       im.bindings[ActionAttack],
		Block:        im.bindings[ActionBlock],
		UseAbility1:  im.bindings[ActionAbility1],
		UseAbility2:  im.bindings[ActionAbility2],
		UseAbility3:  im.bindings[ActionAbility3],
		UseAbility4:  im.bindings[ActionAbility4],
		QuickHeal:    im.bindings[ActionQuickHeal],
		ToggleWeapon: im.bindings[ActionToggleWeapon],

		Interact:     im.bindings[ActionInteract],
		PickUp:       im.bindings[ActionPickUp],
		DropItem:     im.bindings[ActionDropItem],
		UseItem:      im.bindings[ActionUseItem],
		Talk:         im.bindings[ActionTalk],
		ReadSign:     im.bindings[ActionReadSign],
		Mount:        im.bindings[ActionMount],
		EnterVehicle: im.bindings[ActionEnterVehicle],

		Inventory:   im.bindings[ActionInventory],
		Map:         im.bindings[ActionMap],
		QuestLog:    im.bindings[ActionQuestLog],
		CharSheet:   im.bindings[ActionCharSheet],
		SkillTree:   im.bindings[ActionSkillTree],
		Crafting:    im.bindings[ActionCrafting],
		Pause:       im.bindings[ActionPause],
		QuickSave:   im.bindings[ActionQuickSave],
		QuickLoad:   im.bindings[ActionQuickLoad],
		Screenshot:  im.bindings[ActionScreenshot],
		ToggleHUD:   im.bindings[ActionToggleHUD],
		Console:     im.bindings[ActionConsole],
		ChatWindow:  im.bindings[ActionChatWindow],
		SocialMenu:  im.bindings[ActionSocialMenu],
		TradeWindow: im.bindings[ActionTradeWindow],
	}
}

// ValidateKey checks if a key name is valid.
func ValidateKey(key string) bool {
	// Valid keys include:
	// - Single letters A-Z
	// - Numbers 0-9
	// - Function keys F1-F12
	// - Special keys: Space, Tab, Enter, Escape, Backspace, etc.
	// - Modifier keys: ShiftLeft, ShiftRight, ControlLeft, ControlRight, AltLeft, AltRight
	// - Arrow keys: ArrowUp, ArrowDown, ArrowLeft, ArrowRight
	// - Mouse buttons: MouseButtonLeft, MouseButtonRight, MouseButtonMiddle

	key = strings.ToUpper(key)

	validKeys := map[string]bool{
		// Letters
		"A": true, "B": true, "C": true, "D": true, "E": true, "F": true,
		"G": true, "H": true, "I": true, "J": true, "K": true, "L": true,
		"M": true, "N": true, "O": true, "P": true, "Q": true, "R": true,
		"S": true, "T": true, "U": true, "V": true, "W": true, "X": true,
		"Y": true, "Z": true,

		// Numbers
		"0": true, "1": true, "2": true, "3": true, "4": true,
		"5": true, "6": true, "7": true, "8": true, "9": true,

		// Function keys
		"F1": true, "F2": true, "F3": true, "F4": true, "F5": true, "F6": true,
		"F7": true, "F8": true, "F9": true, "F10": true, "F11": true, "F12": true,

		// Special keys
		"SPACE": true, "TAB": true, "ENTER": true, "ESCAPE": true, "BACKSPACE": true,
		"DELETE": true, "INSERT": true, "HOME": true, "END": true,
		"PAGEUP": true, "PAGEDOWN": true, "BACKQUOTE": true,

		// Modifiers
		"SHIFTLEFT": true, "SHIFTRIGHT": true, "CONTROLLEFT": true, "CONTROLRIGHT": true,
		"ALTLEFT": true, "ALTRIGHT": true,

		// Arrows
		"ARROWUP": true, "ARROWDOWN": true, "ARROWLEFT": true, "ARROWRIGHT": true,

		// Mouse buttons
		"MOUSEBUTTONLEFT": true, "MOUSEBUTTONRIGHT": true, "MOUSEBUTTONMIDDLE": true,
	}

	return validKeys[key]
}
