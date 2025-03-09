package bx

import (
	"fmt"
	"log"

	"tomestobot/pkg/gobx/bxclient"
	"tomestobot/pkg/gobx/bxtypes"
)

type bxUser struct {
	bx bxclient.BxClient
	id bxtypes.Id
}

func (u *bxUser) ListDeals() ([]bxtypes.Deal, error) {
	// Make request
	resp, err := u.bx.Do(
		"crm.deal.list",
		bxtypes.ReqCrmDealList{
			ReqArrayParams: bxtypes.ReqArrayParams{
				Select: []string{"ID", "TITLE", "TYPE_ID", "CATEGORY_ID", "STAGE_ID"},
				Filter: map[string]string{
					// What is here TODO
				},
			},
		},
		&bxtypes.ArrayResponse[bxtypes.Deal]{})

	// Check for result to be valid
	if err != nil {
		return nil, fmt.Errorf("during request: %w", err)
	}
	res, ok := resp.(*bxtypes.ArrayResponse[bxtypes.Deal])
	if !ok {
		return nil, fmt.Errorf("failed to parse response")
	}

	return res.Result, nil
}

func (u *bxUser) AddCommentToDeal(dealId bxtypes.Id, comment string) error {
	// Make request
	resp, err := u.bx.Do(
		"crm.timeline.comment.add",
		bxtypes.ReqCrmTimelineCommentAdd{
			Fields: bxtypes.ReqCrmTimelineCommentAddFields{
				EntityId:   dealId,
				EntityType: "deal",
				AuthorId:   u.id,
				Comment:    comment,
			},
		},
		&bxtypes.Response[bxtypes.ResCrmTimelineCommentAdd]{})

	// Check for result to be valid
	if err != nil {
		return fmt.Errorf("during request: %w", err)
	}
	res, ok := resp.(*bxtypes.Response[bxtypes.ResCrmTimelineCommentAdd])
	if !ok {
		return fmt.Errorf("failed to parse response")
	}

	log.Print(res.Result)

	return nil
}

func (u *bxUser) ListDealTasks(dealId bxtypes.Id) ([]bxtypes.Task, error) {
	// Make request
	resp, err := u.bx.Do(
		"tasks.task.list",
		bxtypes.ReqTasksTaskList{
			Select: []string{"ID", "TITLE", "STATUS", "UF_CRM_TASK"},
			Filter: map[string]string{
				"<REAL_STATUS": "5", // Now there are only incomplete ones TODO
				"UF_CRM_TASK":  "D_" + dealId.String(),
			},
			Order: map[string]string{},
		},
		&bxtypes.Response[bxtypes.ResTasksTaskList]{})

	// Check for result to be valid
	if err != nil {
		return nil, fmt.Errorf("during request: %w", err)
	}
	res, ok := resp.(*bxtypes.Response[bxtypes.ResTasksTaskList])
	if !ok {
		return nil, fmt.Errorf("failed to parse response")
	}

	return res.Result.Tasks, nil
}

func (u *bxUser) CompleteTask(taskId bxtypes.Id) error {
	// Make request
	_, err := u.bx.Do(
		"tasks.task.complete",
		bxtypes.ReqTasksTaskComplete{
			TaskId: taskId,
		},
		&bxtypes.Response[any]{})

	// Check for result to be valid
	if err != nil {
		return fmt.Errorf("during request: %w", err)
	}
	return nil
}

func (u *bxUser) GetId() bxtypes.Id {
	return u.id
}

func (u *bxUser) Close() error {
	return nil
}
