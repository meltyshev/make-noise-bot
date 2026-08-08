package bot

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/game"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

func cmdGame() *Command {
	return &Command{
		Name: "game",
		Init: func(c *Ctx, _ string) {
			if !c.IsManager() {
				return
			}

			if engine := c.app.Engine(); engine != nil {
				if err := c.app.store.Update(func(d *store.Data) { d.Game = nil }); err != nil {
					c.app.reportError(err)
					return
				}
				c.Reply(texts.GameOver)
				return
			}

			cfg := c.app.store.GameConfig()
			newGame, err := game.Start(c.ctx, cfg, c.app.env)
			if err != nil {
				c.app.log.Warn("game start failed", "error", err)
				c.Reply(texts.GameCannotStart)
				return
			}
			if err := c.app.store.Update(func(d *store.Data) { d.Game = newGame }); err != nil {
				c.app.reportError(err)
				return
			}
			// newGame is the stored game now, so the link comes from a copy
			// rather than from a pointer the updater also writes to.
			g, ok := c.app.store.Game()
			if !ok {
				return
			}
			c.Reply(game.New(g, c.app.env).Link())
		},
	}
}

func cmdSubscribe() *Command {
	return &Command{
		Name: "subscribe",
		Init: func(c *Ctx, _ string) {
			if !c.IsManager() {
				return
			}
			if _, ok := c.app.store.Game(); !ok {
				c.Reply(texts.NoActiveGame)
				return
			}

			c.ReplyMenu(func(d *store.Data) (string, [][]models.InlineKeyboardButton) {
				return renderSubscriptionDetail(d, true, c.ChatID(), true)
			})
		},
	}
}

func cmdQuestion() *Command {
	return &Command{
		Name:        "question",
		Description: texts.DescQuestion,
		Init: func(c *Ctx, _ string) {
			if !c.EnsureAllowed() {
				return
			}

			engine := c.app.Engine()
			if engine == nil {
				c.Reply(texts.NoActiveGame)
				return
			}

			snap, err := engine.Load(c.ctx)
			if err != nil {
				c.app.log.Warn("engine load failed", "error", err)
				c.Reply(texts.CannotLoadEngine)
				return
			}
			if question := snap.Question(); question != "" {
				c.ReplyHTML(question)
			} else {
				c.Reply(texts.QuestionNone)
			}
		},
	}
}

func cmdLink() *Command {
	return &Command{
		Name:        "link",
		Description: texts.DescLink,
		Init: func(c *Ctx, _ string) {
			if !c.EnsureAllowed() {
				return
			}
			if engine := c.app.Engine(); engine != nil {
				c.Reply(engine.Link())
			} else {
				c.Reply(texts.NoActiveGame)
			}
		},
	}
}

func cmdRestrict() *Command {
	return &Command{
		Name: "restrict",
		Init: func(c *Ctx, _ string) {
			if !c.IsManager() {
				return
			}
			if _, ok := c.app.store.Game(); !ok {
				c.Reply(texts.NoActiveGame)
				return
			}

			restricted := false
			err := c.app.store.UpdateGame(func(g *store.Game) {
				g.Restricted = !g.Restricted
				restricted = g.Restricted
			})
			if err != nil {
				c.app.reportError(err)
				return
			}

			if restricted {
				c.Reply(texts.RestrictOn)
			} else {
				c.Reply(texts.RestrictOff)
			}
		},
	}
}

func cmdBruteForce() *Command {
	return &Command{
		Name: "bruteforce",
		Init: func(c *Ctx, _ string) {
			if !c.EnsureAllowed() {
				return
			}
			if c.ChatType() != string(models.ChatTypePrivate) && !c.IsManager() {
				return
			}

			enabled := false
			err := c.app.store.Update(func(d *store.Data) {
				if chat, ok := d.Chats[c.ChatID()]; ok {
					chat.BruteForce = !chat.BruteForce
					enabled = chat.BruteForce
				}
			})
			if err != nil {
				c.app.reportError(err)
				return
			}

			if enabled {
				c.Reply(texts.BruteForceOn)
			} else {
				c.Reply(texts.BruteForceOff)
			}
		},
	}
}

func cmdPinLevel() *Command {
	apply := func(c *Ctx, input string) {
		level, err := strconv.Atoi(input)
		if err != nil || level < 0 {
			c.Reply(texts.PinLevelRequired)
			return
		}
		if c.app.ClassicEngine() == nil {
			c.Reply(texts.NoActiveGame)
			return
		}

		if err := c.app.store.UpdateGame(func(g *store.Game) { g.PinnedLevel = &level }); err != nil {
			c.app.reportError(err)
			return
		}
		c.Reply(texts.Done)
	}

	return &Command{
		Name: "pinlevel",
		Init: func(c *Ctx, args string) {
			if !c.IsManager() {
				return
			}
			if c.app.ClassicEngine() == nil {
				c.Reply(texts.NoActiveGame)
				return
			}
			if args != "" {
				c.DelConv()
				apply(c, args)
				return
			}
			c.SetConv("pinlevel")
			c.Reply(texts.PinLevelAsk)
		},
		Handle: func(c *Ctx, _ any) {
			c.DelConv()
			apply(c, c.Text())
		},
	}
}

func cmdUnpinLevel() *Command {
	return &Command{
		Name: "unpinlevel",
		Init: func(c *Ctx, _ string) {
			if !c.IsManager() {
				return
			}
			if c.app.ClassicEngine() == nil {
				c.Reply(texts.NoActiveGame)
				return
			}
			if err := c.app.store.UpdateGame(func(g *store.Game) { g.PinnedLevel = nil }); err != nil {
				c.app.reportError(err)
				return
			}
			c.Reply(texts.Done)
		},
	}
}

func cmdRating() *Command {
	return &Command{
		Name:        "rating",
		Description: texts.DescRating,
		Init: func(c *Ctx, _ string) {
			players := c.app.store.Rating()
			if len(players) == 0 {
				c.Reply(texts.RatingNone)
				return
			}

			var lines []string
			for i, player := range players {
				lines = append(lines, fmt.Sprintf("%3s %-2d %s", fmt.Sprintf("%d)", i+1), player.Total, player.Name))
			}
			c.ReplyHTML("<pre>" + html.EscapeString(strings.Join(lines, "\n")) + "</pre>")
		},
	}
}

func cmdClearRating() *Command {
	return &Command{
		Name: "clearrating",
		Init: func(c *Ctx, _ string) {
			if !c.IsManager() {
				return
			}
			err := c.app.store.Update(func(d *store.Data) { d.Players = map[int64]*store.Player{} })
			if err != nil {
				c.app.reportError(err)
				return
			}
			c.Reply(texts.RatingCleared)
		},
	}
}
