package bot

import (
	"strings"
	"testing"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

func TestKindsSummary(t *testing.T) {
	if got := kindsSummary(store.AllUpdates(1)); got != texts.KindsAll {
		t.Errorf("all = %q", got)
	}
	if got := kindsSummary(store.Notifications(1)); got != "АП, подсказки, спойлеры" {
		t.Errorf("notifications = %q", got)
	}
	if got := kindsSummary(store.Subscription{ChatID: 1, Question: true}); got != "задание" {
		t.Errorf("question only = %q", got)
	}
}

func TestRenderSubscribersList(t *testing.T) {
	d := testData()
	d.GameConfig.Subscriptions = store.Subscriptions{
		store.AllUpdates(-100),
		store.Notifications(555),
	}

	_, keyboard := renderSubscribers(d, false)

	// Rows open the per-chat screen instead of toggling everything at once.
	if b := findButton(t, keyboard, "✓ Команда: всё"); b.CallbackData != "cs:d:-100" {
		t.Errorf("subscribed chat = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, "✓ ID 555: АП, подсказки, спойлеры"); b.CallbackData != "cs:d:555" {
		t.Errorf("unknown subscriber = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Вася"); b.CallbackData != "cs:d:10" {
		t.Errorf("candidate = %q", b.CallbackData)
	}

	for _, row := range keyboard {
		if len(row) != 1 {
			t.Errorf("row has %d buttons, want a single one", len(row))
		}
	}

	for _, button := range flatButtons(keyboard) {
		if strings.Contains(button.Text, "Чужие") || strings.Contains(button.Text, "Аня") {
			t.Errorf("not-allowed chat offered: %q", button.Text)
		}
	}
}

func TestRenderSubscribersGameList(t *testing.T) {
	d := testData()
	d.Game = &store.Game{Engine: "DozorClassic", Subscriptions: store.Subscriptions{store.AllUpdates(-100)}}

	_, keyboard := renderSubscribers(d, true)
	if b := findButton(t, keyboard, "✓ Команда: всё"); b.CallbackData != "gs:d:-100" {
		t.Errorf("game row = %q", b.CallbackData)
	}
}

func TestRenderSubscriptionDetail(t *testing.T) {
	d := testData()
	d.GameConfig.Subscriptions = store.Subscriptions{store.Notifications(-100)}

	text, keyboard := renderSubscriptionDetail(d, false, -100, false)
	if !strings.Contains(text, "Команда") {
		t.Errorf("title = %q", text)
	}

	if b := findButton(t, keyboard, mark(true, texts.KindLevelUp)); b.CallbackData != "cs:k:-100:l" {
		t.Errorf("level toggle = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, texts.KindQuestion); b.CallbackData != "cs:k:-100:q" {
		t.Errorf("question toggle = %q", b.CallbackData)
	}
	if strings.Contains(findButton(t, keyboard, texts.KindQuestion).Text, "✓") {
		t.Error("question must be off for a notifications-only subscription")
	}
	if b := findButton(t, keyboard, texts.ButtonAllUpdates); b.CallbackData != "cs:a:-100" {
		t.Errorf("all = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, texts.ButtonUnsubscribe); b.CallbackData != "cs:x:-100" {
		t.Errorf("unsubscribe = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, texts.ButtonBack); b.CallbackData != "cs:list" {
		t.Errorf("back = %q", b.CallbackData)
	}
}

func TestRenderSubscriptionDetailSolo(t *testing.T) {
	d := testData()
	d.Game = &store.Game{Engine: "DozorClassic"}

	_, keyboard := renderSubscriptionDetail(d, true, -100, true)

	// Solo screens carry the marker so edits keep rendering standalone.
	if b := findButton(t, keyboard, texts.KindSpoilers); b.CallbackData != "gs:k:-100:s:solo" {
		t.Errorf("spoilers toggle = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, texts.ButtonClose); b.CallbackData != "gs:close" {
		t.Errorf("close = %q", b.CallbackData)
	}
	// Nothing to unsubscribe from yet.
	for _, button := range flatButtons(keyboard) {
		if button.Text == texts.ButtonUnsubscribe {
			t.Error("unsubscribe offered for a chat with no subscription")
		}
	}
}

func TestSoloMarkerDoesNotClashWithSpoilerKind(t *testing.T) {
	// The spoilers kind code is "s"; the solo marker must stay distinct.
	kind, found := findKind("s")
	if !found || kind.label != texts.KindSpoilers {
		t.Fatalf("kind s = %+v, found=%v", kind, found)
	}
	if soloMarker == "s" {
		t.Error("solo marker collides with the spoilers kind code")
	}
}

func TestFindKind(t *testing.T) {
	for _, code := range []string{"l", "h", "s", "q", "n"} {
		if _, found := findKind(code); !found {
			t.Errorf("kind %q missing", code)
		}
	}
	if _, found := findKind("z"); found {
		t.Error("unknown kind accepted")
	}
}
