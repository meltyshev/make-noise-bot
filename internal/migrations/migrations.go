// Package migrations upgrades the state file to the current schema. Each
// migration lives in its own file and edits the raw JSON, so the store model
// only ever describes the current shape.
package migrations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type migration struct {
	version int
	name    string
	apply   func(state map[string]any) error
}

// all is filled by the init of every migration file.
var all []migration

func register(version int, name string, apply func(state map[string]any) error) {
	all = append(all, migration{version: version, name: name, apply: apply})
}

// Current is the schema version this build expects.
func Current() int {
	current := 0
	for _, m := range all {
		if m.version > current {
			current = m.version
		}
	}
	return current
}

// Apply brings the state file at path up to date and reports the names of
// the migrations it ran. A state that does not exist yet is stamped with the
// current version, so later migrations skip it.
func Apply(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, write(path, map[string]any{"schema_version": Current()})
	}
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}

	var state map[string]any
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("parse state %s: %w", path, err)
	}

	pending := append([]migration{}, all...)
	sort.Slice(pending, func(i, j int) bool { return pending[i].version < pending[j].version })

	current := schemaVersion(state)
	var applied []string
	for _, m := range pending {
		if m.version <= current {
			continue
		}
		if err := m.apply(state); err != nil {
			return nil, fmt.Errorf("migration %d %s: %w", m.version, m.name, err)
		}
		state["schema_version"] = m.version
		applied = append(applied, m.name)
	}
	if len(applied) == 0 {
		return nil, nil
	}

	if err := write(path, state); err != nil {
		return nil, err
	}
	return applied, nil
}

func schemaVersion(state map[string]any) int {
	version, ok := state["schema_version"].(float64)
	if !ok {
		return 0
	}
	return int(version)
}

func write(path string, state map[string]any) error {
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	raw = append(raw, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(path), ".state-*.json")
	if err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("write state: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}

// object returns a nested JSON object for editing in place.
func object(state map[string]any, key string) (map[string]any, bool) {
	value, ok := state[key].(map[string]any)
	return value, ok
}
