// Package cli provides a text frontend for Bagh-Bandi.
//
// It exists so the complete game can be played and debugged without a GUI,
// and it uses exactly the same engine as the Fyne frontend - there is no
// second copy of the rules anywhere in this program.
package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"bagh-bandi/internal/ai"
	"bagh-bandi/internal/game"
)

// Options configures a CLI session.
type Options struct {
	// Rules is the ruleset to play under.
	Rules game.RuleSet
	// TigerAI, when set, lets the computer play the tiger.
	TigerAI ai.Player
	// GoatAI, when set, lets the computer play the goats.
	GoatAI ai.Player
	// Think is the AI time budget per move.
	Think time.Duration
	// In and Out default to stdin/stdout.
	In  io.Reader
	Out io.Writer
}

// DefaultOptions returns a two-human CLI session on the default rules.
func DefaultOptions() Options {
	return Options{Rules: game.DefaultRuleSet(), Think: 3 * time.Second}
}

// Session is a running CLI game.
type Session struct {
	engine  *game.Engine
	opts    Options
	out     *bufio.Writer
	in      *bufio.Scanner
	quit    bool
	lastErr error
}

// NewSession creates a CLI session.
func NewSession(opts Options) *Session {
	if opts.In == nil {
		opts.In = os.Stdin
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Think <= 0 {
		opts.Think = 3 * time.Second
	}
	if opts.Rules.TotalGoats == 0 {
		opts.Rules = game.DefaultRuleSet()
	}
	return &Session{
		engine: game.NewEngine(opts.Rules),
		opts:   opts,
		out:    bufio.NewWriter(opts.Out),
		in:     bufio.NewScanner(opts.In),
	}
}

// Engine exposes the underlying engine, mainly for tests.
func (s *Session) Engine() *game.Engine { return s.engine }

const banner = `
বাঘবন্দী  -  Bagh-Bandi (Tiger & Goats)
1 tiger, 16 goats, Bangladeshi rules.

Type "help" for commands.
`

// Run reads commands until EOF or "quit".
func (s *Session) Run() error {
	s.printf("%s\n", banner)
	s.showBoard()

	for !s.quit {
		s.playAIIfItsTurn()
		// With both sides played by the computer there is nothing left to
		// prompt for once the game ends. showBoard has already printed the
		// result, so just leave.
		if s.engine.IsGameOver() && s.opts.TigerAI != nil && s.opts.GoatAI != nil {
			break
		}
		s.printf("%s> ", s.engine.State().Turn)
		s.flush()

		if !s.in.Scan() {
			s.printf("\n")
			break
		}
		if err := s.Execute(strings.TrimSpace(s.in.Text())); err != nil {
			s.printf("error: %v\n", err)
		}
		s.flush()
	}
	s.flush()
	return s.in.Err()
}

// Execute runs a single command line. It is exported so tests can drive a
// session without any terminal.
func (s *Session) Execute(line string) error {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}
	cmd, args := strings.ToLower(fields[0]), fields[1:]

	switch cmd {
	case "help", "?":
		s.showHelp()
	case "show", "board", "b":
		s.showBoard()
	case "status", "st":
		s.showStatus()
	case "moves", "legal", "l":
		s.showMoves(args)
	case "move", "m", "play", "p":
		return s.doMove(args)
	case "place":
		return s.doPlace(args)
	case "undo", "u":
		if err := s.engine.Undo(); err != nil {
			return err
		}
		s.showBoard()
	case "redo":
		if err := s.engine.Redo(); err != nil {
			return err
		}
		s.showBoard()
	case "restart", "new":
		s.engine.Restart()
		s.showBoard()
	case "history", "h":
		s.showHistory()
	case "quit", "exit", "q":
		s.quit = true
	default:
		return fmt.Errorf("unknown command %q - type \"help\"", cmd)
	}
	return nil
}

func (s *Session) showHelp() {
	s.printf(`Commands:
  show                 draw the board
  status               turn, goat counts, result
  moves [point]        list legal moves, optionally only from a point
  move <from> <to>     move a piece, e.g. "move c3 b3" (captures included)
  place <point>        place a goat during the placement phase, e.g. "place a1"
  undo / redo          step through the move history
  restart              start a new game
  history              list the moves played
  quit                 leave

Points are named a1 (top-left) through e5 (bottom-right).
`)
}

// showBoard draws the Alquerque grid with its lines, so the topology is
// visible rather than implied.
func (s *Session) showBoard() {
	st := s.engine.State()
	b := s.engine.Board()

	s.printf("\n    a   b   c   d   e\n")
	for row := 0; row < game.BoardSize; row++ {
		s.printf("%d  ", row+1)
		for col := 0; col < game.BoardSize; col++ {
			s.printf("%s", glyph(st.Occupancy[game.ID(row, col)]))
			if col < game.BoardSize-1 {
				s.printf(" - ")
			}
		}
		s.printf("  %d\n", row+1)

		if row < game.BoardSize-1 {
			s.printf("   ")
			for col := 0; col < game.BoardSize; col++ {
				s.printf("|")
				if col < game.BoardSize-1 {
					// A diagonal cross is drawn only where the lines exist.
					if b.AreAdjacent(game.ID(row, col), game.ID(row+1, col+1)) {
						s.printf(" X ")
					} else {
						s.printf("   ")
					}
				}
			}
			s.printf("\n")
		}
	}
	s.printf("    a   b   c   d   e\n\n")
	s.showStatus()
}

func glyph(p game.PieceType) string {
	switch p {
	case game.Tiger:
		return "T"
	case game.Goat:
		return "g"
	default:
		return "."
	}
}

func (s *Session) showStatus() {
	st := s.engine.State()
	cfg := s.engine.Config()

	if st.IsOver() {
		s.printf("Game over: %s", st.Status)
		if w := s.engine.Winner(); w != game.NoWinner {
			s.printf(" (%s)", w)
		}
		s.printf("\n")
		return
	}

	phase := "movement"
	if st.InPlacementPhase(&cfg) {
		phase = "placement"
	}
	s.printf("Turn: %-5s | phase: %-9s | goats in hand: %2d | on board: %2d | captured: %d/%d\n",
		st.Turn, phase, st.GoatsInHand(&cfg), st.GoatsOnBoard(), st.GoatsCaptured, cfg.TigerWinCaptures)
	if s.lastErr != nil {
		s.printf("last error: %v\n", s.lastErr)
	}
}

func (s *Session) showMoves(args []string) {
	moves := s.engine.LegalMoves()
	if len(args) > 0 {
		from, ok := game.ParseNotation(args[0])
		if !ok {
			s.printf("bad point %q\n", args[0])
			return
		}
		moves = s.engine.LegalMovesFrom(from)
	}
	if len(moves) == 0 {
		s.printf("no legal moves\n")
		return
	}
	strs := make([]string, len(moves))
	for i, m := range moves {
		strs[i] = m.String()
	}
	sort.Strings(strs)
	s.printf("%d legal moves: %s\n", len(strs), strings.Join(strs, ", "))
}

func (s *Session) doPlace(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: place <point>")
	}
	to, ok := game.ParseNotation(args[0])
	if !ok {
		return fmt.Errorf("bad point %q", args[0])
	}
	if err := s.engine.Play(game.PlaceMove(to)); err != nil {
		return err
	}
	s.showBoard()
	return nil
}

// doMove accepts "move c3 b3" and works out whether that is a slide or a
// jump by asking the engine, so the user never has to spell out captures.
func (s *Session) doMove(args []string) error {
	if len(args) == 1 {
		// "move a1" during placement is a convenient shorthand.
		return s.doPlace(args)
	}
	if len(args) != 2 {
		return fmt.Errorf("usage: move <from> <to>")
	}
	from, ok := game.ParseNotation(args[0])
	if !ok {
		return fmt.Errorf("bad point %q", args[0])
	}
	to, ok := game.ParseNotation(args[1])
	if !ok {
		return fmt.Errorf("bad point %q", args[1])
	}

	for _, m := range s.engine.LegalMovesFrom(from) {
		if m.To == to {
			if err := s.engine.Play(m); err != nil {
				return err
			}
			s.printf("played %s\n", m)
			s.showBoard()
			return nil
		}
	}
	return fmt.Errorf("%w: %s to %s", game.ErrInvalidMove, game.Notation(from), game.Notation(to))
}

func (s *Session) showHistory() {
	h := s.engine.History()
	if len(h) == 0 {
		s.printf("no moves played\n")
		return
	}
	first := s.engine.Config().FirstPlayer
	for i, m := range h {
		side := first
		if i%2 == 1 {
			side = first.Opponent()
		}
		s.printf("%3d. %-6s %s\n", i+1, side, m)
	}
}

// playAIIfItsTurn lets a configured computer player move, possibly several
// times in a row when both sides are computers.
func (s *Session) playAIIfItsTurn() {
	for !s.engine.IsGameOver() {
		var p ai.Player
		if s.engine.State().Turn == game.TigerPlayer {
			p = s.opts.TigerAI
		} else {
			p = s.opts.GoatAI
		}
		if p == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), s.opts.Think)
		m, ok := p.ChooseMove(ctx, s.engine.Rules(), s.engine.State())
		cancel()
		if !ok {
			return
		}
		if err := s.engine.Play(m); err != nil {
			s.printf("AI produced an illegal move %v: %v\n", m, err)
			return
		}
		s.printf("%s (%s) plays %s\n", s.engine.State().Turn.Opponent(), p.Name(), m)
		s.showBoard()
	}
}

func (s *Session) printf(format string, args ...any) {
	fmt.Fprintf(s.out, format, args...)
}

func (s *Session) flush() { _ = s.out.Flush() }
