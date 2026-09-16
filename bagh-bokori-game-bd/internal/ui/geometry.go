package ui

import (
	"fyne.io/fyne/v2"

	"bagh-bandi/internal/game"
)

// Geometry maps the abstract board graph onto whatever space the window
// currently gives us. Nothing in the UI hard-codes a pixel coordinate: every
// drawn object is positioned through this type, so the board scales and
// stays square at any window size.
type Geometry struct {
	Origin fyne.Position // top-left of the playing grid
	Side   float32       // length of one side of the grid
	Cell   float32       // distance between adjacent points
	Point  float32       // radius of an empty point marker
	Piece  float32       // radius of a piece
	Margin float32       // padding between the grid and the board panel
}

const (
	// minBoardSide keeps the board usable if the window is shrunk hard.
	minBoardSide = 160
	// boardMarginRatio is the share of the board panel left as a border.
	boardMarginRatio = 0.11
)

// NewGeometry computes the layout for the given widget size.
func NewGeometry(size fyne.Size) Geometry {
	extent := size.Width
	if size.Height < extent {
		extent = size.Height
	}
	if extent < minBoardSide {
		extent = minBoardSide
	}

	margin := extent * boardMarginRatio
	side := extent - 2*margin
	cell := side / float32(game.BoardSize-1)

	return Geometry{
		Origin: fyne.NewPos((size.Width-side)/2, (size.Height-side)/2),
		Side:   side,
		Cell:   cell,
		Point:  cell * 0.09,
		Piece:  cell * 0.30,
		Margin: margin,
	}
}

// Center returns the pixel centre of a board point.
func (g Geometry) Center(id game.PositionID) fyne.Position {
	row, col := game.RowCol(id)
	return fyne.NewPos(
		g.Origin.X+float32(col)*g.Cell,
		g.Origin.Y+float32(row)*g.Cell,
	)
}

// Lerp returns the point a fraction t of the way from a to b, used by the
// move animation.
func Lerp(a, b fyne.Position, t float32) fyne.Position {
	return fyne.NewPos(a.X+(b.X-a.X)*t, a.Y+(b.Y-a.Y)*t)
}

// Panel returns the rectangle of the board's backing panel.
func (g Geometry) Panel() (fyne.Position, fyne.Size) {
	pos := fyne.NewPos(g.Origin.X-g.Margin, g.Origin.Y-g.Margin)
	extent := g.Side + 2*g.Margin
	return pos, fyne.NewSize(extent, extent)
}

// Nearest returns the board point closest to pos, and whether the click was
// close enough to count as hitting it.
func (g Geometry) Nearest(pos fyne.Position) (game.PositionID, bool) {
	best := game.NoPosition
	bestDist := float32(-1)

	for id := game.PositionID(0); int(id) < game.PositionCount; id++ {
		c := g.Center(id)
		dx, dy := pos.X-c.X, pos.Y-c.Y
		d := dx*dx + dy*dy
		if bestDist < 0 || d < bestDist {
			bestDist, best = d, id
		}
	}

	// A generous hit radius: fiddly clicking is the fastest way to make a
	// board game feel cheap.
	hit := g.Cell * 0.45
	return best, bestDist <= hit*hit
}

// CenterOf is a convenience that offsets a circle's top-left corner so its
// centre lands on pos.
func topLeft(center fyne.Position, radius float32) fyne.Position {
	return fyne.NewPos(center.X-radius, center.Y-radius)
}

func squareSize(radius float32) fyne.Size {
	return fyne.NewSize(radius*2, radius*2)
}
