package api

import (
	"io"

	"github.com/CGSG-2021-AE4/tomestobot/pkg/gobx/bxtypes"
)

type BxUser interface {
	ListDeals() ([]bxtypes.Deal, error)                                     // Deals that are accessable for this user. Later add stage as filter
	AddCommentToDeal(dealId bxtypes.Id, comment string) (bxtypes.Id, error) // Add comment to this deal
	ListDealTasks(dealId bxtypes.Id) ([]bxtypes.Task, error)                // List tasks that are attached to this deal and are not complete
	CompleteTask(taskId bxtypes.Id) error                                   // Compete the task
	Get() bxtypes.User                                                      // Returns user info
	io.Closer
}

type BxWrapper interface {
	AuthUserByPhone(phone string) (BxUser, error) // Check if there is a user with this number and creates BxUser if it is. Nil if auth is successful, error if not
	AuthUserById(id bxtypes.Id) (BxUser, error)   // The same thing but not we know id
	io.Closer
}
