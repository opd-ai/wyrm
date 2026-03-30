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

// InputManager handles key bindings and input state.
type InputManager struct {
	mu           sync.RWMutex
	bindings     map[Action]string   // Action -> key name
	keyToActions map[string][]Action // key name -> actions (for reverse lookup)
	pressedKeys  map[string]bool     // currently pressed keys
	listeners    []InputListener
}

// InputListener receives input events.
type InputListener interface {
	OnActionPressed(action Action)
	OnActionReleased(action Action)
}

// NewInputManager creates a new input manager with default bindings.
func NewInputManager() *InputManager {
	im := &InputManager{
		bindings:     make(map[Action]string),
		keyToActions: make(map[string][]Action),
		pressedKeys:  make(map[string]bool),
		listeners:    make([]InputListener, 0),
	}
	im.setDefaultBindings()
	return im
}

// LoadFromConfig loads key bindings from a config struct.
func (im *InputManager) LoadFromConfig(cfg *config.KeyBindingsConfig) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Clear existing bindings
	im.bindings = make(map[Action]string)
	im.keyToActions = make(map[string][]Action)

	// Movement
	im.setBindingUnsafe(ActionMoveForward, cfg.MoveForward)
	im.setBindingUnsafe(ActionMoveBackward, cfg.MoveBackward)
	im.setBindingUnsafe(ActionMoveLeft, cfg.MoveLeft)
	im.setBindingUnsafe(ActionMoveRight, cfg.MoveRight)
	im.setBindingUnsafe(ActionJump, cfg.Jump)
	im.setBindingUnsafe(ActionCrouch, cfg.Crouch)
	im.setBindingUnsafe(ActionSprint, cfg.Sprint)

	// Combat
	im.setBindingUnsafe(ActionAttack, cfg.Attack)
	im.setBindingUnsafe(ActionBlock, cfg.Block)
	im.setBindingUnsafe(ActionAbility1, cfg.UseAbility1)
	im.setBindingUnsafe(ActionAbility2, cfg.UseAbility2)
	im.setBindingUnsafe(ActionAbility3, cfg.UseAbility3)
	im.setBindingUnsafe(ActionAbility4, cfg.UseAbility4)
	im.setBindingUnsafe(ActionQuickHeal, cfg.QuickHeal)
	im.setBindingUnsafe(ActionToggleWeapon, cfg.ToggleWeapon)

	// Interaction
	im.setBindingUnsafe(ActionInteract, cfg.Interact)
	im.setBindingUnsafe(ActionPickUp, cfg.PickUp)
	im.setBindingUnsafe(ActionDropItem, cfg.DropItem)
	im.setBindingUnsafe(ActionUseItem, cfg.UseItem)
	im.setBindingUnsafe(ActionTalk, cfg.Talk)
	im.setBindingUnsafe(ActionReadSign, cfg.ReadSign)
	im.setBindingUnsafe(ActionMount, cfg.Mount)
	im.setBindingUnsafe(ActionEnterVehicle, cfg.EnterVehicle)

	// UI
	im.setBindingUnsafe(ActionInventory, cfg.Inventory)
	im.setBindingUnsafe(ActionMap, cfg.Map)
	im.setBindingUnsafe(ActionQuestLog, cfg.QuestLog)
	im.setBindingUnsafe(ActionCharSheet, cfg.CharSheet)
	im.setBindingUnsafe(ActionSkillTree, cfg.SkillTree)
	im.setBindingUnsafe(ActionCrafting, cfg.Crafting)
	im.setBindingUnsafe(ActionPause, cfg.Pause)
	im.setBindingUnsafe(ActionQuickSave, cfg.QuickSave)
	im.setBindingUnsafe(ActionQuickLoad, cfg.QuickLoad)
	im.setBindingUnsafe(ActionScreenshot, cfg.Screenshot)
	im.setBindingUnsafe(ActionToggleHUD, cfg.ToggleHUD)
	im.setBindingUnsafe(ActionConsole, cfg.Console)
	im.setBindingUnsafe(ActionChatWindow, cfg.ChatWindow)
	im.setBindingUnsafe(ActionSocialMenu, cfg.SocialMenu)
	im.setBindingUnsafe(ActionTradeWindow, cfg.TradeWindow)
}

// setDefaultBindings sets up the default key bindings.
func (im *InputManager) setDefaultBindings() {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Movement
	im.setBindingUnsafe(ActionMoveForward, "W")
	im.setBindingUnsafe(ActionMoveBackward, "S")
	im.setBindingUnsafe(ActionMoveLeft, "A")
	im.setBindingUnsafe(ActionMoveRight, "D")
	im.setBindingUnsafe(ActionJump, "Space")
	im.setBindingUnsafe(ActionCrouch, "ControlLeft")
	im.setBindingUnsafe(ActionSprint, "ShiftLeft")

	// Combat
	im.setBindingUnsafe(ActionAttack, "MouseButtonLeft")
	im.setBindingUnsafe(ActionBlock, "MouseButtonRight")
	im.setBindingUnsafe(ActionAbility1, "1")
	im.setBindingUnsafe(ActionAbility2, "2")
	im.setBindingUnsafe(ActionAbility3, "3")
	im.setBindingUnsafe(ActionAbility4, "4")
	im.setBindingUnsafe(ActionQuickHeal, "H")
	im.setBindingUnsafe(ActionToggleWeapon, "Tab")

	// Interaction
	im.setBindingUnsafe(ActionInteract, "E")
	im.setBindingUnsafe(ActionPickUp, "F")
	im.setBindingUnsafe(ActionDropItem, "G")
	im.setBindingUnsafe(ActionUseItem, "R")
	im.setBindingUnsafe(ActionTalk, "T")
	im.setBindingUnsafe(ActionReadSign, "V")
	im.setBindingUnsafe(ActionMount, "X")
	im.setBindingUnsafe(ActionEnterVehicle, "C")

	// UI
	im.setBindingUnsafe(ActionInventory, "I")
	im.setBindingUnsafe(ActionMap, "M")
	im.setBindingUnsafe(ActionQuestLog, "J")
	im.setBindingUnsafe(ActionCharSheet, "K")
	im.setBindingUnsafe(ActionSkillTree, "P")
	im.setBindingUnsafe(ActionCrafting, "B")
	im.setBindingUnsafe(ActionPause, "Escape")
	im.setBindingUnsafe(ActionQuickSave, "F5")
	im.setBindingUnsafe(ActionQuickLoad, "F9")
	im.setBindingUnsafe(ActionScreenshot, "F12")
	im.setBindingUnsafe(ActionToggleHUD, "F1")
	im.setBindingUnsafe(ActionConsole, "Backquote")
	im.setBindingUnsafe(ActionChatWindow, "Enter")
	im.setBindingUnsafe(ActionSocialMenu, "O")
	im.setBindingUnsafe(ActionTradeWindow, "Y")
}

// setBindingUnsafe sets a binding without acquiring the lock.
func (im *InputManager) setBindingUnsafe(action Action, key string) {
	key = strings.ToUpper(key)
	im.bindings[action] = key
	im.keyToActions[key] = append(im.keyToActions[key], action)
}

// SetBinding changes the key binding for an action.
func (im *InputManager) SetBinding(action Action, key string) error {
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
func (im *InputManager) removeKeyActionUnsafe(key string, action Action) {
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
func (im *InputManager) GetBinding(action Action) string {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.bindings[action]
}

// GetAllBindings returns a copy of all current bindings.
func (im *InputManager) GetAllBindings() map[Action]string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	result := make(map[Action]string, len(im.bindings))
	for action, key := range im.bindings {
		result[action] = key
	}
	return result
}

// GetActionsForKey returns all actions bound to a key.
func (im *InputManager) GetActionsForKey(key string) []Action {
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
func (im *InputManager) AddListener(listener InputListener) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.listeners = append(im.listeners, listener)
}

// RemoveListener unregisters an input listener.
func (im *InputManager) RemoveListener(listener InputListener) {
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
func (im *InputManager) OnKeyPressed(key string) {
	im.mu.Lock()
	key = strings.ToUpper(key)
	if im.pressedKeys[key] {
		im.mu.Unlock()
		return // Already pressed, ignore repeat
	}
	im.pressedKeys[key] = true
	actions := make([]Action, len(im.keyToActions[key]))
	copy(actions, im.keyToActions[key])
	listeners := make([]InputListener, len(im.listeners))
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
func (im *InputManager) OnKeyReleased(key string) {
	im.mu.Lock()
	key = strings.ToUpper(key)
	if !im.pressedKeys[key] {
		im.mu.Unlock()
		return // Wasn't pressed
	}
	delete(im.pressedKeys, key)
	actions := make([]Action, len(im.keyToActions[key]))
	copy(actions, im.keyToActions[key])
	listeners := make([]InputListener, len(im.listeners))
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
func (im *InputManager) IsActionPressed(action Action) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	key, exists := im.bindings[action]
	if !exists {
		return false
	}
	return im.pressedKeys[key]
}

// IsKeyPressed returns true if the key is currently pressed.
func (im *InputManager) IsKeyPressed(key string) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.pressedKeys[strings.ToUpper(key)]
}

// ResetBinding restores the default binding for an action.
func (im *InputManager) ResetBinding(action Action) error {
	defaultKey, err := im.getDefaultKey(action)
	if err != nil {
		return err
	}
	return im.SetBinding(action, defaultKey)
}

// ResetAllBindings restores all bindings to defaults.
func (im *InputManager) ResetAllBindings() {
	im.setDefaultBindings()
}

// getDefaultKey returns the default key for an action.
func (im *InputManager) getDefaultKey(action Action) (string, error) {
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
func (im *InputManager) ExportToConfig() *config.KeyBindingsConfig {
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
