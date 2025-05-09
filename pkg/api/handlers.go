package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	db "github.com/sakojpa/tasker/pkg/database"
	"github.com/sakojpa/tasker/utils"
	"net/http"
	"time"
)

// TaskRouterHandler routes HTTP requests to appropriate task handlers based on request method.
func TaskRouterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		AddTask(w, r)
	}
	if r.Method == "GET" || r.Method == "PUT" {
		EditTask(w, r)
	}
	if r.Method == "DELETE" {
		DeleteTask(w, r)
	}
}

// GetAllTasksHandler retrieves all tasks filtered by search query or date.
func GetAllTasksHandler(w http.ResponseWriter, r *http.Request) {
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
	tasks, err := db.GetAllTasks(50, query, queryType)
	if err != nil {
		utils.SentErrorJson(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.SentOkMsg(w, TasksResp{Tasks: tasks}, 1)
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
		utils.SentErrorJson(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.SentOkMsg(w, taskInfo.NextDate, 2)
}

// DoneTaskHandler marks a task as done, deletes it if non-repeating, otherwise updates its next due date.
func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	task, err := db.GetTaskById(id)
	if err != nil {
		utils.SentErrorJson(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(task.Repeat) == 0 {
		err = db.DeleteTaskById(id)
		if err != nil {
			utils.SentErrorJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		utils.SentOkMsg(w, struct{}{}, 1)
		return
	} else {
		newNow, _ := time.Parse(dateFormat, task.Date)
		taskInfo, err := validateRepeatRule(newNow, task.Date, task.Repeat)
		if err != nil {
			utils.SentErrorJson(w, err.Error(), http.StatusInternalServerError)
			return
		}
		task.Date = taskInfo.NextDate
		err = db.UpdateTask(task, id)
		if err != nil {
			utils.SentErrorJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		utils.SentOkMsg(w, struct{}{}, 1)
		return
	}
}

// AuthHandler authenticates users by processing JSON-based login credentials.
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	var authRequest utils.AuthRequest
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Printf("body read error: %s\n", err.Error())
		utils.SentErrorJson(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(buf.Bytes(), &authRequest)
	if err != nil {
		fmt.Printf("unmarshal error: %s\n", err.Error())
		utils.SentErrorJson(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	resp, err, code := MakeAuth(&authRequest)
	if err != nil {
		utils.SentErrorJson(w, err.Error(), code)
		return
	}
	utils.SentOkMsg(w, resp, 1)
	return
}
