package updater

import (
	"strings"
	"testing"

	"github.com/meltyshev/make-noise-bot/internal/game"
)

func at(number int, task string, seconds int) levelState {
	return levelState{Number: &number, Task: task, Time: &seconds}
}

// TestIsNewLevel covers the cases an organizer can cause during a game.
func TestIsNewLevel(t *testing.T) {
	tests := []struct {
		name     string
		previous levelState
		current  levelState
		want     bool
	}{
		{
			name:     "team solved the level",
			previous: at(5, "aaa", 1800),
			current:  at(6, "bbb", 4),
			want:     true,
		},
		{
			name:     "organizer deleted an earlier level",
			previous: at(5, "aaa", 1800),
			current:  at(4, "aaa", 1805),
			want:     false,
		},
		{
			name:     "organizer inserted a level before",
			previous: at(5, "aaa", 1800),
			current:  at(6, "aaa", 1805),
			want:     false,
		},
		{
			name:     "organizer fixed a word in the task",
			previous: at(5, "aaa", 1800),
			current:  at(5, "bbb", 1805),
			want:     false,
		},
		{
			name:     "organizer deleted the current level",
			previous: at(5, "aaa", 1800),
			current:  at(5, "bbb", 3),
			want:     true,
		},
		{
			name:     "next level looks the same",
			previous: at(5, "aaa", 1800),
			current:  at(6, "aaa", 2),
			want:     true,
		},
		{
			name:     "nothing changed",
			previous: at(5, "aaa", 1800),
			current:  at(5, "aaa", 1805),
			want:     false,
		},
		{
			name:     "first look at a running game",
			previous: levelState{},
			current:  at(5, "aaa", 1800),
			want:     true,
		},
		{
			name:     "level disappeared",
			previous: at(5, "aaa", 1800),
			current:  levelState{},
			want:     false,
		},
		{
			name:     "engine hides the timer",
			previous: levelState{Number: new(5), Task: "aaa"},
			current:  levelState{Number: new(6), Task: "bbb"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNewLevel(tt.previous, tt.current); got != tt.want {
				t.Errorf("isNewLevel = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimerRestarted(t *testing.T) {
	// A countdown ticking down must not look like a restart.
	if timerRestarted(new(3600), new(3595)) {
		t.Error("a countdown tick counted as a restart")
	}
	if !timerRestarted(new(3600), new(5)) {
		t.Error("a timer falling back to zero is a restart")
	}
	if timerRestarted(new(300), new(400)) {
		t.Error("a growing timer is not a restart")
	}
	if timerRestarted(nil, new(5)) || timerRestarted(new(5), nil) {
		t.Error("an unknown timer cannot restart")
	}
	if !timerRestarted(new(130), new(120)) {
		t.Error("a drop into the start window is a restart")
	}
	if timerRestarted(new(4000), new(3000)) {
		t.Error("a drop that stays far from the start is not a restart")
	}
}

func TestLevelGone(t *testing.T) {
	if !levelGone(at(5, "aaa", 1800), levelState{}) {
		t.Error("a level that stopped being served is gone")
	}
	if levelGone(levelState{}, levelState{}) {
		t.Error("a game without a level yet has nothing to lose")
	}
	if levelGone(at(5, "aaa", 1800), at(6, "bbb", 3)) {
		t.Error("a new level is not a gone level")
	}
	if levelGone(levelState{}, at(6, "bbb", 3)) {
		t.Error("a level appearing is not a gone level")
	}
}

// TestLevelMessage pins that the full message carries the task, while
// events-only chats get the bare shout that announce() picks per chat.
func TestLevelMessage(t *testing.T) {
	if got := levelMessage("Найдите памятник"); got != "АП!\n\nНайдите памятник" {
		t.Errorf("levelMessage with a task = %q, want the shout and the task", got)
	}
	if got := levelMessage(""); got != "АП!" {
		t.Errorf("without a task = %q, want the shout alone", got)
	}
}

func TestSameLevel(t *testing.T) {
	if !sameLevel(at(5, "aaa", 10), at(5, "aaa", 10)) {
		t.Error("identical states differ")
	}
	if sameLevel(at(5, "aaa", 10), at(5, "aaa", 15)) {
		t.Error("a moved timer must be stored")
	}
	if !sameLevel(levelState{}, levelState{}) {
		t.Error("empty states differ")
	}
}

// fakeSnapshot is enough for fingerprinting.
type fakeSnapshot struct {
	number   *int
	question string
	sectors  []game.Sector
}

func (s fakeSnapshot) LevelNumber() *int         { return s.number }
func (s fakeSnapshot) Progress() string          { return "" }
func (s fakeSnapshot) Question() string          { return s.question }
func (s fakeSnapshot) Sectors() []game.Sector    { return s.sectors }
func (s fakeSnapshot) Hint() (int, string, bool) { return 0, "", false }
func (s fakeSnapshot) Spoilers() []game.Spoiler  { return nil }
func (s fakeSnapshot) TimeOnLevel() (int, bool)  { return 0, false }

func TestLevelTaskIgnoresProgress(t *testing.T) {
	level := fakeSnapshot{
		number:   new(5),
		question: "Найдите памятник",
		sectors: []game.Sector{{
			Name: "Основные коды",
			Codes: []game.SectorCode{
				{Number: 1, Hazard: "1.2"},
				{Number: 2, Hazard: "1.3"},
			},
		}},
	}

	entered := level
	entered.sectors = []game.Sector{{
		Name: "Основные коды",
		Codes: []game.SectorCode{
			{Number: 1, Hazard: "1.2", Entered: true},
			{Number: 2, Hazard: "1.3"},
		},
	}}

	if levelTask(level) != levelTask(entered) {
		t.Error("entering a code changed the task fingerprint")
	}

	edited := level
	edited.question = "Найдите памятник у вокзала"
	if levelTask(level) == levelTask(edited) {
		t.Error("an edited question must change the fingerprint")
	}

	relaid := level
	relaid.sectors = []game.Sector{{
		Name: "Основные коды",
		Codes: []game.SectorCode{
			{Number: 1, Hazard: "2.5"},
			{Number: 2, Hazard: "1.3"},
		},
	}}
	if levelTask(level) == levelTask(relaid) {
		t.Error("a different code layout must change the fingerprint")
	}

	if levelTask(fakeSnapshot{}) != "" {
		t.Error("no level means no fingerprint")
	}
	if got := levelTask(level); len(got) != 16 || strings.ContainsAny(got, "ghijklmnopqrstuvwxyz") {
		t.Errorf("fingerprint = %q, want short hex", got)
	}
}
