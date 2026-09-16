package game

import (
	"path/filepath"
	"testing"
)

func TestNewScoreManager(t *testing.T) {
	// Use a temp file to avoid loading previous high score
	tmpPath := filepath.Join(t.TempDir(), "test_highscore.json")
	sm := NewScoreManagerWithPath(tmpPath)

	if sm.GetCurrent() != 0 {
		t.Errorf("Expected current score 0, got %d", sm.GetCurrent())
	}

	if sm.GetHighScore() != 0 {
		t.Errorf("Expected high score 0, got %d", sm.GetHighScore())
	}
}

func TestNewScoreManagerWithPath(t *testing.T) {
	// Create a temp file path
	tmpPath := filepath.Join(t.TempDir(), "test_highscore.json")
	sm := NewScoreManagerWithPath(tmpPath)

	if sm.GetCurrent() != 0 {
		t.Errorf("Expected current score 0, got %d", sm.GetCurrent())
	}
}

func TestAddScore(t *testing.T) {
	// Use a temp file to avoid loading previous high score
	tmpPath := filepath.Join(t.TempDir(), "test_score.json")
	sm := NewScoreManagerWithPath(tmpPath)

	sm.AddScore(100)
	if sm.GetCurrent() != 100 {
		t.Errorf("Expected current score 100, got %d", sm.GetCurrent())
	}
	if sm.GetHighScore() != 100 {
		t.Errorf("Expected high score 100, got %d", sm.GetHighScore())
	}

	sm.AddScore(50)
	if sm.GetCurrent() != 150 {
		t.Errorf("Expected current score 150, got %d", sm.GetCurrent())
	}
	if sm.GetHighScore() != 150 {
		t.Errorf("Expected high score 150, got %d", sm.GetHighScore())
	}
}

func TestResetCurrent(t *testing.T) {
	// Use a temp file to avoid loading previous high score
	tmpPath := filepath.Join(t.TempDir(), "test_reset.json")
	sm := NewScoreManagerWithPath(tmpPath)

	sm.AddScore(100)
	sm.ResetCurrent()

	if sm.GetCurrent() != 0 {
		t.Errorf("Expected current score 0 after reset, got %d", sm.GetCurrent())
	}

	// High score should not be reset
	if sm.GetHighScore() != 100 {
		t.Errorf("Expected high score 100 (not reset), got %d", sm.GetHighScore())
	}
}

func TestIsNewHighScore(t *testing.T) {
	// Use a temp file to avoid loading previous high score
	tmpPath := filepath.Join(t.TempDir(), "test_new_high.json")
	sm := NewScoreManagerWithPath(tmpPath)

	// Initially, current score is 0, not a new high score
	if sm.IsNewHighScore() {
		t.Error("Expected 0 score to not be new high score")
	}

	sm.AddScore(100)
	sm.ResetCurrent()

	// After reset, score is 0, which is equal to high score (100), but score is 0
	if sm.IsNewHighScore() {
		t.Error("Expected 0 score to not be new high score")
	}

	sm.AddScore(150)
	if !sm.IsNewHighScore() {
		t.Error("Expected 150 to be new high score")
	}
}

func TestScorePersistence(t *testing.T) {
	// Use a temp file
	tmpPath := filepath.Join(t.TempDir(), "test_persistence.json")
	sm := NewScoreManagerWithPath(tmpPath)

	// Add score
	sm.AddScore(1234)

	// Create new manager with same path
	sm2 := NewScoreManagerWithPath(tmpPath)

	if sm2.GetHighScore() != 1234 {
		t.Errorf("Expected persisted high score 1234, got %d", sm2.GetHighScore())
	}
}
