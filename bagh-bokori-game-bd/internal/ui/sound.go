package ui

import (
	"sync"

	"fyne.io/fyne/v2"
)

// Effect names a sound cue. The engine knows nothing about these: the
// controller maps game events onto effects, and a SoundPlayer decides what -
// if anything - to do with them.
type Effect int

const (
	// EffectPlace is a goat being set down.
	EffectPlace Effect = iota
	// EffectMove is a piece sliding along a line.
	EffectMove
	// EffectCapture is the tiger eating a goat.
	EffectCapture
	// EffectInvalid is a refused action.
	EffectInvalid
	// EffectVictory is the local player winning.
	EffectVictory
	// EffectDefeat is the local player losing.
	EffectDefeat
)

func (e Effect) String() string {
	switch e {
	case EffectPlace:
		return "place"
	case EffectMove:
		return "move"
	case EffectCapture:
		return "capture"
	case EffectInvalid:
		return "invalid"
	case EffectVictory:
		return "victory"
	default:
		return "defeat"
	}
}

// assetName is the file this effect would play, if the asset is present.
func (e Effect) assetName() string { return "sounds/" + e.String() + ".wav" }

// FyneSound plays bundled sound assets through Fyne.
//
// Sound is deliberately best-effort: assets/sounds is empty in the repository
// because we ship no audio we do not own, so every lookup misses and the
// player silently does nothing. Drop .wav files with the names above into
// assets/sounds and rebuild to hear them.
type FyneSound struct {
	mu      sync.Mutex
	enabled bool
	loader  func(name string) (fyne.Resource, error)
}

// NewFyneSound creates a sound player. loader may be nil, in which case the
// player is a no-op that still records its enabled state.
func NewFyneSound(enabled bool, loader func(name string) (fyne.Resource, error)) *FyneSound {
	return &FyneSound{enabled: enabled, loader: loader}
}

// SetEnabled turns sound on or off.
func (s *FyneSound) SetEnabled(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = on
}

// Enabled reports whether sound is on.
func (s *FyneSound) Enabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enabled
}

// Play implements SoundPlayer. It never blocks and never panics, whatever
// the audio situation on the host turns out to be.
func (s *FyneSound) Play(e Effect) {
	if !s.Enabled() || s.loader == nil {
		return
	}
	go func() {
		defer func() { _ = recover() }() // a missing audio device must never crash a board game
		res, err := s.loader(e.assetName())
		if err != nil || res == nil {
			return
		}
		playResource(res)
	}()
}

var _ SoundPlayer = (*FyneSound)(nil)
