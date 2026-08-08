// Package secret hides sensitive values from text that reaches logs, chats
// or the console.
package secret

import (
	"errors"
	"fmt"
	"net/url"
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

// Redact replaces every registered value in s with a mask. Values below the
// Register threshold are not covered, which is what StripURL is for.
func Redact(s string) string {
	mu.RLock()
	defer mu.RUnlock()

	for _, value := range values {
		s = strings.ReplaceAll(s, value, mask)
	}
	return s
}

// StripURL drops the request URL from a transport error, keeping the operation
// and the cause. Both the bot token and the game password end up in that URL,
// and a value too short for Register to accept would survive Redact.
func StripURL(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("%s: %w", urlErr.Op, urlErr.Err)
	}
	return err
}
