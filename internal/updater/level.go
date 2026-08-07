package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

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

	digest := sha256.New()
	io.WriteString(digest, snap.Question())
	io.WriteString(digest, "\x00")
	io.WriteString(digest, snap.Notes())
	for _, sector := range snap.Sectors() {
		io.WriteString(digest, "\x00"+sector.Name)
		for _, code := range sector.Codes {
			fmt.Fprintf(digest, "|%d:%s", code.Number, code.Hazard)
		}
	}
	return hex.EncodeToString(digest.Sum(nil))[:16]
}

func timeOnLevel(snap game.Snapshot) *int {
	seconds, ok := snap.TimeOnLevel()
	if !ok {
		return nil
	}
	return &seconds
}
