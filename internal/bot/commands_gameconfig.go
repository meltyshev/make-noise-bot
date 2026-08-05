package bot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/game"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// cmdGameConfig is the keyboard-driven editor for the next game's settings.
func cmdGameConfig() *Command {
	menu := func(c *Ctx) {
		c.SetConv("gameconfig")
		cfg := c.app.store.GameConfig()

		formatsJSON, _ := json.Marshal(cfg.CodeFormats)
		subscribersJSON, _ := json.Marshal(cfg.Subscribers)

		c.ReplyKeyboard(texts.SettingsTitle, [][]models.KeyboardButton{
			kbRow(fmt.Sprintf(texts.GameConfigEngineFmt, cfg.Engine)),
			kbRow(fmt.Sprintf(texts.GameConfigCityFmt, cfg.City)),
			kbRow(fmt.Sprintf(texts.GameConfigLoginFmt, cfg.Login)),
			kbRow(fmt.Sprintf(texts.GameConfigPasswordFmt, cfg.Password)),
			kbRow(fmt.Sprintf(texts.GameConfigPincodeFmt, cfg.Pincode)),
			kbRow(fmt.Sprintf(texts.GameConfigGameIDFmt, cfg.GameID)),
			kbRow(fmt.Sprintf(texts.GameConfigLeagueFmt, cfg.League)),
			kbRow(fmt.Sprintf(texts.GameConfigFormatsFmt, formatsJSON)),
			kbRow(fmt.Sprintf(texts.GameConfigSubscribersFmt, subscribersJSON)),
			kbRow(texts.ButtonReset, texts.ButtonFinish),
		})
	}

	askValue := func(c *Ctx, field, prompt string) {
		c.SetConvState("gameconfig", field)
		c.ReplyKeyboard(prompt, [][]models.KeyboardButton{kbRow(texts.ButtonCancel)})
	}

	setValue := func(c *Ctx, field, value string) {
		err := c.app.store.Update(func(d *store.Data) {
			cfg := &d.GameConfig
			switch field {
			case "engine":
				if game.IsKnownEngine(value) {
					cfg.Engine = value
				}
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
					// "null" decodes into a nil slice, which would render as
					// null later.
					if formats == nil {
						formats = [][]string{}
					}
					cfg.CodeFormats = formats
				}
			case "subscribers":
				var subscribers []int64
				if json.Unmarshal([]byte(value), &subscribers) == nil {
					if subscribers == nil {
						subscribers = []int64{}
					}
					cfg.Subscribers = subscribers
				}
			}
		})
		if err != nil {
			c.app.reportError(err)
		}
	}

	return &Command{
		Name: "gameconfig",
		Init: func(c *Ctx, _ string) {
			if !c.IsManager() {
				return
			}
			menu(c)
		},
		Handle: func(c *Ctx, state any) {
			text := c.Text()
			if text == "" {
				c.Reply(texts.TextRequired)
				return
			}

			if field, ok := state.(string); ok {
				if text != texts.ButtonCancel {
					setValue(c, field, text)
				}
				menu(c)
				return
			}

			switch {
			case strings.HasPrefix(text, "Движок"):
				c.SetConvState("gameconfig", "engine")
				keyboard := make([][]models.KeyboardButton, 0, len(game.Names)+1)
				for _, name := range game.Names {
					keyboard = append(keyboard, kbRow(name))
				}
				keyboard = append(keyboard, kbRow(texts.ButtonCancel))
				c.ReplyKeyboard(texts.GameConfigEngineAsk, keyboard)
			case strings.HasPrefix(text, "Город"):
				askValue(c, "city", texts.GameConfigCityAsk)
			case strings.HasPrefix(text, "Логин"):
				askValue(c, "login", texts.GameConfigLoginAsk)
			case strings.HasPrefix(text, "Пароль"):
				askValue(c, "password", texts.GameConfigPasswordAsk)
			case strings.HasPrefix(text, "Пинка"):
				askValue(c, "pincode", texts.GameConfigPincodeAsk)
			case strings.HasPrefix(text, "Номер игры"):
				askValue(c, "game_id", texts.GameConfigGameIDAsk)
			case strings.HasPrefix(text, "Лига"):
				askValue(c, "league", texts.GameConfigLeagueAsk)
			case strings.HasPrefix(text, "Форматы кода"):
				askValue(c, "code_formats", texts.GameConfigFormatsAsk)
			case strings.HasPrefix(text, "Подписчики"):
				askValue(c, "subscribers", texts.GameConfigSubscribersAsk)
			case text == texts.ButtonReset:
				err := c.app.store.Update(func(d *store.Data) { d.GameConfig = store.DefaultGameConfig() })
				if err != nil {
					c.app.reportError(err)
					return
				}
				menu(c)
			case text == texts.ButtonFinish:
				c.DelConv()
				c.ReplyRemoveKeyboard(texts.Done)
			}
		},
	}
}
