# Build a Retro Samsung SGH-C108-Style Pac-Man Game in Go + Fyne

## Objective

Build a complete, playable **retro maze game inspired by the Pac-Man-style game found on early Samsung feature phones such as the Samsung SGH-C108**.

The game must be implemented in **Go (Golang)** using **Fyne** as the cross-platform GUI framework.

The goal is NOT to reproduce Samsung's original proprietary game/assets. Create an original game that captures the same **early feature-phone / monochrome retro gaming experience**.

The application must run on:

- Linux
- Windows
- macOS

Design the architecture so that Android/iOS support could potentially be added later.

---

# 1. Technology Requirements

Use:

- Go
- Fyne
- Standard Go packages where appropriate
- Fyne Canvas APIs for rendering
- No web technologies
- No Electron
- No HTML/CSS/JavaScript
- No external game engine unless absolutely necessary

Use Go modules.

Example:

```bash
go mod init retro-maze
```

The application should be buildable with:

```bash
go run .
```

and:

```bash
go build .
```

---

# 2. Game Concept

Create a small retro maze game with the following gameplay:

- Player controls a yellow/bright circular character.
- Player moves through a maze.
- Collect dots/pellets.
- Avoid enemies.
- Eating all pellets completes the level.
- Losing all lives causes Game Over.
- Player can restart the game.
- Increasing levels should gradually increase difficulty.

The gameplay should feel like a **very old feature-phone arcade game**, rather than a modern high-resolution Pac-Man clone.

Do not copy proprietary Samsung assets, sounds, graphics, or source code.

---

# 3. Retro Visual Style

The visual design is extremely important.

Target the aesthetic of:

- Samsung SGH-C108
- early 2000s feature phones
- monochrome/limited-color LCD games
- tiny low-resolution screens
- simple pixel graphics
- simple geometric sprites
- minimal UI

The game should intentionally look retro.

Avoid:

- modern gradients
- realistic lighting
- 3D graphics
- particle effects
- complex animations
- high-resolution textures
- excessive UI
- modern arcade effects

Use a small logical game resolution such as:

```text
160 × 128
```

or another appropriately small virtual resolution.

Render the game using integer scaling where possible so that pixels remain crisp.

For example:

```text
Logical resolution:
160 × 128

Window:
640 × 512

Scale:
4x
```

The game should still work correctly when the window is resized.

---

# 4. Game Screen

Use a Fyne window containing:

```text
+--------------------------------+
| SCORE: 000120     LEVEL: 01    |
+--------------------------------+
|                                |
|   ########################     |
|   # . . . . # . . . . . #     |
|   # .###.#.# # #.###.#. #     |
|   # .#   #.       .#   #     |
|   # .# ### ##### ### #. #     |
|   # . . .   ...   . . . #     |
|   # ### ### ### ### ### #     |
|   #       #     #       #     |
|   ########################     |
|                                |
+--------------------------------+
|       LIVES: ♥ ♥ ♥             |
+--------------------------------+
```

The actual maze should be generated/rendered programmatically.

---

# 5. Core Game Loop

Implement a deterministic game loop.

The game loop should conceptually perform:

```text
Input
  ↓
Update
  ↓
Collision
  ↓
Enemy AI
  ↓
Animation
  ↓
Render
  ↓
Repeat
```

Separate game logic from Fyne UI/rendering as much as possible.

Do not put the entire game implementation inside `main.go`.

---

# 6. Recommended Architecture

Use a clean modular architecture.

Suggested structure:

```text
retro-maze/
│
├── go.mod
├── go.sum
├── main.go
│
├── game/
│   ├── game.go
│   ├── state.go
│   ├── loop.go
│   ├── level.go
│   ├── maze.go
│   ├── collision.go
│   └── score.go
│
├── player/
│   ├── player.go
│   └── movement.go
│
├── enemy/
│   ├── enemy.go
│   └── ai.go
│
├── input/
│   └── input.go
│
├── renderer/
│   ├── renderer.go
│   ├── maze.go
│   ├── player.go
│   └── enemy.go
│
├── ui/
│   ├── menu.go
│   ├── hud.go
│   └── game_over.go
│
├── assets/
│   └── README.md
│
└── README.md
```

Keep responsibilities clearly separated.

---

# 7. Game States

Implement:

```go
type GameState int

const (
    StateTitle GameState = iota
    StateReady
    StatePlaying
    StatePaused
    StateLevelComplete
    StateGameOver
)
```

State transitions should be explicit.

Example:

```text
TITLE
  ↓
READY
  ↓
PLAYING
  ↓
LEVEL COMPLETE
  ↓
NEXT LEVEL
  ↓
PLAYING

or

PLAYING
  ↓
GAME OVER
  ↓
TITLE
```

---

# 8. Maze Representation

Do NOT hard-code every wall as an individual graphical object.

Represent the maze using a grid.

For example:

```go
type Tile int

const (
    TileWall Tile = iota
    TileEmpty
    TilePellet
    TilePlayerStart
    TileEnemyStart
)
```

Example:

```text
################
#..............#
#.####.#######.#
#..............#
#.####.#######.#
#......#.......#
################
```

Convert the maze into a grid:

```go
[][]Tile
```

The maze renderer should translate grid cells into visual tiles.

---

# 9. Grid-Based Movement

Movement should primarily be grid-based, similar to old feature-phone games.

The player should move through:

```text
UP
DOWN
LEFT
RIGHT
```

Prevent the player from moving through walls.

Movement should feel responsive.

Support:

```text
Arrow keys
WASD
```

Additionally, provide on-screen Fyne controls:

```text
        ↑

    ←       →

        ↓
```

The controls should work with mouse/touch where supported.

---

# 10. Player

Create:

```go
type Player struct {
    X          int
    Y          int
    Direction  Direction
    NextDirection Direction
    Lives      int
    Score      int
}
```

The player should have:

- current position
- current direction
- requested direction
- lives
- score
- movement speed
- animation state

Implement simple mouth/open-close animation.

Do not use a large external sprite.

Prefer drawing the character using Fyne canvas primitives where practical.

---

# 11. Pellets

Each pellet should:

- occupy a maze cell
- disappear when collected
- increase score

Example:

```text
Normal pellet = 10 points
Power pellet  = 50 points
Enemy eaten   = bonus points
```

The game should display:

```text
SCORE: 001250
```

with zero padding.

---

# 12. Enemies

Create several simple enemies.

Example:

```go
type Enemy struct {
    X         int
    Y         int
    Direction Direction
    State     EnemyState
}
```

Use simple AI.

Possible behaviors:

```text
Enemy 1:
Random movement

Enemy 2:
Try to approach player

Enemy 3:
Prefer horizontal movement

Enemy 4:
Prefer vertical movement
```

Do NOT implement an overly complex AI system initially.

The objective is to reproduce the simple behavior of old feature-phone games.

---

# 13. Enemy Collision

If:

```text
Player position == Enemy position
```

then:

```text
Player loses a life
```

Reset:

```text
Player position
Enemy positions
Player direction
```

If lives reach zero:

```text
StateGameOver
```

---

# 14. Power-Up

Implement a simple temporary power mode.

When the player collects a power pellet:

```text
PowerMode = true
PowerModeDuration = 8 seconds
```

Enemies become vulnerable.

The player can temporarily eat enemies.

After the timer expires:

```text
PowerMode = false
```

Keep the effect visually simple.

---

# 15. Level System

Start with:

```text
Level 1
```

After all pellets are collected:

```text
Level 2
```

Increase difficulty gradually.

For example:

```text
Level 1:
Enemy speed = 1.0

Level 2:
Enemy speed = 1.1

Level 3:
Enemy speed = 1.2
```

Increase difficulty without making the game frustrating.

---

# 16. Fyne Rendering

Use Fyne's canvas APIs where appropriate.

Possible primitives:

```go
canvas.Rectangle
canvas.Circle
canvas.Line
canvas.Text
```

The renderer should maintain a collection of visual objects and update their positions rather than unnecessarily recreating the entire UI every frame.

Avoid creating thousands of Fyne objects during gameplay.

Optimize rendering for the small logical resolution.

---

# 17. Rendering Abstraction

Create an abstraction similar to:

```go
type Renderer interface {
    DrawMaze(...)
    DrawPlayer(...)
    DrawEnemies(...)
    DrawPellets(...)
    DrawHUD(...)
    Refresh()
}
```

The game logic should not depend heavily on Fyne-specific implementation details.

This will make the game easier to test and potentially port later.

---

# 18. Input Abstraction

Create:

```go
type InputManager interface {
    IsPressed(direction Direction) bool
}
```

Support:

### Keyboard

```text
Arrow keys
W
A
S
D
```

### Touch / Mouse

Fyne buttons:

```text
UP
LEFT
RIGHT
DOWN
```

The same game logic must receive input regardless of input source.

---

# 19. Audio

Keep audio optional.

If implementing audio, support:

- pellet collection
- power-up
- enemy collision
- level complete
- game over

Use simple retro-style sounds.

Do not use copyrighted Samsung/Pac-Man audio.

If audio support complicates the initial implementation, create an abstraction:

```go
type Audio interface {
    PlayPellet()
    PlayPowerUp()
    PlayDeath()
    PlayLevelComplete()
}
```

and provide a no-op implementation initially.

---

# 20. Menus

Create a simple retro title screen:

```text
+------------------------+
|                        |
|      RETRO MAZE        |
|                        |
|    PRESS START         |
|                        |
|      HIGH SCORE        |
|        001250          |
|                        |
+------------------------+
```

Buttons:

```text
START
HOW TO PLAY
EXIT
```

Keep the UI visually consistent with the retro theme.

---

# 21. High Score

Implement a local high-score system.

At minimum:

```text
Current Score
High Score
```

Persist the high score locally.

Use a simple JSON file or another standard Go mechanism.

Example:

```json
{
    "high_score": 12500
}
```

Do not require a database.

---

# 22. Game Timing

Do not make movement dependent directly on rendering speed.

Use elapsed time / fixed timestep logic.

Conceptually:

```go
deltaTime := time.Since(lastUpdate)
```

Use timing so that the game behaves consistently on:

- Linux
- Windows
- macOS

Avoid CPU-intensive busy loops.

---

# 23. Window Behavior

Create a Fyne application window.

Initial size approximately:

```text
640 × 512
```

Maintain a sensible aspect ratio.

Support resizing.

The game should remain centered.

Do not stretch pixels unevenly.

---

# 24. Code Quality

Follow idiomatic Go.

Use:

- small functions
- clear structs
- interfaces where useful
- meaningful names
- constants instead of magic numbers
- error handling
- Go documentation for exported types
- `go fmt`
- `go vet`

Avoid:

- giant functions
- global mutable state
- unnecessary interfaces
- unnecessary abstractions
- duplicated game logic
- Fyne code mixed throughout every package

---

# 25. Testing

Create unit tests for:

### Maze

```text
IsWall()
IsWalkable()
CollectPellet()
```

### Collision

```text
PlayerEnemyCollision()
PlayerWallCollision()
```

### Movement

```text
MoveUp()
MoveDown()
MoveLeft()
MoveRight()
```

### Score

```text
AddScore()
HighScore()
```

### Level

```text
LevelComplete()
NextLevel()
```

Use Go's standard:

```bash
go test ./...
```

---

# 26. Development Phases

Implement incrementally.

## Phase 1

Create:

- Go project
- Fyne window
- game canvas
- basic game loop

Verify:

```bash
go run .
```

---

## Phase 2

Implement:

- maze grid
- wall rendering
- pellets

---

## Phase 3

Implement:

- player
- keyboard input
- movement
- collision

---

## Phase 4

Implement:

- pellet collection
- scoring
- HUD

---

## Phase 5

Implement:

- enemies
- enemy movement
- collision

---

## Phase 6

Implement:

- lives
- game over
- restart

---

## Phase 7

Implement:

- power pellets
- vulnerable enemies

---

## Phase 8

Implement:

- levels
- increasing difficulty

---

## Phase 9

Implement:

- title screen
- pause
- game over screen
- high score

---

## Phase 10

Implement:

- touch/on-screen controls
- audio abstraction
- polish
- packaging

---

# 27. Important Engineering Rule

Do NOT try to build everything in one huge implementation.

After every phase:

1. Run the application.
2. Test the feature.
3. Fix compilation errors.
4. Run:

```bash
go fmt ./...
go vet ./...
go test ./...
```

5. Only then proceed to the next phase.

---

# 28. Final Deliverables

The completed project must contain:

```text
✓ Complete Go source code
✓ Fyne GUI
✓ Playable maze game
✓ Retro visual style
✓ Keyboard controls
✓ Mouse/touch controls
✓ Player movement
✓ Collision detection
✓ Pellets
✓ Score
✓ Enemies
✓ Enemy AI
✓ Power pellets
✓ Lives
✓ Levels
✓ Game Over
✓ Pause
✓ High score persistence
✓ Cross-platform build support
✓ Unit tests
✓ README
```

README should explain:

```text
Requirements
Installation
Running
Controls
Building on Linux
Building on Windows
Building on macOS
Project architecture
Game architecture
How to add levels
How to modify enemies
How to modify the maze
```

---

# 29. Final Design Principle

The finished game should feel like:

> "A game you could realistically imagine playing on a tiny Samsung feature phone in the early 2000s."

It should be:

```text
Small
Fast
Simple
Pixelated
Responsive
Minimal
Addictive
Retro
```

Do not turn it into a modern Pac-Man clone.

Prioritize the **feel and constraints of a feature-phone game** over visual complexity.

Start by creating the project and implementing Phase 1. After completing each phase, verify that the project builds and runs before continuing.