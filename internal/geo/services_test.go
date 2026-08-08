package geo

import (
	"strings"
	"testing"
)

func TestServiceLinks(t *testing.T) {
	const lat, lon = 56.838011, 60.597465

	tests := []struct {
		name    string
		service Service
		want    string
	}{
		{
			name:    "yandex puts longitude first",
			service: Yandex,
			want:    "https://yandex.ru/maps/?pt=60.597465,56.838011&z=17&l=map",
		},
		{
			name:    "google takes a plain lat,lon query",
			service: Google,
			want:    "https://maps.google.com/?q=56.838011,60.597465",
		},
		{
			name:    "2gis puts longitude first too",
			service: TwoGIS,
			want:    "https://2gis.ru/geo/60.597465,56.838011",
		},
		{
			name:    "osm repeats the point in the fragment",
			service: OSM,
			want:    "https://www.openstreetmap.org/?mlat=56.838011&mlon=60.597465#map=17/56.838011/60.597465",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.service.Link(lat, lon); got != tt.want {
				t.Errorf("%s.Link(%v, %v) = %q, want %q", tt.service, lat, lon, got, tt.want)
			}
		})
	}
}

func TestServiceLinksAreHTTPS(t *testing.T) {
	// Telegram only makes http and https links clickable.
	for _, service := range Services {
		if link := service.Link(-33.8688, 151.2093); !strings.HasPrefix(link, "https://") {
			t.Errorf("%s.Link = %q, want an https link Telegram makes clickable", service, link)
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
			t.Errorf("%s.Valid() = false, want every service in Services to be valid", service)
		}
		if service.Label() == string(service) {
			t.Errorf("%s.Label() = %q, want a label distinct from the key", service, service.Label())
		}
	}
	if Service("nonsense").Valid() {
		t.Error("unknown service accepted")
	}
}
