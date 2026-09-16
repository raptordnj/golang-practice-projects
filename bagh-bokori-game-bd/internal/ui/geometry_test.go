package ui

import (
	"testing"

	"fyne.io/fyne/v2"

	"bagh-bandi/internal/game"
)

func TestGeometryScalesWithTheWindow(t *testing.T) {
	small := NewGeometry(fyne.NewSize(400, 400))
	large := NewGeometry(fyne.NewSize(900, 900))

	if large.Cell <= small.Cell {
		t.Error("board did not grow with the window")
	}
	if large.Piece <= small.Piece {
		t.Error("pieces did not scale with the board")
	}
}

// A non-square window must still produce a square board, centred.
func TestGeometryStaysSquareAndCentred(t *testing.T) {
	g := NewGeometry(fyne.NewSize(1200, 500))

	topLeftPoint := g.Center(game.ID(0, 0))
	bottomRight := g.Center(game.ID(game.BoardSize-1, game.BoardSize-1))

	width := bottomRight.X - topLeftPoint.X
	height := bottomRight.Y - topLeftPoint.Y
	if diff := width - height; diff > 0.01 || diff < -0.01 {
		t.Errorf("board is %v x %v, want square", width, height)
	}

	leftGap := topLeftPoint.X
	rightGap := 1200 - bottomRight.X
	if diff := leftGap - rightGap; diff > 0.01 || diff < -0.01 {
		t.Errorf("board is not horizontally centred: %v vs %v", leftGap, rightGap)
	}
}

func TestGeometryClampsTinyWindows(t *testing.T) {
	g := NewGeometry(fyne.NewSize(10, 10))
	if g.Cell <= 0 || g.Piece <= 0 {
		t.Error("degenerate geometry for a tiny window")
	}
}

func TestNearestFindsTheClickedPoint(t *testing.T) {
	g := NewGeometry(fyne.NewSize(600, 600))

	for id := game.PositionID(0); int(id) < game.PositionCount; id++ {
		got, ok := g.Nearest(g.Center(id))
		if !ok || got != id {
			t.Errorf("click on %s resolved to %s (ok=%v)",
				game.Notation(id), game.Notation(got), ok)
		}
	}
}

func TestNearestRejectsClicksBetweenPoints(t *testing.T) {
	g := NewGeometry(fyne.NewSize(600, 600))
	a := g.Center(game.ID(0, 0))
	b := g.Center(game.ID(0, 1))
	mid := fyne.NewPos((a.X+b.X)/2, (a.Y+b.Y)/2)

	if _, ok := g.Nearest(mid); ok {
		t.Error("a click halfway between two points was accepted")
	}
}

func TestLerpEndpoints(t *testing.T) {
	a := fyne.NewPos(0, 0)
	b := fyne.NewPos(10, 20)
	if got := Lerp(a, b, 0); got != a {
		t.Errorf("t=0 gave %v", got)
	}
	if got := Lerp(a, b, 1); got != b {
		t.Errorf("t=1 gave %v", got)
	}
	if got := Lerp(a, b, 0.5); got.X != 5 || got.Y != 10 {
		t.Errorf("t=0.5 gave %v", got)
	}
}
