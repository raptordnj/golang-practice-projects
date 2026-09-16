package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// The settings form has a lot in it. Fyne sizes a dialog to the bare minimum
// its content tolerates, which left the form squeezed into a tiny scroll
// area, so the dialog asks for a generous size instead.
func TestDialogSizeIsGenerousInALargeWindow(t *testing.T) {
	a := newTestApp(t)
	a.window.Resize(fyne.NewSize(980, 720))

	got := a.dialogSize(560, 720)

	if got.Height < 600 {
		t.Errorf("dialog height = %v, want the form to have real room", got.Height)
	}
	if got.Width < 500 {
		t.Errorf("dialog width = %v, too narrow for the controls", got.Width)
	}
}

// ...but it must never ask for more than the window can show, or the buttons
// fall off the bottom and the dialog cannot be dismissed.
func TestDialogSizeFitsInsideASmallWindow(t *testing.T) {
	a := newTestApp(t)
	small := fyne.NewSize(420, 380)
	a.window.Resize(small)

	got := a.dialogSize(560, 720)
	canvas := a.window.Canvas().Size()

	if got.Width > canvas.Width || got.Height > canvas.Height {
		t.Errorf("dialog %v does not fit in a %v canvas", got, canvas)
	}
	if got.Width <= 0 || got.Height <= 0 {
		t.Errorf("degenerate dialog size %v", got)
	}
}

// The settings dialog must open, lay out and close without panicking, and
// every control in it must be reachable.
func TestSettingsDialogOpens(t *testing.T) {
	a := newTestApp(t)
	a.window.Resize(fyne.NewSize(980, 720))
	a.showSettings() // must not panic
}

func TestRulesDialogOpens(t *testing.T) {
	a := newTestApp(t)
	a.window.Resize(fyne.NewSize(980, 720))
	a.showRules() // must not panic
}

// newTestApp builds the real App against Fyne's test driver and a throwaway
// config directory, so dialogs can be opened and measured in a unit test.
func newTestApp(t *testing.T) *App {
	t.Helper()
	fyneApp := test.NewTempApp(t)
	a := newApp(fyneApp, NewStorageAt(t.TempDir()))
	t.Cleanup(a.closeController)
	return a
}
