package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	db "github.com/sakojpa/tasker/pkg/database"
	"github.com/sakojpa/tasker/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AddTask adds a new task to the database and returns its ID.
func AddTask(w http.ResponseWriter, r *http.Request) {
	task, err := checkBody(r)
	if err != nil {
		utils.SentErrorJson(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := db.CreateTask(task)
	if err != nil {
		http.Error(w, "Database error", http.StatusBadRequest)
		return
	}
	resp := struct {
		ID int64 `json:"id"`
	}{ID: id}
	utils.SentOkMsg(w, resp, 1)
	return
}

// EditTask handles retrieving or updating a specific task by its ID.
func EditTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id := r.FormValue("id")
		task, err := db.GetTaskById(id)
		if err != nil {
			utils.SentErrorJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		utils.SentOkMsg(w, task, 1)
		return

	case "PUT":
		task, err := checkBody(r)
		if err != nil {
			utils.SentErrorJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(task.ID) == "" {
			utils.SentErrorJson(w, "Task ID absent", http.StatusBadRequest)
			return
		}
		if _, err := strconv.Atoi(task.ID); err != nil {
			utils.SentErrorJson(w, "Task ID error", http.StatusBadRequest)
			return
		}
		err = db.UpdateTask(task, task.ID)
		if err != nil {
			utils.SentErrorJson(w, err.Error(), http.StatusBadRequest)
			return
		}
		utils.SentOkMsg(w, struct{}{}, 1)
		return
	}
}

// DeleteTask removes a task by its ID from the database.
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	err := db.DeleteTaskById(id)
	if err != nil {
		utils.SentErrorJson(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.SentOkMsg(w, struct{}{}, 1)
	return
}

func checkDate(task *db.Task) error {
	now := time.Now()
	if task.Date == "" || strings.TrimSpace(task.Date) == "" {
		task.Date = now.Format(dateFormat)
	}
	parsedTime, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("wrong date format, expected format is %s", dateFormat)
	}
	if utils.AfterNow(now, parsedTime) {
		if len(task.Repeat) == 0 {
			task.Date = now.Format(dateFormat)
		} else {
			next, err := validateRepeatRule(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("repeat rule error: %w", err)
			}
			task.Date = next.NextDate
		}
	}
	return nil
}

func checkBody(r *http.Request) (*db.Task, error) {
	var task db.Task
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("body read error")
	}
	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		return nil, fmt.Errorf("marshaller error")
	}
	if strings.TrimSpace(task.Title) == "" {
		return nil, fmt.Errorf("task title absent")
	}
	if err = checkDate(&task); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return &task, nil
}
