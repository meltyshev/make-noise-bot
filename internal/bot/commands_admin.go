package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

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

			chat := store.Chat{
				ID:         c.ChatID(),
				Type:       c.ChatType(),
				Permission: store.PermissionRequested,
				Title:      c.msg.Chat.Title,
				Username:   c.msg.Chat.Username,
				FirstName:  c.msg.Chat.FirstName,
				LastName:   c.msg.Chat.LastName,
			}
			// The store gets its own allocation, so the local chat stays a
			// value this handler can read after the lock is released.
			if err := c.app.store.Update(func(d *store.Data) {
				stored := chat
				d.Chats[chat.ID] = &stored
			}); err != nil {
				c.app.reportError(err)
				return
			}

			var text strings.Builder
			fmt.Fprintf(&text, texts.PermissionRequestTitleFmt, c.app.me.Username)
			fmt.Fprintf(&text, texts.PermissionRequestTypeFmt, chat.Type)
			for _, field := range []struct{ name, value string }{
				{texts.PermissionFieldTitle, chat.Title},
				{texts.PermissionFieldUsername, chat.Username},
				{texts.PermissionFieldFirstName, chat.FirstName},
				{texts.PermissionFieldLastName, chat.LastName},
			} {
				if field.value != "" {
					fmt.Fprintf(&text, texts.PermissionRequestFieldFmt, field.name, field.value)
				}
			}
			fmt.Fprintf(&text, texts.PermissionRequestIDFmt, chat.ID)

			if err := c.app.sendInline(c.ctx, c.app.adminID(), text.String(), permRequestKeyboard(chat.ID)); err != nil {
				c.app.log.Warn("send to admin failed", "error", err)
			}
			c.Reply(texts.PermissionRequestSent)
		},
	}
}

// permissionChange implements /allow and /forbid.
func permissionChange(name string, target store.Permission, notification string) *Command {
	apply := func(c *Ctx, input string) {
		chatID, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
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
	}

	return &Command{
		Name: name,
		Init: func(c *Ctx, args string) {
			if !c.IsAdmin() {
				return
			}
			if args != "" {
				c.DelConv()
				apply(c, args)
				return
			}
			c.SetConv(name)
			c.Reply(texts.AskChatID)
		},
		Handle: func(c *Ctx, _ any) {
			c.DelConv()
			apply(c, c.Text())
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
	apply := func(c *Ctx, input string) {
		chatID, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
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
	}

	return &Command{
		Name: "drop",
		Init: func(c *Ctx, args string) {
			if !c.IsAdmin() {
				return
			}
			if args != "" {
				c.DelConv()
				apply(c, args)
				return
			}
			c.SetConv("drop")
			c.Reply(texts.AskChatID)
		},
		Handle: func(c *Ctx, _ any) {
			c.DelConv()
			apply(c, c.Text())
		},
	}
}

func cmdWrite() *Command {
	deliver := func(c *Ctx, chatID int64, text string) {
		if text == "" {
			c.Reply(texts.TextRequired)
			return
		}
		if err := c.Send(chatID, text); err != nil {
			c.Reply(texts.ChatNotFound)
			return
		}
		c.Reply(texts.Done)
	}

	return &Command{
		Name: "write",
		Init: func(c *Ctx, args string) {
			if !c.IsAdmin() {
				return
			}

			if args != "" {
				idText, text, _ := strings.Cut(args, " ")
				chatID, err := strconv.ParseInt(idText, 10, 64)
				if err == nil {
					text = strings.TrimSpace(text)
					if text == "" {
						c.SetConvState("write", chatID)
						c.Reply(texts.WriteWhat)
						return
					}
					c.DelConv()
					deliver(c, chatID, text)
					return
				}
			}

			c.SetConv("write")
			c.Reply(texts.AskChatID)
		},
		Handle: func(c *Ctx, state any) {
			if state == nil {
				chatID, err := strconv.ParseInt(strings.TrimSpace(c.Text()), 10, 64)
				if err != nil {
					c.Reply(texts.AskChatID)
					return
				}
				c.SetConvState("write", chatID)
				c.Reply(texts.WriteWhat)
				return
			}

			c.DelConv()
			chatID, ok := state.(int64)
			if !ok {
				c.Reply(texts.TextRequired)
				return
			}
			deliver(c, chatID, c.Text())
		},
	}
}

func cmdChats() *Command {
	return &Command{
		Name: "chats",
		Init: func(c *Ctx, _ string) {
			if !c.IsAdmin() || !c.EnsurePrivate() {
				return
			}
			c.ReplyMenu(renderChatsList)
		},
	}
}

func cmdConfig() *Command {
	return &Command{
		Name: "config",
		Init: func(c *Ctx, _ string) {
			if !c.IsAdmin() || !c.EnsurePrivate() {
				return
			}
			c.ReplyMenu(renderConfigMenu)
		},
		Handle: func(c *Ctx, state any) {
			pick, ok := state.(pickManagers)
			if !ok {
				c.DelConv()
				return
			}

			if c.msg.UsersShared != nil {
				err := c.app.store.Update(func(d *store.Data) {
					for _, user := range c.msg.UsersShared.Users {
						name := strings.TrimSpace(strings.TrimSpace(user.FirstName) + " " + strings.TrimSpace(user.LastName))
						if name != "" {
							d.UserNames[user.UserID] = name
						}
						if !d.IsManager(user.UserID) {
							d.Managers = append(d.Managers, user.UserID)
						}
					}
				})
				if err != nil {
					c.app.reportError(err)
					return
				}

				c.DelConv()
				c.ReplyRemoveKeyboard(texts.Done)
				c.app.editMenu(c.ctx, pick.ChatID, pick.MsgID, renderManagers)
				return
			}

			if c.Text() == texts.ButtonCancel {
				c.DelConv()
				c.ReplyRemoveKeyboard(texts.Done)
			}
		},
	}
}
