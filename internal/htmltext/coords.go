package htmltext

import (
	"html"
	"strings"

	"github.com/meltyshev/make-noise-bot/internal/geo"
)

// LinkCoordinates turns coordinates in a converted fragment into map links
// and rewrites geo: links, which Telegram rejects, to the same map service.
// Text inside links and monospace blocks is left alone.
func LinkCoordinates(text string, link func(lat, lon float64) string) string {
	if link == nil {
		return text
	}

	var (
		out   strings.Builder
		depth int
	)
	for _, tok := range tokenize(text) {
		if !tok.isTag {
			if depth > 0 {
				out.WriteString(tok.raw)
				continue
			}
			out.WriteString(linkPlainText(tok.raw, link))
			continue
		}

		switch tok.name {
		case "a":
			if tok.closing {
				depth--
			} else {
				depth++
				out.WriteString(rewriteGeoHref(tok.raw, link))
				continue
			}
		case "code", "pre":
			if tok.closing {
				depth--
			} else {
				depth++
			}
		}
		out.WriteString(tok.raw)
	}
	return out.String()
}

func linkPlainText(escaped string, link func(lat, lon float64) string) string {
	plain := html.UnescapeString(escaped)

	matches := geo.Find(plain)
	if len(matches) == 0 {
		return escaped
	}

	var (
		out  strings.Builder
		last int
	)
	for _, match := range matches {
		out.WriteString(html.EscapeString(plain[last:match.Start]))
		out.WriteString(`<a href="` + html.EscapeString(link(match.Lat, match.Lon)) + `">`)
		out.WriteString(html.EscapeString(plain[match.Start:match.End]))
		out.WriteString("</a>")
		last = match.End
	}
	out.WriteString(html.EscapeString(plain[last:]))
	return out.String()
}

// rewriteGeoHref replaces a geo: URI with a map link, keeping the anchor
// text. Links that carry no readable coordinates lose their href.
func rewriteGeoHref(tag string, link func(lat, lon float64) string) string {
	href, ok := hrefOf(tag)
	if !ok || !strings.HasPrefix(strings.ToLower(href), "geo:") {
		return tag
	}

	lat, lon, ok := geo.ParseURI(html.UnescapeString(href))
	if !ok {
		return "<a>"
	}
	return `<a href="` + html.EscapeString(link(lat, lon)) + `">`
}

func hrefOf(tag string) (string, bool) {
	idx := strings.Index(tag, `href="`)
	if idx < 0 {
		return "", false
	}
	rest := tag[idx+len(`href="`):]
	end := strings.IndexByte(rest, '"')
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}
