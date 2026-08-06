// make-noise-bot is a Telegram bot for playing Dozor city quest games.
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
	"github.com/meltyshev/make-noise-bot/internal/secret"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/updater"
)

var version = "dev" // set by goreleaser

func main() {
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
		return
	}

	logger := slog.New(secret.NewLogHandler(slog.NewTextHandler(os.Stderr, nil)))

	cfg, err := loadConfig(*configPath, *token)
	if err != nil {
		logger.Error("config", "error", err)
		os.Exit(1)
	}
	if *adminID != 0 && cfg.AdminID != *adminID {
		cfg.AdminID = *adminID
		if err := cfg.Save(); err != nil {
			logger.Error("config", "error", err)
			os.Exit(1)
		}
	}
	if *debug {
		cfg.Debug = true
	}

	st, err := store.Open(cfg.StatePath)
	if err != nil {
		logger.Error("state", "error", err)
		os.Exit(1)
	}

	app, err := bot.New(cfg, st, logger)
	if err != nil {
		logger.Error("startup", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go updater.New(st, app.GameEnv(), app.TG(), logger, cfg.UpdateInterval(), app.ReportError).Run(ctx)

	if err := app.Run(ctx); err != nil {
		logger.Error("run", "error", err)
		os.Exit(1)
	}
	logger.Info("bot stopped")
}

func loadConfig(path, token string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err == nil {
		return cfg, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	if token != "" {
		return config.Create(path, token)
	}
	cfg, err = config.Wizard(path, os.Stdin, os.Stdout)
	if err != nil {
		return nil, fmt.Errorf("%w - либо запустите с флагом --token", err)
	}
	return cfg, nil
}
