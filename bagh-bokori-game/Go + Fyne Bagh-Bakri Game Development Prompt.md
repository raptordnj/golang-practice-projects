# Build a Cross-Platform Bagh-Bakri (Tiger-Goat) Game in Go

You are an expert Go game developer and software architect.

Build a complete, polished **Bagh-Bakri / Tiger-Goat** strategy board game in **Go**, using **Fyne** for the desktop GUI.

The architecture must keep the **game engine completely independent from Fyne**, so that the same engine can later be reused for WebAssembly/WebGL or another frontend.

## 1. Game Concept

Implement the traditional Tiger-Goat board game:

- 4 Tigers
- 20 Goats
- Tigers can move along valid board connections.
- Tigers can capture goats by jumping over them to an empty connected position.
- Goats cannot capture Tigers.
- Goats are placed progressively onto the board.
- Tigers win when they capture enough goats / goats can no longer prevent the tigers, according to the selected ruleset.
- Goats win when all Tigers are blocked and no legal Tiger movement/capture is possible.

Use a clearly defined, deterministic board topology.

If multiple regional rules exist, implement the commonly used 4-Tiger / 20-Goat variant and document the exact rules in `README.md`.

Do NOT hard-code game logic into the UI.

---

# 2. Primary Technology

Use:

- Go
- Fyne for desktop GUI
- Go modules
- Standard Go testing
- `go vet`
- `gofmt`

Target:

- Linux
- Windows
- macOS

Keep the architecture compatible with:

- WebAssembly
- future WebGL/browser frontend
- future mobile frontend

Do not introduce a browser-specific dependency into the core game engine.

---

# 3. Architecture

Use a clean layered architecture:

```text
bagh-bakri/
├── cmd/
│   └── bagh-bakri/
│       └── main.go
│
├── internal/
│   ├── game/
│   │   ├── board.go
│   │   ├── position.go
│   │   ├── piece.go
│   │   ├── move.go
│   │   ├── rules.go
│   │   ├── state.go
│   │   ├── engine.go
│   │   └── game_test.go
│   │
│   ├── ai/
│   │   ├── minimax.go
│   │   ├── evaluation.go
│   │   └── ai_test.go
│   │
│   └── ui/
│       ├── board.go
│       ├── renderer.go
│       ├── input.go
│       └── theme.go
│
├── assets/
│   ├── icons/
│   └── sounds/
│
├── README.md
├── go.mod
└── LICENSE
```

You may improve this structure if there is a strong architectural reason.

Important:

`internal/game` must NOT import Fyne.

The dependency direction should be:

```text
Fyne UI
   ↓
Game Engine
   ↓
Board / Rules / State
```

Never:

```text
Game Engine → Fyne
```

---

# 4. Game Engine

Create a deterministic game engine.

Represent the board as a graph.

For example:

```go
type PositionID int

type Position struct {
    ID       PositionID
    X        float64
    Y        float64
    Neighbors []PositionID
}
```

Represent pieces explicitly:

```go
type PieceType int

const (
    Empty PieceType = iota
    Tiger
    Goat
)
```

Represent game state:

```go
type GameState struct {
    Board       Board
    Tigers      []PositionID
    Goats       []PositionID
    GoatsPlaced int
    Turn        Player
    Status      GameStatus
}
```

Avoid storing redundant state whenever practical.

The engine must provide APIs similar to:

```go
NewGame() GameState

LegalMoves(state GameState) []Move

ApplyMove(state GameState, move Move) (GameState, error)

IsValidMove(state GameState, move Move) bool

Winner(state GameState) Winner

IsGameOver(state GameState) bool
```

Use immutable-style state transitions where practical so AI simulations do not mutate the real game accidentally.

---

# 5. Board Rules

Define the complete board topology explicitly.

Do NOT calculate connectivity from screen coordinates.

The graphical coordinates are presentation data only.

The board model should know:

```text
position → neighboring positions
```

and, where necessary:

```text
tiger position
+
goat position
+
destination
=
valid capture
```

Implement:

### Tiger movement

A Tiger may:

1. Move to an adjacent empty connected position.

OR

2. Jump over one Goat to a valid empty destination when the board topology permits the jump.

A capture must remove exactly one Goat.

### Goat movement

During the placement phase:

- The player places a Goat on an empty valid position.

After all 20 Goats have been placed:

- Goats may move according to the selected ruleset.

Make this behavior configurable in the rules implementation rather than scattered throughout UI code.

---

# 6. Rules Engine

Create a `Rules` abstraction.

Example:

```go
type Rules interface {
    LegalMoves(state GameState) []Move
    IsValidMove(state GameState, move Move) bool
    ApplyMove(state GameState, move Move) (GameState, error)
    Winner(state GameState) Winner
}
```

Implement:

```text
TraditionalBaghBakriRules
```

Document every rule.

Do not silently invent rules.

---

# 7. Fyne UI

Create a polished desktop game.

The board should be visually obvious:

- traditional board appearance
- clear lines/connections
- Tiger pieces visually distinct
- Goat pieces visually distinct
- selected piece highlight
- legal move indicators
- capture indication
- turn indicator
- Goat counter
- captured Goat counter
- game status
- restart button
- new game button
- player-vs-player mode
- player-vs-AI mode

The UI should be responsive and scale correctly when the window is resized.

Do not use fixed pixel coordinates everywhere.

Create a board renderer that calculates positions based on available canvas size.

---

# 8. Interaction

The user should be able to:

1. Click a Tiger.
2. See its legal destinations.
3. Click destination.
4. Apply move.

For Goats:

1. Click an empty valid position during placement.
2. Place Goat.

For movable Goats:

1. Select Goat.
2. Show legal destinations.
3. Move Goat.

Invalid actions should provide visual feedback without crashing.

---

# 9. AI

Implement optional computer-player support.

Start with a simple but clean architecture:

```go
type Player interface {
    ChooseMove(state game.GameState) game.Move
}
```

Implement:

```text
HumanPlayer
RandomAI
MinimaxAI
```

The AI should operate ONLY against the game engine.

The AI must never access Fyne widgets.

Use:

- Minimax
- Alpha-beta pruning
- configurable search depth

Example:

```text
Easy   → Random / shallow search
Medium → Minimax depth 3-4
Hard   → Minimax depth 5+
```

Do not block the Fyne UI while AI is thinking.

Run AI calculations asynchronously.

---

# 10. AI Evaluation

Create separate evaluation logic.

Consider:

### Tiger advantages

- Number of captured Goats
- Available Tiger moves
- Available captures
- Tiger mobility
- Goat blocking potential

### Goat advantages

- Number of remaining Goats
- Tiger mobility reduction
- Number of blocked Tigers
- Goat distribution
- Strategic control of important positions

Do not over-engineer the evaluation initially.

Make the evaluator easy to improve later.

---

# 11. Undo / Redo

Implement move history.

Support:

```text
Undo
Redo
Restart
```

Store complete or safely reconstructable game states.

AI games should also support undo without corrupting turn state.

---

# 12. Save / Load

Implement optional local save functionality.

Use a stable format such as JSON.

Example:

```text
~/.config/bagh-bakri/
    saves/
```

The saved game must contain enough information to restore the exact game state.

---

# 13. Sound and Animation

Keep these optional and isolated from the engine.

Potential effects:

- piece placement
- Tiger movement
- Goat capture
- invalid move
- victory
- defeat

Do not make the game engine depend on audio libraries.

---

# 14. Web / WebGL Preparation

The first implementation should be Fyne desktop.

However, design the project so the engine can later compile to WebAssembly.

The core package must avoid:

- OS-specific APIs
- Fyne imports
- filesystem dependencies
- desktop-only APIs
- goroutine assumptions that prevent WASM compatibility

If practical, add a future frontend abstraction:

```text
Game Engine
     │
 ┌───┴───────────┐
 │               │
Fyne           WASM/Web
Frontend       Frontend
```

Do NOT build a separate web implementation unless it is genuinely useful.

If WebGL is attempted, prefer a minimal WebAssembly-compatible rendering layer rather than rewriting game logic.

---

# 15. CLI / Debug Mode

Also provide a useful CLI mode for testing the engine.

Example:

```bash
go run ./cmd/bagh-bakri --cli
```

Allow:

```text
show board
legal moves
move
undo
restart
status
```

This is primarily for development and debugging.

The CLI must use the exact same game engine as Fyne.

---

# 16. Testing

Testing is mandatory.

Write unit tests for:

### Board

- Position connectivity
- Valid neighbors
- Capture paths
- Invalid paths

### Movement

- Valid Tiger movement
- Invalid Tiger movement
- Valid Tiger capture
- Invalid capture
- Goat placement
- Goat movement

### Game state

- Turn switching
- Goat count
- Captured Goat count
- Game-over detection
- Winner detection

### AI

- Legal move generation
- AI never produces illegal moves
- AI does not mutate original state

### Regression

Every discovered bug must receive a regression test.

Run:

```bash
go test ./...
go vet ./...
gofmt -w .
```

---

# 17. Performance

Optimize only after correctness.

However, design for efficient AI simulation.

Avoid:

- unnecessary allocations
- copying huge structures
- UI-dependent game calculations
- repeated board topology reconstruction

Use compact representations where beneficial.

Benchmark:

```bash
go test -bench=. ./...
```

Add benchmarks for:

```text
LegalMoves
ApplyMove
Minimax
```

---

# 18. Error Handling

Never panic for normal invalid gameplay.

Return errors:

```go
var (
    ErrInvalidMove
    ErrOccupiedPosition
    ErrWrongTurn
    ErrGameOver
)
```

UI should convert these errors into user-friendly messages.

---

# 19. UX

Provide:

### Main Menu

```text
Bagh-Bakri

[ Player vs Player ]
[ Player vs Computer ]
[ Settings ]
[ Exit ]
```

### In-game

```text
Tiger-Goat

Turn: Tiger
Goats remaining: 14
Goats captured: 6

[ Undo ] [ Restart ] [ Menu ]
```

Use clear visual feedback.

Avoid clutter.

The game should feel like a small polished desktop game rather than a developer prototype.

---

# 20. Settings

Add:

```text
Difficulty
Sound
Animation
Board theme
First player
```

Persist settings locally if practical.

---

# 21. Documentation

README must include:

1. Game rules
2. How to build
3. How to run
4. CLI mode
5. Fyne requirements
6. Architecture
7. Board topology
8. AI architecture
9. Testing
10. Cross-platform build instructions
11. Future WebAssembly/WebGL architecture

Include commands such as:

```bash
go mod tidy
go test ./...
go vet ./...
go run ./cmd/bagh-bakri
```

---

# 22. Development Process

Work incrementally.

Do NOT generate a giant untested implementation in one step.

Follow this sequence:

## Phase 1 — Project

Create:

- Go module
- directory structure
- minimal executable
- README

Verify build.

## Phase 2 — Board

Implement:

- positions
- topology
- board tests

Verify all connectivity.

## Phase 3 — Game Engine

Implement:

- pieces
- game state
- moves
- rules
- captures
- winner detection

Write comprehensive tests.

## Phase 4 — CLI

Build a playable CLI version.

Verify the entire game without Fyne.

## Phase 5 — Fyne

Build the graphical board.

Connect it to the existing engine.

Do not duplicate game logic.

## Phase 6 — AI

Implement:

- Random AI
- Minimax
- Alpha-beta
- difficulty levels

Benchmark AI.

## Phase 7 — UX

Add:

- animations
- sound
- undo/redo
- save/load
- settings
- polished visuals

## Phase 8 — Cross-platform

Verify:

```text
Linux
Windows
macOS
```

If practical, investigate:

```text
GOOS=js GOARCH=wasm
```

and document what can and cannot be reused for the web frontend.

---

# 23. Agent Rules

You are operating as a coding agent.

Before modifying code:

1. Inspect the repository.
2. Determine what already exists.
3. Do not overwrite working code unnecessarily.
4. Preserve existing functionality.
5. Prefer small, testable changes.
6. Run tests after each major phase.
7. Fix compilation errors immediately.
8. Do not claim something works without testing it.

When making architectural decisions, prefer:

```text
simple
explicit
idiomatic Go
testable
portable
maintainable
```

Avoid unnecessary frameworks and abstractions.

Do not implement speculative features before the core game works.

---

# 24. Definition of Done

The project is complete when:

- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] code is formatted
- [ ] Fyne desktop application starts
- [ ] complete Bagh-Bakri game is playable
- [ ] Tiger movement works
- [ ] Tiger capture works
- [ ] Goat placement works
- [ ] Goat movement works according to documented rules
- [ ] win/lose conditions work
- [ ] Player vs Player works
- [ ] Player vs AI works
- [ ] AI produces legal moves
- [ ] undo/restart works
- [ ] game state is not coupled to Fyne
- [ ] CLI mode works
- [ ] README is complete
- [ ] architecture is ready for a future WebAssembly/WebGL frontend

Start by inspecting the repository and then implement **Phase 1**. After completing each phase, test it before proceeding to the next phase.