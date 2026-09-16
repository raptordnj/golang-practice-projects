package player

// Direction represents player movement direction
type Direction int

const (
	DirNone Direction = iota
	DirUp
	DirDown
	DirLeft
	DirRight
)

// Config holds player configuration
type Config struct {
	Color string
	Speed float64
	Size  float32
}

// DefaultConfig returns default player configuration
func DefaultConfig() Config {
	return Config{
		Color: "yellow",
		Speed: 150.0,
		Size:  8,
	}
}

// Player represents the player entity (game logic)
type Player struct {
	X             int
	Y             int
	Direction     Direction
	NextDirection Direction
	Lives         int
	IsAlive       bool
}

// NewPlayer creates a new player
func NewPlayer(x, y int) *Player {
	return &Player{
		X:             x,
		Y:             y,
		Direction:     DirNone,
		NextDirection: DirNone,
		Lives:         3,
		IsAlive:       true,
	}
}

// Reset resets the player to initial state
func (p *Player) Reset(x, y int) {
	p.X = x
	p.Y = y
	p.Direction = DirNone
	p.NextDirection = DirNone
	p.IsAlive = true
}

// GetPosition returns the player's position
func (p *Player) GetPosition() (int, int) {
	return p.X, p.Y
}

// SetPosition sets the player's position
func (p *Player) SetPosition(x, y int) {
	p.X = x
	p.Y = y
}

// SetDirection sets the player's direction
func (p *Player) SetDirection(dir Direction) {
	p.Direction = dir
}

// SetNextDirection sets the player's next direction
func (p *Player) SetNextDirection(dir Direction) {
	p.NextDirection = dir
}
