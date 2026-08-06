package secret

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	const token = "123456789:AAHsecretvaluehere"
	Register(token)

	text := `Post "https://api.telegram.org/bot` + token + `/sendMessage": timeout`
	got := Redact(text)

	if strings.Contains(got, token) {
		t.Errorf("token survived redaction: %q", got)
	}
	if !strings.Contains(got, mask) {
		t.Errorf("mask missing: %q", got)
	}
	if !strings.Contains(got, "timeout") {
		t.Errorf("message lost: %q", got)
	}
}

func TestRegisterIgnoresShortValues(t *testing.T) {
	Register("abc")
	if got := Redact("abc def"); got != "abc def" {
		t.Errorf("short value was redacted: %q", got)
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
		t.Errorf("token in log output: %q", out)
	}
	if !strings.Contains(out, "chat_id=42") {
		t.Errorf("non-secret attrs lost: %q", out)
	}
	if strings.Count(out, mask) != 3 {
		t.Errorf("expected message, error and url redacted: %q", out)
	}
}

func TestLogHandlerRedactsGroupsAndWithAttrs(t *testing.T) {
	const token = "555555555:CCHgroupsecretvalue"
	Register(token)

	var buf bytes.Buffer
	logger := slog.New(NewLogHandler(slog.NewTextHandler(&buf, nil))).With("base", token)
	logger.Info("ok", slog.Group("g", "inner", token))

	if out := buf.String(); strings.Contains(out, token) {
		t.Errorf("token in log output: %q", out)
	}
}
