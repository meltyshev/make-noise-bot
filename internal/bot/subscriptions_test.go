package bot

import (
	"strings"
	"testing"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

func TestModeSummary(t *testing.T) {
	if got := modeSummary(store.AllUpdates(1)); got != texts.SummaryAll {
		t.Errorf("modeSummary(AllUpdates) = %q, want %q", got, texts.SummaryAll)
	}
	if got := modeSummary(store.OnlyEvents(1)); got != texts.SummaryEvents {
		t.Errorf("modeSummary(OnlyEvents) = %q, want %q", got, texts.SummaryEvents)
	}
}

func TestRenderSubscribersList(t *testing.T) {
	d := testData()
	d.GameConfig.Subscriptions = store.Subscriptions{
		store.AllUpdates(-100),
		store.OnlyEvents(555),
	}

	_, keyboard := renderSubscribers(d, false)

	// Rows open the per-chat screen instead of switching a mode at once.
	if b := findButton(t, keyboard, "✓ Команда: всё"); b.CallbackData != "cs:d:-100" {
		t.Errorf("subscribed chat callback = %q, want cs:d:-100", b.CallbackData)
	}
	if b := findButton(t, keyboard, "✓ ID 555: события"); b.CallbackData != "cs:d:555" {
		t.Errorf("events-only chat callback = %q, want cs:d:555", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Вася"); b.CallbackData != "cs:d:10" {
		t.Errorf("candidate callback = %q, want cs:d:10", b.CallbackData)
	}

	for _, row := range keyboard {
		if len(row) != 1 {
			t.Errorf("subscriber row = %d buttons, want 1", len(row))
		}
	}

	for _, button := range flatButtons(keyboard) {
		if strings.Contains(button.Text, "Чужие") || strings.Contains(button.Text, "Аня") {
			t.Errorf("subscriber candidates include %q, want only allowed chats", button.Text)
		}
	}
}

func TestRenderSubscribersGameList(t *testing.T) {
	d := testData()
	d.Game = &store.Game{Engine: "DozorClassic", Subscriptions: store.Subscriptions{store.AllUpdates(-100)}}

	_, keyboard := renderSubscribers(d, true)
	if b := findButton(t, keyboard, "✓ Команда: всё"); b.CallbackData != "gs:d:-100" {
		t.Errorf("game row callback = %q, want gs:d:-100", b.CallbackData)
	}
}

func TestRenderSubscriptionDetail(t *testing.T) {
	d := testData()
	d.GameConfig.Subscriptions = store.Subscriptions{store.OnlyEvents(-100)}

	text, keyboard := renderSubscriptionDetail(d, false, -100, false)
	if !strings.Contains(text, "Команда") {
		t.Errorf("renderSubscriptionDetail title = %q, want it to name the chat", text)
	}

	// The current mode is marked, the other one is not.
	if b := findButton(t, keyboard, texts.ButtonEventsOnly); b.CallbackData != "cs:v:-100" || !strings.HasPrefix(b.Text, "✓") {
		t.Errorf("events button = %+v, want marked", b)
	}
	if b := findButton(t, keyboard, texts.ButtonAllUpdates); b.CallbackData != "cs:a:-100" || strings.HasPrefix(b.Text, "✓") {
		t.Errorf("all button = %+v, want unmarked", b)
	}
	if b := findButton(t, keyboard, texts.ButtonUnsubscribe); b.CallbackData != "cs:x:-100" {
		t.Errorf("unsubscribe callback = %q, want cs:x:-100", b.CallbackData)
	}
	if b := findButton(t, keyboard, texts.ButtonBack); b.CallbackData != "cs:list" {
		t.Errorf("back callback = %q, want cs:list", b.CallbackData)
	}
}

func TestRenderSubscriptionDetailUnsubscribed(t *testing.T) {
	d := testData()

	_, keyboard := renderSubscriptionDetail(d, false, -100, false)
	for _, button := range flatButtons(keyboard) {
		if button.Text == texts.ButtonUnsubscribe {
			t.Error("renderSubscriptionDetail offers unsubscribe, want it hidden for a chat with no subscription")
		}
		if strings.HasPrefix(button.Text, "✓") {
			t.Errorf("mode label = %q, want no checkmark for a chat with no subscription", button.Text)
		}
	}
}

func TestRenderSubscriptionDetailSolo(t *testing.T) {
	d := testData()
	d.Game = &store.Game{Engine: "DozorClassic", Subscriptions: store.Subscriptions{store.AllUpdates(-100)}}

	_, keyboard := renderSubscriptionDetail(d, true, -100, true)

	// Solo screens carry the marker so edits keep rendering standalone.
	if b := findButton(t, keyboard, texts.ButtonAllUpdates); b.CallbackData != "gs:a:-100:solo" {
		t.Errorf("all button callback = %q, want gs:a:-100:solo", b.CallbackData)
	}
	if b := findButton(t, keyboard, texts.ButtonClose); b.CallbackData != "gs:close" {
		t.Errorf("close callback = %q, want gs:close", b.CallbackData)
	}
}
