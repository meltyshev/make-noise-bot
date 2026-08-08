// Package secret hides sensitive values from text that reaches logs, chats
// or the console.
package secret

import (
	"slices"
	"strings"
	"sync"
)

const mask = "[redacted]"

var (
	mu     sync.RWMutex
	values []string
)

// Register marks a value to be hidden by Redact. Short values are ignored so
// a stray registration cannot blank out unrelated text.
func Register(value string) {
	if len(value) < 8 {
		return
	}

	mu.Lock()
	defer mu.Unlock()
	if slices.Contains(values, value) {
		return
	}
	values = append(values, value)
}

func Redact(s string) string {
	mu.RLock()
	defer mu.RUnlock()

	for _, value := range values {
		s = strings.ReplaceAll(s, value, mask)
	}
	return s
}
