package geo

import (
	"strings"
	"testing"
)

func TestServiceLinks(t *testing.T) {
	const lat, lon = 56.838011, 60.597465

	tests := map[Service]string{
		Yandex: "https://yandex.ru/maps/?pt=60.597465,56.838011&z=17&l=map",
		Google: "https://maps.google.com/?q=56.838011,60.597465",
		TwoGIS: "https://2gis.ru/geo/60.597465,56.838011",
		OSM:    "https://www.openstreetmap.org/?mlat=56.838011&mlon=60.597465#map=17/56.838011/60.597465",
	}

	for service, want := range tests {
		t.Run(string(service), func(t *testing.T) {
			if got := service.Link(lat, lon); got != want {
				t.Errorf("%s.Link(%v, %v) = %q, want %q", service, lat, lon, got, want)
			}
		})
	}
}

func TestServiceLinksAreHTTPS(t *testing.T) {
	// Telegram only makes http and https links clickable.
	for _, service := range Services {
		if link := service.Link(-33.8688, 151.2093); !strings.HasPrefix(link, "https://") {
			t.Errorf("%s link is not https: %q", service, link)
		}
	}
}

func TestLinkerFallsBackToDefault(t *testing.T) {
	want := DefaultService.Link(1.5, 2.5)

	for _, service := range []string{"", "nonsense"} {
		if got := Linker(service)(1.5, 2.5); got != want {
			t.Errorf("Linker(%q) = %q, want the default %q", service, got, want)
		}
	}
	if got := Linker(string(Google))(1.5, 2.5); got != Google.Link(1.5, 2.5) {
		t.Errorf("Linker(google) = %q, want %q", got, Google.Link(1.5, 2.5))
	}
}

func TestServicesHaveLabels(t *testing.T) {
	for _, service := range Services {
		if !service.Valid() {
			t.Errorf("%s missing from Services", service)
		}
		if service.Label() == string(service) {
			t.Errorf("%s has no label", service)
		}
	}
	if Service("nonsense").Valid() {
		t.Error("unknown service accepted")
	}
}
