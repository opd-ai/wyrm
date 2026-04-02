//go:build !noebiten

// Package main provides the quest UI overlay for the Wyrm client.
package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
	"github.com/opd-ai/wyrm/pkg/input"
)

// QuestUI handles quest log display and interaction.
type QuestUI struct {
	isOpen            bool
	genre             string
	playerEntity      ecs.Entity
	inputManager      *input.Manager
	selectedQuestIdx  int
	quests            []*questDisplayInfo
	notifications     []questNotification
	notificationTimer float64
	trackedQuestID    string
	// Pre-allocated rendering buffers for batch GPU uploads
	panelImage   *ebiten.Image // Main quest panel background
	panelPixels  []byte        // Pixel buffer for panel
	trackerImage *ebiten.Image // Quest tracker background
}

// questDisplayInfo holds display information for a quest.
type questDisplayInfo struct {
	ID           string
	Name         string
	Description  string
	Type         string
	Stage        int
	Completed    bool
	Objectives   []objectiveDisplayInfo
	Rewards      map[string]int
	IsTracked    bool
	DynamicQuest *systems.DynamicQuest
}

// objectiveDisplayInfo holds display information for a quest objective.
type objectiveDisplayInfo struct {
	Description string
	Required    int
	Current     int
	Completed   bool
}

// questNotification holds a quest completion notification.
type questNotification struct {
	Message string
	Timer   float64
}

const (
	questUIWidth             = 500
	questUIHeight            = 400
	questListWidth           = 180
	questNotificationTimeout = 5.0
)

// NewQuestUI creates a new quest UI overlay.
func NewQuestUI(genre string, playerEntity ecs.Entity, inputManager *input.Manager) *QuestUI {
	return &QuestUI{
		genre:         genre,
		playerEntity:  playerEntity,
		inputManager:  inputManager,
		quests:        make([]*questDisplayInfo, 0),
		notifications: make([]questNotification, 0),
	}
}

// IsOpen returns whether the quest log is currently open.
func (q *QuestUI) IsOpen() bool {
	return q.isOpen
}

// Open opens the quest log.
func (q *QuestUI) Open() {
	q.isOpen = true
}

// Close closes the quest log.
func (q *QuestUI) Close() {
	q.isOpen = false
}

// Toggle toggles the quest log open/closed state.
func (q *QuestUI) Toggle() {
	q.isOpen = !q.isOpen
}

// Update processes quest UI input and updates quest data from the world.
func (q *QuestUI) Update(world *ecs.World, dt float64) {
	q.updateNotifications(dt)
	q.refreshQuestList(world)

	if !q.isOpen {
		return
	}

	q.handleInput()
}

// updateNotifications decrements notification timers and removes expired ones.
func (q *QuestUI) updateNotifications(dt float64) {
	active := make([]questNotification, 0, len(q.notifications))
	for _, n := range q.notifications {
		n.Timer -= dt
		if n.Timer > 0 {
			active = append(active, n)
		}
	}
	q.notifications = active
}

// refreshQuestList updates the quest list from the ECS world.
func (q *QuestUI) refreshQuestList(world *ecs.World) {
	if world == nil {
		return
	}

	prevCompleted := make(map[string]bool)
	for _, quest := range q.quests {
		if quest.Completed {
			prevCompleted[quest.ID] = true
		}
	}

	q.quests = make([]*questDisplayInfo, 0)

	for _, e := range world.Entities("Quest") {
		comp, ok := world.GetComponent(e, "Quest")
		if !ok {
			continue
		}
		quest := comp.(*components.Quest)
		info := q.questToDisplayInfo(quest)
		q.quests = append(q.quests, info)

		// Check for new completions
		if info.Completed && !prevCompleted[info.ID] {
			q.addNotification(fmt.Sprintf("Quest Completed: %s", info.Name))
		}
	}

	// Clamp selection to valid range
	if q.selectedQuestIdx >= len(q.quests) {
		q.selectedQuestIdx = len(q.quests) - 1
	}
	if q.selectedQuestIdx < 0 {
		q.selectedQuestIdx = 0
	}
}

// questToDisplayInfo converts a Quest component to display info.
func (q *QuestUI) questToDisplayInfo(quest *components.Quest) *questDisplayInfo {
	info := &questDisplayInfo{
		ID:        quest.ID,
		Name:      getQuestName(quest.ID, q.genre),
		Type:      getQuestType(quest.ID),
		Stage:     quest.CurrentStage,
		Completed: quest.Completed,
		Rewards:   make(map[string]int),
		IsTracked: quest.ID == q.trackedQuestID,
	}

	info.Description = getQuestDescription(quest.ID, q.genre)
	info.Objectives = getQuestObjectives(quest, q.genre)

	return info
}

// getQuestName returns a display name for a quest ID.
func getQuestName(questID, genre string) string {
	// This would integrate with QuestAdapter in production
	// For now, generate a readable name from the ID
	return formatQuestID(questID)
}

// getQuestType extracts the quest type from the ID.
func getQuestType(questID string) string {
	// Quest IDs are typically prefix_suffix
	for _, prefix := range []string{"famine", "war", "bandit", "monster", "politics", "quest"} {
		if len(questID) > len(prefix) && questID[:len(prefix)] == prefix {
			return prefix
		}
	}
	return "misc"
}

// getQuestDescription returns a description for a quest.
func getQuestDescription(questID, genre string) string {
	// Default description based on quest type
	qtype := getQuestType(questID)
	switch qtype {
	case "famine":
		return "Help gather supplies to ease the famine."
	case "war":
		return "Assist in the war effort against our enemies."
	case "bandit":
		return "Clear the roads of bandits."
	case "monster":
		return "Hunt down the creature terrorizing the region."
	case "politics":
		return "Navigate the complex political landscape."
	default:
		return "Complete this quest for rewards."
	}
}

// getQuestObjectives returns objectives for a quest.
func getQuestObjectives(quest *components.Quest, genre string) []objectiveDisplayInfo {
	// Generate objectives based on quest stage and flags
	objectives := make([]objectiveDisplayInfo, 0)

	// Create at least one objective based on quest type
	qtype := getQuestType(quest.ID)
	obj := objectiveDisplayInfo{
		Description: getObjectiveDescription(qtype, quest.CurrentStage),
		Required:    1,
		Current:     0,
		Completed:   quest.Completed,
	}

	// Check flags for completion
	if quest.Flags != nil {
		for flagName := range quest.Flags {
			if quest.Flags[flagName] {
				obj.Current++
			}
		}
	}

	if obj.Current >= obj.Required || quest.Completed {
		obj.Completed = true
		obj.Current = obj.Required
	}

	objectives = append(objectives, obj)
	return objectives
}

// getObjectiveDescription returns a description for an objective.
func getObjectiveDescription(questType string, stage int) string {
	switch questType {
	case "famine":
		return "Gather supplies"
	case "war":
		return "Defeat enemies"
	case "bandit":
		return "Eliminate bandits"
	case "monster":
		return "Slay the beast"
	case "politics":
		return "Speak with leaders"
	default:
		return fmt.Sprintf("Complete stage %d", stage+1)
	}
}

// formatQuestID converts a quest ID to a readable name.
func formatQuestID(questID string) string {
	if len(questID) == 0 {
		return "Unknown Quest"
	}
	// Capitalize first letter and replace underscores with spaces
	result := make([]byte, 0, len(questID))
	capitalize := true
	for i := 0; i < len(questID); i++ {
		c := questID[i]
		if c == '_' {
			result = append(result, ' ')
			capitalize = true
		} else if capitalize {
			if c >= 'a' && c <= 'z' {
				result = append(result, c-32)
			} else {
				result = append(result, c)
			}
			capitalize = false
		} else {
			result = append(result, c)
		}
	}
	return string(result)
}

// handleInput processes keyboard/mouse input for the quest UI.
func (q *QuestUI) handleInput() {
	// Navigate quest list
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		q.selectedQuestIdx--
		if q.selectedQuestIdx < 0 {
			q.selectedQuestIdx = len(q.quests) - 1
			if q.selectedQuestIdx < 0 {
				q.selectedQuestIdx = 0
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		q.selectedQuestIdx++
		if q.selectedQuestIdx >= len(q.quests) {
			q.selectedQuestIdx = 0
		}
	}

	// Track quest
	if inpututil.IsKeyJustPressed(ebiten.KeyT) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if q.selectedQuestIdx >= 0 && q.selectedQuestIdx < len(q.quests) {
			quest := q.quests[q.selectedQuestIdx]
			if q.trackedQuestID == quest.ID {
				q.trackedQuestID = "" // Untrack
			} else {
				q.trackedQuestID = quest.ID
			}
		}
	}

	// Handle mouse click on quest list
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		q.handleMouseClick(mx, my)
	}
}

// handleMouseClick processes mouse clicks on the quest UI.
func (q *QuestUI) handleMouseClick(mx, my int) {
	// Check if click is in quest list area
	screenW, screenH := 1280, 720 // Default, would get from ebiten.WindowSize() in production
	startX := (screenW - questUIWidth) / 2
	startY := (screenH - questUIHeight) / 2

	listX := startX + 10
	listY := startY + 40
	listH := questUIHeight - 50

	if mx >= listX && mx < listX+questListWidth && my >= listY && my < listY+listH {
		// Calculate which quest was clicked
		relY := my - listY
		idx := relY / 20 // 20 pixels per quest entry
		if idx >= 0 && idx < len(q.quests) {
			q.selectedQuestIdx = idx
		}
	}
}

// addNotification adds a quest completion notification.
func (q *QuestUI) addNotification(message string) {
	q.notifications = append(q.notifications, questNotification{
		Message: message,
		Timer:   questNotificationTimeout,
	})
}

// Draw renders the quest UI overlay and notifications.
func (q *QuestUI) Draw(screen *ebiten.Image) {
	// Always draw notifications
	q.drawNotifications(screen)

	// Always draw quest tracker
	q.drawQuestTracker(screen)

	// Draw full UI only when open
	if !q.isOpen {
		return
	}

	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()
	startX := (screenW - questUIWidth) / 2
	startY := (screenH - questUIHeight) / 2

	// Draw background
	q.drawBackground(screen, startX, startY)

	// Draw title
	q.drawTitle(screen, startX, startY)

	// Draw quest list
	q.drawQuestList(screen, startX, startY)

	// Draw selected quest details
	q.drawQuestDetails(screen, startX, startY)

	// Draw instructions
	q.drawInstructions(screen, startX, startY)
}

// drawBackground draws the quest UI background panel.
func (q *QuestUI) drawBackground(screen *ebiten.Image, x, y int) {
	bgColor := q.getBackgroundColor()
	for dy := 0; dy < questUIHeight; dy++ {
		for dx := 0; dx < questUIWidth; dx++ {
			screen.Set(x+dx, y+dy, bgColor)
		}
	}

	// Border
	borderColor := q.getBorderColor()
	for dx := 0; dx < questUIWidth; dx++ {
		screen.Set(x+dx, y, borderColor)
		screen.Set(x+dx, y+questUIHeight-1, borderColor)
	}
	for dy := 0; dy < questUIHeight; dy++ {
		screen.Set(x, y+dy, borderColor)
		screen.Set(x+questUIWidth-1, y+dy, borderColor)
	}

	// Divider between list and details
	for dy := 40; dy < questUIHeight-30; dy++ {
		screen.Set(x+questListWidth+15, y+dy, borderColor)
	}
}

// drawTitle draws the quest log title.
func (q *QuestUI) drawTitle(screen *ebiten.Image, x, y int) {
	title := q.getUITitle()
	ebitenutil.DebugPrintAt(screen, title, x+questUIWidth/2-len(title)*3, y+10)
}

// drawQuestList draws the list of quests.
func (q *QuestUI) drawQuestList(screen *ebiten.Image, x, y int) {
	listX := x + 10
	listY := y + 40

	if len(q.quests) == 0 {
		ebitenutil.DebugPrintAt(screen, "No quests", listX, listY)
		return
	}

	for i, quest := range q.quests {
		entryY := listY + i*20
		if entryY > y+questUIHeight-50 {
			break // Out of space
		}

		// Selection highlight
		if i == q.selectedQuestIdx {
			for dx := 0; dx < questListWidth; dx++ {
				screen.Set(listX+dx, entryY-2, color.RGBA{80, 80, 120, 200})
				screen.Set(listX+dx, entryY+14, color.RGBA{80, 80, 120, 200})
			}
		}

		// Quest status indicator
		prefix := "[ ] "
		if quest.Completed {
			prefix = "[X] "
		} else if quest.IsTracked {
			prefix = "[*] "
		}

		// Truncate name if too long
		name := quest.Name
		maxNameLen := 15
		if len(name) > maxNameLen {
			name = name[:maxNameLen-2] + ".."
		}

		textColor := q.getQuestTextColor(quest)
		_ = textColor // Would use with custom font rendering
		ebitenutil.DebugPrintAt(screen, prefix+name, listX, entryY)
	}
}

// drawQuestDetails draws the selected quest's details.
func (q *QuestUI) drawQuestDetails(screen *ebiten.Image, x, y int) {
	detailsX := x + questListWidth + 25
	detailsY := y + 40
	detailsW := questUIWidth - questListWidth - 35

	if q.selectedQuestIdx < 0 || q.selectedQuestIdx >= len(q.quests) {
		ebitenutil.DebugPrintAt(screen, "Select a quest", detailsX, detailsY)
		return
	}

	quest := q.quests[q.selectedQuestIdx]

	// Quest name
	ebitenutil.DebugPrintAt(screen, quest.Name, detailsX, detailsY)
	detailsY += 20

	// Quest status
	status := "Active"
	if quest.Completed {
		status = "Completed"
	}
	ebitenutil.DebugPrintAt(screen, "Status: "+status, detailsX, detailsY)
	detailsY += 20

	// Description (word wrap)
	desc := quest.Description
	wrapped := wrapTextQuest(desc, detailsW/6)
	for _, line := range wrapped {
		ebitenutil.DebugPrintAt(screen, line, detailsX, detailsY)
		detailsY += 14
	}
	detailsY += 10

	// Objectives
	ebitenutil.DebugPrintAt(screen, "Objectives:", detailsX, detailsY)
	detailsY += 16
	for _, obj := range quest.Objectives {
		checkmark := "[ ]"
		if obj.Completed {
			checkmark = "[X]"
		}
		objText := fmt.Sprintf("%s %s (%d/%d)", checkmark, obj.Description, obj.Current, obj.Required)
		ebitenutil.DebugPrintAt(screen, objText, detailsX+10, detailsY)
		detailsY += 14
	}
}

// drawInstructions draws UI control instructions.
func (q *QuestUI) drawInstructions(screen *ebiten.Image, x, y int) {
	instructions := "[W/S] Navigate  [T] Track  [J] Close"
	ebitenutil.DebugPrintAt(screen, instructions, x+10, y+questUIHeight-20)
}

// drawNotifications draws quest completion notifications.
func (q *QuestUI) drawNotifications(screen *ebiten.Image) {
	screenW := screen.Bounds().Dx()

	for i, notif := range q.notifications {
		// Fade out effect
		alpha := uint8(255)
		if notif.Timer < 1.0 {
			alpha = uint8(notif.Timer * 255)
		}

		// Position at top center of screen
		notifY := 50 + i*25
		notifX := screenW/2 - len(notif.Message)*3

		// Draw background
		bgColor := color.RGBA{40, 80, 40, alpha}
		for dx := -5; dx < len(notif.Message)*6+5; dx++ {
			for dy := -2; dy < 14; dy++ {
				screen.Set(notifX+dx, notifY+dy, bgColor)
			}
		}

		ebitenutil.DebugPrintAt(screen, notif.Message, notifX, notifY)
	}
}

// drawQuestTracker draws the active quest tracker on-screen.
func (q *QuestUI) drawQuestTracker(screen *ebiten.Image) {
	if q.trackedQuestID == "" {
		return
	}

	// Find tracked quest
	var tracked *questDisplayInfo
	for _, quest := range q.quests {
		if quest.ID == q.trackedQuestID {
			tracked = quest
			break
		}
	}

	if tracked == nil || tracked.Completed {
		q.trackedQuestID = "" // Clear invalid tracking
		return
	}

	// Draw tracker in top-right corner
	screenW := screen.Bounds().Dx()
	trackerX := screenW - 220
	trackerY := 10

	// Background
	bgColor := color.RGBA{30, 30, 50, 180}
	for dx := 0; dx < 210; dx++ {
		for dy := 0; dy < 60; dy++ {
			screen.Set(trackerX+dx, trackerY+dy, bgColor)
		}
	}

	// Quest name
	name := tracked.Name
	if len(name) > 25 {
		name = name[:22] + "..."
	}
	ebitenutil.DebugPrintAt(screen, name, trackerX+5, trackerY+5)

	// Current objective
	if len(tracked.Objectives) > 0 {
		obj := tracked.Objectives[0]
		objText := fmt.Sprintf("> %s (%d/%d)", obj.Description, obj.Current, obj.Required)
		if len(objText) > 30 {
			objText = objText[:27] + "..."
		}
		ebitenutil.DebugPrintAt(screen, objText, trackerX+5, trackerY+25)
	}
}

// wrapTextQuest splits text into lines of max width.
func wrapTextQuest(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = 40
	}

	var lines []string
	for len(text) > maxWidth {
		// Find break point
		breakIdx := maxWidth
		for i := maxWidth; i > 0; i-- {
			if text[i] == ' ' {
				breakIdx = i
				break
			}
		}
		lines = append(lines, text[:breakIdx])
		text = text[breakIdx:]
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:]
		}
	}
	if len(text) > 0 {
		lines = append(lines, text)
	}
	return lines
}

// Genre-specific styling methods

func (q *QuestUI) getBackgroundColor() color.Color {
	switch q.genre {
	case "fantasy":
		return color.RGBA{40, 30, 20, 230}
	case "sci-fi":
		return color.RGBA{20, 30, 40, 230}
	case "horror":
		return color.RGBA{20, 15, 25, 240}
	case "cyberpunk":
		return color.RGBA{15, 15, 30, 230}
	case "post-apocalyptic":
		return color.RGBA{35, 30, 25, 230}
	default:
		return color.RGBA{30, 30, 30, 230}
	}
}

func (q *QuestUI) getBorderColor() color.Color {
	switch q.genre {
	case "fantasy":
		return color.RGBA{180, 150, 100, 255}
	case "sci-fi":
		return color.RGBA{100, 150, 200, 255}
	case "horror":
		return color.RGBA{100, 60, 80, 255}
	case "cyberpunk":
		return color.RGBA{0, 255, 200, 255}
	case "post-apocalyptic":
		return color.RGBA{180, 140, 80, 255}
	default:
		return color.RGBA{150, 150, 150, 255}
	}
}

func (q *QuestUI) getUITitle() string {
	switch q.genre {
	case "fantasy":
		return "=== Quest Journal ==="
	case "sci-fi":
		return "[ MISSION LOG ]"
	case "horror":
		return "-- Dark Tasks --"
	case "cyberpunk":
		return "// CONTRACTS //"
	case "post-apocalyptic":
		return "- Survival Jobs -"
	default:
		return "Quest Log"
	}
}

func (q *QuestUI) getQuestTextColor(quest *questDisplayInfo) color.Color {
	if quest.Completed {
		return color.RGBA{100, 100, 100, 255}
	}
	if quest.IsTracked {
		return color.RGBA{255, 220, 100, 255}
	}
	return color.RGBA{255, 255, 255, 255}
}
