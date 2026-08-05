package bot

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

func kbRow(labels ...string) []models.KeyboardButton {
	row := make([]models.KeyboardButton, len(labels))
	for i, label := range labels {
		row[i] = models.KeyboardButton{Text: label}
	}
	return row
}

func cmdPermission() *Command {
	return &Command{
		Name: "permission",
		Init: func(c *Ctx, _ string) {
			if chat, ok := c.app.store.Chat(c.ChatID()); ok {
				switch chat.Permission {
				case store.PermissionRequested:
					c.Reply(texts.PermissionStatusRequested)
				case store.PermissionAllowed:
					c.Reply(texts.PermissionStatusAllowed)
				case store.PermissionForbidden:
					c.Reply(texts.PermissionStatusForbidden)
				}
				return
			}

			chat := &store.Chat{
				ID:         c.ChatID(),
				Type:       c.ChatType(),
				Permission: store.PermissionRequested,
				Title:      c.msg.Chat.Title,
				Username:   c.msg.Chat.Username,
				FirstName:  c.msg.Chat.FirstName,
				LastName:   c.msg.Chat.LastName,
			}
			if err := c.app.store.Update(func(d *store.Data) { d.Chats[chat.ID] = chat }); err != nil {
				c.app.reportError(err)
				return
			}

			text := fmt.Sprintf(texts.PermissionRequestTitleFmt, c.app.me.Username)
			text += "type: " + chat.Type
			for _, field := range []struct{ name, value string }{
				{"title", chat.Title},
				{"username", chat.Username},
				{"first_name", chat.FirstName},
				{"last_name", chat.LastName},
			} {
				if field.value != "" {
					text += fmt.Sprintf("\n%s: %s", field.name, field.value)
				}
			}

			if request := c.SendToAdmin(text); request != nil {
				c.app.sendToAdminReply(c.ctx, strconv.FormatInt(chat.ID, 10), request.ID)
			}
			c.Reply(texts.PermissionRequestSent)
		},
	}
}

// permissionChange implements /allow and /forbid.
func permissionChange(name string, target store.Permission, notification string) *Command {
	return &Command{
		Name: name,
		Init: func(c *Ctx, _ string) {
			if !c.IsAdmin() {
				return
			}
			c.SetConv(name)
			c.Reply(texts.AskChatID)
		},
		Handle: func(c *Ctx, _ any) {
			c.DelConv()

			chatID, err := strconv.ParseInt(strings.TrimSpace(c.Text()), 10, 64)
			if err != nil {
				c.Reply(texts.ChatNotFound)
				return
			}

			changed := false
			storeErr := c.app.store.Update(func(d *store.Data) {
				if chat, ok := d.Chats[chatID]; ok && chat.Permission != target {
					chat.Permission = target
					changed = true
				}
			})
			if storeErr != nil {
				c.app.reportError(storeErr)
				return
			}

			if !changed {
				c.Reply(texts.ChatNotFound)
				return
			}
			if err := c.Send(chatID, notification); err != nil {
				c.app.log.Warn("permission notification failed", "chat_id", chatID, "error", err)
			}
			c.Reply(texts.Done)
		},
	}
}

func cmdAllow() *Command {
	return permissionChange("allow", store.PermissionAllowed, texts.PermissionGranted)
}

func cmdForbid() *Command {
	return permissionChange("forbid", store.PermissionForbidden, texts.PermissionForbidden)
}

func cmdDrop() *Command {
	return &Command{
		Name: "drop",
		Init: func(c *Ctx, _ string) {
			if !c.IsAdmin() {
				return
			}
			c.SetConv("drop")
			c.Reply(texts.AskChatID)
		},
		Handle: func(c *Ctx, _ any) {
			c.DelConv()

			chatID, err := strconv.ParseInt(strings.TrimSpace(c.Text()), 10, 64)
			if err != nil {
				c.Reply(texts.ChatNotFound)
				return
			}

			deleted := false
			storeErr := c.app.store.Update(func(d *store.Data) {
				if _, ok := d.Chats[chatID]; ok {
					delete(d.Chats, chatID)
					deleted = true
				}
			})
			if storeErr != nil {
				c.app.reportError(storeErr)
				return
			}

			if deleted {
				c.Reply(texts.Done)
			} else {
				c.Reply(texts.ChatNotFound)
			}
		},
	}
}

func cmdWrite() *Command {
	return &Command{
		Name: "write",
		Init: func(c *Ctx, _ string) {
			if !c.IsAdmin() {
				return
			}
			c.SetConv("write")
			c.Reply(texts.AskChatIDWrite)
		},
		Handle: func(c *Ctx, state any) {
			if state == nil {
				chatID, err := strconv.ParseInt(strings.TrimSpace(c.Text()), 10, 64)
				if err != nil {
					c.Reply(texts.AskChatIDWrite)
					return
				}
				c.SetConvState("write", chatID)
				c.Reply(texts.WriteWhat)
				return
			}

			c.DelConv()
			chatID, ok := state.(int64)
			if !ok || c.Text() == "" {
				c.Reply(texts.TextRequired)
				return
			}
			if err := c.Send(chatID, c.Text()); err != nil {
				c.Reply(texts.ChatNotFound)
				return
			}
			c.Reply(texts.Done)
		},
	}
}

type chatsState struct {
	ChatID *int64
}

func cmdChats() *Command {
	mainMenu := func(c *Ctx) {
		c.SetConvState("chats", &chatsState{})

		var chats []store.Chat
		c.app.store.View(func(d *store.Data) {
			for _, chat := range d.Chats {
				chats = append(chats, *chat)
			}
		})
		sort.Slice(chats, func(i, j int) bool { return chats[i].ID < chats[j].ID })

		var keyboard [][]models.KeyboardButton
		for _, chat := range chats {
			keyboard = append(keyboard, kbRow(fmt.Sprintf("%s | %d", chat.DisplayName(), chat.ID)))
		}
		keyboard = append(keyboard, kbRow(texts.ButtonClose))

		c.ReplyKeyboard(texts.ChatsChoose, keyboard)
	}

	actionsMenu := func(c *Ctx, chat store.Chat) {
		chatID := chat.ID
		c.SetConvState("chats", &chatsState{ChatID: &chatID})
		c.ReplyKeyboard(
			fmt.Sprintf(texts.ChatsActionsFmt, chat.DisplayName(), chat.ID, chat.Type, chat.Permission),
			[][]models.KeyboardButton{
				kbRow(texts.ButtonDelete),
				kbRow(texts.ButtonBack, texts.ButtonClose),
			},
		)
	}

	closeMenu := func(c *Ctx) {
		c.DelConv()
		c.ReplyRemoveKeyboard(texts.ChatsClosed)
	}

	return &Command{
		Name: "chats",
		Init: func(c *Ctx, _ string) {
			if !c.IsAdmin() {
				return
			}
			mainMenu(c)
		},
		Handle: func(c *Ctx, state any) {
			menu, ok := state.(*chatsState)
			if !ok {
				mainMenu(c)
				return
			}

			if menu.ChatID == nil {
				if c.Text() == texts.ButtonClose {
					closeMenu(c)
					return
				}

				idx := strings.LastIndex(c.Text(), " | ")
				if idx < 0 {
					c.Reply(texts.ChatsNoAction)
					mainMenu(c)
					return
				}
				chatID, err := strconv.ParseInt(c.Text()[idx+3:], 10, 64)
				if err != nil {
					c.Reply(texts.ChatsNotFound)
					mainMenu(c)
					return
				}
				chat, found := c.app.store.Chat(chatID)
				if !found {
					c.Reply(texts.ChatsNotFound)
					mainMenu(c)
					return
				}
				actionsMenu(c, chat)
				return
			}

			chat, found := c.app.store.Chat(*menu.ChatID)
			if !found {
				c.Reply(texts.ChatsGone)
				mainMenu(c)
				return
			}

			switch c.Text() {
			case texts.ButtonBack:
				mainMenu(c)
			case texts.ButtonClose:
				closeMenu(c)
			case texts.ButtonDelete:
				err := c.app.store.Update(func(d *store.Data) { delete(d.Chats, chat.ID) })
				if err != nil {
					c.app.reportError(err)
					return
				}
				c.Reply(texts.ChatsDeleted)
				mainMenu(c)
			default:
				c.Reply(texts.ChatsNoAction)
				actionsMenu(c, chat)
			}
		},
	}
}

func cmdConfig() *Command {
	menu := func(c *Ctx) {
		c.SetConv("config")

		var (
			managers  []int64
			leaveMode bool
		)
		c.app.store.View(func(d *store.Data) {
			managers = append([]int64{}, d.Managers...)
			leaveMode = d.LeaveMode
		})

		managersJSON, _ := json.Marshal(managers)
		leaveModeWord := texts.LeaveModeOffWord
		if leaveMode {
			leaveModeWord = texts.LeaveModeOnWord
		}

		c.ReplyKeyboard(texts.SettingsTitle, [][]models.KeyboardButton{
			kbRow(fmt.Sprintf(texts.ManagersFmt, managersJSON)),
			kbRow(fmt.Sprintf(texts.LeaveModeFmt, leaveModeWord)),
			kbRow(texts.ButtonReset, texts.ButtonFinish),
		})
	}

	return &Command{
		Name: "config",
		Init: func(c *Ctx, _ string) {
			if !c.IsAdmin() {
				return
			}
			menu(c)
		},
		Handle: func(c *Ctx, state any) {
			text := c.Text()

			if state == "managers" {
				if text != texts.ButtonCancel {
					var managers []int64
					if err := json.Unmarshal([]byte(text), &managers); err == nil {
						if managers == nil {
							managers = []int64{}
						}
						if err := c.app.store.Update(func(d *store.Data) { d.Managers = managers }); err != nil {
							c.app.reportError(err)
							return
						}
					}
				}
				menu(c)
				return
			}

			switch {
			case strings.HasPrefix(text, "Менеджеры"):
				c.SetConvState("config", "managers")
				c.ReplyKeyboard(texts.ManagersAsk, [][]models.KeyboardButton{kbRow(texts.ButtonCancel)})
			case strings.HasPrefix(text, "Режим выхода"):
				err := c.app.store.Update(func(d *store.Data) { d.LeaveMode = !d.LeaveMode })
				if err != nil {
					c.app.reportError(err)
					return
				}
				menu(c)
			case text == texts.ButtonFinish:
				c.DelConv()
				c.ReplyRemoveKeyboard(texts.Done)
			case text == texts.ButtonReset:
				err := c.app.store.Update(func(d *store.Data) {
					d.Managers = []int64{}
					d.LeaveMode = false
				})
				if err != nil {
					c.app.reportError(err)
					return
				}
				menu(c)
			default:
				menu(c)
			}
		},
	}
}
