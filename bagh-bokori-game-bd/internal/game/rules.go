package game

// Winner names the side that won, if any.
type Winner uint8

const (
	// NoWinner means the game is unfinished or drawn.
	NoWinner Winner = iota
	// TigerWinner means the tiger ate enough goats.
	TigerWinner
	// GoatWinner means the goats fenced the tiger in.
	GoatWinner
)

func (w Winner) String() string {
	switch w {
	case TigerWinner:
		return "Tiger"
	case GoatWinner:
		return "Goats"
	default:
		return "None"
	}
}

// RuleSet holds every tunable rule of the game in one place. Nothing in the
// UI or the AI may reinvent these numbers.
//
// The defaults describe the Bangladeshi বাঘবন্দী variant this project
// implements: one tiger against sixteen goats on the 5x5 Alquerque board.
type RuleSet struct {
	// TotalGoats is how many goats the goat player owns (16).
	TotalGoats int

	// TigerStart is where the tiger begins (the centre point, c3).
	TigerStart PositionID

	// FirstPlayer is the side that moves first (goats, by placing).
	FirstPlayer Player

	// TigerWinCaptures is how many goats the tiger must eat to win. Below
	// this many goats remain, the goats can no longer build a safe fence.
	TigerWinCaptures int

	// GoatsMoveDuringPlacement allows goats already on the board to slide
	// before all goats have been dropped. Traditional play forbids it.
	GoatsMoveDuringPlacement bool

	// ForcedCapture makes capture compulsory for the tiger when available.
	// Traditional play leaves the choice to the tiger.
	ForcedCapture bool

	// HalfmoveLimit is the number of plies without a capture or a placement
	// after which the game is declared a draw. Zero disables the rule.
	//
	// It is generous - 100 plies, 50 moves each side - because closing a
	// fence around the tiger legitimately takes the goats a long sequence of
	// quiet slides, and cutting that short would hand the tiger draws it did
	// not earn.
	HalfmoveLimit int
}

// DefaultRuleSet returns the documented Bangladeshi rules: 1 tiger,
// 16 goats, tiger on the centre point, goats move first.
func DefaultRuleSet() RuleSet {
	return RuleSet{
		TotalGoats:               16,
		TigerStart:               ID(2, 2),
		FirstPlayer:              GoatPlayer,
		TigerWinCaptures:         9,
		GoatsMoveDuringPlacement: false,
		ForcedCapture:            false,
		HalfmoveLimit:            100,
	}
}

// Rules is the behaviour every ruleset variant must provide. The UI, the CLI
// and the AI all talk to the game exclusively through this interface.
type Rules interface {
	NewGame() GameState
	LegalMoves(state GameState) []Move
	IsValidMove(state GameState, move Move) bool
	ApplyMove(state GameState, move Move) (GameState, error)
	Winner(state GameState) Winner
	IsGameOver(state GameState) bool
	Config() RuleSet
}

// TraditionalBaghBandiRules implements Rules for the classic Bengali game.
//
// Rules, in full:
//
//  1. The board is the 5x5 Alquerque grid (see board.go).
//  2. The tiger starts on the centre point. The goat player has 16 goats in
//     hand and moves first.
//  3. Placement phase: on their turn the goat player drops one goat on any
//     empty point. Goats already on the board may not move yet.
//  4. Movement phase: once all 16 goats have been placed, the goat player
//     instead slides one goat along a drawn line to an adjacent empty point.
//  5. The tiger, on every turn, either slides along a line to an adjacent
//     empty point, or jumps in a straight line over exactly one adjacent
//     goat onto the empty point directly beyond it, removing that goat.
//     Only one goat may be taken per jump, and chained jumps are not allowed.
//  6. Goats never capture and never jump.
//  7. The tiger wins on eating TigerWinCaptures goats (9 by default), or if
//     the goat player has no legal move at all.
//  8. The goats win when the tiger has no legal move - neither a slide nor
//     a jump.
//  9. The game is drawn after HalfmoveLimit plies (100 by default) with no
//     capture and no placement.
type TraditionalBaghBandiRules struct {
	cfg   RuleSet
	board *Board
}

// NewRules returns the traditional rules with the supplied configuration.
func NewRules(cfg RuleSet) *TraditionalBaghBandiRules {
	return &TraditionalBaghBandiRules{cfg: cfg, board: NewBoard()}
}

// NewDefaultRules returns the traditional rules with DefaultRuleSet.
func NewDefaultRules() *TraditionalBaghBandiRules {
	return NewRules(DefaultRuleSet())
}

// Config returns the ruleset in force.
func (r *TraditionalBaghBandiRules) Config() RuleSet { return r.cfg }

// Board returns the shared board topology.
func (r *TraditionalBaghBandiRules) Board() *Board { return r.board }

// NewGame returns the opening position.
func (r *TraditionalBaghBandiRules) NewGame() GameState {
	var s GameState
	s.Tiger = r.cfg.TigerStart
	s.Occupancy[r.cfg.TigerStart] = Tiger
	s.Turn = r.cfg.FirstPlayer
	s.Status = InProgress
	return s
}

// LegalMoves returns every move available to the side to move.
func (r *TraditionalBaghBandiRules) LegalMoves(state GameState) []Move {
	return r.AppendLegalMoves(nil, state)
}

// AppendLegalMoves appends the legal moves to dst and returns it. The AI uses
// this form to reuse buffers and keep search allocation-free.
func (r *TraditionalBaghBandiRules) AppendLegalMoves(dst []Move, state GameState) []Move {
	if state.Status != InProgress {
		return dst
	}
	if state.Turn == TigerPlayer {
		return r.appendTigerMoves(dst, &state)
	}
	return r.appendGoatMoves(dst, &state)
}

func (r *TraditionalBaghBandiRules) appendTigerMoves(dst []Move, s *GameState) []Move {
	from := s.Tiger
	start := len(dst)

	for _, j := range r.board.Jumps(from) {
		if s.Occupancy[j.Over] == Goat && s.Occupancy[j.To] == Empty {
			dst = append(dst, CaptureMove(from, j.Over, j.To))
		}
	}
	captures := len(dst) - start
	if captures > 0 && r.cfg.ForcedCapture {
		return dst
	}

	for _, to := range r.board.Neighbors(from) {
		if s.Occupancy[to] == Empty {
			dst = append(dst, StepMove(from, to))
		}
	}
	return dst
}

func (r *TraditionalBaghBandiRules) appendGoatMoves(dst []Move, s *GameState) []Move {
	placing := s.InPlacementPhase(&r.cfg)

	if placing {
		for id := PositionID(0); id < PositionCount; id++ {
			if s.Occupancy[id] == Empty {
				dst = append(dst, PlaceMove(id))
			}
		}
		if !r.cfg.GoatsMoveDuringPlacement {
			return dst
		}
	}

	for id := PositionID(0); id < PositionCount; id++ {
		if s.Occupancy[id] != Goat {
			continue
		}
		for _, to := range r.board.Neighbors(id) {
			if s.Occupancy[to] == Empty {
				dst = append(dst, StepMove(id, to))
			}
		}
	}
	return dst
}

// IsValidMove reports whether move is legal in state.
func (r *TraditionalBaghBandiRules) IsValidMove(state GameState, move Move) bool {
	return r.validate(&state, move) == nil
}

func (r *TraditionalBaghBandiRules) validate(s *GameState, m Move) error {
	if s.Status != InProgress {
		return ErrGameOver
	}
	if !r.board.IsValid(m.To) {
		return ErrOutOfBounds
	}
	if s.Occupancy[m.To] != Empty {
		return ErrOccupiedPosition
	}

	switch m.Kind {
	case Placement:
		if s.Turn != GoatPlayer {
			return ErrWrongTurn
		}
		if !s.InPlacementPhase(&r.cfg) {
			return ErrInvalidMove
		}
		return nil

	case Step:
		if !r.board.IsValid(m.From) {
			return ErrOutOfBounds
		}
		piece := s.Occupancy[m.From]
		if piece == Empty {
			return ErrInvalidMove
		}
		if piece != s.Turn.Piece() {
			return ErrWrongTurn
		}
		if piece == Goat && s.InPlacementPhase(&r.cfg) && !r.cfg.GoatsMoveDuringPlacement {
			return ErrInvalidMove
		}
		if piece == Tiger && r.cfg.ForcedCapture && r.tigerHasCapture(s) {
			return ErrInvalidMove
		}
		if !r.board.AreAdjacent(m.From, m.To) {
			return ErrInvalidMove
		}
		return nil

	case Capture:
		if s.Turn != TigerPlayer {
			return ErrWrongTurn
		}
		if !r.board.IsValid(m.From) || s.Occupancy[m.From] != Tiger {
			return ErrInvalidMove
		}
		over, ok := r.board.JumpOver(m.From, m.To)
		if !ok || over != m.Over || s.Occupancy[over] != Goat {
			return ErrInvalidMove
		}
		return nil

	default:
		return ErrInvalidMove
	}
}

func (r *TraditionalBaghBandiRules) tigerHasCapture(s *GameState) bool {
	for _, j := range r.board.Jumps(s.Tiger) {
		if s.Occupancy[j.Over] == Goat && s.Occupancy[j.To] == Empty {
			return true
		}
	}
	return false
}

// ApplyMove returns the state that results from playing move. The input state
// is never modified - it is passed and returned by value.
func (r *TraditionalBaghBandiRules) ApplyMove(state GameState, move Move) (GameState, error) {
	if err := r.validate(&state, move); err != nil {
		return state, err
	}

	next := state
	switch move.Kind {
	case Placement:
		next.Occupancy[move.To] = Goat
		next.GoatsPlaced++
		next.HalfmoveClock = 0

	case Step:
		piece := next.Occupancy[move.From]
		next.Occupancy[move.From] = Empty
		next.Occupancy[move.To] = piece
		if piece == Tiger {
			next.Tiger = move.To
		}
		next.HalfmoveClock++

	case Capture:
		next.Occupancy[move.From] = Empty
		next.Occupancy[move.Over] = Empty
		next.Occupancy[move.To] = Tiger
		next.Tiger = move.To
		next.GoatsCaptured++
		next.HalfmoveClock = 0
	}

	next.Turn = state.Turn.Opponent()
	next.Status = r.evaluateStatus(&next)
	return next, nil
}

// evaluateStatus decides the outcome of the position that has just arisen.
func (r *TraditionalBaghBandiRules) evaluateStatus(s *GameState) GameStatus {
	if s.GoatsCaptured >= r.cfg.TigerWinCaptures {
		return TigerWon
	}
	if r.cfg.HalfmoveLimit > 0 && s.HalfmoveClock >= r.cfg.HalfmoveLimit {
		return Draw
	}
	// The side to move having no move at all loses; for the tiger that is
	// the goats' fence, for the goats it is total immobilisation.
	if !r.hasAnyMove(s) {
		if s.Turn == TigerPlayer {
			return GoatsWon
		}
		return TigerWon
	}
	return InProgress
}

func (r *TraditionalBaghBandiRules) hasAnyMove(s *GameState) bool {
	if s.Turn == TigerPlayer {
		for _, to := range r.board.Neighbors(s.Tiger) {
			if s.Occupancy[to] == Empty {
				return true
			}
		}
		return r.tigerHasCapture(s)
	}

	if s.InPlacementPhase(&r.cfg) {
		for id := PositionID(0); id < PositionCount; id++ {
			if s.Occupancy[id] == Empty {
				return true
			}
		}
		if !r.cfg.GoatsMoveDuringPlacement {
			return false
		}
	}
	for id := PositionID(0); id < PositionCount; id++ {
		if s.Occupancy[id] != Goat {
			continue
		}
		for _, to := range r.board.Neighbors(id) {
			if s.Occupancy[to] == Empty {
				return true
			}
		}
	}
	return false
}

// Winner returns the winning side of a finished game.
func (r *TraditionalBaghBandiRules) Winner(state GameState) Winner {
	switch state.Status {
	case TigerWon:
		return TigerWinner
	case GoatsWon:
		return GoatWinner
	default:
		return NoWinner
	}
}

// IsGameOver reports whether the game has finished.
func (r *TraditionalBaghBandiRules) IsGameOver(state GameState) bool {
	return state.Status != InProgress
}

var _ Rules = (*TraditionalBaghBandiRules)(nil)
