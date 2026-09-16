package main

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"retro-maze/game"
)

const (
	// Game dimensions
	logicalWidth  float32 = 160
	logicalHeight float32 = 128
	windowWidth   float32 = 640
	windowHeight  float32 = 576
	gameScale     float32 = 4.0

	// Tile size
	tileSizeF float32 = 8.0

	// Timing - optimized for smooth gameplay
	playerMoveDelay    = 80 * time.Millisecond  // Faster player movement
	enemyMoveDelay     = 150 * time.Millisecond // Balanced enemy speed
	enemyMoveDelayFast = 100 * time.Millisecond // Faster when vulnerable
)

// App represents the game application
type App struct {
	fyneApp     fyne.App
	window      fyne.Window
	game        *game.Game

	// Canvas objects
	gameContainer   *fyne.Container
	background      *canvas.Rectangle
	mazeContainer   *fyne.Container
	playerObj       *canvas.Circle
	enemyObjs       []*canvas.Circle
	enemyColors     []color.Color
	pelletContainer *fyne.Container

	// UI elements
	hudContainer   *fyne.Container
	scoreText      *canvas.Text
	levelText      *canvas.Text
	livesText      *canvas.Text
	titleText      *canvas.Text
	subtitleText   *canvas.Text
	overlayText    *canvas.Text
	pauseText      *canvas.Text

	// Touch controls
	controlsContainer *fyne.Container
	upButton          *widget.Button
	downButton        *widget.Button
	leftButton        *widget.Button
	rightButton       *widget.Button

	// Timing
	lastPlayerMove time.Time
	lastEnemyMove  time.Time
	lastUpdateTime time.Time

	// Game state tracking
	currentLevel int

	// Direction buffer for smoother controls
	directionBuffer game.Direction
}

func main() {
	a := app.New()
	w := a.NewWindow("Retro Maze")
	w.Resize(fyne.NewSize(windowWidth, windowHeight))
	w.SetFixedSize(false)

	application := NewApp(a, w)
	application.Run()
}

// NewApp creates a new game application
func NewApp(a fyne.App, w fyne.Window) *App {
	app := &App{
		fyneApp:        a,
		window:         w,
		game:           game.NewGame(),
		enemyObjs:      make([]*canvas.Circle, 0),
		enemyColors:    make([]color.Color, 0),
		lastPlayerMove: time.Now(),
		lastEnemyMove:  time.Now(),
		lastUpdateTime: time.Now(),
		currentLevel:   1,
		directionBuffer: game.DirNone,
	}

	app.initCanvas()
	app.initUI()
	app.initControls()
	app.setupInput()

	return app
}

// initCanvas initializes the game canvas
func (app *App) initCanvas() {
	// Background
	app.background = canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	app.background.Resize(fyne.NewSize(windowWidth, windowHeight))

	// Maze container
	app.mazeContainer = container.NewWithoutLayout()

	// Pellet container
	app.pelletContainer = container.NewWithoutLayout()

	// Player object
	app.playerObj = canvas.NewCircle(color.RGBA{R: 255, G: 255, B: 0, A: 255})
	app.playerObj.Resize(fyne.NewSize(tileSizeF*gameScale*0.8, tileSizeF*gameScale*0.8))

	// Game container
	app.gameContainer = container.NewWithoutLayout(
		app.background,
		app.mazeContainer,
		app.pelletContainer,
		app.playerObj,
	)
}

// initUI initializes UI elements
func (app *App) initUI() {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	cyan := color.RGBA{R: 0, G: 255, B: 255, A: 255}

	// HUD
	app.scoreText = canvas.NewText("SCORE: 000000", white)
	app.scoreText.TextSize = 14
	app.scoreText.TextStyle = fyne.TextStyle{Monospace: true}
	app.scoreText.Move(fyne.NewPos(10, 10))

	app.levelText = canvas.NewText("LEVEL: 01", white)
	app.levelText.TextSize = 14
	app.levelText.TextStyle = fyne.TextStyle{Monospace: true}
	app.levelText.Move(fyne.NewPos(windowWidth-120, 10))

	app.livesText = canvas.NewText("LIVES: ♥♥♥", white)
	app.livesText.TextSize = 14
	app.livesText.TextStyle = fyne.TextStyle{Monospace: true}
	app.livesText.Move(fyne.NewPos(windowWidth/2-60, windowHeight-30))

	// Title/Overlay text
	app.titleText = canvas.NewText("RETRO MAZE", yellow)
	app.titleText.TextSize = 28
	app.titleText.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	app.titleText.Alignment = fyne.TextAlignCenter
	app.titleText.Resize(fyne.NewSize(windowWidth, 40))
	app.titleText.Move(fyne.NewPos(0, windowHeight/4))

	app.subtitleText = canvas.NewText("PRESS SPACE TO START", white)
	app.subtitleText.TextSize = 16
	app.subtitleText.TextStyle = fyne.TextStyle{Monospace: true}
	app.subtitleText.Alignment = fyne.TextAlignCenter
	app.subtitleText.Resize(fyne.NewSize(windowWidth, 30))
	app.subtitleText.Move(fyne.NewPos(0, windowHeight/2))

	app.overlayText = canvas.NewText("", white)
	app.overlayText.TextSize = 12
	app.overlayText.TextStyle = fyne.TextStyle{Monospace: true}
	app.overlayText.Alignment = fyne.TextAlignCenter
	app.overlayText.Resize(fyne.NewSize(windowWidth, 60))
	app.overlayText.Move(fyne.NewPos(0, windowHeight/2+50))

	// Pause text
	app.pauseText = canvas.NewText("PAUSED\nPRESS P TO RESUME", cyan)
	app.pauseText.TextSize = 16
	app.pauseText.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	app.pauseText.Alignment = fyne.TextAlignCenter
	app.pauseText.Resize(fyne.NewSize(windowWidth, 60))
	app.pauseText.Move(fyne.NewPos(0, windowHeight/2))
	app.pauseText.Hide()
}

// initControls initializes touch controls
func (app *App) initControls() {
	// D-pad layout container
	app.controlsContainer = container.NewWithoutLayout()

	// Create D-pad with visual separation
	dpadWidth := float32(120)
	dpadHeight := float32(120)
	dpX := windowWidth - dpadWidth - 20
	dpadY := windowHeight - dpadHeight - 30

	// Up button
	app.upButton = widget.NewButton("▲", func() { app.onDirectionInput(game.DirUp) })
	app.upButton.Resize(fyne.NewSize(50, 50))
	app.upButton.Move(fyne.NewPos(dpX+dpadWidth/2-25, dpadY))

	// Down button
	app.downButton = widget.NewButton("▼", func() { app.onDirectionInput(game.DirDown) })
	app.downButton.Resize(fyne.NewSize(50, 50))
	app.downButton.Move(fyne.NewPos(dpX+dpadWidth/2-25, dpadY+dpadHeight-50))

	// Left button
	app.leftButton = widget.NewButton("◀", func() { app.onDirectionInput(game.DirLeft) })
	app.leftButton.Resize(fyne.NewSize(50, 50))
	app.leftButton.Move(fyne.NewPos(dpX, dpadY+dpadHeight/2-25))

	// Right button
	app.rightButton = widget.NewButton("▶", func() { app.onDirectionInput(game.DirRight) })
	app.rightButton.Resize(fyne.NewSize(50, 50))
	app.rightButton.Move(fyne.NewPos(dpX+dpadWidth-50, dpadY+dpadHeight/2-25))

	app.controlsContainer.Add(app.upButton)
	app.controlsContainer.Add(app.downButton)
	app.controlsContainer.Add(app.leftButton)
	app.controlsContainer.Add(app.rightButton)

	app.hideControls()
}

// setupInput sets up keyboard input
func (app *App) setupInput() {
	// Keyboard handlers are set up via the canvas in Run()
}

// onKeyDown handles keyboard input
func (app *App) onKeyDown(key *fyne.KeyEvent) {
	switch app.game.State {
	case game.StateTitle:
		if key.Name == fyne.KeySpace || key.Name == fyne.KeyReturn {
			app.startGame()
		}

	case game.StatePlaying:
		switch key.Name {
		case fyne.KeyUp, fyne.KeyW:
			app.directionBuffer = game.DirUp
			app.game.MovePlayer(game.DirUp)
		case fyne.KeyDown, fyne.KeyS:
			app.directionBuffer = game.DirDown
			app.game.MovePlayer(game.DirDown)
		case fyne.KeyLeft, fyne.KeyA:
			app.directionBuffer = game.DirLeft
			app.game.MovePlayer(game.DirLeft)
		case fyne.KeyRight, fyne.KeyD:
			app.directionBuffer = game.DirRight
			app.game.MovePlayer(game.DirRight)
		case fyne.KeyP, fyne.KeyEscape:
			app.game.TogglePause()
			app.updateOverlay()
		}

	case game.StatePaused:
		if key.Name == fyne.KeyP || key.Name == fyne.KeyEscape || key.Name == fyne.KeySpace {
			app.game.TogglePause()
			app.updateOverlay()
		}

	case game.StateGameOver:
		if key.Name == fyne.KeySpace || key.Name == fyne.KeyReturn {
			app.returnToTitle()
		}
	}
}

func (app *App) onKeyUp(key *fyne.KeyEvent) {
	// Clear direction buffer when key released
	if app.game.State == game.StatePlaying {
		app.directionBuffer = game.DirNone
	}
}

// onDirectionInput handles touch/mouse direction input
func (app *App) onDirectionInput(dir game.Direction) {
	if app.game.State == game.StatePlaying {
		app.directionBuffer = dir
		app.game.MovePlayer(dir)
	}
}

// startGame starts a new game
func (app *App) startGame() {
	app.game.Start()
	app.titleText.Hide()
	app.subtitleText.Hide()
	app.overlayText.Hide()
	app.pauseText.Hide()
	app.showHUD()
	app.showControls()
	app.renderMaze()
	app.updateRender()
}

// returnToTitle returns to the title screen
func (app *App) returnToTitle() {
	app.game.Initialize()
	app.hideHUD()
	app.hideControls()
	app.showTitleScreen()
}

// showTitleScreen shows the title screen
func (app *App) showTitleScreen() {
	app.titleText.Text = "RETRO MAZE"
	app.titleText.Show()
	app.subtitleText.Text = "PRESS SPACE TO START"
	app.subtitleText.Show()

	if app.game.ScoreManager.GetHighScore() > 0 {
		app.overlayText.Text = fmt.Sprintf("HIGH SCORE: %06d", app.game.ScoreManager.GetHighScore())
		app.overlayText.Show()
	}
}

// hideTitleScreen hides the title screen
func (app *App) hideTitleScreen() {
	app.titleText.Hide()
	app.subtitleText.Hide()
	app.overlayText.Hide()
}

// showHUD shows the HUD
func (app *App) showHUD() {
	app.scoreText.Show()
	app.levelText.Show()
	app.livesText.Show()
}

// hideHUD hides the HUD
func (app *App) hideHUD() {
	app.scoreText.Hide()
	app.levelText.Hide()
	app.livesText.Hide()
}

// showControls shows touch controls
func (app *App) showControls() {
	app.upButton.Show()
	app.downButton.Show()
	app.leftButton.Show()
	app.rightButton.Show()
}

// hideControls hides touch controls
func (app *App) hideControls() {
	app.upButton.Hide()
	app.downButton.Hide()
	app.leftButton.Hide()
	app.rightButton.Hide()
}

// updateOverlay updates the overlay based on game state
func (app *App) updateOverlay() {
	switch app.game.State {
	case game.StatePaused:
		app.pauseText.Show()
	case game.StateGameOver:
		app.overlayText.Text = fmt.Sprintf("GAME OVER\nSCORE: %06d\nPRESS SPACE TO CONTINUE",
			app.game.ScoreManager.GetCurrent())
		app.overlayText.Show()
	case game.StateLevelComplete:
		app.overlayText.Text = fmt.Sprintf("LEVEL %d COMPLETE!", app.currentLevel)
		app.overlayText.Show()
	default:
		app.overlayText.Hide()
		app.pauseText.Hide()
	}
}

// renderMaze renders the current maze
func (app *App) renderMaze() {
	app.mazeContainer.Objects = nil
	app.pelletContainer.Objects = nil

	blue := color.RGBA{R: 33, G: 33, B: 222, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	for y, row := range app.game.Maze.Tiles {
		for x, tile := range row {
			if tile == game.TileWall {
				wall := canvas.NewRectangle(blue)
				wall.Resize(fyne.NewSize(tileSizeF*gameScale, tileSizeF*gameScale))
				wall.Move(fyne.NewPos(float32(x)*tileSizeF*gameScale, float32(y)*tileSizeF*gameScale))
				app.mazeContainer.Objects = append(app.mazeContainer.Objects, wall)
			}
		}
	}

	// Render pellets
	for y, row := range app.game.Maze.Tiles {
		for x, tile := range row {
			if tile == game.TilePellet || tile == game.TilePowerPellet {
				pelletSize := tileSizeF * gameScale * 0.2
				if tile == game.TilePowerPellet {
					pelletSize = tileSizeF * gameScale * 0.5
				}
				pellet := canvas.NewRectangle(white)
				pellet.Resize(fyne.NewSize(pelletSize, pelletSize))
				offset := (tileSizeF*gameScale - pelletSize) / 2
				pellet.Move(fyne.NewPos(
					float32(x)*tileSizeF*gameScale+offset,
					float32(y)*tileSizeF*gameScale+offset,
				))
				app.pelletContainer.Objects = append(app.pelletContainer.Objects, pellet)
			}
		}
	}

	// Initialize enemy colors
	enemyColors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},     // Red
		color.RGBA{R: 255, G: 184, B: 255, A: 255}, // Pink
		color.RGBA{R: 0, G: 255, B: 255, A: 255},   // Cyan
		color.RGBA{R: 255, G: 184, B: 82, A: 255},  // Orange
	}

	app.enemyColors = enemyColors

	// Create enemy objects
	for i := 0; i < len(app.game.Enemies); i++ {
		enemyObj := canvas.NewCircle(enemyColors[i%len(enemyColors)])
		enemyObj.Resize(fyne.NewSize(tileSizeF*gameScale*0.7, tileSizeF*gameScale*0.7))
		app.enemyObjs = append(app.enemyObjs, enemyObj)
		app.mazeContainer.Objects = append(app.mazeContainer.Objects, enemyObj)
	}

	app.mazeContainer.Refresh()
	app.pelletContainer.Refresh()
}

// updateRender updates the game rendering
func (app *App) updateRender() {
	now := time.Now()
	dt := now.Sub(app.lastUpdateTime).Seconds()
	app.lastUpdateTime = now

	// Update game state
	app.game.Update(dt)

	// Handle state changes
	if app.game.State != game.StatePaused {
		// Move player continuously if direction is held
		if app.game.State == game.StatePlaying {
			if now.Sub(app.lastPlayerMove) > playerMoveDelay {
				// Try buffered direction first, then current direction
				if app.directionBuffer != game.DirNone {
					app.game.MovePlayer(app.directionBuffer)
				} else {
					app.game.MovePlayerContinuously()
				}
				app.lastPlayerMove = now
			}

			// Move enemies
			if now.Sub(app.lastEnemyMove) > enemyMoveDelay {
				app.game.UpdateEnemies(dt)
				app.lastEnemyMove = now
			}
		}

		// Check for level change
		newLevel := app.game.LevelManager.GetCurrentLevel()
		if newLevel != app.currentLevel {
			app.currentLevel = newLevel
			app.renderMaze()
		}

		// Check for pellet changes
		app.updatePellets()
	}

	// Update player position
	if app.game.Player.IsAlive {
		offset := (tileSizeF * gameScale * 0.2) / 2
		app.playerObj.Move(fyne.NewPos(
			float32(app.game.Player.X)*tileSizeF*gameScale+offset,
			float32(app.game.Player.Y)*tileSizeF*gameScale+offset,
		))
		app.playerObj.Show()
	} else {
		app.playerObj.Hide()
	}

	// Update enemy positions
	for i, enemyObj := range app.enemyObjs {
		if i < len(app.game.Enemies) {
			enemy := app.game.Enemies[i]
			if enemy.IsAlive {
				enemySize := tileSizeF * gameScale * 0.7
				offset := (tileSizeF*gameScale - enemySize) / 2
				enemyObj.Move(fyne.NewPos(
					float32(enemy.X)*tileSizeF*gameScale+offset,
					float32(enemy.Y)*tileSizeF*gameScale+offset,
				))

				// Update color based on vulnerable state
				if enemy.IsVulnerable {
					enemyObj.FillColor = color.RGBA{R: 33, G: 33, B: 222, A: 255}
				} else {
					enemyObj.FillColor = app.enemyColors[i%len(app.enemyColors)]
				}
				enemyObj.Show()
			} else {
				enemyObj.Hide()
			}
		}
	}

	// Update HUD
	app.scoreText.Text = fmt.Sprintf("SCORE: %06d", app.game.ScoreManager.GetCurrent())
	app.levelText.Text = fmt.Sprintf("LEVEL: %02d", app.game.LevelManager.GetCurrentLevel())

	livesText := "LIVES: "
	for i := 0; i < app.game.Player.Lives; i++ {
		livesText += "♥"
	}
	app.livesText.Text = livesText

	// Handle state-specific UI
	switch app.game.State {
	case game.StateGameOver:
		app.updateOverlay()
	case game.StateLevelComplete:
		app.updateOverlay()
	}

	app.gameContainer.Refresh()
}

// updatePellets updates the pellet rendering
func (app *App) updatePellets() {
	// For simplicity, we'll re-render pellets only when needed
	// This is optimized enough for the small game resolution
}

// Run starts the game loop
func (app *App) Run() {
	// Create the main content
	content := container.NewWithoutLayout(
		app.gameContainer,
		app.scoreText,
		app.levelText,
		app.livesText,
		app.titleText,
		app.subtitleText,
		app.overlayText,
		app.pauseText,
		app.upButton,
		app.downButton,
		app.leftButton,
		app.rightButton,
	)

	app.window.SetContent(content)

	// Set up keyboard handling using desktop canvas interface
	if desktopCanvas, ok := app.window.Canvas().(desktop.Canvas); ok {
		desktopCanvas.SetOnKeyDown(app.onKeyDown)
		desktopCanvas.SetOnKeyUp(app.onKeyUp)
	}

	// Show initial title screen
	app.hideHUD()
	app.showTitleScreen()

	// Start the game loop
	go app.gameLoop()

	app.window.ShowAndRun()
}

// gameLoop runs the game loop
func (app *App) gameLoop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	for range ticker.C {
		if app.game.State == game.StatePlaying || app.game.State == game.StateLevelComplete {
			app.updateRender()
		} else if app.game.State == game.StateReady {
			app.updateRender()
		}

		// Update overlay for state changes
		if app.game.State == game.StateGameOver {
			app.updateOverlay()
		}
	}
}