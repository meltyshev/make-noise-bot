package jsonfile

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteIsPrivate pins the mode both callers rely on: config.json holds the
// bot token and state.json holds the game credentials.
func TestWriteIsPrivate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	for _, tt := range []struct {
		name  string
		value map[string]any
	}{
		{name: "first write", value: map[string]any{"token": "t"}},
		{name: "replacing an existing file", value: map[string]any{"token": "u"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := Write(path, tt.value); err != nil {
				t.Fatalf("Write(%s) = %v, want no error", path, err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if got := info.Mode().Perm(); got != 0o600 {
				t.Errorf("Write(%s) mode = %v, want 0600", path, got)
			}
		})
	}

	// Two writes must leave one file: the rename replaces the target and the
	// temp is cleaned up either way.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 1 || names[0] != "config.json" {
		t.Errorf("directory after two writes = %v, want just config.json", names)
	}
}

func TestWriteThenRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := Write(path, map[string]any{"schema_version": 2}); err != nil {
		t.Fatalf("Write(%s) = %v, want no error", path, err)
	}

	var got map[string]any
	if err := Read(path, &got); err != nil {
		t.Fatalf("Read(%s) = %v, want no error", path, err)
	}
	if got["schema_version"] != float64(2) {
		t.Errorf("Read(%s)[schema_version] = %v, want 2", path, got["schema_version"])
	}
}

func TestReadMissingIsNotExist(t *testing.T) {
	var got map[string]any
	err := Read(filepath.Join(t.TempDir(), "nope.json"), &got)
	if !os.IsNotExist(err) {
		t.Errorf("Read(missing) = %v, want IsNotExist", err)
	}
}
