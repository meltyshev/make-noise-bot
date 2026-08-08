package bot

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// render builds a menu from one consistent view of the state.
type render func(d *store.Data) (string, [][]models.InlineKeyboardButton)

// cb is the context of one inline button click.
type cb struct {
	ctx    context.Context
	app    *App
	query  *models.CallbackQuery
	chatID int64
	msgID  int
}

func (a *App) onCallback(ctx context.Context, query *models.CallbackQuery) {
	defer a.recoverPanic()

	msg := query.Message.Message
	if msg == nil {
		a.answerCallback(ctx, query.ID, texts.AlreadyProcessed)
		return
	}

	c := &cb{ctx: ctx, app: a, query: query, chatID: msg.Chat.ID, msgID: msg.ID}

	parts := strings.Split(query.Data, ":")
	namespace, args := parts[0], parts[1:]

	switch namespace {
	case "perm", "cfg", "ch":
		if query.From.ID != a.adminID() {
			c.answer(texts.NoAccess)
			return
		}
	case "gc", "gs", "cs", "res", "gm":
		if query.From.ID != a.adminID() && !a.store.IsManager(query.From.ID) {
			c.answer(texts.NoAccess)
			return
		}
	}

	switch namespace {
	case "perm":
		a.permCallback(c, args)
	case "cfg":
		a.configCallback(c, args)
	case "ch":
		a.chatsCallback(c, args)
	case "gc":
		a.gameConfigCallback(c, args)
	case "cs":
		a.subscriptionsCallback(c, false, args)
	case "gs":
		a.subscriptionsCallback(c, true, args)
	case "res":
		a.restrictCallback(c)
	case "gm":
		a.stopGameCallback(c)
	default:
		c.answer("")
	}
}

func (c *cb) answer(toast string) {
	c.app.answerCallback(c.ctx, c.query.ID, toast)
}

func (a *App) answerCallback(ctx context.Context, id, toast string) {
	_, err := a.tg.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: id,
		Text:            toast,
	})
	if err != nil {
		a.log.Warn("answer callback failed", "error", err)
	}
}

// edit replaces the menu message the click came from. A nil keyboard removes
// the buttons.
func (c *cb) edit(text string, keyboard [][]models.InlineKeyboardButton) {
	c.app.editMessage(c.ctx, c.chatID, c.msgID, text, keyboard)
}

// show redraws the menu message the click came from.
func (c *cb) show(r render) {
	c.edit(c.app.renderMenu(r))
}

func (a *App) renderMenu(r render) (string, [][]models.InlineKeyboardButton) {
	var (
		text     string
		keyboard [][]models.InlineKeyboardButton
	)
	a.store.View(func(d *store.Data) { text, keyboard = r(d) })
	return text, keyboard
}

// editMenu redraws a menu message the current update did not come from, which
// is how a conversation reply refreshes the menu that started it.
func (a *App) editMenu(ctx context.Context, chatID int64, msgID int, r render) {
	text, keyboard := a.renderMenu(r)
	a.editMessage(ctx, chatID, msgID, text, keyboard)
}

func (a *App) editMessage(ctx context.Context, chatID int64, msgID int, text string, keyboard [][]models.InlineKeyboardButton) {
	params := &tgbot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: text}
	if keyboard != nil {
		params.ReplyMarkup = &models.InlineKeyboardMarkup{InlineKeyboard: keyboard}
	}
	if _, err := a.tg.EditMessageText(ctx, params); err != nil {
		if !strings.Contains(err.Error(), "message is not modified") {
			a.reportError(fmt.Errorf("edit message: %w", err))
		}
	}
}

func (a *App) send(ctx context.Context, chatID int64, text string) error {
	_, err := a.tg.SendMessage(ctx, &tgbot.SendMessageParams{ChatID: chatID, Text: text})
	return err
}

func (a *App) sendInline(ctx context.Context, chatID int64, text string, keyboard [][]models.InlineKeyboardButton) error {
	_, err := a.tg.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	})
	return err
}

func argID(args []string, idx int) (int64, bool) {
	if idx >= len(args) {
		return 0, false
	}
	id, err := strconv.ParseInt(args[idx], 10, 64)
	return id, err == nil
}

func btn(label, data string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{Text: truncateLabel(label), CallbackData: data}
}

// mark prefixes a label when the option is active. Every menu shows state
// the same way, and the mark survives label truncation.
func mark(active bool, label string) string {
	if active {
		return "✓ " + label
	}
	return label
}

func btnRow(buttons ...models.InlineKeyboardButton) []models.InlineKeyboardButton {
	return buttons
}

func truncateLabel(label string) string {
	const maxRunes = 48
	runes := []rune(label)
	if len(runes) <= maxRunes {
		return label
	}
	return string(runes[:maxRunes-3]) + "..."
}

// sortByName orders the ids of a menu list the way they are labelled, so the
// buttons do not jump around between renders.
func sortByName(d *store.Data, ids map[int64]bool) []int64 {
	sorted := slices.Collect(maps.Keys(ids))
	slices.SortFunc(sorted, func(a, b int64) int {
		if n := strings.Compare(d.DisplayName(a), d.DisplayName(b)); n != 0 {
			return n
		}
		return cmp.Compare(a, b)
	})
	return sorted
}
