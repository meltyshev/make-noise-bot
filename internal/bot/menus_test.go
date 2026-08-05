package bot

import (
	"strings"
	"testing"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/store"
)

func testData() *store.Data {
	return &store.Data{
		Managers:  []int64{10},
		UserNames: map[int64]string{30: "Петя"},
		Chats: map[int64]*store.Chat{
			10:   {ID: 10, Type: "private", Permission: store.PermissionAllowed, FirstName: "Вася"},
			20:   {ID: 20, Type: "private", Permission: store.PermissionRequested, FirstName: "Аня"},
			-100: {ID: -100, Type: "supergroup", Permission: store.PermissionAllowed, Title: "Команда"},
			-200: {ID: -200, Type: "group", Permission: store.PermissionForbidden, Title: "Чужие"},
		},
		GameConfig: store.DefaultGameConfig(),
	}
}

func flatButtons(keyboard [][]models.InlineKeyboardButton) []models.InlineKeyboardButton {
	var buttons []models.InlineKeyboardButton
	for _, row := range keyboard {
		buttons = append(buttons, row...)
	}
	return buttons
}

func findButton(t *testing.T, keyboard [][]models.InlineKeyboardButton, labelPart string) models.InlineKeyboardButton {
	t.Helper()
	for _, button := range flatButtons(keyboard) {
		if strings.Contains(button.Text, labelPart) {
			return button
		}
	}
	t.Fatalf("button %q not found", labelPart)
	return models.InlineKeyboardButton{}
}

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
}

func TestRenderManagers(t *testing.T) {
	_, keyboard := renderManagers(testData())

	// Managers list shows private chats and known picked users as toggles.
	if b := findButton(t, keyboard, "Вася ✓"); b.CallbackData != "cfg:mgrt:10" {
		t.Errorf("manager toggle = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, "Аня"); b.CallbackData != "cfg:mgrt:20" {
		t.Errorf("candidate toggle = %q", b.CallbackData)
	}
	findButton(t, keyboard, "Добавить")

	// Group chats are not manager candidates.
	for _, button := range flatButtons(keyboard) {
		if strings.Contains(button.Text, "Команда") {
			t.Errorf("group chat offered as manager: %q", button.Text)
		}
	}
}

func TestRenderManagersUnknownID(t *testing.T) {
	d := testData()
	d.Managers = append(d.Managers, 999)

	_, keyboard := renderManagers(d)
	if b := findButton(t, keyboard, "ID 999 ✓"); b.CallbackData != "cfg:mgrt:999" {
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
	if b := findButton(t, keyboard, "Подписчики: 0"); b.CallbackData != "gc:subs" {
		t.Errorf("subscribers = %q", b.CallbackData)
	}

	// No game subscribers row without an active game.
	for _, button := range flatButtons(keyboard) {
		if button.CallbackData == "gs:list" {
			t.Error("game subscribers offered without a game")
		}
	}

	d.Game = &store.Game{Engine: "DozorClassic", Subscribers: []int64{-100}}
	_, keyboard = renderGameConfigMenu(d)
	if b := findButton(t, keyboard, "Подписчики игры: 1"); b.CallbackData != "gs:list" {
		t.Errorf("game subscribers = %q", b.CallbackData)
	}
}

func TestRenderSubscribers(t *testing.T) {
	d := testData()
	d.GameConfig.Subscribers = []int64{-100, 555}

	_, keyboard := renderSubscribers(d, false)

	// Allowed chats plus already-subscribed unknown ids are offered.
	if b := findButton(t, keyboard, "Команда ✓"); b.CallbackData != "gc:subst:-100" {
		t.Errorf("subscribed chat = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, "ID 555 ✓"); b.CallbackData != "gc:subst:555" {
		t.Errorf("unknown subscriber = %q", b.CallbackData)
	}
	findButton(t, keyboard, "Вася")

	// Forbidden and requested chats are not offered.
	for _, button := range flatButtons(keyboard) {
		if strings.Contains(button.Text, "Чужие") || strings.Contains(button.Text, "Аня") {
			t.Errorf("not-allowed chat offered: %q", button.Text)
		}
	}

	d.Game = &store.Game{Engine: "DozorClassic", Subscribers: []int64{-100}}
	_, keyboard = renderSubscribers(d, true)
	if b := findButton(t, keyboard, "Команда ✓"); b.CallbackData != "gs:t:-100" {
		t.Errorf("game subscriber toggle = %q", b.CallbackData)
	}
}

func TestToggleID(t *testing.T) {
	list := toggleID(nil, 5)
	if len(list) != 1 || list[0] != 5 {
		t.Errorf("add = %v", list)
	}
	list = toggleID(list, 5)
	if len(list) != 0 {
		t.Errorf("remove = %v", list)
	}
}

func TestSetMembership(t *testing.T) {
	list := setMembership(nil, 5, true)
	if len(list) != 1 || list[0] != 5 {
		t.Errorf("add = %v", list)
	}
	if got := setMembership(list, 5, true); len(got) != 1 {
		t.Errorf("add existing = %v", got)
	}
	if got := setMembership(list, 7, false); len(got) != 1 {
		t.Errorf("remove missing = %v", got)
	}
	if got := setMembership(list, 5, false); len(got) != 0 {
		t.Errorf("remove = %v", got)
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
