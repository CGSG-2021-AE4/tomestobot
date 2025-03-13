package api

import "tomestobot/pkg/gobx/bxtypes"

// Some internal errors
const (
	// Bx
	ErrorUserNotFound = bxtypes.ErrorInternal(iota)
	ErrorSeveralUsersFound

	// Dialog
	ErrorDialogInvalidOrder
	ErrorDialogPrevStateNotComplete
)

func ErrorInternalText(err bxtypes.ErrorInternal) string {
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
