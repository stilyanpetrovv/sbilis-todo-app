// db/init.go
package db

import (
	"database/sql"
	"fmt"

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

	// Create the users table if it doesn't exist
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT
	);`

	_, err = DB.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		title TEXT,
		status BOOLEAN DEFAULT 0,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	if _, err = DB.Exec(createTableQuery); err != nil {
		return err
	}

	return nil
}

// Close closes the database connection, to be called when shutting down the app.
func Close() error {
	return DB.Close()
}
