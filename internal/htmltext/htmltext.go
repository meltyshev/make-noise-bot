// Package htmltext converts engine HTML into the tag subset Telegram's HTML
// parse mode understands.
package htmltext

import (
	"html"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	xhtml "golang.org/x/net/html"
)

var (
	// Levels are pasted from word processors, so &nbsp; turns up between
	// coordinates and as the only content of a paragraph. It counts as
	// whitespace here, or it survives every trim below.
	whitespaceRe     = regexp.MustCompile(`[\s\x{00A0}]+`)
	multiSpaceRe     = regexp.MustCompile(` {2,}`)
	spacesAfterNLRe  = regexp.MustCompile(`\n +`)
	spacesBeforeNLRe = regexp.MustCompile(` +\n`)
	extraNewlinesRe  = regexp.MustCompile(`\n\n\n+`)
	tagRe            = regexp.MustCompile(`<[^>]*>`)
)

// inlineTags maps every inline tag an organizer may write onto the canonical
// Telegram set.
var inlineTags = map[string]string{
	"b": "b", "strong": "b",
	"i": "i", "em": "i",
	"u": "u", "ins": "u",
	"s": "s", "strike": "s", "del": "s",
	"code": "code",
}

type openTag struct {
	name string
	raw  string
}

// Convert renders an HTML fragment as Telegram-HTML text, resolving relative
// URLs against baseURL. The output always has balanced tags, whatever the
// input looks like.
func Convert(fragment, baseURL string) string {
	var (
		b          strings.Builder
		stack      []openTag
		hideOutput bool
	)

	base, _ := url.Parse(baseURL)
	makeURL := func(path string) string {
		path = strings.TrimSuffix(path, `\"`)
		if base == nil {
			return path
		}
		ref, err := url.Parse(path)
		if err != nil {
			return path
		}
		return base.ResolveReference(ref).String()
	}

	attr := func(tok xhtml.Token, name string) (string, bool) {
		for _, a := range tok.Attr {
			if a.Key == name {
				return a.Val, true
			}
		}
		return "", false
	}

	// Tags carry no width, so spacing is decided from the last character of
	// the text written so far, not from the builder's tail.
	var (
		lastByte  byte
		needSpace bool
	)
	writeText := func(escaped string) {
		if escaped == "" {
			return
		}
		if needSpace && escaped[0] != ' ' {
			b.WriteString(" ")
		}
		needSpace = false
		b.WriteString(escaped)
		lastByte = escaped[len(escaped)-1]
	}
	writeNewline := func() {
		b.WriteString("\n")
		lastByte = '\n'
		needSpace = false
	}
	// writeURL stands an image in for its address, spaced off the text around
	// it only where that text does not already provide the gap.
	writeURL := func(u string) {
		if lastByte != 0 && lastByte != ' ' && lastByte != '\n' {
			b.WriteString(" ")
		}
		b.WriteString(u)
		lastByte = u[len(u)-1]
		needSpace = true
	}

	open := func(t openTag) {
		stack = append(stack, t)
		b.WriteString(t.raw)
	}
	closeTop := func() {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		b.WriteString("</" + top.name + ">")
	}
	// closeTag closes the named tag if it is open, temporarily closing and
	// reopening anything above it so the output stays properly nested.
	closeTag := func(name string) {
		idx := -1
		for i, v := range slices.Backward(stack) {
			if v.name == name {
				idx = i
				break
			}
		}
		if idx < 0 {
			return
		}
		reopen := append([]openTag{}, stack[idx+1:]...)
		for len(stack) > idx {
			closeTop()
		}
		for _, t := range reopen {
			open(t)
		}
	}

	z := xhtml.NewTokenizer(strings.NewReader(fragment))
	for {
		tt := z.Next()
		if tt == xhtml.ErrorToken {
			break
		}
		tok := z.Token()

		switch tt {
		case xhtml.StartTagToken:
			switch tok.Data {
			case "script", "style":
				hideOutput = true
			case "div", "p", "br":
				if !hideOutput {
					writeNewline()
				}
			case "a":
				if href, ok := attr(tok, "href"); ok {
					// Telegram forbids nested links and rejects messages
					// with link protocols it does not know.
					closeTag("a")
					if target := makeURL(href); supportedLink(target) {
						open(openTag{name: "a", raw: `<a href="` + html.EscapeString(target) + `">`})
					}
				}
			case "img":
				if src, ok := attr(tok, "src"); ok {
					writeURL(html.EscapeString(makeURL(src)))
				}
			default:
				if name, ok := inlineTags[tok.Data]; ok {
					open(openTag{name: name, raw: "<" + name + ">"})
				}
			}

		case xhtml.SelfClosingTagToken:
			switch tok.Data {
			case "br":
				writeNewline()
			case "img":
				if src, ok := attr(tok, "src"); ok {
					writeURL(html.EscapeString(makeURL(src)))
				}
			}

		case xhtml.EndTagToken:
			switch tok.Data {
			case "script", "style":
				hideOutput = false
			case "div", "p":
				writeNewline()
			case "a":
				closeTag("a")
			default:
				if name, ok := inlineTags[tok.Data]; ok {
					closeTag(name)
				}
			}

		case xhtml.TextToken:
			if !hideOutput {
				writeText(html.EscapeString(whitespaceRe.ReplaceAllString(tok.Data, " ")))
			}
		}
	}

	for len(stack) > 0 {
		closeTop()
	}

	text := multiSpaceRe.ReplaceAllString(b.String(), " ")
	text = spacesAfterNLRe.ReplaceAllString(text, "\n")
	text = spacesBeforeNLRe.ReplaceAllString(text, "\n")
	// Organizers pad levels with rows of <br>; one blank line is enough.
	text = extraNewlinesRe.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(unwrapSelfLinks(text))
}

// unwrapSelfLinks turns an anchor whose whole text is its own target back into
// a bare URL. A picture wrapped in a link to itself converts to exactly that,
// and Telegram previews a plain URL where it leaves the anchor alone.
func unwrapSelfLinks(text string) string {
	tokens := tokenize(text)

	var b strings.Builder
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.isTag && !tok.closing && tok.name == "a" && i+2 < len(tokens) {
			inner, end := tokens[i+1], tokens[i+2]
			href, ok := hrefOf(tok.raw)
			if ok && !inner.isTag && end.isTag && end.closing && end.name == "a" &&
				strings.TrimSpace(inner.raw) == href {
				b.WriteString(href)
				i += 2
				continue
			}
		}
		b.WriteString(tok.raw)
	}
	return b.String()
}

// supportedLink keeps geo: links, which LinkCoordinates rewrites later.
func supportedLink(target string) bool {
	lower := strings.ToLower(target)
	for _, scheme := range []string{"http://", "https://", "tg://", "geo:"} {
		if strings.HasPrefix(lower, scheme) {
			return true
		}
	}
	return false
}

// StripTags returns the plain text of a converted fragment.
func StripTags(s string) string {
	return html.UnescapeString(tagRe.ReplaceAllString(s, ""))
}

// Split cuts a converted fragment into parts of at most limit UTF-16 units
// (Telegram counts message length that way), keeping tags balanced in every
// part: open tags are closed at a cut and reopened in the next part. Cuts
// prefer newlines, then spaces, and never land inside a tag or an entity.
func Split(s string, limit int) []string {
	if utf16Len(s) <= limit {
		return []string{s}
	}

	var (
		parts []string
		cur   strings.Builder
		open  []openTag
	)
	curLen := 0
	prefixLen := 0

	closersLen := func() int {
		n := 0
		for _, t := range open {
			n += utf16Len("</" + t.name + ">")
		}
		return n
	}

	flush := func() {
		part := cur.String()
		var b strings.Builder
		b.WriteString(part)
		for _, v := range slices.Backward(open) {
			b.WriteString("</" + v.name + ">")
		}
		if strings.TrimSpace(StripTags(b.String())) != "" {
			parts = append(parts, strings.TrimSpace(b.String()))
		}

		cur.Reset()
		curLen = 0
		for _, t := range open {
			cur.WriteString(t.raw)
			curLen += utf16Len(t.raw)
		}
		prefixLen = curLen
	}

	for _, tok := range tokenize(s) {
		if tok.isTag {
			tagLen := utf16Len(tok.raw)
			if curLen+tagLen+closersLen() > limit && curLen > prefixLen {
				flush()
			}
			cur.WriteString(tok.raw)
			curLen += tagLen
			if tok.closing {
				for i, v := range slices.Backward(open) {
					if v.name == tok.name {
						open = append(open[:i], open[i+1:]...)
						break
					}
				}
			} else {
				open = append(open, openTag{name: tok.name, raw: tok.raw})
			}
			continue
		}

		text := tok.raw
		for text != "" {
			budget := limit - curLen - closersLen()
			fit, rest := cutText(text, budget)
			if fit == "" {
				if curLen > prefixLen {
					flush()
					continue
				}
				// Pathological limit: force one rune to guarantee progress.
				_, size := utf8.DecodeRuneInString(text)
				fit, rest = text[:size], text[size:]
			}
			cur.WriteString(fit)
			curLen += utf16Len(fit)
			text = strings.TrimLeft(rest, " ")
			if rest != "" && curLen > prefixLen {
				flush()
			}
		}
	}
	flush()

	if len(parts) == 0 {
		return []string{strings.TrimSpace(s)}
	}
	return parts
}

type htmlToken struct {
	raw     string
	isTag   bool
	name    string
	closing bool
}

// tokenize scans converted output, where every "<" starts a real tag.
func tokenize(s string) []htmlToken {
	var out []htmlToken
	for s != "" {
		if s[0] == '<' {
			end := strings.IndexByte(s, '>')
			if end < 0 {
				out = append(out, htmlToken{raw: s})
				break
			}
			raw := s[:end+1]
			name := strings.TrimPrefix(strings.TrimPrefix(raw, "<"), "/")
			if idx := strings.IndexAny(name, " >"); idx >= 0 {
				name = name[:idx]
			}
			out = append(out, htmlToken{
				raw:     raw,
				isTag:   true,
				name:    name,
				closing: strings.HasPrefix(raw, "</"),
			})
			s = s[end+1:]
			continue
		}

		idx := strings.IndexByte(s, '<')
		if idx < 0 {
			idx = len(s)
		}
		out = append(out, htmlToken{raw: s[:idx]})
		s = s[idx:]
	}
	return out
}

// cutText returns the longest prefix within budget, preferring to cut at a
// newline, then at a space, and stepping back from a half-consumed entity.
func cutText(text string, budget int) (fit, rest string) {
	if budget <= 0 {
		return "", text
	}

	used := 0
	end := len(text)
	lastNewline, lastSpace := -1, -1
	for i, r := range text {
		width := 1
		if r > 0xFFFF {
			width = 2
		}
		if used+width > budget {
			end = i
			break
		}
		used += width
		switch r {
		case '\n':
			lastNewline = i
		case ' ':
			lastSpace = i
		}
	}
	if end == len(text) {
		return text, ""
	}

	cut := end
	if lastNewline > 0 {
		cut = lastNewline
	} else if lastSpace > 0 {
		cut = lastSpace
	} else {
		// Do not cut a "&...;" entity in half.
		if amp := strings.LastIndexByte(text[:cut], '&'); amp >= 0 && !strings.ContainsRune(text[amp:cut], ';') {
			cut = amp
		}
	}
	return text[:cut], text[cut:]
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}
