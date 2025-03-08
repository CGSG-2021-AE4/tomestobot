package bx

import (
	"fmt"

	"github.com/CGSG-2021-AE4/gobx/client"
	"github.com/CGSG-2021-AE4/gobx/types"
)

type bxUser struct {
	bx client.BxClient
	id types.Id
}

func (u *bxUser) ListDeals() ([]types.Deal, error) {
	// Make request
	resp, err := u.bx.Do(
		"crm.deal.list",
		types.ReqCrmDealList{
			ReqArrayParams: types.ReqArrayParams{
				Select: []string{"ID", "TITLE", "TYPE_ID", "CATEGORY_ID", "STAGE_ID"},
				Filter: map[string]string{
					// What is here TODO
				},
			},
		},
		&types.ArrayResponse[types.Deal]{})

	// Check for result to be valid
	if err != nil {
		return nil, fmt.Errorf("during request: %w", err)
	}
	res, ok := resp.(*types.ArrayResponse[types.Deal])
	if !ok {
		return nil, fmt.Errorf("failed to parse response")
	}

	return res.Result, nil
}
func (u *bxUser) AddCommentToDeal(dealId types.Id, comment string) error {

	return fmt.Errorf("not implemented yet")
}
func (u *bxUser) ListDealTasks(dealId types.Id) ([]types.Task, error) {
	// Make request
	resp, err := u.bx.Do(
		"tasks.task.list",
		types.ReqCrmDealList{
			ReqArrayParams: types.ReqArrayParams{
				Select: []string{"ID", "TITLE", "STATUS"},
				Filter: map[string]string{
					"UF_CRM_TASK": "D_" + dealId.String(),
				},
			},
		},
		&types.Response[types.ResTasksTaskList]{})

	// Check for result to be valid
	if err != nil {
		return nil, fmt.Errorf("during request: %w", err)
	}
	res, ok := resp.(*types.Response[types.ResTasksTaskList])
	if !ok {
		return nil, fmt.Errorf("failed to parse response")
	}

	return res.Result.Tasks, nil
}

func (u *bxUser) CompleteTask(taskId types.Id) error {
	return fmt.Errorf("not implemented yet")
}

func (u *bxUser) GetId() types.Id {
	return u.id
}

func (u *bxUser) Close() error {
	return nil
}
