package game

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	g := NewGame()

	if g.State != StateTitle {
		t.Errorf("Expected initial state StateTitle, got %v", g.State)
	}

	if g.Player == nil {
		t.Error("Expected player to be initialized")
	}

	if g.Player.Lives != 3 {
		t.Errorf("Expected 3 lives, got %d", g.Player.Lives)
	}

	if g.Maze == nil {
		t.Error("Expected maze to be initialized")
	}
	// Note: Enemies may be empty if no enemy start positions exist in default maze
}

func TestInitialize(t *testing.T) {
	g := NewGame()
	g.Initialize()

	if g.State != StateTitle {
		t.Errorf("Expected state StateTitle after Initialize, got %v", g.State)
	}

	if g.ScoreManager.GetCurrent() != 0 {
		t.Error("Expected score to be reset")
	}

	if g.LevelManager.GetCurrentLevel() != 1 {
		t.Error("Expected level to be reset to 1")
	}
}

func TestStart(t *testing.T) {
	g := NewGame()
	g.Start()

	if g.State != StateReady {
		t.Errorf("Expected state StateReady after Start, got %v", g.State)
	}
}

func TestGameStates(t *testing.T) {
	// Test state transitions
	if StateTitle.String() != "TITLE" {
		t.Errorf("Expected StateTitle string 'TITLE', got '%s'", StateTitle.String())
	}
	if StatePlaying.String() != "PLAYING" {
		t.Errorf("Expected StatePlaying string 'PLAYING', got '%s'", StatePlaying.String())
	}
	if StatePaused.String() != "PAUSED" {
		t.Errorf("Expected StatePaused string 'PAUSED', got '%s'", StatePaused.String())
	}
	if StateGameOver.String() != "GAME_OVER" {
		t.Errorf("Expected StateGameOver string 'GAME_OVER', got '%s'", StateGameOver.String())
	}
}

func TestDirection(t *testing.T) {
	tests := []struct {
		dir      Direction
		expected string
	}{
		{DirNone, "NONE"},
		{DirUp, "UP"},
		{DirDown, "DOWN"},
		{DirLeft, "LEFT"},
		{DirRight, "RIGHT"},
	}

	for _, test := range tests {
		if test.dir.String() != test.expected {
			t.Errorf("Expected Direction %d string '%s', got '%s'", test.dir, test.expected, test.dir.String())
		}
	}
}

func TestMovePlayer(t *testing.T) {
	g := NewGame()
	g.Start()
	g.State = StatePlaying

	// Get player starting position
	startX := g.Player.X
	startY := g.Player.Y

	// Move player (should not move into wall)
	g.MovePlayer(g.Player.Direction)

	// Verify player is still within bounds
	if g.Player.X < 0 || g.Player.X >= g.Maze.Width {
		t.Error("Player X position out of bounds")
	}
	if g.Player.Y < 0 || g.Player.Y >= g.Maze.Height {
		t.Error("Player Y position out of bounds")
	}

	// Verify player didn't move into a wall
	if g.Maze.IsWall(g.Player.X, g.Player.Y) {
		t.Error("Player moved into a wall")
	}

	// Suppress unused variable warnings
	_ = startX
	_ = startY
}

func TestPauseAndResume(t *testing.T) {
	g := NewGame()
	g.Start()
	g.State = StatePlaying

	g.Pause()
	if g.State != StatePaused {
		t.Errorf("Expected state StatePaused, got %v", g.State)
	}

	g.Resume()
	if g.State != StatePlaying {
		t.Errorf("Expected state StatePlaying, got %v", g.State)
	}
}

func TestTogglePause(t *testing.T) {
	g := NewGame()
	g.Start()
	g.State = StatePlaying

	g.TogglePause()
	if g.State != StatePaused {
		t.Errorf("Expected state StatePaused, got %v", g.State)
	}

	g.TogglePause()
	if g.State != StatePlaying {
		t.Errorf("Expected state StatePlaying, got %v", g.State)
	}
}

func TestRestart(t *testing.T) {
	g := NewGame()
	g.Start()
	g.State = StatePlaying

	// Modify some state
	g.Player.Lives = 1
	g.ScoreManager.AddScore(100)
	g.LevelManager.NextLevel()

	// Restart
	g.Restart()

	if g.Player.Lives != 3 {
		t.Errorf("Expected 3 lives after restart, got %d", g.Player.Lives)
	}

	if g.LevelManager.GetCurrentLevel() != 1 {
		t.Errorf("Expected level 1 after restart, got %d", g.LevelManager.GetCurrentLevel())
	}

	if g.State != StateReady {
		t.Errorf("Expected state StateReady after restart, got %v", g.State)
	}
}

func TestKillPlayer(t *testing.T) {
	g := NewGame()
	g.Start()
	g.State = StatePlaying

	initialLives := g.Player.Lives
	g.killPlayer()

	if g.Player.Lives != initialLives-1 {
		t.Errorf("Expected %d lives after death, got %d", initialLives-1, g.Player.Lives)
	}

	if g.Player.IsAlive {
		t.Error("Expected player to be dead")
	}
}

func TestGameStateChangeCallback(t *testing.T) {
	g := NewGame()

	var oldState, newState GameState
	g.SetStateChangeCallback(func(old, new GameState) {
		oldState = old
		newState = new
	})

	g.Start()

	if oldState != StateTitle {
		t.Errorf("Expected oldState StateTitle, got %v", oldState)
	}

	if newState != StateReady {
		t.Errorf("Expected newState StateReady, got %v", newState)
	}
}
