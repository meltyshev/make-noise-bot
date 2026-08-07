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

// A state from the plain-list era upgrades all the way to the current shape.
func TestApplyLegacySubscribers(t *testing.T) {
	path := writeState(t, `{
	  "game_config": {"city": "e-burg", "subscribers": [-100, -200, 7]},
	  "game": {"engine": "DozorClassic", "subscribers": [-100, -200]}
	}`)

	applied, err := Apply(path)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(applied) != 2 {
		t.Fatalf("applied = %v, want both migrations", applied)
	}

	state := read(t, path)
	if got := state["schema_version"]; got != float64(Current()) {
		t.Errorf("schema_version = %v, want %d", got, Current())
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

		// The chat that used to receive the level texts keeps everything.
		first := subs[0].(map[string]any)
		if first["chat_id"] != float64(-100) || first["events_only"] == true {
			t.Errorf("%s: first subscriber = %v, want everything", key, first)
		}
		second := subs[1].(map[string]any)
		if second["events_only"] != true {
			t.Errorf("%s: second subscriber = %v, want events only", key, second)
		}
		for _, sub := range subs {
			for _, kind := range []string{"level_up", "hints", "spoilers", "question", "notes"} {
				if _, ok := sub.(map[string]any)[kind]; ok {
					t.Errorf("%s: %s survived", key, kind)
				}
			}
		}
	}

	if city := state["game_config"].(map[string]any)["city"]; city != "e-burg" {
		t.Errorf("city = %v", city)
	}
}

func TestApplySubscriptionModes(t *testing.T) {
	path := writeState(t, `{
	  "schema_version": 1,
	  "game_config": {"subscriptions": [
	    {"chat_id": -100, "level_up": true, "hints": true, "spoilers": true, "question": true, "notes": true},
	    {"chat_id": -200, "level_up": true, "hints": true, "spoilers": true},
	    {"chat_id": 7, "notes": true}
	  ]}
	}`)

	if _, err := Apply(path); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	subs := read(t, path)["game_config"].(map[string]any)["subscriptions"].([]any)
	if len(subs) != 3 {
		t.Fatalf("subscriptions = %d, want 3", len(subs))
	}

	// Chats that wanted the level texts keep everything.
	for i, want := range []bool{false, true, false} {
		sub := subs[i].(map[string]any)
		if got := sub["events_only"] == true; got != want {
			t.Errorf("subscription %d events_only = %v, want %v", i, got, want)
		}
		for _, kind := range []string{"level_up", "hints", "spoilers", "question", "notes"} {
			if _, ok := sub[kind]; ok {
				t.Errorf("subscription %d kept %s", i, kind)
			}
		}
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
	if state["schema_version"] != float64(Current()) {
		t.Errorf("schema_version = %v", state["schema_version"])
	}
}

func TestApplyBrokenState(t *testing.T) {
	path := writeState(t, "{not json")
	if _, err := Apply(path); err == nil {
		t.Error("broken state must fail loudly")
	}
}
