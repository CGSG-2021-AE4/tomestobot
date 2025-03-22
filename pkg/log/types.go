package log

import (
	"context"
	"log/slog"
)

type LogsOutput interface {
	Handle(ctx context.Context, groups []string, record slog.Record) error
}
