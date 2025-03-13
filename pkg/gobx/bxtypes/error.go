package bxtypes

import "fmt"

// This file contains different error types by level

type ErrorResty struct { // Resty POST request finished with error
	Err error
}

type ErrorStatusCode int // When code is >=400

// Errors for future use

type ErrorResponse int // Errors with response like parsing etc.
type ErrorInternal int // My internal errors e.g. no user found

// error interface implementation

func (e ErrorResty) Error() string {
	return fmt.Sprintf("resty: %s", e.Err.Error())
}

func (code ErrorStatusCode) Error() string {
	return fmt.Sprintf("status code %d", int(code))
}

func (code ErrorResponse) Error() string {
	return fmt.Sprintf("response %d", int(code))
}

func (code ErrorInternal) Error() string {
	return fmt.Sprintf("internal %d", int(code))
}
