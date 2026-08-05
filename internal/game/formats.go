package game

import (
	"strings"
	"unicode"
)

// PrepareCode normalizes a chat message into an engine code. A leading "!"
// or "." forces the rest through untouched. Otherwise the message's
// non-digit characters must equal one of the format variants; a non-first
// variant is rewritten into the first one character by character ("др123"
// with ["dr", "др"] becomes "dr123"). ok is false when the message is not
// a code.
func PrepareCode(message string, formats [][]string) (code string, ok bool) {
	code = strings.ToLower(message)

	runes := []rune(code)
	if len(runes) == 0 {
		return "", false
	}
	if runes[0] == '!' || runes[0] == '.' {
		return string(runes[1:]), true
	}

	var nonDigits []rune
	for _, r := range runes {
		if !unicode.IsDigit(r) {
			nonDigits = append(nonDigits, r)
		}
	}
	pattern := string(nonDigits)

	for _, format := range formats {
		if len(format) == 0 {
			continue
		}

		first := format[0]
		if first == pattern {
			return code, true
		}

		for _, variant := range format[1:] {
			if variant != pattern {
				continue
			}

			firstRunes := []rune(first)
			for i, search := range []rune(variant) {
				if i >= len(firstRunes) {
					break
				}
				code = replaceFirstRune(code, search, firstRunes[i])
			}
			return code, true
		}
	}

	return "", false
}

func replaceFirstRune(s string, old, new rune) string {
	idx := strings.IndexRune(s, old)
	if idx < 0 {
		return s
	}
	return s[:idx] + string(new) + s[idx+len(string(old)):]
}
