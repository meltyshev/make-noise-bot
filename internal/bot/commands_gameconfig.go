package bot

import (
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// cmdGameConfig is the inline editor for the next game's settings. Buttons
// are handled in gameConfigCallback; typed field values arrive here.
func cmdGameConfig() *Command {
	return &Command{
		Name: "gameconfig",
		Init: func(c *Ctx, _ string) {
			if !c.IsManager() || !c.EnsurePrivate() {
				return
			}
			c.ReplyMenu(renderGameConfigMenu)
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

			if field.Field == "code_formats" {
				formats, valid := parseCodeFormats(value)
				if !valid {
					c.Reply(texts.FormatsInvalid)
					return
				}
				if err := c.app.applyFormats(formats); err != nil {
					c.app.reportError(err)
					return
				}
			} else {
				err := c.app.store.Update(func(d *store.Data) {
					applyGameConfigField(&d.GameConfig, field.Field, value)
				})
				if err != nil {
					c.app.reportError(err)
					return
				}
			}

			c.DelConv()
			c.app.editMenu(c.ctx, field.ChatID, field.MsgID, renderGameConfigMenu)
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
	}
}
