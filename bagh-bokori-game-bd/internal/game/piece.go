package game

// PieceType is the occupant of a point.
type PieceType uint8

const (
	// Empty means no piece stands on the point.
	Empty PieceType = iota
	// Tiger is the বাঘ.
	Tiger
	// Goat is the ছাগল.
	Goat
)

func (p PieceType) String() string {
	switch p {
	case Tiger:
		return "Tiger"
	case Goat:
		return "Goat"
	default:
		return "Empty"
	}
}

// Bengali returns the Bangla name of the piece, used by the UI.
func (p PieceType) Bengali() string {
	switch p {
	case Tiger:
		return "বাঘ"
	case Goat:
		return "ছাগল"
	default:
		return "খালি"
	}
}

// Player is a side to move.
type Player uint8

const (
	// TigerPlayer moves the single tiger.
	TigerPlayer Player = iota + 1
	// GoatPlayer places and moves the goats.
	GoatPlayer
)

func (p Player) String() string {
	if p == TigerPlayer {
		return "Tiger"
	}
	return "Goat"
}

// Bengali returns the Bangla name of the side.
func (p Player) Bengali() string {
	if p == TigerPlayer {
		return "বাঘ"
	}
	return "ছাগল"
}

// Opponent returns the other side.
func (p Player) Opponent() Player {
	if p == TigerPlayer {
		return GoatPlayer
	}
	return TigerPlayer
}

// Piece returns the piece type this player commands.
func (p Player) Piece() PieceType {
	if p == TigerPlayer {
		return Tiger
	}
	return Goat
}
