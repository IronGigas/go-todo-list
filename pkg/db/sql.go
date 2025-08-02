package db

import (
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

// note that I cannot import fromat from api, as it would cause cyclical import, so have to redefine in here
const parseDateFormat = "02.01.2006"
const dateFormat = "20060102"

// insert sql func
func AddTask(task *Task) (int64, error) {
	var id int64
	//insert statement
	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	//execute insert
	result, err := Db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = result.LastInsertId()
	}

	return id, err
}

// select all sql func
func Tasks(limit int, search string) ([]*Task, error) {
	var query string
	//slice of epmty interfaces serves here as slice that can recieve strings and ints
	var arguments []interface{}

	//search logic added, prepare select
	//select without search condition
	if search == "" {
		query = `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`
		arguments = append(arguments, limit)
		//select with date search condition
	} else if datetime, err := time.Parse(parseDateFormat, search); err == nil {
		dateString := datetime.Format(dateFormat)
		query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? LIMIT ?`
		arguments = append(arguments, dateString, limit)
		//select with string search pattern
	} else {
		query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?`
		searchPattern := "%" + search + "%"
		arguments = append(arguments, searchPattern, searchPattern, limit)
	}

	//run select
	rows, err := Db.Query(query, arguments...)
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	//upload result set into tasks structure
	for rows.Next() {
		task := &Task{}
		var id int64
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		task.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, task)
	}
	//check errors in interation
	if err := rows.Err(); err != nil {
		return nil, err
	}

	//create empty structure if result set is empty
	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, nil
}

// select one row sql func
func GetTask(id string) (*Task, error) {

	var taskId int64

	//select query
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`

	//execure query
	task := &Task{}
	err := Db.QueryRow(query, id).Scan(&taskId, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, err
	}

	task.ID = strconv.FormatInt(taskId, 10)

	return task, nil
}

// update sql func
func UpdateTask(task *Task) error {

	//update query
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`

	//execure query
	result, err := Db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	//check errors
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("incorrect id for updating task")
	}

	return nil
}

// delete sql func
func DeleteTask(id string) error {

	//check if id is ok
	if id == "" {
		return fmt.Errorf("invalid id")
	}

	//delete query
	query := `DELETE FROM scheduler WHERE id = ?`

	//execute query
	result, err := Db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("db delete error: %w", err)
	}

	//check errors
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// update date sql func
func UpdateDate(next, id string) error {

	//check if id is ok
	if id == "" {
		return fmt.Errorf("invalid id")
	}

	//update query
	query := `UPDATE scheduler SET date = ? WHERE id = ?`

	//execute query
	result, err := Db.Exec(query, next, id)
	if err != nil {
		return fmt.Errorf("db update date error: %w", err)
	}

	//check errors
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}
