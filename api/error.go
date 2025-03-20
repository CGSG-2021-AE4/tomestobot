package api

import (
	"fmt"
	"net/http"
	"tomestobot/pkg/gobx/bxtypes"
)

type ErrorInternal int // My internal errors e.g. no user found

// Some internal errors
const (
	// Bx
	ErrorUserNotFound = ErrorInternal(iota)
	ErrorSeveralUsersFound
	ErrorNoContactInMsg
	ErrorInvalidPhoneNumber

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
	case ErrorNoContactInMsg:
		return "ErrorNoContactInMsg"
	case ErrorInvalidPhoneNumber:
		return "ErrorInvalidPhoneNumber"

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

// Formats error string - default case
// Returns:
//   - do add help footer
//   - styled error
func ErrorText(err error) (bool, string) {
	if err, ok := err.(bxtypes.ErrorResty); ok { // Resty
		return true, fmt.Sprintf("ERROR:\n<code>resty level: %s</code>", err.Error())
	}
	if err, ok := err.(bxtypes.ErrorStatusCode); ok { // HTTP status code
		return true, fmt.Sprintf("ERROR:\n<code>http status: %s</code>", http.StatusText(int(err)))
	}
	if err, ok := err.(bxtypes.ErrorResponse); ok { // HTTP status code
		return true, fmt.Sprintf("ERROR:\n<code>with response: %s</code>", ErrorResponseText(err))
	}
	if err, ok := err.(ErrorInternal); ok { // HTTP status code
		switch err { // Special errors
		case ErrorUserNotFound:
			return false, "Пользователь не найден."
		case ErrorSeveralUsersFound:
			return false, "Ошибка: найдено несколько пользователей."
		}
		return true, fmt.Sprintf("ERROR:\n<code>internal level: %s</code>", ErrorInternalText(err))
	}

	return true, fmt.Sprintf("ERROR:\n<code>unknown level: %s</code>", err.Error())
}
