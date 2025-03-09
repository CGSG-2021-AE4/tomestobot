package bot

import (
	"fmt"
	"time"

	"tomestobot/api"

	"github.com/charmbracelet/log"
	"github.com/go-playground/validator/v10"
	tele "gopkg.in/telebot.v4"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type BotDescriptor struct {
	TgBotToken string        `validate:"required"`
	Bx         api.BxWrapper `validate:"required"`
}

type bot struct {
	logger *log.Logger

	bot *tele.Bot     // Telegram bot API wrapper
	bx  api.BxWrapper // Bitrix wrapper

	whitelist map[int64]bool         // List of authorized users' IDs
	sessions  map[int64]*userSession // List of current sessions, sessions will be susspended after some idle time
}

type Bot interface {
	Start() error
}

func New(logger *log.Logger, descr BotDescriptor) (Bot, error) {
	// Validate descriptor
	if err := validate.Struct(descr); err != nil {
		return nil, fmt.Errorf("bot descriptor validation: %w", err)
	}

	// Creating telebot
	pref := tele.Settings{
		Token:  descr.TgBotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	telebot, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("telebot creation: %w", err)
	}

	return &bot{
		logger: logger,

		bot: telebot,
		bx:  descr.Bx,

		whitelist: map[int64]bool{},         // Later will load from a file
		sessions:  map[int64]*userSession{}, // Map of active sessions
	}, nil
}

func (b *bot) authMiddle(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {

		b.logger.Debug("msg", "id", c.Sender().ID)

		// Check that this is private chat
		// if !c.Chat().Private {
		// 	return c.Send("Public chats are not allowed")
		// }
		// Check that this is not a bot
		if c.Sender().IsBot {
			return c.Send("Bots are not allowed")
		}

		// Check if user id is in whitelist if only it is not a Contact message
		// TODO change to groups middle
		if c.Message().Contact == nil && !b.whitelist[c.Sender().ID] {
			return b.reqContact(c) // Request contact for auth
		}
		return next(c)
	}
}

// Requests contact from user
func (b *bot) reqContact(c tele.Context) error {
	// Setup reply markup
	r := &tele.ReplyMarkup{ResizeKeyboard: true}
	r.Reply(r.Row(r.Contact("Share contact")))

	return c.Send("Share your contant for auth", r)
}

func (b *bot) setupAuth() {
	b.bot.Use(b.authMiddle)

	b.bot.Handle(tele.OnContact, func(c tele.Context) error {
		// Validate the message
		if !c.Message().IsReply() {
			return c.Send("ERROR: contact message is not replied")
		}

		b.logger.Debug("contact msg", "id", c.Sender().ID, "name", c.Sender().FirstName+" "+c.Sender().LastName, "phone", c.Message().Contact.PhoneNumber)

		// Do auth
		b.whitelist[c.Sender().ID] = true

		// Clear messages
		b.bot.Delete(c.Message().ReplyTo)
		b.bot.Delete(c.Message())
		return c.Send("Thanks")
	})
}

func (b *bot) Start() error {
	b.setupAuth()

	b.bot.Handle("/start", func(c tele.Context) error {
		return c.Send("Hi")
	})

	b.logger.Debug("bot started")
	b.bot.Start()
	b.logger.Debug("bot ended")
	return nil
}
