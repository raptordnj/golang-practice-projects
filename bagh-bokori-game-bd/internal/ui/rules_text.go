package ui

// RulesMarkdown is the in-game rules sheet. It is the same text as the rules
// section of README.md, kept here so the game explains itself without needing
// the repository.
const RulesMarkdown = `
## বাঘবন্দী - Bagh-Bandi

**One tiger (বাঘ) against sixteen goats (ছাগল)** on the 5×5 Alquerque board
played across Bangladesh.

### The board

Twenty-five points joined by drawn lines. Every point connects to its
orthogonal neighbours; diagonals are drawn only through the points that carry
them, which is why some points have eight lines and others only three. A piece
may only travel along a drawn line - never across a gap.

### Setting up

The tiger starts on the centre point. The goat player holds all sixteen goats
and moves first.

### Placement phase

On each turn the goat player drops one goat onto any empty point. Goats
already on the board may **not** move until all sixteen have been placed.

### Movement phase

Once all sixteen goats are down, the goat player instead slides one goat along
a line to an adjacent empty point.

### The tiger

On every turn the tiger either:

1. slides along a line to an adjacent empty point, or
2. **jumps** in a straight line over exactly one adjacent goat onto the empty
   point directly beyond it, removing that goat.

Only one goat is taken per jump, and there are no chained jumps.
Goats never capture and never jump.

### Winning

* **The tiger wins** by eating enough goats - nine by default, adjustable in
  Settings - or if the goats have no legal move at all.
* **The goats win** when the tiger is fenced in: no slide and no jump.
* The game is a **draw** after a hundred half-moves with no capture and no
  placement.

### How to play well

*As the goats*, place in solid blocks. A goat is safe when the point behind it
is occupied, so build outward from the edges and never leave a lone goat with
an empty point beyond it. Crowd the tiger toward a corner, where it has only
three lines.

*As the tiger*, stay near the centre while you can - the middle point has
eight lines, a corner has three. Take goats that open a second threat, and
avoid stepping into a pocket the goats can seal behind you.
`
