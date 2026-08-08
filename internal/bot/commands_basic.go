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

var maxwellPhrases = []string{
	"Внутри тебя не будет пустоты, если ты шаурма",
	"Не имей 100 рублей, а имей 100 рецептов шаурмы",
	"Не откладывай на завтра ту шаурму, которую можно съесть сегодня",
	"Сколько волка шаурмой не корми, он все равно в лес смотрит",
	"А Васька слушает да ест (шаурму)",
	"В гостях хорошо, а в шаурмечной лучше",
	"Век живи, век шаурму ешь",
	"Ты на кого тут шаурму крошишь?",
	"Шаурмей, шаурмей - кто успел, тот и съел",
	"Шаурма всему голова",
	"Шаурма человека кормит, а дозор портит",
	"Шаурма человеку друг",
	"Шаурме - время, дозору - час",
	"Язык до шаурмячной доведет",
	"Голод не шаурма, вообще только шаурма - шаурма",
	"На чужую шаурму рот не разевай",
	"Готовь сани летом, а ингредиенты для шаурмы - постоянно",
	"Съел свою шаурму - помоги соседу",
	"Глаза боятся, а руки готовят шаурму",
	"Любишь жрать, люби и шаурму готовить",
	"Под лежачую шаурму соус не течет",
	"2 шаурмы - пара",
	"Шаурма не волк - в лес не убежит",
	"Глаза боятся, а руки шаурму делают",
	"После поедания шаурмы грязными кулаками не машут",
	"В чужую шаурмячную со своей шаурмой не ходят",
	"Семь раз съешь - один закажи",
}

func cmdMaxwell() *Command {
	return &Command{
		Name: "maxwell",
		Init: func(c *Ctx, _ string) {
			c.Reply(maxwellPhrases[rand.IntN(len(maxwellPhrases))])
		},
	}
}

var romkaPhrases = []string{
	"Укропчика не желаете?",
	"Развооорооот..",
	"А метку «DR не светить» тоже снимать?",
	"Привет, хе-хе:)",
	"Здорова, ёптыть!",
	"ПЕРВЫЙ КОД ЗА ИГРУ!",
	"Причина остановки? Не выходи из машины!",
	"Через изоплит быстрее! Я по Яндексу пробил..",
	"По-брааатски, включи бутырку!",
}

func cmdRomka() *Command {
	return &Command{
		Name: "romka",
		Init: func(c *Ctx, _ string) {
			c.Reply(romkaPhrases[rand.IntN(len(romkaPhrases))])
		},
	}
}
