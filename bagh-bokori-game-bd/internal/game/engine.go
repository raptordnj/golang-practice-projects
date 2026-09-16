package game

// Engine is a playable game: a ruleset, the current state, and the move
// history that powers undo/redo.
//
// The engine performs no I/O and imports nothing outside the standard
// library, so it compiles unchanged for desktop, WebAssembly and mobile.
type Engine struct {
	rules   *TraditionalBaghBandiRules
	history []GameState // history[0] is the opening position
	moves   []Move      // moves[i] led from history[i] to history[i+1]
	cursor  int         // index into history of the current state
}

// NewEngine starts a new game under the given rules.
func NewEngine(cfg RuleSet) *Engine {
	e := &Engine{rules: NewRules(cfg)}
	e.Restart()
	return e
}

// NewDefaultEngine starts a new game under the default Bangladeshi rules.
func NewDefaultEngine() *Engine { return NewEngine(DefaultRuleSet()) }

// Rules returns the ruleset implementation in use.
func (e *Engine) Rules() *TraditionalBaghBandiRules { return e.rules }

// Board returns the board topology.
func (e *Engine) Board() *Board { return e.rules.Board() }

// Config returns the active rule configuration.
func (e *Engine) Config() RuleSet { return e.rules.Config() }

// State returns a copy of the current state. Callers may do anything they
// like with it without affecting the engine.
func (e *Engine) State() GameState { return e.history[e.cursor] }

// LegalMoves returns the moves available right now.
func (e *Engine) LegalMoves() []Move { return e.rules.LegalMoves(e.State()) }

// LegalMovesFrom returns the legal moves whose piece stands on from. During
// the placement phase it returns nothing, because placements have no origin.
func (e *Engine) LegalMovesFrom(from PositionID) []Move {
	var out []Move
	for _, m := range e.LegalMoves() {
		if m.Kind != Placement && m.From == from {
			out = append(out, m)
		}
	}
	return out
}

// Play applies a move, truncating any redo history.
func (e *Engine) Play(m Move) error {
	next, err := e.rules.ApplyMove(e.State(), m)
	if err != nil {
		return err
	}
	e.history = append(e.history[:e.cursor+1], next)
	e.moves = append(e.moves[:e.cursor], m)
	e.cursor++
	return nil
}

// Restart returns to a fresh opening position and clears the history.
func (e *Engine) Restart() {
	e.history = []GameState{e.rules.NewGame()}
	e.moves = nil
	e.cursor = 0
}

// CanUndo reports whether there is an earlier state.
func (e *Engine) CanUndo() bool { return e.cursor > 0 }

// CanRedo reports whether there is a later state.
func (e *Engine) CanRedo() bool { return e.cursor < len(e.history)-1 }

// Undo steps back one ply. Redo history is preserved.
func (e *Engine) Undo() error {
	if !e.CanUndo() {
		return ErrNothingToUndo
	}
	e.cursor--
	return nil
}

// UndoTurn steps back to the previous position in which player is to move.
// Against the AI this rewinds both the AI's reply and the human's move, so
// the human is never handed a turn that is not theirs.
func (e *Engine) UndoTurn(player Player) error {
	if !e.CanUndo() {
		return ErrNothingToUndo
	}
	start := e.cursor
	for e.cursor > 0 {
		e.cursor--
		s := e.history[e.cursor]
		if s.Turn == player && !s.IsOver() {
			return nil
		}
	}
	if e.cursor == 0 {
		return nil
	}
	e.cursor = start
	return ErrNothingToUndo
}

// Redo steps forward one ply.
func (e *Engine) Redo() error {
	if !e.CanRedo() {
		return ErrNothingToRedo
	}
	e.cursor++
	return nil
}

// History returns the moves played up to the current cursor.
func (e *Engine) History() []Move {
	out := make([]Move, e.cursor)
	copy(out, e.moves[:e.cursor])
	return out
}

// LastMove returns the move that produced the current state.
func (e *Engine) LastMove() (Move, bool) {
	if e.cursor == 0 {
		return Move{}, false
	}
	return e.moves[e.cursor-1], true
}

// IsGameOver reports whether the current game has finished.
func (e *Engine) IsGameOver() bool { return e.State().IsOver() }

// Winner returns the winner of the current position.
func (e *Engine) Winner() Winner { return e.rules.Winner(e.State()) }

// LoadSnapshot replaces the engine contents with a restored game.
func (e *Engine) LoadSnapshot(snap Snapshot) error {
	rules := NewRules(snap.Rules)
	history := []GameState{rules.NewGame()}
	for _, m := range snap.Moves {
		next, err := rules.ApplyMove(history[len(history)-1], m)
		if err != nil {
			return err
		}
		history = append(history, next)
	}
	if snap.Cursor < 0 || snap.Cursor >= len(history) {
		return ErrInvalidMove
	}
	e.rules = rules
	e.history = history
	e.moves = append([]Move(nil), snap.Moves...)
	e.cursor = snap.Cursor
	return nil
}

// Snapshot captures the whole game for saving.
func (e *Engine) Snapshot() Snapshot {
	return Snapshot{
		Version: snapshotVersion,
		Rules:   e.rules.Config(),
		Moves:   append([]Move(nil), e.moves...),
		Cursor:  e.cursor,
	}
}
