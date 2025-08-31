package api

import (
	"go_final_project/pkg/db"
	"net/http"
	"strconv"
	"time"
)

type TasksResponse struct {
	Tasks []*db.Task `json:"tasks"`
}

// Init производит инициализацию обработчиков для эндпоинтов
func Init() {
	http.HandleFunc("/api/nextdate", NextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/task/done", taskDoneHandler)
}

// NextDateHandler обработчик эндпоинта для вычисления следующей даты
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			http.Error(w, "Invalid now parameter", http.StatusBadRequest)
			return
		}
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}

// taskHandler обработчик для отдельной задачи POST, GET, PUT, DELETE
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// tasksHandler обработчик для получения списка задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	search := r.FormValue("search")
	limitStr := r.FormValue("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	var tasks []*db.Task
	var err error
	if search != "" {
		tasks, err = db.SearchTasks(search, limit)
	} else {
		tasks, err = db.GetTasks(limit)
	}
	if err != nil {
		writeJSONError(w, "Failed to get tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}
	response := TasksResponse{Tasks: tasks}
	writeJSON(w, response, http.StatusOK)
}

// deleteTaskHandler удаление задачи
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeJSONError(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	err := db.DeleteTask(id)
	if err != nil {
		writeJSONError(w, "Failed to delete task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}
