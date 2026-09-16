package enemy

// EnemyType represents different enemy AI behaviors
type EnemyType int

const (
	// EnemyRandom moves randomly
	EnemyRandom EnemyType = iota
	// EnemyChaser tries to approach the player
	EnemyChaser
	// EnemyHorizontal prefers horizontal movement
	EnemyHorizontal
	// EnemyVertical prefers vertical movement
	EnemyVertical
)

// Enemy represents an enemy entity
type Enemy struct {
	X            int
	Y            int
	Direction    int
	IsVulnerable bool
	IsAlive      bool
	Type         EnemyType
}

// NewEnemy creates a new enemy
func NewEnemy(x, y int, enemyType EnemyType) *Enemy {
	return &Enemy{
		X:            x,
		Y:            y,
		Direction:    0,
		IsVulnerable: false,
		IsAlive:      true,
		Type:         enemyType,
	}
}

// Reset resets the enemy to initial state
func (e *Enemy) Reset(x, y int) {
	e.X = x
	e.Y = y
	e.Direction = 0
	e.IsVulnerable = false
	e.IsAlive = true
}

// GetPosition returns the enemy's position
func (e *Enemy) GetPosition() (int, int) {
	return e.X, e.Y
}

// SetPosition sets the enemy's position
func (e *Enemy) SetPosition(x, y int) {
	e.X = x
	e.Y = y
}

// SetVulnerable sets the enemy's vulnerable state
func (e *Enemy) SetVulnerable(vulnerable bool) {
	e.IsVulnerable = vulnerable
}

// GetType returns the enemy type
func (e *Enemy) GetType() EnemyType {
	return e.Type
}

// Kill kills the enemy
func (e *Enemy) Kill() {
	e.IsAlive = false
}

// Respawn respawns the enemy
func (e *Enemy) Respawn(x, y int) {
	e.X = x
	e.Y = y
	e.IsAlive = true
	e.IsVulnerable = false
}
