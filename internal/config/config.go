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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/meltyshev/make-noise-bot/internal/secret"
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
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{path: path}
	if err := json.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
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

// Save writes the config with mode 0600: it contains the token.
func (c *Config) Save() error {
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, append(raw, '\n'), 0o600)
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
	fmt.Fprintln(out, "Файл конфигурации не найден - настроим бота.")
	fmt.Fprintln(out, "Создайте бота у @BotFather в Telegram и получите токен.")
	fmt.Fprintln(out)

	reader := bufio.NewReader(in)
	for {
		fmt.Fprint(out, "Вставьте токен бота: ")

		line, err := reader.ReadString('\n')
		token := strings.TrimSpace(line)
		if token == "" {
			if err != nil {
				return nil, errors.New("не введен токен")
			}
			continue
		}

		username, checkErr := checkToken(ctx, token)
		if checkErr != nil {
			fmt.Fprintf(out, "Токен не подошел (%v), попробуйте еще раз.\n", checkErr)
			if err != nil {
				return nil, errors.New("не введен рабочий токен")
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

		fmt.Fprintf(out, "Готово, это @%s. Конфигурация сохранена в %s.\n", username, path)
		fmt.Fprintf(out, "Теперь отправьте боту /start - первый написавший станет админом.\n\n")
		return cfg, nil
	}
}

func checkToken(ctx context.Context, token string) (string, error) {
	link := "https://api.telegram.org/bot" + url.PathEscape(token) + "/getMe"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return "", stripURL(err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", stripURL(err)
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

// stripURL drops the request URL from a transport error: it carries the token.
func stripURL(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("%s: %w", urlErr.Op, urlErr.Err)
	}
	return err
}
