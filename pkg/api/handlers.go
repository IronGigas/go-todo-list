package api

import (
	"encoding/json"
	"io"
	"net/http"
	"go-todo-list/pkg/db"
	"time"
)

const limit = 50

// handler for next date
func nextDateHandler(w http.ResponseWriter, r *http.Request) {

	//check request method
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	nowString := r.FormValue("now")
	dstartString := r.FormValue("date")
	repeatString := r.FormValue("repeat")

	var now time.Time
	var err error

	//parse now to required format
	if nowString == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateFormat, nowString)
		if err != nil {
			http.Error(w, "invalid now format", http.StatusBadRequest)
			return
		}
	}

	//get NextDate
	next, err := NextDate(now, dstartString, repeatString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(next)); err != nil {
		writeJson(w, map[string]string{"error": "failed to write response"}, http.StatusInternalServerError)
		return
	}
}

// handler for insert statement into db
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJson(w, map[string]string{"error": "JSON parsing error"}, http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &task); err != nil {
		writeJson(w, map[string]string{"error": "cannot unmarshal json"}, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJson(w, map[string]string{"error": "title cannot be empty"}, http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": "cannot save task into structure"}, http.StatusInternalServerError)
		return
	}

	writeJson(w, map[string]any{"id": id}, http.StatusOK)
}

// handler for db query:select all rows
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	//check request method
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")

	//search variable added
	tasks, err := db.Tasks(limit, search)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJson(w, TasksResp{Tasks: tasks}, http.StatusOK)
}

// handler for db query:select one row
func getTaskHandler(w http.ResponseWriter, r *http.Request) {

	//get id for select
	id := r.FormValue("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	//select one row
	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	writeJson(w, task, http.StatusOK)
}

// handler for db query:update one row
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {

	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	//check if id is not empty
	if task.ID == "" {
		writeJson(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	//check if title is not empty
	if task.Title == "" {
		writeJson(w, map[string]string{"error": "title is required"}, http.StatusBadRequest)
		return
	}

	// adjust date if necessary
	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return //exit function in case of error
	}

	//run update query
	if err := db.UpdateTask(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return //exit function in case of error
	}

	writeJson(w, map[string]string{}, http.StatusOK)
}

// handler for db query:delete by id
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	//get id
	id := r.FormValue("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}
	//run delete
	if err := db.DeleteTask(id); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	writeJson(w, map[string]string{}, http.StatusOK)
}

// handler for db query:update date by id or delete by id
func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	//check request method
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	nowString := r.FormValue("now")
	if id == "" {
		writeJson(w, map[string]string{"error": "wrong 'now' for some reason"}, http.StatusBadRequest)
		return
	}

	var now time.Time
	var err error
	//parse now back to time
	if nowString == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateFormat, nowString)
		if err != nil {
			writeJson(w, map[string]string{"error": "invalid now format"}, http.StatusBadRequest)
			return
		}
	}

	//get task
	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	// delete if no repeat
	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		// else find date and update it
	} else {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
			return
		}
		if err := db.UpdateDate(next, id); err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
	}

	writeJson(w, map[string]string{}, http.StatusOK)
}
