package ai

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"bagh-bandi/internal/game"
)

// MinimaxAI searches the game tree with alpha-beta pruning and iterative
// deepening. The tiger maximises the evaluation, the goats minimise it.
//
// The search is pure: it copies GameState values (which contain no
// references) and never touches the caller's position, so it is safe to run
// on a background goroutine while the UI redraws.
type MinimaxAI struct {
	// Depth is the maximum search depth in plies.
	Depth int

	eval *Evaluator
	rng  *rand.Rand

	// buffers[ply] is the move list for that depth, reused across the whole
	// search to keep it allocation-free.
	buffers [][]game.Move

	// Stats from the most recent search, for benchmarks and debugging.
	Nodes   int
	Reached int
}

// NewMinimaxAI returns a minimax player searching to the given depth.
// The seed makes tie-breaking between equally good moves reproducible.
func NewMinimaxAI(depth int, seed int64) *MinimaxAI {
	if depth < 1 {
		depth = 1
	}
	return &MinimaxAI{
		Depth: depth,
		eval:  NewEvaluator(DefaultWeights()),
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// SetWeights replaces the evaluation weights.
func (a *MinimaxAI) SetWeights(w Weights) { a.eval = NewEvaluator(w) }

// Name implements Player.
func (a *MinimaxAI) Name() string { return fmt.Sprintf("Minimax (depth %d)", a.Depth) }

// ChooseMove implements Player. It deepens one ply at a time so that an
// interrupted search still returns the best move from the last completed
// depth rather than nothing.
func (a *MinimaxAI) ChooseMove(ctx context.Context, rules *game.TraditionalBaghBandiRules, state game.GameState) (game.Move, bool) {
	moves := rules.LegalMoves(state)
	if len(moves) == 0 {
		return game.Move{}, false
	}
	a.Nodes = 0
	a.Reached = 0

	// Shuffle first so equally scored moves are not always the same one;
	// a deterministic engine that always plays a1 feels broken to a human.
	a.rng.Shuffle(len(moves), func(i, j int) { moves[i], moves[j] = moves[j], moves[i] })

	best := moves[0]
	for depth := 1; depth <= a.Depth; depth++ {
		move, ok := a.searchRoot(ctx, rules, state, moves, depth)
		if !ok {
			break // out of time: keep the previous depth's answer
		}
		best = move
		a.Reached = depth
		// Search the previous best first next time: better ordering means
		// far more alpha-beta cutoffs.
		promote(moves, best)
	}
	return best, true
}

func (a *MinimaxAI) searchRoot(ctx context.Context, rules *game.TraditionalBaghBandiRules, state game.GameState, moves []game.Move, depth int) (game.Move, bool) {
	maximising := state.Turn == game.TigerPlayer
	best := moves[0]
	bestScore := LossScore * 2
	if !maximising {
		bestScore = WinScore * 2
	}

	for _, m := range moves {
		if ctxDone(ctx) {
			return best, false
		}
		next, err := rules.ApplyMove(state, m)
		if err != nil {
			continue
		}
		score := a.search(ctx, rules, next, depth-1, 1, LossScore*2, WinScore*2)
		if maximising && score > bestScore || !maximising && score < bestScore {
			bestScore, best = score, m
		}
	}
	if ctxDone(ctx) {
		return best, false
	}
	return best, true
}

// search is the alpha-beta negamax-style recursion, written as an explicit
// min/max because the two sides have genuinely different move generators.
func (a *MinimaxAI) search(ctx context.Context, rules *game.TraditionalBaghBandiRules, state game.GameState, depth, ply int, alpha, beta Score) Score {
	a.Nodes++
	if state.IsOver() || depth == 0 {
		return a.eval.Evaluate(rules, &state, ply)
	}
	if a.Nodes&1023 == 0 && ctxDone(ctx) {
		return a.eval.Evaluate(rules, &state, ply)
	}

	moves := rules.AppendLegalMoves(a.buffer(ply)[:0], state)
	a.setBuffer(ply, moves)
	if len(moves) == 0 {
		// The rules already turn "no moves" into a finished status, so this
		// is defensive only.
		return a.eval.Evaluate(rules, &state, ply)
	}
	orderMoves(moves)

	if state.Turn == game.TigerPlayer {
		best := LossScore * 2
		for _, m := range moves {
			next, err := rules.ApplyMove(state, m)
			if err != nil {
				continue
			}
			score := a.search(ctx, rules, next, depth-1, ply+1, alpha, beta)
			if score > best {
				best = score
			}
			if best > alpha {
				alpha = best
			}
			if alpha >= beta {
				break // the goats would never allow this line
			}
		}
		return best
	}

	best := WinScore * 2
	for _, m := range moves {
		next, err := rules.ApplyMove(state, m)
		if err != nil {
			continue
		}
		score := a.search(ctx, rules, next, depth-1, ply+1, alpha, beta)
		if score < best {
			best = score
		}
		if best < beta {
			beta = best
		}
		if alpha >= beta {
			break // the tiger would never allow this line
		}
	}
	return best
}

// buffer returns the reusable move slice for a ply, growing the pool lazily.
func (a *MinimaxAI) buffer(ply int) []game.Move {
	for len(a.buffers) <= ply {
		a.buffers = append(a.buffers, make([]game.Move, 0, 32))
	}
	return a.buffers[ply]
}

func (a *MinimaxAI) setBuffer(ply int, moves []game.Move) { a.buffers[ply] = moves }

// orderMoves puts captures first. Captures change material, so they are the
// moves most likely to cause a cutoff.
func orderMoves(moves []game.Move) {
	sort.SliceStable(moves, func(i, j int) bool {
		return moves[i].Kind == game.Capture && moves[j].Kind != game.Capture
	})
}

// promote moves `m` to the front of the slice, preserving the rest's order.
func promote(moves []game.Move, m game.Move) {
	for i, got := range moves {
		if got == m {
			copy(moves[1:i+1], moves[:i])
			moves[0] = m
			return
		}
	}
}

func ctxDone(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

var _ Player = (*MinimaxAI)(nil)
