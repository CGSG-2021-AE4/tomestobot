package bxtypes

import (
	"fmt"
)

type Response[T any] struct {
	Result T `json:"result"`
	// Time   Time `json:"time"` // Do not need now
}

type ArrayResponse[T any] struct {
	Result []T `json:"result"`
	Total  int `json:"total"` // For array results
	// Time   Time `json:"time"` // Do not need now
}

type ResponseError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

func (resp *ResponseError) Error() string {
	return fmt.Sprintf("server responded with error code: %s descr: %s", resp.Code, resp.Description)
}

// Other response structs

type ResTasksTaskList struct {
	Tasks []Task `json:"tasks"`
}
