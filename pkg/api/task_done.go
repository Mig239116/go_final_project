package api

import (
	"go_final_project/pkg/db"
	"net/http"
	"time"
)

// taskDoneHandler обработчик эндпоинта выполнения задачи
func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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
	now := time.Now()
	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			writeJSONError(w, "Failed to delete task: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJSONError(w, "Failed to calculate next date: "+err.Error(), http.StatusInternalServerError)
			return
		}
	err = db.UpdateDate(id, nextDate)
		if err != nil {
			writeJSONError(w, "Failed to update task date: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}
