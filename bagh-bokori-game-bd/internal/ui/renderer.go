package ui

import (
	"image/color"

	"bagh-bandi/assets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"bagh-bandi/internal/game"
)

// boardRenderer draws the board. It creates its canvas objects once and only
// ever repositions and recolours them, so a redraw allocates nothing and the
// GPU keeps the same textures - that is what makes resizing smooth.
//
// Fyne composites all of these through its hardware driver: OpenGL on Linux,
// macOS and Windows, and WebGL when the same code is built for
// GOOS=js GOARCH=wasm. See README.md ("Rendering and WebGL").
type boardRenderer struct {
	view *BoardView

	backdrop *canvas.Rectangle
	panel    *canvas.Rectangle
	border   *canvas.Rectangle

	lines  []*canvas.Line
	points []*canvas.Circle

	// Per-point overlays, indexed by PositionID.
	lastMoveRings []*canvas.Circle
	selectRings   []*canvas.Circle
	moveDots      []*canvas.Circle
	captureMarks  []*canvas.Line

	// Pieces. Each point owns a shadow disc and one image for each animal;
	// unused ones are simply hidden. Reusing the images means a redraw never
	// re-rasterises the SVGs.
	shadows []*canvas.Circle
	tigers  []*canvas.Image
	goats   []*canvas.Image

	objects []fyne.CanvasObject
}

func newBoardRenderer(v *BoardView) *boardRenderer {
	r := &boardRenderer{view: v}

	r.backdrop = canvas.NewRectangle(color.Transparent)
	r.panel = canvas.NewRectangle(color.Transparent)
	r.border = canvas.NewRectangle(color.Transparent)
	r.border.StrokeWidth = 3
	r.border.FillColor = color.Transparent

	board := v.board
	for from := game.PositionID(0); int(from) < game.PositionCount; from++ {
		for _, to := range board.Neighbors(from) {
			if to > from { // draw each edge once
				r.lines = append(r.lines, canvas.NewLine(color.Transparent))
			}
		}
	}

	n := game.PositionCount
	r.points = makeCircles(n)
	r.lastMoveRings = makeCircles(n)
	r.selectRings = makeCircles(n)
	r.moveDots = makeCircles(n)
	r.shadows = makeCircles(n)
	r.tigers = makeImages(n, assets.Tiger())
	r.goats = makeImages(n, assets.Goat())
	r.captureMarks = makeLines(2 * n)

	r.objects = append(r.objects, r.backdrop, r.panel, r.border)
	for _, l := range r.lines {
		r.objects = append(r.objects, l)
	}
	appendAll(&r.objects, r.lastMoveRings, r.points, r.selectRings, r.moveDots, r.shadows)
	appendLines(&r.objects, r.captureMarks)
	appendImages(&r.objects, r.tigers, r.goats)

	return r
}

func makeCircles(n int) []*canvas.Circle {
	out := make([]*canvas.Circle, n)
	for i := range out {
		out[i] = canvas.NewCircle(color.Transparent)
	}
	return out
}

// makeImages creates n copies of the same artwork. Each board point owns its
// own image object so a redraw only moves and resizes existing ones.
func makeImages(n int, res fyne.Resource) []*canvas.Image {
	out := make([]*canvas.Image, n)
	for i := range out {
		img := canvas.NewImageFromResource(res)
		img.FillMode = canvas.ImageFillContain
		img.ScaleMode = canvas.ImageScaleSmooth
		img.Hide()
		out[i] = img
	}
	return out
}

func makeLines(n int) []*canvas.Line {
	out := make([]*canvas.Line, n)
	for i := range out {
		out[i] = canvas.NewLine(color.Transparent)
	}
	return out
}

func appendAll(dst *[]fyne.CanvasObject, groups ...[]*canvas.Circle) {
	for _, g := range groups {
		for _, c := range g {
			*dst = append(*dst, c)
		}
	}
}

func appendLines(dst *[]fyne.CanvasObject, groups ...[]*canvas.Line) {
	for _, g := range groups {
		for _, l := range g {
			*dst = append(*dst, l)
		}
	}
}

func appendImages(dst *[]fyne.CanvasObject, groups ...[]*canvas.Image) {
	for _, g := range groups {
		for _, img := range g {
			*dst = append(*dst, img)
		}
	}
}

// Objects implements fyne.WidgetRenderer.
func (r *boardRenderer) Objects() []fyne.CanvasObject { return r.objects }

// Layout implements fyne.WidgetRenderer.
func (r *boardRenderer) Layout(size fyne.Size) { r.draw(size) }

// Refresh implements fyne.WidgetRenderer.
func (r *boardRenderer) Refresh() {
	r.draw(r.view.Size())
	canvas.Refresh(r.view)
}

// MinSize implements fyne.WidgetRenderer.
func (r *boardRenderer) MinSize() fyne.Size { return r.view.MinSize() }

// Destroy implements fyne.WidgetRenderer.
func (r *boardRenderer) Destroy() {}

// draw is the whole drawing routine, recomputed from the current size so the
// board is correct at any window dimensions.
func (r *boardRenderer) draw(size fyne.Size) {
	g := NewGeometry(size)
	t := r.view.theme
	s := r.view.state

	r.backdrop.FillColor = t.Background
	r.backdrop.Resize(size)
	r.backdrop.Move(fyne.NewPos(0, 0))

	panelPos, panelSize := g.Panel()
	for _, rect := range []*canvas.Rectangle{r.panel, r.border} {
		rect.Move(panelPos)
		rect.Resize(panelSize)
		rect.CornerRadius = g.Margin * 0.5
	}
	r.panel.FillColor = t.BoardFill
	r.border.StrokeColor = t.BoardEdge

	r.drawLines(g, t)
	r.drawPoints(g, t)
	r.drawHighlights(g, t, &s)
	r.drawPieces(g, t, &s)
}

func (r *boardRenderer) drawLines(g Geometry, t BoardTheme) {
	board := r.view.board
	i := 0
	width := g.Cell * 0.045
	if width < 1 {
		width = 1
	}
	for from := game.PositionID(0); int(from) < game.PositionCount; from++ {
		for _, to := range board.Neighbors(from) {
			if to <= from {
				continue
			}
			l := r.lines[i]
			i++
			l.StrokeColor = t.Line
			l.StrokeWidth = width
			l.Position1 = g.Center(from)
			l.Position2 = g.Center(to)
		}
	}
}

func (r *boardRenderer) drawPoints(g Geometry, t BoardTheme) {
	for id := 0; id < game.PositionCount; id++ {
		c := r.points[id]
		c.FillColor = t.PointFill
		place(c, g.Center(game.PositionID(id)), g.Point)
	}
}

func (r *boardRenderer) drawHighlights(g Geometry, t BoardTheme, s *game.GameState) {
	v := r.view

	for id := game.PositionID(0); int(id) < game.PositionCount; id++ {
		center := g.Center(id)

		// Last move: a soft ring on the origin and destination.
		ring := r.lastMoveRings[id]
		if v.lastMove != nil && (v.lastMove.To == id || v.lastMove.From == id) {
			ring.FillColor = color.Transparent
			ring.StrokeColor = t.LastMove
			ring.StrokeWidth = g.Cell * 0.06
			place(ring, center, g.Piece*1.18)
		} else {
			hide(ring)
		}

		// Selection ring, or a red flash when an action was refused.
		sel := r.selectRings[id]
		switch {
		case v.flash == id:
			sel.FillColor = color.Transparent
			sel.StrokeColor = t.Danger
			sel.StrokeWidth = g.Cell * 0.09
			place(sel, center, g.Piece*1.3)
		case v.selected == id:
			sel.FillColor = color.Transparent
			sel.StrokeColor = t.Selection
			sel.StrokeWidth = g.Cell * 0.09
			place(sel, center, g.Piece*1.3)
		default:
			hide(sel)
		}

		// Legal destinations: a dot for a quiet move, a red cross over the
		// goat that would be eaten for a capture.
		dot := r.moveDots[id]
		markA, markB := r.captureMarks[2*int(id)], r.captureMarks[2*int(id)+1]
		hideLine(markA)
		hideLine(markB)

		if m, ok := v.targets[id]; ok {
			if m.Kind == game.Capture {
				dot.FillColor = t.CaptureHint
				place(dot, center, g.Piece*0.55)
				drawCross(markA, markB, g.Center(m.Over), g.Piece*0.7, t.CaptureHint, g.Cell*0.07)
			} else {
				dot.FillColor = t.LegalMove
				place(dot, center, g.Piece*0.42)
			}
		} else {
			hide(dot)
		}
	}
	_ = s
}

func (r *boardRenderer) drawPieces(g Geometry, t BoardTheme, s *game.GameState) {
	v := r.view

	// While a slide is animating, the moving piece is drawn along the path
	// instead of sitting at its destination.
	animFrom, animTo := game.NoPosition, game.NoPosition
	if v.animMove != nil {
		animFrom, animTo = v.animMove.From, v.animMove.To
	}

	for id := game.PositionID(0); int(id) < game.PositionCount; id++ {
		piece := s.Occupancy[id]
		center := g.Center(id)
		if id == animTo && piece != game.Empty {
			center = Lerp(g.Center(animFrom), g.Center(animTo), v.animPhase)
		}

		tiger, goat, shadow := r.tigers[id], r.goats[id], r.shadows[id]

		if piece == game.Empty {
			tiger.Hide()
			goat.Hide()
			hide(shadow)
			continue
		}

		// A themed disc behind the animal keeps it legible on every board
		// colour and gives the piece some weight on the board.
		radius := g.Piece
		shadow.FillColor = t.PieceBacking
		shadow.StrokeColor = t.PieceOutline
		shadow.StrokeWidth = g.Cell * 0.04
		place(shadow, center, radius)

		art, other := tiger, goat
		if piece == game.Goat {
			art, other = goat, tiger
			radius = g.Piece * 0.92
		}
		other.Hide()
		showImage(art, center, radius*0.94)
	}
}

// showImage centres an image on a point, sized to fit inside the disc.
func showImage(img *canvas.Image, center fyne.Position, radius float32) {
	img.Move(topLeft(center, radius))
	img.Resize(squareSize(radius))
	img.Show()
}

// ---- small drawing helpers ----

func place(c *canvas.Circle, center fyne.Position, radius float32) {
	c.Move(topLeft(center, radius))
	c.Resize(squareSize(radius))
	c.Show()
}

func hide(c *canvas.Circle) {
	c.FillColor = color.Transparent
	c.StrokeWidth = 0
	c.Hide()
}

func hideLine(l *canvas.Line) {
	l.StrokeColor = color.Transparent
	l.StrokeWidth = 0
	l.Hide()
}

func stripe(l *canvas.Line, a, b fyne.Position, col color.Color, width float32) {
	l.StrokeColor = col
	l.StrokeWidth = width
	l.Position1 = a
	l.Position2 = b
	l.Show()
}

func drawCross(a, b *canvas.Line, center fyne.Position, size float32, col color.Color, width float32) {
	stripe(a, fyne.NewPos(center.X-size, center.Y-size), fyne.NewPos(center.X+size, center.Y+size), col, width)
	stripe(b, fyne.NewPos(center.X-size, center.Y+size), fyne.NewPos(center.X+size, center.Y-size), col, width)
}

var _ fyne.WidgetRenderer = (*boardRenderer)(nil)
