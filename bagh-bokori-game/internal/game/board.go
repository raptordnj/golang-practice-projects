package game

import (
	"strings"
)

const (
	BoardSize = 25
)

var BoardTopology = map[PositionID][]PositionID{
	// Row 0
	0: {1, 5},
	1: {0, 2, 6},
	2: {1, 3, 7},
	3: {2, 4, 8},
	4: {3, 9},
	// Row 1
	5: {0, 6, 10},
	6: {1, 5, 7, 11},
	7: {2, 6, 8, 12},
	8: {3, 7, 9, 13},
	9: {4, 8, 14},
	// Row 2 (center)
	10: {5, 11, 15},
	11: {6, 10, 12, 16},
	12: {7, 11, 13, 17},
	13: {8, 12, 14, 18},
	14: {9, 13, 19},
	// Row 3
	15: {10, 16, 20},
	16: {11, 15, 17, 21},
	17: {12, 16, 18, 22},
	18: {13, 17, 19, 23},
	19: {14, 18, 24},
	// Row 4
	20: {15, 21},
	21: {16, 20, 22},
	22: {17, 21, 23},
	23: {18, 22, 24},
	24: {19, 23},
}

type Board [BoardSize]Piece

func NewBoard() Board {
	return Board{}
}

func (b Board) Neighbors(pos PositionID) []PositionID {
	return BoardTopology[pos]
}

func (b Board) At(pos PositionID) Piece {
	return b[pos]
}

func (b Board) Set(pos PositionID, p Piece) Board {
	b[pos] = p
	return b
}

func (b Board) EmptyNeighbors(pos PositionID) []PositionID {
	var result []PositionID
	for _, n := range BoardTopology[pos] {
		if b[n].Type == Empty {
			result = append(result, n)
		}
	}
	return result
}

func (b Board) CountPiece(pieceType PieceType) int {
	count := 0
	for _, p := range b {
		if p.Type == pieceType {
			count++
		}
	}
	return count
}

func (b Board) String() string {
	var sb strings.Builder
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			pos := PositionID(row*5 + col)
			p := b[pos]
			switch p.Type {
			case Tiger:
				sb.WriteString(" T ")
			case Goat:
				sb.WriteString(" G ")
			default:
				sb.WriteString(" . ")
			}
			if col < 4 {
				sb.WriteString("-")
			}
		}
		sb.WriteString("\n")
		if row < 4 {
			sb.WriteString("  |   |   |   |\n")
		}
	}
	return sb.String()
}
