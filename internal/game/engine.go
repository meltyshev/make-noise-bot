// Package game talks to the Dozor game engines.
package game

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/meltyshev/make-noise-bot/internal/texts"
)

const (
	requestTimeout = 22 * time.Second
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"
)

type EnterCodeResult struct {
	Message    string
	StatusCode int // 0 when the engine gave no recognizable status
	Accepted   bool
}

type SectorCode struct {
	Number  int
	Hazard  string
	Entered bool
}

type Sector struct {
	Name  string
	Codes []SectorCode
}

// Spoiler is one spoiler of a level. Text is filled once it opens, and only
// by engines that publish it.
type Spoiler struct {
	Number int
	Open   bool
	Text   string
}

// Snapshot is one loaded view of the engine state.
type Snapshot interface {
	LevelNumber() *int        // nil when there is no level
	Progress() string         // "" when unavailable
	Question() string         // "" when unavailable
	Notes() string            // "" when unavailable
	Sectors() []Sector        // nil when unavailable
	Hint() (int, string)      // 0, "" when there is no hint
	Spoilers() []Spoiler      // nil when unsupported or no level
	TimeOnLevel() (int, bool) // seconds, false when the engine hides it
}

type Engine interface {
	Name() string
	Link() string
	// EnterCode never fails hard: transport errors become the result message.
	EnterCode(ctx context.Context, code string, pinnedLevel *int) EnterCodeResult
	Load(ctx context.Context) (Snapshot, error)
}

// Env carries the HTTP setup shared by the engines.
type Env struct {
	ClassicBaseURL string
	LiteBaseURL    string

	// HTTP must not follow redirects: engine answers are err= codes in 302
	// Location headers. HTTPFollow is for the login API only.
	HTTP       *http.Client
	HTTPFollow *http.Client

	// Debug receives raw engine payloads that failed to parse.
	Debug func(kind string, body []byte)

	// OnSessionUpdate is called after an automatic re-login.
	OnSessionUpdate func(session string)
}

func DefaultEnv() *Env {
	return &Env{
		ClassicBaseURL: "https://classic.dzzzr.ru",
		LiteBaseURL:    "https://lite.dzzzr.ru",
		HTTP: &http.Client{
			Timeout: requestTimeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		HTTPFollow: &http.Client{Timeout: requestTimeout},
	}
}

func (e *Env) debug(kind string, body []byte) {
	if e.Debug != nil {
		e.Debug(kind, body)
	}
}

func (e *Env) sessionUpdated(session string) {
	if e.OnSessionUpdate != nil {
		e.OnSessionUpdate(session)
	}
}

func statusCodeFromLocation(location string) int {
	u, err := url.Parse(location)
	if err != nil {
		return 0
	}
	code, err := strconv.Atoi(u.Query().Get("err"))
	if err != nil {
		return 0
	}
	return code
}

func resultFromStatus(statuses map[int]string, accepted map[int]bool, location string) EnterCodeResult {
	code := statusCodeFromLocation(location)
	message, ok := statuses[code]
	if !ok {
		message = texts.EngineUnknown
	}
	return EnterCodeResult{Message: message, StatusCode: code, Accepted: accepted[code]}
}

// encodeCP1251 returns windows-1251 bytes in a string; unmappable runes are
// dropped.
func encodeCP1251(s string) string {
	out, err := charmap.Windows1251.NewEncoder().String(s)
	if err != nil {
		var b strings.Builder
		for _, r := range s {
			if enc, encErr := charmap.Windows1251.NewEncoder().String(string(r)); encErr == nil {
				b.WriteString(enc)
			}
		}
		return b.String()
	}
	return out
}

// decodeBody honors the Content-Type charset and defaults to windows-1251.
func decodeBody(resp *http.Response, raw []byte) string {
	if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type")); err == nil {
		if name := params["charset"]; name != "" {
			if enc, err := htmlindex.Get(name); err == nil {
				if decoded, err := enc.NewDecoder().Bytes(raw); err == nil {
					return string(decoded)
				}
			}
		}
	}
	decoded, err := charmap.Windows1251.NewDecoder().Bytes(raw)
	if err != nil {
		return string(raw)
	}
	return string(decoded)
}

func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, 8<<20))
}

func firstBetween(s, open, close string) (string, bool) {
	start := strings.Index(s, open)
	if start < 0 {
		return "", false
	}
	rest := s[start+len(open):]
	end := strings.Index(rest, close)
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}

// betweenRows returns the pieces strictly between delimiters; the head and
// the tail are dropped.
func betweenRows(block, delimiter string) []string {
	parts := strings.Split(block, delimiter)
	if len(parts) < 3 {
		return nil
	}
	return parts[1 : len(parts)-1]
}

// asInt accepts both numbers and numeric strings.
func asInt(v any) (int, bool) {
	switch value := v.(type) {
	case json.Number:
		n, err := value.Int64()
		if err != nil {
			f, ferr := value.Float64()
			if ferr != nil {
				return 0, false
			}
			return int(f), true
		}
		return int(n), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, false
		}
		return n, true
	case float64:
		return int(value), true
	default:
		return 0, false
	}
}

func asString(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case json.Number:
		return value.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(value)
	}
}

// truthy treats null, false, 0, "" and empty containers as false.
func truthy(v any) bool {
	switch value := v.(type) {
	case nil:
		return false
	case bool:
		return value
	case string:
		return value != ""
	case json.Number:
		f, err := value.Float64()
		return err == nil && f != 0
	case float64:
		return value != 0
	case []any:
		return len(value) > 0
	case map[string]any:
		return len(value) > 0
	default:
		return true
	}
}

func intPtr(n int) *int { return &n }
