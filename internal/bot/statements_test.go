package bot

import "testing"

func TestCodeToSend(t *testing.T) {
	formats := [][]string{{"dr", "др"}}

	tests := []struct {
		name       string
		text       string
		restricted bool
		bruteForce bool
		want       string
		wantOK     bool
	}{
		{name: "code in a format", text: "dr12", want: "dr12", wantOK: true},
		{name: "code in a synonym", text: "др12", want: "dr12", wantOK: true},
		{name: "chatter is not a code", text: "привет"},
		{name: "restricted blocks parsed codes", text: "dr12", restricted: true},
		{name: "restricted keeps marked codes", text: "!dr12", restricted: true, want: "dr12", wantOK: true},
		{name: "restricted keeps dotted codes", text: ".x1", restricted: true, want: "x1", wantOK: true},
		{name: "brute force sends anything", text: "привет", bruteForce: true, want: "привет", wantOK: true},
		{name: "restricted blocks brute force", text: "привет", restricted: true, bruteForce: true},
		{
			name: "a marked code loses its mark in brute force",
			text: "!привет", restricted: true, bruteForce: true,
			want: "привет", wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := codeToSend(tt.text, tt.restricted, tt.bruteForce, formats)
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("codeToSend(%q, restricted=%v, brute=%v) = (%q, %v), want (%q, %v)",
					tt.text, tt.restricted, tt.bruteForce, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
