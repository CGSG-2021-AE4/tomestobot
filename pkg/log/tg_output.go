package log

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tele "gopkg.in/telebot.v4"
)

// Allow you to report errors to tg users

type TgOutput struct {
	bot        *tele.Bot
	recipients []tele.Recipient
}

func (out *TgOutput) Add(r tele.Recipient) {
	out.recipients = append(out.recipients, r)
}

// Output interface implementation
func (out *TgOutput) Handle(ctx context.Context, groups []string, record slog.Record) error {
	for _, r := range out.recipients {
		if _, err := out.bot.Send(r, formatTgError(groups, record)); err != nil {
			return err
		}
	}
	return nil
}

func formatTgError(groups []string, record slog.Record) string {
	str := ""
	f := func(a slog.Attr) bool {
		str += fmt.Sprintf("%s=<code>%s</code>\n", a.Key, a.Value)
		return true
	}
	// Collect attrs
	record.Attrs(f) // Because its the only way...

	return fmt.Sprintf("%s\n<code>%s : %s</code>\n%s", labelByLevel(record.Level), strings.Join(groups, " "), record.Message, str)
}

func NewTgOutput(bot *tele.Bot) *TgOutput {
	return &TgOutput{
		bot:        bot,
		recipients: []tele.Recipient{},
	}
}
