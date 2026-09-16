# Sound effects

This directory is intentionally empty: the game ships no audio, because we do
not distribute audio we do not own.

Sound is wired end to end all the same. Drop WAV files with these names here
and wire a player into `internal/ui/sound_play.go`, and every cue starts
working with no other change:

| File | Played when |
|---|---|
| `place.wav` | a goat is placed |
| `move.wav` | a piece slides along a line |
| `capture.wav` | the tiger eats a goat |
| `invalid.wav` | an action is refused |
| `victory.wav` | you win |
| `defeat.wav` | you lose |

See the "Sound" section of the top-level README for the details.
