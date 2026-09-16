package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"bagh-bandi/internal/game"
)

// BoardView is the interactive board widget.
//
// It holds no rules: it is given a GameState and a set of legal moves to
// draw, and it reports clicks back as board points. Deciding what a click
// means is the controller's job.
type BoardView struct {
	widget.BaseWidget

	state    game.GameState
	board    *game.Board
	theme    BoardTheme
	selected game.PositionID
	targets  map[game.PositionID]game.Move
	lastMove *game.Move
	flash    game.PositionID // point flashed red after an illegal action

	animate     bool
	animMove    *game.Move
	animPhase   float32
	animRunning *fyne.Animation

	// OnPointTapped is called with the board point the user clicked.
	OnPointTapped func(game.PositionID)
}

// NewBoardView creates the board widget.
func NewBoardView(theme BoardTheme, animate bool) *BoardView {
	b := &BoardView{
		board:    game.NewBoard(),
		theme:    theme,
		selected: game.NoPosition,
		flash:    game.NoPosition,
		targets:  map[game.PositionID]game.Move{},
		animate:  animate,
	}
	b.ExtendBaseWidget(b)
	return b
}

// SetState replaces the position being drawn.
func (b *BoardView) SetState(s game.GameState) {
	b.state = s
	b.Refresh()
}

// SetTheme swaps the colour scheme.
func (b *BoardView) SetTheme(t BoardTheme) {
	b.theme = t
	b.Refresh()
}

// SetAnimate turns move animation on or off.
func (b *BoardView) SetAnimate(on bool) { b.animate = on }

// SetSelection highlights a piece and the moves it may make. Pass
// game.NoPosition and nil to clear.
func (b *BoardView) SetSelection(from game.PositionID, moves []game.Move) {
	b.selected = from
	b.targets = make(map[game.PositionID]game.Move, len(moves))
	for _, m := range moves {
		b.targets[m.To] = m
	}
	b.Refresh()
}

// SetLastMove marks the move that produced the current position.
func (b *BoardView) SetLastMove(m *game.Move) {
	b.lastMove = m
	b.Refresh()
}

// Selected returns the currently highlighted point.
func (b *BoardView) Selected() game.PositionID { return b.selected }

// TargetAt returns the highlighted move that lands on id, if any.
func (b *BoardView) TargetAt(id game.PositionID) (game.Move, bool) {
	m, ok := b.targets[id]
	return m, ok
}

// FlashInvalid briefly marks a point red to show an action was refused.
// It never blocks and never alters the game.
func (b *BoardView) FlashInvalid(id game.PositionID) {
	b.flash = id
	b.Refresh()
	go func() {
		time.Sleep(350 * time.Millisecond)
		// Back onto the UI goroutine: every canvas mutation in Fyne must
		// happen there, and this timer does not run there.
		fyne.Do(func() {
			if b.flash == id {
				b.flash = game.NoPosition
				b.Refresh()
			}
		})
	}()
}

// AnimateMove slides a piece from its origin to its destination, calling done
// when the slide finishes. When animation is disabled it calls done at once,
// so callers need no special case.
func (b *BoardView) AnimateMove(m game.Move, done func()) {
	if !b.animate || m.Kind == game.Placement {
		done()
		return
	}
	if b.animRunning != nil {
		b.animRunning.Stop()
	}

	b.animMove = &m
	b.animPhase = 0
	anim := fyne.NewAnimation(180*time.Millisecond, func(f float32) {
		b.animPhase = f
		b.Refresh()
	})
	anim.Curve = fyne.AnimationEaseInOut
	b.animRunning = anim

	go func() {
		time.Sleep(190 * time.Millisecond)
		fyne.Do(func() {
			b.animMove = nil
			b.animRunning = nil
			b.Refresh()
			done()
		})
	}()
	anim.Start()
}

// Tapped implements fyne.Tappable.
func (b *BoardView) Tapped(ev *fyne.PointEvent) {
	if b.OnPointTapped == nil {
		return
	}
	g := NewGeometry(b.Size())
	if id, ok := g.Nearest(ev.Position); ok {
		b.OnPointTapped(id)
	}
}

// MinSize keeps the board from collapsing in a narrow window.
func (b *BoardView) MinSize() fyne.Size {
	return fyne.NewSize(minBoardSide+60, minBoardSide+60)
}

// CreateRenderer implements fyne.Widget.
func (b *BoardView) CreateRenderer() fyne.WidgetRenderer {
	return newBoardRenderer(b)
}

var _ fyne.Tappable = (*BoardView)(nil)
