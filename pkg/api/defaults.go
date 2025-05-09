package api

import (
	db "github.com/sakojpa/tasker/pkg/database"
)

var (
	dateFormat = "20060102"
)

// TasksResp represents a list of tasks returned in API responses.
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}
