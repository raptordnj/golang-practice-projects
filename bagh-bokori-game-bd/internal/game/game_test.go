package game

import (
	"errors"
	"math/rand"
	"testing"
)

// ---------- board topology ----------

func TestBoardHasAllPositions(t *testing.T) {
	b := NewBoard()
	if got := len(b.Positions); got != PositionCount {
		t.Fatalf("positions = %d, want %d", got, PositionCount)
	}
	for i, p := range b.Positions {
		if p.ID != PositionID(i) {
			t.Errorf("position %d has ID %d", i, p.ID)
		}
		if p.X < 0 || p.X > 1 || p.Y < 0 || p.Y > 1 {
			t.Errorf("%s: normalised coords out of range: %v,%v", Notation(p.ID), p.X, p.Y)
		}
	}
}

// The Alquerque pattern gives each point a fixed degree: 8 where all four
// diagonals are drawn, 5 on the even-parity edge points, 4 on the odd-parity
// interior points and 3 at the corners. Assert the exact table rather than
// trusting prose.
func TestNeighborDegrees(t *testing.T) {
	b := NewBoard()
	want := map[string]int{
		"a1": 3, "b1": 3, "c1": 5, "d1": 3, "e1": 3,
		"a2": 3, "b2": 8, "c2": 4, "d2": 8, "e2": 3,
		"a3": 5, "b3": 4, "c3": 8, "d3": 4, "e3": 5,
		"a4": 3, "b4": 8, "c4": 4, "d4": 8, "e4": 3,
		"a5": 3, "b5": 3, "c5": 5, "d5": 3, "e5": 3,
	}
	for notation, degree := range want {
		id, ok := ParseNotation(notation)
		if !ok {
			t.Fatalf("bad notation %q", notation)
		}
		if got := len(b.Neighbors(id)); got != degree {
			t.Errorf("%s: degree = %d, want %d", notation, got, degree)
		}
	}
}

func TestAdjacencyIsSymmetric(t *testing.T) {
	b := NewBoard()
	for a := PositionID(0); a < PositionCount; a++ {
		for c := PositionID(0); c < PositionCount; c++ {
			if b.AreAdjacent(a, c) != b.AreAdjacent(c, a) {
				t.Fatalf("asymmetric edge %s-%s", Notation(a), Notation(c))
			}
		}
		if b.AreAdjacent(a, a) {
			t.Fatalf("%s is adjacent to itself", Notation(a))
		}
	}
}

func TestNoDiagonalThroughOddPoints(t *testing.T) {
	b := NewBoard()
	// b1 (row 0, col 1) has odd row+col, so it must have no diagonal lines.
	b1, _ := ParseNotation("b1")
	for _, n := range b.Neighbors(b1) {
		r, c := RowCol(n)
		nr, nc := RowCol(b1)
		if r != nr && c != nc {
			t.Errorf("b1 has an illegal diagonal to %s", Notation(n))
		}
	}
}

func TestJumpPathsFollowLines(t *testing.T) {
	b := NewBoard()
	for from := PositionID(0); from < PositionCount; from++ {
		for _, j := range b.Jumps(from) {
			if !b.AreAdjacent(from, j.Over) || !b.AreAdjacent(j.Over, j.To) {
				t.Errorf("jump %s-%s-%s does not follow drawn lines",
					Notation(from), Notation(j.Over), Notation(j.To))
			}
			fr, fc := RowCol(from)
			or, oc := RowCol(j.Over)
			tr, tc := RowCol(j.To)
			if or-fr != tr-or || oc-fc != tc-oc {
				t.Errorf("jump %s-%s-%s is not collinear",
					Notation(from), Notation(j.Over), Notation(j.To))
			}
		}
	}
}

func TestOffBoardIDsRejected(t *testing.T) {
	b := NewBoard()
	for _, id := range []PositionID{-1, PositionCount, 999} {
		if b.IsValid(id) {
			t.Errorf("%d reported valid", id)
		}
		if b.Neighbors(id) != nil {
			t.Errorf("%d returned neighbors", id)
		}
	}
}

func TestNotationRoundTrip(t *testing.T) {
	for id := PositionID(0); id < PositionCount; id++ {
		back, ok := ParseNotation(Notation(id))
		if !ok || back != id {
			t.Errorf("round trip failed for %d (%q)", id, Notation(id))
		}
	}
	if _, ok := ParseNotation("z9"); ok {
		t.Error("z9 parsed as a valid point")
	}
}

// ---------- helpers ----------

// position builds a state directly, for focused rule tests.
func position(t *testing.T, tiger PositionID, goats []PositionID, turn Player, placed, captured int) GameState {
	t.Helper()
	var s GameState
	s.Tiger = tiger
	s.Occupancy[tiger] = Tiger
	for _, g := range goats {
		s.Occupancy[g] = Goat
	}
	s.Turn = turn
	s.GoatsPlaced = placed
	s.GoatsCaptured = captured
	return s
}

func at(t *testing.T, notation string) PositionID {
	t.Helper()
	id, ok := ParseNotation(notation)
	if !ok {
		t.Fatalf("bad notation %q", notation)
	}
	return id
}

func hasMove(moves []Move, m Move) bool {
	for _, got := range moves {
		if got == m {
			return true
		}
	}
	return false
}

// ---------- opening position ----------

func TestNewGame(t *testing.T) {
	r := NewDefaultRules()
	s := r.NewGame()

	if s.Tiger != ID(2, 2) {
		t.Errorf("tiger starts at %s, want c3", Notation(s.Tiger))
	}
	if s.Occupancy[ID(2, 2)] != Tiger {
		t.Error("centre point does not hold the tiger")
	}
	if s.GoatsOnBoard() != 0 || s.GoatsPlaced != 0 || s.GoatsCaptured != 0 {
		t.Error("new game should have no goats on the board")
	}
	if s.Turn != GoatPlayer {
		t.Errorf("first turn = %v, want Goat", s.Turn)
	}
	if s.IsOver() {
		t.Error("new game is already over")
	}
	if got := s.GoatsInHand(&r.cfg); got != 16 {
		t.Errorf("goats in hand = %d, want 16", got)
	}
}

func TestOpeningLegalMovesArePlacements(t *testing.T) {
	r := NewDefaultRules()
	moves := r.LegalMoves(r.NewGame())
	if len(moves) != PositionCount-1 {
		t.Fatalf("got %d opening moves, want %d", len(moves), PositionCount-1)
	}
	for _, m := range moves {
		if m.Kind != Placement {
			t.Fatalf("opening move %v is not a placement", m)
		}
		if m.To == ID(2, 2) {
			t.Error("placement offered on the tiger's point")
		}
	}
}

// ---------- goat placement ----------

func TestGoatPlacementUpdatesCounters(t *testing.T) {
	r := NewDefaultRules()
	s := r.NewGame()
	next, err := r.ApplyMove(s, PlaceMove(at(t, "a1")))
	if err != nil {
		t.Fatalf("placement failed: %v", err)
	}
	if next.Occupancy[at(t, "a1")] != Goat {
		t.Error("goat not placed")
	}
	if next.GoatsPlaced != 1 || next.GoatsOnBoard() != 1 {
		t.Errorf("counters wrong: placed=%d onBoard=%d", next.GoatsPlaced, next.GoatsOnBoard())
	}
	if next.Turn != TigerPlayer {
		t.Error("turn did not pass to the tiger")
	}
	if s.GoatsPlaced != 0 || s.Occupancy[at(t, "a1")] != Empty {
		t.Error("ApplyMove mutated the input state")
	}
}

func TestPlacementOnOccupiedPointRejected(t *testing.T) {
	r := NewDefaultRules()
	s := r.NewGame()
	if _, err := r.ApplyMove(s, PlaceMove(ID(2, 2))); !errors.Is(err, ErrOccupiedPosition) {
		t.Errorf("err = %v, want ErrOccupiedPosition", err)
	}
}

func TestPlacementOffBoardRejected(t *testing.T) {
	r := NewDefaultRules()
	if _, err := r.ApplyMove(r.NewGame(), PlaceMove(99)); !errors.Is(err, ErrOutOfBounds) {
		t.Errorf("err = %v, want ErrOutOfBounds", err)
	}
}

func TestGoatsCannotMoveDuringPlacement(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), []PositionID{at(t, "a1")}, GoatPlayer, 1, 0)
	move := StepMove(at(t, "a1"), at(t, "b1"))
	if r.IsValidMove(s, move) {
		t.Error("goat allowed to move before all goats were placed")
	}
	for _, m := range r.LegalMoves(s) {
		if m.Kind == Step {
			t.Fatalf("step move %v offered during placement", m)
		}
	}
}

func TestGoatsMoveAfterPlacementPhase(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), []PositionID{at(t, "a1")}, GoatPlayer, 16, 15)
	moves := r.LegalMoves(s)
	if len(moves) == 0 {
		t.Fatal("no goat moves after placement phase")
	}
	for _, m := range moves {
		if m.Kind != Step {
			t.Fatalf("move %v is not a step", m)
		}
	}
	if !hasMove(moves, StepMove(at(t, "a1"), at(t, "b1"))) {
		t.Error("a1-b1 not offered")
	}
	if hasMove(moves, StepMove(at(t, "a1"), at(t, "c1"))) {
		t.Error("a1-c1 offered but is not adjacent")
	}
}

func TestPlacementRejectedAfterPhaseEnds(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), []PositionID{at(t, "a1")}, GoatPlayer, 16, 15)
	if _, err := r.ApplyMove(s, PlaceMove(at(t, "e5"))); !errors.Is(err, ErrInvalidMove) {
		t.Errorf("err = %v, want ErrInvalidMove", err)
	}
}

// ---------- tiger movement ----------

func TestTigerStepsToAdjacentEmptyPoint(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), nil, TigerPlayer, 1, 0)
	next, err := r.ApplyMove(s, StepMove(at(t, "c3"), at(t, "b2")))
	if err != nil {
		t.Fatalf("tiger step failed: %v", err)
	}
	if next.Tiger != at(t, "b2") || next.Occupancy[at(t, "b2")] != Tiger {
		t.Error("tiger did not move to b2")
	}
	if next.Occupancy[at(t, "c3")] != Empty {
		t.Error("tiger left a copy behind")
	}
}

func TestTigerCannotStepToNonAdjacentPoint(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), nil, TigerPlayer, 1, 0)
	if _, err := r.ApplyMove(s, StepMove(at(t, "c3"), at(t, "a1"))); !errors.Is(err, ErrInvalidMove) {
		t.Errorf("err = %v, want ErrInvalidMove", err)
	}
	// c3-b1 is not a drawn line either (knight-shaped).
	if r.IsValidMove(s, StepMove(at(t, "c3"), at(t, "b1"))) {
		t.Error("c3-b1 accepted but no line exists")
	}
}

func TestTigerCannotStepOntoOccupiedPoint(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), []PositionID{at(t, "c2")}, TigerPlayer, 1, 0)
	if _, err := r.ApplyMove(s, StepMove(at(t, "c3"), at(t, "c2"))); !errors.Is(err, ErrOccupiedPosition) {
		t.Errorf("err = %v, want ErrOccupiedPosition", err)
	}
}

func TestGoatCannotMoveTiger(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), []PositionID{at(t, "a1")}, GoatPlayer, 16, 15)
	if _, err := r.ApplyMove(s, StepMove(at(t, "c3"), at(t, "b2"))); !errors.Is(err, ErrWrongTurn) {
		t.Errorf("err = %v, want ErrWrongTurn", err)
	}
}

func TestTigerCannotMoveGoat(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), []PositionID{at(t, "a1")}, TigerPlayer, 16, 15)
	if _, err := r.ApplyMove(s, StepMove(at(t, "a1"), at(t, "b1"))); !errors.Is(err, ErrWrongTurn) {
		t.Errorf("err = %v, want ErrWrongTurn", err)
	}
}

func TestMovingFromEmptyPointRejected(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), nil, TigerPlayer, 1, 0)
	if _, err := r.ApplyMove(s, StepMove(at(t, "a1"), at(t, "b1"))); !errors.Is(err, ErrInvalidMove) {
		t.Errorf("err = %v, want ErrInvalidMove", err)
	}
}

// ---------- captures ----------

func TestTigerCapturesGoat(t *testing.T) {
	r := NewDefaultRules()
	// c3 -> over c2 -> c1, straight up the centre file.
	s := position(t, at(t, "c3"), []PositionID{at(t, "c2")}, TigerPlayer, 4, 0)
	m := CaptureMove(at(t, "c3"), at(t, "c2"), at(t, "c1"))
	if !r.IsValidMove(s, m) {
		t.Fatal("legal capture rejected")
	}
	next, err := r.ApplyMove(s, m)
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if next.Occupancy[at(t, "c2")] != Empty {
		t.Error("captured goat was not removed")
	}
	if next.Tiger != at(t, "c1") {
		t.Error("tiger did not land on c1")
	}
	if next.GoatsCaptured != 1 || next.GoatsOnBoard() != 3 {
		t.Errorf("captured=%d onBoard=%d", next.GoatsCaptured, next.GoatsOnBoard())
	}
	if next.GoatsPlaced != 4 {
		t.Error("capture must not change the number of goats placed")
	}
}

func TestCaptureRemovesExactlyOneGoat(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "a3"), []PositionID{at(t, "b3"), at(t, "c3"), at(t, "d3")}, TigerPlayer, 5, 0)
	// a3 x b3 -> c3 is blocked because c3 holds a goat.
	if r.IsValidMove(s, CaptureMove(at(t, "a3"), at(t, "b3"), at(t, "c3"))) {
		t.Fatal("capture onto an occupied landing point accepted")
	}
	s2 := position(t, at(t, "a3"), []PositionID{at(t, "b3"), at(t, "d3")}, TigerPlayer, 5, 0)
	next, err := r.ApplyMove(s2, CaptureMove(at(t, "a3"), at(t, "b3"), at(t, "c3")))
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if next.GoatsCaptured != 1 {
		t.Errorf("captured %d goats in one jump", next.GoatsCaptured)
	}
	if next.Occupancy[at(t, "d3")] != Goat {
		t.Error("an uninvolved goat was removed")
	}
}

func TestCannotJumpOverEmptyPoint(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), nil, TigerPlayer, 1, 0)
	if r.IsValidMove(s, CaptureMove(at(t, "c3"), at(t, "c2"), at(t, "c1"))) {
		t.Error("jump over an empty point accepted")
	}
}

func TestCannotJumpOverTiger(t *testing.T) {
	r := NewDefaultRules()
	// There is only one tiger, so verify a jump over a point that holds no
	// goat is refused regardless of what the caller claims in Over.
	s := position(t, at(t, "c3"), []PositionID{at(t, "c2")}, TigerPlayer, 2, 0)
	if r.IsValidMove(s, CaptureMove(at(t, "c3"), at(t, "b2"), at(t, "c1"))) {
		t.Error("capture with a mismatched Over field accepted")
	}
}

func TestCannotJumpOffBoard(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c2"), []PositionID{at(t, "c1")}, TigerPlayer, 2, 0)
	// c2 -> c1 -> off the top edge: no such jump exists.
	for _, j := range NewBoard().Jumps(at(t, "c2")) {
		if j.Over == at(t, "c1") {
			t.Errorf("jump over the top edge exists: %s", Notation(j.To))
		}
	}
	if len(r.LegalMoves(s)) == 0 {
		t.Error("tiger should still have ordinary steps")
	}
}

func TestDiagonalCaptureOnlyOnDiagonalLines(t *testing.T) {
	r := NewDefaultRules()
	// c3-b2-a1 is a drawn diagonal, so the capture is legal.
	s := position(t, at(t, "c3"), []PositionID{at(t, "b2")}, TigerPlayer, 2, 0)
	if !r.IsValidMove(s, CaptureMove(at(t, "c3"), at(t, "b2"), at(t, "a1"))) {
		t.Error("legal diagonal capture c3xb2-a1 rejected")
	}
	// b1-c2-d3 crosses odd points, so no diagonal line exists there.
	s2 := position(t, at(t, "b1"), []PositionID{at(t, "c2")}, TigerPlayer, 2, 0)
	if r.IsValidMove(s2, CaptureMove(at(t, "b1"), at(t, "c2"), at(t, "d3"))) {
		t.Error("capture along a line that is not drawn accepted")
	}
}

func TestCaptureOfferedInLegalMoves(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), []PositionID{at(t, "c2")}, TigerPlayer, 2, 0)
	if !hasMove(r.LegalMoves(s), CaptureMove(at(t, "c3"), at(t, "c2"), at(t, "c1"))) {
		t.Error("capture missing from legal moves")
	}
}

func TestForcedCaptureRuleset(t *testing.T) {
	cfg := DefaultRuleSet()
	cfg.ForcedCapture = true
	r := NewRules(cfg)
	s := position(t, at(t, "c3"), []PositionID{at(t, "c2")}, TigerPlayer, 2, 0)
	moves := r.LegalMoves(s)
	for _, m := range moves {
		if m.Kind != Capture {
			t.Fatalf("non-capture %v offered while a capture exists", m)
		}
	}
	if _, err := r.ApplyMove(s, StepMove(at(t, "c3"), at(t, "b3"))); !errors.Is(err, ErrInvalidMove) {
		t.Errorf("err = %v, want ErrInvalidMove", err)
	}
}

// ---------- outcomes ----------

func TestTigerWinsOnCaptureThreshold(t *testing.T) {
	cfg := DefaultRuleSet()
	r := NewRules(cfg)
	s := position(t, at(t, "c3"), []PositionID{at(t, "c2")}, TigerPlayer, 16, cfg.TigerWinCaptures-1)
	next, err := r.ApplyMove(s, CaptureMove(at(t, "c3"), at(t, "c2"), at(t, "c1")))
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if next.Status != TigerWon {
		t.Errorf("status = %v, want TigerWon", next.Status)
	}
	if r.Winner(next) != TigerWinner {
		t.Errorf("winner = %v, want Tiger", r.Winner(next))
	}
	if !r.IsGameOver(next) {
		t.Error("game should be over")
	}
}

func TestGoatsWinWhenTigerIsFenced(t *testing.T) {
	r := NewDefaultRules()
	// Tiger in the a1 corner. Its lines run to b1, a2 and b2; fill b1 and
	// b2, then the final goat lands on a2 and the tiger is trapped. No
	// jump is possible because the landing points behind the goats are
	// themselves occupied or off-board.
	goats := []PositionID{
		at(t, "b1"), at(t, "b2"), at(t, "c1"), at(t, "c3"),
		at(t, "a3"),
	}
	s := position(t, at(t, "a1"), goats, GoatPlayer, 15, 0)
	next, err := r.ApplyMove(s, PlaceMove(at(t, "a2")))
	if err != nil {
		t.Fatalf("placement failed: %v", err)
	}
	if next.Status != GoatsWon {
		t.Fatalf("status = %v, want GoatsWon (tiger moves: %v)", next.Status, r.LegalMoves(next))
	}
	if r.Winner(next) != GoatWinner {
		t.Errorf("winner = %v, want Goats", r.Winner(next))
	}
}

func TestTigerNotTrappedWhenAJumpRemains(t *testing.T) {
	r := NewDefaultRules()
	// Same corner fence, but c1 and a3 are empty, so the tiger can jump.
	goats := []PositionID{at(t, "b1"), at(t, "b2"), at(t, "a2")}
	s := position(t, at(t, "a1"), goats, TigerPlayer, 16, 0)
	if s.Status != InProgress {
		t.Fatal("test fixture already finished")
	}
	moves := r.LegalMoves(s)
	if len(moves) == 0 {
		t.Fatal("tiger has no moves but jumps should be available")
	}
	for _, m := range moves {
		if m.Kind != Capture {
			t.Errorf("non-capture %v offered from a fenced corner", m)
		}
	}
}

func TestDrawOnHalfmoveLimit(t *testing.T) {
	cfg := DefaultRuleSet()
	cfg.HalfmoveLimit = 2
	r := NewRules(cfg)
	s := position(t, at(t, "c3"), []PositionID{at(t, "a1")}, TigerPlayer, 16, 0)
	s, err := r.ApplyMove(s, StepMove(at(t, "c3"), at(t, "b3")))
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != InProgress {
		t.Fatalf("premature end: %v", s.Status)
	}
	s, err = r.ApplyMove(s, StepMove(at(t, "a1"), at(t, "b1")))
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != Draw {
		t.Errorf("status = %v, want Draw", s.Status)
	}
	if r.Winner(s) != NoWinner {
		t.Error("a draw must have no winner")
	}
}

func TestPlacementResetsHalfmoveClock(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), nil, TigerPlayer, 0, 0)
	s, _ = r.ApplyMove(s, StepMove(at(t, "c3"), at(t, "b3")))
	if s.HalfmoveClock != 1 {
		t.Fatalf("clock = %d, want 1", s.HalfmoveClock)
	}
	s, _ = r.ApplyMove(s, PlaceMove(at(t, "a1")))
	if s.HalfmoveClock != 0 {
		t.Errorf("clock = %d after placement, want 0", s.HalfmoveClock)
	}
}

func TestMovesRejectedAfterGameOver(t *testing.T) {
	r := NewDefaultRules()
	s := position(t, at(t, "c3"), []PositionID{at(t, "a1")}, GoatPlayer, 16, 0)
	s.Status = TigerWon
	if _, err := r.ApplyMove(s, StepMove(at(t, "a1"), at(t, "b1"))); !errors.Is(err, ErrGameOver) {
		t.Errorf("err = %v, want ErrGameOver", err)
	}
	if len(r.LegalMoves(s)) != 0 {
		t.Error("finished game still reports legal moves")
	}
}

func TestUnknownMoveKindRejected(t *testing.T) {
	r := NewDefaultRules()
	s := r.NewGame()
	if r.IsValidMove(s, Move{Kind: 0, From: NoPosition, To: at(t, "a1")}) {
		t.Error("move with zero kind accepted")
	}
}

// ---------- engine, undo, redo ----------

func TestEnginePlayAndUndoRedo(t *testing.T) {
	e := NewDefaultEngine()
	start := e.State()

	if err := e.Play(PlaceMove(at(t, "a1"))); err != nil {
		t.Fatalf("play failed: %v", err)
	}
	if e.State().GoatsPlaced != 1 {
		t.Error("engine did not record the placement")
	}
	if !e.CanUndo() || e.CanRedo() {
		t.Error("undo/redo availability wrong after one move")
	}

	if err := e.Undo(); err != nil {
		t.Fatalf("undo failed: %v", err)
	}
	if e.State() != start {
		t.Error("undo did not restore the opening position")
	}
	if !e.CanRedo() {
		t.Error("redo should be available after undo")
	}

	if err := e.Redo(); err != nil {
		t.Fatalf("redo failed: %v", err)
	}
	if e.State().GoatsPlaced != 1 {
		t.Error("redo did not reapply the move")
	}
}

func TestEngineUndoAtStartFails(t *testing.T) {
	e := NewDefaultEngine()
	if err := e.Undo(); !errors.Is(err, ErrNothingToUndo) {
		t.Errorf("err = %v, want ErrNothingToUndo", err)
	}
	if err := e.Redo(); !errors.Is(err, ErrNothingToRedo) {
		t.Errorf("err = %v, want ErrNothingToRedo", err)
	}
}

func TestPlayTruncatesRedoHistory(t *testing.T) {
	e := NewDefaultEngine()
	mustPlay(t, e, PlaceMove(at(t, "a1")))
	mustPlay(t, e, StepMove(at(t, "c3"), at(t, "b3")))
	if err := e.Undo(); err != nil {
		t.Fatal(err)
	}
	if err := e.Undo(); err != nil {
		t.Fatal(err)
	}
	mustPlay(t, e, PlaceMove(at(t, "e5")))
	if e.CanRedo() {
		t.Error("redo survived a new move")
	}
	if len(e.History()) != 1 {
		t.Errorf("history length = %d, want 1", len(e.History()))
	}
}

// UndoTurn must hand the human back a turn that is actually theirs, even
// though the AI replied in between. Regression test for turn corruption.
func TestUndoTurnSkipsTheAIReply(t *testing.T) {
	e := NewDefaultEngine()
	mustPlay(t, e, PlaceMove(at(t, "a1")))             // goat (human)
	mustPlay(t, e, StepMove(at(t, "c3"), at(t, "b3"))) // tiger (AI)
	if e.State().Turn != GoatPlayer {
		t.Fatal("fixture: goat should be to move")
	}
	mustPlay(t, e, PlaceMove(at(t, "e5")))             // goat (human)
	mustPlay(t, e, StepMove(at(t, "b3"), at(t, "c3"))) // tiger (AI)

	if err := e.UndoTurn(GoatPlayer); err != nil {
		t.Fatalf("undo turn failed: %v", err)
	}
	if e.State().Turn != GoatPlayer {
		t.Errorf("turn = %v after UndoTurn, want Goat", e.State().Turn)
	}
	if e.State().GoatsPlaced != 1 {
		t.Errorf("goats placed = %d, want 1", e.State().GoatsPlaced)
	}
}

func TestRestartClearsHistory(t *testing.T) {
	e := NewDefaultEngine()
	mustPlay(t, e, PlaceMove(at(t, "a1")))
	e.Restart()
	if e.CanUndo() || e.CanRedo() || len(e.History()) != 0 {
		t.Error("restart left history behind")
	}
	if e.State().GoatsPlaced != 0 {
		t.Error("restart did not reset the position")
	}
}

func TestLegalMovesFrom(t *testing.T) {
	e := NewDefaultEngine()
	mustPlay(t, e, PlaceMove(at(t, "a1")))
	moves := e.LegalMovesFrom(at(t, "c3"))
	if len(moves) == 0 {
		t.Fatal("tiger has no moves from c3")
	}
	for _, m := range moves {
		if m.From != at(t, "c3") {
			t.Errorf("move %v does not start at c3", m)
		}
	}
	if len(e.LegalMovesFrom(at(t, "e5"))) != 0 {
		t.Error("moves reported from an empty point")
	}
}

func mustPlay(t *testing.T, e *Engine, m Move) {
	t.Helper()
	if err := e.Play(m); err != nil {
		t.Fatalf("play %v: %v", m, err)
	}
}

// ---------- save / load ----------

func TestSnapshotRoundTrip(t *testing.T) {
	e := NewDefaultEngine()
	mustPlay(t, e, PlaceMove(at(t, "a1")))
	mustPlay(t, e, StepMove(at(t, "c3"), at(t, "b3")))
	mustPlay(t, e, PlaceMove(at(t, "e5")))
	want := e.State()

	data, err := EncodeSnapshot(e.Snapshot())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	snap, err := DecodeSnapshot(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	restored := NewDefaultEngine()
	if err := restored.LoadSnapshot(snap); err != nil {
		t.Fatalf("load: %v", err)
	}
	if restored.State() != want {
		t.Error("restored state differs from the saved one")
	}
	if len(restored.History()) != 3 {
		t.Errorf("restored history length = %d, want 3", len(restored.History()))
	}
	if !restored.CanUndo() {
		t.Error("restored game cannot undo")
	}
}

func TestLoadSnapshotRejectsIllegalMoves(t *testing.T) {
	snap := Snapshot{
		Version: 1,
		Rules:   DefaultRuleSet(),
		Moves:   []Move{PlaceMove(ID(2, 2))}, // onto the tiger
		Cursor:  1,
	}
	if err := NewDefaultEngine().LoadSnapshot(snap); err == nil {
		t.Error("illegal saved game loaded without error")
	}
}

func TestDecodeSnapshotFillsMissingRules(t *testing.T) {
	snap, err := DecodeSnapshot([]byte(`{"version":1,"moves":[],"cursor":0}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if snap.Rules.TotalGoats != 16 {
		t.Errorf("total goats = %d, want the 16-goat default", snap.Rules.TotalGoats)
	}
}

// ---------- invariants under random play ----------

func TestRandomPlayouts(t *testing.T) {
	r := NewDefaultRules()
	rng := rand.New(rand.NewSource(1))

	for game := 0; game < 300; game++ {
		s := r.NewGame()
		for ply := 0; ply < 500 && !s.IsOver(); ply++ {
			moves := r.LegalMoves(s)
			if len(moves) == 0 {
				t.Fatalf("no legal moves in an unfinished position: %+v", s)
			}
			before := s
			m := moves[rng.Intn(len(moves))]
			next, err := r.ApplyMove(s, m)
			if err != nil {
				t.Fatalf("legal move %v rejected: %v", m, err)
			}
			if s != before {
				t.Fatal("ApplyMove mutated its argument")
			}
			checkInvariants(t, next)
			s = next
		}
	}
}

func checkInvariants(t *testing.T, s GameState) {
	t.Helper()
	tigers, goats := 0, 0
	for _, p := range s.Occupancy {
		switch p {
		case Tiger:
			tigers++
		case Goat:
			goats++
		}
	}
	if tigers != 1 {
		t.Fatalf("board holds %d tigers, want exactly 1", tigers)
	}
	if s.Occupancy[s.Tiger] != Tiger {
		t.Fatal("Tiger field disagrees with the occupancy grid")
	}
	if goats != s.GoatsOnBoard() {
		t.Fatalf("counted %d goats, state says %d", goats, s.GoatsOnBoard())
	}
	if s.GoatsPlaced > 16 || s.GoatsCaptured > s.GoatsPlaced {
		t.Fatalf("impossible counters: placed=%d captured=%d", s.GoatsPlaced, s.GoatsCaptured)
	}
}
