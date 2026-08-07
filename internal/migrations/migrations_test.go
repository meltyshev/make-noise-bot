package migrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeState(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func read(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func TestApplyMissingFileStampsCurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	applied, err := Apply(path)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("applied = %v, want none", applied)
	}
	if got := read(t, path)["schema_version"]; got != float64(Current()) {
		t.Errorf("schema_version = %v, want %d", got, Current())
	}
}

func TestApplySubscriptionKinds(t *testing.T) {
	path := writeState(t, `{
	  "game_config": {"city": "e-burg", "subscribers": [-100, -200, 7]},
	  "game": {"engine": "DozorClassic", "subscribers": [-100, -200]}
	}`)

	applied, err := Apply(path)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(applied) != 1 {
		t.Fatalf("applied = %v, want one migration", applied)
	}

	state := read(t, path)
	if got := state["schema_version"]; got != float64(1) {
		t.Errorf("schema_version = %v", got)
	}

	for _, key := range []string{"game_config", "game"} {
		owner := state[key].(map[string]any)
		if _, ok := owner["subscribers"]; ok {
			t.Errorf("%s: legacy subscribers kept", key)
		}

		subs, ok := owner["subscriptions"].([]any)
		if !ok {
			t.Fatalf("%s: subscriptions missing", key)
		}

		first := subs[0].(map[string]any)
		if first["chat_id"] != float64(-100) || first["question"] != true || first["notes"] != true {
			t.Errorf("%s: first subscriber = %v, want everything", key, first)
		}

		second := subs[1].(map[string]any)
		if second["question"] != false || second["notes"] != false {
			t.Errorf("%s: second subscriber = %v, want notifications only", key, second)
		}
		for _, kind := range []string{"level_up", "hints", "spoilers"} {
			if second[kind] != true {
				t.Errorf("%s: second subscriber lost %s", key, kind)
			}
		}
	}

	// Untouched fields survive.
	if city := state["game_config"].(map[string]any)["city"]; city != "e-burg" {
		t.Errorf("city = %v", city)
	}
}

func TestApplyIsIdempotent(t *testing.T) {
	path := writeState(t, `{"game_config": {"subscribers": [1]}}`)

	if _, err := Apply(path); err != nil {
		t.Fatal(err)
	}
	before := read(t, path)

	applied, err := Apply(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 0 {
		t.Errorf("second run applied %v", applied)
	}

	after := read(t, path)
	beforeRaw, _ := json.Marshal(before)
	afterRaw, _ := json.Marshal(after)
	if string(beforeRaw) != string(afterRaw) {
		t.Errorf("state changed on rerun:\n%s\n%s", beforeRaw, afterRaw)
	}
}

func TestApplyWithoutSubscribers(t *testing.T) {
	path := writeState(t, `{"game_config": {"city": "msk"}, "players": {}}`)

	if _, err := Apply(path); err != nil {
		t.Fatal(err)
	}

	state := read(t, path)
	owner := state["game_config"].(map[string]any)
	if subs, ok := owner["subscriptions"]; ok {
		t.Errorf("subscriptions invented: %v", subs)
	}
	if state["schema_version"] != float64(1) {
		t.Errorf("schema_version = %v", state["schema_version"])
	}
}

func TestApplyBrokenState(t *testing.T) {
	path := writeState(t, "{not json")
	if _, err := Apply(path); err == nil {
		t.Error("broken state must fail loudly")
	}
}
