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
		t.Errorf("permRequestKeyboard callbacks = (%q, %q), want (perm:a:-100123, perm:f:-100123)", buttons[0].CallbackData, buttons[1].CallbackData)
	}
}

func TestRenderConfigMenu(t *testing.T) {
	d := testData()
	d.LeaveMode = true

	_, keyboard := renderConfigMenu(d)
	if b := findButton(t, keyboard, "Менеджеры: 1"); b.CallbackData != "cfg:mgr" {
		t.Errorf("managers callback = %q, want cfg:mgr", b.CallbackData)
	}
	if b := findButton(t, keyboard, "включен"); b.CallbackData != "cfg:leave" {
		t.Errorf("leave callback = %q, want cfg:leave", b.CallbackData)
	}
	// An empty setting shows the default service.
	label := fmt.Sprintf(texts.MapServiceFmt, geo.DefaultService.Label())
	if b := findButton(t, keyboard, label); b.CallbackData != "cfg:maps" {
		t.Errorf("maps callback = %q, want cfg:maps", b.CallbackData)
	}
	if len([]rune(label)) > 48 {
		t.Errorf("map label = %q (%d runes), want at most 48", label, len([]rune(label)))
	}
}

func TestRenderMapServices(t *testing.T) {
	d := testData()
	d.MapService = string(geo.Google)

	_, keyboard := renderMapServices(d)
	if b := findButton(t, keyboard, mark(true, geo.Google.Label())); b.CallbackData != "cfg:setmap:google" {
		t.Errorf("active service callback = %q, want cfg:setmap:google", b.CallbackData)
	}
	if b := findButton(t, keyboard, geo.OSM.Label()); b.CallbackData != "cfg:setmap:osm" {
		t.Errorf("other service callback = %q, want cfg:setmap:osm", b.CallbackData)
	}
	for _, button := range flatButtons(keyboard) {
		if button.CallbackData == "cfg:setmap:yandex" && strings.HasPrefix(button.Text, "✓") {
			t.Errorf("inactive service label = %q, want no checkmark", button.Text)
		}
	}
}

func TestRenderManagers(t *testing.T) {
	_, keyboard := renderManagers(testData())

	if b := findButton(t, keyboard, "✓ Вася"); b.CallbackData != "cfg:mgrt:10" {
		t.Errorf("manager toggle = %q, want cfg:mgrt:10", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Дима"); b.CallbackData != "cfg:mgrt:40" {
		t.Errorf("candidate toggle = %q, want cfg:mgrt:40", b.CallbackData)
	}
	findButton(t, keyboard, "Добавить")

	// Only allowed private chats are candidates.
	for _, button := range flatButtons(keyboard) {
		for _, hidden := range []string{"Аня", "Гоша", "Команда", "Чужие"} {
			if strings.Contains(button.Text, hidden) {
				t.Errorf("manager candidates include %q, want %s left out", button.Text, hidden)
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
		t.Errorf("forbidden manager toggle = %q, want cfg:mgrt:50", b.CallbackData)
	}
}

func TestRenderManagersUnknownID(t *testing.T) {
	d := testData()
	d.Managers = append(d.Managers, 999)

	_, keyboard := renderManagers(d)
	if b := findButton(t, keyboard, "✓ ID 999"); b.CallbackData != "cfg:mgrt:999" {
		t.Errorf("unknown manager toggle = %q, want cfg:mgrt:999", b.CallbackData)
	}
}

// TestRenderChatsList pins that the list marks the allowed chat the same way
// every other menu does, and spells the two states a checkmark cannot show.
func TestRenderChatsList(t *testing.T) {
	_, keyboard := renderChatsList(testData())

	if b := findButton(t, keyboard, "✓ Команда"); b.CallbackData != "ch:open:-100" {
		t.Errorf("allowed chat callback = %q, want ch:open:-100", b.CallbackData)
	}
	findButton(t, keyboard, "Чужие: запрещено")
	findButton(t, keyboard, "Аня: запрошено")
}

// TestRenderChatsListKeepsTheStateOnLongNames pins that a long chat title
// costs the name characters, not the permission summary: the state has to
// survive the label budget the way a mark() prefix does.
func TestRenderChatsListKeepsTheStateOnLongNames(t *testing.T) {
	d := testData()
	d.Chats[-200].Title = strings.Repeat("щ", 80)

	_, keyboard := renderChatsList(d)
	b := findButton(t, keyboard, strings.Repeat("щ", 20))
	if !strings.HasSuffix(b.Text, texts.SummaryForbidden) {
		t.Errorf("long-name label = %q, want it to still end with %q", b.Text, texts.SummaryForbidden)
	}
	if len([]rune(b.Text)) > maxLabelRunes {
		t.Errorf("long-name label = %d runes, want at most %d", len([]rune(b.Text)), maxLabelRunes)
	}
}

func TestRenderChatActions(t *testing.T) {
	chat := store.Chat{ID: -100, Type: "supergroup", Permission: store.PermissionAllowed, Title: "Команда"}
	text, keyboard := renderChatActions(chat)

	if !strings.Contains(text, "Команда | -100") || !strings.Contains(text, "allowed") {
		t.Errorf("renderChatActions text = %q, want it to name the chat and its status", text)
	}
	if b := findButton(t, keyboard, "Разрешить"); b.CallbackData != "ch:allow:-100" {
		t.Errorf("allow callback = %q, want ch:allow:-100", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Удалить"); b.CallbackData != "ch:del:-100" {
		t.Errorf("delete callback = %q, want ch:del:-100", b.CallbackData)
	}
}

func TestRenderGameConfigMenu(t *testing.T) {
	d := testData()
	_, keyboard := renderGameConfigMenu(d)

	if b := findButton(t, keyboard, "Движок: DozorClassic"); b.CallbackData != "gc:engine" {
		t.Errorf("engine callback = %q, want gc:engine", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Город: e-burg"); b.CallbackData != "gc:field:city" {
		t.Errorf("city callback = %q, want gc:field:city", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Подписчики: 0"); b.CallbackData != "cs:list" {
		t.Errorf("subscribers callback = %q, want cs:list", b.CallbackData)
	}

	// No game subscribers row without an active game.
	for _, button := range flatButtons(keyboard) {
		if button.CallbackData == "gs:list" {
			t.Error("renderGameConfigMenu offers gs:list, want no game subscribers row without a game")
		}
	}

	d.Game = &store.Game{Engine: "DozorClassic", Subscriptions: store.Subscriptions{store.AllUpdates(-100)}}
	_, keyboard = renderGameConfigMenu(d)
	if b := findButton(t, keyboard, "Подписчики игры: 1"); b.CallbackData != "gs:list" {
		t.Errorf("game subscribers callback = %q, want gs:list", b.CallbackData)
	}
}

func TestRenderEngineChoice(t *testing.T) {
	d := testData()
	d.GameConfig.Engine = "DozorLite"

	_, keyboard := renderEngineChoice(d)
	if b := findButton(t, keyboard, mark(true, "DozorLite")); b.CallbackData != "gc:seteng:DozorLite" {
		t.Errorf("active engine callback = %q, want gc:seteng:DozorLite", b.CallbackData)
	}
	for _, button := range flatButtons(keyboard) {
		if button.CallbackData == "gc:seteng:DozorClassic" && strings.HasPrefix(button.Text, "✓") {
			t.Errorf("inactive engine label = %q, want no checkmark", button.Text)
		}
	}
}

func TestMark(t *testing.T) {
	if got := mark(true, "Вася"); got != "✓ Вася" {
		t.Errorf("mark(true, \"Вася\") = %q, want \"✓ Вася\"", got)
	}
	if got := mark(false, "Вася"); got != "Вася" {
		t.Errorf("mark(false, \"Вася\") = %q, want \"Вася\"", got)
	}
}

func TestTruncateLabel(t *testing.T) {
	if got := truncateLabel("короткий"); got != "короткий" {
		t.Errorf("truncateLabel(\"короткий\") = %q, want it unchanged", got)
	}
	long := strings.Repeat("щ", 60)
	got := truncateLabel(long)
	if len([]rune(got)) != 48 || !strings.HasSuffix(got, "...") {
		t.Errorf("truncateLabel(60 runes) = %q (%d runes), want 48 runes ending in an ellipsis", got, len([]rune(got)))
	}
}
