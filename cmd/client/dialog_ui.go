//go:build !noebiten

// dialog_ui.go provides the dialog overlay for NPC conversations.
// Per PLAN.md Phase 2 Task 2B.

package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/wyrm/pkg/dialog"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/rendering/subtitles"
)

// DialogState tracks the current state of a dialog interaction.
type DialogState int

const (
	DialogStateClosed DialogState = iota
	DialogStateOpen
	DialogStateWaitingForChoice
	DialogStateSkillCheck
	DialogStateExiting
)

// DialogOption represents a selectable dialog response.
type DialogOption struct {
	ID              string
	Text            string
	SkillCheck      *dialog.SkillCheck // Optional skill check required
	ConsequenceType string             // e.g., "friendly", "hostile", "quest"
}

// DialogUI manages the dialog interface and conversation state.
type DialogUI struct {
	state            DialogState
	npcEntity        ecs.Entity
	npcName          string
	npcEmotion       dialog.EmotionalState
	currentText      string
	options          []DialogOption
	selectedOption   int
	dialogManager    *dialog.Manager
	subtitleSystem   *subtitles.SubtitleSystem
	genre            string
	playerEntity     ecs.Entity
	lastInteraction  time.Time
	skillCheckResult *dialog.SkillCheckResult
	world            *ecs.World // World reference for skill lookups
	// Pre-allocated overlay image to avoid per-frame GPU allocation
	overlayImage       *ebiten.Image
	overlayImageWidth  int
	overlayImageHeight int
}

// NewDialogUI creates a new dialog UI system.
func NewDialogUI(genre string, playerEntity ecs.Entity) *DialogUI {
	return &DialogUI{
		state:          DialogStateClosed,
		dialogManager:  dialog.NewManager(time.Now().UnixNano()),
		subtitleSystem: subtitles.NewSubtitleSystem(true),
		genre:          genre,
		playerEntity:   playerEntity,
		options:        make([]DialogOption, 0),
	}
}

// IsOpen returns true if the dialog UI is currently displayed.
func (d *DialogUI) IsOpen() bool {
	return d.state != DialogStateClosed
}

// OpenDialog starts a conversation with the specified NPC entity.
func (d *DialogUI) OpenDialog(world *ecs.World, npcEntity ecs.Entity, npcName string) {
	d.state = DialogStateOpen
	d.npcEntity = npcEntity
	d.npcName = npcName
	d.selectedOption = 0
	d.skillCheckResult = nil
	d.world = world // Store world reference for skill lookups

	// Get NPC's emotional state toward player
	d.npcEmotion = d.dialogManager.GetEmotionalState(
		uint64(npcEntity),
		uint64(d.playerEntity),
		dialog.EmotionNeutral,
	)

	// Generate opening dialog based on emotion and genre
	d.currentText = d.generateGreeting()
	d.generateResponseOptions()

	// Add to subtitle system for accessibility
	d.subtitleSystem.AddDialog(d.npcName, d.currentText)

	d.lastInteraction = time.Now()
}

// CloseDialog ends the current conversation.
func (d *DialogUI) CloseDialog() {
	if d.state == DialogStateClosed {
		return
	}

	// Record farewell in memory
	if d.npcEntity != 0 {
		d.dialogManager.RecordTopic(
			uint64(d.npcEntity),
			uint64(d.playerEntity),
			"conversation_end",
			"player_left",
			"conversation ended",
		)
	}

	d.state = DialogStateClosed
	d.npcEntity = 0
	d.npcName = ""
	d.currentText = ""
	d.options = d.options[:0]
	d.selectedOption = 0
	d.skillCheckResult = nil
	d.subtitleSystem.Clear()
}

// Update handles input and updates the dialog state.
func (d *DialogUI) Update() {
	if d.state == DialogStateClosed {
		return
	}

	d.subtitleSystem.Update()

	switch d.state {
	case DialogStateOpen, DialogStateWaitingForChoice:
		d.handleOptionSelection()
	case DialogStateSkillCheck:
		d.handleSkillCheckResult()
	case DialogStateExiting:
		d.CloseDialog()
	}
}

// handleOptionSelection processes player input for selecting dialog options.
func (d *DialogUI) handleOptionSelection() {
	if len(d.options) == 0 {
		if ebiten.IsKeyPressed(ebiten.KeyE) || ebiten.IsKeyPressed(ebiten.KeyEscape) {
			d.state = DialogStateExiting
		}
		return
	}

	// Navigate options
	if inputJustPressed(ebiten.KeyUp) || inputJustPressed(ebiten.KeyW) {
		d.selectedOption--
		if d.selectedOption < 0 {
			d.selectedOption = len(d.options) - 1
		}
	}
	if inputJustPressed(ebiten.KeyDown) || inputJustPressed(ebiten.KeyS) {
		d.selectedOption++
		if d.selectedOption >= len(d.options) {
			d.selectedOption = 0
		}
	}

	// Select option with Enter or E
	if inputJustPressed(ebiten.KeyEnter) || inputJustPressed(ebiten.KeyE) {
		d.selectOption(d.selectedOption)
	}

	// Close with Escape
	if inputJustPressed(ebiten.KeyEscape) {
		d.state = DialogStateExiting
	}
}

// inputJustPressed checks if a key was just pressed this frame.
// This is a simplified debounce check.
var lastKeyState = make(map[ebiten.Key]bool)

func inputJustPressed(key ebiten.Key) bool {
	pressed := ebiten.IsKeyPressed(key)
	wasPressed := lastKeyState[key]
	lastKeyState[key] = pressed
	return pressed && !wasPressed
}

// selectOption processes the player's selected dialog choice.
func (d *DialogUI) selectOption(index int) {
	if index < 0 || index >= len(d.options) {
		return
	}

	option := d.options[index]

	// Check if this option requires a skill check
	if option.SkillCheck != nil {
		d.performSkillCheck(option)
		return
	}

	// Process the selected option
	d.processDialogChoice(option)
}

// performSkillCheck executes a dialog skill check.
func (d *DialogUI) performSkillCheck(option DialogOption) {
	result := d.dialogManager.PerformSkillCheck(
		uint64(d.npcEntity),
		uint64(d.playerEntity),
		*option.SkillCheck,
	)

	d.skillCheckResult = result
	d.state = DialogStateSkillCheck

	// Update current text with skill check result
	checkName := option.SkillCheck.Type.String()
	if result.Success {
		if result.CriticalSuccess {
			d.currentText = fmt.Sprintf("[Critical %s Success!]\n%s", checkName, result.ResponseText)
		} else {
			d.currentText = fmt.Sprintf("[%s Success]\n%s", checkName, result.ResponseText)
		}
	} else {
		if result.CriticalFailure {
			d.currentText = fmt.Sprintf("[Critical %s Failure!]\n%s", checkName, result.ResponseText)
		} else {
			d.currentText = fmt.Sprintf("[%s Failed]\n%s", checkName, result.ResponseText)
		}
	}

	// Update NPC emotion based on skill check
	d.npcEmotion = d.dialogManager.GetEmotionalState(
		uint64(d.npcEntity),
		uint64(d.playerEntity),
		d.npcEmotion,
	)

	d.subtitleSystem.AddDialog(d.npcName, d.currentText)
}

// handleSkillCheckResult processes input after a skill check.
func (d *DialogUI) handleSkillCheckResult() {
	if inputJustPressed(ebiten.KeyEnter) || inputJustPressed(ebiten.KeyE) {
		d.state = DialogStateWaitingForChoice
		d.generateResponseOptions()
	}
	if inputJustPressed(ebiten.KeyEscape) {
		d.state = DialogStateExiting
	}
}

// processDialogChoice handles a non-skill-check dialog option.
func (d *DialogUI) processDialogChoice(option DialogOption) {
	// Record the choice in NPC memory
	d.dialogManager.RecordTopic(
		uint64(d.npcEntity),
		uint64(d.playerEntity),
		option.ID,
		option.Text,
		d.currentText,
	)

	// Apply emotional consequence
	shift := d.getEmotionShiftForConsequence(option.ConsequenceType)
	if shift != 0 {
		d.dialogManager.ShiftEmotion(
			uint64(d.npcEntity),
			uint64(d.playerEntity),
			shift,
		)
		d.npcEmotion = d.dialogManager.GetEmotionalState(
			uint64(d.npcEntity),
			uint64(d.playerEntity),
			d.npcEmotion,
		)
	}

	// Generate NPC response
	response := d.dialogManager.GenerateResponse(
		uint64(d.npcEntity),
		uint64(d.playerEntity),
		d.genre,
		option.ID,
		d.npcEmotion,
	)

	d.currentText = response.Text
	d.subtitleSystem.AddDialog(d.npcName, d.currentText)

	// Check for exit options
	if option.ConsequenceType == "exit" {
		d.state = DialogStateExiting
		return
	}

	// Generate new response options
	d.generateResponseOptions()
	d.state = DialogStateWaitingForChoice
}

// getEmotionShiftForConsequence returns emotion shift for a consequence type.
func (d *DialogUI) getEmotionShiftForConsequence(consequenceType string) float64 {
	switch consequenceType {
	case "friendly":
		return 5.0
	case "hostile":
		return -10.0
	case "helpful":
		return 10.0
	case "threatening":
		return -15.0
	case "quest":
		return 3.0
	default:
		return 0.0
	}
}

// generateGreeting creates the NPC's opening dialog.
func (d *DialogUI) generateGreeting() string {
	vocab := dialog.GetVocabulary(d.genre)
	mods := dialog.EmotionModifiers[d.npcEmotion]

	// Check for recalled topic
	lastTopic := d.dialogManager.GetLastTopic(uint64(d.npcEntity), uint64(d.playerEntity))
	var recall string
	if lastTopic != nil && time.Since(lastTopic.Timestamp) < 24*time.Hour {
		recall = fmt.Sprintf(" We spoke about %s last time.", lastTopic.Topic)
	}

	// Get genre-appropriate greeting
	greeting := vocab.Greetings[0]
	if len(vocab.Greetings) > 1 {
		greeting = vocab.Greetings[int(time.Now().UnixNano())%len(vocab.Greetings)]
	}

	// Add emotional prefix if applicable
	if len(mods.Prefixes) > 0 && mods.Prefixes[0] != "" {
		greeting = mods.Prefixes[0] + " " + greeting
	}

	return greeting + recall
}

// getPlayerSkillLevel returns the player's level in a given skill.
// Falls back to default if Skills component is not available.
func (d *DialogUI) getPlayerSkillLevel(skillName string, defaultValue int) int {
	if d.world == nil {
		return defaultValue
	}
	skillsComp, ok := d.world.GetComponent(d.playerEntity, "Skills")
	if !ok {
		return defaultValue
	}
	skills := skillsComp.(*components.Skills)
	if skills.Levels == nil {
		return defaultValue
	}
	level, found := skills.Levels[skillName]
	if !found {
		return defaultValue
	}
	return level
}

// generateResponseOptions creates the available dialog choices.
func (d *DialogUI) generateResponseOptions() {
	d.options = d.options[:0]
	d.selectedOption = 0

	vocab := dialog.GetVocabulary(d.genre)

	// Always add a friendly response option
	d.options = append(d.options, DialogOption{
		ID:              "friendly_response",
		Text:            "[Friendly] " + vocab.Affirmatives[0],
		ConsequenceType: "friendly",
	})

	// Add a question option
	if len(vocab.CommonWords) > 0 {
		word := vocab.CommonWords[int(time.Now().UnixNano())%len(vocab.CommonWords)]
		d.options = append(d.options, DialogOption{
			ID:              "ask_about_" + word,
			Text:            fmt.Sprintf("[Ask] Tell me about the %s.", word),
			ConsequenceType: "neutral",
		})
	}

	// Add persuasion option if not already friendly
	if d.npcEmotion != dialog.EmotionFriendly {
		persuasionLevel := d.getPlayerSkillLevel("persuasion", 50)
		d.options = append(d.options, DialogOption{
			ID:   "persuade",
			Text: "[Persuade] Perhaps we could come to an understanding...",
			SkillCheck: &dialog.SkillCheck{
				Type:       dialog.SkillCheckPersuasion,
				Difficulty: dialog.DifficultyMedium,
				SkillLevel: persuasionLevel,
				NPCState:   d.npcEmotion,
				Genre:      d.genre,
			},
			ConsequenceType: "friendly",
		})
	}

	// Add intimidation option
	intimidationLevel := d.getPlayerSkillLevel("intimidation", 30)
	d.options = append(d.options, DialogOption{
		ID:   "intimidate",
		Text: "[Intimidate] You would be wise to cooperate.",
		SkillCheck: &dialog.SkillCheck{
			Type:       dialog.SkillCheckIntimidate,
			Difficulty: dialog.DifficultyMedium,
			SkillLevel: intimidationLevel,
			NPCState:   d.npcEmotion,
			Genre:      d.genre,
		},
		ConsequenceType: "threatening",
	})

	// Always add exit option
	d.options = append(d.options, DialogOption{
		ID:              "farewell",
		Text:            "[Leave] " + vocab.Farewells[0],
		ConsequenceType: "exit",
	})
}

// Draw renders the dialog UI overlay.
func (d *DialogUI) Draw(screen *ebiten.Image) {
	if d.state == DialogStateClosed {
		return
	}

	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()

	// Draw semi-transparent background overlay
	d.drawOverlay(screen, screenWidth, screenHeight)

	// Draw NPC name and emotional state
	d.drawNPCHeader(screen, screenWidth)

	// Draw current dialog text
	d.drawDialogText(screen, screenWidth, screenHeight)

	// Draw response options
	d.drawOptions(screen, screenWidth, screenHeight)

	// Draw subtitle if enabled
	d.drawSubtitle(screen, screenWidth, screenHeight)
}

// drawOverlay draws the semi-transparent dialog background using Fill.
func (d *DialogUI) drawOverlay(screen *ebiten.Image, width, height int) {
	// Draw dialog box at bottom of screen
	boxHeight := height / 3
	boxY := height - boxHeight

	// Pre-allocate overlay image once, or reallocate if size changed
	if d.overlayImage == nil || d.overlayImageWidth != width || d.overlayImageHeight != boxHeight {
		d.overlayImage = ebiten.NewImage(width, boxHeight)
		d.overlayImageWidth = width
		d.overlayImageHeight = boxHeight
	}

	// Draw semi-transparent background using Fill
	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	d.overlayImage.Fill(bgColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(boxY))
	screen.DrawImage(d.overlayImage, op)
}

// drawNPCHeader draws the NPC name and emotional state.
func (d *DialogUI) drawNPCHeader(screen *ebiten.Image, screenWidth int) {
	screenHeight := screen.Bounds().Dy()
	boxY := screenHeight - (screenHeight / 3)

	emotionText := fmt.Sprintf("[%s]", d.npcEmotion.String())
	headerText := fmt.Sprintf("%s %s", d.npcName, emotionText)

	// Center the header
	x := (screenWidth - len(headerText)*6) / 2
	ebitenutil.DebugPrintAt(screen, headerText, x, boxY+10)
}

// drawDialogText draws the current dialog content.
func (d *DialogUI) drawDialogText(screen *ebiten.Image, screenWidth, screenHeight int) {
	boxY := screenHeight - (screenHeight / 3)
	textY := boxY + 35

	// Wrap text to fit screen width
	maxChars := (screenWidth - 40) / 6
	lines := wrapText(d.currentText, maxChars)

	for i, line := range lines {
		ebitenutil.DebugPrintAt(screen, line, 20, textY+(i*15))
	}
}

// drawOptions draws the selectable dialog responses.
func (d *DialogUI) drawOptions(screen *ebiten.Image, screenWidth, screenHeight int) {
	if len(d.options) == 0 {
		ebitenutil.DebugPrintAt(screen, "[Press E to continue]", 20, screenHeight-50)
		return
	}

	optionY := screenHeight - 20 - (len(d.options) * 18)

	for i, option := range d.options {
		prefix := "  "
		if i == d.selectedOption {
			prefix = "> "
		}

		optionText := fmt.Sprintf("%s%d. %s", prefix, i+1, option.Text)
		ebitenutil.DebugPrintAt(screen, optionText, 20, optionY+(i*18))
	}
}

// drawSubtitle draws accessibility subtitle overlay.
func (d *DialogUI) drawSubtitle(screen *ebiten.Image, screenWidth, screenHeight int) {
	renderData := d.subtitleSystem.GetRenderData()
	if renderData == nil {
		return
	}

	// Draw subtitle at bottom center with background
	text := renderData.FormatText()
	textWidth := len(text) * 6
	x := (screenWidth - textWidth) / 2
	y := screenHeight - 30

	ebitenutil.DebugPrintAt(screen, text, x, y)
}

// wrapText wraps text to fit within maxChars per line.
func wrapText(text string, maxChars int) []string {
	if maxChars <= 0 {
		maxChars = 80
	}

	var lines []string
	words := splitWords(text)
	currentLine := ""

	for _, word := range words {
		if word == "\n" {
			lines = append(lines, currentLine)
			currentLine = ""
			continue
		}

		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= maxChars {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// splitWords splits text into words, preserving newlines as separate tokens.
func splitWords(text string) []string {
	var words []string
	currentWord := ""

	for _, ch := range text {
		if ch == '\n' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
			words = append(words, "\n")
		} else if ch == ' ' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(ch)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}
