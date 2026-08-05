// Package htmltext converts engine HTML into the tag subset Telegram's HTML
// parse mode understands.
package htmltext

import (
	"html"
	"net/url"
	"regexp"
	"strings"

	xhtml "golang.org/x/net/html"
)

var (
	spacesRe           = regexp.MustCompile(` +`)
	spacesAfterNLRe    = regexp.MustCompile(`\n +`)
	extraNewlinesRe    = regexp.MustCompile(`\n\n\n+`)
	trailingBackslashQ = `\"`
)

// Convert renders an HTML fragment as Telegram-HTML text, resolving relative
// URLs against baseURL.
func Convert(fragment, baseURL string) string {
	var (
		b          strings.Builder
		hideOutput bool
	)

	base, _ := url.Parse(baseURL)
	makeURL := func(path string) string {
		path = strings.TrimSuffix(path, trailingBackslashQ)
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
					b.WriteString("\n")
				}
			case "b", "strong", "i", "em":
				b.WriteString("<" + tok.Data + ">")
			case "a":
				if href, ok := attr(tok, "href"); ok {
					b.WriteString(`<a href="` + html.EscapeString(makeURL(href)) + `">`)
				}
			case "img":
				if src, ok := attr(tok, "src"); ok {
					b.WriteString(" " + html.EscapeString(makeURL(src)) + " ")
				}
			}

		case xhtml.SelfClosingTagToken:
			switch tok.Data {
			case "br":
				b.WriteString("\n")
			case "img":
				if src, ok := attr(tok, "src"); ok {
					b.WriteString(" " + html.EscapeString(makeURL(src)) + " ")
				}
			}

		case xhtml.EndTagToken:
			switch tok.Data {
			case "script", "style":
				hideOutput = false
			case "div", "p":
				b.WriteString("\n")
			case "b", "strong", "i", "em":
				b.WriteString("</" + tok.Data + ">")
			case "a":
				b.WriteString("</a>")
			}

		case xhtml.TextToken:
			if !hideOutput && tok.Data != "" {
				text := strings.TrimSpace(spacesRe.ReplaceAllString(tok.Data, " "))
				b.WriteString(html.EscapeString(text))
			}
		}
	}

	text := strings.TrimSpace(b.String())
	text = spacesAfterNLRe.ReplaceAllString(text, "\n")
	text = extraNewlinesRe.ReplaceAllString(text, "\n\n")
	return text
}
