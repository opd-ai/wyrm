package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// TestNewInteractionSystem verifies that a new InteractionSystem is created correctly.
func TestNewInteractionSystem(t *testing.T) {
	sys := NewInteractionSystem()

	if sys == nil {
		t.Fatal("NewInteractionSystem returned nil")
	}

	if sys.DefaultInteractionRange != 2.5 {
		t.Errorf("DefaultInteractionRange = %f, want 2.5", sys.DefaultInteractionRange)
	}

	if sys.MaxTargetDistance != 10.0 {
		t.Errorf("MaxTargetDistance = %f, want 10.0", sys.MaxTargetDistance)
	}
}

// TestInteractionSystem_Update verifies that the system's Update method runs without error.
func TestInteractionSystem_Update(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity
	player := world.CreateEntity()
	playerPos := &components.Position{X: 5.0, Y: 5.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create an interactable object
	obj := world.CreateEntity()
	objPos := &components.Position{X: 6.0, Y: 5.0, Z: 0.0}
	envObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInventoriable,
		ObjectType:       "test_item",
		DisplayName:      "Test Item",
		InteractionType:  components.InteractionPickup,
		InteractionRange: 2.0,
	}
	world.AddComponent(obj, objPos)
	world.AddComponent(obj, envObj)

	// Update should not panic
	sys.Update(world, 0.016)
}

// TestInteractionSystem_ProximityHighlight verifies that objects within range get highlighted.
func TestInteractionSystem_ProximityHighlight(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity at origin
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create an object within highlight range
	nearObj := world.CreateEntity()
	nearObjPos := &components.Position{X: 2.0, Y: 0.0, Z: 0.0}
	nearEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInventoriable,
		ObjectType:       "near_item",
		DisplayName:      "Near Item",
		InteractionType:  components.InteractionPickup,
		InteractionRange: 3.0,
	}
	world.AddComponent(nearObj, nearObjPos)
	world.AddComponent(nearObj, nearEnvObj)

	// Create an object outside highlight range
	farObj := world.CreateEntity()
	farObjPos := &components.Position{X: 20.0, Y: 0.0, Z: 0.0}
	farEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInventoriable,
		ObjectType:       "far_item",
		DisplayName:      "Far Item",
		InteractionType:  components.InteractionPickup,
		InteractionRange: 3.0,
	}
	world.AddComponent(farObj, farObjPos)
	world.AddComponent(farObj, farEnvObj)

	// Update system
	sys.Update(world, 0.016)

	// Near object should be highlighted (within HighlightRange of 5.0)
	if nearEnvObj.HighlightState == 0 {
		t.Error("Near object should have HighlightState > 0")
	}

	// Far object should not be highlighted
	if farEnvObj.HighlightState != 0 {
		t.Errorf("Far object HighlightState = %d, want 0", farEnvObj.HighlightState)
	}
}

// TestInteractionSystem_Targeting verifies that look-based targeting works.
func TestInteractionSystem_Targeting(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity facing East (angle 0)
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create an object directly East of player
	eastObj := world.CreateEntity()
	eastObjPos := &components.Position{X: 2.0, Y: 0.0, Z: 0.0}
	eastEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInventoriable,
		ObjectType:       "east_item",
		DisplayName:      "East Item",
		InteractionType:  components.InteractionPickup,
		InteractionRange: 3.0,
	}
	world.AddComponent(eastObj, eastObjPos)
	world.AddComponent(eastObj, eastEnvObj)

	// Update system
	sys.Update(world, 0.016)

	// Check that the east object is targeted
	target, valid := sys.GetCurrentTarget()
	if !valid {
		t.Error("Expected a valid target")
	}
	if target != eastObj {
		t.Errorf("Expected target to be eastObj (%d), got %d", eastObj, target)
	}
}

// TestInteractionSystem_PickupInteraction verifies picking up items.
func TestInteractionSystem_PickupInteraction(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity with inventory
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create a pickupable object
	item := world.CreateEntity()
	itemPos := &components.Position{X: 1.0, Y: 0.0, Z: 0.0}
	itemEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInventoriable,
		ObjectType:       "health_potion",
		DisplayName:      "Health Potion",
		InteractionType:  components.InteractionPickup,
		InteractionRange: 2.0,
		ItemID:           "potion_health_small",
	}
	itemAppearance := &components.Appearance{Visible: true}
	world.AddComponent(item, itemPos)
	world.AddComponent(item, itemEnvObj)
	world.AddComponent(item, itemAppearance)

	// Update to establish targeting
	sys.Update(world, 0.016)

	// Request interaction
	sys.RequestInteraction(1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Check that item was picked up
	if itemEnvObj.InteractionType != components.InteractionNone {
		t.Errorf("Item InteractionType = %v, want InteractionNone", itemEnvObj.InteractionType)
	}

	if itemAppearance.Visible {
		t.Error("Item should be invisible after pickup")
	}

	// Check that item was added to inventory
	if len(playerInv.Items) != 1 {
		t.Errorf("Inventory has %d items, want 1", len(playerInv.Items))
	}

	if len(playerInv.Items) > 0 && playerInv.Items[0] != "potion_health_small" {
		t.Errorf("Inventory item = %s, want potion_health_small", playerInv.Items[0])
	}
}

// TestInteractionSystem_OpenDoor verifies opening doors.
func TestInteractionSystem_OpenDoor(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create an unlocked door
	door := world.CreateEntity()
	doorPos := &components.Position{X: 1.0, Y: 0.0, Z: 0.0}
	doorEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInteractive,
		ObjectType:       "door",
		DisplayName:      "Wooden Door",
		InteractionType:  components.InteractionOpen,
		InteractionRange: 2.0,
		IsLocked:         false,
		IsOpen:           false,
	}
	world.AddComponent(door, doorPos)
	world.AddComponent(door, doorEnvObj)

	// Update to establish targeting
	sys.Update(world, 0.016)

	// Request interaction
	sys.RequestInteraction(1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Check that door was opened
	if !doorEnvObj.IsOpen {
		t.Error("Door should be open after interaction")
	}
}

// TestInteractionSystem_LockedDoorWithKey verifies opening locked doors with key.
func TestInteractionSystem_LockedDoorWithKey(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity with a key in inventory
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{"key_dungeon_1"}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create a locked door
	door := world.CreateEntity()
	doorPos := &components.Position{X: 1.0, Y: 0.0, Z: 0.0}
	doorEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInteractive,
		ObjectType:       "door",
		DisplayName:      "Locked Door",
		InteractionType:  components.InteractionOpen,
		InteractionRange: 2.0,
		IsLocked:         true,
		RequiredKeyID:    "key_dungeon_1",
		IsOpen:           false,
	}
	world.AddComponent(door, doorPos)
	world.AddComponent(door, doorEnvObj)

	// Update to establish targeting
	sys.Update(world, 0.016)

	// Request interaction
	sys.RequestInteraction(1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Check that door was unlocked and opened
	if doorEnvObj.IsLocked {
		t.Error("Door should be unlocked after using key")
	}

	if !doorEnvObj.IsOpen {
		t.Error("Door should be open after interaction with key")
	}
}

// TestInteractionSystem_LockedDoorWithoutKey verifies locked doors can't be opened without key.
func TestInteractionSystem_LockedDoorWithoutKey(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity without the required key
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{"key_other"}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create a locked door
	door := world.CreateEntity()
	doorPos := &components.Position{X: 1.0, Y: 0.0, Z: 0.0}
	doorEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInteractive,
		ObjectType:       "door",
		DisplayName:      "Locked Door",
		InteractionType:  components.InteractionOpen,
		InteractionRange: 2.0,
		IsLocked:         true,
		RequiredKeyID:    "key_dungeon_1",
		IsOpen:           false,
	}
	world.AddComponent(door, doorPos)
	world.AddComponent(door, doorEnvObj)

	// Update to establish targeting
	sys.Update(world, 0.016)

	// Request interaction
	sys.RequestInteraction(1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Check that door is still locked and closed
	if !doorEnvObj.IsLocked {
		t.Error("Door should still be locked without correct key")
	}

	if doorEnvObj.IsOpen {
		t.Error("Door should remain closed without correct key")
	}
}

// TestInteractionSystem_PushObject verifies pushing movable objects.
func TestInteractionSystem_PushObject(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity facing East (angle 0)
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create a pushable object
	obj := world.CreateEntity()
	objPosInitial := 1.0
	objPos := &components.Position{X: objPosInitial, Y: 0.0, Z: 0.0}
	objEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInteractive,
		ObjectType:       "crate",
		DisplayName:      "Wooden Crate",
		InteractionType:  components.InteractionPush,
		InteractionRange: 2.0,
	}
	world.AddComponent(obj, objPos)
	world.AddComponent(obj, objEnvObj)

	// Update to establish targeting
	sys.Update(world, 0.016)

	// Request interaction
	sys.RequestInteraction(1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Check that object was pushed (X should increase since player faces East)
	if objPos.X <= objPosInitial {
		t.Errorf("Object X = %f, expected > %f after push", objPos.X, objPosInitial)
	}

	if !objEnvObj.IsUsed {
		t.Error("Object should be marked as used after push")
	}
}

// TestInteractionSystem_UseObject verifies using interactive objects.
func TestInteractionSystem_UseObject(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create a usable object (lever)
	lever := world.CreateEntity()
	leverPos := &components.Position{X: 1.0, Y: 0.0, Z: 0.0}
	leverEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInteractive,
		ObjectType:       "lever",
		DisplayName:      "Iron Lever",
		InteractionType:  components.InteractionUse,
		InteractionRange: 2.0,
		IsUsed:           false,
	}
	world.AddComponent(lever, leverPos)
	world.AddComponent(lever, leverEnvObj)

	// Update to establish targeting
	sys.Update(world, 0.016)

	// Request interaction
	sys.RequestInteraction(1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Check that lever was used
	if !leverEnvObj.IsUsed {
		t.Error("Lever should be marked as used after interaction")
	}
}

// TestInteractionSystem_ExamineObject verifies examining objects.
func TestInteractionSystem_ExamineObject(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create an examinable object
	statue := world.CreateEntity()
	statuePos := &components.Position{X: 1.0, Y: 0.0, Z: 0.0}
	statueEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryDecorative,
		ObjectType:       "statue",
		DisplayName:      "Ancient Statue",
		InteractionType:  components.InteractionExamine,
		InteractionRange: 3.0,
		ExamineText:      "A weathered statue depicting a forgotten deity.",
		IsUsed:           false,
	}
	world.AddComponent(statue, statuePos)
	world.AddComponent(statue, statueEnvObj)

	// Update to establish targeting
	sys.Update(world, 0.016)

	// Request interaction
	sys.RequestInteraction(1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Check that statue was examined
	if !statueEnvObj.IsUsed {
		t.Error("Statue should be marked as used/examined after interaction")
	}
}

// TestInteractionSystem_OutOfRange verifies that distant objects can't be interacted with.
func TestInteractionSystem_OutOfRange(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create an object far away (beyond max target distance of 10.0)
	farObj := world.CreateEntity()
	farObjPos := &components.Position{X: 15.0, Y: 0.0, Z: 0.0}
	farEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInventoriable,
		ObjectType:       "test_item",
		DisplayName:      "Far Item",
		InteractionType:  components.InteractionPickup,
		InteractionRange: 2.0,
		ItemID:           "test_item_id",
	}
	farAppearance := &components.Appearance{Visible: true}
	world.AddComponent(farObj, farObjPos)
	world.AddComponent(farObj, farEnvObj)
	world.AddComponent(farObj, farAppearance)

	// Update system
	sys.Update(world, 0.016)

	// Should have no valid target
	_, valid := sys.GetCurrentTarget()
	if valid {
		t.Error("Should not have valid target for distant object")
	}

	// Try to interact (should be no-op since no valid target)
	initialInteractionType := farEnvObj.InteractionType
	sys.RequestInteraction(1.0)
	sys.Update(world, 0.016)

	// Item should not have been picked up
	if farEnvObj.InteractionType != initialInteractionType {
		t.Error("Distant object should not be modified")
	}
}

// TestInteractionSystem_RequestInteractionWith verifies direct entity interaction.
func TestInteractionSystem_RequestInteractionWith(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create an object within range but not in the direction player is facing
	obj := world.CreateEntity()
	objPos := &components.Position{X: 0.0, Y: 1.0, Z: 0.0} // North of player, but player faces East
	objEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInteractive,
		ObjectType:       "lever",
		DisplayName:      "Lever",
		InteractionType:  components.InteractionUse,
		InteractionRange: 2.0,
		IsUsed:           false,
	}
	world.AddComponent(obj, objPos)
	world.AddComponent(obj, objEnvObj)

	// Update system (object won't be targeted since not in look direction)
	sys.Update(world, 0.016)

	// Request interaction directly with the entity (e.g., via mouse click)
	sys.RequestInteractionWith(obj, 1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Check that the lever was used
	if !objEnvObj.IsUsed {
		t.Error("Direct interaction should work even if not targeted by look")
	}
}

// TestInteractionSystem_DoorSwing verifies that opening a door triggers swing animation.
func TestInteractionSystem_DoorSwing(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create a door with a PhysicsBody
	door := world.CreateEntity()
	doorPos := &components.Position{X: 1.0, Y: 0.0, Z: 0.0}
	doorEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInteractive,
		ObjectType:       "door",
		DisplayName:      "Swinging Door",
		InteractionType:  components.InteractionOpen,
		InteractionRange: 2.0,
		IsLocked:         false,
		IsOpen:           false,
	}
	doorPhys := &components.PhysicsBody{
		Mass:          50.0,
		IsKinematic:   false,
		IsSwinging:    false, // Will be enabled when door opens
		SwingAngle:    0.0,
		SwingVelocity: 0.0,
		MaxSwingAngle: 0.0, // Will be set to Pi/2
		SwingDamping:  0.0, // Will be set
	}
	world.AddComponent(door, doorPos)
	world.AddComponent(door, doorEnvObj)
	world.AddComponent(door, doorPhys)

	// Update to establish targeting
	sys.Update(world, 0.016)

	// Request interaction to open door
	sys.RequestInteraction(1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Verify door is now open
	if !doorEnvObj.IsOpen {
		t.Error("Door should be open after interaction")
	}

	// Verify swing animation was triggered
	if !doorPhys.IsSwinging {
		t.Error("Door physics should have IsSwinging=true after opening")
	}
	if doorPhys.MaxSwingAngle == 0.0 {
		t.Error("Door should have MaxSwingAngle set for swing animation")
	}
	if doorPhys.SwingVelocity == 0.0 {
		t.Error("Door should have non-zero SwingVelocity to animate")
	}
	if doorPhys.SwingVelocity <= 0.0 {
		t.Error("Opening door should have positive SwingVelocity")
	}

	// Now close the door
	sys.RequestInteraction(1.0)
	sys.Update(world, 0.016)

	// Verify door is now closed
	if doorEnvObj.IsOpen {
		t.Error("Door should be closed after second interaction")
	}

	// Verify swing animation for closing
	if doorPhys.SwingVelocity >= 0.0 {
		t.Error("Closing door should have negative SwingVelocity")
	}
}

// TestInteractionSystem_DoorWithoutPhysicsBody verifies doors without PhysicsBody work.
func TestInteractionSystem_DoorWithoutPhysicsBody(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Create a player entity
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	playerInv := &components.Inventory{Items: []string{}}
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerInv)
	sys.SetPlayerEntity(player)

	// Create a door WITHOUT PhysicsBody
	door := world.CreateEntity()
	doorPos := &components.Position{X: 1.0, Y: 0.0, Z: 0.0}
	doorEnvObj := &components.EnvironmentObject{
		Category:         components.ObjectCategoryInteractive,
		ObjectType:       "door",
		DisplayName:      "Simple Door",
		InteractionType:  components.InteractionOpen,
		InteractionRange: 2.0,
		IsLocked:         false,
		IsOpen:           false,
	}
	world.AddComponent(door, doorPos)
	world.AddComponent(door, doorEnvObj)

	// Update to establish targeting
	sys.Update(world, 0.016)

	// Request interaction to open door
	sys.RequestInteraction(1.0)

	// Update to process interaction
	sys.Update(world, 0.016)

	// Verify door is now open (no panic without PhysicsBody)
	if !doorEnvObj.IsOpen {
		t.Error("Door without PhysicsBody should still open")
	}
}

// TestInteractionSystem_NoPlayer verifies system handles missing player gracefully.
func TestInteractionSystem_NoPlayer(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewInteractionSystem()

	// Don't set player entity

	// Create an object
	obj := world.CreateEntity()
	objPos := &components.Position{X: 1.0, Y: 0.0, Z: 0.0}
	objEnvObj := &components.EnvironmentObject{
		Category:        components.ObjectCategoryInventoriable,
		ObjectType:      "test_item",
		DisplayName:     "Test Item",
		InteractionType: components.InteractionPickup,
	}
	world.AddComponent(obj, objPos)
	world.AddComponent(obj, objEnvObj)

	// Update should not panic
	sys.Update(world, 0.016)

	// No valid target since no player
	_, valid := sys.GetCurrentTarget()
	if valid {
		t.Error("Should not have valid target without player")
	}
}
