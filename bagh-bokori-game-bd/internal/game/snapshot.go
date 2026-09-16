package game

import "encoding/json"

const snapshotVersion = 1

// Snapshot is the serialisable form of a whole game.
//
// Only the ruleset and the move list are stored: replaying the moves through
// the rules reconstructs every intermediate state exactly, which keeps the
// format small and makes a corrupt or forged file impossible to load into an
// illegal position.
type Snapshot struct {
	Version int      `json:"version"`
	Rules   RuleSet  `json:"rules"`
	Moves   []Move   `json:"moves"`
	Cursor  int      `json:"cursor"`
	Label   string   `json:"label,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// MarshalJSON is provided on Snapshot so callers get a stable format without
// the engine ever touching a filesystem - the caller decides where bytes go,
// which keeps the package WebAssembly-safe.
func (s Snapshot) MarshalJSON() ([]byte, error) {
	type alias Snapshot
	return json.Marshal(alias(s))
}

// EncodeSnapshot serialises a snapshot to indented JSON.
func EncodeSnapshot(s Snapshot) ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// DecodeSnapshot parses JSON produced by EncodeSnapshot.
func DecodeSnapshot(data []byte) (Snapshot, error) {
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return Snapshot{}, err
	}
	if s.Rules.TotalGoats <= 0 {
		s.Rules = DefaultRuleSet()
	}
	return s, nil
}
