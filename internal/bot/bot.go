package bot

import (
	"fmt"
	"os"
	"strings"
	"time"

	"tomestobot/api"
	"tomestobot/internal/bot/session"
	"tomestobot/pkg/gobx/bxtypes"

	"github.com/charmbracelet/log"
	"github.com/go-playground/validator/v10"
	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type BotDescriptor struct {
	TgBotToken string        `validate:"required"`
	Bx         api.BxWrapper `validate:"required"`
}

type bot struct {
	logger *log.Logger

	bot       *tele.Bot     // Telegram bot API wrapper
	mainGroup *tele.Group   // Group for main handlers - is neede because I do not need to apply session middle for OnContact endpoint
	bx        api.BxWrapper // Bitrix wrapper

	idStore  api.UsersIdStore   // Store of familiar users' IDs, so they do not have to share their contact every time
	sessions api.SessionManager // Manages sessions
}

func New(logger *log.Logger, descr BotDescriptor) (api.Bot, error) {
	// Validate descriptor
	if err := validate.Struct(descr); err != nil {
		return nil, fmt.Errorf("bot descriptor validation: %w", err)
	}

	// Creating telebot
	pref := tele.Settings{
		Token:     descr.TgBotToken,
		Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tele.ModeHTML,
	}
	telebot, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("telebot creation: %w", err)
	}

	// Setup session group
	mainGroup := telebot.Group()

	b := &bot{
		logger: logger,

		bot:       telebot,
		mainGroup: mainGroup,
		bx:        descr.Bx,

		idStore:  NewJsonUsersIdStore(logger, os.Getenv("ID_STORE_FILE")),
		sessions: session.NewManager(logger, mainGroup),
	}

	if err := b.setupEndpoints(); err != nil {
		return nil, fmt.Errorf("bot setup endpoints: %w", err)
	}
	return b, nil
}

func (b *bot) Start() error {
	b.logger.Debug("bot started")
	b.bot.Start()
	b.logger.Debug("bot ended")
	return nil
}

func (b *bot) setupEndpoints() error {
	// Setup middle
	b.mainGroup.Use(b.sessionMiddle) // For authorization
	b.bot.Use(middleware.AutoRespond())
	b.bot.Use(middleware.Recover(func(err error, c tele.Context) {
		str := fmt.Sprintf("ERROR:\n<code>panic: %s</code>\n\nПерезапутите бот прописав <code>/start</code>}", err.Error())
		b.logger.Warn(str, "username", c.Sender().Username)
		c.Send(str)
	}))

	// Contact for auth
	b.bot.Handle(tele.OnContact, b.onContact) // The method is the only one not in auth group!!!

	b.mainGroup.Handle("/start", func(c tele.Context) error {
		// If user reached this endpoint - session exists
		// Reset session
		return b.sessions.Get(c.Sender().ID).OnStart(c)
	})

	b.mainGroup.Handle("/stop", func(c tele.Context) error { // For debug purposes - ends user's session
		if b.sessions.Exist(c.Sender().ID) {
			b.sessions.Stop(c.Sender().ID)
			return c.Send("Сессия остановлена")
		}
		return c.Send("Не найдено активных сессий")
	})
	return nil
}

// Handles the creation of sessions and authorization of users
func (b *bot) sessionMiddle(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		// Session does not exist - do auth stuff
		if !b.sessions.Exist(c.Sender().ID) {
			b.logger.Debug("auth user", "id", c.Sender().ID)
			// Check if chat is suitable for conversation
			if c.Sender().IsBot {
				return c.Send("Сообщения от ботов не разрешены")
			}

			// Check by id else request contact info
			know, err := b.tryAuthById(c)
			if err != nil {
				return fmt.Errorf("try auth by id: %w", err)
			}
			if !know { // We do not know the id - need to auth by phone
				return b.reqContact(c)
			}
			// We know user and there was know errors with auth => we authed!
			// Continue with request
		}
		return next(c)
	}
}

// Requests contact from user
func (b *bot) reqContact(c tele.Context) error {
	// Setup reply markup
	r := &tele.ReplyMarkup{ResizeKeyboard: true}
	r.Reply(r.Row(r.Contact("Предоставить номер")))

	return c.Send(`Для авторизыации предоставьте номер телефона.(кнопка "Предоставить номер")`, r)
}

// OnContact endpoint callback
func (b *bot) onContact(c tele.Context) error {
	b.logger.Debug("on contact")
	if !b.sessions.Exist(c.Sender().ID) {
		// Session does not exist so we auth

		// Clear messages anyway - differs on desk and mobile versions (ReplyTo would be nil on mobile)
		// b.bot.Delete(c.Message().ReplyTo)
		// b.bot.Delete(c.Message())

		b.logger.Debug(c.Message().Contact)
		// Try to auth
		if err := b.tryAuthByPhone(c); err != nil {
			// Other error
			b.logger.Debug("send authe rror")
			footer, str := api.ErrorText(err)
			b.logger.Warn(str, "username", c.Sender().Username)
			if footer {
				str += "\n\nДля перезапуска отправьте команду <code>/start</code>"
			}
			return c.Send(str)
		}

		if err := c.Send("Авторизация прошла успешно."); err != nil {
			return fmt.Errorf("success authed msg send: %w", err)
		}
		return b.bot.Trigger("/start", c)
	}

	b.logger.Warn("got on contact message but user is authorised")
	return nil
}

// Checks if user is familiar(we know his vx id) and if session does not exist it creates it
// Returns true if we know the user, false if not
func (b *bot) tryAuthById(c tele.Context) (bool, error) {
	// Assume session does not exist
	tgId := c.Sender().ID
	bxId, wok := b.idStore.Get(tgId)

	if wok { // id exists in the list of familiar users and session does not exist
		u, err := b.bx.AuthUserById(bxtypes.Id(bxId))
		if err != nil {
			return true, err
		}
		// Auth is successful
		b.logger.Debug("user authed by id", "username", c.Sender().Username)
		// Create session
		b.sessions.Start(tgId, u)

		return true, nil
	}
	return false, nil
}

// Checks if session exists and if not - auth user by phone, add it to the list of familiar users and create session
func (b *bot) tryAuthByPhone(c tele.Context) error {
	// Assume session does not exist
	tgId := c.Sender().ID

	// Validate message
	if c.Message().Contact == nil {
		return api.ErrorNoContactInMsg
	}
	// Do auth
	b.logger.Debug("bx log by phone")

	phoneNumber := fixPhoneNumber(c.Message().Contact.PhoneNumber)
	b.logger.Debug(phoneNumber)
	u, err := b.bx.AuthUserByPhone(phoneNumber)
	if err != nil {
		return err
	}
	b.logger.Debug("ok")

	// Auth is successful
	b.logger.Debug("user authed by phone", "username", c.Sender().Username)
	// Save user
	b.idStore.Set(tgId, int64(u.Get().Id))
	if err := b.idStore.Save(); err != nil {
		b.logger.Warn(err)
	}
	// Create session
	b.sessions.Start(tgId, u)

	return nil
}

// Fixes phone number because telegram provide it in different style
// Even on mobile it differes
func fixPhoneNumber(in string) string {
	str := in

	// Remove all symbols but numbers
	str = strings.ReplaceAll(str, "(", "")
	str = strings.ReplaceAll(str, ")", "")
	str = strings.ReplaceAll(str, "-", "")
	str = strings.ReplaceAll(str, "+", "")
	str = strings.ReplaceAll(str, " ", "")

	// Add plus if it is not 8
	if len(str) < 10 { // Definetely is not a phone number
		return str
	}
	if str[0] != '8' {
		return "+" + str
	}
	return str
}
