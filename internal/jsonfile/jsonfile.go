// Package jsonfile reads and writes the small JSON files the bot keeps in its
// working directory.
package jsonfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Read unmarshals the file at path into v. A missing file is reported as an
// os.IsNotExist error, so callers can start from defaults.
func Read(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

// Write marshals v indented and replaces path atomically with mode 0600: a
// crash mid-write leaves the previous file untouched rather than a truncated
// one. Both files the bot keeps hold credentials, so neither is world
// readable.
func Write(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	raw = append(raw, '\n')

	// The pattern follows the target, so a stray temp file says which write
	// died. os.CreateTemp already creates it 0600.
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
