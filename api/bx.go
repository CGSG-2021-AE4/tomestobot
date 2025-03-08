package api

import (
	"io"

	"github.com/CGSG-2021-AE4/gobx/types"
)

type BxUser interface {
	ListDeals() ([]types.Deal, error)                       // Deals that are accessable for this user. Later add stage as filter
	AddCommentToDeal(dealId types.Id, comment string) error // Add comment to this deal
	ListDealTasks(dealId types.Id) ([]types.Task, error)    // List tasks that are attached to this deal and are not complete
	CompleteTask(taskId types.Id) error                     // Compete the task
	GetId() types.Id                                        // Id getter
	io.Closer
}

type BxWrapper interface {
	AuthUserByPhone(phone string) (BxUser, error) // Check if there is a user with this number and save if it is. Nil if auth is successful, error if not
	io.Closer
}
