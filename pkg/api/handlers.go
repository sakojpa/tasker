package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sakojpa/tasker/config"
	db "github.com/sakojpa/tasker/pkg/database"
	"net/http"
	"time"
)

// GetAllTasksHandler retrieves all tasks filtered by search query or date.
func GetAllTasksHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	query := r.FormValue("search")
	queryType := ""
	if len(query) > 0 {
		t, err := time.Parse("02.01.2006", query)
		if err != nil {
			queryType = "text"
		} else {
			queryType = "date"
			query = t.Format(dateFormat)
		}
	}
	tasks, err := db.GetAllTasks(ctx, 50, query, queryType)
	if err != nil {
		sentErrorJson(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sentOkMsg(w, TasksResp{Tasks: tasks}, 1)
}

// RepeatTaskHandler calculates next execution date of repeated tasks using provided rules.
func RepeatTaskHandler(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	if now == "" {
		now = time.Now().Format(dateFormat)
	}
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")
	nowTime, _ := time.Parse(dateFormat, now)
	taskInfo, err := validateRepeatRule(nowTime, date, repeat)
	if err != nil {
		sentErrorJson(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sentOkMsg(w, taskInfo.NextDate, 2)
}

// DoneTaskHandler marks a task as done, deletes it if non-repeating, otherwise updates its next due date.
func DoneTaskHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	id := r.FormValue("id")
	task, err := db.GetTaskById(ctx, id)
	if err != nil {
		sentErrorJson(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(task.Repeat) == 0 {
		err = db.DeleteTaskById(ctx, id)
		if err != nil {
			sentErrorJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		sentOkMsg(w, struct{}{}, 1)
		return
	}
	newNow, _ := time.Parse(dateFormat, task.Date)
	taskInfo, err := validateRepeatRule(newNow, task.Date, task.Repeat)
	if err != nil {
		sentErrorJson(w, err.Error(), http.StatusInternalServerError)
		return
	}
	task.Date = taskInfo.NextDate
	err = db.UpdateTask(ctx, task, id)
	if err != nil {
		sentErrorJson(w, err.Error(), http.StatusBadRequest)
		return
	}
	sentOkMsg(w, struct{}{}, 1)
	return
}

// AuthHandler authenticates users by processing JSON-based login credentials.
func AuthHandler(w http.ResponseWriter, r *http.Request, c *config.Config) {
	if c.Auth.Enabled && c.Auth.Password != "" {
		auth := Auth{}
		var authRequest AuthRequest
		var buf bytes.Buffer

		_, err := buf.ReadFrom(r.Body)
		defer r.Body.Close()
		if err != nil {
			fmt.Printf("body read error: %s\n", err.Error())
			sentErrorJson(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(buf.Bytes(), &authRequest)
		if err != nil {
			fmt.Printf("unmarshal error: %s\n", err.Error())
			sentErrorJson(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		resp, err, code := auth.make(&authRequest, c)
		if err != nil {
			sentErrorJson(w, err.Error(), code)
			return
		}
		sentOkMsg(w, resp, 1)
		return
	}
}
