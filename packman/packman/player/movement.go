package player

// Movement handles player movement logic
type Movement struct {
	speed        float64
	moveProgress float64
	targetX      int
	targetY      int
	moving       bool
}

// NewMovement creates a new movement handler
func NewMovement(speed float64) *Movement {
	return &Movement{
		speed:  speed,
		moving: false,
	}
}

// Update updates the movement state
func (m *Movement) Update(dt float64, currentX, currentY, gridX, gridY int) (float64, float64) {
	// Calculate pixel position based on grid position
	targetPixelX := float64(gridX)
	targetPixelY := float64(gridY)

	return targetPixelX, targetPixelY
}

// GetSpeed returns the movement speed
func (m *Movement) GetSpeed() float64 {
	return m.speed
}

// SetSpeed sets the movement speed
func (m *Movement) SetSpeed(speed float64) {
	m.speed = speed
}
