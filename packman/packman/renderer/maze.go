package renderer

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// MazeRenderer handles rendering the maze
type MazeRenderer struct {
	container *fyne.Container
}

// NewMazeRenderer creates a new maze renderer
func NewMazeRenderer() *MazeRenderer {
	return &MazeRenderer{
		container: container.NewWithoutLayout(),
	}
}

// GetContainer returns the maze container
func (mr *MazeRenderer) GetContainer() *fyne.Container {
	return mr.container
}

// Render renders the maze grid
func (mr *MazeRenderer) Render(tiles [][]int, tileSize float32, wallColor color.Color) {
	mr.container.Objects = nil

	for y, row := range tiles {
		for x, tile := range row {
			if tile == 1 { // Wall
				wall := canvas.NewRectangle(wallColor)
				wall.Resize(fyne.NewSize(tileSize-1, tileSize-1))
				wall.Move(fyne.NewPos(float32(x)*tileSize, float32(y)*tileSize))
				mr.container.Objects = append(mr.container.Objects, wall)
			}
		}
	}

	mr.container.Refresh()
}

// Clear clears the maze rendering
func (mr *MazeRenderer) Clear() {
	mr.container.Objects = nil
	mr.container.Refresh()
}
