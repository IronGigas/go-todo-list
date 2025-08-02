package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var Db *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    date    CHAR(8)      DEFAULT '',
    title   VARCHAR(255) DEFAULT '',
    comment TEXT         DEFAULT '',
    repeat  VARCHAR(128) DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`
const defaultDb = "scheduler.db"

func Init(dbFile string) error {
	
	//if TODO_DBFILE variable is empty use default path (TODO_DBFILE should be something like pkg/db/scheduler.db)
	if dbFile == "" {
		dbFile = defaultDb
	}
	
	//check if schduler.db exists
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	//open database file or create one if there is none
	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	//if there was no schduler.db previously: create table and index
	if install {
		if _, err := database.Exec(schema); err != nil {
			return fmt.Errorf("failed to create table or index: %w", err)
		}
	}

	Db = database
	return nil
}
