package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"bagh-bakri/internal/game"
)

const saveDir = ".bagh-bakri"
const saveFile = "save.json"

type SavedGame struct {
	State      game.GameState `json:"state"`
	PlayerTurn Player         `json:"player_turn"`
}

type Player int

const (
	PlayerHuman Player = iota
	PlayerAI
)

func SaveGame(state game.GameState, player Player) error {
	dir, err := savePath()
	if err != nil {
		return err
	}

	saved := SavedGame{
		State:      state,
		PlayerTurn: player,
	}

	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save: %w", err)
	}

	path := filepath.Join(dir, saveFile)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write save: %w", err)
	}

	return nil
}

func LoadGame() (game.GameState, Player, error) {
	dir, err := savePath()
	if err != nil {
		return game.GameState{}, PlayerHuman, err
	}

	path := filepath.Join(dir, saveFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return game.GameState{}, PlayerHuman, fmt.Errorf("no saved game found")
	}

	var saved SavedGame
	if err := json.Unmarshal(data, &saved); err != nil {
		return game.GameState{}, PlayerHuman, fmt.Errorf("failed to parse save: %w", err)
	}

	return saved.State, saved.PlayerTurn, nil
}

func HasSaveGame() bool {
	dir, err := savePath()
	if err != nil {
		return false
	}
	path := filepath.Join(dir, saveFile)
	_, err = os.Stat(path)
	return err == nil
}

func DeleteSave() error {
	dir, err := savePath()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, saveFile)
	return os.Remove(path)
}

func savePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, saveDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}
