package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"bagh-bandi/internal/ai"
	"bagh-bandi/internal/game"
)

func TestSettingsRoundTrip(t *testing.T) {
	want := DefaultSettings()
	want.Difficulty = ai.Hard
	want.Sound = false
	want.BoardTheme = BoardThemes[1].Name

	data, err := EncodeSettings(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeSettings(data)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// A hand-edited or corrupt settings file must never produce an unplayable
// game; every out-of-range value is repaired.
func TestSettingsNormaliseRepairsBadValues(t *testing.T) {
	bad := Settings{
		Difficulty:       ai.Difficulty(99),
		HumanSide:        game.Player(7),
		ThinkSeconds:     -5,
		TigerWinCaptures: 999,
		BoardTheme:       "Nonexistent",
	}
	got := bad.Normalise()

	if got.Difficulty != ai.Medium {
		t.Errorf("difficulty = %v", got.Difficulty)
	}
	if got.HumanSide != game.GoatPlayer {
		t.Errorf("human side = %v", got.HumanSide)
	}
	if got.ThinkTime() <= 0 {
		t.Errorf("think time = %v", got.ThinkTime())
	}
	if got.TigerWinCaptures != game.DefaultRuleSet().TigerWinCaptures {
		t.Errorf("tiger win captures = %d", got.TigerWinCaptures)
	}
	if got.Theme().Name != BoardThemes[0].Name {
		t.Errorf("theme = %q", got.Theme().Name)
	}
}

func TestDecodeSettingsOnGarbageReturnsDefaults(t *testing.T) {
	got, err := DecodeSettings([]byte("not json"))
	if err == nil {
		t.Error("expected a parse error")
	}
	if got != DefaultSettings() {
		t.Error("garbage input did not fall back to defaults")
	}
}

func TestSettingsBuildRules(t *testing.T) {
	s := DefaultSettings()
	s.ForcedCapture = true
	s.TigerWinCaptures = 4
	cfg := s.Rules()

	if !cfg.ForcedCapture || cfg.TigerWinCaptures != 4 {
		t.Errorf("rules = %+v", cfg)
	}
	if cfg.TotalGoats != 16 {
		t.Errorf("total goats = %d, want 16", cfg.TotalGoats)
	}
}

func TestStorageSettingsPersist(t *testing.T) {
	s := NewStorageAt(t.TempDir())

	if got := s.LoadSettings(); got != DefaultSettings() {
		t.Error("empty storage should return defaults")
	}

	want := DefaultSettings()
	want.Sound = false
	want.Difficulty = ai.Hard
	if err := s.SaveSettings(want); err != nil {
		t.Fatal(err)
	}
	if got := s.LoadSettings(); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestStorageSavesAndListsGames(t *testing.T) {
	s := NewStorageAt(t.TempDir())
	e := game.NewDefaultEngine()
	if err := e.Play(game.PlaceMove(game.ID(0, 0))); err != nil {
		t.Fatal(err)
	}

	path, err := s.SaveGame("my game", e.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(path) != ".json" {
		t.Errorf("save path = %q", path)
	}

	names := s.ListGames()
	if len(names) != 1 || names[0] != "my game" {
		t.Fatalf("listed %v", names)
	}

	snap, err := s.LoadGame("my game")
	if err != nil {
		t.Fatal(err)
	}
	restored := game.NewDefaultEngine()
	if err := restored.LoadSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	if restored.State() != e.State() {
		t.Error("the loaded game differs from the saved one")
	}

	if err := s.DeleteGame("my game"); err != nil {
		t.Fatal(err)
	}
	if len(s.ListGames()) != 0 {
		t.Error("delete did not remove the save")
	}
}

// A save name must never be able to escape the saves directory.
func TestSaveNameCannotTraversePaths(t *testing.T) {
	root := t.TempDir()
	s := NewStorageAt(root)

	path, err := s.SaveGame("../../escape", game.NewDefaultEngine().Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, s.SavesDir()) {
		t.Errorf("save escaped to %q", path)
	}
	if strings.Contains(filepath.Base(path), "..") {
		t.Errorf("file name contains traversal: %q", filepath.Base(path))
	}
}

func TestSanitiseName(t *testing.T) {
	cases := map[string]string{
		"":              "game",
		"   ":           "game",
		"../../etc":     "etc",
		"my/save":       "mysave",
		"good name-1_2": "good name-1_2",
	}
	for in, want := range cases {
		if got := sanitiseName(in); got != want {
			t.Errorf("sanitiseName(%q) = %q, want %q", in, got, want)
		}
	}
	if got := sanitiseName(strings.Repeat("a", 200)); len(got) != 60 {
		t.Errorf("long name produced %d characters", len(got))
	}
}

func TestUnavailableStorageDoesNotPanic(t *testing.T) {
	s := NewStorageAt("")
	if s.Available() {
		t.Fatal("empty root should be unavailable")
	}
	if got := s.LoadSettings(); got != DefaultSettings() {
		t.Error("unavailable storage should return defaults")
	}
	if err := s.SaveSettings(DefaultSettings()); err == nil {
		t.Error("saving to unavailable storage should error")
	}
	if _, err := s.SaveGame("x", game.NewDefaultEngine().Snapshot()); err == nil {
		t.Error("saving a game to unavailable storage should error")
	}
	if names := s.ListGames(); names != nil {
		t.Errorf("listed %v from unavailable storage", names)
	}
}

func TestSoundPlayerIsSafeWhenDisabled(t *testing.T) {
	s := NewFyneSound(false, nil)
	s.Play(EffectCapture) // must not panic
	if s.Enabled() {
		t.Error("sound reported enabled")
	}
	s.SetEnabled(true)
	s.Play(EffectVictory) // no loader: still a no-op
	if !s.Enabled() {
		t.Error("sound did not turn on")
	}
}

func TestEffectNamesAreDistinct(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range []Effect{EffectPlace, EffectMove, EffectCapture, EffectInvalid, EffectVictory, EffectDefeat} {
		name := e.String()
		if name == "" || seen[name] {
			t.Errorf("effect %d has a missing or duplicate name %q", e, name)
		}
		seen[name] = true
		if !strings.HasPrefix(e.assetName(), "sounds/") {
			t.Errorf("asset name %q is not under sounds/", e.assetName())
		}
	}
}

func TestThemesAreComplete(t *testing.T) {
	if len(BoardThemes) < 2 {
		t.Fatal("expected several board themes")
	}
	seen := map[string]bool{}
	for _, th := range BoardThemes {
		if th.Name == "" || th.Bangla == "" {
			t.Errorf("theme %q is missing a name", th.Name)
		}
		if seen[th.Name] {
			t.Errorf("duplicate theme name %q", th.Name)
		}
		seen[th.Name] = true
		if th.Background == nil || th.PieceBacking == nil || th.PieceOutline == nil || th.Line == nil {
			t.Errorf("theme %q has unset colours", th.Name)
		}
	}
	if ThemeByName("no such theme").Name != BoardThemes[0].Name {
		t.Error("unknown theme did not fall back to the first")
	}
}
