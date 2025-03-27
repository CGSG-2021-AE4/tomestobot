package session

import (
	"encoding/binary"
	"encoding/hex"
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

	// Payload
	deals     TaggedVar[[]bxtypes.Deal]
	deal      TaggedVar[bxtypes.Deal]
	dealTasks TaggedVar[tasksPayload]

	// Comment - the difficulty is that the msg is just text
	waitingForComment bool          // Writing comment is toggled and now I waiting for text message that will be treated like a comment
	writeCommentMsg   tele.Editable // For future deletion
	addCommentPayload string        // Exception - supposed to be in msg data field

	// Previous request msg for deletion
	prevMsg tele.Editable
}

// Supplement structures
// For inline menu creation functions
type inlineBtnDescr struct {
	text    string
	unique  string
	payload string // Will be marshalled to json later
}

// The same descriptor but with handler function
type inlineBtnWithHandlerDescr struct {
	text    string
	unique  string
	payload string // Will be marshalled to json later
	handler tele.HandlerFunc
}

type tasksPayload struct {
	deal  bxtypes.Deal
	tasks []bxtypes.Task
}

// Create session function
func createSession(logger *slog.Logger, bot *tele.Bot, group *tele.Group, user api.BxUser) *session {
	s := &session{
		logger: logger,
		bot:    bot,
		group:  group,
		bxUser: user,

		// Dynamic data
		deals:     newTaggedVar[[]bxtypes.Deal](),
		deal:      newTaggedVar[bxtypes.Deal](),
		dealTasks: newTaggedVar[tasksPayload](),

		waitingForComment: false,
		writeCommentMsg:   nil,
		addCommentPayload: "",
	}
	// Setup some handlers
	s.group.Handle(tele.OnText, s.onAddComment)

	return s
}

// Commands handlers
// Are named after commands or actions they execute

// Main message - may be consider as help
func (s *session) OnStart(c tele.Context) error {
	s.clearPrev()
	menu := &tele.ReplyMarkup{}

	listDealsBtn := menu.Data("Показать открытые сделки", "list_deals")
	s.group.Handle(&listDealsBtn, s.onListDeals)

	menu.Inline(
		menu.Row(listDealsBtn),
	)

	return s.ask(c, fmt.Sprintf("Здравствуйте, %s %s.\nВыберете действие.", s.bxUser.Get().Name, s.bxUser.Get().LastName), menu)
}

// Handles list deals message
func (s *session) onListDeals(c tele.Context) error {
	s.clearPrev()
	s.logger.Debug("on list deals")

	// Get deals
	deals, err := s.bxUser.ListDeals()
	if err != nil {
		return s.sendError(c, err)
	}
	s.logger.Debug(fmt.Sprint(deals))

	// Store deals
	tagBytes := s.deals.Set(deals).Bytes()

	// Case when no deals found
	if len(deals) == 0 {
		if err = c.Send("Не найдено открытых сделок."); err != nil {
			return s.sendError(c, err)
		}
		return s.OnStart(c)
	}

	// Prepare buttons descriptors
	btnDescrs := []inlineBtnDescr{}
	for i, d := range deals {
		// Encode index
		iBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(iBytes, uint32(i))
		payload := append(tagBytes[:], iBytes...)
		s.logger.Debug(fmt.Sprint(payload))
		// Add button descriptor
		btnDescrs = append(btnDescrs, inlineBtnDescr{
			text:    d.Title,
			unique:  "dealBtn" + d.Id.String(),
			payload: hex.EncodeToString(payload),
		})
	}
	menu, err := creatInlineMenu(s.group, s.onDealActions, btnDescrs)
	if err != nil {
		s.sendError(c, err)
	}
	return s.ask(c, "Выберете сделку:", menu)
}

// Shows actions with select deal
func (s *session) onDealActions(c tele.Context) error {
	s.clearPrev()
	// Get deal of the button
	s.logger.Debug(c.Data())
	s.logger.Debug(fmt.Sprint([]byte(c.Data())))

	// Decode payload
	tag, i, err := decodeTagWithI(c.Data())
	if err != nil {
		s.logger.Debug("decode tag with I err")
		return s.sendError(c, err) // Already typed err
	}
	deals, err := s.deals.Get(tag)
	if err != nil {
		s.logger.Debug("get deals invalid tag")
		return s.sendError(c, err) // Already typed err
	}
	s.logger.Debug(fmt.Sprint(deals))
	s.logger.Debug(fmt.Sprint(i))
	if i >= len(deals) { // To be sure its ok
		return s.sendError(c, fmt.Errorf("invalid deal index"))
	}

	// Save selected deal and encode tag to payload
	deal := deals[i]
	tagBytes := s.deal.Set(deal).Bytes()
	payload := hex.EncodeToString(tagBytes[:])

	// Create buttons
	menu, err := creatInlineMenuWithHandler(s.group, []inlineBtnWithHandlerDescr{
		{
			text:    "Добавить коментарий",
			unique:  "addComment" + deal.Id.String(),
			handler: s.onWriteComment,
			payload: payload,
		},
		{
			text:    "Показать открытые задачи",
			unique:  "listTasks" + deal.Id.String(),
			handler: s.onListTasks,
			payload: payload,
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
	return s.ask(c, fmt.Sprintf("Сделка: <i>%s</i>\nСтатус: <i>%s</i>\n\nВыберете действие:", deal.Title, bxtypes.DealStageText(deal.StageId)), menu)
}

// Asks to write a coomment
func (s *session) onWriteComment(c tele.Context) error {
	// Send message
	msg, err := s.bot.Send(c.Sender(), "Напишите коментарий:")
	if err != nil {
		return err
	}
	// Save payload
	s.writeCommentMsg = msg
	s.addCommentPayload = c.Data() // Redirect payload from button
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

	// Clear previous
	s.clearPrev()
	s.waitingForComment = false // Remove flag before any error
	if s.writeCommentMsg != nil {
		if err := s.bot.Delete(s.writeCommentMsg); err != nil {
			return s.sendError(c, err)
		}
		s.writeCommentMsg = nil
	}
	defer s.bot.Delete(c.Message())

	s.logger.Debug("onAddComment", "msg", c.Text())

	// Decode payload
	tag, err := decodeTag(s.addCommentPayload)
	if err != nil {
		s.logger.Debug("decode tag with I err")
		return s.sendError(c, err) // Already typed err
	}
	deal, err := s.deal.Get(tag)
	if err != nil {
		s.logger.Debug("get deal invalid tag")
		return s.sendError(c, err) // Already typed err
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
			payload: s.addCommentPayload, // Contains deal
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
	return s.ask(c, "Нужно ли закрыть задачу по этой сделке?", menu)
}

// Lists deal tasks
func (s *session) onListTasks(c tele.Context) error {
	// Decode payload
	s.logger.Debug(c.Data())
	tag, err := decodeTag(c.Data())
	if err != nil {
		s.logger.Debug("decode tag with I err")
		return s.sendError(c, err) // Already typed err
	}
	deal, err := s.deal.Get(tag)
	if err != nil {
		s.logger.Debug("get deal invalid tag")
		return s.sendError(c, err) // Already typed err
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
		return s.OnStart(c)
	}

	// Save tasks and encode tag
	tagBytes := s.dealTasks.Set(tasksPayload{
		deal:  deal,
		tasks: tasks,
	}).Bytes()

	// Prepare buttons
	btns := []inlineBtnDescr{}
	r, _ := regexp.Compile("по сделке.*")
	for i, t := range tasks {
		iBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(iBytes, uint32(i))
		payload := append(tagBytes[:], iBytes...)
		btns = append(btns, inlineBtnDescr{
			text:    r.ReplaceAllLiteralString(t.Title, ""),
			unique:  "selectTask" + t.Id.String(),
			payload: hex.EncodeToString(payload),
		})
	}
	menu, err := creatInlineMenu(s.group, s.onCompleteTask, btns)
	if err != nil {
		return s.sendError(c, err)
	}
	return s.ask(c, "Выберете задачу для завершения:", menu)
}

// Completes selected task
func (s *session) onCompleteTask(c tele.Context) error {
	// Decode payload
	tag, i, err := decodeTagWithI(c.Data())
	if err != nil {
		s.logger.Debug("decode tag with I err")
		return s.sendError(c, err) // Already typed err
	}
	tasksPayload, err := s.dealTasks.Get(tag)
	if err != nil {
		s.logger.Debug("get deal tasks invalid tag")
		return s.sendError(c, err) // Already typed err
	}
	s.logger.Debug(fmt.Sprint(tasksPayload))
	s.logger.Debug(fmt.Sprint(i))
	if i >= len(tasksPayload.tasks) { // To be sure its ok
		return s.sendError(c, fmt.Errorf("invalid task index"))
	}

	task := tasksPayload.tasks[i]

	// Make request
	if err := s.bxUser.CompleteTask(task.Id); err != nil {
		return s.sendError(c, err)
	}

	// Send report
	if err := c.Send(fmt.Sprintf("Задача <i>%s</i> успешно завершена.", task.Title)); err != nil {
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
		slog.Debug(b.payload)
		btn := menu.Data(b.text, b.unique, b.payload) // Attach index of deal in deals array
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
		btn := menu.Data(b.text, b.unique, b.payload) // Attach index of deal in deals array
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

// Sends message and saves it for future deletion
func (s *session) ask(c tele.Context, what any, opts ...any) error {
	msg, err := s.bot.Send(c.Chat(), what, opts...)
	if err != nil {
		return err
	}
	s.prevMsg = msg // Do not check if it is nil
	return nil
}

// Deletes previous qustion message
func (s *session) clearPrev() error {
	if prev := s.prevMsg; prev != nil {
		s.prevMsg = nil
		return s.bot.Delete(prev)
	}
	return nil
}
