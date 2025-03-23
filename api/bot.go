package api

import (
	"github.com/CGSG-2021-AE4/tomestobot/pkg/log"
	tele "gopkg.in/telebot.v4"
)

type Bot interface {
	Start() error

	GetLogsOutput() log.Output // For tg logging
}

type Session interface {
	OnStart(c tele.Context) error
}

type SessionManager interface {
	// Get/Exist are separate functions because in most cases I need to know only one of these values
	Get(tgId int64) Session
	Exist(tgId int64) bool

	Start(tgId int64, u BxUser) Session
	Stop(tgId int64)
}
