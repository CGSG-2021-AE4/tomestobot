package api

import (
	"fmt"
	"tomestobot/pkg/gobx/bxtypes"
)

type ErrorInternal int // My internal errors e.g. no user found

// Some internal errors
const (
	// Bx
	ErrorUserNotFound = ErrorInternal(iota)
	ErrorSeveralUsersFound

	// Dialog
	ErrorDialogInvalidOrder
	ErrorDialogPrevStateNotComplete
)

func ErrorInternalText(err ErrorInternal) string {
	switch err {
	case ErrorUserNotFound:
		return "UserNotFound"
	case ErrorSeveralUsersFound:
		return "SeveralUsersFound"
	case ErrorDialogInvalidOrder:
		return "DialogInvalidOrder"
	case ErrorDialogPrevStateNotComplete:
		return "DialogPrevStateNotComplete"
	}
	return "unknown"
}

func (code ErrorInternal) Error() string {
	return fmt.Sprintf("internal %s", ErrorInternalText(code))
}

// Some response errors
const (
	ErrorParseResponse = bxtypes.ErrorResponse(iota)
)

func ErrorResponseText(err bxtypes.ErrorResponse) string {
	switch err {
	case ErrorParseResponse:
		return "ParseResponse"
	}
	return "unknown"
}
