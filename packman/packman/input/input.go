package input

import (
	"fyne.io/fyne/v2"
)

// Direction represents input direction
type Direction int

const (
	DirNone Direction = iota
	DirUp
	DirDown
	DirLeft
	DirRight
)

// InputManager handles keyboard and touch input
type InputManager struct {
	window fyne.Window

	// Current direction state
	currentDirection Direction
	nextDirection    Direction

	// Key states
	keyStates map[fyne.KeyName]bool
}

// NewInputManager creates a new input manager
func NewInputManager(window fyne.Window) *InputManager {
	im := &InputManager{
		window:           window,
		currentDirection: DirNone,
		nextDirection:    DirNone,
		keyStates:        make(map[fyne.KeyName]bool),
	}

	return im
}

// SetupKeyboardHandlers sets up keyboard event handlers
func (im *InputManager) SetupKeyboardHandlers() {
	// Keyboard handling is done in main.go via window.Canvas()
	// This method is kept for interface compatibility
}

// handleKeyDown handles key down events
func (im *InputManager) handleKeyDown(key *fyne.KeyEvent) {
	im.keyStates[key.Name] = true

	switch key.Name {
	case fyne.KeyUp, fyne.KeyW:
		im.nextDirection = DirUp
	case fyne.KeyDown, fyne.KeyS:
		im.nextDirection = DirDown
	case fyne.KeyLeft, fyne.KeyA:
		im.nextDirection = DirLeft
	case fyne.KeyRight, fyne.KeyD:
		im.nextDirection = DirRight
	case fyne.KeyP, fyne.KeyEscape:
		// Pause toggle - handled by game
	case fyne.KeySpace:
		// Start/confirm - handled by game
	case fyne.KeyReturn:
		// Start/confirm - handled by game
	}
}

// handleKeyUp handles key up events
func (im *InputManager) handleKeyUp(key *fyne.KeyEvent) {
	im.keyStates[key.Name] = false
}

// GetDirection returns the current input direction
func (im *InputManager) GetDirection() Direction {
	return im.currentDirection
}

// GetNextDirection returns the next requested direction
func (im *InputManager) GetNextDirection() Direction {
	return im.nextDirection
}

// SetDirection sets the current direction (after successful move)
func (im *InputManager) SetDirection(dir Direction) {
	im.currentDirection = dir
}

// ConsumeNextDirection returns and clears the next direction request
func (im *InputManager) ConsumeNextDirection() Direction {
	dir := im.nextDirection
	// Don't clear - keep buffering until successful move
	return dir
}

// ClearNextDirection clears the buffered direction
func (im *InputManager) ClearNextDirection() {
	im.nextDirection = DirNone
}

// IsKeyPressed returns whether a specific key is currently pressed
func (im *InputManager) IsKeyPressed(key fyne.KeyName) bool {
	return im.keyStates[key]
}

// HandleDirectionInput sets the next direction from external input (touch/mouse)
func (im *InputManager) HandleDirectionInput(dir Direction) {
	im.nextDirection = dir
}
