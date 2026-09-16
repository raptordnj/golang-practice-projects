package ui

import (
	"fmt"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"bagh-bandi/internal/ai"
	"bagh-bandi/internal/game"
)

// AppID is the reverse-DNS identifier Fyne uses for preferences.
const AppID = "com.baghbandi.game"

// appDir is the folder under the user's config directory.
const appDir = "bagh-bandi"

// App is the whole desktop application: window, menu screen, game screen.
type App struct {
	fyneApp fyne.App
	window  fyne.Window

	storage  *Storage
	settings Settings
	sound    *FyneSound

	board      *BoardView
	controller *GameController

	// HUD widgets.
	turnLabel    *widget.Label
	phaseLabel   *widget.Label
	goatLabel    *widget.Label
	capturedBar  *widget.ProgressBar
	statusLabel  *widget.Label
	messageLabel *widget.Label
	undoBtn      *widget.Button
	redoBtn      *widget.Button

	menuScreen fyne.CanvasObject
	gameScreen fyne.CanvasObject
	content    *fyne.Container
}

// NewApp builds the application. Call Run to show it.
func NewApp() *App {
	return newApp(fyneapp.NewWithID(AppID), NewStorage(appDir))
}

// newApp builds the application against a supplied Fyne app and storage, so
// tests can drive the real window and dialogs without touching the user's
// desktop or their configuration directory.
func newApp(fyneApp fyne.App, store *Storage) *App {
	a := &App{fyneApp: fyneApp}
	a.storage = store
	a.settings = a.storage.LoadSettings()
	a.sound = NewFyneSound(a.settings.Sound, a.loadSound)

	a.fyneApp.Settings().SetTheme(newAppTheme(a.settings.Theme()))
	a.window = a.fyneApp.NewWindow("বাঘবন্দী - Bagh-Bandi")
	a.window.Resize(fyne.NewSize(1000, 780))
	a.window.SetMaster()

	a.buildGameScreen()
	a.buildMenuScreen()

	a.content = container.NewStack(a.menuScreen)
	a.window.SetContent(a.content)
	a.window.SetCloseIntercept(a.onClose)
	return a
}

// Run shows the window and blocks until it closes.
func (a *App) Run() {
	a.window.ShowAndRun()
}

// loadSound resolves a sound asset from disk. Missing files are the normal
// case - the repository ships no audio - so the error is simply passed back
// and the sound player stays silent.
func (a *App) loadSound(name string) (fyne.Resource, error) {
	return fyne.LoadResourceFromPath(filepath.Join("assets", filepath.FromSlash(name)))
}

// ---- menu screen ----

func (a *App) buildMenuScreen() {
	title := widget.NewLabelWithStyle("বাঘবন্দী", fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true})
	title.SizeName = theme.SizeNameHeadingText

	subtitle := widget.NewLabelWithStyle(
		"Bagh-Bandi  ·  1 tiger, 16 goats  ·  Bangladeshi rules",
		fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	pvp := widget.NewButtonWithIcon("Player vs Player  (দুইজন খেলোয়াড়)",
		theme.AccountIcon(), func() {
			a.startGame(PlayerVsPlayer, a.settings.HumanSide)
		})
	pvp.Importance = widget.HighImportance

	pvc := widget.NewButtonWithIcon("Player vs Computer  (কম্পিউটারের সাথে)",
		theme.ComputerIcon(), func() {
			a.startGame(PlayerVsComputer, a.settings.HumanSide)
		})
	pvc.Importance = widget.HighImportance

	resume := widget.NewButtonWithIcon("Resume last game", theme.MediaPlayIcon(), a.resumeAutosave)
	if _, err := a.storage.LoadGame(AutosaveName); err != nil {
		resume.Disable()
	}

	settings := widget.NewButtonWithIcon("Settings  (সেটিংস)", theme.SettingsIcon(), a.showSettings)
	rules := widget.NewButtonWithIcon("How to play  (নিয়ম)", theme.HelpIcon(), a.showRules)
	quit := widget.NewButtonWithIcon("Exit", theme.LogoutIcon(), func() { a.onClose() })

	buttons := container.NewVBox(pvp, pvc, resume, settings, rules, quit)

	a.menuScreen = container.NewCenter(container.NewVBox(
		title, subtitle, widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(380, 44),
			buttons.Objects...),
	))
}

// ---- game screen ----

func (a *App) buildGameScreen() {
	a.board = NewBoardView(a.settings.Theme(), a.settings.Animation)
	a.board.OnPointTapped = func(id game.PositionID) {
		if a.controller != nil {
			a.controller.TapPoint(id)
		}
	}

	a.turnLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	a.turnLabel.SizeName = theme.SizeNameSubHeadingText
	a.phaseLabel = widget.NewLabel("")
	a.goatLabel = widget.NewLabel("")
	a.statusLabel = widget.NewLabel("")
	a.messageLabel = widget.NewLabel("")
	a.messageLabel.Wrapping = fyne.TextWrapWord
	a.capturedBar = widget.NewProgressBar()
	a.capturedBar.TextFormatter = func() string {
		cfg := game.DefaultRuleSet()
		if a.controller != nil {
			cfg = a.controller.Config()
		}
		captured := 0
		if a.controller != nil {
			captured = a.controller.State().GoatsCaptured
		}
		return fmt.Sprintf("Goats eaten %d / %d", captured, cfg.TigerWinCaptures)
	}

	a.undoBtn = widget.NewButtonWithIcon("Undo", theme.ContentUndoIcon(), func() { a.controller.Undo() })
	a.redoBtn = widget.NewButtonWithIcon("Redo", theme.ContentRedoIcon(), func() { a.controller.Redo() })
	restart := widget.NewButtonWithIcon("Restart", theme.ViewRefreshIcon(), func() { a.controller.Restart() })
	save := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), a.showSaveDialog)
	load := widget.NewButtonWithIcon("Load", theme.FolderOpenIcon(), a.showLoadDialog)
	menu := widget.NewButtonWithIcon("Menu", theme.HomeIcon(), a.showMenu)

	side := container.NewVBox(
		a.turnLabel,
		a.phaseLabel,
		a.goatLabel,
		a.capturedBar,
		widget.NewSeparator(),
		a.statusLabel,
		a.messageLabel,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, a.undoBtn, a.redoBtn),
		restart,
		container.NewGridWithColumns(2, save, load),
		menu,
	)

	a.gameScreen = container.NewBorder(nil, nil, nil,
		container.NewVScroll(container.NewPadded(side)),
		container.NewPadded(a.board))
}

// startGame creates a controller and switches to the board.
func (a *App) startGame(mode Mode, humanSide game.Player) {
	a.closeController()

	a.board.SetTheme(a.settings.Theme())
	a.board.SetAnimate(a.settings.Animation)
	a.sound.SetEnabled(a.settings.Sound)

	a.controller = NewGameController(ControllerOptions{
		Rules:      a.settings.Rules(),
		Mode:       mode,
		HumanSide:  humanSide,
		Difficulty: a.settings.Difficulty,
		ThinkTime:  a.settings.ThinkTime(),
		Seed:       time.Now().UnixNano(),
		View:       a.board,
		HUD:        a,
		Sound:      a.sound,
		RunOnMain:  fyne.Do,
	})
	a.showGame()
}

func (a *App) showGame() {
	a.content.Objects = []fyne.CanvasObject{a.gameScreen}
	a.content.Refresh()
}

func (a *App) showMenu() {
	a.buildMenuScreen() // rebuild so "Resume" reflects the current autosave
	a.content.Objects = []fyne.CanvasObject{a.menuScreen}
	a.content.Refresh()
}

func (a *App) closeController() {
	if a.controller != nil {
		a.controller.Close()
	}
}

// ---- HUD implementation ----

// Update implements HUD.
func (a *App) Update(c Controller) {
	s := c.State()
	cfg := c.Config()

	turn := fmt.Sprintf("Turn: %s (%s)", s.Turn, s.Turn.Bengali())
	if c.Thinking() {
		turn = fmt.Sprintf("%s is thinking…", c.Difficulty())
	}
	a.turnLabel.SetText(turn)

	phase := "Movement phase"
	if s.InPlacementPhase(&cfg) {
		phase = "Placement phase"
	}
	if c.Mode() == PlayerVsComputer {
		phase += fmt.Sprintf("  ·  you play the %s", c.HumanSide().String())
	}
	a.phaseLabel.SetText(phase)

	a.goatLabel.SetText(fmt.Sprintf(
		"Goats in hand: %d\nGoats on board: %d",
		s.GoatsInHand(&cfg), s.GoatsOnBoard()))

	if cfg.TigerWinCaptures > 0 {
		a.capturedBar.SetValue(float64(s.GoatsCaptured) / float64(cfg.TigerWinCaptures))
	}
	a.capturedBar.Refresh()

	a.statusLabel.SetText("Status: " + s.Status.String())

	setEnabled(a.undoBtn, c.CanUndo())
	setEnabled(a.redoBtn, c.CanRedo())
}

// Message implements HUD.
func (a *App) Message(text string, isError bool) {
	if isError && text != "" {
		a.messageLabel.Importance = widget.DangerImportance
	} else {
		a.messageLabel.Importance = widget.MediumImportance
	}
	a.messageLabel.SetText(text)
}

// GameOver implements HUD.
func (a *App) GameOver(status game.GameStatus, winner game.Winner) {
	title := "Game over"
	body := status.String()
	switch winner {
	case game.TigerWinner:
		title = "বাঘ জিতেছে - the tiger wins"
		body = "The tiger ate enough goats to break the fence."
	case game.GoatWinner:
		title = "ছাগল জিতেছে - the goats win"
		body = "The tiger is fenced in and cannot move."
	default:
		body = "Neither side made progress - the game is a draw."
	}
	dialog.ShowCustomConfirm(title, "New game", "Keep looking",
		widget.NewLabel(body), func(again bool) {
			if again {
				a.controller.Restart()
			}
		}, a.window)
}

// ---- dialogs ----

func (a *App) showSettings() {
	difficulty := widget.NewSelect([]string{"Easy", "Medium", "Hard"}, nil)
	difficulty.SetSelected(a.settings.Difficulty.String())

	side := widget.NewSelect([]string{"Goats", "Tiger"}, nil)
	side.SetSelected(map[game.Player]string{game.GoatPlayer: "Goats", game.TigerPlayer: "Tiger"}[a.settings.HumanSide])

	themeNames := make([]string, len(BoardThemes))
	for i, t := range BoardThemes {
		themeNames[i] = t.Name
	}
	boardTheme := widget.NewSelect(themeNames, nil)
	boardTheme.SetSelected(a.settings.Theme().Name)

	sound := widget.NewCheck("Sound effects", nil)
	sound.SetChecked(a.settings.Sound)
	animation := widget.NewCheck("Animate moves", nil)
	animation.SetChecked(a.settings.Animation)
	forced := widget.NewCheck("Capture is compulsory for the tiger", nil)
	forced.SetChecked(a.settings.ForcedCapture)

	captures := widget.NewSlider(3, 12)
	captures.Step = 1
	captures.SetValue(float64(a.settings.TigerWinCaptures))
	capturesLabel := widget.NewLabel("")
	updateCaptures := func(v float64) {
		capturesLabel.SetText(fmt.Sprintf("Tiger wins after eating %d goats", int(v)))
	}
	captures.OnChanged = updateCaptures
	updateCaptures(captures.Value)

	think := widget.NewSlider(0.5, 10)
	think.Step = 0.5
	think.SetValue(a.settings.ThinkSeconds)
	thinkLabel := widget.NewLabel("")
	updateThink := func(v float64) {
		thinkLabel.SetText(fmt.Sprintf("Computer thinking time: %.1fs", v))
	}
	think.OnChanged = updateThink
	updateThink(think.Value)

	// Label beside control rather than above it: stacked rows doubled the
	// height of the form and pushed the last section off the bottom.
	form := container.NewVBox(
		sectionHeading("Game"),
		widget.NewForm(
			widget.NewFormItem("Difficulty", difficulty),
			widget.NewFormItem("You play as", side),
			widget.NewFormItem("Board theme", boardTheme),
		),

		sectionHeading("Presentation"),
		sound,
		animation,

		sectionHeading("Rules"),
		forced,
		settingSlider(capturesLabel, captures),

		sectionHeading("Computer"),
		settingSlider(thinkLabel, think),
	)

	d := dialog.NewCustomConfirm("Settings", "Save", "Cancel",
		container.NewVScroll(container.NewPadded(form)), func(save bool) {
			if !save {
				return
			}
			a.settings.Difficulty = parseDifficulty(difficulty.Selected)
			if side.Selected == "Tiger" {
				a.settings.HumanSide = game.TigerPlayer
			} else {
				a.settings.HumanSide = game.GoatPlayer
			}
			a.settings.BoardTheme = boardTheme.Selected
			a.settings.Sound = sound.Checked
			a.settings.Animation = animation.Checked
			a.settings.ForcedCapture = forced.Checked
			a.settings.TigerWinCaptures = int(captures.Value)
			a.settings.ThinkSeconds = think.Value
			a.settings = a.settings.Normalise()
			a.applySettings()
		}, a.window)
	d.Resize(a.dialogSize(580, 700))
	d.Show()
}

// dialogSize returns a comfortable dialog size: the preferred dimensions,
// shrunk to fit if the window is smaller than that. Fyne otherwise sizes a
// dialog to the bare minimum its content will tolerate, which leaves a form
// like this one squeezed into a few scrolling centimetres.
func (a *App) dialogSize(preferWidth, preferHeight float32) fyne.Size {
	window := a.window.Canvas().Size()

	width := preferWidth
	if max := window.Width * 0.92; width > max {
		width = max
	}
	height := preferHeight
	if max := window.Height * 0.92; height > max {
		height = max
	}
	return fyne.NewSize(width, height)
}

// sectionHeading is a bold label with a rule above it, used to break the
// settings form into groups.
func sectionHeading(text string) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewVBox(widget.NewSeparator(), label)
}

// settingSlider keeps a slider and its live caption on one row, with the
// caption wide enough that its text does not jump around as the value
// changes.
func settingSlider(caption *widget.Label, slider *widget.Slider) fyne.CanvasObject {
	sized := container.NewGridWrap(fyne.NewSize(250, 34), caption)
	return container.NewBorder(nil, nil, sized, nil, slider)
}

// applySettings pushes the saved preferences into the live game.
func (a *App) applySettings() {
	if err := a.storage.SaveSettings(a.settings); err != nil && a.storage.Available() {
		a.Message("Could not save settings: "+err.Error(), true)
	}
	a.fyneApp.Settings().SetTheme(newAppTheme(a.settings.Theme()))
	a.board.SetTheme(a.settings.Theme())
	a.board.SetAnimate(a.settings.Animation)
	a.sound.SetEnabled(a.settings.Sound)

	if a.controller != nil {
		a.controller.SetThinkTime(a.settings.ThinkTime())
		dialog.ShowConfirm("Apply new rules?",
			"Difficulty, sides and rule changes take effect in a new game. Start one now?",
			func(yes bool) {
				if yes {
					a.controller.NewGameWith(a.controller.Mode(), a.settings.HumanSide,
						a.settings.Difficulty, a.settings.Rules(), time.Now().UnixNano())
				}
			}, a.window)
	}
}

func parseDifficulty(name string) ai.Difficulty {
	switch name {
	case "Hard":
		return ai.Hard
	case "Easy":
		return ai.Easy
	default:
		return ai.Medium
	}
}

func (a *App) showRules() {
	text := widget.NewRichTextFromMarkdown(RulesMarkdown)
	text.Wrapping = fyne.TextWrapWord
	d := dialog.NewCustom("How to play বাঘবন্দী", "Close",
		container.NewVScroll(container.NewPadded(text)), a.window)
	d.Resize(a.dialogSize(640, 720))
	d.Show()
}

func (a *App) showSaveDialog() {
	if a.controller == nil {
		return
	}
	entry := widget.NewEntry()
	entry.SetText(fmt.Sprintf("game %s", time.Now().Format("2006-01-02 1504")))

	dialog.ShowForm("Save game", "Save", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("Name", entry)},
		func(ok bool) {
			if !ok {
				return
			}
			path, err := a.storage.SaveGame(entry.Text, a.controller.Engine().Snapshot())
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			a.Message("Saved to "+path, false)
		}, a.window)
}

func (a *App) showLoadDialog() {
	names := a.storage.ListGames()
	if len(names) == 0 {
		dialog.ShowInformation("Load game", "There are no saved games yet.", a.window)
		return
	}

	list := widget.NewSelect(names, nil)
	list.SetSelectedIndex(0)

	dialog.ShowForm("Load game", "Load", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("Saved game", list)},
		func(ok bool) {
			if !ok || list.Selected == "" {
				return
			}
			snap, err := a.storage.LoadGame(list.Selected)
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			if a.controller == nil {
				a.startGame(PlayerVsPlayer, a.settings.HumanSide)
			}
			if err := a.controller.LoadSnapshot(snap); err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			a.showGame()
			a.Message("Loaded "+list.Selected, false)
		}, a.window)
}

func (a *App) resumeAutosave() {
	snap, err := a.storage.LoadGame(AutosaveName)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.startGame(PlayerVsPlayer, a.settings.HumanSide)
	if err := a.controller.LoadSnapshot(snap); err != nil {
		dialog.ShowError(err, a.window)
	}
}

// onClose autosaves the game in progress, then quits.
func (a *App) onClose() {
	if a.controller != nil && len(a.controller.Engine().History()) > 0 {
		_, _ = a.storage.SaveGame(AutosaveName, a.controller.Engine().Snapshot())
	}
	a.closeController()
	a.fyneApp.Quit()
}

func setEnabled(b *widget.Button, on bool) {
	if on {
		b.Enable()
		return
	}
	b.Disable()
}

var _ HUD = (*App)(nil)
