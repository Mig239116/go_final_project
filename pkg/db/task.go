package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет задачу в базу данных
func AddTask(task *Task) (string, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) 
	          VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return "", fmt.Errorf("can not post task %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("can not get task id %v", err)
	}
	return fmt.Sprintf("%d", id), nil
}

// GetTask получает задачу из базы данных
func GetTask(id string) (*Task, error) {
	query := `SELECT id, date, title, comment, repeat 
	          FROM scheduler 
	          WHERE id = ?`
	var task Task
	var taskId int64
	err := DB.QueryRow(query, id).Scan(&taskId, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %v", err)
	}

	task.ID = fmt.Sprintf("%d", taskId)
	return &task, nil
}

// UpdateTask обновляет задачу в БД
func UpdateTask(task *Task) error {
	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid task ID")
	}

	query := `UPDATE scheduler 
	          SET date = ?, title = ?, comment = ?, repeat = ? 
	          WHERE id = ?`

	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, id)
	if err != nil {
		return fmt.Errorf("failed to update task: %v", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// DeleteTask удаляет задачу из БД
func DeleteTask(id string) error {
    taskId, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
		return fmt.Errorf("invalid task ID")
	}

    query := `DELETE FROM scheduler WHERE id = ?`
    res, err := DB.Exec(query, taskId)
    if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}
    count, err := res.RowsAffected()
    if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
    if count == 0 {
		return fmt.Errorf("task not found")
	}
    return nil
}

// UpdateDate обновляет дату задачи в БД
func UpdateDate(id string, date string) error {
    taskId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid task ID")
	}
    query := `UPDATE scheduler SET date = ? WHERE id = ?`
    res, err := DB.Exec(query, date, taskId)
	if err != nil {
		return fmt.Errorf("failed to update task date: %v", err)
	}
    count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
    if count == 0 {
		return fmt.Errorf("task not found")
	}
    return nil
}

// GetTasks возвращает список задач в указанном количестве (по умолчанию 50)
func GetTasks(limit int) ([]*Task, error) {
	query := `SELECT id, date, title, comment, repeat 
	          FROM scheduler 
	          ORDER BY date ASC 
	          LIMIT ?`
	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks %v", err)
	}
	defer rows.Close()
	var tasks []*Task
	for rows.Next() {
		var task Task
		var id int64
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %v", err)
		}
		task.ID = fmt.Sprintf("%d", id)
		tasks = append(tasks, &task)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %v", err)
	}
	if tasks == nil {
		tasks = []*Task{}
	}
	return tasks, nil
}

// SearchTasks ищет задачи по строке
func SearchTasks(search string, limit int) ([]*Task, error) {
	if isDateSearch(search) {
		date, err := convertSearchDate(search)
		if err != nil {
			return nil, err
		}
		return getTasksByDate(date, limit)
	}
	return getTasksByTextSearch(search, limit)
}

// isDateSearch проверяет что поисковая строка является датой
func isDateSearch(search string) bool {
	if len(search) != 10 {
		return false
	}
	if search[2] != '.' || search[5] != '.' {
		return false
	}
	_, err := time.Parse("02.01.2006", search)
	return err == nil
}

// convertSearchDate конвертирует поисковую строку в дату
func convertSearchDate(search string) (string, error) {
	t, err := time.Parse("02.01.2006", search)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}
	return t.Format("20060102"), nil
}

// getTasksByDate возвращает указанное количество задач по дате
func getTasksByDate(date string, limit int) ([]*Task, error) {
	query := `SELECT id, date, title, comment, repeat 
	          FROM scheduler 
	          WHERE date = ? 
	          ORDER BY date ASC 
	          LIMIT ?`

	rows, err := DB.Query(query, date, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by date: %v", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// getTasksByTextSearch возвращает указанное количество задач по тексту в комменте или заголовке
func getTasksByTextSearch(search string, limit int) ([]*Task, error) {
	searchPattern := "%" + search + "%"
	query := `SELECT id, date, title, comment, repeat 
	          FROM scheduler 
	          WHERE title LIKE ? OR comment LIKE ? 
	          ORDER BY date ASC 
	          LIMIT ?`

	rows, err := DB.Query(query, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by search: %v", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// scanTasks создает список задач
func scanTasks(rows *sql.Rows) ([]*Task, error) {
	var tasks []*Task
	for rows.Next() {
		var task Task
		var id int64
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %v", err)
		}
		task.ID = fmt.Sprintf("%d", id)
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %v", err)
	}

	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, nil
}
