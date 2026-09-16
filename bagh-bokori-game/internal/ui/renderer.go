package ui

import (
	"fmt"

	"bagh-bakri/internal/game"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type FyneApp struct {
	a       fyne.App
	w       fyne.Window
	state   *game.GameState
	boardUI *BoardUI
	status  *widget.Label
}

func NewFyneApp() *FyneApp {
	return &FyneApp{
		a: app.New(),
	}
}

func (f *FyneApp) Run() {
	f.w = f.a.NewWindow("Bagh-Bakri")
	f.state = &game.GameState{}
	*f.state = game.NewGame()
	f.boardUI = NewBoardUI(f.state, f.onMove)

	f.status = widget.NewLabel("Turn: Tiger")
	f.updateStatus()

	restartBtn := widget.NewButton("Restart", f.onRestart)
	undoBtn := widget.NewButton("Undo", f.onUndo)

	top := container.NewHBox(
		f.status,
		widget.NewLabel("|"),
		widget.NewLabel(fmt.Sprintf("Captured: %d", f.state.CapturedGoats)),
		widget.NewLabel("|"),
		undoBtn,
		restartBtn,
	)

	content := container.NewBorder(top, nil, nil, nil, f.boardUI.CreateUI())
	f.w.SetContent(content)
	f.w.Resize(fyne.NewSize(400, 450))
	f.w.ShowAndRun()
}

func (f *FyneApp) onMove(m game.Move) {
	newState, err := game.ApplyMove(*f.state, m)
	if err != nil {
		return
	}
	*f.state = newState
	f.boardUI.SetState(f.state)
	f.updateUI()
}

func (f *FyneApp) updateUI() {
	f.updateStatus()
	f.w.SetContent(container.NewBorder(
		container.NewHBox(
			f.status,
			widget.NewLabel("|"),
			widget.NewLabel(fmt.Sprintf("Captured: %d", f.state.CapturedGoats)),
			widget.NewLabel("|"),
			widget.NewButton("Undo", f.onUndo),
			widget.NewButton("Restart", f.onRestart),
		),
		nil, nil, nil, f.boardUI.CreateUI(),
	))
}

func (f *FyneApp) updateStatus() {
	turn := "Tiger"
	if f.state.Turn == game.PlayerGoat {
		turn = "Goat"
	}
	f.status.SetText(fmt.Sprintf("Turn: %s", turn))
}

func (f *FyneApp) onRestart() {
	*f.state = game.NewGame()
	f.boardUI.SetState(f.state)
	f.updateUI()
}

func (f *FyneApp) onUndo() {
	*f.state = game.Undo(*f.state)
	f.boardUI.SetState(f.state)
	f.updateUI()
}
