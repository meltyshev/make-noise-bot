package geo

import (
	"regexp"
	"strconv"
	"strings"
)

// Match is a coordinate pair found in a text, with its byte offsets.
type Match struct {
	Start int
	End   int
	Lat   float64
	Lon   float64
}

// Level texts are full of numbers, so decimals need at least four fraction
// digits to count as coordinates. That rejects hazard codes like 1.2 and
// house numbers while accepting anything a GPS produces. The degree sign and
// the primes are input organizers actually type, not prose, which is why the
// ASCII-only rule does not reach them.
var (
	decimalRe     = regexp.MustCompile(`(?i)([nsсю])?\s*(-?\d{1,3}\.\d{4,})\s*°?\s*([nsсю])?\s*(?:[,;]\s*|\s+)([ewвз])?\s*(-?\d{1,3}\.\d{4,})\s*°?\s*([ewвз])?`)
	sexagesimalRe = regexp.MustCompile(`(?i)([nsсю])?\s*(\d{1,3})\s*°\s*(\d{1,2}(?:[.,]\d+)?)\s*['′]\s*(?:(\d{1,2}(?:[.,]\d+)?)\s*(?:["″]|''))?\s*([nsсю])?\s*(?:[,;]\s*|\s+)([ewвз])?\s*(\d{1,3})\s*°\s*(\d{1,2}(?:[.,]\d+)?)\s*['′]\s*(?:(\d{1,2}(?:[.,]\d+)?)\s*(?:["″]|''))?\s*([ewвз])?`)
)

// Find returns the coordinate pairs in a text, in order and without
// overlaps.
func Find(text string) []Match {
	matches := append(findSexagesimal(text), findDecimal(text)...)

	var kept []Match
	for _, match := range matches {
		if overlaps(kept, match) {
			continue
		}
		kept = append(kept, match)
	}

	for i := 1; i < len(kept); i++ {
		for j := i; j > 0 && kept[j].Start < kept[j-1].Start; j-- {
			kept[j], kept[j-1] = kept[j-1], kept[j]
		}
	}
	return kept
}

// ParseURI reads a geo: URI, where the numbers are known to be coordinates
// and need no plausibility rules.
func ParseURI(uri string) (lat, lon float64, ok bool) {
	body, found := strings.CutPrefix(strings.ToLower(uri), "geo:")
	if !found {
		return 0, 0, false
	}
	if idx := strings.IndexAny(body, ";?"); idx >= 0 {
		body = body[:idx]
	}

	parts := strings.Split(body, ",")
	if len(parts) < 2 {
		return 0, 0, false
	}
	if lat, ok = parseFloat(strings.TrimSpace(parts[0])); !ok {
		return 0, 0, false
	}
	if lon, ok = parseFloat(strings.TrimSpace(parts[1])); !ok {
		return 0, 0, false
	}
	if abs(lat) > 90 || abs(lon) > 180 {
		return 0, 0, false
	}
	return lat, lon, true
}

func findDecimal(text string) []Match {
	var found []Match
	for _, idx := range decimalRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := trimSpan(text, idx[0], idx[1])
		if !standsAlone(text, start, end) {
			continue
		}

		lat, ok := parseFloat(group(text, idx, 2))
		if !ok {
			continue
		}
		lon, ok := parseFloat(group(text, idx, 5))
		if !ok {
			continue
		}

		lat = applyHemisphere(lat, group(text, idx, 1), group(text, idx, 3))
		lon = applyHemisphere(lon, group(text, idx, 4), group(text, idx, 6))

		if match, ok := newMatch(start, end, lat, lon); ok {
			found = append(found, match)
		}
	}
	return found
}

func findSexagesimal(text string) []Match {
	var found []Match
	for _, idx := range sexagesimalRe.FindAllStringSubmatchIndex(text, -1) {
		start, end := trimSpan(text, idx[0], idx[1])
		if !standsAlone(text, start, end) {
			continue
		}

		lat, ok := degrees(group(text, idx, 2), group(text, idx, 3), group(text, idx, 4))
		if !ok {
			continue
		}
		lon, ok := degrees(group(text, idx, 7), group(text, idx, 8), group(text, idx, 9))
		if !ok {
			continue
		}

		lat = applyHemisphere(lat, group(text, idx, 1), group(text, idx, 5))
		lon = applyHemisphere(lon, group(text, idx, 6), group(text, idx, 10))

		if match, ok := newMatch(start, end, lat, lon); ok {
			found = append(found, match)
		}
	}
	return found
}

// newMatch keeps only plausible pairs. A latitude over 90 means the numbers
// are something else, and guessing that they are swapped would point at a
// confidently wrong place.
func newMatch(start, end int, lat, lon float64) (Match, bool) {
	if abs(lat) > 90 || abs(lon) > 180 {
		return Match{}, false
	}
	if lat == 0 && lon == 0 {
		return Match{}, false
	}
	return Match{Start: start, End: end, Lat: lat, Lon: lon}, true
}

// standsAlone rejects a match that continues a longer number. A trailing dot
// is fine: coordinates often end a sentence.
func standsAlone(text string, start, end int) bool {
	if start > 0 {
		previous := text[start-1]
		if isDigit(previous) || previous == '.' {
			return false
		}
	}
	if end < len(text) {
		if isDigit(text[end]) {
			return false
		}
		if text[end] == '.' && end+1 < len(text) && isDigit(text[end+1]) {
			return false
		}
	}
	return true
}

func isDigit(char byte) bool { return char >= '0' && char <= '9' }

// trimSpan drops the whitespace the optional hemisphere groups swallow, so a
// link covers the coordinates only.
func trimSpan(text string, start, end int) (int, int) {
	for start < end && isSpace(text[start]) {
		start++
	}
	for end > start && isSpace(text[end-1]) {
		end--
	}
	return start, end
}

func isSpace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}

func overlaps(kept []Match, candidate Match) bool {
	for _, match := range kept {
		if candidate.Start < match.End && match.Start < candidate.End {
			return true
		}
	}
	return false
}

func group(text string, idx []int, n int) string {
	if 2*n+1 >= len(idx) || idx[2*n] < 0 {
		return ""
	}
	return text[idx[2*n]:idx[2*n+1]]
}

func parseFloat(value string) (float64, bool) {
	parsed, err := strconv.ParseFloat(strings.Replace(value, ",", ".", 1), 64)
	return parsed, err == nil
}

func degrees(degreeText, minuteText, secondText string) (float64, bool) {
	degree, ok := parseFloat(degreeText)
	if !ok {
		return 0, false
	}
	minute, ok := parseFloat(minuteText)
	if !ok {
		return 0, false
	}
	second := 0.0
	if secondText != "" {
		if second, ok = parseFloat(secondText); !ok {
			return 0, false
		}
	}
	if minute >= 60 || second >= 60 {
		return 0, false
	}
	return degree + minute/60 + second/3600, true
}

func applyHemisphere(value float64, before, after string) float64 {
	letter := strings.ToLower(before + after)
	if strings.ContainsAny(letter, "sюwз") {
		return -abs(value)
	}
	return value
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
