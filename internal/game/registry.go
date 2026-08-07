package game

import (
	"context"
	"fmt"

	"github.com/meltyshev/make-noise-bot/internal/store"
)

// Names is in game config menu order.
var Names = []string{NameClassic, NameLite, NameClassicPrequel, NameLitePrequel}

func IsKnownEngine(name string) bool {
	for _, known := range Names {
		if known == name {
			return true
		}
	}
	return false
}

func newGameFromConfig(cfg store.GameConfig) *store.Game {
	game := &store.Game{
		Engine:        cfg.Engine,
		City:          cfg.City,
		CodeFormats:   make([][]string, len(cfg.CodeFormats)),
		Subscriptions: cfg.Subscriptions.Clone(),
	}
	for i, format := range cfg.CodeFormats {
		game.CodeFormats[i] = append([]string{}, format...)
	}
	return game
}

// Start prepares a new game from the config; the caller persists it.
func Start(ctx context.Context, cfg store.GameConfig, env *Env) (*store.Game, error) {
	switch cfg.Engine {
	case NameClassic:
		return startClassic(ctx, cfg, env)
	case NameLite:
		return startLite(cfg), nil
	case NameClassicPrequel, NameLitePrequel:
		return startPrequel(ctx, cfg.Engine, cfg, env)
	default:
		return nil, fmt.Errorf("unknown engine %q", cfg.Engine)
	}
}

// New returns nil for unknown engine names.
func New(g store.Game, env *Env) Engine {
	switch g.Engine {
	case NameClassic:
		return newClassic(g, env)
	case NameLite:
		return newLite(g, env)
	case NameClassicPrequel, NameLitePrequel:
		return newPrequel(g, env)
	default:
		return nil
	}
}
