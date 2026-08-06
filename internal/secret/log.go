package secret

import (
	"context"
	"log/slog"
)

// NewLogHandler wraps a handler so registered values never reach the log.
func NewLogHandler(inner slog.Handler) slog.Handler {
	return logHandler{inner: inner}
}

type logHandler struct {
	inner slog.Handler
}

func (h logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h logHandler) Handle(ctx context.Context, record slog.Record) error {
	clean := slog.NewRecord(record.Time, record.Level, Redact(record.Message), record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		clean.AddAttrs(redactAttr(attr))
		return true
	})
	return h.inner.Handle(ctx, clean)
}

func (h logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clean := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		clean[i] = redactAttr(attr)
	}
	return logHandler{inner: h.inner.WithAttrs(clean)}
}

func (h logHandler) WithGroup(name string) slog.Handler {
	return logHandler{inner: h.inner.WithGroup(name)}
}

func redactAttr(attr slog.Attr) slog.Attr {
	switch attr.Value.Kind() {
	case slog.KindString:
		return slog.String(attr.Key, Redact(attr.Value.String()))
	case slog.KindAny:
		if err, ok := attr.Value.Any().(error); ok {
			return slog.String(attr.Key, Redact(err.Error()))
		}
	case slog.KindGroup:
		group := attr.Value.Group()
		clean := make([]any, 0, len(group))
		for _, inner := range group {
			clean = append(clean, redactAttr(inner))
		}
		return slog.Group(attr.Key, clean...)
	}
	return attr
}
