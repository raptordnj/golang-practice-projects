package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// MenuScreen represents the title/menu screen
type MenuScreen struct {
	container *fyne.Container
	title     *canvas.Text
	subtitle  *canvas.Text
	highScore *canvas.Text
	instructions *canvas.Text
}

// NewMenuScreen creates a new menu screen
func NewMenuScreen() *MenuScreen {
	title := canvas.NewText("RETRO MAZE", yellowColor())
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("PRESS SPACE TO START", whiteColor())
	subtitle.TextSize = 14
	subtitle.TextStyle = fyne.TextStyle{Monospace: true}
	subtitle.Alignment = fyne.TextAlignCenter

	highScore := canvas.NewText("", whiteColor())
	highScore.TextSize = 12
	highScore.TextStyle = fyne.TextStyle{Monospace: true}
	highScore.Alignment = fyne.TextAlignCenter

	instructions := canvas.NewText("ARROWS/WASD TO MOVE\nP TO PAUSE", grayColor())
	instructions.TextSize = 10
	instructions.TextStyle = fyne.TextStyle{Monospace: true}
	instructions.Alignment = fyne.TextAlignCenter

	return &MenuScreen{
		title:        title,
		subtitle:     subtitle,
		highScore:    highScore,
		instructions: instructions,
	}
}

// GetObject returns the menu screen container
func (ms *MenuScreen) GetObject(width, height float32) fyne.CanvasObject {
	ms.title.Resize(fyne.NewSize(width, 40))
	ms.title.Move(fyne.NewPos(0, height/4))

	ms.subtitle.Resize(fyne.NewSize(width, 30))
	ms.subtitle.Move(fyne.NewPos(0, height/2))

	ms.highScore.Resize(fyne.NewSize(width, 20))
	ms.highScore.Move(fyne.NewPos(0, height/2+40))

	ms.instructions.Resize(fyne.NewSize(width, 40))
	ms.instructions.Move(fyne.NewPos(0, height*3/4))

	ms.container = container.NewWithoutLayout(
		ms.title,
		ms.subtitle,
		ms.highScore,
		ms.instructions,
	)

	return ms.container
}

// UpdateHighScore updates the high score display
func (ms *MenuScreen) UpdateHighScore(score int) {
	if score > 0 {
		ms.highScore.Text = "HIGH SCORE: " + formatScore(score)
	} else {
		ms.highScore.Text = ""
	}
	ms.highScore.Refresh()
}

// Show shows the menu screen
func (ms *MenuScreen) Show() {
	ms.title.Show()
	ms.subtitle.Show()
	ms.highScore.Show()
	ms.instructions.Show()
}

// Hide hides the menu screen
func (ms *MenuScreen) Hide() {
	ms.title.Hide()
	ms.subtitle.Hide()
	ms.highScore.Hide()
	ms.instructions.Hide()
}

// Refresh refreshes the menu screen
func (ms *MenuScreen) Refresh() {
	if ms.container != nil {
		ms.container.Refresh()
	}
}

// Color helper functions
func yellowColor() color.Color {
	return color.RGBA{R: 255, G: 255, B: 0, A: 255}
}

func whiteColor() color.Color {
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

func grayColor() color.Color {
	return color.RGBA{R: 180, G: 180, B: 180, A: 255}
}

// formatScore formats a score with leading zeros
func formatScore(score int) string {
	if score < 0 {
		score = 0
	}
	if score > 999999 {
		score = 999999
	}
	result := "000000"
	s := fmt.Sprintf("%d", score)
	return result[:len(result)-len(s)] + s
}
