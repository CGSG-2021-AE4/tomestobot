package session

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/CGSG-2021-AE4/tomestobot/api"
	"github.com/CGSG-2021-AE4/tomestobot/pkg/gobx/bxtypes"

	tele "gopkg.in/telebot.v4"
)

type session struct {
	logger *slog.Logger
	bot    *tele.Bot   // Because the only way to send a message and get beck it's sign is through this var
	group  *tele.Group // Group for sessions' endpoints

	// tgID   int64
	bxUser api.BxUser

	// Dynamic data

	// Comment
	waitingForComment bool // Writing comment is toggled and now I waiting for text message that will be treated like a comment
	writeCommentMsg   tele.Editable

	// Payload is for any data that cannot be passed directly
	// Now it is used in comment
	payload string // String because telebot uses string
}

// Supplement structures
// For inline menu creation functions
type inlineBtnDescr struct {
	text    string
	unique  string
	payload any // Will be marshalled to json later
}

// The same descriptor but with handler function
type inlineBtnWithHandlerDescr struct {
	text    string
	unique  string
	payload any // Will be marshalled to json later
	handler tele.HandlerFunc
}

type taskBtnPayload struct {
	Deal bxtypes.Deal `json:"deal"`
	Task bxtypes.Task `json:"task"`
}

// Create session function
func createSession(logger *slog.Logger, bot *tele.Bot, group *tele.Group, user api.BxUser) *session {
	s := &session{
		logger: logger,
		bot:    bot,
		group:  group,
		bxUser: user,

		// Dynamic data
		waitingForComment: false,
		writeCommentMsg:   nil,
		payload:           "",
	}
	// Setup some handlers
	s.group.Handle(tele.OnText, s.onAddComment)

	return s
}

// Commands handlers
// Are named after commands or actions they execute

// Main message - may be consider as help
func (s *session) OnStart(c tele.Context) error {
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
	s.logger.Debug("on list deals")

	// Get deals
	deals, err := s.bxUser.ListDeals()
	if err != nil {
		return s.sendError(c, err)
	}
	s.logger.Debug(fmt.Sprint(deals))

	// Case when no deals found
	if len(deals) == 0 {
		if err = c.Send("Не найдено открытых сделок."); err != nil {
			return s.sendError(c, err)
		}
		return s.OnStart(c)
	}

	// Prepare buttons descriptors
	btnDescrs := []inlineBtnDescr{}
	for _, d := range deals {
		btnDescrs = append(btnDescrs, inlineBtnDescr{
			text:    d.Title,
			unique:  "dealBtn" + d.Id.String(),
			payload: d,
		})
	}
	menu, err := creatInlineMenu(s.group, s.onDealActions, btnDescrs)
	if err != nil {
		s.sendError(c, err)
	}
	return c.Send("Выберете сделку:", menu)
}

// Shows actions with select deal
func (s *session) onDealActions(c tele.Context) error {
	// Get deal of the button
	s.logger.Debug(c.Data())

	// Check for payload
	if c.Data() == "" {
		s.logger.Debug("no btn payload")
		return s.sendError(c, api.ErrorInvalidBtnPayload)
	}
	// Parse payload
	deal := bxtypes.Deal{}
	if err := json.Unmarshal([]byte(c.Data()), &deal); err != nil {
		s.logger.Debug(err.Error())
		return s.sendError(c, api.ErrorInvalidBtnPayload)
	}

	// Create buttons
	menu, err := creatInlineMenuWithHandler(s.group, []inlineBtnWithHandlerDescr{
		{
			text:    "Добавить коментарий",
			unique:  "addComment" + deal.Id.String(),
			handler: s.onWriteComment,
			payload: c.Data(),
		},
		{
			text:    "Показать открытые задачи",
			unique:  "listTasks" + deal.Id.String(),
			handler: s.onListTasks,
			payload: c.Data(),
		},
		{
			text:    "Назад",
			unique:  "dealBackBtn" + deal.Id.String(),
			handler: s.onDealActions,
		},
	})
	if err != nil {
		s.sendError(c, err)
	}
	return c.Send(fmt.Sprintf("Сделка: <i>%s</i>\nСтатус: <i>%s</i>\n\nВыберете действие:", deal.Title, bxtypes.DealStageText(deal.StageId)), menu)
}

// Asks to write a coomment
func (s *session) onWriteComment(c tele.Context) error {
	// To be sure it is alright at this step
	if c.Data() == "" {
		s.logger.Debug("No payload")
		return s.sendError(c, api.ErrorInvalidBtnPayload)
	}
	// Send message
	msg, err := s.bot.Send(c.Sender(), "Напишите коментарий:")
	if err != nil {
		return err
	}
	// Save payload
	s.writeCommentMsg = msg
	s.payload = c.Data()
	s.waitingForComment = true
	return nil
}

// Add written comment to deal
// Handles all bare text messages
// Culls them if we do not wait comment
func (s *session) onAddComment(c tele.Context) error {
	// Check if it is comment or just text msg
	if !s.waitingForComment {
		s.logger.Debug("got text when I don't expect it")
		return c.Send("Text msgs are not allowed")
	}
	s.waitingForComment = false // Remove flag before any error
	s.logger.Debug("onAddComment", "msg", c.Text())

	// Check for payload
	if s.payload == "" {
		s.logger.Debug("no btn payload")
		return s.sendError(c, api.ErrorInvalidBtnPayload)
	}
	// Parse payload
	deal := bxtypes.Deal{}
	if err := json.Unmarshal([]byte(s.payload), &deal); err != nil {
		s.logger.Debug(err.Error())
		return s.sendError(c, api.ErrorInvalidBtnPayload)
	}

	// Add comment
	commentId, err := s.bxUser.AddCommentToDeal(deal.Id, c.Text())
	if err != nil {
		return s.sendError(c, err)
	}
	s.logger.Debug("Added comment", "id", commentId)

	// Report status
	if err = c.Send(fmt.Sprintf(`Коментарий "%s" Добавлен к сделке <i>%s</i>`, c.Text(), deal.Title)); err != nil {
		return err
	}

	// Create buttons
	menu, err := creatInlineMenuWithHandler(s.group, []inlineBtnWithHandlerDescr{
		{
			text:    "Да",
			unique:  "listTasks" + deal.Id.String(),
			handler: s.onListTasks,
			payload: s.payload, // Contains deal
		},
		{
			text:    "Нет",
			unique:  "goToStart",
			handler: s.OnStart,
		},
	})
	if err != nil {
		s.sendError(c, err)
	}
	return c.Send("Нужно ли закрыть задачу по этой сделке?", menu)
}

// Lists deal tasks
func (s *session) onListTasks(c tele.Context) error {
	// Check for payload
	if c.Data() == "" {
		s.logger.Debug("no btn payload")
		return s.sendError(c, api.ErrorInvalidBtnPayload)
	}
	// Parse payload
	deal := bxtypes.Deal{}
	if err := json.Unmarshal([]byte(c.Data()), &deal); err != nil {
		s.logger.Debug(err.Error())
		return s.sendError(c, api.ErrorInvalidBtnPayload)
	}

	// Request tasks
	tasks, err := s.bxUser.ListDealTasks(deal.Id)
	if err != nil {
		return s.sendError(c, err)
	}

	// Case when no deals found
	if len(tasks) == 0 {
		if err = c.Send("Нет открытых задач."); err != nil {
			return s.sendError(c, err)
		}
		return s.onDealActions(c)
	}

	// Prepare buttons
	btns := []inlineBtnDescr{}
	r, _ := regexp.Compile("по сделке.*")
	for _, t := range tasks {
		btns = append(btns, inlineBtnDescr{
			text:   r.ReplaceAllLiteralString(t.Title, ""),
			unique: "selectTask" + t.Id.String(),
			payload: taskBtnPayload{
				Deal: deal,
				Task: t,
			},
		})
	}
	menu, err := creatInlineMenu(s.group, s.onCompleteTask, btns)
	if err != nil {
		return s.sendError(c, err)
	}
	return c.Send("Выберете задачу для завершения:", menu)
}

// Completes selected task
func (s *session) onCompleteTask(c tele.Context) error {
	// Check for payload
	if c.Data() == "" {
		s.logger.Debug("no btn payload")
		return s.sendError(c, api.ErrorInvalidBtnPayload)
	}
	// Parse payload
	data := taskBtnPayload{}
	if err := json.Unmarshal([]byte(c.Data()), &data); err != nil {
		s.logger.Debug(err.Error())
		return s.sendError(c, api.ErrorInvalidBtnPayload)
	}

	// Make request
	if err := s.bxUser.CompleteTask(data.Task.Id); err != nil {
		return s.sendError(c, err)
	}

	// Send report
	if err := c.Send(fmt.Sprintf("Задача <i>%s</i> успешно завершена.", data.Task.Title)); err != nil {
		return s.sendError(c, err)
	}

	return s.OnStart(c)
}

// Supporting functions

// Creates inline menu
func creatInlineMenu(group *tele.Group, handler tele.HandlerFunc, btns []inlineBtnDescr) (*tele.ReplyMarkup, error) {
	// Setup buttons
	rows := []tele.Row{}
	menu := &tele.ReplyMarkup{}
	for _, b := range btns {
		payload, err := json.Marshal(b.payload)
		if err != nil {
			return nil, fmt.Errorf("marshal btn payload: %w", err)
		}
		slog.Debug(fmt.Sprint(b.payload))
		slog.Debug(string(payload))
		btn := menu.Data(b.text, b.unique, string(payload)) // Attach index of deal in deals array
		group.Handle(&btn, handler)
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)
	return menu, nil
}

// The same thing but with custom handler for every button
func creatInlineMenuWithHandler(group *tele.Group, btns []inlineBtnWithHandlerDescr) (*tele.ReplyMarkup, error) {
	// Setup buttons
	rows := []tele.Row{}
	menu := &tele.ReplyMarkup{}
	for _, b := range btns {
		payload, err := json.Marshal(b.payload)
		if err != nil {
			return nil, fmt.Errorf("marshal btn payload: %w", err)
		}
		btn := menu.Data(b.text, b.unique, string(payload)) // Attach index of deal in deals array
		group.Handle(&btn, b.handler)
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)
	return menu, nil
}

// Function that analise !my !internal errors and log/ sends report
func (s *session) sendError(c tele.Context, err error) error {
	addFooter, str := api.ErrorText(err)

	s.logger.Warn(str)

	if addFooter {
		str += "\n\nДля перезапуска отправьте команду <code>/start</code>"
	}

	return c.Send(str)
}
