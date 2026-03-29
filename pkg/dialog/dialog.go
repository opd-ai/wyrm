// Package dialog provides deep NPC dialog with topic memory and emotional states.
// Per ROADMAP Phase 6 item 26:
// AC: NPC recalls player's previous interaction topic in follow-up conversation;
// emotional state (fearful/hostile/friendly) changes NPC response vocabulary;
// unit test asserts all 5 genres produce non-overlapping common word sets.
package dialog

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// EmotionalState represents an NPC's emotional disposition.
type EmotionalState int

const (
	EmotionNeutral EmotionalState = iota
	EmotionFriendly
	EmotionHostile
	EmotionFearful
	EmotionSuspicious
)

// String returns the human-readable name of the emotional state.
func (e EmotionalState) String() string {
	switch e {
	case EmotionFriendly:
		return "friendly"
	case EmotionHostile:
		return "hostile"
	case EmotionFearful:
		return "fearful"
	case EmotionSuspicious:
		return "suspicious"
	default:
		return "neutral"
	}
}

// TopicMemory stores what topics have been discussed with a player.
type TopicMemory struct {
	Topic        string
	Timestamp    time.Time
	PlayerAction string // What the player did during the conversation
	NPCResponse  string // How the NPC responded
}

// NPCMemory tracks conversation history with a specific NPC.
type NPCMemory struct {
	NPCID           uint64
	PlayerID        uint64
	Topics          []TopicMemory
	EmotionShift    float64 // Cumulative emotional modifier
	LastInteraction time.Time
}

// DialogManager handles NPC conversations with memory.
type DialogManager struct {
	mu       sync.RWMutex
	memories map[uint64]map[uint64]*NPCMemory // NPCID -> PlayerID -> Memory
	rng      *rand.Rand
}

// NewDialogManager creates a new dialog manager.
func NewDialogManager(seed int64) *DialogManager {
	return &DialogManager{
		memories: make(map[uint64]map[uint64]*NPCMemory),
		rng:      rand.New(rand.NewSource(seed)),
	}
}

// RecordTopic stores a conversation topic in memory.
func (dm *DialogManager) RecordTopic(npcID, playerID uint64, topic, playerAction, npcResponse string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.memories[npcID] == nil {
		dm.memories[npcID] = make(map[uint64]*NPCMemory)
	}

	memory := dm.memories[npcID][playerID]
	if memory == nil {
		memory = &NPCMemory{
			NPCID:    npcID,
			PlayerID: playerID,
			Topics:   make([]TopicMemory, 0),
		}
		dm.memories[npcID][playerID] = memory
	}

	memory.Topics = append(memory.Topics, TopicMemory{
		Topic:        topic,
		Timestamp:    time.Now(),
		PlayerAction: playerAction,
		NPCResponse:  npcResponse,
	})
	memory.LastInteraction = time.Now()

	// Keep only last 20 topics
	if len(memory.Topics) > 20 {
		memory.Topics = memory.Topics[len(memory.Topics)-20:]
	}
}

// GetLastTopic returns the most recent topic discussed with an NPC.
func (dm *DialogManager) GetLastTopic(npcID, playerID uint64) *TopicMemory {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.memories[npcID] == nil {
		return nil
	}
	memory := dm.memories[npcID][playerID]
	if memory == nil || len(memory.Topics) == 0 {
		return nil
	}
	return &memory.Topics[len(memory.Topics)-1]
}

// HasDiscussedTopic checks if a topic has been discussed before.
func (dm *DialogManager) HasDiscussedTopic(npcID, playerID uint64, topic string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.memories[npcID] == nil {
		return false
	}
	memory := dm.memories[npcID][playerID]
	if memory == nil {
		return false
	}

	topic = strings.ToLower(topic)
	for _, t := range memory.Topics {
		if strings.ToLower(t.Topic) == topic {
			return true
		}
	}
	return false
}

// GetTopicHistory returns all topics discussed about a specific subject.
func (dm *DialogManager) GetTopicHistory(npcID, playerID uint64) []TopicMemory {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.memories[npcID] == nil {
		return nil
	}
	memory := dm.memories[npcID][playerID]
	if memory == nil {
		return nil
	}
	return memory.Topics
}

// ShiftEmotion modifies an NPC's emotional state toward a player.
func (dm *DialogManager) ShiftEmotion(npcID, playerID uint64, shift float64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.memories[npcID] == nil {
		dm.memories[npcID] = make(map[uint64]*NPCMemory)
	}

	memory := dm.memories[npcID][playerID]
	if memory == nil {
		memory = &NPCMemory{
			NPCID:    npcID,
			PlayerID: playerID,
			Topics:   make([]TopicMemory, 0),
		}
		dm.memories[npcID][playerID] = memory
	}

	memory.EmotionShift += shift
	// Clamp to -100 to +100
	if memory.EmotionShift > 100 {
		memory.EmotionShift = 100
	}
	if memory.EmotionShift < -100 {
		memory.EmotionShift = -100
	}
}

// GetEmotionalState returns the NPC's emotional state toward a player.
func (dm *DialogManager) GetEmotionalState(npcID, playerID uint64, baseState EmotionalState) EmotionalState {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	memory := dm.getMemory(npcID, playerID)
	if memory == nil {
		return baseState
	}
	return dm.computeEmotionalState(memory.EmotionShift, baseState)
}

// getMemory retrieves the memory for an NPC-player pair, or nil if none exists.
func (dm *DialogManager) getMemory(npcID, playerID uint64) *NPCMemory {
	if dm.memories[npcID] == nil {
		return nil
	}
	return dm.memories[npcID][playerID]
}

// computeEmotionalState determines emotional state based on shift value.
func (dm *DialogManager) computeEmotionalState(shift float64, baseState EmotionalState) EmotionalState {
	if shift >= 30 {
		return EmotionFriendly
	}
	if shift <= -70 {
		return EmotionFearful
	}
	if shift <= -50 {
		return EmotionHostile
	}
	if shift <= -30 {
		return EmotionSuspicious
	}
	return baseState
}

// GenreVocabulary defines genre-specific common words.
type GenreVocabulary struct {
	Greetings    []string
	Farewells    []string
	Affirmatives []string
	Negatives    []string
	Expletives   []string
	Titles       []string
	CommonWords  []string
}

// GenreVocabularies maps genre to vocabulary sets.
// Per AC: all 5 genres produce non-overlapping common word sets.
var GenreVocabularies = map[string]*GenreVocabulary{
	"fantasy": {
		Greetings:    []string{"Hail", "Well met", "Greetings", "Good morrow", "Salutations"},
		Farewells:    []string{"Fare thee well", "Until next we meet", "Safe travels", "Gods watch over you"},
		Affirmatives: []string{"Aye", "Indeed", "Verily", "'Tis so", "As you say"},
		Negatives:    []string{"Nay", "Alas, no", "I think not", "By the gods, no"},
		Expletives:   []string{"By the gods!", "Blazes!", "Curses!", "Forsooth!"},
		Titles:       []string{"milord", "milady", "ser", "goodman", "traveler"},
		CommonWords:  []string{"quest", "dungeon", "kingdom", "magic", "dragon", "sword", "potion", "spell", "guild", "tavern"},
	},
	"sci-fi": {
		Greetings:    []string{"Greetings", "Acknowledged", "Welcome aboard", "Scanner reads friendly"},
		Farewells:    []string{"Clear skies", "Stay frosty", "Transmission ended", "Good hunting"},
		Affirmatives: []string{"Affirmative", "Confirmed", "Roger that", "Copy that"},
		Negatives:    []string{"Negative", "That's a no-go", "Denied", "Abort that thought"},
		Expletives:   []string{"Void!", "Damn the core!", "Frakking!", "By the stars!"},
		Titles:       []string{"commander", "citizen", "operative", "pilot", "tech"},
		CommonWords:  []string{"sector", "hyperspace", "credits", "plasma", "colony", "android", "starship", "quantum", "megacorp", "datalink"},
	},
	"horror": {
		Greetings:    []string{"You're still alive...", "Another soul...", "Don't startle me", "Who goes there?"},
		Farewells:    []string{"May you survive", "Don't look back", "Stay in the light", "Pray we meet again"},
		Affirmatives: []string{"Yes... yes", "I suppose so", "If you say so", "I... agree"},
		Negatives:    []string{"No... never", "I dare not", "Please, no", "I can't... I won't"},
		Expletives:   []string{"Dear god!", "Heaven help us!", "What fresh hell?!", "No... NO!"},
		Titles:       []string{"stranger", "survivor", "poor soul", "wanderer", "lost one"},
		CommonWords:  []string{"darkness", "ritual", "curse", "asylum", "madness", "blood", "sanctuary", "whispers", "haunted", "sacrifice"},
	},
	"cyberpunk": {
		Greetings:    []string{"'Sup choom", "Scan's clean", "What's the biz", "Jack in, samurai"},
		Farewells:    []string{"Stay chrome", "Flatline 'em", "Catch you on the net", "Don't get zeroed"},
		Affirmatives: []string{"Preem", "Solid", "No delta", "That's nova"},
		Negatives:    []string{"Nah, choom", "Hard pass", "No dice", "That's gonk"},
		Expletives:   []string{"Drek!", "Frag it!", "Slot!", "Null sweat!"},
		Titles:       []string{"choom", "netrunner", "edgerunner", "fixer", "corpo"},
		CommonWords:  []string{"eddies", "chrome", "netspace", "flatline", "implant", "synth", "megablock", "braindance", "cyberware", "glitch"},
	},
	"post-apocalyptic": {
		Greetings:    []string{"You breathin'?", "Friendly?", "Wastelander?", "State your business"},
		Farewells:    []string{"Stay rad-free", "Don't rust", "Keep moving", "Watch the wastes"},
		Affirmatives: []string{"Sure thing", "Damn right", "Reckon so", "That'll work"},
		Negatives:    []string{"Hell no", "Not a chance", "Forget it", "That's suicide"},
		Expletives:   []string{"Rad storm!", "Mutant spit!", "Holy fallout!", "Damn the bombs!"},
		Titles:       []string{"wastelander", "drifter", "vault dweller", "raider", "scav"},
		CommonWords:  []string{"bunker", "radiation", "mutant", "wasteland", "scavenge", "settlement", "caravan", "purifier", "rad", "tribe"},
	},
}

// GetVocabulary returns the vocabulary for a genre.
func GetVocabulary(genre string) *GenreVocabulary {
	if vocab, ok := GenreVocabularies[genre]; ok {
		return vocab
	}
	return GenreVocabularies["fantasy"]
}

// EmotionVocabularyModifiers adjusts vocabulary based on emotional state.
type EmotionVocabularyModifiers struct {
	Prefixes []string
	Tone     string
}

// EmotionModifiers maps emotional state to speech patterns.
var EmotionModifiers = map[EmotionalState]EmotionVocabularyModifiers{
	EmotionFriendly: {
		Prefixes: []string{"Ah, friend!", "Good to see you!", "My dear companion,"},
		Tone:     "warm",
	},
	EmotionHostile: {
		Prefixes: []string{"You...", "What do YOU want?", "Leave me be,"},
		Tone:     "aggressive",
	},
	EmotionFearful: {
		Prefixes: []string{"P-please...", "D-don't hurt me!", "I... I..."},
		Tone:     "trembling",
	},
	EmotionSuspicious: {
		Prefixes: []string{"Hmm...", "I'm watching you,", "Don't try anything,"},
		Tone:     "guarded",
	},
	EmotionNeutral: {
		Prefixes: []string{"", "Hmm,", "Yes?"},
		Tone:     "neutral",
	},
}

// DialogResponse represents a generated NPC response.
type DialogResponse struct {
	Text          string
	EmotionalTone string
	RecalledTopic string // If this response references past conversation
}

// GenerateResponse creates an NPC response based on context.
func (dm *DialogManager) GenerateResponse(
	npcID, playerID uint64,
	genre string,
	topic string,
	emotion EmotionalState,
) *DialogResponse {
	vocab := GetVocabulary(genre)
	mods := EmotionModifiers[emotion]

	response := &DialogResponse{
		EmotionalTone: mods.Tone,
	}

	// Check for topic recall
	lastTopic := dm.GetLastTopic(npcID, playerID)
	if lastTopic != nil && time.Since(lastTopic.Timestamp) < 24*time.Hour {
		response.RecalledTopic = lastTopic.Topic
	}

	// Build response based on emotion and vocabulary
	var parts []string

	// Add emotional prefix
	if len(mods.Prefixes) > 0 {
		parts = append(parts, mods.Prefixes[dm.rng.Intn(len(mods.Prefixes))])
	}

	// Reference past conversation if available
	if response.RecalledTopic != "" {
		parts = append(parts, fmt.Sprintf("About %s we discussed earlier...", response.RecalledTopic))
	}

	// Add topic-appropriate response
	if len(vocab.CommonWords) > 0 {
		word := vocab.CommonWords[dm.rng.Intn(len(vocab.CommonWords))]
		parts = append(parts, fmt.Sprintf("Regarding the %s...", word))
	}

	response.Text = strings.Join(parts, " ")
	return response
}

// GetGreeting returns a genre-appropriate greeting.
func GetGreeting(genre string, emotion EmotionalState, rng *rand.Rand) string {
	vocab := GetVocabulary(genre)
	mods := EmotionModifiers[emotion]

	greeting := vocab.Greetings[rng.Intn(len(vocab.Greetings))]

	if len(mods.Prefixes) > 0 && mods.Prefixes[0] != "" {
		prefix := mods.Prefixes[rng.Intn(len(mods.Prefixes))]
		return prefix + " " + greeting
	}
	return greeting
}

// GetFarewell returns a genre-appropriate farewell.
func GetFarewell(genre string, emotion EmotionalState, rng *rand.Rand) string {
	vocab := GetVocabulary(genre)
	return vocab.Farewells[rng.Intn(len(vocab.Farewells))]
}

// MemoryCount returns the number of NPC-player memory pairs.
func (dm *DialogManager) MemoryCount() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	count := 0
	for _, playerMemories := range dm.memories {
		count += len(playerMemories)
	}
	return count
}

// ClearOldMemories removes memories older than the given duration.
func (dm *DialogManager) ClearOldMemories(maxAge time.Duration) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	now := time.Now()
	for npcID, playerMemories := range dm.memories {
		for playerID, memory := range playerMemories {
			if now.Sub(memory.LastInteraction) > maxAge {
				delete(playerMemories, playerID)
			}
		}
		if len(playerMemories) == 0 {
			delete(dm.memories, npcID)
		}
	}
}
