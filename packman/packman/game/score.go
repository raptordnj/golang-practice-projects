package game

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ScoreManager handles scoring and high score persistence
type ScoreManager struct {
	CurrentScore int
	HighScore    int
	savePath     string
}

// scoreData is used for JSON serialization
type scoreData struct {
	HighScore int `json:"high_score"`
}

// NewScoreManager creates a new score manager
func NewScoreManager() *ScoreManager {
	sm := &ScoreManager{
		CurrentScore: 0,
		HighScore:    0,
	}
	sm.loadHighScore()
	return sm
}

// NewScoreManagerWithPath creates a new score manager with a custom save path
func NewScoreManagerWithPath(savePath string) *ScoreManager {
	sm := &ScoreManager{
		CurrentScore: 0,
		HighScore:    0,
		savePath:     savePath,
	}
	sm.loadHighScore()
	return sm
}

// AddScore adds points to the current score
func (sm *ScoreManager) AddScore(points int) {
	sm.CurrentScore += points
	if sm.CurrentScore > sm.HighScore {
		sm.HighScore = sm.CurrentScore
		sm.saveHighScore()
	}
}

// ResetCurrent resets the current score to zero
func (sm *ScoreManager) ResetCurrent() {
	sm.CurrentScore = 0
}

// GetCurrent returns the current score
func (sm *ScoreManager) GetCurrent() int {
	return sm.CurrentScore
}

// GetHighScore returns the high score
func (sm *ScoreManager) GetHighScore() int {
	return sm.HighScore
}

// IsNewHighScore checks if the current score is a new high score
func (sm *ScoreManager) IsNewHighScore() bool {
	return sm.CurrentScore >= sm.HighScore && sm.CurrentScore > 0
}

// getSavePath returns the path to the high score file
func (sm *ScoreManager) getSavePath() string {
	if sm.savePath != "" {
		return sm.savePath
	}
	// Default to user home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return ".retro_maze_highscore.json"
	}
	return filepath.Join(home, ".retro_maze_highscore.json")
}

// loadHighScore loads the high score from disk
func (sm *ScoreManager) loadHighScore() {
	path := sm.getSavePath()
	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist or can't be read - that's fine
		return
	}

	var sd scoreData
	if err := json.Unmarshal(data, &sd); err != nil {
		// Corrupted file - ignore
		return
	}

	sm.HighScore = sd.HighScore
}

// saveHighScore saves the high score to disk
func (sm *ScoreManager) saveHighScore() {
	path := sm.getSavePath()
	sd := scoreData{HighScore: sm.HighScore}

	data, err := json.Marshal(sd)
	if err != nil {
		return
	}

	// Write file with proper permissions
	_ = os.WriteFile(path, data, 0644)
}
