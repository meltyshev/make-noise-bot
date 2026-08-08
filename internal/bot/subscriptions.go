package bot

import (
	"fmt"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

// soloMarker ends a callback opened by /subscribe, where there is no list to
// go back to.
const soloMarker = "solo"

func modeSummary(sub store.Subscription) string {
	if sub.EventsOnly {
		return texts.SummaryEvents
	}
	return texts.SummaryAll
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
			label = mark(true, label+": "+modeSummary(sub))
		}
		keyboard = append(keyboard, btnRow(btn(label, subAction(forGame, "d", id))))
	}
	keyboard = append(keyboard, btnRow(btn(texts.ButtonBack, "gc:menu")))
	return title, keyboard
}

func renderSubscriptionDetail(d *store.Data, forGame bool, chatID int64, solo bool) (string, [][]models.InlineKeyboardButton) {
	sub, subscribed := subscriptionsOf(d, forGame).Find(chatID)

	suffix := ""
	if solo {
		suffix = ":" + soloMarker
	}

	keyboard := [][]models.InlineKeyboardButton{
		btnRow(
			btn(mark(subscribed && !sub.EventsOnly, texts.ButtonAllUpdates), subAction(forGame, "a", chatID)+suffix),
			btn(mark(subscribed && sub.EventsOnly, texts.ButtonEventsOnly), subAction(forGame, "v", chatID)+suffix),
		),
	}
	if subscribed {
		keyboard = append(keyboard, btnRow(btn(texts.ButtonUnsubscribe, subAction(forGame, "x", chatID)+suffix)))
	}
	if solo {
		keyboard = append(keyboard, btnRow(btn(texts.ButtonClose, namespaceOf(forGame)+":close")))
	} else {
		keyboard = append(keyboard, btnRow(btn(texts.ButtonBack, namespaceOf(forGame)+":list")))
	}

	return fmt.Sprintf(texts.SubscriptionTitleFmt, d.DisplayName(chatID)), keyboard
}

// setSubscription stores a chat's mode, or removes the chat. Config changes
// mirror into the running game so they apply without a restart.
func (a *App) setSubscription(forGame bool, sub store.Subscription, subscribed bool) error {
	apply := func(list store.Subscriptions) store.Subscriptions {
		if subscribed {
			return list.Set(sub)
		}
		return list.Remove(sub.ChatID)
	}

	return a.store.Update(func(d *store.Data) {
		if forGame {
			if d.Game != nil {
				d.Game.Subscriptions = apply(d.Game.Subscriptions)
			}
			return
		}

		d.GameConfig.Subscriptions = apply(d.GameConfig.Subscriptions)
		if d.Game != nil {
			d.Game.Subscriptions = apply(d.Game.Subscriptions)
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
		c.show(func(d *store.Data) (string, [][]models.InlineKeyboardButton) {
			return renderSubscribers(d, forGame)
		})
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

	if args[0] == "d" {
		c.answer("")
		c.show(func(d *store.Data) (string, [][]models.InlineKeyboardButton) {
			return renderSubscriptionDetail(d, forGame, chatID, solo)
		})
		return
	}

	var (
		sub        store.Subscription
		subscribed = true
	)
	switch args[0] {
	case "a":
		sub = store.AllUpdates(chatID)
	case "v":
		sub = store.OnlyEvents(chatID)
	case "x":
		sub, subscribed = store.Subscription{ChatID: chatID}, false
	default:
		c.answer("")
		return
	}

	if err := a.setSubscription(forGame, sub, subscribed); err != nil {
		a.reportError(err)
		c.answer("")
		return
	}

	// Every choice here is decisive, so it leaves the screen.
	c.answer("")
	if solo {
		c.edit(texts.Done, nil)
		return
	}
	showList()
}
