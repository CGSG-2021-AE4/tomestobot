package bx

import (
	"github.com/CGSG-2021-AE4/tomestobot/api"
	"github.com/CGSG-2021-AE4/tomestobot/pkg/gobx/bxclient"
	"github.com/CGSG-2021-AE4/tomestobot/pkg/gobx/bxtypes"
)

type bxUser struct {
	bx   bxclient.BxClient
	user bxtypes.User
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
					"ASSIGNED_BY_ID": u.user.Id.String(),
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
				AuthorId:   u.user.Id,
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
				"RESPONSIBLE_ID": u.user.Id.String(),
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

func (u *bxUser) Get() bxtypes.User {
	return u.user
}

func (u *bxUser) Close() error {
	return nil
}
