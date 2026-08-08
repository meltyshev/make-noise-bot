package bot

import (
	"fmt"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

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
