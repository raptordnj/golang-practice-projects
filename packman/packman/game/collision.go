package game

// Position represents a 2D grid position
type Position struct {
	X int
	Y int
}

// CollisionDetector handles collision detection between game entities
type CollisionDetector struct{}

// NewCollisionDetector creates a new collision detector
func NewCollisionDetector() *CollisionDetector {
	return &CollisionDetector{}
}

// PlayerEnemyCollision checks if the player and enemy are at the same position
func (c *CollisionDetector) PlayerEnemyCollision(playerPos, enemyPos Position) bool {
	return playerPos.X == enemyPos.X && playerPos.Y == enemyPos.Y
}

// PlayerEnemyCollisionDistance checks if the player is within a certain distance of an enemy
func (c *CollisionDetector) PlayerEnemyCollisionDistance(playerPos, enemyPos Position, distance float64) bool {
	dx := float64(playerPos.X - enemyPos.X)
	dy := float64(playerPos.Y - enemyPos.Y)
	return (dx*dx + dy*dy) <= distance*distance
}

// IsValidMove checks if a position is valid (not a wall and within bounds)
func (c *CollisionDetector) IsValidMove(maze *Maze, x, y int) bool {
	return maze.IsWalkable(x, y)
}

// IsSamePosition checks if two positions are the same
func (c *CollisionDetector) IsSamePosition(a, b Position) bool {
	return a.X == b.X && a.Y == b.Y
}
