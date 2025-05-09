package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
)

// GetTaskById retrieves a task by its ID from the database.
func GetTaskById(ctx context.Context, id string) (*Task, error) {
	task := &Task{}
	err := dbConn.QueryRowContext(
		ctx,
		"SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id,
	).Scan(
		&task.ID,
		&task.Date,
		&task.Title,
		&task.Comment,
		&task.Repeat,
	)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return task, nil
}

// GetAllTasks retrieves all tasks, optionally filtering by text or date, sorted by date.
func GetAllTasks(ctx context.Context, limit int, searchQuery, searchType string) ([]*Task, error) {
	var rows *sql.Rows
	var err error
	switch searchType {
	case "text":
		queryDb := "SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? ORDER BY date ASC LIMIT ?"
		rows, err = dbConn.QueryContext(ctx, queryDb, "%"+searchQuery+"%", limit)
	case "date":
		queryDb := "SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC LIMIT ?"
		rows, err = dbConn.QueryContext(ctx, queryDb, searchQuery, limit)
	default:
		queryDb := "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?"
		rows, err = dbConn.QueryContext(ctx, queryDb, limit)
	}
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("sql request error: %w", err)
	}

	tasks := []*Task{}
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("read row error: %w", err)
		}
		tasks = append(tasks, task)
	}
	sort.Slice(
		tasks, func(i, j int) bool {
			return tasks[i].Date < tasks[j].Date
		},
	)
	return tasks, nil
}

// CreateTask inserts a new task into the database and returns its ID.
func CreateTask(ctx context.Context, task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := dbConn.ExecContext(ctx, query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

// UpdateTask modifies an existing task's data in the database by its ID.
func UpdateTask(ctx context.Context, task *Task, id string) error {
	query := `
		UPDATE scheduler SET 
                     date = ?, 
                     title = ?, 
                     comment = ?, 
                     repeat = ? 
                 WHERE 
                     id = ?
`
	res, err := dbConn.ExecContext(ctx, query, task.Date, task.Title, task.Comment, task.Repeat, id)
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("incorrect id for updating task")
	}
	return nil
}

// DeleteTaskById removes a task from the database by its ID.
func DeleteTaskById(ctx context.Context, id string) error {
	query := "DELETE FROM scheduler WHERE id = ?"
	res, err := dbConn.ExecContext(ctx, query, id)
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("incorrect id for delete task")
	}
	return nil
}
