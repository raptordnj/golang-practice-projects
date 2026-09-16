package ai

import (
	"bagh-bakri/internal/game"
	"math"
)

type MinimaxAI struct {
	Depth int
}

func NewMinimaxAI(depth int) *MinimaxAI {
	if depth < 1 {
		depth = 1
	}
	return &MinimaxAI{Depth: depth}
}

func (m *MinimaxAI) ChooseMove(state game.GameState) game.Move {
	moves := game.LegalMoves(state)
	if len(moves) == 0 {
		return game.Move{From: game.PositionID(-1), To: game.PositionID(-1)}
	}

	bestMove := moves[0]
	bestScore := math.Inf(-1)

	for _, move := range moves {
		if move.From == game.PositionID(-1) {
			continue
		}
		newState, err := game.ApplyMove(state, move)
		if err != nil {
			continue
		}
		score := minimax(newState, m.Depth-1, math.Inf(-1), math.Inf(1), false)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	return bestMove
}

func minimax(state game.GameState, depth int, alpha, beta float64, maximizing bool) float64 {
	if state.Status != game.StatusPlaying || depth == 0 {
		return evaluate(state)
	}

	moves := game.LegalMoves(state)
	if len(moves) == 0 {
		return evaluate(state)
	}

	if maximizing {
		maxEval := math.Inf(-1)
		for _, move := range moves {
			if move.From == game.PositionID(-1) {
				continue
			}
			newState, err := game.ApplyMove(state, move)
			if err != nil {
				continue
			}
			eval := minimax(newState, depth-1, alpha, beta, false)
			maxEval = math.Max(maxEval, eval)
			alpha = math.Max(alpha, eval)
			if beta <= alpha {
				break
			}
		}
		return maxEval
	} else {
		minEval := math.Inf(1)
		for _, move := range moves {
			if move.From == game.PositionID(-1) {
				continue
			}
			newState, err := game.ApplyMove(state, move)
			if err != nil {
				continue
			}
			eval := minimax(newState, depth-1, alpha, beta, true)
			minEval = math.Min(minEval, eval)
			beta = math.Min(beta, eval)
			if beta <= alpha {
				break
			}
		}
		return minEval
	}
}

func evaluate(state game.GameState) float64 {
	if state.Status == game.StatusTigerWin {
		return 1000
	}
	if state.Status == game.StatusGoatWin {
		return -1000
	}

	score := 0.0

	// Tiger advantages
	score += float64(state.CapturedGoats) * 10
	tigerMoves := len(tigerMovesOnly(state))
	score += float64(tigerMoves) * 2

	// Goat advantages
	goatsRemaining := 20 - state.CapturedGoats - len(state.Goats)
	score -= float64(goatsRemaining) * 3

	return score
}

func tigerMovesOnly(state game.GameState) []game.Move {
	oldTurn := state.Turn
	state.Turn = game.PlayerTiger
	moves := game.LegalMoves(state)
	state.Turn = oldTurn
	return moves
}
