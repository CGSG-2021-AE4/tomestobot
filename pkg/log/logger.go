package log

import (
	"context"
	"log/slog"
)

// Type that implements handler interface and logging itself and storing messages
type logsHandler struct {
	output LogsOutput
	level  slog.Level

	groups []string
	attrs  []slog.Attr
}

// Slog handler interface implementation

func (h *logsHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *logsHandler) Handle(ctx context.Context, record slog.Record) error {
	record.AddAttrs(h.attrs...)
	return h.output.Handle(ctx, h.groups, record)
}

func (h *logsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logsHandler{
		output: h.output,
		level:  h.level,

		groups: h.groups,
		attrs:  attrs,
	}
}

func (h *logsHandler) WithGroup(name string) slog.Handler {
	return &logsHandler{
		output: h.output,
		level:  h.level,

		groups: append(h.groups, name),
		attrs:  h.attrs,
	}
}

func NewLogger(output LogsOutput, level slog.Level) slog.Handler {
	return &logsHandler{
		output: output,
		level:  level,

		groups: []string{},
		attrs:  []slog.Attr{},
	}
}
