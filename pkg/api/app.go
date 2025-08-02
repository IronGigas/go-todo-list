package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-todo-list/pkg/db"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

var dateFormat string = "20060102"

// find next date accroding to repeat rule
func NextDate(now time.Time, dateStr string, repeat string) (string, error) {

	//check if repeat rule is empty
	if repeat == "" {
		return "", errors.New("repeat is empty")
	}

	//convert dstart to time
	dstartTime, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return "", errors.New("error parse dstart")
	}

	//split repeat rule
	parts := strings.Split(repeat, " ")
	rule := parts[0]

	switch rule {
	//years
	case "y":
		nextDate := dstartTime
		//add one year till we go after now
		for {
			nextDate = nextDate.AddDate(1, 0, 0)
			if nextDate.After(now) {
				return nextDate.Format(dateFormat), nil
			}
		}
	//days
	case "d":
		//check foramt
		if len(parts) != 2 {
			return "", errors.New("wrong format repeat for days")
		}
		//convert number of days to int and check foramt once more
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("wrong format repeat for days")
		}

		nextDate := dstartTime
		//add days till we go after now
		for {
			nextDate = nextDate.AddDate(0, 0, days)
			if nextDate.After(now) {
				return nextDate.Format(dateFormat), nil
			}
		}
	//weeks
	case "w":
		//check format
		if len(parts) != 2 {
			return "", errors.New("wrong format repeat for weeks")
		}
		//get week days
		daysOfWeek := strings.Split(parts[1], ",")
		//simple array instead of map
		var weekDays [8]bool
		//fill array
		for _, day := range daysOfWeek {
			//convert to int and check
			day, err := strconv.Atoi(day)
			if err != nil || day < 1 || day > 7 {
				return "", errors.New("invalid weekday format")
			}
			if day == 7 {
				weekDays[0] = true //weird Go foramt where sunday is zero... have to handle it like this
			} else {
				weekDays[day] = true
			}
		}
		nextDate := dstartTime
		//for ten years limit (to exclude eternal loop in tets)
		for i := 0; i < 365*10; i++ {
			//add one day
			nextDate = nextDate.AddDate(0, 0, 1)
			//until we hit correct day of week and after now
			if weekDays[int(nextDate.Weekday())] && nextDate.After(now) {
				return nextDate.Format(dateFormat), nil
			}
		}
		return "", nil
	//months
	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", errors.New("wrong format repeat for months")
		}
		//parse days from month
		daysMth := strings.Split(parts[1], ",")
		//simple array instead of map for days
		var monthDays [32]bool
		//fill array and check -1 and -2 days
		checkLastDay := false
		checkSecondLastDay := false
		for _, day := range daysMth {
			day, err := strconv.Atoi(day)
			if err != nil || day == 0 || day > 31 || day < -2 {
				return "", errors.New("wrong format repeat days for month")
			}
			switch day {
			case -1:
				checkLastDay = true
			case -2:
				checkSecondLastDay = true
			default:
				monthDays[day] = true
			}
		}
		//another array instead of map for months
		var months [13]bool
		//mark if actual moth is present in repeat rule
		allMonths := true
		//fill array
		if len(parts) == 3 {
			allMonths = false
			monthsStr := strings.Split(parts[2], ",")
			if len(monthsStr) == 0 {
				return "", errors.New("invalid number of months")
			}
			for _, monthStr := range monthsStr {
				month, err := strconv.Atoi(monthStr)
				if err != nil || month < 1 || month > 12 {
					return "", errors.New("invalid month should be from 1 to 12")
				}
				months[month] = true
			}
		}
		//find next date
		nextDate := dstartTime
		for i := 0; i < 365*10; i++ { //10 years limit to avoid enetral loop
			//add day
			nextDate = nextDate.AddDate(0, 0, 1)
			//check -1 and -2 conditions in the past
			if !nextDate.After(now) {
				continue
			}
			if !allMonths && !months[int(nextDate.Month())] {
				continue
			}
			//if this date is the date, return it
			dayMatch := monthDays[nextDate.Day()] || (checkLastDay && isLastDayOfMonth(nextDate)) || (checkSecondLastDay && isSecondToLastDayOfMonth(nextDate))
			if dayMatch {
				return nextDate.Format(dateFormat), nil
			}
		}
		return "", errors.New("wrong repeat rule")
	default:
		return "", nil
	}
}

// check date and apply NextDate logic if necessary, separate logic which is used in hadlers few times
func checkDate(task *db.Task) error {
	now := time.Now()
	// set time to zero in roder to avoid errors
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if task.Date == "" {
		task.Date = today.Format(dateFormat)
		return nil
	}

	parsedDate, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}

	// if date in the past, calculate next one
	if parsedDate.Before(today) {
		if task.Repeat == "" {
			task.Date = today.Format(dateFormat) // set as today if repeat rule is empty
		} else {
			// pass today and not now sa we need zero time
			next, err := NextDate(today, task.Date, task.Repeat)
			if err != nil || next == "" {
				return fmt.Errorf("invalid repeat rule")
			}
			task.Date = next
		}
	}
	return nil
}

// small additional func for "m -1" repeat rule logic
func isLastDayOfMonth(t time.Time) bool {
	return t.AddDate(0, 0, 1).Day() == 1
}

// small additional func for "m -2" repeat rule logic
func isSecondToLastDayOfMonth(t time.Time) bool {
	return t.AddDate(0, 0, 2).Day() == 1 || (t.Month() != t.AddDate(0, 0, 2).Month())
}

// special func for writing into json
func writeJson(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		//log error into server
		log.Printf("could not encode json: %v", err)
	}
}
