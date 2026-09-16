package ui

import (
	"testing"
	"time"

	"bagh-bandi/internal/ai"
	"bagh-bandi/internal/game"
)

// fakeView records what the controller asks the board to draw.
type fakeView struct {
	state     game.GameState
	selected  game.PositionID
	targets   []game.Move
	lastMove  *game.Move
	flashed   []game.PositionID
	animated  []game.Move
	animateOn bool
}

func newFakeView() *fakeView {
	return &fakeView{selected: game.NoPosition, animateOn: false}
}

func (v *fakeView) SetState(s game.GameState) { v.state = s }
func (v *fakeView) SetSelection(id game.PositionID, moves []game.Move) {
	v.selected, v.targets = id, moves
}
func (v *fakeView) SetLastMove(m *game.Move)        { v.lastMove = m }
func (v *fakeView) FlashInvalid(id game.PositionID) { v.flashed = append(v.flashed, id) }
func (v *fakeView) AnimateMove(m game.Move, done func()) {
	v.animated = append(v.animated, m)
	done()
}

// fakeHUD records the chrome updates.
type fakeHUD struct {
	updates  int
	messages []string
	errors   []string
	overs    int
	status   game.GameStatus
	winner   game.Winner
}

func (h *fakeHUD) Update(Controller) { h.updates++ }
func (h *fakeHUD) Message(text string, isError bool) {
	if isError {
		h.errors = append(h.errors, text)
		return
	}
	if text != "" {
		h.messages = append(h.messages, text)
	}
}
func (h *fakeHUD) GameOver(status game.GameStatus, winner game.Winner) {
	h.overs++
	h.status, h.winner = status, winner
}

type fakeSound struct{ played []Effect }

func (s *fakeSound) Play(e Effect) { s.played = append(s.played, e) }

func newTestController(t *testing.T, mode Mode, human game.Player) (*GameController, *fakeView, *fakeHUD, *fakeSound) {
	t.Helper()
	v, h, snd := newFakeView(), &fakeHUD{}, &fakeSound{}
	c := NewGameController(ControllerOptions{
		Rules:      game.DefaultRuleSet(),
		Mode:       mode,
		HumanSide:  human,
		Difficulty: ai.Easy,
		ThinkTime:  500 * time.Millisecond,
		Seed:       42,
		View:       v,
		HUD:        h,
		Sound:      snd,
	})
	t.Cleanup(c.Close)
	return c, v, h, snd
}

func at(t *testing.T, notation string) game.PositionID {
	t.Helper()
	id, ok := game.ParseNotation(notation)
	if !ok {
		t.Fatalf("bad notation %q", notation)
	}
	return id
}

func TestTapPlacesGoatDuringPlacement(t *testing.T) {
	c, v, _, snd := newTestController(t, PlayerVsPlayer, game.GoatPlayer)

	c.TapPoint(at(t, "a1"))

	if c.State().GoatsPlaced != 1 {
		t.Fatal("tap did not place a goat")
	}
	if v.state.Occupancy[at(t, "a1")] != game.Goat {
		t.Error("board view was not updated")
	}
	if len(snd.played) == 0 || snd.played[0] != EffectPlace {
		t.Errorf("sounds = %v, want a place effect", snd.played)
	}
}

func TestTapOnOccupiedPointDuringPlacementIsRejected(t *testing.T) {
	c, v, h, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)

	c.TapPoint(at(t, "c3")) // the tiger's point

	if c.State().GoatsPlaced != 0 {
		t.Fatal("a goat was placed on the tiger")
	}
	if len(v.flashed) != 1 || v.flashed[0] != at(t, "c3") {
		t.Errorf("flashed = %v, want one flash on c3", v.flashed)
	}
	if len(h.errors) == 0 {
		t.Error("no message explained the refusal")
	}
}

// Select a tiger, see its destinations, then move it.
func TestSelectThenMoveTiger(t *testing.T) {
	c, v, _, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	c.TapPoint(at(t, "a1")) // goat places, tiger to move

	c.TapPoint(at(t, "c3"))
	if v.selected != at(t, "c3") {
		t.Fatalf("selection = %s, want c3", game.Notation(v.selected))
	}
	if len(v.targets) == 0 {
		t.Fatal("no legal destinations were shown")
	}

	c.TapPoint(at(t, "b3"))
	if c.State().Tiger != at(t, "b3") {
		t.Errorf("tiger at %s, want b3", game.Notation(c.State().Tiger))
	}
	if v.selected != game.NoPosition {
		t.Error("selection was not cleared after the move")
	}
	if len(v.animated) != 2 {
		t.Errorf("animated %d moves, want 2 (the placement and the slide)", len(v.animated))
	}
	if last := v.animated[len(v.animated)-1]; last.To != at(t, "b3") {
		t.Errorf("last animated move was %v, want the slide to b3", last)
	}
}

func TestTapIllegalDestinationClearsSelectionAndWarns(t *testing.T) {
	c, v, h, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	c.TapPoint(at(t, "a1"))

	c.TapPoint(at(t, "c3"))
	c.TapPoint(at(t, "e5")) // nowhere near the tiger

	if c.State().Tiger != at(t, "c3") {
		t.Error("the tiger moved on an illegal click")
	}
	if v.selected != game.NoPosition {
		t.Error("selection should be cleared")
	}
	if len(h.errors) == 0 {
		t.Error("no error message shown")
	}
	if len(v.flashed) == 0 {
		t.Error("no visual feedback for the invalid click")
	}
}

func TestTapEmptyPointWithNoSelectionWarns(t *testing.T) {
	c, _, h, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	c.TapPoint(at(t, "a1")) // place, tiger to move now
	c.TapPoint(at(t, "e5")) // empty, not the tiger

	if len(h.errors) == 0 {
		t.Error("clicking an empty point as the tiger should warn")
	}
}

func TestSelectingAGoatDuringMovementPhase(t *testing.T) {
	c, v, _, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	// Drive the engine straight to the movement phase.
	e := c.Engine()
	for e.State().InPlacementPhase(ptrCfg(e.Config())) {
		played := false
		for _, m := range e.LegalMoves() {
			if err := e.Play(m); err == nil {
				played = true
				break
			}
		}
		if !played {
			t.Fatal("could not reach the movement phase")
		}
	}
	c.sync()
	if e.State().Turn != game.GoatPlayer {
		for _, m := range e.LegalMoves() {
			if err := e.Play(m); err == nil {
				break
			}
		}
		c.sync()
	}

	goats := e.State().Goats()
	if len(goats) == 0 {
		t.Skip("no goats survived the scripted run")
	}
	for _, g := range goats {
		if len(e.LegalMovesFrom(g)) > 0 {
			c.TapPoint(g)
			if v.selected != g {
				t.Errorf("goat %s was not selected", game.Notation(g))
			}
			return
		}
	}
}

func ptrCfg(cfg game.RuleSet) *game.RuleSet { return &cfg }

func TestUndoAndRedo(t *testing.T) {
	c, _, h, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	c.TapPoint(at(t, "a1"))

	c.Undo()
	if c.State().GoatsPlaced != 0 {
		t.Error("undo did not take effect")
	}
	c.Redo()
	if c.State().GoatsPlaced != 1 {
		t.Error("redo did not take effect")
	}

	c.Undo()
	c.Undo() // nothing left
	if len(h.errors) == 0 {
		t.Error("undo past the start should report a message")
	}
}

func TestRestartResetsEverything(t *testing.T) {
	c, v, _, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	c.TapPoint(at(t, "a1"))
	c.TapPoint(at(t, "c3"))

	c.Restart()

	if c.State().GoatsPlaced != 0 || c.CanUndo() {
		t.Error("restart did not clear the game")
	}
	if v.selected != game.NoPosition || v.lastMove != nil {
		t.Error("restart did not clear the board highlights")
	}
}

// In PvC the human must not be able to move for the computer.
func TestHumanCannotMoveForTheComputer(t *testing.T) {
	c, v, h, _ := newTestController(t, PlayerVsComputer, game.GoatPlayer)
	// Force a tiger turn without letting the AI run.
	c.stopAI()
	e := c.Engine()
	if err := e.Play(game.PlaceMove(at(t, "a1"))); err != nil {
		t.Fatal(err)
	}
	c.sync()

	before := c.State()
	c.TapPoint(at(t, "c3"))

	if c.State() != before {
		t.Error("the human moved the computer's tiger")
	}
	if len(h.errors) == 0 || len(v.flashed) == 0 {
		t.Error("no feedback when clicking out of turn")
	}
}

// The AI must reply on its own, without freezing the caller.
func TestAIRepliesInPlayerVsComputer(t *testing.T) {
	v, h, snd := newFakeView(), &fakeHUD{}, &fakeSound{}
	done := make(chan struct{}, 8)

	c := NewGameController(ControllerOptions{
		Rules:      game.DefaultRuleSet(),
		Mode:       PlayerVsComputer,
		HumanSide:  game.GoatPlayer,
		Difficulty: ai.Easy,
		ThinkTime:  time.Second,
		Seed:       5,
		View:       v,
		HUD:        h,
		Sound:      snd,
		RunOnMain: func(f func()) {
			f()
			done <- struct{}{}
		},
	})
	defer c.Close()

	c.TapPoint(at(t, "a1"))

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the AI never replied")
	}

	if c.State().Turn != game.GoatPlayer {
		t.Errorf("turn = %v after the AI reply, want Goat", c.State().Turn)
	}
	if len(c.Engine().History()) != 2 {
		t.Errorf("history has %d moves, want 2", len(c.Engine().History()))
	}
}

// A search that finishes after the player restarted must be discarded, not
// applied to the new game. Regression test for stale AI moves.
func TestStaleAIMoveIsDiscarded(t *testing.T) {
	v, h := newFakeView(), &fakeHUD{}
	// The AI's reply is parked here instead of being applied, so the test
	// controls exactly when it lands. A channel rather than a shared
	// variable, so the handover is synchronised.
	pending := make(chan func(), 1)

	c := NewGameController(ControllerOptions{
		Rules:      game.DefaultRuleSet(),
		Mode:       PlayerVsComputer,
		HumanSide:  game.GoatPlayer,
		Difficulty: ai.Easy,
		ThinkTime:  time.Second,
		Seed:       5,
		View:       v,
		HUD:        h,
		RunOnMain:  func(f func()) { pending <- f }, // hold the reply back
	})
	defer c.Close()

	c.TapPoint(at(t, "a1"))

	var reply func()
	select {
	case reply = <-pending:
	case <-time.After(5 * time.Second):
		t.Fatal("the AI never produced a move")
	}

	c.Restart() // the player starts over while the AI's answer is in flight
	reply()     // ...and only now does it arrive

	if c.State().GoatsPlaced != 0 || len(c.Engine().History()) != 0 {
		t.Errorf("a stale AI move was applied: %+v", c.Engine().History())
	}
}

func TestGameOverIsAnnouncedExactlyOnce(t *testing.T) {
	c, _, h, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	e := c.Engine()

	// Walk the tiger into the a1 corner and seal every line except b2,
	// leaving the winning placement to the player.
	for _, m := range []game.Move{
		game.PlaceMove(at(t, "c1")), game.StepMove(at(t, "c3"), at(t, "c2")),
		game.PlaceMove(at(t, "c3")), game.StepMove(at(t, "c2"), at(t, "b2")),
		game.PlaceMove(at(t, "a3")), game.StepMove(at(t, "b2"), at(t, "a1")),
		game.PlaceMove(at(t, "b1")), game.StepMove(at(t, "a1"), at(t, "b2")),
		game.PlaceMove(at(t, "a2")), game.StepMove(at(t, "b2"), at(t, "a1")),
	} {
		if err := e.Play(m); err != nil {
			t.Fatalf("fixture move %v: %v", m, err)
		}
	}
	c.sync()

	c.TapPoint(at(t, "b2")) // the fence closes

	if h.overs != 1 {
		t.Fatalf("game over announced %d times, want 1 (status %v)", h.overs, c.State().Status)
	}
	if h.status != game.GoatsWon || h.winner != game.GoatWinner {
		t.Errorf("announced %v / %v, want GoatsWon / Goats", h.status, h.winner)
	}

	// Further clicks must be refused, not crash.
	c.TapPoint(at(t, "e5"))
	if h.overs != 1 {
		t.Error("game over was announced again after the game ended")
	}
}

func TestNewGameWithSwitchesSides(t *testing.T) {
	c, _, _, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	c.TapPoint(at(t, "a1"))

	c.NewGameWith(PlayerVsPlayer, game.TigerPlayer, ai.Hard, game.DefaultRuleSet(), 1)

	if c.Mode() != PlayerVsPlayer || c.HumanSide() != game.TigerPlayer {
		t.Error("settings were not applied")
	}
	if c.Difficulty() != ai.Hard {
		t.Error("difficulty was not applied")
	}
	if c.State().GoatsPlaced != 0 {
		t.Error("the new game kept the old position")
	}
}

func TestSaveAndLoadRoundTripThroughTheController(t *testing.T) {
	c, _, _, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	c.TapPoint(at(t, "a1"))
	c.TapPoint(at(t, "c3"))
	c.TapPoint(at(t, "b3"))
	want := c.State()

	data, err := game.EncodeSnapshot(c.Engine().Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	snap, err := game.DecodeSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}

	other, _, _, _ := newTestController(t, PlayerVsPlayer, game.GoatPlayer)
	if err := other.LoadSnapshot(snap); err != nil {
		t.Fatalf("load: %v", err)
	}
	if other.State() != want {
		t.Error("the loaded game differs from the saved one")
	}
}

func TestFriendlyErrorCoversEveryEngineError(t *testing.T) {
	cases := []error{
		game.ErrInvalidMove, game.ErrOccupiedPosition, game.ErrWrongTurn,
		game.ErrGameOver, game.ErrOutOfBounds, game.ErrNothingToUndo,
		game.ErrNothingToRedo,
	}
	for _, err := range cases {
		msg := friendlyError(err)
		if msg == "" || msg == err.Error() {
			t.Errorf("%v has no friendly message (got %q)", err, msg)
		}
	}
	if friendlyError(nil) != "" {
		t.Error("nil should map to an empty message")
	}
}
