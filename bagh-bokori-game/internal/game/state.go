package game

import "fmt"

type GameState struct {
	Board         Board
	Tigers        []PositionID
	Goats         []PositionID
	GoatsPlaced   int
	Turn          Player
	Status        GameStatus
	CapturedGoats int
	MoveHistory   []Move
}

type GameStatus int

const (
	StatusPlaying GameStatus = iota
	StatusTigerWin
	StatusGoatWin
)

type GameResult struct {
	Winner Player
	Status GameStatus
}

func NewGame() GameState {
	state := GameState{
		Board:       NewBoard(),
		Turn:        PlayerTiger,
		Status:      StatusPlaying,
		GoatsPlaced: 0,
	}
	// Place tigers at their starting positions
	tigerPositions := []PositionID{0, 4, 20, 24}
	for _, pos := range tigerPositions {
		state.Board = state.Board.Set(pos, Piece{Type: Tiger, Player: PlayerTiger})
		state.Tigers = append(state.Tigers, pos)
	}
	return state
}

func LegalMoves(state GameState) []Move {
	if state.Status != StatusPlaying {
		return nil
	}
	if state.Turn == PlayerTiger {
		return tigerLegalMoves(state)
	}
	return goatLegalMoves(state)
}

func IsValidMove(state GameState, move Move) bool {
	for _, m := range LegalMoves(state) {
		if m.From == move.From && m.To == move.To {
			return true
		}
	}
	return false
}

func ApplyMove(state GameState, move Move) (GameState, error) {
	if state.Status != StatusPlaying {
		return state, fmt.Errorf("game is over")
	}
	if !IsValidMove(state, move) {
		return state, fmt.Errorf("invalid move")
	}

	newState := state
	if move.From == PositionID(-1) {
		// Goat placement
		newState.Board = newState.Board.Set(move.To, Piece{Type: Goat, Player: PlayerGoat})
	} else {
		newState.Board = newState.Board.Set(move.To, newState.Board.At(move.From))
		newState.Board = newState.Board.Set(move.From, Piece{})
	}
	newState.MoveHistory = append(newState.MoveHistory, move)

	if state.Turn == PlayerTiger {
		// Check if this was a capture (jump over a goat)
		if isCaptureMove(state, move) {
			captured := capturePosition(state, move)
			newState.Board = newState.Board.Set(captured, Piece{})
			newState.CapturedGoats++
			// Remove from goats list
			var newGoats []PositionID
			for _, g := range newState.Goats {
				if g != captured {
					newGoats = append(newGoats, g)
				}
			}
			newState.Goats = newGoats
		}
		// Update tiger positions
		var newTigers []PositionID
		for _, t := range newState.Tigers {
			if t == move.From {
				newTigers = append(newTigers, move.To)
			} else {
				newTigers = append(newTigers, t)
			}
		}
		newState.Tigers = newTigers
	} else {
		// Goat move or placement
		if state.GoatsPlaced < 20 {
			newState.GoatsPlaced++
		}
		newState.Goats = append(newState.Goats, move.To)
	}

	// Switch turn
	if state.Turn == PlayerTiger {
		newState.Turn = PlayerGoat
	} else {
		newState.Turn = PlayerTiger
	}

	// Check game over
	checkGameOver(&newState)

	return newState, nil
}

func Undo(state GameState) GameState {
	if len(state.MoveHistory) == 0 {
		return state
	}
	newHistory := state.MoveHistory[:len(state.MoveHistory)-1]

	// Rebuild state by replaying all moves except the last
	state = NewGame()
	for _, m := range newHistory {
		var err error
		state, err = ApplyMove(state, m)
		if err != nil {
			return state
		}
	}
	return state
}

func Winner(state GameState) GameResult {
	switch state.Status {
	case StatusTigerWin:
		return GameResult{Winner: PlayerTiger, Status: StatusTigerWin}
	case StatusGoatWin:
		return GameResult{Winner: PlayerGoat, Status: StatusGoatWin}
	default:
		return GameResult{Status: StatusPlaying}
	}
}

func IsGameOver(state GameState) bool {
	return state.Status != StatusPlaying
}
