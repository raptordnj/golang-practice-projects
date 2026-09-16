# বাঘবন্দী · Bagh-Bandi

A complete, playable **Tiger-and-Goats** board game in Go — the Bangladeshi
variant: **one tiger (বাঘ) against sixteen goats (ছাগল)** on the 5×5 Alquerque
board that gets scratched into courtyards and drawn on paper all over Bengal.

Desktop GUI with [Fyne](https://fyne.io), a full terminal frontend, minimax AI
with alpha-beta pruning, undo/redo, save/load, three board themes — and the
same binary compiles to WebAssembly and renders through WebGL.

```
    a   b   c   d   e
1   ছ - ছ - ছ - . - .   1
    | X |   | X |   |
2   ছ - . - . - . - ছ   2
    |   | X |   | X |
3   ছ - . - বাঘ - . - ছ  3
    | X |   | X |   |
4   . - ছ - . - ছ - .   4
    |   | X |   | X |
5   . - . - ছ - . - .   5
```

---

## Contents

1. [The rules](#the-rules)
2. [Board topology](#board-topology)
3. [Building](#building)
4. [Running](#running)
5. [CLI / debug mode](#cli--debug-mode)
6. [Architecture](#architecture)
7. [AI architecture](#ai-architecture)
8. [Rendering, WebGL and DirectX](#rendering-webgl-and-directx)
9. [WebAssembly build](#webassembly-build)
10. [Testing and benchmarks](#testing-and-benchmarks)
11. [Cross-platform builds](#cross-platform-builds)
12. [Saved games and settings](#saved-games-and-settings)
13. [Project layout](#project-layout)
14. [Licence](#licence)

---

## The rules

This is one specific, documented variant. Regional rules differ; nothing here
is invented silently, and every number below is configurable in
`game.RuleSet`.

### Pieces and setup

* **1 tiger**, starting on the centre point (c3).
* **16 goats**, all in the goat player's hand at the start.
* **The goat player moves first.**

### Placement phase

On each of their turns the goat player drops **one goat on any empty point**.
Goats already on the board **may not move** until all sixteen have been
placed.

### Movement phase

Once all sixteen goats are down, the goat player instead **slides one goat
along a drawn line to an adjacent empty point**.

### The tiger

On every turn the tiger does exactly one of:

1. **Slide** along a drawn line to an adjacent empty point, or
2. **Jump** in a straight line over exactly **one** adjacent goat, landing on
   the empty point directly beyond it, and **removing that goat**.

A jump takes exactly one goat. There are **no chained jumps**. The jump must
follow drawn lines for *both* steps — the tiger can never cross a gap where no
line is drawn. Goats never capture and never jump.

### Winning

| Outcome | Condition |
|---|---|
| **Tiger wins** | eats `TigerWinCaptures` goats (default **9**), or the goat player has no legal move at all |
| **Goats win** | the tiger has no legal move — neither slide nor jump |
| **Draw** | `HalfmoveLimit` plies (default **100**) with no capture and no placement |

The nine-capture threshold is the point at which seven goats remain, which is
too few to build a safe fence around the tiger. Both this and compulsory
capture can be changed in Settings, or with `--win-captures` and
`--forced-capture` on the command line.

### Configurable rules

Everything above lives in one struct, `game.RuleSet`:

| Field | Default | Meaning |
|---|---|---|
| `TotalGoats` | 16 | goats the goat player owns |
| `TigerStart` | c3 | the tiger's opening point |
| `FirstPlayer` | Goat | who moves first |
| `TigerWinCaptures` | 9 | captures that win for the tiger |
| `GoatsMoveDuringPlacement` | false | may placed goats slide before the hand is empty |
| `ForcedCapture` | false | must the tiger take when it can |
| `HalfmoveLimit` | 100 | plies without progress before a draw |

---

## Board topology

The board is the classic **Alquerque** pattern: 25 points in a 5×5 grid.

* Every point is joined **orthogonally** to its neighbours.
* **Diagonals are drawn only through points whose `(row + col)` is even.**

That is what produces the traditional eight-pointed star figure, and it is why
different points have different numbers of lines:

```
degree   a    b    c    d    e
   1     3    3    5    3    3
   2     3    8    4    8    3
   3     5    4    8    4    5
   4     3    8    4    8    3
   5     3    3    5    3    3
```

The centre (c3) has **eight** lines; the corners have **three**. This is the
whole strategic shape of the game: the tiger wants the centre, the goats want
to push it into a corner.

**The topology is defined explicitly in `internal/game/board.go` and is never
derived from screen coordinates.** The `X`/`Y` fields on a `Position` are
normalised 0–1 values that exist purely so a renderer knows where to draw; the
engine never reads them. Capture paths are precomputed into a jump table at
start-up, so move generation never re-derives geometry.

---

## Building

Requires **Go 1.25+**. Fyne needs a C compiler and the system OpenGL/X11
development headers:

```bash
# Debian / Ubuntu
sudo apt install golang gcc libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install golang gcc libX11-devel libXcursor-devel libXrandr-devel \
                 libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel

# macOS — Xcode command line tools
xcode-select --install

# Windows — a gcc such as MSYS2 / TDM-GCC
```

Then:

```bash
go mod tidy
go build ./...
go build -o bagh-bandi ./cmd/bagh-bandi
```

---

## Running

```bash
go run ./cmd/bagh-bandi           # desktop game
go run ./cmd/bagh-bandi --cli     # terminal game
go run ./cmd/bagh-bandi --help    # every option
```

The main menu offers **Player vs Player**, **Player vs Computer**, resume,
settings and the in-game rules sheet. In a game you get a turn indicator, the
goats-in-hand and goats-on-board counters, a captured-goats progress bar,
undo/redo, restart, save, load and a way back to the menu.

**Playing with the mouse.** During the placement phase, click any empty point
to drop a goat. Otherwise click a piece of your own to select it — its legal
destinations appear as dots, and a capture shows a red cross over the goat
that would be eaten — then click a destination. An illegal click flashes the
point red and explains itself in the side panel; it never crashes and never
changes the board.

---

## CLI / debug mode

The terminal frontend runs the **same engine** as the GUI — there is no second
copy of the rules anywhere in this program.

```bash
go run ./cmd/bagh-bandi --cli                        # two humans
go run ./cmd/bagh-bandi --cli --tiger-ai hard        # you play the goats
go run ./cmd/bagh-bandi --cli --goat-ai medium       # you play the tiger
go run ./cmd/bagh-bandi --tiger-ai medium --goat-ai easy   # watch two AIs
```

| Command | Effect |
|---|---|
| `show` | draw the board, with its real diagonals |
| `status` | turn, phase, goat counts, result |
| `moves [point]` | list legal moves, optionally only from one point |
| `move <from> <to>` | move a piece — captures are resolved for you |
| `place <point>` | place a goat during the placement phase |
| `undo` / `redo` | step through the history |
| `restart` | start again |
| `history` | list the moves played |
| `quit` | leave |

Points are named `a1` (top-left) through `e5` (bottom-right).

---

## Architecture

Strictly layered, with the dependency arrow only ever pointing downwards:

```
  cmd/bagh-bandi          flags, wiring
        │
   ┌────┴─────┬───────────────┐
   ▼          ▼               ▼
internal/ui   internal/cli   (future: WASM frontend)
   │          │
   └────┬─────┘
        ▼
   internal/ai            minimax, evaluation
        │
        ▼
   internal/game          board, rules, state, engine
```

**`internal/game` imports nothing but the standard library.** No Fyne, no
audio, no filesystem, no OS-specific calls. That constraint is what makes the
engine reusable for WebAssembly, for mobile, and for the AI's millions of
simulated positions.

The key design decision is that `GameState` is a **pure value type**: an array
of occupancies plus a handful of integers, with no slices, maps or pointers.
`next := state` is therefore a complete, independent copy, and `ApplyMove`
takes and returns states *by value*. The AI can search freely and can never
corrupt the real game — there is a test that asserts exactly that.

| Package | Responsibility |
|---|---|
| `internal/game` | board graph, pieces, moves, rules, state, engine, undo/redo, save format |
| `internal/ai` | `Player` interface, `RandomAI`, `MinimaxAI`, evaluation |
| `internal/ui` | Fyne widgets, board renderer, geometry, controller, settings, storage, sound seam |
| `internal/cli` | terminal frontend |
| `assets` | embedded SVG artwork |
| `tools/serve` | static server for the WebAssembly build |

The UI is split so that the part worth testing can be tested: `GameController`
holds all the click-interpretation logic and talks to the board through the
small `BoardPresenter` interface, so its behaviour is verified headlessly,
with no graphics driver involved.

### Errors

Normal invalid gameplay never panics. The engine returns
`ErrInvalidMove`, `ErrOccupiedPosition`, `ErrWrongTurn`, `ErrGameOver`,
`ErrOutOfBounds`, `ErrNothingToUndo` or `ErrNothingToRedo`, and the UI turns
each into a sentence a player can act on.

---

## AI architecture

```go
type Player interface {
    Name() string
    ChooseMove(ctx context.Context, rules *game.TraditionalBaghBandiRules,
        state game.GameState) (game.Move, bool)
}
```

The AI touches nothing but the engine — no widgets, no globals, no shared
mutable state.

| Difficulty | Player | Search |
|---|---|---|
| Easy (সহজ) | `RandomAI` | uniform random legal move |
| Medium (মাঝারি) | `MinimaxAI` | alpha-beta, depth 4 |
| Hard (কঠিন) | `MinimaxAI` | alpha-beta, depth 6 |

* **Minimax with alpha-beta pruning.** The tiger maximises, the goats
  minimise.
* **Iterative deepening.** Each depth completes before the next begins, so an
  interrupted search still returns the best move from the last finished depth.
* **Move ordering.** Captures are searched first, and the previous iteration's
  best move is promoted to the front — both produce far more cutoffs.
* **Allocation-free.** Per-ply move buffers are reused across the whole
  search.
* **Cancellable.** `ChooseMove` takes a `context.Context`; the UI gives it a
  budget (2 s by default, adjustable in Settings) and the search returns
  promptly when it expires.
* **Asynchronous.** The search runs on its own goroutine and hands its answer
  back to the UI thread through `fyne.Do`. The interface never freezes. If the
  player restarted or undid while the computer was thinking, the stale answer
  is discarded — there is a regression test for that.

### Evaluation

`internal/ai/evaluation.go`, scored in centi-goats from the tiger's side, with
every weight in one `Weights` struct so it can be retuned without touching the
search:

**For the tiger** — goats captured, goats still in hand, tiger mobility,
captures available right now, and proximity to the centre (eight lines at c3,
three in a corner).

**For the goats** — goat mobility, and *goat safety*: a goat is safe when
every landing point beyond it is blocked, which is the whole idea behind
building goats into solid blocks.

Terminal positions are scored by ply, so a win in 1 beats a win in 5 and the
AI finishes games instead of shuffling.

---

## Rendering, WebGL and DirectX

The board is drawn by a custom Fyne widget (`internal/ui/board.go` and
`renderer.go`) and every pixel goes through the GPU:

| Platform | Graphics path |
|---|---|
| Linux, macOS | OpenGL, via GLFW |
| Windows | OpenGL, via GLFW — mapped onto **Direct3D** when the system or driver uses ANGLE, which is the normal path on modern Windows |
| Browser | **WebGL**, via Fyne's `gl-js` backend under `GOOS=js` |

The important part is that **this is the same code on all three**. The
renderer builds its canvas objects once and thereafter only repositions and
recolours them, so a redraw allocates nothing and the GPU keeps its textures —
which is what makes resizing smooth.

**Nothing uses fixed pixel coordinates.** `internal/ui/geometry.go` maps the
abstract board graph onto whatever space the window currently gives it: the
board stays square and centred in any window shape, and the lines, points,
pieces and highlights all scale together. It is tested at sizes from 240 px to
1000 px, and for non-square windows.

**The pieces are real artwork**, not coloured counters: `assets/icons/tiger.svg`
is a Royal Bengal tiger and `assets/icons/goat.svg` a Black Bengal goat — the
breed that actually grazes across Bangladesh. Both are hand-written SVG using
flat shapes only (no gradients, no filters), so they rasterise identically
through Fyne on the desktop and in the browser, and stay sharp at any board
size or display scaling. They are embedded into the binary with `go:embed`, so
a build is a single file and a WebAssembly build needs no filesystem.

---

## WebAssembly build

The whole program — engine, AI *and* the Fyne UI — already compiles to
WebAssembly and renders through WebGL. Nothing had to be rewritten, because no
core package touches the filesystem, the OS or a desktop-only API.

```bash
./build-wasm.sh                 # builds into build/web/
VERSION=1.0.0 ./build-wasm.sh   # ...with a version stamped in
go run ./tools/serve build/web          # http://localhost:8080
go run ./tools/serve build/web :8088    # ...or any other free port
```

That produces:

| File | Size | |
|---|---|---|
| `main.wasm` | 36 MB | the whole game |
| `main.wasm.gz` | 12 MB | the same thing, pre-compressed — this is what actually goes over the wire |
| `wasm_exec.js` | 17 KB | Go's WebAssembly runtime shim, copied from your `GOROOT` |
| `index.html` | 2 KB | host page; checks for WebAssembly and WebGL before starting |

`tools/serve` exists for two things a quick static server usually gets wrong:
browsers refuse to instantiate a `.wasm` file served with the wrong MIME type,
and a Go wasm binary is big enough that serving the pre-compressed copy
matters. It hands `main.wasm.gz` to any browser that accepts gzip, with
`Content-Type: application/wasm` and `Content-Encoding: gzip`, and falls back
to the plain file otherwise.

Any other static host works too, as long as it sets the wasm MIME type and
(ideally) serves the file compressed.

What is reusable, and what is not:

| Layer | WebAssembly |
|---|---|
| `internal/game` | reused **unchanged** — standard library only |
| `internal/ai` | reused **unchanged** — goroutines and `context` both work |
| `internal/ui` widgets and renderer | reused **unchanged** — Fyne renders to WebGL |
| `internal/ui/storage.go` | **not** reused — it uses `os.UserConfigDir`. Swap in a `localStorage`-backed implementation; it is the only file in the project that touches a filesystem, which is exactly why it is isolated there |
| `internal/cli` | not meaningful in a browser |

If a bespoke WebGL frontend is ever wanted instead of Fyne's, it plugs in at
the same seam the Fyne UI uses — `BoardPresenter` and `HUD` — and the engine,
the AI and the controller all come along unchanged.

---

## Testing and benchmarks

```bash
go test ./...
go vet ./...
gofmt -l .        # silence means everything is formatted
go test -bench=. ./...
```

There are around 90 tests. They cover:

* **Board** — the exact degree of all 25 points, symmetry of every edge, the
  absence of diagonals through odd points, collinearity of every jump path,
  and rejection of off-board ids.
* **Movement** — legal and illegal tiger slides, legal and illegal captures,
  captures onto occupied and off-board landing points, goat placement, the ban
  on goats moving during placement, and goat movement afterwards.
* **Game state** — turn switching, goat counters, capture counters, the
  no-progress draw, game-over detection and winner detection.
* **AI** — every difficulty, on both sides, at every stage, always returns a
  legal move; the search never mutates the state it is given; a free capture
  is taken; a winning fence is played; a cancelled search still returns a
  usable move.
* **Controller** — click interpretation, out-of-turn clicks, invalid-click
  feedback, undo across the AI's reply, restart, and save/load round trips.
* **Rendering** — the board is captured headlessly and compared pixel by
  pixel, so a silently blank board or an invisible highlight fails the build.
* **Storage** — settings round trips, repair of corrupt settings, and the
  guarantee that a save name can never escape the saves directory.
* **Regression** — stale AI moves after a restart, undo handing back the wrong
  player's turn, and the draw clock resetting on placement.

Property-style tests play 300 randomised games end to end, asserting after
every single ply that exactly one tiger is on the board, that the occupancy
grid agrees with the counters, and that no unfinished position is ever without
a legal move.

Benchmarks cover `LegalMoves`, `ApplyMove`, `Evaluate` and minimax at depth 4
and 6. On a Ryzen 5 5600G:

```
BenchmarkAppendLegalMovesTiger-12    19.22 ns/op    0 B/op   0 allocs/op
BenchmarkApplyMoveCapture-12         20.62 ns/op    0 B/op   0 allocs/op
BenchmarkMinimaxDepth4-12             0.73 ms/op
BenchmarkMinimaxDepth6-12            14.81 ms/op
```

---

## Cross-platform builds

Fyne uses cgo, so cross-compiling needs a cross toolchain; building natively on
each platform is simplest.

```bash
# native
go build -o bagh-bandi ./cmd/bagh-bandi

# Windows, from Linux, with mingw-w64 installed
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags "-H windowsgui" -o bagh-bandi.exe ./cmd/bagh-bandi

# WebAssembly
./build-wasm.sh
```

The [`fyne` CLI](https://docs.fyne.io/started/packaging) will produce proper
`.app`, `.exe` and `.deb` bundles:

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
fyne package -os darwin   # or windows, linux, wasm, android, ios
```

Stamp a version in with `go build -ldflags "-X main.version=1.0.0"`.

---

## Saved games and settings

Both live under the OS config directory — `~/.config/bagh-bandi` on Linux,
`~/Library/Application Support/bagh-bandi` on macOS,
`%AppData%\bagh-bandi` on Windows:

```
bagh-bandi/
├── settings.json
└── saves/
    ├── autosave.json
    └── <your saved games>.json
```

A save stores the ruleset and the **move list**, not a board dump. Replaying
the moves through the rules reconstructs every intermediate state exactly, so
the full undo history survives a save/load round trip — and a corrupt or
hand-forged file cannot load the game into an illegal position, because every
move is re-validated on the way in.

The game in progress is autosaved when you close the window, and **Resume last
game** on the main menu picks it up.

Settings cover difficulty, which side you play, board theme, sound,
animation, the tiger's capture threshold, compulsory capture, and the
computer's thinking time. Out-of-range values in a hand-edited file are
repaired rather than trusted.

---

## Project layout

```
bagh-bandi/
├── cmd/bagh-bandi/main.go       flags, and the choice of frontend
├── internal/
│   ├── game/                    the engine — standard library only
│   │   ├── board.go             topology and the precomputed jump table
│   │   ├── position.go          points
│   │   ├── piece.go             pieces and players
│   │   ├── move.go              moves and a1–e5 notation
│   │   ├── state.go             GameState, a pure value type
│   │   ├── rules.go             RuleSet and TraditionalBaghBandiRules
│   │   ├── engine.go            history, undo/redo, restart
│   │   ├── snapshot.go          the save format
│   │   ├── errors.go            the error sentinels
│   │   ├── game_test.go
│   │   └── bench_test.go
│   ├── ai/
│   │   ├── player.go            the Player interface and difficulties
│   │   ├── random.go            RandomAI
│   │   ├── minimax.go           alpha-beta with iterative deepening
│   │   ├── evaluation.go        weights and heuristics
│   │   └── ai_test.go
│   ├── ui/
│   │   ├── app.go               window, menu, HUD, dialogs
│   │   ├── controller.go        click interpretation and AI orchestration
│   │   ├── board.go             the board widget
│   │   ├── renderer.go          the drawing routine
│   │   ├── geometry.go          layout — no fixed pixel coordinates
│   │   ├── theme.go             three board themes
│   │   ├── settings.go          preferences
│   │   ├── storage.go           the only file that touches a filesystem
│   │   ├── sound.go             the audio seam
│   │   └── *_test.go
│   └── cli/                     the terminal frontend
├── assets/
│   ├── assets.go                go:embed of the artwork
│   ├── icons/tiger.svg          Royal Bengal tiger
│   ├── icons/goat.svg           Black Bengal goat
│   └── sounds/                  drop .wav files here to enable sound
├── tools/serve/                 static server for the wasm build
├── web/index.html               WebAssembly host page
├── build-wasm.sh
├── go.mod
├── LICENSE
└── README.md
```

### Sound

Sound is wired end to end — the controller raises `place`, `move`, `capture`,
`invalid`, `victory` and `defeat` cues and the UI routes them to a
`SoundPlayer` — but **no audio ships with this repository**, because we do not
ship audio we do not own. Fyne has no audio API of its own, so
`internal/ui/sound_play.go` is deliberately an empty seam: drop `.wav` files
named after the effects into `assets/sounds/`, wire a player such as
[oto](https://github.com/hajimehoshi/oto) or
[beep](https://github.com/faiface/beep) into `playResource`, and every cue in
the game starts working with no change anywhere else. Keeping it a seam is
what lets `internal/game` stay free of audio dependencies and lets the whole
program build for WebAssembly.

---

## Licence

MIT — see [LICENSE](LICENSE).

The game itself is traditional and belongs to everybody who has ever drawn the
board in the dirt.
