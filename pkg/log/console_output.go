package log

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const (
	// Colors for standart levels
	colorDebug = "\x1b[38;2;72;72;255m"
	colorInfo  = "\x1b[38;2;0;170;50m"
	colorWarn  = "\x1b[38;2;200;200;50m"
	colorError = "\x1b[38;2;200;20;50m"

	// Other colors
	colorReset = "\x1b[39m" // Resets foregroupd
)

func labelByLevel(l slog.Level) string {
	switch l {
	case slog.LevelDebug:
		return "DEBU"
	case slog.LevelInfo:
		return "INFO"
	case slog.LevelWarn:
		return "WARN"
	case slog.LevelError:
		return "ERROR"
	}
	return "unknown"
}
func coloredLabelByLevel(l slog.Level) string {
	switch l {
	case slog.LevelDebug:
		return colorDebug + "DEBU" + colorReset
	case slog.LevelInfo:
		return colorInfo + "INFO" + colorReset
	case slog.LevelWarn:
		return colorWarn + "WARN" + colorReset
	case slog.LevelError:
		return colorError + "ERROR" + colorReset
	}
	return "unknown"
}

func printAttr(a slog.Attr) bool {
	fmt.Print(a.Key + "=" + a.Value.String() + " ")
	return true
}

type consoleLogOutput struct {
}
type coloredConsoleLogOutput struct {
}

func (c consoleLogOutput) Handle(ctx context.Context, groups []string, record slog.Record) error {
	fmt.Print(record.Time.Format(time.DateTime) + " " + // Time
		labelByLevel(record.Level) + " " + // Level label
		strings.Join(groups, " ") + " : " + // Groups
		record.Message + " ") // Message

	record.Attrs(printAttr) // Attrs
	fmt.Print("\n")

	return nil
}

func (c coloredConsoleLogOutput) Handle(ctx context.Context, groups []string, record slog.Record) error {
	fmt.Print(record.Time.Format(time.DateTime) + " " + // Time
		coloredLabelByLevel(record.Level) + " " + // Level label
		strings.Join(groups, " ") + " : " + // Groups
		record.Message + " ") // Message

	record.Attrs(printAttr) // Attrs
	fmt.Print("\n")

	return nil
}

func NewConsoleLogOutput(enableColors bool) Output {
	if enableColors {
		return coloredConsoleLogOutput{}
	}
	return consoleLogOutput{}
}
