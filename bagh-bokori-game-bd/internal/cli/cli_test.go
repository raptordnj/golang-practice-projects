package cli

import (
	"bytes"
	"strings"
	"testing"

	"bagh-bandi/internal/ai"
	"bagh-bandi/internal/game"
)

func newTestSession() (*Session, *bytes.Buffer) {
	var out bytes.Buffer
	opts := DefaultOptions()
	opts.Out = &out
	opts.In = strings.NewReader("")
	return NewSession(opts), &out
}

func exec(t *testing.T, s *Session, line string) {
	t.Helper()
	if err := s.Execute(line); err != nil {
		t.Fatalf("%q: %v", line, err)
	}
}

func TestCLIPlaceAndMove(t *testing.T) {
	s, out := newTestSession()

	exec(t, s, "place a1")
	if s.Engine().State().GoatsPlaced != 1 {
		t.Fatal("goat not placed")
	}
	exec(t, s, "move c3 b3")
	if s.Engine().State().Tiger != game.ID(2, 1) {
		t.Fatalf("tiger at %s, want b3", game.Notation(s.Engine().State().Tiger))
	}
	s.flush()
	if !strings.Contains(out.String(), "Turn:") {
		t.Error("board output missing the status line")
	}
}

// "move a1" with a single argument is placement shorthand.
func TestCLIMoveShorthandPlaces(t *testing.T) {
	s, _ := newTestSession()
	exec(t, s, "move e5")
	if s.Engine().State().Occupancy[game.ID(4, 4)] != game.Goat {
		t.Error("shorthand placement did not place a goat")
	}
}

// The CLI resolves a capture from just "from to" - the user never types the
// jumped-over point.
func TestCLIResolvesCaptureFromFromTo(t *testing.T) {
	s, _ := newTestSession()
	exec(t, s, "place c2")
	if err := s.Execute("move c3 c1"); err != nil {
		t.Fatalf("capture rejected: %v", err)
	}
	st := s.Engine().State()
	if st.GoatsCaptured != 1 || st.Occupancy[game.ID(1, 2)] != game.Empty {
		t.Error("capture was not applied")
	}
}

func TestCLIRejectsIllegalMoveWithoutPanicking(t *testing.T) {
	s, _ := newTestSession()
	err := s.Execute("move c3 a1")
	if err == nil {
		t.Fatal("illegal move accepted")
	}
	if s.Engine().State().Tiger != game.ID(2, 2) {
		t.Error("board changed after an illegal move")
	}
}

func TestCLIRejectsGarbage(t *testing.T) {
	s, _ := newTestSession()
	for _, line := range []string{"frobnicate", "move", "move z9 a1", "place", "place q7"} {
		if err := s.Execute(line); err == nil {
			t.Errorf("%q was accepted", line)
		}
	}
	if err := s.Execute("   "); err != nil {
		t.Errorf("blank line returned %v", err)
	}
}

func TestCLIUndoRedoRestart(t *testing.T) {
	s, _ := newTestSession()
	exec(t, s, "place a1")
	exec(t, s, "undo")
	if s.Engine().State().GoatsPlaced != 0 {
		t.Error("undo did not take effect")
	}
	exec(t, s, "redo")
	if s.Engine().State().GoatsPlaced != 1 {
		t.Error("redo did not take effect")
	}
	exec(t, s, "restart")
	if s.Engine().State().GoatsPlaced != 0 || s.Engine().CanUndo() {
		t.Error("restart did not reset the game")
	}
	if err := s.Execute("undo"); err == nil {
		t.Error("undo at the start should report an error")
	}
}

func TestCLIShowCommands(t *testing.T) {
	s, out := newTestSession()
	for _, cmd := range []string{"help", "show", "status", "moves", "moves c3", "history"} {
		exec(t, s, cmd)
	}
	s.flush()
	text := out.String()
	for _, want := range []string{"Commands:", "a   b   c   d   e", "goats in hand", "legal moves", "no moves played"} {
		if !strings.Contains(text, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

// The board drawing must show diagonals only where the topology has them.
func TestCLIBoardDrawsRealDiagonals(t *testing.T) {
	s, out := newTestSession()
	exec(t, s, "show")
	s.flush()
	lines := strings.Split(out.String(), "\n")
	var connectors []string
	for _, l := range lines {
		if strings.HasPrefix(l, "   |") {
			connectors = append(connectors, l)
		}
	}
	if len(connectors) != game.BoardSize-1 {
		t.Fatalf("got %d connector rows, want %d", len(connectors), game.BoardSize-1)
	}
	// Between rows 1 and 2 the a1-b2 diagonal exists but b1-c2 does not.
	if strings.Count(connectors[0], "X") != 2 {
		t.Errorf("first connector row has %d crosses, want 2: %q",
			strings.Count(connectors[0], "X"), connectors[0])
	}
}

// The full loop, driven entirely by scripted input, must reach game over
// without a single illegal move - AI against AI.
func TestCLIRunPlaysAIGameToCompletion(t *testing.T) {
	var out bytes.Buffer
	opts := DefaultOptions()
	opts.Out = &out
	opts.In = strings.NewReader("")
	opts.TigerAI = ai.NewMinimaxAI(2, 5)
	opts.GoatAI = ai.NewRandomAI(9)

	s := NewSession(opts)
	if err := s.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !s.Engine().IsGameOver() {
		t.Errorf("AI vs AI game did not finish: %v", s.Engine().State().Status)
	}
	if strings.Contains(out.String(), "illegal move") {
		t.Error("an AI produced an illegal move")
	}
}

func TestCLIQuitStopsTheLoop(t *testing.T) {
	var out bytes.Buffer
	opts := DefaultOptions()
	opts.Out = &out
	opts.In = strings.NewReader("place a1\nmove c3 b3\nquit\n")
	s := NewSession(opts)
	if err := s.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if s.Engine().State().GoatsPlaced != 1 {
		t.Error("scripted commands were not executed")
	}
}
