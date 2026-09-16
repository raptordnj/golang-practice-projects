package game

import "fmt"

// MoveKind distinguishes the three legal actions in Bagh-Bandi.
type MoveKind uint8

const (
	// Placement drops a goat onto an empty point during the placement phase.
	Placement MoveKind = iota + 1
	// Step slides a piece along a line to an adjacent empty point.
	Step
	// Capture jumps the tiger over one goat onto the empty point beyond.
	Capture
)

func (k MoveKind) String() string {
	switch k {
	case Placement:
		return "place"
	case Step:
		return "move"
	case Capture:
		return "capture"
	default:
		return "unknown"
	}
}

// Move is a single action. It is a small comparable value type, so moves can
// be compared with == and stored in maps.
//
// For Placement, From is NoPosition. For Capture, Over holds the captured
// goat; it is NoPosition otherwise.
type Move struct {
	Kind MoveKind
	From PositionID
	To   PositionID
	Over PositionID
}

// PlaceMove builds a goat placement.
func PlaceMove(to PositionID) Move {
	return Move{Kind: Placement, From: NoPosition, To: to, Over: NoPosition}
}

// StepMove builds an ordinary slide along a line.
func StepMove(from, to PositionID) Move {
	return Move{Kind: Step, From: from, To: to, Over: NoPosition}
}

// CaptureMove builds a tiger jump capturing the goat on over.
func CaptureMove(from, over, to PositionID) Move {
	return Move{Kind: Capture, From: from, To: to, Over: over}
}

func (m Move) String() string {
	switch m.Kind {
	case Placement:
		return fmt.Sprintf("place %s", Notation(m.To))
	case Capture:
		return fmt.Sprintf("%s x%s -> %s", Notation(m.From), Notation(m.Over), Notation(m.To))
	default:
		return fmt.Sprintf("%s -> %s", Notation(m.From), Notation(m.To))
	}
}

// Notation renders a point as algebraic coordinates, a1 (top-left) to e5.
func Notation(id PositionID) string {
	if id < 0 || int(id) >= PositionCount {
		return "-"
	}
	row, col := RowCol(id)
	return fmt.Sprintf("%c%d", 'a'+rune(col), row+1)
}

// ParseNotation converts "c3" style coordinates back to a PositionID.
func ParseNotation(s string) (PositionID, bool) {
	if len(s) != 2 {
		return NoPosition, false
	}
	col := int(s[0] - 'a')
	row := int(s[1] - '1')
	if !inBounds(row, col) {
		return NoPosition, false
	}
	return ID(row, col), true
}
