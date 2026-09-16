package ui

import (
	"context"
	"sync"
	"time"

	"bagh-bandi/internal/ai"
	"bagh-bandi/internal/game"
)

// Mode is who plays which side.
type Mode int

const (
	// PlayerVsPlayer is two humans on one machine.
	PlayerVsPlayer Mode = iota
	// PlayerVsComputer gives one side to the AI.
	PlayerVsComputer
)

// BoardPresenter is the slice of the board widget the controller needs. It is
// an interface so the controller's whole click-handling logic can be tested
// without a graphics driver.
type BoardPresenter interface {
	SetState(game.GameState)
	SetSelection(game.PositionID, []game.Move)
	SetLastMove(*game.Move)
	FlashInvalid(game.PositionID)
	AnimateMove(game.Move, func())
}

// HUD receives everything the surrounding chrome needs to display.
type HUD interface {
	// Update refreshes the turn indicator, counters and buttons.
	Update(Controller)
	// Message shows a transient line of feedback to the player.
	Message(text string, isError bool)
	// GameOver announces the result once, when a game finishes.
	GameOver(status game.GameStatus, winner game.Winner)
}

// SoundPlayer plays an optional effect. Any implementation must be safe to
// call from any goroutine and must never block.
type SoundPlayer interface {
	Play(effect Effect)
}

// Controller is the read-only view of a game that the HUD renders.
type Controller interface {
	State() game.GameState
	Config() game.RuleSet
	Mode() Mode
	HumanSide() game.Player
	Thinking() bool
	CanUndo() bool
	CanRedo() bool
	Difficulty() ai.Difficulty
}

// GameController wires the board widget, the engine and the AI together.
//
// All engine mutation happens on the UI goroutine. The AI runs on its own
// goroutine and hands its answer back through apply, so the interface never
// freezes while the computer is thinking.
type GameController struct {
	engine *game.Engine
	view   BoardPresenter
	hud    HUD
	sound  SoundPlayer

	mode       Mode
	humanSide  game.Player
	difficulty ai.Difficulty
	opponent   ai.Player
	thinkTime  time.Duration

	selected game.PositionID

	mu       sync.Mutex
	thinking bool
	cancelAI context.CancelFunc

	// runOnMain marshals a callback onto the UI goroutine. Fyne supplies
	// fyne.Do; tests supply a direct call.
	runOnMain func(func())
	// announced stops the game-over dialog firing more than once.
	announced bool
}

// ControllerOptions configures a new controller.
type ControllerOptions struct {
	Rules      game.RuleSet
	Mode       Mode
	HumanSide  game.Player
	Difficulty ai.Difficulty
	ThinkTime  time.Duration
	Seed       int64
	View       BoardPresenter
	HUD        HUD
	Sound      SoundPlayer
	RunOnMain  func(func())
}

// NewGameController starts a new game.
func NewGameController(opts ControllerOptions) *GameController {
	if opts.Rules.TotalGoats == 0 {
		opts.Rules = game.DefaultRuleSet()
	}
	if opts.ThinkTime <= 0 {
		opts.ThinkTime = 2 * time.Second
	}
	if opts.HumanSide == 0 {
		opts.HumanSide = game.GoatPlayer
	}
	if opts.RunOnMain == nil {
		opts.RunOnMain = func(f func()) { f() }
	}

	c := &GameController{
		engine:     game.NewEngine(opts.Rules),
		view:       opts.View,
		hud:        opts.HUD,
		sound:      opts.Sound,
		mode:       opts.Mode,
		humanSide:  opts.HumanSide,
		difficulty: opts.Difficulty,
		opponent:   ai.NewPlayer(opts.Difficulty, opts.Seed),
		thinkTime:  opts.ThinkTime,
		selected:   game.NoPosition,
		runOnMain:  opts.RunOnMain,
	}
	c.sync()
	c.maybeStartAI()
	return c
}

// ---- Controller (read-only view for the HUD) ----

// State returns the current position.
func (c *GameController) State() game.GameState { return c.engine.State() }

// Config returns the ruleset in force.
func (c *GameController) Config() game.RuleSet { return c.engine.Config() }

// Mode returns whether the computer is playing.
func (c *GameController) Mode() Mode { return c.mode }

// HumanSide returns the side the local player controls in PlayerVsComputer.
func (c *GameController) HumanSide() game.Player { return c.humanSide }

// Difficulty returns the AI strength preset.
func (c *GameController) Difficulty() ai.Difficulty { return c.difficulty }

// Thinking reports whether the AI is searching.
func (c *GameController) Thinking() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.thinking
}

// CanUndo reports whether undo is available to the player right now.
func (c *GameController) CanUndo() bool { return c.engine.CanUndo() && !c.Thinking() }

// CanRedo reports whether redo is available.
func (c *GameController) CanRedo() bool { return c.engine.CanRedo() && !c.Thinking() }

// Engine exposes the engine for save/load and for tests.
func (c *GameController) Engine() *game.Engine { return c.engine }

// ---- input ----

// isHumanTurn reports whether the local player may act.
func (c *GameController) isHumanTurn() bool {
	if c.engine.IsGameOver() || c.Thinking() {
		return false
	}
	if c.mode == PlayerVsPlayer {
		return true
	}
	return c.engine.State().Turn == c.humanSide
}

// TapPoint is the single entry point for board clicks. It translates a point
// into a move by asking the engine what is legal - it never decides legality
// itself.
func (c *GameController) TapPoint(id game.PositionID) {
	if !c.isHumanTurn() {
		c.reject(id, "Not your turn yet.")
		return
	}

	s := c.engine.State()
	cfg := c.engine.Config()

	// Placement phase: a click on any empty point drops a goat.
	if s.Turn == game.GoatPlayer && s.InPlacementPhase(&cfg) {
		c.tryPlay(game.PlaceMove(id), id)
		return
	}

	// A click on a highlighted destination completes the selected move.
	if c.selected != game.NoPosition {
		for _, m := range c.engine.LegalMovesFrom(c.selected) {
			if m.To == id {
				c.tryPlay(m, id)
				return
			}
		}
	}

	// Otherwise treat the click as selecting one of your own pieces.
	if s.Occupancy[id] == s.Turn.Piece() {
		moves := c.engine.LegalMovesFrom(id)
		if len(moves) == 0 {
			c.reject(id, "That piece has nowhere to go.")
			return
		}
		c.selected = id
		c.view.SetSelection(id, moves)
		c.hud.Message("", false)
		return
	}

	if c.selected != game.NoPosition {
		c.clearSelection()
		c.reject(id, "That is not a legal destination.")
		return
	}
	c.reject(id, "Select one of your own pieces.")
}

func (c *GameController) tryPlay(m game.Move, clicked game.PositionID) {
	if err := c.engine.Play(m); err != nil {
		c.reject(clicked, friendlyError(err))
		return
	}
	c.clearSelection()
	c.afterMove(m)
}

// afterMove updates everything that follows a successful move, then lets the
// AI reply.
func (c *GameController) afterMove(m game.Move) {
	c.playSound(soundFor(m))
	c.view.SetState(c.engine.State())
	last := m
	c.view.SetLastMove(&last)
	c.hud.Update(c)

	c.view.AnimateMove(m, func() {
		c.hud.Update(c)
		if c.announceIfOver() {
			return
		}
		c.maybeStartAI()
	})
}

func (c *GameController) announceIfOver() bool {
	s := c.engine.State()
	if !s.IsOver() {
		return false
	}
	if c.announced {
		return true
	}
	c.announced = true

	winner := c.engine.Winner()
	switch {
	case c.mode == PlayerVsComputer && winner != game.NoWinner:
		if (winner == game.TigerWinner) == (c.humanSide == game.TigerPlayer) {
			c.playSound(EffectVictory)
		} else {
			c.playSound(EffectDefeat)
		}
	case winner != game.NoWinner:
		c.playSound(EffectVictory)
	default:
		c.playSound(EffectDefeat)
	}
	c.hud.GameOver(s.Status, winner)
	return true
}

// ---- AI ----

// maybeStartAI launches a search if it is the computer's turn.
func (c *GameController) maybeStartAI() {
	if c.mode != PlayerVsComputer || c.engine.IsGameOver() {
		return
	}
	if c.engine.State().Turn == c.humanSide {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.thinkTime)
	c.mu.Lock()
	c.thinking = true
	c.cancelAI = cancel
	c.mu.Unlock()
	c.hud.Update(c)

	rules := c.engine.Rules()
	state := c.engine.State()
	opponent := c.opponent

	go func() {
		defer cancel()
		m, ok := opponent.ChooseMove(ctx, rules, state)
		c.runOnMain(func() {
			c.mu.Lock()
			c.thinking = false
			c.cancelAI = nil
			c.mu.Unlock()

			// The player may have restarted or undone while we searched;
			// only apply the move if the position is still the one we
			// searched from.
			if !ok || c.engine.State() != state {
				c.hud.Update(c)
				return
			}
			if err := c.engine.Play(m); err != nil {
				c.hud.Message("The computer could not move: "+friendlyError(err), true)
				c.hud.Update(c)
				return
			}
			c.afterMove(m)
		})
	}()
}

// stopAI cancels any running search and waits for the goroutine to notice.
func (c *GameController) stopAI() {
	c.mu.Lock()
	cancel := c.cancelAI
	c.cancelAI = nil
	c.thinking = false
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// ---- commands ----

// Undo takes back the local player's last move, stepping over the computer's
// reply so the player is always handed their own turn.
func (c *GameController) Undo() {
	c.stopAI()
	var err error
	if c.mode == PlayerVsComputer {
		err = c.engine.UndoTurn(c.humanSide)
	} else {
		err = c.engine.Undo()
	}
	if err != nil {
		c.hud.Message(friendlyError(err), true)
		return
	}
	c.announced = false
	c.clearSelection()
	c.sync()
	c.maybeStartAI()
}

// Redo steps forward through the history.
func (c *GameController) Redo() {
	c.stopAI()
	if err := c.engine.Redo(); err != nil {
		c.hud.Message(friendlyError(err), true)
		return
	}
	c.announced = false
	c.clearSelection()
	c.sync()
}

// Restart begins a fresh game under the same settings.
func (c *GameController) Restart() {
	c.stopAI()
	c.engine.Restart()
	c.announced = false
	c.clearSelection()
	c.sync()
	c.hud.Message("New game. "+openingHint(c.engine.Config()), false)
	c.maybeStartAI()
}

// NewGameWith restarts under different settings.
func (c *GameController) NewGameWith(mode Mode, humanSide game.Player, difficulty ai.Difficulty, rules game.RuleSet, seed int64) {
	c.stopAI()
	c.mode = mode
	c.humanSide = humanSide
	c.difficulty = difficulty
	c.opponent = ai.NewPlayer(difficulty, seed)
	c.engine = game.NewEngine(rules)
	c.announced = false
	c.clearSelection()
	c.sync()
	c.maybeStartAI()
}

// LoadSnapshot restores a saved game.
func (c *GameController) LoadSnapshot(snap game.Snapshot) error {
	c.stopAI()
	if err := c.engine.LoadSnapshot(snap); err != nil {
		return err
	}
	c.announced = false
	c.clearSelection()
	c.sync()
	c.maybeStartAI()
	return nil
}

// SetThinkTime changes the AI time budget.
func (c *GameController) SetThinkTime(d time.Duration) {
	if d > 0 {
		c.thinkTime = d
	}
}

// Close stops any background search. Call it when the window closes.
func (c *GameController) Close() { c.stopAI() }

// ---- helpers ----

func (c *GameController) clearSelection() {
	c.selected = game.NoPosition
	c.view.SetSelection(game.NoPosition, nil)
}

// sync pushes the whole engine state out to the view and the HUD.
func (c *GameController) sync() {
	c.view.SetState(c.engine.State())
	if m, ok := c.engine.LastMove(); ok {
		last := m
		c.view.SetLastMove(&last)
	} else {
		c.view.SetLastMove(nil)
	}
	c.hud.Update(c)
}

func (c *GameController) reject(id game.PositionID, message string) {
	c.playSound(EffectInvalid)
	c.view.FlashInvalid(id)
	c.hud.Message(message, true)
}

func (c *GameController) playSound(e Effect) {
	if c.sound != nil {
		c.sound.Play(e)
	}
}

func soundFor(m game.Move) Effect {
	switch m.Kind {
	case game.Placement:
		return EffectPlace
	case game.Capture:
		return EffectCapture
	default:
		return EffectMove
	}
}

func openingHint(cfg game.RuleSet) string {
	if cfg.FirstPlayer == game.GoatPlayer {
		return "Goats place first - tap any empty point."
	}
	return "The tiger moves first."
}

// friendlyError turns an engine error into something a player can act on.
func friendlyError(err error) string {
	switch {
	case err == nil:
		return ""
	case errorsIs(err, game.ErrOccupiedPosition):
		return "That point is already taken."
	case errorsIs(err, game.ErrWrongTurn):
		return "That piece is not yours to move."
	case errorsIs(err, game.ErrGameOver):
		return "The game is over - start a new one."
	case errorsIs(err, game.ErrOutOfBounds):
		return "That is not a point on the board."
	case errorsIs(err, game.ErrNothingToUndo):
		return "There is nothing to undo."
	case errorsIs(err, game.ErrNothingToRedo):
		return "There is nothing to redo."
	case errorsIs(err, game.ErrInvalidMove):
		return "That move is not allowed."
	default:
		return err.Error()
	}
}

var _ Controller = (*GameController)(nil)
