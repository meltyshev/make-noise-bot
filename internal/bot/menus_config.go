package bot

import (
	"context"
	"fmt"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/geo"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

func renderConfigMenu(d *store.Data) (string, [][]models.InlineKeyboardButton) {
	leaveWord := texts.LeaveModeOffWord
	if d.LeaveMode {
		leaveWord = texts.LeaveModeOnWord
	}
	return texts.SettingsTitle, [][]models.InlineKeyboardButton{
		btnRow(btn(fmt.Sprintf(texts.ManagersCountFmt, len(d.Managers)), "cfg:mgr")),
		btnRow(btn(fmt.Sprintf(texts.MapServiceFmt, mapService(d).Label()), "cfg:maps")),
		btnRow(btn(fmt.Sprintf(texts.LeaveModeFmt, leaveWord), "cfg:leave")),
		btnRow(btn(texts.ButtonReset, "cfg:reset"), btn(texts.ButtonClose, "cfg:close")),
	}
}

func mapService(d *store.Data) geo.Service {
	service := geo.Service(d.MapService)
	if !service.Valid() {
		return geo.DefaultService
	}
	return service
}

func renderMapServices(d *store.Data) (string, [][]models.InlineKeyboardButton) {
	current := mapService(d)

	var keyboard [][]models.InlineKeyboardButton
	for _, service := range geo.Services {
		keyboard = append(keyboard, btnRow(
			btn(mark(service == current, service.Label()), "cfg:setmap:"+string(service)),
		))
	}
	keyboard = append(keyboard, btnRow(btn(texts.ButtonBack, "cfg:menu")))
	return texts.MapServiceTitle, keyboard
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
		if chat.Type == string(models.ChatTypePrivate) && chat.Permission == store.PermissionAllowed {
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

	switch args[0] {
	case "menu":
		c.answer("")
		c.show(renderConfigMenu)
	case "leave":
		err := a.store.Update(func(d *store.Data) { d.LeaveMode = !d.LeaveMode })
		if err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer("")
		c.show(renderConfigMenu)
	case "reset":
		err := a.store.Update(func(d *store.Data) {
			d.Managers = []int64{}
			d.LeaveMode = false
			d.MapService = ""
		})
		if err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer(texts.Done)
		c.show(renderConfigMenu)
	case "close":
		c.answer("")
		c.edit(texts.Done, nil)
	case "maps":
		c.answer("")
		c.show(renderMapServices)
	case "setmap":
		if len(args) < 2 || !geo.Service(args[1]).Valid() {
			c.answer("")
			return
		}
		if err := a.store.Update(func(d *store.Data) { d.MapService = args[1] }); err != nil {
			a.reportError(err)
			c.answer("")
			return
		}
		c.answer("")
		c.show(renderConfigMenu)
	case "mgr":
		c.answer("")
		c.show(renderManagers)
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
		c.show(renderManagers)
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
