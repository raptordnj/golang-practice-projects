package ui

import "fyne.io/fyne/v2"

// playResource is the single point where audio would actually be emitted.
//
// Fyne has no audio API of its own, so this is a seam rather than an
// implementation: wire in oto, beep or a platform player here and every sound
// cue in the game starts working, with no change anywhere else. Keeping the
// seam empty is what lets internal/game stay free of audio dependencies and
// lets the whole program build for WebAssembly unchanged.
func playResource(res fyne.Resource) {
	_ = res
}
