package game

// GameState represents the current state of the game
type GameState int

const (
	// StateTitle is the title/menu screen
	StateTitle GameState = iota
	// StateReady is the brief countdown before gameplay starts
	StateReady
	// StatePlaying is active gameplay
	StatePlaying
	// StatePaused is when the game is paused
	StatePaused
	// StateLevelComplete is when a level is finished
	StateLevelComplete
	// StateGameOver is when the player loses all lives
	StateGameOver
)

// String returns a human-readable name for the game state
func (s GameState) String() string {
	switch s {
	case StateTitle:
		return "TITLE"
	case StateReady:
		return "READY"
	case StatePlaying:
		return "PLAYING"
	case StatePaused:
		return "PAUSED"
	case StateLevelComplete:
		return "LEVEL_COMPLETE"
	case StateGameOver:
		return "GAME_OVER"
	default:
		return "UNKNOWN"
	}
}
