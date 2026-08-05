// Package updater polls the game engine and broadcasts level-ups, hints and
// solved spoilers to subscribed chats.
package updater

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/game"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

type Updater struct {
	store    *store.Store
	env      *game.Env
	tg       *tgbot.Bot
	log      *slog.Logger
	interval time.Duration
	report   func(error)
}

func New(st *store.Store, env *game.Env, tg *tgbot.Bot, logger *slog.Logger, interval time.Duration, report func(error)) *Updater {
	return &Updater{
		store:    st,
		env:      env,
		tg:       tg,
		log:      logger,
		interval: interval,
		report:   report,
	}
}

// Run polls until ctx is cancelled; ticks never overlap.
func (u *Updater) Run(ctx context.Context) {
	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.tick(ctx)
		}
	}
}

func (u *Updater) tick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			u.report(fmt.Errorf("updater panic: %v\n%s", r, debug.Stack()))
		}
	}()

	g, ok := u.store.Game()
	if !ok || len(g.Subscribers) == 0 {
		return
	}

	engine := game.New(g, u.env)
	if engine == nil {
		return
	}

	snap, err := engine.Load(ctx)
	if err != nil {
		u.log.Warn("engine load failed", "error", err)
		return
	}

	// Persist first, notify after: a failed broadcast is never repeated.
	levelNumber := snap.LevelNumber()
	currentHint := g.HintNumber
	currentSolved := g.SolvedSpoilers

	if !intPtrEqual(levelNumber, g.LevelNumber) {
		err := u.store.UpdateGame(func(g *store.Game) {
			g.LevelNumber = levelNumber
			g.HintNumber = nil
			g.SolvedSpoilers = nil
		})
		if err != nil {
			u.report(err)
			return
		}
		currentHint = nil
		currentSolved = nil

		if levelNumber != nil {
			u.broadcast(ctx, texts.LevelUp, false)

			if question := snap.Question(); question != "" {
				u.sendToFirst(ctx, g.Subscribers, question)
			}
			if notes := snap.Notes(); notes != "" {
				u.sendToFirst(ctx, g.Subscribers, notes)
			}
		}
	}

	hintNumber, hintText := snap.Hint()
	var hintPtr *int
	if hintNumber != 0 {
		hintPtr = &hintNumber
	}
	if !intPtrEqual(hintPtr, currentHint) {
		if err := u.store.UpdateGame(func(g *store.Game) { g.HintNumber = hintPtr }); err != nil {
			u.report(err)
			return
		}
		if hintPtr != nil {
			u.broadcastHTML(ctx, fmt.Sprintf(texts.HintFmt, hintNumber, hintText))
		}
	}

	solved := snap.SolvedSpoilers()
	newSolved := diff(solved, currentSolved)
	if len(newSolved) > 0 {
		if err := u.store.UpdateGame(func(g *store.Game) { g.SolvedSpoilers = solved }); err != nil {
			u.report(err)
			return
		}
		for _, spoiler := range newSolved {
			u.broadcast(ctx, fmt.Sprintf(texts.SpoilerSolved, spoiler), false)
		}
	}
}

// A chat that blocked the bot is unsubscribed automatically.
func (u *Updater) broadcast(ctx context.Context, text string, html bool) {
	g, ok := u.store.Game()
	if !ok {
		return
	}
	for _, subscriber := range g.Subscribers {
		u.sendTo(ctx, subscriber, text, html)
	}
}

func (u *Updater) broadcastHTML(ctx context.Context, text string) {
	u.broadcast(ctx, text, true)
}

// Questions and notes go only to the first subscriber, the team's main chat.
func (u *Updater) sendToFirst(ctx context.Context, subscribers []int64, text string) {
	if len(subscribers) == 0 {
		return
	}
	u.sendTo(ctx, subscribers[0], text, true)
}

func (u *Updater) sendTo(ctx context.Context, chatID int64, text string, html bool) {
	params := &tgbot.SendMessageParams{ChatID: chatID, Text: text}
	if html {
		params.ParseMode = models.ParseModeHTML
	}

	if _, err := u.tg.SendMessage(ctx, params); err != nil {
		if errors.Is(err, tgbot.ErrorForbidden) {
			u.unsubscribe(chatID)
			return
		}
		u.log.Warn("broadcast failed", "chat_id", chatID, "error", err)
	}
}

func (u *Updater) unsubscribe(chatID int64) {
	err := u.store.UpdateGame(func(g *store.Game) {
		kept := g.Subscribers[:0]
		for _, id := range g.Subscribers {
			if id != chatID {
				kept = append(kept, id)
			}
		}
		g.Subscribers = kept
	})
	if err != nil {
		u.report(err)
		return
	}
	u.log.Info("unsubscribed blocked chat", "chat_id", chatID)
}

func intPtrEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func diff(first, second []int) []int {
	present := map[int]bool{}
	for _, item := range second {
		present[item] = true
	}
	var missing []int
	for _, item := range first {
		if !present[item] {
			missing = append(missing, item)
		}
	}
	return missing
}
