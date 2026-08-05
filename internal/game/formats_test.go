package game

import "testing"

var defaultFormats = [][]string{{"dr", "др", "--"}}

func TestPrepareCode(t *testing.T) {
	tests := []struct {
		name    string
		message string
		formats [][]string
		want    string
		wantOK  bool
	}{
		{name: "canonical form passes through", message: "dr123", formats: defaultFormats, want: "dr123", wantOK: true},
		{name: "cyrillic variant is rewritten", message: "др123", formats: defaultFormats, want: "dr123", wantOK: true},
		{name: "dashes variant is rewritten", message: "--123", formats: defaultFormats, want: "dr123", wantOK: true},
		{name: "uppercase is lowered first", message: "ДР123", formats: defaultFormats, want: "dr123", wantOK: true},
		{name: "digits between letters keep positions", message: "д1р2", formats: defaultFormats, want: "d1r2", wantOK: true},
		{name: "exclamation mark forces the code", message: "!whatever", formats: defaultFormats, want: "whatever", wantOK: true},
		{name: "dot forces the code", message: ".код", formats: defaultFormats, want: "код", wantOK: true},
		{name: "mark strips only one char", message: "!!x", formats: defaultFormats, want: "!x", wantOK: true},
		{name: "plain chatter is not a code", message: "привет", formats: defaultFormats, want: "", wantOK: false},
		{name: "wrong letters are not a code", message: "xy123", formats: defaultFormats, want: "", wantOK: false},
		{name: "digits only match empty pattern format", message: "123", formats: [][]string{{""}}, want: "123", wantOK: true},
		{name: "digits only with default formats are not a code", message: "123", formats: defaultFormats, want: "", wantOK: false},
		{name: "empty formats accept nothing", message: "dr123", formats: nil, want: "", wantOK: false},
		{name: "second format wins", message: "кр7", formats: [][]string{{"dr", "др"}, {"kr", "кр"}}, want: "kr7", wantOK: true},
		{name: "variant longer than canonical is trimmed by zip", message: "дрр5", formats: [][]string{{"dr", "дрр"}}, want: "drр5", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := PrepareCode(tt.message, tt.formats)
			if ok != tt.wantOK || got != tt.want {
				t.Fatalf("PrepareCode(%q) = (%q, %v), want (%q, %v)", tt.message, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
