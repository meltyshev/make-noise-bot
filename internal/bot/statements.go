package bot

import (
	"fmt"
	"html"
	"strings"

	tgbot "github.com/go-telegram/bot"

	"github.com/meltyshev/make-noise-bot/internal/game"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// information answers "?" with the sector board and progress.
func (a *App) information(c *Ctx) {
	if !c.IsAllowedSilent() {
		return
	}

	engine := a.Engine()
	if engine == nil {
		return
	}

	snap, err := engine.Load(c.ctx)
	if err != nil {
		c.Reply(texts.CannotLoadEngine)
		return
	}

	var messages []string
	if sectors := snap.Sectors(); sectors != nil {
		var sectorBlocks []string
		for _, sector := range sectors {
			sectorBlocks = append(sectorBlocks, formatSector(sector))
		}
		messages = append(messages, strings.Join(sectorBlocks, "\n\n"))
	}
	if progress := snap.Progress(); progress != "" {
		messages = append(messages, progress)
	}

	if len(messages) == 0 {
		c.Reply(texts.InfoNone)
		return
	}
	c.ReplyHTML("<pre>" + html.EscapeString(strings.Join(messages, "\n\n")) + "</pre>")
}

// formatSector renders the two-column code table.
func formatSector(sector game.Sector) string {
	const cellFormat = "%3s %-2s %-2s"

	cell := func(code game.SectorCode) []any {
		entered := ""
		if code.Entered {
			entered = "ok"
		}
		return []any{fmt.Sprintf("%d)", code.Number), code.Hazard, entered}
	}

	codes := sector.Codes
	balance := len(codes) % 2
	perColumn := len(codes) / 2

	column1 := codes[:perColumn]
	column2 := codes[perColumn+balance:]

	lines := []string{sector.Name + ":"}
	for i := range column1 {
		args := append(cell(column1[i]), cell(column2[i])...)
		lines = append(lines, fmt.Sprintf(cellFormat+"   "+cellFormat, args...))
	}
	if balance == 1 {
		lines = append(lines, fmt.Sprintf(cellFormat, cell(codes[perColumn])...))
	}

	return strings.Join(lines, "\n")
}

// enterPinnedCode handles "$code": codes for the pinned level.
func (a *App) enterPinnedCode(c *Ctx) {
	if !c.IsAllowedSilent() {
		return
	}

	engine := a.ClassicEngine()
	if engine == nil {
		return
	}

	g, ok := a.store.Game()
	if !ok || g.PinnedLevel == nil {
		return
	}

	code := strings.ToLower(c.Text()[1:])
	result := engine.EnterCode(c.ctx, code, g.PinnedLevel)
	c.ReplyAlways(result.Message)
}

// enterCode treats every unclaimed text message in an allowed chat as a
// potential game code.
func (a *App) enterCode(c *Ctx) {
	if !c.IsAllowedSilent() {
		return
	}

	g, ok := a.store.Game()
	if !ok || g.Restricted {
		return
	}

	engine := game.New(g, a.env)
	if engine == nil {
		return
	}

	chat, ok := a.store.Chat(c.ChatID())
	if !ok {
		return
	}

	code := c.Text()
	if !chat.BruteForce {
		prepared, isCode := game.PrepareCode(code, g.CodeFormats)
		if !isCode {
			return
		}
		code = prepared
	}

	result := engine.EnterCode(c.ctx, code, nil)

	if result.Accepted {
		if err := a.store.IncrementPlayer(c.UserID(), c.UserName()); err != nil {
			a.reportError(err)
		}
	}

	// In brute-force mode only accepted codes get an answer.
	if chat.BruteForce && !result.Accepted {
		return
	}

	c.ReplyAlways(result.Message)
}

func (a *App) leaveChat(c *Ctx) {
	if _, err := a.tg.LeaveChat(c.ctx, &tgbot.LeaveChatParams{ChatID: c.ChatID()}); err != nil {
		a.reportError(fmt.Errorf("leave chat: %w", err))
		return
	}
	if a.adminID() != 0 {
		text := fmt.Sprintf(texts.LeftChatFmt, c.msg.Chat.Title, c.ChatID())
		if _, err := a.tg.SendMessage(c.ctx, &tgbot.SendMessageParams{ChatID: a.adminID(), Text: text}); err != nil {
			a.log.Error("send to admin failed", "error", err)
		}
	}
}
