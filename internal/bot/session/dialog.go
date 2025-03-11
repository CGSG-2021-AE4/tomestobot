package session

import (
	"fmt"
)

// Dialog flow handles the right order of requests

type DialogState int

const (
	DialogStarted = DialogState(iota)
	DialogDealsList
	DialogDealActions
	DialogWriteComment
	DialogAddComment
	DialogTasksList
	DialogTaskComplete
)

func (s DialogState) String() string {
	switch s {
	case DialogStarted:
		return "DialogStarted"
	case DialogDealsList:
		return "DialogDealsList"
	case DialogDealActions:
		return "DialogDealActions"
	case DialogWriteComment:
		return "DialogWriteComment"
	case DialogAddComment:
		return "DialogAddComment"
	case DialogTasksList:
		return "DialogTasksList"
	case DialogTaskComplete:
		return "DialogTaskComplete"
	}
	return "unknown"
}

type DialogFlow interface {
	Set(state DialogState) error // Is executed to check if previous state realy was before and that it comleted without erros
	Done()                       // Executed by defer at the end of the function to let know that there were no errors and the state is fully ended
	// If it was not executed DialogFlow guess considers that there was an error during the state
	// I take into account errors from my part(bitrix) if there is an error on telegram's side flow will broke anyway

	// For cases that require other treatment
	// For example DailogAddComment
	Get() DialogState
	IsDone() bool
}

type dialogFlow struct {
	state DialogState
	done  bool
}

func NewDialogFlow() DialogFlow {
	return &dialogFlow{
		state: DialogStarted,
		done:  true,
	}
}

func (f *dialogFlow) Set(newState DialogState) error {
	if !f.done {
		return fmt.Errorf("previous state %s is not complete", f.state)
	}
	err := fmt.Errorf("invalid previous state: %s", f.state) // For future
	switch newState {
	case DialogStarted: // For explicity
		break
	case DialogDealsList:
		if f.state != DialogStarted {
			return err
		}
	case DialogDealActions:
		if f.state != DialogDealsList && f.state != DialogAddComment {
			return err
		}
	case DialogWriteComment:
		if f.state != DialogDealActions {
			return err
		}
	case DialogAddComment:
		if f.state != DialogWriteComment {
			return nil // The rest is handled in session
		}
	case DialogTasksList:
		if f.state != DialogDealActions {
			return err
		}
	case DialogTaskComplete:
		if f.state != DialogTasksList {
			return err
		}
	}
	f.done = false
	f.state = newState
	return nil
}

func (f *dialogFlow) Done() {
	f.done = true
}

func (f *dialogFlow) Get() DialogState {
	return f.state
}

func (f *dialogFlow) IsDone() bool {
	return f.done
}
