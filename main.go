package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID     int
	Title  string
	Status string
}

var db *sql.DB

func main() {
	// Initialize database
	var err error
	db, err = sql.Open("sqlite3", "./todo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	// Setup routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/edit", editHandler)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTable() {
	query := `CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        status TEXT
    )`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, status FROM todos")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, todos)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		title := r.FormValue("title")
		status := "Pending"

		_, err := db.Exec("INSERT INTO todos (title, status) VALUES (?, ?)", title, status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		status := r.FormValue("status")

		_, err = db.Exec("UPDATE todos SET title = ?, status = ? WHERE id = ?", title, status, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
