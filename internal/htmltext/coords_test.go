package htmltext

import (
	"fmt"
	"strings"
	"testing"
)

func testLink(lat, lon float64) string {
	return fmt.Sprintf("https://maps.example/%.4f/%.4f", lat, lon)
}

func TestLinkCoordinates(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "plain coordinates become a link",
			text: "Идите на 56.838011, 60.597465 и ищите код",
			want: `Идите на <a href="https://maps.example/56.8380/60.5975">56.838011, 60.597465</a> и ищите код`,
		},
		{
			name: "formatting around coordinates survives",
			text: "<b>Точка:</b> 56.838011, 60.597465",
			want: `<b>Точка:</b> <a href="https://maps.example/56.8380/60.5975">56.838011, 60.597465</a>`,
		},
		{
			name: "escaped quotes in degrees are handled",
			text: `56°50&#39;16.8&#34;N 60°35&#39;50.9&#34;E`,
			want: `<a href="https://maps.example/56.8380/60.5975">56°50&#39;16.8&#34;N 60°35&#39;50.9&#34;E</a>`,
		},
		{
			name: "text without coordinates is untouched",
			text: "Код 1.2, сектор 3.4 и время 0:15:00",
			want: "Код 1.2, сектор 3.4 и время 0:15:00",
		},
		{
			name: "coordinates inside a link are left alone",
			text: `<a href="https://yandex.ru/maps">56.838011, 60.597465</a>`,
			want: `<a href="https://yandex.ru/maps">56.838011, 60.597465</a>`,
		},
		{
			name: "monospace blocks are left alone",
			text: "<pre>56.838011, 60.597465</pre><code>55.755814, 37.617635</code>",
			want: "<pre>56.838011, 60.597465</pre><code>55.755814, 37.617635</code>",
		},
		{
			name: "geo links point at the map service",
			text: `<a href="geo:56.838011,60.597465">точка старта</a>`,
			want: `<a href="https://maps.example/56.8380/60.5975">точка старта</a>`,
		},
		{
			name: "geo links with parameters",
			text: `<a href="geo:56.838011,60.597465;u=35">точка</a>`,
			want: `<a href="https://maps.example/56.8380/60.5975">точка</a>`,
		},
		{
			name: "unreadable geo links lose the href",
			text: `<a href="geo:непонятно">точка</a>`,
			want: `<a>точка</a>`,
		},
		{
			name: "several coordinates in one text",
			text: "Старт 56.838011, 60.597465 финиш 55.755814, 37.617635",
			want: `Старт <a href="https://maps.example/56.8380/60.5975">56.838011, 60.597465</a> финиш <a href="https://maps.example/55.7558/37.6176">55.755814, 37.617635</a>`,
		},
		{
			name: "ampersands stay escaped",
			text: "Кафе &amp; бар 56.838011, 60.597465",
			want: `Кафе &amp; бар <a href="https://maps.example/56.8380/60.5975">56.838011, 60.597465</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LinkCoordinates(tt.text, testLink); got != tt.want {
				t.Errorf("LinkCoordinates(%q):\n got %q\nwant %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestLinks(t *testing.T) {
	text := `до <a href="https://example.com/one">линк</a> и картинка https://example.com/two.png, ` +
		`и <a href="https://example.com/three">еще</a>`

	links := Links(text)
	want := []string{"https://example.com/one", "https://example.com/two.png", "https://example.com/three"}
	if len(links) != len(want) {
		t.Fatalf("Links = %q, want %q", links, want)
	}
	for i, link := range links {
		if link != want[i] {
			t.Errorf("link %d = %q, want %q", i, link, want[i])
		}
	}
}

func TestLinksSkipCoordinateLinks(t *testing.T) {
	converted := Convert(`<p>55.058638, 82.974920</p><p><img src="https://example.com/pic.png"/></p>`, "https://example.com")

	own := Links(converted)
	if len(own) != 1 || own[0] != "https://example.com/pic.png" {
		t.Fatalf("own links = %q", own)
	}

	// After linking, the map URL is in the text but was never in own.
	linked := LinkCoordinates(converted, testLink)
	if !strings.Contains(linked, "maps.example") {
		t.Fatal("coordinates were not linked")
	}
	for _, link := range own {
		if strings.Contains(link, "maps.example") {
			t.Errorf("a coordinate link leaked into own links: %q", link)
		}
	}
}

func TestLinksHandleEntities(t *testing.T) {
	text := `<a href="https://example.com/a?x=1&amp;y=2">линк</a> и https://example.com/b?p=1&amp;q=2`

	links := Links(text)
	if len(links) != 2 {
		t.Fatalf("Links = %q, want 2", links)
	}
	if links[0] != "https://example.com/a?x=1&y=2" || links[1] != "https://example.com/b?p=1&q=2" {
		t.Errorf("Links = %q, want unescaped urls", links)
	}
}

func TestLinkCoordinatesWithoutLinker(t *testing.T) {
	text := "56.838011, 60.597465"
	if got := LinkCoordinates(text, nil); got != text {
		t.Errorf("nil linker changed the text: %q", got)
	}
}

func TestLinkedTextStaysBalanced(t *testing.T) {
	text := Convert("<b>Точка 56.838011, 60.597465 тут</b>", "https://example.com")
	linked := LinkCoordinates(text, testLink)

	if strings.Count(linked, "<a ") != strings.Count(linked, "</a>") {
		t.Errorf("unbalanced links: %q", linked)
	}
	for _, part := range Split(linked, 60) {
		balanced(t, part)
	}
}

func TestConvertDropsUnsupportedSchemes(t *testing.T) {
	tests := []struct {
		name     string
		fragment string
		want     string
	}{
		{name: "javascript", fragment: `<a href="javascript:alert(1)">клик</a>`, want: "клик"},
		{name: "tel", fragment: `<a href="tel:+79001234567">звонок</a>`, want: "звонок"},
		{name: "geo stays", fragment: `<a href="geo:56.8,60.5">точка</a>`, want: `<a href="geo:56.8,60.5">точка</a>`},
		{name: "relative becomes absolute", fragment: `<a href="/level">уровень</a>`, want: `<a href="https://example.com/level">уровень</a>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Convert(tt.fragment, "https://example.com"); got != tt.want {
				t.Errorf("Convert(%q) = %q, want %q", tt.fragment, got, tt.want)
			}
		})
	}
}
