package ui

import (
	"image"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"bagh-bandi/internal/game"
)

// midGame builds a representative position with both animals on the board.
func midGame(t *testing.T) (game.GameState, *game.TraditionalBaghBandiRules) {
	t.Helper()
	rules := game.NewDefaultRules()
	s := rules.NewGame()
	for _, n := range []string{"a1", "b1", "c1", "a2", "e2", "a3", "e3", "b4", "d4", "c5"} {
		id, ok := game.ParseNotation(n)
		if !ok {
			t.Fatalf("bad fixture point %q", n)
		}
		s.Occupancy[id] = game.Goat
		s.GoatsPlaced++
	}
	s.Turn = game.TigerPlayer
	return s, rules
}

// renderBoard draws the widget offscreen and returns the pixels.
func renderBoard(t *testing.T, v *BoardView, size fyne.Size) image.Image {
	t.Helper()
	w := test.NewWindow(v)
	t.Cleanup(w.Close)
	w.Resize(size)
	img := w.Canvas().Capture()
	if img == nil {
		t.Fatal("capture produced no image")
	}
	return img
}

// The board must actually draw the pieces - a silently blank board is the
// failure mode a headless test is here to catch.
func TestBoardRendersPieces(t *testing.T) {
	s, rules := midGame(t)

	empty := NewBoardView(BoardThemes[0], false)
	empty.SetState(rules.NewGame())
	before := renderBoard(t, empty, fyne.NewSize(400, 400))

	populated := NewBoardView(BoardThemes[0], false)
	populated.SetState(s)
	after := renderBoard(t, populated, fyne.NewSize(400, 400))

	if countDifferingPixels(before, after) < 500 {
		t.Error("adding ten goats barely changed the rendering - are the pieces drawn?")
	}
}

// Highlights must be visible, otherwise a player cannot see their options.
func TestSelectionAndTargetsAreDrawn(t *testing.T) {
	s, rules := midGame(t)

	plain := NewBoardView(BoardThemes[0], false)
	plain.SetState(s)
	before := renderBoard(t, plain, fyne.NewSize(400, 400))

	highlighted := NewBoardView(BoardThemes[0], false)
	highlighted.SetState(s)
	var fromTiger []game.Move
	for _, m := range rules.LegalMoves(s) {
		if m.From == s.Tiger {
			fromTiger = append(fromTiger, m)
		}
	}
	if len(fromTiger) == 0 {
		t.Fatal("fixture gives the tiger no moves")
	}
	highlighted.SetSelection(s.Tiger, fromTiger)
	after := renderBoard(t, highlighted, fyne.NewSize(400, 400))

	if countDifferingPixels(before, after) < 100 {
		t.Error("selecting the tiger produced no visible highlight")
	}
}

// Every theme must render without panicking and must produce a distinct look.
func TestEveryThemeRenders(t *testing.T) {
	s, _ := midGame(t)
	var shots []image.Image

	for _, th := range BoardThemes {
		v := NewBoardView(th, false)
		v.SetState(s)
		shots = append(shots, renderBoard(t, v, fyne.NewSize(360, 360)))
	}
	for i := 1; i < len(shots); i++ {
		if countDifferingPixels(shots[0], shots[i]) < 500 {
			t.Errorf("theme %q renders almost identically to %q",
				BoardThemes[i].Name, BoardThemes[0].Name)
		}
	}
}

// Resizing must not crash and must actually redraw at the new scale.
func TestBoardRendersAtManySizes(t *testing.T) {
	s, _ := midGame(t)
	for _, size := range []fyne.Size{
		fyne.NewSize(240, 240), fyne.NewSize(400, 700),
		fyne.NewSize(900, 500), fyne.NewSize(1000, 1000),
	} {
		v := NewBoardView(BoardThemes[0], false)
		v.SetState(s)
		img := renderBoard(t, v, size)
		if img.Bounds().Empty() {
			t.Errorf("size %v produced an empty image", size)
		}
	}
}

// A tap must be translated into the board point under the cursor.
func TestTapReportsTheClickedPoint(t *testing.T) {
	v := NewBoardView(BoardThemes[0], false)
	v.SetState(game.NewDefaultRules().NewGame())
	v.Resize(fyne.NewSize(500, 500))

	var got game.PositionID = game.NoPosition
	v.OnPointTapped = func(id game.PositionID) { got = id }

	g := NewGeometry(v.Size())
	want := game.ID(1, 3)
	v.Tapped(&fyne.PointEvent{Position: g.Center(want)})

	if got != want {
		t.Errorf("tap reported %s, want %s", game.Notation(got), game.Notation(want))
	}
}

func TestTapBetweenPointsIsIgnored(t *testing.T) {
	v := NewBoardView(BoardThemes[0], false)
	v.SetState(game.NewDefaultRules().NewGame())
	v.Resize(fyne.NewSize(500, 500))

	called := false
	v.OnPointTapped = func(game.PositionID) { called = true }

	g := NewGeometry(v.Size())
	a, b := g.Center(game.ID(0, 0)), g.Center(game.ID(0, 1))
	v.Tapped(&fyne.PointEvent{Position: fyne.NewPos((a.X+b.X)/2, (a.Y+b.Y)/2)})

	if called {
		t.Error("a tap between two points was reported as a move")
	}
}

// With animation off, AnimateMove must complete synchronously - the
// controller relies on the callback always arriving.
func TestAnimateMoveCallsBackWithAnimationOff(t *testing.T) {
	v := NewBoardView(BoardThemes[0], false)
	called := false
	v.AnimateMove(game.StepMove(game.ID(2, 2), game.ID(2, 1)), func() { called = true })
	if !called {
		t.Error("the completion callback was not called")
	}
}

func TestAnimateMoveSkipsPlacements(t *testing.T) {
	v := NewBoardView(BoardThemes[0], true)
	called := false
	v.AnimateMove(game.PlaceMove(game.ID(0, 0)), func() { called = true })
	if !called {
		t.Error("placements should complete immediately, with no slide")
	}
}

func countDifferingPixels(a, b image.Image) int {
	bounds := a.Bounds()
	if bounds != b.Bounds() {
		return bounds.Dx() * bounds.Dy()
	}
	diff := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				diff++
			}
		}
	}
	return diff
}
