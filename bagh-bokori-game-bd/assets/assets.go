// Package assets embeds the game's artwork into the binary.
//
// Embedding matters for two reasons: a desktop build is a single file with no
// stray image directory to lose, and a WebAssembly build has no filesystem to
// read from at all. The art is SVG, so the pieces stay sharp at any board
// size and at any display scaling.
package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed icons/tiger.svg
var tigerSVG []byte

//go:embed icons/goat.svg
var goatSVG []byte

// The Bangla text in this game - বাঘবন্দী, বাঘ, ছাগল - needs a font with
// Bengali glyphs. On the desktop Fyne can borrow one from the system, but a
// WebAssembly build has no system fonts at all, so Bangla renders as blanks
// unless the font travels with the binary.
//
// Noto Sans Bengali carries no Latin letters, so it does not replace the UI
// font: Fyne resolves faces per rune, taking Bengali from this font and
// everything else from its own default. See assets/fonts/LICENSE-OFL.txt.

//go:embed fonts/NotoSansBengali-Regular.ttf
var bengaliRegular []byte

//go:embed fonts/NotoSansBengali-Bold.ttf
var bengaliBold []byte

// Tiger returns the বাঘ artwork.
func Tiger() fyne.Resource { return fyne.NewStaticResource("tiger.svg", tigerSVG) }

// Goat returns the ছাগল artwork.
func Goat() fyne.Resource { return fyne.NewStaticResource("goat.svg", goatSVG) }

// BengaliRegular returns the Bangla text font.
func BengaliRegular() fyne.Resource {
	return fyne.NewStaticResource("NotoSansBengali-Regular.ttf", bengaliRegular)
}

// BengaliBold returns the bold Bangla text font.
func BengaliBold() fyne.Resource {
	return fyne.NewStaticResource("NotoSansBengali-Bold.ttf", bengaliBold)
}
