package bot

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// onUpdate dispatches in a fixed order: commands, an open conversation,
// leave mode, then game statements.
func (a *App) onUpdate(ctx context.Context, _ *tgbot.Bot, update *models.Update) {
	defer a.recoverPanic()

	if update.CallbackQuery != nil {
		a.onCallback(ctx, update.CallbackQuery)
		return
	}

	msg := update.Message
	if msg == nil {
		return
	}

	c := &Ctx{ctx: ctx, app: a, msg: msg}
	text := msg.Text

	// Until an admin is configured, the first /start in a private chat
	// claims the role.
	if a.adminID() == 0 {
		if msg.From != nil && msg.Chat.Type == models.ChatTypePrivate && a.stripBotMention(commandName(text)) == "start" {
			if err := a.claimAdmin(msg.From.ID); err != nil {
				a.reportError(fmt.Errorf("claim admin: %w", err))
				return
			}
			c.Replyf(texts.AdminClaimed, msg.From.FirstName)
		}
		return
	}

	// Unknown commands are swallowed so they are never submitted as codes.
	if strings.HasPrefix(text, "/") && len(text) > 1 {
		name, args := splitCommand(text)
		name = a.stripBotMention(name)

		if cmd, ok := a.commands[name]; ok {
			cmd.Init(c, args)
		}
		return
	}

	if msg.From == nil {
		return
	}

	if conv, ok := c.Conv(); ok {
		if cmd, found := a.commands[conv.Name]; found && cmd.Handle != nil {
			cmd.Handle(c, conv.State)
			return
		}
	}

	if c.isGroup() && text != "" && a.store.LeaveMode() {
		a.leaveChat(c)
		return
	}

	switch {
	case text == "?":
		a.information(c)
	case strings.HasPrefix(text, "$"):
		a.enterPinnedCode(c)
	case text != "":
		a.enterCode(c)
	}
}

// splitCommand splits "/name args" on any whitespace.
func splitCommand(text string) (name, args string) {
	s := text[1:]
	idx := strings.IndexFunc(s, unicode.IsSpace)
	if idx < 0 {
		return s, ""
	}
	return s[:idx], strings.TrimLeftFunc(s[idx:], unicode.IsSpace)
}

// stripBotMention strips "@this_bot"; other bots' mentions stay, so their
// commands remain unknown.
func (a *App) stripBotMention(name string) string {
	if a.me == nil {
		return name
	}
	if idx := strings.LastIndex(name, "@"); idx >= 0 && strings.EqualFold(name[idx+1:], a.me.Username) {
		return name[:idx]
	}
	return name
}

func commandName(text string) string {
	if !strings.HasPrefix(text, "/") || len(text) < 2 {
		return ""
	}
	name, _ := splitCommand(text)
	return name
}
