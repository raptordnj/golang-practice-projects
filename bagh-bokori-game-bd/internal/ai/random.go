package ai

import (
	"context"
	"math/rand"

	"bagh-bandi/internal/game"
)

// RandomAI picks uniformly among the legal moves. It is the Easy opponent and
// doubles as a baseline for testing stronger engines.
type RandomAI struct {
	rng *rand.Rand
}

// NewRandomAI returns a random player. Seeding it makes games reproducible,
// which is what the tests rely on.
func NewRandomAI(seed int64) *RandomAI {
	return &RandomAI{rng: rand.New(rand.NewSource(seed))}
}

// Name implements Player.
func (a *RandomAI) Name() string { return "Random" }

// ChooseMove implements Player.
func (a *RandomAI) ChooseMove(ctx context.Context, rules *game.TraditionalBaghBandiRules, state game.GameState) (game.Move, bool) {
	moves := rules.LegalMoves(state)
	if len(moves) == 0 {
		return game.Move{}, false
	}
	return moves[a.rng.Intn(len(moves))], true
}

var _ Player = (*RandomAI)(nil)
