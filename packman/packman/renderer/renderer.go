package renderer

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// GameRenderer is the main renderer interface
type GameRenderer interface {
	DrawMaze(maze [][]int, pellets [][]bool, powerPellets [][]bool, scale float32)
	DrawPlayer(x, y int, direction int, animFrame int, isAlive bool, scale float32)
	DrawEnemies(enemies []EnemyRenderData, scale float32)
	DrawHUD(score, highScore, level, lives int)
	DrawTitle()
	DrawGameOver(score, highScore int)
	DrawLevelComplete(level int)
	DrawReady()
	DrawPaused()
	GetCanvas() fyne.CanvasObject
	Refresh()
}

// EnemyRenderData holds data needed to render an enemy
type EnemyRenderData struct {
	X            int
	Y            int
	Direction    int
	IsVulnerable bool
	AnimFrame    int
	Alive        bool
	Color        color.Color
}

// FyneRenderer implements GameRenderer using Fyne canvas primitives
type FyneRenderer struct {
	container     *fyne.Container
	mazeContainer *fyne.Container
	playerObj     *canvas.Circle
	enemyObjs     []*canvas.Circle
	pelletObjs    []*canvas.Rectangle
	hudScore      *canvas.Text
	hudLevel      *canvas.Text
	hudLives      *canvas.Text
	titleText     *canvas.Text
	subtitleText  *canvas.Text
	infoText      *canvas.Text

	// Game area
	gameWidth  float32
	gameHeight float32
	scale      float32
	offsetX    float32
	offsetY    float32
}

// Colors for retro style
var (
	colorBlack     = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	colorBlue      = color.RGBA{R: 0, G: 0, B: 139, A: 255}
	colorYellow    = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	colorRed       = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	colorPink      = color.RGBA{R: 255, G: 184, B: 255, A: 255}
	colorCyan      = color.RGBA{R: 0, G: 255, B: 255, A: 255}
	colorOrange    = color.RGBA{R: 255, G: 184, B: 82, A: 255}
	colorWhite     = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	colorGreen     = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	colorDarkBlue  = color.RGBA{R: 33, G: 33, B: 222, A: 255}
)

// NewFyneRenderer creates a new Fyne renderer
func NewFyneRenderer(gameWidth, gameHeight float32) *FyneRenderer {
	r := &FyneRenderer{
		gameWidth:  gameWidth,
		gameHeight: gameHeight,
		scale:      1.0,
	}

	// Create background
	bg := canvas.NewRectangle(colorBlack)
	bg.Resize(fyne.NewSize(gameWidth, gameHeight))

	// Create maze container
	r.mazeContainer = container.NewWithoutLayout()

	// Create player object
	r.playerObj = canvas.NewCircle(colorYellow)
	r.playerObj.Hide()

	// Create HUD
	r.hudScore = canvas.NewText("SCORE: 000000", colorWhite)
	r.hudScore.TextSize = 12
	r.hudScore.TextStyle = fyne.TextStyle{Monospace: true}

	r.hudLevel = canvas.NewText("LEVEL: 01", colorWhite)
	r.hudLevel.TextSize = 12
	r.hudLevel.TextStyle = fyne.TextStyle{Monospace: true}

	r.hudLives = canvas.NewText("LIVES: ♥♥♥", colorWhite)
	r.hudLives.TextSize = 12
	r.hudLives.TextStyle = fyne.TextStyle{Monospace: true}

	// Title screen text
	r.titleText = canvas.NewText("RETRO MAZE", colorYellow)
	r.titleText.TextSize = 24
	r.titleText.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	r.titleText.Alignment = fyne.TextAlignCenter
	r.titleText.Hide()

	r.subtitleText = canvas.NewText("PRESS START", colorWhite)
	r.subtitleText.TextSize = 14
	r.subtitleText.TextStyle = fyne.TextStyle{Monospace: true}
	r.subtitleText.Alignment = fyne.TextAlignCenter
	r.subtitleText.Hide()

	r.infoText = canvas.NewText("", colorWhite)
	r.infoText.TextSize = 10
	r.infoText.TextStyle = fyne.TextStyle{Monospace: true}
	r.infoText.Alignment = fyne.TextAlignCenter
	r.infoText.Hide()

	// Main container
	r.container = container.NewWithoutLayout(
		bg,
		r.mazeContainer,
		r.playerObj,
	)

	return r
}

// GetCanvas returns the main canvas object
func (r *FyneRenderer) GetCanvas() fyne.CanvasObject {
	return r.container
}

// Refresh refreshes all rendered objects
func (r *FyneRenderer) Refresh() {
	r.container.Refresh()
}

// SetScale sets the rendering scale
func (r *FyneRenderer) SetScale(scale float32) {
	r.scale = scale
}

// SetOffset sets the game area offset
func (r *FyneRenderer) SetOffset(x, y float32) {
	r.offsetX = x
	r.offsetY = y
}

// ClearMaze clears all maze objects
func (r *FyneRenderer) ClearMaze() {
	r.mazeContainer.Objects = nil
	r.mazeContainer.Refresh()
}

// ClearEnemies clears all enemy objects
func (r *FyneRenderer) ClearEnemies() {
	for _, obj := range r.enemyObjs {
		obj.Hide()
	}
	r.enemyObjs = nil
}

// ShowTitleScreen shows the title screen
func (r *FyneRenderer) ShowTitleScreen(highScore int) {
	r.ClearMaze()
	r.HidePlayer()
	r.ClearEnemies()
	r.HideHUD()

	r.titleText.Show()
	r.subtitleText.Show()
	r.subtitleText.Text = "PRESS START"

	highScoreText := ""
	if highScore > 0 {
		highScoreText = "HIGH SCORE: " + formatScore(highScore)
	}
	r.infoText.Text = highScoreText
	r.infoText.Show()

	// Position title elements
	titleY := r.gameHeight / 3

	r.titleText.Resize(fyne.NewSize(r.gameWidth, 30))
	r.titleText.Move(fyne.NewPos(0, titleY))

	r.subtitleText.Resize(fyne.NewSize(r.gameWidth, 20))
	r.subtitleText.Move(fyne.NewPos(0, titleY+40))

	r.infoText.Resize(fyne.NewSize(r.gameWidth, 20))
	r.infoText.Move(fyne.NewPos(0, titleY+70))

	r.container.Add(r.titleText)
	r.container.Add(r.subtitleText)
	r.container.Add(r.infoText)
	r.Refresh()
}

// HideTitleScreen hides the title screen
func (r *FyneRenderer) HideTitleScreen() {
	r.titleText.Hide()
	r.subtitleText.Hide()
	r.infoText.Hide()
	r.Refresh()
}

// ShowGameOver shows the game over screen
func (r *FyneRenderer) ShowGameOver(score, highScore int) {
	r.HidePlayer()
	r.ClearEnemies()

	r.titleText.Text = "GAME OVER"
	r.titleText.Show()

	r.subtitleText.Text = "SCORE: " + formatScore(score)
	r.subtitleText.Show()

	highScoreText := "HIGH SCORE: " + formatScore(highScore)
	r.infoText.Text = highScoreText + "\nPRESS START"
	r.infoText.Show()

	centerY := r.gameHeight / 3

	r.titleText.Resize(fyne.NewSize(r.gameWidth, 30))
	r.titleText.Move(fyne.NewPos(0, centerY))

	r.subtitleText.Resize(fyne.NewSize(r.gameWidth, 20))
	r.subtitleText.Move(fyne.NewPos(0, centerY+40))

	r.infoText.Resize(fyne.NewSize(r.gameWidth, 40))
	r.infoText.Move(fyne.NewPos(0, centerY+70))

	if !r.containsObject(r.titleText) {
		r.container.Add(r.titleText)
		r.container.Add(r.subtitleText)
		r.container.Add(r.infoText)
	}
	r.Refresh()
}

// ShowLevelComplete shows level complete screen
func (r *FyneRenderer) ShowLevelComplete(level int) {
	r.HidePlayer()

	r.titleText.Text = "LEVEL COMPLETE!"
	r.titleText.Show()

	r.subtitleText.Text = "LEVEL " + formatLevel(level)
	r.subtitleText.Show()

	r.infoText.Text = "GET READY..."
	r.infoText.Show()

	centerY := r.gameHeight / 3

	r.titleText.Resize(fyne.NewSize(r.gameWidth, 30))
	r.titleText.Move(fyne.NewPos(0, centerY))

	r.subtitleText.Resize(fyne.NewSize(r.gameWidth, 20))
	r.subtitleText.Move(fyne.NewPos(0, centerY+40))

	r.infoText.Resize(fyne.NewSize(r.gameWidth, 20))
	r.infoText.Move(fyne.NewPos(0, centerY+70))

	if !r.containsObject(r.titleText) {
		r.container.Add(r.titleText)
		r.container.Add(r.subtitleText)
		r.container.Add(r.infoText)
	}
	r.Refresh()
}

// ShowReady shows ready screen
func (r *FyneRenderer) ShowReady() {
	r.titleText.Text = "READY!"
	r.titleText.Show()
	r.subtitleText.Hide()
	r.infoText.Hide()

	centerY := r.gameHeight / 2
	r.titleText.Resize(fyne.NewSize(r.gameWidth, 30))
	r.titleText.Move(fyne.NewPos(0, centerY))

	if !r.containsObject(r.titleText) {
		r.container.Add(r.titleText)
	}
	r.Refresh()
}

// HideReady hides ready screen
func (r *FyneRenderer) HideReady() {
	r.titleText.Hide()
	r.subtitleText.Hide()
	r.infoText.Hide()
}

// ShowPaused shows pause overlay
func (r *FyneRenderer) ShowPaused() {
	r.titleText.Text = "PAUSED"
	r.titleText.Show()
	r.subtitleText.Text = "PRESS P TO RESUME"
	r.subtitleText.Show()
	r.infoText.Hide()

	centerY := r.gameHeight / 2
	r.titleText.Resize(fyne.NewSize(r.gameWidth, 30))
	r.titleText.Move(fyne.NewPos(0, centerY))

	r.subtitleText.Resize(fyne.NewSize(r.gameWidth, 20))
	r.subtitleText.Move(fyne.NewPos(0, centerY+40))

	if !r.containsObject(r.titleText) {
		r.container.Add(r.titleText)
		r.container.Add(r.subtitleText)
	}
	r.Refresh()
}

// HidePaused hides pause overlay
func (r *FyneRenderer) HidePaused() {
	r.titleText.Hide()
	r.subtitleText.Hide()
	r.infoText.Hide()
}

// DrawMaze draws the maze walls and pellets
func (r *FyneRenderer) DrawMaze(maze [][]int, pellets [][]bool, powerPellets [][]bool, scale float32, tileSize float32) {
	r.ClearMaze()

	// Draw walls
	for y, row := range maze {
		for x, tile := range row {
			if tile == 1 { // Wall
				wall := canvas.NewRectangle(colorBlue)
				wall.Resize(fyne.NewSize(tileSize*scale-1, tileSize*scale-1))
				wall.Move(fyne.NewPos(r.offsetX+float32(x)*tileSize*scale, r.offsetY+float32(y)*tileSize*scale))
				r.mazeContainer.Objects = append(r.mazeContainer.Objects, wall)
			}
		}
	}

	r.Refresh()
}

// DrawPellets draws the pellets
func (r *FyneRenderer) DrawPellets(pellets [][]bool, powerPellets [][]bool, scale float32, tileSize float32) {
	// Remove old pellets
	r.pelletObjs = nil

	for y, row := range pellets {
		for x, hasPellet := range row {
			if hasPellet {
				pellet := canvas.NewRectangle(colorWhite)
				isPower := false
				if y < len(powerPellets) && x < len(powerPellets[y]) {
					isPower = powerPellets[y][x]
				}

				pelletSize := tileSize * scale * 0.2
				if isPower {
					pelletSize = tileSize * scale * 0.5
				}

				pellet.Resize(fyne.NewSize(pelletSize, pelletSize))
				offset := (tileSize*scale - pelletSize) / 2
				pellet.Move(fyne.NewPos(
					r.offsetX+float32(x)*tileSize*scale+offset,
					r.offsetY+float32(y)*tileSize*scale+offset,
				))

				r.pelletObjs = append(r.pelletObjs, pellet)
				r.mazeContainer.Objects = append(r.mazeContainer.Objects, pellet)
			}
		}
	}

	r.mazeContainer.Refresh()
}

// ShowPlayer shows the player
func (r *FyneRenderer) ShowPlayer(scale float32, tileSize float32) {
	playerSize := tileSize * scale * 0.8
	r.playerObj.Resize(fyne.NewSize(playerSize, playerSize))
	r.playerObj.Show()
	r.Refresh()
}

// HidePlayer hides the player
func (r *FyneRenderer) HidePlayer() {
	r.playerObj.Hide()
}

// UpdatePlayerPosition updates player position
func (r *FyneRenderer) UpdatePlayerPosition(x, y int, scale float32, tileSize float32) {
	offset := (tileSize * scale * 0.2) / 2
	r.playerObj.Move(fyne.NewPos(
		r.offsetX+float32(x)*tileSize*scale+offset,
		r.offsetY+float32(y)*tileSize*scale+offset,
	))
	r.playerObj.Refresh()
}

// DrawEnemies draws all enemies
func (r *FyneRenderer) DrawEnemies(enemies []EnemyRenderData, scale float32, tileSize float32) {
	// Hide old enemies
	for _, obj := range r.enemyObjs {
		obj.Hide()
	}
	r.enemyObjs = nil

	for _, enemy := range enemies {
		if !enemy.Alive {
			continue
		}

		enemySize := tileSize * scale * 0.7
		enemyObj := canvas.NewCircle(enemy.Color)
		enemyObj.Resize(fyne.NewSize(enemySize, enemySize))
		enemyObj.Move(fyne.NewPos(
			r.offsetX+float32(enemy.X)*tileSize*scale+(tileSize*scale-enemySize)/2,
			r.offsetY+float32(enemy.Y)*tileSize*scale+(tileSize*scale-enemySize)/2,
		))

		r.enemyObjs = append(r.enemyObjs, enemyObj)
		r.mazeContainer.Objects = append(r.mazeContainer.Objects, enemyObj)
	}

	r.mazeContainer.Refresh()
}

// UpdateHUD updates the HUD display
func (r *FyneRenderer) UpdateHUD(score, highScore, level, lives int) {
	r.hudScore.Text = "SCORE: " + formatScore(score)
	r.hudLevel.Text = "LEVEL: " + formatLevel(level)

	livesText := "LIVES: "
	for i := 0; i < lives; i++ {
		livesText += "♥"
	}
	r.hudLives.Text = livesText

	r.hudScore.Refresh()
	r.hudLevel.Refresh()
	r.hudLives.Refresh()
}

// ShowHUD shows the HUD
func (r *FyneRenderer) ShowHUD() {
	if !r.containsObject(r.hudScore) {
		r.container.Add(r.hudScore)
		r.container.Add(r.hudLevel)
		r.container.Add(r.hudLives)
	}
	r.hudScore.Show()
	r.hudLevel.Show()
	r.hudLives.Show()
}

// HideHUD hides the HUD
func (r *FyneRenderer) HideHUD() {
	r.hudScore.Hide()
	r.hudLevel.Hide()
	r.hudLives.Hide()
}

// PositionHUD positions the HUD elements
func (r *FyneRenderer) PositionHUD(offsetX, offsetY, gameWidth float32) {
	r.hudScore.Move(fyne.NewPos(offsetX, offsetY))
	r.hudScore.Resize(fyne.NewSize(gameWidth/2, 20))

	r.hudLevel.Move(fyne.NewPos(offsetX+gameWidth/2, offsetY))
	r.hudLevel.Resize(fyne.NewSize(gameWidth/2, 20))

	r.hudLives.Move(fyne.NewPos(offsetX, offsetY+r.gameHeight-20))
	r.hudLives.Resize(fyne.NewSize(gameWidth, 20))
}

// containsObject checks if container contains an object
func (r *FyneRenderer) containsObject(obj fyne.CanvasObject) bool {
	for _, o := range r.container.Objects {
		if o == obj {
			return true
		}
	}
	return false
}

// formatScore formats a score with leading zeros
func formatScore(score int) string {
	if score < 0 {
		score = 0
	}
	if score > 999999 {
		score = 999999
	}
	result := "000000"
	s := fmt.Sprintf("%d", score)
	return result[:len(result)-len(s)] + s
}

// formatLevel formats a level number
func formatLevel(level int) string {
	if level < 1 {
		level = 1
	}
	if level > 99 {
		level = 99
	}
	return fmt.Sprintf("%02d", level)
}
