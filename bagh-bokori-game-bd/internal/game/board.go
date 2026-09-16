package game

// BoardSize is the side length of the Bagh-Bandi board (5x5 points).
const BoardSize = 5

// PositionCount is the total number of points on the board.
const PositionCount = BoardSize * BoardSize

// Board is the immutable topology of the game: the graph of points, the
// edges between them, and the precomputed jump table used for captures.
//
// A Board is created once by NewBoard and is never mutated afterwards, so a
// single instance can be shared by the real game and by any number of
// concurrent AI simulations.
type Board struct {
	Positions []Position

	// adjacency[a][b] is true when a and b are joined by a drawn line.
	adjacency [PositionCount][PositionCount]bool

	// jumps[from] lists every (over, to) triple reachable from `from` by
	// jumping in a straight line over exactly one adjacent point.
	jumps [PositionCount][]Jump
}

// Jump is a precomputed capture path: from -> over -> to, all collinear and
// all connected by real board lines.
type Jump struct {
	Over PositionID
	To   PositionID
}

// ID converts a row/column pair to a PositionID.
func ID(row, col int) PositionID {
	return PositionID(row*BoardSize + col)
}

// RowCol converts a PositionID back to its row/column pair.
func RowCol(id PositionID) (row, col int) {
	return int(id) / BoardSize, int(id) % BoardSize
}

// defaultBoard is the shared, read-only topology instance.
var defaultBoard = buildBoard()

// NewBoard returns the standard Bagh-Bandi board topology.
//
// The board is the classic Alquerque pattern used across Bengal for
// বাঘবন্দী: a 5x5 grid of points where every point is joined orthogonally to
// its neighbours, and diagonals are drawn only through points whose
// (row+col) is even. That yields 8 lines at the centre, 4 at the corners and
// alternating 3/4 elsewhere - see README.md for the diagram.
func NewBoard() *Board { return defaultBoard }

func buildBoard() *Board {
	b := &Board{Positions: make([]Position, PositionCount)}

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			id := ID(row, col)
			b.Positions[id] = Position{
				ID:  id,
				Row: row,
				Col: col,
				X:   float64(col) / float64(BoardSize-1),
				Y:   float64(row) / float64(BoardSize-1),
			}
		}
	}

	// Edges. Orthogonal lines exist everywhere; diagonal lines exist only
	// where (row+col) is even, which is what produces the traditional
	// eight-pointed star pattern of the Alquerque board.
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			from := ID(row, col)
			for _, d := range directions {
				if d.diagonal && (row+col)%2 != 0 {
					continue
				}
				nr, nc := row+d.dr, col+d.dc
				if !inBounds(nr, nc) {
					continue
				}
				b.link(from, ID(nr, nc))
			}
		}
	}

	// Jump table. A jump is legal topology-wise when both the first and the
	// second step follow real lines in the same direction.
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			from := ID(row, col)
			for _, d := range directions {
				mr, mc := row+d.dr, col+d.dc
				tr, tc := row+2*d.dr, col+2*d.dc
				if !inBounds(mr, mc) || !inBounds(tr, tc) {
					continue
				}
				over, to := ID(mr, mc), ID(tr, tc)
				if !b.adjacency[from][over] || !b.adjacency[over][to] {
					continue
				}
				b.jumps[from] = append(b.jumps[from], Jump{Over: over, To: to})
			}
		}
	}

	for i := range b.Positions {
		b.Positions[i].Neighbors = b.neighborsOf(PositionID(i))
	}
	return b
}

type direction struct {
	dr, dc   int
	diagonal bool
}

var directions = []direction{
	{-1, 0, false}, {1, 0, false}, {0, -1, false}, {0, 1, false},
	{-1, -1, true}, {-1, 1, true}, {1, -1, true}, {1, 1, true},
}

func inBounds(row, col int) bool {
	return row >= 0 && row < BoardSize && col >= 0 && col < BoardSize
}

func (b *Board) link(a, c PositionID) {
	b.adjacency[a][c] = true
	b.adjacency[c][a] = true
}

func (b *Board) neighborsOf(id PositionID) []PositionID {
	var out []PositionID
	for other := PositionID(0); other < PositionCount; other++ {
		if b.adjacency[id][other] {
			out = append(out, other)
		}
	}
	return out
}

// AreAdjacent reports whether a line is drawn between a and b.
func (b *Board) AreAdjacent(a, c PositionID) bool {
	if !b.IsValid(a) || !b.IsValid(c) {
		return false
	}
	return b.adjacency[a][c]
}

// Neighbors returns the points connected to id. The slice must not be
// modified by callers.
func (b *Board) Neighbors(id PositionID) []PositionID {
	if !b.IsValid(id) {
		return nil
	}
	return b.Positions[id].Neighbors
}

// Jumps returns every topological jump path starting at id. The slice must
// not be modified by callers.
func (b *Board) Jumps(id PositionID) []Jump {
	if !b.IsValid(id) {
		return nil
	}
	return b.jumps[id]
}

// JumpOver returns the point jumped over when travelling from -> to in a
// straight line, and whether such a jump path exists.
func (b *Board) JumpOver(from, to PositionID) (PositionID, bool) {
	for _, j := range b.jumps[from] {
		if j.To == to {
			return j.Over, true
		}
	}
	return NoPosition, false
}
