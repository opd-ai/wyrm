// Package city provides procedural city generation.
package city

import (
	"fmt"
	"math"
	"math/rand"
)

// cos wraps math.Cos for convenience.
func cos(x float64) float64 { return math.Cos(x) }

// sin wraps math.Sin for convenience.
func sin(x float64) float64 { return math.Sin(x) }

// City represents a procedurally generated city.
type City struct {
	Name             string
	Districts        []District
	Seed             int64
	Genre            string
	ResidentialAreas []ResidentialArea
	IndustrialZones  []IndustrialZone
	Walls            Wall
	Gates            []CityGate
	Roads            []Road // Roads connecting districts and POIs
}

// District represents a city district.
type District struct {
	Name      string
	CenterX   float64
	CenterY   float64
	Buildings int
	Type      string
}

// ResidentialArea represents a neighborhood of housing.
type ResidentialArea struct {
	Name             string
	CenterX          float64
	CenterY          float64
	Radius           float64
	HousingType      string  // "apartments", "houses", "slums", "mansions", "dorms"
	Density          float64 // 0.0-1.0
	Population       int
	WealthLevel      int // 1 (poor) to 5 (wealthy)
	Buildings        []ResidentialBuilding
	Amenities        []string // "park", "school", "clinic", "market", "tavern"
	CrimeLevel       float64  // 0.0-1.0
	MaintenanceLevel float64  // 0.0-1.0 (building condition)
}

// ResidentialBuilding represents a single residential structure.
type ResidentialBuilding struct {
	X, Y         float64
	Width        int
	Height       int
	Floors       int
	Units        int // Number of dwelling units
	Occupied     int // Number of occupied units
	BuildingType string
	Style        string // Genre-specific architectural style
}

// IndustrialZone represents a manufacturing/production area.
type IndustrialZone struct {
	Name       string
	CenterX    float64
	CenterY    float64
	Radius     float64
	ZoneType   string // "factory", "warehouse", "refinery", "workshop", "mining"
	OutputType string // What the zone produces
	Workers    int
	Pollution  float64 // 0.0-1.0
	Automation float64 // 0.0-1.0 (higher = fewer workers needed)
	Facilities []IndustrialFacility
}

// IndustrialFacility represents a single industrial structure.
type IndustrialFacility struct {
	X, Y           float64
	Width          int
	Height         int
	FacilityType   string
	OutputRate     float64
	WorkerCapacity int
	Active         bool
}

// genreNames holds city name prefixes and suffixes by genre.
var genreNames = map[string]struct {
	prefixes []string
	suffixes []string
}{
	"fantasy": {
		prefixes: []string{"Elder", "Golden", "Silver", "Iron", "Dragon", "Storm", "Shadow", "Dawn"},
		suffixes: []string{"hold", "haven", "reach", "fall", "gate", "forge", "keep", "spire"},
	},
	"sci-fi": {
		prefixes: []string{"Neo", "Nova", "Alpha", "Omega", "Stellar", "Quantum", "Helix", "Nexus"},
		suffixes: []string{"prime", "station", "port", "dome", "arc", "core", "hub", "node"},
	},
	"horror": {
		prefixes: []string{"Hollow", "Bleak", "Silent", "Ashen", "Cursed", "Withered", "Pale", "Dread"},
		suffixes: []string{"moor", "vale", "marsh", "hollow", "crypt", "shade", "glen", "barrow"},
	},
	"cyberpunk": {
		prefixes: []string{"Neon", "Chrome", "Data", "Grid", "Cyber", "Synth", "Wire", "Volt"},
		suffixes: []string{"sprawl", "sector", "block", "zone", "grid", "strip", "stack", "layer"},
	},
	"post-apocalyptic": {
		prefixes: []string{"Rust", "Scrap", "Bone", "Dust", "Rad", "Ash", "Iron", "Salt"},
		suffixes: []string{"town", "camp", "fort", "haven", "pit", "wastes", "hold", "ruins"},
	},
}

// genreDistricts holds district types by genre.
var genreDistricts = map[string][]string{
	"fantasy":          {"Market", "Temple", "Castle", "Guild", "Harbor", "Slums", "Noble", "Mage"},
	"sci-fi":           {"Residential", "Industrial", "Commercial", "Military", "Research", "Transit", "Medical", "Admin"},
	"horror":           {"Old Town", "Cemetery", "Manor", "Asylum", "Church", "Docks", "Ruins", "Catacombs"},
	"cyberpunk":        {"Corporate", "Residential", "Industrial", "Entertainment", "Black Market", "Slums", "Tech", "Enforcement"},
	"post-apocalyptic": {"Salvage", "Farm", "Trade", "Shelter", "Water", "Armory", "Med Bay", "Wasteland"},
}

// Generate creates a new procedural city from the given seed and genre.
func Generate(seed int64, genre string) *City {
	rng := rand.New(rand.NewSource(seed))

	// Default to fantasy if unknown genre
	names, ok := genreNames[genre]
	if !ok {
		names = genreNames["fantasy"]
	}
	districts, ok := genreDistricts[genre]
	if !ok {
		districts = genreDistricts["fantasy"]
	}

	// Generate city name
	prefix := names.prefixes[rng.Intn(len(names.prefixes))]
	suffix := names.suffixes[rng.Intn(len(names.suffixes))]
	cityName := prefix + suffix

	// Generate 3-6 districts
	numDistricts := 3 + rng.Intn(4)
	cityDistricts := make([]District, numDistricts)

	for i := 0; i < numDistricts; i++ {
		districtType := districts[rng.Intn(len(districts))]
		cityDistricts[i] = District{
			Name:      fmt.Sprintf("%s %s", cityName, districtType),
			CenterX:   rng.Float64()*1000 - 500,
			CenterY:   rng.Float64()*1000 - 500,
			Buildings: 10 + rng.Intn(41),
			Type:      districtType,
		}
	}

	// Generate residential areas
	residentialAreas := generateResidentialAreas(rng, cityName, genre, cityDistricts)

	// Generate industrial zones
	industrialZones := generateIndustrialZones(rng, cityName, genre, cityDistricts)

	city := &City{
		Name:             cityName,
		Districts:        cityDistricts,
		Seed:             seed,
		Genre:            genre,
		ResidentialAreas: residentialAreas,
		IndustrialZones:  industrialZones,
	}

	// Generate defensive walls and gates
	city.GenerateWallsAndGates(rng)

	// Generate roads connecting districts
	city.GenerateRoads(rng)

	return city
}

// genreHousingTypes maps genres to appropriate housing types.
var genreHousingTypes = map[string][]string{
	"fantasy":          {"cottages", "townhouses", "manor", "slums", "peasant_huts"},
	"sci-fi":           {"apartments", "hab_pods", "luxury_suites", "barracks", "cryo_bunks"},
	"horror":           {"victorian_homes", "tenements", "manor", "asylum_cells", "crypts"},
	"cyberpunk":        {"megablocks", "coffin_hotels", "luxury_penthouses", "slum_stacks", "corp_housing"},
	"post-apocalyptic": {"bunkers", "makeshift_shelters", "fortified_homes", "camps", "ruins"},
}

// genreResidentialNames maps genres to neighborhood naming styles.
var genreResidentialNames = map[string][]string{
	"fantasy":          {"Quarter", "Row", "Lane", "Circle", "Heights"},
	"sci-fi":           {"Sector", "Block", "Ring", "Level", "Zone"},
	"horror":           {"Row", "Heights", "Hollow", "Court", "End"},
	"cyberpunk":        {"Block", "Stack", "Level", "Grid", "Sector"},
	"post-apocalyptic": {"Camp", "Settlement", "Refuge", "Zone", "Compound"},
}

// generateResidentialAreas creates residential neighborhoods.
func generateResidentialAreas(rng *rand.Rand, cityName, genre string, districts []District) []ResidentialArea {
	housingTypes := genreHousingTypes[genre]
	if housingTypes == nil {
		housingTypes = genreHousingTypes["fantasy"]
	}
	nameStyles := genreResidentialNames[genre]
	if nameStyles == nil {
		nameStyles = genreResidentialNames["fantasy"]
	}

	// Generate 2-5 residential areas
	numAreas := 2 + rng.Intn(4)
	areas := make([]ResidentialArea, numAreas)

	for i := 0; i < numAreas; i++ {
		// Base position off of existing districts
		var baseX, baseY float64
		if len(districts) > 0 {
			base := districts[rng.Intn(len(districts))]
			baseX = base.CenterX + (rng.Float64()*200 - 100)
			baseY = base.CenterY + (rng.Float64()*200 - 100)
		} else {
			baseX = rng.Float64()*800 - 400
			baseY = rng.Float64()*800 - 400
		}

		wealthLevel := 1 + rng.Intn(5)
		density := 0.3 + rng.Float64()*0.6
		radius := 50 + rng.Float64()*100

		housingType := housingTypes[rng.Intn(len(housingTypes))]
		nameStyle := nameStyles[rng.Intn(len(nameStyles))]

		area := ResidentialArea{
			Name:             fmt.Sprintf("%s %s", cityName, nameStyle),
			CenterX:          baseX,
			CenterY:          baseY,
			Radius:           radius,
			HousingType:      housingType,
			Density:          density,
			WealthLevel:      wealthLevel,
			CrimeLevel:       (5.0 - float64(wealthLevel)) / 5.0 * 0.5, // Lower wealth = higher crime
			MaintenanceLevel: float64(wealthLevel) / 5.0,
			Amenities:        generateAmenities(rng, wealthLevel, genre),
		}

		// Generate buildings
		area.Buildings = generateResidentialBuildings(rng, &area, genre)
		area.Population = calculatePopulation(&area)

		areas[i] = area
	}

	return areas
}

// generateAmenities creates appropriate amenities for a residential area.
func generateAmenities(rng *rand.Rand, wealthLevel int, genre string) []string {
	amenities := make([]string, 0)

	// Base amenities all areas might have
	baseAmenities := map[string][]string{
		"fantasy":          {"well", "tavern", "shrine", "market"},
		"sci-fi":           {"vendor_kiosk", "med_station", "transit_stop", "park_dome"},
		"horror":           {"church", "pharmacy", "pub", "cemetery"},
		"cyberpunk":        {"vending_wall", "clinic", "bar", "arcade"},
		"post-apocalyptic": {"water_pump", "trade_post", "clinic", "watchtower"},
	}

	genreAmenities := baseAmenities[genre]
	if genreAmenities == nil {
		genreAmenities = baseAmenities["fantasy"]
	}

	// Higher wealth = more amenities
	numAmenities := 1 + wealthLevel/2 + rng.Intn(2)
	for i := 0; i < numAmenities && i < len(genreAmenities); i++ {
		idx := rng.Intn(len(genreAmenities))
		amenity := genreAmenities[idx]
		// Avoid duplicates
		found := false
		for _, a := range amenities {
			if a == amenity {
				found = true
				break
			}
		}
		if !found {
			amenities = append(amenities, amenity)
		}
	}

	return amenities
}

// generateResidentialBuildings creates buildings for a residential area.
func generateResidentialBuildings(rng *rand.Rand, area *ResidentialArea, genre string) []ResidentialBuilding {
	// Number of buildings based on density and radius
	numBuildings := int(area.Density * area.Radius / 10)
	if numBuildings < 3 {
		numBuildings = 3
	}
	if numBuildings > 50 {
		numBuildings = 50
	}

	buildings := make([]ResidentialBuilding, numBuildings)

	for i := 0; i < numBuildings; i++ {
		// Position within radius using polar coordinates
		angle := rng.Float64() * 6.28318 // 2*PI
		dist := rng.Float64() * area.Radius
		x := area.CenterX + dist*cos(angle)
		y := area.CenterY + dist*sin(angle)

		// Size based on wealth and housing type
		var width, height, floors, units int
		switch area.HousingType {
		case "apartments", "megablocks", "hab_pods":
			width = 8 + rng.Intn(8)
			height = 8 + rng.Intn(8)
			floors = 3 + rng.Intn(10+area.WealthLevel*2)
			units = floors * (2 + rng.Intn(4))
		case "houses", "cottages", "victorian_homes", "fortified_homes":
			width = 6 + rng.Intn(4)
			height = 6 + rng.Intn(4)
			floors = 1 + rng.Intn(2)
			units = 1
		case "manor", "luxury_penthouses", "luxury_suites":
			width = 12 + rng.Intn(8)
			height = 12 + rng.Intn(8)
			floors = 2 + rng.Intn(3)
			units = 1 + rng.Intn(3)
		default:
			width = 4 + rng.Intn(4)
			height = 4 + rng.Intn(4)
			floors = 1
			units = 1
		}

		buildings[i] = ResidentialBuilding{
			X:            x,
			Y:            y,
			Width:        width,
			Height:       height,
			Floors:       floors,
			Units:        units,
			Occupied:     int(float64(units) * (0.7 + rng.Float64()*0.25)),
			BuildingType: area.HousingType,
			Style:        getArchitecturalStyle(genre, area.WealthLevel, rng),
		}
	}

	return buildings
}

// calculatePopulation estimates population from buildings.
func calculatePopulation(area *ResidentialArea) int {
	pop := 0
	for _, b := range area.Buildings {
		// Average 2.5 people per occupied unit
		pop += int(float64(b.Occupied) * 2.5)
	}
	return pop
}

// getArchitecturalStyle returns a genre-appropriate building style.
func getArchitecturalStyle(genre string, wealthLevel int, rng *rand.Rand) string {
	styles := map[string]map[int][]string{
		"fantasy": {
			1: {"ramshackle", "crude_wood"},
			2: {"timber_frame", "thatch"},
			3: {"stone_cottage", "half_timber"},
			4: {"brick_townhouse", "slate_roof"},
			5: {"marble_manor", "gilded"},
		},
		"sci-fi": {
			1: {"prefab_basic", "container"},
			2: {"modular_standard", "polymer"},
			3: {"smart_composite", "glass_steel"},
			4: {"designer_alloy", "holographic"},
			5: {"luxury_orbital", "nano_assembled"},
		},
		"cyberpunk": {
			1: {"scrap_metal", "shipping_container"},
			2: {"brutalist_concrete", "neon_lit"},
			3: {"corporate_standard", "chrome_glass"},
			4: {"designer_neo", "holo_facade"},
			5: {"exec_tower", "skybridge"},
		},
		"post-apocalyptic": {
			1: {"rubble_shelter", "tarp_lean_to"},
			2: {"salvage_walls", "reinforced_ruins"},
			3: {"bunker_concrete", "metal_shell"},
			4: {"prewar_restored", "generator_powered"},
			5: {"vault_tech", "pristine_prewar"},
		},
	}

	genreStyles := styles[genre]
	if genreStyles == nil {
		return "standard"
	}
	wealthStyles := genreStyles[wealthLevel]
	if wealthStyles == nil {
		return "standard"
	}
	return wealthStyles[rng.Intn(len(wealthStyles))]
}

// genreIndustrialTypes maps genres to appropriate industrial zone types.
var genreIndustrialTypes = map[string][]string{
	"fantasy":          {"smithy", "tannery", "brewery", "mill", "quarry", "lumber_yard"},
	"sci-fi":           {"factory", "refinery", "shipyard", "power_plant", "mining_rig", "biolab"},
	"horror":           {"meat_packing", "chemical_plant", "morgue", "asylum_workshop", "tannery", "mill"},
	"cyberpunk":        {"factory", "data_center", "chem_plant", "chop_shop", "bio_vat", "recycler"},
	"post-apocalyptic": {"salvage_yard", "water_purifier", "workshop", "fuel_refinery", "farm", "scrap_forge"},
}

// generateIndustrialZones creates industrial/production areas.
func generateIndustrialZones(rng *rand.Rand, cityName, genre string, districts []District) []IndustrialZone {
	zoneTypes := genreIndustrialTypes[genre]
	if zoneTypes == nil {
		zoneTypes = genreIndustrialTypes["fantasy"]
	}

	// Generate 1-3 industrial zones
	numZones := 1 + rng.Intn(3)
	zones := make([]IndustrialZone, numZones)

	for i := 0; i < numZones; i++ {
		// Position away from city center
		baseX := rng.Float64()*600 - 300
		if rng.Float64() > 0.5 {
			baseX += 200
		} else {
			baseX -= 200
		}
		baseY := rng.Float64()*600 - 300
		if rng.Float64() > 0.5 {
			baseY += 200
		} else {
			baseY -= 200
		}

		zoneType := zoneTypes[rng.Intn(len(zoneTypes))]
		radius := 80 + rng.Float64()*120
		automation := rng.Float64() * 0.5 // 0-50% automation
		if genre == "sci-fi" || genre == "cyberpunk" {
			automation = 0.3 + rng.Float64()*0.6 // Higher in tech genres
		}

		zone := IndustrialZone{
			Name:       fmt.Sprintf("%s %s", cityName, zoneType),
			CenterX:    baseX,
			CenterY:    baseY,
			Radius:     radius,
			ZoneType:   zoneType,
			OutputType: getOutputType(zoneType),
			Pollution:  0.2 + rng.Float64()*0.6,
			Automation: automation,
		}

		// Generate facilities
		zone.Facilities = generateIndustrialFacilities(rng, &zone)
		zone.Workers = calculateWorkers(&zone)

		zones[i] = zone
	}

	return zones
}

// getOutputType returns what an industrial zone produces.
func getOutputType(zoneType string) string {
	outputs := map[string]string{
		"smithy":         "metal_goods",
		"tannery":        "leather",
		"brewery":        "alcohol",
		"mill":           "flour",
		"quarry":         "stone",
		"lumber_yard":    "lumber",
		"factory":        "manufactured_goods",
		"refinery":       "fuel",
		"shipyard":       "vehicles",
		"power_plant":    "power",
		"mining_rig":     "ore",
		"biolab":         "medicine",
		"meat_packing":   "food",
		"chemical_plant": "chemicals",
		"data_center":    "data",
		"chop_shop":      "parts",
		"bio_vat":        "organics",
		"recycler":       "recycled_materials",
		"salvage_yard":   "salvage",
		"water_purifier": "clean_water",
		"workshop":       "repairs",
		"fuel_refinery":  "fuel",
		"farm":           "food",
		"scrap_forge":    "metal_goods",
	}
	if output, ok := outputs[zoneType]; ok {
		return output
	}
	return "goods"
}

// generateIndustrialFacilities creates facilities for an industrial zone.
func generateIndustrialFacilities(rng *rand.Rand, zone *IndustrialZone) []IndustrialFacility {
	numFacilities := 2 + rng.Intn(5)
	facilities := make([]IndustrialFacility, numFacilities)

	for i := 0; i < numFacilities; i++ {
		// Position within zone radius
		x := zone.CenterX + (rng.Float64()*2-1)*zone.Radius*0.8
		y := zone.CenterY + (rng.Float64()*2-1)*zone.Radius*0.8

		width := 15 + rng.Intn(20)
		height := 15 + rng.Intn(20)
		workerCap := 10 + rng.Intn(50)

		facilities[i] = IndustrialFacility{
			X:              x,
			Y:              y,
			Width:          width,
			Height:         height,
			FacilityType:   zone.ZoneType,
			OutputRate:     0.5 + rng.Float64()*0.5,
			WorkerCapacity: workerCap,
			Active:         rng.Float64() > 0.1, // 90% chance active
		}
	}

	return facilities
}

// calculateWorkers estimates total workers in an industrial zone.
func calculateWorkers(zone *IndustrialZone) int {
	workers := 0
	for _, f := range zone.Facilities {
		if f.Active {
			// Automation reduces worker count
			needed := float64(f.WorkerCapacity) * (1.0 - zone.Automation*0.8)
			workers += int(needed)
		}
	}
	return workers
}

// GetResidentialPopulation returns total population across all residential areas.
func (c *City) GetResidentialPopulation() int {
	total := 0
	for _, area := range c.ResidentialAreas {
		total += area.Population
	}
	return total
}

// GetIndustrialWorkers returns total workers across all industrial zones.
func (c *City) GetIndustrialWorkers() int {
	total := 0
	for _, zone := range c.IndustrialZones {
		total += zone.Workers
	}
	return total
}

// GetAreaByName finds a residential area by name.
func (c *City) GetAreaByName(name string) *ResidentialArea {
	for i := range c.ResidentialAreas {
		if c.ResidentialAreas[i].Name == name {
			return &c.ResidentialAreas[i]
		}
	}
	return nil
}

// GetZoneByName finds an industrial zone by name.
func (c *City) GetZoneByName(name string) *IndustrialZone {
	for i := range c.IndustrialZones {
		if c.IndustrialZones[i].Name == name {
			return &c.IndustrialZones[i]
		}
	}
	return nil
}

// Wall represents the defensive walls surrounding a city.
type Wall struct {
	// Segments defines wall segments as start/end point pairs.
	Segments []WallSegment
	// Height is the wall height in units.
	Height float64
	// Thickness is the wall thickness in units.
	Thickness float64
	// Material is the construction material (affects durability).
	Material string
	// Condition is the current repair state (0.0-1.0).
	Condition float64
}

// WallSegment represents a single section of wall.
type WallSegment struct {
	StartX, StartY float64
	EndX, EndY     float64
	HasWalkway     bool    // Can guards patrol on top
	Damaged        bool    // Section has breach
	DamageLevel    float64 // 0.0-1.0 damage amount
}

// CityGate represents an entrance through the city walls.
type CityGate struct {
	Name       string
	X, Y       float64
	Width      float64 // Gate opening width
	Facing     string  // "north", "south", "east", "west"
	IsOpen     bool
	OpenHour   int  // Hour gate opens (0-23)
	CloseHour  int  // Hour gate closes (0-23)
	GuardCount int  // Number of guards stationed
	Locked     bool // Requires key or permission
	Style      string
}

// Road represents a path connecting two points in the city.
type Road struct {
	StartX, StartY float64
	EndX, EndY     float64
	Width          float64 // Road width in units
	Type           string  // "main", "side", "path", "highway"
	Material       string  // Surface material (cobblestone, asphalt, dirt, etc.)
}

// RoadPoint represents a point along a road path.
type RoadPoint struct {
	X, Y float64
}

// genreRoadMaterials maps genres to appropriate road surface materials.
var genreRoadMaterials = map[string][]string{
	"fantasy":          {"cobblestone", "brick", "dirt_path", "flagstone"},
	"sci-fi":           {"plasteel_grating", "hover_lane", "mag_rail", "energy_path"},
	"horror":           {"cracked_asphalt", "bone_path", "mud", "rotting_wood"},
	"cyberpunk":        {"neon_asphalt", "holo_guide", "smart_pavement", "chrome_strip"},
	"post-apocalyptic": {"cracked_highway", "rubble_path", "dirt", "scrap_bridge"},
}

// genreWallMaterials maps genres to appropriate wall construction materials.
var genreWallMaterials = map[string][]string{
	"fantasy":          {"stone", "granite", "marble", "enchanted_stone"},
	"sci-fi":           {"plasteel", "forcefield", "nano_composite", "energy_barrier"},
	"horror":           {"crumbling_stone", "iron", "bone", "cursed_masonry"},
	"cyberpunk":        {"concrete", "steel", "smart_glass", "holographic_barrier"},
	"post-apocalyptic": {"scrap_metal", "concrete_rubble", "car_bodies", "shipping_containers"},
}

// genreGateStyles maps genres to appropriate gate architectural styles.
var genreGateStyles = map[string][]string{
	"fantasy":          {"portcullis", "drawbridge", "wooden_doors", "enchanted_arch"},
	"sci-fi":           {"blast_door", "iris_gate", "energy_field", "airlock"},
	"horror":           {"rusted_iron", "bone_arch", "creaking_wood", "mist_barrier"},
	"cyberpunk":        {"security_checkpoint", "scanner_gate", "holo_barrier", "blast_shutter"},
	"post-apocalyptic": {"barricade", "car_gate", "chain_fence", "guard_tower"},
}

// GenerateWallsAndGates generates defensive walls and entrance gates for a city.
func (c *City) GenerateWallsAndGates(rng *rand.Rand) {
	materials := getGenreMaterials(c.Genre, genreWallMaterials)
	gateStyles := getGenreMaterials(c.Genre, genreGateStyles)

	bounds := c.calculateWallBounds()
	c.Walls = c.createWall(rng, materials, bounds)
	c.applyWallDamage(rng)
	c.Gates = c.generateGates(rng, gateStyles, bounds)
}

// getGenreMaterials returns materials for a genre with fantasy fallback.
func getGenreMaterials(genre string, genreMap map[string][]string) []string {
	materials := genreMap[genre]
	if materials == nil {
		materials = genreMap["fantasy"]
	}
	return materials
}

// wallBounds holds the calculated wall boundaries.
type wallBounds struct {
	minX, minY, maxX, maxY float64
}

// calculateWallBounds computes city bounds with padding for walls.
func (c *City) calculateWallBounds() wallBounds {
	minX, minY, maxX, maxY := c.calculateBounds()
	wallPadding := 50.0
	return wallBounds{
		minX: minX - wallPadding,
		minY: minY - wallPadding,
		maxX: maxX + wallPadding,
		maxY: maxY + wallPadding,
	}
}

// createWall creates the wall structure with segments.
func (c *City) createWall(rng *rand.Rand, materials []string, b wallBounds) Wall {
	material := materials[rng.Intn(len(materials))]
	return Wall{
		Height:    8.0 + rng.Float64()*8.0,
		Thickness: 2.0 + rng.Float64()*3.0,
		Material:  material,
		Condition: 0.7 + rng.Float64()*0.3,
		Segments: []WallSegment{
			{StartX: b.minX, StartY: b.minY, EndX: b.maxX, EndY: b.minY, HasWalkway: true}, // North
			{StartX: b.maxX, StartY: b.minY, EndX: b.maxX, EndY: b.maxY, HasWalkway: true}, // East
			{StartX: b.maxX, StartY: b.maxY, EndX: b.minX, EndY: b.maxY, HasWalkway: true}, // South
			{StartX: b.minX, StartY: b.maxY, EndX: b.minX, EndY: b.minY, HasWalkway: true}, // West
		},
	}
}

// applyWallDamage adds random damage to wall segments based on genre.
func (c *City) applyWallDamage(rng *rand.Rand) {
	damageChance := 0.1
	if c.Genre == "horror" || c.Genre == "post-apocalyptic" {
		damageChance = 0.4
	}
	for i := range c.Walls.Segments {
		if rng.Float64() < damageChance {
			c.Walls.Segments[i].Damaged = true
			c.Walls.Segments[i].DamageLevel = 0.2 + rng.Float64()*0.5
		}
	}
}

// generateGates creates the city gates.
func (c *City) generateGates(rng *rand.Rand, gateStyles []string, b wallBounds) []CityGate {
	gateDirections := []struct {
		name   string
		x, y   float64
		facing string
	}{
		{"North Gate", (b.minX + b.maxX) / 2, b.minY, "north"},
		{"South Gate", (b.minX + b.maxX) / 2, b.maxY, "south"},
		{"East Gate", b.maxX, (b.minY + b.maxY) / 2, "east"},
		{"West Gate", b.minX, (b.minY + b.maxY) / 2, "west"},
	}

	// Shuffle directions
	for i := range gateDirections {
		j := rng.Intn(i + 1)
		gateDirections[i], gateDirections[j] = gateDirections[j], gateDirections[i]
	}

	numGates := 2 + rng.Intn(3)
	gates := make([]CityGate, numGates)
	for i := 0; i < numGates && i < len(gateDirections); i++ {
		gates[i] = c.createGate(rng, gateStyles, gateDirections[i])
	}
	return gates
}

// createGate creates a single city gate.
func (c *City) createGate(rng *rand.Rand, gateStyles []string, dir struct {
	name   string
	x, y   float64
	facing string
},
) CityGate {
	style := gateStyles[rng.Intn(len(gateStyles))]
	gate := CityGate{
		Name:       fmt.Sprintf("%s %s", c.Name, dir.name),
		X:          dir.x,
		Y:          dir.y,
		Width:      8.0 + rng.Float64()*4.0,
		Facing:     dir.facing,
		IsOpen:     true,
		OpenHour:   5 + rng.Intn(2),
		CloseHour:  21 + rng.Intn(3),
		GuardCount: 2 + rng.Intn(5),
		Locked:     false,
		Style:      style,
	}
	if (c.Genre == "horror" || c.Genre == "post-apocalyptic") && rng.Float64() < 0.3 {
		gate.Locked = true
	}
	return gate
}

// calculateBounds finds the bounding box of all districts.
func (c *City) calculateBounds() (minX, minY, maxX, maxY float64) {
	if len(c.Districts) == 0 {
		return -100, -100, 100, 100
	}

	minX = c.Districts[0].CenterX
	maxX = c.Districts[0].CenterX
	minY = c.Districts[0].CenterY
	maxY = c.Districts[0].CenterY

	for _, d := range c.Districts {
		// Estimate district radius based on building count
		radius := float64(d.Buildings) * 3.0
		if d.CenterX-radius < minX {
			minX = d.CenterX - radius
		}
		if d.CenterX+radius > maxX {
			maxX = d.CenterX + radius
		}
		if d.CenterY-radius < minY {
			minY = d.CenterY - radius
		}
		if d.CenterY+radius > maxY {
			maxY = d.CenterY + radius
		}
	}

	return minX, minY, maxX, maxY
}

// GetGateByDirection finds a gate facing the specified direction.
func (c *City) GetGateByDirection(direction string) *CityGate {
	for i := range c.Gates {
		if c.Gates[i].Facing == direction {
			return &c.Gates[i]
		}
	}
	return nil
}

// GetGateByName finds a gate by its name.
func (c *City) GetGateByName(name string) *CityGate {
	for i := range c.Gates {
		if c.Gates[i].Name == name {
			return &c.Gates[i]
		}
	}
	return nil
}

// IsGateOpen checks if a gate is open at the given hour.
func (g *CityGate) IsGateOpen(hour int) bool {
	if g.Locked {
		return false
	}
	if g.OpenHour <= g.CloseHour {
		return hour >= g.OpenHour && hour < g.CloseHour
	}
	// Handle overnight hours
	return hour >= g.OpenHour || hour < g.CloseHour
}

// GetWallCondition returns the overall wall condition (average of segments).
func (c *Wall) GetWallCondition() float64 {
	if len(c.Segments) == 0 {
		return c.Condition
	}

	totalDamage := 0.0
	for _, seg := range c.Segments {
		if seg.Damaged {
			totalDamage += seg.DamageLevel
		}
	}
	return c.Condition * (1.0 - totalDamage/float64(len(c.Segments)))
}

// HasBreach returns true if any wall segment is significantly damaged.
func (c *Wall) HasBreach() bool {
	for _, seg := range c.Segments {
		if seg.Damaged && seg.DamageLevel > 0.5 {
			return true
		}
	}
	return false
}

// GenerateRoads creates roads connecting districts and key locations.
// Uses a minimum spanning tree (MST) approach to connect all districts,
// then adds some extra connections for redundancy.
func (c *City) GenerateRoads(rng *rand.Rand) {
	if len(c.Districts) < 2 {
		return
	}

	materials := c.getRoadMaterials()
	roads := make([]Road, 0, len(c.Districts)*2)

	// Build MST connecting all districts
	roads = append(roads, c.generateMSTRoads(rng, materials)...)

	// Add shortcut roads for redundancy
	roads = append(roads, c.generateShortcutRoads(rng, materials)...)

	// Connect gates to nearest districts
	roads = append(roads, c.generateGateRoads(materials)...)

	c.Roads = roads
}

// getRoadMaterials returns the appropriate road materials for the city's genre.
func (c *City) getRoadMaterials() []string {
	materials := genreRoadMaterials[c.Genre]
	if materials == nil {
		return genreRoadMaterials["fantasy"]
	}
	return materials
}

// generateMSTRoads creates the minimum spanning tree of roads using Prim's algorithm.
func (c *City) generateMSTRoads(rng *rand.Rand, materials []string) []Road {
	connected := make([]bool, len(c.Districts))
	connected[0] = true
	numConnected := 1
	roads := make([]Road, 0, len(c.Districts)-1)

	for numConnected < len(c.Districts) {
		from, to := c.findShortestEdge(connected)
		if to < 0 {
			break
		}

		roads = append(roads, c.createRoad(
			c.Districts[from].CenterX, c.Districts[from].CenterY,
			c.Districts[to].CenterX, c.Districts[to].CenterY,
			3.0+rng.Float64()*2.0, "main",
			materials[rng.Intn(len(materials))],
		))

		connected[to] = true
		numConnected++
	}
	return roads
}

// findShortestEdge finds the shortest edge from a connected to an unconnected district.
func (c *City) findShortestEdge(connected []bool) (bestFrom, bestTo int) {
	bestDist := math.MaxFloat64
	bestFrom, bestTo = -1, -1

	for i := range c.Districts {
		if !connected[i] {
			continue
		}
		for j := range c.Districts {
			if connected[j] {
				continue
			}
			dist := c.distSquared(i, j)
			if dist < bestDist {
				bestDist = dist
				bestFrom, bestTo = i, j
			}
		}
	}
	return bestFrom, bestTo
}

// distSquared returns the squared distance between two districts.
func (c *City) distSquared(i, j int) float64 {
	dx := c.Districts[i].CenterX - c.Districts[j].CenterX
	dy := c.Districts[i].CenterY - c.Districts[j].CenterY
	return dx*dx + dy*dy
}

// generateShortcutRoads adds extra connections for redundancy (30% chance per pair).
func (c *City) generateShortcutRoads(rng *rand.Rand, materials []string) []Road {
	var roads []Road
	const shortcutChance = 0.3
	const maxShortcutDist = 400.0

	for i := 0; i < len(c.Districts); i++ {
		for j := i + 2; j < len(c.Districts); j++ {
			if rng.Float64() >= shortcutChance {
				continue
			}
			dist := math.Sqrt(c.distSquared(i, j))
			if dist >= maxShortcutDist {
				continue
			}
			roads = append(roads, c.createRoad(
				c.Districts[i].CenterX, c.Districts[i].CenterY,
				c.Districts[j].CenterX, c.Districts[j].CenterY,
				2.0+rng.Float64()*1.0, "side",
				materials[rng.Intn(len(materials))],
			))
		}
	}
	return roads
}

// generateGateRoads connects city gates to their nearest districts.
func (c *City) generateGateRoads(materials []string) []Road {
	roads := make([]Road, 0, len(c.Gates))
	for _, gate := range c.Gates {
		nearestIdx := c.findNearestDistrict(gate.X, gate.Y)
		roads = append(roads, c.createRoad(
			gate.X, gate.Y,
			c.Districts[nearestIdx].CenterX, c.Districts[nearestIdx].CenterY,
			4.0, "main", materials[0],
		))
	}
	return roads
}

// findNearestDistrict returns the index of the district closest to the given point.
func (c *City) findNearestDistrict(x, y float64) int {
	nearestIdx := 0
	nearestDist := math.MaxFloat64
	for i, d := range c.Districts {
		dx := x - d.CenterX
		dy := y - d.CenterY
		dist := dx*dx + dy*dy
		if dist < nearestDist {
			nearestDist = dist
			nearestIdx = i
		}
	}
	return nearestIdx
}

// createRoad creates a new Road with the given parameters.
func (c *City) createRoad(startX, startY, endX, endY, width float64, roadType, material string) Road {
	return Road{
		StartX:   startX,
		StartY:   startY,
		EndX:     endX,
		EndY:     endY,
		Width:    width,
		Type:     roadType,
		Material: material,
	}
}

// GetRoads returns all roads in the city.
func (c *City) GetRoads() []Road {
	return c.Roads
}

// IsOnRoad checks if a world coordinate is on a road.
// Returns true if within roadWidth/2 of any road segment.
func (c *City) IsOnRoad(worldX, worldY float64) bool {
	for _, road := range c.Roads {
		if isPointNearLineSegment(worldX, worldY, road.StartX, road.StartY, road.EndX, road.EndY, road.Width/2) {
			return true
		}
	}
	return false
}

// isPointNearLineSegment checks if point (px, py) is within distance d of line segment (x1,y1)-(x2,y2).
func isPointNearLineSegment(px, py, x1, y1, x2, y2, d float64) bool {
	// Vector from p1 to p2
	dx := x2 - x1
	dy := y2 - y1
	lenSq := dx*dx + dy*dy

	if lenSq == 0 {
		// Degenerate segment (point)
		return (px-x1)*(px-x1)+(py-y1)*(py-y1) <= d*d
	}

	// Project point onto line, clamped to segment
	t := ((px-x1)*dx + (py-y1)*dy) / lenSq
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Closest point on segment
	closestX := x1 + t*dx
	closestY := y1 + t*dy

	// Distance from point to closest point on segment
	distSq := (px-closestX)*(px-closestX) + (py-closestY)*(py-closestY)
	return distSq <= d*d
}
