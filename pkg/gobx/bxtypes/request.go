package bxtypes

// Naming notation for request parameters: Req<request name in camal case>

type ReqArrayParams struct {
	Select []string          `json:"SELECT"`
	Order  map[string]string `json:"ORDER"`
	Filter map[string]string `json:"FILTER"`
}

type ReqUserGet struct { // Do not use ReqArrayParams because there is no select here
	Sort      string            `json:"SORT"`
	Order     string            `json:"ORDER"`
	Filter    map[string]string `json:"FILTER"`
	AdminMode bool              `json:"ADMIN_MODE"`
	Start     int               `json:"START"`
}

type ReqCrmDealList struct {
	ReqArrayParams
	Start int `json:"START"`
}

type ReqCrmTimelineCommentAddFields struct {
	EntityId   Id     `json:"ENTITY_ID"`
	EntityType string `json:"ENTITY_TYPE"`
	Comment    string `json:"COMMENT"`
	AuthorId   Id     `json:"AUTHOR_ID"`
	// There are also files but I don't use them now
}

type ReqCrmTimelineCommentAdd struct {
	Fields ReqCrmTimelineCommentAddFields `json:"fields"`
}

type ReqTasksTaskList struct {
	Select []string          `json:"select"`
	Order  map[string]string `json:"order"`
	Filter map[string]string `json:"filter"`
	Limit  int               `json:"limit"`
	Start  int               `json:"start"`
}

type ReqTasksTaskComplete struct {
	TaskId Id `json:"taskId"`
}
