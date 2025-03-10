package session

import (
	"tomestobot/api"

	"github.com/charmbracelet/log"
	tele "gopkg.in/telebot.v4"
)

type dialogState int

const (
	dialogStarted = dialogState(iota)
	dialogDealsSelect
	dialogDealActionSelect
	dialogWriteComment
	dialogTasksSelect
)

type session struct {
	logger *log.Logger
	group  *tele.Group // Group for sessions' endpoints

	// tgID   int64
	bxUser api.BxUser
	state  dialogState
	// lastAction time.Time TODO
}

func (s *session) OnStart(c tele.Context) error {
	menu := &tele.ReplyMarkup{}

	listDealsBtn := menu.Data("List deals", "list_deals")
	s.group.Handle(&listDealsBtn, s.onListDeals)

	menu.Inline(
		menu.Row(listDealsBtn),
	)

	return c.Send("Select actions you want to do:", menu)
}

// Resets state to started
func (s *session) reset() {
	s.state = dialogStarted
}

// Handles list deals message
func (s *session) onListDeals(c tele.Context) error {
	s.logger.Debug("on list deals", "username", c.Sender().Username)
	return c.Send("Deals...")
}
