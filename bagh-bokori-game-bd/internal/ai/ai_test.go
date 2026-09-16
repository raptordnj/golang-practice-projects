package ai

import (
	"context"
	"testing"
	"time"

	"bagh-bandi/internal/game"
)

func at(t *testing.T, notation string) game.PositionID {
	t.Helper()
	id, ok := game.ParseNotation(notation)
	if !ok {
		t.Fatalf("bad notation %q", notation)
	}
	return id
}

func players() []Player {
	return []Player{
		NewRandomAI(7),
		NewMinimaxAI(1, 7),
		NewMinimaxAI(3, 7),
		NewMinimaxAI(4, 7),
	}
}

// Every AI, on every side, at every stage, must return a move the rules
// accept. This is the single most important AI property.
func TestAIAlwaysProducesLegalMoves(t *testing.T) {
	rules := game.NewDefaultRules()
	for _, p := range players() {
		p := p
		t.Run(p.Name(), func(t *testing.T) {
			s := rules.NewGame()
			for ply := 0; ply < 200 && !s.IsOver(); ply++ {
				m, ok := p.ChooseMove(context.Background(), rules, s)
				if !ok {
					t.Fatalf("ply %d: AI returned no move in a live position", ply)
				}
				if !rules.IsValidMove(s, m) {
					t.Fatalf("ply %d: AI returned illegal move %v", ply, m)
				}
				next, err := rules.ApplyMove(s, m)
				if err != nil {
					t.Fatalf("ply %d: applying AI move failed: %v", ply, err)
				}
				s = next
			}
		})
	}
}

// The search must not scribble on the caller's position.
func TestAIDoesNotMutateState(t *testing.T) {
	rules := game.NewDefaultRules()
	for _, p := range players() {
		s := rules.NewGame()
		s.Occupancy[at(t, "a1")] = game.Goat
		s.Occupancy[at(t, "b2")] = game.Goat
		s.GoatsPlaced = 2
		s.Turn = game.TigerPlayer
		before := s

		if _, ok := p.ChooseMove(context.Background(), rules, s); !ok {
			t.Fatalf("%s: no move", p.Name())
		}
		if s != before {
			t.Errorf("%s mutated the state it was given", p.Name())
		}
	}
}

func TestAIReturnsNoMoveInFinishedGame(t *testing.T) {
	rules := game.NewDefaultRules()
	s := rules.NewGame()
	s.Status = game.TigerWon
	for _, p := range players() {
		if _, ok := p.ChooseMove(context.Background(), rules, s); ok {
			t.Errorf("%s offered a move in a finished game", p.Name())
		}
	}
}

// A tiger to move with a free goat in front of it must eat it.
func TestMinimaxTakesFreeCapture(t *testing.T) {
	rules := game.NewDefaultRules()
	var s game.GameState
	s.Tiger = at(t, "c3")
	s.Occupancy[s.Tiger] = game.Tiger
	s.Occupancy[at(t, "c2")] = game.Goat
	s.GoatsPlaced = 5
	s.Turn = game.TigerPlayer

	ai := NewMinimaxAI(3, 1)
	m, ok := ai.ChooseMove(context.Background(), rules, s)
	if !ok {
		t.Fatal("no move")
	}
	if m.Kind != game.Capture || m.Over != at(t, "c2") {
		t.Errorf("minimax played %v instead of eating the goat on c2", m)
	}
}

// The tiger must not step into a square where it is immediately fenced in,
// when a safe square exists. This exercises the mobility term.
func TestMinimaxAvoidsWalkingIntoATrap(t *testing.T) {
	rules := game.NewDefaultRules()
	var s game.GameState
	// Tiger on b1. a1 is a dead end that the goats can seal next move.
	s.Tiger = at(t, "b1")
	s.Occupancy[s.Tiger] = game.Tiger
	for _, n := range []string{"a2", "b2", "c2", "d2", "e2", "c3"} {
		s.Occupancy[at(t, n)] = game.Goat
	}
	s.GoatsPlaced = 16
	s.GoatsCaptured = 10
	s.Turn = game.TigerPlayer

	ai := NewMinimaxAI(4, 1)
	m, ok := ai.ChooseMove(context.Background(), rules, s)
	if !ok {
		t.Fatal("no move")
	}
	next, err := rules.ApplyMove(s, m)
	if err != nil {
		t.Fatalf("AI move rejected: %v", err)
	}
	if next.Status == game.GoatsWon {
		t.Errorf("minimax played %v and lost immediately", m)
	}
}

// The goat side must take the win when placing a goat ends the game.
func TestMinimaxGoatPlaysTheFencingMove(t *testing.T) {
	rules := game.NewDefaultRules()
	var s game.GameState
	s.Tiger = at(t, "a1")
	s.Occupancy[s.Tiger] = game.Tiger
	for _, n := range []string{"b1", "b2", "c1", "c3", "a3"} {
		s.Occupancy[at(t, n)] = game.Goat
	}
	s.GoatsPlaced = 15
	s.Turn = game.GoatPlayer

	ai := NewMinimaxAI(2, 1)
	m, ok := ai.ChooseMove(context.Background(), rules, s)
	if !ok {
		t.Fatal("no move")
	}
	next, err := rules.ApplyMove(s, m)
	if err != nil {
		t.Fatalf("AI move rejected: %v", err)
	}
	if next.Status != game.GoatsWon {
		t.Errorf("goat AI played %v instead of the winning fence on a2", m)
	}
}

// A deeper search must be at least as good as a shallow one: depth 4 tiger
// should not lose a match against a random goat player.
func TestMinimaxBeatsRandomAsTiger(t *testing.T) {
	if testing.Short() {
		t.Skip("match play is slow")
	}
	rules := game.NewDefaultRules()
	tiger := NewMinimaxAI(4, 3)
	goat := NewRandomAI(11)

	wins := 0
	const games = 6
	for g := 0; g < games; g++ {
		s := rules.NewGame()
		for ply := 0; ply < 400 && !s.IsOver(); ply++ {
			var p Player = goat
			if s.Turn == game.TigerPlayer {
				p = tiger
			}
			m, ok := p.ChooseMove(context.Background(), rules, s)
			if !ok {
				break
			}
			var err error
			if s, err = rules.ApplyMove(s, m); err != nil {
				t.Fatalf("illegal move %v: %v", m, err)
			}
		}
		if rules.Winner(s) == game.TigerWinner {
			wins++
		}
	}
	if wins < games/2 {
		t.Errorf("minimax tiger won only %d of %d games against a random goat", wins, games)
	}
}

// A cancelled context must return promptly with a usable move.
func TestSearchRespectsContextCancellation(t *testing.T) {
	rules := game.NewDefaultRules()
	s := rules.NewGame()
	ai := NewMinimaxAI(12, 1) // far deeper than we will allow it to go

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	var m game.Move
	var ok bool
	go func() {
		m, ok = ai.ChooseMove(ctx, rules, s)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("search ignored context cancellation")
	}
	if !ok {
		t.Fatal("cancelled search returned no move")
	}
	if !rules.IsValidMove(s, m) {
		t.Errorf("cancelled search returned illegal move %v", m)
	}
	if ai.Reached >= 12 {
		t.Errorf("search reached depth %d despite the timeout", ai.Reached)
	}
}

func TestNewPlayerPresets(t *testing.T) {
	for _, d := range []Difficulty{Easy, Medium, Hard} {
		p := NewPlayer(d, 1)
		if p == nil || p.Name() == "" {
			t.Fatalf("difficulty %v produced no player", d)
		}
		if d.Bengali() == "" {
			t.Errorf("difficulty %v has no Bangla label", d)
		}
	}
}

func TestEvaluatorFavoursTigerOnCaptures(t *testing.T) {
	rules := game.NewDefaultRules()
	e := NewEvaluator(DefaultWeights())

	base := rules.NewGame()
	base.Turn = game.TigerPlayer
	fed := base
	fed.GoatsPlaced = 8
	fed.GoatsCaptured = 5

	hungry := base
	hungry.GoatsPlaced = 8

	if e.Evaluate(rules, &fed, 0) <= e.Evaluate(rules, &hungry, 0) {
		t.Error("evaluation does not reward the tiger for eating goats")
	}
}

func TestEvaluatorScoresTerminalPositions(t *testing.T) {
	rules := game.NewDefaultRules()
	e := NewEvaluator(DefaultWeights())
	s := rules.NewGame()

	s.Status = game.TigerWon
	if got := e.Evaluate(rules, &s, 2); got <= 0 {
		t.Errorf("tiger win scored %d, want a large positive", got)
	}
	s.Status = game.GoatsWon
	if got := e.Evaluate(rules, &s, 2); got >= 0 {
		t.Errorf("goat win scored %d, want a large negative", got)
	}
	s.Status = game.Draw
	if got := e.Evaluate(rules, &s, 2); got != 0 {
		t.Errorf("draw scored %d, want 0", got)
	}
}

// A quicker win must score higher than a slower one, so the AI finishes
// games instead of shuffling.
func TestMateScoresPreferQuickWins(t *testing.T) {
	rules := game.NewDefaultRules()
	e := NewEvaluator(DefaultWeights())
	s := rules.NewGame()
	s.Status = game.TigerWon
	if e.Evaluate(rules, &s, 1) <= e.Evaluate(rules, &s, 5) {
		t.Error("a win in 5 scores at least as well as a win in 1")
	}
}

func BenchmarkMinimaxDepth4(b *testing.B) {
	rules := game.NewDefaultRules()
	s := rules.NewGame()
	for _, n := range []string{"a1", "b1", "c1", "e1", "a2", "e2", "e3", "b3"} {
		id, _ := game.ParseNotation(n)
		s.Occupancy[id] = game.Goat
		s.GoatsPlaced++
	}
	s.Turn = game.TigerPlayer
	ai := NewMinimaxAI(4, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ai.ChooseMove(context.Background(), rules, s)
	}
}

func BenchmarkMinimaxDepth6(b *testing.B) {
	rules := game.NewDefaultRules()
	s := rules.NewGame()
	for _, n := range []string{"a1", "b1", "c1", "e1", "a2", "e2", "e3", "b3"} {
		id, _ := game.ParseNotation(n)
		s.Occupancy[id] = game.Goat
		s.GoatsPlaced++
	}
	s.Turn = game.TigerPlayer
	ai := NewMinimaxAI(6, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ai.ChooseMove(context.Background(), rules, s)
	}
}

func BenchmarkEvaluate(b *testing.B) {
	rules := game.NewDefaultRules()
	s := rules.NewGame()
	e := NewEvaluator(DefaultWeights())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Evaluate(rules, &s, 0)
	}
}
