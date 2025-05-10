package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	db "github.com/sakojpa/tasker/pkg/database"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Task struct{}

func (a Task) checkBody(r *http.Request) (*db.Task, error, int) {
	var task db.Task
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("body read error"), http.StatusInternalServerError
	}
	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		return nil, fmt.Errorf("marshaller error"), http.StatusInternalServerError
	}
	if strings.TrimSpace(task.Title) == "" {
		return nil, fmt.Errorf("task title absent"), http.StatusBadRequest
	}
	now := time.Now()
	if task.Date == "" || strings.TrimSpace(task.Date) == "" {
		task.Date = now.Format(dateFormat)
	}
	parsedTime, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return nil, fmt.Errorf("wrong date format, expected format is %s", dateFormat), http.StatusBadRequest
	}
	if afterNow(now, parsedTime) {
		if len(task.Repeat) == 0 {
			task.Date = now.Format(dateFormat)
		} else {
			next, err := validateRepeatRule(now, task.Date, task.Repeat)
			if err != nil {
				return nil, fmt.Errorf("repeat rule error: %w", err), http.StatusBadRequest
			}
			task.Date = next.NextDate
		}
	}
	return &task, nil, 0
}

// AddTaskHandler adds a new task to the database and returns its ID.
func AddTaskHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	addTask := Task{}
	task, err, code := addTask.checkBody(r)
	if err != nil {
		sentErrorJson(w, err.Error(), code)
		return
	}
	id, err := db.CreateTask(ctx, task)
	if err != nil {
		http.Error(w, "Database error", code)
		return
	}
	resp := struct {
		ID int64 `json:"id"`
	}{ID: id}
	sentOkMsg(w, resp, 1)
	return
}

// UpdateTaskHandler handles updating a specific task by its ID.
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	updateTask := Task{}
	task, err, code := updateTask.checkBody(r)
	if err != nil {
		sentErrorJson(w, err.Error(), code)
		return
	}
	if strings.TrimSpace(task.ID) == "" {
		sentErrorJson(w, "Task ID absent", http.StatusBadRequest)
		return
	}
	if _, err := strconv.Atoi(task.ID); err != nil {
		sentErrorJson(w, "Task ID error", http.StatusBadRequest)
		return
	}
	err = db.UpdateTask(ctx, task, task.ID)
	if err != nil {
		sentErrorJson(w, err.Error(), http.StatusBadRequest)
		return
	}
	sentOkMsg(w, struct{}{}, 1)
	return
}

// DeleteTaskHandler removes a task by its ID from the database.
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	id := r.FormValue("id")
	err := db.DeleteTaskById(ctx, id)
	if err != nil {
		sentErrorJson(w, err.Error(), http.StatusBadRequest)
		return
	}
	sentOkMsg(w, struct{}{}, 1)
	return
}

// EditTaskHandler handles retrieving a specific task by its ID.
func EditTaskHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	id := r.FormValue("id")
	task, err := db.GetTaskById(ctx, id)
	if err != nil {
		sentErrorJson(w, err.Error(), http.StatusBadRequest)
		return
	}
	sentOkMsg(w, task, 1)
	return
}
