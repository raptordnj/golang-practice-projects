package game

// Tile represents a single cell in the maze grid
type Tile int

const (
	// TileWall is an impassable wall
	TileWall Tile = iota
	// TileEmpty is an empty walkable space
	TileEmpty
	// TilePellet contains a normal pellet
	TilePellet
	// TilePowerPellet contains a power pellet
	TilePowerPellet
	// TilePlayerStart is the player's starting position
	TilePlayerStart
	// TileEnemyStart is an enemy's starting position
	TileEnemyStart
)

// Maze represents the game maze as a grid of tiles
type Maze struct {
	Width   int
	Height  int
	Tiles   [][]Tile
	Pellets int
}

// DefaultMazeLayout is the default maze layout using characters:
// # = wall, . = pellet, o = power pellet, P = player start, E = enemy start, ' ' = empty
var DefaultMazeLayout = []string{
	"################",
	"#..............#",
	"#.####.#####.#.#",
	"#.#..........#.#",
	"#.#.##.###.#.#.#",
	"#....#...#.....#",
	"#.##.###.#.###.#",
	"#..............#",
	"#.###.#####.##.#",
	"#..............#",
	"#.##.#.###.#.#.#",
	"#....#...#..E..#",
	"#.####.#####.#.#",
	"#..............#",
	"################",
}

// NewMaze creates a new maze from a string layout
func NewMaze(layout []string) *Maze {
	height := len(layout)
	width := 0
	if height > 0 {
		width = len(layout[0])
	}

	tiles := make([][]Tile, height)
	pellets := 0

	for y, row := range layout {
		tiles[y] = make([]Tile, width)
		for x, ch := range row {
			switch ch {
			case '#':
				tiles[y][x] = TileWall
			case '.':
				tiles[y][x] = TilePellet
				pellets++
			case 'o':
				tiles[y][x] = TilePowerPellet
				pellets++
			case 'P':
				tiles[y][x] = TilePlayerStart
			case 'E':
				tiles[y][x] = TileEnemyStart
			default:
				tiles[y][x] = TileEmpty
			}
		}
	}

	return &Maze{
		Width:   width,
		Height:  height,
		Tiles:   tiles,
		Pellets: pellets,
	}
}

// IsWall returns true if the given position is a wall
func (m *Maze) IsWall(x, y int) bool {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return true
	}
	return m.Tiles[y][x] == TileWall
}

// IsWalkable returns true if the given position can be moved into
func (m *Maze) IsWalkable(x, y int) bool {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return false
	}
	return m.Tiles[y][x] != TileWall
}

// GetTile returns the tile at the given position
func (m *Maze) GetTile(x, y int) Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return TileWall
	}
	return m.Tiles[y][x]
}

// SetTile sets the tile at the given position
func (m *Maze) SetTile(x, y int, tile Tile) {
	if x >= 0 && x < m.Width && y >= 0 && y < m.Height {
		m.Tiles[y][x] = tile
	}
}

// CollectPellet removes a pellet at the given position and returns the points earned
func (m *Maze) CollectPellet(x, y int) int {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return 0
	}

	tile := m.Tiles[y][x]
	switch tile {
	case TilePellet:
		m.Tiles[y][x] = TileEmpty
		m.Pellets--
		return 10
	case TilePowerPellet:
		m.Tiles[y][x] = TileEmpty
		m.Pellets--
		return 50
	}
	return 0
}

// HasPellets returns true if there are still pellets remaining
func (m *Maze) HasPellets() bool {
	return m.Pellets > 0
}

// FindPlayerStart finds the player's starting position
func (m *Maze) FindPlayerStart() (int, int) {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.Tiles[y][x] == TilePlayerStart {
				return x, y
			}
		}
	}
	return 1, 1 // fallback
}

// FindEnemyStarts finds all enemy starting positions
func (m *Maze) FindEnemyStarts() [][2]int {
	var positions [][2]int
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.Tiles[y][x] == TileEnemyStart {
				positions = append(positions, [2]int{x, y})
			}
		}
	}
	return positions
}
