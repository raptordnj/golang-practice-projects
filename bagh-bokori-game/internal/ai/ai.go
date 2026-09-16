package ai

import (
	"bagh-bakri/internal/game"
	"math/rand"
)

type Player interface {
	ChooseMove(state game.GameState) game.Move
}

type HumanPlayer struct{}

func (h HumanPlayer) ChooseMove(state game.GameState) game.Move {
	return game.Move{}
}

type RandomAI struct {
	rng *rand.Rand
}

func NewRandomAI() *RandomAI {
	return &RandomAI{rng: rand.New(rand.NewSource(0))}
}

func (r *RandomAI) ChooseMove(state game.GameState) game.Move {
	moves := game.LegalMoves(state)
	if len(moves) == 0 {
		return game.Move{}
	}
	return moves[r.rng.Intn(len(moves))]
}
