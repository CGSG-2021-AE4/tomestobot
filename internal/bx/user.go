package bx

import (
	"context"
	"fmt"
	"tomestobot/api"

	"github.com/CGSG-2021-AE4/gobx/types"
)

type bxUser struct {
	bx     api.BxWrapper
	userId string
}

func (u *bxUser) ListDeals(ctx context.Context) ([]types.Deal, error) {
	return nil, fmt.Errorf("not implemented yet")
}
func (u *bxUser) AddCommentToDeal(ctx context.Context, dealId int, comment string) error {

	return fmt.Errorf("not implemented yet")
}
func (u *bxUser) ListDealTasks(ctx context.Context, dealId int) ([]types.Task, error) {

	return nil, fmt.Errorf("not implemented yet")
}
func (u *bxUser) CompleteTask(ctx context.Context, taskId int) error {
	return fmt.Errorf("not implemented yet")
}

func (u *bxUser) Close() error {
	return nil
}
