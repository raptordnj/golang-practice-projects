package renderer

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// EnemyRenderer handles rendering enemies
type EnemyRenderer struct {
	objects []*canvas.Circle
}

// NewEnemyRenderer creates a new enemy renderer
func NewEnemyRenderer() *EnemyRenderer {
	return &EnemyRenderer{
		objects: make([]*canvas.Circle, 0),
	}
}

// Update updates the enemy rendering
func (er *EnemyRenderer) Update(enemies []EnemyData, size float32) {
	// Hide old objects
	for _, obj := range er.objects {
		obj.Hide()
	}
	er.objects = nil

	for _, enemy := range enemies {
		if !enemy.Alive {
			continue
		}

		obj := canvas.NewCircle(enemy.Color)
		obj.Resize(fyne.NewSize(size, size))
		obj.Move(fyne.NewPos(enemy.X, enemy.Y))
		obj.Show()

		er.objects = append(er.objects, obj)
	}
}

// GetObjects returns the enemy canvas objects
func (er *EnemyRenderer) GetObjects() []*canvas.Circle {
	return er.objects
}

// Show shows all enemies
func (er *EnemyRenderer) Show() {
	for _, obj := range er.objects {
		obj.Show()
	}
}

// Hide hides all enemies
func (er *EnemyRenderer) Hide() {
	for _, obj := range er.objects {
		obj.Hide()
	}
}

// Refresh refreshes all enemies
func (er *EnemyRenderer) Refresh() {
	for _, obj := range er.objects {
		obj.Refresh()
	}
}

// EnemyData holds data needed to render an enemy
type EnemyData struct {
	X            float32
	Y            float32
	Color        color.Color
	Alive        bool
	IsVulnerable bool
}
