package bot

import (
	"strings"
	"testing"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/game"
)

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		text     string
		wantName string
		wantArgs string
	}{
		{"/morse", "morse", ""},
		{"/morse .- -...", "morse", ".- -..."},
		{"/write\nтекст", "write", "текст"},
		{"/morse   двойные пробелы", "morse", "двойные пробелы"},
		{"/morse ", "morse", ""},
	}
	for _, tt := range tests {
		name, args := splitCommand(tt.text)
		if name != tt.wantName || args != tt.wantArgs {
			t.Errorf("splitCommand(%q) = (%q, %q), want (%q, %q)", tt.text, name, args, tt.wantName, tt.wantArgs)
		}
	}
}

func TestStripBotMention(t *testing.T) {
	a := &App{me: &models.User{Username: "MakeNoiseBot"}}

	if got := a.stripBotMention("morse@MakeNoiseBot"); got != "morse" {
		t.Errorf("own mention: %q", got)
	}
	if got := a.stripBotMention("morse@makenoisebot"); got != "morse" {
		t.Errorf("case-insensitive mention: %q", got)
	}
	if got := a.stripBotMention("morse@OtherBot"); got != "morse@OtherBot" {
		t.Errorf("other bot's command must stay unknown: %q", got)
	}
	if got := a.stripBotMention("morse"); got != "morse" {
		t.Errorf("plain name: %q", got)
	}
}

func TestFormatSector(t *testing.T) {
	sector := game.Sector{
		Name: "Основные коды",
		Codes: []game.SectorCode{
			{Number: 1, Hazard: "1.2", Entered: true},
			{Number: 2, Hazard: "N"},
			{Number: 3, Hazard: "2"},
			{Number: 4, Hazard: "1.5", Entered: true},
			{Number: 5, Hazard: "3"},
		},
	}

	// 5 codes: columns are [1 2] and [4 5], the middle code 3 goes last.
	want := strings.Join([]string{
		"Основные коды:",
		" 1) 1.2 ok    4) 1.5 ok",
		" 2) N        5) 3    ",
		" 3) 2    ",
	}, "\n")

	if got := formatSector(sector); got != want {
		t.Errorf("formatSector:\n%q\nwant:\n%q", got, want)
	}
}

func TestFormatSectorEven(t *testing.T) {
	sector := game.Sector{
		Name: "Бонусные коды",
		Codes: []game.SectorCode{
			{Number: 1, Hazard: "1"},
			{Number: 2, Hazard: "2", Entered: true},
		},
	}
	want := strings.Join([]string{
		"Бонусные коды:",
		" 1) 1        2) 2  ok",
	}, "\n")
	if got := formatSector(sector); got != want {
		t.Errorf("formatSector:\n%q\nwant:\n%q", got, want)
	}
}

func TestJoinLimited(t *testing.T) {
	short := []string{"кот", "ток"}
	if got := joinLimited(short); got != "кот, ток" {
		t.Errorf("short join = %q", got)
	}

	long := make([]string, 0, 2000)
	for range 2000 {
		long = append(long, "слово")
	}
	got := joinLimited(long)
	if len(got) > telegramTextLimit+len("...") {
		t.Errorf("joined length %d exceeds the telegram limit", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("truncated join must end with an ellipsis")
	}
}
