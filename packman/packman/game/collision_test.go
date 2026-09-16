package game

import (
	"testing"
)

func TestPlayerEnemyCollision(t *testing.T) {
	cd := NewCollisionDetector()

	playerPos := Position{X: 5, Y: 5}
	enemyPos := Position{X: 5, Y: 5}

	if !cd.PlayerEnemyCollision(playerPos, enemyPos) {
		t.Error("Expected collision when positions are equal")
	}

	enemyPos = Position{X: 6, Y: 5}
	if cd.PlayerEnemyCollision(playerPos, enemyPos) {
		t.Error("Expected no collision when positions differ")
	}
}

func TestPlayerEnemyCollisionDistance(t *testing.T) {
	cd := NewCollisionDetector()

	playerPos := Position{X: 0, Y: 0}

	// Test exact distance
	enemyPos := Position{X: 3, Y: 4}
	if !cd.PlayerEnemyCollisionDistance(playerPos, enemyPos, 5.0) {
		t.Error("Expected collision within distance 5")
	}

	// Test outside distance
	if cd.PlayerEnemyCollisionDistance(playerPos, enemyPos, 4.9) {
		t.Error("Expected no collision outside distance 4.9")
	}

	// Test at same position
	enemyPos = Position{X: 0, Y: 0}
	if !cd.PlayerEnemyCollisionDistance(playerPos, enemyPos, 0.5) {
		t.Error("Expected collision at same position")
	}
}

func TestIsValidMove(t *testing.T) {
	cd := NewCollisionDetector()
	layout := []string{
		"###",
		"#.#",
		"###",
	}
	maze := NewMaze(layout)

	// Valid move
	if !cd.IsValidMove(maze, 1, 1) {
		t.Error("Expected (1,1) to be a valid move")
	}

	// Wall - invalid
	if cd.IsValidMove(maze, 0, 0) {
		t.Error("Expected (0,0) to be an invalid move (wall)")
	}

	// Out of bounds - invalid
	if cd.IsValidMove(maze, -1, 0) {
		t.Error("Expected (-1,0) to be an invalid move (out of bounds)")
	}
}

func TestIsSamePosition(t *testing.T) {
	cd := NewCollisionDetector()

	a := Position{X: 3, Y: 4}
	b := Position{X: 3, Y: 4}
	c := Position{X: 4, Y: 3}

	if !cd.IsSamePosition(a, b) {
		t.Error("Expected positions to be the same")
	}

	if cd.IsSamePosition(a, c) {
		t.Error("Expected positions to be different")
	}
}
