package config

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAppliesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"token": "t", "admin_id": 5}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.UpdateIntervalSeconds != DefaultUpdateInterval {
		t.Errorf("UpdateIntervalSeconds = %d, want %d", cfg.UpdateIntervalSeconds, DefaultUpdateInterval)
	}
	if cfg.StatePath != filepath.Join(filepath.Dir(path), "state.json") {
		t.Errorf("StatePath = %q, want it next to the config", cfg.StatePath)
	}
}

func TestStripURLDropsToken(t *testing.T) {
	const token = "123456789:AAHsecretvaluehere"
	err := &url.Error{
		Op:  "Get",
		URL: "https://api.telegram.org/bot" + token + "/getMe",
		Err: errors.New("context deadline exceeded"),
	}

	got := stripURL(err).Error()
	if strings.Contains(got, token) {
		t.Errorf("token survived: %q", got)
	}
	if !strings.Contains(got, "context deadline exceeded") {
		t.Errorf("cause lost: %q", got)
	}
}

func TestStripURLKeepsOtherErrors(t *testing.T) {
	err := errors.New("Unauthorized")
	if got := stripURL(err); !errors.Is(got, err) {
		t.Errorf("plain error changed: %v", got)
	}
}

func TestLoadMissingIsNotExist(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if !os.IsNotExist(err) {
		t.Errorf("want IsNotExist, got %v", err)
	}
}

func TestSaveRoundtripAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := &Config{Token: "secret-token", AdminID: 7, UpdateIntervalSeconds: 9, path: path}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config permissions = %v, want 0600: it holds the token", info.Mode().Perm())
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if loaded.Token != "secret-token" || loaded.AdminID != 7 || loaded.UpdateIntervalSeconds != 9 {
		t.Errorf("roundtrip = %+v", loaded)
	}
}
