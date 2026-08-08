package bot

import (
	"strings"
	"testing"

	"github.com/meltyshev/make-noise-bot/internal/texts"
)

func TestParseCodeFormats(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   [][]string
		wantOK bool
	}{
		{
			name:   "single group",
			input:  "dr=др=--",
			want:   [][]string{{"dr", "др", "--"}},
			wantOK: true,
		},
		{
			name:   "moscow style",
			input:  "dr=др, rd=рд, d=д, r=р",
			want:   [][]string{{"dr", "др"}, {"rd", "рд"}, {"d", "д"}, {"r", "р"}},
			wantOK: true,
		},
		{
			name:   "group without synonyms",
			input:  "dr",
			want:   [][]string{{"dr"}},
			wantOK: true,
		},
		{
			name:   "input is lowercased and trimmed",
			input:  " DR = ДР ",
			want:   [][]string{{"dr", "др"}},
			wantOK: true,
		},
		{
			name:   "legacy json still works",
			input:  `[["dr", "др"]]`,
			want:   [][]string{{"dr", "др"}},
			wantOK: true,
		},
		{name: "empty input", input: "  ", wantOK: false},
		{name: "empty variant", input: "dr=", wantOK: false},
		{name: "empty group", input: "dr,,rd", wantOK: false},
		{name: "digits in variant", input: "dr1=др", wantOK: false},
		{name: "broken json", input: "[[", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseCodeFormats(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("parseCodeFormats(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && !formatsEqual(got, tt.want) {
				t.Errorf("parseCodeFormats(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatsLabel(t *testing.T) {
	if got := formatsLabel([][]string{{""}}); got != texts.PresetDigitsOnly {
		t.Errorf("digits preset = %q", got)
	}
	if got := formatsLabel([][]string{{"dr", "др", "--"}}); got != texts.PresetDR {
		t.Errorf("dr preset = %q", got)
	}
	if got := formatsLabel([][]string{{"dr", "др"}, {"rd", "рд"}, {"d", "д"}, {"r", "р"}}); got != texts.PresetMoscow {
		t.Errorf("moscow preset = %q", got)
	}
	if got := formatsLabel([][]string{{"kr", "кр"}}); got != "kr=кр" {
		t.Errorf("custom = %q", got)
	}
	// Configs the friendly syntax cannot express fall back to JSON.
	if got := formatsLabel([][]string{{"a,b"}}); got != `[["a,b"]]` {
		t.Errorf("json fallback = %q", got)
	}
}

func TestFormatsRoundTrip(t *testing.T) {
	for _, preset := range formatPresets[1:] {
		label := formatsLabel(preset.Formats)
		if label == texts.PresetDR || label == texts.PresetMoscow {
			continue
		}
		parsed, ok := parseCodeFormats(label)
		if !ok || !formatsEqual(parsed, preset.Formats) {
			t.Errorf("round trip failed for %v via %q", preset.Formats, label)
		}
	}

	custom := [][]string{{"kr", "кр"}, {"x", "х"}}
	parsed, ok := parseCodeFormats(formatsLabel(custom))
	if !ok || !formatsEqual(parsed, custom) {
		t.Errorf("custom round trip = %v, ok=%v", parsed, ok)
	}
}

// TestPresetLabelsListEveryVariant pins that a preset label names every
// variant it accepts, so labels cannot drift from the formats behind them.
func TestPresetLabelsListEveryVariant(t *testing.T) {
	for _, preset := range formatPresets {
		for _, group := range preset.Formats {
			for _, variant := range group {
				if variant == "" {
					continue
				}
				if !strings.Contains(preset.Label, variant) {
					t.Errorf("%q does not mention %q", preset.Label, variant)
				}
			}
		}
		if len([]rune(mark(true, preset.Label))) > 48 {
			t.Errorf("%q is too long for a button", preset.Label)
		}
	}
}

func TestRenderFormatsMenu(t *testing.T) {
	d := testData()

	_, keyboard := renderFormatsMenu(d)
	if b := findButton(t, keyboard, mark(true, texts.PresetDR)); b.CallbackData != "gc:fmtp:1" {
		t.Errorf("active preset = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, texts.PresetDigitsOnly); b.CallbackData != "gc:fmtp:0" {
		t.Errorf("digits preset = %q", b.CallbackData)
	}
	if b := findButton(t, keyboard, texts.ButtonManual); b.CallbackData != "gc:fmtm" {
		t.Errorf("manual = %q", b.CallbackData)
	}
}
