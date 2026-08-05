package bot

import (
	"encoding/json"
	"slices"
	"strings"
	"unicode"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

type formatPreset struct {
	Label   string
	Formats [][]string
}

var formatPresets = []formatPreset{
	{texts.PresetDigitsOnly, [][]string{{""}}},
	{texts.PresetDR, [][]string{{"dr", "др", "--"}}},
	{texts.PresetMoscow, [][]string{{"dr", "др"}, {"rd", "рд"}, {"d", "д"}, {"r", "р"}}},
}

func formatsEqual(a, b [][]string) bool {
	return slices.EqualFunc(a, b, slices.Equal)
}

// formatsLabel shows a preset name, the manual syntax, or JSON as a last
// resort for configs the friendly syntax cannot express.
func formatsLabel(formats [][]string) string {
	for _, preset := range formatPresets {
		if formatsEqual(formats, preset.Formats) {
			return preset.Label
		}
	}

	var groups []string
	for _, group := range formats {
		for _, variant := range group {
			if variant == "" || strings.ContainsAny(variant, ",=") {
				raw, _ := json.Marshal(formats)
				return string(raw)
			}
		}
		groups = append(groups, strings.Join(group, "="))
	}
	return strings.Join(groups, ", ")
}

// parseCodeFormats understands the manual syntax ("dr=др=--, rd=рд") and,
// for compatibility, JSON when the input starts with "[".
func parseCodeFormats(input string) ([][]string, bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, false
	}

	if strings.HasPrefix(input, "[") {
		var formats [][]string
		if json.Unmarshal([]byte(input), &formats) != nil {
			return nil, false
		}
		if formats == nil {
			formats = [][]string{}
		}
		return formats, true
	}

	var formats [][]string
	for _, rawGroup := range strings.Split(input, ",") {
		var group []string
		for _, rawVariant := range strings.Split(rawGroup, "=") {
			variant := strings.ToLower(strings.TrimSpace(rawVariant))
			if variant == "" || strings.ContainsFunc(variant, unicode.IsDigit) {
				return nil, false
			}
			group = append(group, variant)
		}
		formats = append(formats, group)
	}
	return formats, true
}

// applyFormats updates the config and, when a game is running, the game
// itself: format changes take effect immediately.
func (a *App) applyFormats(formats [][]string) error {
	return a.store.Update(func(d *store.Data) {
		d.GameConfig.CodeFormats = formats
		if d.Game != nil {
			copied := make([][]string, len(formats))
			for i, group := range formats {
				copied[i] = append([]string{}, group...)
			}
			d.Game.CodeFormats = copied
		}
	})
}
