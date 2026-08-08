package words

import (
	"slices"
	"testing"
)

func TestAnagramsRussian(t *testing.T) {
	matches, ok := Anagrams("кот")
	if !ok {
		t.Fatal("cyrillic letters must be supported")
	}
	if !slices.Contains(matches, "ток") {
		t.Errorf("кот anagrams = %v, want to contain ток", matches)
	}
	// The word itself is a dictionary noun and rearranges into itself.
	if !slices.Contains(matches, "кот") {
		t.Errorf("кот anagrams = %v, want to contain кот", matches)
	}
}

func TestAnagramsYoNormalization(t *testing.T) {
	// ё and е must be interchangeable.
	withYo, ok := Anagrams("ёлка")
	if !ok {
		t.Fatal("ёлка must be treated as cyrillic")
	}
	withE, _ := Anagrams("елка")
	if !slices.Equal(withYo, withE) {
		t.Errorf("anagrams with ё = %v, want the same as with е %v", withYo, withE)
	}
}

func TestAnagramsEnglish(t *testing.T) {
	matches, ok := Anagrams("silent")
	if !ok {
		t.Fatal("latin letters must be supported")
	}
	for _, want := range []string{"listen", "enlist", "tinsel"} {
		if !slices.Contains(matches, want) {
			t.Errorf("silent anagrams = %v, want to contain %q", matches, want)
		}
	}
}

func TestAnagramsMixedScript(t *testing.T) {
	if _, ok := Anagrams("aбв"); ok {
		t.Error("mixed scripts must be rejected")
	}
	if _, ok := Anagrams("абв1"); ok {
		t.Error("digits must be rejected")
	}
}

func TestMask(t *testing.T) {
	matches, ok := Mask("к-т")
	if !ok {
		t.Fatal("mask with dashes must work")
	}
	for _, want := range []string{"кот", "кит"} {
		if !slices.Contains(matches, want) {
			t.Errorf("к-т matches = %v, want to contain %q", matches, want)
		}
	}
	for _, match := range matches {
		if len([]rune(match)) != 3 {
			t.Errorf("mask \"к-т\" matched %q, want a 3-letter word", match)
		}
	}
}

func TestMaskUnderscoreAndLatin(t *testing.T) {
	matches, ok := Mask("c_t")
	if !ok {
		t.Fatal("latin mask must work")
	}
	if !slices.Contains(matches, "cat") || !slices.Contains(matches, "cut") {
		t.Errorf("c_t matches = %v, want cat and cut", matches)
	}
}

func TestMaskNoMatches(t *testing.T) {
	matches, ok := Mask("ъъъъъъ")
	if !ok {
		t.Fatal("all-cyrillic mask must be accepted")
	}
	if len(matches) != 0 {
		t.Errorf("mask \"ъъъъъъ\" matched %v, want no words", matches)
	}
}
