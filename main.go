// Command make-noise-bot is a Telegram bot for playing Dozor city quest games.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/meltyshev/make-noise-bot/internal/bot"
	"github.com/meltyshev/make-noise-bot/internal/config"
	"github.com/meltyshev/make-noise-bot/internal/migrations"
	"github.com/meltyshev/make-noise-bot/internal/secret"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/updater"
)

var version = "dev" // set by goreleaser

func main() {
	logger := slog.New(secret.NewLogHandler(slog.NewTextHandler(os.Stderr, nil)))
	if err := run(logger); err != nil {
		logger.Error("bot failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	var (
		configPath  = flag.String("config", "config.json", "path to the config file")
		token       = flag.String("token", "", "bot token; creates the config non-interactively")
		adminID     = flag.Int64("admin-id", 0, "admin user id override")
		debug       = flag.Bool("debug", false, "dump raw engine payloads on parse errors")
		showVersion = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println("make-noise-bot", version)
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig(ctx, *configPath, *token)
	if err != nil {
		return err
	}
	if *adminID != 0 && cfg.AdminID != *adminID {
		cfg.AdminID = *adminID
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
	}
	if *debug {
		cfg.Debug = true
	}

	applied, err := migrations.Apply(cfg.StatePath)
	if err != nil {
		return fmt.Errorf("migrate state: %w", err)
	}
	for _, name := range applied {
		logger.Info("state migrated", "migration", name)
	}

	st, err := store.Open(cfg.StatePath)
	if err != nil {
		return fmt.Errorf("open state: %w", err)
	}

	app, err := bot.New(cfg, st, logger)
	if err != nil {
		return fmt.Errorf("start bot: %w", err)
	}

	go updater.New(st, app.GameEnv(), app.TG(), logger, cfg.UpdateInterval(), app.ReportError).Run(ctx)

	if err := app.Run(ctx); err != nil {
		return err
	}
	logger.Info("bot stopped")
	return nil
}

func loadConfig(ctx context.Context, path, token string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err == nil {
		return cfg, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if token != "" {
		return config.Create(ctx, path, token)
	}
	cfg, err = config.Wizard(ctx, path, os.Stdin, os.Stdout)
	if err != nil {
		return nil, fmt.Errorf("%w - либо запустите с флагом --token", err)
	}
	return cfg, nil
}
