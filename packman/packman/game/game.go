package game

// Direction represents movement direction
type Direction int

const (
	// DirNone means no direction
	DirNone Direction = iota
	// DirUp moves up
	DirUp
	// DirDown moves down
	DirDown
	// DirLeft moves left
	DirLeft
	// DirRight moves right
	DirRight
)

// String returns the string representation of a direction
func (d Direction) String() string {
	switch d {
	case DirUp:
		return "UP"
	case DirDown:
		return "DOWN"
	case DirLeft:
		return "LEFT"
	case DirRight:
		return "RIGHT"
	default:
		return "NONE"
	}
}

// Game represents the main game state and logic
type Game struct {
	State          GameState
	Maze           *Maze
	Player         *PlayerState
	Enemies        []*EnemyState
	ScoreManager   *ScoreManager
	LevelManager   *LevelManager
	Collision      *CollisionDetector
	Loop           *GameLoop

	// Power mode state
	PowerMode       bool
	PowerModeTimer  float64

	// Player starting position (for respawn)
	playerStartX    int
	playerStartY    int

	// Enemy starting positions (for respawn)
	enemyStarts     [][2]int

	// Game loop callback
	onStateChange   func(oldState, newState GameState)
}

// PlayerState represents the player's current state
type PlayerState struct {
	X             int
	Y             int
	Direction     Direction
	NextDirection Direction
	Lives         int
	IsAlive       bool
	DeathTimer    float64
	AnimFrame     int
	AnimTimer     float64
}

// EnemyState represents an enemy's current state
type EnemyState struct {
	X           int
	Y           int
	Direction   Direction
	IsVulnerable bool
	IsAlive     bool
	AnimFrame   int
	AnimTimer   float64
	Type        EnemyType
}

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

// NewGame creates a new game instance
func NewGame() *Game {
	maze := NewMaze(DefaultMazeLayout)
	playerStartX, playerStartY := maze.FindPlayerStart()
	enemyStarts := maze.FindEnemyStarts()

	// Ensure at least one enemy start exists
	if len(enemyStarts) == 0 {
		enemyStarts = append(enemyStarts, [2]int{8, 7})
	}

	// Add more enemy starts if needed
	for len(enemyStarts) < 4 {
		enemyStarts = append(enemyStarts, enemyStarts[0])
	}

	game := &Game{
		State:         StateTitle,
		Maze:          maze,
		ScoreManager:  NewScoreManager(),
		LevelManager:  NewLevelManager(),
		Collision:     NewCollisionDetector(),
		Loop:          NewGameLoop(60),
		PowerMode:     false,
		PowerModeTimer: 0,
		playerStartX:  playerStartX,
		playerStartY:  playerStartY,
		enemyStarts:   make([][2]int, len(enemyStarts)),
	}

	copy(game.enemyStarts, enemyStarts)

	game.Player = &PlayerState{
		X:             playerStartX,
		Y:             playerStartY,
		Direction:     DirNone,
		NextDirection: DirNone,
		Lives:         3,
		IsAlive:       true,
	}

	return game
}

// Initialize prepares the game for a new session
func (g *Game) Initialize() {
	g.ScoreManager.ResetCurrent()
	g.LevelManager.Reset()
	g.setupLevel()
	g.setState(StateTitle)
}

// Start begins the game from the title screen
func (g *Game) Start() {
	g.ScoreManager.ResetCurrent()
	g.LevelManager.Reset()
	g.setupLevel()
	g.setState(StateReady)
}

// setupLevel sets up the current level
func (g *Game) setupLevel() {
	// Create new maze for the level
	layout := g.LevelManager.GetMazeLayout()
	g.Maze = NewMaze(layout)

	playerStartX, playerStartY := g.Maze.FindPlayerStart()
	g.playerStartX = playerStartX
	g.playerStartY = playerStartY

	// Find enemy starts
	g.enemyStarts = g.Maze.FindEnemyStarts()
	for len(g.enemyStarts) < 4 {
		g.enemyStarts = append(g.enemyStarts, g.enemyStarts[0])
	}

	// Reset player
	g.Player.X = playerStartX
	g.Player.Y = playerStartY
	g.Player.Direction = DirNone
	g.Player.NextDirection = DirNone
	g.Player.IsAlive = true

	// Create enemies based on level
	numEnemies := g.LevelManager.GetEnemyCount()
	g.Enemies = make([]*EnemyState, numEnemies)
	enemyTypes := []EnemyType{EnemyRandom, EnemyChaser, EnemyHorizontal, EnemyVertical}

	for i := 0; i < numEnemies; i++ {
		start := g.enemyStarts[i%len(g.enemyStarts)]
		g.Enemies[i] = &EnemyState{
			X:           start[0],
			Y:           start[1],
			Direction:   DirNone,
			IsVulnerable: false,
			IsAlive:     true,
			Type:        enemyTypes[i%len(enemyTypes)],
		}
	}

	// Reset power mode
	g.PowerMode = false
	g.PowerModeTimer = 0
}

// Update updates the game state
func (g *Game) Update(dt float64) {
	switch g.State {
	case StateReady:
		g.updateReady(dt)
	case StatePlaying:
		g.updatePlaying(dt)
	case StateLevelComplete:
		g.updateLevelComplete(dt)
	case StateGameOver:
		g.updateGameOver(dt)
	}
}

// updateReady handles the ready state countdown
func (g *Game) updateReady(dt float64) {
	// Short delay before starting
	g.PowerModeTimer -= dt
	if g.PowerModeTimer <= 0 {
		g.setState(StatePlaying)
		g.Loop.Start()
	}
}

// updatePlaying handles active gameplay
func (g *Game) updatePlaying(dt float64) {
	// Update power mode timer
	if g.PowerMode {
		g.PowerModeTimer -= dt
		if g.PowerModeTimer <= 0 {
			g.PowerMode = false
			// Reset enemy vulnerability
			for _, enemy := range g.Enemies {
				enemy.IsVulnerable = false
			}
		}
	}

	// Update player animation
	g.Player.AnimTimer += dt
	if g.Player.AnimTimer >= 0.1 {
		g.Player.AnimTimer = 0
		g.Player.AnimFrame = (g.Player.AnimFrame + 1) % 4
	}

	// Update enemy animations
	for _, enemy := range g.Enemies {
		enemy.AnimTimer += dt
		if enemy.AnimTimer >= 0.1 {
			enemy.AnimTimer = 0
			enemy.AnimFrame = (enemy.AnimFrame + 1) % 4
		}
	}

	// Check if player is dead
	if !g.Player.IsAlive {
		g.Player.DeathTimer -= dt
		if g.Player.DeathTimer <= 0 {
			if g.Player.Lives <= 0 {
				g.setState(StateGameOver)
			} else {
				g.respawnPlayer()
			}
		}
		return
	}

	// Check for pellet collection
	points := g.Maze.CollectPellet(g.Player.X,	g.Player.Y)
	if points > 0 {
		g.ScoreManager.AddScore(points)
		if points == 50 {
			g.activatePowerMode()
		}

		// Check if level is complete
		if !g.Maze.HasPellets() {
			g.setState(StateLevelComplete)
			return
		}
	}

	// Check for enemy collisions
	g.checkEnemyCollisions()
}

// updateLevelComplete handles the level complete state
func (g *Game) updateLevelComplete(dt float64) {
	g.PowerModeTimer -= dt
	if g.PowerModeTimer <= 0 {
		g.LevelManager.NextLevel()
		g.setupLevel()
		g.setState(StateReady)
		g.PowerModeTimer = 2.0 // 2 second delay
	}
}

// updateGameOver handles the game over state
func (g *Game) updateGameOver(dt float64) {
	// Wait for player input to restart
}

// activatePowerMode activates power mode
func (g *Game) activatePowerMode() {
	g.PowerMode = true
	g.PowerModeTimer = g.LevelManager.GetPowerModeDuration()
	for _, enemy := range g.Enemies {
		enemy.IsVulnerable = true
	}
}

// checkEnemyCollisions checks for collisions between player and enemies
func (g *Game) checkEnemyCollisions() {
	for _, enemy := range g.Enemies {
		if !enemy.IsAlive {
			continue
		}

		if g.Collision.PlayerEnemyCollision(
			Position{X: g.Player.X, Y: g.Player.Y},
			Position{X: enemy.X, Y: enemy.Y},
		) {
			if enemy.IsVulnerable {
				// Enemy is eaten
				enemy.IsAlive = false
				g.ScoreManager.AddScore(200)
			} else {
				// Player dies
				g.killPlayer()
				return
			}
		}
	}

	// Check if all dead enemies should respawn (after a delay)
	// For simplicity, we'll just keep them dead
}

// killPlayer kills the player
func (g *Game) killPlayer() {
	g.Player.IsAlive = false
	g.Player.Lives--
	g.Player.DeathTimer = 2.0 // 2 second death animation
}

// respawnPlayer respawns the player at the start position
func (g *Game) respawnPlayer() {
	g.Player.X = g.playerStartX
	g.Player.Y = g.playerStartY
	g.Player.Direction = DirNone
	g.Player.NextDirection = DirNone
	g.Player.IsAlive = true
	g.Player.DeathTimer = 0

	// Respawn enemies
	for i, enemy := range g.Enemies {
		start := g.enemyStarts[i%len(g.enemyStarts)]
		enemy.X = start[0]
		enemy.Y = start[1]
		enemy.Direction = DirNone
		enemy.IsAlive = true
		enemy.IsVulnerable = false
	}
}

// MovePlayer moves the player in the given direction
func (g *Game) MovePlayer(dir Direction) {
	if g.State != StatePlaying {
		return
	}

	g.Player.NextDirection = dir

	// Try to move in the requested direction
	if g.tryMove(dir) {
		g.Player.Direction = dir
	}
}

// MovePlayerContinuously continues moving in the current direction
func (g *Game) MovePlayerContinuously() {
	if g.State != StatePlaying || !g.Player.IsAlive {
		return
	}

	// First try the current direction
	if g.tryMove(g.Player.Direction) {
		return
	}

	// If blocked, try the next direction
	if g.Player.NextDirection != g.Player.Direction {
		if g.tryMove(g.Player.NextDirection) {
			g.Player.Direction = g.Player.NextDirection
		}
	}
}

// tryMove attempts to move the player in the given direction
func (g *Game) tryMove(dir Direction) bool {
	newX, newY := g.Player.X, g.Player.Y

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

	if g.Maze.IsWalkable(newX, newY) {
		g.Player.X = newX
		g.Player.Y = newY
		return true
	}
	return false
}

// UpdateEnemies updates all enemy positions based on their AI
func (g *Game) UpdateEnemies(dt float64) {
	if g.State != StatePlaying {
		return
	}

	speedMultiplier := g.LevelManager.GetEnemySpeedMultiplier()
	moveInterval := 0.2 / speedMultiplier // Base 200ms between moves

	for _, enemy := range g.Enemies {
		if !enemy.IsAlive {
			continue
		}

		// Accumulate time for movement
		enemy.AnimTimer += dt

		// Simple movement timer - enemies move every N ms
		if enemy.AnimTimer >= moveInterval {
			enemy.AnimTimer = 0
			g.moveEnemy(enemy)
		}
	}
}

// moveEnemy moves a single enemy based on its AI type
func (g *Game) moveEnemy(enemy *EnemyState) {
	var dir Direction

	switch enemy.Type {
	case EnemyRandom:
		dir = g.getRandomDirection(enemy)
	case EnemyChaser:
		dir = g.getChaseDirection(enemy)
	case EnemyHorizontal:
		dir = g.getHorizontalPreferredDirection(enemy)
	case EnemyVertical:
		dir = g.getVerticalPreferredDirection(enemy)
	}

	// Try to move in the chosen direction
	newX, newY := enemy.X, enemy.Y
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

	if g.Maze.IsWalkable(newX, newY) {
		enemy.X = newX
		enemy.Y = newY
		enemy.Direction = dir
	} else {
		// Try a random valid direction
		dir = g.getAnyValidDirection(enemy)
		if dir != DirNone {
			enemy.Direction = dir
			switch dir {
			case DirUp:
				enemy.Y--
			case DirDown:
				enemy.Y++
			case DirLeft:
				enemy.X--
			case DirRight:
				enemy.X++
			}
		}
	}
}

// getRandomDirection returns a random valid direction
func (g *Game) getRandomDirection(enemy *EnemyState) Direction {
	// 70% chance to continue in current direction, 30% chance to turn
	if enemy.Direction != DirNone && g.getRandomNumber() < 0.7 {
		return enemy.Direction
	}
	return g.getAnyValidDirection(enemy)
}

// getChaseDirection returns a direction that moves toward the player
func (g *Game) getChaseDirection(enemy *EnemyState) Direction {
	// 80% chance to chase, 20% random
	if g.getRandomNumber() < 0.2 {
		return g.getRandomDirection(enemy)
	}

	dx := g.Player.X - enemy.X
	dy := g.Player.Y - enemy.Y

	// Prefer the axis with greater distance
	var preferredDirs []Direction
	if abs(dx) > abs(dy) {
		if dx > 0 {
			preferredDirs = []Direction{DirRight, DirUp, DirDown, DirLeft}
		} else {
			preferredDirs = []Direction{DirLeft, DirUp, DirDown, DirRight}
		}
	} else {
		if dy > 0 {
			preferredDirs = []Direction{DirDown, DirLeft, DirRight, DirUp}
		} else {
			preferredDirs = []Direction{DirUp, DirLeft, DirRight, DirDown}
		}
	}

	for _, dir := range preferredDirs {
		if g.canMoveInDirection(enemy, dir) {
			return dir
		}
	}

	return DirNone
}

// getHorizontalPreferredDirection prefers horizontal movement
func (g *Game) getHorizontalPreferredDirection(enemy *EnemyState) Direction {
	if g.getRandomNumber() < 0.6 {
		if g.Player.X < enemy.X && g.canMoveInDirection(enemy, DirLeft) {
			return DirLeft
		}
		if g.Player.X > enemy.X && g.canMoveInDirection(enemy, DirRight) {
			return DirRight
		}
	}
	return g.getAnyValidDirection(enemy)
}

// getVerticalPreferredDirection prefers vertical movement
func (g *Game) getVerticalPreferredDirection(enemy *EnemyState) Direction {
	if g.getRandomNumber() < 0.6 {
		if g.Player.Y < enemy.Y && g.canMoveInDirection(enemy, DirUp) {
			return DirUp
		}
		if g.Player.Y > enemy.Y && g.canMoveInDirection(enemy, DirDown) {
			return DirDown
		}
	}
	return g.getAnyValidDirection(enemy)
}

// getAnyValidDirection returns any valid direction the enemy can move
func (g *Game) getAnyValidDirection(enemy *EnemyState) Direction {
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}
	// Shuffle for variety
	for i := len(directions) - 1; i > 0; i-- {
		j := int(g.getRandomNumber() * float64(i+1))
		directions[i], directions[j] = directions[j], directions[i]
	}

	for _, dir := range directions {
		if g.canMoveInDirection(enemy, dir) {
			return dir
		}
	}

	return DirNone
}

// canMoveInDirection checks if the enemy can move in the given direction
func (g *Game) canMoveInDirection(enemy *EnemyState, dir Direction) bool {
	newX, newY := enemy.X, enemy.Y
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
	return g.Maze.IsWalkable(newX, newY)
}

// Pseudo-random number generator state
var randomSeed = uint32(1)

// getRandomNumber returns a pseudo-random number between 0 and 1
func (g *Game) getRandomNumber() float64 {
	// Simple LCG for deterministic behavior
	randomSeed = randomSeed*1103515245 + 12345
	return float64(randomSeed&0x7fffffff) / float64(0x7fffffff)
}

// setState changes the game state and calls the callback
func (g *Game) setState(newState GameState) {
	oldState := g.State
	g.State = newState

	// Set timers for specific states
	switch newState {
	case StateReady:
		g.PowerModeTimer = 2.0 // 2 second countdown
	case StateLevelComplete:
		g.PowerModeTimer = 2.0 // 2 second delay before next level
	}

	if g.onStateChange != nil {
		g.onStateChange(oldState, newState)
	}
}

// SetStateChangeCallback sets the callback for state changes
func (g *Game) SetStateChangeCallback(callback func(oldState, newState GameState)) {
	g.onStateChange = callback
}

// Pause pauses the game
func (g *Game) Pause() {
	if g.State == StatePlaying {
		g.setState(StatePaused)
		g.Loop.Pause()
	}
}

// Resume resumes the game
func (g *Game) Resume() {
	if g.State == StatePaused {
		g.setState(StatePlaying)
		g.Loop.Resume()
	}
}

// TogglePause toggles pause state
func (g *Game) TogglePause() {
	if g.State == StatePlaying {
		g.Pause()
	} else if g.State == StatePaused {
		g.Resume()
	}
}

// Restart restarts the game from the beginning
func (g *Game) Restart() {
	g.ScoreManager.ResetCurrent()
	g.LevelManager.Reset()
	g.Player.Lives = 3
	g.setupLevel()
	g.setState(StateReady)
}

// GetReady returns to the title screen
func (g *Game) GetReady() {
	g.setState(StateTitle)
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
