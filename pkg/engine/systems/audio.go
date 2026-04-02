package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// AudioSystem drives procedural audio synthesis and spatial audio.
type AudioSystem struct {
	Genre string
	// CombatDetectionRange is the distance to detect combat for music intensity.
	CombatDetectionRange float64
	// AmbientUpdateInterval is seconds between ambient sound checks.
	AmbientUpdateInterval float64
	// timeAccum tracks time for periodic ambient updates.
	timeAccum float64
}

// NewAudioSystem creates a new audio system with default settings.
func NewAudioSystem(genre string) *AudioSystem {
	return &AudioSystem{
		Genre:                 genre,
		CombatDetectionRange:  DefaultCombatDetectionRange,
		AmbientUpdateInterval: DefaultAmbientUpdateInterval,
		timeAccum:             0,
	}
}

// Update advances audio synthesis based on player position and game state.
func (s *AudioSystem) Update(w *ecs.World, dt float64) {
	s.timeAccum += dt

	// Find the audio listener (typically the player)
	listenerPos, listenerFound := s.findListenerPosition(w)
	if !listenerFound {
		return
	}

	// Update audio state component if it exists
	s.updateAudioState(w, listenerPos)

	// Process audio sources for spatial audio calculations
	s.processSpatialAudio(w, listenerPos)

	// Periodically update ambient sounds
	if s.timeAccum >= s.AmbientUpdateInterval {
		s.timeAccum = 0
		s.updateAmbientSounds(w, listenerPos)
	}
}

// findListenerPosition locates the audio listener entity and returns its position.
func (s *AudioSystem) findListenerPosition(w *ecs.World) ([2]float64, bool) {
	for _, e := range w.Entities("AudioListener", "Position") {
		listenerComp, ok := w.GetComponent(e, "AudioListener")
		if !ok {
			continue
		}
		listener := listenerComp.(*components.AudioListener)
		if !listener.Enabled {
			continue
		}
		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)
		return [2]float64{pos.X, pos.Y}, true
	}
	return [2]float64{}, false
}

// updateAudioState updates the AudioState component with current conditions.
func (s *AudioSystem) updateAudioState(w *ecs.World, listenerPos [2]float64) {
	// Calculate combat intensity based on nearby hostile entities
	combatIntensity := s.calculateCombatIntensity(w, listenerPos)

	// Update all AudioState components
	for _, e := range w.Entities("AudioState") {
		comp, ok := w.GetComponent(e, "AudioState")
		if !ok {
			continue
		}
		state := comp.(*components.AudioState)
		state.CombatIntensity = combatIntensity
		state.LastPositionX = listenerPos[0]
		state.LastPositionY = listenerPos[1]
	}
}

// calculateCombatIntensity returns 0.0-1.0 based on nearby hostile entities.
func (s *AudioSystem) calculateCombatIntensity(w *ecs.World, listenerPos [2]float64) float64 {
	hostileCount := s.countNearbyHostiles(w, listenerPos)

	if hostileCount >= MaxHostilesForIntensity {
		return MaxCombatIntensity
	}
	return float64(hostileCount) / float64(MaxHostilesForIntensity)
}

// countNearbyHostiles counts hostile entities within detection range.
func (s *AudioSystem) countNearbyHostiles(w *ecs.World, listenerPos [2]float64) int {
	hostileCount := 0
	rangeSquared := s.CombatDetectionRange * s.CombatDetectionRange

	for _, e := range w.Entities("Health", "Position", "Faction") {
		if s.isEntityHostileAndNearby(w, e, listenerPos, rangeSquared) {
			hostileCount++
		}
	}
	return hostileCount
}

// isEntityHostileAndNearby checks if an entity is both hostile and within range.
func (s *AudioSystem) isEntityHostileAndNearby(w *ecs.World, e ecs.Entity, listenerPos [2]float64, rangeSquared float64) bool {
	posComp, ok := w.GetComponent(e, "Position")
	if !ok {
		return false
	}
	pos := posComp.(*components.Position)

	if !s.isWithinRange(pos, listenerPos, rangeSquared) {
		return false
	}

	return s.isEntityHostile(w, e)
}

// isWithinRange checks if a position is within squared range of listener.
func (s *AudioSystem) isWithinRange(pos *components.Position, listenerPos [2]float64, rangeSquared float64) bool {
	dx := pos.X - listenerPos[0]
	dy := pos.Y - listenerPos[1]
	distSq := dx*dx + dy*dy
	return distSq <= rangeSquared
}

// isEntityHostile checks if an entity has hostile faction reputation.
func (s *AudioSystem) isEntityHostile(w *ecs.World, e ecs.Entity) bool {
	factionComp, ok := w.GetComponent(e, "Faction")
	if !ok {
		return false
	}
	faction := factionComp.(*components.Faction)
	return faction.Reputation < HostileFactionThreshold
}

// processSpatialAudio calculates volume/pan for audio sources based on distance.
func (s *AudioSystem) processSpatialAudio(w *ecs.World, listenerPos [2]float64) {
	for _, e := range w.Entities("AudioSource", "Position") {
		sourceComp, ok := w.GetComponent(e, "AudioSource")
		if !ok {
			continue
		}
		source := sourceComp.(*components.AudioSource)
		if !source.Playing {
			continue
		}

		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)

		// Calculate distance-based attenuation
		dx := pos.X - listenerPos[0]
		dy := pos.Y - listenerPos[1]
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist >= source.Range {
			// Out of range - effectively muted
			continue
		}

		// Linear falloff for now (could be improved to inverse-square)
		attenuation := LinearFalloffBase - (dist / source.Range)
		// TODO: Apply attenuated volume (source.Volume * attenuation) to audio engine
		// when actual audio playback is implemented
		source.EffectiveVolume = source.Volume * attenuation
	}
}

// updateAmbientSounds selects appropriate ambient sound based on environment.
func (s *AudioSystem) updateAmbientSounds(w *ecs.World, listenerPos [2]float64) {
	// Determine ambient type based on location and world state
	ambientType := s.selectAmbientType(w, listenerPos)

	// Update AudioState with new ambient
	for _, e := range w.Entities("AudioState") {
		comp, ok := w.GetComponent(e, "AudioState")
		if !ok {
			continue
		}
		state := comp.(*components.AudioState)
		state.CurrentAmbient = ambientType
	}
}

// selectAmbientType chooses the ambient sound type based on location.
func (s *AudioSystem) selectAmbientType(w *ecs.World, listenerPos [2]float64) string {
	// Check if in a city (near EconomyNode entities)
	for _, e := range w.Entities("EconomyNode", "Position") {
		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)
		dx := pos.X - listenerPos[0]
		dy := pos.Y - listenerPos[1]
		if dx*dx+dy*dy < CityProximityRange*CityProximityRange { // Within city range
			return s.getCityAmbient()
		}
	}

	// Default to wilderness ambient
	return s.getWildernessAmbient()
}

// getCityAmbient returns the genre-appropriate city ambient sound type.
func (s *AudioSystem) getCityAmbient() string {
	switch s.Genre {
	case "fantasy":
		return "city_medieval"
	case "sci-fi":
		return "city_station"
	case "horror":
		return "city_abandoned"
	case "cyberpunk":
		return "city_neon"
	case "post-apocalyptic":
		return "city_ruins"
	default:
		return "city_generic"
	}
}

// getWildernessAmbient returns the genre-appropriate wilderness ambient sound type.
func (s *AudioSystem) getWildernessAmbient() string {
	switch s.Genre {
	case "fantasy":
		return "wilderness_forest"
	case "sci-fi":
		return "wilderness_alien"
	case "horror":
		return "wilderness_dark"
	case "cyberpunk":
		return "wilderness_industrial"
	case "post-apocalyptic":
		return "wilderness_wasteland"
	default:
		return "wilderness_generic"
	}
}
