package bot

import (
	"encoding/json"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// cmdGameConfig is the inline editor for the next game's settings. Buttons
// are handled in gameConfigCallback; typed field values arrive here.
func cmdGameConfig() *Command {
	return &Command{
		Name: "gameconfig",
		Init: func(c *Ctx, _ string) {
			if !c.IsManager() {
				return
			}
			var (
				text     string
				keyboard [][]models.InlineKeyboardButton
			)
			c.app.store.View(func(d *store.Data) { text, keyboard = renderGameConfigMenu(d) })
			c.ReplyInline(text, keyboard)
		},
		Handle: func(c *Ctx, state any) {
			field, ok := state.(gcField)
			if !ok {
				c.DelConv()
				return
			}

			value := c.Text()
			if value == "" {
				c.Reply(texts.TextRequired)
				return
			}

			err := c.app.store.Update(func(d *store.Data) {
				applyGameConfigField(&d.GameConfig, field.Field, value)
			})
			if err != nil {
				c.app.reportError(err)
				return
			}

			c.DelConv()

			var (
				text     string
				keyboard [][]models.InlineKeyboardButton
			)
			c.app.store.View(func(d *store.Data) { text, keyboard = renderGameConfigMenu(d) })
			c.app.editMessage(c.ctx, field.ChatID, field.MsgID, text, keyboard)
		},
	}
}

func applyGameConfigField(cfg *store.GameConfig, field, value string) {
	switch field {
	case "city":
		cfg.City = value
	case "login":
		cfg.Login = value
	case "password":
		cfg.Password = value
	case "pincode":
		cfg.Pincode = value
	case "game_id":
		cfg.GameID = value
	case "league":
		cfg.League = value
	case "code_formats":
		var formats [][]string
		if json.Unmarshal([]byte(value), &formats) == nil {
			if formats == nil {
				formats = [][]string{}
			}
			cfg.CodeFormats = formats
		}
	}
}
