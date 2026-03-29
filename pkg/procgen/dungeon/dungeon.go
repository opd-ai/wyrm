// Package dungeon provides procedural dungeon generation using BSP room graphs.
// Per ROADMAP Phase 6 item 25:
// AC: 100 generated dungeons have 0 unreachable rooms; each genre produces distinct tile aesthetics.
package dungeon

import (
	"math/rand"
)

// TileType represents a dungeon tile.
type TileType int

const (
	TileWall TileType = iota
	TileFloor
	TileDoor
	TileTrap
	TilePuzzle
	TileBossArena
	TileEntrance
	TileExit
)

// Room represents a dungeon room.
type Room struct {
	ID         int
	X, Y       int // Top-left position in dungeon grid
	Width      int
	Height     int
	Type       RoomType
	Connected  []int // Connected room IDs
	HasTrap    bool
	HasPuzzle  bool
	IsBossRoom bool
	Genre      string
}

// RoomType categorizes room purposes.
type RoomType int

const (
	RoomNormal RoomType = iota
	RoomEntrance
	RoomExit
	RoomTreasure
	RoomTrap
	RoomPuzzle
	RoomBoss
)

// Dungeon represents a generated dungeon layout.
type Dungeon struct {
	Width   int
	Height  int
	Tiles   [][]TileType
	Rooms   []*Room
	Genre   string
	Seed    int64
	Depth   int // Dungeon depth/difficulty level
	Palette TilePalette
}

// TilePalette defines genre-specific visual styles.
type TilePalette struct {
	WallColor    uint32
	FloorColor   uint32
	DoorColor    uint32
	TrapColor    uint32
	AccentColor  uint32
	WallTexture  string
	FloorTexture string
}

// GenrePalettes maps genre to visual style.
var GenrePalettes = map[string]TilePalette{
	"fantasy": {
		WallColor:    0x4A3B2A, // Dark brown stone
		FloorColor:   0x6B5B4A, // Lighter stone
		DoorColor:    0x8B4513, // Saddle brown
		TrapColor:    0xFF4500, // Orange-red
		AccentColor:  0xFFD700, // Gold
		WallTexture:  "cobblestone",
		FloorTexture: "flagstone",
	},
	"sci-fi": {
		WallColor:    0x1A1A2E, // Dark blue-grey
		FloorColor:   0x2A2A4E, // Lighter blue-grey
		DoorColor:    0x00FFFF, // Cyan
		TrapColor:    0xFF00FF, // Magenta
		AccentColor:  0x00FF00, // Green
		WallTexture:  "metal_panel",
		FloorTexture: "grate",
	},
	"horror": {
		WallColor:    0x2E2E2E, // Dark grey
		FloorColor:   0x3E3E3E, // Slightly lighter grey
		DoorColor:    0x4E4E4E, // Grey
		TrapColor:    0x8B0000, // Dark red
		AccentColor:  0x556B2F, // Dark olive
		WallTexture:  "cracked_stone",
		FloorTexture: "dirty_tile",
	},
	"cyberpunk": {
		WallColor:    0x0D0D0D, // Near black
		FloorColor:   0x1A1A1A, // Dark grey
		DoorColor:    0xFF1493, // Deep pink
		TrapColor:    0x00CED1, // Dark turquoise
		AccentColor:  0xFF69B4, // Hot pink
		WallTexture:  "neon_panel",
		FloorTexture: "tech_floor",
	},
	"post-apocalyptic": {
		WallColor:    0x5C4033, // Dark rusty brown
		FloorColor:   0x8B7355, // Dusty tan
		DoorColor:    0x8B4513, // Saddle brown
		TrapColor:    0xFFFF00, // Yellow (radiation)
		AccentColor:  0xCD853F, // Peru
		WallTexture:  "rusted_metal",
		FloorTexture: "rubble",
	},
}

// Generator creates procedural dungeons.
type Generator struct {
	Seed  int64
	Genre string
	rng   *rand.Rand
}

// NewGenerator creates a dungeon generator.
func NewGenerator(seed int64, genre string) *Generator {
	return &Generator{
		Seed:  seed,
		Genre: genre,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// Generate creates a new dungeon with the given parameters.
func (g *Generator) Generate(width, height, depth int) *Dungeon {
	d := &Dungeon{
		Width:   width,
		Height:  height,
		Tiles:   make([][]TileType, height),
		Rooms:   make([]*Room, 0),
		Genre:   g.Genre,
		Seed:    g.Seed,
		Depth:   depth,
		Palette: g.getPalette(),
	}

	// Initialize all tiles as walls
	for y := 0; y < height; y++ {
		d.Tiles[y] = make([]TileType, width)
		for x := 0; x < width; x++ {
			d.Tiles[y][x] = TileWall
		}
	}

	// Generate rooms using BSP
	g.generateBSP(d, 1, 1, width-2, height-2, 0)

	// Connect rooms with corridors
	g.connectRooms(d)

	// Add special room features based on depth
	g.addSpecialRooms(d, depth)

	// Add traps and puzzles
	g.addTrapsAndPuzzles(d, depth)

	// Mark entrance and exit
	g.markEntranceExit(d)

	return d
}

// generateBSP recursively splits space using Binary Space Partitioning.
func (g *Generator) generateBSP(d *Dungeon, x, y, w, h, depth int) {
	minRoomSize := 6
	maxDepth := 4

	if g.shouldCreateRoom(w, h, depth, minRoomSize, maxDepth) {
		g.createRoom(d, x, y, w, h)
		return
	}

	horizontal := g.decideSplitDirection(w, h)
	g.performBSPSplit(d, x, y, w, h, depth, horizontal, minRoomSize)
}

// shouldCreateRoom determines if this space should become a room instead of splitting further.
func (g *Generator) shouldCreateRoom(w, h, depth, minRoomSize, maxDepth int) bool {
	return w < minRoomSize*2 || h < minRoomSize*2 || depth >= maxDepth
}

// decideSplitDirection chooses horizontal or vertical split based on aspect ratio.
func (g *Generator) decideSplitDirection(w, h int) bool {
	horizontal := g.rng.Float64() < 0.5
	if float64(w)/float64(h) > 1.25 {
		horizontal = false // Split vertically if too wide
	} else if float64(h)/float64(w) > 1.25 {
		horizontal = true // Split horizontally if too tall
	}
	return horizontal
}

// performBSPSplit splits the space and recursively generates sub-spaces.
func (g *Generator) performBSPSplit(d *Dungeon, x, y, w, h, depth int, horizontal bool, minRoomSize int) {
	if horizontal {
		g.splitHorizontally(d, x, y, w, h, depth, minRoomSize)
	} else {
		g.splitVertically(d, x, y, w, h, depth, minRoomSize)
	}
}

// splitHorizontally performs a horizontal split of the space.
func (g *Generator) splitHorizontally(d *Dungeon, x, y, w, h, depth, minRoomSize int) {
	splitRange := h - minRoomSize*2
	if splitRange <= 0 {
		g.createRoom(d, x, y, w, h)
		return
	}
	split := minRoomSize + g.rng.Intn(splitRange)
	g.generateBSP(d, x, y, w, split, depth+1)
	g.generateBSP(d, x, y+split, w, h-split, depth+1)
}

// splitVertically performs a vertical split of the space.
func (g *Generator) splitVertically(d *Dungeon, x, y, w, h, depth, minRoomSize int) {
	splitRange := w - minRoomSize*2
	if splitRange <= 0 {
		g.createRoom(d, x, y, w, h)
		return
	}
	split := minRoomSize + g.rng.Intn(splitRange)
	g.generateBSP(d, x, y, split, h, depth+1)
	g.generateBSP(d, x+split, y, w-split, h, depth+1)
}

// calculateRoomDimensions computes random room dimensions within bounds.
func (g *Generator) calculateRoomDimensions(maxW, maxH, minSize int) (int, int, bool) {
	if maxW <= minSize+2 || maxH <= minSize+2 {
		return 0, 0, false
	}
	roomW := minSize + g.rng.Intn(maxW-minSize-2)
	roomH := minSize + g.rng.Intn(maxH-minSize-2)
	return roomW, roomH, true
}

// calculateRoomOffset computes random offset for room placement within space.
func (g *Generator) calculateRoomOffset(maxDim, roomDim int) int {
	offsetRange := maxDim - roomDim - 1
	if offsetRange <= 0 {
		offsetRange = 1
	}
	return g.rng.Intn(offsetRange)
}

// clampRoomToBounds adjusts room dimensions to fit within dungeon bounds.
func clampRoomToBounds(roomPos, roomDim, dungeonDim, minSize int) (int, bool) {
	if roomPos+roomDim >= dungeonDim-1 {
		roomDim = dungeonDim - roomPos - 2
	}
	return roomDim, roomDim >= minSize
}

// carveFloorTiles marks floor tiles in the dungeon grid for a room.
func carveFloorTiles(d *Dungeon, roomX, roomY, roomW, roomH int) {
	for ry := roomY; ry < roomY+roomH; ry++ {
		for rx := roomX; rx < roomX+roomW; rx++ {
			if rx >= 0 && rx < d.Width && ry >= 0 && ry < d.Height {
				d.Tiles[ry][rx] = TileFloor
			}
		}
	}
}

// createRoom carves a room into the dungeon.
func (g *Generator) createRoom(d *Dungeon, x, y, maxW, maxH int) {
	minSize := 4

	roomW, roomH, ok := g.calculateRoomDimensions(maxW, maxH, minSize)
	if !ok {
		return
	}

	offsetX := g.calculateRoomOffset(maxW, roomW)
	offsetY := g.calculateRoomOffset(maxH, roomH)

	roomX := x + offsetX + 1
	roomY := y + offsetY + 1

	roomW, ok = clampRoomToBounds(roomX, roomW, d.Width, minSize)
	if !ok {
		return
	}
	roomH, ok = clampRoomToBounds(roomY, roomH, d.Height, minSize)
	if !ok {
		return
	}

	room := &Room{
		ID:        len(d.Rooms),
		X:         roomX,
		Y:         roomY,
		Width:     roomW,
		Height:    roomH,
		Type:      RoomNormal,
		Connected: make([]int, 0),
		Genre:     g.Genre,
	}

	carveFloorTiles(d, roomX, roomY, roomW, roomH)
	d.Rooms = append(d.Rooms, room)
}

// connectRooms creates corridors between rooms ensuring full connectivity.
func (g *Generator) connectRooms(d *Dungeon) {
	if len(d.Rooms) < 2 {
		return
	}

	// Connect each room to the next (ensures all rooms are reachable)
	for i := 0; i < len(d.Rooms)-1; i++ {
		g.carveCorridor(d, d.Rooms[i], d.Rooms[i+1])
		d.Rooms[i].Connected = append(d.Rooms[i].Connected, d.Rooms[i+1].ID)
		d.Rooms[i+1].Connected = append(d.Rooms[i+1].Connected, d.Rooms[i].ID)
	}

	// Add some random extra connections for variety
	extraConnections := len(d.Rooms) / 3
	for i := 0; i < extraConnections; i++ {
		a := g.rng.Intn(len(d.Rooms))
		b := g.rng.Intn(len(d.Rooms))
		if a != b && !g.isConnected(d.Rooms[a], b) {
			g.carveCorridor(d, d.Rooms[a], d.Rooms[b])
			d.Rooms[a].Connected = append(d.Rooms[a].Connected, b)
			d.Rooms[b].Connected = append(d.Rooms[b].Connected, a)
		}
	}
}

// isConnected checks if a room is connected to another room ID.
func (g *Generator) isConnected(room *Room, otherID int) bool {
	for _, id := range room.Connected {
		if id == otherID {
			return true
		}
	}
	return false
}

// carveCorridor creates a corridor between two rooms.
func (g *Generator) carveCorridor(d *Dungeon, a, b *Room) {
	// Get center points of rooms
	ax := a.X + a.Width/2
	ay := a.Y + a.Height/2
	bx := b.X + b.Width/2
	by := b.Y + b.Height/2

	// Carve L-shaped corridor
	if g.rng.Float64() < 0.5 {
		// Horizontal then vertical
		g.carveHorizontal(d, ax, bx, ay)
		g.carveVertical(d, bx, ay, by)
	} else {
		// Vertical then horizontal
		g.carveVertical(d, ax, ay, by)
		g.carveHorizontal(d, ax, bx, by)
	}
}

// carveHorizontal carves a horizontal corridor.
func (g *Generator) carveHorizontal(d *Dungeon, x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if x >= 0 && x < d.Width && y >= 0 && y < d.Height {
			d.Tiles[y][x] = TileFloor
		}
	}
}

// carveVertical carves a vertical corridor.
func (g *Generator) carveVertical(d *Dungeon, x, y1, y2 int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if x >= 0 && x < d.Width && y >= 0 && y < d.Height {
			d.Tiles[y][x] = TileFloor
		}
	}
}

// addSpecialRooms assigns room types based on depth.
func (g *Generator) addSpecialRooms(d *Dungeon, depth int) {
	if len(d.Rooms) < 3 {
		return
	}

	// Boss room at deep dungeons
	if depth >= 3 && len(d.Rooms) >= 5 {
		bossRoom := d.Rooms[len(d.Rooms)-1]
		bossRoom.Type = RoomBoss
		bossRoom.IsBossRoom = true
		// Mark boss area
		g.markBossArea(d, bossRoom)
	}

	// Treasure room (1 per dungeon)
	treasureIdx := 1 + g.rng.Intn(len(d.Rooms)-2)
	d.Rooms[treasureIdx].Type = RoomTreasure
}

// markBossArea marks tiles in a boss room.
func (g *Generator) markBossArea(d *Dungeon, room *Room) {
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			if x >= 0 && x < d.Width && y >= 0 && y < d.Height {
				if d.Tiles[y][x] == TileFloor {
					d.Tiles[y][x] = TileBossArena
				}
			}
		}
	}
}

// addTrapsAndPuzzles places traps and puzzles based on depth.
func (g *Generator) addTrapsAndPuzzles(d *Dungeon, depth int) {
	// More traps at higher depths
	trapChance := 0.1 + float64(depth)*0.05
	puzzleChance := 0.15

	for _, room := range d.Rooms {
		if room.Type == RoomBoss || room.Type == RoomEntrance || room.Type == RoomExit {
			continue
		}

		if g.rng.Float64() < trapChance {
			room.HasTrap = true
			room.Type = RoomTrap
			g.placeTrap(d, room)
		} else if g.rng.Float64() < puzzleChance {
			room.HasPuzzle = true
			room.Type = RoomPuzzle
			g.placePuzzle(d, room)
		}
	}
}

// placeTrap places a trap tile in a room.
func (g *Generator) placeTrap(d *Dungeon, room *Room) {
	x := room.X + 1 + g.rng.Intn(room.Width-2)
	y := room.Y + 1 + g.rng.Intn(room.Height-2)
	if x >= 0 && x < d.Width && y >= 0 && y < d.Height {
		d.Tiles[y][x] = TileTrap
	}
}

// placePuzzle places a puzzle tile in a room.
func (g *Generator) placePuzzle(d *Dungeon, room *Room) {
	x := room.X + room.Width/2
	y := room.Y + room.Height/2
	if x >= 0 && x < d.Width && y >= 0 && y < d.Height {
		d.Tiles[y][x] = TilePuzzle
	}
}

// markEntranceExit marks entrance and exit rooms.
func (g *Generator) markEntranceExit(d *Dungeon) {
	if len(d.Rooms) < 2 {
		return
	}
	g.markEntranceRoom(d)
	g.markExitRoom(d)
}

// markEntranceRoom marks the first room as entrance.
func (g *Generator) markEntranceRoom(d *Dungeon) {
	entrance := d.Rooms[0]
	entrance.Type = RoomEntrance
	cx, cy := entrance.X+entrance.Width/2, entrance.Y+entrance.Height/2
	if g.isValidTilePosition(d, cx, cy) {
		d.Tiles[cy][cx] = TileEntrance
	}
}

// markExitRoom marks the last room as exit (unless it's a boss room).
func (g *Generator) markExitRoom(d *Dungeon) {
	exit := d.Rooms[len(d.Rooms)-1]
	if exit.Type != RoomBoss {
		exit.Type = RoomExit
	}
	cx, cy := exit.X+exit.Width/2, exit.Y+exit.Height/2
	if g.isValidTilePosition(d, cx, cy) && d.Tiles[cy][cx] != TileBossArena {
		d.Tiles[cy][cx] = TileExit
	}
}

// isValidTilePosition checks if coordinates are within dungeon bounds.
func (g *Generator) isValidTilePosition(d *Dungeon, x, y int) bool {
	return x >= 0 && x < d.Width && y >= 0 && y < d.Height
}

// getPalette returns the genre-appropriate tile palette.
func (g *Generator) getPalette() TilePalette {
	if palette, ok := GenrePalettes[g.Genre]; ok {
		return palette
	}
	return GenrePalettes["fantasy"]
}

// ValidateConnectivity checks that all rooms are reachable from entrance.
// Per AC: 0 unreachable rooms.
func (d *Dungeon) ValidateConnectivity() bool {
	if len(d.Rooms) == 0 {
		return true
	}
	visited := d.traverseRooms()
	return len(visited) == len(d.Rooms)
}

// traverseRooms performs BFS from entrance room and returns visited room IDs.
func (d *Dungeon) traverseRooms() map[int]bool {
	visited := make(map[int]bool)
	queue := []int{0} // Start from entrance (room 0)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		queue = d.enqueueConnectedRooms(queue, visited, current)
	}
	return visited
}

// enqueueConnectedRooms adds unvisited connected rooms to the queue.
func (d *Dungeon) enqueueConnectedRooms(queue []int, visited map[int]bool, roomIndex int) []int {
	room := d.Rooms[roomIndex]
	for _, connectedID := range room.Connected {
		if !visited[connectedID] {
			queue = append(queue, connectedID)
		}
	}
	return queue
}

// CountRoomsByType returns the count of rooms with a given type.
func (d *Dungeon) CountRoomsByType(roomType RoomType) int {
	count := 0
	for _, room := range d.Rooms {
		if room.Type == roomType {
			count++
		}
	}
	return count
}

// HasBossRoom returns whether the dungeon has a boss room.
func (d *Dungeon) HasBossRoom() bool {
	for _, room := range d.Rooms {
		if room.IsBossRoom {
			return true
		}
	}
	return false
}

// TrapCount returns the number of trap rooms.
func (d *Dungeon) TrapCount() int {
	count := 0
	for _, room := range d.Rooms {
		if room.HasTrap {
			count++
		}
	}
	return count
}

// PuzzleCount returns the number of puzzle rooms.
func (d *Dungeon) PuzzleCount() int {
	count := 0
	for _, room := range d.Rooms {
		if room.HasPuzzle {
			count++
		}
	}
	return count
}
