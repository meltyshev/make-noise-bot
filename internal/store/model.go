package store

import (
	"fmt"
	"slices"

	"github.com/meltyshev/make-noise-bot/internal/texts"
)

type Permission string

const (
	PermissionRequested Permission = "requested"
	PermissionAllowed   Permission = "allowed"
	PermissionForbidden Permission = "forbidden"
)

type Chat struct {
	ID         int64      `json:"id"`
	Type       string     `json:"type"`
	Permission Permission `json:"permission"`
	BruteForce bool       `json:"brute_force,omitempty"`

	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// DisplayName is the title for groups or the person's name for private
// chats.
func (c *Chat) DisplayName() string {
	name := c.Title
	for _, part := range []string{c.FirstName, c.LastName} {
		if part == "" {
			continue
		}
		if name != "" {
			name += " "
		}
		name += part
	}
	return name
}

// Subscription is a chat that receives updates. An events-only chat gets the
// short notices without the level texts.
type Subscription struct {
	ChatID     int64 `json:"chat_id"`
	EventsOnly bool  `json:"events_only,omitempty"`
}

func AllUpdates(chatID int64) Subscription {
	return Subscription{ChatID: chatID}
}

func OnlyEvents(chatID int64) Subscription {
	return Subscription{ChatID: chatID, EventsOnly: true}
}

type Subscriptions []Subscription

func (s Subscriptions) Find(chatID int64) (Subscription, bool) {
	for _, sub := range s {
		if sub.ChatID == chatID {
			return sub, true
		}
	}
	return Subscription{ChatID: chatID}, false
}

func (s Subscriptions) Set(sub Subscription) Subscriptions {
	for i, item := range s {
		if item.ChatID == sub.ChatID {
			s[i] = sub
			return s
		}
	}
	return append(s, sub)
}

func (s Subscriptions) Remove(chatID int64) Subscriptions {
	for i, item := range s {
		if item.ChatID == chatID {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func (s Subscriptions) Clone() Subscriptions {
	return append(Subscriptions{}, s...)
}

// GameConfig is the template used to start the next game.
type GameConfig struct {
	Engine        string        `json:"engine"`
	City          string        `json:"city"`
	Login         string        `json:"login"`
	Password      string        `json:"password"`
	Pincode       string        `json:"pincode"`
	GameID        string        `json:"game_id"`
	League        string        `json:"league"`
	CodeFormats   [][]string    `json:"code_formats"`
	Subscriptions Subscriptions `json:"subscriptions"`
}

func DefaultGameConfig() GameConfig {
	return GameConfig{
		Engine:        "DozorClassic",
		City:          "e-burg",
		Login:         "-",
		Password:      "-",
		Pincode:       "-",
		GameID:        "-",
		League:        "-",
		CodeFormats:   [][]string{{"dr", "др", "--"}},
		Subscriptions: Subscriptions{},
	}
}

// Game is the currently running game. nil in Data means no active game.
type Game struct {
	Engine        string        `json:"engine"`
	City          string        `json:"city"`
	CodeFormats   [][]string    `json:"code_formats"`
	Subscriptions Subscriptions `json:"subscriptions"`
	Restricted    bool          `json:"restricted,omitempty"`

	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
	Pincode  string `json:"pincode,omitempty"`
	GameID   string `json:"game_id,omitempty"`
	League   string `json:"league,omitempty"`
	Session  string `json:"session,omitempty"`

	LevelNumber *int `json:"level_number,omitempty"`
	// LevelTask fingerprints the task itself and LevelTime is its timer, so
	// a renumbered level is not mistaken for a new one.
	LevelTask      string `json:"level_task,omitempty"`
	LevelTime      *int   `json:"level_time,omitempty"`
	HintNumber     *int   `json:"hint_number,omitempty"`
	SolvedSpoilers []int  `json:"solved_spoilers,omitempty"`
	PinnedLevel    *int   `json:"pinned_level,omitempty"`
}

// Player counts a user's accepted codes.
type Player struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

// Data is everything the bot persists between restarts.
type Data struct {
	SchemaVersion int     `json:"schema_version"`
	Managers      []int64 `json:"managers"`
	LeaveMode     bool    `json:"leave_mode,omitempty"`
	// MapService names the map links coordinates point to; empty means the
	// default.
	MapService string `json:"map_service,omitempty"`
	// UserNames remembers display names of users picked via the manager
	// picker, which may have no chat with the bot.
	UserNames  map[int64]string  `json:"user_names,omitempty"`
	Chats      map[int64]*Chat   `json:"chats"`
	GameConfig GameConfig        `json:"game_config"`
	Game       *Game             `json:"game,omitempty"`
	Players    map[int64]*Player `json:"players"`
}

func newData() *Data {
	return &Data{
		Managers:   []int64{},
		UserNames:  map[int64]string{},
		Chats:      map[int64]*Chat{},
		GameConfig: DefaultGameConfig(),
		Players:    map[int64]*Player{},
	}
}

// DisplayName resolves the best known name for a chat or user id.
func (d *Data) DisplayName(id int64) string {
	if chat, ok := d.Chats[id]; ok {
		if name := chat.DisplayName(); name != "" {
			return name
		}
	}
	if name, ok := d.UserNames[id]; ok && name != "" {
		return name
	}
	return fmt.Sprintf(texts.UnknownNameFmt, id)
}

func (d *Data) IsManager(userID int64) bool {
	return slices.Contains(d.Managers, userID)
}
