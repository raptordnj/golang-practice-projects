// Package ui is the Fyne desktop frontend.
//
// It owns every pixel and every widget, and it owns no game rules at all: it
// asks internal/game what is legal and what happened, and only draws the
// answer. internal/game does not import this package, and never may.
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"bagh-bandi/assets"
)

// BoardTheme is a named colour scheme for the board and its pieces.
type BoardTheme struct {
	Name string
	// Bangla name shown in the settings dialog.
	Bangla string

	Background color.Color
	BoardFill  color.Color
	BoardEdge  color.Color
	Line       color.Color
	PointFill  color.Color

	// PieceBacking is the disc drawn behind a piece's artwork, and
	// PieceOutline its rim. They keep the tiger and the goats legible on
	// every board colour.
	PieceBacking color.Color
	PieceOutline color.Color

	TigerBody   color.Color
	TigerStripe color.Color
	GoatBody    color.Color
	GoatDetail  color.Color

	Selection   color.Color
	LegalMove   color.Color
	CaptureHint color.Color
	LastMove    color.Color
	Danger      color.Color
}

// BoardThemes are the themes offered in settings. The default, "Terracotta",
// is drawn from the red clay and alpona motifs of rural Bangladesh, where
// বাঘবন্দী boards are traditionally scratched into the earth or a courtyard
// floor.
var BoardThemes = []BoardTheme{
	{
		Name:         "Terracotta",
		Bangla:       "পোড়ামাটি",
		Background:   color.NRGBA{R: 0x2A, G: 0x1C, B: 0x14, A: 0xFF},
		BoardFill:    color.NRGBA{R: 0xC4, G: 0x7A, B: 0x4A, A: 0xFF},
		BoardEdge:    color.NRGBA{R: 0x7A, G: 0x40, B: 0x22, A: 0xFF},
		Line:         color.NRGBA{R: 0x3E, G: 0x22, B: 0x12, A: 0xFF},
		PointFill:    color.NRGBA{R: 0xF3, G: 0xE3, B: 0xCE, A: 0xFF},
		PieceBacking: color.NRGBA{R: 0xF7, G: 0xEB, B: 0xD8, A: 0xFF},
		PieceOutline: color.NRGBA{R: 0x5A, G: 0x2E, B: 0x16, A: 0xFF},
		TigerBody:    color.NRGBA{R: 0xE8, G: 0x8A, B: 0x1A, A: 0xFF},
		TigerStripe:  color.NRGBA{R: 0x2B, G: 0x1A, B: 0x0C, A: 0xFF},
		GoatBody:     color.NRGBA{R: 0xF6, G: 0xF2, B: 0xE8, A: 0xFF},
		GoatDetail:   color.NRGBA{R: 0x6B, G: 0x5B, B: 0x49, A: 0xFF},
		Selection:    color.NRGBA{R: 0x2E, G: 0x8B, B: 0x57, A: 0xFF},
		LegalMove:    color.NRGBA{R: 0x2E, G: 0x8B, B: 0x57, A: 0xAA},
		CaptureHint:  color.NRGBA{R: 0xD2, G: 0x2B, B: 0x2B, A: 0xCC},
		LastMove:     color.NRGBA{R: 0xFF, G: 0xD7, B: 0x3A, A: 0xCC},
		Danger:       color.NRGBA{R: 0xD2, G: 0x2B, B: 0x2B, A: 0xFF},
	},
	{
		Name:         "Shapla Night",
		Bangla:       "শাপলা রাত",
		Background:   color.NRGBA{R: 0x0E, G: 0x14, B: 0x1B, A: 0xFF},
		BoardFill:    color.NRGBA{R: 0x1B, G: 0x2A, B: 0x38, A: 0xFF},
		BoardEdge:    color.NRGBA{R: 0x0A, G: 0x10, B: 0x16, A: 0xFF},
		Line:         color.NRGBA{R: 0x6E, G: 0x91, B: 0xAE, A: 0xFF},
		PointFill:    color.NRGBA{R: 0xD8, G: 0xE6, B: 0xF2, A: 0xFF},
		PieceBacking: color.NRGBA{R: 0xE6, G: 0xEF, B: 0xF8, A: 0xFF},
		PieceOutline: color.NRGBA{R: 0x0A, G: 0x10, B: 0x16, A: 0xFF},
		TigerBody:    color.NRGBA{R: 0xFF, G: 0xA5, B: 0x2E, A: 0xFF},
		TigerStripe:  color.NRGBA{R: 0x1A, G: 0x0F, B: 0x06, A: 0xFF},
		GoatBody:     color.NRGBA{R: 0xEC, G: 0xF3, B: 0xFA, A: 0xFF},
		GoatDetail:   color.NRGBA{R: 0x4A, G: 0x5A, B: 0x6A, A: 0xFF},
		Selection:    color.NRGBA{R: 0x3F, G: 0xD1, B: 0x8B, A: 0xFF},
		LegalMove:    color.NRGBA{R: 0x3F, G: 0xD1, B: 0x8B, A: 0xAA},
		CaptureHint:  color.NRGBA{R: 0xFF, G: 0x5C, B: 0x5C, A: 0xCC},
		LastMove:     color.NRGBA{R: 0xFF, G: 0xE0, B: 0x6A, A: 0xCC},
		Danger:       color.NRGBA{R: 0xFF, G: 0x5C, B: 0x5C, A: 0xFF},
	},
	{
		Name:         "Jamdani Green",
		Bangla:       "জামদানি সবুজ",
		Background:   color.NRGBA{R: 0x10, G: 0x22, B: 0x1A, A: 0xFF},
		BoardFill:    color.NRGBA{R: 0x1E, G: 0x4D, B: 0x35, A: 0xFF},
		BoardEdge:    color.NRGBA{R: 0x0C, G: 0x2A, B: 0x1B, A: 0xFF},
		Line:         color.NRGBA{R: 0xD9, G: 0xC8, B: 0x8A, A: 0xFF},
		PointFill:    color.NRGBA{R: 0xF5, G: 0xEF, B: 0xD9, A: 0xFF},
		PieceBacking: color.NRGBA{R: 0xF7, G: 0xF1, B: 0xDE, A: 0xFF},
		PieceOutline: color.NRGBA{R: 0x0C, G: 0x2A, B: 0x1B, A: 0xFF},
		TigerBody:    color.NRGBA{R: 0xF0, G: 0x93, B: 0x21, A: 0xFF},
		TigerStripe:  color.NRGBA{R: 0x24, G: 0x16, B: 0x08, A: 0xFF},
		GoatBody:     color.NRGBA{R: 0xFA, G: 0xF7, B: 0xEE, A: 0xFF},
		GoatDetail:   color.NRGBA{R: 0x5F, G: 0x5A, B: 0x4A, A: 0xFF},
		Selection:    color.NRGBA{R: 0xF2, G: 0xD0, B: 0x4E, A: 0xFF},
		LegalMove:    color.NRGBA{R: 0xF2, G: 0xD0, B: 0x4E, A: 0xAA},
		CaptureHint:  color.NRGBA{R: 0xE2, G: 0x45, B: 0x3C, A: 0xCC},
		LastMove:     color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x99},
		Danger:       color.NRGBA{R: 0xE2, G: 0x45, B: 0x3C, A: 0xFF},
	},
}

// ThemeByName looks up a board theme, falling back to the first one.
func ThemeByName(name string) BoardTheme {
	for _, t := range BoardThemes {
		if t.Name == name {
			return t
		}
	}
	return BoardThemes[0]
}

// appTheme is the Fyne widget theme. It keeps the surrounding chrome dark
// and warm so the board is the brightest thing on screen.
type appTheme struct {
	fyne.Theme
	board BoardTheme
}

func newAppTheme(b BoardTheme) fyne.Theme {
	return &appTheme{Theme: theme.DefaultTheme(), board: b}
}

// Font supplies the Bangla-capable face.
//
// Fyne builds a per-rune face chain with the theme's font first and its own
// default behind it. Noto Sans Bengali has no Latin letters, so returning it
// here does not change the look of the English text - it only means that
// বাঘবন্দী, বাঘ and ছাগল have glyphs to render with. That matters most in the
// WebAssembly build, which has no system fonts to fall back on.
func (t *appTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return t.Theme.Font(style)
	}
	if style.Bold {
		return assets.BengaliBold()
	}
	return assets.BengaliRegular()
}

func (t *appTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return t.board.Background
	case theme.ColorNameButton, theme.ColorNameInputBackground:
		return t.board.BoardEdge
	case theme.ColorNamePrimary, theme.ColorNameFocus:
		return t.board.Selection
	case theme.ColorNameError:
		return t.board.Danger
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xF2, G: 0xEA, B: 0xDD, A: 0xFF}
	}
	return t.Theme.Color(name, theme.VariantDark)
}
