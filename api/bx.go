package api

import (
	"context"
	"io"

	"github.com/CGSG-2021-AE4/gobx/types"
)

type BxUser interface {
	ListDeals(ctx context.Context) ([]types.Deal, error)                    // Deals that are accessable for this user. Later add stage as filter
	AddCommentToDeal(ctx context.Context, dealId int, comment string) error // Add comment to this deal
	ListDealTasks(ctx context.Context, dealId int) ([]types.Task, error)    // List tasks that are attached to this deal and are not complete
	CompleteTask(ctx context.Context, taskId int) error                     // Compete the task
	io.Closer
}

type BxWrapper interface {
	AuthUserByPhone(phone string) (BxUser, error) // Check if there is a user with this number and save if it is. Nil if auth is successful, error if not
	io.Closer
}
