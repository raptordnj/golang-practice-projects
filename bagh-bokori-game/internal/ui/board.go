package ui

import (
	"image/color"

	"bagh-bakri/internal/game"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	cellSize  = 60.0
	boardLeft = 50.0
	boardTop  = 50.0
)

type BoardUI struct {
	state       *game.GameState
	onMove      func(game.Move)
	selected    game.PositionID
	selectedSet bool
}

func NewBoardUI(state *game.GameState, onMove func(game.Move)) *BoardUI {
	return &BoardUI{
		state:  state,
		onMove: onMove,
	}
}

func (b *BoardUI) CreateUI() fyne.CanvasObject {
	return b.buildBoard()
}

func (b *BoardUI) buildBoard() fyne.CanvasObject {
	content := container.NewWithoutLayout()

	// Draw board lines
	for pos := 0; pos < game.BoardSize; pos++ {
		x, y := b.posToCoords(game.PositionID(pos))
		// Draw node
		node := canvas.NewCircle(color.NRGBA{R: 200, G: 200, B: 200, A: 255})
		node.Resize(fyne.NewSize(8, 8))
		node.Move(fyne.NewPos(x-4, y-4))
		content.Add(node)

		// Draw lines to neighbors (only right and down to avoid duplicates)
		for _, n := range game.BoardTopology[game.PositionID(pos)] {
			if n > game.PositionID(pos) {
				nx, ny := b.posToCoords(n)
				line := canvas.NewLine(color.NRGBA{R: 100, G: 80, B: 60, A: 255})
				line.Position1 = fyne.NewPos(x, y)
				line.Position2 = fyne.NewPos(nx, ny)
				line.StrokeWidth = 2
				content.Add(line)
			}
		}
	}

	// Draw pieces
	for pos := 0; pos < game.BoardSize; pos++ {
		piece := b.state.Board.At(game.PositionID(pos))
		if piece.Type != game.Empty {
			x, y := b.posToCoords(game.PositionID(pos))
			var c color.Color
			if piece.Player == game.PlayerTiger {
				c = color.NRGBA{R: 30, G: 30, B: 30, A: 255}
			} else {
				c = color.NRGBA{R: 240, G: 240, B: 200, A: 255}
			}
			circle := canvas.NewCircle(c)
			circle.Resize(fyne.NewSize(36, 36))
			circle.Move(fyne.NewPos(x-18, y-18))
			content.Add(circle)
		}
	}

	// Highlight selected position
	if b.selectedSet {
		x, y := b.posToCoords(b.selected)
		highlight := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, A: 128, B: 0})
		highlight.Resize(fyne.NewSize(44, 44))
		highlight.Move(fyne.NewPos(x-22, y-22))
		content.Add(highlight)
	}

	// Add click targets using buttons
	b.addClickTargets(content)

	return content
}

func (b *BoardUI) posToCoords(pos game.PositionID) (float32, float32) {
	row := int(pos) / 5
	col := int(pos) % 5
	return boardLeft + float32(col)*cellSize, boardTop + float32(row)*cellSize
}

func (b *BoardUI) addClickTargets(content *fyne.Container) {
	for pos := 0; pos < game.BoardSize; pos++ {
		x, y := b.posToCoords(game.PositionID(pos))
		btn := widget.NewButton("", func() {
			b.handleClick(game.PositionID(pos))
		})
		btn.Importance = widget.LowImportance
		btn.Resize(fyne.NewSize(cellSize, cellSize))
		btn.Move(fyne.NewPos(x-cellSize/2, y-cellSize/2))
		content.Add(btn)
	}
}

func (b *BoardUI) handleClick(pos game.PositionID) {
	if b.selectedSet && b.selected != pos {
		// Try move
		moves := game.LegalMoves(*b.state)
		for _, m := range moves {
			if m.From == b.selected && m.To == pos {
				if b.onMove != nil {
					b.onMove(m)
				}
				b.selectedSet = false
				return
			}
		}
		// Deselect if invalid
		b.selectedSet = false
		return
	}
	// Select a tiger
	piece := b.state.Board.At(pos)
	if piece.Type == game.Tiger && piece.Player == game.PlayerTiger && b.state.Turn == game.PlayerTiger {
		b.selected = pos
		b.selectedSet = true
	} else if b.state.Turn == game.PlayerGoat && b.state.GoatsPlaced < 20 {
		// Goat placement
		moves := game.LegalMoves(*b.state)
		for _, m := range moves {
			if m.To == pos {
				if b.onMove != nil {
					b.onMove(m)
				}
				return
			}
		}
	} else if b.state.Turn == game.PlayerGoat && b.state.GoatsPlaced >= 20 {
		// Goat movement
		piece := b.state.Board.At(pos)
		if piece.Type == game.Goat {
			b.selected = pos
			b.selectedSet = true
		}
	}
}

func (b *BoardUI) SetState(state *game.GameState) {
	b.state = state
	b.selectedSet = false
}

func (b *BoardUI) SetSelected(pos game.PositionID) {
	b.selected = pos
	b.selectedSet = true
}

func (b *BoardUI) ClearSelected() {
	b.selectedSet = false
}
