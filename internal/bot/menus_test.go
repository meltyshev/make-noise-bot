package bot

import (
	"fmt"
	"strings"
	"testing"

	"github.com/meltyshev/make-noise-bot/internal/geo"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

func TestPermRequestKeyboard(t *testing.T) {
	keyboard := permRequestKeyboard(-100123)
	buttons := flatButtons(keyboard)
	if len(buttons) != 2 {
		t.Fatalf("buttons = %d, want 2", len(buttons))
	}
	if buttons[0].CallbackData != "perm:a:-100123" || buttons[1].CallbackData != "perm:f:-100123" {
		t.Errorf("callbacks = %q, %q", buttons[0].CallbackData, buttons[1].CallbackData)
	}
}

func TestRenderConfigMenu(t *testing.T) {
	d := testData()
	d.LeaveMode = true

	_, keyboard := renderConfigMenu(d)
	if b := findButton(t, keyboard, "Менеджеры: 1"); b.CallbackData != "cfg:mgr" {
		t.Errorf("managers callback = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, "включен"); b.CallbackData != "cfg:leave" {
		t.Errorf("leave callback = %q", b.CallbackData)
	}
	// An empty setting shows the default service.
	label := fmt.Sprintf(texts.MapServiceFmt, geo.DefaultService.Label())
	if b := findButton(t, keyboard, label); b.CallbackData != "cfg:maps" {
		t.Errorf("maps callback = %q", b.CallbackData)
	}
	if len([]rune(label)) > 48 {
		t.Errorf("map row does not fit a button: %q", label)
	}
}

func TestRenderMapServices(t *testing.T) {
	d := testData()
	d.MapService = string(geo.Google)

	_, keyboard := renderMapServices(d)
	if b := findButton(t, keyboard, mark(true, geo.Google.Label())); b.CallbackData != "cfg:setmap:google" {
		t.Errorf("active service = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, geo.OSM.Label()); b.CallbackData != "cfg:setmap:osm" {
		t.Errorf("other service = %q", b.CallbackData)
	}
	for _, button := range flatButtons(keyboard) {
		if button.CallbackData == "cfg:setmap:yandex" && strings.HasPrefix(button.Text, "✓") {
			t.Errorf("inactive service marked: %q", button.Text)
		}
	}
}

func TestRenderManagers(t *testing.T) {
	_, keyboard := renderManagers(testData())

	if b := findButton(t, keyboard, "✓ Вася"); b.CallbackData != "cfg:mgrt:10" {
		t.Errorf("manager toggle = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Дима"); b.CallbackData != "cfg:mgrt:40" {
		t.Errorf("candidate toggle = %q", b.CallbackData)
	}
	findButton(t, keyboard, "Добавить")

	// Only allowed private chats are candidates.
	for _, button := range flatButtons(keyboard) {
		for _, hidden := range []string{"Аня", "Гоша", "Команда", "Чужие"} {
			if strings.Contains(button.Text, hidden) {
				t.Errorf("%s offered as manager: %q", hidden, button.Text)
			}
		}
	}
}

func TestRenderManagersKeepsExistingManagers(t *testing.T) {
	d := testData()
	// A manager whose chat was later forbidden must stay removable.
	d.Managers = append(d.Managers, 50)

	_, keyboard := renderManagers(d)
	if b := findButton(t, keyboard, "✓ Гоша"); b.CallbackData != "cfg:mgrt:50" {
		t.Errorf("forbidden manager toggle = %q", b.CallbackData)
	}
}

func TestRenderManagersUnknownID(t *testing.T) {
	d := testData()
	d.Managers = append(d.Managers, 999)

	_, keyboard := renderManagers(d)
	if b := findButton(t, keyboard, "✓ ID 999"); b.CallbackData != "cfg:mgrt:999" {
		t.Errorf("unknown manager toggle = %q", b.CallbackData)
	}
}

func TestRenderChatsList(t *testing.T) {
	_, keyboard := renderChatsList(testData())

	if b := findButton(t, keyboard, "✅ Команда"); b.CallbackData != "ch:open:-100" {
		t.Errorf("allowed chat = %q", b.CallbackData)
	}
	findButton(t, keyboard, "🚫 Чужие")
	findButton(t, keyboard, "❓ Аня")
}

func TestRenderChatActions(t *testing.T) {
	chat := store.Chat{ID: -100, Type: "supergroup", Permission: store.PermissionAllowed, Title: "Команда"}
	text, keyboard := renderChatActions(chat)

	if !strings.Contains(text, "Команда | -100") || !strings.Contains(text, "allowed") {
		t.Errorf("actions text = %q", text)
	}
	if b := findButton(t, keyboard, "Разрешить"); b.CallbackData != "ch:allow:-100" {
		t.Errorf("allow = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Удалить"); b.CallbackData != "ch:del:-100" {
		t.Errorf("delete = %q", b.CallbackData)
	}
}

func TestRenderGameConfigMenu(t *testing.T) {
	d := testData()
	_, keyboard := renderGameConfigMenu(d)

	if b := findButton(t, keyboard, "Движок: DozorClassic"); b.CallbackData != "gc:engine" {
		t.Errorf("engine = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Город: e-burg"); b.CallbackData != "gc:field:city" {
		t.Errorf("city = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Подписчики: 0"); b.CallbackData != "cs:list" {
		t.Errorf("subscribers = %q", b.CallbackData)
	}

	// No game subscribers row without an active game.
	for _, button := range flatButtons(keyboard) {
		if button.CallbackData == "gs:list" {
			t.Error("game subscribers offered without a game")
		}
	}

	d.Game = &store.Game{Engine: "DozorClassic", Subscriptions: store.Subscriptions{store.AllUpdates(-100)}}
	_, keyboard = renderGameConfigMenu(d)
	if b := findButton(t, keyboard, "Подписчики игры: 1"); b.CallbackData != "gs:list" {
		t.Errorf("game subscribers = %q", b.CallbackData)
	}
}

func TestRenderEngineChoice(t *testing.T) {
	d := testData()
	d.GameConfig.Engine = "DozorLite"

	_, keyboard := renderEngineChoice(d)
	if b := findButton(t, keyboard, mark(true, "DozorLite")); b.CallbackData != "gc:seteng:DozorLite" {
		t.Errorf("active engine = %q", b.CallbackData)
	}
	for _, button := range flatButtons(keyboard) {
		if button.CallbackData == "gc:seteng:DozorClassic" && strings.HasPrefix(button.Text, "✓") {
			t.Errorf("inactive engine marked: %q", button.Text)
		}
	}
}

func TestMark(t *testing.T) {
	if got := mark(true, "Вася"); got != "✓ Вася" {
		t.Errorf("active = %q", got)
	}
	if got := mark(false, "Вася"); got != "Вася" {
		t.Errorf("inactive = %q", got)
	}
}

func TestTruncateLabel(t *testing.T) {
	if got := truncateLabel("короткий"); got != "короткий" {
		t.Errorf("short = %q", got)
	}
	long := strings.Repeat("щ", 60)
	got := truncateLabel(long)
	if len([]rune(got)) != 48 || !strings.HasSuffix(got, "...") {
		t.Errorf("long = %q (%d runes)", got, len([]rune(got)))
	}
}
