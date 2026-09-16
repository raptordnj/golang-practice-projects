// Package ai provides computer opponents for Bagh-Bandi.
//
// Everything here talks only to internal/game. It imports no UI toolkit, no
// OS-specific package and touches no filesystem, so it compiles for desktop,
// WebAssembly and mobile alike, and it can be exercised entirely from tests.
package ai

import (
	"context"

	"bagh-bandi/internal/game"
)

// Player chooses a move for the side to move.
//
// Implementations must never mutate the state they are given and must always
// return a move that the rules accept. ChooseMove may be called on a
// background goroutine, so it must not touch UI state.
type Player interface {
	// Name is a short human-readable label, e.g. "Minimax (depth 4)".
	Name() string
	// ChooseMove returns a legal move, or ok == false if there is none.
	// The context lets a caller abandon a search that is taking too long;
	// the best move found so far is returned in that case.
	ChooseMove(ctx context.Context, rules *game.TraditionalBaghBandiRules, state game.GameState) (move game.Move, ok bool)
}

// Difficulty selects a preset opponent strength.
type Difficulty int

const (
	// Easy plays at random.
	Easy Difficulty = iota
	// Medium searches a few plies ahead.
	Medium
	// Hard searches deeply.
	Hard
)

func (d Difficulty) String() string {
	switch d {
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	default:
		return "Easy"
	}
}

// Bengali returns the Bangla label for the difficulty, used by the UI.
func (d Difficulty) Bengali() string {
	switch d {
	case Medium:
		return "মাঝারি"
	case Hard:
		return "কঠিন"
	default:
		return "সহজ"
	}
}

// NewPlayer builds the opponent for a difficulty preset.
//
// Easy   -> random legal move
// Medium -> minimax with alpha-beta, depth 4
// Hard   -> minimax with alpha-beta, depth 6
func NewPlayer(d Difficulty, seed int64) Player {
	switch d {
	case Medium:
		return NewMinimaxAI(4, seed)
	case Hard:
		return NewMinimaxAI(6, seed)
	default:
		return NewRandomAI(seed)
	}
}
