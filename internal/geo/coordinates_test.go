package geo

import (
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 0.0001 }

func TestFindDecimal(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantLat float64
		wantLon float64
	}{
		{"comma and space", "Точка 56.838011, 60.597465 рядом", 56.838011, 60.597465},
		{"comma only", "56.838011,60.597465", 56.838011, 60.597465},
		{"space only", "56.838011 60.597465", 56.838011, 60.597465},
		{"semicolon", "56.838011; 60.597465", 56.838011, 60.597465},
		{"ends a sentence", "Идите на 56.838011, 60.597465.", 56.838011, 60.597465},
		{"latin hemispheres", "N 56.838011 E 60.597465", 56.838011, 60.597465},
		{"southern and western", "S 33.868800, W 151.209300", -33.8688, -151.2093},
		{"cyrillic hemispheres", "с 56.838011, в 60.597465", 56.838011, 60.597465},
		{"degree signs", "56.838011° 60.597465°", 56.838011, 60.597465},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := Find(tt.text)
			if len(matches) != 1 {
				t.Fatalf("Find(%q) = %d matches, want 1", tt.text, len(matches))
			}
			if !near(matches[0].Lat, tt.wantLat) || !near(matches[0].Lon, tt.wantLon) {
				t.Errorf("Find(%q) = (%v, %v), want (%v, %v)", tt.text, matches[0].Lat, matches[0].Lon, tt.wantLat, tt.wantLon)
			}
		})
	}
}

func TestFindSexagesimal(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantLat float64
		wantLon float64
	}{
		{"degrees minutes seconds", `56°50'16.8"N 60°35'50.9"E`, 56.8380, 60.5975},
		{"hemispheres in front", `N 56°50'16.8" E 60°35'50.9"`, 56.8380, 60.5975},
		{"decimal minutes", `56°50.28' 60°35.84'`, 56.8380, 60.5973},
		{"comma separated", `56°50'16.8"N, 60°35'50.9"E`, 56.8380, 60.5975},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := Find(tt.text)
			if len(matches) != 1 {
				t.Fatalf("Find(%q) = %d matches, want 1", tt.text, len(matches))
			}
			if math.Abs(matches[0].Lat-tt.wantLat) > 0.001 || math.Abs(matches[0].Lon-tt.wantLon) > 0.001 {
				t.Errorf("Find(%q) = (%v, %v), want about (%v, %v)", tt.text, matches[0].Lat, matches[0].Lon, tt.wantLat, tt.wantLon)
			}
		})
	}
}

func TestFindRejectsNonCoordinates(t *testing.T) {
	for _, text := range []string{
		"Коды сложности: 1.2, 1.3",
		"Всего 12.5, принято 3.75",
		"Время 0:15:00, уровень 3",
		"Телефон 8-900-123-45-67",
		"Версия 1.10.2024",
		"95.838011, 60.597465",   // latitude out of range
		"56.838011, 190.597465",  // longitude out of range
		"0.000000, 0.000000",     // null island
		"1156.838011, 60.597465", // part of a longer number
		"56.83, 60.59",           // too few decimals to be a fix
	} {
		if matches := Find(text); len(matches) != 0 {
			t.Errorf("Find(%q) = %+v, want none", text, matches)
		}
	}
}

func TestFindSeveralAndOffsets(t *testing.T) {
	text := "Старт 56.838011, 60.597465 финиш 55.755814, 37.617635"

	matches := Find(text)
	if len(matches) != 2 {
		t.Fatalf("matches = %d, want 2", len(matches))
	}
	if matches[0].Start > matches[1].Start {
		t.Error("matches are not in order")
	}
	if got := text[matches[0].Start:matches[0].End]; got != "56.838011, 60.597465" {
		t.Errorf("first span = %q, want the first coordinate pair", got)
	}
	if got := text[matches[1].Start:matches[1].End]; got != "55.755814, 37.617635" {
		t.Errorf("second span = %q, want the second coordinate pair", got)
	}
}

func TestParseURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		wantLat float64
		wantLon float64
		wantOK  bool
	}{
		{name: "plain", uri: "geo:56.838011,60.597465", wantLat: 56.838011, wantLon: 60.597465, wantOK: true},
		{name: "short", uri: "geo:56.83,60.59", wantLat: 56.83, wantLon: 60.59, wantOK: true},
		{name: "uncertainty", uri: "geo:56.838011,60.597465;u=35", wantLat: 56.838011, wantLon: 60.597465, wantOK: true},
		{name: "query", uri: "geo:56.838011,60.597465?q=точка", wantLat: 56.838011, wantLon: 60.597465, wantOK: true},
		{name: "uppercase scheme", uri: "GEO:56.838011,60.597465", wantLat: 56.838011, wantLon: 60.597465, wantOK: true},
		{name: "other scheme", uri: "https://example.com"},
		{name: "not a pair", uri: "geo:broken"},
		{name: "latitude out of range", uri: "geo:956.8,60.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lon, ok := ParseURI(tt.uri)
			if ok != tt.wantOK {
				t.Fatalf("ParseURI(%q) ok = %v, want %v", tt.uri, ok, tt.wantOK)
			}
			if ok && (!near(lat, tt.wantLat) || !near(lon, tt.wantLon)) {
				t.Errorf("ParseURI(%q) = (%v, %v), want (%v, %v)", tt.uri, lat, lon, tt.wantLat, tt.wantLon)
			}
		})
	}
}
