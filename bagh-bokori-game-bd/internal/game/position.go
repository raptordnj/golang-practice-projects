package game

// PositionID identifies a single point (ghar) on the board graph.
// IDs are stable, deterministic and independent of any rendering.
type PositionID int

// NoPosition is the zero value used to signal "no point".
const NoPosition PositionID = -1

// Position is a single intersection of the board.
//
// X and Y are *presentation data only*. The engine never derives
// connectivity from them; connectivity lives in Neighbors.
type Position struct {
	ID        PositionID
	Row       int
	Col       int
	X         float64 // normalised 0..1, for renderers
	Y         float64 // normalised 0..1, for renderers
	Neighbors []PositionID
}

// IsValid reports whether id addresses a real point of the board.
func (b *Board) IsValid(id PositionID) bool {
	return id >= 0 && int(id) < len(b.Positions)
}
