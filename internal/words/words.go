// Package words answers anagram and mask queries from dictionaries embedded
// into the binary (see NOTICE for their sources).
package words

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"embed"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode"
)

//go:embed assets/ru.txt.gz assets/en.txt.gz
var assets embed.FS

type dictionary struct {
	once     sync.Once
	file     string
	words    []string
	anagrams map[string][]string
}

var (
	ru = &dictionary{file: "assets/ru.txt.gz"}
	en = &dictionary{file: "assets/en.txt.gz"}
)

func (d *dictionary) load() {
	d.once.Do(func() {
		raw, err := assets.ReadFile(d.file)
		if err != nil {
			return
		}
		zr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return
		}
		defer zr.Close()

		d.anagrams = map[string][]string{}
		scanner := bufio.NewScanner(zr)
		for scanner.Scan() {
			word := normalize(strings.TrimSpace(scanner.Text()))
			if word == "" {
				continue
			}
			d.words = append(d.words, word)
			key := anagramKey(word)
			d.anagrams[key] = append(d.anagrams[key], word)
		}
	})
}

func normalize(word string) string {
	return strings.ReplaceAll(strings.ToLower(word), "ё", "е")
}

func anagramKey(word string) string {
	runes := []rune(word)
	slices.Sort(runes)
	return string(runes)
}

func isCyrillic(word string) bool {
	for _, r := range word {
		if !unicode.Is(unicode.Cyrillic, r) {
			return false
		}
	}
	return word != ""
}

func isLatin(word string) bool {
	for _, r := range word {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return word != ""
}

func pick(word string) *dictionary {
	switch {
	case isCyrillic(word):
		return ru
	case isLatin(word):
		return en
	default:
		return nil
	}
}

// Anagrams returns dictionary words that are exact rearrangements of the
// letters. ok is false for mixed or unsupported scripts.
func Anagrams(letters string) (matches []string, ok bool) {
	letters = normalize(letters)

	d := pick(letters)
	if d == nil {
		return nil, false
	}
	d.load()

	matches = append(matches, d.anagrams[anagramKey(letters)]...)
	slices.Sort(matches)
	return matches, true
}

// Mask returns dictionary words matching a pattern where "-", "_" and "*"
// stand for exactly one letter.
func Mask(pattern string) (matches []string, ok bool) {
	pattern = normalize(pattern)
	pattern = strings.NewReplacer("-", "*", "_", "*").Replace(pattern)

	letters := strings.ReplaceAll(pattern, "*", "")
	d := ru
	if letters != "" {
		if d = pick(strings.ReplaceAll(letters, "*", "")); d == nil {
			return nil, false
		}
	}
	d.load()

	var re strings.Builder
	re.WriteString("^")
	for _, r := range pattern {
		if r == '*' {
			re.WriteString(".")
		} else {
			re.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	re.WriteString("$")

	maskRe, err := regexp.Compile(re.String())
	if err != nil {
		return nil, false
	}

	for _, word := range d.words {
		if maskRe.MatchString(word) {
			matches = append(matches, word)
		}
	}
	return matches, true
}
