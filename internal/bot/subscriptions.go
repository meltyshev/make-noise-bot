package bot

import (
	"fmt"
	"strings"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// soloMarker ends a callback opened by /subscribe, where there is no list to
// go back to.
const soloMarker = "solo"

type kindSpec struct {
	code    string
	label   string
	short   string
	content bool
	get     func(store.Subscription) bool
	set     func(*store.Subscription, bool)
}

// kinds are ordered notifications first, level content after.
var kinds = []kindSpec{
	{
		code: "l", label: texts.KindLevelUp, short: texts.KindShortLevelUp,
		get: func(s store.Subscription) bool { return s.LevelUp },
		set: func(s *store.Subscription, v bool) { s.LevelUp = v },
	},
	{
		code: "h", label: texts.KindHints, short: texts.KindShortHints,
		get: func(s store.Subscription) bool { return s.Hints },
		set: func(s *store.Subscription, v bool) { s.Hints = v },
	},
	{
		code: "s", label: texts.KindSpoilers, short: texts.KindShortSpoilers,
		get: func(s store.Subscription) bool { return s.Spoilers },
		set: func(s *store.Subscription, v bool) { s.Spoilers = v },
	},
	{
		code: "q", label: texts.KindQuestion, short: texts.KindShortQuestion, content: true,
		get: func(s store.Subscription) bool { return s.Question },
		set: func(s *store.Subscription, v bool) { s.Question = v },
	},
	{
		code: "n", label: texts.KindNotes, short: texts.KindShortNotes, content: true,
		get: func(s store.Subscription) bool { return s.Notes },
		set: func(s *store.Subscription, v bool) { s.Notes = v },
	},
}

func kindsSummary(sub store.Subscription) string {
	if sub.All() {
		return texts.KindsAll
	}
	var parts []string
	for _, kind := range kinds {
		if kind.get(sub) {
			parts = append(parts, kind.short)
		}
	}
	return strings.Join(parts, ", ")
}

func namespaceOf(forGame bool) string {
	if forGame {
		return "gs"
	}
	return "cs"
}

func subAction(forGame bool, action string, chatID int64) string {
	return fmt.Sprintf("%s:%s:%d", namespaceOf(forGame), action, chatID)
}

func subscriptionsOf(d *store.Data, forGame bool) store.Subscriptions {
	if forGame {
		if d.Game == nil {
			return nil
		}
		return d.Game.Subscriptions
	}
	return d.GameConfig.Subscriptions
}

func renderSubscribers(d *store.Data, forGame bool) (string, [][]models.InlineKeyboardButton) {
	title := texts.SubscribersTitle
	if forGame {
		title = texts.GameSubscribersTitle
	}
	subs := subscriptionsOf(d, forGame)

	ids := map[int64]bool{}
	for id, chat := range d.Chats {
		if chat.Permission == store.PermissionAllowed {
			ids[id] = true
		}
	}
	for _, sub := range subs {
		ids[sub.ChatID] = true
	}

	var keyboard [][]models.InlineKeyboardButton
	for _, id := range sortByName(d, ids) {
		label := d.DisplayName(id)
		if sub, ok := subs.Find(id); ok {
			label = mark(true, label+": "+kindsSummary(sub))
		}
		keyboard = append(keyboard, btnRow(btn(label, subAction(forGame, "d", id))))
	}
	keyboard = append(keyboard, btnRow(btn(texts.ButtonBack, "gc:menu")))
	return title, keyboard
}

func renderSubscriptionDetail(d *store.Data, forGame bool, chatID int64, solo bool) (string, [][]models.InlineKeyboardButton) {
	sub, _ := subscriptionsOf(d, forGame).Find(chatID)

	suffix := ""
	if solo {
		suffix = ":" + soloMarker
	}

	var notifications, content []models.InlineKeyboardButton
	for _, kind := range kinds {
		label := mark(kind.get(sub), kind.label)
		button := btn(label, subAction(forGame, "k", chatID)+":"+kind.code+suffix)
		if kind.content {
			content = append(content, button)
		} else {
			notifications = append(notifications, button)
		}
	}

	keyboard := [][]models.InlineKeyboardButton{
		btnRow(
			btn(texts.ButtonAllUpdates, subAction(forGame, "a", chatID)+suffix),
			btn(texts.ButtonNotificationsOnly, subAction(forGame, "v", chatID)+suffix),
		),
		notifications,
		content,
	}
	if sub.Any() {
		keyboard = append(keyboard, btnRow(btn(texts.ButtonUnsubscribe, subAction(forGame, "x", chatID)+suffix)))
	}
	if solo {
		keyboard = append(keyboard, btnRow(btn(texts.ButtonClose, namespaceOf(forGame)+":close")))
	} else {
		keyboard = append(keyboard, btnRow(btn(texts.ButtonBack, namespaceOf(forGame)+":list")))
	}

	return fmt.Sprintf(texts.SubscriptionTitleFmt, d.DisplayName(chatID)), keyboard
}

// updateSubscription edits one chat's subscription. Config changes mirror
// into the running game so they apply without a restart.
func (a *App) updateSubscription(forGame bool, chatID int64, mutate func(sub *store.Subscription)) error {
	return a.store.Update(func(d *store.Data) {
		if forGame {
			if d.Game == nil {
				return
			}
			sub, _ := d.Game.Subscriptions.Find(chatID)
			mutate(&sub)
			d.Game.Subscriptions = d.Game.Subscriptions.Set(sub)
			return
		}

		sub, _ := d.GameConfig.Subscriptions.Find(chatID)
		mutate(&sub)
		d.GameConfig.Subscriptions = d.GameConfig.Subscriptions.Set(sub)
		if d.Game != nil {
			d.Game.Subscriptions = d.Game.Subscriptions.Set(sub)
		}
	})
}

func (a *App) subscriptionsCallback(c *cb, forGame bool, args []string) {
	if len(args) == 0 {
		c.answer("")
		return
	}
	if forGame {
		if _, ok := a.store.Game(); !ok {
			c.answer(texts.NoActiveGame)
			return
		}
	}

	showList := func() {
		var (
			text     string
			keyboard [][]models.InlineKeyboardButton
		)
		a.store.View(func(d *store.Data) { text, keyboard = renderSubscribers(d, forGame) })
		c.edit(text, keyboard)
	}
	showDetail := func(chatID int64, solo bool) {
		var (
			text     string
			keyboard [][]models.InlineKeyboardButton
		)
		a.store.View(func(d *store.Data) { text, keyboard = renderSubscriptionDetail(d, forGame, chatID, solo) })
		c.edit(text, keyboard)
	}

	switch args[0] {
	case "list":
		c.answer("")
		showList()
		return
	case "close":
		c.answer("")
		c.edit(texts.Done, nil)
		return
	}

	chatID, ok := argID(args, 1)
	if !ok {
		c.answer("")
		return
	}
	rest := args[2:]
	solo := len(rest) > 0 && rest[len(rest)-1] == soloMarker

	apply := func(mutate func(sub *store.Subscription)) bool {
		if err := a.updateSubscription(forGame, chatID, mutate); err != nil {
			a.reportError(err)
			c.answer("")
			return false
		}
		return true
	}

	// The presets are decisive, so they leave the screen; ticking single
	// kinds keeps it open.
	finish := func() {
		c.answer("")
		if solo {
			c.edit(texts.Done, nil)
			return
		}
		showList()
	}

	switch args[0] {
	case "d":
		c.answer("")
		showDetail(chatID, solo)
	case "a":
		if apply(func(sub *store.Subscription) { *sub = store.AllUpdates(chatID) }) {
			finish()
		}
	case "v":
		if apply(func(sub *store.Subscription) { *sub = store.Notifications(chatID) }) {
			finish()
		}
	case "x":
		if apply(func(sub *store.Subscription) { *sub = store.Subscription{ChatID: chatID} }) {
			finish()
		}
	case "k":
		if len(rest) == 0 {
			c.answer("")
			return
		}
		kind, found := findKind(rest[0])
		if !found {
			c.answer("")
			return
		}
		if apply(func(sub *store.Subscription) { kind.set(sub, !kind.get(*sub)) }) {
			c.answer("")
			showDetail(chatID, solo)
		}
	default:
		c.answer("")
	}
}

func findKind(code string) (kindSpec, bool) {
	for _, kind := range kinds {
		if kind.code == code {
			return kind, true
		}
	}
	return kindSpec{}, false
}
