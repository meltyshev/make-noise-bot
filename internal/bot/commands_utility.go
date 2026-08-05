package bot

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/meltyshev/make-noise-bot/internal/avatar"
	"github.com/meltyshev/make-noise-bot/internal/texts"
	"github.com/meltyshev/make-noise-bot/internal/words"
)

// telegramTextLimit keeps long word lists under Telegram's 4096-char cap.
const telegramTextLimit = 4000

var morseEN = map[string]string{
	".-": "a", "-...": "b", "-.-.": "c", "-..": "d", ".": "e", "..-.": "f",
	"--.": "g", "....": "h", "..": "i", ".---": "j", "-.-": "k", ".-..": "l",
	"--": "m", "-.": "n", "---": "o", ".--.": "p", "--.-": "q", ".-.": "r",
	"...": "s", "-": "t", "..-": "u", "...-": "v", ".--": "w", "-..-": "x",
	"-.--": "y", "--..": "z",
}

var morseRU = map[string]string{
	".-": "а", "-...": "б", ".--": "в", "--.": "г", "-..": "д", ".": "е",
	"...-": "ж", "--..": "з", "..": "и", ".---": "й", "-.-": "к", ".-..": "л",
	"--": "м", "-.": "н", "---": "о", ".--.": "п", ".-.": "р", "...": "с",
	"-": "т", "..-": "у", "..-.": "ф", "....": "х", "-.-.": "ц", "---.": "ч",
	"----": "ш", "--.-": "щ", "--.--": "ъ", "-.--": "ы", "-..-": "ь",
	"..-..": "э", "..--": "ю", ".-.-": "я",
}

var morseCommon = map[string]string{
	"-----": "0", ".----": "1", "..---": "2", "...--": "3", "....-": "4",
	".....": "5", "-....": "6", "--...": "7", "---..": "8", "----.": "9",
	"......": ".", ".-.-.-": ",", "---...": ":", "-.-.-": ";", "-.--.-": "|",
	".----.": "'", ".-..-.": `"`, "-....-": "-", "-..-.": "/", "..--..": "?",
	"--..--": "!", ".--.-.": "@",
}

func cmdMorse() *Command {
	return askFlow("morse", texts.DescMorse, texts.MorseAsk, texts.MorseRequired, func(c *Ctx, input string) {
		input = strings.ReplaceAll(input, "_", "-")

		var wordEN, wordRU strings.Builder
		for _, letter := range strings.Fields(input) {
			if common, ok := morseCommon[letter]; ok {
				wordEN.WriteString(common)
				wordRU.WriteString(common)
				continue
			}
			if en, ok := morseEN[letter]; ok {
				wordEN.WriteString(en)
			} else {
				fmt.Fprintf(&wordEN, "(%s)", letter)
			}
			if ru, ok := morseRU[letter]; ok {
				wordRU.WriteString(ru)
			} else {
				fmt.Fprintf(&wordRU, "(%s)", letter)
			}
		}

		c.Reply(wordEN.String())
		c.Reply(wordRU.String())
	})
}

const lowercaseEN = "abcdefghijklmnopqrstuvwxyz"

// ё included, at position 7.
var lowercaseRU = []rune("абвгдеёжзийклмнопрстуфхцчшщъыьэюя")

func cmdNumbersToLetters() *Command {
	return askFlow("numberstoletters", texts.DescNumbersToLetters, texts.NumbersAsk, texts.NumbersRequired, func(c *Ctx, input string) {
		var wordEN, wordRU strings.Builder

		for _, token := range strings.Fields(input) {
			n, err := strconv.Atoi(token)
			if err != nil || len(token) > 2 || strings.ContainsAny(token, "+-") {
				fmt.Fprintf(&wordEN, "(%s)", token)
				fmt.Fprintf(&wordRU, "(%s)", token)
				continue
			}

			// "0" wraps around to the last letter.
			if n <= 26 {
				wordEN.WriteString(string(lowercaseEN[wrapIndex(n, len(lowercaseEN))]))
			} else {
				fmt.Fprintf(&wordEN, "(%d)", n)
			}
			if n <= 33 {
				wordRU.WriteString(string(lowercaseRU[wrapIndex(n, len(lowercaseRU))]))
			} else {
				fmt.Fprintf(&wordRU, "(%d)", n)
			}
		}

		c.Reply(wordEN.String())
		c.Reply(wordRU.String())
	})
}

func wrapIndex(n, length int) int {
	idx := n - 1
	if idx < 0 {
		idx += length
	}
	return idx
}

func cmdLettersToNumbers() *Command {
	return askFlow("letterstonumbers", texts.DescLettersToNumbers, texts.LettersAsk, texts.LettersRequired, func(c *Ctx, input string) {
		for _, word := range strings.Fields(input) {
			var numbers []string
			for _, letter := range strings.ToLower(word) {
				if idx := strings.IndexRune(lowercaseEN, letter); idx >= 0 {
					numbers = append(numbers, strconv.Itoa(idx+1))
				} else if idx := runeIndex(lowercaseRU, letter); idx >= 0 {
					numbers = append(numbers, strconv.Itoa(idx+1))
				} else {
					numbers = append(numbers, fmt.Sprintf("(%c)", letter))
				}
			}
			c.Reply(strings.Join(numbers, " "))
		}
	})
}

func runeIndex(runes []rune, r rune) int {
	for i, candidate := range runes {
		if candidate == r {
			return i
		}
	}
	return -1
}

func cmdIntersection() *Command {
	return askFlow("intersection", texts.DescIntersection, texts.IntersectionAsk, texts.IntersectionRequired, func(c *Ctx, input string) {
		fields := strings.Fields(input)
		if len(fields) < 2 {
			c.Reply(texts.IntersectionTooFew)
			return
		}

		sets := make([]map[rune]bool, len(fields))
		for i, word := range fields {
			sets[i] = map[rune]bool{}
			for _, r := range strings.ToLower(word) {
				sets[i][r] = true
			}
		}

		var intersection []rune
		seen := map[rune]bool{}
		for _, r := range strings.ToLower(fields[0]) {
			if seen[r] {
				continue
			}
			seen[r] = true

			inAll := true
			for _, set := range sets[1:] {
				if !set[r] {
					inAll = false
					break
				}
			}
			if inAll {
				intersection = append(intersection, r)
			}
		}

		if len(intersection) > 0 {
			c.Reply(string(intersection))
		} else {
			c.Reply(texts.IntersectionEmpty)
		}
	})
}

func isAlpha(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func cmdAnagram() *Command {
	return askFlow("anagram", texts.DescAnagram, texts.AnagramAsk, texts.AnagramRequired, func(c *Ctx, input string) {
		if !isAlpha(input) {
			c.Reply(texts.AnagramOnlyLetters)
			return
		}

		matches, ok := words.Anagrams(input)
		if !ok {
			c.Reply(texts.AnagramUnavailable)
			return
		}
		if len(matches) == 0 {
			c.Reply(texts.AnagramNotFound)
			return
		}
		c.Reply(joinLimited(matches))
	})
}

func cmdMask() *Command {
	return askFlow("mask", texts.DescMask, texts.MaskAsk, texts.MaskRequired, func(c *Ctx, input string) {
		for _, r := range input {
			if !unicode.IsLetter(r) && r != '-' && r != '_' {
				c.Reply(texts.MaskOnlyLetters)
				return
			}
		}

		matches, ok := words.Mask(input)
		if !ok {
			return
		}
		if len(matches) == 0 {
			c.Reply(texts.MaskNotFound)
			return
		}
		c.Reply(joinLimited(matches))
	})
}

func joinLimited(items []string) string {
	joined := strings.Join(items, ", ")
	if len(joined) <= telegramTextLimit {
		return joined
	}
	cut := strings.LastIndex(joined[:telegramTextLimit], ", ")
	if cut < 0 {
		cut = telegramTextLimit
	}
	return joined[:cut] + "..."
}

func cmdCoordinates() *Command {
	replyDMS := func(c *Ctx, latitude, longitude float64) {
		latDeg, latMin, latSec := toDMS(latitude)
		lngDeg, lngMin, lngSec := toDMS(longitude)

		c.Replyf(`%d°%02d'%05.2f"N %d°%02d'%05.2f"E`,
			latDeg, latMin, latSec,
			lngDeg, lngMin, lngSec,
		)
	}

	return &Command{
		Name: "coordinates",
		Init: func(c *Ctx, args string) {
			if !c.EnsureAllowed("coordinates") {
				return
			}
			if args != "" {
				c.DelConv()
				if latitude, longitude, ok := parseCoordinates(args); ok {
					replyDMS(c, latitude, longitude)
				} else {
					c.Reply(texts.CoordinatesRequired)
				}
				return
			}
			c.SetConv("coordinates")
			c.Reply(texts.CoordinatesAsk)
		},
		Handle: func(c *Ctx, _ any) {
			if c.msg.Location != nil {
				c.DelConv()
				replyDMS(c, c.msg.Location.Latitude, c.msg.Location.Longitude)
				return
			}
			if latitude, longitude, ok := parseCoordinates(c.Text()); ok {
				c.DelConv()
				replyDMS(c, latitude, longitude)
				return
			}
			c.Reply(texts.CoordinatesRequired)
		},
	}
}

func parseCoordinates(input string) (latitude, longitude float64, ok bool) {
	fields := strings.Fields(strings.ReplaceAll(input, ",", " "))
	if len(fields) != 2 {
		return 0, 0, false
	}

	latitude, latErr := strconv.ParseFloat(fields[0], 64)
	longitude, lngErr := strconv.ParseFloat(fields[1], 64)
	if latErr != nil || lngErr != nil {
		return 0, 0, false
	}
	return latitude, longitude, true
}

func toDMS(decimal float64) (int, int, float64) {
	degrees := math.Floor(decimal)
	minutes := math.Floor(60 * (decimal - degrees))
	seconds := math.Round(3600*((decimal-degrees)-minutes/60)*100) / 100
	return int(degrees), int(minutes), seconds
}

func cmdAvatar() *Command {
	run := func(c *Ctx, input string) {
		fields := strings.Fields(input)
		if len(fields) < 3 {
			c.Reply(texts.AvatarUsage)
			return
		}

		background, foreground := fields[0], fields[1]
		for _, nickname := range fields[2:] {
			png, err := avatar.Generate(background, foreground, nickname)
			if err != nil {
				c.Reply(texts.AvatarUsage)
				return
			}
			c.ReplyPhoto(png, "")
		}
	}

	return &Command{
		Name: "avatar",
		Init: func(c *Ctx, args string) {
			if !c.IsManager() {
				return
			}
			if args != "" {
				c.DelConv()
				run(c, args)
				return
			}
			c.SetConv("avatar")
			c.Reply(texts.AvatarAsk)
		},
		Handle: func(c *Ctx, _ any) {
			c.DelConv()
			run(c, c.Text())
		},
	}
}
