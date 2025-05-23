package session

import (
	"log/slog"

	"github.com/CGSG-2021-AE4/tomestobot/api"

	tele "gopkg.in/telebot.v4"
)

// Manages start/stop of sessions
type sessionManager struct {
	logger *slog.Logger
	bot    *tele.Bot
	group  *tele.Group

	users map[int64]*session
}

func NewManager(logger *slog.Logger, bot *tele.Bot, group *tele.Group) api.SessionManager {
	m := &sessionManager{
		logger: logger,
		bot:    bot,
		group:  group,

		users: map[int64]*session{},
	}

	return m
}

func (m *sessionManager) Exist(tgId int64) bool {
	_, exists := m.users[tgId]
	return exists
}

func (m *sessionManager) Get(tgId int64) api.Session {
	return m.users[tgId]
}

func (m *sessionManager) Start(tgId int64, u api.BxUser) api.Session {
	// If session exists return it
	if s := m.users[tgId]; s != nil {
		m.logger.Warn("trying to start session that already exists", "tgId", tgId)
		return s
	}
	s := createSession(m.logger.With("tgId", tgId), m.bot, m.group, u)
	m.users[tgId] = s
	return s
}

func (m *sessionManager) Stop(tgId int64) {
	if m.users[tgId] == nil {
		m.logger.Warn("trying to stop session that does not exist", "tgId", tgId)
		return
	}
	delete(m.users, tgId)
}
