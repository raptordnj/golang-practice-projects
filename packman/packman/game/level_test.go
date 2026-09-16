package game

import (
	"testing"
)

func TestNewLevelManager(t *testing.T) {
	lm := NewLevelManager()

	if lm.GetCurrentLevel() != 1 {
		t.Errorf("Expected level 1, got %d", lm.GetCurrentLevel())
	}
}

func TestNextLevel(t *testing.T) {
	lm := NewLevelManager()

	lm.NextLevel()
	if lm.GetCurrentLevel() != 2 {
		t.Errorf("Expected level 2, got %d", lm.GetCurrentLevel())
	}

	lm.NextLevel()
	if lm.GetCurrentLevel() != 3 {
		t.Errorf("Expected level 3, got %d", lm.GetCurrentLevel())
	}
}

func TestReset(t *testing.T) {
	lm := NewLevelManager()

	lm.NextLevel()
	lm.NextLevel()
	lm.Reset()

	if lm.GetCurrentLevel() != 1 {
		t.Errorf("Expected level 1 after reset, got %d", lm.GetCurrentLevel())
	}
}

func TestGetEnemySpeedMultiplier(t *testing.T) {
	lm := NewLevelManager()

	// Level 1: 1.0
	if mult := lm.GetEnemySpeedMultiplier(); mult != 1.0 {
		t.Errorf("Expected speed multiplier 1.0 at level 1, got %f", mult)
	}

	// Level 2: 1.1
	lm.NextLevel()
	if mult := lm.GetEnemySpeedMultiplier(); mult != 1.1 {
		t.Errorf("Expected speed multiplier 1.1 at level 2, got %f", mult)
	}

	// Level 11: should cap at 2.0
	for i := 0; i < 10; i++ {
		lm.NextLevel()
	}
	if mult := lm.GetEnemySpeedMultiplier(); mult != 2.0 {
		t.Errorf("Expected speed multiplier capped at 2.0, got %f", mult)
	}
}

func TestGetEnemyCount(t *testing.T) {
	lm := NewLevelManager()

	// Level 1: 2 enemies
	if count := lm.GetEnemyCount(); count != 2 {
		t.Errorf("Expected 2 enemies at level 1, got %d", count)
	}

	// Level 3: 3 enemies
	lm.NextLevel()
	lm.NextLevel()
	if count := lm.GetEnemyCount(); count != 3 {
		t.Errorf("Expected 3 enemies at level 3, got %d", count)
	}

	// Level 11+: should cap at 6
	for i := 0; i < 10; i++ {
		lm.NextLevel()
	}
	if count := lm.GetEnemyCount(); count != 6 {
		t.Errorf("Expected 6 enemies at level 11+, got %d", count)
	}
}

func TestGetPowerModeDuration(t *testing.T) {
	lm := NewLevelManager()

	// Level 1: 8 seconds
	if dur := lm.GetPowerModeDuration(); dur != 8.0 {
		t.Errorf("Expected power mode duration 8.0 at level 1, got %f", dur)
	}

	// Level 2: 7.5 seconds
	lm.NextLevel()
	if dur := lm.GetPowerModeDuration(); dur != 7.5 {
		t.Errorf("Expected power mode duration 7.5 at level 2, got %f", dur)
	}

	// Level 11+: should be at minimum 3.0
	for i := 0; i < 12; i++ {
		lm.NextLevel()
	}
	if dur := lm.GetPowerModeDuration(); dur < 3.0 {
		t.Errorf("Expected power mode duration >= 3.0, got %f", dur)
	}
}

func TestGetMazeLayout(t *testing.T) {
	lm := NewLevelManager()

	layout := lm.GetMazeLayout()
	if len(layout) == 0 {
		t.Error("Expected non-empty maze layout")
	}

	// Verify layout contains expected elements
	foundWall := false
	foundPellet := false
	for _, row := range layout {
		for _, ch := range row {
			if ch == '#' {
				foundWall = true
			}
			if ch == '.' {
				foundPellet = true
			}
		}
	}

	if !foundWall {
		t.Error("Expected maze layout to contain walls")
	}
	if !foundPellet {
		t.Error("Expected maze layout to contain pellets")
	}
}
