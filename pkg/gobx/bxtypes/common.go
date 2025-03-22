package bxtypes

import (
	"strconv"
)

// Now in this package there are only the types I need
// The same is for their fields - only the ones I use
// I do not want to add all possible values

// Here are request body data structures

// User

type User struct {
	Id       Id     `json:"ID"`
	Name     string `json:"NAME"`
	LastName string `json:"LAST_NAME"`
}

var NilUser = User{
	Id:       0,
	Name:     "",
	LastName: "",
}

// Deal

type Deal struct {
	Id         Id     `json:"ID"`
	Title      string `json:"TITLE"`
	TypeId     string `json:"TYPE_ID"`
	CategoryId string `json:"CATEGORY_ID"`
	StageId    string `json:"STAGE_ID"`
}

var NilDeal = Deal{
	Id:         0,
	Title:      "",
	TypeId:     "",
	CategoryId: "",
	StageId:    "",
}

// Deal stages
// Now only constants
const (
	DealStageNew               = "C1:NEW"
	DealStagePreparation       = "C1:PREPARATION"
	DealStageGetDecision       = "C1:9"
	DealStagePrepaymentInvoice = "C1:PREPAYMENT_INVOICE"
	DealStageExecuting         = "C1:EXECUTING"
)

func DealStageText(str string) string {
	switch str {
	case DealStageNew:
		return "Новая сделка"
	case DealStagePreparation:
		return "Сделать предложение"
	case DealStageGetDecision:
		return "Получить решение"
	case DealStagePrepaymentInvoice:
		return "Получить анкету"
	case DealStageExecuting:
		return "Получить договор"
	}
	return str // Default case if is is unknown
}

// Task status type

type TaskState int

const (
	TaskStateNew = TaskState(iota + 1)
	TaskStatePending
	TaskStateInProgress
	TaskStateSupposedlyCompleted
	TaskStateCompleted
	TaskStateDeferred
	TaskStateDeclined
)

func (s TaskState) String() string {
	return strconv.Itoa(int(s))
}

func (s *TaskState) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' { // Because in Bitrix' responses id is sometimes number sometimes string...!?
		b = b[1 : len(b)-1]
	}
	sn, err := strconv.Atoi(string(b))
	*s = TaskState(sn)
	return err
}

type Task struct {
	Id     Id        `json:"ID"`
	Title  string    `json:"TITLE"`
	Status TaskState `json:"STATUS"`
}

var NilTask = Task{
	Id:     0,
	Title:  "",
	Status: 0,
}

// Resource id type

type Id int

func (id Id) String() string {
	return strconv.Itoa(int(id))
}

func (id *Id) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' { // Because in Bitrix' responses id is sometimes number sometimes string...!?
		b = b[1 : len(b)-1]
	}
	i, err := strconv.Atoi(string(b))
	*id = Id(i)
	return err
}
