package ai

import (
	"bagh-bakri/internal/game"
	"testing"
)

func TestRandomAI_choosesLegalMove(t *testing.T) {
	ai := NewRandomAI()
	state := game.NewGame()
	move := ai.ChooseMove(state)
	moves := game.LegalMoves(state)
	if len(moves) == 0 {
		t.Fatal("no legal moves available")
	}
	found := false
	for _, m := range moves {
		if m == move {
			found = true
			break
		}
	}
	if !found {
		t.Error("random AI chose illegal move")
	}
}

func TestMinimaxAI_choosesLegalMove(t *testing.T) {
	ai := NewMinimaxAI(3)
	state := game.NewGame()
	move := ai.ChooseMove(state)
	moves := game.LegalMoves(state)
	if len(moves) == 0 {
		t.Fatal("no legal moves available")
	}
	found := false
	for _, m := range moves {
		if m == move {
			found = true
			break
		}
	}
	if !found {
		t.Error("minimax AI chose illegal move")
	}
}

func TestAI_doesNotMutateState(t *testing.T) {
	ai := NewMinimaxAI(2)
	state := game.NewGame()
	originalBoard := state.Board
	ai.ChooseMove(state)
	if state.Board != originalBoard {
		t.Error("AI mutated original state")
	}
}
