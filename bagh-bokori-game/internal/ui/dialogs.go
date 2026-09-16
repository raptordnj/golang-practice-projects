package ui

import (
	"bagh-bakri/internal/game"
	"fyne.io/fyne/v2"
)

func ShowSaveDialog(w fyne.Window, state game.GameState, player Player) {
	SaveGame(state, player)
}

func ShowLoadDialog(w fyne.Window) (game.GameState, Player, bool) {
	state, player, err := LoadGame()
	if err != nil {
		return game.NewGame(), PlayerHuman, false
	}
	return state, player, true
}
