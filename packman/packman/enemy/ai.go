package enemy

// AI handles enemy artificial intelligence
type AI struct{}

// NewAI creates a new AI handler
func NewAI() *AI {
	return &AI{}
}

// Direction constants for AI
const (
	DirNone  = 0
	DirUp    = 1
	DirDown  = 2
	DirLeft  = 3
	DirRight = 4
)

// DecideMovement calculates the enemy's next move
func (ai *AI) DecideMovement(e *Enemy, playerX, playerY int, isWalkable func(x, y int) bool) int {
	switch e.Type {
	case EnemyRandom:
		return ai.randomMove(e, isWalkable)
	case EnemyChaser:
		return ai.chaseMove(e, playerX, playerY, isWalkable)
	case EnemyHorizontal:
		return ai.horizontalPreferredMove(e, playerX, isWalkable)
	case EnemyVertical:
		return ai.verticalPreferredMove(e, playerY, isWalkable)
	}
	return DirNone
}

// randomMove returns a random valid direction
func (ai *AI) randomMove(e *Enemy, isWalkable func(x, y int) bool) int {
	directions := []int{DirUp, DirDown, DirLeft, DirRight}

	for _, dir := range directions {
		if ai.canMove(e, dir, isWalkable) {
			return dir
		}
	}
	return e.Direction
}

// chaseMove returns a direction toward the player
func (ai *AI) chaseMove(e *Enemy, playerX, playerY int, isWalkable func(x, y int) bool) int {
	dx := playerX - e.X
	dy := playerY - e.Y

	// Prioritize the axis with greater distance
	var preferredDirs []int
	if abs(dx) > abs(dy) {
		if dx > 0 {
			preferredDirs = []int{DirRight, DirUp, DirDown, DirLeft}
		} else {
			preferredDirs = []int{DirLeft, DirUp, DirDown, DirRight}
		}
	} else {
		if dy > 0 {
			preferredDirs = []int{DirDown, DirLeft, DirRight, DirUp}
		} else {
			preferredDirs = []int{DirUp, DirLeft, DirRight, DirDown}
		}
	}

	for _, dir := range preferredDirs {
		if ai.canMove(e, dir, isWalkable) {
			return dir
		}
	}
	return e.Direction
}

// horizontalPreferredMove prefers horizontal movement
func (ai *AI) horizontalPreferredMove(e *Enemy, playerX int, isWalkable func(x, y int) bool) int {
	if playerX < e.X && ai.canMove(e, DirLeft, isWalkable) {
		return DirLeft
	}
	if playerX > e.X && ai.canMove(e, DirRight, isWalkable) {
		return DirRight
	}
	return ai.randomMove(e, isWalkable)
}

// verticalPreferredMove prefers vertical movement
func (ai *AI) verticalPreferredMove(e *Enemy, playerY int, isWalkable func(x, y int) bool) int {
	if playerY < e.Y && ai.canMove(e, DirUp, isWalkable) {
		return DirUp
	}
	if playerY > e.Y && ai.canMove(e, DirDown, isWalkable) {
		return DirDown
	}
	return ai.randomMove(e, isWalkable)
}

// canMove checks if the enemy can move in the given direction
func (ai *AI) canMove(e *Enemy, dir int, isWalkable func(x, y int) bool) bool {
	newX, newY := e.X, e.Y

	switch dir {
	case DirUp:
		newY--
	case DirDown:
		newY++
	case DirLeft:
		newX--
	case DirRight:
		newX++
	}

	return isWalkable(newX, newY)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
