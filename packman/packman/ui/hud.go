package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// HUD represents the heads-up display
type HUD struct {
	container    *fyne.Container
	scoreText    *canvas.Text
	levelText    *canvas.Text
	livesText    *canvas.Text
	highScoreText *canvas.Text
}

// NewHUD creates a new HUD
func NewHUD() *HUD {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	scoreText := canvas.NewText("SCORE: 000000", white)
	scoreText.TextSize = 12
	scoreText.TextStyle = fyne.TextStyle{Monospace: true}

	levelText := canvas.NewText("LEVEL: 01", white)
	levelText.TextSize = 12
	levelText.TextStyle = fyne.TextStyle{Monospace: true}

	livesText := canvas.NewText("LIVES: ♥♥♥", white)
	livesText.TextSize = 12
	livesText.TextStyle = fyne.TextStyle{Monospace: true}

	highScoreText := canvas.NewText("HI: 000000", yellow)
	highScoreText.TextSize = 12
	highScoreText.TextStyle = fyne.TextStyle{Monospace: true}

	return &HUD{
		scoreText:     scoreText,
		levelText:     levelText,
		livesText:     livesText,
		highScoreText: highScoreText,
	}
}

// GetObject returns the HUD container
func (h *HUD) GetObject(width float32) fyne.CanvasObject {
	h.scoreText.Resize(fyne.NewSize(width/3, 20))
	h.scoreText.Move(fyne.NewPos(5, 0))

	h.levelText.Resize(fyne.NewSize(width/3, 20))
	h.levelText.Move(fyne.NewPos(width/3, 0))

	h.livesText.Resize(fyne.NewSize(width/3, 20))
	h.livesText.Move(fyne.NewPos(width*2/3, 0))

	h.highScoreText.Resize(fyne.NewSize(width/2, 20))
	h.highScoreText.Move(fyne.NewPos(width/2, 0))

	h.container = container.NewWithoutLayout(
		h.scoreText,
		h.levelText,
		h.livesText,
		h.highScoreText,
	)

	return h.container
}

// Update updates the HUD with current game state
func (h *HUD) Update(score, highScore, level, lives int) {
	h.scoreText.Text = "SCORE: " + FormatScore6(score)
	h.highScoreText.Text = "HI: " + FormatScore6(highScore)
	h.levelText.Text = "LV: " + formatLevel(level)

	livesStr := "LIVES: "
	for i := 0; i < lives; i++ {
		livesStr += "♥"
	}
	h.livesText.Text = livesStr

	h.Refresh()
}

// Show shows the HUD
func (h *HUD) Show() {
	h.scoreText.Show()
	h.levelText.Show()
	h.livesText.Show()
	h.highScoreText.Show()
}

// Hide hides the HUD
func (h *HUD) Hide() {
	h.scoreText.Hide()
	h.levelText.Hide()
	h.livesText.Hide()
	h.highScoreText.Hide()
}

// Refresh refreshes the HUD
func (h *HUD) Refresh() {
	h.scoreText.Refresh()
	h.levelText.Refresh()
	h.livesText.Refresh()
	h.highScoreText.Refresh()
}

// Position positions the HUD at the given coordinates
func (h *HUD) Position(x, y float32) {
	// Get current positions relative to container
	scoreX := h.scoreText.Position().X
	levelX := h.levelText.Position().X
	livesX := h.livesText.Position().X
	highScoreX := h.highScoreText.Position().X

	h.scoreText.Move(fyne.NewPos(x+scoreX, y))
	h.levelText.Move(fyne.NewPos(x+levelX, y))
	h.livesText.Move(fyne.NewPos(x+livesX, y))
	h.highScoreText.Move(fyne.NewPos(x+highScoreX, y))
}

// FormatScore6 formats a score with leading zeros
func FormatScore6(score int) string {
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

func formatLevel(level int) string {
	if level < 1 {
		level = 1
	}
	if level > 99 {
		level = 99
	}
	return fmt.Sprintf("%02d", level)
}
