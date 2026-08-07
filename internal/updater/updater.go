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
	"github.com/meltyshev/make-noise-bot/internal/geo"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
	"github.com/meltyshev/make-noise-bot/internal/tgsend"
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
	if !ok || len(g.Subscriptions) == 0 {
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
	previous := levelState{Number: g.LevelNumber, Task: g.LevelTask, Time: g.LevelTime}
	current := levelState{Number: snap.LevelNumber(), Task: levelTask(snap), Time: timeOnLevel(snap)}

	currentHint := g.HintNumber
	currentSolved := g.SolvedSpoilers

	newLevel := isNewLevel(previous, current)
	gone := levelGone(previous, current)

	if !sameLevel(previous, current) {
		err := u.store.UpdateGame(func(g *store.Game) {
			g.LevelNumber = current.Number
			g.LevelTask = current.Task
			g.LevelTime = current.Time
			if newLevel {
				g.HintNumber = nil
				g.SolvedSpoilers = nil
			}
		})
		if err != nil {
			u.report(err)
			return
		}
	}

	if newLevel {
		currentHint = nil
		currentSolved = nil

		// A restriction outlives the level it was set on, so the new level
		// carries the button to lift it.
		var markup models.ReplyMarkup
		if g.Restricted {
			markup = buttonRow(texts.ButtonAllowCodes, texts.CallbackAllowCodes)
		}
		u.broadcastWith(ctx, wantLevelUp, texts.LevelUp, false, markup)

		if question := snap.Question(); question != "" {
			u.broadcast(ctx, wantQuestion, question, true)
		}
		if notes := snap.Notes(); notes != "" {
			u.broadcast(ctx, wantNotes, notes, true)
		}
	}

	if gone {
		u.broadcastWith(ctx, wantLevelUp, texts.LevelGone, false, buttonRow(texts.ButtonStopGame, texts.CallbackStopGame))
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
			u.broadcast(ctx, wantHints, fmt.Sprintf(texts.HintFmt, hintNumber, hintText), true)
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
			u.broadcast(ctx, wantSpoilers, fmt.Sprintf(texts.SpoilerSolved, spoiler), false)
		}
	}
}

func buttonRow(label, data string) models.ReplyMarkup {
	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{
		{Text: label, CallbackData: data},
	}}}
}

type wants func(store.Subscription) bool

func wantLevelUp(s store.Subscription) bool  { return s.LevelUp }
func wantHints(s store.Subscription) bool    { return s.Hints }
func wantSpoilers(s store.Subscription) bool { return s.Spoilers }
func wantQuestion(s store.Subscription) bool { return s.Question }
func wantNotes(s store.Subscription) bool    { return s.Notes }

func (u *Updater) broadcast(ctx context.Context, want wants, text string, html bool) {
	u.broadcastWith(ctx, want, text, html, nil)
}

// A chat that blocked the bot is unsubscribed automatically.
func (u *Updater) broadcastWith(ctx context.Context, want wants, text string, html bool, markup models.ReplyMarkup) {
	g, ok := u.store.Game()
	if !ok {
		return
	}
	for _, sub := range g.Subscriptions {
		if want(sub) {
			u.sendTo(ctx, sub.ChatID, text, html, markup)
		}
	}
}

func (u *Updater) sendTo(ctx context.Context, chatID int64, text string, html bool, markup models.ReplyMarkup) {
	var err error
	if html {
		err = tgsend.HTML(ctx, u.tg, chatID, text, geo.Linker(u.store.MapService()), nil)
	} else {
		_, err = u.tg.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ReplyMarkup: markup,
		})
	}

	if err != nil {
		if errors.Is(err, tgbot.ErrorForbidden) {
			u.unsubscribe(chatID)
			return
		}
		u.log.Warn("broadcast failed", "chat_id", chatID, "error", err)
	}
}

func (u *Updater) unsubscribe(chatID int64) {
	err := u.store.UpdateGame(func(g *store.Game) {
		g.Subscriptions = g.Subscriptions.Remove(chatID)
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
