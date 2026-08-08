package bot

import (
	"fmt"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/game"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

var gameConfigFields = map[string]string{
	"city":         texts.GameConfigCityAsk,
	"login":        texts.GameConfigLoginAsk,
	"password":     texts.GameConfigPasswordAsk,
	"pincode":      texts.GameConfigPincodeAsk,
	"game_id":      texts.GameConfigGameIDAsk,
	"league":       texts.GameConfigLeagueAsk,
	"code_formats": texts.FormatsManualAsk,
}

// gcField is the conversation state while the admin types a value for one
// game config field.
type gcField struct {
	Field  string
	ChatID int64
	MsgID  int
}

func renderGameConfigMenu(d *store.Data) (string, [][]models.InlineKeyboardButton) {
	cfg := d.GameConfig

	keyboard := [][]models.InlineKeyboardButton{
		btnRow(btn(fmt.Sprintf(texts.GameConfigEngineFmt, cfg.Engine), "gc:engine")),
		btnRow(btn(fmt.Sprintf(texts.GameConfigCityFmt, cfg.City), "gc:field:city")),
		btnRow(btn(fmt.Sprintf(texts.GameConfigLoginFmt, cfg.Login), "gc:field:login")),
		btnRow(btn(fmt.Sprintf(texts.GameConfigPasswordFmt, cfg.Password), "gc:field:password")),
		btnRow(btn(fmt.Sprintf(texts.GameConfigPincodeFmt, cfg.Pincode), "gc:field:pincode")),
		btnRow(btn(fmt.Sprintf(texts.GameConfigGameIDFmt, cfg.GameID), "gc:field:game_id")),
		btnRow(btn(fmt.Sprintf(texts.GameConfigLeagueFmt, cfg.League), "gc:field:league")),
		btnRow(btn(fmt.Sprintf(texts.GameConfigFormatsFmt, formatsLabel(cfg.CodeFormats)), "gc:fmt")),
		btnRow(btn(fmt.Sprintf(texts.GameConfigSubscribersFmt, len(cfg.Subscriptions)), "cs:list")),
	}
	if d.Game != nil {
		keyboard = append(keyboard, btnRow(
			btn(fmt.Sprintf(texts.GameSubscribersCountFmt, len(d.Game.Subscriptions)), "gs:list"),
		))
	}
	keyboard = append(keyboard, btnRow(btn(texts.ButtonReset, "gc:reset"), btn(texts.ButtonClose, "gc:close")))
	return texts.SettingsTitle, keyboard
}

func renderFormatsMenu(d *store.Data) (string, [][]models.InlineKeyboardButton) {
	var keyboard [][]models.InlineKeyboardButton
	for i, preset := range formatPresets {
		active := formatsEqual(d.GameConfig.CodeFormats, preset.Formats)
		keyboard = append(keyboard, btnRow(
			btn(mark(active, preset.Label), fmt.Sprintf("gc:fmtp:%d", i)),
		))
	}
	keyboard = append(keyboard,
		btnRow(btn(texts.ButtonManual, "gc:fmtm")),
		btnRow(btn(texts.ButtonBack, "gc:menu")),
	)
	return texts.FormatsTitle, keyboard
}

func renderEngineChoice(d *store.Data) (string, [][]models.InlineKeyboardButton) {
	var keyboard [][]models.InlineKeyboardButton
	for _, name := range game.Names {
		keyboard = append(keyboard, btnRow(
			btn(mark(d.GameConfig.Engine == name, name), "gc:seteng:"+name),
		))
	}
	keyboard = append(keyboard, btnRow(btn(texts.ButtonBack, "gc:menu")))
	return texts.GameConfigEngineAsk, keyboard
}

func (a *App) gameConfigCallback(c *cb, args []string) {
	if len(args) == 0 {
		c.answer("")
		return
	}

	showMenu := func() { c.show(renderGameConfigMenu) }

	switch args[0] {
	case "menu":
		c.answer("")
		showMenu()
	case "engine":
		c.answer("")
		c.show(renderEngineChoice)
	case "seteng":
		if len(args) < 2 || !game.IsKnownEngine(args[1]) {
			c.answer("")
			return
		}
		err := a.store.Update(func(d *store.Data) { d.GameConfig.Engine = args[1] })
		if err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer("")
		showMenu()
	case "field":
		if len(args) < 2 {
			c.answer("")
			return
		}
		ask, known := gameConfigFields[args[1]]
		if !known {
			c.answer("")
			return
		}
		a.conv.Set(c.query.From.ID, c.chatID, "gameconfig", gcField{Field: args[1], ChatID: c.chatID, MsgID: c.msgID})
		c.answer("")
		if err := a.send(c.ctx, c.chatID, ask); err != nil {
			a.reportError(err)
		}
	case "fmt":
		c.answer("")
		c.show(renderFormatsMenu)
	case "fmtp":
		if len(args) < 2 {
			c.answer("")
			return
		}
		idx, ok := argID(args, 1)
		if !ok || idx < 0 || int(idx) >= len(formatPresets) {
			c.answer("")
			return
		}
		if err := a.applyFormats(formatPresets[idx].Formats); err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer("")
		showMenu()
	case "fmtm":
		a.conv.Set(c.query.From.ID, c.chatID, "gameconfig", gcField{Field: "code_formats", ChatID: c.chatID, MsgID: c.msgID})
		c.answer("")
		if err := a.send(c.ctx, c.chatID, texts.FormatsManualAsk); err != nil {
			a.reportError(err)
		}
	case "reset":
		err := a.store.Update(func(d *store.Data) { d.GameConfig = store.DefaultGameConfig() })
		if err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer(texts.Done)
		showMenu()
	case "close":
		c.answer("")
		c.edit(texts.Done, nil)
	default:
		c.answer("")
	}
}
