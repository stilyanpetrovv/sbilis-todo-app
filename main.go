package main

import (
	"fmt"
	"log"
	"net/http"
	"sbilis-todo-app/db"
	"sbilis-todo-app/handlers"
)

func main() {
	// Initialize the database
	if err := db.Initialize(); err != nil {
		log.Fatalf("Could not set up the database: %v", err)
	}
	defer db.Close()

	// Serve static files (should be before other routes)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Register routes that do NOT require authentication
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// Register routes that DO require authentication
	http.Handle("/tasks", handlers.AuthMiddleware(handlers.TasksHandler))
	http.Handle("/add", handlers.AuthMiddleware(handlers.AddHandler))
	http.Handle("/delete", handlers.AuthMiddleware(handlers.DeleteHandler))
	http.Handle("/edit", handlers.AuthMiddleware(handlers.EditHandler))

	// Start the server
	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
