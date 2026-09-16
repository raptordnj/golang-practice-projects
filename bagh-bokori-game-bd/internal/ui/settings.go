package ui

import (
	"encoding/json"
	"time"

	"bagh-bandi/internal/ai"
	"bagh-bandi/internal/game"
)

// Settings are the player's persisted preferences.
type Settings struct {
	Difficulty   ai.Difficulty `json:"difficulty"`
	Sound        bool          `json:"sound"`
	Animation    bool          `json:"animation"`
	BoardTheme   string        `json:"boardTheme"`
	HumanSide    game.Player   `json:"humanSide"`
	ThinkSeconds float64       `json:"thinkSeconds"`
	// ForcedCapture and TigerWinCaptures expose the two rules people
	// genuinely disagree about regionally.
	ForcedCapture    bool `json:"forcedCapture"`
	TigerWinCaptures int  `json:"tigerWinCaptures"`
}

// DefaultSettings returns sensible starting preferences.
func DefaultSettings() Settings {
	cfg := game.DefaultRuleSet()
	return Settings{
		Difficulty:       ai.Medium,
		Sound:            true,
		Animation:        true,
		BoardTheme:       BoardThemes[0].Name,
		HumanSide:        game.GoatPlayer,
		ThinkSeconds:     2,
		ForcedCapture:    cfg.ForcedCapture,
		TigerWinCaptures: cfg.TigerWinCaptures,
	}
}

// Normalise repairs values that a hand-edited or outdated settings file might
// contain, so a bad file can never put the game into an unplayable state.
func (s Settings) Normalise() Settings {
	out := s
	if out.Difficulty < ai.Easy || out.Difficulty > ai.Hard {
		out.Difficulty = ai.Medium
	}
	if out.HumanSide != game.TigerPlayer && out.HumanSide != game.GoatPlayer {
		out.HumanSide = game.GoatPlayer
	}
	if out.ThinkSeconds < 0.2 || out.ThinkSeconds > 30 {
		out.ThinkSeconds = 2
	}
	if out.TigerWinCaptures < 1 || out.TigerWinCaptures > game.DefaultRuleSet().TotalGoats {
		out.TigerWinCaptures = game.DefaultRuleSet().TigerWinCaptures
	}
	if ThemeByName(out.BoardTheme).Name != out.BoardTheme {
		out.BoardTheme = BoardThemes[0].Name
	}
	return out
}

// ThinkTime is the AI budget as a duration.
func (s Settings) ThinkTime() time.Duration {
	return time.Duration(s.ThinkSeconds * float64(time.Second))
}

// Rules builds the engine ruleset described by these settings.
func (s Settings) Rules() game.RuleSet {
	cfg := game.DefaultRuleSet()
	cfg.ForcedCapture = s.ForcedCapture
	cfg.TigerWinCaptures = s.TigerWinCaptures
	return cfg
}

// Theme returns the chosen board theme.
func (s Settings) Theme() BoardTheme { return ThemeByName(s.BoardTheme) }

// EncodeSettings serialises settings to JSON.
func EncodeSettings(s Settings) ([]byte, error) { return json.MarshalIndent(s, "", "  ") }

// DecodeSettings parses settings JSON, repairing anything out of range.
func DecodeSettings(data []byte) (Settings, error) {
	out := DefaultSettings()
	if err := json.Unmarshal(data, &out); err != nil {
		return DefaultSettings(), err
	}
	return out.Normalise(), nil
}
