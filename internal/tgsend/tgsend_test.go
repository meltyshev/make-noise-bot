package tgsend

import (
	"testing"

	"github.com/meltyshev/make-noise-bot/internal/geo"
	"github.com/meltyshev/make-noise-bot/internal/htmltext"
)

func TestPreviewPicksTheEnginesOwnLink(t *testing.T) {
	const image = "http://classic.dzzzr.ru/uploaded/zadan.png"

	converted := htmltext.Convert(
		`<p>55.058638, 82.974920</p><p><img src="`+image+`"/></p>`,
		"https://lite.dzzzr.ru/novosib/go/",
	)
	own := htmltext.Links(converted)
	linked := htmltext.LinkCoordinates(converted, geo.Linker(string(geo.Yandex)))

	options := preview(linked, own)
	if options.URL == nil {
		t.Fatalf("preview for %q = none, want the page's own link", linked)
	}
	if *options.URL != image {
		t.Errorf("preview = %q, want the image %q", *options.URL, image)
	}
}

func TestPreviewIsOffWithoutOwnLinks(t *testing.T) {
	converted := htmltext.Convert("<p>Точка 55.058638, 82.974920</p>", "https://lite.dzzzr.ru/novosib/go/")
	own := htmltext.Links(converted)
	linked := htmltext.LinkCoordinates(converted, geo.Linker(string(geo.Yandex)))

	options := preview(linked, own)
	if options.URL != nil {
		t.Errorf("preview = %q, want no coordinate link chosen", *options.URL)
	}
	if options.IsDisabled == nil || !*options.IsDisabled {
		t.Error("preview should be disabled when only our own links remain")
	}
}

func TestPreviewPerPart(t *testing.T) {
	own := []string{"https://example.com/first.png", "https://example.com/second.png"}

	if options := preview("текст https://example.com/second.png", own); options.URL == nil || *options.URL != own[1] {
		t.Errorf("part preview = %+v, want the link inside the part", options)
	}
	if options := preview("текст без ссылок", own); options.IsDisabled == nil || !*options.IsDisabled {
		t.Error("a part without links should have no preview")
	}
}
