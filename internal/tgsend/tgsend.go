// Package tgsend delivers HTML messages: long texts are split into balanced
// parts, and markup Telegram rejects is resent as a code block.
package tgsend

import (
	"context"
	"errors"
	"html"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/htmltext"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// Telegram's message limit is 4096 UTF-16 units; keep a margin.
const limit = 4000

type Message struct {
	ChatID  int64
	Text    string
	MapLink func(lat, lon float64) string
	Reply   *models.ReplyParameters
	Markup  models.ReplyMarkup
}

// HTML sends a converted fragment, linking coordinates and splitting
// anything over the length limit. The reply lands on the first part and the
// buttons on the last one.
func HTML(ctx context.Context, b *tgbot.Bot, msg Message) error {
	// The engine's own links are collected first, so the coordinate links
	// added below never become the preview.
	own := htmltext.Links(msg.Text)
	text := htmltext.LinkCoordinates(msg.Text, msg.MapLink)

	parts := htmltext.Split(text, limit)
	for i, part := range parts {
		params := &tgbot.SendMessageParams{
			ChatID:             msg.ChatID,
			Text:               part,
			ParseMode:          models.ParseModeHTML,
			LinkPreviewOptions: preview(part, own),
		}
		if i == 0 {
			params.ReplyParameters = msg.Reply
		}
		if i == len(parts)-1 {
			params.ReplyMarkup = msg.Markup
		}

		if _, err := b.SendMessage(ctx, params); err != nil {
			if !errors.Is(err, tgbot.ErrorBadRequest) {
				return err
			}
			if err := sendUnparsed(ctx, b, msg.ChatID, part); err != nil {
				return err
			}
		}
	}
	return nil
}

// preview points the link preview at the first of the engine's own links in
// the part, and turns it off when the part has none.
func preview(part string, own []string) *models.LinkPreviewOptions {
	plain := html.UnescapeString(part)
	for _, link := range own {
		if strings.Contains(plain, link) {
			target := link
			return &models.LinkPreviewOptions{URL: &target}
		}
	}

	disabled := true
	return &models.LinkPreviewOptions{IsDisabled: &disabled}
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
