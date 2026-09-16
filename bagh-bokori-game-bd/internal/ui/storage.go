package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bagh-bandi/internal/game"
)

// Storage persists settings and saved games under the user's config
// directory, e.g. ~/.config/bagh-bandi on Linux.
//
// This is the only file in the project that touches a filesystem. Keeping it
// here - rather than in internal/game - is what lets the engine build for
// WebAssembly, where a different Storage implementation would use the
// browser's local storage instead.
type Storage struct {
	root string
}

// NewStorage returns storage rooted at the OS config directory. When that is
// unavailable the returned Storage is still usable: every call simply reports
// an error instead of crashing the game.
func NewStorage(appName string) *Storage {
	dir, err := os.UserConfigDir()
	if err != nil {
		return &Storage{}
	}
	return &Storage{root: filepath.Join(dir, appName)}
}

// NewStorageAt returns storage rooted at an explicit directory, used by tests.
func NewStorageAt(root string) *Storage { return &Storage{root: root} }

// Available reports whether a config directory was found.
func (s *Storage) Available() bool { return s.root != "" }

// Root returns the configuration directory.
func (s *Storage) Root() string { return s.root }

// SavesDir returns the directory holding saved games.
func (s *Storage) SavesDir() string { return filepath.Join(s.root, "saves") }

const settingsFile = "settings.json"

// LoadSettings reads the stored preferences, returning the defaults when no
// file exists yet or when the file is unreadable.
func (s *Storage) LoadSettings() Settings {
	if !s.Available() {
		return DefaultSettings()
	}
	data, err := os.ReadFile(filepath.Join(s.root, settingsFile))
	if err != nil {
		return DefaultSettings()
	}
	settings, err := DecodeSettings(data)
	if err != nil {
		return DefaultSettings()
	}
	return settings
}

// SaveSettings writes the preferences.
func (s *Storage) SaveSettings(settings Settings) error {
	if !s.Available() {
		return os.ErrNotExist
	}
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return err
	}
	data, err := EncodeSettings(settings.Normalise())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.root, settingsFile), data, 0o644)
}

// SaveGame writes a snapshot under the given name (without extension).
func (s *Storage) SaveGame(name string, snap game.Snapshot) (string, error) {
	if !s.Available() {
		return "", os.ErrNotExist
	}
	if err := os.MkdirAll(s.SavesDir(), 0o755); err != nil {
		return "", err
	}
	data, err := game.EncodeSnapshot(snap)
	if err != nil {
		return "", err
	}
	path := filepath.Join(s.SavesDir(), sanitiseName(name)+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// LoadGame reads a snapshot back.
func (s *Storage) LoadGame(name string) (game.Snapshot, error) {
	path := filepath.Join(s.SavesDir(), sanitiseName(name)+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return game.Snapshot{}, err
	}
	return game.DecodeSnapshot(data)
}

// ListGames returns the saved game names, newest first.
func (s *Storage) ListGames() []string {
	entries, err := os.ReadDir(s.SavesDir())
	if err != nil {
		return nil
	}

	type row struct {
		name string
		mod  time.Time
	}
	var rows []row
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		rows = append(rows, row{strings.TrimSuffix(e.Name(), ".json"), info.ModTime()})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].mod.After(rows[j].mod) })

	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.name
	}
	return out
}

// DeleteGame removes a saved game.
func (s *Storage) DeleteGame(name string) error {
	return os.Remove(filepath.Join(s.SavesDir(), sanitiseName(name)+".json"))
}

// AutosaveName is the slot used for the automatic save on exit.
const AutosaveName = "autosave"

// sanitiseName keeps a user-supplied save name to something safe to put in a
// path: no separators, no traversal, never empty.
func sanitiseName(name string) string {
	cleaned := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '-', r == '_', r == ' ':
			return r
		default:
			return -1
		}
	}, name)
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return "game"
	}
	if len(cleaned) > 60 {
		cleaned = cleaned[:60]
	}
	return cleaned
}
