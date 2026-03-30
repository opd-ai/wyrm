//go:build !noebiten

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"
	"reflect"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/faction"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
)

// FactionAdapter wraps Venture's faction generator for Wyrm's politics system.
type FactionAdapter struct {
	generator *faction.Generator
}

// NewFactionAdapter creates a new faction adapter.
func NewFactionAdapter() *FactionAdapter {
	return &FactionAdapter{
		generator: faction.NewGenerator(),
	}
}

// FactionData holds generated faction data for integration with Wyrm.
type FactionData struct {
	ID             string
	Name           string
	Type           string
	Description    string
	MemberCount    int
	TerritoryColor [4]uint8
	Relationships  map[string]int
}

// GenerateFactions creates factions for the world and returns Wyrm-compatible data.
func (a *FactionAdapter) GenerateFactions(seed int64, genre string, depth int) ([]*FactionData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Depth:      depth,
		Difficulty: DefaultGenerationDifficulty,
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("faction generation failed: %w", err)
	}

	// Use reflection to extract faction data without importing engine package
	// (which pulls in Ebiten and requires a display)
	data, err := extractFactionData(result)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// extractFactionData uses reflection to extract faction data from Venture's
// engine.Faction slice without directly importing the engine package.
func extractFactionData(result interface{}) ([]*FactionData, error) {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("invalid faction result type: expected slice, got %T", result)
	}

	data := make([]*FactionData, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		elem, err := getStructElement(rv.Index(i), i)
		if err != nil {
			return nil, err
		}
		data[i] = extractSingleFaction(elem)
	}
	return data, nil
}

// getStructElement dereferences pointers and validates struct type.
func getStructElement(elem reflect.Value, index int) (reflect.Value, error) {
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("invalid faction element type at index %d", index)
	}
	return elem, nil
}

// extractSingleFaction extracts all fields from a single faction struct.
func extractSingleFaction(elem reflect.Value) *FactionData {
	fd := &FactionData{}
	extractStringFields(elem, fd)
	extractTypeField(elem, fd)
	extractMemberCount(elem, fd)
	extractTerritoryColor(elem, fd)
	extractRelationships(elem, fd)
	return fd
}

// extractStringFields extracts ID, Name, and Description string fields.
func extractStringFields(elem reflect.Value, fd *FactionData) {
	if f := elem.FieldByName("ID"); f.IsValid() && f.Kind() == reflect.String {
		fd.ID = f.String()
	}
	if f := elem.FieldByName("Name"); f.IsValid() && f.Kind() == reflect.String {
		fd.Name = f.String()
	}
	if f := elem.FieldByName("Description"); f.IsValid() && f.Kind() == reflect.String {
		fd.Description = f.String()
	}
}

// extractTypeField extracts the Type field, converting to string.
func extractTypeField(elem reflect.Value, fd *FactionData) {
	if f := elem.FieldByName("Type"); f.IsValid() {
		fd.Type = fmt.Sprintf("%v", f.Interface())
	}
}

// extractMemberCount extracts the MemberCount integer field.
func extractMemberCount(elem reflect.Value, fd *FactionData) {
	if f := elem.FieldByName("MemberCount"); f.IsValid() && f.Kind() == reflect.Int {
		fd.MemberCount = int(f.Int())
	}
}

// extractTerritoryColor extracts the TerritoryColor [4]uint8 array.
func extractTerritoryColor(elem reflect.Value, fd *FactionData) {
	if f := elem.FieldByName("TerritoryColor"); f.IsValid() && f.Kind() == reflect.Array {
		for j := 0; j < 4 && j < f.Len(); j++ {
			fd.TerritoryColor[j] = uint8(f.Index(j).Uint())
		}
	}
}

// extractRelationships extracts the Relationships map[string]int field.
func extractRelationships(elem reflect.Value, fd *FactionData) {
	f := elem.FieldByName("Relationships")
	if !f.IsValid() || f.Kind() != reflect.Map {
		return
	}
	fd.Relationships = make(map[string]int)
	for _, key := range f.MapKeys() {
		if key.Kind() != reflect.String {
			continue
		}
		val := f.MapIndex(key)
		if val.Kind() == reflect.Int {
			fd.Relationships[key.String()] = int(val.Int())
		}
	}
}

// RegisterFactionsWithPoliticsSystem adds generated factions to the politics system.
func RegisterFactionsWithPoliticsSystem(fps *systems.FactionPoliticsSystem, factions []*FactionData) {
	for _, f1 := range factions {
		for f2ID, relationship := range f1.Relationships {
			rel := relationshipToFactionRelation(relationship)
			fps.SetRelation(f1.ID, f2ID, rel)
		}
	}
}

// relationshipToFactionRelation converts Venture's relationship int to Wyrm's FactionRelation.
func relationshipToFactionRelation(relationship int) systems.FactionRelation {
	if relationship <= FactionHostileThreshold {
		return systems.RelationHostile
	}
	if relationship >= FactionAllyThreshold {
		return systems.RelationAlly
	}
	return systems.RelationNeutral
}
