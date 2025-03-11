package session

import (
	"fmt"
	"strconv"
	"tomestobot/api"
	"tomestobot/pkg/gobx/bxtypes"

	"github.com/charmbracelet/log"
	tele "gopkg.in/telebot.v4"
)

type dialogState int

const (
	dialogStarted = dialogState(iota)
	dialogDealsList
	dialogDealActions
	dialogWriteComment
	dialogAddComment
	dialogTasksList
)

type session struct {
	logger *log.Logger
	group  *tele.Group // Group for sessions' endpoints

	// tgID   int64
	bxUser api.BxUser

	// Local state dynamic data
	state dialogState
	deals []bxtypes.Deal // Current user deals
	deal  bxtypes.Deal   // Deal the user is working with
	// lastAction time.Time TODO
}

// Commands handlers
// Are named after commands or actions they execute

func (s *session) OnStart(c tele.Context) error {
	s.state = dialogStarted // Update state - do not check because it does not matter
	menu := &tele.ReplyMarkup{}

	listDealsBtn := menu.Data("List deals", "list_deals")
	s.group.Handle(&listDealsBtn, s.onListDeals)

	menu.Inline(
		menu.Row(listDealsBtn),
	)

	return c.Send("Select actions you want to do:", menu)
}

// Handles list deals message
func (s *session) onListDeals(c tele.Context) error {
	// Check possible entry states
	if s.state != dialogStarted && s.state != dialogDealActions {
		return s.sendError(c, fmt.Errorf("invalid entry state: %d", s.state))
	}
	s.state = dialogDealsList // Update state

	s.logger.Debug("on list deals", "username", c.Sender().Username)

	// Get deals
	deals, err := s.bxUser.ListDeals()
	if err != nil {
		s.logger.Warn("list deals error", "username", c.Sender().Username, "err", err.Error())
		return c.Send(fmt.Sprintf("Got error: %s\n Try to restart bot", err.Error()))
	}
	s.deals = deals // Save deals
	s.logger.Debug(deals)

	// Setup buttons
	btns := []tele.Row{}
	menu := &tele.ReplyMarkup{}
	for i, d := range deals {
		btn := menu.Data(d.Title, "selectDeal", strconv.Itoa(i)) // Attach id of deal in deals array
		s.group.Handle(&btn, s.onDealActions)
		btns = append(btns, menu.Row(btn))
	}
	menu.Inline(btns...)

	return c.Send("Select the deal you want to work with", menu)
}

// Shows actions with select deal
func (s *session) onDealActions(c tele.Context) error {
	// Check possible entry states
	if s.state != dialogDealsList && s.state != dialogAddComment {
		return s.sendError(c, fmt.Errorf("invalid entry state: %d", s.state))
	}
	s.state = dialogDealActions // Update state

	// Get deal of the button
	s.logger.Debug(c.Data())
	i, err := strconv.Atoi(c.Data()) // Deal index in deals array
	if err != nil {
		return s.sendError(c, fmt.Errorf("parsing deal index: %w", err))
	}
	if i >= len(s.deals) {
		return s.sendError(c, fmt.Errorf("invlide deal index in deals array"))
	}
	s.deal = s.deals[i]
	s.logger.Debug(s.deal)

	menu := &tele.ReplyMarkup{}
	addCommentBtn := menu.Data("Add comment", "addComment")
	s.group.Handle(&addCommentBtn, s.onWriteComment)
	listTasksBtn := menu.Data("List tasks", "listTasks")
	s.group.Handle(&listTasksBtn, s.onListTasks)

	menu.Inline(
		menu.Row(addCommentBtn),
		menu.Row(listTasksBtn),
	)

	return c.Send(fmt.Sprintf("Deal: %s\nStatus: %s\nSelect the action:", s.deal.Title, s.deal.StageId), menu)
}

// Asks to write a coomment
func (s *session) onWriteComment(c tele.Context) error {
	// Check possible entry states
	if s.state != dialogDealActions {
		return s.sendError(c, fmt.Errorf("invalid entry state: %d", s.state))
	}
	s.state = dialogAddComment // Update state

	return c.Send("NOT IMPLEMENTED YET")
}

// Add written comment to deal
func (s *session) onAddComment(c tele.Context) error {
	// Check possible entry states
	if s.state != dialogDealActions {
		return s.sendError(c, fmt.Errorf("invalid entry state: %d", s.state))
	}
	s.state = dialogAddComment // Update state

	return c.Send("NOT IMPLEMENTED YET")
}

// Lists deal tasks
func (s *session) onListTasks(c tele.Context) error {
	// Check possible entry states
	if s.state != dialogDealsList {
		return s.sendError(c, fmt.Errorf("invalid entry state: %d", s.state))
	}
	s.state = dialogDealActions // Update state

	return c.Send("NOT IMPLEMENTED YET")
}

// Supporting functions

// Resets state to started
func (s *session) reset() {
	s.state = dialogStarted
}

func (s *session) sendError(c tele.Context, err error) error {
	s.logger.Warn("Session error", "username", c.Sender().Username, "err", err.Error())
	return c.Send(fmt.Sprintf("ERROR: %s\nTry restart bot", err.Error()))
}
