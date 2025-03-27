package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/CGSG-2021-AE4/tomestobot/pkg/gobx/bxtypes"
)

type ErrorInternal int // My internal errors e.g. no user found

// Some internal errors
const (
	// Bx
	ErrorUserNotFound = ErrorInternal(iota)
	ErrorSeveralUsersFound
	ErrorNoContactInMsg
	ErrorInvalidPhoneNumber

	// Session
	ErrorInvalidBtnPayload // Invalid payload len for ex

	// Tag
	ErrorInvalidTag
)

func ErrorInternalText(err ErrorInternal) string {
	switch err {
	case ErrorUserNotFound:
		return "UserNotFound"
	case ErrorSeveralUsersFound:
		return "SeveralUsersFound"
	case ErrorNoContactInMsg:
		return "ErrorNoContactInMsg"
	case ErrorInvalidBtnPayload:
		return "ErrorInvalidBtnPayload"
	case ErrorInvalidPhoneNumber:
		return "ErrorInvalidPhoneNumber"
	case ErrorInvalidTag:
		return "ErrorInvalidTag"
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
			return false, "Пользователь с таким номером не найден."
		case ErrorSeveralUsersFound:
			return false, "Ошибка: в системе зарегистрировано несколько пользователей с таким номером, обратитесь к администрации."
		}
		return true, fmt.Sprintf("ERROR:\n<code>internal level: %s</code>", ErrorInternalText(err))
	}

	return true, fmt.Sprintf("ERROR:\n<code>unknown level: %s</code>", err.Error())
}

// Global flags
// Like env var but with boolean type
// The easiest and not error prone way

// By default logs only warnings

// Enables debug logs
var EnableDebugLogs = false

// Enables full resty requests logs - separate var because these logs are very big
var EnableRestyLogs = false

// Setups this variable
func SetupGlobalFlags() {
	EnableDebugLogs = os.Getenv("ENABLE_DEBUG_LOGS") == "true"
	EnableRestyLogs = os.Getenv("ENABLE_RESTY_LOGS") == "true"
}
