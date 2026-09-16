package game

// LevelManager handles level progression and difficulty
type LevelManager struct {
	CurrentLevel int
}

// NewLevelManager creates a new level manager
func NewLevelManager() *LevelManager {
	return &LevelManager{
		CurrentLevel: 1,
	}
}

// GetCurrentLevel returns the current level number
func (lm *LevelManager) GetCurrentLevel() int {
	return lm.CurrentLevel
}

// NextLevel advances to the next level
func (lm *LevelManager) NextLevel() {
	lm.CurrentLevel++
}

// Reset resets the level to 1
func (lm *LevelManager) Reset() {
	lm.CurrentLevel = 1
}

// GetEnemySpeedMultiplier returns the speed multiplier for enemies based on current level
func (lm *LevelManager) GetEnemySpeedMultiplier() float64 {
	// Each level increases speed by 10%, capped at 2.0
	multiplier := 1.0 + float64(lm.CurrentLevel-1)*0.1
	if multiplier > 2.0 {
		multiplier = 2.0
	}
	return multiplier
}

// GetEnemyCount returns the number of enemies for the current level
func (lm *LevelManager) GetEnemyCount() int {
	// Start with 2 enemies, add 1 every 2 levels, max 6
	count := 2 + (lm.CurrentLevel-1)/2
	if count > 6 {
		count = 6
	}
	return count
}

// GetPowerModeDuration returns the power mode duration in seconds for the current level
func (lm *LevelManager) GetPowerModeDuration() float64 {
	// Base 8 seconds, decreases by 0.5 per level, minimum 3 seconds
	duration := 8.0 - float64(lm.CurrentLevel-1)*0.5
	if duration < 3.0 {
		duration = 3.0
	}
	return duration
}

// GetMazeLayout returns the maze layout for the current level
func (lm *LevelManager) GetMazeLayout() []string {
	// For now, return the default maze. In the future, different levels
	// could have different layouts
	return DefaultMazeLayout
}
