package input

import (
	"sync"
	"testing"

	"github.com/opd-ai/wyrm/config"
)

func TestNewManager(t *testing.T) {
	im := NewManager()
	if im == nil {
		t.Fatal("NewManager returned nil")
	}

	// Check that default bindings are set
	if im.GetBinding(ActionMoveForward) != "W" {
		t.Errorf("expected MoveForward=W, got %s", im.GetBinding(ActionMoveForward))
	}
	if im.GetBinding(ActionJump) != "SPACE" {
		t.Errorf("expected Jump=SPACE, got %s", im.GetBinding(ActionJump))
	}
	if im.GetBinding(ActionPause) != "ESCAPE" {
		t.Errorf("expected Pause=ESCAPE, got %s", im.GetBinding(ActionPause))
	}
}

func TestSetBinding(t *testing.T) {
	im := NewManager()

	// Change a binding
	err := im.SetBinding(ActionMoveForward, "UpArrow")
	if err != nil {
		t.Fatalf("SetBinding failed: %v", err)
	}

	// Verify new binding
	if im.GetBinding(ActionMoveForward) != "UPARROW" {
		t.Errorf("expected MoveForward=UPARROW, got %s", im.GetBinding(ActionMoveForward))
	}

	// Verify old binding is removed
	actions := im.GetActionsForKey("W")
	for _, a := range actions {
		if a == ActionMoveForward {
			t.Error("old binding W should not map to MoveForward")
		}
	}
}

func TestGetActionsForKey(t *testing.T) {
	im := NewManager()

	// Check actions for W key
	actions := im.GetActionsForKey("W")
	found := false
	for _, a := range actions {
		if a == ActionMoveForward {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected W to map to MoveForward")
	}

	// Check case insensitivity
	actions = im.GetActionsForKey("w")
	found = false
	for _, a := range actions {
		if a == ActionMoveForward {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected lowercase w to map to MoveForward")
	}
}

func TestKeyPressRelease(t *testing.T) {
	im := NewManager()

	// Press W key
	im.OnKeyPressed("W")
	if !im.IsActionPressed(ActionMoveForward) {
		t.Error("ActionMoveForward should be pressed after W pressed")
	}
	if !im.IsKeyPressed("W") {
		t.Error("W should be pressed")
	}

	// Release W key
	im.OnKeyReleased("W")
	if im.IsActionPressed(ActionMoveForward) {
		t.Error("ActionMoveForward should not be pressed after W released")
	}
	if im.IsKeyPressed("W") {
		t.Error("W should not be pressed")
	}
}

type testListener struct {
	pressed  []Action
	released []Action
	mu       sync.Mutex
}

func (l *testListener) OnActionPressed(action Action) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pressed = append(l.pressed, action)
}

func (l *testListener) OnActionReleased(action Action) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.released = append(l.released, action)
}

func TestListener(t *testing.T) {
	im := NewManager()
	listener := &testListener{}
	im.AddListener(listener)

	// Press and release W
	im.OnKeyPressed("W")
	im.OnKeyReleased("W")

	listener.mu.Lock()
	defer listener.mu.Unlock()

	if len(listener.pressed) != 1 || listener.pressed[0] != ActionMoveForward {
		t.Errorf("expected single press of MoveForward, got %v", listener.pressed)
	}
	if len(listener.released) != 1 || listener.released[0] != ActionMoveForward {
		t.Errorf("expected single release of MoveForward, got %v", listener.released)
	}
}

func TestRemoveListener(t *testing.T) {
	im := NewManager()
	listener := &testListener{}
	im.AddListener(listener)
	im.RemoveListener(listener)

	// Press W - listener should not receive
	im.OnKeyPressed("W")

	listener.mu.Lock()
	defer listener.mu.Unlock()

	if len(listener.pressed) != 0 {
		t.Error("removed listener should not receive events")
	}
}

func TestLoadFromConfig(t *testing.T) {
	im := NewManager()

	cfg := &config.KeyBindingsConfig{
		MoveForward:  "UpArrow",
		MoveBackward: "DownArrow",
		MoveLeft:     "LeftArrow",
		MoveRight:    "RightArrow",
		Jump:         "Z",
		Crouch:       "X",
		Sprint:       "C",

		Attack:       "MouseButtonLeft",
		Block:        "MouseButtonRight",
		UseAbility1:  "1",
		UseAbility2:  "2",
		UseAbility3:  "3",
		UseAbility4:  "4",
		QuickHeal:    "Q",
		ToggleWeapon: "Tab",

		Interact:     "E",
		PickUp:       "F",
		DropItem:     "G",
		UseItem:      "R",
		Talk:         "T",
		ReadSign:     "V",
		Mount:        "M",
		EnterVehicle: "N",

		Inventory:   "I",
		Map:         "Tab",
		QuestLog:    "J",
		CharSheet:   "K",
		SkillTree:   "P",
		Crafting:    "B",
		Pause:       "Escape",
		QuickSave:   "F5",
		QuickLoad:   "F9",
		Screenshot:  "F12",
		ToggleHUD:   "F1",
		Console:     "Backquote",
		ChatWindow:  "Enter",
		SocialMenu:  "O",
		TradeWindow: "Y",
	}

	im.LoadFromConfig(cfg)

	if im.GetBinding(ActionMoveForward) != "UPARROW" {
		t.Errorf("expected MoveForward=UPARROW, got %s", im.GetBinding(ActionMoveForward))
	}
	if im.GetBinding(ActionJump) != "Z" {
		t.Errorf("expected Jump=Z, got %s", im.GetBinding(ActionJump))
	}
}

func TestExportToConfig(t *testing.T) {
	im := NewManager()

	// Change some bindings
	im.SetBinding(ActionMoveForward, "Up")
	im.SetBinding(ActionJump, "Z")

	cfg := im.ExportToConfig()

	if cfg.MoveForward != "UP" {
		t.Errorf("expected MoveForward=UP, got %s", cfg.MoveForward)
	}
	if cfg.Jump != "Z" {
		t.Errorf("expected Jump=Z, got %s", cfg.Jump)
	}
}

func TestResetBinding(t *testing.T) {
	im := NewManager()

	// Change binding
	im.SetBinding(ActionMoveForward, "Up")
	if im.GetBinding(ActionMoveForward) != "UP" {
		t.Errorf("expected UP, got %s", im.GetBinding(ActionMoveForward))
	}

	// Reset to default
	err := im.ResetBinding(ActionMoveForward)
	if err != nil {
		t.Fatalf("ResetBinding failed: %v", err)
	}
	if im.GetBinding(ActionMoveForward) != "W" {
		t.Errorf("expected W after reset, got %s", im.GetBinding(ActionMoveForward))
	}
}

func TestResetAllBindings(t *testing.T) {
	im := NewManager()

	// Change several bindings
	im.SetBinding(ActionMoveForward, "Up")
	im.SetBinding(ActionJump, "Z")
	im.SetBinding(ActionPause, "P")

	// Reset all
	im.ResetAllBindings()

	if im.GetBinding(ActionMoveForward) != "W" {
		t.Errorf("expected W, got %s", im.GetBinding(ActionMoveForward))
	}
	if im.GetBinding(ActionJump) != "SPACE" {
		t.Errorf("expected SPACE, got %s", im.GetBinding(ActionJump))
	}
	if im.GetBinding(ActionPause) != "ESCAPE" {
		t.Errorf("expected ESCAPE, got %s", im.GetBinding(ActionPause))
	}
}

func TestValidateKey(t *testing.T) {
	tests := []struct {
		key   string
		valid bool
	}{
		{"W", true},
		{"w", true},
		{"Space", true},
		{"SPACE", true},
		{"F1", true},
		{"F12", true},
		{"MouseButtonLeft", true},
		{"ShiftLeft", true},
		{"Invalid", false},
		{"F13", false},
		{"", false},
	}

	for _, tc := range tests {
		result := ValidateKey(tc.key)
		if result != tc.valid {
			t.Errorf("ValidateKey(%q) = %v, want %v", tc.key, result, tc.valid)
		}
	}
}

func TestGetAllBindings(t *testing.T) {
	im := NewManager()

	bindings := im.GetAllBindings()

	// Check some expected bindings
	if bindings[ActionMoveForward] != "W" {
		t.Errorf("expected MoveForward=W, got %s", bindings[ActionMoveForward])
	}
	if bindings[ActionJump] != "SPACE" {
		t.Errorf("expected Jump=SPACE, got %s", bindings[ActionJump])
	}

	// Modify returned map should not affect original
	bindings[ActionMoveForward] = "UP"
	if im.GetBinding(ActionMoveForward) != "W" {
		t.Error("modifying returned map should not affect original bindings")
	}
}

func TestKeyRepeatIgnored(t *testing.T) {
	im := NewManager()
	listener := &testListener{}
	im.AddListener(listener)

	// Press W multiple times
	im.OnKeyPressed("W")
	im.OnKeyPressed("W")
	im.OnKeyPressed("W")

	listener.mu.Lock()
	defer listener.mu.Unlock()

	// Should only fire once
	if len(listener.pressed) != 1 {
		t.Errorf("expected 1 press event, got %d", len(listener.pressed))
	}
}

func TestConcurrentAccess(t *testing.T) {
	im := NewManager()

	var wg sync.WaitGroup

	// Multiple goroutines pressing and releasing keys
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				im.OnKeyPressed("W")
				im.IsActionPressed(ActionMoveForward)
				im.OnKeyReleased("W")
				im.GetBinding(ActionMoveForward)
				im.GetAllBindings()
			}
		}()
	}

	// Another goroutine changing bindings
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 50; j++ {
			im.SetBinding(ActionMoveForward, "Up")
			im.SetBinding(ActionMoveForward, "W")
		}
	}()

	wg.Wait()
}

func TestMultipleActionsOnSameKey(t *testing.T) {
	im := NewManager()

	// Bind two actions to the same key
	im.SetBinding(ActionMoveForward, "W")
	im.SetBinding(ActionSprint, "W")

	// Press W
	im.OnKeyPressed("W")

	// Both actions should be pressed
	if !im.IsActionPressed(ActionMoveForward) {
		t.Error("ActionMoveForward should be pressed")
	}
	if !im.IsActionPressed(ActionSprint) {
		t.Error("ActionSprint should be pressed")
	}
}
