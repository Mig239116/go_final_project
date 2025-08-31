package api

import (
	"encoding/json"
	"fmt"
	"go_final_project/pkg/db"
	"net/http"
	"time"
)

// addTaskHandler обработчик эндпоинта добавления задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := validateAndProcessTask(&task); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSONError(w, "Failed to add task to database", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id": id,
	}
	writeJSON(w, response, http.StatusOK)
}

// validateAndProcessTask валидация полей задачи и заполнение дефолтными значениями
func validateAndProcessTask(task *db.Task) error {
	if task.Title == "" {
		return fmt.Errorf("title is required")
	}

	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	}
	t, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}
	if !afterNow(t, now) && task.Repeat != "" && !isSameDate(t, now) {
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("invalid repeat rule: %v", err)
		}
		task.Date = nextDate
	} else if !afterNow(t, now) {
		task.Date = now.Format(DateFormat)
	}
	if task.Repeat != "" {
		_, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("invalid repeat rule: %v", err)
		}
	}
	return nil
}

// writeJSON кодирует данные для ответа
func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// writeJSONError кодирует данные об ошибке
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]string{
		"error": message,
	}
	writeJSON(w, response, statusCode)
}

// getTaskHandler обработка эндпоинта для получения задачи из БД
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeJSONError(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

// isSameDate проверяет что даты одинаковые
func isSameDate(t1, t2 time.Time) bool {
    return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}

// updateTaskHandler обработка эндпоинта для обновления задачи
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		writeJSONError(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	if err := validateAndProcessTask(&task); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		writeJSONError(w, "Failed to update task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}