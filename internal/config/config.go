// Package config handles config.json and the interactive first-run setup.
package config

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/meltyshev/make-noise-bot/internal/jsonfile"
	"github.com/meltyshev/make-noise-bot/internal/secret"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

const DefaultUpdateInterval = 5

type Config struct {
	Token                 string `json:"token"`
	AdminID               int64  `json:"admin_id"`
	UpdateIntervalSeconds int    `json:"update_interval_seconds"`
	StatePath             string `json:"state_path,omitempty"`
	Debug                 bool   `json:"debug,omitempty"`

	path string
}

// Load reads the config; a missing file is reported via os.IsNotExist.
func Load(path string) (*Config, error) {
	cfg := &Config{path: path}
	if err := jsonfile.Read(path, cfg); err != nil {
		return nil, err
	}
	cfg.applyDefaults()
	secret.Register(cfg.Token)
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.UpdateIntervalSeconds <= 0 {
		c.UpdateIntervalSeconds = DefaultUpdateInterval
	}
	if c.StatePath == "" {
		c.StatePath = filepath.Join(filepath.Dir(c.path), "state.json")
	}
}

// Save replaces the config atomically; jsonfile keeps it at mode 0600, which
// matters because it holds the token.
func (c *Config) Save() error {
	return jsonfile.Write(c.path, c)
}

func (c *Config) UpdateInterval() time.Duration {
	return time.Duration(c.UpdateIntervalSeconds) * time.Second
}

// Create writes a fresh config non-interactively.
func Create(ctx context.Context, path, token string) (*Config, error) {
	if _, err := checkToken(ctx, token); err != nil {
		return nil, fmt.Errorf("check token: %w", err)
	}
	cfg := &Config{
		Token:                 token,
		UpdateIntervalSeconds: DefaultUpdateInterval,
		path:                  path,
	}
	cfg.applyDefaults()
	secret.Register(cfg.Token)
	if err := cfg.Save(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Wizard interactively asks for a token, validates it and saves the config.
func Wizard(ctx context.Context, path string, in io.Reader, out io.Writer) (*Config, error) {
	fmt.Fprintln(out, texts.WizardNoConfig)
	fmt.Fprintln(out, texts.WizardCreateBot)
	fmt.Fprintln(out)

	reader := bufio.NewReader(in)
	for {
		fmt.Fprint(out, texts.WizardAskToken)

		line, err := reader.ReadString('\n')
		token := strings.TrimSpace(line)
		if token == "" {
			if err != nil {
				return nil, errors.New(texts.WizardNoToken)
			}
			continue
		}

		username, checkErr := checkToken(ctx, token)
		if checkErr != nil {
			fmt.Fprintf(out, texts.WizardBadToken, checkErr)
			if err != nil {
				return nil, errors.New(texts.WizardNoGoodToken)
			}
			continue
		}

		cfg := &Config{
			Token:                 token,
			UpdateIntervalSeconds: DefaultUpdateInterval,
			path:                  path,
		}
		cfg.applyDefaults()
		secret.Register(cfg.Token)
		if err := cfg.Save(); err != nil {
			return nil, err
		}

		fmt.Fprintf(out, texts.WizardSavedFmt, username, path)
		fmt.Fprint(out, texts.WizardClaimAdmin)
		return cfg, nil
	}
}

func checkToken(ctx context.Context, token string) (string, error) {
	link := "https://api.telegram.org/bot" + url.PathEscape(token) + "/getMe"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return "", secret.StripURL(err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", secret.StripURL(err)
	}
	defer resp.Body.Close()

	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if !payload.OK {
		return "", errors.New(payload.Description)
	}
	return payload.Result.Username, nil
}
