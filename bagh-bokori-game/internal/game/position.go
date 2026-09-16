package game

import "fmt"

type PositionID int

type PieceType int

const (
	Empty PieceType = iota
	Tiger
	Goat
)

type Player int

const (
	PlayerTiger Player = iota
	PlayerGoat
)

type Move struct {
	From PositionID
	To   PositionID
}

func (m Move) String() string {
	return fmtMove(m.From) + "->" + fmtMove(m.To)
}

func fmtMove(p PositionID) string {
	return fmt.Sprintf("%c", 'A'+int(p))
}

type Piece struct {
	Type   PieceType
	Player Player
}
