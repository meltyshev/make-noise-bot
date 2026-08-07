// Package geo finds coordinates in text and turns them into map links.
package geo

import (
	"fmt"
	"strconv"
)

type Service string

const (
	Yandex Service = "yandex"
	Google Service = "google"
	TwoGIS Service = "2gis"
	OSM    Service = "osm"

	DefaultService = Yandex
)

// Services is the order shown in the settings menu.
var Services = []Service{Yandex, Google, TwoGIS, OSM}

func (s Service) Valid() bool {
	for _, known := range Services {
		if known == s {
			return true
		}
	}
	return false
}

func (s Service) Label() string {
	switch s {
	case Yandex:
		return "Яндекс.Карты"
	case Google:
		return "Google Maps"
	case TwoGIS:
		return "2ГИС"
	case OSM:
		return "OpenStreetMap"
	default:
		return string(s)
	}
}

// Link builds a web link that opens the point, and the map app on phones.
func (s Service) Link(lat, lon float64) string {
	latText, lonText := format(lat), format(lon)

	switch s {
	case Google:
		return fmt.Sprintf("https://maps.google.com/?q=%s,%s", latText, lonText)
	case TwoGIS:
		return fmt.Sprintf("https://2gis.ru/geo/%s,%s", lonText, latText)
	case OSM:
		return fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s#map=17/%s/%s", latText, lonText, latText, lonText)
	default:
		return fmt.Sprintf("https://yandex.ru/maps/?pt=%s,%s&z=17&l=map", lonText, latText)
	}
}

// Linker returns the link builder for a service name, falling back to the
// default when the setting is empty or unknown.
func Linker(service string) func(lat, lon float64) string {
	chosen := Service(service)
	if !chosen.Valid() {
		chosen = DefaultService
	}
	return chosen.Link
}

func format(value float64) string {
	return strconv.FormatFloat(value, 'f', 6, 64)
}
