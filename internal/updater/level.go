package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/meltyshev/make-noise-bot/internal/game"
)

// levelStartWindow is how long after a restart the timer still looks fresh.
const levelStartWindow = 120

// levelState is what a tick knows about the level the team is on.
type levelState struct {
	Number *int
	Task   string
	Time   *int
}

// isNewLevel votes on three independent signals, because the engine renumbers
// levels when an organizer deletes one and rewrites the task when they fix a
// typo. Two signals have to agree, so neither alone can announce a level.
func isNewLevel(previous, current levelState) bool {
	if current.Number == nil {
		return false
	}

	signals := 0
	if !intPtrEqual(previous.Number, current.Number) {
		signals++
	}
	if previous.Task != current.Task {
		signals++
	}
	if timerRestarted(previous.Time, current.Time) {
		signals++
	}
	return signals >= 2
}

// levelGone reports the moment the engine stops serving a level. A failed
// request never gets here, so an empty answer is the engine's own word.
func levelGone(previous, current levelState) bool {
	return previous.Number != nil && current.Number == nil
}

func sameLevel(previous, current levelState) bool {
	return intPtrEqual(previous.Number, current.Number) &&
		previous.Task == current.Task &&
		intPtrEqual(previous.Time, current.Time)
}

// timerRestarted looks for a timer that fell back to the start of a level,
// which a countdown ticking down does not do.
func timerRestarted(previous, current *int) bool {
	if previous == nil || current == nil {
		return false
	}
	return *current < *previous && *current <= levelStartWindow
}

// levelTask fingerprints what the team has to solve: the texts and the code
// layout, but not the codes they entered, the timer or the hints, which all
// change while the level stays the same.
func levelTask(snap game.Snapshot) string {
	if snap.LevelNumber() == nil {
		return ""
	}

	var text strings.Builder
	text.WriteString(snap.Question())
	for _, sector := range snap.Sectors() {
		text.WriteString("\x00" + sector.Name)
		for _, code := range sector.Codes {
			fmt.Fprintf(&text, "|%d:%s", code.Number, code.Hazard)
		}
	}

	digest := sha256.Sum256([]byte(text.String()))
	return hex.EncodeToString(digest[:])[:16]
}

func spoilerText(spoilers []game.Spoiler, number int) string {
	for _, spoiler := range spoilers {
		if spoiler.Number == number {
			return spoiler.Text
		}
	}
	return ""
}

func timeOnLevel(snap game.Snapshot) *int {
	seconds, ok := snap.TimeOnLevel()
	if !ok {
		return nil
	}
	return &seconds
}
