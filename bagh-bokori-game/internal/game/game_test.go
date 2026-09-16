package game

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	state := NewGame()
	if state.Turn != PlayerTiger {
		t.Error("expected first turn to be Tiger")
	}
	if len(state.Tigers) != 4 {
		t.Errorf("expected 4 tigers, got %d", len(state.Tigers))
	}
	if state.GoatsPlaced != 0 {
		t.Errorf("expected 0 goats placed, got %d", state.GoatsPlaced)
	}
	if state.Status != StatusPlaying {
		t.Error("expected game to be playing")
	}
}

func TestBoardConnectivity(t *testing.T) {
	board := NewBoard()
	neighbors := board.Neighbors(0)
	expected := []PositionID{1, 5}
	if len(neighbors) != len(expected) {
		t.Errorf("position 0 neighbors: expected %v, got %v", expected, neighbors)
	}
	// Check center connectivity
	centerNeighbors := board.Neighbors(6)
	expectedCenter := []PositionID{1, 5, 7, 11}
	if len(centerNeighbors) != len(expectedCenter) {
		t.Errorf("position 6 neighbors: expected %v, got %v", expectedCenter, centerNeighbors)
	}
}

func TestValidNeighbors(t *testing.T) {
	board := NewBoard()
	neighbors := board.Neighbors(0)
	for _, n := range neighbors {
		if n != 1 && n != 5 {
			t.Errorf("unexpected neighbor for position 0: %d", n)
		}
	}
}

func TestCapturePaths(t *testing.T) {
	state := NewGame()
	// Place goat at 1 between tiger at 0 and empty 2
	state.Board = state.Board.Set(1, Piece{Type: Goat, Player: PlayerGoat})
	state.Goats = append(state.Goats, 1)
	state.Turn = PlayerTiger

	// Tiger at 0 should be able to jump over goat at 1 to position 2
	moves := LegalMoves(state)
	found := false
	for _, m := range moves {
		if m.From == 0 && m.To == 2 {
			found = true
		}
	}
	if !found {
		t.Error("expected capture path from 0 over 1 to 2")
	}
}

func TestInvalidPaths(t *testing.T) {
	state := NewGame()
	// Tiger at 0 cannot move to 2 directly (not adjacent)
	moves := LegalMoves(state)
	for _, m := range moves {
		if m.From == 0 && m.To == 2 {
			t.Error("position 2 is not adjacent to 0")
		}
	}
}

func TestTigerMovement(t *testing.T) {
	state := NewGame()
	moves := LegalMoves(state)
	found := false
	for _, m := range moves {
		if m.From == 0 && m.To == 1 {
			found = true
		}
	}
	if !found {
		t.Error("expected tiger at 0 to be able to move to 1")
	}
}

func TestInvalidTigerMovement(t *testing.T) {
	state := NewGame()
	// Tiger at 0 cannot move to 4 (not adjacent)
	moves := LegalMoves(state)
	for _, m := range moves {
		if m.From == 0 && m.To == 4 {
			t.Error("tiger at 0 should not be able to move to 4")
		}
	}
}

func TestValidTigerCapture(t *testing.T) {
	state := NewGame()
	state.Board = state.Board.Set(1, Piece{Type: Goat, Player: PlayerGoat})
	state.Goats = append(state.Goats, 1)
	state.Turn = PlayerTiger

	moves := LegalMoves(state)
	found := false
	for _, m := range moves {
		if m.From == 0 && m.To == 2 {
			found = true
		}
	}
	if !found {
		t.Error("expected valid tiger capture from 0 over 1 to 2")
	}
}

func TestInvalidCapture(t *testing.T) {
	state := NewGame()
	state.Board = state.Board.Set(1, Piece{Type: Goat, Player: PlayerGoat})
	state.Board = state.Board.Set(6, Piece{Type: Tiger, Player: PlayerTiger})
	state.Goats = append(state.Goats, 1)
	state.Turn = PlayerTiger

	// Position 6 is occupied by tiger, so no capture to 6
	moves := LegalMoves(state)
	for _, m := range moves {
		if m.From == 0 && m.To == 6 {
			t.Error("should not capture to position 6 (occupied by tiger)")
		}
	}
}

func TestGoatPlacement(t *testing.T) {
	state := NewGame()
	state.Turn = PlayerGoat
	moves := LegalMoves(state)
	if len(moves) != 21 {
		t.Errorf("expected 21 placement moves, got %d", len(moves))
	}
}

func TestGoatMovement(t *testing.T) {
	state := NewGame()
	state.GoatsPlaced = 20
	state.Turn = PlayerGoat
	// Place a goat that can move
	state.Board = state.Board.Set(7, Piece{Type: Goat, Player: PlayerGoat})
	state.Goats = append(state.Goats, 7)
	moves := LegalMoves(state)
	found := false
	for _, m := range moves {
		if m.From == 7 {
			found = true
		}
	}
	if !found {
		t.Error("expected goat at 7 to be movable")
	}
}

func TestTurnSwitching(t *testing.T) {
	state := NewGame()
	newState, _ := ApplyMove(state, Move{From: 0, To: 1})
	if newState.Turn != PlayerGoat {
		t.Error("expected turn to switch to goat after tiger move")
	}
	newState2, _ := ApplyMove(newState, Move{From: PositionID(-1), To: 6})
	if newState2.Turn != PlayerTiger {
		t.Error("expected turn to switch to tiger after goat move")
	}
}

func TestGoatCount(t *testing.T) {
	state := NewGame()
	state.Turn = PlayerGoat
	state, _ = ApplyMove(state, Move{From: PositionID(-1), To: 6})
	state.Turn = PlayerGoat
	state, _ = ApplyMove(state, Move{From: PositionID(-1), To: 7})
	state.Turn = PlayerGoat
	state, _ = ApplyMove(state, Move{From: PositionID(-1), To: 8})
	if state.GoatsPlaced != 3 {
		t.Errorf("expected 3 goats placed, got %d", state.GoatsPlaced)
	}
}

func TestCapturedGoatCount(t *testing.T) {
	state := NewGame()
	state.Board = state.Board.Set(1, Piece{Type: Goat, Player: PlayerGoat})
	state.Board = state.Board.Set(2, Piece{})
	state.Goats = append(state.Goats, 1)
	state.Turn = PlayerTiger

	moves := LegalMoves(state)
	found := false
	for _, m := range moves {
		if m.From == 0 && m.To == 2 {
			found = true
		}
	}
	if !found {
		t.Fatal("expected tiger capture move")
	}
	state, _ = ApplyMove(state, Move{From: 0, To: 2})
	if state.CapturedGoats != 1 {
		t.Errorf("expected 1 captured goat, got %d", state.CapturedGoats)
	}
}

func TestGameOverDetection(t *testing.T) {
	state := NewGame()
	state.Status = StatusTigerWin
	if !IsGameOver(state) {
		t.Error("tiger win should be game over")
	}
	state.Status = StatusGoatWin
	if !IsGameOver(state) {
		t.Error("goat win should be game over")
	}
}

func TestWinnerDetection(t *testing.T) {
	state := NewGame()
	state.Status = StatusTigerWin
	w := Winner(state)
	if w.Winner != PlayerTiger {
		t.Error("expected tiger winner")
	}

	state.Status = StatusGoatWin
	w = Winner(state)
	if w.Winner != PlayerGoat {
		t.Error("expected goat winner")
	}
}

func TestGameOverApplication(t *testing.T) {
	state := NewGame()
	state.Turn = PlayerTiger
	state.Board = state.Board.Set(1, Piece{Type: Goat, Player: PlayerGoat})
	state.Board = state.Board.Set(5, Piece{Type: Goat, Player: PlayerGoat})
	state.Goats = append(state.Goats, 1, 5)
	if IsGameOver(state) {
		t.Error("game should not be over with tigers still mobile")
	}
}

func TestUndoMultiple(t *testing.T) {
	state := NewGame()
	state, _ = ApplyMove(state, Move{From: 0, To: 1})
	state, _ = ApplyMove(state, Move{From: PositionID(-1), To: 6})
	state = Undo(state)
	if state.Turn != PlayerGoat {
		t.Error("expected goat turn after undo of goat placement")
	}
	if state.GoatsPlaced != 0 {
		t.Error("expected 0 goats placed after undo")
	}
}

func TestIsValidMove(t *testing.T) {
	state := NewGame()
	if !IsValidMove(state, Move{From: 0, To: 1}) {
		t.Error("expected move 0->1 to be valid")
	}
	if IsValidMove(state, Move{From: 0, To: 4}) {
		t.Error("expected move 0->4 to be invalid")
	}
}

func TestNoMovesWhenGameOver(t *testing.T) {
	state := NewGame()
	state.Status = StatusTigerWin
	moves := LegalMoves(state)
	if len(moves) != 0 {
		t.Error("expected no moves when game is over")
	}
}

// Benchmarks

func BenchmarkLegalMoves(b *testing.B) {
	state := NewGame()
	for i := 0; i < 10; i++ {
		state.Turn = PlayerGoat
		state.GoatsPlaced = 20
		state, _ = ApplyMove(state, Move{From: PositionID(-1), To: 7})
		state.Turn = PlayerTiger
		state, _ = ApplyMove(state, Move{From: 0, To: 1})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LegalMoves(state)
	}
}

func BenchmarkApplyMove(b *testing.B) {
	state := NewGame()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st := state
		st, _ = ApplyMove(st, Move{From: 0, To: 1})
		_ = st
	}
}
