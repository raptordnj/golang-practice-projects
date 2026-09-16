package renderer

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// PlayerRenderer handles rendering the player
type PlayerRenderer struct {
	object *canvas.Circle
}

// NewPlayerRenderer creates a new player renderer
func NewPlayerRenderer(playerColor color.Color) *PlayerRenderer {
	return &PlayerRenderer{
		object: canvas.NewCircle(playerColor),
	}
}

// GetObject returns the player canvas object
func (pr *PlayerRenderer) GetObject() *canvas.Circle {
	return pr.object
}

// Update updates the player's visual state
func (pr *PlayerRenderer) Update(x, y, size float32, isAlive bool) {
	pr.object.Resize(fyne.NewSize(size, size))
	pr.object.Move(fyne.NewPos(x, y))

	if isAlive {
		pr.object.Show()
	} else {
		pr.object.Hide()
	}
}

// Show shows the player
func (pr *PlayerRenderer) Show() {
	pr.object.Show()
}

// Hide hides the player
func (pr *PlayerRenderer) Hide() {
	pr.object.Hide()
}

// Refresh refreshes the player display
func (pr *PlayerRenderer) Refresh() {
	pr.object.Refresh()
}
