package game

import "errors"

// Engine errors. Normal invalid gameplay never panics; it returns one of
// these so the UI can present a friendly message.
var (
	// ErrInvalidMove means the move is not legal in the current position.
	ErrInvalidMove = errors.New("invalid move")
	// ErrOccupiedPosition means the destination point already holds a piece.
	ErrOccupiedPosition = errors.New("position is occupied")
	// ErrWrongTurn means the move belongs to the side that is not to move.
	ErrWrongTurn = errors.New("not this player's turn")
	// ErrGameOver means the game has already finished.
	ErrGameOver = errors.New("game is over")
	// ErrOutOfBounds means a position id does not exist on the board.
	ErrOutOfBounds = errors.New("position is off the board")
	// ErrNothingToUndo means the history has no earlier state.
	ErrNothingToUndo = errors.New("nothing to undo")
	// ErrNothingToRedo means the history has no later state.
	ErrNothingToRedo = errors.New("nothing to redo")
)
