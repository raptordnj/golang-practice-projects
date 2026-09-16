package game

func tigerLegalMoves(state GameState) []Move {
	var moves []Move
	for _, t := range state.Tigers {
		// Movement to empty neighbors
		for _, n := range state.Board.EmptyNeighbors(t) {
			moves = append(moves, Move{From: t, To: n})
		}
		// Capture jumps
		for _, n := range BoardTopology[t] {
			if state.Board.At(n).Type == Goat {
				// Jump over goat to empty beyond
				for _, beyond := range BoardTopology[n] {
					if beyond != t && state.Board.At(beyond).Type == Empty {
						moves = append(moves, Move{From: t, To: beyond})
					}
				}
			}
		}
	}
	return moves
}

func goatLegalMoves(state GameState) []Move {
	var moves []Move
	if state.GoatsPlaced < 20 {
		// Placement phase
		for i := 0; i < BoardSize; i++ {
			if state.Board.At(PositionID(i)).Type == Empty {
				moves = append(moves, Move{From: PositionID(-1), To: PositionID(i)})
			}
		}
	} else {
		// Movement phase - goats move to empty adjacent positions
		for _, g := range state.Goats {
			for _, n := range state.Board.EmptyNeighbors(g) {
				moves = append(moves, Move{From: g, To: n})
			}
		}
	}
	return moves
}

func isCaptureMove(state GameState, move Move) bool {
	if state.Turn != PlayerTiger {
		return false
	}
	// Check if there's a goat between From and To
	for _, mid := range BoardTopology[move.From] {
		if mid == move.To {
			return false // Adjacent move, not a jump
		}
		if state.Board.At(mid).Type == Goat {
			// Check if To is a neighbor of mid
			for _, n := range BoardTopology[mid] {
				if n == move.To {
					return true
				}
			}
		}
	}
	return false
}

func capturePosition(state GameState, move Move) PositionID {
	for _, mid := range BoardTopology[move.From] {
		if state.Board.At(mid).Type == Goat {
			for _, n := range BoardTopology[mid] {
				if n == move.To {
					return mid
				}
			}
		}
	}
	return PositionID(-1)
}

func checkGameOver(state *GameState) {
	// Check tiger win: goats captured >= 5 or goats can't block
	if state.CapturedGoats >= 5 {
		state.Status = StatusTigerWin
		return
	}
	// Check goat win: all tigers blocked
	tigersCanMove := false
	for range state.Tigers {
		if len(tigerLegalMoves(*state)) > 0 {
			tigersCanMove = true
			break
		}
	}
	if !tigersCanMove {
		state.Status = StatusGoatWin
	}
}
