// Package tgsend delivers HTML messages: long texts are split into balanced
// parts, and markup Telegram rejects is resent as a code block.
package tgsend

import (
	"context"
	"errors"
	"html"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/htmltext"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// Telegram's message limit is 4096 UTF-16 units; keep a margin.
const limit = 4000

// HTML sends a converted fragment, linking coordinates with mapLink and
// splitting anything over the length limit.
func HTML(ctx context.Context, b *tgbot.Bot, chatID int64, text string, mapLink func(lat, lon float64) string, reply *models.ReplyParameters) error {
	text = htmltext.LinkCoordinates(text, mapLink)

	for i, part := range htmltext.Split(text, limit) {
		params := &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      part,
			ParseMode: models.ParseModeHTML,
		}
		if i == 0 {
			params.ReplyParameters = reply
		}

		if _, err := b.SendMessage(ctx, params); err != nil {
			if !errors.Is(err, tgbot.ErrorBadRequest) {
				return err
			}
			if err := sendUnparsed(ctx, b, chatID, part); err != nil {
				return err
			}
		}
	}
	return nil
}

func sendUnparsed(ctx context.Context, b *tgbot.Bot, chatID int64, part string) error {
	fallback := texts.HTMLFallback + "\n<pre>" + html.EscapeString(part) + "</pre>"
	for _, piece := range htmltext.Split(fallback, limit) {
		_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      piece,
			ParseMode: models.ParseModeHTML,
		})
		if err == nil {
			continue
		}
		if _, plainErr := b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   htmltext.StripTags(piece),
		}); plainErr != nil {
			return plainErr
		}
	}
	return nil
}
