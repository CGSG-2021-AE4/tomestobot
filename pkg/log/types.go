package log

import (
	"context"
	"log/slog"
)

type Output interface {
	Handle(ctx context.Context, groups []string, record slog.Record) error
}
