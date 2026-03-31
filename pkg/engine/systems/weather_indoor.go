package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// LocationType represents where an entity is located.
type LocationType int

const (
	LocationOutdoor LocationType = iota
	LocationIndoor
	LocationUnderground
	LocationUnderwater
)

// IndoorOutdoorZone represents a zone with indoor/outdoor properties.
type IndoorOutdoorZone struct {
	ID               string
	LocationType     LocationType
	MinX, MinY, MinZ float64
	MaxX, MaxY, MaxZ float64
	WeatherShielded  bool
	LightOverride    float64
	AmbientSound     string
}

// IndoorOutdoorSystem detects whether entities are inside or outside.
type IndoorOutdoorSystem struct {
	Zones       map[string]*IndoorOutdoorZone
	EntityZones map[ecs.Entity]string
	DefaultType LocationType
	weatherSys  *WeatherSystem
}

// NewIndoorOutdoorSystem creates a new indoor/outdoor detection system.
func NewIndoorOutdoorSystem(weatherSys *WeatherSystem) *IndoorOutdoorSystem {
	return &IndoorOutdoorSystem{
		Zones:       make(map[string]*IndoorOutdoorZone),
		EntityZones: make(map[ecs.Entity]string),
		DefaultType: LocationOutdoor,
		weatherSys:  weatherSys,
	}
}

// Update checks all entities' positions and updates their location status.
func (s *IndoorOutdoorSystem) Update(w *ecs.World, dt float64) {
	if w == nil {
		return
	}
	for _, e := range w.Entities("Position") {
		s.updateEntityLocation(w, e)
	}
}

// updateEntityLocation determines which zone an entity is in.
func (s *IndoorOutdoorSystem) updateEntityLocation(w *ecs.World, e ecs.Entity) {
	posComp, ok := w.GetComponent(e, "Position")
	if !ok {
		return
	}

	type positioner interface {
		GetX() float64
		GetY() float64
		GetZ() float64
	}

	if pos, ok := posComp.(positioner); ok {
		x, y, z := pos.GetX(), pos.GetY(), pos.GetZ()
		for id, zone := range s.Zones {
			if s.isInZone(x, y, z, zone) {
				s.EntityZones[e] = id
				return
			}
		}
	}
	delete(s.EntityZones, e)
}

// isInZone checks if coordinates are within a zone's bounds.
func (s *IndoorOutdoorSystem) isInZone(x, y, z float64, zone *IndoorOutdoorZone) bool {
	return x >= zone.MinX && x <= zone.MaxX &&
		y >= zone.MinY && y <= zone.MaxY &&
		z >= zone.MinZ && z <= zone.MaxZ
}

// RegisterZone adds a new zone to the system.
func (s *IndoorOutdoorSystem) RegisterZone(zone *IndoorOutdoorZone) {
	s.Zones[zone.ID] = zone
}

// UnregisterZone removes a zone from the system.
func (s *IndoorOutdoorSystem) UnregisterZone(id string) {
	delete(s.Zones, id)
	for e, zoneID := range s.EntityZones {
		if zoneID == id {
			delete(s.EntityZones, e)
		}
	}
}

// GetEntityLocationType returns the location type for an entity.
func (s *IndoorOutdoorSystem) GetEntityLocationType(e ecs.Entity) LocationType {
	if zoneID, ok := s.EntityZones[e]; ok {
		if zone, ok := s.Zones[zoneID]; ok {
			return zone.LocationType
		}
	}
	return s.DefaultType
}

// GetEntityZone returns the zone ID for an entity, or empty string if outside.
func (s *IndoorOutdoorSystem) GetEntityZone(e ecs.Entity) string {
	return s.EntityZones[e]
}

// IsEntityIndoor checks if an entity is in an indoor location.
func (s *IndoorOutdoorSystem) IsEntityIndoor(e ecs.Entity) bool {
	return s.GetEntityLocationType(e) == LocationIndoor
}

// IsEntityOutdoor checks if an entity is in an outdoor location.
func (s *IndoorOutdoorSystem) IsEntityOutdoor(e ecs.Entity) bool {
	return s.GetEntityLocationType(e) == LocationOutdoor
}

// IsEntityUnderground checks if an entity is underground.
func (s *IndoorOutdoorSystem) IsEntityUnderground(e ecs.Entity) bool {
	return s.GetEntityLocationType(e) == LocationUnderground
}

// IsEntityUnderwater checks if an entity is underwater.
func (s *IndoorOutdoorSystem) IsEntityUnderwater(e ecs.Entity) bool {
	return s.GetEntityLocationType(e) == LocationUnderwater
}

// IsWeatherShielded checks if an entity is protected from weather effects.
func (s *IndoorOutdoorSystem) IsWeatherShielded(e ecs.Entity) bool {
	if zoneID, ok := s.EntityZones[e]; ok {
		if zone, ok := s.Zones[zoneID]; ok {
			return zone.WeatherShielded
		}
	}
	return false
}

// GetEffectiveWeatherModifiers returns weather modifiers adjusted for location.
func (s *IndoorOutdoorSystem) GetEffectiveWeatherModifiers(e ecs.Entity) WeatherModifiers {
	if s.weatherSys == nil {
		return WeatherModifiers{
			Visibility: 1.0,
			Movement:   1.0,
			Accuracy:   1.0,
			Damage:     0.0,
			Stealth:    1.0,
		}
	}

	baseMods := s.weatherSys.GetWeatherModifiers()

	if s.IsWeatherShielded(e) {
		return WeatherModifiers{
			Visibility: 1.0,
			Movement:   1.0,
			Accuracy:   1.0,
			Damage:     0.0,
			Stealth:    baseMods.Stealth,
		}
	}

	locType := s.GetEntityLocationType(e)
	switch locType {
	case LocationUnderground:
		return WeatherModifiers{
			Visibility: 0.5,
			Movement:   1.0,
			Accuracy:   1.0,
			Damage:     0.0,
			Stealth:    0.7,
		}
	case LocationUnderwater:
		return WeatherModifiers{
			Visibility: 0.4,
			Movement:   0.6,
			Accuracy:   0.5,
			Damage:     0.0,
			Stealth:    0.8,
		}
	}

	return baseMods
}

// GetEffectiveLightLevel returns the light level adjusted for location.
func (s *IndoorOutdoorSystem) GetEffectiveLightLevel(e ecs.Entity, hour int) float64 {
	if zoneID, ok := s.EntityZones[e]; ok {
		if zone, ok := s.Zones[zoneID]; ok {
			if zone.LightOverride > 0 {
				return zone.LightOverride
			}
			switch zone.LocationType {
			case LocationIndoor:
				return 0.7
			case LocationUnderground:
				return 0.2
			case LocationUnderwater:
				return 0.4
			}
		}
	}

	if s.weatherSys != nil {
		return s.weatherSys.GetLightLevel(hour)
	}
	return 1.0
}

// GetAmbientSound returns the appropriate ambient sound for an entity's location.
func (s *IndoorOutdoorSystem) GetAmbientSound(e ecs.Entity) string {
	if sound := s.getZoneAmbientSound(e); sound != "" {
		return sound
	}
	return s.getWeatherAmbientSound()
}

// getZoneAmbientSound returns the ambient sound for an entity's zone.
func (s *IndoorOutdoorSystem) getZoneAmbientSound(e ecs.Entity) string {
	zoneID, ok := s.EntityZones[e]
	if !ok {
		return ""
	}
	zone, ok := s.Zones[zoneID]
	if !ok {
		return ""
	}
	if zone.AmbientSound != "" {
		return zone.AmbientSound
	}
	return locationTypeToAmbient(zone.LocationType)
}

// locationTypeToAmbient maps location types to their default ambient sounds.
func locationTypeToAmbient(locType LocationType) string {
	switch locType {
	case LocationIndoor:
		return "ambient_indoor"
	case LocationUnderground:
		return "ambient_cave"
	case LocationUnderwater:
		return "ambient_underwater"
	default:
		return ""
	}
}

// getWeatherAmbientSound returns the ambient sound based on current weather.
func (s *IndoorOutdoorSystem) getWeatherAmbientSound() string {
	if s.weatherSys == nil {
		return "ambient_outdoor"
	}
	switch s.weatherSys.CurrentWeather {
	case "rain", "thunderstorm", "acid_rain":
		return "ambient_rain"
	case "fog", "mist":
		return "ambient_fog"
	case "dust_storm", "ash_fall":
		return "ambient_wind"
	default:
		return "ambient_outdoor"
	}
}

// zoneDefaults holds the non-positional defaults for a zone type.
type zoneDefaults struct {
	locationType    LocationType
	weatherShielded bool
	lightOverride   float64
	ambientSound    string
}

// zoneTypeDefaults maps location types to their standard defaults.
var zoneTypeDefaults = map[LocationType]zoneDefaults{
	LocationIndoor:      {LocationIndoor, true, 0.7, "ambient_indoor"},
	LocationUnderground: {LocationUnderground, true, 0.2, "ambient_cave"},
	LocationUnderwater:  {LocationUnderwater, true, 0.4, "ambient_underwater"},
}

// createZoneWithDefaults creates a zone with standard defaults for the given type.
func (s *IndoorOutdoorSystem) createZoneWithDefaults(id string, locType LocationType, minX, minY, minZ, maxX, maxY, maxZ float64) *IndoorOutdoorZone {
	defaults := zoneTypeDefaults[locType]
	zone := &IndoorOutdoorZone{
		ID:              id,
		LocationType:    defaults.locationType,
		MinX:            minX,
		MinY:            minY,
		MinZ:            minZ,
		MaxX:            maxX,
		MaxY:            maxY,
		MaxZ:            maxZ,
		WeatherShielded: defaults.weatherShielded,
		LightOverride:   defaults.lightOverride,
		AmbientSound:    defaults.ambientSound,
	}
	s.RegisterZone(zone)
	return zone
}

// CreateBuildingZone creates a standard indoor zone for a building.
func (s *IndoorOutdoorSystem) CreateBuildingZone(id string, minX, minY, minZ, maxX, maxY, maxZ float64) *IndoorOutdoorZone {
	return s.createZoneWithDefaults(id, LocationIndoor, minX, minY, minZ, maxX, maxY, maxZ)
}

// CreateCaveZone creates a standard underground zone.
func (s *IndoorOutdoorSystem) CreateCaveZone(id string, minX, minY, minZ, maxX, maxY, maxZ float64) *IndoorOutdoorZone {
	return s.createZoneWithDefaults(id, LocationUnderground, minX, minY, minZ, maxX, maxY, maxZ)
}

// CreateUnderwaterZone creates a standard underwater zone.
func (s *IndoorOutdoorSystem) CreateUnderwaterZone(id string, minX, minY, minZ, maxX, maxY, maxZ float64) *IndoorOutdoorZone {
	return s.createZoneWithDefaults(id, LocationUnderwater, minX, minY, minZ, maxX, maxY, maxZ)
}

// GetZoneCount returns the number of registered zones.
func (s *IndoorOutdoorSystem) GetZoneCount() int {
	return len(s.Zones)
}

// GetTrackedEntityCount returns the number of entities in zones.
func (s *IndoorOutdoorSystem) GetTrackedEntityCount() int {
	return len(s.EntityZones)
}

// ClearEntityTracking removes all entity zone associations.
func (s *IndoorOutdoorSystem) ClearEntityTracking() {
	s.EntityZones = make(map[ecs.Entity]string)
}

// GetZone returns a zone by ID.
func (s *IndoorOutdoorSystem) GetZone(id string) *IndoorOutdoorZone {
	return s.Zones[id]
}

// SetEntityZone manually sets an entity's zone.
func (s *IndoorOutdoorSystem) SetEntityZone(e ecs.Entity, zoneID string) bool {
	if _, ok := s.Zones[zoneID]; ok {
		s.EntityZones[e] = zoneID
		return true
	}
	return false
}

// ClearEntityZone removes an entity's zone association.
func (s *IndoorOutdoorSystem) ClearEntityZone(e ecs.Entity) {
	delete(s.EntityZones, e)
}
