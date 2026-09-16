# Bagh-Bakri (Tiger-Goat) Game

A cross-platform strategy board game implemented in Go using Fyne for the desktop GUI.

## Game Rules

Bagh-Bakri (Tiger-Goat) is a traditional Indian strategy board game played between:
- **4 Tigers** (Bagh)
- **20 Goats** (Bokri)

### Objective
- **Tigers** win by capturing 5 goats or blocking all goat movement
- **Goats** win by blocking all tiger movement/captures

### Rules
1. **Placement Phase**: Goats are placed one at a time on empty positions
2. **Movement Phase**: After all 20 goats are placed, both sides alternate moving
3. **Tiger Movement**: Tigers can move to adjacent empty positions or jump over a goat to capture it
4. **Goat Movement**: Goats move to adjacent empty positions (no captures)
5. **Capture**: A tiger jumps over an adjacent goat to an empty position beyond

## Building

```bash
go mod tidy
go build ./cmd/bagh-bakri
```

## Running

### GUI Mode
```bash
go run ./cmd/bagh-bakri
```

### CLI Mode (debug)
```bash
go run ./cmd/bagh-bakri --cli
```

## Fyne Requirements

Fyne requires:
- Linux: GTK3, OpenGL
- macOS: Cocoa
- Windows: GDI+

## Architecture

```
cmd/bagh-bakri/     Entry point
internal/game/      Game engine (no Fyne dependency)
internal/ai/        AI players
internal/ui/        Fyne UI components
```

The game engine is completely independent from Fyne, enabling future WebAssembly/WebGL reuse.

## Board Topology

5x5 grid with diagonal connections forming the traditional Bagh-Bakri pattern.

## AI Architecture

- **RandomAI**: Chooses random legal moves
- **MinimaxAI**: Minimax with alpha-beta pruning, configurable depth
  - Easy: depth 1-2
  - Medium: depth 3-4
  - Hard: depth 5+

## Testing

```bash
go test ./...
go vet ./...
gofmt -w .
```

## Cross-Platform Builds

```bash
# Linux
GOOS=linux GOARCH=amd64 go build ./cmd/bagh-bakri

# Windows
GOOS=windows GOARCH=amd64 go build ./cmd/bagh-bakri

# macOS
GOOS=darwin GOARCH=amd64 go build ./cmd/bagh-bakri
```

## Future: WebAssembly/WebGL

The core engine avoids OS-specific APIs, making it suitable for WebAssembly compilation.
To compile for WASM:

```bash
GOOS=js GOARCH=wasm go build -o bagh-bakri.wasm ./cmd/bagh-bakri
```

Note: Fyne does not support WebAssembly. A separate lightweight frontend (Canvas/WebGL) would be needed for the browser version while reusing `internal/game`.