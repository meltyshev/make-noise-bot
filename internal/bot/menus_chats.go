package bot

import (
	"fmt"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

func renderChatsList(d *store.Data) (string, [][]models.InlineKeyboardButton) {
	ids := map[int64]bool{}
	for id := range d.Chats {
		ids[id] = true
	}

	var keyboard [][]models.InlineKeyboardButton
	for _, id := range sortByName(d, ids) {
		chat := d.Chats[id]
		allowed := chat.Permission == store.PermissionAllowed
		label := labelWithState(d.DisplayName(id), permissionSummary(chat.Permission))
		keyboard = append(keyboard, btnRow(btn(mark(allowed, label), fmt.Sprintf("ch:open:%d", id))))
	}
	keyboard = append(keyboard, btnRow(btn(texts.ButtonClose, "ch:close")))
	return texts.ChatsChoose, keyboard
}

// labelWithState shortens the name rather than the state, so a long chat title
// cannot push the summary past the button's label budget.
func labelWithState(name, summary string) string {
	if summary == "" {
		return name
	}

	suffix := ": " + summary
	room := maxLabelRunes - len([]rune(suffix))
	if runes := []rune(name); len(runes) > room {
		name = string(runes[:room-3]) + "..."
	}
	return name + suffix
}

// permissionSummary names the states mark() cannot show, and is empty for the
// allowed one, which the checkmark already covers.
func permissionSummary(p store.Permission) string {
	switch p {
	case store.PermissionRequested:
		return texts.SummaryRequested
	case store.PermissionForbidden:
		return texts.SummaryForbidden
	default:
		return ""
	}
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

	showList := func() { c.show(renderChatsList) }
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
		case !changed:
			c.answer(texts.AlreadyProcessed)
		default:
			c.answer(texts.Done)
			if err := a.send(c.ctx, id, notification); err != nil {
				a.log.Warn("permission notification failed", "chat_id", id, "error", err)
			}
		}
		showList()
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
