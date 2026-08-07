package bot

import (
	"context"
	"fmt"
	"sort"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/game"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// Permission requests.

func permRequestKeyboard(chatID int64) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{btnRow(
		btn(texts.ButtonAllow, fmt.Sprintf("perm:a:%d", chatID)),
		btn(texts.ButtonForbid, fmt.Sprintf("perm:f:%d", chatID)),
	)}
}

func (a *App) permCallback(c *cb, args []string) {
	if len(args) < 2 {
		c.answer("")
		return
	}

	target, mark, notification := store.PermissionAllowed, texts.PermissionAllowedMark, texts.PermissionGranted
	if args[0] == "f" {
		target, mark, notification = store.PermissionForbidden, texts.PermissionForbiddenMark, texts.PermissionForbidden
	}
	chatID, ok := argID(args, 1)
	if !ok {
		c.answer("")
		return
	}

	found, changed := false, false
	err := a.store.Update(func(d *store.Data) {
		if chat, exists := d.Chats[chatID]; exists {
			found = true
			if chat.Permission != target {
				chat.Permission = target
				changed = true
			}
		}
	})
	if err != nil {
		a.reportError(err)
		c.answer("")
		return
	}

	base := c.query.Message.Message.Text
	switch {
	case !found:
		c.answer(texts.ChatNotFound)
		c.edit(base+"\n\n"+texts.ChatNotFound, nil)
	case !changed:
		c.answer(texts.AlreadyProcessed)
		c.edit(base+"\n\n"+mark, nil)
	default:
		c.answer(texts.Done)
		c.edit(base+"\n\n"+mark, nil)
		if err := a.send(c.ctx, chatID, notification); err != nil {
			a.log.Warn("permission notification failed", "chat_id", chatID, "error", err)
		}
	}
}

// /config: managers and leave mode.

func renderConfigMenu(d *store.Data) (string, [][]models.InlineKeyboardButton) {
	leaveWord := texts.LeaveModeOffWord
	if d.LeaveMode {
		leaveWord = texts.LeaveModeOnWord
	}
	return texts.SettingsTitle, [][]models.InlineKeyboardButton{
		btnRow(btn(fmt.Sprintf(texts.ManagersCountFmt, len(d.Managers)), "cfg:mgr")),
		btnRow(btn(fmt.Sprintf(texts.LeaveModeFmt, leaveWord), "cfg:leave")),
		btnRow(btn(texts.ButtonReset, "cfg:reset"), btn(texts.ButtonClose, "cfg:close")),
	}
}

func renderManagers(d *store.Data) (string, [][]models.InlineKeyboardButton) {
	managers := map[int64]bool{}
	for _, id := range d.Managers {
		managers[id] = true
	}

	// Candidates are allowed private chats; current managers stay listed so
	// they can be removed.
	ids := map[int64]bool{}
	for id, chat := range d.Chats {
		if chat.Type == "private" && chat.Permission == store.PermissionAllowed {
			ids[id] = true
		}
	}
	for id := range managers {
		ids[id] = true
	}

	var keyboard [][]models.InlineKeyboardButton
	for _, id := range sortByName(d, ids) {
		keyboard = append(keyboard, btnRow(
			btn(mark(managers[id], d.DisplayName(id)), fmt.Sprintf("cfg:mgrt:%d", id)),
		))
	}
	keyboard = append(keyboard, btnRow(btn(texts.ButtonAdd, "cfg:mgradd"), btn(texts.ButtonBack, "cfg:menu")))
	return texts.ManagersTitle, keyboard
}

// pickManagers is the conversation state while the admin uses the native
// user picker; it remembers which menu message to refresh.
type pickManagers struct {
	ChatID int64
	MsgID  int
}

func (a *App) configCallback(c *cb, args []string) {
	if len(args) == 0 {
		c.answer("")
		return
	}

	editWith := func(render func(d *store.Data) (string, [][]models.InlineKeyboardButton)) {
		var (
			text     string
			keyboard [][]models.InlineKeyboardButton
		)
		a.store.View(func(d *store.Data) { text, keyboard = render(d) })
		c.edit(text, keyboard)
	}

	switch args[0] {
	case "menu":
		c.answer("")
		editWith(renderConfigMenu)
	case "leave":
		err := a.store.Update(func(d *store.Data) { d.LeaveMode = !d.LeaveMode })
		if err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer("")
		editWith(renderConfigMenu)
	case "reset":
		err := a.store.Update(func(d *store.Data) {
			d.Managers = []int64{}
			d.LeaveMode = false
		})
		if err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer(texts.Done)
		editWith(renderConfigMenu)
	case "close":
		c.answer("")
		c.edit(texts.Done, nil)
	case "mgr":
		c.answer("")
		editWith(renderManagers)
	case "mgrt":
		id, ok := argID(args, 1)
		if !ok {
			c.answer("")
			return
		}
		err := a.store.Update(func(d *store.Data) {
			for i, manager := range d.Managers {
				if manager == id {
					d.Managers = append(d.Managers[:i], d.Managers[i+1:]...)
					return
				}
			}
			d.Managers = append(d.Managers, id)
		})
		if err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer("")
		editWith(renderManagers)
	case "mgradd":
		a.conv.Set(c.query.From.ID, c.chatID, "config", pickManagers{ChatID: c.chatID, MsgID: c.msgID})
		c.answer("")
		a.sendUserPicker(c.ctx, c.chatID)
	default:
		c.answer("")
	}
}

func (a *App) sendUserPicker(ctx context.Context, chatID int64) {
	_, err := a.tg.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   texts.PickUserAsk,
		ReplyMarkup: &models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{{
					Text: texts.ButtonPickUser,
					RequestUsers: &models.KeyboardButtonRequestUsers{
						RequestID:   1,
						RequestName: true,
						MaxQuantity: 10,
					},
				}},
				{{Text: texts.ButtonCancel}},
			},
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		},
	})
	if err != nil {
		a.reportError(err)
	}
}

// /chats: list and per-chat actions.

func renderChatsList(d *store.Data) (string, [][]models.InlineKeyboardButton) {
	ids := map[int64]bool{}
	for id := range d.Chats {
		ids[id] = true
	}

	var keyboard [][]models.InlineKeyboardButton
	for _, id := range sortByName(d, ids) {
		chat := d.Chats[id]
		marker := "❓"
		switch chat.Permission {
		case store.PermissionAllowed:
			marker = "✅"
		case store.PermissionForbidden:
			marker = "🚫"
		}
		keyboard = append(keyboard, btnRow(btn(marker+" "+d.DisplayName(id), fmt.Sprintf("ch:open:%d", id))))
	}
	keyboard = append(keyboard, btnRow(btn(texts.ButtonClose, "ch:close")))
	return texts.ChatsChoose, keyboard
}

func renderChatActions(chat store.Chat) (string, [][]models.InlineKeyboardButton) {
	text := fmt.Sprintf(texts.ChatsActionsFmt, chat.DisplayName(), chat.ID, chat.Type, chat.Permission)
	return text, [][]models.InlineKeyboardButton{
		btnRow(
			btn(texts.ButtonAllow, fmt.Sprintf("ch:allow:%d", chat.ID)),
			btn(texts.ButtonForbid, fmt.Sprintf("ch:forbid:%d", chat.ID)),
		),
		btnRow(btn(texts.ButtonDelete, fmt.Sprintf("ch:del:%d", chat.ID))),
		btnRow(btn(texts.ButtonBack, "ch:list"), btn(texts.ButtonClose, "ch:close")),
	}
}

func (a *App) chatsCallback(c *cb, args []string) {
	if len(args) == 0 {
		c.answer("")
		return
	}

	showList := func() {
		var (
			text     string
			keyboard [][]models.InlineKeyboardButton
		)
		a.store.View(func(d *store.Data) { text, keyboard = renderChatsList(d) })
		c.edit(text, keyboard)
	}
	showActions := func(id int64) bool {
		chat, ok := a.store.Chat(id)
		if !ok {
			return false
		}
		text, keyboard := renderChatActions(chat)
		c.edit(text, keyboard)
		return true
	}

	switch args[0] {
	case "list":
		c.answer("")
		showList()
	case "close":
		c.answer("")
		c.edit(texts.ChatsClosed, nil)
	case "open":
		id, ok := argID(args, 1)
		c.answer("")
		if !ok || !showActions(id) {
			showList()
		}
	case "allow", "forbid":
		id, ok := argID(args, 1)
		if !ok {
			c.answer("")
			return
		}
		target, notification := store.PermissionAllowed, texts.PermissionGranted
		if args[0] == "forbid" {
			target, notification = store.PermissionForbidden, texts.PermissionForbidden
		}

		found, changed := false, false
		err := a.store.Update(func(d *store.Data) {
			if chat, exists := d.Chats[id]; exists {
				found = true
				if chat.Permission != target {
					chat.Permission = target
					changed = true
				}
			}
		})
		if err != nil {
			a.reportError(err)
			c.answer("")
			return
		}

		switch {
		case !found:
			c.answer(texts.ChatNotFound)
			showList()
		case !changed:
			c.answer(texts.AlreadyProcessed)
		default:
			c.answer(texts.Done)
			showActions(id)
			if err := a.send(c.ctx, id, notification); err != nil {
				a.log.Warn("permission notification failed", "chat_id", id, "error", err)
			}
		}
	case "del":
		id, ok := argID(args, 1)
		if !ok {
			c.answer("")
			return
		}
		err := a.store.Update(func(d *store.Data) { delete(d.Chats, id) })
		if err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer(texts.ChatsDeleted)
		showList()
	default:
		c.answer("")
	}
}

// /gameconfig: settings for the next game plus live game subscribers.

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

	editWith := func(render func(d *store.Data) (string, [][]models.InlineKeyboardButton)) {
		var (
			text     string
			keyboard [][]models.InlineKeyboardButton
		)
		a.store.View(func(d *store.Data) { text, keyboard = render(d) })
		c.edit(text, keyboard)
	}
	showMenu := func() { editWith(renderGameConfigMenu) }

	switch args[0] {
	case "menu":
		c.answer("")
		showMenu()
	case "engine":
		c.answer("")
		editWith(renderEngineChoice)
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
		editWith(renderFormatsMenu)
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
		c.answer(texts.Done)
		editWith(renderFormatsMenu)
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

func sortByName(d *store.Data, ids map[int64]bool) []int64 {
	sorted := make([]int64, 0, len(ids))
	for id := range ids {
		sorted = append(sorted, id)
	}
	sort.Slice(sorted, func(i, j int) bool {
		ni, nj := d.DisplayName(sorted[i]), d.DisplayName(sorted[j])
		if ni != nj {
			return ni < nj
		}
		return sorted[i] < sorted[j]
	})
	return sorted
}
