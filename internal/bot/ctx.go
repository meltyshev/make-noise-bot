package bot

import (
	"bytes"
	"context"
	"fmt"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/geo"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
	"github.com/meltyshev/make-noise-bot/internal/tgsend"
)

// Ctx bundles one incoming message with everything a handler needs.
type Ctx struct {
	ctx context.Context
	app *App
	msg *models.Message
	cmd string // the command being handled, set by the dispatcher
}

func (c *Ctx) Text() string  { return c.msg.Text }
func (c *Ctx) ChatID() int64 { return c.msg.Chat.ID }
func (c *Ctx) ChatType() string {
	return string(c.msg.Chat.Type)
}

func (c *Ctx) UserID() int64 {
	if c.msg.From == nil {
		return 0
	}
	return c.msg.From.ID
}

func (c *Ctx) UserName() string {
	if c.msg.From == nil {
		return ""
	}
	name := c.msg.From.FirstName
	if c.msg.From.LastName != "" {
		if name != "" {
			name += " "
		}
		name += c.msg.From.LastName
	}
	return name
}

func (c *Ctx) isGroup() bool {
	return c.msg.Chat.Type == models.ChatTypeGroup || c.msg.Chat.Type == models.ChatTypeSupergroup
}

// EnsurePrivate keeps settings menus out of group chats.
func (c *Ctx) EnsurePrivate() bool {
	if c.msg.Chat.Type == models.ChatTypePrivate {
		return true
	}
	c.Reply(texts.PrivateOnly)
	return false
}

// replyParams quotes the triggering message in group chats, and returns nil
// everywhere else.
func (c *Ctx) replyParams() *models.ReplyParameters {
	if !c.isGroup() {
		return nil
	}
	return &models.ReplyParameters{MessageID: c.msg.ID}
}

// send delivers one reply. A failure is that chat's problem, not the
// maintainer's, so it stays at the level every single-chat send uses.
func (c *Ctx) send(params *tgbot.SendMessageParams) {
	if _, err := c.app.tg.SendMessage(c.ctx, params); err != nil {
		c.app.log.Warn("send message failed", "chat_id", c.ChatID(), "error", err)
	}
}

func (c *Ctx) Reply(text string) {
	c.send(&tgbot.SendMessageParams{ChatID: c.ChatID(), Text: text, ReplyParameters: c.replyParams()})
}

func (c *Ctx) Replyf(format string, args ...any) {
	c.Reply(fmt.Sprintf(format, args...))
}

func (c *Ctx) ReplyHTML(text string) {
	err := tgsend.HTML(c.ctx, c.app.tg, tgsend.Message{
		ChatID:  c.ChatID(),
		Text:    text,
		MapLink: geo.Linker(c.app.store.MapService()),
		Reply:   c.replyParams(),
	})
	if err != nil {
		c.app.log.Warn("send message failed", "chat_id", c.ChatID(), "error", err)
	}
}

// ReplyAlways quotes the triggering message even in private chats.
func (c *Ctx) ReplyAlways(text string) {
	c.send(&tgbot.SendMessageParams{
		ChatID:          c.ChatID(),
		Text:            text,
		ReplyParameters: &models.ReplyParameters{MessageID: c.msg.ID},
	})
}

func (c *Ctx) ReplyInline(text string, keyboard [][]models.InlineKeyboardButton) {
	c.send(&tgbot.SendMessageParams{
		ChatID:          c.ChatID(),
		Text:            text,
		ReplyParameters: c.replyParams(),
		ReplyMarkup:     &models.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	})
}

// ReplyMenu sends a freshly rendered menu as a new message.
func (c *Ctx) ReplyMenu(r render) {
	c.ReplyInline(c.app.renderMenu(r))
}

func (c *Ctx) ReplyRemoveKeyboard(text string) {
	c.send(&tgbot.SendMessageParams{
		ChatID:          c.ChatID(),
		Text:            text,
		ReplyParameters: c.replyParams(),
		ReplyMarkup:     &models.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
}

func (c *Ctx) ReplyPhoto(png []byte, caption string) {
	_, err := c.app.tg.SendPhoto(c.ctx, &tgbot.SendPhotoParams{
		ChatID:          c.ChatID(),
		Photo:           &models.InputFileUpload{Filename: "avatar.png", Data: bytes.NewReader(png)},
		Caption:         caption,
		ReplyParameters: c.replyParams(),
	})
	if err != nil {
		c.app.log.Warn("send photo failed", "chat_id", c.ChatID(), "error", err)
	}
}

func (c *Ctx) Send(chatID int64, text string) error {
	_, err := c.app.tg.SendMessage(c.ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: text})
	return err
}

func (c *Ctx) SetConv(name string)             { c.app.conv.Set(c.UserID(), c.ChatID(), name, nil) }
func (c *Ctx) SetConvState(name string, s any) { c.app.conv.Set(c.UserID(), c.ChatID(), name, s) }
func (c *Ctx) DelConv()                        { c.app.conv.Delete(c.UserID(), c.ChatID()) }
func (c *Ctx) conv() (conversation, bool)      { return c.app.conv.Get(c.UserID(), c.ChatID()) }

func (c *Ctx) IsAdmin() bool {
	return c.UserID() == c.app.adminID()
}

func (c *Ctx) IsManager() bool {
	return c.IsAdmin() || c.app.store.IsManager(c.UserID())
}

// EnsureAllowed replies with instructions when the chat has no permission.
func (c *Ctx) EnsureAllowed() bool {
	chat, ok := c.app.store.Chat(c.ChatID())
	if !ok {
		c.Replyf(texts.PermissionNeeded, c.cmd)
		return false
	}

	switch chat.Permission {
	case store.PermissionRequested:
		c.Replyf(texts.PermissionPending, c.cmd)
	case store.PermissionForbidden:
		c.Reply(texts.PermissionForbidden)
	}
	return chat.Permission == store.PermissionAllowed
}

func (c *Ctx) IsAllowedSilent() bool {
	chat, ok := c.app.store.Chat(c.ChatID())
	return ok && chat.Permission == store.PermissionAllowed
}
