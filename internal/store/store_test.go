package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "state.json")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return s, path
}

func TestOpenMissingFileGivesDefaults(t *testing.T) {
	s, path := openTemp(t)

	cfg := s.GameConfig()
	if cfg.Engine != "DozorClassic" || cfg.City != "e-burg" {
		t.Errorf("GameConfig() = %+v, want engine DozorClassic in e-burg", cfg)
	}
	if len(cfg.CodeFormats) != 1 || cfg.CodeFormats[0][0] != "dr" {
		t.Errorf("GameConfig().CodeFormats = %v, want one group starting with dr", cfg.CodeFormats)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Open(%s) created the file, want it written only on the first update", path)
	}
}

func TestUpdatePersistsAndReloads(t *testing.T) {
	s, path := openTemp(t)

	err := s.Update(func(d *Data) {
		d.Managers = []int64{7}
		d.Chats[42] = &Chat{ID: 42, Type: "group", Permission: PermissionAllowed, Title: "Team"}
		level := 3
		d.Game = &Game{Engine: "DozorClassic", City: "e-burg", LevelNumber: &level, Subscriptions: Subscriptions{AllUpdates(42)}}
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	if !reopened.IsManager(7) || reopened.IsManager(8) {
		t.Errorf("IsManager(7), IsManager(8) = %v, %v, want true, false", reopened.IsManager(7), reopened.IsManager(8))
	}
	chat, ok := reopened.Chat(42)
	if !ok || chat.Permission != PermissionAllowed || chat.Title != "Team" {
		t.Errorf("Chat(42) = (%+v, %v), want the allowed chat \"Team\"", chat, ok)
	}
	game, ok := reopened.Game()
	sub, subscribed := game.Subscriptions.Find(42)
	if !ok || game.LevelNumber == nil || *game.LevelNumber != 3 || !subscribed || sub.EventsOnly {
		t.Errorf("Game() = (%+v, %v), want level 3 with chat 42 on all updates", game, ok)
	}
}

// TestGameReturnsCopies covers every field a caller can reach through: the
// slices, and each *int, which a shallow struct copy would share.
func TestGameReturnsCopies(t *testing.T) {
	s, _ := openTemp(t)
	level, levelTime, hint, pinned := 3, 40, 2, 7
	if err := s.Update(func(d *Data) {
		d.Game = &Game{
			Engine:         "DozorLite",
			CodeFormats:    [][]string{{"dr", "др"}},
			Subscriptions:  Subscriptions{AllUpdates(1)},
			SolvedSpoilers: []int{1},
			LevelNumber:    &level,
			LevelTime:      &levelTime,
			HintNumber:     &hint,
			PinnedLevel:    &pinned,
		}
	}); err != nil {
		t.Fatal(err)
	}

	game, _ := s.Game()
	game.Engine = "hacked"
	game.Subscriptions[0].ChatID = 999
	game.CodeFormats[0][0] = "hacked"
	game.SolvedSpoilers[0] = 999
	for _, p := range []*int{game.LevelNumber, game.LevelTime, game.HintNumber, game.PinnedLevel} {
		*p = 999
	}

	fresh, _ := s.Game()
	for _, tt := range []struct {
		name string
		got  any
		want any
	}{
		{name: "engine", got: fresh.Engine, want: "DozorLite"},
		{name: "subscription chat id", got: fresh.Subscriptions[0].ChatID, want: int64(1)},
		{name: "code format", got: fresh.CodeFormats[0][0], want: "dr"},
		{name: "solved spoiler", got: fresh.SolvedSpoilers[0], want: 1},
		{name: "level number", got: *fresh.LevelNumber, want: 3},
		{name: "level time", got: *fresh.LevelTime, want: 40},
		{name: "hint number", got: *fresh.HintNumber, want: 2},
		{name: "pinned level", got: *fresh.PinnedLevel, want: 7},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("Game().%s = %v after mutating the copy %+v, want %v", tt.name, tt.got, game, tt.want)
			}
		})
	}
}

func TestRating(t *testing.T) {
	s, _ := openTemp(t)

	for range 3 {
		if err := s.IncrementPlayer(1, "Один"); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.IncrementPlayer(2, "Два"); err != nil {
		t.Fatal(err)
	}

	rating := s.Rating()
	if len(rating) != 2 {
		t.Fatalf("Rating() = %d players, want 2", len(rating))
	}
	if rating[0].Name != "Один" || rating[0].Total != 3 || rating[1].Name != "Два" {
		t.Errorf("Rating() = %+v, want Один with 3 ahead of Два", rating)
	}
}

// TestEmptyListsMarshalAsArrays pins the /gameconfig display: empty lists
// must render as [], not null (nil slices marshal as JSON null in Go).
func TestEmptyListsMarshalAsArrays(t *testing.T) {
	s, _ := openTemp(t)
	if err := s.Update(func(d *Data) {
		d.GameConfig.Subscriptions = Subscriptions{}
		d.GameConfig.CodeFormats = [][]string{}
		d.Game = &Game{Engine: "DozorClassic"}
	}); err != nil {
		t.Fatal(err)
	}

	cfg := s.GameConfig()
	for name, value := range map[string]any{
		"subscriptions": cfg.Subscriptions,
		"code_formats":  cfg.CodeFormats,
	} {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != "[]" {
			t.Errorf("%s marshals as %s, want []", name, raw)
		}
	}

	// Copies of a game loaded from a hand-edited state with null lists must
	// also come out as [].
	game, _ := s.Game()
	if game.Subscriptions == nil || game.SolvedSpoilers == nil {
		t.Errorf("Game() lists = (%v, %v), want both non-nil so they marshal as []", game.Subscriptions, game.SolvedSpoilers)
	}
}

func TestOpenToleratesNulledCollections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte(`{"managers": null, "chats": null, "players": null}`), 0o600); err != nil {
		t.Fatal(err)
	}

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Update(func(d *Data) { d.Chats[1] = &Chat{ID: 1} }); err != nil {
		t.Fatalf("Update on nulled maps: %v", err)
	}
}
