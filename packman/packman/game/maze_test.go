package game

import (
	"testing"
)

func TestNewMaze(t *testing.T) {
	layout := []string{
		"####",
		"#..#",
		"#.o#",
		"####",
	}

	maze := NewMaze(layout)

	if maze.Width != 4 {
		t.Errorf("Expected width 4, got %d", maze.Width)
	}

	if maze.Height != 4 {
		t.Errorf("Expected height 4, got %d", maze.Height)
	}

	// Count pellets: 3 regular + 1 power = 4
	expectedPellets := 4
	if maze.Pellets != expectedPellets {
		t.Errorf("Expected %d pellets, got %d", expectedPellets, maze.Pellets)
	}
}

func TestIsWall(t *testing.T) {
	layout := []string{
		"###",
		"#.#",
		"###",
	}

	maze := NewMaze(layout)

	// Test walls
	if !maze.IsWall(0, 0) {
		t.Error("Expected (0,0) to be a wall")
	}

	if !maze.IsWall(1, 0) {
		t.Error("Expected (1,0) to be a wall")
	}

	// Test non-wall
	if maze.IsWall(1, 1) {
		t.Error("Expected (1,1) to not be a wall")
	}

	// Test out of bounds
	if !maze.IsWall(-1, 0) {
		t.Error("Expected (-1,0) to be a wall (out of bounds)")
	}

	if !maze.IsWall(100, 100) {
		t.Error("Expected (100,100) to be a wall (out of bounds)")
	}
}

func TestIsWalkable(t *testing.T) {
	layout := []string{
		"###",
		"#.#",
		"###",
	}

	maze := NewMaze(layout)

	// Test walkable
	if !maze.IsWalkable(1, 1) {
		t.Error("Expected (1,1) to be walkable")
	}

	// Test not walkable (wall)
	if maze.IsWalkable(0, 0) {
		t.Error("Expected (0,0) to not be walkable")
	}

	// Test out of bounds
	if maze.IsWalkable(-1, 0) {
		t.Error("Expected (-1,0) to not be walkable")
	}
}

func TestCollectPellet(t *testing.T) {
	layout := []string{
		"###",
		"#.#", // pellet at (1,1)
		"###",
	}

	maze := NewMaze(layout)

	// Collect pellet
	points := maze.CollectPellet(1, 1)
	if points != 10 {
		t.Errorf("Expected 10 points for regular pellet, got %d", points)
	}

	if maze.Pellets != 0 {
		t.Errorf("Expected 0 pellets after collection, got %d", maze.Pellets)
	}

	// Try to collect again (should be 0 points)
	points = maze.CollectPellet(1, 1)
	if points != 0 {
		t.Errorf("Expected 0 points for empty cell, got %d", points)
	}
}

func TestCollectPowerPellet(t *testing.T) {
	layout := []string{
		"###",
		"#o#", // power pellet at (1,1)
		"###",
	}

	maze := NewMaze(layout)

	// Collect power pellet
	points := maze.CollectPellet(1, 1)
	if points != 50 {
		t.Errorf("Expected 50 points for power pellet, got %d", points)
	}
}

func TestHasPellets(t *testing.T) {
	layout := []string{
		"###",
		"#.#",
		"###",
	}

	maze := NewMaze(layout)

	if !maze.HasPellets() {
		t.Error("Expected maze to have pellets")
	}

	// Collect all pellets
	maze.CollectPellet(1, 1)

	if maze.HasPellets() {
		t.Error("Expected maze to not have pellets")
	}
}

func TestFindPlayerStart(t *testing.T) {
	layout := []string{
		"#####",
		"#P..#",
		"#####",
	}

	maze := NewMaze(layout)

	x, y := maze.FindPlayerStart()
	if x != 1 || y != 1 {
		t.Errorf("Expected player start at (1,1), got (%d,%d)", x, y)
	}
}

func TestFindEnemyStarts(t *testing.T) {
	layout := []string{
		"#######",
		"#P...E#",
		"#.###.#",
		"#E...E#",
		"#######",
	}

	maze := NewMaze(layout)

	starts := maze.FindEnemyStarts()
	if len(starts) != 3 {
		t.Errorf("Expected 3 enemy starts, got %d", len(starts))
	}
}

func TestGetTile(t *testing.T) {
	layout := []string{
		"###",
		"#.#",
		"###",
	}

	maze := NewMaze(layout)

	tile := maze.GetTile(1, 1)
	if tile != TilePellet {
		t.Errorf("Expected TilePellet, got %v", tile)
	}

	tile = maze.GetTile(0, 0)
	if tile != TileWall {
		t.Errorf("Expected TileWall, got %v", tile)
	}
}

func TestSetTile(t *testing.T) {
	layout := []string{
		"###",
		"#.#",
		"###",
	}

	maze := NewMaze(layout)

	maze.SetTile(1, 1, TileEmpty)
	if maze.GetTile(1, 1) != TileEmpty {
		t.Error("Expected tile to be set to TileEmpty")
	}

	// Out of bounds should not panic
	maze.SetTile(100, 100, TileWall)
}
