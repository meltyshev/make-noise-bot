// Package migrations upgrades the state file to the current schema. Each
// migration lives in its own file and edits the raw JSON, so the store model
// only ever describes the current shape.
package migrations

import (
	"cmp"
	"fmt"
	"os"
	"slices"

	"github.com/meltyshev/make-noise-bot/internal/jsonfile"
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
	var state map[string]any
	err := jsonfile.Read(path, &state)
	if os.IsNotExist(err) {
		return nil, jsonfile.Write(path, map[string]any{"schema_version": Current()})
	}
	if err != nil {
		return nil, err
	}

	pending := slices.Clone(all)
	slices.SortFunc(pending, func(a, b migration) int { return cmp.Compare(a.version, b.version) })

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

	if err := jsonfile.Write(path, state); err != nil {
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

// object returns a nested JSON object for editing in place.
func object(state map[string]any, key string) (map[string]any, bool) {
	value, ok := state[key].(map[string]any)
	return value, ok
}
