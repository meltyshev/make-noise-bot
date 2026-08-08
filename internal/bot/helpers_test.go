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
			40:   {ID: 40, Type: "private", Permission: store.PermissionAllowed, FirstName: "Дима"},
			50:   {ID: 50, Type: "private", Permission: store.PermissionForbidden, FirstName: "Гоша"},
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
