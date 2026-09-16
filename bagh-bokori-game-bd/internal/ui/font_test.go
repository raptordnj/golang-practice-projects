package ui

import (
	"image"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"golang.org/x/image/font/sfnt"

	"bagh-bandi/assets"
)

// The game's own Bangla strings. If any of these cannot be drawn, players see
// blank boxes - which is exactly what happens in a WebAssembly build that has
// no bundled Bengali font.
var banglaStrings = []string{
	"বাঘবন্দী", // the game's name
	"বাঘ",      // tiger
	"ছাগল",     // goat
	"সহজ",      // easy
	"মাঝারি",   // medium
	"কঠিন",     // hard
	"খালি",     // empty
}

// The embedded font must actually contain the Bengali block.
func TestBundledFontCoversBengali(t *testing.T) {
	for _, res := range []fyne.Resource{assets.BengaliRegular(), assets.BengaliBold()} {
		f, err := sfnt.Parse(res.Content())
		if err != nil {
			t.Fatalf("%s: not a parsable font: %v", res.Name(), err)
		}

		var buf sfnt.Buffer
		for _, s := range banglaStrings {
			for _, r := range s {
				idx, err := f.GlyphIndex(&buf, r)
				if err != nil || idx == 0 {
					t.Errorf("%s: no glyph for %q (U+%04X)", res.Name(), r, r)
				}
			}
		}
	}
}

// The theme must hand that font to Fyne, for both weights.
func TestThemeSuppliesTheBengaliFont(t *testing.T) {
	th := newAppTheme(BoardThemes[0])

	regular := th.Font(fyne.TextStyle{})
	if regular.Name() != assets.BengaliRegular().Name() {
		t.Errorf("regular font = %q, want the bundled Bangla font", regular.Name())
	}

	bold := th.Font(fyne.TextStyle{Bold: true})
	if bold.Name() != assets.BengaliBold().Name() {
		t.Errorf("bold font = %q, want the bundled bold Bangla font", bold.Name())
	}

	// Monospace has no Bangla in this UI and should stay as Fyne's own.
	mono := th.Font(fyne.TextStyle{Monospace: true})
	if mono.Name() == assets.BengaliRegular().Name() {
		t.Error("monospace should keep Fyne's default font")
	}
}

// Smoke test: render the Bangla text through Fyne and confirm ink lands on the
// canvas.
//
// Note what this does and does not prove. On a desktop with Bengali system
// fonts installed, Fyne finds glyphs through fontscan whether or not we bundle
// a font, so this test passes either way here - it is guarding against a
// broken or unreadable font resource, not against the WebAssembly problem.
// The tests that actually pin the fix down are TestBundledFontCoversBengali
// (the embedded font really contains these glyphs) and
// TestThemeSuppliesTheBengaliFont (the theme really hands it to Fyne). In a
// browser there is no fontscan and no system font, so those two are what
// decides whether Bangla renders.
func TestBanglaTextActuallyRenders(t *testing.T) {
	test.NewTempApp(t).Settings().SetTheme(newAppTheme(BoardThemes[0]))

	for _, s := range banglaStrings {
		text := canvas.NewText(s, colorBlack())
		text.TextSize = 42

		w := test.NewWindow(text)
		w.Resize(fyne.NewSize(400, 90))
		img := w.Canvas().Capture()
		inked := countInkedPixels(img)
		w.Close()

		if inked < 40 {
			t.Errorf("%q rendered only %d inked pixels - the Bangla glyphs are missing", s, inked)
		}
	}
}

// Latin text must still render. The Bangla font has no Latin letters at all,
// so if Fyne's per-rune fallback to its default font were not working, making
// Noto Sans Bengali the theme font would have wiped out every English label in
// the game.
func TestLatinTextStillRenders(t *testing.T) {
	test.NewTempApp(t).Settings().SetTheme(newAppTheme(BoardThemes[0]))

	text := canvas.NewText("Bagh-Bandi: Tiger wins", colorBlack())
	text.TextSize = 42

	w := test.NewWindow(text)
	t.Cleanup(w.Close)
	w.Resize(fyne.NewSize(600, 90))

	if inked := countInkedPixels(w.Canvas().Capture()); inked < 40 {
		t.Errorf("Latin text rendered only %d inked pixels", inked)
	}
}

// Different Bangla words must produce different pixels - all-identical output
// would mean every glyph fell back to the same "missing glyph" box.
func TestBanglaWordsRenderDistinctly(t *testing.T) {
	test.NewTempApp(t).Settings().SetTheme(newAppTheme(BoardThemes[0]))

	render := func(s string) image.Image {
		text := canvas.NewText(s, colorBlack())
		text.TextSize = 42
		w := test.NewWindow(text)
		t.Cleanup(w.Close)
		w.Resize(fyne.NewSize(400, 90))
		return w.Canvas().Capture()
	}

	tiger, goat := render("বাঘ"), render("ছাগল")
	if countDifferingPixels(tiger, goat) < 50 {
		t.Error("বাঘ and ছাগল render identically - glyphs are probably missing boxes")
	}
}

func countInkedPixels(img image.Image) int {
	bounds := img.Bounds()
	// The background is whatever the theme paints; count pixels that differ
	// from the top-left corner, which no glyph reaches.
	bg := img.At(bounds.Min.X, bounds.Min.Y)
	n := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.At(x, y) != bg {
				n++
			}
		}
	}
	return n
}

func colorBlack() color.Color { return color.NRGBA{A: 0xFF} }
