// main.go
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

	// Set up routes
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/add", handlers.AddHandler)
	http.HandleFunc("/delete", handlers.DeleteHandler)
	http.HandleFunc("/edit", handlers.EditHandler)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
