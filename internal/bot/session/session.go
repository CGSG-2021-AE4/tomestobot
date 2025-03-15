package session

import (
	"fmt"
	"net/http"
	"regexp"
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

	listDealsBtn := menu.Data("Показать открытые сделки", "list_deals")
	s.group.Handle(&listDealsBtn, s.onListDeals)

	menu.Inline(
		menu.Row(listDealsBtn),
	)

	return c.Send(fmt.Sprintf("Здравствуйте, %s %s.\nВыберете действие.", s.bxUser.Get().Name, s.bxUser.Get().LastName), menu)
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
		return s.sendError(c, err)
	}
	s.deals = deals // Save deals
	s.logger.Debug(deals)

	// Case when no deals found
	if len(deals) == 0 {
		if err := c.Send("Не найдено открытых сделок."); err != nil {
			return s.sendError(c, err)
		}
		s.flow.Done()
		return s.OnStart(c)
	}

	// Setup buttons
	btns := []tele.Row{}
	menu := &tele.ReplyMarkup{}
	for i, d := range deals {
		btn := menu.Data(d.Title, "selectDeal", strconv.Itoa(i)) // Attach index of deal in deals array
		s.group.Handle(&btn, s.onDealActions)
		btns = append(btns, menu.Row(btn))
	}
	menu.Inline(btns...)

	return c.Send("Выберете сделку:", menu)
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
	addCommentBtn := menu.Data("Добавить коментарий", "addComment")
	s.group.Handle(&addCommentBtn, s.onWriteComment)
	listTasksBtn := menu.Data("Показать открытые задачи", "listTasks")
	s.group.Handle(&listTasksBtn, s.onListTasks)
	backBtn := menu.Data("Назад", "back")
	s.group.Handle(&backBtn, s.onListDeals)

	menu.Inline(
		menu.Row(addCommentBtn),
		menu.Row(listTasksBtn),
		menu.Row(backBtn),
	)

	return c.Send(fmt.Sprintf("Сделка: <i>%s</i>\nСтатус: <i>%s</i>\n\nВыберете действие:", s.deal.Title, s.deal.StageId), menu)
}

// Asks to write a coomment
func (s *session) onWriteComment(c tele.Context) error {
	if err := s.flow.Set(DialogWriteComment); err != nil {
		return s.sendError(c, err)
	}
	defer s.flow.Done()

	s.group.Handle(tele.OnText, s.onAddComment)

	return c.Send("Напишите коментарий:")
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
	defer s.flow.Done()

	s.logger.Debug("onAddComment", c.Text())
	commentId, err := s.bxUser.AddCommentToDeal(s.deal.Id, c.Text())
	if err != nil {
		return s.sendError(c, err)
	}
	s.logger.Debug("Added comment", "id", commentId)

	// Report status
	if err := c.Send(fmt.Sprintf(`Коментарий "%s" Добавлен к сделке <i>%s</i>`, c.Text(), s.deal.Title)); err != nil {
		return err
	}

	// Create buttons
	menu := &tele.ReplyMarkup{}
	goToStartBtn := menu.Data("Нет", "goToStart")
	s.group.Handle(&goToStartBtn, s.OnStart)
	listTasksBtn := menu.Data("Да", "listTasks")
	s.group.Handle(&listTasksBtn, s.onListTasks)

	menu.Inline(
		menu.Row(goToStartBtn),
		menu.Row(listTasksBtn),
	)
	return c.Send("Нужно ли закрыть задачу по этой сделке?", menu)
}

// Lists deal tasks
func (s *session) onListTasks(c tele.Context) error {
	if err := s.flow.Set(DialogTasksList); err != nil {
		return s.sendError(c, err)
	}
	defer s.flow.Done()

	tasks, err := s.bxUser.ListDealTasks(s.deal.Id)
	if err != nil {
		return s.sendError(c, err)
	}
	s.tasks = tasks

	// Case when no deals found
	if len(tasks) == 0 {
		if err := c.Send("Нет открытых задач."); err != nil {
			return s.sendError(c, err)
		}
		s.flow.Done()
		return s.onDealActions(c)
	}

	btns := []tele.Row{}
	menu := &tele.ReplyMarkup{}
	r, _ := regexp.Compile("по сделке.*")
	for i, t := range tasks {
		btn := menu.Data(r.ReplaceAllLiteralString(t.Title, ""), "selectTask", strconv.Itoa(i)) // Attach index of task in tasks array
		s.group.Handle(&btn, s.onCompleteTask)
		btns = append(btns, menu.Row(btn))
	}
	menu.Inline(btns...)

	return c.Send("Выберете задачу для завершения:", menu)
}

// Completes selected task
func (s *session) onCompleteTask(c tele.Context) error {
	if err := s.flow.Set(DialogTaskComplete); err != nil {
		return s.sendError(c, err)
	}

	// s.logger.Debug("complete task", "i", c.Data())

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
		return s.sendError(c, err)
	}

	if err := c.Send(fmt.Sprintf("Задача <i>%s</i> успешно завершена.", task.Title)); err != nil {
		return s.sendError(c, err)
	}

	s.flow.Done()
	return s.onDealActions(c)
}

// Supporting functions

// Resets state to started
func (s *session) reset() {
	s.flow.Done() // To ensure there is no error
	s.flow.Set(DialogStarted)
	s.flow.Done()
}

// Function that analise !my !internal errors and log/ sends report
func (s *session) sendError(c tele.Context, err error) error {
	addFooter, str := errorText(err)

	s.logger.Warn("str", "username", c.Sender().Username)

	if addFooter {
		str += "\n\nДля перезапуска отправьте команду <code>/start</code>"
	}

	return c.Send(str)
}

// Styles error string
// Returns:
//   - do add help footer
//   - styled error
func errorText(err error) (bool, string) {
	if err, ok := err.(bxtypes.ErrorResty); ok { // Resty
		return true, fmt.Sprintf("ERROR:\n<code>resty level: %s</code>", err.Error())
	}
	if err, ok := err.(bxtypes.ErrorStatusCode); ok { // HTTP status code
		return true, fmt.Sprintf("ERROR:\n<code>http status: %s</code>", http.StatusText(int(err)))
	}
	if err, ok := err.(bxtypes.ErrorResponse); ok { // HTTP status code
		return true, fmt.Sprintf("ERROR:\n<code>with response: %s</code>", api.ErrorResponseText(err))
	}
	if err, ok := err.(api.ErrorInternal); ok { // HTTP status code
		switch err { // Special errors
		case api.ErrorUserNotFound:
			return false, "Пользователь не найден."
		case api.ErrorSeveralUsersFound:
			return false, "Ошибка: найдено несколько пользователей"
		}
		return true, fmt.Sprintf("ERROR:\n<code>internal level: %s</code>", api.ErrorInternalText(err))
	}

	return true, fmt.Sprintf("ERROR:\n<code>unknown level: %s</code>", err.Error())
}
