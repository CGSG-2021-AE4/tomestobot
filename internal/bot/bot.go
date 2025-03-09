package bot

import (
	"fmt"
	"log"
	"time"

	"tomestobot/api"
	bxwrapper "tomestobot/internal/bx"

	"github.com/go-playground/validator/v10"
	tele "gopkg.in/telebot.v4"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type BotDescriptor struct {
	TgBotToken string `validate:"required"`
	BxDomain   string `validate:"required,fqdn"` // Full Qualified Domain Name
	BxUserId   int    `validate:"required"`
	BxHook     string `validate:"required"`
}

type bot struct {
	bot *tele.Bot
	bx  api.BxWrapper

	whitelist map[int64]bool         // List of authorized users' IDs
	sessions  map[int64]*userSession // List of current sessions, sessions will be susspended after some idle time
}

type Bot interface {
	Start() error
}

func New(descr BotDescriptor) (Bot, error) {
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

	// Creating bx wrapper
	bx, err := bxwrapper.New(descr.BxDomain, descr.BxUserId, descr.BxHook)
	if err != nil {
		return nil, fmt.Errorf("bx wrapper creation: %w", err)
	}

	return &bot{
		bot: telebot,
		bx:  bx,

		whitelist: map[int64]bool{},         // Later will load from a file
		sessions:  map[int64]*userSession{}, // Map of active sessions
	}, nil
}

func (b *bot) authMiddle(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		log.Print("--- MSG FROM ---")
		log.Print(c.Sender().ID)
		log.Print(c.Message().Contact)

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

	log.Print("Start bot")
	b.bot.Start()
	log.Print("End bot")
	return nil
}
