package ai

import "bagh-bandi/internal/game"

// Score is a position evaluation in centi-goats, always from the tiger's
// point of view: positive favours the tiger, negative favours the goats.
type Score int32

const (
	// WinScore is the value of a won position. Mate scores are adjusted by
	// ply so the search prefers a quick win over a slow one.
	WinScore Score = 1_000_000
	// LossScore is the value of a lost position.
	LossScore Score = -WinScore
)

// Weights tunes the evaluation. Keeping the knobs in a struct means the
// evaluator can be retuned - or swapped for a learned one - without touching
// the search.
type Weights struct {
	// CapturedGoat is what each eaten goat is worth to the tiger.
	CapturedGoat Score
	// GoatInHand values goats still waiting to be placed slightly, since an
	// unplaced goat is a goat the tiger cannot eat yet.
	GoatInHand Score
	// TigerMobility rewards each empty point the tiger can step to.
	TigerMobility Score
	// CaptureThreat rewards each capture available right now.
	CaptureThreat Score
	// GoatMobility rewards the goats for keeping their own moves open.
	GoatMobility Score
	// TigerEdgePenalty pushes the tiger away from corners and edges, where
	// it is easiest to fence in.
	TigerEdgePenalty Score
	// GoatConnectivity rewards goats that defend each other: a goat with a
	// friendly goat behind it cannot be jumped.
	GoatConnectivity Score
}

// DefaultWeights are deliberately simple and readable. They are a starting
// point, not a tuned engine.
func DefaultWeights() Weights {
	return Weights{
		CapturedGoat:     1000,
		GoatInHand:       60,
		TigerMobility:    45,
		CaptureThreat:    120,
		GoatMobility:     8,
		TigerEdgePenalty: 35,
		GoatConnectivity: 25,
	}
}

// Evaluator scores positions for the search.
type Evaluator struct {
	W     Weights
	board *game.Board
}

// NewEvaluator returns an evaluator with the given weights.
func NewEvaluator(w Weights) *Evaluator {
	return &Evaluator{W: w, board: game.NewBoard()}
}

// Evaluate scores a position from the tiger's perspective.
func (e *Evaluator) Evaluate(rules *game.TraditionalBaghBandiRules, s *game.GameState, ply int) Score {
	switch s.Status {
	case game.TigerWon:
		return WinScore - Score(ply)
	case game.GoatsWon:
		return LossScore + Score(ply)
	case game.Draw:
		return 0
	}

	cfg := rules.Config()
	var score Score

	// Material: goats eaten, and goats the tiger has not had a shot at yet.
	score += Score(s.GoatsCaptured) * e.W.CapturedGoat
	score += Score(s.GoatsInHand(&cfg)) * e.W.GoatInHand

	// Tiger freedom. A tiger with nowhere to go is a dead tiger, so mobility
	// is the goats' main lever and the tiger's main asset.
	steps, captures := e.tigerOptions(s)
	score += Score(steps) * e.W.TigerMobility
	score += Score(captures) * e.W.CaptureThreat

	// Positional: the tiger is safest with many lines, i.e. near the centre.
	score -= Score(edgeDistancePenalty(s.Tiger)) * e.W.TigerEdgePenalty

	// Goat structure: mobility keeps the fence flexible, and mutual support
	// makes goats un-jumpable.
	goatMoves, supported := e.goatStructure(s)
	score -= Score(goatMoves) * e.W.GoatMobility
	score -= Score(supported) * e.W.GoatConnectivity

	return score
}

// tigerOptions counts the tiger's quiet steps and available captures.
func (e *Evaluator) tigerOptions(s *game.GameState) (steps, captures int) {
	for _, n := range e.board.Neighbors(s.Tiger) {
		if s.Occupancy[n] == game.Empty {
			steps++
		}
	}
	for _, j := range e.board.Jumps(s.Tiger) {
		if s.Occupancy[j.Over] == game.Goat && s.Occupancy[j.To] == game.Empty {
			captures++
		}
	}
	return steps, captures
}

// goatStructure counts the goats' available slides and how many goats stand
// on a point whose jump-landing square is blocked - those goats are safe.
func (e *Evaluator) goatStructure(s *game.GameState) (moves, supported int) {
	for id := game.PositionID(0); int(id) < game.PositionCount; id++ {
		if s.Occupancy[id] != game.Goat {
			continue
		}
		for _, n := range e.board.Neighbors(id) {
			if s.Occupancy[n] == game.Empty {
				moves++
			}
		}
		if e.isGoatSafe(s, id) {
			supported++
		}
	}
	return moves, supported
}

// isGoatSafe reports whether this goat can never be jumped from where it
// stands: for every line through it, either the tiger could not stand on the
// near side or the landing point beyond is blocked.
func (e *Evaluator) isGoatSafe(s *game.GameState, goat game.PositionID) bool {
	for _, n := range e.board.Neighbors(goat) {
		// The tiger attacks from n only if n is free for it to reach.
		if s.Occupancy[n] == game.Goat {
			continue
		}
		to, ok := e.jumpBeyond(n, goat)
		if ok && s.Occupancy[to] == game.Empty {
			return false
		}
	}
	return true
}

// jumpBeyond returns the point directly beyond `over` when travelling from
// `from`, if that whole path follows drawn lines.
func (e *Evaluator) jumpBeyond(from, over game.PositionID) (game.PositionID, bool) {
	for _, j := range e.board.Jumps(from) {
		if j.Over == over {
			return j.To, true
		}
	}
	return game.NoPosition, false
}

// edgeDistancePenalty grows as a point offers fewer lines: 0 at the centre,
// up to 2 in a corner.
func edgeDistancePenalty(id game.PositionID) int {
	row, col := game.RowCol(id)
	mid := game.BoardSize / 2
	return absInt(row-mid) + absInt(col-mid)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
