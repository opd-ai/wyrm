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

	return &City{
		Name:             cityName,
		Districts:        cityDistricts,
		Seed:             seed,
		Genre:            genre,
		ResidentialAreas: residentialAreas,
		IndustrialZones:  industrialZones,
	}
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
