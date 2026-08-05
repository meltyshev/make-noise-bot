package store

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

// GameConfig is the template used to start the next game.
type GameConfig struct {
	Engine      string     `json:"engine"`
	City        string     `json:"city"`
	Login       string     `json:"login"`
	Password    string     `json:"password"`
	Pincode     string     `json:"pincode"`
	GameID      string     `json:"game_id"`
	League      string     `json:"league"`
	CodeFormats [][]string `json:"code_formats"`
	Subscribers []int64    `json:"subscribers"`
}

func DefaultGameConfig() GameConfig {
	return GameConfig{
		Engine:      "DozorClassic",
		City:        "e-burg",
		Login:       "-",
		Password:    "-",
		Pincode:     "-",
		GameID:      "-",
		League:      "-",
		CodeFormats: [][]string{{"dr", "др", "--"}},
		Subscribers: []int64{},
	}
}

// Game is the currently running game. nil in Data means no active game.
type Game struct {
	Engine      string     `json:"engine"`
	City        string     `json:"city"`
	CodeFormats [][]string `json:"code_formats"`
	Subscribers []int64    `json:"subscribers"`
	Restricted  bool       `json:"restricted,omitempty"`

	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
	Pincode  string `json:"pincode,omitempty"`
	GameID   string `json:"game_id,omitempty"`
	League   string `json:"league,omitempty"`
	Session  string `json:"session,omitempty"`

	LevelNumber    *int  `json:"level_number,omitempty"`
	HintNumber     *int  `json:"hint_number,omitempty"`
	SolvedSpoilers []int `json:"solved_spoilers,omitempty"`
	PinnedLevel    *int  `json:"pinned_level,omitempty"`
}

func (g *Game) HasSubscriber(chatID int64) bool {
	for _, id := range g.Subscribers {
		if id == chatID {
			return true
		}
	}
	return false
}

// Player counts a user's accepted codes.
type Player struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

// Data is everything the bot persists between restarts.
type Data struct {
	Managers   []int64           `json:"managers"`
	LeaveMode  bool              `json:"leave_mode,omitempty"`
	Chats      map[int64]*Chat   `json:"chats"`
	GameConfig GameConfig        `json:"game_config"`
	Game       *Game             `json:"game,omitempty"`
	Players    map[int64]*Player `json:"players"`
}

func newData() *Data {
	return &Data{
		Managers:   []int64{},
		Chats:      map[int64]*Chat{},
		GameConfig: DefaultGameConfig(),
		Players:    map[int64]*Player{},
	}
}

func (d *Data) IsManager(userID int64) bool {
	for _, id := range d.Managers {
		if id == userID {
			return true
		}
	}
	return false
}
