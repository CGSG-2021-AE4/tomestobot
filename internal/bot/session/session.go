package session

import (
	"fmt"
	"strconv"
	"tomestobot/api"
	"tomestobot/pkg/gobx/bxtypes"

	"github.com/charmbracelet/log"
	tele "gopkg.in/telebot.v4"
)

type session struct {
	logger *log.Logger
	group  *tele.Group // Group for sessions' endpoints

	// tgID   int64
	bxUser api.BxUser

	// Local state dynamic data
	flow  DialogFlow     // Controls right dialog order
	deals []bxtypes.Deal // Current user deals
	deal  bxtypes.Deal   // Deal the user is working with
	tasks []bxtypes.Task // Tasks of current deal
	// lastAction time.Time TODO
}

// Commands handlers
// Are named after commands or actions they execute

// Main message - may be consider as help
func (s *session) OnStart(c tele.Context) error {
	if err := s.flow.Set(DialogStarted); err != nil {
		return s.sendError(c, err)
	}
	defer s.flow.Done()

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
	if err := s.flow.Set(DialogDealsList); err != nil {
		return s.sendError(c, err)
	}
	defer s.flow.Done()

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
		btn := menu.Data(d.Title, "selectDeal", strconv.Itoa(i)) // Attach index of deal in deals array
		s.group.Handle(&btn, s.onDealActions)
		btns = append(btns, menu.Row(btn))
	}
	menu.Inline(btns...)

	return c.Send("Select the deal you want to work with", menu)
}

// Shows actions with select deal
func (s *session) onDealActions(c tele.Context) error {
	if err := s.flow.Set(DialogDealActions); err != nil {
		return s.sendError(c, err)
	}
	defer s.flow.Done()

	// Get deal of the button
	s.logger.Debug(c.Data())
	if c.Data() != "" || s.deal == bxtypes.NilDeal { // Means we came here not from deal list but from add comment
		i, err := strconv.Atoi(c.Data()) // Deal index in deals array
		if err != nil {
			return s.sendError(c, fmt.Errorf("parsing deal index: %w", err))
		}
		if i >= len(s.deals) {
			return s.sendError(c, fmt.Errorf("invlide deal index in deals array"))
		}
		s.deal = s.deals[i]
	}
	s.logger.Debug(s.deal)

	// Create buttons
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
	if err := s.flow.Set(DialogWriteComment); err != nil {
		return s.sendError(c, err)
	}
	defer s.flow.Done()

	s.group.Handle(tele.OnText, s.onAddComment)

	return c.Send("Please, write a message:")
}

// Add written comment to deal
func (s *session) onAddComment(c tele.Context) error {
	if err := s.flow.Set(DialogAddComment); err != nil { // Hooks only uncomplete error
		return s.sendError(c, err)
	}
	if s.flow.Get() != DialogAddComment || s.flow.IsDone() { // Means it is just text - not for comment
		s.logger.Warn("got raw text outside comment", "username", c.Sender().Username)
		return c.Send("DEBUG  WARNING:\nraw text messages work only while adding comment\n\nFor menu type <code>/start</code>") // DEBUG
	}
	s.flow.Done()

	s.logger.Debug("onAddComment", c.Text())
	commentId, err := s.bxUser.AddCommentToDeal(s.deal.Id, c.Text())
	if err != nil {
		return s.sendError(c, fmt.Errorf("bx add comment to deal: %w", err))
	}
	s.logger.Debug("Added comment", "id", commentId)

	if err := c.Send("comment added"); err != nil {
		return err
	}
	s.flow.Done()
	return s.onDealActions(c)
}

// Lists deal tasks
func (s *session) onListTasks(c tele.Context) error {
	if err := s.flow.Set(DialogTasksList); err != nil {
		return s.sendError(c, err)
	}
	defer s.flow.Done()

	tasks, err := s.bxUser.ListDealTasks(s.deal.Id)
	if err != nil {
		return s.sendError(c, fmt.Errorf("bx list deal tasks: %w", err))
	}
	s.tasks = tasks

	btns := []tele.Row{}
	menu := &tele.ReplyMarkup{}
	for i, t := range tasks {
		btn := menu.Data(t.Title, "selectTask", strconv.Itoa(i)) // Attach index of task in tasks array
		s.group.Handle(&btn, s.onCompleteTask)
		btns = append(btns, menu.Row(btn))
	}
	menu.Inline(btns...)

	return c.Send("Select the task you want to complete", menu)
}

// Completes selected task
func (s *session) onCompleteTask(c tele.Context) error {
	if err := s.flow.Set(DialogTaskComplete); err != nil {
		return s.sendError(c, err)
	}
	defer s.flow.Done()

	s.logger.Debug("Complete task", "i", c.Data())

	// Validate task index and task itself
	index, err := strconv.Atoi(c.Data())
	if err != nil {
		return s.sendError(c, fmt.Errorf("parsing task index: %w", err))
	}
	if index > len(s.tasks) {
		return s.sendError(c, fmt.Errorf("invalid task index"))
	}
	task := s.tasks[index]
	if err := s.bxUser.CompleteTask(task.Id); err != nil {
		return s.sendError(c, fmt.Errorf("bx complete task: %w", err))
	}

	return c.Send(fmt.Sprintf("Task <u>%s</u> was successfully completed.", task.Title))
}

// Supporting functions

// Resets state to started
func (s *session) reset() {
	s.flow.Done() // To ensure there is no error
	s.flow.Set(DialogStarted)
	s.flow.Done()
}

func (s *session) sendError(c tele.Context, err error) error {
	s.logger.Warn("ERROR", "username", c.Sender().Username, "err", err.Error())
	return c.Send(fmt.Sprintf("ERROR:\n<code>%s</code>\n\nTry restart bot", err.Error()))
}
