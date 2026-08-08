package bot

import (
	"strings"
	"testing"

	"github.com/go-telegram/bot/models"

	"github.com/meltyshev/make-noise-bot/internal/game"
	"github.com/meltyshev/make-noise-bot/internal/tgsend"
)

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantName string
		wantArgs string
	}{
		{name: "no arguments", text: "/morse", wantName: "morse"},
		{name: "arguments", text: "/morse .- -...", wantName: "morse", wantArgs: ".- -..."},
		{name: "newline separator", text: "/write\nтекст", wantName: "write", wantArgs: "текст"},
		{name: "repeated spaces", text: "/morse   двойные пробелы", wantName: "morse", wantArgs: "двойные пробелы"},
		{name: "trailing space", text: "/morse ", wantName: "morse"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, args := splitCommand(tt.text)
			if name != tt.wantName || args != tt.wantArgs {
				t.Errorf("splitCommand(%q) = (%q, %q), want (%q, %q)", tt.text, name, args, tt.wantName, tt.wantArgs)
			}
		})
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

func TestParseCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLat float64
		wantLon float64
		wantOK  bool
	}{
		{name: "space separated", input: "56.838011 60.597465", wantLat: 56.838011, wantLon: 60.597465, wantOK: true},
		{name: "comma and space", input: "56.838011, 60.597465", wantLat: 56.838011, wantLon: 60.597465, wantOK: true},
		{name: "comma only", input: "56.838011,60.597465", wantLat: 56.838011, wantLon: 60.597465, wantOK: true},
		{name: "negative latitude", input: "-12.5 30", wantLat: -12.5, wantLon: 30, wantOK: true},
		{name: "plain text", input: "адрес какой-то"},
		{name: "one number", input: "56.8"},
		{name: "three numbers", input: "56.8 60.5 12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lon, ok := parseCoordinates(tt.input)
			if ok != tt.wantOK || lat != tt.wantLat || lon != tt.wantLon {
				t.Errorf("parseCoordinates(%q) = (%v, %v, %v), want (%v, %v, %v)",
					tt.input, lat, lon, ok, tt.wantLat, tt.wantLon, tt.wantOK)
			}
		})
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
	if len(got) > tgsend.Limit+len("...") {
		t.Errorf("joined length %d exceeds the telegram limit", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("truncated join must end with an ellipsis")
	}
}
