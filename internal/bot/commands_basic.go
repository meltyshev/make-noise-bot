package bot

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/meltyshev/make-noise-bot/internal/texts"
)

func cmdStart() *Command {
	return &Command{
		Name: "start",
		Init: func(c *Ctx, _ string) {
			firstName := ""
			if c.msg.From != nil {
				firstName = c.msg.From.FirstName
			}
			c.Replyf(texts.Start, firstName)
		},
	}
}

func cmdHelp() *Command {
	return &Command{
		Name:        "help",
		Description: texts.DescHelp,
		Init: func(c *Ctx, _ string) {
			var lines []string
			for _, cmd := range c.app.order {
				if cmd.Description != "" {
					lines = append(lines, fmt.Sprintf("/%s - %s", cmd.Name, cmd.Description))
				}
			}
			c.Reply(strings.Join(lines, "\n"))
		},
	}
}

func cmdCancel() *Command {
	return &Command{
		Name:        "cancel",
		Description: texts.DescCancel,
		Init: func(c *Ctx, _ string) {
			if conv, ok := c.conv(); ok {
				c.DelConv()
				c.Replyf(texts.CancelDone, conv.Name)
			} else {
				c.Reply(texts.CancelNothing)
			}
		},
	}
}

func cmdChatID() *Command {
	return &Command{
		Name: "chatid",
		Init: func(c *Ctx, _ string) {
			c.Reply(strconv.FormatInt(c.ChatID(), 10))
		},
	}
}

func cmdUserID() *Command {
	return &Command{
		Name: "userid",
		Init: func(c *Ctx, _ string) {
			c.Reply(strconv.FormatInt(c.UserID(), 10))
		},
	}
}

func cmdTop() *Command {
	return &Command{
		Name: "top",
		Init: func(c *Ctx, _ string) {
			c.Reply(texts.TopAll)
		},
	}
}

func cmdMaxwell() *Command {
	return &Command{
		Name: "maxwell",
		Init: func(c *Ctx, _ string) {
			c.Reply(pickPhrase(texts.MaxwellPhrases))
		},
	}
}

func cmdRomka() *Command {
	return &Command{
		Name: "romka",
		Init: func(c *Ctx, _ string) {
			c.Reply(pickPhrase(texts.RomkaPhrases))
		},
	}
}

func pickPhrase(phrases []string) string {
	return phrases[rand.IntN(len(phrases))]
}
