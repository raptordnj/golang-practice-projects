package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// GameOverScreen represents the game over screen
type GameOverScreen struct {
	container  *fyne.Container
	title      *canvas.Text
	scoreText  *canvas.Text
	highScore  *canvas.Text
	newHighScore *canvas.Text
	continueText *canvas.Text
}

// NewGameOverScreen creates a new game over screen
func NewGameOverScreen() *GameOverScreen {
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	title := canvas.NewText("GAME OVER", red)
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	title.Alignment = fyne.TextAlignCenter

	scoreText := canvas.NewText("SCORE: 000000", white)
	scoreText.TextSize = 14
	scoreText.TextStyle = fyne.TextStyle{Monospace: true}
	scoreText.Alignment = fyne.TextAlignCenter

	highScore := canvas.NewText("HIGH SCORE: 000000", yellow)
	highScore.TextSize = 12
	highScore.TextStyle = fyne.TextStyle{Monospace: true}
	highScore.Alignment = fyne.TextAlignCenter

	newHighScore := canvas.NewText("NEW HIGH SCORE!", green)
	newHighScore.TextSize = 14
	newHighScore.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	newHighScore.Alignment = fyne.TextAlignCenter

	continueText := canvas.NewText("PRESS SPACE TO CONTINUE", white)
	continueText.TextSize = 10
	continueText.TextStyle = fyne.TextStyle{Monospace: true}
	continueText.Alignment = fyne.TextAlignCenter

	return &GameOverScreen{
		title:        title,
		scoreText:    scoreText,
		highScore:    highScore,
		newHighScore: newHighScore,
		continueText: continueText,
	}
}

// GetObject returns the game over container
func (gos *GameOverScreen) GetObject(width, height float32) fyne.CanvasObject {
	centerY := height / 3

	gos.title.Resize(fyne.NewSize(width, 30))
	gos.title.Move(fyne.NewPos(0, centerY))

	gos.scoreText.Resize(fyne.NewSize(width, 20))
	gos.scoreText.Move(fyne.NewPos(0, centerY+40))

	gos.highScore.Resize(fyne.NewSize(width, 20))
	gos.highScore.Move(fyne.NewPos(0, centerY+65))

	gos.newHighScore.Resize(fyne.NewSize(width, 20))
	gos.newHighScore.Move(fyne.NewPos(0, centerY+90))

	gos.continueText.Resize(fyne.NewSize(width, 20))
	gos.continueText.Move(fyne.NewPos(0, centerY+130))

	gos.container = container.NewWithoutLayout(
		gos.title,
		gos.scoreText,
		gos.highScore,
		gos.newHighScore,
		gos.continueText,
	)

	return gos.container
}

// Update updates the game over screen with current scores
func (gos *GameOverScreen) Update(score, highScore int, isNewHighScore bool) {
	gos.scoreText.Text = "SCORE: " + fmt.Sprintf("%06d", score)
	gos.highScore.Text = "HIGH SCORE: " + fmt.Sprintf("%06d", highScore)

	if isNewHighScore {
		gos.newHighScore.Show()
	} else {
		gos.newHighScore.Hide()
	}

	gos.Refresh()
}

// Show shows the game over screen
func (gos *GameOverScreen) Show() {
	gos.title.Show()
	gos.scoreText.Show()
	gos.highScore.Show()
	gos.continueText.Show()
}

// Hide hides the game over screen
func (gos *GameOverScreen) Hide() {
	gos.title.Hide()
	gos.scoreText.Hide()
	gos.highScore.Hide()
	gos.newHighScore.Hide()
	gos.continueText.Hide()
}

// Refresh refreshes the game over screen
func (gos *GameOverScreen) Refresh() {
	if gos.container != nil {
		gos.container.Refresh()
	}
}

