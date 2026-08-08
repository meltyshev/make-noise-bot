package htmltext

import (
	"fmt"
	"testing"
)

func testLink(lat, lon float64) string {
	return fmt.Sprintf("https://maps.example/%.4f/%.4f", lat, lon)
}

// balanced fails when a split part carries markup Telegram would reject.
func balanced(t *testing.T, part string) {
	t.Helper()
	var stack []string
	for _, tok := range tokenize(part) {
		if !tok.isTag {
			continue
		}
		if tok.closing {
			if len(stack) == 0 || stack[len(stack)-1] != tok.name {
				t.Fatalf("Split part = %q, want balanced tags", part)
			}
			stack = stack[:len(stack)-1]
		} else {
			stack = append(stack, tok.name)
		}
	}
	if len(stack) != 0 {
		t.Fatalf("Split part = %q, want no unclosed tags, got %v", part, stack)
	}
}
