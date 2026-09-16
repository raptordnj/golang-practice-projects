package game

// GameStatus is the outcome of a position.
type GameStatus uint8

const (
	// InProgress means the game is still being played.
	InProgress GameStatus = iota
	// TigerWon means the tiger captured enough goats.
	TigerWon
	// GoatsWon means the tiger has no legal move left.
	GoatsWon
	// Draw means the no-progress limit was reached.
	Draw
)

func (s GameStatus) String() string {
	switch s {
	case TigerWon:
		return "Tiger wins"
	case GoatsWon:
		return "Goats win"
	case Draw:
		return "Draw"
	default:
		return "In progress"
	}
}

// GameState is a complete, self-contained snapshot of a game.
//
// It is a pure value type: it contains no slices, maps or pointers into
// shared storage, so `next := state` produces a fully independent copy. That
// is what lets the AI search millions of nodes without ever touching the real
// game. The board topology itself is not stored here because it is constant;
// use NewBoard to obtain it.
type GameState struct {
	// Occupancy is indexed by PositionID.
	Occupancy [PositionCount]PieceType

	// Tiger is where the single বাঘ stands.
	Tiger PositionID

	// GoatsPlaced counts goats dropped so far (placement phase progress).
	GoatsPlaced int
	// GoatsCaptured counts goats eaten by the tiger.
	GoatsCaptured int

	// Turn is the side to move.
	Turn Player
	// Status is the current outcome.
	Status GameStatus

	// HalfmoveClock counts plies since the last capture or placement. It
	// drives the draw rule and prevents endless shuffling.
	HalfmoveClock int
}

// GoatsOnBoard returns the number of goats currently standing on the board.
func (s GameState) GoatsOnBoard() int {
	return s.GoatsPlaced - s.GoatsCaptured
}

// InPlacementPhase reports whether goats are still being dropped onto the
// board rather than moved.
func (s GameState) InPlacementPhase(r *RuleSet) bool {
	return s.GoatsPlaced < r.TotalGoats
}

// GoatsInHand returns how many goats are still waiting to be placed.
func (s GameState) GoatsInHand(r *RuleSet) int {
	return r.TotalGoats - s.GoatsPlaced
}

// Goats returns the positions of every goat on the board, in board order.
func (s GameState) Goats() []PositionID {
	out := make([]PositionID, 0, s.GoatsOnBoard())
	for id, p := range s.Occupancy {
		if p == Goat {
			out = append(out, PositionID(id))
		}
	}
	return out
}

// IsEmpty reports whether the point exists and holds no piece.
func (s GameState) IsEmpty(id PositionID) bool {
	return id >= 0 && int(id) < PositionCount && s.Occupancy[id] == Empty
}

// IsOver reports whether the game has finished.
func (s GameState) IsOver() bool { return s.Status != InProgress }
