package bot

import (
	"tomestobot/api"
)

type DialogState int

const (
	Started = DialogState(iota)
	DealsSelect
	DealActionSelect
	WriteComment
	TasksSelect
)

type userSession struct {
	tgID   int64
	bxUser api.BxUser
	state  DialogState
	// lastAction time.Time TODO
}

// Resets state to started
func (s *userSession) reset() {
	s.state = Started
}
