// Package bot wires the Telegram side of make-noise-bot: dispatching,
// conversations, permissions and all commands.
package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/config"
	"github.com/meltyshev/make-noise-bot/internal/game"
	"github.com/meltyshev/make-noise-bot/internal/secret"
	"github.com/meltyshev/make-noise-bot/internal/store"
)

// Command is one bot command. Init receives the command with arguments;
// Handle receives the next message while a conversation is open.
type Command struct {
	Name        string
	Description string // empty = hidden from /help and the command menu
	Init        func(c *Ctx, args string)
	Handle      func(c *Ctx, state any)
}

type App struct {
	cfg   *config.Config
	store *store.Store
	tg    *tgbot.Bot
	me    *models.User
	env   *game.Env
	conv  *convStore
	log   *slog.Logger

	commands map[string]*Command
	order    []*Command

	admin       atomic.Int64
	lastErrorDM atomic.Int64 // unix seconds of the last error DM, for rate limiting
	configMu    sync.Mutex
}

func New(cfg *config.Config, st *store.Store, logger *slog.Logger) (*App, error) {
	a := &App{
		cfg:      cfg,
		store:    st,
		conv:     newConvStore(),
		log:      logger,
		commands: map[string]*Command{},
	}
	a.admin.Store(cfg.AdminID)

	a.env = game.DefaultEnv()
	a.env.OnSessionUpdate = func(session string) {
		if err := st.UpdateGame(func(g *store.Game) { g.Session = session }); err != nil {
			a.log.Error("persist session failed", "error", err)
		}
		a.log.Info("engine session refreshed")
	}
	if cfg.Debug {
		a.env.Debug = a.debugDump
	}

	tg, err := tgbot.New(
		cfg.Token,
		tgbot.WithDefaultHandler(a.onUpdate),
		// Handle updates strictly in order: codes must not race each other.
		tgbot.WithNotAsyncHandlers(),
		tgbot.WithErrorsHandler(func(err error) {
			a.log.Error("telegram client failed", "error", err)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create telegram client: %w", err)
	}
	a.tg = tg

	a.registerCommands()
	return a, nil
}

func (a *App) adminID() int64 { return a.admin.Load() }

func (a *App) TG() *tgbot.Bot        { return a.tg }
func (a *App) GameEnv() *game.Env    { return a.env }
func (a *App) ReportError(err error) { a.reportError(err) }

func (a *App) Run(ctx context.Context) error {
	me, err := a.tg.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("identify bot: %w", err)
	}
	a.me = me

	if err := a.publishCommandMenu(ctx); err != nil {
		a.log.Warn("publish command menu failed", "error", err)
	}

	a.conv.StartJanitor(ctx)

	a.log.Info("bot started", "username", me.Username)
	if a.adminID() == 0 {
		a.log.Info("no admin configured, send /start to the bot to become admin")
	}

	a.tg.Start(ctx)
	return nil
}

func (a *App) Engine() game.Engine {
	g, ok := a.store.Game()
	if !ok {
		return nil
	}
	return game.New(g, a.env)
}

func (a *App) ClassicEngine() game.Engine {
	g, ok := a.store.Game()
	if !ok || g.Engine != game.NameClassic {
		return nil
	}
	return game.New(g, a.env)
}

func (a *App) publishCommandMenu(ctx context.Context) error {
	var commands []models.BotCommand
	for _, cmd := range a.order {
		if cmd.Description != "" {
			commands = append(commands, models.BotCommand{Command: cmd.Name, Description: cmd.Description})
		}
	}
	_, err := a.tg.SetMyCommands(ctx, &tgbot.SetMyCommandsParams{Commands: commands})
	return err
}

func (a *App) register(cmd *Command) {
	a.commands[cmd.Name] = cmd
	a.order = append(a.order, cmd)
}

func (a *App) claimAdmin(userID int64) error {
	a.configMu.Lock()
	defer a.configMu.Unlock()

	a.cfg.AdminID = userID
	if err := a.cfg.Save(); err != nil {
		return err
	}
	a.admin.Store(userID)
	a.log.Info("admin claimed", "user_id", userID)
	return nil
}

// reportError DMs the admin at most once per 30 seconds.
func (a *App) reportError(err error) {
	if err == nil {
		return
	}
	a.log.Error("handler error", "error", err)

	adminID := a.adminID()
	if adminID == 0 {
		return
	}

	now := time.Now().Unix()
	last := a.lastErrorDM.Load()
	if now-last < 30 || !a.lastErrorDM.CompareAndSwap(last, now) {
		return
	}

	text := secret.Redact(err.Error())
	if len(text) > 3500 {
		text = text[:3500] + "..."
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, sendErr := a.tg.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: adminID, Text: "⚠️ " + text}); sendErr != nil {
		a.log.Error("error DM failed", "error", sendErr)
	}
}

// debugDump saves a raw engine payload next to the state file (--debug).
func (a *App) debugDump(kind string, body []byte) {
	dir := filepath.Join(filepath.Dir(a.cfg.StatePath), "debug")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		a.log.Warn("debug dump failed", "error", err)
		return
	}
	name := filepath.Join(dir, fmt.Sprintf("%s-%s.txt", time.Now().Format("20060102-150405"), kind))
	if err := os.WriteFile(name, body, 0o644); err != nil {
		a.log.Warn("debug dump failed", "error", err)
		return
	}
	a.log.Info("engine payload dumped", "file", name)
}

func (a *App) recoverPanic() {
	if r := recover(); r != nil {
		a.reportError(fmt.Errorf("panic: %v\n%s", r, debug.Stack()))
	}
}
