package bot

import (
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// restrictCallback lifts the code restriction from the button the updater
// attaches to a level announced while codes are blocked.
func (a *App) restrictCallback(c *cb) {
	current, ok := a.store.Game()
	if !ok {
		c.answer(texts.NoActiveGame)
		c.edit(c.query.Message.Message.Text, nil)
		return
	}
	if !current.Restricted {
		c.answer(texts.AlreadyProcessed)
		c.edit(c.query.Message.Message.Text, nil)
		return
	}

	if err := a.store.UpdateGame(func(g *store.Game) { g.Restricted = false }); err != nil {
		a.reportError(err)
		c.answer("")
		return
	}

	c.answer(texts.RestrictOff)
	c.edit(c.query.Message.Message.Text+"\n\n"+texts.RestrictOff, nil)
}

// stopGameCallback ends the game from the button the updater attaches when
// the engine stops serving a level.
func (a *App) stopGameCallback(c *cb) {
	// The button outlives the message: a finished game, or a level served
	// again, leaves nothing for an old tap to stop.
	current, ok := a.store.Game()
	if !ok || current.LevelNumber != nil {
		c.answer(texts.AlreadyProcessed)
		c.edit(c.query.Message.Message.Text, nil)
		return
	}

	if err := a.store.Update(func(d *store.Data) { d.Game = nil }); err != nil {
		a.reportError(err)
		c.answer("")
		return
	}

	c.answer(texts.GameOver)
	c.edit(c.query.Message.Message.Text+"\n\n"+texts.GameOver, nil)
}
