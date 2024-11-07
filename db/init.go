// db/init.go
package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Initialize opens the database connection and creates the todos table if it doesn't exist.
func Initialize() error {
	var err error
	DB, err = sql.Open("sqlite3", "./db/todo.db")
	if err != nil {
		return err
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        status TEXT
    )`

	if _, err = DB.Exec(createTableQuery); err != nil {
		return err
	}

	return nil
}

// Close closes the database connection, to be called when shutting down the app.
func Close() error {
	return DB.Close()
}
