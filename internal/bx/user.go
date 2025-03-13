package bx

import (
	"github.com/charmbracelet/log"

	"tomestobot/api"
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
					"ASSIGNED_BY_ID": u.id.String(),
				},
			},
		},
		&bxtypes.ArrayResponse[bxtypes.Deal]{})

	// Check for result to be valid
	if err != nil {
		return nil, err
	}
	res, ok := resp.Result().(*bxtypes.ArrayResponse[bxtypes.Deal])
	if !ok {
		return nil, api.ErrorParseResponse
	}

	return res.Result, nil
}

func (u *bxUser) AddCommentToDeal(dealId bxtypes.Id, comment string) (bxtypes.Id, error) {
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
		return 0, err
	}
	res, ok := resp.Result().(*bxtypes.Response[bxtypes.ResCrmTimelineCommentAdd])
	if !ok {
		return 0, api.ErrorParseResponse
	}

	log.Print(res.Result)

	return bxtypes.Id(res.Result), nil
}

func (u *bxUser) ListDealTasks(dealId bxtypes.Id) ([]bxtypes.Task, error) {
	// Make request
	resp, err := u.bx.Do(
		"tasks.task.list",
		bxtypes.ReqTasksTaskList{
			Select: []string{"ID", "TITLE", "STATUS", "UF_CRM_TASK"},
			Filter: map[string]string{
				"<REAL_STATUS":   "5", // Now there are only incomplete ones TODO
				"RESPONSIBLE_ID": u.id.String(),
				"UF_CRM_TASK":    "D_" + dealId.String(),
			},
			Order: map[string]string{},
		},
		&bxtypes.Response[bxtypes.ResTasksTaskList]{})

	// Check for result to be valid
	if err != nil {
		return nil, err
	}
	res, ok := resp.Result().(*bxtypes.Response[bxtypes.ResTasksTaskList])
	if !ok {
		return nil, api.ErrorParseResponse
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
		return err
	}
	return nil
}

func (u *bxUser) GetId() bxtypes.Id {
	return u.id
}

func (u *bxUser) Close() error {
	return nil
}
