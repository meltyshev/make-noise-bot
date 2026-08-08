package secret

import (
	"bytes"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	const token = "123456789:AAHsecretvaluehere"
	Register(token)

	text := `Post "https://api.telegram.org/bot` + token + `/sendMessage": timeout`
	got := Redact(text)

	if strings.Contains(got, token) {
		t.Errorf("Redact(text with token) = %q, want the token replaced", got)
	}
	if !strings.Contains(got, mask) {
		t.Errorf("Redact(text with token) = %q, want it to contain %q", got, mask)
	}
	if !strings.Contains(got, "timeout") {
		t.Errorf("Redact(text with token) = %q, want the rest of the message kept", got)
	}
}

func TestRegisterIgnoresShortValues(t *testing.T) {
	Register("abc")
	if got := Redact("abc def"); got != "abc def" {
		t.Errorf("Redact(\"abc def\") = %q, want it unchanged: short values are not registered", got)
	}
}

func TestLogHandlerRedacts(t *testing.T) {
	const token = "987654321:BBHanothersecret"
	Register(token)

	var buf bytes.Buffer
	logger := slog.New(NewLogHandler(slog.NewTextHandler(&buf, nil)))

	logger.Error("telegram "+token+" failed",
		"error", errors.New("Get \"https://api.telegram.org/bot"+token+"/getMe\": timeout"),
		"url", "https://api.telegram.org/bot"+token+"/getMe",
		"chat_id", 42,
	)

	out := buf.String()
	if strings.Contains(out, token) {
		t.Errorf("log output = %q, want the token replaced", out)
	}
	if !strings.Contains(out, "chat_id=42") {
		t.Errorf("log output = %q, want chat_id=42 kept", out)
	}
	if got := strings.Count(out, mask); got != 3 {
		t.Errorf("masks in log output = %d, want 3: the message, the error and the url", got)
	}
}

func TestStripURLDropsTheURL(t *testing.T) {
	const token = "123456789:AAHsecretvaluehere"
	err := &url.Error{
		Op:  "Get",
		URL: "https://api.telegram.org/bot" + token + "/getMe",
		Err: errors.New("context deadline exceeded"),
	}

	got := StripURL(err).Error()
	if strings.Contains(got, token) {
		t.Errorf("StripURL(url.Error) = %q, want the token gone", got)
	}
	if !strings.Contains(got, "context deadline exceeded") {
		t.Errorf("StripURL(url.Error) = %q, want the cause kept", got)
	}
}

func TestStripURLKeepsOtherErrors(t *testing.T) {
	err := errors.New("Unauthorized")
	if got := StripURL(err); !errors.Is(got, err) {
		t.Errorf("StripURL(plain) = %v, want the error unchanged", got)
	}
}

// TestGamePasswordIsHidden pins both halves of the promise in
// docs/architecture.md: the engine login puts the password in a query string,
// so a transport error must lose the URL and a registered password must be
// masked wherever the text still carries it.
func TestGamePasswordIsHidden(t *testing.T) {
	const password = "correct-horse-battery"
	Register(password)

	err := &url.Error{
		Op:  "Get",
		URL: "https://classic.dzzzr.ru/e-burg/API/login.php?login=team&password=" + password,
		Err: errors.New("dial tcp: i/o timeout"),
	}

	if got := StripURL(err).Error(); strings.Contains(got, password) {
		t.Errorf("StripURL(login error) = %q, want the password gone", got)
	}
	if got := Redact(err.Error()); strings.Contains(got, password) {
		t.Errorf("Redact(login error) = %q, want the password masked", got)
	}
}

func TestLogHandlerRedactsGroupsAndWithAttrs(t *testing.T) {
	const token = "555555555:CCHgroupsecretvalue"
	Register(token)

	var buf bytes.Buffer
	logger := slog.New(NewLogHandler(slog.NewTextHandler(&buf, nil))).With("base", token)
	logger.Info("ok", slog.Group("g", "inner", token))

	if out := buf.String(); strings.Contains(out, token) {
		t.Errorf("log output = %q, want the token replaced in groups and WithAttrs too", out)
	}
}
