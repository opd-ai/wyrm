//go:build !noebiten

// character_creation.go provides the character creation screen for new players.
package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// CharacterCreationState represents the current step in character creation.
type CharacterCreationState int

const (
	CCStateGenreSelection CharacterCreationState = iota
	CCStateNameInput
	CCStateSkillAllocation
	CCStateEquipmentChoice
	CCStateConfirm
	CCStateComplete
)

// GenreInfo contains genre name and description.
type GenreInfo struct {
	ID          string
	Name        string
	Description string
}

// SkillSchool represents a skill school for allocation.
type SkillSchool struct {
	Name        string
	Description string
	Points      int
}

// StartingEquipment represents a starting equipment choice.
type StartingEquipment struct {
	Name        string
	Description string
	Items       []string
}

// CharacterCreation manages the character creation process.
type CharacterCreation struct {
	state              CharacterCreationState
	selectedIndex      int
	genres             []GenreInfo
	selectedGenre      string
	playerName         string
	nameInputActive    bool
	skillSchools       []SkillSchool
	remainingPoints    int
	startingEquipment  []StartingEquipment
	selectedEquipIndex int
	confirmed          bool
}

// NewCharacterCreation creates a new character creation screen.
func NewCharacterCreation() *CharacterCreation {
	cc := &CharacterCreation{
		state:           CCStateGenreSelection,
		selectedIndex:   0,
		playerName:      "Hero",
		remainingPoints: 10,
	}
	cc.initGenres()
	cc.initSkillSchools()
	cc.initEquipment()
	return cc
}

// initGenres sets up the available genres.
func (cc *CharacterCreation) initGenres() {
	cc.genres = []GenreInfo{
		{
			ID:          "fantasy",
			Name:        "Fantasy",
			Description: "A magic-infused medieval world of kingdoms, guilds, and ancient mysteries.",
		},
		{
			ID:          "sci-fi",
			Name:        "Sci-Fi",
			Description: "Explore a colony planet with corporations, military factions, and alien technology.",
		},
		{
			ID:          "horror",
			Name:        "Horror",
			Description: "Survive a cursed island plagued by cults, monsters, and unspeakable terrors.",
		},
		{
			ID:          "cyberpunk",
			Name:        "Cyberpunk",
			Description: "Navigate a neon-lit megacity of 2140 with megacorps, hackers, and street gangs.",
		},
		{
			ID:          "post-apocalyptic",
			Name:        "Post-Apocalyptic",
			Description: "Wander an irradiated wasteland of tribes, raiders, and forgotten technology.",
		},
	}
}

// initSkillSchools sets up the skill schools for allocation.
func (cc *CharacterCreation) initSkillSchools() {
	cc.skillSchools = []SkillSchool{
		{Name: "Combat", Description: "Melee and physical combat prowess", Points: 0},
		{Name: "Ranged", Description: "Archery, firearms, and thrown weapons", Points: 0},
		{Name: "Magic", Description: "Spellcasting and magical abilities", Points: 0},
		{Name: "Stealth", Description: "Sneaking, lockpicking, and thievery", Points: 0},
		{Name: "Speech", Description: "Persuasion, intimidation, and trading", Points: 0},
		{Name: "Survival", Description: "Crafting, alchemy, and wilderness skills", Points: 0},
	}
}

// initEquipment sets up starting equipment choices.
func (cc *CharacterCreation) initEquipment() {
	cc.startingEquipment = []StartingEquipment{
		{
			Name:        "Warrior's Kit",
			Description: "A sturdy sword and basic armor for combat.",
			Items:       []string{"Iron Sword", "Leather Armor", "Health Potion x2"},
		},
		{
			Name:        "Ranger's Kit",
			Description: "A bow with arrows and light gear for mobility.",
			Items:       []string{"Hunting Bow", "Arrows x20", "Scout Cloak", "Rations x5"},
		},
		{
			Name:        "Mage's Kit",
			Description: "A staff and robes for magical pursuits.",
			Items:       []string{"Apprentice Staff", "Mage Robes", "Mana Potion x2", "Spellbook"},
		},
		{
			Name:        "Rogue's Kit",
			Description: "Daggers and tools for stealth and subterfuge.",
			Items:       []string{"Steel Daggers x2", "Lockpicks x5", "Dark Cloak", "Smoke Bomb x2"},
		},
		{
			Name:        "Adventurer's Kit",
			Description: "A balanced loadout for any situation.",
			Items:       []string{"Shortsword", "Torch x3", "Rope", "Backpack", "Gold x50"},
		},
	}
}

// IsComplete returns true if character creation is finished.
func (cc *CharacterCreation) IsComplete() bool {
	return cc.state == CCStateComplete
}

// GetSelectedGenre returns the selected genre ID.
func (cc *CharacterCreation) GetSelectedGenre() string {
	return cc.selectedGenre
}

// GetPlayerName returns the entered player name.
func (cc *CharacterCreation) GetPlayerName() string {
	return cc.playerName
}

// Update processes input for character creation.
func (cc *CharacterCreation) Update() {
	switch cc.state {
	case CCStateGenreSelection:
		cc.updateGenreSelection()
	case CCStateNameInput:
		cc.updateNameInput()
	case CCStateSkillAllocation:
		cc.updateSkillAllocation()
	case CCStateEquipmentChoice:
		cc.updateEquipmentChoice()
	case CCStateConfirm:
		cc.updateConfirm()
	}
}

// updateGenreSelection handles genre selection input.
func (cc *CharacterCreation) updateGenreSelection() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		cc.selectedIndex--
		if cc.selectedIndex < 0 {
			cc.selectedIndex = len(cc.genres) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		cc.selectedIndex++
		if cc.selectedIndex >= len(cc.genres) {
			cc.selectedIndex = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		cc.selectedGenre = cc.genres[cc.selectedIndex].ID
		cc.state = CCStateNameInput
		cc.nameInputActive = true
	}
}

// updateNameInput handles name text input.
func (cc *CharacterCreation) updateNameInput() {
	// Handle backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(cc.playerName) > 0 {
			cc.playerName = cc.playerName[:len(cc.playerName)-1]
		}
	}

	// Handle character input
	chars := ebiten.AppendInputChars(nil)
	for _, c := range chars {
		if len(cc.playerName) < 20 { // Max name length
			cc.playerName += string(c)
		}
	}

	// Confirm name with Enter
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if len(cc.playerName) > 0 {
			cc.state = CCStateSkillAllocation
			cc.selectedIndex = 0
		}
	}

	// Go back with Escape
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.state = CCStateGenreSelection
	}
}

// updateSkillAllocation handles skill point allocation.
func (cc *CharacterCreation) updateSkillAllocation() {
	cc.handleSkillNavigation()
	cc.handleSkillPointAdjustment()
	cc.handleSkillAllocationConfirm()
}

// handleSkillNavigation processes up/down navigation through skill schools.
func (cc *CharacterCreation) handleSkillNavigation() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		cc.selectedIndex--
		if cc.selectedIndex < 0 {
			cc.selectedIndex = len(cc.skillSchools) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		cc.selectedIndex++
		if cc.selectedIndex >= len(cc.skillSchools) {
			cc.selectedIndex = 0
		}
	}
}

// handleSkillPointAdjustment processes adding/removing skill points.
func (cc *CharacterCreation) handleSkillPointAdjustment() {
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		if cc.remainingPoints > 0 && cc.skillSchools[cc.selectedIndex].Points < 5 {
			cc.skillSchools[cc.selectedIndex].Points++
			cc.remainingPoints--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		if cc.skillSchools[cc.selectedIndex].Points > 0 {
			cc.skillSchools[cc.selectedIndex].Points--
			cc.remainingPoints++
		}
	}
}

// handleSkillAllocationConfirm processes confirm and back actions.
func (cc *CharacterCreation) handleSkillAllocationConfirm() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		cc.state = CCStateEquipmentChoice
		cc.selectedIndex = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.state = CCStateNameInput
	}
}

// updateEquipmentChoice handles equipment selection.
func (cc *CharacterCreation) updateEquipmentChoice() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		cc.selectedIndex--
		if cc.selectedIndex < 0 {
			cc.selectedIndex = len(cc.startingEquipment) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		cc.selectedIndex++
		if cc.selectedIndex >= len(cc.startingEquipment) {
			cc.selectedIndex = 0
		}
	}

	// Confirm with Enter
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		cc.selectedEquipIndex = cc.selectedIndex
		cc.state = CCStateConfirm
	}

	// Go back with Escape
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.state = CCStateSkillAllocation
		cc.selectedIndex = 0
	}
}

// updateConfirm handles final confirmation.
func (cc *CharacterCreation) updateConfirm() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		cc.confirmed = true
		cc.state = CCStateComplete
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cc.state = CCStateEquipmentChoice
	}
}

// Draw renders the character creation screen.
func (cc *CharacterCreation) Draw(screen *ebiten.Image) {
	// Fill background
	screen.Fill(color.RGBA{20, 20, 30, 255})

	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	centerX := w / 2

	// Draw title
	title := "=== CHARACTER CREATION ==="
	ebitenutil.DebugPrintAt(screen, title, centerX-len(title)*3, 30)

	switch cc.state {
	case CCStateGenreSelection:
		cc.drawGenreSelection(screen, centerX, h)
	case CCStateNameInput:
		cc.drawNameInput(screen, centerX, h)
	case CCStateSkillAllocation:
		cc.drawSkillAllocation(screen, centerX, h)
	case CCStateEquipmentChoice:
		cc.drawEquipmentChoice(screen, centerX, h)
	case CCStateConfirm:
		cc.drawConfirm(screen, centerX, h)
	}
}

// drawGenreSelection renders the genre selection screen.
func (cc *CharacterCreation) drawGenreSelection(screen *ebiten.Image, centerX, h int) {
	subtitle := "Choose Your World"
	ebitenutil.DebugPrintAt(screen, subtitle, centerX-len(subtitle)*3, 70)

	startY := 120
	for i, genre := range cc.genres {
		label := genre.Name
		if i == cc.selectedIndex {
			label = "> " + label + " <"
		}
		x := centerX - len(label)*3
		ebitenutil.DebugPrintAt(screen, label, x, startY+i*30)
	}

	// Draw selected genre description
	if cc.selectedIndex < len(cc.genres) {
		desc := cc.genres[cc.selectedIndex].Description
		// Word wrap description
		maxWidth := 60
		y := startY + len(cc.genres)*30 + 40
		for len(desc) > 0 {
			line := desc
			if len(line) > maxWidth {
				line = line[:maxWidth]
				// Find last space for clean break
				for i := len(line) - 1; i > 0; i-- {
					if line[i] == ' ' {
						line = line[:i]
						break
					}
				}
			}
			ebitenutil.DebugPrintAt(screen, line, centerX-len(line)*3, y)
			desc = desc[len(line):]
			if len(desc) > 0 && desc[0] == ' ' {
				desc = desc[1:]
			}
			y += 15
		}
	}

	hints := "UP/DOWN: Select | ENTER: Confirm"
	ebitenutil.DebugPrintAt(screen, hints, centerX-len(hints)*3, h-30)
}

// drawNameInput renders the name input screen.
func (cc *CharacterCreation) drawNameInput(screen *ebiten.Image, centerX, h int) {
	subtitle := "Enter Your Name"
	ebitenutil.DebugPrintAt(screen, subtitle, centerX-len(subtitle)*3, 70)

	// Draw name input box
	nameLabel := fmt.Sprintf("Name: %s_", cc.playerName)
	ebitenutil.DebugPrintAt(screen, nameLabel, centerX-len(nameLabel)*3, 150)

	genreLabel := fmt.Sprintf("Genre: %s", cc.selectedGenre)
	ebitenutil.DebugPrintAt(screen, genreLabel, centerX-len(genreLabel)*3, 200)

	hints := "TYPE: Enter Name | ENTER: Confirm | ESC: Back"
	ebitenutil.DebugPrintAt(screen, hints, centerX-len(hints)*3, h-30)
}

// drawSkillAllocation renders the skill allocation screen.
func (cc *CharacterCreation) drawSkillAllocation(screen *ebiten.Image, centerX, h int) {
	subtitle := "Allocate Skill Points"
	ebitenutil.DebugPrintAt(screen, subtitle, centerX-len(subtitle)*3, 70)

	pointsLabel := fmt.Sprintf("Remaining Points: %d", cc.remainingPoints)
	ebitenutil.DebugPrintAt(screen, pointsLabel, centerX-len(pointsLabel)*3, 100)

	startY := 140
	for i, school := range cc.skillSchools {
		bar := ""
		for j := 0; j < 5; j++ {
			if j < school.Points {
				bar += "█"
			} else {
				bar += "░"
			}
		}
		label := fmt.Sprintf("%-10s [%s] %d", school.Name, bar, school.Points)
		if i == cc.selectedIndex {
			label = "> " + label + " <"
		}
		x := centerX - len(label)*3
		ebitenutil.DebugPrintAt(screen, label, x, startY+i*25)
	}

	// Description
	if cc.selectedIndex < len(cc.skillSchools) {
		desc := cc.skillSchools[cc.selectedIndex].Description
		ebitenutil.DebugPrintAt(screen, desc, centerX-len(desc)*3, startY+len(cc.skillSchools)*25+30)
	}

	hints := "UP/DOWN: Select | LEFT/RIGHT: Adjust | ENTER: Confirm | ESC: Back"
	ebitenutil.DebugPrintAt(screen, hints, centerX-len(hints)*3, h-30)
}

// drawEquipmentChoice renders the equipment selection screen.
func (cc *CharacterCreation) drawEquipmentChoice(screen *ebiten.Image, centerX, h int) {
	subtitle := "Choose Starting Equipment"
	ebitenutil.DebugPrintAt(screen, subtitle, centerX-len(subtitle)*3, 70)

	startY := 120
	for i, equip := range cc.startingEquipment {
		label := equip.Name
		if i == cc.selectedIndex {
			label = "> " + label + " <"
		}
		x := centerX - len(label)*3
		ebitenutil.DebugPrintAt(screen, label, x, startY+i*25)
	}

	// Show selected equipment details
	if cc.selectedIndex < len(cc.startingEquipment) {
		equip := cc.startingEquipment[cc.selectedIndex]
		y := startY + len(cc.startingEquipment)*25 + 30
		ebitenutil.DebugPrintAt(screen, equip.Description, centerX-len(equip.Description)*3, y)
		y += 25
		ebitenutil.DebugPrintAt(screen, "Contains:", centerX-50, y)
		y += 20
		for _, item := range equip.Items {
			itemLabel := "  - " + item
			ebitenutil.DebugPrintAt(screen, itemLabel, centerX-70, y)
			y += 15
		}
	}

	hints := "UP/DOWN: Select | ENTER: Confirm | ESC: Back"
	ebitenutil.DebugPrintAt(screen, hints, centerX-len(hints)*3, h-30)
}

// drawConfirm renders the confirmation screen.
func (cc *CharacterCreation) drawConfirm(screen *ebiten.Image, centerX, h int) {
	subtitle := "Confirm Your Character"
	ebitenutil.DebugPrintAt(screen, subtitle, centerX-len(subtitle)*3, 70)

	y := 120
	nameLabel := fmt.Sprintf("Name: %s", cc.playerName)
	ebitenutil.DebugPrintAt(screen, nameLabel, centerX-len(nameLabel)*3, y)
	y += 25

	genreLabel := fmt.Sprintf("World: %s", cc.selectedGenre)
	ebitenutil.DebugPrintAt(screen, genreLabel, centerX-len(genreLabel)*3, y)
	y += 25

	equipLabel := fmt.Sprintf("Equipment: %s", cc.startingEquipment[cc.selectedEquipIndex].Name)
	ebitenutil.DebugPrintAt(screen, equipLabel, centerX-len(equipLabel)*3, y)
	y += 40

	ebitenutil.DebugPrintAt(screen, "Skills:", centerX-30, y)
	y += 20
	for _, school := range cc.skillSchools {
		if school.Points > 0 {
			skillLabel := fmt.Sprintf("  %s: %d", school.Name, school.Points)
			ebitenutil.DebugPrintAt(screen, skillLabel, centerX-50, y)
			y += 15
		}
	}

	confirmLabel := "Press ENTER to begin your adventure!"
	ebitenutil.DebugPrintAt(screen, confirmLabel, centerX-len(confirmLabel)*3, h-80)

	hints := "ENTER: Start Game | ESC: Go Back"
	ebitenutil.DebugPrintAt(screen, hints, centerX-len(hints)*3, h-30)
}

// ApplyToPlayer applies character creation choices to a player entity.
func (cc *CharacterCreation) ApplyToPlayer(world *ecs.World, playerEntity ecs.Entity) {
	// Apply skill allocations
	skillsComp, ok := world.GetComponent(playerEntity, "Skills")
	if ok {
		skills := skillsComp.(*components.Skills)
		for _, school := range cc.skillSchools {
			if school.Points > 0 {
				skills.Levels[school.Name] = school.Points
			}
		}
	}

	// Apply starting equipment to inventory
	invComp, ok := world.GetComponent(playerEntity, "Inventory")
	if ok {
		inv := invComp.(*components.Inventory)
		equip := cc.startingEquipment[cc.selectedEquipIndex]
		inv.Items = append(inv.Items, equip.Items...)
	}

	// Apply name (store in a Name component or use player entity ID)
	// For now, name is stored locally and used for display
}
